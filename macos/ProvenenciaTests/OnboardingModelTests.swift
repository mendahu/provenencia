import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct OnboardingModelTests {
    private let jane = InstallIdentity(
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jane Smith",
        ref: "USR-A1B2C"
    )
    private let jake = InstallIdentity(
        userID: "00000000-0000-7000-8000-000000000002",
        displayName: "Jake Robins",
        ref: "USR-D3E4F"
    )

    @Test func loadWithNoIdentityShowsChooseFileUnlocked() async throws {
        let (model, _, _) = try harness()
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(!model.researcherLocked)
        #expect(model.session == nil)
        #expect(model.error == nil)
    }

    @Test func loadWithIdentityAndNoActiveShowsChooseFileLocked() async throws {
        let (model, _, _) = try harness(identity: jane)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(model.researcherLocked)
        #expect(model.session == jane)
        #expect(model.error == nil)
    }

    @Test func loadWithIdentityAndExistingActiveDirShowsHome() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Robins Family.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        try touchCatalog(project)
        let store = FakeStore(identity: jane, activeProjectDir: project.path)
        store.projectInfos[project.path] = ProjectInfo(
            label: "Robins Family",
            folderName: "Robins Family.provenencia",
            createdAt: "2026-01-02T03:04:05Z",
            updatedAt: "2026-01-03T03:04:05Z",
            updatedByUserID: jane.userID,
            updatedByDisplayName: jane.displayName,
            updatedByRef: jane.ref
        )
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        #expect(model.phase == .home(projectDir: project.path, userID: jane.userID))
        #expect(model.project?.folderName == "Robins Family.provenencia")
        #expect(model.project?.label == "Robins Family")
        #expect(model.session?.ref == jane.ref)
        #expect(model.project?.updatedByRef == jane.ref)
        #expect(model.researcherLocked)
    }

    @Test func folderNamePreviewIsKebabSlug() async throws {
        let (model, _, _) = try harness()
        model.familyName = "Robins Family"
        #expect(model.folderNamePreview == "robins-family.provenencia")
    }

    @Test func createProjectStoresLabelAndSlugFolder() async throws {
        let (model, store, _) = try harness()
        await model.load()
        model.mode = .create
        model.phase = .identify
        model.displayName = "Jake"
        model.familyName = "Robins Family"
        await model.submit()
        #expect(model.phase.isHome)
        #expect(model.project?.label == "Robins Family")
        #expect(model.project?.folderName == "robins-family.provenencia")
        #expect(store.activeProjectDir?.hasSuffix("robins-family.provenencia") == true)
        #expect(model.session?.ref.isEmpty == false)
    }

    @Test func loadWithIdentityAndDirMissingCatalogClearsPointerAndShowsError() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Robins Family.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, activeProjectDir: project.path)
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(store.activeProjectDir == nil)
        #expect(model.error?.localizedDescription == String(localized: L10n.Onboarding.missingProject))
        #expect(model.researcherLocked)
    }

    @Test func loadWithMissingActivePathClearsPointerAndShowsError() async throws {
        let missing = FileManager.default.temporaryDirectory
            .appendingPathComponent("missing-\(UUID().uuidString).provenencia", isDirectory: true)
        let (model, store, _) = try harness(identity: jane, active: missing.path)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(store.activeProjectDir == nil)
        #expect(model.error?.localizedDescription == String(localized: L10n.Onboarding.missingProject))
        #expect(model.researcherLocked)
    }

    @Test func loadWhenStoreThrowsShowsErrorAndChooseFile() async {
        let folders = (try? makeFolders()) ?? OnboardingFolders.previewEmpty()
        let model = OnboardingModel(store: ThrowingStore(), folders: folders)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(model.error?.localizedDescription == String(localized: L10n.Errors.unknown))
    }

    @Test func canContinueCreateOpenAndIdentify() async throws {
        let (model, _, _) = try harness()
        await model.load()

        model.mode = .create
        #expect(model.canContinue)

        model.mode = .open
        #expect(!model.canContinue)
        model.selectedProject = URL(fileURLWithPath: "/tmp/x.provenencia")
        #expect(model.canContinue)

        model.phase = .identify
        model.mode = .create
        model.displayName = ""
        model.familyName = ""
        #expect(!model.canContinue)
        model.displayName = "Jane"
        model.familyName = "Smith"
        #expect(model.canContinue)

        model.session = jane
        model.displayName = ""
        model.familyName = "Smith"
        #expect(model.canContinue)

        model.mode = .open
        model.session = nil
        model.selectedContributorID = OnboardingModel.newContributorID
        model.displayName = ""
        #expect(!model.canContinue)
        model.displayName = "New person"
        #expect(model.canContinue)

        model.catalogUsers = [jane]
        model.selectedContributorID = jane.userID
        model.displayName = ""
        #expect(model.canContinue)
    }

    @Test func chooseAppendsURLAndSetsOpen() async throws {
        let (model, _, _) = try harness()
        let url = URL(fileURLWithPath: "/tmp/Picked.provenencia")
        await model.choose(url)
        #expect(model.mode == .open)
        #expect(model.selectedProject == url)
        #expect(model.availableProjects.map(\.path) == [url.path])
        await model.choose(url)
        #expect(model.availableProjects.count == 1)
    }

    @Test func goBackReturnsToChooseFile() async throws {
        let (model, _, _) = try harness()
        model.phase = .identify
        model.error = StoreBoom.boom
        model.catalogUsers = [jane]
        model.goBack()
        #expect(model.phase == .chooseFile)
        #expect(model.error == nil)
        #expect(model.catalogUsers.isEmpty)
    }

    @Test func submitCreateLandsHomeAndUpdatesStore() async throws {
        let (model, store, folders) = try harness()
        await model.load()
        model.mode = .create
        model.displayName = "Jane Smith"
        model.familyName = "Robins"
        await model.submit()
        #expect(model.phase == .identify)
        await model.submit()
        #expect(model.phase.isHome)
        #expect(model.project?.folderName == "robins.provenencia")
        #expect(model.project?.label == "Robins")
        #expect(store.identity?.displayName == "Jane Smith")
        #expect(model.session?.ref.isEmpty == false)
        #expect(store.activeProjectDir == folders.documentsDirectory.appendingPathComponent("robins.provenencia").path)
    }

    @Test func submitOpenWithAdoptCallsStoreAndLandsHome() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Shared.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: nil, catalogUsers: [jane, jake])
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        await model.choose(project)
        await model.submit()
        #expect(model.phase == .identify)
        #expect(model.selectedContributorID == OnboardingModel.newContributorID)
        model.selectedContributorID = jane.userID
        await model.submit()
        #expect(model.phase.isHome)
        #expect(store.identity == jane)
        #expect(store.activeProjectDir == project.path)
        #expect(model.session?.displayName == jane.displayName)
    }

    @Test func goToIdentifyPreselectsMatchingIdentityUser() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Shared.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, catalogUsers: [jane, jake])
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        await model.choose(project)
        await model.submit()
        #expect(model.phase == .identify)
        #expect(model.selectedContributorID == jane.userID)
    }

    @Test func goToIdentifyDefaultsToNewContributorWhenIdentityNotInCatalog() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Shared.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, catalogUsers: [jake])
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        await model.choose(project)
        await model.submit()
        #expect(model.selectedContributorID == OnboardingModel.newContributorID)
    }

    @Test func signOutClearsIdentityActiveAndReturnsUnlockedChooseFile() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Robins Family.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        try touchCatalog(project)
        let store = FakeStore(identity: jane, activeProjectDir: project.path)
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        #expect(model.phase.isHome)
        await model.signOut()
        #expect(model.phase == .chooseFile)
        #expect(!model.researcherLocked)
        #expect(model.session == nil)
        #expect(store.identity == nil)
        #expect(store.activeProjectDir == nil)
    }

    @Test func selectModeOpenFillsAvailableProjectsFromProvenenciaFolders() async throws {
        let folders = try makeFolders()
        let keep = folders.documentsDirectory.appendingPathComponent("Keep.provenencia", isDirectory: true)
        try FileManager.default.createDirectory(at: keep, withIntermediateDirectories: true)
        try touchCatalog(keep)
        try FileManager.default.createDirectory(
            at: folders.documentsDirectory.appendingPathComponent("Husk.provenencia", isDirectory: true),
            withIntermediateDirectories: true
        )
        try FileManager.default.createDirectory(
            at: folders.documentsDirectory.appendingPathComponent("plain", isDirectory: true),
            withIntermediateDirectories: true
        )
        try Data("x".utf8).write(to: folders.documentsDirectory.appendingPathComponent("file.provenencia"))
        let model = OnboardingModel(store: FakeStore(), folders: folders)
        await model.selectMode(.open)
        #expect(model.mode == .open)
        #expect(model.availableProjects.map(\.lastPathComponent).sorted() == ["Keep.provenencia"])
    }

    private func harness(
        identity: InstallIdentity? = nil,
        active: String? = nil
    ) throws -> (OnboardingModel, FakeStore, OnboardingFolders) {
        let folders = try makeFolders()
        let store = FakeStore(identity: identity, activeProjectDir: active)
        return (OnboardingModel(store: store, folders: folders), store, folders)
    }

    private func makeFolders() throws -> OnboardingFolders {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("onboarding-\(UUID().uuidString)", isDirectory: true)
        let identity = root.appendingPathComponent("identity", isDirectory: true)
        let documents = root.appendingPathComponent("Documents", isDirectory: true)
        try FileManager.default.createDirectory(at: identity, withIntermediateDirectories: true)
        try FileManager.default.createDirectory(at: documents, withIntermediateDirectories: true)
        return OnboardingFolders(identityDirectory: identity, documentsDirectory: documents)
    }
}

private func touchCatalog(_ project: URL) throws {
    try Data().write(to: project.appendingPathComponent(InstallPaths.catalogFile))
}

private enum StoreBoom: Error {
    case boom
}

private struct ThrowingStore: GenealogyStore {
    func installIdentity(identityDir _: String) async throws -> InstallIdentity? {
        throw CoreInvokeError.coded(status: 1, code: "internal.unknown", kind: .internal, params: [])
    }
    func completeOnboarding(
        identityDir _: String,
        parentDir _: String,
        displayName _: String,
        familyName _: String
    ) async throws -> OnboardingResult { throw StoreBoom.boom }
    func activeProject(identityDir _: String) async throws -> String? { throw StoreBoom.boom }
    func listProjectUsers(projectDir _: String) async throws -> [InstallIdentity] { throw StoreBoom.boom }
    func openProject(
        identityDir _: String,
        projectDir _: String,
        displayName _: String,
        adoptUserID _: String
    ) async throws -> OnboardingResult { throw StoreBoom.boom }
    func removeActiveProject(identityDir _: String) async throws { throw StoreBoom.boom }
    func signOut(identityDir _: String) async throws { throw StoreBoom.boom }
    func closeCatalogSession(projectDir _: String) async throws { throw StoreBoom.boom }
    func projectInfo(projectDir _: String) async throws -> ProjectInfo { throw StoreBoom.boom }
    func listSources(projectDir _: String) async throws -> [CatalogSource] { throw StoreBoom.boom }
    func getSourceWorkspace(projectDir _: String, sourceID _: String) async throws -> CatalogSourceWorkspace {
        throw StoreBoom.boom
    }
    func createSource(
        projectDir _: String,
        userID _: String,
        sourceTypeID _: String,
        title _: String,
        description _: String
    ) async throws -> CatalogSource { throw StoreBoom.boom }
    func updateSource(
        projectDir _: String,
        userID _: String,
        sourceID _: String,
        sourceTypeID _: String,
        title _: String,
        description _: String
    ) async throws -> CatalogSource { throw StoreBoom.boom }
    func addSourceNote(projectDir _: String, userID _: String, sourceID _: String, body _: String) async throws
        -> CatalogSourceNote { throw StoreBoom.boom }
    func updateSourceNote(projectDir _: String, userID _: String, noteID _: String, body _: String) async throws
        -> CatalogSourceNote { throw StoreBoom.boom }
    func deleteSourceNote(projectDir _: String, userID _: String, noteID _: String) async throws {
        throw StoreBoom.boom
    }
    func setSourceMetadata(
        projectDir _: String,
        userID _: String,
        sourceID _: String,
        fieldID _: String,
        valueText _: String,
        date _: CatalogDateValueInput?
    ) async throws -> CatalogMetadataEntry { throw StoreBoom.boom }
    func clearSourceMetadata(projectDir _: String, userID _: String, sourceID _: String, fieldID _: String)
        async throws { throw StoreBoom.boom }
    func dismissSourceMetadataSuggestion(
        projectDir _: String,
        userID _: String,
        sourceID _: String,
        fieldID _: String
    ) async throws -> [CatalogMetadataEntry] { throw StoreBoom.boom }
    func reorderSourceMetadata(
        projectDir _: String,
        userID _: String,
        sourceID _: String,
        fieldIDs _: [String]
    ) async throws -> [CatalogMetadataEntry] { throw StoreBoom.boom }
    func createArtifact(
        projectDir _: String,
        userID _: String,
        sourceID _: String,
        fileID _: String,
        label _: String,
        description _: String
    ) async throws -> CatalogArtifact { throw StoreBoom.boom }
    func updateArtifact(
        projectDir _: String,
        userID _: String,
        artifactID _: String,
        label _: String,
        description _: String
    ) async throws -> CatalogArtifact { throw StoreBoom.boom }
    func ingestArtifactFile(
        projectDir _: String,
        userID _: String,
        artifactID _: String,
        path _: String
    ) async throws -> (artifact: CatalogArtifact, file: CatalogFileRef, reused: Bool) { throw StoreBoom.boom }
    func ensureFileThumbnail(
        projectDir _: String,
        fileID _: String
    ) async throws -> (relPath: String, skipped: Bool) { throw StoreBoom.boom }
    func listSourceCredibilityGrades(projectDir _: String) async throws -> [CatalogCredibilityGrade] {
        throw StoreBoom.boom
    }
    func upsertSourceCredibilityAssessment(
        projectDir _: String,
        userID _: String,
        sourceID _: String,
        gradeID _: String,
        argument _: String
    ) async throws -> CatalogCredibilityAssessment { throw StoreBoom.boom }
    func listSourceTypes(projectDir _: String) async throws -> [CatalogSourceType] { throw StoreBoom.boom }
    func createSourceType(
        projectDir _: String,
        userID _: String,
        label _: String,
        description _: String,
        iconKey _: String
    ) async throws -> CatalogSourceType { throw StoreBoom.boom }
    func updateSourceType(
        projectDir _: String,
        userID _: String,
        typeID _: String,
        label _: String,
        description _: String,
        iconKey _: String
    ) async throws -> CatalogSourceType { throw StoreBoom.boom }
    func deleteSourceType(
        projectDir _: String,
        userID _: String,
        typeID _: String
    ) async throws { throw StoreBoom.boom }
    func listTypeSuggestions(projectDir _: String, typeID _: String) async throws
        -> [CatalogTypeSuggestion] { throw StoreBoom.boom }
    func assignTypeField(
        projectDir _: String,
        userID _: String,
        typeID _: String,
        fieldID _: String
    ) async throws -> [CatalogTypeSuggestion] { throw StoreBoom.boom }
    func removeTypeField(
        projectDir _: String,
        userID _: String,
        typeID _: String,
        fieldID _: String
    ) async throws -> [CatalogTypeSuggestion] { throw StoreBoom.boom }
    func listMetadataFields(projectDir _: String) async throws -> [CatalogMetadataField] { throw StoreBoom.boom }
    func createMetadataField(
        projectDir _: String,
        userID _: String,
        label _: String,
        dataType _: String,
        description _: String
    ) async throws -> CatalogMetadataField { throw StoreBoom.boom }
    func updateMetadataField(
        projectDir _: String,
        userID _: String,
        fieldID _: String,
        label _: String,
        dataType _: String,
        description _: String
    ) async throws -> CatalogMetadataField { throw StoreBoom.boom }
    func deleteMetadataField(
        projectDir _: String,
        userID _: String,
        fieldID _: String
    ) async throws { throw StoreBoom.boom }
    func countFiles(projectDir _: String) async throws -> Int { throw StoreBoom.boom }
    func workspaceNavCounts(projectDir _: String) async throws -> WorkspaceNavCounts {
        throw StoreBoom.boom
    }
}
