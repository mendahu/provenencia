import SwiftUI
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
        let orders = NameFormatOrders.western
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

    private func james(nicks: [String]) -> CatalogNameValue {
        CatalogNameValue(form: "unused", parts: [CatalogNameValuePart(value: "James", type: "given")]
            + nicks.map { CatalogNameValuePart(value: $0, type: "nick") }
            + [CatalogNameValuePart(value: "Robins", type: "surname")])
    }

    private var mary: CatalogNameValue {
        CatalogNameValue(form: "unused", parts: [
            CatalogNameValuePart(value: "Mary", type: "given"),
            CatalogNameValuePart(value: "Smith", type: "former_surname"),
            CatalogNameValuePart(value: "Jones", type: "surname"),
        ])
    }

    @Test func nicknamesAreQuoted() {
        let orders = NameFormatOrders.western
        let name = james(nicks: ["Jake"])
        #expect(PersonHeaderDisplay.title(header(name: name), orders: orders) == "James “Jake” Robins")
        #expect(PersonHeaderDisplay.title(header(name: name), style: .sorted, orders: orders) == "Robins, James “Jake”")
        #expect(PersonHeaderDisplay.title(header(name: james(nicks: ["Jake", "Jim"])), orders: orders)
            == "James “Jake” “Jim” Robins")
        let conclusion = CatalogConclusionHeader.person(header(name: name))
        #expect(ConclusionHeaderDisplay.title(conclusion, orders: orders) == "James “Jake” Robins")
    }

    @Test func formerSurnameIsParenthesized() {
        let orders = NameFormatOrders.western
        #expect(PersonHeaderDisplay.title(header(name: mary), orders: orders) == "Mary (Smith) Jones")
        #expect(PersonHeaderDisplay.title(header(name: mary), style: .sorted, orders: orders) == "Jones, Mary (Smith)")
    }

    @Test func unlistedFormerSurnameFollows() {
        var orders = NameFormatOrders.western
        orders.natural.removeAll { $0.partType == "former_surname" }
        orders.sorted.removeAll { $0.partType == "former_surname" }
        #expect(PersonHeaderDisplay.title(header(name: mary), orders: orders) == "Mary Jones (Smith)")
    }

    private func emphasizedRuns(_ styled: AttributedString) -> [String] {
        styled.runs.filter { $0[AttributeScopes.FoundationAttributes.InlinePresentationIntentAttribute.self] == .emphasized }
            .map { String(styled[$0.range].characters) }
    }

    @Test func formerSurnameRunIsEmphasized() throws {
        let orders = NameFormatOrders.western
        for (style, text) in [(NameDisplayStyle.natural, "Mary (Smith) Jones"), (.sorted, "Jones, Mary (Smith)")] {
            let d = NameValueDisplay.display(for: mary, style: style, orders: orders)
            let styled = try #require(d.styled)
            #expect(d.text == text)
            #expect(String(styled.characters) == d.text)
            #expect(emphasizedRuns(styled) == ["Smith"])
        }
        #expect(NameValueDisplay.display(for: james(nicks: ["Jake"]), style: .natural, orders: orders).styled == nil)

        let d = NameValueDisplay.display(for: mary, style: .natural, orders: orders)
        #expect(PersonHeaderDisplay.titleSource(header(name: mary), orders: orders)
            == .name("Mary (Smith) Jones", styled: d.styled))
        let plain = CatalogNameValue(form: "x", parts: [
            CatalogNameValuePart(value: "James", type: "given"),
            CatalogNameValuePart(value: "Robins", type: "surname"),
        ])
        #expect(PersonHeaderDisplay.titleSource(header(name: plain), orders: orders) == .name("James Robins"))

        let italic = ConclusionListRow.italicizingEmphasis(try #require(d.styled), font: .body)
        for run in italic.runs {
            let isSmith = String(italic[run.range].characters) == "Smith"
            #expect((run[AttributeScopes.SwiftUIAttributes.FontAttribute.self] != nil) == isSmith)
        }
    }
}
