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
            id: .sourceDetail,
            presentation: .sourcePage,
            priority: 100,
            matches: { $0.section == .sources && $0.sourceId != nil },
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
            id: .sourceFields,
            presentation: .sourceFields,
            priority: 80,
            matches: { $0.section == .sourceFields },
            queryKeys: { project, _ in
                [.metadataFieldsList(project: project)]
            },
            deepId: { $0.fieldId }
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
