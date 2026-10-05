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
    static let built: Set<PromoteStep> = [.chooseTarget, .claim]

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
/// running, the leave guard is asking, a write was refused for good, or the
/// flow is over.
///
/// **Navigation.** Moving between steps is `.advance` and `.stepBack`, and
/// `controls` is everything the footer shows: where Back goes, whether Next
/// moves or writes, and what is enabled. A new step is a case on
/// `PromoteStep`, an entry in `built`, its screen, and the edits it owns
/// (`Edit.step`) — the footer and model don't change.
///
/// **Draft across steps.** Back and forward keep the whole draft. Nothing is
/// drafted from the target yet, so changing it keeps the claim fields; when
/// Compare drafts the argument (S9-19), a changed target must clear what was
/// drafted from it.
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

    /// The Identity Claim's status. Promote writes accepted claims only
    /// (S9-D10); Provisional and Rejected are later cases, a data change for
    /// the step's Select.
    enum ClaimStatus: String, CaseIterable, Sendable {
        case accepted
    }

    /// What the researcher has entered for the current subject. Grows with
    /// later steps (confirmed pairs in S9-19).
    struct Draft: Equatable, Sendable {
        var choice: Choice = .none
        var target: Target?
        var status: ClaimStatus = .accepted
        var confidenceGradeID: String?
        var argument = ""

        var isEmpty: Bool { self == Draft() }
    }

    /// A write the engine refused for good — retrying cannot succeed.
    struct Failure: Equatable, Sendable {
        let message: String
        /// The handle the subject was filed on elsewhere, once the graph shows it.
        var filedOn: String?
    }

    enum Phase: Equatable, Sendable {
        /// The current step takes input.
        case editing
        /// The subject's write is in flight; input is locked. A navigation
        /// asked for meanwhile is held and honoured once the write lands.
        case saving(heldLeave: PendingNavigation?)
        /// The leave guard is asking about this navigation.
        case confirmingLeave(PendingNavigation)
        /// The write was refused for good (the subject was filed elsewhere).
        /// Nothing was written, so leaving never asks; the step stays on
        /// screen with the reason.
        case blocked(Failure)
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

    /// An input on a step's screen. Each edit belongs to one step; `send`
    /// ignores it anywhere else.
    enum Edit: Equatable, Sendable {
        case choose(Choice)
        case selectTarget(Target)
        case setConfidence(String?)
        case setArgument(String)

        var step: PromoteStep {
            switch self {
            case .choose, .selectTarget: .chooseTarget
            case .setConfidence, .setArgument: .claim
            }
        }
    }

    enum Event: Equatable, Sendable {
        case edit(Edit)
        /// Next: the following built step, or the write after the last one.
        case advance
        /// Back within the flow (to the previous built step). Never asks:
        /// the draft is kept.
        case stepBack
        /// The write landed; `entityRef` is the handle the subject is now on.
        case saveSucceeded(entityRef: String)
        /// The write failed. A retryable failure keeps the step editable; one
        /// that isn't blocks it.
        case saveFailed(message: String, retryable: Bool)
        /// The subject left the graph or was promoted elsewhere (`filedOn`
        /// is that handle's ref, when known).
        case subjectUnavailable(filedOn: String?)
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
        /// Tell the researcher a subject was filed (a toast that outlives the place).
        case announceFiled(subjectName: String, entityRef: String)
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

    /// What the primary button does from the current step.
    enum Advance: Equatable, Sendable {
        /// Moves to this step.
        case step(PromoteStep)
        /// Writes the subject's claim.
        case save
    }

    /// Everything the footer needs, derived in one place.
    struct Controls: Equatable, Sendable {
        /// Where Back goes; nil hides it.
        var back: PromoteStep?
        var advance: Advance
        var canGoBack: Bool
        var canAdvance: Bool
        /// Step inputs are read-only (writing, refused, asking, or over).
        var isLocked: Bool
    }

    // MARK: State

    private(set) var queue: [Subject]
    private(set) var index = 0
    private(set) var draft = Draft()
    private(set) var step: PromoteStep = .chooseTarget
    private(set) var phase: Phase = .editing
    /// The last retryable write failure, shown on the step until the next edit.
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

    var failure: Failure? {
        if case .blocked(let failure) = phase { return failure }
        return nil
    }

    var isBlocked: Bool { failure != nil }

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

    var controls: Controls {
        let back = previousBuiltStep(before: step)
        return Controls(
            back: back,
            advance: nextBuiltStep(after: step).map(Advance.step) ?? .save,
            canGoBack: isEditing && back != nil,
            canAdvance: isEditing && isStepComplete,
            isLocked: !isEditing
        )
    }

    var canAdvance: Bool { controls.canAdvance }

    /// Leaving now would drop something the researcher entered.
    var hasUnsavedWork: Bool {
        switch phase {
        case .editing, .confirmingLeave: return !draft.isEmpty
        case .saving: return true
        case .blocked, .finished: return false
        }
    }

    // MARK: Transitions

    /// Applies one event and returns the effects to run, in order. Events that
    /// don't fit the current phase are ignored (no state change, no effects).
    mutating func send(_ event: Event) -> [Effect] {
        switch (phase, event) {
        case (.editing, .edit(let edit)):
            guard edit.step == step else { return [] }
            apply(edit)
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

        case (.saving(let held), .saveSucceeded(let entityRef)):
            savedCount += 1
            var effects: [Effect] = [
                .refreshAfterSave,
                .announceFiled(subjectName: subject.name, entityRef: entityRef),
            ]
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

        case (.saving(let held), .saveFailed(let message, let retryable)):
            if retryable {
                phase = .editing
                error = message
                return held != nil ? [.cancelNavigation] : []
            }
            // Someone else filed the subject: reload so the graph (and the
            // callout's handle) catch up. Nothing is unsaved, so a held
            // navigation goes ahead.
            phase = .blocked(Failure(message: message))
            error = nil
            return held != nil ? [.refreshAfterSave, .resumeNavigation] : [.refreshAfterSave]

        case (.blocked(var failure), .subjectUnavailable(let filedOn)):
            // Stay on the refusal; just learn which handle it was.
            if let filedOn, failure.filedOn != filedOn {
                failure.filedOn = filedOn
                phase = .blocked(failure)
            }
            return []

        case (_, .subjectUnavailable) where phase != .finished:
            let wasHolding: Bool
            switch phase {
            case .saving(let held): wasHolding = held != nil
            case .confirmingLeave: wasHolding = true
            default: wasHolding = false
            }
            phase = .finished
            return wasHolding ? [.cancelNavigation, .navigateToGraph] : [.navigateToGraph]

        case (.editing, .done), (.blocked, .done), (.finished, .done):
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
        case .finished, .blocked:
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

    private mutating func apply(_ edit: Edit) {
        switch edit {
        case .choose(let choice):
            guard draft.choice != choice else { return }
            draft.choice = choice
            if choice != .existing { draft.target = nil }
        case .selectTarget(let target):
            draft.choice = .existing
            draft.target = target
        case .setConfidence(let gradeID):
            draft.confidenceGradeID = gradeID
        case .setArgument(let argument):
            draft.argument = argument
        }
        error = nil
    }

    private func nextBuiltStep(after step: PromoteStep) -> PromoteStep? {
        guard let at = plan.firstIndex(of: step) else { return nil }
        return plan[(at + 1)...].first(where: builtSteps.contains)
    }

    private func previousBuiltStep(before step: PromoteStep) -> PromoteStep? {
        guard let at = plan.firstIndex(of: step) else { return nil }
        return plan[..<at].last(where: builtSteps.contains)
    }
}
