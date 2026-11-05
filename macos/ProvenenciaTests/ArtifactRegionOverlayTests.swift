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

    @Test func handleSizesAreInScreenPoints() {
        let overlay = ArtifactRegionOverlayView(frame: .zero)
        overlay.magnification = 0.5
        #expect(overlay.handleHitRadius == 14)
        #expect(overlay.handleDrawSize == 14)
        overlay.magnification = 1
        #expect(overlay.handleHitRadius == 7)
        #expect(overlay.handleDrawSize == 7)
    }

    private static let square = [
        CGPoint(x: 0.2, y: 0.2), CGPoint(x: 0.4, y: 0.2), CGPoint(x: 0.4, y: 0.4), CGPoint(x: 0.2, y: 0.4),
    ]

    private func near(_ a: CGPoint, _ b: CGPoint) -> Bool {
        abs(a.x - b.x) < 1e-4 && abs(a.y - b.y) < 1e-4
    }

    /// One drag: down at `from`, dragged and released at `to`.
    private func drag(
        from: CGPoint,
        to: CGPoint,
        in host: ArtifactPDFHostView,
        window: NSWindow
    ) {
        let overlay = host.overlayView
        overlay.mouseDown(with: event(.leftMouseDown, at: from, in: host, window: window))
        overlay.mouseDragged(with: event(.leftMouseDragged, at: to, in: host, window: window))
        overlay.mouseUp(with: event(.leftMouseUp, at: to, in: host, window: window))
    }

    @Test func vertexDragCommitsOnce() {
        let (host, window, _) = makeHost(committed: Self.triangle)
        defer { withExtendedLifetime(window) {} }
        var commits: [ArtifactRegionDraft] = []
        host.overlayView.onCommit = { commits.append($0) }

        drag(from: CGPoint(x: 0.5, y: 0.9), to: CGPoint(x: 0.6, y: 0.7), in: host, window: window)
        #expect(commits.count == 1)
        #expect(commits.last.map { near($0.points[2], CGPoint(x: 0.6, y: 0.7)) } == true)
        #expect(commits.last.map { near($0.points[0], Self.triangle[0]) && near($0.points[1], Self.triangle[1]) } == true)

        host.overlayView.committed = ArtifactRegionDraft(kind: .freeform, points: Self.square)
        commits = []
        drag(from: CGPoint(x: 0.3, y: 0.2), to: CGPoint(x: 0.4, y: 0.25), in: host, window: window)
        #expect(commits.count == 1)
        for (moved, original) in zip(commits.last?.points ?? [], Self.square) {
            #expect(near(moved, CGPoint(x: original.x + 0.1, y: original.y + 0.05)))
        }

        host.overlayView.committed = ArtifactRegionDraft(kind: .freeform, points: Self.square)
        commits = []
        drag(from: CGPoint(x: 0.2, y: 0.2), to: CGPoint(x: 0.2, y: 0.2), in: host, window: window)
        #expect(commits.isEmpty)
    }

    @Test func invalidEditRevertsAndDoesNotCommit() {
        let (host, window, _) = makeHost(committed: Self.square)
        defer { withExtendedLifetime(window) {} }
        let overlay = host.overlayView
        var commits: [ArtifactRegionDraft] = []
        overlay.onCommit = { commits.append($0) }

        drag(from: CGPoint(x: 0.2, y: 0.2), to: CGPoint(x: 0.5, y: 0.3), in: host, window: window)
        #expect(overlay.committed?.points == Self.square)
        #expect(commits.isEmpty)

        let five = [
            CGPoint(x: 0.1, y: 0.1), CGPoint(x: 0.9, y: 0.1), CGPoint(x: 0.9, y: 0.9),
            CGPoint(x: 0.5, y: 0.2), CGPoint(x: 0.1, y: 0.9),
        ]
        overlay.committed = ArtifactRegionDraft(kind: .freeform, points: five)
        overlay.rightMouseDown(with: event(.rightMouseDown, at: five[0], in: host, window: window))
        #expect(overlay.committed?.points == five)
        #expect(commits.isEmpty)

        let squareWithSpur = Self.square + [CGPoint(x: 0.1, y: 0.3)]
        overlay.committed = ArtifactRegionDraft(kind: .freeform, points: squareWithSpur)
        overlay.rightMouseDown(with: event(.rightMouseDown, at: squareWithSpur[4], in: host, window: window))
        #expect(commits.count == 1)
        #expect(commits.last?.points == Self.square)
    }
}
