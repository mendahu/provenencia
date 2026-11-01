import Foundation
import Testing
@testable import Provenencia

@Suite
@MainActor
struct PVDiscardGateTests {
    @Test func discardHandsTheStepOutOnce() {
        var ran = 0
        var gate = PVDiscardGate()
        gate.hold { ran += 1 }
        #expect(gate.holdsStep)
        let step = gate.discard()
        step?()
        #expect(ran == 1)
        let again = gate.discard()
        #expect(again == nil)
        #expect(!gate.holdsStep)
    }

    @Test func keepEditingDropsTheStep() {
        var gate = PVDiscardGate()
        gate.hold {}
        let held = gate.keepEditing()
        #expect(held)
        let after = gate.discard()
        #expect(after == nil)
        #expect(!gate.holdsStep)
    }

    @Test func keepEditingWithNothingHeld() {
        var gate = PVDiscardGate()
        let held = gate.keepEditing()
        #expect(!held)
    }

    @Test func sheetCloseOutcomes() {
        var ran = 0
        var gate = PVDiscardGate()
        let step: @MainActor () -> Void = { ran += 1 }

        #expect(PVSheetCloseRule.outcome(
            isRunning: true, isDirty: true, isConfirmingDiscard: false, next: step, gate: &gate
        ) == .nothing)
        #expect(!gate.holdsStep)

        #expect(PVSheetCloseRule.outcome(
            isRunning: false, isDirty: true, isConfirmingDiscard: true, next: step, gate: &gate
        ) == .keepEditing)
        #expect(!gate.holdsStep)

        #expect(PVSheetCloseRule.outcome(
            isRunning: false, isDirty: false, isConfirmingDiscard: false, next: step, gate: &gate
        ) == .closeThenRun)
        #expect(!gate.holdsStep)
        #expect(ran == 0)

        // Dirty: the step waits in the gate. Keep editing drops it.
        #expect(PVSheetCloseRule.outcome(
            isRunning: false, isDirty: true, isConfirmingDiscard: false, next: step, gate: &gate
        ) == .askToDiscard)
        #expect(gate.holdsStep)
        #expect(gate.keepEditing())
        #expect(ran == 0)

        // Dirty again: Discard hands the step out to run after the close.
        _ = PVSheetCloseRule.outcome(
            isRunning: false, isDirty: true, isConfirmingDiscard: false, next: step, gate: &gate
        )
        gate.discard()?()
        #expect(ran == 1)
    }
}
