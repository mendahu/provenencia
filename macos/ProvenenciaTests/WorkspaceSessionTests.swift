import Foundation
import os
import Testing
@testable import Provenencia

/// Counts loader calls. Loaders run off the main actor and can overlap, so the
/// count is guarded by a lock.
private final class Counter: Sendable {
    private let value = OSAllocatedUnfairLock(initialState: 0)

    var count: Int { value.withLock { $0 } }

    func increment() { value.withLock { $0 += 1 } }
}

@Suite(.serialized)
@MainActor
struct WorkspaceSessionTests {
    private let projectDir = "/tmp/workspace-session.provenencia"

    private func makeSession(store: FakeStore = FakeStore()) -> WorkspaceSession {
        WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    }

    @Test func evidenceGraphIndexFollowsTheListRevision() {
        let session = makeSession()
        let key = session.evidenceGraphsKey
        #expect(session.evidenceGraphs(sourceID: "s1") == nil)
        #expect(session.evidenceGraphsBySource.isEmpty)

        session.setQueryValue(key, value: [
            evidenceGraph("g1", sourceID: "s1", name: "Default"),
            evidenceGraph("g2", sourceID: "s2", name: "Default"),
            evidenceGraph("g3", sourceID: "s1", name: "Jones"),
        ])
        #expect(session.evidenceGraph(id: "g3")?.name == "Jones")
        #expect(session.evidenceGraphs(sourceID: "s1")?.map(\.id) == ["g1", "g3"])
        #expect(session.evidenceGraphs(sourceID: "s2")?.map(\.id) == ["g2"])
        #expect(session.evidenceGraphs(sourceID: "s9")?.isEmpty == true)
        #expect(session.evidenceGraphsBySource.keys.sorted() == ["s1", "s2"])

        session.setQueryValue(key, value: [
            evidenceGraph("g1", sourceID: "s1", name: "Default"),
        ])
        #expect(session.evidenceGraph(id: "g3") == nil)
        #expect(session.evidenceGraphs(sourceID: "s1")?.map(\.id) == ["g1"])
        #expect(session.evidenceGraphsBySource.keys.sorted() == ["s1"])
    }

    @Test func evidenceGraphsReadFollowsTheHandle() async {
        struct TestFailure: Error {}
        let session = makeSession()
        func isLoading(_ read: QueryRead<[CatalogEvidenceGraph]>) -> Bool { read.isLoading }

        // No handle yet.
        #expect(isLoading(session.evidenceGraphsRead(sourceID: "s1")))

        // A handle that has not loaded yet.
        let pending: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        #expect(isLoading(session.evidenceGraphsRead(sourceID: "s1")))
        await waitForQuery(pending)

        // A failure with no value.
        let failedSession = makeSession()
        let failing = failedSession.ensureQuery(failedSession.evidenceGraphsKey) { () async throws -> [CatalogEvidenceGraph] in
            throw TestFailure()
        }
        await waitForQuery(failing)
        guard case .failed(let error) = failedSession.evidenceGraphsRead(sourceID: "s1") else {
            Issue.record("expected .failed")
            return
        }
        #expect(error is TestFailure)

        // Loaded: the Source's graphs, or loading when the list has none for it.
        let loadedSession = makeSession()
        loadedSession.setQueryValue(loadedSession.evidenceGraphsKey, value: [
            evidenceGraph("g1", sourceID: "s1", name: "Default"),
            evidenceGraph("g3", sourceID: "s1", name: "Jones"),
        ])
        guard case .loaded(let graphs) = loadedSession.evidenceGraphsRead(sourceID: "s1") else {
            Issue.record("expected .loaded")
            return
        }
        #expect(graphs.map(\.id) == ["g1", "g3"])
        #expect(isLoading(loadedSession.evidenceGraphsRead(sourceID: "s9")))
    }

    @Test func artifactIDsInNoGraphReadsTheIndex() {
        let session = makeSession()
        #expect(session.artifactIDsInNoGraph(sourceID: "s1", artifactIDs: ["a1"]) == nil)

        session.setQueryValue(session.evidenceGraphsKey, value: [
            evidenceGraph("g1", sourceID: "s1", name: "Default", artifactIDs: ["a1"]),
            evidenceGraph("g3", sourceID: "s1", name: "Jones", artifactIDs: ["a1", "a2"]),
            evidenceGraph("g2", sourceID: "s2", name: "Default", artifactIDs: ["a3"]),
        ])
        #expect(session.artifactIDsInNoGraph(sourceID: "s1", artifactIDs: ["a1", "a2", "a3"]) == ["a3"])
        #expect(session.artifactIDsInNoGraph(sourceID: "s1", artifactIDs: []) == [])
        // A list with no graph for the Source is a list not yet refreshed.
        #expect(session.artifactIDsInNoGraph(sourceID: "s9", artifactIDs: ["a1"]) == nil)
    }

    @Test func artifactIDsInNoGraphIsNilWhileTheListRefetches() async {
        let session = makeSession()
        let rows = [evidenceGraph("g1", sourceID: "s1", name: "Default", artifactIDs: ["a1"])]
        session.setQueryValue(session.evidenceGraphsKey, value: rows)
        #expect(session.artifactIDsInNoGraph(sourceID: "s1", artifactIDs: ["a1", "a2"]) == ["a2"])

        // A refetch of the loaded list: the old value stays readable, but it
        // may lack a link just saved, so nothing is marked until it settles.
        let hold = FakeHold()
        hold.isOn = true
        session.invalidate(session.evidenceGraphsKey)
        let handle = session.ensureQuery(session.evidenceGraphsKey) { () async throws -> [CatalogEvidenceGraph] in
            await hold.wait()
            return [evidenceGraph("g1", sourceID: "s1", name: "Default", artifactIDs: ["a1", "a2"])]
        }
        #expect(handle.isFetching)
        #expect(session.artifactIDsInNoGraph(sourceID: "s1", artifactIDs: ["a1", "a2"]) == nil)

        _ = await waitUntil { hold.count == 1 }
        hold.release()
        await waitForQuery(handle)
        #expect(session.artifactIDsInNoGraph(sourceID: "s1", artifactIDs: ["a1", "a2"]) == [])
    }

    @Test func queryReadAndProgressReads() async {
        struct TestFailure: Error {}
        let handle = QueryHandle<Int>()
        #expect(handle.read.isLoading)
        handle.applySuccess(3)
        guard case .loaded(let value) = handle.read else {
            Issue.record("expected .loaded")
            return
        }
        #expect(value == 3)
        handle.applyFailure(TestFailure(), clearValue: true)
        guard case .failed(let error) = handle.read else {
            Issue.record("expected .failed")
            return
        }
        #expect(error is TestFailure)
        handle.applySuccess(4)
        handle.applyFailure(TestFailure(), clearValue: false)
        guard case .loaded(let kept) = handle.read else {
            Issue.record("expected .loaded with the kept value")
            return
        }
        #expect(kept == 4)

        let store = FakeStore()
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 3, observationCount: 5
        )
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g-1", sourceID: "s1", name: "Default")]
        let session = makeSession(store: store)
        #expect(session.evidenceGraphCountsLoading)
        #expect(session.evidenceGraphProgress(graphID: "g-1") == nil)
        let progress: QueryHandle<[String: EvidenceGraphProgress]> = session.query(session.evidenceGraphProgressKey)
        let graphs: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        await waitForQuery(progress)
        await waitForQuery(graphs)
        #expect(session.evidenceGraphCountsLoading == false)
        #expect(session.evidenceGraphProgress(graphID: "g-1")?.subjectCount == 3)
    }

    @Test func progressFailureReadsUnavailable() {
        struct TestFailure: Error {}
        let session = makeSession()
        let progress: QueryHandle<[String: EvidenceGraphProgress]> = session.query(session.evidenceGraphProgressKey)
        progress.applyFailure(TestFailure(), clearValue: true)
        let error = session.evidenceGraphCountsError
        #expect(error is TestFailure)
        #expect(
            SourcesGraphCounts(progress: nil, countsLoading: false, countsError: error)
                == .unavailable(message: L10n.Errors.message(for: TestFailure()))
        )

        let kept = EvidenceGraphProgress(graphId: "g-1", subjectCount: 3, observationCount: 5)
        progress.applySuccess(["g-1": kept])
        progress.applyFailure(TestFailure(), clearValue: false)
        #expect(
            SourcesGraphCounts(
                progress: session.evidenceGraphProgress(graphID: "g-1"),
                countsLoading: false,
                countsError: session.evidenceGraphCountsError
            ) == .counts(subjects: 3, observations: 5, filed: 0)
        )
    }

    @Test func loadsAndCachesValue() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let counter = Counter()

        let handle = session.ensureQuery(key) {
            counter.increment()
            return ["alpha"]
        }
        await waitForQuery(handle)
        #expect(handle.status == .ready)
        #expect(handle.value == ["alpha"])
        #expect(counter.count == 1)

        _ = session.ensureQuery(key) {
            counter.increment()
            return ["beta"]
        }
        await Task.yield()
        #expect(counter.count == 1)
        #expect(handle.value == ["alpha"])
    }

    @Test func dedupesConcurrentLoads() async {
        let session = makeSession()
        let key = CatalogQueryKey.metadataFieldsList(project: session.projectKey)
        let counter = Counter()
        let gate = Gate()

        let loads = (0..<2).map { _ in
            Task {
                _ = session.ensureQuery(key) {
                    counter.increment()
                    await gate.waitForOpen()
                    return ["shared"]
                }
            }
        }
        gate.open()
        for load in loads { await load.value }

        let handle: QueryHandle<[String]> = session.queryHandle(key)!
        await waitForQuery(handle)
        #expect(counter.count == 1)
        #expect(handle.status == .ready)
        #expect(handle.value == ["shared"])
    }

    @Test func invalidateTriggersReload() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        let counter = Counter()

        let handle = session.ensureQuery(key) {
            counter.increment()
            return ["first"]
        }
        await waitForQuery(handle)
        #expect(counter.count == 1)

        session.invalidate(key)
        _ = session.ensureQuery(key) {
            counter.increment()
            return ["second"]
        }
        await waitForQuery(handle)
        #expect(counter.count == 2)
        #expect(handle.value == ["second"])
    }

    @Test func invalidateAllMatching() async {
        let session = makeSession()
        let otherProject = ProjectKey(projectDir: "/tmp/other.provenencia")
        let keyA = CatalogQueryKey.sourcesList(project: session.projectKey)
        let keyB = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        let keyOther = CatalogQueryKey.sourcesList(project: otherProject)
        let counter = Counter()

        let handleA = session.ensureQuery(keyA) {
            counter.increment()
            return ["a"]
        }
        let handleB = session.ensureQuery(keyB) {
            counter.increment()
            return ["b"]
        }
        let handleOther = session.ensureQuery(keyOther) {
            counter.increment()
            return ["other"]
        }
        await waitForQuery(handleA)
        await waitForQuery(handleB)
        await waitForQuery(handleOther)
        #expect(counter.count == 3)

        session.invalidateAll { $0.project == session.projectKey }

        _ = session.ensureQuery(keyA) {
            counter.increment()
            return ["a2"]
        }
        _ = session.ensureQuery(keyB) {
            counter.increment()
            return ["b2"]
        }
        _ = session.ensureQuery(keyOther) {
            counter.increment()
            return ["other2"]
        }
        await waitForQuery(handleA)
        await waitForQuery(handleB)
        await waitForQuery(handleOther)
        #expect(counter.count == 5)
        #expect(handleOther.value == ["other"])
    }

    @Test func staleWhileRevalidate() async {
        let session = makeSession()
        let key = CatalogQueryKey.typeSuggestions(project: session.projectKey, typeId: "t1")
        let counter = Counter()
        let gate = Gate()

        let handle = session.ensureQuery(key) {
            counter.increment()
            return ["first"]
        }
        await waitForQuery(handle)

        session.invalidate(key)
        _ = session.ensureQuery(key) {
            counter.increment()
            await gate.waitForOpen()
            return ["second"]
        }

        #expect(handle.value == ["first"])
        #expect(handle.isFetching)
        gate.open()
        await waitForQuery(handle)
        #expect(handle.value == ["second"])
        #expect(handle.isFetching == false)
        #expect(counter.count == 2)
    }

    @Test func propagatesLoaderError() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: "missing")
        struct TestFailure: Error {}

        let handle = session.ensureQuery(key) {
            throw TestFailure()
        }
        await waitForQuery(handle)
        #expect(handle.status == .error)
        #expect(handle.error is TestFailure)
        #expect(handle.value == nil)
    }

    private func seedStore(_ store: FakeStore) {
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Alpha", description: ""),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t1", key: "book", origin: "user", label: "Book", description: ""),
        ]
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f1", key: "author", origin: "user", label: "Author", dataType: "text", description: ""
            ),
        ]
        store.suggestionsByType["t1"] = [
            CatalogTypeSuggestion(
                field: CatalogMetadataField(
                    id: "f1", key: "author", origin: "user", label: "Author", dataType: "text", description: ""
                ),
                sortOrder: 0
            ),
        ]
    }

    @Test func applyLocationSourcesRoot() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey

        session.apply(location: .sectionRoot(.sources))

        let sourcesHandle: QueryHandle<[CatalogSource]>? = session.queryHandle(
            CatalogQueryKey.sourcesList(project: project)
        )
        let typesHandle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(
            CatalogQueryKey.sourceTypesList(project: project)
        )
        let progressHandle: QueryHandle<[String: EvidenceGraphProgress]>? = session.queryHandle(
            CatalogQueryKey.evidenceGraphProgress(project: project)
        )
        let fieldsHandle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(
            CatalogQueryKey.metadataFieldsList(project: project)
        )
        #expect(sourcesHandle != nil)
        #expect(typesHandle != nil)
        #expect(progressHandle != nil)
        #expect(fieldsHandle == nil)

        await waitForQuery(sourcesHandle!)
        await waitForQuery(typesHandle!)
        await waitForQuery(progressHandle!)
        #expect(sourcesHandle?.value?.first?.title == "Alpha")
        #expect(typesHandle?.value?.first?.label == "Book")
    }

    @Test func applyLocationSourceDetail() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey

        session.apply(location: WorkspaceLocation(section: .sources, sourceId: "s1", title: "Alpha"))

        let listHandle: QueryHandle<[CatalogSource]>? = session.queryHandle(
            CatalogQueryKey.sourcesList(project: project)
        )
        #expect(listHandle == nil)
        let workspaceHandle: QueryHandle<CatalogSourceWorkspace>? = session.queryHandle(
            CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        )
        #expect(workspaceHandle != nil)
        await waitForQuery(workspaceHandle!)
        #expect(workspaceHandle?.value?.source.id == "s1")
    }

    @Test func applyLocationSourceTypesWithSelection() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey

        session.apply(location: WorkspaceLocation(section: .sourceTypes, typeId: "t1", title: "Book"))

        let typesHandle: QueryHandle<[CatalogSourceType]>? = session.queryHandle(
            CatalogQueryKey.sourceTypesList(project: project)
        )
        let fieldsHandle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(
            CatalogQueryKey.metadataFieldsList(project: project)
        )
        let suggestionsHandle: QueryHandle<[CatalogTypeSuggestion]>? = session.queryHandle(
            CatalogQueryKey.typeSuggestions(project: project, typeId: "t1")
        )
        #expect(typesHandle != nil)
        #expect(fieldsHandle != nil)
        #expect(suggestionsHandle != nil)

        await waitForQuery(typesHandle!)
        await waitForQuery(fieldsHandle!)
        await waitForQuery(suggestionsHandle!)
        #expect(suggestionsHandle?.value?.count == 1)
    }

    @Test func applyLocationIsNonBlocking() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey

        session.apply(location: .sectionRoot(.sources))
        let handle: QueryHandle<[CatalogSource]>? = session.queryHandle(
            CatalogQueryKey.sourcesList(project: project)
        )
        #expect(handle?.status == .loading)

        await waitForQuery(handle!)
        #expect(handle?.status == .ready)
    }

    @Test func applyLocationCacheHit() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let location = WorkspaceLocation(section: .sources, sourceId: "s1")

        session.apply(location: location)
        let workspaceHandle: QueryHandle<CatalogSourceWorkspace>? = session.queryHandle(
            CatalogQueryKey.sourceWorkspace(project: session.projectKey, sourceId: "s1")
        )
        await waitForQuery(workspaceHandle!)

        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s2", ref: "SRC-2", sourceTypeID: "", title: "Beta", description: ""),
        ]

        session.apply(location: location)
        await Task.yield()
        #expect(workspaceHandle?.value?.source.id == "s1")
    }

    @Test func loadsSourcesListViaFakeStore() async throws {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "1", ref: "SRC-1", sourceTypeID: "", title: "A", description: ""),
        ]
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)

        let handle = session.ensureQuery(key) {
            try await store.listSources(projectDir: projectDir)
        }
        await waitForQuery(handle)
        #expect(handle.status == .ready)
        #expect(handle.value?.count == 1)
        #expect(handle.value?.first?.ref == "SRC-1")
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func applyCitationMutationReloadsGraphWithoutASecondQuery() async {
        let store = FakeStore()
        seedStore(store)
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person",
                key: "person",
                origin: "provenencia",
                label: "Person",
                description: "",
                refPrefix: "PER",
                candidateRefPrefix: "CPR"
            ),
        ]
        store.subjectsByGraph["g-1"] = [
            CatalogSubject(
                id: "sub-1",
                ref: "CPR-1",
                sourceID: "s1",
                subjectTypeID: "type-person",
                label: "Alice",
                description: ""
            ),
        ]
        store.subjectPositionsBySubject["sub-1"] = CatalogSubjectPosition(
            subjectID: "sub-1",
            gridX: 0,
            gridY: 0
        )
        let session = makeSession(store: store)
        let key = CatalogQueryKey.evidenceGraph(project: session.projectKey, graphId: "g-1")
        let handle: QueryHandle<EvidenceGraphRows> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.value?.observations.isEmpty == true)

        store.observationsByGraph["g-1"] = [
            CatalogObservation(
                id: "obs-1",
                ref: "OBS-1",
                citationID: "cit-1",
                subjectID: "sub-1",
                propertyID: "prop-name",
                polarity: "positive",
                valueText: "Alice",
                valueInteger: nil,
                valueDateID: "",
                valueNameID: "",
                valueSubjectID: "",
                valueTermID: "",
                propertyKey: "name",
                propertyLabel: "Name",
                propertyValueType: "name"
            ),
        ]
        session.apply(CatalogWriteEffects(evidenceGraphs: ["g-1"]))
        await waitForQuery(handle)
        #expect(handle.value?.observations.count == 1)
    }

    /// A Promote leaves a Persons list that isn't on screen stale: the old rows
    /// stay, nothing fetches, and the next `query` reloads from Go's cache.
    @Test func applyPromoteMarksHiddenPersonsListStale() async throws {
        let store = FakeStore()
        seedStore(store)
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.subjectsByGraph["g-1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "Alice", description: ""),
        ]
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForQuery(handle)
        #expect(handle.value?.isEmpty == true)

        let promoted = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")
        session.apply(.promotedSubject(graphId: "g-1"))
        #expect(!handle.isFetching)
        #expect(handle.value?.isEmpty == true)

        let _: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForQuery(handle)
        #expect(handle.value?.map(\.entity.ref) == [promoted.entity.ref])
    }

    /// The rule-2 exception (S9-15): a trigger evicts Conclusion detail
    /// pages that aren't on screen, revalidates the one that is, and marks the
    /// Persons list (also off screen) stale until its next query.
    @Test func applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne() async throws {
        let store = FakeStore()
        seedStore(store)
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.subjectsByGraph["g-1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
            CatalogSubject(id: "sub-2", ref: "CPR-2", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
        ]
        let first = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")
        let second = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-2")
        let session = makeSession(store: store)
        let firstKey = CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: first.entity.id)
        let secondKey = CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: second.entity.id)
        let list: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForQuery(list)

        session.apply(location: .personDetail(entityId: first.entity.id, ref: first.entity.ref, title: nil))
        let firstHandle: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(firstKey))
        await waitForQuery(firstHandle)
        session.apply(location: .personDetail(entityId: second.entity.id, ref: second.entity.ref, title: nil))
        #expect(session.visibleKeys == [secondKey])
        let secondHandle: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(secondKey))
        await waitForQuery(secondHandle)
        #expect(secondHandle.value?.fields.first?.state == .empty)

        store.observationsByGraph["g-1"] = [
            CatalogObservation(
                id: "o1", ref: "OBS-1", citationID: "c1", subjectID: "sub-2", propertyID: "p-name",
                polarity: "positive", valueText: "Jim Robins", valueInteger: nil, valueDateID: "",
                valueNameID: "n-1", nameForm: "Jim Robins",
                nameParts: [
                    CatalogNameValuePart(value: "Jim", type: "given"),
                    CatalogNameValuePart(value: "Robins", type: "surname"),
                ],
                valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            ),
        ]
        session.apply(CatalogWriteEffects(handles: [first.entity.id, second.entity.id], evidenceGraphs: ["g-1"]))

        // Off screen: dropped, so it reloads only when next visited.
        let evicted: QueryHandle<CatalogConclusionDetail>? = session.queryHandle(firstKey)
        #expect(evicted == nil)
        // On screen: revalidated in place.
        await waitForQuery(secondHandle)
        #expect(secondHandle.value?.fields.first?.state == .single)
        // The list is off screen: stale, not reloaded. The next query reloads it.
        #expect(!list.isFetching)
        #expect(list.value?.first?.name?.form != "Jim Robins")
        let _: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForQuery(list)
        #expect(list.value?.first?.name?.form == "Jim Robins")

        // Leaving for the list hides the second page too.
        session.apply(location: .sectionRoot(.persons))
        session.apply(CatalogWriteEffects(handles: [first.entity.id, second.entity.id], evidenceGraphs: ["g-1"]))
        let hidden: QueryHandle<CatalogConclusionDetail>? = session.queryHandle(secondKey)
        #expect(hidden == nil)
        await waitForQuery(list)
        #expect(list.status == .ready)

        // Revisiting loads afresh.
        session.apply(location: .personDetail(entityId: first.entity.id, ref: first.entity.ref, title: nil))
        let reloaded: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(firstKey))
        await waitForQuery(reloaded)
        #expect(reloaded.value?.entity == first.entity)
    }

    private func personHeader(_ id: String, revision: Int64, at: String = "") -> CatalogPersonHeader {
        CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: id, ref: "PER-\(id)", subjectTypeID: "t", label: ""),
            name: nil,
            nameValueCount: 0,
            updated: CatalogActivity(revision: revision, at: at)
        )
    }

    private func eventHeader(_ id: String) -> CatalogEventHeader {
        CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: id, ref: "EVT-\(id)", subjectTypeID: "t", label: ""),
            title: CatalogEventTitle(rule: .ref, ref: "EVT-\(id)")
        )
    }

    private func placeHeader(_ id: String) -> CatalogPlaceHeader {
        CatalogPlaceHeader(
            entity: CatalogCanonicalEntity(id: id, ref: "PLC-\(id)", subjectTypeID: "t", label: "")
        )
    }

    /// A write that reports its activity moves the cached rows it names, in
    /// place: no list load starts, an older or empty activity moves nothing, and
    /// a load already in flight is restarted so its old answer cannot win.
    @Test func applyEffectsPatchesLastUpdatedInPlace() async throws {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let people = CatalogQueryKey.personsList(project: session.projectKey)
        let sources = CatalogQueryKey.sourcesList(project: session.projectKey)
        let at = "2026-10-15T16:00:00Z"
        session.setQueryValue(people, value: [personHeader("e1", revision: 10), personHeader("e2", revision: 9)])
        session.setQueryValue(sources, value: [
            CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Alpha", description: "",
                          updated: CatalogActivity(revision: 5)),
        ])
        let list: QueryHandle<[CatalogPersonHeader]> = try #require(session.queryHandle(people))
        let sourceList: QueryHandle<[CatalogSource]> = try #require(session.queryHandle(sources))

        // AC-2.5: only e1 moves, and nothing loads.
        session.apply(CatalogWriteEffects(handleScopes: ["e1"], activity: CatalogActivity(revision: 12, at: at)))
        #expect(list.value?.map(\.updated) == [CatalogActivity(revision: 12, at: at), CatalogActivity(revision: 9)])
        #expect(!list.isFetching)
        // An older revision, and revision 0, change nothing.
        session.apply(CatalogWriteEffects(handleScopes: ["e1"], activity: CatalogActivity(revision: 11, at: "x")))
        session.apply(CatalogWriteEffects(handleScopes: ["e1", "e2"], activity: CatalogActivity()))
        #expect(list.value?.map(\.updated) == [CatalogActivity(revision: 12, at: at), CatalogActivity(revision: 9)])

        // AC-2.5b: a Source moves and the Sources list is neither loaded nor marked stale.
        session.apply(CatalogWriteEffects(sourceScopes: ["s1"], activity: CatalogActivity(revision: 12, at: at)))
        #expect(sourceList.value?.first?.updated == CatalogActivity(revision: 12, at: at))
        #expect(!sourceList.isFetching)
        let _: QueryHandle<[CatalogSource]> = session.query(sources)
        await Task.yield()
        #expect(!sourceList.isFetching)
        #expect(sourceList.value?.first?.updated.revision == 12)

        // AC-2.5c: a load in flight is cancelled and restarted.
        session.invalidate(people)
        let oldRows = [personHeader("e1", revision: 10)]
        let slow = session.ensureQuery(people) {
            try await Task.sleep(nanoseconds: 200_000_000)
            return oldRows
        }
        #expect(slow.isFetching)
        session.apply(CatalogWriteEffects(handleScopes: ["e2"], activity: CatalogActivity(revision: 20, at: at)))
        #expect(list.value?.last?.updated == CatalogActivity(revision: 20, at: at))
        #expect(list.isFetching)
        try await Task.sleep(nanoseconds: 300_000_000)
        await waitForQuery(list)
        // The restarted load answered from the store (empty), not the slow loader's old rows.
        #expect(list.value?.isEmpty == true)
    }

    /// A write that changes list values reloads the list on screen and leaves
    /// the others stale until they are shown.
    @Test func hiddenListsGoStaleVisibleListReloads() async throws {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let people = CatalogQueryKey.personsList(project: project)
        let events = CatalogQueryKey.eventsList(project: project)
        let places = CatalogQueryKey.placesList(project: project)
        func seed() throws -> (QueryHandle<[CatalogPersonHeader]>, QueryHandle<[CatalogEventHeader]>, QueryHandle<[CatalogPlaceHeader]>) {
            session.setQueryValue(people, value: [personHeader("e1", revision: 1)])
            session.setQueryValue(events, value: [eventHeader("e1")])
            session.setQueryValue(places, value: [placeHeader("e1")])
            return (
                try #require(session.queryHandle(people)),
                try #require(session.queryHandle(events)),
                try #require(session.queryHandle(places))
            )
        }

        // AC-2.6: none on screen. Nothing loads; the old value shows; the next query reloads.
        var (peopleHandle, eventsHandle, placesHandle) = try seed()
        session.apply(CatalogWriteEffects(handles: ["e1"]))
        #expect(!peopleHandle.isFetching && !eventsHandle.isFetching && !placesHandle.isFetching)
        #expect(peopleHandle.value?.count == 1 && eventsHandle.value?.count == 1 && placesHandle.value?.count == 1)
        let _: QueryHandle<[CatalogPersonHeader]> = session.query(people)
        await waitForQuery(peopleHandle)
        #expect(peopleHandle.value?.isEmpty == true)

        // AC-2.7: People on screen. It reloads now; Events and Places go stale.
        session.apply(location: .sectionRoot(.persons))
        await waitForQuery(peopleHandle)
        (peopleHandle, eventsHandle, placesHandle) = try seed()
        session.apply(CatalogWriteEffects(handles: ["e1"]))
        #expect(peopleHandle.isFetching)
        await waitForQuery(peopleHandle)
        #expect(peopleHandle.value?.isEmpty == true)
        #expect(!eventsHandle.isFetching && !placesHandle.isFetching)
        #expect(eventsHandle.value?.count == 1 && placesHandle.value?.count == 1)
        // The stale Events list reloads on its next showing.
        let _: QueryHandle<[CatalogEventHeader]> = session.query(events)
        #expect(eventsHandle.isFetching)
        await waitForQuery(eventsHandle)
        #expect(eventsHandle.value?.isEmpty == true)
    }

    /// A Source-page write reports its effects: the Sources row is patched in
    /// place, the three conclusion lists are neither reloaded nor stale unless
    /// the write reports handles, and the page's own workspace follows the
    /// mutation.
    @Test func sourceWriteEffectsPatchSourcesAndLeaveListsAlone() async throws {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let sources = CatalogQueryKey.sourcesList(project: project)
        let people = CatalogQueryKey.personsList(project: project)
        let events = CatalogQueryKey.eventsList(project: project)
        let places = CatalogQueryKey.placesList(project: project)
        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        session.apply(location: WorkspaceLocation(section: .sources, sourceId: "s1", title: "Alpha"))
        let workspace: QueryHandle<CatalogSourceWorkspace> = try #require(session.queryHandle(workspaceKey))
        await waitForQuery(workspace)
        session.setQueryValue(sources, value: [
            CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Alpha", description: ""),
        ])
        session.setQueryValue(people, value: [personHeader("e1", revision: 1)])
        session.setQueryValue(events, value: [eventHeader("e1")])
        session.setQueryValue(places, value: [placeHeader("e1")])
        let sourcesHandle: QueryHandle<[CatalogSource]> = try #require(session.queryHandle(sources))
        let peopleHandle: QueryHandle<[CatalogPersonHeader]> = try #require(session.queryHandle(people))
        let eventsHandle: QueryHandle<[CatalogEventHeader]> = try #require(session.queryHandle(events))
        let placesHandle: QueryHandle<[CatalogPlaceHeader]> = try #require(session.queryHandle(places))
        func listsAreQuiet() -> Bool {
            !sourcesHandle.isFetching && !peopleHandle.isFetching && !eventsHandle.isFetching
                && !placesHandle.isFetching
        }

        // AC-2.11: a note. Sources row patched; the conclusion lists untouched
        // and not stale; the page's workspace invalidated once.
        let first = CatalogActivity(revision: 12, at: "2026-10-15T16:00:00Z")
        session.apply(CatalogWriteEffects(sourceScopes: ["s1"], activity: first), .mutatedSourceWorkspace(sourceId: "s1"))
        #expect(sourcesHandle.value?.first?.updated == first)
        #expect(listsAreQuiet())
        let _: QueryHandle<[CatalogPersonHeader]> = session.query(people)
        let _: QueryHandle<[CatalogEventHeader]> = session.query(events)
        let _: QueryHandle<[CatalogPlaceHeader]> = session.query(places)
        let _: QueryHandle<[CatalogSource]> = session.query(sources)
        #expect(listsAreQuiet())
        #expect(workspace.isFetching)
        await waitForQuery(workspace)

        // AC-2.12: a credibility change reports the handle. The hidden lists
        // go stale and reload on their next `query`; the Sources row still patches.
        let second = CatalogActivity(revision: 13, at: "2026-10-15T16:05:00Z")
        session.apply(
            CatalogWriteEffects(handles: ["e1"], sourceScopes: ["s1"], activity: second),
            .mutatedSourceWorkspace(sourceId: "s1")
        )
        #expect(sourcesHandle.value?.first?.updated == second)
        #expect(listsAreQuiet())
        let _: QueryHandle<[CatalogPersonHeader]> = session.query(people)
        #expect(peopleHandle.isFetching)
        await waitForQuery(peopleHandle)
        let _: QueryHandle<[CatalogSource]> = session.query(sources)
        #expect(!sourcesHandle.isFetching)

        // AC-2.12, on screen: a Conclusion page showing the reported handle reloads.
        session.apply(location: .personDetail(entityId: "e1", ref: "PER-e1", title: nil))
        let detailKey = CatalogQueryKey.conclusionDetail(project: project, entityId: "e1")
        let detail: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(detailKey))
        await waitForQuery(detail)
        session.apply(
            CatalogWriteEffects(handles: ["e1"], sourceScopes: ["s1"], activity: CatalogActivity(revision: 14, at: "2026-10-15T16:10:00Z")),
            .mutatedSourceWorkspace(sourceId: "s1")
        )
        // The page has no value in the fake, so a reload shows as loading.
        #expect(detail.status == .loading)
        await waitForQuery(detail)
    }

    /// A note moves the Source's last activity and nothing it reads as
    /// evidence: the Properties workspace, the Source graph, and the citations
    /// stay warm and current (D-25, D-26).
    @Test func sourceNoteLeavesEvidenceReadsAlone() async throws {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let sources = CatalogQueryKey.sourcesList(project: project)
        let propertiesKey = CatalogQueryKey.propertiesWorkspace(project: project)
        let graphKey = CatalogQueryKey.evidenceGraph(project: project, graphId: "g-1")
        let citationsKey = CatalogQueryKey.citationsByArtifact(project: project, graphId: "g-1", artifactId: "a1")
        session.setQueryValue(sources, value: [
            CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Alpha", description: ""),
        ])
        session.setQueryValue(propertiesKey, value: PropertiesSnapshot.empty)
        session.setQueryValue(graphKey, value: EvidenceGraphRows(graphId: "g-1"))
        session.setQueryValue(citationsKey, value: [CatalogListedCitation]())
        let activity = CatalogActivity(revision: 12, at: "2026-10-15T16:00:00Z")

        session.apply(
            CatalogWriteEffects(sourceScopes: ["s1"], activity: activity),
            .mutatedSourceWorkspace(sourceId: "s1")
        )

        let list: QueryHandle<[CatalogSource]> = try #require(session.queryHandle(sources))
        #expect(list.value?.first?.updated == activity)
        // Nothing is fetching, and a `query` finds nothing stale to reload.
        let properties: QueryHandle<PropertiesSnapshot> = session.query(propertiesKey)
        let graph: QueryHandle<EvidenceGraphRows> = session.query(graphKey)
        let citations: QueryHandle<[CatalogListedCitation]> = session.query(citationsKey)
        #expect(!properties.isFetching && !graph.isFetching && !citations.isFetching)
    }

    private func detail(_ id: String) -> CatalogConclusionDetail {
        CatalogConclusionDetail(
            entity: CatalogCanonicalEntity(id: id, ref: "PER-\(id)", subjectTypeID: "t", label: ""),
            fields: []
        )
    }

    private func claims(onGraph graphID: String) -> [CatalogPersonClaimGroup] {
        [CatalogPersonClaimGroup(
            sourceID: "s1", sourceRef: "SRC-1", title: "Alpha", sourceTypeID: "t1",
            claims: [CatalogPersonClaim(
                claimID: "c1", graphID: graphID, graphRef: "EG-1", graphName: "Default",
                subject: CatalogSubject(
                    id: "sub1", ref: "SUB-1", sourceID: "s1", subjectTypeID: "t", label: "Ada", description: ""
                ),
                observations: []
            )]
        )]
    }

    /// T-7 (AC-2.1): a write's touched handles revalidate only their own detail
    /// pages. A visible page reloads; a hidden one is evicted.
    @Test func touchedHandlesRevalidateOnlyTheirDetailPages() async {
        let store = FakeStore()
        seedStore(store)
        store.seededConclusionDetails = ["e1": detail("e1"), "e3": detail("e3")]
        let session = makeSession(store: store)
        let project = session.projectKey
        session.apply(location: .personDetail(entityId: "e1", ref: "PER-e1", title: "One"))
        let first = CatalogQueryKey.conclusionDetail(project: project, entityId: "e1")
        let third = CatalogQueryKey.conclusionDetail(project: project, entityId: "e3")
        let visible: QueryHandle<CatalogConclusionDetail> = session.query(first)
        let hidden: QueryHandle<CatalogConclusionDetail> = session.query(third)
        await waitForQuery(visible)
        await waitForQuery(hidden)

        let effects = CatalogWriteEffects(handles: ["e1", "e2"])
        #expect(session.revalidatedKeys(for: effects) == [first])
        #expect(CatalogQueryRegistry.standard.invalidations(for: effects, project: project).isEmpty)
        session.apply(effects)
        #expect(visible.isFetching)
        await session.settleWrites()
        #expect(session.queryHandle(third) as QueryHandle<CatalogConclusionDetail>? != nil)

        session.apply(CatalogWriteEffects(handles: ["e3"]))
        #expect(session.queryHandle(third) as QueryHandle<CatalogConclusionDetail>? == nil)
    }

    /// T-8 (AC-2.6): a key that two refreshes and a mutation all name is
    /// revalidated once. A hidden `.personClaims` key is evicted by the pass
    /// that names it, and nothing loads it again.
    @Test func aKeyInvalidatedAndRefreshedRevalidatesOnce() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let key = CatalogQueryKey.personClaims(project: project, entityId: "e1")
        session.setQueryValue(key, value: claims(onGraph: "g1"))
        let effects = CatalogWriteEffects(handles: ["e1"], evidenceGraphs: ["g1"])

        #expect(session.revalidatedKeys(for: effects, .promotedSubject(graphId: "g1")) == [key])
        #expect(CatalogQueryRegistry.standard.invalidations(for: effects, by: .promotedSubject(graphId: "g1"), project: project)
            .contains(.allCached(.personClaims)))
        session.apply(effects, .promotedSubject(graphId: "g1"))
        await session.settleWrites()
        #expect(session.queryHandle(key) as QueryHandle<[CatalogPersonClaimGroup]>? == nil)
    }

    /// T-9 (AC-2.1, AC-2.3): person claims follow their handle and the graphs
    /// they hold a claim on; the detail page follows only the handle.
    @Test func personClaimsFollowHandlesAndGraphs() {
        let session = makeSession()
        let project = session.projectKey
        let claimsKey = CatalogQueryKey.personClaims(project: project, entityId: "e1")
        let detailKey = CatalogQueryKey.conclusionDetail(project: project, entityId: "e1")
        session.setQueryValue(claimsKey, value: claims(onGraph: "g1"))
        session.setQueryValue(detailKey, value: detail("e1"))

        #expect(session.revalidatedKeys(for: CatalogWriteEffects(evidenceGraphs: ["g1"])) == [claimsKey])
        #expect(session.revalidatedKeys(for: CatalogWriteEffects(evidenceGraphs: ["g9"])).isEmpty)
        #expect(Set(session.revalidatedKeys(for: CatalogWriteEffects(handles: ["e1"]))) == [claimsKey, detailKey])
    }

    /// A `.personClaims` key whose first load is in flight follows every
    /// graph; once it holds a value it follows only the graphs it holds.
    @Test func personClaimsFollowEveryGraphUntilLoaded() {
        let session = makeSession()
        let project = session.projectKey
        let key = CatalogQueryKey.personClaims(project: project, entityId: "e1")
        let _: QueryHandle<[CatalogPersonClaimGroup]> = session.query(key)
        #expect((session.queryHandle(key) as QueryHandle<[CatalogPersonClaimGroup]>?)?.value == nil)
        #expect(session.revalidatedKeys(for: CatalogWriteEffects(evidenceGraphs: ["g9"])) == [key])

        session.setQueryValue(key, value: claims(onGraph: "g1"))
        #expect(session.revalidatedKeys(for: CatalogWriteEffects(evidenceGraphs: ["g9"])).isEmpty)
        #expect(session.revalidatedKeys(for: CatalogWriteEffects(evidenceGraphs: ["g1"])) == [key])
    }

    /// T-12 (AC-2.7): the Evidence graph and list keys declare the slots they
    /// follow, so a write reloads the graphs it names and every list it touches.
    @Test func effectSlotsRevalidateTheirKeys() {
        let session = makeSession()
        let project = session.projectKey
        let graph1 = CatalogQueryKey.evidenceGraph(project: project, graphId: "g1")
        let graph2 = CatalogQueryKey.evidenceGraph(project: project, graphId: "g2")
        session.setQueryValue(graph1, value: EvidenceGraphRows(graphId: "g1"))
        session.setQueryValue(graph2, value: EvidenceGraphRows(graphId: "g2"))
        session.setQueryValue(CatalogQueryKey.personsList(project: project), value: [CatalogPersonHeader]())
        session.setQueryValue(CatalogQueryKey.eventsList(project: project), value: [CatalogEventHeader]())
        session.setQueryValue(CatalogQueryKey.placesList(project: project), value: [CatalogPlaceHeader]())

        let named = session.revalidatedKeys(for: CatalogWriteEffects(handles: ["e1"], evidenceGraphs: ["g1"]))
        #expect(Set(named) == [
            graph1, .personsList(project: project), .eventsList(project: project), .placesList(project: project),
        ])
        #expect(named.count == 4)
        #expect(session.revalidatedKeys(for: CatalogWriteEffects(handleScopes: ["e1"])).isEmpty)
    }

    private func citation(_ id: String, artifactID: String) -> CatalogCitation {
        CatalogCitation(
            id: id, ref: "CIT-\(id)", artifactID: artifactID, locatorJSON: "{}", transcription: "",
            description: "", transcriptionUncertain: false, transcriptionNote: ""
        )
    }

    /// Each read follows only what feeds it (D-27): with the graph on screen, a
    /// Subject create reloads the graph and nothing else, and an Observation
    /// create also reloads the citation counts and the Properties `usedBy`.
    @Test func eachReadFollowsOnlyWhatFeedsIt() async throws {
        let store = FakeStore()
        seedStore(store)
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g-1", sourceID: "s1", artifactIDs: ["a1"])]
        store.citationsByID["c1"] = citation("c1", artifactID: "a1")
        let session = makeSession(store: store)
        let project = session.projectKey
        session.apply(location: WorkspaceLocation(
            section: .sources, sourceId: "s1", graphId: "g-1", sourceSurface: .graph, title: "Alpha"
        ))
        let propertiesKey = CatalogQueryKey.propertiesWorkspace(project: project)
        let countsKey = CatalogQueryKey.citationCounts(project: project, graphId: "g-1")
        #expect(session.visibleKeys.contains(propertiesKey))
        #expect(session.visibleKeys.contains(countsKey))
        let properties: QueryHandle<PropertiesSnapshot> = session.query(propertiesKey)
        let counts: QueryHandle<[String: Int]> = session.query(countsKey)
        await waitForQuery(properties)
        await waitForQuery(counts)
        #expect(counts.value == ["a1": 1])

        store.citationsByID["c2"] = citation("c2", artifactID: "a1")
        session.apply(CatalogWriteEffects(evidenceGraphs: ["g-1"]))
        #expect(!properties.isFetching && !counts.isFetching)
        try? await Task.sleep(nanoseconds: 50_000_000)
        #expect(counts.value == ["a1": 1])

        session.apply(CatalogWriteEffects(
            evidenceGraphs: ["g-1"],
            reads: [CatalogWriteEffectsRead(read: .citations, ids: ["g-1"])],
            lists: [.propertyUsage]
        ))
        #expect(properties.isFetching)
        #expect(await waitUntil { counts.value == ["a1": 2] })
    }

    /// T-7: a `citations` read for one graph reloads that graph's visible
    /// counts and leaves another graph's warmed counts alone. The fake counts every
    /// seeded Citation on its Artifact for any graph, so the reload shows in which
    /// graph the store was asked, not in the values.
    @Test func citationReadsRefreshOnlyTheirGraph() async throws {
        let store = FakeStore()
        seedStore(store)
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("g-1", sourceID: "s1", artifactIDs: ["a1"]),
            evidenceGraph("g-2", sourceID: "s1", artifactIDs: ["a2"]),
        ]
        store.citationsByID["c1"] = citation("c1", artifactID: "a1")
        store.citationsByID["c2"] = citation("c2", artifactID: "a2")
        let session = makeSession(store: store)
        let project = session.projectKey
        session.apply(location: WorkspaceLocation(
            section: .sources, sourceId: "s1", graphId: "g-1", sourceSurface: .graph, title: "Alpha"
        ))
        let first: QueryHandle<[String: Int]> = session.query(.citationCounts(project: project, graphId: "g-1"))
        let second: QueryHandle<[String: Int]> = session.query(.citationCounts(project: project, graphId: "g-2"))
        await waitForQuery(first)
        await waitForQuery(second)
        #expect(first.value == ["a1": 1, "a2": 1] && second.value == ["a1": 1, "a2": 1])
        let loaded = store.citationCountsGraphIDs
        #expect(loaded.filter { $0 == "g-1" }.count == 1 && loaded.filter { $0 == "g-2" }.count == 1)

        store.citationsByID["c3"] = citation("c3", artifactID: "a1")
        session.apply(CatalogWriteEffects(reads: [CatalogWriteEffectsRead(read: .citations, ids: ["g-1"])]))

        await session.settleWrites()
        #expect(first.value == ["a1": 2, "a2": 1])
        #expect(store.citationCountsGraphIDs.filter { $0 == "g-1" }.count == 2)
        #expect(store.citationCountsGraphIDs.filter { $0 == "g-2" }.count == 1)
        #expect(second.value == ["a1": 1, "a2": 1])
    }

    /// T-8: the composer's location names its Artifact, so the Artifact's
    /// citations key is visible and `applyWrite` is `apply` alone; the key
    /// revalidates from the write's `citations` read, with no `query` after.
    @Test func composerWriteReloadsTheArtifactListItShows() async throws {
        let store = FakeStore()
        seedStore(store)
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g-1", sourceID: "s1", artifactIDs: ["a1"])]
        store.citationsByID["c1"] = citation("c1", artifactID: "a1")
        let session = makeSession(store: store)
        let project = session.projectKey
        let key = CatalogQueryKey.citationsByArtifact(project: project, graphId: "g-1", artifactId: "a1")
        session.apply(location: WorkspaceLocation(
            section: .sources, sourceId: "s1", graphId: "g-1", citationId: "c1", artifactId: "a1",
            sourceSurface: .citationComposer, title: "Alpha"
        ))
        #expect(session.visibleKeys.contains(key))
        let handle: QueryHandle<[CatalogListedCitation]> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.value?.count == 1)
        let context = CitationComposerContext(
            store: store, session: session, userID: "u1", sourceID: "s1", graphID: "g-1"
        )
        context.artifactID = "a1"

        store.citationsByID["c2"] = citation("c2", artifactID: "a1")
        context.applyWrite(CatalogWriteEffects(
            evidenceGraphs: ["g-1"], reads: [CatalogWriteEffectsRead(read: .citations, ids: ["g-1"])]
        ))

        #expect(await waitUntil { handle.value?.count == 2 })
    }

    /// `usedBy` on the Properties inspector counts Observations, and it
    /// moves whenever a citation save creates or deletes one.
    @Test func applySavedCitationInvalidatesPropertiesUsedBy() async {
        let store = FakeStore()
        seedStore(store)
        store.propertiesByProject[projectDir] = [
            CatalogProperty(
                id: "prop-name",
                key: "name",
                origin: "provenencia",
                label: "Name",
                description: "",
                valueType: "name",
                usedBy: 2
            ),
        ]
        let session = makeSession(store: store)
        let fieldsKey = CatalogQueryKey.propertiesWorkspace(project: session.projectKey)
        let fields: QueryHandle<PropertiesSnapshot> = session.query(fieldsKey)
        await waitForQuery(fields)
        #expect(fields.value?.properties.first?.usedBy == 2)

        store.propertiesByProject[projectDir]?[0].usedBy = 1
        session.apply(CatalogWriteEffects(evidenceGraphs: ["g-1"], lists: [.propertyUsage]))

        let _: QueryHandle<PropertiesSnapshot> = session.query(fieldsKey)
        await waitForQuery(fields)
        #expect(fields.value?.properties.first?.usedBy == 1)
    }

    @Test func readyValueIgnoresStaleCache() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let handle = session.ensureQuery(key) { ["alpha"] }
        await waitForQuery(handle)
        session.invalidate(key)
        let ready: [String]? = await session.readyValue(key)
        #expect(ready == nil)
    }

    @Test func staleOlderLoadDoesNotOverwriteNewer() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let olderGate = Gate()
        let handle = session.ensureQuery(key) {
            await olderGate.waitForOpen()
            return ["old"]
        }
        session.invalidate(key)
        _ = session.ensureQuery(key) { ["new"] }
        await waitForQuery(handle)
        #expect(handle.value == ["new"])
        olderGate.open()
        await Task.yield()
        #expect(handle.value == ["new"])
        #expect(handle.isFetching == false)
    }

    @Test func setQueryValueWinsOverInFlightLoad() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let gate = Gate()
        let stale = CatalogSource(
            id: "stale",
            ref: "SRC-S",
            sourceTypeID: "t",
            title: "Stale",
            description: ""
        )
        let patched = CatalogSource(
            id: "patched",
            ref: "SRC-P",
            sourceTypeID: "t",
            title: "Patched",
            description: ""
        )
        let handle = session.ensureQuery(key) {
            await gate.waitForOpen()
            return [stale]
        }
        session.setQueryValue(key, value: [patched])
        #expect(handle.value?.first?.id == "patched")
        gate.open()
        await Task.yield()
        #expect(handle.value?.first?.id != "stale")
    }

    @Test func setQueryValueRestartsInFlightLoad() async {
        let store = FakeStore()
        store.sourcesByProject[projectDir] = [
            CatalogSource(
                id: "s1",
                ref: "SRC-1",
                sourceTypeID: "t",
                title: "Census",
                description: ""
            ),
        ]
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let gate = Gate()
        let handle = session.ensureQuery(key) {
            await gate.waitForOpen()
            return [CatalogSource]()
        }
        session.setQueryValue(
            key,
            value: [
                CatalogSource(
                    id: "patched",
                    ref: "SRC-P",
                    sourceTypeID: "t",
                    title: "Patched",
                    description: ""
                ),
            ]
        )
        #expect(handle.value?.first?.id == "patched")
        await waitForQuery(handle)
        #expect(handle.value?.first?.id == "s1")
    }

    /// T-15: a write's effects name the graphs whose progress to patch; the
    /// other graph of the same Source keeps its entry and is not fetched.
    @Test func applyGraphEffectsPatchProgressOnly() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 12, observationCount: 48
        )
        store.graphProgressByGraph["g-2"] = EvidenceGraphProgress(
            graphId: "g-2", subjectCount: 4, observationCount: 0
        )
        let session = makeSession(store: store)
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        let progressKey = CatalogQueryKey.evidenceGraphProgress(project: session.projectKey)
        let listHandle: QueryHandle<[CatalogSource]> = session.query(listKey)
        let progressHandle: QueryHandle<[String: EvidenceGraphProgress]> = session.query(progressKey)
        await waitForQuery(listHandle)
        await waitForQuery(progressHandle)
        #expect(progressHandle.value?["g-1"]?.subjectCount == 12)
        #expect(progressHandle.value?["g-2"]?.subjectCount == 4)
        let listCallsBefore = store.listEvidenceGraphProgressCalls
        let getCallsBefore = store.getEvidenceGraphProgressCalls
        let sourcesAtLoad = listHandle.value

        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 99, observationCount: 48
        )
        store.graphProgressByGraph["g-2"] = EvidenceGraphProgress(
            graphId: "g-2", subjectCount: 5, observationCount: 0
        )
        session.apply(CatalogWriteEffects(
            evidenceGraphs: ["g-2"], reads: [CatalogWriteEffectsRead(read: .evidenceGraphProgress, ids: ["g-2"])]
        ))
        var waited: UInt64 = 0
        while waited < 2_000_000_000 {
            if progressHandle.value?["g-2"]?.subjectCount == 5 { break }
            await Task.yield()
            try? await Task.sleep(nanoseconds: 10_000_000)
            waited += 10_000_000
        }
        #expect(progressHandle.value?["g-2"]?.subjectCount == 5)
        #expect(progressHandle.value?["g-1"]?.subjectCount == 12)
        #expect(listHandle.value == sourcesAtLoad)
        #expect(listHandle.isFetching == false)
        #expect(store.listEvidenceGraphProgressCalls == listCallsBefore)
        #expect(store.getEvidenceGraphProgressCalls == getCallsBefore + 1)
    }

    @Test func applyPromotedSubjectPatchesProgress() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 2, observationCount: 1, filedCount: 0
        )
        let session = makeSession(store: store)
        let progressHandle: QueryHandle<[String: EvidenceGraphProgress]> = session.query(
            CatalogQueryKey.evidenceGraphProgress(project: session.projectKey)
        )
        await waitForQuery(progressHandle)
        #expect(progressHandle.value?["g-1"]?.filedCount == 0)
        let callsBefore = store.getEvidenceGraphProgressCalls

        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 2, observationCount: 1, filedCount: 2
        )
        session.apply(.promotedSubject(graphId: "g-1"))
        var waited: UInt64 = 0
        while waited < 2_000_000_000 {
            if progressHandle.value?["g-1"]?.filedCount == 2 { break }
            await Task.yield()
            try? await Task.sleep(nanoseconds: 10_000_000)
            waited += 10_000_000
        }
        #expect(store.getEvidenceGraphProgressCalls == callsBefore + 1)
        #expect(progressHandle.value?["g-1"]?.filedCount == 2)
    }

    @Test func applySourceEffectsDoesNotRefetchSourcesList() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 1, observationCount: 1
        )
        let session = makeSession(store: store)
        let listHandle: QueryHandle<[CatalogSource]> = session.query(
            CatalogQueryKey.sourcesList(project: session.projectKey)
        )
        let progressHandle: QueryHandle<[String: EvidenceGraphProgress]> = session.query(
            CatalogQueryKey.evidenceGraphProgress(project: session.projectKey)
        )
        await waitForQuery(listHandle)
        await waitForQuery(progressHandle)
        let sourcesAtLoad = listHandle.value
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 1, observationCount: 2
        )
        session.apply(CatalogWriteEffects(
            evidenceGraphs: ["g-1"], reads: [CatalogWriteEffectsRead(read: .evidenceGraphProgress, ids: ["g-1"])]
        ))
        var waited: UInt64 = 0
        while waited < 2_000_000_000 {
            if progressHandle.value?["g-1"]?.observationCount == 2 { break }
            await Task.yield()
            try? await Task.sleep(nanoseconds: 10_000_000)
            waited += 10_000_000
        }
        #expect(progressHandle.value?["g-1"]?.observationCount == 2)
        #expect(listHandle.value == sourcesAtLoad)
        #expect(listHandle.isFetching == false)
    }

    @Test func graphCameraIsPerGraph() {
        let session = makeSession()
        let first = GraphCanvasCamera(
            magnification: 0.5,
            contentOffset: CGPoint(x: 120, y: 80)
        )
        session.setGraphCamera(first, for: "g-1")
        #expect(session.graphCamera(for: "g-1") == first)
        #expect(session.graphCamera(for: "g-2") == nil)

        let second = GraphCanvasCamera(
            magnification: 1.25,
            contentOffset: CGPoint(x: 10, y: 20)
        )
        session.setGraphCamera(second, for: "g-2")
        #expect(session.graphCamera(for: "g-1") == first)
        #expect(session.graphCamera(for: "g-2") == second)
        #expect(session.graphCamera(for: "g-1") != session.graphCamera(for: "g-2"))
    }

    @Test func applyRunsNavCountsHookForTriggers() async {
        let session = makeSession()
        var calls = 0
        session.onNavCountsMoved = { calls += 1 }
        session.apply(.createdSource)
        session.apply(.deletedSource(id: "s1"))
        session.apply(.promotedSubject(graphId: "g-1"))
        await session.settleWrites()
        #expect(calls == 3)
    }

    @Test func applySkipsNavCountsHookForOtherWrites() async {
        let session = makeSession()
        var calls = 0
        session.onNavCountsMoved = { calls += 1 }
        session.apply(.mutatedSourceWorkspace(sourceId: "s1"))
        session.apply(CatalogWriteEffects(sourceScopes: ["s1"], activity: CatalogActivity(revision: 1)))
        session.apply(.createdSource)
        await session.settleWrites()
        #expect(calls == 1)
    }

    /// T-10: a key marked stale while hidden keeps its value but is not settled
    /// until it reloads.
    @Test func staleKeyIsNotSettled() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.isSettled)

        session.apply(.deletedSource(id: "gone"))
        #expect(handle.isStale)
        #expect(!handle.isSettled)
        #expect(handle.status == .ready)
        #expect(!handle.isFetching)

        let again: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForQuery(again)
        #expect(!handle.isStale)
        #expect(handle.isSettled)
    }

    /// T-11: a whole-value write makes a stale key current; a partial patch does not.
    @Test func setQueryValueClearsStaleAndPatchKeepsIt() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForQuery(handle)
        session.apply(.deletedSource(id: "gone"))
        #expect(handle.isStale)

        let result = session.patchCached(key, as: [CatalogSource].self) { rows in
            rows[0].title = "Patched"
            return true
        }
        #expect(result == .applied)
        #expect(handle.isStale)

        session.setQueryValue(key, value: handle.value ?? [])
        #expect(!handle.isStale)
    }

    /// T-18 (a): a key marked stale while its first load is in flight restarts
    /// that load when it is queried, instead of keeping an answer that may be older.
    @Test func staleKeyQueriedDuringLoadReloads() async {
        let store = FakeStore()
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g1", sourceID: "s1")]
        store.listEvidenceGraphsHold.isOn = true
        let session = makeSession(store: store)
        let key = session.evidenceGraphsKey
        let handle: QueryHandle<[CatalogEvidenceGraph]> = session.query(key)
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 1 })

        session.apply(.deletedSource(id: "s9"))
        #expect(handle.isStale)
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("g2", sourceID: "s1")]
        let _: QueryHandle<[CatalogEvidenceGraph]> = session.query(key)
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 2 })
        store.listEvidenceGraphsHold.isOn = false
        store.listEvidenceGraphsHold.release()
        store.listEvidenceGraphsHold.release()
        await session.settleWrites()
        #expect(handle.isSettled)
        #expect(!handle.isStale)
        #expect(handle.value?.map(\.id) == ["g2"])
    }

    /// T-18 (b): a patch that finds the key's first load in flight restarts the load.
    @Test func patchDuringFirstLoadFallsBack() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        session.apply(location: .sectionRoot(.sources))
        let handle: QueryHandle<[CatalogSource]>? = session.queryHandle(
            CatalogQueryKey.sourcesList(project: session.projectKey)
        )
        #expect(handle?.value == nil)

        session.apply(CatalogWriteEffects(
            sourceScopes: ["s1"], activity: CatalogActivity(revision: 5, at: "2026-01-01T00:00:00Z")
        ))
        await session.settleWrites()
        #expect(await waitUntil { store.listSourcesCalls == 2 })
        #expect(handle?.value?.first?.id == "s1")
    }

    /// T-7b: a failed progress re-read reloads the whole map.
    @Test func progressReReadFailureReloadsTheMap() async {
        struct Boom: Error {}
        let store = FakeStore()
        seedStore(store)
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", subjectCount: 1, observationCount: 1
        )
        let session = makeSession(store: store)
        let progressHandle: QueryHandle<[String: EvidenceGraphProgress]> = session.query(
            CatalogQueryKey.evidenceGraphProgress(project: session.projectKey)
        )
        await waitForQuery(progressHandle)
        let listCalls = store.listEvidenceGraphProgressCalls

        store.getEvidenceGraphProgressError = Boom()
        session.apply(CatalogWriteEffects(
            reads: [CatalogWriteEffectsRead(read: .evidenceGraphProgress, ids: ["g-1"])]
        ))
        await session.settleWrites()
        #expect(store.listEvidenceGraphProgressCalls == listCalls + 1)
        #expect(!progressHandle.isFetching)
    }

    // MARK: Graph list re-reads (S14-11)

    private func graphRead(_ sourceIDs: String...) -> CatalogWriteEffects {
        CatalogWriteEffects(reads: [CatalogWriteEffectsRead(read: .evidenceGraphs, ids: sourceIDs)])
    }

    /// A session whose graph list holds `rows` (set without a load) and whose
    /// place shows it when `visible`.
    private func graphListSession(
        store: FakeStore,
        rows: [CatalogEvidenceGraph],
        visible: Bool
    ) -> (WorkspaceSession, QueryHandle<[CatalogEvidenceGraph]>) {
        let session = makeSession(store: store)
        session.setQueryValue(session.evidenceGraphsKey, value: rows)
        if visible {
            session.apply(location: .sourcePage(sourceId: "S"))
        }
        let handle: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        return (session, handle)
    }

    /// T-6: a graph write names one Source, so only that Source's graphs are
    /// asked for and merged; the other Source's rows are untouched.
    @Test func graphListMergesOneSourcesReRead() async {
        let store = FakeStore()
        let (session, handle) = graphListSession(
            store: store,
            rows: [
                evidenceGraph("s-default", sourceID: "S", name: "Default"),
                evidenceGraph("s-jones", sourceID: "S", name: "Jones"),
                evidenceGraph("t-default", sourceID: "T", name: "Default"),
            ],
            visible: true
        )
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("s-default", sourceID: "S", name: "Default"),
            evidenceGraph("s-jones", sourceID: "S", name: "Jones", artifactIDs: ["page2"]),
            evidenceGraph("s-smith", sourceID: "S", name: "Smith"),
            evidenceGraph("t-default", sourceID: "T", name: "Default", artifactIDs: ["other"]),
        ]
        session.apply(graphRead("S"))
        #expect(handle.isFetching)
        await session.settleWrites()
        #expect(!handle.isFetching)
        #expect(store.listEvidenceGraphsSourceIDs == ["S"])
        #expect(session.evidenceGraphs(sourceID: "S")?.map(\.id) == ["s-default", "s-jones", "s-smith"])
        #expect(session.evidenceGraph(id: "s-jones")?.artifactIDs == ["page2"])
        #expect(session.evidenceGraphs(sourceID: "T") == [evidenceGraph("t-default", sourceID: "T", name: "Default")])
    }

    /// T-7 (a): while the re-read waits the list reads as refetching, and a
    /// waiter on `readyValue` gets the merged rows.
    @Test func graphListPartialReReadSetsFetching() async {
        let store = FakeStore()
        let (session, handle) = graphListSession(
            store: store,
            rows: [evidenceGraph("s-default", sourceID: "S", name: "Default")],
            visible: true
        )
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("s-default", sourceID: "S", name: "Default"),
            evidenceGraph("s-new", sourceID: "S", name: "New"),
        ]
        store.listEvidenceGraphsHold.isOn = true
        session.apply(graphRead("S"))
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 1 })
        #expect(handle.isFetching)
        if case .unknown = session.evidenceGraphPresence(id: "s-new") {} else {
            Issue.record("expected .unknown while the re-read runs")
        }
        #expect(session.artifactIDsInNoGraph(sourceID: "S", artifactIDs: ["a"]) == nil)

        let returned = Flag()
        let waiter = Task { @MainActor in
            let rows: [CatalogEvidenceGraph]? = await session.readyValue(session.evidenceGraphsKey)
            returned.isSet = true
            return rows
        }
        await Task.yield()
        #expect(!returned.isSet)
        store.listEvidenceGraphsHold.release()
        let rows = await waiter.value
        #expect(rows?.map(\.id) == ["s-default", "s-new"])
        #expect(!handle.isFetching)
    }

    /// T-7 (b): a failed re-read reloads the whole list through the key's policy.
    @Test func graphListPartialReReadFailureReloadsTheList() async {
        struct Boom: Error {}
        let store = FakeStore()
        let (session, handle) = graphListSession(
            store: store,
            rows: [evidenceGraph("s-default", sourceID: "S", name: "Default")],
            visible: true
        )
        store.listEvidenceGraphsError = Boom()
        session.apply(graphRead("S"))
        await session.settleWrites()
        #expect(store.listEvidenceGraphsSourceIDs.count == 2)
        #expect(store.listEvidenceGraphsSourceIDs.first == .some("S"))
        #expect(store.listEvidenceGraphsSourceIDs.last == .some(nil))
        #expect(handle.status == .error)
        #expect(handle.value?.map(\.id) == ["s-default"])
        #expect(!handle.isFetching)
    }

    /// T-7 (c): a visible list whose first load is in flight restarts that load
    /// instead of re-reading one Source.
    @Test func graphListFirstLoadInFlightRestartsWhenVisible() async {
        let store = FakeStore()
        store.listEvidenceGraphsHold.isOn = true
        let session = makeSession(store: store)
        session.apply(location: .sourcePage(sourceId: "S"))
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 1 })

        session.apply(graphRead("S"))
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 2 })
        store.listEvidenceGraphsHold.isOn = false
        store.listEvidenceGraphsHold.release()
        store.listEvidenceGraphsHold.release()
        await session.settleWrites()
        #expect(store.listEvidenceGraphsSourceIDs == [nil, nil])
    }

    /// T-7 (d): the same with the list hidden: no new call, and the list is
    /// stale once the first load lands.
    @Test func graphListFirstLoadInFlightGoesStaleWhenHidden() async {
        let store = FakeStore()
        store.listEvidenceGraphsHold.isOn = true
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 1 })

        session.apply(graphRead("S"))
        #expect(handle.isStale)
        store.listEvidenceGraphsHold.isOn = false
        store.listEvidenceGraphsHold.release()
        await session.settleWrites()
        #expect(store.listEvidenceGraphsSourceIDs == [nil])
        #expect(handle.isStale)
        #expect(!handle.isSettled)
    }

    /// T-7 (e): overlapping re-reads of two Sources keep `isFetching` true
    /// until the last one ends, and both merge.
    @Test func overlappingGraphReReadsKeepFetchingUntilTheLastEnds() async {
        let store = FakeStore()
        let (session, handle) = graphListSession(
            store: store,
            rows: [
                evidenceGraph("s-default", sourceID: "S", name: "Default"),
                evidenceGraph("t-default", sourceID: "T", name: "Default"),
            ],
            visible: true
        )
        store.evidenceGraphsByProject[projectDir] = [
            evidenceGraph("s-default", sourceID: "S", name: "Default"),
            evidenceGraph("s-new", sourceID: "S", name: "New"),
            evidenceGraph("t-default", sourceID: "T", name: "Default"),
            evidenceGraph("t-new", sourceID: "T", name: "New"),
        ]
        store.listEvidenceGraphsHold.isOn = true
        session.apply(graphRead("S"))
        session.apply(graphRead("T"))
        #expect(await waitUntil { store.listEvidenceGraphsHold.count == 2 })

        store.listEvidenceGraphsHold.release()
        #expect(await waitUntil { handle.value?.count == 3 })
        #expect(handle.isFetching)
        store.listEvidenceGraphsHold.release()
        await session.settleWrites()
        #expect(!handle.isFetching)
        #expect(session.evidenceGraphs(sourceID: "S")?.map(\.id) == ["s-default", "s-new"])
        #expect(session.evidenceGraphs(sourceID: "T")?.map(\.id) == ["t-default", "t-new"])
    }

    /// T-12: a hidden graph list is flagged stale by the read, fetches nothing,
    /// and reloads whole on the next query.
    @Test func hiddenGraphListGoesStaleOnRead() async {
        let store = FakeStore()
        let (session, handle) = graphListSession(
            store: store,
            rows: [evidenceGraph("s-default", sourceID: "S", name: "Default")],
            visible: false
        )
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("s-default", sourceID: "S", name: "Default")]
        session.apply(graphRead("S"))
        #expect(handle.isStale)
        await session.settleWrites()
        #expect(store.listEvidenceGraphsSourceIDs.isEmpty)

        let again: QueryHandle<[CatalogEvidenceGraph]> = session.query(session.evidenceGraphsKey)
        await waitForQuery(again)
        await session.settleWrites()
        #expect(store.listEvidenceGraphsSourceIDs == [nil])
        #expect(!handle.isStale)
    }

    /// T-13: deleting a Source drops its graphs through the same re-read.
    @Test func deletedSourceDropsItsGraphs() async {
        let store = FakeStore()
        let (session, handle) = graphListSession(
            store: store,
            rows: [
                evidenceGraph("s-default", sourceID: "S", name: "Default"),
                evidenceGraph("t-default", sourceID: "T", name: "Default"),
            ],
            visible: true
        )
        store.evidenceGraphsByProject[projectDir] = [evidenceGraph("t-default", sourceID: "T", name: "Default")]
        session.apply(.deletedSource(id: "S"))
        await session.settleWrites()
        #expect(store.listEvidenceGraphsSourceIDs == ["S"])
        #expect(handle.value?.map(\.id) == ["t-default"])
        #expect(session.evidenceGraphs(sourceID: "S")?.isEmpty == true)
    }
}

/// A flag a test task sets once it returns.
@MainActor
private final class Flag {
    var isSet = false
}

/// One-shot gate so concurrent loaders block until tests observe stale data.
@MainActor
private final class Gate {
    private var continuations: [CheckedContinuation<Void, Never>] = []
    private var isOpen = false

    func waitForOpen() async {
        if isOpen { return }
        await withCheckedContinuation { continuation in
            if isOpen {
                continuation.resume()
            } else {
                continuations.append(continuation)
            }
        }
    }

    func open() {
        isOpen = true
        for continuation in continuations {
            continuation.resume()
        }
        continuations.removeAll()
    }
}
