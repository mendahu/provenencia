import SwiftUI

/// Curated Marks pack — the one place the Mac client draws its own artwork
/// instead of SF Symbols. Three families: `file_*` (MIME stand-ins),
/// `type_*` (Source types on `source_types.icon_key`), and `subject_*`
/// (evidence-graph / Subject-fields kinds, including `subject_source`).
///
/// Artwork lives in `Assets.xcassets/Marks` as template-rendered vector
/// image sets named `provenencia_<key>`. The mono type label on file glyphs is
/// composited here (Xcode's SVG importer does not resolve `<text>`).
/// Tint comes from call-site `.foregroundStyle` / ambient foreground — never
/// baked into assets. `type_*` names a held Source type; `subject_source` names
/// the reified source kind on the graph — different families, never swapped.

enum PVMarkFamily {
    case file
    case type
    case subject
}

enum PVMarkKey: String, CaseIterable, Sendable {
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
    case typePassport = "type_passport"

    // Subject kinds (graph + Subject fields)
    case subjectPerson = "subject_person"
    case subjectEvent = "subject_event"
    case subjectPlace = "subject_place"
    case subjectRelationship = "subject_relationship"
    case subjectParticipation = "subject_participation"
    case subjectLocation = "subject_location"
    case subjectSource = "subject_source"

    /// Unknown catalog strings resolve here rather than failing.
    static let fallback: PVMarkKey = .typeEvidence

    /// Default for a new user Source type until the researcher picks another.
    static let defaultTypeMark: PVMarkKey = .typeEvidence

    init(catalogKey: String?) {
        self = PVMarkKey(rawValue: catalogKey ?? "") ?? .fallback
    }

    var family: PVMarkFamily {
        if rawValue.hasPrefix("file_") { return .file }
        if rawValue.hasPrefix("subject_") { return .subject }
        return .type
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
        case .filePDF: L10n.DesignSystem.markFilePDF
        case .fileDoc: L10n.DesignSystem.markFileDoc
        case .fileTxt: L10n.DesignSystem.markFileTxt
        case .fileSheet: L10n.DesignSystem.markFileSheet
        case .fileSlides: L10n.DesignSystem.markFileSlides
        case .fileVideo: L10n.DesignSystem.markFileVideo
        case .fileAudio: L10n.DesignSystem.markFileAudio
        case .fileImageMissing: L10n.DesignSystem.markFileImageMissing
        case .fileGeneric: L10n.DesignSystem.markFileGeneric
        case .typeCertificate: L10n.DesignSystem.markTypeCertificate
        case .typeBook: L10n.DesignSystem.markTypeBook
        case .typeDocument: L10n.DesignSystem.markTypeDocument
        case .typeScroll: L10n.DesignSystem.markTypeScroll
        case .typePhotograph: L10n.DesignSystem.markTypePhotograph
        case .typeNewspaper: L10n.DesignSystem.markTypeNewspaper
        case .typeMap: L10n.DesignSystem.markTypeMap
        case .typeMicrofilm: L10n.DesignSystem.markTypeMicrofilm
        case .typeCassette: L10n.DesignSystem.markTypeCassette
        case .typeOralHistory: L10n.DesignSystem.markTypeOralHistory
        case .typeVideo: L10n.DesignSystem.markTypeVideo
        case .typeWebsite: L10n.DesignSystem.markTypeWebsite
        case .typeCensus: L10n.DesignSystem.markTypeCensus
        case .typeDNA: L10n.DesignSystem.markTypeDNA
        case .typeGEDCOM: L10n.DesignSystem.markTypeGEDCOM
        case .typeGrave: L10n.DesignSystem.markTypeGrave
        case .typeScrapbook: L10n.DesignSystem.markTypeScrapbook
        case .typeEvidence: L10n.DesignSystem.markTypeEvidence
        case .typeFolderArchive: L10n.DesignSystem.markTypeFolderArchive
        case .typeEmail: L10n.DesignSystem.markTypeEmail
        case .typePostcard: L10n.DesignSystem.markTypePostcard
        case .typePassport: L10n.DesignSystem.markTypePassport
        case .subjectPerson: L10n.DesignSystem.markSubjectPerson
        case .subjectEvent: L10n.DesignSystem.markSubjectEvent
        case .subjectPlace: L10n.DesignSystem.markSubjectPlace
        case .subjectRelationship: L10n.DesignSystem.markSubjectRelationship
        case .subjectParticipation: L10n.DesignSystem.markSubjectParticipation
        case .subjectLocation: L10n.DesignSystem.markSubjectLocation
        case .subjectSource: L10n.DesignSystem.markSubjectSource
        }
    }

    /// Short title in the Source-type icon picker grid and form tile.
    var typePickerTitle: LocalizedStringResource {
        switch self {
        case .typeCertificate: L10n.DesignSystem.markTypeCertificateTitle
        case .typeBook: L10n.DesignSystem.markTypeBookTitle
        case .typeDocument: L10n.DesignSystem.markTypeDocumentTitle
        case .typeScroll: L10n.DesignSystem.markTypeScrollTitle
        case .typePhotograph: L10n.DesignSystem.markTypePhotographTitle
        case .typeNewspaper: L10n.DesignSystem.markTypeNewspaperTitle
        case .typeMap: L10n.DesignSystem.markTypeMapTitle
        case .typeMicrofilm: L10n.DesignSystem.markTypeMicrofilmTitle
        case .typeCassette: L10n.DesignSystem.markTypeCassetteTitle
        case .typeOralHistory: L10n.DesignSystem.markTypeOralHistoryTitle
        case .typeVideo: L10n.DesignSystem.markTypeVideoTitle
        case .typeWebsite: L10n.DesignSystem.markTypeWebsiteTitle
        case .typeCensus: L10n.DesignSystem.markTypeCensusTitle
        case .typeDNA: L10n.DesignSystem.markTypeDNATitle
        case .typeGEDCOM: L10n.DesignSystem.markTypeGEDCOMTitle
        case .typeGrave: L10n.DesignSystem.markTypeGraveTitle
        case .typeScrapbook: L10n.DesignSystem.markTypeScrapbookTitle
        case .typeEvidence: L10n.DesignSystem.markTypeEvidenceTitle
        case .typeFolderArchive: L10n.DesignSystem.markTypeFolderArchiveTitle
        case .typeEmail: L10n.DesignSystem.markTypeEmailTitle
        case .typePostcard: L10n.DesignSystem.markTypePostcardTitle
        case .typePassport: L10n.DesignSystem.markTypePassportTitle
        default: accessibilityName
        }
    }

    /// One-line metaphor shown under the icon picker dialog while this mark is selected.
    var typeMetaphor: LocalizedStringResource {
        switch self {
        case .typeCertificate: L10n.DesignSystem.markTypeCertificateMetaphor
        case .typeBook: L10n.DesignSystem.markTypeBookMetaphor
        case .typeDocument: L10n.DesignSystem.markTypeDocumentMetaphor
        case .typeScroll: L10n.DesignSystem.markTypeScrollMetaphor
        case .typePhotograph: L10n.DesignSystem.markTypePhotographMetaphor
        case .typeNewspaper: L10n.DesignSystem.markTypeNewspaperMetaphor
        case .typeMap: L10n.DesignSystem.markTypeMapMetaphor
        case .typeMicrofilm: L10n.DesignSystem.markTypeMicrofilmMetaphor
        case .typeCassette: L10n.DesignSystem.markTypeCassetteMetaphor
        case .typeOralHistory: L10n.DesignSystem.markTypeOralHistoryMetaphor
        case .typeVideo: L10n.DesignSystem.markTypeVideoMetaphor
        case .typeWebsite: L10n.DesignSystem.markTypeWebsiteMetaphor
        case .typeCensus: L10n.DesignSystem.markTypeCensusMetaphor
        case .typeDNA: L10n.DesignSystem.markTypeDNAMetaphor
        case .typeGEDCOM: L10n.DesignSystem.markTypeGEDCOMMetaphor
        case .typeGrave: L10n.DesignSystem.markTypeGraveMetaphor
        case .typeScrapbook: L10n.DesignSystem.markTypeScrapbookMetaphor
        case .typeEvidence: L10n.DesignSystem.markTypeEvidenceMetaphor
        case .typeFolderArchive: L10n.DesignSystem.markTypeFolderArchiveMetaphor
        case .typeEmail: L10n.DesignSystem.markTypeEmailMetaphor
        case .typePostcard: L10n.DesignSystem.markTypePostcardMetaphor
        case .typePassport: L10n.DesignSystem.markTypePassportMetaphor
        default: L10n.DesignSystem.markTypeEvidenceMetaphor
        }
    }

    static var fileKeys: [PVMarkKey] { allCases.filter { $0.family == .file } }
    static var typeKeys: [PVMarkKey] { allCases.filter { $0.family == .type } }
    static var subjectKeys: [PVMarkKey] { allCases.filter { $0.family == .subject } }
}

enum PVMarkSize: CGFloat {
    case inline = 16
    case row = 24
    case tile = 40
}

struct PVMark: View {
    private let key: PVMarkKey
    private let size: CGFloat
    private let showLabel: Bool?
    private let decorative: Bool

    init(
        _ key: PVMarkKey,
        size: CGFloat = PVMarkSize.tile.rawValue,
        showLabel: Bool? = nil,
        decorative: Bool = false
    ) {
        self.key = key
        self.size = size
        self.showLabel = showLabel
        self.decorative = decorative
    }

    init(
        _ key: PVMarkKey,
        size: PVMarkSize,
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
            .accessibilityLabel(Text(key.accessibilityName))
    }

    private var band: some View {
        Rectangle()
            .frame(width: bandWidth, height: bandHeight)
            .overlay {
                if let label = key.label {
                    Text(verbatim: label)
                        .font(PVFont.mono(size: labelSize, weight: PVFontWeight.semibold))
                        .tracking(size * 0.28 / 24)
                        .blendMode(.destinationOut)
                }
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
    ) -> PVMarkKey {
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

    private static func fromMediaType(_ raw: String?) -> PVMarkKey? {
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
        case "text/markdown", "text/x-markdown":
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

    private static func fromFilename(_ raw: String?) -> PVMarkKey? {
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

#Preview("Marks") {
    ScrollView {
        VStack(alignment: .leading, spacing: PVSpacing.space7) {
            ForEach([PVMarkKey.fileKeys, PVMarkKey.typeKeys, PVMarkKey.subjectKeys], id: \.first) { group in
                LazyVGrid(columns: Array(repeating: GridItem(.flexible()), count: 6), spacing: PVSpacing.space6) {
                    ForEach(group, id: \.self) { key in
                        PVMark(key, size: .tile)
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
