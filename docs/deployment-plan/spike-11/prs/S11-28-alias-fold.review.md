# Review — S11-28 Fold a nickname into the given name it stands for

**Verdict:** Fits with changes — Not ready → fixed in round 1
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. Product calls were asked before planning (2026-10-19): shorter folds into longer, nicknames and abbreviations both fold, the folded name disappears.

## Writer's own calls
- Direction by rune length because the catalog stores no canonical token per alias set; *Abel* / *Ebbie* folds the wrong way (needs both on one person). Canonical tokens are parked in `docs/ideas/international-names.md`.
- The reconciler reads the **product seed** of alias tables, not a catalog's rows, so the cache still changes only with `CacheVersion`; a guard test pins the seed to the version.
- `Compatible` (Promote exhibits) follows the fold, so *Jim* / *James* now counts as agreeing there.
- Which part types alias-fold is decided by the alias tables' roles: given, nick and untyped (`undetermined`, untyped parts). Surnames never.

## Round 1

### X-1 · blocker · Promote's profile builder re-reconciles with alias-free rules
`core/database/promotealign/profiles.go:169` passes `matchregistry.Default().Value` to `subjectprofile.Build`, which re-runs `autoreconcile.Reconcile` for `Options.Excluding`.
**Answer:** Fixed — D-6: `autoreconciler.ValueRules` (exported) is the one value; `profiles.go` passes it; `subjectprofile.Build` comment and `add-match-rule` skill updated in STEP-3; AC-2.3, T-6. `promotealign` already imports `autoreconciler` (checked with `go list -deps`).

### X-2 · major · T-1 rows leave expected results to the implementer
**Answer:** Fixed — §2 gives members, parts and state for the folding ACs; "unchanged" ACs assert equality with the alias-free run (T-1), which is deterministic without guessing the tie order.

### X-3 · major · D-7 test flip under-specified
**Answer:** Fixed — D-7 / T-3: keep the `Compare == Contains` assertion, read `aliasRules`, add Jose / José → not compatible.

### X-4 · major · Existing tests that may move are not named
**Answer:** Fixed — §8 "Must stay green": invariants run under `testRules`; AC-1.4 runs reversed in T-1; a `promotealign` change is allowed only for an alias pair now agreeing, listed in the PR; matching stop condition.

### X-5 · minor · `completed.md` path
**Answer:** Fixed — full path in STEP-3 and DoD.

### X-6 · minor · `typ` / role lookup assumption
**Answer:** Fixed — D-3 rewritten (see M-1, M-6).

### X-7 · minor · T-2 and manual check
**Answer:** Fixed — T-2 folded into T-1; manual check names the sorted Western style.

### X-8 · nit · Doc lines to replace
**Answer:** Fixed — STEP-3 quotes each sentence replaced (conclusion-reconciliation lines 132 and 167, `names.go` header).

### M-1 · blocker · Hard-coded given / nick beside the alias tables' Roles
**Answer:** Fixed — D-3: `Aliases.Hit(a, b, role, role)` with `role := Western.Role(typ)`; the module names no role. Format-agnostic bullet amended in STEP-3.

### M-2 · major · `AliasTables()` includes plugin tables; seed edits don't bump `CacheVersion`
**Answer:** Fixed — D-6: `namealiases.ProductTables()` (seed only); T-7 `TestProductAliasTablesPinCacheVersion` pins a hash of the seed.

### M-3 · major · Docs and skills made false
**Answer:** Fixed — STEP-3 adds `seeded-vocabulary.md` §4.5, `add-seeded-vocabulary` and `add-match-rule` skills, `namealiases` comments, Format-agnostic bullets, `compareValue` comment; DoD line extended.

### M-4 · major · Replace `seedRules` in `namealiases/registry_test.go`
**Answer:** Declined — `seedRules` reads `AliasTables()`, which includes `Register`ed tables on purpose (`TestRegisterAddsAnAliasTable`). After M-2 the production value reads the seed only, so the two are different values.

### M-5 · minor · Assemble product rules in `valuerules`
**Answer:** Fixed — `valuerules.Product()` beside `Load`; no import cycle (`go list -deps`).

### M-6 · minor · Use `NamePattern.Role`
**Answer:** Fixed — in the D-3 sketch.
