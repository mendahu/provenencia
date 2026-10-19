import SwiftUI

extension View {
    /// Runs `action` when the view appears and after each `revision` bump, but
    /// only while `handle.isSettled`. A list publishes its sidebar count here:
    /// a stale refetch keeps `status == .ready`, so a reload goes `.ready →
    /// .ready` and `onChange(of: status)` would miss it.
    func onSettled<Value>(_ handle: QueryHandle<Value>, perform action: @escaping () -> Void) -> some View {
        onAppear {
            if handle.isSettled { action() }
        }
        .onChange(of: handle.revision) { _, _ in
            if handle.isSettled { action() }
        }
    }
}
