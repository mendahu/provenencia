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

    private func seedCitations(_ store: FakeStore, artifactID: String, count: Int) {
        for i in 0..<count {
            let id = "cit-\(artifactID)-\(i)"
            store.citationsByID[id] = CatalogCitation(
                id: id,
                ref: "CIT-\(i)",
                artifactID: artifactID,
                locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
                transcription: "",
                description: "",
                transcriptionUncertain: false,
                transcriptionNote: ""
            )
        }
    }

    private func setSourceThumbnail(_ store: FakeStore, artifactID: String) {
        guard var sources = store.sourcesByProject[projectDir],
              let index = sources.firstIndex(where: { $0.id == sourceID })
        else { return }
        sources[index].coverMode = "artifact"
        sources[index].primaryArtifactID = artifactID
        store.sourcesByProject[projectDir] = sources
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

    @Test func prepareMultiArtifactOpensComposeWithDefault() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-0")
        #expect(model.showsArtifactSwitcher)
        #expect(model.observations.isEmpty)
        await model.selectArtifactAndLoad("art-1")
        #expect(model.selectedArtifactID == "art-1")
    }

    @Test func prepareDefaultsToFirstArtifactWithFile() async {
        let store = makeStore()
        seedArtifact(store, count: 3)
        stripAttachedFile(store, artifactID: "art-0")
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-1")
    }

    @Test func prepareFallsBackToFirstWhenNoneHaveFiles() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        stripAttachedFile(store, artifactID: "art-0")
        stripAttachedFile(store, artifactID: "art-1")
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-0")
    }

    @Test func prepareDefaultsToMostCitedArtifact() async {
        let store = makeStore()
        seedArtifact(store, count: 3)
        seedCitations(store, artifactID: "art-0", count: 1)
        seedCitations(store, artifactID: "art-2", count: 3)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-2")
    }

    @Test func prepareCitationTiePrefersThumbnail() async {
        let store = makeStore()
        seedArtifact(store, count: 3)
        seedCitations(store, artifactID: "art-0", count: 2)
        seedCitations(store, artifactID: "art-2", count: 2)
        setSourceThumbnail(store, artifactID: "art-2")
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.selectedArtifactID == "art-2")
    }

    @Test func defaultPendingArtifactIDLayers() {
        func artifact(_ id: String, file: Bool = true) -> CatalogArtifact {
            CatalogArtifact(
                id: id,
                ref: id,
                sourceID: sourceID,
                fileID: file ? "file-\(id)" : "",
                label: id,
                description: "",
                file: file
                    ? CatalogFileRef(
                        id: "file-\(id)",
                        relPath: "objects/\(id).pdf",
                        originalFilename: "\(id).pdf",
                        mediaType: "application/pdf",
                        byteSize: 10
                    )
                    : nil
            )
        }

        #expect(
            CitationComposerModel.defaultPendingArtifactID(
                artifacts: [artifact("art-0"), artifact("art-1"), artifact("art-2")],
                citationCounts: ["art-0": 1, "art-1": 4, "art-2": 2],
                selectedID: nil,
                coverMode: "type_icon",
                primaryArtifactID: ""
            ) == "art-1"
        )
        #expect(
            CitationComposerModel.defaultPendingArtifactID(
                artifacts: [artifact("art-0"), artifact("art-1")],
                citationCounts: ["art-0": 2, "art-1": 2],
                selectedID: nil,
                coverMode: "artifact",
                primaryArtifactID: "art-1"
            ) == "art-1"
        )
        #expect(
            CitationComposerModel.defaultPendingArtifactID(
                artifacts: [artifact("art-0"), artifact("art-1")],
                citationCounts: [:],
                selectedID: nil,
                coverMode: "artifact",
                primaryArtifactID: "art-1"
            ) == "art-1"
        )
        #expect(
            CitationComposerModel.defaultPendingArtifactID(
                artifacts: [artifact("art-0", file: false), artifact("art-1"), artifact("art-2")],
                citationCounts: [:],
                selectedID: nil,
                coverMode: "type_icon",
                primaryArtifactID: ""
            ) == "art-1"
        )
        #expect(
            CitationComposerModel.defaultPendingArtifactID(
                artifacts: [artifact("art-0", file: false), artifact("art-1", file: false)],
                citationCounts: [:],
                selectedID: nil,
                coverMode: "type_icon",
                primaryArtifactID: ""
            ) == "art-0"
        )
        #expect(
            CitationComposerModel.defaultPendingArtifactID(
                artifacts: [artifact("art-0"), artifact("art-1")],
                citationCounts: ["art-0": 1, "art-1": 3],
                selectedID: "art-0",
                coverMode: "type_icon",
                primaryArtifactID: ""
            ) == "art-0"
        )
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

    @Test func addObservationCommitsInlineTextRow() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        #expect(model.observationDialog == nil)
        #expect(model.observations.count == 1)
        let id = model.observations[0].id
        model.updateObservationProperty(id: id, propertyID: occupationPropertyID)
        model.updateObservationText(id: id, text: "Farmer")
        #expect(model.observations[0].valueText == "Farmer")
        #expect(model.observations[0].subjectID == subjectID)
    }

    @Test func observationDialogCommitsNameValueParts() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        let id = model.observations[0].id
        model.updateObservationProperty(id: id, propertyID: namePropertyID)
        model.beginEditObservation(model.observations[0])
        var draft = model.observationDialog!
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

    @Test func submitEmptyObservationsPersistsReading() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.phase == .compose)
        #expect(model.locator.isArtifactOnly)
        #expect(model.hasLocator)
        model.transcription = "transcribe first"

        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        #expect(location?.sourceId == sourceID)
        #expect(store.citationsByID.values.contains(where: { $0.transcription == "transcribe first" }))
        let listed = try await store.listObservationsBySource(
            projectDir: projectDir,
            sourceID: sourceID
        )
        #expect(listed.isEmpty)
        let json = store.citationsByID.values.first?.locatorJSON ?? ""
        #expect(json.contains("\"artifact\""))
        #expect(!json.contains("\"page\""))
    }

    @Test func submitWritesObservationForEntrySubject() async throws {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        model.beginAddObservation()
        let id = model.observations[0].id
        model.updateObservationProperty(id: id, propertyID: occupationPropertyID)
        model.updateObservationText(id: id, text: "Farmer")
        let location = await model.submit()
        #expect(location?.sourceSurface == .graph)
        let listed = try await store.listObservationsBySource(
            projectDir: projectDir,
            sourceID: sourceID
        )
        #expect(listed.count == 1)
        #expect(listed[0].valueText == "Farmer")
        #expect(listed[0].propertyID == occupationPropertyID)
        #expect(listed[0].subjectID == subjectID)
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
        #expect(model.selectedArtifactID == "art-0")
        #expect(model.setPage(2))
        #expect(model.setRegion(sampleRectangle()))
        #expect(model.locator.page == 2)
        #expect(model.locator.region != nil)

        await model.selectArtifactAndLoad("art-1")
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
        model.updateObservationText(id: model.observations[0].id, text: "Miller")
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

    @Test func prepareLoadsAllSubjectsOnSharedCitation() async {
        let store = makeStore()
        seedArtifact(store)
        let otherID = "sub-person-2"
        store.subjectsBySource[sourceID]?.append(
            CatalogSubject(
                id: otherID,
                ref: "CPR-2",
                sourceID: sourceID,
                subjectTypeID: personTypeID,
                label: "Thomas",
                description: ""
            )
        )
        store.subjectPositionsBySubject[otherID] = CatalogSubjectPosition(
            subjectID: otherID, gridX: 1, gridY: 0
        )
        let citation = CatalogCitation(
            id: "cit-shared",
            ref: "CIT-44",
            artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "household",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.citationsByID[citation.id] = citation
        store.observationsBySource[sourceID] = [
            CatalogObservation(
                id: "obs-m",
                ref: "OBS-1",
                citationID: citation.id,
                subjectID: subjectID,
                propertyID: "prop-age",
                polarity: "positive",
                valueText: "",
                valueInteger: 52,
                valueDateID: "",
                valueNameID: "",
                valueSubjectID: "",
                valueTermID: "",
                propertyKey: "age",
                propertyLabel: "Age",
                propertyValueType: "integer"
            ),
            CatalogObservation(
                id: "obs-t",
                ref: "OBS-2",
                citationID: citation.id,
                subjectID: otherID,
                propertyID: occupationPropertyID,
                polarity: "positive",
                valueText: "Blacksmith",
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
        #expect(model.observations.count == 2)
        #expect(Set(model.observations.map(\.subjectID)) == [subjectID, otherID])
        #expect(model.focusedObservationID == model.observations.first(where: { $0.subjectID == subjectID })?.id)
    }

    @Test func addPropertyCanReuseExistingCitation() async {
        let store = makeStore()
        seedArtifact(store)
        let otherID = "sub-person-2"
        store.subjectsBySource[sourceID]?.append(
            CatalogSubject(
                id: otherID,
                ref: "CPR-2",
                sourceID: sourceID,
                subjectTypeID: personTypeID,
                label: "Thomas",
                description: ""
            )
        )
        store.subjectPositionsBySubject[otherID] = CatalogSubjectPosition(
            subjectID: otherID, gridX: 1, gridY: 0
        )
        let citation = CatalogCitation(
            id: "cit-reuse",
            ref: "CIT-43",
            artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "Margt. Alderwick",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.citationsByID[citation.id] = citation
        store.observationsBySource[sourceID] = [
            CatalogObservation(
                id: "obs-t",
                ref: "OBS-2",
                citationID: citation.id,
                subjectID: otherID,
                propertyID: occupationPropertyID,
                polarity: "positive",
                valueText: "Blacksmith",
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
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.activeCitationID == nil)
        await model.selectCitationAndLoad(citation.id)
        #expect(model.activeCitationID == citation.id)
        #expect(model.transcription == "Margt. Alderwick")
        #expect(model.observations.contains(where: { $0.subjectID == otherID }))
        #expect(model.observations.contains(where: { $0.subjectID == subjectID && $0.propertyID.isEmpty }))
    }

    @Test func dirtyArtifactChangeOnSavedCitationConfirms() async {
        let store = makeStore()
        seedArtifact(store, count: 2)
        let citation = CatalogCitation(
            id: "cit-dirty",
            ref: "CIT-44",
            artifactID: "art-0",
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
            transcription: "saved",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.citationsByID[citation.id] = citation
        let model = makeModel(store: store, citationID: citation.id)
        await model.prepare()
        #expect(model.isEditingExisting)
        #expect(!model.isDirty)
        await model.selectArtifactAndLoad("art-1")
        #expect(model.pendingArtifactAbandon == nil)
        #expect(model.activeCitationID == nil)
        #expect(model.selectedArtifactID == "art-1")

        let model2 = makeModel(store: store, citationID: citation.id)
        await model2.prepare()
        model2.transcription = "edited"
        #expect(model2.isDirty)
        await model2.selectArtifactAndLoad("art-1")
        #expect(model2.pendingArtifactAbandon?.targetArtifactID == "art-1")
        #expect(model2.selectedArtifactID == "art-0")
        model2.cancelAbandonArtifact()
        #expect(model2.pendingArtifactAbandon == nil)
        model2.transcription = "edited"
        await model2.selectArtifactAndLoad("art-1")
        await model2.confirmAbandonArtifactAndLoad()
        #expect(model2.selectedArtifactID == "art-1")
        #expect(model2.activeCitationID == nil)
    }

    @Test func graphLocationReturnsToEvidenceGraph() async {
        let store = makeStore()
        seedArtifact(store)
        let model = makeModel(store: store)
        await model.prepare()
        #expect(model.graphLocation().sourceSurface == .graph)
        #expect(model.graphLocation().sourceId == sourceID)
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
