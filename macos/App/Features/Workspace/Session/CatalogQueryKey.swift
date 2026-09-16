import Foundation

/// Cache identity for one catalog read. New workspace places add cases here;
/// loaders register in `CatalogQueryRegistry` (S4-02+).
enum CatalogQueryKey: Hashable, Sendable {
    case sourcesList(project: ProjectKey)
    case sourceTypesList(project: ProjectKey)
    case metadataFieldsList(project: ProjectKey)
    case credibilityGradesList(project: ProjectKey)
    case sourceWorkspace(project: ProjectKey, sourceId: String)
    case typeSuggestions(project: ProjectKey, typeId: String)

    /// Case identity without associated payload — used by `CatalogQueryRegistry` specs.
    enum Kind: Hashable, Sendable {
        case sourcesList
        case sourceTypesList
        case metadataFieldsList
        case credibilityGradesList
        case sourceWorkspace
        case typeSuggestions
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
        case .typeSuggestions:
            return .typeSuggestions
        }
    }

    var project: ProjectKey {
        switch self {
        case .sourcesList(let project),
             .sourceTypesList(let project),
             .metadataFieldsList(let project),
             .credibilityGradesList(let project),
             .sourceWorkspace(let project, _),
             .typeSuggestions(let project, _):
            project
        }
    }
}
