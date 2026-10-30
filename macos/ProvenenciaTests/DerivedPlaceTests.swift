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

    @Test func shownSkipsANamelessPlace() {
        let nameless = CatalogHeaderPlace(
            entity: CatalogCanonicalEntity(id: "p0", ref: "PLC-0", subjectTypeID: "t", label: ""), names: []
        )
        let paris = CatalogHeaderPlace(
            entity: CatalogCanonicalEntity(id: "p1", ref: "PLC-1", subjectTypeID: "t", label: ""), names: ["Paris"]
        )
        #expect(DerivedPlace.shown([nameless, paris])?.entity.id == "p1")
        #expect(DerivedPlace.shown([nameless]) == nil)
        #expect(DerivedPlace.name([nameless, paris]) == "Paris")
        #expect(DerivedPlace.extra([nameless, paris]) == 1)
        let row = PersonDetailContent.rows([], header: CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-1", subjectTypeID: "t", label: ""),
            death: CatalogLifeFacts(places: [nameless, paris])
        ), locale: Locale(identifier: "en_US"))[4]
        #expect(row.opens?.location == .placeDetail(entityId: "p1", ref: "PLC-1", title: "Paris"))
    }
}
