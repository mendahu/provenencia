import SwiftUI

/// The Events list (S9-23, board S9-D3). The Events place, query key, and
/// history entry stay Events. The title includes subjects; the secondary line
/// adds the place.
typealias EventsListView = ConclusionListPage<EventsList>

enum EventsList: ConclusionListKind {
    static func key(project: ProjectKey) -> CatalogQueryKey { .eventsList(project: project) }
    static var title: LocalizedStringResource { L10n.Workspace.eventsTitle }
    static var identifierRoot: String { "events" }
    static var emptyIcon: PVSymbol { .calendar }
    static var emptyTitle: LocalizedStringResource { L10n.Workspace.eventsEmptyTitle }
    static var emptyMessage: LocalizedStringResource { L10n.Workspace.eventsEmptyMessage }
    static func countMeta(_ count: Int) -> String { L10n.Workspace.eventCount(count) }
    static func refreshingMeta(_ count: Int) -> String { L10n.Workspace.eventCountRefreshing(count) }
    static var mark: PVMarkKey { .subjectEvent }

    static func ref(_ header: CatalogEventHeader) -> String { header.entity.ref }

    static func titleSource(_ header: CatalogEventHeader) -> ConclusionTitleSource {
        EventTitleDisplay.titleSource(header.title)
    }

    static func accessibilityLabel(_ header: CatalogEventHeader) -> String {
        let date = DateRowDisplay.line(date: header.date, start: header.startDate, end: header.endDate)
        var label = L10n.Workspace.eventRowAccessibility(title: titleSource(header).text, date: date, ref: header.entity.ref)
        let place = DerivedPlace.name(header.places)
        if !place.isEmpty {
            label = L10n.Conclusions.a11yList(label, rest: place)
        }
        let extra = DerivedPlace.extra(header.places)
        if extra > 0 {
            label = L10n.Conclusions.a11yList(label, rest: L10n.Conclusions.morePlaces(count: extra))
        }
        return label
    }

    static func location(_ header: CatalogEventHeader) -> WorkspaceLocation {
        .eventDetail(entityId: header.entity.id, ref: header.entity.ref, title: EventTitleDisplay.title(header.title))
    }

    @MainActor
    static func secondary(_ header: CatalogEventHeader) -> EventSecondaryLine {
        EventSecondaryLine(
            date: DateRowDisplay.line(date: header.date, start: header.startDate, end: header.endDate),
            place: DerivedPlace.name(header.places),
            extraPlaces: DerivedPlace.extra(header.places)
        )
    }
}

/// The Events row's date, then its place. The +N is other kept places.
struct EventSecondaryLine: View {
    let date: String
    let place: String
    let extraPlaces: Int

    /// Nothing when neither is recorded, so `PVList` lays out no second line.
    var body: some View {
        if !date.isEmpty || !place.isEmpty {
            HStack(spacing: PVSpacing.space4) {
                if !date.isEmpty {
                    Text(verbatim: date)
                        .font(PVFont.mono(size: PVTypeScale.caption))
                        .fixedSize(horizontal: true, vertical: false)
                }
                if !date.isEmpty && !place.isEmpty {
                    Text(verbatim: "·")
                        .foregroundStyle(PVColor.textFaint)
                }
                if !place.isEmpty {
                    Text(verbatim: place)
                        .italic()
                        .lineLimit(1)
                }
                if extraPlaces > 0 {
                    PVBadge(text: L10n.Conclusions.moreCount(extraPlaces), tone: .neutral, subtle: true)
                        .accessibilityHidden(true)
                }
            }
        }
    }
}
