import Foundation

/// Stable input for ``ArtifactViewerModel/load`` — no catalog types.
struct ArtifactViewerSource: Equatable, Sendable {
    var projectDir: String
    var relPath: String
    var mediaType: String
}

/// MIME → viewer kind. Hosts pass `mediaType` through; they do not branch on strings.
enum ArtifactViewerKind: String, Equatable, Sendable, CaseIterable {
    case image
    case pdf
    case audio
    case video
    case unsupported

    static func classify(mediaType: String) -> ArtifactViewerKind {
        let trimmed = mediaType.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if trimmed.hasPrefix("image/") { return .image }
        if trimmed.contains("pdf") { return .pdf }
        if trimmed.hasPrefix("audio/") { return .audio }
        if trimmed.hasPrefix("video/") { return .video }
        return .unsupported
    }

    /// Page chevrons + page field (board Frame 1).
    var supportsPages: Bool { self == .pdf }

    /// Spatial pan/zoom viewport (board Frames 1–2).
    var supportsSpatialZoom: Bool {
        switch self {
        case .image, .pdf: return true
        case .audio, .video, .unsupported: return false
        }
    }

    /// S7-06 implements image + PDF paint; A/V uses the empty state until players ship.
    var isRenderableInS706: Bool {
        switch self {
        case .image, .pdf: return true
        case .audio, .video, .unsupported: return false
        }
    }
}
