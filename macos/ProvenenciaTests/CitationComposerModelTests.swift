import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CitationComposerModelTests {
    private let projectDir = "/tmp/citation-composer-model.provenencia"
    private let sourceID = "src-1"
    private let subjectID = "sub-person-1"
    private let personTypeID = "type-person"
    private let participationTypeID = "type-participation"
    private let namePropertyID = "prop-name"
    private let occupationPropertyID = "prop-occupation"
    private let personEdgePropertyID = "prop-person-edge"
    private let eventEdgePropertyID = "prop-event-edge"

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
                id: participationTypeID,
                key: "participation",
                origin: "provenencia",
                label: "Participation",
                description: "",
                refPrefix: "PTN",
                candidateRefPrefix: "CPA"
            ),
        ]
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: namePropertyID,
                key: "name",
                origin: "provenencia",
                label: "Name",
                description: "",
                valueType: "name"
            ),
            CatalogProperty(
                id: occupationPropertyID,
                key: "occupation",
                origin: "provenencia",
                label: "Occupation",
                description: "",
                valueType: "text"
            ),
            CatalogProperty(
                id: personEdgePropertyID,
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                valueType: "subject"
            ),
            CatalogProperty(
                id: eventEdgePropertyID,
                key: "event",
                origin: "provenencia",
                label: "Event",
                description: "",
                valueType: "subject"
            ),
            CatalogProperty(
                id: "prop-age",
                key: "age",
                origin: "user",
                label: "Age",
                description: "",
                valueType: "integer"
            ),
        ]
        store.subjectTypeFieldsByType[participationTypeID] = [
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![2],
                sortOrder: 0,
                locked: true
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![3],
                sortOrder: 1,
                locked: true
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![1],
                sortOrder: 2,
                locked: false
            ),
        ]
        store.subjectTypeFieldsByType[personTypeID] = [
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![0],
                sortOrder: 0,
                locked: false
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![1],
                sortOrder: 1,
                locked: false
            ),
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]![4],
                sortOrder: 2,
                locked: false
            ),
        ]
        store.connectRules = [
            CatalogConnectRule(
                fromTypeKey: "person",
                toTypeKey: "event",
                bridgeTypeKey: "participation",
                edgePropertyKeys: ["person", "event"],
                disambiguation: "role",
                refuse: false
            ),
        ]
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID,
                ref: "SRC-1",
                sourceTypeID: "stype",
                title: "Census",
                description: ""
            ),
        ]
        let person = CatalogSubject(
            id: subjectID,
            ref: "CPR-1",
            sourceID: sourceID,
            subjectTypeID: personTypeID,
            label: "Margt.",
            description: ""
        )
        store.subjectsBySource[sourceID] = [person]
        store.subjectPositionsBySubject[subjectID] = CatalogSubjectPosition(
            subjectID: subjectID,
            gridX: 0,
            gridY: 0
        )
        return store
    }

    private func seedArtifact(_ store: FakeStore, count: Int = 1, pdf: Bool = true) {
        store.artifactsBySource[sourceID] = (0..<count).map { i in
            CatalogArtifact(
                id: "art-\(i)",
                ref: "ART-\(i)",
                sourceID: sourceID,
                fileID: "file-\(i)",
                label: "Scan \(i + 1)",
                description: "",
                file: CatalogFileRef(
                    id: "file-\(i)",
                    relPath: "objects/file-\(i).pdf",
                    originalFilename: "scan\(i).pdf",
                    mediaType: pdf ? "application/pdf" : "image/jpeg",
                    byteSize: 10
                )
            )
        }
    }

    private func makeModel(
        store: FakeStore,
        subjectID: String? = nil,
        citationID: String? = nil
    ) -> CitationComposerModel {
        let session = WorkspaceSession(
            projectKey: ProjectKey(projectDir: projectDir),
            store: store
        )
        return CitationComposerModel(
            sourceID: sourceID,
            subjectID: subjectID ?? self.subjectID,
            citationID: citationID,
            session: session,
            store: store,
            userID: "user-1"
        )
    }

    @Test func excludedEdgePropertyKeysForBridgeType() {
        let keys = CitationComposerModel.excludedEdgePropertyKeys(
            typeKey: "participation",
            rules: [
                CatalogConnectRule(
                    fromTypeKey: "person",
                    toTypeKey: "event",
                    bridgeTypeKey: "participation",
                    edgePropertyKeys: ["person", "event"],
                    disambiguation: "role",
                    refuse: false
                ),
            ]
        )
        #expect(keys == Set(["person", "event"]))
    }

    @Test func prepareAutoSelectsSingleArtifactAndFiltersNameType() async {
        let store = makeStore()
        seedArtifact(store, count: 1)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-0")
        #expect(model.observations.isEmpty)
        #expect(model.availableProperties.map(\.id).sorted() == ["prop-age", occupationPropertyID].sorted())
        #expect(!model.availableProperties.contains(where: { $0.valueType == "name" }))
    }

    @Test func prepareShowsPickerUntilContinue() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .pickArtifact)
        #expect(model.selectedArtifactID == nil)
        model.selectPendingArtifact("art-1")
        #expect(model.phase == .pickArtifact)
        model.confirmArtifactSelection()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-1")
        #expect(model.observations.isEmpty)
    }

    @Test func prepareNoArtifactsOpensInertCompose() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.hasNoArtifacts)
        #expect(model.sourcePageLocation().sourceSurface == .page)
        #expect(model.sourcePageLocation().sourceId == sourceID)
    }

    @Test func prepareExcludesConnectEdgePropertiesOnBridge() async {
        let store = makeStore()
        seedArtifact(store)
        let bridgeID = "sub-bridge-1"
        store.subjectsBySource[sourceID]?.append(
            CatalogSubject(
                id: bridgeID,
                ref: "CPA-1",
                sourceID: sourceID,
                subjectTypeID: participationTypeID,
                label: "Participation",
                description: ""
            )
        )
        store.subjectPositionsBySubject[bridgeID] = CatalogSubjectPosition(
            subjectID: bridgeID,
            gridX: 1,
            gridY: 0
        )
        let model = makeModel(store: store, subjectID: bridgeID)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.availableProperties.map(\.key) == ["occupation"])
        #expect(!model.availableProperties.contains(where: { $0.key == "person" }))
    }

    @Test func observationDialogCommitsRow() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        var draft = model.observationDialog!
        draft.propertyID = occupationPropertyID
        draft.valueText = "Farmer"
        model.updateObservationDialog(draft)
        model.confirmObservationDialog()
        #expect(model.observationDialog == nil)
        #expect(model.observations.count == 1)
        #expect(model.observations[0].valueText == "Farmer")
    }

    @Test func submitRequiresLocatorAndObservationThenWrites() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)

        let noObs = await model.submit()
        #expect(noObs == nil)
        #expect(model.locatorError != nil || model.formError != nil)

        model.markWholeImageLocator()
        #expect(model.hasLocator)

        let stillEmpty = await model.submit()
        #expect(stillEmpty == nil)
        #expect(model.formError != nil)

        model.beginAddObservation()
        var draft = model.observationDialog!
        draft.propertyID = occupationPropertyID
        draft.valueText = "Farmer"
        model.updateObservationDialog(draft)
        model.confirmObservationDialog()

        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        #expect(location?.sourceId == sourceID)
        let listed = try await store.listObservationsBySource(
            projectDir: projectDir,
            sourceID: sourceID
        )
        #expect(listed.count == 1)
        #expect(listed[0].valueText == "Farmer")
        #expect(listed[0].propertyID == occupationPropertyID)
        #expect(listed[0].subjectID == subjectID)
    }

    @Test func pdfPageNavCommitsLocator() async {
        let store = makeStore()
        seedArtifact(store, count: 1, pdf: true)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(!model.hasLocator)
        // Stub page count is 1, so next is a no-op — Draw region commits page for PDF.
        model.markWholeImageLocator()
        #expect(model.locatorPage == 1)
        model.clearLocator()
        #expect(!model.hasLocator)
    }

    @Test func missingSubjectFallsBack() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store, subjectID: "missing")
        await model.prepare()
        #expect(model.phase == .subjectMissing)
        #expect(model.shouldFallbackToGraph)
    }

    @Test func prepareLoadsExistingCitationForEdit() async throws {
        let store = makeStore()
        seedArtifact(store)
        let citation = CatalogCitation(
            id: "cit-edit",
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"page","artifact_page":3}]}"#,
            transcription: "Farmer",
            description: "Note",
            transcriptionUncertain: true,
            transcriptionNote: "blurry"
        )
        store.citationsByID[citation.id] = citation
        store.observationsBySource[sourceID] = [
            CatalogObservation(
                id: "obs-1",
                ref: "OBS-1",
                citationID: citation.id,
                subjectID: subjectID,
                propertyID: occupationPropertyID,
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
            ),
        ]
        let model = makeModel(store: store, citationID: citation.id)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.isEditingExisting)
        #expect(model.selectedArtifactID == "art-0")
        #expect(model.transcription == "Farmer")
        #expect(model.transcriptionUncertain)
        #expect(model.locatorPage == 3)
        #expect(model.observations.count == 1)
        #expect(model.observations[0].valueText == "Farmer")
    }

    @Test func submitUpdatesExistingCitation() async throws {
        let store = makeStore()
        seedArtifact(store)
        let citation = CatalogCitation(
            id: "cit-edit",
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"page","artifact_page":1}]}"#,
            transcription: "Old",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.citationsByID[citation.id] = citation
        store.observationsBySource[sourceID] = [
            CatalogObservation(
                id: "obs-1",
                ref: "OBS-1",
                citationID: citation.id,
                subjectID: subjectID,
                propertyID: occupationPropertyID,
                polarity: "positive",
                valueText: "Old",
                valueInteger: nil,
                valueDateID: "",
                valueNameID: "",
                valueSubjectID: "",
                valueTermID: "",
                propertyKey: "occupation",
                propertyLabel: "Occupation",
                propertyValueType: "text"
            ),
        ]
        let model = makeModel(store: store, citationID: citation.id)
        await model.prepare()
        model.transcription = "New"
        let row = model.observations[0]
        model.beginEditObservation(row)
        var draft = model.observationDialog!
        draft.valueText = "Miller"
        model.updateObservationDialog(draft)
        model.confirmObservationDialog()
        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        let (_, _, listed) = try await store.getCitation(
            projectDir: projectDir,
            citationID: citation.id
        )
        #expect(store.citationsByID[citation.id]?.transcription == "New")
        #expect(listed.count == 1)
        #expect(listed[0].valueText == "Miller")
    }
}
