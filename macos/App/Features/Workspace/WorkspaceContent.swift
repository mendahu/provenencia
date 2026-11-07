import SwiftUI

/// The workspace's single content host (W-3): App Layout toolbar (Back/Forward,
/// breadcrumbs, omnibar shell) plus the active destination below.
struct WorkspaceContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let projectDir: String
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore

    /// Hosted here (not on the 52pt toolbar row) so the jump menu / omnibar
    /// results can paint and receive hits over the page below.
    @State private var jumpMenu = HistoryJumpMenuModel()
    @State private var omnibarResults = OmnibarResultsModel()
    /// The content column's width, for placing overlays only. Pages read \.pvWidthClass.
    @State private var contentWidth: CGFloat = 0

    var body: some View {
        VStack(spacing: 0) {
            WorkspaceToolbar(
                jumpMenu: jumpMenu,
                omnibarResults: omnibarResults,
                projectDir: projectDir,
                store: store
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
        .coordinateSpace(name: HistoryJumpMenuModel.contentCoordinateSpace)
        .overlay(alignment: .topLeading) {
            HistoryJumpMenuHost(jumpMenu: jumpMenu, contentWidth: contentWidth)
        }
        .overlay(alignment: .topLeading) {
            OmnibarResultsHost(
                results: omnibarResults,
                projectDir: projectDir,
                onActivate: { hit in
                    omnibarResults.activate(hit, navigation: navigation)
                }
            )
        }
        .accessibilityIdentifier("workspace.content")
    }
}
