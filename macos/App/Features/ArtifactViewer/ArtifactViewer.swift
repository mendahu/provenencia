import SwiftUI

/// Board Frames 1–2 page + zoom groups, plus Set page and the region radios.
struct ArtifactViewerToolChrome: View {
    @Bindable var model: ArtifactViewerModel

    var locatorPage: Int?
    var onSetPage: () -> Void
    /// False while a region is drawn: the Citation's page cannot move under it.
    var isSetPageEnabled: Bool
    /// The radios drawn, in order; the host narrows them (a depiction is a rectangle).
    var regionTools: [ArtifactRegionTool]
    @Binding var armedRegionTool: ArtifactRegionTool?

    @State private var pageFieldText = "1"

    var body: some View {
        HStack(spacing: 12) {
            if model.supportsPages {
                pageGroup
                setPageButton
                toolSep
            }
            if model.supportsSpatialZoom {
                zoomGroup
            }
            if model.locatorCapabilities.supportsRegionLocator {
                if model.supportsPages || model.supportsSpatialZoom {
                    toolSep
                }
                regionToolGroup
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
        return PVChip(
            text: isSet ? L10n.ArtifactViewer.pageSet(page: model.page) : L10n.string(L10n.ArtifactViewer.setPage),
            isSelected: isSet,
            tone: .accent,
            size: .sm,
            action: onSetPage
        )
        .disabled(!isSetPageEnabled)
        .accessibilityIdentifier("artifactViewer.setPage")
    }

    private var regionToolGroup: some View {
        HStack(spacing: 2) {
            ForEach(regionTools, id: \.self) { tool in
                regionRadio(tool)
            }
        }
    }

    private func regionRadio(_ tool: ArtifactRegionTool) -> some View {
        let selected = armedRegionTool == tool
        return PVIconButton(
            tool.symbol,
            label: label(for: tool),
            size: .sm,
            isSelected: selected
        ) {
            armedRegionTool = selected ? nil : tool
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
                    Text(verbatim: L10n.ArtifactViewer.zoomPercent(percent: Int((model.zoom * 100).rounded())))
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

/// Self-contained Artifact document canvas. The page / zoom / region strip is
/// ``ArtifactViewerToolChrome``, composed by the host.
struct ArtifactViewer: View {
    @Bindable var model: ArtifactViewerModel
    var locatorPage: Int?
    var armedRegionTool: ArtifactRegionTool?
    var committedRegion: ArtifactRegionDraft?
    var onCommitRegion: ((ArtifactRegionDraft) -> Void)?
    var onDisarmRegionTool: (() -> Void)?
    var freeformDeleteTooltip: String = ""

    var body: some View {
        canvas
            .frame(maxWidth: .infinity, maxHeight: .infinity)
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
                    findActivationID: model.findActivationID,
                    onUserSelectionChange: { model.setUserSelection(string: $0, page: $1) }
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
        case .imageTooLarge: L10n.ArtifactViewer.imageTooLargeTitle
        }
    }

    private var emptyMessage: LocalizedStringResource {
        switch model.emptyReason {
        case .idle: L10n.ArtifactViewer.emptyIdleMessage
        case .unsupportedKind: L10n.ArtifactViewer.unsupportedMessage
        case .mediaNotYetAvailable: L10n.ArtifactViewer.mediaComingSoonMessage
        case .missingFile: L10n.ArtifactViewer.missingFileMessage
        case .loadFailed: L10n.ArtifactViewer.loadFailedMessage
        case .imageTooLarge: L10n.ArtifactViewer.imageTooLargeMessage
        }
    }

    private var contentIdentity: AnyHashable {
        "\(model.kind.rawValue)-\(model.page)-\(model.pageCount)-\(model.displayImage?.size.width ?? model.pdfPageMediaSize?.width ?? 0)"
    }

    private var regionOverlayInput: ArtifactRegionOverlayInput {
        ArtifactRegionOverlayInput(
            armedTool: armedRegionTool,
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
