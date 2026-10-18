import SwiftUI

/// The Copy Source form's Artifact checklist: an "All Artifacts" row and one
/// row per Artifact, each with its thumbnail, label and ref. State is the set
/// of unchecked ids, so an Artifact that appears later starts checked.
/// A snowflake until the Graph form reuses it (promote then).
struct SourceArtifactChecklist: View {
    let artifacts: [CatalogArtifact]
    @Binding var excluded: Set<String>
    let projectDir: String
    let sourceTypeIconKey: String?
    let isDisabled: Bool

    private static let thumbnailSize: CGFloat = 32
    private static let maxRowsHeight: CGFloat = 264

    var body: some View {
        if artifacts.isEmpty {
            Text(L10n.Sources.copySourceNoArtifacts)
                .font(PVFont.body(size: PVTypeScale.caption, italic: true))
                .foregroundStyle(PVColor.textMuted)
                .accessibilityIdentifier("sources.copy.artifacts.empty")
        } else {
            VStack(alignment: .leading, spacing: PVSpacing.space4) {
                header
                VStack(spacing: 0) {
                    allRow
                    ScrollView {
                        VStack(spacing: 0) {
                            ForEach(artifacts, id: \.id) { artifact in
                                PVDivider()
                                row(artifact)
                            }
                        }
                    }
                    .frame(maxHeight: Self.maxRowsHeight)
                }
                .clipShape(RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous))
                .overlay {
                    RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                        .strokeBorder(PVColor.borderSubtle, lineWidth: 1)
                }
            }
            .accessibilityIdentifier("sources.copy.artifacts")
        }
    }

    private var header: some View {
        HStack {
            Text(L10n.Sources.copySourceArtifactsHeading)
                .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                .foregroundStyle(PVColor.textPrimary)
            Spacer(minLength: 0)
            Text(verbatim: L10n.Sources.copySourceArtifactCount(artifacts.count - excluded.count, artifacts.count))
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
        }
    }

    private var allRow: some View {
        PVCheckbox(
            L10n.Sources.copySourceAllArtifacts,
            isChecked: Binding(
                get: { excluded.isEmpty },
                set: { on in excluded = on ? [] : Set(artifacts.map(\.id)) }
            ),
            isIndeterminate: !excluded.isEmpty && excluded.count < artifacts.count,
            isDisabled: isDisabled
        )
        .accessibilityIdentifier("sources.copy.artifacts.all")
        .padding(.vertical, PVSpacing.space2)
        .padding(.trailing, PVSpacing.space5)
        // Row padding + thumbnail + gap, so this box lines up with the rows' boxes.
        .padding(.leading, PVSpacing.space5 + Self.thumbnailSize + PVSpacing.space5)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(PVColor.surfaceSunken)
    }

    private func row(_ artifact: CatalogArtifact) -> some View {
        HStack(spacing: PVSpacing.space5) {
            CachedThumbnail(
                projectDir: projectDir,
                relPath: artifact.thumbnailRelPath,
                mediaType: artifact.file?.mediaType ?? "",
                originalFilename: artifact.file?.originalFilename ?? "",
                typeIconKey: artifact.thumbnailTypeIconKey(sourceTypeIconKey: sourceTypeIconKey),
                size: Self.thumbnailSize
            )
            PVCheckbox(
                verbatim: artifact.hasDisplayLabel ? artifact.displayTitle : "",
                isChecked: Binding(
                    get: { !excluded.contains(artifact.id) },
                    set: { on in
                        if on { excluded.remove(artifact.id) } else { excluded.insert(artifact.id) }
                    }
                ),
                isDisabled: isDisabled
            )
            .accessibilityLabel(Text(verbatim: artifact.displayTitle))
            .accessibilityIdentifier("sources.copy.artifacts.\(artifact.id)")
            Spacer(minLength: 0)
            Text(verbatim: artifact.ref)
                .font(PVFont.mono(size: PVTypeScale.micro))
                .foregroundStyle(PVColor.textMuted)
        }
        .padding(.vertical, PVSpacing.space2)
        .padding(.horizontal, PVSpacing.space5)
    }
}
