import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct UnpromoteFlowTests {
    private let projectDir = ArtifactFileFixtures.uniqueProjectDir()
    private let sourceID = "src-1"
    private let graphID = "g-1"

    private func subject(_ id: String, _ prefix: String, _ typeID: String, _ label: String) -> CatalogSubject {
        CatalogSubject(id: id, ref: "\(prefix)-\(id)", sourceID: sourceID, subjectTypeID: typeID, label: label, description: "")
    }

    private func edge(_ id: String, bridge: String, key: String, value: String) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: "OBS-\(id)", citationID: "c-1", subjectID: bridge, propertyID: "p-\(key)",
            polarity: "positive", valueText: "", valueInteger: nil, valueDateID: "", valueNameID: "",
            valueSubjectID: value, valueTermID: "", propertyKey: key, propertyLabel: key,
            propertyValueType: "subject"
        )
    }

    private func type(_ id: String, _ key: String, _ label: String, _ ref: String, _ candidate: String) -> CatalogSubjectType {
        CatalogSubjectType(
            id: id, key: key, origin: "provenencia", label: label, description: "",
            refPrefix: ref, candidateRefPrefix: candidate
        )
    }

    /// A graph with J. Weber, John, a census, and two bridges: b1 (relationship J. Weber - John)
    /// and b2 (participation J. Weber in the census). Nothing is promoted.
    private func makeFixture() -> (FakeStore, WorkspaceSession) {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            type("t-person", "person", "Person", "PER", "CPR"),
            type("t-event", "event", "Event", "EVT", "CEV"),
            type("t-rel", "relationship", "Relationship", "REL", "CRL"),
            type("t-part", "participation", "Participation", "PTN", "CPA"),
        ]
        store.subjectsByGraph[graphID] = [
            subject("s-weber", "CPR", "t-person", "J. Weber"),
            subject("s-john", "CPR", "t-person", "John Weber"),
            subject("s-census", "CEV", "t-event", "1881 census"),
            subject("b1", "CRL", "t-rel", ""),
            subject("b2", "CPA", "t-part", ""),
        ]
        for (index, subject) in store.subjectsByGraph[graphID, default: []].enumerated() {
            store.subjectPositionsBySubject[subject.id] = CatalogSubjectPosition(
                subjectID: subject.id, gridX: Int64(index * 3), gridY: 0
            )
        }
        store.observationsByGraph[graphID] = [
            edge("e1", bridge: "b1", key: "person", value: "s-weber"),
            edge("e2", bridge: "b1", key: "related_to", value: "s-john"),
            edge("e3", bridge: "b2", key: "person", value: "s-weber"),
            edge("e4", bridge: "b2", key: "event", value: "s-census"),
        ]
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        return (store, session)
    }

    private func sentence(_ session: WorkspaceSession, _ bridgeID: String) async throws -> String {
        let snapshot = try #require(await session.evidenceGraphSnapshot(graphID: graphID))
        let bridge = try #require(snapshot.bridges.first { $0.id == bridgeID })
        return EvidenceBridgeEdgeSummary.sentence(for: bridge, in: snapshot, orders: session.nameFormat)
    }

    private func group(total: Int, listed: [(String, String)], via: String = DeleteImpactCascadeVia.deletesHandles.wire) -> CatalogDeleteImpactGroup {
        CatalogDeleteImpactGroup(
            via: via, kind: "canonical_entity", total: total,
            listed: listed.map { id, ref in
                CatalogDeleteImpactListed(
                    id: id, ref: ref, title: ref, location: WorkspaceLocation(section: .persons, entityId: id)
                )
            }
        )
    }

    private func item(
        handleID: String? = "h-maria", deletes: CatalogDeleteImpactGroup? = nil, failed: Bool = false
    ) -> UnpromoteConfirmItem {
        var item = UnpromoteConfirmItem(id: "s", name: "Maria", kind: .person, handleID: handleID, connections: .listed([]))
        item.deletedHandles = deletes
        item.impactCheckFailed = failed
        return item
    }

    @Test func messageFollowsThePreflight() {
        let own = item(deletes: group(total: 1, listed: [("h-maria", "PER-8JD2L")]))
        #expect(own.message == "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph.")
        #expect(own.confirmTone == .danger)

        let three = item(deletes: group(total: 3, listed: [("h-maria", "PER-8JD2L"), ("h-a", "REL-1"), ("h-b", "REL-2")]))
        let cascade = PVDeleteImpactCopy.cascadeLines(
            CatalogDeleteImpact(allowed: true, gate: .ok, groups: [], cascades: [three.deletedHandles!])
        ).first
        #expect(three.message == cascade)
        #expect(three.confirmTone == .danger)

        let other = item(deletes: group(total: 1, listed: [("h-rel", "REL-1")]))
        let otherCascade = PVDeleteImpactCopy.cascadeLines(
            CatalogDeleteImpact(allowed: true, gate: .ok, groups: [], cascades: [other.deletedHandles!])
        ).first
        #expect(!other.deletesOwnHandle)
        #expect(other.message == otherCascade)
        #expect(other.confirmTone == .danger)

        let failed = item(failed: true)
        #expect(failed.message == L10n.Unpromote.confirmPreflightFailed(kind: .person))
        #expect(failed.confirmTone == .danger)

        let none = item()
        #expect(none.message == L10n.string(L10n.Unpromote.confirmMessage))
        #expect(none.confirmTone == .irreversible)
    }

    @Test func listComesFromThePreflight() async throws {
        let (store, session) = makeFixture()
        let flow = UnpromoteFlow()
        func open() async {
            await flow.open(
                subjectID: "s-weber", name: "J. Weber", kind: .person, handleID: "h-weber",
                graphID: graphID, session: session
            )
        }
        func script(_ ids: [String]) {
            store.unpromoteImpactBySubject["s-weber"] = CatalogUnpromoteImpact(deletes: nil, unpromotedBridgeSubjectIDs: ids)
        }
        let first = try await sentence(session, "b1")
        let second = try await sentence(session, "b2")

        script(["b2", "b1"])
        await open()
        #expect(flow.pending?.connections == .listed([second, first]))

        script(["b2", "gone"])
        await open()
        #expect(flow.pending?.connections == .listed([second]))

        script(["gone"])
        await open()
        #expect(flow.pending?.connections == .unlisted)

        // Rows that cannot be read, with ids named.
        let (failingStore, failingSession) = makeFixture()
        struct Boom: Error {}
        failingStore.listSubjectsError = Boom()
        failingStore.unpromoteImpactBySubject["s-weber"] = CatalogUnpromoteImpact(
            deletes: nil, unpromotedBridgeSubjectIDs: ["b1"]
        )
        let failingFlow = UnpromoteFlow()
        await failingFlow.open(
            subjectID: "s-weber", name: "J. Weber", kind: .person, handleID: nil,
            graphID: graphID, session: failingSession
        )
        #expect(failingFlow.pending?.connections == .unlisted)

        // No ids, and a failing rows load: nothing to list.
        let (quietStore, quietSession) = makeFixture()
        quietStore.listSubjectsError = Boom()
        let quietFlow = UnpromoteFlow()
        await quietFlow.open(
            subjectID: "s-weber", name: "J. Weber", kind: .person, handleID: nil,
            graphID: graphID, session: quietSession
        )
        #expect(quietFlow.pending?.connections == .listed([]))

        // The preflight throws.
        store.getUnpromoteImpactError = Boom()
        await open()
        #expect(flow.pending?.connections == .unlisted)
        #expect(flow.pending?.impactCheckFailed == true)
    }

    @Test func staleOpenIsDropped() async {
        let (store, session) = makeFixture()
        let flow = UnpromoteFlow()
        store.getUnpromoteImpactHold.isOn = true
        let a = Task {
            await flow.open(subjectID: "s-weber", name: "A", kind: .person, handleID: nil, graphID: graphID, session: session)
        }
        #expect(await waitUntil { store.getUnpromoteImpactHold.count == 1 })
        let b = Task {
            await flow.open(subjectID: "s-john", name: "B", kind: .person, handleID: nil, graphID: graphID, session: session)
        }
        #expect(await waitUntil { store.getUnpromoteImpactHold.count == 2 })
        store.getUnpromoteImpactHold.release()
        #expect(await waitUntil { store.getUnpromoteImpactHold.count == 1 })
        store.getUnpromoteImpactHold.isOn = false
        store.getUnpromoteImpactHold.release()
        await a.value
        await b.value
        #expect(flow.pending?.id == "s-john")
        #expect(!flow.isBusy)
    }

    @Test func confirmReportsTheEffects() async throws {
        let (store, session) = makeFixture()
        let flow = UnpromoteFlow()
        func open() async {
            await flow.open(
                subjectID: "s-weber", name: "J. Weber", kind: .person, handleID: "h-maria",
                graphID: graphID, session: session
            )
        }
        // Busy while the open reads.
        store.getUnpromoteImpactHold.isOn = true
        let opening = Task { await open() }
        #expect(await waitUntil { store.getUnpromoteImpactHold.count == 1 })
        #expect(flow.isBusy)
        store.getUnpromoteImpactHold.isOn = false
        store.getUnpromoteImpactHold.release()
        await opening.value
        #expect(!flow.isBusy)

        // The person is deleted: `apply` receives the effects, and leaving there evicts her page.
        store.nextWriteEffects = CatalogWriteEffects(
            handles: ["h-maria"], deleted: [CatalogWriteEffectsDeleted(entity: .handle, ids: ["h-maria"])]
        )
        let detailKey = CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: "h-maria")
        session.apply(location: .personDetail(entityId: "h-maria", ref: "PER-8JD2L", title: "Maria"))
        let warmed: QueryHandle<CatalogConclusionDetail>? = session.queryHandle(detailKey)
        #expect(warmed != nil)
        var seen: [String] = []
        store.unpromoteSubjectHold.isOn = true
        let writing = Task {
            await flow.confirm(session: session, userID: "user-1", apply: { effects in
                seen = effects.deletedHandles
                session.apply(effects, leavingPage: { session.apply(location: .sectionRoot(.persons)) })
            })
        }
        #expect(await waitUntil { store.unpromoteSubjectHold.count == 1 })
        #expect(flow.isBusy)
        store.unpromoteSubjectHold.isOn = false
        store.unpromoteSubjectHold.release()
        #expect(await writing.value)
        #expect(seen == ["h-maria"])
        #expect(flow.pending == nil)
        #expect(!flow.isBusy)
        let evicted: QueryHandle<CatalogConclusionDetail>? = session.queryHandle(detailKey)
        #expect(evicted == nil)

        // The person stayed.
        store.nextWriteEffects = CatalogWriteEffects(handles: ["h-maria"])
        await open()
        var stayedSeen: CatalogWriteEffects?
        #expect(await flow.confirm(session: session, userID: "user-1", apply: { stayedSeen = $0 }))
        #expect(stayedSeen?.deletedHandles.isEmpty == true)
    }

    @Test func confirmFailureStaysOpen() async {
        let (store, session) = makeFixture()
        let flow = UnpromoteFlow()
        await flow.open(
            subjectID: "s-weber", name: "J. Weber", kind: .person, handleID: nil, graphID: graphID, session: session
        )
        let failure = CoreInvokeError.coded(status: 1, code: "promote.stale", kind: .user, params: [])
        store.unpromoteSubjectError = failure
        let result = await flow.confirm(session: session, userID: "user-1")
        #expect(!result)
        #expect(flow.pending?.error == L10n.Errors.message(for: failure))
        store.unpromoteSubjectError = nil
        // A retry clears the error before the write returns.
        store.unpromoteSubjectHold.isOn = true
        let retry = Task { await flow.confirm(session: session, userID: "user-1") }
        #expect(await waitUntil { store.unpromoteSubjectHold.count == 1 })
        #expect(flow.pending?.error == nil)
        store.unpromoteSubjectHold.isOn = false
        store.unpromoteSubjectHold.release()
        #expect(await retry.value)
        #expect(flow.pending == nil)
    }
}
