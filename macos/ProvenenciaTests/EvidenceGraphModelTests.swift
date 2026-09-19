import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct EvidenceGraphModelTests {
    private let projectDir = "/tmp/evidence-graph-model.provenencia"
    private let sourceID = "src-1"
    private let personTypeID = "type-person"

    private func makeStore() -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: personTypeID,
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                refPrefix: "PER",
                candidateRefPrefix: "CPR"
            ),
            CatalogSubjectType(
                id: "type-event",
                key: "event",
                origin: "provenencia",
                label: "Event",
                description: "",
                refPrefix: "EVT",
                candidateRefPrefix: "CEV"
            ),
        ]
        return store
    }

    private func makeModel(store: FakeStore) -> EvidenceGraphModel {
        let session = WorkspaceSession(
            projectKey: ProjectKey(projectDir: projectDir),
            store: store
        )
        return EvidenceGraphModel(
            sourceID: sourceID,
            session: session,
            store: store,
            userID: "user-1"
        )
    }

    @Test func prepareCachesPrimaryTypeIDs() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.typeIDByKind["person"] == personTypeID)
        #expect(model.typeIDByKind["event"] == "type-event")
        #expect(model.typeIDByKind["location"] == nil)
    }

    @Test func confirmCreateWritesSubjectAndPositionThenDisarms() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let _: QueryHandle<SourceGraphSnapshot> = model.session.query(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        )

        model.toggleArm(.person)
        model.beginCreate(at: CGPoint(x: 100, y: 140))
        #expect(model.isCreating)
        #expect(model.pendingGridX == 2)
        #expect(model.pendingGridY == 3)
        model.draft.label = "Alice"

        let id = await model.confirmCreate()
        #expect(id != nil)
        #expect(model.armedKind == nil)
        #expect(model.isCreating == false)

        let subjects = try await store.listSubjects(projectDir: projectDir, sourceID: sourceID)
        #expect(subjects.count == 1)
        #expect(subjects.first?.label == "Alice")
        let positions = try await store.listSubjectPositions(projectDir: projectDir, sourceID: sourceID)
        #expect(positions.first?.gridX == 2)
        #expect(positions.first?.gridY == 3)

        // Session must re-query so the canvas leaves the stale empty snapshot.
        let handle: QueryHandle<SourceGraphSnapshot>? = model.session.queryHandle(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        )
        var waited = 0
        while handle?.value?.subjects.isEmpty != false, waited < 40 {
            try await Task.sleep(nanoseconds: 25_000_000)
            waited += 1
        }
        #expect(handle?.value?.subjects.count == 1)
        #expect(handle?.value?.subjects.first?.subject.label == "Alice")
    }

    @Test func cancelCreateDisarmsTool() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        model.toggleArm(.event)
        model.beginCreate(at: .zero)
        model.cancelCreate()
        #expect(model.isCreating == false)
        #expect(model.armedKind == nil)
        #expect(model.inputMode == .idle)
    }

    @Test func inputModePlacingWhenArmedIdleWhenCreating() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.inputMode == .idle)
        model.toggleArm(.person)
        #expect(model.inputMode == .placing(.person))
        model.beginCreate(at: CGPoint(x: 40, y: 40))
        #expect(model.inputMode == .idle)
    }

    @Test func activateSubjectThenSelectOtherClearsActivation() {
        let store = makeStore()
        let model = makeModel(store: store)
        model.activateSubject(id: "a")
        #expect(model.selectedSubjectID == "a")
        #expect(model.activatedSubjectID == "a")
        model.selectSubject(id: "b")
        #expect(model.selectedSubjectID == "b")
        #expect(model.activatedSubjectID == nil)
        model.activateSubject(id: "b")
        model.deactivateSubject()
        #expect(model.activatedSubjectID == nil)
        #expect(model.selectedSubjectID == "b")
    }

    @Test func moveSubjectPersistFailureRevertsAndToasts() async {
        let store = makeStore()
        let subject = CatalogSubject(
            id: "s1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "A",
            description: ""
        )
        store.subjectsBySource[sourceID] = [subject]
        store.subjectPositionsBySubject["s1"] = CatalogSubjectPosition(
            subjectID: "s1",
            gridX: 1,
            gridY: 2
        )
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphSnapshot(
                sourceId: sourceID,
                subjects: [
                    SourceGraphPlacedSubject(
                        subject: subject,
                        kind: .person,
                        typeLabel: "Person",
                        gridX: 1,
                        gridY: 2,
                        isCited: false
                    ),
                ]
            )
        )

        store.setSubjectPositionError = CoreInvokeError.coded(
            status: 1, code: "internal.unknown", kind: .internal, params: []
        )
        let result = await model.moveSubject(
            subjectID: "s1",
            fromGridX: 1,
            fromGridY: 2,
            deltaX: 1,
            deltaY: 0
        )
        #expect(result == nil)
        let handle: QueryHandle<SourceGraphSnapshot>? = model.session.queryHandle(key)
        #expect(handle?.value?.subjects.first?.gridX == 1)
        #expect(handle?.value?.subjects.first?.gridY == 2)
        #expect(model.toast?.tone == .danger)
        #expect(model.toast?.title == String(localized: L10n.EvidenceGraph.positionPersistFailedTitle))
    }

    @Test func updatingPositionPatchesSnapshot() {
        let subject = CatalogSubject(
            id: "s1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "A",
            description: ""
        )
        let snapshot = SourceGraphSnapshot(
            sourceId: sourceID,
            subjects: [
                SourceGraphPlacedSubject(
                    subject: subject,
                    kind: .person,
                    typeLabel: "Person",
                    gridX: 1,
                    gridY: 1,
                    isCited: false
                ),
            ]
        )
        let next = snapshot.updatingPosition(subjectID: "s1", gridX: 4, gridY: 5)
        #expect(next.subjects[0].gridX == 4)
        #expect(next.subjects[0].gridY == 5)
    }

    @Test func createdSubjectInvalidatesSourceGraph() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        let effects = registry.invalidations(
            by: .createdSubject(sourceId: sourceID),
            project: project
        )
        #expect(effects == [.allCached(.sourceGraph)])
    }
}
