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

    static func titleSource(_ summary: CatalogEventSummary) -> PVTitleSource {
        titleSource(summary, orders: NameFormatOrders())
    }

    static func titleSource(_ summary: CatalogEventSummary, orders: NameFormatOrders) -> PVTitleSource {
        EventTitleDisplay.titleSource(summary, orders: orders)
    }

    static func accessibilityLabel(_ summary: CatalogEventSummary) -> String {
        accessibilityLabel(summary, orders: NameFormatOrders())
    }

    static func accessibilityLabel(
        _ summary: CatalogEventSummary,
        orders: NameFormatOrders,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        let line = EventSecondaryDisplay.content(summary, locale: locale)
        let head = L10n.Workspace.eventRowAccessibility(
            title: EventTitleDisplay.titleSource(summary, locale: locale, orders: orders).text,
            date: line.date,
            ref: summary.entity.ref
        )
        let tail = placeTail(line)
        return tail.isEmpty ? head : L10n.DesignSystem.spokenJoin(head, tail)
    }

    /// The event as one spoken line without its ref: title, date, place and
    /// "+N places", each only when there is something to say. Another surface
    /// that names the event (a Timeline row) adds its own tail.
    static func spokenLine(
        _ summary: CatalogEventSummary,
        orders: NameFormatOrders,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        let line = EventSecondaryDisplay.content(summary, locale: locale)
        let parts = [
            EventTitleDisplay.titleSource(summary, locale: locale, orders: orders).text,
            line.date,
            placeTail(line),
        ]
        return L10n.DesignSystem.spokenJoin(parts) ?? ""
    }

    /// The place, then "N more places", spoken; empty when there is neither.
    private static func placeTail(_ line: EventSecondaryDisplay.Content) -> String {
        var parts: [String] = []
        if !line.place.isEmpty { parts.append(line.place) }
        if line.extraPlaces > 0 { parts.append(L10n.Conclusions.morePlaces(count: line.extraPlaces)) }
        return L10n.DesignSystem.spokenJoin(parts) ?? ""
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
        EventSecondaryLine(content: EventSecondaryDisplay.content(summary))
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
        line(content(summary, locale: locale))
    }

    /// The date and place of an already-built `Content` on one line.
    static func line(_ content: Content) -> String {
        PVDatePlaceLine.text(date: content.date, place: content.place)
    }
}

/// The Events row's date, then its place, on the Person header's line. The +N is
/// other kept places. The row's own label speaks it, so the line is hidden from VoiceOver.
struct EventSecondaryLine: View {
    let content: EventSecondaryDisplay.Content
    /// The date's size; the Events row's caption size unless a caller (the Timeline's
    /// event row) draws the line larger.
    var dateSize: CGFloat = PVTypeScale.caption

    /// Nothing when neither is recorded, so `PVList` lays out no second line.
    var body: some View {
        if !content.date.isEmpty || !content.place.isEmpty {
            PVDatePlaceLine(
                date: content.date.isEmpty ? nil : content.date,
                place: content.place.isEmpty ? nil : content.place,
                accessibilityLabel: EventSecondaryDisplay.line(content),
                missing: .omitted,
                dateSize: dateSize,
                dateTone: PVColor.textMuted,
                extraCount: content.extraPlaces
            )
            .accessibilityHidden(true)
        }
    }
}
