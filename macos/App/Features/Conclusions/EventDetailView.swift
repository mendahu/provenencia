import SwiftUI

/// The Event page (S9-24, board S9-D6 frame 2f). A configuration of
/// `ConclusionDetailPage`: the event mark, a date · place line, and the
/// Date and Place rows. The title and places come from the Event header the
/// detail carries.
struct EventDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    var body: some View {
        ConclusionDetailPage<EventDetailContent, EventDetailSummary>(
            session: session,
            entityId: entityId,
            mark: .subjectEvent,
            thumbnailLabel: L10n.Conclusions.eventNoImage,
            pageIdentifier: "events.detail",
            errorIdentifier: "events.detail.error",
            make: { EventDetailContent(detail: $0) },
            summary: { EventDetailSummary(date: $0.summaryDate, place: $0.summaryPlace) }
        )
    }
}

/// Compact date, then the place, under the Event title.
private struct EventDetailSummary: View {
    let date: String?
    let place: String?

    var body: some View {
        ConclusionDatePlaceLine(
            date: date,
            place: place,
            accessibilityLabel: L10n.Conclusions.a11yList(
                date ?? L10n.string(L10n.Conclusions.dateUnknown),
                rest: place ?? L10n.string(L10n.Conclusions.placeUnknown)
            )
        )
    }
}
