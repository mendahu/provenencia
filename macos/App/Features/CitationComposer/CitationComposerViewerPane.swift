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
                .padding(.horizontal, PVSpacing.space7)
                .padding(.bottom, PVSpacing.space7)
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
            Spacer(minLength: PVSpacing.space3)
            if model.artifacts.count > 1 {
                Button {
                    model.changeArtifact()
                } label: {
                    Text(L10n.CitationComposer.changeArtifact)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.accent)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("citationComposer.changeArtifact")
            }
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
        VStack(alignment: .leading, spacing: PVSpacing.space3) {
            HStack {
                Text(L10n.CitationComposer.locatorSection)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textSecondary)
                Spacer(minLength: 0)
                Text(L10n.CitationComposer.locatorOuterToInner)
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textFaint)
            }

            locatorRow(
                title: String(localized: L10n.CitationComposer.locatorEntireArtifact),
                note: String(localized: L10n.CitationComposer.locatorAlwaysIncluded),
                locked: true,
                accessibilityIdentifier: "citationComposer.locator.artifact"
            )

            if model.locatorCapabilities.supportsPageLocator, let page = model.locator.page {
                locatorRow(
                    title: L10n.CitationComposer.locatorPage(page),
                    note: nil,
                    locked: false,
                    removeLabel: L10n.CitationComposer.removePage(page),
                    accessibilityIdentifier: "citationComposer.locator.page",
                    onRemove: { model.removePage() }
                )
            }

            if model.locatorCapabilities.supportsRegionLocator, let region = model.locator.region {
                locatorRow(
                    title: regionNoun(region.kind),
                    note: nil,
                    locked: false,
                    removeLabel: String(localized: L10n.CitationComposer.removeRegion),
                    accessibilityIdentifier: "citationComposer.locator.region",
                    onRemove: { model.removeRegion() }
                )
            }
        }
        .accessibilityIdentifier("citationComposer.locatorList")
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

    private func locatorRow(
        title: String,
        note: String?,
        locked: Bool,
        removeLabel: String? = nil,
        accessibilityIdentifier: String,
        onRemove: (() -> Void)? = nil
    ) -> some View {
        HStack(spacing: PVSpacing.space3) {
            if locked {
                PVIcon(.lock, size: 11)
                    .foregroundStyle(PVColor.textFaint)
            } else {
                Color.clear.frame(width: 11, height: 11)
            }
            VStack(alignment: .leading, spacing: 1) {
                Text(verbatim: title)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                if let note {
                    Text(verbatim: note)
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textFaint)
                }
            }
            Spacer(minLength: 0)
            if let onRemove, let removeLabel {
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
        .padding(.vertical, 4)
        .accessibilityIdentifier(accessibilityIdentifier)
    }
}
