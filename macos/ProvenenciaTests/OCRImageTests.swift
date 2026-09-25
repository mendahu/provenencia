import AppKit
import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

struct OCRImageTests {
    @Test func cgImageRejectsEmptySize() {
        let image = NSImage(size: .zero)
        #expect(OCRImage.cgImage(from: image) == nil)
    }

    @Test func cgImageConvertsPaintedBitmap() throws {
        let image = try #require(pixelExactImage(width: 8, height: 8, topRed: true))
        let cgImage = try #require(OCRImage.cgImage(from: image))
        #expect(cgImage.width == 8)
        #expect(cgImage.height == 8)
    }

    @Test func cropRejectsZeroAreaPixelRect() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        #expect(OCRImage.crop(image, pixelRect: .zero) == nil)
        #expect(OCRImage.crop(image, pixelRect: CGRect(x: 2, y: 2, width: 0, height: 4)) == nil)
    }

    @Test func cropRejectsZeroAreaNormalizedRect() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        #expect(OCRImage.crop(image, normalizedRect: .zero) == nil)
    }

    @Test func cropNormalizedQuarterIsHalfSize() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        let cropped = OCRImage.crop(
            image,
            normalizedRect: CGRect(x: 0.25, y: 0.25, width: 0.5, height: 0.5)
        )
        #expect(cropped?.width == 4)
        #expect(cropped?.height == 4)
    }

    @Test func cropYDownTopHalfIsRed() throws {
        let image = try #require(makeRGBImage(width: 8, height: 8, topRed: true))
        let cropped = try #require(
            OCRImage.crop(image, normalizedRect: CGRect(x: 0, y: 0, width: 1, height: 0.5))
        )
        #expect(cropped.width == 8)
        #expect(cropped.height == 4)
        let pixel = try #require(firstPixel(cropped))
        #expect(pixel.0 > 200 && pixel.1 < 40 && pixel.2 < 40)
    }

    @Test func isOversizedUsesLongestEdgeAndPixelCount() {
        #expect(!OCRImage.isOversized(CGSize(width: 800, height: 600)))
        #expect(OCRImage.isOversized(CGSize(width: 4000, height: 10)))
        #expect(OCRImage.isOversized(CGSize(width: 10, height: 4000)))
        #expect(OCRImage.isOversized(CGSize(width: 6200, height: 8400)))
        #expect(!OCRImage.isOversized(.zero))
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
        guard let data = image.dataProvider?.data as Data?, data.count >= 4 else { return nil }
        return (data[0], data[1], data[2])
    }
}
