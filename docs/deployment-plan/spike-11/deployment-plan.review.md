# Review — Spike 11 PR map

**Verdict:** Fits with changes. **Rounds:** 1. **Reviewers:** maintainability (fresh agent, map questions).

## To user

- **Two preparatory PRs added: S11-22 (pure value types) and S11-23 (one home for primary kinds).** Recommendation: keep both. Without S11-22, every layer would either import `core/database` or define its own date and name structs. Without S11-23, the registry would add the fifth to seventh hand-written kind lists.
- **Two splits added: S11-24 (delete frequency weighting, out of S11-13) and S11-25 (claim evidence RPC and pin policy, out of S11-19).** Recommendation: keep both. S11-13 was well over 10 steps, and S11-19 would have broken the shipped sheet between PRs.
- **A matching skill (`add-match-rule`)** is added in S11-02 and grown by S11-10, S11-14 and S11-16. Recommendation: keep it. The name is a working name.

## Writer's own calls

- New PR IDs are S11-22 – S11-25, numbered in the order they were cut. Their place in the order comes from the slices and dependencies, not the number.
- The pin-table migration is conditional: S11-19's plan checks the tables, and cuts a migration PR ahead of S11-19 only if they can't hold one incoming Observation against many member Observations.

## Round 1

### K-1 · blocker · Pure value types need a preparatory move, not layer-1 copies
`datevalues.Value`, `namevalues.Value` and `properties`' value-type and cardinality constants live in packages that import `core/database`. About 65 files use the structs.
**Answer:** Fixed. S11-22 is added ahead of S11-02, with the skills updated. Working rule "Pure value types and kinds come first".

### K-2 · major · Alias seeds and name-pattern types have no home once `core/match` goes
**Answer:** Fixed in S11-03. The types move to `valuecompare`. The seeds, `Register` and `third_party/nicknames/` move to `namealiases`, with `add-seeded-vocabulary` and `seeded-vocabulary.md` updated.

### K-3 · major · `core/database/matching` is not dead; S11-13 too big
`promotealign/layer.go:48` calls `matching.SubjectsValues`.
**Answer:** Fixed. The working rule now says `SubjectsValues` moves to the loader in S11-13. Frequency removal is split out as S11-24.

### K-4 · major · A registry is due for primary kinds and property keys
**Answer:** Fixed. S11-23 is added. S11-10 ports `TestDefaultProfileCardinality` as "every registry `PropertyKey` is a seeded Property".

### K-5 · major · One guard, not one per layer package
**Answer:** Fixed. Working rule "One guard, not one per package". S11-02 ships it, modeled on `core/writes/begin_test.go`.

### K-6 · major · S11-19 mixed engine, FFI and the Swift pin flow
**Answer:** Fixed. The work is split into S11-19 (Go query) and S11-25 (RPC, Go default pins at Apply, Swift drafting moved, proto field 9 deleted). Working rule "The pin policy has one owner, in Go". The migration is conditional (see Writer's own calls).

### K-7 · major · Extend `RowFlags` / `Reason`; one wire vocabulary
**Answer:** Fixed in S11-16 and S11-17. Working rule "One wire vocabulary".

### K-8 · major · No skill teaches the new pattern; docs go stale
**Answer:** Fixed. Working rule "One matching skill, grown with the layers". Each migration PR names the docs and idea files it makes false (S11-05, S11-06, S11-16).

### K-9 · minor · S11-08: the Swift helper already exists
**Answer:** Fixed. `PVComboBoxMatch.normalize` is promoted, with a locale parameter. `DateValueTranscriptionMatch` is added, and the `shared-list-helpers.md` section is closed.

### K-10 · minor · Reuse the graphcache stats slot; name who deletes fan-out
**Answer:** Fixed. S11-14 deletes fan-out. S11-15 reuses the stats slot.

### K-11 · minor · S11-21: extend the existing benchmark; make the budget run
**Answer:** Fixed. S11-21 is a budgeted `Test…` on `bench_test.go`'s fixture. S11-24 deletes `BenchmarkLoadStatsNames`.
