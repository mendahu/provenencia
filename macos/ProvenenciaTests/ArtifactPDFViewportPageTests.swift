import AppKit
import PDFKit
import Testing
@testable import Provenencia

@Suite
@MainActor
struct ArtifactPDFViewportPageTests {
    @Test func pageChangedNotificationUpdatesCallback() async throws {
        let document = ArtifactFileFixtures.blankPDFDocument(pageCount: 3)
        var reportedPages: [Int] = []
        let host = bootstrapHost(document: document) { page in
            reportedPages.append(page)
        }

        let window = hostWindow(for: host)
        defer { withExtendedLifetime(window) {} }

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

    @Test func syncZoomOnCurrentPageDoesNotSnapToPageOne() throws {
        let document = ArtifactFileFixtures.blankPDFDocument(pageCount: 3)
        let host = bootstrapHost(document: document) { _ in }

        let window = hostWindow(for: host)
        defer { withExtendedLifetime(window) {} }

        guard let pageThree = document.page(at: 2) else {
            Issue.record("Expected page at index 2")
            return
        }
        host.pdfView.go(to: pageThree)

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
        let host = bootstrapHost(document: document) { reported in
            page = reported
        }

        let window = hostWindow(for: host)
        defer { withExtendedLifetime(window) {} }

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

    private func bootstrapHost(
        document: PDFDocument,
        onPageChange: @escaping (Int) -> Void
    ) -> ArtifactPDFHostView {
        let host = ArtifactPDFHostView()
        host.frame = NSRect(x: 0, y: 0, width: 400, height: 500)
        host.onPageChange = onPageChange
        host.sync(
            document: document,
            page: 1,
            zoom: 1,
            resetPage: true,
            selectionEnabled: true,
            findSelection: nil,
            findActivationID: 0
        )
        host.startObserving()
        return host
    }

    private func hostWindow(for host: ArtifactPDFHostView) -> NSWindow {
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 400, height: 500),
            styleMask: .borderless,
            backing: .buffered,
            defer: false
        )
        window.contentView = host
        window.orderFrontRegardless()
        return window
    }
}
