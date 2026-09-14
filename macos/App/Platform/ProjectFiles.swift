import AppKit
import Foundation
import UniformTypeIdentifiers

/// Resolves content-addressed object paths under a `.provenencia` project and
/// opens user-selected files for ingest. Swift never writes `objects/` itself.
enum ProjectFiles {
    /// Joins `projectDir` with a slash-separated `relPath` from the catalog
    /// (e.g. `objects/ab/cd/abcd…`). Returns `nil` when `relPath` is empty,
    /// contains `..` / absolute-ish segments, or resolves outside `projectDir`.
    static func objectURL(projectDir: String, relPath: String) -> URL? {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let projectURL = URL(fileURLWithPath: projectDir, isDirectory: true).standardizedFileURL
        var url = projectURL
        for part in trimmed.split(separator: "/") where !part.isEmpty {
            let segment = String(part)
            if segment == ".." || segment == "." {
                return nil
            }
            // Absolute or home-relative segments must not be joined as relatives.
            if segment.hasPrefix("/") || segment.hasPrefix("~") {
                return nil
            }
            url.appendPathComponent(segment, isDirectory: false)
        }

        let resolved = url.standardizedFileURL
        let projectPath = projectURL.path
        let resolvedPath = resolved.path
        // Require a real child of the project directory (not the dir itself,
        // and not a sibling that only shares a path prefix).
        guard resolvedPath.hasPrefix(projectPath + "/") else {
            return nil
        }
        return resolved
    }

    /// Opens a stored object in the default app. Returns false when missing
    /// or when `relPath` is rejected by `objectURL`.
    @discardableResult
    static func openObject(projectDir: String, relPath: String) -> Bool {
        guard let url = objectURL(projectDir: projectDir, relPath: relPath) else {
            return false
        }
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
