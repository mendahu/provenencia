import CoreGraphics
import Foundation

/// Document-space hit rect for AppKit pointer routing (product-agnostic).
///
/// Array order is bottom → top; the last matching frame wins.
struct GraphCanvasHitTarget: Equatable, Sendable, Identifiable {
    var id: String
    var frame: CGRect
    /// When `false`, connecting mode ignores this target (e.g. bridge cards).
    var acceptsConnect: Bool
}

/// Tool mode for ``GraphCanvasPointerController`` (product maps its own modes here).
enum GraphCanvasPointerToolMode: Equatable, Sendable {
    case idle
    case placing
    case connecting
}

/// Pure hit-test / click-vs-drag helpers (unit-tested).
enum GraphCanvasPointerHitTesting {
    /// Movement (document points) before a press becomes a drag.
    static let dragThreshold: CGFloat = 4

    /// Topmost target whose frame contains `point`, or `nil` on a miss.
    static func topmostTarget(
        at point: CGPoint,
        in targets: [GraphCanvasHitTarget]
    ) -> GraphCanvasHitTarget? {
        for target in targets.reversed() {
            if target.frame.contains(point) {
                return target
            }
        }
        return nil
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
