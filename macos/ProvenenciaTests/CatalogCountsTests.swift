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
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "1", key: "photograph", origin: "provenencia", label: "Photograph", description: ""),
            CatalogSourceType(id: "2", key: "scrapbook", origin: "user", label: "Scrapbook", description: ""),
            CatalogSourceType(id: "3", key: "grave", origin: "plugin:findagrave", label: "Grave", description: ""),
        ]
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(id: "1", key: "date_taken", origin: "provenencia", label: "Date taken", dataType: "date", description: ""),
            CatalogMetadataField(id: "2", key: "notes", origin: "user", label: "Notes", dataType: "text", description: ""),
        ]
        let counts = makeCounts(store: store)
        await counts.refreshAll()

        #expect(counts.sources == 2)
        #expect(counts.sourceTypes?.total == 3)
        #expect(counts.sourceTypes?.seeded == 1)
        #expect(counts.sourceTypes?.user == 1)
        #expect(counts.sourceTypes?.plugin == 1)
        #expect(counts.sourceFields?.total == 2)
        #expect(counts.sourceFields?.seeded == 1)
        #expect(counts.sourceFields?.user == 1)
        #expect(counts.sourceFields?.plugin == 0)
        #expect(counts.badge(for: .sources) == 2)
        #expect(counts.badge(for: .sourceTypes) == 3)
        #expect(counts.badge(for: .sourceFields) == 2)
    }

    @Test func refreshAllOnEmptyProjectYieldsZero() async {
        let counts = makeCounts()
        await counts.refreshAll()
        #expect(counts.sources == 0)
        #expect(counts.sourceTypes?.total == 0)
        #expect(counts.sourceFields?.total == 0)
        #expect(counts.lastRefreshError == nil)
    }

    @Test func refreshAllFailureSetsLastRefreshErrorAndLeavesBadgesUnset() async {
        let store = FakeStore()
        enum Boom: Error { case boom }
        store.workspaceNavCountsError = Boom.boom

        let counts = makeCounts(store: store)
        await counts.refreshAll()

        #expect(counts.lastRefreshError == String(localized: L10n.Errors.unknown))
        #expect(counts.sources == nil)
        #expect(counts.sourceTypes == nil)
        #expect(counts.sourceFields == nil)
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

    @Test func publishReplacesVocabularySummariesWithoutStoreRoundTrip() {
        let counts = makeCounts()
        counts.publishSourceFields(CatalogCountSummary(total: 5, seeded: 2, user: 2, plugin: 1))
        counts.publishSourceTypes(CatalogCountSummary(total: 3, seeded: 1, user: 2, plugin: 0))
        counts.publishSources(7)
        #expect(counts.badge(for: .sourceFields) == 5)
        #expect(counts.sourceFields?.plugin == 1)
        #expect(counts.badge(for: .sourceTypes) == 3)
        #expect(counts.sourceTypes?.user == 2)
        #expect(counts.badge(for: .sources) == 7)
    }

    @Test func summaryFromRowsSplitsByOrigin() {
        let rows = [
            CatalogMetadataField(id: "1", key: "a", origin: "provenencia", label: "A", dataType: "text", description: ""),
            CatalogMetadataField(id: "2", key: "b", origin: "user", label: "B", dataType: "text", description: ""),
            CatalogMetadataField(id: "3", key: "c", origin: "plugin:x", label: "C", dataType: "text", description: ""),
        ]
        let summary = CatalogCountSummary.from(rows)
        #expect(summary == CatalogCountSummary(total: 3, seeded: 1, user: 1, plugin: 1))
    }
}
