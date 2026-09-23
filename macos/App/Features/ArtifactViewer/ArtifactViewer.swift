import SwiftUI

/// Board Frames 1–2 page + zoom groups (no Draw region — hosts append that).
struct ArtifactViewerToolChrome: View {
    @Bindable var model: ArtifactViewerModel

    @State private var pageFieldText = "1"

    var body: some View {
        HStack(spacing: 12) {
            if model.supportsPages {
                pageGroup
                toolSep
            }
            if model.supportsSpatialZoom {
                zoomGroup
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
/// Prefer composing ``ArtifactViewerToolChrome`` in a host strip with Draw region
/// (Citation Composer board Frame 1) and using `showsToolStrip: false` here.
struct ArtifactViewer: View {
    @Bindable var model: ArtifactViewerModel
    /// When true, paints page/zoom above the canvas (standalone hosts).
    var showsToolStrip: Bool = true

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            if showsToolStrip,
               model.supportsPages || model.supportsSpatialZoom
            {
                ArtifactViewerToolChrome(model: model)
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
        } else if let image = model.displayImage, model.supportsSpatialZoom {
            ArtifactMediaViewport(
                image: image,
                zoom: model.zoom,
                contentID: contentIdentity,
                onZoomChange: { model.setZoom($0) }
            )
            .clipShape(RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous))
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .strokeBorder(PVColor.borderDefault, lineWidth: 1)
            )
            .accessibilityLabel(Text(L10n.ArtifactViewer.canvasAccessibility))
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
        "\(model.kind.rawValue)-\(model.page)-\(model.pageCount)-\(model.displayImage?.size.width ?? 0)"
    }
}
