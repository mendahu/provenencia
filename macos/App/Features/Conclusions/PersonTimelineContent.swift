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
    static let eventFold = ConclusionFold(limit: 50, minimumHidden: 1)

    /// Other people shown on one event before "Show N more": the board's five
    /// at Regular and twelve at Wide, so a fold is one or two rows of cards.
    static func participantFold(_ widthClass: PVWidthClass) -> ConclusionFold {
        switch widthClass {
        case .regular: ConclusionFold(limit: 5, minimumHidden: 1)
        case .wide: ConclusionFold(limit: 12, minimumHidden: 1)
        }
    }

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
        /// The person's age on the event, when Go could tell it.
        var age: CatalogAgeRange?
        /// The other people shown, in the read's order (most important role first).
        var participants: [PersonCardLink]
        var participantsHidden: Int
        /// "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject, age 0".
        var spokenLabel: String

        /// The age flag's text: "27", "17–18", "≤ 63", "≥ 32", or "?" with no age.
        var ageText: String { PersonTimelineContent.ageWords(age).flag }

        /// The flag's tooltip.
        var ageHelp: String { PersonTimelineContent.ageWords(age).help }
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
        /// The row's spoken label, the name of its point on the line.
        var spokenLabel: String
    }

    private(set) var items: [Item] = []
    private(set) var allSpans: [Span] = []
    /// What each point or cap on the lifeline is called, by item id: a row's or
    /// marker's spoken label, including the rows a fold hides.
    private(set) var pointLabels: [String: String] = [:]
    /// The first item each fold would reveal, folded or not: a row's id, or a
    /// participant card's.
    private var revealed: [Fold: String] = [:]
    /// The ids of the event rows the dated fold counts, in order: the dated rows before the
    /// death marker, a recorded birth's row among them.
    private var datedFoldIDs: [String] = []
    let isEmpty: Bool

    /// The id of the first item that unfolding `fold` adds, for the focus to land
    /// on; nil when the list has nothing to fold.
    func firstRevealed(by fold: Fold) -> String? { revealed[fold] }

    /// How many pages of the dated fold show the row `eventID`, for a click on its point on
    /// the lifeline; nil for an id the dated fold never hides (an undated row, a placeholder,
    /// a row after the death marker).
    func pages(toShow eventID: String) -> Int? {
        datedFoldIDs.firstIndex(of: eventID).map {
            Self.eventFold.pages(toShow: $0, count: datedFoldIDs.count)
        }
    }

    /// Whether `id` (a row's event id or a participant card's id) is drawn.
    func isShown(_ id: String) -> Bool {
        items.compactMap(\.eventRow).contains { $0.id == id || $0.participants.contains { $0.id == id } }
    }

    /// The event id of the row that shows the card `id`; nil for a row's title or an unknown id.
    func row(holdingCard id: String) -> String? {
        items.compactMap(\.eventRow).first { $0.participants.contains { $0.id == id } }?.id
    }

    init(
        timeline: CatalogPersonTimeline,
        folds: ConclusionFoldState<Fold> = ConclusionFoldState(),
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders(),
        widthClass: PVWidthClass = .regular
    ) {
        isEmpty = timeline.events.isEmpty
        let participantFold = Self.participantFold(widthClass)
        let rows = timeline.events.map {
            Self.row($0, folds: folds, fold: participantFold, locale: locale, orders: orders)
        }
        allSpans = rows.compactMap { row in row.span.map { Span(id: row.id, span: $0, spokenLabel: row.spokenLabel) } }
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
        for event in timeline.events {
            revealed[.participants(eventId: event.event.id)] =
                participantFold.firstRevealed(
                    event.participants, pages: folds.pages(.participants(eventId: event.event.id))
                ).map {
                    Self.participantLinkID(eventId: event.event.id, personId: $0.person.id)
                }
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
        // among them; a placeholder ahead of the hidden run is never folded.
        // Only event rows count toward the limit; a marker or placeholder among the hidden
        // rows stays where it is, after the one "Show N more" for them.
        let counted = before.filter { $0.eventRow != nil }
        datedFoldIDs = counted.map(\.id)
        revealed[.dated] = Self.eventFold.firstRevealed(counted, pages: folds.pages(.dated))?.id
        let datedFold = Self.eventFold.shown(counted, pages: folds.pages(.dated))
        if datedFold.hidden > 0 {
            let hiddenIDs = Set(counted.dropFirst(datedFold.shown.count).map(\.id))
            var placed = false
            before = before.compactMap { item in
                guard item.eventRow != nil, hiddenIDs.contains(item.id) else { return item }
                if placed { return nil }
                placed = true
                return .more(.dated, count: datedFold.hidden)
            }
        }

        revealed[.undated] = Self.eventFold.firstRevealed(undated, pages: folds.pages(.undated))?.id
        var undatedItems: [Item] = []
        if !undated.isEmpty {
            undatedItems.append(.undatedHeader)
            let fold = Self.eventFold.shown(undated, pages: folds.pages(.undated))
            undatedItems += fold.shown.map { .row($0) }
            if fold.hidden > 0 { undatedItems.append(.more(.undated, count: fold.hidden)) }
        }

        items = before + undatedItems + [dated[deathAt]] + after

        pointLabels = Dictionary(allSpans.map { ($0.id, $0.spokenLabel) }, uniquingKeysWith: { first, _ in first })
        for item in items {
            if let row = item.eventRow {
                pointLabels[item.id] = row.spokenLabel
            } else if case .marker(let marker) = item {
                pointLabels[item.id] = marker.spokenLabel
            }
        }
    }

    /// The names of the role terms among `values`; one that is not a term, or has no name, is skipped.
    private static func roleNames(_ values: [CatalogConclusionValue]) -> [String] {
        values.compactMap { PropertyTermDisplay.name($0, propertyKey: SeededPropertyKey.role) }
    }

    /// What a row says of the person's age: the flag ("27", "17–18", "≤ 63",
    /// "≥ 32", or "?" for an age Go could not tell), its tooltip, and the spoken
    /// tail, nil when the age is unknown. Only one side of an open age is known.
    static func ageWords(_ age: CatalogAgeRange?) -> (flag: String, help: String, spoken: String?) {
        guard let age else { return ("?", L10n.string(L10n.Conclusions.timelineAgeUnknown), nil) }
        if age.openLo {
            return (L10n.Conclusions.timelineAgeAtMost(age.hi), L10n.Conclusions.timelineAgeAtMostHelp(age.hi), L10n.Conclusions.timelineAgeAtMostSpoken(age.hi))
        }
        if age.openHi {
            return (L10n.Conclusions.timelineAgeAtLeast(age.lo), L10n.Conclusions.timelineAgeAtLeastHelp(age.lo), L10n.Conclusions.timelineAgeAtLeastSpoken(age.lo))
        }
        if age.lo == age.hi {
            let text = String(age.hi)
            return (text, L10n.Conclusions.timelineAgeHelp(text), L10n.Conclusions.timelineAgeSpoken(text))
        }
        let text = L10n.Conclusions.timelineAgeRange(lo: age.lo, hi: age.hi)
        return (text, L10n.Conclusions.timelineAgeHelp(text), L10n.Conclusions.timelineAgeSpokenRange(lo: age.lo, hi: age.hi))
    }

    private static func participantLinkID(eventId: String, personId: String) -> String { "\(eventId)-\(personId)" }

    private static func row(
        _ event: CatalogPersonEvent,
        folds: ConclusionFoldState<Fold>,
        fold participantFold: ConclusionFold,
        locale: Locale,
        orders: NameFormatOrders
    ) -> Row {
        let summary = event.event
        let roles = event.roles.compactMap { value -> RoleBadge? in
            roleNames([value]).first.map { RoleBadge(text: $0, isSubject: value.productKey == SeededRoleTerm.subject) }
        }
        var spoken = EventsList.spokenLine(summary, orders: orders, locale: locale)
        if let joined = L10n.DesignSystem.spokenJoin(roles.map(\.text)) {
            spoken = L10n.DesignSystem.spokenJoin(spoken, L10n.Conclusions.timelineAsRole(joined))
        }
        if let tail = ageWords(event.age).spoken {
            spoken = L10n.DesignSystem.spokenJoin(spoken, tail)
        }
        let shown = participantFold.shown(
            event.participants, pages: folds.pages(.participants(eventId: summary.id))
        ).shown
        return Row(
            id: summary.id,
            title: EventTitleDisplay.titleSource(summary, locale: locale, orders: orders),
            line: EventSecondaryDisplay.content(summary, locale: locale),
            imageRelPath: summary.pictureImageRelPath,
            roles: roles,
            location: EventsList.location(summary, orders: orders),
            span: event.span,
            age: event.age,
            participants: shown.map { participant in
                PersonCardLink.link(
                    id: participantLinkID(eventId: summary.id, personId: participant.person.id),
                    summary: participant.person,
                    role: L10n.DesignSystem.spokenJoin(roleNames(participant.roles)),
                    opensOn: .timeline,
                    locale: locale,
                    orders: orders
                )
            },
            participantsHidden: event.participants.count - shown.count,
            spokenLabel: spoken
        )
    }
}
