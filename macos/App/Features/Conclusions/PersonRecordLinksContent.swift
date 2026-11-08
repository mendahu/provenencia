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

    /// Whether node `index` is one end of this arc.
    func touches(_ index: Int) -> Bool { a == index || b == index }

    /// The nodes sharing a person with node `index`: the other end of each of its arcs.
    static func neighbors(of index: Int, in arcs: [RecordLinksArc]) -> Set<Int> {
        var result = Set<Int>()
        for arc in arcs {
            if arc.a == index { result.insert(arc.b) } else if arc.b == index { result.insert(arc.a) }
        }
        return result
    }
}

/// What the record links diagram shows for a Person: the Person adapter over the S13-27 read
/// and the claims content. Pure; the layout and the diagram take only nodes and arcs.
struct PersonRecordLinksContent: Equatable {
    let nodes: [RecordLinksNode]
    let arcs: [RecordLinksArc]

    // decided: S13-28 round 1 (SEC-1), the caps-for-adversarial-only rule: a high total cap a researcher never reaches, so a crafted catalog cannot send a million lines to a main-thread draw; 20 times the 3,675 arcs of S13-27's 50-graph household Scale
    static let maxArcs = 75_000
    // decided: S13-28 round 1 (SEC-1), same rule: 20 times the 50 graphs of that Scale
    static let maxNodes = 1_000

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
        for graph in links.graphs where indexes[graph.graphID] == nil && graphs.count < Self.maxNodes {
            indexes[graph.graphID] = graphs.count
            graphs.append(graph)
        }
        // Past `maxArcs` the strong lines are kept first, then the thin ones, each in the read's order.
        var valid = 0
        for pair in links.pairs where Self.ends(of: pair, in: indexes) != nil { valid += pair.lines.count }
        var arcs: [RecordLinksArc] = []
        arcs.reserveCapacity(min(valid, Self.maxArcs))
        let passes: [CatalogRecordLinkStrength?] = valid <= Self.maxArcs ? [nil] : [.strong, .weak]
        for pass in passes {
            for pair in links.pairs {
                guard let (a, b) = Self.ends(of: pair, in: indexes) else { continue }
                for (slot, line) in pair.lines.enumerated() where pass == nil || line.strength == pass {
                    guard arcs.count < Self.maxArcs else { break }
                    arcs.append(RecordLinksArc(a: a, b: b, strong: line.strength == .strong, slot: slot, slots: pair.lines.count))
                }
            }
        }
        nodes = graphs.enumerated().map { index, graph in
            let name = graph.graphName.trimmedNonEmpty
            let ref = graph.graphRef.trimmedNonEmpty
            let graphTitle = name ?? ref ?? ""
            let sourceTitle = graph.sourceTitle.trimmedNonEmpty ?? graphSourceRefs[graph.graphID] ?? ""
            let label = ref ?? name ?? ""
            let connected = RecordLinksArc.neighbors(of: index, in: arcs).count
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

    private static func ends(of pair: CatalogRecordLinksPair, in indexes: [String: Int]) -> (Int, Int)? {
        guard let a = indexes[pair.graphAID], let b = indexes[pair.graphBID], a != b else { return nil }
        return (a, b)
    }
}
