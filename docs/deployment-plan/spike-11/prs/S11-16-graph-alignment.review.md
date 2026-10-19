# Review — S11-16 Rewrite graph alignment

**Verdict:** Fits with changes — Not ready (round 1); every finding answered below. The plan stays **Blocked — on S11-15**, so round 2 runs with the refresh after S11-15 lands, against its Landed note.
**Rounds:** 1. **Reviewers:** maintainability (opus), executability (sonnet)

## To user
- None. The candidate gate (similarity, 0.5) was asked during planning and answered 2026-10-19.

## Writer's own calls
- `ConflictWithFixed` is kept (wire field unchanged) and measured against the settle margin (1) instead of `ConflictPenalty` (3); contested is a separate flag (D-8).
- An unplaced neighbor lines up at neighbor evidence ≥ 2 (`segment.UnplacedAlike`, new estimate) and differs below 0 (D-3).
- Move reasons are three new `Reason` codes in a separate `Row.Move`; `Row.Reason` keeps its meaning (D-9).
- Bridge properties are not compared: no golden needs them (D-15).
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
