import SwiftUI

/// Loads a project object thumbnail through `ThumbnailCache` and feeds
/// `PVThumbnail` (empty / loading / image / type icon / MIME glyph).
/// Features should not call `NSImage(contentsOf:)` on the render path.
///
/// Preference: raster → type icon → MIME glyph → empty.
struct CachedThumbnail: View {
    /// Square list tile, or a frame-filling well whose raster covers and whose glyph stays 56pt.
    enum Layout {
        case tile
        case well
    }

    /// Glyph size inside a card well. The board draws the type or file mark at 56.
    static let wellGlyph: CGFloat = 56

    let projectDir: String
    let relPath: String
    var mediaType: String = ""
    var originalFilename: String = ""
    /// `source_types.icon_key` for Source cover / fileless Artifact fallback.
    var typeIconKey: String? = nil
    var size: CGFloat = 44
    var cornerRadius: CGFloat = PVRadius.sm
    var layout: Layout = .tile

    @State private var content: PVThumbnail.Content = .empty

    var body: some View {
        Group {
            switch layout {
            case .tile:
                PVThumbnail(content, size: size, cornerRadius: cornerRadius)
            case .well:
                wellBody
            }
        }
        .task(id: taskID) {
            await load()
        }
    }

    /// Fills the parent. A raster cover-crops. A mark or empty state stays centered, without tile chrome.
    private var wellBody: some View {
        ZStack {
            if content.loading {
                ProgressView()
            } else if let image = content.image {
                image
                    .resizable()
                    .scaledToFill()
            } else if case .mark(let key)? = content.glyph {
                PVMark(key, size: Self.wellGlyph, showLabel: false, decorative: true)
                    .foregroundStyle(PVColor.textMuted)
            } else {
                Text(verbatim: "—")
                    .font(PVFont.mono(size: PVTypeScale.caption))
                    .foregroundStyle(PVColor.textFaint)
            }
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .clipped()
        .accessibilityHidden(true)
    }

    private var taskID: String {
        "\(projectDir)\0\(relPath)\0\(mediaType)\0\(originalFilename)\0\(typeIconKey ?? "")"
    }

    private var hasFileMetadata: Bool {
        !mediaType.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || !originalFilename.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var resolvedTypeIcon: PVMarkKey? {
        guard let raw = typeIconKey?.trimmingCharacters(in: .whitespacesAndNewlines), !raw.isEmpty else {
            return nil
        }
        let key = PVMarkKey(catalogKey: raw)
        guard key.family == .type else { return nil }
        return key
    }

    private func mimeGlyphContent() -> PVThumbnail.Content {
        let key = PVFileTypeGlyph.key(
            mediaType: mediaType.isEmpty ? nil : mediaType,
            originalFilename: originalFilename.isEmpty ? nil : originalFilename
        )
        return PVThumbnail.Content(glyph: .mark(key), label: key.accessibilityName)
    }

    private func typeGlyphContent(_ key: PVMarkKey) -> PVThumbnail.Content {
        PVThumbnail.Content(glyph: .mark(key), label: key.accessibilityName)
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
