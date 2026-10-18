import AppKit
import CoreGraphics
import Foundation
import PDFKit
@testable import Provenencia

enum ArtifactFileFixtureError: Error {
    case failed
}

/// Artifact files on disk for tests that load a viewer: a PNG, a PDF, and the
/// catalog row that points at them.
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
        // A bitmap context, not NSImage.lockFocus: that draws at the display's
        // backing scale and would double the pixel size on a Retina Mac.
        guard let context = CGContext(
            data: nil,
            width: width,
            height: height,
            bitsPerComponent: 8,
            bytesPerRow: 0,
            space: CGColorSpaceCreateDeviceRGB(),
            bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
        ) else { throw ArtifactFileFixtureError.failed }
        context.setFillColor(red: 1, green: 0, blue: 0, alpha: 1)
        context.fill(CGRect(x: 0, y: 0, width: width, height: height))
        if topHalfRed {
            // CGContext's origin is bottom-left: the blue lower half is the bottom of the picture.
            context.setFillColor(red: 0, green: 0, blue: 1, alpha: 1)
            context.fill(CGRect(x: 0, y: 0, width: width, height: height / 2))
        }
        guard let cgImage = context.makeImage(),
              let data = NSBitmapImageRep(cgImage: cgImage).representation(using: .png, properties: [:])
        else {
            throw ArtifactFileFixtureError.failed
        }
        try write(data, projectDir: projectDir, relPath: relPath)
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
                    byteSize: 10
                )
            ),
        ]
    }

    private static func write(_ data: Data, projectDir: String, relPath: String) throws {
        let url = URL(fileURLWithPath: projectDir).appendingPathComponent(relPath)
        try FileManager.default.createDirectory(
            at: url.deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        try data.write(to: url)
    }
}
