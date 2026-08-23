import Foundation

struct GoStore: GenealogyStore {
    func ping(_ message: String) async throws -> String {
        try await Task.detached {
            var req = Provenance_Engine_V1_PingRequest()
            req.message = message
            let out = try provenanceInvoke(method: CoreMethod.ping, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_PingResponse(serializedBytes: out)
            return resp.message
        }.value
    }

    func version() async throws -> String {
        try await Task.detached {
            let req = Provenance_Engine_V1_GetVersionRequest()
            let out = try provenanceInvoke(method: CoreMethod.getVersion, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_GetVersionResponse(serializedBytes: out)
            return resp.version
        }.value
    }

    func installIdentity(identityDir: String) async throws -> InstallIdentity? {
        try await Task.detached {
            var req = Provenance_Engine_V1_GetInstallIdentityRequest()
            req.identityDir = identityDir
            let out = try provenanceInvoke(method: CoreMethod.getInstallIdentity, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_GetInstallIdentityResponse(serializedBytes: out)
            guard resp.found else {
                return nil
            }
            return InstallIdentity(userID: resp.userID, displayName: resp.displayName)
        }.value
    }

    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        try await Task.detached {
            var req = Provenance_Engine_V1_CompleteOnboardingRequest()
            req.identityDir = identityDir
            req.parentDir = parentDir
            req.displayName = displayName
            req.familyName = familyName
            let out = try provenanceInvoke(method: CoreMethod.completeOnboarding, request: try req.serializedData())
            let resp = try Provenance_Engine_V1_CompleteOnboardingResponse(serializedBytes: out)
            return OnboardingResult(
                projectDir: resp.projectDir,
                userID: resp.userID,
                displayName: resp.displayName
            )
        }.value
    }
}
