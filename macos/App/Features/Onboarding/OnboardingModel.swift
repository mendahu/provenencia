import AppKit
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
    var error: Error?
    var phase: Phase = .loading
    /// Install identity for this Mac once known (locks the researcher name field).
    var session: InstallIdentity?
    /// Catalog project metadata for the active or selected project.
    var project: ProjectInfo?
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

    var researcherLocked: Bool { session != nil }

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
        error = nil
        session = nil
        selectedProject = nil
        catalogUsers = []
        project = nil
        do {
            let identityDir = try identityDir()
            if let id = try await store.installIdentity(identityDir: identityDir.path) {
                session = id
                displayName = id.displayName
                if let active = try await store.activeProject(identityDir: identityDir.path) {
                    if InstallPaths.isProjectDirectory(active) {
                        await loadProjectInfo(path: active)
                        phase = .home
                        return
                    }
                    try await store.removeActiveProject(identityDir: identityDir.path)
                    present(L10n.Onboarding.missingProject)
                }
                phase = .chooseFile
                return
            }
            phase = .chooseFile
        } catch {
            present(error)
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
        error = nil
        catalogUsers = []
        phase = .chooseFile
    }

    func selectMode(_ mode: Mode) async {
        self.mode = mode
        if mode == .open {
            await refreshDocumentsProjects()
            await refreshSelectedProjectInfo()
        } else {
            project = nil
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

    /// Presents the AppKit folder picker and adopts the chosen project directory.
    func chooseFolder() {
        let panel = NSOpenPanel()
        panel.canChooseFiles = false
        panel.canChooseDirectories = true
        panel.allowsMultipleSelection = false
        panel.treatsFilePackagesAsDirectories = true
        panel.prompt = String(localized: L10n.Onboarding.openPanelPrompt)
        panel.message = String(localized: L10n.Onboarding.openPanelMessage)
        guard panel.runModal() == .OK, let url = panel.url else {
            return
        }
        Task { await choose(url) }
    }

    /// Loads catalog project metadata for the current picker selection (open flow).
    func refreshSelectedProjectInfo() async {
        guard let selectedProject else {
            project = nil
            return
        }
        await loadProjectInfo(path: selectedProject.path)
    }

    func signOut() async {
        isBusy = true
        error = nil
        defer { isBusy = false }
        do {
            let identityDir = try identityDir()
            try await store.signOut(identityDir: identityDir.path)
            displayName = ""
            familyName = ""
            session = nil
            project = nil
            selectedProject = nil
            catalogUsers = []
            selectedContributorID = Self.newContributorID
            availableProjects = []
            mode = .create
            phase = .chooseFile
        } catch {
            present(error)
        }
    }

    private func refreshDocumentsProjects() async {
        do {
            let loc = try resolvedFolders()
            availableProjects = try InstallPaths.provenenciaProjects(in: loc.documentsDirectory)
        } catch {
            present(error)
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
        error = nil
        defer { isBusy = false }
        do {
            catalogUsers = try await store.listProjectUsers(projectDir: selectedProject.path)
            if let userID = session?.userID, catalogUsers.contains(where: { $0.userID == userID }) {
                selectedContributorID = userID
            } else {
                selectedContributorID = Self.newContributorID
            }
            phase = .identify
        } catch {
            present(error)
        }
    }

    private func createProject() async {
        isBusy = true
        error = nil
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
            present(error)
        }
    }

    private func open(_ url: URL) async {
        isBusy = true
        error = nil
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
            present(error)
        }
    }

    private func applyOpened(_ result: OnboardingResult) {
        session = InstallIdentity(userID: result.userID, displayName: result.displayName, ref: result.ref)
        project = normalizedProject(result.project, path: result.projectDir)
        displayName = result.displayName
        phase = .home
    }

    private func loadProjectInfo(path: String) async {
        do {
            let info = try await store.projectInfo(projectDir: path)
            project = normalizedProject(info, path: path)
        } catch {
            project = ProjectInfo(
                label: ProjectSlug.labelFromFolder(path),
                folderName: URL(fileURLWithPath: path).lastPathComponent,
                createdAt: "",
                updatedAt: "",
                updatedByUserID: "",
                updatedByDisplayName: "",
                updatedByRef: ""
            )
            present(error)
        }
    }

    private func normalizedProject(_ info: ProjectInfo, path: String) -> ProjectInfo {
        var info = info
        if info.folderName.isEmpty {
            info.folderName = URL(fileURLWithPath: path).lastPathComponent
        }
        return info
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

    private func present(_ error: Error) {
        self.error = error
    }

    private func present(_ resource: LocalizedStringResource) {
        self.error = LocalizedMessageError(resource)
    }
}

/// Client-only localized message (e.g. missing active project) for toast display.
private struct LocalizedMessageError: LocalizedError {
    let resource: LocalizedStringResource

    init(_ resource: LocalizedStringResource) {
        self.resource = resource
    }

    var errorDescription: String? {
        String(localized: resource)
    }
}
