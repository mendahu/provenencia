import SwiftUI

/// The Event page (S9-24, board S9-D6 frame 2f). A configuration of
/// `ConclusionDetailPage`: the event mark, a date · place line, and the
/// Date and Place rows. The title and places come from the Event summary the
/// detail carries.
struct EventDetailView: View {
    let session: WorkspaceSession
    let entityId: String
    let userID: String

    var body: some View {
        ConclusionDetailPage(
            session: session,
            entityId: entityId,
            pageIdentifier: "events.detail",
            errorIdentifier: "events.detail.error",
            userID: userID,
            make: { EventDetailContent(detail: $0, orders: session.nameFormat) },
            secondary: { EventDetailSecondary(date: $0.secondaryDate, place: $0.secondaryPlace) }
        )
    }
}

/// Compact date, then the place, under the Event title.
private struct EventDetailSecondary: View {
    let date: String?
    let place: String?

    var body: some View {
        ConclusionDatePlaceLine(
            date: date,
            place: place,
            accessibilityLabel: L10n.DesignSystem.spokenJoin(
                date ?? L10n.string(L10n.Conclusions.dateUnknown),
                place ?? L10n.string(L10n.Conclusions.placeUnknown)
            )
        )
    }
}
