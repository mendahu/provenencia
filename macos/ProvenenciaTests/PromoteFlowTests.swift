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
            reason: "agrees", reasonPropertyKey: "name", reasonPropertyOrigin: "provenencia",
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

    private func fact(_ id: String, anchor: Bool = false, handleID: String = "e-anchor") -> PromoteFlow.SubjectFact {
        PromoteFlow.SubjectFact(
            id: id, kind: .person, name: id, ref: "CPR-\(id)",
            anchor: anchor ? PromoteFlow.Anchor(id: handleID, ref: "PER-A", title: "Filed") : nil
        )
    }

    @Test func matchMenusUseConclusionTitles() {
        let birth = CatalogEventTitle(rule: .typeAtPlace, ref: "EVT-D1JBR", typeLabel: "Birth", place: "Gumpertz")
        var event = row(id: "e", assessment: "medium")
        event.kind = "event"
        event.handleRef = "EVT-D1JBR"
        event.event = CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-D1JBR", subjectTypeID: "t", label: ""),
            eventName: "",
            title: birth
        )
        event.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(
                handleID: "e2", handleRef: "EVT-2", score: 1,
                event: CatalogEventHeader(
                    entity: CatalogCanonicalEntity(id: "e2", ref: "EVT-2", subjectTypeID: "t", label: ""),
                    title: CatalogEventTitle(rule: .label, label: "Working title", ref: "EVT-2")
                )
            ),
        ]

        var person = row(id: "p", assessment: "medium")
        person.person = CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: "Grandpa"),
            name: nil,
            nameValueCount: 0
        )

        var place = row(id: "l", assessment: "medium")
        place.kind = "place"
        place.handleRef = "PLC-1"
        place.place = CatalogPlaceHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PLC-1", subjectTypeID: "t", label: "The farm"),
            names: [],
            parents: ["Ontario", "Canada"]
        )

        var flow = PromoteFlow(entryID: "e")
        flow.load(
            entryID: "e",
            proposal: proposal(rows: [event, person, place]),
            subjects: [
                PromoteFlow.SubjectFact(id: "e", kind: .event, name: "e", ref: "CPR-e"),
                PromoteFlow.SubjectFact(id: "p", kind: .person, name: "p", ref: "CPR-p"),
                PromoteFlow.SubjectFact(id: "l", kind: .place, name: "l", ref: "CPR-l"),
            ],
            bridges: []
        )
        let eventRow = flow.rows.first { $0.subjectID == "e" }
        let personRow = flow.rows.first { $0.subjectID == "p" }
        let placeRow = flow.rows.first { $0.subjectID == "l" }
        #expect(eventRow?.menu.map(\.title) == [
            EventTitleDisplay.title(birth),
            "Working title",
        ])
        #expect(eventRow?.target == .handle(id: "e1", ref: "EVT-D1JBR", title: EventTitleDisplay.title(birth)))
        #expect(personRow?.menu.first?.title == "Grandpa")
        #expect(placeRow?.menu.first?.title == "The farm")
        #expect(placeRow?.menu.first?.subtitle == PlaceChainDisplay.line(parents: ["Ontario", "Canada"]))
    }

    @Test func aMediumMatchStartsOnItsHandle() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a", assessment: "medium")]),
            subjects: [fact("a")],
            bridges: []
        )
        #expect(flow.rows.first?.target == .handle(id: "e1", ref: "PER-1", title: "PER-1"))
        #expect(flow.rows.first?.selectionReadout == PromoteFlow.SelectionReadout(
            assessment: .medium, reason: .agrees(key: "name", origin: "provenencia")
        ))
    }

    @Test func aWeakMatchStartsOnItsHandle() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a", assessment: "weak")]),
            subjects: [fact("a")],
            bridges: []
        )
        let ada = flow.rows.first
        #expect(ada?.target == .handle(id: "e1", ref: "PER-1", title: "PER-1"))
        #expect(ada?.selectionReadout == PromoteFlow.SelectionReadout(
            assessment: .weak, reason: .agrees(key: "name", origin: "provenencia")
        ))
        #expect(flow.targetsChosen)
    }

    /// A score under the accept bar is still a named candidate. The menu
    /// opens on it; only a no-match stays empty.
    @Test func aWeakCandidateBelowTheBarStartsSelected() {
        var below = row(id: "a", target: "skip", handleID: "", assessment: "weak")
        below.reason = "weak"
        below.comparisons = []
        below.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(
                handleID: "e9", handleRef: "PER-9", score: 1,
                assessment: "weak", reason: "weak"
            ),
        ]
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [below]),
            subjects: [fact("a")],
            bridges: []
        )
        let ada = flow.rows.first
        #expect(ada?.target == .handle(id: "e9", ref: "PER-9", title: "PER-9"))
        #expect(ada?.selectionReadout?.assessment == .weak)
        #expect(flow.targetsChosen)
    }

    @Test func theAssessmentFollowsTheSelectedRecord() {
        var chosen = row(id: "a", assessment: "medium")
        chosen.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(
                handleID: "e2", handleRef: "PER-2", score: 2.5,
                assessment: "weak", reason: "agrees",
                reasonPropertyKey: "name", reasonPropertyOrigin: "provenencia"
            ),
        ]
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [chosen]),
            subjects: [fact("a")],
            bridges: []
        )
        #expect(flow.rows.first?.selectionReadout?.assessment == .medium)

        flow.setTarget(subjectID: "a", token: "handle:e2")
        #expect(flow.rows.first?.selectionReadout == PromoteFlow.SelectionReadout(
            assessment: .weak, reason: .agrees(key: "name", origin: "provenencia")
        ))

        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.rows.first?.selectionReadout == nil)
        flow.setTarget(subjectID: "a", token: "new")
        #expect(flow.rows.first?.selectionReadout == nil)

        flow.setTarget(subjectID: "a", token: "handle:e2")
        var rescored = row(id: "a", handleID: "e2", assessment: "medium")
        rescored.handleRef = "PER-2"
        rescored.reason = "decided"
        flow.merge(proposal(rows: [rescored], revision: 4), subjects: [fact("a")])
        #expect(flow.rows.first?.selectionReadout == PromoteFlow.SelectionReadout(
            assessment: .medium, reason: .agrees(key: "name", origin: "provenencia")
        ))
    }

    @Test func anUnmatchedRowStartsUnselected() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a", target: "new", handleID: "", assessment: "none"),
                row(id: "b", assessment: "strong"),
            ]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        #expect(flow.rows.first { $0.subjectID == "a" }?.target == PromoteFlow.Target.unset)
        #expect(flow.rows.first { $0.subjectID == "a" }?.selectionReadout == nil)
        #expect(flow.targetsChosen == false)
        flow.mapRest()
        #expect(flow.batchRows().map(\.subjectID) == ["b"])

        flow.setTarget(subjectID: "a", token: "new")
        #expect(flow.rows.first { $0.subjectID == "a" }?.target == .newKind)
        #expect(flow.targetsChosen)
        #expect(flow.batchRows().first { $0.subjectID == "a" }?.target == "new")

        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.rows.first { $0.subjectID == "a" }?.target == .skip)
        #expect(flow.batchRows().first { $0.subjectID == "a" }?.target == "skip")
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
        #expect(bea?.updated == true)
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
        #expect(flow.connectionLines().first?.state == .off)
    }

    @Test func aDeclinedBridgeStartsOffAndStaysDeclinedOnDone() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a")]),
            subjects: [fact("a")],
            bridges: [
                PromoteFlow.BridgeFact(id: "kept-off", sentence: "rejected", endA: "a", endB: "b", declined: true),
                PromoteFlow.BridgeFact(id: "filed", sentence: "filed", endA: "a", endB: "b", alreadyFiled: true, declined: true),
            ]
        )
        #expect(flow.connectionLines().map(\.id) == ["kept-off"])
        #expect(flow.connectionLines().first?.state == .off)
        #expect(flow.skipBridgeIDs() == ["kept-off"])
        #expect(flow.connectionChanges == 0)
        #expect(!flow.manual)

        // Switching it back on drops it from the declined set Done sends.
        flow.toggleBridge("kept-off")
        #expect(flow.skipBridgeIDs().isEmpty)
        #expect(flow.connectionChanges == 1)
    }

    @Test func promoteAllStartsExpandedAndHidesFiledBridges() {
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), row(id: "b")]),
            subjects: [fact("a", anchor: true, handleID: "e-a"), fact("b", anchor: true, handleID: "e-b")],
            bridges: [
                PromoteFlow.BridgeFact(id: "old", sentence: "old", endA: "a", endB: "b", alreadyFiled: true),
                PromoteFlow.BridgeFact(id: "new", sentence: "new", endA: "a", endB: "b"),
            ]
        )
        #expect(flow.visibleRows.map(\.subjectID) == ["a", "b"])
        #expect(flow.targetsChosen)
        #expect(flow.connectionLines().map(\.id) == ["new"])
        #expect(flow.connectionLines().first?.state == .files)
        #expect(flow.connectionCount == 1)
        #expect(flow.skipBridgeIDs().isEmpty)
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

    @Test func aSkipRowFilesNoArgumentConfidenceOrPins() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a", assessment: "weak"), row(id: "b")]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        flow.mapRest()
        flow.setTarget(subjectID: "a", token: "skip")
        flow.setConfidence(subjectID: "b", gradeID: "grade-1")
        flow.setArgument(subjectID: "b", argument: "names agree")
        flow.setTarget(subjectID: "b", token: "skip")
        for line in flow.batchRows() {
            #expect(line.target == "skip")
            #expect(line.argument.isEmpty)
            #expect(line.confidenceGradeID == nil)
            #expect(line.pairs.isEmpty)
        }
    }

    @Test func aRetargetRedraftsPinsAndArgumentForTheNewTarget() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [row(id: "a")]), subjects: [fact("a")], bridges: [])
        #expect(flow.rows[0].argument == "James")

        flow.setTarget(subjectID: "a", token: "handle:e2")
        #expect(flow.rows[0].pins.isEmpty)
        #expect(flow.rows[0].comparisons.isEmpty)
        #expect(flow.batchRows()[0].pairs.isEmpty)

        var onE2 = row(id: "a", handleID: "e2")
        onE2.comparisons[0].incomingObservationID = "obs-in-2"
        onE2.comparisons[0].memberObservationID = "obs-mem-2"
        onE2.comparisons[0].incomingDisplay = "Jim"
        flow.merge(proposal(rows: [onE2], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].pins == [onE2.comparisons[0].id])
        #expect(flow.rows[0].argument == "Jim")
        #expect(flow.batchRows()[0].pairs.first?.memberObservationID == "obs-mem-2")

        flow.setArgument(subjectID: "a", argument: "my reasons")
        flow.setTarget(subjectID: "a", token: "new")
        #expect(flow.rows[0].argument == "my reasons")
    }

    @Test func aManualPinOnTheSameTargetSurvivesAReproposal() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [row(id: "a")]), subjects: [fact("a")], bridges: [])
        let line = flow.rows[0].comparisons[0].id
        flow.togglePin(subjectID: "a", comparisonID: line)
        #expect(flow.rows[0].pins.isEmpty)
        flow.merge(proposal(rows: [row(id: "a")], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].pins.isEmpty)
    }

    @Test func duplicateNotesNameTheOtherRowAndItsRef() {
        var lost = row(id: "b", target: "skip", handleID: "", assessment: "strong")
        lost.possibleDuplicate = true
        lost.duplicateOfSubjectID = "a"
        var newA = row(id: "c", target: "new", handleID: "", assessment: "strong")
        newA.possibleDuplicate = true
        newA.duplicateOfSubjectID = "d"
        var newB = row(id: "d", target: "new", handleID: "", assessment: "strong")
        newB.possibleDuplicate = true
        newB.duplicateOfSubjectID = "c"
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), lost, newA, newB, row(id: "e", handleID: "e2")]),
            subjects: [fact("a"), fact("b"), fact("c"), fact("d"), fact("e")],
            bridges: []
        )
        let note = { (id: String) in flow.rows.first { $0.subjectID == id }?.duplicateNote }
        #expect(note("b") == .sharesHandle(otherName: "a", ref: "PER-1"))
        #expect(note("c") == .alsoNew(otherName: "d"))
        #expect(note("d") == .alsoNew(otherName: "c"))

        flow.mapRest()
        flow.setTarget(subjectID: "e", token: "handle:e2")
        flow.setTarget(subjectID: "a", token: "handle:e2")
        #expect(note("a") == .sharesHandle(otherName: "e", ref: "PER-2"))
    }

    @Test func aConflictingComparisonIsNotDraftedAsAPin() {
        var proposed = row(id: "a")
        proposed.comparisons.append(CatalogPromoteGraphAlignmentComparison(
            propertyKey: "start_date",
            propertyOrigin: "provenencia",
            outcome: "conflict",
            valueType: "date",
            pinned: true,
            incomingObservationID: "obs-date",
            incomingDisplay: "1990",
            memberObservationID: "obs-date-there",
            memberDisplay: "1985"
        ))
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [proposed]), subjects: [fact("a")], bridges: [])
        let conflict = proposed.comparisons[1].id
        #expect(flow.rows[0].pins.contains(proposed.comparisons[0].id))
        #expect(!flow.rows[0].pins.contains(conflict))
    }

    @Test func aConflictNamesTheStrongerHandleNotTheDecidedOne() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [row(id: "a")]), subjects: [fact("a")], bridges: [])
        flow.setTarget(subjectID: "a", token: "handle:e2")
        var held = row(id: "a", handleID: "e2")
        held.handleRef = "PER-2"
        held.conflictWithFixed = true
        held.alternatives = [CatalogPromoteGraphAlignmentAlternative(handleID: "e1", handleRef: "PER-1", score: 6)]
        flow.merge(proposal(rows: [held], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].conflictNote == "PER-1")
    }
}
