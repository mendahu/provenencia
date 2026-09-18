import SwiftUI

/// Workspace destination for a Source’s Evidence graph.
///
/// Product place under Sources (`sourceSurface: .graph`). Composes the
/// reusable [`GraphCanvas`](../GraphCanvas/) shell; later PRs add subjects,
/// palette, and connect on top of that tooling — other visualizations can
/// use the same canvas without living in this feature.
struct EvidenceGraphView: View {
    /// Large enough to pan; Source-scoped graphs stay small (design note §7.1).
    private static let contentSize = CGSize(width: 4_000, height: 4_000)

    var body: some View {
        GraphCanvasScrollView(contentSize: Self.contentSize) {
            GraphCanvasGridView(contentSize: Self.contentSize)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .accessibilityIdentifier("workspace.destination.evidenceGraph")
        .accessibilityLabel(Text(L10n.Workspace.evidenceGraphTitle))
    }
}
