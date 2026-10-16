# S10-27 — Name frequency in this project

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-27--name-frequency)
**Brief:** — (no new chrome: the score changes, no view does)
**Board:** —
**Depends on:** S10-24 (landed [#378](https://github.com/mendahu/provenencia/pull/378))
**Blocks:** — (S10-28 sequences after S10-24 and touches the same comparer and fixtures; it does not wait on this PR)
**Branch:** `s10-27-name-frequency`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`name-matching-enhancements.md`](../name-matching-enhancements.md), [`docs/matching.md`](../../../matching.md), [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §6, working rules "Name frequency is this project's inverse frequency" and "Accent fold is matching only" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-16 against `main` @ `1fc683b`
**Review:** [`S10-27-name-frequency.review.md`](S10-27-name-frequency.review.md) — Fits with changes / Ready with changes, 1 round, no open blocker

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

A shared surname that is rare in the researcher's own project counts for more than a shared common one. Today "John Smith" against "John Smith" earns the same points as "John Quirk" against "John Quirk", and that is the main source of false suggestions in Promote. After this PR a word that many filed people carry earns up to 20% fewer points, and a word that one person carries earns all of them.

**Why now.** Slice 7 makes names match the way records are written. S10-24 and S10-25 widened what counts as the same word. This PR narrows how much a common word proves. It comes before S10-28 so support weighting and negative evidence are tuned once against frequency-weighted fixtures.

**What stays.** A name's similarity. An identical name is still similarity 1, still `agree`, and still pins in Promote. Display, `NormalizeForm`, `sort_key`, and the auto-reconciler are untouched. Callers with no counts (`DefaultNames`, `NameComparer{}`) score exactly as before: counts are data a caller supplies.

## 2. User stories

**US-1.** As a researcher with many filed people, when Promote compares an incoming "John Quirk" to a filed "John Quirk" and another incoming "John Smith" to a filed "John Smith", I see the Quirk suggestion rank above the Smith one, so that a common surname alone does not look like a confident match.

Counts for AC-1.1 and AC-1.2: People 100; `smith` on 40 handles, `quirk` on 1, `john` on 10. A word's discount is `0.2 × ln(df) / ln(100)`: `john` 0.100, `smith` 0.160, `quirk` 0.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | The counts above. `John Quirk` vs `John Quirk`, then `John Smith` vs `John Smith`, ranked by `DefaultProfile("person")` | Both name similarities are 1. Discounts are 0.040 and 0.136. The contributions are 9.60 and 8.64 |
| AC-1.2 | The counts above, Promote's fixed name scale (weight 2.5) | `agree`, pinned, weight 2.40 for Quirk and 2.16 for Smith |
| AC-1.3 | The project has no filed person with a name | Discount is 0 and every score equals today's |
| AC-1.4 | One household: 8 filed people, all with `robins` | `Robins` vs `Robins` has discount 0.090, not 0.2 |
| AC-1.5 | `J. Robins` vs `James Robins` with `robins` on 40 handles and `james` on 10 | The initial has discount 0. Name discount is 0.116 |
| AC-1.6 | Names `José` and `Jose` | Both count under one folded key |
| AC-1.7 | A handle carries `Smith` in two name values, or is merged into another handle | It counts once for `smith`. A merged handle counts for nothing |
| AC-1.8 | A handle commit changes a name | The next Promote read recounts. A read with no commit between reads the count once |
| AC-1.9 | The 60-person fixture of T-10 | The Quirk Subject's top candidate scores above the Smith Subject's top candidate in `ForSubject` |
| AC-1.10 | A candidate whose name alone scored between 3 and 3.75 (surname-only match, a surname on every filed person) | It drops below `MinScore` 3 and is not returned. Deliberate: a common surname alone no longer clears the bar |

**US-2.** As a contributor, when I score two names with a bare `NameComparer{}`, I get today's numbers, so the 0–1 invariants in `names_test.go` keep holding.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `NameComparer{}` and `DefaultNames` score every case in `nameCases` | Unchanged from the current values |
| AC-2.2 | `TestRegistrySetsEverySetting` runs | Passes. Each new setting has a registry value |

## 3. Scope

| In | Covered by |
| --- | --- |
| Count name words over the filed people of a project, one count per handle per word, scoped to Person handles | STEP-1, STEP-5, AC-1.6, AC-1.7 |
| A discount from a word's count, scaled by role weight, applied to points and not to similarity | STEP-2, STEP-3, STEP-4, AC-1.1, AC-1.2, AC-1.5 |
| A floor for small projects (a minimum population) so the first Smith is not rare forever and an empty project still matches | STEP-2, AC-1.3, AC-1.4 |
| Promote reads the counts from `graphalign.Stats` (computed once, dropped by a handle commit); merge-hint ranking counts the candidates it already loaded | STEP-5, STEP-6, AC-1.8, AC-1.9 |
| The `MinScore` cut sees the discounted score (D-11) | STEP-3, AC-1.10 |
| Docs: `matching.md`, `promote-graph-alignment.md`, the map, `completed.md` | STEP-7 |

| Out | Home |
| --- | --- |
| A seeded population table | [`docs/ideas/name-matching-enhancements.md`](../../../ideas/name-matching-enhancements.md) |
| Counting unfiled Subjects' names or every name Observation | not planned (D-2) |
| Frequency for place, event, text, or date values | not planned. Promote's `u` frequency scales already cover text and term values |
| Support-weighted scores and negative evidence | S10-28 |
| A setting or UI to change the discount | not planned |
| A migration, a stored count table, or a new graph slot | not planned (D-3) |
| Counts by name role | not planned (D-4) |

## 4. Current state

Verified at `1fc683b`.

| Area | Where | Today |
| --- | --- | --- |
| Name comparer | `core/match/names.go` `NameComparer.compareNames` | Pairs words one-to-one by best assignment of `similarity × affinity × (weightX + weightY)`, divides by the counted weight. `totalEarned, usedA, usedB := bestAssignment(earned)`: `usedA` indexes `wa`, `usedB` indexes `wb`, and only positive-scoring cells are marked used. Knows nothing about how common a word is |
| Name comparer settings | `core/match/registry.go` `DefaultNames`; `core/match/registry_test.go` `TestRegistrySetsEverySetting` | The test fails on a nil pointer or map in `DefaultNames` (it does not recurse into pointers). A new pointer field needs a registry value |
| Comparer interface | `core/match/compare.go` `Comparer.Compare(a, b Value) (float64, bool)` | Similarity only. `NameComparer` has a value receiver; `Feature.Comparer` holds the value |
| Point scoring | `core/match/match.go` `Score`, `best`, `Rank` | `best` returns `(float64, bool)`. `Contribution = Weight × Similarity`, or `−Contradiction` at 0. `Rank` keeps `m.Score >= p.MinScore` |
| Pairwise evaluation | `core/match/pair.go` `Evaluate`, `EvaluateWithNames`, `evaluate`, `compareProperty`, `PropertyComparison` | `Evaluate` passes a nil names comparer, so it scores names with `DefaultNames`. `EvaluateWithNames` takes a comparer. `PropertyComparison` is built only in `evaluate` |
| Promote scale | `core/graphalign/score.go` `ScoreSimilarity`, `ApplyScale`; `core/graphalign/registry.go` name scale `{Floor: 0.75, Weight: 2.5, Contradiction: 3}` | Points are `Weight × similarity`. Outcome `agree` needs similarity 1 |
| Scale call sites | `core/graphalign/score.go` `propertyPoints` (line 38), `core/graphalign/align.go` `comparisonsFrom` (line 885), `core/database/promotealign/exhibits.go` (line 221, with `comparerSimilarity` at line 255) | The only three callers of `ScoreSimilarity`; no test calls it. `exhibits.go` picks the member whose similarity `s` is highest and pins when `outcome == match.OutcomeAgree` |
| Ranking readers | `core/database/promotealign/canon.go` (line 108, `match.Rank` over seeds), `core/graphalign/align.go` (line 421, `match.Rank` for unreachable Subjects), `core/database/matching/matching.go` `rank` | All receive the name comparer built once per read |
| Comparer build | `core/database/promotealign/promotealign.go` `Propose` (`nameformats.Load`, `namealiases.Comparer`, `loadCanon`, then `loadStats`); `core/database/matching/matching.go` `rank` (`loadCandidates` runs first) | `Propose` builds names before stats. `rank` has no graph and has already loaded every candidate's name blobs |
| Stats | `core/graphalign/types.go` `Stats` (`ValueFreq`, `FanOut`); `core/database/promotealign/stats.go` `loadStats`, `cloneStats`, `computeStats`; `core/database/graphcache/graph.go` `Stats`, `SetStats`, `clear`; `core/database/graphcache/commit.go` `OnCommit` | Counted once, kept on the graph, cleared by a handle commit and a structure drop. `computeStats` queries `auto_reconciler_values` joined to unmerged `canonical_entities`. `TestStatsCacheInvalidatesOnWrite` covers the drop |
| Handle names | `auto_reconciler_values.value_name` blob; `core/valuecodec/valuecodec.go` `UnmarshalName`, `MarshalName` | `matching.go` `scanEntityValues` decodes the same blob |
| Tests and bench | `core/match/names_test.go` `nameCases`; `core/database/matching/matching_test.go` `newFixture`, `handle`, `named`, `TestForSubjectQueryCountIsConstant`; `core/database/promotealign/promotealign_test.go` `newFixture`, `person`, `promote`, `propose`; `core/database/promotealign/bench_test.go` `bulkCanon` | Green at the base. `bulkCanon` inserts entities and no name rows |

**Precedent.** `loadStats` in `core/database/promotealign/stats.go` is the shape: derived from `auto_reconciler_values`, computed on the first read, kept on the open catalog's graph, dropped by a handle commit. Copy that, including `cloneStats`. Do not copy S10-25's `Warm` at project open ([#379](https://github.com/mendahu/provenencia/pull/379)): the count is only read by Promote.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Frequency scales **points**. The comparer's similarity is unchanged: an identical name stays 1, `agree`, and pinned. | `exhibits.go` and `align.go` pin only on `agree`, and `agree` needs similarity 1. A lower similarity would unpin every common identical name. User answer 2026-10-16. | Lowering the similarity — unpins identical common names. Boosting rare names above baseline — leaves common names where they are, and the aim is fewer false suggestions on common names |
| D-2 | The population is the unmerged Person handles' reconciled name values in `auto_reconciler_values` for the product `name` Property, every rank. A handle counts once per word. User answer 2026-10-16. | Matches how Promote's `stats` already measures frequency (handles). Transcription volume does not move it. Scoping to the person type keeps a researcher-defined type with a `name` Property out of the count. | Every name Observation — ten records of one person would make a surname look common |
| D-3 | Promote's counts are a field of `graphalign.Stats` (`Words match.WordCounts`), filled by `computeStats`. No migration, no stored table, no new graph slot. `matching.rank` counts the candidates it already loaded, before the exclude filter, and only when the profile has a name Feature. | `Stats` already has the lifetime and invalidation this needs. `rank` has no graph and already holds every name, so a second scan is waste. | A `namefrequency` package with its own graph slot — a second invalidation owner. A count table rebuilt with the search index — a second derived store. Reading `g.Kind` nodes — loads every Person into memory to count words, which `computeStats`'s SQL avoids |
| D-4 | A word's key is its comparison token: `splitWords` output (particles rejoined) passed through `foldWord` with the comparer's folds. Initials (`autoreconcile.IsInitial`) are not counted. Every part word is counted whatever its role. Counting is `NameComparer.CountWords`, and `strength` looks words up through the same `Words.resolve` rules. `computeStats` counts with `match.DefaultNames.CountWords`. | One key path for count and lookup, so a rule change cannot desynchronize them. Counting every part keeps the count independent of the name format. Ignored-role words (titles, particles) are counted and never looked up. | Counting per role — would make the pattern a cache key. A free function with its own default rules — the desync M-4 of the review found. Pooling aliases (`Jim` with `James`) — a nickname table is not a frequency table |
| D-5 | A word's discount is `FrequencyDiscount × ln(df) / ln(pop)`. `df` is the word's handle count, floored at 1 and capped at `pop`. `pop = max(People, FrequencyMinPeople, 2)`. An initial and a word with `df ≤ 1` have discount 0. The initial's side pays nothing; the matched full word pays its own. | `df = 1` costs nothing; `df = pop` costs the full discount. The log keeps the ramp gentle: a word on 10 of 100 people loses 10%, half the cap. `FrequencyMinPeople` keeps a small project's discount small: 8 people sharing one surname lose 9%, not 20%. An empty project has `People = 0` and discount 0. | Linear in `df / pop` — a surname on 4% of people would lose under 1%. Plain `ln(N / df)` — the first Smith is rare forever, and an empty project divides by zero |
| D-6 | A pair of names has **discount** = the role-weighted mean of the word discounts of the words the best assignment paired, both sides. Role weights are the pattern's (`nameWord.weight`). With no paired word, discount is 0. The multiplier on points is `1 − discount`. | The family word carries 1.5 against 1.0 for a first given name, so a shared rare surname moves the result more than a shared rare given name. It needs only the `usedA` and `usedB` that `bestAssignment` returns. A zero value means no discount, so a hand-built struct needs no sentinel. | Strength as a multiplier stored directly — its zero value would score nothing, and the code would need a `≤ 0 means 1` rule. Re-running the assignment per word — more code for no change in the result |
| D-7 | Registry values: `Frequency` is `&noWordCounts` (`var noWordCounts WordCounts`, empty). `FrequencyDiscount` is `Set(0.2)` (user answer 2026-10-16). `FrequencyMinPeople` is `Set(100)`. | 0.2 keeps an identical fully common name at 0.8 of its points. A conflict stays `−Contradiction`. 100 is where a project's own counts become a fair sample. | 0.4 and 0.1 — user chose 0.2 |
| D-8 | The discount reaches the scorers through an optional interface, `DiscountComparer.CompareDiscount(a, b Value) (similarity, discount float64, comparable bool)`, and a helper `match.CompareDiscount(cmp, a, b)` that returns discount 0 for a plain `Comparer`. `NameComparer.Compare` returns the similarity from `CompareDiscount`. `compareNames` returns `(similarity, discount float64, ok bool)`. Both are on the value receiver. | `Comparer` stays the contract every other comparer meets, so text, term, date, and integer comparers need no change. An interface keeps `match.best` free of a type switch on `NameComparer`. | A package-level type switch — couples `best` to one comparer. A new method on `Comparer` — forces four no-op methods |
| D-9 | `best` picks the pair with the highest `similarity × (1 − discount)` and reports that pair's similarity and discount. When every pair has similarity 0, the discount is 0. `Score` sets `Contribution = Weight × similarity × (1 − discount)` for a positive similarity. A contradiction (`similarity == 0`) is `−Contradiction`, unscaled, with `Reason.Discount` 0. `exhibits.go` picks the member by the same product. `ScoreSimilarity(similarity, discount float64, comparable bool, …)` multiplies positive points by `1 − discount`. | One rule picks the pair everywhere. A contradiction is not weaker because the surname is common. | A tie-break on equal similarity — a second sort key that S10-28 would extend again. Scaling a contradiction |
| D-10 | The discounted word counts for Promote are read before the comparer is used: in `Propose`, `loadStats` runs before `loadCanon`, and `names.Frequency = &stats.Words` is set before `loadCanon` and before `cfg.Names = &names`. `cloneStats` copies the `Words` struct and shares its map, which no code mutates. In `rank`, `names.Frequency` is set from the candidates before `profile.WithNameComparer`. | `canon.go` (seed ranking) and `align.go` (unreachable Subjects) use `match.Rank` with this comparer, so the counts must be on it first. | Setting counts after `loadCanon` — leaves the seed ranking unweighted |
| D-11 | The `MinScore` cut sees the discounted score. The `DefaultProfile` comment in `registry.go` states it. `canon.go` line 108 and `align.go` line 421 are affected readers and need no code change. | Hiding a candidate that only a common surname supports is the point of the PR. | Cutting on the undiscounted score — a common surname alone would keep clearing the bar |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/match/frequency.go` *(new)* | file | `core/match/aliases.go` | Counts and discount are their own concern; `aliases.go` holds a table registry |
| `match.WordCounts`, `NameComparer.CountWords` *(new)* | type, method | `core/match/aliases.go` `AliasTable` | A count is not a table of equivalence sets |
| `NameComparer.Frequency`, `FrequencyDiscount`, `FrequencyMinPeople` *(new)* | fields | `NameComparer.GivenOnlyFactor`; `WordRules.Aliases` | Same pointer-with-registry-default pattern. Name-only, so not on `WordRules`, which `TextComparer` shares |
| `match.DiscountComparer`, `match.CompareDiscount` *(new)* | interface, func | `match.Comparer` | See D-8 |
| `Reason.Discount`, `PropertyComparison.Discount` *(new)* | fields | `Reason.Similarity`, `PropertyComparison.Similarity` | A reason that omits it hides why 10 points became 8.6 |
| `Stats.Words` *(new)* | field | `Stats.ValueFreq` | Same object, same lifetime. The value is a count per word, not a fraction per value |
| `graphalign.ScoreSimilarity` discount parameter *(new)* | parameter | `ScoreCandidate`'s provenance multiplier | One place multiplies; `ApplyScale` stays pure |
| `BenchmarkLoadStatsNames` *(new)* | benchmark | `BenchmarkProposeColdHub` | `bulkCanon` inserts no name rows, so the existing benchmark cannot show the count's cost |

**Builds on.** `loadStats` and `Stats` for the lazy, graph-held, commit-dropped count; the registry rule that every setting has a registry value; `bestAssignment`'s used flags.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-28 multiplies support into the same pair-selection rule in `best` (D-9) and adds a field beside `Discount` when it needs one. A seeded population table (ideas) becomes another source of `WordCounts` on `Stats`. Counts by role become a key change in `CountWords` only.

### Schema

None — no catalog change. The count reads `auto_reconciler_values`, which exists.

### Go core

| Package | Change |
| --- | --- |
| `core/match` `frequency.go` *(new)* | `type WordCounts struct { People int; Words map[string]int }`. `func (c NameComparer) CountWords(people [][]namevalues.Value) WordCounts`: one entry per handle; D-4 keys from `c.Words.resolve(false)`. `func (c NameComparer) discount(wa []nameWord, usedA []bool, wb []nameWord, usedB []bool, r wordRules) float64` (D-5, D-6) |
| `core/match` `compare.go` | `DiscountComparer` interface and `CompareDiscount(cmp, a, b)` (D-8) |
| `core/match` `names.go` | `NameComparer` gains the three fields. `compareNames` returns the discount beside the score. `CompareDiscount` is the exported path; `Compare` returns its similarity |
| `core/match` `registry.go` | `noWordCounts`; `DefaultNames` gains `Frequency: &noWordCounts`, `FrequencyDiscount: Set(0.2)`, `FrequencyMinPeople: Set(100)`. The `DefaultProfile` comment says the `MinScore` cut sees the discounted score (D-11) |
| `core/match` `match.go` | `best` returns `(similarity, discount, ok)` and picks by D-9. `Score` applies it. `Reason` gains `Discount` |
| `core/match` `pair.go` | `compareProperty` returns the discount. `PropertyComparison` gains `Discount` |
| `core/graphalign` `types.go`, `score.go`, `align.go` | `Stats` gains `Words match.WordCounts`. `ScoreSimilarity(similarity, discount float64, comparable bool, …)` multiplies positive points by `1 − discount`. `propertyPoints` and `comparisonsFrom` pass `pc.Discount` |
| `core/database/promotealign` `stats.go` | `computeStats` fills `Words` with one query (below) and `match.DefaultNames.CountWords`. `cloneStats` copies the `Words` struct and shares its map |
| `core/database/promotealign` `promotealign.go` | `Propose` calls `loadStats` before `loadCanon` and sets `names.Frequency = &stats.Words` (D-10) |
| `core/database/promotealign` `exhibits.go` | `comparerSimilarity` becomes `comparerDiscount` returning the discount through `match.CompareDiscount`. The member is chosen by `s × (1 − d)`. Its `ScoreSimilarity` call passes the discount. A pinned pair with no comparer result keeps discount 0 |
| `core/database/matching` `matching.go` | In `rank`, when the profile has a `name` Feature, `names.Frequency` is set from `names.CountWords` over the loaded candidates (before the exclude filter) |

The Promote count query:

```sql
SELECT r.entity_id, r.value_name
FROM auto_reconciler_values r
JOIN properties p ON p.id = r.property_id AND p.key = 'name' AND p.origin = 'provenencia'
JOIN canonical_entities e ON e.id = r.entity_id AND e.merged_into_id IS NULL
JOIN subject_types st ON st.id = e.subject_type_id AND st.key = 'person' AND st.origin = 'provenencia'
WHERE r.value_name IS NOT NULL
ORDER BY r.entity_id
```

Rows with the same `entity_id` are one handle. Errors: none new. A decode failure returns the wrapped error, as `scanEntityValues` does.

### FFI

None — scores change; no message does.

### macOS

None — no view changes. Promote's row weights shift by at most 20%; no copy, state, or layout changes.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Count name words.**
Files: `core/match/frequency.go` *(new)*, `core/match/frequency_test.go` *(new)*.
Do: `WordCounts` and `NameComparer.CountWords` per D-4. The three fields and registry values (D-7) are added here so `TestRegistrySetsEverySetting` stays green.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match -run 'TestCountNameWords|TestRegistrySetsEverySetting'` passes (T-1, T-4).

**STEP-2 — Discount on the name comparer.**
Files: `core/match/names.go`, `core/match/registry.go`, and the `core/match/frequency.go` *(new)* and `core/match/frequency_test.go` *(new)* from STEP-1.
Do: `discount` (D-5, D-6) and `compareNames` returning it. Update the `DefaultProfile` comment (D-11).
Check: `go test -tags fts5 ./core/match` passes, with every existing name case unchanged (T-2, T-3).

**STEP-3 — Discount through Score and Evaluate.**
Files: `core/match/compare.go`, `core/match/match.go`, `core/match/pair.go`, `core/match/match_test.go`.
Do: D-8, D-9, D-11. `Reason.Discount`, `PropertyComparison.Discount`.
Check: `go test -tags fts5 ./core/match` passes (T-5).

**STEP-4 — Discount in Promote's scale.**
Files: `core/graphalign/score.go`, `core/graphalign/align.go`, `core/database/promotealign/exhibits.go`, `core/graphalign/align_test.go`.
Do: the `ScoreSimilarity` parameter and the three call sites (D-9).
Check: `go test -tags fts5 ./core/graphalign ./core/database/promotealign` passes (T-6).

**STEP-5 — Counts on Stats.**
Files: `core/graphalign/types.go`, `core/database/promotealign/stats.go`, `core/database/promotealign/promotealign_test.go`, `core/database/promotealign/bench_test.go`.
Do: `Stats.Words`, the query, `cloneStats`, and `BenchmarkLoadStatsNames` (inserts 20,000 Person handles with one `name` row each, then times `LoadStatsForTest` after `Graph().Drop()`).
Check: `go test -tags fts5 ./core/database/promotealign -run 'TestStatsCountsNameWords|TestStatsCacheInvalidatesOnWrite'` passes (T-7, T-8). `go test -tags fts5 -bench BenchmarkLoadStatsNames -run '^$' ./core/database/promotealign` runs; report ns/op.

**STEP-6 — Wire the two readers.**
Files: `core/database/promotealign/promotealign.go`, `core/database/matching/matching.go`, `core/database/matching/matching_test.go`, `core/database/promotealign/promotealign_test.go`.
Do: D-10 and the `rank` change (D-3).
Check: `go test -tags fts5 ./core/database/matching ./core/database/promotealign` passes (T-9, T-10, T-11).

**STEP-7 — Docs.**
Files: `docs/matching.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-27-name-frequency.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Rank a shared rare surname above a shared common one."). No WIP commits.

## 8. Tests

Fixture for T-9 to T-11, the **60-person fixture**: 60 filed Person handles named `John <Surname>`; 30 with `Smith`, 1 with `Quirk`, 29 with `Surname01`…`Surname29`. In `matching_test.go` build them with `f.handle("person", named("John Smith"))`. In `promotealign_test.go` build them with the `person` and `promote` helpers. A reconciled handle with two name values (T-7) is inserted with SQL into `auto_reconciler_values`, using the column list of the prepared statement in `bulkCanon` (`bench_test.go`) plus `value_name` from `valuecodec.MarshalName`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/match/frequency_test.go` `TestCountNameWords` *(new)* | Handle A has `Smith` in two values and `José`; handle B has `Jose Smith` and `J.`. → People 2, `smith` 2, `jose` 2, no `j`. A handle with only an initial adds no People. A comparer with custom folds keys by its own folds | AC-1.6, AC-1.7, D-4 |
| T-2 | Go | `core/match/frequency_test.go` `TestNameDiscount` *(new)* | Table-driven. Counts as AC-1.1: `John Quirk`=`John Quirk` → 0.040; `John Smith`=`John Smith` → 0.136. People 8, `robins` 8: `Robins`=`Robins` → 0.090. People 0 → 0. `J. Robins` vs `James Robins` with `robins` 40, `james` 10 → 0.116. A zero-similarity pair → 0 | AC-1.1, AC-1.3, AC-1.4, AC-1.5, D-5, D-6 |
| T-3 | Go | `core/match/names_test.go` `TestNameComparerScores`, `TestNameComparerInvariants` | Unchanged cases pass with `Frequency` unset | AC-2.1 |
| T-4 | Go | `core/match/registry_test.go` `TestRegistrySetsEverySetting`, `TestBareComparersMatchTheRegistry` | Pass with the three new fields set | AC-2.2, D-7 |
| T-5 | Go | `core/match/match_test.go` `TestDiscountedScoring` *(new)* | Table-driven. Counts swapped in through `Profile.WithNameComparer`: `Rank` of the person profile gives 9.60 (Quirk) and 8.64 (Smith); `Reason.Discount` 0.040 and 0.136. `EvaluateWithNames` with the same comparer on identical `John Smith` → `agree`, `PropertyComparison.Discount` 0.136. A zero-similarity name → `−Contradiction`, `Reason.Discount` 0. A candidate at 3.2 undiscounted and a 0.136 discount → 2.76, dropped by `MinScore` 3 | AC-1.1, AC-1.10, D-8, D-9, D-11 |
| T-6 | Go | `core/graphalign/align_test.go` `TestScoreSimilarityDiscount` *(new)* | Table-driven. Name scale, similarity 1, discount 0.136 → `agree`, 2.16. Discount 0 → 2.5. Conflict with discount 0.5 → −3 | AC-1.2, D-9 |
| T-7 | Go | `core/database/promotealign/promotealign_test.go` `TestStatsCountsNameWords` *(new)* | Three Person handles through `LoadStatsForTest`: one with `Smith` in two reconciled values, one merged into another, one named `José Smith`. → People 2, `smith` 2 for the unmerged pair, the merged handle absent | AC-1.7, D-2 |
| T-8 | Go | same file `TestStatsCacheInvalidatesOnWrite` | Extended: after a write that renames a handle, `LoadStatsForTest` returns new `Words`; two reads with no write between return equal `Words` and issue one count | AC-1.8, D-3 |
| T-9 | Go | `core/database/matching/matching_test.go` `TestForSubjectRanksRareSurnameFirst` *(new)* | The 60-person fixture. `ForSubject` for a `John Smith` Subject and a `John Quirk` Subject: the Quirk top score is above the Smith top score. `TestForSubjectQueryCountIsConstant` still holds | AC-1.9, D-3 |
| T-10 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeWeighsACommonNameLess` *(new)* | The 60-person fixture; a Source with an incoming `John Smith` and `John Quirk`. Both name comparisons are `agree` and pinned; the Quirk weight is above the Smith weight | AC-1.2, D-1, D-10 |
| T-11 | Go | `core/database/matching/matching_test.go` `TestRankSkipsCountsWithoutANameFeature` *(new)* | `ForSubject` for a place Subject issues the same number of queries as before and leaves `Frequency` empty | D-3 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break are `core/match` (name invariants) and `core/database/promotealign` (weights asserted as numbers).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host with no xcodebuild, the Go and plan-doc steps pass and the Xcode step is reported as not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] BenchmarkLoadStatsNames ns/op is reported; it stays under 250 ms/op at 20,000 persons
- [ ] No new user-facing copy; no Swift change
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-27 Plan line status and the Slice 7 list line; completed.md: one short paragraph of what shipped
- [ ] docs/matching.md: a worked example for a common and a rare surname; the "Not yet handled" bullet no longer names project name frequency (leave the other items in that bullet as they are)
- [ ] docs/promote-graph-alignment.md §6: the fixed name scale's points are scaled by the name discount; similarity and pinning are unchanged
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice has no new chrome, so the check is a catalog run, not a screen.

1. Run `go test -tags fts5 -run TestProposeWeighsACommonNameLess -v ./core/database/promotealign`. See two `agree` name rows, Quirk above Smith.
2. Open a project with at least 100 filed people and a common surname on many of them. Open a Source with an incoming Person of that surname and one of a rarer surname, and open Promote. Both name rows read as agreed; the rarer surname's suggestion sorts above the common one when everything else is equal.
3. Open a project with fewer than 10 filed people. Promote scores match the base within 10%.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The count runs on the first Promote read after every handle commit | It is dropped only by a handle commit, as the rest of `Stats` is. `BenchmarkLoadStatsNames` measures it at 20,000 persons with name rows |
| A stored-fixture score changes in a test that does not mean to cover frequency | `Frequency` is empty in every existing fixture, so the discount is 0. A failing existing case means the discount leaked into a path D-1 rules out |
| A candidate that only a common surname supported disappears from merge hints and seeds | Intended (D-11, AC-1.10). The `DefaultProfile` comment says so |
| Fixture churn with S10-28 | S10-28 sequences after this PR. It extends the same pair-selection rule (D-9) |

**Stop and report if:**

- `compareNames` no longer returns `usedA` and `usedB` from `bestAssignment` (§4), so D-6 cannot be computed from them.
- Any existing `names_test.go` case changes value with `Frequency` unset.
- `ScoreSimilarity` has a caller outside the three in §4.
- `BenchmarkLoadStatsNames` reports 250 ms/op or more.
- An identical name loses `agree` in a test (D-1 broken).

## 12. Open questions

None.

## 13. PR

**Title:** S10-27 — Name frequency in this project
**Body:** A shared rare surname now earns more than a shared common one; what shipped against US-1 and US-2; AC evidence; deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-27-name-frequency.md`.
