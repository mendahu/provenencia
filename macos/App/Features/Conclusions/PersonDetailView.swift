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
        ConclusionDetailPage(
            session: session,
            entityId: entityId,
            pageIdentifier: "persons.detail",
            errorIdentifier: "persons.detail.error",
            make: { PersonDetailContent(detail: $0, orders: session.nameFormat) },
            summary: { PersonDetailLifeLines(lines: $0.lifeLines) },
            portrait: portrait
        )
        .vocabularyToastOverlay($portrait.toast, identifier: "persons.detail.toast")
    }
}

/// The two life lines under a Person's name.
private struct PersonDetailLifeLines: View {
    let lines: [PersonDetailContent.LifeLine]

    var body: some View {
        VStack(alignment: .leading, spacing: PVSpacing.space1) {
            ForEach(lines, id: \.abbreviation) { line in
                ConclusionDatePlaceLine(
                    leading: line.abbreviation,
                    date: line.date,
                    place: line.place,
                    accessibilityLabel: line.accessibilityLabel
                )
            }
        }
    }
}
