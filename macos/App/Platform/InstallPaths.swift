import Foundation

enum InstallPaths {
    static let appSupportFolder = "Provenance"

    /// `{Application Support}/Provenance`. Holds `identity.json` and `active-project.json`.
    static func identityDirectory(fileManager: FileManager = .default) throws -> URL {
        let base = try fileManager.url(
            for: .applicationSupportDirectory,
            in: .userDomainMask,
            appropriateFor: nil,
            create: true
        )
        return base.appendingPathComponent(appSupportFolder, isDirectory: true)
    }

    static func documentsDirectory(fileManager: FileManager = .default) throws -> URL {
        try fileManager.url(
            for: .documentDirectory,
            in: .userDomainMask,
            appropriateFor: nil,
            create: true
        )
    }

    /// Folder names ending in `.provenance` under Documents. Does not open SQLite.
    /// Call only after the user chooses to open an existing file (macOS Files and Folders TCC).
    static func provenanceProjects(
        in documents: URL,
        fileManager: FileManager = .default
    ) throws -> [URL] {
        var isDir: ObjCBool = false
        guard fileManager.fileExists(atPath: documents.path, isDirectory: &isDir), isDir.boolValue else {
            return []
        }
        let urls = try fileManager.contentsOfDirectory(
            at: documents,
            includingPropertiesForKeys: [.isDirectoryKey],
            options: [.skipsHiddenFiles]
        )
        return urls.filter { url in
            url.pathExtension == "provenance"
                && ((try? url.resourceValues(forKeys: [.isDirectoryKey]).isDirectory) ?? false)
        }
    }

    static func isProjectDirectory(_ path: String, fileManager: FileManager = .default) -> Bool {
        var isDir: ObjCBool = false
        guard fileManager.fileExists(atPath: path, isDirectory: &isDir), isDir.boolValue else {
            return false
        }
        return true
    }
}

struct OnboardingFolders: Sendable {
    var identityDirectory: URL
    var documentsDirectory: URL

    static func liveIdentity() throws -> URL {
        try InstallPaths.identityDirectory()
    }

    static func live() throws -> OnboardingFolders {
        OnboardingFolders(
            identityDirectory: try InstallPaths.identityDirectory(),
            documentsDirectory: try InstallPaths.documentsDirectory()
        )
    }

    static var previewEmpty: OnboardingFolders {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("provenance-preview-empty-\(UUID().uuidString)")
        return OnboardingFolders(identityDirectory: root, documentsDirectory: root)
    }

    static var previewWithProject: OnboardingFolders {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("provenance-preview-proj-\(UUID().uuidString)")
        let docs = root.appendingPathComponent("Documents", isDirectory: true)
        let project = docs.appendingPathComponent("Smith Family.provenance", isDirectory: true)
        try? FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        return OnboardingFolders(identityDirectory: root.appendingPathComponent("identity", isDirectory: true), documentsDirectory: docs)
    }
}
