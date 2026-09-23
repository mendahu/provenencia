import Foundation
import Testing
@testable import Provenencia

/// `PVTable`'s keyboard behaviour lives in two pure helpers so it can be
/// tested without mounting a view — see `PVTable.swift`.
@Suite
struct PVTableSelectionTests {
    @Test func movesDownAndUp() {
        #expect(PVTableSelection.moveIndex(0, delta: 1, count: 5) == 1)
        #expect(PVTableSelection.moveIndex(3, delta: -1, count: 5) == 2)
    }

    @Test func clampsAtBothEndsRatherThanWrapping() {
        #expect(PVTableSelection.moveIndex(4, delta: 1, count: 5) == 4)
        #expect(PVTableSelection.moveIndex(0, delta: -1, count: 5) == 0)
    }

    @Test func clampsAPageJumpToTheLastRow() {
        #expect(PVTableSelection.moveIndex(2, delta: 10, count: 5) == 4)
        #expect(PVTableSelection.moveIndex(8, delta: -10, count: 12) == 0)
    }

    @Test func startsFromTheEdgeTheMoveCameFromWhenNothingIsSelected() {
        #expect(PVTableSelection.moveIndex(-1, delta: 1, count: 5) == 0)
        #expect(PVTableSelection.moveIndex(-1, delta: -1, count: 5) == 4)
    }

    @Test func reportsNoSelectionForAnEmptyTable() {
        #expect(PVTableSelection.moveIndex(0, delta: 1, count: 0) == -1)
        #expect(PVTableSelection.moveIndex(-1, delta: -1, count: 0) == -1)
    }
}

@Suite
struct PVTableTypeSelectMatcherTests {
    private let labels = ["Author", "Certificate number", "Interview date", "Photographer", "Publication date"]

    @Test func matchesTheFirstRowWithThePrefix() {
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "pho") == 3)
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "cert") == 1)
    }

    @Test func matchesCaseInsensitively() {
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "AUTH") == 0)
    }

    @Test func matchesOnlyThePrefixNotAnySubstring() {
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "date") == -1)
    }

    @Test func wrapsForwardFromTheGivenIndexSoRepeatedKeysCycle() {
        // "P" from just past Photographer lands on Publication date, then
        // wraps back around to Photographer.
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "p", fromIndex: 4) == 4)
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "p", fromIndex: 5) == 3)
    }

    @Test func reportsNoMatchForAnEmptyPrefixOrNoRows() {
        #expect(PVTableTypeSelectMatcher.index(in: labels, prefix: "") == -1)
        #expect(PVTableTypeSelectMatcher.index(in: [], prefix: "a") == -1)
    }

    @Test func accumulatesKeystrokesWithinTheResetWindow() {
        var matcher = PVTableTypeSelectMatcher()
        let start = Date()
        #expect(matcher.append("p", now: start) == "p")
        #expect(matcher.append("u", now: start.addingTimeInterval(0.2)) == "pu")
        #expect(matcher.append("b", now: start.addingTimeInterval(0.4)) == "pub")
    }

    @Test func startsOverAfterThePause() {
        var matcher = PVTableTypeSelectMatcher()
        let start = Date()
        _ = matcher.append("p", now: start)
        let afterPause = start.addingTimeInterval(PVTableTypeSelectMatcher.resetInterval + 0.01)
        #expect(matcher.append("u", now: afterPause) == "u")
    }

    @Test func resetDropsTheBuffer() {
        var matcher = PVTableTypeSelectMatcher()
        let start = Date()
        _ = matcher.append("p", now: start)
        matcher.reset()
        #expect(matcher.append("u", now: start.addingTimeInterval(0.1)) == "u")
    }
}

@Suite
struct PVTableFilterSelectTests {
    @Test func selectOptionsMapLabelsAndCounts() {
        var received: String?
        let filter = PVTableColumnFilter(
            value: "text",
            active: true,
            options: [
                .init(value: "", label: LocalizedStringResource(stringLiteral: "All data types")),
                .init(value: "text", label: LocalizedStringResource(stringLiteral: "Text"), count: 3),
            ],
            onChange: { received = $0 }
        )
        let options = filter.selectOptions()
        #expect(options.map(\.id) == ["", "text"])
        #expect(options[0].label == "All data types")
        #expect(options[1].label == L10n.DesignSystem.tableFilterOptionCount(label: "Text", count: 3))
        filter.onChange("text")
        #expect(received == "text")
    }

    @Test func filterColumnLabelNamesTheAxis() {
        #expect(
            L10n.DesignSystem.tableFilterColumn(column: "Data type")
                == String(format: String(localized: LocalizedStringResource(
                    "designSystem.table.filterColumn",
                    defaultValue: "Filter %@",
                    comment: "Accessibility label for a PVTable column's filter menu; argument is the column title"
                )), locale: .current, "Data type")
        )
    }
}
