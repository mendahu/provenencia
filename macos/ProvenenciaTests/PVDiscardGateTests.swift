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
}
