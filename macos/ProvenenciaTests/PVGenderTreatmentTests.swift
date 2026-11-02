import Testing
@testable import Provenencia

struct PVGenderTreatmentTests {
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
        let woman = PVGenderTreatment.read(summary(gender: gender("woman", "Woman"), sex: female))
        #expect(woman?.family == .woman && woman?.basis == .gender)
        let man = PVGenderTreatment.read(summary(gender: gender("man", "Man"), sex: female))
        #expect(man?.family == .man && man?.basis == .gender)
        #expect(PVGenderTreatment.read(summary(gender: gender("non_binary", "Non-binary")))?.family == .nonBinary)
        #expect(PVGenderTreatment.read(summary(gender: gender("two_spirit", "Two-spirit")))?.family == .other)
    }

    @Test func aResearchersTermKeyedLikeAProductOneKeepsTheOtherFamily() {
        let theirs = gender("woman", "Woman", origin: CatalogOrigin.user)
        #expect(PVGenderTreatment.read(summary(gender: theirs))?.family == .other)
        let theirSex = sex("female", "Female", origin: CatalogOrigin.user)
        #expect(PVGenderTreatment.read(summary(sex: theirSex))?.family == .other)
    }

    @Test func mixedGenderIsNeutralAndDoesNotFallBack() {
        let mixed = gender("woman", "Woman", count: 2)
        #expect(PVGenderTreatment.read(summary(gender: mixed, sex: sex("female", "Female"))) == nil)
    }

    @Test func sexAtBirthBorrowsAFamily() {
        let female = PVGenderTreatment.read(summary(sex: sex("female", "Female")))
        #expect(female?.family == .woman && female?.basis == .sexAtBirth)
        #expect(PVGenderTreatment.read(summary(sex: sex("male", "Male")))?.family == .man)
        #expect(PVGenderTreatment.read(summary(sex: sex("intersex", "Intersex")))?.family == .other)
        let emptyGender = CatalogTermFacts(term: nil, count: 0, propertyLabel: "Gender")
        #expect(PVGenderTreatment.read(summary(gender: emptyGender, sex: sex("male", "Male")))?.family == .man)
    }

    @Test func nothingUsableIsNeutral() {
        #expect(PVGenderTreatment.read(summary()) == nil)
        #expect(PVGenderTreatment.read(nil) == nil)
        #expect(PVGenderTreatment.read(summary(sex: sex("female", "Female", count: 2))) == nil)
        let noTerm = CatalogTermFacts(term: nil, count: 1, propertyLabel: "Gender")
        #expect(PVGenderTreatment.read(summary(gender: noTerm)) == nil)
        #expect(PVGenderTreatment.read(summary(gender: gender("", ""))) == nil)
    }

    @Test func aRelabelledProductTermReadsLikeTheRowsString() {
        let woman = CatalogConclusionValue.term(id: "t", key: "woman", label: "Female (self-described)", origin: CatalogOrigin.provenencia)
        let treatment = PVGenderTreatment.read(summary(gender: CatalogTermFacts(term: woman, count: 1, propertyLabel: "Gender")))
        #expect(treatment?.text == ReconciledValueDisplay.string(for: woman, propertyKey: SeededPropertyKey.gender))
        #expect(treatment?.text == "Woman")
    }

    @Test func badgeAndSpokenText() throws {
        func badge(_ treatment: PVGenderTreatment?) throws -> ConclusionDetailBadge {
            PersonDetailContent.badge(try #require(treatment))
        }
        let woman = try badge(PVGenderTreatment.read(summary(gender: gender("woman", "Woman"))))
        #expect(woman.text == "Woman")
        #expect(woman.accessibilityLabel == "Gender, Woman")
        let male = try badge(PVGenderTreatment.read(summary(sex: sex("male", "Male"))))
        #expect(male.text == "Male · at birth")
        #expect(male.accessibilityLabel == "Sex at birth, Male")
        let relabeled = try badge(PVGenderTreatment.read(summary(gender: facts("woman", "Woman", propertyLabel: "Gender identity"))))
        #expect(relabeled.accessibilityLabel == "Gender identity, Woman")
    }
}
