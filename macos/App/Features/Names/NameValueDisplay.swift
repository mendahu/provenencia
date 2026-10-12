import Foundation

/// Which of a profile's two orders a surface uses. Call sites choose this.
/// The part list itself comes from the catalog.
enum NameDisplayStyle: Sendable {
    case natural
    case sorted
}

/// One part type in a display order, and the separator that follows that group.
/// An empty separator is a normal space.
struct NameFormatPart: Sendable, Equatable {
    var partType: String
    var trailingSeparator: String
}

/// The default profile's two orders, loaded when a catalog opens.
/// Empty lists keep stored part order.
struct NameFormatOrders: Sendable, Equatable {
    var natural: [NameFormatPart] = []
    var sorted: [NameFormatPart] = []

    func order(_ style: NameDisplayStyle) -> [NameFormatPart] {
        switch style {
        case .natural: natural
        case .sorted: sorted
        }
    }
}

/// Compact NameValue summary for list rows and host previews.
enum NameValueDisplay {
    static func string(for draft: NameValueDraft) -> String {
        draft.trimmedForm
    }

    static func string(form: String) -> String {
        form.trimmingCharacters(in: .whitespacesAndNewlines)
    }

    /// Display text for a structured NameValue. `order` is the profile's
    /// part list for `style`. An empty order joins parts in stored order.
    /// A part type the order does not list follows, still in stored order.
    /// No parts falls back to `form`. `form` is never rewritten.
    static func string(
        for name: CatalogNameValue,
        order: [NameFormatPart] = []
    ) -> String {
        if name.parts.isEmpty {
            return string(form: name.form)
        }
        let formatted = join(name.parts, order: order)
        if !formatted.isEmpty { return formatted }
        return string(form: name.form)
    }

    static func string(
        for name: CatalogNameValue,
        style: NameDisplayStyle,
        orders: NameFormatOrders
    ) -> String {
        string(for: name, order: orders.order(style))
    }

    private static func join(_ parts: [CatalogNameValuePart], order: [NameFormatPart]) -> String {
        var byType: [String: [String]] = [:]
        var stored: [String] = []
        var seen = Set<String>()
        for part in parts {
            let word = part.value.trimmingCharacters(in: .whitespacesAndNewlines)
            guard !word.isEmpty else { continue }
            if byType[part.type] == nil {
                if !seen.contains(part.type) {
                    seen.insert(part.type)
                    stored.append(part.type)
                }
            }
            byType[part.type, default: []].append(word)
        }
        var listed = Set<String>()
        var groups: [(text: String, trail: String)] = []
        func add(_ type: String, trail: String) {
            guard let words = byType[type], !words.isEmpty else { return }
            groups.append((words.joined(separator: " "), trail))
        }
        for row in order {
            listed.insert(row.partType)
            add(row.partType, trail: row.trailingSeparator)
        }
        for type in stored where !listed.contains(type) {
            add(type, trail: "")
        }
        guard !groups.isEmpty else { return "" }
        var out = ""
        for (index, group) in groups.enumerated() {
            if index > 0 {
                let trail = groups[index - 1].trail
                out += trail.isEmpty ? " " : trail
            }
            out += group.text
        }
        return out
    }
}
