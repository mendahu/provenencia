import CoreGraphics
import Foundation

/// The geometry of the record links diagram (S13-28, board S13-D10): nodes on a circle,
/// clockwise from the top, and a quadratic curve for each arc. Pure and Person-free: it takes
/// a node count, the arcs and a width, so it is tested without a view and an Event or Place
/// diagram reuses it.
struct RecordLinksLayout: Equatable {
    /// Which side of its dot a node's ref is printed on.
    enum LabelSide: Equatable {
        case right, left, top, bottom
    }

    struct Node: Equatable {
        var point: CGPoint
        var side: LabelSide
    }

    /// One arc as a quadratic curve from `from` to `to`, with the arc it draws.
    struct Curve: Equatable {
        var arc: RecordLinksArc
        var from: CGPoint
        var to: CGPoint
        var control: CGPoint
    }

    /// The smallest circle: a window too short for the column shrinks the circle to this and
    /// no further (the column then scrolls with the page instead).
    static let minimumDiameter: CGFloat = 172
    /// Room kept outside the circle, left and right of it (one ref plus 14pt).
    static let sideRoom: CGFloat = 84
    /// Room kept outside the circle, above and below it.
    static let verticalRoom: CGFloat = 30
    /// Where a single record's dot sits, and the height of the region it is drawn in: no
    /// circle, so just the dot and its ref.
    static let singleDotY: CGFloat = 24
    static let singleHeight: CGFloat = 64
    /// A node is labelled beside its dot when its dot is this far, as a fraction of the radius,
    /// from the vertical through the centre.
    private static let sideThreshold: CGFloat = 0.35

    /// The circle's diameter in a column `width` wide, shrunk to `fitting` (the height left for
    /// the drawing) but never below `minimumDiameter` and never beyond the column.
    static func diameter(width: CGFloat, fitting: CGFloat?) -> CGFloat {
        let widest = max(0, width - 2 * sideRoom)
        guard let fitting else { return widest }
        return min(widest, max(minimumDiameter, min(widest, fitting - 2 * verticalRoom)))
    }

    let nodes: [Node]
    /// One curve per arc, in the arcs' order.
    let curves: [Curve]
    let center: CGPoint
    let diameter: CGFloat
    let size: CGSize

    var radius: CGFloat { diameter / 2 }

    init(nodeCount: Int, arcs: [RecordLinksArc], width: CGFloat, diameter: CGFloat? = nil) {
        let diameter = diameter ?? Self.diameter(width: width, fitting: nil)
        let radius = diameter / 2
        let center = CGPoint(x: width / 2, y: Self.verticalRoom + radius)
        self.diameter = diameter
        self.center = center
        size = CGSize(width: width, height: nodeCount == 1 ? Self.singleHeight : diameter + 2 * Self.verticalRoom)
        nodes = (0..<max(0, nodeCount)).map { index in
            guard nodeCount > 1 else { return Node(point: CGPoint(x: width / 2, y: Self.singleDotY), side: .right) }
            let angle = 2 * Double.pi * Double(index) / Double(nodeCount)
            let dx = CGFloat(sin(angle))
            let dy = CGFloat(-cos(angle))
            let side: LabelSide
            if dx > Self.sideThreshold {
                side = .right
            } else if dx < -Self.sideThreshold {
                side = .left
            } else {
                side = dy <= 0 ? .top : .bottom
            }
            return Node(point: CGPoint(x: center.x + radius * dx, y: center.y + radius * dy), side: side)
        }
        let points = nodes
        curves = arcs.compactMap { arc in
            guard points.indices.contains(arc.a), points.indices.contains(arc.b), arc.a != arc.b else { return nil }
            return Self.curve(from: points[arc.a].point, to: points[arc.b].point, arc: arc, center: center, radius: radius)
        }
    }

    /// The gap between neighbouring arcs of one pair: 6% of the radius between 7 and 14pt, and
    /// at most the radius over the gaps, so a pair with many shared people stays in the circle.
    static func spacing(radius: CGFloat, slots: Int) -> CGFloat {
        let base = min(14, max(7, 0.06 * radius))
        guard slots > 1 else { return base }
        return min(base, radius / CGFloat(slots - 1))
    }

    private static func curve(
        from: CGPoint, to: CGPoint, arc: RecordLinksArc, center: CGPoint, radius: CGFloat
    ) -> Curve {
        let mid = CGPoint(x: (from.x + to.x) / 2, y: (from.y + to.y) / 2)
        // Pulled to 30% of the way from the centre to the chord's midpoint: the curve bows
        // toward the middle of the circle.
        var control = CGPoint(x: center.x + 0.3 * (mid.x - center.x), y: center.y + 0.3 * (mid.y - center.y))
        let dx = to.x - from.x
        let dy = to.y - from.y
        let length = (dx * dx + dy * dy).squareRoot()
        if length > 0 {
            let offset = (CGFloat(arc.slot) - CGFloat(arc.slots - 1) / 2) * spacing(radius: radius, slots: arc.slots)
            control.x += -dy / length * offset
            control.y += dx / length * offset
        }
        return Curve(arc: arc, from: from, to: to, control: control)
    }
}
