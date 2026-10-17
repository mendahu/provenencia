import AppKit
import CoreGraphics
import Foundation
import ImageIO
import PDFKit
import UniformTypeIdentifiers

enum ArtifactRegionCropError: Error {
    case encodeFailed
}

/// Cuts a region out of viewer content: a pixel crop of an image, a PDFKit
/// render of a PDF page. Stateless. Normalized rects are **y-down** (unit
/// square, origin top-left), matching Artifact locators.
enum ArtifactRegionCrop {
    /// Longest edge of an image the app sends. Go stores at most this edge
    /// (`core/ingest/stage.go` `stageMaxEdge`), so the app never sends more.
    static let maxLongEdge: CGFloat = 2048
    /// Longest edge of the page bitmap rendered before cropping.
    static let maxPageLongEdge: CGFloat = 4096
    /// Render scale cap, in pixels per point.
    static let maxPageScale: CGFloat = 4

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

    /// The uncapped crop of the region's bounding box.
    static func boundingBox(of image: CGImage, region: ArtifactRegionDraft) -> CGImage? {
        crop(image, normalizedRect: ArtifactRegionGeometry.boundingRect(of: region.points))
    }

    /// The region's bounding box as the image to save, at most `maxLongEdge`
    /// on its long edge. `page` is the 1-based PDF page. Nil when the viewer
    /// holds nothing this can cut.
    @MainActor
    static func render(_ viewer: ArtifactViewerModel, page: Int?, region: ArtifactRegionDraft) -> CGImage? {
        guard region.isValid else { return nil }
        let box: CGImage?
        switch viewer.kind {
        case .image:
            guard let raster = viewer.displayImage, let image = cgImage(from: raster) else { return nil }
            box = boundingBox(of: image, region: region)
        case .pdf:
            guard let page, page >= 1,
                  let pdfPage = viewer.pdfDocument?.page(at: page - 1),
                  let image = pageBitmap(pdfPage, region: region)
            else { return nil }
            box = boundingBox(of: image, region: region)
        default:
            return nil
        }
        guard let box else { return nil }
        return scaledToFit(box, maxLongEdge: maxLongEdge)
    }

    /// The page as displayed (rotation applied), sized so the region's box
    /// reaches `maxLongEdge` where the page allows.
    private static func pageBitmap(_ page: PDFPage, region: ArtifactRegionDraft) -> CGImage? {
        let media = page.bounds(for: .mediaBox).size
        let rotated = abs(page.rotation) % 180 == 90
        let visual = rotated ? CGSize(width: media.height, height: media.width) : media
        guard visual.width > 0, visual.height > 0 else { return nil }
        let box = ArtifactRegionGeometry.boundingRect(of: region.points)
        let boxLong = max(box.width * visual.width, box.height * visual.height)
        let visualLong = max(visual.width, visual.height)
        guard boxLong > 0 else { return nil }
        let scale = min(maxPageScale, maxLongEdge / boxLong, maxPageLongEdge / visualLong)
        let size = CGSize(width: visual.width * scale, height: visual.height * scale)
        return cgImage(from: page.thumbnail(of: size, for: .mediaBox))
    }

    /// Redraws `image` so its long edge is at most `maxLongEdge`; returns it as is when it already fits.
    static func scaledToFit(_ image: CGImage, maxLongEdge: CGFloat) -> CGImage? {
        let long = CGFloat(max(image.width, image.height))
        guard long > maxLongEdge else { return image }
        let factor = maxLongEdge / long
        let width = max(1, Int((CGFloat(image.width) * factor).rounded()))
        let height = max(1, Int((CGFloat(image.height) * factor).rounded()))
        guard let space = CGColorSpace(name: CGColorSpace.sRGB),
              let context = CGContext(
                  data: nil,
                  width: width,
                  height: height,
                  bitsPerComponent: 8,
                  bytesPerRow: 0,
                  space: space,
                  bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue
              )
        else { return nil }
        context.interpolationQuality = .high
        context.draw(image, in: CGRect(x: 0, y: 0, width: width, height: height))
        return context.makeImage()
    }

    /// Encodes `image` as a PNG in the temporary directory. The caller removes the file.
    static func writeTemporaryPNG(_ image: CGImage) throws -> URL {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("provenencia-depiction-\(UUID().uuidString).png")
        guard let destination = CGImageDestinationCreateWithURL(
            url as CFURL,
            UTType.png.identifier as CFString,
            1,
            nil
        ) else {
            throw ArtifactRegionCropError.encodeFailed
        }
        CGImageDestinationAddImage(destination, image, nil)
        guard CGImageDestinationFinalize(destination) else {
            try? FileManager.default.removeItem(at: url)
            throw ArtifactRegionCropError.encodeFailed
        }
        return url
    }
}
