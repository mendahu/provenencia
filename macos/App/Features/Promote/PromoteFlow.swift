import Foundation

/// One page of Promote (S9-44): a row list, not a step machine. Decisions
/// live here. The model loads the proposal and files Done.
struct PromoteFlow: Equatable, Sendable {
    /// Default profile orders for row titles. Empty keeps stored part order.
    var nameOrders = NameFormatOrders()

    enum Target: Equatable, Sendable {
        /// No choice yet. Every row Promote suggests no record for starts here, so Skip and New are explicit.
        case unset
        case handle(id: String, ref: String, title: String)
        case newKind
        case skip

        var token: String {
            switch self {
            case .unset: ""
            case .handle(let id, _, _): Self.handleToken(id)
            case .newKind: "new"
            case .skip: "skip"
            }
        }

        /// The menu option value that stands for the record with this id.
        static func handleToken(_ id: String) -> String { "handle:\(id)" }

        /// The menu option value for "Find another record…": a sentinel that opens a search and is
        /// never a target.
        static let findToken = "find"

        var handleID: String? {
            if case .handle(let id, _, _) = self { return id }
            return nil
        }

        var handleRef: String? {
            if case .handle(_, let ref, _) = self { return ref }
            return nil
        }
    }

    enum Assessment: String, Equatable, Sendable {
        case strong, medium, weak, none

        init(wire: String) { self = Assessment(rawValue: wire) ?? .none }
    }

    /// Why a suggested row moved on the last proposal. The page words it.
    enum Reason: Equatable, Sendable {
        /// Reached from this neighbor row once it was matched.
        case via(neighborID: String)
        /// Settling put this row's pair here because of this neighbor.
        case neighborPlaced(neighborID: String)
        /// Settling moved this row's pair because of this neighbor.
        case neighborMoved(neighborID: String)
        /// Settling released this row's pair because of this neighbor.
        case neighborReleased(neighborID: String)

        /// Why settling released an earlier pair of this row; nil when it did not.
        init?(move proposal: CatalogPromoteGraphAlignmentRow) {
            guard let code = proposal.moveReason, !proposal.moveNeighborSubjectID.isEmpty else { return nil }
            let id = proposal.moveNeighborSubjectID
            switch code {
            case .neighborPlaced: self = .neighborPlaced(neighborID: id)
            case .neighborMoved: self = .neighborMoved(neighborID: id)
            case .neighborReleased: self = .neighborReleased(neighborID: id)
            default: return nil
            }
        }

        /// The neighbor the proposal reached this row from; nil when it did not.
        init?(via proposal: CatalogPromoteGraphAlignmentRow) {
            guard !proposal.viaNeighborSubjectID.isEmpty else { return nil }
            self = .via(neighborID: proposal.viaNeighborSubjectID)
        }
    }

    /// Why a connection's end does not file.
    enum WaitReason: Equatable, Sendable {
        case skipped, unset, notOnPage
    }

    /// Whether a connection files on Done, and if not, why.
    enum ConnectionState: Equatable, Sendable {
        case files
        /// An end does not file: `endID` is the Subject that waits (the connection's first end when both do).
        case waits(endID: String?, why: WaitReason)
        /// Both ends land on one record; `ref` is that record's.
        case selfLink(ref: String)

        /// The Subject the connection waits for; nil when it files or is a self-link.
        var waitingEndID: String? {
            if case .waits(let endID, _) = self { return endID }
            return nil
        }
    }

    /// Why a row may be the same entity as another row on this page.
    enum DuplicateNote: Equatable, Sendable {
        /// Both rows land on one handle (or the other took the one this wanted).
        case sharesHandle(otherName: String, ref: String)
        /// Both rows mint New handles and their values match.
        case alsoNew(otherName: String)
    }

    struct Alternative: Equatable, Sendable, Identifiable {
        var id: String
        var ref: String
        /// The menu's form: a person sorted ("Smith, John"). Events and places read the same everywhere.
        var title: String
        /// Dates, place, or other secondary line under the name.
        var subtitle: String = ""
    }

    struct Row: Equatable, Sendable, Identifiable {
        var subjectID: String
        var kind: EvidencePrimaryKind
        var name: String
        var ref: String
        var anchor: Bool
        var target: Target
        var menu: [Alternative]
        var decided: Bool
        /// A suggested row whose target moved on the last proposal.
        var updated: Bool
        /// Ref of the stronger handle a decided row's neighbors point at.
        var conflictNote: String?
        /// The row the proposal says this one may duplicate.
        var proposedDuplicateOf: String?
        var duplicateNote: DuplicateNote?
        /// What the Updated caption words: why settling moved this row, else the neighbor the proposal reached it from; nil when it names neither.
        var caption: Reason? = nil
        /// An Event's date as the Events list lists it, without the source's words for it; empty for anything else.
        var date: String = ""
        var id: String { subjectID }

        /// The menu item a record token names, the one decoder of `Target.handleToken`.
        func menuItem(forToken token: String) -> Alternative? {
            menu.first { Target.handleToken($0.id) == token }
        }

        /// The ref the page names this row by: its record's when it has one, else its Source ref.
        var shownRef: String { target.handleRef ?? ref }
    }

    @dynamicMemberLookup
    struct Bridge: Equatable, Sendable, Identifiable {
        var fact: BridgeFact
        var state: ConnectionState
        var id: String { fact.id }
        var files: Bool { state == .files }

        subscript<T>(dynamicMember key: KeyPath<BridgeFact, T> & Sendable) -> T { fact[keyPath: key] }

        func name(of endID: String?) -> String { fact.name(of: endID) }
    }

    /// What a row says of a connection's other end when that end is not a row to work on.
    struct OtherEnd: Equatable, Sendable {
        var name: String
        /// The end is already filed (an anchor), not just missing from the page.
        var filed: Bool
    }

    /// A connection as one row shows it.
    struct RowConnection: Equatable, Sendable, Identifiable {
        var bridge: Bridge
        var otherEnd: OtherEnd?
        var id: String { bridge.id }
    }

    struct Anchor: Equatable, Sendable {
        var id: String
        var ref: String
        var title: String
    }

    struct SubjectFact: Equatable, Sendable {
        var id: String
        var kind: EvidencePrimaryKind
        var name: String
        var ref: String
        /// Set when the subject is already filed.
        var anchor: Anchor?
        /// An Event's date as the Events list lists it, without the source's words for it; empty for anything else.
        var date: String = ""
    }

    struct BridgeFact: Equatable, Sendable {
        var id: String
        var sentence: String
        /// The connection's term in lowercase ("spouse"): a row's reason names it; nil when the
        /// connection has none.
        var term: String?
        var endAName: String = ""
        var endBName: String = ""
        var mark: PVMarkKey = .subjectRelationship
        var endA: String?
        var endB: String?
        /// Already an accepted association. Hidden from the will-file list;
        /// Done does not file it again.
        var alreadyFiled: Bool = false

        /// The name of one end, as its row reads; empty when `endID` is neither end.
        func name(of endID: String?) -> String {
            guard let endID else { return "" }
            if endID == endA { return endAName }
            if endID == endB { return endBName }
            return ""
        }
    }

    var entryID: String
    var mappedRest = false
    var rows: [Row] = []
    var bridges: [BridgeFact] = []
    var revision: Int64 = 0
    var saving = false
    var staleNote: String?
    var manual = false
    var pendingLeave: PendingNavigation?

    var visibleRows: [Row] {
        guard mappedRest else { return rows.filter { $0.subjectID == entryID } }
        return rows
    }

    var restCount: Int { rows.filter { $0.subjectID != entryID }.count }

    var fileCount: Int {
        scope.filter { row in
            guard !row.anchor else { return false }
            switch row.target {
            case .handle, .newKind: return true
            case .skip, .unset: return false
            }
        }.count
    }

    var skipCount: Int { scope.filter { !$0.anchor && $0.target == .skip }.count }

    /// Rows with no target yet.
    var unsetCount: Int { scope.filter { !$0.anchor && $0.target == .unset }.count }

    /// Every visible row has a target. An unset row blocks Done.
    var targetsChosen: Bool {
        unsetCount == 0
    }

    var connectionCount: Int { connections.filter(\.files).count }

    private var scope: [Row] { visibleRows }

    private func lined(_ bridge: BridgeFact, _ state: ConnectionState) -> Bridge {
        Bridge(fact: bridge, state: state)
    }

    /// Every connection not yet filed, with whether it files and, if not, which end waits and why.
    var connections: [Bridge] {
        let rowByID = Dictionary(rows.map { ($0.subjectID, $0) }, uniquingKeysWith: { first, _ in first })
        return bridges.filter { !$0.alreadyFiled }.map { bridge in
            let left = resolve(bridge.endA, in: rowByID)
            let right = resolve(bridge.endB, in: rowByID)
            switch (left, right) {
            case (.waits(let why), _):
                return lined(bridge, .waits(endID: bridge.endA, why: why))
            case (_, .waits(let why)):
                return lined(bridge, .waits(endID: bridge.endB, why: why))
            case (.handle(let leftID, let ref), .handle(let rightID, _)):
                return lined(bridge, leftID == rightID ? .selfLink(ref: ref) : .files)
            }
        }
    }

    /// Each connection under every end that is a working row, keyed by the row's subject id and
    /// ordered by state, then sentence. The other end gets a note when it is not a working row.
    var connectionsByRow: [String: [RowConnection]] {
        let working = Set(scope.filter { !$0.anchor }.map(\.subjectID))
        let anchors = Set(rows.filter(\.anchor).map(\.subjectID))
        var byRow: [String: [RowConnection]] = [:]
        for bridge in connections {
            let ends = [bridge.endA, bridge.endB]
            var seen = Set<String>()
            for case let end? in ends where working.contains(end) && seen.insert(end).inserted {
                let far = end == bridge.endA ? bridge.endB : bridge.endA
                var note: OtherEnd?
                if let far, !working.contains(far) {
                    let name = bridge.name(of: far)
                    if !name.isEmpty { note = OtherEnd(name: name, filed: anchors.contains(far)) }
                }
                byRow[end, default: []].append(RowConnection(bridge: bridge, otherEnd: note))
            }
        }
        for key in byRow.keys {
            byRow[key]?.sort(by: Self.shownBefore)
        }
        return byRow
    }

    private static func stateRank(_ state: ConnectionState) -> Int {
        switch state {
        case .files: 0
        case .waits: 1
        case .selfLink: 2
        }
    }

    private static func shownBefore(_ a: RowConnection, _ b: RowConnection) -> Bool {
        let rankA = stateRank(a.bridge.state)
        let rankB = stateRank(b.bridge.state)
        if rankA != rankB { return rankA < rankB }
        switch a.bridge.sentence.localizedStandardCompare(b.bridge.sentence) {
        case .orderedAscending: return true
        case .orderedDescending: return false
        case .orderedSame: return a.id < b.id
        }
    }

    /// The first bridge on the page, in page order, that joins these two Subjects (either way round).
    func bridge(between a: String, and b: String) -> BridgeFact? {
        bridges.first { ($0.endA == a && $0.endB == b) || ($0.endA == b && $0.endB == a) }
    }

    /// The row for this Subject, or nil when it is not on the page.
    func row(_ subjectID: String?) -> Row? {
        guard let subjectID else { return nil }
        return rows.first { $0.subjectID == subjectID }
    }

    private enum EndResolution: Equatable {
        case handle(id: String, ref: String)
        case waits(WaitReason)
    }

    /// The handle a subject will be, or why it stays unfiled. The one place that decides.
    private func resolve(_ subjectID: String?, in rowByID: [String: Row]) -> EndResolution {
        guard let subjectID, let row = rowByID[subjectID] else { return .waits(.notOnPage) }
        if row.anchor {
            guard let id = row.target.handleID, let ref = row.target.handleRef else { return .waits(.notOnPage) }
            return .handle(id: id, ref: ref)
        }
        if !mappedRest && subjectID != entryID { return .waits(.notOnPage) }
        switch row.target {
        case .handle(let id, let ref, _): return .handle(id: id, ref: ref)
        case .newKind: return .handle(id: "new:" + subjectID, ref: row.ref)
        case .skip: return .waits(.skipped)
        case .unset: return .waits(.unset)
        }
    }

    mutating func load(
        entryID: String,
        proposal: CatalogPromoteGraphAlignmentProposal,
        subjects: [SubjectFact],
        bridges: [BridgeFact]
    ) {
        self.entryID = entryID
        self.bridges = bridges
        revision = proposal.revision
        let byID = Dictionary(uniqueKeysWithValues: subjects.map { ($0.id, $0) })
        rows = proposal.rows.compactMap { proposalRow in
            guard let fact = byID[proposalRow.subjectID] else { return nil }
            return Self.makeRow(proposalRow, fact: fact, previous: nil, orders: nameOrders)
        }
        noteDuplicates()
    }

    mutating func mapRest() {
        mappedRest = true
    }

    /// Returns true when suggested rows should be proposed again.
    @discardableResult
    mutating func setTarget(subjectID: String, token: String) -> Bool {
        guard let index = rows.firstIndex(where: { $0.subjectID == subjectID }), !rows[index].anchor else { return false }
        guard let next = target(for: token, on: rows[index]) else { return false }
        return select(next, at: index)
    }

    /// Holds a record Promote did not offer: it joins the row's menu as the
    /// researcher's choice, titled as the menu titles every record, then is
    /// selected like a menu pick. Returns true when suggested rows should be
    /// proposed again.
    @discardableResult
    mutating func hold(subjectID: String, id: String, ref: String, header: CatalogConclusionHeader?) -> Bool {
        guard !id.isEmpty,
              let index = rows.firstIndex(where: { $0.subjectID == subjectID }),
              !rows[index].anchor else { return false }
        let item: Alternative
        if let offered = rows[index].menu.first(where: { $0.id == id }) {
            item = offered
        } else {
            item = Self.item(
                id: id, ref: ref, kind: rows[index].kind,
                person: header?.person, event: header?.event, place: header?.place,
                orders: nameOrders
            )
            rows[index].menu.append(item)
        }
        return select(.handle(id: item.id, ref: item.ref, title: item.title), at: index)
    }

    /// Selects a target on a working row: the old target's caption goes, the
    /// row is decided. Returns true.
    private mutating func select(_ next: Target, at index: Int) -> Bool {
        if next != rows[index].target {
            // The caption worded a move of the old target.
            rows[index].target = next
            rows[index].caption = nil
        }
        rows[index].decided = true
        rows[index].updated = false
        manual = true
        noteDuplicates()
        return true
    }

    /// The targets for these rows as the page holds them, for a proposal: a handle, New, or Skip.
    /// An unset row has none.
    func fixed(for rows: [Row]) -> [CatalogPromoteGraphAlignmentFixed] {
        rows.compactMap { row in
            switch row.target {
            case .handle(let id, _, _):
                return CatalogPromoteGraphAlignmentFixed(subjectID: row.subjectID, handleID: id)
            case .newKind:
                return CatalogPromoteGraphAlignmentFixed(subjectID: row.subjectID, target: .new)
            case .skip:
                return CatalogPromoteGraphAlignmentFixed(subjectID: row.subjectID, target: .skip)
            case .unset:
                return nil
            }
        }
    }

    /// Fold a new proposal. Decided rows keep their choice.
    mutating func merge(_ proposal: CatalogPromoteGraphAlignmentProposal, subjects: [SubjectFact]) {
        revision = proposal.revision
        let byID = Dictionary(uniqueKeysWithValues: subjects.map { ($0.id, $0) })
        let previous = Dictionary(uniqueKeysWithValues: rows.map { ($0.subjectID, $0) })
        rows = proposal.rows.compactMap { proposalRow in
            guard let fact = byID[proposalRow.subjectID] else { return nil }
            return Self.makeRow(proposalRow, fact: fact, previous: previous[proposalRow.subjectID], orders: nameOrders)
        }
        noteDuplicates()
    }

    /// One line per row to file.
    func batchRows() -> [CatalogPromoteBatchRow] {
        filedRows.map(batchRow)
    }

    /// The rows Done files: visible, not already filed, with a target.
    private var filedRows: [Row] {
        scope.filter { !$0.anchor && $0.target != .unset }
    }

    private func batchRow(_ row: Row) -> CatalogPromoteBatchRow {
        let entityID: String?
        let target: String
        switch row.target {
        case .handle(let id, _, _):
            target = "handle"
            entityID = id
        case .newKind:
            target = "new"
            entityID = nil
        case .skip, .unset:
            target = "skip"
            entityID = nil
        }
        return CatalogPromoteBatchRow(
            subjectID: row.subjectID, target: target, entityID: entityID,
            confidenceGradeID: nil, argument: "", defaultPins: false, pinIDs: []
        )
    }

    mutating func requestLeave(_ pending: PendingNavigation) -> Bool {
        guard manual, !saving else { return false }
        pendingLeave = pending
        return true
    }

    mutating func cancelLeave() { pendingLeave = nil }

    /// Rows that would land on one handle, or that the proposal says may be
    /// the same entity, name each other.
    private mutating func noteDuplicates() {
        var owners: [String: [Int]] = [:]
        for index in rows.indices {
            rows[index].duplicateNote = nil
            if let id = rows[index].target.handleID, !rows[index].anchor {
                owners[id, default: []].append(index)
            }
        }
        for (_, indexes) in owners where indexes.count > 1 {
            for index in indexes {
                guard let ref = rows[index].target.handleRef else { continue }
                let other = indexes.first { $0 != index } ?? index
                rows[index].duplicateNote = .sharesHandle(otherName: rows[other].name, ref: ref)
            }
        }
        for index in rows.indices where rows[index].duplicateNote == nil && !rows[index].anchor {
            guard let otherID = rows[index].proposedDuplicateOf,
                  let other = row(otherID)
            else { continue }
            if let ref = other.target.handleRef {
                rows[index].duplicateNote = .sharesHandle(otherName: other.name, ref: ref)
            } else if rows[index].target == .newKind, other.target == .newKind {
                rows[index].duplicateNote = .alsoNew(otherName: other.name)
            }
        }
    }

    /// The target a menu token names on this row; nil for a token that names none (the search
    /// entry, an unknown token, a record not in the menu).
    private func target(for token: String, on row: Row) -> Target? {
        if token == Target.skip.token { return .skip }
        if token == Target.newKind.token { return .newKind }
        if let alt = row.menuItem(forToken: token) {
            return .handle(id: alt.id, ref: alt.ref, title: alt.title)
        }
        return nil
    }

    private static func makeRow(
        _ proposal: CatalogPromoteGraphAlignmentRow,
        fact: SubjectFact,
        previous: Row?,
        orders: NameFormatOrders
    ) -> Row {
        if let anchor = fact.anchor {
            return Row(
                subjectID: fact.id, kind: fact.kind, name: fact.name, ref: fact.ref,
                anchor: true,
                target: .handle(id: anchor.id, ref: anchor.ref, title: anchor.title),
                menu: [],
                decided: false, updated: false, conflictNote: nil, proposedDuplicateOf: nil, duplicateNote: nil,
                date: fact.date
            )
        }
        let menu = menu(for: proposal, kind: fact.kind, orders: orders)
        let drafted = draftTarget(proposal, menu: menu)
        let caption = Reason(move: proposal) ?? Reason(via: proposal)
        if let previous, previous.decided {
            var kept = previous
            kept.conflictNote = proposal.conflictWithFixed ? proposal.alternatives.first?.handleRef : nil
            kept.caption = caption
            kept.proposedDuplicateOf = duplicateOf(proposal)
            return kept
        }
        let updated = previous.map { !$0.anchor && $0.target != drafted } ?? false
        return Row(
            subjectID: fact.id, kind: fact.kind, name: fact.name, ref: fact.ref,
            anchor: false,
            target: drafted, menu: menu,
            decided: false, updated: updated, conflictNote: nil,
            proposedDuplicateOf: duplicateOf(proposal), duplicateNote: nil,
            caption: caption, date: fact.date
        )
    }

    private static func duplicateOf(_ proposal: CatalogPromoteGraphAlignmentRow) -> String? {
        proposal.possibleDuplicate && !proposal.duplicateOfSubjectID.isEmpty ? proposal.duplicateOfSubjectID : nil
    }

    private static func draftTarget(_ proposal: CatalogPromoteGraphAlignmentRow, menu: [Alternative]) -> Target {
        let assessment = Assessment(wire: proposal.assessment)
        // No suggestion: leave the choice open so Skip and New are the researcher's.
        if assessment == .none {
            return .unset
        }
        if proposal.target == .handle, let first = menu.first, first.id == proposal.handleID {
            return .handle(id: first.id, ref: first.ref, title: first.title)
        }
        // Another row on the page took this record. Drafting it again would
        // file two subjects on one handle, so the choice stays open, with the
        // duplicate note naming the other row, whatever the band.
        if proposal.reason == .taken {
            return .unset
        }
        // Below the accept bar (a registry where the weak bar sits under it)
        // the candidate is an alternative, not the row target. Still open on
        // that match.
        if assessment == .weak, let match = menu.first {
            return .handle(id: match.id, ref: match.ref, title: match.title)
        }
        return .unset
    }

    private static func menu(for proposal: CatalogPromoteGraphAlignmentRow, kind: EvidencePrimaryKind, orders: NameFormatOrders) -> [Alternative] {
        var out: [Alternative] = []
        if proposal.target == .handle, !proposal.handleID.isEmpty {
            out.append(item(
                id: proposal.handleID, ref: proposal.handleRef,
                kind: EvidencePrimaryKind(rawValue: proposal.kind),
                person: proposal.person, event: proposal.event, place: proposal.place,
                orders: orders
            ))
        }
        for alt in proposal.alternatives where alt.handleID != proposal.handleID {
            out.append(item(
                id: alt.handleID, ref: alt.handleRef, kind: kind,
                person: alt.person, event: alt.event, place: alt.place,
                orders: orders
            ))
        }
        return out
    }

    /// One menu item, titled sorted and subtitled from the record's header, whichever way the
    /// record reached the menu.
    private static func item(
        id: String,
        ref: String,
        kind: EvidencePrimaryKind?,
        person: CatalogPersonHeader?,
        event: CatalogEventHeader?,
        place: CatalogPlaceHeader?,
        orders: NameFormatOrders
    ) -> Alternative {
        let header: CatalogConclusionHeader?
        switch kind {
        case .person: header = person.map { .person($0) }
        case .event: header = event.map { .event($0) }
        case .place: header = place.map { .place($0) }
        case nil: header = nil
        }
        return Alternative(
            id: id, ref: ref,
            title: header.map { ConclusionHeaderDisplay.title($0, style: .sorted, orders: orders) } ?? ref,
            subtitle: header.map(ConclusionHeaderDisplay.secondary) ?? ""
        )
    }
}
