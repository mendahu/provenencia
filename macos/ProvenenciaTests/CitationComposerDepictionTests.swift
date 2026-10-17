import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CitationComposerDepictionTests {
    private let sourceID = "src-1"
    private let adaID = "sub-ada"
    private let benID = "sub-ben"
    private let claraID = "sub-clara"
    private let personTypeID = "type-person"
    private let likenessID = "prop-likeness"
    private let occupationID = "prop-occupation"
    private let savedRelPath = "objects/ab/cd.png"

    // MARK: - Fixtures

    private func makeStore(projectDir: String, withLikeness: Bool = true) -> FakeStore {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: personTypeID, key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPER"
            ),
        ]
        var properties = [
            CatalogProperty(
                id: occupationID, key: "occupation", origin: "provenencia", label: "Occupation",
                description: "", valueType: "text"
            ),
        ]
        if withLikeness {
            properties.append(
                CatalogProperty(
                    id: likenessID, key: "likeness", origin: "provenencia", label: "Likeness",
                    description: "", valueType: "depiction"
                )
            )
        }
        store.propertiesByProject[projectDir] = properties
        store.subjectTypePropertiesByType[personTypeID] = properties.enumerated().map {
            CatalogSubjectTypeProperty(property: $0.element, sortOrder: $0.offset, locked: false)
        }
        store.connectRules = CatalogConnectRule.productMatrix
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: sourceID, ref: "SRC-1", sourceTypeID: "stype", title: "Wedding", description: ""),
        ]
        store.subjectsBySource[sourceID] = [
            person(adaID, ref: "CPR-1", label: "Ada Lovelace"),
            person(benID, ref: "CPR-2", label: "Ben Ford"),
            person(claraID, ref: "CPR-3", label: "Clara Ford"),
        ]
        for (index, id) in [adaID, benID, claraID].enumerated() {
            store.subjectPositionsBySubject[id] = CatalogSubjectPosition(
                subjectID: id, gridX: Int64(index * 8), gridY: 0
            )
        }
        return store
    }

    private func person(_ id: String, ref: String, label: String) -> CatalogSubject {
        CatalogSubject(
            id: id, ref: ref, sourceID: sourceID, subjectTypeID: personTypeID, label: label, description: ""
        )
    }

    private enum Media { case image, pdf, audio }

    /// A project folder with one Artifact of that kind, on disk where a viewer can load it.
    private func seeded(
        _ media: Media = .image,
        withLikeness: Bool = true
    ) throws -> (projectDir: String, store: FakeStore) {
        let projectDir = ArtifactFileFixtures.uniqueProjectDir()
        let store = makeStore(projectDir: projectDir, withLikeness: withLikeness)
        switch media {
        case .image:
            try ArtifactFileFixtures.writePNG(
                projectDir: projectDir, relPath: "objects/photo.png", width: 8, height: 8, topHalfRed: true
            )
            ArtifactFileFixtures.seedArtifact(store, mediaType: "image/png", relPath: "objects/photo.png")
        case .pdf:
            try ArtifactFileFixtures.writeBlankPDF(projectDir: projectDir, relPath: "objects/news.pdf", pageCount: 2)
            ArtifactFileFixtures.seedArtifact(store, mediaType: "application/pdf", relPath: "objects/news.pdf")
        case .audio:
            ArtifactFileFixtures.seedArtifact(store, mediaType: "audio/mpeg", relPath: "objects/voice.mp3")
        }
        return (projectDir, store)
    }

    private func makeModel(
        store: FakeStore,
        projectDir: String,
        subjectID: String? = nil,
        citationID: String? = nil,
        observationID: String? = nil
    ) -> CitationComposerModel {
        let entry: CitationComposerEntry
        if let citationID {
            entry = .edit(
                sourceID: sourceID,
                subjectID: subjectID ?? adaID,
                citationID: citationID,
                artifactID: "art-0",
                observationID: observationID
            )
        } else {
            entry = .addProperty(sourceID: sourceID, subjectID: subjectID ?? adaID, artifactID: "art-0")
        }
        return CitationComposerModel(
            entry: entry,
            session: WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store),
            store: store,
            userID: "user-1"
        )
    }

    private func rect(_ x0: CGFloat, _ y0: CGFloat, _ x1: CGFloat, _ y1: CGFloat) -> ArtifactRegionDraft {
        ArtifactRegionDraft(
            kind: .rectangle,
            points: ArtifactRegionGeometry.rectanglePoints(from: CGPoint(x: x0, y: y0), to: CGPoint(x: x1, y: y1))
        )
    }

    private func regionJSON(_ region: ArtifactRegionDraft?, page: Int? = nil) -> String {
        CitationLocatorDraft(page: page, region: region, extras: []).encodeJSON()
    }

    private func seedCitation(
        _ store: FakeStore,
        locatorJSON: String = #"{"version":1,"selectors":[{"type":"artifact"}]}"#,
        observations: [CatalogObservation] = []
    ) {
        store.citationsByID["cit-1"] = CatalogCitation(
            id: "cit-1",
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: locatorJSON,
            transcription: "",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.observationsBySource[sourceID] = observations
    }

    private func likenessObservation(
        id: String = "obs-1",
        subjectID: String? = nil,
        polarity: String = ObservationPolarity.positive.rawValue,
        depictionID: String = "dep-1",
        region: String
    ) -> CatalogObservation {
        CatalogObservation(
            id: id, ref: "OBS-L2BEN", citationID: "cit-1", subjectID: subjectID ?? benID,
            propertyID: likenessID, polarity: polarity,
            valueText: "", valueInteger: nil, valueDateID: "", valueNameID: "", valueSubjectID: "",
            valueTermID: "", propertyKey: "likeness", propertyLabel: "Likeness",
            propertyValueType: "depiction",
            valueDepictionID: depictionID,
            depiction: CatalogDepictionValue(regionJSON: region, imageRelPath: savedRelPath)
        )
    }

    /// A model on a photo with a saved likeness of Ben, its edit dialog open.
    private func editingSavedLikeness(
        polarity: String = ObservationPolarity.positive.rawValue
    ) async throws -> (CitationComposerModel, FakeStore) {
        let (projectDir, store) = try seeded()
        seedCitation(
            store,
            observations: [
                likenessObservation(polarity: polarity, region: regionJSON(rect(0.1, 0.1, 0.4, 0.4))),
            ]
        )
        let model = makeModel(store: store, projectDir: projectDir, citationID: "cit-1", observationID: "obs-1")
        await model.prepare()
        return (model, store)
    }

    private func openLikenessDraft(_ model: CitationComposerModel, subjectID: String? = nil) {
        model.beginPropertyDraft(subjectID: subjectID ?? adaID)
        model.mutateProperty { $0.values.propertyID = likenessID }
    }

    private func draftValues(_ model: CitationComposerModel) -> ObservationRowValues? {
        if case .property(let draft) = model.authoring { return draft.values }
        return nil
    }

    private func editValues(_ model: CitationComposerModel) -> ObservationRowValues? {
        if case .property(let values, _, _) = model.editSession?.kind { return values }
        return nil
    }

    private func isMissing(_ paths: [String]) -> Bool {
        !paths.isEmpty && paths.allSatisfy { !FileManager.default.fileExists(atPath: $0) }
    }

    // MARK: - T-1 … T-3

    @Test func depictionIsOfferedOnImageAndPDFOnly() async throws {
        for media in [Media.image, .pdf] {
            let (projectDir, store) = try seeded(media)
            let model = makeModel(store: store, projectDir: projectDir)
            await model.prepare()
            let options = model.vocabulary.propertyOptions(
                forSubjectID: adaID, capabilities: model.depictionCapabilities
            )
            #expect(options.contains { $0.key == "likeness" })
        }
        let (audioDir, audioStore) = try seeded(.audio)
        let audio = makeModel(store: audioStore, projectDir: audioDir)
        await audio.prepare()
        let audioOptions = audio.vocabulary.propertyOptions(
            forSubjectID: adaID, capabilities: audio.depictionCapabilities
        )
        #expect(!audioOptions.contains { $0.key == "likeness" })
        #expect(audioOptions.contains { $0.key == "occupation" })

        let (bareDir, bareStore) = try seeded(.image, withLikeness: false)
        let bare = makeModel(store: bareStore, projectDir: bareDir)
        await bare.prepare()
        let bareOptions = bare.vocabulary.propertyOptions(
            forSubjectID: adaID, capabilities: bare.depictionCapabilities
        )
        #expect(!bareOptions.contains { $0.key == "likeness" })
    }

    @Test func loadedDepictionKeepsItsValue() throws {
        let region = rect(0.1, 0.2, 0.5, 0.6)
        var observation = likenessObservation(region: regionJSON(region, page: 2))
        let row = ObservationRow.loaded(from: observation)
        #expect(row.depictionDraft.locator.region == region)
        #expect(row.depictionDraft.locator.page == 2)
        #expect(row.depictionDraft.valueDepictionID == "dep-1")
        #expect(row.depictionDraft.imageRelPath == savedRelPath)
        #expect(!row.depictionDraft.regionHidden)
        let property = CatalogProperty(
            id: likenessID, key: "likeness", origin: "provenencia", label: "Likeness",
            description: "", valueType: "depiction"
        )
        let rebuilt = try #require(row.asCatalogObservation(property: property).0)
        #expect(rebuilt.valueDepictionID == "dep-1")

        observation.depiction = CatalogDepictionValue(regionJSON: "", imageRelPath: savedRelPath)
        let hidden = ObservationRow.loaded(from: observation)
        #expect(hidden.depictionDraft.regionHidden)
    }

    @Test func depictionNeedsARegionToSave() async throws {
        let (projectDir, store) = try seeded()
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        openLikenessDraft(model)
        #expect(!model.authoringCanConfirm)
        model.mutateProperty { $0.values.depictionDraft.locator.region = rect(0.25, 0.25, 0.75, 0.75) }
        #expect(model.authoringCanConfirm)
    }

    // MARK: - T-8

    @Test func fakeStoreReturnsTheDepiction() async throws {
        let (projectDir, store) = try seeded()
        seedCitation(store)
        let json = regionJSON(rect(0.1, 0.1, 0.4, 0.4))
        var draft = CatalogObservationDraft(
            subjectID: adaID, propertyID: likenessID, polarity: ObservationPolarity.positive.rawValue
        )
        draft.depiction = CatalogDepictionValueInput(regionJSON: json, imagePath: "/nonexistent.png")
        let written = try await store.addObservationsToCitation(
            projectDir: projectDir, userID: "user-1", citationID: "cit-1", observations: [draft]
        )
        let saved = try #require(written.0.first)
        #expect(saved.depiction?.regionJSON == json)
        #expect(saved.depiction?.imageRelPath.hasPrefix("objects/") == true)
        #expect(!saved.valueDepictionID.isEmpty)

        let updated = try await store.updateObservation(
            projectDir: projectDir, userID: "user-1", observation: saved
        )
        #expect(updated.0.valueDepictionID == saved.valueDepictionID)
        #expect(updated.0.depiction == saved.depiction)
    }

    // MARK: - Drawing (T-9 … T-12, T-20)

    @Test func drawingLeavesTheCitationRegion() async throws {
        let (projectDir, store) = try seeded()
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        let citationRegion = rect(0.5, 0.5, 0.9, 0.9)
        #expect(model.setRegion(citationRegion))
        openLikenessDraft(model)
        #expect(model.depictionPreview == nil)
        model.startDepictionDrawing()
        #expect(model.presented == .drawingDepiction(over: .authoring))
        #expect(model.armedRegionTool == .rectangle)
        let face = rect(0.25, 0.25, 0.75, 0.75)
        #expect(model.commitViewerRegion(face))
        #expect(draftValues(model)?.depictionDraft.locator.region == face)
        #expect(model.locator.region == citationRegion)
        #expect(model.presented == .authoring)
        #expect(model.viewerRegion == face)
        #expect(model.depictionPreview?.size == CGSize(width: 4, height: 4))

        model.startDepictionDrawing()
        model.cancelDepictionDrawing()
        #expect(draftValues(model)?.depictionDraft.locator.region == face)
        #expect(model.presented == .authoring)
        #expect(model.armedRegionTool == nil)
    }

    @Test func newPDFCitationTakesTheDrawnPage() async throws {
        let (projectDir, store) = try seeded(.pdf)
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        model.artifactViewer.setPage(2)
        openLikenessDraft(model, subjectID: benID)
        model.startDepictionDrawing()
        #expect(model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75)))
        #expect(model.locator.page == 2)
        #expect(draftValues(model)?.depictionDraft.locator.page == 2)
    }

    @Test func existingPDFCitationKeepsItsPage() async throws {
        let (projectDir, store) = try seeded(.pdf)
        seedCitation(
            store,
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"},{"type":"page","artifact_page":1}]}"#
        )
        let model = makeModel(store: store, projectDir: projectDir, citationID: "cit-1")
        await model.prepare()
        model.artifactViewer.setPage(2)
        openLikenessDraft(model)
        model.startDepictionDrawing()
        #expect(model.artifactViewer.page == 1)
        model.artifactViewer.setPage(2)
        #expect(model.depictionWrongPage)
        #expect(!model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75)))
        #expect(model.depictionWrongPage)
        #expect(draftValues(model)?.depictionDraft.hasRegion == false)
        #expect(model.presented == .drawingDepiction(over: .authoring))
        model.artifactViewer.setPage(1)
        #expect(!model.depictionWrongPage)
        #expect(model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75)))

        let (bareDir, bareStore) = try seeded(.pdf)
        seedCitation(bareStore)
        let bare = makeModel(store: bareStore, projectDir: bareDir, citationID: "cit-1")
        await bare.prepare()
        openLikenessDraft(bare)
        #expect(bare.depictionNeedsPage)
        bare.startDepictionDrawing()
        #expect(bare.presented == .authoring)
    }

    @Test func viewerToolIsUnarmedOnTheWrongPage() async throws {
        let (projectDir, store) = try seeded(.pdf)
        seedCitation(
            store,
            locatorJSON: #"{"version":1,"selectors":[{"type":"artifact"},{"type":"page","artifact_page":1}]}"#
        )
        let model = makeModel(store: store, projectDir: projectDir, citationID: "cit-1")
        await model.prepare()
        openLikenessDraft(model)
        model.startDepictionDrawing()
        #expect(model.viewerArmedRegionTool == .rectangle)
        model.artifactViewer.setPage(2)
        #expect(model.depictionWrongPage)
        #expect(model.viewerArmedRegionTool == nil)
        #expect(model.armedRegionTool == .rectangle)
        model.artifactViewer.setPage(1)
        #expect(model.viewerArmedRegionTool == .rectangle)
        model.disarmRegionTool()
        #expect(model.presented == .authoring)
        #expect(model.armedRegionTool == nil)
    }

    @Test func tabFromPropertySkipsADisabledDraw() {
        #expect(
            CitationAuthoringNavigation.tabDestination(
                from: .property, committed: likenessID, via: .tab,
                draftAfterCommit: CitationAuthoringDraftAfterCommit()
            ) == .value
        )
        #expect(
            CitationAuthoringNavigation.tabDestination(
                from: .property, committed: likenessID, via: .tab,
                draftAfterCommit: CitationAuthoringDraftAfterCommit(propertyValueTakesFocus: false)
            ) == nil
        )
    }

    @Test func unsavedDepictionHoldsLeave() async throws {
        let (projectDir, store) = try seeded()
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        openLikenessDraft(model)
        model.startDepictionDrawing()
        #expect(model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75)))
        #expect(model.shouldHoldNavigation(.back))
        model.keepEditingAfterLeave()
        #expect(model.presented == .authoring)
    }

    @Test func keepEditingAfterACitationSwitchKeepsTheRegion() async throws {
        let (projectDir, store) = try seeded()
        seedCitation(store)
        var second = try #require(store.citationsByID["cit-1"])
        second.id = "cit-2"
        second.ref = "CIT-2"
        store.citationsByID["cit-2"] = second
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        model.setAuthoringCitation("cit-1")
        await model.awaitIdentitySwitch()
        openLikenessDraft(model)
        model.startDepictionDrawing()
        let face = rect(0.25, 0.25, 0.75, 0.75)
        #expect(model.commitViewerRegion(face))
        model.transcription = "unsaved note"

        model.setAuthoringCitation("cit-2")
        #expect(model.pendingLeave?.keepDraft == true)
        #expect(draftValues(model)?.depictionDraft.locator.region == face)
        model.keepEditingAfterLeave()
        #expect(model.activeCitationID == "cit-1")
        #expect(draftValues(model)?.depictionDraft.locator.region == face)

        model.setAuthoringCitation("cit-2")
        model.discardLeaveChanges()
        await model.awaitIdentitySwitch()
        #expect(model.activeCitationID == "cit-2")
        #expect(draftValues(model)?.depictionDraft.hasRegion == false)
    }

    @Test func drawingDisablesCitationControlsAndLeaveReturns() async throws {
        let (model, _) = try await editingSavedLikeness()
        let line = try #require(model.subjectGroups.flatMap(\.lines).first { $0.observationID == "obs-1" })
        model.discardEdit()
        openLikenessDraft(model, subjectID: adaID)
        model.startDepictionDrawing()
        #expect(model.isDrawingDepiction)
        let before = model.authoring

        model.beginEdit(line: line)
        model.duplicate(line: line)
        await model.askDeleteLine(line)
        model.beginAddObservation()
        model.selectCitation(nil)
        #expect(model.editSession == nil)
        #expect(model.authoring == before)
        #expect(model.pendingImpact == nil)
        #expect(model.presented == .drawingDepiction(over: .authoring))
        #expect(!model.setPageFromViewer())

        let locatorBefore = model.locator
        let artifactBefore = model.selectedArtifactID
        let transcriptionBefore = model.transcription
        model.clearRegion()
        model.requestSelectArtifact("art-other")
        model.requestAutoTranscribe()
        model.requestPasteTranscription()
        #expect(model.locator == locatorBefore)
        #expect(model.selectedArtifactID == artifactBefore)
        #expect(model.pendingTranscriptionConfirm == nil)
        #expect(model.transcription == transcriptionBefore)

        #expect(model.shouldHoldNavigation(.back))
        #expect(model.pendingLeave != nil)
        model.keepEditingAfterLeave()
        #expect(model.presented == .drawingDepiction(over: .authoring))
    }

    // MARK: - Save (T-13 … T-15)

    @Test func savingSendsAPNGAndDeletesIt() async throws {
        let (projectDir, store) = try seeded()
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        openLikenessDraft(model)
        model.startDepictionDrawing()
        #expect(model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75)))
        #expect(model.transcription.isEmpty)
        await model.saveAuthoring()

        let call = try #require(store.recordedCalls.last { $0.hasPrefix("addObservationsToCitation") })
        #expect(call.hasSuffix(" depiction=1 depictionFile=present"))
        let saved = try #require(store.observationsBySource[sourceID]?.last?.depiction)
        let sent = CitationLocatorDraft.decode(saved.regionJSON)
        #expect(sent.page == nil)
        #expect(sent.region != nil)
        let selectors = try JSONSerialization.jsonObject(with: Data(saved.regionJSON.utf8)) as? [String: Any]
        let types = (selectors?["selectors"] as? [[String: Any]])?.compactMap { $0["type"] as? String }
        #expect(types?.filter { $0 == "region" }.count == 1)
        #expect(types?.contains("page") == false)
        #expect(store.recordedDepictionPaths.count == 1)
        #expect(isMissing(store.recordedDepictionPaths))
    }

    @Test func nextFaceKeepsThePropertyAndClearsThePerson() async throws {
        let (projectDir, store) = try seeded()
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        openLikenessDraft(model)
        model.startDepictionDrawing()
        _ = model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))
        await model.saveAuthoring()

        let values = try #require(draftValues(model))
        #expect(values.propertyID == likenessID)
        #expect(values.subjectID.isEmpty)
        #expect(!values.depictionDraft.hasRegion)
        #expect(model.presented == .authoring)
        #expect(
            CitationAuthoringNavigation.initialPropertyFocus(
                subjectID: values.subjectID,
                propertyID: values.propertyID,
                isDepiction: model.vocabulary.isDepiction(propertyID: values.propertyID)
            ) == .subject
        )
        let options = model.vocabulary.propertyOptions(
            forSubjectID: values.subjectID,
            keeping: values.propertyID,
            capabilities: model.depictionCapabilities
        )
        #expect(options.first { $0.id == likenessID }?.label == "Likeness")
        var next = values
        model.applyObservationSubject(benID, field: .propertySubject, values: &next)
        #expect(next.subjectID == benID)
        #expect(next.propertyID == likenessID)
    }

    @Test func refusedSaveKeepsTheRegion() async throws {
        let (projectDir, store) = try seeded()
        let refusal = CoreInvokeError.coded(
            status: 1, code: "ingest.unsupported_image", kind: .user, params: []
        )
        store.addObservationsToCitationError = refusal
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        openLikenessDraft(model)
        model.startDepictionDrawing()
        let face = rect(0.25, 0.25, 0.75, 0.75)
        _ = model.commitViewerRegion(face)
        await model.saveAuthoring()

        guard case .property(let draft) = model.authoring else {
            Issue.record("the dialog should stay open")
            return
        }
        #expect(draft.error == L10n.Errors.message(for: refusal))
        #expect(draft.values.depictionDraft.locator.region == face)
        #expect(!draft.isSaving)
        #expect(model.presented == .authoring)
        #expect(isMissing(store.recordedDepictionPaths))
    }

    // MARK: - Edit and duplicate (T-16 … T-18, T-21)

    @Test func editWithoutRedrawKeepsTheImage() async throws {
        let (model, store) = try await editingSavedLikeness()
        model.mutateEdit { session in
            guard case .property(var values, let id, let ref) = session.kind else { return }
            values.subjectID = claraID
            session.kind = .property(values: values, persistedID: id, persistedRef: ref)
        }
        store.recordedCalls = []
        await model.saveEdit()
        let call = try #require(store.recordedCalls.first { $0.hasPrefix("updateObservation") })
        #expect(call.contains(" valueDepictionID=dep-1"))
        #expect(!call.contains(" depiction=1"))
        #expect(model.editSession == nil)
    }

    @Test func redrawSendsANewImage() async throws {
        let (model, store) = try await editingSavedLikeness()
        model.startDepictionDrawing()
        #expect(model.presented == .drawingDepiction(over: .edit))
        #expect(model.commitViewerRegion(rect(0.5, 0.5, 0.9, 0.9)))
        #expect(model.isDepictionRedrawn)
        store.recordedCalls = []
        await model.saveEdit()
        let call = try #require(store.recordedCalls.first { $0.hasPrefix("updateObservation") })
        #expect(call.hasSuffix(" depiction=1 depictionFile=present"))
        #expect(isMissing(store.recordedDepictionPaths))
    }

    @Test func duplicateCopiesTheRegionNotTheImage() async throws {
        let (model, store) = try await editingSavedLikeness(polarity: ObservationPolarity.negative.rawValue)
        let line = try #require(model.subjectGroups.flatMap(\.lines).first { $0.observationID == "obs-1" })
        model.discardEdit()
        store.recordedCalls = []
        model.duplicate(line: line)
        #expect(store.recordedCalls.isEmpty)
        let values = try #require(draftValues(model))
        #expect(values.subjectID == benID)
        #expect(values.propertyID == likenessID)
        #expect(values.polarity == ObservationPolarity.negative.rawValue)
        #expect(values.depictionDraft.locator.region == rect(0.1, 0.1, 0.4, 0.4))
        #expect(values.depictionDraft.valueDepictionID.isEmpty)
        #expect(values.depictionDraft.imageRelPath.isEmpty)
        #expect(
            CitationAuthoringNavigation.initialPropertyFocus(
                subjectID: values.subjectID,
                propertyID: values.propertyID,
                isDepiction: model.vocabulary.isDepiction(propertyID: values.propertyID)
            ) == .subject
        )
    }

    @Test func hiddenRegionAsksForARedraw() async throws {
        let (projectDir, store) = try seeded()
        seedCitation(store, observations: [likenessObservation(region: "")])
        let model = makeModel(store: store, projectDir: projectDir, citationID: "cit-1", observationID: "obs-1")
        await model.prepare()
        #expect(editValues(model)?.depictionDraft.regionHidden == true)
        #expect(!model.editCanConfirm)
        model.startDepictionDrawing()
        #expect(model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75)))
        #expect(editValues(model)?.depictionDraft.regionHidden == false)
        #expect(model.editCanConfirm)
    }

    // MARK: - Lines, copy (T-19, T-22, T-23)

    @Test func savedLineCarriesItsImage() async throws {
        let (model, _) = try await editingSavedLikeness()
        let line = try #require(model.subjectGroups.flatMap(\.lines).first { $0.observationID == "obs-1" })
        #expect(line.imageRelPath == savedRelPath)
        #expect(line.valueText.isEmpty)
    }

    @Test func captionFollowsTheDraft() async throws {
        let (projectDir, store) = try seeded()
        let model = makeModel(store: store, projectDir: projectDir)
        await model.prepare()
        openLikenessDraft(model)
        #expect(caption(model) == L10n.string(L10n.CitationComposer.depictionCaptionEmpty))
        model.mutateProperty { $0.values.depictionDraft.locator.region = rect(0.25, 0.25, 0.75, 0.75) }
        #expect(caption(model) == L10n.string(L10n.CitationComposer.depictionCaptionBox))

        let (editModel, _) = try await editingSavedLikeness()
        #expect(caption(editModel) == L10n.string(L10n.CitationComposer.depictionCaptionBox))
        editModel.startDepictionDrawing()
        _ = editModel.commitViewerRegion(rect(0.5, 0.5, 0.9, 0.9))
        #expect(caption(editModel) == L10n.string(L10n.CitationComposer.depictionCaptionRedrawn))

        let (pdfDir, pdfStore) = try seeded(.pdf)
        seedCitation(pdfStore)
        let pdf = makeModel(store: pdfStore, projectDir: pdfDir, citationID: "cit-1")
        await pdf.prepare()
        openLikenessDraft(pdf)
        #expect(caption(pdf) == L10n.string(L10n.CitationComposer.depictionCaptionNeedsPage))

        let (hiddenDir, hiddenStore) = try seeded()
        seedCitation(hiddenStore, observations: [likenessObservation(region: "")])
        let hidden = makeModel(
            store: hiddenStore, projectDir: hiddenDir, citationID: "cit-1", observationID: "obs-1"
        )
        await hidden.prepare()
        #expect(caption(hidden) == L10n.string(L10n.CitationComposer.depictionCaptionHidden))
    }

    @Test func depictionCopyNamesThePropertyAndPerson() async throws {
        let (model, _) = try await editingSavedLikeness()
        let line = try #require(model.subjectGroups.flatMap(\.lines).first { $0.observationID == "obs-1" })
        model.discardEdit()
        openLikenessDraft(model, subjectID: claraID)
        model.startDepictionDrawing()
        _ = model.commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))
        #expect(
            model.unsavedSummary == L10n.CitationComposer.leaveDepictionDraft(
                property: "Likeness",
                person: "Clara Ford",
                ref: model.activeCitationRef
            )
        )

        model.discardAuthoring()
        model.duplicate(line: line)
        guard case .property(let draft) = model.authoring else {
            Issue.record("duplicate should open a draft")
            return
        }
        #expect(draft.note == L10n.CitationComposer.depictionDuplicatedFrom("Ben Ford", "Likeness"))
    }

    private func caption(_ model: CitationComposerModel) -> String {
        L10n.string(model.depictionCaption)
    }
}
