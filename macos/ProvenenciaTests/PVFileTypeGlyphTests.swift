import Foundation
import Testing
@testable import Provenencia

@Suite
struct PVFileTypeGlyphTests {
    @Test func pdfMediaType() {
        #expect(PVFileTypeGlyph.key(mediaType: "application/pdf") == .filePDF)
        #expect(PVFileTypeGlyph.key(mediaType: "application/pdf; charset=binary") == .filePDF)
    }

    @Test func officeFamilies() {
        #expect(PVFileTypeGlyph.key(mediaType: "application/msword") == .fileDoc)
        #expect(PVFileTypeGlyph.key(mediaType: "text/plain") == .fileTxt)
        #expect(PVFileTypeGlyph.key(mediaType: "text/csv") == .fileSheet)
        #expect(
            PVFileTypeGlyph.key(
                mediaType: "application/vnd.openxmlformats-officedocument.presentationml.presentation"
            ) == .fileSlides
        )
    }

    @Test func mediaFamilies() {
        #expect(PVFileTypeGlyph.key(mediaType: "video/mp4") == .fileVideo)
        #expect(PVFileTypeGlyph.key(mediaType: "video/quicktime") == .fileVideo)
        #expect(PVFileTypeGlyph.key(mediaType: "audio/wav") == .fileAudio)
        #expect(PVFileTypeGlyph.key(mediaType: "audio/mpeg") == .fileAudio)
    }

    @Test func imageWithoutRasterIsMissing() {
        #expect(PVFileTypeGlyph.key(mediaType: "image/jpeg") == .fileImageMissing)
        #expect(PVFileTypeGlyph.key(mediaType: nil, originalFilename: "scan.png") == .fileImageMissing)
    }

    @Test func filenameFallback() {
        #expect(PVFileTypeGlyph.key(mediaType: nil, originalFilename: "deed.PDF") == .filePDF)
        #expect(PVFileTypeGlyph.key(mediaType: nil, originalFilename: "notes.docx") == .fileDoc)
        #expect(PVFileTypeGlyph.key(mediaType: nil, originalFilename: "clip.mov") == .fileVideo)
    }

    @Test func unknownIsGeneric() {
        #expect(PVFileTypeGlyph.key(mediaType: "application/octet-stream") == .fileGeneric)
        #expect(PVFileTypeGlyph.key(mediaType: nil, originalFilename: "blob.xyz") == .fileGeneric)
        #expect(PVFileTypeGlyph.key(mediaType: nil, originalFilename: nil) == .fileGeneric)
        #expect(PVFileTypeGlyph.key(mediaType: "  ") == .fileGeneric)
    }

    @Test func mediaTypeWinsOverFilename() {
        #expect(
            PVFileTypeGlyph.key(mediaType: "application/pdf", originalFilename: "scan.jpg") == .filePDF
        )
    }
}
