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
        ref: String = "SRC-AAAAA",
        updatedRevision: Int64 = 0
    ) -> CatalogSource {
        CatalogSource(
            id: id,
            ref: ref,
            sourceTypeID: typeID,
            title: title,
            description: "",
            updatedRevision: updatedRevision
        )
    }

    private func makeSession(store: FakeStore) -> WorkspaceSession {
        WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    }

    private func makeModel(
        store: FakeStore = FakeStore(),
        sources: [CatalogSource] = [],
        types: [CatalogSourceType] = [],
        catalogCounts: CatalogCounts? = nil
    ) -> (SourcesModel, WorkspaceSession) {
        store.sourcesByProject[projectDir] = sources
        store.sourceTypesByProject[projectDir] = types
        let session = makeSession(store: store)
        let model = SourcesModel(
            session: session,
            userID: userID,
            store: store,
            catalogCounts: catalogCounts
        )
        return (model, session)
    }

    private func waitForQuery<Value>(_ handle: QueryHandle<Value>) async {
        var waited: UInt64 = 0
        let step: UInt64 = 10_000_000
        while waited < 2_000_000_000 {
            if handle.status == .ready || handle.status == .error { return }
            await Task.yield()
            try? await Task.sleep(nanoseconds: step)
            waited += step
        }
    }

    private func warmLists(_ model: SourcesModel, session: WorkspaceSession) async {
        model.warmListQueries()
        if let sourcesHandle: QueryHandle<[CatalogSource]> = session.queryHandle(
            SourcesModel.sourcesListKey(for: session)
        ) {
            await waitForQuery(sourcesHandle)
        }
        if let typesHandle: QueryHandle<[CatalogSourceType]> = session.queryHandle(
            SourcesModel.sourceTypesListKey(for: session)
        ) {
            await waitForQuery(typesHandle)
        }
        model.syncCatalogCounts()
    }

    @Test func sessionQueryPopulatesSourcesTypesAndPublishesCount() async {
        let store = FakeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let (model, session) = makeModel(
            store: store,
            sources: [
                source(id: "s1", title: "Album", typeID: "t1", ref: "SRC-0001"),
                source(id: "s2", title: "Deed", typeID: "t2", ref: "SRC-0002"),
            ],
            types: [photoType(), bookType()],
            catalogCounts: counts
        )
        await warmLists(model, session: session)
        #expect(model.sources.count == 2)
        #expect(model.types.count == 2)
        #expect(counts.sources == 2)
        let album = model.sources.first { $0.id == "s1" }
        #expect(album != nil)
        #expect(model.typeLabel(for: album!) == "Photograph")
    }

    @Test func listSourcesFillsThumbnailFromPinnedArtifact() async {
        let store = FakeStore()
        store.artifactsBySource["s1"] = [
            CatalogArtifact(
                id: "a1", ref: "ART-1", sourceID: "s1", fileID: "f1",
                label: "Front", description: "",
                file: CatalogFileRef(
                    id: "f1", relPath: "objects/aa/bb/file", originalFilename: "front.png",
                    mediaType: "image/png", byteSize: 12
                ),
                thumbnailRelPath: "objects/aa/bb/thumb"
            ),
        ]
        var src = source(id: "s1", title: "Album", typeID: "t1")
        src.coverMode = "artifact"
        src.primaryArtifactID = "a1"
        let (model, session) = makeModel(
            store: store,
            sources: [src],
            types: [photoType()]
        )
        await warmLists(model, session: session)
        #expect(model.sources.first?.thumbnailRelPath == "objects/aa/bb/thumb")
        #expect(model.sources.first?.coverMode == "artifact")
    }

    @Test func typeFilterNarrowsTheList() async {
        let (model, session) = makeModel(
            sources: [
                source(id: "s1", title: "A", typeID: "t1"),
                source(id: "s2", title: "B", typeID: "t2"),
                source(id: "s3", title: "C", typeID: "t1"),
            ],
            types: [photoType(), bookType()]
        )
        await warmLists(model, session: session)
        model.typeFilterID = "t1"
        #expect(model.visibleSources.map(\.id) == ["s3", "s1"])
    }

    @Test func sortOrdersByAddedUpdatedAndTitle() async {
        let (model, session) = makeModel(
            sources: [
                source(id: "s1", title: "Charlie", typeID: "t1", updatedRevision: 1),
                source(id: "s2", title: "Alpha", typeID: "t1", updatedRevision: 3),
                source(id: "s3", title: "Bravo", typeID: "t1", updatedRevision: 2),
            ],
            types: [photoType()]
        )
        await warmLists(model, session: session)

        // FakeStore + model: id DESC (s3 > s2 > s1).
        model.sort = .added
        #expect(model.visibleSources.map(\.id) == ["s3", "s2", "s1"])

        model.sort = .updated
        #expect(model.visibleSources.map(\.id) == ["s2", "s3", "s1"])

        model.sort = .az
        #expect(model.visibleSources.map(\.title) == ["Alpha", "Bravo", "Charlie"])

        model.sort = .za
        #expect(model.visibleSources.map(\.title) == ["Charlie", "Bravo", "Alpha"])
    }

    @Test func createRequiresTypeAndTitle() async {
        let (model, session) = makeModel(types: [photoType(), bookType()])
        await warmLists(model, session: session)
        model.openAdd()
        #expect(model.draft.sourceTypeID.isEmpty)
        let created = await model.create()
        #expect(created == nil)
        #expect(model.typeError != nil)
        #expect(model.titleError != nil)
        #expect(model.isAdding)
    }

    @Test func createReturnsSourceAndPublishesCount() async {
        let store = FakeStore()
        let counts = CatalogCounts(projectDir: projectDir, store: store)
        let (model, session) = makeModel(
            store: store,
            types: [photoType()],
            catalogCounts: counts
        )
        await warmLists(model, session: session)
        #expect(counts.sources == 0)

        model.openAdd()
        model.draft.sourceTypeID = "t1"
        model.draft.title = "  Family album  "
        model.draft.description = "Nan's prints"
        let created = await model.create()

        #expect(!model.isAdding)
        #expect(model.sources.count == 1)
        #expect(created?.title == "Family album")
        #expect(counts.sources == 1)
        #expect(model.toast != nil)
    }

    @Test func cancelAddDoesNotCreate() async {
        let (model, session) = makeModel(types: [photoType()])
        await warmLists(model, session: session)
        model.openAdd()
        model.draft.title = "Unused"
        model.cancelAdd()
        #expect(!model.isAdding)
        #expect(model.sources.isEmpty)
    }

    @Test func refreshTypesWarmsTypesHandle() async {
        let store = FakeStore()
        let (model, session) = makeModel(store: store, types: [])
        model.warmListQueries()
        store.sourceTypesByProject[projectDir] = [photoType(), bookType()]
        session.invalidate(CatalogQueryKey.sourceTypesList(project: session.projectKey))
        model.refreshTypes()
        if let typesHandle: QueryHandle<[CatalogSourceType]> = session.queryHandle(
            SourcesModel.sourceTypesListKey(for: session)
        ) {
            await waitForQuery(typesHandle)
        }
        #expect(model.types.map(\.id) == ["t1", "t2"])
        #expect(model.typeComboOptions.count == 2)
    }

    @Test func typesHandleLoadsWhenSourcesListFails() async {
        enum Boom: Error { case boom }
        let store = FakeStore()
        store.listSourcesError = Boom.boom
        let (model, session) = makeModel(store: store, types: [photoType()])
        await warmLists(model, session: session)
        #expect(model.types.map(\.id) == ["t1"])
        #expect(model.loadError != nil)
    }

    @Test func typeIconKeyResolvesFromSourceType() async {
        var photo = photoType()
        photo.iconKey = "type_photograph"
        let (model, session) = makeModel(
            sources: [source(id: "s1", title: "Album", typeID: "t1")],
            types: [photo]
        )
        await warmLists(model, session: session)
        #expect(model.typeIconKey(for: model.sources[0]) == "type_photograph")
    }

    @Test func cacheHitSkipsSecondListLoad() async {
        let store = FakeStore()
        let (model, session) = makeModel(
            store: store,
            sources: [source(id: "s1", title: "Album", typeID: "t1")],
            types: [photoType()]
        )
        await warmLists(model, session: session)
        store.sourcesByProject[projectDir] = [
            source(id: "s2", title: "Other", typeID: "t1", ref: "SRC-0002"),
        ]

        let _: QueryHandle<[CatalogSource]> = session.query(
            CatalogQueryKey.sourcesList(project: session.projectKey)
        )
        await Task.yield()
        #expect(model.sources.first?.title == "Album")
    }

    @Test func updatedSourcePatchesListRow() async {
        let store = FakeStore()
        let (model, session) = makeModel(
            store: store,
            sources: [source(id: "s1", title: "Deed", typeID: "t1")],
            types: [photoType()]
        )
        await warmLists(model, session: session)

        var updated = source(id: "s1", title: "Deed renamed", typeID: "t1")
        updated.coverMode = "type_icon"
        session.apply(.updatedSource(updated))

        #expect(model.sources.first?.title == "Deed renamed")
        #expect(model.sources.first?.coverMode == "type_icon")
    }

    @Test func warmListQueriesLoadsGraphProgress() async {
        let store = FakeStore()
        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 4, observationCount: 0, uncitedCount: 4
        )
        let (model, session) = makeModel(
            store: store,
            sources: [source(id: "s1", title: "Album", typeID: "t1")],
            types: [photoType()]
        )
        await warmLists(model, session: session)
        if let progressHandle: QueryHandle<[String: SourceGraphProgress]> = session.queryHandle(
            SourcesModel.sourceGraphProgressKey(for: session)
        ) {
            await waitForQuery(progressHandle)
        }
        #expect(model.graphProgress(for: "s1")?.subjectCount == 4)
        #expect(model.graphProgress(for: "s1")?.uncitedCount == 4)
        #expect(model.graphCountsLoading == false)
    }

    @Test func graphProgressCopyMatchesBoard() {
        #expect(L10n.Sources.graphCountLine(subjects: 12, observations: 48) == "12 subjects · 48 observations")
        #expect(L10n.Sources.graphUncitedCount(3) == "3 uncited")
        #expect(L10n.Sources.graphSubjectCount(0) == "0 subjects")
        #expect(
            L10n.Sources.graphZoneAccessibility(progress: nil, countsLoading: true)
                == "Open graph, counts loading"
        )
        #expect(
            L10n.Sources.graphZoneAccessibility(
                progress: .zeros(sourceId: "s1"),
                countsLoading: false
            ) == "Open graph, 0 subjects, not started"
        )
        #expect(
            L10n.Sources.graphZoneAccessibility(
                progress: SourceGraphProgress(
                    sourceId: "s1", subjectCount: 12, observationCount: 48, uncitedCount: 3
                ),
                countsLoading: false
            ) == "Open graph, 12 subjects, 48 observations, 3 uncited"
        )
    }
}
