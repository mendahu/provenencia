# S11-05 — Migrate matching onto value compare

**Status:** Draft
**Spike / slice:** Spike 11 · Slice 2 — Value compare, everywhere
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Go scoring moves onto value compare; no Swift or wire change)
**Board:** —
**Depends on:** S11-03 (landed [#447](https://github.com/mendahu/provenencia/pull/447)), S11-04 (landed [#445](https://github.com/mendahu/provenencia/pull/445))
**Blocks:** S11-06, S11-07, S11-09, S11-24
**Branch:** `s11-05-migrate-matching`
**Version bump:** none — Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships, as Spike 10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-match-rule`](../../../../.agents/skills/add-match-rule/SKILL.md), [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/match-rules.mdc`
**Context:** [layer 1](../comparison-layers.md#layer-1-value-comparison), [relations](../comparison-layers.md#relations), [frequency](../comparison-layers.md#frequency), [outcome from relation](../comparison-layers.md#outcome-from-relation), [what moves where](../comparison-layers.md#what-moves-where-in-todays-code); [`docs/matching-model.md`](../../../matching-model.md) "Removed knobs"; [`docs/matching.md`](../../../matching.md); [S11-02 plan](S11-02-value-compare-core.md) D-10, [S11-03 plan](S11-03-value-compare-names.md) and [S11-04 plan](S11-04-value-compare-dates.md) with their Landed notes; [S11-01 plan](S11-01-matching-model.md) D-5 (who updates the old docs); working rules "Build, then migrate", "High tolerance for churn", "Frequency weighting is gone", "One match registry", "Sparse data first"
**Planned:** 2026-10-17 against `main` @ `b6d1d48` (S11-03 and S11-04 landed)
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

Promote and `Rank` compare values with value compare. Every place `core/match`, `core/graphalign` and `core/database/promotealign` judged two values with `match`'s own comparers now calls `valuecompare.Compare` under the registry's value rules, with the catalog's name pattern and alias tables passed in as `Rules`. `match`'s five comparers, the `Comparer` interface, `ComparerFor`, `Set`, the word code and the name frequency discount are deleted, along with Promote's per-type floors, which the value layer's `Near` minimum replaces. Researchers see the design's rules: a name with words on one side only (`Robins` / `James Robins`) is a resemblance below a full match, `ABT 1817` reads as 1816 – 1818, toponyms compare in word order, and a common surname costs nothing extra.

**Why now.** S11-03 and S11-04 built the name and date comparers; until callers move, two comparers exist for every type and the registry's numbers are read twice. S11-06, S11-07, S11-09 and S11-24 all start from a `core/match` with no comparer.

**What stays.** `core/match` keeps its layer-2 and layer-3 work until S11-13: `pairValues` (leaders, support weights, negatives, cardinality), `poolPeriods`, `Score`, `Rank`, `Evaluate`, the profiles and their weights. `graphalign`'s scales, weights, frequency weights for text, terms, dates and integers (`Stats.ValueFreq`, `MPrior`, `UPrior`), edge support and bands stay (S11-16, S11-24). The reconciler's sameness test, `autoreconcile.Compatible`, still decides pins and negatives (S11-06). No Swift, wire or catalog change.

## 2. User stories

Numbers use the registry's starting values. Name word weights: family `F` 1.5, first given `G1` 1, other given `G` 0.5. A name's similarity is what its pairs earn over the weight of every word on both sides (`meanSimilarity`). Promote's name scale is weight 2.5, contradiction 3; the bands are weak 2, medium 3, strong 4.

**US-1.** As a researcher, a Subject whose name is shorter than a filed person's still reads as that person, a little less sure, and two different first names never do.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Promote with no anchor: Subject `James Robins`, one handle `James Kenneth Robins` (`TestPartialNamesAreWeak`) | Handle, weak. Name `partial`, similarity (2·G1 + 2·F) / (2·G1 + 2·F + G) = 5 / 5.5, score 2.5 × 5 / 5.5 ≈ 2.273 |
| AC-1.2 | Pairwise: `Mrs. Robins` (prefix, surname) / `Marion Robins` (`TestMissingGivenOrPrefixIsNotANameConflict`) | Name `partial`, similarity 2·F / (2·F + G1) = 0.75, score 2.5 × 0.75 = 1.875; not a conflict |
| AC-1.3 | Pairwise: `Mary Robins` / `James Robins` | Name `conflict`, −3 (`Different`: a first given name on each side) |
| AC-1.4 | Promote: Subject `Mary Robins`, one handle `James Robins` (`TestUnreachableFallbackScoresOnTheWalkScale`) | Skip, no band: `Rank` scores the name 0 and drops the handle under `MinScore` 3 |
| AC-1.5 | `Rank` with the person profile: probe `James Robins`; candidates `James Robins`, `Robins` (surname only), `Mary Robins` | `James Robins` 10; `Robins` 10 × 2·F / (2·F + G1) = 7.5; `Mary Robins` not returned (0 < 3) |

**US-2.** As a researcher, an approximate date reads as the range it stands for.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Event profile `Score`: `birth`, `ABT 1817` / `birth`, `1818` | Date contribution 6 × 0.7 = 4.2 (`Contains`, span score of 3 years); total 4 + 4.2 = 8.2 |
| AC-2.2 | Pairwise: `start_date` `BEF 31 MAR 2001` / `BEF 1 FEB 1985` (`exhibits_pin_test.go`) | Exhibit `partial`, unpinned, weight > 0 (`Contains`, open span 0.2) |

**US-3.** As a researcher, Promote compares toponyms as written and weighs no name by how common it is.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Pairwise toponym: `York, Upper Canada` / `Upper Canada, York` | `conflict` (`Different`, cause `order`) |
| AC-3.2 | Pairwise toponym: `York` / `York, Upper Canada`, multiple cardinality | `partial`, similarity 0.7 × 2 / 4 = 0.35 (`Contains`) |
| AC-3.3 | A catalog where 40 of 100 filed people are named Smith: Promote's name points for `John Smith` / `John Smith` | 2.5, the same as for a rare surname; `rg -n 'Discount|WordCounts|CountWords|Stats.Words|countNameWords' core api` prints nothing |

**US-4.** As a contributor, `core/match` holds no comparer and reads no number; the registry's value section is passed in.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `rg -n 'type (Name|Date|Text|Term|Integer)Comparer|DiscountComparer|ComparerFor|func Set\[|DefaultNames|DefaultDates|DefaultText|DefaultWordRules|matchregistry' core/match` | Nothing |
| AC-4.2 | `match.Compare(rules, "term", male, neutral unknown)`; `match.Compare(rules, "date", phrase-only, 1817)`; `match.Compare(rules, "text", a name value, a text value)`; `match.Compare(rules, "integer", {Integer: 0}, {HasInteger: true})` | Not comparable each |
| AC-4.3 | `namealiases.Rules` on an installed catalog (`namealiases` `TestWarmHoldsAliasTablesUntilInvalidated`, `TestUserAndPluginRowsScore`; `catalogsession` `TestDoKeepsAliasTablesForTheOpenProject`); `nameformats.Loaded.Rules` on an installed catalog (`TestInstallLoadsWesternAndFormats`) | `Jas.` / `James` is `Equivalent` with similarity `Words.AliasCredit` (0.98) while the alias row is there, and not after it is deleted and the cache dropped; the installed load's `Names.Patterns` holds `western` |
| AC-4.4 | `scripts/check-pr.sh` | `check-pr: PASS`; every expected value that moved is a row of §8 "Expected changes" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `match.Compare`: one adapter from `match.Value` to `valuecompare.Compare` (D-2) | STEP-3, AC-4.2 |
| `Feature.ValueType` replaces `Feature.Comparer`; `Rank`, `Score`, `Evaluate`, `pairValues` take `valuecompare.Rules` (D-1, D-3) | STEP-3, AC-1.5, AC-2.1 |
| Delete `match`'s comparers, word code, `Set`, the registry defaults they read, and their tests (D-3, D-11) | STEP-3, AC-4.1 |
| Delete the name frequency discount end to end, `Stats.Words` and `countNameWords` included (D-4) | STEP-2, AC-3.3 |
| Delete `graphalign.Scale.Floor` (D-5) | STEP-3, AC-1.2 |
| `graphalign.Config.Value` replaces `Config.Names` (D-6) | STEP-3, AC-1.1 |
| `nameformats.Loaded.Rules` and `namealiases.Rules` put the catalog's pattern and alias tables into `Rules` (D-7, D-8) | STEP-3, AC-4.3 |
| `promotealign` (`Propose`, `loadCanon`, `pairExhibits`) and `core/database/matching` (`rank`) build and pass those rules (D-9, D-10) | STEP-3, AC-2.2 |
| Name-rule cases with no counterpart in `valuecompare` ported before `names_test.go` goes (D-11) | STEP-1 |
| Expected values updated where the design moves them, and only there (D-12) | STEP-3, AC-4.4 |
| `add-match-rule` names where catalog data enters `Rules` (D-13) | STEP-4 |
| `docs/matching.md`, `docs/promote-graph-alignment.md` §6, `docs/ideas/daitch-mokotoff.md` stop describing deleted code (D-14) | STEP-5 |

| Out | Home |
| --- | --- |
| Outcome read from the relation (`Equivalent` agrees; layer 2's outcome table); `ApplyScale` and `ScoreSimilarity` themselves | S11-09 builds it, S11-13 moves Promote onto it |
| The neutral-term check in `match.Compare` | S11-09 (layer 2, step 1) |
| `Stats.ValueFreq`, `MPrior`, `UPrior`, the frequency scales and their lift, `PropertyAgreement`, `BenchmarkLoadStatsNames` | S11-24 |
| `pairValues`' leaders and support weights; `poolPeriods`; `match.Value`'s reconciler fields; `Rank`'s weight set | S11-11, S11-10, S11-13 |
| `autoreconcile.Compatible` in `pairValues` and `exhibitCompatible` | S11-06 |
| `pairExhibits`' own member picker | S11-19 |
| Deleting `core/match`, `core/database/matching`, `docs/matching.md` | S11-13 |
| `core/graphalign/registry.go` numbers into `matchregistry` | S11-16 |
| Search, term dedupe and alias tokens on `valuecompare.Key` | S11-07 |
| Tuning any registry number to restore an old score | not planned: the working rule "Registry numbers: carry over, then estimate" |

## 4. Current state

Verified at `b6d1d48`.

| Area | Where | Today |
| --- | --- | --- |
| Comparer interface | `core/match/compare.go` `Comparer`, `DiscountComparer`, `CompareDiscount`, `ComparerFor`, `Set`, `setting`, `TextComparer`, `TermComparer`, `IntegerComparer` | `match`'s own layer 1: bare 0…1 score; `TermComparer` treats a `Neutral` term as not comparable |
| Name comparer | `core/match/names.go` `NamePatterns`, `NamePatternSet`, `NameComparer`, `nameWords`, `compareNames`, `familyConflict`; `core/match/frequency.go` `WordCounts`, `CountWords`, `wordDiscount`, `discount` | One-sided words are neutral (`Robins` / `James Robins` scores 1); `GivenOnlyFactor` 0.5 and `SuffixConflict` 0.3 are literals; the discount scales points by word frequency |
| Date comparer | `core/match/dates.go` `DateComparer`, `dateRules`, `dateSpan`, `spanOf`, `shares`, `spanScore`, `apart` | `ABT` doubles the tolerance (`ApproxToleranceFactor` 2) |
| Words | `core/match/words.go` `foldOnly`, `WordRules`, `wordRules`, `TextResemblance`, `compareForms`, `wordSimilarity`, `wordDice`, `bestPairing` | Text compares words in any order (Dice); a second word table beside `valuecompare.wordSimilarity` |
| Registry defaults | `core/match/registry.go` `value`, `SpacedParticles`, `LetterFolds`, `DefaultWordRules`, `AliasCredit`, `noAliases`, `MaxExactPairing`, `DefaultNames`, `WesternNamePattern`, `BuiltinNamePatterns`, `DefaultDates`, `DefaultText`, `DefaultIntegers`; keeps `PeriodProperties`, `DatePeriods`, `DefaultSuggestionLimit`, `DefaultProfile` | Comparer defaults read `matchregistry.Default().Value`; profiles name a comparer per Feature |
| Profiles | `core/match/profiles.go` `Feature.Comparer`, `Profile.WithNameComparer`, `With`, `Properties`, `product` | A Feature carries a comparer value |
| Scoring | `core/match/match.go` `Reason.Discount`, `Rank`, `Score` (`Weight × sim × (1 − Discount)`) | Reads the comparer from the Feature |
| Pairwise | `core/match/pair.go` `PropertyComparison.Discount`, `Evaluate`, `EvaluateWithNames`, `evaluate`, `compareProperty`, `outcomeAtFloor` | Picks `ComparerFor` or the passed `*NameComparer` |
| Pair selection | `core/match/pairs.go` `PairResult.Discount`, `pairValues`, `plainPair` (best `sim × (1 − disc)`), `valueType`, `compatible` | Calls `CompareDiscount` |
| Periods | `core/match/periods.go` `poolPeriods`, `period` | Builds a date range from start / end with `valuecompare.DateEnds`; unchanged here |
| Alignment config | `core/graphalign/registry.go` `DefaultConfig`, `Config.Names *match.NameComparer`, `Scale.Floor` (name 0.75, sex / gender 1, term / integer / name / subject value types 1, text / date 0) | Floors decide partial against conflict on top of the comparer |
| Alignment scoring | `core/graphalign/score.go` `ScoreSimilarity(similarity, discount, …)`, `ApplyScale`, `resolvedScale`; `core/graphalign/align.go` `state.evaluate` (`EvaluateWithNames`), `fallbackUnreachable` (`WithNameComparer`, `Rank`), `comparisonsFrom`; `core/graphalign/types.go` `Stats.Words` | — |
| Promote loader | `core/database/promotealign/promotealign.go` `Propose` (`loaded.Comparer()`, `namealiases.Comparer`, `names.Frequency = &stats.Words`, `cfg.Names`); `canon.go` `loadCanon(…, names match.NameComparer, namePattern)`, `Rank(…, seedCandidateLimit)` | — |
| Exhibits | `core/database/promotealign/exhibits.go` `pairExhibits` (`ComparerFor`, `cfg.Names`, `comparerDiscount`), `comparerDiscount`, `exhibitCompatible` | — |
| Stats | `core/database/promotealign/stats.go` `cloneStats` (`out.Words`), `computeStats` (`st.Words`), `sqlNameWords`, `countNameWords` | Name counts feed the discount only |
| Name formats | `core/database/nameformats/nameformats.go` `Loaded{Key, Profile, Set, Builtin}`, `Loaded.Comparer`, `Load` | Builds a `match.NameComparer` with the stored pattern |
| Alias tables | `core/database/namealiases/namealiases.go` `Comparer`, `withAliases` | Sets `NameComparer.Words.Aliases` |
| Old adapter | `core/database/matching/matching.go` `rank` (`CountWords`, `Frequency`, `WithNameComparer`, `Rank`), `candidateNames`, `loadCandidates` comment | No caller outside tests; deleted in S11-13 |
| Tests on `match`'s comparers | `core/match/names_test.go`, `dates_test.go`, `frequency_test.go`, `registry_test.go`; `core/match/match_test.go` `TestComparers`, `TestComparerFor`, `TestDiscountedScoring`; `core/match/pairs_test.go` (`TermComparer{}`, `DefaultNames`, `DefaultText`, `ComparerFor`); `core/match/pair_test.go` `TestEvaluateTextResemblance` (`TextResemblance`) | — |
| Tests on callers | `core/graphalign/align_test.go` (35 tests; `TestScoreSimilarityDiscount`, `TestScoreSimilarityWeightedCredit`, `TestPartialNamesAreWeak`, `TestMissingGivenOrPrefixIsNotANameConflict`, `TestGracieObituaryGolden`), `core/graphalign/pick_internal_test.go`; `core/database/promotealign/promotealign_test.go` (`TestStatsCountsNameWords`, `TestProposeWeighsACommonNameLess`), `exhibits_pin_test.go`, `bench_test.go` `BenchmarkLoadStatsNames`; `core/database/nameformats/nameformats_test.go` (`Loaded.Comparer`, `DefaultNames`); `core/database/namealiases/namealiases_test.go` (`Comparer`, `assertScore`, `AliasCredit`); `core/catalogsession/alias_cache_test.go` (`namealiases.Comparer`, `AliasCredit`); `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` | — |
| Value compare | `core/valuecompare/valuecompare.go` `Compare`, `Value`, `Result`; `core/valuecompare/rules.go` `Rules`, `NameRules.Patterns`, `NameRules.Aliases`; `core/valuecompare/name_test.go` `nameCases`, `withTables` | Built by S11-02 – S11-04; no production caller |
| Registry | `core/matchregistry/registry.go` `Default`; `core/matchregistry/value.go` `defaultValue` (`Names.NearMinimum` 0.75 from the name floor, `Dates.NearMinimum` 0 from the date floor) | — |
| Docs | `docs/matching.md` (file table L9 – L21, Scoring L41 – L52, Comparers and Names L54 – L167, Default profiles L169 – L181, Configuring L183 – L206, Extending L208 – L217); `docs/promote-graph-alignment.md` §6 L166 – L175; `docs/ideas/daitch-mokotoff.md` L5, L13, L123; map S11-24 scope entry | Describe the deleted code |

**Precedent.** S11-02 D-10 and S11-03 D-8 ([plans](S11-02-value-compare-core.md)): `match` was repointed at `valuecompare` helpers in the commit that moved them, with no copy left behind. S11-03's Landed note: the alias index is built once per load in `namealiases`, which this PR keeps (it now lands in `Rules.Names.Aliases`). Copy the shape "move every caller, delete the old code, in one commit per behavior".

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **A Feature names its value type.** `match.Feature` loses `Comparer` and gains `ValueType string` (a `propertymeta.ValueType*` constant). `DefaultProfile`: name → `ValueTypeName`; `sex_at_birth`, `gender`, `event_type` → `ValueTypeTerm`; `date`, `start_date`, `end_date` → `ValueTypeDate`; `toponym` → `ValueTypeText`. Weights, contradictions, cardinalities and `MinScore` stay. The person comment becomes: "`Mary Robins` ~ `James Robins` is `Different`: 0, hidden. `Robins` ~ `James Robins` is 10 × 0.75 = 7.5: shown." `Profile.WithNameComparer` is deleted; `With` and `Properties` stay | Value compare dispatches on the value type; the Feature already knows it (`TestDefaultProfileCardinality` reads the seed beside it) | Keeping a `Comparer` interface wrapping `valuecompare` — because the map says no comparer types remain in `core/match` |
| D-2 | **One adapter.** `core/match/compare.go` is rewritten to hold only `func Compare(rules valuecompare.Rules, valueType string, a, b Value) (similarity float64, comparable bool)` and the unexported `func (v Value) compareValue(valueType string) valuecompare.Value` (`Type` = `valueType`, and `Text`, `Term`, `Integer`, `Name`, `NamePattern`, `Date` copied; `Ref` nil). `Compare` is not comparable when `a.Neutral` or `b.Neutral`; when `a.valueType()` or `b.valueType()` (the existing method in `pairs.go`) is not `valueType`; or when the relation is `Unknown`. Otherwise it returns `Result.Similarity`. Its doc comment says the neutral check moves to property compare (S11-09) | One place turns a `match.Value` into a layer-1 value; `pairExhibits` and `pairValues` share it. `valueType()` also covers "an integer is always present" in `valuecompare.Value`: a value with no integer is not an integer value | Copying the conversion into `promotealign` — because a second conversion drifts; returning `Result` and `Detail` — because no caller reads them until S11-09 |
| D-3 | **Rules are an argument; `core/match` reads no number.** New signatures: `Rank(p Profile, rules valuecompare.Rules, probe Values, candidates []Candidate, limit int)`, `Score(p Profile, rules valuecompare.Rules, probe Values, c Candidate)`, `Evaluate(probe, candidate Values, metas []PropertyMeta, rules valuecompare.Rules)`, `pairValues(rules valuecompare.Rules, valueType, card string, probe, candidate []Value)`, `plainPair(rules, valueType, probe, candidate)`. `EvaluateWithNames`, `compareProperty`'s comparer choice and `match`'s `matchregistry` import are deleted. Deleted files: `core/match/names.go`, `dates.go`, `words.go`, `frequency.go`. `registry.go` keeps only `PeriodProperties`, `DatePeriods`, `DefaultSuggestionLimit`, `DefaultProfile` and a header comment saying the value numbers live in `core/matchregistry` and reach `match` as `Rules` | The working rule "each layer gets its section as an argument"; `match` is deleted in S11-13, so it gains no registry import to carry | Keeping `var value = matchregistry.Default().Value` in `match` — because callers that load a name format must pass their own rules anyway |
| D-4 | **The name frequency discount is deleted end to end.** `frequency.go`; `Reason.Discount`, `PropertyComparison.Discount`, `PairResult.Discount`; `ScoreSimilarity`'s `discount` parameter (new signature `ScoreSimilarity(similarity float64, comparable bool, cardinality string, property match.Property, valueType string, vals []match.Value, cfg Config, stats Stats)`); `graphalign.Stats.Words`; `promotealign` `sqlNameWords`, `countNameWords`, the `Words` lines in `computeStats` and `cloneStats`, `names.Frequency = &stats.Words`; `matching` `candidateNames` and its counting. `Score` adds `Weight × similarity`; `plainPair` and `pairValues` keep the highest `similarity × weights`. `BenchmarkLoadStatsNames` stays (S11-24 deletes it with the rest of the stats). The map's S11-24 scope already dropped `Stats.Words` and `countNameWords` when this plan was written | The map row says the discount goes here, and `Stats.Words` and `countNameWords` have no other reader once it does ("High tolerance for churn": no dead code); [frequency](../comparison-layers.md#frequency) | Leaving `Stats.Words` to S11-24 — because it would be counted and never read for a whole PR |
| D-5 | **Promote's floors are deleted.** `graphalign.Scale` loses `Floor`, and every `DefaultConfig` entry loses it. `ApplyScale`: not comparable → `unknown`, 0; similarity > 0 → `agree` at 1 or more, else `partial`, points `Weight × similarity`; otherwise `unknown` for a multiple-cardinality Property, else `conflict`, −`Contradiction`. `match`'s `outcomeAtFloor(sim, ok, card, floor)` becomes `outcomeOf(sim, ok, card)` with the floor-0 behavior it has today | The value layer's `Near` minimum is the old name floor (`value.Names.NearMinimum` 0.75 carries it, `value.Dates.NearMinimum` 0 the date floor); term and integer similarities are only 0 or 1 at today's tolerance 0, so their floor 1 changed nothing. Keeping the name floor would turn every `Contains` under 0.75 (`Robins` / `James Kenneth Jake Robins`, 0.6) into a conflict, which [outcome from relation](../comparison-layers.md#outcome-from-relation) calls partial. `matching-model.md` "Removed knobs" already lists the floors | Keeping the floors until S11-13 — because they would contradict the relation the comparer just returned |
| D-6 | **Alignment carries the value rules.** `graphalign.Config.Names *match.NameComparer` becomes `Value valuecompare.Rules`, doc comment "the value section, with the catalog's name patterns and alias tables". `DefaultConfig()` sets `Value: matchregistry.Default().Value`. `state.evaluate` calls `match.Evaluate(a, b, st.metas, st.cfg.Value)`; `fallbackUnreachable` calls `match.Rank(profile, st.cfg.Value, s.Values, …)`. `pairExhibits` reads `cfg.Value` | One argument instead of a name-only override; `core/graphalign/registry.go` is already the defaults file for the old walk, and S11-16 deletes it and passes registry sections from the caller, so the `matchregistry` import goes with it | A pass-through `match.DefaultRules()` — because it adds an export to a package S11-13 deletes; leaving `Value` zero in `DefaultConfig` — because every `Align(…, nil)` test would compare under zero weights |
| D-7 | **Name formats fill `Rules`.** `nameformats.Loaded` loses `Set` and `Comparer()`. New `func (l Loaded) Rules(base valuecompare.Rules) valuecompare.Rules`: a builtin load returns `base` unchanged; otherwise it returns `base` with `Names.Patterns` = `map[string]valuecompare.NamePattern{l.Key: l.Profile.Pattern}`. Values keep carrying `NamePattern = loaded.Key` from their loaders (`handleValues`, `loadExhibitObservations`, `matching` scans), so a name finds its stored pattern by key and any other key reads `Names.Western` | The layer reads patterns from `Rules.Names.Patterns` (`nameWords`); this is S11-03's "catalog data comes in through `Rules`" | A `NamePatterns` source interface in `valuecompare` — because one stored default format is all the catalog has |
| D-8 | **Alias tables fill `Rules`.** `namealiases.Comparer(g, q, base match.NameComparer)` becomes `func Rules(g *graphcache.Graph, q Querier, base valuecompare.Rules) (valuecompare.Rules, error)`; `withAliases` sets `base.Names.Aliases = valuecompare.NewAliasIndex(tables)`. The warmed-graph and cold-read paths stay as they are | S11-03's index-once-per-load, landed in `namealiases`; the same data, a new destination | — |
| D-9 | **Callers build the rules once per run.** `promotealign.Propose`: `rules, err := namealiases.Rules(c.Graph(), q, loaded.Rules(matchregistry.Default().Value))`, then `cfg.Value = rules` and `loadCanon(c.Graph(), q, layer, anchors, rules, loaded.Key)` (parameter `names match.NameComparer` becomes `rules valuecompare.Rules`), which calls `match.Rank(profile, rules, s.Values, cands, seedCandidateLimit)`. `matching.rank`: `namealiases.Rules(nil, q, loaded.Rules(matchregistry.Default().Value))` and `match.Rank(profile, rules, probe, kept, opts.Limit)`; `loadCandidates`' comment loses its name-word sentence. The stats load stays before `loadCanon` only for the fan-outs; its comment "seed ranking in loadCanon reads the name counts" is deleted | Two call sites, the same three calls; `core/database/matching` is deleted in S11-13, so a shared helper would outlive one of its two callers by a single PR | A helper in `namealiases` taking a `nameformats.Loaded` — because it would make `namealiases` import `nameformats` for one line |
| D-10 | **Exhibits compare through the adapter.** `pairExhibits` calls `match.Compare(cfg.Value, in.valueType, in.value, mem.value)` where it called `comparerDiscount`, keeps the first `exhibitCompatible` member (pinned) and otherwise the member with the highest similarity, and calls the new `ScoreSimilarity` (D-4). `comparerDiscount` is deleted. The member picker itself is unchanged | The map row: `pairExhibits` uses value compare after S11-05 and is deleted in S11-19 | Rewriting the picker on relations — because S11-19 replaces it |
| D-11 | **Tests that tested deleted code go; their uncovered cases move first.** Deleted whole: `core/match/names_test.go`, `dates_test.go`, `frequency_test.go`, `registry_test.go`. Deleted tests: `match_test.go` `TestComparers`, `TestComparerFor`, `TestDiscountedScoring`; `align_test.go` `TestScoreSimilarityDiscount`; `promotealign_test.go` `TestStatsCountsNameWords`, `TestProposeWeighsACommonNameLess`. Before that, STEP-1 adds `TestNameRulesComeFromRules` to `core/valuecompare/name_test.go` with the four name cases `match` tests that `valuecompare` does not (T-1). `match`'s own tests (`pairs_test.go`, `pair_test.go`, `match_test.go`, `periods_test.go`) build rules with `matchregistry.Default().Value` (test files may import it) through one helper `rules()` in `match_test.go` *(new)* for `package match_test` and one in `pairs_test.go` for `package match`; `pairs_test.go` passes value types where it passed `TermComparer{}`, `DefaultNames`, `DefaultText` or `ComparerFor(…)`. `TestEvaluateTextResemblance` asserts the partial's similarity is `valuecompare.Compare`'s for the same pair instead of `TextResemblance`'s. `TestDefaultProfileCardinality` also checks each Feature's `ValueType` against the seeded Property's (T-6) | The S11-03 and S11-04 plans ported every ladder, table and invariant of `names_test.go` and `dates_test.go` onto relations (their D-10 / D-16); four rule-source cases were not | Keeping `match`'s comparer tests against the adapter — because they would test `valuecompare` twice under other names |
| D-12 | **Expected values move only where §8 "Expected changes" says.** Each row names the test, the old value, the new one, and the design rule behind it. A failing assertion that is not a row is a Stop condition | "Goldens are updated, not preserved" applies to changes the design makes; anything else is a bug in the migration | Updating whatever fails — because it would hide a migration bug as a design change |
| D-13 | **The matching skill names where catalog data enters `Rules`.** `add-match-rule` "Rules that hold everywhere", bullet "Catalog data comes in through `Rules`", gains: "Promote and `Rank` fill them with `nameformats.Loaded.Rules` (the stored default pattern under its key) and `namealiases.Rules` (the alias index, built once per load) on a copy of `matchregistry.Default().Value`." | The next caller (S11-06, S11-07) needs the same three calls | — |
| D-14 | **Docs that describe deleted code are corrected in this PR.** `docs/matching.md`: the file table lists `compare.go` as "`Compare`: one value pair through `core/valuecompare` under the caller's `Rules`" and drops `names.go`, `dates.go`, `words.go`; Scoring step 3 reads "adds `Weight × similarity`" and the discount sentences go; "Comparers" and "Names: part types are data, not gates" through the name tests paragraph (L54 – L157 up to "Support weights and negatives") become one section "Comparing values": values compare with [`core/valuecompare`](../../../../core/valuecompare) under the registry's value section ([`matching-model.md`](../../../matching-model.md) layer 1 and knob inventory); a `Neutral` term and an `Unknown` relation are not comparable; the catalog's pattern and alias tables arrive in `Rules`; "Name frequency" is deleted; "Support weights and negatives" stays, with `(1 − discount)` removed from Credit; the L165 – L167 paragraphs say `valuecompare` instead of `NameComparer` and drop `ComparerFor`; Default profiles' worked example reads "`Mary Robins` against `James Robins` is `Different`: 0, hidden; `Robins` against `James Robins` is 7.5"; Configuring keeps only `Profile.With`, the registry sentence (value numbers in `core/matchregistry`, profiles in `registry.go`), the neutral-term sentence (checked in `match.Compare`) and "Not yet"; Extending loses the name-word count sentences. `docs/promote-graph-alignment.md` §6: L166 reads "`match.Compare` returns a similarity from value compare"; the two floor bullets become "A positive similarity contributes `weight × similarity`; 1 agrees, less is partial. A comparable 0 conflicts for a single-value property and is unknown for a multiple-value one."; L175's name sentence becomes "Name: weight 2.5, contradiction 3. An identical name is 2.5, still under medium. A name with words on one side only is a partial (`Jake Robins` against `James Kenneth Jake Robins` is 2.5 × 0.75)." with the discount and "agrees" sentences gone. `docs/ideas/daitch-mokotoff.md`: L5 says S11-05 deleted `match`'s copy, L13 links `../../core/valuecompare/words.go`, L123 names `valuecompare.WordRules` and `core/matchregistry/value.go` | Working rule "One matching skill, grown with the layers" (each migration PR fixes the docs it makes false); S11-01 D-5 leaves the full rewrite and deletion to S11-13 and S11-16 | Rewriting `matching.md` against the model now — because S11-13 deletes it |
| D-15 | **Commits:** (1) STEP-1's test; (2) STEP-2, the discount; (3) STEP-3, value compare everywhere, with its expected values; (4) STEP-4 and STEP-5, skill and docs; (5) STEP-7, plan and map. Each commit builds and passes `go test -tags fts5 ./core/... ./api/...` | One behavior per commit; STEP-3 cannot be split without a commit where callers and `match` disagree on a signature | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `match.Compare` *(new)* | func | `match.CompareDiscount`, `ComparerFor` | Replaces both (D-2); the conversion to `valuecompare.Value` has no other home while `match.Value` carries layer-2 fields |
| `Value.compareValue` *(new)* | func | `Value.Auto` (`pairs.go`), the conversion to `autoreconcile.Value` | Same shape for the other layer; `Auto` stays for `Compatible` until S11-06 |
| `Feature.ValueType` *(new)* | field | `PropertyMeta.ValueType` | The same string `Evaluate` already reads; `Feature` keeps its own type until S11-10 |
| `outcomeOf` *(new name)* | func | `outcomeAtFloor` | Renamed: its only floor was 0 (D-5) |
| `graphalign.Config.Value` *(new)* | field | `Config.Names` | Replaces it (D-6) |
| `nameformats.Loaded.Rules` *(new)* | func | `Loaded.Comparer` | Replaces it (D-7) |
| `namealiases.Rules` *(new)* | func | `namealiases.Comparer` | Replaces it (D-8) |
| `TestNameRulesComeFromRules` *(new)* | test | `core/valuecompare/name_test.go` `TestNameCompare` | Its table asserts exact similarities; these four cases assert only that a rule's source is honored (D-11) |
| `rules()` test helpers *(new)* | helper | `core/valuecompare/name_test.go` builds rules from `matchregistry.Default().Value` inline | One per test package in `core/match`; no shared test package for two files (D-11) |

Searches run for this table: `rg -n 'func Compare\b|CompareDiscount|ComparerFor' core` (only `core/match/compare.go` and `core/valuecompare`), `rg -n 'Loaded\) (Comparer|Rules)' core` (only `Comparer`), `rg -n 'func (Comparer|Rules)\(' core/database/namealiases` (only `Comparer`), `rg -n 'outcomeAtFloor' core` (only `pair.go`).

**Builds on.** `valuecompare.Compare` and `Rules` (S11-02 – S11-04); `matchregistry.Default()`; the alias index built once per load in `namealiases` (S11-03); the `nameformats` loader.
**Deletes.** `core/match/names.go`, `dates.go`, `words.go`, `frequency.go`, `names_test.go`, `dates_test.go`, `frequency_test.go`, `registry_test.go`; from `compare.go` every comparer type, `Comparer`, `DiscountComparer`, `CompareDiscount`, `ComparerFor`, `Set`, `setting`; from `registry.go` every default but the four D-3 keeps; `Profile.WithNameComparer`; `EvaluateWithNames`; every `Discount` field; `graphalign.Scale.Floor`, `Config.Names`, `Stats.Words`; `promotealign` `countNameWords`, `sqlNameWords`, `comparerDiscount`; `matching` `candidateNames`; `nameformats.Loaded.Set`, `Loaded.Comparer`; `namealiases.Comparer`; the six deleted tests in D-11; the name-frequency and comparer sections of `docs/matching.md`.
**Next adds.** S11-06 calls `valuecompare` for the reconciler with the same `Loaded.Rules` / `namealiases.Rules` pair. S11-09 takes the neutral check out of `match.Compare` into property compare. S11-13 replaces `match.Compare`'s callers with entity compare and deletes `core/match`. S11-24 deletes the remaining frequency fields beside the gap `Stats.Words` leaves.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/match` | `compare.go` rewritten (D-2); `names.go`, `dates.go`, `words.go`, `frequency.go` deleted; `registry.go` trimmed (D-3); `profiles.go` `Feature.ValueType`, no `WithNameComparer` (D-1); `match.go` `Rank`, `Score` take rules, no `Discount` (D-3, D-4); `pair.go` `Evaluate` takes rules, `outcomeOf` (D-3, D-5); `pairs.go` `pairValues`, `plainPair` take rules and a value type (D-3) |
| `core/graphalign` | `registry.go` `Config.Value`, no `Floor`, no `Names` (D-5, D-6); `score.go` `ScoreSimilarity`, `ApplyScale` (D-4, D-5); `align.go` `evaluate`, `fallbackUnreachable`, `comparisonsFrom` (D-6); `types.go` no `Stats.Words` (D-4) |
| `core/database/promotealign` | `promotealign.go`, `canon.go` (D-9); `exhibits.go` (D-10); `stats.go` (D-4) |
| `core/database/matching` | `matching.go` `rank` (D-9) |
| `core/database/nameformats` | `nameformats.go` `Loaded.Rules` (D-7) |
| `core/database/namealiases` | `namealiases.go` `Rules` (D-8) |

```go
// Compare is a and b, values of valueType, under rules: the similarity value
// compare gives, or not comparable when either value is neutral, either is
// not a value of valueType, or the relation is Unknown. The neutral check
// moves to property compare (S11-09).
func Compare(rules valuecompare.Rules, valueType string, a, b Value) (similarity float64, comparable bool)

type Feature struct {
	Property      Property
	ValueType     string // propertymeta.ValueType*
	Weight        float64
	Contradiction float64
	Cardinality   string
}

func Rank(p Profile, rules valuecompare.Rules, probe Values, candidates []Candidate, limit int) []Match
func Score(p Profile, rules valuecompare.Rules, probe Values, c Candidate) Match
func Evaluate(probe, candidate Values, metas []PropertyMeta, rules valuecompare.Rules) Evaluation

func (l Loaded) Rules(base valuecompare.Rules) valuecompare.Rules           // nameformats
func Rules(g *graphcache.Graph, q Querier, base valuecompare.Rules) (valuecompare.Rules, error) // namealiases
```

Errors: none new. `namealiases.Rules` returns the errors `Comparer` returned.

### FFI

None — no wire change. `Exhibit.Weight` and `Comparison.Weight` keep their fields; their values follow §8.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Port the name-rule cases**
Files: `core/valuecompare/name_test.go`. Skill: `add-match-rule`.
Do: T-1 (D-11).
Check: `go test -run TestNameRulesComeFromRules -count=1 ./core/valuecompare/` passes with no tags.

**STEP-2 — Delete the name frequency discount**
Files: `core/match/frequency.go`, `frequency_test.go`, `names.go`, `match.go`, `pair.go`, `pairs.go`, `compare.go`, `registry.go`, `match_test.go`; `core/graphalign/score.go`, `align.go`, `types.go`, `align_test.go`; `core/database/promotealign/promotealign.go`, `stats.go`, `exhibits.go`, `promotealign_test.go`; `core/database/matching/matching.go`.
Do: D-4. In this step `NameComparer` loses `Frequency`, `FrequencyDiscount`, `FrequencyMinPeople`, `CountWords`, `wordDiscount`, `discount` and its `CompareDiscount` method; `DiscountComparer` and `CompareDiscount` go; `pairValues` calls `cmp.Compare`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; the only changed expected values are §8 rows marked STEP-2; `rg -n 'Discount|WordCounts|CountWords|countNameWords|sqlNameWords|Stats\.Words|\.Words = ' core/match core/graphalign core/database/promotealign core/database/matching` prints nothing (AC-3.3).

**STEP-3 — Compare values with value compare**
Files: every `core/match` file (D-1 – D-3, D-5); `core/graphalign/registry.go`, `score.go`, `align.go`, `align_test.go`, `pick_internal_test.go`; `core/database/nameformats/nameformats.go`, `nameformats_test.go`; `core/database/namealiases/namealiases.go`, `namealiases_test.go`; `core/database/promotealign/promotealign.go`, `canon.go`, `exhibits.go`, `layer.go` if it names a comparer, `promotealign_test.go`, `exhibits_pin_test.go`; `core/database/matching/matching.go`, `matching_test.go`; `core/catalogsession/alias_cache_test.go`; `core/database/subjectvocab/profile_test.go`. Skill: `add-match-rule`, `add-name-value`.
Do, in this order, then build: D-7, D-8 (catalog data into `Rules`); D-1 – D-3 and D-5 in `core/match`; D-5, D-6 in `core/graphalign`; D-9, D-10 in the loaders; D-11's test deletions and helpers; T-2 – T-6; then run the suite and update each expected value listed in §8 "Expected changes", nothing else.
Check: `go build ./...` and `go vet -tags fts5 ./core/... ./api/...` pass; `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; AC-4.1's `rg` prints nothing; `rg -n 'Floor' core/graphalign` prints nothing; `git diff b6d1d48 --stat -- core/valuecompare core/matchregistry` lists only `core/valuecompare/name_test.go` (from STEP-1).

**STEP-4 — The matching skill**
Files: `.agents/skills/add-match-rule/SKILL.md`.
Do: D-13.
Check: `rg -n 'namealiases.Rules' .agents/skills/add-match-rule/SKILL.md` prints one line.

**STEP-5 — Docs that named the deleted code**
Files: `docs/matching.md`, `docs/promote-graph-alignment.md`, `docs/ideas/daitch-mokotoff.md`.
Do: D-14.
Check: `rg -n 'NameComparer|DateComparer|TextComparer|ComparerFor|CompareDiscount|Frequency|discount|match\.Set|names_test|core/match/words.go|core/match/registry.go#|DefaultWordRules' docs/matching.md docs/promote-graph-alignment.md docs/ideas/daitch-mokotoff.md` prints nothing.

**STEP-6 — Full gate**
Files: none.
Do: run the gate on the user's Mac, as `ship-pr` requires (`xcode-test` cannot pass elsewhere).
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Docs**
Files: `docs/deployment-plan/spike-11/deployment-plan.md` (map row), `docs/deployment-plan/spike-11/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-05-migrate-matching.md` reports no errors.

Commits: one per behavior (D-15). The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Compare Promote's values with value compare and drop the per-type floors."). No WIP commits.

## 8. Tests

New and changed tests are table-driven where they hold more than one case (`.cursor/rules/go-tests.mdc`). Expected values are written as the arithmetic.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/valuecompare/name_test.go` `TestNameRulesComeFromRules` *(new)* | Under `Default().Value` unless named: `s=O. Brien` / `s=O'Brien` → not `Same` (a period initial never rejoins); `Words.Particles` empty: `s=O Brien` / `s=O'Brien` → not `Same`, and with the default particles `Same`; `Words.Folds` empty: `g=José` / `g=Jose` → `Equivalent`, cause `fold` (accents still strip), and `s=Groß` / `s=Gross` → neither `Same` nor `Equivalent`; alias tables from `withTables(nicknameTable)`: `g=Xyzzy` / `g=Plugh` → `Different`, and with no tables `g=Jas.` / `g=James` → cause not `abbreviation` | D-11 |
| T-2 | Go | `core/match/match_test.go` `TestCompare` *(new, replaces `TestComparers`)* | Under `rules()`: term `male` / `male` → 1, comparable; `male` / `female` → 0, comparable; neutral `unknown` / `male` → not comparable; text `York` / `york` → 1; a name value against a text value as `text` → not comparable; integer value with `HasInteger` false → not comparable; a phrase-only date / `1817` → not comparable; `James Robins` / `James Kenneth Robins` as `name` → 5 / 5.5 | AC-4.2, D-2 |
| T-3 | Go | `core/match/match_test.go` `TestRankPersons`, `TestRankEventsAndPlaces`, `TestScoreDatePeriod` (changed signatures) | AC-1.5's ranking; AC-2.1's event score | AC-1.5, AC-2.1, D-1, D-3 |
| T-4 | Go | `core/match/pair_test.go` `TestEvaluateTextResemblance` (changed) | `Provenance` / `Provenanced` partial at `valuecompare`'s similarity; `York` / `York, Upper Canada` multiple → 0.35; `York, Upper Canada` / `Upper Canada, York` single → conflict | AC-3.1, AC-3.2 |
| T-5 | Go | `core/graphalign/align_test.go` `TestPartialNamesAreWeak`, `TestMissingGivenOrPrefixIsNotANameConflict`, `TestUnreachableFallbackScoresOnTheWalkScale` (expected values per "Expected changes") | AC-1.1 – AC-1.4 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-5 |
| T-6 | Go | `core/database/subjectvocab/profile_test.go` `TestDefaultProfileCardinality` (extended) | Every `DefaultProfile` Feature's `ValueType` equals its seeded Property's `ValueType` | D-1 |
| T-7 | Go | `core/database/promotealign/exhibits_pin_test.go`; `core/database/nameformats/nameformats_test.go` `TestMissingDefaultUsesWestern`, `TestInstallLoadsWesternAndFormats`; `core/database/namealiases/namealiases_test.go` (every test that called `Comparer`; its `assertScore` helper compares through `valuecompare.Compare` under the returned rules and checks the similarity); `core/catalogsession/alias_cache_test.go` (changed to `Rules`) | AC-2.2; the builtin load's `Rules(Default().Value)` equals `Default().Value` and compares `Vincent van Gogh` the same; the installed load's `Names.Patterns["western"]` is the stored pattern; AC-4.3 through `valuecompare.Compare` | AC-2.2, AC-4.3, D-7, D-8 |
| T-8 | Go | The whole Go suite | Passes; every moved expected value is a row below | AC-3.3, AC-4.1, AC-4.4, D-12 |

**Expected changes.** EXPECTED-CHANGES-PLACEHOLDER

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/graphalign` (35 tests, the golden), `core/database/promotealign` (scores read from a real catalog), `core/match` `pairs_test.go` (support and negatives on new similarities), `TestLayerPackages` (no layer package may import `core/match`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI in this PR
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: matching.md, promote-graph-alignment.md §6, ideas/daitch-mokotoff.md (STEP-5); add-match-rule (STEP-4)
- [ ] PR description: what shipped, AC evidence, deviations, every "Expected changes" row with its rule, manual check
```

## 10. Manual check

The reviewer runs, from the repo root:

1. AC-4.1's `rg` and STEP-2's `rg`: no output.
2. `CGO_ENABLED=1 go test -tags fts5 -run 'TestPartialNamesAreWeak|TestMissingGivenOrPrefixIsNotANameConflict|TestGracieObituaryGolden' -v ./core/graphalign/`: passes (AC-1.1, AC-1.2).
3. In the app, open a project with a filed person `James Kenneth Robins` and a Source holding a person `James Robins`; open Promote on that Source. The row proposes `James Kenneth Robins`, weak. A Source person `Mary Robins` proposes no one (AC-1.1, AC-1.4).
4. On a Source with a birth dated `ABT 1817` beside a filed birth dated `1818` for the same person, the evidence sheet shows the date line as a resemblance, unpinned (AC-2.1).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A support-weighted or negative case in `pairs_test.go` moves because a similarity changed under it, and the update hides a `pairValues` regression | D-12: only rows in "Expected changes" move; each names its rule |
| A Promote test passes by an accident of the old floor (a `Contains` under 0.75 now partial) | STEP-3's run lists every changed assertion; the table was measured on a prototype of this exact design |
| `graphalign` importing `core/matchregistry` blocks a later layer | S11-16 deletes `registry.go`; D-6 records it; `TestLayerPackages` does not list `core/graphalign` |
| The `rules()` helpers drift from `Default()` | They return `matchregistry.Default().Value` and nothing else |

**Stop and report if:**

- A row in §4 is no longer true at your base (for example `match.NameComparer` is already gone or `namealiases.Comparer` has another caller).
- An expected value changes that is not a row of "Expected changes", or a row's new value differs from what you get.
- A test outside D-11's list fails only because a deleted symbol is gone and no new signature serves it.
- `core/valuecompare` or `core/matchregistry` needs a change beyond STEP-1's test (a new knob, a changed number, a new cause).
- Any package under `core/valuecompare`, `core/matchregistry`, `core/datevalue`, `core/namevalue`, `core/propertymeta` or `core/vocabkey` would import `core/match` or `core/graphalign`.

## 12. Open questions

None.

## 13. PR

**Title:** S11-05 — Migrate matching onto value compare
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence (test names); the "Expected changes" rows with their rules; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-11/prs/S11-05-migrate-matching.md`.
