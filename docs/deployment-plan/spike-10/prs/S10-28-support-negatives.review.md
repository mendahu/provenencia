# S10-28 review log

**Plan:** [`S10-28-support-negatives.md`](S10-28-support-negatives.md)
**Lint:** `python3 scripts/check-plan-docs.py` — 0 errors, 0 warnings, run after the round-1 fixes
**Round 1:** maintainability reviewer: Fits with changes. Executability reviewer: Not ready (two blockers). Both reviewed against `a4e9de4` with S10-27's plan as the assumed base.
**Round 2:** run 2026-10-16 after S10-27 landed (#409) and the plan's Current state was redone against `6577a43`. Maintainability: no new blocker, two majors (Fits with changes). Executability: two blockers, one major (Not ready). All answered below.
**Verdict:** Planned. No open blocker. Nothing for the user.

## Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | `SameValue` 0.95 is a second definition of "same value" beside the reconciler's | **Fixed.** R1 and R2 use `autoreconcile.Compatible`; `exhibitAuto` moves to `match.Value.Auto` (D-5). No similarity constant |
| M-2 | blocker | Leaders ignore `reason`, so a non-displayed value could lead | **Fixed.** D-3: leaders are the `kept` voting values; neutral and `provisional` never vote. D-2 is narrowed to the user's wording: only `denied` contradicts, and a `kept` value with `against ≥ support` is not disputed. AC-1.8 and AC-1.9 test it |
| M-3 | major | R3 and fractional weights fight each other; the Risks row claims the opposite | **Fixed in part.** The leader-relative weights stay: the user chose them (2026-10-16), and they act wherever a comparer returns partials (names, dates, text; AC-1.6, AC-2.5). The Risks row now says the married-name case loses credit by design; AC-2.5 is corrected. Leader-set-only was declined for that reason |
| M-4 | major | Exhibits are a third consumer; the Precedent says otherwise | **Fixed.** D-10, an Out line, and a corrected Precedent: exhibits keep their own picker |
| M-5 | major | Near-copies: value builder, label chain, Reason fields | **Fixed in part.** D-8 adds `SetEvidence`, called by both loaders. **Declined:** unifying `Value.Label` with the exhibit chain (exhibits take term labels from catalog SQL that `match.Value` does not have; Out line). **Declined:** embedding `PairResult` in `Reason` (it would rename S10-27's `Reason.Similarity` and `Reason.Discount`; D-7) |
| M-6 | major | Edge cases undecided | **Fixed.** D-4: tie-break order, negatives excluded from R3, empty voting set is not comparable, contradiction discount 0. D-3: `Counted` replaces "support 0 means unknown". D-2: units note. D-4: cardinality note |
| M-7 | minor | Negatives in shared `Values`; readers missed | **Fixed.** D-6 adds `Values.HasPositive` and the two missed readers (`canon.go` gate, `rank` early return) |
| M-8 | minor | File name, `matching.md` file table, Depends-on drift, AC-1.4 two views | **Fixed.** `pairs.go` (D-11); DoD updates the file table and the map's Depends-on and In lines; AC-1.4 states both |

## Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | blocker | R3 with negative probe values contradicts AC-2.2 | **Fixed.** Negatives take part only in R1 (D-4); AC-2.2 follows because the probe has no positive leader |
| X-2 | blocker | Negatives on the candidate side and disputed values on the probe side undefined | **Fixed.** D-4: rules are symmetric; a negative against a negative is ignored |
| X-3 | major | `pairValues` edge cases left to the implementer | **Fixed.** Same as M-6 |
| X-4 | major | Exhibits not covered; Precedent wrong | **Fixed.** Same as M-4 |
| X-5 | major | D-5's claim about names reaching 0.95 is false | **Fixed by removal.** There is no threshold now (D-5) |
| X-6 | major | Two readers missing from D-6 | **Fixed.** Same as M-7 |
| X-7 | major | T-6 and T-8 fixtures not writable | **Fixed.** T-6 and T-8 give the SQL, the probe name that clears `MinScore`, and `Graph().Drop()` after the edit |
| X-8 | minor | Real `support` may move existing numbers | **Fixed.** A stop condition and a DoD box: a changed expectation in a pre-existing test is a stop-and-report |
| X-9 | minor | Label fallbacks, `Negative` flag, command forms | **Fixed.** D-7 (`Label` rules, `ValueRef.Negative`); the same `CGO_ENABLED=1 go test -tags fts5` form in every check; D-7 sets the fields from the `PairResult` |
| X-10 | minor | Weighted credits feed the scale floor | **Fixed.** D-9 states it; T-8 has the case |
| X-11 | nit | Doc drift | **Fixed.** DoD |

## Round 2

Round-1 answers were checked by both reviewers: M-1, M-2, M-4, M-6, M-7, X-1, X-2, X-5 confirmed fixed; the M-3 and M-5 declines confirmed as backed by evidence. Every Current state line number was re-checked against the landed code.

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| R2-M1 | major | D-3 drops a handle's only displayed value when `against ≥ support`, or a concluded value with support 0, so the feature is skipped | **Fixed.** D-2, D-3: a `kept` value is floored at effective support 1 and always votes; other values floor at 0. AC-1.10, T-1 |
| R2-M2 | major | `pairValues` sorts and builds labels per candidate per Feature; no cost bound | **Fixed.** D-4 cost bound: plain values run today's `best` loop with no sort, no `Label`, no allocation; leaders in one pass otherwise; `Label` only for the reported pair and ties. T-10 (`BenchmarkPairValues`), the cold-hub benchmark in the DoD and Stop conditions |
| R2-M3 | minor | `ORDER BY o.id` belongs in the `sqlbatch.Query` tail; AC-3.3 needs a final tie-break | **Fixed.** D-8 names the `) ORDER BY o.id` tail and adds `r.rank` to the entity-value orders; D-4's tie-break ends at `Rank` or `ObservationID`, then input order |
| R2-M4 | minor | `Similarity` is now weighted; `Pinned` goes false for a minority match | **Fixed.** D-9 says both |
| R2-M5 | minor | Header and log still say round 2 pending; line drift | **Fixed.** This section; `comparerDiscount` 257, `rank` 159 |
| X-12 | blocker | AC-2.5 "probe male +0.5" contradicts R3 | **Fixed.** With effective supports 1 and 2 the consensus is female, so probe male is a contradiction (−2); AC-2.5 and T-1 say so. A weight example R3 does not preempt is in T-1 (`Robins` 3 / `Robbins` 1, similarity 0.857) |
| X-13 | blocker | `Auto()` and `Compatible` never match a term across the two loaders: `handleValues` sets a catalog id, the other loaders set only the key | **Fixed.** D-5: `Auto()` keys a term by `Term`, ignoring the catalog id. AC-2.8 and a T-8 case cover the Promote path. Exhibits are unaffected when both sides carry the key, as they do today |
| X-14 | major | T-6's `denied` case inserts the wrong row | **Fixed.** T-6: update the male row to `denied`, insert a `kept` female so it is the only leader |
| X-15 | minor | T-8 floor case has no fixture; `ForEntity` probe weighting unstated; sort tie; negative zero | **Fixed.** The floor case moved to T-9 (`ScoreSimilarity` table). D-3 says the `ForEntity` probe is weighted. D-4 and D-8 end the tie-break at `Rank` or `ObservationID`. T-4 compares with `== 0` |
