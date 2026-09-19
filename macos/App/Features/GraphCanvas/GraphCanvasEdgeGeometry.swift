import CoreGraphics
import Foundation

/// Pure edge helpers for canvas documents (card-edge attachment + cubic path).
///
/// Keep catalog / product types out — callers pass rectangles and points.
/// Endpoints tuck under the card fill along the hit-edge normal so diagonal
/// approaches still end inside the chrome (not merely toward the rect center).
enum GraphCanvasEdgeGeometry {
    /// Inward distance from the AABB edge under the fill.
    static let endpointTuck: CGFloat = 16

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
        let sx = abs(dx) < 0.001 ? CGFloat.greatestFiniteMagnitude : halfW / abs(dx)
        let sy = abs(dy) < 0.001 ? CGFloat.greatestFiniteMagnitude : halfH / abs(dy)
        let t = min(sx, sy)
        return CGPoint(x: fromCenter.x + dx * t, y: fromCenter.y + dy * t)
    }

    /// Moves `point` toward `target` by `distance` (clamped so it never passes target).
    static func moveToward(_ point: CGPoint, target: CGPoint, distance: CGFloat) -> CGPoint {
        let dx = target.x - point.x
        let dy = target.y - point.y
        let length = hypot(dx, dy)
        guard length > 0.001 else { return point }
        let t = min(distance, length) / length
        return CGPoint(x: point.x + dx * t, y: point.y + dy * t)
    }

    /// Pulls an edge hit straight inward along the rect normal so Y tuck is not
    /// diluted on diagonal approaches. Corner hits tuck on both axes.
    static func tuckInside(_ point: CGPoint, rect: CGRect, distance: CGFloat = endpointTuck) -> CGPoint {
        let eps: CGFloat = 1.5
        let onTop = abs(point.y - rect.minY) <= eps
        let onBottom = abs(point.y - rect.maxY) <= eps
        let onLeft = abs(point.x - rect.minX) <= eps
        let onRight = abs(point.x - rect.maxX) <= eps

        var x = point.x
        var y = point.y
        if onTop { y = min(y + distance, rect.midY) }
        if onBottom { y = max(y - distance, rect.midY) }
        if onLeft { x = min(x + distance, rect.midX) }
        if onRight { x = max(x - distance, rect.midX) }

        if onTop || onBottom || onLeft || onRight {
            return CGPoint(x: x, y: y)
        }
        return moveToward(point, target: CGPoint(x: rect.midX, y: rect.midY), distance: distance)
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

    /// Segment between two card frames: attach on each rect edge, tuck under
    /// the fill along the edge normal, then draw a cubic.
    ///
    /// `start` lies on `fromRect`, `end` on `toRect`.
    static func segment(
        fromRect: CGRect,
        toRect: CGRect,
        tuck: CGFloat = endpointTuck
    ) -> (start: CGPoint, end: CGPoint, path: CGPath) {
        let fromCenter = CGPoint(x: fromRect.midX, y: fromRect.midY)
        let toCenter = CGPoint(x: toRect.midX, y: toRect.midY)
        let edgeStart = attachmentPoint(fromRect: fromRect, toward: toCenter)
        let edgeEnd = attachmentPoint(fromRect: toRect, toward: fromCenter)
        let start = tuckInside(edgeStart, rect: fromRect, distance: tuck)
        let end = tuckInside(edgeEnd, rect: toRect, distance: tuck)
        return (start, end, cubicPath(from: start, to: end))
    }
}
