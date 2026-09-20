import Foundation

/// Clamped ↑/↓ movement for overlay floating menus (context, select, history).
/// No wrap — matches Finder lists and the omnibar.
enum PVFloatingMenuSelection {
    /// - Parameters:
    ///   - index: Current highlight (`-1` = none).
    ///   - delta: Typically `±1`.
    ///   - count: Number of activatable rows.
    /// - Returns: New index, or `-1` when `count` is 0.
    static func moveIndex(from index: Int, delta: Int, count: Int) -> Int {
        guard count > 0 else { return -1 }
        if index < 0 { return delta > 0 ? 0 : count - 1 }
        return min(count - 1, max(0, index + delta))
    }
}
