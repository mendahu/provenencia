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
    /// Partial reload of a warmed value. Nil when this kind has none.
    var refresh: CatalogQueryRefresh? = nil
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
        refresh: CatalogQueryRefresh? = nil,
        load: @escaping @Sendable (CatalogQueryKey, any GenealogyStore) async throws -> Value
    ) -> CatalogQueryDescriptor {
        CatalogQueryDescriptor(
            kind: kind,
            stalePolicy: stalePolicy,
            invalidateOn: invalidateOn,
            whenHidden: whenHidden,
            invalidation: invalidation,
            patches: patches,
            refresh: refresh,
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

    /// Cache effects of one write's reported handles and Sources. Each read
    /// follows what feeds it: each Evidence graph per `evidenceGraphs`, the citation
    /// reads per `citationSources`, the catalog-wide lists per `lists`, and the
    /// Properties workspace on structure. Handle scopes invalidate nothing: they move
    /// through each list's `patches`.
    func invalidations(for effects: CatalogWriteEffects, project: ProjectKey) -> [CatalogQueryInvalidation] {
        var out: [CatalogQueryInvalidation] = []
        for graphID in effects.evidenceGraphs {
            out.append(.key(.evidenceGraph(project: project, graphId: graphID)))
        }
        // The effects name Sources and the keys name graphs: drop every cached
        // citation read once.
        if !effects.citationSources.isEmpty {
            out.append(.allCached(.citationCounts))
            out.append(.allCached(.citationsByArtifact))
        }
        var listed = CatalogList.allCases.filter { effects.lists.contains($0) }
            .flatMap { Self.invalidations(for: $0, project: project) }
        if effects.structure {
            listed.append(.key(.propertiesWorkspace(project: project)))
        }
        for invalidation in listed where !out.contains(invalidation) {
            out.append(invalidation)
        }
        if !effects.touchedHandles.isEmpty {
            out.append(.key(.personsList(project: project)))
            out.append(.key(.eventsList(project: project)))
            out.append(.key(.placesList(project: project)))
        }
        for entityID in effects.touchedHandles {
            out.append(.key(.conclusionDetail(project: project, entityId: entityID)))
        }
        return out
    }

    /// The cached reads one catalog-wide list feeds.
    static func invalidations(for list: CatalogList, project: ProjectKey) -> [CatalogQueryInvalidation] {
        switch list {
        case .vocabulary, .propertyUsage:
            return [.key(.propertiesWorkspace(project: project))]
        case .evidenceGraphs:
            return [.key(.evidenceGraphsList(project: project))]
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
                invalidateOn: [.createdSource, .deletedSource],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.sourcesList(project: project)) },
                patches: [
                    .lastUpdated(CatalogSource.self, follows: .source),
                    .updatedSourceRow(),
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
        case .confidenceGradesList:
            // Seeded vocabulary with no CRUD surface, so nothing stales it.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.confidenceGradesList(project: project)) },
                load: { key, store in
                    guard case .confidenceGradesList(let project) = key else {
                        mismatchedCatalogKey(key, .confidenceGradesList)
                    }
                    return try await store.listClaimConfidenceGrades(projectDir: project.projectDir)
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
            // Writes that report effects reload the graphs they name
            // (`WriteEffects.evidenceGraphs`); Promote reports none.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [.deletedSource, .promotedSubject],
                invalidation: { project, mutation in
                    if let graphId = mutation.evidenceGraphId {
                        return .key(.evidenceGraph(project: project, graphId: graphId))
                    }
                    return .allCached(.evidenceGraph)
                },
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
            // Reloads follow the write's `citationSources` effects.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { _, _ in .allCached(.citationCounts) },
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
            // Reloads follow the write's `citationSources` effects.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { _, _ in .allCached(.citationsByArtifact) },
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
            // Canvas writes patch one graph via Get + setQueryValue — do not
            // restale the project-wide map (or sourcesList).
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.evidenceGraphProgress(project: project)) },
                refresh: CatalogQueryRefresh(follows: .evidenceGraphs) { session, key, ids in
                    for graphId in ids {
                        guard let _: QueryHandle<[String: EvidenceGraphProgress]> = session.queryHandle(key) else {
                            return
                        }
                        do {
                            let one = try await session.store.getEvidenceGraphProgress(
                                projectDir: key.project.projectDir,
                                graphID: graphId
                            )
                            var map: [String: EvidenceGraphProgress] = session.queryHandle(key)?.value ?? [:]
                            if one.isZero {
                                map.removeValue(forKey: graphId)
                            } else {
                                map[graphId] = one
                            }
                            session.setQueryValue(key, value: map)
                        } catch {
                            // Keep last counts; the next mutation or the next place visit that warms it retries.
                        }
                    }
                },
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
            // Graph and link writes report the `evidence_graphs` list in
            // `WriteEffects`; `deleteSource` reports none, so `.deletedSource` drops it.
            // A visible key still revalidates (`WorkspaceSession.apply(invalidations:)`);
            // a hidden list just goes stale until it is shown again.
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [.deletedSource],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.evidenceGraphsList(project: project)) },
                load: { key, store in
                    guard case .evidenceGraphsList(let project) = key else {
                        mismatchedCatalogKey(key, .evidenceGraphsList)
                    }
                    return try await store.listEvidenceGraphs(projectDir: project.projectDir)
                }
            )
        case .personsList:
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.personsList(project: project)) },
                patches: [.lastUpdated(CatalogPersonHeader.self, follows: .handle)],
                load: { key, store in
                    guard case .personsList(let project) = key else {
                        mismatchedCatalogKey(key, .personsList)
                    }
                    return try await store.listPersonHeaders(projectDir: project.projectDir)
                }
            )
        case .eventsList:
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.eventsList(project: project)) },
                patches: [.lastUpdated(CatalogEventHeader.self, follows: .handle)],
                load: { key, store in
                    guard case .eventsList(let project) = key else {
                        mismatchedCatalogKey(key, .eventsList)
                    }
                    return try await store.listEventHeaders(projectDir: project.projectDir)
                }
            )
        case .placesList:
            return .make(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.placesList(project: project)) },
                patches: [.lastUpdated(CatalogPlaceHeader.self, follows: .handle)],
                load: { key, store in
                    guard case .placesList(let project) = key else {
                        mismatchedCatalogKey(key, .placesList)
                    }
                    return try await store.listPlaceHeaders(projectDir: project.projectDir)
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
                load: { key, store in
                    guard case .conclusionDetail(let project, let entityId) = key else {
                        mismatchedCatalogKey(key, .conclusionDetail)
                    }
                    return try await store.getConclusionDetail(projectDir: project.projectDir, entityID: entityId)
                }
            )
        }
    }
}
