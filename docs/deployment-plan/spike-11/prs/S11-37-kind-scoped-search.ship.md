# Ship log — S11-37 One kind-scoped catalog search

**Plan:** [`S11-37-kind-scoped-search.md`](S11-37-kind-scoped-search.md) · **Branch:** `s11-37-kind-scoped-search` · **PR:** #551
**State:** ready

## Round 0 — implement
Base `main` @ `55611a1c` (plan written against `7f8117d`). Worktree: `/Users/jakerobins/Repos/provenencia-s11-37`.
Head `17a52a87`. Gate PASS (orchestrator run; xcode-test 51s). Report deviations:
- `waitForSearch` (`QueryHandleWaiting.swift`) waits 10 s, not the 2 s default. The implementer's first gate run timed out 12 search expectations under full-suite load. The old private helper counted iterations.
- Steps 2–4 landed as two commits. The pbxproj registers the new files together.
- A FakeStore doc comment now reads "vocabulary fake" after the D-3 rename. This is outside the plan's file list.
- The base moved from `7f8117d` to `55611a1c`. The Current state rows were checked against it and none are affected.
- AC-1.2 (manual check) was not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Nothing skipped. UX runs because the diff touches `macos/App/**`. SEC+PRF runs because the diff touches a performance path (`macos/App/Features/**` models: `CatalogSearchModel`, `OmnibarResultsModel`).
Verdicts: CON clean, MNT clean, UX clean, SEC+PRF clean (no findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 (MNT-1, MNT-2) | CON | minor | Stale "omnibar" local and comment in `FakeStore` after the D-3 rename | Fix now → F-1 | Minor in a file this PR changes. Confirmed with `rg -n omnibar macos/App/Platform/FakeStore.swift` at lines 106 and 1233 |
| UX-1 | UX | minor | The shared hit row leaves the VoiceOver label to each caller | Later PR → S11-36 | D-8 keeps the label at the call site. That is right for this PR: the omnibar's announcement is unchanged. The risk is S11-36 forgetting it, so S11-36's map scope entry now requires the label (or moving it onto the row). Written in `deployment-plan.md` § S11-36 |
| CON-2 | CON | nit | Review link points to a file not in the diff | Rejected | `S11-37-kind-scoped-search.review.md` already exists on `main` |

Nits (logged): `MNT-3 · CatalogSearchHitRow.swift:6 · selected is always false at its only caller` (also noted in S11-36's scope entry for its planner)

### Fix plan 1
**F-1 — Finish the D-3 rename in FakeStore** (from CON-1, MNT-1, MNT-2)
Files: `macos/App/Platform/FakeStore.swift`.
Failing test first: none. This is a rename and a comment, with no behavior change.
Do: rename the local `let omnibar = try await vocabularySearchHits(...)` (around line 1233) and its uses to `vocabulary`. Reword the `searchCatalog` error-hook comment at line 106 so it does not say "omnibar error UI". The hook now serves any search caller's error UI.
Check: `rg -n omnibar macos/App/Platform/FakeStore.swift` finds no reference to the local or the error-hook comment. scripts/check-pr.sh PASS.

## Round 1 — fix
Head `554e6730`. Gate PASS (orchestrator run; xcode-test 45s). F-1 done; read by the orchestrator. **Resolved**: `rg -n omnibar FakeStore.swift` finds nothing.

## Round 2 — review
Not run. Round 1 had no Fix-now blocker or major, so no reviewer goes back.

## Result
Gate PASS · CI pending · open: none · later PR: UX-1 and MNT-3 in S11-36's scope entry · follow-ups: none · to user: none · CON miss: none · manual check (§10) not yet run
