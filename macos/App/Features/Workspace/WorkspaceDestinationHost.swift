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
        case .sourceCitationComposer:
            if let entry = CitationComposerEntry(location: navigation.currentLocation) {
                CitationComposerView(
                    entry: entry,
                    session: session,
                    store: store,
                    userID: userID
                )
                .id(entry.identityKey)
            }
        case .sourcePromote:
            if let entry = PromoteEntry(location: navigation.currentLocation) {
                PromoteView(
                    entry: entry,
                    session: session,
                    store: store,
                    userID: userID,
                    catalogCounts: catalogCounts
                )
                .id(entry.identityKey)
            }
        case .metadata:
            MetadataView(
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
        case .properties:
            PropertiesView(
                session: session,
                userID: userID,
                store: store,
                catalogCounts: catalogCounts
            )
            .accessibilityIdentifier("workspace.destination.properties")
        case .personsList:
            PersonsListView(session: session)
                .accessibilityIdentifier("workspace.destination.persons")
        case .personDetail:
            PersonDetailView(session: session, entityId: navigation.currentLocation.entityId ?? "")
                .accessibilityIdentifier("workspace.destination.personDetail")
        case .eventsList:
            EventsListView(session: session)
                .accessibilityIdentifier("workspace.destination.events")
        case .eventDetail:
            EventDetailView(session: session, entityId: navigation.currentLocation.entityId ?? "")
                .accessibilityIdentifier("workspace.destination.eventDetail")
        case .placesList:
            PlacesListView(session: session)
                .accessibilityIdentifier("workspace.destination.places")
        case .placeDetail:
            ConclusionStubView(section: .places)
                .accessibilityIdentifier("workspace.destination.placeDetail")
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
        case .sourcesList, .sourcePage, .sourceGraph, .sourceCitationComposer, .sourcePromote:
            return .sources
        case .metadata:
            return .metadata
        case .sourceTypes:
            return .sourceTypes
        case .properties:
            return .properties
        case .personsList, .personDetail:
            return .persons
        case .eventsList, .eventDetail:
            return .events
        case .placesList, .placeDetail:
            return .places
        }
    }
}

/// Top-level destination view mounted by `WorkspaceDestinationHost`.
enum WorkspaceDestinationKind: Equatable {
    case sources
    case metadata
    case sourceTypes
    case properties
    case persons
    case events
    case places
}
