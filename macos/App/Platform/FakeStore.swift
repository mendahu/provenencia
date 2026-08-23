import Foundation

/// In-memory `GenealogyStore` for SwiftUI previews and tests. Not used in the shipped app.
struct FakeStore: GenealogyStore {
    var identity: InstallIdentity?
    var lastResult = OnboardingResult(
        projectDir: "/tmp/Robins Family.provenance",
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jake Robins"
    )

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
        OnboardingResult(
            projectDir: parentDir + "/" + familyName + ".provenance",
            userID: lastResult.userID,
            displayName: displayName
        )
    }
}
