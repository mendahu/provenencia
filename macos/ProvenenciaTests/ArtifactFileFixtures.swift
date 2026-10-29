import AppKit
import CoreGraphics
import Foundation
import PDFKit
import UniformTypeIdentifiers
@testable import Provenencia

enum ArtifactFileFixtureError: Error {
    case failed
}

/// Artifact files on disk for tests: an image (PNG, JPEG, TIFF), a PDF, raw
/// bytes, and the catalog row that points at them.
enum ArtifactFileFixtures {
    static func uniqueProjectDir() -> String {
        FileManager.default.temporaryDirectory
            .appendingPathComponent("pv-artifact-\(UUID().uuidString).provenencia", isDirectory: true)
            .path
    }

    /// An all-red PNG, or with `topHalfRed` a red top half over a blue bottom half.
    static func writePNG(
        projectDir: String,
        relPath: String,
        width: Int = 8,
        height: Int = 8,
        topHalfRed: Bool = false
    ) throws {
        try writeImage(
            projectDir: projectDir, relPath: relPath,
            width: width, height: height, fileType: .png, topHalfRed: topHalfRed
        )
    }

    /// An all-red image of `fileType`, or with `topHalfRed` a red top half over a blue bottom half.
    /// `dpi` and `orientation` (EXIF 1...8) are written into the file when given;
    /// `grayscale` writes one gray channel, so a large JPEG stays small in memory.
    /// `contentType` names a type `NSBitmapImageRep.FileType` has no case for (HEIC) and wins over `fileType`.
    static func writeImage(
        projectDir: String,
        relPath: String,
        width: Int = 8,
        height: Int = 8,
        fileType: NSBitmapImageRep.FileType = .png,
        topHalfRed: Bool = false,
        dpi: CGFloat? = nil,
        orientation: Int? = nil,
        grayscale: Bool = false,
        contentType: UTType? = nil
    ) throws {
        // A bitmap context, not NSImage.lockFocus: that draws at the display's
        // backing scale and would double the pixel size on a Retina Mac.
        guard let context = CGContext(
            data: nil,
            width: width,
            height: height,
            bitsPerComponent: 8,
            bytesPerRow: 0,
            space: grayscale ? CGColorSpaceCreateDeviceGray() : CGColorSpaceCreateDeviceRGB(),
            bitmapInfo: grayscale ? CGImageAlphaInfo.none.rawValue : CGImageAlphaInfo.premultipliedLast.rawValue
        ) else { throw ArtifactFileFixtureError.failed }
        context.setFillColor(red: 1, green: 0, blue: 0, alpha: 1)
        context.fill(CGRect(x: 0, y: 0, width: width, height: height))
        if topHalfRed {
            // CGContext's origin is bottom-left: the blue lower half is the bottom of the picture.
            context.setFillColor(red: 0, green: 0, blue: 1, alpha: 1)
            context.fill(CGRect(x: 0, y: 0, width: width, height: height / 2))
        }
        let type: UTType
        switch fileType {
        case .jpeg: type = .jpeg
        case .tiff: type = .tiff
        case .bmp: type = .bmp
        case .gif: type = .gif
        default: type = .png
        }
        let writtenType = contentType ?? type
        var properties: [CFString: Any] = [:]
        if let dpi {
            properties[kCGImagePropertyDPIWidth] = dpi
            properties[kCGImagePropertyDPIHeight] = dpi
        }
        if let orientation { properties[kCGImagePropertyOrientation] = orientation }
        let data = NSMutableData()
        guard let cgImage = context.makeImage(),
              let destination = CGImageDestinationCreateWithData(data, writtenType.identifier as CFString, 1, nil)
        else { throw ArtifactFileFixtureError.failed }
        CGImageDestinationAddImage(destination, cgImage, properties as CFDictionary)
        guard CGImageDestinationFinalize(destination) else { throw ArtifactFileFixtureError.failed }
        try write(data as Data, projectDir: projectDir, relPath: relPath)
    }

    /// A PNG that declares `width` x `height` and holds no real pixels: its IDAT is filler.
    /// ImageIO reads the declared size from it and nothing decodes. ImageIO refuses a header
    /// whose data is far too small for the size it declares, so the filler is one byte in
    /// 3,000 pixels.
    static func declaredPNGHeaderData(width: Int, height: Int) -> Data {
        func be32(_ value: UInt32) -> [UInt8] {
            [UInt8(value >> 24), UInt8((value >> 16) & 0xff), UInt8((value >> 8) & 0xff), UInt8(value & 0xff)]
        }
        func chunk(_ type: String, _ body: [UInt8]) -> [UInt8] {
            let typed = Array(type.utf8) + body
            return be32(UInt32(body.count)) + typed + be32(crc32(typed))
        }
        let ihdr = be32(UInt32(width)) + be32(UInt32(height)) + [8, 2, 0, 0, 0]
        let filler = [UInt8(0x78), 0x9c] + [UInt8](repeating: 0, count: max(1, width * height / 3000))
        let signature: [UInt8] = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]
        return Data(signature + chunk("IHDR", ihdr) + chunk("IDAT", filler) + chunk("IEND", []))
    }

    static func writeDeclaredPNGHeader(projectDir: String, relPath: String, width: Int, height: Int) throws {
        try write(declaredPNGHeaderData(width: width, height: height), projectDir: projectDir, relPath: relPath)
    }

    private static func crc32(_ bytes: [UInt8]) -> UInt32 {
        var crc: UInt32 = 0xffff_ffff
        for byte in bytes {
            crc ^= UInt32(byte)
            for _ in 0..<8 {
                crc = (crc & 1) != 0 ? (crc >> 1) ^ 0xedb8_8320 : crc >> 1
            }
        }
        return ~crc
    }

    static func writeBlankPDF(projectDir: String, relPath: String, pageCount: Int) throws {
        let document = PDFDocument()
        for _ in 0..<pageCount {
            document.insert(PDFPage(), at: document.pageCount)
        }
        guard let data = document.dataRepresentation() else {
            throw ArtifactFileFixtureError.failed
        }
        try write(data, projectDir: projectDir, relPath: relPath)
    }

    /// One 100×200 pt page with a black 10×10 pt square at its unrotated
    /// top-left, shown at `rotation` degrees.
    static func writeMarkedPDF(projectDir: String, relPath: String, rotation: Int) throws {
        let data = NSMutableData()
        var box = CGRect(x: 0, y: 0, width: 100, height: 200)
        guard let consumer = CGDataConsumer(data: data as CFMutableData),
              let context = CGContext(consumer: consumer, mediaBox: &box, nil)
        else { throw ArtifactFileFixtureError.failed }
        context.beginPDFPage(nil)
        context.setFillColor(red: 1, green: 1, blue: 1, alpha: 1)
        context.fill(box)
        context.setFillColor(red: 0, green: 0, blue: 0, alpha: 1)
        context.fill(CGRect(x: 0, y: 190, width: 10, height: 10))
        context.endPDFPage()
        context.closePDF()
        guard let document = PDFDocument(data: data as Data), let page = document.page(at: 0) else {
            throw ArtifactFileFixtureError.failed
        }
        page.rotation = rotation
        guard let rotated = document.dataRepresentation() else { throw ArtifactFileFixtureError.failed }
        try write(rotated, projectDir: projectDir, relPath: relPath)
    }

    /// Makes a symlink at `linkPath` to a PNG in a fresh outside folder, creating
    /// the link's parent folders. Returns the outside folder; the caller removes it.
    static func linkToOutsidePNG(at linkPath: String) throws -> String {
        let outside = uniqueProjectDir()
        try writePNG(projectDir: outside, relPath: "outside.png")
        let link = URL(fileURLWithPath: linkPath)
        try FileManager.default.createDirectory(
            at: link.deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        try FileManager.default.createSymbolicLink(
            atPath: linkPath,
            withDestinationPath: outside + "/outside.png"
        )
        return outside
    }

    /// Writes a Finder alias (bookmark file) at `aliasPath` to a PNG in a fresh
    /// outside folder, creating the alias's parent folders. Returns the outside
    /// folder; the caller removes it.
    static func aliasToOutsidePNG(at aliasPath: String) throws -> String {
        let outside = uniqueProjectDir()
        try writePNG(projectDir: outside, relPath: "outside.png")
        let alias = URL(fileURLWithPath: aliasPath)
        try FileManager.default.createDirectory(
            at: alias.deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        let data = try URL(fileURLWithPath: outside + "/outside.png").bookmarkData(
            options: .suitableForBookmarkFile,
            includingResourceValuesForKeys: nil,
            relativeTo: nil
        )
        try URL.writeBookmarkData(data, to: alias)
        return outside
    }

    /// Makes a symlink at `linkPath` to a fresh outside folder holding a PNG at
    /// `relPath`, creating the link's parent folders. Returns the outside folder;
    /// the caller removes it.
    static func linkToOutsideFolder(at linkPath: String, holding relPath: String) throws -> String {
        let outside = uniqueProjectDir()
        try writePNG(projectDir: outside, relPath: relPath)
        try FileManager.default.createDirectory(
            at: URL(fileURLWithPath: linkPath).deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        try FileManager.default.createSymbolicLink(atPath: linkPath, withDestinationPath: outside)
        return outside
    }

    static func seedArtifact(_ store: FakeStore, sourceID: String = "src-1", mediaType: String, relPath: String) {
        store.artifactsBySource[sourceID] = [
            CatalogArtifact(
                id: "art-0",
                ref: "ART-0",
                sourceID: sourceID,
                fileID: "file-0",
                label: "Scan 1",
                description: "",
                file: CatalogFileRef(
                    id: "file-0",
                    relPath: relPath,
                    originalFilename: URL(fileURLWithPath: relPath).lastPathComponent,
                    mediaType: mediaType,
                    byteSize: 10,
                    typeIdentifiers: FakeStore.typeIdentifiers(forMediaType: mediaType)
                )
            ),
        ]
    }

    /// The image as flat RGBA bytes (premultiplied, row 0 at the top), read
    /// through a redraw into a known layout so any source layout works. Empty
    /// when the bitmap cannot be made.
    static func rgbaBytes(_ image: CGImage) -> [UInt8] {
        let width = image.width
        let height = image.height
        var data = [UInt8](repeating: 0, count: width * height * 4)
        let drawn = data.withUnsafeMutableBytes { buffer -> Bool in
            guard let context = CGContext(
                data: buffer.baseAddress,
                width: width,
                height: height,
                bitsPerComponent: 8,
                bytesPerRow: width * 4,
                space: CGColorSpaceCreateDeviceRGB(),
                bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
            ) else { return false }
            context.draw(image, in: CGRect(x: 0, y: 0, width: width, height: height))
            return true
        }
        return drawn ? data : []
    }

    /// Writes `data` at `relPath` under `projectDir` and sets its permission bits to
    /// `mode`, so a test can make an object with an execute bit.
    static func write(_ data: Data, projectDir: String, relPath: String, mode: mode_t = 0o644) throws {
        let url = URL(fileURLWithPath: projectDir).appendingPathComponent(relPath)
        try FileManager.default.createDirectory(
            at: url.deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        try data.write(to: url)
        guard chmod(url.path, mode) == 0 else { throw ArtifactFileFixtureError.failed }
    }
}
