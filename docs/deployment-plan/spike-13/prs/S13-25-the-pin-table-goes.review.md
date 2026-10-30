# Review — S13-25 The pin table goes

**Verdict:** Fits with changes — Ready with changes (round 1)
**Rounds:** 1. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None yet.

## Writer's own calls
- One format step that rebuilds `identity_claims` (D-1), not `DROP COLUMN`: SQLite refuses to drop a column with an FK or index.
- `releaseClaim` keeps its `releaseFacets(KindIdentityClaim)` call with no facet registered, so a later claim facet is released (D-3).
- The whole `claimconfidencegrades` package and `apperr.CodeClaimConfidenceInvalid` go; S13-24's grade subtest goes with the column (D-5).
- Pin tests are rewritten to keep their non-pin cover (member delete with cache check, handle reporting, bridge delete), not deleted (D-7).
- No `CacheVersion` bump, no proto change (D-11).
- `ideas/identity-claim-review.md` is marked superseded (D-9); three docs added to the map's S13-25 list.

## Round 1

### X-1 · blocker · STEP-1's check cannot pass: `effects` init panics once `catalogmodel` drops the tables
`core/database/listen.go` imports `effects`; its `init` panics on an entry for an undeclared table, so `TestCreateOpen` dies after STEP-1.
**Answer:** Fixed — STEP-1 now deletes D-4's three `registry.go` entries with the model; STEP-3 keeps the effects tests and comment.

### X-2 · major · The migration ACs have no runnable check on a Linux host
**Answer:** Fixed — §10 says the user runs it on a Mac; an implementer without Xcode reports §10 and STEP-5's Xcode check as not run and writes no upgrade harness (`add-catalog-migration` forbids one).

### X-3 · major · D-11 and the Stop-condition `rg` give other output than stated
**Answer:** Fixed — D-11 lists `catalogmodel/model.go` among the expected hits; the Stop condition greps `confidence_grade_id` only, against the three named files.

### X-4 · minor · Test edits where the implementer must choose
**Answer:** Fixed — D-7 now says imports follow the compiler, `cb` and `obsA` go from the fixture, T-1 replaces the two subtests in place, and T-4 drops `person`, the `edge` loop, the `PinTx` revision and both audit-type assertions. D-4 cites line 186.

### X-5 · minor · Swift and docs edits
**Answer:** Fixed — D-8 says no new Swift test (the existing `alsoAffects` test covers the default branch); D-9 gives one rule for every doc: delete the DDL and paragraphs for dropped tables and columns, leave one "dropped by S13-25" sentence; `seeded-vocabulary.md` §5.5 becomes that sentence.

### M-1 · major · `cacheMatchesRebuild` is the fourth copy of "cache upkeep equals a rebuild"
Copies: `pinnedHandle.cacheMatchesRebuild`, `bridgeWorld.cacheMatchesRebuild` + `cacheDump`, `assertCacheMatchesRebuild` + `cacheLines`, `fixture.assertUpkeepEqualsRebuild`.
**Answer:** Fixed — D-7 and Codebase fit add `autoreconcilertest.AssertCacheMatchesRebuild` (new package beside `autoreconciler`, the `promotetest` shape); T-2 and the five `promote` call sites use it; `bridgeWorld.cacheMatchesRebuild`, `cacheDump`, `assertCacheMatchesRebuild` and `cacheLines` are deleted. `assertUpkeepEqualsRebuild` stays: it also compares search documents.

### M-2 · major · The sweep is name-based and misses prose the PR makes false
**Answer:** Fixed — STEP-6 adds `add-audit-scope` lines 28 and 54, `catalog-deletes.md` line 46, conclusion §3 item 2; D-10 adds the `batch_test.go` line-305 comment; the DoD gains a prose `rg -w 'pins?|pinned|exhibit'` line. Declined for `docs/deployment-plan/spike-14/app-health-audit.md` L-2: it is a dated audit record, and its finding stays true for `deleteimpact.release_incomplete`.

### M-3 · minor · Near-identical `newFixture` vocabulary setups
**Answer:** Logged — §3 Out, not planned; it touches about 30 test files, and this PR shrinks the copies.

### M-4 · minor · `cascadeLines` keys copy on via strings with no parity guard
**Answer:** Logged — §3 Out, home S13-08's plan (its unpromote preflight confirm keys on vias too). This PR removes a case and adds none.
