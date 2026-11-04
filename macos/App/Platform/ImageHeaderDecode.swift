import AppKit
import Foundation
import ImageIO
import UniformTypeIdentifiers

/// What a caller admits as a picture file: the sniffed type must conform to one of `typeIdentifiers`
/// (an empty list refuses), both edges must be in `1...maxEdge`, and a file over `maxBytes`, when set,
/// refuses before ImageIO opens it.
struct ImageDecodeLimits: Equatable, Sendable {
    var typeIdentifiers: [String]
    var maxEdge = BoundedImageDecode.maxEdge
    var maxBytes: Int?
}

/// The ImageIO steps the thumbnail cache, the Artifact viewer and the likeness viewer share:
/// admitting an image by its header, and a bounded decode.
enum ImageHeaderDecode {
    /// An image declaring more 8-bit pixels than this is refused before any pixel decodes;
    /// the count is divided by the bytes per color component.
    static let maxDeclaredPixels = 256000000

    /// Why an image is refused.
    enum Refusal: Error, Equatable, Sendable {
        /// The header declares a size outside the edge or depth-scaled pixel budget, or the file is
        /// over the byte limit.
        case tooLarge
        /// Anything else: a sniffed type that conforms to none of the identifiers, an unreadable
        /// header, or a failed decode.
        case unreadable
    }

    /// Bytes per color component, at least 1: a 16-bit image counts twice, a 32-bit float image four times.
    static func depthFactor(_ bitsPerComponent: Int) -> Int {
        max(1, bitsPerComponent / 8)
    }

    /// The one admission of a picture file by its header: the sniffed type conforms to one of
    /// `limits.typeIdentifiers`, and the declared size is within `limits.maxEdge` and the depth-scaled
    /// pixel budget. Reads the header only; no pixels decode, and `limits.maxBytes` is the caller's
    /// (it needs the file's size, not the source).
    static func admit(_ source: CGImageSource, limits: ImageDecodeLimits) -> Result<DeclaredImage, Refusal> {
        guard let sniffed = (CGImageSourceGetType(source) as String?).flatMap({ UTType($0) }),
              limits.typeIdentifiers.contains(where: { identifier in
                  UTType(identifier).map { sniffed.conforms(to: $0) } ?? false
              }),
              let declared = declaredImage(source)
        else { return .failure(.unreadable) }
        guard declared.fits(
            maxEdge: limits.maxEdge, maxPixels: maxDeclaredPixels / depthFactor(declared.bitsPerComponent)
        ) else { return .failure(.tooLarge) }
        return .success(declared)
    }

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

        /// EXIF orientations 5...8 turn the image a quarter, so it is shown with its edges swapped.
        var isTurned: Bool { orientation >= 5 }

        /// The size in pixels as the image is shown: the declared edges, swapped when ``isTurned``.
        var uprightSize: (width: Int, height: Int) {
            isTurned ? (height, width) : (width, height)
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
