import Foundation
import Testing
@testable import Provenencia

@Suite
struct DateRowDisplayTests {
    private let en = Locale(identifier: "en_US")

    @Test func dateLineUsesThePointOrTheSpan() {
        let day = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        let start = CatalogDateValueInput(kind: "point", startYear: 1849)
        let end = CatalogDateValueInput(kind: "point", startYear: 1851)
        #expect(DateRowDisplay.line(date: day, start: start, end: end, locale: en) == "14 May 1817")
        #expect(DateRowDisplay.line(date: nil, start: start, end: end, locale: en) == "1849–1851")
        #expect(DateRowDisplay.line(date: nil, start: start, locale: en) == "1849")
        #expect(DateRowDisplay.line(date: nil, end: end, locale: en) == "1851")
        #expect(DateRowDisplay.line(date: nil, locale: en) == "")
    }

    @Test func dateLineCanLeaveOffThePhrase() {
        var spring = CatalogDateValueInput(kind: "point", startYear: 1842)
        spring.phrase = "in the spring"
        #expect(DateRowDisplay.line(date: spring, locale: en).contains("in the spring"))
        #expect(DateRowDisplay.line(date: spring, includesPhrase: false, locale: en) == "1842")
    }

    @Test func dateLineKeepsQualifiersAndRanges() {
        let about = CatalogDateValueInput(kind: "point", qualifier: "ABT", startYear: 1810)
        let before = CatalogDateValueInput(kind: "point", qualifier: "BEF", startYear: 1790, startMonth: 3)
        let after = CatalogDateValueInput(kind: "point", qualifier: "AFT", startYear: 1880)
        var range = CatalogDateValueInput(kind: "range", startYear: 1803)
        range.endYear = 1806
        #expect(DateRowDisplay.line(date: about, locale: en) == "abt 1810")
        #expect(DateRowDisplay.line(date: before, locale: en) == "bef Mar 1790")
        #expect(DateRowDisplay.line(date: after, locale: en) == "aft 1880")
        #expect(DateRowDisplay.line(date: range, locale: en) == "bet 1803 and 1806")
    }

    @Test func yearPrecision() {
        func year(_ input: CatalogDateValueInput) -> String {
            DateRowDisplay.line(date: input, precision: .year, locale: en)
        }
        let day = CatalogDateValueInput(kind: "point", startYear: 1790, startMonth: 5, startDay: 3)
        #expect(year(day) == "1790")
        #expect(DateRowDisplay.line(date: day, precision: .full, locale: en) == "3 May 1790")
        #expect(DateRowDisplay.line(date: day, locale: en) == "3 May 1790")
        var about = day
        about.qualifier = "ABT"
        #expect(year(about) == "abt 1790")
        #expect(year(CatalogDateValueInput(kind: "point", qualifier: "BEF", startYear: 1790)) == "bef 1790")
        var range = CatalogDateValueInput(kind: "range", startYear: 1790)
        range.endYear = 1795
        #expect(year(range) == "bet 1790 and 1795")
        range.endYear = 1790
        #expect(year(range) == "1790")
        #expect(year(CatalogDateValueInput(kind: "point", startYear: 1790, startMonth: 5)) == "1790")
        #expect(year(CatalogDateValueInput(kind: "point", startMonth: 5)) == "")
        var phraseOnly = CatalogDateValueInput(kind: "point")
        phraseOnly.phrase = "in the spring"
        #expect(year(phraseOnly) == "in the spring")
    }
}
