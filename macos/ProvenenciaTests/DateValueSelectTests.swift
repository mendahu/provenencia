import Foundation
import Testing
@testable import Provenencia

@Suite
struct DateValueSelectTests {
    @Test func calendarOptionsUseTheFourProductCalendars() {
        let values = DateValueSelectOptions.calendars.map(\.id)
        #expect(values == ["gregorian", "julian", "french-republican", "hebrew"])
        #expect(
            DateValueSelectOptions.calendars.map(\.label)
                == [
                    String(localized: L10n.Sources.dateCalendarGregorian),
                    String(localized: L10n.Sources.dateCalendarJulian),
                    String(localized: L10n.Sources.dateCalendarFrenchRepublican),
                    String(localized: L10n.Sources.dateCalendarHebrew),
                ]
        )
    }

    @Test func monthOptionsAreEmptyPlusOneThroughTwelve() {
        let values = DateValueSelectOptions.months.map(\.id)
        #expect(values == ["", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"])
        #expect(DateValueSelectOptions.months[0].label == String(localized: L10n.Sources.dateMonthNone))
        #expect(DateValueSelectOptions.months[1].label == String(localized: L10n.Sources.dateMonthJanuary))
        #expect(DateValueSelectOptions.months[12].label == String(localized: L10n.Sources.dateMonthDecember))
    }

    @Test func monthIsDisabledWhenTheYearIsEmpty() {
        #expect(DateValueSelectOptions.isMonthDisabled(yearIsEmpty: true))
        #expect(!DateValueSelectOptions.isMonthDisabled(yearIsEmpty: false))
    }

    @Test func applyingAMonthOptionStillCascadesTheDraft() {
        var draft = DateValueDraft.empty()
        draft.startYear = 1880
        draft.startMonth = 3
        draft.applyStartCascade()
        #expect(draft.startMonth == 3)
        #expect(draft.isValid)
    }
}
