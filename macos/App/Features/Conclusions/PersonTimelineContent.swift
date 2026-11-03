import Foundation

/// What a Person's Timeline section shows (S13-13, board S13-D4): the events in
/// the one display order the list and the lifeline both read, with the birth and
/// death markers placed and the long lists folded. Pure; the section draws it.
///
/// Go gives the rows in date order, undated last, with each row's span and the
/// markers (S13-33); this places the markers and the undated group and folds,
/// and never sorts or dates anything itself.
struct PersonTimelineContent: Equatable {
    /// Dated events shown before "Show N more", and likewise the undated ones.
    static let eventFoldLimit = 50
    /// Other people shown on one event before "Show N more".
    static let participantFoldLimit = 15

    /// A list that can be folded; a row's participant fold is drawn inside the row.
    enum Fold: Hashable {
        case dated, undated
        case participants(eventId: String)
    }

    /// One role of the person on an event, drawn as a Badge; `isSubject` draws
    /// the outline.
    struct RoleBadge: Equatable {
        var text: String
        var isSubject: Bool
    }

    struct Row: Equatable, Identifiable {
        /// The event id.
        var id: String
        var title: PVTitleSource
        var line: EventSecondaryDisplay.Content
        var imageRelPath: String
        var roles: [RoleBadge]
        var location: WorkspaceLocation
        var span: CatalogTimelineSpan?
        /// The other people shown, in the read's order (most important role first).
        var participants: [PersonCardLink]
        var participantsHidden: Int
        /// "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject".
        var spokenLabel: String
    }

    struct Marker: Equatable {
        enum Kind: Equatable { case birth, death }

        var kind: Kind
        /// The recorded event's row; nil for a placeholder.
        var row: Row?
        var spokenLabel: String
    }

    enum Item: Equatable, Identifiable {
        case marker(Marker)
        case row(Row)
        case undatedHeader
        /// A fold of `.dated` or `.undated` rows.
        case more(Fold, count: Int)

        /// A row's and a recorded marker's id is the event id; a placeholder's is
        /// "birth" or "death".
        var id: String {
            switch self {
            case .marker(let marker):
                marker.row?.id ?? (marker.kind == .birth ? "birth" : "death")
            case .row(let row): row.id
            case .undatedHeader: "undated"
            case .more(let fold, _): fold == .dated ? "more.dated" : "more.undated"
            }
        }

        /// The row behind a row or a recorded marker.
        var eventRow: Row? {
            switch self {
            case .marker(let marker): marker.row
            case .row(let row): row
            case .undatedHeader, .more: nil
            }
        }
    }

    /// Every dated row's span, folded or not, for the lifeline.
    struct Span: Equatable {
        var id: String
        var span: CatalogTimelineSpan
    }

    private(set) var items: [Item] = []
    private(set) var allSpans: [Span] = []
    let isEmpty: Bool

    init(
        timeline: CatalogPersonTimeline,
        expanded: Set<Fold> = [],
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) {
        isEmpty = timeline.events.isEmpty
        let rows = timeline.events.map { Self.row($0, expanded: expanded, locale: locale, orders: orders) }
        allSpans = rows.compactMap { row in row.span.map { Span(id: row.id, span: $0) } }
        guard !rows.isEmpty else { return }
        let rowsByID = Dictionary(rows.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })

        func marker(_ kind: Marker.Kind, _ source: CatalogLifeMarker) -> Marker {
            if let id = source.eventId, let row = rowsByID[id] {
                return Marker(kind: kind, row: row, spokenLabel: row.spokenLabel)
            }
            return Marker(
                kind: kind,
                row: nil,
                spokenLabel: L10n.string(kind == .birth ? L10n.Conclusions.timelineBirthUnknown : L10n.Conclusions.timelineDeathUnknown)
            )
        }
        let birth = marker(.birth, timeline.birth)
        let death = marker(.death, timeline.death)

        // (1) The dated rows in the read's order; a recorded marker with a span
        // is its row, in place.
        var dated: [Item] = rows.filter { $0.span != nil }.map { row in
            if birth.row?.id == row.id { return .marker(birth) }
            if death.row?.id == row.id { return .marker(death) }
            return .row(row)
        }
        // (2), (3) Any other marker goes beside the row it names: the birth
        // before it, the death after it; with none, first and last.
        let besideBirth = timeline.birth.besideEventId
        let besideDeath = timeline.death.besideEventId
        if birth.row?.span == nil {
            let at = besideBirth.flatMap { id in dated.firstIndex { $0.id == id } } ?? 0
            dated.insert(.marker(birth), at: at)
        }
        if death.row?.span == nil {
            let at = besideDeath.flatMap { id in dated.firstIndex { $0.id == id }.map { $0 + 1 } } ?? dated.count
            dated.insert(.marker(death), at: at)
        }
        let deathAt = dated.firstIndex { item in
            if case .marker(let marker) = item { return marker.kind == .death }
            return false
        } ?? dated.count - 1

        // (4) The undated rows, less a recorded birth or death (it is its
        // marker), go immediately before the death marker.
        let undated = rows.filter { row in
            row.span == nil && row.id != birth.row?.id && row.id != death.row?.id
        }

        var before = Array(dated[..<deathAt])
        let after = Array(dated[(deathAt + 1)...])

        // Folds: the dated rows before the death marker, a recorded birth's row
        // among them; a birth placeholder is never folded.
        if !expanded.contains(.dated) {
            var shown = 0
            var folded: [Item] = []
            var hidden = 0
            for item in before {
                let counts = item.eventRow != nil
                if counts && shown >= Self.eventFoldLimit {
                    if hidden == 0 { folded.append(.more(.dated, count: 0)) }
                    hidden += 1
                } else {
                    if counts { shown += 1 }
                    folded.append(item)
                }
            }
            if hidden > 0, let at = folded.firstIndex(where: { $0.id == "more.dated" }) {
                folded[at] = .more(.dated, count: hidden)
            }
            before = folded
        }

        var undatedItems: [Item] = []
        if !undated.isEmpty {
            undatedItems.append(.undatedHeader)
            let limit = expanded.contains(.undated) ? undated.count : Self.eventFoldLimit
            undatedItems += undated.prefix(limit).map { .row($0) }
            if undated.count > limit { undatedItems.append(.more(.undated, count: undated.count - limit)) }
        }

        items = before + undatedItems + [dated[deathAt]] + after
    }

    private static func row(
        _ event: CatalogPersonEvent,
        expanded: Set<Fold>,
        locale: Locale,
        orders: NameFormatOrders
    ) -> Row {
        let summary = event.event
        let roles = event.roles.compactMap { value -> RoleBadge? in
            PropertyTermDisplay.name(value, propertyKey: SeededPropertyKey.role).map {
                RoleBadge(text: $0, isSubject: value.productKey == SeededRoleTerm.subject)
            }
        }
        var spoken = EventsList.spokenLine(summary, orders: orders, locale: locale)
        if let joined = L10n.DesignSystem.spokenJoin(roles.map(\.text)) {
            spoken = L10n.DesignSystem.spokenJoin(spoken, L10n.Conclusions.timelineAsRole(joined))
        }
        let fold = Fold.participants(eventId: summary.id)
        let shown = expanded.contains(fold) ? event.participants : Array(event.participants.prefix(participantFoldLimit))
        return Row(
            id: summary.id,
            title: EventTitleDisplay.titleSource(summary, locale: locale, orders: orders),
            line: EventSecondaryDisplay.content(summary, locale: locale),
            imageRelPath: summary.pictureImageRelPath,
            roles: roles,
            location: EventsList.location(summary, orders: orders),
            span: event.span,
            participants: shown.map { participant in
                let names = participant.roles.compactMap {
                    PropertyTermDisplay.name($0, propertyKey: SeededPropertyKey.role)
                }
                return PersonCardLink(
                    id: "\(summary.id)-\(participant.person.id)",
                    content: PVPersonCardContent(
                        summary: participant.person,
                        role: L10n.DesignSystem.spokenJoin(names),
                        isFocus: false,
                        locale: locale,
                        orders: orders
                    ),
                    location: PersonsList.location(participant.person, orders: orders).withPageSection(.timeline)
                )
            },
            participantsHidden: event.participants.count - shown.count,
            spokenLabel: spoken
        )
    }
}
