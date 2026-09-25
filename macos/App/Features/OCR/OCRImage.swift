import AppKit
import CoreGraphics
import Foundation

/// Stateless raster helpers for OCR. Coordinates for normalized rects are
/// **y-down** (unit square, origin top-left), matching Artifact locators.
enum OCRImage {
    /// Longest edge (points/pixels) at or above this is oversized.
    static let oversizedLongestEdge: CGFloat = 4000
    /// Width × height at or above this is oversized.
    static let oversizedPixelCount: CGFloat = 16_000_000

    static func cgImage(from image: NSImage) -> CGImage? {
        var rect = CGRect(origin: .zero, size: image.size)
        guard rect.width > 0, rect.height > 0 else { return nil }
        return image.cgImage(forProposedRect: &rect, context: nil, hints: nil)
    }

    /// Crops in the image's pixel space. `pixelRect` is y-down from the top-left.
    static func crop(_ image: CGImage, pixelRect: CGRect) -> CGImage? {
        let width = CGFloat(image.width)
        let height = CGFloat(image.height)
        guard width > 0, height > 0 else { return nil }
        let integral = pixelRect.integral
        guard integral.width > 0, integral.height > 0 else { return nil }
        let bounds = CGRect(x: 0, y: 0, width: width, height: height)
        let clipped = integral.intersection(bounds)
        guard clipped.width > 0, clipped.height > 0 else { return nil }
        // Bitmap / `CGImage.cropping(to:)` match locator y-down (origin top-left).
        return image.cropping(to: clipped)
    }

    /// `normalizedRect` is a y-down unit-square box (0…1).
    static func crop(_ image: CGImage, normalizedRect: CGRect) -> CGImage? {
        guard normalizedRect.width > 0, normalizedRect.height > 0 else { return nil }
        let pixel = CGRect(
            x: normalizedRect.minX * CGFloat(image.width),
            y: normalizedRect.minY * CGFloat(image.height),
            width: normalizedRect.width * CGFloat(image.width),
            height: normalizedRect.height * CGFloat(image.height)
        )
        return crop(image, pixelRect: pixel)
    }

    static func isOversized(_ size: CGSize) -> Bool {
        let width = max(size.width, 0)
        let height = max(size.height, 0)
        return max(width, height) >= oversizedLongestEdge
            || (width * height) >= oversizedPixelCount
    }
}
