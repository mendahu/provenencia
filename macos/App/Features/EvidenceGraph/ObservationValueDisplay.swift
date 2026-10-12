import Foundation

/// User-visible Observation value strings for graph cards and other list rows.
/// Handles every Property value_type so structured date/name/term/subject
/// Observations never collapse to an empty “—” when value_text is unset in SQLite.
enum ObservationValueDisplay {
    static func string(
        for observation: CatalogObservation,
        locale: Locale = .autoupdatingCurrent,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        // Prefer a structured NameValue over the denormalized form in value_text.
        if let name = structuredName(observation, orders: orders) {
            return name
        }

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

        if let value = observation.valueInteger {
            return "\(value)"
        }

        return ""
    }

    /// NameValueDisplay joins parts in the profile order. No parts falls back
    /// to the stored form. Nil when this row has no name.
    private static func structuredName(
        _ observation: CatalogObservation,
        orders: NameFormatOrders
    ) -> String? {
        let form = observation.nameForm.trimmingCharacters(in: .whitespacesAndNewlines)
        let hasPart = observation.nameParts.contains {
            !$0.value.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        }
        let isName = observation.propertyValueType == PropertyValueType.name.rawValue
        guard isName || hasPart || !form.isEmpty else { return nil }
        let rendered = NameValueDisplay.string(
            for: CatalogNameValue(form: observation.nameForm, parts: observation.nameParts),
            style: .natural,
            orders: orders
        )
        return rendered.isEmpty ? nil : rendered
    }
}
