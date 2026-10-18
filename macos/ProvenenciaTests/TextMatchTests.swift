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
}
