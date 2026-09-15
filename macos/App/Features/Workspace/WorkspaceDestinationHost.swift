import SwiftUI

/// Routes workspace content by place-registry presentation id (S4-05 shim).
/// Reads `navigation.currentLocation`; S4-06+ splits Sources list/detail here.
struct WorkspaceDestinationHost: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    let projectDir: String
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore
    let catalogCounts: CatalogCounts

    var body: some View {
        switch Self.presentation(for: navigation.currentLocation, projectDir: projectDir) {
        case .sourcesList, .sourcePage:
            SourcesView(
                projectDir: projectDir,
                userID: userID,
                sessionDisplayName: sessionDisplayName,
                store: store,
                catalogCounts: catalogCounts
            )
        case .sourceFields:
            SourceFieldsView(
                projectDir: projectDir,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        case .sourceTypes:
            SourceTypesView(
                projectDir: projectDir,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        }
    }

    static func presentation(
        for location: WorkspaceLocation,
        projectDir: String,
        registry: PlaceRegistry = .standard
    ) -> WorkspacePresentationID {
        let project = ProjectKey(projectDir: projectDir)
        return registry.resolve(location, project: project)?.presentation ?? .sourcesList
    }

    /// Mirrors the host `switch` for unit tests (which view family mounts).
    static func destinationKind(for presentation: WorkspacePresentationID) -> WorkspaceDestinationKind {
        switch presentation {
        case .sourcesList, .sourcePage:
            return .sources
        case .sourceFields:
            return .sourceFields
        case .sourceTypes:
            return .sourceTypes
        }
    }
}

/// Top-level destination view mounted by `WorkspaceDestinationHost`.
enum WorkspaceDestinationKind: Equatable {
    case sources
    case sourceFields
    case sourceTypes
}
