import Foundation

/// Locale-aware compact summary strings for genealogical DateValues (list rows, previews).
/// Source of truth for user-visible DateValue text on macOS — not Source-page-owned.
enum DateValueDisplay {
    static func string(for draft: DateValueDraft, locale: Locale = .autoupdatingCurrent) -> String {
        guard draft.isValid else { return "" }

        let phrase = draft.phrase.trimmingCharacters(in: .whitespacesAndNewlines)

        switch draft.kind {
        case "point":
            guard let point = formatSide(
                year: draft.startYear,
                month: draft.startMonth,
                day: draft.startDay,
                hour: draft.startHour,
                minute: draft.startMinute,
                second: draft.startSecond,
                locale: locale
            ) else {
                return phrase
            }
            let qualified = applyQualifier(draft.qualifier, to: point, locale: locale)
            return appendPhrase(phrase, to: qualified)

        case "range":
            guard
                let start = formatSide(
                    year: draft.startYear,
                    month: draft.startMonth,
                    day: draft.startDay,
                    hour: draft.startHour,
                    minute: draft.startMinute,
                    second: draft.startSecond,
                    locale: locale
                ),
                let end = formatSide(
                    year: draft.endYear,
                    month: draft.endMonth,
                    day: draft.endDay,
                    hour: draft.endHour,
                    minute: draft.endMinute,
                    second: draft.endSecond,
                    locale: locale
                )
            else {
                return ""
            }
            let between = L10n.Dates.displayBetween(start: start, end: end, locale: locale)
            return appendPhrase(phrase, to: between)

        default:
            return ""
        }
    }

    static func string(for input: CatalogDateValueInput, locale: Locale = .autoupdatingCurrent) -> String {
        string(for: DateValueDraft(from: input), locale: locale)
    }

    // MARK: Private

    private static func appendPhrase(_ phrase: String, to structured: String) -> String {
        guard !phrase.isEmpty else { return structured }
        return "\(structured) · \(phrase)"
    }

    private static func applyQualifier(_ raw: String, to point: String, locale: Locale) -> String {
        switch raw.trimmingCharacters(in: .whitespacesAndNewlines) {
        case "ABT":
            return L10n.Dates.displayAbout(point, locale: locale)
        case "BEF":
            return L10n.Dates.displayBefore(point, locale: locale)
        case "AFT":
            return L10n.Dates.displayAfter(point, locale: locale)
        default:
            return point
        }
    }

    private static func formatSide(
        year: Int32?,
        month: Int32?,
        day: Int32?,
        hour: Int32?,
        minute: Int32?,
        second: Int32?,
        locale: Locale
    ) -> String? {
        if year == nil, month == nil, day == nil, hour == nil, minute == nil, second == nil {
            return nil
        }

        var calendar = Calendar(identifier: .gregorian)
        calendar.locale = locale

        var components = DateComponents()
        components.year = year.map(Int.init) ?? 2000
        components.month = month.map(Int.init) ?? 1
        components.day = day.map(Int.init) ?? 1

        var result = ""
        if year != nil || month != nil || day != nil {
            let template: String
            switch (year != nil, month != nil, day != nil) {
            case (true, true, true): template = "yMMMd"
            case (true, true, false): template = "yMMM"
            case (true, false, true): template = "yd"
            case (true, false, false): template = "y"
            case (false, true, true): template = "MMMMd"
            case (false, true, false): template = "MMMM"
            case (false, false, true): template = "d"
            case (false, false, false): template = ""
            }
            if let date = calendar.date(from: components),
               let format = DateFormatter.dateFormat(fromTemplate: template, options: 0, locale: locale)
            {
                let dateFormatter = DateFormatter()
                dateFormatter.locale = locale
                dateFormatter.calendar = calendar
                dateFormatter.dateFormat = format
                result = dateFormatter.string(from: date)
            }
        }

        if hour != nil || minute != nil || second != nil {
            if let hour { components.hour = Int(hour) }
            if let minute { components.minute = Int(minute) }
            if let second { components.second = Int(second) }
            let timeTemplate: String
            switch (hour != nil, minute != nil, second != nil) {
            case (true, true, true): timeTemplate = "jms"
            case (true, true, false): timeTemplate = "jm"
            case (true, false, true): timeTemplate = "js"
            case (true, false, false): timeTemplate = "j"
            case (false, true, true): timeTemplate = "ms"
            case (false, true, false): timeTemplate = "m"
            case (false, false, true): timeTemplate = "s"
            case (false, false, false): timeTemplate = ""
            }
            if let timeDate = calendar.date(from: components),
               let format = DateFormatter.dateFormat(fromTemplate: timeTemplate, options: 0, locale: locale)
            {
                let timeFormatter = DateFormatter()
                timeFormatter.locale = locale
                timeFormatter.calendar = calendar
                timeFormatter.dateFormat = format
                let clock = timeFormatter.string(from: timeDate)
                result = result.isEmpty ? clock : "\(result) \(clock)"
            }
        }

        return result.isEmpty ? nil : result
    }
}
