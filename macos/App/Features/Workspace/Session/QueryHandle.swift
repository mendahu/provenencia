import Foundation
import Observation

/// What a handle currently offers a reader: a value, a failure with no value,
/// or neither yet.
enum QueryRead<Value> {
    case loading
    case failed(Error)
    case loaded(Value)

    var isLoading: Bool {
        if case .loading = self { return true }
        return false
    }

    var isFailed: Bool {
        if case .failed = self { return true }
        return false
    }
}

/// What a reader that shows a refresh failure beside its value sees: nothing
/// yet, a failure with no value, or a value with the failure of its latest
/// reload (nil when that reload succeeded).
enum QueryReading<Value> {
    case loading
    case failed(Error)
    case loaded(Value, refreshFailure: Error?)
}

/// Typed view of one cached catalog query. Owned by `WorkspaceSession` per key.
@MainActor
@Observable
final class QueryHandle<Value> {
    private(set) var status: QueryStatus = .idle
    private(set) var value: Value?
    private(set) var error: Error?
    /// True while a load is refetching with prior `value` still visible
    /// (stale-while-revalidate), or a partial re-read of part of `value` is
    /// running (`beginRefresh()`). Readers treat both as "rows may be out of date".
    var isFetching: Bool { refetching || refreshing > 0 }
    /// True while a reload is owed: a write moved this value and nothing has
    /// reloaded it yet. The old `value` stays until the next load starts.
    private(set) var isStale = false
    private var refetching = false
    private var refreshing = 0
    /// Bumps when `value` is replaced or cleared. Readers compare this instead
    /// of the stored rows.
    private(set) var revision = 0

    /// Loaded, not refetching and not stale: `value` is the latest the loader returned.
    var isSettled: Bool { status == .ready && !isFetching && !isStale }

    /// Whether a value is cached.
    var hasValue: Bool { value != nil }

    /// The last load's error while the handle is in the error state, even beside a stale value.
    /// Nil while a retry runs; a notice with Try again reads `reading`.
    var failure: Error? { status == .error ? error : nil }

    /// Bumps in every `applyFailure`, so a notice can speak again for each new failure.
    private(set) var failureRevision = 0

    /// True while a first load or a refetch is running.
    var isReading: Bool { status == .loading || isFetching }

    /// A value with the failure of its latest reload beside it (nil when it
    /// succeeded), else the stored error, else still loading. The error
    /// holds while a retry after a first-load failure runs, so a notice with
    /// Try again stays on screen. `read` and `failure` are unchanged.
    var reading: QueryReading<Value> {
        if let value { return .loaded(value, refreshFailure: failure) }
        if let error { return .failed(error) }
        return .loading
    }

    /// Folds value / error / loading: a value wins (even a stale one), else a
    /// failure, else still loading.
    var read: QueryRead<Value> {
        if let value { return .loaded(value) }
        if status == .error, let error { return .failed(error) }
        return .loading
    }

    init() {}

    /// Starts a load, which makes the handle current. `keepingValue` says a
    /// value is already shown, so the load refetches behind it.
    func beginLoading(keepingValue: Bool) {
        isStale = false
        if keepingValue {
            refetching = true
        } else {
            status = .loading
        }
    }

    /// Flags that a reload is owed; the next `query` or place visit does it.
    func markStale() {
        isStale = true
    }

    /// Says the value is current again (a whole-value write).
    func clearStale() {
        isStale = false
    }

    /// A partial re-read of `value` starts. Leaves `status` and `isStale`.
    func beginRefresh() {
        refreshing += 1
    }

    /// A partial re-read ends. Overlapping re-reads keep `isFetching` true
    /// until the last one ends.
    func endRefresh() {
        refreshing = max(0, refreshing - 1)
    }

    func applySuccess(_ value: Value) {
        self.value = value
        error = nil
        status = .ready
        refetching = false
        revision += 1
    }

    func applyFailure(_ error: Error, clearValue: Bool) {
        self.error = error
        status = .error
        refetching = false
        failureRevision += 1
        if clearValue {
            value = nil
            revision += 1
        }
    }

    func resetForInvalidation() {
        status = .idle
        value = nil
        error = nil
        refetching = false
        revision += 1
    }
}

/// The erased view of a handle's cache state, so the session can keep handles
/// of every `Value` type in one map.
@MainActor
protocol AnyQueryHandle: AnyObject {
    var isStale: Bool { get }
    var hasValue: Bool { get }
    var isFetching: Bool { get }
    func markStale()
    func clearStale()
    func beginRefresh()
    func endRefresh()
}

extension QueryHandle: AnyQueryHandle {}
