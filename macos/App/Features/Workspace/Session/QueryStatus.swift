import Foundation

/// Load state for one cached catalog query.
enum QueryStatus: Equatable, Sendable {
    case idle
    case loading
    case ready
    case error
}
