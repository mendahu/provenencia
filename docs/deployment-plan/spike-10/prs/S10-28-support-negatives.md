# S10-28 — Support, negatives, and which values matched

**Status:** Landed in #411 (2026-10-16)
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-28--support-negatives-and-which-values-matched)
**Brief:** — (no new chrome: the score and its reasons change, no view does)
**Board:** —
**Depends on:** S10-24 (landed [#378](https://github.com/mendahu/provenencia/pull/378)); S10-27 (landed [#409](https://github.com/mendahu/provenencia/pull/409)), whose pair-selection seam this PR extends
**Blocks:** —
**Branch:** `s10-28-support-negatives`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`name-matching-enhancements.md`](../name-matching-enhancements.md), [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §4–§6, [`docs/matching.md`](../../../matching.md), [S10-27 plan](S10-27-name-frequency.md) D-8 and D-9 and its Landed note, working rule "Support weighting and negatives land together" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-16 against `main` @ `6577a43` (S10-27 landed). First written against `a4e9de4` while S10-27 was in implementation; §4 was redone on landing
**Review:** [`S10-28-support-negatives.review.md`](S10-28-support-negatives.review.md) — 2 rounds: Fits with changes / Not ready, then Fits with changes / Not ready; every finding answered; no open blocker

**Landed:** The code matches the plan except for these.
- `Values.HasPositive` lives in `core/match/pairs.go` with the other new `Value` methods, not in `match.go`.
- The plan's T-1 row for AC-2.5 is a separate test, `TestPairValuesAgainstReducesSupport`; the minority-weight and order cases are `TestPairValuesOrderIndependent` and the T-1 table.
- Test helpers (`cachedPerson`, `cached`, `kept`, `negative`, `negated`, `addCachedTerm`) are test-only.
- The plain path calls `Label` once for the reported pair, so a name without `Form` or a date allocates there.
- F-1, a user-approved amendment to D-4 (ship log, CON-1): `Feature.Cardinality` and `pairValues(cmp, card, …)`; a multiple Property's voting values all lead at weight 1 and R3 is skipped. The toponym Feature is `multiple`, and `TestDefaultProfileCardinality` in `core/database/subjectvocab` keeps Features and the vocabulary together.
- F-2, a user-approved amendment to D-10 (ship log, UX-1): Promote's sheet shows a row-level conflict as an unpinnable line (`withRowConflicts`), with a proto field `incoming_negative` and the string `promote.sheet.negatedValue`.
- `BenchmarkPairValues/plain` uses `TermComparer`; a text comparer allocates on its own. `BenchmarkProposeColdHub` did not move (about 535-575 ms/op before and after).

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

A handle's values count in proportion to how well supported they are, and a negative Observation counts against a match. A handle that is mostly one sex, or mostly one name, is judged against that majority, so one stray reading cannot make a mixed handle look like a clean match. Each reason says which two values produced it.

**Why now.** Slice 7 makes names compare the way records are written. S10-24, S10-25, and S10-27 changed what a name word is worth. This PR changes which of a handle's values are allowed to speak. It follows S10-27 so fixtures settle against frequency-weighted scores once.

**What stays.** A handle with one value, or with values of equal support, scores exactly as before. The auto-reconciler, the cache, and every Promote view are untouched. No catalog change: the cache already stores `support`, `against`, and a `reason` on every value, and the matcher never read them. Promote's exhibit chooser keeps its own member picker (D-10).

## 2. User stories

**US-1.** As a researcher, when a Subject says "female" and the handle it might belong to is "male" on three Sources and "female" on one, I see a contradiction, so that one dissenting reading does not hide a sex mismatch.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Probe `sex_at_birth` female. Handle values male (support 3), female (support 1). Person profile | Contribution −2. The reason names `female` and `male` |
| AC-1.2 | Probe male, the same handle | Contribution +1. The reason names `male` and `male` |
| AC-1.3 | Probe female. Handle values male (support 2), female (support 2) | Contribution +1: both values lead, so the match stands |
| AC-1.4 | Probe `James Robins`. Handle names `Mary Smith` (support 5), `James Robins` (support 1) | Name similarity 0. `Rank` contribution 0, because the name feature has no contradiction points; Promote's name scale scores it a conflict (−3). The reason names `James Robins` and `Mary Smith`. Not 10 |
| AC-1.5 | Probe `Mary Jones`. Handle names `Mary Smith` (support 1), `Mary Jones` (support 1) | Name similarity 1, contribution 10: equal support keeps both names in full |
| AC-1.6 | Probe `Robbins`. Handle names `Robins` (support 3), `Robbins` (support 1) | Name similarity 0.857, the pair `Robbins` and `Robins`, not 1 |
| AC-1.7 | Values loaded without cache evidence (`Counted` false) | Every weight is 1 and scores equal today's |
| AC-1.8 | Probe male. Handle values male (support 1, reason `kept`), female (support 3, reason `weak`) | Contribution +1. Probe female: −2. The displayed value leads, as on the Person page |
| AC-1.10 | A handle whose only value is male with reason `kept`, support 1, against 3 (or support 0, as a concluded value would carry). Probe male, then probe female | Probe male +1; probe female −2: a displayed value always votes |
| AC-1.9 | Handle values `unknown` (neutral term, support 5), male (support 1). Probe female | Contribution −2: the neutral value never leads |

**US-2.** As a researcher, when a Subject carries a negative Observation ("not male", "not born in York") or the handle carries evidence against one of its values, I see that count as a contradiction.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Probe has a negative `sex_at_birth` male Observation. Handle male (support 3), female (support 1) | Contribution −2. The reason names the negative `male` and `male` |
| AC-2.2 | The same negative probe against a handle that is female (support 3), male (support 1) | No contradiction, no credit. With no positive value on the probe, the feature is not comparable and is skipped |
| AC-2.3 | A probe has a positive female and a negative male. Handle male (support 3), female (support 1) | Contribution −2 (the negative matches the leader; the positive does not rescue it) |
| AC-2.4 | Candidate value male has reason `denied`. Probe male | Contribution −2: the probe asserts a value the handle's evidence refutes. The reason names the two |
| AC-2.5 | Candidate values male (support 3, against 2), female (support 2, against 0). Probe female, then probe male | Effective supports are 1 and 2, so female is the consensus. Probe female: +1. Probe male: −2 |
| AC-2.8 | Promote: a Subject with a negative `sex_at_birth` male Observation, against a handle that is male on three Sources | The sex comparison is a conflict (terms match by key in both loaders) |
| AC-2.6 | `SubjectsValues` for a Subject with one positive and one negative Observation | Both come back; the negative has `Negative` set and its Observation id |
| AC-2.7 | A Subject that has only a negative Observation on a profile Property, ranked with `ForSubject` | The feature is skipped, and a probe whose profile Properties hold only negatives ranks nothing |

**US-3.** As a contributor, when I read a `Reason`, I can tell which two values produced it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Any credited or contradicted reason | `Reason.Probe` and `Reason.Candidate` carry the label, the negative flag, and for a probe value its Observation id |
| AC-3.2 | `EvaluateWithNames` for the stray-name case of AC-1.4 | The comparison is a conflict, not an agree |
| AC-3.3 | The same Subject and handle scored twice, rows read in a different order | The same pair is reported |

## 3. Scope

| In | Covered by |
| --- | --- |
| Weight each value by its effective support relative to the leading value on its side | STEP-1, STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-2.5 |
| Judge contradictions against the leading values of each side | STEP-1, AC-1.1, AC-1.3, AC-1.4, AC-1.8, AC-1.9 |
| Load negative Observations of a Subject; score a negative that matches a leading value as a contradiction | STEP-3, STEP-4, AC-2.1, AC-2.2, AC-2.3, AC-2.6, AC-2.7, AC-2.8 |
| Load `support`, `against`, and `reason` from the handle cache, for the matching loader and Promote's graph loader | STEP-3, STEP-4, AC-2.4, AC-2.5 |
| `Reason` and `PropertyComparison` name the probe value and the candidate value | STEP-2, AC-3.1, AC-3.2, AC-3.3 |
| Docs: `matching.md`, the spike design note, the map, `completed.md` | STEP-5 |

| Out | Home |
| --- | --- |
| A concluded value dominating the others (Reconciliation Claims) | [`docs/ideas/archive/name-matching-enhancements.md`](../../../ideas/archive/name-matching-enhancements.md) |
| Remembering a rejected claim | same idea note |
| A Promote layout for the reason text, or a sentence built from the pair | S10-36 covers crops only; not planned |
| Support weights in Promote's exhibit chooser (`exhibits.go` keeps picking members by similarity) | not planned (D-10) |
| A support weight on the incoming Subject's own values (each Observation counts 1) | not planned |
| Claim confidence or Source credibility as a weight | not planned: the spike decided "share of support, not rank and not claim confidence" |
| Changing the auto-reconciler, its cache, or `against` counting | not planned |
| Unifying `Value.Label` with the exhibit display chain | not planned: exhibits resolve term labels from the catalog (D-7) |
| A migration | not planned |

## 4. Current state

Verified at `6577a43`, with S10-27 landed ([#409](https://github.com/mendahu/provenencia/pull/409)). Line numbers are at that commit.

| Area | Where | Today |
| --- | --- | --- |
| Pair selection | `core/match/match.go` `best` (line 136; two callers: `Score` line 118, `compareProperty` in `pair.go` line 100) | Picks the comparable pair with the highest `similarity × (1 − discount)` and returns `(similarity, discount, ok)`, with discount 0 when similarity is 0. It reads each pair through `CompareDiscount`. Two values of any support count alike |
| Scoring | `core/match/match.go` `Score`, `Reason`, `Rank` | `Contribution = Weight × Similarity × (1 − Discount)`, or `−Contradiction` at similarity 0. `Reason` has `Property`, `Similarity`, `Discount`, `Contribution` |
| Pairwise evaluation | `core/match/pair.go` `evaluate`, `compareProperty`, `PropertyComparison` | `PropertyComparison` has `Similarity`, `Discount`, `Comparable`. A comparable zero is a conflict (single) or unknown (multiple) |
| Value shape | `core/match/match.go` `Value` | No support, no reason, no polarity, no label |
| Discount seam | `core/match/compare.go` `DiscountComparer`, `CompareDiscount`; `core/match/names.go` `NameComparer.CompareDiscount` | `CompareDiscount` returns a plain `Comparer`'s result with discount 0 |
| Same-value test | `core/autoreconcile/modules.go` `Compatible(valueType, a, b)`; `core/database/promotealign/exhibits.go` `exhibitCompatible`, `exhibitAuto` | The test the reconciler uses to deny and to count `against`; Promote's agreement test. `exhibitAuto` adapts a `match.Value` to it. `core/match/words.go` already imports `autoreconcile` |
| Handle cache | `auto_reconciler_values` (`support`, `against`, `reason`, `rank`, all but `rank` NOT NULL or defaulted); `core/autoreconcile/autoreconcile.go` `Reason*` constants and `ReconciledValue.Displayed` | One row per distinct value. `support` is distinct Sources, `against` is negative Observations that counted against the value, `reason` is `kept`, `outvoted`, `weak`, `denied`, or `provisional`. `Displayed` is `reason == kept` |
| Matching loader | `core/database/matching/matching.go` `sqlEntityValues`, `scanEntityValues`, `rank` (line 159; the early return is line 160), `candidateNames` | Selects none of `support`, `against`, `reason`. Rows come `ORDER BY e.ref, r.entity_id` with no order within a handle. `rank` returns early on `len(probe) == 0`; since S10-27 it also counts name words over the candidates it loaded |
| Subject loader | `core/database/matching/matching.go` `sqlSubjectsValues` (line 72, `WHERE o.polarity = 'positive'`), `SubjectsValues`; callers `loadSubjectValues`, `core/database/promotealign/layer.go` line 48 | Negatives never load. No `ORDER BY` |
| Promote's handle values | `core/database/graphcache/graph.go` `Value` (`Support`, `Against`, `Reason`, `Rank`); `core/database/graphcache/load.go` selects all three; `core/database/promotealign/canon.go` `handleValues` (line 245) | The graph row has them; `handleValues` copies none |
| Readers of `Values` that would misread a negative | `core/graphalign/score.go` `uFor` (line 145); `core/graphalign/align.go` `subjectHasValues` (line 972); `core/database/promotealign/canon.go` line 80 (`len(s.Values) == 0`); `core/database/matching/matching.go` `rank` line 160 (`len(probe) == 0`) | Each treats any value as a value. `exhibits.go` and `terms.go` already filter `polarity = 'positive'` in SQL |
| Subject-versus-Subject checks | `core/graphalign/align.go` lines 774, 854, 865 call `evaluate` on two Subjects | Both sides hold Subject values, so both can now carry negatives |
| Exhibit chooser | `core/database/promotealign/exhibits.go` (loop at lines 195–225, `comparerDiscount` at line 257, `exhibitAuto` at line 268) | Compares an incoming Observation with member Observations by `s × (1 − d)` since S10-27, and pins on `agree` |
| Tests | `core/match/match_test.go`, `core/match/pair_test.go`, `core/database/matching/matching_test.go` (`newFixture`, `handle`, `termed`, `named`), `core/database/promotealign/promotealign_test.go`, `bench_test.go` `bulkCanon` | Green at the base. `matching_test.go` is package `matching_test` and cannot call `bulkCanon`. S10-27 re-based absolute scores in `TestForSubjectPersons`, `TestForSubjectComparesNameParts`, and `TestForEntityMergeHints` (11.0 → 10.5, 10.0 → 9.7): they now assert discounted numbers |

**Precedent.** S10-27 (landed) put a per-pair factor behind one seam: `best` selects and reports the pair, and `Score` and `compareProperty` call it. Copy that. Exhibits are not a third caller: they pick a member Observation with their own loop, and neither plan changes that (D-10).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A value's weight is its effective support divided by the leaders' effective support (D-3). A leader weighs 1. User answer 2026-10-16. | A stray one-of-five weighs 0.2. Two equally supported names (a maiden and a married name) both lead, so a probe that matches either is a full match. | Share of the total — two equal names would each weigh 0.5 |
| D-2 | A candidate value whose reason is `denied` has weight 0, never leads, and a probe value that is Compatible with it is a contradiction (R2). `against` reduces effective support by that count. A displayed (`kept`) value is floored at effective support 1, so a handle's only displayed value always votes and is never a contradiction: the reconciler shows it, and a concluded value (`kept`, support 0) is compared as it is today. Any other value is floored at 0, and a non-displayed value at 0 has weight 0, never leads, and does not contradict. User answer 2026-10-16, narrowed to the stated wording. | The reconciler already decided that a stronger negative refutes a `denied` value. `against` counts negative Observations and `support` counts distinct Sources; the PR offsets one for one by decision and does not convert. | Treating `against ≥ support` as disputed — contradicts a value the Person page shows |
| D-3 | **Definitions.** `Counted` is true on values a cache loader fills; every other value has weight 1. Effective support is `max(Support − Against, 1)` for a `Counted` value whose reason is `kept`, `max(Support − Against, 0)` for any other `Counted` value, and 1 for a value that is not `Counted`. A **voting** value is positive, not `denied`, not `provisional`, not `Neutral`, and has effective support above 0. The **leaders** of a side are the voting values with reason `kept` (or not `Counted`) that share the highest effective support; when none is `kept`, the voting values that share it. A voting value's weight is `min(1, effective ÷ leader effective)`. Negative values are never voting. A handle probe (`ForEntity`) is `Counted` and weighted the same way as a candidate. | Leaders follow what the reconciler displays, so Promote agrees with the Person page. A neutral or provisional value never decides a match. | Leaders by raw support — three weak readings could outrank one strong displayed value |
| D-4 | One function, `pairValues(cmp Comparer, card string, probe, candidate []Value) PairResult`, replaces `best`. Rules in order. **R1** a negative value on one side that is Compatible (D-5) with a leader on the other is a contradiction, naming the pair. **R2** a positive value on one side that is Compatible with a `denied` value on the other, when no voting value on that other side is Compatible with it, is a contradiction. **R3** when both sides have leaders, at least one leader pair is comparable, and every comparable leader pair has similarity 0, that is a contradiction. Otherwise the credit is the highest `similarity × weightA × weightB × (1 − discount)` over comparable voting positive pairs, reported with similarity `similarity × weightA × weightB`. With no comparable voting pair the result is not comparable. A contradiction has similarity 0, discount 0, and is comparable. Negative values take part only in R1. A negative against a negative is ignored. Equal keys break in this order: displayed first, then effective support descending, then `Label` ascending, then `Rank` ascending (a cache value) or `ObservationID` (a Subject value), then input order; the probe value decides first, then the candidate value. **Cost bound.** When every value on both sides is plain (not `Counted`, not `Negative`, not `Denied`), `pairValues` runs the loop `best` runs today: no sorting, no `Label` call, no allocation. Otherwise it finds the leaders in one pass per side and the best pair by key comparison, without sorting copies. `Label` is computed only for the reported pair and for equal keys. | One rule picks the pair everywhere. R3 judges against the consensus, so a minority that agrees does not rescue a mismatch. AC-2.2 follows from R1 alone: the probe has no positive value, so no leader pair exists on its side. On a `multiple` Property (`Feature.Cardinality`, `pairValues(cmp, card, probe, candidate)`) every voting value is a leader with weight 1 and R3 is skipped; R1 and R2 run unchanged, and `outcomeAtFloor` already reads a zero on a multiple Property as unknown. Amended 2026-10-16 in the ship log (CON-1, user decision). | Keeping `best` and adding a second function — two selection rules |
| D-5 | "Compatible" is `autoreconcile.Compatible`. The value type is read from the fields a `Value` sets: `Name` → `name`, `Date` → `date`, `Term` non-empty → `term`, `HasInteger` → `integer`, `HasText` → `text`. `exhibitAuto` moves to `match.Value.Auto()`, and `exhibits.go` calls it. `Auto()` keys a term by its key: when `Term` is non-empty it sets `TermID` to `[]byte(Term)`, ignoring any catalog id, because `SubjectsValues` and `scanEntityValues` set only the key while `handleValues` sets the id, and `Compatible` compares `TermID` as a string. There is no similarity threshold. | The reconciler's own test decides what a negative counts against, including folds (`J. Robins` against `James Robins`). Two tests for "the same value" would drift as dates and names are tuned. | A `SameValue` similarity constant — disagrees with the reconciler both ways (a nickname scores 0.98; an initial scores under 0.95) |
| D-6 | Negative values ride in `Values` with `Value.Negative` set. `Values.HasPositive()` *(new)* reports whether any value is positive. `uFor` skips negatives. `subjectHasValues`, the `canon.go` gate, and the `rank` early return use `HasPositive`. | `Values` is the type every caller already passes. A parallel map would need a new parameter on `Rank`, `Evaluate`, and `Align`. | A parallel `Negatives` map — touches five signatures |
| D-7 | `Reason` and `PropertyComparison` each gain `Probe` and `Candidate`, both `ValueRef{Label string; ObservationID []byte; Negative bool}`, set from the `PairResult`. `Value.Label()` is the text; the term key; the integer in decimal; `datevalues.CompactDisplay` for a date; for a name its `Form`, else its part values joined by spaces. Only a probe value loaded from an Observation has an id. | "Names the two values (or their ids)" with no schema change. A candidate value is a cache cluster, which has no single Observation. Embedding `PairResult` would rename S10-27's `Reason.Similarity` and `Reason.Discount` fields. | Candidate Observation ids from `auto_reconciler_outcomes` — a second query per candidate |
| D-8 | `Value.SetEvidence(support, against, rank int, reason string)` *(new)* sets `Counted`, `Support`, `Against`, `Rank`, `Reason`, and `Denied` (`reason == autoreconcile.ReasonDenied`). `scanEntityValues` and `handleValues` both call it. `scanEntityValues` selects `r.support, r.against, r.rank, r.reason`, and the three entity-value queries order by `r.rank` last (`ORDER BY e.ref, r.entity_id, r.rank`; `ORDER BY r.rank` for one handle; `ORDER BY r.entity_id, r.rank` for a batch). `SubjectsValues` drops the positive filter, selects `o.id` and `o.polarity`, passes `) ORDER BY o.id` as the `sqlbatch.Query` tail (order holds within a batch, and one Subject's rows share a batch), and sets `Negative` and `ObservationID`. | One place maps the cache columns to the fields. The loaders are the only places the data enters. | Copying the mapping into both loaders — the near-copy the review found |
| D-9 | `Rank`'s `MinScore` cut and Promote's scales are unchanged. A contradiction reaches them as similarity 0. A weighted credit below 1 reaches `ApplyScale` as it is, so a minority exact match (0.2, 0.5) is a partial at or above the scale's floor and a conflict below it. `PropertyComparison.Similarity` is now `similarity × weightA × weightB`, so `Pinned` (outcome `agree`) is false for a minority exact match. | Nothing downstream needs to know why. A probe that matches only a minority is weaker evidence, and the name scale's 0.75 floor says how weak is too weak. | A new outcome — no consumer would read it |
| D-10 | `exhibits.go` keeps its own member picker. A Promote row can read conflict while an exhibit line shows an agreeing pinned pair for a minority value; the sheet also shows the row's conflict as a line beside it. | Exhibits compare an incoming Observation with member Observations, which carry no support; the cache cluster does. | Wiring `pairValues` into the exhibit loop — S10-27 already reshaped it, and the sheet's job is to show each pair. Amended 2026-10-16 in the ship log (UX-1, user decision). |
| D-11 | The file is `core/match/pairs.go` *(new)*. | `pair.go` is pairwise evaluation; this is pair selection. | `support.go` — names the input, not the job |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/match/pairs.go` *(new)* | file | `core/match/match.go` `best` | The weights, the three rules, and `PairResult` outgrow a ten-line helper. `best` is deleted |
| `match.PairResult`, `match.ValueRef`, `pairValues` *(new)* | type, type, func | S10-27 D-8 `CompareDiscount`, S10-27 `best` | The result is more than a triple now |
| `Value.Counted`, `Support`, `Against`, `Reason`, `Denied`, `Negative`, `ObservationID`, `Value.SetEvidence`, `Value.Label`, `Value.Auto`, `Values.HasPositive` *(new)* | fields, methods | `graphcache.Value` `Support`, `Against`, `Reason`; `exhibitAuto` | `graphcache.Value` is the cache row; `match.Value` is the shape comparers read. `exhibitAuto` moves, it is not copied |
| `Reason.Probe`, `Reason.Candidate`, `PropertyComparison.Probe`, `PropertyComparison.Candidate` *(new)* | fields | `Reason.Property` | See D-7 |

**Builds on.** The cache columns the auto-reconciler already fills; `graphcache.Value`; `autoreconcile.Compatible` and the `Reason*` constants; the pair-selection seam S10-27 created.
**Deletes.** `match.best`. `exhibitAuto` in `exhibits.go` (it moves to `match.Value.Auto`).
**Next adds.** A concluded value (Reconciliation Claims) becomes another input to the weight in `pairValues`, and the displayed-value floor (D-3) already lets a concluded value with support 0 vote. A reason sentence for Promote reads `Reason.Probe` and `Reason.Candidate` and adds no data.

### Schema

None — no catalog change. The columns read exist in `auto_reconciler_values`.

### Go core

| Package | Change |
| --- | --- |
| `core/match` `match.go` | `Value` gains the fields of the Codebase-fit row. `Reason` gains `Probe`, `Candidate`. `Score` calls `pairValues`; `best` is deleted. `Values.HasPositive` |
| `core/match` `pairs.go` *(new)* | `ValueRef`, `PairResult`, `pairValues` (reads each pair through `CompareDiscount`), `Value.SetEvidence`, `Value.Label`, `Value.Auto` (D-1 through D-8) |
| `core/match` `pair.go` | `compareProperty` calls `pairValues`; `PropertyComparison` gains `Probe`, `Candidate` |
| `core/database/matching` `matching.go` | `sqlEntityValues` selects `r.support, r.against, r.reason`; `scanEntityValues` calls `SetEvidence`. `sqlSubjectsValues` selects `o.id, o.polarity` without the positive filter and orders by `o.id`; `SubjectsValues` sets `Negative` and `ObservationID`. `rank` returns early when `!probe.HasPositive()` |
| `core/database/promotealign` `canon.go` | `handleValues` calls `SetEvidence`. The line-80 gate uses `HasPositive` |
| `core/database/promotealign` `exhibits.go` | `exhibitAuto` is replaced by `match.Value.Auto()` at its call sites. No other change |
| `core/graphalign` | `uFor` skips `Negative` values. `subjectHasValues` uses `HasPositive` |

### FFI

None — no message changes.

### macOS

A Promote row can move from agree to conflict when a handle's majority disagrees with the incoming Subject. The sheet then shows a conflict line (and the callout) for the Property: a row-level conflict with no conflicting exhibit gets an unpinnable line, and an incoming negative Observation reads "not <value>" (`L10n.Promote.negatedValue`). No layout change. Amended 2026-10-16 in the ship log (UX-1, user decision).

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Value shape and the pair function.**
Files: `core/match/match.go`, `core/match/pairs.go` *(new)*, `core/match/pairs_test.go` *(new)*.
Do: the `Value` fields and methods, `ValueRef`, `PairResult`, and `pairValues` with D-1 through D-5 and D-11. Keep `best` until STEP-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match -run 'TestPairValues|TestValueLabel|TestSetEvidence'` passes (T-1, T-2, T-3).

**STEP-2 — Use it in Score and Evaluate.**
Files: `core/match/match.go`, `core/match/pair.go`, `core/match/match_test.go`, `core/match/pair_test.go`, `core/match/pairs_test.go` *(new)*.
Do: `Score` and `compareProperty` call `pairValues`; delete `best`; `Reason` and `PropertyComparison` carry the pair (D-7).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match` passes, with every existing case unchanged (T-4, T-5). `CGO_ENABLED=1 go test -tags fts5 -bench BenchmarkPairValues -run '^$' ./core/match` reports 0 allocs/op for the plain-values case (T-10).

**STEP-3 — Load support, against, reason, and negatives.**
Files: `core/database/matching/matching.go`, `core/database/matching/matching_test.go`.
Do: D-8 for `scanEntityValues`, the three entity-value queries' order, and `SubjectsValues` with its `) ORDER BY o.id` tail; the `rank` early return (D-6).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/matching` passes, including `TestForSubjectQueryCountIsConstant` (T-6, T-7).

**STEP-4 — Promote's handle values, the readers, and the moved adapter.**
Files: `core/database/promotealign/canon.go`, `core/database/promotealign/exhibits.go`, `core/graphalign/score.go`, `core/graphalign/align.go`, `core/database/promotealign/promotealign_test.go`.
Do: `handleValues` calls `SetEvidence`; the readers in D-6; `exhibitAuto` → `Value.Auto`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/graphalign ./core/database/promotealign` passes (T-8).

**STEP-5 — Docs.**
Files: `docs/matching.md`, `docs/deployment-plan/spike-10/name-matching-enhancements.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-28-support-negatives.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Judge a handle's values by their support."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/match/pairs_test.go` `TestPairValues` *(new)* | Table-driven with hand-built `Value`s and `TermComparer`. Male 3 / female 1: probe female → similarity 0, pair (`female`, `male`); probe male → 1. Male 2 / female 2: probe female → 1. Male 1 `kept` / female 3 `weak`: probe female → 0, probe male → 1. `unknown` neutral 5 / male 1: probe female → 0. Male (3, against 2) / female 2: probe female → 1, probe male → 0. Sole `kept` male with against 3 / support 1, and with support 0: probe male → 1, probe female → 0. `Robins` 3 / `Robbins` 1 with `NameComparer`, probe `Robbins` → 0.857. Values with `Counted` false → equals the old best. Equal keys reported in the D-4 order, whatever the input order | AC-1.1, AC-1.2, AC-1.3, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-2.5, AC-3.3, D-1, D-3, D-4 |
| T-2 | Go | same file `TestPairValuesNegatives` *(new)* | Negative male vs leader male → contradiction. Negative male vs minority male → no contradiction, no credit. Negative male alone vs female-led handle → not comparable. Positive female plus negative male vs male leader → contradiction. `Denied` male, probe male → contradiction. A `weak` value with against ≥ support → weight 0: no credit, no contradiction. A negative never raises similarity. `Compatible` folds: negative `J. Robins` vs leader `James Robins` → contradiction | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-2, D-4, D-5 |
| T-3 | Go | same file `TestValueLabel`, `TestSetEvidence` *(new)* | Text, term, integer, date, a name with a form, and a parts-only name each give their label. `SetEvidence` sets `Counted` and `Denied` from `denied` only | AC-3.1, D-7, D-8 |
| T-4 | Go | `core/match/match_test.go` `TestScoreSupport` *(new)* | `DefaultProfile("person")`. Probe `James Robins`, handle `Mary Smith` 5 / `James Robins` 1 → name similarity 0, contribution equal to 0, reason pair (`James Robins`, `Mary Smith`); compare the contribution with `== 0`. Probe `Mary Jones`, handle `Mary Smith` 1 / `Mary Jones` 1 → 10. Probe `Robbins`, handle `Robins` 3 / `Robbins` 1 → similarity 0.857. The sex cases of AC-1.1 and AC-1.2 → −2 and +1 | AC-1.1, AC-1.2, AC-1.4, AC-1.5, AC-1.6, AC-3.1 |
| T-5 | Go | `core/match/pair_test.go` `TestEvaluateStrayName` *(new)* | `EvaluateWithNames` for the stray-name case → `conflict`, with the pair on the comparison | AC-3.2 |
| T-6 | Go | `core/database/matching/matching_test.go` `TestForSubjectJudgesAgainstTheMajority` *(new)* | `f.handle("person", named("Mary Robins"), termed("sex_at_birth", "male"))`; then by SQL: `UPDATE auto_reconciler_values SET support = 3` for the handle's sex row, and `INSERT INTO auto_reconciler_values (entity_id, property_id, rank, value_term_id, support, against, reason) VALUES (…, 2, <female term id>, 1, 0, 'kept')`. `ForSubject` for a Subject `Mary Robins`, female → the sex reason contribution is −2 and names `female` and `male`. A second case, on a fresh handle: `UPDATE` its male row to reason `denied`, insert the female row (rank 2, support 1, `kept`, so female is the only leader), and a male Subject `Mary Robins` → −2 | AC-1.1, AC-2.4, D-8 |
| T-7 | Go | same file `TestSubjectsValuesNegatives` *(new)* | A Subject with a positive and a negative sex Observation (`observations.Input{Polarity: observations.PolarityNegative}`) → both values, the negative with `Negative` and its Observation id, in `o.id` order. A Subject with only the negative → `ForSubject` returns no matches. `TestForSubjectQueryCountIsConstant` still holds | AC-2.1, AC-2.6, AC-2.7, D-8 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeJudgesAgainstTheMajority` *(new)* | Promote a `Mary Smith` person; by SQL set its name row's support to 5 and insert a second name row (rank 2, `value_name` from `valuecodec.MarshalName`, support 1, reason `kept`) for `James Robins`; call `f.c.Graph().Drop()`; an incoming `James Robins` → the name comparison is a conflict, not an agree. A second case: a Subject `Mary Smith` with a negative `sex_at_birth` male Observation, against a handle `Mary Smith` whose cache sex row is male with support 3 → the sex comparison is a conflict | AC-3.2, AC-2.8, D-5 |
| T-9 | Go | `core/graphalign/align_test.go` `TestScoreSimilarityWeightedCredit` *(new)* | Table-driven on `ScoreSimilarity` with the name Property: similarity 0.5, discount 0 → `conflict`, −3. Similarity 0.8 → `partial`, 2.0. Similarity 1 → `agree`, 2.5 | D-9 |
| T-10 | Go | `core/match/pairs_test.go` `BenchmarkPairValues` *(new)* | Plain values (not `Counted`, no negatives): 0 allocs/op. A `Counted` case is reported, with no bar | D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. A pre-existing test whose expected number changes is a stop-and-report, not an edit (see Stop conditions).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host with no xcodebuild, the Go and plan-doc steps pass and the Xcode step is reported as not run
- [ ] No pre-existing test needed a changed expectation; if one did, the work stopped and reported
- [ ] BenchmarkProposeColdHub ns/op and BenchmarkPairValues allocs/op are reported against the base; the first rises by under 15%
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] No new user-facing copy; no Swift change
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-28 Plan line status; the S10-28 "Depends on" reads "S10-24; sequenced after S10-27"; the S10-28 "In" says "relative to the best-supported value"; completed.md: one short paragraph of what shipped
- [ ] docs/matching.md: the file table lists `pairs.go` and the new `match.go` and `pair.go` roles; a section on support weights and negatives with the stray-name and one-female examples; the "Not yet handled" bullet no longer names support weighting (leave the other items in that bullet as they are)
- [ ] name-matching-enhancements.md: "Weigh values by support" and "Negative evidence" say they landed in S10-28 and quote the leader-relative rule
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice has no new chrome, so the check is a catalog run, not a screen.

1. Run `CGO_ENABLED=1 go test -tags fts5 -run 'TestScoreSupport|TestProposeJudgesAgainstTheMajority' -v ./core/match ./core/database/promotealign`. See the stray-name case score 0 against a mostly-Mary-Smith handle, and the Promote name comparison read as a conflict.
2. In a project where one handle has `male` on three Sources and `female` on one, add a Subject whose Observation says female and open Promote. The suggestion's sex comparison reads as a conflict, and the sheet shows a conflict line and the callout.
3. Add a negative Observation "sex at birth: male" to a Subject whose handle is male on most Sources. Promote reads the sex comparison as a conflict, and the sheet shows a conflict line reading "not male" that cannot be pinned.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-27's re-based test numbers move again when a fixture handle holds values of unequal support | Stop-and-report rule below; no existing expectation is edited in this PR |
| A mixed handle that was a clean match before drops in Promote | Intended (US-1, AC-1.4). T-4 and T-8 assert it |
| A name seen on one Source against a different name on five loses its name credit (a married name against a maiden name, probe on the married name) | Intended: R3 judges against the consensus, so that match is a contradiction in Promote's scale and zero in `Rank`. Equal support keeps both names (AC-1.5). Known-as handling is later ([`docs/ideas/archive/name-matching-enhancements.md`](../../../ideas/archive/name-matching-enhancements.md)) |
| A negative-only Subject reaches a reader that treats any value as a value | D-6 names the four readers; T-7 asserts the `ForSubject` case |
| A Promote row reads conflict while its exhibit lines show an agreeing pair | Accepted (D-10); the conflict line now appears beside the agreeing pair |
| `pairValues` runs for every candidate and every Feature | D-4's cost bound and T-10; the cold-hub benchmark is reported |

**Stop and report if:**

- `best` is no longer the single pair-selection seam §4 describes (callers `Score` and `compareProperty` only). Redo §4 and tell the user what moved.
- Any existing test in `core/match`, `core/database/matching`, `core/database/promotealign`, or `core/graphalign` needs a changed expected value to pass.
- `auto_reconciler_values` does not fill `support`, `against`, and `reason` for a Property this plan scores (a constant 0 for every row).
- A caller of `best` exists outside `Score` and `compareProperty`.
- `SubjectsValues` has a caller outside `loadSubjectValues` and `layer.go` that cannot tolerate a negative value.
- `autoreconcile.Compatible` cannot be called from `core/match` without an import cycle.
- `BenchmarkProposeColdHub` rises by 15% or more against the base and `pairValues` is the cause.

## 12. Open questions

None.

## 13. PR

**Title:** S10-28 — Support, negatives, and which values matched
**Body:** A handle's values are now judged by their support, negative Observations count against a match, and each reason names the two values; what shipped against US-1, US-2, and US-3; AC evidence; deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-28-support-negatives.md`.
