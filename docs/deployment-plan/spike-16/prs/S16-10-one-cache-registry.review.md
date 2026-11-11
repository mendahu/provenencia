# Review — S16-10 One cache registry

**Verdict:** Fits with changes — Ready with changes, 2 rounds; To user: answered (write path, typed get-or-fill)
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

- **M-1 · rule K-21 and a write's graph.** Answered 2026-11-11: allow catalog configuration only (option a) → D-21, Rules this plan bends, STEP-8.

## Writer's own calls

- D-9 / D-12: the drop-on-any-handle rules of `for`, `siblings` and the match index move into their owners' Apply now, so `DropOnHandleCommit` and `dropHandleMemos` go in this PR, not S16-16; the map's S16-16 and S16-20 wording follows (STEP-8).
- D-2: no keyed Fill callback; the owner's `SetMemo` plus Apply replace the design's "Fill plus Deps".
- D-11: `WarmNeeded`, `MarkWarm` and the re-warm after each `Do` go; the alias index and name format survive every in-app change already.
- D-15: Evidence-graph memos stop dropping on an alias index or name format change.
- D-6: keyed caps 65,536; past the cap a key is not stored.
- D-17: only persons grow for the 16,000 fixture.

## Round 1

### M-1 · blocker · A whole Build runs on the write path (`Over`), against working rule K-21
**Answer:** To user — answered 2026-11-11, option (a): D-21; the rule is reworded in STEP-8 and listed under Rules this plan bends.

### M-2 · major · A keyed entry with nil Apply keeps everything on a commit
**Answer:** Fixed — D-3: `RegisterKeyed` panics without an Apply; AC-1.8, T-6.

### M-3 · major · `PerKey bool` plus runtime panics where two slot types would do
**Answer:** Fixed — D-1, D-8: `KeyedEntry`/`KeySlot`/`RegisterKeyed` and `WholeEntry`/`WholeSlot`/`RegisterWhole`; kind-mismatch panics gone from AC-1.8.

### M-4 · major · `Reads` already means something else on the commit path
**Answer:** Fixed — D-4 `DerivedFrom`; whole accessors `Value`/`SetValue`/`DropValue` (D-8); STEP-8 rewords the rule, the design and the map.

### M-5 · major · The cache doc is held only by a manual `rg`
**Answer:** Fixed — D-19 `EntryNames` and `TestCacheDocListsEveryEntry` (T-16), after `TestKnobInventoryMatchesRegistry`.

### M-6 · major · Evidence-graph memos keep a hand-written vocabulary drop in `OnCommit`
**Answer:** Fixed (doc and map) — D-15: the loop is listed in `docs/catalog-caches.md` as the store's one hand-written drop, owned by S16-13; STEP-8 adds it to S16-13's scope (rule K-18 keeps per-graph memos in the store, which S16-13 reshapes). Out row added.

### M-7 · major · The Swift half of the doc hand-copies `descriptor`
**Answer:** Fixed — D-18 (d): one row pointing to `CatalogQueryKey.Kind.descriptor`, separate rows only for caches outside it.

### M-8 · major · `errNoGraph` beside `database.ErrClosed`
**Answer:** Fixed — D-10 returns `database.ErrClosed`; `errNoGraph` removed from Codebase fit.

### M-9 · minor · `valuerules.Querier` and `evidencelayer.Load`'s comment
**Answer:** Fixed — **Deletes**, STEP-4 files.

### M-10 · minor · Benchmark tied to `promotealign`'s private fixture
**Answer:** Declined — the fixture is built from `promotealign_test`'s `fixture` methods (`newFixture`, `promote`, `Place`, `filedBirth`, `propose` in `promotealign_test.go`), so moving it is a package split, not a file move; no PR planned so far measures outside `promotealign`. Out row: not planned until one does.

### M-11 · nit · `Cached` belongs in `graph.go`
**Answer:** Fixed — D-8, STEP-1.

### X-1 · blocker · Initialization cycle, and the "fresh profiler" shares the stored map
**Answer:** Fixed — D-12 registers in `init()`; D-13 adds `newProfilerWith(…, done)` and `buildShortlist(g, p, cfg)`, and the check reads `loadConfig(g)` and `valuerules.Metas(g)`.

### X-2 · major · Fixture at 16,000 persons not derivable
**Answer:** Fixed — D-17: `largeShape(…, persons int)` changes only `Persons`; `newLargeFixture(tb testing.TB, persons int)` (`newFixture` and `must` already take `testing.TB`); the benchmark reports `handles` and key counts; D-6's "37,000" claim replaced by a measured Stop condition (half the cap).

### X-3 · major · Benchmark fill steps open
**Answer:** Fixed — D-17 lists each call in order; kinship/for/siblings is one metric with three key counts.

### X-4 · major · T-6 needs a registry snapshot hook
**Answer:** Fixed — D-20 `SnapshotRegistryForTest` in `graphcache/export_test.go` *(new)*.

### X-5 · major · Does a Build's store drop readers?
**Answer:** Fixed — D-2: a first fill by Build drops no reader; `SetValue` does; AC-1.5, T-4.

### X-6 · major · Drops move after the reload; error path
**Answer:** Fixed — D-3: Applies run in registration order; the first error is `OnCommit`'s, and `Catalog.Notify` (`core/database/listen.go`) then calls `Drop()` on the graph, so no stale entry survives a failed commit.

### X-7 · minor · Small choices
**Answer:** Fixed — drops return no error (D-8); `Warm`/`Invalidate` signatures and nil-graph behavior (D-10); `SetMemo` nil and short-id rule (D-8); held-key replace under Cap (D-6); the extra call sites are in § 4; `TestDropLeaves*` cases go to T-1/T-9; `assocBetween` call sites (D-16); the Known trap is a new line (STEP-8, § 4 Guides row); the pointless line-241 note removed. The Swift enumeration is now one descriptor row plus the named outside caches (D-18).

## Round 2

### X-8 · blocker · STEP-3 deletes `warmNeeded` while `dropDerived` still sets it
**Answer:** Fixed — STEP-3 deletes the `g.warmNeeded = true` line from `dropDerived`; the function goes in STEP-4.

### X-9 · major · Match-index drop on a vocabulary commit between steps; D-13 errors
**Answer:** Fixed — STEP-1 keeps the `dropDerived()` call in the vocabulary branch until STEP-4; STEP-3 removes `TestMatchIndexDrops`' alias and name format cases and states the intended gap until STEP-4's DerivedFrom; D-13 returns `loadConfig`'s and `Metas`' errors.

### M-12 · blocker · D-21 guards Build but not `SetValue` on a write's graph
**Answer:** Fixed — D-21: `SetValue` of a `LifeGraph` whole entry on an `Over` graph stores nothing; AC-1.10, T-4; the reworded rule says "neither builds nor stores".

### M-13 · blocker · Key counts need exported slots
**Answer:** Fixed — `MemoCount(KeySlot)` replaced by `KeyCounts() map[string]int` by entry `Name` (D-8); the benchmark loops it (D-17).

### M-14 · blocker · D-2 departs from the registry rule; two map lines stay false
**Answer:** To user (plan § 12 Open question 1, recommendation: accept) for the Fill wording; Fixed for the rest — STEP-8 rewords the "Every cache names the signal" rule's `DropOnHandleCommit` clause to name the three known exceptions, and D-18 marks them in the table.

Minor, round 2: `SourcesModel.searchKeyCache` and `EvidenceGraphModel.snapshotMemo` are not single-entry `Memo` holders — Fixed in D-18. D-13's `Metas` error — already Fixed under X-9.

## After round 2 — the user's answer on M-14 (2026-11-11)

The user asked for the Fill question to be analyzed against the batch's purpose. Owners, not display code, read the registry; the cost the choice decides is the read–cast–compute–store block every owner writes (six today, one more per Batch 2 cache PR). The user chose typed get-or-fill handles: `Keyed[T]` with `Get(g, id, fill)` and `GetMany(g, ids, fill)` (batch, for S16-18's frontier loads) and `Whole[T]` with a lazy Build. Plan: D-1, D-2, D-8, D-9, D-10, D-12, D-21 rewritten; Rules this plan bends records it; STEP-8 rewords the rule; `NewReader` drops its always-default `rules` parameter (D-9, Stop condition added). Not re-reviewed: the change is the shape the reviewers' M-3 and M-13 pointed toward, and it touches no Current state row.
