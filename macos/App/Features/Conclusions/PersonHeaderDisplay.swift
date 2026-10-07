import Foundation

/// Text for a Person header (S9-07). Go returns structures; this is the one
/// place a Person's row title is made: auto-reconciled name, else the handle's
/// label, else its ref. `titleSource` says which, so a row can style each
/// case (S9-D2: name plain, label italic, ref mono).
enum PersonHeaderDisplay {
    typealias TitleSource = ConclusionTitleSource

    static func titleSource(name: CatalogNameValue?, entity: CatalogCanonicalEntity) -> TitleSource {
        if let name {
            let text = NameValueDisplay.string(for: name)
            if !text.isEmpty { return .name(text) }
        }
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? .ref(entity.ref) : .label(label)
    }

    static func titleSource(_ header: CatalogPersonHeader) -> TitleSource {
        titleSource(name: header.name, entity: header.entity)
    }

    static func title(_ header: CatalogPersonHeader) -> String {
        titleSource(header).text
    }
}
