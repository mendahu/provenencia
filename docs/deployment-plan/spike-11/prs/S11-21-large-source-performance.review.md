# Review — S11-21 Large-Source performance check

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. Budgets and remedies were answered before review (user, 2026-10-24: "Snappy" budgets; bound + memos).

## Writer's own calls
- The benchmarks keep their catalog through `benchShape`, except that hub events past the first 20,000 become deaths (D-4). Their numbers are reported only.
- No name-format commit effect, although the scope entry names one: the only writer is `nameformats.Install` at onboarding, so the name format is warmed at open like the alias tables (D-9). The scope entry is rewritten in STEP-9.
- Warm Evidence keeps `SoleGraph`'s 2 SELECTs until S12-09 deletes it; the test measures `SoleGraph`'s count in place (AC-3.3).
- Not built, measured inside budget: in-place index update, key-value profiling, a bounded canon walk.
- Map: S11-21 depends on S11-19 as well as S11-16; S11-31 depends on S11-21 (it sizes its cap from this fixture).

## Round 1

### M-1 · blocker · The warm name format belongs to `nameformats`, not `valuerules`
`conclusionheaders` `persons` and `handlers.GetNameFormat` also read `nameformats.Load`; a warm slot read only inside `valuerules` makes two ways to read one fact.
**Answer:** Fixed — D-9 now adds `nameformats.Warm`, `Invalidate` and `LoadFor(g, q)`; all three readers move to `LoadFor`; T-9 tests `nameformats`. Checked: `graphcache` imports neither `nameformats` nor `core/database`, so there is no cycle (stop condition added).

### M-2 · major · Steps and comments still describe nominations as unbounded
`promote-graph-alignment.md` §4 steps 1 and 4, `place`'s doc comment, `Rules.TopK` / `TieCap` comments, the `matching-model.md` glossary row **candidate**.
**Answer:** Fixed — §6 Go core gives the three comments; STEP-9 gives the doc sentences; §4 Docs row lists them.

### M-3 · major · `qualifiedEvent` copies `datedOn`'s body
**Answer:** Fixed — `(*fixture).eventOn` takes `*datevalue.Value` (nil: undated); `datedOn` calls it (STEP-1, Codebase fit).

### M-4 · major · The candidates memo holds a whole index pointer
**Answer:** Fixed — D-8 uses a `matchIndex.gen` generation from `indexGen atomic.Uint64`, like `profilesPattern` checks profiles; `loadShortlist` returns it.

### M-5 · minor · The fixture copies the open-time warm list
**Answer:** Fixed — D-9a `catalogsession.WarmOnOpen`, called by `openResearcher` and the fixture.

### M-6 · minor · `NominatedForTest` repeats `Align`'s steps
**Answer:** Fixed — D-7a splits `Align` into `run` plus `proposal`; the seam calls `run`.

### M-7 · nit · Name-format session test in `alias_cache_test.go`
**Answer:** Fixed — `core/catalogsession/warm_test.go` *(new)*.

### X-1 · blocker · How `bulkCanon` writes a date is not decided
**Answer:** Fixed — D-4: one `auto_reconciler_values` row with `value_date = valuecodec.MarshalDate`, `date_lo` / `date_hi` / `sort_key` NULL (`graphcache` `loadValues` reads only `value_date`; the planning prototype did exactly this); `fixtureDate` maps BEF / AFT / none / full day; death term and toponym writes named.

### X-2 · major · Fixture inputs underspecified
**Answer:** Fixed — D-4a: the cumulative-weight sampler, `letters`, one PCG stream and its draw order; STEP-3: hub page is its own Source, its persons' names, participations and `locate`; Source A's events are not located; the 100-Subject count.

### X-3 · major · `CountSelects` can deadlock; nothing checks it early
**Answer:** Fixed — D-10 closes the `*sql.Conn` after `Raw` and clears the authorizer in `t.Cleanup`; STEP-2 adds `TestCountSelects` (T-10). `sqlite3.SQLITE_OK` is exported by mattn v1.14.50 (`sqlite3.go` line 357).

### X-4 · major · STEP-1 and STEP-6 hide many changes
**Answer:** Fixed — STEP-1 split into shape (STEP-1), counter (STEP-2), fixture (STEP-3); subtest `fixture` asserts 100 rows and ≥ 40 handle rows. Nine steps.

### X-5 · minor · T-7 reuses a local table
**Answer:** Fixed — STEP-4 hoists the table; the move is In.

### X-6 · minor · Nil graph
**Answer:** Fixed — D-9: `NameFormat` / `SetNameFormat` nil-receiver safe; `LoadFor(nil, q)` reads `q`; T-8 and T-9 cover both.

### X-7 · minor · Empty candidates entry; T-5 shape
**Answer:** Fixed — D-8: an empty non-nil slice is an entry. T-5 now has one layer birth, named ids and refs, and each case's dates. T-6 names `f.filedBirth` and `f.cite`.

### X-8 · minor · Cold runs under parallel `go test`
**Answer:** Fixed — D-1: cold and after-commit take the fastest of 2; the 4-core parallel assumption is stated.

### X-9 · minor · STEP-7 paths; scope entry and map disagree
**Answer:** Fixed — STEP-9 uses full paths and rewrites the scope entry's Ships line plus a Not needed line. Map: S11-21 depends on S11-16 and S11-19; S11-31 depends on S11-19 and S11-21. §4 Docs and Benchmarks rows corrected.

### X-10 · nit
**Answer:** Logged.
