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
        #expect(model.pdfDocument != nil)
        #expect(model.displayImage == nil)
        #expect(model.pdfPageMediaSize?.width == 612)
        #expect(model.pdfPageMediaSize?.height == 792)

        model.goToNextPage()
        #expect(model.page == 2)
        #expect(model.pdfDocument != nil)
        #expect(model.displayImage == nil)
        model.setPage(99)
        #expect(model.page == 3)
        model.goToPreviousPage()
        #expect(model.page == 2)
    }

    @Test func unloadClearsUserSelection() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/sample.pdf"
        writeFile(project: project, relPath: rel, data: try #require(minimalPDFData(pageCount: 2)))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        model.installUserSelectionForTesting(text: "Alice", page: 2)
        #expect(model.hasUserSelection)
        #expect(model.userSelectionPage == 2)

        model.unload()
        #expect(!model.hasUserSelection)
        #expect(model.userSelectionPage == nil)

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        #expect(!model.hasUserSelection)
    }

    @Test func findHitsChangePageAndWrap() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/find.pdf"
        writeFile(project: project, relPath: rel, data: try #require(textPDFData(pages: [
            "Alice lived in Boston",
            "Boston city register",
            "Alice later moved",
        ])))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        #expect(model.findHasTextLayer)
        model.setPage(2)
        model.openFind()
        model.findQuery = "Alice"
        #expect(model.findMatchCount == 2)
        #expect(model.page == 1)
        #expect(model.activeFindIndex == 0)
        #expect(model.findNote == .match(page: 1, jumped: true, nextMatchPage: 3))

        model.findNext()
        #expect(model.page == 3)
        #expect(model.activeFindIndex == 1)

        model.findNext()
        #expect(model.page == 1)
        #expect(model.activeFindIndex == 0)
        #expect(model.findNote == .wrappedFirst(page: 1))

        model.findPrevious()
        #expect(model.page == 3)
        #expect(model.findNote == .wrappedLast(page: 3))
    }

    @Test func emptyFindQueryDoesNotChangePage() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/find.pdf"
        writeFile(project: project, relPath: rel, data: try #require(textPDFData(pages: [
            "Alice lived in Boston",
            "Boston city register",
            "Alice later moved",
        ])))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        model.setPage(2)
        model.openFind()
        model.findQuery = "   "
        model.findNext()
        #expect(model.page == 2)
        #expect(model.findMatchCount == 0)
        #expect(model.findNote == .idle)
        #expect(model.activeFindSelection == nil)
    }

    @Test func findNoMatchesLeavesPageAndNotes() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/find.pdf"
        writeFile(project: project, relPath: rel, data: try #require(textPDFData(pages: [
            "Alice lived in Boston",
            "Boston city register",
            "Alice later moved",
        ])))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        model.setPage(2)
        model.openFind()
        model.findQuery = "zzzz-no-match"
        #expect(model.page == 2)
        #expect(model.findMatchCount == 0)
        #expect(model.findNote == .noMatches(pageCount: 3))
        #expect(model.findIsInvalid)
        #expect(model.findCountSuffix == "0")
        #expect(model.findStatusNote == L10n.ArtifactViewer.findNoteNoMatches(pageCount: 3))
    }

    @Test func findOnScannedPDFShowsNoTextLayer() async throws {
        let model = ArtifactViewerModel()
        let project = makeTempProject()
        defer { try? FileManager.default.removeItem(at: project) }
        let rel = "objects/ab/cd/scan.pdf"
        writeFile(project: project, relPath: rel, data: try #require(minimalPDFData(pageCount: 2)))

        await model.load(
            ArtifactViewerSource(
                projectDir: project.path,
                relPath: rel,
                mediaType: "application/pdf"
            )
        )
        #expect(!model.findHasTextLayer)
        model.openFind()
        #expect(model.isFindPresented)
        #expect(model.findNote == .noTextLayer)
        #expect(model.findStatusNote == String(localized: L10n.ArtifactViewer.findNoteNoTextLayer))
        model.findQuery = "Alice"
        #expect(model.findMatchCount == 0)
        #expect(model.page == 1)
        #expect(model.findNote == .noTextLayer)
    }

    @Test func findIsNoOpOnImage() async throws {
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
        model.openFind()
        model.findQuery = "Alice"
        model.findNext()
        #expect(!model.isFindPresented)
        #expect(model.findMatchCount == 0)
        #expect(model.kind == .image)
        #expect(model.page == 1)
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

    /// Born-digital pages with `/Tj` strings so `PDFDocument.findString` can hit.
    private func textPDFData(pages: [String]) -> Data? {
        func escape(_ text: String) -> String {
            text
                .replacingOccurrences(of: "\\", with: "\\\\")
                .replacingOccurrences(of: "(", with: "\\(")
                .replacingOccurrences(of: ")", with: "\\)")
        }

        var body = "%PDF-1.4\n"
        var offsets: [Int] = [0]

        func addObject(_ object: String) {
            offsets.append(body.utf8.count)
            body += object
        }

        var nextObject = 4
        var pageObjects: [Int] = []
        var contents: [(page: Int, contents: Int, text: String)] = []
        for text in pages {
            pageObjects.append(nextObject)
            contents.append((page: nextObject, contents: nextObject + 1, text: text))
            nextObject += 2
        }

        let kids = pageObjects.map { "\($0) 0 R" }.joined(separator: " ")
        addObject("1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n")
        addObject("2 0 obj << /Type /Pages /Kids [\(kids)] /Count \(pages.count) >> endobj\n")
        addObject("3 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n")
        for item in contents {
            addObject(
                "\(item.page) 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents \(item.contents) 0 R /Resources << /Font << /F1 3 0 R >> >> >> endobj\n"
            )
            let stream = "BT /F1 24 Tf 72 720 Td (\(escape(item.text))) Tj ET\n"
            addObject(
                "\(item.contents) 0 obj << /Length \(stream.utf8.count) >> stream\n\(stream)endstream\nendobj\n"
            )
        }

        let xrefOffset = body.utf8.count
        var xref = "xref\n0 \(offsets.count)\n0000000000 65535 f \n"
        for offset in offsets.dropFirst() {
            xref += String(format: "%010d 00000 n \n", offset)
        }
        body += xref
        body += "trailer << /Size \(offsets.count) /Root 1 0 R >>\nstartxref\n\(xrefOffset)\n%%EOF\n"
        return Data(body.utf8)
    }
}
