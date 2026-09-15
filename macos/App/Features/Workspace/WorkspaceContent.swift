import SwiftUI

/// The workspace's single content host (W-3): App Layout toolbar (Back/Forward,
/// breadcrumbs, omnibar shell) plus the active destination below.
struct WorkspaceContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let projectDir: String
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore
    let catalogCounts: CatalogCounts

    /// Hosted here (not on the 52pt toolbar row) so the jump menu / omnibar
    /// results can paint and receive hits over the page below — same reason
    /// `pvContextMenu` wants a large ancestor.
    @State private var jumpMenu = HistoryJumpMenuModel()
    @State private var omnibarResults = OmnibarResultsModel()

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
                store: store,
                catalogCounts: catalogCounts
            )
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .coordinateSpace(name: HistoryJumpMenuModel.contentCoordinateSpace)
        .overlay(alignment: .topLeading) {
            HistoryJumpMenuHost(jumpMenu: jumpMenu)
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
