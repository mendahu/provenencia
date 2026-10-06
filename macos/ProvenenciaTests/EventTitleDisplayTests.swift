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
                == L10n.EventTitle.unspecified(type: "birth", locale: en)
        )
        #expect(EventTitleDisplay.titleSource(parts(typeLabel: "Birth"), locale: en) == .name("Unspecified birth"))
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
        let start = CatalogDateValueInput(kind: "point", startYear: 1849)
        let end = CatalogDateValueInput(kind: "point", startYear: 1851)
        #expect(EventTitleDisplay.dateLine(date: day, start: start, end: end, locale: en) == "14 May 1817")
        #expect(EventTitleDisplay.dateLine(date: nil, start: start, end: end, locale: en) == "1849–1851")
        #expect(EventTitleDisplay.dateLine(date: nil, start: start, locale: en) == "1849")
        #expect(EventTitleDisplay.dateLine(date: nil, end: end, locale: en) == "1851")
        #expect(EventTitleDisplay.dateLine(date: nil, locale: en) == "")
    }

    @Test func dateLineKeepsQualifiersAndRanges() {
        let about = CatalogDateValueInput(kind: "point", qualifier: "ABT", startYear: 1810)
        let before = CatalogDateValueInput(kind: "point", qualifier: "BEF", startYear: 1790, startMonth: 3)
        let after = CatalogDateValueInput(kind: "point", qualifier: "AFT", startYear: 1880)
        var range = CatalogDateValueInput(kind: "range", startYear: 1803)
        range.endYear = 1806
        #expect(EventTitleDisplay.dateLine(date: about, locale: en) == "abt 1810")
        #expect(EventTitleDisplay.dateLine(date: before, locale: en) == "bef Mar 1790")
        #expect(EventTitleDisplay.dateLine(date: after, locale: en) == "aft 1880")
        #expect(EventTitleDisplay.dateLine(date: range, locale: en) == "bet 1803 and 1806")
    }

    @Test func labelIsItalicSourceAndRefIsMonoSource() {
        #expect(EventTitleDisplay.titleSource(parts(label: "Baptism in St James register, f. 12"), locale: en) == .label("Baptism in St James register, f. 12"))
        #expect(EventTitleDisplay.titleSource(parts(ref: "EVT-9ZZ02"), locale: en) == .ref("EVT-9ZZ02"))
    }
}
