import AppKit
import Foundation
import ImageIO

/// The two ImageIO steps the thumbnail cache and the Artifact viewer share: reading what an
/// image's header declares, and a bounded decode.
enum ImageHeaderDecode {
    /// What an image's header declares: its size, resolution and EXIF orientation.
    struct DeclaredImage: Equatable, Sendable {
        var width: Int
        var height: Int
        var dpiWidth: CGFloat
        var dpiHeight: CGFloat
        var orientation: Int
        /// Bits per color component (8 when the header gives none or an odd one), so a 16-bit or
        /// float image counts for the bytes it holds.
        var bitsPerComponent: Int

        /// True when both edges are in `1...maxEdge` and `width * height <= maxPixels`.
        func fits(maxEdge: Int, maxPixels: Int) -> Bool {
            (1...maxEdge).contains(width) && (1...maxEdge).contains(height) && width * height <= maxPixels
        }
    }

    /// What the first image's header declares, or nil when it declares no size. One header
    /// read; no pixels decode. The caller applies its own bounds through ``DeclaredImage/fits(maxEdge:maxPixels:)``. A resolution that is absent, zero, negative or beyond
    /// 100,000 reads as 72, an orientation outside 1...8 as 1, and a depth outside 1...64 as 8.
    static func declaredImage(_ source: CGImageSource) -> DeclaredImage? {
        guard let props = CGImageSourceCopyPropertiesAtIndex(source, 0, nil) as? [CFString: Any],
              let width = (props[kCGImagePropertyPixelWidth] as? NSNumber)?.intValue,
              let height = (props[kCGImagePropertyPixelHeight] as? NSNumber)?.intValue
        else { return nil }
        func dpi(_ key: CFString) -> CGFloat {
            guard let value = (props[key] as? NSNumber)?.doubleValue, value.isFinite,
                  (1...100_000).contains(value)
            else { return 72 }
            return CGFloat(value)
        }
        let orientation = (props[kCGImagePropertyOrientation] as? NSNumber)?.intValue ?? 1
        let depth = (props[kCGImagePropertyDepth] as? NSNumber)?.intValue ?? 8
        return DeclaredImage(
            width: width,
            height: height,
            dpiWidth: dpi(kCGImagePropertyDPIWidth),
            dpiHeight: dpi(kCGImagePropertyDPIHeight),
            orientation: (1...8).contains(orientation) ? orientation : 1,
            bitsPerComponent: (1...64).contains(depth) ? depth : 8
        )
    }

    /// The first image decoded eagerly at no more than `maxPixelSize` on its
    /// longest edge, turned upright by its EXIF orientation.
    static func decode(_ source: CGImageSource, maxPixelSize: Int) -> CGImage? {
        CGImageSourceCreateThumbnailAtIndex(source, 0, [
            kCGImageSourceCreateThumbnailFromImageAlways: true,
            kCGImageSourceCreateThumbnailWithTransform: true,
            kCGImageSourceShouldCacheImmediately: true,
            kCGImageSourceThumbnailMaxPixelSize: maxPixelSize,
        ] as CFDictionary)
    }
}
