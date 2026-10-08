import Foundation

/// One page of Promote (S9-44): a row list, not a step machine. Decisions
/// live here. The model loads the proposal and files Done.
struct PromoteFlow: Equatable, Sendable {
    enum Target: Equatable, Sendable {
        case handle(id: String, ref: String, title: String)
        case newKind
        case skip

        var token: String {
            switch self {
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

    struct Alternative: Equatable, Sendable, Identifiable {
        var id: String
        var ref: String
        var title: String
    }

    struct Row: Equatable, Sendable, Identifiable {
        var subjectID: String
        var kind: EvidencePrimaryKind
        var name: String
        var ref: String
        var anchor: Bool
        var assessment: String
        var reason: String
        var target: Target
        var menu: [Alternative]
        var comparisons: [CatalogPromoteGraphAlignmentComparison]
        var pins: Set<String>
        var confidenceGradeID: String?
        var argument: String
        var decided: Bool
        var updatedNote: String?
        var conflictNote: String?
        var duplicateNote: String?
        var id: String { subjectID }
    }

    struct Bridge: Equatable, Sendable, Identifiable {
        var id: String
        var sentence: String
        var endA: String?
        var endB: String?
        var files: Bool
        var why: String
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
        var endA: String?
        var endB: String?
    }

    var entryID: String
    var mappedRest = false
    var rows: [Row] = []
    var bridges: [BridgeFact] = []
    var bridgeOff: Set<String> = []
    var revision: Int64 = 0
    var saving = false
    var staleNote: String?
    var manual = false
    var pendingLeave: PendingNavigation?
    var sheetSubjectID: String?
    /// Kind or assessment.
    var group = "kind"
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
            case .skip: return false
            }
        }.count
    }

    var skipCount: Int { scope.filter { !$0.anchor && $0.target.token == "skip" }.count }

    var connectionCount: Int { filedBridges.filter(\.files).count }

    private var scope: [Row] { visibleRows }

    func connectionLines() -> [Bridge] { filedBridges }

    private var filedBridges: [Bridge] {
        bridges.map { bridge in
            if bridgeOff.contains(bridge.id) {
                return Bridge(id: bridge.id, sentence: bridge.sentence, endA: bridge.endA, endB: bridge.endB, files: false, why: "off")
            }
            guard let left = resolve(bridge.endA), let right = resolve(bridge.endB) else {
                return Bridge(id: bridge.id, sentence: bridge.sentence, endA: bridge.endA, endB: bridge.endB, files: false, why: "skipped")
            }
            if left == right {
                return Bridge(id: bridge.id, sentence: bridge.sentence, endA: bridge.endA, endB: bridge.endB, files: false, why: "self")
            }
            return Bridge(id: bridge.id, sentence: bridge.sentence, endA: bridge.endA, endB: bridge.endB, files: true, why: "")
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
        case .skip: return nil
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
        rows[index].target = target(for: token, on: rows[index])
        rows[index].decided = true
        rows[index].updatedNote = nil
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

    func batchRows() -> [CatalogPromoteBatchRow] {
        scope.filter { !$0.anchor }.map { row in
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
            case .skip:
                target = "skip"
                entityID = nil
            }
            return CatalogPromoteBatchRow(
                subjectID: row.subjectID,
                target: target,
                entityID: entityID,
                confidenceGradeID: row.confidenceGradeID,
                argument: row.argument,
                pairs: pairs
            )
        }
    }

    func skipBridgeIDs() -> [String] { bridgeOff.sorted() }

    mutating func requestLeave(_ pending: PendingNavigation) -> Bool {
        guard manual, !saving else { return false }
        pendingLeave = pending
        return true
    }

    mutating func cancelLeave() { pendingLeave = nil }

    private mutating func noteDuplicates() {
        var owners: [String: [Int]] = [:]
        for index in rows.indices {
            rows[index].duplicateNote = nil
            if let id = rows[index].target.handleID, !rows[index].anchor {
                owners[id, default: []].append(index)
            }
        }
        for (_, indexes) in owners where indexes.count > 1 {
            let names = indexes.map { rows[$0].name }
            for index in indexes {
                let other = names.filter { $0 != rows[index].name }.first ?? names[0]
                let ref = rows[index].target.handleID ?? ""
                rows[index].duplicateNote = other + "|" + ref
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
                anchor: true, assessment: "strong", reason: "fixed",
                target: .handle(id: anchor.id, ref: anchor.ref, title: anchor.title),
                menu: [], comparisons: [], pins: [], confidenceGradeID: nil, argument: "",
                decided: false, updatedNote: nil, conflictNote: nil, duplicateNote: nil
            )
        }
        let menu = menu(for: proposal, kind: fact.kind)
        let drafted = draftTarget(proposal, menu: menu)
        let pins = Set(proposal.comparisons.filter(\.pinned).map(\.id))
        let argument = proposal.comparisons
            .filter { $0.outcome == "agree" && !$0.incomingDisplay.isEmpty }
            .map(\.incomingDisplay)
            .joined(separator: "; ")
        if let previous, previous.decided {
            var kept = previous
            kept.assessment = proposal.assessment
            kept.reason = reason(proposal, kind: fact.kind)
            kept.comparisons = proposal.comparisons
            if proposal.conflictWithFixed, let rival = menu.first {
                kept.conflictNote = rival.ref
            }
            return kept
        }
        var updated: String?
        if let previous, !previous.anchor, previous.target.token != drafted.token {
            updated = drafted.token
        }
        return Row(
            subjectID: fact.id, kind: fact.kind, name: fact.name, ref: fact.ref,
            anchor: false, assessment: proposal.assessment,
            reason: reason(proposal, kind: fact.kind),
            target: drafted, menu: menu, comparisons: proposal.comparisons,
            pins: pins, confidenceGradeID: nil, argument: argument,
            decided: false, updatedNote: updated, conflictNote: nil, duplicateNote: nil
        )
    }

    private static func draftTarget(_ proposal: CatalogPromoteGraphAlignmentRow, menu: [Alternative]) -> Target {
        if proposal.assessment == "strong", proposal.target == "handle", let first = menu.first, first.id == proposal.handleID {
            return .handle(id: first.id, ref: first.ref, title: first.title)
        }
        if proposal.assessment == "strong", proposal.target == "new" {
            return .newKind
        }
        return .skip
    }

    private static func menu(for proposal: CatalogPromoteGraphAlignmentRow, kind: EvidencePrimaryKind) -> [Alternative] {
        var out: [Alternative] = []
        if proposal.target == "handle", !proposal.handleID.isEmpty {
            out.append(Alternative(id: proposal.handleID, ref: proposal.handleRef, title: headerTitle(proposal)))
        }
        for alt in proposal.alternatives where alt.handleID != proposal.handleID {
            out.append(Alternative(id: alt.handleID, ref: alt.handleRef, title: headerTitle(alt, kind: kind)))
        }
        return out
    }

    private static func headerTitle(_ row: CatalogPromoteGraphAlignmentRow) -> String {
        if let person = row.person, let name = person.name?.form, !name.isEmpty { return name }
        if let event = row.event { return event.eventName.isEmpty ? row.handleRef : event.eventName }
        if let place = row.place, let name = place.names.first, !name.isEmpty { return name }
        return row.handleRef
    }

    private static func headerTitle(_ alt: CatalogPromoteGraphAlignmentAlternative, kind: EvidencePrimaryKind) -> String {
        switch kind {
        case .person:
            if let name = alt.person?.name?.form, !name.isEmpty { return name }
        case .event:
            if let name = alt.event?.eventName, !name.isEmpty { return name }
        case .place:
            if let name = alt.place?.names.first, !name.isEmpty { return name }
        }
        return alt.handleRef
    }

    private static func reason(_ row: CatalogPromoteGraphAlignmentRow, kind _: EvidencePrimaryKind) -> String {
        if !row.viaNeighborSubjectID.isEmpty {
            let role = row.viaRole.isEmpty ? row.viaBridgeType : row.viaRole
            return row.viaNeighborSubjectID + "|" + role
        }
        return row.reasons.first ?? ""
    }
}
