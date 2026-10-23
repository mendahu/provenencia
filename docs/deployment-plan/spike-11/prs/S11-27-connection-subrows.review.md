# Review — S11-27 Connection subrows on Promote

**Verdict:** Fits with changes — Ready with changes (both answered; no blocker open)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

Round 2 was not run: the round-1 blockers (M-1, X-1) were fixed by a narrower change that the reviewers proposed, and no Reshape was made.

## To user
- None. The writer asked the user four questions before writing (2026-10-23): build on S11-D1 as agreed; word a subrow with the existing connection sentence rather than the board's "A · term · B"; adopt the board's footer with an unselected part; count connections that file even when no subrow shows them. All four took the writer's recommendation.

## Writer's own calls
- D-2 hosting: a place relationship is hosted by its from end for both seeded types (`part_of`, `succeeded_by`). The brief names only `part_of`.
- D-1: when both ends wait, the state names end A.
- D-5: with a whole sentence, the other end's status ("Birth is already filed") is a second line, not a suffix after the name.
- D-9: the footer parts are joined with `" · "` in Swift, not written as four whole-sentence keys.
- D-10: expansion and highlight are view `@State`, not flow state.
- D-12: a pointer expands every host whose list hides one of the pointed-at connections.

## Round 1

### M-1 · blocker · D-1 works out "why an end waits" beside `resolve`
**Answer:** Fixed. D-1 now makes `resolve(_:)` return `EndResolution { handle(id:ref:), waits(WaitReason) }`, and it carries the self-link ref. No other code reads targets to decide a state.

### M-2 · major · `isSymmetric` adds a reader of the Swift inverse map that `ideas/term-management.md` retires
**Answer:** Fixed. D-3 adds `SeededRelationshipTypeTerm.isSymmetric(_:)` beside the map, and `EvidenceBridgeEdgeSummary.isSymmetric` calls it. STEP-6 adds a line to the idea file's item 2, and the DoD has a box for it. STEP-2 checks `rg "\.inverse\b"`.

### M-3 · major · T-7 copies the `viaReason` helper
**Answer:** Fixed. The symmetry cases move to `EvidenceBridgeEdgeSummaryTests`, using that file's helpers. `viaReason`'s setup is split into `bridgeModel(kind:edgeKeys:term:)`, which both tests use (STEP-2, T-7).

### M-4 · major · The `phrase` chain becomes dead code
**Answer:** Fixed in part. `BridgeFact.phrase` and `Bridge.phrase`, and the `bridgeFacts` argument, are deleted (D-4, Deletes, STEP-5's `rg`). Declined for `EvidenceBridgeEdgeSummary.phrase(for:)`: `wholeNameFallback(for:)` in `EvidenceBridgeEdgeSummary.swift` (line 154) still calls it, so it is not dead.

### M-5 · major · End names are looked up in three places
**Answer:** Fixed. `Bridge.name(of:)` is the one lookup (D-5), and `ConnectionState.waitingEndID` feeds it. The model function `connectionEndName` was dropped.

### M-6 · minor · `unsetCount` repeats the `targetsChosen` predicate
**Answer:** Fixed. D-8 now defines `targetsChosen` as `unsetCount == 0`.

### M-7 · minor · Scroll and timer are hand-rolled
**Answer:** Fixed. D-12 uses `PVMotion.fastStandard` as `PVTable` does, and `.task(id: highlightGeneration)` handles cancellation.

### M-8 · minor · The spoken keys repeat the sentence slot; two functions share a name
**Answer:** Fixed. The spoken keys hold only the state sentence, and `connectionSpoken` joins two whole sentences. The L10n functions are `connectionState(_:endName:)` and `connectionSpoken(sentence:state:endName:)`, and no model method shares their names.

### M-9 · minor · The footer adds a fourth hand-written `" · "` join
**Answer:** Declined for this PR. A localizable separator helper would change three existing callers outside this PR's files (the composer's two `menuLocator` copies and `NameValueDraft`), which the Implementer block forbids. Logged here as a follow-up for an L10n tidy-up, together with the duplicated `menuLocator`.

### X-1 · blocker · STEP-6 moves S11-D1 without listing the links to rewrite
**Answer:** Fixed. STEP-6 names every file that links the brief: this plan, S11-18, S11-20, the S11-D2 archive, `deployment-plan.md` and `design/README.md`. It also says to fix the brief's own relative links and gives a precise `rg` check. Running `rg` at the base showed that `docs/versioning.md`, `docs/promote-graph-alignment.md` and `promote-connection-subrows.md` do not link the brief.

### X-2 · major · STEP-3 and STEP-5 leave `doneSummary` and the old keys in an undefined state
**Answer:** Fixed. The new `doneSummary` and the `promote.done.*` keys come in STEP-5, which also deletes the old ones. STEP-3 deletes the three state keys along with the old `connectionState(_:)`.

### X-3 · major · L10n and model signatures elided or clashing
**Answer:** Fixed. §6 Strings gives the full signatures on `L10n.Promote`. The model function is gone (M-5). STEP-3 gives the interim `connectionRow` code.

### X-4 · major · T-6 fails without `mappedRest`
**Answer:** Fixed. T-6 sets `entryID: "a"` and `mappedRest = true`.

### X-5 · minor · Test gaps
**Answer:** Fixed. T-5 gives every handle. T-1 fixes the bridge ids and every pointer. T-3 drafts the unset row without the contradictory `target: "skip"`. T-7 uses the split helper. §8 says how to set `kind` and `symmetric`.

### X-6 · minor · D-1 and D-12 holes
**Answer:** Fixed. D-1 states that an anchor always resolves. D-12 expands every host that hides a pointed-at connection, uses `PromoteLayout.subrowCap`, and computes page order from `sections` in the view.

### X-7 · minor · D-4 keeps dead `phrase`
**Answer:** Fixed with M-4.

### X-8 · minor · README "(#N)" before the PR exists
**Answer:** Fixed. STEP-6 says to open the draft PR before the step and fill in N.

Executability limits it reported (not findings): Xcode checks need macOS, the board link is not reachable (§6 describes the frames in words), and the view has no automated test (§10 covers it).
