import Foundation

/// Parts of an Event title. Go supplies the ones it has composed. The
/// Persons and Events lists pass subjects and place through in S9-32. The
/// date is not part of the title.
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

/// The first named Place's first kept name, and how many other Places a walk
/// reached. A Place's other names (Montréal and Montreal) are one place, not
/// a disagreement. A chain is not part of the name.
enum DerivedPlace {
    static func name(_ places: [CatalogHeaderPlace]) -> String {
        for place in places {
            for name in place.names {
                let trimmed = name.trimmingCharacters(in: .whitespacesAndNewlines)
                if !trimmed.isEmpty { return trimmed }
            }
        }
        return ""
    }

    /// Places beyond the one `name` shows: the list's +N.
    static func extra(_ places: [CatalogHeaderPlace]) -> Int {
        max(0, places.count - (name(places).isEmpty ? 0 : 1))
    }
}

/// A Person's birth and death as one secondary line: dates, then places.
/// A missing half is left out. A birth with no death keeps the dash.
enum PersonLifeDisplay {
    struct Line: Equatable {
        var text: String
        /// Places beyond the ones in `text`. The list's +N.
        var extraPlaces: Int
    }

    static func line(_ header: CatalogPersonHeader, locale: Locale = .autoupdatingCurrent) -> Line {
        let dates = dateSpan(birth: header.birth.date, death: header.death.date, locale: locale)
        let places = placeSpan(birth: header.birth.places, death: header.death.places)
        let extra = DerivedPlace.extra(header.birth.places) + DerivedPlace.extra(header.death.places)
        let text: String
        switch (dates.isEmpty, places.isEmpty) {
        case (true, true):
            text = ""
        case (false, true):
            text = dates
        case (true, false):
            text = places
        case (false, false):
            text = "\(dates) · \(places)"
        }
        return Line(text: text, extraPlaces: extra)
    }

    /// The date a life row shows. Empty when none is recorded.
    static func dateText(_ date: CatalogDateValueInput?, locale: Locale = .autoupdatingCurrent) -> String {
        guard let date else { return "" }
        return EventTitleDisplay.dateLine(date: date, locale: locale)
    }

    private static func dateSpan(
        birth: CatalogDateValueInput?,
        death: CatalogDateValueInput?,
        locale: Locale
    ) -> String {
        let born = dateText(birth, locale: locale)
        let died = dateText(death, locale: locale)
        switch (born.isEmpty, died.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return "\(born) –"
        case (true, false):
            return died
        case (false, false):
            return "\(born) – \(died)"
        }
    }

    private static func placeSpan(birth: [CatalogHeaderPlace], death: [CatalogHeaderPlace]) -> String {
        let born = DerivedPlace.name(birth)
        let died = DerivedPlace.name(death)
        switch (born.isEmpty, died.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return born
        case (true, false):
            return died
        case (false, false):
            return "\(born) → \(died)"
        }
    }
}
