import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SourcePageModelTests {
    private let projectDir = "/tmp/source-page.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"
    private let sourceID = "src-1"

    private func photoType() -> CatalogSourceType {
        CatalogSourceType(
            id: "t1", key: "photograph", origin: "provenencia", label: "Photograph",
            description: ""
        )
    }

    private func seededGrades() -> [CatalogCredibilityGrade] {
        [
            CatalogCredibilityGrade(id: "g-low", key: "low_trust", origin: "provenencia", label: "Low trust", sortOrder: 1),
            CatalogCredibilityGrade(id: "g-std", key: "standard", origin: "provenencia", label: "Standard", sortOrder: 2),
            CatalogCredibilityGrade(id: "g-high", key: "high_trust", origin: "provenencia", label: "High trust", sortOrder: 3),
        ]
    }

    private func makeStore(
        source: CatalogSource? = nil,
        notes: [CatalogSourceNote] = [],
        artifacts: [CatalogArtifact] = [],
        credibility: CatalogCredibilityAssessment? = nil,
        metadata: [CatalogMetadataEntry] = [],
        fields: [CatalogMetadataField] = []
    ) -> FakeStore {
        let store = FakeStore()
        let src = source ?? CatalogSource(
            id: sourceID, ref: "SRC-AAAAA", sourceTypeID: "t1",
            title: "Family album", description: "Held by Mary"
        )
        store.sourcesByProject[projectDir] = [src]
        store.sourceTypesByProject[projectDir] = [photoType()]
        store.notesBySource[sourceID] = notes
        store.artifactsBySource[sourceID] = artifacts
        store.metadataBySource[sourceID] = metadata
        store.fieldsByProject[projectDir] = fields
        store.credibilityGradesByProject[projectDir] = seededGrades()
        if let credibility {
            store.credibilityBySource[sourceID] = credibility
        }
        return store
    }

    private func makeModel(store: FakeStore, session: WorkspaceSession? = nil) -> SourcePageModel {
        let session = session ?? WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        return SourcePageModel(
            sourceID: sourceID,
            session: session,
            userID: userID,
            sessionDisplayName: "Jake Robins",
            store: store
        )
    }

    @Test func loadPopulatesWorkspaceAndDefaultsCredibilityToStandard() async {
        let model = makeModel(store: makeStore())
        await model.warmFromSession()
        #expect(model.workspace?.source.ref == "SRC-AAAAA")
        #expect(model.identity.title == "Family album")
        #expect(model.credibility.savedKey == "standard")
        #expect(model.workspace?.credibility == nil)
        #expect(model.credibility.grades.count == 3)
    }

    @Test func openCopyUsesThePageSource() async {
        let model = makeModel(store: makeStore())
        model.openCopy()
        #expect(model.copying == nil)

        await model.warmFromSession()
        model.openCopy()
        #expect(model.copying?.original.id == sourceID)
        #expect(model.copying?.choices.draft.title == "Copy of Family album")
    }

    @Test func saveTitleUpdatesCommittedTitle() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditTitle()
        model.identity.titleDraft = "Revised title"
        await model.identity.saveTitle()
        #expect(model.identity.editingTitle == false)
        #expect(model.identity.title == "Revised title")
        #expect(model.workspace?.source.title == "Revised title")
        #expect(store.sourcesByProject[projectDir]?.first?.title == "Revised title")
    }

    @Test func saveTypeUpdatesCommittedType() async {
        let store = makeStore()
        store.sourceTypesByProject[projectDir] = [
            photoType(),
            CatalogSourceType(
                id: "t2", key: "census", origin: "provenencia", label: "Census", description: ""
            ),
        ]
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditType()
        model.identity.typeDraftID = "t2"
        await model.identity.saveType()
        #expect(model.identity.editingType == false)
        #expect(model.identity.sourceTypeID == "t2")
        #expect(model.identity.typeLabel == "Census")
        #expect(store.sourcesByProject[projectDir]?.first?.sourceTypeID == "t2")
    }

    @Test func cancelEditTypeRestoresSavedType() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditType()
        model.identity.typeDraftID = "t-other"
        model.identity.cancelEditType()
        #expect(model.identity.editingType == false)
        #expect(model.identity.typeDraftID == "t1")
        #expect(model.identity.sourceTypeID == "t1")
    }

    @Test func saveDescriptionUpdatesCommittedDescription() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditDescription()
        model.identity.descriptionDraft = "Revised notes"
        await model.identity.saveDescription()
        #expect(model.identity.editingDescription == false)
        #expect(model.identity.description == "Revised notes")
        #expect(store.sourcesByProject[projectDir]?.first?.description == "Revised notes")
    }

    @Test func cancelEditDescriptionDiscardsDraft() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditDescription()
        model.identity.descriptionDraft = "Should not stick"
        model.identity.cancelEditDescription()
        #expect(model.identity.editingDescription == false)
        #expect(model.identity.description == "Held by Mary")
        #expect(store.sourcesByProject[projectDir]?.first?.description == "Held by Mary")
    }

    @Test func updateSourceFailureSurfacesOnTitleError() async {
        let store = makeStore()
        store.updateSourceError = CoreInvokeError.coded(
            status: 1, code: "internal.unknown", kind: .internal, params: []
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditTitle()
        model.identity.titleDraft = "Nope"
        await model.identity.saveTitle()
        #expect(model.identity.titleError != nil)
        #expect(model.identity.editingTitle == true)
        #expect(model.identity.title == "Family album")
    }

    @Test func addNoteFailureSurfacesPageError() async {
        let store = makeStore()
        store.addSourceNoteError = CoreInvokeError.coded(
            status: 1, code: "internal.unknown", kind: .internal, params: []
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.notes.draft = "Will fail"
        await model.notes.add()
        #expect(model.pageError != nil)
        #expect(model.notes.items.isEmpty)
        #expect(model.notes.draft == "Will fail")
    }

    @Test func emptyTitleDraftSetsValidationError() async {
        let model = makeModel(store: makeStore())
        await model.warmFromSession()
        model.identity.beginEditTitle()
        model.identity.titleDraft = "   "
        await model.identity.saveTitle()
        #expect(model.identity.titleError != nil)
        #expect(model.identity.editingTitle == true)
        #expect(model.workspace?.source.title == "Family album")
    }

    @Test func selectCredibilityDraftDoesNotPersistUntilSave() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.credibility.selectDraft(key: "high_trust")
        #expect(model.credibility.isDirty)
        #expect(model.workspace?.credibility == nil)
        #expect(store.credibilityBySource[sourceID] == nil)
        await model.credibility.save()
        #expect(model.workspace?.credibility?.gradeKey == "high_trust")
        #expect(store.credibilityBySource[sourceID]?.gradeKey == "high_trust")
        #expect(!model.credibility.isDirty)
    }

    @Test func standardWithEmptyArgumentDoesNotWriteRow() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.credibility.selectDraft(key: "standard")
        await model.credibility.save()
        #expect(model.workspace?.credibility == nil)
        #expect(store.credibilityBySource[sourceID] == nil)
        #expect(!model.credibility.hasSavedAssessment)
    }

    @Test func cancelCredibilityRestoresDraftFromSaved() async {
        let store = makeStore(credibility: CatalogCredibilityAssessment(
            id: "cred-1",
            sourceID: sourceID,
            gradeID: "g-high",
            gradeKey: "high_trust",
            gradeLabel: "High trust",
            argument: "Film"
        ))
        let model = makeModel(store: store)
        await model.warmFromSession()
        #expect(model.credibility.hasSavedAssessment)
        model.credibility.selectDraft(key: "low_trust")
        model.credibility.argumentDraft = "changed"
        #expect(model.credibility.isDirty)
        model.credibility.cancel()
        #expect(model.credibility.draftKey == "high_trust")
        #expect(model.credibility.argumentDraft == "Film")
        #expect(!model.credibility.isDirty)
    }

    @Test func notesCRUD() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.notes.draft = "First look"
        await model.notes.add()
        #expect(model.notes.items.count == 1)
        #expect(model.notes.draft.isEmpty)

        let noteID = try #require(model.notes.items.first?.id)
        #expect(model.notes.items.first?.authorDisplayName == "Jake Robins")
        #expect(!(model.notes.items.first?.createdAt.isEmpty ?? true))

        model.notes.beginEdit(id: noteID)
        #expect(model.notes.editingNoteID == noteID)
        #expect(model.notes.bodyDraft == "First look")
        model.notes.bodyDraft = "Updated look"
        await model.notes.saveEdit()
        #expect(model.notes.editingNoteID == nil)
        #expect(model.notes.items.first?.body == "Updated look")
        #expect(model.notes.items.first?.authorDisplayName == "Jake Robins")

        await model.notes.delete(id: noteID)
        #expect(model.notes.items.isEmpty)
    }

    @Test func cancelNoteEditRestoresCommittedBody() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.notes.draft = "Original"
        await model.notes.add()
        let noteID = try #require(model.notes.items.first?.id)

        model.notes.beginEdit(id: noteID)
        model.notes.bodyDraft = "Changed"
        model.notes.cancelEdit()
        #expect(model.notes.editingNoteID == nil)
        #expect(model.notes.items.first?.body == "Original")
    }

    @Test func emptyNoteBodyDraftSetsValidationError() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.notes.draft = "Keep me"
        await model.notes.add()
        let noteID = try #require(model.notes.items.first?.id)

        model.notes.beginEdit(id: noteID)
        model.notes.bodyDraft = "   "
        await model.notes.saveEdit()
        #expect(model.notes.bodyError != nil)
        #expect(model.notes.editingNoteID == noteID)
        #expect(model.notes.items.first?.body == "Keep me")
    }

    @Test func createFilelessArtifact() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Physical copy"
        model.artifacts.draft.description = "At the archive"
        await model.artifacts.create()
        #expect(model.artifacts.items.count == 1)
        #expect(model.artifacts.items.first?.label == "Physical copy")
        #expect(model.artifacts.items.first?.fileID.isEmpty == true)
        #expect(model.artifacts.isAdding == false)
    }

    @Test func createArtifactRequiresLabel() async {
        let model = makeModel(store: makeStore())
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "  "
        await model.artifacts.create()
        #expect(model.artifacts.draftLabelError != nil)
        #expect(model.artifacts.items.isEmpty)
    }

    @Test func createArtifactWithPDFLeavesTypeIconCover() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Deed"
        let path = FileManager.default.temporaryDirectory
            .appendingPathComponent("deed-\(UUID().uuidString).pdf").path
        try Data("%PDF-1.1".utf8).write(to: URL(fileURLWithPath: path))
        model.artifacts.draft.filePath = path
        await model.artifacts.create()
        #expect(model.artifacts.items.count == 1)
        #expect(model.source?.coverMode == "type_icon")
        #expect(model.source?.primaryArtifactID.isEmpty == true)
        #expect(model.source?.thumbnailMediaType.isEmpty == true)
        #expect(!model.artifacts.canUseAsThumbnail(model.artifacts.items[0]))
    }

    @Test func creatingAnArtifactWithAFileOpensTheGraphJump() async throws {
        let store = makeStore()
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: "g-1", ref: "GPH-1", sourceID: sourceID, name: "Default", description: "", artifactIDs: []
            ),
        ]
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourcePageModel(
            sourceID: sourceID, session: session, userID: userID,
            sessionDisplayName: "Jake Robins", store: store
        )
        await model.warmFromSession()
        await waitForQuery(session.query(session.evidenceGraphsKey) as QueryHandle<[CatalogEvidenceGraph]>)
        #expect(entryGraphLocation(model) == nil)

        model.artifacts.openAdd()
        model.artifacts.draft.label = "Scan"
        let path = FileManager.default.temporaryDirectory
            .appendingPathComponent("scan-\(UUID().uuidString).pdf").path
        try Data("%PDF-1.1".utf8).write(to: URL(fileURLWithPath: path))
        model.artifacts.draft.filePath = path
        await model.artifacts.create()

        // The create linked the Artifact to the entry graph; the file ingest
        // after it moved no graph link, so the list reloads only if the
        // create's effects were applied too.
        await waitForQuery(session.query(session.evidenceGraphsKey) as QueryHandle<[CatalogEvidenceGraph]>)
        #expect(entryGraphLocation(model)?.graphId == "g-1")
    }

    @Test func useAsThumbnailAndRevertToTypeIcon() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Scan"
        let path = FileManager.default.temporaryDirectory
            .appendingPathComponent("scan-\(UUID().uuidString).png").path
        // Minimal valid PNG so FakeStore marks a raster thumb path.
        try Data([
            0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
        ]).write(to: URL(fileURLWithPath: path))
        model.artifacts.draft.filePath = path
        await model.artifacts.create()
        let first = try #require(model.artifacts.items.first)
        #expect(!model.artifacts.isCover(first))
        #expect(model.artifacts.canUseAsThumbnail(first))
        #expect(model.source?.coverMode == "type_icon")

        await model.artifacts.useAsThumbnail(first)
        #expect(model.source?.coverMode == "artifact")
        #expect(model.source?.primaryArtifactID == first.id)
        #expect(!model.source!.thumbnailRelPath.isEmpty)

        model.artifacts.openAdd()
        model.artifacts.draft.label = "Other"
        let path2 = FileManager.default.temporaryDirectory
            .appendingPathComponent("other-\(UUID().uuidString).pdf").path
        try Data("%PDF".utf8).write(to: URL(fileURLWithPath: path2))
        model.artifacts.draft.filePath = path2
        await model.artifacts.create()
        let second = try #require(model.artifacts.items.last)
        #expect(!model.artifacts.isCover(second))
        #expect(!model.artifacts.canUseAsThumbnail(second))

        await model.artifacts.revertCoverToTypeIcon()
        #expect(model.source?.coverMode == "type_icon")
        #expect(model.source?.primaryArtifactID.isEmpty == true)
        #expect(model.source?.thumbnailRelPath.isEmpty == true)
    }

    @Test func filelessArtifactCannotBeCover() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Note only"
        await model.artifacts.create()
        let art = try #require(model.artifacts.items.first)
        #expect(!model.artifacts.canUseAsThumbnail(art))
        #expect(!model.artifacts.isCover(art))
        #expect(model.source?.coverMode == "type_icon")
    }

    @Test func createFilelessArtifactClearsMIMECoverForTypeIcon() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Physical register"
        await model.artifacts.create()
        #expect(model.source?.thumbnailRelPath.isEmpty == true)
        #expect(model.source?.thumbnailMediaType.isEmpty == true)
        #expect(model.source?.coverMode == "type_icon")
        #expect(model.identity.typeLabel == "Photograph")
    }

    @Test func aRefusedFileKeepsTheCreatedArtifactAndARetryAttachesToIt() async throws {
        let store = makeStore()
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: "g-1", ref: "GPH-1", sourceID: sourceID, name: "Default", description: "", artifactIDs: []
            ),
        ]
        store.ingestArtifactFileError = CoreInvokeError.coded(
            status: 1, code: "ingest.invalid", kind: .user, params: []
        )
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourcePageModel(
            sourceID: sourceID, session: session, userID: userID,
            sessionDisplayName: "Jake Robins", store: store
        )
        await model.warmFromSession()
        await waitForQuery(session.query(session.evidenceGraphsKey) as QueryHandle<[CatalogEvidenceGraph]>)
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Scan"
        model.artifacts.draft.filePath = "/tmp/provenencia-missing-\(UUID().uuidString).bin"

        await model.artifacts.create()

        // Go created the Artifact before the file was refused: the page shows
        // it and its graph link, and the dialog stays open on the refusal.
        #expect(model.artifacts.isAdding)
        #expect(model.artifacts.draft.reject != nil)
        #expect(model.artifacts.items.map(\.label) == ["Scan"])
        #expect(model.artifacts.items.first?.fileID.isEmpty == true)
        await waitForQuery(session.query(session.evidenceGraphsKey) as QueryHandle<[CatalogEvidenceGraph]>)
        #expect(entryGraphLocation(model)?.graphId == "g-1")

        store.ingestArtifactFileError = nil
        let path = FileManager.default.temporaryDirectory
            .appendingPathComponent("scan-\(UUID().uuidString).pdf").path
        try Data("%PDF-1.1".utf8).write(to: URL(fileURLWithPath: path))
        model.artifacts.applyPickedFile(url: URL(fileURLWithPath: path), into: .create)
        model.artifacts.draft.label = "Scan, front"

        await model.artifacts.create()

        #expect(!model.artifacts.isAdding)
        #expect(store.artifactsBySource[sourceID]?.count == 1)
        #expect(model.artifacts.items.count == 1)
        #expect(model.artifacts.items.first?.label == "Scan, front")
        #expect(model.artifacts.items.first?.fileID.isEmpty == false)
    }

    @Test func ingestMissingFileSurfacesIngestInvalid() async {
        let store = makeStore()
        store.ingestArtifactFileError = CoreInvokeError.coded(
            status: 1,
            code: "ingest.invalid",
            kind: .user,
            params: []
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Scan"
        model.artifacts.draft.filePath = "/tmp/provenencia-missing-\(UUID().uuidString).bin"

        await model.artifacts.create()

        #expect(model.artifacts.draft.reject != nil)
        #expect(model.artifacts.draft.filePath == nil)
        #expect(model.artifacts.draftLabelError == nil)
        #expect(model.artifacts.isAdding)
    }

    @Test func applyPickedFileRejectsSpreadsheetWithoutFFI() throws {
        let model = makeModel(store: makeStore())
        model.artifacts.openAdd()
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("ingest-ui-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let url = dir.appendingPathComponent("notes.xlsx")
        try Data("PK\u{0003}\u{0004}".utf8).write(to: url)

        model.artifacts.applyPickedFile(url: url, into: .create)

        #expect(model.artifacts.draft.filePath == nil)
        #expect(model.artifacts.draft.fileName == "notes.xlsx")
        #expect(model.artifacts.draft.reject != nil)
        #expect(!model.artifacts.canSubmitDraft) // label still empty too
        model.artifacts.draft.label = "Letter"
        #expect(!model.artifacts.canSubmitDraft)
    }

    @Test func ingestThenRejectSecondAttach() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Scan"
        await model.artifacts.create()
        let artID = try #require(model.artifacts.items.first?.id)

        let path = FileManager.default.temporaryDirectory
            .appendingPathComponent("scan-\(UUID().uuidString).bin").path
        try Data("bytes".utf8).write(to: URL(fileURLWithPath: path))

        let first = try await store.ingestArtifactFile(
            projectDir: projectDir, userID: userID, artifactID: artID, path: path
        )
        #expect(!first.artifact.fileID.isEmpty)

        // Refresh model artifact state from store.
        await model.warmFromSession()
        #expect(model.artifacts.items.first?.fileID.isEmpty == false)

        do {
            _ = try await store.ingestArtifactFile(
                projectDir: projectDir, userID: userID, artifactID: artID, path: path
            )
            Issue.record("expected second ingest to fail")
        } catch let error as CoreInvokeError {
            guard case .coded(_, let code, _, _) = error else {
                Issue.record("unexpected CoreInvokeError")
                return
            }
            #expect(code == "artifacts.file_already_attached")
        }
    }

    @Test func workspaceArtifactsCarryThumbnailRelPath() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "f1",
                    label: "Scan", description: "",
                    file: CatalogFileRef(
                        id: "f1", relPath: "objects/aa/bb/prim",
                        originalFilename: "scan.png", mediaType: "image/png", byteSize: 12
                    ),
                    thumbnailRelPath: "objects/aa/bb/thumb"
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        #expect(model.artifacts.items.first?.thumbnailRelPath == "objects/aa/bb/thumb")
    }

    private func authorField() -> CatalogMetadataField {
        CatalogMetadataField(
            id: "f-author", key: "author", origin: "provenencia",
            label: "Author", dataType: "text", description: ""
        )
    }

    private func repositoryField() -> CatalogMetadataField {
        CatalogMetadataField(
            id: "f-repo", key: "repository", origin: "provenencia",
            label: "Repository", dataType: "text", description: ""
        )
    }

    @Test func loadMetadataSuggestionsAndValues() async {
        let author = authorField()
        let repo = repositoryField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: author, valueText: "",
                    hasValue: false, suggested: true, sortOrder: 0
                ),
                CatalogMetadataEntry(
                    field: repo, valueText: "NRO",
                    hasValue: true, suggested: true, sortOrder: 1
                ),
            ],
            fields: [author, repo]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        #expect(model.metadata.entries.count == 2)
        #expect(model.metadata.saved.map(\.field.key) == ["repository"])
        #expect(model.metadata.suggested.map(\.field.key) == ["author"])
    }

    @Test func saveMetadataValueFillsSuggestion() async {
        let author = authorField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: author, valueText: "",
                    hasValue: false, suggested: true, sortOrder: 0
                ),
            ],
            fields: [author]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        #expect(model.metadata.suggested.count == 1)
        model.metadata.drafts[author.id] = "Mary Robins"
        await model.metadata.save(fieldID: author.id)
        #expect(model.metadata.saved.map(\.valueText) == ["Mary Robins"])
        #expect(model.metadata.suggested.isEmpty)
    }

    @Test func beginEditMetadataThenCancelRestoresDraft() async {
        let repo = repositoryField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: repo, valueText: "NRO",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
            ],
            fields: [repo]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.beginEdit(fieldID: repo.id)
        #expect(model.metadata.editingFieldID == repo.id)
        model.metadata.drafts[repo.id] = "Changed"
        model.metadata.cancelEdit()
        #expect(model.metadata.editingFieldID == nil)
        #expect(model.metadata.drafts[repo.id] == "NRO")
    }

    @Test func saveMetadataValueExitsEditMode() async {
        let repo = repositoryField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: repo, valueText: "NRO",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
            ],
            fields: [repo]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.beginEdit(fieldID: repo.id)
        model.metadata.drafts[repo.id] = "Norfolk Record Office"
        await model.metadata.save(fieldID: repo.id)
        #expect(model.metadata.editingFieldID == nil)
        #expect(model.metadata.saved.map(\.valueText) == ["Norfolk Record Office"])
    }

    @Test func dismissSuggestionRemovesEmptyRow() async {
        let author = authorField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: author, valueText: "",
                    hasValue: false, suggested: true, sortOrder: 0
                ),
            ],
            fields: [author]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.metadata.dismissSuggestion(fieldID: author.id)
        #expect(model.metadata.entries.isEmpty)
        #expect(model.metadata.suggested.isEmpty)
    }

    @Test func reorderSavedMetadataLeavesSuggestions() async {
        let author = authorField()
        let repo = repositoryField()
        let issue = CatalogMetadataField(
            id: "f-issue", key: "issue", origin: "provenencia",
            label: "Issue", dataType: "text", description: ""
        )
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: author, valueText: "A",
                    hasValue: true, suggested: true, sortOrder: 0
                ),
                CatalogMetadataEntry(
                    field: issue, valueText: "",
                    hasValue: false, suggested: true, sortOrder: 1
                ),
                CatalogMetadataEntry(
                    field: repo, valueText: "B",
                    hasValue: true, suggested: true, sortOrder: 2
                ),
            ],
            fields: [author, repo, issue]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.metadata.moveSaved(from: IndexSet(integer: 0), to: 2)
        #expect(model.metadata.saved.map(\.field.id) == [repo.id, author.id])
        #expect(model.metadata.suggested.map(\.field.id) == [issue.id])
        #expect(model.metadata.entries.map(\.field.id) == [repo.id, author.id, issue.id])
    }

    @Test func reorderSavedMetadataFailureRevertsAndSetsPageError() async {
        let author = authorField()
        let repo = repositoryField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: author, valueText: "A",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
                CatalogMetadataEntry(
                    field: repo, valueText: "B",
                    hasValue: true, suggested: false, sortOrder: 1
                ),
            ],
            fields: [author, repo]
        )
        store.reorderSourceMetadataError = CoreInvokeError.coded(
            status: 1, code: "internal.unknown", kind: .internal, params: []
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        let before = model.metadata.saved.map(\.field.id)
        await model.metadata.moveSaved(from: IndexSet(integer: 0), to: 2)
        #expect(model.pageError != nil)
        #expect(model.metadata.saved.map(\.field.id) == before)
        #expect(store.metadataBySource[sourceID]?.map(\.field.id) == before)
    }

    @Test func addMetadataFromDialog() async {
        let author = authorField()
        let store = makeStore(fields: [author])
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.openAdd()
        model.metadata.addFieldID = author.id
        model.metadata.addValue = "Eliza"
        await model.metadata.createFromAdd()
        #expect(model.metadata.isAdding == false)
        #expect(model.metadata.entries.contains { $0.field.id == author.id && $0.valueText == "Eliza" })
    }

    /// A field added on the Metadata page reaches the Add-metadata dropdown
    /// without the page reloading: the dropdown reads the shared field list.
    @Test func addMetadataDropdownSeesFieldCreatedElsewhere() async throws {
        let store = makeStore(fields: [authorField()])
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourcePageModel(
            sourceID: sourceID, session: session, userID: userID, store: store
        )
        await model.warmFromSession()
        #expect(model.metadata.fieldComboOptions.map(\.label) == ["Author"])

        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: sourceID)
        let workspaceBefore: QueryHandle<CatalogSourceWorkspace>? = session.queryHandle(workspaceKey)
        let payloadBefore = workspaceBefore?.value

        store.fieldsByProject[projectDir]?.append(
            CatalogMetadataField(
                id: "f-folio", key: "folio", origin: "user", label: "Folio",
                dataType: "text", description: ""
            )
        )
        session.apply(.createdMetadataField)
        session.apply(location: WorkspaceLocation(section: .sources, sourceId: sourceID))

        let fieldsHandle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(
            .metadataFieldsList(project: session.projectKey)
        )
        await waitForQuery(try #require(fieldsHandle))

        #expect(model.metadata.fieldComboOptions.map(\.label) == ["Author", "Folio"])
        // The page payload itself was never refetched.
        #expect(workspaceBefore?.value == payloadBefore)
    }

    @Test func addMetadataRequiresFieldAndValue() async {
        let author = authorField()
        let store = makeStore(fields: [author])
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.openAdd()
        await model.metadata.createFromAdd()
        #expect(model.metadata.addFieldError != nil)
        #expect(model.metadata.isAdding == true)

        model.metadata.addFieldID = author.id
        model.metadata.addValue = "   "
        await model.metadata.createFromAdd()
        #expect(model.metadata.addValueError != nil)
        #expect(model.metadata.isAdding == true)
        #expect(model.metadata.entries.isEmpty)
    }

    @Test func cancelArtifactFieldsRestoresDrafts() async throws {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Front", description: "Original", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        let artID = try #require(model.artifacts.items.first?.id)
        model.artifacts.labels[artID] = "Changed"
        model.artifacts.descriptions[artID] = "Changed desc"
        #expect(model.artifacts.fieldsDirty(artID))
        model.artifacts.cancelFields(id: artID)
        #expect(!model.artifacts.fieldsDirty(artID))
        #expect(model.artifacts.labels[artID] == "Front")
        #expect(model.artifacts.descriptions[artID] == "Original")
    }

    private func recordDateField() -> CatalogMetadataField {
        CatalogMetadataField(
            id: "f-date", key: "record_date", origin: "provenencia",
            label: "Date of record", dataType: "text", description: ""
        )
    }

    private func urlField() -> CatalogMetadataField {
        CatalogMetadataField(
            id: "f-url", key: "url", origin: "provenencia",
            label: "URL", dataType: "url", description: ""
        )
    }

    @Test func saveDateNamedFieldAsText() async {
        let dateField = recordDateField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: dateField, valueText: "about the year 1890",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
            ],
            fields: [dateField]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.beginEdit(fieldID: dateField.id)
        model.metadata.drafts[dateField.id] = "abt 1890"
        await model.metadata.save(fieldID: dateField.id)
        #expect(model.metadata.editingFieldID == nil)
        #expect(model.metadata.saved.map(\.valueText) == ["abt 1890"])
        #expect(model.pageError == nil)
    }

    @Test func clearSuggestedMetadataReturnsToDashedList() async {
        let author = authorField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: author, valueText: "Mary Robins",
                    hasValue: true, suggested: true, sortOrder: 0
                ),
            ],
            fields: [author]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.askClear(fieldID: author.id)
        await model.metadata.confirmClear()
        #expect(model.metadata.pendingDelete == nil)
        #expect(model.metadata.saved.isEmpty)
        #expect(model.metadata.suggested.map(\.field.id) == [author.id])
        #expect(store.metadataBySource[sourceID]?.first?.hasValue == false)
    }

    @Test func clearExtraMetadataRemovesRow() async {
        let repo = repositoryField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: repo, valueText: "NRO",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
            ],
            fields: [repo]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.askClear(fieldID: repo.id)
        await model.metadata.confirmClear()
        #expect(model.metadata.pendingDelete == nil)
        #expect(model.metadata.entries.isEmpty)
        #expect(store.metadataBySource[sourceID]?.isEmpty == true)
    }

    @Test func saveSchemeLessURLValue() async {
        let field = urlField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: field, valueText: "",
                    hasValue: false, suggested: true, sortOrder: 0
                ),
            ],
            fields: [field]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.drafts[field.id] = "www.url.com"
        await model.metadata.save(fieldID: field.id)
        #expect(model.metadata.saved.map(\.valueText) == ["www.url.com"])
        #expect(model.metadata.suggested.isEmpty)
        #expect(model.pageError == nil)
        #expect(model.metadata.fieldError == nil)
    }

    @Test func invalidMetadataSetLandsOnFieldError() async {
        let field = urlField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: field, valueText: "https://example.com",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
            ],
            fields: [field]
        )
        store.setSourceMetadataError = CoreInvokeError.coded(
            status: 1, code: "sourcemetadata.invalid", kind: .user, params: []
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.beginEdit(fieldID: field.id)
        model.metadata.drafts[field.id] = "file:/tmp"
        await model.metadata.save(fieldID: field.id)
        #expect(model.metadata.fieldError == L10n.Errors.message(for: store.setSourceMetadataError!))
        #expect(model.metadata.fieldErrorID == field.id)
        #expect(model.metadata.editingFieldID == field.id)
        #expect(model.pageError == nil)
    }

    @Test func invalidMetadataAddLandsOnValueField() async {
        let field = urlField()
        let store = makeStore(fields: [field])
        store.setSourceMetadataError = CoreInvokeError.coded(
            status: 1, code: "sourcemetadata.invalid", kind: .user, params: []
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.openAdd()
        model.metadata.addFieldID = field.id
        model.metadata.addValue = "javascript:alert(1)"
        await model.metadata.createFromAdd()
        #expect(model.metadata.addValueError == L10n.Errors.message(for: store.setSourceMetadataError!))
        #expect(model.metadata.isAdding == true)
        #expect(model.pageError == nil)
        #expect(model.metadata.entries.isEmpty)
    }

    @Test func metadataSetIOFailureLandsOnPageError() async {
        let field = urlField()
        let store = makeStore(
            metadata: [
                CatalogMetadataEntry(
                    field: field, valueText: "https://example.com",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
            ],
            fields: [field]
        )
        store.setSourceMetadataError = CoreInvokeError.failed(status: 2)
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.metadata.beginEdit(fieldID: field.id)
        model.metadata.drafts[field.id] = "https://example.org"
        await model.metadata.save(fieldID: field.id)
        #expect(model.pageError == L10n.Errors.message(for: store.setSourceMetadataError!))
        #expect(model.metadata.fieldError == nil)
        #expect(model.metadata.editingFieldID == field.id)
    }

    @Test func editingArtifactIsOneCardAtATime() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Front", description: "", file: nil
                ),
                CatalogArtifact(
                    id: "a2", ref: "ART-BBBBB", sourceID: sourceID, fileID: "",
                    label: "Back", description: "", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.artifacts.beginEdit("a1")
        #expect(model.artifacts.editingArtifactID == "a1")
        model.artifacts.labels["a1"] = "Changed"
        #expect(model.artifacts.canSaveFields("a1"))
        model.artifacts.labels["a1"] = "   "
        #expect(!model.artifacts.canSaveFields("a1"))
        #expect(
            model.artifacts.cardLabelError(id: "a1")
                == L10n.string(L10n.Sources.artifactCardLabelRequired)
        )
        model.artifacts.beginEdit("a2")
        #expect(model.artifacts.editingArtifactID == "a2")
        #expect(model.artifacts.labels["a1"] == "Front")
        model.artifacts.cancelFields(id: "a2")
        #expect(model.artifacts.editingArtifactID == nil)
        #expect(!model.artifacts.canSaveFields("a2"))
    }

    @Test func pageDropStagesFileWithoutCreating() async throws {
        let url = try writeTempPNG()
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Front", description: "", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        let before = model.artifacts.items.count
        model.artifacts.beginEdit("a1")
        model.artifacts.labels["a1"] = "Changed"
        model.artifacts.stagePageDrop([url])
        #expect(model.artifacts.isAdding)
        #expect(model.artifacts.editingArtifactID == nil)
        #expect(model.artifacts.labels["a1"] == "Front")
        #expect(model.artifacts.draft.label.isEmpty)
        #expect(model.artifacts.draft.description.isEmpty)
        #expect(model.artifacts.draft.fileName == url.lastPathComponent)
        #expect(model.artifacts.draft.filePath == url.path)
        #expect(model.artifacts.draft.reject == nil)
        #expect(model.artifacts.items.count == before)
        #expect(store.artifactsBySource[sourceID]?.count == before)
    }

    @Test func pageDropOfTwoFilesOpensRejectCallout() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Front", description: "", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        let before = model.artifacts.items.count
        model.artifacts.stagePageDrop([
            URL(fileURLWithPath: "/tmp/recto.png"),
            URL(fileURLWithPath: "/tmp/verso.png"),
        ])
        #expect(model.artifacts.isAdding)
        #expect(model.artifacts.draft.label.isEmpty)
        #expect(model.artifacts.draft.filePath == nil)
        #expect(model.artifacts.draft.reject != nil)
        #expect(!model.artifacts.canSubmitDraft)
        #expect(model.artifacts.items.count == before)
        #expect(store.artifactsBySource[sourceID]?.count == before)
    }

    private func writeTempPNG() throws -> URL {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("s10-18-\(UUID().uuidString).png")
        try Data([0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A]).write(to: url)
        return url
    }

    @Test func askDeleteSourcePresentsAllowedImpact() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.askDeleteSource()
        #expect(model.pendingImpact?.report.allowed == true)
        #expect(model.pendingImpact?.target.kind == "source")
        #expect(model.pendingImpact?.target.ref == "SRC-AAAAA")
    }

    @Test func confirmDeleteSourceCallsStore() async {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.askDeleteSource()
        let left = await model.confirmResourceDelete()
        #expect(left)
        #expect(store.deleteSourceCalls == 1)
        #expect(model.pendingImpact == nil)
    }

    @Test func blockedSourceDoesNotCallDelete() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Scan", description: "", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.askDeleteSource()
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.groups.first?.via == "artifacts.source_id")
        let left = await model.confirmResourceDelete()
        #expect(!left)
        #expect(store.deleteSourceCalls == 0)
    }

    /// Where Open goes for the page's first graph; nil while it links no Artifact.
    private func entryGraphLocation(_ model: SourcePageModel) -> WorkspaceLocation? {
        model.graphs.rows?.first.flatMap { model.graphs.location(for: $0) }
    }

    @Test func graphRowsAreTheSourcesGraphsByName() async {
        let store = makeStore()
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g-smith", name: "Smith household"),
            evidenceGraph("g-default", name: "Default"),
            evidenceGraph("g-jones", name: "Jones household"),
            evidenceGraph("g-other", sourceID: "src-2", name: "Other source"),
        ]
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = makeModel(store: store, session: session)
        await model.warmFromSession()
        await waitForQuery(session.evidenceGraphsHandle!)
        #expect(model.graphs.rows?.map(\.name) == ["Default", "Jones household", "Smith household"])
        #expect(model.graphs.rows?.first?.name == "Default")
        #expect(L10n.Sources.graphsCount(3) == "3 evidence graphs")
    }

    @Test func graphRowsNilUntilListLoadsAndErrorWhenItFails() async {
        let store = makeStore()
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g-default", name: "Default")]
        let model = makeModel(store: store)
        #expect(model.graphs.rows == nil)
        #expect(model.graphs.loadError == nil)

        struct Boom: Error {}
        let failing = makeStore()
        failing.listEvidenceGraphsError = Boom()
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: failing)
        let failed = makeModel(store: failing, session: session)
        await failed.warmFromSession()
        await waitForQuery(session.evidenceGraphsHandle!)
        #expect(failed.graphs.rows == nil)
        #expect(failed.graphs.loadError != nil)
    }

    @Test func addGraphNeedsAnArtifact() async {
        let bare = makeModel(store: makeStore())
        await bare.warmFromSession()
        #expect(!bare.graphs.canAdd)
        #expect(bare.graphs.addBlockedReason == "Add an artifact before you add an evidence graph")
        bare.graphs.openAdd()
        #expect(bare.graphs.form == nil)

        let withArtifact = makeModel(store: makeStore(artifacts: [
            CatalogArtifact(id: "a1", ref: "ART-A1", sourceID: sourceID, fileID: "", label: "Page 1", description: ""),
        ]))
        await withArtifact.warmFromSession()
        #expect(withArtifact.graphs.canAdd)
        #expect(withArtifact.graphs.addBlockedReason == nil)
        withArtifact.graphs.openAdd()
        #expect(withArtifact.graphs.form != nil)
    }

    @Test func deleteEmptyGraphCallsStoreAndToasts() async {
        let store = makeStore()
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g-default", name: "Default"),
            evidenceGraph("g-jones", name: "Jones household"),
        ]
        store.deleteImpactByID["g-jones"] = CatalogDeleteImpact(allowed: true, gate: .ok, groups: [])
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = makeModel(store: store, session: session)
        session.apply(location: WorkspaceLocation.sourcePage(sourceId: sourceID))
        await model.warmFromSession()
        await waitForQuery(session.evidenceGraphsHandle!)
        await model.askDeleteGraph(id: "g-jones")
        #expect(model.pendingImpact?.target.kind == "evidence_graph")
        #expect(model.pendingImpact?.target.ref == "GPH-G-JONES")
        let left = await model.confirmResourceDelete()
        #expect(!left)
        #expect(store.deleteEvidenceGraphCalls == 1)
        #expect(store.deleteArtifactCalls == 0)
        #expect(model.toast?.title == "Evidence graph deleted")
        #expect(model.toast?.body == "GPH-G-JONES was erased from this source.")
        // The page is the visible place, so the effects revalidate the list.
        #expect(await waitUntil { model.graphs.rows?.map(\.id) == ["g-default"] })
    }

    @Test(arguments: [
        CatalogDeleteImpact(
            allowed: false, gate: .inbound,
            groups: [CatalogDeleteImpactGroup(via: "subjects.graph_id", kind: "subject", total: 3, listed: [])]
        ),
        CatalogDeleteImpact(allowed: false, gate: .lastGraph, groups: []),
    ])
    func blockedGraphDoesNotCallDelete(report: CatalogDeleteImpact) async {
        let store = makeStore()
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g-default", name: "Default")]
        store.deleteImpactByID["g-default"] = report
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = makeModel(store: store, session: session)
        await model.warmFromSession()
        await waitForQuery(session.evidenceGraphsHandle!)
        await model.askDeleteGraph(id: "g-default")
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.gate == report.gate)
        #expect(model.pendingImpact?.report.groups.first?.total == report.groups.first?.total)
        let left = await model.confirmResourceDelete()
        #expect(!left)
        #expect(store.deleteEvidenceGraphCalls == 0)
    }

    @Test func deleteUncitedArtifactStaysOnPageAndClearsCover() async {
        var source = CatalogSource(
            id: sourceID, ref: "SRC-AAAAA", sourceTypeID: "t1",
            title: "Family album", description: "Held by Mary"
        )
        source.coverMode = "artifact"
        source.primaryArtifactID = "a1"
        let store = makeStore(
            source: source,
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "f1",
                    label: "Cover scan", description: "", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.askDeleteArtifact(id: "a1")
        #expect(model.pendingImpact?.report.allowed == true)
        let left = await model.confirmResourceDelete()
        #expect(!left)
        #expect(store.deleteArtifactCalls == 1)
        #expect(model.artifacts.items.isEmpty)
        #expect(model.source?.primaryArtifactID.isEmpty == true)
        #expect(model.source?.coverMode == "type_icon")
    }

    /// Deleting a Source's only Artifact changes row fields the activity patch
    /// cannot (`hasArtifact`, the cover), so the cached Sources row is refreshed
    /// from the engine as Artifact create and attach do (D-26).
    @Test func deleteOnlyArtifactClearsHasArtifactOnTheCachedRow() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Scan", description: "", file: nil
                ),
            ]
        )
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        var row = CatalogSource(
            id: sourceID, ref: "SRC-AAAAA", sourceTypeID: "t1",
            title: "Family album", description: "Held by Mary"
        )
        row.hasArtifact = true
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        session.setQueryValue(listKey, value: [row])
        let model = SourcePageModel(
            sourceID: sourceID, session: session, userID: userID,
            sessionDisplayName: "Jake Robins", store: store
        )
        await model.warmFromSession()
        await model.askDeleteArtifact(id: "a1")
        let workspaceReads = store.getSourceWorkspaceCalls
        _ = await model.confirmResourceDelete()
        let list: QueryHandle<[CatalogSource]>? = session.queryHandle(listKey)
        #expect(list?.value?.first?.hasArtifact == false)
        // The refresh is the one read of the Source's workspace, as create does.
        #expect(store.getSourceWorkspaceCalls == workspaceReads + 1)
    }

    @Test func blockedArtifactDoesNotCallDelete() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "",
                    label: "Cited", description: "", file: nil
                ),
            ]
        )
        store.citationsByID["c1"] = CatalogCitation(
            id: "c1",
            ref: "CIT-AAAAA",
            artifactID: "a1",
            locatorJSON: "{}",
            transcription: "John Hartley, worsted weaver",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: "g-1", ref: "GPH-1", sourceID: sourceID, name: "Default", description: "", artifactIDs: ["a1"]
            ),
        ]
        let model = makeModel(store: store)
        await model.warmFromSession()
        await model.askDeleteArtifact(id: "a1")
        #expect(model.pendingImpact?.report.allowed == false)
        #expect(model.pendingImpact?.report.groups.first?.via == "citations.artifact_id")
        let jump = model.pendingImpact?.report.groups.first?.listed.first?.location
        #expect(jump?.sourceId == sourceID)
        #expect(jump?.citationId == "c1")
        #expect(jump?.artifactId == "a1")
        #expect(jump?.graphId == "g-1")
        #expect(jump?.sourceSurface == .citationComposer)
        #expect(model.pendingImpact?.report.groups.first?.listed.first?.title == "John Hartley, worsted weaver")
        #expect(model.pendingImpact?.report.groups.first?.listed.first?.ref == "CIT-AAAAA")
        let left = await model.confirmResourceDelete()
        #expect(!left)
        #expect(store.deleteArtifactCalls == 0)
    }

    @Test func deletingSecondaryArtifactKeepsIdentityDraft() async {
        let store = makeStore(
            artifacts: [
                CatalogArtifact(
                    id: "a1", ref: "ART-AAAAA", sourceID: sourceID, fileID: "f1",
                    label: "Cover scan", description: "", file: nil
                ),
                CatalogArtifact(
                    id: "a2", ref: "ART-BBBBB", sourceID: sourceID, fileID: "f2",
                    label: "Extra scan", description: "", file: nil
                ),
            ]
        )
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.identity.beginEditTitle()
        model.identity.titleDraft = "Unsaved title"
        model.notes.draft = "Scratch note"
        await model.askDeleteArtifact(id: "a2")
        let left = await model.confirmResourceDelete()
        #expect(!left)
        #expect(store.deleteArtifactCalls == 1)
        #expect(model.identity.editingTitle)
        #expect(model.identity.titleDraft == "Unsaved title")
        #expect(model.notes.draft == "Scratch note")
        #expect(model.artifacts.items.map(\.id) == ["a1"])
    }

    @Test func noteDeleteWaitsForConfirm() async throws {
        let store = makeStore()
        let model = makeModel(store: store)
        await model.warmFromSession()
        model.notes.draft = "First look"
        await model.notes.add()
        let noteID = try #require(model.notes.items.first?.id)
        model.notes.askDelete(id: noteID)
        #expect(model.notes.pendingDelete?.id == noteID)
        #expect(model.notes.items.count == 1)
        await model.notes.confirmDelete()
        #expect(model.notes.items.isEmpty)
        #expect(model.notes.pendingDelete == nil)
    }

    @Test func confirmDeleteSourceRecountsThroughSession() async {
        let store = makeStore()
        store.sourcesByProject[projectDir]?.append(
            CatalogSource(
                id: "src-2", ref: "SRC-BBBBB", sourceTypeID: "t1",
                title: "Letters", description: ""
            )
        )
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        session.onNavCountsMoved = { await counts.refreshAll() }
        let model = SourcePageModel(
            sourceID: sourceID,
            session: session,
            userID: userID,
            sessionDisplayName: "Jake Robins",
            store: store
        )
        await model.warmFromSession()
        await model.askDeleteSource()
        let left = await model.confirmResourceDelete()
        #expect(left)
        #expect(await waitUntil { counts.sources == 1 })
    }

    // MARK: Artifact ↔ graph links (S12-13)

    private func graph(_ id: String, _ name: String, _ artifactIDs: [String]) -> CatalogEvidenceGraph {
        evidenceGraph(id, sourceID: sourceID, name: name, artifactIDs: artifactIDs)
    }

    private func page(_ id: String) -> CatalogArtifact {
        CatalogArtifact(
            id: id, ref: "ART-\(id)", sourceID: sourceID, fileID: "", label: "Page \(id)", description: ""
        )
    }

    /// The census: Brown (`g-b`, P9), Jones (`g-j`, P3 P4), Smith (`g-s`, P3 P6), with the
    /// page location applied so the graph list is a visible key and graph writes refresh it.
    private func makeCensus(
        graphs: [CatalogEvidenceGraph]? = nil
    ) async -> (model: SourcePageModel, store: FakeStore, session: WorkspaceSession) {
        let artifacts = ["P3", "P4", "P6", "P9"].map(page)
        let store = makeStore(artifacts: artifacts)
        store.evidenceGraphsByProject[projectDir] = graphs ?? [
            graph("g-b", "Brown household", ["P9"]),
            graph("g-j", "Jones household", ["P3", "P4"]),
            graph("g-s", "Smith household", ["P3", "P6"]),
        ]
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourcePageModel(
            sourceID: sourceID, session: session, userID: userID,
            sessionDisplayName: "Jake Robins", store: store
        )
        session.apply(location: WorkspaceLocation.sourcePage(sourceId: sourceID))
        await model.warmFromSession()
        await waitForQuery(session.evidenceGraphsHandle!)
        return (model, store, session)
    }

    private func smithBlock(citations: Int = 2) -> CatalogEvidenceGraphUnlinkBlock {
        CatalogEvidenceGraphUnlinkBlock(
            artifactID: "P3",
            citations: CatalogDeleteImpactGroup(
                via: "evidencegraphs.artifact_in_use", kind: "citation", total: citations, listed: []
            )
        )
    }

    private func linkCalls(_ store: FakeStore) -> [String] {
        store.recordedCalls.filter { $0.contains("EvidenceGraphArtifact") }
    }

    @Test func addArtifactOnOneGraphSendsEveryGraph() async {
        let (model, store, _) = await makeCensus(graphs: [graph("g-1", "Default", [])])
        #expect(model.artifacts.visibleGraphs == nil)
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Page 7"
        await model.artifacts.create()
        // `[]` links every graph: the one graph linked it.
        #expect(store.evidenceGraphsByProject[projectDir]?.first?.artifactIDs.count == 1)
        #expect(!model.artifacts.isAdding)

        // A graph list with no value (loading) shows no group and sends `[]`.
        let loading = makeStore()
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: loading)
        let other = makeModel(store: loading, session: session)
        #expect(other.artifacts.visibleGraphs == nil)
        #expect(other.artifacts.draftGraphIDs.isEmpty)
    }

    @Test func addArtifactLinksOnlyTheCheckedGraphs() async {
        let (model, store, _) = await makeCensus()
        #expect(model.artifacts.visibleGraphs?.map(\.id) == ["g-b", "g-j", "g-s"])
        model.artifacts.openAdd()
        #expect(model.artifacts.draftGraphIDs == ["g-b", "g-j", "g-s"])
        model.artifacts.draft.label = "Page 7"
        model.artifacts.draft.excludedGraphIDs = ["g-b", "g-s"]
        #expect(model.artifacts.draftGraphIDs == ["g-j"])
        await model.artifacts.create()
        let created = store.artifactsBySource[sourceID]?.last
        #expect(created?.label == "Page 7")
        let graphs = store.evidenceGraphsByProject[projectDir] ?? []
        #expect(graphs.first { $0.id == "g-j" }?.artifactIDs.contains(created?.id ?? "") == true)
        #expect(graphs.first { $0.id == "g-b" }?.artifactIDs.contains(created?.id ?? "") == false)
        #expect(graphs.first { $0.id == "g-s" }?.artifactIDs.contains(created?.id ?? "") == false)
        #expect(!model.artifacts.isAdding)
        #expect(model.toast != nil)
    }

    @Test func addArtifactNeedsAGraph() async {
        let (model, store, _) = await makeCensus()
        model.artifacts.openAdd()
        model.artifacts.draft.label = "Page 7"
        model.artifacts.draft.excludedGraphIDs = ["g-b", "g-j", "g-s"]
        #expect(!model.artifacts.canSubmitDraft)
        #expect(model.artifacts.draftNeedsGraph)
        await model.artifacts.create()
        #expect(store.artifactsBySource[sourceID]?.count == 4)
        #expect(model.artifacts.isAdding)
    }

    @Test func addArtifactExcludedGraphIsDirty() async {
        let (model, _, _) = await makeCensus()
        model.artifacts.openAdd()
        #expect(!model.artifacts.addSheetIsDirty)
        model.artifacts.draft.excludedGraphIDs = ["g-b"]
        #expect(model.artifacts.addSheetIsDirty)
    }

    @Test func cardEditOnOneGraphHasNoLinks() async {
        let (model, store, _) = await makeCensus(graphs: [graph("g-1", "Default", ["P3"])])
        model.artifacts.beginEdit("P3")
        await model.artifacts.linkLoadTask?.value
        #expect(model.artifacts.linkEdit == nil)
        #expect(store.getEvidenceGraphUnlinkImpactCalls == 0)
    }

    @Test func cardEditLocksInUseAndOnlyArtifactGraphs() async {
        let (model, store, _) = await makeCensus()
        store.unlinkImpactBlocksByGraph["g-s"] = [smithBlock()]
        model.artifacts.beginEdit("P3")
        #expect(model.artifacts.linkEdit?.loadState == .loading || model.artifacts.linkEdit?.loadState == .ready)
        await model.artifacts.linkLoadTask?.value
        let edit = model.artifacts.linkEdit
        #expect(edit?.loadState == .ready)
        #expect(edit?.locked["g-s"] == "2 citations in this graph use this artifact")
        #expect(edit?.locked["g-j"] == nil)
        #expect(edit?.excluded == ["g-b"])
        #expect(store.getEvidenceGraphUnlinkImpactGraphIDs == ["g-j", "g-s"])

        store.getEvidenceGraphUnlinkImpactGraphIDs = []
        model.artifacts.beginEdit("P9")
        await model.artifacts.linkLoadTask?.value
        #expect(model.artifacts.linkEdit?.locked["g-b"] == L10n.string(L10n.Sources.artifactGraphOnlyArtifact))
        #expect(model.artifacts.linkEdit?.excluded == ["g-j", "g-s"])
        #expect(store.getEvidenceGraphUnlinkImpactGraphIDs.isEmpty)
    }

    @Test func cardSaveLinksThenUnlinks() async {
        let (model, store, session) = await makeCensus()
        model.artifacts.beginEdit("P3")
        #expect(!model.artifacts.canSaveFields("P3"))
        await model.artifacts.linkLoadTask?.value
        model.artifacts.linkEdit?.excluded = ["g-j"]
        #expect(model.artifacts.checkedGraphIDs == ["g-b", "g-s"])
        #expect(model.artifacts.canSaveFields("P3"))
        await model.artifacts.saveFields(id: "P3")
        #expect(linkCalls(store) == [
            "linkEvidenceGraphArtifact graph=g-b artifact=P3",
            "unlinkEvidenceGraphArtifact graph=g-j artifact=P3",
        ])
        #expect(store.updateArtifactCalls == 0)
        #expect(model.artifacts.editingArtifactID == nil)
        #expect(model.artifacts.linkEdit == nil)
        #expect(session.evidenceGraph(id: "g-b")?.artifactIDs.contains("P3") == true)
        #expect(session.evidenceGraph(id: "g-j")?.artifactIDs.contains("P3") == false)
    }

    @Test func cardSaveFieldsOnlyHasNoLinkCalls() async {
        let (model, store, _) = await makeCensus()
        model.artifacts.beginEdit("P3")
        await model.artifacts.linkLoadTask?.value
        model.artifacts.labels["P3"] = "Page 3, front"
        await model.artifacts.saveFields(id: "P3")
        #expect(store.updateArtifactCalls == 1)
        #expect(linkCalls(store).isEmpty)
        #expect(model.artifacts.items.first { $0.id == "P3" }?.label == "Page 3, front")
        #expect(model.artifacts.editingArtifactID == nil)
    }

    @Test func cardSaveRefusedKeepsEditAndRereadsLocks() async {
        let (model, store, _) = await makeCensus()
        model.artifacts.beginEdit("P3")
        await model.artifacts.linkLoadTask?.value
        model.artifacts.linkEdit?.excluded = ["g-j"]
        store.unlinkEvidenceGraphArtifactError = CoreInvokeError.coded(
            status: 1, code: "evidencegraphs.artifact_in_use", kind: .conflict, params: []
        )
        store.unlinkImpactBlocksByGraph["g-j"] = [smithBlock(citations: 1)]
        let impactCalls = store.getEvidenceGraphUnlinkImpactCalls
        await model.artifacts.saveFields(id: "P3")

        #expect(model.artifacts.editingArtifactID == "P3")
        #expect(model.artifacts.cardSaveError != nil)
        #expect(model.pageError == nil)
        #expect(model.artifacts.linkEdit?.excluded == ["g-j"])
        #expect(linkCalls(store) == ["linkEvidenceGraphArtifact graph=g-b artifact=P3"])
        // The save awaited the reloaded list, so the new link is current and Jones locks.
        #expect(model.artifacts.linkedGraphIDs.contains("g-b"))
        #expect(store.getEvidenceGraphUnlinkImpactCalls > impactCalls)
        #expect(model.artifacts.linkEdit?.locked["g-j"] == "1 citation in this graph uses this artifact")

        store.unlinkEvidenceGraphArtifactError = nil
        store.unlinkImpactBlocksByGraph["g-j"] = []
        await model.artifacts.loadLinkLocks()
        await model.artifacts.saveFields(id: "P3")
        #expect(linkCalls(store) == [
            "linkEvidenceGraphArtifact graph=g-b artifact=P3",
            "unlinkEvidenceGraphArtifact graph=g-j artifact=P3",
        ])
        #expect(model.artifacts.editingArtifactID == nil)
        #expect(model.artifacts.cardSaveError == nil)
    }

    @Test func cardLinkReadFailsAndReadAgain() async {
        let (model, store, _) = await makeCensus()
        store.getEvidenceGraphUnlinkImpactError = CoreInvokeError.failed(status: 3)
        model.artifacts.beginEdit("P3")
        await model.artifacts.linkLoadTask?.value
        guard case .failed = model.artifacts.linkEdit?.loadState else {
            Issue.record("expected a failed read, got \(String(describing: model.artifacts.linkEdit?.loadState))")
            return
        }
        model.artifacts.labels["P3"] = "Changed"
        #expect(!model.artifacts.canSaveFields("P3"))
        store.getEvidenceGraphUnlinkImpactError = nil
        await model.artifacts.loadLinkLocks()
        #expect(model.artifacts.linkEdit?.loadState == .ready)
        #expect(model.artifacts.canSaveFields("P3"))
    }

    @Test func cardCancelDropsLinkChanges() async {
        let (model, store, _) = await makeCensus()
        model.artifacts.beginEdit("P3")
        await model.artifacts.linkLoadTask?.value
        model.artifacts.linkEdit?.excluded.insert("g-j")
        #expect(model.artifacts.fieldsDirty("P3"))
        model.artifacts.cancelFields(id: "P3")
        #expect(model.artifacts.linkEdit == nil)
        #expect(linkCalls(store).isEmpty)
        model.artifacts.beginEdit("P3")
        await model.artifacts.linkLoadTask?.value
        #expect(model.artifacts.checkedGraphIDs.contains("g-j"))
        #expect(!model.artifacts.fieldsDirty("P3"))
    }
}
