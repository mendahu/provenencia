import Foundation
import Observation

/// Project-scoped catalog query cache. S4-02 adds registry-backed loaders;
/// S4-04 wires this into workspace navigation.
@MainActor
@Observable
final class WorkspaceSession {
    let projectKey: ProjectKey
    let store: any GenealogyStore
    let registry: CatalogQueryRegistry
    /// Default name-format orders. Empty until the catalog read finishes,
    /// which keeps stored part order.
    var nameFormat = NameFormatOrders()

    private var handles: [CatalogQueryKey: AnyObject] = [:]
    private var valueTypes: [CatalogQueryKey: ObjectIdentifier] = [:]
    private var generations: [CatalogQueryKey: Int] = [:]
    private var inFlight: [CatalogQueryKey: (generation: Int, task: Task<Void, Never>)] = [:]
    private var invalidatedKeys: Set<CatalogQueryKey> = []
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

    /// Waits out an in-flight load, then returns the cached value.
    /// Does not start a load. Nil when the key was never warmed, is stale, or the load failed.
    func readyValue<Value>(_ key: CatalogQueryKey) async -> Value? {
        while let task = inFlight[key]?.task {
            await task.value
        }
        if invalidatedKeys.contains(key) {
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
            invalidatedKeys.remove(key)
        }
        if hadInFlight {
            invalidatedKeys.insert(key)
            revalidate(key)
        }
    }

    /// Edits the cached value in place. `change` returns whether it changed
    /// anything; nothing happens when it did not or the key has no value. The
    /// stale flag is left as it was: a partial patch does not make the key current.
    /// Returns whether a patch applied.
    @discardableResult
    func patchCached<Value>(
        _ key: CatalogQueryKey,
        as _: Value.Type,
        _ change: (inout Value) -> Bool
    ) -> Bool {
        guard let handle: QueryHandle<Value> = queryHandle(key), var value = handle.value else { return false }
        guard change(&value) else { return false }
        replaceValue(key, value, clearStale: false)
        return true
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
    /// Then it invalidates the deduped union of what the effects
    /// reported (each Evidence graph it names, its Sources' citation reads, the Conclusions whose
    /// handles changed) and what the mutation names (the Source page's own
    /// workspace, the metadata field list), so a key both name is invalidated
    /// once, not cancelled and restarted twice. Keys the registry evicts or
    /// marks stale when hidden are not loaded unless on screen. Last, it
    /// refreshes graph progress for the Evidence graphs the effects and the mutation
    /// moved, and runs `onNavCountsMoved` for a mutation in
    /// `CatalogQueryRegistry.navCountTriggers`. Features only call `apply`;
    /// they do not query afterward or tell sibling views to refresh.
    func apply(_ effects: CatalogWriteEffects = .empty, _ mutation: CatalogMutation? = nil) {
        applyPatches(effects, mutation)
        apply(invalidations: registry.invalidations(for: effects, by: mutation, project: projectKey))
        for key in Array(handles.keys) where key.project == projectKey {
            guard let refresh = key.kind.descriptor.refresh else { continue }
            let ids = refresh.follows.ids(in: effects, mutation: mutation)
            if !ids.isEmpty {
                Task { await refresh.run(self, key, ids) }
            }
        }
        if let kind = mutation?.invalidationKind,
           CatalogQueryRegistry.navCountTriggers.contains(kind),
           let recount = onNavCountsMoved
        {
            Task { await recount() }
        }
    }

    /// A mutation with no reported effects.
    func apply(_ mutation: CatalogMutation) {
        apply(.empty, mutation)
    }

    /// Runs each warmed key's descriptor patches. A patch with `follows` gets
    /// the ids that slot names, and only when the write reported activity.
    /// A patch with no `follows` reads the mutation.
    private func applyPatches(_ effects: CatalogWriteEffects, _ mutation: CatalogMutation?) {
        for key in Array(handles.keys) where key.project == projectKey {
            for patch in key.kind.descriptor.patches {
                if let follows = patch.follows {
                    guard effects.activity.revision > 0 else { continue }
                    let ids = Set(follows.ids(in: effects, mutation: mutation))
                    if !ids.isEmpty {
                        patch.apply(self, key, ids, effects.activity, mutation)
                    }
                } else {
                    patch.apply(self, key, [], effects.activity, mutation)
                }
            }
        }
    }

    private func apply(invalidations: [CatalogQueryInvalidation]) {
        var keys: [CatalogQueryKey] = []
        for key in cachedKeys(for: invalidations) {
            switch visibleKeys.contains(key) ? CatalogHiddenPolicy.revalidate : registry.hiddenPolicy(key) {
            case .evict:
                evict(key)
            case .markStale:
                // The old value stays; the next `query` reloads it.
                if handles[key] != nil { invalidatedKeys.insert(key) }
            case .revalidate:
                keys.append(key)
            }
        }
        for key in keys {
            invalidate(key)
        }
        for key in keys {
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

        let isStale = invalidatedKeys.contains(key)
        if handle.status == .ready, !isStale, inFlight[key] == nil {
            return handle
        }
        if inFlight[key] != nil {
            return handle
        }

        invalidatedKeys.remove(key)
        let captured = bumpGeneration(key)
        let hadStaleValue = handle.value != nil
        handle.beginLoading(stale: hadStaleValue)

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
        invalidatedKeys.insert(key)
    }

    /// Drops a key: its in-flight load is cancelled and its handle forgotten,
    /// so the next `query` starts fresh.
    private func evict(_ key: CatalogQueryKey) {
        bumpGeneration(key)
        inFlight[key]?.task.cancel()
        inFlight[key] = nil
        invalidatedKeys.remove(key)
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
