import Foundation
import Testing
@testable import Provenencia

struct PersonTileTests {
    private let en = Locale(identifier: "en_US")

    private func place(_ name: String) -> CatalogChainedPlace {
        CatalogChainedPlace(
            entity: CatalogCanonicalEntity(id: "pl-\(name)", ref: "PLC-\(name)", subjectTypeID: "t", label: ""),
            names: [name]
        )
    }

    private func summary(
        id: String = "e1",
        ref: String = "PER-7KD45",
        name: String? = nil,
        birth: CatalogLifeFacts = CatalogLifeFacts(),
        death: CatalogLifeFacts = CatalogLifeFacts()
    ) -> CatalogPersonSummary {
        CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: id, ref: ref, subjectTypeID: "t", label: ""),
            name: name.map { CatalogNameValue(form: $0) },
            nameValueCount: name == nil ? 0 : 1,
            birth: birth,
            death: death
        )
    }

    private func date(_ year: Int32, _ month: Int32? = nil, _ day: Int32? = nil) -> CatalogDateValueInput {
        CatalogDateValueInput(kind: "point", startYear: year, startMonth: month, startDay: day)
    }

    private var anna: CatalogPersonSummary {
        var h = summary(
            id: "anna", name: "Anna Weber",
            birth: CatalogLifeFacts(date: date(1790, 5, 3), dateCount: 1, places: [place("Bremen")]),
            death: CatalogLifeFacts(date: date(1851), dateCount: 1)
        )
        h.gender = CatalogTermFacts(
            term: .term(id: "t-woman", key: "woman", label: "Woman", origin: CatalogOrigin.provenencia),
            count: 1, propertyLabel: "Gender"
        )
        return h
    }

    @Test func tileContentFromSummary() {
        let tile = PersonTileContent(summary: anna, role: "Parent", isFocus: false, locale: en)
        #expect(tile.id == "anna")
        #expect(tile.title.text == "Anna Weber")
        #expect(tile.lines.map(\.abbreviation) == ["b.", "d."])
        #expect(tile.lines.map(\.date) == ["3 May 1790", "1851"])
        #expect(tile.lines.map(\.place) == ["Bremen", nil])
        #expect(tile.colors != nil)

        let wilhelm = PersonTileContent(summary: summary(id: "wilhelm", name: "Wilhelm Weber"), role: nil, isFocus: false, locale: en)
        #expect(wilhelm.lines.isEmpty && wilhelm.colors == nil && wilhelm.imageRelPath.isEmpty)

        let child = PersonTileContent(summary: summary(id: "kid", ref: "PER-7KQ2M"), role: "Child", isFocus: false, locale: en)
        #expect(child.title == .ref("PER-7KQ2M"))
    }

    @Test func spokenLabels() {
        #expect(PersonTileContent(summary: anna, role: "Parent", isFocus: false, locale: en).spokenLabel
            == "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851")
        #expect(PersonTileContent(summary: anna, role: nil, isFocus: false, locale: en).spokenLabel
            == "Anna Weber, born 3 May 1790, Bremen, died 1851")
        let johann = summary(
            id: "johann", name: "Johann Weber",
            birth: CatalogLifeFacts(date: date(1817, 3, 12), dateCount: 1, places: [place("Bremen")]),
            death: CatalogLifeFacts(date: date(1880), dateCount: 1, places: [place("Hamburg")])
        )
        #expect(PersonTileContent(summary: johann, role: nil, isFocus: true, locale: en).spokenLabel
            == "Johann Weber, born 12 Mar 1817, Bremen, died 1880, Hamburg, this person")
        #expect(PersonTileContent(summary: summary(name: "Wilhelm Weber"), role: "Sibling", isFocus: false, locale: en).spokenLabel
            == "Sibling, Wilhelm Weber")
    }
}
