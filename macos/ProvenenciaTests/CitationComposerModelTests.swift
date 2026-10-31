import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CitationComposerModelTests {
    private let projectDir = "/tmp/citation-composer-model.provenencia"
    private let sourceID = "src-1"
    private let graphID = "g-1"
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
    private let regionLocatorJSON = """
    {"version":1,"selectors":[{"type":"artifact"},{"type":"region","unit":"normalized","points":[{"x":0.1,"y":0.1},{"x":0.4,"y":0.1},{"x":0.4,"y":0.4},{"x":0.1,"y":0.4}]}]}
    """

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
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: graphID, ref: "GPH-1", sourceID: sourceID, name: "Default", description: "", artifactIDs: ["art-0", "art-1"]
            ),
        ]
        store.subjectsByGraph[graphID] = [
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
        store.nextWriteEffects = CatalogWriteEffects(
            evidenceGraphs: [graphID], reads: [CatalogWriteEffectsRead(read: .citations, ids: [graphID])],
            lists: [.propertyUsage]
        )
        return store
    }

    private func artifact(_ n: Int) -> CatalogArtifact {
        CatalogArtifact(
            id: "art-\(n)",
            ref: "ART-\(n)",
            sourceID: sourceID,
            fileID: "file-\(n)",
            label: "Scan \(n + 1)",
            description: "",
            file: CatalogFileRef(
                id: "file-\(n)",
                relPath: "objects/file-\(n).pdf",
                originalFilename: "scan.pdf",
                mediaType: "application/pdf",
                byteSize: 10
            )
        )
    }

    private func seedArtifact(_ store: FakeStore) {
        store.artifactsBySource[sourceID] = [artifact(0)]
    }

    /// Replaces the fixture graph's links.
    private func link(_ store: FakeStore, _ artifactIDs: [String]) {
        store.evidenceGraphsByProject[projectDir]?[0].artifactIDs = artifactIDs
    }

    /// A second graph of the Source, Smith, linking art-0, with one person.
    private func seedSecondGraph(_ store: FakeStore) {
        store.evidenceGraphsByProject[projectDir]?.append(
            CatalogEvidenceGraph(
                id: "g-2", ref: "GPH-2", sourceID: sourceID, name: "Smith", description: "", artifactIDs: ["art-0"]
            )
        )
        store.subjectsByGraph["g-2"] = [
            subject(id: "sub-person-2", typeID: personTypeID, ref: "CPR-2", label: "Clara"),
        ]
        store.subjectPositionsBySubject["sub-person-2"] = CatalogSubjectPosition(
            subjectID: "sub-person-2", gridX: 0, gridY: 0
        )
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
        store.observationsByGraph[graphID] = observations
    }

    private func makeModel(
        store: FakeStore,
        graphID: String? = nil,
        subjectID: String? = nil,
        citationID: String? = nil,
        observationID: String? = nil,
        connectFrom: String? = nil,
        connectTo: String? = nil,
        bridge: String? = nil
    ) -> CitationComposerModel {
        let graphID = graphID ?? self.graphID
        let entry: CitationComposerEntry
        if let connectFrom, let connectTo, let bridge {
            entry = .connect(
                sourceID: sourceID,
                graphID: graphID,
                fromSubjectID: connectFrom,
                toSubjectID: connectTo,
                bridgeTypeKey: bridge
            )
        } else if let citationID {
            entry = .edit(
                sourceID: sourceID,
                graphID: graphID,
                subjectID: subjectID ?? self.subjectID,
                citationID: citationID,
                artifactID: "art-0",
                observationID: observationID
            )
        } else {
            entry = .addProperty(
                sourceID: sourceID,
                graphID: graphID,
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
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations graph=\(graphID) observations=0") })
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
        await model.observationRows.commit(rowID: rowID, citation: model.fields.values)
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
        store.subjectsByGraph[graphID]?.append(bridge)
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
        store.subjectsByGraph[graphID]?.append(bridge)
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

    // MARK: - Evidence graph scope (S12-10)

    @Test func composerListsOnlyTheGraphsLinkedArtifacts() async {
        let store = makeStore()
        store.artifactsBySource[sourceID] = [artifact(0), artifact(1)]
        link(store, ["art-1"])
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.artifacts.map(\.id) == ["art-1"])
        #expect(model.selectedArtifactID == "art-1")
    }

    @Test func composerKeepsPageOrderForLinkedArtifacts() async {
        let store = makeStore()
        store.artifactsBySource[sourceID] = [artifact(0), artifact(1)]
        link(store, ["art-1", "art-0"])
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.artifacts.map(\.id) == ["art-0", "art-1"])
    }

    @Test func graphWithNoLinksShowsTheNoArtifactGate() async {
        let store = makeStore()
        seedArtifact(store)
        link(store, [])
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.hasNoArtifacts)
    }

    @Test func unlinkedEntryArtifactOpensANewCitation() async {
        let store = makeStore()
        store.artifactsBySource[sourceID] = [artifact(0), artifact(1)]
        link(store, ["art-1"])
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.selectedArtifactID == "art-1")
        #expect(model.activeCitationID == nil)
    }

    @Test func graphListFailureFailsTheLoad() async {
        enum Boom: Error { case boom }
        let store = makeStore()
        seedArtifact(store)
        store.listEvidenceGraphsError = Boom.boom
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .loadFailed)
    }

    @Test func stepperListsOnlyThisGraphsCitations() async {
        let store = makeStore()
        seedArtifact(store)
        seedSecondGraph(store)
        for (id, ref) in [("cit-a", "CIT-A"), ("cit-b", "CIT-B"), ("cit-c", "CIT-C")] {
            store.citationsByID[id] = CatalogCitation(
                id: id, ref: ref, artifactID: "art-0", locatorJSON: "{}", transcription: "",
                description: "", transcriptionUncertain: false, transcriptionNote: ""
            )
        }
        store.observationsByGraph[graphID] = [occupationObservation(id: "obs-a", ref: "OBS-A", citationID: "cit-a")]
        var other = occupationObservation(id: "obs-b", ref: "OBS-B", citationID: "cit-b")
        other.subjectID = "sub-person-2"
        store.observationsByGraph["g-2"] = [other]

        let jones = makeModel(store: store)
        await jones.prepare()
        #expect(jones.listedCitations.map(\.citation.ref).sorted() == ["CIT-A", "CIT-B", "CIT-C"])
        #expect(jones.listedCount(for: "art-0") == 3)

        let smith = makeModel(store: store, graphID: "g-2", subjectID: "sub-person-2")
        await smith.prepare()
        #expect(smith.listedCitations.map(\.citation.ref).sorted() == ["CIT-A", "CIT-B", "CIT-C"])
    }

    @Test func firstObservationMovesANewCitationIntoTheGraph() async {
        let store = makeStore()
        seedArtifact(store)
        seedSecondGraph(store)
        let jones = makeModel(store: store)
        await jones.prepare()
        jones.beginPropertyDraft(subjectID: subjectID)
        jones.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        store.nextWriteEffects = CatalogWriteEffects(
            reads: [CatalogWriteEffectsRead(read: .citations, ids: [graphID])], lists: [.propertyUsage]
        )
        await jones.saveAuthoring()
        let created = jones.activeCitationID
        #expect(created != nil)
        #expect(jones.listedCitations.contains { $0.citation.id == created })

        let smith = makeModel(store: store, graphID: "g-2", subjectID: "sub-person-2")
        await smith.prepare()
        #expect(smith.listedCitations.contains { $0.id == created })
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
        #expect(store.recordedCalls.contains { $0 == "createCitationWithObservations graph=\(graphID) observations=0" })
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
        #expect(creates.first == "createCitationWithObservations graph=\(graphID) observations=0")
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
        #expect(model.hasUnsavedDocumentWork == false)
        model.transcription = "dirty"
        #expect(model.hasUnsavedDocumentWork)
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
        #expect(connect.hasUnsavedDocumentWork)
        connect.mutateBridge { $0.termID = "term-witness" }
        #expect(connect.hasUnsavedDocumentWork)
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

    @Test func newCitationClearsThePreviousLocator() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, locatorJSON: regionLocatorJSON)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.fields.locator.hasRegion)
        model.selectCitation(nil)
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == nil)
        #expect(model.fields.locator.isArtifactOnly)
        #expect(model.fields.isDirty == false)
    }

    @Test func newCitationFromTheObservationDialogClearsTheLocator() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, locatorJSON: regionLocatorJSON)
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.selectCitation(nil)
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == nil)
        #expect(model.fields.locator.isArtifactOnly)
        guard case .property = model.authoring else {
            Issue.record("the observation dialog stays open")
            return
        }
    }

    @Test func pendingConnectionOnNewAndExisting() async {
        let store = makeStore()
        seedArtifact(store)
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
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations graph=\(graphID) observations=0") })
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
        #expect(store.recordedCalls.contains { $0.hasPrefix("createCitationWithObservations graph=\(graphID) observations=0") })
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
        let orders = NameFormatOrders.givenSurname
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
        store.subjectsByGraph[graphID]?.append(bridge)
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
        store.subjectsByGraph[graphID]?.append(
            subject(id: "bridge-1", typeID: participationTypeID, ref: "PTN-1", label: "")
        )
        store.observationsByGraph[graphID, default: []].append(
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
        model.session.apply(CatalogWriteEffects(evidenceGraphs: [graphID]))
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

    /// T-18: the composer entry needs the location's graph, and an existing
    /// Citation read straight from the store names that graph.
    @Test func entryNeedsTheGraph() async {
        let location = WorkspaceLocation.citationComposer(
            sourceId: sourceID, graphId: graphID, subjectId: subjectID, artifactId: "art-0"
        )
        #expect(CitationComposerEntry(location: location)?.graphID == graphID)
        #expect(CitationComposerEntry(location: location)?.sourceID == sourceID)
        // Without a graph the location is the Source page, which is no entry.
        let bare = WorkspaceLocation(
            section: .sources, sourceId: sourceID, subjectId: subjectID, sourceSurface: .citationComposer
        )
        #expect(bare.sourceSurface == .page)
        #expect(CitationComposerEntry(location: bare) == nil)

        // A Citation on a file the Source no longer has is not in the cached
        // lists, so the composer reads it from the store, on its graph.
        let store = makeStore()
        seedArtifact(store)
        store.citationsByID["cit-gone"] = CatalogCitation(
            id: "cit-gone", ref: "CIT-9", artifactID: "art-gone",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "", description: "", transcriptionUncertain: false, transcriptionNote: ""
        )
        let model = makeModel(store: store, citationID: "cit-gone")
        await model.prepare()
        #expect(store.recordedCalls.contains("getCitation cit-gone graphID=\(graphID)"))
    }

    @Test func connectEntryIgnoresLegacyTermAndGrid() throws {
        let json = """
        {
          "section":"sources",
          "sourceId":"\(sourceID)",
          "graphId":"\(graphID)",
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
            graphID: graphID,
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

    // MARK: term keys honour the term's origin

    private func termVocabulary(
        propertyKey: String, terms: [(key: String, label: String, origin: String)], endTypeKey: String
    ) -> CitationComposerVocabulary {
        let property = CatalogProperty(
            id: "p-\(propertyKey)", key: propertyKey, origin: CatalogOrigin.provenencia, label: propertyKey,
            description: "", valueType: "term"
        )
        func end(_ id: String) -> CitationComposerModel.GraphSubjectOption {
            .init(id: id, label: id, ref: id, typeID: "t", typeKey: endTypeKey)
        }
        var vocabulary = CitationComposerVocabulary.empty
        vocabulary.propertiesByID = [property.id: property]
        vocabulary.subjectsByID = ["a": end("a"), "b": end("b")]
        vocabulary.connectRules = CatalogConnectRule.productMatrix
        vocabulary.termsByPropertyID[property.id] = terms.map {
            CatalogPropertyTerm(
                id: "term-\($0.origin)-\($0.key)", propertyID: property.id, key: $0.key,
                origin: $0.origin, label: $0.label, description: ""
            )
        }
        return vocabulary
    }

    private func readyBridge(_ vocabulary: CitationComposerVocabulary, termID: String) -> CitationReadyBridge? {
        guard case .ready(let ready) = CitationComposerAuthoring.resolvePair(
            endA: "a", endB: "b", termID: termID, vocabulary: vocabulary
        ) else { return nil }
        return ready
    }

    @Test func aProductTermsInverseIsNeverAResearchersTerm() {
        let vocabulary = termVocabulary(
            propertyKey: "relationship_type",
            terms: [
                (key: "parent", label: "Parent", origin: CatalogOrigin.provenencia),
                (key: "child", label: "Kid", origin: CatalogOrigin.user),
            ],
            endTypeKey: "person"
        )
        let ready = readyBridge(vocabulary, termID: "term-provenencia-parent")
        #expect(ready?.endBTitle == "Parent")
        let theirs = termVocabulary(
            propertyKey: "relationship_type",
            terms: [
                (key: "parent", label: "Mine", origin: CatalogOrigin.user),
                (key: "child", label: "Child", origin: CatalogOrigin.provenencia),
            ],
            endTypeKey: "person"
        )
        #expect(readyBridge(theirs, termID: "term-user-parent")?.endBTitle == "Mine")
        let product = termVocabulary(
            propertyKey: "relationship_type",
            terms: [
                (key: "parent", label: "Parent", origin: CatalogOrigin.provenencia),
                (key: "child", label: "Child", origin: CatalogOrigin.provenencia),
            ],
            endTypeKey: "person"
        )
        #expect(readyBridge(product, termID: "term-provenencia-parent")?.endBTitle == "Child")
    }

    @Test func aResearchersSpouseTermGetsNoSpouseSentence() {
        let user = termVocabulary(
            propertyKey: "relationship_type",
            terms: [(key: "spouse", label: "Partner", origin: CatalogOrigin.user)],
            endTypeKey: "person"
        )
        #expect(readyBridge(user, termID: "term-user-spouse")?.reading == L10n.CitationComposer.isATermOf("a", "partner", "b"))
        let product = termVocabulary(
            propertyKey: "relationship_type",
            terms: [(key: "spouse", label: "Spouse", origin: CatalogOrigin.provenencia)],
            endTypeKey: "person"
        )
        #expect(readyBridge(product, termID: "term-provenencia-spouse")?.reading == L10n.CitationComposer.areSpouses("a", "b"))
    }

    @Test func aResearchersPlaceTermKeyedPartOfKeepsItsOwnLabel() {
        let theirs = CatalogPropertyTerm(
            id: "t", propertyID: "p", key: "part_of", origin: CatalogOrigin.user, label: "Inside of", description: ""
        )
        #expect(CitationBridgeSentence.displayLabel(theirs) == "inside of")
        var product = theirs
        product.origin = CatalogOrigin.provenencia
        #expect(CitationBridgeSentence.displayLabel(product) == L10n.string(L10n.CitationComposer.placeTermPartOf))
        let vocabulary = termVocabulary(
            propertyKey: "place_relationship_type",
            terms: [(key: "part_of", label: "Inside of", origin: CatalogOrigin.user)],
            endTypeKey: "place"
        )
        let ready = readyBridge(vocabulary, termID: "term-user-part_of")
        #expect(ready?.endATitle == "Inside of" && ready?.reading == nil)
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
        model.requestSelectArtifact("art-1")
        #expect(model.pendingLeave != nil)
        #expect(model.activeCitationID == "cit-1")
        model.keepEditingAfterLeave()
        #expect(model.pendingLeave == nil)
        #expect(model.transcription == "unsaved")
    }

    @Test func artifactSwitchShowsAFailedCitationLoad() async {
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
        store.citationsByID["cit-2"] = CatalogCitation(
            id: "cit-2",
            ref: "CIT-2",
            artifactID: "art-1",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "other",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        let listKey = CatalogQueryKey.citationsByArtifact(
            project: model.session.projectKey, graphId: graphID, artifactId: "art-1"
        )
        let _: QueryHandle<[CatalogListedCitation]> = model.session.query(listKey)
        _ = await model.session.readyValue(listKey) as [CatalogListedCitation]?

        // The Citation can no longer be fetched, and the graph's cache cannot answer for it.
        store.citationsByID["cit-2"] = nil
        store.listSubjectsError = ComposerLoadFailure()
        let graphHandle: QueryHandle<EvidenceGraphRows>? = model.session.queryHandle(model.graphKey)
        graphHandle?.resetForInvalidation()

        model.requestSelectArtifact("art-1")
        let switched = await waitUntil { !model.isSwitchingCitation && model.selectedArtifactID == "art-1" }
        #expect(switched)
        #expect(model.fields.error != nil)
        #expect(model.activeCitationID == nil)
    }

    /// D-8: selecting an Artifact rewrites the current location's Artifact in
    /// place and keeps every other id and display field; nothing is pushed.
    @Test func selectingAnArtifactRewritesTheLocationInPlace() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store, citationID: "cit-1")
        let navigation = WorkspaceNavigation()
        let location = WorkspaceLocation(
            section: .sources, sourceId: sourceID, graphId: graphID, subjectId: subjectID,
            citationId: "cit-1", artifactId: "art-0", observationId: "obs-1",
            sourceSurface: .citationComposer, ref: "CIT-1", title: "Ada", sourceTitle: "Census"
        )
        navigation.replaceCurrent(with: location)
        model.navigation = navigation

        model.activeCitationID = "cit-1"
        model.selectedArtifactID = "art-1"

        var expected = location
        expected.artifactId = "art-1"
        #expect(navigation.currentLocation == expected)
        #expect(navigation.currentLocation.observationId == "obs-1")
        #expect(navigation.currentLocation.title == "Ada")
        #expect(navigation.currentLocation.sourceTitle == "Census")
        #expect(CitationComposerEntry(location: expected)?.identityKey
            == CitationComposerEntry(location: location)?.identityKey)

        model.selectedArtifactID = nil
        #expect(navigation.currentLocation.artifactId == nil)
    }

    /// An Artifact switch that drops the open Citation (as selectArtifactAndLoad
    /// does before it sets the Artifact) clears the Citation's ids from the
    /// location; the display fields stay.
    @Test func switchingArtifactsAwayFromAnEditedCitationClearsItsIds() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store, citationID: "cit-1")
        let navigation = WorkspaceNavigation()
        let location = WorkspaceLocation(
            section: .sources, sourceId: sourceID, graphId: graphID, subjectId: subjectID,
            citationId: "cit-1", artifactId: "art-0", observationId: "obs-1",
            sourceSurface: .citationComposer, ref: "CIT-1", title: "Ada", sourceTitle: "Census"
        )
        navigation.replaceCurrent(with: location)
        model.navigation = navigation

        model.activeCitationID = nil
        model.selectedArtifactID = "art-1"

        #expect(navigation.currentLocation.artifactId == "art-1")
        #expect(navigation.currentLocation.citationId == nil)
        #expect(navigation.currentLocation.observationId == nil)
        #expect(navigation.currentLocation.title == "Ada")
        #expect(navigation.currentLocation.ref == "CIT-1")
        #expect(navigation.currentLocation.sourceTitle == "Census")
    }

    @Test func lineRefreshShowsAFailedCitationLoad() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(
            store,
            observations: [occupationObservation(id: "obs-1", ref: "OBS-1", citationID: "cit-1")]
        )
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        let rowCount = model.observationRows.rows.count
        #expect(rowCount == 1)

        let kept = store.citationsByID["cit-1"]
        store.citationsByID["cit-1"] = nil
        model.session.invalidate(
            .citationsByArtifact(project: model.session.projectKey, graphId: graphID, artifactId: "art-0")
        )
        await model.refreshCitationLines()
        #expect(model.fields.error != nil)
        #expect(model.observationRows.rows.count == rowCount)

        store.citationsByID["cit-1"] = kept
        model.session.invalidate(
            .citationsByArtifact(project: model.session.projectKey, graphId: graphID, artifactId: "art-0")
        )
        await model.refreshCitationLines()
        #expect(model.fields.error == nil)
    }

    @Test func customTermReturnsToTheNewObservationDialog() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.beginAddCustomTerm(propertyID: occupationPropertyID)
        guard case .customTerm(_, let over) = model.presented else {
            Issue.record("custom term should open over the new observation")
            return
        }
        #expect(over == .authoring)
        model.customTermDraft?.label = "Weaver"
        model.cancelCustomTermDialog()
        #expect(model.presented == .authoring)
        #expect(model.customTermDraft == nil)
        if case .property = model.authoring {} else {
            Issue.record("the new observation draft should still be open")
        }
    }

    @Test func keepEditingRestoresACustomTermDraft() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty {
            $0.values.propertyID = occupationPropertyID
            $0.values.valueText = "miller"
        }
        model.beginAddCustomTerm(propertyID: occupationPropertyID)
        model.customTermDraft?.label = "Weaver"
        #expect(model.shouldHoldNavigation(.back))
        guard case .leave(_, let returnTo) = model.presented,
              case .customTerm(let draft, .authoring) = returnTo
        else {
            Issue.record("leave should return to the custom term, got \(String(describing: model.presented))")
            return
        }
        #expect(draft.label == "Weaver")
        model.keepEditingAfterLeave()
        guard case .customTerm(let restored, .authoring) = model.presented else {
            Issue.record("keep editing should reopen the custom term")
            return
        }
        #expect(restored.label == "Weaver")
    }

    @Test func discardingLeaveDropsTheChildDialog() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginPropertyDraft(subjectID: subjectID)
        model.mutateProperty { $0.values.valueText = "miller" }
        model.beginAddCustomTerm(propertyID: occupationPropertyID)
        model.customTermDraft?.label = "Weaver"
        #expect(model.shouldHoldNavigation(.back))
        model.discardLeaveChanges()
        #expect(model.presented == nil)
        #expect(model.customTermDraft == nil)
        #expect(model.authoring == .rest)
    }

    @Test func keepEditingRestoresTheEditDialog() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.beginEdit(observationID: "obs-1", subjectID: subjectID))
        model.mutateEdit { session in
            guard case .property(var values, let id, let ref) = session.kind else { return }
            values.valueText = "changed"
            session.kind = .property(values: values, persistedID: id, persistedRef: ref)
        }
        #expect(model.shouldHoldNavigation(.back))
        guard case .leave(_, let returnTo) = model.presented else {
            Issue.record("leave should be showing")
            return
        }
        #expect(returnTo == .edit)
        model.keepEditingAfterLeave()
        #expect(model.presented == .edit)
        if case .property(let values, _, _) = model.editSession?.kind {
            #expect(values.valueText == "changed")
        } else {
            Issue.record("the edit session should still hold the draft")
        }
    }

    @Test func newSubjectFromTheEditDialogSelectsIt() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.beginEdit(observationID: "obs-1", subjectID: subjectID))
        model.beginNewSubject(typeKey: "person", field: .propertySubject)
        guard case .newSubject(_, let over) = model.presented else {
            Issue.record("new subject should open over the edit dialog")
            return
        }
        #expect(over == .edit)
        model.newSubjectDraft?.label = "Ada"
        await model.confirmNewSubject()
        #expect(model.presented == .edit)
        #expect(model.authoring == .rest)
        guard case .property(let values, _, _) = model.editSession?.kind else {
            Issue.record("the edit dialog should still be a property observation")
            return
        }
        let created = store.subjectsByGraph[graphID]?.last
        #expect(values.subjectID == created?.id)
        #expect(values.subjectID != subjectID)
    }

    @Test func saveEditClosesTheDialog() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.beginEdit(observationID: "obs-1", subjectID: subjectID))
        model.mutateEdit { session in
            guard case .property(var values, let id, let ref) = session.kind else { return }
            values.valueText = "changed"
            session.kind = .property(values: values, persistedID: id, persistedRef: ref)
        }
        store.recordedCalls = []
        await model.saveEdit()
        #expect(model.presented == nil)
        #expect(model.editSession == nil)
        #expect(store.recordedCalls.contains("updateObservation id=obs-1"))
    }

    @Test func saveEditKeepsTheDialogWhenThePropertyIsMissing() async {
        let store = makeStore()
        seedArtifact(store)
        seedCitation(store, observations: [occupationObservation(id: "obs-1", ref: "OBS-ABC", citationID: "cit-1")])
        let model = makeModel(store: store, citationID: "cit-1")
        await model.prepare()
        #expect(model.beginEdit(observationID: "obs-1", subjectID: subjectID))
        model.mutateEdit { session in
            guard case .property(var values, let id, let ref) = session.kind else { return }
            values.propertyID = "missing"
            session.kind = .property(values: values, persistedID: id, persistedRef: ref)
        }
        await model.saveEdit()
        #expect(model.presented == .edit)
        #expect(model.editSession?.error != nil)
        #expect(model.editSession?.isSaving == false)
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
        #expect(CitationAuthoringNavigation.next(after: .subject) == .property)
        #expect(CitationAuthoringNavigation.next(after: .property) == .value)
        #expect(CitationAuthoringNavigation.next(after: .value) == nil)
    }

    @Test func bridgeDraftWalksTheEndsAndStopsWhenThereIsNoTerm() {
        #expect(CitationAuthoringNavigation.next(after: .endA) == .endB)
        #expect(CitationAuthoringNavigation.next(after: .endB) == nil)
        #expect(CitationAuthoringNavigation.next(
            after: .endB,
            draftAfterCommit: CitationAuthoringDraftAfterCommit(bridgeHasTerm: true)
        ) == .term)
        #expect(CitationAuthoringNavigation.next(
            after: .term,
            draftAfterCommit: CitationAuthoringDraftAfterCommit(bridgeHasTerm: true)
        ) == nil)
    }

    @Test func creatingASubjectDoesNotAdvance() {
        #expect(CitationAuthoringNavigation.shouldAdvance("sub-1"))
        #expect(!CitationAuthoringNavigation.shouldAdvance(""))
        #expect(!CitationAuthoringNavigation.shouldAdvance("  "))
        #expect(!CitationAuthoringNavigation.shouldAdvance(CitationComposerModel.newSubjectPrefix + "person"))
    }

    @Test func tabAlwaysMovesToTheFieldAfterTheCommit() {
        let withTerm = CitationAuthoringDraftAfterCommit(bridgeHasTerm: true)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .subject, committed: "sub-1", via: .tab, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == .property)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .subject, committed: "sub-2", via: .tab, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == .property)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .property, committed: "prop-1", via: .tab, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == .value)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .property, committed: "prop-1", via: .click, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == nil)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .property, committed: "prop-1", via: .return, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == nil)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .endB, committed: "end-b", via: .tab, draftAfterCommit: withTerm
        ) == .term)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .endB, committed: "end-b", via: .tab, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == nil)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .value, committed: "kept", via: .tab, draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == nil)
        #expect(CitationAuthoringNavigation.tabDestination(
            from: .subject,
            committed: CitationComposerModel.newSubjectPrefix + "person",
            via: .tab,
            draftAfterCommit: CitationAuthoringDraftAfterCommit()
        ) == nil)
    }

    @Test func noArtifactsCalloutNamesTheGraphLinks() {
        #expect(L10n.string(L10n.CitationComposer.noArtifactsCallout)
            == "Every citation needs an artifact linked to this evidence graph. Add one on the Source page, or link an existing one from its card, then return here.")
    }
}

private struct ComposerLoadFailure: Error {}
