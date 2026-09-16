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
    private var inFlight: [CatalogQueryKey: Task<Void, Never>] = [:]
    private var invalidatedKeys: Set<CatalogQueryKey> = []

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
    func setQueryValue<Value>(_ key: CatalogQueryKey, value: Value) {
        let handle = typedHandle(for: key, as: Value.self)
        handle.applySuccess(value)
        invalidatedKeys.remove(key)
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
    /// busts whatever else the registry says the write reached. Patching first
    /// keeps the edit on screen while a busted page revalidates.
    func apply(_ mutation: CatalogMutation) {
        if let source = mutation.patchedSource {
            patchUpdatedSource(source)
        }
        for invalidation in registry.invalidations(by: mutation, project: projectKey) {
            switch invalidation {
            case .key(let key):
                invalidate(key)
            case .allCached(let kind):
                invalidateAll { $0.kind == kind && $0.project == projectKey }
            }
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
        let hadStaleValue = handle.value != nil
        handle.beginLoading(stale: hadStaleValue)

        let task = Task { @MainActor in
            defer { self.inFlight[key] = nil }
            do {
                let result = try await loader()
                handle.applySuccess(result)
            } catch {
                handle.applyFailure(error, clearValue: !hadStaleValue)
            }
        }
        inFlight[key] = task
        return handle
    }

    func invalidate(_ key: CatalogQueryKey) {
        inFlight[key]?.cancel()
        inFlight[key] = nil
        invalidatedKeys.insert(key)
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
        case .sourceWorkspace:
            let _: QueryHandle<CatalogSourceWorkspace> = query(key)
        case .typeSuggestions:
            let _: QueryHandle<[CatalogTypeSuggestion]> = query(key)
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
