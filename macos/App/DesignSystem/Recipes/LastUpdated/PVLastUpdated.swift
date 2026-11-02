import SwiftUI

/// The last-updated cell of a list row, shared by the Conclusion lists (S10-22)
/// and the Sources list (S10-41; it draws the cell on the row's second line):
/// the history icon and the relative time with the full stamp as help, a faint
/// dash when no time is recorded, and the spoken suffix of the row label.
/// `PVListDetailCell` draws it. `now` is read as each row builds (no timer).
enum PVLastUpdated {
    /// Width of the column, icon included. Layout, not formatting.
    static let columnWidth: CGFloat = 136

    /// The cell: relative text with the full stamp as help, or a faint dash
    /// when the time is empty or unparseable.
    static func detail(_ updatedAt: String, in context: TimestampFormat.Context) -> PVListDetail {
        let text = TimestampFormat.string(updatedAt, .relative, in: context)
        return PVListDetail(
            text: text ?? "—",
            help: TimestampFormat.string(updatedAt, .stamp, in: context).map(L10n.LastUpdated.help(stamp:)),
            icon: .history,
            faint: text == nil
        )
    }

    /// A row's spoken label with its last-updated suffix.
    static func label(_ base: String, updatedAt: String, in context: TimestampFormat.Context) -> String {
        L10n.DesignSystem.spokenJoin(
            base,
            TimestampFormat.string(updatedAt, .spoken, in: context).map(L10n.LastUpdated.a11yUpdated)
                ?? L10n.LastUpdated.a11yUnknown
        )
    }
}

#if DEBUG
#Preview("Known and unknown") {
    VStack(alignment: .leading, spacing: PVSpacing.space4) {
        PVListDetailCell(
            detail: PVLastUpdated.detail("2026-10-14T21:15:00Z", in: .current),
            width: PVLastUpdated.columnWidth
        )
        PVListDetailCell(
            detail: PVLastUpdated.detail("", in: .current),
            width: PVLastUpdated.columnWidth
        )
    }
    .padding(PVSpacing.space6)
    .background(PVColor.surfaceCard)
}
#endif
