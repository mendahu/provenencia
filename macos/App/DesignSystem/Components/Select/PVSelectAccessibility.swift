import Foundation

/// Spoken / trait mapping for `PVSelect`. The trigger is the VoiceOver surface
/// (activedescendant); the menu panel is not a second tab stop.
enum PVSelectAccessibility {
    struct Spoken: Equatable {
        var value: String
        var isExpanded: Bool
        var position: (current: Int, count: Int)?

        static func == (lhs: Spoken, rhs: Spoken) -> Bool {
            lhs.value == rhs.value
                && lhs.isExpanded == rhs.isExpanded
                && lhs.position?.current == rhs.position?.current
                && lhs.position?.count == rhs.position?.count
        }
    }

    static func identifier(_ prefix: String?, suffix: String? = nil) -> String? {
        guard let prefix, !prefix.isEmpty else { return nil }
        guard let suffix, !suffix.isEmpty else { return prefix }
        return "\(prefix).\(suffix)"
    }

    static func menuIdentifier(_ prefix: String?) -> String? {
        identifier(prefix, suffix: "menu")
    }

    static func spoken(
        session: PVSelectSession,
        placeholder: String? = nil,
        displayLabel: String? = nil
    ) -> Spoken {
        if session.isOpen {
            let value = session.highlightedLabel
                ?? session.committedLabel
                ?? placeholder
                ?? ""
            let position: (Int, Int)?
            if session.options.indices.contains(session.highlightIndex) {
                position = (session.highlightIndex + 1, session.options.count)
            } else {
                position = nil
            }
            return Spoken(value: value, isExpanded: true, position: position)
        }
        if let displayLabel {
            return Spoken(value: displayLabel, isExpanded: false, position: nil)
        }
        let value = session.committedLabel ?? placeholder ?? ""
        return Spoken(value: value, isExpanded: false, position: nil)
    }

    static func triggerLabel(
        session: PVSelectSession,
        displayLabel: String?,
        placeholder: String?
    ) -> String {
        if let displayLabel { return displayLabel }
        return session.committedLabel ?? placeholder ?? ""
    }

    static func hasRequiredLabel(iconOnly: Bool, hasLabel: Bool) -> Bool {
        !iconOnly || hasLabel
    }
}

/// Visual / AT split for one menu row. Checkmark is the committed mark;
/// fill is the keyboard or pointer highlight. They must not look equally selected.
enum PVSelectRowChrome {
    struct Appearance: Equatable {
        var showsCheckmark: Bool
        var fillsHighlight: Bool
        var isSelectedTrait: Bool
    }

    static func appearance(committed: Bool, highlighted: Bool) -> Appearance {
        Appearance(
            showsCheckmark: committed,
            fillsHighlight: highlighted,
            isSelectedTrait: committed
        )
    }
}
