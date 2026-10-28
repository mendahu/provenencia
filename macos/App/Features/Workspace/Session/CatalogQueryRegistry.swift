import Foundation

enum CatalogQueryRegistryError: Error {
    case typeMismatch
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
/// the last-updated patch for lists that move a row in place.
struct CatalogQueryDescriptor: Sendable {
    let kind: CatalogQueryKey.Kind
    let stalePolicy: CatalogQueryStalePolicy
    let invalidateOn: Set<CatalogMutationKind>
    /// What an invalidated key that isn't on screen does. The exception
    /// to cache rule 2 for lists and per-handle pages, so one trigger
    /// doesn't reload every list or page the researcher has opened.
    var whenHidden: CatalogHiddenPolicy = .revalidate
    let invalidation: @Sendable (ProjectKey, CatalogMutation) -> CatalogQueryInvalidation
    /// How a write that reports its activity moves this list's rows in
    /// place (no load). `nil` for lists without a last-updated cell.
    var lastUpdated: CatalogLastUpdatedPatch?
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

    func load(key: CatalogQueryKey, store: any GenealogyStore) async throws -> Any {
        switch key {
        case .sourcesList(let project):
            return try await store.listSources(projectDir: project.projectDir)
        case .sourceTypesList(let project):
            return try await store.listSourceTypes(projectDir: project.projectDir)
        case .metadataFieldsList(let project):
            return try await store.listMetadataFields(projectDir: project.projectDir)
        case .credibilityGradesList(let project):
            return try await store.listSourceCredibilityGrades(projectDir: project.projectDir)
        case .confidenceGradesList(let project):
            return try await store.listClaimConfidenceGrades(projectDir: project.projectDir)
        case .sourceWorkspace(let project, let sourceId):
            return try await store.getSourceWorkspace(projectDir: project.projectDir, sourceID: sourceId)
        case .typeSuggestions(let project, let typeId):
            return try await store.listTypeSuggestions(projectDir: project.projectDir, typeID: typeId)
        case .citationCounts(let project, let graphId):
            return try await store.citationCountsByEvidenceGraph(
                projectDir: project.projectDir,
                graphID: graphId
            )
        case .evidenceGraph(let project, let graphId):
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
        case .connectRules:
            return try await store.listConnectRules()
        case .propertyTerms(let project, let propertyId):
            return try await store.listPropertyTerms(
                projectDir: project.projectDir,
                propertyID: propertyId
            )
        case .citationsByArtifact(let project, let graphId, let artifactId):
            return try await store.listCitationsByArtifact(
                projectDir: project.projectDir,
                graphID: graphId,
                artifactID: artifactId
            )
        case .evidenceGraphProgress(let project):
            let rows = try await store.listEvidenceGraphProgress(projectDir: project.projectDir)
            var map: [String: EvidenceGraphProgress] = [:]
            for row in rows {
                map[row.graphId] = row
            }
            return map
        case .evidenceGraphsList(let project):
            return try await store.listEvidenceGraphs(projectDir: project.projectDir)
        case .propertiesWorkspace(let project):
            return try await store.getPropertiesWorkspace(projectDir: project.projectDir)
        case .personsList(let project):
            return try await store.listPersonHeaders(projectDir: project.projectDir)
        case .eventsList(let project):
            return try await store.listEventHeaders(projectDir: project.projectDir)
        case .placesList(let project):
            return try await store.listPlaceHeaders(projectDir: project.projectDir)
        case .conclusionDetail(let project, let entityId):
            return try await store.getConclusionDetail(projectDir: project.projectDir, entityID: entityId)
        }
    }

    /// Cache effects of one write's reported handles and Sources. Each read
    /// follows what feeds it: each Evidence graph per `evidenceGraphs`, the citation
    /// reads per `citationSources`, the catalog-wide lists per `lists`, and the
    /// Properties workspace on structure. Handle scopes invalidate nothing: they move
    /// through `lastUpdatedPatch`.
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

    /// How a write that reports its activity moves `key`'s rows in place.
    func lastUpdatedPatch(_ key: CatalogQueryKey) -> CatalogLastUpdatedPatch? {
        descriptor(for: key).lastUpdated
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
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [.createdSource, .deletedSource],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.sourcesList(project: project)) },
                lastUpdated: .rows(CatalogSource.self, scope: .source)
            )
        case .sourceTypesList:
            // `usedBy` counts sources per type, so adding a source or moving one
            // restates rows no type edit touched.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .createdSourceType, .updatedSourceType, .deletedSourceType,
                    .createdSource, .deletedSource, .changedSourceType,
                ],
                invalidation: { project, _ in .key(.sourceTypesList(project: project)) }
            )
        case .metadataFieldsList:
            // `usedBy` counts sources holding a value for the field, and it
            // gates Delete.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .createdMetadataField, .updatedMetadataField, .deletedMetadataField,
                    .mutatedSourceMetadata,
                ],
                invalidation: { project, _ in .key(.metadataFieldsList(project: project)) }
            )
        case .credibilityGradesList:
            // Seeded vocabulary with no CRUD surface, so nothing stales it.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.credibilityGradesList(project: project)) }
            )
        case .confidenceGradesList:
            // Seeded vocabulary with no CRUD surface, so nothing stales it.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.confidenceGradesList(project: project)) }
            )
        case .sourceWorkspace:
            // Vocabulary lists are no longer folded in here, so plain
            // create/relabel of a type or field leaves the page alone. What
            // remains is the metadata row set the engine derives per source:
            // suggestions ∪ values, each row embedding its field.
            return CatalogQueryDescriptor(
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
                }
            )
        case .evidenceGraph:
            // Writes that report effects reload the graphs they name
            // (`WriteEffects.evidenceGraphs`); Promote reports none.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [.deletedSource, .promotedSubject],
                invalidation: { project, mutation in
                    if let graphId = mutation.evidenceGraphId {
                        return .key(.evidenceGraph(project: project, graphId: graphId))
                    }
                    return .allCached(.evidenceGraph)
                }
            )
        case .citationCounts:
            // Reloads follow the write's `citationSources` effects.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { _, _ in .allCached(.citationCounts) }
            )
        case .typeSuggestions:
            // Each suggestion embeds a full field row, so relabelling or
            // deleting a field reaches types the edit never named.
            return CatalogQueryDescriptor(
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
                }
            )
        case .propertiesWorkspace:
            // `usedBy` counts Observations per property. Observation writes
            // report their Source in `WriteEffects`, which invalidates it.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [
                    .createdProperty, .updatedProperty, .deletedProperty,
                    .mutatedSubjectTypeProperties, .deletedSubject,
                ],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.propertiesWorkspace(project: project)) }
            )
        case .connectRules:
            // Seeded product matrix with no CRUD surface.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.connectRules(project: project)) }
            )
        case .propertyTerms:
            return CatalogQueryDescriptor(
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
                }
            )
        case .citationsByArtifact:
            // Reloads follow the write's `citationSources` effects.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                whenHidden: .markStale,
                invalidation: { _, _ in .allCached(.citationsByArtifact) }
            )
        case .evidenceGraphProgress:
            // Canvas writes patch one graph via Get + setQueryValue — do not
            // restale the project-wide map (or sourcesList).
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [],
                invalidation: { project, _ in .key(.evidenceGraphProgress(project: project)) }
            )
        case .evidenceGraphsList:
            // Graph and link writes report the `evidence_graphs` list in
            // `WriteEffects`; `deleteSource` reports none, so `.deletedSource` drops it.
            // A visible key still revalidates (`WorkspaceSession.apply(invalidations:)`);
            // a hidden list just goes stale until it is shown again.
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: [.deletedSource],
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.evidenceGraphsList(project: project)) }
            )
        case .personsList:
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.personsList(project: project)) },
                lastUpdated: .rows(CatalogPersonHeader.self, scope: .handle)
            )
        case .eventsList:
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.eventsList(project: project)) },
                lastUpdated: .rows(CatalogEventHeader.self, scope: .handle)
            )
        case .placesList:
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .markStale,
                invalidation: { project, _ in .key(.placesList(project: project)) },
                lastUpdated: .rows(CatalogPlaceHeader.self, scope: .handle)
            )
        case .conclusionDetail:
            return CatalogQueryDescriptor(
                kind: self,
                stalePolicy: .sessionFresh,
                invalidateOn: CatalogQueryRegistry.conclusionTriggers,
                whenHidden: .evict,
                // Any Conclusion edit can touch any handle (a merge, a member's
                // Observation), so every cached detail is stale.
                invalidation: { _, _ in .allCached(.conclusionDetail) }
            )
        }
    }
}
