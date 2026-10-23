import Foundation

/// Pure navigation helpers for the Sources list split-row (S5-08 / S5-D2).
enum SourcesListNavigation {
    /// Opens the Source filing page.
    static func pageLocation(for source: CatalogSource) -> WorkspaceLocation {
        WorkspaceLocation(
            section: .sources,
            sourceId: source.id,
            sourceSurface: .page,
            ref: source.ref,
            title: source.title
        )
    }

    /// The graph the Source-level entry points open: the Source's first graph
    /// in the graph list, which is sorted by name. The interim rule until each
    /// graph has its own Open (S12-11, S12-14).
    static func entryGraph(sourceID: String, in graphs: [CatalogEvidenceGraph]) -> CatalogEvidenceGraph? {
        graphs.first { $0.sourceID == sourceID }
    }

    /// Opens the entry graph's canvas when that graph links at least one
    /// Artifact. Returns `nil` when the graph zone should stay blocked.
    static func graphLocation(for source: CatalogSource, graphs: [CatalogEvidenceGraph]) -> WorkspaceLocation? {
        guard let graph = entryGraph(sourceID: source.id, in: graphs), !graph.artifactIDs.isEmpty else { return nil }
        return .evidenceGraph(
            sourceId: source.id,
            graphId: graph.id,
            ref: source.ref,
            title: source.title
        )
    }
}
