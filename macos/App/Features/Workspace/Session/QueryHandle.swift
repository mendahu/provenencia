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
}

/// Typed view of one cached catalog query. Owned by `WorkspaceSession` per key.
@MainActor
@Observable
final class QueryHandle<Value> {
    private(set) var status: QueryStatus = .idle
    private(set) var value: Value?
    private(set) var error: Error?
    /// True while refetching with prior `value` still visible (stale-while-revalidate).
    private(set) var isFetching = false
    /// Bumps when `value` is replaced or cleared. Readers compare this instead
    /// of the stored rows.
    private(set) var revision = 0

    /// Loaded and not refetching: `value` is the latest the loader returned.
    var isSettled: Bool { status == .ready && !isFetching }

    /// The last load's error while the handle is in the error state, even beside a stale value.
    var failure: Error? { status == .error ? error : nil }

    /// Folds value / error / loading: a value wins (even a stale one), else a
    /// failure, else still loading.
    var read: QueryRead<Value> {
        if let value { return .loaded(value) }
        if status == .error, let error { return .failed(error) }
        return .loading
    }

    init() {}

    func beginLoading(stale: Bool) {
        if stale {
            isFetching = true
        } else {
            status = .loading
        }
    }

    func applySuccess(_ value: Value) {
        self.value = value
        error = nil
        status = .ready
        isFetching = false
        revision += 1
    }

    func applyFailure(_ error: Error, clearValue: Bool) {
        self.error = error
        status = .error
        isFetching = false
        if clearValue {
            value = nil
            revision += 1
        }
    }

    func resetForInvalidation() {
        status = .idle
        value = nil
        error = nil
        isFetching = false
        revision += 1
    }
}

@MainActor
protocol QueryHandleResetting: AnyObject {
    func resetForInvalidation()
}

extension QueryHandle: QueryHandleResetting {}
