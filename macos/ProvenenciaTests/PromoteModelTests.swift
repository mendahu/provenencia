import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct PromoteModelTests {
    private let projectDir = "/tmp/promote-model.provenencia"
    private let sourceID = "src-1"

    private func waitUntil(_ condition: () -> Bool) async -> Bool {
        for _ in 0..<200 {
            if condition() { return true }
            try? await Task.sleep(nanoseconds: 10_000_000)
        }
        return condition()
    }

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
        store.subjectsBySource[sourceID] = subjects
        for (index, subject) in subjects.enumerated() {
            store.subjectPositionsBySubject[subject.id] = CatalogSubjectPosition(
                subjectID: subject.id, gridX: 0, gridY: Int64(index * 2)
            )
        }
        return store
    }

    private func row(_ id: String, handleID: String, assessment: String = "strong") -> CatalogPromoteGraphAlignmentRow {
        CatalogPromoteGraphAlignmentRow(
            subjectID: id, kind: "person", target: "handle", handleID: handleID, handleRef: "PER-\(handleID)",
            score: 5, assessment: assessment, reason: "agrees", reasonPropertyKey: "name", reasonPropertyOrigin: "provenencia",
            comparisons: [
                CatalogPromoteGraphAlignmentComparison(
                    propertyKey: "name", propertyOrigin: "provenencia", outcome: "agree", valueType: "name",
                    pinned: true, incomingObservationID: "obs-\(id)", incomingDisplay: id,
                    memberObservationID: "mem-\(id)", memberDisplay: id
                ),
            ],
            alternatives: [
                CatalogPromoteGraphAlignmentAlternative(handleID: "other", handleRef: "PER-OTHER", score: 1),
            ],
            conflictWithFixed: false, possibleDuplicate: false
        )
    }

    private func model(store: FakeStore, subjectID: String) -> PromoteModel {
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let entry = PromoteEntry(location: .promote(
            sourceId: sourceID, subjectId: subjectID, kind: .person, ref: "CPR-1", title: "James", sourceTitle: "Census"
        ))!
        let model = PromoteModel(entry: entry, session: session, store: store, userID: "user-1", catalogCounts: nil)
        let _: QueryHandle<SourceGraphRows> = session.query(model.graphKey)
        let _: QueryHandle<PropertiesSnapshot> = session.query(model.propertiesKey)
        let _: QueryHandle<[CatalogConnectRule]> = session.query(model.rulesKey)
        let _: QueryHandle<[CatalogClaimConfidenceGrade]> = session.query(model.confidenceKey)
        return model
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

    @Test func doneSendsThePinnedPair() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1")])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        model.done()
        let filed = await waitUntil { store.lastPromoteBatch != nil }
        #expect(filed)
        #expect(store.lastPromoteBatch?.rows.first?.pairs.first?.incomingObservationID == "obs-a")
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
            projectDir: projectDir, userID: "user-1", sourceID: sourceID,
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
            CatalogPromoteGraphAlignmentFixed(subjectID: "a", target: "new"),
            CatalogPromoteGraphAlignmentFixed(subjectID: "b", target: "skip"),
        ])
    }

    @Test func reasonsAreWordedWithLabelsAndRefs() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada"), subject("b", ref: "CPR-B", label: "Bea")]
        let store = store(subjects: subjects)
        store.propertiesByProject[projectDir] = [
            CatalogProperty(id: "p-name", key: "name", origin: "provenencia", label: "Full name", description: "", valueType: "name"),
        ]
        var viaA = row("b", handleID: "e2")
        viaA.reason = "via"
        viaA.viaNeighborSubjectID = "a"
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1"), viaA])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        let ada = model.flow.rows.first { $0.subjectID == "a" }!
        let bea = model.flow.rows.first { $0.subjectID == "b" }!
        #expect(model.reasonText(for: ada) == L10n.Promote.agreesOn(property: "Full name"))
        #expect(model.reasonText(for: bea) == L10n.Promote.viaNeighbor(neighbor: "Ada", ref: "PER-e1"))
    }

    @Test func leaveGuardAsksOnlyAfterAManualChange() async {
        let subjects = [subject("a", ref: "CPR-A", label: "Ada")]
        let store = store(subjects: subjects)
        store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("a", handleID: "e1")])]
        let model = model(store: store, subjectID: "a")
        await model.load()
        #expect(model.shouldHoldNavigation(.back) == false)
        model.setArgument(subjectID: "a", argument: "because the name agrees")
        #expect(model.shouldHoldNavigation(.back) == true)
    }

    @Test func aPinnedObservationDeleteNamesThePerson() async throws {
        let store = FakeStore()
        let person = CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "type-person", label: "")
        store.observationsBySource[sourceID] = [
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
