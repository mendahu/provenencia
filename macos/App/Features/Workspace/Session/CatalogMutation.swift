import Foundation

/// Workspace catalog write that may patch or bust cached queries.
enum CatalogMutation: Sendable, Equatable {
    /// Patch path: update list row + cached workspace shell synchronously.
    case updatedSource(CatalogSource)

    /// Patch the same shell as `updatedSource`, then bust: the engine derives a
    /// source's suggested metadata rows from its type, and counts sources per
    /// type, so neither survives the move.
    case changedSourceType(CatalogSource)

    case createdSource

    case createdSourceType
    case updatedSourceType(id: String)
    case deletedSourceType(id: String)

    case createdMetadataField
    case updatedMetadataField(id: String)
    case deletedMetadataField(id: String)

    case assignedTypeSuggestion(typeId: String)
    case removedTypeSuggestion(typeId: String)

    /// Bust-only: non-identity source page edits (notes, artifacts, credibility).
    case mutatedSourceWorkspace(sourceId: String)

    /// Metadata values for one source changed. Narrower than
    /// `mutatedSourceWorkspace` because it also moves each field's `usedBy`.
    case mutatedSourceMetadata(sourceId: String)

    /// A primary subject was created on this Source's Evidence graph.
    case createdSubject(sourceId: String)

    case createdProperty
    case updatedProperty
    case deletedProperty
    case mutatedSubjectTypeFields
}

/// Mutation kind for registry invalidation tags (no associated payload).
enum CatalogMutationKind: Hashable, Sendable {
    case changedSourceType
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
    case mutatedSourceMetadata
    case createdSubject
    case createdProperty
    case updatedProperty
    case deletedProperty
    case mutatedSubjectTypeFields
}

extension CatalogMutation {
    /// Kind tag for registry `invalidateOn` lookup. Nil for patch-only mutations.
    var invalidationKind: CatalogMutationKind? {
        switch self {
        case .updatedSource:
            return nil
        case .changedSourceType:
            return .changedSourceType
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
        case .mutatedSourceMetadata:
            return .mutatedSourceMetadata
        case .createdSubject:
            return .createdSubject
        case .createdProperty:
            return .createdProperty
        case .updatedProperty:
            return .updatedProperty
        case .deletedProperty:
            return .deletedProperty
        case .mutatedSubjectTypeFields:
            return .mutatedSubjectTypeFields
        }
    }

    /// Enriched Source to patch into the cached list row and page shell, if any.
    var patchedSource: CatalogSource? {
        switch self {
        case .updatedSource(let source), .changedSourceType(let source):
            return source
        default:
            return nil
        }
    }
}
