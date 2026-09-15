import Foundation
import Testing
@testable import Provenencia

@Suite(.serialized)
@MainActor
struct CatalogQueryRegistryTests {
    private let projectDir = "/tmp/catalog-query-registry.provenencia"

    private func makeSession(store: FakeStore = FakeStore()) -> WorkspaceSession {
        WorkspaceSession(projectKey: ProjectKey(projectDir: projectDir), store: store)
    }

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

    @Test func registryLoadsEveryKey() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey

        let sourcesHandle: QueryHandle<[CatalogSource]> = session.query(
            CatalogQueryKey.sourcesList(project: project)
        )
        let typesHandle: QueryHandle<[CatalogSourceType]> = session.query(
            CatalogQueryKey.sourceTypesList(project: project)
        )
        let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.query(
            CatalogQueryKey.metadataFieldsList(project: project)
        )
        let workspaceHandle: QueryHandle<CatalogSourceWorkspace> = session.query(
            CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        )
        let suggestionsHandle: QueryHandle<[CatalogTypeSuggestion]> = session.query(
            CatalogQueryKey.typeSuggestions(project: project, typeId: "t1")
        )

        await waitForFetchComplete(sourcesHandle)
        await waitForFetchComplete(typesHandle)
        await waitForFetchComplete(fieldsHandle)
        await waitForFetchComplete(workspaceHandle)
        await waitForFetchComplete(suggestionsHandle)

        #expect(sourcesHandle.value?.first?.title == "Alpha")
        #expect(typesHandle.value?.first?.label == "Book")
        #expect(fieldsHandle.value?.first?.label == "Author")
        #expect(workspaceHandle.value?.source.id == "s1")
        #expect(suggestionsHandle.value?.count == 1)
    }

    @Test func registryBackedQueryDoesNotRecallLoader() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)

        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForFetchComplete(handle)

        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s2", ref: "SRC-2", sourceTypeID: "", title: "Beta", description: ""),
        ]

        let _: QueryHandle<[CatalogSource]> = session.query(key)
        await Task.yield()
        #expect(handle.value?.first?.title == "Alpha")
    }

    @Test func setQueryValueSkipsLoader() async {
        let store = FakeStore()
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        let seeded = [
            CatalogSource(id: "seed", ref: "SRC-SEED", sourceTypeID: "", title: "Seeded", description: ""),
        ]

        session.setQueryValue(key, value: seeded)
        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await Task.yield()
        #expect(handle.status == .ready)
        #expect(handle.value == seeded)
        #expect(store.heldCatalogProjectDir == nil)
    }

    @Test func applyUpdatedSourcePatchesListAndWorkspace() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let listKey = CatalogQueryKey.sourcesList(project: project)
        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")

        let listHandle: QueryHandle<[CatalogSource]> = session.query(listKey)
        let workspaceHandle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        await waitForFetchComplete(listHandle)
        await waitForFetchComplete(workspaceHandle)
        store.heldCatalogProjectDir = nil

        let updated = CatalogSource(
            id: "s1", ref: "SRC-1", sourceTypeID: "t1", title: "Renamed", description: "New desc"
        )
        session.apply(.updatedSource(updated))

        #expect(listHandle.value?.first?.title == "Renamed")
        #expect(workspaceHandle.value?.source.title == "Renamed")
        #expect(store.heldCatalogProjectDir == nil)
    }

    @Test func applyCreatedSourceInvalidatesList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)

        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForFetchComplete(handle)

        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "", title: "Alpha", description: ""),
            CatalogSource(id: "s2", ref: "SRC-2", sourceTypeID: "", title: "New Row", description: ""),
        ]

        session.apply(.createdSource)
        let _: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.count == 2)
        #expect(handle.value?.last?.title == "New Row")
    }

    @Test func applyFieldCRUDInvalidatesMetadataList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let fieldsKey = CatalogQueryKey.metadataFieldsList(project: project)
        let typesKey = CatalogQueryKey.sourceTypesList(project: project)

        let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.query(fieldsKey)
        let typesHandle: QueryHandle<[CatalogSourceType]> = session.query(typesKey)
        await waitForFetchComplete(fieldsHandle)
        await waitForFetchComplete(typesHandle)

        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f2", key: "date", origin: "user", label: "Date", dataType: "date", description: ""
            ),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t9", key: "map", origin: "user", label: "Map", description: ""),
        ]

        for mutation in [
            CatalogMutation.createdMetadataField,
            .updatedMetadataField(id: "f1"),
            .deletedMetadataField(id: "f1"),
        ] {
            session.apply(mutation)
            let _: QueryHandle<[CatalogMetadataField]> = session.query(fieldsKey)
            await waitForFetchComplete(fieldsHandle)
            #expect(fieldsHandle.value?.first?.label == "Date")

            let _: QueryHandle<[CatalogSourceType]> = session.query(typesKey)
            await Task.yield()
            #expect(typesHandle.value?.first?.label == "Book")
        }
    }

    @Test func applyTypeCRUDInvalidatesTypesList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let typesKey = CatalogQueryKey.sourceTypesList(project: project)
        let fieldsKey = CatalogQueryKey.metadataFieldsList(project: project)

        let typesHandle: QueryHandle<[CatalogSourceType]> = session.query(typesKey)
        let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.query(fieldsKey)
        await waitForFetchComplete(typesHandle)
        await waitForFetchComplete(fieldsHandle)

        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t2", key: "photo", origin: "user", label: "Photo", description: ""),
        ]
        store.fieldsByProject[projectDir] = [
            CatalogMetadataField(
                id: "f9", key: "place", origin: "user", label: "Place", dataType: "text", description: ""
            ),
        ]

        for mutation in [
            CatalogMutation.createdSourceType,
            .updatedSourceType(id: "t1"),
            .deletedSourceType(id: "t1"),
        ] {
            session.apply(mutation)
            let _: QueryHandle<[CatalogSourceType]> = session.query(typesKey)
            await waitForFetchComplete(typesHandle)
            #expect(typesHandle.value?.first?.label == "Photo")

            let _: QueryHandle<[CatalogMetadataField]> = session.query(fieldsKey)
            await Task.yield()
            #expect(fieldsHandle.value?.first?.label == "Author")
        }
    }

    @Test func applySuggestionMutationsInvalidateTypeKey() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let suggestionsKey = CatalogQueryKey.typeSuggestions(project: project, typeId: "t1")
        let otherSuggestionsKey = CatalogQueryKey.typeSuggestions(project: project, typeId: "t2")
        let typesKey = CatalogQueryKey.sourceTypesList(project: project)

        store.suggestionsByType["t2"] = []

        let suggestionsHandle: QueryHandle<[CatalogTypeSuggestion]> = session.query(suggestionsKey)
        let otherHandle: QueryHandle<[CatalogTypeSuggestion]> = session.query(otherSuggestionsKey)
        let typesHandle: QueryHandle<[CatalogSourceType]> = session.query(typesKey)
        await waitForFetchComplete(suggestionsHandle)
        await waitForFetchComplete(otherHandle)
        await waitForFetchComplete(typesHandle)

        store.suggestionsByType["t1"] = []
        store.suggestionsByType["t2"] = [
            CatalogTypeSuggestion(
                field: CatalogMetadataField(
                    id: "f2", key: "date", origin: "user", label: "Date", dataType: "date", description: ""
                ),
                sortOrder: 0
            ),
        ]
        store.sourceTypesByProject[projectDir] = [
            CatalogSourceType(id: "t9", key: "map", origin: "user", label: "Map", description: ""),
        ]

        session.apply(.assignedTypeSuggestion(typeId: "t1"))
        let _: QueryHandle<[CatalogTypeSuggestion]> = session.query(suggestionsKey)
        await waitForFetchComplete(suggestionsHandle)
        #expect(suggestionsHandle.value?.isEmpty == true)

        let _: QueryHandle<[CatalogTypeSuggestion]> = session.query(otherSuggestionsKey)
        await Task.yield()
        #expect(otherHandle.value?.isEmpty == true)

        let _: QueryHandle<[CatalogSourceType]> = session.query(typesKey)
        await Task.yield()
        #expect(typesHandle.value?.first?.label == "Book")
    }

    @Test func sessionUsesStandardRegistryByDefault() {
        let session = makeSession()
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)
        #expect(session.registry.stalePolicy(for: key) == .sessionFresh)
    }
}
