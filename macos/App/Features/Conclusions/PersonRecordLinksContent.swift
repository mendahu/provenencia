import Foundation

/// What the record links diagram shows for a Person: the one Person-aware piece, the adapter
/// over the S13-27 read and the claims content. Pure; the diagram takes only a dataset of nodes and arcs.
struct PersonRecordLinksContent: Equatable {
    /// The nodes, keyed by graph id, and one arc per line of a pair, resolved once.
    let dataset: CircularLinkDataset<String>
    /// Where a click on a graph's node opens its record; a graph the claims read has not named is absent.
    let locations: [String: WorkspaceLocation]
    /// Why the drawing is not the whole read, nil when it is.
    let partial: Partial?

    /// The one note for a drawing that leaves something out, whichever side cut it.
    enum Partial: Equatable {
        /// The engine did not compare the records, so it sent no link.
        case notCounted
        /// Records or links are left out: the engine dropped lines, or there are more records than nodes drawn.
        case leftOut
    }

    /// The board's line widths (S13-D10): a strong line and a weak one.
    static let strongLineWidth: CGFloat = 3
    static let weakLineWidth: CGFloat = 1.25

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
        var ids = Set<String>()
        var graphs: [CatalogRecordLinksGraph] = []
        var graphsCut = false
        for graph in links.graphs where !ids.contains(graph.graphID) {
            if graphs.count == Self.maxNodes {
                graphsCut = true
                break
            }
            ids.insert(graph.graphID)
            graphs.append(graph)
        }
        let reading: Partial? =
            if links.notCounted { .notCounted } else if links.linesDropped || graphsCut { .leftOut } else { nil }
        partial = reading
        // Go sends at most `persondetail.maxRecordLinkLines` lines, strong first; every line it sends is an arc, in the read's order.
        var arcs: [CircularLinkArc<String>] = []
        for pair in links.pairs where Self.isKept(pair, in: ids) {
            for line in pair.lines {
                arcs.append(CircularLinkArc(
                    from: pair.graphAID, to: pair.graphBID,
                    lineWidth: line.strength == .strong ? Self.strongLineWidth : Self.weakLineWidth
                ))
            }
        }
        let graphsByID = Dictionary(uniqueKeysWithValues: graphs.map { ($0.graphID, $0) })
        let bare = CircularLinkDataset(
            nodes: graphs.map { CircularLinkNode(id: $0.graphID, label: "") }, arcs: arcs
        )
        dataset = bare.mappingNodes { _, node in
            guard let graph = graphsByID[node.id] else { return node }
            let name = graph.graphName.trimmedNonEmpty
            let ref = graph.graphRef.trimmedNonEmpty
            let graphTitle = name ?? ref ?? ""
            let sourceTitle = graph.sourceTitle.trimmedNonEmpty ?? graphSourceRefs[graph.graphID] ?? ""
            let label = ref ?? name ?? ""
            let connected = bare.neighborCount(of: graph.graphID)
            return CircularLinkNode(
                id: graph.graphID,
                label: label,
                help: L10n.Conclusions.recordLinksNodeHelp(graph: graphTitle, source: sourceTitle),
                // A node with no line in a drawing that left some out may have lost its lines.
                accessibilityLabel: reading == .notCounted || (reading == .leftOut && connected == 0)
                    ? L10n.Conclusions.recordLinksNodeUncountedAccessibility(ref: label, graph: graphTitle, source: sourceTitle)
                    : connected == 0
                    ? L10n.Conclusions.recordLinksNodeApartAccessibility(ref: label, graph: graphTitle, source: sourceTitle)
                    : L10n.Conclusions.recordLinksNodeAccessibility(
                        ref: label, graph: graphTitle, source: sourceTitle, connected: connected
                    ),
                isEnabled: graphLocations[graph.graphID] != nil
            )
        }
        locations = graphLocations.filter { ids.contains($0.key) }
    }

    private static func isKept(_ pair: CatalogRecordLinksPair, in ids: Set<String>) -> Bool {
        pair.graphAID != pair.graphBID && ids.contains(pair.graphAID) && ids.contains(pair.graphBID)
    }
}
