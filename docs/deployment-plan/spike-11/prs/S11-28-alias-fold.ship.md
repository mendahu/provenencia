# Ship log — S11-28 Fold a nickname into the given name it stands for

**Plan:** [`S11-28-alias-fold.md`](S11-28-alias-fold.md) · **Branch:** `s11-28-alias-fold` · **PR:** #589
**State:** ready

## Round 0 — implement
Base `main` @ `67988a5d`. Local head `cd44ffeb` (3 commits, not pushed). Gate FAIL go-test (`core/search` only); every other check PASS.

**Stopped on:** §11 "A test outside … T-11's named assertions changes result." `core/search/handles_test.go` records *James Robins* / *Jim Robins* (lines 129, 219, 279, 285, 315). *Jim* now folds into *James*, so `TestHandleSearch` (two subtests), `TestHandleSearchFindsOutvotedNames` and `TestHeaderSearchFindsTheListTitle` fail. This is the same James/Jim fixture that T-11 already rewrites elsewhere; T-11 missed this file. Verified: the fixture lines are as reported.

Deviations: `promotealign/profiles.go` did not already import `autoreconciler` (the plan's §4 said it did); the implementer added the import and dropped an unused `matchregistry` import. `autoreconciler_test.go` local `jim` renamed to `kenneth` (T-11 follow-through).

**Resolved (user, 2026-10-27):** "Amend T-11 to include core/search/handles_test.go (Jim → Kenneth, expected forms follow), note it in the review log, and resume the ship loop." Plan amended in `c593a517`; the implementer resumes round 0 with that fixture change.

Resumed run: head `abc254d0` (T-11 `core/search` swap done). Gate FAIL go-test, two failures, both verified by the orchestrator:

1. `TestHandleSearchFindsOutvotedNames` (`core/search/handles_test.go:185`): fixture `Thomas Lee`, `thomas lee`, `Tom Lee`; search "Tom" returns `[]`. *Tom* is in `{thomas, thom, tom, tommy}`, so it now folds into *Thomas* and is no longer a cached value. Outside T-11 (T-11 covers only *Jim*). Note: this means a folded nickname is no longer searchable.
2. `TestLayerPackages/core/valuecompare/numbers` (`core/matchregistry`): `aliases.go:45, 88, 90, 160` — the `2` in `map[[2]string]bool`. The guard counts the array length as a number. The plan's D-1a specifies `prefer map[[2]string]bool`, so this is a plan defect, not an implementer slip. The first round-0 run did not report it.

**Resolved (user, 2026-10-27):** "Amend D-1a to key prefer by AliasDirection, add the Tom → Thommas swap in TestHandleSearchFindsOutvotedNames to T-11, log both in the review log, record alias-aware search as a follow-up, and resume." Plan amended in `233481cc`.

**Follow-up FU-1 — Alias-aware search.** A folded nickname is no longer a reconciled value, so search for "Tom" no longer finds a Person shown as *Thomas*. Search should expand a typed nickname through the alias tables. Home: `core/search` (omnibar-search), spike close triage.

Resumed run: head `b8d15e6e`, pushed, draft PR #589, landing docs committed. Orchestrator gate on `b8d15e6e`: PASS (diff-check, gofmt, go-vet, go-test, localizable, guidance-names, plan-docs, xcode-test, swift-warnings). Report: no deviations.

## Round 1 — review
Base `67988a5d` · head `b8d15e6e`.
Reviewers: CON (Sonnet), MNT, SEC+PRF (matching layers `core/autoreconcile`, `core/valuecompare`, `core/database/namealiases`; `core/database/**`). Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | No Landed note listing differences from the plan | Fix now → F-1 | verified: plan line 3 is Landed, DoD line 255 requires a note, `rg Landed` finds none. Differences: `profiles.go` import, T-11 widened twice, D-1a key type, plus F-2's supersession and the search regression |
| MNT-1 | MNT | major | `valuerules.Product` builds its alias index outside `namealiases.buildIndex` | Fix now → F-2 (same-pass cleanup) | verified: `valuerules.go:43` and `namealiases.go:226` are the only non-test `NewAliasIndex` calls; `buildIndex` is the declared one keying rule. Supersedes plan §4's `valuerules` row (`valuecompare.NewAliasIndex(..., rules.Words)`); behavior identical, so not To user. Recorded in F-1's Landed note |
| MNT-2 | MNT | major | `nicknames_seed.go:27` and `matching-model.md:213` now false about who reads the seed | Fix now → F-3 | verified: "Scoring reads the catalog copy, not this variable."; "(catalog rows, passed in at run time)" |
| MNT-3 | MNT | minor | `subjectprofile` test `reconcile` helper says "the way the cache does" but passes `matchregistry.Default().Value` | Fix now → F-4 | verified `helpers_test.go:63-72`; the PR's `Build` comment makes it false. A `subjectprofile` result change is within §11's allowed packages |
| MNT-4 | MNT | minor | Fourth copy of the names test loop | Fix now → F-5 (same-pass cleanup) | PR adds the fourth copy (`names_test.go:737-756`) |
| MNT-5 | MNT | nit | `Expands` comment says "the table's Directions"; `prefer` is index-wide | Fix now → F-6 (stale comment in a file this PR adds) | `aliases.go:143-146` |
| CON-2 | CON | minor | Alias-aware search recorded only in logs | Follow-up (FU-1) | User asked for it as a follow-up; it rides in the PR body's follow-up list and spike close (S11-99) triages ship-log follow-ups. F-1's Landed note names it |
| SEC-out-1 | SEC→MNT | minor | `autoreconciler.ValueRules` is an exported mutable var | Rejected | D-6 names `var ValueRules` exported, used by `promotealign`; no code writes it, and T-7 pins the seed to `CacheVersion`. An accessor is churn the Decision ruled out |

Nits (logged): `MNT-6 · core/valuecompare/aliases.go:91-104 · set head in the existing token loop` · `CON-3 · S11-28-alias-fold.md:208 · "None —." dangling dash (rides along in F-1)`
SEC+PRF: no findings (verdict Clean).

### Fix plan 1

**F-1 — Landed note on the plan** (from CON-1, CON-2, CON-3)
Files: `docs/deployment-plan/spike-11/prs/S11-28-alias-fold.md`.
Do: under the Status line add a **Landed note** listing every difference from the plan: (1) `promotealign/profiles.go` did not already import `autoreconciler` (§4 said it did); the import was added and the unused `matchregistry` import dropped. (2) T-11 grew to `core/search/handles_test.go` (Jim → Kenneth; outvoted Tom → Thommas), user 2026-10-27. (3) D-1a keys `prefer` by `AliasDirection`, not `[2]string` (layer-numbers guard), user 2026-10-27. (4) `valuerules.Product` gets its index from `namealiases.ProductIndex()` (F-2), superseding §4's `valuerules` row; same keying, one rule. (5) Consequence: a folded nickname is no longer a reconciled value, so searching "Tom" no longer finds a Person shown as *Thomas*; alias-aware search is a follow-up. Also fix §6 Strings "None —." → "None."
Check: `python3 scripts/check-plan-docs.py` 0 errors.

**F-2 — One alias-index builder** (from MNT-1)
Files: `core/database/namealiases/namealiases.go` (or `registry.go`, beside `ProductTables`), `core/database/valuerules/valuerules.go`.
Do: add `func ProductIndex() valuecompare.AliasIndex { return buildIndex(ProductTables()) }` with a one-line doc comment; `Product` sets `rules.Names.Aliases = namealiases.ProductIndex()`. Keep `ProductTables` (T-7 uses it).
Check: `git grep -n "NewAliasIndex(" -- '*.go' ':!*_test.go'` prints only `core/database/namealiases/namealiases.go`; `TestProductAliasTablesPinCacheVersion` and `TestValueRulesFoldNicknamesOnly` pass.

**F-3 — Fix the two statements the PR made false** (from MNT-2)
Files: `core/database/namealiases/nicknames_seed.go:27`, `docs/matching-model.md:213`.
Do: seed comment → "Matching reads the catalog rows; the reconciler cache reads this seed through valuerules.Product, so an edit bumps autoreconciler.CacheVersion." `matching-model.md` "Not knobs": `value.Names.Aliases` is catalog rows for matching, and the product seed for the reconciler (`valuerules.Product`).
Check: `git grep -n "Scoring reads" core docs .agents` prints nothing.

**F-4 — `subjectprofile` test helper reconciles with the cache's rules** (from MNT-3)
Files: `core/subjectprofile/helpers_test.go`.
Do: pass `autoreconciler.ValueRules` instead of `matchregistry.Default().Value`; drop the `matchregistry` import if unused. If a `subjectprofile` test changes result, stop and report which (do not edit expectations).
Check: `go test -tags fts5 ./core/subjectprofile/...` passes.

**F-5 — Shared names test helper** (from MNT-4)
Files: `core/autoreconcile/names_test.go`.
Do: extract one helper (e.g. `checkNames(t, rules, in, want, state, reasons)`) holding the Reconcile → `have` → compare → state → reasons block, and call it from all four tests (copies at ~246, ~599, ~666, ~737). No assertion changes.
Check: `git grep -n "have = append(have, cl{shape" core/autoreconcile` prints exactly one line; `go test ./core/autoreconcile/` passes.

**F-6 — `Expands` comment** (from MNT-5)
Files: `core/valuecompare/aliases.go:143-146`.
Do: "the table's Directions" → "a Direction in any table".
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `b8d583c9`, pushed. Implementer gate PASS. F-1, F-2, F-3, F-5, F-6 done. F-3 also fixed a live "Scoring reads" claim in `name-candidates.md:42`; the remaining hit is the plan's own quote of the old string (S11-28-alias-fold.md:225), left as is.

F-4 blocked and reverted: passing `autoreconciler.ValueRules` from `core/subjectprofile` tests makes `TestLayerPackages` fail (a layer package's test reaches the catalog through `database/autoreconciler`). Decision (orchestrator): take MNT-3's offered alternative — reword the helper comment, keep `matchregistry.Default().Value`. → F-4b.

**F-4b — Say what the `subjectprofile` test helper reconciles with** (from MNT-3)
Files: `core/subjectprofile/helpers_test.go`.
Do: the `reconcile` comment no longer says "the way the cache does"; say it reconciles with the shipped registry (no alias tables, since a layer package cannot reach the product seed in `core/database`), with the property's value type and cardinality and no concluded value.
Check: scripts/check-pr.sh PASS.

F-4b done, head `d3983c7e`, pushed. Orchestrator gate on `d3983c7e`: PASS.

## Round 2 — review
Scope `b8d15e6e...d3983c7e`. Reviewers: CON (Sonnet; CON-1 major), MNT (MNT-1, MNT-2 majors). Not re-sent: SEC+PRF (no findings).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | major | Landed note | Resolved | Note under Status lists five differences; plan-docs 0 errors |
| MNT-1 | MNT | major | One alias-index builder | Resolved | only non-test `NewAliasIndex(` call is in `namealiases.buildIndex` |
| MNT-2 | MNT | major | Statements about who reads the seed | Resolved | seed comment, `matching-model.md:213`, `name-candidates.md:42` |
| MNT-3, MNT-4, MNT-5 | MNT | minor/nit | Helper comment, shared names check, `Expands` comment | Resolved | read by the orchestrator and confirmed by MNT round 2 |
| CON-4 | CON | minor | No blank line after the Landed note | Logged | round 2 adds only blockers; the header lines below it carry no blank lines either |

No new blockers. No Fix plan 2.

## Finish
Merged `origin/main` (`e1f54022`, S12-20 plan docs only) into the branch; orchestrator gate on `59c71fd9`: PASS.

## Result
Gate PASS · CI pending · open: none · follow-ups: FU-1 (alias-aware search; CON-2) · to user: three plan amendments during round 0 (T-11 `core/search` Jim; T-11 Tom → Thommas; D-1a `AliasDirection`), all answered 2026-10-27 · CON miss: none
