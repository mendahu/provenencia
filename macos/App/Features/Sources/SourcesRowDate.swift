import SwiftUI

/// The text a row says about the Source's date, in one place: a date whose line is
/// empty counts as undated.
enum SourcesRowDate {
    static func line(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String? {
        guard let date else { return nil }
        let line = DateRowDisplay.line(date: date, locale: locale)
        return line.isEmpty ? nil : line
    }

    static func help(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String? {
        line(date, locale: locale).map(L10n.Sources.publishedHelp(date:))
    }

    static func spoken(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String {
        line(date, locale: locale).map(L10n.Sources.publishedSpoken(date:))
            ?? L10n.string(L10n.Sources.noPublishedDateSpoken)
    }
}
