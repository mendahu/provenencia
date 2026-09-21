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
            CatalogSubjectType(
                id: "type-location",
                key: "location",
                origin: "provenencia",
                label: "Location",
                description: "",
                refPrefix: "LOC",
                candidateRefPrefix: "CLO"
            ),
            CatalogSubjectType(
                id: "type-participation",
                key: "participation",
                origin: "provenencia",
                label: "Participation",
                description: "",
                refPrefix: "PTN",
                candidateRefPrefix: "CPA"
            ),
            CatalogSubjectType(
                id: "type-relationship",
                key: "relationship",
                origin: "provenencia",
                label: "Relationship",
                description: "",
                refPrefix: "REL",
                candidateRefPrefix: "CRL"
            ),
        ]
        return store
    }

    private func makeModel(
        store: FakeStore,
        linkStore: (any EvidenceProvisionalLinkStoring)? = nil
    ) -> EvidenceGraphModel {
        let session = WorkspaceSession(
            projectKey: ProjectKey(projectDir: projectDir),
            store: store
        )
        return EvidenceGraphModel(
            sourceID: sourceID,
            session: session,
            store: store,
            userID: "user-1",
            linkStore: linkStore ?? InMemoryEvidenceProvisionalLinkStore()
        )
    }

    @Test func prepareCachesPrimaryTypeIDs() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.typeIDByKind["person"] == personTypeID)
        #expect(model.typeIDByKind["event"] == "type-event")
        #expect(model.typeIDByKind["location"] == "type-location")
        #expect(model.typeIDByKind["participation"] == "type-participation")
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
        #expect(effects == [.key(.sourceGraph(project: project, sourceId: sourceID))])
    }

    @Test func connectExcludesPlacingAndPickCreatesBridge() async throws {
        let store = makeStore()
        let links = InMemoryEvidenceProvisionalLinkStore()
        let model = makeModel(store: store, linkStore: links)
        await model.prepare()

        let person = CatalogSubject(
            id: "p1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Alice",
            description: ""
        )
        let event = CatalogSubject(
            id: "e1",
            ref: "CEV-1",
            sourceID: sourceID,
            subjectTypeID: "type-event",
            label: "Birth",
            description: ""
        )
        store.subjectsBySource[sourceID] = [person, event]
        store.subjectPositionsBySubject["p1"] = CatalogSubjectPosition(
            subjectID: "p1", gridX: 1, gridY: 1
        )
        store.subjectPositionsBySubject["e1"] = CatalogSubjectPosition(
            subjectID: "e1", gridX: 5, gridY: 1
        )

        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: [person, event],
                positions: [
                    CatalogSubjectPosition(subjectID: "p1", gridX: 1, gridY: 1),
                    CatalogSubjectPosition(subjectID: "e1", gridX: 5, gridY: 1),
                ],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )

        model.toggleConnect()
        #expect(model.inputMode == .connecting)
        model.toggleArm(.person)
        #expect(model.armedConnect == false)
        #expect(model.inputMode == .placing(.person))

        model.toggleConnect()
        model.handleConnectPick(subjectID: "p1", kind: .person, label: "Alice")
        #expect(model.connectOriginID == "p1")
        model.handleConnectPick(subjectID: "e1", kind: .event, label: "Birth")
        #expect(model.isCreating)
        #expect(model.pendingBridgeKind == .participation)

        model.draft.label = "Witness"
        let bridgeID = await model.confirmCreate()
        #expect(bridgeID != nil)
        #expect(model.armedConnect == false)
        #expect(links.links(for: sourceID).first?.endpointAID == "p1")
        #expect(links.links(for: sourceID).first?.endpointBID == "e1")

        let subjects = try await store.listSubjects(projectDir: projectDir, sourceID: sourceID)
        #expect(subjects.contains(where: { $0.label == "Witness" }))
    }

    @Test func cancelBridgeCreateKeepsOriginHeld() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()

        store.subjectTypesByProject[projectDir]?.append(
            CatalogSubjectType(
                id: "type-place",
                key: "place",
                origin: "provenencia",
                label: "Place",
                description: "",
                refPrefix: "PLC",
                candidateRefPrefix: "CPL"
            )
        )
        await model.prepare()

        let person = CatalogSubject(
            id: "p1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Alice",
            description: ""
        )
        let placeSubject = CatalogSubject(
            id: "pl1",
            ref: "CPL-1",
            sourceID: sourceID,
            subjectTypeID: "type-place",
            label: "Town",
            description: ""
        )
        store.subjectsBySource[sourceID] = [person, placeSubject]
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: [person, placeSubject],
                positions: [
                    CatalogSubjectPosition(subjectID: "p1", gridX: 0, gridY: 0),
                    CatalogSubjectPosition(subjectID: "pl1", gridX: 4, gridY: 0),
                ],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )

        model.toggleConnect()
        model.handleConnectPick(subjectID: "p1", kind: .person, label: "Alice")
        model.handleConnectPick(subjectID: "pl1", kind: .place, label: "Town")
        #expect(model.pendingBridgeKind == .location)
        model.cancelCreate()
        #expect(model.isCreating == false)
        #expect(model.armedConnect == true)
        #expect(model.connectOriginID == "p1")
    }

    @Test func invalidConnectPairToastsAndKeepsOrigin() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let e1 = CatalogSubject(
            id: "e1",
            ref: "CEV-1",
            sourceID: sourceID,
            subjectTypeID: "type-event",
            label: "A",
            description: ""
        )
        let e2 = CatalogSubject(
            id: "e2",
            ref: "CEV-2",
            sourceID: sourceID,
            subjectTypeID: "type-event",
            label: "B",
            description: ""
        )
        store.subjectsBySource[sourceID] = [e1, e2]
        store.subjectPositionsBySubject["e1"] = CatalogSubjectPosition(
            subjectID: "e1", gridX: 0, gridY: 0
        )
        store.subjectPositionsBySubject["e2"] = CatalogSubjectPosition(
            subjectID: "e2", gridX: 2, gridY: 0
        )
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: [e1, e2],
                positions: [
                    CatalogSubjectPosition(subjectID: "e1", gridX: 0, gridY: 0),
                    CatalogSubjectPosition(subjectID: "e2", gridX: 2, gridY: 0),
                ],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )

        model.toggleConnect()
        model.handleConnectPick(subjectID: "e1", kind: .event, label: "A")
        model.handleConnectPick(subjectID: "e2", kind: .event, label: "B")
        #expect(model.isCreating == false)
        #expect(model.connectOriginID == "e1")
        #expect(model.toast?.tone == .danger)
    }

    @Test func confirmEditUpdatesSubjectAndClosesSheet() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let person = CatalogSubject(
            id: "p1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Old",
            description: "Was"
        )
        store.subjectsBySource[sourceID] = [person]
        store.subjectPositionsBySubject["p1"] = CatalogSubjectPosition(
            subjectID: "p1", gridX: 1, gridY: 2
        )
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphSnapshot.build(
                sourceId: sourceID,
                subjects: [person],
                positions: [CatalogSubjectPosition(subjectID: "p1", gridX: 1, gridY: 2)],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )

        model.beginEdit(subjectID: "p1")
        #expect(model.editingSubjectID == "p1")
        #expect(model.draft.label == "Old")
        model.draft.label = "New name"
        model.draft.description = "Updated"
        let id = await model.confirmEdit()
        #expect(id == "p1")
        #expect(model.editingSubjectID == nil)
        #expect(store.subjectsBySource[sourceID]?.first?.label == "New name")
        #expect(store.subjectsBySource[sourceID]?.first?.description == "Updated")
    }

    @Test func noArtifactDisablesArmingAndComposerLocation() async {
        let store = makeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "type-book",
                title: "Bare",
                description: "",
                hasArtifact: false
            ),
        ]
        let model = makeModel(store: store)
        let listKey = CatalogQueryKey.sourcesList(project: model.session.projectKey)
        model.session.setQueryValue(listKey, value: store.sourcesByProject[projectDir] ?? [])
        await model.prepare()
        #expect(model.canCite == false)
        model.toggleArm(.person)
        #expect(model.armedKind == nil)
        #expect(model.composerLocation(for: "p1") == nil)
    }
}
