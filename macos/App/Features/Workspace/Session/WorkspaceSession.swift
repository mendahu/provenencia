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

    private var handles: [CatalogQueryKey: AnyObject] = [:]
    private var valueTypes: [CatalogQueryKey: ObjectIdentifier] = [:]
    private var generations: [CatalogQueryKey: Int] = [:]
    private var inFlight: [CatalogQueryKey: (generation: Int, task: Task<Void, Never>)] = [:]
    private var invalidatedKeys: Set<CatalogQueryKey> = []
    /// Transient notice after leaving a place (composer save, etc.).
    var noticeToast: VocabularyToast?

    init(
        projectKey: ProjectKey,
        store: any GenealogyStore,
        registry: CatalogQueryRegistry = .standard
    ) {
        self.projectKey = projectKey
        self.store = store
        self.registry = registry
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
        let hadInFlight = inFlight[key] != nil
        bumpGeneration(key)
        inFlight[key]?.task.cancel()
        inFlight[key] = nil
        let handle = typedHandle(for: key, as: Value.self)
        handle.applySuccess(value)
        invalidatedKeys.remove(key)
        if hadInFlight {
            invalidatedKeys.insert(key)
            revalidate(key)
        }
    }

    /// Resolves the place for `location` and starts loading its query keys (non-blocking).
    func apply(
        location: WorkspaceLocation,
        placeRegistry: PlaceRegistry = .standard
    ) {
        guard let place = placeRegistry.resolve(location, project: projectKey) else { return }
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
    /// reached. Features only call `apply`; they do not query afterward or
    /// tell sibling views to refresh.
    func apply(_ mutation: CatalogMutation) {
        if let source = mutation.patchedSource {
            patchUpdatedSource(source)
        }
        var keys: [CatalogQueryKey] = []
        for invalidation in registry.invalidations(by: mutation, project: projectKey) {
            switch invalidation {
            case .key(let key):
                keys.append(key)
            case .allCached(let kind):
                keys.append(contentsOf: handles.keys.filter {
                    $0.kind == kind && $0.project == projectKey
                })
            }
        }
        for key in keys {
            invalidate(key)
        }
        for key in keys {
            revalidate(key)
        }
        refreshSourceGraphProgress(for: mutation)
    }

    /// Surgical Get-one + map merge. Skips when the list never warmed the map.
    private func refreshSourceGraphProgress(for mutation: CatalogMutation) {
        let sourceId: String
        switch mutation {
        case .mutatedSourceGraph(let id), .savedCitation(let id):
            sourceId = id
        default:
            return
        }
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
        case .sourceWorkspace:
            let _: QueryHandle<CatalogSourceWorkspace> = query(key)
        case .typeSuggestions:
            let _: QueryHandle<[CatalogTypeSuggestion]> = query(key)
        case .sourceGraph:
            let _: QueryHandle<SourceGraphRows> = query(key)
        case .citationCounts:
            let _: QueryHandle<[String: Int]> = query(key)
        case .subjectFieldsWorkspace:
            let _: QueryHandle<SubjectFieldsSnapshot> = query(key)
        case .connectRules:
            let _: QueryHandle<[CatalogConnectRule]> = query(key)
        case .propertyTerms:
            let _: QueryHandle<[CatalogPropertyTerm]> = query(key)
        case .citationsByArtifact:
            let _: QueryHandle<[CatalogListedCitation]> = query(key)
        case .sourceGraphProgress:
            let _: QueryHandle<[String: SourceGraphProgress]> = query(key)
        }
    }

    private func patchUpdatedSource(_ source: CatalogSource) {
        let listKey = CatalogQueryKey.sourcesList(project: projectKey)
        if let listHandle: QueryHandle<[CatalogSource]> = queryHandle(listKey),
           var sources = listHandle.value,
           let index = sources.firstIndex(where: { $0.id == source.id }) {
            sources[index] = source
            listHandle.applySuccess(sources)
            invalidatedKeys.remove(listKey)
        }

        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: projectKey, sourceId: source.id)
        if let workspaceHandle: QueryHandle<CatalogSourceWorkspace> = queryHandle(workspaceKey),
           var workspace = workspaceHandle.value {
            workspace.source = source
            workspaceHandle.applySuccess(workspace)
            invalidatedKeys.remove(workspaceKey)
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
