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

/// Declarative catalog read loaders and mutation invalidation map.
struct CatalogQueryRegistry: Sendable {
    static let standard = CatalogQueryRegistry()

    private struct Spec {
        let kind: CatalogQueryKey.Kind
        let stalePolicy: CatalogQueryStalePolicy
        let invalidateOn: Set<CatalogMutationKind>
    }

    private let specs: [Spec] = [
        Spec(
            kind: .sourcesList,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdSource]
        ),
        Spec(
            kind: .sourceTypesList,
            stalePolicy: .sessionFresh,
            // `usedBy` counts sources per type, so adding a source or moving one
            // restates rows no type edit touched.
            invalidateOn: [
                .createdSourceType, .updatedSourceType, .deletedSourceType,
                .createdSource, .changedSourceType,
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
            ]
        ),
        Spec(
            kind: .sourceGraph,
            stalePolicy: .sessionFresh,
            invalidateOn: [.mutatedSourceGraph, .savedCitation]
        ),
        Spec(
            kind: .citationCounts,
            stalePolicy: .sessionFresh,
            invalidateOn: [.savedCitation]
        ),
        Spec(
            kind: .subjectFieldsWorkspace,
            stalePolicy: .sessionFresh,
            invalidateOn: [
                .createdProperty, .updatedProperty, .deletedProperty,
                .mutatedSubjectTypeFields,
            ]
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
            invalidateOn: [.savedCitation]
        ),
        Spec(
            kind: .sourceGraphProgress,
            stalePolicy: .sessionFresh,
            // Canvas writes patch one Source via Get + setQueryValue — do not
            // restale the project-wide map (or sourcesList).
            invalidateOn: []
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
        case .sourceWorkspace(let project, let sourceId):
            return try await store.getSourceWorkspace(projectDir: project.projectDir, sourceID: sourceId)
        case .typeSuggestions(let project, let typeId):
            return try await store.listTypeSuggestions(projectDir: project.projectDir, typeID: typeId)
        case .citationCounts(let project, let sourceId):
            return try await store.citationCountsBySource(
                projectDir: project.projectDir,
                sourceID: sourceId
            )
        case .sourceGraph(let project, let sourceId):
            async let subjects = store.listSubjects(
                projectDir: project.projectDir,
                sourceID: sourceId
            )
            async let positions = store.listSubjectPositions(
                projectDir: project.projectDir,
                sourceID: sourceId
            )
            async let observations = store.listObservationsBySource(
                projectDir: project.projectDir,
                sourceID: sourceId
            )
            return SourceGraphRows(
                sourceId: sourceId,
                subjects: try await subjects,
                positions: try await positions,
                observations: try await observations
            )
        case .connectRules:
            return try await store.listConnectRules()
        case .propertyTerms(let project, let propertyId):
            return try await store.listPropertyTerms(
                projectDir: project.projectDir,
                propertyID: propertyId
            )
        case .citationsByArtifact(let project, let artifactId):
            return try await store.listCitationsByArtifact(
                projectDir: project.projectDir,
                artifactID: artifactId
            )
        case .sourceGraphProgress(let project):
            let rows = try await store.listSourceGraphProgress(projectDir: project.projectDir)
            var map: [String: SourceGraphProgress] = [:]
            for row in rows {
                map[row.sourceId] = row
            }
            return map
        case .subjectFieldsWorkspace(let project):
            return try await store.getSubjectFieldsWorkspace(projectDir: project.projectDir)
        }
    }

    /// Resolves cache effects from registry `invalidateOn` tags + mutation payload.
    func invalidations(by mutation: CatalogMutation, project: ProjectKey) -> [CatalogQueryInvalidation] {
        guard let mutationKind = mutation.invalidationKind else { return [] }
        return specs
            .filter { $0.invalidateOn.contains(mutationKind) }
            .map { $0.kind.invalidation(project: project, mutation: mutation) }
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
        case .sourceWorkspace:
            switch mutation {
            case .mutatedSourceWorkspace(let sourceId), .mutatedSourceMetadata(let sourceId):
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
        case .sourceGraph:
            switch mutation {
            case .mutatedSourceGraph(let sourceId), .savedCitation(let sourceId):
                return .key(.sourceGraph(project: project, sourceId: sourceId))
            default:
                return .allCached(.sourceGraph)
            }
        case .citationCounts:
            switch mutation {
            case .savedCitation(let sourceId):
                return .key(.citationCounts(project: project, sourceId: sourceId))
            default:
                return .allCached(.citationCounts)
            }
        case .subjectFieldsWorkspace:
            return .key(.subjectFieldsWorkspace(project: project))
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
        case .sourceGraphProgress:
            return .key(.sourceGraphProgress(project: project))
        }
    }
}
