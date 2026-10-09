import Foundation

/// One page of Promote (S9-44): a row list, not a step machine. Decisions
/// live here. The model loads the proposal and files Done.
struct PromoteFlow: Equatable, Sendable {
    enum Target: Equatable, Sendable {
        /// No choice yet. A no-match row starts here so Skip and New are explicit.
        case unset
        case handle(id: String, ref: String, title: String)
        case newKind
        case skip

        var token: String {
            switch self {
            case .unset: ""
            case .handle(let id, _, _): "handle:\(id)"
            case .newKind: "new"
            case .skip: "skip"
            }
        }

        var handleID: String? {
            if case .handle(let id, _, _) = self { return id }
            return nil
        }
    }

    enum Assessment: String, Equatable, Sendable {
        case strong, medium, weak, none

        init(wire: String) { self = Assessment(rawValue: wire) ?? .none }
    }

    /// Why a row reads as it does. The page words it.
    enum Reason: Equatable, Sendable {
        /// Reached from this neighbor row once it was matched.
        case via(neighborID: String)
        /// Filed before this page opened.
        case filed
        /// The researcher's choice, held as given.
        case decided
        /// A property-only match; this Property agrees most.
        case agrees(key: String, origin: String)
        case weak
        /// Another row on the page took this match.
        case taken
        case noMatch
        case empty

        init(code: String, propertyKey: String, propertyOrigin: String, viaNeighborID: String) {
            switch code {
            case "via" where !viaNeighborID.isEmpty: self = .via(neighborID: viaNeighborID)
            case "decided": self = .decided
            case "agrees": self = .agrees(key: propertyKey, origin: propertyOrigin)
            case "weak": self = .weak
            case "taken": self = .taken
            case "no_match": self = .noMatch
            default: self = .empty
            }
        }

        init(_ row: CatalogPromoteGraphAlignmentRow) {
            self.init(
                code: row.reason,
                propertyKey: row.reasonPropertyKey,
                propertyOrigin: row.reasonPropertyOrigin,
                viaNeighborID: row.viaNeighborSubjectID
            )
        }

        init(_ alt: CatalogPromoteGraphAlignmentAlternative) {
            self.init(
                code: alt.reason,
                propertyKey: alt.reasonPropertyKey,
                propertyOrigin: alt.reasonPropertyOrigin,
                viaNeighborID: alt.viaNeighborSubjectID
            )
        }
    }

    enum Grouping: String, Equatable, Sendable {
        case kind, assessment
    }

    /// Whether a connection files on Done, and if not, why.
    enum ConnectionState: Equatable, Sendable {
        case files
        /// The researcher switched it off.
        case off
        /// An end is skipped or not on the page.
        case endSkipped
        /// Both ends land on one handle.
        case selfLink
    }

    /// Why a row may be the same entity as another row on this page.
    enum DuplicateNote: Equatable, Sendable {
        /// Both rows land on one handle (or the other took the one this wanted).
        case sharesHandle(otherName: String, ref: String)
        /// Both rows mint New handles and their values match.
        case alsoNew(otherName: String)
    }

    /// The assessment column for one selected record.
    struct SelectionReadout: Equatable, Sendable {
        var assessment: Assessment
        var reason: Reason
    }

    struct Alternative: Equatable, Sendable, Identifiable {
        var id: String
        var ref: String
        var title: String
        /// Dates, place, or other secondary line under the name.
        var subtitle: String = ""
        var assessment: Assessment = .none
        var reason: Reason = .empty
    }

    struct Row: Equatable, Sendable, Identifiable {
        var subjectID: String
        var kind: EvidencePrimaryKind
        var name: String
        var ref: String
        var anchor: Bool
        var assessment: Assessment
        var reason: Reason
        var target: Target
        var menu: [Alternative]
        var comparisons: [CatalogPromoteGraphAlignmentComparison]
        var pins: Set<String>
        var confidenceGradeID: String?
        var argument: String
        /// The researcher typed the argument; a retarget keeps it instead of re-drafting.
        var argumentEdited = false
        var decided: Bool
        /// A suggested row whose target moved on the last proposal.
        var updated: Bool
        /// Ref of the stronger handle a decided row's neighbors point at.
        var conflictNote: String?
        /// The row the proposal says this one may duplicate.
        var proposedDuplicateOf: String?
        var duplicateNote: DuplicateNote?
        var id: String { subjectID }

        /// The selected record's band and why. Skip, New, and an empty menu show nothing.
        var selectionReadout: SelectionReadout? {
            guard case .handle(let id, _, _) = target,
                  let alt = menu.first(where: { $0.id == id }) else { return nil }
            return SelectionReadout(assessment: alt.assessment, reason: alt.reason)
        }
    }

    struct Bridge: Equatable, Sendable, Identifiable {
        var id: String
        var sentence: String
        var phrase: String
        var endAName: String
        var endBName: String
        var mark: PVMarkKey
        var endA: String?
        var endB: String?
        var state: ConnectionState
        var files: Bool { state == .files }
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
    }

    struct BridgeFact: Equatable, Sendable {
        var id: String
        var sentence: String
        var phrase: String = ""
        var endAName: String = ""
        var endBName: String = ""
        var mark: PVMarkKey = .subjectRelationship
        var endA: String?
        var endB: String?
        /// Already an accepted association. Hidden from the will-file list;
        /// Done does not file it again.
        var alreadyFiled: Bool = false
        /// Switched off on an earlier Done. Starts off; Done keeps it unfiled
        /// unless the researcher switches it back on.
        var declined: Bool = false
    }

    var entryID: String
    var mappedRest = false
    var rows: [Row] = []
    var bridges: [BridgeFact] = []
    var bridgeOff: Set<String> = []
    /// The declined bridges the page opened with.
    var bridgeOffAtLoad: Set<String> = []
    var revision: Int64 = 0
    var saving = false
    var staleNote: String?
    var manual = false
    var pendingLeave: PendingNavigation?
    var sheetSubjectID: String?
    var group = Grouping.kind
    var bridgesOpen = false

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

    /// Every visible row has a target. An unselected no-match menu blocks Done.
    var targetsChosen: Bool {
        scope.allSatisfy { $0.anchor || $0.target != .unset }
    }

    var connectionCount: Int { filedBridges.filter(\.files).count }

    /// Connections switched on or off since the page opened.
    var connectionChanges: Int { bridgeOff.symmetricDifference(bridgeOffAtLoad).count }

    private var scope: [Row] { visibleRows }

    func connectionLines() -> [Bridge] { filedBridges }

    private func lined(_ bridge: BridgeFact, _ state: ConnectionState) -> Bridge {
        Bridge(
            id: bridge.id, sentence: bridge.sentence, phrase: bridge.phrase,
            endAName: bridge.endAName, endBName: bridge.endBName, mark: bridge.mark,
            endA: bridge.endA, endB: bridge.endB, state: state
        )
    }

    private var filedBridges: [Bridge] {
        bridges.filter { !$0.alreadyFiled }.map { bridge in
            if bridgeOff.contains(bridge.id) {
                return lined(bridge, .off)
            }
            guard let left = resolve(bridge.endA), let right = resolve(bridge.endB) else {
                return lined(bridge, .endSkipped)
            }
            if left == right {
                return lined(bridge, .selfLink)
            }
            return lined(bridge, .files)
        }
    }

    /// The handle a subject will be, or nil when it stays unfiled.
    private func resolve(_ subjectID: String?) -> String? {
        guard let subjectID, let row = rows.first(where: { $0.subjectID == subjectID }) else { return nil }
        if row.anchor { return row.target.handleID }
        if !mappedRest && subjectID != entryID { return nil }
        switch row.target {
        case .handle(let id, _, _): return id
        case .newKind: return "new:" + subjectID
        case .skip, .unset: return nil
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
        bridgeOff = Set(bridges.filter { $0.declined && !$0.alreadyFiled }.map(\.id))
        bridgeOffAtLoad = bridgeOff
        revision = proposal.revision
        let byID = Dictionary(uniqueKeysWithValues: subjects.map { ($0.id, $0) })
        rows = proposal.rows.compactMap { proposalRow in
            guard let fact = byID[proposalRow.subjectID] else { return nil }
            return Self.makeRow(proposalRow, fact: fact, previous: nil)
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
        let next = target(for: token, on: rows[index])
        if next != rows[index].target {
            // The lines and pins compared the old target; the next proposal
            // drafts them for this one.
            rows[index].target = next
            rows[index].comparisons = []
            rows[index].pins = []
            if !rows[index].argumentEdited { rows[index].argument = "" }
        }
        rows[index].decided = true
        rows[index].updated = false
        manual = true
        noteDuplicates()
        return true
    }

    mutating func togglePin(subjectID: String, comparisonID: String) {
        guard let index = rows.firstIndex(where: { $0.subjectID == subjectID }) else { return }
        if rows[index].pins.contains(comparisonID) {
            rows[index].pins.remove(comparisonID)
        } else {
            rows[index].pins.insert(comparisonID)
        }
        rows[index].decided = true
        manual = true
    }

    mutating func setArgument(subjectID: String, argument: String) {
        guard let index = rows.firstIndex(where: { $0.subjectID == subjectID }) else { return }
        rows[index].argument = argument
        rows[index].argumentEdited = true
        rows[index].decided = true
        manual = true
    }

    mutating func setConfidence(subjectID: String, gradeID: String?) {
        guard let index = rows.firstIndex(where: { $0.subjectID == subjectID }) else { return }
        rows[index].confidenceGradeID = gradeID
        rows[index].decided = true
        manual = true
    }

    mutating func toggleBridge(_ id: String) {
        if bridgeOff.contains(id) { bridgeOff.remove(id) } else { bridgeOff.insert(id) }
        manual = true
    }

    /// Fold a new proposal. Decided rows keep their choice.
    mutating func merge(_ proposal: CatalogPromoteGraphAlignmentProposal, subjects: [SubjectFact]) {
        revision = proposal.revision
        let byID = Dictionary(uniqueKeysWithValues: subjects.map { ($0.id, $0) })
        let previous = Dictionary(uniqueKeysWithValues: rows.map { ($0.subjectID, $0) })
        rows = proposal.rows.compactMap { proposalRow in
            guard let fact = byID[proposalRow.subjectID] else { return nil }
            return Self.makeRow(proposalRow, fact: fact, previous: previous[proposalRow.subjectID])
        }
        noteDuplicates()
    }

    /// One line per row to file. A Skip writes nothing, so it carries no
    /// argument, confidence, or pins, whatever the row was drafted with.
    func batchRows() -> [CatalogPromoteBatchRow] {
        scope.compactMap { row in
            if row.anchor || row.target == .unset { return nil }
            return batchRow(row)
        }
    }

    private func batchRow(_ row: Row) -> CatalogPromoteBatchRow {
        let pairs: [CatalogPromoteGraphAlignmentPair]
        if case .handle = row.target {
            pairs = row.comparisons.compactMap { comparison in
                guard row.pins.contains(comparison.id),
                      !comparison.incomingObservationID.isEmpty,
                      !comparison.memberObservationID.isEmpty
                else { return nil }
                return CatalogPromoteGraphAlignmentPair(
                    incomingObservationID: comparison.incomingObservationID,
                    memberObservationID: comparison.memberObservationID
                )
            }
        } else {
            pairs = []
        }
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
        let files = target != "skip"
        return CatalogPromoteBatchRow(
            subjectID: row.subjectID,
            target: target,
            entityID: entityID,
            confidenceGradeID: files ? row.confidenceGradeID : nil,
            argument: files ? row.argument : "",
            pairs: pairs
        )
    }

    func skipBridgeIDs() -> [String] { bridgeOff.sorted() }

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
                guard case .handle(_, let ref, _) = rows[index].target else { continue }
                let other = indexes.first { $0 != index } ?? index
                rows[index].duplicateNote = .sharesHandle(otherName: rows[other].name, ref: ref)
            }
        }
        for index in rows.indices where rows[index].duplicateNote == nil && !rows[index].anchor {
            guard let otherID = rows[index].proposedDuplicateOf,
                  let other = rows.first(where: { $0.subjectID == otherID })
            else { continue }
            if case .handle(_, let ref, _) = other.target {
                rows[index].duplicateNote = .sharesHandle(otherName: other.name, ref: ref)
            } else if rows[index].target == .newKind, other.target == .newKind {
                rows[index].duplicateNote = .alsoNew(otherName: other.name)
            }
        }
    }

    private func target(for token: String, on row: Row) -> Target {
        if token == "skip" { return .skip }
        if token == "new" { return .newKind }
        if token.hasPrefix("handle:") {
            let id = String(token.dropFirst("handle:".count))
            if let alt = row.menu.first(where: { $0.id == id }) {
                return .handle(id: alt.id, ref: alt.ref, title: alt.title)
            }
        }
        return row.target
    }

    private static func makeRow(
        _ proposal: CatalogPromoteGraphAlignmentRow,
        fact: SubjectFact,
        previous: Row?
    ) -> Row {
        if let anchor = fact.anchor {
            return Row(
                subjectID: fact.id, kind: fact.kind, name: fact.name, ref: fact.ref,
                anchor: true, assessment: .strong, reason: .filed,
                target: .handle(id: anchor.id, ref: anchor.ref, title: anchor.title),
                menu: [], comparisons: [], pins: [], confidenceGradeID: nil, argument: "",
                decided: false, updated: false, conflictNote: nil, proposedDuplicateOf: nil, duplicateNote: nil
            )
        }
        let menu = menu(for: proposal, kind: fact.kind)
        let drafted = draftTarget(proposal, menu: menu)
        if let previous, previous.decided {
            var kept = previous
            kept.assessment = Assessment(wire: proposal.assessment)
            kept.reason = Reason(proposal)
            // The column shows the selected record. A held handle is rescored,
            // so refresh that record's band. "Your choice" stays off the column;
            // the match sentence already on the menu item remains.
            if case .handle(let id, _, _) = kept.target,
               proposal.target == "handle", proposal.handleID == id,
               let menuIndex = kept.menu.firstIndex(where: { $0.id == id }) {
                kept.menu[menuIndex].assessment = Assessment(wire: proposal.assessment)
                let matchReason = Reason(proposal)
                if matchReason != .decided {
                    kept.menu[menuIndex].reason = matchReason
                }
            }
            // Lines only describe the row's own target: a decided handle is
            // sent as fixed, so the proposal compares exactly that handle.
            let lines = proposal.target == "handle" && proposal.handleID == previous.target.handleID
                ? proposal.comparisons : []
            if previous.comparisons.isEmpty {
                kept.pins = draftedPins(lines)
            } else {
                kept.pins = previous.pins.intersection(lines.map(\.id))
            }
            kept.comparisons = lines
            if !previous.argumentEdited {
                kept.argument = draftedArgument(lines)
            }
            kept.conflictNote = proposal.conflictWithFixed ? proposal.alternatives.first?.handleRef : nil
            kept.proposedDuplicateOf = duplicateOf(proposal)
            return kept
        }
        let lines = drafted.handleID == nil ? [] : proposal.comparisons
        let updated = previous.map { !$0.anchor && $0.target != drafted } ?? false
        return Row(
            subjectID: fact.id, kind: fact.kind, name: fact.name, ref: fact.ref,
            anchor: false, assessment: Assessment(wire: proposal.assessment),
            reason: Reason(proposal),
            target: drafted, menu: menu, comparisons: proposal.comparisons,
            pins: draftedPins(lines), confidenceGradeID: nil, argument: draftedArgument(lines),
            decided: false, updated: updated, conflictNote: nil,
            proposedDuplicateOf: duplicateOf(proposal), duplicateNote: nil
        )
    }

    private static func duplicateOf(_ proposal: CatalogPromoteGraphAlignmentRow) -> String? {
        proposal.possibleDuplicate && !proposal.duplicateOfSubjectID.isEmpty ? proposal.duplicateOfSubjectID : nil
    }

    private static func draftedPins(_ lines: [CatalogPromoteGraphAlignmentComparison]) -> Set<String> {
        Set(lines.filter { $0.pinned && $0.outcome != "conflict" }.map(\.id))
    }

    /// The agreeing values, as a starting argument the researcher can edit.
    private static func draftedArgument(_ lines: [CatalogPromoteGraphAlignmentComparison]) -> String {
        lines
            .filter { $0.outcome == "agree" && !$0.incomingDisplay.isEmpty }
            .map(\.incomingDisplay)
            .joined(separator: "; ")
    }

    private static func draftTarget(_ proposal: CatalogPromoteGraphAlignmentRow, menu: [Alternative]) -> Target {
        let assessment = Assessment(wire: proposal.assessment)
        // No match: leave the menu empty so Skip and New are a choice.
        if assessment == .none {
            return .unset
        }
        if proposal.target == "handle", let first = menu.first, first.id == proposal.handleID {
            return .handle(id: first.id, ref: first.ref, title: first.title)
        }
        // Another row on the page took this record. Drafting it again would
        // file two subjects on one handle, so the choice stays open, with the
        // duplicate note naming the other row, whatever the band.
        if Reason(proposal) == .taken {
            return .unset
        }
        // Below the accept bar (a registry where the weak bar sits under it)
        // the candidate is an alternative, not the row target. Still open on
        // that match.
        if assessment == .weak, let match = menu.first {
            return .handle(id: match.id, ref: match.ref, title: match.title)
        }
        return .skip
    }

    private static func menu(for proposal: CatalogPromoteGraphAlignmentRow, kind: EvidencePrimaryKind) -> [Alternative] {
        var out: [Alternative] = []
        if proposal.target == "handle", !proposal.handleID.isEmpty {
            out.append(Alternative(
                id: proposal.handleID, ref: proposal.handleRef,
                title: headerTitle(proposal), subtitle: headerSubtitle(proposal),
                assessment: Assessment(wire: proposal.assessment), reason: Reason(proposal)
            ))
        }
        for alt in proposal.alternatives where alt.handleID != proposal.handleID {
            out.append(Alternative(
                id: alt.handleID, ref: alt.handleRef,
                title: headerTitle(alt, kind: kind), subtitle: headerSubtitle(alt, kind: kind),
                assessment: Assessment(wire: alt.assessment), reason: Reason(alt)
            ))
        }
        return out
    }

    private static func headerSubtitle(_ row: CatalogPromoteGraphAlignmentRow) -> String {
        conclusionSubtitle(
            kind: EvidencePrimaryKind(rawValue: row.kind),
            person: row.person, event: row.event, place: row.place
        )
    }

    private static func headerSubtitle(_ alt: CatalogPromoteGraphAlignmentAlternative, kind: EvidencePrimaryKind) -> String {
        conclusionSubtitle(kind: kind, person: alt.person, event: alt.event, place: alt.place)
    }

    private static func headerTitle(_ row: CatalogPromoteGraphAlignmentRow) -> String {
        conclusionTitle(
            kind: EvidencePrimaryKind(rawValue: row.kind),
            person: row.person, event: row.event, place: row.place,
            fallback: row.handleRef
        )
    }

    private static func headerTitle(_ alt: CatalogPromoteGraphAlignmentAlternative, kind: EvidencePrimaryKind) -> String {
        conclusionTitle(
            kind: kind,
            person: alt.person, event: alt.event, place: alt.place,
            fallback: alt.handleRef
        )
    }

    private static func conclusionTitle(
        kind: EvidencePrimaryKind?,
        person: CatalogPersonHeader?,
        event: CatalogEventHeader?,
        place: CatalogPlaceHeader?,
        fallback: String
    ) -> String {
        switch kind {
        case .person:
            if let person { return PersonHeaderDisplay.title(person) }
        case .event:
            if let event { return EventTitleDisplay.title(event) }
        case .place:
            if let place { return PlaceTitleDisplay.title(place) }
        case nil:
            break
        }
        return fallback
    }

    private static func conclusionSubtitle(
        kind: EvidencePrimaryKind?,
        person: CatalogPersonHeader?,
        event: CatalogEventHeader?,
        place: CatalogPlaceHeader?
    ) -> String {
        switch kind {
        case .person:
            if let person { return PersonLifeDisplay.line(person).text }
        case .event:
            if let event { return EventSecondaryDisplay.line(event) }
        case .place:
            if let place {
                return PlaceChainDisplay.line(parents: place.parents, candidates: place.parentsAreCandidates)
            }
        case nil:
            break
        }
        return ""
    }
}
