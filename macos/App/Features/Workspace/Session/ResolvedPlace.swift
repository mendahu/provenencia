import Foundation

/// One workspace place resolved from navigation location + registry row.
struct ResolvedPlace: Equatable, Sendable {
    let placeID: PlaceID
    let presentation: WorkspacePresentationID
    let queryKeys: [CatalogQueryKey]
    /// Active deep id for this place (`sourceId`, `typeId`, or `fieldId`); nil at list roots.
    let deepId: String?
}
