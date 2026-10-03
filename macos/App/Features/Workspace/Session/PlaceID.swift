import Foundation

/// Stable identity for one registered workspace place.
enum PlaceID: Hashable, Sendable, CaseIterable {
    case sourcesList
    case sourceDetail
    case sourceGraph
    case sourceCitationComposer
    case metadata
    case sourceTypes
    case sourceTypesDetail
    case properties
    case personsList
    case eventsList
    case placesList
}
