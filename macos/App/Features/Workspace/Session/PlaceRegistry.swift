import Foundation

/// Maps `WorkspaceLocation` to catalog query keys and presentation ids.
struct PlaceRegistry: Sendable {
    static let standard = PlaceRegistry()

    private struct Spec {
        let id: PlaceID
        let presentation: WorkspacePresentationID
        let priority: Int
        let matches: @Sendable (WorkspaceLocation) -> Bool
        let queryKeys: @Sendable (ProjectKey, WorkspaceLocation) -> [CatalogQueryKey]
        let deepId: @Sendable (WorkspaceLocation) -> String?
    }

    /// Higher priority wins when multiple specs could match.
    private let specs: [Spec] = [
        Spec(
            id: .sourcePromote,
            presentation: .sourcePromote,
            priority: 120,
            matches: {
                $0.section == .sources
                    && $0.sourceId != nil
                    && $0.sourceSurface == .promote
                    && $0.subjectId != nil
            },
            queryKeys: { project, location in
                guard let sourceId = location.sourceId, let subjectId = location.subjectId else { return [] }
                return [
                    .sourceGraph(project: project, sourceId: sourceId),
                    .sourcesList(project: project),
                    .promoteTargets(project: project, subjectId: subjectId),
                    .confidenceGradesList(project: project),
                    .propertiesWorkspace(project: project),
                ]
            },
            deepId: { $0.subjectId }
        ),
        Spec(
            id: .sourceCitationComposer,
            presentation: .sourceCitationComposer,
            priority: 120,
            matches: {
                $0.section == .sources
                    && $0.sourceId != nil
                    && $0.sourceSurface == .citationComposer
                    && ($0.subjectId != nil || $0.isConnectPrefill || $0.citationId != nil)
            },
            queryKeys: { project, location in
                guard let sourceId = location.sourceId else { return [] }
                var keys: [CatalogQueryKey] = [
                    .sourceGraph(project: project, sourceId: sourceId),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                    .sourceWorkspace(project: project, sourceId: sourceId),
                    .sourceTypesList(project: project),
                    .citationCounts(project: project, sourceId: sourceId),
                ]
                if let artifactId = location.artifactId, !artifactId.isEmpty {
                    keys.append(.citationsByArtifact(project: project, artifactId: artifactId))
                }
                return keys
            },
            deepId: { $0.subjectId ?? $0.citationId }
        ),
        Spec(
            id: .sourceGraph,
            presentation: .sourceGraph,
            priority: 110,
            matches: {
                $0.section == .sources && $0.sourceId != nil && $0.sourceSurface == .graph
            },
            queryKeys: { project, location in
                guard let sourceId = location.sourceId else { return [] }
                return [
                    .sourceGraph(project: project, sourceId: sourceId),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                ]
            },
            deepId: { $0.sourceId }
        ),
        Spec(
            id: .sourceDetail,
            presentation: .sourcePage,
            priority: 100,
            matches: {
                $0.section == .sources && $0.sourceId != nil && $0.sourceSurface == .page
            },
            queryKeys: { project, location in
                guard let sourceId = location.sourceId else { return [] }
                // The page's vocabulary (type picker, credibility chips,
                // Add-metadata list) comes from the shared lists, which are
                // usually already warm from the sidebar and list places.
                return [
                    .sourceWorkspace(project: project, sourceId: sourceId),
                    .sourceTypesList(project: project),
                    .metadataFieldsList(project: project),
                    .credibilityGradesList(project: project),
                ]
            },
            deepId: { $0.sourceId }
        ),
        Spec(
            id: .sourceTypesDetail,
            presentation: .sourceTypes,
            priority: 90,
            matches: { $0.section == .sourceTypes && $0.typeId != nil },
            queryKeys: { project, location in
                guard let typeId = location.typeId else { return [] }
                return [
                    .sourceTypesList(project: project),
                    .metadataFieldsList(project: project),
                    .typeSuggestions(project: project, typeId: typeId),
                ]
            },
            deepId: { $0.typeId }
        ),
        Spec(
            id: .metadata,
            presentation: .metadata,
            priority: 80,
            matches: { $0.section == .metadata },
            queryKeys: { project, _ in
                [.metadataFieldsList(project: project)]
            },
            deepId: { $0.fieldId }
        ),
        Spec(
            id: .properties,
            presentation: .properties,
            priority: 80,
            matches: { $0.section == .properties },
            queryKeys: { project, _ in
                [.propertiesWorkspace(project: project)]
            },
            deepId: { _ in nil }
        ),
        Spec(
            id: .sourcesList,
            presentation: .sourcesList,
            priority: 10,
            matches: { $0.section == .sources && $0.sourceId == nil },
            queryKeys: { project, _ in
                [
                    .sourcesList(project: project),
                    .sourceTypesList(project: project),
                    .sourceGraphProgress(project: project),
                ]
            },
            deepId: { _ in nil }
        ),
        Spec(
            id: .sourceTypes,
            presentation: .sourceTypes,
            priority: 10,
            matches: { $0.section == .sourceTypes && $0.typeId == nil },
            queryKeys: { project, _ in
                [
                    .sourceTypesList(project: project),
                    .metadataFieldsList(project: project),
                ]
            },
            deepId: { _ in nil }
        ),
        Spec(
            id: .personsList,
            presentation: .personsList,
            priority: 10,
            matches: { $0.section == .persons && $0.entityId == nil },
            queryKeys: { project, _ in [.personsList(project: project)] },
            deepId: { _ in nil }
        ),
        Spec(
            id: .personDetail,
            presentation: .personDetail,
            priority: 20,
            matches: { $0.section == .persons && $0.entityId != nil },
            // Still the stub page until S9-16 draws it; the detail loads.
            queryKeys: { project, location in
                location.entityId.map { [.conclusionDetail(project: project, entityId: $0)] } ?? []
            },
            deepId: { $0.entityId }
        ),
        Spec(
            id: .eventsList,
            presentation: .eventsList,
            priority: 10,
            matches: { $0.section == .events && $0.entityId == nil },
            queryKeys: { project, _ in [.eventsList(project: project)] },
            deepId: { _ in nil }
        ),
        Spec(
            id: .eventDetail,
            presentation: .eventDetail,
            priority: 20,
            matches: { $0.section == .events && $0.entityId != nil },
            // Stub page until S9-24; the detail loads.
            queryKeys: { project, location in
                location.entityId.map { [.conclusionDetail(project: project, entityId: $0)] } ?? []
            },
            deepId: { $0.entityId }
        ),
        Spec(
            id: .placesList,
            presentation: .placesList,
            priority: 10,
            matches: { $0.section == .places },
            // Stub page until S9-26.
            queryKeys: { _, _ in [] },
            deepId: { _ in nil }
        ),
    ]

    private var orderedSpecs: [Spec] {
        specs.sorted { $0.priority > $1.priority }
    }

    func resolve(_ location: WorkspaceLocation, project: ProjectKey) -> ResolvedPlace? {
        guard let spec = orderedSpecs.first(where: { $0.matches(location) }) else {
            return nil
        }
        return ResolvedPlace(
            placeID: spec.id,
            presentation: spec.presentation,
            queryKeys: spec.queryKeys(project, location),
            deepId: spec.deepId(location)
        )
    }

    func allPlaceIDs() -> [PlaceID] {
        specs.map(\.id)
    }
}
