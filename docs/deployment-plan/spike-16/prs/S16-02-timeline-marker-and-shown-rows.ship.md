# Ship log — S16-02 The Timeline places a missing death and builds what it shows

**Plan:** [`S16-02-timeline-marker-and-shown-rows.md`](S16-02-timeline-marker-and-shown-rows.md) · **Branch:** `s16-02-timeline-marker-and-shown-rows` · **PR:** [#764](https://github.com/mendahu/provenencia/pull/764)
**State:** implementing (amendment)

## Round 0 — implement
Head `59e4a6a1`. Gate: implementer PASS; orchestrator re-run on `59e4a6a1` PASS (9/9; xcode-test 130s).
Report deviations:
1. STEP-1 and STEP-3 tests not seen failing first; code and tests in the same commits.
2. 7 commits, not strictly one per behavior: the Swift stub placement, shown-rows build, shared click target and kept scale share files (`5377df6c`).
3. Version 0.5.1 (main was already 0.5.0).
4. T-6 same-year case uses day dates: two year-wide spans put the baptism at the death stub's y.
5. T-7 `foldedMark(nearest:)` compares `pointY` (points of one year share a y; earlier wins the tie).
6. `KeptScale` keeps its four fields private; init and `pointsPerYear(for:viewportHeight:)` match the plan.
Noted: an earlier implementer gate run failed `go-test` on `core/database/promotealign` `TestLargeSourceBudgets/warm_at_open` (2.95s vs 2.5s) while overlapping an `xcodebuild` run; untouched package, re-run passed. Landing docs (map row, `completed.md`, plan Landed note) already committed in `59e4a6a1`.
Not done in round 0: manual check §10 (the user's).

## Round 1 — review
Reviewers: CON, MNT, UX (`macos/App/Features/Conclusions/**`, `macos/App/Platform/**`), SEC+PRF (`api/ffi/**`, `core/database/persondetail`, Timeline content model building rows for up to 2,000 events). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | Folded points after a birth placed past the fold are drawn inside its stub | Rejected (record the edge) → F-3 | D-9 as planned: "the `.more(.dated)` item stands for every folded point at its place". Unreachable from Go: `lifeline.beforeBirth` is empty (`lifeline.go`), so a placeholder birth is always `.before` the first dated row, which a fold (a suffix, S16-07 T-4) never hides; adding a before-birth type is Out "not planned". F-3 records the edge beside "Next adds" so whoever adds one sees it |
| MNT-1 | MNT | minor | "A recorded marker has its own span" check written three times in `PersonTimelineContent.init` | Fix now → F-1 | verified: `resolved(_:rowID:)` and the birth and death insertion guards each re-derive it |
| MNT-2 | MNT | minor | New test helpers written twice (Swift `timeline(_:birth:death:)`; Go `before`/`after`/`same(Place)`) | Fix now → F-2 | same-pass cleanup: this PR adds both copies |
| UX | UX | — | — | — | Clean |
| SEC+PRF | SEC/PRF | — | — | — | Clean; place ids checked against dated rows; build cost now bounded by shown rows at 2,000 events |

Nits (logged): `MNT-3 · lifeline.go:52 · matches is a method with an unused receiver; D-3 names a plain function` · `MNT-4 · PersonTimelineLifeline.swift · where spansByID[row] != nil then spansByID[row]!; bind with if let` · `MNT-5 · PersonTimelineLifeline.swift · Item.marker(marker).id built only to read the id; a Marker.id would do` · `CON-2 · 4d61ddb6 · app does not build at that commit (mapLifeMarker reads besideEventID until 5377df6c)` · `CON-3 · 5377df6c · three behaviours in one commit`

### Fix plan 1

**F-1 — Work out a recorded marker's own span once** (from MNT-1)
Files: `macos/App/Features/Conclusions/PersonTimelineContent.swift`.
Failing test first: none (refactor; behavior pinned by `PersonTimelineContentTests` and `PersonTimelineLifelineTests`).
Do: next to `birthRowID` and `deathRowID`, compute `birthHasSpan` and `deathHasSpan` (`rowID.flatMap { events[$0]?.span } != nil`) once; use them in `resolved(_:rowID:)` (pass the flag instead of re-reading `events`) and in both insertion guards. No behavior change.
Check: `rg -n 'events\[\$0\]\?\.span|events\[rowID\]\?\.span' macos/App/Features/Conclusions/PersonTimelineContent.swift` prints only the two `…HasSpan` lines; scripts/check-pr.sh PASS.

**F-2 — One copy of each new test helper** (from MNT-2)
Files: `macos/App/Features/Conclusions/PersonTimelinePreviewData.swift` (or a shared test extension under `macos/ProvenenciaTests/`), `macos/ProvenenciaTests/PersonTimelineContentTests.swift`, `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift`, `core/database/persondetail/lifeline_test.go`, `core/database/persondetail/timeline_test.go`, new `core/database/persondetail/export_test.go`.
Failing test first: none (test refactor).
Do: Swift — move the local `timeline(_ events:birth:death:) -> CatalogPersonTimeline` into one shared place both test files reach (prefer a `PersonTimelinePreviewData` extension in `macos/ProvenenciaTests/`, so no test-only code ships in the app); delete both local copies. Go — keep `before`, `after` and the place comparator in the internal `lifeline_test.go` (package `persondetail`) and expose them to the external `timeline_test.go` through `export_test.go` (`var SamePlace = same` etc., the standard export_test pattern); delete the external copies (`before`, `after`, `samePlace`). No exported production API.
Check: `rg -n 'func timeline\(_ events' macos/ProvenenciaTests macos/App` prints one line; `rg -n '^func (before|after|same|samePlace)\(' core/database/persondetail` prints only `lifeline_test.go` lines; scripts/check-pr.sh PASS.

**F-3 — Record the before-birth folded-point edge** (from CON-1)
Files: `docs/deployment-plan/spike-16/prs/S16-02-timeline-marker-and-shown-rows.md` (Landed note only; do not edit Decisions, Scope or ACs).
Do: add one sentence to the Landed note: adding a before-birth type (one `EventRole` in `lifelineTypes.beforeBirth`) can place a birth after the `.more(.dated)` item, and then D-9 draws the folded points dated after it inside the birth stub; that add must also shift those folded points (CON-1, S16-02 ship log).
Check: scripts/check-pr.sh PASS (plan-docs).

## Round 1 — fix
Head `e1a8b2ec`. Gate PASS (orchestrator re-run; xcode-test 112s). F-1 done (`7a3fa2a5`); F-2 done (`1f9445e5` + call-site fix `e1a8b2ec`: `1f9445e5` alone does not build the tests); F-3 done (`d11e4fa0`).
Deviations: F-2 Go helpers were closures inside `TestLifeMarkers`, so they became package-level `markerBefore`/`markerAfter`/`sameMarkerPlace` in `lifeline_test.go`, exported through `export_test.go`; Swift helper is `PersonTimelinePreviewData.timeline` in `macos/ProvenenciaTests/PersonTimelinePreviewData+Tests.swift`.
Read by the orchestrator (minors only, no round 2): MNT-1 **Resolved** (the Check `rg` prints only the two `…HasSpan` lines); MNT-2 **Resolved** (one Swift `timeline(_:birth:death:)`; Go comparator and builders only in `lifeline_test.go`); CON-1 edge recorded in the plan's Landed note.

## Round 2 — review
Not run: no round-1 blocker or major.

## Result
Gate PASS · CI green (Go tests, macOS tests, localizable) · open: none · follow-ups: none · to user: —

## Amendment — scope added after ready (user, 2026-11-09)
Dogfood: burial "abt 2011", death "29 Mar 2011", funeral "aft 29 Mar 2011" listed Burial, Death, Funeral. User chose to add the fix to S16-02 (Timeline only; funeral before burial when dates overlap). PR returned to draft. Plan amended through plan-pr and review-plan (`bbdf916b`, `b81958e8`): D-16 – D-22, STEP-8 – STEP-12, T-10 – T-14; Status Planned. User's rule on the planner's five calls: "the dates that are on the events are authoritative, but if they don't conflict, then we follow the rule of death, funeral, burial … order them by the dates, and then, if we need a tiebreaker, we use the types." That matches D-16 – D-20, so no call changed.

## Round A0 — implement amendment
Head `b3b49581`. Gate: implementer PASS; orchestrator re-run on `b3b49581` PASS (9/9; xcode-test 120s). Failing first seen for T-11, T-10 (7 cases), T-14; T-12 passed when written (STEP-9 already in). Deviations: `lifelineRow` gains a private `after` rank field (read once); PR body not edited (orchestrator's). Implementer saw one load flake (`promotealign` budget, four `WorkspaceDestinationHostTests`); passed on rerun, unchanged code.

## Round A1 — review (scope `f9ce9203...b3b49581`)
Reviewers: CON, MNT, UX (`macos/App/Features/Conclusions/**`), SEC+PRF (`core/database/**`; `order` loops over a person's rows, up to 2,000). Skipped: none.
