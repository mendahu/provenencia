import Foundation

/// View-routing identity for `WorkspaceDestinationHost` (S4-05+).
enum WorkspacePresentationID: Hashable, Sendable, CaseIterable {
    case sourcesList
    case sourcePage
    case sourceGraph
    case sourceFields
    case sourceTypes
    case subjectFields
}
