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
        let progressHandle: QueryHandle<[String: SourceGraphProgress]>? = session.queryHandle(
            CatalogQueryKey.sourceGraphProgress(project: project)
        )
        let fieldsHandle: QueryHandle<[CatalogMetadataField]>? = session.queryHandle(
            CatalogQueryKey.metadataFieldsList(project: project)
        )
        #expect(sourcesHandle != nil)
        #expect(typesHandle != nil)
        #expect(progressHandle != nil)
        #expect(fieldsHandle == nil)

        await waitForFetchComplete(sourcesHandle!)
        await waitForFetchComplete(typesHandle!)
        await waitForFetchComplete(progressHandle!)
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
        store.subjectsBySource["s1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "Alice", description: ""),
        ]
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForFetchComplete(handle)
        #expect(handle.value?.isEmpty == true)

        let promoted = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")
        session.apply(.promotedSubject(sourceId: "s1"))
        #expect(!handle.isFetching)
        #expect(handle.value?.isEmpty == true)

        let _: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForFetchComplete(handle)
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
        store.subjectsBySource["s1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
            CatalogSubject(id: "sub-2", ref: "CPR-2", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
        ]
        let first = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")
        let second = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-2")
        let session = makeSession(store: store)
        let firstKey = CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: first.entity.id)
        let secondKey = CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: second.entity.id)
        let list: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForFetchComplete(list)

        session.apply(location: .personDetail(entityId: first.entity.id, ref: first.entity.ref, title: nil))
        let firstHandle: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(firstKey))
        await waitForFetchComplete(firstHandle)
        session.apply(location: .personDetail(entityId: second.entity.id, ref: second.entity.ref, title: nil))
        #expect(session.visibleKeys == [secondKey])
        let secondHandle: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(secondKey))
        await waitForFetchComplete(secondHandle)
        #expect(secondHandle.value?.fields.first?.state == .empty)

        store.observationsBySource["s1"] = [
            CatalogObservation(
                id: "o1", ref: "OBS-1", citationID: "c1", subjectID: "sub-2", propertyID: "p-name",
                polarity: "positive", valueText: "Jim Robins", valueInteger: nil, valueDateID: "",
                valueNameID: "n-1", nameForm: "Jim Robins", valueSubjectID: "", valueTermID: "",
                propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
            ),
        ]
        session.apply(.savedCitation(sourceId: "s1"))

        // Off screen: dropped, so it reloads only when next visited.
        let evicted: QueryHandle<CatalogConclusionDetail>? = session.queryHandle(firstKey)
        #expect(evicted == nil)
        // On screen: revalidated in place.
        await waitForFetchComplete(secondHandle)
        #expect(secondHandle.value?.fields.first?.state == .single)
        // The list is off screen: stale, not reloaded. The next query reloads it.
        #expect(!list.isFetching)
        #expect(list.value?.first?.name?.form != "Jim Robins")
        let _: QueryHandle<[CatalogPersonHeader]> = session.query(.personsList(project: session.projectKey))
        await waitForFetchComplete(list)
        #expect(list.value?.first?.name?.form == "Jim Robins")

        // Leaving for the list hides the second page too.
        session.apply(location: .sectionRoot(.persons))
        session.apply(.savedCitation(sourceId: "s1"))
        let hidden: QueryHandle<CatalogConclusionDetail>? = session.queryHandle(secondKey)
        #expect(hidden == nil)
        await waitForFetchComplete(list)
        #expect(list.status == .ready)

        // Revisiting loads afresh.
        session.apply(location: .personDetail(entityId: first.entity.id, ref: first.entity.ref, title: nil))
        let reloaded: QueryHandle<CatalogConclusionDetail> = try #require(session.queryHandle(firstKey))
        await waitForFetchComplete(reloaded)
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
        session.apply(CatalogWriteEffects(sources: ["s1"], activity: CatalogActivity(revision: 12, at: at)))
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
        await waitForFetchComplete(list)
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
        await waitForFetchComplete(peopleHandle)
        #expect(peopleHandle.value?.isEmpty == true)

        // AC-2.7: People on screen. It reloads now; Events and Places go stale.
        session.apply(location: .sectionRoot(.persons))
        await waitForFetchComplete(peopleHandle)
        (peopleHandle, eventsHandle, placesHandle) = try seed()
        session.apply(CatalogWriteEffects(handles: ["e1"]))
        #expect(peopleHandle.isFetching)
        await waitForFetchComplete(peopleHandle)
        #expect(peopleHandle.value?.isEmpty == true)
        #expect(!eventsHandle.isFetching && !placesHandle.isFetching)
        #expect(eventsHandle.value?.count == 1 && placesHandle.value?.count == 1)
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
        await waitForFetchComplete(fields)
        #expect(fields.value?.properties.first?.usedBy == 2)

        store.propertiesByProject[projectDir]?[0].usedBy = 1
        session.apply(.savedCitation(sourceId: "s1"))

        let _: QueryHandle<PropertiesSnapshot> = session.query(fieldsKey)
        await waitForFetchComplete(fields)
        #expect(fields.value?.properties.first?.usedBy == 1)
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
        await gate.open()
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
        await waitForFetchComplete(handle)
        #expect(handle.value?.first?.id == "s1")
    }

    @Test func applyMutatedSourceGraphPatchesProgressOnly() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 12, observationCount: 48
        )
        store.graphProgressBySource["s2"] = SourceGraphProgress(
            sourceId: "s2", subjectCount: 4, observationCount: 0
        )
        let session = makeSession(store: store)
        let listKey = CatalogQueryKey.sourcesList(project: session.projectKey)
        let progressKey = CatalogQueryKey.sourceGraphProgress(project: session.projectKey)
        let listHandle: QueryHandle<[CatalogSource]> = session.query(listKey)
        let progressHandle: QueryHandle<[String: SourceGraphProgress]> = session.query(progressKey)
        await waitForFetchComplete(listHandle)
        await waitForFetchComplete(progressHandle)
        #expect(progressHandle.value?["s1"]?.subjectCount == 12)
        #expect(progressHandle.value?["s2"]?.subjectCount == 4)
        let listCallsBefore = store.listSourceGraphProgressCalls
        let sourcesAtLoad = listHandle.value

        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 13, observationCount: 48
        )
        session.apply(.mutatedSourceGraph(sourceId: "s1"))
        var waited: UInt64 = 0
        while waited < 2_000_000_000 {
            if progressHandle.value?["s1"]?.subjectCount == 13 { break }
            await Task.yield()
            try? await Task.sleep(nanoseconds: 10_000_000)
            waited += 10_000_000
        }
        #expect(progressHandle.value?["s1"]?.subjectCount == 13)
        #expect(progressHandle.value?["s2"]?.subjectCount == 4)
        #expect(listHandle.value == sourcesAtLoad)
        #expect(listHandle.isFetching == false)
        #expect(store.listSourceGraphProgressCalls == listCallsBefore)
        #expect(store.getSourceGraphProgressCalls >= 1)
    }

    @Test func applySavedCitationDoesNotRefetchSourcesList() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 1, observationCount: 1
        )
        let session = makeSession(store: store)
        let listHandle: QueryHandle<[CatalogSource]> = session.query(
            CatalogQueryKey.sourcesList(project: session.projectKey)
        )
        let progressHandle: QueryHandle<[String: SourceGraphProgress]> = session.query(
            CatalogQueryKey.sourceGraphProgress(project: session.projectKey)
        )
        await waitForFetchComplete(listHandle)
        await waitForFetchComplete(progressHandle)
        let sourcesAtLoad = listHandle.value
        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 1, observationCount: 2
        )
        session.apply(.savedCitation(sourceId: "s1"))
        var waited: UInt64 = 0
        while waited < 2_000_000_000 {
            if progressHandle.value?["s1"]?.observationCount == 2 { break }
            await Task.yield()
            try? await Task.sleep(nanoseconds: 10_000_000)
            waited += 10_000_000
        }
        #expect(progressHandle.value?["s1"]?.observationCount == 2)
        #expect(listHandle.value == sourcesAtLoad)
        #expect(listHandle.isFetching == false)
    }

    @Test func graphCameraIsPerSource() {
        let session = makeSession()
        let first = GraphCanvasCamera(
            magnification: 0.5,
            contentOffset: CGPoint(x: 120, y: 80)
        )
        session.setGraphCamera(first, for: "src-a")
        #expect(session.graphCamera(for: "src-a") == first)
        #expect(session.graphCamera(for: "src-b") == nil)

        let second = GraphCanvasCamera(
            magnification: 1.25,
            contentOffset: CGPoint(x: 10, y: 20)
        )
        session.setGraphCamera(second, for: "src-b")
        #expect(session.graphCamera(for: "src-a") == first)
        #expect(session.graphCamera(for: "src-b") == second)
        #expect(session.graphCamera(for: "src-a") != session.graphCamera(for: "src-b"))
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
