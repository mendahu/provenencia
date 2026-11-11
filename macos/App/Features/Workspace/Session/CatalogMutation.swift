import Foundation

/// Workspace catalog write that may patch or bust cached queries.
enum CatalogMutation: Sendable, Equatable {
    /// Patch path: update list row + cached workspace shell synchronously.
    case updatedSource(CatalogSource)

    /// Patch the same shell as `updatedSource`, then bust: the engine derives a
    /// source's suggested metadata rows from its type, and counts sources per
    /// type, so neither survives the move.
    case changedSourceType(CatalogSource)

    /// Add or Copy Source: the row the engine returned, which the Sources list
    /// takes as a new row (`CatalogQueryPatch.sourceRow()`).
    case createdSource(CatalogSource)
    case deletedSource(id: String)

    /// A field was relabelled or deleted. Each Source page restates the
    /// field's derived rows, which `source_vocabulary` cannot tell from a field
    /// create, so these two name the Source pages.
    case updatedMetadataField(id: String)
    case deletedMetadataField(id: String)

    /// The type–field join is not audited and reports no effects, so these
    /// name the type's suggestions key.
    case assignedTypeSuggestion(typeId: String)
    case removedTypeSuggestion(typeId: String)

    /// Bust-only: non-identity source page edits (notes, artifacts, credibility).
    case mutatedSourceWorkspace(sourceId: String)

    /// Metadata values for one source changed. Narrower than
    /// `mutatedSourceWorkspace` because it also moves each field's `usedBy`.
    case mutatedSourceMetadata(sourceId: String)
}

/// A catalog-wide list a write changed. The raw values are Go's
/// `effects.List` wire keys (`WriteEffects.lists`); a key this build does not
/// know is dropped when the effects are decoded.
enum CatalogList: String, Sendable, CaseIterable {
    /// A Source type or a metadata field changed.
    case sourceVocabulary = "source_vocabulary"
    /// A subject type, Property or Property term label, or a subject type's
    /// Property binding, changed.
    case propertyVocabulary = "property_vocabulary"
    /// An Observation was created, deleted, or moved to another Property, so
    /// the Properties workspace's `usedBy` counts changed.
    case propertyUsage = "property_usage"
    /// A Source or a handle (of any type) was created, deleted or merged, so a
    /// sidebar count may have changed. It names no cache key: the session
    /// recounts (`WorkspaceSession.onNavCountsMoved`).
    case counts = "counts"
}

/// A keyed partial read a write moved, per id of the read's key table
/// (Evidence graphs for `citations` and `evidence_graph_progress`, Sources for
/// `evidence_graphs`). The raw values are
/// Go's `effects.Read` wire keys (`WriteEffects.reads`); a key this build does
/// not know is dropped when the effects are decoded.
enum CatalogRead: String, Sendable, CaseIterable {
    /// The citation counts and the citations by Artifact of a graph.
    case citations = "citations"
    /// The Subject, filed and Observation counts of a graph.
    case evidenceGraphProgress = "evidence_graph_progress"
    /// The graphs of a Source with their linked Artifacts (the graph list).
    case evidenceGraphs = "evidence_graphs"
    /// The terms of a Property (the term list).
    case propertyTerms = "property_terms"
}

/// One keyed read and the key ids (Evidence graphs, Sources for
/// `evidence_graphs`, Properties for `property_terms`) it moved for.
struct CatalogWriteEffectsRead: Sendable, Equatable {
    var read: CatalogRead
    var ids: [String]
}

/// An entity whose deleted ids a write reports. The raw values are Go's
/// `effects.DeletedEntities` wire keys (`WriteEffects.deleted`); a key this
/// build does not know is dropped when the effects are decoded.
enum CatalogDeletedEntity: String, Sendable, CaseIterable {
    /// Person, Event, Place and bridge handles.
    case handle = "canonical_entity"
}

/// The ids of one entity a write deleted.
struct CatalogWriteEffectsDeleted: Sendable, Equatable {
    var entity: CatalogDeletedEntity
    var ids: [String]
}

/// Handles and Sources a write reported. The session applies each cached
/// key's hidden policy to the reads the write moved (a visible key reloads, a
/// hidden one is marked stale), evicts the hidden keys of what `deleted`
/// names, and patches the rows whose last activity moved.
struct CatalogWriteEffects: Sendable, Equatable {
    var handles: [String] = []
    /// Evidence graphs whose loaded rows changed (graphcache drops each). A
    /// note, metadata, or credibility change is only in `sourceScopes`.
    var evidenceGraphs: [String] = []
    /// Keyed partial reads this write moved, each with the key ids it moved
    /// them for: citation counts, citations by Artifact and graph progress.
    var reads: [CatalogWriteEffectsRead] = []
    /// Catalog-wide lists this write changed. A key that holds one follows it
    /// through its descriptor `refreshes`.
    var lists: Set<CatalogList> = []
    var structure = false
    var summaryHandles: [String] = []
    /// Handles whose last-updated time this write moved. The session patches
    /// those rows in the cached lists with `activity`; `touchedHandles` (detail
    /// pages) does not include them.
    var handleScopes: [String] = []
    /// Sources whose last-updated time this write moved. The session patches
    /// those rows in the cached Sources list with `activity`, and the detail
    /// keys whose records cite one of those Sources follow it through `.source`.
    var sourceScopes: [String] = []
    /// Ids this write deleted, by entity. `apply` evicts the hidden keys whose
    /// descriptor's `deletedWith` names the entity; a page that shows a deleted id applies the
    /// effects with `apply(_:leavingPage:)` (`UnpromoteFlow.confirm(session:userID:leavingPage:announce:)`).
    var deleted: [CatalogWriteEffectsDeleted] = []
    /// The write's own audit revision and time; revision 0 when it recorded none.
    var activity = CatalogActivity()

    static let empty = CatalogWriteEffects()

    /// The handles this write deleted: the `canonical_entity` entry of `deleted`.
    var deletedHandles: [String] {
        deleted.first { $0.entity == .handle }?.ids ?? []
    }

    var touchedHandles: [String] {
        var seen = Set<String>()
        return (handles + summaryHandles).filter { seen.insert($0).inserted && !$0.isEmpty }
    }
}

/// Mutation kind for registry invalidation tags (no associated payload).
enum CatalogMutationKind: Hashable, Sendable {
    case changedSourceType
    case createdSource
    case deletedSource
    case updatedMetadataField
    case deletedMetadataField
    case assignedTypeSuggestion
    case removedTypeSuggestion
    case mutatedSourceWorkspace
    case mutatedSourceMetadata
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
        case .deletedSource:
            return .deletedSource
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
        }
    }

    /// Enriched Source to patch into the cached list row and page shell, if any.
    var patchedSource: CatalogSource? {
        switch self {
        case .updatedSource(let source), .changedSourceType(let source), .createdSource(let source):
            return source
        default:
            return nil
        }
    }
}
