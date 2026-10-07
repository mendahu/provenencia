import SwiftUI

/// The Persons list (S9-09, board S9-D2). A configuration of
/// `ConclusionListPage`; the place, query key, and history entry stay
/// Persons. The secondary line is the birth and death.
struct PersonsListView: View {
    let session: WorkspaceSession

    var body: some View {
        ConclusionListPage<CatalogPersonHeader, PersonLifeLine>(
            session: session,
            key: .personsList(project: session.projectKey),
            title: L10n.Workspace.personsTitle,
            pageAccessibilityIdentifier: "persons.list",
            emptyIcon: .person,
            emptyTitle: L10n.Workspace.personsEmptyTitle,
            emptyMessage: L10n.Workspace.personsEmptyMessage,
            emptyAccessibilityIdentifier: "persons.empty",
            countMeta: L10n.Workspace.personCount,
            refreshingMeta: L10n.Workspace.personCountRefreshing,
            mark: .subjectPerson,
            ref: { $0.entity.ref },
            rowAccessibilityIdentifier: { "persons.row.\($0.entity.ref)" },
            titleSource: { PersonHeaderDisplay.titleSource($0) },
            accessibilityLabel: Self.rowLabel,
            location: { header in
                .personDetail(
                    entityId: header.entity.id,
                    ref: header.entity.ref,
                    title: PersonHeaderDisplay.title(header)
                )
            },
            secondary: Self.lifeLine
        )
    }

    private static func rowLabel(_ header: CatalogPersonHeader) -> String {
        let title = ConclusionListRow.accessibilityLabel(
            PersonHeaderDisplay.titleSource(header), ref: header.entity.ref
        )
        let life = PersonLifeDisplay.line(header).text
        guard !life.isEmpty else { return title }
        return L10n.Conclusions.a11yList(title, rest: life)
    }

    private static func lifeLine(_ header: CatalogPersonHeader) -> PersonLifeLine {
        PersonLifeLine(line: PersonLifeDisplay.line(header))
    }
}

/// Birth and death under a Person's name. The +N is other kept places.
private struct PersonLifeLine: View {
    let line: PersonLifeDisplay.Line

    var body: some View {
        HStack(spacing: PVSpacing.space4) {
            Text(verbatim: line.text)
                .lineLimit(1)
                .frame(maxHeight: line.text.isEmpty ? 0 : nil)
            if line.extraPlaces > 0 {
                PVBadge(text: "+\(line.extraPlaces)", tone: .neutral, subtle: true)
                    .accessibilityHidden(true)
            }
        }
        .accessibilityHidden(line.text.isEmpty)
    }
}
