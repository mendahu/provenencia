import Foundation
import Testing
@testable import Provenance

@Suite
@MainActor
struct OnboardingModelTests {
    private let jane = InstallIdentity(
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jane Smith"
    )
    private let jake = InstallIdentity(
        userID: "00000000-0000-7000-8000-000000000002",
        displayName: "Jake Robins"
    )

    @Test func loadWithNoIdentityShowsChooseFileUnlocked() async throws {
        let (model, _, _) = try harness()
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(!model.researcherLocked)
        #expect(model.identityUserID.isEmpty)
        #expect(model.errorText == nil)
    }

    @Test func loadWithIdentityAndNoActiveShowsChooseFileLocked() async throws {
        let (model, _, _) = try harness(identity: jane)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(model.researcherLocked)
        #expect(model.identityUserID == jane.userID)
        #expect(model.sessionDisplayName == jane.displayName)
        #expect(model.errorText == nil)
    }

    @Test func loadWithIdentityAndExistingActiveDirShowsHome() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Robins Family.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, activeProjectDir: project.path)
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        #expect(model.phase == .home)
        #expect(model.projectBasename == "Robins Family.provenance")
        #expect(model.researcherLocked)
    }

    @Test func loadWithMissingActivePathClearsPointerAndShowsError() async throws {
        let missing = FileManager.default.temporaryDirectory
            .appendingPathComponent("missing-\(UUID().uuidString).provenance", isDirectory: true)
        let (model, store, _) = try harness(identity: jane, active: missing.path)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(store.activeProjectDir == nil)
        #expect(model.errorText == "The last project could not be found. Create or open a project.")
        #expect(model.researcherLocked)
    }

    @Test func loadWhenStoreThrowsShowsErrorAndChooseFile() async {
        let folders = (try? makeFolders()) ?? OnboardingFolders.previewEmpty
        let model = OnboardingModel(store: ThrowingStore(), folders: folders)
        await model.load()
        #expect(model.phase == .chooseFile)
        #expect(model.errorText != nil)
    }

    @Test func canContinueCreateOpenAndIdentify() async throws {
        let (model, _, _) = try harness()
        await model.load()

        model.mode = .create
        #expect(model.canContinue)

        model.mode = .open
        #expect(!model.canContinue)
        model.selectedProject = URL(fileURLWithPath: "/tmp/x.provenance")
        #expect(model.canContinue)

        model.phase = .identify
        model.mode = .create
        model.displayName = ""
        model.familyName = ""
        #expect(!model.canContinue)
        model.displayName = "Jane"
        model.familyName = "Smith"
        #expect(model.canContinue)

        model.researcherLocked = true
        model.displayName = ""
        model.familyName = "Smith"
        #expect(model.canContinue)

        model.mode = .open
        model.researcherLocked = false
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
        let url = URL(fileURLWithPath: "/tmp/Picked.provenance")
        model.choose(url)
        #expect(model.mode == .open)
        #expect(model.selectedProject == url)
        #expect(model.availableProjects.map(\.path) == [url.path])
        model.choose(url)
        #expect(model.availableProjects.count == 1)
    }

    @Test func goBackReturnsToChooseFile() async throws {
        let (model, _, _) = try harness()
        model.phase = .identify
        model.errorText = "x"
        model.catalogUsers = [jane]
        model.goBack()
        #expect(model.phase == .chooseFile)
        #expect(model.errorText == nil)
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
        #expect(model.phase == .home)
        #expect(model.projectBasename == "Robins.provenance")
        #expect(store.identity?.displayName == "Jane Smith")
        #expect(store.activeProjectDir == folders.documentsDirectory.appendingPathComponent("Robins.provenance").path)
    }

    @Test func submitOpenWithAdoptCallsStoreAndLandsHome() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Shared.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: nil, catalogUsers: [jane, jake])
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        model.choose(project)
        await model.submit()
        #expect(model.phase == .identify)
        #expect(model.selectedContributorID == OnboardingModel.newContributorID)
        model.selectedContributorID = jane.userID
        await model.submit()
        #expect(model.phase == .home)
        #expect(store.identity == jane)
        #expect(store.activeProjectDir == project.path)
        #expect(model.sessionDisplayName == jane.displayName)
    }

    @Test func goToIdentifyPreselectsMatchingIdentityUser() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Shared.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, catalogUsers: [jane, jake])
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        model.choose(project)
        await model.submit()
        #expect(model.phase == .identify)
        #expect(model.selectedContributorID == jane.userID)
    }

    @Test func goToIdentifyDefaultsToNewContributorWhenIdentityNotInCatalog() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Shared.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, catalogUsers: [jake])
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        model.choose(project)
        await model.submit()
        #expect(model.selectedContributorID == OnboardingModel.newContributorID)
    }

    @Test func signOutClearsIdentityActiveAndReturnsUnlockedChooseFile() async throws {
        let folders = try makeFolders()
        let project = folders.documentsDirectory.appendingPathComponent("Robins Family.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        let store = FakeStore(identity: jane, activeProjectDir: project.path)
        let model = OnboardingModel(store: store, folders: folders)
        await model.load()
        #expect(model.phase == .home)
        await model.signOut()
        #expect(model.phase == .chooseFile)
        #expect(!model.researcherLocked)
        #expect(model.identityUserID.isEmpty)
        #expect(store.identity == nil)
        #expect(store.activeProjectDir == nil)
    }

    @Test func selectModeOpenFillsAvailableProjectsFromProvenanceFolders() async throws {
        let folders = try makeFolders()
        let keep = folders.documentsDirectory.appendingPathComponent("Keep.provenance", isDirectory: true)
        try FileManager.default.createDirectory(at: keep, withIntermediateDirectories: true)
        try FileManager.default.createDirectory(
            at: folders.documentsDirectory.appendingPathComponent("plain", isDirectory: true),
            withIntermediateDirectories: true
        )
        try Data("x".utf8).write(to: folders.documentsDirectory.appendingPathComponent("file.provenance"))
        let model = OnboardingModel(store: FakeStore(), folders: folders)
        await model.selectMode(.open)
        #expect(model.mode == .open)
        #expect(model.availableProjects.map(\.lastPathComponent).sorted() == ["Keep.provenance"])
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

private enum StoreBoom: Error, LocalizedError {
    case boom
    var errorDescription: String? { "store failed" }
}

private struct ThrowingStore: GenealogyStore {
    func ping(_ message: String) async throws -> String { throw StoreBoom.boom }
    func version() async throws -> String { throw StoreBoom.boom }
    func installIdentity(identityDir _: String) async throws -> InstallIdentity? { throw StoreBoom.boom }
    func completeOnboarding(
        identityDir _: String,
        parentDir _: String,
        displayName _: String,
        familyName _: String
    ) async throws -> OnboardingResult { throw StoreBoom.boom }
    func removeInstallIdentity(identityDir _: String) async throws { throw StoreBoom.boom }
    func activeProject(identityDir _: String) async throws -> String? { throw StoreBoom.boom }
    func listProjectUsers(projectDir _: String) async throws -> [InstallIdentity] { throw StoreBoom.boom }
    func openProject(
        identityDir _: String,
        projectDir _: String,
        displayName _: String,
        adoptUserID _: String
    ) async throws -> OnboardingResult { throw StoreBoom.boom }
    func removeActiveProject(identityDir _: String) async throws { throw StoreBoom.boom }
}
