import SwiftUI

/// The Person page (S9-16, board S9-D5). A configuration of
/// `ConclusionDetailPage`: portrait, name, and the b. / d. lines. Life rows
/// stay stated empty until the birth and death Events arrive (S9-32).
struct PersonDetailView: View {
    let session: WorkspaceSession
    let entityId: String

    var body: some View {
        ConclusionDetailPage<PersonDetailContent, PersonDetailVitals>(
            session: session,
            entityId: entityId,
            mark: .subjectPerson,
            noLikeness: L10n.Conclusions.personNoLikeness,
            pageIdentifier: "person.detail",
            errorIdentifier: "person.detail.error",
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
                vitalLine(vital)
            }
        }
    }

    private func vitalLine(_ vital: PersonDetailContent.Vital) -> some View {
        HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space4) {
            Text(verbatim: vital.abbreviation)
                .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                .foregroundStyle(PVColor.textMuted)
            if let date = vital.date {
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
            if let place = vital.place {
                Text(verbatim: place)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textSecondary)
            } else {
                Text(L10n.Conclusions.personPlaceUnknown)
                    .font(PVFont.body(size: PVTypeScale.bodySmall, italic: true))
                    .foregroundStyle(PVColor.textMuted)
            }
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(vital.accessibilityLabel)
    }
}
