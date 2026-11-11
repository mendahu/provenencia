import Testing
@testable import Provenencia

@Suite
@MainActor
struct ConclusionPageSectionSwitchTests {
    @Test func switchSpeaksAndClaims() {
        let fromColumn = ConclusionPageSectionSwitch(
            from: .family, to: .timeline, sectionsFocused: true, switcherFocused: false)
        #expect(fromColumn?.announcement == "Timeline shown.")
        #expect(fromColumn?.claimsSwitcher == true)

        let fromSwitcher = ConclusionPageSectionSwitch(
            from: .vitals, to: .gallery, sectionsFocused: false, switcherFocused: true)
        #expect(fromSwitcher?.announcement == "Gallery shown.")
        #expect(fromSwitcher?.claimsSwitcher == true)

        let elsewhere = ConclusionPageSectionSwitch(
            from: .vitals, to: .family, sectionsFocused: false, switcherFocused: false)
        #expect(elsewhere?.announcement == "Family shown.")
        #expect(elsewhere?.claimsSwitcher == false)

        #expect(ConclusionPageSectionSwitch(
            from: .timeline, to: .timeline, sectionsFocused: true, switcherFocused: true) == nil)
    }
}
