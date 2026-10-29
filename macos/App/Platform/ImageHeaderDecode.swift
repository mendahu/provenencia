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
    }

    /// The first image's declared size, or nil unless both edges are in
    /// `1...maxEdge` and `width * height <= maxPixels`. One header read; no
    /// pixels decode. A resolution that is absent, zero, negative or beyond
    /// 100,000 reads as 72, and an orientation outside 1...8 as 1.
    static func declaredImage(_ source: CGImageSource, maxEdge: Int, maxPixels: Int) -> DeclaredImage? {
        guard let props = CGImageSourceCopyPropertiesAtIndex(source, 0, nil) as? [CFString: Any],
              let width = (props[kCGImagePropertyPixelWidth] as? NSNumber)?.intValue,
              let height = (props[kCGImagePropertyPixelHeight] as? NSNumber)?.intValue,
              (1...maxEdge).contains(width), (1...maxEdge).contains(height),
              width * height <= maxPixels
        else { return nil }
        func dpi(_ key: CFString) -> CGFloat {
            guard let value = (props[key] as? NSNumber)?.doubleValue, value.isFinite,
                  (1...100_000).contains(value)
            else { return 72 }
            return CGFloat(value)
        }
        let orientation = (props[kCGImagePropertyOrientation] as? NSNumber)?.intValue ?? 1
        return DeclaredImage(
            width: width,
            height: height,
            dpiWidth: dpi(kCGImagePropertyDPIWidth),
            dpiHeight: dpi(kCGImagePropertyDPIHeight),
            orientation: (1...8).contains(orientation) ? orientation : 1
        )
    }

    /// The first image decoded eagerly at no more than `maxPixelSize` on its
    /// longest edge, turned upright when `applyTransform` is set.
    static func decode(_ source: CGImageSource, maxPixelSize: Int, applyTransform: Bool) -> CGImage? {
        CGImageSourceCreateThumbnailAtIndex(source, 0, [
            kCGImageSourceCreateThumbnailFromImageAlways: true,
            kCGImageSourceCreateThumbnailWithTransform: applyTransform,
            kCGImageSourceShouldCacheImmediately: true,
            kCGImageSourceThumbnailMaxPixelSize: maxPixelSize,
        ] as CFDictionary)
    }
}
