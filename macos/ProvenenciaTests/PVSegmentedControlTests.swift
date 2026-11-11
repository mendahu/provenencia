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
}
