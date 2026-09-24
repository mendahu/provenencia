import SwiftUI

/// Artifact identity row inside the in-form context menu (file mark + count).
struct CitationComposerArtifactMenuRow: View {
    let artifact: CatalogArtifact
    let citationCount: Int
    var isSelected: Bool
    var sourceTypeIconKey: String = ""
    var index: Int = 0
    var action: () -> Void

    @Environment(\.pvContextMenuDismiss) private var dismiss

    var body: some View {
        Button {
            dismiss()
            action()
        } label: {
            HStack(alignment: .center, spacing: 10) {
                PVMark(markKey, size: 22, decorative: true)
                VStack(alignment: .leading, spacing: 1) {
                    Text(verbatim: artifact.label)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.medium))
                        .foregroundStyle(PVColor.textPrimary)
                        .lineLimit(1)
                    Text(verbatim: metaLine)
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                        .lineLimit(1)
                }
                Spacer(minLength: 0)
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.vertical, PVSpacing.space3)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .background(isSelected ? PVColor.surfaceSunken : Color.clear)
        .accessibilityIdentifier("citationComposer.artifact.row.\(artifact.id)")
    }

    private var markKey: PVMarkKey {
        PVFileTypeGlyph.key(
            mediaType: artifact.file?.mediaType,
            originalFilename: artifact.file?.originalFilename
        )
    }

    private var metaLine: String {
        L10n.CitationComposer.artifactMenuMeta(
            kind: kindLabel,
            count: citationCount
        )
    }

    private var kindLabel: String {
        let media = artifact.file?.mediaType ?? ""
        if media.localizedCaseInsensitiveContains("pdf") {
            return String(localized: L10n.CitationComposer.artifactKindPDF)
        }
        if media.hasPrefix("image/") {
            return String(localized: L10n.CitationComposer.artifactKindImage)
        }
        if media.hasPrefix("text/") {
            return String(localized: L10n.CitationComposer.artifactKindUnknown)
        }
        return String(localized: L10n.CitationComposer.artifactKindUnknown)
    }
}

/// Citation identity row: ref + observation count + transcription snippet.
struct CitationComposerCitationMenuRow: View {
    let listed: CatalogListedCitation
    var isSelected: Bool
    var index: Int
    var action: () -> Void

    @Environment(\.pvContextMenuDismiss) private var dismiss

    var body: some View {
        Button {
            dismiss()
            action()
        } label: {
            VStack(alignment: .leading, spacing: 2) {
                HStack(spacing: 8) {
                    Text(verbatim: listed.citation.ref)
                        .font(PVFont.mono(size: PVTypeScale.caption, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                    Text(verbatim: L10n.CitationComposer.citationMenuCount(count: listed.observationCount))
                        .font(PVFont.body(size: PVTypeScale.micro))
                        .foregroundStyle(PVColor.textMuted)
                    Spacer(minLength: 0)
                }
                if !snippet.isEmpty {
                    Text(verbatim: snippet)
                        .font(PVFont.body(size: PVTypeScale.caption))
                        .foregroundStyle(PVColor.textSecondary)
                        .lineLimit(2)
                }
            }
            .padding(.horizontal, PVSpacing.space5)
            .padding(.vertical, PVSpacing.space3)
            .frame(maxWidth: .infinity, alignment: .leading)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .background(isSelected ? PVColor.surfaceSunken : Color.clear)
        .accessibilityIdentifier("citationComposer.citation.row.\(listed.id)")
    }

    private var snippet: String {
        listed.citation.transcription
            .replacingOccurrences(of: "\n", with: " — ")
            .trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
