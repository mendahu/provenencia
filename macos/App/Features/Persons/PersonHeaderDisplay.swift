import Foundation

/// Text for a Person header (S9-07). Go returns structures; this is the one
/// place a Person's row title is made: resolved name, else the handle's
/// label, else its ref.
enum PersonHeaderDisplay {
    static func title(_ header: CatalogPersonHeader) -> String {
        if let name = header.name {
            let text = NameValueDisplay.string(for: name)
            if !text.isEmpty { return text }
        }
        let label = header.entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? header.entity.ref : label
    }
}
