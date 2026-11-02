import Testing
@testable import Provenencia

struct PersonGenderTreatmentTests {
    private func facts(
        _ key: String, _ label: String, origin: String = CatalogOrigin.provenencia,
        count: Int = 1, propertyLabel: String
    ) -> CatalogTermFacts {
        CatalogTermFacts(
            term: .term(id: "t-\(key)", key: key, label: label, origin: origin),
            count: count, propertyLabel: propertyLabel
        )
    }

    private func gender(_ key: String, _ label: String, origin: String = CatalogOrigin.provenencia, count: Int = 1) -> CatalogTermFacts {
        facts(key, label, origin: origin, count: count, propertyLabel: "Gender")
    }

    private func sex(_ key: String, _ label: String, origin: String = CatalogOrigin.provenencia, count: Int = 1) -> CatalogTermFacts {
        facts(key, label, origin: origin, count: count, propertyLabel: "Sex at birth")
    }

    private func summary(gender: CatalogTermFacts = CatalogTermFacts(), sex: CatalogTermFacts = CatalogTermFacts()) -> CatalogPersonSummary {
        var h = CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: ""),
            name: nil, nameValueCount: 0
        )
        h.gender = gender
        h.sexAtBirth = sex
        return h
    }

    @Test func genderWinsOverSexAtBirth() {
        let female = sex("female", "Female")
        let woman = PersonGenderTreatment.read(summary(gender: gender("woman", "Woman"), sex: female))
        #expect(woman?.family == .woman && woman?.basis == .gender)
        let man = PersonGenderTreatment.read(summary(gender: gender("man", "Man"), sex: female))
        #expect(man?.family == .man && man?.basis == .gender)
        #expect(PersonGenderTreatment.read(summary(gender: gender("non_binary", "Non-binary")))?.family == .nonBinary)
        #expect(PersonGenderTreatment.read(summary(gender: gender("two_spirit", "Two-spirit")))?.family == .other)
    }

    @Test func aResearchersTermKeyedLikeAProductOneKeepsTheOtherFamily() {
        let theirs = gender("woman", "Woman", origin: CatalogOrigin.user)
        #expect(PersonGenderTreatment.read(summary(gender: theirs))?.family == .other)
        let theirSex = sex("female", "Female", origin: CatalogOrigin.user)
        #expect(PersonGenderTreatment.read(summary(sex: theirSex))?.family == .other)
    }

    @Test func mixedGenderIsNeutralAndDoesNotFallBack() {
        let mixed = gender("woman", "Woman", count: 2)
        #expect(PersonGenderTreatment.read(summary(gender: mixed, sex: sex("female", "Female"))) == nil)
    }

    @Test func sexAtBirthBorrowsAFamily() {
        let female = PersonGenderTreatment.read(summary(sex: sex("female", "Female")))
        #expect(female?.family == .woman && female?.basis == .sexAtBirth)
        #expect(PersonGenderTreatment.read(summary(sex: sex("male", "Male")))?.family == .man)
        #expect(PersonGenderTreatment.read(summary(sex: sex("intersex", "Intersex")))?.family == .other)
        let emptyGender = CatalogTermFacts(term: nil, count: 0, propertyLabel: "Gender")
        #expect(PersonGenderTreatment.read(summary(gender: emptyGender, sex: sex("male", "Male")))?.family == .man)
    }

    @Test func nothingUsableIsNeutral() {
        #expect(PersonGenderTreatment.read(summary()) == nil)
        #expect(PersonGenderTreatment.read(nil) == nil)
        #expect(PersonGenderTreatment.read(summary(sex: sex("female", "Female", count: 2))) == nil)
        let noTerm = CatalogTermFacts(term: nil, count: 1, propertyLabel: "Gender")
        #expect(PersonGenderTreatment.read(summary(gender: noTerm)) == nil)
        #expect(PersonGenderTreatment.read(summary(gender: gender("", ""))) == nil)
    }

    @Test func badgeAndSpokenText() {
        let woman = PersonGenderTreatment.read(summary(gender: gender("woman", "Woman")))
        #expect(woman?.badge.text == "Woman")
        #expect(woman?.badge.accessibilityLabel == "Gender, Woman")
        let male = PersonGenderTreatment.read(summary(sex: sex("male", "Male")))
        #expect(male?.badge.text == "Male · at birth")
        #expect(male?.badge.accessibilityLabel == "Sex at birth, Male")
        let relabeled = PersonGenderTreatment.read(summary(gender: facts("woman", "Woman", propertyLabel: "Gender identity")))
        #expect(relabeled?.badge.accessibilityLabel == "Gender identity, Woman")
    }
}
