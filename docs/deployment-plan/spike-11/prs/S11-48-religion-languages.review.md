# Review — S11-48 Religion and languages spoken on a person

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1 (no blocker, so no round 2). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user

None. The product calls were asked before writing: religion is cardinality multiple; the PR goes in Spike 11; existing catalogs get the Properties through S11-44's widened command (all user, 2026-10-25).

## Writer's own calls

- Keys `religion` and `languages_spoken`, term keys `catholic` and `english` (D-1).
- No Swift constants or L10n wording; labels come from the catalog (D-3).
- No `vocabkey` constants (D-4); bindings unlocked at sort 5 and 6 (D-5); the description text (D-6).
- One registry-driven Install test instead of a fourth hand-copied one, and tests outside `subjectvocab` stop restating the seed (D-7, D-8, from M-1 and M-2).
- S11-44 depends on S11-48, and its command adds missing seeded subject vocabulary but restores a binding only for a Property it adds itself.

## Round 1

### M-1 · major · T-1 would be the fourth hand-written "Install seeded X" test
One `TestInstallMatchesRegistry` that checks every installed row against its registry row, plus a check of this PR's registry values.
**Answer:** Fixed. D-7, STEP-1, T-1 and T-2. AC-1.4 dropped: user-term create is generic and covered by the gender test's "Agender". The older three tests stay; they also check user-term create and `Neutral`.

### M-2 · major · The person binding order and Property count are copied into four other tests
**Answer:** Fixed. D-8, STEP-2, T-3 and T-4: `graphcache` reads `BindingsForType("person")`; the handler count goes; `claimevidence` `seedOrder` is untouched.

### M-3 · major · A second sentence in conclusion-reconciliation.md becomes false
**Answer:** Fixed. STEP-4 rewrites the **Multiple** bullet's example; §4 and the DoD name it.

### M-4 · major · D-3 disagrees with add-seeded-vocabulary
**Answer:** Fixed. STEP-4 rewrites the skill's sentence; §4, Scope and the DoD name it.

### M-5 · nit · Where the term-multiple reconcile case lives
**Answer:** Logged. The pipeline table stays the home (one table for "multiple keeps distinct values").

### X-1 · major · The manual check is unreachable for the implementer
**Answer:** Fixed. §10 says to run it if the session can launch the app, else report "not run" with the reason. The Go tests cover every AC.

### X-2 · major · T-1 setup and the `propertyterms` calls are implicit
**Answer:** Fixed, superseded by M-1. STEP-1 names the setup, the lookups and the stored field names.

### X-3 · major · Placement of registry rows and doc lines is open
**Answer:** Fixed. §6 Go core gives positions and spacing; STEP-4 gives the doc lines.

### X-4 · major · Docs steps lack final wording
**Answer:** Fixed. STEP-4 gives each replaced and inserted sentence.

### X-5 · minor · check-plan-docs.py warnings
**Answer:** Fixed. STEP-4's Check says the `ideas/…` link-text warnings are expected.

### X-6 · minor · Doc housekeeping is underspecified
**Answer:** Fixed. STEP-4 spells out the map Status, `completed.md` heading and text, and the **Landed** line, as S11-41 does.

### X-7 · minor · Other `sections_test.go` cases
**Answer:** Fixed. With D-8, `sections_test.go` is not touched; STEP-3 says it is test-only.

### X-8 · nit · `subjectcomparetest.Metas()`
**Answer:** Fixed. STEP-2 says it is not touched.
