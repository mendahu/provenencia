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
/// dataset groups them by pair, in input order, and the layout fans them side by side.
struct CircularLinkArc<ID: Hashable>: Equatable {
    var from: ID
    var to: ID
    var lineWidth: CGFloat
    /// Nil: the style's `arcColor`. A highlighted arc draws in `style.highlightedArcColor` instead.
    var color: Color? = nil
}

/// An unordered pair of node indexes.
private struct CircularLinkPair: Hashable {
    var low: Int
    var high: Int
}

/// A diagram's nodes and arcs, resolved once: the one place an arc's ends are found, an arc
/// that names no node or one node is skipped, an arc gets its place among those on its pair,
/// and each node's neighbours are collected. The layout and the highlight both read it, so
/// neither finds a node by id again.
struct CircularLinkDataset<ID: Hashable>: Equatable {
    /// An arc whose two ends are different nodes.
    struct ResolvedArc: Equatable {
        /// The arc's index in the input `arcs`.
        var arcIndex: Int
        /// The node indexes of its ends.
        var from: Int
        var to: Int
        /// Its place among the arcs on its unordered pair, in input order, and how many there are.
        var slot: Int
        var slots: Int
        /// True when it runs against its pair's reference direction, the first arc's `from` to `to`.
        var flipped: Bool
        var lineWidth: CGFloat
        var color: Color?
    }

    /// A duplicate id keeps the first.
    private(set) var nodes: [CircularLinkNode<ID>]
    /// In input order.
    let resolvedArcs: [ResolvedArc]
    private let indexes: [ID: Int]
    private let adjacency: [Set<Int>]

    init(nodes input: [CircularLinkNode<ID>], arcs: [CircularLinkArc<ID>]) {
        var nodes: [CircularLinkNode<ID>] = []
        var indexes: [ID: Int] = [:]
        for node in input where indexes[node.id] == nil {
            indexes[node.id] = nodes.count
            nodes.append(node)
        }
        var counts: [CircularLinkPair: Int] = [:]
        var references: [CircularLinkPair: Int] = [:]
        var resolved: [(arc: Int, from: Int, to: Int, slot: Int, pair: CircularLinkPair)] = []
        var neighbors = [Set<Int>](repeating: [], count: nodes.count)
        for (arcIndex, arc) in arcs.enumerated() {
            guard let from = indexes[arc.from], let to = indexes[arc.to], from != to else { continue }
            let pair = CircularLinkPair(low: min(from, to), high: max(from, to))
            if references[pair] == nil { references[pair] = from }
            resolved.append((arcIndex, from, to, counts[pair, default: 0], pair))
            counts[pair, default: 0] += 1
            neighbors[from].insert(to)
            neighbors[to].insert(from)
        }
        self.nodes = nodes
        self.indexes = indexes
        adjacency = neighbors
        resolvedArcs = resolved.map { item in
            ResolvedArc(
                arcIndex: item.arc, from: item.from, to: item.to, slot: item.slot,
                slots: counts[item.pair] ?? 1, flipped: references[item.pair] != item.from,
                lineWidth: arcs[item.arc].lineWidth, color: arcs[item.arc].color
            )
        }
    }

    /// The node indexes sharing an arc with the node at `index`.
    func neighbors(ofIndex index: Int) -> Set<Int> { adjacency[index] }

    /// How many nodes share an arc with the node `id`; zero for an id that is no node.
    func neighborCount(of id: ID) -> Int { indexes[id].map { adjacency[$0].count } ?? 0 }

    /// The index of the node `id`.
    func index(of id: ID) -> Int? { indexes[id] }

    /// The same dataset with each node replaced by `transform(index, node)`, for a node's text
    /// that reads this dataset's neighbours. The id is put back after the transform.
    func mappingNodes(_ transform: (Int, CircularLinkNode<ID>) -> CircularLinkNode<ID>) -> Self {
        var copy = self
        copy.nodes = nodes.enumerated().map { index, node in
            var mapped = transform(index, node)
            mapped.id = node.id
            return mapped
        }
        return copy
    }
}

/// The node the pointer is on, on a node or on any part of something that stands for it. Its
/// own observable, so a hover redraws only the views that read it and never the caller's
/// content or the diagram's layout.
@MainActor @Observable
final class CircularLinkHover<ID: Hashable> {
    var id: ID?
}
