import Foundation

/// Text for a Person summary (S9-07). Go returns structures; this is the one
/// place a Person's row title is made: auto-reconciled name, else the handle's
/// label, else its ref. `titleSource` says which, so a row can style each
/// case (S9-D2: name plain, label italic, ref mono).
enum PersonTitleDisplay {
    static func titleSource(
        name: CatalogNameValue?,
        entity: CatalogCanonicalEntity,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> ConclusionTitleSource {
        if let name {
            let d = NameValueDisplay.display(for: name, style: style, orders: orders)
            if !d.text.isEmpty { return .name(d.text, styled: d.styled, spoken: d.spoken) }
        }
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? .ref(entity.ref) : .label(label)
    }

    static func titleSource(
        _ summary: CatalogPersonSummary,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> ConclusionTitleSource {
        titleSource(name: summary.name, entity: summary.entity, style: style, orders: orders)
    }

    static func title(
        _ summary: CatalogPersonSummary,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        titleSource(summary, style: style, orders: orders).text
    }
}
