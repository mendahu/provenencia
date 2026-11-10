import AppKit
import Testing
@testable import Provenencia

@Suite(.serialized, .hostedWindows)
@MainActor
struct GraphCanvasViewportControllerTests {
    private struct Rig {
        let scrollView: NSScrollView
        let document: GraphCanvasDocumentView
        let controller: GraphCanvasViewportController
    }

    private func rig() -> Rig {
        let scrollView = NSScrollView(frame: NSRect(x: 0, y: 0, width: 1_000, height: 800))
        scrollView.scrollerStyle = .overlay
        scrollView.hasVerticalScroller = false
        scrollView.hasHorizontalScroller = false
        scrollView.allowsMagnification = true
        scrollView.minMagnification = 0.1
        scrollView.maxMagnification = 4
        let document = GraphCanvasDocumentView(
            pointer: GraphCanvasPointerController(),
            frame: NSRect(x: 0, y: 0, width: 4_000, height: 4_000)
        )
        scrollView.documentView = document
        hostInBorderlessWindow(nsView: scrollView, size: CGSize(width: 1_000, height: 800))
        scrollView.layoutSubtreeIfNeeded()
        let controller = GraphCanvasViewportController()
        controller.scrollView = scrollView
        return Rig(scrollView: scrollView, document: document, controller: controller)
    }

    private func nextTurn() async {
        await withCheckedContinuation { continuation in
            DispatchQueue.main.async { continuation.resume() }
        }
    }

    @Test func centerOnGridPoint() async {
        let rig = rig()
        let clip = rig.scrollView.contentView

        rig.controller.center(onGridPoint: CGPoint(x: 3_000, y: 1_200))
        await nextTurn()
        #expect(clip.bounds.origin == CGPoint(x: 2_500, y: 800))

        // The document's top-left stands on grid (-500, -500).
        rig.document.contentOrigin = CGPoint(x: -500, y: -500)
        rig.controller.center(onGridPoint: CGPoint(x: 2_500, y: 700))
        await nextTurn()
        #expect(clip.bounds.origin == CGPoint(x: 2_500, y: 800))

        rig.document.contentOrigin = .zero
        rig.scrollView.magnification = 2
        rig.controller.center(onGridPoint: CGPoint(x: 3_000, y: 1_200))
        await nextTurn()
        #expect(clip.bounds.origin == CGPoint(x: 2_750, y: 1_000))
        #expect(rig.scrollView.magnification == 2)

        // A scroll queued ahead of the call: the focus wins.
        DispatchQueue.main.async { clip.scroll(to: .zero) }
        rig.controller.center(onGridPoint: CGPoint(x: 3_000, y: 1_200))
        await nextTurn()
        await nextTurn()
        #expect(clip.bounds.origin == CGPoint(x: 2_750, y: 1_000))
    }
}
