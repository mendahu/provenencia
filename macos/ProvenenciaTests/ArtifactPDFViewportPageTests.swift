import AppKit
import PDFKit
import Testing
@testable import Provenencia

@Suite(.hostedWindows)
@MainActor
struct ArtifactPDFViewportPageTests {
    @Test func pageChangedNotificationUpdatesCallback() async throws {
        let document = ArtifactFileFixtures.blankPDFDocument(pageCount: 3)
        var reportedPages: [Int] = []
        let host = ArtifactFileFixtures.pdfHost(document: document) { page in
            reportedPages.append(page)
        }

        hostInBorderlessWindow(nsView: host, size: CGSize(width: 400, height: 500))

        guard let pageThree = document.page(at: 2) else {
            Issue.record("Expected page at index 2")
            return
        }
        host.pdfView.go(to: pageThree)
        #expect(await waitUntil { reportedPages.last == 3 })

        guard let pageOne = document.page(at: 0) else {
            Issue.record("Expected page at index 0")
            return
        }
        host.pdfView.go(to: pageOne)
        #expect(await waitUntil { reportedPages.last == 1 })
    }

    @Test func syncZoomOnCurrentPageDoesNotSnapToPageOne() async throws {
        let document = ArtifactFileFixtures.blankPDFDocument(pageCount: 3)
        let host = ArtifactFileFixtures.pdfHost(document: document) { _ in }

        hostInBorderlessWindow(nsView: host, size: CGSize(width: 400, height: 500))

        guard let pageThree = document.page(at: 2) else {
            Issue.record("Expected page at index 2")
            return
        }
        host.pdfView.go(to: pageThree)
        #expect(await waitUntil {
            host.pdfView.currentPage.map { document.index(for: $0) == 2 } ?? false
        })

        host.sync(
            document: document,
            page: 3,
            zoom: 2.0,
            resetPage: false,
            selectionEnabled: true,
            findSelection: nil,
            findActivationID: 0
        )

        guard let currentPage = host.pdfView.currentPage else {
            Issue.record("Expected a current PDF page")
            return
        }
        #expect(document.index(for: currentPage) == 2)
        #expect(abs(host.pdfView.scaleFactor - 2.0) < 0.01)
    }

    @Test func pageReportThenSyncZoomKeepsPage() async throws {
        let document = ArtifactFileFixtures.blankPDFDocument(pageCount: 3)
        var page = 1
        let host = ArtifactFileFixtures.pdfHost(document: document) { reported in
            page = reported
        }

        hostInBorderlessWindow(nsView: host, size: CGSize(width: 400, height: 500))

        guard let pageThree = document.page(at: 2) else {
            Issue.record("Expected page at index 2")
            return
        }
        host.pdfView.go(to: pageThree)
        #expect(await waitUntil { page == 3 })

        host.sync(
            document: document,
            page: page,
            zoom: 1.5,
            resetPage: false,
            selectionEnabled: true,
            findSelection: nil,
            findActivationID: 0
        )

        guard let currentPage = host.pdfView.currentPage else {
            Issue.record("Expected a current PDF page")
            return
        }
        #expect(document.index(for: currentPage) == 2)
    }
}
