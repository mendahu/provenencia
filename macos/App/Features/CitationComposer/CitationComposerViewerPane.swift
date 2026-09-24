import SwiftUI

/// Left pane: Artifact header + capability tool strip + canvas + locator list.
struct CitationComposerViewerPane: View {
    @Bindable var model: CitationComposerModel
    @State private var clearMenu = PVContextMenuState()
    @State private var clearKeyboard = PVContextMenuKeyboard.inactive

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
                .padding(.horizontal, PVSpacing.space7)
                .padding(.top, PVSpacing.space7)
                .padding(.bottom, PVSpacing.space3)

            toolStrip

            ArtifactViewer(
                model: model.artifactViewer,
                showsToolStrip: false,
                locatorPage: model.locator.page,
                onSetPage: { model.setPageFromViewer() },
                armedRegionTool: $model.armedRegionTool,
                committedRegion: model.locator.region,
                onCommitRegion: { model.setRegion($0) },
                onDisarmRegionTool: { model.disarmRegionTool() },
                freeformDeleteTooltip: String(localized: L10n.CitationComposer.freeformDeleteVertex)
            )
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .padding(PVSpacing.space7)

            locatorList
        }
    }

    private var header: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3) {
            Text(verbatim: model.sourceTitle.isEmpty ? "—" : model.sourceTitle)
                .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textDisplay)
                .lineLimit(1)
            Text(verbatim: "·")
                .foregroundStyle(PVColor.textMuted)
            Text(verbatim: artifactIndexCaption)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .lineLimit(1)
            Spacer(minLength: 0)
        }
    }

    private var artifactIndexCaption: String {
        guard let index = model.selectedArtifactIndex else {
            return String(localized: L10n.CitationComposer.artifactKindUnknown)
        }
        let kind: String
        switch model.artifactViewer.kind {
        case .pdf:
            kind = String(localized: L10n.CitationComposer.artifactKindPDF)
        case .image:
            kind = String(localized: L10n.CitationComposer.artifactKindImage)
        default:
            kind = String(localized: L10n.CitationComposer.artifactKindUnknown)
        }
        return L10n.CitationComposer.artifactIndexOf(
            index: index + 1,
            total: model.artifacts.count,
            kind: kind
        )
    }

    /// Capability layout: viewer chrome (page / Set page / zoom / radios) + Clear.
    private var toolStrip: some View {
        HStack(spacing: 12) {
            ArtifactViewerToolChrome(
                model: model.artifactViewer,
                locatorPage: model.locator.page,
                onSetPage: { model.setPageFromViewer() },
                armedRegionTool: $model.armedRegionTool
            )

            Spacer(minLength: 0)

            clearMenuButton
        }
        .frame(height: 44)
        .padding(.horizontal, 16)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
        }
        .pvContextMenu($clearMenu, keyboard: $clearKeyboard) {
            PVContextMenuPanel(
                accessibilityIdentifier: "citationComposer.clearMenu.menu"
            ) {
                PVContextMenuItem(
                    L10n.CitationComposer.clearRegion,
                    index: 0,
                    isEnabled: model.locator.hasRegion
                        && model.locatorCapabilities.supportsRegionLocator,
                    accessibilityIdentifier: "citationComposer.clearRegion"
                ) {
                    model.clearRegion()
                }
                PVContextMenuItem(
                    L10n.CitationComposer.resetToEntireArtifact,
                    index: 1,
                    accessibilityIdentifier: "citationComposer.resetToEntireArtifact"
                ) {
                    model.resetToEntireArtifact()
                }
            }
        }
    }

    private var clearMenuButton: some View {
        Button {
            if clearMenu.isPresented {
                clearMenu.dismiss()
            } else {
                clearKeyboard = PVContextMenuKeyboard(itemCount: 2, activeIndex: -1)
                clearMenu.present(at: CGPoint(x: 0, y: 28))
            }
        } label: {
            Text(L10n.CitationComposer.clearMenu)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.accent)
        }
        .buttonStyle(.plain)
        .disabled(model.isLocatorArtifactOnly)
        .accessibilityIdentifier("citationComposer.clearMenu")
    }

    private var locatorList: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack(alignment: .firstTextBaseline, spacing: 8) {
                Text(L10n.CitationComposer.locatorSection)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textSecondary)
                Text(L10n.CitationComposer.locatorOuterToInner)
                    .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }

            locatorRow(
                icon: .file,
                title: String(localized: L10n.CitationComposer.locatorEntireArtifact),
                helper: artifactHelper,
                helperMono: false,
                depth: 0,
                locked: true,
                accessibilityIdentifier: "citationComposer.locator.artifact"
            )

            if model.locatorCapabilities.supportsPageLocator, let page = model.locator.page {
                locatorRow(
                    icon: .bookmark,
                    title: L10n.CitationComposer.locatorPage(page),
                    helper: L10n.ArtifactViewer.pageOf(total: model.artifactViewer.pageCount),
                    helperMono: false,
                    depth: 1,
                    removeLabel: L10n.CitationComposer.removePage(page),
                    accessibilityIdentifier: "citationComposer.locator.page",
                    onRemove: { model.removePage() }
                )
            }

            if model.locatorCapabilities.supportsRegionLocator, let region = model.locator.region {
                let depth = model.locator.page == nil ? 1 : 2
                locatorRow(
                    icon: regionSymbol(region.kind),
                    title: regionNoun(region.kind),
                    helper: L10n.CitationComposer.locatorPoints(region.points.count),
                    helperMono: true,
                    depth: depth,
                    removeLabel: String(localized: L10n.CitationComposer.removeRegion),
                    accessibilityIdentifier: "citationComposer.locator.region",
                    onRemove: { model.removeRegion() }
                )
            }
        }
        .padding(.horizontal, 16)
        .padding(.top, 10)
        .padding(.bottom, 12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .top) {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
        }
        .accessibilityIdentifier("citationComposer.locatorList")
    }

    private var artifactHelper: String? {
        let label = model.selectedArtifact?.label.trimmingCharacters(in: .whitespacesAndNewlines) ?? ""
        return label.isEmpty ? nil : label
    }

    private func regionNoun(_ kind: ArtifactRegionKind) -> String {
        switch kind {
        case .rectangle:
            return String(localized: L10n.CitationComposer.locatorRectangle)
        case .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft:
            return String(localized: L10n.CitationComposer.locatorLShape)
        case .circle:
            return String(localized: L10n.CitationComposer.locatorCircle)
        case .freeform:
            return String(localized: L10n.CitationComposer.locatorPolygon)
        }
    }

    private func regionSymbol(_ kind: ArtifactRegionKind) -> PVSymbol {
        switch kind {
        case .rectangle: .regionRectangle
        case .lOpenTopRight: .regionLTopRight
        case .lOpenTopLeft: .regionLTopLeft
        case .lOpenBottomRight: .regionLBottomRight
        case .lOpenBottomLeft: .regionLBottomLeft
        case .circle: .regionCircle
        case .freeform: .regionFreeform
        }
    }

    private func locatorRow(
        icon: PVSymbol,
        title: String,
        helper: String?,
        helperMono: Bool,
        depth: Int,
        locked: Bool = false,
        removeLabel: String? = nil,
        accessibilityIdentifier: String,
        onRemove: (() -> Void)? = nil
    ) -> some View {
        HStack(alignment: .center, spacing: 8) {
            if depth > 0 {
                LocatorHierarchyElbow()
            }
            PVIcon(icon, size: 14)
                .foregroundStyle(PVColor.textSecondary)
            Text(verbatim: title)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
            if let helper {
                Text(verbatim: helper)
                    .font(
                        helperMono
                            ? PVFont.mono(size: PVTypeScale.micro)
                            : PVFont.body(size: 12)
                    )
                    .foregroundStyle(PVColor.textMuted)
            }
            Spacer(minLength: 0)
            if locked {
                HStack(spacing: 5) {
                    PVIcon(.lock, size: 12)
                    Text(L10n.CitationComposer.locatorAlwaysIncluded)
                        .font(PVFont.body(size: 12, italic: true))
                }
                .foregroundStyle(PVColor.textFaint)
            } else if let onRemove, let removeLabel {
                PVIconButton(
                    .dismiss,
                    label: L10n.CitationComposer.removeLocatorLayer,
                    size: .sm
                ) {
                    onRemove()
                }
                .accessibilityLabel(Text(verbatim: removeLabel))
                .accessibilityIdentifier("\(accessibilityIdentifier).remove")
            }
        }
        .frame(minHeight: 28)
        .padding(.leading, CGFloat(depth) * 20)
        .accessibilityIdentifier(accessibilityIdentifier)
    }
}

/// └ connector between nested locator rows (S7-D9 board).
private struct LocatorHierarchyElbow: View {
    var body: some View {
        Path { path in
            path.move(to: CGPoint(x: 0.5, y: 0))
            path.addLine(to: CGPoint(x: 0.5, y: 14))
            path.addLine(to: CGPoint(x: 10, y: 14))
        }
        .stroke(PVColor.borderDefault, lineWidth: 1)
        .frame(width: 10, height: 14)
        .offset(y: -6)
        .accessibilityHidden(true)
    }
}
