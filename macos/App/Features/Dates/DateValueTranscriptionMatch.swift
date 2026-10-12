import Foundation

/// The first Gregorian date found in a citation transcription.
/// Patterns live here, keyed by the locale the caller passes, not in the catalog.
struct DateValueTranscriptionMatch: Equatable, Sendable {
    var year: Int32
    var month: Int32?
    var day: Int32?

    /// Earliest match in `text` for `locale`. The caller chooses that locale
    /// so a source language can replace the user's. A longer match that starts
    /// at the same place wins over a shorter one, so a full date beats the year inside it.
    static func first(in text: String, locale: Locale) -> DateValueTranscriptionMatch? {
        DateValueTranscriptionScanner(locale: locale).first(in: text)
    }
}

private struct DateValueTranscriptionScanner {
    let locale: Locale

    func first(in text: String) -> DateValueTranscriptionMatch? {
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return nil }
        let candidates = collect(in: trimmed)
        return candidates.min { lhs, rhs in
            if lhs.start != rhs.start {
                return lhs.start < rhs.start
            }
            return lhs.length > rhs.length
        }?.match
    }

    private struct Candidate {
        var start: String.Index
        var length: Int
        var match: DateValueTranscriptionMatch
    }

    private func collect(in text: String) -> [Candidate] {
        var found: [Candidate] = []
        let months = monthAlternation
        if !months.isEmpty {
            add(
                #"\b(\d{1,2})(?:st|nd|rd|th)?(?:\s+of)?\s+(\#(months))\.?,?\s+(\d{3,4})\b"#,
                in: text,
                to: &found
            ) { groups in
                guard let day = Int(groups[0]), let month = monthNumber(groups[1]), let year = Int(groups[2]) else {
                    return nil
                }
                return validated(year: year, month: month, day: day)
            }
            add(
                #"\b(\#(months))\.?\s+(\d{1,2})(?:st|nd|rd|th)?,?\s+(\d{3,4})\b"#,
                in: text,
                to: &found
            ) { groups in
                guard let month = monthNumber(groups[0]), let day = Int(groups[1]), let year = Int(groups[2]) else {
                    return nil
                }
                return validated(year: year, month: month, day: day)
            }
        }
        add(
            #"\b(\d{4})-(\d{2})-(\d{2})\b"#,
            in: text,
            to: &found
        ) { groups in
            guard let year = Int(groups[0]), let month = Int(groups[1]), let day = Int(groups[2]) else {
                return nil
            }
            return validated(year: year, month: month, day: day)
        }
        let order = numericOrder
        add(
            #"\b(\d{1,2})[./-](\d{1,2})[./-](\d{3,4})\b"#,
            in: text,
            to: &found
        ) { groups in
            guard let first = Int(groups[0]), let second = Int(groups[1]), let year = Int(groups[2]) else {
                return nil
            }
            switch order {
            case .monthDayYear:
                return validated(year: year, month: first, day: second)
            case .dayMonthYear:
                return validated(year: year, month: second, day: first)
            case .yearMonthDay:
                return nil
            }
        }
        add(
            #"\b(\d{3,4})\b"#,
            in: text,
            to: &found
        ) { groups in
            guard let year = Int(groups[0]) else { return nil }
            return validated(year: year, month: nil, day: nil)
        }
        return found
    }

    private func add(
        _ pattern: String,
        in text: String,
        to found: inout [Candidate],
        map: ([String]) -> DateValueTranscriptionMatch?
    ) {
        guard let expression = try? NSRegularExpression(
            pattern: pattern,
            options: [.caseInsensitive]
        ) else { return }
        let range = NSRange(text.startIndex..<text.endIndex, in: text)
        expression.enumerateMatches(in: text, range: range) { result, _, _ in
            guard let result, let span = Range(result.range, in: text) else { return }
            var groups: [String] = []
            for index in 1..<result.numberOfRanges {
                guard let groupRange = Range(result.range(at: index), in: text) else { return }
                groups.append(String(text[groupRange]))
            }
            guard let match = map(groups) else { return }
            found.append(Candidate(
                start: span.lowerBound,
                length: text.distance(from: span.lowerBound, to: span.upperBound),
                match: match
            ))
        }
    }

    private enum NumericOrder {
        case dayMonthYear
        case monthDayYear
        case yearMonthDay
    }

    /// `yMd` template order for this locale. `en_US` is month, day, year.
    private var numericOrder: NumericOrder {
        let template = DateFormatter.dateFormat(fromTemplate: "yMd", options: 0, locale: locale) ?? ""
        let letters = template.filter { $0 == "y" || $0 == "M" || $0 == "d" }
        if letters.first == "d" { return .dayMonthYear }
        if letters.first == "y" { return .yearMonthDay }
        return .monthDayYear
    }

    private var monthAlternation: String {
        var calendar = Calendar(identifier: .gregorian)
        calendar.locale = locale
        let symbols = calendar.monthSymbols + calendar.shortMonthSymbols
        let names = Set(
            symbols.map {
                $0.trimmingCharacters(in: .whitespaces)
                    .trimmingCharacters(in: CharacterSet(charactersIn: "."))
            }.filter { $0.count >= 3 }
        )
        return names
            .sorted { $0.count > $1.count }
            .map { NSRegularExpression.escapedPattern(for: $0) }
            .joined(separator: "|")
    }

    private func monthNumber(_ name: String) -> Int? {
        var calendar = Calendar(identifier: .gregorian)
        calendar.locale = locale
        let folded = name.folding(options: [.caseInsensitive, .diacriticInsensitive], locale: locale)
            .trimmingCharacters(in: CharacterSet(charactersIn: "."))
        let lists = [calendar.monthSymbols, calendar.shortMonthSymbols]
        for symbols in lists {
            if let index = symbols.firstIndex(where: { symbol in
                symbol.folding(options: [.caseInsensitive, .diacriticInsensitive], locale: locale)
                    .trimmingCharacters(in: CharacterSet(charactersIn: ".")) == folded
            }) {
                return index + 1
            }
        }
        return nil
    }

    private func validated(year: Int, month: Int?, day: Int?) -> DateValueTranscriptionMatch? {
        guard (1...9999).contains(year) else { return nil }
        if let month {
            guard (1...12).contains(month) else { return nil }
        }
        if let day {
            guard let month else { return nil }
            var calendar = Calendar(identifier: .gregorian)
            calendar.locale = locale
            var components = DateComponents()
            components.year = year
            components.month = month
            components.day = day
            guard let date = calendar.date(from: components) else { return nil }
            let readBack = calendar.dateComponents([.year, .month, .day], from: date)
            guard readBack.year == year, readBack.month == month, readBack.day == day else { return nil }
        }
        return DateValueTranscriptionMatch(
            year: Int32(year),
            month: month.map(Int32.init),
            day: day.map(Int32.init)
        )
    }
}
