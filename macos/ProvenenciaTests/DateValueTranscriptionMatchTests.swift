import Foundation
import Testing
@testable import Provenencia

@Suite
struct DateValueTranscriptionMatchTests {
    private let unitedStates = Locale(identifier: "en_US")
    private let unitedKingdom = Locale(identifier: "en_GB")

    @Test func dayMonthYearFillsYearMonthAndDay() {
        let match = DateValueTranscriptionMatch.first(in: "born 14 May 1817 at home", locale: unitedStates)
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: 5, day: 14))
    }

    @Test func monthDayYearFillsTheSameDate() {
        let match = DateValueTranscriptionMatch.first(in: "May 14, 1817", locale: unitedStates)
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: 5, day: 14))
    }

    @Test func isoFillsTheSameDate() {
        let match = DateValueTranscriptionMatch.first(in: "1817-05-14", locale: unitedStates)
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: 5, day: 14))
    }

    @Test func yearAloneLeavesMonthAndDayEmpty() {
        let match = DateValueTranscriptionMatch.first(in: "about 1817", locale: unitedStates)
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: nil, day: nil))
    }

    @Test func theFirstOfTwoDatesWins() {
        let match = DateValueTranscriptionMatch.first(
            in: "14 May 1817 and 2 June 1818",
            locale: unitedStates
        )
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: 5, day: 14))
    }

    @Test func noDateReturnsNil() {
        #expect(DateValueTranscriptionMatch.first(in: "no date written here", locale: unitedStates) == nil)
        #expect(DateValueTranscriptionMatch.first(in: "   ", locale: unitedStates) == nil)
    }

    @Test func unitedStatesNumericOrderIsMonthThenDay() {
        let match = DateValueTranscriptionMatch.first(in: "5/14/1817", locale: unitedStates)
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: 5, day: 14))
    }

    @Test func unitedKingdomNumericOrderIsDayThenMonth() {
        let match = DateValueTranscriptionMatch.first(in: "14/5/1817", locale: unitedKingdom)
        #expect(match == DateValueTranscriptionMatch(year: 1817, month: 5, day: 14))
    }

    @Test func anImpossibleDayIsNotFilled() {
        let match = DateValueTranscriptionMatch.first(in: "31 February 1817", locale: unitedStates)
        #expect(match != DateValueTranscriptionMatch(year: 1817, month: 2, day: 31))
    }

    @Test func accentedMonthNameReadsAtAnyCase() {
        let match = DateValueTranscriptionMatch.first(in: "3 FÉVRIER 1851", locale: Locale(identifier: "fr_FR"))
        #expect(match == DateValueTranscriptionMatch(year: 1851, month: 2, day: 3))
    }
}
