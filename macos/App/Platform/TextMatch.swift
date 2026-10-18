import Foundation

/// The app's one rule for "the same text" in filters, vocabulary lookups and
/// the fake store. Mirrors Go `valuecompare.TextKey` (S11-07 D-1): case,
/// spaces, punctuation and word breaks never matter; at `equivalent` accents
/// and the registry's letter folds do not either.
enum TextMatch {
    enum Level {
        case same
        case equivalent
    }

    /// Lowercases as Go's per-rune `unicode.ToLower` does; for fixed
    /// vocabularies and the fake store, which must not depend on the machine's
    /// locale.
    static let posix = Locale(identifier: "en_US_POSIX")

    /// Go `matchregistry` `Value.Words.Folds`; `TestSwiftTextMatchMirrorsTextKey` keeps them equal.
    static let letterFolds: [Unicode.Scalar: String] = [
        "ß": "ss",
        "æ": "ae",
        "œ": "oe",
        "ø": "o",
        "ł": "l",
        "đ": "d",
        "þ": "th",
        "ı": "i",
    ]

    /// Shared cases with `valuecompare.TextKey` at `equivalent`;
    /// `TestSwiftTextMatchMirrorsTextKey` reads this table.
    static let fixtures: [(text: String, key: String)] = [
        ("god-father", "godfather"),
        ("God Father", "godfather"),
        ("Mont-réal", "montreal"),
        ("Ó Riain", "oriain"),
        ("O'Brien", "obrien"),
        ("A+B", "a+b"),
        ("Straße", "strasse"),
        ("Søren Ærø", "sorenaero"),
        ("Łódź", "lodz"),
        ("Þórr", "thorr"),
        ("event_type", "eventtype"),
        ("SRC-4F2NP", "src4f2np"),
        ("München / Munich", "munchenmunich"),
        ("İstanbul", "istanbul"),
        ("!!!", ""),
    ]

    /// The letters of `text` with no separator. Returns `""` when nothing survives.
    static func key(_ text: String, at level: Level = .equivalent, locale: Locale = .autoupdatingCurrent) -> String {
        var out = String.UnicodeScalarView()
        for ch in text {
            out.append(contentsOf: characterKey(ch, level: level, locale: locale))
        }
        return String(out)
    }

    /// Go's word breaks (`rawTokens`); unlike Go `SplitWords`, particles do not rejoin.
    static func words(_ text: String, at level: Level = .equivalent, locale: Locale = .autoupdatingCurrent) -> [String] {
        var pieces: [String.UnicodeScalarView] = []
        var current = String.UnicodeScalarView()
        for scalar in text.unicodeScalars {
            if scalar.properties.isWhitespace || scalar.properties.generalCategory == .dashPunctuation || scalar == "/" {
                pieces.append(current)
                current = String.UnicodeScalarView()
            } else {
                current.append(scalar)
            }
        }
        pieces.append(current)
        return pieces.map { key(String($0), at: level, locale: locale) }.filter { !$0.isEmpty }
    }

    /// Typed text, keyed once.
    struct Query {
        let key: String
        let locale: Locale

        init(_ typed: String, locale: Locale = .autoupdatingCurrent) {
            self.key = TextMatch.key(typed, locale: locale)
            self.locale = locale
        }

        var isEmpty: Bool { key.isEmpty }

        /// True for an empty query, or when any text's key contains the query.
        func matches(_ texts: String...) -> Bool {
            isEmpty || texts.contains { TextMatch.key($0, locale: locale).contains(key) }
        }

        /// `matches` for a text that is already keyed.
        func matchesKey(_ textKey: String) -> Bool {
            isEmpty || textKey.contains(key)
        }
    }

    /// The span of `text` whose keyed characters hold `query`, or nil.
    static func range(of query: Query, in text: String) -> Range<String.Index>? {
        if query.isEmpty { return nil }
        var keyed: [Unicode.Scalar] = []
        var owner: [String.Index] = []
        for idx in text.indices {
            for scalar in characterKey(text[idx], level: .equivalent, locale: query.locale) {
                keyed.append(scalar)
                owner.append(idx)
            }
        }
        let needle = Array(query.key.unicodeScalars)
        let n = needle.count
        guard keyed.count >= n else { return nil }
        for i in 0...(keyed.count - n) where Array(keyed[i..<i + n]) == needle {
            return owner[i]..<text.index(after: owner[i + n - 1])
        }
        return nil
    }

    private static func characterKey(_ ch: Character, level: Level, locale: Locale) -> [Unicode.Scalar] {
        var lowered = String(ch).lowercased(with: locale)
        if level == .equivalent {
            lowered = lowered.decomposedStringWithCanonicalMapping
        }
        var out: [Unicode.Scalar] = []
        for scalar in lowered.unicodeScalars {
            let props = scalar.properties
            if props.isWhitespace { continue }
            switch props.generalCategory {
            case .connectorPunctuation, .dashPunctuation, .openPunctuation, .closePunctuation,
                 .initialPunctuation, .finalPunctuation, .otherPunctuation:
                continue
            case .nonspacingMark where level == .equivalent:
                continue
            default:
                break
            }
            if level == .equivalent, let fold = letterFolds[scalar] {
                out.append(contentsOf: fold.unicodeScalars)
            } else {
                out.append(scalar)
            }
        }
        return out
    }
}
