import Testing
@testable import Provenencia

/// The Promote state machine on its own: every transition, no store or view.
@Suite
struct PromoteFlowTests {
    private let james = PromoteFlow.Subject(id: "sub-1", ref: "CPR-1", name: "James Robins", kind: .person)
    private let mary = PromoteFlow.Subject(id: "sub-2", ref: "CPR-2", name: "Mary Robins", kind: .person)
    private let per1 = PromoteFlow.Target(entityID: "e1", ref: "PER-1", title: "James Robins", memberCount: 2)
    private let allSteps = Set(PromoteStep.allCases)

    /// A flow on the claim step of the mint path, ready to save.
    private func mintOnClaim(_ subject: PromoteFlow.Subject? = nil) -> PromoteFlow {
        var flow = PromoteFlow(subject: subject ?? james)
        _ = flow.send(.edit(.choose(.new)))
        _ = flow.send(.advance)
        return flow
    }

    /// A flow whose write is in flight.
    private func saving(_ subject: PromoteFlow.Subject? = nil) -> PromoteFlow {
        var flow = mintOnClaim(subject)
        _ = flow.send(.advance)
        return flow
    }

    // MARK: Choosing

    @Test func startsOnChooseTargetWithNothingChosen() {
        let flow = PromoteFlow(subject: james)
        #expect(flow.phase == .editing)
        #expect(flow.step == .chooseTarget)
        #expect(flow.plan == [.chooseTarget, .claim])
        #expect(flow.stepNumber == 1)
        #expect(!flow.canAdvance)
        #expect(!flow.hasUnsavedWork)
    }

    @Test func theChoiceShapesThePlan() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.edit(.choose(.existing))).isEmpty)
        #expect(flow.plan == [.chooseTarget, .compare, .claim])
        #expect(!flow.canAdvance)
        _ = flow.send(.edit(.selectTarget(per1)))
        #expect(flow.canAdvance)
        _ = flow.send(.edit(.choose(.new)))
        #expect(flow.draft.target == nil)
        #expect(flow.plan == [.chooseTarget, .claim])
        #expect(flow.canAdvance)
    }

    @Test func selectingATargetChoosesExisting() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        #expect(flow.draft.choice == .existing)
        #expect(flow.draft.target == per1)
    }

    // MARK: Steps

    @Test func nextOnTheTargetGoesToClaimFields() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.choose(.new)))
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.step == .claim)
        #expect(flow.stepNumber == 2)
        #expect(flow.plan.count == 2)
    }

    @Test func aJoinSkipsCompareUntilItIsBuilt() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        #expect(flow.step == .claim)
        #expect(flow.plan == [.chooseTarget, .compare, .claim])
        #expect(flow.stepNumber == 3)
    }

    @Test func theJoinPathWalksEveryBuiltStep() {
        var flow = PromoteFlow(subject: james, builtSteps: allSteps)
        _ = flow.send(.edit(.selectTarget(per1)))
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.step == .compare)
        #expect(flow.stepNumber == 2)
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.step == .claim)
        #expect(flow.stepNumber == 3)
        _ = flow.send(.stepBack)
        #expect(flow.step == .compare)
        _ = flow.send(.advance)
        if case .save = flow.send(.advance).first {} else { Issue.record("expected a save from the last step") }
    }

    @Test func theMintPathSkipsCompare() {
        var flow = PromoteFlow(subject: james, builtSteps: allSteps)
        _ = flow.send(.edit(.choose(.new)))
        _ = flow.send(.advance)
        #expect(flow.step == .claim)
        #expect(flow.stepNumber == 2)
    }

    @Test func stepBackStopsAtTheFirstStep() {
        var flow = PromoteFlow(subject: james, builtSteps: allSteps)
        _ = flow.send(.stepBack)
        #expect(flow.step == .chooseTarget)
    }

    @Test func backAndForwardKeepTheWholeDraft() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        _ = flow.send(.edit(.setConfidence("cg-mod")))
        _ = flow.send(.edit(.setArgument("Same household")))
        let draft = flow.draft
        #expect(flow.send(.stepBack).isEmpty)
        #expect(flow.step == .chooseTarget)
        #expect(flow.draft == draft)
        // Back never asks: nothing is lost.
        #expect(flow.phase == .editing)
        _ = flow.send(.advance)
        #expect(flow.step == .claim)
        #expect(flow.draft == draft)
    }

    @Test func changingTheTargetKeepsTheClaimFields() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        _ = flow.send(.edit(.setArgument("Same household")))
        _ = flow.send(.stepBack)
        _ = flow.send(.edit(.choose(.new)))
        #expect(flow.draft.argument == "Same household")
        #expect(flow.draft.target == nil)
    }

    // MARK: Edits belong to their step

    @Test func claimFieldsEditOnlyOnTheClaimStep() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.setArgument("Too early")))
        _ = flow.send(.edit(.setConfidence("cg-low")))
        #expect(flow.draft.argument.isEmpty)
        #expect(flow.draft.confidenceGradeID == nil)
        flow = mintOnClaim()
        _ = flow.send(.edit(.setConfidence("cg-low")))
        _ = flow.send(.edit(.setArgument("Same household")))
        #expect(flow.draft.confidenceGradeID == "cg-low")
        #expect(flow.draft.argument == "Same household")
        _ = flow.send(.edit(.setConfidence(nil)))
        #expect(flow.draft.confidenceGradeID == nil)
    }

    @Test func theChoiceIsLockedOnLaterSteps() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        _ = flow.send(.edit(.choose(.new)))
        _ = flow.send(.edit(.selectTarget(PromoteFlow.Target(entityID: "e2", ref: "PER-2", title: "J", memberCount: 1))))
        #expect(flow.draft.choice == .existing)
        #expect(flow.draft.target == per1)
    }

    // MARK: Controls

    @Test func controlsFollowThePlanAndTheBuiltSteps() {
        typealias Row = (built: Set<PromoteStep>, join: Bool, advances: Int, back: PromoteStep?, advance: PromoteFlow.Advance)
        let rows: [Row] = [
            // Shipped steps.
            (PromoteStep.built, false, 0, nil, .step(.claim)),
            (PromoteStep.built, false, 1, .chooseTarget, .save),
            (PromoteStep.built, true, 0, nil, .step(.claim)),
            (PromoteStep.built, true, 1, .chooseTarget, .save),
            // With Compare built (S9-19), the join path stops on it — no other changes.
            (allSteps, true, 0, nil, .step(.compare)),
            (allSteps, true, 1, .chooseTarget, .step(.claim)),
            (allSteps, true, 2, .compare, .save),
            (allSteps, false, 1, .chooseTarget, .save),
        ]
        for row in rows {
            var flow = PromoteFlow(subject: james, builtSteps: row.built)
            _ = row.join ? flow.send(.edit(.selectTarget(per1))) : flow.send(.edit(.choose(.new)))
            for _ in 0..<row.advances { _ = flow.send(.advance) }
            let controls = flow.controls
            #expect(controls.back == row.back, "back for \(row)")
            #expect(controls.advance == row.advance, "advance for \(row)")
            #expect(controls.canGoBack == (row.back != nil))
            #expect(controls.canAdvance)
            #expect(!controls.isLocked)
        }
    }

    @Test func controlsLockWhileSavingAndWhenBlocked() {
        var flow = saving()
        var controls = flow.controls
        #expect(controls.back == .chooseTarget)
        #expect(!controls.canGoBack && !controls.canAdvance && controls.isLocked)
        _ = flow.send(.saveFailed(message: "Filed elsewhere", retryable: false))
        controls = flow.controls
        #expect(!controls.canGoBack && !controls.canAdvance && controls.isLocked)
    }

    @Test func nothingChosenCannotAdvance() {
        let controls = PromoteFlow(subject: james).controls
        #expect(controls.back == nil)
        #expect(!controls.canAdvance)
        #expect(!controls.isLocked)
    }

    // MARK: Saving

    @Test func advanceWithoutAChoiceDoesNothing() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.phase == .editing)
    }

    @Test func saveOnTheLastStepWritesTheDraft() {
        var flow = mintOnClaim()
        _ = flow.send(.edit(.setConfidence("cg-mod")))
        _ = flow.send(.edit(.setArgument("Same household")))
        let effects = flow.send(.advance)
        #expect(effects == [.save(PromoteFlow.Save(
            subjectID: "sub-1", entityID: nil, confidenceGradeID: "cg-mod", argument: "Same household"
        ))])
        #expect(flow.phase == .saving(heldLeave: nil))
        #expect(flow.isSaving && !flow.canAdvance)
    }

    @Test func aJoinSavesOntoTheTarget() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        #expect(flow.send(.advance) == [.save(PromoteFlow.Save(
            subjectID: "sub-1", entityID: "e1", confidenceGradeID: nil, argument: ""
        ))])
    }

    @Test func inputIsLockedWhileSaving() {
        var flow = saving()
        #expect(flow.send(.edit(.setArgument("late"))).isEmpty)
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.send(.stepBack).isEmpty)
        #expect(flow.send(.done).isEmpty)
        #expect(flow.draft.argument.isEmpty)
        #expect(flow.step == .claim)
    }

    @Test func aSuccessfulSaveAnnouncesItAndReturnsToTheGraph() {
        var flow = saving()
        #expect(flow.send(.saveSucceeded(entityRef: "PER-9")) == [
            .refreshAfterSave,
            .announceFiled(subjectName: "James Robins", entityRef: "PER-9"),
            .navigateToGraph,
        ])
        #expect(flow.phase == .finished)
        #expect(flow.savedCount == 1)
        #expect(!flow.hasUnsavedWork)
    }

    @Test func aRetryableFailureKeepsTheStepAndShowsWhy() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        _ = flow.send(.edit(.setArgument("Same household")))
        _ = flow.send(.advance)
        #expect(flow.send(.saveFailed(message: "Disk full", retryable: true)).isEmpty)
        #expect(flow.phase == .editing)
        #expect(flow.step == .claim)
        #expect(flow.error == "Disk full")
        #expect(flow.draft.argument == "Same household")
        #expect(flow.canAdvance)
        #expect(flow.hasUnsavedWork)
        // The next edit clears it.
        _ = flow.send(.edit(.setArgument("Same household, York")))
        #expect(flow.error == nil)
    }

    @Test func aFinalFailureBlocksTheStep() {
        var flow = saving()
        #expect(flow.send(.saveFailed(message: "Already a member", retryable: false)) == [.refreshAfterSave])
        #expect(flow.failure == PromoteFlow.Failure(message: "Already a member"))
        #expect(flow.isBlocked)
        #expect(flow.error == nil)
        #expect(!flow.canAdvance)
        #expect(!flow.hasUnsavedWork)
        #expect(flow.send(.edit(.setArgument("x"))).isEmpty)
        #expect(flow.send(.advance).isEmpty)
        // Nothing to lose: leaving and Done go without asking.
        #expect(flow.requestLeave(.back) == .allow)
        #expect(flow.send(.done) == [.navigateToGraph])
    }

    @Test func aBlockedStepLearnsWhichHandleWithoutLeaving() {
        var flow = saving()
        _ = flow.send(.saveFailed(message: "Already a member", retryable: false))
        #expect(flow.send(.subjectUnavailable(filedOn: "PER-7")).isEmpty)
        #expect(flow.failure?.filedOn == "PER-7")
        #expect(flow.isBlocked)
        #expect(flow.send(.subjectUnavailable(filedOn: nil)).isEmpty)
        #expect(flow.failure?.filedOn == "PER-7")
    }

    @Test func aFinalFailureLetsAHeldLeaveGo() {
        var flow = saving()
        _ = flow.requestLeave(.back)
        #expect(flow.send(.saveFailed(message: "Already a member", retryable: false)) == [
            .refreshAfterSave, .resumeNavigation,
        ])
    }

    @Test func resultsOutsideAWriteAreIgnored() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.saveSucceeded(entityRef: "PER-1")).isEmpty)
        #expect(flow.send(.saveFailed(message: "x", retryable: true)).isEmpty)
        #expect(flow.send(.saveFailed(message: "x", retryable: false)).isEmpty)
        #expect(flow.savedCount == 0)
        #expect(flow.error == nil)
        #expect(!flow.isBlocked)
    }

    // MARK: Leaving

    @Test func leavingUntouchedIsAllowed() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.requestLeave(.back) == .allow)
        #expect(flow.phase == .editing)
    }

    @Test func leavingWithAChoiceAsks() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.choose(.new)))
        #expect(flow.requestLeave(.back) == .hold)
        #expect(flow.pendingLeave == .back)
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.send(.leaveCancelled) == [.cancelNavigation])
        #expect(flow.phase == .editing)
        #expect(flow.requestLeave(.forward) == .hold)
        #expect(flow.send(.leaveConfirmed) == [.resumeNavigation])
        #expect(flow.phase == .finished)
        #expect(flow.requestLeave(.back) == .allow)
    }

    @Test func doneOnClaimFieldsAsksThroughTheGuard() {
        var flow = mintOnClaim()
        #expect(flow.send(.done) == [.navigateToGraph])
        // The navigation it starts is what the guard holds.
        #expect(flow.requestLeave(.location(.sectionRoot(.sources))) == .hold)
        #expect(flow.pendingLeave != nil)
    }

    @Test func aSecondLeaveReplacesTheQuestion() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.choose(.new)))
        _ = flow.requestLeave(.back)
        _ = flow.requestLeave(.forward)
        #expect(flow.pendingLeave == .forward)
    }

    @Test func leavingDuringAWriteWaitsForIt() {
        var flow = saving()
        #expect(flow.requestLeave(.back) == .hold)
        #expect(flow.phase == .saving(heldLeave: .back))
        #expect(flow.pendingLeave == nil)
        #expect(flow.send(.saveSucceeded(entityRef: "PER-9")) == [
            .refreshAfterSave,
            .announceFiled(subjectName: "James Robins", entityRef: "PER-9"),
            .resumeNavigation,
        ])
    }

    @Test func aRetryableFailureDropsTheHeldLeave() {
        var flow = saving()
        _ = flow.requestLeave(.back)
        #expect(flow.send(.saveFailed(message: "x", retryable: true)) == [.cancelNavigation])
        #expect(flow.phase == .editing)
    }

    @Test func doneAsksTheGraphToOpen() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.done) == [.navigateToGraph])
    }

    // MARK: Subject gone

    @Test func aGoneSubjectEndsTheFlow() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.choose(.new)))
        #expect(flow.send(.subjectUnavailable(filedOn: nil)) == [.navigateToGraph])
        #expect(flow.phase == .finished)
        #expect(flow.send(.subjectUnavailable(filedOn: "PER-1")).isEmpty)
    }

    @Test func aGoneSubjectWhileAskingDropsTheQuestion() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.edit(.choose(.new)))
        _ = flow.requestLeave(.back)
        #expect(flow.send(.subjectUnavailable(filedOn: "PER-1")) == [.cancelNavigation, .navigateToGraph])
    }

    // MARK: Walk

    @Test func aWalkMovesToTheNextSubjectAfterEachSave() {
        var flow = mintOnClaim()
        flow.enqueue([mary, james])
        #expect(flow.queue.map(\.id) == ["sub-1", "sub-2"])
        _ = flow.send(.advance)
        #expect(flow.send(.saveSucceeded(entityRef: "PER-9")) == [
            .refreshAfterSave,
            .announceFiled(subjectName: "James Robins", entityRef: "PER-9"),
        ])
        #expect(flow.subject == mary)
        #expect(flow.phase == .editing)
        #expect(flow.step == .chooseTarget)
        #expect(flow.draft == PromoteFlow.Draft())
        #expect(flow.savedCount == 1)
        _ = flow.send(.edit(.selectTarget(per1)))
        _ = flow.send(.advance)
        if case .save(let save) = flow.send(.advance).first {
            #expect(save.subjectID == "sub-2")
        } else {
            Issue.record("expected a save")
        }
        #expect(flow.send(.saveSucceeded(entityRef: "PER-1")) == [
            .refreshAfterSave,
            .announceFiled(subjectName: "Mary Robins", entityRef: "PER-1"),
            .navigateToGraph,
        ])
        #expect(flow.savedCount == 2)
        #expect(flow.phase == .finished)
    }
}
