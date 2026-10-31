import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct SourceGraphFormModelTests {
    private let projectDir = "/tmp/source-graph-form.provenencia"
    private let userID = "00000000-0000-7000-8000-000000000001"
    private let sourceID = "src-1"

    private struct Boom: Error {}

    private func artifact(_ id: String, label: String) -> CatalogArtifact {
        CatalogArtifact(
            id: id, ref: "ART-\(id)", sourceID: "src-1", fileID: "", label: label, description: ""
        )
    }

    /// `a1` – `a<count>` on Source `src-1`, in that page order.
    private func artifacts(_ count: Int) -> [CatalogArtifact] {
        (1...max(count, 1)).prefix(count).map { artifact("a\($0)", label: "Page \($0)") }
    }

    private func makeStore(artifactCount: Int, graphs: [CatalogEvidenceGraph]) -> FakeStore {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: sourceID, ref: "SRC-AAAAA", sourceTypeID: "t1",
                title: "1851 census", description: ""
            ),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(
                id: "t1", key: "census", origin: "provenencia", label: "Census", description: ""
            ),
        ]
        store.artifactsBySource[sourceID] = artifacts(artifactCount)
        store.evidenceGraphsByProject[projectDir] = graphs
        store.nextWriteEffects = CatalogWriteEffects(lists: [.evidenceGraphs])
        return store
    }

    /// The Source page with its location applied, so the graph list is a
    /// visible key and a graph write's effects refresh it.
    private func makeModel(
        artifactCount: Int = 6,
        graphs: [CatalogEvidenceGraph] = []
    ) async -> (model: SourcePageModel, store: FakeStore, session: WorkspaceSession) {
        let store = makeStore(artifactCount: artifactCount, graphs: graphs)
        let session = WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
        let model = SourcePageModel(
            sourceID: sourceID, session: session, userID: userID,
            sessionDisplayName: "Jake Robins", store: store
        )
        session.apply(location: WorkspaceLocation.sourcePage(sourceId: sourceID))
        await model.warmFromSession()
        await waitForQuery(session.evidenceGraphsHandle!)
        return (model, store, session)
    }

    /// `g-smith` linked to a3 and a4, with a3 in use by two Citations; the edit form open and loaded.
    private func openSmithEdit(
        store: FakeStore,
        model: SourcePageModel,
        blocks: [CatalogEvidenceGraphUnlinkBlock]? = nil
    ) async -> SourceGraphFormModel? {
        store.unlinkImpactBlocks = blocks ?? [unlinkBlock("a3", graphID: "g-smith", citations: 2)]
        model.graphs.openEdit(id: "g-smith")
        guard let form = model.graphs.form else { return nil }
        _ = await waitUntil { form.loadState != .loading }
        return form
    }

    private func smith() -> CatalogEvidenceGraph {
        evidenceGraph("g-smith", name: "Smith household", description: "Household 9", artifactIDs: ["a3", "a4"])
    }

    @Test func createPrefillsEveryArtifactChecked() async throws {
        let (model, _, _) = await makeModel()
        model.graphs.openAdd()
        let form = try #require(model.graphs.form)
        #expect(form.graph == nil)
        #expect(form.draft.name == "")
        #expect(form.checkedArtifactIDs == ["a1", "a2", "a3", "a4", "a5", "a6"])
        #expect(form.loadState == .ready)
        #expect(!form.isDirty)
    }

    @Test func createSavesOneGraphAndToasts() async throws {
        let (model, store, _) = await makeModel()
        model.graphs.openAdd()
        let form = try #require(model.graphs.form)
        form.draft.name = "Jones household"
        form.draft.description = " Household 14, page 3 "
        form.draft.excludedArtifactIDs = ["a1", "a2", "a4", "a5", "a6"]
        let saved = await form.save()
        #expect(saved)
        #expect(store.createEvidenceGraphCalls == 1)
        #expect(await waitUntil { model.graphs.rows?.contains { $0.name == "Jones household" } == true })
        let created = try #require(model.graphs.rows?.first { $0.name == "Jones household" })
        #expect(created.sourceID == "src-1")
        #expect(created.description == "Household 14, page 3")
        #expect(created.artifactIDs == ["a3"])
        #expect(model.toast?.title == "Evidence graph added")
        #expect(model.toast?.body == "\(created.ref) is ready to open.")
    }

    enum Blocked: CaseIterable, Sendable {
        case blankName
        case noArtifact
    }

    @Test(arguments: Blocked.allCases)
    func blankNameAndNoArtifactBlockSave(_ blocked: Blocked) async throws {
        let (model, store, _) = await makeModel()
        model.graphs.openAdd()
        let form = try #require(model.graphs.form)
        switch blocked {
        case .blankName:
            form.draft.name = "  "
            let saved = await form.save()
            #expect(!saved)
            #expect(form.nameError == "Give the evidence graph a name.")
            #expect(store.createEvidenceGraphCalls == 0)
        case .noArtifact:
            form.draft.name = "Jones household"
            form.draft.excludedArtifactIDs = Set(["a1", "a2", "a3", "a4", "a5", "a6"])
            #expect(form.needsArtifact)
            #expect(!form.canSave)
            let saved = await form.save()
            #expect(!saved)
            #expect(store.createEvidenceGraphCalls == 0)
        }
    }

    @Test(arguments: [false, true])
    func saveFailureKeepsChoices(edit: Bool) async throws {
        let (model, store, _) = await makeModel(graphs: [smith()])
        let form: SourceGraphFormModel
        if edit {
            store.updateEvidenceGraphError = Boom()
            form = try #require(await openSmithEdit(store: store, model: model, blocks: []))
        } else {
            store.createEvidenceGraphError = Boom()
            model.graphs.openAdd()
            form = try #require(model.graphs.form)
        }
        form.draft.name = "Renamed"
        form.draft.excludedArtifactIDs.insert("a1")
        let before = form.draft
        let saved = await form.save()
        #expect(!saved)
        #expect(form.saveError != nil)
        #expect(form.draft == before)
    }

    @Test func editPrefillsAndLocksInUseArtifacts() async throws {
        let (model, store, _) = await makeModel(graphs: [smith()])
        let form = try #require(await openSmithEdit(store: store, model: model))
        #expect(store.getEvidenceGraphUnlinkImpactCalls == 1)
        #expect(form.graph?.ref == "GPH-G-SMITH")
        #expect(form.draft.name == "Smith household")
        #expect(form.draft.description == "Household 9")
        #expect(form.draft.excludedArtifactIDs == ["a1", "a2", "a5", "a6"])
        #expect(form.locked == ["a3": "2 citations in this graph use this artifact"])
        #expect(form.blocks.count == 1)
        #expect(form.keptLinked.map(\.artifact.id) == ["a3"])
        #expect(form.loadState == .ready)
    }

    @Test func editLocksEveryBlock() async throws {
        let (model, store, _) = await makeModel(graphs: [smith()])
        let form = try #require(await openSmithEdit(store: store, model: model, blocks: [
            unlinkBlock("a3", graphID: "g-smith", citations: 1, reason: .artifactInUse),
            unlinkBlock("a4", graphID: "g-smith", citations: 0, reason: .lastArtifact),
        ]))
        #expect(form.locked == [
            "a3": "1 citation in this graph uses this artifact",
            "a4": "An evidence graph needs at least one artifact.",
        ])
        #expect(form.keptLinked.map(\.artifact.id) == ["a3"])
        #expect(form.checkedArtifactIDs == ["a3", "a4"])

        let (other, otherStore, _) = await makeModel(graphs: [smith()])
        let lastOnly = try #require(await openSmithEdit(store: otherStore, model: other, blocks: [
            unlinkBlock("a4", graphID: "g-smith", citations: 0, reason: .lastArtifact),
        ]))
        #expect(lastOnly.locked.keys.sorted() == ["a4"])
        #expect(lastOnly.keptLinked.isEmpty)
    }

    @Test func createPrefillsTheArtifactsInNoGraph() async throws {
        let (model, _, _) = await makeModel(graphs: [
            evidenceGraph("g-default", name: "Default", artifactIDs: ["a1", "a2", "a3", "a4"]),
        ])
        model.graphs.openAdd()
        let form = try #require(model.graphs.form)
        #expect(form.checkedArtifactIDs == ["a5", "a6"])
        #expect(!form.isDirty)

        let (full, _, _) = await makeModel(graphs: [
            evidenceGraph("g-default", name: "Default", artifactIDs: ["a1", "a2", "a3", "a4", "a5", "a6"]),
        ])
        full.graphs.openAdd()
        let fullForm = try #require(full.graphs.form)
        #expect(fullForm.checkedArtifactIDs == ["a1", "a2", "a3", "a4", "a5", "a6"])
        #expect(!fullForm.isDirty)
    }

    @Test func editSavesNameAndFullLinkSet() async throws {
        let (model, store, _) = await makeModel(graphs: [smith()])
        let form = try #require(await openSmithEdit(store: store, model: model))
        form.draft.name = "Smith family"
        form.draft.excludedArtifactIDs.insert("a4")
        form.draft.excludedArtifactIDs.remove("a5")
        let saved = await form.save()
        #expect(saved)
        #expect(store.updateEvidenceGraphCalls == 1)
        #expect(await waitUntil { model.graphs.rows?.contains { $0.name == "Smith family" } == true })
        let updated = try #require(model.graphs.rows?.first { $0.id == "g-smith" })
        #expect(updated.artifactIDs == ["a3", "a5"])
        #expect(model.toast == nil)
    }

    @Test func editRefusedInUseRereadsKeptLinks() async throws {
        let linked = evidenceGraph("g-pair", name: "Pair", artifactIDs: ["a2", "a3"])
        let (model, store, _) = await makeModel(graphs: [linked])
        store.unlinkImpactBlocks = []
        model.graphs.openEdit(id: "g-pair")
        let form = try #require(model.graphs.form)
        _ = await waitUntil { form.loadState != .loading }
        #expect(form.locked.isEmpty)
        form.draft.excludedArtifactIDs.insert("a2")
        store.updateEvidenceGraphError = CoreInvokeError.coded(
            status: 1, code: "evidencegraphs.artifact_in_use", kind: .conflict, params: []
        )
        store.unlinkImpactBlocks = [unlinkBlock("a2", graphID: "g-pair", citations: 1)]
        let before = form.draft
        let saved = await form.save()
        #expect(!saved)
        #expect(form.saveError != nil)
        #expect(form.draft == before)
        #expect(form.locked.keys.contains("a2"))
        #expect(form.checkedArtifactIDs.contains("a2"))
    }

    @Test func editRefusedArtifactInUseRereadsKeptLinks() async throws {
        let linked = evidenceGraph("g-pair", name: "Pair", artifactIDs: ["a2", "a3"])
        let (model, store, _) = await makeModel(graphs: [linked])
        store.unlinkImpactBlocks = []
        model.graphs.openEdit(id: "g-pair")
        let form = try #require(model.graphs.form)
        _ = await waitUntil { form.loadState != .loading }
        form.draft.excludedArtifactIDs.insert("a2")
        store.updateEvidenceGraphError = CoreInvokeError.coded(
            status: 1, code: "evidencegraphs.artifact_in_use", kind: .conflict, params: []
        )
        store.unlinkImpactBlocks = [unlinkBlock("a2", graphID: "g-pair", citations: 1, reason: .artifactInUse)]
        let calls = store.getEvidenceGraphUnlinkImpactCalls
        let saved = await form.save()
        #expect(!saved)
        #expect(store.getEvidenceGraphUnlinkImpactCalls == calls + 1)
        #expect(form.saveError != nil)
        #expect(form.locked["a2"] == "1 citation in this graph uses this artifact")
        #expect(form.checkedArtifactIDs.contains("a2"))

        // A refusal with no code re-reads too and keeps the refusal in view.
        store.updateEvidenceGraphError = CoreInvokeError.failed(status: 3)
        let again = store.getEvidenceGraphUnlinkImpactCalls
        let savedAgain = await form.save()
        #expect(!savedAgain)
        #expect(store.getEvidenceGraphUnlinkImpactCalls == again + 1)
        #expect(form.saveError != nil)
    }

    @Test func editPreflightFailureThenReadAgain() async throws {
        let (model, store, _) = await makeModel(graphs: [smith()])
        store.getEvidenceGraphUnlinkImpactError = Boom()
        store.unlinkImpactBlocks = [unlinkBlock("a3", graphID: "g-smith", citations: 2)]
        model.graphs.openEdit(id: "g-smith")
        let form = try #require(model.graphs.form)
        _ = await waitUntil { form.loadState != .loading }
        guard case .failed = form.loadState else {
            Issue.record("expected a failed load, got \(form.loadState)")
            return
        }
        #expect(!form.canSave)
        store.getEvidenceGraphUnlinkImpactError = nil
        await form.loadKeptLinks()
        #expect(form.loadState == .ready)
        #expect(store.getEvidenceGraphUnlinkImpactCalls == 2)
    }

    @Test func editDefaultWithNoArtifactsKeepsLinks() async throws {
        let (model, store, _) = await makeModel(
            artifactCount: 0,
            graphs: [evidenceGraph("g-default", name: "Default")]
        )
        model.graphs.openEdit(id: "g-default")
        let form = try #require(model.graphs.form)
        #expect(store.getEvidenceGraphUnlinkImpactCalls == 0)
        #expect(form.canSave)
        form.draft.name = "Whole census"
        let saved = await form.save()
        #expect(saved)
        #expect(await waitUntil { model.graphs.rows?.first?.name == "Whole census" })
        #expect(model.graphs.rows?.first?.artifactIDs == [])
    }
}
