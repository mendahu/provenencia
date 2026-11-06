# Review — S13-17 Remove a claim

**Verdict:** Ready after the refresh rounds below (2026-11-06); no blocker open. Earlier: Reshape — Not ready (round 1, on the draft). Reshaped with the user's answer of 2026-11-05; every finding Fixed.
**Rounds:** 1 (round 2 runs when S13-37 is planned, S13-16 has landed and board S13-D6 is agreed). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **The engine reports what Unpromote removes, as a preparatory PR?** (M-2, M-3) **Answered 2026-11-05: yes.** S13-37 is added to the map ahead of S13-17. It covers the preflight's bridge ids and `WriteEffects.deleted_handles`.

## Writer's own calls

- D-5: the Evidence graph's single-handle confirm shows the per-kind sentence approved for S13-08 (2026-11-04). That sentence is defined but has no caller today. This changes copy the user sees on the graph.
- D-7: the page leaves for the Persons list before the detail reload paints.
- D-8: `userID` is added to the page-sections context for every section.
- D-10: after a kept Unpromote, focus moves to the next card's control, else the previous one's, else the first group title.

## Round 1 — maintainability

### M-1 · blocker · Lift the flow, not just its item
**Answer:** Fixed. D-1 adds `UnpromoteFlow`, shaped like `DeleteImpactFlow`, and both pages hold one. `PersonClaimRemoval` is gone, along with its duplicate tests.

### M-2 · blocker · The person-gone fact should come from the write
**Answer:** To user, then Fixed. With the user's yes, D-6 reads `effects.deletedHandles` (S13-37). The preflight inference, the detail read and T-8's fallback are gone.

### M-3 · blocker · The bridge rule is copied in Swift
**Answer:** To user, then Fixed. D-3 takes the preflight's bridge ids (S13-37), and the Swift filter is deleted.

### M-4 · major · The fourth copy of reading a snapshot
**Answer:** Fixed. D-4 adds `WorkspaceSession.evidenceGraphSnapshot(graphID:)`, and `PromoteModel.graphSnapshot()` moves onto it.

### M-5 · major · "Deletes its own handle" is decided twice
**Answer:** Fixed. D-2 adds one `deletesOwnHandle`. The leave rule no longer needs it (D-6).

### M-6 · minor · One action has two names
**Answer:** Fixed. Code names, keys and ids say Unpromote; "Remove" stays only in the map title.

### M-7 · minor · `userID` and `store`
**Answer:** Fixed. D-8 adds `userID` to the context, and the writes go through `session.store`.

### M-8 · minor · T-4 compared the builder with itself
**Answer:** Fixed. The flow is tested once in `UnpromoteFlowTests`.

## Round 1 — executability

### X-1 · blocker · The not-found error has kind `.user`, not `.notFound`
**Answer:** Fixed by removal. D-6 reads the effects, and no error kind is matched.

### X-2 · blocker · Step checks named tests that do not prove the step
**Answer:** Fixed. Each step's check now names its tests. The view's leave and focus rule is a static, testable `outcome(effects:entityId:)` (T-8).

### X-3 · blocker · The message had no testable function
**Answer:** Fixed. `UnpromoteConfirmItem.message` (D-2) is covered by T-2.

### X-4 · blocker · The three reads were unspecified
**Answer:** Fixed. D-4 names the keys and the nil rules.

### X-5 · blocker · `handleID`, the gone check and the `listed` cap
**Answer:** Fixed. D-2 defines `handleID` and `deletesOwnHandle` (`total == 1`, so the cap cannot hide it). The gone check reads the effects (D-6).

### X-6 · blocker · Leaving can race the not-found Callout
**Answer:** Fixed. D-7 leaves before the reload paints, and manual item 3 checks it.

### X-7 · major · Wiring of `canBegin`, `userID`, navigation and `PVCopy`
**Answer:** Fixed in D-8 and D-9.

### X-8 · major · Focus was underspecified
**Answer:** Fixed. D-10 adds a flattened reading order, applied after the reload. T-8 covers it.

### X-9 · major · Current state rows
**Answer:** Fixed. The error-kind row is corrected, the test row names its tests, and the precedent cites `DeleteImpactFlow` and `PromoteModel.graphSnapshot()` (both verified).

### X-10 · major · `isUnpromoting` after the lift
**Answer:** Fixed. The flow owns `isRunning` and `isOpening` (D-1).

### X-11 · minor · Step order and size
**Answer:** Fixed. Each key is added in the step that first uses it, and the strings step only checks the catalog. The steps are re-cut to seven.

### X-12 · minor · `add-workspace-location` is listed but unused
**Answer:** Fixed. It is dropped from the skills to load.

## Round 2 — the agreed board, 2026-11-05

A fresh reviewer checked the refresh against S13-16's types and the board. **Verdict:** Ready with changes. Every finding is Fixed. The plan stays at `Blocked — on S13-37, S13-16`: it is refreshed and set to `Planned` once both have landed.

- **R-1 · blocker · D-9 reads fields `Line` does not have.** Fixed. S13-16's `Line` gains `graphID`, and D-9 reads `line.graphID`, `line.subjectID` and `line.subjectName`.
- **R-2 · major · The fallback name contradicts S13-16 D-2.** Fixed. The confirm title is `Line.subjectName` (D-4, D-9).
- **R-3 · major · Stale card wording.** Fixed. The plan now says "record line", and STEP-5 updates the `evidenceGraph.subject.unpromote` key's comment.
- **R-4 · major · The list has no input for the control or its focus.** Fixed. `PersonSourcesList` gets `unpromote:`, `unpromoteDisabled:` and a `FocusState<PersonSourcesFocus?>` binding (D-10).
- **R-5 · major · T-8's signature cannot compute the focus.** Fixed. `outcome(effects:entityId:removed:in:)` takes the removed claim and the content, and gains two more cases.
- **R-6 · major · The brief's inventory still names the IconButton.** Fixed. The brief's RC-1 and its Remove inventory row now give the agreed ghost Button with the glyph and the word.

## Refresh round 1 — S13-37 and S13-16 landed, 2026-11-06

The plan was refreshed against `main` @ `f48846c3`: every Current state row re-read, and the Decisions rechecked. The refresh itself added D-11 (section switches push history entries, so RC-7's "Back does not return" needed more than D-7) and the failed-preflight rule in D-3.

**Verdicts:** maintainability (Opus) Fits with changes; executability (Sonnet) Not ready.

### To user

- **M-3 · Where the shared Unpromote confirm lives.** **Answered 2026-11-06: split into a recipe.** D-2 adds `Recipes/UnpromoteConfirm/PVUnpromoteConfirm.swift` with neutral props, and the flow sits in `Features/Catalog/Unpromote/`. No rule is bent.

### Maintainability

- **M-1 · blocker · Fourth hand-rolled code match, no parity.** Fixed. D-11 adds `CoreInvokeError.code` and `ConclusionDetailsError`; the three code-only matches move on; T-10 is the parity row. `apperr.KindNotFound` is rejected in D-11: an engine change this PR does not need.
- **M-2 · blocker · The session reloads a deleted handle's keys.** Fixed. D-7 leaves before `apply` through `confirm`'s `beforeApply`, so `.evict` drops the keys and nothing reloads. `onLocationCommit` runs `session.apply(location:)` in the same call. T-5 proves it.
- **M-4 · major · Third hand-written leave.** Fixed. D-11 adds `leavesForSectionRoot(when:)`; the graph and the Source page move on; the `.failed` branch draws the loading page.
- **M-5 · major · `open`'s closure and `orders:`.** Fixed. D-1 takes `graphID:`; the flow reads the snapshot and `session.nameFormat`. D-4's `query` covers a stale key (T-7).
- **M-6 · major · `UnpromoteResult`, `leaves(...)`.** Fixed. `confirm` returns `CatalogWriteEffects?`; both wrappers go.
- **M-7 · major · Focus after a sheet closes.** Fixed. D-10 uses `FocusClaim.claim`.
- **M-8 · minor · Focus state and the walk.** Fixed. The list keeps its focus; the walk is `PersonSourcesContent.lineAfterRemoving(claimID:)`.
- **M-9 · minor · `isOpening || isRunning`.** Fixed. D-1 has `isBusy` over a private open count.
- **M-10 · minor · `userID` on the shared context.** Fixed. D-8 passes it through `PersonDetailSections`, which supersedes round 1 M-7.
- **M-11 · minor · Next adds and skills.** Fixed. The S13-28 clause goes; `add-workspace-location` is back.

### Executability

- **X-1 · blocker · No route for `userID`.** Fixed by M-10: the route no longer goes through the private page body.
- **X-2 · blocker · The focus hook cannot see its state.** Fixed by M-8: `pendingFocus`, `.onChange` and the confirm are all on `PersonSourcesList`.
- **X-3 · major · Fixture and write path.** Fixed. STEP-1 names the session, seeds, `nextWriteEffects`, the two `withState` blocks and the `graphRows` move.
- **X-4 · major · L10n shapes and comments.** Fixed. Strings gives keys, types, values and comments.
- **X-5 · major · Focus signatures.** Fixed. D-10 opens with them. A removed line is never the only visible one of a folded Source: a fold shows five.
- **X-6 · major · Undecided edges.** Fixed. A Reach row accepts the unmount, and the walk's fold case is in T-8.
- **X-7 · minor · `readyValue` and Promote.** Fixed. D-4 says Promote keeps its `reread` and its keys.
- **X-8 · minor · Stale doc.** Fixed. STEP-8 and the DoD name `docs/macos-client-patterns.md` and `CatalogWriteEffects.deleted`'s comment.
- **X-9 · minor · Smaller choices.** Fixed for the preview (D-10). The AC-2.6 transition stays a manual check: T-9 tests the predicate, and the modifier is two lines.

## Refresh round 2 — 2026-11-06

**Verdicts:** maintainability (Opus) Fits with changes; executability (Sonnet) Ready with changes. The refresh-round-1 answers held, except that two left new blockers. Every finding is Fixed, and the two-round cap is reached.

- **M-1 · blocker · D-4's warming read turns two Promote tests red.** Fixed. D-4 says Promote's read now warms, a real visit is unchanged because the place warms the keys, and T-11 rewrites both tests onto a failed rows load. The Stop clause goes.
- **M-2 / X-1 · blocker · Steps ran before what they depend on.** Fixed. The order is now snapshot read, flow with D-3, then recipe and message; T-1's `item.message` assertion moves to STEP-3.
- **X-2 · major · The Source page's `model.sync`.** Fixed. D-11 keeps it in its own `.onChange`.
- **X-3 · major · STEP-6's `rg` output.** Fixed. D-11 gives the accessor's form, and the check prints that one line.
- **X-4 · major · `beforeApply` isolation.** Fixed. D-1 types it `@MainActor`.
- **X-5 · minor · Row and line inputs.** Fixed in D-10; D-9 adds `.fixedSize()`.
- **X-6 · minor · Small gaps.** Fixed. The leave modifier's form, "visible line", STEP-2's files and T-5's call are named.
