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
    /// Evidence-graph pan and zoom, keyed by Source id. Not catalog data; quit drops it.
    @ObservationIgnored
    private var graphCameras: [String: GraphCanvasCamera] = [:]
    /// Bridge to select the next time that Source's graph appears. Consumed once.
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

    func graphCamera(for sourceID: String) -> GraphCanvasCamera? {
        graphCameras[sourceID]
    }

    func setGraphCamera(_ camera: GraphCanvasCamera, for sourceID: String) {
        guard graphCameras[sourceID] != camera else { return }
        graphCameras[sourceID] = camera
    }

    func setPendingGraphSelection(subjectID: String, sourceID: String) {
        pendingGraphSelections[sourceID] = subjectID
    }

    func consumePendingGraphSelection(sourceID: String) -> String? {
        let id = pendingGraphSelections[sourceID]
        pendingGraphSelections[sourceID] = nil
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
        ensureQuery(key) { [store, registry] in
            let result = try await registry.load(key: key, store: store)
            guard let typed = result as? Value else {
                #if DEBUG
                assertionFailure("CatalogQueryKey \(key) value type mismatch for \(Value.self)")
                #endif
                throw CatalogQueryRegistryError.typeMismatch
            }
            return typed
        }
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
            warmQuery(key)
        }
    }

    /// Patches the list row + cached page shell on identity/cover save, then
    /// invalidates and revalidates whatever else the registry says the write
    /// reached. Keys the registry evicts when hidden are dropped instead
    /// unless on screen. Features only call `apply`; they do not query
    /// afterward or tell sibling views to refresh.
    /// Drops the cached reads named by a write's effects: that Source's graph
    /// and citation counts, and only the conclusion pages whose handles changed.
    func apply(_ effects: CatalogWriteEffects) {
        patchLastUpdated(effects)
        apply(invalidations: registry.invalidations(for: effects, project: projectKey))
        for sourceID in effects.touchedSources {
            refreshSourceGraphProgress(sourceId: sourceID)
        }
    }

    func apply(_ mutation: CatalogMutation) {
        if let source = mutation.patchedSource {
            patchUpdatedSource(source)
        }
        apply(invalidations: registry.invalidations(by: mutation, project: projectKey))
        refreshSourceGraphProgress(for: mutation)
    }

    /// A write that reports effects and still has a mutation naming keys the
    /// effects cannot (the Source page's own workspace, the metadata field
    /// list). One application, so a key both name is invalidated once, not
    /// cancelled and restarted twice.
    func apply(_ effects: CatalogWriteEffects, _ mutation: CatalogMutation) {
        patchLastUpdated(effects)
        var seen = Set<CatalogQueryInvalidation>()
        let union = (
            registry.invalidations(for: effects, project: projectKey)
                + registry.invalidations(by: mutation, project: projectKey)
        ).filter { seen.insert($0).inserted }
        apply(invalidations: union)
        if let source = mutation.patchedSource {
            patchUpdatedSource(source)
        }
        var progressSources = effects.touchedSources
        if let sourceID = graphProgressSource(of: mutation), !progressSources.contains(sourceID) {
            progressSources.append(sourceID)
        }
        for sourceID in progressSources {
            refreshSourceGraphProgress(sourceId: sourceID)
        }
    }

    /// Moves the last-updated cell of every cached row the write reports, using
    /// the write's own activity. Each list declares its patch in the registry.
    /// No loader runs.
    private func patchLastUpdated(_ effects: CatalogWriteEffects) {
        guard effects.activity.revision > 0 else { return }
        for key in Array(handles.keys) where key.project == projectKey {
            guard let patch = registry.lastUpdatedPatch(key) else { continue }
            let ids = Set(patch.scope.ids(in: effects))
            if !ids.isEmpty {
                patch.apply(self, key, ids, effects.activity)
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

    /// The Source whose graph progress a mutation moved, if it moved one.
    private func graphProgressSource(of mutation: CatalogMutation) -> String? {
        switch mutation {
        case .mutatedSourceGraph(let id), .deletedSubject(let id):
            return id
        default:
            return nil
        }
    }

    /// Surgical Get-one + map merge. Skips when the list never warmed the map.
    private func refreshSourceGraphProgress(for mutation: CatalogMutation) {
        if let sourceId = graphProgressSource(of: mutation) {
            refreshSourceGraphProgress(sourceId: sourceId)
        }
    }

    private func refreshSourceGraphProgress(sourceId: String) {
        let key = CatalogQueryKey.sourceGraphProgress(project: projectKey)
        let existing: QueryHandle<[String: SourceGraphProgress]>? = queryHandle(key)
        guard existing != nil else { return }
        Task { await self.patchSourceGraphProgress(sourceId: sourceId) }
    }

    private func patchSourceGraphProgress(sourceId: String) async {
        let key = CatalogQueryKey.sourceGraphProgress(project: projectKey)
        let existing: QueryHandle<[String: SourceGraphProgress]>? = queryHandle(key)
        guard existing != nil else { return }
        do {
            let one = try await store.getSourceGraphProgress(
                projectDir: projectKey.projectDir,
                sourceID: sourceId
            )
            let handle: QueryHandle<[String: SourceGraphProgress]>? = queryHandle(key)
            var map = handle?.value ?? [:]
            if one.isZero {
                map.removeValue(forKey: sourceId)
            } else {
                map[sourceId] = one
            }
            setQueryValue(key, value: map)
        } catch {
            // Keep last counts; the next mutation or Sources place visit retries.
        }
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
        warmQuery(key)
    }

    func invalidateAll(matching predicate: (CatalogQueryKey) -> Bool) {
        for key in handles.keys where predicate(key) {
            invalidate(key)
        }
    }

    private func warmQuery(_ key: CatalogQueryKey) {
        switch key {
        case .sourcesList:
            let _: QueryHandle<[CatalogSource]> = query(key)
        case .sourceTypesList:
            let _: QueryHandle<[CatalogSourceType]> = query(key)
        case .metadataFieldsList:
            let _: QueryHandle<[CatalogMetadataField]> = query(key)
        case .credibilityGradesList:
            let _: QueryHandle<[CatalogCredibilityGrade]> = query(key)
        case .confidenceGradesList:
            let _: QueryHandle<[CatalogClaimConfidenceGrade]> = query(key)
        case .sourceWorkspace:
            let _: QueryHandle<CatalogSourceWorkspace> = query(key)
        case .typeSuggestions:
            let _: QueryHandle<[CatalogTypeSuggestion]> = query(key)
        case .sourceGraph:
            let _: QueryHandle<SourceGraphRows> = query(key)
        case .citationCounts:
            let _: QueryHandle<[String: Int]> = query(key)
        case .propertiesWorkspace:
            let _: QueryHandle<PropertiesSnapshot> = query(key)
        case .connectRules:
            let _: QueryHandle<[CatalogConnectRule]> = query(key)
        case .propertyTerms:
            let _: QueryHandle<[CatalogPropertyTerm]> = query(key)
        case .citationsByArtifact:
            let _: QueryHandle<[CatalogListedCitation]> = query(key)
        case .sourceGraphProgress:
            let _: QueryHandle<[String: SourceGraphProgress]> = query(key)
        case .personsList:
            let _: QueryHandle<[CatalogPersonHeader]> = query(key)
        case .eventsList:
            let _: QueryHandle<[CatalogEventHeader]> = query(key)
        case .placesList:
            let _: QueryHandle<[CatalogPlaceHeader]> = query(key)
        case .conclusionDetail:
            let _: QueryHandle<CatalogConclusionDetail> = query(key)
        }
    }

    private func patchUpdatedSource(_ source: CatalogSource) {
        let listKey = CatalogQueryKey.sourcesList(project: projectKey)
        // Leaves any stale flag on the list: other rows may still owe a reload.
        patchCached(listKey, as: [CatalogSource].self) { sources in
            guard let index = sources.firstIndex(where: { $0.id == source.id }) else { return false }
            sources[index] = source
            return true
        }

        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: projectKey, sourceId: source.id)
        let patched = patchCached(workspaceKey, as: CatalogSourceWorkspace.self) {
            $0.source = source
            return true
        }
        if patched { invalidatedKeys.remove(workspaceKey) }  // that page is now current
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
