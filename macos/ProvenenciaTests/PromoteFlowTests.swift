import Testing
@testable import Provenencia

/// The Promote state machine on its own: every transition, no store or view.
@Suite
struct PromoteFlowTests {
    private let james = PromoteFlow.Subject(id: "sub-1", ref: "CPR-1", name: "James Robins", kind: .person)
    private let mary = PromoteFlow.Subject(id: "sub-2", ref: "CPR-2", name: "Mary Robins", kind: .person)
    private let per1 = PromoteFlow.Target(entityID: "e1", ref: "PER-1", title: "James Robins", memberCount: 2)
    private let allSteps = Set(PromoteStep.allCases)

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
        #expect(flow.send(.choose(.existing)).isEmpty)
        #expect(flow.plan == [.chooseTarget, .compare, .claim])
        #expect(!flow.canAdvance)
        _ = flow.send(.selectTarget(per1))
        #expect(flow.canAdvance)
        _ = flow.send(.choose(.new))
        #expect(flow.draft.target == nil)
        #expect(flow.plan == [.chooseTarget, .claim])
        #expect(flow.canAdvance)
    }

    @Test func selectingATargetChoosesExisting() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.selectTarget(per1))
        #expect(flow.draft.choice == .existing)
        #expect(flow.draft.target == per1)
    }

    // MARK: Advancing and saving

    @Test func advanceWithoutAChoiceDoesNothing() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.phase == .editing)
    }

    @Test func advancePastTheLastBuiltStepSaves() {
        var flow = PromoteFlow(subject: james, builtSteps: [.chooseTarget])
        _ = flow.send(.choose(.new))
        let effects = flow.send(.advance)
        #expect(effects == [.save(PromoteFlow.Save(subjectID: "sub-1", entityID: nil, confidenceGradeID: nil, argument: ""))])
        #expect(flow.phase == .saving(heldLeave: nil))
        #expect(flow.isSaving && !flow.canAdvance)
    }

    @Test func aJoinSavesOntoTheTarget() {
        var flow = PromoteFlow(subject: james, builtSteps: [.chooseTarget])
        _ = flow.send(.selectTarget(per1))
        #expect(flow.send(.advance) == [.save(PromoteFlow.Save(subjectID: "sub-1", entityID: "e1", confidenceGradeID: nil, argument: ""))])
    }

    @Test func inputIsLockedWhileSaving() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        _ = flow.send(.advance)
        #expect(flow.send(.choose(.existing)).isEmpty)
        #expect(flow.send(.advance).isEmpty)
        #expect(flow.send(.done).isEmpty)
        #expect(flow.draft.choice == .new)
    }

    @Test func aSuccessfulSaveFinishesAndReturnsToTheGraph() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        _ = flow.send(.advance)
        #expect(flow.send(.saveSucceeded) == [.refreshAfterSave, .navigateToGraph])
        #expect(flow.phase == .finished)
        #expect(flow.savedCount == 1)
        #expect(!flow.hasUnsavedWork)
    }

    @Test func aFailedSaveKeepsTheDraftAndShowsWhy() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.selectTarget(per1))
        _ = flow.send(.advance)
        #expect(flow.send(.saveFailed(message: "Already a member")).isEmpty)
        #expect(flow.phase == .editing)
        #expect(flow.error == "Already a member")
        #expect(flow.draft.target == per1)
        #expect(flow.canAdvance)
        // The next edit clears it.
        _ = flow.send(.choose(.new))
        #expect(flow.error == nil)
    }

    @Test func resultsOutsideAWriteAreIgnored() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.saveSucceeded).isEmpty)
        #expect(flow.send(.saveFailed(message: "x")).isEmpty)
        #expect(flow.savedCount == 0)
        #expect(flow.error == nil)
    }

    // MARK: Steps (as later PRs build them)

    @Test func theJoinPathWalksEveryBuiltStep() {
        var flow = PromoteFlow(subject: james, builtSteps: allSteps)
        _ = flow.send(.selectTarget(per1))
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
        _ = flow.send(.choose(.new))
        _ = flow.send(.advance)
        #expect(flow.step == .claim)
        #expect(flow.stepNumber == 2)
    }

    @Test func unbuiltStepsAreSkipped() {
        var flow = PromoteFlow(subject: james, builtSteps: [.chooseTarget, .claim])
        _ = flow.send(.selectTarget(per1))
        _ = flow.send(.advance)
        #expect(flow.step == .claim)
        #expect(flow.plan == [.chooseTarget, .compare, .claim])
        #expect(flow.stepNumber == 3)
    }

    @Test func stepBackStopsAtTheFirstStep() {
        var flow = PromoteFlow(subject: james, builtSteps: allSteps)
        _ = flow.send(.stepBack)
        #expect(flow.step == .chooseTarget)
    }

    @Test func theChoiceIsLockedOnLaterSteps() {
        var flow = PromoteFlow(subject: james, builtSteps: allSteps)
        _ = flow.send(.selectTarget(per1))
        _ = flow.send(.advance)
        _ = flow.send(.choose(.new))
        #expect(flow.draft.choice == .existing)
    }

    // MARK: Leaving

    @Test func leavingUntouchedIsAllowed() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.requestLeave(.back) == .allow)
        #expect(flow.phase == .editing)
    }

    @Test func leavingWithAChoiceAsks() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
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

    @Test func aSecondLeaveReplacesTheQuestion() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        _ = flow.requestLeave(.back)
        _ = flow.requestLeave(.forward)
        #expect(flow.pendingLeave == .forward)
    }

    @Test func leavingDuringAWriteWaitsForIt() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        _ = flow.send(.advance)
        #expect(flow.requestLeave(.back) == .hold)
        #expect(flow.phase == .saving(heldLeave: .back))
        #expect(flow.pendingLeave == nil)
        #expect(flow.send(.saveSucceeded) == [.refreshAfterSave, .resumeNavigation])
    }

    @Test func aFailedWriteDropsTheHeldLeave() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        _ = flow.send(.advance)
        _ = flow.requestLeave(.back)
        #expect(flow.send(.saveFailed(message: "x")) == [.cancelNavigation])
        #expect(flow.phase == .editing)
    }

    @Test func doneAsksTheGraphToOpen() {
        var flow = PromoteFlow(subject: james)
        #expect(flow.send(.done) == [.navigateToGraph])
    }

    // MARK: Subject gone

    @Test func aGoneSubjectEndsTheFlow() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        #expect(flow.send(.subjectUnavailable) == [.navigateToGraph])
        #expect(flow.phase == .finished)
        #expect(flow.send(.subjectUnavailable).isEmpty)
    }

    @Test func aGoneSubjectWhileAskingDropsTheQuestion() {
        var flow = PromoteFlow(subject: james)
        _ = flow.send(.choose(.new))
        _ = flow.requestLeave(.back)
        #expect(flow.send(.subjectUnavailable) == [.cancelNavigation, .navigateToGraph])
    }

    // MARK: Walk

    @Test func aWalkMovesToTheNextSubjectAfterEachSave() {
        var flow = PromoteFlow(subject: james)
        flow.enqueue([mary, james])
        #expect(flow.queue.map(\.id) == ["sub-1", "sub-2"])
        _ = flow.send(.choose(.new))
        _ = flow.send(.advance)
        #expect(flow.send(.saveSucceeded) == [.refreshAfterSave])
        #expect(flow.subject == mary)
        #expect(flow.phase == .editing)
        #expect(flow.step == .chooseTarget)
        #expect(flow.draft == PromoteFlow.Draft())
        #expect(flow.savedCount == 1)
        _ = flow.send(.selectTarget(per1))
        if case .save(let save) = flow.send(.advance).first {
            #expect(save.subjectID == "sub-2")
        } else {
            Issue.record("expected a save")
        }
        #expect(flow.send(.saveSucceeded) == [.refreshAfterSave, .navigateToGraph])
        #expect(flow.savedCount == 2)
        #expect(flow.phase == .finished)
    }
}
