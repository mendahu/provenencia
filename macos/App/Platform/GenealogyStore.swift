import Foundation

struct InstallIdentity: Sendable, Equatable {
    var userID: String
    var displayName: String
}

struct OnboardingResult: Sendable, Equatable {
    var projectDir: String
    var userID: String
    var displayName: String
}

protocol GenealogyStore: Sendable {
    func ping(_ message: String) async throws -> String
    func version() async throws -> String
    func installIdentity(identityDir: String) async throws -> InstallIdentity?
    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult
    func removeInstallIdentity(identityDir: String) async throws
    func activeProject(identityDir: String) async throws -> String?
    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String
    ) async throws -> OnboardingResult
    func removeActiveProject(identityDir: String) async throws
}
