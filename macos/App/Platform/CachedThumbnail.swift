import SwiftUI

/// Loads a project object thumbnail through `ThumbnailCache` and feeds
/// `PVThumbnail` (empty / loading / image / MIME glyph). Features should not
/// call `NSImage(contentsOf:)` on the render path.
struct CachedThumbnail: View {
    let projectDir: String
    let relPath: String
    var mediaType: String = ""
    var originalFilename: String = ""
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
        "\(projectDir)\0\(relPath)\0\(mediaType)\0\(originalFilename)"
    }

    private var hasFileMetadata: Bool {
        !mediaType.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || !originalFilename.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private func glyphContent() -> PVThumbnail.Content {
        let key = PVFileTypeGlyph.key(
            mediaType: mediaType.isEmpty ? nil : mediaType,
            originalFilename: originalFilename.isEmpty ? nil : originalFilename
        )
        return PVThumbnail.Content(evidenceIcon: key, label: key.accessibilityName)
    }

    private func load() async {
        let trimmed = relPath.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            content = hasFileMetadata ? glyphContent() : .empty
            return
        }
        content = PVThumbnail.Content(loading: true)
        if let nsImage = await ThumbnailCache.shared.image(
            projectDir: projectDir,
            relPath: trimmed
        ) {
            content = PVThumbnail.Content(image: Image(nsImage: nsImage))
        } else if hasFileMetadata {
            content = glyphContent()
        } else {
            content = .empty
        }
    }
}
