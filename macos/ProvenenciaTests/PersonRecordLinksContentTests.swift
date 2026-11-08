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
        #expect(content.nodes.map(\.label) == [
            "EVG-2K7PQ", "EVG-5MD8R", "EVG-8TZ4N", "EVG-1QR9D", "EVG-6WB3M", "EVG-4PL7T", "EVG-9CN2H",
        ])
        #expect(content.nodes.map(\.connected) == [3, 2, 0, 2, 2, 1, 0])
        #expect(content.nodes.allSatisfy { $0.location != nil })
        #expect(content.nodes[0].help == "Household 214 — Hamburg census 1867")
    }

    @Test func johannHasEightArcsWithSlotsOnSharedPairs() {
        let arcs = johann().arcs
        #expect(arcs.count == 8)
        #expect(arcs.map { [$0.a, $0.b] } == [[0, 1], [0, 1], [0, 3], [0, 3], [1, 3], [0, 4], [4, 5], [4, 5]])
        #expect(arcs.map(\.slot) == [0, 1, 0, 1, 0, 0, 0, 1])
        #expect(arcs.map(\.slots) == [2, 2, 2, 2, 1, 1, 2, 2])
        #expect(arcs.map(\.strong) == [true, false, true, true, false, true, false, false])
    }

    @Test func neighborsAreTheOtherEndsOfANodesArcs() {
        let content = johann()
        #expect(content.neighbors(of: 0) == [1, 3, 4])
        #expect(content.neighbors(of: 2).isEmpty)
    }

    @Test func aNodeThatSharesNoOneIsSpokenAsApart() {
        let content = johann()
        #expect(content.nodes[2].accessibilityLabel == "EVG-8TZ4N, Volume 3, Hamburger Adressbuch 1861, shares no one")
        #expect(content.nodes[0].accessibilityLabel
            == "EVG-2K7PQ, Household 214, Hamburg census 1867, connected to 3 records")
        #expect(content.nodes[5].accessibilityLabel.hasSuffix("connected to 1 record"))
    }

    @Test func aGraphWithNoRefIsLabelledWithItsName() {
        let links = CatalogPersonRecordLinks(
            graphs: [CatalogRecordLinksGraph(
                graphID: "g", graphRef: "", graphName: "Page 4", sourceID: "s", sourceTitle: "", sourceTypeID: "t"
            )],
            pairs: []
        )
        let content = PersonRecordLinksContent(links: links, graphLocations: [:], graphSourceRefs: ["g": "SRC-1"])
        #expect(content.nodes[0].label == "Page 4")
        #expect(content.nodes[0].help == "Page 4 — SRC-1")
    }

    @Test func aGraphWithNoLocationIsDrawnAndNotClickable() {
        let content = johann(removing: "g-388")
        #expect(content.nodes.count == 7)
        #expect(content.nodes[1].location == nil)
        #expect(content.nodes[0].location != nil)
    }

    @Test func aDuplicateGraphIDKeepsTheFirst() {
        var links = PersonSourcesPreviewData.johannRecordLinks
        links.graphs.append(links.graphs[0])
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        #expect(content.nodes.count == 7)
    }

    @Test func aPairNamingAnUnknownGraphIsDropped() {
        var links = PersonSourcesPreviewData.johannRecordLinks
        links.pairs.append(CatalogRecordLinksPair(
            graphAID: "g-214", graphBID: "g-gone", lines: [CatalogRecordLink(entityID: "x", strength: .strong)]
        ))
        let content = PersonRecordLinksContent(links: links, graphLocations: [:])
        #expect(content.arcs.count == 8)
    }

    @Test func manyHasFiftyNodesAndNinetyOneArcs() {
        let sources = PersonSourcesPreviewData.manyClaims(graphs: 50)
        #expect(sources.count == 14)
        #expect(sources[0].claims.count == 9)
        let content = PersonRecordLinksContent(
            links: PersonSourcesPreviewData.manyRecordLinks(graphs: 50), graphLocations: [:]
        )
        #expect(content.nodes.count == 50)
        #expect(content.arcs.count == 91)
        #expect(content.nodes.suffix(3).allSatisfy { $0.connected == 0 })
    }
}
