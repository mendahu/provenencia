import SwiftUI

/// A date · place line under a Conclusion's title: a Person's *b.* and *d.*
/// lines, an Event's summary. A missing half says it is unknown. One
/// VoiceOver element with the caller's spoken label.
struct ConclusionDatePlaceLine: View {
    /// A leading abbreviation (*b.*, *d.*); nil for none.
    var leading: String?
    let date: String?
    let place: String?
    let accessibilityLabel: String

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            if let leading {
                Text(verbatim: leading)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            if let date {
                Text(verbatim: date)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textPrimary)
            } else {
                Text(L10n.Conclusions.dateUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            Text(verbatim: "·")
                .foregroundStyle(PVColor.textFaint)
            if let place {
                Text(verbatim: place)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textSecondary)
            } else {
                Text(L10n.Conclusions.placeUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(accessibilityLabel)
    }
}
