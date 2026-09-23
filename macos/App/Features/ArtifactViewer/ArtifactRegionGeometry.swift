import CoreGraphics
import Foundation

/// Normalized (0…1) region geometry for the raster Artifact overlay.
enum ArtifactRegionGeometry {
    static let circleRingCount = 32
    static let minimumPolygonArea: CGFloat = 1e-12
    static let freeformMinimumVertices = 3

    // MARK: Image rect + normalize

    /// Media rect inside the document pad (centering inset is not the image).
    static func imageRect(documentSize: CGSize, imageSize: CGSize) -> CGRect {
        guard imageSize.width > 0, imageSize.height > 0 else { return .zero }
        return CGRect(
            x: (documentSize.width - imageSize.width) / 2,
            y: (documentSize.height - imageSize.height) / 2,
            width: imageSize.width,
            height: imageSize.height
        )
    }

    static func clampNormalized(_ point: CGPoint) -> CGPoint {
        CGPoint(
            x: min(max(point.x, 0), 1),
            y: min(max(point.y, 0), 1)
        )
    }

    static func normalize(documentPoint: CGPoint, imageRect: CGRect) -> CGPoint {
        guard imageRect.width > 0, imageRect.height > 0 else { return .zero }
        return clampNormalized(
            CGPoint(
                x: (documentPoint.x - imageRect.minX) / imageRect.width,
                y: (documentPoint.y - imageRect.minY) / imageRect.height
            )
        )
    }

    static func documentPoint(normalized: CGPoint, imageRect: CGRect) -> CGPoint {
        CGPoint(
            x: imageRect.minX + clampNormalized(normalized).x * imageRect.width,
            y: imageRect.minY + clampNormalized(normalized).y * imageRect.height
        )
    }

    static func normalizeAll(_ points: [CGPoint], imageRect: CGRect) -> [CGPoint] {
        points.map { normalize(documentPoint: $0, imageRect: imageRect) }
    }

    static func documentPoints(normalized: [CGPoint], imageRect: CGRect) -> [CGPoint] {
        normalized.map { documentPoint(normalized: $0, imageRect: imageRect) }
    }

    // MARK: Shape builders (normalized)

    static func rectanglePoints(from a: CGPoint, to b: CGPoint) -> [CGPoint] {
        let minX = min(a.x, b.x)
        let maxX = max(a.x, b.x)
        let minY = min(a.y, b.y)
        let maxY = max(a.y, b.y)
        return [
            clampNormalized(CGPoint(x: minX, y: minY)),
            clampNormalized(CGPoint(x: maxX, y: minY)),
            clampNormalized(CGPoint(x: maxX, y: maxY)),
            clampNormalized(CGPoint(x: minX, y: maxY)),
        ]
    }

    static func lShapePoints(
        kind: ArtifactRegionKind,
        from a: CGPoint,
        to b: CGPoint
    ) -> [CGPoint] {
        let minX = min(a.x, b.x)
        let maxX = max(a.x, b.x)
        let minY = min(a.y, b.y)
        let maxY = max(a.y, b.y)
        let midX = (minX + maxX) / 2
        let midY = (minY + maxY) / 2
        let raw: [CGPoint]
        switch kind {
        case .lOpenTopRight:
            raw = [
                CGPoint(x: minX, y: minY),
                CGPoint(x: midX, y: minY),
                CGPoint(x: midX, y: midY),
                CGPoint(x: maxX, y: midY),
                CGPoint(x: maxX, y: maxY),
                CGPoint(x: minX, y: maxY),
            ]
        case .lOpenTopLeft:
            raw = [
                CGPoint(x: midX, y: minY),
                CGPoint(x: maxX, y: minY),
                CGPoint(x: maxX, y: maxY),
                CGPoint(x: minX, y: maxY),
                CGPoint(x: minX, y: midY),
                CGPoint(x: midX, y: midY),
            ]
        case .lOpenBottomRight:
            raw = [
                CGPoint(x: minX, y: minY),
                CGPoint(x: maxX, y: minY),
                CGPoint(x: maxX, y: midY),
                CGPoint(x: midX, y: midY),
                CGPoint(x: midX, y: maxY),
                CGPoint(x: minX, y: maxY),
            ]
        case .lOpenBottomLeft:
            raw = [
                CGPoint(x: minX, y: minY),
                CGPoint(x: maxX, y: minY),
                CGPoint(x: maxX, y: maxY),
                CGPoint(x: midX, y: maxY),
                CGPoint(x: midX, y: midY),
                CGPoint(x: minX, y: midY),
            ]
        default:
            return rectanglePoints(from: a, to: b)
        }
        return raw.map(clampNormalized)
    }

    static func circleRing(center: CGPoint, radius: CGFloat) -> [CGPoint] {
        let c = clampNormalized(center)
        let maxR = min(c.x, 1 - c.x, c.y, 1 - c.y)
        let r = min(max(radius, 0), maxR)
        guard r > 0 else { return [] }
        return (0..<circleRingCount).map { index in
            let angle = CGFloat(index) / CGFloat(circleRingCount) * 2 * .pi
            return clampNormalized(
                CGPoint(
                    x: c.x + r * cos(angle),
                    y: c.y + r * sin(angle)
                )
            )
        }
    }

    static func circleRing(from a: CGPoint, to b: CGPoint) -> [CGPoint] {
        let dx = b.x - a.x
        let dy = b.y - a.y
        let radius = sqrt(dx * dx + dy * dy)
        return circleRing(center: a, radius: radius)
    }

    // MARK: Constrained edit

    static func moveRectangleHandle(
        points: [CGPoint],
        handle: Int,
        to location: CGPoint
    ) -> [CGPoint] {
        guard points.count == 4 else { return points }
        let p = clampNormalized(location)
        var minX = points.map(\.x).min() ?? 0
        var maxX = points.map(\.x).max() ?? 1
        var minY = points.map(\.y).min() ?? 0
        var maxY = points.map(\.y).max() ?? 1
        // 0 TL, 1 TR, 2 BR, 3 BL, 4 top, 5 right, 6 bottom, 7 left
        switch handle {
        case 0:
            minX = p.x
            minY = p.y
        case 1:
            maxX = p.x
            minY = p.y
        case 2:
            maxX = p.x
            maxY = p.y
        case 3:
            minX = p.x
            maxY = p.y
        case 4:
            minY = p.y
        case 5:
            maxX = p.x
        case 6:
            maxY = p.y
        case 7:
            minX = p.x
        default:
            break
        }
        if minX > maxX { swap(&minX, &maxX) }
        if minY > maxY { swap(&minY, &maxY) }
        return rectanglePoints(from: CGPoint(x: minX, y: minY), to: CGPoint(x: maxX, y: maxY))
    }

    static func moveLHandle(
        kind: ArtifactRegionKind,
        points: [CGPoint],
        handle: Int,
        to location: CGPoint
    ) -> [CGPoint] {
        guard points.count == 6, kind.isLShape else { return points }
        let p = clampNormalized(location)
        var minX = points.map(\.x).min() ?? 0
        var maxX = points.map(\.x).max() ?? 1
        var minY = points.map(\.y).min() ?? 0
        var maxY = points.map(\.y).max() ?? 1
        // Outer-box handles 0…3 (corners of the bounding rect).
        switch handle {
        case 0:
            minX = p.x
            minY = p.y
        case 1:
            maxX = p.x
            minY = p.y
        case 2:
            maxX = p.x
            maxY = p.y
        case 3:
            minX = p.x
            maxY = p.y
        default:
            break
        }
        if minX > maxX { swap(&minX, &maxX) }
        if minY > maxY { swap(&minY, &maxY) }
        return lShapePoints(
            kind: kind,
            from: CGPoint(x: minX, y: minY),
            to: CGPoint(x: maxX, y: maxY)
        )
    }

    static func moveCircleRadius(points: [CGPoint], to location: CGPoint) -> [CGPoint] {
        guard let center = circleCenter(points) else { return points }
        let p = clampNormalized(location)
        let dx = p.x - center.x
        let dy = p.y - center.y
        return circleRing(center: center, radius: sqrt(dx * dx + dy * dy))
    }

    static func moveFreeformVertex(points: [CGPoint], index: Int, to location: CGPoint) -> [CGPoint] {
        guard points.indices.contains(index) else { return points }
        var next = points
        next[index] = clampNormalized(location)
        return next
    }

    static func deleteFreeformVertex(points: [CGPoint], index: Int) -> [CGPoint]? {
        guard points.indices.contains(index), points.count > freeformMinimumVertices else {
            return nil
        }
        var next = points
        next.remove(at: index)
        return next
    }

    // MARK: Queries

    static func isAxisAlignedRectangle(_ points: [CGPoint]) -> Bool {
        guard points.count == 4 else { return false }
        let xs = Set(points.map { rounded($0.x) })
        let ys = Set(points.map { rounded($0.y) })
        return xs.count == 2 && ys.count == 2
    }

    static func looksLikeCircleRing(_ points: [CGPoint]) -> Bool {
        guard points.count == circleRingCount, let center = circleCenter(points) else {
            return false
        }
        let radii = points.map { hypot($0.x - center.x, $0.y - center.y) }
        guard let first = radii.first, first > 0 else { return false }
        return radii.allSatisfy { abs($0 - first) < 0.04 }
    }

    static func circleCenter(_ points: [CGPoint]) -> CGPoint? {
        guard !points.isEmpty else { return nil }
        let sx = points.reduce(CGFloat(0)) { $0 + $1.x }
        let sy = points.reduce(CGFloat(0)) { $0 + $1.y }
        return CGPoint(x: sx / CGFloat(points.count), y: sy / CGFloat(points.count))
    }

    static func circleRadius(_ points: [CGPoint]) -> CGFloat? {
        guard let center = circleCenter(points), let first = points.first else { return nil }
        return hypot(first.x - center.x, first.y - center.y)
    }

    static func inferredLOpen(from points: [CGPoint]) -> ArtifactRegionKind? {
        guard points.count == 6 else { return nil }
        let minX = points.map(\.x).min() ?? 0
        let maxX = points.map(\.x).max() ?? 1
        let minY = points.map(\.y).min() ?? 0
        let maxY = points.map(\.y).max() ?? 1
        let candidates: [ArtifactRegionKind] = [
            .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft,
        ]
        return candidates.min { a, b in
            distance(
                points,
                lShapePoints(
                    kind: a,
                    from: CGPoint(x: minX, y: minY),
                    to: CGPoint(x: maxX, y: maxY)
                )
            ) < distance(
                points,
                lShapePoints(
                    kind: b,
                    from: CGPoint(x: minX, y: minY),
                    to: CGPoint(x: maxX, y: maxY)
                )
            )
        }
    }

    static func boundingRect(of points: [CGPoint]) -> CGRect {
        guard let first = points.first else { return .zero }
        var minX = first.x
        var maxX = first.x
        var minY = first.y
        var maxY = first.y
        for point in points.dropFirst() {
            minX = min(minX, point.x)
            maxX = max(maxX, point.x)
            minY = min(minY, point.y)
            maxY = max(maxY, point.y)
        }
        return CGRect(x: minX, y: minY, width: maxX - minX, height: maxY - minY)
    }

    static func polygonArea(_ points: [CGPoint]) -> CGFloat {
        let n = points.count
        guard n >= 3 else { return 0 }
        var sum: CGFloat = 0
        for i in 0..<n {
            let j = (i + 1) % n
            sum += points[i].x * points[j].y - points[j].x * points[i].y
        }
        return abs(sum) / 2
    }

    static func isValidPolygon(_ points: [CGPoint]) -> Bool {
        guard points.count >= 3 else { return false }
        if nearlyEqual(points[0], points[points.count - 1]) { return false }
        for point in points {
            if point.x < 0 || point.x > 1 || point.y < 0 || point.y > 1 {
                return false
            }
        }
        return polygonArea(points) >= minimumPolygonArea
    }

    static func nearlyEqual(_ a: CGPoint, _ b: CGPoint, epsilon: CGFloat = 1e-12) -> Bool {
        abs(a.x - b.x) < epsilon && abs(a.y - b.y) < epsilon
    }

    static func closeEnoughToOrigin(_ point: CGPoint, origin: CGPoint, threshold: CGFloat = 0.03) -> Bool {
        hypot(point.x - origin.x, point.y - origin.y) <= threshold
    }

    // MARK: Handles (normalized)

    static func handlePoints(for draft: ArtifactRegionDraft) -> [CGPoint] {
        switch draft.kind {
        case .rectangle:
            guard draft.points.count == 4 else { return draft.points }
            let tl = draft.points[0]
            let tr = draft.points[1]
            let br = draft.points[2]
            let bl = draft.points[3]
            return [
                tl, tr, br, bl,
                CGPoint(x: (tl.x + tr.x) / 2, y: tl.y),
                CGPoint(x: tr.x, y: (tr.y + br.y) / 2),
                CGPoint(x: (bl.x + br.x) / 2, y: bl.y),
                CGPoint(x: bl.x, y: (tl.y + bl.y) / 2),
            ]
        case .lOpenTopRight, .lOpenTopLeft, .lOpenBottomRight, .lOpenBottomLeft:
            let box = boundingRect(of: draft.points)
            return [
                CGPoint(x: box.minX, y: box.minY),
                CGPoint(x: box.maxX, y: box.minY),
                CGPoint(x: box.maxX, y: box.maxY),
                CGPoint(x: box.minX, y: box.maxY),
            ]
        case .circle:
            guard let center = circleCenter(draft.points),
                  let radius = circleRadius(draft.points)
            else { return [] }
            return [CGPoint(x: center.x + radius, y: center.y)]
        case .freeform:
            return draft.points
        }
    }

    // MARK: - Private

    private static func rounded(_ value: CGFloat) -> Int {
        Int((value * 10_000).rounded())
    }

    private static func distance(_ a: [CGPoint], _ b: [CGPoint]) -> CGFloat {
        guard a.count == b.count else { return .greatestFiniteMagnitude }
        return zip(a, b).reduce(0) { $0 + hypot($1.0.x - $1.1.x, $1.0.y - $1.1.y) }
    }
}
