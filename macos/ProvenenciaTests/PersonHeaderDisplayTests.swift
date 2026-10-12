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

    @Test func partsWinOverTheTranscription() {
        let h = header(name: CatalogNameValue(form: " Jeremiah Arthurt Gumtree ", parts: [
            CatalogNameValuePart(value: "Jeremiah", type: "given"),
            CatalogNameValuePart(value: "Arthur", type: "given"),
            CatalogNameValuePart(value: "Gumtree", type: "surname"),
        ]), label: "Grandpa", clusters: 1)
        #expect(PersonHeaderDisplay.title(h) == "Jeremiah Arthur Gumtree")
    }

    @Test func formIsTheFallbackWhenThereAreNoParts() {
        let h = header(name: CatalogNameValue(form: " James Robins ", parts: [
            CatalogNameValuePart(value: " ", type: "given"),
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

    @Test func sortedOrderUsesTheCatalogSeparator() {
        let name = CatalogNameValue(form: "unused", parts: [
            CatalogNameValuePart(value: "Vincent", type: "given"),
            CatalogNameValuePart(value: "van", type: "surname_prefix"),
            CatalogNameValuePart(value: "Gogh", type: "surname"),
        ])
        let orders = NameFormatOrders(
            natural: [
                NameFormatPart(partType: "given", trailingSeparator: ""),
                NameFormatPart(partType: "surname_prefix", trailingSeparator: ""),
                NameFormatPart(partType: "surname", trailingSeparator: ""),
            ],
            sorted: [
                NameFormatPart(partType: "surname_prefix", trailingSeparator: ""),
                NameFormatPart(partType: "surname", trailingSeparator: ", "),
                NameFormatPart(partType: "given", trailingSeparator: ""),
            ]
        )
        #expect(NameValueDisplay.string(for: name, style: .sorted, orders: orders) == "van Gogh, Vincent")
        #expect(PersonHeaderDisplay.title(header(name: name), style: .sorted, orders: orders) == "van Gogh, Vincent")
        #expect(NameValueDisplay.string(for: name, style: .natural, orders: orders) == "Vincent van Gogh")
    }

    @Test func emptyNameFallsBackToLabelThenRef() {
        let empty = CatalogNameValue(form: "", parts: [])
        #expect(PersonHeaderDisplay.title(header(name: empty, label: "Mother of James")) == "Mother of James")
        #expect(PersonHeaderDisplay.title(header(name: nil, label: "Mother of James")) == "Mother of James")
        #expect(PersonHeaderDisplay.title(header(name: nil, label: "  ")) == "PER-7KD45")
    }
}
