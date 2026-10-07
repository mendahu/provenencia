import SwiftUI

/// The Events list (S9-23, board S9-D3). A configuration of
/// `ConclusionListPage`. The Events place, query key, and history entry stay
/// Events. The title includes subjects; the secondary line adds the place.
struct EventsListView: View {
    let session: WorkspaceSession

    var body: some View {
        ConclusionListPage<CatalogEventHeader, EventSecondaryLine>(
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
            titleSource: { EventTitleDisplay.titleSource($0.title) },
            accessibilityLabel: Self.rowLabel,
            location: Self.location,
            secondary: Self.secondary
        )
    }

    private static func rowLabel(_ header: CatalogEventHeader) -> String {
        let source = EventTitleDisplay.titleSource(header.title)
        let date = EventTitleDisplay.dateLine(
            date: header.date, start: header.startDate, end: header.endDate
        )
        var label = L10n.Workspace.eventRowAccessibility(title: source.text, date: date, ref: header.entity.ref)
        let place = DerivedPlace.name(header.places)
        if !place.isEmpty {
            label = L10n.Conclusions.a11yList(label, rest: place)
        }
        return label
    }

    private static func location(_ header: CatalogEventHeader) -> WorkspaceLocation {
        .eventDetail(
            entityId: header.entity.id,
            ref: header.entity.ref,
            title: EventTitleDisplay.title(header.title)
        )
    }

    private static func secondary(_ header: CatalogEventHeader) -> EventSecondaryLine {
        EventSecondaryLine(
            date: EventTitleDisplay.dateLine(
                date: header.date, start: header.startDate, end: header.endDate
            ),
            place: DerivedPlace.name(header.places),
            extraPlaces: DerivedPlace.extra(header.places)
        )
    }
}

/// The Events row's date, then its place. The +N is other kept places.
private struct EventSecondaryLine: View {
    let date: String
    let place: String
    let extraPlaces: Int

    var body: some View {
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
                PVBadge(text: "+\(extraPlaces)", tone: .neutral, subtle: true)
                    .accessibilityHidden(true)
            }
        }
        .frame(maxHeight: date.isEmpty && place.isEmpty ? 0 : nil)
        .accessibilityHidden(date.isEmpty && place.isEmpty)
    }
}
