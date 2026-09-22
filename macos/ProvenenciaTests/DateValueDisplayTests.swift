import Foundation
import Testing
@testable import Provenencia

@Suite
struct DateValueDisplayTests {
    private let enUS = Locale(identifier: "en_US")
    private let enGB = Locale(identifier: "en_GB")
    private let frFR = Locale(identifier: "fr_FR")

    @Test func invalidDraftIsEmpty() {
        #expect(DateValueDisplay.string(for: DateValueDraft.empty(), locale: enUS) == "")
    }

    @Test func yearOnlyUsesLocaleYear() {
        var d = DateValueDraft.empty()
        d.startYear = 1882
        #expect(DateValueDisplay.string(for: d, locale: enUS) == "1882")
        #expect(DateValueDisplay.string(for: d, locale: frFR) == "1882")
    }

    @Test func yearMonthUsesLocaleTemplate() {
        var d = DateValueDraft.empty()
        d.startYear = 1882
        d.startMonth = 4
        let us = DateValueDisplay.string(for: d, locale: enUS)
        let gb = DateValueDisplay.string(for: d, locale: enGB)
        let fr = DateValueDisplay.string(for: d, locale: frFR)
        #expect(us.contains("1882"))
        #expect(us.contains("Apr"))
        #expect(gb.contains("1882"))
        #expect(gb.contains("Apr"))
        #expect(fr.contains("1882"))
        #expect(fr.lowercased().contains("avr"))
    }

    @Test func fullDayUsesLocaleMediumDate() {
        var d = DateValueDraft.empty()
        d.startYear = 1882
        d.startMonth = 4
        d.startDay = 3
        let us = DateValueDisplay.string(for: d, locale: enUS)
        let gb = DateValueDisplay.string(for: d, locale: enGB)
        #expect(us == "Apr 3, 1882")
        #expect(gb == "3 Apr 1882")
    }

    @Test(arguments: [
        ("ABT", "About Apr 1882"),
        ("BEF", "Before Apr 1882"),
        ("AFT", "After Apr 1882"),
    ])
    func pointQualifier(qualifier: String, expected: String) {
        var d = DateValueDraft.empty()
        d.qualifier = qualifier
        d.startYear = 1882
        d.startMonth = 4
        #expect(DateValueDisplay.string(for: d, locale: enUS) == expected)
    }

    @Test func rangeWithAsymmetricPrecision() {
        var d = DateValueDraft.empty()
        d.setKind("range")
        d.startYear = 1880
        d.endYear = 1882
        d.endMonth = 4
        d.endDay = 3
        #expect(
            DateValueDisplay.string(for: d, locale: enUS)
                == "Between 1880 and Apr 3, 1882"
        )
    }

    @Test func phraseOnlyReturnsPhrase() {
        var d = DateValueDraft.empty()
        d.phrase = "Christmas"
        #expect(DateValueDisplay.string(for: d, locale: enUS) == "Christmas")
    }

    @Test func structurePlusPhraseAppendsMiddleDot() {
        var d = DateValueDraft.empty()
        d.startYear = 1882
        d.startMonth = 4
        d.startDay = 3
        d.phrase = "baptism day"
        #expect(
            DateValueDisplay.string(for: d, locale: enUS)
                == "Apr 3, 1882 · baptism day"
        )
    }

    @Test func catalogInputOverload() {
        var d = DateValueDraft.empty()
        d.startYear = 1900
        d.qualifier = "ABT"
        let input = d.toInput()
        #expect(DateValueDisplay.string(for: input, locale: enUS) == "About 1900")
    }
}
