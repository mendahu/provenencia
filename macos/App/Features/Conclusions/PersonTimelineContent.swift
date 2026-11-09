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
        /// Where a marker with a stub sits, as resolved against the dated rows: nil for a
        /// marker whose row has a span, and for a place whose id is not a dated row.
        var place: CatalogLifeMarker.Place? = nil
    }

    enum Item: Equatable, Identifiable {
        case marker(Marker)
        case row(Row)
        case undatedHeader
        /// A fold of `.dated` or `.undated` rows.
        case more(Fold, count: Int)

        /// A row's and a recorded marker's id is the event id; a placeholder's is
        /// "birth" or "death".
        var id: String { slot.id }

        /// This item before it was built.
        fileprivate var slot: Slot {
            switch self {
            case .marker(let marker): .marker(marker.kind, rowID: marker.row?.id)
            case .row(let row): .event(row.id)
            case .undatedHeader: .undatedHeader
            case .more(let fold, let count): .more(fold, count: count)
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

        /// A row's or a marker's spoken label, the name of its point on the line; nil for the
        /// Undated heading and a Show more.
        var spokenLabel: String? {
            switch self {
            case .marker(let marker): marker.spokenLabel
            case .row(let row): row.spokenLabel
            case .undatedHeader, .more: nil
            }
        }
    }

    /// An item before it is built: the content lays out slots first, so a row folded away is never built.
    /// It owns the one id rule.
    fileprivate enum Slot: Equatable {
        case event(String)
        case marker(Marker.Kind, rowID: String?)
        case undatedHeader
        case more(Fold, count: Int)

        var id: String {
            switch self {
            case .event(let id): id
            case .marker(let kind, let rowID): rowID ?? (kind == .birth ? "birth" : "death")
            case .undatedHeader: "undated"
            case .more(let fold, _): fold == .dated ? "more.dated" : "more.undated"
            }
        }

        /// An event row or a recorded marker: what a fold counts.
        var isEvent: Bool {
            switch self {
            case .event: true
            case .marker(_, let rowID): rowID != nil
            case .undatedHeader, .more: false
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
    /// The first item each fold's next page would reveal: a row's id, or a participant
    /// card's; absent when the fold hides nothing.
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

        // Pass 1: ids and spans only. No row is built here.
        allSpans = timeline.events.compactMap { event in event.span.map { Span(id: event.id, span: $0) } }
        guard !timeline.events.isEmpty else { return }
        let events = Dictionary(timeline.events.map { ($0.id, $0) }, uniquingKeysWith: { first, _ in first })

        // A marker's row is its recorded event when that is one of the events.
        func recordedID(_ source: CatalogLifeMarker) -> String? {
            source.eventId.flatMap { events[$0] == nil ? nil : $0 }
        }
        let birthRowID = recordedID(timeline.birth)
        let deathRowID = recordedID(timeline.death)
        // Whether a recorded marker is a dated row with its own span.
        let birthHasSpan = birthRowID.flatMap { events[$0]?.span } != nil
        let deathHasSpan = deathRowID.flatMap { events[$0]?.span } != nil

        // Pass 2: lay out the slots and fold them.
        // (1) The dated rows in the read's order; a recorded marker with a span
        // is its row, in place.
        var dated: [Slot] = timeline.events.filter { $0.span != nil }.map { event in
            if birthRowID == event.id { return .marker(.birth, rowID: event.id) }
            if deathRowID == event.id { return .marker(.death, rowID: event.id) }
            return .event(event.id)
        }
        // The place a marker with a stub names, when it names a dated row.
        let datedIDs = Set(dated.map(\.id))
        func resolved(_ source: CatalogLifeMarker, hasSpan: Bool) -> CatalogLifeMarker.Place? {
            if hasSpan { return nil }
            switch source.place {
            case .after(let id)?, .before(let id)?: return datedIDs.contains(id) ? source.place : nil
            case nil: return nil
            }
        }
        let birthPlace = resolved(timeline.birth, hasSpan: birthHasSpan)
        let deathPlace = resolved(timeline.death, hasSpan: deathHasSpan)

        // (2), (3) Any other marker goes next to the row it names, after or before it;
        // with none, the birth first and the death last. The birth is placed first, and a
        // death placed after a row steps past a birth placed right there.
        let birthSlot = Slot.marker(.birth, rowID: birthRowID)
        let deathSlot = Slot.marker(.death, rowID: deathRowID)
        if !birthHasSpan {
            let at: Int
            switch birthPlace {
            case .before(let id)?: at = dated.firstIndex { $0.id == id } ?? 0
            case .after(let id)?: at = dated.firstIndex { $0.id == id }.map { $0 + 1 } ?? 0
            case nil: at = 0
            }
            dated.insert(birthSlot, at: at)
        }
        if !deathHasSpan {
            var at = dated.count
            switch deathPlace {
            case .before(let id)?: at = dated.firstIndex { $0.id == id } ?? at
            case .after(let id)?:
                if let index = dated.firstIndex(where: { $0.id == id }) {
                    at = index + 1
                    if at < dated.count, dated[at] == birthSlot, birthPlace != nil { at += 1 }
                }
            case nil: break
            }
            dated.insert(deathSlot, at: at)
        }
        let deathAt = dated.firstIndex(of: deathSlot) ?? dated.count - 1

        // (4) The undated rows, less a recorded birth or death (it is its
        // marker), go immediately before the death marker.
        let undated: [Slot] = timeline.events.compactMap { event in
            event.span == nil && event.id != birthRowID && event.id != deathRowID ? .event(event.id) : nil
        }

        var before = Array(dated[..<deathAt])
        let after = Array(dated[(deathAt + 1)...])

        // Folds: the dated rows before the death marker, a recorded birth's row
        // among them; a placeholder ahead of the hidden run is never folded.
        // Only event rows count toward the limit; a marker or placeholder among the hidden
        // rows stays where it is, after the one "Show N more" for them.
        let counted = before.filter(\.isEvent)
        datedFoldIDs = counted.map(\.id)
        revealed[.dated] = Self.eventFold.firstRevealed(counted, pages: folds.pages(.dated))?.id
        let datedFold = Self.eventFold.shown(counted, pages: folds.pages(.dated))
        if datedFold.hidden > 0 {
            let hiddenIDs = Set(counted.dropFirst(datedFold.shown.count).map(\.id))
            var placed = false
            before = before.compactMap { slot in
                guard slot.isEvent, hiddenIDs.contains(slot.id) else { return slot }
                if placed { return nil }
                placed = true
                return .more(.dated, count: datedFold.hidden)
            }
        }

        revealed[.undated] = Self.eventFold.firstRevealed(undated, pages: folds.pages(.undated))?.id
        var undatedSlots: [Slot] = []
        if !undated.isEmpty {
            undatedSlots.append(.undatedHeader)
            let fold = Self.eventFold.shown(undated, pages: folds.pages(.undated))
            undatedSlots += fold.shown
            if fold.hidden > 0 { undatedSlots.append(.more(.undated, count: fold.hidden)) }
        }

        // Pass 3: build the rows and cards of the slots that are left.
        func built(_ id: String) -> Row? {
            guard let event = events[id] else { return nil }
            revealed[.participants(eventId: id)] =
                participantFold.firstRevealed(event.participants, pages: folds.pages(.participants(eventId: id)))
                .map { Self.participantLinkID(eventId: id, personId: $0.person.id) }
            return Self.row(event, folds: folds, fold: participantFold, locale: locale, orders: orders)
        }
        items = (before + undatedSlots + [dated[deathAt]] + after).compactMap { slot in
            switch slot {
            case .event(let id):
                return built(id).map { .row($0) }
            case .marker(let kind, let rowID):
                let row = rowID.flatMap(built)
                return .marker(Marker(
                    kind: kind,
                    row: row,
                    spokenLabel: row?.spokenLabel
                        ?? L10n.string(kind == .birth ? L10n.Conclusions.timelineBirthUnknown : L10n.Conclusions.timelineDeathUnknown),
                    place: kind == .birth ? birthPlace : deathPlace
                ))
            case .undatedHeader: return .undatedHeader
            case .more(let fold, let count): return .more(fold, count: count)
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
