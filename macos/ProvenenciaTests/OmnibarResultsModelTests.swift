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

    @Test func accessibilityLabelComposesVisibleRowMetadata() {
        let hit = CatalogSearchHit(
            kind: "source",
            id: "s1",
            ref: "SRC-3K9M2",
            title: "Ilminster parish register",
            subtitle: "Parish register",
            matchReason: "notes",
            matchSnippet: "Zemblanity",
            location: .sectionRoot(.sources)
        )
        let label = OmnibarHitPresentation.accessibilityLabel(for: hit)
        #expect(label.contains("Ilminster parish register"))
        #expect(label.contains(String(localized: L10n.Workspace.omnibarKindSource)))
        #expect(label.contains("Parish register"))
        #expect(label.contains("SRC-3K9M2"))
        #expect(label.contains(L10n.Workspace.omnibarMatchNote(snippet: "Zemblanity")))
    }

    @Test func accessibilityLabelOmitsEmptyOptionalParts() {
        let hit = CatalogSearchHit(
            kind: "source_field",
            id: "f1",
            ref: "",
            title: "Citation",
            subtitle: "",
            matchReason: "title",
            location: .sectionRoot(.sourceFields)
        )
        let label = OmnibarHitPresentation.accessibilityLabel(for: hit)
        #expect(label.contains("Citation"))
        #expect(label.contains(String(localized: L10n.Workspace.omnibarKindField)))
        #expect(!label.contains("Note:"))
        #expect(!label.contains("Metadata:"))
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
        model.query = "Deed"
        model.hits = [hit]
        model.hasSearched = true

        model.activate(hit, navigation: navigation)

        #expect(navigation.currentLocation.sourceId == "s1")
        #expect(navigation.selectedSection == .sources)
        #expect(model.query.isEmpty)
        #expect(model.hits.isEmpty)
        #expect(!model.hasSearched)
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
        model.query = "Author"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sourceFields),
            store: store
        )
        for _ in 0..<80 {
            if !model.isLoading, model.hasSearched || model.searchError != nil {
                break
            }
            try? await Task.sleep(nanoseconds: 25_000_000)
        }
        let hit = try? #require(model.hits.first { $0.id == "f9" })
        guard let hit else { return }

        let navigation = WorkspaceNavigation()
        model.activate(hit, navigation: navigation)

        #expect(navigation.currentLocation.fieldId == "f9")
        #expect(navigation.selectedSection == .sourceFields)
        #expect(model.query.isEmpty)
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

    @Test func searchFailureSurfacesMappedErrorNotEmpty() async {
        let store = FakeStore()
        store.searchCatalogError = CoreInvokeError.coded(
            status: 1,
            code: "catalog.closed",
            kind: .user,
            params: []
        )
        let model = OmnibarResultsModel()
        model.query = "Ilminster"
        model.scheduleSearch(
            projectDir: projectDir,
            location: .sectionRoot(.sources),
            store: store
        )
        await waitForSearch(model)
        #expect(model.searchError == L10n.Errors.message(code: "catalog.closed"))
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
