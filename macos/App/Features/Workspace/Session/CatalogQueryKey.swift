import Foundation

/// Cache identity for one catalog read. New workspace places add cases here;
/// loaders register in `CatalogQueryRegistry` (S4-02+).
enum CatalogQueryKey: Hashable, Sendable {
    case sourcesList(project: ProjectKey)
    case sourceTypesList(project: ProjectKey)
    case metadataFieldsList(project: ProjectKey)
    case credibilityGradesList(project: ProjectKey)
    /// Identity Claim confidence grades (seeded vocabulary; S9-12).
    case confidenceGradesList(project: ProjectKey)
    case sourceWorkspace(project: ProjectKey, sourceId: String)
    case sourceGraph(project: ProjectKey, sourceId: String)
    case citationCounts(project: ProjectKey, sourceId: String)
    case typeSuggestions(project: ProjectKey, typeId: String)
    case propertiesWorkspace(project: ProjectKey)
    case connectRules(project: ProjectKey)
    case propertyTerms(project: ProjectKey, propertyId: String)
    case citationsByArtifact(project: ProjectKey, artifactId: String)
    case sourceGraphProgress(project: ProjectKey)
    /// Every Person header, composed from the auto-reconciler cache (S9-07).
    case personsList(project: ProjectKey)
    /// Every Event header, composed from the auto-reconciler cache (S9-22).
    case eventsList(project: ProjectKey)
    /// Existing handles a Subject could join in Promote, best first (S9-11).
    case promoteTargets(project: ProjectKey, subjectId: String)
    /// One handle's detail, any kind (S9-15). Evicted, not revalidated, while
    /// off screen.
    case conclusionDetail(project: ProjectKey, entityId: String)

    /// Case identity without associated payload — used by `CatalogQueryRegistry` specs.
    enum Kind: Hashable, Sendable {
        case sourcesList
        case sourceTypesList
        case metadataFieldsList
        case credibilityGradesList
        case confidenceGradesList
        case sourceWorkspace
        case sourceGraph
        case citationCounts
        case typeSuggestions
        case propertiesWorkspace
        case connectRules
        case propertyTerms
        case citationsByArtifact
        case sourceGraphProgress
        case personsList
        case eventsList
        case promoteTargets
        case conclusionDetail
    }

    var kind: Kind {
        switch self {
        case .sourcesList:
            return .sourcesList
        case .sourceTypesList:
            return .sourceTypesList
        case .metadataFieldsList:
            return .metadataFieldsList
        case .credibilityGradesList:
            return .credibilityGradesList
        case .confidenceGradesList:
            return .confidenceGradesList
        case .sourceWorkspace:
            return .sourceWorkspace
        case .sourceGraph:
            return .sourceGraph
        case .citationCounts:
            return .citationCounts
        case .typeSuggestions:
            return .typeSuggestions
        case .propertiesWorkspace:
            return .propertiesWorkspace
        case .connectRules:
            return .connectRules
        case .propertyTerms:
            return .propertyTerms
        case .citationsByArtifact:
            return .citationsByArtifact
        case .sourceGraphProgress:
            return .sourceGraphProgress
        case .personsList:
            return .personsList
        case .eventsList:
            return .eventsList
        case .promoteTargets:
            return .promoteTargets
        case .conclusionDetail:
            return .conclusionDetail
        }
    }

    var project: ProjectKey {
        switch self {
        case .sourcesList(let project),
             .sourceTypesList(let project),
             .metadataFieldsList(let project),
             .credibilityGradesList(let project),
             .confidenceGradesList(let project),
             .sourceWorkspace(let project, _),
             .sourceGraph(let project, _),
             .citationCounts(let project, _),
             .typeSuggestions(let project, _),
             .propertiesWorkspace(let project),
             .connectRules(let project),
             .propertyTerms(let project, _),
             .citationsByArtifact(let project, _),
             .sourceGraphProgress(let project),
             .personsList(let project),
             .eventsList(let project),
             .promoteTargets(let project, _),
             .conclusionDetail(let project, _):
            project
        }
    }
}
