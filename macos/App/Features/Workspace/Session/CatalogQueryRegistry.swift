import Foundation

enum CatalogQueryRegistryError: Error {
    case typeMismatch
}

/// Declarative catalog read loaders and mutation invalidation map.
struct CatalogQueryRegistry: Sendable {
    static let standard = CatalogQueryRegistry()

    private struct Spec {
        let kind: CatalogQueryKey.Kind
        let stalePolicy: CatalogQueryStalePolicy
        let invalidateOn: Set<CatalogMutationKind>
    }

    private let specs: [Spec] = [
        Spec(
            kind: .sourcesList,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdSource]
        ),
        Spec(
            kind: .sourceTypesList,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdSourceType, .updatedSourceType, .deletedSourceType]
        ),
        Spec(
            kind: .metadataFieldsList,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdMetadataField, .updatedMetadataField, .deletedMetadataField]
        ),
        Spec(
            kind: .sourceWorkspace,
            stalePolicy: .sessionFresh,
            invalidateOn: [.mutatedSourceWorkspace]
        ),
        Spec(
            kind: .typeSuggestions,
            stalePolicy: .sessionFresh,
            invalidateOn: [.assignedTypeSuggestion, .removedTypeSuggestion]
        ),
    ]

    func stalePolicy(for key: CatalogQueryKey) -> CatalogQueryStalePolicy {
        specs.first { $0.kind == key.kind }?.stalePolicy ?? .sessionFresh
    }

    func load(key: CatalogQueryKey, store: any GenealogyStore) async throws -> Any {
        switch key {
        case .sourcesList(let project):
            return try await store.listSources(projectDir: project.projectDir)
        case .sourceTypesList(let project):
            return try await store.listSourceTypes(projectDir: project.projectDir)
        case .metadataFieldsList(let project):
            return try await store.listMetadataFields(projectDir: project.projectDir)
        case .sourceWorkspace(let project, let sourceId):
            return try await store.getSourceWorkspace(projectDir: project.projectDir, sourceID: sourceId)
        case .typeSuggestions(let project, let typeId):
            return try await store.listTypeSuggestions(projectDir: project.projectDir, typeID: typeId)
        }
    }

    /// Resolves concrete cache keys from registry `invalidateOn` tags + mutation payload.
    func keysAffected(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryKey] {
        guard let mutationKind = mutation.invalidationKind else { return [] }
        return specs.compactMap { spec in
            guard spec.invalidateOn.contains(mutationKind) else { return nil }
            return spec.kind.cacheKey(project: project, mutation: mutation)
        }
    }
}

private extension CatalogQueryKey.Kind {
    /// Builds the concrete cache key for one registry row and mutation.
    func cacheKey(project: ProjectKey, mutation: CatalogMutation) -> CatalogQueryKey? {
        switch self {
        case .sourcesList:
            return .sourcesList(project: project)
        case .sourceTypesList:
            return .sourceTypesList(project: project)
        case .metadataFieldsList:
            return .metadataFieldsList(project: project)
        case .sourceWorkspace:
            guard case .mutatedSourceWorkspace(let sourceId) = mutation else { return nil }
            return .sourceWorkspace(project: project, sourceId: sourceId)
        case .typeSuggestions:
            switch mutation {
            case .assignedTypeSuggestion(let typeId), .removedTypeSuggestion(let typeId):
                return .typeSuggestions(project: project, typeId: typeId)
            default:
                return nil
            }
        }
    }
}
