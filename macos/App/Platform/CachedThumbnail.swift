import SwiftUI

/// Loads a project object thumbnail through `ThumbnailCache` and feeds
/// `PVThumbnail` (empty / loading / image / type icon / MIME glyph).
/// Features should not call `NSImage(contentsOf:)` on the render path.
///
/// Preference: raster → fallback mark → type icon → MIME glyph → empty.
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
    /// A mark the caller holds (a Person's `subject_*` mark) to show when there is
    /// no raster; it wins over the type icon and the MIME glyph.
    var fallbackMark: PVMarkKey? = nil
    /// Spoken label of `fallbackMark`.
    var fallbackLabel: LocalizedStringResource? = nil
    /// Spoken label while a raster shows (`.tile` layout); a formatted string, not a resource.
    var imageLabel: String? = nil
    /// Tint and line for the `.tile` thumb (a Person's gender treatment).
    var colors: PVThumbnail.Colors? = nil

    @State private var content: PVThumbnail.Content = .empty

    var body: some View {
        Group {
            switch layout {
            case .tile:
                tileBody
            case .well:
                wellBody
            }
        }
        .task(id: taskID) {
            await load()
        }
    }

    private var tileBody: some View {
        // A row with no path draws its fallback at once, not the empty dash first.
        let hasPath = !relPath.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        let shown = hasPath ? content : fallbackContent()
        return Group {
            if let imageLabel, shown.image != nil {
                PVThumbnail(shown, size: size, cornerRadius: cornerRadius, colors: colors)
                    .accessibilityLabel(Text(verbatim: imageLabel))
            } else {
                PVThumbnail(shown, size: size, cornerRadius: cornerRadius, colors: colors)
            }
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
        "\(projectDir)\0\(relPath)\0\(mediaType)\0\(originalFilename)\0\(typeIconKey ?? "")\0\(fallbackMark?.rawValue ?? "")"
    }

    /// The fallback when there is no raster: `fallbackMark`, then the type icon,
    /// then the MIME glyph, then empty.
    static func fallbackContent(
        fallbackMark: PVMarkKey?,
        fallbackLabel: LocalizedStringResource?,
        typeIconKey: String?,
        mediaType: String,
        originalFilename: String
    ) -> PVThumbnail.Content {
        if let fallbackMark {
            return PVThumbnail.Content(glyph: .mark(fallbackMark), label: fallbackLabel)
        }
        if let raw = typeIconKey?.trimmingCharacters(in: .whitespacesAndNewlines), !raw.isEmpty {
            let key = PVMarkKey(catalogKey: raw)
            if key.family == .type {
                return PVThumbnail.Content(glyph: .mark(key), label: key.accessibilityName)
            }
        }
        let hasFileMetadata = !mediaType.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
            || !originalFilename.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        if hasFileMetadata {
            let key = PVFileTypeGlyph.key(
                mediaType: mediaType.isEmpty ? nil : mediaType,
                originalFilename: originalFilename.isEmpty ? nil : originalFilename
            )
            return PVThumbnail.Content(glyph: .mark(key), label: key.accessibilityName)
        }
        return .empty
    }

    private func fallbackContent() -> PVThumbnail.Content {
        Self.fallbackContent(
            fallbackMark: fallbackMark,
            fallbackLabel: fallbackLabel,
            typeIconKey: typeIconKey,
            mediaType: mediaType,
            originalFilename: originalFilename
        )
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
