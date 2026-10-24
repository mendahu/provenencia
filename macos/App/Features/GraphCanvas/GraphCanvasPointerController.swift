@preconcurrency import AppKit
import Observation

/// Owns canvas pointer policy: hit-test, select, drag, empty-canvas pan,
/// selection rectangle, place, connect.
///
/// Every point it takes or reports is in grid space (see the GraphCanvas
/// README, "Coordinate seams"); ``GraphCanvasDocumentView`` converts each event once.
///
/// Product hosts publish ``hitTargets`` and tool ``mode``, and wire the callbacks.
/// Live drag offsets live here; ``GraphCanvasDocumentView`` applies them to item
/// views directly. Observation is per property, so views that only need to know
/// *whether* something is dragging read ``draggingItemID``, which changes at
/// drag start and end — reading ``offsets`` re-runs a view on every pointer event.
@MainActor
@Observable
final class GraphCanvasPointerController {
    /// Grid-space hit rects, in array order for targets that are not items.
    var hitTargets: [GraphCanvasHitTarget] = []
    /// Where the item with an id is drawn among the items (higher is on top);
    /// nil for an id that is not an item. The document view wires it to its
    /// item layer, so a click hits the item drawn on top.
    var stackIndex: (String) -> Int? = { _ in nil }
    /// Exclusive tool mode from the product host.
    var mode: GraphCanvasPointerToolMode = .idle

    /// Mid-gesture item offsets, keyed by hit target id. A drag delta is the
    /// same in grid and document space.
    /// Changes on every pointer event during a drag.
    private(set) var offsets: [String: CGSize] = [:]
    /// The item being dragged (the anchor), if any. Changes only at drag start and end.
    private(set) var draggingItemID: String?
    /// Items that move with the anchor. Empty for a one-item drag. Changes only
    /// at drag start and end.
    private(set) var draggingPeerIDs: Set<String> = []
    /// Idle hover over a nested item action (`itemID`, `actionID`), for paint.
    private(set) var hoveredItemAction: (itemID: String, actionID: String)?
    /// Tool-tip text of the hovered nested item action; nil when none is hovered
    /// or the action has no help. Reads ``hoveredItemAction``, so every path that
    /// clears hover clears it.
    var hoveredHelp: String? {
        guard let hovered = hoveredItemAction,
              let target = hitTargets.first(where: { $0.id == hovered.itemID })
        else { return nil }
        return target.actions.first(where: { $0.id == hovered.actionID })?.help
    }
    /// Idle press held on a nested item action, for pressed paint. Cleared on
    /// release and when the press becomes an item drag.
    private(set) var pressedItemAction: (itemID: String, actionID: String)?
    /// Shift is down, as reported by the document view.
    private(set) var shiftHeld = false
    /// The selection rectangle being drawn, normalized, in grid space.
    /// Nil unless the gesture is a marquee.
    private(set) var marqueeRect: CGRect?

    weak var viewport: GraphCanvasViewportController?

    var onSelect: ((String) -> Void)?
    /// An item drag started on this id. Returns every id that moves, anchor included.
    var onDragBegan: ((String) -> Set<String>)?
    var onDeselect: (() -> Void)?
    var onDragEnded: ((String, CGSize) -> Void)?
    /// Shift-click on a hit target: add it to the selection or take it out.
    /// Replaces ``onSelect`` and the item action for that click.
    var onToggleSelect: ((String) -> Void)?
    /// An empty-canvas rectangle was released: ids of the hit targets it overlaps,
    /// and whether Shift was down at press (add to the selection).
    var onMarqueeEnded: ((Set<String>, Bool) -> Void)?
    /// Shift is down over the canvas, as the document view reports it. The host
    /// decides what it engages.
    var onShiftHeldChanged: ((Bool) -> Void)?
    var onPlace: ((CGPoint) -> Void)?
    var onConnectPick: ((String) -> Void)?
    /// Idle click on a nested item action (an edit pencil); `(targetID, actionID)`.
    var onItemAction: ((String, String) -> Void)?
    /// Grid point while hovering in place/connect; `nil` when pointer leaves.
    var onHover: ((CGPoint?) -> Void)?

    private enum Gesture {
        case none
        case pendingClick(start: CGPoint, targetID: String?, shiftDown: Bool)
        case draggingItem(id: String, start: CGPoint, peers: Set<String>)
        /// Empty-canvas selection rectangle between two grid points.
        case marquee(origin: CGPoint, current: CGPoint, additive: Bool)
        /// Empty-canvas pan (idle only); `lastWindow` for incremental deltas.
        case backgroundPanning(lastWindow: CGPoint)
    }

    private var gesture: Gesture = .none
    private var panCursorPushed = false

    /// One dictionary write for the anchor and every item moving with it.
    /// Assigning each id on its own notifies observers once per item.
    private func replaceDragOffsets(anchor: String, peers: Set<String>, offset: CGSize) {
        var next: [String: CGSize] = [:]
        next.reserveCapacity(peers.count + 1)
        next[anchor] = offset
        for peer in peers {
            next[peer] = offset
        }
        offsets = next
    }

    func clearOffset(_ id: String) {
        offsets[id] = nil
    }

    func clearAllOffsets() {
        offsets = [:]
    }

    private static func normalizedRect(from a: CGPoint, to b: CGPoint) -> CGRect {
        CGRect(
            x: min(a.x, b.x),
            y: min(a.y, b.y),
            width: abs(a.x - b.x),
            height: abs(a.y - b.y)
        )
    }

    func setShiftHeld(_ held: Bool) {
        guard held != shiftHeld else { return }
        shiftHeld = held
        onShiftHeldChanged?(held)
    }

    func setPointerInside(_ inside: Bool) {
        if !inside {
            hoveredItemAction = nil
            onHover?(nil)
        }
    }

    // MARK: - Mouse (document view)

    func mouseDown(gridPoint: CGPoint, windowPoint: CGPoint, shiftDown: Bool = false) {
        _ = windowPoint
        switch mode {
        case .placing:
            // Item under cursor blocks place (click is ignored).
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: gridPoint, in: hitTargets, stackIndex: stackIndex)
            gesture = .pendingClick(start: gridPoint, targetID: hit?.id, shiftDown: shiftDown)
        case .connecting:
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: gridPoint, in: hitTargets, stackIndex: stackIndex)
            if let hit, hit.acceptsConnect {
                gesture = .pendingClick(start: gridPoint, targetID: hit.id, shiftDown: shiftDown)
            } else {
                gesture = .pendingClick(start: gridPoint, targetID: nil, shiftDown: shiftDown)
            }
        case .idle, .selecting:
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: gridPoint, in: hitTargets, stackIndex: stackIndex)
            gesture = .pendingClick(start: gridPoint, targetID: hit?.id, shiftDown: shiftDown)
            if let hit, let action = GraphCanvasPointerHitTesting.action(at: gridPoint, in: hit) {
                pressedItemAction = (hit.id, action.id)
            }
        }
    }

    func mouseDragged(gridPoint: CGPoint, windowPoint: CGPoint) {
        switch gesture {
        case .backgroundPanning(let lastWindow):
            let viewDelta = CGSize(
                width: windowPoint.x - lastWindow.x,
                height: -(windowPoint.y - lastWindow.y)
            )
            viewport?.panByViewDelta(viewDelta)
            gesture = .backgroundPanning(lastWindow: windowPoint)

        case .marquee(let origin, _, let additive):
            gesture = .marquee(origin: origin, current: gridPoint, additive: additive)
            marqueeRect = Self.normalizedRect(from: origin, to: gridPoint)

        case .pendingClick(let start, let targetID, let shiftDown):
            guard GraphCanvasPointerHitTesting.isDrag(from: start, to: gridPoint) else { return }
            let pointsAtItems = mode == .idle || mode == .selecting
            if pointsAtItems, let targetID {
                pressedItemAction = nil
                let peers = (onDragBegan?(targetID) ?? [targetID]).subtracting([targetID])
                gesture = .draggingItem(id: targetID, start: start, peers: peers)
                draggingItemID = targetID
                draggingPeerIDs = peers
                let offset = CGSize(
                    width: gridPoint.x - start.x,
                    height: gridPoint.y - start.y
                )
                replaceDragOffsets(anchor: targetID, peers: peers, offset: offset)
            } else if pointsAtItems, targetID == nil {
                if mode == .selecting || shiftDown {
                    // Select armed, or Shift held: draw a rectangle.
                    gesture = .marquee(origin: start, current: gridPoint, additive: shiftDown)
                    marqueeRect = Self.normalizedRect(from: start, to: gridPoint)
                } else {
                    // Empty canvas → click-drag pan.
                    gesture = .backgroundPanning(lastWindow: windowPoint)
                    pushPanCursor()
                }
            } else {
                // Past threshold with no item drag (place/connect): cancel click.
                gesture = .none
            }

        case .draggingItem(let id, let start, let peers):
            let offset = CGSize(
                width: gridPoint.x - start.x,
                height: gridPoint.y - start.y
            )
            replaceDragOffsets(anchor: id, peers: peers, offset: offset)

        case .none:
            break
        }
    }

    func mouseUp(gridPoint: CGPoint, windowPoint: CGPoint) {
        _ = windowPoint
        pressedItemAction = nil
        switch gesture {
        case .backgroundPanning:
            popPanCursor()
            gesture = .none

        case .marquee(_, _, let additive):
            let rect = marqueeRect
            marqueeRect = nil
            gesture = .none
            if let rect {
                onMarqueeEnded?(
                    GraphCanvasPointerHitTesting.idsIntersecting(rect, in: hitTargets),
                    additive
                )
            }

        case .pendingClick(_, let targetID, let shiftDown):
            defer { gesture = .none }
            switch mode {
            case .placing:
                if targetID == nil {
                    onPlace?(gridPoint)
                }
            case .connecting:
                if let targetID {
                    onConnectPick?(targetID)
                }
            case .idle, .selecting:
                if let targetID, shiftDown {
                    onToggleSelect?(targetID)
                } else if let targetID {
                    let target = hitTargets.first(where: { $0.id == targetID })
                    if let target,
                       let action = GraphCanvasPointerHitTesting.action(
                        at: gridPoint,
                        in: target
                       )
                    {
                        onSelect?(targetID)
                        onItemAction?(targetID, action.id)
                    } else {
                        onSelect?(targetID)
                    }
                } else {
                    onDeselect?()
                }
            }

        case .draggingItem(let id, let start, let peers):
            let delta = CGSize(
                width: gridPoint.x - start.x,
                height: gridPoint.y - start.y
            )
            clearOffset(id)
            for peer in peers { clearOffset(peer) }
            draggingItemID = nil
            draggingPeerIDs = []
            onDragEnded?(id, delta)
            gesture = .none

        case .none:
            break
        }
    }

    func mouseMoved(gridPoint: CGPoint) {
        switch mode {
        case .placing, .connecting:
            hoveredItemAction = nil
            onHover?(gridPoint)
        case .idle, .selecting:
            updateHoveredAction(at: gridPoint)
        }
    }

    func mouseExited() {
        setPointerInside(false)
    }

    func mouseEntered() {
        setPointerInside(true)
    }

    private func updateHoveredAction(at gridPoint: CGPoint) {
        guard let target = GraphCanvasPointerHitTesting.topmostTarget(
            at: gridPoint,
            in: hitTargets,
            stackIndex: stackIndex
        ),
            let action = GraphCanvasPointerHitTesting.action(at: gridPoint, in: target)
        else {
            if hoveredItemAction != nil {
                hoveredItemAction = nil
            }
            return
        }
        let next = (itemID: target.id, actionID: action.id)
        if hoveredItemAction?.itemID != next.itemID || hoveredItemAction?.actionID != next.actionID {
            hoveredItemAction = next
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
