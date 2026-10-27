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

/// The session's cached Evidence graph list and graph progress: the one place
/// their keys and handles are looked up, and the graph's linked Artifacts read
/// from them.
extension WorkspaceSession {
    var evidenceGraphsKey: CatalogQueryKey { .evidenceGraphsList(project: projectKey) }

    var evidenceGraphsHandle: QueryHandle<[CatalogEvidenceGraph]>? {
        queryHandle(evidenceGraphsKey)
    }

    var evidenceGraphProgressKey: CatalogQueryKey { .evidenceGraphProgress(project: projectKey) }

    var evidenceGraphProgressHandle: QueryHandle<[String: EvidenceGraphProgress]>? {
        queryHandle(evidenceGraphProgressKey)
    }

    /// The counts of one Evidence graph, by graph id; `nil` while the progress
    /// map has no value or omits the graph (Go omits a graph with zero counts).
    func evidenceGraphProgress(graphID: String) -> EvidenceGraphProgress? {
        evidenceGraphProgressHandle?.value?[graphID]
    }

    /// True while the progress map or the graph list is missing or still
    /// loading, so a graph's counts can still arrive.
    var evidenceGraphCountsLoading: Bool {
        guard let progress = evidenceGraphProgressHandle, let graphs = evidenceGraphsHandle else { return true }
        return progress.read.isLoading || graphs.read.isLoading
    }

    /// The progress map's failure while it has no value to show; `nil` otherwise.
    var evidenceGraphCountsError: Error? {
        if case .failed(let error) = evidenceGraphProgressHandle?.read { return error }
        return nil
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

    /// One Source's graphs as a read. A loaded list with no graph for the
    /// Source reads `.loading`: every Source has a Default graph from creation,
    /// so that is a list not yet refreshed.
    func evidenceGraphsRead(sourceID: String) -> QueryRead<[CatalogEvidenceGraph]> {
        guard let handle = evidenceGraphsHandle else { return .loading }
        switch handle.read {
        case .loading: return .loading
        case .failed(let error): return .failed(error)
        case .loaded:
            guard let graphs = evidenceGraphs(sourceID: sourceID), !graphs.isEmpty else { return .loading }
            return .loaded(graphs)
        }
    }
}
