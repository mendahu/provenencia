import AppKit
import CoreGraphics
import Foundation
import PDFKit
import Testing
@testable import Provenencia

@MainActor
struct ArtifactRegionCropTests {
    @Test func cgImageRejectsEmptySize() {
        let image = NSImage(size: .zero)
        #expect(ArtifactRegionCrop.cgImage(from: image) == nil)
    }

    @Test func cgImageConvertsPaintedBitmap() throws {
        let image = try #require(pixelExactImage(width: 8, height: 8, topRed: true))
        let cgImage = try #require(ArtifactRegionCrop.cgImage(from: image))
        #expect(cgImage.width == 8)
        #expect(cgImage.height == 8)
    }

    @Test func cropRejectsZeroAreaPixelRect() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        #expect(ArtifactRegionCrop.crop(image, pixelRect: .zero) == nil)
        #expect(ArtifactRegionCrop.crop(image, pixelRect: CGRect(x: 2, y: 2, width: 0, height: 4)) == nil)
    }

    @Test func cropRejectsZeroAreaNormalizedRect() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        #expect(ArtifactRegionCrop.crop(image, normalizedRect: .zero) == nil)
    }

    @Test func cropNormalizedQuarterIsHalfSize() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        let cropped = ArtifactRegionCrop.crop(
            image,
            normalizedRect: CGRect(x: 0.25, y: 0.25, width: 0.5, height: 0.5)
        )
        #expect(cropped?.width == 4)
        #expect(cropped?.height == 4)
    }

    @Test func cropYDownTopHalfIsRed() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        let cropped = try #require(
            ArtifactRegionCrop.crop(image, normalizedRect: CGRect(x: 0, y: 0, width: 1, height: 0.5))
        )
        #expect(cropped.width == 8)
        #expect(cropped.height == 4)
        let pixel = try #require(firstPixel(cropped))
        #expect(pixel.0 > 200 && pixel.1 < 40 && pixel.2 < 40)
    }

    @Test func imageCropIsTheBoundingBox() async throws {
        let projectDir = ArtifactFileFixtures.uniqueProjectDir()
        try ArtifactFileFixtures.writePNG(
            projectDir: projectDir, relPath: "objects/a.png", width: 8, height: 8, topHalfRed: true
        )
        let viewer = ArtifactViewerModel()
        await viewer.load(source(projectDir, "objects/a.png", "image/png"))
        let middle = try #require(
            ArtifactRegionCrop.render(viewer, page: nil, region: rectangle(0.25, 0.25, 0.75, 0.75))
        )
        #expect(middle.width == 4)
        #expect(middle.height == 4)
        let top = try #require(
            ArtifactRegionCrop.render(viewer, page: nil, region: rectangle(0, 0, 1, 0.25))
        )
        let pixel = try #require(firstPixel(top))
        #expect(pixel.0 > 200 && pixel.1 < 40 && pixel.2 < 40)
    }

    @Test func cropIsCappedAt2048() async throws {
        let projectDir = ArtifactFileFixtures.uniqueProjectDir()
        try ArtifactFileFixtures.writePNG(projectDir: projectDir, relPath: "objects/wide.png", width: 4096, height: 1024)
        let viewer = ArtifactViewerModel()
        await viewer.load(source(projectDir, "objects/wide.png", "image/png"))
        let image = try #require(ArtifactRegionCrop.render(viewer, page: nil, region: rectangle(0, 0, 1, 1)))
        #expect(image.width == 2048)
        #expect(image.height == 512)
    }

    @Test func pdfRegionFollowsTheDisplayedPage() async throws {
        let projectDir = ArtifactFileFixtures.uniqueProjectDir()
        try ArtifactFileFixtures.writeMarkedPDF(projectDir: projectDir, relPath: "objects/m.pdf", rotation: 90)
        let viewer = ArtifactViewerModel()
        await viewer.load(source(projectDir, "objects/m.pdf", "application/pdf"))
        // Rotated 90 degrees the unrotated top-left square sits at the displayed top-right.
        let topRight = try #require(
            ArtifactRegionCrop.render(viewer, page: 1, region: rectangle(0.5, 0, 1, 0.5))
        )
        let topLeft = try #require(
            ArtifactRegionCrop.render(viewer, page: 1, region: rectangle(0, 0, 0.5, 0.5))
        )
        #expect(containsBlack(topRight))
        #expect(!containsBlack(topLeft))
        #expect(topRight.width == 2 * topRight.height)
        #expect(topLeft.width == 2 * topLeft.height)
    }

    @Test func temporaryPNGIsAPNG() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        let url = try ArtifactRegionCrop.writeTemporaryPNG(image)
        defer { try? FileManager.default.removeItem(at: url) }
        #expect(url.deletingLastPathComponent().standardizedFileURL
            == FileManager.default.temporaryDirectory.standardizedFileURL)
        let head = try Data(contentsOf: url).prefix(8)
        #expect(Array(head) == [0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A])
    }

    // MARK: - Helpers

    private func source(_ projectDir: String, _ relPath: String, _ mediaType: String) -> ArtifactViewerSource {
        ArtifactViewerSource(projectDir: projectDir, relPath: relPath, mediaType: mediaType)
    }

    private func rectangle(_ x0: CGFloat, _ y0: CGFloat, _ x1: CGFloat, _ y1: CGFloat) -> ArtifactRegionDraft {
        ArtifactRegionDraft(
            kind: .rectangle,
            points: ArtifactRegionGeometry.rectanglePoints(from: CGPoint(x: x0, y: y0), to: CGPoint(x: x1, y: y1))
        )
    }

    private func pixelExactImage(width: Int, height: Int, topRed: Bool) -> NSImage? {
        guard let cgImage = makeRGBImage(width: width, height: height, topRed: topRed) else { return nil }
        return NSImage(cgImage: cgImage, size: NSSize(width: width, height: height))
    }

    /// Top half red, bottom half blue, in y-down (top-left origin) terms.
    private func makeRGBImage(width: Int, height: Int, topRed: Bool) -> CGImage? {
        let colorSpace = CGColorSpaceCreateDeviceRGB()
        guard let context = CGContext(
            data: nil,
            width: width,
            height: height,
            bitsPerComponent: 8,
            bytesPerRow: 0,
            space: colorSpace,
            bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
        ) else { return nil }
        // CGContext is origin bottom-left: fill bottom then top.
        context.setFillColor(red: 0, green: 0, blue: 1, alpha: 1)
        context.fill(CGRect(x: 0, y: 0, width: width, height: height / 2))
        context.setFillColor(red: topRed ? 1 : 0, green: 0, blue: topRed ? 0 : 1, alpha: 1)
        context.fill(CGRect(x: 0, y: height / 2, width: width, height: height / 2))
        return context.makeImage()
    }

    private func firstPixel(_ image: CGImage) -> (UInt8, UInt8, UInt8)? {
        pixels(image).first
    }

    private func containsBlack(_ image: CGImage) -> Bool {
        pixels(image).contains { $0.0 < 60 && $0.1 < 60 && $0.2 < 60 }
    }

    /// RGB triples read through an RGBA bitmap redraw, so any source layout works.
    private func pixels(_ image: CGImage) -> [(UInt8, UInt8, UInt8)] {
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
        guard drawn else { return [] }
        return stride(from: 0, to: data.count, by: 4).map { (data[$0], data[$0 + 1], data[$0 + 2]) }
    }
}
