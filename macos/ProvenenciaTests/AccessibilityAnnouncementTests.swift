import Foundation
import Testing
@testable import Provenencia

@Suite
struct AccessibilityAnnouncementTests {
    @Test func joinsTitleAndMessage() {
        #expect(AccessibilityAnnouncement.text(title: "Evidence graph added", message: nil) == "Evidence graph added")
        #expect(
            AccessibilityAnnouncement.text(
                title: "Couldn't save the evidence graph",
                message: "The name is taken."
            ) == "Couldn't save the evidence graph, The name is taken."
        )
        #expect(
            AccessibilityAnnouncement.text(title: "Couldn't save the evidence graph", message: "The name is taken.")
                == L10n.DesignSystem.announcementJoin("Couldn't save the evidence graph", "The name is taken.")
        )
    }

    @Test func nothingToSayIsNil() {
        #expect(AccessibilityAnnouncement.text(title: nil, message: "") == nil)
        #expect(AccessibilityAnnouncement.text(title: nil, message: nil) == nil)
    }

    @Test func dangerIsHigh() {
        #expect(AccessibilityAnnouncement.priority(danger: true) == .high)
        #expect(AccessibilityAnnouncement.priority(danger: false) == .medium)
    }

    @Test func copyHasAString() {
        #expect(("Jones" as String).pvString == "Jones")
        #expect(L10n.Sources.graphFormSaveErrorTitle.pvString == L10n.string(L10n.Sources.graphFormSaveErrorTitle))
        #expect(AttributedString("Smith").pvString == "Smith")
    }
}
