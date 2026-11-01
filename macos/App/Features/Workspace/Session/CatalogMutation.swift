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
    case deletedSource(id: String)

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

    /// A Subject (including a bridge) was erased. Releases connection facets
    /// so Observation `usedBy` must reload; the citation reads follow the
    /// write's effects.
    case deletedSubject

    /// A Subject on this Evidence graph was promoted (an Identity Claim filed it
    /// onto a handle), so its graph card's membership changed. The S9-04 card
    /// dispatches it after `promoteSubject` succeeds.
    case promotedSubject(graphId: String)

    case createdProperty
    case updatedProperty
    case deletedProperty
    case mutatedSubjectTypeProperties

    /// A term was added to one Property's vocabulary.
    case createdPropertyTerm(propertyId: String)
}

/// A catalog-wide list a write changed. The raw values are Go's
/// `effects.List` wire keys (`WriteEffects.lists`); a key this build does not
/// know is dropped when the effects are decoded.
enum CatalogList: String, Sendable, CaseIterable {
    /// A vocabulary label changed.
    case vocabulary = "vocabulary"
    /// An Observation was created, deleted, or moved to another Property, so
    /// the Properties workspace's `usedBy` counts changed.
    case propertyUsage = "property_usage"
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
}

/// One keyed read and the key ids (Evidence graphs, or Sources for
/// `evidence_graphs`) it moved for.
struct CatalogWriteEffectsRead: Sendable, Equatable {
    var read: CatalogRead
    var ids: [String]
}

/// Handles and Sources a write reported. The session drops those cached
/// reads and patches the rows whose last activity it moved.
struct CatalogWriteEffects: Sendable, Equatable {
    var handles: [String] = []
    /// Evidence graphs whose loaded rows changed (graphcache drops each). A
    /// note, metadata, or credibility change is only in `sourceScopes`.
    var evidenceGraphs: [String] = []
    /// Keyed partial reads this write moved, each with the key ids it moved
    /// them for: citation counts, citations by Artifact and graph progress.
    var reads: [CatalogWriteEffectsRead] = []
    /// Catalog-wide lists this write changed. Each is a cached read the write
    /// cannot patch row by row; see `CatalogQueryRegistry.invalidations(for:project:)`.
    var lists: Set<CatalogList> = []
    var structure = false
    var headerHandles: [String] = []
    /// Handles whose last-updated time this write moved. The session patches
    /// those rows in the cached lists with `activity`; `touchedHandles` (detail
    /// pages) does not include them.
    var handleScopes: [String] = []
    /// Sources whose last-updated time this write moved. The session patches
    /// those rows in the cached Sources list with `activity`; it reloads nothing.
    var sourceScopes: [String] = []
    /// The write's own audit revision and time; revision 0 when it recorded none.
    var activity = CatalogActivity()

    static let empty = CatalogWriteEffects()

    var touchedHandles: [String] {
        var seen = Set<String>()
        return (handles + headerHandles).filter { seen.insert($0).inserted && !$0.isEmpty }
    }
}

/// Mutation kind for registry invalidation tags (no associated payload).
enum CatalogMutationKind: Hashable, Sendable {
    case changedSourceType
    case createdSource
    case deletedSource
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
    case deletedSubject
    case promotedSubject
    case createdProperty
    case updatedProperty
    case deletedProperty
    case mutatedSubjectTypeProperties
    case createdPropertyTerm
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
        case .deletedSubject:
            return .deletedSubject
        case .promotedSubject:
            return .promotedSubject
        case .createdProperty:
            return .createdProperty
        case .updatedProperty:
            return .updatedProperty
        case .deletedProperty:
            return .deletedProperty
        case .mutatedSubjectTypeProperties:
            return .mutatedSubjectTypeProperties
        case .createdPropertyTerm:
            return .createdPropertyTerm
        }
    }

    /// The Evidence graph this mutation changes, for a mutation that names one
    /// and brings no effects with it. The registry reloads it and the session
    /// patches its progress; nil for every other mutation. S14-16 removes this
    /// when Promote reports effects.
    var evidenceGraphId: String? {
        switch self {
        case .promotedSubject(let graphId):
            return graphId
        default:
            return nil
        }
    }

    /// The ids of `read` this mutation names for a write that reports no
    /// effects: Promote's graph for graph progress, a deleted Source for the
    /// graph list. S14-16 removes this when Promote and DeleteSource report effects.
    func readIDs(for read: CatalogRead) -> [String] {
        switch (read, self) {
        case (.evidenceGraphProgress, _):
            return evidenceGraphId.map { [$0] } ?? []
        case (.evidenceGraphs, .deletedSource(let id)):
            return [id]
        default:
            return []
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
