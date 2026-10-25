import Foundation

/// The session's cached Sources list: the one place a Source is read by id.
extension WorkspaceSession {
    /// One Source by id; `nil` while the Sources list has no value or has no such Source.
    func source(id: String) -> CatalogSource? {
        let handle: QueryHandle<[CatalogSource]>? = queryHandle(.sourcesList(project: projectKey))
        return handle?.value?.first { $0.id == id }
    }
}
