import Foundation
import Testing
@testable import Provenencia

@Suite
struct PlaceTitleDisplayTests {
    private func parts(
        names: [String] = [],
        label: String = "",
        ref: String = "PLC-1"
    ) -> PlaceTitleParts {
        PlaceTitleParts(names: names, label: label, ref: ref)
    }

    @Test func firstNameWins() {
        let title = PlaceTitleDisplay.titleSource(parts(names: [" Montréal ", "Montreal"], label: "Home", ref: "PLC-9"))
        #expect(title == .name("Montréal"))
        #expect(PlaceTitleDisplay.extraNameCount(parts(names: ["Montréal", "Montreal"])) == 1)
    }

    @Test func labelWhenUnnamed() {
        let title = PlaceTitleDisplay.titleSource(parts(label: " Home ", ref: "PLC-9"))
        #expect(title == .label("Home"))
        #expect(PlaceTitleDisplay.extraNameCount(parts(label: "Home")) == 0)
    }

    @Test func refWhenBare() {
        let title = PlaceTitleDisplay.titleSource(parts(ref: " PLC-9 "))
        #expect(title == .ref("PLC-9"))
        #expect(PlaceTitleDisplay.extraNameCount(parts()) == 0)
    }

    @Test func chainStaysEmpty() {
        #expect(PlaceTitleDisplay.chain(parts(names: ["York", "Toronto"], label: "Home")) == "")
    }
}
