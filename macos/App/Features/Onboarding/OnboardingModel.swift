import Foundation
import Observation

@MainActor
@Observable
final class OnboardingModel {
    enum Phase {
        case loading
        case chooseFile
        case identify
        case home
    }

    enum Mode: String {
        case create
        case open
    }

    /// Empty string means “add a new contributor.”
    static let newContributorID = ""

    var displayName = ""
    var familyName = ""
    var isBusy = false
    var errorText: String?
    var phase: Phase = .loading
    var sessionDisplayName = ""
    var sessionRef = ""
    var projectBasename = ""
    var projectLabel = ""
    var projectCreatedAt = ""
    var projectUpdatedAt = ""
    var projectUpdatedByDisplayName = ""
    var projectUpdatedByRef = ""
    var researcherLocked = false
    var identityUserID = ""
    var mode: Mode = .create
    var availableProjects: [URL] = []
    var selectedProject: URL?
    var catalogUsers: [InstallIdentity] = []
    var selectedContributorID: String = newContributorID

    private let store: any GenealogyStore
    private let folders: OnboardingFolders?

    init(store: any GenealogyStore, folders: OnboardingFolders? = nil) {
        self.store = store
        self.folders = folders
    }

    var folderNamePreview: String {
        let trimmedLabel = trimmed(familyName)
        guard !trimmedLabel.isEmpty else {
            return ""
        }
        return ProjectSlug.folderName(from: trimmedLabel)
    }

    var canContinue: Bool {
        guard !isBusy else {
            return false
        }
        switch phase {
        case .chooseFile:
            return mode == .create || selectedProject != nil
        case .identify:
            if mode == .create {
                let nameOK = researcherLocked || !trimmed(displayName).isEmpty
                return nameOK && !trimmed(familyName).isEmpty
            }
            if selectedContributorID == Self.newContributorID {
                return researcherLocked || !trimmed(displayName).isEmpty
            }
            return catalogUsers.contains { $0.userID == selectedContributorID }
        default:
            return false
        }
    }

    func load() async {
        errorText = nil
        researcherLocked = false
        identityUserID = ""
        selectedProject = nil
        catalogUsers = []
        clearProjectMeta()
        do {
            let identityDir = try identityDir()
            if let id = try await store.installIdentity(identityDir: identityDir.path) {
                sessionDisplayName = id.displayName
                sessionRef = id.ref
                displayName = id.displayName
                identityUserID = id.userID
                researcherLocked = true
                if let active = try await store.activeProject(identityDir: identityDir.path) {
                    if InstallPaths.isProjectDirectory(active) {
                        await applyProjectInfo(path: active)
                        phase = .home
                        return
                    }
                    try await store.removeActiveProject(identityDir: identityDir.path)
                    errorText = String(localized: L10n.Onboarding.missingProject)
                }
                phase = .chooseFile
                return
            }
            phase = .chooseFile
        } catch {
            errorText = error.localizedDescription
            phase = .chooseFile
        }
    }

    func submit() async {
        guard canContinue else {
            return
        }
        switch phase {
        case .chooseFile:
            await goToIdentify()
        case .identify:
            if mode == .create {
                await createProject()
            } else if let selectedProject {
                await open(selectedProject)
            }
        default:
            break
        }
    }

    func goBack() {
        errorText = nil
        catalogUsers = []
        phase = .chooseFile
    }

    func selectMode(_ mode: Mode) async {
        self.mode = mode
        if mode == .open {
            await refreshDocumentsProjects()
            await refreshSelectedProjectInfo()
        } else {
            clearProjectMeta()
        }
    }

    func choose(_ url: URL) async {
        selectedProject = url
        if !availableProjects.contains(where: { $0.path == url.path }) {
            availableProjects.append(url)
        }
        mode = .open
        await refreshSelectedProjectInfo()
    }

    /// Loads catalog project metadata for the current picker selection (open flow).
    func refreshSelectedProjectInfo() async {
        guard let selectedProject else {
            clearProjectMeta()
            return
        }
        await applyProjectInfo(path: selectedProject.path)
    }

    func signOut() async {
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            let identityDir = try identityDir()
            try await store.signOut(identityDir: identityDir.path)
            displayName = ""
            familyName = ""
            sessionDisplayName = ""
            sessionRef = ""
            clearProjectMeta()
            identityUserID = ""
            researcherLocked = false
            selectedProject = nil
            catalogUsers = []
            selectedContributorID = Self.newContributorID
            availableProjects = []
            mode = .create
            phase = .chooseFile
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func refreshDocumentsProjects() async {
        do {
            let loc = try resolvedFolders()
            availableProjects = try InstallPaths.provenenciaProjects(in: loc.documentsDirectory)
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func goToIdentify() async {
        if mode == .create {
            selectedContributorID = Self.newContributorID
            phase = .identify
            return
        }
        guard let selectedProject else {
            return
        }
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            catalogUsers = try await store.listProjectUsers(projectDir: selectedProject.path)
            if !identityUserID.isEmpty, catalogUsers.contains(where: { $0.userID == identityUserID }) {
                selectedContributorID = identityUserID
            } else {
                selectedContributorID = Self.newContributorID
            }
            phase = .identify
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func createProject() async {
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            let loc = try resolvedFolders()
            let result = try await store.completeOnboarding(
                identityDir: loc.identityDirectory.path,
                parentDir: loc.documentsDirectory.path,
                displayName: trimmed(displayName),
                familyName: trimmed(familyName)
            )
            applyOpened(result)
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func open(_ url: URL) async {
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            let loc = try resolvedFolders()
            let adopt = selectedContributorID == Self.newContributorID ? "" : selectedContributorID
            let result = try await store.openProject(
                identityDir: loc.identityDirectory.path,
                projectDir: url.path,
                displayName: trimmed(displayName),
                adoptUserID: adopt
            )
            applyOpened(result)
        } catch {
            errorText = error.localizedDescription
        }
    }

    private func applyOpened(_ result: OnboardingResult) {
        sessionDisplayName = result.displayName
        sessionRef = result.ref
        identityUserID = result.userID
        applyProject(result.project, path: result.projectDir)
        researcherLocked = true
        displayName = result.displayName
        phase = .home
    }

    private func applyProjectInfo(path: String) async {
        projectBasename = URL(fileURLWithPath: path).lastPathComponent
        do {
            let info = try await store.projectInfo(projectDir: path)
            applyProject(info, path: path)
        } catch {
            clearProjectMeta()
            projectBasename = URL(fileURLWithPath: path).lastPathComponent
            projectLabel = ProjectSlug.labelFromFolder(path)
            errorText = error.localizedDescription
        }
    }

    private func applyProject(_ info: ProjectInfo, path: String) {
        projectBasename = info.folderName.isEmpty
            ? URL(fileURLWithPath: path).lastPathComponent
            : info.folderName
        projectLabel = info.label
        projectCreatedAt = info.createdAt
        projectUpdatedAt = info.updatedAt
        projectUpdatedByDisplayName = info.updatedByDisplayName
        projectUpdatedByRef = info.updatedByRef
    }

    private func clearProjectMeta() {
        projectBasename = ""
        projectLabel = ""
        projectCreatedAt = ""
        projectUpdatedAt = ""
        projectUpdatedByDisplayName = ""
        projectUpdatedByRef = ""
    }

    private func identityDir() throws -> URL {
        if let folders {
            return folders.identityDirectory
        }
        return try OnboardingFolders.liveIdentity()
    }

    private func resolvedFolders() throws -> OnboardingFolders {
        if let folders {
            return folders
        }
        return try OnboardingFolders.live()
    }

    private func trimmed(_ s: String) -> String {
        s.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
