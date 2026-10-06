import SwiftUI

/// The Persons list (S9-09, board S9-D2). A configuration of
/// `ConclusionListPage`; the place, query key, and history entry stay
/// Persons. The secondary line stays empty until S9-32.
struct PersonsListView: View {
    let session: WorkspaceSession

    var body: some View {
        ConclusionListPage<CatalogPersonHeader, EmptyView>(
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
            accessibilityLabel: { header in
                ConclusionListRow.accessibilityLabel(
                    PersonHeaderDisplay.titleSource(header), ref: header.entity.ref
                )
            },
            location: { header in
                .personDetail(
                    entityId: header.entity.id,
                    ref: header.entity.ref,
                    title: PersonHeaderDisplay.title(header)
                )
            },
            secondary: { (_: CatalogPersonHeader) in EmptyView() }
        )
    }
}
