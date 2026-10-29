import AppKit
import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

/// The Artifact viewer's bounded image decode: the file's sniffed type must match
/// the type Go assigned, the declared size must be within budget, and the one
/// decode is held to ``maxDecodePixels``. That one raster serves display, crop and OCR.
enum ArtifactImageDecode {
    /// An edge outside 1 to this is refused.
    static let maxEdge = 65535
    /// An image declaring more pixels than this is refused before any pixel decodes.
    static let maxDeclaredPixels = 256000000
    /// Most pixels the viewer holds decoded. Keep equal to Go's `raster.MaxDecodePixels`;
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

    /// The `MaxPixelSize` that brings a `width` x `height` image within ``maxDecodePixels``:
    /// the long edge itself when it fits, else the long edge scaled by the square root of
    /// the ratio. Orientation does not change the long edge or the product.
    static func maxPixelSize(width: Int, height: Int) -> Int {
        let longEdge = max(width, height)
        let pixels = width * height
        guard pixels > maxDecodePixels else { return longEdge }
        return Int((Double(longEdge) * (Double(maxDecodePixels) / Double(pixels)).squareRoot()).rounded(.down))
    }

    /// Decodes the image at `url`, or nil for any refusal: a type that does not match
    /// `mediaType` (else the extension of `relPath`), a declared size over budget, an
    /// unreadable header, or a failed decode.
    static func decode(url: URL, mediaType: String, relPath: String) -> Decoded? {
        guard let source = CGImageSourceCreateWithURL(url as CFURL, nil),
              let sniffed = CGImageSourceGetType(source) as String?,
              let expected = expectedType(mediaType: mediaType, relPath: relPath),
              UTType(sniffed)?.conforms(to: expected) == true,
              let declared = ThumbnailDecodePolicy.declaredImage(
                  source, maxEdge: maxEdge, maxPixels: maxDeclaredPixels
              ),
              let raster = ThumbnailDecodePolicy.decode(
                  source,
                  maxPixelSize: maxPixelSize(width: declared.width, height: declared.height),
                  applyTransform: appliesOrientation
              )
        else { return nil }
        let turned = appliesOrientation && declared.orientation >= 5
        let pixelWidth = turned ? declared.height : declared.width
        let pixelHeight = turned ? declared.width : declared.height
        let dpiWidth = turned ? declared.dpiHeight : declared.dpiWidth
        let dpiHeight = turned ? declared.dpiWidth : declared.dpiHeight
        let size = NSSize(
            width: CGFloat(pixelWidth) * 72 / dpiWidth,
            height: CGFloat(pixelHeight) * 72 / dpiHeight
        )
        return Decoded(
            raster: raster,
            image: NSImage(cgImage: raster, size: size),
            pixelSize: (pixelWidth, pixelHeight)
        )
    }

    /// The type Go assigned: what macOS knows for its media type, else for the extension
    /// Go gave the object from the same table. A media type macOS only makes up a
    /// placeholder for (`image/tif`) counts as unknown.
    private static func expectedType(mediaType: String, relPath: String) -> UTType? {
        if let known = UTType(mimeType: mediaType), !known.isDynamic { return known }
        return UTType(filenameExtension: (relPath as NSString).pathExtension)
    }
}
