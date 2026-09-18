import Foundation
import Testing
@testable import Provenencia

@Suite
struct PVFloatingMenuSelectionTests {
    @Test func emptyCountReturnsNone() {
        #expect(PVFloatingMenuSelection.moveIndex(from: 0, delta: 1, count: 0) == -1)
        #expect(PVFloatingMenuSelection.moveIndex(from: -1, delta: -1, count: 0) == -1)
    }

    @Test func fromNoneOpensAtEndMatchingDelta() {
        #expect(PVFloatingMenuSelection.moveIndex(from: -1, delta: 1, count: 5) == 0)
        #expect(PVFloatingMenuSelection.moveIndex(from: -1, delta: -1, count: 5) == 4)
    }

    @Test func clampsAtEnds() {
        #expect(PVFloatingMenuSelection.moveIndex(from: 0, delta: -1, count: 5) == 0)
        #expect(PVFloatingMenuSelection.moveIndex(from: 4, delta: 1, count: 5) == 4)
    }

    @Test func stepsByDelta() {
        #expect(PVFloatingMenuSelection.moveIndex(from: 2, delta: 1, count: 5) == 3)
        #expect(PVFloatingMenuSelection.moveIndex(from: 2, delta: -1, count: 5) == 1)
    }
}
