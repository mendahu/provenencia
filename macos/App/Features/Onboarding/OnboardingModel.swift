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

    private let store: any GenealogyStore

    init(store: any GenealogyStore) {
        self.store = store
    }

    var canContinue: Bool {
        !trimmed(displayName).isEmpty && !trimmed(familyName).isEmpty && !isBusy
    }

    func load() async {
        errorText = nil
        do {
            let dir = try InstallPaths.identityDirectory().path
            if let id = try await store.installIdentity(identityDir: dir) {
                sessionDisplayName = id.displayName
                projectBasename = ""
                phase = .home
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
            let identityDir = try InstallPaths.identityDirectory().path
            let parentDir = try InstallPaths.documentsDirectory().path
            let result = try await store.completeOnboarding(
                identityDir: identityDir,
                parentDir: parentDir,
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

    private func trimmed(_ s: String) -> String {
        s.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
