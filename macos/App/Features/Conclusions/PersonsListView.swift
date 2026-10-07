import SwiftUI

/// The Persons list (S9-09, board S9-D2). The place, query key, and history
/// entry stay Persons. The secondary line is the birth and death.
typealias PersonsListView = ConclusionListPage<PersonsList>

enum PersonsList: ConclusionListKind {
    static func key(project: ProjectKey) -> CatalogQueryKey { .personsList(project: project) }
    static var title: LocalizedStringResource { L10n.Workspace.personsTitle }
    static var identifierRoot: String { "persons" }
    static var emptyIcon: PVSymbol { .person }
    static var emptyTitle: LocalizedStringResource { L10n.Workspace.personsEmptyTitle }
    static var emptyMessage: LocalizedStringResource { L10n.Workspace.personsEmptyMessage }
    static func countMeta(_ count: Int) -> String { L10n.Workspace.personCount(count) }
    static func refreshingMeta(_ count: Int) -> String { L10n.Workspace.personCountRefreshing(count) }
    static var mark: PVMarkKey { .subjectPerson }

    static func ref(_ header: CatalogPersonHeader) -> String { header.entity.ref }

    static func titleSource(_ header: CatalogPersonHeader) -> ConclusionTitleSource {
        PersonHeaderDisplay.titleSource(header)
    }

    static func accessibilityLabel(_ header: CatalogPersonHeader) -> String {
        let title = ConclusionListRow.accessibilityLabel(titleSource(header), ref: header.entity.ref)
        let line = PersonLifeDisplay.line(header)
        guard !line.text.isEmpty else { return title }
        let label = L10n.Conclusions.a11yList(title, rest: line.text)
        guard line.extraPlaces > 0 else { return label }
        return L10n.Conclusions.a11yList(label, rest: L10n.Conclusions.morePlaces(count: line.extraPlaces))
    }

    static func location(_ header: CatalogPersonHeader) -> WorkspaceLocation {
        .personDetail(entityId: header.entity.id, ref: header.entity.ref, title: PersonHeaderDisplay.title(header))
    }

    @MainActor
    static func secondary(_ header: CatalogPersonHeader) -> PersonLifeLine {
        PersonLifeLine(line: PersonLifeDisplay.line(header))
    }
}

/// Birth and death under a Person's name. The +N is other kept places.
struct PersonLifeLine: View {
    let line: PersonLifeDisplay.Line

    /// Nothing when nothing is recorded, so `PVList` lays out no second line.
    var body: some View {
        if !line.text.isEmpty {
            HStack(spacing: PVSpacing.space4) {
                Text(verbatim: line.text)
                    .lineLimit(1)
                if line.extraPlaces > 0 {
                    PVBadge(text: L10n.Conclusions.moreCount(line.extraPlaces), tone: .neutral, subtle: true)
                        .accessibilityHidden(true)
                }
            }
        }
    }
}
