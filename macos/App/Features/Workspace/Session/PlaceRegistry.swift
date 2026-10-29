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
        /// The id this place's view resets on: a different value remounts the destination. `nil` keeps one view across locations (list roots, master–detail places that follow `currentLocation`).
        let viewIdentity: @Sendable (WorkspaceLocation) -> String?
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
                guard location.sourceId != nil, let graphId = location.graphId else { return [] }
                return [
                    .evidenceGraph(project: project, graphId: graphId),
                    .sourcesList(project: project),
                    .sourceTypesList(project: project),
                    .propertiesWorkspace(project: project),
                    .connectRules(project: project),
                    .evidenceGraphsList(project: project),
                ]
            },
            viewIdentity: { PromoteEntry(location: $0)?.identityKey }
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
                guard let sourceId = location.sourceId, let graphId = location.graphId else { return [] }
                var keys: [CatalogQueryKey] = [
                    .evidenceGraph(project: project, graphId: graphId),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                    .sourceWorkspace(project: project, sourceId: sourceId),
                    .sourceTypesList(project: project),
                    .citationCounts(project: project, graphId: graphId),
                    .evidenceGraphsList(project: project),
                ]
                if let artifactId = location.artifactId, !artifactId.isEmpty {
                    keys.append(.citationsByArtifact(project: project, graphId: graphId, artifactId: artifactId))
                }
                return keys
            },
            viewIdentity: { CitationComposerEntry(location: $0)?.identityKey }
        ),
        Spec(
            id: .sourceGraph,
            presentation: .sourceGraph,
            priority: 110,
            matches: {
                $0.section == .sources && $0.sourceId != nil && $0.sourceSurface == .graph
            },
            queryKeys: { project, location in
                guard let sourceId = location.sourceId, let graphId = location.graphId else { return [] }
                return [
                    .evidenceGraph(project: project, graphId: graphId),
                    .propertiesWorkspace(project: project),
                    .sourcesList(project: project),
                    .connectRules(project: project),
                    .sourceWorkspace(project: project, sourceId: sourceId),
                    .citationCounts(project: project, graphId: graphId),
                    .evidenceGraphsList(project: project),
                ]
            },
            viewIdentity: { $0.graphId }
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
                    .evidenceGraphsList(project: project),
                    .evidenceGraphProgress(project: project),
                ]
            },
            viewIdentity: { $0.sourceId }
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
            viewIdentity: { _ in nil }
        ),
        Spec(
            id: .metadata,
            presentation: .metadata,
            priority: 80,
            matches: { $0.section == .metadata },
            queryKeys: { project, _ in
                [.metadataFieldsList(project: project)]
            },
            viewIdentity: { _ in nil }
        ),
        Spec(
            id: .properties,
            presentation: .properties,
            priority: 80,
            matches: { $0.section == .properties },
            queryKeys: { project, _ in
                [.propertiesWorkspace(project: project)]
            },
            viewIdentity: { _ in nil }
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
                    .evidenceGraphsList(project: project),
                    .evidenceGraphProgress(project: project),
                ]
            },
            viewIdentity: { _ in nil }
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
            viewIdentity: { _ in nil }
        ),
        Spec(
            id: .personsList,
            presentation: .personsList,
            priority: 10,
            matches: { $0.section == .persons && $0.entityId == nil },
            queryKeys: { project, _ in [.personsList(project: project)] },
            viewIdentity: { _ in nil }
        ),
        Spec(
            id: .personDetail,
            presentation: .personDetail,
            priority: 20,
            matches: { $0.section == .persons && $0.entityId != nil },
            queryKeys: { project, location in
                location.entityId.map { [.conclusionDetail(project: project, entityId: $0)] } ?? []
            },
            viewIdentity: { $0.entityId }
        ),
        Spec(
            id: .eventsList,
            presentation: .eventsList,
            priority: 10,
            matches: { $0.section == .events && $0.entityId == nil },
            queryKeys: { project, _ in [.eventsList(project: project)] },
            viewIdentity: { _ in nil }
        ),
        Spec(
            id: .eventDetail,
            presentation: .eventDetail,
            priority: 20,
            matches: { $0.section == .events && $0.entityId != nil },
            queryKeys: { project, location in
                location.entityId.map { [.conclusionDetail(project: project, entityId: $0)] } ?? []
            },
            viewIdentity: { $0.entityId }
        ),
        Spec(
            id: .placesList,
            presentation: .placesList,
            priority: 10,
            matches: { $0.section == .places && $0.entityId == nil },
            queryKeys: { project, _ in [.placesList(project: project)] },
            viewIdentity: { _ in nil }
        ),
        Spec(
            id: .placeDetail,
            presentation: .placeDetail,
            priority: 20,
            matches: { $0.section == .places && $0.entityId != nil },
            queryKeys: { project, location in
                location.entityId.map { [.conclusionDetail(project: project, entityId: $0)] } ?? []
            },
            viewIdentity: { $0.entityId }
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
            viewIdentity: spec.viewIdentity(location)
        )
    }

    func allPlaceIDs() -> [PlaceID] {
        specs.map(\.id)
    }
}
