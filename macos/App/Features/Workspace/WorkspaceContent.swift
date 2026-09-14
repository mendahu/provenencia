import SwiftUI

/// The workspace's single content host (W-3): App Layout toolbar (Back/Forward,
/// breadcrumbs, omnibar shell) plus the active destination below. **Sources**
/// (S2-17), **Source fields** (S2-15), and **Source types** (S2-16) mount their
/// own full-height views; Files still shows the labeled empty placeholder
/// (project Files list was descoped with S2-20/S2-21).
struct WorkspaceContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let projectDir: String
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore
    let catalogCounts: CatalogCounts

    /// Hosted here (not on the 52pt toolbar row) so the jump menu can paint and
    /// receive hits over the page below — same reason `pvContextMenu` wants a
    /// large ancestor.
    @State private var jumpMenu = HistoryJumpMenuModel()

    private var section: WorkspaceSection { navigation.selectedSection }

    var body: some View {
        VStack(spacing: 0) {
            WorkspaceToolbar(jumpMenu: jumpMenu)
            switch section {
            case .sourceFields:
                SourceFieldsView(
                    projectDir: projectDir,
                    userID: userID,
                    store: store,
                    catalogCounts: catalogCounts
                )
            case .sourceTypes:
                SourceTypesView(
                    projectDir: projectDir,
                    userID: userID,
                    store: store,
                    catalogCounts: catalogCounts
                )
            case .sources:
                SourcesView(
                    projectDir: projectDir,
                    userID: userID,
                    sessionDisplayName: sessionDisplayName,
                    store: store,
                    catalogCounts: catalogCounts
                )
            default:
                ScrollView {
                    VStack(alignment: .leading, spacing: PVSpacing.space7) {
                        Text(section.label)
                            .font(PVFont.display(size: PVTypeScale.h1))
                            .foregroundStyle(PVColor.textDisplay)
                        placeholder
                    }
                    .frame(maxWidth: PVSpacing.widthContentMax, alignment: .leading)
                    .frame(maxWidth: .infinity, alignment: .center)
                    .padding(.top, PVSpacing.space8)
                    .padding(.horizontal, PVSpacing.gutterPage)
                    .padding(.bottom, PVSpacing.space9)
                }
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(PVColor.surfacePage)
        .coordinateSpace(name: HistoryJumpMenuModel.contentCoordinateSpace)
        .overlay(alignment: .topLeading) {
            HistoryJumpMenuHost(jumpMenu: jumpMenu)
        }
        .accessibilityIdentifier("workspace.content")
    }

    @ViewBuilder
    private var placeholder: some View {
        VStack(spacing: PVSpacing.space2) {
            Text(section.label)
                .font(PVFont.display(size: PVTypeScale.h3, weight: PVFontWeight.semibold))
                .foregroundStyle(PVColor.textSecondary)
            Text(section.placeholderNote)
                .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: PVSpacing.measureNarrow)
        .padding(PVSpacing.space10)
        .frame(maxWidth: .infinity, minHeight: 240)
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [4, 3]))
        )
        .accessibilityIdentifier("workspace.content.placeholder")
    }
}
