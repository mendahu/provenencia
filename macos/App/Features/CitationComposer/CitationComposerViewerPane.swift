import SwiftUI

/// Left pane: Artifact header + board tool strip (page/zoom + Draw region) + canvas.
struct CitationComposerViewerPane: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            header
                .padding(.horizontal, PVSpacing.space7)
                .padding(.top, PVSpacing.space7)
                .padding(.bottom, PVSpacing.space3)

            toolStrip

            if let locatorError = model.locatorError, model.submitAttempted {
                PVCallout(tone: .danger, message: locatorError, compact: true)
                    .padding(.horizontal, PVSpacing.space7)
                    .padding(.top, PVSpacing.space3)
            } else if model.hasLocator {
                locatorCrumb
                    .padding(.horizontal, PVSpacing.space7)
                    .padding(.top, PVSpacing.space3)
            }

            ArtifactViewer(model: model.artifactViewer, showsToolStrip: false)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .padding(PVSpacing.space7)
                .onChange(of: model.artifactViewer.page) { _, _ in
                    if model.artifactViewer.supportsPages {
                        model.syncLocatorFromViewerPage()
                    }
                }
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
        let kind = model.isPDFArtifact
            ? String(localized: L10n.CitationComposer.artifactKindPDF)
            : (model.isImageArtifact
                ? String(localized: L10n.CitationComposer.artifactKindImage)
                : String(localized: L10n.CitationComposer.artifactKindUnknown))
        return L10n.CitationComposer.artifactIndexOf(
            index: index + 1,
            total: model.artifacts.count,
            kind: kind
        )
    }

    /// Board Frame 1/2: page + zoom (module) then Draw region / Clear (host stub).
    private var toolStrip: some View {
        HStack(spacing: 12) {
            if model.artifactViewer.supportsPages || model.artifactViewer.supportsSpatialZoom {
                ArtifactViewerToolChrome(model: model.artifactViewer)
                toolSep
            }

            PVButton(L10n.CitationComposer.drawRegion, variant: .secondary, size: .sm) {
                model.markWholeImageLocator()
            }
            .accessibilityIdentifier("citationComposer.drawRegion")

            Spacer(minLength: 0)

            if model.hasLocator {
                Button {
                    model.clearLocator()
                } label: {
                    Text(L10n.CitationComposer.clearLocator)
                        .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.accent)
                }
                .buttonStyle(.plain)
                .accessibilityIdentifier("citationComposer.clearLocator")
            }
        }
        .frame(height: 44)
        .padding(.horizontal, 16)
        .background(PVColor.surfaceCard)
        .overlay(alignment: .bottom) {
            Rectangle()
                .fill(PVColor.borderSubtle)
                .frame(height: 1)
        }
    }

    private var toolSep: some View {
        Rectangle()
            .fill(PVColor.borderSubtle)
            .frame(width: 1, height: 18)
    }

    private var locatorCrumb: some View {
        HStack(spacing: PVSpacing.space2) {
            if model.locatorIsWholeImage {
                Text(L10n.CitationComposer.locatorWholeImage)
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
            } else if let page = model.locatorPage {
                Text(verbatim: L10n.CitationComposer.locatorPage(page))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textSecondary)
            }
        }
    }
}
