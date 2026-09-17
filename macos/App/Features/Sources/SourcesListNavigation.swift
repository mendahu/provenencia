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

    /// Opens the Evidence graph when the Source has at least one Artifact.
    /// Returns `nil` when the graph zone should stay blocked.
    static func graphLocation(for source: CatalogSource) -> WorkspaceLocation? {
        guard source.hasArtifact else { return nil }
        return WorkspaceLocation(
            section: .sources,
            sourceId: source.id,
            sourceSurface: .graph,
            ref: source.ref,
            title: source.title
        )
    }
}
