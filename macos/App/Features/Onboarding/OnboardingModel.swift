import Foundation
import Observation

@Observable
final class OnboardingModel {
    enum Phase {
        case loading
        case form
        case home
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

    private let store: any GenealogyStore
    private let folders: OnboardingFolders?

    init(store: any GenealogyStore, folders: OnboardingFolders? = nil) {
        self.store = store
        self.folders = folders
    }

    var canContinue: Bool {
        let familyOK = !trimmed(familyName).isEmpty
        let nameOK = researcherLocked || !trimmed(displayName).isEmpty
        return nameOK && familyOK && !isBusy
    }

    func load() async {
        errorText = nil
        researcherLocked = false
        do {
            let loc = try resolvedFolders()
            let projects = try InstallPaths.provenanceProjects(in: loc.documentsDirectory)
            if let id = try await store.installIdentity(identityDir: loc.identityDirectory.path) {
                sessionDisplayName = id.displayName
                displayName = id.displayName
                researcherLocked = true
                if projects.isEmpty {
                    phase = .form
                } else {
                    projectBasename = ""
                    phase = .home
                }
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
            sessionDisplayName = result.displayName
            projectBasename = URL(fileURLWithPath: result.projectDir).lastPathComponent
            phase = .home
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
            try await store.removeInstallIdentity(identityDir: loc.identityDirectory.path)
            displayName = ""
            familyName = ""
            sessionDisplayName = ""
            projectBasename = ""
            researcherLocked = false
            phase = .form
        } catch {
            errorText = error.localizedDescription
        }
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
