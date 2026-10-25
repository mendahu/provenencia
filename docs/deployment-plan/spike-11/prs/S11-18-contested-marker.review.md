# Review — S11-18 Contested marker on Promote rows (and brief S11-D1)

**Verdict:** Fits — Ready (after round 2 fixes)
**Rounds:** 2 (round 2 after the board was agreed 2026-10-20). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None. The board was agreed 2026-10-20; the plan was refreshed against it (D-8, D-10).

## Writer's own calls
- D-1: the readout carries `contested` (marker only where the band shows), with a `false` default so existing literals compile.
- D-2: a retarget clears `contested` and `move` until the re-proposal.
- D-3: `.danger` tone and `.warning` icon for the badge (kit `PVBadge`).
- D-5: copy for the three move sentences and the fallback "after a connected row changed".
- D-6: `Target.handleRef` and `Row.shownRef` in the flow (from M-3).
- D-7: `Row.captionReason`; `reasonText(for:)` deleted (from M-4).
- Brief: one host per connection by type (participation/location under the event, place relationship under the part, directed relationship under the end named last, symmetric under the first); three subrows then "+N more"; already-filed connections stay hidden.
- D-10: the Updated caption takes `PVColor.infoForeground`, because the agreed board draws it in `--info-fg`.
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
**Answer:** To user, then resolved: the board was agreed 2026-10-20 and the plan refreshed against it.

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

## Round 2

Board refresh checked against `Promote flow.dc.html` frames 14, 15 and 06: copy, tone, icon, VoiceOver and the caption rule match CM-1 – CM-6. One visual change, the caption's info ink, became D-10. The board's differences for S11-27 are in the brief's §8.

Maintainability: all round-1 answers hold; D-10 adds no token or mechanism. Verdict: Fits.

### X-8 · blocker · STEP-3 deletes `reasonText(for:)` while the view still calls it
**Answer:** Fixed. The deletion and the test call-site edit moved to STEP-4, beside the caption change; STEP-3 says to leave the forwarder.

### X-9 · major · T-4 and T-6 fixtures leave the setup open
**Answer:** Fixed. Both name the `reasonsAreWordedWithLabelsAndRefs` setup; T-6 is one load with a contested Ada and a clean Bea.

### X-10 · minor · xcstrings placement has no rule
**Answer:** Fixed. STEP-2 names the entries to insert after.

### X-11 · minor · badge accessibility identifier may not be exposed
**Answer:** Fixed. D-3 drops the identifier; nothing reads it.

