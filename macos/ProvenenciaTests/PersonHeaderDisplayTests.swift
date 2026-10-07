import Testing
@testable import Provenencia

struct PersonHeaderDisplayTests {
    private func header(name: CatalogNameValue?, label: String = "", clusters: Int = 0) -> CatalogPersonHeader {
        CatalogPersonHeader(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: label),
            name: name,
            nameValueCount: clusters
        )
    }

    @Test func nameFormWins() {
        let h = header(name: CatalogNameValue(form: " James Robins ", parts: [
            CatalogNameValuePart(value: "Jim", type: "given"),
        ]), label: "Grandpa", clusters: 1)
        #expect(PersonHeaderDisplay.title(h) == "James Robins")
    }

    @Test func blankFormFallsBackToParts() {
        let name = CatalogNameValue(form: "  ", parts: [
            CatalogNameValuePart(value: "James", type: "given"),
            CatalogNameValuePart(value: " ", type: ""),
            CatalogNameValuePart(value: "Robins", type: "surname"),
        ])
        #expect(NameValueDisplay.string(for: name) == "James Robins")
        #expect(PersonHeaderDisplay.title(header(name: name, clusters: 1)) == "James Robins")
    }

    @Test func emptyNameFallsBackToLabelThenRef() {
        let empty = CatalogNameValue(form: "", parts: [])
        #expect(PersonHeaderDisplay.title(header(name: empty, label: "Mother of James")) == "Mother of James")
        #expect(PersonHeaderDisplay.title(header(name: nil, label: "Mother of James")) == "Mother of James")
        #expect(PersonHeaderDisplay.title(header(name: nil, label: "  ")) == "PER-7KD45")
    }
}
