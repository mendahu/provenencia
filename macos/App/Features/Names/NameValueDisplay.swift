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
    /// A nickname shows in quotes and a former surname in parentheses.
    /// No parts falls back to `form`. `form` is never rewritten.
    static func string(
        for name: CatalogNameValue,
        order: [NameFormatPart] = []
    ) -> String {
        display(for: name, order: order).text
    }

    static func string(
        for name: CatalogNameValue,
        style: NameDisplayStyle,
        orders: NameFormatOrders
    ) -> String {
        string(for: name, order: orders.order(style))
    }

    /// The text of ``string(for:style:orders:)``, plus a styled copy when a word is
    /// emphasized (a former surname). The styled copy is nil otherwise, so a name
    /// with no emphasized word builds no `AttributedString`.
    static func display(
        for name: CatalogNameValue,
        style: NameDisplayStyle,
        orders: NameFormatOrders
    ) -> NameDisplay {
        display(for: name, order: orders.order(style))
    }

    private static func display(for name: CatalogNameValue, order: [NameFormatPart]) -> NameDisplay {
        let runs = runs(name.parts, order: order)
        guard !runs.isEmpty else { return NameDisplay(text: string(form: name.form), styled: nil) }
        let text = runs.map(\.plainText).joined()
        guard runs.contains(where: { $0.decoration?.emphasized == true }) else {
            return NameDisplay(text: text, styled: nil)
        }
        var styled = AttributedString()
        for run in runs { styled += run.styledText }
        return NameDisplay(text: text, styled: styled)
    }

    /// The same order and separators as Go `nameformats.Format`; only this side decorates.
    /// One run per part word, one undecorated run for each space between words and each
    /// separator between groups. An empty separator is a run of one space.
    private static func runs(_ parts: [CatalogNameValuePart], order: [NameFormatPart]) -> [NameDisplayRun] {
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
        var groups: [(type: String, words: [String], trail: String)] = []
        func add(_ type: String, trail: String) {
            guard let words = byType[type], !words.isEmpty else { return }
            groups.append((type, words, trail))
        }
        for row in order {
            listed.insert(row.partType)
            add(row.partType, trail: row.trailingSeparator)
        }
        for type in stored where !listed.contains(type) {
            add(type, trail: "")
        }
        var out: [NameDisplayRun] = []
        for (index, group) in groups.enumerated() {
            if index > 0 {
                let trail = groups[index - 1].trail
                out.append(NameDisplayRun(text: trail.isEmpty ? " " : trail, decoration: nil))
            }
            let decoration = NamePartType(rawValue: group.type)?.decoration
            for (wordIndex, word) in group.words.enumerated() {
                if wordIndex > 0 { out.append(NameDisplayRun(text: " ", decoration: nil)) }
                out.append(NameDisplayRun(text: word, decoration: decoration))
            }
        }
        return out
    }
}

/// A displayed name: the plain text, and a styled copy when a word is emphasized.
struct NameDisplay: Equatable {
    var text: String
    var styled: AttributedString?
}

/// One word or separator of a displayed name, with the decoration of its part type.
private struct NameDisplayRun {
    var text: String
    var decoration: NamePartDecoration?

    /// The run with its decoration's punctuation, as plain text.
    var plainText: String {
        guard let decoration else { return text }
        return String(StyledFormat.attributed(decoration.format, [StyledFormat.Argument(text: text)]).characters)
    }

    /// The run with its decoration's punctuation; an emphasized word is marked.
    var styledText: AttributedString {
        guard let decoration else { return AttributedString(text) }
        var attributes = AttributeContainer()
        if decoration.emphasized { attributes.inlinePresentationIntent = .emphasized }
        return StyledFormat.attributed(decoration.format, [StyledFormat.Argument(text: text, attributes: attributes)])
    }
}
