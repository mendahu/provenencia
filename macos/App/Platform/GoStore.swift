import Foundation

struct GoStore: GenealogyStore {
    func ping(_ message: String) async throws -> String {
        var req = Provenance_Engine_V1_PingRequest()
        req.message = message
        let resp: Provenance_Engine_V1_PingResponse = try await provenanceCall(
            method: CoreMethod.ping,
            request: req
        )
        return resp.message
    }

    func version() async throws -> String {
        let resp: Provenance_Engine_V1_GetVersionResponse = try await provenanceCall(
            method: CoreMethod.getVersion,
            request: Provenance_Engine_V1_GetVersionRequest()
        )
        return resp.version
    }

    func installIdentity(identityDir: String) async throws -> InstallIdentity? {
        var req = Provenance_Engine_V1_GetInstallIdentityRequest()
        req.identityDir = identityDir
        let resp: Provenance_Engine_V1_GetInstallIdentityResponse = try await provenanceCall(
            method: CoreMethod.getInstallIdentity,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return InstallIdentity(userID: resp.userID, displayName: resp.displayName)
    }

    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        var req = Provenance_Engine_V1_CompleteOnboardingRequest()
        req.identityDir = identityDir
        req.parentDir = parentDir
        req.displayName = displayName
        req.familyName = familyName
        let resp: Provenance_Engine_V1_CompleteOnboardingResponse = try await provenanceCall(
            method: CoreMethod.completeOnboarding,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName
        )
    }

    func removeInstallIdentity(identityDir: String) async throws {
        var req = Provenance_Engine_V1_RemoveInstallIdentityRequest()
        req.identityDir = identityDir
        let _: Provenance_Engine_V1_RemoveInstallIdentityResponse = try await provenanceCall(
            method: CoreMethod.removeInstallIdentity,
            request: req
        )
    }
}
