import Foundation

enum InstallPaths {
    static let appSupportFolder = "Provenencia"
    /// Filename of the catalog inside a `.provenencia` folder. Existence only; Swift does not open SQLite.
    static let catalogFile = "provenencia.sqlite"

    /// `{Application Support}/Provenencia`. Holds `identity.json` and `active-project.json`.
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

    /// `.provenencia` folders under Documents that contain `provenencia.sqlite` (name + file exists; does not open SQLite).
    /// Call only after the user chooses to open an existing file (macOS Files and Folders TCC).
    static func provenenciaProjects(
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
            url.pathExtension == "provenencia" && isProjectDirectory(url.path, fileManager: fileManager)
        }
    }

    static func isProjectDirectory(_ path: String, fileManager: FileManager = .default) -> Bool {
        var isDir: ObjCBool = false
        guard fileManager.fileExists(atPath: path, isDirectory: &isDir), isDir.boolValue else {
            return false
        }
        let catalog = URL(fileURLWithPath: path).appendingPathComponent(catalogFile)
        var isCatalogDir: ObjCBool = false
        guard fileManager.fileExists(atPath: catalog.path, isDirectory: &isCatalogDir), !isCatalogDir.boolValue else {
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

#if DEBUG
    static func previewEmpty() -> OnboardingFolders {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("provenencia-preview-empty-\(UUID().uuidString)")
        return OnboardingFolders(identityDirectory: root, documentsDirectory: root)
    }

    static func previewWithProject() -> OnboardingFolders {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent("provenencia-preview-proj-\(UUID().uuidString)")
        let docs = root.appendingPathComponent("Documents", isDirectory: true)
        let project = docs.appendingPathComponent("Smith Family.provenencia", isDirectory: true)
        try? FileManager.default.createDirectory(at: project, withIntermediateDirectories: true)
        try? Data().write(to: project.appendingPathComponent(InstallPaths.catalogFile))
        return OnboardingFolders(identityDirectory: root.appendingPathComponent("identity", isDirectory: true), documentsDirectory: docs)
    }
#endif
}
