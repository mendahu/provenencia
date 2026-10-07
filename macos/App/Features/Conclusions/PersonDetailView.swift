import SwiftUI

/// The Person page (S9-16, board S9-D5). A configuration of
/// `ConclusionDetailPage`: portrait, name, and the b. / d. lines. Life dates
/// and places come from the Person header the detail carries.
struct PersonDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    var body: some View {
        ConclusionDetailPage<PersonDetailContent, PersonDetailVitals>(
            session: session,
            entityId: entityId,
            mark: .subjectPerson,
            thumbnailLabel: L10n.Conclusions.personNoLikeness,
            pageIdentifier: "persons.detail",
            errorIdentifier: "persons.detail.error",
            make: { PersonDetailContent(detail: $0) },
            summary: { PersonDetailVitals(vitals: $0.vitals) }
        )
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
