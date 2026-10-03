import Foundation

/// Text for a Person header (S9-07). Go returns structures; this is the one
/// place a Person's row title is made: resolved name, else the handle's
/// label, else its ref. `titleSource` says which, so a row can style each
/// case (S9-D2: name plain, label italic, ref mono).
enum PersonHeaderDisplay {
    enum TitleSource: Equatable {
        case name(String)
        case label(String)
        case ref(String)

        var text: String {
            switch self {
            case .name(let text), .label(let text), .ref(let text): text
            }
        }
    }

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
