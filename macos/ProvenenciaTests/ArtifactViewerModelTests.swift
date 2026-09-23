import AppKit
import Foundation
import PDFKit
import Testing
@testable import Provenencia

@Suite
@MainActor
struct ArtifactViewerModelTests {
    @Test func classifiesMediaTypes() {
        #expect(ArtifactViewerKind.classify(mediaType: "image/png") == .image)
        #expect(ArtifactViewerKind.classify(mediaType: "image/jpeg") == .image)
        #expect(ArtifactViewerKind.classify(mediaType: "application/pdf") == .pdf)
        #expect(ArtifactViewerKind.classify(mediaType: "application/x-pdf") == .pdf)
        #expect(ArtifactViewerKind.classify(mediaType: "audio/mpeg") == .audio)
        #expect(ArtifactViewerKind.classify(mediaType: "video/mp4") == .video)
        #expect(ArtifactViewerKind.classify(mediaType: "text/plain") == .unsupported)
        #expect(ArtifactViewerKind.classify(mediaType: "") == .unsupported)
    }

    @Test func capabilitiesMatchKind() {
        let image = ArtifactViewerKind.image.locatorCapabilities
        #expect(!image.supportsPageLocator)
        #expect(image.supportsRegionLocator)
        #expect(image.supportsSpatialZoom)
        #expect(!image.supportsTimeRangeLocator)

        let pdf = ArtifactViewerKind.pdf.locatorCapabilities
        #expect(pdf.supportsPageLocator)
        #expect(pdf.supportsRegionLocator)
        #expect(pdf.supportsSpatialZoom)
        #expect(!pdf.supportsTimeRangeLocator)

        let audio = ArtifactViewerKind.audio.locatorCapabilities
        #expect(!audio.supportsPageLocator)
        #expect(!audio.supportsRegionLocator)
        #expect(!audio.supportsSpatialZoom)
        #expect(audio.supportsTimeRangeLocator)

        let video = ArtifactViewerKind.video.locatorCapabilities
        #expect(!video.supportsPageLocator)
        #expect(!video.supportsRegionLocator)
        #expect(video.supportsTimeRangeLocator)

        let unsupported = ArtifactViewerKind.unsupported.locatorCapabilities
        #expect(!unsupported.supportsPageLocator)
        #expect(!unsupported.supportsRegionLocator)
        #expect(!unsupported.supportsTimeRangeLocator)
        #expect(!ArtifactViewerKind.video.isRenderableInS706)
    }

    @Test func clampsZoomAndPage() {
        #expect(ArtifactViewerModel.clampZoom(0.1) == ArtifactViewerModel.minZoom)
        #expect(ArtifactViewerModel.clampZoom(9) == ArtifactViewerModel.maxZoom)
        #expect(ArtifactViewerModel.clampZoom(1.5) == 1.5)
        #expect(ArtifactViewerModel.clampPage(0, pageCount: 5) == 1)
        #expect(ArtifactViewerModel.clampPage(99, pageCount: 5) == 5)
        #expect(ArtifactViewerModel.clampPage(3, pageCount: 5) == 3)
    }

    @Test func audioKindShowsComingSoonWithoutCrash() async {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/aa/bb/dummy.mp3"
        writeFile(project: project, relPath: rel, data: Data([1, 2, 3]))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "audio/mpeg"
            )
        )
        #expect(model.kind == .audio)
        #expect(model.emptyReason == .mediaNotYetAvailable)
        #expect(model.displayImage == nil)
        #expect(!model.supportsPages)
    }

    @Test func missingFileSetsEmptyReason() async {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: "objects/aa/bb/nope.png",
                mediaType: "image/png"
            )
        )
        #expect(model.emptyReason == .missingFile)
    }

    @Test func loadsPNGImageViaProjectFiles() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/sample.png"
        let png = try #require(minimalPNGData())
        writeFile(project: project, relPath: rel, data: png)

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "image/png"
            )
        )
        #expect(model.kind == .image)
        #expect(model.emptyReason == .idle)
        #expect(model.displayImage != nil)
        #expect(model.pageCount == 1)
        #expect(model.supportsSpatialZoom)
    }

    @Test func loadsPDFAndPages() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/sample.pdf"
        let pdfData = try #require(minimalPDFData(pageCount: 3))
        writeFile(project: project, relPath: rel, data: pdfData)

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        #expect(model.kind == .pdf)
        #expect(model.pageCount == 3)
        #expect(model.page == 1)
        #expect(model.displayImage != nil)
        // Layout size is media-box points (not the backing-scale pixel raster).
        #expect(model.displayImage?.size.width == 612)
        #expect(model.displayImage?.size.height == 792)

        model.goToNextPage()
        #expect(model.page == 2)
        model.setPage(99)
        #expect(model.page == 3)
        model.goToPreviousPage()
        #expect(model.page == 2)
    }

    @Test func zoomStepsClamp() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/sample.png"
        writeFile(project: project, relPath: rel, data: try #require(minimalPNGData()))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "image/png"
            )
        )
        model.zoomOut()
        #expect(model.zoom == ArtifactViewerModel.minZoom || model.zoom < 1)
        model.setZoom(10)
        #expect(model.zoom == ArtifactViewerModel.maxZoom)
        #expect(model.zoomPercentLabel.hasSuffix("%"))
    }

    // MARK: - Fixtures

    private func makeTempProject() -> URL {
        let url = FileManager.default.temporaryDirectory
            .appendingPathComponent("pv-artifact-viewer-\(UUID().uuidString)", isDirectory: true)
        try? FileManager.default.createDirectory(at: url, withIntermediateDirectories: true)
        return url
    }

    private func writeFile(project: URL, relPath: String, data: Data) {
        let url = project.appendingPathComponent(relPath)
        try? FileManager.default.createDirectory(
            at: url.deletingLastPathComponent(),
            withIntermediateDirectories: true
        )
        try? data.write(to: url)
    }

    private func minimalPNGData() -> Data? {
        let image = NSImage(size: NSSize(width: 8, height: 8))
        image.lockFocus()
        NSColor.red.setFill()
        NSBezierPath(rect: NSRect(x: 0, y: 0, width: 8, height: 8)).fill()
        image.unlockFocus()
        guard let tiff = image.tiffRepresentation,
              let rep = NSBitmapImageRep(data: tiff)
        else { return nil }
        return rep.representation(using: .png, properties: [:])
    }

    private func minimalPDFData(pageCount: Int) -> Data? {
        let document = PDFDocument()
        for index in 0..<pageCount {
            let page = PDFPage()
            // Ensure non-empty media box via annotation of page size.
            _ = index
            document.insert(page, at: document.pageCount)
        }
        return document.dataRepresentation()
    }
}
