import SwiftUI

/// Board Frames 1–2 page + zoom groups, plus optional Set page / region radios.
struct ArtifactViewerToolChrome: View {
    @Bindable var model: ArtifactViewerModel

    var locatorPage: Int?
    var onSetPage: (() -> Void)?
    var armedRegionTool: Binding<ArtifactRegionTool?>?

    @State private var pageFieldText = "1"

    var body: some View {
        HStack(spacing: 12) {
            if model.supportsPages {
                pageGroup
                if onSetPage != nil {
                    setPageButton
                }
                toolSep
            }
            if model.supportsSpatialZoom {
                zoomGroup
            }
            if model.locatorCapabilities.supportsRegionLocator, armedRegionTool != nil {
                if model.supportsPages || model.supportsSpatialZoom {
                    toolSep
                }
                regionTools
            }
        }
        .onAppear { syncPageField() }
        .onChange(of: model.page) { _, _ in syncPageField() }
        .onChange(of: model.pageCount) { _, _ in syncPageField() }
    }

    private var pageGroup: some View {
        HStack(spacing: 6) {
            PVIconButton(.chevronBack, label: L10n.ArtifactViewer.previousPage, size: .sm) {
                model.goToPreviousPage()
            }
            .disabled(model.page <= 1)
            .accessibilityIdentifier("artifactViewer.previousPage")

            TextField(
                text: $pageFieldText,
                prompt: Text(verbatim: "1")
            ) {
                EmptyView()
            }
            .font(PVFont.mono(size: PVTypeScale.caption))
            .foregroundStyle(PVColor.textPrimary)
            .multilineTextAlignment(.center)
            .frame(width: 34, height: 24)
            .background(PVColor.surfaceRaised)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .strokeBorder(PVColor.borderDefault, lineWidth: 1)
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
            .onSubmit { commitPageField() }
            .accessibilityLabel(Text(L10n.ArtifactViewer.pageField))
            .accessibilityIdentifier("artifactViewer.pageField")

            Text(verbatim: L10n.ArtifactViewer.pageOf(total: model.pageCount))
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)

            PVIconButton(.chevronForward, label: L10n.ArtifactViewer.nextPage, size: .sm) {
                model.goToNextPage()
            }
            .disabled(model.page >= model.pageCount)
            .accessibilityIdentifier("artifactViewer.nextPage")
        }
    }

    private var setPageButton: some View {
        let isSet = locatorPage == model.page
        return Button {
            onSetPage?()
        } label: {
            Text(isSet ? L10n.ArtifactViewer.pageSet(page: model.page) : L10n.ArtifactViewer.setPage)
                .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                .foregroundStyle(isSet ? PVColor.accentSoftForeground : PVColor.textPrimary)
                .padding(.horizontal, 8)
                .frame(height: 24)
                .background(
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .fill(isSet ? PVColor.accentSoft : PVColor.surfaceRaised)
                )
                .overlay(
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .strokeBorder(isSet ? PVColor.accentLine : PVColor.borderDefault, lineWidth: 1)
                )
        }
        .buttonStyle(.plain)
        .accessibilityIdentifier("artifactViewer.setPage")
    }

    @ViewBuilder
    private var regionTools: some View {
        if let armedRegionTool {
            HStack(spacing: 2) {
                ForEach(ArtifactRegionTool.allCases, id: \.self) { tool in
                    regionRadio(tool, selection: armedRegionTool)
                }
            }
        }
    }

    private func regionRadio(_ tool: ArtifactRegionTool, selection: Binding<ArtifactRegionTool?>) -> some View {
        let selected = selection.wrappedValue == tool
        return PVIconButton(
            tool.symbol,
            label: label(for: tool),
            size: .sm,
            isSelected: selected
        ) {
            selection.wrappedValue = selected ? nil : tool
        }
        .accessibilityIdentifier("artifactViewer.regionTool.\(tool.rawValue)")
    }

    private func label(for tool: ArtifactRegionTool) -> LocalizedStringResource {
        switch tool {
        case .rectangle: L10n.ArtifactViewer.regionRectangle
        case .lOpenTopRight: L10n.ArtifactViewer.regionLOpenTopRight
        case .lOpenTopLeft: L10n.ArtifactViewer.regionLOpenTopLeft
        case .lOpenBottomRight: L10n.ArtifactViewer.regionLOpenBottomRight
        case .lOpenBottomLeft: L10n.ArtifactViewer.regionLOpenBottomLeft
        case .circle: L10n.ArtifactViewer.regionCircle
        case .freeform: L10n.ArtifactViewer.regionFreeform
        }
    }

    private var zoomGroup: some View {
        HStack(spacing: 6) {
            PVIconButton(.zoomOut, label: L10n.ArtifactViewer.zoomOut, size: .sm) {
                model.zoomOut()
            }
            .disabled(model.zoom <= ArtifactViewerModel.minZoom + 0.001)
            .accessibilityIdentifier("artifactViewer.zoomOut")

            Text(verbatim: model.zoomPercentLabel)
                .font(PVFont.mono(size: PVTypeScale.caption))
                .foregroundStyle(PVColor.textSecondary)
                .frame(minWidth: 36)
                .accessibilityLabel(
                    Text(L10n.ArtifactViewer.zoomPercent(percent: Int((model.zoom * 100).rounded())))
                )

            PVIconButton(.zoomIn, label: L10n.ArtifactViewer.zoomIn, size: .sm) {
                model.zoomIn()
            }
            .disabled(model.zoom >= ArtifactViewerModel.maxZoom - 0.001)
            .accessibilityIdentifier("artifactViewer.zoomIn")
        }
    }

    private var toolSep: some View {
        Rectangle()
            .fill(PVColor.borderSubtle)
            .frame(width: 1, height: 18)
    }

    private func syncPageField() {
        pageFieldText = "\(model.page)"
    }

    private func commitPageField() {
        let trimmed = pageFieldText.trimmingCharacters(in: .whitespacesAndNewlines)
        if let value = Int(trimmed) {
            model.setPage(value)
        }
        syncPageField()
    }
}

/// Self-contained Artifact document canvas (+ optional embedded tool strip).
///
/// Prefer composing ``ArtifactViewerToolChrome`` in a host strip (Citation Composer)
/// and using `showsToolStrip: false` here.
struct ArtifactViewer: View {
    @Bindable var model: ArtifactViewerModel
    /// When true, paints page/zoom above the canvas (standalone hosts).
    var showsToolStrip: Bool = true
    var locatorPage: Int?
    var onSetPage: (() -> Void)?
    var armedRegionTool: Binding<ArtifactRegionTool?>?
    var committedRegion: ArtifactRegionDraft?
    var onCommitRegion: ((ArtifactRegionDraft) -> Void)?
    var onDisarmRegionTool: (() -> Void)?
    var freeformDeleteTooltip: String = ""

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if showsToolStrip,
               model.supportsPages || model.supportsSpatialZoom
                || model.locatorCapabilities.supportsRegionLocator
            {
                ArtifactViewerToolChrome(
                    model: model,
                    locatorPage: locatorPage,
                    onSetPage: onSetPage,
                    armedRegionTool: armedRegionTool
                )
                    .frame(height: 44)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 16)
                    .background(PVColor.surfaceCard)
                    .overlay(alignment: .bottom) {
                        Rectangle()
                            .fill(PVColor.borderSubtle)
                            .frame(height: 1)
                    }
            }
            canvas
                .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
    }

    @ViewBuilder
    private var canvas: some View {
        if model.isLoading {
            emptyChrome {
                ProgressView()
                    .controlSize(.small)
            }
        } else if model.kind == .pdf, let document = model.pdfDocument, model.supportsSpatialZoom {
            spatialCanvas {
                ArtifactPDFViewport(
                    document: document,
                    page: model.page,
                    zoom: model.zoom,
                    contentID: contentIdentity,
                    onZoomChange: { model.setZoom($0) },
                    overlayEnabled: model.locatorCapabilities.supportsRegionLocator,
                    overlayInput: regionOverlayInput,
                    onCommitRegion: onCommitRegion,
                    onDisarmRegionTool: onDisarmRegionTool,
                    freeformDeleteTooltip: freeformDeleteTooltip,
                    findSelection: model.activeFindSelection,
                    findActivationID: model.findActivationID
                )
            }
        } else if let image = model.displayImage, model.supportsSpatialZoom {
            spatialCanvas {
                ArtifactMediaViewport(
                    image: image,
                    zoom: model.zoom,
                    contentID: contentIdentity,
                    onZoomChange: { model.setZoom($0) },
                    overlayEnabled: model.locatorCapabilities.supportsRegionLocator,
                    overlayInput: regionOverlayInput,
                    onCommitRegion: onCommitRegion,
                    onDisarmRegionTool: onDisarmRegionTool,
                    freeformDeleteTooltip: freeformDeleteTooltip
                )
            }
        } else {
            emptyChrome {
                VStack(spacing: PVSpacing.space3) {
                    Text(emptyTitle)
                        .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textMuted)
                        .multilineTextAlignment(.center)
                    Text(emptyMessage)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textFaint)
                        .multilineTextAlignment(.center)
                }
                .padding(PVSpacing.space7)
            }
        }
    }

    private func emptyChrome<Content: View>(@ViewBuilder content: () -> Content) -> some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [6, 4]))
            .background(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(PVColor.surfaceSunken)
            )
            .overlay { content() }
    }

    private var emptyTitle: LocalizedStringResource {
        switch model.emptyReason {
        case .idle: L10n.ArtifactViewer.emptyIdleTitle
        case .unsupportedKind: L10n.ArtifactViewer.unsupportedTitle
        case .mediaNotYetAvailable: L10n.ArtifactViewer.mediaComingSoonTitle
        case .missingFile: L10n.ArtifactViewer.missingFileTitle
        case .loadFailed: L10n.ArtifactViewer.loadFailedTitle
        }
    }

    private var emptyMessage: LocalizedStringResource {
        switch model.emptyReason {
        case .idle: L10n.ArtifactViewer.emptyIdleMessage
        case .unsupportedKind: L10n.ArtifactViewer.unsupportedMessage
        case .mediaNotYetAvailable: L10n.ArtifactViewer.mediaComingSoonMessage
        case .missingFile: L10n.ArtifactViewer.missingFileMessage
        case .loadFailed: L10n.ArtifactViewer.loadFailedMessage
        }
    }

    private var contentIdentity: AnyHashable {
        "\(model.kind.rawValue)-\(model.page)-\(model.pageCount)-\(model.displayImage?.size.width ?? model.pdfPageMediaSize?.width ?? 0)"
    }

    private var regionOverlayInput: ArtifactRegionOverlayInput {
        ArtifactRegionOverlayInput(
            armedTool: armedRegionTool?.wrappedValue,
            committed: ArtifactRegionOverlayInput.committedOnCurrentPage(
                committedRegion,
                locatorPage: locatorPage,
                viewerPage: model.page,
                supportsPageLocator: model.locatorCapabilities.supportsPageLocator
            )
        )
    }

    private func spatialCanvas<Content: View>(@ViewBuilder content: () -> Content) -> some View {
        content()
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(PVColor.borderDefault, lineWidth: 1)
            )
            .accessibilityLabel(Text(L10n.ArtifactViewer.canvasAccessibility))
    }
}
