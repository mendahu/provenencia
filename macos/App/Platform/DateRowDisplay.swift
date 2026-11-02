import Foundation

/// A date as a list row or page summary shows it: compact genealogical form
/// (`14 May 1817`, `abt 1810`, `bet 1803 and 1806`), a point date over a
/// start–end span. Editors keep `DateValueDisplay`.
enum DateRowDisplay {
    /// How much of a date a line shows. `.year` is a life span's: only the year of each
    /// side, with its qualifier ("abt 1790", "bet 1790 and 1795"); a side with no year
    /// shows nothing, and a date with only its phrase shows the phrase.
    enum Precision: Equatable, Sendable { case full, year }

    /// A point date wins; otherwise the start–end span. Empty when none of
    /// them is set. `includesPhrase: false` leaves off the words the source gave for the date
    /// ("in the spring"), for a place where the date sits inside a sentence.
    static func line(
        date: CatalogDateValueInput?,
        start: CatalogDateValueInput? = nil,
        end: CatalogDateValueInput? = nil,
        includesPhrase: Bool = true,
        precision: Precision = .full,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        if let date {
            return rowDate(date, includesPhrase: includesPhrase, precision: precision, locale: locale)
        }
        let startText = start.map { rowDate($0, includesPhrase: includesPhrase, precision: precision, locale: locale) } ?? ""
        let endText = end.map { rowDate($0, includesPhrase: includesPhrase, precision: precision, locale: locale) } ?? ""
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
    /// and `bet … and …` for a range. At `.year` each side is only its year.
    private static func rowDate(
        _ input: CatalogDateValueInput,
        includesPhrase: Bool,
        precision: Precision,
        locale: Locale
    ) -> String {
        let draft = DateValueDraft(from: input)
        let phrase = includesPhrase ? draft.phrase : ""
        if draft.isRange {
            guard
                let start = side(year: draft.startYear, month: draft.startMonth, day: draft.startDay, precision: precision, locale: locale),
                let end = side(year: draft.endYear, month: draft.endMonth, day: draft.endDay, precision: precision, locale: locale)
            else {
                return ""
            }
            if precision == .year, start == end {
                return appendPhrase(start, phrase: phrase, locale: locale)
            }
            return appendPhrase(L10n.Dates.rowBetween(start: start, end: end, locale: locale), phrase: phrase, locale: locale)
        }
        guard let point = side(
            year: draft.startYear, month: draft.startMonth, day: draft.startDay, precision: precision, locale: locale
        ) else {
            switch precision {
            case .full: return DateValueDisplay.string(for: input, locale: locale)
            case .year: return appendPhrase("", phrase: phrase, locale: locale)
            }
        }
        return appendPhrase(applyQualifier(draft.qualifier, to: point, locale: locale), phrase: phrase, locale: locale)
    }

    /// One side of a date at `precision`; nil when it has nothing to show.
    private static func side(year: Int32?, month: Int32?, day: Int32?, precision: Precision, locale: Locale) -> String? {
        switch precision {
        case .full: civil(year: year, month: month, day: day, locale: locale)
        case .year: civil(year: year, month: nil, day: nil, locale: locale)
        }
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
            return L10n.Dates.rowDayMonthYear(day: day, month: month, year: year, locale: locale)
        case let (nil, month?, year?):
            return L10n.Dates.rowMonthYear(month: month, year: year, locale: locale)
        case let (day?, nil, year?):
            return L10n.Dates.rowDayYear(day: day, year: year, locale: locale)
        case let (nil, nil, year?):
            return "\(year)"
        case let (day?, month?, nil):
            return L10n.Dates.rowDayMonth(day: day, month: month, locale: locale)
        case let (nil, month?, nil):
            return month
        case let (day?, nil, nil):
            return "\(day)"
        case (nil, nil, nil):
            return nil
        }
    }

    private static func appendPhrase(_ structured: String, phrase: String, locale: Locale) -> String {
        let trimmed = phrase.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return structured }
        if structured.isEmpty { return trimmed }
        return L10n.Dates.rowWithPhrase(structured, phrase: trimmed, locale: locale)
    }
}
