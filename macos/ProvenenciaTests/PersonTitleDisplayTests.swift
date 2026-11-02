import SwiftUI
import Testing
@testable import Provenencia

struct PersonTitleDisplayTests {
    private func summary(name: CatalogNameValue?, label: String = "", clusters: Int = 0) -> CatalogPersonSummary {
        CatalogPersonSummary(
            entity: CatalogCanonicalEntity(id: "e1", ref: "PER-7KD45", subjectTypeID: "t", label: label),
            name: name,
            nameValueCount: clusters
        )
    }

    @Test func partsWinOverTheTranscription() {
        let h = summary(name: CatalogNameValue(form: " Jeremiah Arthurt Gumtree ", parts: [
            CatalogNameValuePart(value: "Jeremiah", type: "given"),
            CatalogNameValuePart(value: "Arthur", type: "given"),
            CatalogNameValuePart(value: "Gumtree", type: "surname"),
        ]), label: "Grandpa", clusters: 1)
        #expect(PersonTitleDisplay.title(h) == "Jeremiah Arthur Gumtree")
    }

    @Test func formIsTheFallbackWhenThereAreNoParts() {
        let h = summary(name: CatalogNameValue(form: " James Robins ", parts: [
            CatalogNameValuePart(value: " ", type: "given"),
        ]), label: "Grandpa", clusters: 1)
        #expect(PersonTitleDisplay.title(h) == "James Robins")
    }

    @Test func blankFormFallsBackToParts() {
        let name = CatalogNameValue(form: "  ", parts: [
            CatalogNameValuePart(value: "James", type: "given"),
            CatalogNameValuePart(value: " ", type: ""),
            CatalogNameValuePart(value: "Robins", type: "surname"),
        ])
        #expect(NameValueDisplay.string(for: name) == "James Robins")
        #expect(PersonTitleDisplay.title(summary(name: name, clusters: 1)) == "James Robins")
    }

    @Test func sortedOrderUsesTheCatalogSeparator() {
        let name = CatalogNameValue(form: "unused", parts: [
            CatalogNameValuePart(value: "Vincent", type: "given"),
            CatalogNameValuePart(value: "van", type: "surname_prefix"),
            CatalogNameValuePart(value: "Gogh", type: "surname"),
        ])
        let orders = NameFormatOrders.western
        #expect(NameValueDisplay.string(for: name, style: .sorted, orders: orders) == "van Gogh, Vincent")
        #expect(PersonTitleDisplay.title(summary(name: name), style: .sorted, orders: orders) == "van Gogh, Vincent")
        #expect(NameValueDisplay.string(for: name, style: .natural, orders: orders) == "Vincent van Gogh")
    }

    @Test func emptyNameFallsBackToLabelThenRef() {
        let empty = CatalogNameValue(form: "", parts: [])
        #expect(PersonTitleDisplay.title(summary(name: empty, label: "Mother of James")) == "Mother of James")
        #expect(PersonTitleDisplay.title(summary(name: nil, label: "Mother of James")) == "Mother of James")
        #expect(PersonTitleDisplay.title(summary(name: nil, label: "  ")) == "PER-7KD45")
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
        #expect(PersonTitleDisplay.title(summary(name: name), orders: orders) == "James “Jake” Robins")
        #expect(PersonTitleDisplay.title(summary(name: name), style: .sorted, orders: orders) == "Robins, James “Jake”")
        #expect(PersonTitleDisplay.title(summary(name: james(nicks: ["Jake", "Jim"])), orders: orders)
            == "James “Jake” “Jim” Robins")
        let conclusion = CatalogConclusionSummary.person(summary(name: name))
        #expect(ConclusionSummaryDisplay.title(conclusion, orders: orders) == "James “Jake” Robins")
    }

    @Test func formerSurnameIsParenthesized() {
        let orders = NameFormatOrders.western
        #expect(PersonTitleDisplay.title(summary(name: mary), orders: orders) == "Mary (Smith) Jones")
        #expect(PersonTitleDisplay.title(summary(name: mary), style: .sorted, orders: orders) == "Jones, Mary (Smith)")
    }

    @Test func unlistedFormerSurnameFollows() {
        var orders = NameFormatOrders.western
        orders.natural.removeAll { $0.partType == "former_surname" }
        orders.sorted.removeAll { $0.partType == "former_surname" }
        #expect(PersonTitleDisplay.title(summary(name: mary), orders: orders) == "Mary Jones (Smith)")
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
        #expect(PersonTitleDisplay.titleSource(summary(name: mary), orders: orders)
            == .name("Mary (Smith) Jones", styled: d.styled, spoken: d.spoken))
        let plain = CatalogNameValue(form: "x", parts: [
            CatalogNameValuePart(value: "James", type: "given"),
            CatalogNameValuePart(value: "Robins", type: "surname"),
        ])
        #expect(PersonTitleDisplay.titleSource(summary(name: plain), orders: orders) == .name("James Robins"))

        let italic = ConclusionListRow.italicizingEmphasis(try #require(d.styled), font: .body)
        for run in italic.runs {
            let isSmith = String(italic[run.range].characters) == "Smith"
            #expect((run[AttributeScopes.SwiftUIAttributes.FontAttribute.self] != nil) == isSmith)
        }
    }

    @Test func spokenNameMarksDecoratedParts() {
        let orders = NameFormatOrders.western
        let jake = NameValueDisplay.display(for: james(nicks: ["Jake"]), style: .natural, orders: orders)
        #expect(jake.spoken == "James Jake, nickname Robins")
        let m = NameValueDisplay.display(for: mary, style: .natural, orders: orders)
        #expect(m.spoken == "Mary former surname Smith Jones")
        #expect(NameValueDisplay.display(for: james(nicks: []), style: .natural, orders: orders).spoken == nil)
        let source = PersonTitleDisplay.titleSource(name: mary, entity: summary(name: mary).entity, orders: orders)
        #expect(source.text == "Mary (Smith) Jones")
        #expect(ConclusionListRow.accessibilityLabel(source, ref: "PER-7KD45").contains("former surname Smith"))
    }
}
