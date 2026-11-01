import Foundation

/// A scope's newest audit revision and when it was recorded (RFC3339 UTC).
/// Mirrors the engine's `effects.Activity`: revision 0 means none. Ordered
/// by revision, never by the time string (the wall clock is not monotonic).
struct CatalogActivity: Sendable, Equatable {
    var revision: Int64 = 0
    var at: String = ""
}

/// A cached list row that carries its handle's or Source's last activity.
/// `Sendable` so `CatalogQueryPatch.lastUpdated` can capture `Row.Type` in its
/// `@Sendable` closure.
protocol CatalogLastUpdatedRow: Sendable {
    var id: String { get }
    var updated: CatalogActivity { get set }
}

extension CatalogPersonHeader: CatalogLastUpdatedRow {}
extension CatalogEventHeader: CatalogLastUpdatedRow {}
extension CatalogPlaceHeader: CatalogLastUpdatedRow {}
extension CatalogSource: CatalogLastUpdatedRow {}

/// Which ids of a write a patch or refresh follows. Each slot carries its own.
enum CatalogQueryFollows: Sendable, Equatable {
    /// `CatalogWriteEffects.handleScopes`: Person, Event, and Place handles.
    case handle
    /// `CatalogWriteEffects.sourceScopes`.
    case source
    /// The key ids of one entry of `CatalogWriteEffects.reads`. A read also
    /// follows the ids its mutation names when the effects did not
    /// (`CatalogMutation.readIDs(for:)`: Promote's graph, a deleted Source;
    /// S14-16 removes these special cases).
    case reads(CatalogRead)
    /// `CatalogWriteEffects.touchedHandles`: every handle a write changed or
    /// whose header it changed.
    case touchedHandles
    /// `CatalogWriteEffects.evidenceGraphs`.
    case evidenceGraphs

    /// The ids of a write that this slot follows.
    func ids(in effects: CatalogWriteEffects, mutation: CatalogMutation?) -> [String] {
        switch self {
        case .handle:
            return effects.handleScopes
        case .source:
            return effects.sourceScopes
        case .reads(let read):
            var ids = effects.reads.first { $0.read == read }?.ids ?? []
            for id in mutation?.readIDs(for: read) ?? [] where !ids.contains(id) {
                ids.append(id)
            }
            return ids
        case .touchedHandles:
            return effects.touchedHandles
        case .evidenceGraphs:
            return effects.evidenceGraphs
        }
    }
}

/// One in-place update of a warmed cache key. `follows == nil` means the patch
/// reads the mutation (an enriched Source); otherwise the session passes the
/// ids that `follows` names. Array order on a descriptor is composition order.
struct CatalogQueryPatch: Sendable {
    var follows: CatalogQueryFollows?
    let apply: @MainActor @Sendable (
        WorkspaceSession, CatalogQueryKey, Set<String>, CatalogActivity, CatalogMutation?
    ) -> CatalogPatchResult

    /// Moves each row whose id is in the set and whose revision is older than
    /// the write's. A row only moves forward; nothing is fetched.
    static func lastUpdated<Row: CatalogLastUpdatedRow>(
        _: Row.Type,
        follows: CatalogQueryFollows
    ) -> Self {
        Self(follows: follows) { session, key, ids, activity, _ in
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

    /// Puts the mutation's Source into the Sources list: it replaces the row
    /// with the same id, and a `.createdSource` appends its row when the list
    /// lacks it (the list sorts for display). An update or type change of a
    /// row the list lacks adds nothing. Leaves any stale flag: other rows may
    /// still owe a reload.
    static func sourceRow() -> Self {
        Self(follows: nil) { session, key, _, _, mutation in
            guard let source = mutation?.patchedSource else { return .unchanged }
            var isCreate = false
            if case .createdSource = mutation { isCreate = true }
            return session.patchCached(key, as: [CatalogSource].self) { sources in
                if let index = sources.firstIndex(where: { $0.id == source.id }) {
                    sources[index] = source
                    return true
                }
                guard isCreate else { return false }
                sources.append(source)
                return true
            }
        }
    }

    /// Writes the mutation's Source onto the cached page shell. A partial
    /// patch: a page marked stale stays stale and reloads on its next visit.
    static func updatedSourceShell() -> Self {
        Self(follows: nil) { session, key, _, _, mutation in
            guard let source = mutation?.patchedSource else { return .unchanged }
            guard case .sourceWorkspace(_, let sourceId) = key, sourceId == source.id else { return .unchanged }
            return session.patchCached(key, as: CatalogSourceWorkspace.self) { workspace in
                workspace.source = source
                return true
            }
        }
    }
}

/// What a warmed key does when a write reports ids its `follows` slot names.
/// The session reads a descriptor's list in order.
struct CatalogQueryRefresh: Sendable {
    typealias Run = @MainActor @Sendable (WorkspaceSession, CatalogQueryKey, [String]) async -> Void

    enum Action: Sendable {
        /// Re-applies the hidden policy to the key, synchronously and at most
        /// once per `apply`, when the write names an id the key depends on.
        /// The closure returns those ids, read from the key or its cached
        /// value; `nil` means any id the slot names.
        case revalidate(@MainActor @Sendable (WorkspaceSession, CatalogQueryKey) -> Set<String>?)
        /// A partial reload merged into the cached value, run in a task with
        /// the ids `follows` names.
        case run(Run)
    }

    var follows: CatalogQueryFollows
    var action: Action

    init(follows: CatalogQueryFollows, action: Action) {
        self.follows = follows
        self.action = action
    }

    /// A partial reload (`.run`). `merged` is the only way to declare one.
    fileprivate init(follows: CatalogQueryFollows, run: @escaping Run) {
        self.init(follows: follows, action: .run(run))
    }

    /// A partial reload of a warmed value: `read` fetches one id's part and
    /// `merge` folds it into the cached value (`WorkspaceSession.refreshCached`).
    /// The key's `isFetching` stays true while it runs; a failed read reloads
    /// the key through its hidden policy.
    static func merged<Value: Sendable, Part: Sendable>(
        follows: CatalogQueryFollows,
        as _: Value.Type,
        read: @escaping @MainActor @Sendable (WorkspaceSession, CatalogQueryKey, String) async throws -> Part,
        merge: @escaping @MainActor @Sendable (inout Value, String, Part) -> Void
    ) -> CatalogQueryRefresh {
        CatalogQueryRefresh(follows: follows, run: { session, key, ids in
            await session.refreshCached(
                key,
                as: Value.self,
                ids: ids,
                read: { id in try await read(session, key, id) },
                merge: merge
            )
        })
    }

    /// A refresh that re-applies the hidden policy to a warmed key when the
    /// write names one of the `ids` the key depends on.
    static func applyHiddenPolicy(
        follows: CatalogQueryFollows,
        ids: @escaping @MainActor @Sendable (WorkspaceSession, CatalogQueryKey) -> Set<String>
    ) -> CatalogQueryRefresh {
        CatalogQueryRefresh(follows: follows, action: .revalidate { session, key in ids(session, key) })
    }

    /// A refresh whose ids come from the key's cached `Value`. Until the key
    /// holds a value its first load is in flight and its ids are unknown, so
    /// the refresh matches any id the slot names.
    static func applyHiddenPolicy<Value>(
        followsValue follows: CatalogQueryFollows,
        ids: @escaping @Sendable (Value) -> Set<String>
    ) -> CatalogQueryRefresh {
        CatalogQueryRefresh(follows: follows, action: .revalidate { session, key in
            let value: Value? = session.queryHandle(key)?.value
            return value.map(ids)
        })
    }

    /// A refresh that re-applies the hidden policy to a warmed key whenever the
    /// write names any id.
    static func applyHiddenPolicyOnAny(follows: CatalogQueryFollows) -> CatalogQueryRefresh {
        CatalogQueryRefresh(follows: follows, action: .revalidate { _, _ in nil })
    }
}
