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
        handleID: String = "e1"
    ) -> CatalogPromoteGraphAlignmentRow {
        CatalogPromoteGraphAlignmentRow(
            subjectID: id,
            kind: "person",
            target: target,
            handleID: handleID,
            handleRef: "PER-1",
            alternatives: [
                CatalogPromoteGraphAlignmentAlternative(handleID: "e2", handleRef: "PER-2"),
            ],
            conflictWithFixed: false,
            possibleDuplicate: false
        )
    }

    private func fact(_ id: String, anchor: Bool = false, handleID: String = "e-anchor") -> PromoteFlow.SubjectFact {
        PromoteFlow.SubjectFact(
            id: id, kind: .person, name: id, ref: "CPR-\(id)",
            anchor: anchor ? PromoteFlow.Anchor(id: handleID, ref: "PER-A", title: "Promoted") : nil
        )
    }

    @Test func aRowCarriesItsEventsDate() {
        for anchored in [true, false] {
            var flow = PromoteFlow(entryID: "m")
            flow.load(
                entryID: "m",
                proposal: proposal(rows: [row(id: "m")]),
                subjects: [
                    PromoteFlow.SubjectFact(
                        id: "m", kind: .event, name: "Marriage", ref: "EVT-M",
                        anchor: anchored ? PromoteFlow.Anchor(id: "e-anchor", ref: "EVT-A", title: "Promoted") : nil,
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
        var event = row(id: "e")
        event.kind = "event"
        event.handleRef = "EVT-D1JBR"
        event.header = .event(CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-D1JBR", subjectTypeID: "t", label: ""),
            eventName: "",
            title: birth
        ))
        event.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(
                handleID: "e2", handleRef: "EVT-2",
                header: .event(CatalogEventHeader(
                    entity: CatalogCanonicalEntity(id: "e2", ref: "EVT-2", subjectTypeID: "t", label: ""),
                    title: CatalogEventTitle(rule: .label, label: "Working title", ref: "EVT-2")
                ))
            ),
        ]

        var person = row(id: "p")
        person.header = .person(CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: "Grandpa"),
            name: nil,
            nameValueCount: 0
        ))

        var place = row(id: "l")
        place.kind = "place"
        place.handleRef = "PLC-1"
        place.header = .place(CatalogPlaceHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PLC-1", subjectTypeID: "t", label: "The farm"),
            names: [],
            parents: ["Ontario", "Canada"]
        ))

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

    @Test func aPersonRecordReadsSortedInTheMenu() {
        var person = row(id: "p")
        person.header = .person(CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: ""),
            name: CatalogNameValue(form: "John Smith", parts: [
                CatalogNameValuePart(value: "John", type: "given"),
                CatalogNameValuePart(value: "Smith", type: "surname"),
            ]),
            nameValueCount: 1
        ))
        var flow = PromoteFlow(entryID: "e")
        flow.nameOrders = .givenSurname
        flow.load(entryID: "e", proposal: proposal(rows: [person]), subjects: [fact("p")], bridges: [])
        let personRow = flow.rows.first
        #expect(personRow?.menu.first?.title == "Smith, John")
    }

    /// Where a row starts is Go's call (`graphalign.Row.Start`): the page reads `target` and `handleID`
    /// and starts on that record when it heads the menu, else leaves the choice open, so Skip and New
    /// are the researcher's and a record another row took is never drafted twice.
    @Test func aRowStartsWhereTheProposalSays() {
        func load(_ r: CatalogPromoteGraphAlignmentRow) -> PromoteFlow {
            var flow = PromoteFlow(entryID: "a")
            flow.load(entryID: "a", proposal: proposal(rows: [r]), subjects: [fact("a")], bridges: [])
            return flow
        }
        var placed = row(id: "a", handleID: "e9")
        placed.handleRef = "PER-9"
        placed.alternatives = [
            CatalogPromoteGraphAlignmentAlternative(handleID: "e9", handleRef: "PER-9"),
            CatalogPromoteGraphAlignmentAlternative(handleID: "e2", handleRef: "PER-2"),
        ]
        let started = load(placed)
        #expect(started.rows.first?.target == .handle(id: "e9", ref: "PER-9", title: "PER-9"))
        #expect(started.rows.first?.menu.map(\.id) == ["e9", "e2"])
        #expect(started.targetsChosen)

        var unplaced = row(id: "a", target: .unplaced, handleID: "")
        unplaced.alternatives = [CatalogPromoteGraphAlignmentAlternative(handleID: "e2", handleRef: "PER-2")]
        let open = load(unplaced)
        #expect(open.rows.first?.target == .unset)
        #expect(open.targetsChosen == false)

        let new = load(row(id: "a", target: .new, handleID: ""))
        #expect(new.rows.first?.target == .unset)
        #expect(new.targetsChosen == false)
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
        var lost = row(id: "b", target: .unplaced, handleID: "")
        lost.alternatives = []
        flow.merge(proposal(rows: [row(id: "a"), lost], revision: 4), subjects: [fact("a"), fact("b")])
        let bea = flow.rows.first { $0.subjectID == "b" }
        #expect(bea?.target == .unset)
        #expect(bea?.updated == true)
    }

    @Test func anUnmatchedRowStartsUnselected() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a", target: .unplaced, handleID: ""),
                row(id: "b"),
            ]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        #expect(flow.rows.first { $0.subjectID == "a" }?.target == PromoteFlow.Target.unset)
        #expect(flow.targetsChosen == false)
        flow.mapRest()
        #expect(flow.batchRows().map(\.subjectID) == ["b"])

        flow.setTarget(subjectID: "a", token: "new")
        #expect(flow.rows.first { $0.subjectID == "a" }?.target == .newKind)
        #expect(flow.targetsChosen)
        #expect(flow.batchRows().first { $0.subjectID == "a" }?.target == .new)

        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.rows.first { $0.subjectID == "a" }?.target == .skip)
        #expect(flow.batchRows().first { $0.subjectID == "a" }?.target == .skip)
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

    @Test func retargetingClearsTheCaptionUntilTheNextProposal() {
        var a = row(id: "a")
        a.moveReason = .neighborMoved
        a.moveNeighborSubjectID = "b"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a, row(id: "b", handleID: "e3")]), subjects: [fact("a"), fact("b")], bridges: [])
        flow.mapRest()
        flow.setTarget(subjectID: "a", token: "handle:e2")
        let cleared = flow.rows.first { $0.subjectID == "a" }
        #expect(cleared?.caption == nil)
        #expect(cleared?.updated == false)
        var next = row(id: "a", handleID: "e2")
        next.moveReason = .neighborMoved
        next.moveNeighborSubjectID = "b"
        flow.merge(proposal(rows: [next], revision: 4), subjects: [fact("a")])
        let again = flow.rows.first { $0.subjectID == "a" }
        #expect(again?.target.handleID == "e2")
        #expect(again?.caption == .neighborMoved(neighborID: "b"))
    }

    @Test func theCaptionIsTheMoveElseTheViaNeighbor() {
        func caption(
            move: PromoteMoveReason? = nil, moveNeighbor: String = "", via: String = ""
        ) -> PromoteFlow.Reason? {
            var r = row(id: "a")
            r.moveReason = move
            r.moveNeighborSubjectID = moveNeighbor
            r.viaNeighborSubjectID = via
            var flow = PromoteFlow(entryID: "a")
            flow.load(entryID: "a", proposal: proposal(rows: [r]), subjects: [fact("a")], bridges: [])
            return flow.rows.first?.caption
        }
        #expect(caption(move: .neighborMoved, moveNeighbor: "n", via: "a") == .neighborMoved(neighborID: "n"))
        #expect(caption(via: "a") == .via(neighborID: "a"))
        #expect(caption() == nil)
        #expect(caption(move: .neighborPlaced, moveNeighbor: "") == nil)
        #expect(caption(move: .unrecognized("later_move"), moveNeighbor: "n") == nil)

        var a = row(id: "a")
        a.moveReason = .neighborMoved
        a.moveNeighborSubjectID = "b"
        var flow = PromoteFlow(entryID: "a")
        flow.load(entryID: "a", proposal: proposal(rows: [a, row(id: "b", handleID: "e3")]), subjects: [fact("a"), fact("b")], bridges: [])
        let rowB = flow.rows.first { $0.subjectID == "b" }
        #expect(rowB?.caption == nil)
        #expect(flow.rows.first { $0.subjectID == "a" }?.shownRef == "PER-1")
        flow.setTarget(subjectID: "a", token: "skip")
        #expect(flow.rows.first { $0.subjectID == "a" }?.shownRef == "CPR-a")
    }

    @Test func reasonCodesParseThroughTheWireVocabulary() {
        let wire: [(String, PromoteMoveReason)] = [
            ("neighbor_placed", .neighborPlaced),
            ("neighbor_moved", .neighborMoved), ("neighbor_released", .neighborReleased),
        ]
        for (code, expected) in wire {
            #expect(PromoteMoveReason(wire: code) == expected)
        }
        #expect(PromoteMoveReason(wire: "later_code") == .unrecognized("later_code"))
        #expect(PromoteMoveReason(wire: "taken") == .unrecognized("taken"))
        let targets: [(String, PromoteTargetCode)] = [("", .unplaced), ("handle", .handle), ("new", .new), ("skip", .skip)]
        for (code, expected) in targets {
            #expect(PromoteTargetCode(wire: code) == expected)
        }
        #expect(PromoteTargetCode(wire: "later") == .unrecognized("later"))
        for (_, c) in targets {
            #expect(PromoteTargetCode(wire: c.wire) == c)
        }
    }

    @Test func connectionsFollowTheirEnds() {
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a", handleID: "e1"), row(id: "b", handleID: "e2")]),
            subjects: [fact("a"), fact("b")],
            bridges: [PromoteFlow.BridgeFact(id: "br", sentence: "link", endA: "a", endB: "b")]
        )
        #expect(flow.connections.first?.state == .promotes)
        flow.setTarget(subjectID: "b", token: "skip")
        #expect(flow.connections.first?.state == .waits(endID: "b", why: .skipped))
        let batch = flow.batchRows()
        #expect(batch.count == 2)
        #expect(batch[0].subjectID == "a")
        #expect(batch[1].subjectID == "b")
        #expect(batch[1].target == .skip)
        flow.setTarget(subjectID: "b", token: "new")
        #expect(flow.connections.first?.state == .promotes)
    }

    private func link(
        _ id: String, _ a: String?, _ b: String?,
        sentence: String? = nil, alreadyPromoted: Bool = false
    ) -> PromoteFlow.BridgeFact {
        var bridge = PromoteFlow.BridgeFact(id: id, sentence: sentence ?? id, endA: a, endB: b)
        bridge.endAName = a ?? ""
        bridge.endBName = b ?? ""
        bridge.alreadyPromoted = alreadyPromoted
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
        // (a) The event is an anchor: it shows under the person only, with a note that it is promoted.
        let promoted = mapped(ids: ["j", "b"], anchors: ["b"], bridges: [link("pa", "j", "b")])
        let promotedRows = promoted.connectionsByRow
        #expect(promotedRows.keys.sorted() == ["j"])
        #expect(promotedRows["j"]?.first?.otherEnd == PromoteFlow.OtherEnd(name: "b", promoted: true))
        #expect(promotedRows["j"]?.first?.bridge.state == .promotes)

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
        #expect(shown["j"]?.first?.otherEnd == PromoteFlow.OtherEnd(name: "b", promoted: false))

        // (c) Both ends anchors: shown nowhere, still counted.
        let both = mapped(ids: ["a", "b"], anchors: ["a", "b"], bridges: [link("x", "a", "b")])
        #expect(both.connectionsByRow.isEmpty)
        #expect(both.connectionCount == 1)

        // (d) Already promoted: neither shown nor counted.
        let old = mapped(ids: ["a", "b"], bridges: [link("old", "a", "b", alreadyPromoted: true)])
        #expect(old.connectionsByRow.isEmpty)
        #expect(old.connections.isEmpty)
        #expect(old.connectionCount == 0)

        // (e) Neither end is a row: shown nowhere, nothing promotes.
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
            proposal: proposal(rows: [row(id: "a", handleID: "e1"), row(id: "b", handleID: "")]),
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
                row(id: "b", handleID: ""),
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

    @Test func promoteAllStartsExpandedAndHidesPromotedBridges() {
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a"), row(id: "b")]),
            subjects: [fact("a", anchor: true, handleID: "e-a"), fact("b", anchor: true, handleID: "e-b")],
            bridges: [
                PromoteFlow.BridgeFact(id: "old", sentence: "old", endA: "a", endB: "b", alreadyPromoted: true),
                PromoteFlow.BridgeFact(id: "new", sentence: "new", endA: "a", endB: "b"),
            ]
        )
        #expect(flow.visibleRows.map(\.subjectID) == ["a", "b"])
        #expect(flow.targetsChosen)
        #expect(flow.connections.map(\.id) == ["new"])
        #expect(flow.connections.first?.state == .promotes)
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
        let unset = row(id: "a", target: .unplaced, handleID: "")
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
        flow.setTarget(subjectID: "missing", token: "skip")
        #expect(flow.manual == false)
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [row(id: "a")]),
            subjects: [fact("a")],
            bridges: []
        )
        #expect(flow.requestLeave(.back) == false)
        flow.setTarget(subjectID: "a", token: "handle:e2")
        #expect(flow.requestLeave(.back) == true)
    }

    @Test func doneSendsATargetPerPromotedRow() {
        var flow = PromoteFlow(entryID: "a")
        flow.mappedRest = true
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [
                row(id: "a"),
                row(id: "b", target: .unplaced, handleID: ""),
                row(id: "c"),
                row(id: "d", target: .unplaced, handleID: ""),
            ]),
            subjects: [fact("a"), fact("b"), fact("c"), fact("d")],
            bridges: []
        )
        flow.setTarget(subjectID: "a", token: "handle:e1")
        flow.setTarget(subjectID: "b", token: "new")
        flow.setTarget(subjectID: "c", token: "skip")
        #expect(flow.batchRows() == [
            CatalogPromoteGraphAlignmentFixed(subjectID: "a", handleID: "e1", target: .handle),
            CatalogPromoteGraphAlignmentFixed(subjectID: "b", handleID: "", target: .new),
            CatalogPromoteGraphAlignmentFixed(subjectID: "c", handleID: "", target: .skip),
        ])
    }

    @Test func duplicateNotesNameTheOtherRowAndItsRef() {
        var lost = row(id: "b", target: .unplaced, handleID: "")
        lost.possibleDuplicate = true
        lost.duplicateOfSubjectID = "a"
        var newA = row(id: "c", target: .unplaced, handleID: "")
        newA.possibleDuplicate = true
        newA.duplicateOfSubjectID = "d"
        var newB = row(id: "d", target: .unplaced, handleID: "")
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
        held.alternatives = [CatalogPromoteGraphAlignmentAlternative(handleID: "e1", handleRef: "PER-1")]
        flow.merge(proposal(rows: [held], revision: 4), subjects: [fact("a")])
        #expect(flow.rows[0].conflictNote == "PER-1")
    }

    // MARK: - A record Promote did not offer

    /// A row Promote found no record for: no target, no band, no alternatives, so it opens unset.
    private func unmatched(_ id: String) -> CatalogPromoteGraphAlignmentRow {
        var unmatched = row(id: id, target: .unplaced, handleID: "")
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
            id: "e7", ref: "PER-7", title: "Jones, Polly",
            subtitle: PersonLifeDisplay.line(polly).text
        ))
        #expect(held.decided)
        #expect(flow.manual)

        let fixed = flow.fixed(for: flow.rows.filter(\.decided))
        #expect(fixed == [CatalogPromoteGraphAlignmentFixed(subjectID: "a", handleID: "e7")])
        let batch = flow.batchRows().first { $0.subjectID == "a" }
        #expect(batch?.target == .handle)
        #expect(batch?.handleID == "e7")

        var bare = started()
        let bareAccepted = bare.hold(subjectID: "a", id: "e7", ref: "PER-7", header: nil)
        #expect(bareAccepted)
        let item = bare.rows[0].menu.last
        #expect(item?.title == "PER-7")
        #expect(item?.subtitle == "")
    }

    @Test func aFoundRecordSurvivesAReproposal() {
        var flow = PromoteFlow(entryID: "a")
        flow.load(
            entryID: "a",
            proposal: proposal(rows: [unmatched("a"), row(id: "b")]),
            subjects: [fact("a"), fact("b")],
            bridges: []
        )
        flow.mapRest()
        flow.hold(subjectID: "a", id: "e7", ref: "PER-7", header: nil)

        var heldA = row(id: "a", handleID: "e7")
        heldA.handleRef = "PER-7"
        flow.merge(proposal(rows: [heldA, row(id: "b")], revision: 4), subjects: [fact("a"), fact("b")])
        var a = flow.rows.first { $0.subjectID == "a" }
        #expect(a?.menu.contains { $0.id == "e7" } == true)
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
        #expect(flow.rows[0].target == .handle(id: "e2", ref: "PER-2", title: offered[1].title))
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
