import CoreGraphics
import Foundation

/// Pure edge helpers for canvas documents (card-edge attachment + cubic path).
///
/// Keep catalog / product types out — callers pass rectangles and points.
enum GraphCanvasEdgeGeometry {
    /// Intersection of the ray from `fromCenter` toward `toCenter` with the
    /// boundary of `fromRect`. Falls back to `fromCenter` if the ray is degenerate.
    static func attachmentPoint(fromRect: CGRect, toward toCenter: CGPoint) -> CGPoint {
        let fromCenter = CGPoint(x: fromRect.midX, y: fromRect.midY)
        let dx = toCenter.x - fromCenter.x
        let dy = toCenter.y - fromCenter.y
        if abs(dx) < 0.001, abs(dy) < 0.001 {
            return fromCenter
        }

        let halfW = fromRect.width / 2
        let halfH = fromRect.height / 2
        // Scale so the vector hits the nearer side of the axis-aligned box.
        let sx = abs(dx) < 0.001 ? CGFloat.greatestFiniteMagnitude : halfW / abs(dx)
        let sy = abs(dy) < 0.001 ? CGFloat.greatestFiniteMagnitude : halfH / abs(dy)
        let t = min(sx, sy)
        return CGPoint(x: fromCenter.x + dx * t, y: fromCenter.y + dy * t)
    }

    /// Cubic control points for a gentle horizontal-biased curve between edges.
    static func cubicControls(from: CGPoint, to: CGPoint) -> (CGPoint, CGPoint) {
        let midX = (from.x + to.x) / 2
        return (
            CGPoint(x: midX, y: from.y),
            CGPoint(x: midX, y: to.y)
        )
    }

    /// Cubic path from `from` to `to` using ``cubicControls``.
    static func cubicPath(from: CGPoint, to: CGPoint) -> CGPath {
        let (c1, c2) = cubicControls(from: from, to: to)
        let path = CGMutablePath()
        path.move(to: from)
        path.addCurve(to: to, control1: c1, control2: c2)
        return path
    }

    /// Segment between two card frames: attachment on each rect edge.
    static func segment(
        fromRect: CGRect,
        toRect: CGRect
    ) -> (start: CGPoint, end: CGPoint, path: CGPath) {
        let start = attachmentPoint(fromRect: fromRect, toward: CGPoint(x: toRect.midX, y: toRect.midY))
        let end = attachmentPoint(fromRect: toRect, toward: CGPoint(x: fromRect.midX, y: fromRect.midY))
        return (start, end, cubicPath(from: start, to: end))
    }
}
