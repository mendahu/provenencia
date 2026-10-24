import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct GraphCanvasPointerHitTestingTests {
    @Test func shiftReportIgnoresTextFocus() {
        #expect(GraphCanvasDocumentView.reportedShift([.shift], textHasFocus: false))
        #expect(!GraphCanvasDocumentView.reportedShift([.shift], textHasFocus: true))
        #expect(!GraphCanvasDocumentView.reportedShift([], textHasFocus: false))
        #expect(GraphCanvasDocumentView.reportedShift([.shift, .command], textHasFocus: false))
    }

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
            in: [bottom, top],
            stackIndex: { _ in nil }
        )
        #expect(hit?.id == "top")
    }

    private func target(_ id: String) -> GraphCanvasHitTarget {
        GraphCanvasHitTarget(id: id, frame: CGRect(x: 0, y: 0, width: 50, height: 50), acceptsConnect: true)
    }

    @Test func topmostTargetPrefersTheHigherStackIndexOverArrayOrder() {
        let stack = ["first": 1, "second": 0]
        let hit = GraphCanvasPointerHitTesting.topmostTarget(
            at: CGPoint(x: 10, y: 10),
            in: [target("first"), target("second")],
            stackIndex: { stack[$0] }
        )
        #expect(hit?.id == "first")
    }

    @Test func topmostTargetRanksATargetWithNoItemBelowEveryItem() {
        let hit = GraphCanvasPointerHitTesting.topmostTarget(
            at: CGPoint(x: 10, y: 10),
            in: [target("item"), target("base")],
            stackIndex: { $0 == "item" ? 0 : nil }
        )
        #expect(hit?.id == "item")
    }

    @Test func topmostTargetFallsBackToArrayOrderWithoutStackIndexes() {
        let hit = GraphCanvasPointerHitTesting.topmostTarget(
            at: CGPoint(x: 10, y: 10),
            in: [target("under"), target("over")],
            stackIndex: { _ in nil }
        )
        #expect(hit?.id == "over")
    }

    @Test func topmostTargetReturnsNilOnMiss() {
        let target = GraphCanvasHitTarget(
            id: "a",
            frame: CGRect(x: 10, y: 10, width: 20, height: 20),
            acceptsConnect: true
        )
        let hit = GraphCanvasPointerHitTesting.topmostTarget(
            at: CGPoint(x: 0, y: 0),
            in: [target],
            stackIndex: { _ in nil }
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
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 70), windowPoint: .zero)
        #expect(pointer.pressedCardAction?.cardID == "card")
        #expect(pointer.pressedCardAction?.actionID == "promote")
        pointer.mouseUp(gridPoint: CGPoint(x: 10, y: 70), windowPoint: .zero)
        #expect(pointer.pressedCardAction == nil)
    }

    @Test func pressOffAnActionIsNotHeld() {
        let pointer = controller()
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero)
        #expect(pointer.pressedCardAction == nil)
    }

    @Test func dragClearsThePress() {
        let pointer = controller()
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 70), windowPoint: .zero)
        pointer.mouseDragged(gridPoint: CGPoint(x: 40, y: 70), windowPoint: .zero)
        #expect(pointer.pressedCardAction == nil)
    }

    @Test func hoverExposesTargetHelp() {
        let pointer = GraphCanvasPointerController()
        pointer.hitTargets = [
            GraphCanvasHitTarget(
                id: "card",
                frame: CGRect(x: 0, y: 0, width: 100, height: 100),
                acceptsConnect: true,
                actions: [
                    GraphCanvasActionTarget(id: "unpromote", frame: CGRect(x: 60, y: 60, width: 36, height: 36), help: "Unpromote"),
                    GraphCanvasActionTarget(id: "openHandle", frame: CGRect(x: 0, y: 60, width: 59, height: 36)),
                ]
            ),
        ]
        pointer.mouseMoved(gridPoint: CGPoint(x: 70, y: 70))
        #expect(pointer.hoveredHelp == "Unpromote")
        pointer.mouseMoved(gridPoint: CGPoint(x: 10, y: 70))
        #expect(pointer.hoveredHelp == nil)
        pointer.mouseMoved(gridPoint: CGPoint(x: 70, y: 70))
        pointer.mouseExited()
        #expect(pointer.hoveredHelp == nil)
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
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        pointer.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(pointer.marqueeRect == CGRect(x: 10, y: 10, width: 70, height: 80))
        #expect(pointer.draggingItemID == nil)
        pointer.mouseUp(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(selectionChanges == 0)

        // Without Shift the same drag pans, and the selection stays.
        let panning = marqueeController()
        panning.onSelect = { _ in selectionChanges += 1 }
        panning.onDeselect = { selectionChanges += 1 }
        panning.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: false)
        panning.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(panning.marqueeRect == nil)
        panning.mouseUp(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(selectionChanges == 0)

        // A Shift press that starts on a card is a card drag.
        let onCard = marqueeController()
        onCard.mouseDown(gridPoint: CGPoint(x: 220, y: 220), windowPoint: .zero, shiftDown: true)
        onCard.mouseDragged(gridPoint: CGPoint(x: 260, y: 260), windowPoint: .zero)
        #expect(onCard.draggingItemID == "card")
        #expect(onCard.marqueeRect == nil)
    }

    @Test func selectingModeMarqueesWithoutShift() {
        let pointer = marqueeController(mode: .selecting)
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero)
        pointer.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(pointer.marqueeRect != nil)

        // A drag that starts on a card still moves the card.
        let onCard = marqueeController(mode: .selecting)
        onCard.mouseDown(gridPoint: CGPoint(x: 220, y: 220), windowPoint: .zero)
        onCard.mouseDragged(gridPoint: CGPoint(x: 260, y: 260), windowPoint: .zero)
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
        var deliveredAdditive: Bool?
        pointer.onMarqueeEnded = { ids, additive in
            delivered = ids
            deliveredAdditive = additive
        }
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        pointer.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        pointer.mouseUp(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(delivered == ["inside"])
        #expect(deliveredAdditive == true)
        #expect(pointer.marqueeRect == nil)
    }

    @Test func marqueeReportsAdditiveFromShift() {
        var additives: [Bool] = []
        let shifted = marqueeController()
        shifted.onMarqueeEnded = { _, additive in additives.append(additive) }
        shifted.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        shifted.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        shifted.mouseUp(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)

        let plain = marqueeController(mode: .selecting)
        plain.onMarqueeEnded = { _, additive in additives.append(additive) }
        plain.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero)
        plain.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        plain.mouseUp(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        #expect(additives == [true, false])
    }

    @Test func shiftClickOnACardToggles() {
        for mode in [GraphCanvasPointerToolMode.idle, .selecting] {
            var toggled: [String] = []
            var selected: [String] = []
            let pointer = marqueeController(mode: mode)
            pointer.onToggleSelect = { toggled.append($0) }
            pointer.onSelect = { selected.append($0) }
            pointer.mouseDown(gridPoint: CGPoint(x: 220, y: 220), windowPoint: .zero, shiftDown: true)
            pointer.mouseUp(gridPoint: CGPoint(x: 220, y: 220), windowPoint: .zero)
            #expect(toggled == ["card"])
            #expect(selected.isEmpty)
        }
    }

    @Test func shiftClickOnAnActionSkipsTheAction() {
        var toggled = 0
        var selected = 0
        var actions = 0
        let pointer = controller()
        pointer.onToggleSelect = { _ in toggled += 1 }
        pointer.onSelect = { _ in selected += 1 }
        pointer.onCardAction = { _, _ in actions += 1 }
        pointer.mouseDown(gridPoint: CGPoint(x: 50, y: 70), windowPoint: .zero, shiftDown: true)
        pointer.mouseUp(gridPoint: CGPoint(x: 50, y: 70), windowPoint: .zero)
        #expect(toggled == 1)
        #expect(selected == 0)
        #expect(actions == 0)

        pointer.mouseDown(gridPoint: CGPoint(x: 50, y: 70), windowPoint: .zero)
        pointer.mouseUp(gridPoint: CGPoint(x: 50, y: 70), windowPoint: .zero)
        #expect(toggled == 1)
        #expect(selected == 1)
        #expect(actions == 1)
    }

    @Test func emptyClickDeselectsWithOrWithoutShift() {
        var deselects = 0
        var toggles = 0
        let pointer = marqueeController()
        pointer.onDeselect = { deselects += 1 }
        pointer.onToggleSelect = { _ in toggles += 1 }
        for shift in [false, true] {
            pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: shift)
            pointer.mouseUp(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero)
        }
        #expect(deselects == 2)
        #expect(toggles == 0)
    }

    @Test func shiftHeldReportsOnlyChanges() {
        var reported: [Bool] = []
        let pointer = marqueeController()
        pointer.onShiftHeldChanged = { reported.append($0) }
        pointer.setShiftHeld(true)
        pointer.setShiftHeld(true)
        pointer.setShiftHeld(false)
        #expect(reported == [true, false])
        #expect(pointer.shiftHeld == false)
    }

    @MainActor
    @Test func aContentOriginChangeLeavesGestureGridPointsAlone() {
        let pointer = marqueeController()
        let document = GraphCanvasDocumentView(pointer: pointer, frame: CGRect(x: 0, y: 0, width: 600, height: 600))
        pointer.mouseDown(gridPoint: CGPoint(x: 10, y: 10), windowPoint: .zero, shiftDown: true)
        pointer.mouseDragged(gridPoint: CGPoint(x: 80, y: 90), windowPoint: .zero)
        document.contentOrigin = CGPoint(x: 100, y: 0)
        #expect(pointer.marqueeRect == CGRect(x: 10, y: 10, width: 70, height: 80))
        pointer.mouseDragged(gridPoint: CGPoint(x: 90, y: 100), windowPoint: .zero)
        #expect(pointer.marqueeRect == CGRect(x: 10, y: 10, width: 80, height: 90))
        pointer.mouseUp(gridPoint: CGPoint(x: 90, y: 100), windowPoint: .zero)

        // A card drag keeps its anchor too.
        pointer.mouseDown(gridPoint: CGPoint(x: 220, y: 220), windowPoint: .zero)
        pointer.mouseDragged(gridPoint: CGPoint(x: 260, y: 230), windowPoint: .zero)
        document.contentOrigin = CGPoint(x: 300, y: 40)
        pointer.mouseDragged(gridPoint: CGPoint(x: 270, y: 240), windowPoint: .zero)
        #expect(pointer.offsets["card"] == CGSize(width: 50, height: 20))
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
        pointer.mouseDown(gridPoint: CGPoint(x: 220, y: 220), windowPoint: .zero)
        pointer.mouseDragged(gridPoint: CGPoint(x: 260, y: 230), windowPoint: .zero)
        #expect(began == "card")
        #expect(pointer.draggingItemID == "card")
        #expect(pointer.draggingPeerIDs == ["peer"])
        #expect(pointer.offsets["card"] == CGSize(width: 40, height: 10))
        #expect(pointer.offsets["peer"] == pointer.offsets["card"])
        pointer.mouseUp(gridPoint: CGPoint(x: 260, y: 230), windowPoint: .zero)
        #expect(pointer.offsets["card"] == nil)
        #expect(pointer.offsets["peer"] == nil)
        #expect(pointer.draggingPeerIDs.isEmpty)
        #expect(ended?.0 == "card")
        #expect(ended?.1 == CGSize(width: 40, height: 10))
    }
}
