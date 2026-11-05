import AppKit
import PDFKit
import Testing
@testable import Provenencia

@Suite
@MainActor
struct ArtifactRegionOverlayTests {
    private static let triangle = [
        CGPoint(x: 0.1, y: 0.1), CGPoint(x: 0.9, y: 0.1), CGPoint(x: 0.5, y: 0.9),
    ]

    /// A 300×400 PDF page in a 400×500 host, laid out, with `committed` on the overlay.
    private func makeHost(
        committed: [CGPoint],
        kind: ArtifactRegionKind = .freeform
    ) -> (host: ArtifactPDFHostView, window: NSWindow, media: CGRect) {
        let document = ArtifactFileFixtures.sizedPDFDocument(size: CGSize(width: 300, height: 400))
        let host = ArtifactFileFixtures.pdfHost(document: document)
        let window = hostInBorderlessWindow(nsView: host, size: CGSize(width: 400, height: 500))
        host.layoutSubtreeIfNeeded()
        host.overlayView.committed = ArtifactRegionDraft(kind: kind, points: committed)
        return (host, window, host.overlayView.mediaRectOverride ?? .zero)
    }

    private func event(
        _ type: NSEvent.EventType,
        at normalized: CGPoint,
        in host: ArtifactPDFHostView,
        window: NSWindow
    ) -> NSEvent {
        let media = host.overlayView.mediaRectOverride ?? .zero
        return NSEvent.mouseEvent(
            with: type,
            location: host.overlayView.convert(
                ArtifactRegionGeometry.documentPoint(normalized: normalized, imageRect: media),
                to: nil
            ),
            modifierFlags: [],
            timestamp: 0,
            windowNumber: window.windowNumber,
            context: nil,
            eventNumber: 0,
            clickCount: 1,
            pressure: 1
        )!
    }

    @Test func pdfHostHitTestFindsVertex() {
        let (host, window, media) = makeHost(committed: Self.triangle)
        defer { withExtendedLifetime(window) {} }
        #expect(abs(media.width - 300) <= 1 && abs(media.height - 400) <= 1)

        for vertex in [Self.triangle[0], Self.triangle[2]] {
            let overlayPoint = ArtifactRegionGeometry.documentPoint(normalized: vertex, imageRect: media)
            let point = host.convert(overlayPoint, from: host.overlayView)
            #expect(host.overlayView.hitTest(point) === host.overlayView)
        }
        let above = ArtifactRegionGeometry.documentPoint(
            normalized: CGPoint(x: 0.5, y: 0.02), imageRect: media
        )
        #expect(host.overlayView.hitTest(host.convert(above, from: host.overlayView)) == nil)
    }
}
