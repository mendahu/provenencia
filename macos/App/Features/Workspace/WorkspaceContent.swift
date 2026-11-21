import SwiftUI

/// The workspace's single content host (W-3): App Layout toolbar (Back/Forward,
/// breadcrumbs, omnibar shell) plus the active destination below.
struct WorkspaceContent: View {
    /// The content column's coordinate space: overlays and popups place
    /// themselves in it.
    nonisolated static let coordinateSpace = "workspace.content"

    let projectDir: String
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore

    @State private var omnibarResults = OmnibarResultsModel()
    /// The content column's width, for the toolbar popups' width rules. Pages read \.pvWidthClass.
    @State private var contentWidth: CGFloat = 0

    var body: some View {
        VStack(spacing: 0) {
            WorkspaceToolbar(
                omnibarResults: omnibarResults,
                projectDir: projectDir,
                store: store,
                contentWidth: contentWidth
            )
            WorkspaceDestinationHost(
                userID: userID,
                sessionDisplayName: sessionDisplayName,
                store: store
            )
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .onGeometryChange(for: CGFloat.self) { $0.size.width } action: { contentWidth = $0 }
        .coordinateSpace(name: WorkspaceContent.coordinateSpace)
        .accessibilityIdentifier("workspace.content")
    }
}
