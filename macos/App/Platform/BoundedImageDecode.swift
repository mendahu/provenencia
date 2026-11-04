import AppKit
import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

/// The bounded full-size image decode of a project object, for the Artifact viewer and the likeness
/// viewer: the caller's ``ImageDecodeLimits`` decide what is admitted (the sniffed type, the file's
/// bytes, the declared size), and the one decode is held to ``maxDecodePixels``. In the Artifact viewer
/// that one raster serves display, crop and OCR.
enum BoundedImageDecode {
    /// The default edge limit: an edge outside 1 to this is refused.
    static let maxEdge = 65535
    /// Most 8-bit pixels the viewer holds decoded; divided by the bytes per color component.
    /// Keep equal to Go's `raster.MaxDecodePixels`; `TestSwiftVocabularyMatchesGo` holds them.
    static let maxDecodePixels = 64000000

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
        let budget = maxDecodePixels / ImageHeaderDecode.depthFactor(depth)
        guard pixels > budget else { return longEdge }
        return Int((Double(longEdge) * (Double(budget) / Double(pixels)).squareRoot()).rounded(.down))
    }

    /// Why a decode refused.
    typealias Refusal = ImageHeaderDecode.Refusal

    /// The size in pixels the file declares, as the image is shown (turned for EXIF 5...8), or the
    /// reason it refused. One header read; no pixels decode. The same admission as ``decode(url:limits:)``.
    static func declaredSize(
        url: URL, limits: ImageDecodeLimits
    ) -> Result<(width: Int, height: Int), Refusal> {
        ImageHeaderDecode.admit(url: url, limits: limits).map { $0.1.uprightSize }
    }

    /// Decodes the image at `url`, or the reason it refused, under the same admission as ``declaredSize(url:limits:)``.
    static func decode(url: URL, limits: ImageDecodeLimits) -> Result<Decoded, Refusal> {
        let source: CGImageSource
        let declared: ImageHeaderDecode.DeclaredImage
        switch ImageHeaderDecode.admit(url: url, limits: limits) {
        case .success(let admitted): (source, declared) = admitted
        case .failure(let refusal): return .failure(refusal)
        }
        guard let raster = ImageHeaderDecode.decode(
            source,
            maxPixelSize: maxPixelSize(
                width: declared.width, height: declared.height, depth: declared.bitsPerComponent
            )
        ) else { return .failure(.unreadable) }
        let turned = declared.isTurned
        let pixelSize = declared.uprightSize
        let dpiWidth = turned ? declared.dpiHeight : declared.dpiWidth
        let dpiHeight = turned ? declared.dpiWidth : declared.dpiHeight
        let size = NSSize(
            width: CGFloat(pixelSize.width) * 72 / dpiWidth,
            height: CGFloat(pixelSize.height) * 72 / dpiHeight
        )
        return .success(Decoded(
            raster: raster,
            image: NSImage(cgImage: raster, size: size),
            pixelSize: pixelSize
        ))
    }
}
