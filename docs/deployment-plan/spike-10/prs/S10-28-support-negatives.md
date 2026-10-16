# S10-28 — Support, negatives, and which values matched

**Status:** Blocked — on S10-27
**Spike / slice:** Spike 10 · Slice 7 — How names compare (no new chrome)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-28--support-negatives-and-which-values-matched)
**Brief:** — (no new chrome: the score and its reasons change, no view does)
**Board:** —
**Depends on:** S10-24 (landed [#378](https://github.com/mendahu/provenencia/pull/378)); sequenced after S10-27 ([plan](S10-27-name-frequency.md), in implementation), whose pair-selection seam this PR extends
**Blocks:** —
**Branch:** `s10-28-support-negatives`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`name-matching-enhancements.md`](../name-matching-enhancements.md), [`docs/conclusion-reconciliation.md`](../../../conclusion-reconciliation.md) §4–§6, [`docs/matching.md`](../../../matching.md), [S10-27 plan](S10-27-name-frequency.md) D-8 and D-9, working rule "Support weighting and negatives land together" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-16 against `main` @ `a4e9de4` plus the S10-27 plan as merged. S10-27 code has not landed. Redo §4 on landing (see Stop conditions)
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

A handle's values count in proportion to how well supported they are, and a negative Observation counts against a match. A handle that is mostly one sex, or mostly one name, is judged against that majority, so one stray reading cannot make a mixed handle look like a clean match. Each reason says which two values produced it.

**Why now.** Slice 7 makes names compare the way records are written. S10-24, S10-25, and S10-27 changed what a name word is worth. This PR changes which of a handle's values are allowed to speak. It follows S10-27 so fixtures settle against frequency-weighted scores once.

**What stays.** A handle with one value, or with values of equal support, scores exactly as before. The auto-reconciler, the cache, and every Promote view are untouched. No catalog change: the cache already stores `support`, `against`, and a `denied` reason on every value, and the matcher never read them.

## 2. User stories

**US-1.** As a researcher, when a Subject says "female" and the handle it might belong to is "male" on three Sources and "female" on one, I see a contradiction, so that one dissenting reading does not hide a sex mismatch.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Probe `sex_at_birth` female. Handle values male (support 3), female (support 1). Person profile | Contribution −2. The reason names `female` and `male` |
| AC-1.2 | Probe male, the same handle | Contribution +1. Reason names `male` and `male` |
| AC-1.3 | Probe female. Handle values male (support 2), female (support 2) | Contribution +1: both values lead, so the match stands |
| AC-1.4 | Probe `James Robins`. Handle names `Mary Smith` (support 5), `James Robins` (support 1) | Name similarity 0, contribution 0 (the name feature has no contradiction points). The reason names `James Robins` and `Mary Smith`. Not 10 |
| AC-1.5 | Probe `Mary Jones`. Handle names `Mary Smith` (support 1), `Mary Jones` (support 1) | Name similarity 1, contribution 10: equal support keeps both names in full |
| AC-1.6 | Probe `Robbins`. Handle names `Robins` (support 3), `Robbins` (support 1) | Name similarity 0.857, with the pair `Robbins` and `Robins`, not 1 |
| AC-1.7 | Values with support 0 (unknown) | Every weight is 1 and scores equal today's |

**US-2.** As a researcher, when a Subject carries a negative Observation ("not male", "not born in York") or the handle carries evidence against one of its values, I see that count as a contradiction.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Probe has a negative `sex_at_birth` male Observation. Handle male (support 3), female (support 1) | Contribution −2. The reason names the negative value and `male` |
| AC-2.2 | The same negative probe against a handle that is female (support 3), male (support 1) | No contradiction: the negative matches only a minority value. The feature scores from the probe's positives, or is not comparable when the probe has none |
| AC-2.3 | A negative probe value matches a candidate value | The negative never adds credit |
| AC-2.4 | Candidate value `male` is `denied`, or has `against` ≥ `support`. Probe male | Contribution −2: the probe asserts a value the handle's evidence refutes. The reason names the two |
| AC-2.5 | Candidate values male (support 3, against 2), female (support 2, against 0). Probe female | Both values lead after the reduction (1 and 2 → weights 0.5 and 1 for male and female). Probe female scores +1 |
| AC-2.6 | `SubjectsValues` for a Subject with one positive and one negative Observation | Both come back; the negative has `Negative` set and its Observation id |

**US-3.** As a contributor, when I read a `Reason`, I can tell which two values produced it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Any credited or contradicted reason | `Reason.Probe` and `Reason.Candidate` carry the value's label; the probe's carries its Observation id when it has one |
| AC-3.2 | `EvaluateWithNames` for the stray-name case of AC-1.4 | The comparison is a conflict, not an agree |

## 3. Scope

| In | Covered by |
| --- | --- |
| Weight each value by its support relative to the leading value on its side, after subtracting `against` | STEP-1, STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-2.5 |
| Judge contradictions against the leading values of each side | STEP-2, AC-1.1, AC-1.3, AC-1.4 |
| Load negative Observations of a Subject; score a negative that matches a leading value as a contradiction | STEP-3, AC-2.1, AC-2.2, AC-2.3, AC-2.6 |
| Load `support`, `against`, and `denied` from the handle cache, for both the matching loader and Promote's graph loader | STEP-3, STEP-4, AC-2.4, AC-2.5 |
| `Reason` names the probe value and the candidate value | STEP-1, STEP-2, AC-3.1 |
| Docs: `matching.md`, the spike design note, the map, `completed.md` | STEP-5 |

| Out | Home |
| --- | --- |
| A concluded value dominating the others (Reconciliation Claims) | [`docs/ideas/name-matching-enhancements.md`](../../../ideas/name-matching-enhancements.md) |
| Remembering a rejected claim | same idea note |
| A Promote layout for the reason text, or a sentence built from the pair | S10-36 covers crops only; not planned |
| A support weight on the incoming Subject's own values (each Observation counts 1) | not planned |
| Claim confidence or Source credibility as a weight | not planned: the spike decided "share of support, not rank and not claim confidence" |
| Changing the auto-reconciler, its cache, or `against` counting | not planned |
| A migration | not planned |

## 4. Current state

Verified at `a4e9de4` (S10-27's plan merged, no S10-27 code). Rows marked **S10-27** describe the plan's design, not code on `main`. Check them on landing.

| Area | Where | Today |
| --- | --- | --- |
| Pair selection | `core/match/match.go` `best` | The highest similarity over every comparable pair. Two values of any support count alike. **S10-27** changes it to pick by `similarity × (1 − discount)` and return `(similarity, discount, ok)` (S10-27 D-9) |
| Scoring | `core/match/match.go` `Score`, `Reason`, `Rank` | `Contribution = Weight × Similarity`, or `−Contradiction` at similarity 0. `Reason` has `Property`, `Similarity`, `Contribution`. **S10-27** adds `Reason.Discount` |
| Pairwise evaluation | `core/match/pair.go` `evaluate`, `compareProperty`, `PropertyComparison` | Calls `best`. A comparable zero is a conflict (single) or unknown (multiple). **S10-27** adds `PropertyComparison.Discount` |
| Value shape | `core/match/match.go` `Value` | No support, no polarity, no label |
| Comparers | `core/match/compare.go` `Comparer`; **S10-27** `DiscountComparer`, `CompareDiscount` | `TermComparer` returns 0 for two different terms and not comparable for a neutral one |
| Handle cache | `auto_reconciler_values` (`support`, `against`, `reason`, `rank`); `core/database/autoreconciler/autoreconciler.go` | One row per distinct value. `support` is distinct Sources, `against` is negatives that counted against the value, `reason` is `kept`, `outvoted`, `weak`, `denied`, or `provisional` |
| Matching loader | `core/database/matching/matching.go` `sqlEntityValues`, `scanEntityValues` | Selects neither `support`, `against`, nor `reason` |
| Subject loader | `core/database/matching/matching.go` `sqlSubjectsValues`, `SubjectsValues` | `WHERE o.polarity = 'positive'`: negatives never load. Shared by `ForSubject` and `core/database/promotealign/layer.go` |
| Promote's handle values | `core/database/graphcache/graph.go` `Value` (`Support`, `Against`, `Reason`); `core/database/promotealign/canon.go` `handleValues` | The graph row has all three; `handleValues` copies none of them |
| Readers of `Values` that would see a negative | `core/graphalign/score.go` `uFor`; `core/graphalign/align.go` `subjectHasValues` | Both iterate every value of a Property. `exhibits.go` and `terms.go` already filter `polarity = 'positive'` in SQL |
| Tests | `core/match/match_test.go`, `core/match/pair_test.go`, `core/database/matching/matching_test.go` (`newFixture`, `handle`, `termed`, `named`) | Green at the base |

**Precedent.** S10-27 added a per-pair factor to `best`, `Score`, and `Evaluate` through one seam (its D-8, D-9). Copy that: one function selects and reports the pair, and `Score`, `compareProperty`, and the exhibit chooser all call it. Do not copy its optional-interface plumbing: this PR needs none.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A value's weight is its effective support divided by the highest effective support among the standing values on its side (D-3). The leading value weighs 1. User answer 2026-10-16. | A stray one-of-five weighs 0.2. Two equally supported names (a maiden and a married name) both weigh 1, so a probe that matches either is a full match. | Share of the total — two equal names would each weigh 0.5, so a probe matching either would score a half match |
| D-2 | A candidate value is **disputed** when its reason is `denied`, or when `support > 0` and `against ≥ support`. A disputed value casts no credit, and a probe that asserts it contradicts. A value with `against < support` stays standing with effective support `support − against`. User answer 2026-10-16. | The reconciler already decided that a stronger negative refutes a `denied` value. `against` is the count of negatives that did not win but still count. | Reading only the incoming Subject's negatives — leaves the evidence the reconciler already cached unused |
| D-3 | Effective support is `support − against` when `support > 0`, else 1 (unknown). A standing value is one that is not disputed. Weights are computed per side, for the Subject probe and for the handle candidate alike. A Subject's own values all weigh 1. | A handle probe (merge hints) has the same stray-reading problem as a handle candidate. Support 0 means the loader did not know, so scores equal today's (AC-1.7). | Weighting only the candidate — makes `ForEntity` asymmetric |
| D-4 | One function, `pairValues(cmp Comparer, probe, candidate []Value) PairResult`, returns the similarity, the discount, whether the pair is comparable, and the two values that produced it. `best` is deleted. Its rules, in order: **R1** a negative probe value that matches a leading candidate value is a contradiction. **R2** a probe value that matches a disputed value, with no standing value matching it, is a contradiction. **R3** when the leading values of both sides are all comparable and every leader pair has similarity 0, that is a contradiction. Otherwise the credit is the maximum of `similarity × weightA × weightB × (1 − discount)` over standing positive pairs, reported with similarity `similarity × weightA × weightB`. A contradiction reports similarity 0. | One rule picks the pair everywhere, which is also what S10-27 asked of `best`. R3 is "judge against the consensus": a minority that agrees does not rescue a mismatch. | Keeping `best` and adding a second function — two selection rules for `Score`, `Evaluate`, and exhibits to choose between |
| D-5 | "Matches" for R1 and R2 means similarity ≥ `SameValue` = 0.95 (a `const` in `registry.go` beside `AliasCredit`). Text, terms, and integers return 0 or 1, dates reach it only on an exact match, and names reach it on an exact or accent-fold match. | A negative says "not exactly this". A near-spelling or a partial date is not that value. | A setting — a threshold with one caller; a const is enough until a second appears |
| D-6 | Negative values ride in `Values` with `Value.Negative` set. Only `pairValues` consumes them. The two readers in §4 (`uFor`, `subjectHasValues`) skip a negative value. | `Values` is the type every caller already passes. A separate negatives map would need a new parameter on `Rank`, `Evaluate`, and `Align`. | A parallel `Negatives` map — touches five signatures |
| D-7 | `Reason` gains `Probe` and `Candidate`, both `ValueRef{Label string; ObservationID []byte}`. `Value.Label()` is the label: the text, the term key, the integer, `datevalues.CompactDisplay`, or the name's `Form`. Only a loaded probe Observation has an id. | "Names the two values (or their ids)" with no schema change. A candidate value is a cache cluster, which has no single Observation. | Candidate Observation ids from `auto_reconciler_outcomes` — a second query per candidate for an id that does not name a value |
| D-8 | `scanEntityValues` selects `support`, `against`, `reason` and sets `Value.Support`, `Value.Against`, `Value.Denied = reason == 'denied'`. `handleValues` in `canon.go` copies the same three from `graphcache.Value`. `SubjectsValues` drops the `polarity = 'positive'` filter, selects `o.polarity` and `o.id`, and sets `Negative` and `ObservationID`. | The loaders are the only places the data enters. The columns and the graph fields exist. | A second loader for negatives — a second query per read |
| D-9 | `Rank`'s `MinScore` cut and Promote's scales are unchanged. A contradiction reaches them as similarity 0, which both already treat as a conflict. | Nothing downstream needs to know why. | A new outcome — no consumer would read it |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/match/support.go` *(new)* | file | `core/match/match.go` `best` | The weights, the rules, and `PairResult` outgrow a ten-line helper. `best` is deleted |
| `match.PairResult`, `match.ValueRef`, `pairValues` *(new)* | type, type, func | S10-27 plan D-8 `CompareDiscount`, S10-27 `best` | The result is more than a triple now; the helper replaces `best` |
| `Value.Support`, `Against`, `Denied`, `Negative`, `ObservationID`, `Value.Label` *(new)* | fields, method | `graphcache.Value` `Support`, `Against`, `Reason` | The cache row type belongs to `graphcache`; `match.Value` is the shape comparers read |
| `Reason.Probe`, `Reason.Candidate` *(new)* | fields | `Reason.Property` | See D-7 |
| `SameValue` *(new)* | const | `AliasCredit` in `registry.go` | See D-5 |

**Builds on.** The cache columns the auto-reconciler already fills; `graphcache.Value`; the pair-selection seam S10-27 creates.
**Deletes.** `match.best`. Its two callers (`Score`, `compareProperty`) call `pairValues`.
**Next adds.** A concluded value (Reconciliation Claims) becomes another input to the weight in `pairValues`. A reason sentence for Promote reads `Reason.Probe` and `Reason.Candidate` and adds no data.

### Schema

None — no catalog change. The columns read exist in `auto_reconciler_values`.

### Go core

| Package | Change |
| --- | --- |
| `core/match` `match.go` | `Value` gains `Support`, `Against int`, `Denied`, `Negative bool`, `ObservationID []byte`. `Reason` gains `Probe`, `Candidate ValueRef`. `Score` calls `pairValues`; `best` is deleted |
| `core/match` `support.go` *(new)* | `ValueRef`, `PairResult`, `pairValues`, `Value.Label` (D-1, D-2, D-3, D-4, D-5, D-6, D-7) |
| `core/match` `pair.go` | `compareProperty` calls `pairValues`; `PropertyComparison` gains `Probe`, `Candidate ValueRef` |
| `core/match` `registry.go` | `const SameValue = 0.95` |
| `core/database/matching` `matching.go` | `sqlEntityValues` selects `r.support, r.against, r.reason`; `scanEntityValues` sets the three. `sqlSubjectsValues` selects `o.id, o.polarity` without the positive filter; `SubjectsValues` sets `Negative` and `ObservationID` |
| `core/database/promotealign` `canon.go` | `handleValues` copies `Support`, `Against`, and `Denied` from the graph row |
| `core/graphalign` | `uFor` and `subjectHasValues` skip `Negative` values |

### FFI

None — no message changes.

### macOS

None — no view changes. A Promote row can move from agree to conflict when a handle's majority disagrees with the incoming Subject; no copy, state, or layout changes.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Value shape and the pair function.**
Files: `core/match/match.go`, `core/match/support.go` *(new)*, `core/match/support_test.go` *(new)*.
Do: the `Value` fields, `ValueRef`, `PairResult`, `Value.Label`, and `pairValues` with D-1, D-2, D-3, D-4, D-5. Keep `best` until STEP-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/match -run 'TestPairValues|TestValueLabel'` passes (T-1, T-2, T-3).

**STEP-2 — Use it in Score and Evaluate.**
Files: `core/match/match.go`, `core/match/pair.go`, `core/match/registry.go`, `core/match/match_test.go`, `core/match/pair_test.go`.
Do: `Score` and `compareProperty` call `pairValues`; delete `best`; `Reason` and `PropertyComparison` carry the pair (D-7); add `SameValue`.
Check: `go test -tags fts5 ./core/match` passes, with every existing case unchanged (T-4, T-5).

**STEP-3 — Load support, against, denied, and negatives.**
Files: `core/database/matching/matching.go`, `core/database/matching/matching_test.go`.
Do: D-8 for `scanEntityValues` and `SubjectsValues`.
Check: `go test -tags fts5 ./core/database/matching` passes, including `TestForSubjectQueryCountIsConstant` (T-6, T-7).

**STEP-4 — Promote's handle values and the two readers.**
Files: `core/database/promotealign/canon.go`, `core/graphalign/score.go`, `core/graphalign/align.go`, `core/database/promotealign/promotealign_test.go`.
Do: `handleValues` copies the three fields; `uFor` and `subjectHasValues` skip `Negative`.
Check: `go test -tags fts5 ./core/graphalign ./core/database/promotealign` passes (T-8).

**STEP-5 — Docs.**
Files: `docs/matching.md`, `docs/deployment-plan/spike-10/name-matching-enhancements.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-28-support-negatives.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period ("Judge a handle's values by their support."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/match/support_test.go` `TestPairValues` *(new)* | Table-driven with hand-built `Value`s and `TermComparer`. Handle male 3 / female 1: probe female → similarity 0, pair (`female`, `male`); probe male → 1. Male 2 / female 2: probe female → 1. Support 0 on both → equals the old best. Male 3 against 2 and female 2: probe female → 1 | AC-1.1, AC-1.2, AC-1.3, AC-1.7, AC-2.5, D-1, D-3, D-4 |
| T-2 | Go | same file `TestPairValuesNegatives` *(new)* | Negative male vs leader male → contradiction. Negative male vs minority male → no contradiction, no credit. `Denied` male, probe male → contradiction. `Against ≥ Support` same. A negative never raises similarity | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-2, D-5 |
| T-3 | Go | `core/match/support_test.go` `TestValueLabel` *(new)* | Text, term, integer, date, and name values each give their label | AC-3.1, D-7 |
| T-4 | Go | `core/match/match_test.go` `TestScoreSupport` *(new)* | `DefaultProfile("person")`. Probe `James Robins`, handle `Mary Smith` 5 / `James Robins` 1 → name similarity 0, contribution 0, reason pair (`James Robins`, `Mary Smith`). Probe `Mary Jones`, handle `Mary Smith` 1 / `Mary Jones` 1 → 10. Probe `Robbins`, handle `Robins` 3 / `Robbins` 1 → similarity 0.857. Sex cases of AC-1.1 and AC-1.2 → −2 and +1 | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-3.1 |
| T-5 | Go | `core/match/pair_test.go` `TestEvaluateStrayName` *(new)* | `EvaluateWithNames` for the stray-name case → `conflict`, with the pair on the comparison | AC-3.2 |
| T-6 | Go | `core/database/matching/matching_test.go` `TestScanEntityValuesSupport` *(new)* | A handle whose cache rows are inserted by SQL: male support 3, female support 1 (reason `kept`), a `denied` row. `ForSubject` for a female Subject gets −2 from sex; reasons name the pair | AC-1.1, AC-2.4, D-8 |
| T-7 | Go | same file `TestSubjectsValuesNegatives` *(new)* | A Subject with a positive and a negative sex Observation → both values, the negative with `Negative` and its Observation id. `TestForSubjectQueryCountIsConstant` still holds | AC-2.1, AC-2.6, D-8 |
| T-8 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeJudgesAgainstTheMajority` *(new)* | A handle with cache rows `Mary Smith` support 5 and `James Robins` support 1; an incoming `James Robins`. The name comparison is a conflict, not an agree | AC-3.2, D-8 |

Fixture note: a handle's cache rows with chosen `support` are inserted with SQL into `auto_reconciler_values`, using the column list of the insert in `core/database/promotealign/bench_test.go` `bulkCanon` plus `support`, `against`, `reason`.

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break are `core/match` (name invariants) and `core/database/promotealign` (weights asserted as numbers).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); on a host with no xcodebuild, the Go and plan-doc steps pass and the Xcode step is reported as not run
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] No new user-facing copy; no Swift change
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-28 Plan line status; completed.md: one short paragraph of what shipped
- [ ] docs/matching.md: a section on support weights and negatives with the stray-name and one-female examples; the "Not yet handled" bullet no longer names support weighting (leave the other items in that bullet as they are)
- [ ] name-matching-enhancements.md: "Weigh values by support" and "Negative evidence" say they landed in S10-28 and quote the leader-relative rule
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The slice has no new chrome, so the check is a catalog run, not a screen.

1. Run `go test -tags fts5 -run 'TestScoreSupport|TestProposeJudgesAgainstTheMajority' -v ./core/match ./core/database/promotealign`. See the stray-name case score 0 against a mostly-Mary-Smith handle, and the Promote name comparison read as a conflict.
2. In a project where one handle has `male` on three Sources and `female` on one, add a Subject whose Observation says female and open Promote. The suggestion's sex comparison reads as a conflict.
3. Add a negative Observation "sex at birth: male" to a Subject whose handle is male on most Sources. Promote reads the sex comparison as a conflict.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-27 lands differently from its plan, so `best`, `Reason`, or `PropertyComparison` differ from §4 | §4 marks every S10-27 row. The first action is to redo them against the landed code |
| A mixed handle that was a clean match before drops in Promote | Intended (US-1, AC-1.4). The fixtures that assert it are T-4 and T-8 |
| A name with two real aliases of unequal support loses recall | The weaker alias weighs support ÷ leader support and still credits. Recorded in the spike note as the cost of "share of support" |
| A negative Observation on a Property with `Contradiction` 0 (names in `DefaultProfile`) scores nothing | Intended: the contradiction points belong to the profile. Promote's name scale has −3 |

**Stop and report if:**

- S10-27 has not landed, or `best` is not the single pair-selection seam §4 describes. Redo §4 and tell the user what moved.
- Any existing `core/match` case changes value when every `Value.Support` is 0.
- `auto_reconciler_values` does not fill `support` and `against` for a Property this plan scores (a table of NULL or constant 0 for every row).
- A caller of `best` exists outside `Score` and `compareProperty`.
- `SubjectsValues` has a caller that cannot tolerate a negative value in its result.

## 12. Open questions

None.

## 13. PR

**Title:** S10-28 — Support, negatives, and which values matched
**Body:** A handle's values are now judged by their support, negative Observations count against a match, and each reason names the two values; what shipped against US-1, US-2, and US-3; AC evidence; deviations from the plan; the manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-28-support-negatives.md`.
