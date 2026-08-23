import Foundation

/// In-memory `GenealogyStore` for SwiftUI previews and tests. Not used in the shipped app.
final class FakeStore: GenealogyStore, @unchecked Sendable {
    var identity: InstallIdentity?
    var activeProjectDir: String?
    var catalogUsers: [InstallIdentity]
    var lastResult = OnboardingResult(
        projectDir: "/tmp/Robins Family.provenance",
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jake Robins"
    )

    init(
        identity: InstallIdentity? = nil,
        activeProjectDir: String? = nil,
        catalogUsers: [InstallIdentity] = [
            InstallIdentity(userID: "00000000-0000-7000-8000-000000000001", displayName: "Jane Smith")
        ]
    ) {
        self.identity = identity
        self.activeProjectDir = activeProjectDir
        self.catalogUsers = catalogUsers
    }

    func ping(_ message: String) async throws -> String {
        message
    }

    func version() async throws -> String {
        "0.0.0"
    }

    func installIdentity(identityDir _: String) async throws -> InstallIdentity? {
        identity
    }

    func completeOnboarding(
        identityDir _: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        identity = InstallIdentity(userID: lastResult.userID, displayName: displayName)
        let result = OnboardingResult(
            projectDir: parentDir + "/" + familyName + ".provenance",
            userID: lastResult.userID,
            displayName: displayName
        )
        activeProjectDir = result.projectDir
        catalogUsers = [InstallIdentity(userID: lastResult.userID, displayName: displayName)]
        return result
    }

    func removeInstallIdentity(identityDir _: String) async throws {
        identity = nil
    }

    func activeProject(identityDir _: String) async throws -> String? {
        activeProjectDir
    }

    func listProjectUsers(projectDir _: String) async throws -> [InstallIdentity] {
        catalogUsers
    }

    func openProject(
        identityDir _: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult {
        if !adoptUserID.isEmpty, let match = catalogUsers.first(where: { $0.userID == adoptUserID }) {
            identity = match
            activeProjectDir = projectDir
            return OnboardingResult(projectDir: projectDir, userID: match.userID, displayName: match.displayName)
        }
        if identity == nil {
            identity = InstallIdentity(userID: lastResult.userID, displayName: displayName)
        }
        let name = identity?.displayName ?? displayName
        let result = OnboardingResult(
            projectDir: projectDir,
            userID: identity?.userID ?? lastResult.userID,
            displayName: name
        )
        activeProjectDir = projectDir
        return result
    }

    func removeActiveProject(identityDir _: String) async throws {
        activeProjectDir = nil
    }
}
