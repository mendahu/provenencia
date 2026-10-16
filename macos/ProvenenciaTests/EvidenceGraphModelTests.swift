import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct EvidenceGraphModelTests {
    private let projectDir = "/tmp/evidence-graph-model.provenencia"
    private let sourceID = "src-1"
    private let personTypeID = "type-person"

    private func waitUntil(
        timeout: Duration = .seconds(5),
        _ predicate: () -> Bool
    ) async -> Bool {
        let deadline = ContinuousClock.now + timeout
        while ContinuousClock.now < deadline {
            if predicate() { return true }
            try? await Task.sleep(for: .milliseconds(5))
        }
        return predicate()
    }

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
            CatalogQueryKey.propertiesWorkspace(project: session.projectKey),
            value: PropertiesSnapshot(
                properties: store.propertiesByProject[projectDir] ?? [],
                types: store.subjectTypesByProject[projectDir] ?? [],
                propertiesByTypeID: store.subjectTypePropertiesByType,
                presentationsByKey: [:]
            )
        )
    }

    /// Session graph payload. `types` is ignored — vocabulary lives on `propertiesWorkspace`.
    private func graphRows(
        sourceId: String,
        subjects: [CatalogSubject],
        positions: [CatalogSubjectPosition],
        types: [CatalogSubjectType] = [],
        observations: [CatalogObservation] = [],
        memberships: [CatalogSubjectMembership] = []
    ) -> SourceGraphRows {
        _ = types
        return SourceGraphRows(
            sourceId: sourceId,
            subjects: subjects,
            positions: positions,
            observations: observations,
            memberships: memberships
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

    /// Card copy is formatted once per snapshot, matches the card formatters,
    /// and rebuilds when the graph rows change.
    @Test func displayTextFollowsTheSnapshot() throws {
        let store = makeStore()
        let model = makeModel(store: store)
        let types = store.subjectTypesByProject[projectDir] ?? []
        func rows(label: String) -> SourceGraphRows {
            SourceGraphRows(
                sourceId: sourceID,
                subjects: [CatalogSubject(
                    id: "s1",
                    ref: "CPR-1",
                    sourceID: sourceID,
                    subjectTypeID: personTypeID,
                    label: label,
                    description: ""
                )],
                positions: [CatalogSubjectPosition(subjectID: "s1", gridX: 0, gridY: 0)]
            )
        }

        let first = rows(label: "Wm Robins")
        let placed = try #require(model.displaySnapshot(rows: first, types: types).subjects.first)
        let text = model.displayText(rows: first, types: types)[subject: placed]
        #expect(text.accessibilityLabel == EvidenceSubjectCard.accessibilityLabel(for: placed))
        #expect(text.footerActionName == EvidenceSubjectCard.footerAccessibilityActionName(for: placed))
        #expect(text.deleteActionName == L10n.EvidenceGraph.deleteAccessibility(
            kind: "person", label: "Wm Robins", ref: "CPR-1"
        ))
        #expect(model.displayText(rows: first, types: types) == model.displayText(rows: first, types: types))

        let renamed = rows(label: "William Robins")
        let renamedPlaced = try #require(model.displaySnapshot(rows: renamed, types: types).subjects.first)
        #expect(model.displayText(rows: renamed, types: types)[subject: renamedPlaced].accessibilityLabel
            .contains("William Robins"))
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
        let gridX = 2 + EvidenceGraphPlacement.originReserveCells
        let gridY = 3 + EvidenceGraphPlacement.originReserveCells
        #expect(model.isCreating)
        #expect(model.pendingGridX == gridX)
        #expect(model.pendingGridY == gridY)
        model.draft.label = "Alice"
        store.recordedCalls = []

        let id = await model.confirmCreate()
        #expect(id != nil)
        #expect(model.armedKind == nil)
        #expect(model.isCreating == false)
        #expect(store.recordedCalls.contains {
            $0.hasPrefix("createSubject") && $0.contains("placement=\(gridX),\(gridY)")
        })
        #expect(store.recordedCalls.contains { $0.hasPrefix("setSubjectPosition") } == false)

        let subjects = try await store.listSubjects(projectDir: projectDir, sourceID: sourceID)
        #expect(subjects.count == 1)
        #expect(subjects.first?.label == "Alice")
        let positions = try await store.listSubjectPositions(projectDir: projectDir, sourceID: sourceID)
        #expect(positions.first?.gridX == gridX)
        #expect(positions.first?.gridY == gridY)

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
        #expect(model.toast?.title == L10n.string(L10n.EvidenceGraph.positionPersistFailedTitle))
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
        #expect(L10n.string(L10n.EvidenceGraph.connectInvalidPairBody).contains("Person and place"))
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

    @Test func failedOlderDragDoesNotRevertNewerMove() async throws {
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
        store.setSubjectPositionErrors = [
            CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: []),
            nil,
        ]
        store.holdsSetSubjectPosition = true
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 1,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        let olderWrite = model.lastPositionWrite
        try #require(await waitUntil { store.heldSetSubjectPositionCount == 1 })
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 5,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        let newerWrite = model.lastPositionWrite
        try #require(await waitUntil { store.heldSetSubjectPositionCount == 2 })
        // Settle the older write (fails) before the newer one (succeeds).
        store.releaseSetSubjectPosition()
        await olderWrite?.value
        store.releaseSetSubjectPosition()
        await newerWrite?.value
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.positions.first?.gridX == 9)
        #expect(handle?.value?.positions.first?.gridY == 2)
        #expect(store.subjectPositionsBySubject["s1"]?.gridX == 9)
    }

    @Test func successfulOlderDragIsConfirmedWhenNewerFails() async throws {
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
        store.setSubjectPositionErrors = [
            nil,
            CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: []),
        ]
        store.holdsSetSubjectPosition = true
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 1,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        let olderWrite = model.lastPositionWrite
        try #require(await waitUntil { store.heldSetSubjectPositionCount == 1 })
        _ = model.commitDrag(
            subjectID: "s1",
            originGridX: 5,
            originGridY: 2,
            documentDelta: CGSize(width: 160, height: 0)
        )
        let newerWrite = model.lastPositionWrite
        try #require(await waitUntil { store.heldSetSubjectPositionCount == 2 })
        // Settle the older write (succeeds) before the newer one (fails).
        store.releaseSetSubjectPosition()
        await olderWrite?.value
        store.releaseSetSubjectPosition()
        await newerWrite?.value
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
        await model.beginDelete(subjectID: "s-uncited")
        #expect(model.pendingImpact?.target.kind == "subject")
        #expect(model.pendingImpact?.target.ref == "CPR-1")
        #expect(model.pendingImpact?.target.title == "Alice")
        #expect(model.pendingImpact?.report.allowed == true)
        let ok = await model.confirmPendingImpact()
        #expect(ok)
        #expect(model.pendingImpact == nil)
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
        await model.beginDelete(subjectID: "b1")
        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.ref == "CPA-1")
        let ok = await model.confirmPendingImpact()
        #expect(ok)
        try await Task.sleep(for: .milliseconds(50))
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.subjects.contains { $0.id == "b1" } == false)
        #expect(handle?.value?.subjects.count == 2)
    }

    @Test func beginDeleteCitedEndpointPresentsNotice() async {
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
            id: "s-cited",
            ref: "CPR-2",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Bob",
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
            label: "",
            description: ""
        )
        store.subjectsBySource[sourceID] = [person, event, bridge]
        let edge = CatalogObservation(
            id: "obs-person",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "b1",
            propertyID: "prop-person",
            polarity: "positive",
            valueText: "Bob",
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "s-cited",
            valueTermID: "",
            propertyKey: "person",
            propertyLabel: "Person",
            propertyValueType: "subject"
        )
        store.observationsBySource[sourceID] = [edge]
        store.edgeObservationIDs.insert(edge.id)
        let model = makeModel(store: store)
        await model.prepare()
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        model.session.setQueryValue(
            key,
            value: graphRows(
                sourceId: sourceID,
                subjects: [person, event, bridge],
                positions: [
                    CatalogSubjectPosition(subjectID: "s-cited", gridX: 0, gridY: 0),
                    CatalogSubjectPosition(subjectID: "e1", gridX: 4, gridY: 0),
                    CatalogSubjectPosition(subjectID: "b1", gridX: 2, gridY: 0),
                ],
                observations: [edge]
            )
        )
        await model.beginDelete(subjectID: "s-cited")
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.groups.first?.via == "observations.value_subject_id")
        let ok = await model.confirmPendingImpact()
        #expect(!ok)
        #expect(store.subjectsBySource[sourceID]?.contains { $0.id == "s-cited" } == true)
        #expect(!store.recordedCalls.contains { $0.hasPrefix("deleteSubject") })
    }

    @Test func deleteFacetsOnlyBridgeReleasesEdges() async throws {
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
            label: "",
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
        let edges = [
            CatalogObservation(
                id: "obs-person",
                ref: "OBS-1",
                citationID: "cit-1",
                subjectID: "b1",
                propertyID: "prop-person",
                polarity: "positive",
                valueText: "Alice",
                valueInteger: nil,
                valueDateID: "",
                valueNameID: "",
                valueSubjectID: "p1",
                valueTermID: "",
                propertyKey: "person",
                propertyLabel: "Person",
                propertyValueType: "subject"
            ),
            CatalogObservation(
                id: "obs-event",
                ref: "OBS-2",
                citationID: "cit-1",
                subjectID: "b1",
                propertyID: "prop-event",
                polarity: "positive",
                valueText: "Birth",
                valueInteger: nil,
                valueDateID: "",
                valueNameID: "",
                valueSubjectID: "e1",
                valueTermID: "",
                propertyKey: "event",
                propertyLabel: "Event",
                propertyValueType: "subject"
            ),
            CatalogObservation(
                id: "obs-role",
                ref: "OBS-3",
                citationID: "cit-1",
                subjectID: "b1",
                propertyID: "prop-role",
                polarity: "positive",
                valueText: "Witness",
                valueInteger: nil,
                valueDateID: "",
                valueNameID: "",
                valueSubjectID: "",
                valueTermID: "term-witness",
                propertyKey: "role",
                propertyLabel: "Role",
                propertyValueType: "term"
            ),
        ]
        store.observationsBySource[sourceID] = edges
        store.edgeObservationIDs.formUnion(["obs-person", "obs-event"])
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
                types: store.subjectTypesByProject[projectDir] ?? [],
                observations: edges
            )
        )
        await model.beginDelete(subjectID: "b1")
        #expect(model.pendingImpact?.report.allowed == true)
        let ok = await model.confirmPendingImpact()
        #expect(ok)
        #expect(store.subjectsBySource[sourceID]?.contains { $0.id == "b1" } == false)
        #expect(store.subjectsBySource[sourceID]?.contains { $0.id == "p1" } == true)
        #expect(store.subjectsBySource[sourceID]?.contains { $0.id == "e1" } == true)
        #expect(store.observationsBySource[sourceID]?.isEmpty == true)
        try await Task.sleep(for: .milliseconds(50))
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        #expect(handle?.value?.subjects.contains { $0.id == "b1" } == false)
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
        #expect(effect.moves == [
            EvidenceGraphModel.KeyEffect.SubjectMove(
                subjectID: "p1", fromX: 2, fromY: 3, deltaX: 1, deltaY: 0
            ),
        ])
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

    // MARK: Promote (S9-04)

    private func promotableModel() async -> (FakeStore, EvidenceGraphModel) {
        let store = makeStore()
        let subject = CatalogSubject(
            id: "s-james", ref: "CPR-2AB91", sourceID: sourceID, subjectTypeID: personTypeID,
            label: "James Robins", description: ""
        )
        store.subjectsBySource[sourceID] = [subject]
        store.subjectPositionsBySubject["s-james"] = CatalogSubjectPosition(subjectID: "s-james", gridX: 1, gridY: 1)
        let model = makeModel(store: store)
        await model.prepare()
        model.session.setQueryValue(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID),
            value: graphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s-james", gridX: 1, gridY: 1)]
            )
        )
        return (store, model)
    }

    /// The card's Promote opens the Promote place (S9-11) and writes nothing.
    @Test func promoteActionReturnsThePromoteLocation() async {
        let (store, model) = await promotableModel()
        store.recordedCalls = []
        let location = model.performCardAction(subjectID: "s-james", actionID: EvidenceSubjectCard.promoteActionID)
        #expect(location?.section == .sources)
        #expect(location?.sourceSurface == .promote)
        #expect(location?.sourceId == sourceID)
        #expect(location?.subjectId == "s-james")
        #expect(location?.subjectTypeKey == "person")
        #expect(location?.ref == "CPR-2AB91")
        #expect(location?.title == "James Robins")
        #expect(location.flatMap(PromoteEntry.init(location:))?.kind == .person)
        #expect(!store.recordedCalls.contains { $0.hasPrefix("promoteSubject") })
    }

    /// Promote all opens the same place with every subject expanded, including
    /// when the card itself is already filed.
    @Test func promoteAllOpensTheExpandedPromotePlace() async {
        let (_, model) = await promotableModel()
        let card = model.promoteLocation(for: "s-james")
        let location = model.promoteAllLocation()
        #expect(location?.promoteAll == true)
        #expect(location?.sourceSurface == .promote)
        #expect(location?.sourceId == sourceID)
        #expect(location?.subjectId == "s-james")
        #expect(location?.subjectTypeKey == "person")
        #expect(location != card)
        #expect(PromoteEntry(location: location!)?.mapAll == true)
        #expect(PromoteEntry(location: location!)?.identityKey == "promote|\(sourceID)|all")
    }

    @Test func promoteAllOpensWhenEveryCardIsAlreadyFiled() async {
        let store = makeStore()
        let subject = CatalogSubject(
            id: "s-james", ref: "CPR-2AB91", sourceID: sourceID, subjectTypeID: personTypeID,
            label: "James Robins", description: ""
        )
        store.subjectsBySource[sourceID] = [subject]
        store.subjectPositionsBySubject["s-james"] = CatalogSubjectPosition(subjectID: "s-james", gridX: 1, gridY: 1)
        let model = makeModel(store: store)
        await model.prepare()
        model.session.setQueryValue(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID),
            value: SourceGraphRows(
                sourceId: sourceID,
                subjects: [subject],
                positions: [CatalogSubjectPosition(subjectID: "s-james", gridX: 1, gridY: 1)],
                memberships: [
                    CatalogSubjectMembership(
                        subjectID: "s-james",
                        claimID: "claim-1",
                        entity: CatalogCanonicalEntity(id: "e-james", ref: "PER-7KD45", subjectTypeID: personTypeID, label: ""),
                        kind: "person"
                    ),
                ]
            )
        )
        #expect(model.promoteLocation(for: "s-james") == nil)
        let location = model.promoteAllLocation()
        #expect(location?.promoteAll == true)
        #expect(location?.subjectId == "s-james")
    }

    /// Promote names the subject by its name Observation's form, falling back
    /// to the label, then the ref.
    @Test func promoteTitleUsesTheNameFormThenTheLabel() async {
        let (_, model) = await promotableModel()
        let base = CatalogSubject(
            id: "s-james", ref: "CPR-2AB91", sourceID: sourceID, subjectTypeID: personTypeID,
            label: "James Robins", description: ""
        )
        func name(_ id: String, form: String, polarity: String = "positive") -> CatalogObservation {
            CatalogObservation(
                id: id, ref: "OBS-\(id)", citationID: "c1", subjectID: "s-james", propertyID: "p-name",
                polarity: polarity, valueText: "", valueInteger: nil, valueDateID: "",
                valueNameID: "n-\(id)", nameForm: form, valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            )
        }
        func title(label: String, observations: [CatalogObservation]) -> String? {
            var subject = base
            subject.label = label
            model.session.setQueryValue(
                CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID),
                value: graphRows(
                    sourceId: sourceID,
                    subjects: [subject],
                    positions: [CatalogSubjectPosition(subjectID: "s-james", gridX: 1, gridY: 1)],
                    observations: observations
                )
            )
            return model.promoteLocation(for: "s-james")?.title
        }
        #expect(title(label: "James", observations: [name("o1", form: "James Henry Robins")]) == "James Henry Robins")
        #expect(title(label: "James", observations: [name("o1", form: "Not James", polarity: "negative")]) == "James")
        #expect(title(label: "James", observations: [name("o1", form: "  ")]) == "James")
        #expect(title(label: "James", observations: []) == "James")
        #expect(title(label: " ", observations: []) == "CPR-2AB91")
    }

    // MARK: Unpromote (S10-44)

    private func handleMembership(_ subjectID: String, ref: String, typeID: String, kind: String) -> CatalogSubjectMembership {
        CatalogSubjectMembership(
            subjectID: subjectID,
            claimID: "claim-\(subjectID)",
            entity: CatalogCanonicalEntity(id: "ent-\(subjectID)", ref: ref, subjectTypeID: typeID, label: ""),
            kind: kind
        )
    }

    private func edgeObservation(_ id: String, bridge: String, key: String, value: String) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: "OBS-\(id)", citationID: "c-1", subjectID: bridge, propertyID: "p-\(key)",
            polarity: "positive", valueText: "", valueInteger: nil, valueDateID: "", valueNameID: "",
            valueSubjectID: value, valueTermID: "", propertyKey: key, propertyLabel: key,
            propertyValueType: "subject"
        )
    }

    /// The household of D-17: James, John, Mary, the 1881 census event; son-of
    /// (James-John), participation (James in the event), and spouse (John-Mary),
    /// every Subject filed except, with `fileSonOf: false`, the son-of bridge.
    private func householdModel(fileSonOf: Bool = true) async -> (FakeStore, EvidenceGraphModel) {
        let store = makeStore()
        func subject(_ id: String, _ prefix: String, _ typeID: String, _ label: String, x: Int64, y: Int64) -> CatalogSubject {
            store.subjectPositionsBySubject[id] = CatalogSubjectPosition(subjectID: id, gridX: x, gridY: y)
            return CatalogSubject(
                id: id, ref: "\(prefix)-\(id)", sourceID: sourceID, subjectTypeID: typeID, label: label, description: ""
            )
        }
        let subjects = [
            subject("s-james", "CPR", personTypeID, "James Robins", x: 0, y: 0),
            subject("s-john", "CPR", personTypeID, "John Robins", x: 3, y: 0),
            subject("s-mary", "CPR", personTypeID, "Mary Smith", x: 6, y: 0),
            subject("s-census", "CEV", "type-event", "1881 census", x: 3, y: 3),
            subject("b1-sonof", "CRL", "type-relationship", "", x: 1, y: 1),
            subject("b2-part", "CPA", "type-participation", "", x: 1, y: 3),
            subject("b3-spouse", "CRL", "type-relationship", "", x: 4, y: 1),
        ]
        store.subjectsBySource[sourceID] = subjects
        let observations = [
            edgeObservation("e1", bridge: "b1-sonof", key: "person", value: "s-james"),
            edgeObservation("e2", bridge: "b1-sonof", key: "related_to", value: "s-john"),
            edgeObservation("e3", bridge: "b2-part", key: "person", value: "s-james"),
            edgeObservation("e4", bridge: "b2-part", key: "event", value: "s-census"),
            edgeObservation("e5", bridge: "b3-spouse", key: "person", value: "s-john"),
            edgeObservation("e6", bridge: "b3-spouse", key: "related_to", value: "s-mary"),
        ]
        store.observationsBySource[sourceID] = observations
        var memberships = [
            handleMembership("s-james", ref: "PER-4M2QX", typeID: personTypeID, kind: "person"),
            handleMembership("s-john", ref: "PER-7KD45", typeID: personTypeID, kind: "person"),
            handleMembership("s-mary", ref: "PER-9TT3B", typeID: personTypeID, kind: "person"),
            handleMembership("s-census", ref: "EVT-2WQ8C", typeID: "type-event", kind: "event"),
            handleMembership("b2-part", ref: "CPA-5RN7K", typeID: "type-participation", kind: "participation"),
            handleMembership("b3-spouse", ref: "REL-6HX2D", typeID: "type-relationship", kind: "relationship"),
        ]
        if fileSonOf {
            memberships.append(handleMembership("b1-sonof", ref: "REL-3AB91", typeID: "type-relationship", kind: "relationship"))
        }
        for m in memberships { store.membershipBySubject[m.subjectID] = m }
        let model = makeModel(store: store)
        await model.prepare()
        model.session.setQueryValue(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID),
            value: graphRows(
                sourceId: sourceID,
                subjects: subjects,
                positions: subjects.compactMap { store.subjectPositionsBySubject[$0.id] },
                observations: observations,
                memberships: memberships
            )
        )
        return (store, model)
    }

    private var sonOfSentence: String {
        L10n.EvidenceGraph.bridgeSummaryRelationshipFallback(person: "James Robins", related: "John Robins")
    }

    private var participationSentence: String {
        L10n.EvidenceGraph.bridgeSummaryParticipationFallback(person: "James Robins", event: "1881 census")
    }

    private func placedSubject(_ model: EvidenceGraphModel, _ id: String) -> SourceGraphPlacedSubject? {
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        let snapshot = SourceGraphSnapshot.build(
            rows: handle?.value ?? SourceGraphRows(sourceId: sourceID),
            types: model.session.queryHandle(CatalogQueryKey.propertiesWorkspace(project: model.session.projectKey)).flatMap { (h: QueryHandle<PropertiesSnapshot>) in h.value?.types } ?? [],
            rules: CatalogConnectRule.productMatrix
        )
        return snapshot.subjects.first { $0.id == id }
    }

    @Test func unpromoteVoiceOverActionFollowsFooterAction() async throws {
        let (_, model) = await householdModel()
        let content = EvidenceCanvasCardContent(model: model, navigation: WorkspaceNavigation(), layouts: EvidenceCardLayoutStore())
        func card(_ placed: SourceGraphPlacedSubject) -> EvidenceCanvasCard {
            EvidenceCanvasCard(
                placed: .subject(placed, nil),
                text: EvidenceCardText(subject: placed),
                state: .init(
                    isSelected: false, isActivated: false, isConnectingFrom: false, isDragging: false,
                    canCite: true, hoveredActionID: nil, pressedActionID: nil
                )
            )
        }
        let james = try #require(placedSubject(model, "s-james"))
        let filed = card(james)
        let names = content.namedActions(for: filed).map(\.name)
        let footer = try #require(filed.text.footerActionName)
        let index = try #require(names.firstIndex(of: footer))
        #expect(names[index + 1] == "Unpromote")
        let action = try #require(content.namedActions(for: filed).first { $0.name == "Unpromote" })
        action.run()
        #expect(model.pendingUnpromote?.id == "s-james")

        var unfiled = james
        unfiled.membership = nil
        #expect(!content.namedActions(for: card(unfiled)).map(\.name).contains("Unpromote"))
    }

    @Test func unpromoteActionOpensConfirmNamingBridges() async {
        let (store, model) = await householdModel()
        store.recordedCalls = []
        let location = model.performCardAction(subjectID: "s-james", actionID: EvidenceSubjectCard.unpromoteActionID)
        #expect(location == nil)
        #expect(!store.recordedCalls.contains { $0.hasPrefix("unpromoteSubject") })
        #expect(model.pendingUnpromote?.name == "James Robins")
        #expect(model.pendingUnpromote?.bridges == [sonOfSentence, participationSentence])
        #expect(model.pendingUnpromote?.error == nil)

        let (_, unfiledSonOf) = await householdModel(fileSonOf: false)
        unfiledSonOf.beginUnpromote(subjectID: "s-james")
        #expect(unfiledSonOf.pendingUnpromote?.bridges == [participationSentence])
    }

    @Test func unpromoteWithNoBridgesHasEmptyDetail() async throws {
        let (store, model) = await promotableModel()
        _ = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "s-james")
        model.session.apply(.promotedSubject(sourceId: sourceID))
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        let loaded = await waitUntil {
            let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
            return handle?.value?.memberships.first?.subjectID == "s-james"
        }
        #expect(loaded)
        model.beginUnpromote(subjectID: "s-james")
        #expect(model.pendingUnpromote?.bridges == [])

        let (_, unfiled) = await promotableModel()
        unfiled.beginUnpromote(subjectID: "s-james")
        #expect(unfiled.pendingUnpromote == nil)
    }

    @Test func confirmUnpromoteRefreshesToPromote() async {
        let (store, model) = await householdModel()
        model.beginUnpromote(subjectID: "s-james")
        model.pendingUnpromote = nil
        #expect(!store.recordedCalls.contains { $0.hasPrefix("unpromoteSubject") })

        model.beginUnpromote(subjectID: "s-james")
        await model.confirmUnpromote()
        #expect(store.recordedCalls.contains("unpromoteSubject id=s-james"))
        #expect(model.pendingUnpromote == nil)
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        let refreshed = await waitUntil {
            let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
            let ids = Set(handle?.value?.memberships.map(\.subjectID) ?? [])
            return !ids.contains("s-james")
        }
        #expect(refreshed)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        let ids = Set(handle?.value?.memberships.map(\.subjectID) ?? [])
        #expect(ids == ["s-john", "s-mary", "s-census", "b3-spouse"])
    }

    @Test func fakeUnpromoteReportsTheEnginesEffects() async throws {
        let (store, _) = await householdModel()
        let removed = ["s-james", "b1-sonof", "b2-part"].compactMap { store.membershipBySubject[$0]?.entity.id }
        #expect(removed.count == 3)
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID, ref: "SRC-1", sourceTypeID: "type-book", title: "Household", description: "",
                hasArtifact: true
            ),
        ]
        let before = store.sourcesByProject
        let effects = try await store.unpromoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "s-james")
        #expect(effects == CatalogWriteEffects(handles: removed, sources: [sourceID], handleScopes: removed))
        #expect(store.sourcesByProject == before)
    }

    @Test func unpromoteFailureStaysInConfirm() async {
        let (store, model) = await householdModel()
        let failure = CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: [])
        store.unpromoteSubjectError = failure
        model.beginUnpromote(subjectID: "s-james")
        await model.confirmUnpromote()
        #expect(model.pendingUnpromote != nil)
        #expect(model.pendingUnpromote?.error == L10n.Errors.message(for: failure))
        #expect(!model.isUnpromoting)
        model.beginUnpromote(subjectID: "s-james")
        #expect(model.pendingUnpromote?.error == nil)
    }

    /// Files s-james the way the Promote place does: write, then invalidate.
    private func promoteJames(_ store: FakeStore, _ model: EvidenceGraphModel) async throws {
        _ = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "s-james")
        model.session.apply(.promotedSubject(sourceId: sourceID))
    }

    /// A promoted person's membership row opens its Person page (S9-09 stub)
    /// and shows the handle's auto-reconciled name.
    @Test func promotedPersonOpensItsPageAndShowsItsName() async {
        let (store, model) = await promotableModel()
        store.observationsBySource[sourceID] = [
            CatalogObservation(
                id: "o-name", ref: "OBS-1", citationID: "c1", subjectID: "s-james", propertyID: "p-name",
                polarity: "positive", valueText: "James Robins", valueInteger: nil, valueDateID: "",
                valueNameID: "n1", nameForm: "James Robins", valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            ),
        ]
        #expect(model.openHandle(subjectID: "s-james") == nil)
        try? await promoteJames(store, model)
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        let loaded = await waitUntil {
            let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
            return handle?.value?.memberships.first?.name != nil
        }
        #expect(loaded)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        let membership = handle?.value?.memberships.first
        #expect(membership.flatMap(EvidenceSubjectCard.membershipName) == "James Robins")

        let location = model.openHandle(subjectID: "s-james")
        #expect(location?.section == .persons)
        #expect(location?.entityId == membership?.entity.id)
        #expect(location?.ref == membership?.entity.ref)
        #expect(location?.title == "James Robins")
    }

    /// A promoted Event opens its page. A promoted Place opens its page. An unpromoted card does not.
    @Test func promotedEventOpensItsPage() async throws {
        let store = makeStore()
        let event = CatalogSubject(
            id: "s-fire", ref: "CEV-1", sourceID: sourceID, subjectTypeID: "type-event",
            label: "Fire at York", description: ""
        )
        let place = CatalogSubject(
            id: "s-york", ref: "CPL-1", sourceID: sourceID, subjectTypeID: "type-place",
            label: "York", description: ""
        )
        store.subjectsBySource[sourceID] = [event, place]
        store.subjectPositionsBySubject["s-fire"] = CatalogSubjectPosition(subjectID: "s-fire", gridX: 1, gridY: 1)
        store.subjectPositionsBySubject["s-york"] = CatalogSubjectPosition(subjectID: "s-york", gridX: 4, gridY: 1)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.openHandle(subjectID: "s-fire") == nil)

        _ = try await store.promoteSubject(
            projectDir: projectDir, userID: "user-1", subjectID: "s-fire",
            entityID: nil, confidenceGradeID: nil, argument: ""
        )
        _ = try await store.promoteSubject(
            projectDir: projectDir, userID: "user-1", subjectID: "s-york",
            entityID: nil, confidenceGradeID: nil, argument: ""
        )
        model.session.apply(.promotedSubject(sourceId: sourceID))
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        let loaded = await waitUntil {
            let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
            return handle?.value?.memberships.contains { $0.subjectID == "s-fire" && $0.kind == "event" } == true
        }
        #expect(loaded)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        let membership = handle?.value?.memberships.first { $0.subjectID == "s-fire" }
        let location = model.openHandle(subjectID: "s-fire")
        #expect(location?.section == .events)
        #expect(location?.entityId == membership?.entity.id)
        #expect(location?.title == "Fire at York")
        let york = handle?.value?.memberships.first { $0.subjectID == "s-york" }
        let yorkPage = model.openHandle(subjectID: "s-york")
        #expect(yorkPage?.section == .places)
        #expect(yorkPage?.entityId == york?.entity.id)
        #expect(yorkPage?.title == "York")
    }

    @Test func promotedCardShowsMembershipAndNoLongerOffersPromote() async throws {
        let (store, model) = await promotableModel()
        try await promoteJames(store, model)
        let key = CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID)
        let reloaded = await waitUntil {
            let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
            return handle?.value?.memberships.first?.subjectID == "s-james"
        }
        #expect(reloaded)
        let handle: QueryHandle<SourceGraphRows>? = model.session.queryHandle(key)
        let snapshot = SourceGraphSnapshot.build(
            rows: handle?.value ?? SourceGraphRows(sourceId: sourceID),
            types: store.subjectTypesByProject[projectDir] ?? [],
            rules: []
        )
        let card = snapshot.subjects.first { $0.id == "s-james" }
        #expect(card?.membership?.entity.ref.hasPrefix("PER-") == true)
        let ids = card.map { EvidenceSubjectCard.actionTargets(for: $0, canCite: true).map(\.id) } ?? []
        #expect(ids.contains(EvidenceSubjectCard.openHandleActionID))
        // Promote is not offered again.
        #expect(model.promoteLocation(for: "s-james") == nil)
    }

    @Test func openHandleHasNoPageYet() async {
        let (_, model) = await promotableModel()
        #expect(model.performCardAction(subjectID: "s-james", actionID: EvidenceSubjectCard.openHandleActionID) == nil)
    }

    // MARK: Open a file (S10-20)

    private func artifact(
        _ id: String,
        label: String,
        ref: String = "",
        fileID: String? = nil,
        relPath: String? = nil,
        mediaType: String = "image/jpeg",
        withFile: Bool = true
    ) -> CatalogArtifact {
        CatalogArtifact(
            id: id,
            ref: ref.isEmpty ? "ART-\(id.uppercased())" : ref,
            sourceID: sourceID,
            fileID: fileID ?? "file-\(id)",
            label: label,
            description: "",
            file: withFile
                ? CatalogFileRef(
                    id: fileID ?? "file-\(id)",
                    relPath: relPath ?? "objects/\(id).jpg",
                    originalFilename: "\(id).jpg",
                    mediaType: mediaType,
                    byteSize: 4_200_000
                )
                : nil
        )
    }

    @Test func fileBearingListKeepsCoverAndDropsFileless() {
        let artifacts = [
            artifact("a", label: "Page 1", ref: "ART-AAAA"),
            artifact("b", label: "Page 2"),
            artifact("c", label: "Loose note", withFile: false),
            artifact("d", label: "", ref: "ART-7Q2KD"),
            artifact("e", label: "No file id", fileID: ""),
            artifact("f", label: "Blank path", relPath: "   "),
        ]
        let rows = EvidenceGraphFiles.fileBearing(artifacts, coverArtifactID: "a")
        #expect(rows.map(\.id) == ["a", "b", "d"])
        #expect(rows.last?.title == "ART-7Q2KD")
        #expect(rows.map(\.titleIsRef) == [false, false, true])
        #expect(rows.map(\.isCover) == [true, false, false])
        #expect(EvidenceGraphFiles.initialID(in: rows, coverArtifactID: "a") == "a")
        // A cover with no file falls back to the first row with one.
        #expect(EvidenceGraphFiles.initialID(in: rows, coverArtifactID: "c") == "a")
        #expect(EvidenceGraphFiles.initialID(in: [], coverArtifactID: nil) == nil)
    }

    @Test func fileKindTokenAndMissingCopy() {
        #expect(EvidenceGraphFiles.kindToken(mediaType: "image/jpeg") == "JPEG")
        #expect(EvidenceGraphFiles.kindToken(mediaType: "application/pdf")
            == L10n.string(L10n.CitationComposer.artifactKindPDF))
        #expect(EvidenceGraphFiles.kindToken(mediaType: "text/plain")
            == L10n.string(L10n.CitationComposer.artifactKindUnknown))
        #expect(
            L10n.EvidenceGraph.fileNotFoundMessage(name: "Page 7", filename: "1881-peel-n-p7.jpg")
                == "Page 7 points at 1881-peel-n-p7.jpg, which is no longer where Provenencia stored it."
        )
    }

    @Test func resolvedFileFallsBackAndSelectFileClearsTheError() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let source = CatalogSource(
            id: sourceID,
            ref: "SRC-1",
            sourceTypeID: "type-book",
            title: "Census",
            description: "",
            coverMode: "artifact",
            primaryArtifactID: "a",
            hasArtifact: true
        )
        model.session.setQueryValue(
            CatalogQueryKey.sourcesList(project: model.session.projectKey),
            value: [source]
        )
        model.session.setQueryValue(
            CatalogQueryKey.sourceWorkspace(project: model.session.projectKey, sourceId: sourceID),
            value: CatalogSourceWorkspace(
                source: source,
                notes: [],
                metadata: [],
                artifacts: [
                    artifact("b", label: "Page 2"),
                    artifact("a", label: "Page 1"),
                    artifact("c", label: "Loose note", withFile: false),
                ],
                credibility: nil
            )
        )
        #expect(model.fileRows.map(\.id) == ["b", "a"])
        #expect(model.resolvedFileArtifactID == "a")
        model.selectedFileArtifactID = "missing"
        #expect(model.resolvedFileArtifactID == "a")
        model.fileOpenError = "File not found"
        model.selectFile("b")
        #expect(model.fileOpenError == nil)
        #expect(model.resolvedFileArtifactID == "b")
        #expect(model.citationCount(for: "b") == 0)
    }

    /// A model whose Source has the given file-bearing Artifacts. The project folder does
    /// not exist, so every object is missing from disk.
    private func fileModel(artifactIDs: [String]) async -> EvidenceGraphModel {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        let source = CatalogSource(
            id: sourceID,
            ref: "SRC-1",
            sourceTypeID: "type-book",
            title: "Census",
            description: "",
            coverMode: "artifact",
            primaryArtifactID: artifactIDs.first ?? "",
            hasArtifact: true
        )
        model.session.setQueryValue(
            CatalogQueryKey.sourcesList(project: model.session.projectKey),
            value: [source]
        )
        model.session.setQueryValue(
            CatalogQueryKey.sourceWorkspace(project: model.session.projectKey, sourceId: sourceID),
            value: CatalogSourceWorkspace(
                source: source,
                notes: [],
                metadata: [],
                artifacts: artifactIDs.map { artifact($0, label: "Page \($0)") },
                credibility: nil
            )
        )
        return model
    }

    @Test func openSelectedFileReportsAMissingFileAndKeepsTheSelection() async {
        let model = await fileModel(artifactIDs: ["a", "b"])
        model.selectFile("b")
        model.openSelectedFile()
        #expect(
            model.fileOpenError
                == L10n.EvidenceGraph.fileNotFoundMessage(name: "Page b", filename: "b.jpg")
        )
        #expect(model.selectedFileArtifactID == "b")
        #expect(model.resolvedFileArtifactID == "b")

        model.selectFile("a")
        #expect(model.fileOpenError == nil)
        #expect(model.selectedFileArtifactID == "a")
    }

    @Test func armingAToolClearsTheMissingFileError() async {
        let model = await fileModel(artifactIDs: ["a"])
        func missingFileError() -> String? {
            model.openSelectedFile()
            return model.fileOpenError
        }

        #expect(missingFileError() != nil)
        model.toggleArm(.person)
        #expect(model.armedKind == .person)
        #expect(model.fileOpenError == nil)

        #expect(missingFileError() != nil)
        model.toggleConnect()
        #expect(model.armedConnect)
        #expect(model.fileOpenError == nil)

        #expect(missingFileError() != nil)
        model.toggleSelect()
        #expect(model.armedSelect)
        #expect(model.fileOpenError == nil)
    }

    // MARK: Box select and group move (S10-20)

    /// Persons `p1` (2, 3), `p2` (4, 3), `p3` (8, 3) and a participation bridge `b1` between them.
    private func groupModel() -> (FakeStore, EvidenceGraphModel) {
        let store = makeStore()
        func person(_ id: String) -> CatalogSubject {
            CatalogSubject(
                id: id, ref: "CPR-\(id)", sourceID: sourceID, subjectTypeID: personTypeID,
                label: id, description: ""
            )
        }
        let bridge = CatalogSubject(
            id: "b1", ref: "CPA-1", sourceID: sourceID, subjectTypeID: "type-participation",
            label: "", description: ""
        )
        let subjects = [person("p1"), person("p2"), person("p3"), bridge]
        let positions = [
            CatalogSubjectPosition(subjectID: "p1", gridX: 2, gridY: 3),
            CatalogSubjectPosition(subjectID: "p2", gridX: 4, gridY: 3),
            CatalogSubjectPosition(subjectID: "p3", gridX: 8, gridY: 3),
            CatalogSubjectPosition(subjectID: "b1", gridX: 3, gridY: 3),
        ]
        store.subjectsBySource[sourceID] = subjects
        for position in positions { store.subjectPositionsBySubject[position.subjectID] = position }
        let model = makeModel(store: store)
        seedFields(model.session, store: store)
        model.session.setQueryValue(
            CatalogQueryKey.sourceGraph(project: model.session.projectKey, sourceId: sourceID),
            value: graphRows(
                sourceId: sourceID,
                subjects: subjects,
                positions: positions,
                types: store.subjectTypesByProject[projectDir] ?? []
            )
        )
        return (store, model)
    }

    @Test func rectangleSelectionKeepsBridges() {
        let (_, model) = groupModel()
        model.replacePrimarySelection(["p1", "b1"])
        #expect(model.selectedSubjectIDs == ["p1", "b1"])
        model.selectSubject(id: "b1")
        #expect(model.selectedSubjectID == "b1")
        model.replacePrimarySelection(["p1", "p2"])
        #expect(model.selectedSubjectID == nil)
        model.armedSelect = true
        model.selectSubject(id: nil)
        #expect(model.armedSelect)
        #expect(model.selectedSubjectIDs.isEmpty)
    }

    @Test func groupDragKeepsTheSet() {
        let (_, model) = groupModel()
        model.replacePrimarySelection(["p1", "p2"])
        #expect(model.beginGroupDrag(anchor: "p1") == ["p1", "p2"])
        #expect(model.groupDragIDs == ["p1", "p2"])
        #expect(model.selectedSubjectIDs == ["p1", "p2"])
        model.replacePrimarySelection(["p1", "b1"])
        #expect(model.beginGroupDrag(anchor: "b1") == ["b1", "p1"])
        #expect(model.groupDragIDs == ["b1", "p1"])
        #expect(model.selectedSubjectIDs == ["b1", "p1"])
        #expect(model.beginGroupDrag(anchor: "p3") == ["p3"])
        #expect(model.groupDragIDs == ["p3"])
        #expect(model.selectedSubjectIDs == ["p3"])
    }

    @Test func arrowMovesEachSelectedSubject() {
        let (_, model) = groupModel()
        model.replacePrimarySelection(["p2", "p1", "b1"])
        model.armedSelect = true
        let effect = model.handleKey(.arrow(dx: 1, dy: 0), hasFocus: false)
        #expect(effect.handled)
        #expect(effect.moves == [
            EvidenceGraphModel.KeyEffect.SubjectMove(subjectID: "p1", fromX: 2, fromY: 3, deltaX: 1, deltaY: 0),
            EvidenceGraphModel.KeyEffect.SubjectMove(subjectID: "p2", fromX: 4, fromY: 3, deltaX: 1, deltaY: 0),
            EvidenceGraphModel.KeyEffect.SubjectMove(subjectID: "b1", fromX: 3, fromY: 3, deltaX: 1, deltaY: 0),
        ])
        // Return and Space do nothing while Select is armed or several are selected.
        #expect(model.handleKey(.return, hasFocus: false) == EvidenceGraphModel.KeyEffect())
        #expect(model.handleKey(.space, hasFocus: false) == EvidenceGraphModel.KeyEffect())
    }

    @Test func moveSubjectsPatchesEachCellAndPersistsInOrder() async {
        let (store, model) = groupModel()
        model.replacePrimarySelection(["p1", "p2"])
        let effect = model.handleKey(.arrow(dx: 1, dy: 0), hasFocus: false)
        model.moveSubjects(effect.moves)
        await model.lastPositionWrite?.value
        #expect(model.gridCell(for: "p1")?.x == 3)
        #expect(model.gridCell(for: "p2")?.x == 5)
        #expect(model.gridCell(for: "p3")?.x == 8)
        #expect(store.subjectPositionsBySubject["p1"]?.gridX == 3)
        #expect(store.subjectPositionsBySubject["p2"]?.gridX == 5)
    }

    @Test func commitGroupDragSnapsEachCard() async {
        let (store, model) = groupModel()
        model.replacePrimarySelection(["p1", "p2"])
        _ = model.beginGroupDrag(anchor: "p1")
        let moved = model.commitGroupDrag(anchor: "p1", documentDelta: CGSize(width: 40, height: 0))
        #expect(moved.map(\.subjectID) == ["p1", "p2"])
        #expect(moved.map(\.gridX) == [3, 5])
        #expect(moved.map(\.gridY) == [3, 3])
        #expect(model.lastPositionWrite != nil)
        #expect(model.groupDragIDs.isEmpty)
        await model.lastPositionWrite?.value
        #expect(store.subjectPositionsBySubject["p1"]?.gridX == 3)
        #expect(store.subjectPositionsBySubject["p2"]?.gridX == 5)
        // b1 was not selected, so the relationship stays on its cell.
        #expect(model.gridCell(for: "b1")?.x == 3)
        #expect(store.subjectPositionsBySubject["b1"]?.gridX == 3)
    }

    @Test func commitGroupDragMovesASelectedBridge() async {
        let (store, model) = groupModel()
        model.replacePrimarySelection(["p1", "b1"])
        _ = model.beginGroupDrag(anchor: "p1")
        let moved = model.commitGroupDrag(anchor: "p1", documentDelta: CGSize(width: 40, height: 0))
        #expect(moved.map(\.subjectID) == ["p1", "b1"])
        #expect(moved.map(\.gridX) == [3, 4])
        await model.lastPositionWrite?.value
        #expect(store.subjectPositionsBySubject["p1"]?.gridX == 3)
        #expect(store.subjectPositionsBySubject["b1"]?.gridX == 4)
        #expect(store.subjectPositionsBySubject["p2"]?.gridX == 4)
    }

    @Test func selectStaysAvailableWithoutAnArtifact() async {
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
        model.session.setQueryValue(
            CatalogQueryKey.sourcesList(project: model.session.projectKey),
            value: store.sourcesByProject[projectDir] ?? []
        )
        await model.prepare()
        #expect(model.canCite == false)
        model.replacePrimarySelection([])
        model.toggleSelect()
        #expect(model.armedSelect)
        #expect(model.inputMode == .selecting)
        model.toggleArm(.person)
        #expect(model.armedKind == nil)
        #expect(model.armedSelect)

        // The cards chosen before Escape stay chosen.
        let (_, grouped) = groupModel()
        grouped.replacePrimarySelection(["p1", "p2"])
        grouped.toggleSelect()
        let effect = grouped.handleKey(.escape, hasFocus: false)
        #expect(effect.handled)
        #expect(grouped.armedSelect == false)
        #expect(grouped.selectedSubjectIDs == ["p1", "p2"])
        #expect(L10n.string(L10n.EvidenceGraph.toolSelectHelp) == "Hold Shift to select")
    }

    @Test func selectAndTheCiteToolsDisarmEachOther() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        model.toggleSelect()
        model.toggleArm(.person)
        #expect(model.armedSelect == false)
        #expect(model.inputMode == .placing(.person))
        model.toggleSelect()
        #expect(model.armedKind == nil)
        #expect(model.inputMode == .selecting)
        model.toggleConnect()
        #expect(model.armedSelect == false)
        #expect(model.inputMode == .connecting)
        model.toggleSelect()
        #expect(model.armedConnect == false)
        // Pressing Select again disarms it. Disarm clears it as well.
        model.toggleSelect()
        #expect(model.inputMode == .idle)
        model.toggleSelect()
        model.disarm()
        #expect(model.armedSelect == false)
    }
}
