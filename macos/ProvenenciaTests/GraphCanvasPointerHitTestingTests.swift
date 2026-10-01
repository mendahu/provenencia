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

    @Test func actionPrefersNestedFrameInsideTarget() {
        let target = GraphCanvasHitTarget(
            id: "card",
            frame: CGRect(x: 0, y: 0, width: 200, height: 100),
            acceptsConnect: true,
            actions: [
                GraphCanvasActionTarget(
                    id: "edit",
                    frame: CGRect(x: 160, y: 8, width: 28, height: 28)
                ),
            ]
        )
        let hit = GraphCanvasPointerHitTesting.action(
            at: CGPoint(x: 170, y: 20),
            in: target
        )
        #expect(hit?.id == "edit")
        #expect(GraphCanvasPointerHitTesting.action(at: CGPoint(x: 20, y: 20), in: target) == nil)
    }
}

@Suite
@MainActor
struct GraphCanvasPointerPressTests {
    private func controller() -> GraphCanvasPointerController {
        let pointer = GraphCanvasPointerController()
        pointer.hitTargets = [
            GraphCanvasHitTarget(
                id: "card",
                frame: CGRect(x: 0, y: 0, width: 100, height: 100),
                acceptsConnect: true,
                actions: [GraphCanvasActionTarget(id: "promote", frame: CGRect(x: 0, y: 56, width: 100, height: 44))]
            ),
        ]
        return pointer
    }

    @Test func pressOnActionIsHeldUntilRelease() {
        let pointer = controller()
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 70), windowPoint: .zero)
        #expect(pointer.pressedCardAction?.cardID == "card")
        #expect(pointer.pressedCardAction?.actionID == "promote")
        pointer.mouseUp(documentPoint: CGPoint(x: 10, y: 70), windowPoint: .zero)
        #expect(pointer.pressedCardAction == nil)
    }

    @Test func pressOffAnActionIsNotHeld() {
        let pointer = controller()
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 10), windowPoint: .zero)
        #expect(pointer.pressedCardAction == nil)
    }

    @Test func dragClearsThePress() {
        let pointer = controller()
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 70), windowPoint: .zero)
        pointer.mouseDragged(documentPoint: CGPoint(x: 40, y: 70), windowPoint: .zero)
        #expect(pointer.pressedCardAction == nil)
    }
}
