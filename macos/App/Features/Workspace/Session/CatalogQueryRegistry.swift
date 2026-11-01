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
    /// Drop the key; it loads fresh when next visited.
    case evict
}

/// One catalog cache kind: its policies, how a mutation invalidates it, and
/// the in-place patches for lists and the Source page.
struct CatalogQueryDescriptor: Sendable {
    let kind: CatalogQueryKey.Kind
    let stalePolicy: CatalogQueryStalePolicy
    let invalidateOn: Set<CatalogMutationKind>
    /// What an invalidated key that isn't on screen does. The exception
    /// to cache rule 2 for lists and per-handle pages, so one trigger
    /// doesn't reload every list or page the researcher has opened.
    var whenHidden: CatalogHiddenPolicy = .revalidate
    let invalidation: @Sendable (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation
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
        stalePolicy: CatalogQueryStalePolicy,
        invalidateOn: Set<CatalogMutationKind>,
        whenHidden: CatalogHiddenPolicy = .revalidate,
        invalidation: @escaping @Sendable (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation,
        patches: [CatalogQueryPatch] = [],
        refreshes: [CatalogQueryRefresh] = [],
        load: @escaping @Sendable (CatalogQueryKey, any GenealogyStore) async throws -> Value
    ) -> CatalogQueryDescriptor {
        CatalogQueryDescriptor(
            kind: kind,
            stalePolicy: stalePolicy,
            invalidateOn: invalidateOn,
            whenHidden: whenHidden,
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

/// Declarative catalog read loaders and mutation invalidation map.
struct CatalogQueryRegistry: Sendable {
    static let standard = CatalogQueryRegistry()

    /// Writes that can change any Conclusion value and report no
    /// `WriteEffects` (Spike 9 Q5). Every Conclusion key — lists and details —
    /// invalidates on this one set. A write that reports its handles (a
    /// citation, a credibility change, a Subject delete) invalidates through
    /// `invalidations(for:)` instead, and a note reports none, so a Source-page
    /// write no longer reloads every list. A Property update, including a
    /// cardinality change, rides `updatedProperty` (S9-37).
    static let conclusionTriggers: Set<CatalogMutationKind> = [
        .promotedSubject, .deletedSource, .updatedProperty,
    ]

    /// Writes that change a sidebar count (Sources, Persons, Events, Places).
    /// `WorkspaceSession.apply` recounts after each; features never call
    /// `CatalogCounts.refreshAll()`.
    static let navCountTriggers: Set<CatalogMutationKind> = [
        .createdSource, .deletedSource, .promotedSubject,
    ]

    func stalePolicy(for key: CatalogQueryKey) -> CatalogQueryStalePolicy {
        descriptor(for: key).stalePolicy
    }

    /// Cache effects of one write's reported lists and structure: the
    /// catalog-wide lists per `lists`, and the Properties workspace on
    /// structure. Evidence graphs, handles and keyed `reads` invalidate nothing
    /// here: each kind's `refreshes` follow them (`WorkspaceSession.revalidatedKeys`).
    /// Handle scopes invalidate nothing: they move through each list's `patches`.
    func invalidations(for effects: CatalogWriteEffects, project: ProjectKey) -> [CatalogQueryInvalidation] {
        var listed = CatalogList.allCases.filter { effects.lists.contains($0) }
            .flatMap { Self.invalidations(for: $0, project: project) }
        if effects.structure {
            listed.append(.key(.propertiesWorkspace(project: project)))
        }
        var out: [CatalogQueryInvalidation] = []
        for invalidation in listed where !out.contains(invalidation) {
            out.append(invalidation)
        }
        return out
    }

    /// The cached reads one catalog-wide list feeds.
    static func invalidations(for list: CatalogList, project: ProjectKey) -> [CatalogQueryInvalidation] {
        switch list {
        case .vocabulary, .propertyUsage:
            return [.key(.propertiesWorkspace(project: project))]
        }
    }

    /// The deduped union of what a write's effects reported and what its
    /// mutation names, in that order, so a key both name is invalidated once.
    func invalidations(
        for effects: CatalogWriteEffects,
        by mutation: CatalogMutation?,
        project: ProjectKey
    ) -> [CatalogQueryInvalidation] {
        var all = invalidations(for: effects, project: project)
        if let mutation {
            all += invalidations(by: mutation, project: project)
        }
        var seen = Set<CatalogQueryInvalidation>()
        return all.filter { seen.insert($0).inserted }
    }

    /// Resolves cache effects from each kind's `invalidateOn` tags + mutation payload.
    func invalidations(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryInvalidation] {
        guard let mutationKind = mutation.invalidationKind else { return [] }
        return CatalogQueryKey.Kind.allCases.compactMap { kind in
            let descriptor = kind.descriptor
            guard descriptor.invalidateOn.contains(mutationKind) else { return nil }
            return descriptor.invalidation(project, mutation)
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
                stalePolicy: .sessionFresh,
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
            // `usedBy` counts sources per type, so adding a source or moving one
            // restates rows no type edit touched.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .createdSourceType, .updatedSourceType, .deletedSourceType,
                    .createdSource, .deletedSource, .changedSourceType,
                ],
                invalidation: { project, _ in .key(.sourceTypesList(project: project)) },
                load: { key, store in
                    guard case .sourceTypesList(let project) = key else {
                        mismatchedCatalogKey(key, .sourceTypesList)
                    }
                    return try await store.listSourceTypes(projectDir: project.projectDir)
                }
            )
        case .metadataFieldsList:
            // `usedBy` counts sources holding a value for the field, and it
            // gates Delete.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .createdMetadataField, .updatedMetadataField, .deletedMetadataField,
                    .mutatedSourceMetadata,
                ],
                invalidation: { project, _ in .key(.metadataFieldsList(project: project)) },
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
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.credibilityGradesList(project: project)) },
                load: { key, store in
                    guard case .credibilityGradesList(let project) = key else {
                        mismatchedCatalogKey(key, .credibilityGradesList)
                    }
                    return try await store.listSourceCredibilityGrades(projectDir: project.projectDir)
                }
            )
        case .sourceWorkspace:
            // Vocabulary lists are no longer folded in here, so plain
            // create/relabel of a type or field leaves the page alone. What
            // remains is the metadata row set the engine derives per source:
            // suggestions ∪ values, each row embedding its field.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
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
            // Writes that report effects name the graphs they move
            // (`WriteEffects.evidenceGraphs`); Promote reports none. The graph
            // on screen reloads; a visited graph that is hidden goes stale and
            // reloads when its canvas, composer or Promote is shown again.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [.deletedSource, .promotedSubject],
                whenHidden: .markStale,
                invalidation: { project, mutation in
                    if let graphId = mutation.evidenceGraphId {
                        return .key(.evidenceGraph(project: project, graphId: graphId))
                    }
                    return .allCached(.evidenceGraph)
                },
                refreshes: [.applyHiddenPolicy(follows: .evidenceGraphs) { _, key in
                    guard case .evidenceGraph(_, let graphId) = key else { return [] }
                    return [graphId]
                }],
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
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { _, _ in .allCached(.citationCounts) },
                refreshes: [.applyHiddenPolicy(follows: .reads(.citations)) { _, key in
                    guard case .citationCounts(_, let graphId) = key else { return [] }
                    return [graphId]
                }],
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
            // deleting a field reaches types the edit never named.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .assignedTypeSuggestion, .removedTypeSuggestion,
                    .updatedMetadataField, .deletedMetadataField,
                    .deletedSourceType,
                ],
                invalidation: { project, mutation in
                    switch mutation {
                    case .assignedTypeSuggestion(let typeId), .removedTypeSuggestion(let typeId):
                        return .key(.typeSuggestions(project: project, typeId: typeId))
                    default:
                        return .allCached(.typeSuggestions)
                    }
                },
                load: { key, store in
                    guard case .typeSuggestions(let project, let typeId) = key else {
                        mismatchedCatalogKey(key, .typeSuggestions)
                    }
                    return try await store.listTypeSuggestions(projectDir: project.projectDir, typeID: typeId)
                }
            )
        case .propertiesWorkspace:
            // `usedBy` counts Observations per property. Observation writes
            // report their Source in `WriteEffects`, which invalidates it.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .createdProperty, .updatedProperty, .deletedProperty,
                    .mutatedSubjectTypeProperties, .deletedSubject,
                ],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.propertiesWorkspace(project: project)) },
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
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.connectRules(project: project)) },
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
                stalePolicy: .sessionFresh,
                invalidateOn: [.createdPropertyTerm],
                invalidation: { project, mutation in
                    switch mutation {
                    case .createdPropertyTerm(let propertyId):
                        return .key(.propertyTerms(project: project, propertyId: propertyId))
                    default:
                        return .allCached(.propertyTerms)
                    }
                },
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
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { _, _ in .allCached(.citationsByArtifact) },
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
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.evidenceGraphProgress(project: project)) },
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
            // graphs and merges them into the warmed list. `deleteSource` reports
            // none, so `CatalogMutation.readIDs(for:)` names the deleted Source.
            // A hidden list goes stale until it is shown again.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.evidenceGraphsList(project: project)) },
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
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.personsList(project: project)) },
                patches: [.lastUpdated(CatalogPersonSummary.self, follows: .handle)],
                refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)],
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
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.eventsList(project: project)) },
                patches: [.lastUpdated(CatalogEventSummary.self, follows: .handle)],
                refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)],
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
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.placesList(project: project)) },
                patches: [.lastUpdated(CatalogPlaceSummary.self, follows: .handle)],
                refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)],
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
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .evict,
                // Any Conclusion edit can touch any handle (a merge, a member's
                // Observation), so every cached detail is stale.
                invalidation: { _, _ in .allCached(.conclusionDetail) },
                refreshes: [.applyHiddenPolicy(follows: .touchedHandles) { _, key in
                    guard case .conclusionDetail(_, let entityId) = key else { return [] }
                    return [entityId]
                }],
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
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .evict,
                invalidation: { _, _ in .allCached(.personClaims) },
                refreshes: [
                    .applyHiddenPolicy(follows: .touchedHandles) { _, key in
                        guard case .personClaims(_, let entityId) = key else { return [] }
                        return [entityId]
                    },
                    .applyHiddenPolicy(followsValue: .evidenceGraphs) { (groups: [CatalogPersonClaimGroup]) in
                        Set(groups.flatMap(\.claims).map(\.graphID))
                    },
                ],
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
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .evict,
                invalidation: { _, _ in .allCached(.personFamily) },
                refreshes: [.applyHiddenPolicyOnAny(follows: .touchedHandles)],
                load: { key, store in
                    guard case .personFamily(let project, let entityId) = key else {
                        mismatchedCatalogKey(key, .personFamily)
                    }
                    return try await store.getPersonFamily(projectDir: project.projectDir, entityID: entityId)
                }
            )
        }
    }
}
