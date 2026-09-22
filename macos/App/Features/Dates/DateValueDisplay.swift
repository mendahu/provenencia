import Foundation

/// Locale-aware compact summary strings for genealogical DateValues (list rows, previews).
/// Source of truth for user-visible DateValue text on macOS — not Source-page-owned.
enum DateValueDisplay {
    static func string(for draft: DateValueDraft, locale: Locale = .autoupdatingCurrent) -> String {
        guard draft.isValid else { return "" }

        let phrase = draft.phrase.trimmingCharacters(in: .whitespacesAndNewlines)

        switch draft.kind {
        case "point":
            if draft.startYear == nil {
                return phrase
            }
            guard let point = formatSide(
                year: draft.startYear,
                month: draft.startMonth,
                day: draft.startDay,
                hour: draft.startHour,
                minute: draft.startMinute,
                second: draft.startSecond,
                locale: locale
            ) else {
                return ""
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
        guard let year else { return nil }

        var calendar = Calendar(identifier: .gregorian)
        calendar.locale = locale

        var components = DateComponents()
        components.year = Int(year)

        let template: String
        if let month {
            components.month = Int(month)
            if let day {
                components.day = Int(day)
                template = "yMMMd"
            } else {
                template = "yMMM"
            }
        } else {
            template = "y"
        }

        guard let date = calendar.date(from: components) else { return nil }
        guard let format = DateFormatter.dateFormat(fromTemplate: template, options: 0, locale: locale)
        else { return nil }

        let dateFormatter = DateFormatter()
        dateFormatter.locale = locale
        dateFormatter.calendar = calendar
        dateFormatter.dateFormat = format
        var result = dateFormatter.string(from: date)

        if let hour {
            components.hour = Int(hour)
            components.minute = Int(minute ?? 0)
            components.second = Int(second ?? 0)
            if let timeDate = calendar.date(from: components) {
                let timeFormatter = DateFormatter()
                timeFormatter.locale = locale
                timeFormatter.calendar = calendar
                timeFormatter.dateStyle = .none
                timeFormatter.timeStyle = .short
                result += " \(timeFormatter.string(from: timeDate))"
            }
        }

        return result
    }
}
