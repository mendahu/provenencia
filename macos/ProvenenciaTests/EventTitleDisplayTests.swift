import Foundation
import Testing
@testable import Provenencia

@Suite
struct EventTitleDisplayTests {
    private let en = Locale(identifier: "en_US")

    private func title(
        _ rule: CatalogEventTitle.Rule,
        recordedName: String = "",
        label: String = "",
        ref: String = "EVT-1",
        typeKey: String = "",
        typeLabel: String = "",
        subjects: [String?] = [],
        place: String = ""
    ) -> CatalogEventTitle {
        CatalogEventTitle(
            rule: rule,
            recordedName: recordedName,
            label: label,
            ref: ref,
            typeKey: typeKey,
            typeLabel: typeLabel,
            subjects: subjects.map { $0.map { CatalogNameValue(form: $0) } },
            place: place
        )
    }

    // Go chooses the rule (core/eventtitle); these check each rule's template.

    @Test func recordedNameIsTheName() {
        #expect(EventTitleDisplay.title(title(.recordedName, recordedName: "The Great Fire", label: "Grandpa's fire"), locale: en) == "The Great Fire")
    }

    @Test func oneSubject() {
        #expect(
            EventTitleDisplay.title(title(.subject, typeLabel: "Birth", subjects: ["James Robins"]), locale: en)
                == L10n.EventTitle.ofOne(type: "Birth", subject: "James Robins", locale: en)
        )
    }

    @Test func couple() {
        #expect(
            EventTitleDisplay.title(title(.couple, typeKey: "marriage", typeLabel: "Marriage", subjects: ["James Robins", "Mary Smith"]), locale: en)
                == L10n.EventTitle.marriage(a: "James Robins", b: "Mary Smith", locale: en)
        )
    }

    @Test func severalSubjectsUseEtAl() {
        #expect(
            EventTitleDisplay.title(title(.subjects, typeLabel: "Census Enumeration", subjects: ["James Robins", "Mary Smith"]), locale: en)
                == L10n.EventTitle.etAl(type: "Census Enumeration", first: "James Robins", locale: en)
        )
    }

    @Test func typeAtPlace() {
        #expect(
            EventTitleDisplay.title(title(.typeAtPlace, typeLabel: "Birth", place: "York"), locale: en)
                == L10n.EventTitle.atPlace(type: "Birth", place: "York", locale: en)
        )
    }

    @Test func unspecifiedType() {
        #expect(EventTitleDisplay.titleSource(title(.type, typeLabel: "Birth"), locale: en) == .name("Unspecified birth"))
    }

    @Test func missingTypeUsesEvent() {
        #expect(
            EventTitleDisplay.title(title(.subject, subjects: ["James Robins"]), locale: en)
                == L10n.EventTitle.ofOne(type: L10n.string(L10n.EventTitle.fallbackType), subject: "James Robins", locale: en)
        )
        #expect(
            EventTitleDisplay.title(title(.typeAtPlace, place: "York"), locale: en)
                == L10n.EventTitle.atPlace(type: L10n.string(L10n.EventTitle.fallbackType), place: "York", locale: en)
        )
    }

    @Test func unnamedPerson() {
        let unnamed = L10n.EventTitle.ofOne(type: "Birth", subject: L10n.string(L10n.EventTitle.unnamedPerson), locale: en)
        #expect(EventTitleDisplay.title(title(.subject, typeLabel: "Birth", subjects: [nil]), locale: en) == unnamed)
        #expect(EventTitleDisplay.title(title(.subject, typeLabel: "Birth", subjects: [" "]), locale: en) == unnamed)
    }

    @Test func labelIsItalicSourceAndRefIsMonoSource() {
        #expect(EventTitleDisplay.titleSource(title(.label, label: "Baptism in St James register, f. 12"), locale: en) == .label("Baptism in St James register, f. 12"))
        #expect(EventTitleDisplay.titleSource(title(.ref, ref: "EVT-9ZZ02"), locale: en) == .ref("EVT-9ZZ02"))
    }

    @Test func untitledIsLabelThenRef() {
        let entity = CatalogCanonicalEntity(id: "e", ref: "EVT-1", subjectTypeID: "t", label: " ")
        #expect(EventTitleDisplay.untitled(entity) == .ref("EVT-1"))
        var labelled = entity
        labelled.label = "Grandpa's fire"
        #expect(EventTitleDisplay.untitled(labelled) == .label("Grandpa's fire"))
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

    @Test(arguments: [
        (names: [["York", "Tkaronto"]], name: "York", extra: 0),
        (names: [["York"], ["Toronto"]], name: "York", extra: 1),
        (names: [[], ["Toronto"]], name: "Toronto", extra: 1),
        (names: [[String]](), name: "", extra: 0),
        (names: [[]], name: "", extra: 1),
    ])
    func derivedPlaceCountsPlacesNotNames(names: [[String]], name: String, extra: Int) {
        let places = names.enumerated().map { index, names in
            CatalogHeaderPlace(
                entity: CatalogCanonicalEntity(id: "p\(index)", ref: "PLC-\(index)", subjectTypeID: "t", label: ""),
                names: names,
                nameCount: names.count
            )
        }
        #expect(DerivedPlace.name(places) == name)
        #expect(DerivedPlace.extra(places) == extra)
    }
}
