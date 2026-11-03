# Review — S14-20 The Promote page recovers and speaks

**Verdict:** Fits with changes — Ready with changes (every round-2 blocker Fixed or answered by the user)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **M-4 · Move `CatalogCounts.refreshAll` onto the same queue?** The writer recommended logging it. User, 2026-11-03: no. Keep the Promote flow simple; the page may assume its data is static while it is open, as long as the backend refuses a stale Done with an error. The generic runner went with that answer (D-1).
- **M-10 · One retry callout for every failed read?** The writer recommended composing Promote's callout and logging the unification. User, 2026-11-03: extract the shared piece now. D-14 adds `PVRetryCallout`, and the Evidence graph canvas and Source forms move onto it. Their looks change without a board, which the plan's **Rules this plan bends** records.
- **M-6 · Load-error copy.** The writer recommended rewording `graphUnavailable` and keeping "Try again". User, 2026-11-03: reword, keep "Try again" (D-11).

## Writer's own calls

- The proposal queue is two private members of `PromoteModel`, not a shared type (D-1), following the user's answer to M-4.
- PN-3 (a failed re-proposal clears the stale notice) stays: it is on the agreed board (frame 4) and costs one line.
- Try again reads only the graph key again (D-7): properties and rules are optional in `graphSnapshot()`.
- A remount (`.id(loadAttempt)`) re-announces a repeated failure. The kit gets no repeat key (D-8; M-5).
- Version bump 0.2.13 (`main` is at 0.2.12 after S14-22).

## Board

The S14-D7 board (exported `Promote flow.dc.html`, section "S14-D7", frames 1a – 4) was checked against the brief. It uses kit Callout and Button only; PN-1 is option (a), with option (b) deleted; every PN requirement has a frame. Try again is `size: 'sm'` (`PVControlSize.sm` exists). The second load message on frame 1a is marked a board placeholder.

## Round 1

### M-1 · blocker · The scroll to the notices is built by hand next to the kit's scroll mechanism
`PVScrollRowIntoView.swift` `pvScrollsRowIntoView(…whenRequested:)` already scrolls to an id on a token, waits a turn for an inserted view, and honors Reduce Motion; `DesignSystem/README.md` forbids another `ScrollViewReader`. (Same finding as X-1.)
**Answer:** Fixed — D-10 uses `.pvScrollsRowIntoView(PromoteView.noticesAnchor, anchor: .top, whenRequested: model.noticeToken)`; Codebase fit updated.

### M-2 · blocker · Try again skips the session's read-again path, so a graph-missing load cannot recover
`WorkspaceSession.readyValue` returns nil for a failed key; `reread` is the read-again path (`EvidenceGraphModel.readGraphAgain()`).
**Answer:** Fixed — D-7: `retryLoad()` calls `session.reread(graphKey)` when the key has no value; T-10 retries the graph-missing case; Codebase fit and §4 name `readGraphAgain` and `SourceFormLoadArea`.

### M-3 · major · The tests pause the fake with sleeps; `add-swift-test` says to use `FakeHold`
**Answer:** Fixed — D-13 adds `proposePromoteGraphAlignmentHold`; `promoteProposeDelays` and its claim go (Deletes); STEP-2 moves the two delay tests onto the hold; §8 fixtures use the hold and `waitUntil`.

### M-4 · major · `LatestRunner` has one caller; the second one with the same defect is not named
**Answer:** To user (answered above) — the generic runner is gone; the queue is private to `PromoteModel` (D-1); the recount is an Out row, declined by the user.

### M-5 · minor · The re-announcement is fixed with a remount at the call site, not in the kit
**Answer:** Declined — one call site in this PR; D-8's rejected column and an Out row record it. A second retry callout that announces is where the kit key earns its place.

### M-6 · major · The load-error copy still describes the workaround Try again replaces
**Answer:** To user (answered above) — D-11 and Strings reword `graphUnavailable`; the button stays "Try again".

### M-7 · minor · The domain doc sweep misses §12
**Answer:** Fixed — STEP-1 rewords §12 line 242; the DoD sweep `rg`s `provenance is 1`. (Same as X-7.)

### M-8 · minor · Load state is spread across three members, and the notice token is bumped in three places
**Answer:** Fixed in part — D-10's one helper, `showingNotices(_:)`, owns every bump. Declined for the load enum: three members on one page, and an enum would not remove one (D-7's rejected column).

### X-1 · major · D-10 adds a bare `ScrollViewReader`, which the kit README forbids
**Answer:** Fixed — see M-1.

### X-2 · major · `loadAttempt` is bumped when the retry starts, so VoiceOver hears the old message at that moment
**Answer:** Fixed — D-7 bumps `loadAttempt` on each load failure; T-6 and T-7 count it.

### X-3 · major · `retryLoad()` is not testable as written, and its guard can race
**Answer:** Fixed — D-7 sets `isLoading = true` synchronously in `retryLoad()`; T-6 and T-7 wait on `!isLoading`, and T-7 presses twice in a row.

### X-4 · major · `noticeToken` bumping is underspecified
**Answer:** Fixed — D-10: one helper compares the three notices before and after a change and bumps once when any went from nil to set; T-8 states the counts.

### X-5 · major · Construction of the runner and the proposal helper are left open
**Answer:** Fixed — the runner type is gone (M-4). D-1 names the private members, the loop, its weak-per-iteration capture and what happens when the page leaves; D-6 names `submitProposal()` as the one path for both callers.

### X-6 · major · AC-3.1's `rg` cannot give "no hit" as written
**Answer:** Fixed — AC-3.1 and T-9 use `rg -n 'Subject\.Provenance|Provenance:|provenance float64' core`.

### X-7 · minor · Stale docs lines the plan does not touch
**Answer:** Fixed — §12 in STEP-1; `types.go` line 18 and the `score.go` doc lines in Deletes.

### X-8 · minor · Test fixtures for T-5 and T-13
**Answer:** Fixed — §8 Fixtures: `mapAll: true` for mapped pages, the hold protocol, and D-13 records a call before it waits or throws.

### X-9 · minor · Checks I could not run
**Answer:** Declined — `xcodebuild` runs on macOS in `scripts/check-pr.sh` and CI; the manual check covers the scroll and announcement of the stale notice, and T-8 covers when each notice bumps the token.

### X-10 · minor · Header state
**Answer:** Fixed — status moves to Planned when round 2 closes.

### X-11 · nit · Minor wording
Logged. STEP-3 was split: the FakeStore hooks are STEP-2.

## Round 2

### M-9 · blocker · `isProposing` and `proposalRunning` store one fact twice
**Answer:** Fixed. D-1 and D-3 now use `isProposing` as the queue's only running flag, and `proposalRunning` is gone from D-1 and Codebase fit.

### M-10 · blocker · A third hand-built "failed read, retry" callout, each with its own busy state
**Answer:** To user, answered 2026-11-03: extract it now. D-14 adds `PVRetryCallout` in `Components/Callout/`, and STEP-4 moves `SourceFormLoadArea` and `EvidenceGraphView.readErrorCallout` onto it. T-14 checks that both copies are gone. **Rules this plan bends** records the look change on the canvas.

### M-11 · blocker · `retryLoad()` decides "the graph read is not usable" with a different rule than `readyValue`
**Answer:** Fixed. D-7 rereads the graph key unconditionally, as `readGraphAgain()` does, and T-12 covers a stale key that still holds a value. (Same as X-3.)

### Note · the map's scope entry still promised a shared runner
**Answer:** Fixed. The S14-20 row and its **Ships** entry in `deployment-plan.md` now describe the queue as private to `PromoteModel`, and name `PVRetryCallout`.

### X-1 · blocker · AC-3.1 and T-9's `rg` cannot return "no hit"
**Answer:** Fixed. The pattern is scoped to `core/graphalign core/database/promotealign`, in AC-3.1, T-9 and STEP-1.

### X-2 · blocker · STEP-2 cannot move `anOlderProposalAnsweringLastIsDropped` onto `FakeHold`
**Answer:** Fixed. STEP-2 moves only `doneWaitsForAPendingProposal`. STEP-3 deletes `anOlderProposalAnsweringLastIsDropped`, because the queue rules out the overlap it tests and T-2 replaces it. `promoteProposeDelays` goes in STEP-3 as well.

### X-3 · blocker · D-7's reread condition leaves a stale graph key unrecoverable
**Answer:** Fixed. See M-11.

### X-4 · major · T-2 leaves the test setup to the implementer
**Answer:** Fixed. T-2 now names the subjects, the three scripted answers, the tokens, both releases and their waits. STEP-2 places the hold's `wait()` after the `withState` claim.

### X-5 · major · A fixture sentence is false for T-1 to T-8
**Answer:** Fixed. The sentence now says only that T-2 opens with `mapAll: true`.

### X-6 · minor · The Status and Review lines still read "Blocked" and "round 2 pending"
**Answer:** Fixed. Status is Planned, and the Review line links this log.
