import Foundation

/// The Source's files the Evidence graph header can open. Pure rules: which
/// Artifacts have a file, in what order, which one is selected first, and how
/// a file's kind is named.
enum EvidenceGraphFiles {
    struct Row: Equatable {
        var id: String
        var ref: String
        var title: String
        var titleIsRef: Bool
        var filename: String
        var relPath: String
        var mediaType: String
        var byteSize: Int64
        var thumbnailRelPath: String
        var isCover: Bool

        /// `CatalogArtifact.fileDisplayName` for this row.
        var displayFilename: String {
            CatalogArtifact.fileDisplayName(originalFilename: filename, relPath: relPath, ref: ref)
        }
    }

    /// Artifacts that have an openable file, in workspace order.
    static func fileBearing(_ artifacts: [CatalogArtifact], coverArtifactID: String?) -> [Row] {
        artifacts.compactMap { artifact in
            guard let relPath = artifact.openableRelPath, let file = artifact.file else { return nil }
            return Row(
                id: artifact.id,
                ref: artifact.ref,
                title: artifact.displayTitle,
                titleIsRef: !artifact.hasDisplayLabel,
                filename: file.originalFilename,
                relPath: relPath,
                mediaType: file.mediaType,
                byteSize: file.byteSize,
                thumbnailRelPath: artifact.thumbnailRelPath,
                isCover: artifact.id == coverArtifactID
            )
        }
    }

    /// The cover when it is listed, otherwise the first row.
    static func initialID(in rows: [Row], coverArtifactID: String?) -> String? {
        if let coverArtifactID, rows.contains(where: { $0.id == coverArtifactID }) {
            return coverArtifactID
        }
        return rows.first?.id
    }

    /// Short kind word for a file's media type. Format names are verbatim
    /// tokens. Other families reuse the citation composer's words.
    static func kindToken(mediaType: String) -> String {
        switch mediaType {
        case "image/jpeg", "image/jpg": return "JPEG"
        case "image/png": return "PNG"
        case "image/tiff", "image/tif": return "TIFF"
        case "image/gif": return "GIF"
        case "image/webp": return "WEBP"
        case "application/pdf": return L10n.string(L10n.CitationComposer.artifactKindPDF)
        default:
            if mediaType.hasPrefix("image/") {
                return L10n.string(L10n.CitationComposer.artifactKindImage)
            }
            return L10n.string(L10n.CitationComposer.artifactKindUnknown)
        }
    }
}
