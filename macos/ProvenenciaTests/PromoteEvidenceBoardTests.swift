import Foundation
import Testing
@testable import Provenencia

/// The evidence board: the sheet's words, made once per load.
@Suite
struct PromoteEvidenceBoardTests {
    // MARK: - Fixtures

    private func pin(_ id: String, source: String = "src-1", value: CatalogConclusionValue = .name(CatalogNameValue(form: "John Smith"))) -> CatalogClaimEvidencePin {
        CatalogClaimEvidencePin(id: id, sourceID: source, value: value)
    }

    private func line(_ id: String, pins: [String], incoming: CatalogConclusionValue = .name(CatalogNameValue(form: "John Smith"))) -> CatalogClaimEvidenceLine {
        CatalogClaimEvidenceLine(
            propertyKey: "name", outcome: .agree, relation: .same,
            incomingObservationID: id, incoming: incoming,
            members: pins.map { pin($0) }
        )
    }

    private func build(_ evidence: CatalogClaimEvidence, neighbor: @escaping (String) -> PromoteFlow.Row? = { _ in nil }) -> EvidenceBoard {
        EvidenceBoard.build(
            evidence: evidence, neighbor: neighbor,
            labels: { $0.capitalized },
            worded: { value in
                switch value {
                case .name(let name): name.form
                case .depiction: "likeness"
                default: ""
                }
            }
        )
    }

    /// A sheet of 200 lines of one property, each with one record, and a page row on it.
    private func twoHundredLines() -> (board: EvidenceBoard, flow: PromoteFlow) {
        let lines = (0..<200).map { line("in-\($0)", pins: ["l:\($0)"]) }
        let board = build(CatalogClaimEvidence(
            sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: lines)],
            offers: (0..<200).map { CatalogClaimEvidenceOffer(id: "l:\($0)", checked: true) },
            sourceTitles: ["src-1": "Census 1851"]
        ))
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [
                CatalogPromoteGraphAlignmentRow(
                    subjectID: "a", kind: "person", target: "handle", handleID: "e1", handleRef: "PER-1",
                    score: 5, assessment: "strong", reason: .agrees, reasonPropertyKey: "name", reasonPropertyOrigin: "provenencia",
                    alternatives: [], conflictWithFixed: false, possibleDuplicate: false
                ),
            ]),
            subjects: [PromoteFlow.SubjectFact(id: "a", kind: .person, name: "a", ref: "CPR-a", anchor: nil)],
            bridges: []
        )
        flow.setEvidence(subjectID: "a", handleID: "e1", board: board)
        return (board, flow)
    }

    /// What the sheet hands one line: the line, its own checked pins and its disclosure.
    private func inputs(_ board: EvidenceBoard, pins: Set<String>, expanded: Set<String> = []) -> [String] {
        board.cards.flatMap(\.lines).map { line in
            "\(line.id)|\(pins.intersection(line.pinIDs).sorted())|\(expanded.contains(line.key))"
        }
    }

    // MARK: - T-6

    @Test func aKeystrokeLeavesLinesEqual() {
        var (board, flow) = twoHundredLines()
        let before = flow.rows[0].board
        let inputsBefore = inputs(board, pins: flow.rows[0].pins)
        flow.setArgument(subjectID: "a", argument: "J")
        flow.setArgument(subjectID: "a", argument: "Jo")
        #expect(flow.rows[0].board == before)
        #expect(flow.rows[0].board?.cards.flatMap(\.lines) == board.cards.flatMap(\.lines))
        #expect(inputs(flow.rows[0].board ?? board, pins: flow.rows[0].pins) == inputsBefore)
        board = flow.rows[0].board ?? board
        #expect(board.cards.flatMap(\.lines).count == 200)
    }

    @Test func aToggleChangesOneLine() {
        var (board, flow) = twoHundredLines()
        let pinsBefore = flow.rows[0].pins
        let before = inputs(board, pins: pinsBefore)
        let states = board.cards.flatMap(\.lines).map { PromoteEvidenceDisplay.checkState(pinIDs: $0.pinIDs, pins: pinsBefore) }
        flow.togglePin(subjectID: "a", pinIDs: ["l:7"])
        board = flow.rows[0].board ?? board
        let after = inputs(board, pins: flow.rows[0].pins)
        let statesAfter = board.cards.flatMap(\.lines).map { PromoteEvidenceDisplay.checkState(pinIDs: $0.pinIDs, pins: flow.rows[0].pins) }
        #expect(zip(before, after).enumerated().filter { $0.element.0 != $0.element.1 }.map(\.offset) == [7])
        #expect(zip(states, statesAfter).enumerated().filter { $0.element.0 != $0.element.1 }.map(\.offset) == [7])
        #expect(statesAfter[7] == .unchecked)
    }

    // MARK: - What the board words

    @Test func nothingToCompareWhenNoLineOrRelationship() {
        #expect(build(CatalogClaimEvidence()).hasNothingToCompare)
        #expect(build(CatalogClaimEvidence(sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [])])).hasNothingToCompare)
        let withLine = CatalogClaimEvidence(sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [line("in", pins: ["p1"])])])
        #expect(!build(withLine).hasNothingToCompare)
        let relationship = CatalogClaimEvidenceRelationship(
            neighborSubjectID: "n", key: "k", bridgeType: "participation", outcome: .agree,
            bundles: [CatalogClaimEvidenceBundle(id: "b1")]
        )
        #expect(!build(CatalogClaimEvidence(relationships: [relationship])).hasNothingToCompare)
    }

    @Test func aLineCarriesItsWordsAndItsPins() {
        let evidence = CatalogClaimEvidence(
            sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [line("in-1", pins: ["p1", "p2"])])],
            offers: [CatalogClaimEvidenceOffer(id: "p1", checked: true), CatalogClaimEvidenceOffer(id: "p2", checked: false)],
            sourceTitles: ["src-1": "Census 1851"]
        )
        let board = build(evidence)
        #expect(board.offeredPinIDs == ["p1", "p2"])
        #expect(board.checkedPinIDs == ["p1"])
        let card = board.cards[0]
        #expect(card.id == "name")
        #expect(card.header == "they have a similarity in name")
        #expect(card.thisSource == "John Smith")
        let first = card.lines[0]
        #expect(first.pinIDs == ["p1", "p2"] && first.key == "p1" && first.collapses)
        #expect(first.value == .text("John Smith"))
        #expect(first.source == "Census 1851")
        #expect(first.members.map(\.source) == ["Census 1851", "Census 1851"])
        #expect(first.thisSource == nil)
        #expect(board.checkedCount(pins: ["p1", "x"]) == 1)
    }

    @Test func aSourceWithNoTitleIsARecordOnEveryLineAndMember() {
        let evidence = CatalogClaimEvidence(
            sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [line("in-1", pins: ["p1", "p2"])])],
            sourceTitles: ["src-1": ""]
        )
        let first = build(evidence).cards[0].lines[0]
        #expect(first.source == "a record")
        #expect(first.members.map(\.source) == ["a record", "a record"])
    }

    // MARK: - T-11

    @Test func spokenCountsAndUntitled() {
        func spouse(_ bundles: Int) -> CatalogClaimEvidenceRelationship {
            CatalogClaimEvidenceRelationship(
                neighborSubjectID: "n1", key: "relationship·spouse", bridgeType: "relationship", outcome: .agree,
                bundles: (0..<bundles).map { CatalogClaimEvidenceBundle(id: "b\($0)") }
            )
        }
        let three = build(CatalogClaimEvidence(relationships: [spouse(3)])).relationships[0].spoken
        #expect(three.hasSuffix("Agrees, 3 records."))
        let one = build(CatalogClaimEvidence(relationships: [spouse(1)])).relationships[0].spoken
        #expect(one.hasSuffix("Agrees.") && !one.contains("records"))

        // A likeness of two records.
        func crop(_ path: String) -> CatalogConclusionValue {
            .depiction(CatalogDepictionValue(regionJSON: "", imageRelPath: path))
        }
        var likeness = line("in-1", pins: [], incoming: crop("objects/a.png"))
        likeness.propertyKey = "likeness"
        likeness.members = [pin("p1", value: crop("objects/b.png")), pin("p2", value: crop("objects/c.png"))]
        let board = build(CatalogClaimEvidence(
            sections: [CatalogClaimEvidenceSection(propertyKey: "likeness", lines: [likeness])],
            sourceTitles: ["src-1": "Funeral card"]
        ))
        #expect(board.cards[0].lines[0].spoken.hasSuffix(", 2 records."))
        #expect(board.cards[0].lines[0].spoken.hasPrefix("Likeness from this source and Funeral card."))

        // A member under an untitled Source is spoken as shown.
        let untitled = build(CatalogClaimEvidence(
            sections: [CatalogClaimEvidenceSection(propertyKey: "name", lines: [line("in-2", pins: ["p1", "p2"])])],
            sourceTitles: [:]
        ))
        #expect(untitled.cards[0].lines[0].members.map(\.spoken) == ["a record: John Smith.", "a record: John Smith."])
        #expect(untitled.cards[0].lines[0].members.map(\.source) == ["a record", "a record"])
    }

    @Test func aLikenessShowsItsTwoCropsAndSaysNoThisSource() {
        let crop = CatalogConclusionValue.depiction(CatalogDepictionValue(regionJSON: "", imageRelPath: "objects/a.png"))
        let memberCrop = CatalogConclusionValue.depiction(CatalogDepictionValue(regionJSON: "", imageRelPath: "objects/b.png"))
        var likeness = line("in-1", pins: [], incoming: crop)
        likeness.propertyKey = "likeness"
        likeness.members = [pin("p1", value: memberCrop)]
        let board = build(CatalogClaimEvidence(sections: [CatalogClaimEvidenceSection(propertyKey: "likeness", lines: [likeness])]))
        let first = board.cards[0].lines[0]
        #expect(first.isLikeness)
        #expect(first.value == .crops(["objects/a.png", "objects/b.png"]))
        #expect(first.thisSource == nil)
        #expect(board.cards[0].thisSource == nil)
    }

    @Test func differingValuesSayThisSourceUnderEachLine() {
        var other = line("in-2", pins: ["p2"], incoming: .name(CatalogNameValue(form: "J. Smith")))
        other.propertyKey = "name"
        let board = build(CatalogClaimEvidence(sections: [
            CatalogClaimEvidenceSection(propertyKey: "name", lines: [line("in-1", pins: ["p1"]), other]),
        ]))
        #expect(board.cards[0].thisSource == nil)
        #expect(board.cards[0].lines.map(\.thisSource) == ["This source: John Smith", "This source: J. Smith"])
    }

    @Test func aRelationshipCardIsWordedFromTheNeighborsSide() {
        var spouse = CatalogClaimEvidenceRelationship(
            neighborSubjectID: "n1", key: "relationship·spouse", bridgeType: "relationship", outcome: .partial,
            bundles: [CatalogClaimEvidenceBundle(id: "b1"), CatalogClaimEvidenceBundle(id: "b2")]
        )
        spouse.neighborTerm = CatalogEvidenceTerm(key: "spouse", label: "Spouse", origin: CatalogOrigin.provenencia)
        spouse.neighborTermNamesNeighbor = true
        let board = build(CatalogClaimEvidence(relationships: [spouse]))
        let card = board.relationships[0]
        #expect(card.header == "they share a similar spouse")
        #expect(card.pinIDs == ["b1", "b2"] && card.recordCount == 2)
        #expect(card.outcome == .partial)
        #expect(card.groups.isEmpty)
        #expect(card.id == "n1|relationship·spouse")
    }
}
