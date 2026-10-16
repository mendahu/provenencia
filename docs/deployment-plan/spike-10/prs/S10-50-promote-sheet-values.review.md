# Review — S10-50 Promote sheet lines show the values they compared

**Verdict:** round 1: Fits with changes — Not ready
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
None.

## Writer's own calls
- Term labels reach every `ValueRef`, so `Rank` / ForSubject reasons name terms by label too (D-1).
- A no-exhibit row's `agree` lines are no longer drawn as checked, disabled pins, and their values join the drafted argument (D-7).
- The PR also moves `conclusiondetails.termLabels` and `handlers.loadTermInfo` onto the new `Vocabulary.TermsByID` (D-2), beyond the two files the S10-28 follow-up named.
- The plan is stacked on the S10-28 branch and stays Blocked until #411 merges.

## Round 1

### M-1 · blocker · Two mechanisms still produce `Exhibit.*Display`
`loadExhibitObservations` builds `exhibitObs.display` with its own chain beside `Value.Label()`, and leaves `TermLabel` unset although it selects `t.label`.
**Answer:** Fixed — D-3 fills `TermLabel` in `loadExhibitObservations`; D-4 deletes `exhibitObs.display` and its chain, and `pairExhibits` uses `Value.Label()`. The name-without-`Form` Out line is removed and becomes AC-1.5 / T-7. Objective "What stays" and Deletes corrected.

### M-2 · major · `termsByID` would be a third id → term map
`conclusiondetails.termLabels` and `handlers.loadTermInfo` build the same map.
**Answer:** Fixed — D-2: `graphcache.Vocabulary.TermsByID()`; `neutralTerms`, `termLabels`/`termName`, `loadTermInfo`/`termInfo` are deleted (STEP-2, T-4). Stop condition added for any other reader of those structs.

### M-3 · major · `graphalign.Comparison.Pinned` becomes dead
Its only production reader is the FFI branch D-7 deleted.
**Answer:** Fixed — D-8 deletes the field, its setter, and the two test assertions (STEP-6); the `Comparison` doc comment is reworded.

### M-4 · minor · The tie-break should not depend on display copy
**Answer:** Fixed — D-1 adds `sortKey()` (`Term` for terms); `order` uses it. T-2 tests `order` directly. The tie-break Risks row is removed.

### M-5 · minor · Keep the sheet assembly in `attachExhibits`
**Answer:** Fixed — D-6: the early return becomes a skip, and `attachExhibits` ends with the `withRowLines` loop over every row. `Propose` unchanged.

### M-6 · nit · `refLine` dates should use `cloneDate`
**Answer:** Fixed — D-6.

### X-1 · blocker · A Person has no date Property
**Answer:** Fixed — AC-2.3 and AC-2.5 use birth Event rows; T-9 and T-10 build them after `TestProposeOneHopDateExhibit`; T-8 and T-11 (Person) check name and sex only; Manual check step 3 uses an Event.

### X-2 · blocker · The stop condition forbade test updates the plan needs
**Answer:** Fixed — §4 lists the tests that read term keys or `Comparison.Pinned`; STEP-3, STEP-4, and STEP-6 name each update with its new value; the stop condition now covers only unnamed tests.

### X-3 · major · T-5 and T-7 fixtures under-specified; T-7 is in another package
**Answer:** Fixed — T-8 gives the exact `promotealign` fixture steps (`f.cite` for sex, the `UPDATE` through `f.c.DB()`, `Graph().Drop()`); T-11 does the same inside `withProjectCatalog` with `personOnNewSource`. The "provisional still yields comparisons" risk stays a stop condition.

### X-4 · major · New and Skip rows under "every row"
**Answer:** Fixed — verified in `graphalign/align.go` `proposal`: only handle-target rows get `Comparisons`. D-6 and §4 say New and Skip rows get no lines, as today.

### X-5 · major · Handler shape left open
**Answer:** Fixed — D-8: unwrap to a plain `for _, ex := range r.Exhibits`.

### X-6 · minor · Column order in the SQL
**Answer:** Fixed — D-3: `COALESCE(t.label, '')` immediately after `COALESCE(t.neutral, 0)`, scanned into `termLabel string` after the neutral variable.

### X-7 · minor · Missing term id in `handleValues`
**Answer:** Fixed — D-3: zero values; `loadCanon`'s local renamed to `terms`.

### X-8 · minor · Doc targets
**Answer:** Fixed — STEP-7 names the "Reasons." bullet, the `completed.md` heading placement, and the map edits.

### X-9 · minor · Stale base statement
**Answer:** Fixed — the **Planned** line says the branch records S10-28 as landed ahead of the merge, and that `gh pr view 411` must report `MERGED` before work starts.

### X-10 · minor · T-2 helper signature
**Answer:** Fixed — T-3 (renumbered) calls `pairValues(ComparerFor(properties.ValueTypeDate), "", …)`.
