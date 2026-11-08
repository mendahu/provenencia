import Foundation

/// One record of a Person on the record links diagram: an Evidence graph (S13-28).
struct RecordLinksNode: Equatable, Identifiable {
    /// The graph id.
    var id: String
    /// The graph's ref, else its name: what the circle prints beside the dot.
    var label: String
    /// The help tag: "Household 214 — Hamburg census 1867".
    var help: String
    /// Where a click opens the record; nil when the claims read has not named the graph.
    var location: WorkspaceLocation?
    /// How many other graphs this one shares a person with.
    var connected: Int
    var accessibilityLabel: String
}

/// One person two records both name: an arc between two nodes. `slot` of `slots` orders the
/// arcs of one pair, in the read's line order, so they are drawn side by side.
struct RecordLinksArc: Equatable {
    /// Indexes into the nodes.
    var a: Int
    var b: Int
    var strong: Bool
    var slot: Int
    var slots: Int
}

/// What the record links diagram shows for a Person: the Person adapter over the S13-27 read
/// and the claims content. Pure; the layout and the diagram take only nodes and arcs.
struct PersonRecordLinksContent: Equatable {
    let nodes: [RecordLinksNode]
    let arcs: [RecordLinksArc]

    /// `graphLocations` and `graphSourceRefs` come from `PersonSourcesContent`, the one
    /// builder of a graph's location. A duplicate graph id keeps the first; a pair that
    /// names a graph not in the list is dropped.
    init(
        links: CatalogPersonRecordLinks,
        graphLocations: [String: WorkspaceLocation],
        graphSourceRefs: [String: String] = [:]
    ) {
        var indexes: [String: Int] = [:]
        var graphs: [CatalogRecordLinksGraph] = []
        for graph in links.graphs where indexes[graph.graphID] == nil {
            indexes[graph.graphID] = graphs.count
            graphs.append(graph)
        }
        var arcs: [RecordLinksArc] = []
        for pair in links.pairs {
            guard let a = indexes[pair.graphAID], let b = indexes[pair.graphBID], a != b else { continue }
            for (slot, line) in pair.lines.enumerated() {
                arcs.append(RecordLinksArc(a: a, b: b, strong: line.strength == .strong, slot: slot, slots: pair.lines.count))
            }
        }
        var others = Array(repeating: Set<Int>(), count: graphs.count)
        for arc in arcs {
            others[arc.a].insert(arc.b)
            others[arc.b].insert(arc.a)
        }
        nodes = graphs.enumerated().map { index, graph in
            let name = graph.graphName.trimmedNonEmpty
            let ref = graph.graphRef.trimmedNonEmpty
            let graphTitle = name ?? ref ?? ""
            let sourceTitle = graph.sourceTitle.trimmedNonEmpty ?? graphSourceRefs[graph.graphID] ?? ""
            let label = ref ?? name ?? ""
            let connected = others[index].count
            return RecordLinksNode(
                id: graph.graphID,
                label: label,
                help: L10n.Conclusions.recordLinksNodeHelp(graph: graphTitle, source: sourceTitle),
                location: graphLocations[graph.graphID],
                connected: connected,
                accessibilityLabel: connected == 0
                    ? L10n.Conclusions.recordLinksNodeApartAccessibility(ref: label, graph: graphTitle, source: sourceTitle)
                    : L10n.Conclusions.recordLinksNodeAccessibility(
                        ref: label, graph: graphTitle, source: sourceTitle, connected: connected
                    )
            )
        }
        self.arcs = arcs
    }

    /// The nodes sharing a person with node `index`: the other end of each of its arcs.
    func neighbors(of index: Int) -> Set<Int> {
        var result = Set<Int>()
        for arc in arcs {
            if arc.a == index { result.insert(arc.b) } else if arc.b == index { result.insert(arc.a) }
        }
        return result
    }
}
