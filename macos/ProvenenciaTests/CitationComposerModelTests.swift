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
        store.subjectTypePropertiesByType[personTypeID] = [
            field(store.propertiesByProject[projectDir]![0], 0),
            field(store.propertiesByProject[projectDir]![1], 1),
        ]
        store.subjectTypePropertiesByType[eventTypeID] = [
            field(store.propertiesByProject[projectDir]![3], 0),
        ]
        store.subjectTypePropertiesByType[placeTypeID] = [
            field(store.propertiesByProject[projectDir]![4], 0),
        ]
        store.subjectTypePropertiesByType[participationTypeID] = [
            field(store.propertiesByProject[projectDir]![2], 0),
            field(store.propertiesByProject[projectDir]![3], 1),
            field(store.propertiesByProject[projectDir]![6], 2),
        ]
        store.subjectTypePropertiesByType[relationshipTypeID] = [
            field(store.propertiesByProject[projectDir]![2], 0),
            field(store.propertiesByProject[projectDir]![5], 1),
            field(store.propertiesByProject[projectDir]![7], 2),
        ]
        store.subjectTypePropertiesByType[locationTypeID] = [
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
        locatorJSON: String = #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
        observations: [CatalogObservation] = []
    ) {
        store.citationsByID[id] = CatalogCitation(
            id: id,
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: locatorJSON,
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
        observationID: String? = nil,
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
                observationID: observationID
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

    @Test func addPropertyEntryOpensTheObservationWithThatSubject() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        guard case .property(let draft) = model.authoring else {
            Issue.record("expected the observation dialog")
            return
        }
        #expect(draft.values.subjectID == subjectID)
        #expect(CitationAuthoringNavigation.initialPropertyFocus(subjectID: draft.values.subjectID) == .property)
        #expect(model.presented == .authoring)
        #expect(model.subjectLabel == "Margt.")
    }

    @Test func editPropertyEntryOpensThatObservation() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1", observationID: "obs-1")
        await model.prepare()
        guard case .property(let values, let persistedID, _) = model.editSession?.kind else {
            Issue.record("expected the edit dialog")
            return
        }
        #expect(persistedID == "obs-1")
        #expect(values.subjectID == subjectID)
        #expect(values.propertyID == occupationPropertyID)
        #expect(values.valueText == "miller")
        #expect(model.activeCitationID == "cit-1")
        #expect(model.editSession?.isDirty == false)
    }

    @Test func firstRowSaveOnNewCreatesCitationAndExposesRef() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        await model.saveAuthoring()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations observations=0") })
        #expect(store.recordedCalls.contains { $0.hasPrefix("addObservationsToCitation") })
        #expect(model.activeCitationID != nil)
        #expect(model.observations.contains { $0.persistedRef == "OBS-FAKE1" && $0.state == .saved })
    }

    @Test func draftHasNoRefLoadedHasStoredRef() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.observations[0].persistedRef == "OBS-ABC")
        model.beginPropertyDraft(subjectID: subjectID)
        #expect(model.observations.count == 1)
        if case .property(let draft) = model.authoring {
            #expect(draft.values.propertyID.isEmpty)
        } else {
            Issue.record("expected a property draft")
        }
    }

    @Test func secondRowSaveCallsAdd() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "weaver"
        }
        store.recordedCalls = []
        await model.saveAuthoring()
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
        model.observationRows.applyText(rowID: rowID, text: "changed")
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
        model.beginPropertyDraft(subjectID: subjectID)
        model.discardAuthoring()
        #expect(model.observations.count == 1)
        let rowID = model.observations[0].id
        model.observationRows.applyText(rowID: rowID, text: "changed")
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
        await model.askDeleteObservation(rowID: model.observations[0].id)
        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.kind == "observation")
        await model.confirmPendingImpact()
        #expect(store.recordedCalls.contains { $0 == "deleteObservation id=obs-1" })
        #expect(model.observations.isEmpty)
        #expect(model.activeCitationID == "cit-1")
        #expect(store.citationsByID["cit-1"] != nil)
    }

    @Test func emptyCitationDeleteConfirmsAndBecomesNew() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        await model.askDeleteCitation()
        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.kind == "citation")
        #expect(model.pendingImpact?.target.ref == "CIT-1")
        await model.confirmPendingImpact()
        #expect(store.recordedCalls.contains { $0 == "deleteCitation id=cit-1" })
        #expect(model.activeCitationID == nil)
        #expect(model.isEditingExisting == false)
        #expect(store.citationsByID["cit-1"] == nil)
    }

    @Test func citationWithRowsNoticesAndDoesNotDelete() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(
            store,
            observations: [
                occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1"),
                occupationObservation(id: "obs-2", ref: "OBS-DEF", citationID: "cit-1"),
            ]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        await model.askDeleteCitation()
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.groups.first?.via == "observations.citation_id")
        #expect(model.pendingImpact?.report.groups.first?.total == 2)
        store.recordedCalls = []
        await model.confirmPendingImpact()
        #expect(store.recordedCalls.contains { $0.hasPrefix("deleteCitation") } == false)
        #expect(store.citationsByID["cit-1"] != nil)
    }

    @Test func connectionOnlyCitationNoticesEdges() async {
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
            ]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        await model.askDeleteCitation()
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.groups.first?.via == "observations.citation_id")
        #expect(model.pendingImpact?.report.groups.first?.total == 3)
        store.recordedCalls = []
        await model.confirmPendingImpact()
        #expect(store.recordedCalls.contains { $0.hasPrefix("deleteCitation") } == false)
    }

    @Test func lastRowThenDeleteCitation() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        await model.askDeleteObservation(rowID: model.observations[0].id)
        await model.confirmPendingImpact()
        #expect(model.activeCitationID == "cit-1")
        await model.askDeleteCitation()
        #expect(model.pendingImpact?.report.allowed == true)
        await model.confirmPendingImpact()
        #expect(store.recordedCalls.contains { $0 == "deleteCitation id=cit-1" })
        #expect(model.activeCitationID == nil)
    }

    @Test func savedConnectionRowHasNoDeletePath() async {
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
            ]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.connections.rows.count == 1)
        #expect(model.observations.isEmpty)
        await model.askDeleteObservation(rowID: UUID())
        #expect(model.pendingImpact == nil)
    }

    @Test func ghostArtifactResetsToANewCitation() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        if var citation = store.citationsByID["cit-1"] {
            citation.artifactID = "art-gone"
            store.citationsByID["cit-1"] = citation
        }
        let ghostArt = makeModel(store: store, citationID: "cit-1")
        await ghostArt.prepare()
        #expect(ghostArt.phase == .compose)
        #expect(ghostArt.activeCitationID == nil)
        store.recordedCalls = []
        await ghostArt.fields.saveCitation()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations") })
        #expect(store.recordedCalls.contains { $0.hasPrefix("updateCitation") } == false)
    }

    @Test func incompatibleSubjectClearsPropertyAndDisablesSave() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        store.recordedCalls = []
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        model.selectAuthoringSubject(eventID, field: .propertySubject)
        guard case .property(let draft) = model.authoring else {
            Issue.record("expected a property draft")
            return
        }
        #expect(draft.values.subjectID == eventID)
        #expect(draft.values.propertyID.isEmpty)
        #expect(draft.values.valueText.isEmpty)
        #expect(model.authoringCanConfirm == false)
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

    @Test func concurrentFirstWritesShareOneCitation() async {
        let store = makeStore()
        store.createCitationDelayNanoseconds = 20_000_000
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        await model.saveAuthoring()
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "weaver"
        }
        await model.saveAuthoring()
        let creates = store.recordedCalls.filter { $0.hasPrefix("createCitationWithObservations") }
        #expect(creates.count == 1)
        #expect(creates.first == "createCitationWithObservations observations=0")
        let adds = store.recordedCalls.filter { $0.hasPrefix("addObservationsToCitation") }
        #expect(adds.count == 2)
        #expect(model.activeCitationID != nil)
        #expect(model.observations.filter { $0.state == .saved }.count == 2)
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
        if case .bridge(let draft) = connect.authoring {
            #expect(draft.termID.isEmpty)
            #expect(draft.holdsLeave)
        } else {
            Issue.record("Connect opens a bridge draft")
        }
        #expect(connect.identityMenusDisabled)
        connect.mutateBridge { $0.termID = "term-witness" }
        #expect(connect.identityMenusDisabled)
        #expect(connect.shouldHoldLeave)
        #expect(connect.shouldHoldNavigation(.back))
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
        #expect(connect.shouldHoldNavigation(.back))
        connect.discardAuthoring()
        #expect(connect.shouldHoldNavigation(.back) == false)
    }

    @Test func editPrepareKeepsRegionLocatorAndIsNotDirty() async {
        let store = makeStore()
        seedArtifact(store)
        let locatorJSON = """
        {"version":1,"selectors":[{"type":"artifact"},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.4},{"x":0.1,"y":0.4}]}]}
        """
        seedCitation(
            store,
            locatorJSON: locatorJSON,
            observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.fields.locator.hasRegion)
        #expect(model.fields.locator.region?.points.count == 4)
        #expect(model.fields.isDirty == false)
        #expect(model.hasUnsavedDocumentWork == false)
        #expect(model.shouldHoldNavigation(.back) == false)
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
        await existing.connections.saveConnection()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitedBridge citationID=cit-1 observations=3") })

        let fresh = makeModel(
            store: store,
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await fresh.prepare()
        fresh.mutateBridge { $0.termID = "term-witness" }
        store.recordedCalls = []
        await fresh.saveAuthoring()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations observations=0") })
        #expect(store.recordedCalls.contains { call in
            call.hasPrefix("createCitedBridge citationID=") && call.contains("observations=3")
                && !call.contains("citationID=nil")
        })
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
        if case .bridge = model.authoring {} else {
            Issue.record("Connect opens one bridge draft")
        }
        model.selectCitation("cit-1")
        await model.awaitIdentitySwitch()
        #expect(model.pendingLeave == nil)
        #expect(model.activeCitationID == "cit-1")
        guard case .bridge(let draft) = model.authoring else {
            Issue.record("the bridge draft stays open")
            return
        }
        #expect(draft.endA == subjectID)
        #expect(draft.endB == eventID)
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
        guard case .bridge(let draft) = model.authoring else {
            Issue.record("location connect opens a bridge draft")
            return
        }
        let pair = CitationComposerAuthoring.resolvePair(
            endA: draft.endA, endB: draft.endB, termID: "", vocabulary: model.vocabulary
        )
        guard case .ready(let ready) = pair else {
            Issue.record("event and place should be ready")
            return
        }
        #expect(ready.termProperty == nil)
        #expect(model.shouldHoldNavigation(.back))
        model.transcription = "dirty"
        #expect(model.shouldHoldNavigation(.back))
        store.recordedCalls = []
        await model.saveAuthoring()
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations observations=0") })
        #expect(store.recordedCalls.contains { call in
            call.hasPrefix("createCitedBridge citationID=") && call.contains("observations=2")
                && !call.contains("citationID=nil")
        })
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
            ).termFieldLabel == L10n.string(L10n.CitationComposer.connectionRole)
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
            ).termFieldLabel == L10n.string(L10n.CitationComposer.connectionRelationship)
        )
    }

    @Test func polarityMenuTitleFollowsPolarity() {
        var row = ObservationRow.draft(subjectID: subjectID)
        #expect(!row.isNegated)
        #expect(row.polarityMenuTitle == L10n.string(L10n.CitationComposer.negateObservation))
        row.polarity = ObservationPolarity.negative.rawValue
        #expect(row.isNegated)
        #expect(row.polarityMenuTitle == L10n.string(L10n.CitationComposer.affirmObservation))
    }

    @Test func nameRowUsesDisplayOrderNotTheStoredForm() {
        let fields = CitationObservationValue.Fields(
            nameDraft: NameValueDraft(
                form: "Gumtree, Jeremiah Arthurt",
                parts: [
                    CatalogNameValuePart(value: "Jeremiah", type: "given"),
                    CatalogNameValuePart(value: "Arthur", type: "given"),
                    CatalogNameValuePart(value: "Gumtree", type: "surname"),
                ]
            )
        )
        let orders = NameFormatOrders(
            natural: [
                NameFormatPart(partType: "given", trailingSeparator: ""),
                NameFormatPart(partType: "surname", trailingSeparator: ""),
            ],
            sorted: [
                NameFormatPart(partType: "surname", trailingSeparator: ", "),
                NameFormatPart(partType: "given", trailingSeparator: ""),
            ]
        )
        #expect(
            CitationObservationValue.summary(
                valueType: PropertyValueType.name.rawValue,
                fields: fields,
                termLabel: nil,
                orders: orders
            ) == "Jeremiah Arthur Gumtree"
        )
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
        model.mutateBridge { $0.termID = "term-witness" }
        let before = model.observations.count
        await model.saveAuthoring()
        #expect(model.observations.filter { $0.persistedID != nil }.count >= before)
        #expect(model.observations.count == before)
        #expect(model.vocabulary.graphSubjects.contains { $0.typeKey == "participation" })
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
        await model.connections.commitRole(connectionID: connectionID)
        #expect(store.recordedCalls.contains { $0.hasPrefix("updateObservation") })
    }

    @Test func newPersonUsesComposerSlotAndReusesLandingColumn() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: "")
        model.beginNewSubject(typeKey: "person", field: .propertySubject)
        model.newSubjectDraft?.label = "Ada"
        store.recordedCalls = []
        await model.confirmNewSubject()
        #expect(model.landingColumn != nil)
        #expect(store.recordedCalls.contains { $0.hasPrefix("createSubject") && $0.contains("placement=") })
        let firstColumn = model.landingColumn
        model.beginPropertyDraft(subjectID: "")
        model.beginNewSubject(typeKey: "person", field: .propertySubject)
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
        model.beginNewSubject(typeKey: "person", field: .propertySubject)
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

    @Test func modalCitationSwitchUpdatesTheOpenCitationAndKeepsTheDraft() async {
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
            transcription: "B", description: "Schedule B", transcriptionUncertain: false, transcriptionNote: ""
        )
        let model = makeModel(store: store)
        await model.prepare()
        model.mutateProperty { $0.values.propertyID = occupationPropertyID }
        model.setAuthoringCitation("cit-b")
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == "cit-b")
        #expect(model.authoringCitationID == "cit-b")
        #expect(model.transcription == "B")
        #expect(model.citationDescription == "Schedule B")
        guard case .property(let draft) = model.authoring else {
            Issue.record("the observation dialog stays open")
            return
        }
        #expect(draft.values.subjectID == subjectID)
        #expect(draft.values.propertyID == occupationPropertyID)
        model.selectCitation("cit-a")
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == "cit-a")
        #expect(model.authoringCitationID == "cit-a")
        #expect(model.transcription == "A")
        guard case .property(let still) = model.authoring else {
            Issue.record("the page citation menu keeps the dialog open")
            return
        }
        #expect(still.values.subjectID == subjectID)
        #expect(store.recordedCalls.contains { $0.hasPrefix("getCitation") } == false)
    }

    @Test func dirtyCitationSwitchAsksAndKeepsTheDraft() async {
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
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty { $0.values.propertyID = occupationPropertyID }
        model.selectCitation("cit-b")
        await model.awaitIdentitySwitch()
        model.transcription = "unsaved note"
        model.selectCitation("cit-a")
        #expect(model.pendingLeave?.keepDraft == true)
        #expect(model.activeCitationID == "cit-b")
        #expect(model.transcription == "unsaved note")
        model.keepEditingAfterLeave()
        #expect(model.pendingLeave == nil)
        #expect(model.activeCitationID == "cit-b")
        #expect(model.transcription == "unsaved note")
        model.selectCitation("cit-a")
        model.discardLeaveChanges()
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == "cit-a")
        #expect(model.transcription == "A")
        #expect(model.fields.isDirty == false)
        guard case .property(let draft) = model.authoring else {
            Issue.record("discarding the citation edit keeps the observation draft")
            return
        }
        #expect(draft.values.subjectID == subjectID)
        #expect(draft.values.propertyID == occupationPropertyID)
    }

    @Test func savePropertyKeepsDirtyTranscription() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1", observationID: "")
        await model.prepare()
        #expect(model.editSession == nil)
        model.transcription = "edited before save"
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        await model.saveAuthoring()
        #expect(model.transcription == "edited before save")
        #expect(model.fields.isDirty)
        #expect(model.citationDescription == "")
    }

    @Test func connectEntryIgnoresLegacyTermAndGrid() throws {
        let json = """
        {
          "section":"sources",
          "sourceId":"\(sourceID)",
          "connectFromSubjectId":"\(subjectID)",
          "connectToSubjectId":"\(eventID)",
          "connectBridgeTypeKey":"participation",
          "connectDisambiguationTermId":"term-witness",
          "connectGridX":2,
          "connectGridY":3,
          "sourceSurface":"citationComposer"
        }
        """
        let location = try JSONDecoder().decode(WorkspaceLocation.self, from: Data(json.utf8))
        let entry = CitationComposerEntry(location: location)
        #expect(entry == .connect(
            sourceID: sourceID,
            fromSubjectID: subjectID,
            toSubjectID: eventID,
            bridgeTypeKey: "participation"
        ))
    }

    @Test func missingCitationShowsNeutralLoadFailed() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store, citationID: "cit-gone")
        await model.prepare()
        #expect(model.phase == .loadFailed)
        #expect(model.loadFailureIsMissingCitation)
        #expect(model.loadError == L10n.string(L10n.CitationComposer.citationMissing))
    }

    @Test func missingSubjectFallsBackWithoutCompose() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store, subjectID: "sub-gone")
        await model.prepare()
        #expect(model.phase == .subjectMissing)
    }

    @Test func newSubjectOptionsComeFromPlaceableTypes() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        let values = Set(model.subjectOptions.map(\.value))
        #expect(values.contains(CitationComposerModel.newSubjectPrefix + "person"))
        #expect(values.contains(CitationComposerModel.newSubjectPrefix + "event"))
        #expect(values.contains(CitationComposerModel.newSubjectPrefix + "place"))
        #expect(!values.contains(where: { $0.hasSuffix("participation") }))
        #expect(!values.contains(where: { $0.hasSuffix("relationship") }))
    }

    @Test func confirmCustomTermWritesTheTermAndClosesTheModal() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty { $0.values.propertyID = occupationPropertyID }
        model.beginAddCustomTerm(propertyID: occupationPropertyID)
        model.customTermDraft?.label = "Weaver"
        await model.confirmCustomTerm()
        #expect(model.customTermDraft == nil)
        if case .property(let draft) = model.authoring {
            #expect(draft.values.valueTermID.isEmpty == false)
        } else {
            Issue.record("the term should land on the draft")
        }
        #expect(store.propertyTermsByProperty[occupationPropertyID]?.contains { $0.label == "Weaver" } == true)
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

    private func field(_ property: CatalogProperty, _ order: Int) -> CatalogSubjectTypeProperty {
        CatalogSubjectTypeProperty(property: property, sortOrder: order, locked: false)
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

    @Test func propertyAndBridgeShareTheAuthoringBox() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        if case .property = model.authoring {} else {
            Issue.record("expected a property draft")
        }
        model.beginBridgeDraft(endA: subjectID, endB: eventID, termID: "", origin: .blank)
        if case .bridge(let draft) = model.authoring {
            #expect(draft.endA == subjectID)
            #expect(draft.termID.isEmpty)
        } else {
            Issue.record("expected a bridge draft")
        }
        if case .property = model.authoring {
            Issue.record("property draft should have been replaced")
        }
    }

    @Test func savedBridgeIsListedUnderBothEnds() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        let bridgeID = "018f0000-0000-7000-8000-000000000001"
        model.connections.replaceSaved([
            ConnectionRow(
                id: UUID(),
                isPending: false,
                fromSubjectID: subjectID,
                toSubjectID: eventID,
                fromLabel: "Margt.",
                toLabel: "Birth",
                bridgeTypeKey: "participation",
                bridgeID: bridgeID,
                bridgeRef: "CPT-1",
                sentence: "",
                termProperty: store.propertiesByProject[projectDir]!.first { $0.id == rolePropertyID },
                rolePersistedID: "obs-role",
                rolePersistedRef: "OBS-R",
                roleTermID: "term-witness",
                roleBaselineTermID: "term-witness",
                termTouched: false,
                isSaving: false
            ),
        ])
        let groups = model.subjectGroups
        let person = groups.first { $0.subjectID == subjectID }
        let event = groups.first { $0.subjectID == eventID }
        #expect(person?.lines.contains { $0.bridgeID == bridgeID } == true)
        #expect(event?.lines.contains { $0.bridgeID == bridgeID } == true)
        #expect(person?.lines.first?.body.contains("Birth") == true)
        model.beginEdit(line: person!.lines[0])
        if case .bridge(_, _, let termID, let editedID, _, _) = model.editSession?.kind {
            #expect(editedID == bridgeID)
            #expect(termID == "term-witness")
        } else {
            Issue.record("edit should open the one bridge")
        }
    }

    @Test func duplicateDoesNotWriteAndConfirmsADirtyDraft() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(
            store,
            observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        let line = model.subjectGroups.flatMap(\.lines).first { $0.observationID == "obs-1" }
        #expect(line != nil)
        store.recordedCalls = []
        model.duplicate(line: line!)
        #expect(store.recordedCalls.isEmpty)
        if case .property(let draft) = model.authoring {
            #expect(draft.isDuplicate)
            #expect(draft.values.valueText == "miller")
        } else {
            Issue.record("duplicate should open a property draft")
        }
        model.mutateProperty { $0.values.valueText = "changed" }
        model.duplicate(line: line!)
        #expect(model.pendingLeave != nil)
        if case .property(let draft) = model.authoring {
            #expect(draft.values.valueText == "changed")
        }
        model.discardLeaveChanges()
        if case .property(let draft) = model.authoring {
            #expect(draft.values.valueText == "miller")
        } else {
            Issue.record("confirm should replace the dirty draft")
        }
    }

    @Test func connectSeedsAnEmptyRole() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(
            store: store,
            citationID: "cit-1",
            connectFrom: subjectID,
            connectTo: eventID,
            bridge: "participation"
        )
        await model.prepare()
        if case .bridge(let draft) = model.authoring {
            #expect(draft.endA == subjectID)
            #expect(draft.endB == eventID)
            #expect(draft.termID.isEmpty)
            #expect(draft.origin == .graph)
            #expect(
                CitationAuthoringNavigation.initialBridgeFocus(
                    endA: draft.endA,
                    endB: draft.endB,
                    hasTermField: true,
                    termID: draft.termID
                ) == .term
            )
        } else {
            Issue.record("Connect should open a bridge draft")
        }
    }

    @Test func personAndPlaceCannotSave() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        let pair = CitationComposerAuthoring.resolvePair(
            endA: subjectID,
            endB: placeID,
            termID: "",
            vocabulary: model.vocabulary
        )
        guard case .refused = pair else {
            Issue.record("person and place should be refused, got \(pair)")
            return
        }
    }

    @Test func saveKeepsTheSubject() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        await model.saveAuthoring()
        if case .property(let draft) = model.authoring {
            #expect(draft.values.subjectID == subjectID)
            #expect(draft.values.propertyID.isEmpty)
            #expect(draft.values.valueText.isEmpty)
            #expect(draft.holdsLeave == false)
        } else {
            Issue.record("the box should stay open on the subject")
        }
    }

    @Test func citationPreviewFallsBack() {
        #expect(CitationComposerAuthoring.preview(transcription: "line", description: "desc") == .text("line"))
        #expect(CitationComposerAuthoring.preview(transcription: "", description: "desc") == .text("desc"))
        #expect(CitationComposerAuthoring.preview(transcription: "  ", description: "") == .empty)
    }

    @Test func steppingWithDirtyCitationFieldsAsks() async {
        let store = makeStore()
        seedArtifact(store)
        store.artifactsBySource[sourceID]?.append(
            CatalogArtifact(
                id: "art-1",
                ref: "ART-1",
                sourceID: sourceID,
                fileID: "file-1",
                label: "Scan 2",
                description: "",
                file: CatalogFileRef(
                    id: "file-1",
                    relPath: "objects/file-1.pdf",
                    originalFilename: "scan-2.pdf",
                    mediaType: "application/pdf",
                    byteSize: 10
                )
            )
        )
        seedCitation(store)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.transcription = "unsaved"
        model.stepArtifact(by: 1)
        #expect(model.pendingLeave != nil)
        #expect(model.activeCitationID == "cit-1")
        model.keepEditingAfterLeave()
        #expect(model.pendingLeave == nil)
        #expect(model.transcription == "unsaved")
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

@Suite
struct CitationAuthoringNavigationTests {
    @Test func propertyDraftWalksSubjectThenPropertyThenValue() {
        #expect(CitationAuthoringNavigation.next(after: .subject, bridgeHasTerm: false) == .property)
        #expect(CitationAuthoringNavigation.next(after: .property, bridgeHasTerm: false) == .value)
        #expect(CitationAuthoringNavigation.next(after: .value, bridgeHasTerm: false) == nil)
    }

    @Test func bridgeDraftWalksTheEndsAndStopsWhenThereIsNoTerm() {
        #expect(CitationAuthoringNavigation.next(after: .endA, bridgeHasTerm: false) == .endB)
        #expect(CitationAuthoringNavigation.next(after: .endB, bridgeHasTerm: false) == nil)
        #expect(CitationAuthoringNavigation.next(after: .endB, bridgeHasTerm: true) == .term)
        #expect(CitationAuthoringNavigation.next(after: .term, bridgeHasTerm: true) == nil)
    }

    @Test func creatingASubjectDoesNotAdvance() {
        #expect(CitationAuthoringNavigation.shouldAdvance("sub-1"))
        #expect(!CitationAuthoringNavigation.shouldAdvance(""))
        #expect(!CitationAuthoringNavigation.shouldAdvance("  "))
        #expect(!CitationAuthoringNavigation.shouldAdvance(CitationComposerModel.newSubjectPrefix + "person"))
    }

    @Test func insertedPropertyFieldsClaimTabUntilTheyAreOnScreen() {
        #expect(CitationAuthoringNavigation.next(after: .subject, bridgeHasTerm: false) == .property)
        #expect(CitationAuthoringNavigation.next(after: .property, bridgeHasTerm: false) == .value)
        #expect(CitationAuthoringNavigation.claimsForwardTab(
            nextFieldAlreadyFocusable: false,
            nextFieldWillExist: true
        ))
        #expect(!CitationAuthoringNavigation.claimsForwardTab(
            nextFieldAlreadyFocusable: true,
            nextFieldWillExist: true
        ))
    }

    @Test func tabClaimsTheBridgeTermOnlyWhenTheCommitInsertsIt() {
        #expect(CitationAuthoringNavigation.claimsForwardTab(
            nextFieldAlreadyFocusable: false,
            nextFieldWillExist: true
        ))
        #expect(!CitationAuthoringNavigation.claimsForwardTab(
            nextFieldAlreadyFocusable: true,
            nextFieldWillExist: true
        ))
        #expect(!CitationAuthoringNavigation.claimsForwardTab(
            nextFieldAlreadyFocusable: false,
            nextFieldWillExist: false
        ))
    }
}
