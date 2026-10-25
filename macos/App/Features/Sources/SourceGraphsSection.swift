import Foundation
import Observation

/// The Source page's Evidence graphs: one row per graph of the page's Source,
/// read from the session's graph list. A snapshot of the list, so the rows
/// follow a create, rename or delete without a page-local copy.
@MainActor
@Observable
final class SourceGraphsSection {
    private let context: SourcePageContext

    init(context: SourcePageContext) {
        self.context = context
    }

    /// The Source's graphs in list order (sorted by name); `nil` while the
    /// graph list has no value.
    var rows: [CatalogEvidenceGraph]? {
        context.session.evidenceGraphs(sourceID: context.sourceID)
    }

    /// The graph list's error while it has no value to show.
    var loadError: Error? {
        guard let handle = context.session.evidenceGraphsHandle,
              handle.status == .error, handle.value == nil
        else { return nil }
        return handle.error
    }

    /// Where Open goes: the graph's canvas while it links an Artifact.
    func location(for graph: CatalogEvidenceGraph) -> WorkspaceLocation? {
        guard let source = context.source else { return nil }
        return SourcesListNavigation.graphLocation(for: source, graph: graph)
    }

    /// Why Open is blocked; nil when it opens.
    func blockedReason(for graph: CatalogEvidenceGraph) -> String? {
        SourcesListNavigation.graphBlockedReason(for: graph)
    }
}
