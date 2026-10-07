import Foundation

/// A date as a list row or page summary shows it: compact genealogical form
/// (`14 May 1817`, `abt 1810`, `bet 1803 and 1806`), a point date over a
/// start–end span. Editors keep `DateValueDisplay`.
enum DateRowDisplay {
    /// A point date wins; otherwise the start–end span. Empty when none of
    /// them is set.
    static func line(
        date: CatalogDateValueInput?,
        start: CatalogDateValueInput? = nil,
        end: CatalogDateValueInput? = nil,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        if let date {
            return rowDate(date, locale: locale)
        }
        let startText = start.map { rowDate($0, locale: locale) } ?? ""
        let endText = end.map { rowDate($0, locale: locale) } ?? ""
        switch (startText.isEmpty, endText.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return startText
        case (true, false):
            return endText
        case (false, false):
            return L10n.Dates.rowSpan(start: startText, end: endText, locale: locale)
        }
    }

    /// One date as the list shows it: day–month–year, qualifier abbreviations,
    /// and `bet … and …` for a range.
    private static func rowDate(_ input: CatalogDateValueInput, locale: Locale) -> String {
        let draft = DateValueDraft(from: input)
        if draft.isRange {
            guard
                let start = civil(
                    year: draft.startYear, month: draft.startMonth, day: draft.startDay, locale: locale
                ),
                let end = civil(
                    year: draft.endYear, month: draft.endMonth, day: draft.endDay, locale: locale
                )
            else {
                return ""
            }
            return appendPhrase(
                L10n.Dates.rowBetween(start: start, end: end, locale: locale),
                phrase: draft.phrase
            )
        }
        guard let point = civil(
            year: draft.startYear, month: draft.startMonth, day: draft.startDay, locale: locale
        ) else {
            return DateValueDisplay.string(for: input, locale: locale)
        }
        return appendPhrase(applyQualifier(draft.qualifier, to: point, locale: locale), phrase: draft.phrase)
    }

    private static func applyQualifier(_ raw: String, to point: String, locale: Locale) -> String {
        switch raw.trimmingCharacters(in: .whitespacesAndNewlines) {
        case DateQualifier.about:
            return L10n.Dates.rowAbout(point, locale: locale)
        case DateQualifier.before:
            return L10n.Dates.rowBefore(point, locale: locale)
        case DateQualifier.after:
            return L10n.Dates.rowAfter(point, locale: locale)
        default:
            return point
        }
    }

    /// Day, then the locale's abbreviated month, then the year. A missing
    /// piece is left out (`14 May 1817`, `Mar 1790`, `1851`).
    private static func civil(year: Int32?, month: Int32?, day: Int32?, locale: Locale) -> String? {
        let monthName: String? = {
            guard let month, (1...12).contains(month) else { return nil }
            var calendar = Calendar(identifier: .gregorian)
            calendar.locale = locale
            let symbols = calendar.shortMonthSymbols
            let index = Int(month) - 1
            guard symbols.indices.contains(index) else { return nil }
            return symbols[index]
        }()
        switch (day, monthName, year) {
        case let (day?, month?, year?):
            return "\(day) \(month) \(year)"
        case let (nil, month?, year?):
            return "\(month) \(year)"
        case let (day?, nil, year?):
            return "\(day) \(year)"
        case let (nil, nil, year?):
            return "\(year)"
        case let (day?, month?, nil):
            return "\(day) \(month)"
        case let (nil, month?, nil):
            return month
        case let (day?, nil, nil):
            return "\(day)"
        case (nil, nil, nil):
            return nil
        }
    }

    private static func appendPhrase(_ structured: String, phrase: String) -> String {
        let trimmed = phrase.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return structured }
        return "\(structured) · \(trimmed)"
    }
}
