import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SourcesModelTests {
    private let projectDir = "/tmp/sources.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"

    private func photoType(id: String = "t1") -> CatalogSourceType {
        CatalogSourceType(
            id: id, key: "photograph", origin: "provenencia", label: "Photograph",
            description: "A photographic image."
        )
    }

    private func bookType(id: String = "t2") -> CatalogSourceType {
        CatalogSourceType(
            id: id, key: "book", origin: "provenencia", label: "Book",
            description: "A published monograph."
        )
    }

    private func source(
        id: String,
        title: String,
        typeID: String,
        ref: String = "SRC-AAAAA"
    ) -> CatalogSource {
        CatalogSource(id: id, ref: ref, sourceTypeID: typeID, title: title, description: "")
    }

    private func makeModel(
        store: FakeStore = FakeStore(),
        sources: [CatalogSource] = [],
        types: [CatalogSourceType] = [],
        catalogCounts: CatalogCounts? = nil
    ) -> SourcesModel {
        store.sourcesByProject[projectDir] = sources
        store.sourceTypesByProject[projectDir] = types
        return SourcesModel(
            projectDir: projectDir,
            userID: userID,
            store: store,
            catalogCounts: catalogCounts
        )
    }

    @Test func loadPopulatesSourcesTypesAndPublishesCount() async {
        let store = FakeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let model = makeModel(
            store: store,
            sources: [
                source(id: "s1", title: "Album", typeID: "t1", ref: "SRC-0001"),
                source(id: "s2", title: "Deed", typeID: "t2", ref: "SRC-0002"),
            ],
            types: [photoType(), bookType()],
            catalogCounts: counts
        )
        await model.load()
        #expect(model.sources.count == 2)
        #expect(model.types.count == 2)
        #expect(counts.sources == 2)
        #expect(model.typeLabel(for: model.sources[0]) == "Photograph")
    }

    @Test func listSourcesFillsThumbnailFromArtifact() async {
        let store = FakeStore()
        store.artifactsBySource["s1"] = [
            CatalogArtifact(
                id: "a1", ref: "ART-1", sourceID: "s1", fileID: "f1",
                label: "Front", description: "", thumbnailRelPath: "objects/aa/bb/thumb"
            ),
        ]
        let model = makeModel(
            store: store,
            sources: [source(id: "s1", title: "Album", typeID: "t1")],
            types: [photoType()]
        )
        await model.load()
        #expect(model.sources.first?.thumbnailRelPath == "objects/aa/bb/thumb")
    }

    @Test func searchFiltersByTitleRefAndTypeLabel() async {
        let model = makeModel(
            sources: [
                source(id: "s1", title: "Alderwick photograph", typeID: "t1", ref: "SRC-0412"),
                source(id: "s2", title: "Parish register", typeID: "t2", ref: "SRC-0287"),
            ],
            types: [photoType(), bookType(id: "t2")]
        )
        await model.load()

        model.query = "alderwick"
        #expect(model.visibleSources.map(\.id) == ["s1"])

        model.query = "SRC-0287"
        #expect(model.visibleSources.map(\.id) == ["s2"])

        model.query = "photograph"
        #expect(model.visibleSources.map(\.id) == ["s1"])

        model.query = "nomatch"
        #expect(model.visibleSources.isEmpty)
    }

    @Test func typeFilterNarrowsTheList() async {
        let model = makeModel(
            sources: [
                source(id: "s1", title: "A", typeID: "t1"),
                source(id: "s2", title: "B", typeID: "t2"),
                source(id: "s3", title: "C", typeID: "t1"),
            ],
            types: [photoType(), bookType()]
        )
        await model.load()
        model.typeFilterID = "t1"
        #expect(model.visibleSources.map(\.id) == ["s1", "s3"])
    }

    @Test func sortOrdersByTitleAndCatalogOrder() async {
        let model = makeModel(
            sources: [
                source(id: "s1", title: "Charlie", typeID: "t1"),
                source(id: "s2", title: "Alpha", typeID: "t1"),
                source(id: "s3", title: "Bravo", typeID: "t1"),
            ],
            types: [photoType()]
        )
        await model.load()
        #expect(model.visibleSources.map(\.id) == ["s1", "s2", "s3"])

        model.sort = .az
        #expect(model.visibleSources.map(\.title) == ["Alpha", "Bravo", "Charlie"])

        model.sort = .za
        #expect(model.visibleSources.map(\.title) == ["Charlie", "Bravo", "Alpha"])

        model.sort = .updated
        #expect(model.visibleSources.map(\.id) == ["s3", "s2", "s1"])
    }

    @Test func createRequiresTypeAndTitle() async {
        let model = makeModel(types: [photoType(), bookType()])
        await model.load()
        model.openAdd()
        #expect(model.draft.sourceTypeID.isEmpty)
        await model.create()
        #expect(model.typeError != nil)
        #expect(model.titleError != nil)
        #expect(model.openedSourceID == nil)
        #expect(model.isAdding)
    }

    @Test func createNavigatesPublishesAndDismissesDialog() async {
        let store = FakeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let model = makeModel(
            store: store,
            types: [photoType()],
            catalogCounts: counts
        )
        await model.load()
        #expect(counts.sources == 0)

        model.openAdd()
        model.draft.sourceTypeID = "t1"
        model.draft.title = "  Family album  "
        model.draft.description = "Nan's prints"
        await model.create()

        #expect(!model.isAdding)
        #expect(model.sources.count == 1)
        #expect(model.sources.first?.title == "Family album")
        #expect(model.openedSourceID == model.sources.first?.id)
        #expect(counts.sources == 1)
        #expect(model.toast != nil)
    }

    @Test func cancelAddDoesNotNavigate() async {
        let model = makeModel(types: [photoType()])
        await model.load()
        model.openAdd()
        model.draft.title = "Unused"
        model.cancelAdd()
        #expect(!model.isAdding)
        #expect(model.openedSourceID == nil)
        #expect(model.sources.isEmpty)
    }

    @Test func refreshTypesRepopulatesAnEmptyPool() async {
        let store = FakeStore()
        let model = makeModel(store: store, types: [])
        await model.load()
        #expect(model.types.isEmpty)

        store.sourceTypesByProject[projectDir] = [photoType(), bookType()]
        await model.refreshTypes()
        #expect(model.types.map(\.id) == ["t1", "t2"])
        #expect(model.typeComboOptions.count == 2)
    }

    @Test func loadKeepsTypesWhenSourcesListFails() async {
        enum Boom: Error { case boom }
        let store = FakeStore()
        store.listSourcesError = Boom.boom
        store.sourceTypesByProject[projectDir] = [photoType()]
        let model = SourcesModel(projectDir: projectDir, userID: userID, store: store)
        await model.load()
        #expect(model.types.map(\.id) == ["t1"])
        #expect(model.loadError != nil)
    }

    @Test func openAndCloseSourcePage() async {
        let model = makeModel(
            sources: [source(id: "s1", title: "Album", typeID: "t1")],
            types: [photoType()]
        )
        await model.load()
        model.openSource(id: "s1")
        #expect(model.openedSourceID == "s1")
        #expect(model.openedSource?.title == "Album")
        model.closeSource()
        #expect(model.openedSourceID == nil)
    }

    @Test func applyUpdatedSourceMergesCoverWhenIdentityOmitsThumbs() async {
        var row = source(id: "s1", title: "Deed", typeID: "t1")
        row.thumbnailMediaType = "application/pdf"
        row.thumbnailOriginalFilename = "deed.pdf"
        let model = makeModel(sources: [row], types: [photoType()])
        await model.load()
        model.applyUpdatedSource(
            CatalogSource(
                id: "s1",
                ref: "SRC-AAAAA",
                sourceTypeID: "t1",
                title: "Deed renamed",
                description: ""
            )
        )
        #expect(model.sources.first?.title == "Deed renamed")
        #expect(model.sources.first?.thumbnailMediaType == "application/pdf")
        #expect(model.sources.first?.thumbnailOriginalFilename == "deed.pdf")
    }

    @Test func applyUpdatedSourceReplacesCoverWhenIncomingHasGlyph() async {
        var row = source(id: "s1", title: "Deed", typeID: "t1")
        row.thumbnailMediaType = "application/pdf"
        let model = makeModel(sources: [row], types: [photoType()])
        await model.load()
        var updated = source(id: "s1", title: "Deed", typeID: "t1")
        updated.thumbnailRelPath = "objects/aa/bb/thumb"
        model.applyUpdatedSource(updated)
        #expect(model.sources.first?.thumbnailRelPath == "objects/aa/bb/thumb")
        #expect(model.sources.first?.thumbnailMediaType.isEmpty == true)
    }
}
