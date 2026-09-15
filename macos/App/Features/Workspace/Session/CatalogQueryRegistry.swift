import Foundation

enum CatalogQueryRegistryError: Error {
    case typeMismatch
}

/// Declarative catalog read loaders and mutation invalidation map.
struct CatalogQueryRegistry: Sendable {
    static let standard = CatalogQueryRegistry()

    private enum KeyKind {
        case sourcesList
        case sourceTypesList
        case metadataFieldsList
        case sourceWorkspace
        case typeSuggestions
    }

    private struct Spec {
        let kind: KeyKind
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
        spec(for: key).stalePolicy
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

    func keysAffected(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryKey] {
        switch mutation {
        case .updatedSource:
            return []
        case .createdSource:
            return [.sourcesList(project: project)]
        case .createdSourceType, .updatedSourceType, .deletedSourceType:
            return [.sourceTypesList(project: project)]
        case .createdMetadataField, .updatedMetadataField, .deletedMetadataField:
            return [.metadataFieldsList(project: project)]
        case .assignedTypeSuggestion(let typeId), .removedTypeSuggestion(let typeId):
            return [.typeSuggestions(project: project, typeId: typeId)]
        case .mutatedSourceWorkspace(let sourceId):
            return [.sourceWorkspace(project: project, sourceId: sourceId)]
        }
    }

    private func spec(for key: CatalogQueryKey) -> Spec {
        let kind = keyKind(for: key)
        return specs.first { $0.kind == kind } ?? Spec(kind: kind, stalePolicy: .sessionFresh, invalidateOn: [])
    }

    private func keyKind(for key: CatalogQueryKey) -> KeyKind {
        switch key {
        case .sourcesList:
            return .sourcesList
        case .sourceTypesList:
            return .sourceTypesList
        case .metadataFieldsList:
            return .metadataFieldsList
        case .sourceWorkspace:
            return .sourceWorkspace
        case .typeSuggestions:
            return .typeSuggestions
        }
    }
}
