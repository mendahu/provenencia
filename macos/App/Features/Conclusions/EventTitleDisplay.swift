import Foundation

/// Parts of an Event title. Go supplies the ones it has composed; subject
/// and place parts stay empty until S9-31 and S9-32. The date is not part
/// of the title.
struct EventTitleParts: Equatable, Sendable {
    var recordedName: String = ""
    var label: String = ""
    var ref: String = ""
    var typeKey: String = ""
    var typeLabel: String = ""
    var subjects: [String] = []
    var place: String = ""
}

/// Event titles from the naming matrix (R4). A recorded name wins, then
/// subjects, then the working label, then type and place, then the ref.
enum EventTitleDisplay {
    static func titleSource(_ parts: EventTitleParts, locale: Locale = .autoupdatingCurrent) -> ConclusionTitleSource {
        let recorded = trim(parts.recordedName)
        if !recorded.isEmpty {
            return .name(recorded)
        }
        if !parts.subjects.isEmpty {
            return .name(subjectTitle(parts, locale: locale))
        }
        let label = trim(parts.label)
        if !label.isEmpty {
            return .label(label)
        }
        let place = trim(parts.place)
        let type = typeWord(parts)
        if !place.isEmpty {
            return .name(L10n.EventTitle.atPlace(type: type, place: place, locale: locale))
        }
        if hasType(parts) {
            return .name(L10n.EventTitle.unspecified(type: type.lowercased(with: locale), locale: locale))
        }
        return .ref(trim(parts.ref))
    }

    static func title(_ parts: EventTitleParts, locale: Locale = .autoupdatingCurrent) -> String {
        titleSource(parts, locale: locale).text
    }

    /// The row's date: a compact genealogical line. A point date wins; otherwise
    /// the start–end span. Empty when none of them is set. Editors keep
    /// `DateValueDisplay`.
    static func dateLine(
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
        if draft.kind == "range" {
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
        case "ABT":
            return L10n.Dates.rowAbout(point, locale: locale)
        case "BEF":
            return L10n.Dates.rowBefore(point, locale: locale)
        case "AFT":
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

    private static func subjectTitle(_ parts: EventTitleParts, locale: Locale) -> String {
        let type = typeWord(parts)
        let names = parts.subjects.map { name in
            let trimmed = trim(name)
            return trimmed.isEmpty ? L10n.string(L10n.EventTitle.unnamedPerson) : trimmed
        }
        if names.count == 1 {
            return L10n.EventTitle.ofOne(type: type, subject: names[0], locale: locale)
        }
        if names.count == 2, parts.typeKey == "marriage" {
            return L10n.EventTitle.marriage(a: names[0], b: names[1], locale: locale)
        }
        return L10n.EventTitle.etAl(type: type, first: names[0], locale: locale)
    }

    private static func typeWord(_ parts: EventTitleParts) -> String {
        let label = trim(parts.typeLabel)
        if !label.isEmpty {
            return label
        }
        return L10n.string(L10n.EventTitle.fallbackType)
    }

    private static func hasType(_ parts: EventTitleParts) -> Bool {
        !trim(parts.typeLabel).isEmpty || !trim(parts.typeKey).isEmpty
    }

    private static func trim(_ value: String) -> String {
        value.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}
