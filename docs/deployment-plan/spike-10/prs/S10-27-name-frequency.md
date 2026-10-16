# S10-27 — Name frequency in this project

**Status:** Draft
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-27--name-frequency)
**Brief:** — (no new chrome: the score changes, no view does)
**Board:** —
**Depends on:** S10-24 (landed [#378](https://github.com/mendahu/provenencia/pull/378))
**Blocks:** — (S10-28 sequences after S10-24 and touches the same comparer and fixtures; it does not wait on this PR)
**Branch:** `s10-27-name-frequency`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-name-value`](../../../../.agents/skills/add-name-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/name-values.mdc`
**Context:** [`name-matching-enhancements.md`](../name-matching-enhancements.md), [`docs/matching.md`](../../../matching.md), [`docs/promote-graph-alignment.md`](../../../promote-graph-alignment.md) §6, working rules "Name frequency is this project's inverse frequency" and "Accent fold is matching only" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-16 against `main` @ `1fc683b`
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

A shared surname that is rare in the researcher's own project counts for more than a shared common one. Today "John Smith" against "John Smith" earns the same points as "John Quirk" against "John Quirk", and that is the main source of false suggestions in Promote. After this PR a word that many filed people carry earns up to 20% fewer points, and a word that one person carries earns all of them.

**Why now.** Slice 7 makes names match the way records are written. S10-24 and S10-25 widened what counts as the same word. This PR narrows how much a common word proves. It comes before S10-28 so support weighting and negative evidence are tuned once against frequency-weighted fixtures.

**What stays.** A name's similarity. An identical name is still similarity 1, still `agree`, and still pins in Promote. Display, `NormalizeForm`, `sort_key`, and the auto-reconciler are untouched. Context-free callers (`DefaultNames`, `NameComparer{}`) score exactly as before: frequency is data a catalog load supplies, the way alias tables are.

## 2. User stories

**US-1.** As a researcher with many filed people, when Promote compares an incoming "John Quirk" to a filed "John Quirk" and another incoming "John Smith" to a filed "John Smith", I see the Quirk suggestion rank above the Smith one, so that a common surname alone does not look like a confident match.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | 100 filed people; "smith" is on 40 handles, "quirk" on 1, "john" on 10. `John Quirk` vs `John Quirk`, then `John Smith` vs `John Smith`, scored by `DefaultProfile("person")` | Both name similarities are 1. Strengths are 0.960 and 0.864. The contributions are 9.60 and 8.64 |
| AC-1.2 | The same 100 people, Promote's fixed name scale (weight 2.5) | `agree`, pinned, with weight 2.40 for Quirk and 2.16 for Smith |
| AC-1.3 | The catalog has no filed person with a name | Strength is 1 and every score equals today's |
| AC-1.4 | One household: 8 filed people, all with "robins" | "Robins" strength is 0.910, not 0 and not below 0.8 |
| AC-1.5 | A name word is an initial (`J.` vs `James`) | The initial takes strength 1. Only `James` is discounted |
| AC-1.6 | A word that is only an accent-fold of another (`José`, `Jose`) | Both count under one folded key |
| AC-1.7 | A handle carries "Smith" in two name values, or the handle is merged into another | It counts once for "smith". A merged handle counts for nothing |
| AC-1.8 | A handle commit changes a name | The next Promote read recounts. A read with no commit between reads the count once |

**US-2.** As a contributor, when I score two names with a bare `NameComparer{}`, I get today's numbers, so the 0–1 invariants in `names_test.go` keep holding.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `NameComparer{}` and `DefaultNames` score every case in `nameCases` | Unchanged from the current values |
| AC-2.2 | `TestRegistrySetsEverySetting` runs | Passes. Each new setting has a registry value |

## 3. Scope

| In | Covered by |
| --- | --- |
| Count name words over the filed people of a project, one count per handle per word | STEP-1, STEP-6, AC-1.6, AC-1.7 |
| A strength factor from a word's count, scaled by role weight, applied to points and not to similarity | STEP-2, STEP-3, STEP-4, AC-1.1, AC-1.2, AC-1.5 |
| A floor for small projects (a minimum population) so the first Smith is not rare forever and an empty project still matches | STEP-2, AC-1.3, AC-1.4 |
| The count is loaded at score time from the catalog, kept on the open catalog's graph, and dropped by a handle commit | STEP-5, STEP-6, AC-1.8 |
| Promote's proposal and merge-hint ranking both read it | STEP-7, AC-1.1, AC-1.2 |
| Docs: `matching.md`, `promote-graph-alignment.md`, the map, `completed.md` | STEP-8 |

| Out | Home |
| --- | --- |
| A seeded population table | [`docs/ideas/name-matching-enhancements.md`](../../../ideas/name-matching-enhancements.md) |
| Counting unfiled Subjects' names or every name Observation | not planned (D-2) |
| Frequency for place, event, text, or date values | not planned. Promote's `u` frequency scales already cover text and term values |
| Support-weighted scores and negative evidence | S10-28 |
| A setting or UI to change the discount | not planned |
| A migration or a stored count table | not planned (D-3) |

## 4. Current state

Verified at `1fc683b`.

| Area | Where | Today |
| --- | --- | --- |
| Name comparer | `core/match/names.go` `NameComparer.compareNames` | Pairs words one-to-one by best assignment of `similarity × affinity × (weightX + weightY)`, divides by the counted weight. Knows nothing about how common a word is. `bestAssignment` returns which words of each side were paired |
| Name comparer settings | `core/match/registry.go` `DefaultNames` | Every pointer field has a value. A new pointer field without one fails `TestRegistrySetsEverySetting` |
| Comparer interface | `core/match/compare.go` `Comparer.Compare(a, b Value) (float64, bool)` | Similarity only |
| Point scoring | `core/match/match.go` `Score`, `best` | `Contribution = Weight × Similarity` of the best pair, or `−Contradiction` at 0. `best` returns similarity and comparability only |
| Pairwise evaluation | `core/match/pair.go` `evaluate`, `compareProperty`, `PropertyComparison` | Similarity and comparable per Property. No strength |
| Promote scale | `core/graphalign/score.go` `ScoreSimilarity`, `ApplyScale`; `core/graphalign/registry.go` name scale `{Floor: 0.75, Weight: 2.5, Contradiction: 3}` | Points are `Weight × similarity`. Outcome `agree` needs similarity 1 |
| Scale call sites | `core/graphalign/score.go` `propertyPoints`, `core/graphalign/align.go` `comparisonsFrom`, `core/database/promotealign/exhibits.go` `comparerSimilarity` and its `ScoreSimilarity` call | Each passes similarity to `ScoreSimilarity`. `exhibits.go` pins when `outcome == match.OutcomeAgree` |
| Catalog comparer build | `core/database/promotealign/promotealign.go` `Propose` (`nameformats.Load`, `namealiases.Comparer(c.Graph(), q, loaded.Comparer())`); `core/database/matching/matching.go` `rank` (`namealiases.Comparer(nil, q, loaded.Comparer())`) | The one place a catalog supplies the name comparer. `canon.go` and `align.go` receive it |
| Alias precedent | `core/database/namealiases/namealiases.go` `Comparer`, `Warm`; `core/database/graphcache/graph.go` `AliasTables`, `SetAliasTables` | Catalog rows become a comparer field; the graph holds them for the open project |
| Stats precedent | `core/database/promotealign/stats.go` `loadStats`; `core/database/graphcache/graph.go` `Stats`, `SetStats`, `clear`; `core/database/graphcache/commit.go` `OnCommit` | Counted once, kept on the graph, cleared by a handle commit and a structure drop. Computed lazily on the first read |
| Handle names | `auto_reconciler_values` (`entity_id`, `property_id`, `value_name` blob); `core/valuecodec/valuecodec.go` `UnmarshalName` | Every rank of every handle's reconciled name values. `matching.go` `scanEntityValues` reads the same blob |
| Tests | `core/match/names_test.go`, `core/match/registry_test.go` `TestRegistrySetsEverySetting`, `TestBareComparersMatchTheRegistry`; `core/database/matching/matching_test.go` `TestForSubjectQueryCountIsConstant`; `core/database/promotealign/promotealign_test.go`; `core/database/promotealign/bench_test.go` `BenchmarkProposeColdHub` (20,000 persons) | Green at the base. Name invariants assert a name matches itself at 1 |

**Precedent.** S10-25 ([#379](https://github.com/mendahu/provenencia/pull/379)) added catalog-supplied alias tables to the name comparer through `namealiases.Comparer`. Copy: a loader package whose `Comparer(g, q, base)` returns `base` with one field set, and a graph slot for the open project. Do not copy: its `Warm` at project open (D-3 loads lazily, as `loadStats` does) or its tables in `catalogmodel` (no table here).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Frequency scales **points**. The comparer's similarity is unchanged: an identical name stays 1, `agree`, and pinned. | `exhibits.go` and `align.go` pin only on `agree`, and `agree` needs similarity 1. A lower similarity would unpin every common identical name. User answer 2026-10-16. | Lowering the similarity — unpins identical common names. Boosting rare names above baseline — leaves common names where they are, and the aim is fewer false suggestions on common names |
| D-2 | The population is the unmerged handles' reconciled name values in `auto_reconciler_values` for the product `name` Property, every rank. A handle counts once per word. User answer 2026-10-16. | Matches how Promote's `stats` already measures frequency (handles). Transcription volume does not move it. | Every name Observation — ten records of one person would make a surname look common |
| D-3 | Counts are read at score time and kept on the open catalog's graph. No migration, no stored table. The graph slot is cleared with `stats` (a handle commit and `clear`). It loads lazily on the first read, not at project open. | A count is derived from `auto_reconciler_values` and is wrong the moment a handle changes. `loadStats` is the shape. The scope allows either; this one adds no schema. | A count table rebuilt with the search index — a second derived store to keep in step. Warming at open — makes open slower for a value Promote alone reads |
| D-4 | A word's key is its comparison token: `splitWords` output (particles rejoined) passed through `foldWord` with the registry folds. Initials (`autoreconcile.IsInitial`) are not counted. Every part word is counted whatever its role. | One key for `José` and `Jose` (the S10-24 fold). Counting every part keeps the count independent of the name format, so it needs no re-count when the default format changes. Ignored-role words (titles, particles) are counted and never looked up. | Counting per role — `Wood` as given and surname are different frequencies, but the pattern would then be a cache key. Counting pooled aliases (`Jim` with `James`) — a nickname table is not a frequency table |
| D-5 | A word's credit is `1 − Discount × ln(df) / ln(pop)`. `df` is the word's handle count, floored at 1 and capped at `pop`. `pop = max(People, MinPeople)`. An initial and a word with `df ≤ 1` credit 1. | `df = 1` costs nothing; `df = pop` costs the full discount. The log keeps the ramp gentle (a word on 10 of 100 people loses 10% of the 20%). `MinPeople` makes a small project's discount small: 8 people sharing one surname lose 9%, not 20%. An empty project has `People = 0` and strength 1. | Linear in `df / pop` — a surname on 4% of people would lose under 1%. Plain `ln(N / df)` — the first Smith is rare forever, and an empty project divides by zero |
| D-6 | A pair of names has **strength** = the role-weighted mean of the credits of the words the best assignment paired, both sides. Role weights are the pattern's (`nameWord.weight`). With no paired word, strength is 1. | The family word carries 1.5 against 1.0 for a first given name, so a shared rare surname moves the result more than a shared rare given name. It needs only the `usedA` and `usedB` that `bestAssignment` already returns. | Strength from the best pair's single word — ignores the other paired words. Re-running the assignment per word — more code for no change in the result |
| D-7 | Registry values: `Frequency` is `&noWordCounts` (empty). `FrequencyDiscount` is `Set(0.2)` (user answer 2026-10-16). `FrequencyMinPeople` is `Set(100)`. | 0.2 keeps an identical fully common name at 0.8 of its points. A conflict stays `−Contradiction`. 100 is where a project's own counts become a fair sample. | 0.4 and 0.1 — user chose 0.2. A min below 2 — `ln(pop)` must be positive, so the code clamps `pop` to at least 2 |
| D-8 | Strength reaches the scorers through an optional interface, `StrengthComparer.CompareStrength(a, b Value) (similarity, strength float64, comparable bool)`, and a helper `match.CompareStrength(cmp, a, b)` that returns strength 1 for a plain `Comparer`. `NameComparer.Compare` returns the similarity from `CompareStrength`. | `Comparer` stays the contract every other comparer meets. Text, term, date, and integer comparers need no change. | A second field on `Value` — pairs, not values, have strength. A new method on `Comparer` — forces four no-op methods |
| D-9 | Where `best` finds two pairs with equal similarity, it keeps the higher strength. Strength applies only when the contribution is positive. A contradiction (`similarity == 0`) is `−Contradiction`, unscaled. `ScoreSimilarity` treats strength `≤ 0` as 1. | Matches `ScoreCandidate`'s rule for provenance `≤ 0`. A hand-built `PropertyComparison` with a zero strength field then still scores. | Scaling a contradiction — a clear disagreement is not weaker because the surname is common |
| D-10 | New code lives in `core/match/frequency.go` *(new)* (pure counts, credit), `core/database/namefrequency` (the catalog load and the graph slot owner), and a graph slot in `core/database/graphcache`. `matching.rank` and `promotealign.Propose` each call `namefrequency.Comparer` after `namealiases.Comparer`. | Same split as `namealiases`: `core/match` stays free of catalog access. | Folding the load into `namealiases` — a different data set with a different lifetime |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/match/frequency.go` *(new)* | file | `core/match/aliases.go` | Counts and credit are their own concern; `aliases.go` holds a table registry |
| `match.WordCounts`, `match.CountNameWords` *(new)* | type, func | `core/match/aliases.go` `AliasTable` | A count is not a table of equivalence sets |
| `NameComparer.Frequency`, `FrequencyDiscount`, `FrequencyMinPeople` *(new)* | fields | `NameComparer.GivenOnlyFactor`; `WordRules.Aliases` | Same pointer-with-registry-default pattern. Name-only, so not on `WordRules`, which `TextComparer` shares |
| `match.StrengthComparer`, `match.CompareStrength` *(new)* | interface, func | `match.Comparer` | See D-8 |
| `Reason.Strength`, `PropertyComparison.Strength` *(new)* | fields | `Reason.Similarity`, `PropertyComparison.Similarity` | A reason that omits it hides why 10 points became 8.6 |
| `core/database/namefrequency` *(new)* | package | `core/database/namealiases` | See D-10 |
| `Graph.WordCounts`, `Graph.SetWordCounts` *(new)* | methods | `Graph.Stats`, `Graph.SetStats` | Different lifetime from `stats`, same invalidation |
| `graphalign.ScoreSimilarity` strength parameter *(new)* | parameter | `ScoreCandidate`'s provenance rule | One place multiplies; `ApplyScale` stays pure |

**Builds on.** `namealiases.Comparer` for the catalog-to-comparer seam, `loadStats` for the lazy graph slot, the registry rule that every setting has a registry value.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-28 reads the same `StrengthComparer` seam: a support share is another factor on the same contribution. A seeded population table (ideas) becomes a second source of `WordCounts`. Name frequency by role becomes a key change in `CountNameWords` only.

### Schema

None — no catalog change. The count reads `auto_reconciler_values`, which exists.

### Go core

| Package | Change |
| --- | --- |
| `core/match` | `frequency.go` *(new)*: `type WordCounts struct { People int; Words map[string]int }`; `func CountNameWords(people [][]namevalues.Value) WordCounts` (one entry per handle; D-4 keys); `func (c NameComparer) strength(wa []nameWord, usedA []bool, wb []nameWord, usedB []bool, folds map[rune]string) float64` (D-5, D-6) |
| `core/match` `compare.go` | `StrengthComparer` interface and `CompareStrength(cmp, a, b)` helper (D-8) |
| `core/match` `names.go` | `NameComparer` gains the three fields. `compareNames` returns strength beside the score; `CompareStrength` is the exported path; `Compare` returns its similarity |
| `core/match` `registry.go` | `noWordCounts`; `DefaultNames` gains `Frequency: &noWordCounts`, `FrequencyDiscount: Set(0.2)`, `FrequencyMinPeople: Set(100)` |
| `core/match` `match.go` | `best` returns `(similarity, strength, ok)`. `Score` sets `Contribution = Weight × Similarity × Strength` for a positive similarity. `Reason` gains `Strength` |
| `core/match` `pair.go` | `compareProperty` returns strength. `PropertyComparison` gains `Strength` (1 when not comparable) |
| `core/graphalign` | `ScoreSimilarity(similarity, strength float64, comparable bool, …)` multiplies positive points by `strength` when `strength > 0`. `propertyPoints` and `comparisonsFrom` pass `pc.Strength` |
| `core/database/promotealign` `exhibits.go` | `comparerSimilarity` becomes `comparerStrength` returning strength through `match.CompareStrength`; its `ScoreSimilarity` call passes it. A pinned pair with no comparer result keeps strength 1 |
| `core/database/graphcache` | `Graph.WordCounts() (any, bool)` and `Graph.SetWordCounts(any)`. Cleared beside `stats` in `OnCommit` and `clear` |
| `core/database/namefrequency` *(new)* | `func Load(q Querier) (match.WordCounts, error)`: one query (below), group by entity, `valuecodec.UnmarshalName`, `match.CountNameWords`. `func Comparer(g *graphcache.Graph, q Querier, base match.NameComparer) (match.NameComparer, error)`: a graph hit uses the stored counts; a miss calls `Load`, stores the result when `g` is not nil, and sets `base.Frequency` |
| `core/database/promotealign` `promotealign.go` | After `namealiases.Comparer`, `names, err = namefrequency.Comparer(c.Graph(), q, names)` |
| `core/database/matching` `matching.go` | Same call with a nil graph in `rank` |

The load query:

```sql
SELECT r.entity_id, r.value_name
FROM auto_reconciler_values r
JOIN properties p ON p.id = r.property_id AND p.key = 'name' AND p.origin = 'provenencia'
JOIN canonical_entities e ON e.id = r.entity_id AND e.merged_into_id IS NULL
WHERE r.value_name IS NOT NULL
ORDER BY r.entity_id
```

Errors: none new. A decode failure returns the wrapped error, as `scanEntityValues` does.

### FFI

None — scores change; no message does.

### macOS

None — no view changes. Promote's row weights shift by at most 20%; no copy, state, or layout changes.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Count name words.**
Files: `core/match/frequency.go` *(new)*, `core/match/frequency_test.go` *(new)*. Skill: `add-name-value`.
Do: `WordCounts`, `CountNameWords` per D-4. Uses `DefaultWordRules.resolve(false)` for particles and folds.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match -run 'TestCountNameWords'` passes (T-1).

**STEP-2 — Strength on the name comparer.**
Files: `core/match/names.go`, `core/match/registry.go`, `core/match/frequency.go` *(new)*, `core/match/frequency_test.go` *(new)*.
Do: the three fields, the registry values, `strength` (D-5, D-6, D-7), `compareNames` returning strength.
Check: `go test -tags fts5 ./core/match` passes, including `TestRegistrySetsEverySetting`, `TestBareComparersMatchTheRegistry`, and every existing name case unchanged (T-2, T-3, T-4).

**STEP-3 — Strength through Score and Evaluate.**
Files: `core/match/compare.go`, `core/match/match.go`, `core/match/pair.go`, `core/match/match_test.go`.
Do: D-8 and D-9. `Reason.Strength`, `PropertyComparison.Strength`.
Check: `go test -tags fts5 ./core/match` passes (T-5, T-6).

**STEP-4 — Strength in Promote's scale.**
Files: `core/graphalign/score.go`, `core/graphalign/align.go`, `core/database/promotealign/exhibits.go`, `core/graphalign/align_test.go`.
Do: the `ScoreSimilarity` parameter and the three call sites.
Check: `go test -tags fts5 ./core/graphalign ./core/database/promotealign` passes (T-7).

**STEP-5 — Graph slot.**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/commit.go`.
Do: `WordCounts`, `SetWordCounts`; reset in `clear` and where `OnCommit` resets `stats`.
Check: `go test -tags fts5 ./core/database/graphcache` passes (T-8).

**STEP-6 — Load the counts.**
Files: `core/database/namefrequency/namefrequency.go` *(new)*, `core/database/namefrequency/namefrequency_test.go` *(new)*.
Do: `Load`, `Comparer` (D-2, D-3, D-10).
Check: `go test -tags fts5 ./core/database/namefrequency` passes (T-9, T-10).

**STEP-7 — Wire the two callers.**
Files: `core/database/promotealign/promotealign.go`, `core/database/matching/matching.go`, `core/database/matching/matching_test.go`, `core/database/promotealign/promotealign_test.go`.
Do: one call each.
Check: `go test -tags fts5 ./core/database/matching ./core/database/promotealign` passes (T-11, T-12). `go test -tags fts5 -bench BenchmarkProposeColdHub -run '^$' ./core/database/promotealign` runs; record ns/op against the base.

**STEP-8 — Docs.**
Files: `docs/matching.md`, `docs/promote-graph-alignment.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-27-name-frequency.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Rank a shared rare surname above a shared common one."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/match/frequency_test.go` `TestCountNameWords` *(new)* | Two handles; handle A has `Smith` in two values and `José`; handle B has `Jose Smith`, `J.`. → People 2, `smith` 2, `jose` 2, no `j`. A handle with only an initial adds no People | AC-1.6, AC-1.7, D-4 |
| T-2 | Go | `core/match/frequency_test.go` `TestNameStrength` *(new)* | Counts People 100, `smith` 40, `quirk` 1, `john` 10. `John Quirk`=`John Quirk` → 0.960. `John Smith`=`John Smith` → 0.864. People 8, `robins` 8, `Robins`=`Robins` → 0.910. People 0 → 1. `J. Robins` vs `James Robins` with `robins` 40 and `james` 10 → 0.884 (the initial credits 1). Table-driven | AC-1.1, AC-1.3, AC-1.4, AC-1.5, D-5, D-6 |
| T-3 | Go | `core/match/names_test.go` `TestNameComparerScores` | Unchanged cases pass with `Frequency` unset | AC-2.1 |
| T-4 | Go | `core/match/registry_test.go` `TestRegistrySetsEverySetting` | Passes with the three new fields set | AC-2.2, D-7 |
| T-5 | Go | `core/match/match_test.go` `TestRankRareSurnameFirst` *(new)* | `Rank` of person profile, probe `John Quirk`, candidate `John Quirk`; probe `John Smith`, candidate `John Smith`; counts as T-2 → scores 9.60 and 8.64; reasons carry Strength 0.960 and 0.864 | AC-1.1, D-8 |
| T-6 | Go | `core/match/match_test.go` `TestStrengthLeavesContradictionsAlone` *(new)* | A zero-similarity name with counts set → `−Contradiction`. `Evaluate` on identical `John Smith` → `agree` and Strength 0.864 | AC-1.2, D-1, D-9 |
| T-7 | Go | `core/graphalign/align_test.go` `TestScoreSimilarityStrength` *(new)* | Name scale, similarity 1, strength 0.864 → `agree`, 2.16. Strength 0 → 2.5. Conflict with strength 0.5 → −3 | AC-1.2, D-9 |
| T-8 | Go | `core/database/graphcache/graph_test.go` `TestWordCountsClearOnHandleCommit` *(new)* | Set counts, `OnCommit` with a handle effect → `WordCounts()` is false. A vocabulary-only commit leaves it | AC-1.8, D-3 |
| T-9 | Go | `core/database/namefrequency/namefrequency_test.go` `TestLoadCountsFiledPeople` *(new)* | Catalog with three person handles (one merged into another, one with `Smith` twice) → People and per-word counts as T-1; the merged handle counts for nothing | AC-1.7, D-2 |
| T-10 | Go | same file `TestComparerKeepsCountsOnTheGraph` *(new)* | First call loads; second call with a `Querier` that fails the test on use returns the same counts; after a handle commit the next call queries again | AC-1.8, D-3 |
| T-11 | Go | `core/database/matching/matching_test.go` `TestForSubjectRanksRareSurnameFirst` *(new)* | 60 filed persons: 30 `Smith`, 1 `Quirk`, 29 others with distinct surnames. `ForSubject` for a `John Smith` Subject and for a `John Quirk` Subject; the Quirk top score is above the Smith top score. `TestForSubjectQueryCountIsConstant` still holds | AC-1.1 |
| T-12 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeWeighsACommonNameLess` *(new)* | The same 60 persons; a Source with an incoming `John Smith` and `John Quirk`. Both name comparisons are `agree` and pinned; the Quirk weight is above the Smith weight | AC-1.2 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break are `core/match` (name invariants) and `core/database/promotealign` (weights asserted as numbers).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] BenchmarkProposeColdHub ns/op is reported against the base; the rise is under 15%
- [ ] No new user-facing copy; no Swift change
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-27 map line and Plan link; completed.md: one short paragraph of what shipped
- [ ] docs/matching.md: a worked example for a common and a rare surname, and "Not yet handled" no longer lists project name frequency
- [ ] docs/promote-graph-alignment.md §6: the fixed name scale's points are scaled by name strength; similarity and pinning are unchanged
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
| The count runs on every Promote read after a handle commit | The graph slot is dropped only by a handle commit, as `stats` is. STEP-7 reports the benchmark (20,000 persons). The rise must stay under 15% |
| A stored-fixture score changes in a test that does not mean to cover frequency | `Frequency` is empty in every existing fixture, so strength is 1. A failing existing case means strength leaked into a path D-1 rules out |
| Fixture churn with S10-28 | S10-28 sequences after this PR. It reads the same seam (D-8) |
| The first Promote read after open is slower | One query and one decode per name value. The benchmark covers the cold path (`Drop` before each run) |

**Stop and report if:**

- `compareNames` no longer returns `usedA` and `usedB` from `bestAssignment` (§4), so D-6 cannot be computed from them.
- Any existing `names_test.go` case changes value with `Frequency` unset.
- `ScoreSimilarity` has a caller outside the three in §4.
- `BenchmarkProposeColdHub` rises by 15% or more and the cause is the count.
- An identical name loses `agree` in a test (D-1 broken).

## 12. Open questions

None.

## 13. PR

**Title:** S10-27 — Name frequency in this project
**Body:** A shared rare surname now earns more than a shared common one; what shipped against US-1 and US-2; AC evidence; deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-27-name-frequency.md`.
