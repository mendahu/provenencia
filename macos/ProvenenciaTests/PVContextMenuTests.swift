import CoreGraphics
import Foundation
import Testing
@testable import Provenencia

@Suite
struct PVContextMenuStateTests {
    @Test func presentSetsOriginAndOpens() {
        var state = PVContextMenuState()
        state.present(at: CGPoint(x: 12, y: 34))
        #expect(state.isPresented)
        #expect(state.origin == CGPoint(x: 12, y: 34))
    }

    @Test func presentReplacesOriginWhileOpen() {
        var state = PVContextMenuState()
        state.present(at: CGPoint(x: 1, y: 2))
        state.present(at: CGPoint(x: 9, y: 8))
        #expect(state.isPresented)
        #expect(state.origin == CGPoint(x: 9, y: 8))
    }

    @Test func dismissClosesWithoutClearingOrigin() {
        var state = PVContextMenuState()
        state.present(at: CGPoint(x: 4, y: 5))
        state.dismiss()
        #expect(!state.isPresented)
        #expect(state.origin == CGPoint(x: 4, y: 5))
    }
}

@Suite
struct PVContextMenuKeyboardTests {
    @Test func inactiveHasNoRowsOrHighlight() {
        #expect(PVContextMenuKeyboard.inactive.itemCount == 0)
        #expect(PVContextMenuKeyboard.inactive.activeIndex == -1)
        #expect(PVContextMenuKeyboard.inactive.itemTitles.isEmpty)
    }

    @Test func typeSelectJumpsToMatchingTitle() {
        var nav = PVContextMenuKeyboard(
            itemCount: 4,
            activeIndex: 0,
            itemTitles: ["(none)", "Given name", "Surname", "Suffix"]
        )
        let moved = nav.applyTypeSelect("s")
        #expect(moved)
        #expect(nav.activeIndex == 2)
    }

    @Test func typeSelectCyclesRepeatedSingleLetters() {
        var nav = PVContextMenuKeyboard(
            itemCount: 3,
            activeIndex: 0,
            itemTitles: ["Prefix", "Suffix", "Surname"]
        )
        let start = Date()
        let first = nav.applyTypeSelect("s", now: start)
        #expect(first)
        #expect(nav.activeIndex == 1)
        let next = start.addingTimeInterval(PVTypeSelectMatcher.resetInterval + 0.01)
        let second = nav.applyTypeSelect("s", now: next)
        #expect(second)
        #expect(nav.activeIndex == 2)
    }

    @Test func typeSelectAccumulatesAPrefix() {
        var nav = PVContextMenuKeyboard(
            itemCount: 3,
            activeIndex: 0,
            itemTitles: ["Given name", "Suffix", "Surname"]
        )
        let start = Date()
        let first = nav.applyTypeSelect("s", now: start)
        #expect(first)
        #expect(nav.activeIndex == 1)
        let second = nav.applyTypeSelect("u", now: start.addingTimeInterval(0.1))
        #expect(second)
        #expect(nav.activeIndex == 1)
        let third = nav.applyTypeSelect("r", now: start.addingTimeInterval(0.2))
        #expect(third)
        #expect(nav.activeIndex == 2)
    }

    @Test func typeSelectIsANoOpWithoutTitles() {
        var nav = PVContextMenuKeyboard(itemCount: 2, activeIndex: 0)
        let moved = nav.applyTypeSelect("s")
        #expect(!moved)
        #expect(nav.activeIndex == 0)
    }
}

@Suite
struct PVContextMenuActionRegistryTests {
    @Test func activateRunsTheRegisteredAction() {
        let registry = PVContextMenuActionRegistry()
        var fired = 0
        registry.register(0) { fired += 1 }
        registry.register(2) { fired += 10 }
        registry.activate(2)
        #expect(fired == 10)
    }

    @Test func activateOnMissingIndexIsANoOp() {
        let registry = PVContextMenuActionRegistry()
        var fired = false
        registry.register(0) { fired = true }
        registry.activate(3)
        #expect(!fired)
    }

    @Test func clearRemovesPriorRegistrations() {
        let registry = PVContextMenuActionRegistry()
        var fired = false
        registry.register(0) { fired = true }
        registry.clear()
        registry.activate(0)
        #expect(!fired)
    }

    @Test func registerOverwritesTheSameIndex() {
        let registry = PVContextMenuActionRegistry()
        var which = ""
        registry.register(1) { which = "first" }
        registry.register(1) { which = "second" }
        registry.activate(1)
        #expect(which == "second")
    }
}

@Suite
struct PVContextMenuPlacementTests {
    /// AppKit screen coordinates: y grows upward. Anchor sits with its bottom
    /// at y=400 and top at y=428 (28pt tall field).
    private let anchor = CGRect(x: 100, y: 400, width: 240, height: 28)

    @Test func panelSitsAtTopLeadingOriginInsideTheAnchor() {
        let frame = PVContextMenuPlacement.panelFrame(
            anchorOnScreen: anchor,
            origin: .zero,
            size: CGSize(width: 180, height: 120)
        )
        #expect(frame.minX == anchor.minX)
        #expect(frame.maxY == anchor.maxY)
        #expect(frame.width == 180)
        #expect(frame.height == 120)
    }

    @Test func panelDropsBelowTheTriggerUsingSwiftUIDownwardOrigin() {
        // PVSelect presents at (0, triggerHeight + gap) — top of menu below the field.
        let origin = CGPoint(x: 0, y: 30)
        let size = CGSize(width: 210, height: 96)
        let frame = PVContextMenuPlacement.panelFrame(
            anchorOnScreen: anchor,
            origin: origin,
            size: size
        )
        #expect(frame.minX == anchor.minX)
        #expect(frame.maxY == anchor.maxY - origin.y)
        #expect(frame.minY == anchor.maxY - origin.y - size.height)
        #expect(!frame.intersects(anchor), "menu should clear the trigger when origin is below it")
    }

    @Test func panelFollowsAHorizontalOriginOffset() {
        let origin = CGPoint(x: 40, y: 0)
        let frame = PVContextMenuPlacement.panelFrame(
            anchorOnScreen: anchor,
            origin: origin,
            size: CGSize(width: 180, height: 80)
        )
        #expect(frame.minX == anchor.minX + origin.x)
        #expect(frame.maxY == anchor.maxY)
    }

    @Test func panelClampsDegenerateSizeToAtLeastOnePoint() {
        let frame = PVContextMenuPlacement.panelFrame(
            anchorOnScreen: anchor,
            origin: .zero,
            size: .zero
        )
        #expect(frame.width == 1)
        #expect(frame.height == 1)
    }

    @Test func panelUsesNegativeOriginAsAboveTheAnchorTop() {
        // Right-click near the top edge can report a small negative local y
        // after layout rounding; screen top should move up with it.
        let origin = CGPoint(x: 0, y: -8)
        let frame = PVContextMenuPlacement.panelFrame(
            anchorOnScreen: anchor,
            origin: origin,
            size: CGSize(width: 180, height: 40)
        )
        #expect(frame.maxY == anchor.maxY - origin.y)
    }
}
