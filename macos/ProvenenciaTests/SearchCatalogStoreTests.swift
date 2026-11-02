import Foundation
import Testing
@testable import Provenencia

@Suite
struct SearchCatalogStoreTests {
    private let projectDir = "/tmp/search.provenencia"

    private func hit(kind: String, id: String) -> CatalogSearchHit {
        CatalogSearchHit(
            kind: kind, id: id, ref: id.uppercased(), title: id, subtitle: "", matchReason: "title",
            location: WorkspaceLocation(section: .sources)
        )
    }

    private func seededStore() -> FakeStore {
        let store = FakeStore()
        store.seededSearchHits = [
            hit(kind: "source", id: "s1"), hit(kind: "person", id: "p1"), hit(kind: "place", id: "l1"),
        ]
        return store
    }

    @Test func returnsSeededHitsInSeededOrder() async throws {
        let store = seededStore()
        let hits = try await store.searchCatalog(
            projectDir: projectDir, query: "anything", location: .sectionRoot(.sources), kinds: []
        )
        #expect(hits.map(\.id) == ["s1", "p1", "l1"])
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func filtersSeededHitsByRequestedKinds() async throws {
        let store = seededStore()
        let some = try await store.searchCatalog(
            projectDir: projectDir, query: "anything", location: nil, kinds: ["person", "place"]
        )
        #expect(some.map(\.id) == ["p1", "l1"])
        let none = try await store.searchCatalog(
            projectDir: projectDir, query: "anything", location: nil, kinds: ["nonsense"]
        )
        #expect(none.isEmpty)
    }

    @Test func emptyQueryReturnsNoHits() async throws {
        let store = seededStore()
        let hits = try await store.searchCatalog(
            projectDir: projectDir,
            query: "   ",
            location: .sectionRoot(.sources),
            kinds: []
        )
        #expect(hits.isEmpty)
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func searchErrorIsThrown() async {
        let store = seededStore()
        store.searchCatalogError = CoreInvokeError.coded(status: 1, code: "catalog.closed", kind: .user, params: [])
        await #expect(throws: CoreInvokeError.self) {
            _ = try await store.searchCatalog(
                projectDir: projectDir, query: "anything", location: nil, kinds: []
            )
        }
    }

    @Test func entityIdRoundTripsThroughTheEngineLocation() {
        let location = WorkspaceLocation(section: .persons, entityId: "e-1", ref: "PER-1", title: "James")
        var proto = Provenencia_Engine_V1_WorkspaceLocation()
        proto.section = "persons"
        proto.entityID = "e-1"
        #expect(GoStore.mapWorkspaceLocationFromProto(proto) == location)
    }
}
