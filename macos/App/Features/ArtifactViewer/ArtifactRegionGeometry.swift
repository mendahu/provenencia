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

    static func lShapePoints(kind: ArtifactRegionKind, params: LParams) -> [CGPoint] {
        var p = params
        p.clamp()
        return rebuiltL(kind: kind, params: p)
    }

    private static func rebuiltL(kind: ArtifactRegionKind, params p: LParams) -> [CGPoint] {
        let raw: [CGPoint]
        switch kind {
        case .lOpenTopRight:
            raw = [
                CGPoint(x: p.minX, y: p.minY),
                CGPoint(x: p.midX, y: p.minY),
                CGPoint(x: p.midX, y: p.midY),
                CGPoint(x: p.maxX, y: p.midY),
                CGPoint(x: p.maxX, y: p.maxY),
                CGPoint(x: p.minX, y: p.maxY),
            ]
        case .lOpenTopLeft:
            raw = [
                CGPoint(x: p.midX, y: p.minY),
                CGPoint(x: p.maxX, y: p.minY),
                CGPoint(x: p.maxX, y: p.maxY),
                CGPoint(x: p.minX, y: p.maxY),
                CGPoint(x: p.minX, y: p.midY),
                CGPoint(x: p.midX, y: p.midY),
            ]
        case .lOpenBottomRight:
            raw = [
                CGPoint(x: p.minX, y: p.minY),
                CGPoint(x: p.maxX, y: p.minY),
                CGPoint(x: p.maxX, y: p.midY),
                CGPoint(x: p.midX, y: p.midY),
                CGPoint(x: p.midX, y: p.maxY),
                CGPoint(x: p.minX, y: p.maxY),
            ]
        case .lOpenBottomLeft:
            raw = [
                CGPoint(x: p.minX, y: p.minY),
                CGPoint(x: p.maxX, y: p.minY),
                CGPoint(x: p.maxX, y: p.maxY),
                CGPoint(x: p.midX, y: p.maxY),
                CGPoint(x: p.midX, y: p.midY),
                CGPoint(x: p.minX, y: p.midY),
            ]
        default:
            return rectanglePoints(
                from: CGPoint(x: p.minX, y: p.minY),
                to: CGPoint(x: p.maxX, y: p.maxY)
            )
        }
        return raw.map(clampNormalized)
    }

    /// Circle in **image pixels** (isotropic), then normalized. A unit-square
    /// ring would stretch into an oval on a non-square Artifact.
    static func circleRing(
        centerNormalized: CGPoint,
        throughNormalized: CGPoint,
        imageRect: CGRect
    ) -> [CGPoint] {
        let media = usableImageRect(imageRect)
        let centerDoc = documentPoint(normalized: centerNormalized, imageRect: media)
        let edgeDoc = documentPoint(normalized: throughNormalized, imageRect: media)
        let radius = hypot(edgeDoc.x - centerDoc.x, edgeDoc.y - centerDoc.y)
        return circleRing(centerDocument: centerDoc, radiusDocument: radius, imageRect: media)
    }

    static func circleRing(
        centerDocument: CGPoint,
        radiusDocument: CGFloat,
        imageRect: CGRect
    ) -> [CGPoint] {
        let media = usableImageRect(imageRect)
        let maxR = min(
            centerDocument.x - media.minX,
            media.maxX - centerDocument.x,
            centerDocument.y - media.minY,
            media.maxY - centerDocument.y
        )
        let r = min(max(radiusDocument, 0), max(maxR, 0))
        guard r > 0 else { return [] }
        return (0..<circleRingCount).map { index in
            let angle = CGFloat(index) / CGFloat(circleRingCount) * 2 * .pi
            return normalize(
                documentPoint: CGPoint(
                    x: centerDocument.x + r * cos(angle),
                    y: centerDocument.y + r * sin(angle)
                ),
                imageRect: media
            )
        }
    }

    /// Unit-square helper for tests (square image). Prefer the image-rect overload in the overlay.
    static func circleRing(center: CGPoint, radius: CGFloat) -> [CGPoint] {
        let media = CGRect(x: 0, y: 0, width: 1, height: 1)
        return circleRing(
            centerDocument: documentPoint(normalized: center, imageRect: media),
            radiusDocument: radius,
            imageRect: media
        )
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
        guard points.count == 6, kind.isLShape, var params = lParams(kind: kind, points: points) else {
            return points
        }
        let p = clampNormalized(location)
        applyLVertex(&params, kind: kind, handle: handle, to: p)
        return lShapePoints(kind: kind, params: params)
    }

    static func moveCircleRadius(
        points: [CGPoint],
        to location: CGPoint,
        imageRect: CGRect
    ) -> [CGPoint] {
        let media = usableImageRect(imageRect)
        guard let center = circleCenter(points) else { return points }
        return circleRing(
            centerNormalized: center,
            throughNormalized: clampNormalized(location),
            imageRect: media
        )
    }

    static func moveCircleRadius(points: [CGPoint], to location: CGPoint) -> [CGPoint] {
        moveCircleRadius(
            points: points,
            to: location,
            imageRect: CGRect(x: 0, y: 0, width: 1, height: 1)
        )
    }

    /// Translate every vertex, clamping so the polygon stays inside the unit square.
    static func translate(_ points: [CGPoint], by delta: CGPoint) -> [CGPoint] {
        guard !points.isEmpty else { return points }
        let minX = points.map(\.x).min() ?? 0
        let maxX = points.map(\.x).max() ?? 1
        let minY = points.map(\.y).min() ?? 0
        let maxY = points.map(\.y).max() ?? 1
        let dx = min(max(delta.x, -minX), 1 - maxX)
        let dy = min(max(delta.y, -minY), 1 - maxY)
        return points.map { clampNormalized(CGPoint(x: $0.x + dx, y: $0.y + dy)) }
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
        points.count == circleRingCount && circleCenter(points) != nil
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
        guard polygonArea(points) >= minimumPolygonArea else { return false }
        return !selfIntersects(points)
    }

    /// True when non-adjacent edges of the closed polygon cross.
    /// Matches `locator.validateRegion` in Go.
    private static func selfIntersects(_ points: [CGPoint]) -> Bool {
        let count = points.count
        guard count >= 4 else { return false }
        for i in 0..<count {
            let a1 = points[i]
            let a2 = points[(i + 1) % count]
            for j in (i + 1)..<count {
                if j == i || (j + 1) % count == i || (i + 1) % count == j {
                    continue
                }
                let b1 = points[j]
                let b2 = points[(j + 1) % count]
                if segmentsIntersect(a1, a2, b1, b2) {
                    return true
                }
            }
        }
        return false
    }

    private static func segmentsIntersect(_ a1: CGPoint, _ a2: CGPoint, _ b1: CGPoint, _ b2: CGPoint) -> Bool {
        let d1 = cross(a2.x - a1.x, a2.y - a1.y, b1.x - a1.x, b1.y - a1.y)
        let d2 = cross(a2.x - a1.x, a2.y - a1.y, b2.x - a1.x, b2.y - a1.y)
        let d3 = cross(b2.x - b1.x, b2.y - b1.y, a1.x - b1.x, a1.y - b1.y)
        let d4 = cross(b2.x - b1.x, b2.y - b1.y, a2.x - b1.x, a2.y - b1.y)
        return ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0))
            && ((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0))
    }

    private static func cross(_ ax: CGFloat, _ ay: CGFloat, _ bx: CGFloat, _ by: CGFloat) -> CGFloat {
        ax * by - ay * bx
    }

    static func nearlyEqual(_ a: CGPoint, _ b: CGPoint, epsilon: CGFloat = 1e-12) -> Bool {
        abs(a.x - b.x) < epsilon && abs(a.y - b.y) < epsilon
    }

    static func closeEnoughToOrigin(_ point: CGPoint, origin: CGPoint, threshold: CGFloat = 0.03) -> Bool {
        hypot(point.x - origin.x, point.y - origin.y) <= threshold
    }

    // MARK: Handles (normalized)

    /// Resize handles only (not the move anchor). L uses the six real vertices.
    static func handlePoints(for draft: ArtifactRegionDraft, imageRect: CGRect = .unitSquare) -> [CGPoint] {
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
            return draft.points
        case .circle:
            let media = usableImageRect(imageRect)
            guard let center = circleCenter(draft.points) else { return [] }
            let centerDoc = documentPoint(normalized: center, imageRect: media)
            let radiusDoc = circleRadiusDocument(draft.points, imageRect: media) ?? 0
            guard radiusDoc > 0 else { return [] }
            return [
                normalize(
                    documentPoint: CGPoint(x: centerDoc.x + radiusDoc, y: centerDoc.y),
                    imageRect: media
                )
            ]
        case .freeform:
            return draft.points
        }
    }

    static func moveAnchor(for draft: ArtifactRegionDraft) -> CGPoint? {
        circleCenter(draft.points)
    }

    /// Eight-way resize cursor. Overlay y increases downward (flipped).
    static func resizeCursorKind(handle: CGPoint, centroid: CGPoint) -> ArtifactResizeCursorKind {
        let dx = handle.x - centroid.x
        let dy = handle.y - centroid.y
        guard abs(dx) > 1e-6 || abs(dy) > 1e-6 else { return .east }
        var degrees = atan2(dy, dx) * 180 / .pi
        if degrees < 0 { degrees += 360 }
        switch Int((degrees + 22.5) / 45) % 8 {
        case 0: return .east
        case 1: return .southEast
        case 2: return .south
        case 3: return .southWest
        case 4: return .west
        case 5: return .northWest
        case 6: return .north
        default: return .northEast
        }
    }

    static func missingLCorner(kind: ArtifactRegionKind, points: [CGPoint]) -> CGPoint? {
        let box = boundingRect(of: points)
        switch kind {
        case .lOpenTopRight: return CGPoint(x: box.maxX, y: box.minY)
        case .lOpenTopLeft: return CGPoint(x: box.minX, y: box.minY)
        case .lOpenBottomRight: return CGPoint(x: box.maxX, y: box.maxY)
        case .lOpenBottomLeft: return CGPoint(x: box.minX, y: box.maxY)
        default: return nil
        }
    }

    // MARK: - Private

    private static func usableImageRect(_ rect: CGRect) -> CGRect {
        (rect.width > 0 && rect.height > 0) ? rect : .unitSquare
    }

    static func circleRadiusDocument(_ points: [CGPoint], imageRect: CGRect) -> CGFloat? {
        let media = usableImageRect(imageRect)
        guard let center = circleCenter(points), let first = points.first else { return nil }
        let c = documentPoint(normalized: center, imageRect: media)
        let p = documentPoint(normalized: first, imageRect: media)
        return hypot(p.x - c.x, p.y - c.y)
    }

    static func lParams(kind: ArtifactRegionKind, points: [CGPoint]) -> LParams? {
        guard points.count == 6 else { return nil }
        switch kind {
        case .lOpenTopRight:
            return LParams(
                minX: points[0].x, maxX: points[4].x,
                minY: points[0].y, maxY: points[4].y,
                midX: points[2].x, midY: points[2].y
            )
        case .lOpenTopLeft:
            return LParams(
                minX: points[3].x, maxX: points[1].x,
                minY: points[0].y, maxY: points[2].y,
                midX: points[5].x, midY: points[5].y
            )
        case .lOpenBottomRight:
            return LParams(
                minX: points[0].x, maxX: points[1].x,
                minY: points[0].y, maxY: points[5].y,
                midX: points[3].x, midY: points[3].y
            )
        case .lOpenBottomLeft:
            return LParams(
                minX: points[0].x, maxX: points[1].x,
                minY: points[0].y, maxY: points[2].y,
                midX: points[4].x, midY: points[4].y
            )
        default:
            return nil
        }
    }

    private static func applyLVertex(
        _ params: inout LParams,
        kind: ArtifactRegionKind,
        handle: Int,
        to p: CGPoint
    ) {
        switch (kind, handle) {
        case (.lOpenTopRight, 0): params.minX = p.x; params.minY = p.y
        case (.lOpenTopRight, 1): params.midX = p.x
        case (.lOpenTopRight, 2): params.midX = p.x; params.midY = p.y
        case (.lOpenTopRight, 3): params.midY = p.y
        case (.lOpenTopRight, 4): params.maxX = p.x; params.maxY = p.y
        case (.lOpenTopRight, 5): params.minX = p.x; params.maxY = p.y
        case (.lOpenTopLeft, 0): params.midX = p.x; params.minY = p.y
        case (.lOpenTopLeft, 1): params.maxX = p.x; params.minY = p.y
        case (.lOpenTopLeft, 2): params.maxX = p.x; params.maxY = p.y
        case (.lOpenTopLeft, 3): params.minX = p.x; params.maxY = p.y
        case (.lOpenTopLeft, 4): params.minX = p.x; params.midY = p.y
        case (.lOpenTopLeft, 5): params.midX = p.x; params.midY = p.y
        case (.lOpenBottomRight, 0): params.minX = p.x; params.minY = p.y
        case (.lOpenBottomRight, 1): params.maxX = p.x; params.minY = p.y
        case (.lOpenBottomRight, 2): params.maxX = p.x; params.midY = p.y
        case (.lOpenBottomRight, 3): params.midX = p.x; params.midY = p.y
        case (.lOpenBottomRight, 4): params.midX = p.x; params.maxY = p.y
        case (.lOpenBottomRight, 5): params.minX = p.x; params.maxY = p.y
        case (.lOpenBottomLeft, 0): params.minX = p.x; params.minY = p.y
        case (.lOpenBottomLeft, 1): params.maxX = p.x; params.minY = p.y
        case (.lOpenBottomLeft, 2): params.maxX = p.x; params.maxY = p.y
        case (.lOpenBottomLeft, 3): params.midX = p.x; params.maxY = p.y
        case (.lOpenBottomLeft, 4): params.midX = p.x; params.midY = p.y
        case (.lOpenBottomLeft, 5): params.minX = p.x; params.midY = p.y
        default: break
        }
        params.clamp()
    }

    private static func rounded(_ value: CGFloat) -> Int {
        Int((value * 10_000).rounded())
    }

    private static func distance(_ a: [CGPoint], _ b: [CGPoint]) -> CGFloat {
        guard a.count == b.count else { return .greatestFiniteMagnitude }
        return zip(a, b).reduce(0) { $0 + hypot($1.0.x - $1.1.x, $1.0.y - $1.1.y) }
    }
}

enum ArtifactResizeCursorKind: Equatable {
    case east, southEast, south, southWest, west, northWest, north, northEast
}

struct LParams: Equatable {
    var minX: CGFloat
    var maxX: CGFloat
    var minY: CGFloat
    var maxY: CGFloat
    var midX: CGFloat
    var midY: CGFloat

    mutating func clamp() {
        if minX > maxX { swap(&minX, &maxX) }
        if minY > maxY { swap(&minY, &maxY) }
        let gap: CGFloat = 0.02
        midX = min(max(midX, minX + gap), maxX - gap)
        midY = min(max(midY, minY + gap), maxY - gap)
    }
}

extension CGRect {
    static let unitSquare = CGRect(x: 0, y: 0, width: 1, height: 1)
}
