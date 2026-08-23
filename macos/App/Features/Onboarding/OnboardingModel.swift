import Foundation
import Observation

@Observable
final class OnboardingModel {
    enum Phase {
        case loading
        case chooseFile
        case identify
        case home
    }

    enum Mode: String, CaseIterable, Identifiable {
        case create
        case open

        var id: String { rawValue }
    }

    /// Empty string means “add a new contributor.”
    static let newContributorID = ""

    var displayName = ""
    var familyName = ""
    var isBusy = false
    var errorText: String?
    var phase: Phase = .loading
    var sessionDisplayName = ""
    var projectBasename = ""
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
        do {
            let identityDir = try identityDir()
            if let id = try await store.installIdentity(identityDir: identityDir.path) {
                sessionDisplayName = id.displayName
                displayName = id.displayName
                identityUserID = id.userID
                researcherLocked = true
                if let active = try await store.activeProject(identityDir: identityDir.path) {
                    if InstallPaths.isProjectDirectory(active) {
                        projectBasename = URL(fileURLWithPath: active).lastPathComponent
                        phase = .home
                        return
                    }
                    try await store.removeActiveProject(identityDir: identityDir.path)
                    errorText = "The last project could not be found. Create or open a project."
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

    func selectMode(_ mode: Mode) {
        self.mode = mode
        if mode == .open {
            Task { await refreshDocumentsProjects() }
        }
    }

    func choose(_ url: URL) {
        selectedProject = url
        if !availableProjects.contains(where: { $0.path == url.path }) {
            availableProjects.append(url)
        }
        mode = .open
    }

    func signOut() async {
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            let identityDir = try identityDir()
            try await store.removeActiveProject(identityDir: identityDir.path)
            try await store.removeInstallIdentity(identityDir: identityDir.path)
            displayName = ""
            familyName = ""
            sessionDisplayName = ""
            projectBasename = ""
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
            availableProjects = try InstallPaths.provenanceProjects(in: loc.documentsDirectory)
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
        projectBasename = URL(fileURLWithPath: result.projectDir).lastPathComponent
        researcherLocked = true
        displayName = result.displayName
        phase = .home
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
