# S11-28 — Fold a nickname into the given name it stands for

**Status:** Planned
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere (follow-up to S11-06)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the reconciler's fold rule changes in Go; pages render the cache as they do today)
**Board:** —
**Depends on:** S11-06 (landed [#456](https://github.com/mendahu/provenencia/pull/456))
**Blocks:** —
**Branch:** `s11-28-alias-fold`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`, `.cursor/rules/name-values.mdc`
**Context:** [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §7 (grouping policy), §7.2; [S11-06 plan](S11-06-migrate-reconciler.md); working rules "The reconciler keeps its own decisions", "High tolerance for churn"
**Planned:** 2026-10-19 against `main` @ `84a9b31`
**Review:** [`S11-28-alias-fold.review.md`](S11-28-alias-fold.review.md) — Fits with changes / Not ready → fixed, 1 round

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

A Person whose records say *Chris Robins* and *Christopher Robins* shows one given name, *Christopher*, on the Persons list and the Person page. Today the reconciler keeps both (`given: [Christopher, Chris]`) because S11-06 decided that an alias never folds. The user asked (2026-10-19) that the product nickname and abbreviation tables fold a nickname into the formal name it stands for, the way an initial folds into the name it begins. The direction comes from the data: each product set lists its formal name first, and this PR curates the 144 pairs the upstream list records both ways (user, 2026-10-19).

**Why now.** Dogfood: every Person recorded once by a nickname and once by a formal name reads as two given names.

**What stays.** Accent folds never fold (*Jose* stays beside *José*). Majority still only outvotes spelling variants. Surnames never alias-fold. Promote scoring is unchanged: aliases already score there.

## 2. User stories

**US-1.** As a researcher, when one record calls a person *Chris* and another *Christopher*, I see one given name, *Christopher*, so that the title reads as one name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Candidates `given=Chris\|surname=Robins` and `given=Christopher\|surname=Robins` | One value `given=Christopher\|surname=Robins`, both members, state merged |
| AC-1.2 | `given=Wm\|surname=Robins` and `given=William\|surname=Robins` | One value `given=William\|surname=Robins`, members {1,2}, merged |
| AC-1.3 | `given=Chris\|given=R\|surname=Robins` and `given=Christopher\|given=Robert\|surname=Robins` | One value `given=Christopher\|given=Robert\|surname=Robins`, members {1,2}, merged |
| AC-1.4 | Records 1 `given=Christopher\|surname=Robins`, 2 and 3 `given=Christina\|surname=Robins`, 4 `given=Chris\|surname=Robins` | Chris folds into Christina, the best supported: one value, members {1,2,3,4}, `given=Christina\|given=Christopher\|surname=Robins`, merged (the ambiguous-initial rule at `names_test.go` "an ambiguous initial folds into the best supported", applied to an alias) |
| AC-1.5 | `given=Jose\|surname=Robins` and `given=José\|given=Kenneth\|surname=Robins` | Unchanged: the same result as without aliases (an accent fold never folds) |
| AC-1.6 | `nick=Chris\|surname=Robins` and `given=Christopher\|surname=Robins` | Unchanged: the same result as without aliases (a fold never crosses part types) |
| AC-1.7 | `given=Ann\|surname=Bill` and `given=Ann\|surname=William` (an alias set holds both) | Unchanged: the same result as without aliases; a surname never alias-folds |
| AC-1.8 | Two nicknames of one formal name: `given=Chris\|surname=Robins` and `given=Kit\|surname=Robins` (set `{christopher, chris, kit}`) | Unchanged: the same result as without aliases; neither is the other's formal name |
| AC-1.9 | A nickname longer than its formal name: `given=Ebbie\|surname=Robins` and `given=Abel\|surname=Robins` (set `{abel, ebbie}`) | One value `given=Abel\|surname=Robins`, members {1,2}, merged |
| AC-1.10 | A pair listed both ways: `given=Allan\|surname=Robins` and `given=Allen\|surname=Robins` (sets `{allan, allen}` and `{allen, allan}`) | Unchanged: the same result as without aliases |

**US-2.** As the engine, the reconciler cache reads the product alias tables, so that the fold happens on every catalog and the cache changes only when `CacheVersion` does.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A catalog with two Observations naming the same person *Chris Robins* and *Christopher Robins*, joined in one entity | `autoreconciler` writes one displayed name `Christopher Robins` |
| AC-2.2 | `autoreconciler.CacheVersion` | `18`; a cache written at 17 rebuilds on open |
| AC-2.4 | The curated product seed, under `autoreconciler.ValueRules` | *Bill* + *William* → *William*; *Jamie* + *James* → *James*; *Tony* + *Anthony* → *Anthony*; *Allan* + *Allen* → both stay |
| AC-2.5 | The product nickname seed after curation | The only pairs listed both ways are Appendix A's "Kept two-way" list; no set holds only its formal name |
| AC-2.3 | `core/database/promotealign/profiles.go` builds its profiler | `reconcile` is `autoreconciler.ValueRules`, the cache's value, so a future `Options.Excluding` re-reconcile folds as the cache does (no caller passes `Excluding` today: `profiles.go:269` passes `Options{}`) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `valuecompare.AliasTable.FormalFirst` and `AliasIndex.Expands`; `nameModule.fold` accepts an alias or abbreviation pair whose `into` word is the formal name of a set holding the `from` word | STEP-1, AC-1.1 – AC-1.10 |
| The product nickname seed curated per Appendix A; both product tables marked `FormalFirst` | STEP-2, AC-2.4, AC-2.5 |
| `autoreconciler.ValueRules` (exported) carries the product alias tables; `CacheVersion` 18; `promotealign` profiles read it | STEP-2, AC-2.1, AC-2.2, AC-2.3 |
| `Compatible` follows the new fold (it calls `fold`); its alias test flips | STEP-1, T-3 |
| Docs: grouping policy and §7.2 fold rule, `names.go` header, `subjectprofile.Build` comment, `add-match-rule` skill | STEP-3 |

| Out | Home |
| --- | --- |
| Storing the formal name in the catalog (`name_alias_members` keeps no order), so catalog-read alias rows can fold | not planned; matching does not need a direction, and the reconciler reads the product seed |
| Re-curating existing catalogs' nickname rows | not planned; their matching is unchanged (every pair still shares the formal name's set) |
| The reconciler reading a catalog's own alias rows (an alias editor) | not planned; no alias editor exists |
| Nickname quotes, former surnames, italics on the title | S11-29 |
| Folding across part types (`nick=Chris` into `given=Christopher`) | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Fold rule | `core/autoreconcile/names.go` `nameModule.fold` | Folds only when `valuecompare.Compare(from, into)` is `Contains` with `Broader == SideA` and every pair is `Same` or cause `initial`. `Chris` vs `Christopher` is `Equivalent` with one pair `Equivalent`/`alias`, so it never folds. The type argument is ignored (`_ string`) |
| Unit compare value | `core/autoreconcile/names.go` `nameUnit.compareValue` | Strips part types, so every word is untyped (role `untyped`); the product alias tables allow `untyped`, so alias pairs are found for every part type including surnames |
| Header comment | `core/autoreconcile/names.go` package comment, "Fold" bullet | "An accent fold or an alias never folds." |
| Compatible | `core/autoreconcile/modules.go` `Compatible` | Two values are compatible when every shared unit has the same key or one folds into the other. Called from `core/database/promotealign/exhibits.go` |
| Reconciler rules | `core/database/autoreconciler/autoreconciler.go` `valueRules` | `matchregistry.Default().Value`: no alias tables, by design, "so the cache changes only when CacheVersion does" |
| Cache version | `core/database/autoreconciler/autoreconciler.go` `CacheVersion` | 17, with a changelog comment above it |
| Promote profiles | `core/database/promotealign/profiles.go` profiler constructor (`reconcile: matchregistry.Default().Value`, line 169); `core/subjectprofile/subjectprofile.go` `Build` doc comment ("pass matchregistry.Default().Value, as the cache does") | A second copy of the cache's rules. `promotealign` already imports `autoreconciler` (checked with `go list -deps`) |
| Skill text | `.agents/skills/add-match-rule/SKILL.md`, bullets "The profile builder" and "The reconciler composes value compare" | Say the cache and `Build` callers pass `matchregistry.Default().Value` |
| Product alias seed | `core/database/namealiases/registry.go` `productAliasTables`, `AliasTables`; `nicknames_seed.go` `nicknameSets` | The abbreviations and nicknames tables, roles given / nick / untyped. Every set lists its formal name first (abbreviations: `{james, jas}`; nicknames: `{christopher, chris, kit}`, `{abel, ab, abe, eb, ebbie}`). 1,232 nickname sets; 144 pairs are listed both ways (`{bill, …, william, …}` and `{william, …, bill, …}`) |
| Alias index | `core/valuecompare/aliases.go` `AliasTable`, `AliasIndex`, `NewAliasIndex`, `Hit` | Sets are unordered token maps; nothing records a formal name. Catalog `Load` returns tokens ordered alphabetically |
| Western roles | `core/matchregistry/value.go` `Names.Western.PartRoles` | `given` → given, `nick` → nick, `surname` → family, `undetermined` and `""` → untyped |
| Tests | `core/autoreconcile/names_test.go` `TestReconcileNames` (group "fold"), `TestReconcileNamesWords` ("an accent-folded pair never folds"); `modules_test.go` `TestCompatibleIgnoresAliases`; `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledNames` | `testRules` (in `autoreconcile_test.go`) has no aliases. `TestCompatibleIgnoresAliases` asserts Jim Robins is not compatible with James Kenneth Robins |

Value compare's output for the cases, checked at the planning base with the product rules plus a nickname table: `Chris` / `Christopher` → `equivalent`, one pair `equivalent alias`; `Chris, R` / `Christopher, Robert` → `contains`, `Broader = a`, pairs `alias` and `initial`; `Wm` / `William` → `equivalent`, pair cause `abbreviation`; `Jose` / `José` → `equivalent`, pair cause `fold`.

**Precedent.** S11-06 ([#456](https://github.com/mendahu/provenencia/pull/456), commits `fc01031a`, `1b407323`) changed the same fold and outvote rules: copy its table-driven cases in `names_test.go` and its `CacheVersion` changelog line. Do not copy its `outvotes` change; outvoting is not touched.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A unit folds into another when `Compare(from, into)` is `Contains` with `Broader == SideA`, or `Equivalent`; every pair is `Same`, cause `initial`, or an alias pair (cause `alias` or `abbreviation`) for which `Aliases.Expands(from word, into word, role)` holds; and at least one pair is not `Same` | The data says which name is formal (user, 2026-10-19: "we control that data"); `Equivalent` covers a unit with no unpaired words, `Contains` one with extra words | Shorter folds into longer — wrong for *Abel* / *Ebbie*, *James* / *Jamie*, *Michael* / *Mike* |
| D-1a | `valuecompare.AliasTable.FormalFirst bool` *(new)*: each set's first token is the formal name the others stand for. `NewAliasIndex` records it per set (`aliasSet.formal`, the first token's `AliasToken`; empty when the table is not `FormalFirst`). `AliasIndex.Expands(from, into string, role NameRole) bool` *(new)*: true when a set that allows `role` has `formal == into` and holds `from`, and no set that allows `role` has `formal == from` and holds `into`. Both product tables set `FormalFirst: true`; catalog `Load` leaves it false | One explicit flag instead of a positional assumption every reader must know; a pair listed both ways never folds | A per-set `Formal` field — every set would repeat its first token; deriving it in `namealiases` — the index is where tokens are keyed |
| D-1b | Curate `nicknameSets` per Appendix A: for each directed pair `formal ← nickname`, remove the formal name from every set headed by the nickname; then delete any set left with only its head (50 sets). The 36 "Kept two-way" pairs stay as they are. The `nicknames_seed.go` header gains: "Curated (S11-28): a pair listed both ways upstream keeps one direction, the formal name first; spelling variants and translations stay both ways." | User call (2026-10-19): curate in the PR; the user reviews Appendix A on the PR. Matching is unchanged: every directed pair still shares the formal name's set | Leaving two-way pairs to never fold — *Bill* / *William* would not fold; a length rule for them — wrong for *James* / *Jamie* |
| D-2 | Two nicknames of one formal name (*Chris* / *Kit*) and a pair listed both ways (*Allan* / *Allen*) do not fold | Neither is the other's formal name; both stay, as today | Alphabetical tie-break — arbitrary |
| D-3 | An alias pair folds only when the catalog's alias data allows the unit's role: `role := m.rules.Names.Western.Role(typ)` and `m.rules.Names.Aliases.Hit(p.A.Text, p.B.Text, role, role)` must hit. The product tables allow given, nick and untyped, so `given`, `nick`, `undetermined` and untyped parts alias-fold; `surname` (family) and the others never do | Which roles may use an alias table is already data (`AliasTable.Roles`, checked by `AliasIndex.Hit`); the module names no role. The reconciler strips roles before comparing, so the role is re-read here | Hard-coding given / nick in the module — a second rule beside the table roles (review M-1); indexing `PartRoles` directly — `NamePattern.Role` is the one lookup (M-6) |
| D-4 | Both nickname and abbreviation tables fold | User call (2026-10-19): "Just Christopher", nicknames and abbreviations both fold | Nicknames only |
| D-5 | The folded name disappears from the displayed value; it is not turned into a `nick` part | User call (2026-10-19); matches how an initial folds | Showing *Christopher "Chris"* — a name no source wrote |
| D-6 | `valuerules.Product() valuecompare.Rules` *(new)* = the shipped registry plus the **product seed** alias tables (`namealiases.ProductTables()` *(new)*, a copy of `productAliasTables`, not `AliasTables()`, which includes `Register`ed plugin tables). `autoreconciler.valueRules` is renamed `ValueRules` and set to `valuerules.Product()`; `promotealign/profiles.go` passes `autoreconciler.ValueRules` as `reconcile`, the one value every reconciler caller uses. A guard test pins the product tables to `CacheVersion` (T-7) | Keeps the cache's rule that it changes only when `CacheVersion` does: a plugin table cannot move it, and a seed edit fails a test until the version is bumped. `valuerules` is the declared one place `Rules` are assembled. Promote's `Excluding` re-reconcile agrees with the cache | Reading the catalog's alias rows — stale cache on an alias change; `AliasTables()` — plugin tables change the cache silently (M-2); building the rules in `namealiases` — `valuerules` owns assembly (M-5) |
| D-7 | `Compatible` changes with the fold under rules whose tables are `FormalFirst`; `TestCompatibleIgnoresAliases` becomes `TestCompatibleFoldsAliases`, reads `aliasRules`, asserts Jim Robins is compatible with James Kenneth Robins, keeps its whole-name `Compare` == `Contains` assertion, and adds Jose Robins vs José Robins → not compatible. Promote exhibits that count through `Compatible` may now count an alias pair as agreeing | `Compatible` is defined as "the same value as the pipeline would see it". Promote's exhibits (deleted in S11-19) call it with the catalog's alias rows (`valuerules.Load`), which are not `FormalFirst`, so exhibits do not alias-fold and are unchanged | Keeping `Compatible` alias-blind through a flag — a second fold rule |
| D-8 | `CacheVersion` 18 with the changelog line `18: a nickname or abbreviation of a given name folds into its formal name; the reconciler reads the curated product alias seed (S11-28).` | Cached values change | — |
| D-9 | No new registry knob | The rule has no number: the alias data and the cause codes decide | A minimum length — nothing needs tuning |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `valuerules.Product()` *(new)* | func | `valuerules.Load` (catalog rules) | Same assembly from the product seed instead of catalog rows; sits beside `Load` |
| `namealiases.ProductTables()` *(new)* | func | `namealiases.AliasTables()` (seed + registered) | Product seed only (D-6) |
| `TestProductAliasTablesPinCacheVersion` *(new)* | test | `TestKnobInventoryMatchesRegistry` (a drift guard) | Fails when the product seed changes without a `CacheVersion` bump |
| `AliasTable.FormalFirst`, `aliasSet.formal`, `AliasIndex.Expands` *(new)* | field, method | `AliasIndex.Hit` | `Hit` answers "same set"; `Expands` adds the direction, on the same index |
| `TestNicknameSetsHaveOneDirection` *(new, `core/database/namealiases/registry_test.go`)* | test | `TestSharedNicknameDoesNotMergeCanonicalNames` | Guards Appendix A: the only two-way pairs are the kept list (`keptTwoWay`, a test var) and no set is a lone head |
| `aliasRules` *(new, test var in `names_test.go`)* | helper | the inline index in `TestCompatibleIgnoresAliases` | Moves that table to one var both tests use |
| `autoreconciler.ValueRules` *(exported rename of `valueRules`)* | var | `valueRules`; `promotealign` `matchregistry.Default().Value` | One value for both reconciler callers |

**Builds on.** S11-06's fold in `nameModule`; value compare's `Detail.Parts` causes; `AliasIndex` and the tables' `Roles`; the seed's formal-first convention; `NamePattern.Role`; `valuerules` as the one place rules are assembled.
**Deletes.** 108 backwards tokens and 50 lone-head sets in `nicknameSets` (Appendix A). The "an accent fold or an alias never folds" sentence in `names.go` and in §7 and §7.2 of `conclusion-reconciliation.md`; replaced in STEP-3. The map's working rule was amended at planning. The inline alias table in `TestCompatibleIgnoresAliases`.
**Next adds.** An alias editor that stores order (or a formal flag) in the catalog sets `FormalFirst` on loaded tables; nothing else changes.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` (`aliases.go`) | D-1a: `FormalFirst` on `AliasTable`; `formal` on `aliasSet`; `func (a AliasIndex) Expands(from, into string, role NameRole) bool` *(new)*. `aliases_test.go` / `words_test.go`: T-8 |
| `core/autoreconcile` (`names.go`) | `fold(typ string, from, into unit) bool` reads `typ`. Sketch below. Header comment "Fold" bullet rewritten (STEP-3 text) |
| `core/database/namealiases` (`registry.go`, `nicknames_seed.go`) | Both product tables `FormalFirst: true`; `nicknameSets` curated per D-1b. `func ProductTables() []valuecompare.AliasTable` *(new)*: a copy of `productAliasTables`. Comments that say scoring never reads the in-code list (package doc in `namealiases.go`; `productAliasTables`, `aliasTables`, `AliasTables` in `registry.go`) now say: "Matching reads the catalog rows; the reconciler cache reads the product seed through `valuerules.Product`." |
| `core/database/valuerules` (`valuerules.go`) | `func Product() valuecompare.Rules` *(new)*: `matchregistry.Default().Value` with `Names.Aliases = valuecompare.NewAliasIndex(namealiases.ProductTables(), rules.Words)`; doc: "the shipped registry and the product alias seed, never a catalog's own rows or name formats; the reconciler cache reads it" |
| `core/database/promotealign` (`profiles.go`) | `reconcile: autoreconciler.ValueRules`; drop the `matchregistry` import if it becomes unused |
| `core/database/autoreconciler` (`autoreconciler.go`, `truth.go`) | `var ValueRules = valuerules.Product()` (renamed from `valueRules`; both call sites updated); `CacheVersion = 18` with D-8's line. `autoreconciler_test.go`: `TestProductAliasTablesPinCacheVersion` *(new)*: a pinned pair `{CacheVersion: 18, Hash: "<hex>"}`: the test fails when the SHA-256 of the sorted product tables (key, kind, roles, sorted sets) differs from `Hash`, or when `CacheVersion` differs from the pinned version, so re-pinning means writing the new version beside the new hash; on a hash mismatch the message says "the product alias seed changed: bump autoreconciler.CacheVersion and re-pin this hash" |

```go
func (m nameModule) fold(typ string, from, into unit) bool {
	res, det := valuecompare.Compare(from.data.(nameUnit).cmp, into.data.(nameUnit).cmp, m.rules)
	switch {
	case res.Relation == valuecompare.Contains && det.Broader == valuecompare.SideA:
	case res.Relation == valuecompare.Equivalent:
	default:
		return false
	}
	role := m.rules.Names.Western.Role(typ)
	changed := false
	for _, p := range det.Parts {
		switch {
		case p.Relation == valuecompare.Same:
		case p.Cause == valuecompare.CauseInitial:
			changed = true
		case (p.Cause == valuecompare.CauseAlias || p.Cause == valuecompare.CauseAbbreviation) &&
			m.rules.Names.Aliases.Expands(p.A.Text, p.B.Text, role):
			changed = true
		default:
			return false
		}
	}
	return changed || res.Relation == valuecompare.Contains
}


// Expands: from stands for into. A set that allows role has into as its
// formal name and holds from, and no such set says the reverse.
func (a AliasIndex) Expands(from, into string, role NameRole) bool {
	forward, reverse := false, false
	for _, i := range a.byToken[from] {
		set := a.sets[i]
		if !set.roles[role] || !set.tokens[into] {
			continue
		}
		forward = forward || set.formal == into
		reverse = reverse || set.formal == from
	}
	return forward && !reverse
}
```

`typ` is the trimmed part type key (`readName` trims it); `Role` returns untyped for a type the pattern does not list. `p.A` is the `from` side because `from` is the first argument to `Compare`.

### FFI

None — the cache shape is unchanged.

### macOS

None — pages render the cache.

### Strings

None —.

## 7. Steps

**STEP-1 — Fold aliases of given names into their formal names**
Files: `core/valuecompare/aliases.go`, `core/valuecompare/words_test.go`, `core/autoreconcile/names.go`, `core/autoreconcile/names_test.go`, `core/autoreconcile/modules_test.go`. Skill: `add-match-rule` (no knob; read it for the layer rules).
Do: implement D-1, D-1a, D-2, D-3 per the sketch; add T-8. Add `aliasRules` (product `matchregistry.Default().Value` with a `NewAliasIndex` over two `FormalFirst` tables: nickname `{christopher, chris, kit}`, `{christina, chris}`, `{william, bill}`, `{james, jim}`, `{abel, ebbie}`, `{allan, allen}`, `{allen, allan}`; abbreviation `{william, wm}`; roles given / nick / untyped). Add `TestReconcileNamesAliases` *(new)* with T-1's rows, in the shape of `TestReconcileNames` (`in`, `want []cl`, `state`, `reasons`). Rename `TestCompatibleIgnoresAliases` to `TestCompatibleFoldsAliases` and flip it per D-7, reading `aliasRules`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/valuecompare/ ./core/autoreconcile/` passes, including the untouched `TestReconcileNamesWords` row "an accent-folded pair never folds".

**STEP-2 — The reconciler cache reads the curated product alias tables**
Files: `core/database/namealiases/nicknames_seed.go`, `core/database/namealiases/registry_test.go`, `core/database/namealiases/registry.go`, `core/database/namealiases/namealiases.go` (comment), `core/database/valuerules/valuerules.go`, `core/database/autoreconciler/autoreconciler.go`, `core/database/autoreconciler/truth.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/promotealign/profiles.go`.
Do: curate `nicknameSets` per D-1b and Appendix A (write a throwaway script; do not commit it); set `FormalFirst` on both product tables; add T-9 with `keptTwoWay` copied from Appendix A. Add `namealiases.ProductTables` and `valuerules.Product`; rename `valueRules` to `ValueRules` and set it; `profiles.go` passes it; `CacheVersion` 18 (D-6, D-8). Add T-4's subtest, T-7 and T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/ ./core/database/namealiases/ ./core/database/valuerules/ ./core/database/promotealign/ ./core/subjectprofile/` passes; `go build ./...` succeeds.

**STEP-3 — Docs**
Files: `core/autoreconcile/names.go` (header "Fold" bullet), `docs/conclusion-reconciliation.md` (line 132 grouping policy; line 167 §7.2 Fold bullet), `core/subjectprofile/subjectprofile.go` (`Build` comment: "pass autoreconciler.ValueRules, as the cache does"), `.agents/skills/add-match-rule/SKILL.md` (the two bullets "The profile builder" and "The reconciler composes value compare": the cache and `Build` callers pass `autoreconciler.ValueRules` = `valuerules.Product()`, the shipped registry plus the product alias seed, never catalog data; and the `valuerules.Load` line names `Product` as the reconciler's exception), `.agents/skills/add-seeded-vocabulary/SKILL.md` and `docs/seeded-vocabulary.md` §4.5 ("Scoring reads these rows and no in-code list" → "Matching reads these rows; the reconciler cache reads the product seed, so a seed edit bumps `CacheVersion`. Each product set lists its formal name first (`FormalFirst`); a pair the upstream list gives both ways was curated to one direction in S11-28, except spelling variants and translations"; the skill's "Scoring reads the catalog only" line gets the same change), `core/autoreconcile/names.go` "Format-agnostic" bullet and `docs/conclusion-reconciliation.md` §7.2 "Format-agnostic" bullet (add: "A type's role in the name format decides only whether an alias table applies (§7.2 Fold)."), `nameUnit.compareValue` comment (the role is re-read for the alias check), `docs/deployment-plan/spike-11/deployment-plan.md` (map row; the working rule was amended at planning), `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: in the `names.go` header and the §7.2 Fold bullet, replace "An accent fold or an alias never folds (`[Jose]` stays beside `[José, Kenneth]`)." with: "A nickname or abbreviation folds into the formal name it stands for, as the product alias seed lists it (*Chris* → *Christopher*, *Wm* → *William*, *Ebbie* → *Abel*; decided 2026-10-19, S11-28). Two nicknames of one name, or a pair the seed lists both ways, never fold. An alias applies only to the part roles its table allows; the product tables cover given names and nicknames, so a surname never alias-folds. An accent fold never folds (`[Jose]` stays beside `[José, Kenneth]`)." In the §7 grouping policy (line 132), replace "`Equivalent` (an accent fold, an alias) never groups or folds" with "`Equivalent` (an accent fold, an alias) never groups; it folds only through an alias of a given name (§7.2)". Then the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-28-alias-fold.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Fold a nickname into the given name it stands for.").

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/autoreconcile/names_test.go` `TestReconcileNamesAliases` *(new)* | Folding rows AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.9 with the members, parts and state in §2 and reasons `kept`; AC-1.4 also run with its input reversed, same result. "Unchanged" rows AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.10 assert that `Reconcile` under `aliasRules` equals `Reconcile` under `testRules` on the same input (values, members, state, reasons). AC-1.7 uses the `{william, bill}` set on surnames | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, D-1 – D-5 |
| T-3 | Go | `core/autoreconcile/modules_test.go` `TestCompatibleFoldsAliases` *(renamed)* | With `aliasRules`: Jim Robins vs James Kenneth Robins → compatible, and their whole-name `Compare` is still `Contains`; Jose Robins vs José Robins → not compatible | D-7 |
| T-6 | Go | `core/database/promotealign` (existing profile tests) | Pass unchanged with `reconcile: autoreconciler.ValueRules` | AC-2.3, D-6 (no behavior to observe until a caller passes `Excluding`; the check is the code reading `ValueRules`) |
| T-7 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestProductAliasTablesPinCacheVersion` *(new)* | The hash of `namealiases.ProductTables()` (key, kind, sorted roles, `FormalFirst`, each set as its first token then the rest sorted) equals the pinned constant | D-6, AC-2.2 |
| T-8 | Go | `core/valuecompare/words_test.go` `TestAliasIndexExpands` *(new)* | `FormalFirst` table `{christopher, chris, kit}`, `{abel, ebbie}`, `{allan, allen}`, `{allen, allan}`: `Expands(chris, christopher, given)` true; reverse false; `(chris, kit)` false; `(ebbie, abel)` true; `(allan, allen)` and reverse false; a role the table does not allow → false; the same table without `FormalFirst` → all false | D-1a, D-2 |
| T-9 | Go | `core/database/namealiases/registry_test.go` `TestNicknameSetsHaveOneDirection` *(new)* | The pairs listed both ways in `nicknameSets` equal `keptTwoWay` (Appendix A, 36 pairs); no set has fewer than two tokens; `william` is not in any set headed by `bill` | AC-2.5, D-1b |
| T-10 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestValueRulesFoldCuratedNicknames` *(new)* | `autoreconcile.Reconcile(ValueRules, name, …)` on AC-2.4's four pairs (surname Robins) → the formal name alone for Bill/William, Jamie/James, Tony/Anthony; both for Allan/Allen | AC-2.4 |
| T-4 | Go | `core/database/autoreconciler/autoreconciler_test.go` `TestReconciledNames` (new subtest "a nickname folds into the given name it stands for", shaped like "an initial expands into the full name") | `nameIn(a, f.props["name"], "given=Chris\|surname=Robins")` and `nameIn(b, f.props["name"], "given=Christopher\|surname=Robins")`, joined → one name row, `Support` 2, parts `given=Christopher\|surname=Robins` | AC-2.1, D-6 |
| T-5 | Go | `core/database/autoreconciler/autoreconciler_test.go` (existing check at the `StoredVersion(db)` call, line 1113 at the base) | Unchanged test; passes with `CacheVersion` 18 | AC-2.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestReconcileNamesInvariants` (runs under `testRules`, so alias folds are not in it; T-1's reversed AC-1.4 row covers order), `core/database/promotealign` tests that count exhibits through `Compatible`. A `promotealign` expectation may change only where the changed count comes from an alias pair now agreeing; list each in the PR.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; docs/deployment-plan/spike-11/completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: conclusion-reconciliation.md §7 and §7.2; seeded-vocabulary §4.5; subjectprofile.Build comment; add-match-rule and add-seeded-vocabulary skills; namealiases comments
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that names a person *Chris Robins* and another Source that names *Christopher Robins*; join both into one Person in Promote.
2. Open Persons (sorted style, Western format). The row reads *Robins, Christopher*, not *Robins, Christopher Chris*.
3. Open the Person. The title reads *Christopher Robins*; Why on the name shows both records counted toward it.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A curated direction is wrong (Appendix A) | The user reviews Appendix A on the plan PR before implementation; a fix is one token in the seed plus a `CacheVersion` bump (T-7) |
| Existing catalogs keep the uncurated rows | Matching treats sets as unordered, so their scores do not change; the reconciler reads the seed |
| Order dependence when one short name aliases two longer ones | AC-1.4 uses the pipeline's existing best-supported rule; `TestReconcileNamesInvariants` guards order |

**Stop and report if:**

- A pair in Appendix A is not listed both ways in `nicknameSets` at your base, or a two-way pair exists that Appendix A does not list.
- `namealiases` `AddProductNicknames` tests fail after the curation.

- `valuecompare.Compare` on `Chris` / `Christopher` does not return `Equivalent` with one pair of cause `alias` at your base.
- `TestReconcileNamesInvariants` fails after STEP-1 and the failure is not one of T-1's new expectations.
- Importing `namealiases` from `autoreconciler` creates an import cycle.
- A test outside `core/autoreconcile`, `core/database/autoreconciler`, `core/database/promotealign` and `core/subjectprofile` changes result.
- A `promotealign` test changes for a reason other than an alias pair now agreeing.

## 12. Open questions

None.

## 13. PR

**Title:** S11-28 — Fold a nickname into the given name it stands for
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-28-alias-fold.md`.

## Appendix A — Nickname directions

Generated at the planning base from `nicknameSets`: every pair `a`, `b` where a set headed by `a` holds `b` and a set headed by `b` holds `a`. Formal names were chosen by hand where the longer name is not the formal one (Jamie → James, Mike → Michael, Frankie → Frank, Annie → Ann / Anna, Debbie → Debra, Jessie → Jane / Janet, Harry → Henry). The user reviews this list on the plan PR.

144 two-way pairs at `84a9b31`: 108 get a direction, 36 stay two-way.

**Directed** — `formal ← nickname`. Remove the formal name from every set headed by the nickname:

```text
adela ← della
adelaide ← della
amanda ← mandy
ann ← annie
anna ← annie
anthony ← tony
arabella ← bella
barbara ← babs
barbara ← barbie
barnabas ← barney
bartholomew ← bart
beatrice ← bea
isabella ← bella
benjamin ← benjy
elizabeth ← beth
robert ← bill
william ← bill
robert ← billy
william ← billy
robert ← bob
bradford ← brad
carolann ← carol
caroline ← carol
cassandra ← sandra
catherine ← cathy
cathleen ← cathy
charles ← charlie
chester ← chet
christina ← tina
cinderella ← cindy
clarissa ← clara
clifford ← cliff
cordelia ← delia
daniel ← danny
deborah ← debbie
debra ← debbie
fidelia ← delia
delilah ← della
richard ← dick
elizabeth ← eliza
emanuel ← manuel
florence ← flora
francine ← francie
francis ← frankie
frank ← frankie
geraldine ← jerry
gertrude ← gertie
gertrude ← trudy
harold ← harry
henry ← harry
james ← jamie
jane ← jessie
janet ← jessie
josophine ← josey
judith ← judy
julia ← julie
kenneth ← kenny
lamont ← monty
laurence ← larry
lawrence ← larry
louise ← lois
louis ← lou
lucas ← luke
lucinda ← lucy
mehitabel ← mabel
madeline ← maddy
madelyn ← maddy
madeline ← madie
madelyn ← madie
manuel ← manny
margaret ← marge
margaretta ← marge
marjorie ← margie
marvin ← marv
mary ← mitzi
melissa ← missy
mervin ← merv
michael ← mike
wilhelmina ← minnie
mitchell ← mitch
newton ← newt
obediah ← obie
oliver ← ollie
patricia ← patty
patricia ← trish
prudence ← prudy
raymond ← ray
reginald ← reggie
ronald ← ronny
sandra ← sandy
susan ← sue
tabitha ← tabby
tammie ← tammy
terence ← terry
teresa ← tess
teresa ← tessa
theresa ← tess
theresa ← tessa
thaddeus ← thad
theodore ← theo
thomas ← thom
thomas ← tom
victor ← vic
victoria ← vicki
wilbur ← will
william ← willie
winnifred ← winnie
winnifred ← winny
```

**Kept two-way** — spelling variants, translations, and nicknames of the same third name. Both sets stay as they are, so these never fold:

```text
agnes ↔ inez
allan ↔ allen
amos ↔ moses
anselm ↔ selma
bess ↔ bessie
carla ↔ karla
casper ↔ jasper
chris ↔ kris
ellen ↔ helen
eloise ↔ heloise
erwin ↔ irwin
esther ↔ hester
hailey ↔ haylee
hailey ↔ hayley
haylee ↔ hayley
ian ↔ john
jane ↔ jean
jim ↔ jimmie
joe ↔ joey
john ↔ jon
johnathan ↔ johnathon
johnathan ↔ jonathan
johnathan ↔ jonathon
johnathon ↔ jonathan
johnathon ↔ jonathon
jonathan ↔ jonathon
kimberley ↔ kimberly
lil ↔ lilly
lil ↔ lily
lilly ↔ lily
marie ↔ mary
orlando ↔ roland
rodger ↔ roger
shaun ↔ shawn
stacey ↔ stacie
zach ↔ zack
```
