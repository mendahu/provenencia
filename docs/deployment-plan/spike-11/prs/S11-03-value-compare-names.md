# S11-03 — Value compare: names

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a Go comparer, moved types and seeds, docs)
**Board:** —
**Depends on:** S11-02 (landed [#441](https://github.com/mendahu/provenencia/pull/441))
**Blocks:** S11-05
**Branch:** `s11-03-value-compare-names`
**Version bump:** none — no caller uses the new comparer yet; the moved types, seeds and knobs behave exactly as before ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md), [`add-seeded-vocabulary`](../../../../.agents/skills/add-seeded-vocabulary/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [layer 1](../comparison-layers.md#layer-1-value-comparison), [relations](../comparison-layers.md#relations) and its worked cases, [detail](../comparison-layers.md#detail), [layer 1 tests](../comparison-layers.md#tests), [match registry](../comparison-layers.md#match-registry); [`docs/matching-model.md`](../../../matching-model.md) knob inventory; [`docs/matching.md`](../../../matching.md) Names; [S11-02 plan](S11-02-value-compare-core.md) D-7, D-8, D-10, D-16 and its Landed note; working rules "Build, then migrate", "High tolerance for churn", "Registry numbers: carry over, then estimate", "Value compare returns a relation, a similarity and a detail", "A normalized `Key` per value type", "One match registry", "One matching skill, grown with the layers", "Frequency weighting is gone"
**Planned:** 2026-10-17 against `main` @ `458c20f` (S11-02 landed)
**Review:** — (not reviewed yet)

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

Value compare learns names. `valuecompare.Compare` relates two name values by their parts and returns a relation (`Same`, `Equivalent`, `Contains`, `Near`, `Different`, `Unknown`), a similarity, and a detail that lists which word paired with which, in which role, and why. Initials and alias tables join value compare's word similarity. The name types (`NameRole`, `NamePattern`, `AliasTable`) move into `core/valuecompare`; the product alias and nickname seeds, `Register`, and the nickname license move to `core/database/namealiases`, where `Install` already writes them. The name knobs join the registry's value section.

**Why now.** S11-02 landed the contract, the word helpers and the registry. S11-05 moves every name caller onto this comparer; it needs the comparer first.

**What stays.** Nothing changes for researchers. `core/match`'s `NameComparer` keeps scoring Promote until S11-05; it reads the moved types and the registry's numbers, so every score it gives is today's. `namealiases.Install` writes the same rows.

## 2. User stories

Weights in this section are the Western pattern's: surname F = 1.5, first given G1 = 1, other given G2 = 0.5, nick N = 0.3, untyped U = 1, a generation word Gen = 0.5 (D-4). A name is written `given | surname` unless a part type is named; `(u)` is untyped, `(n)` nick, `(x)` suffix, `(p)` prefix. Similarity is Σ earned / Σ weight of every word on both sides (D-11 step 9), written as the arithmetic.

**US-1.** As a contributor building a matching layer, I compare two names and read what kind of match it is, how close, and which parts paired, with no database.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `James Robins` / `James Robins`; `JAMES Robins.` / `james robins`; surname listed first / given first; `Rev. James Robins` / `Dr. James Robins`; `Vincent van Gogh` (surname prefix) / `Vincent Gogh`; `蒋 浩` / `浩 蒋` | `Same`, 1, cause `exact` |
| AC-1.2 | `José Núñez` / `Jose Nunez` | `Equivalent`, (2·0.95·G1 + 0.95·2F) / (2G1 + 2F) = 0.95, cause `fold` |
| AC-1.3 | With the abbreviation table `{james, jas}`: `Jas.` / `James`. With the nickname table `{james, jim}`: `Jim` / `James` | `Equivalent` 0.98 cause `abbreviation`; `Equivalent` 0.98 cause `alias` |
| AC-1.4 | `Jim` / `James` with no alias table; `Bill` / `William` as surnames with the nickname table `{bill, william}` (its roles are given, nick, untyped) | `Different`, 0, cause `no_match` for both |
| AC-1.5 | `Robins` (surname only) / `James Robins` | `Contains`, broader A, cause `missing_part`, OnlyB `james` (kind `given`), 2F / (2F + G1) = 3 / 4 = 0.75 |
| AC-1.6 | `James` / `James Robins`; `James Robins` / `James William Robins`; `James Robins Jr.` / `James Robins` | `Contains` 2G1 / (2G1 + F) = 2 / 3.5 ≈ 0.571, broader A; `Contains` (2G1 + 2F) / (2G1 + 2F + G2) = 5 / 5.5 ≈ 0.909, broader A; `Contains` 5 / 5.5 ≈ 0.909, broader B, OnlyA `jr` (kind `generation`); each cause `missing_part` |
| AC-1.7 | `J. Robins` / `James Robins` | `Contains`, broader A, cause `initial`, (0.5·2G1 + 2F) / (2G1 + 2F) = 4 / 5 = 0.8 |
| AC-1.8 | `James Robbins` / `James Robins`; `Ann Robins` / `Anne Robins` | `Near`, cause `spelling`: (2G1 + (6/7)·2F) / 5 ≈ 0.914; (0.75·2G1 + 2F) / 5 = 0.9 |
| AC-1.9 | `James Robins` / `James (u) Robins (u)` | `Near`, cause `role`, (0.8·(G1 + U) + 0.8·(F + U)) / (G1 + F + 2U) = 0.8 |
| AC-1.10 | `James Kenneth Robins` / `James William Robins`; `James K. Robins` / `James W. Robins` | `Near`, cause `overlap`, (2G1 + 2F) / (2G1 + 2G2 + 2F) = 5 / 6 ≈ 0.833, OnlyA and OnlyB each hold the middle word |
| AC-1.11 | `Ann` / `Anne Robins` | `Near`, cause `spelling`, paired mean 0.75 meets the minimum (D-11 step 10), similarity 0.75·2G1 / (2G1 + F) = 1.5 / 3.5 ≈ 0.429 |
| AC-1.12 | `Mary Robins` / `James Robins`; `Mary Smith` / `Mary Robins`; `Maria Silva Costa` / `Maria Costa Pereira`; `W. Robins` / `James Robins`; `Ada Lovelace` / `James Robins` | `Different`, 0, cause `no_match` for each (a first given name or a surname left over on both sides; or no pair) |
| AC-1.13 | Swapped types: given `Robins`, surname `James` / `James Robins`; `J. Robbins` / `James Robins` | `Different`, 0, cause `below_minimum` (paired means 0.5 and (0.5·2 + (6/7)·3) / 5 ≈ 0.714, under 0.75); `Parts` still lists the pairs |
| AC-1.14 | `James Robins Jr.` / `James Robins Sr.` | `Different`, 0, cause `generation_conflict` |
| AC-1.15 | A name with no parts (form only); a name whose parts are only a title and a suffix; no name | `Unknown`, 0, cause `empty` |
| AC-1.16 | A name carrying its own `Pattern` that maps `given` to family and `surname` to given, with parts given `Robins`, surname `James` / `James Robins` under the Western pattern | `Same`, 1 (each name reads its roles from its own pattern) |

**US-2.** As a contributor, the name comparer keeps the layer 1 contract, and a name key agrees with `Compare`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The name sample table and 500 generated pairs per seed, under `Default()` | The AC-2.1 – AC-2.4 bounds of the S11-02 contract (`TestComparerContract`), including the swap: same relation, similarity and cause, `Broader` flipped, `Parts` mirrored |
| AC-2.2 | The same pairs through `Key` under `Default()` (no alias tables) | Equal at `Same` exactly when `Same`; equal at `Equivalent` exactly when `Same` or `Equivalent` (`TestKeyContract`) |
| AC-2.3 | `Jim` / `James` with the nickname table `{james, jim}` | `Equivalent`, and their `Equivalent` keys differ: keys never read alias tables (D-10) |
| AC-2.4 | `Key` of `Robins James` (surname first) and `James Robins` at `Same` | Both `family=robins\|given=james` |
| AC-2.5 | Any pair compared under its sides' surname parts moved to the front | The same result |

**US-3.** As a contributor, every name number is in the registry, and today's matching reads it from there.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `go test ./core/matchregistry/...` | `TestKnobInventoryMatchesRegistry` passes with no `value.Names.` or `value.Words.` prefix left in `notYetBuilt` |
| AC-3.2 | An inventory row is added for `value.Names.Aliases`, or `value.Names.Aliases` is dropped from the "Not knobs" section | `TestKnobInventoryMatchesRegistry` fails naming `value.Names.Aliases` |
| AC-3.3 | `rg -n 'value\.(Words\.(InitialCredit\|AliasCredit\|MaxExactPairing)\|Names\.(UntypedAffinity\|NickAffinity\|CrossRole\|Western))' core/match/registry.go` | Seven lines, one per knob D-13 names; the only name literals left in that file are `GivenOnlyFactor`, `SuffixConflict` and the frequency settings (removed knobs S11-05 and S11-24 delete) |

**US-4.** As a researcher, nothing behaves differently.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The whole Go and Xcode suite | `scripts/check-pr.sh` prints `check-pr: PASS`; no expected value changes in `core/match`, `core/database/namealiases`, `core/database/nameformats`, `core/database/matching` or `core/database/promotealign` tests |
| AC-4.2 | A new catalog | `namealiases.Install` writes the same tables, sets and roles as before (`core/database/namealiases` tests unchanged in what they assert) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `NameRole` and its roles, `NameWeights`, `NamePattern` (without `Key`), `WesternPattern`, `AliasKind`, `AliasAbbreviation`, `AliasNickname`, `AliasTable` move from `core/match` to `core/valuecompare` (D-2) | STEP-1, AC-4.1 |
| The alias and nickname seeds, `Register`, `AliasTables`, `ResetAliasTablesForTest` and `core/match/third_party/nicknames/` move from `core/match` to `core/database/namealiases` (D-3) | STEP-1, AC-4.2 |
| The name knobs in `Rules` and the registry: `Words.InitialCredit`, `AliasCredit`, `MaxExactPairing`; a `Names` group (D-4); the inventory and its guard (D-5, D-6) | STEP-2, AC-3.1, AC-3.2 |
| `core/match` reads those knobs from the registry and calls value compare's pairing; no second copy (D-13) | STEP-2, AC-3.3, AC-4.1 |
| Initials and aliases in value compare's word similarity; the alias index with normalized tokens (D-7, D-8) | STEP-3, AC-1.3, AC-1.4, AC-1.7 |
| The name comparer and its `Key` (D-1, D-9 – D-12) | STEP-4, AC-1.1 – AC-1.16, AC-2.3, AC-2.4 |
| Contract rows and generator for names | STEP-5, AC-2.1, AC-2.2, AC-2.5 |
| Today's name ladders and generated invariants, ported onto relations (D-14) | STEP-6 |
| The matching skill and the seeded-vocabulary skill; docs that name the moved code; the relations table's worked cases (D-15) | STEP-7, STEP-9 |

| Out | Home |
| --- | --- |
| Moving any caller onto the name comparer: `match.NameComparer`, `nameformats.Loaded.Comparer`, `namealiases.Comparer`, `graphalign`, `promotealign` | S11-05 |
| Deleting `match`'s name comparer, `wordSimilarity`, `aliasHit`, `indexAliases`, `GivenOnlyFactor`, `SuffixConflict`, `NamePatternSet` | S11-05 |
| Name frequency (`NameComparer.Frequency`, `CompareDiscount`, `Stats.Words`) | S11-05 (comparer), S11-24 (stats) |
| The reconciler's name module (`core/autoreconcile/names.go` `partFits`, `subsumes`, `outvotes`) | S11-06 |
| Normalizing alias tokens at storage (`namealiases` writes, `Key` callers) | S11-07 |
| A `generation` weight slot in stored name formats (`nameformats` seed, `name_format_profile_weights`) | not planned: the generation weight is a `Names` knob, not a pattern weight (D-4) |
| A shortlist key that joins alias sets | S11-15 |
| A name `Overlaps` relation | not planned: a name whose sides each say something the other lacks is `Near` (D-11, user call 2026-10-17) |

## 4. Current state

Verified at `458c20f`.

| Area | Where | Today |
| --- | --- | --- |
| Name types | `core/match/names.go` `NameRole`, `RoleFamily`, `RoleGiven`, `RoleNick`, `RoleUntyped`, `RoleGeneration`, `RoleIgnored`, `NameWeights`, `NamePattern` (`Key`, `PartRoles`, `Weights`; unexported `role`, `weight`), `NamePatterns`, `NamePatternSet` | Matching's types; the layer needs them and cannot import `core/match` |
| Name comparer today | `core/match/names.go` `NameComparer`, `Compare`, `CompareDiscount`, `pattern`, `nameWords`, `affinity`, `compareNames`, `countedWeight`, `familyConflict`, `sharesAny` | Bare 0…1 score: one-sided words are neutral (`James` / `James Robins` is 1), a shared given name with unrelated surnames is scaled by `GivenOnlyFactor`, a suffix conflict by `SuffixConflict` |
| Alias types and seed | `core/match/registry.go` `AliasCredit` (const 0.98), `AliasKind`, `AliasAbbreviation`, `AliasNickname`, `AliasTable`, `aliasPartRoles`, `noAliasTables`, `productAliasTables`, `MaxExactPairing` (const 12); `core/match/nicknames_seed.go` `nicknameSets`; `core/match/aliases.go` `aliasTables`, `init`, `resetAliasTables`, `AliasTables`, `Register`, `ResetAliasTablesForTest`, `aliasTablesOf`; `core/match/third_party/nicknames/LICENSE`, `NOTICE` (NOTICE names `core/match/nicknames_seed.go`) | Seed data in the matching package; only `namealiases` reads it |
| Name knobs today | `core/match/registry.go` `DefaultWordRules.InitialCredit` `Set(0.5)`; `DefaultNames` `UntypedAffinity` 0.8, `NickAffinity` 0.9, `CrossRole` 0.5, `GivenOnlyFactor` 0.5, `SuffixConflict` 0.3, frequency settings; `WesternPattern`, `WesternNamePattern`, `BuiltinNamePatterns` | Literals in `match` |
| Word helpers in `match` | `core/match/words.go` `WordRules`, `wordRules`, `aliasSet`, `resolve`, `indexAliases`, `aliasHit`, `wordSimilarity`, `bestPairing`, `bestAssignment`, `exactAssign`, `greedyAssign` | Pairing and alias code `match` still owns; `bestAssignment` reads `MaxExactPairing` |
| Value compare | `core/valuecompare/valuecompare.go` `Value` (`Type`, `Text`, `Term`, `Integer`, `Ref`), `comparers`, `Compare`, `Key`, `PartKind` (`PartWord` only), `Cause` codes; `core/valuecompare/rules.go` `Rules{Words, Text, Integers}`, `WordRules` (no initial, alias or pairing knob); `core/valuecompare/words.go` `wordSimilarity(x, y, fx, fy string, w WordRules)`, `alignWords`; `core/valuecompare/text.go` | No name comparer; word similarity has no initials or aliases |
| Registry | `core/matchregistry/value.go` `defaultValue`; `core/matchregistry/inventory_test.go` `notYetBuilt` (`value.Words.InitialCredit`, `value.Words.AliasCredit`, `value.Words.MaxExactPairing`, `value.Names.`, …), `registryKeys` | Names not built |
| Inventory | `docs/matching-model.md` rows `value.Words.InitialCredit`, `AliasCredit`, `MaxExactPairing`, `value.Names.NearMinimum` through `value.Names.Western.PartRoles`; "Not knobs" names `DefaultWordRules.Aliases`, `WesternPattern`; "Removed knobs" | Rows wait for this PR |
| Seed readers | `core/database/namealiases/namealiases.go` `Install`, `Load`, `Warm`, `Comparer`, `withAliases`; `core/database/namealiases/add.go` `AddProductNicknames`, `productNicknames` (both call `match.AliasTables()`) | Read the seed from `match` |
| Name formats | `core/database/nameformats/nameformats.go` `Profile.Pattern` (`match.NamePattern`), `Loaded.Set` (`match.NamePatternSet`), `Loaded.Comparer`, `LoadID` (sets `Pattern.Key`), `builtinLoaded`; `core/database/nameformats/registry.go` `seedRole.Role` (`match.NameRole`), `seedProfiles` (`match.WesternPattern`, `match.Role*`) | Use the match types |
| Tests | `core/match/names_test.go` `nameCases`, `TestNameComparerScores`, `TestLetterFoldsComeFromTheMap`, `withAliasSeed`, `TestSharedNicknameDoesNotMergeCanonicalNames`, `TestAliasTablesComeFromTheList`, `TestSurnamesDoNotUseGivenAliases`, `TestUnknownTokenFallsThrough`, `TestRegisterAddsAnAliasTable`, `TestSpacedParticlesComeFromTheList`, `TestNameComparerNotComparable`, `TestOneSidedNameWordsAreNeutral`, `TestNameComparerLadders`, `TestNameComparerInvariants`, `TestNameComparerPatterns`, helpers `nv`, `patterned`, `untyped`, `surnamesFirst`, `addPart`; `core/database/namealiases/namealiases_test.go`; `core/database/nameformats/nameformats_test.go`; `core/database/matching/matching_test.go` (uses `match.WesternPattern`) | `withAliasSeed` reads `productAliasTables`; `TestRegisterAddsAnAliasTable` and `TestSharedNicknameDoesNotMergeCanonicalNames` test the seed |
| Contract | `core/valuecompare/contract_test.go` `samples`, `generators`, `TestComparerContract`, `TestKeyContract`; `core/valuecompare/valuecompare_test.go` `TestCompareUnsupported` (expects `ComparerTypes()` = `integer`, `term`, `text`; a name value is `unsupported`) | Names not covered |
| Docs naming moved code | `docs/seeded-vocabulary.md` L562, L582; `docs/matching.md` L94; `.agents/skills/add-seeded-vocabulary/SKILL.md` L32; `docs/deployment-plan/spike-11/comparison-layers.md` relations table name row, worked cases, §Detail cause list | Name `core/match` paths |

**Precedent.** S11-02 ([plan](S11-02-value-compare-core.md), [#441](https://github.com/mendahu/provenencia/pull/441)): moved the word helpers into `valuecompare` with no aliases and repointed `match` in the same commit (D-8), made `match` read the registry instead of its literals (D-10), and built the text comparer on a numbered rule list (D-7). Copy all three shapes. S11-22 ([#434](https://github.com/mendahu/provenencia/pull/434)) moved types out of a package and repointed every caller in one commit.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Name values.** `Value` gains `Name *namevalue.Value` and `Pattern *NamePattern`. `comparers[propertymeta.ValueTypeName] = name{}` in `core/valuecompare/name.go` (new). A name reads its roles and weights from `v.Pattern`, or from `rules.Names.Western` when `Pattern` is nil. Only `Parts` are read; `Form` never is | One entry point (S11-02 D-3). A stored name format is catalog data, so it rides on the value the loader built; the built-in Western pattern is the registry's | A `Patterns` map in `Rules` keyed by `Value.NamePattern` — because the inventory guard would walk it as `value.Names.Patterns.western.*` and catalog patterns are not knobs |
| D-2 | **Types move to `core/valuecompare`.** `core/valuecompare/namepattern.go` (new): `type NameRole string` with `RoleFamily`, `RoleGiven`, `RoleNick`, `RoleUntyped`, `RoleGeneration`, `RoleIgnored` (same strings); `NameWeights{Family, FirstGiven, OtherGiven, Nick, Untyped float64}`; `NamePattern{PartRoles map[string]NameRole; Weights NameWeights}` (no `Key`: a profile's key is `nameformats.Profile.Key`, and `match.NamePatternSet` keys its map); exported methods `(NamePattern) Role(partType string) NameRole` and `(NamePattern) Weight(role NameRole, firstGiven bool) float64` (today's `role` / `weight`); `const WesternPattern = "western"` (the key catalogs store for the built-in pattern). `core/valuecompare/aliases.go` (new): `AliasKind`, `AliasAbbreviation`, `AliasNickname`, `AliasTable{Key, Origin, Label string; Kind AliasKind; Roles []NameRole; Sets [][]string}` unchanged. Every caller is repointed in the same commit; `match` keeps `NamePatterns` and `NamePatternSet` (its comparer plumbing) typed on `valuecompare.NamePattern`. `nameformats.LoadID` builds `valuecompare.NamePattern{PartRoles: …}` with no `Key`; `names_test.go` `patterned` and `TestNameComparerPatterns` drop their `Key` assignments | The layer cannot import `core/match` (impure); a type in two places drifts | Type aliases in `match` — because the spike rules out aliases and shims |
| D-3 | **Seeds move to `core/database/namealiases`.** `seed.go` (new) holds `productAliasTables`, `aliasPartRoles`, the `aliasMu` / `aliasTables` state, `init`, `resetAliasTables`, `AliasTables`, `Register`, `ResetAliasTablesForTest`, moved unchanged from `core/match/registry.go` and `aliases.go`. `git mv core/match/nicknames_seed.go core/database/namealiases/nicknames_seed.go` (package line changed) and `git mv core/match/third_party core/database/namealiases/third_party`; NOTICE names the new path. `Install` and `productNicknames` call the package's own `AliasTables()`. `match` keeps `noAliasTables` (`DefaultWordRules.Aliases`) and moves `aliasTablesOf` into `words.go`; `core/match/aliases.go` is deleted | The map scope entry; `Install` is the only reader of the seed, and seeds live beside their `Install` (`add-seeded-vocabulary`) | A pure seed package — because nothing pure reads the seed |
| D-4 | **Knobs.** `WordRules` gains `InitialCredit float64` (0.5, `DefaultWordRules.InitialCredit`), `AliasCredit float64` (0.98, `match.AliasCredit`), `MaxExactPairing int` (12, `match.MaxExactPairing`). New `NameRules` group, `Rules.Names`: `NearMinimum` 0.75 (`graphalign` name `PropertyScale.Floor`), `UntypedAffinity` 0.8, `NickAffinity` 0.9, `CrossRole` 0.5 (`DefaultNames`), `Generation` 0.5 (new, estimate: a Jr. on one side only says about as much as a middle name, G2 0.5), `Western NamePattern` (today's `WesternNamePattern` roles and weights 1.5 / 1 / 0.5 / 0.3 / 1), `Aliases AliasIndex` (catalog data, D-5; `Default()` leaves it empty). `core/matchregistry/value.go` sets each with a comment naming its source, and builds `PartRoles` fresh on each call (S11-02 D-2). `GivenOnlyFactor` and `SuffixConflict` get no field: a surname or first given name left over on both sides, and a generation conflict, are `Different` (D-11), so neither scales anything | Carry over today's values; a generation word needs a weight once one-sided words count (D-11 step 9); one knob for all patterns because catalog patterns have no generation slot | A `Generation` weight on `NamePattern` — because stored name formats have no such slot and would read 0 |
| D-5 | **Catalog data in `Rules` is not a knob.** `core/matchregistry/inventory_test.go` gains `var catalogData = []string{"value.Names.Aliases"}`: registry paths under one of these prefixes are skipped as fields, and the test fails when a prefix matches no registry field, when an inventory row starts with it, or when the "## Not knobs" section does not name it in backticks. `matching-model.md` "Not knobs" replaces `DefaultWordRules.Aliases` with `value.Names.Aliases` (alias tables, passed in at run time) and moves `WesternPattern` to `valuecompare.WesternPattern` | The design puts alias tables in `Rules` ("Language rules: alias and nickname tables, name patterns" are `Rules`) and lists them under Not knobs; the guard must know which field is which, by a list it checks both ways | An inventory row for a catalog field — because the inventory is the tunables; a struct tag — because the guard reads paths, not tags (S11-02 D-16) |
| D-6 | **Inventory rows.** Rename `value.Names.Western.Family`, `.FirstGiven`, `.OtherGiven`, `.Nick`, `.Untyped` to `value.Names.Western.Weights.Family`, `.FirstGiven`, `.OtherGiven`, `.Nick`, `.Untyped` (the path of the moved struct). Add `value.Names.Generation` (0.5, estimate, "Weight of a generation word, Jr. or III, on either side"). Delete the `value.Names.GivenOnlyFactor` and `value.Names.SuffixConflict` rows; add to "Removed knobs": "`DefaultNames.GivenOnlyFactor`, `SuffixConflict` — A surname or first given name left over on both sides, and a generation conflict, are `Different` (S11-03)". Delete the prefixes `value.Words.InitialCredit`, `value.Words.AliasCredit`, `value.Words.MaxExactPairing`, `value.Names.` from `notYetBuilt` | The guard's rule: every field a row, every row a field | — |
| D-7 | **Word similarity gains initials and aliases** (`core/valuecompare/words.go`): `wordSimilarity(x, y, fx, fy string, w WordRules, o wordOptions) (float64, Relation, Cause)` with `type wordOptions struct { initials bool; aliases AliasIndex; roleX, roleY NameRole }`. In order: byte-equal → 1 `Same` `exact`; either token `IsInitial`: when `o.initials` and the first runes of `x` and `y` are equal → `InitialCredit`, `Contains`, `initial`, else no pair; `o.aliases.hit(x, y, o.roleX, o.roleY)` → `AliasCredit`, `Equivalent`, cause `abbreviation` for an `AliasAbbreviation` table, `alias` for any other; then today's fold, length cap and spelling steps. `alignWords` (text) passes `wordOptions{}`, so text keeps every result | One word table for both comparers ("adds initials and aliases to word similarity"); the order is today's `match.wordSimilarity` | A second word table for names — because a fold or spelling fix would land in one copy |
| D-8 | **Alias index.** `type AliasIndex struct { sets []aliasSet }` with unexported `aliasSet{tokens map[string]bool; roles map[NameRole]bool; kind AliasKind}`; `func NewAliasIndex(tables []AliasTable) AliasIndex` normalizes every token with `NormalizeForm`, drops empty tokens and sets left with none; `(AliasIndex) hit(x, y string, rx, ry NameRole) (AliasKind, bool)` is true when one set holds both tokens and both roles. A zero `AliasIndex` hits nothing | Index once per load, not per compare (today `resolve` indexes on every call); normalized tokens close "`match.indexAliases` doesn't normalize alias tokens" ([found in passing](../comparison-layers.md#reuse-across-the-app)) | Indexing inside `Compare` — because it repeats per pair |
| D-9 | **Name words.** For each part in order, its role under the pattern; `RoleIgnored` parts are dropped; each `RoleGeneration` part's tokens go to the side's generation list; every other part's tokens (`SplitWords` with `Particles`) become words with that role, and the first `RoleGiven` word in part order is the first given word, weighted `FirstGiven`, later given words `OtherGiven`. Words are then sorted by role (family, given, nick, untyped) and token, stably, so part order never changes the pairing (only which given word is first). A side with no word is empty; generation tokens alone do not make a side non-empty | Today's `nameWords` rule, plus a fixed order so the pairing and its tie-breaks do not depend on how the parts were entered (AC-2.5) | — |
| D-10 | **Name `Key`.** For each role in the order family, given, nick, untyped, generation that has tokens: `role=` and its tokens sorted and joined by one space; roles joined by `\|` (`family=robins\|given=james william\|generation=jr`). At `Equivalent` each token is `FoldWord`ed first. An empty name has no key. Keys never read alias tables, and `Key`'s doc comment says so: "Name keys ignore alias tables: an alias joins two words that share a set, and sets do not chain (Chris ~ Christine, Chris ~ Christopher, Christine ≁ Christopher), so no key can hold them." `TestKeyContract` runs names under `Default()`, which has no tables | A key per role and sorted words is exactly the multiset D-11 step 2 compares; folding is the only `Equivalent` without tables | Alias-aware keys — because aliases are not transitive |
| D-11 | **The name rule**, in order. (1) Either side empty → `Unknown`, `empty`. (2) Equal `Same` keys → `Same`, 1, `exact`; `Parts` pairs the sorted words position by position. (3) Both sides have generation tokens and share none → `Different`, `generation_conflict`. (4) **Pair:** `earned[i][j]` = word similarity (D-7, `initials` true, `rules.Names.Aliases`, the two roles) × affinity × (wᵢ + wⱼ), where affinity is 1 for the same role, `UntypedAffinity` when one role is untyped, `NickAffinity` for nick against given, `CrossRole` otherwise; rows are the side of the smaller `Same` key (canonical orientation: when `Key(a, Same) > Key(b, Same)` compare b against a and mirror the result: `Parts` A↔B, `OnlyA`↔`OnlyB`, `Broader` flipped); pair with `BestAssignment(earned, w.MaxExactPairing)` (D-13); a cell pairs only above 0. Generation tokens on both sides pair with their equal token, `Same`, `exact`, earned 2·`Generation`. (5) No pair → `Different`, `no_match`. (6) **Leftovers** are the unpaired words and the unshared generation tokens of each side. When both sides have leftovers and either both have a family leftover or both sides' first given words are leftovers → `Different`, `no_match`. (7) **Contributors**, each with a relation and cause: every pair, at its word relation, except a pair whose roles differ, which is `Near` with cause `role` (or `spelling` if the word relation was `Near`); an initial pair is `Contains` with `Broader` the side holding the initial; leftovers on one side only are `Contains`, `missing_part`, `Broader` the other side; leftovers on both sides are `Near`, `overlap`; `Contains` contributors that name both sides as broader are together one `Near`, `overlap`. (8) The value's relation is the weakest contributor in the order `Same`, `Equivalent`, `Contains`, `Near`; its cause is the first contributor at that relation, pairs in A's sorted order before leftovers. (9) **Similarity** = Σ earned ÷ Σ weight of every word and generation token on both sides (generation tokens weigh `Generation`); one-sided words count 0. (10) **Near minimum:** when the relation is `Near` and the paired mean (Σ earned ÷ Σ weight of paired words and shared generation tokens only) is below `NearMinimum` → `Different`, 0, cause `below_minimum`, `Parts` kept. **Detail:** `Parts` in A's sorted order, each `PartMatch` with `Part{Kind: PartKind(role), Text: token}` on each side, the word relation (D-7) adjusted by step 7, similarity word similarity × affinity, and cause; `OnlyA` / `OnlyB` the leftovers with their kinds | Satisfies every name cell of the [relations table](../comparison-layers.md#relations) and the worked cases, with the user's calls of 2026-10-17: swapped part types stay under the Near minimum (`Different`), and only a surname or first given name left over on both sides makes a disagreement. The paired mean gates `Near` so a one-sided surname does not turn `Ann` / `Anne Robins` into `Different`. One-sided words at 0 is the design's "weighted mean of the part similarities", and makes `Contains` score below 1 | Today's neutral one-sided words — because `Contains` must score below 1; gating `Near` on the full similarity — because adding a surname to one side would turn a spelling variant into `Different`; an `Overlaps` relation for names — because the user chose `Near` for leftovers on both sides |
| D-12 | **New codes.** `PartKind` gains `PartFamily = "family"`, `PartGiven = "given"`, `PartNick = "nick"`, `PartUntyped = "untyped"`, `PartGeneration = "generation"` (the role strings, so `PartKind(role)` converts). `Cause` gains `CauseRole = "role"` (words paired across part types) and `CauseBelowMinimum = "below_minimum"` (a `Near` pair under its type's minimum). S11-04 declares `CauseBelowMinimum` too: whichever of S11-03 and S11-04 lands second keeps the one declaration on rebase | Codes are strings for the wire (S11-02 D-5); the design's part kinds name family and given | — |
| D-13 | **`core/match` keeps no second copy** (S11-02 D-10). `core/match/registry.go`: `DefaultWordRules.InitialCredit` reads `value.Words.InitialCredit`; `AliasCredit` and `MaxExactPairing` become `var`s reading `value.Words.AliasCredit` / `MaxExactPairing`; `DefaultNames` `UntypedAffinity`, `NickAffinity`, `CrossRole` read `value.Names.*`; `var WesternNamePattern = value.Names.Western`; `BuiltinNamePatterns` keys it by `valuecompare.WesternPattern`. `GivenOnlyFactor`, `SuffixConflict` and the frequency settings keep their literals (removed knobs; S11-05 and S11-24 delete them). `bestAssignment`, `exactAssign`, `greedyAssign` move from `core/match/words.go` to `core/valuecompare/pairing.go` (new) as `func BestAssignment(score [][]float64, maxExact int) (total float64, rowTo []int)` (`rowTo[i]` is the paired column or -1); `match.bestPairing` and `compareNames` call it and derive their used flags. Every `match` expected value stays | One number, one place; one pairing search | Copying the pairing into `valuecompare` — because the skill's "do not" forbids a second copy |
| D-14 | **Tests.** `core/match/names_test.go`: `withAliasSeed` becomes `withAliasFixture`, a table built in the test from the abbreviation set `{james, jas}` and every set of `nicknames_seed.go` holding `james`, `jim`, `jacob`, `ben`, `benjamin`, `doug`, `douglas`, `bill` or `william`, copied verbatim; every case keeps its expected value. `TestSharedNicknameDoesNotMergeCanonicalNames` and `TestRegisterAddsAnAliasTable` move to `core/database/namealiases/seed_test.go` (new), asserting through `valuecompare.Compare` with `NewAliasIndex(AliasTables())`: Chris / Christine `Equivalent` 0.98 `alias`, Tina / Christopher `Different`; `Register` appends and `AliasTables` returns it, and `Default()` rules do not hit it. The new comparer's tests are in `core/valuecompare/name_test.go` (new), `package valuecompare_test`, with a `nv` helper like `match`'s | `match`'s tests keep guarding live code until S11-05; the seed's tests follow the seed | Deleting `match`'s alias tests — because Promote still runs that code |
| D-15 | **Docs and skills.** `add-match-rule` gains: a name value rides its own pattern (D-1); catalog data in `Rules` goes in `catalogData` and the Not knobs section (D-5); `BestAssignment` and `wordSimilarity` are the shared pairing and word table. `add-seeded-vocabulary` L32 and `seeded-vocabulary.md` L562, L582 name `core/database/namealiases/seed.go` (new), and `nicknames_seed.go` and `third_party/nicknames/` under `core/database/namealiases/`; `matching.md` L94 the same. `comparison-layers.md`: the cause list gains `role` and `below_minimum`; the name row of the relations table reads Near "spelling variant; different middle or untyped words", Different "a surname or first given name on each side that the other lacks (`Mary Robins` / `James Robins`), generation conflict, under the Near minimum"; the worked case "surname James, given Robins / James Robins" reads `Different` (under the Near minimum; decided 2026-10-17, S11-03) | The working rule "Each migration PR also updates the domain docs and idea files it makes false"; the worked case changed by a user call | — |
| D-16 | Commits, each passing the full suite: (1) types and seeds move; (2) knobs, inventory and `match` reads them; (3) word similarity and the alias index; (4) the name comparer; (5) contract rows; (6) ladders and invariants; (7) skills and docs | One concern per commit | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/valuecompare/name.go` *(new)*: `name` comparer | file, type | `core/match/names.go` `NameComparer` | Returns a bare score and reads frequency; build, then migrate: S11-05 deletes it |
| `core/valuecompare/namepattern.go`, `aliases.go` *(new)* | files | `core/match/names.go`, `registry.go` | The types **move** here (D-2) |
| `valuecompare.AliasIndex`, `NewAliasIndex` *(new)* | type, func | `core/match/words.go` `indexAliases`, `aliasSet`, `aliasHit` | `match`'s index is rebuilt per call and leaves tokens raw; `match`'s copy is deleted with its comparer in S11-05 |
| `valuecompare.NameRules` *(new)* | type | `match.NameComparer` pointer settings | Whole values passed in (S11-02 D-1) |
| `wordOptions` *(new, unexported)* | type | `match.wordRules.initials` | Extends the one word table (D-7) |
| `valuecompare.BestAssignment`, `core/valuecompare/pairing.go` *(new)* | func, file | `core/match/words.go` `bestAssignment` | **Moves** here (D-13) |
| `PartFamily`, `PartGiven`, `PartNick`, `PartUntyped`, `PartGeneration`, `CauseRole`, `CauseBelowMinimum` *(new)* | codes | `valuecompare.PartWord`, cause list | Extends the code sets (D-12) |
| `core/database/namealiases/seed.go`, `seed_test.go` *(new)* | files | `core/match/aliases.go`, `registry.go` seed | The seed **moves** beside `Install` (D-3) |
| `catalogData` *(new)* | test list | `notYetBuilt` | A second list with the opposite meaning: fields that never get a row (D-5) |
| `core/valuecompare/name_test.go` *(new)* | test file | `core/match/names_test.go` | Tests the new comparer; `match`'s stay with live code (D-14) |

Searches run for this table: `rg -n 'AliasIndex|NewAliasIndex' core` (none), `rg -n 'func BestAssignment|bestAssignment' core` (only `core/match/words.go`), `rg -n 'NameRole|NamePattern' core --type go -l` (`core/match`, `nameformats`, `namealiases`), `rg -n 'catalogData|notKnob' core` (none).

**Builds on.** S11-02's contract, `comparers` map, word helpers, registry and guards; the `add-match-rule` skill; `add-seeded-vocabulary`'s rule that a seed lives beside its `Install`.
**Deletes.** `core/match/aliases.go`; `core/match/nicknames_seed.go` and `core/match/third_party/` (moved); the name and alias types, `productAliasTables`, `aliasPartRoles`, the literal values of `InitialCredit`, `AliasCredit`, `MaxExactPairing`, `UntypedAffinity`, `NickAffinity`, `CrossRole` and `WesternNamePattern` from `core/match/registry.go`; `bestAssignment`, `exactAssign`, `greedyAssign` from `core/match/words.go`; the `value.Names.GivenOnlyFactor` and `SuffixConflict` inventory rows; four `notYetBuilt` prefixes. `match.NameComparer`, `wordSimilarity`, `aliasHit`, `indexAliases` become dead once S11-05 moves their callers, and S11-05 deletes them.
**Next adds.** S11-05 builds name values with `Pattern` from `nameformats.Load` and `Rules.Names.Aliases` from `NewAliasIndex` over `namealiases.Load`, then deletes `match`'s comparer. S11-06 reads name relations in the reconciler. S11-07 normalizes stored alias tokens. S11-10 weighs a name's relation per Subject type.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/valuecompare` | `namepattern.go`, `aliases.go`, `pairing.go`, `name.go` *(new)* (D-1, D-2, D-8, D-11, D-13); `valuecompare.go`: `Value.Name`, `Value.Pattern`, the `name` entry, codes (D-12), `Key` doc comment (D-10); `rules.go`: three `WordRules` fields, `NameRules`, `Rules.Names` (D-4); `words.go`: `wordOptions`, `wordSimilarity`'s new step order (D-7); `text.go` / `alignWords` pass `wordOptions{}` |
| `core/matchregistry` | `value.go`: the D-4 values; `inventory_test.go`: `catalogData` and its checks, four prefixes gone (D-5, D-6) |
| `core/match` | Types and seed gone (D-2, D-3); `registry.go` reads the registry (D-13); `words.go` calls `valuecompare.BestAssignment` and holds `aliasTablesOf`; `names.go` uses `valuecompare.NameRole`, `NamePattern.Role` / `Weight` |
| `core/database/namealiases` | `seed.go` *(new)*, `nicknames_seed.go` and `third_party/nicknames/` (moved) (D-3); `Install`, `Load`, `Comparer`, `add.go` typed on `valuecompare.AliasTable` |
| `core/database/nameformats` | `match.NamePattern` → `valuecompare.NamePattern` (no `Key`), `match.NameRole` / `Role*` → `valuecompare.*`, `match.WesternPattern` → `valuecompare.WesternPattern` |

```go
type NameRules struct {
	NearMinimum     float64     // least paired mean for a Near name
	UntypedAffinity float64     // a typed word against an untyped one
	NickAffinity    float64     // a nickname against a given name
	CrossRole       float64     // any other role mismatch
	Generation      float64     // weight of a generation word (Jr., III)
	Western         NamePattern // roles and weights for a name with no pattern
	Aliases         AliasIndex  // catalog alias tables; not a knob
}

func NewAliasIndex(tables []AliasTable) AliasIndex
func BestAssignment(score [][]float64, maxExact int) (total float64, rowTo []int)
```

Errors: none. Every input has a result.

### FFI

None — no wire change.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Move the name types and the alias seed**
Files: `core/valuecompare/namepattern.go` (new), `core/valuecompare/aliases.go` (new), `core/database/namealiases/seed.go` (new), `core/database/namealiases/seed_test.go` (new), `core/database/namealiases/nicknames_seed.go` (moved), `core/database/namealiases/third_party/nicknames/` (moved), `core/match/names.go`, `core/match/registry.go`, `core/match/words.go`, `core/match/aliases.go` (deleted), `core/match/names_test.go`, `core/database/namealiases/namealiases.go`, `core/database/namealiases/add.go`, `core/database/namealiases/namealiases_test.go`, `core/database/nameformats/nameformats.go`, `core/database/nameformats/registry.go`, `core/database/nameformats/nameformats_test.go`, `core/database/matching/matching_test.go`. Skill: `add-seeded-vocabulary`.
Do: D-2, D-3, D-14 (the fixture and the two moved tests). Move with `git mv`; no type aliases.
Check: `rg -n 'match\.(NameRole|Role[A-Z]\w*|NamePattern\b|NameWeights|WesternPattern|AliasTable|AliasKind|Alias(Abbreviation|Nickname)|AliasTables|Register|ResetAliasTablesForTest)' core api` prints nothing; `ls core/match/third_party core/match/nicknames_seed.go` fails; `CGO_ENABLED=1 go test -tags fts5 ./core/match/... ./core/database/namealiases/... ./core/database/nameformats/... ./core/database/matching/... ./core/database/promotealign/...` passes with no expected value changed.

**STEP-2 — Name knobs in the registry; `match` reads them**
Files: `core/valuecompare/rules.go`, `core/valuecompare/pairing.go` (new), `core/matchregistry/value.go`, `core/matchregistry/inventory_test.go`, `docs/matching-model.md`, `core/match/registry.go`, `core/match/words.go`, `core/match/names.go`. Skill: `add-match-rule` "Add a knob".
Do: D-4, D-5, D-6, D-13. `NameRules.Aliases` is declared as `AliasIndex` with an empty struct body in this step; STEP-3 fills it.
Check: `go test ./core/matchregistry/...` passes (no tags); temporarily add a row `` | `value.Names.Aliases` | — | — | — | `` to the inventory and see `TestKnobInventoryMatchesRegistry` fail naming it, then revert; `CGO_ENABLED=1 go test -tags fts5 ./core/match/... ./core/database/promotealign/...` passes with no expected value changed; AC-3.3's `rg` prints seven lines.

**STEP-3 — Initials and aliases in word similarity**
Files: `core/valuecompare/words.go`, `core/valuecompare/aliases.go` (new), `core/valuecompare/text.go`, `core/valuecompare/words_test.go`.
Do: D-7, D-8. Add `TestWordSimilarityNames` (T-1) and `TestAliasIndex` (T-2), using `export_test.go` to expose `wordSimilarity` as `WordSimilarity` and `wordOptions` through a constructor `WordOptions(initials bool, aliases AliasIndex, rx, ry NameRole)`.
Check: `go test ./core/valuecompare/...` passes with no tags and no text test changed.

**STEP-4 — The name comparer**
Files: `core/valuecompare/name.go` (new), `core/valuecompare/valuecompare.go`, `core/valuecompare/name_test.go` (new), `core/valuecompare/valuecompare_test.go`.
Do: D-1, D-9 – D-12. In `TestCompareUnsupported`, `ComparerTypes()` becomes `integer`, `name`, `term`, `text` and the name value leaves the unsupported list. T-3, T-4.
Check: `go test ./core/valuecompare/...` passes; every AC-1.x pair is a case in `TestNameCompare`.

**STEP-5 — Contract rows for names**
Files: `core/valuecompare/contract_test.go`.
Do: a `propertymeta.ValueTypeName` row in `samples` holding every AC-1.x pair, and a generator: names of 0 to 4 parts from the pools of `match`'s `TestNameComparerInvariants` (given, nick, surname, surname prefix, suffix, prefix, undetermined, untyped), half of them a variant of the other side (case, a dropped part, an added part, a spelling from the pool, parts reordered). T-5.
Check: `go test -run 'TestComparerContract|TestKeyContract' -count=1 ./core/valuecompare/` passes.

**STEP-6 — Ladders and invariants on relations**
Files: `core/valuecompare/name_test.go` (new).
Do: T-6, T-7.
Check: `go test -run 'TestNameLadders|TestNameInvariants' -count=1 ./core/valuecompare/` passes.

**STEP-7 — Skills and docs**
Files: `.agents/skills/add-match-rule/SKILL.md`, `.agents/skills/add-seeded-vocabulary/SKILL.md`, `docs/seeded-vocabulary.md`, `docs/matching.md`, `docs/deployment-plan/spike-11/comparison-layers.md`.
Do: D-15.
Check: `rg -n 'core/match/(nicknames_seed|third_party|aliases\.go)|productAliasTables. in .core/match' docs .agents --glob '!docs/deployment-plan/archive/**' --glob '!docs/deployment-plan/spike-10/**' --glob '!docs/deployment-plan/spike-11/prs/**'` prints nothing; every path the skill names exists (`ls`).

**STEP-8 — Full suite**
Files: none.
Do: run the gate on the user's Mac, as `ship-pr` requires (`xcode-test` cannot pass elsewhere).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-11/deployment-plan.md`, `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-03-value-compare-names.md` reports no errors.

Commits: one per behavior (D-16). The subject is one full imperative sentence that says what changed, ending in a period, like the rest of `git log` ("Compare names by relation in value compare."). No WIP commits.

## 8. Tests

All new tests are table-driven (`.cursor/rules/go-tests.mdc`), in `package valuecompare_test` with no build tags unless named otherwise. Expected values are written as the arithmetic, with the weight constants of §2.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/words_test.go` `TestWordSimilarityNames` *(new)* | `j` / `james` with initials → 0.5 `Contains` `initial`; without → no pair; `w` / `james` → no pair; `jas` / `james` with an abbreviation index → 0.98 `Equivalent` `abbreviation`; `jim` / `james` with a nickname index, roles given / given → 0.98 `alias`; roles family / family → no alias hit, falls to spelling → no pair | AC-1.3, AC-1.4, AC-1.7, D-7 |
| T-2 | Go | `core/valuecompare/words_test.go` `TestAliasIndex` *(new)* | A table with tokens `Jas.`, ` JIM ` → hits `jas` / `jim` (normalized); an empty token is dropped; the zero index hits nothing | D-8 |
| T-3 | Go | `core/valuecompare/name_test.go` `TestNameCompare` *(new)* | Every AC-1.1 – AC-1.16 pair → relation, similarity (±1e-9), cause, `Broader`, `OnlyA`, `OnlyB`, and for AC-1.5, AC-1.7, AC-1.9 and AC-1.13 the `Parts` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, AC-1.12, AC-1.13, AC-1.14, AC-1.15, AC-1.16, D-11 |
| T-4 | Go | `core/valuecompare/name_test.go` `TestNameKey` *(new)* | AC-2.4; `José` / `Jose` → `Same` keys differ, `Equivalent` keys equal; AC-2.3; an empty name → `"", false`; level `Near` → `"", false` | AC-2.3, AC-2.4, D-10 |
| T-5 | Go | `core/valuecompare/contract_test.go` `TestComparerContract`, `TestKeyContract` (name rows) | The name samples and generated pairs → the S11-02 contract | AC-2.1, AC-2.2 |
| T-6 | Go | `core/valuecompare/name_test.go` `TestNameLadders` *(new)* | Each ladder's rungs score strictly in order, then each listed rung is `Different`: probe `James K. Robins`: `James K. Robins` (1) > `James K. Robbins` (5.571 / 6 ≈ 0.929) > `J. K. Robins` (5 / 6), `Different`: `Mary Robins`, `James K. Smith`, `Ada Lovelace`. Probe `James Robins Jr.`: `James Robins Jr.` (1) > `Jim` given, `James` nick, `Robbins` surname ((0.9·1.3 + (6/7)·3) / 5.8 ≈ 0.645), `Different`: `Mary Robins`, `James Robins Sr.`, `Mary Robins Sr.`. Probe `Maria Luisa Silva Costa`: `Maria`, `Luisa`, `Silva`, `Costa` as separate parts (1) > `M. Costa` (4 / 7), `Different`: `Luisa Pereira`. Probe `James Robins`: `James Robins` (1) > untyped `James Robins` (0.8) > untyped `Mary Robins` (2 / 4.5), `Different`: surname `James` given `Robins`, `Ada Lovelace` | D-11, D-14 |
| T-7 | Go | `core/valuecompare/name_test.go` `TestNameInvariants` *(new)* | `match`'s generator (seeds 1 – 4, 2000 pairs) under `Default()`, a heavy copy (Western `Family` 3, `CrossRole` 0.2) and fuzzy off (`FuzzyFloor` 1) → the form never changes a result; moving surname parts first never changes it; adding `Jr.` to one and `Sr.` to the other never raises the similarity; adding a title to one and a surname prefix to the other never changes it; the untyped copy of a name is `Same` or `Near`, never above 1 | AC-2.5, D-9, D-11 |
| T-8 | Go (fts5) | `core/database/namealiases/seed_test.go` *(new)* `TestSharedNicknameDoesNotMergeCanonicalNames`, `TestRegisterAddsAnAliasTable` (moved) | Chris / Christine → `Equivalent` 0.98 `alias`; Tina / Christopher → `Different`; `Register` appends a table `AliasTables` returns; `Default()` rules do not hit it | AC-4.2, D-3, D-14 |
| T-9 | Go | `core/matchregistry/inventory_test.go` `TestKnobInventoryMatchesRegistry` | `catalogData` entries skipped and checked both ways; no `value.Names.` prefix | AC-3.1, AC-3.2, D-5, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/match` names tests (STEP-1 and STEP-2 repoint types and knobs), `core/database/namealiases` (seed moved), `core/database/nameformats` (`Key` dropped), `TestLayerPackages` (`valuecompare` imports `core/namevalue`, pure).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: matching-model.md inventory and Not knobs / Removed knobs, seeded-vocabulary.md, matching.md L94, comparison-layers.md relations table, worked cases and cause list (STEP-2, STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

No app change. The reviewer runs, from the repo root:

1. `go test ./core/valuecompare/... ./core/matchregistry/...` with no tags: passes.
2. `go list -deps ./core/valuecompare | grep 'provenencia/core/database'`: no output.
3. `rg -n 'third_party|nicknames_seed' core/match`: no output (AC-4.2); `ls core/database/namealiases/third_party/nicknames`: `LICENSE`, `NOTICE`.
4. AC-3.3's `rg` on `core/match/registry.go`: seven lines.
5. Create a new project and open Promote on a Source with a person: suggestions and bands read as before (AC-4.1, AC-4.2).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving `bestAssignment` changes a tie-break, so a `match` score moves | `BestAssignment` is moved, not rewritten; STEP-2 runs the `match` and `promotealign` suites unchanged |
| The test alias fixture misses a set, so a `match` alias case changes | D-14 names the tokens; STEP-1 requires no expected value to change |
| The canonical orientation or a tie leaves a swap asymmetry | T-5's swap invariant over generated pairs; D-9's sorted words |
| S11-04 lands first and edits the same files (`valuecompare.go`, `rules.go`, `value.go`, `inventory_test.go`, `contract_test.go`, `matching-model.md`, the skill) | Each edit is additive; on rebase keep both sides, and one `CauseBelowMinimum` (D-12) |

**Stop and report if:**

- A row in §4 is no longer true at your base (for example `match.NamePattern` has no `Key`, or `namealiases` no longer reads `match.AliasTables()`).
- STEP-1 or STEP-2 changes any expected value in an existing test.
- An AC-1.x pair cannot hold under D-11's rule as written, or T-5's swap check fails on a pair D-11 cannot settle without a tie-break the plan does not give.
- `TestLayerPackages/core/valuecompare/numbers` fails and the fix would be anything other than a registry field.
- `valuecompare` needs to import anything under `core/database` or `core/match`.

## 12. Open questions

None.

## 13. PR

**Title:** S11-03 — Value compare: names
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (test names); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-03-value-compare-names.md`.
