import Foundation

/// Event titles from the naming matrix (R4). Go chooses the rule and its
/// parts (`CatalogEventTitle`); this fills the rule's L10n template. It never
/// chooses a rule, so every surface titles an Event the same way.
enum EventTitleDisplay {
    static func titleSource(_ title: CatalogEventTitle, locale: Locale = .autoupdatingCurrent) -> ConclusionTitleSource {
        let type = typeWord(title)
        switch title.rule {
        case .recordedName:
            return .name(title.recordedName)
        case .subject:
            return .name(L10n.EventTitle.ofOne(type: type, subject: subjectName(title, at: 0), locale: locale))
        case .couple:
            return .name(L10n.EventTitle.marriage(
                a: subjectName(title, at: 0), b: subjectName(title, at: 1), locale: locale
            ))
        case .subjects:
            return .name(L10n.EventTitle.etAl(type: type, first: subjectName(title, at: 0), locale: locale))
        case .label:
            return .label(title.label)
        case .typeAtPlace:
            return .name(L10n.EventTitle.atPlace(type: type, place: title.place, locale: locale))
        case .type:
            return .name(L10n.EventTitle.unspecified(type: type.lowercased(with: locale), locale: locale))
        case .ref:
            return .ref(title.ref)
        }
    }

    static func title(_ title: CatalogEventTitle, locale: Locale = .autoupdatingCurrent) -> String {
        titleSource(title, locale: locale).text
    }

    /// An Event whose header has not loaded: its label, else its ref.
    static func untitled(_ entity: CatalogCanonicalEntity) -> ConclusionTitleSource {
        let label = entity.label.trimmingCharacters(in: .whitespacesAndNewlines)
        return label.isEmpty ? .ref(entity.ref) : .label(label)
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

    /// A subject's displayed name, or "unnamed person".
    private static func subjectName(_ title: CatalogEventTitle, at index: Int) -> String {
        guard title.subjects.indices.contains(index), let name = title.subjects[index] else {
            return L10n.string(L10n.EventTitle.unnamedPerson)
        }
        let text = NameValueDisplay.string(for: name).trimmingCharacters(in: .whitespacesAndNewlines)
        return text.isEmpty ? L10n.string(L10n.EventTitle.unnamedPerson) : text
    }

    /// The type's label, else "Event".
    private static func typeWord(_ title: CatalogEventTitle) -> String {
        title.typeLabel.isEmpty ? L10n.string(L10n.EventTitle.fallbackType) : title.typeLabel
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
