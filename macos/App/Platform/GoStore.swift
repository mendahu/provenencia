import Foundation

struct GoStore: GenealogyStore {
    func installIdentity(identityDir: String) async throws -> InstallIdentity? {
        var req = Provenencia_Engine_V1_GetInstallIdentityRequest()
        req.identityDir = identityDir
        let resp: Provenencia_Engine_V1_GetInstallIdentityResponse = try await provenenciaCall(
            method: CoreMethod.getInstallIdentity,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return InstallIdentity(userID: resp.userID, displayName: resp.displayName, ref: resp.ref)
    }

    func completeOnboarding(
        identityDir: String,
        parentDir: String,
        displayName: String,
        familyName: String
    ) async throws -> OnboardingResult {
        var req = Provenencia_Engine_V1_CompleteOnboardingRequest()
        req.identityDir = identityDir
        req.parentDir = parentDir
        req.displayName = displayName
        req.familyName = familyName
        let resp: Provenencia_Engine_V1_CompleteOnboardingResponse = try await provenenciaCall(
            method: CoreMethod.completeOnboarding,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName,
            ref: resp.ref,
            project: Self.mapProject(resp.project)
        )
    }

    func activeProject(identityDir: String) async throws -> String? {
        var req = Provenencia_Engine_V1_GetActiveProjectRequest()
        req.identityDir = identityDir
        let resp: Provenencia_Engine_V1_GetActiveProjectResponse = try await provenenciaCall(
            method: CoreMethod.getActiveProject,
            request: req
        )
        guard resp.found else {
            return nil
        }
        return resp.projectDir
    }

    func openProject(
        identityDir: String,
        projectDir: String,
        displayName: String,
        adoptUserID: String
    ) async throws -> OnboardingResult {
        var req = Provenencia_Engine_V1_OpenProjectRequest()
        req.identityDir = identityDir
        req.projectDir = projectDir
        req.displayName = displayName
        req.adoptUserID = adoptUserID
        let resp: Provenencia_Engine_V1_OpenProjectResponse = try await provenenciaCall(
            method: CoreMethod.openProject,
            request: req
        )
        return OnboardingResult(
            projectDir: resp.projectDir,
            userID: resp.userID,
            displayName: resp.displayName,
            ref: resp.ref,
            project: Self.mapProject(resp.project)
        )
    }

    func listProjectUsers(projectDir: String) async throws -> [InstallIdentity] {
        var req = Provenencia_Engine_V1_ListProjectUsersRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_ListProjectUsersResponse = try await provenenciaCall(
            method: CoreMethod.listProjectUsers,
            request: req
        )
        return resp.users.map {
            InstallIdentity(userID: $0.userID, displayName: $0.displayName, ref: $0.ref)
        }
    }

    func removeActiveProject(identityDir: String) async throws {
        var req = Provenencia_Engine_V1_RemoveActiveProjectRequest()
        req.identityDir = identityDir
        let _: Provenencia_Engine_V1_RemoveActiveProjectResponse = try await provenenciaCall(
            method: CoreMethod.removeActiveProject,
            request: req
        )
    }

    func signOut(identityDir: String) async throws {
        var req = Provenencia_Engine_V1_SignOutRequest()
        req.identityDir = identityDir
        let _: Provenencia_Engine_V1_SignOutResponse = try await provenenciaCall(
            method: CoreMethod.signOut,
            request: req
        )
    }

    func projectInfo(projectDir: String) async throws -> ProjectInfo {
        var req = Provenencia_Engine_V1_GetProjectInfoRequest()
        req.projectDir = projectDir
        let resp: Provenencia_Engine_V1_GetProjectInfoResponse = try await provenenciaCall(
            method: CoreMethod.getProjectInfo,
            request: req
        )
        return Self.mapProject(resp.project)
    }

    private static func mapProject(_ p: Provenencia_Engine_V1_ProjectInfo) -> ProjectInfo {
        ProjectInfo(
            label: p.label,
            folderName: p.folderName,
            createdAt: p.createdAt,
            updatedAt: p.updatedAt,
            updatedByUserID: p.updatedByUserID,
            updatedByDisplayName: p.updatedByDisplayName,
            updatedByRef: p.updatedByRef
        )
    }
}
