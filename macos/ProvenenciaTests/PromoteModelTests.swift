import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct PromoteModelTests {
    private let projectDir = "/tmp/promote-model.provenencia"
    private let sourceID = "src-1"
    private let graphID = "g-1"

    private func subject(_ id: String, ref: String, label: String) -> CatalogSubject {
        CatalogSubject(id: id, ref: ref, sourceID: sourceID, subjectTypeID: "type-person", label: label, description: "")
    }

    private func store(subjects: [CatalogSubject]) -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: graphID, ref: "GPH-1", sourceID: sourceID, name: "Default", description: "", artifactIDs: []
            ),
        ]
        store.subjectsByGraph[graphID] = subjects
        for (index, subject) in subjects.enumerated() {
            store.subjectPositionsBySubject[subject.id] = CatalogSubjectPosition(
                subjectID: subject.id, gridX: 0, gridY: Int64(index * 2)
            )
        }
        return store
    }

    private func row(_ id: String, handleID: String, assessment: String = "strong") -> CatalogPromoteGraphAlignmentRow {
        CatalogPromoteGraphAlignmentRow(
            subjectID: id, kind: "person", target: .handle, handleID: handleID, handleRef: "PER-\(handleID)",
            score: 5, assessment: assessment, reason: .agrees, reasonPropertyKey: "name", reasonPropertyOrigin: "provenencia",
            alternatives: [
                CatalogPromoteGraphAlignmentAlternative(handleID: "other", handleRef: "PER-OTHER", score: 1),
            ],
            conflictWithFixed: false, possibleDuplicate: false
        )
    }

    private func model(store: FakeStore, subjectID: String, graphID: String? = nil) -> PromoteModel {
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let entry = PromoteEntry(location: .promote(
            sourceId: sourceID, graphId: graphID ?? self.graphID, subjectId: subjectID, kind: .person, ref: "CPR-1", title: "James", sourceTitle: "Census"
        ))!
        let model = PromoteModel(entry: entry, session: session, store: store, userID: "user-1")
        let _: QueryHandle<EvidenceGraphRows> = session.query(model.graphKey)
        let _: QueryHandle<PropertiesSnapshot> = session.query(model.propertiesKey)
        let _: QueryHandle<[CatalogConnectRule]> = session.query(model.rulesKey)
        let _: QueryHandle<[CatalogClaimConfidenceGrade]> = session.query(model.confidenceKey)
        return model
    }

    @Test func aPageWithoutItsGraphSaysSoInsteadOfLookingEmpty() async {
        let store = store(subjects: [subject("a", ref: "CPR-A", label: "Ada")])
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1")])]
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let entry = PromoteEntry(location: .promote(
            sourceId: sourceID, graphId: graphID, subjectId: "a", kind: .person, ref: "CPR-A", title: "Ada", sourceTitle: "Census"
        ))!
        // The graph key was never warmed, so readyValue has nothing to give.
        let model = PromoteModel(entry: entry, session: session, store: store, userID: "user-1")
        await model.load()
        #expect(model.loadError == L10n.string(L10n.Promote.graphUnavailable))
        #expect(model.flow.rows.isEmpty)
        #expect(model.canFinish == false)
    }

    @Test func opensOnTheClickedSubjectThenMapsTheRest() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e1")])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        #expect(model.flow.visibleRows.map(\.subjectID) == ["a"])
        model.mapRest()
        #expect(model.flow.visibleRows.map(\.subjectID) == ["a", "b"])
    }

    @Test func reproposeKeepsBridges() async {
        let subjects = [
            subject("a", ref: "CPR-A", label: "Ada"),
            subject("b", ref: "CPR-B", label: "Bea"),
            CatalogSubject(id: "br", ref: "CPA-1", sourceID: sourceID, subjectTypeID: "type-relationship", label: "Ada and Bea", description: ""),
        ]
        let store = store(subjects: subjects)
        store.subjectTypesByProject[projectDir]?.append(CatalogSubjectType(
            id: "type-relationship", key: "relationship", origin: "provenencia", label: "Relationship",
            description: "", refPrefix: "REL", candidateRefPrefix: "CPA"
        ))
        func cite(_ key: String, _ subjectID: String) -> CatalogObservation {
            CatalogObservation(
                id: "obs-\(key)", ref: "OBS-\(key)", citationID: "c1", subjectID: "br", propertyID: "p-\(key)",
                polarity: "positive", valueText: "", valueInteger: nil, valueDateID: "", valueNameID: "",
                valueSubjectID: subjectID, valueTermID: "", propertyKey: key, propertyLabel: key, propertyValueType: "subject"
            )
        }
        store.observationsByGraph[graphID] = [cite("person", "a"), cite("related_to", "b")]
        store.promoteProposals = [
            CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e1")]),
            CatalogPromoteGraphAlignmentProposal(revision: 2, rows: [row("a", handleID: "e1"), row("b", handleID: "e9")]),
        ]
        let model = model(store: store, subjectID: "a")
        await model.load()
        let loaded = model.flow.bridges
        #expect(loaded.count == 1)
        // The graph changes under the page, then a decision proposes again: the page keeps its connections.
        store.observationsByGraph[graphID] = []
        model.session.invalidate(model.graphKey)
        let _: QueryHandle<EvidenceGraphRows> = model.session.query(model.graphKey)
        model.mapRest()
        model.setTarget(subjectID: "a", token: "handle:other")
        let proposed = await waitUntil { model.flow.rows.first { $0.subjectID == "b" }?.updated == true }
        #expect(proposed)
        #expect(model.flow.bridges == loaded)
    }

    @Test func retargetKeepsTheDecidedRowAndMarksTheSuggestionUpdated() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        store.promoteProposals = [
            CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e1")]),
            CatalogPromoteGraphAlignmentProposal(revision: 2, rows: [row("a", handleID: "e1"), row("b", handleID: "e9", assessment: "strong")]),
        ]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        model.setTarget(subjectID: "a", token: "handle:other")
        let moved = await waitUntil { model.flow.rows.first { $0.subjectID == "b" }?.updated == true }
        #expect(moved)
        let ada = model.flow.rows.first { $0.subjectID == "a" }
        let bea = model.flow.rows.first { $0.subjectID == "b" }
        #expect(ada?.target.handleID == "other")
        #expect(bea?.decided == false)
    }

    @Test func promoteSendsItsGraph() async {
        let store = store(subjects: [subject("a", ref: "CPR-A", label: "Ada")])
        store.evidenceGraphsByProject[projectDir]?.append(CatalogEvidenceGraph(
            id: "g-2", ref: "GPH-2", sourceID: sourceID, name: "Jones", description: "", artifactIDs: []
        ))
        store.subjectsByGraph["g-2"] = [subject("z", ref: "CPR-Z", label: "Zed")]
        store.subjectPositionsBySubject["z"] = CatalogSubjectPosition(subjectID: "z", gridX: 0, gridY: 0)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("z", handleID: "e1")])]
        let model = model(store: store, subjectID: "z", graphID: "g-2")
        await model.load()
        model.done()
        let filed = await waitUntil { store.lastPromoteBatch != nil }
        #expect(filed)
        #expect(store.recordedCalls.contains("proposePromoteGraphAlignment graph=g-2"))
        #expect(store.recordedCalls.contains("applyPromoteGraphAlignment graph=g-2 rows=1"))
        #expect((store.lastPromoteBatch ?? []).map(\.subjectID) == ["z"])
    }

    // MARK: - T-12

    @Test func doneFilesEveryRowWithATargetOnly() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e2")])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        model.done()
        let filed = await waitUntil { store.lastPromoteBatch != nil }
        #expect(filed)
        let rows = store.lastPromoteBatch ?? []
        #expect(rows.map(\.subjectID) == ["a", "b"])
        #expect(rows.map(\.entityID) == ["e1", "e2"])
        #expect(store.recordedCalls.contains { $0.hasPrefix("getPromoteClaimEvidence") } == false)
        #expect(store.recordedCalls.contains { $0.hasPrefix("listClaimConfidenceGrades") } == false)
    }

    @Test func aStaleDoneReproposes() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada")]
        let store = store(subjects: subjects)
        store.promoteProposals = [
            CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1")]),
            CatalogPromoteGraphAlignmentProposal(revision: 2, rows: [row("a", handleID: "e9")]),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: sourceID, ref: "SRC-1", sourceTypeID: "", title: "Census", description: ""),
        ]
        let model = model(store: store, subjectID: "a")
        await model.load()
        _ = try? await store.createSubject(
            projectDir: projectDir, userID: "user-1", graphID: graphID,
            subjectTypeID: "type-person", label: "Later", description: "", placement: nil
        )
        model.done()
        let stale = await waitUntil { model.flow.staleNote != nil }
        #expect(stale)
        #expect(model.isSaving == false)
    }

    @Test func anOlderProposalAnsweringLastIsDropped() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        store.promoteProposals = [
            CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e1")]),
            CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "other"), row("b", handleID: "e-old")]),
            CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e-new")]),
        ]
        store.promoteProposeDelays = [.zero, .milliseconds(300), .zero]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        model.setTarget(subjectID: "a", token: "handle:other")
        // The first retarget has to claim its slow proposal before the next
        // call takes a slot. Otherwise the newer request can sleep on the
        // older answer and, being current, keep it.
        let claimed = await waitUntil {
            store.recordedCalls.filter { $0 == "proposePromoteGraphAlignment graph=g-1" }.count >= 2
        }
        #expect(claimed)
        model.setTarget(subjectID: "a", token: "handle:e1")
        let settled = await waitUntil { !model.isProposing }
        #expect(settled)
        try? await Task.sleep(for: .milliseconds(400))
        #expect(model.flow.rows.first { $0.subjectID == "b" }?.target.handleID == "e-new")
    }

    @Test func doneWaitsForAPendingProposal() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1")])]
        store.promoteProposeDelays = [.zero, .milliseconds(200)]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.setTarget(subjectID: "a", token: "handle:other")
        #expect(model.canFinish == false)
        model.done()
        #expect(model.isSaving == false)
        let settled = await waitUntil { model.canFinish }
        #expect(settled)
        #expect(store.lastPromoteBatch == nil)
    }

    @Test func newAndSkipDecisionsAreHeldOnTheNextProposal() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), row("b", handleID: "e2")])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        model.setTarget(subjectID: "a", token: "new")
        _ = await waitUntil { !model.isProposing }
        model.setTarget(subjectID: "b", token: "skip")
        _ = await waitUntil { !model.isProposing }
        let held = store.lastPromoteFixed.sorted { $0.subjectID < $1.subjectID }
        #expect(held == [
            CatalogPromoteGraphAlignmentFixed(subjectID: "a", target: .new),
            CatalogPromoteGraphAlignmentFixed(subjectID: "b", target: .skip),
        ])
    }

    @Test func aFoundRecordIsProposedHeldAndFiled() async {
        let store = storeWithFoundRecords([subject("a", ref: "CPR-A", label: "Ada")])
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        #expect(model.flow.row("a")?.target == .unset)

        model.hold(subjectID: "a", id: "e-far", ref: "PER-FAR", header: nil)
        _ = await waitUntil { !model.isProposing }
        #expect(store.lastPromoteFixed == [CatalogPromoteGraphAlignmentFixed(subjectID: "a", handleID: "e-far")])
        let ada = model.flow.row("a")
        #expect(ada?.target.handleID == "e-far")

        model.done()
        let filed = await waitUntil { store.lastPromoteBatch != nil }
        #expect(filed)
        let batch = store.lastPromoteBatch?.first { $0.subjectID == "a" }
        #expect(batch?.target == "handle")
        #expect(batch?.entityID == "e-far")
    }

    private func storeWithFoundRecords(_ subjects: [CatalogSubject]) -> FakeStore {
        let store = store(subjects: subjects)
        store.membershipBySubject["x"] = CatalogSubjectMembership(
            subjectID: "x", claimID: "claim-x",
            entity: CatalogCanonicalEntity(id: "e-far", ref: "PER-FAR", subjectTypeID: "type-person", label: "Polly Jones"),
            kind: "person"
        )
        store.membershipBySubject["y"] = CatalogSubjectMembership(
            subjectID: "y", claimID: "claim-y",
            entity: CatalogCanonicalEntity(id: "p-1", ref: "PLC-1", subjectTypeID: "type-place", label: "Polly Street"),
            kind: "place"
        )
        return store
    }

    @Test func aRecordSearchOpensOnTheRowsKindAndDecidesNothing() async {
        let store = storeWithFoundRecords([subject("a", ref: "CPR-A", label: "Ada")])
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        let fixedBefore = store.lastPromoteFixed

        model.setTarget(subjectID: "a", token: PromoteFlow.Target.findToken)
        #expect(model.recordSearch?.subjectID == "a")
        #expect(model.recordSearch?.search.kinds == ["person"])
        #expect(model.flow.row("a")?.decided == false)
        #expect(model.flow.row("a")?.target == .unset)
        #expect(model.isProposing == false)
        #expect(store.lastPromoteFixed == fixedBefore)
    }

    @Test func cancellingOrMovingARecordSearchChangesNoRow() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = storeWithFoundRecords(subjects)
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        let rowsBefore = model.flow.rows
        let fixedBefore = store.lastPromoteFixed

        model.openRecordSearch(subjectID: "a")
        model.openRecordSearch(subjectID: "b")
        #expect(model.recordSearch?.subjectID == "b")
        model.closeRecordSearch()
        #expect(model.recordSearch == nil)
        #expect(model.flow.rows == rowsBefore)
        #expect(store.lastPromoteFixed == fixedBefore)

        model.openRecordSearch(subjectID: "a")
        model.done()
        #expect(model.recordSearch == nil)
    }

    @Test func aFoundRecordIsSearchedPickedAndHeld() async {
        let store = storeWithFoundRecords([subject("a", ref: "CPR-A", label: "Ada")])
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()

        model.openRecordSearch(subjectID: "a")
        model.recordSearchQueryChanged("Polly")
        let search = model.recordSearch!.search
        await waitForSearch(search)
        #expect(search.hits.map(\.id) == ["e-far"])

        model.pickSearchHit(id: "e-far")
        #expect(model.recordSearch == nil)
        _ = await waitUntil { !model.isProposing }
        #expect(store.lastPromoteFixed == [CatalogPromoteGraphAlignmentFixed(subjectID: "a", handleID: "e-far")])
        #expect(model.flow.row("a")?.decided == true)
        #expect(model.flow.row("a")?.target.handleID == "e-far")

        model.done()
        let filed = await waitUntil { store.lastPromoteBatch != nil }
        #expect(filed)
        #expect(store.lastPromoteBatch?.first { $0.subjectID == "a" }?.entityID == "e-far")
    }

    @Test func aRecordSearchShowsFailuresAndIgnoresUnknownHits() async {
        struct Boom: Error {}
        let store = storeWithFoundRecords([subject("a", ref: "CPR-A", label: "Ada")])
        store.searchCatalogError = Boom()
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        let before = model.flow.row("a")

        model.openRecordSearch(subjectID: "a")
        model.recordSearchQueryChanged("Polly")
        let search = model.recordSearch!.search
        await waitForSearch(search)
        #expect(search.showsErrorState)

        model.pickSearchHit(id: "nope")
        #expect(model.recordSearch?.subjectID == "a")
        #expect(model.flow.row("a") == before)
    }

    /// Ada and Bea on one page, each placed by the given proposal rows.
    private func loadedAdaAndBea(
        _ ada: CatalogPromoteGraphAlignmentRow,
        _ bea: CatalogPromoteGraphAlignmentRow,
        properties: [CatalogProperty] = []
    ) async -> PromoteModel {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        if !properties.isEmpty { store.propertiesByProject[projectDir] = properties }
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [ada, bea])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        return model
    }

    @Test func aViaCaptionNamesTheNeighborByItsRecord() async {
        var viaA = row("b", handleID: "e2")
        viaA.viaNeighborSubjectID = "a"
        var viaZ = row("c", handleID: "e3")
        viaZ.viaNeighborSubjectID = "x"
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea"), subject("c", ref: "CPR-C", label: "Cy")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), viaA, viaZ])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.mapRest()
        let bea = model.flow.row("b")!
        let cy = model.flow.row("c")!
        #expect(model.reasonText(bea.caption!, on: bea) == L10n.Promote.viaNeighbor(neighbor: "Ada", ref: "PER-e1"))
        #expect(model.reasonText(cy.caption!, on: cy) == L10n.string(L10n.Promote.reasonVia))
        #expect(L10n.Promote.updatedDetail("x") == "Updated · x")
    }

    /// Ada and Bea on one page, joined by one bridge of `kind` that carries `term`.
    private func viaReason(
        kind: EvidenceBridgeKind,
        edgeKeys: (String, String),
        term: (property: String, key: String, origin: String, label: String)?
    ) async -> String {
        let bridgeSubject = CatalogSubject(
            id: "br", ref: "CPA-1", sourceID: sourceID, subjectTypeID: "type-bridge", label: "", description: ""
        )
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea"), bridgeSubject]
        let store = store(subjects: subjects)
        store.subjectTypesByProject[projectDir]?.append(CatalogSubjectType(
            id: "type-bridge", key: kind.rawValue, origin: "provenencia", label: "Bridge",
            description: "", refPrefix: "CPA", candidateRefPrefix: "CPA"
        ))
        func edge(_ key: String, _ subjectID: String) -> CatalogObservation {
            var o = CatalogObservation(
                id: "o-\(key)", ref: "OBS-\(key)", citationID: "c", subjectID: "br", propertyID: "p-\(key)",
                polarity: ObservationPolarity.positive.rawValue, valueText: "", valueInteger: nil, valueDateID: "",
                valueNameID: "", valueSubjectID: subjectID, valueTermID: "", propertyKey: key,
                propertyLabel: key, propertyValueType: PropertyValueType.subject.rawValue
            )
            o.valueSubjectID = subjectID
            return o
        }
        var observations = [edge(edgeKeys.0, "a"), edge(edgeKeys.1, "b")]
        if let term {
            var o = edge(term.property, "")
            o.id = "o-term"
            o.valueText = term.label
            o.valueTermID = "term-1"
            o.valueTermKey = term.key
            o.valueTermOrigin = term.origin
            o.propertyValueType = PropertyValueType.term.rawValue
            observations.append(o)
        }
        store.observationsByGraph[graphID] = observations
        var viaA = row("b", handleID: "e2")
        viaA.viaNeighborSubjectID = "a"
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), viaA])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        let bea = model.flow.row("b")!
        return model.reasonText(bea.caption!, on: bea)
    }

    @Test func aViaReasonNamesTheTermInLowercase() async {
        let provenencia = CatalogOrigin.provenencia
        #expect(await viaReason(
            kind: .participation, edgeKeys: ("person", "event"),
            term: ("role", "subject", provenencia, "Subject")
        ) == L10n.Promote.via(neighbor: "Ada", ref: "PER-e1", role: "subject"))
        #expect(await viaReason(
            kind: .relationship, edgeKeys: ("person", "related_to"),
            term: ("relationship_type", "spouse", provenencia, "Spouse")
        ) == L10n.Promote.via(neighbor: "Ada", ref: "PER-e1", role: "spouse"))
        #expect(await viaReason(
            kind: .placeRelationship, edgeKeys: ("from", "to"),
            term: ("place_relationship_type", "part_of", provenencia, "Part of")
        ) == L10n.Promote.via(neighbor: "Ada", ref: "PER-e1", role: "part of"))
        // A location has no term, so the reason names only the neighbor.
        #expect(await viaReason(
            kind: .location, edgeKeys: ("event", "place"), term: nil
        ) == L10n.Promote.viaNeighbor(neighbor: "Ada", ref: "PER-e1"))
        // A researcher's term keeps its own label, lowercased.
        #expect(await viaReason(
            kind: .participation, edgeKeys: ("person", "event"),
            term: ("role", "godparent", "user", "Godparent")
        ) == L10n.Promote.via(neighbor: "Ada", ref: "PER-e1", role: "godparent"))
    }

    @Test func movesAreWordedNamingTheNeighbor() async {
        let model = await loadedAdaAndBea(row("a", handleID: "e1"), row("b", handleID: "e2"))
        let bea = model.flow.row("b")!
        #expect(model.reasonText(.neighborPlaced(neighborID: "a"), on: bea) == L10n.Promote.neighborPlaced(neighbor: "Ada", ref: "PER-e1"))
        #expect(model.reasonText(.neighborMoved(neighborID: "a"), on: bea) == L10n.Promote.neighborMoved(neighbor: "Ada", ref: "PER-e1"))
        #expect(model.reasonText(.neighborReleased(neighborID: "a"), on: bea) == L10n.Promote.neighborReleased(neighbor: "Ada"))
        #expect(model.reasonText(.neighborMoved(neighborID: "zz"), on: bea) == L10n.string(L10n.Promote.reasonMoveUnknown))
        // A neighbor with no record to name is not named by its Source ref.
        model.setTarget(subjectID: "a", token: "skip")
        _ = await waitUntil { !model.isProposing }
        let unknown = L10n.string(L10n.Promote.reasonMoveUnknown)
        #expect(model.reasonText(.neighborPlaced(neighborID: "a"), on: bea) == unknown)
        #expect(model.reasonText(.neighborMoved(neighborID: "a"), on: bea) == unknown)
        #expect(model.reasonText(.neighborReleased(neighborID: "a"), on: bea) == L10n.Promote.neighborReleased(neighbor: "Ada"))
    }

    @Test func aPlacedEventsRowCarriesItsDate() async {
        let wedding = CatalogSubject(
            id: "m", ref: "CPR-M", sourceID: sourceID, subjectTypeID: "type-event", label: "Wedding", description: ""
        )
        let store = store(subjects: [subject("a", ref: "CPR-A", label: "Ada"), wedding])
        store.subjectTypesByProject[projectDir]?.append(CatalogSubjectType(
            id: "type-event", key: "event", origin: "provenencia", label: "Event",
            description: "", refPrefix: "EVT", candidateRefPrefix: "CEV"
        ))
        let married = CatalogDateValueInput(kind: "point", startYear: 1842, startMonth: 5, startDay: 12)
        store.observationsByGraph[graphID] = [
            CatalogObservation(
                id: "o-date", ref: "OBS-D", citationID: "c", subjectID: "m", propertyID: "p-date",
                polarity: "positive", valueText: "", valueInteger: nil, valueDateID: "", date: married,
                valueNameID: "", valueSubjectID: "", valueTermID: "",
                propertyKey: "date", propertyLabel: "Date", propertyValueType: "date"
            ),
        ]
        var eventRow = row("m", handleID: "e2")
        eventRow.kind = "event"
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), eventRow])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        #expect(model.flow.row("m")?.date == DateRowDisplay.line(date: married, includesPhrase: false))
        #expect(model.flow.row("m")?.date.isEmpty == false)
        #expect(model.flow.row("a")?.date == "")
    }

    @Test func leaveGuardAsksOnlyAfterAManualChange() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1")])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        #expect(model.shouldHoldNavigation(.back) == false)
        model.setTarget(subjectID: "a", token: "handle:other")
        #expect(model.shouldHoldNavigation(.back) == true)
    }

    @Test func aPinnedObservationDeleteNamesThePerson() async throws {
        let store = FakeStore()
        let person = CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "type-person", label: "")
        store.observationsByGraph[graphID] = [
            CatalogObservation(
                id: "obs-1", ref: "OBS-1", citationID: "c1", subjectID: "a", propertyID: "p-name",
                polarity: "positive", valueText: "Ada", valueInteger: nil, valueDateID: "",
                valueNameID: "n1", nameForm: "Ada", valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            ),
        ]
        store.pinsByObservation["obs-1"] = [person]
        let impact = try await store.getDeleteImpact(projectDir: projectDir, kind: "observation", id: "obs-1")
        #expect(impact.allowed)
        #expect(impact.cascades.first?.listed.map(\.ref) == ["PER-7KD45"])
        let copy = PVDeleteImpactCopy.confirmCopy(
            for: PVDeleteImpactTarget(kind: "observation", id: "obs-1", ref: "OBS-1", title: "Name"),
            report: impact
        )
        #expect(copy.message.contains(L10n.DeleteImpact.leavesEvidence(handleRefs: "PER-7KD45")))
    }
}
