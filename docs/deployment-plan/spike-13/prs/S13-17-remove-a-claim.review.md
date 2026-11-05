# Review — S13-17 Remove a claim

**Verdict:** Reshape — Not ready (round 1, on the draft). Reshaped with the user's answer of 2026-11-05; every finding Fixed.
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
