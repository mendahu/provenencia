import Testing
@testable import Provenencia

struct FocusKeeperTests {
    @Test func keepsTheStopAcrossAWidthChangeAndAReload() {
        var keeper = FocusKeeper<String>()
        keeper.took("b", fallback: "row", revealedPending: nil)
        #expect(keeper.widthStop(current: nil, isShown: { _ in true }) == "b")
        #expect(keeper.widthStop(current: "b", isShown: { _ in true }) == "b")
        #expect(keeper.widthStop(current: nil, isShown: { $0 != "b" }) == "row")

        // A reload took the card: its row takes the keyboard.
        #expect(keeper.lost(isShown: { _ in false }) == .fallback("row"))

        // A click elsewhere: nothing is pulled back, and nothing is handed on a resize.
        #expect(keeper.lost(isShown: { _ in true }) == .none)
        keeper.forgetLast()
        #expect(keeper.widthStop(current: nil, isShown: { _ in true }) == nil)

        // A stop with no fallback that is gone has nowhere to go.
        keeper.took("t", fallback: nil, revealedPending: nil)
        #expect(keeper.lost(isShown: { _ in false }) == .none)
        #expect(keeper.widthStop(current: nil, isShown: { _ in false }) == nil)
    }

    @Test func pendingStopSurvivesItsOwnLandingAndEndsOnAnotherStop() {
        var keeper = FocusKeeper<String>()
        keeper.give("x")
        #expect(keeper.pending == "x")
        #expect(keeper.lost(isShown: { _ in true }) == .reclaim)
        keeper.took("z", fallback: nil, revealedPending: "x")
        #expect(keeper.pending == nil)

        keeper.give(".useAs")
        keeper.took(".image", fallback: nil, revealedPending: ".image")
        #expect(keeper.pending == ".useAs")
        keeper.took(".useAs", fallback: nil, revealedPending: ".image")
        #expect(keeper.pending == ".useAs")

        keeper.give("x")
        keeper.expire("y")
        #expect(keeper.pending == "x")
        keeper.expire("x")
        #expect(keeper.pending == nil)
    }
}
