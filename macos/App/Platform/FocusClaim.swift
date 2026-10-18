import SwiftUI

/// Moves keyboard focus to a field that may not be ready to take it yet.
enum FocusClaim {
    /// Writes `value` to `binding` on the next main-queue turn, runs `then`,
    /// and writes it again a turn later if the binding still differs. The
    /// write happens twice because the target may not be in the view
    /// hierarchy until the following turn (a field inserted by the same
    /// commit), and a sheet that is still dismissing can swallow the first.
    @MainActor
    static func claim<V: Hashable>(
        _ binding: FocusState<V>.Binding,
        _ value: V,
        then: @escaping @MainActor () -> Void = {}
    ) {
        DispatchQueue.main.async {
            binding.wrappedValue = value
            then()
            DispatchQueue.main.async {
                if binding.wrappedValue != value { binding.wrappedValue = value }
            }
        }
    }
}
