import Foundation
import Testing
@testable import Provenencia

struct PersonRecordLinksContentTests {
    private func johann(removing graphID: String? = nil) -> PersonRecordLinksContent {
        let sources = PersonSourcesContent(
            groups: PersonSourcesPreviewData.johannClaims, sourceTypes: PersonSourcesPreviewData.types,
            orders: NameFormatOrders(), expanded: []
        )
        var locations = sources.graphLocations
        if let graphID { locations[graphID] = nil }
        return PersonRecordLinksContent(
            links: PersonSourcesPreviewData.johannRecordLinks,
            graphLocations: locations,
            graphSourceRefs: sources.graphSourceRefs
        )
    }

    @Test func johannHasSevenNodesInTheReadsOrder() {
        let content = johann()
        #expect(content.dataset.nodes.map(\.label) == [
            "EVG-2K7PQ", "EVG-5MD8R", "EVG-8TZ4N", "EVG-1QR9D", "EVG-6WB3M", "EVG-4PL7T", "EVG-9CN2H",
        ])
        #expect(content.dataset.nodes.map { content.dataset.neighborCount(of: $0.id) } == [3, 2, 0, 2, 2, 1, 0])
        #expect(content.dataset.nodes.allSatisfy { $0.isEnabled })
        #expect(content.dataset.nodes[0].help == "Household 214 — Hamburg census 1867")
    }

    @Test func johannHasEightArcsOnSharedPairs() {
        let content = johann()
        let arcs = content.dataset.resolvedArcs
        #expect(arcs.count == 8)
        #expect(arcs.map { [$0.from, $0.to] } == [[0, 1], [0, 1], [0, 3], [0, 3], [1, 3], [0, 4], [4, 5], [4, 5]])
        #expect(arcs.map { $0.lineWidth == PersonRecordLinksContent.strongLineWidth }
            == [true, false, true, true, false, true, false, false])
    }

    @Test func strongAndWeakLinesMapToTheBoardsWidths() {
        #expect(PersonRecordLinksContent.strongLineWidth == 3)
        #expect(PersonRecordLinksContent.weakLineWidth == 1.25)
        #expect(Set(johann().dataset.resolvedArcs.map(\.lineWidth)) == [3, 1.25])
    }

    @Test func aNodeThatSharesNoOneIsSpokenAsApart() {
        let content = johann()
        #expect(content.dataset.nodes[2].accessibilityLabel == "EVG-8TZ4N, Volume 3, Hamburger Adressbuch 1861, shares no one")
        #expect(content.dataset.nodes[0].accessibilityLabel
            == "EVG-2K7PQ, Household 214, Hamburg census 1867, connected to 3 records")
        #expect(content.dataset.nodes[5].accessibilityLabel?.hasSuffix("connected to 1 record") == true)
    }

    @Test func aGraphWithNoRefIsLabelledWithItsName() {
        let links = CatalogPersonRecordLinks(
            graphs: [CatalogRecordLinksGraph(
                graphID: "g", graphRef: "", graphName: "Page 4", sourceID: "s", sourceTitle: "", sourceTypeID: "t"
            )],
            pairs: []
        )
        let content = PersonRecordLinksContent(links: links, graphLocations: [:], graphSourceRefs: ["g": "SRC-1"])
        #expect(content.dataset.nodes[0].label == "Page 4")
        #expect(content.dataset.nodes[0].help == "Page 4 — SRC-1")
    }

    @Test func aGraphWithNoLocationIsDrawnAndNotClickable() {
        let content = johann(removing: "g-388")
        #expect(content.dataset.nodes.count == 7)
        #expect(content.dataset.nodes[1].isEnabled == false)
        #expect(content.dataset.nodes[0].isEnabled)
        #expect(content.locations["g-388"] == nil)
    }

    @Test func aDuplicateGraphIDKeepsTheFirst() {
        var links = PersonSourcesPreviewData.johannRecordLinks
        links.graphs.append(links.graphs[0])
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        #expect(content.dataset.nodes.count == 7)
    }

    @Test func aPairNamingAnUnknownGraphIsDropped() {
        var links = PersonSourcesPreviewData.johannRecordLinks
        links.pairs.append(CatalogRecordLinksPair(
            graphAID: "g-214", graphBID: "g-gone", lines: [CatalogRecordLink(entityID: "x", strength: .strong)]
        ))
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        #expect(content.dataset.resolvedArcs.count == 8)
    }

    @Test func manyHasFiftyNodesAndNinetyOneArcs() {
        let sources = PersonSourcesPreviewData.manyClaims(graphs: 50)
        #expect(sources.count == 14)
        #expect(sources[0].claims.count == 9)
        let content = PersonRecordLinksContent(
            links: PersonSourcesPreviewData.manyRecordLinks(graphs: 50), graphLocations: [:]
        )
        #expect(content.dataset.nodes.count == 50)
        #expect(content.dataset.resolvedArcs.count == 91)
        #expect(content.dataset.nodes.suffix(3).allSatisfy { content.dataset.neighborCount(of: $0.id) == 0 })
    }

    @Test func householdScale() {
        let content = PersonRecordLinksContent(
            links: PersonSourcesPreviewData.householdRecordLinks(graphs: 50), graphLocations: [:]
        )
        #expect(content.dataset.nodes.count == 50)
        #expect(content.dataset.resolvedArcs.count == 3_675)
    }

    @Test func allLinesDrawAndNodesStopAtTheCap() {
        func graph(_ id: String) -> CatalogRecordLinksGraph {
            CatalogRecordLinksGraph(graphID: id, graphRef: id, graphName: "", sourceID: "s", sourceTitle: "", sourceTypeID: "t")
        }
        // Go owns the line bound: every line it sends is an arc, however many.
        let lines = (0..<80_000).map { CatalogRecordLink(entityID: "e\($0)", strength: .weak) }
        let many = PersonRecordLinksContent(
            links: CatalogPersonRecordLinks(
                graphs: [graph("a"), graph("b")],
                pairs: [CatalogRecordLinksPair(graphAID: "a", graphBID: "b", lines: lines)]
            ),
            graphLocations: [:]
        )
        #expect(many.dataset.resolvedArcs.count == 80_000)

        let ids = (0..<(PersonRecordLinksContent.maxNodes + 5)).map { "g\($0)" }
        let wide = PersonRecordLinksContent(
            links: CatalogPersonRecordLinks(
                graphs: ids.map(graph),
                pairs: [
                    CatalogRecordLinksPair(graphAID: "g0", graphBID: "g1", lines: [CatalogRecordLink(entityID: "x", strength: .weak)]),
                    CatalogRecordLinksPair(
                        graphAID: "g0", graphBID: ids[PersonRecordLinksContent.maxNodes + 2],
                        lines: [CatalogRecordLink(entityID: "y", strength: .strong)]
                    ),
                ]
            ),
            graphLocations: [:]
        )
        #expect(wide.dataset.nodes.count == PersonRecordLinksContent.maxNodes)
        #expect(wide.dataset.resolvedArcs.count == 1)
        #expect(wide.dataset.resolvedArcs.map { [wide.dataset.nodes[$0.from].id, wide.dataset.nodes[$0.to].id] } == [["g0", "g1"]])
    }

    @Test func catalogRecordLinksDefaultsToAnOrdinaryRead() {
        let links = CatalogPersonRecordLinks()
        #expect(!links.notCounted)
        #expect(!links.linesDropped)
    }

    @Test func goStoreCarriesThePartialFlags() {
        var resp = Provenencia_Engine_V1_GetPersonRecordLinksResponse()
        resp.notCounted = true
        resp.linesDropped = true
        let partial = GoStore.mapPersonRecordLinks(resp)
        #expect(partial.notCounted)
        #expect(partial.linesDropped)
        let ordinary = GoStore.mapPersonRecordLinks(Provenencia_Engine_V1_GetPersonRecordLinksResponse())
        #expect(!ordinary.notCounted)
        #expect(!ordinary.linesDropped)
    }
}
