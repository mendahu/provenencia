import CoreGraphics
import Testing
@testable import Provenencia

@Suite
struct PVSelectPointerTrackingTests {
    private let trigger = CGRect(x: 0, y: 100, width: 180, height: 28)
    private let rows = [
        CGRect(x: 0, y: 60, width: 210, height: 20),
        CGRect(x: 0, y: 40, width: 210, height: 20),
        CGRect(x: 0, y: 20, width: 210, height: 20),
    ]

    private func partSession() -> PVSelectSession {
        PVSelectSession(
            options: [
                PVSelectOption(value: "given", label: "Given name"),
                PVSelectOption(value: "surname", label: "Surname"),
                PVSelectOption(value: "suffix", label: "Suffix"),
            ],
            selection: "given"
        )
    }

    @Test func hitTestPrefersARowOverTheTrigger() {
        let target = PVSelectPointerTracking.hitTest(
            point: CGPoint(x: 10, y: 50),
            trigger: trigger,
            rows: rows
        )
        #expect(target == .row(1))
    }

    @Test func hitTestReturnsTriggerWhenThePointIsOnTheField() {
        let target = PVSelectPointerTracking.hitTest(
            point: CGPoint(x: 10, y: 110),
            trigger: trigger,
            rows: rows
        )
        #expect(target == .trigger)
    }

    @Test func hitTestMissIsOutside() {
        let target = PVSelectPointerTracking.hitTest(
            point: CGPoint(x: 400, y: 400),
            trigger: trigger,
            rows: rows
        )
        #expect(target == .outside)
    }

    @Test func rowFramesSliceThePanelTopDown() {
        let panel = CGRect(x: 0, y: 0, width: 200, height: 80)
        let frames = PVSelectPointerTracking.rowFrames(panel: panel, count: 2, padding: 0)
        #expect(frames.count == 2)
        #expect(frames[0].maxY == panel.maxY)
        #expect(frames[1].minY == panel.minY)
    }

    @Test func pressOnTriggerOpensWithCommittedHighlight() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        #expect(nav.isOpen)
        #expect(nav.highlightIndex == 0)
        #expect(nav.selection == "given")
    }

    @Test func pressOnTriggerWhenDisabledStaysClosed() {
        var nav = partSession()
        nav.isDisabled = true
        nav.handlePointer(.press(.trigger))
        #expect(!nav.isOpen)
    }

    @Test func dragOverARowMovesHighlightWithoutCommit() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.drag(.row(2)))
        #expect(nav.highlightIndex == 2)
        #expect(nav.selection == "given")
    }

    @Test func dragLeavingTheMenuClearsHighlight() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.drag(.row(1)))
        nav.handlePointer(.drag(.outside))
        #expect(nav.highlightIndex == -1)
        #expect(nav.selection == "given")
        #expect(nav.isOpen)
    }

    @Test func releaseOnARowCommitsAndCloses() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.drag(.row(1)))
        nav.handlePointer(.release(.row(1)))
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
    }

    @Test func releaseOnTriggerLeavesTheMenuOpen() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.release(.trigger))
        #expect(nav.isOpen)
        #expect(nav.selection == "given")
    }

    @Test func releaseOutsideRestoresTheSnapshot() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.drag(.row(2)))
        nav.handlePointer(.release(.outside))
        #expect(!nav.isOpen)
        #expect(nav.selection == "given")
    }

    @Test func clickAwayWhileOpenMatchesEscape() {
        var nav = partSession()
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.drag(.row(2)))
        nav.handlePointer(.clickAway)
        #expect(!nav.isOpen)
        #expect(nav.selection == "given")
    }
}
