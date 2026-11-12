# Review — S16-12 Source records in the graph cache

**Verdict:** round 1: Fits with changes — Not ready; round 2 pending
**Rounds:** 1 (round 2 pending). **Reviewers:** maintainability (Opus), executability (Opus)

## To user

- **M-2: Delete the title fallbacks and the planned `source_ref`, or write them once?** Recommended: delete them. Answered 2026-11-12: delete. A Source title and a graph name are never empty (`sources.Create`/`Update`, migrations 000011 and 000055). D-11 now adds no `source_ref`, and D-14 shows titles as stored.
- **M-4: Keep `SourceScopes` as the signal, or add an exact keyed read?** Recommended: keep it. Answered 2026-11-12: keep it. Scale, Edges and STEP-8 now state the fan-out and document the dependency.
- **X-8: Leave the Source's cover off the record?** This bends the rule "A Source's words live once". Answered 2026-11-12: leave it out; the rule now says cover is added with its first reader. The row is recorded under **Rules this plan bends**.

## Writer's own calls

- D-2: the entry has graph lifetime. A read on a write's graph returns `ErrRecordsOnWrite`; a read on a nil graph returns `database.ErrClosed`.
- D-7: Apply patches the index in place, then calls `Records.Set` with the same pointer.
- D-13: a graph or Artifact with no Source record is an error, not a skipped row.
- D-12: `ClaimGroup` holds one `Source sources.Record`; `PersonRecordLinks.Graphs` becomes `[]SourcedGraph`.
- The record links wire keeps `source_id`. The new `.source` follow reads it from the value, so it is no longer unread (D-11; this changes the map row).

## Round 1

### M-1 · blocker · Apply changes the held value in place, outside the handle
Apply bypassed `Whole.Set`, so an entry declared `DerivedFrom: Records` (S16-13, S16-15 and S16-30 may add one) would never be emptied. This also goes against `catalog-queries.mdc`.
**Answer:** Fixed. D-7 now calls `Records.Set(g, idx)` after the patch, which costs O(1). T-2 registers a test entry derived from `Records` and asserts that each commit empties it.

### M-2 · major · The fallbacks guard states the catalog forbids
**Answer:** To user, answered: delete. D-11 and D-14 are rewritten, US-3 and AC-3.x are reworded, and T-12's fixtures now use titled Sources and named graphs. The STEP-8 docs drop P-22's "missing `source_ref`".

### M-3 · major · The credibility rule belongs with the grades, not the reconciler
**Answer:** Fixed. D-4: `sourcecredibilitygrades.DefaultKey` and `RelativeOrderColumn` hold the rule, and both the reconciler and `sources` use them. `seedGrades` uses `DefaultKey`.

### M-4 · major · `SourceScopes` fires on nearly every write in a Source, and the coupling is undocumented
**Answer:** Fixed for Scale, Edges, T-2 (an Observation edit is now a row), T-13, and the STEP-8 lines in `add-audit-scope` and the effects README. The question of a narrower signal went To user, who chose to keep `SourceScopes`.

### M-5 · minor · A surviving fallback belongs in Platform's title-display files
**Answer:** Moot. No fallback survives (M-2).

### X-1 · blocker · T-2 asked for an Artifact move and a credibility clear, and no API makes either
**Answer:** Fixed. Both cases are dropped from T-2 and AC-2.1. The move goes to S15-24 (Out row), and Next adds notes that `readChangeColumn` names the old and new Source.

### X-2 · major · STEP-3's check could not compile because the handler test was missing
**Answer:** Fixed. `api/ffi/handlers/person_detail_test.go` is now in STEP-3's Files, and STEP-5's check runs `./api/ffi/handlers/...`.

### X-3 · major · `registry_completeness_test.go` was not named
**Answer:** Fixed. D-9 and T-8 name both rows, and the file is in STEP-5.

### X-4 · major · STEP-5's `rg` check could never pass
**Answer:** Fixed. The check is narrowed to non-test field reads plus `eg.Source…` in the tests.

### X-5 · major · The graphfill rewrite dropped coverage
**Answer:** Fixed. T-5 assigns each case to a test, keeps the not-found and graph-head assertions, and updates `TestTwoEvidenceGraphsLoadApart`.

### X-6 · major · A Swift test reads the deleted `graphSourceRefs`
**Answer:** Fixed. T-12 renames the test and drops that assertion.

### X-7 · major · `invalidationsComeFromRegistrySpecs` was not covered
**Answer:** Fixed. It is added to T-11, and the other constructors are noted in T-10 and T-11.

### X-8 · major · D-3 bends a working rule
**Answer:** To user, answered (see To user). The rule is now listed under **Rules this plan bends**, and STEP-8 rewords it.

### X-9 · minor · How the detail scan gets the index
**Answer:** Fixed. D-15: `fieldsOf` calls `RecordsOf` once and passes the index to `loadRecords`.

### X-10 · minor · Edges of `sources.Records` left open
**Answer:** Fixed. D-3 says `ArtifactLabels` is always non-nil. D-2 says a nil graph returns `database.ErrClosed`. T-2 calls `g.Verify()`. D-6 names the loader. D-1 puts the tests in the internal package.

### X-11 · minor · Benchmark set-up and the DoD benchmark line
**Answer:** Fixed. T-13 names the raw-`INSERT` seeding and times Apply through `g.OnCommit`. The DoD says "before and after" for `BenchmarkCacheSizes` and "after" for `BenchmarkSourceRecords`. STEP-1's figures go in the report to the orchestrator.

### X-12 · minor · Three Current state wordings
**Answer:** Fixed. `Install` is in `registry.go`; the count test compares a small and a big Person and counts only graph queries; 19 entries set `SourceScope`. AC-2.4 no longer claims a count for the warm-up.

### X-13 · nit · Wording slips
**Answer:** Fixed: the STEP number, the T-4 file, the `measure` body, and the README sentence span.
