# Review — S16-32 Promote re-ranks only the Subjects a write moved

**Verdict:** (round 2 pending)
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None yet.

## Writer's own calls
- The signal is every key a changed handle was filed under, before and after the commit, not only the keys that gained or lost it (the map's earlier wording). A handle whose values change and keys do not can still move into or out of a Subject's top candidates. The map's row, scope entry and Check and the design's line were reworded at planning.
- `shortlist.Index.Filed` is a read accessor; `Put` and `Remove` keep their signatures (S16-20's **Next adds** sketched them returning keys).
- `candidates` keeps `DerivedFrom MatchIndex` instead of declaring the three sources directly.
- A forgotten Subject is ranked again lazily, on its graph's next proposal, not inside the commit.
- The forget is a forward scan over held Subjects' keys (about 2 ms at 64 Source-A-sized graphs, about 5 ms with one 10,000-Subject graph), not an inverted index.
- Bar 10's ratio is gated at 4,000 persons in `TestLargeSourceBudgets` and recorded at both sizes by the benchmark; the existing `budgetAfterCommit` stays.
- After M-1: a new registry mechanism, `Whole.Patched` plus `KeyedEntry.Narrow`, carries the hand-off from the match index to the candidates.

## Round 1

### M-1 · blocker · The candidates' per-commit update is a call from another entry's Apply, which neither entry declares
`applyMatchIndex` called `forgetCandidates` directly; `docs/catalog-caches.md` line 5 lists three signal forms and the `.mdc` rule says an entry declares its own; producers never name consumers today (`Whole.Set` → `dropReaders`).
**Answer:** Fixed — D-3 adds `Whole.Patched(g, change)` and `KeyedEntry.Narrow`; D-5 ends Apply with `w.Patched(g, moved)`; D-6/D-7 make `forgetCandidates` the candidates' declared `Narrow`; T-10 holds it; STEP-3 updates `catalog-caches.md` line 5 and the rule. The map's rule "the registry carries the plumbing" (user, 2026-11-13) backs it, so no rule is bent.

### M-2 · major · STEP-8 and the DoD sweep leave the cache rule text false
**Answer:** Fixed — STEP-3 rewrites `catalog-caches.md` line 5 to name the fourth form and the `.mdc` clause to name `Patched`, `Narrow` and `Keyed.Update`; the DoD sweep `rg`s the old three-form sentence.

### M-3 · major · D-9 adds a third hand-rolled "GC, then time, best of N" loop
**Answer:** Fixed — D-9 adds `databasetest.Collected(n, setup, read)`; `Warm`, the two `recordlinks_perf_test.go` subtests and `proposalAfterCommit` use it; T-11; STEP-1.

### M-4 · minor · Each ranked Subject's lookup keys are computed twice
**Answer:** Fixed — D-2 adds `Index.Find`; `Lookup` is `Find(LookupKeys)`; D-4's `rankCandidates` computes the keys once and passes them to `seedHandles(m, kind, keys)`.

### M-5 · minor · Four new test hooks beside three near-identical ones
**Answer:** Fixed — D-11: `HoldCandidatesForTest` gains `subjects, keysEach`; one `HeldSubjectsForTest` returning a set replaces the two probes; `ForgetForTest` stays (no existing hook narrows).

### M-6 · minor · Next adds claims B-11 as a second reader without evidence
**Answer:** Fixed — the B-11 clause is gone; Next adds names any later keyed reader of a patched whole entry.

### X-1 · major · Status Draft and no review, but the Implementer block assumes Planned
**Answer:** Fixed — this log; the status moves to Planned only after round 2. STEP-8 names the map status value `Landed #N`.

### X-2 · major · T-6 / D-11 leave the synthetic ids and keys to the implementer
**Answer:** Fixed — T-6 gives the 64 graph ids (index big-endian in the first 4 bytes) and the 20 keys; D-11 gives each Subject's id and keys.

### X-3 · major · A memo forgotten to empty: nil vs empty map under DeepEqual
**Answer:** Fixed — D-6 builds the new map with `make`, and says why Verify passes; D-11's hold uses `make` too.

### X-4 · minor · D-9 does not say how one `after` serves two checks
**Answer:** Fixed — D-9: measured once, passed to both `WithinBudget` calls.

### X-5 · minor · STEP-8's size cell and the 1 MB stop condition lack a format and a base
**Answer:** Fixed — STEP-8 gives the cell's text; the Stop condition compares with STEP-1's figure on the same machine.

### X-6 · minor · `CandidateHeldForTest` vs `CandidatesHeldForTest`; export_test.go marked new
**Answer:** Fixed — with M-5 the one-letter hook is gone; D-11 says the file exists.

### X-7 · minor · T-1's held states depend on date buckets too
**Answer:** Fixed — the Stop condition now covers a shared date key (1790 against 1817) as well as names.

### X-8 · nit · cache-architecture line numbers
**Answer:** Fixed — STEP-8's Files line names line 67 by its bullet.
