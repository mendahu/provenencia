import Foundation
import Testing
@testable import Provenencia

@Suite
struct ObservationValueDisplayTests {
    private let enUS = Locale(identifier: "en_US")

    @Test func textUsesValueText() {
        let obs = base(valueText: "Boston", propertyValueType: "text")
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "Boston")
    }

    @Test func integerUsesValueInteger() {
        var obs = base(propertyValueType: "integer")
        obs.valueInteger = 42
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "42")
    }

    @Test func datePrefersStructuredLocaleFormat() {
        var obs = base(propertyValueType: "date")
        obs.date = CatalogDateValueInput(
            kind: "point",
            qualifier: "ABT",
            startYear: 1882,
            startMonth: 4
        )
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "About Apr 1882")
    }

    @Test func dateFallsBackToValueText() {
        let obs = base(valueText: "1882-04-03", propertyValueType: "date")
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "1882-04-03")
    }

    @Test func nameUsesFormWhenValueTextEmpty() {
        var obs = base(propertyValueType: "name")
        obs.nameForm = "Ada Lovelace"
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "Ada Lovelace")
    }

    @Test func namePartsWinOverTheDenormalizedForm() {
        var obs = base(valueText: "Jeremiah Arthurt Gumtree", propertyValueType: "name")
        obs.nameForm = "Jeremiah Arthurt Gumtree"
        obs.nameParts = [
            CatalogNameValuePart(value: "Jeremiah", type: "given"),
            CatalogNameValuePart(value: "Arthur", type: "given"),
            CatalogNameValuePart(value: "Gumtree", type: "surname"),
        ]
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "Jeremiah Arthur Gumtree")
    }

    @Test func nameUsesTheNaturalOrder() {
        var obs = base(valueText: "unused", propertyValueType: "name")
        obs.nameForm = "unused"
        obs.nameParts = [
            CatalogNameValuePart(value: "Vincent", type: "given"),
            CatalogNameValuePart(value: "van", type: "surname_prefix"),
            CatalogNameValuePart(value: "Gogh", type: "surname"),
        ]
        let orders = NameFormatOrders(
            natural: [
                NameFormatPart(partType: "given", trailingSeparator: ""),
                NameFormatPart(partType: "surname_prefix", trailingSeparator: ""),
                NameFormatPart(partType: "surname", trailingSeparator: ""),
            ],
            sorted: [
                NameFormatPart(partType: "surname", trailingSeparator: ", "),
                NameFormatPart(partType: "given", trailingSeparator: ""),
            ]
        )
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS, orders: orders) == "Vincent van Gogh")
    }

    @Test func emptyIsEmptyString() {
        let obs = base(propertyValueType: "date")
        #expect(ObservationValueDisplay.string(for: obs, locale: enUS) == "")
    }

    private func base(
        valueText: String = "",
        propertyValueType: String
    ) -> CatalogObservation {
        CatalogObservation(
            id: "obs-1",
            ref: "OBS-1",
            citationID: "cit-1",
            subjectID: "sub-1",
            propertyID: "prop-1",
            polarity: "positive",
            valueText: valueText,
            valueInteger: nil,
            valueDateID: "",
            valueNameID: "",
            valueSubjectID: "",
            valueTermID: "",
            propertyKey: "k",
            propertyLabel: "Label",
            propertyValueType: propertyValueType
        )
    }
}
