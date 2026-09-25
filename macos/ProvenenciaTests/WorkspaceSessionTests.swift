import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct WorkspaceSessionTests {
    private let projectDir = "/tmp/workspace-session.provenencia"

    private func makeSession(store: FakeStore = FakeStore()) -> WorkspaceSession {
        WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    }

    /// Waits until the handle is not actively refetching and has settled on `.ready` or `.error`.
    private func waitForFetchComplete<Value>(
        _ handle: QueryHandle<Value>,
        timeoutNanoseconds: UInt64 = 2_000_000_000
    ) async {
        var waited: UInt64 = 0
        let step: UInt64 = 10_000_000
        while waited < timeoutNanoseconds {
            if handle.isFetching {
                // stale-while-revalidate still in flight
            } else if handle.status == .ready || handle.status == .error {
                return
            }
            await Task.yield()
            try? await Task.sleep(nanoseconds: step)
            waited += step
        }
    }

    @Test func loadsAndCachesValue() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()

        let handle = session.ensureQuery(key) {
            counter.count += 1
            return ["alpha"]
        }
        await waitForFetchComplete(handle)
        #expect(handle.status == .ready)
        #expect(handle.value == ["alpha"])
        #expect(counter.count == 1)

        _ = session.ensureQuery(key) {
            counter.count += 1
            return ["beta"]
        }
        await Task.yield()
        #expect(counter.count == 1)
        #expect(handle.value == ["alpha"])
    }

    @Test func dedupesConcurrentLoads() async {
        let session = makeSession()
        let key = CatalogQueryKey.metadataFieldsList(project: session.projectKey)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()
        let gate = Gate()

        await withTaskGroup(of: Void.self) { group in
            group.addTask { @MainActor in
                _ = session.ensureQuery(key) {
                    counter.count += 1
                    await gate.waitForOpen()
                    return ["shared"]
                }
            }
            group.addTask { @MainActor in
                _ = session.ensureQuery(key) {
                    counter.count += 1
                    await gate.waitForOpen()
                    return ["shared"]
                }
            }
            await gate.open()
            await group.waitForAll()
        }

        let handle: QueryHandle<[String]> = session.queryHandle(key)!
        await waitForFetchComplete(handle)
        #expect(counter.count == 1)
        #expect(handle.status == .ready)
        #expect(handle.value == ["shared"])
    }

    @Test func invalidateTriggersReload() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()

        let handle = session.ensureQuery(key) {
            counter.count += 1
            return ["first"]
        }
        await waitForFetchComplete(handle)
        #expect(counter.count == 1)

        session.invalidate(key)
        _ = session.ensureQuery(key) {
            counter.count += 1
            return ["second"]
        }
        await waitForFetchComplete(handle)
        #expect(counter.count == 2)
        #expect(handle.value == ["second"])
    }

    @Test func invalidateAllMatching() async {
        let session = makeSession()
        let otherProject = ProjectKey(projectDir: "/tmp/other.provenencia")
        let keyA = CatalogQueryKey.sourcesList(project: session.projectKey)
        let keyB = CatalogQueryKey.sourceTypesList(project: session.projectKey)
        let keyOther = CatalogQueryKey.sourcesList(project: otherProject)
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()

        let handleA = session.ensureQuery(keyA) {
            counter.count += 1
            return ["a"]
        }
        let handleB = session.ensureQuery(keyB) {
            counter.count += 1
            return ["b"]
        }
        let handleOther = session.ensureQuery(keyOther) {
            counter.count += 1
            return ["other"]
        }
        await waitForFetchComplete(handleA)
        await waitForFetchComplete(handleB)
        await waitForFetchComplete(handleOther)
        #expect(counter.count == 3)

        session.invalidateAll { $0.project == session.projectKey }

        _ = session.ensureQuery(keyA) {
            counter.count += 1
            return ["a2"]
        }
        _ = session.ensureQuery(keyB) {
            counter.count += 1
            return ["b2"]
        }
        _ = session.ensureQuery(keyOther) {
            counter.count += 1
            return ["other2"]
        }
        await waitForFetchComplete(handleA)
        await waitForFetchComplete(handleB)
        await waitForFetchComplete(handleOther)
        #expect(counter.count == 5)
        #expect(handleOther.value == ["other"])
    }

    @Test func staleWhileRevalidate() async {
        let session = makeSession()
        let key = CatalogQueryKey.typeSuggestions(project: session.projectKey, typeId: "t1")
        final class Counter: @unchecked Sendable {
            var count = 0
        }
        let counter = Counter()
        let gate = Gate()

        let handle = session.ensureQuery(key) {
            counter.count += 1
            return ["first"]
        }
        await waitForFetchComplete(handle)

        session.invalidate(key)
        _ = session.ensureQuery(key) {
            counter.count += 1
            await gate.waitForOpen()
            return ["second"]
        }

        #expect(handle.value == ["first"])
        #expect(handle.isFetching)
        await gate.open()
        await waitForFetchComplete(handle)
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
        await waitForFetchComplete(handle)
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
        let fieldsHandle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(
            CatalogQueryKey.metadataFieldsList(project: project)
        )
        #expect(sourcesHandle != nil)
        #expect(typesHandle != nil)
        #expect(fieldsHandle == nil)

        await waitForFetchComplete(sourcesHandle!)
        await waitForFetchComplete(typesHandle!)
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
        await waitForFetchComplete(workspaceHandle!)
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

        await waitForFetchComplete(typesHandle!)
        await waitForFetchComplete(fieldsHandle!)
        await waitForFetchComplete(suggestionsHandle!)
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

        await waitForFetchComplete(handle!)
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
        await waitForFetchComplete(workspaceHandle!)

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
        await waitForFetchComplete(handle)
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
        store.subjectsBySource["s1"] = [
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
        let key = CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: "s1")
        let handle: QueryHandle<SourceGraphRows> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.observations.isEmpty == true)

        store.observationsBySource["s1"] = [
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
        session.apply(.savedCitation(sourceId: "s1"))
        await waitForFetchComplete(handle)
        #expect(handle.value?.observations.count == 1)
    }

    @Test func applySavedCitationDoesNotReloadSubjectFields() async {
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
        let session = makeSession(store: store)
        let fieldsKey = CatalogQueryKey.subjectFieldsWorkspace(project: session.projectKey)
        let fields: QueryHandle<SubjectFieldsSnapshot> = session.query(fieldsKey)
        await waitForFetchComplete(fields)
        let typesAtLoad = fields.value?.types
        store.subjectTypesByProject[projectDir] = []
        session.apply(.savedCitation(sourceId: "s1"))
        await Task.yield()
        #expect(fields.value?.types == typesAtLoad)
        #expect(fields.isFetching == false)
    }

    @Test func readyValueIgnoresStaleCache() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let handle = session.ensureQuery(key) { ["alpha"] }
        await waitForFetchComplete(handle)
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
        await waitForFetchComplete(handle)
        #expect(handle.value == ["new"])
        await olderGate.open()
        await Task.yield()
        #expect(handle.value == ["new"])
        #expect(handle.isFetching == false)
    }

    @Test func setQueryValueWinsOverInFlightLoad() async {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let gate = Gate()
        let handle = session.ensureQuery(key) {
            await gate.waitForOpen()
            return ["stale"]
        }
        session.setQueryValue(key, value: ["patched"])
        #expect(handle.value == ["patched"])
        await gate.open()
        await Task.yield()
        #expect(handle.value == ["patched"])
        let ready: [String]? = await session.readyValue(key)
        #expect(ready == ["patched"])
    }
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
