@preconcurrency import AppKit
import Observation

/// Owns canvas pointer policy: hit-test, select, drag, empty-canvas pan,
/// selection rectangle, place, connect.
///
/// Product hosts publish ``hitTargets`` and tool ``mode``, and wire the callbacks.
/// Live drag offsets live here; ``GraphCanvasDocumentView`` applies them to item
/// views directly. Observation is per property, so views that only need to know
/// *whether* something is dragging read ``draggingItemID``, which changes at
/// drag start and end — reading ``offsets`` re-runs a view on every pointer event.
@MainActor
@Observable
final class GraphCanvasPointerController {
    /// Bottom → top document hit rects.
    var hitTargets: [GraphCanvasHitTarget] = []
    /// Exclusive tool mode from the product host.
    var mode: GraphCanvasPointerToolMode = .idle

    /// Mid-gesture card offsets (document space), keyed by hit target id.
    /// Changes on every pointer event during a drag.
    private(set) var offsets: [String: CGSize] = [:]
    /// The item being dragged, if any. Changes only at drag start and end.
    private(set) var draggingItemID: String?
    /// Pointer is inside the document (hover for place/connect).
    private(set) var pointerInside = false
    /// Idle hover over a nested card action (`cardID`, `actionID`), for paint.
    private(set) var hoveredCardAction: (cardID: String, actionID: String)?
    /// Idle press held on a nested card action, for pressed paint. Cleared on
    /// release and when the press becomes a card drag.
    private(set) var pressedCardAction: (cardID: String, actionID: String)?
    /// The selection rectangle being drawn, normalized, in document space.
    /// Nil unless the gesture is a marquee.
    private(set) var marqueeRect: CGRect?

    weak var viewport: GraphCanvasViewportController?

    var onSelect: ((String) -> Void)?
    var onDeselect: (() -> Void)?
    var onDragEnded: ((String, CGSize) -> Void)?
    /// An empty-canvas rectangle was released: ids of the hit targets it overlaps.
    var onMarqueeEnded: ((Set<String>) -> Void)?
    var onPlace: ((CGPoint) -> Void)?
    var onConnectPick: ((String) -> Void)?
    /// Idle click on a nested card action (edit / Add property); `(targetID, actionID)`.
    var onCardAction: ((String, String) -> Void)?
    /// Document point while hovering in place/connect; `nil` when pointer leaves.
    var onHover: ((CGPoint?) -> Void)?

    private enum Gesture {
        case none
        case pendingClick(start: CGPoint, targetID: String?, shiftDown: Bool)
        case draggingCard(id: String, start: CGPoint)
        /// Empty-canvas selection rectangle between two document points.
        case marquee(origin: CGPoint, current: CGPoint)
        /// Empty-canvas pan (idle only); `lastWindow` for incremental deltas.
        case backgroundPanning(lastWindow: CGPoint)
    }

    private var gesture: Gesture = .none
    private var panCursorPushed = false

    func setOffset(_ id: String, offset: CGSize) {
        offsets[id] = offset
    }

    func clearOffset(_ id: String) {
        offsets[id] = nil
    }

    func clearAllOffsets() {
        offsets = [:]
    }

    /// The document's coordinate origin moved by `delta` (old origin minus new).
    /// Gesture anchors are document points, so they move with it and a drag
    /// already in progress does not jump.
    func shiftActiveDocumentOrigin(by delta: CGSize) {
        switch gesture {
        case .draggingCard(let id, let start):
            gesture = .draggingCard(id: id, start: shifted(start, by: delta))
        case .pendingClick(let start, let targetID, let shiftDown):
            gesture = .pendingClick(
                start: shifted(start, by: delta),
                targetID: targetID,
                shiftDown: shiftDown
            )
        case .marquee(let origin, let current):
            let next = (shifted(origin, by: delta), shifted(current, by: delta))
            gesture = .marquee(origin: next.0, current: next.1)
            marqueeRect = Self.normalizedRect(from: next.0, to: next.1)
        case .backgroundPanning, .none:
            break
        }
    }

    private func shifted(_ point: CGPoint, by delta: CGSize) -> CGPoint {
        CGPoint(x: point.x + delta.width, y: point.y + delta.height)
    }

    private static func normalizedRect(from a: CGPoint, to b: CGPoint) -> CGRect {
        CGRect(
            x: min(a.x, b.x),
            y: min(a.y, b.y),
            width: abs(a.x - b.x),
            height: abs(a.y - b.y)
        )
    }

    func setPointerInside(_ inside: Bool) {
        pointerInside = inside
        if !inside {
            hoveredCardAction = nil
            onHover?(nil)
        }
    }

    // MARK: - Mouse (document view)

    func mouseDown(documentPoint: CGPoint, windowPoint: CGPoint, shiftDown: Bool = false) {
        _ = windowPoint
        switch mode {
        case .placing:
            // Card under cursor blocks place (click is ignored).
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: documentPoint, in: hitTargets)
            gesture = .pendingClick(start: documentPoint, targetID: hit?.id, shiftDown: shiftDown)
        case .connecting:
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: documentPoint, in: hitTargets)
            if let hit, hit.acceptsConnect {
                gesture = .pendingClick(start: documentPoint, targetID: hit.id, shiftDown: shiftDown)
            } else {
                gesture = .pendingClick(start: documentPoint, targetID: nil, shiftDown: shiftDown)
            }
        case .idle, .selecting:
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: documentPoint, in: hitTargets)
            gesture = .pendingClick(start: documentPoint, targetID: hit?.id, shiftDown: shiftDown)
            if let hit, let action = GraphCanvasPointerHitTesting.action(at: documentPoint, in: hit) {
                pressedCardAction = (hit.id, action.id)
            }
        }
    }

    func mouseDragged(documentPoint: CGPoint, windowPoint: CGPoint) {
        switch gesture {
        case .backgroundPanning(let lastWindow):
            let viewDelta = CGSize(
                width: windowPoint.x - lastWindow.x,
                height: -(windowPoint.y - lastWindow.y)
            )
            viewport?.panByViewDelta(viewDelta)
            gesture = .backgroundPanning(lastWindow: windowPoint)

        case .marquee(let origin, _):
            gesture = .marquee(origin: origin, current: documentPoint)
            marqueeRect = Self.normalizedRect(from: origin, to: documentPoint)

        case .pendingClick(let start, let targetID, let shiftDown):
            guard GraphCanvasPointerHitTesting.isDrag(from: start, to: documentPoint) else { return }
            let pointsAtCards = mode == .idle || mode == .selecting
            if pointsAtCards, let targetID {
                pressedCardAction = nil
                gesture = .draggingCard(id: targetID, start: start)
                draggingItemID = targetID
                let offset = CGSize(
                    width: documentPoint.x - start.x,
                    height: documentPoint.y - start.y
                )
                setOffset(targetID, offset: offset)
                onSelect?(targetID)
            } else if pointsAtCards, targetID == nil {
                if mode == .selecting || shiftDown {
                    // Select armed, or Shift held: draw a rectangle.
                    gesture = .marquee(origin: start, current: documentPoint)
                    marqueeRect = Self.normalizedRect(from: start, to: documentPoint)
                } else {
                    // Empty canvas → click-drag pan.
                    gesture = .backgroundPanning(lastWindow: windowPoint)
                    pushPanCursor()
                }
            } else {
                // Past threshold with no card drag (place/connect): cancel click.
                gesture = .none
            }

        case .draggingCard(let id, let start):
            let offset = CGSize(
                width: documentPoint.x - start.x,
                height: documentPoint.y - start.y
            )
            setOffset(id, offset: offset)

        case .none:
            break
        }
    }

    func mouseUp(documentPoint: CGPoint, windowPoint: CGPoint) {
        _ = windowPoint
        pressedCardAction = nil
        switch gesture {
        case .backgroundPanning:
            popPanCursor()
            gesture = .none

        case .marquee:
            let rect = marqueeRect
            marqueeRect = nil
            gesture = .none
            if let rect {
                onMarqueeEnded?(
                    GraphCanvasPointerHitTesting.idsIntersecting(rect, in: hitTargets)
                )
            }

        case .pendingClick(_, let targetID, _):
            defer { gesture = .none }
            switch mode {
            case .placing:
                if targetID == nil {
                    onPlace?(documentPoint)
                }
            case .connecting:
                if let targetID {
                    onConnectPick?(targetID)
                }
            case .idle, .selecting:
                if let targetID {
                    let target = hitTargets.first(where: { $0.id == targetID })
                    if let target,
                       let action = GraphCanvasPointerHitTesting.action(
                        at: documentPoint,
                        in: target
                       )
                    {
                        onSelect?(targetID)
                        onCardAction?(targetID, action.id)
                    } else {
                        onSelect?(targetID)
                    }
                } else {
                    onDeselect?()
                }
            }

        case .draggingCard(let id, let start):
            let delta = CGSize(
                width: documentPoint.x - start.x,
                height: documentPoint.y - start.y
            )
            clearOffset(id)
            draggingItemID = nil
            onDragEnded?(id, delta)
            gesture = .none

        case .none:
            break
        }
    }

    func mouseMoved(documentPoint: CGPoint) {
        switch mode {
        case .placing, .connecting:
            hoveredCardAction = nil
            onHover?(documentPoint)
        case .idle, .selecting:
            updateHoveredAction(at: documentPoint)
        }
    }

    func mouseExited() {
        setPointerInside(false)
    }

    func mouseEntered() {
        setPointerInside(true)
    }

    private func updateHoveredAction(at documentPoint: CGPoint) {
        guard let target = GraphCanvasPointerHitTesting.topmostTarget(
            at: documentPoint,
            in: hitTargets
        ),
            let action = GraphCanvasPointerHitTesting.action(at: documentPoint, in: target)
        else {
            if hoveredCardAction != nil {
                hoveredCardAction = nil
            }
            return
        }
        let next = (cardID: target.id, actionID: action.id)
        if hoveredCardAction?.cardID != next.cardID || hoveredCardAction?.actionID != next.actionID {
            hoveredCardAction = next
        }
    }

    private func pushPanCursor() {
        guard !panCursorPushed else { return }
        NSCursor.closedHand.push()
        panCursorPushed = true
    }

    private func popPanCursor() {
        guard panCursorPushed else { return }
        NSCursor.pop()
        panCursorPushed = false
    }
}
