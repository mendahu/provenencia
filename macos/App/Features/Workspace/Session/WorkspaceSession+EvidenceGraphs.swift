import Foundation

/// The session's cached Evidence graph list: the one place its key and handle
/// are looked up, and the graph's linked Artifacts read from it.
extension WorkspaceSession {
    var evidenceGraphsKey: CatalogQueryKey { .evidenceGraphsList(project: projectKey) }

    var evidenceGraphsHandle: QueryHandle<[CatalogEvidenceGraph]>? {
        queryHandle(evidenceGraphsKey)
    }

    /// The Artifacts of `artifacts` that the graph links, in `artifacts` order.
    /// `nil` while the graph list has no value; empty when the list has no graph
    /// with that id.
    func linkedArtifacts(graphID: String, artifacts: [CatalogArtifact]) -> [CatalogArtifact]? {
        guard evidenceGraphsHandle?.value != nil else { return nil }
        return evidenceGraph(id: graphID)?.linkedArtifacts(artifacts) ?? []
    }

    /// One graph by id; `nil` while the graph list has no value or has no such graph.
    func evidenceGraph(id: String) -> CatalogEvidenceGraph? {
        evidenceGraphsHandle?.value?.first { $0.id == id }
    }

    /// One Source's graphs in list order (sorted by name); `nil` while the graph
    /// list has no value.
    func evidenceGraphs(sourceID: String) -> [CatalogEvidenceGraph]? {
        evidenceGraphsHandle?.value?.filter { $0.sourceID == sourceID }
    }
}
