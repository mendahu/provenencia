import SwiftUI

/// Frame 7: choose which Artifact this Citation cites.
struct CitationComposerArtifactPicker: View {
    @Bindable var model: CitationComposerModel
    var onCancel: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            Spacer(minLength: PVSpacing.space9)
            VStack(alignment: .leading, spacing: PVSpacing.space7) {
                VStack(alignment: .leading, spacing: PVSpacing.space3) {
                    Text(L10n.CitationComposer.pickArtifactTitle)
                        .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textDisplay)
                    Text(verbatim: pickArtifactCaption)
                        .font(PVFont.body(size: PVTypeScale.body))
                        .foregroundStyle(PVColor.textSecondary)
                }
                LazyVGrid(
                    columns: [GridItem(.adaptive(minimum: 160), spacing: PVSpacing.space5)],
                    spacing: PVSpacing.space5
                ) {
                    ForEach(model.artifacts, id: \.id) { artifact in
                        pickCard(artifact)
                    }
                }
            }
            .padding(PVSpacing.space9)
            .frame(maxWidth: 720)
            .background(
                RoundedRectangle(cornerRadius: PVRadius.lg, style: .continuous)
                    .fill(PVColor.surfaceRaised)
            )
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.lg, style: .continuous)
                    .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
            )
            Spacer(minLength: PVSpacing.space9)
            HStack {
                PVButton(L10n.CitationComposer.cancel, variant: .secondary, action: onCancel)
                Spacer()
                PVButton(L10n.CitationComposer.continuePick, variant: .primary) {
                    model.confirmArtifactSelection()
                }
                .disabled(model.pendingArtifactID == nil)
                .accessibilityIdentifier("citationComposer.pick.continue")
            }
            .padding(PVSpacing.space7)
            .frame(maxWidth: 720)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .padding(.horizontal, PVSpacing.space9)
    }

    private var pickArtifactCaption: String {
        let title = model.sourceTitle.isEmpty ? "—" : model.sourceTitle
        return L10n.CitationComposer.pickArtifactCount(title: title, count: model.artifacts.count)
    }

    private func pickCard(_ artifact: CatalogArtifact) -> some View {
        let selected = model.pendingArtifactID == artifact.id
        return Button {
            model.selectPendingArtifact(artifact.id)
        } label: {
            VStack(alignment: .leading, spacing: PVSpacing.space3) {
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .fill(PVColor.surfaceSunken)
                    .frame(height: 100)
                    .overlay {
                        Text(verbatim: displayName(artifact))
                            .font(PVFont.body(size: PVTypeScale.caption))
                            .foregroundStyle(PVColor.textMuted)
                            .multilineTextAlignment(.center)
                            .padding(PVSpacing.space4)
                    }
                    .overlay(
                        RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                            .strokeBorder(
                                selected ? PVColor.accent : PVColor.borderDefault,
                                lineWidth: selected ? 2 : 1
                            )
                    )
                Text(verbatim: displayName(artifact))
                    .font(PVFont.body(size: PVTypeScale.caption, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(2)
                Text(verbatim: mediaCaption(artifact))
                    .font(PVFont.body(size: PVTypeScale.micro))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(verbatim: displayName(artifact)))
        .accessibilityAddTraits(selected ? .isSelected : [])
    }

    private func displayName(_ artifact: CatalogArtifact) -> String {
        artifact.label.isEmpty ? artifact.ref : artifact.label
    }

    private func mediaCaption(_ artifact: CatalogArtifact) -> String {
        if CitationComposerModel.isPDF(artifact) {
            return String(localized: L10n.CitationComposer.mediaCaptionPDF)
        }
        if CitationComposerModel.isImage(artifact) {
            return String(localized: L10n.CitationComposer.mediaCaptionImage)
        }
        return artifact.ref
    }
}
