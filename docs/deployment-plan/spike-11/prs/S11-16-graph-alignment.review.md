# Review — S11-16 Rewrite graph alignment

**Verdict:** Fits with changes — Ready with changes (round 2); every finding fixed. Round 1 ran on the draft against `219a3b8`; round 2 on the refresh against `8915340` (S11-15 merged).
**Rounds:** 2. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. The candidate gate (similarity, 0.5) was asked during planning and answered 2026-10-19.

## Writer's own calls
- `ConflictWithFixed` is kept (wire field unchanged) and measured against the settle margin (1) instead of `ConflictPenalty` (3); contested is a separate flag (D-8).
- An unplaced neighbor lines up at neighbor evidence ≥ 2 (`segment.UnplacedAlike`, new estimate) and differs below 0 (D-3).
- Move reasons are three new `Reason` codes in a separate `Row.Move`; `Row.Reason` keeps its meaning (D-9).
- Bridge properties are not compared: no golden needs them (D-15).
- A row whose only candidate is under the weak band comes back Skip, not New, keeping `promote-graph-alignment.md` principle 5 (D-8, AC-2.6).
- Ported tests may change expected values only where a named Decision causes it; the sparse, birth-vs-death, lifespan, obituary and part-of tests keep targets and bands, or the implementer stops (D-13).

## Round 1

### M-1 · blocker · Align builds a second shortlist index and runs `Candidates` again
**Answer:** Fixed. D-5 drops Align's index; `Canon.Candidates` (new field) carries the loader's per-Subject lists (D-10). Tests fill it with `withCandidates`. Also answers X-4.

### M-2 · blocker · Contested copies segment compare's conflict loop
**Answer:** Fixed. D-8 adds `subjectcompare.Result.Conflicts()`; `segmentcompare.Compare` and contested call it (STEP-1).

### M-3 · major · Registry sections copied into `Config` by hand in several places
**Answer:** Fixed. D-1 adds `Registry.AlignConfig()`; `promotealign.config` and the `config(R)` test helper are gone.

### M-4 · major · A second import walker beside `checkPure`
**Answer:** Fixed. D-11 extends `checkPure` with the two forbidden imports; no new test.

### M-5 · major · S11-19 would copy the placement rule
**Answer:** Fixed. D-6: `segments` takes a placement lookup and returns per-link results kept on `cand.links`; Next adds names it for S11-19.

### M-6 · minor · "Contested" with two meanings
**Answer:** Fixed. D-13 renames `TestContestedHandleIsDecidedTheSameEveryRun` → `TestSharedHandleIsDecidedTheSameEveryRun`.

### M-7 · minor · "(deleted)" suffix not used elsewhere
**Answer:** Fixed. D-14 uses the plain "carried from <identifier>" form.

### M-8 · minor · AC-1.2 asserted twice
**Answer:** Fixed. AC-1.2 maps to T-11 (c) only.

### M-9 · nit · Skill "Add a knob" names only valuecompare's rules file
**Answer:** Fixed in D-14.

### X-1 · blocker · Fixed Subjects get no candidates
**Answer:** Fixed. D-5 (1) scores a held handle as the first candidate; D-7 `place` nominates for fixed neighbors too; D-8 says where a fixed row's candidates come from; the final re-score includes fixed Subjects.

### X-2 · blocker · AC-4.4's grep matches registry comments
**Answer:** Fixed. AC-4.4 excludes `core/matchregistry/**`.

### X-3 · blocker · `len(idx) < 2` in `flagSharedHandles` fails the numbers guard
**Answer:** Fixed. D-11 names the line: `<= 1`.

### X-4 · major · Canon index ≠ loader's candidates
**Answer:** Fixed by M-1: one list, from the loader.

### X-5 · major · Settle rules leave choices open
**Answer:** Fixed. D-7: the first dirty mark wins until the round clears it; `releasedFrom` / `move` take the latest release; a round checks every dirty Subject before one `walk`. The round-cap case is AC-3.2 (T-7); the Risks row cites it.

### X-6 · major · Stale Current state; STEP-2 file list incomplete
**Answer:** Fixed. Counts are 40 and 30 (`walk` is listed twice because `Align` calls it twice). STEP-2 lists `index.go`, `export_test.go`, `exhibits_pin_test.go`, `layer.go`, `exhibits.go`, `promotealign_test.go`. The `loadCanon` signature is STEP-0's check.

### X-7 · major · Manual check 3 cannot be done
**Answer:** Fixed. T-13 adds `TestProposeMarksAConflictingRowContested`; the manual check names it.

### X-8 · minor · Small ambiguities
**Answer:** Fixed. D-8 says which Skip rows are contested; STEP-3 names the fallback call and the rule for a changed expectation; MNT-3 scope is every comment naming deleted `core/match` / `graphalign` identifiers; D-6 keeps the most-evidence unplaced result per link; AC-2.2's EVT-1 is a birth in 1817; STEP-7 gives the benchmark command.

## Round 2

All round-1 answers were checked against the plan and the tree; both reviewers found them landed.

### M-1 · blocker · `newPairSegments` decides "an unplaced neighbor lines up" with its own accept-bar gate
**Answer:** Fixed. D-3: `newPairSegments` compares a different neighbor with `Unplaced` and counts agree or partial; its `AcceptScore` gate goes. "What stays" and STEP-4 name it.

### M-2 · major · The old nomination and accept path is missing from Deletes and the grep
**Answer:** Fixed. Deletes lists `pushFromAnchor`, `enqueue`, `accept`, `scoredCand`, `sortCands`, `scoreCandidate`; `lostContest` stays; `queueItem` gains `version`; STEP-4's grep names them.

### M-3 · minor · `subject.go` comments end in a bare `ConflictPenalty`
**Answer:** Fixed in D-14 and STEP-8.

### M-4 · minor · A false `Result.Neighbor` doc and a copied Apart block
**Answer:** Fixed. D-3 adds an `apart` helper both readings call and rewrites the field doc.

### X-1 · blocker · Who sets `via` when a nominated handle is already a candidate
**Answer:** Fixed. D-7: a nomination sets `via` on a candidate without one and keeps an existing one; D-13 pins reasons and `Via` in the Promote tests.

### X-2 · major · A weak candidate turns New into Skip
**Answer:** Fixed. Kept as Skip (principle 5); said under "What changes", AC-2.6 and T-6.

### X-3 / X-4 · minor · §8 has no "stats" line
**Answer:** Fixed. The §8 edit and the Docs row mention are dropped (S11-15 already rewrote §8).

### X-5 · minor · `checkPure` change under-specified
**Answer:** Fixed. D-11 states the rule per hop and per target.

### X-6 · minor · Small gaps
**Answer:** Fixed. MNT-3 scope is a grep; `withCandidates` returns a `Canon`; a released Subject may re-place on the same handle (`Move` nil); STEP-2 says `alikeNew` / `newPairSegments` read `AcceptBar` until STEP-4.

## Refresh — STEP-3 stop (2026-10-19)

The implementer stopped at STEP-3. D-4's similarity gate admits a handle that agrees on one property with nothing else to compare (similarity 1, evidence 0.5). Under D-8 that row is Skip `weak`, which conflicts with D-13's pin on `TestBirthIsNeverProposedOnADeath/AC-1.8` and with three unpinned tests.

**Answer:** To user → option (a): keep D-4 and D-8; AC-1.8 expects Skip `weak` with an unbanded alternative. D-13 names the exception and the three tests' new expectations, written to keep each test's intent (`ViaNeighbor` empty, property-only score). §11's stop condition excepts AC-1.8. Coverage-aware similarity goes to [`docs/ideas/coverage-aware-similarity.md`](../../../ideas/coverage-aware-similarity.md).

A fresh reviewer checked the refresh against the branch: Ready with changes.
- **F-1 · major:** D-13's STEP-3 rule still forbade the changes. **Fixed:** the rule admits a handle that is at least half alike but under the weak band; the exception holds from STEP-3 on.
- **F-2 · major:** Promote never shortlists an undated event. **Fixed:** "What changes" and the idea lead with surname-only; AC-1.8 pins `graphalign`'s rule only.
- **F-3 · minor:** "may be an alternative" left a choice. **Fixed:** exact `Alternatives`, plus the real reason each score is property-only.
- **F-4 · minor:** stale test comments. **Fixed:** D-13 names the rewording.
- **F-5 · minor:** "unbanded" not asserted. **Fixed:** case renamed "offered as weak".
- **F-6 · minor:** T-4 now pins the surname-only case.

## Refresh — STEP-4 stop (2026-10-19)

The implementer stopped at STEP-4. D-7 set `via` on any existing candidate a placed neighbor points at. In `TestPartOfChainIsStrongWithoutADecision` the state is placed first and relabels the city's and country's own toponym candidates as `via`. D-13 pins that test's reasons (`agrees` / toponym).

**Answer:** To user → limit the clause to a fixed receiving row. An unfixed row's existing candidate keeps its own reason; a nominated (new) candidate still gets `via`; a fixed row still gets `via` on its candidates (D-8). The implementer ran this variant: `core/graphalign`, `core/database/promotealign` and `api/ffi/handlers` all pass. This narrows review X-1's answer to fixed rows.
