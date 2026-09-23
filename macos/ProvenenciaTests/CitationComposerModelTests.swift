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

    private func seedArtifact(
        _ store: FakeStore,
        count: Int = 1,
        pdf: Bool = true,
        mediaType: String? = nil
    ) {
        let type = mediaType ?? (pdf ? "application/pdf" : "image/jpeg")
        let ext = type.contains("pdf") ? "pdf" : (type.hasPrefix("image/") ? "jpg" : "bin")
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
                    relPath: "objects/file-\(i).\(ext)",
                    originalFilename: "scan\(i).\(ext)",
                    mediaType: type,
                    byteSize: 10
                )
            )
        }
    }

    private func stripAttachedFile(_ store: FakeStore, artifactID: String) {
        guard var artifacts = store.artifactsBySource[sourceID],
              let index = artifacts.firstIndex(where: { $0.id == artifactID })
        else { return }
        artifacts[index].fileID = ""
        artifacts[index].file = nil
        store.artifactsBySource[sourceID] = artifacts
    }

    private func sampleRectangle() -> ArtifactRegionDraft {
        ArtifactRegionDraft(
            kind: .rectangle,
            points: [
                CGPoint(x: 0.1, y: 0.1),
                CGPoint(x: 0.4, y: 0.1),
                CGPoint(x: 0.4, y: 0.3),
                CGPoint(x: 0.1, y: 0.3),
            ]
        )
    }

    private func makeModel(
        store: FakeStore,
        subjectID: String? = nil,
        citationID: String? = nil,
        connectFromSubjectID: String? = nil,
        connectToSubjectID: String? = nil,
        connectBridgeTypeKey: String? = nil,
        connectDisambiguationTermID: String? = nil
    ) -> CitationComposerModel {
        let session = WorkspaceSession(
            projectKey: ProjectKey(projectDir: projectDir),
            store: store
        )
        return CitationComposerModel(
            sourceID: sourceID,
            subjectID: subjectID ?? self.subjectID,
            citationID: citationID,
            connectFromSubjectID: connectFromSubjectID,
            connectToSubjectID: connectToSubjectID,
            connectBridgeTypeKey: connectBridgeTypeKey,
            connectDisambiguationTermID: connectDisambiguationTermID,
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

    @Test func prepareAutoSelectsSingleArtifactAndIncludesNameType() async {
        let store = makeStore()
        seedArtifact(store, count: 1)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-0")
        #expect(model.observations.isEmpty)
        #expect(model.availableProperties.map(\.id).sorted() == ["prop-age", namePropertyID, occupationPropertyID].sorted())
        #expect(model.availableProperties.contains(where: { $0.valueType == "name" }))
    }

    @Test func prepareShowsPickerUntilContinue() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .pickArtifact)
        #expect(model.selectedArtifactID == nil)
        #expect(model.pendingArtifactID == "art-0")
        model.selectPendingArtifact("art-1")
        #expect(model.phase == .pickArtifact)
        await model.confirmArtifactSelectionAndLoad()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-1")
        #expect(model.observations.isEmpty)
    }

    @Test func preparePickerDefaultsToFirstArtifactWithFile() async {
        let store = makeStore()
        seedArtifact(store, count: 3)
        stripAttachedFile(store, artifactID: "art-0")
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .pickArtifact)
        #expect(model.pendingArtifactID == "art-1")
    }

    @Test func preparePickerFallsBackToFirstWhenNoneHaveFiles() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        stripAttachedFile(store, artifactID: "art-0")
        stripAttachedFile(store, artifactID: "art-1")
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .pickArtifact)
        #expect(model.pendingArtifactID == "art-0")
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
        #expect(model.observations.isEmpty)
    }

    @Test func preparePrefillsFixedConnectEdgeRowsFromLocationPayload() async {
        let store = makeStore()
        seedArtifact(store)
        let eventTypeID = "type-event"
        store.subjectTypesByProject[projectDir]?.append(
            CatalogSubjectType(
                id: eventTypeID,
                key: "event",
                origin: "provenencia",
                label: "Event",
                description: "",
                refPrefix: "EVT",
                candidateRefPrefix: "CEV"
            )
        )
        store.propertiesByProject[projectDir]?.append(
            CatalogProperty(
                id: "prop-role",
                key: "role",
                origin: "provenencia",
                label: "Role",
                description: "",
                valueType: "term"
            )
        )
        store.subjectTypeFieldsByType[participationTypeID]?.append(
            CatalogSubjectTypeField(
                property: store.propertiesByProject[projectDir]!.last!,
                sortOrder: 3,
                locked: false
            )
        )
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
        let eventID = "sub-event-1"
        store.subjectsBySource[sourceID] = [
            CatalogSubject(
                id: subjectID,
                ref: "CPR-1",
                sourceID: sourceID,
                subjectTypeID: personTypeID,
                label: "Margt.",
                description: ""
            ),
            CatalogSubject(
                id: eventID,
                ref: "CEV-1",
                sourceID: sourceID,
                subjectTypeID: eventTypeID,
                label: "Enumeration, 1871",
                description: ""
            ),
        ]
        store.subjectPositionsBySubject[eventID] = CatalogSubjectPosition(
            subjectID: eventID, gridX: 2, gridY: 0
        )
        let model = makeModel(
            store: store,
            subjectID: "",
            connectFromSubjectID: subjectID,
            connectToSubjectID: eventID,
            connectBridgeTypeKey: "participation",
            connectDisambiguationTermID: "term-witness"
        )
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.isConnectPrefill)
        #expect(model.observations.count == 3)
        #expect(model.observations.filter(\.isConnectFixed).count == 2)
        #expect(model.observations.filter(\.isConnectFixed).map(\.valueSubjectID).sorted() == [eventID, subjectID].sorted())
        #expect(model.observations.contains(where: { $0.valueTermID == "term-witness" && !$0.isConnectFixed }))
        #expect(
            model.subjectLabel == EvidenceBridgeEdgeSummary.sentence(
                kind: .participation,
                person: "Margt.",
                related: nil,
                event: "Enumeration, 1871",
                place: nil,
                term: String(localized: L10n.PropertyTerm.roleWitness)
            )
        )
        model.removeObservation(id: model.observations[0].id)
        #expect(model.observations.filter(\.isConnectFixed).count == 2)
    }

    @Test func connectEdgePrefillRowsAssignEndpointsByPropertyKey() {
        let rows = CitationComposerModel.connectEdgePrefillRows(
            bridgeTypeKey: "participation",
            rules: [
                CatalogConnectRule(
                    fromTypeKey: "person",
                    toTypeKey: "event",
                    bridgeTypeKey: "participation",
                    edgePropertyKeys: ["person", "event"],
                    disambiguation: "role",
                    refuse: false
                ),
            ],
            properties: [
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
            ],
            endpointA: (id: "p1", label: "Alice", typeKey: "person"),
            endpointB: (id: "e1", label: "Birth", typeKey: "event")
        )
        #expect(rows.count == 2)
        #expect(rows[0].propertyID == personEdgePropertyID)
        #expect(rows[0].valueSubjectID == "p1")
        #expect(rows[0].valueText == "Alice")
        #expect(rows[0].isConnectFixed)
        #expect(rows[1].propertyID == eventEdgePropertyID)
        #expect(rows[1].valueSubjectID == "e1")
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

    @Test func observationDialogCommitsNameValueParts() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        var draft = model.observationDialog!
        draft.propertyID = namePropertyID
        draft.nameDraft = NameValueDraft(
            form: "John W. Alderwick",
            parts: [
                CatalogNameValuePart(value: "John", type: "given"),
                CatalogNameValuePart(value: "W.", type: "initial"),
                CatalogNameValuePart(value: "Alderwick", type: "surname"),
            ]
        )
        model.updateObservationDialog(draft)
        #expect(model.canConfirmObservation)
        model.confirmObservationDialog()
        #expect(model.observations.count == 1)
        #expect(NameValueDisplay.string(for: model.observations[0].nameDraft) == "John W. Alderwick")

        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        let listed = try await store.listObservationsBySource(
            projectDir: projectDir,
            sourceID: sourceID
        )
        #expect(listed.count == 1)
        #expect(listed[0].nameForm == "John W. Alderwick")
        #expect(listed[0].nameParts.map(\.type) == ["given", "initial", "surname"])
        #expect(listed[0].nameParts.map(\.value) == ["John", "W.", "Alderwick"])
    }

    @Test func submitRequiresObservationAndWritesArtifactLocator() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.locator.isArtifactOnly)
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
        let json = store.citationsByID.values.first?.locatorJSON ?? ""
        #expect(json.contains("\"artifact\""))
        #expect(!json.contains("\"page\""))
    }

    @Test func setPageDoesNotFollowBrowse() async {
        let store = makeStore()
        seedArtifact(store, count: 1, pdf: true)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.locator.isArtifactOnly)
        model.goToNextPage()
        #expect(model.locator.page == nil)
        #expect(model.setPageFromViewer())
        #expect(model.locator.page == model.artifactViewer.page)
        #expect(model.isLocatorPageSetOnViewer)
        model.resetToEntireArtifact()
        #expect(model.locator.isArtifactOnly)
    }

    @Test func pdfRegionAutoInsertsCurrentPage() async {
        let store = makeStore()
        seedArtifact(store, count: 1, pdf: true)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.setRegion(sampleRectangle()))
        #expect(model.locator.page == model.artifactViewer.page)
        #expect(model.locator.region?.kind == .rectangle)
        #expect(model.locator.encodeJSON().contains("\"page\""))
        #expect(model.locator.encodeJSON().contains("\"region\""))
    }

    @Test func imageRegionHasNoPage() async {
        let store = makeStore()
        seedArtifact(store, pdf: false)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(!model.setPageFromViewer())
        #expect(model.setRegion(sampleRectangle()))
        #expect(model.locator.page == nil)
        #expect(model.locator.region != nil)
        #expect(!model.locator.encodeJSON().contains("\"page\""))
    }

    @Test func audioRefusesPageAndRegion() async {
        let store = makeStore()
        seedArtifact(store, mediaType: "audio/mpeg")
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.artifactViewer.kind == .audio)
        #expect(!model.setPage(2))
        #expect(!model.setRegion(sampleRectangle()))
        #expect(model.locator.isArtifactOnly)
        #expect(model.locator.encodeJSON().contains("\"artifact\""))
    }

    @Test func switchingKindPeelsIllegalLayers() async {
        let store = makeStore()
        seedArtifact(store, count: 2, pdf: true)
        if var artifacts = store.artifactsBySource[sourceID], artifacts.indices.contains(1) {
            artifacts[1].file?.mediaType = "image/jpeg"
            store.artifactsBySource[sourceID] = artifacts
        }
        let model = makeModel(store: store)
        await model.prepare()
        model.selectPendingArtifact("art-0")
        await model.confirmArtifactSelectionAndLoad()
        #expect(model.setPage(2))
        #expect(model.setRegion(sampleRectangle()))
        #expect(model.locator.page == 2)
        #expect(model.locator.region != nil)

        model.selectPendingArtifact("art-1")
        await model.confirmArtifactSelectionAndLoad()
        #expect(model.locator.isArtifactOnly)

        var draft = CitationLocatorDraft.artifactOnly()
        let setPage = draft.setPage(3, capabilities: ArtifactViewerKind.pdf.locatorCapabilities)
        let setRegion = draft.setRegion(
            sampleRectangle(),
            capabilities: ArtifactViewerKind.pdf.locatorCapabilities,
            autoPage: 3
        )
        #expect(setPage)
        #expect(setRegion)
        draft.peelIllegalLayers(capabilities: ArtifactViewerKind.image.locatorCapabilities)
        #expect(draft.page == nil)
        #expect(draft.region != nil)
        draft.peelIllegalLayers(capabilities: ArtifactViewerKind.audio.locatorCapabilities)
        #expect(draft.page == nil)
        #expect(draft.region == nil)
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
        #expect(model.locator.encodeJSON().contains("\"artifact\""))
        #expect(model.locator.encodeJSON().contains("\"page\""))
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
        #expect(store.citationsByID[citation.id]?.locatorJSON.contains("\"artifact\"") == true)
        #expect(listed.count == 1)
        #expect(listed[0].valueText == "Miller")
    }

    @Test func submitConnectPrefillCreatesCitedBridge() async throws {
        let store = makeStore()
        seedArtifact(store)
        let eventTypeID = "type-event"
        store.subjectTypesByProject[projectDir]?.append(
            CatalogSubjectType(
                id: eventTypeID,
                key: "event",
                origin: "provenencia",
                label: "Event",
                description: "",
                refPrefix: "EVT",
                candidateRefPrefix: "CEV"
            )
        )
        let eventID = "sub-event-1"
        store.subjectsBySource[sourceID] = [
            CatalogSubject(
                id: subjectID,
                ref: "CPR-1",
                sourceID: sourceID,
                subjectTypeID: personTypeID,
                label: "Margt.",
                description: ""
            ),
            CatalogSubject(
                id: eventID,
                ref: "CEV-1",
                sourceID: sourceID,
                subjectTypeID: eventTypeID,
                label: "Birth",
                description: ""
            ),
        ]
        store.subjectPositionsBySubject[subjectID] = CatalogSubjectPosition(
            subjectID: subjectID, gridX: 0, gridY: 0
        )
        store.subjectPositionsBySubject[eventID] = CatalogSubjectPosition(
            subjectID: eventID, gridX: 2, gridY: 0
        )
        let before = (store.subjectsBySource[sourceID] ?? []).count
        let model = makeModel(
            store: store,
            subjectID: "",
            connectFromSubjectID: subjectID,
            connectToSubjectID: eventID,
            connectBridgeTypeKey: "participation"
        )
        await model.prepare()
        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        let after = store.subjectsBySource[sourceID] ?? []
        #expect(after.count == before + 1)
        #expect(after.contains(where: { $0.subjectTypeID == participationTypeID }))
        #expect(!store.citationsByID.isEmpty)
    }

    @Test func fakeStoreRefusesPersonPlaceCitedBridge() async {
        let store = makeStore()
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
        store.subjectsBySource[sourceID] = [
            CatalogSubject(
                id: subjectID,
                ref: "CPR-1",
                sourceID: sourceID,
                subjectTypeID: personTypeID,
                label: "Alice",
                description: ""
            ),
            CatalogSubject(
                id: "pl1",
                ref: "CPL-1",
                sourceID: sourceID,
                subjectTypeID: "type-place",
                label: "Leeds",
                description: ""
            ),
        ]
        do {
            _ = try await store.createCitedBridge(
                projectDir: projectDir,
                userID: "user-1",
                sourceID: sourceID,
                fromSubjectID: subjectID,
                toSubjectID: "pl1",
                bridgeTypeKey: "",
                label: "Nope",
                description: "",
                gridX: 0,
                gridY: 0,
                artifactID: "art-0",
                locatorJSON: "{}",
                transcription: "",
                citationDescription: "",
                transcriptionUncertain: false,
                transcriptionNote: "",
                citationNotes: [],
                observations: []
            )
            Issue.record("expected refuse")
        } catch let error as CoreInvokeError {
            guard case .coded(_, let code, _, _) = error else {
                Issue.record("expected coded error")
                return
            }
            #expect(code == "connect.refused")
        } catch {
            Issue.record("expected CoreInvokeError")
        }
    }

    @Test func prepareFailureIsNotTheEmptyArtifactGate() async {
        let store = makeStore()
        store.listSubjectsError = CoreInvokeError.failed(status: 1)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .loadFailed)
        #expect(model.loadError != nil)
        #expect(model.hasNoArtifacts == false)
    }

    @Test func customTermFailureStaysOnTheTermDialog() async {
        let store = makeStore()
        store.createPropertyTermError = CoreInvokeError.failed(status: 1)
        let model = makeModel(store: store)
        let term = await model.createCustomTerm(propertyID: occupationPropertyID, label: "Farmer")
        #expect(term == nil)
        #expect(model.termError != nil)
        #expect(model.formError == nil)
    }

    @Test func observationIntegerDraftRejectsWords() {
        var draft = CatalogObservationDraft(subjectID: "s", propertyID: "p")
        let failure = CitationObservationValue.apply(
            valueType: "integer",
            fields: CitationObservationValue.Fields(valueIntegerText: "nope"),
            to: &draft
        )
        #expect(failure == .invalidInteger)
        #expect(draft.valueInteger == nil)
    }
}
