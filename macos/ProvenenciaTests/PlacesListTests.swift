import Foundation
import Testing
@testable import Provenencia

struct PlacesListTests {

    private func header(
        _ ref: String,
        names: [String] = [],
        label: String = "",
        parents: [String] = []
    ) -> CatalogPlaceHeader {
        CatalogPlaceHeader(
            entity: CatalogCanonicalEntity(id: "id-\(ref)", ref: ref, subjectTypeID: "t", label: label),
            names: names,
            parents: parents
        )
    }

    @Test func headerMetaCountsAndShowsRefreshing() {
        let one = [header("PLC-1", names: ["York"])]
        let two = one + [header("PLC-2")]
        #expect(ConclusionListPresentation.rows(one, refreshing: false).meta(
            count: L10n.Workspace.placeCount, refreshing: L10n.Workspace.placeCountRefreshing
        ) == "1 place · by name")
        #expect(ConclusionListPresentation.rows(two, refreshing: false).meta(
            count: L10n.Workspace.placeCount, refreshing: L10n.Workspace.placeCountRefreshing
        ) == "2 places · by name")
        #expect(ConclusionListPresentation.rows(two, refreshing: true).meta(
            count: L10n.Workspace.placeCount, refreshing: L10n.Workspace.placeCountRefreshing
        ) == "2 places · by name · refreshing")
        #expect(ConclusionListPresentation<[CatalogPlaceHeader]>.empty.meta(
            count: L10n.Workspace.placeCount, refreshing: L10n.Workspace.placeCountRefreshing
        ) == nil)
    }

    @Test func rowAccessibilityNamesTheExtraAndTheChain() {
        #expect(L10n.Workspace.placeRowAccessibility(
            title: "Montréal", extra: 0, chain: "", ref: "PLC-2MT40"
        ) == "Montréal, PLC-2MT40")
        #expect(L10n.Workspace.placeRowAccessibility(
            title: "Montréal", extra: 1, chain: "", ref: "PLC-2MT40"
        ) == "Montréal, and 1 other name, PLC-2MT40")
        #expect(L10n.Workspace.placeRowAccessibility(
            title: "Montréal", extra: 2, chain: "", ref: "PLC-2MT40"
        ) == "Montréal, and 2 other names, PLC-2MT40")
        #expect(L10n.Workspace.placeRowAccessibility(
            title: "Montréal", extra: 1, chain: "Québec, Canada", ref: "PLC-2MT40"
        ) == "Montréal, and 1 other name, in Québec, Canada, PLC-2MT40")
    }

    @Test func chainLineJoinsParentsAndOmitsAnEmptyChain() {
        #expect(PlaceChainDisplay.line(parents: []) == "")
        #expect(PlaceChainDisplay.line(parents: [" ", ""]) == "")
        #expect(PlaceChainDisplay.line(parents: [" Ontario ", "Canada"]) == "Ontario, Canada")
    }

    @Test func rowOpensThePlaceDetailPlace() {
        let place = header("PLC-1", names: ["York"])
        let location = WorkspaceLocation.placeDetail(
            entityId: place.entity.id,
            ref: place.entity.ref,
            title: PlaceTitleDisplay.titleSource(place.titleParts).text
        )
        #expect(location.section == .places)
        #expect(location.entityId == "id-PLC-1")
        let resolved = PlaceRegistry.standard.resolve(location, project: ProjectKey(projectDir: "/tmp/p.provenencia"))
        #expect(resolved?.placeID == .placeDetail)
        #expect(resolved?.queryKeys == [
            .conclusionDetail(project: ProjectKey(projectDir: "/tmp/p.provenencia"), entityId: "id-PLC-1"),
        ])
        #expect(PlaceChainDisplay.line(parents: place.parents) == "")
    }
}
