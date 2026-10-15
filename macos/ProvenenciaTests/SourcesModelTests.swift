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
            updated: CatalogActivity(revision: updatedRevision)
        )
    }

    private func makeSession(store: FakeStore) -> WorkspaceSession {
        WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    }

    private func makeModel(
        store: FakeStore = FakeStore(),
        sources: [CatalogSource] = [],
        types: [CatalogSourceType] = [],
        catalogCounts: CatalogCounts? = nil,
        settingsStore: SourcesListSettingsStore? = nil
    ) -> (SourcesModel, WorkspaceSession) {
        store.sourcesByProject[projectDir] = sources
        store.sourceTypesByProject[projectDir] = types
        let session = makeSession(store: store)
        let model = SourcesModel(
            session: session,
            userID: userID,
            store: store,
            catalogCounts: catalogCounts,
            settingsStore: settingsStore
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

    private func warmProgress(_ model: SourcesModel, session: WorkspaceSession) async {
        if let handle: QueryHandle<[String: SourceGraphProgress]> = session.queryHandle(
            SourcesModel.sourceGraphProgressKey(for: session)
        ) {
            await waitForQuery(handle)
        }
    }

    private func progress(
        _ id: String, primary: Int, filed: Int
    ) -> SourceGraphProgress {
        SourceGraphProgress(
            sourceId: id, subjectCount: primary, observationCount: 0,
            primaryCount: primary, filedCount: filed
        )
    }

    private func tempSettingsFile() throws -> URL {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent("sources-settings-\(UUID().uuidString)", isDirectory: true)
        try FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        return dir.appendingPathComponent("settings.json", isDirectory: false)
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
        model.settings.typeFilterID = "t1"
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
        model.settings.sort = .added
        #expect(model.visibleSources.map(\.id) == ["s3", "s2", "s1"])

        model.settings.sort = .updated
        #expect(model.visibleSources.map(\.id) == ["s2", "s3", "s1"])

        model.settings.sort = .az
        #expect(model.visibleSources.map(\.title) == ["Alpha", "Bravo", "Charlie"])

        model.settings.sort = .za
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

    /// A write's activity moves the row in the Updated sort without a reload.
    @Test func writeActivityMovesTheRowInTheUpdatedSort() async {
        let (model, session) = makeModel(
            sources: [
                source(id: "s1", title: "Deed", typeID: "t1", updatedRevision: 5),
                source(id: "s2", title: "Letter", typeID: "t1", updatedRevision: 9),
            ],
            types: [photoType()]
        )
        await warmLists(model, session: session)
        model.settings.sort = .updated
        #expect(model.visibleSources.map(\.id) == ["s2", "s1"])

        session.apply(CatalogWriteEffects(
            sourceScopes: ["s1"],
            activity: CatalogActivity(revision: 12, at: "2026-10-15T16:00:00Z")
        ))

        #expect(model.visibleSources.map(\.id) == ["s1", "s2"])
    }

    @Test func warmListQueriesLoadsGraphProgress() async {
        let store = FakeStore()
        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 4, observationCount: 0
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
        #expect(model.graphCountsLoading == false)
    }

    @Test func graphProgressCopyMatchesBoard() {
        #expect(L10n.Sources.graphCountLine(subjects: 12, observations: 48) == "12 subjects · 48 observations")
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
                    sourceId: "s1", subjectCount: 12, observationCount: 48
                ),
                countsLoading: false
            ) == "Open graph, 12 subjects, 48 observations"
        )
    }

    @Test func evidenceStatusFromCounts() {
        #expect(SourcesModel.EvidenceStatus(nil) == .empty)
        #expect(SourcesModel.EvidenceStatus(progress("s", primary: 0, filed: 0)) == .empty)
        #expect(SourcesModel.EvidenceStatus(progress("s", primary: 2, filed: 1)) == .inProgress)
        #expect(SourcesModel.EvidenceStatus(progress("s", primary: 2, filed: 2)) == .filed)
    }

    @Test func settingsRestoreOnNewModel() async throws {
        let url = try tempSettingsFile()
        let uuid = "00000000-0000-7000-8000-0000000000aa"
        let storeA = try #require(SourcesListSettingsStore(projectUuid: uuid, userID: userID, fileURL: url))
        let (a, _) = makeModel(types: [photoType()], settingsStore: storeA)
        a.settings.sort = .az
        a.settings.typeFilterID = "t1"
        a.settings.status = .filed
        a.settings.query = "smith"

        let storeB = try #require(SourcesListSettingsStore(projectUuid: uuid, userID: userID, fileURL: url))
        let (b, _) = makeModel(types: [photoType()], settingsStore: storeB)
        #expect(b.settings == a.settings)
        #expect(b.settings.sort == .az)
        #expect(b.settings.status == .filed)
    }

    @Test func searchMatchesTitleAndRefIgnoringCaseAndAccents() async {
        let (model, session) = makeModel(
            sources: [
                source(id: "s1", title: "Smith family Bible", typeID: "t1", ref: "SRC-1AAAA"),
                source(id: "s2", title: "1851 census", typeID: "t1", ref: "SRC-4F2NP"),
                source(id: "s3", title: "José Ruiz letter", typeID: "t1", ref: "SRC-3CCCC"),
            ],
            types: [photoType()]
        )
        await warmLists(model, session: session)

        model.settings.query = "smith"
        #expect(model.visibleSources.map(\.id) == ["s1"])
        model.settings.query = "jose"
        #expect(model.visibleSources.map(\.id) == ["s3"])
        model.settings.query = "src-4f"
        #expect(model.visibleSources.map(\.id) == ["s2"])
        model.settings.query = "  "
        #expect(model.visibleSources.count == 3)
        #expect(!model.isFiltering)
    }

    @Test func filtersCombineAndClear() async {
        let store = FakeStore()
        var sources: [CatalogSource] = []
        for index in 1 ... 8 {
            let typeID = index <= 4 ? "t1" : "t2"
            let title = index == 1 || index == 2 ? "Smith record \(index)" : "Other record \(index)"
            sources.append(source(id: "s\(index)", title: title, typeID: typeID, ref: "SRC-0000\(index)"))
        }
        for id in ["s1", "s2", "s5"] {
            store.graphProgressBySource[id] = progress(id, primary: 2, filed: 1)
        }
        store.graphProgressBySource["s3"] = progress("s3", primary: 1, filed: 1)
        let (model, session) = makeModel(
            store: store, sources: sources, types: [photoType(), bookType()]
        )
        await warmLists(model, session: session)
        await warmProgress(model, session: session)

        model.settings.sort = .az
        model.settings.typeFilterID = "t1"
        model.settings.status = .inProgress
        model.settings.query = "smith"
        #expect(model.visibleSources.map(\.id) == ["s1", "s2"])
        #expect(model.isFiltering)
        #expect(model.countLine == "2 of 8 sources")

        model.settings.query = ""
        model.settings.typeFilterID = ""
        #expect(model.visibleSources.map(\.id).sorted() == ["s1", "s2", "s5"])
        #expect(model.countLine == "3 of 8 sources")

        model.clearFilters()
        #expect(model.visibleSources.count == 8)
        #expect(model.settings.sort == .az)
        #expect(model.settings.typeFilterID.isEmpty)
        #expect(model.settings.status == nil)
        #expect(model.settings.query.isEmpty)
        #expect(model.countLine == "8 sources")
    }

    @Test func restoredTypeMissingFromTypesShowsEveryRow() async {
        let (model, session) = makeModel(
            sources: [
                source(id: "s1", title: "A", typeID: "t1"),
                source(id: "s2", title: "B", typeID: "t1"),
            ],
            types: [photoType()]
        )
        await warmLists(model, session: session)
        model.settings.typeFilterID = "deleted-type"
        #expect(model.visibleSources.count == 2)
        #expect(!model.isFiltering)
        #expect(model.effectiveTypeFilterID.isEmpty)
        #expect(model.settings.typeFilterID == "deleted-type")
    }

    @Test func statusFilterWaitsForCounts() async {
        let store = FakeStore()
        store.graphProgressBySource["s1"] = progress("s1", primary: 1, filed: 0)
        let (model, session) = makeModel(
            store: store,
            sources: [
                source(id: "s1", title: "A", typeID: "t1"),
                source(id: "s2", title: "B", typeID: "t1"),
            ],
            types: [photoType()]
        )
        model.settings.status = .empty
        #expect(model.isAwaitingFilterData)
        await warmLists(model, session: session)
        await warmProgress(model, session: session)
        #expect(!model.isAwaitingFilterData)
        #expect(model.visibleSources.map(\.id) == ["s2"])

        let restored = SourcesModel(
            session: makeSession(store: FakeStore()),
            userID: userID,
            store: FakeStore()
        )
        restored.settings.typeFilterID = "t1"
        #expect(restored.isAwaitingFilterData)
        #expect(restored.countLine == L10n.Sources.countLine(total: 0))
    }

    @Test func statusCountsFailureSurfacesAnErrorAndUnfilteredCount() async {
        enum Boom: Error { case boom }
        let store = FakeStore()
        store.listSourceGraphProgressError = Boom.boom
        let (model, session) = makeModel(
            store: store,
            sources: [
                source(id: "s1", title: "A", typeID: "t1"),
                source(id: "s2", title: "B", typeID: "t1"),
            ],
            types: [photoType()]
        )
        await warmLists(model, session: session)
        await warmProgress(model, session: session)
        #expect(model.statusCountsError == nil)

        model.settings.status = .filed
        #expect(model.statusCountsError != nil)
        #expect(model.countLine == L10n.Sources.countLine(total: 2))
    }
}
