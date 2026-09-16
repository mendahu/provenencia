import Foundation
import Testing
@testable import Provenencia

/// `WindowDrag` is shared static state, so every test here starts from its
/// own `pressBegan()` and the bodies stay synchronous — a `@MainActor` test
/// with no suspension point cannot interleave with another suite's.
@MainActor
@Suite
struct WindowDragTests {
    @Test func pressThatDidNotMoveTheWindowRunsTheAction() {
        WindowDrag.pressBegan()

        var ran = false
        WindowDrag.unlessDragging { ran = true }

        #expect(ran)
    }

    @Test func pressThatMovedTheWindowSkipsTheAction() {
        WindowDrag.pressBegan()
        WindowDrag.windowMovedUnderPress()

        var ran = false
        WindowDrag.unlessDragging { ran = true }

        #expect(!ran)
    }

    @Test func dragSuppressesEveryActionOfTheSamePress() {
        WindowDrag.pressBegan()
        WindowDrag.windowMovedUnderPress()

        // A drag off Back fires the long-press jump menu mid-drag and the
        // button action on the mouse-up that ends it. One press, both skipped.
        var longPressOpenedMenu = false
        var buttonStepped = false
        WindowDrag.unlessDragging { longPressOpenedMenu = true }
        WindowDrag.unlessDragging { buttonStepped = true }

        #expect(!longPressOpenedMenu)
        #expect(!buttonStepped)
    }

    @Test func nextPressClearsThePreviousDrag() {
        WindowDrag.pressBegan()
        WindowDrag.windowMovedUnderPress()
        #expect(WindowDrag.didDrag)

        WindowDrag.pressBegan()

        var ran = false
        WindowDrag.unlessDragging { ran = true }

        #expect(ran)
    }
}
