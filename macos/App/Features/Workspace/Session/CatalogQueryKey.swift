import Foundation

/// Cache identity for one catalog read. New workspace places add cases here;
/// loaders register in `CatalogQueryRegistry` (S4-02+).
enum CatalogQueryKey: Hashable, Sendable {
    case sourcesList(project: ProjectKey)
    case sourceTypesList(project: ProjectKey)
    case metadataFieldsList(project: ProjectKey)
    case credibilityGradesList(project: ProjectKey)
    case sourceWorkspace(project: ProjectKey, sourceId: String)
    case sourceGraph(project: ProjectKey, sourceId: String)
    case citationCounts(project: ProjectKey, sourceId: String)
    case typeSuggestions(project: ProjectKey, typeId: String)
    case subjectFieldsWorkspace(project: ProjectKey)
    case connectRules(project: ProjectKey)
    case propertyTerms(project: ProjectKey, propertyId: String)
    case citationsByArtifact(project: ProjectKey, artifactId: String)

    /// Case identity without associated payload — used by `CatalogQueryRegistry` specs.
    enum Kind: Hashable, Sendable {
        case sourcesList
        case sourceTypesList
        case metadataFieldsList
        case credibilityGradesList
        case sourceWorkspace
        case sourceGraph
        case citationCounts
        case typeSuggestions
        case subjectFieldsWorkspace
        case connectRules
        case propertyTerms
        case citationsByArtifact
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
        case .sourceWorkspace:
            return .sourceWorkspace
        case .sourceGraph:
            return .sourceGraph
        case .citationCounts:
            return .citationCounts
        case .typeSuggestions:
            return .typeSuggestions
        case .subjectFieldsWorkspace:
            return .subjectFieldsWorkspace
        case .connectRules:
            return .connectRules
        case .propertyTerms:
            return .propertyTerms
        case .citationsByArtifact:
            return .citationsByArtifact
        }
    }

    var project: ProjectKey {
        switch self {
        case .sourcesList(let project),
             .sourceTypesList(let project),
             .metadataFieldsList(let project),
             .credibilityGradesList(let project),
             .sourceWorkspace(let project, _),
             .sourceGraph(let project, _),
             .citationCounts(let project, _),
             .typeSuggestions(let project, _),
             .subjectFieldsWorkspace(let project),
             .connectRules(let project),
             .propertyTerms(let project, _),
             .citationsByArtifact(let project, _):
            project
        }
    }
}
