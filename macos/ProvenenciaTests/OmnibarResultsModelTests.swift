import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct OmnibarResultsModelTests {
    private let projectDir = "/tmp/omnibar-results.provenencia"

    private func waitForSearch(_ model: OmnibarResultsModel) async {
        for _ in 0..<80 {
            if !model.isLoading, model.hasSearched || model.searchError != nil {
                return
            }
            try? await Task.sleep(nanoseconds: 25_000_000)
        }
    }

    @Test func shortQueryDoesNotPresent() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-AAAAA", sourceTypeID: "", title: "Ilminster", description: ""),
        ]
        let model = OmnibarResultsModel()
        model.query = "I"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        try? await Task.sleep(nanoseconds: 250_000_000)
        #expect(!model.isPresented)
        #expect(model.hits.isEmpty)
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
        model.query = "Ilminster"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model)
        #expect(model.isPresented)
        #expect(model.hits.first?.id == "s1")
        #expect(model.hits.first?.subtitle == "Parish")
        #expect(model.hits.first?.iconKey == "type_certificate")
    }

    @Test func exactRefHitIsTop() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-ZZ9K2", sourceTypeID: "", title: "Quiet", description: ""),
            CatalogSource(id: "s2", ref: "SRC-OTHER", sourceTypeID: "", title: "SRC-ZZ9K2 in title", description: ""),
        ]
        let model = OmnibarResultsModel()
        model.query = "SRC-ZZ9K2"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model)
        #expect(model.hits.first?.id == "s1")
        #expect(model.hits.first?.matchReason == "ref")
        if let top = model.hits.first {
            #expect(OmnibarHitPresentation.refAccent(for: top))
        }
    }

    @Test func clearAfterNavigateResetsQuery() async {
        let model = OmnibarResultsModel()
        model.query = "abc"
        model.hasSearched = true
        model.hits = [
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
        #expect(model.query.isEmpty)
        #expect(!model.isPresented)
        #expect(model.hits.isEmpty)
    }

    @Test func matchContextLocalizesBodyFieldsAndHidesTitleRef() {
        #expect(OmnibarHitPresentation.matchContextText(field: "title", snippet: "") == "")
        #expect(OmnibarHitPresentation.matchContextText(field: "ref", snippet: "") == "")
        #expect(OmnibarHitPresentation.matchContextText(field: "fuzzy", snippet: "") == "")
        #expect(
            OmnibarHitPresentation.matchContextText(field: "notes", snippet: "Zemblanity")
                == L10n.Workspace.omnibarMatchNote(snippet: "Zemblanity")
        )
        #expect(
            OmnibarHitPresentation.matchContextText(field: "metadata", snippet: "author")
                == L10n.Workspace.omnibarMatchMetadata(snippet: "author")
        )
        #expect(
            OmnibarHitPresentation.matchContextText(field: "filename", snippet: "scan.png")
                == L10n.Workspace.omnibarMatchFilename(snippet: "scan.png")
        )
        #expect(
            OmnibarHitPresentation.matchContextText(field: "description", snippet: "")
                == String(localized: L10n.Workspace.omnibarMatchDescription)
        )
    }

    @Test func closePanelSetsUserDismissedWithoutClearingQuery() {
        let model = OmnibarResultsModel()
        model.query = "Ilminster"
        model.hasSearched = true
        model.isLoading = true
        model.closePanel()
        #expect(model.userDismissed)
        #expect(!model.isPresented)
        #expect(model.query == "Ilminster")
        #expect(!model.isLoading)
    }

    @Test func loadingPresentsBeforeHitsArrive() {
        let model = OmnibarResultsModel()
        model.query = "Il"
        model.isLoading = true
        #expect(model.isPresented)
        #expect(model.showsLoadingState)
        #expect(!model.showsEmptyState)
        #expect(!model.showsErrorState)
    }

    @Test func emptyStateRequiresCompletedSearchWithoutError() {
        let model = OmnibarResultsModel()
        model.query = "zz"
        model.hasSearched = true
        model.isLoading = false
        #expect(model.showsEmptyState)
        model.searchError = "boom"
        #expect(!model.showsEmptyState)
        #expect(model.showsErrorState)
    }

    @Test func searchFailureSurfacesErrorNotEmpty() async {
        enum Boom: Error { case boom }
        let store = FakeStore()
        store.searchCatalogError = Boom.boom
        let model = OmnibarResultsModel()
        model.query = "Ilminster"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model)
        #expect(model.searchError != nil)
        #expect(model.hits.isEmpty)
        #expect(model.showsErrorState)
        #expect(!model.showsEmptyState)
        #expect(model.isPresented)
    }

    @Test func moveSelectionClampsAndSelectedHit() {
        let model = OmnibarResultsModel()
        model.hits = [
            CatalogSearchHit(
                kind: "source", id: "a", ref: "", title: "A", subtitle: "",
                matchReason: "title", location: .sectionRoot(.sources)
            ),
            CatalogSearchHit(
                kind: "source", id: "b", ref: "", title: "B", subtitle: "",
                matchReason: "title", location: .sectionRoot(.sources)
            ),
        ]
        #expect(model.selectedHit()?.id == "a")
        model.moveSelection(delta: 1)
        #expect(model.selectedIndex == 1)
        #expect(model.selectedHit()?.id == "b")
        model.moveSelection(delta: 5)
        #expect(model.selectedIndex == 1)
        model.moveSelection(delta: -10)
        #expect(model.selectedIndex == 0)
    }
}
