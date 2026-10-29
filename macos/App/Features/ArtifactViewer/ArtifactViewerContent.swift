import AppKit
import Foundation
import PDFKit

/// Loaded document payload. Extend with `.audio` / `.video` without changing `load(source:)`.
///
/// `@unchecked Sendable`: `PDFDocument` and the decoded `NSImage` are not Sendable, but we only
/// hand the value from a one-shot load task onto the main actor.
enum ArtifactViewerContent: Equatable, @unchecked Sendable {
    case image(ArtifactImageDecode.Decoded)
    case pdf(PDFDocument)

    static func == (lhs: ArtifactViewerContent, rhs: ArtifactViewerContent) -> Bool {
        switch (lhs, rhs) {
        case (.image(let a), .image(let b)):
            return a.raster === b.raster
        case (.pdf(let a), .pdf(let b)):
            return a === b || a.documentURL == b.documentURL
        default:
            return false
        }
    }
}

/// Why the viewer is empty (no document paint).
enum ArtifactViewerEmptyReason: Equatable, Sendable {
    case idle
    case unsupportedKind
    case mediaNotYetAvailable
    case missingFile
    case loadFailed
}
