import Foundation
import Testing
@testable import Provenencia

@Suite
struct DerivedPlaceTests {
    @Test(arguments: [
        (names: [["York", "Tkaronto"]], name: "York", extra: 0),
        (names: [["York"], ["Toronto"]], name: "York", extra: 1),
        (names: [[], ["Toronto"]], name: "Toronto", extra: 1),
        (names: [[String]](), name: "", extra: 0),
        (names: [[]], name: "", extra: 1),
    ])
    func derivedPlaceCountsPlacesNotNames(names: [[String]], name: String, extra: Int) {
        let places = names.enumerated().map { index, names in
            CatalogHeaderPlace(
                entity: CatalogCanonicalEntity(id: "p\(index)", ref: "PLC-\(index)", subjectTypeID: "t", label: ""),
                names: names
            )
        }
        #expect(DerivedPlace.name(places) == name)
        #expect(DerivedPlace.extra(places) == extra)
    }

    @Test func placeLineIncludesParents() {
        let place = CatalogHeaderPlace(
            entity: CatalogCanonicalEntity(id: "p0", ref: "PLC-0", subjectTypeID: "t", label: ""),
            names: ["Toronto"],
            parents: ["Ontario", "Canada"]
        )
        #expect(DerivedPlace.name([place]) == "Toronto, Ontario, Canada")
        #expect(PlaceChainDisplay.placeLine(place) == "Toronto, Ontario, Canada")
    }

    @Test func placeLineJoinsCandidateParentsWithOr() {
        let place = CatalogHeaderPlace(
            entity: CatalogCanonicalEntity(id: "p0", ref: "PLC-0", subjectTypeID: "t", label: ""),
            names: ["The Farm"],
            parents: ["York Township", "York County"],
            parentsAreCandidates: true
        )
        #expect(DerivedPlace.name([place]) == "The Farm, York Township or York County")
    }
}
