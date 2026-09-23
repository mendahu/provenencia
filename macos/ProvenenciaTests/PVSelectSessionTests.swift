import Foundation
import Testing
@testable import Provenencia

@Suite
struct PVSelectSessionTests {
    private func partTypes() -> [PVSelectOption] {
        [
            PVSelectOption(value: "", label: "(none)"),
            PVSelectOption(value: "given", label: "Given name"),
            PVSelectOption(value: "surname", label: "Surname"),
            PVSelectOption(value: "suffix", label: "Suffix"),
        ]
    }

    private func session(selection: String = "given", disabled: Bool = false) -> PVSelectSession {
        PVSelectSession(options: partTypes(), selection: selection, isDisabled: disabled)
    }

    // MARK: Closed field

    @Test func emptyOrDisabledCannotOpen() {
        var empty = PVSelectSession(options: [], selection: "")
        #expect(!empty.canOpen)
        empty.handleKey(.space)
        #expect(!empty.isOpen)
        empty.handleKey(.down)
        #expect(empty.selection == "")

        var disabled = session(disabled: true)
        #expect(!disabled.canOpen)
        disabled.handleKey(.space)
        #expect(!disabled.isOpen)
        disabled.handleKey(.down)
        #expect(disabled.selection == "given")
    }

    @Test func closedArrowsCommitAndStayClosed() {
        var nav = session()
        nav.handleKey(.down)
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
        nav.handleKey(.up)
        #expect(nav.selection == "given")
        #expect(!nav.isOpen)
    }

    @Test func closedArrowsClampAndDoNotWrap() {
        var nav = session(selection: "")
        nav.handleKey(.up)
        #expect(nav.selection == "")
        var last = session(selection: "suffix")
        last.handleKey(.down)
        #expect(last.selection == "suffix")
    }

    @Test func closedArrowsFromNoSelectionLandOnTheNearEnd() {
        var nav = PVSelectSession(options: partTypes(), selection: "missing")
        nav.handleKey(.down)
        #expect(nav.selection == "")
        var up = PVSelectSession(options: partTypes(), selection: "missing")
        up.handleKey(.up)
        #expect(up.selection == "suffix")
    }

    @Test func closedHomeAndEndCommitFirstAndLast() {
        var nav = session()
        nav.handleKey(.home)
        #expect(nav.selection == "")
        #expect(!nav.isOpen)
        nav.handleKey(.end)
        #expect(nav.selection == "suffix")
        #expect(!nav.isOpen)
    }

    @Test func closedTypeSelectCommitsWithoutOpening() {
        var nav = session()
        nav.handleKey(.character("s"))
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
    }

    @Test func closedTypeSelectCyclesAfterReset() {
        var nav = session(selection: "")
        let start = Date()
        nav.handleKey(.character("s"), now: start)
        #expect(nav.selection == "surname")
        let next = start.addingTimeInterval(PVTypeSelectMatcher.resetInterval + 0.01)
        nav.handleKey(.character("s"), now: next)
        #expect(nav.selection == "suffix")
    }

    @Test func closedTypeSelectAccumulatesAPrefix() {
        let options = [
            PVSelectOption(value: "given", label: "Given name"),
            PVSelectOption(value: "suffix", label: "Suffix"),
            PVSelectOption(value: "surname", label: "Surname"),
        ]
        var nav = PVSelectSession(options: options, selection: "given")
        let start = Date()
        nav.handleKey(.character("s"), now: start)
        #expect(nav.selection == "suffix")
        nav.handleKey(.character("u"), now: start.addingTimeInterval(0.1))
        #expect(nav.selection == "suffix")
        #expect(!nav.isOpen)
    }

    @Test func closedEscapeIsANoOp() {
        var nav = session()
        nav.handleKey(.escape)
        #expect(nav.selection == "given")
        #expect(!nav.isOpen)
    }

    @Test func triggerLabelPrefersDisplayLabelThenCommittedThenPlaceholder() {
        let nav = session()
        #expect(
            PVSelectAccessibility.triggerLabel(session: nav, displayLabel: "Sorted by date", placeholder: nil)
                == "Sorted by date"
        )
        #expect(
            PVSelectAccessibility.triggerLabel(session: nav, displayLabel: nil, placeholder: "Choose")
                == "Given name"
        )
        let empty = PVSelectSession(options: [], selection: "")
        #expect(
            PVSelectAccessibility.triggerLabel(session: empty, displayLabel: nil, placeholder: "No projects")
                == "No projects"
        )
    }

    // MARK: Open menu

    @Test func openSnapshotsAndHighlightsTheCommittedRow() {
        var nav = session()
        nav.handleKey(.space)
        #expect(nav.isOpen)
        #expect(nav.selectionOnOpen == "given")
        #expect(nav.highlightIndex == 1)
        #expect(nav.selection == "given")
    }

    @Test func openArrowsMoveHighlightWithoutCommitting() {
        var nav = session()
        nav.handleKey(.space)
        nav.handleKey(.down)
        #expect(nav.highlightIndex == 2)
        #expect(nav.selection == "given")
        nav.handleKey(.up)
        #expect(nav.highlightIndex == 1)
        #expect(nav.selection == "given")
    }

    @Test func openHomeAndEndJumpHighlightOnly() {
        var nav = session()
        nav.handleKey(.space)
        nav.handleKey(.end)
        #expect(nav.highlightIndex == 3)
        #expect(nav.selection == "given")
        nav.handleKey(.home)
        #expect(nav.highlightIndex == 0)
        #expect(nav.selection == "given")
    }

    @Test func openTypeSelectJumpsHighlightAndEscapeRestores() {
        var nav = session()
        nav.handleKey(.space)
        nav.handleKey(.character("s"))
        #expect(nav.highlightIndex == 2)
        #expect(nav.selection == "given")
        nav.handleKey(.escape)
        #expect(!nav.isOpen)
        #expect(nav.selection == "given")
    }

    @Test func returnCommitsTheHighlight() {
        var nav = session()
        nav.handleKey(.space)
        nav.handleKey(.down)
        nav.handleKey(.return)
        #expect(!nav.isOpen)
        #expect(nav.selection == "surname")
    }

    @Test func clickAwayRestoresTheSnapshot() {
        var nav = session()
        nav.handleKey(.space)
        nav.handleKey(.down)
        nav.handlePointer(.clickAway)
        #expect(!nav.isOpen)
        #expect(nav.selection == "given")
    }

    @Test func incrementAndDecrementMatchClosedArrows() {
        var nav = session()
        nav.increment()
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
        nav.decrement()
        #expect(nav.selection == "given")
    }

    @Test func incrementWhileOpenMovesHighlightOnly() {
        var nav = session()
        nav.handleKey(.space)
        nav.increment()
        #expect(nav.selection == "given")
        #expect(nav.highlightIndex == 2)
        #expect(nav.isOpen)
        nav.decrement()
        #expect(nav.selection == "given")
        #expect(nav.highlightIndex == 1)
    }

    @Test func emptyOptionsShowPlaceholderAndCannotOpen() {
        var nav = PVSelectSession(options: [], selection: "")
        #expect(
            PVSelectAccessibility.triggerLabel(session: nav, displayLabel: nil, placeholder: "Choose")
                == "Choose"
        )
        nav.handlePointer(.press(.trigger))
        nav.increment()
        #expect(!nav.isOpen)
        #expect(nav.selection == "")
    }

    @Test func iconOnlyDoesNotChangeTheSessionContract() {
        #expect(PVSelectAccessibility.hasRequiredLabel(iconOnly: true, hasLabel: true))
        var nav = session()
        nav.handleKey(.down)
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
    }

    // MARK: Chrome / a11y

    @Test func committedRowKeepsCheckmarkWithoutFill() {
        let chrome = PVSelectRowChrome.appearance(committed: true, highlighted: false)
        #expect(chrome.showsCheckmark)
        #expect(!chrome.fillsHighlight)
        #expect(chrome.isSelectedTrait)
    }

    @Test func highlightFillIsNotASelectedTrait() {
        let chrome = PVSelectRowChrome.appearance(committed: false, highlighted: true)
        #expect(!chrome.showsCheckmark)
        #expect(chrome.fillsHighlight)
        #expect(!chrome.isSelectedTrait)
    }

    @Test func committedAndHighlightedKeepsOneFillPlusCheck() {
        let chrome = PVSelectRowChrome.appearance(committed: true, highlighted: true)
        #expect(chrome.showsCheckmark)
        #expect(chrome.fillsHighlight)
        #expect(chrome.isSelectedTrait)
    }

    @Test func identifiersMintDottedMenuSuffix() {
        #expect(PVSelectAccessibility.identifier("nameValue.part.0.type") == "nameValue.part.0.type")
        #expect(
            PVSelectAccessibility.menuIdentifier("nameValue.part.0.type")
                == "nameValue.part.0.type.menu"
        )
        #expect(PVSelectAccessibility.identifier(nil) == nil)
    }

    @Test func spokenValueIsCommittedWhenClosedAndHighlightWhenOpen() {
        var nav = session()
        let closed = PVSelectAccessibility.spoken(session: nav)
        #expect(closed.value == "Given name")
        #expect(!closed.isExpanded)
        #expect(closed.position == nil)

        nav.handleKey(.space)
        nav.handleKey(.down)
        let open = PVSelectAccessibility.spoken(session: nav)
        #expect(open.value == "Surname")
        #expect(open.isExpanded)
        #expect(open.position?.current == 3)
        #expect(open.position?.count == 4)
    }

    @Test func iconOnlyRequiresALabel() {
        #expect(PVSelectAccessibility.hasRequiredLabel(iconOnly: true, hasLabel: true))
        #expect(!PVSelectAccessibility.hasRequiredLabel(iconOnly: true, hasLabel: false))
        #expect(PVSelectAccessibility.hasRequiredLabel(iconOnly: false, hasLabel: false))
    }

    // MARK: Board try-it integration

    @Test func closedDownThenUpReturnsToTheStart() {
        var nav = session()
        nav.handleKey(.down)
        nav.handleKey(.up)
        #expect(nav.selection == "given")
        #expect(!nav.isOpen)
    }

    @Test func typeSWhileClosedLandsSurname() {
        var nav = session()
        nav.handleKey(.character("s"))
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
    }

    @Test func spaceThenTypeGJumpsHighlightWithoutCommit() {
        var nav = session()
        nav.handleKey(.character("s"))
        nav.handleKey(.space)
        #expect(nav.highlightIndex == 2)
        nav.handleKey(.character("g"))
        #expect(nav.highlightIndex == 1)
        #expect(nav.selection == "surname")
    }

    @Test func escapeAfterOpenTypeSelectRestoresSurname() {
        var nav = session()
        nav.handleKey(.character("s"))
        nav.handleKey(.space)
        nav.handleKey(.character("g"))
        nav.handleKey(.escape)
        #expect(!nav.isOpen)
        #expect(nav.selection == "surname")
    }

    @Test func spaceDownReturnCommitsSuffix() {
        var nav = session()
        nav.handleKey(.character("s"))
        nav.handleKey(.space)
        nav.handleKey(.down)
        nav.handleKey(.return)
        #expect(nav.selection == "suffix")
        #expect(!nav.isOpen)
    }

    @Test func pressDragReleaseOntoGivenNameCommits() {
        var nav = session()
        nav.handleKey(.character("s"))
        nav.handlePointer(.press(.trigger))
        #expect(nav.isOpen)
        nav.handlePointer(.drag(.row(1)))
        #expect(nav.selection == "surname")
        nav.handlePointer(.release(.row(1)))
        #expect(nav.selection == "given")
        #expect(!nav.isOpen)
    }

    @Test func pressReleaseOnTriggerThenClickAwayRestores() {
        var nav = session()
        nav.handleKey(.character("s"))
        nav.handleKey(.space)
        nav.handleKey(.down)
        nav.handleKey(.return)
        #expect(nav.selection == "suffix")
        nav.handlePointer(.press(.trigger))
        nav.handlePointer(.release(.trigger))
        #expect(nav.isOpen)
        #expect(nav.selection == "suffix")
        nav.handlePointer(.clickAway)
        #expect(!nav.isOpen)
        #expect(nav.selection == "suffix")
    }

    @Test func disabledSessionIgnoresTheBoardScript() {
        var nav = session(disabled: true)
        nav.handleKey(.down)
        nav.handleKey(.character("s"))
        nav.handleKey(.space)
        nav.handlePointer(.press(.trigger))
        #expect(nav.selection == "given")
        #expect(!nav.isOpen)
    }

    @Test func nameValuePartTypesTypeSelectReachesSurname() {
        var nav = PVSelectSession(options: NamePartType.selectOptions, selection: "")
        let start = Date()
        nav.handleKey(.character("s"), now: start)
        #expect(nav.selection == "surname_prefix")
        let later = start.addingTimeInterval(PVTypeSelectMatcher.resetInterval + 0.01)
        nav.handleKey(.character("s"), now: later)
        #expect(nav.selection == "surname")
        #expect(!nav.isOpen)
    }
}
