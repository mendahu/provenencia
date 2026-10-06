import SwiftUI

/// The Event page (S9-24, board S9-D6 frame 2f). A configuration of
/// `ConclusionDetailPage`: the event mark, a date · place line, and the
/// Date and Place rows. Subjects and filled places arrive in S9-32.
struct EventDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    var body: some View {
        ConclusionDetailPage<EventDetailContent, EventDetailSummary>(
            session: session,
            entityId: entityId,
            mark: .subjectEvent,
            noLikeness: L10n.Conclusions.personNoLikeness,
            pageIdentifier: "events.detail",
            errorIdentifier: "events.detail.error",
            make: { EventDetailContent(detail: $0) },
            summary: { EventDetailSummary(date: $0.summaryDate) }
        )
    }
}

/// Compact date, then "place unknown", under the Event title.
private struct EventDetailSummary: View {
    let date: String?

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            if let date {
                Text(verbatim: date)
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textPrimary)
            } else {
                Text(L10n.Conclusions.personDateUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
            Text(verbatim: "·")
                .foregroundStyle(PVColor.textFaint)
            Text(L10n.Conclusions.personPlaceUnknown)
                .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                .foregroundStyle(PVColor.textMuted)
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(L10n.Conclusions.a11yList(
            date ?? L10n.string(L10n.Conclusions.personDateUnknown),
            rest: L10n.string(L10n.Conclusions.personPlaceUnknown)
        ))
    }
}
