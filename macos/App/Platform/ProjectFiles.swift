import AppKit
import Foundation
import UniformTypeIdentifiers

/// Resolves content-addressed object paths under a `.provenencia` project and
/// opens user-selected files for ingest. Swift never writes `objects/` itself.
enum ProjectFiles {
    /// The project folder that holds every object; mirrors `objectpath.Dir`.
    private static let objectsDir = "objects"

    /// Joins `projectDir` with a slash-separated `relPath` from the catalog
    /// (e.g. `objects/ab/cd/abcd…`). Returns `nil` when `relPath` is empty,
    /// is not under `objects/`, contains `..` / absolute-ish segments, resolves
    /// outside `projectDir`, names a missing file, passes through a symlink at
    /// or under `objects/`, is a Finder alias, or is not a regular file.
    static func objectURL(projectDir: String, relPath: String) -> URL? {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }

        let projectURL = URL(fileURLWithPath: projectDir, isDirectory: true).standardizedFileURL
        var url = projectURL
        var segments: [String] = []
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
            segments.append(segment)
        }
        guard segments.count >= 2, segments[0] == objectsDir else { return nil }

        let resolved = url.standardizedFileURL
        let projectPath = projectURL.path
        let resolvedPath = resolved.path
        // The segment rules above already guarantee a real child of the project
        // directory; this stays as defence in depth in case they change.
        guard resolvedPath.hasPrefix(projectPath + "/") else {
            return nil
        }
        guard isSafeObject(segments: segments, projectURL: projectURL) else { return nil }
        return resolved
    }

    /// Mirrors core/objectstore openObjectsRoot + openRegular: every folder from
    /// objects/ down is a real folder and the object is a regular file (lstat,
    /// never followed) and not a Finder alias. The project folder path itself
    /// may be a symlink.
    private static func isSafeObject(segments: [String], projectURL: URL) -> Bool {
        var url = projectURL
        for (i, segment) in segments.enumerated() {
            url.appendPathComponent(segment, isDirectory: false)
            var st = stat()
            guard lstat(url.path, &st) == 0 else { return false }
            let kind = st.st_mode & S_IFMT
            let isObject = i == segments.count - 1
            guard kind == (isObject ? S_IFREG : S_IFDIR) else { return false }
            if isObject {
                // A Finder alias is a regular file to lstat, but Launch Services
                // resolves it when the object is opened. Read the key from a
                // freshly built URL so no cached value is trusted; an error
                // refuses the object.
                let fresh = URL(fileURLWithPath: url.path)
                guard let isAlias = try? fresh.resourceValues(forKeys: [.isAliasFileKey]).isAliasFile,
                      !isAlias
                else { return false }
            }
        }
        return true
    }

    /// Opens a stored object in the default app. Returns false when missing
    /// or when `relPath` is rejected by `objectURL`.
    @discardableResult
    static func openObject(projectDir: String, relPath: String) -> Bool {
        guard let url = objectURL(projectDir: projectDir, relPath: relPath) else {
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
