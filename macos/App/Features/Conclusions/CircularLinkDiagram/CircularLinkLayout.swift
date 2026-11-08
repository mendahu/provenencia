import SwiftUI
import Foundation

/// Which side of its dot a node's label is printed on.
enum CircularLinkLabelSide: Equatable {
    case right, left, top, bottom
}

/// The geometry of a circular link diagram: nodes on a circle, clockwise from the top, and a
/// quadratic curve for each arc. Pure and free of what a node is: it takes ids, arcs, a width
/// and a style, so it is tested without a view.
struct CircularLinkLayout<ID: Hashable>: Equatable {
    typealias LabelSide = CircularLinkLabelSide

    struct Node: Equatable {
        var point: CGPoint
        var side: LabelSide
    }

    /// One arc as a quadratic curve from `from` to `to`.
    struct Curve: Equatable {
        /// The arc's index in the input `arcs`.
        var arcIndex: Int
        /// The node indexes of its ends.
        var fromNode: Int
        var toNode: Int
        var lineWidth: CGFloat
        var color: Color?
        var from: CGPoint
        var to: CGPoint
        var control: CGPoint

        func touches(_ node: Int) -> Bool { fromNode == node || toNode == node }
    }

    /// The circle's diameter in a column `width` wide, shrunk to `fitting` (the height left for
    /// the drawing) but never below `minimumDiameter` and never beyond the column.
    static func diameter(width: CGFloat, fitting: CGFloat?, style: CircularLinkDiagramStyle = .standard) -> CGFloat {
        let widest = max(0, width - 2 * style.sideRoom)
        guard let fitting else { return widest }
        return min(widest, max(style.minimumDiameter, min(widest, fitting - 2 * style.verticalRoom)))
    }

    /// The gap between neighbouring arcs of one pair, at most the radius over the gaps, so a
    /// pair with many arcs stays in the circle.
    static func spacing(radius: CGFloat, slots: Int, style: CircularLinkDiagramStyle = .standard) -> CGFloat {
        let base = min(style.arcSpacingRange.upperBound, max(style.arcSpacingRange.lowerBound, style.arcSpacingFraction * radius))
        guard slots > 1 else { return base }
        return min(base, radius / CGFloat(slots - 1))
    }

    let nodes: [Node]
    /// Thinnest first, then in the arcs' order.
    let curves: [Curve]
    let center: CGPoint
    let diameter: CGFloat
    let size: CGSize

    var radius: CGFloat { diameter / 2 }

    private struct Pair: Hashable {
        var low: Int
        var high: Int
    }

    init(
        nodes ids: [ID], arcs: [CircularLinkArc<ID>], width: CGFloat, diameter: CGFloat? = nil,
        style: CircularLinkDiagramStyle = .standard
    ) {
        let diameter = diameter ?? Self.diameter(width: width, fitting: nil, style: style)
        let radius = diameter / 2
        let center = CGPoint(x: width / 2, y: style.verticalRoom + radius)
        let nodeCount = ids.count
        self.diameter = diameter
        self.center = center
        size = CGSize(width: width, height: nodeCount == 1 ? style.singleNodeHeight : diameter + 2 * style.verticalRoom)
        let nodes: [Node] = (0..<nodeCount).map { index in
            guard nodeCount > 1 else { return Node(point: CGPoint(x: width / 2, y: style.singleNodeY), side: .right) }
            let angle = 2 * Double.pi * Double(index) / Double(nodeCount)
            let dx = CGFloat(sin(angle))
            let dy = CGFloat(-cos(angle))
            let side: LabelSide
            if dx > style.labelSideThreshold {
                side = .right
            } else if dx < -style.labelSideThreshold {
                side = .left
            } else {
                side = dy <= 0 ? .top : .bottom
            }
            return Node(point: CGPoint(x: center.x + radius * dx, y: center.y + radius * dy), side: side)
        }
        self.nodes = nodes

        var indexes: [ID: Int] = [:]
        for (index, id) in ids.enumerated() where indexes[id] == nil { indexes[id] = index }
        // The arcs that name two different nodes, each with its place among those on its pair.
        var ends: [(arc: Int, from: Int, to: Int)] = []
        var counts: [Pair: Int] = [:]
        var slotOf: [Int] = []
        for (arcIndex, arc) in arcs.enumerated() {
            guard let from = indexes[arc.from], let to = indexes[arc.to], from != to else { continue }
            let pair = Pair(low: min(from, to), high: max(from, to))
            slotOf.append(counts[pair, default: 0])
            counts[pair, default: 0] += 1
            ends.append((arcIndex, from, to))
        }
        var curves: [Curve] = []
        curves.reserveCapacity(ends.count)
        for (position, end) in ends.enumerated() {
            let arc = arcs[end.arc]
            let slots = counts[Pair(low: min(end.from, end.to), high: max(end.from, end.to))] ?? 1
            let from = nodes[end.from].point
            let to = nodes[end.to].point
            let mid = CGPoint(x: (from.x + to.x) / 2, y: (from.y + to.y) / 2)
            // Pulled toward the chord's midpoint from the centre: the curve bows toward the middle of the circle.
            var control = CGPoint(
                x: center.x + style.arcCurvature * (mid.x - center.x),
                y: center.y + style.arcCurvature * (mid.y - center.y)
            )
            let dx = to.x - from.x
            let dy = to.y - from.y
            let length = (dx * dx + dy * dy).squareRoot()
            if length > 0 {
                let offset = (CGFloat(slotOf[position]) - CGFloat(slots - 1) / 2)
                    * Self.spacing(radius: radius, slots: slots, style: style)
                control.x += -dy / length * offset
                control.y += dx / length * offset
            }
            curves.append(Curve(
                arcIndex: end.arc, fromNode: end.from, toNode: end.to, lineWidth: arc.lineWidth, color: arc.color,
                from: from, to: to, control: control
            ))
        }
        // Thinnest first; a stable sort keeps the input order within one width.
        var ordered = curves.enumerated().map { $0 }
        ordered.sort { ($0.element.lineWidth, $0.offset) < ($1.element.lineWidth, $1.offset) }
        self.curves = ordered.map(\.element)
    }
}
