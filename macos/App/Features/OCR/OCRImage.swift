import CoreGraphics
import Foundation

/// Oversize rule for OCR input. Cutting a region from an image lives in
/// `ArtifactRegionCrop`.
enum OCRImage {
    /// Longest edge (points/pixels) at or above this is oversized.
    static let oversizedLongestEdge: CGFloat = 4000
    /// Width × height at or above this is oversized.
    static let oversizedPixelCount: CGFloat = 16_000_000

    static func isOversized(_ size: CGSize) -> Bool {
        let width = max(size.width, 0)
        let height = max(size.height, 0)
        return max(width, height) >= oversizedLongestEdge
            || (width * height) >= oversizedPixelCount
    }
}
