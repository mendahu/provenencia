import Foundation
import Observation

/// Typed view of one cached catalog query. Owned by `WorkspaceSession` per key.
@MainActor
@Observable
final class QueryHandle<Value> {
    private(set) var status: QueryStatus = .idle
    private(set) var value: Value?
    private(set) var error: Error?
    /// True while refetching with prior `value` still visible (stale-while-revalidate).
    private(set) var isFetching = false

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
    }

    func applyFailure(_ error: Error, clearValue: Bool) {
        self.error = error
        status = .error
        isFetching = false
        if clearValue {
            value = nil
        }
    }

    func resetForInvalidation() {
        status = .idle
        value = nil
        error = nil
        isFetching = false
    }
}

@MainActor
protocol QueryHandleResetting: AnyObject {
    func resetForInvalidation()
}

extension QueryHandle: QueryHandleResetting {}
