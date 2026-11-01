/// The one step a "Discard changes?" question waits on. The owner keeps its own
/// question on screen (the sheet's `confirmingDiscard`, a card's pending switch,
/// the composer's leave dialog); the gate holds what to run once the answer is
/// Discard, and drops it when the answer is Keep editing.
///
/// A gate holds one step: `hold` replaces whatever was held.
struct PVDiscardGate {
    private var step: (@MainActor () -> Void)?

    /// Whether a step waits on the question.
    var holdsStep: Bool { step != nil }

    /// Holds `step` until the question is answered.
    mutating func hold(_ step: @escaping @MainActor () -> Void) {
        self.step = step
    }

    /// The answer was Discard: hands the held step out, once. Nil when none is held.
    mutating func discard() -> (@MainActor () -> Void)? {
        defer { step = nil }
        return step
    }

    /// The answer was Keep editing: drops the held step and says whether one was held.
    mutating func keepEditing() -> Bool {
        defer { step = nil }
        return step != nil
    }
}
