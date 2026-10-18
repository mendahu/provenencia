import AppKit
import CoreGraphics
import Foundation
import ImageIO
import PDFKit
import UniformTypeIdentifiers

enum ArtifactRegionCropError: Error {
    case encodeFailed
}

/// Cuts a region out of viewer content: a pixel crop of an image, or a Core
/// Graphics draw of only the region's box on a PDF page, off the main actor. Stateless. Normalized rects are **y-down** (unit
/// square, origin top-left), matching Artifact locators.
enum ArtifactRegionCrop {
    /// Longest edge of an image the app sends. Go stores at most this edge
    /// (`core/ingest/stage.go` `stageMaxEdge`), so the app never sends more.
    static let maxLongEdge: CGFloat = 2048
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

    /// What `cut` needs from the viewer, held as values that are safe to hand
    /// to another thread.
    enum Snapshot: Sendable {
        case image(CGImage)
        /// The file the viewer opened and the 0-based page; `cut` opens its own copy.
        case pdf(url: URL, pageIndex: Int)
    }

    /// Reads the viewer on the main actor. `page` is the 1-based PDF page. Nil
    /// when the viewer holds nothing `cut` can cut.
    @MainActor
    static func snapshot(_ viewer: ArtifactViewerModel, page: Int?) -> Snapshot? {
        switch viewer.kind {
        case .image:
            guard let raster = viewer.displayImage, let image = cgImage(from: raster) else { return nil }
            return .image(image)
        case .pdf:
            guard let page, page >= 1, let url = viewer.pdfDocument?.documentURL else { return nil }
            return .pdf(url: url, pageIndex: page - 1)
        default:
            return nil
        }
    }

    /// The region's bounding box as the image to save, at most `maxLongEdge`
    /// on its long edge. Runs on any thread. Nil when the region is invalid or
    /// the snapshot cannot be cut.
    nonisolated static func cut(_ snapshot: Snapshot, region: ArtifactRegionDraft) -> CGImage? {
        guard region.isValid else { return nil }
        let box: CGImage?
        switch snapshot {
        case .image(let image):
            box = boundingBox(of: image, region: region)
        case .pdf(let url, let pageIndex):
            box = pageBox(url: url, pageIndex: pageIndex, region: region)
        }
        guard let box else { return nil }
        return scaledToFit(box, maxLongEdge: maxLongEdge)
    }

    /// The region's bounding box as the image to save. `page` is the 1-based
    /// PDF page. The cut runs off the main actor. Nil when the viewer holds
    /// nothing this can cut.
    @MainActor
    static func render(_ viewer: ArtifactViewerModel, page: Int?, region: ArtifactRegionDraft) async -> CGImage? {
        guard let snapshot = snapshot(viewer, page: page) else { return nil }
        return await Task.detached(priority: .userInitiated) {
            cut(snapshot, region: region)
        }.value
    }

    /// Draws only the region's box of the page as displayed (rotation applied),
    /// at the scale that brings the box to `maxLongEdge` where `maxPageScale` allows.
    private nonisolated static func pageBox(url: URL, pageIndex: Int, region: ArtifactRegionDraft) -> CGImage? {
        guard let document = CGPDFDocument(url as CFURL),
              let page = document.page(at: pageIndex + 1)
        else { return nil }
        let media = page.getBoxRect(.mediaBox)
        let rotated = abs(Int(page.rotationAngle)) % 180 == 90
        let visual = rotated ? CGSize(width: media.height, height: media.width) : media.size
        guard visual.width > 0, visual.height > 0 else { return nil }
        let n = ArtifactRegionGeometry.boundingRect(of: region.points)
        let box = CGRect(
            x: n.minX * visual.width,
            y: n.minY * visual.height,
            width: n.width * visual.width,
            height: n.height * visual.height
        )
        guard box.width > 0, box.height > 0 else { return nil }
        let scale = min(maxPageScale, maxLongEdge / max(box.width, box.height))
        let width = max(1, Int((box.width * scale).rounded()))
        let height = max(1, Int((box.height * scale).rounded()))
        guard let context = outputContext(width: width, height: height) else { return nil }
        context.setFillColor(red: 1, green: 1, blue: 1, alpha: 1)
        context.fill(CGRect(x: 0, y: 0, width: width, height: height))
        context.scaleBy(x: scale, y: scale)
        context.translateBy(x: -box.minX, y: -(visual.height - box.maxY))
        context.concatenate(
            page.getDrawingTransform(
                .mediaBox,
                rect: CGRect(origin: .zero, size: visual),
                rotate: 0,
                preserveAspectRatio: true
            )
        )
        context.drawPDFPage(page)
        guard let image = context.makeImage() else { return nil }
        return scaledToFit(image, maxLongEdge: maxLongEdge)
    }

    /// The one output bitmap: sRGB, 8 bits per component, premultiplied RGBA.
    private nonisolated static func outputContext(width: Int, height: Int) -> CGContext? {
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
        return context
    }

    /// Redraws `image` so its long edge is at most `maxLongEdge`; returns it as is when it already fits.
    nonisolated static func scaledToFit(_ image: CGImage, maxLongEdge: CGFloat) -> CGImage? {
        let long = CGFloat(max(image.width, image.height))
        guard long > maxLongEdge else { return image }
        let factor = maxLongEdge / long
        let width = max(1, Int((CGFloat(image.width) * factor).rounded()))
        let height = max(1, Int((CGFloat(image.height) * factor).rounded()))
        guard let context = outputContext(width: width, height: height) else { return nil }
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
