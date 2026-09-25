import Foundation
import Testing
@testable import Provenencia

@Suite
struct ConnectEndpointBindingTests {
    @Test func bindMapsDistinctTypesByEndpoint() {
        let rule = CatalogConnectRule.productMatrix.first {
            $0.bridgeTypeKey == "participation" && $0.fromTypeKey == "person"
        }!
        let bound = ConnectEndpointBinding.bind(
            rule: rule,
            fromID: "p1",
            fromTypeKey: "person",
            fromLabel: "Ada",
            toID: "e1",
            toTypeKey: "event",
            toLabel: "Birth"
        )
        #expect(bound?.map(\.subjectID) == ["p1", "e1"])
        #expect(bound?.map(\.propertyKey) == ["person", "event"])
    }

    @Test func bindMapsSameTypeByOrder() {
        let rule = CatalogConnectRule.productMatrix.first {
            $0.bridgeTypeKey == "relationship"
        }!
        let bound = ConnectEndpointBinding.bind(
            rule: rule,
            fromID: "p1",
            fromTypeKey: "person",
            fromLabel: "Ada",
            toID: "p2",
            toTypeKey: "person",
            toLabel: "Ben"
        )
        #expect(bound?.map(\.subjectID) == ["p1", "p2"])
        #expect(bound?.map(\.propertyKey) == ["person", "related_to"])
    }

    @Test func pendingSentenceUsesBoundEndpoints() {
        let rule = CatalogConnectRule.productMatrix.first {
            $0.bridgeTypeKey == "participation" && $0.fromTypeKey == "person"
        }!
        let sentence = ConnectEndpointBinding.pendingSentence(
            rule: rule,
            fromID: "p1",
            fromTypeKey: "person",
            fromLabel: "Ada",
            toID: "e1",
            toTypeKey: "event",
            toLabel: "Birth"
        )
        #expect(
            sentence == L10n.EvidenceGraph.bridgeSummaryParticipationFallback(
                person: "Ada",
                event: "Birth"
            )
        )
    }
}
