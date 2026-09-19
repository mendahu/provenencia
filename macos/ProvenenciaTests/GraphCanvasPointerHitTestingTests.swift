import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasPointerHitTestingTests {
    @Test func topmostTargetPrefersLastMatchingRect() {
        let bottom = GraphCanvasHitTarget(
            id: "bottom",
            frame: CGRect(x: 0, y: 0, width: 100, height: 100),
            acceptsConnect: true
        )
        let top = GraphCanvasHitTarget(
            id: "top",
            frame: CGRect(x: 20, y: 20, width: 40, height: 40),
            acceptsConnect: true
        )
        let hit = GraphCanvasPointerHitTesting.topmostTarget(
            at: CGPoint(x: 30, y: 30),
            in: [bottom, top]
        )
        #expect(hit?.id == "top")
    }

    @Test func topmostTargetReturnsNilOnMiss() {
        let target = GraphCanvasHitTarget(
            id: "a",
            frame: CGRect(x: 10, y: 10, width: 20, height: 20),
            acceptsConnect: true
        )
        let hit = GraphCanvasPointerHitTesting.topmostTarget(
            at: CGPoint(x: 0, y: 0),
            in: [target]
        )
        #expect(hit == nil)
    }

    @Test func isDragRespectsThreshold() {
        let start = CGPoint(x: 0, y: 0)
        #expect(!GraphCanvasPointerHitTesting.isDrag(from: start, to: CGPoint(x: 3, y: 0)))
        #expect(GraphCanvasPointerHitTesting.isDrag(from: start, to: CGPoint(x: 4, y: 0)))
        #expect(GraphCanvasPointerHitTesting.isDrag(from: start, to: CGPoint(x: 0, y: 5)))
    }
}
