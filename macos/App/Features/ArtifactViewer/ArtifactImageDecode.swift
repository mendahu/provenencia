import AppKit
import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

/// The Artifact viewer's bounded image decode: the file's sniffed type must conform to one of
/// the type identifiers Go's stored-type table gives its media type, the declared size must be within budget, and the one
/// decode is held to ``maxDecodePixels``. That one raster serves display, crop and OCR.
enum ArtifactImageDecode {
    /// An edge outside 1 to this is refused.
    static let maxEdge = 65535
    /// An image declaring more 8-bit pixels than this is refused before any pixel decodes;
    /// the count is divided by the bytes per color component.
    static let maxDeclaredPixels = 256000000
    /// Most 8-bit pixels the viewer holds decoded; divided by the bytes per color component. Keep equal to Go's `raster.MaxDecodePixels`;
    /// `TestSwiftVocabularyMatchesGo` holds them.
    static let maxDecodePixels = 64000000
    /// Whether the decode turns an EXIF-rotated image upright, as `NSImage(contentsOf:)` shows it.
    static let appliesOrientation = true

    /// The decoded image and what describes it.
    ///
    /// `@unchecked Sendable`: `NSImage` is not Sendable, but the value is only
    /// handed once from a load task onto the main actor.
    struct Decoded: @unchecked Sendable {
        /// The decoded pixels, at most ``maxDecodePixels``.
        let raster: CGImage
        /// `raster` at the size the file declares in points, so layout and zoom match the original.
        let image: NSImage
        /// The declared size in pixels, as the image is shown (rotated for EXIF 5...8).
        let pixelSize: (width: Int, height: Int)
    }

    /// The `MaxPixelSize` that brings a `width` x `height` image of `depth` bits per component
    /// within ``maxDecodePixels`` (scaled by bytes per component):
    /// the long edge itself when it fits, else the long edge scaled by the square root of
    /// the ratio. Orientation does not change the long edge or the product.
    static func maxPixelSize(width: Int, height: Int, depth: Int = 8) -> Int {
        let longEdge = max(width, height)
        let pixels = width * height
        let budget = maxDecodePixels / depthFactor(depth)
        guard pixels > budget else { return longEdge }
        return Int((Double(longEdge) * (Double(budget) / Double(pixels)).squareRoot()).rounded(.down))
    }

    /// Bytes per color component, at least 1: a 16-bit image counts twice, a 32-bit float image four times.
    private static func depthFactor(_ bitsPerComponent: Int) -> Int {
        max(1, bitsPerComponent / 8)
    }

    /// Why a decode refused.
    enum Refusal: Error, Equatable, Sendable {
        /// The header declares a size outside the edge or depth-scaled pixel budget.
        case tooLarge
        /// Anything else: a sniffed type that conforms to none of the identifiers, an unreadable
        /// header, or a failed decode.
        case unreadable
    }

    /// Decodes the image at `url`, or the reason it refused. The sniffed type must conform to one
    /// of `typeIdentifiers` (an empty list refuses); the declared size must be within budget.
    static func decode(url: URL, typeIdentifiers: [String]) -> Result<Decoded, Refusal> {
        guard let source = CGImageSourceCreateWithURL(url as CFURL, nil),
              let sniffed = (CGImageSourceGetType(source) as String?).flatMap({ UTType($0) }),
              typeIdentifiers.contains(where: { identifier in
                  UTType(identifier).map { sniffed.conforms(to: $0) } ?? false
              }),
              let declared = ImageHeaderDecode.declaredImage(source)
        else { return .failure(.unreadable) }
        guard declared.fits(maxEdge: maxEdge, maxPixels: maxDeclaredPixels / depthFactor(declared.bitsPerComponent))
        else { return .failure(.tooLarge) }
        guard let raster = ImageHeaderDecode.decode(
            source,
            maxPixelSize: maxPixelSize(
                width: declared.width, height: declared.height, depth: declared.bitsPerComponent
            ),
            applyTransform: appliesOrientation
        ) else { return .failure(.unreadable) }
        let turned = appliesOrientation && declared.orientation >= 5
        let pixelWidth = turned ? declared.height : declared.width
        let pixelHeight = turned ? declared.width : declared.height
        let dpiWidth = turned ? declared.dpiHeight : declared.dpiWidth
        let dpiHeight = turned ? declared.dpiWidth : declared.dpiHeight
        let size = NSSize(
            width: CGFloat(pixelWidth) * 72 / dpiWidth,
            height: CGFloat(pixelHeight) * 72 / dpiHeight
        )
        return .success(Decoded(
            raster: raster,
            image: NSImage(cgImage: raster, size: size),
            pixelSize: (pixelWidth, pixelHeight)
        ))
    }
}
