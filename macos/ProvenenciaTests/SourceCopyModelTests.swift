import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SourceCopyModelTests {
    private let projectDir = "/tmp/source-copy.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"

    private func field(_ id: String, _ label: String) -> CatalogMetadataField {
        CatalogMetadataField(
            id: id, key: id, origin: "provenencia", label: label, dataType: "text", description: ""
        )
    }

    private func artifact(_ id: String, label: String) -> CatalogArtifact {
        CatalogArtifact(
            id: id, ref: "ART-\(id)", sourceID: "s1", fileID: "", label: label, description: ""
        )
    }

    private func original() -> CatalogSource {
        CatalogSource(
            id: "s1", ref: "SRC-AAAAA", sourceTypeID: "t1", title: "Obit",
            description: "From the Herald"
        )
    }

    /// Fixture: `s1` with a description, two metadata values, a credibility
    /// grade and two Artifacts (one unlabeled).
    private func makeStore(
        original: CatalogSource? = nil,
        metadata: Bool = true,
        credibility: Bool = true,
        artifacts: Bool = true
    ) -> FakeStore {
        let store = FakeStore()
        let src = original ?? self.original()
        store.sourcesByProject[projectDir] = [src]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1", key: "obituary", origin: "provenencia", label: "Obituary", description: ""
            ),
        ]
        if metadata {
            store.metadataBySource["s1"] = [
                CatalogMetadataEntry(
                    field: field("f1", "Publication"), valueText: "Herald",
                    hasValue: true, suggested: false, sortOrder: 0
                ),
                CatalogMetadataEntry(
                    field: field("f2", "Page"), valueText: "5",
                    hasValue: true, suggested: false, sortOrder: 1
                ),
            ]
        }
        if credibility {
            store.credibilityBySource["s1"] = CatalogCredibilityAssessment(
                id: "c1", sourceID: "s1", gradeID: "g1", gradeKey: "high_trust",
                gradeLabel: "High trust", argument: "Signed"
            )
        }
        if artifacts {
            store.artifactsBySource["s1"] = [
                artifact("a1", label: "Front page"),
                artifact("a2", label: ""),
            ]
        }
        return store
    }

    private func makeModel(
        store: FakeStore,
        original: CatalogSource? = nil
    ) -> (SourceCopyModel, WorkspaceSession) {
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourceCopyModel(
            original: original ?? self.original(),
            session: session,
            userID: userID,
            store: store
        )
        return (model, session)
    }

    private func workspaceKey(_ session: WorkspaceSession, _ id: String = "s1") -> CatalogQueryKey {
        .sourceWorkspace(project: session.projectKey, sourceId: id)
    }

    private func loaded(
        _ store: FakeStore,
        original: CatalogSource? = nil
    ) async -> (SourceCopyModel, WorkspaceSession) {
        let (model, session) = makeModel(store: store, original: original)
        let id = original?.id ?? "s1"
        let handle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey(session, id))
        await waitForQuery(handle)
        return (model, session)
    }

    @Test func prefillsTitleAndType() {
        let (model, _) = makeModel(store: makeStore())
        #expect(model.choices.draft.title == "Copy of Obit")
        #expect(model.choices.draft.sourceTypeID == "t1")
        #expect(model.loadState == .loading)
        #expect(!model.canCopy)
        #expect(!model.isDirty)
    }

    @Test func allPartsOnWhenLoaded() async {
        let (model, _) = await loaded(makeStore())
        #expect(model.descriptionAvailable)
        #expect(model.metadataAvailable)
        #expect(model.credibilityAvailable)
        #expect(model.choices.includeDescription)
        #expect(model.choices.includeMetadata)
        #expect(model.choices.includeCredibility)
        #expect(model.artifactIDsToCopy == ["a1", "a2"])
        #expect(model.loadState == .ready)
    }

    @Test func copyAllCreatesCopyAndReturnsIt() async {
        let store = makeStore()
        let (model, _) = await loaded(store)
        let copy = await model.copy()

        #expect(copy?.title == "Copy of Obit")
        #expect(copy?.sourceTypeID == "t1")
        #expect(store.sourcesByProject[projectDir]?.count == 2)
        let id = copy?.id ?? ""
        #expect(copy?.description == "From the Herald")
        #expect(store.metadataBySource[id]?.count == 2)
        #expect(store.credibilityBySource[id] != nil)
        #expect(store.artifactsBySource[id]?.map(\.label) == ["Front page", ""])
        #expect(store.metadataBySource["s1"]?.count == 2)
        #expect(store.artifactsBySource["s1"]?.count == 2)
        #expect(model.copyError == nil)
    }

    @Test func uncheckedPartsAreNotCopied() async {
        let store = makeStore()
        let (model, _) = await loaded(store)
        model.choices.includeCredibility = false
        model.choices.excludedArtifactIDs = ["a2"]
        let copy = await model.copy()

        let id = copy?.id ?? ""
        #expect(store.credibilityBySource[id] == nil)
        #expect(store.artifactsBySource[id]?.map(\.label) == ["Front page"])
    }

    @Test func copyFailureKeepsChoices() async {
        let missing = CatalogSource(
            id: "s9", ref: "SRC-ZZZZZ", sourceTypeID: "t1", title: "Gone", description: ""
        )
        let store = makeStore()
        let (model, _) = await loaded(store, original: missing)
        model.choices.includeMetadata = false
        let before = model.choices
        let copy = await model.copy()

        #expect(copy == nil)
        #expect(model.copyError != nil)
        #expect(model.choices == before)
        #expect(!model.isSaving)
    }

    @Test func absentPartsAreUnavailableAndNotSent() async {
        var bare = original()
        bare.description = ""
        let store = makeStore(original: bare, metadata: false, credibility: false, artifacts: false)
        let (model, _) = await loaded(store, original: bare)

        #expect(!model.descriptionAvailable)
        #expect(!model.credibilityAvailable)
        #expect(model.artifactIDsToCopy.isEmpty)
        let copy = await model.copy()
        let id = copy?.id ?? ""
        #expect(copy?.description == "")
        #expect(store.credibilityBySource[id] == nil)
        #expect((store.artifactsBySource[id] ?? []).isEmpty)
    }

    @Test func copyRefreshesSourcesList() async {
        let store = makeStore()
        let (model, session) = await loaded(store)
        let list: QueryHandle<[CatalogSource]> = session.query(.sourcesList(project: session.projectKey))
        await waitForQuery(list)
        #expect(list.value?.count == 1)

        let copy = await model.copy()
        // The list is not on screen after a copy from the Source page: the
        // mutation marks it stale and the next open of the list reloads it.
        let reopened: QueryHandle<[CatalogSource]> = session.query(.sourcesList(project: session.projectKey))
        await waitForQuery(reopened)
        #expect(reopened.value?.contains { $0.id == copy?.id } == true)
    }

    @Test func validationBlocksCopy() async {
        let store = makeStore()
        let (model, _) = await loaded(store)
        model.choices.draft.title = " "
        #expect(!model.canCopy)
        let copy = await model.copy()
        #expect(copy == nil)
        #expect(model.titleError != nil)
        #expect(store.sourcesByProject[projectDir]?.count == 1)
    }

    @Test func isDirtyTracksChanges() async {
        let (model, _) = await loaded(makeStore())
        #expect(!model.isDirty)
        model.choices.includeMetadata = false
        #expect(model.isDirty)
        model.choices.includeMetadata = true
        #expect(!model.isDirty)
    }

    @Test func artifactDeletedElsewhereIsDropped() async {
        let store = makeStore()
        let (model, session) = await loaded(store)
        #expect(model.artifactIDsToCopy == ["a1", "a2"])

        store.artifactsBySource["s1"] = [artifact("a1", label: "Front page")]
        session.invalidate(workspaceKey(session))
        let handle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey(session))
        await waitForQuery(handle)
        #expect(model.artifactIDsToCopy == ["a1"])
    }

    @Test func readAgainReloadsWorkspace() async {
        let store = makeStore()
        let (model, session) = await loaded(store)
        let calls = store.getSourceWorkspaceCalls
        model.reloadWorkspace()
        let handle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey(session))
        await waitForQuery(handle)
        #expect(store.getSourceWorkspaceCalls == calls + 1)
        #expect(model.loadState == .ready)
    }

    @Test func loadStatePrecedence() async {
        let (model, _) = makeModel(store: makeStore())
        #expect(model.loadState == .loading)
        let (loadedModel, _) = await loaded(makeStore())
        #expect(loadedModel.loadState == .ready)
    }
}
