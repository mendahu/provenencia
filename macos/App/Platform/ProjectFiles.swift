import AppKit
import Foundation
import UniformTypeIdentifiers

/// Resolves content-addressed object paths under a `.provenencia` project and
/// opens user-selected files for ingest. Swift never writes `objects/` itself.
enum ProjectFiles {
    /// The project folder that holds every object; mirrors `objectpath.Dir`.
    private static let objectsDir = "objects"

    /// Why a stored object was not opened.
    enum ObjectRefusal: Error, Equatable {
        /// The object is not there to open: an empty, rejected or missing path, or a
        /// symlink, Finder alias or other non-regular file at it.
        case missing
        /// The object is there but could run a program: it has an execute bit, no
        /// extension, or an executable type.
        case unsafe
    }

    /// Joins `projectDir` with a slash-separated `relPath` from the catalog
    /// (e.g. `objects/ab/cd/abcd…`). Returns `nil` when `relPath` is empty,
    /// is not under `objects/`, contains `..` / absolute-ish segments, resolves
    /// outside `projectDir`, names a missing file, passes through a symlink at
    /// or under `objects/`, is a Finder alias, is not a regular file, or has an
    /// execute bit. It does not refuse by extension: thumbnails and the viewer
    /// read through it.
    static func objectURL(projectDir: String, relPath: String) -> URL? {
        try? checkedObject(projectDir: projectDir, relPath: relPath).get()
    }

    /// The URL `objectURL` would return, or why it refuses. A symlink, alias or
    /// non-regular object stays `.missing`; an execute bit is `.unsafe`.
    private static func checkedObject(projectDir: String, relPath: String) -> Result<URL, ObjectRefusal> {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return .failure(.missing) }

        let projectURL = URL(fileURLWithPath: projectDir, isDirectory: true).standardizedFileURL
        var url = projectURL
        var segments: [String] = []
        for part in trimmed.split(separator: "/") where !part.isEmpty {
            let segment = String(part)
            if segment == ".." || segment == "." {
                return .failure(.missing)
            }
            // Absolute or home-relative segments must not be joined as relatives.
            if segment.hasPrefix("/") || segment.hasPrefix("~") {
                return .failure(.missing)
            }
            url.appendPathComponent(segment, isDirectory: false)
            segments.append(segment)
        }
        guard segments.count >= 2, segments[0] == objectsDir else { return .failure(.missing) }

        let resolved = url.standardizedFileURL
        let projectPath = projectURL.path
        let resolvedPath = resolved.path
        // The segment rules above already guarantee a real child of the project
        // directory; this stays as defence in depth in case they change.
        guard resolvedPath.hasPrefix(projectPath + "/") else {
            return .failure(.missing)
        }
        if let refusal = objectRefusal(segments: segments, projectURL: projectURL) {
            return .failure(refusal)
        }
        return .success(resolved)
    }

    /// Mirrors core/objectstore openObjectsRoot + openRegular: every folder from
    /// objects/ down is a real folder and the object is a regular file (lstat,
    /// never followed), not a Finder alias, with no execute bit (`.unsafe`; the
    /// rest are `.missing`). The project folder path itself may be a symlink.
    private static func objectRefusal(segments: [String], projectURL: URL) -> ObjectRefusal? {
        var url = projectURL
        for (i, segment) in segments.enumerated() {
            url.appendPathComponent(segment, isDirectory: false)
            var st = stat()
            guard lstat(url.path, &st) == 0 else { return .missing }
            let kind = st.st_mode & S_IFMT
            let isObject = i == segments.count - 1
            guard kind == (isObject ? S_IFREG : S_IFDIR) else { return .missing }
            if isObject {
                // A Finder alias is a regular file to lstat, but Launch Services
                // resolves it when the object is opened. Read the key from a
                // freshly built URL so no cached value is trusted; an error
                // refuses the object.
                let fresh = URL(fileURLWithPath: url.path)
                guard let isAlias = try? fresh.resourceValues(forKeys: [.isAliasFileKey]).isAliasFile,
                      !isAlias
                else { return .missing }
                if st.st_mode & 0o111 != 0 { return .unsafe }
            }
        }
        return nil
    }

    /// The URL to hand to Launch Services for a stored object: `checkedObject`'s,
    /// then `.unsafe` when the name has no extension or its type could run a
    /// program (`IngestMediaPolicy.isExecutableType`). Go decides which types are
    /// stored; this refuses only what no stored type should look like.
    static func openableObjectURL(projectDir: String, relPath: String) -> Result<URL, ObjectRefusal> {
        let checked = checkedObject(projectDir: projectDir, relPath: relPath)
        guard case .success(let url) = checked else { return checked }
        let ext = url.pathExtension
        if ext.isEmpty { return .failure(.unsafe) }
        if let type = UTType(filenameExtension: ext), IngestMediaPolicy.isExecutableType(type) {
            return .failure(.unsafe)
        }
        return checked
    }

    /// Opens a stored object in the default app. Returns `nil` when it opened,
    /// else why not: `.unsafe` from `openableObjectURL`, `.missing` when the path
    /// is rejected or Launch Services declines it.
    static func openObject(projectDir: String, relPath: String) -> ObjectRefusal? {
        switch openableObjectURL(projectDir: projectDir, relPath: relPath) {
        case .failure(let refusal):
            return refusal
        case .success(let url):
            return NSWorkspace.shared.open(url) ? nil : .missing
        }
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
