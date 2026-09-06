import Foundation

struct InstallIdentity: Sendable, Equatable {
    var userID: String
    var displayName: String
    var ref: String
}

struct ProjectInfo: Sendable, Equatable {
    var label: String
    var folderName: String
    var createdAt: String
    var updatedAt: String
    var updatedByUserID: String
    var updatedByDisplayName: String
    var updatedByRef: String
}

struct OnboardingResult: Sendable, Equatable {
    var projectDir: String
    var userID: String
    var displayName: String
    var ref: String
    var project: ProjectInfo
}

protocol GenealogyStore: Sendable {
    func installIdentity(identityDir: String) async throws -> InstallIdentity?
    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult
    func signOut(identityDir: String) async throws
    func activeProject(identityDir: String) async throws -> String?
    func listProjectUsers(projectDir: String) async throws -> [InstallIdentity]
    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult
    func removeActiveProject(identityDir: String) async throws
    func projectInfo(projectDir: String) async throws -> ProjectInfo
}
