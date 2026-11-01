# S14-18 — Name folds never crash

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the reconciler's Go fold only; every view draws as before)
**Board:** —
**Depends on:** —
**Blocks:** S14-26
**Branch:** `s14-18-name-folds-never-crash`
**Version bump:** patch — a crash fix that changes no feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md) (the reconciler section and the `CacheVersion` rule)
**Rules that bind:** `.cursor/rules/match-rules.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [audit 2 B-1](../app-health-audit-2.md#b-1--the-reconciler-panics-on-folds-that-do-not-chain); [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §5 step 3 and §7.2 Fold; working rules: "A nickname folds into its set's formal name, as today", "Proposals are liberal; the researcher decides", "Audit items keep their IDs"
**Planned:** 2026-11-01 against `main` @ `be51621f`
**Review:** [`S14-18-name-folds-never-crash.review.md`](S14-18-name-folds-never-crash.review.md) — not reviewed yet

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A person whose records hold a chain of name forms, such as B., Bill and William, files, reopens and shows one formal name instead of crashing the app. Today the reconciler assumes that when B. folds into Bill and Bill into William, B. also folds into William. Alias and prefix folds break that assumption, and the reconciler writes to a nil map (audit 2 B-1). The same call runs on every write that changes a person's members, in Promote's profile build, and in the cache rebuild on open, so a project holding such a person can crash on every launch. This PR makes the fold resolve each value through the chain to the best formal value it reaches.

**Why now.** It is a crash on real data: the probe at the planning base panics on eight everyday shapes with the product seed (§6 Edges).

**What stays.** Which pairs fold (`nameModule.fold`, `AliasIndex.Expands`, the product seed and `nicknameDirections`) is unchanged. A nickname still folds into its set's formal name (John, John, Ivan still shows Ivan: the user, 2026-11-01, reversing audit 2's B-3 decision). Majority, weak drop and assembly are unchanged.

## 2. User stories

**US-1.** As a researcher, when a person's records hold name forms that fold in a chain, the person files, reopens and shows the formal name the chain reaches, so the app does not crash.

Every example: one candidate per name, `surname=Robins` on each, product rules (`autoreconciler.ValueRules`), cardinality single.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Given B., Bill, William | One value, given `William`; no panic |
| AC-1.2 | Given D., Dick, Richard | given `Richard` |
| AC-1.3 | Given P., Peggy, Margaret | given `Margaret` |
| AC-1.4 | Given Mo, Molly, Mary | given `Mary` |
| AC-1.5 | Given Caroline, Chick, Charles | given `Charles` |
| AC-1.6 | Given Ann, Anna, Hannah | given `Hannah` |
| AC-1.7 | Given Dick, Ricky, Rick (a fold cycle with no way out) | given `Dick`, `Ricky`, `Rick`, in that order; no panic |
| AC-1.8 | Given Joseph, José, and `given=Joseph|given=Jose` | given `Joseph`, `Jose` |
| AC-1.9 | A Subject with given Ann, one with Anna, one with Hannah, promoted and joined into one person, then the cache rebuilt | The joins succeed; the cached name is `given=Hannah|surname=Robins`; upkeep equals rebuild |

**US-2.** As a contributor, the reconciler's fold has no assumption that random lists of nicknames, initials, prefixes and accented forms can break, and the cache rebuilds once for the new resolution.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | 2,000 seeded random lists of up to 8 candidates over a pool of nicknames, initials, prefixes and accented forms, one to three words a unit, under product rules | No panic; each structured positive candidate is in exactly one value; shuffling the input changes nothing |
| AC-2.2 | Any list under a test alias table with a chain and a cycle, after `foldValues` | Every value with `into == nil` has a non-nil `total`; every `into` points at a value whose own `into` is nil |
| AC-2.3 | A catalog stamped `CacheVersion` 21 | `NeedsRebuild` is true; `autoreconciler.CacheVersion` is 22; the product alias pin holds version 22 with the hash unchanged |
| AC-2.4 | Pairs that fold directly today (Bill + William, Des + Desiree, Allan + Allen) | Same result as at the planning base (`TestValueRulesFoldNicknamesOnly` unchanged and green) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `foldValues` resolves each value through the fold graph to the best supported, then best ranked, value with no fold of its own that it can reach; a value that reaches none is its own root (audit 2 B-1) | STEP-1, AC-1.1 – AC-1.8, AC-2.2, D-1, D-2 |
| Pure tests: a chain-and-cycle alias table, fold-chain rows, a `foldValues` roots test, and `TestReconcileNamesInvariants` widened to nicknames, initials, prefixes and accented forms under that table | STEP-2, AC-2.2, D-4 |
| Product-rule tests: fold-chain rows and a product-rule invariants test in `core/database/autoreconciler`; a DB subtest that files and rebuilds Ann, Anna, Hannah | STEP-3, AC-1.1 – AC-1.9, AC-2.1, D-4 |
| `CacheVersion` 22 and its pin | STEP-4, AC-2.3, D-3 |
| `conclusion-reconciliation.md` says a fold follows chains, and what a cycle does | STEP-5 |
| Patch version bump | STEP-6 |

| Out | Home |
| --- | --- |
| Audit 2 B-3: changing which name a fold keeps (`AliasIndex.Expands`, heads) | not planned: the user, 2026-11-01, "a nickname should be folded into the proper name" when both are recorded; map working rule updated |
| Adding pairs to `nicknameDirections` or editing the seed | not planned in this PR; the user adds Directions as dogfood finds them |
| The FFI panic recover | S14-13 |
| Accented alias lookups | [promote-ranking idea](../../../ideas/promote-ranking.md) (descoped with S12-46) |
| Showing a folded nickname as a `nick` part | S14-26 |
| Collapsing a fold cycle into one of its members | not planned (D-2) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Fold | `core/autoreconcile/pipeline.go` `foldValues` | Marks a value maximal when it folds into no other value, then points each value at the best supported, then best ranked, maximal value it folds into **directly**. Gives `total` only to maximal values. A value whose only folds go to non-maximal values keeps `into == nil` without being maximal, so `total` stays nil and `r.total[s] = true` panics. A cycle has no maximal value at all |
| Root | `core/autoreconcile/pipeline.go` `root` | Follows `into` until nil. Today one hop at most |
| Consumers of the fold | `pipeline.go` `reconcile` (step 3 – 6), `elect`, `dropWeak`, `settledUnits` | Read `into == nil` as "a root with a `total`", and `root(v)` |
| Module fold rules | `core/autoreconcile/names.go` `nameModule.fold`; `core/valuecompare/aliases.go` `AliasIndex.Expands` | Initial and contains folds chain; alias (S11-28) and prefix (S11-45) folds do not. Unchanged by this PR |
| Other modules | `core/autoreconcile/modules.go`, `dates.go` | Dates, text, terms, subjects and depictions run the same `foldValues` |
| Callers | `core/database/autoreconciler/autoreconciler.go` (upkeep and `Rebuild`, line 345), `truth.go` line 46, `core/database/promotealign/profiles.go` (through `subjectprofile.Build`) | Every one passes `autoreconciler.ValueRules` and panics on the B-1 shapes |
| Cache version | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` | 21, with a history comment, one line per version |
| Pin | `core/database/autoreconciler/autoreconciler_test.go` `productAliasPin`, `TestProductAliasTablesPinCacheVersion` | `{CacheVersion: 21, Hash: "0dee1d25…"}` |
| Pure tests | `core/autoreconcile/names_test.go` `TestReconcileNamesInvariants`, `aliasRules`, `TestReconcileNamesAliases`, `checkNames`; `pipeline_test.go` `TestPipelineInvariants` | Invariants run under `testRules` only, which has no alias table, so no alias fold is ever generated |
| Product-rule tests | `core/database/autoreconciler/autoreconciler_test.go` `TestValueRulesFoldNicknamesOnly` (two-name pairs), `TestReconciledNames` (DB, `f.promote`, `f.join`, `assertUpkeepEqualsRebuild`) | No three-name chain |
| Docs | `docs/conclusion-reconciliation.md` line 89 (§5 step 3) | "When a value fits several, it folds into the best supported, then the strongest." Says nothing on chains or cycles |

**Precedent.** `ce33e3a2` ("Fold a given name into the name it starts in the reconciler, never a surname.", S11-45): a reconciler fold change, its tests, the `CacheVersion` line and the pin bump in one commit. Copy that shape for STEP-1 to STEP-4. Do not copy its `names.go` change; this PR does not touch which pairs fold.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `foldValues` computes the fold relation once per unit as an n × n bool matrix (`m.fold(name, vals[a].u, vals[b].u)` for a ≠ b). A **sink** is a value with no fold. For each value, a depth-first walk over the matrix collects every sink it reaches; its target is the best of those: most `sources`, then lowest `rank`. A value with a target gets `into = target` (one hop, flattened); a value that reaches no sink keeps `into == nil`. Every value with `into == nil` gets `total`. `root` is unchanged | Matches audit 2's fix and the existing tie-break. `into` stays one hop, so `elect`, `dropWeak`, `settledUnits` and `root` keep their reading. `rank` is unique per value, so the target is a total order and input order cannot change it. The matrix makes the n² `m.fold` calls once, the same count as today | Following `into` chains as stored: a chain may end at a non-sink. Stopping at the first sink found: depends on walk order. Running `m.fold` inside the walk: n³ value-compare calls |
| D-2 | A value that reaches no sink (only a cycle with no way out: Dick, Ricky, Rick) is its own root. The cycle's members stay distinct values and the name shows each (AC-1.7) | A cycle gives no formal name, so the reconciler shows what was recorded, as it does for Jake + James. The same values reach Richard as soon as a Richard record is joined, because Dick → Richard is a Direction | Collapsing the cycle into its best-supported member: invents a formal name the table does not give |
| D-3 | `CacheVersion` 22, line `22: a fold follows chains to the formal value it reaches; a fold cycle keeps each value (S14-18).` `productAliasPin.CacheVersion` 22; the hash is unchanged | Resolution changes stored rows: a value with a direct target may now reach a better-supported sink through a chain. A project that crashed never stored a newer version, so it rebuilds either way | No bump: a cached row could disagree with upkeep until the next edit |
| D-4 | Product-rule tests live in `core/database/autoreconciler` (it already reconciles under `autoreconciler.ValueRules`). `core/autoreconcile` tests use a test table `chainRules` *(new)* in the shape of the product seed | `core/autoreconcile` is a pure layer; its tests do not import `core/database` (`add-match-rule` "Layers are pure"). The map's "`TestReconcileNamesInvariants` runs with the product alias tables" is met by the two together | Importing `namealiases` from `core/autoreconcile` tests: a catalog import into a pure layer's tests |
| D-5 | No change to `names.go`, `aliases.go`, `nicknames_seed.go` or `nickname_directions.go` | Working rule "A nickname folds into its set's formal name, as today" (the user, 2026-11-01) | Audit 2 B-3 option a: stopped 696 of 2,627 product folds, Jim → James among them |

**Rules this plan bends.** None. The B-3 working rule was replaced in the map by the user's answer (2026-11-01) in this planning commit.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `chainRules` *(new)* | test helper | `core/autoreconcile/names_test.go` `aliasRules` | `aliasRules` backs AC-1.x rows of `TestReconcileNamesAliases`, which pin results with no chain; adding chain sets there would change what those rows exercise. `chainRules` copies its shape |
| `TestReconcileNamesFoldChains` *(new)* | test | `TestReconcileNamesAliases` (`folds` table + `checkNames`) | Same table shape, own rules |
| `TestFoldValuesRoots` *(new)* | test | `pipeline_test.go` `TestPipelineInvariants` | That test checks results, not the fold's internal roots |
| `TestValueRulesFoldChains` *(new)* | test | `autoreconciler_test.go` `TestValueRulesFoldNicknamesOnly` | That table is two-name pairs with a one-given-name expectation; chains need a list of names and an expected given list |
| `TestValueRulesNamesInvariants` *(new)* | test | `TestReconcileNamesInvariants` | Same generator idea under product rules, in the package that may import them (D-4) |

No production type, function or file is added; `foldValues` is rewritten in place.

**Builds on.** The pipeline's existing tie-break (sources, then rank); `add-match-rule`'s `CacheVersion` rule; precedent `ce33e3a2`.
**Deletes.** The `maximal` slice and the "maximal value it folds into directly" loop inside `foldValues`, replaced by the matrix walk. Nothing else becomes dead.
**Next adds.** S14-26 keeps a folded nickname as a `nick` part: it reads each value's root, which is now always one hop and always has a `total`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A unit with several values that each have `into == nil` and are not sinks (a cycle) | `elect`, `dropWeak`, `groupAll` / `groupDisplayed`, `settledUnits` | Each treats them as ordinary roots: they vote, can be outvoted as spelling variants, and assemble side by side (AC-1.7) |
| A value whose target is two or more folds away | `reconcile` step 6 (`root(v)`), outcomes `folded` | One hop after flattening; outcome `folded` as for a direct fold |

### Scale

At the map's Scale (92,000 Observations; the heaviest person in `BenchmarkReconcileNames` has 200 name candidates): `m.fold` runs n² times per unit, as today; the walk adds n³ bool steps, n being the distinct values of one unit of one person. Probe at the planning base: `BenchmarkReconcileNames` 3.0 – 3.2 ms/op at `main`, 3.2 – 3.4 ms/op with D-1, allocations within 1%. No write, key or cache changes; the bump rebuilds every catalog's reconciler table once on open, as every `CacheVersion` bump does.

### Edges

- **Assumptions.** Old: the fold relation is transitive and has a maximal element on every path. New: none; `into` is one hop and every `into == nil` value has a `total` (T-3). `rank` is unique per value, so the target is order-free (T-4 shuffles).
- **Neighbors.** B., Bill → Bill (no chain: unchanged). Bill, William → William and the other `TestValueRulesFoldNicknamesOnly` rows: unchanged (AC-2.4). Dates, text and terms fold transitively, so reachable sinks are the direct targets and results are unchanged: `TestPipeline` and the `dates_test.go` suites stay green unchanged (STEP-1 check).
- **Removed inputs.** Drop William from B., Bill, William → given `Bill` (B. folds into Bill). Drop Bill → given `B.`, `William`: B. is not an initial of William, so the chain is gone and both show (T-1 rows). Add Richard to Dick, Ricky, Rick → all four reach Richard (T-1 row).
- **Real data.** The product alias index, given role: 2,247 tokens, 2,627 fold edges, 671 alias chains a → b → c where a does not fold into c, two 3-cycles (`rich → ricky → rick`, `dick → ricky → rick`), no 2-cycles. Initial and prefix folds add more chains. Eight everyday shapes panic at the planning base (AC-1.1 – AC-1.8).
- **Untrusted size.** Values come from the catalog's Observations, not from a table the app ships, but n is the distinct values of one unit of one person. The matrix is n² bools and the walk n³ steps; n is already the bound on today's n² `m.fold` calls, which cost far more per step. No new bound.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/autoreconcile` | `foldValues(m module, name string, vals []*unitValue)` rewritten per D-1 and D-2; doc comment says it follows chains and what a value that reaches no sink does. `root` unchanged |
| `core/database/autoreconciler` | `CacheVersion = 22` with its history line (D-3) |

Sketch of the shape (not the implementation):

```go
// edge[a][b]: vals[a] folds into vals[b]; sink[a]: vals[a] folds into nothing.
// For each a: walk edge from a; among the sinks reached, keep the most
// sources, then the lowest rank; vals[a].into = that sink, or nil.
```

Errors: none.

### FFI

None — no wire change.

### macOS

None — no app change.

### Strings

None — no copy changes.

## 7. Steps

**STEP-1 — Resolve folds through chains**
Files: `core/autoreconcile/pipeline.go`. Skill: `add-match-rule`.
Do: rewrite `foldValues` per D-1 and D-2. Keep the signature and `root`. Update the doc comment above `foldValues`: "foldValues points each value at the best supported, then best ranked, value it reaches through folds that folds into nothing, and totals each root's Sources and strength. A value that reaches none, as in a fold cycle, is its own root."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/autoreconcile/ ./core/subjectprofile/ ./core/database/autoreconciler/` passes with no test edited.

**STEP-2 — Pure tests**
Files: `core/autoreconcile/names_test.go`, `core/autoreconcile/pipeline_test.go`. Rule: `go-tests.mdc`.
Do: add `chainRules` beside `aliasRules`: `testRules` plus one `FormalFirst` nickname table, roles given / nick / untyped, sets `{william, bill}`, `{richard, dick}`, `{ricky, dick}`, `{rick, ricky}`, `{dick, rick}`, `{hannah, anna}`, `{anna, ann}`. Add T-1, T-2, T-3. Widen `TestReconcileNamesInvariants` (T-4).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestReconcileNames|TestFoldValuesRoots' ./core/autoreconcile/` passes; with STEP-1 reverted (`git stash`), T-1 panics. Restore.

**STEP-3 — Product-rule tests**
Files: `core/database/autoreconciler/autoreconciler_test.go`. Rule: `go-tests.mdc`.
Do: add T-5, T-6 and the T-7 subtest in `TestReconciledNames`, before its `assertUpkeepEqualsRebuild`.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestValueRules|TestReconciledNames' ./core/database/autoreconciler/` passes.

**STEP-4 — Cache version**
Files: `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/autoreconciler_test.go`. Skill: `add-match-rule`.
Do: per D-3. Add `21` to the stale list in `TestEnsureCatalogRebuildsStaleVersion` (T-8).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestProductAliasTablesPinCacheVersion|TestEnsureCatalogRebuildsStaleVersion' ./core/database/autoreconciler/` passes.

**STEP-5 — Domain doc**
Files: `docs/conclusion-reconciliation.md`.
Do: §5 step 3 (line 89), after "When a value fits several, it folds into the best supported, then the strongest.", add: "A fold follows chains: *B.* folds into *Bill* and *Bill* into *William*, so *B.* reaches *William*; each value folds into the best supported, then the strongest, value it reaches that folds into nothing. Values in a fold cycle with no way out (*Dick*, *Ricky*, *Rick*) each stay a value."
Check: `rg -n "fold cycle" docs/conclusion-reconciliation.md` hits the new sentence.

**STEP-6 — Version bump**
Files: `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (every `MARKETING_VERSION` line). Rule: `versioning.mdc`.
Do: product patch bump, all together.
Check: `VERSION`, `core.Version` and every `MARKETING_VERSION` line show the same new version.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (map row), this plan (status).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-18-name-folds-never-crash.md` reports no errors.

Commits: one per behavior. (1) "Resolve a name fold through its chain so a person recorded as B., Bill and William no longer crashes the app." (STEP-1 – STEP-4). (2) "Say in the reconciliation doc that a fold follows chains." (STEP-5). (3) "Bump the product version for the name fold crash fix." (STEP-6). (4) Docs (STEP-7).

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/names_test.go` `TestReconcileNamesFoldChains` *(new)* | Under `chainRules`, `checkNames`, `StateMerged`, one value each: B., Bill, William → `given=William|surname=Robins`; Ann, Anna, Hannah → `given=Hannah|…`; Dick, Ricky, Rick → `given=Dick|given=Ricky|given=Rick|…`; Dick, Ricky, Rick, Richard → `given=Richard|…`; B., Bill → `given=Bill|…`; B., William → `given=B.|given=William|…` | AC-1.1, AC-1.6, AC-1.7, D-1, D-2 |
| T-2 | Go | same file, `TestReconcileNamesFoldChains` reversed-order rows | Each T-1 list's candidates, same ids, passed in reverse slice order → the same result | D-1 (order-free) |
| T-3 | Go | `core/autoreconcile/pipeline_test.go` `TestFoldValuesRoots` *(new)* | Seeded random lists over the T-4 pool under `chainRules`; split with `moduleFor(chainRules, propertymeta.ValueTypeName).split`, grouped by unit name as `reconcile` does, then `foldValues` → every `into == nil` value has `total != nil`; every `into != nil` value's `into.into == nil` | AC-2.2 |
| T-4 | Go | `core/autoreconcile/names_test.go` `TestReconcileNamesInvariants` | Runs once under `testRules` and once under `chainRules`. The `given` pool adds B., Bill, William, D., Dick, Ricky, Rick, Richard, Ann, Anna, Hannah, Des, Desiree, José, Jose, Joseph; `nick` adds Bill, Dick. Existing invariants unchanged | AC-2.2, D-4 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestValueRulesFoldChains` *(new)* | `autoreconciler.ValueRules`, one candidate per name: the eight AC-1.1 – AC-1.8 lists → the given lists in §2 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8 |
| T-6 | Go | same file, `TestValueRulesNamesInvariants` *(new)* | Seeds 1 – 4 × 500 lists, up to 8 candidates, given units of one to three words from: B., Bill, William, D., Dick, Richard, Rick, Ricky, P., Peggy, Margaret, Mo, Molly, Mary, Caroline, Chick, Charles, Ann, Anna, Hannah, Joseph, José, Jose, Des, Desiree, J., Jim, James, Jake; surname Robins → no panic, each positive candidate in exactly one value, shuffled input gives a deeply equal result | AC-2.1 |
| T-7 | Go | same file, `TestReconciledNames` subtest "a chain of folds files and rebuilds" *(new)* | Three Subjects with given Ann, Anna, Hannah (surname Robins), `f.promote` the first and `f.join` the others → one name row, `given=Hannah|surname=Robins`; the test's closing `assertUpkeepEqualsRebuild` passes | AC-1.9 |
| T-8 | Go | same file, `TestEnsureCatalogRebuildsStaleVersion`, `TestProductAliasTablesPinCacheVersion` | Stale version 21 rebuilds; pin is version 22 with the unchanged hash | AC-2.3, D-3 |
| T-9 | Go | same file, `TestValueRulesFoldNicknamesOnly` (unchanged) | Every row as at the planning base | AC-2.4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `./core/autoreconcile/` (`TestPipeline`, date tests), `./core/database/autoreconciler/` (`TestRebuildEqualsUpkeep_*`), `./core/database/promotealign/`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): patch
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: none)
- [ ] Domain doc updated: docs/conclusion-reconciliation.md §5 step 3
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. On a project at the planning base, make a person from three Sources whose names read B. Robins, Bill Robins and William Robins (promote one, join the other two). Note the crash. On the PR build, open the same project: it opens (the cache rebuilds to version 22) and the person shows **William Robins**.
2. Quit and reopen: the person still shows William Robins.
3. Do the same with Ann, Anna and Hannah Robins: files, reopens, shows **Hannah Robins**.
4. A person with Bill and William (two records) still shows **William Robins**.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A date, text or term result changes because its fold was not transitive | STEP-1 check runs every existing suite unedited; a failure there is a stop condition |
| A cycle member outvotes another in `elect` as a spelling variant | Allowed: they are ordinary roots (Reach). T-1's Dick, Ricky, Rick row pins the result |
| The walk is slow on a person with many distinct values | Scale probe shows no change at 200 candidates |

**Stop and report if:**

- `foldValues`, `root`, `elect` or `dropWeak` no longer read as §4 says at your base.
- Any existing test in `./core/autoreconcile/`, `./core/subjectprofile/`, `./core/database/autoreconciler/` or `./core/database/promotealign/` fails after STEP-1 without a test edit.
- An AC-1.x result under product rules differs from §2. Report the output; do not change the expected value.
- `TestProductAliasTablesPinCacheVersion` reports a changed hash (something edited the seed).
- `BenchmarkReconcileNames` is more than 25% slower than at the planning base.

## 12. Open questions

None.

## 13. PR

**Title:** S14-18 — Name folds never crash
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-18-name-folds-never-crash.md`.
