import SwiftUI

/// Frame 8: choose which Artifact this Citation cites (multi-artifact Sources).
struct CitationComposerArtifactPicker: View {
    @Bindable var model: CitationComposerModel
    var onCancel: () -> Void

    private let cardWidth: CGFloat = 180
    private let thumbHeight: CGFloat = 150

    var body: some View {
        VStack(spacing: PVSpacing.space8) {
            VStack(spacing: PVSpacing.space3) {
                Text(L10n.CitationComposer.pickArtifactTitle)
                    .font(PVFont.display(size: PVTypeScale.h2, weight: PVFontWeight.medium))
                    .foregroundStyle(PVColor.textDisplay)
                    .multilineTextAlignment(.center)
                Text(verbatim: pickArtifactCaption)
                    .font(PVFont.body(size: PVTypeScale.body))
                    .foregroundStyle(PVColor.textSecondary)
                    .multilineTextAlignment(.center)
            }

            ViewThatFits(in: .horizontal) {
                HStack(alignment: .top, spacing: PVSpacing.space7) {
                    ForEach(model.artifacts, id: \.id) { artifact in
                        pickCard(artifact)
                    }
                }
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(alignment: .top, spacing: PVSpacing.space7) {
                        ForEach(model.artifacts, id: \.id) { artifact in
                            pickCard(artifact)
                        }
                    }
                    .padding(.horizontal, PVSpacing.space2)
                }
            }

            HStack(spacing: PVSpacing.space5) {
                PVButton(L10n.CitationComposer.cancel, variant: .ghost, size: .sm, action: onCancel)
                PVButton(L10n.CitationComposer.continuePick, variant: .primary, size: .sm) {
                    model.confirmArtifactSelection()
                }
                .disabled(model.pendingArtifactID == nil)
                .accessibilityIdentifier("citationComposer.pick.continue")
            }
        }
        .padding(PVSpacing.space9)
        .frame(maxWidth: .infinity, maxHeight: .infinity)
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
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                ArtifactPickThumb(
                    projectDir: model.session.projectKey.projectDir,
                    artifact: artifact,
                    typeIconKey: model.sourceTypeIconKey,
                    selected: selected,
                    width: cardWidth,
                    height: thumbHeight
                )

                Text(verbatim: displayName(artifact))
                    .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(2)
                    .frame(width: cardWidth, alignment: .leading)

                Text(verbatim: mediaCaption(artifact))
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
                    .lineLimit(2)
                    .frame(width: cardWidth, alignment: .leading)
            }
            .frame(width: cardWidth, alignment: .leading)
        }
        .buttonStyle(.plain)
        .accessibilityLabel(Text(verbatim: accessibilityLabel(for: artifact)))
        .accessibilityAddTraits(selected ? .isSelected : [])
        .accessibilityIdentifier("citationComposer.pick.artifact.\(artifact.id)")
    }

    private func displayName(_ artifact: CatalogArtifact) -> String {
        artifact.label.isEmpty ? artifact.ref : artifact.label
    }

    private func mediaCaption(_ artifact: CatalogArtifact) -> String {
        let fileless = artifact.file == nil || artifact.fileID.isEmpty
        if fileless {
            let detail = artifact.description.trimmingCharacters(in: .whitespacesAndNewlines)
            if detail.isEmpty {
                return String(localized: L10n.CitationComposer.mediaCaptionNoFile)
            }
            return L10n.CitationComposer.mediaCaptionNoFileDetail(detail: detail)
        }
        if CitationComposerModel.isPDF(artifact) {
            return String(localized: L10n.CitationComposer.mediaCaptionPDF)
        }
        if CitationComposerModel.isImage(artifact) {
            return String(localized: L10n.CitationComposer.mediaCaptionImage)
        }
        return artifact.ref
    }

    private func accessibilityLabel(for artifact: CatalogArtifact) -> String {
        "\(displayName(artifact)), \(mediaCaption(artifact))"
    }
}

/// Rectangular Artifact preview for the picker (Frame 8 thumb size).
private struct ArtifactPickThumb: View {
    let projectDir: String
    let artifact: CatalogArtifact
    let typeIconKey: String
    let selected: Bool
    let width: CGFloat
    let height: CGFloat

    @State private var image: NSImage?

    var body: some View {
        ZStack {
            PVColor.surfaceSunken
            content
        }
        .frame(width: width, height: height)
        .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
        .overlay(
            RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                .strokeBorder(
                    selected ? PVColor.accent : PVColor.borderDefault,
                    lineWidth: selected ? 2 : 1
                )
        )
        .pvShadow(selected ? PVElevation.md : [])
        .overlay(alignment: .topTrailing) {
            if selected {
                selectionBadge
                    .padding(PVSpacing.space3)
            }
        }
        .task(id: "\(projectDir)\0\(artifact.thumbnailRelPath)\0\(artifact.id)") {
            await loadImage()
        }
    }

    @ViewBuilder
    private var content: some View {
        if let image {
            Image(nsImage: image)
                .resizable()
                .scaledToFill()
                .frame(width: width, height: height)
                .clipped()
        } else if let mark = resolvedTypeMark, isFileless {
            PVMark(mark, size: 40, decorative: true)
                .foregroundStyle(PVColor.textMuted)
        } else if hasFileMetadata {
            PVMark(mimeMark, size: 40, decorative: true)
                .foregroundStyle(PVColor.textMuted)
        } else {
            Text(verbatim: "—")
                .font(PVFont.mono(size: PVTypeScale.body))
                .foregroundStyle(PVColor.textFaint)
                .accessibilityHidden(true)
        }
    }

    private var selectionBadge: some View {
        Circle()
            .fill(PVColor.accent)
            .frame(width: 20, height: 20)
            .overlay {
                PVIcon(.check, size: 12)
                    .foregroundStyle(PVColor.accentForeground)
            }
            .accessibilityHidden(true)
    }

    private var isFileless: Bool {
        artifact.file == nil || artifact.fileID.isEmpty
    }

    private var hasFileMetadata: Bool {
        let media = artifact.file?.mediaType ?? ""
        let name = artifact.file?.originalFilename ?? ""
        return !media.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || !name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var resolvedTypeMark: PVMarkKey? {
        let raw = typeIconKey.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !raw.isEmpty else { return nil }
        let key = PVMarkKey(catalogKey: raw)
        guard key.family == .type else { return nil }
        return key
    }

    private var mimeMark: PVMarkKey {
        PVFileTypeGlyph.key(
            mediaType: artifact.file?.mediaType,
            originalFilename: artifact.file?.originalFilename
        )
    }

    private func loadImage() async {
        let trimmed = artifact.thumbnailRelPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            image = nil
            return
        }
        image = await ThumbnailCache.shared.image(projectDir: projectDir, relPath: trimmed)
    }
}
