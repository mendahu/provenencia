import Foundation

/// The graph list indexed for reads: by graph id and grouped by Source, each
/// Source's graphs in list order (sorted by name). Valid for one handle revision.
struct EvidenceGraphIndex {
    let revision: Int
    let byID: [String: CatalogEvidenceGraph]
    let bySource: [String: [CatalogEvidenceGraph]]

    init(revision: Int, rows: [CatalogEvidenceGraph]) {
        self.revision = revision
        byID = Dictionary(rows.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })
        bySource = Dictionary(grouping: rows, by: \.sourceID)
    }
}

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

    /// The index of the list's current value, rebuilt only when the handle's
    /// revision moved; `nil` while the list has no value.
    private var currentEvidenceGraphIndex: EvidenceGraphIndex? {
        guard let handle = evidenceGraphsHandle, let rows = handle.value else { return nil }
        if let index = evidenceGraphIndex, index.revision == handle.revision { return index }
        let index = EvidenceGraphIndex(revision: handle.revision, rows: rows)
        evidenceGraphIndex = index
        return index
    }

    /// One graph by id; `nil` while the graph list has no value or has no such graph.
    func evidenceGraph(id: String) -> CatalogEvidenceGraph? {
        currentEvidenceGraphIndex?.byID[id]
    }

    /// One Source's graphs in list order (sorted by name); `nil` while the graph
    /// list has no value.
    func evidenceGraphs(sourceID: String) -> [CatalogEvidenceGraph]? {
        guard let index = currentEvidenceGraphIndex else { return nil }
        return index.bySource[sourceID] ?? []
    }

    /// The whole list grouped by Source, each Source's graphs in list order;
    /// empty while the list has no value.
    var evidenceGraphsBySource: [String: [CatalogEvidenceGraph]] {
        currentEvidenceGraphIndex?.bySource ?? [:]
    }
}
