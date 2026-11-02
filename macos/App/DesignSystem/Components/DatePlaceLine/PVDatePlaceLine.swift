import SwiftUI

/// A date · place line under a title: a Person's *b.* and *d.* lines, an
/// Event's summary. A missing half says it is unknown, or with
/// `missing: .omitted` is left out with its dot, and the line stays one line,
/// its place cut at the tail. One VoiceOver element with the caller's spoken
/// label. The date and place come already worded: this view draws, it does not
/// format.
struct PVDatePlaceLine: View {
    /// What a line says of a half that is not recorded. `.stated` is the
    /// header: both halves always, the missing one "unknown". `.omitted` is a
    /// tile or a list row: the missing half is left out, with its dot.
    enum Missing: Equatable { case stated, omitted }

    /// A leading abbreviation (*b.*, *d.*); nil for none.
    var leading: String?
    let date: String?
    let place: String?
    let accessibilityLabel: String
    var missing: Missing = .stated
    var dateSize: CGFloat = PVTypeScale.caption
    /// Places beyond the one shown; a "+N" badge when above zero.
    var extraCount: Int = 0

    init(
        leading: String? = nil,
        date: String?,
        place: String?,
        accessibilityLabel: String,
        missing: Missing = .stated,
        dateSize: CGFloat = PVTypeScale.caption,
        extraCount: Int = 0
    ) {
        self.leading = leading
        self.date = date
        self.place = place
        self.accessibilityLabel = accessibilityLabel
        self.missing = missing
        self.dateSize = dateSize
        self.extraCount = extraCount
    }

    /// The date half as spoken: the date, or the "date unknown" copy.
    static func spokenDate(_ date: String?) -> String {
        date ?? L10n.string(L10n.DesignSystem.datePlaceLineDateUnknown)
    }

    /// The place half as spoken: the place, or the "place unknown" copy.
    static func spokenPlace(_ place: String?) -> String {
        place ?? L10n.string(L10n.DesignSystem.datePlaceLinePlaceUnknown)
    }

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            if let leading {
                Text(verbatim: leading)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let date {
                Text(verbatim: date)
                    .font(PVFont.mono(size: dateSize))
                    .foregroundStyle(PVColor.textPrimary)
                    .lineLimit(missing == .omitted ? 1 : nil)
                    .layoutPriority(1)
            } else if missing == .stated {
                Text(L10n.DesignSystem.datePlaceLineDateUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if missing == .stated || (date != nil && place != nil) {
                Text(verbatim: "·")
                    .foregroundStyle(PVColor.textFaint)
            }
            if let place {
                Text(verbatim: place)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textSecondary)
                    .lineLimit(missing == .omitted ? 1 : nil)
                    .truncationMode(.tail)
            } else if missing == .stated {
                Text(L10n.DesignSystem.datePlaceLinePlaceUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if extraCount > 0 {
                PVBadge.moreCount(extraCount)
                    .layoutPriority(1)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(accessibilityLabel)
    }
}
