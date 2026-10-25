import Foundation

/// Resolves a Property term’s display name: product L10n first, then catalog `label`.
/// Only a product term (`CatalogOrigin.provenencia`) reads from L10n; a researcher's
/// term keyed like one stays on its own label.
enum PropertyTermDisplay {
    static func name(key: String, origin: String, propertyKey: String, catalogLabel: String) -> String {
        if CatalogOrigin.isProduct(origin),
           let resource = L10n.PropertyTerm.resource(propertyKey: propertyKey, termKey: key) {
            return L10n.string(resource)
        }
        let trimmed = catalogLabel.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? key : trimmed
    }

    static func name(term: CatalogPropertyTerm, propertyKey: String) -> String {
        name(key: term.key, origin: term.origin, propertyKey: propertyKey, catalogLabel: term.label)
    }
}
