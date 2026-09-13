import SwiftUI

/// Shared layout metrics for Source page section rows — replaces ad-hoc
/// `170` / `150` / `20 + space5 + …` paddings scattered through the page.
enum SourcePageLayout {
    /// Field-label column on metadata saved + suggestion rows (board: 132).
    static let metadataLabelWidth: CGFloat = 132
    /// Matches `PVReorderHandle`'s frame width.
    static let reorderHandleWidth: CGFloat = 20
    /// Spacing between handle, label, and value on metadata rows.
    static let metadataColumnSpacing = PVSpacing.space5
    /// Leading inset so suggestion labels line up with saved-row labels
    /// (saved rows reserve a reorder-handle column).
    static var metadataSuggestionLabelInset: CGFloat { reorderHandleWidth }
    /// Author / timestamp column on note rows and the composer byline.
    static let notesBylineWidth: CGFloat = 150
    /// Accordion chevron frame width in an artifact row.
    static let artifactChevronWidth: CGFloat = 18
    /// Thumbnail size in a collapsed artifact row.
    static let artifactRowThumbnailSize: CGFloat = 44
    /// Reserved trailing slot for “Use as thumbnail” so ART- refs stay aligned.
    static let useAsThumbnailSlotWidth: CGFloat = 156
    /// Trailing mono ART- ref column.
    static let artifactRefMinWidth: CGFloat = 72
    /// Indent for expanded artifact detail under chevron + thumbnail.
    static var artifactDetailLeading: CGFloat {
        artifactChevronWidth + artifactRowThumbnailSize
    }
}

/// Resting metadata value chrome that matches `PVTextArea` `.mono` insets so
/// entering inline edit does not shift the glyphs when the border appears.
struct SourcePageMetadataValueChrome: ViewModifier {
    func body(content: Content) -> some View {
        content
            .padding(.horizontal, PVInputChrome.horizontalInset)
            .padding(.vertical, PVTextArea.Typography.mono.verticalPadding)
            .overlay(
                RoundedRectangle(cornerRadius: PVRadius.sm, style: .continuous)
                    .stroke(Color.clear, lineWidth: 1)
            )
    }
}

extension View {
    func sourcePageMetadataValueChrome() -> some View {
        modifier(SourcePageMetadataValueChrome())
    }
}

/// Label column used by metadata saved / suggestion rows.
///
/// Saved rows pass `dataType` so the board's kind microcaps (icon + TEXT/DATE)
/// sit under the field name. Suggestions omit it and stay a muted caption.
struct SourcePageMetadataLabel: View {
    let text: String
    /// When set, renders the data-type microcaps under the label (saved rows).
    var dataType: String? = nil
    /// When true, pads leading by the reorder-handle gutter so suggestion
    /// labels align with saved-row labels.
    var alignWithReorderHandle: Bool = false
    var topPadding: CGFloat = 0

    var body: some View {
        Group {
            if let dataType {
                VStack(alignment: .leading, spacing: PVSpacing.space2) {
                    Text(text)
                        .font(PVFont.body(size: PVTypeScale.bodySmall, weight: PVFontWeight.semibold))
                        .foregroundStyle(PVColor.textPrimary)
                        .fixedSize(horizontal: false, vertical: true)
                    HStack(spacing: PVSpacing.space2) {
                        PVIcon(dataType == CatalogFieldDataType.date ? .calendar : .textType, size: 11)
                        Text(CatalogFieldDataType.label(for: dataType))
                            .font(PVFont.mono(size: PVTypeScale.micro))
                            .tracking(PVTypeScale.micro * PVTracking.caps)
                            .textCase(.uppercase)
                    }
                    .foregroundStyle(PVColor.textMuted)
                }
            } else {
                Text(text)
                    .font(PVFont.body(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .frame(width: SourcePageLayout.metadataLabelWidth, alignment: .leading)
        .padding(.leading, alignWithReorderHandle ? SourcePageLayout.metadataSuggestionLabelInset : 0)
        .padding(.top, topPadding)
    }
}
