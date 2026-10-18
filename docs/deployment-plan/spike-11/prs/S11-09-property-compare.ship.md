# Ship log — S11-09 Property compare

**Plan:** [`S11-09-property-compare.md`](S11-09-property-compare.md) · **Branch:** `s11-09-property-compare` · **PR:** #468
**State:** ready

Base `95d1f344` (`main`). The branch starts with the plan refresh commit `a743c97c`, made at the user's request before the loop.

## Round 0 — implement
Head `eddb1901`. Gate PASS (orchestrator run: diff-check, gofmt, go-vet, go-test, localizable, plan-docs, xcode-test). Report deviations: STEPs 3–7 in one commit (`b72f822b`), with the `propertycompare` guard row there; AC-2.6 checked in a separate `TestCompareRefs`; no similarity checked where §2 gives none (AC-1.6, AC-2.5, AC-3.3, AC-3.4); the AC-5.1 neutral extra uses ref `0xff`; no Landed note (no code deviation). No stop condition fired.

## Round 1 — review
Verdicts: CON changes needed · MNT changes needed · UX clean · SEC changes needed · PERF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | A new relation needs `relationOrder`, `outcomeOf` and `contradicts`; skill names one, no test catches a miss | Fix now → F-2 | verified: `outcomeOf` falls to `OutcomeUnknown`; `TestStronger` checks a hand copy of the order |
| CON-1 | CON | major | No Landed note though the work differs from the plan | Fix now → F-4 | verified: plan has no `## Landed`; DoD requires it |
| SEC-1 | SEC | major | All-pairs compare materializes every pair; no cap on value count | Part 1 (stream to a running best) Fix now → F-1. Part 2 (registry `MaxValues` + new cause) Rejected | Part 2 contradicts D-3 ("No registry section") and the Out line "A `property` section … Not planned". Time is O(n·m) either way, as in today's `match.pairValues` (`core/match/pairs.go`); the values are one Subject's concluded values for one property, collapsed by the strict profile (S11-11), so a count bound belongs where values are gathered, not in layer 2. Re-rated: memory part minor, fixed with F-1 at no cost |
| CON-3 | CON | minor | Tied contradicting pairs offered positive-A first, not in A's input order (D-6) | Fix now → F-1 | verified by trace: `bestPair(compareAll(posA, negB), compareAll(negA, posB), …)`; `{¬male, male}` / `{male, ¬male}` reports B negative, D-6 says A |
| MNT-4 | MNT | nit | `bestPair` two lists plus optional predicate | Fix now → F-1 | removed by the F-1 single loop |
| CON-2 | CON | minor | AC-5.1 neutral term is a pool term, D-15 says `unknown` | Fix now → F-3 | verified: `neutralValue` sets `Neutral` on a pool value for every type |
| CON-4 | CON | minor | No test that a caller's `Type` is overwritten (D-4) | Fix now → F-3 | verified: every helper leaves `Type` empty |
| MNT-2 | MNT | minor | `weakestOf` compares positions by hand beside `Stronger` | Rejected | D-16: "`weakestOf` keeps its `slices.Index(relationOrder, …)` comparison unchanged"; same behavior |
| MNT-3 | MNT | minor | `valuecompare/date_test.go` builders twin `datevaluetest` | Follow-up | plan Out line "each moves when its file is next edited"; name `valuecompare/date_test.go` first in the PR follow-ups |
| CON-5 | CON | nit | STEPs 3–7 in one commit | Logged | goes in the Landed note (F-4) |

### Fix plan 1

**F-1 — One pass over the pairs in D-6 offer order, keeping running bests** (from CON-3, SEC-1 part 1, MNT-4)
Files: `core/propertycompare/propertycompare.go`, `core/propertycompare/propertycompare_test.go`. Skill: `add-match-rule`.
Failing test first: `TestCompare/"D-6 tie on contradicting pairs keeps A's input order"` — term, single, a `{¬male, male}`, b `{male, ¬male}` → conflict, cause `negative`, pair `male` / `male`, `aNeg: true`, `Same`, 1 (today B is reported negative).
Do: `split` keeps each side's non-neutral values in input order in one list (each `keyed` already holds `v.Negative`), instead of separate `pos` / `neg` lists. One nested loop offers A's values in input order and, for each, B's in input order; skip two negatives; compare the pair once with `valuecompare.Compare`; a pair with exactly one negative that `contradicts` is a candidate for the running best contradiction, a pair of two positives for the running best positive pair, each kept with `better` (first offered wins a tie). Then the existing outcome logic: a contradiction wins; else a side with no positive → unknown / `no_values`; else the positive best with `outcomeOf` and cardinality. Delete `compareAll` and `bestPair`; no `[]pair` slice is built. Count of positives per side is read from the split, not from the loop, so `{}` sides still give `no_values`.
Check: the new row and every existing `TestCompare` row and `TestInvariants` pass; `scripts/check-pr.sh` PASS.

**F-2 — Every relation has an order, an outcome and a contradiction rule, checked by tests** (from MNT-1)
Files: `core/valuecompare/valuecompare.go`, `core/valuecompare/valuecompare_test.go`, `core/propertycompare/relations_internal_test.go` *(new, package `propertycompare`)*, `.agents/skills/add-match-rule/SKILL.md`.
Failing test first: none fails today; the tests guard the next relation.
Do: add `func Relations() []Relation` to `valuecompare` returning `slices.Clone(relationOrder)` (doc: "every relation, strongest first"). `TestStronger` walks `Relations()` for its adjacent pairs instead of its own list, and checks `len(Relations()) == 7` against the `Relation` constants (list the seven constants once in the test so a new constant not added to `relationOrder` fails). `relations_internal_test.go` `TestEveryRelationHasAnOutcome`: a table of all seven relations → expected `outcomeOf` and `contradicts` (Same, Equivalent → agree, true; Contains → partial, true; Overlaps, Near → partial, false; Different → conflict, false; Unknown → unknown, false), and a check that every relation in `valuecompare.Relations()` has a row. Skill "Add a relation, cause or unit code": after "a new relation also in `relationOrder` at its strength" add ", its outcome in `propertycompare.outcomeOf` and whether a negative contradicts it in `contradicts` (`TestEveryRelationHasAnOutcome` fails until both are decided)".
Check: `go test ./core/valuecompare/... ./core/propertycompare/... ./core/matchregistry/...` with no tags passes (`TestLayerPackages` included: no literal 7 in non-test code).

**F-3 — Tests say what D-4 and D-15 say** (from CON-4, CON-2)
Files: `core/propertycompare/propertycompare_test.go`, `core/propertycompare/invariants_test.go`.
Do: add the `TestCompare` row "D-4 a caller's type is overwritten": text, single, a `{York}` built with `Type: propertymeta.ValueTypeName`, b `{york}` → agree, `relation`, `Same`, 1, both pair types `text`. In `invariants_test.go` `neutralValue`, for terms add `neutral(term("unknown"))` instead of a pool term.
Check: `go test ./core/propertycompare/... -v` passes.

**F-4 — Landed note** (from CON-1, CON-5)
Files: `docs/deployment-plan/spike-11/prs/S11-09-property-compare.md`.
Do: add a `## Landed` section in the shape of `S11-05-migrate-matching.md`'s, listing: STEPs 3–7 in one commit `b72f822b`, with the `propertycompare` guard row there; AC-2.6 in `TestCompareRefs`; no similarity checked where §2 gives none (AC-1.6, AC-2.5, AC-3.3, AC-3.4); the AC-5.1 neutral extra uses ref `0xff`; the pairs are found in one pass with running bests (no pair slice), in D-6 offer order; `valuecompare.Relations()` and `TestEveryRelationHasAnOutcome` added in review.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-11/prs/S11-09-property-compare.md` reports no errors.

## Round 1 — fix
Head `e5f8f3e4`. Gate PASS (orchestrator run). F-1 done (row failed first: B reported negative), commit `33563294`; F-2 done, `6737ab28`; F-3 done, `7d29cd06` (D-4 row rode in `33563294`); F-4 done, `ad7ac4e8`; `completed.md` sentence, `e5f8f3e4`.

## Round 2 — review
Lenses: CON, MNT, SEC (had Fix-now findings). Scope `eddb1901...e5f8f3e4`.
Verdicts: CON clean · MNT clean · SEC clean. Round 1 Fix-now findings all resolved (CON-1, CON-2, CON-3, CON-4, MNT-1, MNT-4, SEC-1 part 1). No new finding. MNT logged, not a finding: `valuecompare.Relations()` has only test callers today.

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-3 (move `core/valuecompare/date_test.go` builders onto `datevaluetest`) · rejected for the user to veto: SEC-1 part 2 (value-count cap; D-3), MNT-2 (`weakestOf` onto `Stronger`; D-16) · to user: —
