@preconcurrency import AppKit
import Observation

/// Owns canvas pointer policy: hit-test, select, drag, empty-canvas pan, place, connect.
///
/// Product hosts publish ``hitTargets`` and tool ``mode``, and wire the callbacks.
/// Live drag offsets live here so SwiftUI paint can follow without owning gestures.
@MainActor
@Observable
final class GraphCanvasPointerController {
    /// Bottom → top document hit rects.
    var hitTargets: [GraphCanvasHitTarget] = []
    /// Exclusive tool mode from the product host.
    var mode: GraphCanvasPointerToolMode = .idle

    /// Mid-gesture card offsets (document space), keyed by hit target id.
    private(set) var offsets: [String: CGSize] = [:]
    /// Pointer is inside the document (hover for place/connect).
    private(set) var pointerInside = false
    /// Idle hover over a nested card action (`cardID`, `actionID`), for paint.
    private(set) var hoveredCardAction: (cardID: String, actionID: String)?

    weak var viewport: GraphCanvasViewportController?

    var onSelect: ((String) -> Void)?
    var onDeselect: (() -> Void)?
    var onDragEnded: ((String, CGSize) -> Void)?
    var onPlace: ((CGPoint) -> Void)?
    var onConnectPick: ((String) -> Void)?
    /// Idle click on a nested card action (edit / Add property); `(targetID, actionID)`.
    var onCardAction: ((String, String) -> Void)?
    /// Document point while hovering in place/connect; `nil` when pointer leaves.
    var onHover: ((CGPoint?) -> Void)?

    private enum Gesture {
        case none
        case pendingClick(start: CGPoint, targetID: String?)
        case draggingCard(id: String, start: CGPoint)
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

    func setPointerInside(_ inside: Bool) {
        pointerInside = inside
        if !inside {
            hoveredCardAction = nil
            onHover?(nil)
        }
    }

    // MARK: - Mouse (document view)

    func mouseDown(documentPoint: CGPoint, windowPoint: CGPoint) {
        _ = windowPoint
        switch mode {
        case .placing:
            // Card under cursor blocks place (click is ignored).
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: documentPoint, in: hitTargets)
            gesture = .pendingClick(start: documentPoint, targetID: hit?.id)
        case .connecting:
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: documentPoint, in: hitTargets)
            if let hit, hit.acceptsConnect {
                gesture = .pendingClick(start: documentPoint, targetID: hit.id)
            } else {
                gesture = .pendingClick(start: documentPoint, targetID: nil)
            }
        case .idle:
            let hit = GraphCanvasPointerHitTesting.topmostTarget(at: documentPoint, in: hitTargets)
            gesture = .pendingClick(start: documentPoint, targetID: hit?.id)
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

        case .pendingClick(let start, let targetID):
            guard GraphCanvasPointerHitTesting.isDrag(from: start, to: documentPoint) else { return }
            if mode == .idle, let targetID {
                gesture = .draggingCard(id: targetID, start: start)
                let offset = CGSize(
                    width: documentPoint.x - start.x,
                    height: documentPoint.y - start.y
                )
                setOffset(targetID, offset: offset)
                onSelect?(targetID)
            } else if mode == .idle, targetID == nil {
                // Empty canvas → click-drag pan.
                gesture = .backgroundPanning(lastWindow: windowPoint)
                pushPanCursor()
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
        switch gesture {
        case .backgroundPanning:
            popPanCursor()
            gesture = .none

        case .pendingClick(_, let targetID):
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
            case .idle:
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
        case .idle:
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
