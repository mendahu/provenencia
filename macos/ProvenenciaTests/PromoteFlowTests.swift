import Testing
@testable import Provenencia

@Suite
struct PromoteFlowTests {
    private func proposal(rows: [CatalogPromoteGraphAlignmentRow], revision: Int64 = 3) -> CatalogPromoteGraphAlignmentProposal {
        CatalogPromoteGraphAlignmentProposal(revision: revision, rows: rows)
    }

    private func row(
        id: String,
        target: PromoteTargetCode = .handle,
        handleID: String = "e1",
        assessment: String = "strong"
    ) -> CatalogPromoteGraphAlignmentRow {
        CatalogPromoteGraphAlignmentRow(
            subjectID: id,
            kind: "person",
            target: target,
            handleID: handleID,
            handleRef: "PER-1",
            score: 5,
            assessment: assessment,
            reason: .agrees, reasonPropertyKey: "name", reasonPropertyOrigin: "provenencia",
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

    @Test func aRowCarriesItsEventsDate() {
        for anchored in [true, false] {
            var flow = PromoteFlow(entryID: "m")
            flow.load(
                entryID: "m",
                proposal: proposal(rows: [row(id: "m", assessment: "medium")]),
                subjects: [
                    PromoteFlow.SubjectFact(
                        id: "m", kind: .event, name: "Marriage", ref: "EVT-M",
                        anchor: anchored ? PromoteFlow.Anchor(id: "e-anchor", ref: "EVT-A", title: "Filed") : nil,
                        date: "12 May 1842"
                    ),
                ],
                bridges: []
            )
            #expect(flow.row("m")?.date == "12 May 1842")
        }
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

    @Test func aPersonRecordReadsNaturallyInTheSheetAndSortedInTheMenu() {
        var person = row(id: "p", assessment: "medium")
        person.person = CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "John Smith", parts: [
                CatalogNameValuePart(value: "John", type: "given"),
                CatalogNameValuePart(value: "Smith", type: "surname"),
            ]),
            nameValueCount: 1
        )
        var flow = PromoteFlow(entryID: "e")
        flow.nameOrders = .givenSurname
        flow.load(entryID: "e", proposal: proposal(rows: [person]), subjects: [fact("p")], bridges: [])
        let personRow = flow.rows.first
        #expect(personRow?.menu.first?.title == "Smith, John")
        #expect(personRow?.menu.first?.readingTitle == "John Smith")
        #expect(personRow?.readingTitle == "John Smith")
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

    /// Promote suggests no record for a row it did not place, whatever its
    /// band, so Skip and New are the researcher's choice, never a draft.
    @Test func anUnplacedRowNeverDraftsNewOrSkip() {
        for band in ["medium", "strong"] {
            var unplaced = row(id: "a", target: .unplaced, handleID: "", assessment: band)
            unplaced.reason = .unplaced
            unplaced.alternatives = [
                CatalogPromoteGraphAlignmentAlternative(handleID: "e2", handleRef: "PER-2", score: 2.5, assessment: band, reason: .agrees),
            ]
            var flow = PromoteFlow(entryID: "a")
            flow.load(entryID: "a", proposal: proposal(rows: [unplaced]), subjects: [fact("a")], bridges: [])
            #expect(flow.rows.first?.target == .unset, "band \(band)")
            #expect(flow.targetsChosen == false, "band \(band)")
        }
    }

    /// A row that had a suggested record and loses it on a re-proposal opens
    /// unset and says no record is suggested.
    @Test func aRowThatLosesItsRecordSaysNoRecordIsSuggested() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), row(id: "b", handleID: "e1")]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        var lost = row(id: "b", target: .unplaced, handleID: "", assessment: "none")
        lost.reason = .unplaced
        lost.alternatives = []
        flow.merge(proposal(rows: [row(id: "a"), lost], revision: 4), subjects: [fact("a"), fact("b")])
        let bea = flow.rows.first { $0.subjectID == "b" }
        #expect(bea?.target == .unset)
        #expect(bea?.updated == true)
        #expect(bea?.captionReason == .unplaced)
    }

    /// A row Promote suggests no record for, drafted onto a weak record by the
    /// page, captions that record's reason, not "No record suggested".
    @Test func anUnplacedRowDraftedOntoAWeakRecordCaptionsItsReason() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [row(id: "a")]), subjects: [fact("a")], bridges: [])
        var weak = row(id: "a", target: .unplaced, handleID: "", assessment: "weak")
        weak.reason = .unplaced
        weak.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(
                handleID: "e9", handleRef: "PER-9", score: 1,
                assessment: "weak", reason: .weak
            ),
        ]
        flow.merge(proposal(rows: [weak], revision: 4), subjects: [fact("a")])
        let ada = flow.rows.first
        #expect(ada?.updated == true)
        #expect(ada?.target == .handle(id: "e9", ref: "PER-9", title: "PER-9"))
        #expect(ada?.captionReason == .weak)
    }

    /// A row whose record another row took starts open, whatever its band,
    /// so Done can't file two subjects on one handle without a choice.
    @Test func aRowWhoseRecordWasTakenStartsUnset() {
        let winner = row(id: "a", handleID: "e1", assessment: "weak")
        for band in ["weak", "medium", "strong"] {
            var loser = row(id: "b", target: .unplaced, handleID: "", assessment: band)
            loser.reason = .taken
            loser.possibleDuplicate = true
            loser.duplicateOfSubjectID = "a"
            loser.alternatives = [
                CatalogPromoteGraphAlignmentAlternative(handleID: "e1", handleRef: "PER-1", score: 2.5, assessment: band, reason: .agrees),
            ]
            var flow = PromoteFlow(entryID: "a")
            flow.mappedRest = true
            flow.load(entryID: "a", proposal: proposal(rows: [winner, loser]), subjects: [fact("a"), fact("b")], bridges: [])
            let bea = flow.rows.first { $0.subjectID == "b" }
            #expect(bea?.target == .unset, "band \(band)")
            #expect(bea?.duplicateNote == .sharesHandle(otherName: "a", ref: "PER-1"), "band \(band)")
            #expect(flow.targetsChosen == false, "band \(band)")
        }
    }

    /// A score under the accept bar is still a named candidate. The menu
    /// opens on it; a row with no band, or a medium or strong one Promote did
    /// not place, stays empty.
    @Test func aWeakCandidateBelowTheBarStartsSelected() {
        var below = row(id: "a", target: .unplaced, handleID: "", assessment: "weak")
        below.reason = .unplaced
        below.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(
                handleID: "e9", handleRef: "PER-9", score: 1,
                assessment: "weak", reason: .weak
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
                assessment: "weak", reason: .agrees,
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
            assessment: .weak, reason: .agrees(key: "name", origin: "provenencia"), chosen: true
        ))

        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.rows.first?.selectionReadout == nil)
        flow.setTarget(subjectID: "a", token: "new")
        #expect(flow.rows.first?.selectionReadout == nil)

        flow.setTarget(subjectID: "a", token: "handle:e2")
        var rescored = row(id: "a", handleID: "e2", assessment: "medium")
        rescored.handleRef = "PER-2"
        rescored.reason = .decided
        flow.merge(proposal(rows: [rescored], revision: 4), subjects: [fact("a")])
        #expect(flow.rows.first?.selectionReadout == PromoteFlow.SelectionReadout(
            assessment: .medium, reason: .agrees(key: "name", origin: "provenencia"), chosen: true
        ))
    }

    @Test func anUnmatchedRowStartsUnselected() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a", target: .unplaced, handleID: "", assessment: "none"),
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

    @Test func theBridgeBetweenTwoSubjectsIsTheFirstOnThePageEitherWayRound() {
        var flow = PromoteFlow(entryID: "john")
        flow.bridges = [
            PromoteFlow.BridgeFact(id: "other", sentence: "", endA: "mary", endB: "birth"),
            PromoteFlow.BridgeFact(id: "first", sentence: "", endA: "birth", endB: "john"),
            PromoteFlow.BridgeFact(id: "second", sentence: "", endA: "john", endB: "birth"),
        ]
        #expect(flow.bridge(between: "john", and: "birth")?.id == "first")
        #expect(flow.bridge(between: "birth", and: "john")?.id == "first")
        #expect(flow.bridge(between: "john", and: "death") == nil)
    }

    @Test func aContestedRowAndItsMoveReachTheFlow() {
        var a = row(id: "a")
        a.contested = true
        a.moveReason = .neighborMoved
        a.moveNeighborSubjectID = "b"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a, row(id: "b", handleID: "e2")]), subjects: [fact("a"), fact("b")], bridges: [])
        let rowA = flow.rows.first { $0.subjectID == "a" }
        let rowB = flow.rows.first { $0.subjectID == "b" }
        #expect(rowA?.contested == true)
        #expect(rowA?.move == .neighborMoved(neighborID: "b"))
        #expect(rowB?.contested == false)
        #expect(rowB?.move == nil)

        func move(_ code: PromoteReasonCode?, neighbor: String) -> PromoteFlow.Reason? {
            var r = row(id: "a")
            r.moveReason = code
            r.moveNeighborSubjectID = neighbor
            return PromoteFlow.Reason(move: r)
        }
        #expect(move(nil, neighbor: "b") == nil)
        #expect(move(.neighborReleased, neighbor: "") == .empty)
        #expect(move(.unrecognized("later_move"), neighbor: "b") == .empty)
    }

    @Test func aContestedRowShowsContestedInItsReadout() {
        var a = row(id: "a")
        a.contested = true
        var b = row(id: "b", handleID: "e3")
        b.handleRef = "PER-3"
        var c = row(id: "c", target: .unplaced, handleID: "", assessment: "weak")
        c.contested = true
        c.alternatives[0].assessment = "weak"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a, b, c]), subjects: [fact("a"), fact("b"), fact("c")], bridges: [])
        let rowA = flow.rows.first { $0.subjectID == "a" }
        let rowB = flow.rows.first { $0.subjectID == "b" }
        let rowC = flow.rows.first { $0.subjectID == "c" }
        #expect(rowA?.selectionReadout?.contested == true)
        #expect(rowB?.selectionReadout?.contested == false)
        #expect(rowC?.target.handleID == "e2")
        #expect(rowC?.selectionReadout == PromoteFlow.SelectionReadout(assessment: .weak, reason: .empty, contested: true))
    }

    @Test func aContestedRowWithoutAHandleShowsNoReadout() {
        var n = row(id: "n", target: .unplaced, handleID: "")
        n.contested = true
        var t = row(id: "t", target: .unplaced, handleID: "", assessment: "medium")
        t.reason = .taken
        t.contested = true
        var z = row(id: "z", target: .unplaced, handleID: "", assessment: "none")
        z.contested = true
        var x = row(id: "x")
        x.contested = true
        var flow = PromoteFlow(entryID: "n")
        flow.load(
            entryID: "n", proposal: proposal(rows: [n, t, z, x]),
            subjects: [fact("n"), fact("t"), fact("z"), fact("x", anchor: true)], bridges: []
        )
        for id in ["n", "t", "z", "x"] {
            #expect(flow.rows.first { $0.subjectID == id }?.selectionReadout == nil, "row \(id)")
        }
    }

    @Test func retargetingClearsContestedAndMoveUntilTheNextProposal() {
        var a = row(id: "a")
        a.contested = true
        a.moveReason = .neighborMoved
        a.moveNeighborSubjectID = "b"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a, row(id: "b", handleID: "e3")]), subjects: [fact("a"), fact("b")], bridges: [])
        flow.mapRest()
        flow.setTarget(subjectID: "a", token: "handle:e2")
        let cleared = flow.rows.first { $0.subjectID == "a" }
        #expect(cleared?.contested == false)
        #expect(cleared?.move == nil)
        var next = row(id: "a", handleID: "e2")
        next.contested = true
        flow.merge(proposal(rows: [next], revision: 4), subjects: [fact("a")])
        let again = flow.rows.first { $0.subjectID == "a" }
        #expect(again?.target.handleID == "e2")
        #expect(again?.contested == true)
    }

    @Test func theCaptionReasonIsTheMoveWhenThereIsOne() {
        var a = row(id: "a")
        a.moveReason = .neighborMoved
        a.moveNeighborSubjectID = "b"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a, row(id: "b", handleID: "e3")]), subjects: [fact("a"), fact("b")], bridges: [])
        let rowA = flow.rows.first { $0.subjectID == "a" }
        let rowB = flow.rows.first { $0.subjectID == "b" }
        #expect(rowA?.captionReason == .neighborMoved(neighborID: "b"))
        #expect(rowB?.captionReason == rowB?.reason)
        #expect(rowA?.shownRef == "PER-1")
        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.rows.first { $0.subjectID == "a" }?.shownRef == "CPR-a")
    }

    @Test func aDecidedRowTakesTheLatestContestedFlag() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [row(id: "a")]), subjects: [fact("a")], bridges: [])
        flow.mapRest()
        flow.setTarget(subjectID: "a", token: "handle:e1")
        var contested = row(id: "a")
        contested.contested = true
        flow.merge(proposal(rows: [contested], revision: 4), subjects: [fact("a")])
        let a = flow.rows.first { $0.subjectID == "a" }
        #expect(a?.target.handleID == "e1")
        #expect(a?.decided == true)
        #expect(a?.contested == true)
    }

    @Test func anAnchorRowIsNeverContested() {
        var a = row(id: "a")
        a.contested = true
        a.moveReason = .neighborPlaced
        a.moveNeighborSubjectID = "b"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a]), subjects: [fact("a", anchor: true)], bridges: [])
        let anchor = flow.rows.first { $0.subjectID == "a" }
        #expect(anchor?.anchor == true)
        #expect(anchor?.contested == false)
        #expect(anchor?.move == nil)
    }

    @Test func reasonCodesParseThroughTheWireVocabulary() {
        let wire: [(String, PromoteReasonCode)] = [
            ("via", .via), ("decided", .decided), ("agrees", .agrees), ("weak", .weak), ("taken", .taken),
            ("", .unplaced), ("empty", .empty), ("neighbor_placed", .neighborPlaced),
            ("neighbor_moved", .neighborMoved), ("neighbor_released", .neighborReleased),
        ]
        for (code, expected) in wire {
            #expect(PromoteReasonCode(wire: code) == expected)
        }
        #expect(PromoteReasonCode(wire: "later_code") == .unrecognized("later_code"))
        #expect(PromoteReasonCode(wire: "no_match") == .unrecognized("no_match"))
        let targets: [(String, PromoteTargetCode)] = [("", .unplaced), ("handle", .handle), ("new", .new), ("skip", .skip)]
        for (code, expected) in targets {
            #expect(PromoteTargetCode(wire: code) == expected)
        }
        #expect(PromoteTargetCode(wire: "later") == .unrecognized("later"))
        for (_, c) in targets {
            #expect(PromoteTargetCode(wire: c.wire) == c)
        }

        func reason(_ code: PromoteReasonCode, neighbor: String = "") -> PromoteFlow.Reason {
            PromoteFlow.Reason(code: code, propertyKey: "name", propertyOrigin: "provenencia", neighborID: neighbor)
        }
        #expect(reason(.via, neighbor: "n") == .via(neighborID: "n"))
        #expect(reason(.via) == .empty)
        #expect(reason(.decided) == .decided)
        #expect(reason(.agrees) == .agrees(key: "name", origin: "provenencia"))
        #expect(reason(.weak) == .weak)
        #expect(reason(.taken) == .taken)
        #expect(reason(.unplaced) == .unplaced)
        #expect(reason(.empty) == .empty)
        #expect(reason(.neighborPlaced, neighbor: "n") == .neighborPlaced(neighborID: "n"))
        #expect(reason(.neighborPlaced) == .empty)
        #expect(reason(.unrecognized("later_code")) == .empty)
    }

    @Test func doneSendsDefaultPinsAndConnectionsFollowEnds() {
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a", handleID: "e1"), row(id: "b", handleID: "e2")]),
            subjects: [fact("a"), fact("b")],
            bridges: [PromoteFlow.BridgeFact(id: "br", sentence: "link", endA: "a", endB: "b")]
        )
        #expect(flow.connections.first?.state == .files)
        flow.setTarget(subjectID: "b", token: "skip")
        #expect(flow.connections.first?.state == .waits(endID: "b", why: .skipped))
        let batch = flow.batchRows()
        #expect(batch.count == 2)
        #expect(batch[0].subjectID == "a")
        #expect(batch[0].defaultPins)
        #expect(batch[0].pinIDs.isEmpty)
        #expect(batch[1].subjectID == "b")
        #expect(batch[1].target == "skip")
        flow.setTarget(subjectID: "b", token: "new")
        #expect(flow.connections.first?.state == .files)
    }

    private func link(
        _ id: String, _ a: String?, _ b: String?,
        sentence: String? = nil, alreadyFiled: Bool = false
    ) -> PromoteFlow.BridgeFact {
        var bridge = PromoteFlow.BridgeFact(id: id, sentence: sentence ?? id, endA: a, endB: b)
        bridge.endAName = a ?? ""
        bridge.endBName = b ?? ""
        bridge.alreadyFiled = alreadyFiled
        return bridge
    }

    private func mapped(
        ids: [String], handleBy: (String) -> String = { "e-" + $0 }, anchors: Set<String> = [],
        bridges: [PromoteFlow.BridgeFact]
    ) -> PromoteFlow {
        var flow = PromoteFlow(entryID: ids[0])
        flow.mappedRest = true
        flow.load(
            entryID: ids[0],
            proposal: proposal(rows: ids.map { row(id: $0, handleID: handleBy($0)) }),
            subjects: ids.map { fact($0, anchor: anchors.contains($0), handleID: handleBy($0)) },
            bridges: bridges
        )
        return flow
    }

    @Test func connectionsShowUnderEveryEndThatIsARow() {
        let flow = mapped(ids: ["j", "b", "t", "o", "m", "n", "a", "c"], bridges: [
            link("pa", "j", "b"),
            link("lo", "b", "t"),
            link("pr", "t", "o"),
            link("di", "m", "n"),
            link("sy", "a", "c"),
        ])
        let byRow = flow.connectionsByRow
        #expect(Set(byRow["j"]?.map(\.id) ?? []) == ["pa"])
        #expect(Set(byRow["b"]?.map(\.id) ?? []) == ["pa", "lo"])
        #expect(Set(byRow["t"]?.map(\.id) ?? []) == ["lo", "pr"])
        #expect(Set(byRow["o"]?.map(\.id) ?? []) == ["pr"])
        #expect(Set(byRow["m"]?.map(\.id) ?? []) == ["di"])
        #expect(Set(byRow["n"]?.map(\.id) ?? []) == ["di"])
        #expect(Set(byRow["a"]?.map(\.id) ?? []) == ["sy"])
        #expect(Set(byRow["c"]?.map(\.id) ?? []) == ["sy"])
        #expect(byRow.values.joined().allSatisfy { $0.otherEnd == nil })
    }

    @Test func aConnectionWithOneRowEndShowsOnlyThere() {
        // (a) The event is an anchor: it shows under the person only, with a note that it is filed.
        let filed = mapped(ids: ["j", "b"], anchors: ["b"], bridges: [link("pa", "j", "b")])
        let filedRows = filed.connectionsByRow
        #expect(filedRows.keys.sorted() == ["j"])
        #expect(filedRows["j"]?.first?.otherEnd == PromoteFlow.OtherEnd(name: "b", filed: true))
        #expect(filedRows["j"]?.first?.bridge.state == .files)

        // (b) Not mapped: the other end is not on this page.
        var alone = PromoteFlow(entryID: "j")
        alone.load(
            entryID: "j",
            proposal: proposal(rows: [row(id: "j"), row(id: "b", handleID: "e2")]),
            subjects: [fact("j"), fact("b")],
            bridges: [link("pa", "j", "b")]
        )
        let shown = alone.connectionsByRow
        #expect(shown.keys.sorted() == ["j"])
        #expect(shown["j"]?.first?.bridge.state == .waits(endID: "b", why: .notOnPage))
        #expect(shown["j"]?.first?.otherEnd == PromoteFlow.OtherEnd(name: "b", filed: false))

        // (c) Both ends anchors: shown nowhere, still counted.
        let both = mapped(ids: ["a", "b"], anchors: ["a", "b"], bridges: [link("x", "a", "b")])
        #expect(both.connectionsByRow.isEmpty)
        #expect(both.connectionCount == 1)

        // (d) Already filed: neither shown nor counted.
        let old = mapped(ids: ["a", "b"], bridges: [link("old", "a", "b", alreadyFiled: true)])
        #expect(old.connectionsByRow.isEmpty)
        #expect(old.connections.isEmpty)
        #expect(old.connectionCount == 0)

        // (e) Neither end is a row: shown nowhere, nothing files.
        let neither = mapped(ids: ["xx", "ww"], bridges: [link("x", "zz", "yy")])
        #expect(neither.connectionsByRow.isEmpty)
        #expect(neither.connectionCount == 0)

        // (f) Both ends are one row: shown once.
        let same = mapped(ids: ["a"], bridges: [link("r", "a", "a")])
        #expect(same.connectionsByRow["a"]?.map(\.id) == ["r"])
    }

    @Test func waitingNamesTheEndAndWhy() {
        let bridges = [link("r", "a", "b")]
        var flow = mapped(ids: ["a", "b"], bridges: bridges)
        flow.setTarget(subjectID: "b", token: "skip")
        #expect(flow.connections.first?.state == .waits(endID: "b", why: .skipped))
        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.connections.first?.state == .waits(endID: "a", why: .skipped))

        var unset = PromoteFlow(entryID: "a")
        unset.mappedRest = true
        unset.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a", handleID: "e1"), row(id: "b", handleID: "", assessment: "none")]),
            subjects: [fact("a"), fact("b")],
            bridges: bridges
        )
        #expect(unset.connections.first?.state == .waits(endID: "b", why: .unset))

        let missing = mapped(ids: ["a", "b"], bridges: [link("r", "a", nil)])
        #expect(missing.connections.first?.state == .waits(endID: nil, why: .notOnPage))
    }

    @Test func bothEndsOnOneRecordNameTheRecord() {
        let flow = mapped(ids: ["a", "b"], handleBy: { _ in "e1" }, bridges: [link("r", "a", "b")])
        #expect(flow.connections.first?.state == .selfLink(ref: "PER-1"))
        #expect(flow.connectionCount == 0)
    }

    @Test func rowConnectionsAreOrderedByStateThenSentence() {
        var flow = mapped(
            ids: ["b", "p1", "p2", "p3", "p4", "p5"],
            handleBy: { $0 == "p5" ? "e-b" : "e-" + $0 },
            bridges: [
                link("x1", "p1", "b", sentence: "d"),
                link("x2", "p2", "b", sentence: "c"),
                link("x3", "p3", "b", sentence: "b"),
                link("x4", "p4", "b", sentence: "a"),
                link("x5", "p5", "b", sentence: "e"),
            ]
        )
        flow.setTarget(subjectID: "p2", token: "skip")
        flow.setTarget(subjectID: "p3", token: "skip")
        #expect(flow.connectionsByRow["b"]?.map(\.bridge.sentence) == ["a", "d", "b", "c", "e"])
    }

    @Test func theUnsetCountCountsRowsWithNoTarget() {
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a", handleID: "e1"),
                row(id: "b", handleID: "", assessment: "none"),
                row(id: "c"),
            ]),
            subjects: [fact("a"), fact("b"), fact("c", anchor: true)],
            bridges: []
        )
        #expect(flow.unsetCount == 1)
        #expect(!flow.targetsChosen)
        flow.setTarget(subjectID: "b", token: "new")
        #expect(flow.unsetCount == 0)
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
        #expect(flow.connections.map(\.id) == ["new"])
        #expect(flow.connections.first?.state == .files)
        #expect(flow.connectionCount == 1)
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

    @Test func anUnresolvedTokenDecidesNothing() {
        var flow = PromoteFlow(entryID: "a")
        let unset = row(id: "a", target: .unplaced, handleID: "", assessment: "none")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [unset, row(id: "b")]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        flow.setTarget(subjectID: "b", token: PromoteFlow.Target.handleToken("e1"))
        let before = flow
        for subject in ["a", "b"] {
            for token in [PromoteFlow.Target.findToken, "bogus", PromoteFlow.Target.handleToken("not-in-menu")] {
                #expect(flow.setTarget(subjectID: subject, token: token) == false)
            }
        }
        #expect(flow == before)
        #expect(flow.rows.map(\.target) == before.rows.map(\.target))
        #expect(flow.rows.map(\.decided) == [false, true])
    }

    @Test func menuTokensAreSpelledAndDecodedOnce() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a")]),
            subjects: [fact("a")],
            bridges: []
        )
        let handle = PromoteFlow.Target.handle(id: "e1", ref: "PER-1", title: "A")
        #expect(handle.token == PromoteFlow.Target.handleToken("e1"))
        let first = flow.rows[0]
        #expect(first.menuItem(forToken: PromoteFlow.Target.handleToken("e1"))?.id == "e1")
        #expect(first.menuItem(forToken: PromoteFlow.Target.newKind.token) == nil)
        #expect(first.menuItem(forToken: PromoteFlow.Target.findToken) == nil)
        flow.setTarget(subjectID: "a", token: PromoteFlow.Target.handleToken("e2"))
        flow.setTarget(subjectID: "a", token: PromoteFlow.Target.handleToken("e1"))
        #expect(flow.rows[0].target.handleID == "e1")
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
            #expect(line.pinIDs.isEmpty)
            #expect(line.defaultPins == false)
        }
    }

    /// A board that offers these pins, checked or not, and has nothing else to say.
    private func evidence(_ offers: [(String, Bool)]) -> EvidenceBoard {
        EvidenceBoard.build(
            evidence: CatalogClaimEvidence(offers: offers.map { CatalogClaimEvidenceOffer(id: $0.0, checked: $0.1) }),
            neighbor: { _ in nil }, labels: { $0 }, worded: { _ in "" }
        )
    }

    private func loaded(_ rows: [CatalogPromoteGraphAlignmentRow] = [], ids: [String] = ["a"]) -> PromoteFlow {
        var flow = PromoteFlow(entryID: ids[0])
        flow.mappedRest = true
        flow.load(
            entryID: ids[0],
            proposal: proposal(rows: rows.isEmpty ? ids.map { row(id: $0) } : rows),
            subjects: ids.map { fact($0) },
            bridges: []
        )
        return flow
    }

    @Test func aNeverOpenedRowSendsDefaultPins() {
        let flow = loaded()
        #expect(flow.rows[0].board == nil)
        let line = flow.batchRows()[0]
        #expect(line.defaultPins)
        #expect(line.pinIDs.isEmpty)
    }

    @Test func openingDraftsTheCheckedPins() {
        var flow = loaded()
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:2", true), ("l:1", true), ("l:3", false)]))
        #expect(flow.rows[0].pins == ["l:1", "l:2"])
        #expect(flow.rows[0].decided == false)
        let line = flow.batchRows()[0]
        #expect(line.defaultPins == false)
        #expect(line.pinIDs == ["l:1", "l:2"])
    }

    @Test func reopeningKeepsOnlyOfferedPins() {
        var flow = loaded()
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true), ("l:2", true), ("l:3", true)]))
        flow.togglePin(subjectID: "a", pinIDs: ["l:2"])
        #expect(flow.rows[0].pins == ["l:1", "l:3"])
        // The neighbor was set to Skip: l:3 is no longer offered, l:2 still is, unchecked.
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true), ("l:2", true)]))
        #expect(flow.rows[0].pins == ["l:1"])
    }

    @Test func aLineTogglesEveryMember() {
        var flow = loaded()
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:a", true), ("l:b", true)]))
        flow.togglePin(subjectID: "a", pinIDs: ["l:a", "l:b"])
        #expect(flow.rows[0].pins.isEmpty)
        #expect(flow.rows[0].decided)
        flow.togglePin(subjectID: "a", pinIDs: ["l:a", "l:b"])
        #expect(flow.rows[0].pins == ["l:a", "l:b"])
        // One member left unchecked makes the line unchecked; a tap checks them all.
        flow.togglePin(subjectID: "a", pinIDs: ["l:a"])
        flow.togglePin(subjectID: "a", pinIDs: ["l:a", "l:b"])
        #expect(flow.rows[0].pins == ["l:a", "l:b"])
    }

    @Test func evidenceForAnOldTargetIsIgnored() {
        var flow = loaded()
        flow.setTarget(subjectID: "a", token: "handle:e2")
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true)]))
        #expect(flow.rows[0].board == nil)
        #expect(flow.rows[0].pins.isEmpty)
    }

    @Test func evidenceNeighborsAreTheRowsDoneFiles() {
        let anchor = row(id: "anchor")
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a"), row(id: "h", handleID: "e7"),
                row(id: "n", target: .unplaced, handleID: "", assessment: "strong"),
                row(id: "s", target: .unplaced, handleID: "", assessment: "strong"),
                row(id: "u", target: .unplaced, handleID: "", assessment: "none"),
                anchor,
            ]),
            subjects: [fact("a"), fact("h"), fact("n"), fact("s"), fact("u"), fact("anchor", anchor: true)],
            bridges: []
        )
        flow.setTarget(subjectID: "n", token: "new")
        flow.setTarget(subjectID: "s", token: "skip")
        let unset = flow.rows.first { $0.subjectID == "u" }
        #expect(unset?.target == .unset)
        let got = flow.evidenceNeighbors(for: "a")
        #expect(got.map(\.subjectID).sorted() == ["h", "n", "s"])
        #expect(got.first { $0.subjectID == "h" }?.handleID == "e7")
        #expect(got.first { $0.subjectID == "n" }?.target == "new")
        #expect(got.first { $0.subjectID == "s" }?.target == "skip")
        flow.mappedRest = false
        #expect(flow.evidenceNeighbors(for: "a").isEmpty)
    }

    @Test func aRetargetClearsEvidenceAndPins() {
        var flow = loaded()
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true)]))
        #expect(flow.rows[0].pins == ["l:1"])

        flow.setTarget(subjectID: "a", token: "handle:e2")
        #expect(flow.rows[0].pins.isEmpty)
        #expect(flow.rows[0].board == nil)
        let line = flow.batchRows()[0]
        #expect(line.defaultPins)
        #expect(line.pinIDs.isEmpty)
    }

    @Test func everyRowStartsWithAnEmptyArgument() {
        let flow = loaded(ids: ["a", "b"])
        #expect(flow.rows.map(\.argument) == ["", ""])
        #expect(flow.batchRows().map(\.argument) == ["", ""])
    }

    @Test func aTypedArgumentSurvivesARetargetAndAReproposal() {
        var flow = loaded()
        flow.setArgument(subjectID: "a", argument: "my reasons")
        flow.setTarget(subjectID: "a", token: "new")
        #expect(flow.rows[0].argument == "my reasons")
        flow.merge(proposal(rows: [row(id: "a", handleID: "e2")], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].argument == "my reasons")

        // A row nobody decided keeps it across a reproposal too.
        var other = loaded(ids: ["a", "b"])
        other.setArgument(subjectID: "b", argument: "typed")
        other.merge(proposal(rows: [row(id: "a"), row(id: "b", handleID: "e2")], revision: 4), subjects: [fact("a"), fact("b")])
        #expect(other.rows[1].argument == "typed")
    }

    @Test func aManualPinOnTheSameTargetSurvivesAReproposal() {
        var flow = loaded()
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true)]))
        flow.togglePin(subjectID: "a", pinIDs: ["l:1"])
        #expect(flow.rows[0].pins.isEmpty)
        flow.merge(proposal(rows: [row(id: "a")], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].pins.isEmpty)
        #expect(flow.rows[0].board != nil)
    }

    @Test func anUnsettledRowKeepsItsEvidenceOnlyOnTheSameTarget() {
        var flow = loaded()
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true)]))
        flow.merge(proposal(rows: [row(id: "a")], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].pins == ["l:1"])
        flow.merge(proposal(rows: [row(id: "a", handleID: "e2")], revision: 5), subjects: [fact("a")])
        #expect(flow.rows[0].board == nil)
        #expect(flow.rows[0].pins.isEmpty)
    }

    @Test func duplicateNotesNameTheOtherRowAndItsRef() {
        var lost = row(id: "b", target: .unplaced, handleID: "", assessment: "strong")
        lost.possibleDuplicate = true
        lost.duplicateOfSubjectID = "a"
        var newA = row(id: "c", target: .unplaced, handleID: "", assessment: "strong")
        newA.possibleDuplicate = true
        newA.duplicateOfSubjectID = "d"
        var newB = row(id: "d", target: .unplaced, handleID: "", assessment: "strong")
        newB.possibleDuplicate = true
        newB.duplicateOfSubjectID = "c"
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), lost, newA, newB, row(id: "e", handleID: "e2")]),
            subjects: [fact("a"), fact("b"), fact("c"), fact("d"), fact("e")],
            bridges: []
        )
        flow.setTarget(subjectID: "c", token: "new")
        flow.setTarget(subjectID: "d", token: "new")
        let note = { (id: String) in flow.rows.first { $0.subjectID == id }?.duplicateNote }
        #expect(note("b") == .sharesHandle(otherName: "a", ref: "PER-1"))
        #expect(note("c") == .alsoNew(otherName: "d"))
        #expect(note("d") == .alsoNew(otherName: "c"))

        flow.mapRest()
        flow.setTarget(subjectID: "e", token: "handle:e2")
        flow.setTarget(subjectID: "a", token: "handle:e2")
        #expect(note("a") == .sharesHandle(otherName: "e", ref: "PER-2"))
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

    // MARK: - A record Promote did not offer

    /// A row Promote found no record for: no target, no band, no alternatives, so it opens unset.
    private func unmatched(_ id: String) -> CatalogPromoteGraphAlignmentRow {
        var unmatched = row(id: id, target: .unplaced, handleID: "", assessment: "none")
        unmatched.reason = .unplaced
        unmatched.alternatives = []
        return unmatched
    }

    @Test func aFoundRecordIsHeldAndJoinsTheMenu() {
        let polly = CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e7", ref: "PER-7", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "Polly Jones", parts: [
                CatalogNameValuePart(value: "Polly", type: "given"),
                CatalogNameValuePart(value: "Jones", type: "surname"),
            ]),
            nameValueCount: 1
        )
        let orders = NameFormatOrders.givenSurname
        func started() -> PromoteFlow {
            var flow = PromoteFlow(entryID: "a")
            flow.nameOrders = orders
            flow.load(entryID: "a", proposal: proposal(rows: [unmatched("a")]), subjects: [fact("a")], bridges: [])
            flow.mapRest()
            return flow
        }

        var flow = started()
        #expect(flow.rows[0].target == .unset)
        #expect(flow.rows[0].menu.isEmpty)
        let accepted = flow.hold(subjectID: "a", id: "e7", ref: "PER-7", header: .person(polly))
        #expect(accepted)
        let held = flow.rows[0]
        #expect(held.target == .handle(id: "e7", ref: "PER-7", title: "Jones, Polly"))
        #expect(held.menu.last == PromoteFlow.Alternative(
            id: "e7", ref: "PER-7", title: "Jones, Polly", readingTitle: "Polly Jones",
            subtitle: PersonLifeDisplay.line(polly).text, assessment: .none, reason: .decided
        ))
        #expect(held.decided)
        #expect(flow.manual)
        #expect(held.readingTitle == "Polly Jones")
        #expect(held.selectionReadout == PromoteFlow.SelectionReadout(assessment: .none, reason: .decided, chosen: true))

        let fixed = flow.fixed(for: flow.rows.filter(\.decided))
        #expect(fixed == [CatalogPromoteGraphAlignmentFixed(subjectID: "a", handleID: "e7")])
        let batch = flow.batchRows().first { $0.subjectID == "a" }
        #expect(batch?.target == "handle")
        #expect(batch?.entityID == "e7")
        #expect(batch?.defaultPins == true)

        var bare = started()
        let bareAccepted = bare.hold(subjectID: "a", id: "e7", ref: "PER-7", header: nil)
        #expect(bareAccepted)
        let item = bare.rows[0].menu.last
        #expect(item?.title == "PER-7")
        #expect(item?.readingTitle == "PER-7")
        #expect(item?.subtitle == "")
    }

    @Test func aFoundRecordSurvivesAReproposalWithItsBandRefreshed() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [unmatched("a"), row(id: "b")]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        flow.mapRest()
        flow.hold(subjectID: "a", id: "e7", ref: "PER-7", header: nil)

        var heldA = row(id: "a", handleID: "e7", assessment: "weak")
        heldA.handleRef = "PER-7"
        heldA.reason = .decided
        heldA.contested = true
        flow.merge(proposal(rows: [heldA, row(id: "b")], revision: 4), subjects: [fact("a"), fact("b")])
        var a = flow.rows.first { $0.subjectID == "a" }
        #expect(a?.menu.contains { $0.id == "e7" } == true)
        #expect(a?.menu.first { $0.id == "e7" }?.assessment == .weak)
        #expect(a?.menu.first { $0.id == "e7" }?.reason == .decided)
        #expect(a?.contested == true)
        #expect(a?.target.handleRef == "PER-7")
        if case .handle(let id, _, _)? = a?.target { #expect(id == "e7") } else { Issue.record("target is not a handle") }

        flow.setTarget(subjectID: "b", token: "handle:e2")
        flow.merge(
            proposal(rows: [unmatched("a"), row(id: "b")], revision: 5),
            subjects: [fact("a"), fact("b")]
        )
        a = flow.rows.first { $0.subjectID == "a" }
        if case .handle(let id, _, _)? = a?.target { #expect(id == "e7") } else { Issue.record("target is not a handle") }
        #expect(a?.menu.contains { $0.id == "e7" } == true)
    }

    @Test func aFoundRecordAlreadyInTheMenuIsPickedNotAddedAgain() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [row(id: "a")]), subjects: [fact("a")], bridges: [])
        let offered = flow.rows[0].menu
        #expect(offered.map(\.id) == ["e1", "e2"])
        let accepted = flow.hold(subjectID: "a", id: "e2", ref: "PER-2", header: nil)
        #expect(accepted)
        #expect(flow.rows[0].menu == offered)
        #expect(flow.rows[0].menu[1].reason == offered[1].reason)
        #expect(flow.rows[0].target == .handle(id: "e2", ref: "PER-2", title: offered[1].title))
    }

    @Test func holdingAFoundRecordResetsTheOldTargetsEvidence() {
        var contested = row(id: "a")
        contested.contested = true
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [contested]), subjects: [fact("a")], bridges: [])
        flow.setEvidence(subjectID: "a", handleID: "e1", board: evidence([("l:1", true)]))
        #expect(flow.rows[0].board != nil)
        #expect(flow.rows[0].pins == ["l:1"])
        #expect(flow.rows[0].contested)

        flow.hold(subjectID: "a", id: "e7", ref: "PER-7", header: nil)
        #expect(flow.rows[0].board == nil)
        #expect(flow.rows[0].pins.isEmpty)
        #expect(flow.rows[0].contested == false)
        #expect(flow.rows[0].move == nil)
    }

    @Test func aFoundRecordSharingAHandleNamesTheOtherRow() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [unmatched("a"), row(id: "b")]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        flow.mapRest()
        flow.hold(subjectID: "a", id: "e1", ref: "PER-1", header: nil)
        #expect(flow.rows.first { $0.subjectID == "a" }?.duplicateNote == .sharesHandle(otherName: "b", ref: "PER-1"))
    }

    @Test func anAnchorOrMissingRowOrEmptyIDHoldsNothing() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), row(id: "b")]),
            subjects: [fact("a", anchor: true), fact("b")],
            bridges: []
        )
        let before = flow.rows
        let onAnchor = flow.hold(subjectID: "a", id: "e7", ref: "PER-7", header: nil)
        let onMissing = flow.hold(subjectID: "zz", id: "e7", ref: "PER-7", header: nil)
        let withNoID = flow.hold(subjectID: "b", id: "", ref: "PER-7", header: nil)
        #expect(!onAnchor && !onMissing && !withNoID)
        #expect(flow.rows == before)
    }
}
