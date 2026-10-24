import CoreGraphics
import Foundation

/// Nested clickable region inside a card hit target (edit pencil, Add property).
struct GraphCanvasActionTarget: Equatable, Sendable, Identifiable {
    var id: String
    var frame: CGRect
    /// Tool-tip text shown while the pointer is over this action; nil shows none.
    var help: String? = nil

    /// A copy with only `frame` transformed (e.g. into document space), so
    /// every other field survives the mapping.
    func mapFrame(_ transform: (CGRect) -> CGRect) -> GraphCanvasActionTarget {
        var copy = self
        copy.frame = transform(frame)
        return copy
    }
}

/// Document-space hit rect for AppKit pointer routing (product-agnostic).
///
/// Among overlapping targets the one drawn on top wins: the highest item stack
/// index, then array order for targets that are not items.
struct GraphCanvasHitTarget: Equatable, Sendable, Identifiable {
    var id: String
    var frame: CGRect
    /// When `false`, connecting mode ignores this target (e.g. bridge cards).
    var acceptsConnect: Bool
    /// Document-space action zones; idle clicks prefer these over select.
    var actions: [GraphCanvasActionTarget] = []
}

/// Tool mode for ``GraphCanvasPointerController`` (product maps its own modes here).
enum GraphCanvasPointerToolMode: Equatable, Sendable {
    case idle
    case placing
    case connecting
    /// Card click, drag, and hover match ``idle``; an empty-canvas drag draws a
    /// selection rectangle instead of panning.
    case selecting
}

/// Pure hit-test / click-vs-drag helpers (unit-tested).
enum GraphCanvasPointerHitTesting {
    /// Movement (document points) before a press becomes a drag.
    static let dragThreshold: CGFloat = 4

    /// Topmost target whose frame contains `point`, or `nil` on a miss.
    ///
    /// `stackIndex` is where the item with that id is drawn (higher is on top).
    /// A target with no stack index is base-layer content, under every item;
    /// ties go to the later target in the array.
    static func topmostTarget(
        at point: CGPoint,
        in targets: [GraphCanvasHitTarget],
        stackIndex: (String) -> Int?
    ) -> GraphCanvasHitTarget? {
        var best: (target: GraphCanvasHitTarget, rank: Int)?
        for target in targets where target.frame.contains(point) {
            let rank = stackIndex(target.id) ?? -1
            if best == nil || rank >= best!.rank {
                best = (target, rank)
            }
        }
        return best?.target
    }

    /// First action whose frame contains `point` (topmost action wins).
    static func action(
        at point: CGPoint,
        in target: GraphCanvasHitTarget
    ) -> GraphCanvasActionTarget? {
        for action in target.actions.reversed() {
            if action.frame.contains(point) {
                return action
            }
        }
        return nil
    }

    /// Ids whose frame overlaps `rect`. A shared edge is not a hit.
    static func idsIntersecting(
        _ rect: CGRect,
        in targets: [GraphCanvasHitTarget]
    ) -> Set<String> {
        Set(targets.filter { $0.frame.intersects(rect) }.map(\.id))
    }

    /// `true` when the pointer has moved at least `threshold` from `start`.
    static func isDrag(
        from start: CGPoint,
        to current: CGPoint,
        threshold: CGFloat = dragThreshold
    ) -> Bool {
        hypot(current.x - start.x, current.y - start.y) >= threshold
    }
}
