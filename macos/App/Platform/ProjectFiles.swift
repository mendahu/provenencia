import AppKit
import Foundation
import UniformTypeIdentifiers

/// Resolves content-addressed object paths under a `.provenencia` project and
/// opens user-selected files for ingest. Swift never writes `objects/` itself.
enum ProjectFiles {
    /// Joins `projectDir` with a slash-separated `relPath` from the catalog
    /// (e.g. `objects/ab/cd/abcd…`).
    static func objectURL(projectDir: String, relPath: String) -> URL {
        var url = URL(fileURLWithPath: projectDir, isDirectory: true)
        for part in relPath.split(separator: "/") where !part.isEmpty {
            url.appendPathComponent(String(part), isDirectory: false)
        }
        return url
    }

    /// Opens a stored object in the default app. Returns false when missing.
    @discardableResult
    static func openObject(projectDir: String, relPath: String) -> Bool {
        let url = objectURL(projectDir: projectDir, relPath: relPath)
        guard FileManager.default.fileExists(atPath: url.path) else {
            return false
        }
        return NSWorkspace.shared.open(url)
    }

    /// Presents a single-file open panel. Returns an absolute path, or nil if cancelled.
    @MainActor
    static func pickFileForIngest(
        prompt: String,
        message: String
    ) -> String? {
        let panel = NSOpenPanel()
        panel.canChooseFiles = true
        panel.canChooseDirectories = false
        panel.allowsMultipleSelection = false
        panel.allowedContentTypes = IngestMediaPolicy.allowedContentTypes
        panel.prompt = prompt
        panel.message = message
        guard panel.runModal() == .OK, let url = panel.url else {
            return nil
        }
        return url.path
    }
}
