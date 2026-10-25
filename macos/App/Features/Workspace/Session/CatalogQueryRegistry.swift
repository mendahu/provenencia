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

    private struct Spec {
        let kind: CatalogQueryKey.Kind
        let stalePolicy: CatalogQueryStalePolicy
        let invalidateOn: Set<CatalogMutationKind>
        /// What an invalidated key that isn't on screen does. The exception
        /// to cache rule 2 for lists and per-handle pages, so one trigger
        /// doesn't reload every list or page the researcher has opened.
        var whenHidden: CatalogHiddenPolicy = .revalidate
        /// How a write that reports its activity moves this list's rows in
        /// place (no load). `nil` for lists without a last-updated cell.
        var lastUpdated: CatalogLastUpdatedPatch?
    }

    private let specs: [Spec] = [
        Spec(
            kind: .sourcesList,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdSource, .deletedSource],
            // Work under a Source reports its activity in `WriteEffects`, which
            // patches the row through `lastUpdated`.
            whenHidden: .markStale,
            lastUpdated: .rows(CatalogSource.self, scope: .source)
        ),
        Spec(
            kind: .sourceTypesList,
            stalePolicy: .sessionFresh,
            // `usedBy` counts sources per type, so adding a source or moving one
            // restates rows no type edit touched.
            invalidateOn: [
                .createdSourceType, .updatedSourceType, .deletedSourceType,
                .createdSource, .deletedSource, .changedSourceType,
            ]
        ),
        Spec(
            kind: .metadataFieldsList,
            stalePolicy: .sessionFresh,
            // `usedBy` counts sources holding a value for the field, and it
            // gates Delete.
            invalidateOn: [
                .createdMetadataField, .updatedMetadataField, .deletedMetadataField,
                .mutatedSourceMetadata,
            ]
        ),
        Spec(
            kind: .credibilityGradesList,
            stalePolicy: .sessionFresh,
            // Seeded vocabulary with no CRUD surface, so nothing stales it.
            invalidateOn: []
        ),
        Spec(
            kind: .confidenceGradesList,
            stalePolicy: .sessionFresh,
            // Seeded vocabulary with no CRUD surface, so nothing stales it.
            invalidateOn: []
        ),
        Spec(
            kind: .sourceWorkspace,
            stalePolicy: .sessionFresh,
            // Vocabulary lists are no longer folded in here, so plain
            // create/relabel of a type or field leaves the page alone. What
            // remains is the metadata row set the engine derives per source:
            // suggestions ∪ values, each row embedding its field.
            invalidateOn: [
                .mutatedSourceWorkspace, .mutatedSourceMetadata, .changedSourceType,
                .updatedMetadataField, .deletedMetadataField,
                .assignedTypeSuggestion, .removedTypeSuggestion,
                .deletedSource,
            ]
        ),
        Spec(
            kind: .typeSuggestions,
            stalePolicy: .sessionFresh,
            // Each suggestion embeds a full field row, so relabelling or
            // deleting a field reaches types the edit never named.
            invalidateOn: [
                .assignedTypeSuggestion, .removedTypeSuggestion,
                .updatedMetadataField, .deletedMetadataField,
                .deletedSourceType,
            ]
        ),
        Spec(
            kind: .evidenceGraph,
            stalePolicy: .sessionFresh,
            // Writes that report effects reload the graphs they name
            // (`WriteEffects.evidenceGraphs`); Promote reports none.
            invalidateOn: [.deletedSource, .promotedSubject]
        ),
        Spec(
            kind: .citationCounts,
            stalePolicy: .sessionFresh,
            // Reloads follow the write's `citationSources` effects.
            invalidateOn: [],
            whenHidden: .markStale
        ),
        Spec(
            kind: .propertiesWorkspace,
            stalePolicy: .sessionFresh,
            // `usedBy` counts Observations per property. Observation writes
            // report their Source in `WriteEffects`, which invalidates it.
            invalidateOn: [
                .createdProperty, .updatedProperty, .deletedProperty,
                .mutatedSubjectTypeProperties, .deletedSubject,
            ],
            whenHidden: .markStale
        ),
        Spec(
            kind: .connectRules,
            stalePolicy: .sessionFresh,
            // Seeded product matrix with no CRUD surface.
            invalidateOn: []
        ),
        Spec(
            kind: .propertyTerms,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdPropertyTerm]
        ),
        Spec(
            kind: .citationsByArtifact,
            stalePolicy: .sessionFresh,
            // Reloads follow the write's `citationSources` effects.
            invalidateOn: [],
            whenHidden: .markStale
        ),
        Spec(
            kind: .evidenceGraphProgress,
            stalePolicy: .sessionFresh,
            // Canvas writes patch one graph via Get + setQueryValue — do not
            // restale the project-wide map (or sourcesList).
            invalidateOn: []
        ),
        Spec(
            kind: .evidenceGraphsList,
            stalePolicy: .sessionFresh,
            // Graph and link writes report the `evidence_graphs` list in
            // `WriteEffects`; `deleteSource` reports none, so `.deletedSource` drops it.
            invalidateOn: [.deletedSource],
            // A visible key still revalidates (`WorkspaceSession.apply(invalidations:)`);
            // a hidden list just goes stale until it is shown again.
            whenHidden: .markStale
        ),
        Spec(
            kind: .personsList,
            stalePolicy: .sessionFresh,
            invalidateOn: CatalogQueryRegistry.conclusionTriggers,
            whenHidden: .markStale,
            lastUpdated: .rows(CatalogPersonHeader.self, scope: .handle)
        ),
        Spec(
            kind: .eventsList,
            stalePolicy: .sessionFresh,
            invalidateOn: CatalogQueryRegistry.conclusionTriggers,
            whenHidden: .markStale,
            lastUpdated: .rows(CatalogEventHeader.self, scope: .handle)
        ),
        Spec(
            kind: .placesList,
            stalePolicy: .sessionFresh,
            invalidateOn: CatalogQueryRegistry.conclusionTriggers,
            whenHidden: .markStale,
            lastUpdated: .rows(CatalogPlaceHeader.self, scope: .handle)
        ),
        Spec(
            kind: .conclusionDetail,
            stalePolicy: .sessionFresh,
            invalidateOn: CatalogQueryRegistry.conclusionTriggers,
            whenHidden: .evict
        ),
    ]

    func stalePolicy(for key: CatalogQueryKey) -> CatalogQueryStalePolicy {
        specs.first { $0.kind == key.kind }?.stalePolicy ?? .sessionFresh
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

    /// Resolves cache effects from registry `invalidateOn` tags + mutation payload.
    func invalidations(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryInvalidation] {
        guard let mutationKind = mutation.invalidationKind else { return [] }
        return specs
            .filter { $0.invalidateOn.contains(mutationKind) }
            .map { $0.kind.invalidation(project: project, mutation: mutation) }
    }

    /// What an invalidated `key` does while off screen (see `Spec`).
    func hiddenPolicy(_ key: CatalogQueryKey) -> CatalogHiddenPolicy {
        specs.first { $0.kind == key.kind }?.whenHidden ?? .revalidate
    }

    /// How a write that reports its activity moves `key`'s rows in place.
    func lastUpdatedPatch(_ key: CatalogQueryKey) -> CatalogLastUpdatedPatch? {
        specs.first { $0.kind == key.kind }?.lastUpdated
    }

}

private extension CatalogQueryKey.Kind {
    /// Builds the cache effect for one registry row and mutation. Id-bearing
    /// kinds fall back to `allCached` when the mutation names no single owner.
    func invalidation(project: ProjectKey, mutation: CatalogMutation) -> CatalogQueryInvalidation {
        switch self {
        case .sourcesList:
            return .key(.sourcesList(project: project))
        case .sourceTypesList:
            return .key(.sourceTypesList(project: project))
        case .metadataFieldsList:
            return .key(.metadataFieldsList(project: project))
        case .credibilityGradesList:
            return .key(.credibilityGradesList(project: project))
        case .confidenceGradesList:
            return .key(.confidenceGradesList(project: project))
        case .sourceWorkspace:
            switch mutation {
            case .mutatedSourceWorkspace(let sourceId), .mutatedSourceMetadata(let sourceId),
                 .deletedSource(let sourceId):
                return .key(.sourceWorkspace(project: project, sourceId: sourceId))
            case .changedSourceType(let source):
                return .key(.sourceWorkspace(project: project, sourceId: source.id))
            default:
                return .allCached(.sourceWorkspace)
            }
        case .typeSuggestions:
            switch mutation {
            case .assignedTypeSuggestion(let typeId), .removedTypeSuggestion(let typeId):
                return .key(.typeSuggestions(project: project, typeId: typeId))
            default:
                return .allCached(.typeSuggestions)
            }
        case .evidenceGraph:
            if let graphId = mutation.evidenceGraphId {
                return .key(.evidenceGraph(project: project, graphId: graphId))
            }
            return .allCached(.evidenceGraph)
        case .citationCounts:
            return .allCached(.citationCounts)
        case .propertiesWorkspace:
            return .key(.propertiesWorkspace(project: project))
        case .connectRules:
            return .key(.connectRules(project: project))
        case .propertyTerms:
            switch mutation {
            case .createdPropertyTerm(let propertyId):
                return .key(.propertyTerms(project: project, propertyId: propertyId))
            default:
                return .allCached(.propertyTerms)
            }
        case .citationsByArtifact:
            return .allCached(.citationsByArtifact)
        case .evidenceGraphProgress:
            return .key(.evidenceGraphProgress(project: project))
        case .evidenceGraphsList:
            return .key(.evidenceGraphsList(project: project))
        case .personsList:
            return .key(.personsList(project: project))
        case .eventsList:
            return .key(.eventsList(project: project))
        case .placesList:
            return .key(.placesList(project: project))
        case .conclusionDetail:
            // Any Conclusion edit can touch any handle (a merge, a member's
            // Observation), so every cached detail is stale.
            return .allCached(.conclusionDetail)
        }
    }
}
