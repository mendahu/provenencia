import SwiftUI

/// Left pane: Artifact viewer chrome + locator stub (Frames 1 / 2 / 12).
struct CitationComposerViewerPane: View {
    @Bindable var model: CitationComposerModel

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space5) {
            header
            toolStrip
            if let locatorError = model.locatorError, model.submitAttempted {
                PVCallout(tone: .danger, message: locatorError, compact: true)
            } else if model.hasLocator {
                locatorCrumb
            }
            ZStack {
                canvas
                    .frame(maxWidth: 520, maxHeight: .infinity)
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .padding(PVSpacing.space7)
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

    private var toolStrip: some View {
        HStack(spacing: PVSpacing.space4) {
            if model.isPDFArtifact {
                HStack(spacing: PVSpacing.space2) {
                    PVIconButton(.chevronBack, label: L10n.CitationComposer.previousPage, size: .sm) {
                        model.goToPreviousPage()
                    }
                    .disabled(model.viewerPage <= 1)
                    Text(verbatim: "\(model.viewerPage)")
                        .font(PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .frame(minWidth: 20)
                    Text(verbatim: L10n.CitationComposer.pageOf(total: model.viewerPageCount))
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textMuted)
                    PVIconButton(.chevronForward, label: L10n.CitationComposer.nextPage, size: .sm) {
                        model.goToNextPage()
                    }
                    .disabled(model.viewerPage >= model.viewerPageCount)
                }
                toolSep
            }

            HStack(spacing: PVSpacing.space2) {
                Text(verbatim: "100%")
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
            toolSep

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
    }

    private var toolSep: some View {
        Rectangle()
            .fill(PVColor.borderSubtle)
            .frame(width: 1, height: 16)
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

    private var canvas: some View {
        RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
            .strokeBorder(PVColor.borderDefault, style: StrokeStyle(lineWidth: 1, dash: [6, 4]))
            .background(
                RoundedRectangle(cornerRadius: PVRadius.md, style: .continuous)
                    .fill(PVColor.surfaceSunken)
            )
            .overlay {
                VStack(spacing: PVSpacing.space3) {
                    Text(verbatim: canvasCaption)
                        .font(PVFont.body(size: PVTypeScale.body, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textMuted)
                        .multilineTextAlignment(.center)
                    Text(L10n.CitationComposer.viewerPlaceholderMessage)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textFaint)
                        .multilineTextAlignment(.center)
                }
                .padding(PVSpacing.space7)
            }
    }

    private var canvasCaption: String {
        if !model.hasLocator {
            return String(localized: L10n.CitationComposer.canvasNothingSelected)
        }
        if let artifact = model.selectedArtifact {
            return artifact.label.isEmpty ? artifact.ref : artifact.label
        }
        return String(localized: L10n.CitationComposer.viewerPlaceholderTitle)
    }
}
