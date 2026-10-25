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
        guard let graphs = evidenceGraphsHandle?.value else { return nil }
        return graphs.first { $0.id == graphID }?.linkedArtifacts(artifacts) ?? []
    }
}
