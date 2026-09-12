import SwiftUI

/// Loads a project object thumbnail through `ThumbnailCache` and feeds
/// `PVThumbnail` (empty / loading / image / type icon / MIME glyph).
/// Features should not call `NSImage(contentsOf:)` on the render path.
///
/// Preference: raster → type icon → MIME glyph → empty.
struct CachedThumbnail: View {
    let projectDir: String
    let relPath: String
    var mediaType: String = ""
    var originalFilename: String = ""
    /// `source_types.icon_key` for Source cover / fileless Artifact fallback.
    var typeIconKey: String? = nil
    var size: CGFloat = 44
    var cornerRadius: CGFloat = PVRadius.sm

    @State private var content: PVThumbnail.Content = .empty

    var body: some View {
        PVThumbnail(content, size: size, cornerRadius: cornerRadius)
            .task(id: taskID) {
                await load()
            }
    }

    private var taskID: String {
        "\(projectDir)\0\(relPath)\0\(mediaType)\0\(originalFilename)\0\(typeIconKey ?? "")"
    }

    private var hasFileMetadata: Bool {
        !mediaType.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || !originalFilename.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var resolvedTypeIcon: PVEvidenceIconKey? {
        guard let raw = typeIconKey?.trimmingCharacters(in: .whitespacesAndNewlines), !raw.isEmpty else {
            return nil
        }
        let key = PVEvidenceIconKey(catalogKey: raw)
        guard key.family == .type else { return nil }
        return key
    }

    private func mimeGlyphContent() -> PVThumbnail.Content {
        let key = PVFileTypeGlyph.key(
            mediaType: mediaType.isEmpty ? nil : mediaType,
            originalFilename: originalFilename.isEmpty ? nil : originalFilename
        )
        return PVThumbnail.Content(evidenceIcon: key, label: key.accessibilityName)
    }

    private func typeGlyphContent(_ key: PVEvidenceIconKey) -> PVThumbnail.Content {
        PVThumbnail.Content(evidenceIcon: key, label: key.accessibilityName)
    }

    private func fallbackContent() -> PVThumbnail.Content {
        if let typeKey = resolvedTypeIcon {
            return typeGlyphContent(typeKey)
        }
        if hasFileMetadata {
            return mimeGlyphContent()
        }
        return .empty
    }

    private func load() async {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            content = fallbackContent()
            return
        }
        content = PVThumbnail.Content(loading: true)
        if let nsImage = await ThumbnailCache.shared.image(
            projectDir: projectDir,
            relPath: trimmed
        ) {
            content = PVThumbnail.Content(image: Image(nsImage: nsImage))
        } else {
            content = fallbackContent()
        }
    }
}
