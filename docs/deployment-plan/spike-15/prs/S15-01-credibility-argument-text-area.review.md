# Review — S15-01 The credibility argument is a text area

**Plan:** [`S15-01-credibility-argument-text-area.md`](S15-01-credibility-argument-text-area.md)
**Lint:** `python3 scripts/check-plan-docs.py` — 0 errors before round 1 and after fixes.
**Verdict:** Maintainability **Fits with changes**; Executability **Ready with changes**. One round; both blockers Fixed; no round 2 needed (no reshape). Status → Planned.

## Round 1 — Maintainability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| M-1 | blocker | D-4 named the field with a call-site `.accessibilityLabel`, a second naming mechanism beside `PVFieldLabel.resolve(own:field:)`; `PVSelect` already takes an own label | **Fixed.** D-4 now adds `accessibilityLabel:` to `PVTextArea` and `own:` to `pvFieldNamed`; STEP-1 added; README Field row updated; Codebase fit rows added. The optional prompt fallback went **To user** as R-1 |
| M-2 | blocker | The view's `private var isDirty` repeats `SourceCredibilitySection.isDirty`, which only tests read; the plan had moved removing it to an Out row | **Fixed.** D-5: `isDirty(argument:)` on the model, view copy deleted (STEP-2, Deletes); AC-2.6 and T-2 pin the inner-newline case; the Out row is gone |
| M-3 | minor | The trim test belongs beside the rule in `sourcecredibility`, not in the handler test | **Fixed.** T-1 and STEP-3 moved to `TestAssessments`; `add-ffi-handler` dropped from Skills to load |

## Round 1 — Executability

| ID | Severity | Finding | Answer |
| --- | --- | --- | --- |
| X-1 | major | Status Draft but the base check names a Planned commit | **Fixed.** Status Planned after this review; **Planned** line says the plan's own commits on top are docs only |
| X-2 | major | Four `MARKETING_VERSION` lines, plan said "Debug and Release" | **Fixed.** D-8 and STEP-5: all four, per bump `9509e881` |
| X-3 | major | Stop condition used the minimum width, ACs the default | **Fixed.** One width: the default window size; D-1 says the 700-pt minimum may scroll |
| X-4 | minor | "500-character" text is 508 characters | **Fixed.** "about 500 characters" |
| X-5 | minor | Docs step left status words, order and the `completed.md` form open | **Fixed.** STEP-6 names them and gives the paragraph |
| X-6 | minor | T-1 fixture choices | **Fixed.** STEP-3 names the helpers, the action and `grades[2]` |
| X-7 | minor | FakeStore stores as given, so T-2 proves the model trims | **Fixed.** Stated in § 4 FakeStore row |
| X-8 | minor | Tab reaching Save depends on Keyboard navigation; VoiceOver only by hand | **Fixed.** AC-2.5 and D-3: Tab leaves the field; AC-2.7 is a manual line and a Stop condition |

## To user

| ID | Question | Recommendation | Answer |
| --- | --- | --- | --- |
| R-1 | Should `PVTextArea` speak its `prompt` when it has no own or field label? It would name the notes composer and other prompt-only areas too | Not in this PR: it changes what VoiceOver says on pages this PR does not check. Recorded as an Out line, not planned | Planned with the recommendation; the user can overrule before the PR ships |
