import SwiftUI

/// One node of a circular link diagram: a dot on the circle and the label beside it. Generic
/// over the caller's id, so the diagram knows nothing of what a node is.
struct CircularLinkNode<ID: Hashable>: Identifiable, Equatable {
    var id: ID
    /// Printed beside the dot.
    var label: String
    /// The native `.help` tooltip; nil draws none.
    var help: String? = nil
    /// The spoken label; nil reads `label`.
    var accessibilityLabel: String? = nil
    /// False: drawn, not clickable, no Return.
    var isEnabled: Bool = true
}

/// One arc between two nodes. Several arcs on one unordered pair are the pair's "amount": the
/// layout groups them by pair, in input order, and fans them side by side.
struct CircularLinkArc<ID: Hashable>: Equatable {
    var from: ID
    var to: ID
    var lineWidth: CGFloat
    /// Nil: the style's `arcColor`.
    var color: Color? = nil

    /// The nodes sharing an arc with `id`: the other end of each of its arcs.
    static func neighbors(of id: ID, in arcs: [Self]) -> Set<ID> {
        var result = Set<ID>()
        for arc in arcs where arc.from != arc.to {
            if arc.from == id { result.insert(arc.to) } else if arc.to == id { result.insert(arc.from) }
        }
        return result
    }

    /// Every node's neighbours, in one pass over the arcs.
    static func neighborSets(in arcs: [Self]) -> [ID: Set<ID>] {
        var result: [ID: Set<ID>] = [:]
        for arc in arcs where arc.from != arc.to {
            result[arc.from, default: []].insert(arc.to)
            result[arc.to, default: []].insert(arc.from)
        }
        return result
    }
}

/// The node the pointer is on, on a node or on any part of something that stands for it. Its
/// own observable, so a hover redraws only the views that read it and never the caller's
/// content or the diagram's layout.
@MainActor @Observable
final class CircularLinkHover<ID: Hashable> {
    var id: ID?
}
