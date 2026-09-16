import Foundation

enum CatalogQueryRegistryError: Error {
    case typeMismatch
}

/// One cache effect of a mutation. Keys carrying an id resolve to `allCached`
/// when the write names no single owner — a new metadata field belongs to every
/// cached Source page, not one of them.
enum CatalogQueryInvalidation: Hashable, Sendable {
    case key(CatalogQueryKey)
    case allCached(CatalogQueryKey.Kind)
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
            // `usedBy` counts sources per type, so adding a source or moving one
            // restates rows no type edit touched.
            invalidateOn: [
                .createdSourceType, .updatedSourceType, .deletedSourceType,
                .createdSource, .changedSourceType,
            ]
        ),
        Spec(
            kind: .metadataFieldsList,
            stalePolicy: .sessionFresh,
            // `usedBy` counts sources holding a value for the field, and it
            // gates Delete.
            invalidateOn: [
                .createdMetadataField, .updatedMetadataField, .deletedMetadataField,
                .mutatedSourceMetadata,
            ]
        ),
        Spec(
            kind: .sourceWorkspace,
            stalePolicy: .sessionFresh,
            // The page payload folds in the whole type and field vocabulary
            // (type picker, Add-metadata list) plus the type's suggested rows,
            // so vocabulary edits stale every cached page, not just one.
            invalidateOn: [
                .mutatedSourceWorkspace, .mutatedSourceMetadata, .changedSourceType,
                .createdSourceType, .updatedSourceType, .deletedSourceType,
                .createdMetadataField, .updatedMetadataField, .deletedMetadataField,
                .assignedTypeSuggestion, .removedTypeSuggestion,
            ]
        ),
        Spec(
            kind: .typeSuggestions,
            stalePolicy: .sessionFresh,
            // Each suggestion embeds a full field row, so relabelling or
            // deleting a field reaches types the edit never named.
            invalidateOn: [
                .assignedTypeSuggestion, .removedTypeSuggestion,
                .updatedMetadataField, .deletedMetadataField,
            ]
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

    /// Resolves cache effects from registry `invalidateOn` tags + mutation payload.
    func invalidations(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryInvalidation] {
        guard let mutationKind = mutation.invalidationKind else { return [] }
        return specs
            .filter { $0.invalidateOn.contains(mutationKind) }
            .map { $0.kind.invalidation(project: project, mutation: mutation) }
    }
}

private extension CatalogQueryKey.Kind {
    /// Builds the cache effect for one registry row and mutation. Id-bearing
    /// kinds fall back to `allCached` when the mutation names no single owner.
    func invalidation(project: ProjectKey, mutation: CatalogMutation) -> CatalogQueryInvalidation {
        switch self {
        case .sourcesList:
            return .key(.sourcesList(project: project))
        case .sourceTypesList:
            return .key(.sourceTypesList(project: project))
        case .metadataFieldsList:
            return .key(.metadataFieldsList(project: project))
        case .sourceWorkspace:
            switch mutation {
            case .mutatedSourceWorkspace(let sourceId), .mutatedSourceMetadata(let sourceId):
                return .key(.sourceWorkspace(project: project, sourceId: sourceId))
            case .changedSourceType(let source):
                return .key(.sourceWorkspace(project: project, sourceId: source.id))
            default:
                return .allCached(.sourceWorkspace)
            }
        case .typeSuggestions:
            switch mutation {
            case .assignedTypeSuggestion(let typeId), .removedTypeSuggestion(let typeId):
                return .key(.typeSuggestions(project: project, typeId: typeId))
            default:
                return .allCached(.typeSuggestions)
            }
        }
    }
}
