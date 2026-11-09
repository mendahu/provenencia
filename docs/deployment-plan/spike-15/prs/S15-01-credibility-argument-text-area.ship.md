# Ship log — S15-01 The credibility argument is a text area

**Plan:** [`S15-01-credibility-argument-text-area.md`](S15-01-credibility-argument-text-area.md) · **Branch:** `s15-01-credibility-argument-text-area` · **PR:** #758
**State:** ready

## Round 0 — implement
Head `d77bd4b6`. Gate PASS (orchestrator run). Report: version bump is 0.4.3, not 0.4.2 (`main` was already at 0.4.2; STEP-5 says next patch). Manual § 10 checks not run (app not launched).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`. SEC+PRF: `macos/App/Features/**` model (`SourceCredibilitySection`) and a free-text argument now multi-line with no length bound.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | minor | An unsaved argument is lost when the researcher leaves the Source page | Follow-up | Verified: `.onAppear { argumentLocal = model.credibility.argumentDraft }` and `argumentDraft` is set only on Save or `resetDrafts()`. Pre-existing (lines unchanged by this PR; plan § 4 "Local draft … Unchanged"). The fix is a page-wide product call covering description, notes and argument together, which the plan's Out line excludes. Home to be chosen by the user |

CON: no findings. MNT: no findings. SEC+PRF: no findings (the unbounded argument length was considered and left out: same exposure as the description `PVTextArea`; plan Out line "A Source credibility argument length limit — not planned").

Nits (logged): `MNT-1 · SourceCredibilitySection.swift:38 · var isDirty now read only by tests; D-5 kept it` · `MNT-2 · SourcePageCredibilityView.swift:56-69 · isDirty(argument:) evaluated three times; a let would read better` · `MNT-3 · DesignSystem/README.md:159 · TextArea row could mention accessibilityLabel` · `UX-2 · SourcePageCredibilityView.swift:80-84 · typing during an in-flight save can be overwritten when the saved text had end whitespace (pre-existing)`

No Fix-now findings: no Fix plan 1, no round 2.

## Result
Gate PASS · CI pending · open: none · follow-ups: UX-1 · to user: —
