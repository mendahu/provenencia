# Review — S11-28 Fold a nickname into the given name it stands for

**Verdict:** Fits with changes — Not ready → fixed in round 1; round 2 Ready
**Rounds:** 4. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. Product calls were asked before planning (2026-10-19): shorter folds into longer, nicknames and abbreviations both fold, the folded name disappears.

## Writer's own calls
- Direction (user rule, 2026-10-19): two given names the alias list pairs fold; the set's head is kept; for the 144 pairs listed both ways, `nicknameDirections` (Appendix A, 108 directed, 36 never fold) decides. One-way pairs keep the head. Pairs that joined two different names (31, e.g. *Wilma* / *William*, *Andrew* / *Andrea*) were removed from the seed in #499 after a review with the user.
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

## Round 2

Reviewer (opus, both lenses): no new blockers; every round-1 answer accepted, M-4's decline upheld. Verdict: Ready. Non-blocking notes, all fixed:
- `AliasIndex.Hit` returns two values — sketch uses an `aliasAllowed` helper.
- STEP-3 quoted the skill's wording for `seeded-vocabulary.md` §4.5 — both quotes now exact.
- AC-2.3 had no observable test (`Excluding` has no caller) — AC reworded to the code fact; T-6 says so.
- D-7 overstated exhibit agreement (exhibits read catalog alias rows) — reworded.
- T-7 re-pin did not force a bump — pins `(CacheVersion, hash)` together.

## Round 3 (after the user changed the fold direction, 2026-10-19)

The user asked that the direction come from data we control, not word length. The first revision curated the 144 pairs listed both ways in the seed. A fresh reviewer (opus, both lenses) found:

### R3-1 · blocker · A set's first word is not always formal
712 one-way pairs between set heads would fold real names together (*William* → *Wilma*, *John* → *Ivan*, *Peter* → *Patrick*).
**Answer:** To user, then fixed. The user chose to curate which of the 406 dual-role heads are nicknames. D-1a now folds only a nickname into a formal name; two formal names never fold. Appendix A lists 152 nickname heads and 254 kept formal; AC-2.4 / T-10 include the negative pairs.

### R3-2 · major · Editing the seed removed 46 alias pairs from new catalogs
**Answer:** Fixed. `nicknameSets` is no longer edited; `nicknameHeads` sits beside it, so matching is unchanged.

### R3-3 · minor · D-3 and the sketch disagreed on the role check
**Answer:** Fixed. D-3: `Expands` carries the role check; `fold` does not call `Hit`.

### R3-4 · nit · `aliases_test.go`; AC order
**Answer:** Fixed.

## Round 4

Reviewer (opus, both lenses), on the nickname-heads version. Verdict: Not ready.

### R4-1 · blocker · §6 `Expands` sketch was the old design
**Answer:** Fixed — sketch matches D-1a (head, `prefer`).

### R4-2 · blocker · Existing *James* + *Jim* assertions break in packages outside the stop-condition list
**Answer:** Fixed — T-11 names each assertion and swaps *Jim* for *Kenneth*; stop condition allows them.

### R4-3 · major · Real names among the 1,020 nickname-only names fold into other names (Ralph → Raphael, Carl → Charles)
**Answer:** To user. The user ruled (2026-10-19): two given names the list pairs fold regardless of whether either is a "real" name; a name alone never folds. This dropped the nickname-heads list: the head of a set holding the other name is kept, and only the 144 two-way pairs need a recorded direction (Appendix A, as data beside the seed).

### R4-4 · major · Steve, Jon on the formal list
**Answer:** Moot — that list is gone. *Steve* + *Stephen* folds (one-way pair); AC-2.4.

### Minors
**Answer:** Fixed — hash description points to T-7; STEP-3 rewrites the grouping-policy and §7.2 Fold first sentences; header round count; Writer's own calls updated.

## After round 4 (user, 2026-10-19)

The user asked whether *Wilma* is a nickname of *William*. It is not: the upstream row `{wilma, billiewilhelm, william}` is an error. With the user's agreement, #499 (the PR carrying this plan) also edits the seed: 31 pairs that join two different given names — masculine and feminine forms (*Andrew* / *Andrea*, *Louis* / *Louise*, *Felix* / *Felicia*) and upstream errors (*Paul* / *Polly*, *William* / *Bela*) — were removed; five sets left with one name were dropped (1,227 sets). Candidates came from an offline name-gender dataset; unisex short forms (*Alex*, *Chris*, *Jamie*, *Kit*) were kept. Appendix A was regenerated (143 two-way pairs: 108 directed, 35 kept). Matching for new catalogs no longer pairs these names; `go test -tags fts5 ./...` passes.

2026-10-21 — S11-25 deleted `autoreconcile.Compatible` with its last caller (`promotealign/exhibits.go`); D-7, T-3 and their rows removed. Other numbering unchanged.
