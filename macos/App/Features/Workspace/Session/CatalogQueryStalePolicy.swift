import Foundation

/// When a cached query may be refetched without an explicit invalidation.
enum CatalogQueryStalePolicy: Sendable, Equatable {
    /// Valid until `invalidate` or session end. Default for all keys in S4-02.
    case sessionFresh
    /// Reserved: refetch when revisiting a place even if not invalidated (S4-04+).
    case refetchOnRevisit
}
