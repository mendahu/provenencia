import Foundation

/// Stable identity for one registered workspace place.
enum PlaceID: Hashable, Sendable, CaseIterable {
    case sourcesList
    case sourceDetail
    case sourceFields
    case sourceTypes
    case sourceTypesDetail
}
