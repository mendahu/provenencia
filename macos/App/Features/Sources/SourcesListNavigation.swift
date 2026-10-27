import Foundation

/// Pure navigation helpers for the Sources list rows and their graph buttons and
/// the Source page graph cards.
enum SourcesListNavigation {
    /// Opens the Source filing page.
    static func pageLocation(for source: CatalogSource) -> WorkspaceLocation {
        .sourcePage(sourceId: source.id, ref: source.ref, title: source.title)
    }

    /// Why one graph's Open is blocked: it links no Artifact. Nil when it opens.
    static func graphBlockedReason(for graph: CatalogEvidenceGraph) -> String? {
        graph.hasArtifact ? nil : L10n.string(L10n.Sources.graphNeedsArtifactTooltip)
    }

    /// Opens one graph's canvas when it links at least one Artifact; `nil`
    /// keeps its Open blocked (`graphBlockedReason(for:)` says why).
    static func graphLocation(for source: CatalogSource, graph: CatalogEvidenceGraph) -> WorkspaceLocation? {
        guard graph.hasArtifact else { return nil }
        return canvasLocation(for: source, graph: graph)
    }

    /// One graph's canvas, whether or not it links an Artifact: what the Sources
    /// list's graph button and an omnibar graph hit both open. The location
    /// carries the Source's ref and title, as Go's `deleteimpact.graphLocation` does.
    static func canvasLocation(for source: CatalogSource, graph: CatalogEvidenceGraph) -> WorkspaceLocation {
        .evidenceGraph(
            sourceId: source.id,
            graphId: graph.id,
            ref: source.ref,
            title: source.title,
            sourceTitle: source.title
        )
    }
}
