# Review — S11-18 Contested marker on Promote rows (and brief S11-D1)

**Verdict:** Fits with changes — Not ready (blocked on board S11-D1); every other finding fixed
**Rounds:** 1; round 2 runs with the board refresh, since round 1 had blockers and the board may change D-3 – D-5. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- **Agree the S11-D1 board.** The plan stays `Blocked — on board S11-D1` until the board is drawn and agreed. Recommendation: draw frames 1–2 (S11-18's) first; on agreement, refresh the plan against the board (D-8), record its link on **Board**, and run a round-2 review if the board changes a Decision.

## Writer's own calls
- D-1: the readout carries `contested` (marker only where the band shows), with a `false` default so existing literals compile.
- D-2: a retarget clears `contested` and `move` until the re-proposal.
- D-3: `.danger` tone and `.warning` icon for the badge (kit `PVBadge`).
- D-5: copy for the three move sentences and the fallback "after a connected row changed".
- D-6: `Target.handleRef` and `Row.shownRef` in the flow (from M-3).
- D-7: `Row.captionReason`; `reasonText(for:)` deleted (from M-4).
- Brief: one host per connection by type (participation/location under the event, place relationship under the part, directed relationship under the end named last, symmetric under the first); three subrows then "+N more"; already-filed connections stay hidden.
- Out: no per-alternative contested mark; no move shown outside the Updated caption.

## Round 1

### M-1 · blocker · D-4 and T-6 call file-private `PromoteAssessmentCopy`
**Answer:** Fixed. D-4 adds `L10n.Promote.assessment(_:)` beside `outcome(_:)`; its four view call sites move onto it and the enum is deleted (STEP-2, STEP-4, Codebase fit, Deletes).

### M-2 · major · `evidenceForContested` duplicates the sentence
**Answer:** Fixed. One key `promote.row.assessmentContested` ("%@, contested") and `PromoteModel.assessmentText(_:)`; the label keeps `evidenceFor`. S11-20's scope now notes the sheet subtitle uses `assessmentText`.

### M-3 · major · private `neighborNaming` keeps the shown-ref rule in the model
**Answer:** Fixed. D-6 is `Target.handleRef` and `Row.shownRef` in `PromoteFlow`; T-5 covers the fallback to the Source ref. `noteDuplicates` / `targetColumn` logged in Next adds, not touched.

### M-4 · minor · forwarding wrapper `updatedReasonText(for:)`
**Answer:** Fixed. D-7 is `Row.captionReason`; `reasonText(for:)` is deleted; T-5 is a flow test.

### M-5 · minor · brief CS-3 highlight has no inventory row
**Answer:** Fixed. Brief inventory row "Subrow highlight": snowflake, `PVColor.surfaceSelected`, no new token or modifier.

### M-6 · minor · sheet names neighbors by Source ref
**Answer:** Fixed (logged). Noted on the S11-20 scope entry in `deployment-plan.md`.

### X-1 · blocker · plan is Blocked with no board
**Answer:** To user (above). Expected for a plan whose brief was written in this pass.

### X-2 · blocker · T-3 retarget token matches nothing
**Answer:** Fixed. T-3 retargets to the helper's alternative `e2` and merges a row on `e2`.

### X-3 · major · T-1 / T-2 fixtures underspecified
**Answer:** Fixed. Both spell each row as `row(...)` plus field assignments, including `target` for the taken case and the alternative's `assessment`.

### X-4 · major · `completed.md` has no path
**Answer:** Fixed. STEP-5 names `docs/deployment-plan/spike-11/completed.md` and the section shape.

### X-5 · minor · xcstrings placement and check
**Answer:** Fixed. STEP-2 inserts in key order and runs `scripts/check-localizable-xcstrings.py`.

### X-6 · minor · small open choices
**Answer:** Fixed. D-4 states the badge gets no accessibility modifier; `assessmentText` sits beside `reasonText`; D-6 keeps `reasonVia`; the §3.1 bullet was checked verbatim (`docs/promote-graph-alignment.md:62`); a stop condition covers XT not running.

### X-7 · blocker (brief) · CM-1 claims None band
**Answer:** Fixed. CM-1 now names Strong, Medium, Weak; None only on a researcher-picked record (a decided handle can be rescored below Weak, so the readout does show None there).

### X-8 · major (brief) · frames leave choices open
**Answer:** Fixed. Frames 1–2 are marked S11-18's, frame 1 adds a Medium row, sample names and the 1200 pt page width; badge column stays fixed.

### X-9 · minor (brief) · smaller details
**Answer:** Declined for S11-18: marks per connection type, highlight timing and the two-non-row-ends case are S11-27's frames, and S11-27's plan refreshes against the board. The ritual's "delete this board's reference" is the verbatim Paste block from `add-design-brief`.
