import SwiftUI
import Testing
@testable import Provenencia

@MainActor
struct PVDisclosureButtonTests {
    @Test func chevronPointsWhereTheSectionWillGo() {
        #expect(PVExpandedState.chevron(isExpanded: false) == .chevronDown)
        #expect(PVExpandedState.chevron(isExpanded: true) == .chevronUp)
    }

    @Test func toggleFlipsTheCallersState() {
        final class Box { var open = false }
        let box = Box()
        let button = PVDisclosureButton(
            "Why",
            isExpanded: Binding(get: { box.open }, set: { box.open = $0 })
        )
        button.toggle()
        #expect(box.open)
        button.toggle()
        #expect(!box.open)
    }

    @Test func voiceOverHearsTheState() {
        #expect(L10n.string(PVExpandedState.spoken(isExpanded: true)) == "Expanded")
        #expect(L10n.string(PVExpandedState.spoken(isExpanded: false)) == "Collapsed")
        #expect(L10n.string(L10n.DesignSystem.disclosureState) == "State")
    }
}
