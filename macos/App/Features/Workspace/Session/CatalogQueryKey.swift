import Foundation

/// Cache identity for one catalog read. New workspace places add cases here;
/// loaders register in `CatalogQueryRegistry` (S4-02+).
enum CatalogQueryKey: Hashable, Sendable {
    case sourcesList(project: ProjectKey)
    case sourceTypesList(project: ProjectKey)
    case metadataFieldsList(project: ProjectKey)
    case sourceWorkspace(project: ProjectKey, sourceId: String)
    case typeSuggestions(project: ProjectKey, typeId: String)

    var project: ProjectKey {
        switch self {
        case .sourcesList(let project),
             .sourceTypesList(let project),
             .metadataFieldsList(let project),
             .sourceWorkspace(let project, _),
             .typeSuggestions(let project, _):
            project
        }
    }
}
