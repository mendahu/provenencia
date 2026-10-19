import Foundation
@testable import Provenencia

/// Waits until a query handle has settled: ready or failed, and no refetch of a
/// stale value in flight (`status` stays `.ready` while a refetch runs).
@MainActor
func waitForQuery<Value>(_ handle: QueryHandle<Value>) async {
    var waited: UInt64 = 0
    let step: UInt64 = 10_000_000
    while waited < 2_000_000_000 {
        if (handle.status == .ready || handle.status == .error) && !handle.isFetching { return }
        await Task.yield()
        try? await Task.sleep(nanoseconds: step)
        waited += step
    }
}
