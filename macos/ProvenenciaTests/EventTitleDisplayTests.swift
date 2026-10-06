import Foundation
import Testing
@testable import Provenencia

@Suite
struct EventTitleDisplayTests {
    private let en = Locale(identifier: "en_US")

    private func parts(
        recordedName: String = "",
        label: String = "",
        ref: String = "EVT-1",
        typeKey: String = "",
        typeLabel: String = "",
        subjects: [String] = [],
        place: String = ""
    ) -> EventTitleParts {
        EventTitleParts(
            recordedName: recordedName,
            label: label,
            ref: ref,
            typeKey: typeKey,
            typeLabel: typeLabel,
            subjects: subjects,
            place: place
        )
    }

    @Test func recordedNameWins() {
        let title = EventTitleDisplay.title(
            parts(recordedName: " The Great Fire ", label: "Grandpa's fire", typeLabel: "Birth", subjects: ["James"], place: "York"),
            locale: en
        )
        #expect(title == "The Great Fire")
    }

    @Test func oneSubject() {
        #expect(
            EventTitleDisplay.title(parts(typeLabel: "Birth", subjects: ["James Robins"]), locale: en)
                == L10n.EventTitle.ofOne(type: "Birth", subject: "James Robins", locale: en)
        )
    }

    @Test func marriageOfTwo() {
        #expect(
            EventTitleDisplay.title(parts(typeKey: "marriage", typeLabel: "Marriage", subjects: ["James Robins", "Mary Smith"]), locale: en)
                == L10n.EventTitle.marriage(a: "James Robins", b: "Mary Smith", locale: en)
        )
    }

    @Test func severalSubjectsUseEtAl() {
        #expect(
            EventTitleDisplay.title(parts(typeLabel: "Census Enumeration", subjects: ["James Robins", "Mary Smith"]), locale: en)
                == L10n.EventTitle.etAl(type: "Census Enumeration", first: "James Robins", locale: en)
        )
        #expect(
            EventTitleDisplay.title(parts(typeKey: "marriage", typeLabel: "Marriage", subjects: ["James", "Mary", "Ann"]), locale: en)
                == L10n.EventTitle.etAl(type: "Marriage", first: "James", locale: en)
        )
    }

    @Test func labelBeatsTypeAndPlace() {
        #expect(
            EventTitleDisplay.title(parts(label: " Grandpa's fire ", typeLabel: "Birth", place: "York"), locale: en)
                == "Grandpa's fire"
        )
    }

    @Test func typeAtPlace() {
        #expect(
            EventTitleDisplay.title(parts(typeLabel: "Birth", place: "York"), locale: en)
                == L10n.EventTitle.atPlace(type: "Birth", place: "York", locale: en)
        )
    }

    @Test func unspecifiedType() {
        #expect(
            EventTitleDisplay.title(parts(typeLabel: "Birth"), locale: en)
                == L10n.EventTitle.unspecified(type: "Birth", locale: en)
        )
    }

    @Test func missingTypeUsesEvent() {
        #expect(
            EventTitleDisplay.title(parts(subjects: ["James Robins"]), locale: en)
                == L10n.EventTitle.ofOne(type: L10n.string(L10n.EventTitle.fallbackType), subject: "James Robins", locale: en)
        )
        #expect(
            EventTitleDisplay.title(parts(place: "York"), locale: en)
                == L10n.EventTitle.atPlace(type: L10n.string(L10n.EventTitle.fallbackType), place: "York", locale: en)
        )
    }

    @Test func unnamedPerson() {
        #expect(
            EventTitleDisplay.title(parts(typeLabel: "Birth", subjects: [" "]), locale: en)
                == L10n.EventTitle.ofOne(type: "Birth", subject: L10n.string(L10n.EventTitle.unnamedPerson), locale: en)
        )
    }

    @Test func refIsLast() {
        #expect(EventTitleDisplay.title(parts(ref: "EVT-4MA10"), locale: en) == "EVT-4MA10")
    }

    @Test func dateLineUsesThePointOrTheSpan() {
        let day = CatalogDateValueInput(kind: "point", startYear: 1817, startMonth: 5, startDay: 14)
        let start = CatalogDateValueInput(kind: "point", startYear: 1900)
        let end = CatalogDateValueInput(kind: "point", startYear: 1910)
        #expect(EventTitleDisplay.dateLine(date: day, start: start, end: end, locale: en) == DateValueDisplay.string(for: day, locale: en))
        #expect(
            EventTitleDisplay.dateLine(date: nil, start: start, end: end, locale: en)
                == L10n.Dates.displayBetween(
                    start: DateValueDisplay.string(for: start, locale: en),
                    end: DateValueDisplay.string(for: end, locale: en),
                    locale: en
                )
        )
        #expect(EventTitleDisplay.dateLine(date: nil, start: start, locale: en) == DateValueDisplay.string(for: start, locale: en))
        #expect(EventTitleDisplay.dateLine(date: nil, end: end, locale: en) == DateValueDisplay.string(for: end, locale: en))
        #expect(EventTitleDisplay.dateLine(date: nil, locale: en) == "")
    }
}
