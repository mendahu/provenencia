import Foundation

/// Workspace catalog write that may patch or bust cached queries.
enum CatalogMutation: Sendable, Equatable {
    /// Patch path: update list row + cached workspace shell synchronously.
    case updatedSource(CatalogSource)

    case createdSource

    case createdSourceType
    case updatedSourceType(id: String)
    case deletedSourceType(id: String)

    case createdMetadataField
    case updatedMetadataField(id: String)
    case deletedMetadataField(id: String)

    case assignedTypeSuggestion(typeId: String)
    case removedTypeSuggestion(typeId: String)

    /// Bust-only: non-identity source page edits (metadata, notes, artifacts).
    case mutatedSourceWorkspace(sourceId: String)
}

/// Mutation kind for registry invalidation tags (no associated payload).
enum CatalogMutationKind: Hashable, Sendable {
    case createdSource
    case createdSourceType
    case updatedSourceType
    case deletedSourceType
    case createdMetadataField
    case updatedMetadataField
    case deletedMetadataField
    case assignedTypeSuggestion
    case removedTypeSuggestion
    case mutatedSourceWorkspace
}

extension CatalogMutation {
    /// Kind tag for registry `invalidateOn` lookup. Nil for patch-only mutations.
    var invalidationKind: CatalogMutationKind? {
        switch self {
        case .updatedSource:
            return nil
        case .createdSource:
            return .createdSource
        case .createdSourceType:
            return .createdSourceType
        case .updatedSourceType:
            return .updatedSourceType
        case .deletedSourceType:
            return .deletedSourceType
        case .createdMetadataField:
            return .createdMetadataField
        case .updatedMetadataField:
            return .updatedMetadataField
        case .deletedMetadataField:
            return .deletedMetadataField
        case .assignedTypeSuggestion:
            return .assignedTypeSuggestion
        case .removedTypeSuggestion:
            return .removedTypeSuggestion
        case .mutatedSourceWorkspace:
            return .mutatedSourceWorkspace
        }
    }
}
