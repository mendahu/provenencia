import SwiftUI

/// The Events list (S9-23, board S9-D3). A configuration of
/// `ConclusionListPage`. The Events place, query key, and history entry stay
/// Events. Subject titles and places wait for S9-32.
struct EventsListView: View {
    let session: WorkspaceSession

    var body: some View {
        ConclusionListPage<CatalogEventHeader, EventDateLine>(
            session: session,
            key: .eventsList(project: session.projectKey),
            title: L10n.Workspace.eventsTitle,
            pageAccessibilityIdentifier: "events.list",
            emptyIcon: .calendar,
            emptyTitle: L10n.Workspace.eventsEmptyTitle,
            emptyMessage: L10n.Workspace.eventsEmptyMessage,
            emptyAccessibilityIdentifier: "events.empty",
            countMeta: L10n.Workspace.eventCount,
            refreshingMeta: L10n.Workspace.eventCountRefreshing,
            mark: .subjectEvent,
            ref: { $0.entity.ref },
            rowAccessibilityIdentifier: { "events.row.\($0.entity.ref)" },
            titleSource: { EventTitleDisplay.titleSource($0.titleParts) },
            accessibilityLabel: Self.rowLabel,
            location: Self.location,
            secondary: Self.dateLine
        )
    }

    private static func rowLabel(_ header: CatalogEventHeader) -> String {
        let source = EventTitleDisplay.titleSource(header.titleParts)
        let date = EventTitleDisplay.dateLine(
            date: header.date, start: header.startDate, end: header.endDate
        )
        return L10n.Workspace.eventRowAccessibility(title: source.text, date: date, ref: header.entity.ref)
    }

    private static func location(_ header: CatalogEventHeader) -> WorkspaceLocation {
        .eventDetail(
            entityId: header.entity.id,
            ref: header.entity.ref,
            title: EventTitleDisplay.title(header.titleParts)
        )
    }

    private static func dateLine(_ header: CatalogEventHeader) -> EventDateLine {
        EventDateLine(
            text: EventTitleDisplay.dateLine(
                date: header.date, start: header.startDate, end: header.endDate
            )
        )
    }
}

/// The Events row's date, mono, omitted when the line is empty. The date
/// does not share the title's truncation.
private struct EventDateLine: View {
    let text: String

    var body: some View {
        Text(verbatim: text)
            .font(PVFont.mono(size: PVTypeScale.caption))
            .fixedSize(horizontal: true, vertical: false)
            .frame(maxHeight: text.isEmpty ? 0 : nil)
            .accessibilityHidden(text.isEmpty)
    }
}
