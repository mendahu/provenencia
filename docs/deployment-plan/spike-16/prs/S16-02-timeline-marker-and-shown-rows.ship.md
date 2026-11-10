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

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Pass (3) moves a funeral above rows the dates clearly put before it | To user → user chose "move the burial down" (2026-11-09) → Fix now → F-1 | Reviewer's seeded probe: 830 non-overlapping pairs flipped, 0 with funerals removed; failing case `[4 2 3]` → `[3 4 2]`. Contradicts US-4 and D-21/D-22 wording; fix changes D-16's mechanism, so asked |
| PRF-1 (SEC) | PRF/SEC | major | Pass (3) quadratic in a person's own burials/funerals, uncapped, inside the held session | Fix now → F-2 | Reviewer's overlay probe: 147 ms at 2,000 rows with nothing moving, ~2 s at 16,000; profile 97% in `order` (`inSet` map lookups, `find` scans, shifts). Map rule "Crafted fan-out is capped, not engineered"; the plan's Scale paragraph counted `Overlaps` calls only |
| UX-1 | UX | major | A moved row's bar button covers the death or birth point; clicking it opens the moved row | Fix now → F-3 | verified: `PersonTimelineLifeline.swift` appends `targets` in item order; `PersonTimelineSection.swift` draws them in a `ZStack` `ForEach` (later on top); `Mark.hit` of a bar spans its extent ±6. Dogfood order D, F, U puts the burial's year bar over both points |
| MNT-3 | MNT | minor | `EventSpanAndAge` returns five positional results | To user → user chose a named struct (2026-11-09) → Fix now → F-4 | D-17 named the signature |
| MNT-1 | MNT | minor | `order` moves rows two ways (pass 3's inline shift and `inSet`) | Fix now → folded into F-1/F-2 | pass 3 is rewritten there |
| MNT-2 | MNT | minor | `TestLifelineOrder` copies `TestLifeMarkers`' fixture | Fix now → F-5 | same-pass cleanup; this PR adds both |
| CON-2 | CON | minor | An `.after` death stub can sit above a recorded birth pass (1) put before its row | Fix now → F-6 | code this PR changes; mirror of D-20; F-3 touches the file |
| UX note → CON | CON | minor | Plan still has the "Amendment —" §4 table row and Edges bullets STEP-12 said to delete | Fix now → F-7 | `rg -n 'Amendment —' S16-02…md` |

Nits (logged): `MNT-4 · lifeline.go · lifelineRow.after stores one rank, pass 1 recomputes the other` · `MNT-5 · PersonTimelineLifeline.swift:654 · nested firstIndex.map…compactMap; name a local` · `MNT-6 · GenealogyStore.swift:775, PersonTimelinePreviewData.swift:6, PersonTimelineContent.swift:7 · comments name "S16-02's moves"; name lifelineTypes.order` · `UX-2 · PersonTimelineContent.swift:7-8 · rejoin wrapped comment` · `CON-3 · PersonTimelinePreviewData.swift:6 · long doc line`

### Fix plan A1

**F-1 — Pass (3) moves the burial down, not the funeral up** (from CON-1, MNT-1; user, 2026-11-09)
Files: `core/database/persondetail/lifeline.go`, `lifeline_test.go`, the plan (Landed note: D-16 pass (3) superseded, with the user's answer).
Failing test first: `TestLifelineOrder` "a funeral does not pass a row the dates put before it": rows burial 4 `YMD(2011,0,0)`, census 2 `YMD(2011,2,0)`, funeral 3 `After(2011,3,1)`, no birth/death → want `[2, 3, 4]` (today `[3 4 2]`). Also a seeded property test over `order` with the shipped `lifeline` (fixed seed, ≥ 5,000 cases of 2–8 rows in 2010–2011 with year/month/day/AFT/BEF dates, optional dated birth and death, funerals and burials): every pair whose relative order changed has overlapping windows, and the output is a permutation of the input.
Do: among the after-death rows after the death (or all of them when there is no recorded dated death), a burial (higher `afterDeath` rank) that has a later overlapping funeral (lower rank) moves to just after that funeral, by the same move shape as `moveAfterOverlapping` (reuse it or a shared single-row helper beside it, so `order` moves rows one way). Drop `inSet` and the per-id `find`: track positions, test the int rank before anything else. Every T-10 and T-12 expectation holds unchanged.
Check: both tests pass; T-10, T-12 unchanged; scripts/check-pr.sh PASS.

**F-2 — Cap pass (3) and keep the no-move path linear** (from PRF-1)
Files: `core/database/persondetail/lifeline.go`, `lifeline_test.go`, the plan (Landed note: correct the Scale paragraph's count).
Failing test first: a timing test (pattern of the repo's existing budget tests, e.g. `core/database/promotealign` `TestLargeSourceBudgets`) that runs `lifeline.timeline` on 20 × the cap own burials and funerals with one window and a dated death, interleaved, and asserts it finishes within a budget well under today's (today ~0.15 s at 2,000; pick the budget so the current code fails at 20 × cap).
Do: a package constant `maxReorderedAfterDeath` (64) with a `// decided:` comment (a real person has one to three; only a crafted catalog passes it; above it pass (3) leaves those rows in date order). Count the after-death rows pass (3) would consider before doing pair work. Passes (1) and (2) stay one `moveAfterOverlapping` call each. No cap on the person's events (Out line stands).
Check: the timing test passes; scripts/check-pr.sh PASS.

**F-3 — The narrowest click target sits on top** (from UX-1)
Files: `macos/App/Features/Conclusions/PersonTimelineLifeline.swift`, `macos/ProvenenciaTests/PersonTimelineLifelineTests.swift`.
Failing test first: `PersonTimelineLifelineTests.aPointUnderAWiderBarIsClickable`: the dogfood fixture (death `.day(2011,3,29)` recorded, funeral AFT 29 Mar 2011, burial year 2011, in that item order) and a recorded birth `.day(1850,3,12)` followed by baptism year 1850: the last target in `targets` whose `hit` contains `mark("d")!.pointY` (and `mark("b")!.pointY`, and the funeral's `pointY`) is that mark's own target.
Do: build `targets` in drawing order: longer `hit` first, item order for ties (stable), so a later, narrower button is on top in `PersonTimelineLine`'s `ZStack`. Keep the click handler and labels unchanged.
Check: the test passes; existing lifeline tests unchanged; scripts/check-pr.sh PASS.

**F-4 — `EventSpanAndAge` returns a named struct** (from MNT-3; user, 2026-11-09)
Files: `core/database/conclusionsummaries/order.go`, `conclusionsummaries_test.go`, `core/database/persondetail/timeline.go`, the plan (Landed note: D-17 signature superseded).
Failing test first: none (refactor; `TestEventSpanDays` and the span/age tests pin behavior).
Do: `type EventPlacing struct { Span TimelineSpan; Days valuecompare.Window; Dated bool; Age AgeRange; Aged bool }` (name per the package's conventions); `EventSpanAndAge` returns it; update the one production caller and the test calls.
Check: `rg -n 'EventSpanAndAge\(' core api` shows no positional `_, _` destructuring; scripts/check-pr.sh PASS.

**F-5 — One `lifeline_test.go` fixture** (from MNT-2)
Files: `core/database/persondetail/lifeline_test.go`.
Do: move `id`, the type map with `types`, and a row builder (`lifelineTestRow(n byte, roles ...vocabkey.Role) PersonEvent`) to package-level test helpers beside `markerBefore`; use them in `TestLifeMarkers`, `TestLifelineOrder` and F-1/F-2's new tests.
Check: `rg -n 'typeOf := map\[byte\]' core/database/persondetail` prints at most one line; scripts/check-pr.sh PASS.

**F-6 — An `.after` stub sits below every shown dated item before it** (from CON-2)
Files: `macos/App/Features/Conclusions/PersonTimelineLifeline.swift`, `PersonTimelineLifelineTests.swift`, the plan (Landed note: D-20 mirrored).
Failing test first: `PersonTimelineLifelineTests.anAfterStubBelowALaterEndingRowBeforeIt`: `Data.timeline([birth "b" .day(1850,3,12), "x" .day(1850,1,1)], birth: CatalogLifeMarker(eventId: "b"), death: CatalogLifeMarker(place: .after(eventId: "x")))` → the death stub's lowerBound ≥ `mark("b")!.pointY`.
Do: mirror D-20: an `.after(row)` stub's `at` is the greatest `to` among that row's span and the spans of the shown dated items before the marker in `content.items`. Name the scan as a local (covers MNT-5).
Check: the test passes; T-6 unchanged; scripts/check-pr.sh PASS.

**F-7 — Remove the plan's leftover amendment notes** (from the UX note)
Files: the S16-02 plan.
Do: delete the "**Amendment — this branch @ `4588b62e`.**" table in §4 and the "Amendment —" Edges bullets (fold anything still true into the Landed note).
Check: `rg -n 'Amendment —' docs/deployment-plan/spike-16/prs/S16-02-timeline-marker-and-shown-rows.md` prints nothing; plan-docs PASS.
