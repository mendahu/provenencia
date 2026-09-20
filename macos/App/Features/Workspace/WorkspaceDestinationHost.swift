import SwiftUI

/// Routes workspace content by place-registry presentation id (S4-05+).
/// Reads `navigation.currentLocation`; list and detail mount as separate views.
struct WorkspaceDestinationHost: View {
    @Environment(WorkspaceNavigation.self) private var navigation
    @Environment(WorkspaceSession.self) private var session
    let userID: String
    let sessionDisplayName: String
    let store: any GenealogyStore
    let catalogCounts: CatalogCounts

    var body: some View {
        switch Self.presentation(for: navigation.currentLocation, projectDir: session.projectKey.projectDir) {
        case .sourcesList:
            SourcesListView(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        case .sourcePage:
            if let sourceID = navigation.currentLocation.sourceId {
                SourcePageView(
                    sourceID: sourceID,
                    session: session,
                    userID: userID,
                    sessionDisplayName: sessionDisplayName,
                    store: store
                )
            }
        case .sourceGraph:
            if let sourceID = navigation.currentLocation.sourceId {
                EvidenceGraphView(
                    sourceID: sourceID,
                    session: session,
                    store: store,
                    userID: userID
                )
            }
        case .sourceFields:
            SourceFieldsView(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        case .sourceTypes:
            SourceTypesView(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
        case .subjectTypes:
            WorkspaceComingSoonView(
                icon: .shapes,
                title: L10n.Workspace.subjectTypesTitle,
                message: L10n.Workspace.subjectTypesStubBody
            )
            .accessibilityIdentifier("workspace.destination.subjectTypes")
        case .subjectFields:
            SubjectFieldsView(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
            .accessibilityIdentifier("workspace.destination.subjectFields")
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
        case .sourcesList, .sourcePage, .sourceGraph:
            return .sources
        case .sourceFields:
            return .sourceFields
        case .sourceTypes:
            return .sourceTypes
        case .subjectTypes:
            return .subjectTypes
        case .subjectFields:
            return .subjectFields
        }
    }
}

/// Top-level destination view mounted by `WorkspaceDestinationHost`.
enum WorkspaceDestinationKind: Equatable {
    case sources
    case sourceFields
    case sourceTypes
    case subjectTypes
    case subjectFields
}
