import Foundation

/// A comparison's user-visible value. A structured date uses DateValueDisplay
/// so the sheet follows the researcher's locale. Anything else keeps the
/// portable string the core sent.
enum ComparisonValueDisplay {
    static func string(date: CatalogDateValueInput?, fallback: String, locale: Locale = .autoupdatingCurrent) -> String {
        if let date {
            let formatted = DateValueDisplay.string(for: date, locale: locale)
            if !formatted.isEmpty {
                return formatted
            }
        }
        return fallback
    }
}
