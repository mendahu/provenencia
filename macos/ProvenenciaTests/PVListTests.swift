import SwiftUI
import Testing
@testable import Provenencia

struct PVListTests {
    private func action(_ key: KeyEquivalent, option: Bool = false, at index: Int, of count: Int = 25) -> PVListKeyAction {
        PVListKeyboard.action(for: key, option: option, index: index, count: count)
    }

    @Test func arrowsMoveOneAndClampWithoutWrapping() {
        #expect(action(.downArrow, at: 3) == .move(to: 4))
        #expect(action(.upArrow, at: 3) == .move(to: 2))
        #expect(action(.downArrow, at: 24) == .move(to: 24))
        #expect(action(.upArrow, at: 0) == .move(to: 0))
    }

    @Test func optionArrowsAndHomeEndJumpToTheEnds() {
        #expect(action(.downArrow, option: true, at: 3) == .move(to: 24))
        #expect(action(.upArrow, option: true, at: 20) == .move(to: 0))
        #expect(action(.home, at: 12) == .move(to: 0))
        #expect(action(.end, at: 12) == .move(to: 24))
    }

    @Test func pageKeysMoveByTenAndClamp() {
        #expect(action(.pageDown, at: 3) == .move(to: 13))
        #expect(action(.pageDown, at: 20) == .move(to: 24))
        #expect(action(.pageUp, at: 13) == .move(to: 3))
        #expect(action(.pageUp, at: 4) == .move(to: 0))
    }

    @Test func returnAndSpaceActivateTheFocusedRow() {
        #expect(action(.return, at: 7) == .activate(7))
        #expect(action(.space, at: 0) == .activate(0))
    }

    @Test func otherKeysAndEmptyListsAreIgnored() {
        #expect(action(KeyEquivalent("a"), at: 2) == .ignore)
        #expect(action(.downArrow, at: 0, of: 0) == .ignore)
        #expect(action(.return, at: 0, of: 0) == .ignore)
    }

    @Test func densityPaddingMatchesTheKit() {
        #expect(PVListDensity.comfortable.vertical == 11)
        #expect(PVListDensity.comfortable.horizontal == 16)
        #expect(PVListDensity.compact.vertical == 7)
        #expect(PVListDensity.compact.horizontal == 12)
    }
}
