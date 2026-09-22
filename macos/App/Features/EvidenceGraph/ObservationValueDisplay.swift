import Foundation

/// User-visible Observation value strings for graph cards and other list rows.
/// Handles every Property value_type so structured date/name/term/subject
/// Observations never collapse to an empty “—” when value_text is unset in SQLite.
enum ObservationValueDisplay {
    static func string(
        for observation: CatalogObservation,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        // Prefer structured DateValue when present (locale-aware).
        if let date = observation.date {
            let formatted = DateValueDisplay.string(for: date, locale: locale)
            if !formatted.isEmpty {
                return formatted
            }
        }

        let text = observation.valueText.trimmingCharacters(in: .whitespacesAndNewlines)
        if !text.isEmpty {
            return text
        }

        let name = observation.nameForm.trimmingCharacters(in: .whitespacesAndNewlines)
        if !name.isEmpty {
            return name
        }

        if let value = observation.valueInteger {
            return "\(value)"
        }

        return ""
    }
}
