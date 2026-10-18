import SwiftUI

/// The Person page (S9-16, board S9-D5). A configuration of
/// `ConclusionDetailPage`: portrait, name, and the b. / d. lines. Life dates
/// and places come from the Person header the detail carries.
struct PersonDetailView: View {
    let session: WorkspaceSession
    let entityId: String
    @State private var portrait: ConclusionPortraitModel

    init(session: WorkspaceSession, entityId: String, userID: String) {
        self.session = session
        self.entityId = entityId
        _portrait = State(
            initialValue: ConclusionPortraitModel(session: session, userID: userID, entityId: entityId)
        )
    }

    var body: some View {
        ConclusionDetailPage<PersonDetailContent, PersonDetailVitals>(
            session: session,
            entityId: entityId,
            pageIdentifier: "persons.detail",
            errorIdentifier: "persons.detail.error",
            make: { PersonDetailContent(detail: $0, orders: session.nameFormat) },
            summary: { PersonDetailVitals(vitals: $0.vitals) },
            portrait: portrait
        )
        .vocabularyToastOverlay($portrait.toast, identifier: "persons.detail.toast")
    }
}

/// The two life lines under a Person's name.
private struct PersonDetailVitals: View {
    let vitals: [PersonDetailContent.Vital]

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space1) {
            ForEach(vitals, id: \.abbreviation) { vital in
                ConclusionDatePlaceLine(
                    leading: vital.abbreviation,
                    date: vital.date,
                    place: vital.place,
                    accessibilityLabel: vital.accessibilityLabel
                )
            }
        }
    }
}
