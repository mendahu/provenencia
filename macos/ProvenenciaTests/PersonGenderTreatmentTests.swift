import Testing
@testable import Provenencia

struct PersonGenderTreatmentTests {
    private typealias D = ReconciledValueDisplayTests

    private func term(_ key: String, _ label: String) -> CatalogConclusionValue {
        .term(id: "t-\(key)", key: key, label: label)
    }

    private func gender(_ values: [CatalogConclusionValue], state: String = "single") -> CatalogConclusionField {
        D.field(
            values.count > 1 ? "mixed" : state, key: "gender", label: "Gender", valueType: "term",
            values.enumerated().map { D.value($0.offset + 1, "kept", $0.element) }
        )
    }

    private func sex(_ values: [CatalogConclusionValue], state: String = "single") -> CatalogConclusionField {
        D.field(
            values.count > 1 ? "mixed" : state, key: "sex_at_birth", label: "Sex at birth", valueType: "term",
            values.enumerated().map { D.value($0.offset + 1, "kept", $0.element) }
        )
    }

    @Test func genderWinsOverSexAtBirth() {
        let female = sex([term("female", "Female")])
        let woman = PersonGenderTreatment.read([gender([term("woman", "Woman")]), female])
        #expect(woman?.family == .woman && woman?.basis == .gender)
        let man = PersonGenderTreatment.read([gender([term("man", "Man")]), female])
        #expect(man?.family == .man && man?.basis == .gender)
        #expect(PersonGenderTreatment.read([gender([term("non_binary", "Non-binary")])])?.family == .nonBinary)
        #expect(PersonGenderTreatment.read([gender([term("two_spirit", "Two-spirit")])])?.family == .other)
    }

    @Test func mixedGenderIsNeutralAndDoesNotFallBack() {
        let mixed = gender([term("woman", "Woman"), term("man", "Man")])
        #expect(PersonGenderTreatment.read([mixed, sex([term("female", "Female")])]) == nil)
    }

    @Test func sexAtBirthBorrowsAFamily() {
        let female = PersonGenderTreatment.read([sex([term("female", "Female")])])
        #expect(female?.family == .woman && female?.basis == .sexAtBirth)
        #expect(PersonGenderTreatment.read([sex([term("male", "Male")])])?.family == .man)
        #expect(PersonGenderTreatment.read([sex([term("intersex", "Intersex")])])?.family == .other)
        let emptyGender = D.field("", key: "gender", label: "Gender", valueType: "term")
        #expect(PersonGenderTreatment.read([emptyGender, sex([term("male", "Male")])])?.family == .man)
    }

    @Test func nothingUsableIsNeutral() {
        let neutral = D.field(
            "", key: "sex_at_birth", label: "Sex at birth", valueType: "term",
            outcomes: [
                D.outcome("u", "no_evidence", rank: nil, recorded: term("unknown", "Unknown")),
                D.outcome("i", "no_evidence", rank: nil, recorded: term("indeterminate", "Indeterminate")),
            ]
        )
        #expect(PersonGenderTreatment.read([neutral]) == nil)
        let mixedSex = sex([term("female", "Female"), term("male", "Male")])
        #expect(PersonGenderTreatment.read([mixedSex]) == nil)
        #expect(PersonGenderTreatment.read([]) == nil)
        #expect(PersonGenderTreatment.read([gender([.text("x")])]) == nil)
        #expect(PersonGenderTreatment.read([gender([term("", "")])]) == nil)
    }

    @Test func badgeAndSpokenText() {
        let woman = PersonGenderTreatment.read([gender([term("woman", "Woman")])])
        #expect(woman?.badge.text == "Woman")
        #expect(woman?.badge.accessibilityLabel == "Gender, Woman")
        let male = PersonGenderTreatment.read([sex([term("male", "Male")])])
        #expect(male?.badge.text == "Male · at birth")
        #expect(male?.badge.accessibilityLabel == "Sex at birth, Male")
    }
}
