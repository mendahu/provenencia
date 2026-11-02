import Foundation
import Observation

/// Project-scoped catalog query cache. Each kind's loader, policies, patches,
/// and refresh live on `CatalogQueryKey.Kind.descriptor`.
@MainActor
@Observable
final class WorkspaceSession {
    let projectKey: ProjectKey
    let store: any GenealogyStore
    let registry: CatalogQueryRegistry
    /// Default name-format orders. Empty until the catalog read finishes,
    /// which keeps stored part order.
    var nameFormat = NameFormatOrders()

    private var handles: [CatalogQueryKey: any AnyQueryHandle] = [:]
    private var valueTypes: [CatalogQueryKey: ObjectIdentifier] = [:]
    private var generations: [CatalogQueryKey: Int] = [:]
    private var inFlight: [CatalogQueryKey: (generation: Int, task: Task<Void, Never>)] = [:]
    /// Keys of the place on screen, from the last `apply(location:)`.
    private(set) var visibleKeys: Set<CatalogQueryKey> = []
    /// Transient notice after leaving a place (composer save, etc.).
    var noticeToast: VocabularyToast?
    /// Recounts the sidebar after a write in `CatalogQueryRegistry.navCountTriggers`. Set by `WorkspaceView`.
    @ObservationIgnored
    var onNavCountsMoved: (@MainActor () async -> Void)?
    /// Evidence-graph pan and zoom, keyed by graph id. Not catalog data; quit drops it.
    @ObservationIgnored
    private var graphCameras: [String: GraphCanvasCamera] = [:]
    /// The Evidence graph list indexed by id and by Source, rebuilt when the
    /// list handle's revision moves (`WorkspaceSession+EvidenceGraphs.swift`).
    /// Not observed: filling it while a view reads must not invalidate that view.
    @ObservationIgnored
    var evidenceGraphIndex: EvidenceGraphIndex?
    /// How many times the index was built in full (a test hook: a merged
    /// re-read splices instead).
    @ObservationIgnored
    var evidenceGraphIndexBuildCount = 0
    /// Bridge to select the next time that graph appears. Consumed once.
    private var pendingGraphSelections: [String: String] = [:]

    init(
        projectKey: ProjectKey,
        store: any GenealogyStore,
        registry: CatalogQueryRegistry = .standard
    ) {
        self.projectKey = projectKey
        self.store = store
        self.registry = registry
    }

    func loadNameFormat() async {
        do {
            nameFormat = try await store.nameFormat(projectDir: projectKey.projectDir)
        } catch {
            nameFormat = NameFormatOrders()
        }
    }

    func graphCamera(for graphID: String) -> GraphCanvasCamera? {
        graphCameras[graphID]
    }

    func setGraphCamera(_ camera: GraphCanvasCamera, for graphID: String) {
        guard graphCameras[graphID] != camera else { return }
        graphCameras[graphID] = camera
    }

    func setPendingGraphSelection(subjectID: String, graphID: String) {
        pendingGraphSelections[graphID] = subjectID
    }

    func consumePendingGraphSelection(graphID: String) -> String? {
        let id = pendingGraphSelections[graphID]
        pendingGraphSelections[graphID] = nil
        return id
    }

    /// Returns an existing handle without starting a load.
    func queryHandle<Value>(_ key: CatalogQueryKey) -> QueryHandle<Value>? {
        handles[key] as? QueryHandle<Value>
    }

    /// True while the key's handle holds a value and is reloading it behind
    /// that value: "refreshing" for a section whose first load is its own spinner.
    func isRefreshing(_ key: CatalogQueryKey) -> Bool {
        handles[key].map { $0.hasValue && $0.isFetching } ?? false
    }

    /// Waits out an in-flight load and every partial re-read a write started
    /// for the key, then returns the cached value.
    /// Does not start a load. Nil when the key was never warmed, is stale, or the load failed.
    func readyValue<Value>(_ key: CatalogQueryKey) async -> Value? {
        while true {
            var pending: [Task<Void, Never>] = []
            if let task = inFlight[key]?.task { pending.append(task) }
            pending += writeRefreshes.values.filter { $0.key == key }.map(\.task)
            if pending.isEmpty { break }
            for task in pending {
                await task.value
            }
        }
        if handles[key]?.isStale == true {
            return nil
        }
        return queryHandle(key)?.value
    }

    /// Registry-backed get-or-load.
    @discardableResult
    func query<Value>(_ key: CatalogQueryKey) -> QueryHandle<Value> {
        key.kind.descriptor.warm(self, key)
        guard let handle: QueryHandle<Value> = queryHandle(key) else {
            preconditionFailure("CatalogQueryKey \(key) warmed a different value type than \(Value.self)")
        }
        return handle
    }

    /// Invalidates the key, then queries it: the read-it-again path. Keeps any
    /// stale value on the handle while the new read runs.
    @discardableResult
    func reread<Value>(_ key: CatalogQueryKey) -> QueryHandle<Value> {
        invalidate(key)
        return query(key)
    }

    /// Synchronous cache write. Clears stale flag without calling the loader.
    /// If a load was in flight, the patch wins immediately and that load is
    /// restarted so a later refetch is not dropped.
    func setQueryValue<Value>(_ key: CatalogQueryKey, value: Value) {
        replaceValue(key, value, clearStale: true)
    }

    /// The one patch primitive: replaces the cached value now. An in-flight
    /// load is cancelled and restarted, so an older response cannot overwrite
    /// the patch. `clearStale` says whether the patch makes the key current.
    private func replaceValue<Value>(_ key: CatalogQueryKey, _ value: Value, clearStale: Bool) {
        let hadInFlight = inFlight[key] != nil
        bumpGeneration(key)
        inFlight[key]?.task.cancel()
        inFlight[key] = nil
        let handle = typedHandle(for: key, as: Value.self)
        handle.applySuccess(value)
        if clearStale {
            handle.clearStale()
        }
        if hadInFlight {
            handle.markStale()
            revalidate(key)
        }
    }

    /// Edits the cached value in place. `change` returns whether it changed
    /// anything; nothing happens when it did not or the key has no value. The
    /// stale flag is left as it was: a partial patch does not make the key current.
    /// A key whose first load is in flight has no value to patch and that load
    /// may have read before the write: the result says so (`.awaitingFirstValue`)
    /// and `apply` restarts or stales the key.
    @discardableResult
    func patchCached<Value>(
        _ key: CatalogQueryKey,
        as _: Value.Type,
        _ change: (inout Value) -> Bool
    ) -> CatalogPatchResult {
        if awaitsFirstValue(key) { return .awaitingFirstValue }
        guard let handle: QueryHandle<Value> = queryHandle(key), var value = handle.value else { return .unchanged }
        guard change(&value) else { return .unchanged }
        replaceValue(key, value, clearStale: false)
        return .applied
    }

    /// Re-reads part of a warmed value and merges it in: `read` runs once per
    /// id, in order, and every part is merged in one `patchCached`. A failed
    /// read falls back to the key's hidden policy (a reload) and merges
    /// nothing. Nothing happens for a key evicted before it starts. `apply`
    /// holds the handle's `isFetching` count for the whole run. After an
    /// applied merge, `afterMerge` gets the parts by id and the handle
    /// revision from just before the patch.
    ///
    /// Called only by `CatalogQueryRefresh.merged`, inside `apply`'s gate and
    /// count; it is not a feature API.
    func refreshCached<Value, Part>(
        _ key: CatalogQueryKey,
        as _: Value.Type,
        ids: [String],
        read: @MainActor (String) async throws -> Part,
        merge: @MainActor (inout Value, String, Part) -> Void,
        afterMerge: (@MainActor ([String: Part], _ previousRevision: Int) -> Void)? = nil
    ) async {
        guard let _: QueryHandle<Value> = queryHandle(key) else { return }
        var parts: [(String, Part)] = []
        for id in ids {
            do {
                parts.append((id, try await read(id)))
            } catch {
                applyHiddenPolicy(to: key)
                return
            }
        }
        let previousRevision = queryHandle(key).map { (handle: QueryHandle<Value>) in handle.revision }
        let result = patchCached(key, as: Value.self) { value in
            for (id, part) in parts {
                merge(&value, id, part)
            }
            return true
        }
        // Any other result (awaiting a first value, unchanged) is `apply`'s to settle.
        if case .applied = result, let previousRevision, let afterMerge {
            afterMerge(Dictionary(parts, uniquingKeysWith: { _, last in last }), previousRevision)
        }
    }

    /// What a key does when a write reaches it: a visible one revalidates, a
    /// hidden one follows its descriptor's `whenHidden`.
    private func effectivePolicy(for key: CatalogQueryKey) -> CatalogHiddenPolicy {
        visibleKeys.contains(key) ? .revalidate : registry.hiddenPolicy(key)
    }

    /// The key has no value yet and its first load is in flight; that load
    /// may have read before the write that is being applied.
    private func awaitsFirstValue(_ key: CatalogQueryKey) -> Bool {
        handles[key].map { !$0.hasValue } == true && inFlight[key] != nil
    }

    /// Resolves the place for `location` and starts loading its query keys (non-blocking).
    func apply(
        location: WorkspaceLocation,
        placeRegistry: PlaceRegistry = .standard
    ) {
        guard let place = placeRegistry.resolve(location, project: projectKey) else {
            visibleKeys = []
            return
        }
        visibleKeys = Set(place.queryKeys)
        #if DEBUG
        if ProcessInfo.processInfo.environment["PROVENENCIA_DEBUG_SESSION_APPLY"] == "1" {
            print("WorkspaceSession.apply location=\(location) place=\(place.placeID) keys=\(place.queryKeys)")
        }
        #endif
        for key in place.queryKeys {
            key.kind.descriptor.warm(self, key)
        }
    }

    /// The one application of a write. Patches in place first, with no load:
    /// each warmed key runs its descriptor `patches` in order (last-updated
    /// cells from the effects, then an enriched Source row and page shell).
    /// Then it revalidates, once per key through one dedupe set, the union of
    /// what the registry invalidates (the catalog-wide lists and structure the
    /// effects reported, and what the mutation names: the Source page's own
    /// workspace, the metadata field list), the warmed keys whose descriptor
    /// `refreshes` follow ids the effects reported (`revalidatedKeys`: each
    /// Evidence graph it names, the handles that changed), the keys a patch
    /// found with their first load still in flight, and the keys a `.run`
    /// refresh cannot run for. Keys the registry evicts or marks stale when
    /// hidden are not loaded unless on screen. Last, it starts each warmed
    /// key's `.run` refreshes for the ids the effects reported (graph
    /// progress per graph, the graph list per Source, only the ids named),
    /// gated per key: one runs only for a key that would revalidate (visible,
    /// or `whenHidden: .revalidate`), holds a value and is not stale, and
    /// holds a count on the handle's `isFetching` until it ends. Any other
    /// warmed key such an entry names joins the dedupe pass instead (reloaded
    /// when visible, marked stale when hidden), and a key in that pass starts
    /// no `.run`. It also runs `onNavCountsMoved`
    /// for a mutation in `CatalogQueryRegistry.navCountTriggers`. Features only
    /// call `apply`; they do not query afterward or tell sibling views to
    /// refresh.
    func apply(_ effects: CatalogWriteEffects = .empty, _ mutation: CatalogMutation? = nil) {
        let awaiting = applyPatches(effects, mutation)
        let invalidated = cachedKeys(for: registry.invalidations(for: effects, by: mutation, project: projectKey))
        let refreshed = revalidatedKeys(for: effects, mutation)
        var running: [(key: CatalogQueryKey, run: CatalogQueryRefresh.Run, ids: [String])] = []
        var fallbacks: [CatalogQueryKey] = []
        for key in Array(handles.keys) where key.project == projectKey {
            guard let handle = handles[key] else { continue }
            for refresh in key.kind.descriptor.refreshes {
                guard case .run(let run) = refresh.action else { continue }
                let ids = refresh.follows.ids(in: effects, mutation: mutation)
                if ids.isEmpty { continue }
                if effectivePolicy(for: key) == .revalidate, !handle.isStale, handle.hasValue {
                    running.append((key, run, ids))
                } else {
                    fallbacks.append(key)
                }
            }
        }
        var seen = Set<CatalogQueryKey>()
        for key in invalidated + refreshed + awaiting + fallbacks where seen.insert(key).inserted {
            applyHiddenPolicy(to: key)
        }
        for entry in running where !seen.contains(entry.key) {
            guard let handle = handles[entry.key] else { continue }
            handle.beginRefresh()
            trackWrite(for: entry.key) {
                await entry.run(self, entry.key, entry.ids)
                handle.endRefresh()
            }
        }
        if let kind = mutation?.invalidationKind,
           CatalogQueryRegistry.navCountTriggers.contains(kind),
           let recount = onNavCountsMoved
        {
            trackWrite { await recount() }
        }
    }

    /// The warmed keys of this project that a write revalidates through their
    /// descriptor `refreshes`: a `.revalidate` refresh names a key once when the
    /// ids its `follows` slot reports meet the ids the key depends on, or, for a
    /// key that depends on any id, when the slot reports one. Synchronous, so a
    /// key declared here is revalidated in the same pass as the registry's
    /// invalidations.
    func revalidatedKeys(for effects: CatalogWriteEffects, _ mutation: CatalogMutation? = nil) -> [CatalogQueryKey] {
        var named: [CatalogQueryKey] = []
        var seen = Set<CatalogQueryKey>()
        for key in Array(handles.keys) where key.project == projectKey {
            for refresh in key.kind.descriptor.refreshes {
                guard case .revalidate(let dependsOn) = refresh.action else { continue }
                let reported = Set(refresh.follows.ids(in: effects, mutation: mutation))
                if reported.isEmpty { continue }
                let meets = dependsOn(self, key).map { !$0.isDisjoint(with: reported) } ?? true
                if meets, seen.insert(key).inserted {
                    named.append(key)
                }
            }
        }
        return named
    }

    /// The refresh and recount tasks `apply` has started and not yet finished.
    /// Each removes itself when it ends, so the list does not grow.
    /// A refresh names its key so `readyValue` can wait for the rows it merges.
    private var writeRefreshes: [UUID: (key: CatalogQueryKey?, task: Task<Void, Never>)] = [:]

    private func trackWrite(for key: CatalogQueryKey? = nil, _ work: @escaping @MainActor () async -> Void) {
        let id = UUID()
        writeRefreshes[id] = (key, Task { @MainActor [weak self] in
            await work()
            self?.writeRefreshes[id] = nil
        })
    }

    /// Returns once every refresh and recount `apply` started and every query
    /// load in flight have finished, following any started while it waited.
    /// Tests await this instead of polling against a deadline, which a busy
    /// main actor can outlast. Each task is awaited once, so a load that
    /// finished under a superseded generation cannot keep it spinning.
    func settleWrites() async {
        var awaited = Set<Task<Void, Never>>()
        while true {
            let pending = writeRefreshes.values.map(\.task) + inFlight.values.map(\.task)
            let next = pending.filter { !awaited.contains($0) }
            if next.isEmpty { return }
            for task in next {
                awaited.insert(task)
                await task.value
            }
            await Task.yield()
        }
    }

    /// A mutation with no reported effects.
    func apply(_ mutation: CatalogMutation) {
        apply(.empty, mutation)
    }

    /// Runs each warmed key's descriptor patches. A patch with `follows` gets
    /// the ids that slot names, and only when the write reported activity.
    /// A patch with no `follows` reads the mutation. Returns the keys a patch
    /// could not write because their first load is still in flight.
    private func applyPatches(_ effects: CatalogWriteEffects, _ mutation: CatalogMutation?) -> [CatalogQueryKey] {
        var awaiting: [CatalogQueryKey] = []
        for key in Array(handles.keys) where key.project == projectKey {
            for patch in key.kind.descriptor.patches {
                let result: CatalogPatchResult
                if let follows = patch.follows {
                    guard effects.activity.revision > 0 else { continue }
                    let ids = Set(follows.ids(in: effects, mutation: mutation))
                    if ids.isEmpty { continue }
                    result = patch.apply(self, key, ids, effects.activity, mutation)
                } else {
                    result = patch.apply(self, key, [], effects.activity, mutation)
                }
                if result == .awaitingFirstValue { awaiting.append(key) }
            }
        }
        return awaiting
    }

    /// What an invalidated key does: a visible one reloads, a hidden one
    /// follows its descriptor's `whenHidden` (evict, mark stale, or reload).
    /// The one path for a registry invalidation and for a descriptor
    /// `refreshes` entry that revalidates a key rather than patching it.
    func applyHiddenPolicy(to key: CatalogQueryKey) {
        switch effectivePolicy(for: key) {
        case .evict:
            evict(key)
        case .markStale:
            // The old value stays; the next `query` reloads it.
            handles[key]?.markStale()
        case .revalidate:
            invalidate(key)
            revalidate(key)
        }
    }

    private func cachedKeys(for invalidations: [CatalogQueryInvalidation]) -> [CatalogQueryKey] {
        var keys: [CatalogQueryKey] = []
        for invalidation in invalidations {
            switch invalidation {
            case .key(let key):
                keys.append(key)
            case .allCached(let kind):
                keys.append(contentsOf: handles.keys.filter {
                    $0.kind == kind && $0.project == projectKey
                })
            }
        }
        return keys
    }

    /// Get-or-load with in-flight dedupe. Reuses cached `.ready` data until invalidated.
    @discardableResult
    func ensureQuery<Value>(
        _ key: CatalogQueryKey,
        loader: @Sendable @escaping () async throws -> Value
    ) -> QueryHandle<Value> {
        let handle = typedHandle(for: key, as: Value.self)

        if handle.status == .ready, !handle.isStale, inFlight[key] == nil {
            return handle
        }
        if inFlight[key] != nil {
            if !handle.isStale { return handle }
            // A write marked the key while its load was in flight: that load
            // may have read before the write, so start over.
            bumpGeneration(key)
            inFlight[key]?.task.cancel()
            inFlight[key] = nil
        }

        let captured = bumpGeneration(key)
        let hadStaleValue = handle.value != nil
        handle.beginLoading(keepingValue: hadStaleValue)

        let task = Task { @MainActor in
            let outcome: Result<Value, Error>
            do {
                outcome = .success(try await loader())
            } catch {
                outcome = .failure(error)
            }
            if self.generations[key] != captured {
                return
            }
            switch outcome {
            case .success(let result):
                handle.applySuccess(result)
            case .failure(let error):
                handle.applyFailure(error, clearValue: !hadStaleValue)
            }
            if self.inFlight[key]?.generation == captured {
                self.inFlight[key] = nil
            }
        }
        inFlight[key] = (captured, task)
        return handle
    }

    func invalidate(_ key: CatalogQueryKey) {
        bumpGeneration(key)
        inFlight[key]?.task.cancel()
        inFlight[key] = nil
        handles[key]?.markStale()
    }

    /// Drops a key: its in-flight load is cancelled and its handle forgotten,
    /// so the next `query` starts fresh.
    private func evict(_ key: CatalogQueryKey) {
        bumpGeneration(key)
        inFlight[key]?.task.cancel()
        inFlight[key] = nil
        handles[key] = nil
        valueTypes[key] = nil
    }

    @discardableResult
    private func bumpGeneration(_ key: CatalogQueryKey) -> Int {
        let next = (generations[key] ?? 0) + 1
        generations[key] = next
        return next
    }

    /// Starts the registry loader when this key is already cached.
    /// Unwarmed keys load on the next `apply(location:)` / `query`.
    private func revalidate(_ key: CatalogQueryKey) {
        guard handles[key] != nil else { return }
        key.kind.descriptor.warm(self, key)
    }

    func invalidateAll(matching predicate: (CatalogQueryKey) -> Bool) {
        for key in handles.keys where predicate(key) {
            invalidate(key)
        }
    }

    private func typedHandle<Value>(for key: CatalogQueryKey, as type: Value.Type) -> QueryHandle<Value> {
        let typeID = ObjectIdentifier(type)
        if let existing = valueTypes[key], existing != typeID {
            #if DEBUG
            assertionFailure("CatalogQueryKey \(key) already bound to a different Value type")
            #endif
        }
        valueTypes[key] = typeID

        if let existing = handles[key] as? QueryHandle<Value> {
            return existing
        }
        let handle = QueryHandle<Value>()
        handles[key] = handle
        return handle
    }
}

/// What `WorkspaceSession.patchCached` did.
enum CatalogPatchResult: Sendable, Equatable {
    /// The cached value changed.
    case applied
    /// Nothing to change: no handle, no value, or the patch changed nothing.
    case unchanged
    /// The key's first load is in flight and may have read before the write.
    case awaitingFirstValue
}
