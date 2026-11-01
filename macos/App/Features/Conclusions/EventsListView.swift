import SwiftUI

/// The Events list (S9-23, board S9-D3). The Events place, query key, and
/// history entry stay Events. The title includes subjects; the secondary line
/// adds the place.
typealias EventsListView = ConclusionListPage<EventsList>

enum EventsList: ConclusionListKind {
    static func key(project: ProjectKey) -> CatalogQueryKey { .eventsList(project: project) }
    static var title: LocalizedStringResource { L10n.Workspace.eventsTitle }
    static var identifierRoot: String { "events" }
    static var emptyTitle: LocalizedStringResource { L10n.Workspace.eventsEmptyTitle }
    static var emptyMessage: LocalizedStringResource { L10n.Workspace.eventsEmptyMessage }
    static func countMeta(_ count: Int) -> String { L10n.Workspace.eventCount(count) }
    static func refreshingMeta(_ count: Int) -> String { L10n.Workspace.eventCountRefreshing(count) }
    static var mark: PVMarkKey { .subjectEvent }

    static func ref(_ summary: CatalogEventSummary) -> String { summary.entity.ref }
    static func updatedAt(_ summary: CatalogEventSummary) -> String { summary.updated.at }
    static func thumbnailRelPath(_ summary: CatalogEventSummary) -> String { summary.pictureImageRelPath }

    static func titleSource(_ summary: CatalogEventSummary) -> ConclusionTitleSource {
        titleSource(summary, orders: NameFormatOrders())
    }

    static func titleSource(_ summary: CatalogEventSummary, orders: NameFormatOrders) -> ConclusionTitleSource {
        EventTitleDisplay.titleSource(summary, orders: orders)
    }

    static func accessibilityLabel(_ summary: CatalogEventSummary) -> String {
        accessibilityLabel(summary, orders: NameFormatOrders())
    }

    static func accessibilityLabel(_ summary: CatalogEventSummary, orders: NameFormatOrders) -> String {
        let line = EventSecondaryDisplay.content(summary)
        var label = L10n.Workspace.eventRowAccessibility(title: titleSource(summary, orders: orders).text, date: line.date, ref: summary.entity.ref)
        if !line.place.isEmpty {
            label = L10n.DesignSystem.spokenJoin(label, line.place)
        }
        if line.extraPlaces > 0 {
            label = L10n.DesignSystem.spokenJoin(label, L10n.Conclusions.morePlaces(count: line.extraPlaces))
        }
        return label
    }

    static func location(_ summary: CatalogEventSummary) -> WorkspaceLocation {
        location(summary, orders: NameFormatOrders())
    }

    static func location(_ summary: CatalogEventSummary, orders: NameFormatOrders) -> WorkspaceLocation {
        .eventDetail(
            entityId: summary.entity.id,
            ref: summary.entity.ref,
            title: EventTitleDisplay.title(summary, orders: orders)
        )
    }

    @MainActor
    static func secondary(_ summary: CatalogEventSummary) -> EventSecondaryLine {
        let line = EventSecondaryDisplay.content(summary)
        return EventSecondaryLine(date: line.date, place: line.place, extraPlaces: line.extraPlaces)
    }
}

/// Date, place, and how many other places an Event row counts. Not a view,
/// so the list and Promote can share it off the main actor.
enum EventSecondaryDisplay {
    struct Content: Equatable, Sendable {
        var date: String
        var place: String
        var extraPlaces: Int
    }

    static func content(_ summary: CatalogEventSummary, locale: Locale = .autoupdatingCurrent) -> Content {
        Content(
            date: DateRowDisplay.line(date: summary.date, start: summary.startDate, end: summary.endDate, locale: locale),
            place: DerivedPlace.name(summary.places),
            extraPlaces: DerivedPlace.extra(summary.places)
        )
    }

    /// The same date and place on one line. Empty when the event has neither.
    static func line(_ summary: CatalogEventSummary, locale: Locale = .autoupdatingCurrent) -> String {
        let line = content(summary, locale: locale)
        if line.date.isEmpty { return line.place }
        if line.place.isEmpty { return line.date }
        return L10n.Conclusions.lineJoin(first: line.date, rest: line.place)
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
