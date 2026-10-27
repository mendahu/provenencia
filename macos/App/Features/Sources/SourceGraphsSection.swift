import Foundation
import Observation

/// The Source page's Evidence graphs: one row per graph of the page's Source,
/// read from the session's graph list. A snapshot of the list, so the rows
/// follow a create, rename or delete without a page-local copy.
@MainActor
@Observable
final class SourceGraphsSection {
    private let context: SourcePageContext

    /// The Add / Edit form while it is open.
    var form: SourceGraphFormModel?

    init(context: SourcePageContext) {
        self.context = context
    }

    /// The Source's graphs in list order (sorted by name); `nil` while the
    /// graph list has no value.
    var rows: [CatalogEvidenceGraph]? {
        guard case .loaded(let graphs) = context.session.evidenceGraphsRead(sourceID: context.sourceID) else { return nil }
        return graphs
    }

    /// The graph list's error while it has no value to show.
    var loadError: Error? {
        guard case .failed(let error) = context.session.evidenceGraphsRead(sourceID: context.sourceID) else { return nil }
        return error
    }

    /// The counts of one graph; `nil` while the progress map has no value or
    /// omits the graph (zero counts).
    func progress(for graph: CatalogEvidenceGraph) -> EvidenceGraphProgress? {
        context.session.evidenceGraphProgress(graphID: graph.id)
    }

    /// True while the counts can still arrive.
    var countsLoading: Bool { context.session.evidenceGraphCountsLoading }

    /// The progress map's failure while it has no value to show.
    var countsError: Error? { context.session.evidenceGraphCountsError }

    /// Where Open goes: the graph's canvas while it links an Artifact.
    func location(for graph: CatalogEvidenceGraph) -> WorkspaceLocation? {
        guard let source = context.source else { return nil }
        return SourcesListNavigation.graphLocation(for: source, graph: graph)
    }

    /// Why Open is blocked; nil when it opens.
    func blockedReason(for graph: CatalogEvidenceGraph) -> String? {
        SourcesListNavigation.graphBlockedReason(for: graph)
    }

    // MARK: Add and edit

    /// A new graph needs an Artifact to link.
    var canAdd: Bool { !(context.workspace?.artifacts.isEmpty ?? true) }

    /// Why Add is disabled; nil when it is enabled, and while the page loads.
    var addBlockedReason: String? {
        guard let workspace = context.workspace, workspace.artifacts.isEmpty else { return nil }
        return L10n.string(L10n.Sources.addEvidenceGraphNeedsArtifact)
    }

    func openAdd() {
        guard canAdd else { return }
        form = SourceGraphFormModel(context: context, mode: .create)
    }

    /// The graph with `id` when it belongs to this page's Source.
    func ownGraph(id: String) -> CatalogEvidenceGraph? {
        guard let graph = context.session.evidenceGraph(id: id), graph.sourceID == context.sourceID else { return nil }
        return graph
    }

    func openEdit(id: String) {
        guard context.workspace != nil, let graph = ownGraph(id: id) else { return }
        let model = SourceGraphFormModel(context: context, mode: .edit(graph))
        form = model
        Task { await model.loadKeptLinks() }
    }
}
