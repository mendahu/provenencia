import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct OmnibarResultsModelTests {
    private let projectDir = "/tmp/omnibar-results.provenencia"

    @Test func shortQueryDoesNotPresent() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-AAAAA", sourceTypeID: "", title: "Ilminster", description: ""),
        ]
        let model = OmnibarResultsModel()
        model.search.query = "I"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        try? await Task.sleep(nanoseconds: 250_000_000)
        #expect(!model.isPresented)
        #expect(model.search.hits.isEmpty)
    }

    @Test func debouncedSearchReturnsHits() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-ILMIN", sourceTypeID: "t1", title: "Ilminster parish", description: ""),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t1", key: "parish", origin: "provenencia", label: "Parish", description: "", iconKey: "type_certificate"),
        ]
        let model = OmnibarResultsModel()
        model.search.query = "Ilminster"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model.search)
        #expect(model.isPresented)
        #expect(model.search.hits.first?.id == "s1")
        #expect(model.search.hits.first?.subtitle == "Parish")
        #expect(model.search.hits.first?.iconKey == "type_certificate")
    }

    @Test func clearAfterNavigateResetsQuery() async {
        let model = OmnibarResultsModel()
        model.search.query = "abc"
        model.search.hasSearched = true
        model.search.hits = [
            CatalogSearchHit(
                kind: "source",
                id: "s1",
                ref: "SRC-AAAAA",
                title: "A",
                subtitle: "",
                matchReason: "title",
                location: .sectionRoot(.sources)
            ),
        ]
        model.clearAfterNavigate()
        #expect(model.search.query.isEmpty)
        #expect(!model.isPresented)
        #expect(model.search.hits.isEmpty)
    }

    @Test func activateCommitsNavigationAndClearsResults() {
        let hit = CatalogSearchHit(
            kind: "source",
            id: "s1",
            ref: "SRC-AAAAA",
            title: "Deed",
            subtitle: "Photograph",
            matchReason: "title",
            location: WorkspaceLocation(section: .sources, sourceId: "s1", ref: "SRC-AAAAA", title: "Deed")
        )
        let navigation = WorkspaceNavigation()
        let model = OmnibarResultsModel()
        model.search.query = "Deed"
        model.search.hits = [hit]
        model.search.hasSearched = true

        model.activate(hit, navigation: navigation)

        #expect(navigation.currentLocation.sourceId == "s1")
        #expect(navigation.selectedSection == .sources)
        #expect(model.search.query.isEmpty)
        #expect(model.search.hits.isEmpty)
        #expect(!model.search.hasSearched)
    }

    @Test func activateFieldHitFromSearch() async {
        let store = FakeStore()
        let projectDir = "/tmp/omnibar-activate.provenencia"
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f9", key: "author", origin: "provenencia",
                label: "Author", dataType: "text", description: ""
            ),
        ]
        let model = OmnibarResultsModel()
        model.search.query = "Author"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.metadata),
            store: store
        )
        await waitForSearch(model.search)
        guard let hit = model.search.hits.first(where: { $0.id == "f9" }) else {
            Issue.record("no f9 hit")
            return
        }

        let navigation = WorkspaceNavigation()
        model.activate(hit, navigation: navigation)

        #expect(navigation.currentLocation.fieldId == "f9")
        #expect(navigation.selectedSection == .metadata)
        #expect(model.search.query.isEmpty)
    }

    @Test func activateEvidenceGraphHitOpensTheGraph() async {
        let store = FakeStore()
        let projectDir = "/tmp/omnibar-graph.provenencia"
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: "src-w3", ref: "SRC-W3CEN", sourceTypeID: "", title: "1881 Census, Ward 3", description: ""
            ),
        ]
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: "g-jones", ref: "GPH-J0NES", sourceID: "src-w3", name: "Jones household",
                description: "", artifactIDs: []
            ),
        ]
        let model = OmnibarResultsModel()
        model.search.query = "jones"
        model.scheduleSearch(projectDir: projectDir, location: .sectionRoot(.sources), store: store)
        await waitForSearch(model.search)
        guard let hit = model.search.hits.first(where: { $0.kind == "evidence_graph" }) else {
            Issue.record("no graph hit")
            return
        }

        let navigation = WorkspaceNavigation()
        model.activate(hit, navigation: navigation)

        #expect(navigation.currentLocation.graphId == "g-jones")
        #expect(navigation.currentLocation.sourceId == "src-w3")
        #expect(navigation.currentLocation.sourceSurface == .graph)
        #expect(navigation.selectedSection == .sources)
        #expect(model.search.query.isEmpty)
    }

    @Test func closePanelSetsUserDismissedWithoutClearingQuery() {
        let model = OmnibarResultsModel()
        model.search.query = "Ilminster"
        model.search.hasSearched = true
        model.search.isLoading = true
        model.closePanel()
        #expect(model.userDismissed)
        #expect(!model.isPresented)
        #expect(model.search.query == "Ilminster")
        #expect(!model.search.isLoading)
    }

    @Test func loadingPresentsBeforeHitsArrive() {
        let model = OmnibarResultsModel()
        model.search.query = "Il"
        model.search.isLoading = true
        #expect(model.isPresented)
        #expect(model.search.showsLoadingState)
        #expect(!model.search.showsEmptyState)
        #expect(!model.search.showsErrorState)
    }

    @Test func searchFailureSurfacesMappedErrorNotEmpty() async {
        let store = FakeStore()
        store.searchCatalogError = CoreInvokeError.coded(
            status: 1,
            code: "catalog.closed",
            kind: .user,
            params: []
        )
        let model = OmnibarResultsModel()
        model.search.query = "Ilminster"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model.search)
        #expect(model.search.searchError == L10n.Errors.message(code: "catalog.closed"))
        #expect(model.search.hits.isEmpty)
        #expect(model.search.showsErrorState)
        #expect(!model.search.showsEmptyState)
        #expect(model.isPresented)
    }
}
