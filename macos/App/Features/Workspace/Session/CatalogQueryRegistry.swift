import Foundation

private func mismatchedCatalogKey(_ key: CatalogQueryKey, _ kind: CatalogQueryKey.Kind) -> Never {
    preconditionFailure("CatalogQueryKey \(key) is not \(kind)")
}

/// One cache effect of a mutation. Keys carrying an id resolve to `allCached`
/// when the write names no single owner — a new metadata field belongs to every
/// cached Source page, not one of them.
enum CatalogQueryInvalidation: Hashable, Sendable {
    case key(CatalogQueryKey)
    case allCached(CatalogQueryKey.Kind)
}

/// What an invalidation does to a cached key that is not on screen.
enum CatalogHiddenPolicy: Sendable, Equatable {
    /// Reload now (the default).
    case revalidate
    /// Keep the old value and reload on the next `query`; nothing fetches now.
    case markStale
}

/// One catalog cache kind: its policies, how a mutation invalidates it, and
/// the in-place patches for lists and the Source page.
struct CatalogQueryDescriptor: Sendable {
    let kind: CatalogQueryKey.Kind
    let invalidateOn: Set<CatalogMutationKind>
    /// What an invalidated key that isn't on screen does. The exception
    /// to cache rule 2 for lists and per-handle pages, so one trigger
    /// doesn't reload every list or page the researcher has opened.
    var whenHidden: CatalogHiddenPolicy = .revalidate
    /// The entity whose id `ownId` is. Its deletion evicts the hidden key
    /// (`apply`), and `.handle` keys are the pages the hidden-entity cap counts
    /// (`apply(location:)`).
    var deletedWith: CatalogDeletedEntity? = nil
    /// Whether this is a per-entity page key, one the hidden-entity cap counts
    /// and `shownOrder` orders: its `ownId` is a handle.
    var isEntityPage: Bool { deletedWith == .handle }
    /// Nil when the kind names no mutation in `invalidateOn`.
    let invalidation: (@Sendable (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation)?
    /// In-place updates run, in order, before invalidation. Empty for kinds
    /// that only reload.
    var patches: [CatalogQueryPatch] = []
    /// How a warmed key follows the ids of a write's effects: which slot, and
    /// whether it revalidates the key or reloads part of its value. Empty when
    /// this kind follows none.
    var refreshes: [CatalogQueryRefresh] = []
    /// Starts the typed load for this kind. Closed over `Value` by `make`.
    let warm: @MainActor @Sendable (WorkspaceSession, CatalogQueryKey) -> Void
}

extension CatalogQueryDescriptor {
    /// One factory per kind so the loader and the warm path share `Value`.
    /// `load` `guard case`s the key; a mismatch is unreachable when warm runs
    /// from `key.kind.descriptor`.
    static func make<Value>(
        kind: CatalogQueryKey.Kind,
        invalidateOn: Set<CatalogMutationKind> = [],
        whenHidden: CatalogHiddenPolicy = .revalidate,
        deletedWith: CatalogDeletedEntity? = nil,
        invalidation: (@Sendable (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation)? = nil,
        patches: [CatalogQueryPatch] = [],
        refreshes: [CatalogQueryRefresh] = [],
        load: @escaping @Sendable (CatalogQueryKey, any GenealogyStore) async throws -> Value
    ) -> CatalogQueryDescriptor {
        precondition(
            invalidateOn.isEmpty == (invalidation == nil),
            "\(kind): a kind names mutations exactly when it says how they invalidate"
        )
        return CatalogQueryDescriptor(
            kind: kind,
            invalidateOn: invalidateOn,
            whenHidden: whenHidden,
            deletedWith: deletedWith,
            invalidation: invalidation,
            patches: patches,
            refreshes: refreshes,
            warm: { session, key in
                let store = session.store
                let _: QueryHandle<Value> = session.ensureQuery(key) {
                    try await load(key, store)
                }
            }
        )
    }
}

extension CatalogQueryDescriptor {
    /// A per-person section key that a relative's or co-participant's change
    /// reaches: that change touches their id, not the Person's, so any touched
    /// handle reloads it.
    static func personSection<Value>(
        kind: CatalogQueryKey.Kind,
        load: @escaping @Sendable (CatalogQueryKey, any GenealogyStore) async throws -> Value
    ) -> CatalogQueryDescriptor {
        .make(
            kind: kind,
            whenHidden: .markStale,
            deletedWith: .handle,
            refreshes: [
                .applyHiddenPolicyOnAny(follows: .touchedHandles),
                .followsPropertyLabels,
            ],
            load: load
        )
    }

    /// A per-person section key read from the Person's Evidence graphs: it
    /// reloads when its own handle or one of the graphs its value holds is
    /// written.
    static func personGraphsSection<Value>(
        kind: CatalogQueryKey.Kind,
        graphIDs: @escaping @Sendable (Value) -> Set<String>,
        load: @escaping @Sendable (CatalogQueryKey, any GenealogyStore) async throws -> Value
    ) -> CatalogQueryDescriptor {
        .make(
            kind: kind,
            whenHidden: .markStale,
            deletedWith: .handle,
            refreshes: [
                .applyHiddenPolicy(followsOwnId: .touchedHandles),
                .applyHiddenPolicy(followsValue: .evidenceGraphs, ids: graphIDs),
                .followsPropertyLabels,
            ],
            load: load
        )
    }
}

/// Declarative catalog read loaders and mutation invalidation map.
struct CatalogQueryRegistry: Sendable {
    static let standard = CatalogQueryRegistry()

    /// Resolves cache effects from each kind's `invalidateOn` tags + mutation payload.
    func invalidations(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryInvalidation] {
        guard let mutationKind = mutation.invalidationKind else { return [] }
        return CatalogQueryKey.Kind.allCases.compactMap { kind in
            let descriptor = kind.descriptor
            guard descriptor.invalidateOn.contains(mutationKind),
                  let invalidation = descriptor.invalidation else { return nil }
            return invalidation(project, mutation)
        }
    }

    /// What an invalidated `key` does while off screen.
    func hiddenPolicy(_ key: CatalogQueryKey) -> CatalogHiddenPolicy {
        descriptor(for: key).whenHidden
    }

    /// The kind's descriptor. A kind whose arm reported a different kind cannot
    /// happen with the exhaustive switch; that is a programmer error.
    private func descriptor(for key: CatalogQueryKey) -> CatalogQueryDescriptor {
        let descriptor = key.kind.descriptor
        guard descriptor.kind == key.kind else {
            preconditionFailure("missing descriptor for \(key.kind)")
        }
        return descriptor
    }
}

extension CatalogQueryKey.Kind {
    /// One arm owns this kind's policies and invalidation. Id-bearing kinds
    /// fall back to `allCached` when the mutation names no single owner.
    var descriptor: CatalogQueryDescriptor {
        switch self {
        case .sourcesList:
            // Work under a Source reports its activity in `WriteEffects`, which
            // patches the row through `lastUpdated`.
            return .make(
                kind: self,
                invalidateOn: [.deletedSource],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.sourcesList(project: project)) },
                patches: [
                    .lastUpdated(CatalogSource.self, follows: .source),
                    .sourceRow(),
                ],
                load: { key, store in
                    guard case .sourcesList(let project) = key else {
                        mismatchedCatalogKey(key, .sourcesList)
                    }
                    return try await store.listSources(projectDir: project.projectDir)
                }
            )
        case .sourceTypesList:
            // A type write reports `source_vocabulary`, which this key follows.
            // `usedBy` counts sources per type, so adding a source or moving one
            // restates rows no type edit touched.
            return .make(
                kind: self,
                invalidateOn: [.createdSource, .deletedSource, .changedSourceType],
                invalidation: { project, _ in .key(.sourceTypesList(project: project)) },
                refreshes: [.applyHiddenPolicyOnAny(follows: .list(.sourceVocabulary))],
                load: { key, store in
                    guard case .sourceTypesList(let project) = key else {
                        mismatchedCatalogKey(key, .sourceTypesList)
                    }
                    return try await store.listSourceTypes(projectDir: project.projectDir)
                }
            )
        case .metadataFieldsList:
            // A field write reports `source_vocabulary`, which this key follows.
            // `usedBy` counts sources holding a value for the field, and it
            // gates Delete.
            return .make(
                kind: self,
                invalidateOn: [.mutatedSourceMetadata],
                invalidation: { project, _ in .key(.metadataFieldsList(project: project)) },
                refreshes: [.applyHiddenPolicyOnAny(follows: .list(.sourceVocabulary))],
                load: { key, store in
                    guard case .metadataFieldsList(let project) = key else {
                        mismatchedCatalogKey(key, .metadataFieldsList)
                    }
                    return try await store.listMetadataFields(projectDir: project.projectDir)
                }
            )
        case .credibilityGradesList:
            // Seeded vocabulary with no CRUD surface, so nothing stales it.
            return .make(
                kind: self,
                load: { key, store in
                    guard case .credibilityGradesList(let project) = key else {
                        mismatchedCatalogKey(key, .credibilityGradesList)
                    }
                    return try await store.listSourceCredibilityGrades(projectDir: project.projectDir)
                }
            )
        case .sourceWorkspace:
            // The page follows no vocabulary list: a type create, or a type
            // relabel, leaves it alone. What it holds is the metadata row set
            // the engine derives per source (suggestions ∪ values, each row
            // embedding its field), so a field relabel or delete, which a list
            // cannot tell from a field create, marks every visited page stale
            // through `.updatedMetadataField` / `.deletedMetadataField`.
            return .make(
                kind: self,
                invalidateOn: [
                    .mutatedSourceWorkspace, .mutatedSourceMetadata, .changedSourceType,
                    .updatedMetadataField, .deletedMetadataField,
                    .assignedTypeSuggestion, .removedTypeSuggestion,
                    .deletedSource,
                ],
                whenHidden: .markStale,
                invalidation: { project, mutation in
                    switch mutation {
                    case .mutatedSourceWorkspace(let sourceId), .mutatedSourceMetadata(let sourceId),
                         .deletedSource(let sourceId):
                        return .key(.sourceWorkspace(project: project, sourceId: sourceId))
                    case .changedSourceType(let source):
                        return .key(.sourceWorkspace(project: project, sourceId: source.id))
                    default:
                        return .allCached(.sourceWorkspace)
                    }
                },
                patches: [.updatedSourceShell()],
                load: { key, store in
                    guard case .sourceWorkspace(let project, let sourceId) = key else {
                        mismatchedCatalogKey(key, .sourceWorkspace)
                    }
                    return try await store.getSourceWorkspace(projectDir: project.projectDir, sourceID: sourceId)
                }
            )
        case .evidenceGraph:
            // Every write that moves a graph, Promote and a Source delete
            // included, names it in `WriteEffects.evidenceGraphs`. The graph
            // on screen reloads; a visited graph that is hidden goes stale and
            // reloads when its canvas, composer or Promote is shown again.
            return .make(
                kind: self,
                whenHidden: .markStale,
                refreshes: [.applyHiddenPolicy(followsOwnId: .evidenceGraphs)],
                load: { key, store in
                    guard case .evidenceGraph(let project, let graphId) = key else {
                        mismatchedCatalogKey(key, .evidenceGraph)
                    }
                    async let subjects = store.listSubjects(
                        projectDir: project.projectDir,
                        graphID: graphId
                    )
                    async let positions = store.listSubjectPositions(
                        projectDir: project.projectDir,
                        graphID: graphId
                    )
                    async let observations = store.listObservationsByEvidenceGraph(
                        projectDir: project.projectDir,
                        graphID: graphId
                    )
                    async let memberships = store.listSubjectMemberships(
                        projectDir: project.projectDir,
                        graphID: graphId
                    )
                    async let eventTitles = store.listEvidenceGraphEventTitles(
                        projectDir: project.projectDir,
                        graphID: graphId
                    )
                    return EvidenceGraphRows(
                        graphId: graphId,
                        subjects: try await subjects,
                        positions: try await positions,
                        observations: try await observations,
                        memberships: try await memberships,
                        eventTitles: try await eventTitles
                    )
                }
            )
        case .citationCounts:
            // Reloads follow the write's `citations` read for this key's graph.
            return .make(
                kind: self,
                whenHidden: .markStale,
                refreshes: [.applyHiddenPolicy(followsOwnId: .reads(.citations))],
                load: { key, store in
                    guard case .citationCounts(let project, let graphId) = key else {
                        mismatchedCatalogKey(key, .citationCounts)
                    }
                    return try await store.citationCountsByEvidenceGraph(
                        projectDir: project.projectDir,
                        graphID: graphId
                    )
                }
            )
        case .typeSuggestions:
            // Each suggestion embeds a full field row, so relabelling or
            // deleting a field reaches types the edit never named: the key
            // follows `source_vocabulary`. The join is not audited and reports
            // no effects, so the type–field mutations name the key.
            return .make(
                kind: self,
                invalidateOn: [.assignedTypeSuggestion, .removedTypeSuggestion],
                invalidation: { project, mutation in
                    switch mutation {
                    case .assignedTypeSuggestion(let typeId), .removedTypeSuggestion(let typeId):
                        return .key(.typeSuggestions(project: project, typeId: typeId))
                    default:
                        return .allCached(.typeSuggestions)
                    }
                },
                refreshes: [.applyHiddenPolicyOnAny(follows: .list(.sourceVocabulary))],
                load: { key, store in
                    guard case .typeSuggestions(let project, let typeId) = key else {
                        mismatchedCatalogKey(key, .typeSuggestions)
                    }
                    return try await store.listTypeSuggestions(
                        projectDir: project.projectDir, typeID: typeId, level: .source
                    )
                }
            )
        case .propertiesWorkspace:
            // `usedBy` counts Observations per Property: a write that creates,
            // deletes or re-points an Observation reports `property_usage`, which
            // this key follows, as it follows a vocabulary label change and graph
            // structure.
            return .make(
                kind: self,
                whenHidden: .markStale,
                refreshes: [
                    .followsPropertyLabels,
                    .applyHiddenPolicyOnAny(follows: .list(.propertyUsage)),
                    .applyHiddenPolicyOnAny(follows: .structure),
                ],
                load: { key, store in
                    guard case .propertiesWorkspace(let project) = key else {
                        mismatchedCatalogKey(key, .propertiesWorkspace)
                    }
                    return try await store.getPropertiesWorkspace(projectDir: project.projectDir)
                }
            )
        case .connectRules:
            // Seeded product matrix with no CRUD surface.
            return .make(
                kind: self,
                load: { key, store in
                    guard case .connectRules = key else {
                        mismatchedCatalogKey(key, .connectRules)
                    }
                    return try await store.listConnectRules()
                }
            )
        case .propertyTerms:
            return .make(
                kind: self,
                // A term write reports the `property_terms` read for its Property.
                // The rows hold no Property label, binding or type, so the key
                // follows no vocabulary list.
                refreshes: [.applyHiddenPolicy(followsOwnId: .reads(.propertyTerms))],
                load: { key, store in
                    guard case .propertyTerms(let project, let propertyId) = key else {
                        mismatchedCatalogKey(key, .propertyTerms)
                    }
                    return try await store.listPropertyTerms(
                        projectDir: project.projectDir,
                        propertyID: propertyId
                    )
                }
            )
        case .citationsByArtifact:
            // Reloads follow the write's `citations` read for this key's graph.
            return .make(
                kind: self,
                whenHidden: .markStale,
                refreshes: [.applyHiddenPolicy(follows: .reads(.citations)) { _, key in
                    guard case .citationsByArtifact(_, let graphId, _) = key else { return [] }
                    return [graphId]
                }],
                load: { key, store in
                    guard case .citationsByArtifact(let project, let graphId, let artifactId) = key else {
                        mismatchedCatalogKey(key, .citationsByArtifact)
                    }
                    return try await store.listCitationsByArtifact(
                        projectDir: project.projectDir,
                        graphID: graphId,
                        artifactID: artifactId
                    )
                }
            )
        case .evidenceGraphProgress:
            // A write re-reads the counts of the graphs it names and merges them
            // into the warmed map (`.merged`); it does not restale the
            // project-wide map (or sourcesList). It keeps merging while hidden.
            return .make(
                kind: self,
                refreshes: [.merged(
                    follows: .reads(.evidenceGraphProgress),
                    as: [String: EvidenceGraphProgress].self,
                    read: { session, key, graphId in
                        try await session.store.getEvidenceGraphProgress(
                            projectDir: key.project.projectDir,
                            graphID: graphId
                        )
                    },
                    merge: { map, graphId, one in
                        if one.isZero {
                            map.removeValue(forKey: graphId)
                        } else {
                            map[graphId] = one
                        }
                    }
                )],
                load: { key, store in
                    guard case .evidenceGraphProgress(let project) = key else {
                        mismatchedCatalogKey(key, .evidenceGraphProgress)
                    }
                    let rows = try await store.listEvidenceGraphProgress(projectDir: project.projectDir)
                    var map: [String: EvidenceGraphProgress] = [:]
                    for row in rows {
                        map[row.graphId] = row
                    }
                    return map
                }
            )
        case .evidenceGraphsList:
            // Graph and link writes report the `evidence_graphs` read for their
            // Source (`WriteEffects.reads`), and the arm re-reads that Source's
            // graphs and merges them into the warmed list; `deleteSource` reports
            // it for the deleted Source too. A hidden list goes stale until it is shown again.
            return .make(
                kind: self,
                whenHidden: .markStale,
                refreshes: [.merged(
                    follows: .reads(.evidenceGraphs),
                    as: [CatalogEvidenceGraph].self,
                    read: { session, key, sourceId in
                        try await session.store.listEvidenceGraphs(
                            projectDir: key.project.projectDir,
                            sourceID: sourceId
                        )
                    },
                    merge: { rows, sourceId, part in
                        rows.removeAll { $0.sourceID == sourceId }
                        rows.append(contentsOf: part)
                    },
                    afterMerge: { session, _, parts, previous in
                        session.spliceEvidenceGraphIndex(parts, previousRevision: previous)
                    }
                )],
                load: { key, store in
                    guard case .evidenceGraphsList(let project) = key else {
                        mismatchedCatalogKey(key, .evidenceGraphsList)
                    }
                    return try await store.listEvidenceGraphs(projectDir: project.projectDir, sourceID: nil)
                }
            )
        case .personsList:
            return .make(
                kind: self,
                whenHidden: .markStale,
                patches: [.lastUpdated(CatalogPersonSummary.self, follows: .handle)],
                refreshes: [
                    .applyHiddenPolicyOnAny(follows: .touchedHandles),
                    .followsPropertyLabels,
                ],
                load: { key, store in
                    guard case .personsList(let project) = key else {
                        mismatchedCatalogKey(key, .personsList)
                    }
                    return try await store.listPersonSummaries(projectDir: project.projectDir)
                }
            )
        case .eventsList:
            return .make(
                kind: self,
                whenHidden: .markStale,
                patches: [.lastUpdated(CatalogEventSummary.self, follows: .handle)],
                refreshes: [
                    .applyHiddenPolicyOnAny(follows: .touchedHandles),
                    .followsPropertyLabels,
                ],
                load: { key, store in
                    guard case .eventsList(let project) = key else {
                        mismatchedCatalogKey(key, .eventsList)
                    }
                    return try await store.listEventSummaries(projectDir: project.projectDir)
                }
            )
        case .placesList:
            return .make(
                kind: self,
                whenHidden: .markStale,
                patches: [.lastUpdated(CatalogPlaceSummary.self, follows: .handle)],
                refreshes: [
                    .applyHiddenPolicyOnAny(follows: .touchedHandles),
                    .followsPropertyLabels,
                ],
                load: { key, store in
                    guard case .placesList(let project) = key else {
                        mismatchedCatalogKey(key, .placesList)
                    }
                    return try await store.listPlaceSummaries(projectDir: project.projectDir)
                }
            )
        case .conclusionDetail:
            return .make(
                kind: self,
                whenHidden: .markStale,
                deletedWith: .handle,
                refreshes: [
                    .applyHiddenPolicy(followsOwnId: .touchedHandles),
                    // The Sources its records cite: a rename or a Citation or
                    // Artifact label fix reaches the provenance sheet, the
                    // Gallery captions and the viewer's record row.
                    .applyHiddenPolicy(followsValue: .source) { (detail: CatalogConclusionDetail) in
                        var ids = Set<String>()
                        for field in detail.fields {
                            ids.formUnion(field.outcomes.map(\.sourceID))
                        }
                        for row in detail.derivedRows {
                            for owner in row.owners {
                                ids.formUnion(owner.field.outcomes.map(\.sourceID))
                            }
                        }
                        ids.formUnion(detail.pictures.map(\.record.sourceID))
                        return ids
                    },
                    .followsPropertyLabels,
                ],
                load: { key, store in
                    guard case .conclusionDetail(let project, let entityId) = key else {
                        mismatchedCatalogKey(key, .conclusionDetail)
                    }
                    return try await store.getConclusionDetail(projectDir: project.projectDir, entityID: entityId)
                }
            )
        case .personClaims:
            // A Person's claims by Source (S13-01). Any Conclusion edit can
            // change who a claim belongs to, so every cached key is stale on a
            // trigger. Otherwise a write reloads it for its own handle, and for
            // the Evidence graphs it holds a claim on.
            return .personGraphsSection(
                kind: self,
                graphIDs: { (groups: [CatalogPersonClaimGroup]) in
                    Set(groups.flatMap(\.claims).map(\.graphID))
                },
                load: { key, store in
                    guard case .personClaims(let project, let entityId) = key else {
                        mismatchedCatalogKey(key, .personClaims)
                    }
                    return try await store.listPersonClaims(projectDir: project.projectDir, entityID: entityId)
                }
            )
        case .personFamily:
            // A Person's family units (S13-04). A relative's change touches the
            // relative's id, not this Person's, and the relatives are up to two
            // hops away, so any write that names a handle reloads a visible
            // Family section; the read is cheap warm. A newly filed relative is
            // not in the cached value, so the key cannot follow the ids it names.
            return .personSection(kind: self) { key, store in
                guard case .personFamily(let project, let entityId) = key else {
                    mismatchedCatalogKey(key, .personFamily)
                }
                return try await store.getPersonFamily(projectDir: project.projectDir, entityID: entityId)
            }
        case .personEvents:
            // A Person's timeline (S13-05). An edit to an event or a
            // co-participant touches their id, not this Person's, and a newly
            // filed participation is not in the cached value, so any write that
            // names a handle reloads a visible Timeline section; the read is
            // cheap warm.
            return .personSection(kind: self) { key, store in
                guard case .personEvents(let project, let entityId) = key else {
                    mismatchedCatalogKey(key, .personEvents)
                }
                return try await store.listPersonEvents(projectDir: project.projectDir, entityID: entityId)
            }
        case .personRecordLinks:
            // A Person's Evidence graphs and the lines between them (S13-27).
            // Every write that can change a line names a graph the value holds
            // (a neighbor's claim names its graph; a Subject, Observation or
            // Source write names its graph) or the Person's own handle (a graph
            // gained or lost), so a write elsewhere does not reload the pairs.
            return .personGraphsSection(
                kind: self,
                graphIDs: { (links: CatalogPersonRecordLinks) in Set(links.graphs.map(\.graphID)) },
                load: { key, store in
                    guard case .personRecordLinks(let project, let entityId) = key else {
                        mismatchedCatalogKey(key, .personRecordLinks)
                    }
                    return try await store.getPersonRecordLinks(projectDir: project.projectDir, entityID: entityId)
                }
            )
        }
    }
}
