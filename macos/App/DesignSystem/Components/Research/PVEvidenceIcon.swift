import SwiftUI

/// Curated evidence icon pack — the one place the Mac client draws its own
/// artwork instead of SF Symbols. Two families: `file_*` (MIME stand-ins) and
/// `type_*` (kinds of evidence on `source_types.icon_key`).
///
/// Artwork lives in `Assets.xcassets/EvidenceIcons` as template-rendered vector
/// image sets named `provenencia_<key>`. The mono type label on file glyphs is
/// composited here (Xcode's SVG importer does not resolve `<text>`).

enum PVEvidenceIconFamily {
    case file
    case type
}

enum PVEvidenceIconKey: String, CaseIterable, Sendable {
    // File-type fallbacks
    case filePDF = "file_pdf"
    case fileDoc = "file_doc"
    case fileTxt = "file_txt"
    case fileSheet = "file_sheet"
    case fileSlides = "file_slides"
    case fileVideo = "file_video"
    case fileAudio = "file_audio"
    case fileImageMissing = "file_image_missing"
    case fileGeneric = "file_generic"

    // Source-type marks
    case typeCertificate = "type_certificate"
    case typeBook = "type_book"
    case typeDocument = "type_document"
    case typeScroll = "type_scroll"
    case typePhotograph = "type_photograph"
    case typeNewspaper = "type_newspaper"
    case typeMap = "type_map"
    case typeMicrofilm = "type_microfilm"
    case typeCassette = "type_cassette"
    case typeOralHistory = "type_oral_history"
    case typeVideo = "type_video"
    case typeWebsite = "type_website"
    case typeCensus = "type_census"
    case typeDNA = "type_dna"
    case typeGEDCOM = "type_gedcom"
    case typeGrave = "type_grave"
    case typeScrapbook = "type_scrapbook"
    case typeEvidence = "type_evidence"
    case typeFolderArchive = "type_folder_archive"
    case typeEmail = "type_email"
    case typePostcard = "type_postcard"

    /// Unknown catalog strings resolve here rather than failing.
    static let fallback: PVEvidenceIconKey = .typeEvidence

    /// Default for a new user Source type until the researcher picks another.
    static let defaultTypeIcon: PVEvidenceIconKey = .typeEvidence

    init(catalogKey: String?) {
        self = PVEvidenceIconKey(rawValue: catalogKey ?? "") ?? .fallback
    }

    var family: PVEvidenceIconFamily {
        rawValue.hasPrefix("file_") ? .file : .type
    }

    var assetName: String { "provenencia_" + rawValue }

    /// Stamped in the ink band. `nil` for source-type marks.
    var label: String? {
        switch self {
        case .filePDF: "PDF"
        case .fileDoc: "DOC"
        case .fileTxt: "TXT"
        case .fileSheet: "CSV"
        case .fileSlides: "PPT"
        case .fileVideo: "MP4"
        case .fileAudio: "WAV"
        case .fileImageMissing: "IMG"
        case .fileGeneric: "FILE"
        default: nil
        }
    }

    var accessibilityName: LocalizedStringResource {
        switch self {
        case .filePDF: L10n.DesignSystem.evidenceIconFilePDF
        case .fileDoc: L10n.DesignSystem.evidenceIconFileDoc
        case .fileTxt: L10n.DesignSystem.evidenceIconFileTxt
        case .fileSheet: L10n.DesignSystem.evidenceIconFileSheet
        case .fileSlides: L10n.DesignSystem.evidenceIconFileSlides
        case .fileVideo: L10n.DesignSystem.evidenceIconFileVideo
        case .fileAudio: L10n.DesignSystem.evidenceIconFileAudio
        case .fileImageMissing: L10n.DesignSystem.evidenceIconFileImageMissing
        case .fileGeneric: L10n.DesignSystem.evidenceIconFileGeneric
        case .typeCertificate: L10n.DesignSystem.evidenceIconTypeCertificate
        case .typeBook: L10n.DesignSystem.evidenceIconTypeBook
        case .typeDocument: L10n.DesignSystem.evidenceIconTypeDocument
        case .typeScroll: L10n.DesignSystem.evidenceIconTypeScroll
        case .typePhotograph: L10n.DesignSystem.evidenceIconTypePhotograph
        case .typeNewspaper: L10n.DesignSystem.evidenceIconTypeNewspaper
        case .typeMap: L10n.DesignSystem.evidenceIconTypeMap
        case .typeMicrofilm: L10n.DesignSystem.evidenceIconTypeMicrofilm
        case .typeCassette: L10n.DesignSystem.evidenceIconTypeCassette
        case .typeOralHistory: L10n.DesignSystem.evidenceIconTypeOralHistory
        case .typeVideo: L10n.DesignSystem.evidenceIconTypeVideo
        case .typeWebsite: L10n.DesignSystem.evidenceIconTypeWebsite
        case .typeCensus: L10n.DesignSystem.evidenceIconTypeCensus
        case .typeDNA: L10n.DesignSystem.evidenceIconTypeDNA
        case .typeGEDCOM: L10n.DesignSystem.evidenceIconTypeGEDCOM
        case .typeGrave: L10n.DesignSystem.evidenceIconTypeGrave
        case .typeScrapbook: L10n.DesignSystem.evidenceIconTypeScrapbook
        case .typeEvidence: L10n.DesignSystem.evidenceIconTypeEvidence
        case .typeFolderArchive: L10n.DesignSystem.evidenceIconTypeFolderArchive
        case .typeEmail: L10n.DesignSystem.evidenceIconTypeEmail
        case .typePostcard: L10n.DesignSystem.evidenceIconTypePostcard
        }
    }

    static var fileKeys: [PVEvidenceIconKey] { allCases.filter { $0.family == .file } }
    static var typeKeys: [PVEvidenceIconKey] { allCases.filter { $0.family == .type } }
}

enum PVEvidenceIconSize: CGFloat {
    case inline = 16
    case row = 24
    case tile = 40
}

struct PVEvidenceIcon: View {
    private let key: PVEvidenceIconKey
    private let size: CGFloat
    private let showLabel: Bool?
    private let decorative: Bool

    init(
        _ key: PVEvidenceIconKey,
        size: CGFloat = PVEvidenceIconSize.tile.rawValue,
        showLabel: Bool? = nil,
        decorative: Bool = false
    ) {
        self.key = key
        self.size = size
        self.showLabel = showLabel
        self.decorative = decorative
    }

    init(
        _ key: PVEvidenceIconKey,
        size: PVEvidenceIconSize,
        decorative: Bool = false
    ) {
        self.init(key, size: size.rawValue, decorative: decorative)
    }

    private static let labelFloor: CGFloat = 22

    private var labelled: Bool {
        guard key.family == .file, key.label != nil else { return false }
        return showLabel ?? (size >= Self.labelFloor)
    }

    private var bandWidth: CGFloat { size * 15.0 / 24 }
    private var bandHeight: CGFloat { size * 6.9 / 24 }
    private var bandOffsetY: CGFloat { size * 6.05 / 24 }
    private var labelSize: CGFloat {
        size * ((key.label?.count ?? 3) > 3 ? 4.4 : 5.4) / 24
    }

    var body: some View {
        Image(key.assetName)
            .renderingMode(.template)
            .resizable()
            .frame(width: size, height: size)
            .overlay { if labelled { band } }
            .accessibilityHidden(decorative)
            .accessibilityLabel(decorative ? Text(verbatim: "") : Text(key.accessibilityName))
    }

    private var band: some View {
        Rectangle()
            .frame(width: bandWidth, height: bandHeight)
            .overlay {
                Text(key.label ?? "")
                    .font(PVFont.mono(size: labelSize, weight: PVFontWeight.semibold))
                    .tracking(size * 0.28 / 24)
                    .blendMode(.destinationOut)
            }
            .compositingGroup()
            .offset(y: bandOffsetY)
    }
}

/// Maps `files.media_type` and/or a filename extension to a `file_*` key.
enum PVFileTypeGlyph {
    /// Prefer MIME; fall back to filename extension. Empty inputs → `file_generic`.
    /// Call only when there is no successful raster thumbnail — image MIME maps
    /// to `file_image_missing`.
    static func key(
        mediaType: String?,
        originalFilename: String? = nil
    ) -> PVEvidenceIconKey {
        if let mimeKey = fromMediaType(mediaType) {
            return mimeKey
        }
        if let extKey = fromFilename(originalFilename) {
            return extKey
        }
        return .fileGeneric
    }

    private static func normalizeMIME(_ raw: String) -> String {
        var s = raw.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
        if let semi = s.firstIndex(of: ";") {
            s = String(s[..<semi]).trimmingCharacters(in: .whitespacesAndNewlines)
        }
        return s
    }

    private static func fromMediaType(_ raw: String?) -> PVEvidenceIconKey? {
        guard let raw, !raw.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            return nil
        }
        let mime = normalizeMIME(raw)
        if mime.hasPrefix("image/") {
            return .fileImageMissing
        }
        switch mime {
        case "application/pdf":
            return .filePDF
        case "application/msword",
             "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
             "application/rtf",
             "text/rtf",
             "application/vnd.oasis.opendocument.text":
            return .fileDoc
        case "text/plain":
            return .fileTxt
        case "text/csv",
             "application/csv",
             "application/vnd.ms-excel",
             "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
             "application/vnd.oasis.opendocument.spreadsheet":
            return .fileSheet
        case "application/vnd.ms-powerpoint",
             "application/vnd.openxmlformats-officedocument.presentationml.presentation",
             "application/vnd.oasis.opendocument.presentation":
            return .fileSlides
        case "video/mp4", "video/quicktime", "video/x-msvideo", "video/webm":
            return .fileVideo
        case "audio/mpeg", "audio/mp3", "audio/wav", "audio/wave", "audio/x-wav", "audio/aac", "audio/ogg":
            return .fileAudio
        case "application/octet-stream":
            return .fileGeneric
        default:
            if mime.hasPrefix("video/") { return .fileVideo }
            if mime.hasPrefix("audio/") { return .fileAudio }
            if mime.hasPrefix("text/") { return .fileTxt }
            return .fileGeneric
        }
    }

    private static func fromFilename(_ raw: String?) -> PVEvidenceIconKey? {
        guard let raw, !raw.isEmpty else { return nil }
        let ext = (raw as NSString).pathExtension.lowercased()
        guard !ext.isEmpty else { return nil }
        switch ext {
        case "pdf":
            return .filePDF
        case "doc", "docx", "rtf", "odt":
            return .fileDoc
        case "txt", "text", "md":
            return .fileTxt
        case "xls", "xlsx", "csv", "ods":
            return .fileSheet
        case "ppt", "pptx", "odp":
            return .fileSlides
        case "mp4", "mov", "m4v", "avi", "webm":
            return .fileVideo
        case "mp3", "wav", "aac", "m4a", "ogg", "flac":
            return .fileAudio
        case "jpg", "jpeg", "png", "gif", "webp", "bmp", "tif", "tiff", "heic":
            return .fileImageMissing
        default:
            return .fileGeneric
        }
    }
}

#Preview("Evidence icons") {
    ScrollView {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            ForEach([PVEvidenceIconKey.fileKeys, PVEvidenceIconKey.typeKeys], id: \.first) { group in
                LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: PVSpacing.space6) {
                    ForEach(group, id: \.self) { key in
                        PVEvidenceIcon(key, size: .tile)
                            .foregroundStyle(PVColor.textSecondary)
                            .frame(width: 52, height: 52)
                            .background(PVColor.surfaceInset)
                    }
                }
            }
        }
        .padding(PVSpacing.space9)
    }
    .background(PVColor.surfacePage)
}
