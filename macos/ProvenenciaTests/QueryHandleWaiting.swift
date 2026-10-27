import Foundation
@testable import Provenencia

/// Waits until a query handle has settled: ready or failed, and no refetch of a
/// stale value in flight (`status` stays `.ready` while a refetch runs).
@MainActor
func waitForQuery<Value>(
    _ handle: QueryHandle<Value>,
    timeoutNanoseconds: UInt64 = 2_000_000_000
) async {
    var waited: UInt64 = 0
    let step: UInt64 = 10_000_000
    while waited < timeoutNanoseconds {
        if (handle.status == .ready || handle.status == .error) && !handle.isFetching { return }
        await Task.yield()
        try? await Task.sleep(nanoseconds: step)
        waited += step
    }
}

/// Polls `condition` every 10 ms until it holds or `timeout` passes (two
/// seconds by default); true once it holds.
@MainActor
func waitUntil(timeout: Duration = .seconds(2), _ condition: () -> Bool) async -> Bool {
    let deadline = ContinuousClock.now + timeout
    while ContinuousClock.now < deadline {
        if condition() { return true }
        try? await Task.sleep(for: .milliseconds(10))
    }
    return condition()
}

/// Waits until a catalog search has settled by awaiting the search itself. No
/// deadline: a full-suite run keeps the main actor busy for longer than any
/// fixed timeout the debounce could be given (CI runs have taken 17 seconds).
@MainActor
func waitForSearch(_ model: CatalogSearchModel) async {
    await model.settle()
}
