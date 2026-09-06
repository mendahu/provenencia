import Foundation

/// In-memory `GenealogyStore` for SwiftUI previews and tests. Not used in the shipped app.
final class FakeStore: GenealogyStore, @unchecked Sendable {
    var identity: InstallIdentity?
    var activeProjectDir: String?
    var catalogUsers: [InstallIdentity]
    var projectInfos: [String: ProjectInfo] = [:]
    var lastResult = OnboardingResult(
        projectDir: "/tmp/robins-family.provenencia",
        userID: "00000000-0000-7000-8000-000000000001",
        displayName: "Jake Robins",
        ref: "USR-F4N2P",
        project: ProjectInfo(
            label: "Robins Family",
            folderName: "robins-family.provenencia",
            createdAt: "2026-01-01T00:00:00Z",
            updatedAt: "2026-01-01T00:00:00Z",
            updatedByUserID: "00000000-0000-7000-8000-000000000001",
            updatedByDisplayName: "Jake Robins",
            updatedByRef: "USR-F4N2P"
        )
    )

    init(
        identity: InstallIdentity? = nil,
        activeProjectDir: String? = nil,
        catalogUsers: [InstallIdentity] = [
            InstallIdentity(
                userID: "00000000-0000-7000-8000-000000000001",
                displayName: "Jane Smith",
                ref: "USR-A1B2C"
            )
        ]
    ) {
        self.identity = identity
        self.activeProjectDir = activeProjectDir
        self.catalogUsers = catalogUsers
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
        let folder = ProjectSlug.folderName(from: familyName)
        let projectDir = parentDir + "/" + folder
        let ref = identity?.ref ?? lastResult.ref
        let userID = identity?.userID ?? lastResult.userID
        identity = InstallIdentity(userID: userID, displayName: displayName, ref: ref)
        let now = ISO8601DateFormatter().string(from: Date())
        let info = ProjectInfo(
            label: familyName,
            folderName: folder,
            createdAt: now,
            updatedAt: now,
            updatedByUserID: userID,
            updatedByDisplayName: displayName,
            updatedByRef: ref
        )
        let result = OnboardingResult(
            projectDir: projectDir,
            userID: userID,
            displayName: displayName,
            ref: ref,
            project: info
        )
        activeProjectDir = result.projectDir
        catalogUsers = [InstallIdentity(userID: userID, displayName: displayName, ref: ref)]
        projectInfos[projectDir] = info
        return result
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
            let info = projectInfos[projectDir] ?? ProjectInfo(
                label: ProjectSlug.labelFromFolder(projectDir),
                folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
                createdAt: "",
                updatedAt: "",
                updatedByUserID: match.userID,
                updatedByDisplayName: match.displayName,
                updatedByRef: match.ref
            )
            return OnboardingResult(
                projectDir: projectDir,
                userID: match.userID,
                displayName: match.displayName,
                ref: match.ref,
                project: info
            )
        }
        if identity == nil {
            identity = InstallIdentity(userID: lastResult.userID, displayName: displayName, ref: lastResult.ref)
        }
        let name = identity?.displayName ?? displayName
        let ref = identity?.ref ?? lastResult.ref
        let userID = identity?.userID ?? lastResult.userID
        let info = projectInfos[projectDir] ?? ProjectInfo(
            label: ProjectSlug.labelFromFolder(projectDir),
            folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
            createdAt: "",
            updatedAt: "",
            updatedByUserID: userID,
            updatedByDisplayName: name,
            updatedByRef: ref
        )
        let result = OnboardingResult(
            projectDir: projectDir,
            userID: userID,
            displayName: name,
            ref: ref,
            project: info
        )
        activeProjectDir = projectDir
        projectInfos[projectDir] = info
        return result
    }

    func removeActiveProject(identityDir _: String) async throws {
        activeProjectDir = nil
    }

    func signOut(identityDir _: String) async throws {
        activeProjectDir = nil
        identity = nil
    }

    func projectInfo(projectDir: String) async throws -> ProjectInfo {
        if let info = projectInfos[projectDir] {
            return info
        }
        return ProjectInfo(
            label: ProjectSlug.labelFromFolder(projectDir),
            folderName: URL(fileURLWithPath: projectDir).lastPathComponent,
            createdAt: "",
            updatedAt: "",
            updatedByUserID: "",
            updatedByDisplayName: "",
            updatedByRef: ""
        )
    }
}
