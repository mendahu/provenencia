import Foundation

/// View-routing identity for `WorkspaceDestinationHost` (S4-05+).
enum WorkspacePresentationID: Hashable, Sendable, CaseIterable {
    case sourcesList
    case sourcePage
    case sourceGraph
    case sourceCitationComposer
    case metadata
    case sourceTypes
    case properties
    case personsList
    case eventsList
    case placesList
}
