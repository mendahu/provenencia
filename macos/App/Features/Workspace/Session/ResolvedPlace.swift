import Foundation

/// One workspace place resolved from navigation location + registry row.
struct ResolvedPlace: Equatable, Sendable {
    let placeID: PlaceID
    let presentation: WorkspacePresentationID
    let queryKeys: [CatalogQueryKey]
    /// The id this place's view resets on: a different value remounts the destination. `nil` keeps one view across locations (list roots, master–detail places that follow `currentLocation`).
    let viewIdentity: String?
}
