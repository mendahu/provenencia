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

/// Type-to-select buffering for lists and menus. Prefix matching is testable
/// without mounting a view. Same contract as a native macOS popup / table.
struct PVTypeSelectMatcher: Equatable {
    /// The buffer starts over after this long a pause.
    static let resetInterval: TimeInterval = 0.8

    private var buffer = ""
    private var lastKeystroke = Date.distantPast

    init() {}

    /// Appends a character, resetting the buffer first if the user paused.
    mutating func append(_ character: Character, now: Date = Date()) -> String {
        buffer = now.timeIntervalSince(lastKeystroke) > Self.resetInterval
            ? String(character)
            : buffer + String(character)
        lastKeystroke = now
        return buffer
    }

    mutating func reset() {
        buffer = ""
        lastKeystroke = .distantPast
    }

    /// First index whose value starts with `prefix`, searching forward from
    /// `fromIndex` and wrapping, so repeated single keystrokes cycle matches.
    static func index(in values: [String], prefix: String, fromIndex: Int = 0) -> Int {
        let query = prefix.lowercased()
        guard !query.isEmpty, !values.isEmpty else { return -1 }
        for offset in 0..<values.count {
            let idx = (max(0, fromIndex) + offset) % values.count
            if values[idx].lowercased().hasPrefix(query) { return idx }
        }
        return -1
    }
}
