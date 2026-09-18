import SwiftUI

/// Evidence graph workspace destination — Source-scoped canvas shell (S6-01).
///
/// Pan/zoom only. Bubbles, palette, and subject loading arrive in later Spike 6 PRs.
struct EvidenceGraphView: View {
    /// Large enough to pan; not a whole-project graph (design note §7.1).
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
