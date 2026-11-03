import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct CatalogCountsTests {
    private let projectDir = "/tmp/counts.provenencia"

    private func makeCounts(store: FakeStore = FakeStore()) -> CatalogCounts {
        CatalogCounts(projectDir: projectDir, store: store)
    }

    @Test func refreshAllPopulatesEveryDestination() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "1", ref: "SRC-1", sourceTypeID: "", title: "A", description: ""),
            CatalogSource(id: "2", ref: "SRC-2", sourceTypeID: "", title: "B", description: ""),
        ]
        let counts = makeCounts(store: store)
        await counts.refreshAll()

        #expect(counts.sources == 2)
        #expect(counts.persons == 0)
        #expect(counts.events == 0)
        #expect(counts.places == 0)
        #expect(counts.badge(for: .sources) == 2)
        #expect(counts.badge(for: .persons) == 0)
        #expect(counts.badge(for: .events) == 0)
        #expect(counts.badge(for: .places) == 0)
        // Configure rows draw no badge.
        #expect(counts.badge(for: .sourceTypes) == nil)
        #expect(counts.badge(for: .metadata) == nil)
        #expect(counts.badge(for: .properties) == nil)
    }

    @Test func refreshAllCountsPromotedPersons() async throws {
        let store = FakeStore()
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.subjectsByGraph["g1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
            CatalogSubject(id: "sub-2", ref: "CPR-2", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
        ]
        _ = store.seedPromotion(projectDir: projectDir, subjectID: "sub-1")
        _ = store.seedPromotion(projectDir: projectDir, subjectID: "sub-2")
        let counts = makeCounts(store: store)
        await counts.refreshAll()
        #expect(counts.persons == 2)
        #expect(counts.badge(for: .persons) == 2)
    }

    @Test func refreshAllOnEmptyProjectYieldsZero() async {
        let counts = makeCounts()
        await counts.refreshAll()
        #expect(counts.sources == 0)
        #expect(counts.persons == 0)
        #expect(counts.events == 0)
        #expect(counts.places == 0)
        #expect(counts.lastRefreshError == nil)
    }

    @Test func refreshAllFailureSetsLastRefreshErrorAndLeavesBadgesUnset() async {
        let store = FakeStore()
        enum Boom: Error { case boom }
        store.workspaceNavCountsError = Boom.boom

        let counts = makeCounts(store: store)
        await counts.refreshAll()

        #expect(counts.lastRefreshError == L10n.string(L10n.Errors.unknown))
        #expect(counts.sources == nil)
        #expect(counts.badge(for: .sources) == nil)
    }

    @Test func refreshAllSuccessClearsLastRefreshError() async {
        let store = FakeStore()
        enum Boom: Error { case boom }
        store.workspaceNavCountsError = Boom.boom
        let counts = makeCounts(store: store)
        await counts.refreshAll()
        #expect(counts.lastRefreshError != nil)

        store.workspaceNavCountsError = nil
        await counts.refreshAll()
        #expect(counts.lastRefreshError == nil)
        #expect(counts.sources == 0)
    }

    /// Two recounts overlap: the one started last reads the newest catalog, so the earlier,
    /// slower one finishing after it must not overwrite its counts.
    @Test func newestRecountWins() async throws {
        let store = FakeStore()
        store.workspaceNavCountsHold.isOn = true
        let counts = makeCounts(store: store)

        let first = Task { await counts.refreshAll() }
        // The first recount has read zero Sources and is held before returning them.
        try #require(await waitUntil(timeout: .seconds(15)) { store.workspaceNavCountsHold.count == 1 })
        store.workspaceNavCountsHold.isOn = false
        _ = try await store.createSource(
            projectDir: projectDir, userID: "u", sourceTypeID: "t1", title: "New", description: ""
        )

        await counts.refreshAll()
        #expect(counts.sources == 1)
        store.workspaceNavCountsHold.release()
        await first.value
        #expect(counts.sources == 1)
    }
}
