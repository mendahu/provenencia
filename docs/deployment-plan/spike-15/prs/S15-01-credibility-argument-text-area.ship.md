# Ship log — S15-01 The credibility argument is a text area

**Plan:** [`S15-01-credibility-argument-text-area.md`](S15-01-credibility-argument-text-area.md) · **Branch:** `s15-01-credibility-argument-text-area` · **PR:** #758
**State:** ready

## Round 0 — implement
Head `d77bd4b6`. Gate PASS (orchestrator run). Report: version bump is 0.4.3, not 0.4.2 (`main` was already at 0.4.2; STEP-5 says next patch). Manual § 10 checks not run (app not launched).

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**`. SEC+PRF: `macos/App/Features/**` model (`SourceCredibilitySection`) and a free-text argument now multi-line with no length bound.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | minor | An unsaved argument is lost when the researcher leaves the Source page | Follow-up → dropped by the user (2026-11-09) | Verified: `.onAppear { argumentLocal = model.credibility.argumentDraft }` and `argumentDraft` is set only on Save or `resetDrafts()`. Pre-existing (lines unchanged by this PR; plan § 4 "Local draft … Unchanged"). The fix is a page-wide product call covering description, notes and argument together, which the plan's Out line excludes. Home to be chosen by the user |

CON: no findings. MNT: no findings. SEC+PRF: no findings (the unbounded argument length was considered and left out: same exposure as the description `PVTextArea`; plan Out line "A Source credibility argument length limit — not planned").

Nits (logged): `MNT-1 · SourceCredibilitySection.swift:38 · var isDirty now read only by tests; D-5 kept it` · `MNT-2 · SourcePageCredibilityView.swift:56-69 · isDirty(argument:) evaluated three times; a let would read better` · `MNT-3 · DesignSystem/README.md:159 · TextArea row could mention accessibilityLabel` · `UX-2 · SourcePageCredibilityView.swift:80-84 · typing during an in-flight save can be overwritten when the saved text had end whitespace (pre-existing)`

No Fix-now findings: no Fix plan 1, no round 2.


## User report — crash (after round 1, CI green)
The user ran the branch in Xcode: `EXC_BAD_ACCESS (code=2)` at a stack address in `PVTextAreaTabRouter.responds(to:)` (`PVTextArea.swift:368`), so a stack overflow from endless recursion.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| USR-1 | CON | blocker | Two text-area Tab routers wrap each other on one text view and `responds(to:)` recurses forever | Fix now → F-1, F-2 | Verified by reading: `SourcePageView.overviewColumns` stacks `SourcePageDescriptionView` (a `PVTextArea` while editing) and `SourcePageCredibilityView` (now a `PVTextArea`) in one `VStack`. `PVTextEditorInsetView.nearestTextView()` climbs up to 4 ancestors and takes the *first* descendant text view, so one area's clearer can reach the other area's text view. `PVTextAreaTabRouter.attach(to:)` then wraps a router: B.wrapped = A, and A's recheck (`scheduleTabRouterRecheck`) sees delegate B ≠ self and sets A.wrapped = B. `responds(to:)` → `wrapped.responds(to:)` loops. Pre-existing in the kit, reached for the first time by this PR's second text area in one column. All four reviewers missed it; no test covered two areas sharing an ancestor |

### Fix plan 1
**F-1 — A text view is wrapped by at most one Tab router, so no delegate cycle** (from USR-1)
Files: `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/ProvenenciaTests/PVTextAreaTests.swift`. Skill: `add-swift-test`.
Failing test first: `PVTextAreaTests.twoRoutersOnOneTextViewDoNotLoop`. Using `RouterFixture`'s text view and original delegate: router A attaches, router B attaches, then A attaches again (the recheck path). Then `textView.delegate!.responds(to: #selector(NSTextDelegate.textDidChange(_:)))` returns true without recursing, Tab still routes, and an edit still reaches the original delegate exactly once. Today this crashes with the same stack overflow, which is the red run.
Do: in `PVTextAreaTabRouter.attach(to:)`, never store another `PVTextAreaTabRouter` as `wrapped`. When the text view's delegate is already a different router, leave the text view alone and return. A text view has one router, the first that claimed it. Keep the existing `=== self` early return. Do not add a recursion depth guard; the invariant "wrapped is never a router" is the fix.
Check: the test passes; the existing `PVTextAreaTests` still pass.

**F-2 — Each area finds its own text view, not a neighbour's** (from USR-1)
Files: `macos/App/DesignSystem/Components/TextArea/PVTextArea.swift`, `macos/ProvenenciaTests/PVTextAreaTests.swift`.
Failing test first: `PVTextAreaTests.textViewSearchPicksTheTextViewUnderTheArea`. Build a container `NSView` holding two text views, each in an `NSScrollView` and stacked vertically with no overlap, plus a marker view whose frame covers the second one. The search from the marker returns the second text view. Today it returns the first.
Do: replace "first descendant within 4 hops" in `PVTextEditorInsetView.nearestTextView()` with a search that, at each ancestor, returns the non-field-editor descendant text view whose enclosing scroll view frame (or its own frame when it has none), converted to window coordinates, contains the clearer's frame midpoint in window coordinates. Put the geometric match in an `NSView` extension beside `firstDescendantTextView()`, for example `descendantTextView(containing point: NSPoint)`, reusing its field-editor skip so the test calls it directly. Keep the hop limit as a bound only, and rewrite the comment so it no longer claims the hop limit protects a second area. If `firstDescendantTextView()` is then unused outside tests, fold it into the new search and move `textViewSearchSkipsAFieldEditor` onto the new function. Do not leave two searches.
Check: the test passes; `scripts/check-pr.sh` PASS; the plan's Landed note records the kit fix (two text areas in one column could crash).

## Round 1 — fix
Head `0f792ee2`. Gate PASS (orchestrator run). F-1 done (red: the test host crashed with the stack overflow); F-2 done (red: the search returned the first text view). `firstDescendantTextView()` removed; `rg` finds no reference.

## Round 2 — review
Reviewers: CON (fresh). USR-1 was the only Fix-now blocker; the other lenses had none. Scope `6550b09b...0f792ee2`.

USR-1: **Resolved.** F-1 makes `wrapped` never a router, so no cycle can form; F-2 matches by the clearer's midpoint in window coordinates. Edge cases checked: zero-size frames before layout return nil and recover on `layout()`/`updateNSView`; other `PVTextArea` callers do not regress. No new blockers.

Outside lens (UX, unverified): an area may draw one frame with the default `TextEditor` inset before the first layout match. Logged; it is part of the manual check.

## Result
Gate PASS · CI see PR · open: none · follow-ups: none (UX-1 dropped by the user) · to user: — · the app run after the crash fix is with the user
