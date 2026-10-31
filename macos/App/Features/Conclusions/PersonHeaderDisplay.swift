import Foundation

/// Text for a Person header (S9-07). Go returns structures; this is the one
/// place a Person's row title is made: auto-reconciled name, else the handle's
/// label, else its ref. `titleSource` says which, so a row can style each
/// case (S9-D2: name plain, label italic, ref mono).
enum PersonHeaderDisplay {
    typealias TitleSource = ConclusionTitleSource

    static func titleSource(
        name: CatalogNameValue?,
        entity: CatalogCanonicalEntity,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> TitleSource {
        if let name {
            let d = NameValueDisplay.display(for: name, style: style, orders: orders)
            if !d.text.isEmpty { return .name(d.text, styled: d.styled) }
        }
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? .ref(entity.ref) : .label(label)
    }

    static func titleSource(
        _ header: CatalogPersonHeader,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> TitleSource {
        titleSource(name: header.name, entity: header.entity, style: style, orders: orders)
    }

    static func title(
        _ header: CatalogPersonHeader,
        style: NameDisplayStyle = .natural,
        orders: NameFormatOrders = NameFormatOrders()
    ) -> String {
        titleSource(header, style: style, orders: orders).text
    }
}
