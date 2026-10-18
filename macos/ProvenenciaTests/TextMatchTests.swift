import Foundation
import Testing
@testable import Provenencia

@Suite
struct TextMatchTests {
    private let enUS = Locale(identifier: "en_US")

    @Test func fixturesMatchGo() {
        for fixture in TextMatch.fixtures {
            #expect(TextMatch.key(fixture.text, locale: enUS) == fixture.key, "\(fixture.text) → \(fixture.key)")
        }
    }

    @Test func sameLevelKeepsAccentsAndLowersEachCharacter() {
        #expect(TextMatch.key("José", at: .same, locale: enUS) == "josé")
        #expect(TextMatch.key("İstanbul", at: .same, locale: Locale(identifier: "tr_TR")) == "istanbul")
        #expect(TextMatch.key("ΟΔΟΣ", at: .same, locale: enUS) == "οδοσ")
    }

    @Test func wordsBreakOnSpacesDashesAndSlashes() {
        #expect(TextMatch.words("Mary-Ann O'Brien/Smith", at: .same, locale: enUS) == ["mary", "ann", "obrien", "smith"])
        #expect(TextMatch.words("  ", at: .same, locale: enUS) == [])
    }

    @Test func queryMatchesIgnoringCaseAccentsAndWordBreaks() {
        #expect(TextMatch.Query("jose", locale: enUS).matches("José Ruiz letter"))
        #expect(TextMatch.Query("god-father", locale: enUS).matches("Godfather"))
        #expect(!TextMatch.Query("jones", locale: enUS).matches("Smith"))
        let dash = TextMatch.Query("-", locale: enUS)
        #expect(dash.matches("Smith"))
        #expect(dash.isEmpty)
        #expect(TextMatch.Query("x", locale: enUS).matches("a", "box"))
    }

    @Test func rangeMarksTheMatchedSpan() {
        func span(_ query: String, _ text: String) -> String? {
            TextMatch.range(of: TextMatch.Query(query, locale: enUS), in: text).map { String(text[$0]) }
        }
        #expect(span("god-f", "God-father") == "God-f")
        #expect(span("godf", "God-father") == "God-f")
        #expect(span("strasse", "Straße") == "Straße")
        #expect(span("jose", "Mr José Ruiz") == "José")
        #expect(span("xyz", "José") == nil)
    }

    @Test func queryMatchesInsideAGraphemeLikeGo() {
        #expect(TextMatch.Query("가", locale: enUS).matches("간장"))
        #expect(TextMatch.Query("क", locale: enUS).matches("किताब"))
        #expect(PVTypeSelectMatcher.index(in: ["간장"], prefix: "가") == 0)
        let options = [
            PVComboBoxOption(value: "late", label: "나가", subtext: "", isDisabled: false),
            PVComboBoxOption(value: "first", label: "간장", subtext: "", isDisabled: false),
        ]
        #expect(PVComboBoxMatch.filter(options, query: "가").map(\.value) == ["first", "late"])
    }

    /// D-2's formula, inline: lower each character with the locale, decompose,
    /// drop whitespace, punctuation and (at equivalent) accents, then fold.
    private func d2Key(_ text: String, level: TextMatch.Level, locale: Locale) -> String {
        var out = String.UnicodeScalarView()
        for ch in text {
            var lowered = String(ch).lowercased(with: locale)
            if level == .equivalent { lowered = lowered.decomposedStringWithCanonicalMapping }
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
                if level == .equivalent, let fold = TextMatch.letterFolds[scalar] {
                    out.append(contentsOf: fold.unicodeScalars)
                } else {
                    out.append(scalar)
                }
            }
        }
        return String(out)
    }

    @Test func asciiFastPathKeysLikeFoundation() {
        let locales = [enUS, Locale(identifier: "tr_TR"), TextMatch.posix]
        for value in 0...0x7F {
            let ch = String(Character(Unicode.Scalar(UInt8(value))))
            for text in [ch, "a\(ch)b"] {
                for level in [TextMatch.Level.same, .equivalent] {
                    for locale in locales {
                        #expect(
                            TextMatch.key(text, at: level, locale: locale) == d2Key(text, level: level, locale: locale),
                            "U+\(String(value, radix: 16)) \(level) \(locale.identifier)"
                        )
                    }
                }
            }
        }
    }
}
