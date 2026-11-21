import AppKit
import SwiftUI
import Testing
@testable import Provenencia

@Suite
struct PVPopupDecisionTests {
    fileprivate struct Row: CustomTestStringConvertible, Sendable {
        let input: PVPopupInput
        let pending: Bool
        let keysStayInField: Bool
        let decision: PVPopupDecision
        var testDescription: String {
            "\(input) pending=\(pending) keysStayInField=\(keysStayInField)"
        }
    }

    /// The §6 Edges table, every `any` expanded to both values.
    fileprivate static let rows: [Row] = {
        var rows: [Row] = []
        let bools = [false, true]
        func add(_ input: PVPopupInput, pending: [Bool], stay: [Bool], _ decision: PVPopupDecision) {
            for p in pending {
                for s in stay {
                    rows.append(Row(input: input, pending: p, keysStayInField: s, decision: decision))
                }
            }
        }
        add(.key(.escape), pending: bools, stay: [false], .close(consumesEvent: true))
        for key in [KeyEquivalent.downArrow, KeyEquivalent("c")] {
            add(.key(key), pending: bools, stay: [false], .toPopup(key))
        }
        add(.key(.escape), pending: bools, stay: [true], .pass)
        add(.key(.downArrow), pending: bools, stay: [true], .pass)
        for button in [PVPopupMouseButton.left, .right, .other] {
            for inside in bools {
                add(.mouseUp(button, inside: inside), pending: [true], stay: bools, .endOpeningPress)
            }
        }
        add(.mouseUp(.left, inside: false), pending: [false], stay: bools, .close(consumesEvent: false))
        add(.mouseUp(.left, inside: true), pending: [false], stay: bools, .pass)
        for button in [PVPopupMouseButton.right, .other] {
            for inside in bools {
                add(.mouseUp(button, inside: inside), pending: [false], stay: bools, .pass)
            }
        }
        for inside in bools {
            add(.mouseDown(.left, inside: inside), pending: bools, stay: bools, .pass)
        }
        for button in [PVPopupMouseButton.right, .other] {
            add(.mouseDown(button, inside: false), pending: bools, stay: bools, .close(consumesEvent: false))
            add(.mouseDown(button, inside: true), pending: bools, stay: bools, .pass)
        }
        add(.windowResignedKey, pending: bools, stay: bools, .close(consumesEvent: false))
        return rows
    }()

    @Test(arguments: rows)
    fileprivate func decidesEveryRow(_ row: Row) {
        let decision = PVPopupDecision.decide(
            row.input,
            openingPressPending: row.pending,
            options: PVPopupOptions(keysStayInField: row.keysStayInField)
        )
        #expect(decision == row.decision)
    }

    fileprivate struct FieldRow: CustomTestStringConvertible, Sendable {
        let input: PVPopupInput
        let pending: Bool
        let onAnchor: Bool
        let staysOpen: Bool
        let keysStayInField: Bool
        let decision: PVPopupDecision
        var testDescription: String {
            "\(input) pending=\(pending) onAnchor=\(onAnchor) staysOpen=\(staysOpen)"
        }
    }

    fileprivate static let fieldRows: [FieldRow] = {
        func row(
            _ input: PVPopupInput, pending: Bool = false, onAnchor: Bool, staysOpen: Bool,
            keysStayInField: Bool = false, _ decision: PVPopupDecision
        ) -> FieldRow {
            FieldRow(
                input: input, pending: pending, onAnchor: onAnchor, staysOpen: staysOpen,
                keysStayInField: keysStayInField, decision: decision
            )
        }
        let close = PVPopupDecision.close(consumesEvent: false)
        return [
            row(.mouseUp(.left, inside: false), onAnchor: true, staysOpen: true, .pass),
            row(.mouseUp(.left, inside: false), onAnchor: true, staysOpen: false, close),
            row(.mouseUp(.left, inside: false), onAnchor: false, staysOpen: true, close),
            row(.mouseUp(.left, inside: false), pending: true, onAnchor: true, staysOpen: true, .endOpeningPress),
            row(.mouseDown(.right, inside: false), onAnchor: true, staysOpen: true, .pass),
            row(.mouseDown(.other, inside: false), onAnchor: true, staysOpen: true, .pass),
            row(.mouseDown(.right, inside: false), onAnchor: true, staysOpen: false, close),
            row(.mouseDown(.other, inside: false), onAnchor: true, staysOpen: false, close),
            row(.key(.escape), onAnchor: true, staysOpen: true, .close(consumesEvent: true)),
            row(.windowResignedKey, onAnchor: true, staysOpen: true, close),
        ]
    }()

    @Test(arguments: fieldRows)
    fileprivate func staysOpenOnFieldClickRows(_ row: FieldRow) {
        let decision = PVPopupDecision.decide(
            row.input,
            openingPressPending: row.pending,
            onAnchor: row.onAnchor,
            options: PVPopupOptions(keysStayInField: row.keysStayInField, staysOpenOnFieldClick: row.staysOpen)
        )
        #expect(decision == row.decision)
    }

    @Test func rightClicksPassWhenIgnored() {
        let options = PVPopupOptions(ignoresRightClicks: true)
        let close = PVPopupDecision.close(consumesEvent: false)
        func decide(_ input: PVPopupInput, pending: Bool = false) -> PVPopupDecision {
            PVPopupDecision.decide(input, openingPressPending: pending, options: options)
        }
        #expect(decide(.mouseDown(.right, inside: false)) == .pass)
        #expect(decide(.mouseDown(.right, inside: true)) == .pass)
        #expect(decide(.mouseUp(.right, inside: false)) == .pass)
        #expect(decide(.mouseUp(.right, inside: false), pending: true) == .endOpeningPress)
        #expect(decide(.mouseDown(.other, inside: false)) == close)
        #expect(decide(.mouseUp(.left, inside: false)) == close)
        #expect(decide(.key(.escape)) == .close(consumesEvent: true))
    }

    @Test func aClickIsWhereItsPressBegan() {
        let options = PVPopupOptions()
        let release = PVPopupInput.mouseUp(.left, inside: false)
        #expect(PVPopupDecision.decide(release, openingPressPending: false, pressBeganInside: true, options: options) == .pass)
        #expect(
            PVPopupDecision.decide(release, openingPressPending: false, pressBeganInside: false, options: options)
                == .close(consumesEvent: false)
        )
        #expect(PVPopupDecision.decide(release, openingPressPending: true, pressBeganInside: true, options: options) == .endOpeningPress)
        #expect(
            PVPopupDecision.decide(.mouseDown(.left, inside: false), openingPressPending: false, options: options) == .pass
        )
    }

    @Test func aPressWithNoReleaseClosesLikeARelease() {
        let options = PVPopupOptions()
        let ended = PVPopupInput.pressEndedUnseen
        #expect(
            PVPopupDecision.decide(ended, openingPressPending: false, pressBeganInside: false, options: options)
                == .close(consumesEvent: false)
        )
        #expect(PVPopupDecision.decide(ended, openingPressPending: false, pressBeganInside: true, options: options) == .pass)
    }
}

@Suite
struct PVPopupInputTests {
    @MainActor @Test func mapsKeys() {
        let none: NSEvent.ModifierFlags = []
        #expect(PVPopupInput.keyEquivalent(keyCode: 53, characters: nil, modifiers: none) == .escape)
        #expect(PVPopupInput.keyEquivalent(keyCode: 126, characters: nil, modifiers: none) == .upArrow)
        #expect(PVPopupInput.keyEquivalent(keyCode: 125, characters: nil, modifiers: none) == .downArrow)
        #expect(PVPopupInput.keyEquivalent(keyCode: 115, characters: nil, modifiers: none) == .home)
        #expect(PVPopupInput.keyEquivalent(keyCode: 119, characters: nil, modifiers: none) == .end)
        #expect(PVPopupInput.keyEquivalent(keyCode: 36, characters: "\r", modifiers: none) == .return)
        #expect(PVPopupInput.keyEquivalent(keyCode: 76, characters: "\r", modifiers: none) == .return)
        #expect(PVPopupInput.keyEquivalent(keyCode: 8, characters: "c", modifiers: none) == KeyEquivalent("c"))
        #expect(PVPopupInput.keyEquivalent(keyCode: 26, characters: "7", modifiers: none) == KeyEquivalent("7"))
        #expect(PVPopupInput.keyEquivalent(keyCode: 8, characters: "c", modifiers: .command) == nil)
        #expect(PVPopupInput.keyEquivalent(keyCode: 8, characters: "c", modifiers: .control) == nil)
        #expect(PVPopupInput.keyEquivalent(keyCode: 27, characters: "-", modifiers: none) == nil)
        #expect(PVPopupInput.keyEquivalent(keyCode: 121, characters: nil, modifiers: none) == nil)

        let escape = NSEvent.keyEvent(
            with: .keyDown, location: .zero, modifierFlags: [], timestamp: 0, windowNumber: 0,
            context: nil, characters: "\u{1b}", charactersIgnoringModifiers: "\u{1b}",
            isARepeat: false, keyCode: 53
        )
        #expect(escape.flatMap(PVPopupInput.init) == .key(.escape))
    }

    @MainActor @Test func mapsMouseEventsAndLevels() {
        func input(_ type: NSEvent.EventType) -> PVPopupInput? {
            NSEvent.mouseEvent(
                with: type, location: .zero, modifierFlags: [], timestamp: 0, windowNumber: 0,
                context: nil, eventNumber: 0, clickCount: 1, pressure: 1
            ).flatMap(PVPopupInput.init)
        }
        #expect(input(.leftMouseUp) == .mouseUp(.left, inside: false))
        #expect(input(.rightMouseUp) == .mouseUp(.right, inside: false))
        #expect(input(.rightMouseDown) == .mouseDown(.right, inside: false))
        #expect(input(.otherMouseDown) == .mouseDown(.other, inside: false))
        #expect(input(.leftMouseDown) == .mouseDown(.left, inside: false))

        #expect(!PVPopupInput.isPopupWindow(level: nil))
        #expect(!PVPopupInput.isPopupWindow(level: .normal))
        #expect(!PVPopupInput.isPopupWindow(level: .floating))
        #expect(PVPopupInput.isPopupWindow(level: .popUpMenu))
        #expect(PVPopupInput.isPopupWindow(level: .screenSaver))
    }
}

@Suite
struct PVPopupOpeningPressTests {
    private func mouse(_ type: NSEvent.EventType, number: Int) -> NSEvent {
        NSEvent.mouseEvent(
            with: type, location: .zero, modifierFlags: [], timestamp: 0, windowNumber: 0,
            context: nil, eventNumber: number, clickCount: 1, pressure: 1
        )!
    }

    @Test func queuedReleaseOfTheOpeningPressEndsIt() {
        var press = PVPopupOpeningPress(
            openingEvent: mouse(.leftMouseDown, number: 7), pressedMouseButtons: 0
        )
        let r1 = press.endsOpeningPress(mouse(.leftMouseUp, number: 7))
        #expect(r1)
        let r2 = press.endsOpeningPress(mouse(.leftMouseUp, number: 7))
        #expect(!r2)
    }

    @Test func anotherReleaseIsNotTheOpeningRelease() {
        var press = PVPopupOpeningPress(
            openingEvent: mouse(.leftMouseDown, number: 7), pressedMouseButtons: 1
        )
        let r3 = press.endsOpeningPress(mouse(.leftMouseUp, number: 8))
        #expect(!r3)
        let r4 = press.endsOpeningPress(mouse(.leftMouseUp, number: 7))
        #expect(r4)
        let r5 = press.endsOpeningPress(mouse(.leftMouseUp, number: 7))
        #expect(!r5)
    }

    @Test func noRecordedPressFallsBackToHeldButtons() {
        var held = PVPopupOpeningPress(openingEvent: nil, pressedMouseButtons: 1)
        let r6 = held.endsOpeningPress(mouse(.rightMouseUp, number: 3))
        #expect(r6)
        let r7 = held.endsOpeningPress(mouse(.leftMouseUp, number: 4))
        #expect(!r7)

        var none = PVPopupOpeningPress(openingEvent: nil, pressedMouseButtons: 0)
        let r8 = none.endsOpeningPress(mouse(.leftMouseUp, number: 4))
        #expect(!r8)
    }

    @Test func aMouseUpOpeningIsNotRecorded() {
        var held = PVPopupOpeningPress(
            openingEvent: mouse(.leftMouseUp, number: 7), pressedMouseButtons: 1
        )
        let r9 = held.endsOpeningPress(mouse(.leftMouseUp, number: 9))
        #expect(r9)

        var none = PVPopupOpeningPress(
            openingEvent: mouse(.leftMouseUp, number: 7), pressedMouseButtons: 0
        )
        let r10 = none.endsOpeningPress(mouse(.leftMouseUp, number: 7))
        #expect(!r10)
    }
}
