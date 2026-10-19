# Review — S11-15 Shortlist index

**Document:** [`S11-15-shortlist-index.md`](S11-15-shortlist-index.md) (Draft, Blocked on S11-13)
**Round 1:** 2026-10-19 — maintainability (fresh context): Fits with changes; executability (fresh context, dry run): Not ready
**Round 2:** due at the plan's STEP-0, once S11-13 has landed and the *(S11-13)* rows are redone against code
**Verdict so far:** every round-1 finding answered; no blocker open; no To user item

## Maintainability

### M-1 · blocker · `Dated.Dates` is a third hand-written date list and skips a period's middle
**Answer:** Fixed. D-5 exports `subjectcompare.Side`; D-6 makes `Dated{Type, Date}`; D-7 reads every value through `Side`, so a period is one range and a pooled property counts with its lead. Inventory row `Dated.event.Dates` → `Dated.event.Date` (D-11). AC-1.8 is now 1840–1870 against 1855 plus an open period; AC-1.9 checks a `date` hides the period.

### M-2 · blocker · The date Near reach is recomputed outside layer 1
**Answer:** Fixed. D-4 adds `valuecompare.YearReach` on `readDate` in `date.go`; T-4's guarantee moved into `core/valuecompare` (`TestYearReachCoversEveryRelatedPair`). `core/shortlist` only buckets (D-7).

### M-3 · major · `matchregistry.Alignment` is a section type declared in the registry
**Answer:** Fixed. D-10's doc on `Alignment` says it is temporary and why; **Next adds** names S11-16's move; D-15 rewrites S11-16's scope entry ("grows the section S11-15 opened, and moves its type into `graphalign`").

### M-4 · major · Two staleness mechanisms, one guarding a write that does not exist
**Answer:** Fixed. D-12 keeps drop-on-change only (`SetAliasTables` drops the slot); `namePattern` is gone from `matchIndex`; the `nameformats` package doc gains the line for a future writer. Out lists that writer.

### M-5 · major · `WordKeys` adds a second per-type switch beside `comparers`
**Answer:** Fixed. D-2: optional `wordKeyer` on `name` and `text`, dispatched through `comparers`; D-15 adds a step to the skill's "Add a comparer".

### M-6 · major · The skill update covers only the Check command
**Answer:** Fixed. D-15: package list, a "Shortlist keys come from layer 1" rule naming T-7, an "Add a shortlist key" section, and the "Add a comparer" step.

### M-7 · minor · Codebase fit misses two existing maps
**Answer:** Fixed. §4 rows "Kind index" and "Search"; the `core/shortlist` Codebase fit row names search and why it stays apart.

### M-8 · minor · Every handle commit rebuilds the whole index
**Answer:** Declined for this PR, logged. The design says "dropped with a handle commit" ([layer 6 procedure](../comparison-layers.md#layer-6-procedure) step 1); **Next adds** records the additive path (dirty ids, `Remove` and re-`Add`) for S11-21; Out names it.

### M-9 · minor · Three names for one slot
**Answer:** Fixed. One name: `matchIndex` (`promotealign`) / `MatchIndex` (`graphcache`); `cachedIndex` / `storeIndex` became one reader, `matchIndexOf` (D-9).

### M-10 · minor · `Keys` copies layer 2's type-from-meta step
**Answer:** Fixed. D-5: `propertycompare.Typed`, used by `split` and `Keys`.

### M-11 · nit · A second `floorDiv`
**Answer:** Logged. `core/database/connect/connect.go`'s is unreachable from a pure package; Out records it.

## Executability

### X-1 · blocker · D-5 does not key a period's interior
**Answer:** Fixed with M-1 (D-5, D-7, AC-1.8, AC-1.9; T-4 has a 1840–1870 range).

### X-2 · major · D-9 does not say how the slot behaves in either S11-14 order, or how two loaders share it
**Answer:** Fixed. STEP-0 records the order with an `rg` check; D-9 gives each order's shape; fields are pointers so nil means not built; `matchIndexOf` plus a read-modify-write store; the name-pattern rebuild is gone (M-4), so nothing rebuilds `stats` by accident.

### X-3 · major · `datevaluetest` cannot build BEF / AFT
**Answer:** Fixed. §2 Conventions defines `bef` / `aft` as test-local helpers shaped like `datevaluetest.About`.

### X-4 · major · D-2 / D-3 signature details left open
**Answer:** Fixed. D-2: a word with no `AliasToken` adds no alias key; aliases read from `r.Names.Aliases`; text adds no alias key. §2 Conventions spells the `bill` table and how it reaches the rules.

### X-5 · minor · `loadShortlist` repeats what the profiler holds
**Answer:** Fixed. D-13: `loadShortlist(g, p, cfg, value)`; metas and types come from the profiler only.

### X-6 · minor · Warm-path cost not covered
**Answer:** Fixed. D-13: one lookup pass, one `g.Nodes`, one `profile` for all Subjects; T-10 `TestProposeQueryCountStaysFlatWhenWarm`.

### X-7 · minor · Smaller open choices
**Answer:** Fixed. D-11 gives each inventory row's cells; D-15 anchors the `promote-graph-alignment.md` edit by its sentence; T-7 says "the kind's `Dated.Date` lead"; STEP-6's check names the expected `g.Kind(` hit; T-7 lives in `core/graphalign/candidates_test.go`, listed in STEP-4. The seeding `Handle` fields are named in D-13.

**Not checkable now (from the dry run):** the S11-14 plan does not exist yet, and the base comparison waits for S11-13's code. Both are STEP-0's job.
