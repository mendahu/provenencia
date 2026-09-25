import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CitationComposerModelTests {
    private let projectDir = "/tmp/citation-composer-model.provenencia"
    private let sourceID = "src-1"
    private let subjectID = "sub-person-1"
    private let eventID = "sub-event-1"
    private let placeID = "sub-place-1"
    private let personTypeID = "type-person"
    private let eventTypeID = "type-event"
    private let placeTypeID = "type-place"
    private let participationTypeID = "type-participation"
    private let relationshipTypeID = "type-relationship"
    private let locationTypeID = "type-location"
    private let namePropertyID = "prop-name"
    private let occupationPropertyID = "prop-occupation"
    private let personEdgePropertyID = "prop-person-edge"
    private let eventEdgePropertyID = "prop-event-edge"
    private let placeEdgePropertyID = "prop-place-edge"
    private let relatedToPropertyID = "prop-related-to"
    private let rolePropertyID = "prop-role"
    private let relationshipTypePropertyID = "prop-relationship-type"

    private func makeStore() -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            type(id: personTypeID, key: "person", label: "Person", prefix: "PER"),
            type(id: eventTypeID, key: "event", label: "Event", prefix: "EVT"),
            type(id: placeTypeID, key: "place", label: "Place", prefix: "PLC"),
            type(id: participationTypeID, key: "participation", label: "Participation", prefix: "PTN"),
            type(id: relationshipTypeID, key: "relationship", label: "Relationship", prefix: "REL"),
            type(id: locationTypeID, key: "location", label: "Location", prefix: "LOC"),
        ]
        store.propertiesByProject[projectDir] = [
            property(id: namePropertyID, key: "name", label: "Name", valueType: "name"),
            property(id: occupationPropertyID, key: "occupation", label: "Occupation", valueType: "text"),
            property(id: personEdgePropertyID, key: "person", label: "Person", valueType: "subject"),
            property(id: eventEdgePropertyID, key: "event", label: "Event", valueType: "subject"),
            property(id: placeEdgePropertyID, key: "place", label: "Place", valueType: "subject"),
            property(id: relatedToPropertyID, key: "related_to", label: "Related to", valueType: "subject"),
            property(id: rolePropertyID, key: "role", label: "Role", valueType: "term"),
            property(
                id: relationshipTypePropertyID,
                key: "relationship_type",
                label: "Relationship type",
                valueType: "term"
            ),
        ]
        store.subjectTypeFieldsByType[personTypeID] = [
            field(store.propertiesByProject[projectDir]![0], 0),
            field(store.propertiesByProject[projectDir]![1], 1),
        ]
        store.subjectTypeFieldsByType[eventTypeID] = [
            field(store.propertiesByProject[projectDir]![3], 0),
        ]
        store.subjectTypeFieldsByType[placeTypeID] = [
            field(store.propertiesByProject[projectDir]![4], 0),
        ]
        store.subjectTypeFieldsByType[participationTypeID] = [
            field(store.propertiesByProject[projectDir]![2], 0),
            field(store.propertiesByProject[projectDir]![3], 1),
            field(store.propertiesByProject[projectDir]![6], 2),
        ]
        store.subjectTypeFieldsByType[relationshipTypeID] = [
            field(store.propertiesByProject[projectDir]![2], 0),
            field(store.propertiesByProject[projectDir]![5], 1),
            field(store.propertiesByProject[projectDir]![7], 2),
        ]
        store.subjectTypeFieldsByType[locationTypeID] = [
            field(store.propertiesByProject[projectDir]![3], 0),
            field(store.propertiesByProject[projectDir]![4], 1),
        ]
        store.connectRules = CatalogConnectRule.productMatrix
        store.propertyTermsByProperty[rolePropertyID] = [
            CatalogPropertyTerm(
                id: "term-witness",
                propertyID: rolePropertyID,
                key: "witness",
                origin: "provenencia",
                label: "Witness",
                description: ""
            ),
        ]
        store.propertyTermsByProperty[relationshipTypePropertyID] = [
            CatalogPropertyTerm(
                id: "term-spouse",
                propertyID: relationshipTypePropertyID,
                key: "spouse",
                origin: "provenencia",
                label: "Spouse",
                description: ""
            ),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID, ref: "SRC-1", sourceTypeID: "stype", title: "Census", description: ""
            ),
        ]
        store.subjectsBySource[sourceID] = [
            subject(id: subjectID, typeID: personTypeID, ref: "CPR-1", label: "Margt."),
            subject(id: eventID, typeID: eventTypeID, ref: "CEV-1", label: "Birth"),
            subject(id: placeID, typeID: placeTypeID, ref: "CPL-1", label: "Boston"),
        ]
        store.subjectPositionsBySubject[subjectID] = CatalogSubjectPosition(
            subjectID: subjectID, gridX: 0, gridY: 0
        )
        store.subjectPositionsBySubject[eventID] = CatalogSubjectPosition(
            subjectID: eventID, gridX: 8, gridY: 0
        )
        store.subjectPositionsBySubject[placeID] = CatalogSubjectPosition(
            subjectID: placeID, gridX: 16, gridY: 0
        )
        return store
    }

    private func seedArtifact(_ store: FakeStore) {
        store.artifactsBySource[sourceID] = [
            CatalogArtifact(
                id: "art-0",
                ref: "ART-0",
                sourceID: sourceID,
                fileID: "file-0",
                label: "Scan 1",
                description: "",
                file: CatalogFileRef(
                    id: "file-0",
                    relPath: "objects/file-0.pdf",
                    originalFilename: "scan.pdf",
                    mediaType: "application/pdf",
                    byteSize: 10
                )
            ),
        ]
    }

    private func seedCitation(
        _ store: FakeStore,
        id: String = "cit-1",
        observations: [CatalogObservation] = []
    ) {
        store.citationsByID[id] = CatalogCitation(
            id: id,
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "saved text",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.observationsBySource[sourceID] = observations
    }

    private func makeModel(
        store: FakeStore,
        subjectID: String? = nil,
        citationID: String? = nil,
        connectFrom: String? = nil,
        connectTo: String? = nil,
        bridge: String? = nil
    ) -> CitationComposerModel {
        let entry: CitationComposerEntry
        if let connectFrom, let connectTo, let bridge {
            entry = .connect(
                sourceID: sourceID,
                fromSubjectID: connectFrom,
                toSubjectID: connectTo,
                bridgeTypeKey: bridge
            )
        } else if let citationID {
            entry = .edit(
                sourceID: sourceID,
                subjectID: subjectID ?? self.subjectID,
                citationID: citationID,
                artifactID: "art-0",
                observationID: nil
            )
        } else {
            entry = .addProperty(
                sourceID: sourceID,
                subjectID: subjectID ?? self.subjectID,
                artifactID: "art-0"
            )
        }
        return CitationComposerModel(
            entry: entry,
            session: WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store),
            store: store,
            userID: "user-1"
        )
    }

    @Test func firstRowSaveOnNewCreatesCitationAndExposesRef() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        let rowID = model.observations[0].id
        model.updateObservationProperty(id: rowID, propertyID: occupationPropertyID)
        model.updateObservationText(id: rowID, text: "miller")
        await model.observationRows.commit(rowID: rowID)
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations observations=1") })
        #expect(model.activeCitationID != nil)
        #expect(model.observations[0].persistedRef == "OBS-FAKE1")
        #expect(model.observations[0].state == .saved)
    }

    @Test func draftHasNoRefLoadedHasStoredRef() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.observations[0].persistedRef == "OBS-ABC")
        model.beginAddObservation()
        #expect(model.observations.last?.persistedRef == nil)
    }

    @Test func secondRowSaveCallsAdd() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginAddObservation()
        let rowID = model.observations.last!.id
        model.updateObservationProperty(id: rowID, propertyID: occupationPropertyID)
        model.updateObservationText(id: rowID, text: "weaver")
        store.recordedCalls = []
        await model.observationRows.commit(rowID: rowID)
        #expect(store.recordedCalls.contains { $0.hasPrefix("addObservationsToCitation") })
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations") } == false)
    }

    @Test func editedRowSaveCallsUpdateOnly() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        let rowID = model.observations[0].id
        model.updateObservationText(id: rowID, text: "changed")
        #expect(model.observations[0].state == .edited)
        store.recordedCalls = []
        await model.observationRows.commit(rowID: rowID)
        #expect(store.recordedCalls == ["updateObservation id=obs-1"])
    }

    @Test func revertDraftRemovesAndRevertEditedRestores() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginAddObservation()
        let draftID = model.observations.last!.id
        model.observationRows.revert(rowID: draftID)
        #expect(model.observations.count == 1)
        let rowID = model.observations[0].id
        model.updateObservationText(id: rowID, text: "changed")
        model.observationRows.revert(rowID: rowID)
        #expect(model.observations[0].valueText == "miller")
        #expect(model.observations[0].state == .saved)
    }

    @Test func deleteAsksThenDeletesAndKeepsCitation() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.observationRows.requestDelete(rowID: model.observations[0].id)
        #expect(model.observationRows.pendingDelete != nil)
        await model.observationRows.confirmDelete()
        #expect(store.recordedCalls.contains { $0 == "deleteObservation id=obs-1" })
        #expect(model.observations.isEmpty)
        #expect(model.activeCitationID == "cit-1")
        #expect(store.citationsByID["cit-1"] != nil)
    }

    @Test func incompatibleSubjectClearsPropertyAndDisablesSave() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        store.recordedCalls = []
        model.updateObservationSubject(id: model.observations[0].id, subjectID: eventID)
        #expect(store.recordedCalls.isEmpty)
        #expect(model.observations[0].propertyID.isEmpty)
        #expect(model.observations[0].canSave == false)
        #expect(model.observations[0].propertyError != nil)
    }

    @Test func saveCitationOnNewAndSaved() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.transcription = "note"
        await model.fields.saveCitation()
        #expect(store.recordedCalls.contains { $0 == "createCitationWithObservations observations=0" })
        #expect(model.activeCitationID != nil)
        store.recordedCalls = []
        model.transcription = "changed"
        await model.fields.saveCitation()
        #expect(store.recordedCalls.contains { $0.hasPrefix("updateCitation") })
        #expect(store.recordedCalls.contains { $0.hasPrefix("addObservations") } == false)
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitation") } == false)
    }

    @Test func menusDisabledWithUnsavedWorkEnabledWithPendingConnection() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.identityMenusDisabled == false)
        model.transcription = "dirty"
        #expect(model.identityMenusDisabled)
        let connect = makeModel(
            store: store,
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await connect.prepare()
        #expect(connect.connections.rows.first?.isPending == true)
        #expect(connect.identityMenusDisabled == false)
    }

    @Test func leaveGuardHoldsDirtyEditedAndTouchedNotUntouchedOrEmpty() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.shouldHoldNavigation(.back) == false)
        model.transcription = "dirty"
        #expect(model.shouldHoldNavigation(.back))
        model.keepEditingAfterLeave()
        model.fields.resetBlank()
        model.beginAddObservation()
        #expect(model.shouldHoldNavigation(.back) == false)
        let connect = makeModel(
            store: store,
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await connect.prepare()
        #expect(connect.shouldHoldNavigation(.back) == false)
        connect.connections.applyTerm(
            connectionID: connect.connections.rows[0].id,
            termID: "term-witness"
        )
        #expect(connect.shouldHoldNavigation(.back))
    }

    @Test func pendingConnectionOnNewAndExisting() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let existing = makeModel(store: store, citationID: "cit-1")
        await existing.prepare()
        existing.connections.seedPending(
            fromSubjectID: subjectID,
            toSubjectID: eventID,
            fromLabel: "Margt.",
            toLabel: "Birth",
            bridgeTypeKey: "participation",
            termProperty: store.propertiesByProject[projectDir]!.first { $0.id == rolePropertyID },
            sentence: "pending"
        )
        existing.connections.applyTerm(connectionID: existing.connections.rows[0].id, termID: "term-witness")
        store.recordedCalls = []
        await existing.performSaveConnection()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitedBridge citationID=cit-1 observations=3") })

        let fresh = makeModel(
            store: store,
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await fresh.prepare()
        fresh.connections.applyTerm(connectionID: fresh.connections.rows[0].id, termID: "term-witness")
        store.recordedCalls = []
        await fresh.performSaveConnection()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitedBridge citationID=nil observations=3") })
        #expect(fresh.activeCitationID != nil)
    }

    @Test func switchingCitationKeepsPendingConnection() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(
            store: store,
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await model.prepare()
        #expect(model.connections.rows.contains { $0.isPending })
        model.selectCitation("cit-1")
        await model.awaitIdentitySwitch()
        #expect(model.connections.rows.contains { $0.isPending })
        #expect(model.activeCitationID == "cit-1")
    }

    @Test func pendingLocationCanSaveWithoutTermAndIsNeverTouched() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(
            store: store,
            connectFrom: eventID,
            connectTo: placeID,
            bridge: "location"
        )
        await model.prepare()
        let pending = model.connections.rows[0]
        #expect(pending.canSave)
        #expect(pending.isTouched == false)
        #expect(pending.termProperty == nil)
        #expect(model.shouldHoldNavigation(.back) == false)
        model.transcription = "dirty"
        #expect(model.shouldHoldNavigation(.back))
        store.recordedCalls = []
        await model.performSaveConnection()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitedBridge citationID=nil observations=2") })
    }

    @Test func savedLocationExposesNoRoleCommands() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.connections.replaceSaved([
            ConnectionRow(
                id: UUID(),
                isPending: false,
                fromSubjectID: eventID,
                toSubjectID: placeID,
                fromLabel: "Birth",
                toLabel: "Boston",
                bridgeTypeKey: "location",
                bridgeID: "bridge-loc",
                bridgeRef: "CLO-1",
                sentence: "Birth in Boston",
                termProperty: nil,
                rolePersistedID: nil,
                rolePersistedRef: nil,
                roleTermID: "",
                roleBaselineTermID: nil,
                termTouched: false,
                isSaving: false,
                error: nil
            ),
        ])
        #expect(model.connections.rows[0].canCommitRole == false)
        #expect(model.connections.rows[0].isLocation)
    }

    @Test func participationUsesRoleRelationshipUsesRelationship() {
        #expect(
            ConnectionRow(
                id: UUID(), isPending: true, fromSubjectID: "", toSubjectID: "", fromLabel: "",
                toLabel: "", bridgeTypeKey: "participation", sentence: "",
                termProperty: CatalogProperty(
                    id: rolePropertyID, key: "role", origin: "provenencia", label: "Role",
                    description: "", valueType: "term"
                ),
                roleTermID: "", termTouched: false, isSaving: false
            ).termFieldLabel == String(localized: L10n.CitationComposer.connectionRole)
        )
        #expect(
            ConnectionRow(
                id: UUID(), isPending: true, fromSubjectID: "", toSubjectID: "", fromLabel: "",
                toLabel: "", bridgeTypeKey: "relationship", sentence: "",
                termProperty: CatalogProperty(
                    id: relationshipTypePropertyID, key: "relationship_type", origin: "provenencia",
                    label: "Type", description: "", valueType: "term"
                ),
                roleTermID: "", termTouched: false, isSaving: false
            ).termFieldLabel == String(localized: L10n.CitationComposer.connectionRelationship)
        )
    }

    @Test func polarityMenuTitleFollowsPolarity() {
        var row = ObservationRow.draft(subjectID: subjectID)
        #expect(row.polarityMenuTitle == String(localized: L10n.CitationComposer.negateObservation))
        row.polarity = ObservationPolarity.negative.rawValue
        #expect(row.polarityMenuTitle == String(localized: L10n.CitationComposer.affirmObservation))
    }

    @Test func nameDialogApplyOnlyUpdatesRow() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(
            store,
            observations: [
                CatalogObservation(
                    id: "obs-name",
                    ref: "OBS-N",
                    citationID: "cit-1",
                    subjectID: subjectID,
                    propertyID: namePropertyID,
                    polarity: ObservationPolarity.positive.rawValue,
                    valueText: "Ada",
                    valueInteger: nil,
                    valueDateID: "",
                    valueNameID: "",
                    nameForm: "Ada",
                    valueSubjectID: "",
                    valueTermID: "",
                    propertyKey: "name",
                    propertyLabel: "Name",
                    propertyValueType: "name"
                ),
            ]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginEditObservation(model.observations[0])
        var draft = model.observationDialog!
        draft.nameDraft = NameValueDraft(form: "Ada Lovelace", parts: [], formTouched: true)
        model.updateObservationDialog(draft)
        store.recordedCalls = []
        model.confirmObservationDialog()
        #expect(store.recordedCalls.isEmpty)
        #expect(model.observations[0].state == .edited)
    }

    @Test func loadingBridgeEdgesYieldsOneConnectionAndKeepsExtraRole() async {
        let store = makeStore()
        seedArtifact(store)
        let bridge = subject(id: "bridge-1", typeID: participationTypeID, ref: "PTN-1", label: "")
        store.subjectsBySource[sourceID]?.append(bridge)
        store.subjectPositionsBySubject[bridge.id] = CatalogSubjectPosition(
            subjectID: bridge.id, gridX: 4, gridY: 0
        )
        seedCitation(
            store,
            observations: [
                edgeObservation(id: "e1", ref: "OBS-E1", subjectID: bridge.id, propertyID: personEdgePropertyID),
                edgeObservation(id: "e2", ref: "OBS-E2", subjectID: bridge.id, propertyID: eventEdgePropertyID),
                roleObservation(id: "r1", ref: "OBS-R1", subjectID: bridge.id, termID: "term-witness"),
                roleObservation(id: "r2", ref: "OBS-R2", subjectID: bridge.id, termID: "term-witness"),
            ]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.connections.rows.count == 1)
        #expect(model.connections.rows[0].bridgeRef == "PTN-1")
        #expect(model.observations.count == 1)
        #expect(model.observations[0].persistedRef == "OBS-R2")
        #expect(model.observations.contains { $0.propertyID == personEdgePropertyID } == false)
    }

    @Test func saveConnectionDoesNotAppendObservationRows() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(
            store: store,
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await model.prepare()
        model.connections.applyTerm(connectionID: model.connections.rows[0].id, termID: "term-witness")
        let before = model.observations.count
        await model.performSaveConnection()
        #expect(model.connections.rows[0].isPending == false)
        #expect(model.observations.count == before)
    }

    @Test func roleChangeCallsUpdateOrAddAndHasNoDelete() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        let connectionID = UUID()
        model.connections.replaceSaved([
            ConnectionRow(
                id: connectionID,
                isPending: false,
                fromSubjectID: subjectID,
                toSubjectID: eventID,
                fromLabel: "Margt.",
                toLabel: "Birth",
                bridgeTypeKey: "participation",
                bridgeID: "bridge-1",
                bridgeRef: "PTN-1",
                sentence: "sentence",
                termProperty: store.propertiesByProject[projectDir]!.first { $0.id == rolePropertyID },
                rolePersistedID: "role-1",
                rolePersistedRef: "OBS-ROLE",
                roleTermID: "term-witness",
                roleBaselineTermID: "term-witness",
                termTouched: false,
                isSaving: false,
                error: nil
            ),
        ])
        store.subjectsBySource[sourceID]?.append(
            subject(id: "bridge-1", typeID: participationTypeID, ref: "PTN-1", label: "")
        )
        store.observationsBySource[sourceID, default: []].append(
            roleObservation(id: "role-1", ref: "OBS-ROLE", subjectID: "bridge-1", termID: "term-witness")
        )
        model.connections.applyTerm(connectionID: connectionID, termID: "term-witness")
        #expect(model.connections.rows[0].canCommitRole == false)
        model.connections.applyTerm(connectionID: connectionID, termID: "term-other")
        store.recordedCalls = []
        await model.performCommitRole(connectionID: connectionID)
        #expect(store.recordedCalls.contains { $0.hasPrefix("updateObservation") })
    }

    @Test func newPersonUsesComposerSlotAndReusesLandingColumn() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        let rowID = model.observations[0].id
        model.beginNewSubject(typeKey: "person", rowID: rowID)
        model.newSubjectDraft?.label = "Ada"
        store.recordedCalls = []
        await model.confirmNewSubject()
        #expect(model.landingColumn != nil)
        #expect(store.recordedCalls.contains { $0.hasPrefix("createSubject") && $0.contains("placement=") })
        let firstColumn = model.landingColumn
        model.beginAddObservation()
        let secondID = model.observations.last!.id
        model.beginNewSubject(typeKey: "person", rowID: secondID)
        model.newSubjectDraft?.label = "Ben"
        await model.confirmNewSubject()
        #expect(model.landingColumn == firstColumn)
    }

    @Test func newSubjectDisabledWhileGraphReloading() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.session.apply(.mutatedSourceGraph(sourceId: sourceID))
        #expect(model.isGraphReloading)
        model.beginAddObservation()
        model.beginNewSubject(typeKey: "person", rowID: model.observations[0].id)
        #expect(model.newSubjectDraft == nil)
    }

    @Test func twoQuickCitationSwitchesApplyOnlyLast() async {
        let store = makeStore()
        seedArtifact(store)
        store.citationsByID["cit-a"] = CatalogCitation(
            id: "cit-a", ref: "CIT-A", artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "A", description: "", transcriptionUncertain: false, transcriptionNote: ""
        )
        store.citationsByID["cit-b"] = CatalogCitation(
            id: "cit-b", ref: "CIT-B", artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "B", description: "", transcriptionUncertain: false, transcriptionNote: ""
        )
        let model = makeModel(store: store)
        await model.prepare()
        model.selectCitation("cit-a")
        model.selectCitation("cit-b")
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == "cit-b")
        #expect(model.transcription == "B")
    }

    @Test func connectEntryIgnoresLegacyTermAndGrid() {
        let location = WorkspaceLocation(
            section: .sources,
            sourceId: sourceID,
            connectFromSubjectId: subjectID,
            connectToSubjectId: eventID,
            connectBridgeTypeKey: "participation",
            connectDisambiguationTermId: "term-witness",
            connectGridX: 2,
            connectGridY: 3,
            sourceSurface: .citationComposer
        )
        let entry = CitationComposerEntry(location: location)
        #expect(entry == .connect(
            sourceID: sourceID,
            fromSubjectID: subjectID,
            toSubjectID: eventID,
            bridgeTypeKey: "participation"
        ))
    }

    private func type(id: String, key: String, label: String, prefix: String) -> CatalogSubjectType {
        CatalogSubjectType(
            id: id, key: key, origin: "provenencia", label: label, description: "",
            refPrefix: prefix, candidateRefPrefix: "C" + prefix
        )
    }

    private func property(id: String, key: String, label: String, valueType: String) -> CatalogProperty {
        CatalogProperty(
            id: id, key: key, origin: "provenencia", label: label, description: "", valueType: valueType
        )
    }

    private func field(_ property: CatalogProperty, _ order: Int) -> CatalogSubjectTypeField {
        CatalogSubjectTypeField(property: property, sortOrder: order, locked: false)
    }

    private func subject(id: String, typeID: String, ref: String, label: String) -> CatalogSubject {
        CatalogSubject(
            id: id, ref: ref, sourceID: sourceID, subjectTypeID: typeID, label: label, description: ""
        )
    }

    private func occupationObservation(id: String, ref: String, citationID: String) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: ref, citationID: citationID, subjectID: subjectID,
            propertyID: occupationPropertyID, polarity: ObservationPolarity.positive.rawValue,
            valueText: "miller", valueInteger: nil, valueDateID: "", valueNameID: "", valueSubjectID: "",
            valueTermID: "", propertyKey: "occupation", propertyLabel: "Occupation",
            propertyValueType: "text"
        )
    }

    private func edgeObservation(
        id: String, ref: String, subjectID: String, propertyID: String
    ) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: ref, citationID: "cit-1", subjectID: subjectID, propertyID: propertyID,
            polarity: ObservationPolarity.positive.rawValue, valueText: "", valueInteger: nil,
            valueDateID: "", valueNameID: "", valueSubjectID: self.subjectID, valueTermID: "",
            propertyKey: storeKey(propertyID), propertyLabel: "edge", propertyValueType: "subject"
        )
    }

    private func roleObservation(id: String, ref: String, subjectID: String, termID: String) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: ref, citationID: "cit-1", subjectID: subjectID, propertyID: rolePropertyID,
            polarity: ObservationPolarity.positive.rawValue, valueText: "Witness", valueInteger: nil,
            valueDateID: "", valueNameID: "", valueSubjectID: "", valueTermID: termID, propertyKey: "role",
            propertyLabel: "Role", propertyValueType: "term"
        )
    }

    private func storeKey(_ propertyID: String) -> String {
        switch propertyID {
        case personEdgePropertyID: return "person"
        case eventEdgePropertyID: return "event"
        case placeEdgePropertyID: return "place"
        default: return ""
        }
    }
}
