import Foundation

/// Pure navigation helpers for the Sources list split-row (S5-08 / S5-D2).
enum SourcesListNavigation {
    /// Opens the Source filing page.
    static func pageLocation(for source: CatalogSource) -> WorkspaceLocation {
        .sourcePage(sourceId: source.id, ref: source.ref, title: source.title)
    }

    /// The graph the Source-level entry points open: the Source's first graph
    /// in the graph list, which is sorted by name. The interim rule until each
    /// graph has its own Open (S12-11, S12-14).
    static func entryGraph(sourceID: String, in graphs: [String: [CatalogEvidenceGraph]]) -> CatalogEvidenceGraph? {
        graphs[sourceID]?.first
    }

    /// The graph list grouped by Source in one pass, each Source's graphs in
    /// list order (sorted by name).
    static func graphsBySource(_ graphs: [CatalogEvidenceGraph]) -> [String: [CatalogEvidenceGraph]] {
        Dictionary(grouping: graphs, by: \.sourceID)
    }

    /// Why Open graph is blocked, or nil when it is
    /// open or the reason is not yet known. With no location, a Source without
    /// an Artifact, or one whose graph list has a value but whose entry graph
    /// links none, needs an Artifact; while the graph list has no value
    /// (loading or failed) the reason is unknown.
    static func graphBlockedReason(hasArtifact: Bool, graphListLoaded: Bool, location: WorkspaceLocation?) -> String? {
        guard location == nil else { return nil }
        guard !hasArtifact || graphListLoaded else { return nil }
        return L10n.string(L10n.Sources.needsArtifactTooltip)
    }

    /// Why one graph's Open is blocked: it links no Artifact. Nil when it opens.
    static func graphBlockedReason(for graph: CatalogEvidenceGraph) -> String? {
        graph.artifactIDs.isEmpty ? L10n.string(L10n.Sources.graphNeedsArtifactTooltip) : nil
    }

    /// Opens the entry graph's canvas when that graph links at least one
    /// Artifact. Returns `nil` when the graph zone should stay blocked.
    static func graphLocation(for source: CatalogSource, graphs: [String: [CatalogEvidenceGraph]]) -> WorkspaceLocation? {
        guard let graph = entryGraph(sourceID: source.id, in: graphs) else { return nil }
        return graphLocation(for: source, graph: graph)
    }

    /// Opens one graph's canvas when it links at least one Artifact; `nil`
    /// keeps its Open blocked (`graphBlockedReason(for:)` says why).
    static func graphLocation(for source: CatalogSource, graph: CatalogEvidenceGraph) -> WorkspaceLocation? {
        guard !graph.artifactIDs.isEmpty else { return nil }
        return .evidenceGraph(
            sourceId: source.id,
            graphId: graph.id,
            ref: source.ref,
            title: source.title
        )
    }
}
