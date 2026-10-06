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
    static func title(_ parts: EventTitleParts, locale: Locale = .autoupdatingCurrent) -> String {
        let recorded = trim(parts.recordedName)
        if !recorded.isEmpty {
            return recorded
        }
        if !parts.subjects.isEmpty {
            return subjectTitle(parts, locale: locale)
        }
        let label = trim(parts.label)
        if !label.isEmpty {
            return label
        }
        let place = trim(parts.place)
        let type = typeWord(parts)
        if !place.isEmpty {
            return L10n.EventTitle.atPlace(type: type, place: place, locale: locale)
        }
        if hasType(parts) {
            return L10n.EventTitle.unspecified(type: type, locale: locale)
        }
        return trim(parts.ref)
    }

    /// The row's date: the point date, else the start–end span. Empty when
    /// none of them is set.
    static func dateLine(
        date: CatalogDateValueInput?,
        start: CatalogDateValueInput? = nil,
        end: CatalogDateValueInput? = nil,
        locale: Locale = .autoupdatingCurrent
    ) -> String {
        if let date {
            return DateValueDisplay.string(for: date, locale: locale)
        }
        let startText = start.map { DateValueDisplay.string(for: $0, locale: locale) } ?? ""
        let endText = end.map { DateValueDisplay.string(for: $0, locale: locale) } ?? ""
        switch (startText.isEmpty, endText.isEmpty) {
        case (true, true):
            return ""
        case (false, true):
            return startText
        case (true, false):
            return endText
        case (false, false):
            return L10n.Dates.displayBetween(start: startText, end: endText, locale: locale)
        }
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
