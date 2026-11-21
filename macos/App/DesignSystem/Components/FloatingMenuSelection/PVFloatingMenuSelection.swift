import Foundation

/// Clamped ↑/↓ movement for overlay floating menus (context menus, select).
/// No wrap — matches Finder lists and the omnibar.
enum PVFloatingMenuSelection {
    /// - Parameters:
    ///   - index: Current highlight (`-1` = none).
    ///   - delta: Typically `±1`.
    ///   - count: Number of activatable rows.
    /// - Returns: New index, or `-1` when `count` is 0.
    static func moveIndex(from index: Int, delta: Int, count: Int) -> Int {
        moveIndex(from: index, delta: delta, count: count, edge: nil, isDisabled: { _ in false })
    }

    /// Which end of the rows a jump walks in from.
    enum Edge {
        case first, last
    }

    /// The same clamped move that steps over disabled rows in the direction of
    /// travel. An `edge` starts at that end and walks inward. Returns `index`
    /// unchanged when no enabled row lies that way, and `-1` when `count` is 0.
    ///
    /// `index` may be -1 (nothing active yet): a downward move lands on the
    /// first row and an upward move on the last.
    static func moveIndex(
        from index: Int,
        delta: Int,
        count: Int,
        edge: Edge? = nil,
        isDisabled: (Int) -> Bool
    ) -> Int {
        let last = count - 1
        guard last >= 0 else { return -1 }
        var i: Int
        switch edge {
        case .first: i = 0
        case .last: i = last
        case nil: i = index < 0 ? (delta > 0 ? 0 : last) : index + delta
        }
        i = min(max(0, i), last)
        let step = (edge == .last || delta < 0) ? -1 : 1
        while isDisabled(i), i + step >= 0, i + step <= last {
            i += step
        }
        return isDisabled(i) ? index : i
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

    /// One type-select keystroke: appends `character` to the buffer, then finds
    /// the match. A single keystroke starts past the current row, so pressing
    /// the same letter repeatedly cycles through its matches. Returns the hit,
    /// or -1.
    mutating func advance(
        _ character: Character,
        from current: Int,
        in titles: [String],
        isDisabled: (Int) -> Bool = { _ in false },
        now: Date = Date()
    ) -> Int {
        let buffer = append(character, now: now)
        let from = buffer.count == 1 ? current + 1 : max(0, current)
        return Self.index(in: titles, prefix: buffer, fromIndex: from, isDisabled: isDisabled)
    }

    /// The one test for a key that type-selects.
    static func isTypeSelectCharacter(_ character: Character) -> Bool {
        character.isLetter || character.isNumber
    }

    /// The character a key press type-selects: one letter or digit with neither
    /// ⌘ nor ⌃ held. Nil for anything else.
    static func typeSelectCharacter(_ characters: String?, command: Bool, control: Bool) -> Character? {
        guard let characters, characters.count == 1, let character = characters.first,
              isTypeSelectCharacter(character), !command, !control
        else { return nil }
        return character
    }

    /// First index whose value starts with `prefix`, searching forward from
    /// `fromIndex` and wrapping, so repeated single keystrokes cycle matches.
    /// Case, accents and punctuation do not count (`TextMatch`).
    static func index(in values: [String], prefix: String, fromIndex: Int = 0) -> Int {
        index(in: values, prefix: prefix, fromIndex: fromIndex, isDisabled: { _ in false })
    }

    /// As above, and a disabled index never matches.
    static func index(
        in values: [String],
        prefix: String,
        fromIndex: Int = 0,
        isDisabled: (Int) -> Bool
    ) -> Int {
        let query = TextMatch.Query(prefix)
        guard !query.isEmpty, !values.isEmpty else { return -1 }
        for offset in 0..<values.count {
            let idx = (max(0, fromIndex) + offset) % values.count
            if isDisabled(idx) { continue }
            if query.isPrefix(ofKey: TextMatch.key(values[idx])) { return idx }
        }
        return -1
    }
}
