# Review — S16-04 The Person page says when a reload failed or the diagram is partial

Covers the plan [`S16-04-person-page-notices.md`](S16-04-person-page-notices.md) and its brief [S16-D1](../design/S16-D1-person-page-notices.md).

**Verdict:** Fits with changes — Ready with changes (round 2 findings all Fixed; no open blocker). Round 1 was Fits with changes — Not ready (plan) and Fits with changes — Ready with changes (brief).
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet), brief (Opus, maintainability UI rows and executability for briefs)

## To user

- **Board S16-D1.** Paste the brief into Claude Design ("Person Detail" board, new turn "S16-D1 · Notices") and agree the board. Until then the plan stays `Blocked — on board S16-D1`.
- **B-2: Remove on a stale line.** After a failed reload, a record line already removed stays under the notice, and a second Remove on it is a silent no-op (`core/database/promote/unpromote.go`). Recommendation: leave it as it is in this PR. The notice says the rows are old and Try again is one press away. Add a line to [`ideas/person-page-followups.md`](../../../ideas/person-page-followups.md). The other choices are to disable Remove under a stale notice, or to have a no-op Remove say so; either one adds a frame and a Decision.
- **M-3 / M2-4 widen the PR to Go, in a matching-layer package.** A not-counted read now lists no pair. The change sits in `recordlinks.Pairs` (D-9), with two rewritten subtests and two lines in `docs/matching-model.md`. The writer accepted it because the reviewers' reason holds: `Pairs` owns `MaxReadWork`, and it stops building up to about 1.1M lines only to have them dropped. If you want S16-04 to stay Swift-only, say so; the drop would then move back to Swift and be logged as debt.

## Writer's own calls

- D-1: the name `QueryReading` and the property `reading` (the working rule's word). `failure` stays.
- D-8: the stale notice's id is `.refreshError`, the first-load and gone notices' id is `.error`.
- D-12 and the brief's copy, all new strings: "Try again" (as on Promote), the stale sentence, "This page could not be read", the Person gone line, and the two partial notes.
- PN-9: a not-counted read draws no arc, rather than giving uncompared arcs a stroke style of their own.
- PN-10: when the Sources rows fail, there is no record links column (today's structure).
- D-14: the manual check runs failure states in previews, because a real project cannot be made to fail on demand.

## Round 1

### M-1 · blocker · The new reading is a second way to see a refresh failure; `QueryHandle.failure` and the Evidence graph's counter already do this job

`QueryHandle.failure` and `EvidenceGraphModel.graphReadAttempt` / `graphReadFailedSettled` / the wait `Task` cover the same ground. Move the callers onto one mechanism, delete the counter and update the doc line.

**Answer:** Fixed in part, and Declined in part.
- Fixed: `failureRevision` and `isReading` replace the Evidence graph's counter and wait `Task`. See D-5, STEP-4, T-10 and **Deletes**. `reading` is built from `failure` (D-1). STEP-10's doc line now says when to use `read` and when to use `reading`.
- Declined: deleting `failure`. `SourcesModel.swift:147-148, 203` and `SourceTypesModel.swift:100` read it for list-level states that replace the list, not a stale value beside it. `reading` is composed from `failure`, not built beside it, so there is still one notion of a failure. This is in the Out table.

### M-2 · blocker · The re-announce and focus-after-Try-again rule sits in a feature snowflake; it belongs in `PVRetryCallout`

The rule would have three copies. The S15-21 sheet is a recipe and cannot name a `Features/` type.

**Answer:** Fixed. `ConclusionReadNotice` is gone. `PVRetryCallout` gains `failureRevision: Int?` (D-4), which carries the remount, `retryPressed` and the claim after a press. Callers that do not pass it are unchanged. The brief's inventory marks `PVRetryCallout` as Extend and has no wrapper.

### M-3 · blocker · Dropping uncompared arcs in Swift is a client decision keyed to an engine flag

**Answer:** Fixed. In D-9, STEP-8 and T-9, `persondetail` sends no pair when not counted. Swift draws what it is sent and uses the flag only for the note and the spoken label (D-10). This is listed under To user because it widens the PR to Go.

### M-4 · minor · `QueryRead` vs `QueryReading` names are easily confused

**Answer:** Declined, but the doc half is Fixed. The name is the working rule's own word ("one reading"). The `macos-client-patterns.md` row in STEP-10 says which accessor a reader picks, and what `isReading` means.

### M-5 · minor · `goneNotice` is threaded through six members; the precedent is a kind-keyed L10n function

**Answer:** Fixed. D-7 adds `L10n.Conclusions.goneNotice(section:)`. The page and the host resolve it from `WorkspaceNavigation`, which the host now reads from the environment. No initializer changes.

### X-1 · blocker · The plan is Blocked on an undrawn board

**Answer:** Declined, because this is the intended state. `plan-pr` step 3 says a written brief whose board is not agreed leaves the plan `Blocked — on board`. The map row now says so too. It is in To user.

### X-2 · blocker · `check-plan-docs.py` fails: missing review log, map has no link

**Answer:** Fixed. This log exists, and the map row links the plan.

### X-3 · blocker · No Swift check can run on this machine

**Answer:** Declined. The repo is a macOS app, and `scripts/check-pr.sh` runs `xcodebuild` by design. The **Planned** line now says that the Swift steps need a Mac with Xcode.

### X-4 · major · Manual checks need failure hooks that do not exist

**Answer:** Fixed. D-13 adds FakeStore error hooks for the detail, Family and claims reads. D-14 adds `PersonDetailNoticePreviews.swift` with one preview per failure state. §10 says, for each line, whether it runs in a Preview or in the App.

### X-5 · major · §10 numbering and scope inconsistent

**Answer:** Fixed. §10 has been rewritten, with nine numbered lines, each marked Preview or App.

### X-6 · major · T-3 and T-4 not written in any step; T-3 does not test wiring

**Answer:** Fixed. STEP-3 writes T-3 and T-4. T-3 now tests `goneNotice(section:)` for Persons, Events and Places, which is the wiring D-7 relies on.

### X-7 · major · The partial-note previews point at a view that does not draw it

**Answer:** Fixed. STEP-9 names `PersonRecordLinksColumn` previews: the environment viewport, the hover state, the closures, and the `manyRecordLinks(graphs:)` fixture for each of the four states.

### X-8 · major · Plumbing `goneNotice` and the band's retry under-specified

**Answer:** Fixed. M-5 removes the plumbing. D-6 names the band's new inputs and gives the `onRetry` closure verbatim. STEP-5 says neither initializer changes.

### X-9 · major · When `retryPressed` resets is not decided

**Answer:** Fixed. D-4 clears it in the `FocusClaim` `then:` closure after the claim. A notice that succeeds unmounts and takes the flag with it (Edges).

### X-10 · minor · Column spacing and the stale notice's id are unspecified

**Answer:** Fixed. D-8 sets `space7` in the band and `space4` in the column, and gives the stale notice the id `.refreshError`, distinct from `.error`.

### X-11 · minor · D-8 and STEP-6 refer to UI tests that do not exist

**Answer:** Fixed. The claim is dropped. §4's Tests row says that no UI test target exists.

### X-12 · minor · Step 9 paths are bare names

**Answer:** Fixed. STEP-10 spells out every path, including `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj`.

### X-13 · minor · D-9's distinct-graph count wording

**Answer:** Fixed. D-10 now reads: "met a graph id not yet kept while `graphs.count == maxNodes`".

### X-14 · nit · Temptation

Logged.

### B-1 · major · "The strongest links are drawn" is false when the 1,000-node cap does the cutting

**Answer:** Fixed. The note now reads "Too many records to draw them all. Some records and links are left out." in brief PN-9, the plan's Strings and AC-4.2.

### B-2 · major · A Remove on a stale line is still a silent no-op

**Answer:** To user (above). The brief's §2.1 and the plan's Out table both point to the question.

### B-3 · major · The inventory leaves out the snowflake the gated plan builds

**Answer:** Fixed by M-2. There is no snowflake any more. The inventory's `PVRetryCallout` row is Extend, and the non-goals say there is no wrapper.

### B-4 · minor · Combined states left for Claude Design

**Answer:** Fixed. The new PN-10 covers four things: no column when the rows fail or are gone; no rows under a gone notice; a notice above the EmptyState; and the column's top aligned with the rows' top.

### B-5 · minor · The not-compared spoken label has no copy

**Answer:** Fixed. PN-9 gives the label verbatim with an example, and the plan's Strings carry `nodeUncountedAccessibility`.

### B-6 · minor · PN-W2's stacked column and Event/Place at Wide have no frame

**Answer:** Fixed. There is a new frame 2c, the column stacked at a 900 pt pane. The Event/Place Wide row now says "not drawn: nothing moves".

### B-7 · minor · Frame 8a asks for a Persons list the board does not hold

**Answer:** Fixed. Frame 8 now draws the Toast alone, over a blank pane.

### B-8 · nit · Toast row Home and README wording

**Answer:** Fixed. The inventory row gives both paths, and the README row says "partial".

### B-9 · nit · "Try again" vs "Read again"

**Answer:** Logged. The inventory says "Try again", as on Promote. The split is an Out line ("not planned").

### B-10 · nit · Sample error text is not a real engine message

**Answer:** Fixed. The frames use the generic error "Something went wrong. Please try again.", marked as sample text.

## Round 2

Fresh reviewers had the plan, the brief and this log. They accepted the round-1 answers M-4, M-5, X-1 – X-14 and B-1 – B-10, and M-1's decline to delete `failure`. Every new finding is Fixed.

### M2-1 / X2-1 · blocker · The Evidence graph's notice reads `failure`, which is nil during a first-load retry, so the notice unmounts
**Answer:** Fixed. In D-5, `graphReadError` reads `handle.reading`. D-1 now says `.failed` comes from the stored `error`, and that it holds through a retry. `failure`'s doc comment says it is nil while a retry runs. A new test, T-10 `aFirstLoadRetryKeepsTheReadError`, checks that the error stays during the retry.

### M2-2 · blocker · Promote still remounts on its own counter
**Answer:** Fixed. In D-4, `PromoteView` passes `failureRevision: model.loadAttempt` and drops `.id`. It keeps its binding and its focus on appear, and the claim goes through the caller's binding when one is passed. The Out row is deleted. Manual check 9 covers Promote.

### M2-3 · blocker · `CitationComposerModel.isGraphReloading` is `isReading` already
**Answer:** Fixed. In D-2, `isGraphReloading` returns `isReading`. Codebase fit and Deletes now list it.

### M2-4 / X2-3 · blocker · The "not compared → no pair" rule belongs in `recordlinks.Pairs`; existing tests and docs assert the old shape
**Answer:** Fixed. In D-9, `Pairs` returns `nil, false` past `MaxReadWork`, and the `skip` path goes; `persondetail` changes only a comment. T-9 rewrites `TestPairsCounted` "AC-2.2" and `TestRecordLinksSaysWhenPartial` "past the work bound by pair work". STEP-8 updates `docs/matching-model.md` (line 33 and the knob row) and loads `add-match-rule` and `match-rules.mdc`. A Stop condition covers any other caller of `Pairs`.

### M2-5 / X2-5 · blocker · The host's new `WorkspaceNavigation` breaks the Sources previews; the page previews had no fixture and the wrong section
**Answer:** Fixed. In D-14, `PersonSourcesPreview` gains a `WorkspaceNavigation` sent to a Person location, in STEP-6, the same commit as the host change. The record links previews reuse that harness: its existing "Diagram error", plus the four partial states. The page harness names its seeded detail, ids, `userID`, navigation and frames.

### X2-2 · major · `generate-proto.sh` cannot run on Linux
**Answer:** Fixed. The FFI section and STEP-8 run it on the Mac, name the tools it needs and the two generated files, and check that the diff is comment-only.

### X2-4 · major · `manyRecordLinks(graphs: 1_001)` yields 50 graphs
**Answer:** Fixed. STEP-9 adds `crowdClaims(graphs:)` and `crowdRecordLinks(graphs:)` (in Codebase fit) and spells out "pairs emptied" as a copy changed through its `var`s.

### X2-6 · minor · `readAgainShowsItIsReading` reads `graphReadAttempt`
**Answer:** Fixed. STEP-4 names the line's replacement. T-10 is a new test.

### X2-7 · minor · `goneNotice(section:)` return type
**Answer:** Fixed. D-7 now declares the return type: `-> LocalizedStringResource`.

### X2-8 · minor · Small gaps
**Answer:** Fixed. All four `MARKETING_VERSION` lines are now covered. D-6 builds the `.conclusionDetail` key inline. D-4 says what happens when a caller passes both parameters. STEP-2 names the `.content` patterns. STEP-1 says how T-1 matches a non-`Equatable` value.
