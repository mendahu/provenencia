import Foundation

/// A scope's newest audit revision and when it was recorded (RFC3339 UTC).
/// Mirrors the engine's `graphcache.Activity`: revision 0 means none. Ordered
/// by revision, never by the time string (the wall clock is not monotonic).
struct CatalogActivity: Sendable, Equatable {
    var revision: Int64 = 0
    var at: String = ""
}

/// A cached list row that carries its handle's or Source's last activity.
protocol CatalogLastUpdatedRow {
    var id: String { get }
    var updated: CatalogActivity { get set }
}

extension CatalogPersonHeader: CatalogLastUpdatedRow {}
extension CatalogEventHeader: CatalogLastUpdatedRow {}
extension CatalogPlaceHeader: CatalogLastUpdatedRow {}
extension CatalogSource: CatalogLastUpdatedRow {}

/// Which ids of a write's effects a list's rows are keyed by.
enum CatalogLastUpdatedScope: Sendable {
    /// `CatalogWriteEffects.handleScopes`: Person, Event, and Place handles.
    case handle
    /// `CatalogWriteEffects.sources`.
    case source

    /// The ids of a write's effects that name rows of this scope.
    func ids(in effects: CatalogWriteEffects) -> [String] {
        switch self {
        case .handle: effects.handleScopes
        case .source: effects.sources
        }
    }
}

/// How one cached list moves a row's last activity when a write reports it.
/// Declared on the list's registry row, so a new list adds one line and the
/// session needs no per-kind switch.
struct CatalogLastUpdatedPatch: Sendable {
    let scope: CatalogLastUpdatedScope
    let apply: @MainActor @Sendable (WorkspaceSession, CatalogQueryKey, Set<String>, CatalogActivity) -> Void

    /// Moves each row whose id is in the set and whose revision is older than
    /// the write's. A row only moves forward; nothing is fetched.
    static func rows<Row: CatalogLastUpdatedRow>(_: Row.Type, scope: CatalogLastUpdatedScope) -> Self {
        Self(scope: scope) { session, key, ids, activity in
            session.patchCached(key, as: [Row].self) { rows in
                var moved = false
                for index in rows.indices
                where ids.contains(rows[index].id) && rows[index].updated.revision < activity.revision {
                    rows[index].updated = activity
                    moved = true
                }
                return moved
            }
        }
    }
}
