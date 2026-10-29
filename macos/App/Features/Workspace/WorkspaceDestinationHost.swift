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
        let place = Self.place(for: navigation.currentLocation, projectDir: session.projectKey.projectDir)
        Group {
            destination(for: place?.presentation ?? .sourcesList)
        }
        .id(place?.viewIdentity)
        // Fill the detail area, then measure it once: every destination reads \.pvWidthClass from here.
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .pvMeasuresWidthClass()
    }

    @ViewBuilder
    private func destination(for presentation: WorkspacePresentationID) -> some View {
        switch presentation {
        case .sourcesList:
            SourcesListView(
                session: session,
                userID: userID,
                projectUuid: navigation.projectUuid,
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
            if let sourceID = navigation.currentLocation.sourceId,
               let graphID = navigation.currentLocation.graphId
            {
                EvidenceGraphView(
                    graphID: graphID,
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
            }
        case .sourcePromote:
            if let entry = PromoteEntry(location: navigation.currentLocation) {
                PromoteView(
                    entry: entry,
                    session: session,
                    store: store,
                    userID: userID
                )
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
            PersonDetailView(session: session, entityId: navigation.currentLocation.entityId ?? "", userID: userID)
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
            PlaceDetailView(session: session, entityId: navigation.currentLocation.entityId ?? "")
                .accessibilityIdentifier("workspace.destination.placeDetail")
        }
    }

    /// The place `location` resolves to; `body` mounts its presentation and resets on its `viewIdentity`.
    static func place(
        for location: WorkspaceLocation,
        projectDir: String,
        registry: PlaceRegistry = .standard
    ) -> ResolvedPlace? {
        registry.resolve(location, project: ProjectKey(projectDir: projectDir))
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
