import Foundation

/// Resolves a Property term’s display name: product L10n first, then catalog `label`.
/// User-minted terms have no L10n key and stay on the catalog label.
enum PropertyTermDisplay {
    static func name(key: String, propertyKey: String, catalogLabel: String) -> String {
        if let resource = L10n.PropertyTerm.resource(propertyKey: propertyKey, termKey: key) {
            return String(localized: resource)
        }
        let trimmed = catalogLabel.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? key : trimmed
    }

    static func name(term: CatalogPropertyTerm, propertyKey: String) -> String {
        name(key: term.key, propertyKey: propertyKey, catalogLabel: term.label)
    }
}
