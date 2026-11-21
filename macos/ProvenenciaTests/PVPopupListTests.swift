import Foundation
import SwiftUI
import Testing
@testable import Provenencia

@Suite
struct PVPopupListHighlightTests {
    private func option(_ value: String, _ label: String, disabled: Bool = false) -> PVComboBoxOption {
        PVComboBoxOption(value: value, label: label, isDisabled: disabled)
    }

    private var pool: [PVComboBoxOption] {
        [option("f1", "Author"), option("f2", "Publication date"), option("f3", "Publisher")]
    }

    /// One key from `index` over `rows`, and where the highlight lands.
    private func moved(
        from index: Int, _ key: KeyEquivalent, command: Bool = false, in rows: [PVComboBoxOption]
    ) -> Int {
        var highlight = PVPopupListHighlight(index: index)
        highlight.move(key, command: command, count: rows.count, isDisabled: { rows[$0].isDisabled })
        return highlight.index
    }

    private func landed(in rows: [PVComboBoxOption]) -> Int {
        var highlight = PVPopupListHighlight(index: 1)
        highlight.land(count: rows.count, isDisabled: { rows[$0].isDisabled })
        return highlight.index
    }

    // MARK: Spotlight landing

    @Test func firstEnabledSkipsDisabledLeadingRows() {
        #expect(landed(in: [option("a", "A", disabled: true), option("b", "B")]) == 1)
    }

    @Test func firstEnabledIsNegativeWhenEveryRowIsDisabled() {
        #expect(landed(in: [option("a", "A", disabled: true), option("b", "B", disabled: true)]) == -1)
        #expect(landed(in: []) == -1)
    }

    @Test func hoverTakesOnlyAnEnabledRowInRange() {
        let rows = [option("a", "A"), option("b", "B", disabled: true), option("c", "C")]
        var highlight = PVPopupListHighlight(index: 0)
        highlight.hover(at: 1, count: rows.count, isDisabled: { rows[$0].isDisabled })
        #expect(highlight.index == 0)
        highlight.hover(at: 2, count: rows.count, isDisabled: { rows[$0].isDisabled })
        #expect(highlight.index == 2)
        highlight.hover(at: 5, count: rows.count, isDisabled: { rows[$0].isDisabled })
        #expect(highlight.index == 2)
    }

    @Test func currentIsNilOutOfRangeAndClearGoesToNone() {
        var highlight = PVPopupListHighlight(index: 2)
        #expect(highlight.current(count: 2) == nil)
        #expect(highlight.current(count: 3) == 2)
        highlight.clear()
        #expect(highlight.index == -1)
    }

    // MARK: Movement

    @Test func movingFromNothingActiveLandsOnTheNearEnd() {
        #expect(moved(from: -1, .downArrow, in: pool) == 0)
        #expect(moved(from: -1, .upArrow, in: pool) == 2)
    }

    @Test func movementClampsAtBothEndsRatherThanCycling() {
        #expect(moved(from: 2, .downArrow, in: pool) == 2)
        #expect(moved(from: 0, .upArrow, in: pool) == 0)
        // A page jump past the end stops at the end, it does not wrap.
        #expect(moved(from: 0, .pageDown, in: pool) == 2)
    }

    @Test func edgeJumpsGoToFirstAndLast() {
        #expect(moved(from: 1, .end, in: pool) == 2)
        #expect(moved(from: 1, .home, in: pool) == 0)
    }

    @Test func movementStepsOverDisabledRows() {
        let rows = [option("a", "A"), option("b", "B", disabled: true), option("c", "C")]
        #expect(moved(from: 0, .downArrow, in: rows) == 2)
        #expect(moved(from: 2, .upArrow, in: rows) == 0)
    }

    @Test func movementStaysPutWhenOnlyDisabledRowsRemain() {
        let rows = [option("a", "A"), option("b", "B", disabled: true)]
        #expect(moved(from: 0, .downArrow, in: rows) == 0)
    }

    @Test func movementOnAnEmptyListIsNegative() {
        #expect(moved(from: -1, .downArrow, in: []) == -1)
    }

    @Test func commandJumpsAndOtherKeys() {
        #expect(moved(from: 0, .downArrow, command: true, in: pool) == 2)
        #expect(moved(from: 2, .upArrow, command: true, in: pool) == 0)
        #expect(moved(from: 2, .pageUp, in: pool) == 0)
        let twelve = (0..<12).map { option("o\($0)", "Option \($0)") }
        #expect(moved(from: 0, .pageDown, in: twelve) == 8)

        let leadingDisabled = [option("a", "A", disabled: true), option("b", "B"), option("c", "C")]
        #expect(moved(from: -1, .upArrow, command: true, in: leadingDisabled) == 1)
        #expect(moved(from: -1, .home, in: leadingDisabled) == 1)

        for key in [KeyEquivalent.return, .escape, .tab, .space, KeyEquivalent("a")] {
            #expect(!PVPopupListHighlight.isMovementKey(key))
            #expect(moved(from: 1, key, in: pool) == 1)
        }
        for key in [KeyEquivalent.upArrow, .downArrow, .home, .end, .pageUp, .pageDown] {
            #expect(PVPopupListHighlight.isMovementKey(key))
        }
    }
}

@Suite
struct PVPopupListLineKindTests {
    @Test func searchingLineShowsOnlyWhileLoadingWithNothingIn() {
        #expect(PVPopupListLineKind.of(rowCount: 0, isLoading: true) == .loading)
    }

    @Test func rowsAlreadyInStayWhileANewSearchRuns() {
        #expect(PVPopupListLineKind.of(rowCount: 3, isLoading: true) == .rows)
    }

    @Test func noRowsAndNoSearchShowsTheEmptyLine() {
        #expect(PVPopupListLineKind.of(rowCount: 0, isLoading: false) == .empty)
        #expect(PVPopupListLineKind.of(rowCount: 2, isLoading: false) == .rows)
    }
}

@Suite
struct PVPopupListPlacementTests {
    /// A roomy 1440x900 display with the menu bar taken off the top, in
    /// AppKit screen coordinates (y grows upward).
    private let screen = CGRect(x: 0, y: 0, width: 1440, height: 875)

    private func anchor(y: CGFloat, width: CGFloat = 300) -> CGRect {
        CGRect(x: 100, y: y, width: width, height: 28)
    }

    @Test func popupHangsBelowTheFieldWhenThereIsRoom() {
        let field = anchor(y: 600)
        let frame = PVPopupListPlacement.popupFrame(
            anchor: field, contentHeight: 220, visibleFrame: screen, gap: 4
        )
        #expect(frame.maxY == field.minY - 4)
        #expect(frame.height == 220)
        #expect(frame.minX == field.minX)
        #expect(frame.width == field.width)
    }

    @Test func popupFlipsAboveTheFieldWhenItWillNotFitBelow() {
        let field = anchor(y: 60)
        let frame = PVPopupListPlacement.popupFrame(
            anchor: field, contentHeight: 220, visibleFrame: screen, gap: 4
        )
        #expect(frame.minY == field.maxY + 4)
        #expect(frame.height == 220)
    }

    /// The whole point of the child window: it is placed strictly outside the
    /// anchor, so it can never sit on top of what is being typed.
    @Test func popupNeverOverlapsTheField() {
        for y in stride(from: CGFloat(0), through: 860, by: 20) {
            let field = anchor(y: y)
            let frame = PVPopupListPlacement.popupFrame(
                anchor: field, contentHeight: 220, visibleFrame: screen
            )
            #expect(!frame.intersects(field), "overlapped with the field at y=\(y)")
        }
    }

    @Test func popupStaysInsideTheVisibleFrame() {
        for y in stride(from: CGFloat(0), through: 860, by: 20) {
            let frame = PVPopupListPlacement.popupFrame(
                anchor: anchor(y: y), contentHeight: 220, visibleFrame: screen
            )
            #expect(screen.contains(frame) || frame.height == 0, "escaped the screen at y=\(y)")
        }
    }

    @Test func popupTrimsItsHeightToTheRoomAvailable() {
        // Cramped both ways: it takes the taller side and scrolls inside.
        let cramped = CGRect(x: 0, y: 0, width: 1440, height: 300)
        let field = anchor(y: 100)
        let frame = PVPopupListPlacement.popupFrame(
            anchor: field, contentHeight: 220, visibleFrame: cramped, gap: 4
        )
        #expect(frame.height < 220)
        #expect(cramped.contains(frame))
    }

    @Test func popupIsClampedToTheScreensHorizontalEdges() {
        let offRight = CGRect(x: 1300, y: 600, width: 300, height: 28)
        let frame = PVPopupListPlacement.popupFrame(
            anchor: offRight, contentHeight: 220, visibleFrame: screen
        )
        #expect(frame.maxX <= screen.maxX)
        #expect(frame.minX >= screen.minX)
    }

    @Test func anchorFrameAlignsTheWidthToAnEdge() {
        let field = CGRect(x: 600, y: 100, width: 300, height: 24)
        #expect(PVPopupListPlacement.anchorFrame(field, width: nil, alignment: .trailing) == field)
        #expect(
            PVPopupListPlacement.anchorFrame(field, width: 640, alignment: .leading)
                == CGRect(x: 600, y: 100, width: 640, height: 24))
        #expect(
            PVPopupListPlacement.anchorFrame(field, width: 640, alignment: .trailing)
                == CGRect(x: 260, y: 100, width: 640, height: 24))
    }

    @Test func trailingListStaysOnScreen() {
        let field = CGRect(x: 100, y: 500, width: 300, height: 24)
        let anchor = PVPopupListPlacement.anchorFrame(field, width: 640, alignment: .trailing)
        #expect(anchor.minX == -240)
        let frame = PVPopupListPlacement.popupFrame(
            anchor: anchor, contentHeight: 200, visibleFrame: CGRect(x: 0, y: 0, width: 1440, height: 900)
        )
        #expect(frame.minX == 0)
        #expect(frame.width == 640)
        #expect(frame.maxY == 496)
    }

    @Test(arguments: [
        (CGFloat(640), CGFloat(1070), HorizontalEdge.trailing, CGFloat(0), CGFloat(420), CGFloat(640)),
        (CGFloat(640), CGFloat(1070), HorizontalEdge.trailing, CGFloat(1102), CGFloat(420), CGFloat(640)),
        (CGFloat(640), CGFloat(404), HorizontalEdge.trailing, CGFloat(436), CGFloat(200), CGFloat(388)),
        (CGFloat(640), CGFloat(210), HorizontalEdge.trailing, CGFloat(436), CGFloat(200), CGFloat(200)),
        (CGFloat(600), CGFloat(60), HorizontalEdge.leading, CGFloat(436), CGFloat(0), CGFloat(360)),
        (CGFloat(600), CGFloat(12), HorizontalEdge.leading, CGFloat(1102), CGFloat(0), CGFloat(600)),
        (CGFloat(600), CGFloat(12), HorizontalEdge.leading, CGFloat(436), CGFloat(0), CGFloat(408)),
        (CGFloat(600), CGFloat(12), HorizontalEdge.leading, CGFloat(0), CGFloat(0), CGFloat(600)),
        (CGFloat(600), CGFloat(12), HorizontalEdge.leading, CGFloat(628), CGFloat(0), CGFloat(600)),
        (CGFloat(600), CGFloat(12), HorizontalEdge.leading, CGFloat(627), CGFloat(0), CGFloat(599)),
        (CGFloat(600), CGFloat(32), HorizontalEdge.leading, CGFloat(436), CGFloat(0), CGFloat(388)),
        (CGFloat(600), CGFloat(82), HorizontalEdge.leading, CGFloat(436), CGFloat(0), CGFloat(338)),
    ])
    func fittedWidthStaysInTheColumn(
        preferred: CGFloat, edge: CGFloat, alignment: HorizontalEdge,
        column: CGFloat, minimum: CGFloat, expected: CGFloat
    ) {
        #expect(
            PVPopupListPlacement.fittedWidth(
                preferred: preferred, hangsFrom: edge, alignment: alignment,
                columnWidth: column, minimum: minimum) == expected)
    }
}

@Suite
struct PVPopupListCardTests {
    @Test func cardHeightAddsThePaddingTwice() {
        #expect(PVPopupListCard<EmptyView>.height(rows: 100, verticalPadding: 8) == 116)
        #expect(
            PVPopupListCard<EmptyView>.height(rows: 100)
                == 100 + PVPopupListCard<EmptyView>.defaultVerticalPadding * 2
        )
    }

    @Test func aMeasurementIsUnmeasuredUntilRecordedAndResets() {
        var measurement = PVPopupListMeasurement()
        #expect(!measurement.isMeasured)
        measurement.record(40)
        #expect(measurement.isMeasured && measurement.height == 40)
        measurement.reset()
        #expect(measurement == PVPopupListMeasurement())
    }
}

@Suite
struct PVPopupListCommitKeyTests {
    @Test func returnAndKeypadEnterCommit() {
        #expect(PVPopupListHighlight.isCommitKey(.return))
        #expect(PVPopupListHighlight.isCommitKey(KeyEquivalent("\u{3}")))
        #expect(!PVPopupListHighlight.isCommitKey(.escape))
        #expect(!PVPopupListHighlight.isCommitKey(.downArrow))
    }
}
