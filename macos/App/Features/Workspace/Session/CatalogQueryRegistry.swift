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
            invalidateOn: [.createdSubject, .createdCitation, .addedObservations]
        ),
        Spec(
            kind: .citationCounts,
            stalePolicy: .sessionFresh,
            invalidateOn: [.createdCitation]
        ),
        Spec(
            kind: .subjectFieldsWorkspace,
            stalePolicy: .sessionFresh,
            invalidateOn: [
                .createdProperty, .updatedProperty, .deletedProperty,
                .mutatedSubjectTypeFields,
            ]
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
            async let types = store.listSubjectTypes(projectDir: project.projectDir)
            async let observations = store.listObservationsBySource(
                projectDir: project.projectDir,
                sourceID: sourceId
            )
            return SourceGraphSnapshot.build(
                sourceId: sourceId,
                subjects: try await subjects,
                positions: try await positions,
                types: try await types,
                observations: try await observations
            )
        case .subjectFieldsWorkspace(let project):
            let dir = project.projectDir
            async let properties = store.listProperties(projectDir: dir)
            async let types = store.listSubjectTypes(projectDir: dir)
            let loadedTypes = try await types
            var fieldsByTypeID: [String: [CatalogSubjectTypeField]] = [:]
            var presentationsByKey: [String: CatalogSubjectTypePresentation] = [:]
            for type in loadedTypes {
                fieldsByTypeID[type.id] = try await store.listSubjectTypeFields(
                    projectDir: dir,
                    subjectTypeID: type.id
                )
                if let presentation = try? await store.getSubjectTypePresentation(typeKey: type.key) {
                    presentationsByKey[type.key] = presentation
                }
            }
            return SubjectFieldsSnapshot(
                properties: try await properties,
                types: loadedTypes,
                fieldsByTypeID: fieldsByTypeID,
                presentationsByKey: presentationsByKey
            )
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
            case .createdSubject(let sourceId),
                 .createdCitation(let sourceId),
                 .addedObservations(let sourceId):
                return .key(.sourceGraph(project: project, sourceId: sourceId))
            default:
                return .allCached(.sourceGraph)
            }
        case .citationCounts:
            switch mutation {
            case .createdCitation(let sourceId):
                return .key(.citationCounts(project: project, sourceId: sourceId))
            default:
                return .allCached(.citationCounts)
            }
        case .subjectFieldsWorkspace:
            return .key(.subjectFieldsWorkspace(project: project))
        }
    }
}
