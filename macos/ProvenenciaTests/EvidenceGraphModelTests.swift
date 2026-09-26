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
            CatalogSubjectType(
                id: "type-place",
                key: "place",
                origin: "provenencia",
                label: "Place",
                description: "",
                refPrefix: "PLC",
                candidateRefPrefix: "CPL"
            ),
        ]
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: "prop-role",
                key: "role",
                origin: "provenencia",
                label: "Role",
                description: "",
                valueType: "term"
            ),
            CatalogProperty(
                id: "prop-rel-type",
                key: "relationship_type",
                origin: "provenencia",
                label: "Relationship type",
                description: "",
                valueType: "term"
            ),
        ]
        store.propertyTermsByProperty["prop-role"] = [
            CatalogPropertyTerm(
                id: "term-witness",
                propertyID: "prop-role",
                key: "witness",
                origin: "provenencia",
                label: "Witness",
                description: ""
            ),
        ]
        store.propertyTermsByProperty["prop-rel-type"] = [
            CatalogPropertyTerm(
                id: "term-spouse",
                propertyID: "prop-rel-type",
                key: "spouse",
                origin: "provenencia",
                label: "Spouse",
                description: ""
            ),
        ]
        return store
    }

    private func seedFields(_ session: WorkspaceSession, store: FakeStore) {
        session.setQueryValue(
            CatalogQueryKey.subjectFieldsWorkspace(project: session.projectKey),
            value: SubjectFieldsSnapshot(
                properties: store.propertiesByProject[projectDir] ?? [],
                types: store.subjectTypesByProject[projectDir] ?? [],
                fieldsByTypeID: store.subjectTypeFieldsByType,
                presentationsByKey: [:]
            )
        )
    }

    /// Session graph payload. `types` is ignored — vocabulary lives on `subjectFieldsWorkspace`.
    private func graphRows(
        sourceId: String,
        subjects: [CatalogSubject],
        positions: [CatalogSubjectPosition],
        types: [CatalogSubjectType] = [],
        observations: [CatalogObservation] = []
    ) -> SourceGraphRows {
        _ = types
        return SourceGraphRows(
            sourceId: sourceId,
            subjects: subjects,
            positions: positions,
            observations: observations
        )
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
        #expect(model.typeIDByKind["location"] == "type-location")
        #expect(model.typeIDByKind["participation"] == "type-participation")
    }

    @Test func confirmCreateWritesSubjectAndPositionThenDisarms() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let _: QueryHandle<SourceGraphRows> = model.session.query(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        )

        model.toggleArm(.person)
        model.beginCreate(at: CGPoint(x: 100, y: 140))
        #expect(model.isCreating)
        #expect(model.pendingGridX == 2)
        #expect(model.pendingGridY == 3)
        model.draft.label = "Alice"
        store.recordedCalls = []

        let id = await model.confirmCreate()
        #expect(id != nil)
        #expect(model.armedKind == nil)
        #expect(model.isCreating == false)
        #expect(store.recordedCalls.contains { $0.hasPrefix("createSubject") && $0.contains("placement=2,3") })
        #expect(store.recordedCalls.contains { $0.hasPrefix("setSubjectPosition") } == false)

        let subjects = try await store.listSubjects(projectDir: projectDir, sourceID: sourceID)
        #expect(subjects.count == 1)
        #expect(subjects.first?.label == "Alice")
        let positions = try await store.listSubjectPositions(projectDir: projectDir, sourceID: sourceID)
        #expect(positions.first?.gridX == 2)
        #expect(positions.first?.gridY == 3)

        // Session must re-query so the canvas leaves the stale empty snapshot.
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        )
        var waited = 0
        while handle?.value?.subjects.isEmpty != false, waited < 40 {
            try await Task.sleep(nanoseconds: 25_000_000)
            waited += 1
        }
        #expect(handle?.value?.subjects.count == 1)
        #expect(handle?.value?.subjects.first?.label == "Alice")
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
            value: SourceGraphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 1, gridY: 2)]
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
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.positions.first?.gridX == 1)
        #expect(handle?.value?.positions.first?.gridY == 2)
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
        let rows = SourceGraphRows(
            sourceId: sourceID,
            subjects: [subject],
            positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 1, gridY: 1)]
        )
        let next = rows.updatingPosition(subjectID: "s1", gridX: 4, gridY: 5)
        #expect(next.positions[0].gridX == 4)
        #expect(next.positions[0].gridY == 5)
    }

    @Test func mutatedSourceGraphInvalidatesSourceGraph() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        let effects = registry.invalidations(
            by: .mutatedSourceGraph(sourceId: sourceID),
            project: project
        )
        #expect(effects == [.key(.sourceGraph(project: project, sourceId: sourceID))])
    }

    @Test func connectExcludesPlacingAndPickCreatesBridge() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
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
            value: graphRows(
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
        await model.handleConnectPick(subjectID: "p1", kind: .person, label: "Alice")
        #expect(model.connectOriginID == "p1")
        await model.handleConnectPick(subjectID: "e1", kind: .event, label: "Birth")
        #expect(model.isCreating == false)
        let handoff = model.consumeComposerHandoff()
        #expect(handoff?.isConnectPrefill == true)
        #expect(handoff?.subjectId == nil)
        #expect(handoff?.connectFromSubjectId == "p1")
        #expect(handoff?.connectToSubjectId == "e1")
        #expect(handoff?.connectBridgeTypeKey == "participation")
        #expect(handoff?.sourceSurface == .citationComposer)

        let subjects = try await store.listSubjects(projectDir: projectDir, sourceID: sourceID)
        #expect(subjects.count == 2)
    }

    @Test func eventPlaceConnectSkipsSheetAndHandsOffComposer() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let event = CatalogSubject(
            id: "e1",
            ref: "CEV-1",
            sourceID: sourceID,
            subjectTypeID: "type-event",
            label: "Birth",
            description: ""
        )
        let place = CatalogSubject(
            id: "pl1",
            ref: "CPL-1",
            sourceID: sourceID,
            subjectTypeID: "type-place",
            label: "Leeds",
            description: ""
        )
        store.subjectsBySource[sourceID] = [event, place]
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [event, place],
                positions: [
                    CatalogSubjectPosition(subjectID: "e1", gridX: 0, gridY: 0),
                    CatalogSubjectPosition(subjectID: "pl1", gridX: 4, gridY: 0),
                ],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )
        model.toggleConnect()
        await model.handleConnectPick(subjectID: "e1", kind: .event, label: "Birth")
        await model.handleConnectPick(subjectID: "pl1", kind: .place, label: "Leeds")
        #expect(model.isCreating == false)
        let handoff = model.consumeComposerHandoff()
        #expect(handoff?.isConnectPrefill == true)
        #expect(handoff?.connectBridgeTypeKey == "location")
        #expect(handoff?.subjectId == nil)
    }

    @Test func personPlaceConnectToastsAndKeepsOrigin() async {
        let store = makeStore()
        let model = makeModel(store: store)
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
            value: graphRows(
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
        await model.handleConnectPick(subjectID: "p1", kind: .person, label: "Alice")
        await model.handleConnectPick(subjectID: "pl1", kind: .place, label: "Town")
        #expect(model.isCreating == false)
        #expect(model.armedConnect == true)
        #expect(model.connectOriginID == "p1")
        #expect(model.toast?.tone == .danger)
        #expect(String(localized: L10n.EvidenceGraph.connectInvalidPairBody).contains("Person and place"))
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
            value: graphRows(
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
        await model.handleConnectPick(subjectID: "e1", kind: .event, label: "A")
        await model.handleConnectPick(subjectID: "e2", kind: .event, label: "B")
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
            value: graphRows(
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

    @Test func bridgeEditSendsStoredLabelUnchanged() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let bridge = CatalogSubject(
            id: "b1",
            ref: "CPA-1",
            sourceID: sourceID,
            subjectTypeID: "type-participation",
            label: "Stored",
            description: "Was"
        )
        store.subjectsBySource[sourceID] = [bridge]
        store.subjectPositionsBySubject["b1"] = CatalogSubjectPosition(
            subjectID: "b1", gridX: 2, gridY: 2
        )
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [bridge],
                positions: [CatalogSubjectPosition(subjectID: "b1", gridX: 2, gridY: 2)],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )
        model.beginEdit(subjectID: "b1")
        model.draft.label = "Changed in the dialog"
        model.draft.description = "New note"
        store.recordedCalls = []
        let id = await model.confirmEdit()
        #expect(id == "b1")
        #expect(store.recordedCalls.contains { $0 == "updateSubject id=b1 label=Stored" })
        #expect(store.subjectsBySource[sourceID]?.first?.label == "Stored")
        #expect(store.subjectsBySource[sourceID]?.first?.description == "New note")
    }

    @Test func failedOlderDragDoesNotRevertNewerMove() async {
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
            subjectID: "s1", gridX: 1, gridY: 2
        )
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 1, gridY: 2)]
            )
        )
        store.setSubjectPositionDelayNanoseconds = 80_000_000
        store.setSubjectPositionErrors = [
            CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: []),
            nil,
        ]
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 1,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        try? await Task.sleep(nanoseconds: 10_000_000)
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 5,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        try? await Task.sleep(nanoseconds: 150_000_000)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.positions.first?.gridX == 9)
        #expect(handle?.value?.positions.first?.gridY == 2)
        #expect(store.subjectPositionsBySubject["s1"]?.gridX == 9)
    }

    @Test func successfulOlderDragIsConfirmedWhenNewerFails() async {
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
            subjectID: "s1", gridX: 1, gridY: 2
        )
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: SourceGraphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 1, gridY: 2)]
            )
        )
        store.setSubjectPositionDelayNanoseconds = 80_000_000
        store.setSubjectPositionErrors = [
            nil,
            CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: []),
        ]
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 1,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        try? await Task.sleep(nanoseconds: 10_000_000)
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 5,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        try? await Task.sleep(nanoseconds: 200_000_000)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.positions.first?.gridX == 5)
        #expect(handle?.value?.positions.first?.gridY == 2)
        #expect(store.subjectPositionsBySubject["s1"]?.gridX == 5)
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

    @Test func deleteUncitedSubjectSucceeds() async throws {
        let store = makeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "type-book",
                title: "Census",
                description: "",
                hasArtifact: true
            ),
        ]
        let subject = CatalogSubject(
            id: "s-uncited",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Alice",
            description: ""
        )
        store.subjectsBySource[sourceID] = [subject]
        store.subjectPositionsBySubject["s-uncited"] = CatalogSubjectPosition(
            subjectID: "s-uncited",
            gridX: 0,
            gridY: 0
        )
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s-uncited", gridX: 0, gridY: 0)]
            )
        )
        model.beginDelete(subjectID: "s-uncited")
        #expect(model.pendingDelete?.id == "s-uncited")
        #expect(model.pendingDelete?.label == "Alice")
        #expect(model.pendingDelete?.ref == "CPR-1")
        let ok = await model.confirmDeleteSubject()
        #expect(ok)
        #expect(model.pendingDelete == nil)
        #expect(store.subjectsBySource[sourceID]?.isEmpty == true)
        // Allow the invalidated sourceGraph query to finish reloading.
        try await Task.sleep(for: .milliseconds(50))
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        )
        #expect(handle?.value?.subjects.isEmpty == true)
    }

    @Test func deleteUncitedBridgeRemovesCardAndProvisionalLink() async throws {
        let store = makeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "type-book",
                title: "Census",
                description: "",
                hasArtifact: true
            ),
        ]
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
        let bridge = CatalogSubject(
            id: "b1",
            ref: "CPA-1",
            sourceID: sourceID,
            subjectTypeID: "type-participation",
            label: "Witness",
            description: ""
        )
        store.subjectsBySource[sourceID] = [person, event, bridge]
        store.subjectPositionsBySubject["p1"] = CatalogSubjectPosition(
            subjectID: "p1", gridX: 0, gridY: 0
        )
        store.subjectPositionsBySubject["e1"] = CatalogSubjectPosition(
            subjectID: "e1", gridX: 4, gridY: 0
        )
        store.subjectPositionsBySubject["b1"] = CatalogSubjectPosition(
            subjectID: "b1", gridX: 2, gridY: 0
        )
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [person, event, bridge],
                positions: [
                    CatalogSubjectPosition(subjectID: "p1", gridX: 0, gridY: 0),
                    CatalogSubjectPosition(subjectID: "e1", gridX: 4, gridY: 0),
                    CatalogSubjectPosition(subjectID: "b1", gridX: 2, gridY: 0),
                ],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )
        model.beginDelete(subjectID: "b1")
        #expect(model.pendingDelete?.id == "b1")
        let ok = await model.confirmDeleteSubject()
        #expect(ok)
        try await Task.sleep(for: .milliseconds(50))
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.subjects.contains { $0.id == "b1" } == false)
        #expect(handle?.value?.subjects.count == 2)
    }

    @Test func beginDeleteIgnoresCitedSubject() async {
        let store = makeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "type-book",
                title: "Census",
                description: "",
                hasArtifact: true
            ),
        ]
        let subject = CatalogSubject(
            id: "s-cited",
            ref: "CPR-2",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Bob",
            description: ""
        )
        let observation = CatalogObservation(
            id: "obs-1",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "s-cited",
            propertyID: "prop-1",
            polarity: "positive",
            valueText: "Farmer",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "occupation",
            propertyLabel: "Occupation",
            propertyValueType: "text"
        )
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s-cited", gridX: 0, gridY: 0)],
                observations: [observation]
            )
        )
        model.beginDelete(subjectID: "s-cited")
        #expect(model.pendingDelete == nil)
    }

    @Test func propertyEditLocationsShareCitationId() async {
        let store = makeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "type-book",
                title: "Census",
                description: "",
                hasArtifact: true
            ),
        ]
        store.artifactsBySource[sourceID] = [
            CatalogArtifact(
                id: "art-0",
                ref: "ART-0",
                sourceID: sourceID,
                fileID: "file-0",
                label: "Scan",
                description: "",
                file: nil
            ),
        ]
        let subject = CatalogSubject(
            id: "s1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Alice",
            description: ""
        )
        let obsA = CatalogObservation(
            id: "obs-a",
            ref: "OBS-A",
            citationID: "cit-shared",
            subjectID: "s1",
            propertyID: "p1",
            polarity: "positive",
            valueText: "A",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "occupation",
            propertyLabel: "Occupation",
            propertyValueType: "text"
        )
        let obsB = CatalogObservation(
            id: "obs-b",
            ref: "OBS-B",
            citationID: "cit-shared",
            subjectID: "s1",
            propertyID: "p2",
            polarity: "positive",
            valueText: "B",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "age",
            propertyLabel: "Age",
            propertyValueType: "integer"
        )
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.canCite)
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 0, gridY: 0)],
                observations: [obsA, obsB]
            )
        )
        let locA = model.composerLocation(forObservationID: "obs-a", subjectID: "s1")
        let locB = model.composerLocation(forObservationID: "obs-b", subjectID: "s1")
        #expect(locA != locB)
        #expect(locA?.citationId == "cit-shared")
        #expect(locB?.citationId == "cit-shared")
        #expect(locA?.observationId == "obs-a")
        #expect(locB?.observationId == "obs-b")
        #expect(locA?.subjectId == "s1")
        #expect(model.composerLocation(for: "s1")?.citationId == nil)
    }

    @Test func bridgeCitationEditOpensExistingCitation() async {
        let store = makeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "type-book",
                title: "Census",
                description: "",
                hasArtifact: true
            ),
        ]
        store.artifactsBySource[sourceID] = [
            CatalogArtifact(
                id: "art-0",
                ref: "ART-0",
                sourceID: sourceID,
                fileID: "file-0",
                label: "Scan",
                description: "",
                file: nil
            ),
        ]
        let bridgeSubject = CatalogSubject(
            id: "b1",
            ref: "CPA-1",
            sourceID: sourceID,
            subjectTypeID: "type-participation",
            label: "Working",
            description: ""
        )
        let observation = CatalogObservation(
            id: "obs-edge",
            ref: "OBS-1",
            citationID: "cit-bridge",
            subjectID: "b1",
            propertyID: "p-person",
            polarity: "positive",
            valueText: "Alice",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "s1",
            valueTermID: "",
            propertyKey: "person",
            propertyLabel: "Person",
            propertyValueType: "subject"
        )
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.canCite)
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [bridgeSubject],
                positions: [CatalogSubjectPosition(subjectID: "b1", gridX: 0, gridY: 0)],
                observations: [observation]
            )
        )
        let location = model.composerLocationForBridgeCitation(subjectID: "b1")
        #expect(location?.citationId == "cit-bridge")
        #expect(location?.subjectId == "b1")
        #expect(location?.observationId == "obs-edge")
        #expect(model.composerLocation(for: "b1")?.citationId == nil)
        #expect(
            model.composerLocation(forObservationID: "obs-edge", subjectID: "b1")?.citationId
                == "cit-bridge"
        )
        #expect(
            model.composerLocation(forObservationID: "obs-edge", subjectID: "b1")?.observationId
                == "obs-edge"
        )
    }

    @Test func connectRulesFailureDisablesConnect() async {
        let store = makeStore()
        store.listConnectRulesError = CoreInvokeError.failed(status: 1)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.connectRules.isEmpty)
        #expect(model.connectRulesError != nil)
        model.toggleConnect()
        #expect(model.armedConnect == false)
        #expect(model.toast?.tone == .danger)
    }

    @Test func escapeDisarmsConnect() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        model.toggleConnect()
        #expect(model.armedConnect)
        let effect = model.handleKey(.escape, hasFocus: false)
        #expect(effect.handled)
        #expect(model.armedConnect == false)
    }

    @Test func arrowRequestsMoveOfSelectedSubject() {
        let store = makeStore()
        let model = makeModel(store: store)
        let person = CatalogSubject(
            id: "p1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Alice",
            description: ""
        )
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        seedFields(model.session, store: store)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [person],
                positions: [CatalogSubjectPosition(subjectID: "p1", gridX: 2, gridY: 3)],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )
        model.selectSubject(id: "p1")
        let effect = model.handleKey(.arrow(dx: 1, dy: 0), hasFocus: false)
        #expect(effect.handled)
        #expect(effect.moveSubjectID == "p1")
        #expect(effect.moveFromX == 2)
        #expect(effect.moveFromY == 3)
        #expect(effect.moveDeltaX == 1)
    }

    @Test func cardActionRoutesEdit() {
        let store = makeStore()
        let model = makeModel(store: store)
        let person = CatalogSubject(
            id: "p1",
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Alice",
            description: ""
        )
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        seedFields(model.session, store: store)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [person],
                positions: [CatalogSubjectPosition(subjectID: "p1", gridX: 0, gridY: 0)],
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )
        let location = model.performCardAction(
            subjectID: "p1",
            actionID: EvidenceSubjectCard.editActionID
        )
        #expect(location == nil)
        #expect(model.editingSubjectID == "p1")
    }

    @Test func sourcePageLocationUsesSameSourceAndPageSurface() {
        let model = makeModel(store: makeStore())
        let location = model.sourcePageLocation(title: "1851 England Census", ref: "SRC-1")
        #expect(location.section == .sources)
        #expect(location.sourceId == sourceID)
        #expect(location.sourceSurface == .page)
        #expect(location.title == "1851 England Census")
        #expect(location.ref == "SRC-1")
        #expect(location.sourceSurface != .graph)
    }
}
