import Testing
@testable import Provenencia

@Suite
@MainActor
struct PVSegmentedControlTests {
    private func items(disabled: Set<String> = []) -> [PVSegmentedControl<String>.Item] {
        ["a", "b", "c"].map {
            .init(text: $0, value: $0, isDisabled: disabled.contains($0), accessibilityIdentifier: $0)
        }
    }

    @Test func focusLandsOnTheSelectedSegment() {
        #expect(PVSegmentedControl<String>.focusTarget(items: items(), selection: "c") == "c")
        #expect(PVSegmentedControl<String>.focusTarget(items: items(disabled: ["b"]), selection: "b") == "a")
        #expect(PVSegmentedControl<String>.focusTarget(items: items(disabled: ["a"]), selection: "a") == "b")
        #expect(PVSegmentedControl<String>.focusTarget(items: items(disabled: ["a", "b", "c"]), selection: "a") == nil)
    }

    @Test func arrowsMoveToTheNextEnabledSegment() {
        typealias C = PVSegmentedControl<String>
        #expect(C.arrowTarget(items: items(), selection: "b", forward: true) == "c")
        #expect(C.arrowTarget(items: items(), selection: "b", forward: false) == "a")
        #expect(C.arrowTarget(items: items(), selection: "c", forward: true) == nil)
        #expect(C.arrowTarget(items: items(), selection: "a", forward: false) == nil)
        #expect(C.arrowTarget(items: items(disabled: ["b"]), selection: "a", forward: true) == "c")
        #expect(C.arrowTarget(items: items(disabled: ["a", "b", "c"]), selection: "a", forward: true) == nil)
    }
}
