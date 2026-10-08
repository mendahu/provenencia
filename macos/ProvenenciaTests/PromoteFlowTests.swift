import Testing
@testable import Provenencia

@Suite
struct PromoteFlowTests {
    private func proposal(rows: [CatalogPromoteGraphAlignmentRow], revision: Int64 = 3) -> CatalogPromoteGraphAlignmentProposal {
        CatalogPromoteGraphAlignmentProposal(revision: revision, rows: rows)
    }

    private func row(
        id: String,
        target: String = "handle",
        handleID: String = "e1",
        assessment: String = "strong",
        pinned: Bool = true
    ) -> CatalogPromoteGraphAlignmentRow {
        CatalogPromoteGraphAlignmentRow(
            subjectID: id,
            kind: "person",
            target: target,
            handleID: handleID,
            handleRef: "PER-1",
            score: 5,
            assessment: assessment,
            reasons: ["agree name"],
            comparisons: [
                CatalogPromoteGraphAlignmentComparison(
                    propertyKey: "name",
                    propertyOrigin: "provenencia",
                    outcome: "agree",
                    valueType: "name",
                    pinned: pinned,
                    incomingObservationID: "obs-in",
                    incomingDisplay: "James",
                    incomingSource: "Census",
                    memberObservationID: "obs-mem",
                    memberDisplay: "James",
                    memberSource: "Register"
                ),
            ],
            alternatives: [
                CatalogPromoteGraphAlignmentAlternative(handleID: "e2", handleRef: "PER-2", score: 1),
            ],
            conflictWithFixed: false,
            possibleDuplicate: false
        )
    }

    private func fact(_ id: String, anchor: Bool = false) -> PromoteFlow.SubjectFact {
        PromoteFlow.SubjectFact(
            id: id, kind: .person, name: id, ref: "CPR-\(id)",
            anchor: anchor ? PromoteFlow.Anchor(id: "e-anchor", ref: "PER-A", title: "Filed") : nil
        )
    }

    @Test func weakAndUnmatchedStartOnSkip() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a", assessment: "weak"),
                row(id: "b", target: "new", handleID: "", assessment: "none"),
            ]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        let ada = flow.rows.first { $0.subjectID == "a" }
        let bea = flow.rows.first { $0.subjectID == "b" }
        #expect(ada?.target == PromoteFlow.Target.skip)
        #expect(ada?.menu.first?.id == "e1")
        #expect(bea?.target == PromoteFlow.Target.skip)
    }

    @Test func decidedRowsSurviveAReproposal() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), row(id: "b", handleID: "e1")]),
            subjects: [fact("a"), fact("b")],
            bridges: [PromoteFlow.BridgeFact(id: "br", sentence: "a wife of b", endA: "a", endB: "b")]
        )
        flow.mapRest()
        flow.setTarget(subjectID: "a", token: "handle:e2")
        var moved = row(id: "b", handleID: "e9")
        moved.handleRef = "PER-9"
        flow.merge(proposal(rows: [row(id: "a"), moved], revision: 4), subjects: [fact("a"), fact("b")])
        let ada = flow.rows.first { $0.subjectID == "a" }
        let bea = flow.rows.first { $0.subjectID == "b" }
        #expect(ada?.target.handleID == "e2")
        #expect(bea?.updatedNote != nil)
        #expect(bea?.decided == false)
    }

    @Test func doneSendsPinsAndASwitchedOffBridge() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a")]),
            subjects: [fact("a")],
            bridges: [PromoteFlow.BridgeFact(id: "br", sentence: "link", endA: "a", endB: "a")]
        )
        flow.toggleBridge("br")
        let batch = flow.batchRows()
        #expect(batch.count == 1)
        #expect(batch[0].pairs.count == 1)
        #expect(batch[0].pairs[0].incomingObservationID == "obs-in")
        #expect(flow.skipBridgeIDs() == ["br"])
        #expect(flow.connectionLines().first?.why == "self" || flow.connectionLines().first?.why == "off")
    }

    @Test func anchorsAreReadOnly() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), row(id: "b")]),
            subjects: [fact("a"), fact("b", anchor: true)],
            bridges: []
        )
        flow.mapRest()
        let bea = flow.rows.first { $0.subjectID == "b" }
        #expect(bea?.anchor == true)
        #expect(flow.setTarget(subjectID: "b", token: "skip") == false)
        #expect(flow.visibleRows.count == 2)
        #expect(flow.restCount == 1)
    }

    @Test func leaveAsksOnlyAfterAManualChange() {
        var flow = PromoteFlow(entryID: "a")
        #expect(flow.requestLeave(.back) == false)
        flow.setArgument(subjectID: "missing", argument: "x")
        #expect(flow.manual == false)
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a")]),
            subjects: [fact("a")],
            bridges: []
        )
        #expect(flow.requestLeave(.back) == false)
        flow.setArgument(subjectID: "a", argument: "because")
        #expect(flow.requestLeave(.back) == true)
    }
}
