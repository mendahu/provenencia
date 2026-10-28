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
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(
                id: "g1", ref: "GPH-1", sourceID: "s1", name: "Default", description: "", artifactIDs: ["art-0"]
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

        await waitForQuery(sourcesHandle)
        await waitForQuery(typesHandle)
        await waitForQuery(fieldsHandle)
        await waitForQuery(workspaceHandle)
        await waitForQuery(suggestionsHandle)

        #expect(sourcesHandle.value?.first?.title == "Alpha")
        #expect(typesHandle.value?.first?.label == "Book")
        #expect(fieldsHandle.value?.first?.label == "Author")
        #expect(workspaceHandle.value?.source.id == "s1")
        #expect(suggestionsHandle.value?.count == 1)
    }

    @Test func confidenceGradesLoadInOrder() async {
        let store = FakeStore()
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogClaimConfidenceGrade]> = session.query(
            CatalogQueryKey.confidenceGradesList(project: session.projectKey)
        )
        await waitForQuery(handle)
        #expect(handle.value?.map(\.key) == ["low_confidence", "moderate", "high_confidence"])
    }

    @Test func citationCountsLoadByEvidenceGraph() async {
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
            CatalogQueryKey.citationCounts(project: session.projectKey, graphId: "g1")
        )
        await waitForQuery(handle)
        #expect(handle.value?["art-0"] == 2)
        #expect(handle.value?["art-1"] == nil)
    }

    @Test func evidenceGraphProgressLoadsMap() async {
        let store = FakeStore()
        seedStore(store)
        store.graphProgressByGraph["g-1"] = EvidenceGraphProgress(
            graphId: "g-1", sourceId: "s1", subjectCount: 12, observationCount: 48
        )
        let session = makeSession(store: store)
        let handle: QueryHandle<[String: EvidenceGraphProgress]> = session.query(
            CatalogQueryKey.evidenceGraphProgress(project: session.projectKey)
        )
        await waitForQuery(handle)
        #expect(handle.value?["g-1"]?.subjectCount == 12)
        #expect(handle.value?["g-1"]?.observationCount == 48)
        #expect(handle.value?["s1"] == nil)
    }

    @Test func connectRulesLoadFromStore() async {
        let store = FakeStore()
        store.connectRules = CatalogConnectRule.productMatrix
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogConnectRule]> = session.query(
            CatalogQueryKey.connectRules(project: session.projectKey)
        )
        await waitForQuery(handle)
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
        await waitForQuery(handle)
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
        await waitForQuery(handle)
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
            CatalogQueryKey.citationsByArtifact(project: session.projectKey, graphId: "g1", artifactId: "art-0")
        )
        await waitForQuery(handle)
        #expect(handle.value?.map(\.id) == ["cit-1"])
    }

    @Test func evidenceGraphLoadsPlacedPrimaries() async {
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
        store.subjectsByGraph["g-1"] = [
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
        let handle: QueryHandle<EvidenceGraphRows> = session.query(
            CatalogQueryKey.evidenceGraph(project: session.projectKey, graphId: "g-1")
        )
        await waitForQuery(handle)

        #expect(handle.value?.graphId == "g-1")
        #expect(handle.value?.subjects.count == 2)
        #expect(handle.value?.subjects.map(\.id).sorted() == ["sub-1", "sub-bridge"])
        #expect(handle.value?.positions.first { $0.subjectID == "sub-1" }?.gridX == 2)
        let snapshot = EvidenceGraphSnapshot.build(
            rows: handle.value ?? EvidenceGraphRows(graphId: "g-1"),
            types: store.subjectTypesByProject[projectDir] ?? [],
            rules: CatalogConnectRule.productMatrix
        )
        #expect(snapshot.subjects.count == 1)
        #expect(snapshot.subjects.first?.kind == .person)
        #expect(snapshot.bridges.count == 1)
        #expect(snapshot.bridges.first?.kind == .location)
        #expect(snapshot.subjects.first?.membership == nil)
    }

    @Test func evidenceGraphLoadsMembershipOfPromotedSubjects() async throws {
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
            CatalogSubject(id: "sub-2", ref: "CPR-2", sourceID: "s1", subjectTypeID: "type-person", label: "Bob", description: ""),
        ]
        for (id, x) in [("sub-1", Int64(1)), ("sub-2", Int64(4))] {
            store.subjectPositionsBySubject[id] = CatalogSubjectPosition(subjectID: id, gridX: x, gridY: 1)
        }
        let promoted = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")

        let session = makeSession(store: store)
        let handle: QueryHandle<EvidenceGraphRows> = session.query(
            CatalogQueryKey.evidenceGraph(project: session.projectKey, graphId: "g-1")
        )
        await waitForQuery(handle)

        #expect(handle.value?.memberships.map(\.subjectID) == ["sub-1"])
        let snapshot = EvidenceGraphSnapshot.build(
            rows: handle.value ?? EvidenceGraphRows(graphId: "g-1"),
            types: store.subjectTypesByProject[projectDir] ?? [],
            rules: []
        )
        let alice = snapshot.subjects.first { $0.id == "sub-1" }
        let bob = snapshot.subjects.first { $0.id == "sub-2" }
        #expect(alice?.membership?.entity.ref == promoted.entity.ref)
        #expect(alice?.membership?.kind == "person")
        #expect(alice?.membership?.claimID == promoted.claim.id)
        #expect(bob != nil && bob?.membership == nil)
    }

    @Test func personsListLoadsHeadersOfPromotedPersons() async throws {
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
        store.observationsByGraph["g-1"] = [
            nameObservation(id: "o1", subjectID: "sub-1", form: "James Robins",
                            parts: [part("James", "given"), part("Robins", "surname")]),
            nameObservation(id: "o2", subjectID: "sub-1", form: "Jim Robins",
                            parts: [part("Jim", "given"), part("Robins", "surname")]),
            nameObservation(id: "o3", subjectID: "sub-1", form: "ROBINS, James",
                            parts: [part("ROBINS", "surname"), part("James", "given")]),
        ]
        let james = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")
        let bare = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-2")

        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogPersonHeader]> = session.query(
            CatalogQueryKey.personsList(project: session.projectKey)
        )
        await waitForQuery(handle)

        let headers = try #require(handle.value)
        #expect(headers.map(\.entity.ref) == [james.entity.ref, bare.entity.ref])
        #expect(headers[0].name == CatalogNameValue(
            form: "James Robins", parts: [part("James", "given"), part("Robins", "surname")]
        ))
        #expect(headers[0].nameValueCount == 2)
        #expect(headers[1].name == nil && PersonHeaderDisplay.title(headers[1]) == bare.entity.ref)
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func personsListKeepsAccentVariantsApart() async throws {
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
        ]
        store.observationsByGraph["g-1"] = [
            nameObservation(id: "o1", subjectID: "sub-1", form: "José Robins",
                            parts: [part("José", "given"), part("Robins", "surname")]),
            nameObservation(id: "o2", subjectID: "sub-1", form: "Jose Robins",
                            parts: [part("Jose", "given"), part("Robins", "surname")]),
        ]
        _ = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")

        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogPersonHeader]> = session.query(
            CatalogQueryKey.personsList(project: session.projectKey)
        )
        await waitForQuery(handle)

        #expect(try #require(handle.value)[0].nameValueCount == 2)
    }

    @Test func conclusionDetailLoadsAPersonsNameValuesAndOutcomes() async throws {
        let store = FakeStore()
        seedStore(store)
        store.subjectTypesByProject[projectDir] = [
            CatalogSubjectType(
                id: "type-person", key: "person", origin: "provenencia", label: "Person",
                description: "", refPrefix: "PER", candidateRefPrefix: "CPR"
            ),
        ]
        store.evidenceGraphsByProject[projectDir] = [
            CatalogEvidenceGraph(id: "g-1", ref: "GPH-1", sourceID: "s1", name: "Default", description: "", artifactIDs: []),
        ]
        store.subjectsByGraph["g-1"] = [
            CatalogSubject(id: "sub-1", ref: "CPR-1", sourceID: "s1", subjectTypeID: "type-person", label: "", description: ""),
        ]
        store.observationsByGraph["g-1"] = [
            nameObservation(id: "o1", subjectID: "sub-1", form: "James Robins",
                            parts: [part("James", "given"), part("Robins", "surname")]),
            nameObservation(id: "o2", subjectID: "sub-1", form: "Jim Robins",
                            parts: [part("Jim", "given"), part("Robins", "surname")]),
            nameObservation(id: "o3", subjectID: "sub-1", form: "ROBINS, James",
                            parts: [part("ROBINS", "surname"), part("James", "given")]),
            nameObservation(id: "o4", subjectID: "sub-1", form: "J. Robins", parts: []),
        ]
        let james = try await store.promoteSubject(projectDir: projectDir, userID: "user-1", subjectID: "sub-1")

        let session = makeSession(store: store)
        let handle: QueryHandle<CatalogConclusionDetail> = session.query(
            CatalogQueryKey.conclusionDetail(project: session.projectKey, entityId: james.entity.id)
        )
        await waitForQuery(handle)

        let detail = try #require(handle.value)
        #expect(detail.entity == james.entity)
        let name = try #require(detail.fields.first { $0.propertyKey == "name" })
        // One Source: o1 and o3 are one value by parts, o2 another; o4 has no parts.
        #expect(name.state == .single)
        #expect(name.values.map(\.reason) == [.kept, .outvoted])
        let james1 = CatalogNameValue(form: "James Robins", parts: [part("James", "given"), part("Robins", "surname")])
        #expect(name.displayedValues.map(\.value) == [.name(james1)])
        #expect(name.outcomes.map(\.observationRef) == ["OBS-o1", "OBS-o3", "OBS-o2", "OBS-o4"])
        #expect(name.outcomes.map(\.valueRank) == [1, 1, 2, nil])
        #expect(name.outcomes.map(\.reason) == [.kept, .kept, .outvoted, .noEvidence])
        #expect(name.outcomes.allSatisfy { $0.subjectRef == "CPR-1" && $0.sourceID == "s1" })
        // One Source voted, so the outvoted Jim lost 1 of 1.
        #expect(name.outcomes.map(\.voteSupport) == [0, 0, 1, 0])
        #expect(name.outcomes.map(\.voteTotal) == [0, 0, 1, 0])
        #expect(detail.memberCount == 1)
        #expect(store.heldCatalogProjectDir == projectDir)
    }

    @Test func hiddenPolicies() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        #expect(registry.hiddenPolicy(.conclusionDetail(project: project, entityId: "e1")) == .evict)
        #expect(registry.hiddenPolicy(.personsList(project: project)) == .markStale)
        #expect(registry.hiddenPolicy(.eventsList(project: project)) == .markStale)
        #expect(registry.hiddenPolicy(.placesList(project: project)) == .markStale)
        #expect(registry.hiddenPolicy(.sourcesList(project: project)) == .markStale)
        #expect(registry.hiddenPolicy(.sourceWorkspace(project: project, sourceId: "s1")) == .revalidate)
    }

    @Test func lastUpdatedPatchesAreDeclaredOnTheFourLists() {
        let project = ProjectKey(projectDir: projectDir)
        func follows(_ key: CatalogQueryKey) -> CatalogQueryFollows? {
            key.kind.descriptor.patches.compactMap(\.follows).first
        }
        for key in [CatalogQueryKey.personsList(project: project), .eventsList(project: project), .placesList(project: project)] {
            #expect(follows(key) == .handle, "\(key)")
        }
        #expect(follows(.sourcesList(project: project)) == .source)
        #expect(follows(.sourceTypesList(project: project)) == nil)
    }

    @Test func conclusionDetailOfAnUnknownHandleIsCodedNotFound() async throws {
        let store = FakeStore()
        seedStore(store)
        await #expect(throws: CoreInvokeError.coded(status: 1, code: "conclusiondetails.not_found", kind: .user, params: [])) {
            _ = try await store.getConclusionDetail(projectDir: projectDir, entityID: "nope")
        }
        #expect(
            CoreInvokeError.coded(status: 1, code: "conclusiondetails.not_found", kind: .user, params: []).localizedDescription
                == "This record no longer exists. It may have been merged into another."
        )
    }

    @Test func eventAndPlaceListsAreGosRowsAsGiven() async throws {
        let store = FakeStore()
        let late = CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: "e2", ref: "EVT-2", subjectTypeID: "t", label: ""),
            title: CatalogEventTitle(rule: .ref, ref: "EVT-2")
        )
        let early = CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "EVT-1", subjectTypeID: "t", label: ""),
            title: CatalogEventTitle(rule: .ref, ref: "EVT-1")
        )
        store.seededEventHeaders = [late, early]
        let place = CatalogPlaceHeader(
            entity: CatalogCanonicalEntity(id: "p1", ref: "PLC-1", subjectTypeID: "t", label: ""),
            names: ["Montréal", "Montreal"]
        )
        store.seededPlaceHeaders = [place]
        let detail = CatalogConclusionDetail(entity: early.entity, fields: [], memberCount: 1, header: .event(early))
        store.seededConclusionDetails = [early.entity.id: detail]
        let session = makeSession(store: store)

        let events: QueryHandle<[CatalogEventHeader]> = session.query(.eventsList(project: session.projectKey))
        let places: QueryHandle<[CatalogPlaceHeader]> = session.query(.placesList(project: session.projectKey))
        let page: QueryHandle<CatalogConclusionDetail> = session.query(
            .conclusionDetail(project: session.projectKey, entityId: early.entity.id)
        )
        await waitForQuery(events)
        await waitForQuery(places)
        await waitForQuery(page)
        // Order is Go's; the fake does not re-sort or re-derive.
        #expect(events.value == [late, early])
        #expect(places.value == [place])
        #expect(page.value?.header?.event == early)
    }

    private func part(_ value: String, _ type: String) -> CatalogNameValuePart {
        CatalogNameValuePart(value: value, type: type)
    }

    private func nameObservation(
        id: String, subjectID: String, form: String, parts: [CatalogNameValuePart]
    ) -> CatalogObservation {
        var observation = CatalogObservation(
            id: id, ref: "OBS-\(id)", citationID: "c1", subjectID: subjectID, propertyID: "p-name",
            polarity: "positive", valueText: form, valueInteger: nil, valueDateID: "",
            valueNameID: "n-\(id)", nameForm: form, valueSubjectID: "", valueTermID: "",
            propertyKey: "name", propertyLabel: "Name", propertyValueType: "name"
        )
        observation.nameParts = parts
        return observation
    }

    @Test func registryBackedQueryDoesNotRecallLoader() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)

        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForQuery(handle)

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
        await waitForQuery(listHandle)
        await waitForQuery(workspaceHandle)
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
        await waitForQuery(handle)

        store.sourcesByProject[projectDir] = [
            CatalogSource(id: "s1", ref: "SRC-1", sourceTypeID: "", title: "Alpha", description: ""),
            CatalogSource(id: "s2", ref: "SRC-2", sourceTypeID: "", title: "New Row", description: ""),
        ]

        session.apply(.createdSource)
        let _: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.value?.count == 2)
        // FakeStore listSources is newest-created-first (id DESC), matching Go.
        #expect(handle.value?.map(\.title) == ["New Row", "Alpha"])
    }

    /// T-14: a write's effects name the Evidence graphs to reload, one key each;
    /// a mutation that names a graph does the same, and a Source delete reloads every cached graph.
    @Test func evidenceGraphEffectsReloadOnlyThatGraph() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        let graph2 = CatalogQueryInvalidation.key(.evidenceGraph(project: project, graphId: "g-2"))
        #expect(registry.invalidations(for: CatalogWriteEffects(evidenceGraphs: ["g-2"]), project: project) == [graph2])
        #expect(registry.invalidations(by: .promotedSubject(graphId: "g-1"), project: project).first
            == .key(.evidenceGraph(project: project, graphId: "g-1")))
        let deleted = registry.invalidations(by: .deletedSubject, project: project)
        #expect(!deleted.contains { $0 == .allCached(.evidenceGraph) || $0 == graph2 })
        #expect(registry.invalidations(by: .deletedSource(id: "src-1"), project: project)
            .contains(.allCached(.evidenceGraph)))
    }

    /// Work under a Source moves its "Updated" revision. The write reports its
    /// activity, so the cached row is patched in place: no reload, no stale mark.
    @Test func applySourceWriteEffectsPatchesTheListRowWithoutReloading() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourcesList(project: session.projectKey)

        let handle: QueryHandle<[CatalogSource]> = session.query(key)
        await waitForQuery(handle)
        store.heldCatalogProjectDir = nil

        let activity = CatalogActivity(revision: 42, at: "2026-10-15T16:00:00Z")
        session.apply(CatalogWriteEffects(sourceScopes: ["s1"], activity: activity))
        await Task.yield()
        #expect(handle.value?.first?.updated == activity)
        #expect(!handle.isFetching)
        #expect(store.heldCatalogProjectDir == nil)

        // Not stale either: the next `query` is served from the patched value.
        let _: QueryHandle<[CatalogSource]> = session.query(key)
        #expect(!handle.isFetching)
        #expect(handle.value?.first?.updated == activity)
    }

    /// Writes under a Source report their activity in `WriteEffects`, so no
    /// mutation reloads the Sources list; the row is patched instead.
    @Test func sourceWritesDoNotInvalidateTheSourcesList() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        let list = CatalogQueryInvalidation.key(.sourcesList(project: project))
        let writes: [CatalogMutation] = [
            .mutatedSourceWorkspace(sourceId: "s1"), .mutatedSourceMetadata(sourceId: "s1"),
            .deletedSubject,
            .promotedSubject(graphId: "g-1"),
        ]
        for mutation in writes {
            #expect(!registry.invalidations(by: mutation, project: project).contains(list), "\(mutation)")
        }
    }

    /// T-17 (AC-2.14): Source-page and graph writes no longer ride the blanket
    /// Conclusion triggers; their effects name the handles they moved.
    @Test func blanketTriggersAreGone() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        #expect(!CatalogQueryRegistry.conclusionTriggers.contains(.mutatedSourceWorkspace))
        #expect(!CatalogQueryRegistry.conclusionTriggers.contains(.deletedSubject))
        let conclusionKeys: [CatalogQueryInvalidation] = [
            .key(.personsList(project: project)), .key(.eventsList(project: project)),
            .key(.placesList(project: project)), .allCached(.conclusionDetail),
        ]
        for mutation: CatalogMutation in [
            .mutatedSourceWorkspace(sourceId: "s1"), .mutatedSourceMetadata(sourceId: "s1"),
            .deletedSubject,
        ] {
            let invalidations = registry.invalidations(by: mutation, project: project)
            for key in conclusionKeys {
                #expect(!invalidations.contains(key), "\(mutation) \(key)")
            }
        }
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
        await waitForQuery(fieldsHandle)
        await waitForQuery(typesHandle)

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
            await waitForQuery(fieldsHandle)
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
        await waitForQuery(typesHandle)
        await waitForQuery(fieldsHandle)

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
            await waitForQuery(typesHandle)
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
        await waitForQuery(suggestionsHandle)
        await waitForQuery(otherHandle)
        await waitForQuery(typesHandle)

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
        await waitForQuery(suggestionsHandle)
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
        #expect(registry.invalidations(by: .deletedSource(id: "s1"), project: project) == [
            .key(.sourcesList(project: project)),
            .key(.sourceTypesList(project: project)),
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
            .allCached(.evidenceGraph),
            .key(.evidenceGraphsList(project: project)),
            .key(.personsList(project: project)),
            .key(.eventsList(project: project)),
            .key(.placesList(project: project)),
            .allCached(.conclusionDetail),
        ])
        // Adding vocabulary touches only the list that owns it — the Source
        // page reads that list rather than carrying its own copy.
        #expect(registry.invalidations(by: .createdMetadataField, project: project) == [
            .key(.metadataFieldsList(project: project)),
        ])
        #expect(registry.invalidations(by: .createdSourceType, project: project) == [
            .key(.sourceTypesList(project: project)),
        ])
        #expect(registry.invalidations(by: .deletedSourceType(id: "t1"), project: project) == [
            .key(.sourceTypesList(project: project)),
            .allCached(.typeSuggestions),
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
        #expect(registry.invalidations(for: CatalogWriteEffects(
            handles: ["e1"],
            evidenceGraphs: ["g-1"],
            citationSources: ["s1"],
            lists: [.propertyUsage],
            headerHandles: ["e2"]
        ), project: project) == [
            .key(.evidenceGraph(project: project, graphId: "g-1")),
            .allCached(.citationCounts),
            .allCached(.citationsByArtifact),
            .key(.propertiesWorkspace(project: project)),
            .key(.personsList(project: project)),
            .key(.eventsList(project: project)),
            .key(.placesList(project: project)),
            .key(.conclusionDetail(project: project, entityId: "e1")),
            .key(.conclusionDetail(project: project, entityId: "e2")),
        ])
        // A key the effects and the mutation both name appears once.
        let both = registry.invalidations(
            for: CatalogWriteEffects(evidenceGraphs: ["g-1"]),
            by: .promotedSubject(graphId: "g-1"),
            project: project
        )
        #expect(both.filter { $0 == .key(.evidenceGraph(project: project, graphId: "g-1")) }.count == 1)
        // A Subject delete reports its handles and its graph in `WriteEffects`, so
        // the lists, the details and the graph are not named here.
        // The citation reads follow the write's `citationSources`, not this mutation.
        #expect(registry.invalidations(by: .deletedSubject, project: project) == [
            .key(.propertiesWorkspace(project: project)),
        ])
        let erased = registry.invalidations(by: .deletedSubject, project: project)
        #expect(!erased.contains(.allCached(.citationCounts)) && !erased.contains(.allCached(.citationsByArtifact)))
        #expect(registry.invalidations(for: CatalogWriteEffects(citationSources: ["s1"]), project: project) == [
            .allCached(.citationCounts), .allCached(.citationsByArtifact),
        ])
        #expect(registry.hiddenPolicy(.citationCounts(project: project, graphId: "g1")) == .markStale)
        // A Promote changes that graph's cards (their membership row) and every
        // Conclusion key.
        #expect(registry.invalidations(by: .promotedSubject(graphId: "g-1"), project: project) == [
            .key(.evidenceGraph(project: project, graphId: "g-1")),
            .key(.personsList(project: project)),
            .key(.eventsList(project: project)),
            .key(.placesList(project: project)),
            .allCached(.conclusionDetail),
        ])
        #expect(registry.invalidations(by: .createdPropertyTerm(propertyId: "p1"), project: project) == [
            .key(.propertyTerms(project: project, propertyId: "p1")),
        ])
        // Grades are seeded vocabulary with no CRUD surface, so nothing stales them.
        let everyMutation: [CatalogMutation] = [
            .updatedSource(source), .changedSourceType(source), .createdSource, .deletedSource(id: "s1"),
            .createdSourceType, .updatedSourceType(id: "t1"), .deletedSourceType(id: "t1"),
            .createdMetadataField, .updatedMetadataField(id: "f1"), .deletedMetadataField(id: "f1"),
            .assignedTypeSuggestion(typeId: "t1"), .removedTypeSuggestion(typeId: "t1"),
            .mutatedSourceWorkspace(sourceId: "s1"), .mutatedSourceMetadata(sourceId: "s1"),
            .deletedSubject,
            .promotedSubject(graphId: "g-1"),
            .createdPropertyTerm(propertyId: "p1"),
        ]
        #expect(everyMutation.allSatisfy { mutation in
            !registry.invalidations(by: mutation, project: project).contains(
                .key(.credibilityGradesList(project: project))
            )
        })
        // Claim confidence grades are seeded the same way (S9-12).
        #expect(everyMutation.allSatisfy { mutation in
            !registry.invalidations(by: mutation, project: project).contains(
                .key(.confidenceGradesList(project: project))
            )
        })
        // Writes that name their source stay narrow: a Source-page write reports
        // its handles in `WriteEffects`, so the mutation names only the page.
        #expect(registry.invalidations(by: .mutatedSourceWorkspace(sourceId: "s1"), project: project) == [
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
        ])
        // Every Conclusion trigger stales the Person, Event, and Place lists and
        // every cached detail; nothing else does.
        for mutation in everyMutation {
            let invalidations = registry.invalidations(by: mutation, project: project)
            let isTrigger = mutation.invalidationKind.map { CatalogQueryRegistry.conclusionTriggers.contains($0) } ?? false
            #expect(invalidations.contains(.key(.personsList(project: project))) == isTrigger, "\(mutation)")
            #expect(invalidations.contains(.key(.eventsList(project: project))) == isTrigger, "\(mutation)")
            #expect(invalidations.contains(.key(.placesList(project: project))) == isTrigger, "\(mutation)")
            #expect(invalidations.contains(.allCached(.conclusionDetail)) == isTrigger, "\(mutation)")
        }
        #expect(CatalogQueryRegistry.conclusionTriggers == [.promotedSubject, .deletedSource, .updatedProperty])
        #expect(registry.invalidations(by: .mutatedSourceMetadata(sourceId: "s1"), project: project) == [
            .key(.metadataFieldsList(project: project)),
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
        ])
        #expect(registry.invalidations(by: .changedSourceType(source), project: project) == [
            .key(.sourceTypesList(project: project)),
            .key(.sourceWorkspace(project: project, sourceId: "s1")),
        ])
    }

    /// A scope move (a note save) changes a row's last-updated time, not what
    /// the lists or detail pages show, so it invalidates nothing; the session
    /// patches the row from the write's activity.
    @Test func handleScopesInvalidateNothing() {
        let registry = CatalogQueryRegistry.standard
        let project = ProjectKey(projectDir: projectDir)
        let lists: [CatalogQueryInvalidation] = [
            .key(.personsList(project: project)),
            .key(.eventsList(project: project)),
            .key(.placesList(project: project)),
        ]
        #expect(registry.invalidations(for: CatalogWriteEffects(handleScopes: ["e1"]), project: project).isEmpty)
        #expect(registry.invalidations(
            for: CatalogWriteEffects(handles: ["e1"], handleScopes: ["e1", "e2"]),
            project: project
        ) == lists + [.key(.conclusionDetail(project: project, entityId: "e1"))])
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
        await waitForQuery(workspaceHandle)
        await waitForQuery(fieldsHandle)
        store.heldCatalogProjectDir = nil

        store.fieldsByProject[projectDir]?.append(
            CatalogMetadataField(
                id: "f2", key: "folio", origin: "user", label: "Folio", dataType: "text", description: ""
            )
        )
        session.apply(.createdMetadataField)
        session.apply(.createdSourceType)
        await waitForQuery(fieldsHandle)
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
        await waitForQuery(firstHandle)
        await waitForQuery(secondHandle)

        store.metadataBySource["s1"] = [suggestedAuthorEntry(label: "Author name")]
        store.metadataBySource["s2"] = [suggestedAuthorEntry(label: "Author name")]
        session.apply(.updatedMetadataField(id: "f1"))

        let _: QueryHandle<CatalogSourceWorkspace> = session.query(firstKey)
        let _: QueryHandle<CatalogSourceWorkspace> = session.query(secondKey)
        await waitForQuery(firstHandle)
        await waitForQuery(secondHandle)
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
        await waitForQuery(firstHandle)
        await waitForQuery(secondHandle)

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
        await waitForQuery(firstHandle)
        await waitForQuery(secondHandle)
        #expect(firstHandle.value?.first?.field.label == "Author name")
        #expect(secondHandle.value?.first?.field.label == "Author name")
    }

    /// `usedBy` gates Delete on the Metadata page, and it moves whenever a
    /// source gains or loses a value.
    @Test func sourceMetadataWriteInvalidatesFieldsList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.metadataFieldsList(project: session.projectKey)

        let handle: QueryHandle<[CatalogMetadataField]> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.value?.first?.usedBy == 0)

        store.fieldsByProject[projectDir]?[0].usedBy = 1
        session.apply(.mutatedSourceMetadata(sourceId: "s1"))

        let _: QueryHandle<[CatalogMetadataField]> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.value?.first?.usedBy == 1)
    }

    /// Adding a source moves its type's `usedBy`, which gates deleting the type.
    @Test func createdSourceInvalidatesTypesList() async {
        let store = FakeStore()
        seedStore(store)
        let session = makeSession(store: store)
        let key = CatalogQueryKey.sourceTypesList(project: session.projectKey)

        let handle: QueryHandle<[CatalogSourceType]> = session.query(key)
        await waitForQuery(handle)
        #expect(handle.value?.first?.usedBy == 0)

        store.sourceTypesByProject[projectDir]?[0].usedBy = 1
        session.apply(.createdSource)

        let _: QueryHandle<[CatalogSourceType]> = session.query(key)
        await waitForQuery(handle)
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
        await waitForQuery(listHandle)
        await waitForQuery(workspaceHandle)

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
        await waitForQuery(workspaceHandle)
        #expect(workspaceHandle.value?.source.sourceTypeID == "t2")
        #expect(workspaceHandle.value?.metadata.count == 1)
    }

    private struct ListCase: CustomTestStringConvertible, Sendable {
        let name: String
        let lists: Set<CatalogList>
        let structure: Bool
        let propertiesInvalidations: Int
        let graphListInvalidations: Int

        var testDescription: String { name }
    }

    nonisolated private static let listCases: [ListCase] = [
        ListCase(name: "vocabulary", lists: [.vocabulary], structure: false, propertiesInvalidations: 1, graphListInvalidations: 0),
        ListCase(name: "property usage", lists: [.propertyUsage], structure: false, propertiesInvalidations: 1, graphListInvalidations: 0),
        ListCase(
            name: "vocabulary and property usage", lists: [.vocabulary, .propertyUsage],
            structure: false, propertiesInvalidations: 1, graphListInvalidations: 0
        ),
        ListCase(name: "structure alone", lists: [], structure: true, propertiesInvalidations: 1, graphListInvalidations: 0),
        ListCase(name: "vocabulary and structure", lists: [.vocabulary], structure: true, propertiesInvalidations: 1, graphListInvalidations: 0),
        ListCase(name: "evidence graphs alone", lists: [.evidenceGraphs], structure: false, propertiesInvalidations: 0, graphListInvalidations: 1),
    ]

    @Test(arguments: listCases)
    private func listInvalidations(_ tc: ListCase) {
        let project = ProjectKey(projectDir: projectDir)
        let got = CatalogQueryRegistry.standard.invalidations(
            for: CatalogWriteEffects(lists: tc.lists, structure: tc.structure),
            project: project
        )
        #expect(got.filter { $0 == .key(.propertiesWorkspace(project: project)) }.count == tc.propertiesInvalidations)
        #expect(got.filter { $0 == .key(.evidenceGraphsList(project: project)) }.count == tc.graphListInvalidations)
    }

    @Test func evidenceGraphsListLoadsFromStore() async throws {
        let store = FakeStore()
        let graph = CatalogEvidenceGraph(
            id: "g1", ref: "GPH-AAAAA", sourceID: "s1", name: "Default", description: "", artifactIDs: ["a1"]
        )
        store.evidenceGraphsByProject[projectDir] = [graph]
        let session = makeSession(store: store)
        let handle: QueryHandle<[CatalogEvidenceGraph]> = session.query(
            .evidenceGraphsList(project: session.projectKey)
        )
        await waitForQuery(handle)
        #expect(handle.value == [graph])
    }

    @Test func deletedSourceInvalidatesEvidenceGraphsList() {
        let project = ProjectKey(projectDir: projectDir)
        let got = CatalogQueryRegistry.standard.invalidations(by: .deletedSource(id: "s1"), project: project)
        #expect(got.contains(.key(.evidenceGraphsList(project: project))))
    }

    @Test func evidenceGraphsListMarksStaleWhenHidden() {
        let project = ProjectKey(projectDir: projectDir)
        #expect(CatalogQueryRegistry.standard.hiddenPolicy(.evidenceGraphsList(project: project)) == .markStale)
    }

    @Test func navCountTriggersAreSourceAndPromoteWrites() {
        #expect(CatalogQueryRegistry.navCountTriggers == [.createdSource, .deletedSource, .promotedSubject])
    }
}
