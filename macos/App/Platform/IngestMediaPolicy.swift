import Foundation
import UniformTypeIdentifiers

/// Fail-fast ingest allowlist / reject taxonomy. Keep in sync with
/// `core/ingest/mediatypes` — Go remains authoritative on sniff.
enum IngestMediaPolicy {
    /// Largest file ingest accepts (matches `ingest.MaxBytes`).
    static let maxBytes: Int64 = 512 << 20

    /// Open-panel / drop content types (UX filter only).
    static let allowedContentTypes: [UTType] = {
        var types: [UTType] = [
            .image,
            .pdf,
            .plainText,
            .commaSeparatedText,
            .audio,
            .movie,
            .mpeg4Movie,
            .quickTimeMovie,
            .wav,
            .mp3,
        ]
        if let avi = UTType(filenameExtension: "avi") {
            types.append(avi)
        }
        if let webm = UTType(filenameExtension: "webm") {
            types.append(webm)
        }
        if let flac = UTType(filenameExtension: "flac") {
            types.append(flac)
        }
        if let md = UTType(filenameExtension: "md") {
            types.append(md)
        }
        if let markdown = UTType(filenameExtension: "markdown") {
            types.append(markdown)
        }
        if let doc = UTType(filenameExtension: "doc") {
            types.append(doc)
        }
        if let docx = UTType(filenameExtension: "docx") {
            types.append(docx)
        }
        return types
    }()

    enum Reason: Equatable {
        case office
        case archive
        case executable
        case empty
        case tooLarge
        case unidentified
        case disallowedSniff
        case genericType
        case notAFile
        case symlink
        case missing
        case permission
        case multiFile
        case unreadable
    }

    enum Validation: Equatable {
        case ok
        case reject(Reason)
    }

    /// Validates a picked or dropped file without calling FFI.
    static func validate(url: URL) -> Validation {
        var isDir: ObjCBool = false
        guard FileManager.default.fileExists(atPath: url.path, isDirectory: &isDir) else {
            return .reject(.missing)
        }
        if isDir.boolValue {
            return .reject(.notAFile)
        }

        let values = try? url.resourceValues(forKeys: [
            .isSymbolicLinkKey,
            .fileSizeKey,
            .contentTypeKey,
        ])
        if values?.isSymbolicLink == true {
            return .reject(.symlink)
        }

        let size = Int64(values?.fileSize ?? 0)
        if size == 0 {
            // resourceValues can miss size; fall back to attributes.
            let attrs = try? FileManager.default.attributesOfItem(atPath: url.path)
            let attrSize = (attrs?[.size] as? NSNumber)?.int64Value ?? 0
            if attrSize == 0 {
                return .reject(.empty)
            }
            if attrSize > maxBytes {
                return .reject(.tooLarge)
            }
        } else {
            if size > maxBytes {
                return .reject(.tooLarge)
            }
        }

        if let type = values?.contentType, isAllowed(type) {
            return .ok
        }

        let name = url.lastPathComponent
        switch classifyFilename(name) {
        case .ok:
            return .ok
        case .office:
            return .reject(.office)
        case .archive:
            return .reject(.archive)
        case .executable:
            return .reject(.executable)
        case .generic:
            // UTType may still identify a family when the extension is odd.
            if let type = values?.contentType {
                if (type.conforms(to: .spreadsheet) && !type.conforms(to: .commaSeparatedText))
                    || type.conforms(to: .presentation)
                    || type.conforms(to: .rtf)
                    || type.identifier.contains("excel")
                    || type.identifier.contains("powerpoint")
                    || type.identifier.contains("spreadsheetml")
                    || type.identifier.contains("presentationml")
                {
                    return .reject(.office)
                }
                if type.conforms(to: .archive) || type.conforms(to: .diskImage) {
                    return .reject(.archive)
                }
                if type.conforms(to: .executable) || type.conforms(to: .unixExecutable)
                    || type.conforms(to: .application)
                {
                    return .reject(.executable)
                }
            }
            return .reject(.genericType)
        }
    }

    static func byteSize(of url: URL) -> Int64 {
        let attrs = try? FileManager.default.attributesOfItem(atPath: url.path)
        return (attrs?[.size] as? NSNumber)?.int64Value ?? 0
    }

    static func formatByteSize(_ n: Int64) -> String {
        ByteCountFormatter.string(fromByteCount: n, countStyle: .file)
    }

    // MARK: - Private

    private enum ExtClass {
        case ok, office, archive, executable, generic
    }

    private static let allowedExtensions: Set<String> = [
        "jpg", "jpeg", "png", "gif", "webp", "tif", "tiff", "bmp", "heic", "heif",
        "pdf", "doc", "docx", "txt", "text", "csv", "md", "markdown",
        "mp3", "m4a", "aac", "wav", "ogg", "flac", "aif", "aiff",
        "mp4", "m4v", "mov", "webm", "avi",
    ]

    private static func classifyFilename(_ name: String) -> ExtClass {
        let ext = (name as NSString).pathExtension.lowercased()
        guard !ext.isEmpty else { return .generic }
        if allowedExtensions.contains(ext) { return .ok }
        switch ext {
        case "xls", "xlsx", "ppt", "pptx", "rtf", "odt", "ods", "odp":
            return .office
        case "zip", "rar", "7z", "tar", "gz", "tgz", "bz2", "dmg":
            return .archive
        case "exe", "msi", "bat", "cmd", "com", "scr", "app", "sh", "bin", "dll", "so", "dylib":
            return .executable
        default:
            return .generic
        }
    }

    private static func isAllowed(_ type: UTType) -> Bool {
        if type.conforms(to: .image) { return true }
        if type.conforms(to: .pdf) { return true }
        if type.conforms(to: .plainText) { return true }
        if type.conforms(to: .commaSeparatedText) { return true }
        if type.identifier.contains("markdown") { return true }
        // Word (.doc / .docx); Excel / PowerPoint stay rejected via extension.
        if type.identifier.contains("wordprocessingml")
            || type.identifier.contains("msword")
            || type.identifier == "com.microsoft.word.doc"
            || type.identifier == "org.openxmlformats.wordprocessingml.document"
        {
            return true
        }
        if type.conforms(to: .audio) { return true }
        if type.conforms(to: .movie) || type.conforms(to: .video) { return true }
        return false
    }
}
