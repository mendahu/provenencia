import Foundation
import Testing
@testable import Provenencia

/// `WindowDrag` is shared static state, so every test here starts from its
/// own `beginPress()` and the bodies stay synchronous — a `@MainActor` test
/// with no suspension point cannot interleave with another suite's.
@MainActor
@Suite
struct WindowDragTests {
    @Test func pressWithoutMovementRunsTheAction() {
        WindowDrag.beginPress()

        var ran = false
        WindowDrag.unlessDragging { ran = true }

        #expect(ran)
    }

    @Test func pressThatMovedTheWindowSkipsTheAction() {
        WindowDrag.beginPress()
        WindowDrag.noteWindowMoved()

        var ran = false
        WindowDrag.unlessDragging { ran = true }

        #expect(!ran)
    }

    @Test func dragSuppressesEveryActionOfTheSamePress() {
        WindowDrag.beginPress()
        WindowDrag.noteWindowMoved()

        // The long-press jump menu fires mid-drag, the button action fires on
        // the mouse-up that ends it. Both belong to the one press.
        var longPressOpenedMenu = false
        var buttonStepped = false
        WindowDrag.unlessDragging { longPressOpenedMenu = true }
        WindowDrag.unlessDragging { buttonStepped = true }

        #expect(!longPressOpenedMenu)
        #expect(!buttonStepped)
    }

    @Test func nextPressClearsThePreviousDrag() {
        WindowDrag.beginPress()
        WindowDrag.noteWindowMoved()
        #expect(WindowDrag.didDrag)

        WindowDrag.beginPress()

        var ran = false
        WindowDrag.unlessDragging { ran = true }

        #expect(ran)
    }
}
