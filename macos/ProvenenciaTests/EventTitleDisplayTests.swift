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

    @Test func typeKeyNamesTheEventWhenTheLabelIsEmpty() {
        let birth = L10n.string(L10n.PropertyTerm.eventTypeBirth)
        #expect(
            EventTitleDisplay.title(title(.subject, typeKey: "birth", subjects: ["Jeremiah Arthur Gumtree"]), locale: en)
                == L10n.EventTitle.ofOne(type: birth, subject: "Jeremiah Arthur Gumtree", locale: en)
        )
    }

    @Test func headerTypeFillsATitleThatOmittedIt() {
        let header = CatalogEventHeader(
            entity: CatalogCanonicalEntity(id: "e", ref: "EVT-1", subjectTypeID: "t", label: ""),
            eventTypeKey: "birth",
            title: title(.subject, subjects: ["Jeremiah Arthur Gumtree"])
        )
        #expect(
            EventTitleDisplay.title(header, locale: en)
                == L10n.EventTitle.ofOne(
                    type: L10n.string(L10n.PropertyTerm.eventTypeBirth),
                    subject: "Jeremiah Arthur Gumtree",
                    locale: en
                )
        )
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
}
