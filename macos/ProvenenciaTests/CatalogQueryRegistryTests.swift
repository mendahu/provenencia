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

    @Test func citationCountsLoadBySource() async {
        let store = FakeStore()
        seedStore(store)
        store.artifactsBySource["s1"] = [
            CatalogArtifact(
                id: "art-0",
                ref: "ART-0",
                sourceID: "s1",
                fileID: "file-0",
                label: "Scan 1",
                description: ""
            ),
            CatalogArtifact(
                id: "art-1",
                ref: "ART-1",
                sourceID: "s1",
                fileID: "file-1",
                label: "Scan 2",
                description: ""
            ),
        ]
        store.citationsByID["cit-1"] = CatalogCitation(
            id: "cit-1",
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: "{}",
            transcription: "",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        store.citationsByID["cit-2"] = CatalogCitation(
            id: "cit-2",
            ref: "CIT-2",
            artifactID: "art-0",
            locatorJSON: "{}",
            transcription: "",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        let session = makeSession(store: store)
        let handle: QueryHandle<[String: Int]> = session.query(
            CatalogQueryKey.citationCounts(project: session.projectKey, sourceId: "s1")
        )
        await waitForFetchComplete(handle)
        #expect(handle.value?["art-0"] == 2)
        #expect(handle.value?["art-1"] == nil)
    }

    @Test func sourceGraphProgressLoadsMap() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressBySource["s1"] = SourceGraphProgress(
            sourceId: "s1", subjectCount: 12, observationCount: 48
        )
        let session = makeSession(store: store)
        let handle: QueryHandle<[String: SourceGraphProgress]> = session.query(
            CatalogQueryKey.sourceGraphProgress(project: session.projectKey)
        )
        await waitForFetchComplete(handle)
        #expect(handle.value?["s1"]?.subjectCount == 12)
        #expect(handle.value?["s1"]?.observationCount == 48)
    }

    @Test func connectRulesLoadFromStore() async {
        let store = FakeStore()
        store.connectRules = CatalogConnectRule.productMatrix
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogConnectRule]> = session.query(
            CatalogQueryKey.connectRules(project: session.projectKey)
        )
        await waitForFetchComplete(handle)
        #expect(handle.value?.isEmpty == false)
    }

    @Test func applyCreatedPropertyTermReloadsTerms() async {
        let store = FakeStore()
        store.propertyTermsByProperty["p1"] = [
            CatalogPropertyTerm(
                id: "t1",
                propertyID: "p1",
                key: "witness",
                origin: "provenencia",
                label: "Witness",
                description: ""
            ),
        ]
        let session = makeSession(store: store)
        let key = CatalogQueryKey.propertyTerms(project: session.projectKey, propertyId: "p1")
        let handle: QueryHandle<[CatalogPropertyTerm]> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.count == 1)

        store.propertyTermsByProperty["p1"]?.append(
            CatalogPropertyTerm(
                id: "t2",
                propertyID: "p1",
                key: "neighbor",
                origin: "user",
                label: "Neighbor",
                description: ""
            )
        )
        session.apply(.createdPropertyTerm(propertyId: "p1"))
        await waitForFetchComplete(handle)
        #expect(handle.value?.count == 2)
    }

    @Test func citationsByArtifactLoadsListedRows() async {
        let store = FakeStore()
        store.citationsByID["cit-1"] = CatalogCitation(
            id: "cit-1",
            ref: "CIT-1",
            artifactID: "art-0",
            locatorJSON: "{}",
            transcription: "",
            description: "",
            transcriptionUncertain: false,
            transcriptionNote: ""
        )
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogListedCitation]> = session.query(
            CatalogQueryKey.citationsByArtifact(project: session.projectKey, artifactId: "art-0")
        )
        await waitForFetchComplete(handle)
        #expect(handle.value?.map(\.id) == ["cit-1"])
    }

    @Test func sourceGraphLoadsPlacedPrimaries() async {
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
            CatalogSubjectType(
                id: "type-location",
                key: "location",
                origin: "provenencia",
                label: "Location",
                description: "",
                refPrefix: "LOC",
                candidateRefPrefix: "CLO"
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
            CatalogSubject(
                id: "sub-bridge",
                ref: "CLO-1",
                sourceID: "s1",
                subjectTypeID: "type-location",
                label: "Somewhere",
                description: ""
            ),
        ]
        store.subjectPositionsBySubject["sub-1"] = CatalogSubjectPosition(
            subjectID: "sub-1",
            gridX: 2,
            gridY: 3
        )
        store.subjectPositionsBySubject["sub-bridge"] = CatalogSubjectPosition(
            subjectID: "sub-bridge",
            gridX: 4,
            gridY: 5
        )

        let session = makeSession(store: store)
        let handle: QueryHandle<SourceGraphRows> = session.query(
            CatalogQueryKey.sourceGraph(project: session.projectKey, sourceId: "s1")
        )
        await waitForFetchComplete(handle)

        #expect(handle.value?.sourceId == "s1")
        #expect(handle.value?.subjects.count == 2)
        #expect(handle.value?.subjects.map(\.id).sorted() == ["sub-1", "sub-bridge"])
        #expect(handle.value?.positions.first { $0.subjectID == "sub-1" }?.gridX == 2)
        let snapshot = SourceGraphSnapshot.build(
            rows: handle.value ?? SourceGraphRows(sourceId: "s1"),
            types: store.subjectTypesByProject[projectDir] ?? [],
            rules: CatalogConnectRule.productMatrix
        )
        #expect(snapshot.subjects.count == 1)
        #expect(snapshot.subjects.first?.kind == .person)
        #expect(snapshot.bridges.count == 1)
        #expect(snapshot.bridges.first?.kind == .location)
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
        // FakeStore listSources is newest-created-first (id DESC), matching Go.
        #expect(handle.value?.map(\.title) == ["New Row", "Alpha"])
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

    @Test func invalidationsComeFromRegistrySpecs() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        let source = CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "t2", title: "T", description: "")

        #expect(registry.invalidations(by: .updatedSource(source), project: project).isEmpty)

        // A new source moves a type's `usedBy`.
        #expect(registry.invalidations(by: .createdSource, project: project) == [
            .key(.sourcesList(project: project)),
            .key(.sourceTypesList(project: project)),
        ])
        // Adding vocabulary touches only the list that owns it — the Source
        // page reads that list rather than carrying its own copy.
        #expect(registry.invalidations(by: .createdMetadataField, project: project) == [
            .key(.metadataFieldsList(project: project)),
        ])
        #expect(registry.invalidations(by: .createdSourceType, project: project) == [
            .key(.sourceTypesList(project: project)),
        ])
        // Edits that restate a *derived* row still fan out, since the mutation
        // names no single page: metadata rows embed their field, and the row set
        // comes from the type's suggestions.
        #expect(registry.invalidations(by: .deletedMetadataField(id: "f1"), project: project) == [
            .key(.metadataFieldsList(project: project)),
            .allCached(.sourceWorkspace),
            .allCached(.typeSuggestions),
        ])
        #expect(registry.invalidations(by: .assignedTypeSuggestion(typeId: "t1"), project: project) == [
            .allCached(.sourceWorkspace),
            .key(.typeSuggestions(project: project, typeId: "t1")),
        ])
        #expect(registry.invalidations(by: .mutatedSourceGraph(sourceId: "s1"), project: project) == [
            .key(.sourceGraph(project: project, sourceId: "s1")),
        ])
        #expect(
            !registry.invalidations(by: .mutatedSourceGraph(sourceId: "s1"), project: project).contains(
                .key(.sourceGraphProgress(project: project))
            )
        )
        #expect(
            !registry.invalidations(by: .savedCitation(sourceId: "s1"), project: project).contains(
                .key(.sourceGraphProgress(project: project))
            )
        )
        #expect(
            !registry.invalidations(by: .mutatedSourceGraph(sourceId: "s1"), project: project).contains(
                .key(.sourcesList(project: project))
            )
        )
        #expect(registry.invalidations(by: .savedCitation(sourceId: "s1"), project: project) == [
            .key(.sourceGraph(project: project, sourceId: "s1")),
            .key(.citationCounts(project: project, sourceId: "s1")),
            .allCached(.citationsByArtifact),
        ])
        #expect(registry.invalidations(by: .createdPropertyTerm(propertyId: "p1"), project: project) == [
            .key(.propertyTerms(project: project, propertyId: "p1")),
        ])
        // Grades are seeded vocabulary with no CRUD surface, so nothing stales them.
        let everyMutation: [CatalogMutation] = [
            .updatedSource(source), .changedSourceType(source), .createdSource,
            .createdSourceType, .updatedSourceType(id: "t1"), .deletedSourceType(id: "t1"),
            .createdMetadataField, .updatedMetadataField(id: "f1"), .deletedMetadataField(id: "f1"),
            .assignedTypeSuggestion(typeId: "t1"), .removedTypeSuggestion(typeId: "t1"),
            .mutatedSourceWorkspace(sourceId: "s1"), .mutatedSourceMetadata(sourceId: "s1"),
            .mutatedSourceGraph(sourceId: "s1"),
            .savedCitation(sourceId: "s1"),
            .createdPropertyTerm(propertyId: "p1"),
        ]
        #expect(everyMutation.allSatisfy { mutation in
            !registry.invalidations(by: mutation, project: project).contains(
                .key(.credibilityGradesList(project: project))
            )
        })
        // Writes that name their source stay narrow.
        #expect(registry.invalidations(by: .mutatedSourceWorkspace(sourceId: "s1"), project: project) == [
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
        ])
        #expect(registry.invalidations(by: .mutatedSourceMetadata(sourceId: "s1"), project: project) == [
            .key(.metadataFieldsList(project: project)),
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
        ])
        #expect(registry.invalidations(by: .changedSourceType(source), project: project) == [
            .key(.sourceTypesList(project: project)),
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
        ])
    }

    /// Adding vocabulary no longer disturbs open Source pages: the page reads the
    /// shared lists, so there is nothing stale on it to refetch.
    @Test func addingVocabularyLeavesCachedSourcePagesAlone() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let project = session.projectKey
        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        let fieldsKey = CatalogQueryKey.metadataFieldsList(project: project)

        let workspaceHandle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        let fieldsHandle: QueryHandle<[CatalogMetadataField]> = session.query(fieldsKey)
        await waitForFetchComplete(workspaceHandle)
        await waitForFetchComplete(fieldsHandle)
        store.heldCatalogProjectDir = nil

        store.fieldsByProject[projectDir]?.append(
            CatalogMetadataField(
                id: "f2", key: "folio", origin: "user", label: "Folio", dataType: "text", description: ""
            )
        )
        session.apply(.createdMetadataField)
        session.apply(.createdSourceType)
        await waitForFetchComplete(fieldsHandle)
        #expect(fieldsHandle.value?.map(\.label) == ["Author", "Folio"])
        #expect(workspaceHandle.isFetching == false)

        // The page is untouched — re-querying it opens no catalog.
        store.heldCatalogProjectDir = nil
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        await Task.yield()
        #expect(store.heldCatalogProjectDir == nil)
    }

    /// Edits that restate a derived metadata row still reach every cached page,
    /// because the mutation cannot name which sources hold that field or type.
    @Test func derivedMetadataEditsStillInvalidateEveryCachedPage() async {
        let store = FakeStore()
        seedStore(store)
        store.sourcesByProject[projectDir]?.append(
            CatalogSource(id: "s2", ref: "SRC-2", sourceTypeID: "t1", title: "Beta", description: "")
        )
        store.metadataBySource["s1"] = [suggestedAuthorEntry(label: "Author")]
        store.metadataBySource["s2"] = [suggestedAuthorEntry(label: "Author")]
        let session = makeSession(store: store)
        let project = session.projectKey
        let firstKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")
        let secondKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s2")

        let firstHandle: QueryHandle<CatalogSourceWorkspace> = session.query(firstKey)
        let secondHandle: QueryHandle<CatalogSourceWorkspace> = session.query(secondKey)
        await waitForFetchComplete(firstHandle)
        await waitForFetchComplete(secondHandle)

        store.metadataBySource["s1"] = [suggestedAuthorEntry(label: "Author name")]
        store.metadataBySource["s2"] = [suggestedAuthorEntry(label: "Author name")]
        session.apply(.updatedMetadataField(id: "f1"))

        let _: QueryHandle<CatalogSourceWorkspace> = session.query(firstKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(secondKey)
        await waitForFetchComplete(firstHandle)
        await waitForFetchComplete(secondHandle)
        #expect(firstHandle.value?.metadata.first?.field.label == "Author name")
        #expect(secondHandle.value?.metadata.first?.field.label == "Author name")
    }

    private func suggestedAuthorEntry(label: String) -> CatalogMetadataEntry {
        CatalogMetadataEntry(
            field: CatalogMetadataField(
                id: "f1", key: "author", origin: "user", label: label,
                dataType: "text", description: ""
            ),
            valueText: "", hasValue: false, suggested: true, sortOrder: 0
        )
    }

    /// Suggestions embed a whole field row, so a relabel reaches types the edit
    /// never named.
    @Test func fieldRelabelInvalidatesEveryCachedSuggestionList() async {
        let store = FakeStore()
        seedStore(store)
        store.suggestionsByType["t2"] = store.suggestionsByType["t1"]
        let session = makeSession(store: store)
        let project = session.projectKey
        let firstKey = CatalogQueryKey.typeSuggestions(project: project, typeId: "t1")
        let secondKey = CatalogQueryKey.typeSuggestions(project: project, typeId: "t2")

        let firstHandle: QueryHandle<[CatalogTypeSuggestion]> = session.query(firstKey)
        let secondHandle: QueryHandle<[CatalogTypeSuggestion]> = session.query(secondKey)
        await waitForFetchComplete(firstHandle)
        await waitForFetchComplete(secondHandle)

        let renamed = CatalogTypeSuggestion(
            field: CatalogMetadataField(
                id: "f1", key: "author", origin: "user", label: "Author name",
                dataType: "text", description: ""
            ),
            sortOrder: 0
        )
        store.suggestionsByType["t1"] = [renamed]
        store.suggestionsByType["t2"] = [renamed]
        session.apply(.updatedMetadataField(id: "f1"))

        let _: QueryHandle<[CatalogTypeSuggestion]> = session.query(firstKey)
        let _: QueryHandle<[CatalogTypeSuggestion]> = session.query(secondKey)
        await waitForFetchComplete(firstHandle)
        await waitForFetchComplete(secondHandle)
        #expect(firstHandle.value?.first?.field.label == "Author name")
        #expect(secondHandle.value?.first?.field.label == "Author name")
    }

    /// `usedBy` gates Delete on the Source Fields page, and it moves whenever a
    /// source gains or loses a value.
    @Test func sourceMetadataWriteInvalidatesFieldsList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.metadataFieldsList(project: session.projectKey)

        let handle: QueryHandle<[CatalogMetadataField]> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.first?.usedBy == 0)

        store.fieldsByProject[projectDir]?[0].usedBy = 1
        session.apply(.mutatedSourceMetadata(sourceId: "s1"))

        let _: QueryHandle<[CatalogMetadataField]> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.first?.usedBy == 1)
    }

    /// Adding a source moves its type's `usedBy`, which gates deleting the type.
    @Test func createdSourceInvalidatesTypesList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourceTypesList(project: session.projectKey)

        let handle: QueryHandle<[CatalogSourceType]> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.first?.usedBy == 0)

        store.sourceTypesByProject[projectDir]?[0].usedBy = 1
        session.apply(.createdSource)

        let _: QueryHandle<[CatalogSourceType]> = session.query(key)
        await waitForFetchComplete(handle)
        #expect(handle.value?.first?.usedBy == 1)
    }

    /// Moving a source to another type patches the list row for an instant
    /// repaint, then refetches the page because the engine derives suggested
    /// metadata rows from the type.
    @Test func changedSourceTypePatchesRowThenRefetchesPage() async {
        let store = FakeStore()
        seedStore(store)
        store.sourceTypesByProject[projectDir]?.append(
            CatalogSourceType(id: "t2", key: "photo", origin: "user", label: "Photo", description: "")
        )
        let session = makeSession(store: store)
        let project = session.projectKey
        let listKey = CatalogQueryKey.sourcesList(project: project)
        let workspaceKey = CatalogQueryKey.sourceWorkspace(project: project, sourceId: "s1")

        let listHandle: QueryHandle<[CatalogSource]> = session.query(listKey)
        let workspaceHandle: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        await waitForFetchComplete(listHandle)
        await waitForFetchComplete(workspaceHandle)

        let moved = CatalogSource(
            id: "s1", ref: "SRC-1", sourceTypeID: "t2", title: "Alpha", description: ""
        )
        store.sourcesByProject[projectDir] = [moved]
        store.metadataBySource["s1"] = [
            CatalogMetadataEntry(
                field: CatalogMetadataField(
                    id: "f1", key: "author", origin: "user", label: "Author",
                    dataType: "text", description: ""
                ),
                valueText: "", hasValue: false, suggested: true, sortOrder: 0
            ),
        ]
        session.apply(.changedSourceType(moved))

        // Patched synchronously — no refetch needed for the row itself.
        #expect(listHandle.value?.first?.sourceTypeID == "t2")

        let _: QueryHandle<CatalogSourceWorkspace> = session.query(workspaceKey)
        await waitForFetchComplete(workspaceHandle)
        #expect(workspaceHandle.value?.source.sourceTypeID == "t2")
        #expect(workspaceHandle.value?.metadata.count == 1)
    }
}
