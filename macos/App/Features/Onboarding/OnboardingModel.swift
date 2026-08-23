import Foundation
import Observation

@Observable
final class OnboardingModel {
    enum Phase {
        case loading
        case form
        case home
    }

    enum Mode: String, CaseIterable, Identifiable {
        case create
        case open

        var id: String { rawValue }
    }

    var displayName = ""
    var familyName = ""
    var isBusy = false
    var errorText: String?
    var phase: Phase = .loading
    var sessionDisplayName = ""
    var projectBasename = ""
    /// True when identity.json loaded; researcher name is not collected again.
    var researcherLocked = false
    var mode: Mode = .create
    var availableProjects: [URL] = []
    var selectedProject: URL?

    private let store: any GenealogyStore
    private let folders: OnboardingFolders?

    init(store: any GenealogyStore, folders: OnboardingFolders? = nil) {
        self.store = store
        self.folders = folders
    }

    var canContinue: Bool {
        let nameOK = researcherLocked || !trimmed(displayName).isEmpty
        guard nameOK, !isBusy else {
            return false
        }
        switch mode {
        case .create:
            return !trimmed(familyName).isEmpty
        case .open:
            return selectedProject != nil
        }
    }

    func load() async {
        errorText = nil
        researcherLocked = false
        selectedProject = nil
        do {
            let loc = try resolvedFolders()
            availableProjects = try InstallPaths.provenanceProjects(in: loc.documentsDirectory)
            if let id = try await store.installIdentity(identityDir: loc.identityDirectory.path) {
                sessionDisplayName = id.displayName
                displayName = id.displayName
                researcherLocked = true
                if let active = try await store.activeProject(identityDir: loc.identityDirectory.path) {
                    if InstallPaths.isProjectDirectory(active) {
                        projectBasename = URL(fileURLWithPath: active).lastPathComponent
                        phase = .home
                        return
                    }
                    try await store.removeActiveProject(identityDir: loc.identityDirectory.path)
                    errorText = "The last project could not be found. Create or open a project."
                }
                phase = .form
                return
            }
            phase = .form
        } catch {
            errorText = error.localizedDescription
            phase = .form
        }
    }

    func submit() async {
        guard canContinue else {
            return
        }
        switch mode {
        case .create:
            await createProject()
        case .open:
            if let selectedProject {
                await open(selectedProject)
            }
        }
    }

    func open(_ url: URL) async {
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            let loc = try resolvedFolders()
            let result = try await store.openProject(
                identityDir: loc.identityDirectory.path,
                projectDir: url.path,
                displayName: trimmed(displayName)
            )
            applyOpened(result)
        } catch {
            errorText = error.localizedDescription
        }
    }

    func signOut() async {
        isBusy = true
        errorText = nil
        defer { isBusy = false }
        do {
            let loc = try resolvedFolders()
            try await store.removeActiveProject(identityDir: loc.identityDirectory.path)
            try await store.removeInstallIdentity(identityDir: loc.identityDirectory.path)
            displayName = ""
            familyName = ""
            sessionDisplayName = ""
            projectBasename = ""
            researcherLocked = false
            selectedProject = nil
            availableProjects = try InstallPaths.provenanceProjects(in: loc.documentsDirectory)
            mode = .create
            phase = .form
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

    private func applyOpened(_ result: OnboardingResult) {
        sessionDisplayName = result.displayName
        projectBasename = URL(fileURLWithPath: result.projectDir).lastPathComponent
        researcherLocked = true
        displayName = result.displayName
        phase = .home
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
