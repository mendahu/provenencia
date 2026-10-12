import Foundation
import SwiftUI
import Testing
@testable import Provenencia

@Suite
struct PVComboBoxTests {
    private func option(_ value: String, _ label: String, subtext: String = "", disabled: Bool = false) -> PVComboBoxOption {
        PVComboBoxOption(value: value, label: label, subtext: subtext, isDisabled: disabled)
    }

    private var pool: [PVComboBoxOption] {
        [
            option("f1", "Author", subtext: "author"),
            option("f2", "Publication date", subtext: "publication-date"),
            option("f3", "Publisher", subtext: "publisher"),
        ]
    }

    // MARK: Filtering

    @Test func emptyQueryKeepsEveryRow() {
        #expect(PVComboBoxMatch.filter(pool, query: "   ").map(\.value) == ["f1", "f2", "f3"])
    }

    @Test func filterMatchesLabelOrSubtextAsSubstring() {
        // "date" is only reachable mid-token, so this must not be a prefix match.
        #expect(PVComboBoxMatch.filter(pool, query: "date").map(\.value) == ["f2"])
        #expect(PVComboBoxMatch.filter(pool, query: "publi").map(\.value) == ["f2", "f3"])
        // The key alone finds the row, even though the label reads differently.
        #expect(PVComboBoxMatch.filter(pool, query: "publication-").map(\.value) == ["f2"])
    }

    @Test func filterIgnoresCaseAndDiacritics() {
        let accented = [option("p1", "Ó Riain"), option("p2", "Smith")]
        #expect(PVComboBoxMatch.filter(accented, query: "o ri").map(\.value) == ["p1"])
        #expect(PVComboBoxMatch.filter(accented, query: "SMITH").map(\.value) == ["p2"])
    }

    @Test func filterWithNoHitsReturnsNothing() {
        #expect(PVComboBoxMatch.filter(pool, query: "nomatch").isEmpty)
    }

    @Test func prefixOutranksALaterWordWhichOutranksAMidWordMatch() {
        let rows = [
            option("g", "grandparent"),
            option("p", "parent"),
            option("s", "step-parent"),
            option("k", "kinship", subtext: "parental"),
        ]
        #expect(PVComboBoxMatch.filter(rows, query: "par").map(\.value) == ["p", "s", "g", "k"])
    }

    @Test func orderInsideABandStaysTheCallerOrder() {
        let rows = [
            option("a", "parentage"),
            option("b", "parent"),
            option("c", "step parent"),
            option("d", "foster-parent"),
        ]
        #expect(PVComboBoxMatch.filter(rows, query: "par").map(\.value) == ["a", "b", "c", "d"])
    }

    // MARK: Spotlight rule

    @Test func firstEnabledIndexSkipsDisabledLeadingRows() {
        let rows = [option("a", "A", disabled: true), option("b", "B")]
        #expect(PVComboBoxMatch.firstEnabledIndex(in: rows) == 1)
    }

    @Test func firstEnabledIndexIsNegativeWhenEveryRowIsDisabled() {
        let rows = [option("a", "A", disabled: true), option("b", "B", disabled: true)]
        #expect(PVComboBoxMatch.firstEnabledIndex(in: rows) == -1)
        #expect(PVComboBoxMatch.firstEnabledIndex(in: []) == -1)
    }

    // MARK: Movement

    @Test func movingFromNothingActiveLandsOnTheNearEnd() {
        #expect(PVComboBoxMatch.move(from: -1, delta: 1, in: pool) == 0)
        #expect(PVComboBoxMatch.move(from: -1, delta: -1, in: pool) == 2)
    }

    @Test func movementClampsAtBothEndsRatherThanCycling() {
        #expect(PVComboBoxMatch.move(from: 2, delta: 1, in: pool) == 2)
        #expect(PVComboBoxMatch.move(from: 0, delta: -1, in: pool) == 0)
        // A page jump past the end stops at the end, it does not wrap.
        #expect(PVComboBoxMatch.move(from: 0, delta: 8, in: pool) == 2)
    }

    @Test func edgeJumpsGoToFirstAndLast() {
        #expect(PVComboBoxMatch.move(from: 1, delta: 1, in: pool, to: .last) == 2)
        #expect(PVComboBoxMatch.move(from: 1, delta: -1, in: pool, to: .first) == 0)
    }

    @Test func movementStepsOverDisabledRows() {
        let rows = [option("a", "A"), option("b", "B", disabled: true), option("c", "C")]
        #expect(PVComboBoxMatch.move(from: 0, delta: 1, in: rows) == 2)
        #expect(PVComboBoxMatch.move(from: 2, delta: -1, in: rows) == 0)
    }

    @Test func movementStaysPutWhenOnlyDisabledRowsRemain() {
        let rows = [option("a", "A"), option("b", "B", disabled: true)]
        #expect(PVComboBoxMatch.move(from: 0, delta: 1, in: rows) == 0)
    }

    @Test func movementOnAnEmptyListIsNegative() {
        #expect(PVComboBoxMatch.move(from: -1, delta: 1, in: []) == -1)
    }

    // MARK: Placement

    /// A roomy 1440x900 display with the menu bar taken off the top, in
    /// AppKit screen coordinates (y grows upward).
    private let screen = CGRect(x: 0, y: 0, width: 1440, height: 875)

    private func anchor(y: CGFloat, width: CGFloat = 300) -> CGRect {
        CGRect(x: 100, y: y, width: width, height: 28)
    }

    @Test func popupHangsBelowTheFieldWhenThereIsRoom() {
        let field = anchor(y: 600)
        let frame = PVComboBoxPlacement.popupFrame(
            anchor: field, contentHeight: 220, visibleFrame: screen, gap: 4
        )
        #expect(frame.maxY == field.minY - 4)
        #expect(frame.height == 220)
        #expect(frame.minX == field.minX)
        #expect(frame.width == field.width)
    }

    @Test func popupFlipsAboveTheFieldWhenItWillNotFitBelow() {
        let field = anchor(y: 60)
        let frame = PVComboBoxPlacement.popupFrame(
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
            let frame = PVComboBoxPlacement.popupFrame(
                anchor: field, contentHeight: 220, visibleFrame: screen
            )
            #expect(!frame.intersects(field), "overlapped with the field at y=\(y)")
        }
    }

    @Test func popupStaysInsideTheVisibleFrame() {
        for y in stride(from: CGFloat(0), through: 860, by: 20) {
            let frame = PVComboBoxPlacement.popupFrame(
                anchor: anchor(y: y), contentHeight: 220, visibleFrame: screen
            )
            #expect(screen.contains(frame) || frame.height == 0, "escaped the screen at y=\(y)")
        }
    }

    @Test func popupTrimsItsHeightToTheRoomAvailable() {
        // Cramped both ways: it takes the taller side and scrolls inside.
        let cramped = CGRect(x: 0, y: 0, width: 1440, height: 300)
        let field = anchor(y: 100)
        let frame = PVComboBoxPlacement.popupFrame(
            anchor: field, contentHeight: 220, visibleFrame: cramped, gap: 4
        )
        #expect(frame.height < 220)
        #expect(cramped.contains(frame))
    }

    @Test func popupIsClampedToTheScreensHorizontalEdges() {
        let offRight = CGRect(x: 1300, y: 600, width: 300, height: 28)
        let frame = PVComboBoxPlacement.popupFrame(
            anchor: offRight, contentHeight: 220, visibleFrame: screen
        )
        #expect(frame.maxX <= screen.maxX)
        #expect(frame.minX >= screen.minX)
    }

    // MARK: Highlight

    @Test func highlightMarksTheMatchedSpanOnly() {
        let marked = PVComboBoxHighlight.attributed("Publication date", query: "date")
        let runs = marked.runs.map { run in
            (
                String(marked[run.range].characters),
                run[AttributeScopes.SwiftUIAttributes.BackgroundColorAttribute.self]
            )
        }
        #expect(runs.map(\.0) == ["Publication ", "date"])
        #expect(runs[0].1 == nil)
        #expect(runs[1].1 == PVColor.markBackground)
    }

    @Test func highlightMatchesCaseInsensitively() {
        let marked = PVComboBoxHighlight.attributed("Author", query: "AUTH")
        let runs = marked.runs.map { String(marked[$0.range].characters) }
        #expect(runs == ["Auth", "or"])
    }

    @Test func highlightLeavesTextAloneWithoutAQuery() {
        let marked = PVComboBoxHighlight.attributed("Author", query: "  ")
        #expect(marked.runs.count == 1)
        #expect(String(marked.characters) == "Author")
    }

    @Test func highlightLeavesTextAloneWhenTheQueryIsAbsent() {
        let marked = PVComboBoxHighlight.attributed("Author", query: "zz")
        #expect(marked.runs.count == 1)
    }
}
