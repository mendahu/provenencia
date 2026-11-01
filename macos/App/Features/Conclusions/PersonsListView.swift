import SwiftUI

/// The Persons list (S9-09, board S9-D2). The place, query key, and history
/// entry stay Persons. The secondary line is the birth and death.
typealias PersonsListView = ConclusionListPage<PersonsList>

enum PersonsList: ConclusionListKind {
    static func key(project: ProjectKey) -> CatalogQueryKey { .personsList(project: project) }
    static var title: LocalizedStringResource { L10n.Workspace.personsTitle }
    static var identifierRoot: String { "persons" }
    static var emptyTitle: LocalizedStringResource { L10n.Workspace.personsEmptyTitle }
    static var emptyMessage: LocalizedStringResource { L10n.Workspace.personsEmptyMessage }
    static func countMeta(_ count: Int) -> String { L10n.Workspace.personCount(count) }
    static func refreshingMeta(_ count: Int) -> String { L10n.Workspace.personCountRefreshing(count) }
    static var mark: PVMarkKey { .subjectPerson }

    static func ref(_ summary: CatalogPersonSummary) -> String { summary.entity.ref }
    static func updatedAt(_ summary: CatalogPersonSummary) -> String { summary.updated.at }
    static func thumbnailRelPath(_ summary: CatalogPersonSummary) -> String { summary.pictureImageRelPath }

    static func titleSource(_ summary: CatalogPersonSummary) -> ConclusionTitleSource {
        titleSource(summary, orders: NameFormatOrders())
    }

    static func titleSource(_ summary: CatalogPersonSummary, orders: NameFormatOrders) -> ConclusionTitleSource {
        PersonTitleDisplay.titleSource(summary, style: .sorted, orders: orders)
    }

    static func accessibilityLabel(_ summary: CatalogPersonSummary) -> String {
        accessibilityLabel(summary, orders: NameFormatOrders())
    }

    static func accessibilityLabel(_ summary: CatalogPersonSummary, orders: NameFormatOrders) -> String {
        let title = ConclusionListRow.accessibilityLabel(titleSource(summary, orders: orders), ref: summary.entity.ref)
        let line = PersonLifeDisplay.line(summary)
        guard !line.text.isEmpty else { return title }
        let label = L10n.DesignSystem.spokenJoin(title, line.text)
        guard line.extraPlaces > 0 else { return label }
        return L10n.DesignSystem.spokenJoin(label, L10n.Conclusions.morePlaces(count: line.extraPlaces))
    }

    static func location(_ summary: CatalogPersonSummary) -> WorkspaceLocation {
        location(summary, orders: NameFormatOrders())
    }

    static func location(_ summary: CatalogPersonSummary, orders: NameFormatOrders) -> WorkspaceLocation {
        .personDetail(
            entityId: summary.entity.id,
            ref: summary.entity.ref,
            title: PersonTitleDisplay.title(summary, style: .sorted, orders: orders)
        )
    }

    @MainActor
    static func secondary(_ summary: CatalogPersonSummary) -> PersonLifeLine {
        PersonLifeLine(line: PersonLifeDisplay.line(summary))
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
