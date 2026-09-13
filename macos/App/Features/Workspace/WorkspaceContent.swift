import SwiftUI

/// The workspace's single content host (W-3): a slim header carrying page
/// context plus project identity (W-9, W-10), and the active
/// destination's content below it. **Sources** (S2-17), **Source fields**
/// (S2-15), and **Source types** (S2-16) mount their own full-height views
/// below that header; Files still shows the labeled empty placeholder
/// (project Files list was descoped with S2-20/S2-21).
struct WorkspaceContent: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let project: ProjectInfo?
    let projectDir: String
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore
    let catalogCounts: CatalogCounts

    private var section: WorkspaceSection { navigation.selectedSection }

    var body: some View {
        VStack(spacing: 0) {
            header
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
        .accessibilityIdentifier("workspace.content")
    }

    @ViewBuilder
    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space6) {
            Text(section.label)
                .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textSecondary)
            Spacer(minLength: PVSpacing.space6)
            if let project {
                HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
                    Text(project.label)
                        .font(PVFont.display(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(1)
                    Text(project.folderName)
                        .font(PVFont.mono(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textFaint)
                        .lineLimit(1)
                }
                .accessibilityElement(children: .combine)
                .accessibilityIdentifier("workspace.content.projectIdentity")
            }
        }
        .padding(.horizontal, PVSpacing.gutterPage)
        .frame(maxWidth: .infinity)
        .pvWorkspaceHeaderRow()
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
