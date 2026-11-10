import AppKit
import PDFKit
import Testing
@testable import Provenencia

@Suite(.hostedWindows)
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
        let (host, _, media) = makeHost(committed: Self.triangle)
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
        #expect(overlay.moveAnchorHitRadius == 16)
        overlay.magnification = 1
        #expect(overlay.handleHitRadius == 7)
        #expect(overlay.handleDrawSize == 7)
        #expect(overlay.moveAnchorHitRadius == 8)
    }

    // MARK: Freeform close distance

    /// A 3000×4000 image drawn at 1:1 on an armed freeform overlay, at `magnification`.
    private func makeArmedFreeform(
        magnification: CGFloat = 1
    ) -> (overlay: ArtifactRegionOverlayView, window: NSWindow, media: CGRect) {
        let overlay = ArtifactRegionOverlayView(frame: CGRect(x: 0, y: 0, width: 400, height: 500))
        let media = CGRect(x: 0, y: 0, width: 3000, height: 4000)
        overlay.mediaRectOverride = media
        overlay.magnification = magnification
        let window = hostInBorderlessWindow(nsView: overlay, size: CGSize(width: 400, height: 500))
        overlay.armedTool = .freeform
        return (overlay, window, media)
    }

    /// A click at `document` (image pixels, overlay space).
    private func click(_ document: CGPoint, on overlay: ArtifactRegionOverlayView, window: NSWindow) {
        overlay.mouseDown(with: NSEvent.mouseEvent(
            with: .leftMouseDown,
            location: overlay.convert(document, to: nil),
            modifierFlags: [],
            timestamp: 0,
            windowNumber: window.windowNumber,
            context: nil,
            eventNumber: 0,
            clickCount: 1,
            pressure: 1
        )!)
    }

    private static let fourDocumentPoints = [
        CGPoint(x: 300, y: 300), CGPoint(x: 900, y: 300), CGPoint(x: 900, y: 900), CGPoint(x: 300, y: 360),
    ]

    @Test func freeformFourthClickNearFirstAddsVertex() {
        let (overlay, window, media) = makeArmedFreeform()
        var commits: [ArtifactRegionDraft] = []
        overlay.onCommit = { commits.append($0); return true }

        for point in Self.fourDocumentPoints { click(point, on: overlay, window: window) }
        #expect(commits.isEmpty)

        // A fifth click within a handle's grab radius of the first closes the four points.
        click(CGPoint(x: 303, y: 300), on: overlay, window: window)
        #expect(commits.count == 1)
        #expect(commits.first?.points.count == 4)
        #expect(commits.first?.points.first
            == ArtifactRegionGeometry.normalize(documentPoint: Self.fourDocumentPoints[0], imageRect: media))
    }

    @Test func freeformCloseZoneIsHandleRadiusAtAnyZoom() {
        for (offset, closes) in [(CGFloat(1.5), true), (CGFloat(2), false)] {
            let (overlay, window, _) = makeArmedFreeform(magnification: 4)
            var commits: [ArtifactRegionDraft] = []
            overlay.onCommit = { commits.append($0); return true }

            for point in Self.fourDocumentPoints.prefix(3) { click(point, on: overlay, window: window) }
            click(CGPoint(x: 300 + offset, y: 300), on: overlay, window: window)
            // 6 screen points away closes the triangle; 8 adds a fourth vertex.
            #expect(commits.count == (closes ? 1 : 0))
            #expect(commits.first?.points.count == (closes ? 3 : nil))
        }
    }

    private static let square = [
        CGPoint(x: 0.2, y: 0.2), CGPoint(x: 0.4, y: 0.2), CGPoint(x: 0.4, y: 0.4), CGPoint(x: 0.2, y: 0.4),
    ]

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
        var commits: [ArtifactRegionDraft] = []
        host.overlayView.onCommit = { commits.append($0); return true }

        drag(from: CGPoint(x: 0.5, y: 0.9), to: CGPoint(x: 0.6, y: 0.7), in: host, window: window)
        #expect(commits.count == 1)
        #expect(commits.last.map { ArtifactRegionGeometry.nearlyEqual($0.points[2], CGPoint(x: 0.6, y: 0.7), epsilon: 1e-4) } == true)
        #expect(commits.last.map { ArtifactRegionGeometry.nearlyEqual($0.points[0], Self.triangle[0], epsilon: 1e-4)
                && ArtifactRegionGeometry.nearlyEqual($0.points[1], Self.triangle[1], epsilon: 1e-4) } == true)

        host.overlayView.committed = ArtifactRegionDraft(kind: .freeform, points: Self.square)
        commits = []
        drag(from: CGPoint(x: 0.3, y: 0.2), to: CGPoint(x: 0.4, y: 0.25), in: host, window: window)
        #expect(commits.count == 1)
        for (moved, original) in zip(commits.last?.points ?? [], Self.square) {
            #expect(ArtifactRegionGeometry.nearlyEqual(
                moved, CGPoint(x: original.x + 0.1, y: original.y + 0.05), epsilon: 1e-4
            ))
        }

        host.overlayView.committed = ArtifactRegionDraft(kind: .freeform, points: Self.square)
        commits = []
        drag(from: CGPoint(x: 0.2, y: 0.2), to: CGPoint(x: 0.2, y: 0.2), in: host, window: window)
        #expect(commits.isEmpty)
    }

    @Test func invalidEditRevertsAndDoesNotCommit() {
        let (host, window, _) = makeHost(committed: Self.square)
        let overlay = host.overlayView
        var commits: [ArtifactRegionDraft] = []
        overlay.onCommit = { commits.append($0); return true }

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

    @Test func refusedCommitRevertsEdit() {
        let (host, window, _) = makeHost(committed: Self.triangle)
        let overlay = host.overlayView
        var commits: [ArtifactRegionDraft] = []
        overlay.onCommit = { commits.append($0); return false }

        drag(from: CGPoint(x: 0.5, y: 0.9), to: CGPoint(x: 0.6, y: 0.7), in: host, window: window)
        #expect(commits.count == 1)
        #expect(overlay.committed?.points == Self.triangle)

        let five = [
            CGPoint(x: 0.1, y: 0.1), CGPoint(x: 0.9, y: 0.1), CGPoint(x: 0.9, y: 0.9),
            CGPoint(x: 0.5, y: 0.9), CGPoint(x: 0.1, y: 0.9),
        ]
        overlay.committed = ArtifactRegionDraft(kind: .freeform, points: five)
        commits = []
        overlay.rightMouseDown(with: event(.rightMouseDown, at: five[3], in: host, window: window))
        #expect(commits.count == 1)
        #expect(overlay.committed?.points == five)
    }

    @Test func subThresholdDragDoesNotEdit() {
        let (host, window, media) = makeHost(committed: Self.triangle)
        let overlay = host.overlayView
        var commits: [ArtifactRegionDraft] = []
        overlay.onCommit = { commits.append($0); return true }
        let jitter = 2 / media.width

        drag(from: Self.triangle[2], to: CGPoint(x: Self.triangle[2].x + jitter, y: Self.triangle[2].y), in: host, window: window)
        #expect(commits.isEmpty)
        #expect(overlay.committed?.points == Self.triangle)

        let onEdge = CGPoint(x: 0.5, y: 0.1)
        drag(from: onEdge, to: CGPoint(x: onEdge.x + jitter, y: onEdge.y), in: host, window: window)
        #expect(commits.isEmpty)
        #expect(overlay.committed?.points == Self.triangle)
    }

    @Test func vertexKeepsGrabOffset() {
        let (host, window, media) = makeHost(committed: Self.triangle)
        let overlay = host.overlayView
        overlay.onCommit = { _ in true }
        let vertex = Self.triangle[2]
        let grab = CGPoint(x: vertex.x + 5 / media.width, y: vertex.y)
        let release = CGPoint(x: grab.x + 20 / media.width, y: grab.y)

        drag(from: grab, to: release, in: host, window: window)
        let moved = overlay.committed?.points[2] ?? .zero
        #expect(ArtifactRegionGeometry.nearlyEqual(
            moved, CGPoint(x: vertex.x + 20 / media.width, y: vertex.y), epsilon: 1e-4
        ))
    }
}
