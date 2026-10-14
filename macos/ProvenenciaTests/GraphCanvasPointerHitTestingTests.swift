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

    @Test func idsIntersectingUsesCardFrames() {
        func target(_ id: String, _ frame: CGRect) -> GraphCanvasHitTarget {
            GraphCanvasHitTarget(id: id, frame: frame, acceptsConnect: true)
        }
        let rect = CGRect(x: 0, y: 0, width: 100, height: 100)
        let ids = GraphCanvasPointerHitTesting.idsIntersecting(
            rect,
            in: [
                target("p1", CGRect(x: 10, y: 10, width: 20, height: 20)),
                target("p2", CGRect(x: 50, y: 50, width: 20, height: 20)),
                target("p3", CGRect(x: 200, y: 0, width: 20, height: 20)),
                // Shares the rectangle's right edge: not a hit.
                target("edge", CGRect(x: 100, y: 0, width: 20, height: 20)),
            ]
        )
        #expect(ids == ["p1", "p2"])
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

    // MARK: Marquee (S10-20)

    private func marqueeController(mode: GraphCanvasPointerToolMode = .idle) -> GraphCanvasPointerController {
        let pointer = GraphCanvasPointerController()
        pointer.mode = mode
        pointer.hitTargets = [
            GraphCanvasHitTarget(
                id: "card",
                frame: CGRect(x: 200, y: 200, width: 100, height: 80),
                acceptsConnect: true
            ),
        ]
        return pointer
    }

    @Test func shiftDragOnEmptyCanvasMarquees() {
        var selectionChanges = 0
        let pointer = marqueeController()
        pointer.onSelect = { _ in selectionChanges += 1 }
        pointer.onDeselect = { selectionChanges += 1 }
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        pointer.mouseDragged(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(pointer.marqueeRect == CGRect(x: 10, y: 10, width: 70, height: 80))
        #expect(pointer.draggingItemID == nil)
        pointer.mouseUp(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(selectionChanges == 0)

        // Without Shift the same drag pans, and the selection stays.
        let panning = marqueeController()
        panning.onSelect = { _ in selectionChanges += 1 }
        panning.onDeselect = { selectionChanges += 1 }
        panning.mouseDown(documentPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: false)
        panning.mouseDragged(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(panning.marqueeRect == nil)
        panning.mouseUp(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(selectionChanges == 0)

        // A Shift press that starts on a card is a card drag.
        let onCard = marqueeController()
        onCard.mouseDown(documentPoint: CGPoint(x: 220, y: 220), windowPoint: .zero, shiftDown: true)
        onCard.mouseDragged(documentPoint: CGPoint(x: 260, y: 260), windowPoint: .zero)
        #expect(onCard.draggingItemID == "card")
        #expect(onCard.marqueeRect == nil)
    }

    @Test func selectingModeMarqueesWithoutShift() {
        let pointer = marqueeController(mode: .selecting)
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 10), windowPoint: .zero)
        pointer.mouseDragged(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(pointer.marqueeRect != nil)

        // A drag that starts on a card still moves the card.
        let onCard = marqueeController(mode: .selecting)
        onCard.mouseDown(documentPoint: CGPoint(x: 220, y: 220), windowPoint: .zero)
        onCard.mouseDragged(documentPoint: CGPoint(x: 260, y: 260), windowPoint: .zero)
        #expect(onCard.draggingItemID == "card")
        #expect(onCard.marqueeRect == nil)
    }

    @Test func marqueeReleaseClearsTheRect() {
        let pointer = marqueeController()
        pointer.hitTargets.append(
            GraphCanvasHitTarget(
                id: "inside",
                frame: CGRect(x: 20, y: 20, width: 30, height: 30),
                acceptsConnect: true
            )
        )
        var delivered: Set<String>?
        pointer.onMarqueeEnded = { delivered = $0 }
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        pointer.mouseDragged(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        pointer.mouseUp(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(delivered == ["inside"])
        #expect(pointer.marqueeRect == nil)
    }

    @Test func marqueeFollowsADocumentOriginShift() {
        let pointer = marqueeController()
        pointer.mouseDown(documentPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        pointer.mouseDragged(documentPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        pointer.shiftActiveDocumentOrigin(by: CGSize(width: 100, height: 0))
        #expect(pointer.marqueeRect == CGRect(x: 110, y: 10, width: 70, height: 80))
    }

    @Test func dragPeersShareTheAnchorOffset() {
        let pointer = marqueeController()
        pointer.hitTargets.append(
            GraphCanvasHitTarget(
                id: "peer",
                frame: CGRect(x: 400, y: 200, width: 100, height: 80),
                acceptsConnect: true
            )
        )
        var began: String?
        pointer.onDragBegan = { id in
            began = id
            return [id, "peer"]
        }
        var ended: (String, CGSize)?
        pointer.onDragEnded = { ended = ($0, $1) }
        pointer.mouseDown(documentPoint: CGPoint(x: 220, y: 220), windowPoint: .zero)
        pointer.mouseDragged(documentPoint: CGPoint(x: 260, y: 230), windowPoint: .zero)
        #expect(began == "card")
        #expect(pointer.draggingItemID == "card")
        #expect(pointer.draggingPeerIDs == ["peer"])
        #expect(pointer.offsets["card"] == CGSize(width: 40, height: 10))
        #expect(pointer.offsets["peer"] == pointer.offsets["card"])
        pointer.mouseUp(documentPoint: CGPoint(x: 260, y: 230), windowPoint: .zero)
        #expect(pointer.offsets["card"] == nil)
        #expect(pointer.offsets["peer"] == nil)
        #expect(pointer.draggingPeerIDs.isEmpty)
        #expect(ended?.0 == "card")
        #expect(ended?.1 == CGSize(width: 40, height: 10))
    }
}
