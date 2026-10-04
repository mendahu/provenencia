import Foundation

/// One step screen of the Promote flow (S9-D9 … D12).
enum PromoteStep: String, CaseIterable, Sendable {
    case chooseTarget
    /// Compare the subject with the chosen handle's members (S9-19).
    case compare
    /// Status, confidence and argument (S9-12).
    case claim

    /// Steps whose screens exist. The step row shows the whole plan as
    /// designed; advancing skips steps that aren't built yet, so a later PR
    /// adds its step here and the flow starts stopping on it.
    static let built: Set<PromoteStep> = [.chooseTarget]

    var isBuilt: Bool { Self.built.contains(self) }
}

/// The Promote flow as a state machine: the single source of truth for where
/// the researcher is, what they've chosen, and what is in flight.
///
/// It is a pure value. `send(_:)` applies one event and returns the effects
/// the model must run (write, navigate, recount); results come back as more
/// events. Nothing else changes flow state, so every transition is in one
/// switch and testable without a store or a view.
///
/// **Shape.** A walk is a queue of subjects (S9-11: one; S9-30 appends
/// neighbours as they are filed). Each subject runs its *plan* — the steps
/// its choice implies — with a draft that only becomes a write when the last
/// built step advances. `phase` says whether input is accepted, a write is
/// running, the leave guard is asking, or the flow is over.
struct PromoteFlow: Equatable, Sendable {
    // MARK: Types

    struct Subject: Equatable, Sendable {
        let id: String
        /// Candidate ref (CPR-…).
        let ref: String
        /// Working label, else the ref.
        let name: String
        let kind: EvidencePrimaryKind
    }

    enum Choice: Equatable, Sendable {
        case none, new, existing
    }

    /// An existing handle the subject would join.
    struct Target: Equatable, Sendable {
        let entityID: String
        let ref: String
        let title: String
        let memberCount: Int
    }

    /// What the researcher has entered for the current subject. Grows with
    /// later steps (claim fields in S9-12, confirmed pairs in S9-19).
    struct Draft: Equatable, Sendable {
        var choice: Choice = .none
        var target: Target?
        var confidenceGradeID: String?
        var argument = ""

        var isEmpty: Bool { self == Draft() }
    }

    enum Phase: Equatable, Sendable {
        /// The current step takes input.
        case editing
        /// The subject's write is in flight; input is locked. A navigation
        /// asked for meanwhile is held and honoured once the write lands.
        case saving(heldLeave: PendingNavigation?)
        /// The leave guard is asking about this navigation.
        case confirmingLeave(PendingNavigation)
        /// The flow is over; leaving never asks.
        case finished
    }

    /// The write for one subject.
    struct Save: Equatable, Sendable {
        let subjectID: String
        /// Nil mints a new handle.
        let entityID: String?
        let confidenceGradeID: String?
        let argument: String
    }

    enum Event: Equatable, Sendable {
        case choose(Choice)
        case selectTarget(Target)
        /// Next: the following built step, or the write after the last one.
        case advance
        /// Back within the flow (to the previous built step).
        case stepBack
        case saveSucceeded
        case saveFailed(message: String)
        /// The subject left the graph or was promoted elsewhere.
        case subjectUnavailable
        /// Done: end the walk; the navigation it starts goes through the guard
        /// (`requestLeave`).
        case done
        case leaveConfirmed
        case leaveCancelled
    }

    enum Effect: Equatable, Sendable {
        case save(Save)
        /// After a write: invalidate the graph and recount the sidebar.
        case refreshAfterSave
        case navigateToGraph
        case resumeNavigation
        case cancelNavigation
    }

    /// How the guard answers a navigation away from the place.
    enum LeaveAnswer: Equatable, Sendable {
        case allow
        /// Hold it; the flow has asked (or will act once its write lands).
        case hold
    }

    // MARK: State

    private(set) var queue: [Subject]
    private(set) var index = 0
    private(set) var draft = Draft()
    private(set) var step: PromoteStep = .chooseTarget
    private(set) var phase: Phase = .editing
    /// The last write's failure, shown on the step until the next edit.
    private(set) var error: String?
    /// Subjects filed in this walk.
    private(set) var savedCount = 0
    /// Which steps have screens (`PromoteStep.built`; tests pass others).
    let builtSteps: Set<PromoteStep>

    init(subject: Subject, builtSteps: Set<PromoteStep> = PromoteStep.built) {
        queue = [subject]
        self.builtSteps = builtSteps
    }

    // MARK: Derived

    var subject: Subject { queue[index] }

    /// The steps this subject's choice implies, as designed: joining adds
    /// Compare. Until Existing is chosen the plan is the mint path.
    var plan: [PromoteStep] {
        draft.choice == .existing ? [.chooseTarget, .compare, .claim] : [.chooseTarget, .claim]
    }

    /// 1-based position of the current step in the plan.
    var stepNumber: Int { (plan.firstIndex(of: step) ?? 0) + 1 }

    var isEditing: Bool { phase == .editing }

    var isSaving: Bool {
        if case .saving = phase { return true }
        return false
    }

    var pendingLeave: PendingNavigation? {
        if case .confirmingLeave(let pending) = phase { return pending }
        return nil
    }

    /// The current step has what it needs to move on.
    var isStepComplete: Bool {
        switch step {
        case .chooseTarget:
            switch draft.choice {
            case .new: return true
            case .existing: return draft.target != nil
            case .none: return false
            }
        case .compare, .claim:
            return true
        }
    }

    var canAdvance: Bool { isEditing && isStepComplete }

    /// Leaving now would drop something the researcher entered.
    var hasUnsavedWork: Bool {
        switch phase {
        case .editing, .confirmingLeave: return !draft.isEmpty
        case .saving: return true
        case .finished: return false
        }
    }

    // MARK: Transitions

    /// Applies one event and returns the effects to run, in order. Events that
    /// don't fit the current phase are ignored (no state change, no effects).
    mutating func send(_ event: Event) -> [Effect] {
        switch (phase, event) {
        case (.editing, .choose(let choice)):
            guard step == .chooseTarget, draft.choice != choice else { return [] }
            draft.choice = choice
            if choice != .existing { draft.target = nil }
            error = nil
            return []

        case (.editing, .selectTarget(let target)):
            guard step == .chooseTarget else { return [] }
            draft.choice = .existing
            draft.target = target
            error = nil
            return []

        case (.editing, .advance):
            guard isStepComplete else { return [] }
            if let next = nextBuiltStep(after: step) {
                step = next
                return []
            }
            phase = .saving(heldLeave: nil)
            error = nil
            return [.save(Save(
                subjectID: subject.id,
                entityID: draft.choice == .existing ? draft.target?.entityID : nil,
                confidenceGradeID: draft.confidenceGradeID,
                argument: draft.argument
            ))]

        case (.editing, .stepBack):
            if let previous = previousBuiltStep(before: step) {
                step = previous
            }
            return []

        case (.saving(let held), .saveSucceeded):
            savedCount += 1
            var effects: [Effect] = [.refreshAfterSave]
            if index + 1 < queue.count {
                index += 1
                draft = Draft()
                step = .chooseTarget
                phase = .editing
                if held != nil { effects.append(.resumeNavigation) }
                return effects
            }
            phase = .finished
            effects.append(held != nil ? .resumeNavigation : .navigateToGraph)
            return effects

        case (.saving(let held), .saveFailed(let message)):
            phase = .editing
            error = message
            return held != nil ? [.cancelNavigation] : []

        case (_, .subjectUnavailable) where phase != .finished:
            let wasHolding: Bool
            switch phase {
            case .saving(let held): wasHolding = held != nil
            case .confirmingLeave: wasHolding = true
            default: wasHolding = false
            }
            phase = .finished
            return wasHolding ? [.cancelNavigation, .navigateToGraph] : [.navigateToGraph]

        case (.editing, .done), (.finished, .done):
            return [.navigateToGraph]

        case (.confirmingLeave, .leaveConfirmed):
            phase = .finished
            return [.resumeNavigation]

        case (.confirmingLeave, .leaveCancelled):
            phase = .editing
            return [.cancelNavigation]

        default:
            return []
        }
    }

    /// The guard's answer to a navigation away, applying it to the flow.
    mutating func requestLeave(_ pending: PendingNavigation) -> LeaveAnswer {
        switch phase {
        case .finished:
            return .allow
        case .editing:
            guard !draft.isEmpty else { return .allow }
            phase = .confirmingLeave(pending)
            return .hold
        case .saving:
            // Let the write land first; the navigation follows it.
            phase = .saving(heldLeave: pending)
            return .hold
        case .confirmingLeave:
            // Already asking about another navigation: replace it.
            phase = .confirmingLeave(pending)
            return .hold
        }
    }

    // MARK: Walk

    /// Adds subjects to walk after the current one (S9-29 / S9-30).
    mutating func enqueue(_ subjects: [Subject]) {
        let known = Set(queue.map(\.id))
        queue.append(contentsOf: subjects.filter { !known.contains($0.id) })
    }

    // MARK: Helpers

    private func nextBuiltStep(after step: PromoteStep) -> PromoteStep? {
        guard let at = plan.firstIndex(of: step) else { return nil }
        return plan[(at + 1)...].first(where: builtSteps.contains)
    }

    private func previousBuiltStep(before step: PromoteStep) -> PromoteStep? {
        guard let at = plan.firstIndex(of: step) else { return nil }
        return plan[..<at].last(where: builtSteps.contains)
    }
}
