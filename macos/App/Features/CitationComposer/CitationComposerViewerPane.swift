import SwiftUI

/// File viewer: tool strip, a one-line locator, then the canvas.
struct CitationComposerViewerPane: View {
    @Bindable var model: CitationComposerModel
    @State private var clearMenu = PVContextMenuState()
    @State private var clearKeyboard = PVContextMenuKeyboard.inactive
    @FocusState private var findFieldFocused: Bool

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            toolStrip
            locatorStrip
                .layoutPriority(-1)
            if model.isDrawingDepiction {
                drawingBand
            }

            ArtifactViewer(
                model: model.artifactViewer,
                showsToolStrip: false,
                locatorPage: model.locator.page,
                onSetPage: { model.setPageFromViewer() },
                armedRegionTool: $model.viewerArmedRegionTool,
                committedRegion: model.viewerRegion,
                onCommitRegion: { model.commitViewerRegion($0) },
                onDisarmRegionTool: { model.disarmRegionTool() },
                freeformDeleteTooltip: L10n.string(L10n.CitationComposer.freeformDeleteVertex)
            )
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .padding(PVSpacing.space6)
        }
    }

    /// While a depiction's region is drawn: the hint (or the wrong-page warning)
    /// and Cancel. Escape cancels too.
    private var drawingBand: some View {
        HStack(alignment: .center, spacing: PVSpacing.space5) {
            if model.depictionWrongPage, let page = model.depictionPage {
                PVCallout(
                    tone: .warning,
                    message: L10n.CitationComposer.depictionWrongPage(page),
                    compact: true
                )
                .frame(maxWidth: .infinity)
            } else {
                PVCallout(
                    tone: .info,
                    message: L10n.string(L10n.CitationComposer.depictionDrawingHint),
                    compact: true
                )
                .frame(maxWidth: .infinity)
            }
            PVButton(L10n.CitationComposer.depictionCancel, variant: .ghost, size: .sm) {
                model.cancelDepictionDrawing()
            }
            .keyboardShortcut(.cancelAction)
            .accessibilityIdentifier("citationComposer.depiction.cancel")
        }
        .padding(.horizontal, PVSpacing.space6)
        .padding(.vertical, PVSpacing.space4)
        .background(PVColor.surfaceCard)
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("citationComposer.depiction.drawingBand")
    }

    /// Capability layout: viewer chrome (page / Set page / zoom / radios) + Clear + Find.
    private var toolStrip: some View {
        VStack(spacing: 0) {
            HStack(spacing: PVSpacing.space5) {
                ArtifactViewerToolChrome(
                    model: model.artifactViewer,
                    locatorPage: model.locator.page,
                    onSetPage: { model.setPageFromViewer() },
                    armedRegionTool: $model.armedRegionTool
                )

                Spacer(minLength: 0)

                clearMenuButton

                if model.artifactViewer.kind == .pdf {
                    findToggleButton
                }
            }
            .frame(height: PVSpacing.hitMin)
            .padding(.horizontal, PVSpacing.space6)
            .disabled(model.isTranscribing)
            PVDivider()
            if model.artifactViewer.kind == .pdf, model.artifactViewer.isFindPresented {
                findRow
                    .frame(height: PVSpacing.space10)
                    .padding(.horizontal, PVSpacing.space6)
                    .disabled(model.isTranscribing)
                PVDivider()
            }
        }
        .background(PVColor.surfaceCard)
    }

    private var findToggleButton: some View {
        PVIconButton(
            .search,
            label: L10n.ArtifactViewer.findToggle,
            accessibilityLabel: model.artifactViewer.findHasTextLayer
                ? L10n.ArtifactViewer.findToggle
                : L10n.ArtifactViewer.findUnavailable,
            size: .sm,
            isSelected: model.artifactViewer.isFindPresented
        ) {
            model.artifactViewer.toggleFind()
            if model.artifactViewer.isFindPresented {
                findFieldFocused = true
            }
        }
        .background {
            Button {
                model.artifactViewer.openFind()
                findFieldFocused = true
            } label: {
                EmptyView()
            }
            .keyboardShortcut("f", modifiers: .command)
            .opacity(0)
            .accessibilityHidden(true)
        }
        .accessibilityIdentifier("artifactViewer.findToggle")
    }

    private var findRow: some View {
        let viewer = model.artifactViewer
        let note = viewer.findStatusNote
        return HStack(spacing: PVSpacing.space4) {
            PVInput(
                text: Binding(
                    get: { viewer.findQuery },
                    set: { viewer.findQuery = $0 }
                ),
                size: .sm,
                prompt: L10n.ArtifactViewer.findToggle,
                icon: .search,
                suffix: viewer.findCountSuffix,
                isInvalid: viewer.findIsInvalid,
                focused: $findFieldFocused,
                activateOnAppear: true
            )
            .frame(width: 220)
            .disabled(!viewer.findHasTextLayer)
            .onSubmit { viewer.findNext() }
            .onKeyPress(keys: [.return], phases: .down) { press in
                if press.modifiers.contains(.shift) {
                    viewer.findPrevious()
                    return .handled
                }
                return .ignored
            }
            .accessibilityLabel(Text(L10n.ArtifactViewer.findField))
            .accessibilityIdentifier("artifactViewer.findField")

            PVIconButton(
                .chevronUp,
                label: L10n.ArtifactViewer.findPrevious,
                size: .sm
            ) {
                viewer.findPrevious()
            }
            .disabled(viewer.findMatchCount == 0)
            .accessibilityIdentifier("artifactViewer.findPrevious")

            PVIconButton(
                .chevronDown,
                label: L10n.ArtifactViewer.findNext,
                size: .sm
            ) {
                viewer.findNext()
            }
            .disabled(viewer.findMatchCount == 0)
            .accessibilityIdentifier("artifactViewer.findNext")

            if !note.isEmpty {
                Text(verbatim: note)
                    .font(PVFont.body(size: 12, italic: true))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
                    .frame(maxWidth: .infinity, alignment: .leading)
            } else {
                Spacer(minLength: 0)
            }

            PVButton(L10n.ArtifactViewer.findDone, variant: .ghost, size: .sm) {
                viewer.closeFind()
            }
            .keyboardShortcut(.cancelAction)
            .accessibilityIdentifier("artifactViewer.findDone")
        }
        .onKeyPress(.escape) {
            viewer.closeFind()
            return .handled
        }
        .accessibilityElement(children: .contain)
        .accessibilityLabel(Text(L10n.ArtifactViewer.findToggle))
        .accessibilityIdentifier("artifactViewer.findBar")
    }

    private var clearMenuButton: some View {
        PVButton(L10n.CitationComposer.clearMenu, variant: .link, size: .sm) {
            toggleClearMenu()
        }
        .disabled(model.isLocatorArtifactOnly || model.isDrawingDepiction)
        .accessibilityIdentifier("citationComposer.clearMenu")
        .pvContextMenu(
            $clearMenu,
            keyboard: $clearKeyboard,
            keyboardTitles: [
                L10n.string(L10n.CitationComposer.clearRegion),
                L10n.string(L10n.CitationComposer.resetToEntireArtifact),
            ],
            dismissOnClickAway: true
        ) {
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

    private func toggleClearMenu() {
        if clearMenu.isPresented {
            clearMenu.dismiss()
            return
        }
        clearMenu.present(at: CGPoint(x: 0, y: PVSpacing.controlHeightSmall))
    }

    private var locatorStrip: some View {
        HStack(spacing: PVSpacing.space3) {
            Text(L10n.CitationComposer.locatorEyebrow)
                .pvMicroCaps()
                .foregroundStyle(PVColor.textMuted)
                .padding(.trailing, PVSpacing.space2)
            locatorFloor
            if model.locatorCapabilities.supportsPageLocator, let page = model.locator.page {
                locatorChevron
                locatorChip(
                    icon: .file,
                    title: L10n.CitationComposer.locatorPage(page),
                    detail: model.artifactViewer.pageCount > 0
                        ? L10n.CitationComposer.pageOfCount(model.artifactViewer.pageCount)
                        : nil,
                    monoDetail: false,
                    accessibilityIdentifier: "citationComposer.locator.page",
                    onRemove: { model.removePage() }
                )
            }
            if model.locatorCapabilities.supportsRegionLocator, let region = model.locator.region {
                locatorChevron
                locatorChip(
                    icon: regionSymbol(region.kind),
                    title: regionNoun(region.kind),
                    detail: L10n.CitationComposer.locatorPoints(region.points.count),
                    monoDetail: true,
                    accessibilityIdentifier: "citationComposer.locator.region",
                    onRemove: { model.removeRegion() }
                )
            }
            Spacer(minLength: PVSpacing.space4)
            Text(L10n.CitationComposer.locatorDrawHint)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .lineLimit(1)
                .layoutPriority(-1)
        }
        .frame(height: 36)
        .padding(.leading, PVSpacing.space6)
        .padding(.trailing, PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            PVDivider(axis: .horizontal, color: PVColor.borderSubtle)
        }
        .accessibilityElement(children: .contain)
        .accessibilityIdentifier("citationComposer.locatorList")
    }

    private var locatorFloor: some View {
        HStack(spacing: PVSpacing.space2) {
            PVIcon(.scrollText, size: 14)
            Text(L10n.CitationComposer.locatorEntireArtifact)
                .font(PVFont.body(size: PVTypeScale.caption))
                .lineLimit(1)
        }
        .foregroundStyle(PVColor.textMuted)
        .accessibilityIdentifier("citationComposer.locator.artifact")
    }

    private var locatorChevron: some View {
        PVIcon(.chevronForward, size: 12)
            .foregroundStyle(PVColor.textFaint)
    }

    private func regionNoun(_ kind: ArtifactRegionKind) -> String {
        switch kind {
        case .rectangle:
            return L10n.string(L10n.CitationComposer.locatorRectangle)
        case .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft:
            return L10n.string(L10n.CitationComposer.locatorLShape)
        case .circle:
            return L10n.string(L10n.CitationComposer.locatorCircle)
        case .freeform:
            return L10n.string(L10n.CitationComposer.locatorPolygon)
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

    private func locatorChip(
        icon: PVSymbol,
        title: String,
        detail: String?,
        monoDetail: Bool,
        accessibilityIdentifier: String,
        onRemove: @escaping () -> Void
    ) -> some View {
        HStack(spacing: PVSpacing.space2) {
            PVIcon(icon, size: 12)
                .foregroundStyle(PVColor.textSecondary)
            Text(verbatim: title)
                .font(PVFont.body(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textPrimary)
                .lineLimit(1)
            if let detail {
                Text(verbatim: detail)
                    .font(monoDetail ? PVFont.mono(size: PVTypeScale.micro) : PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(1)
            }
            PVIconButton(.dismiss, label: L10n.CitationComposer.removeLocatorLayer, size: .sm, action: onRemove)
                .accessibilityIdentifier("\(accessibilityIdentifier).remove")
        }
        .padding(.leading, PVSpacing.space4)
        .padding(.trailing, PVSpacing.space1)
        .frame(height: PVSpacing.controlHeightSmall)
        .background(PVColor.surfaceSunken)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .stroke(PVColor.borderSubtle, lineWidth: 1)
        )
        .accessibilityIdentifier(accessibilityIdentifier)
    }
}
