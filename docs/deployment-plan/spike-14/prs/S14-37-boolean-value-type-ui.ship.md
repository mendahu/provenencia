# Ship log — S14-37 Boolean value type (UI)

**Plan:** [`S14-37-boolean-value-type-ui.md`](S14-37-boolean-value-type-ui.md) · **Branch:** `s14-37-boolean-value-type-ui`, stacked on `s14-39-properties-single-only` (#737, user 2026-11-07) · **PR:** #739
**State:** fixing 1

## Round 0 — implement
Head `210181ab`. Gate PASS (run by the orchestrator on the pushed head, `--base s14-39-properties-single-only`). Report: STEP-3 is one commit; the D-4 rename reached more sites than the plan's regex; a private `segment(_:)` helper in `PVSegmentedControl`. All are in the plan's Landed note. Manual check not run: Tab landing on Yes and the nil-selection drawing, both stop conditions, are untested.

## Round 1 — review
Base `39d8ea4c`, head `210181ab`. Reviewers: CON, MNT, UX (`macos/App/**`, `Localizable.xcstrings`), SEC+PRF (`api/ffi/**`; `core/database/**`, `macos/App/Features/**` models). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 (CON-1) | UX, CON | blocker | Yes and No cannot be reached from the keyboard at default macOS settings; the plan's focus stop condition was never run | Fix now → F-1 | verified: `PVChip` is a plain `Button` with no `.focusable`; `PVSelect.swift:164` and the composer's line row declare it; repo record S13-10 D-9 says kit chips are Tab stops only with Full Keyboard Access. D-10's mechanism alone cannot meet AC-5.2. AC-5.2 stands as written; the board note says Tab lands on Yes, then No |
| MNT-1 | MNT | blocker | One saved line reads two states; `markSaved` ignores the written row | To user → Fix now → F-2 | verified: `groups` takes `valueText` from `saved` and the rest from `values`; `saveEdit` replaces values before the commit, and `discardEdit` only closes. The user chose one source (2026-11-07), superseding D-14's "read as today" clause |
| CON-2 | CON | major | No test proves a changed answer reaches the store | Fix now → F-3 | verified: T-5 covers only an unchanged edit |
| MNT-2 (CON-7) | MNT, CON | major | D-8 leaves the rest of the dead draft-row path | Fix now → F-4 | verified: `addDraft` has no caller; every `rows` assignment goes through `loaded(from:)`. Widens D-8's list; T-6's `isEmptyDraft` line and T-8's draft-row case go with it |
| CON-4 | CON | minor | T-2 is not table-driven and its "row unchanged" is weak | Fix now → F-5 | verified at `observations_test.go:1632` |
| MNT-4 | MNT | minor | The control takes the handle but drops its Tab neighbours | Fix now → F-1 | verified: only `focus.apply` is read |
| MNT-5 | MNT | minor | `ObservationValueDisplay` serves three features from one feature's folder | Fix now → F-6 | verified: its siblings are in `macos/App/Platform/` |
| MNT-6 (CON-3) | MNT, CON | minor | Two docs miss the never-negative rule | Fix now → F-7 | verified: `docs/seeded-vocabulary.md:351`, `docs/conclusion-reconciliation.md:144` |
| MNT-3 | MNT | minor | `ObservationValueDisplay` keeps a silent fallthrough | Rejected | D-11's Rejected column declines "reading `propertyValueType` first": the value field says the type |
| PRF-1 | PRF | minor | Saved lines are reworded on each draft keystroke | Follow-up | no worse than base: the deleted `dateSummary` used the same `DateValueDisplay`, and the term arm got cheaper; the file that would change, `CitationComposerSubjectGroups.swift`, is not in this PR |
| CON-6 | CON | nit | `completed.md` misquotes the removed copy ("date unset"; it read "No date set") | Fix now → F-7 | stale text this PR added |
| MNT-7 | MNT | nit | `CitationObservationValue`'s header still says "Display" | Fix now → F-7 | stale comment in a changed file |

Nits (logged): `CON-5 · CitationObservationRows.swift:37,42 · the subjectID and propertyID setters have no caller` · `MNT-8 · CitationComposerVocabulary.swift:186 · isDepiction through valueType(propertyID:)` · `MNT-9 · CitationComposerVocabulary.swift:26 · vocabulary.fields could be snapshot` · `MNT-10 · CitationComposerAuthoring.swift:342 · mono compares raw strings` · `MNT-11 · CitationComposerModelTests.swift:972-978 · share one lines helper` · `UX-2 · CitationComposerAuthoringBox.swift:426 · "Value" may be read twice; check in manual check 3` · `UX-3 · PropertiesView.swift:875 · the picker segment speaks the raw key "boolean"; D-10 of S14-39 kept keys`

### Fix plan 1

Plan Decisions still bind, except F-2, which the user approved over D-14's clause.

**F-1 — Yes and No take keyboard focus at default settings** (from UX-1, CON-1, MNT-4)
Files: `macos/App/DesignSystem/Components/Chip/PVSegmentedControl.swift`, `macos/App/DesignSystem/README.md`; `macos/ProvenenciaTests/PVSegmentedControlTests.swift`. Skill: `add-ui-component`.
Failing test first: in `PVSegmentedControlTests`, which hosts the control in a window, `focusHandleLandsOnTheFirstSegment` hosts a control with `focus:`, writes the host's `FocusState`, and asserts the window's first responder moved off the window. If a hosted window cannot show this, say so in your report and leave it to manual check 3; do not weaken the change.
Do: when `focus` is non-nil, each enabled segment is `.focusable()` and `.onKeyPress(.space)` runs its `select`, the shape of the composer's line row (`CitationComposerSubjectGroups.swift:201-205`), so Tab reaches Yes, then No, as the board says. `focus.apply` stays on the first enabled segment. A control with no handle behaves exactly as today: the five existing call sites do not change. State in the control's doc comment and the README's SegmentedControl row that it does not read the handle's `next` / `previous` neighbours and leaves Tab between segments to the key loop.
Check: the test passes, or the report says why it cannot; `scripts/check-pr.sh --base s14-39-properties-single-only` PASS.

**F-2 — A saved line reads the stored Observation only** (from MNT-1, user 2026-11-07)
Files: `macos/App/Features/CitationComposer/CitationObservationRows.swift`, `CitationComposerAuthoring.swift`, `CitationComposerModel+Authoring.swift`; `macos/ProvenenciaTests/CitationComposerModelTests.swift`.
Failing test first: `failedEditThenDiscardKeepsTheStoredLine` — a saved Occupation, an edit that ticks Negative, `updateObservation` failing, then `discardEdit` → the line is not negative and an unchanged `valueText`; reopening the edit shows the stored polarity.
Do: in `groups`, read subject, Property, polarity and image from `saved` (`saved.subjectID`, `saved.propertyID`, `saved.polarity`, the depiction's `imageRelPath`). `markSaved` takes the written `CatalogObservation` and stores it in `saved`; `commit` passes the row `updateObservation` returns. `persistedID` and `persistedRef` become computed from `saved`; `savedLineCount` follows. `discardEdit` reverts the row before closing, through `observationRows.revert`.
Check: the test passes; `rg -n "row\.(subjectID|propertyID|isNegated)|values\.valueFields\.depictionDraft" macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` finds nothing in `groups`.

**F-3 — A changed answer reaches the store** (from CON-2)
Files: `macos/ProvenenciaTests/CitationComposerModelTests.swift`.
Test (it passes today; do not report it Not reproduced): `editingABooleanSavesTheNewAnswerPositive` — seed `canReadObservation(value: true)` with polarity negative, edit `valueBoolean` to false, `saveEdit` → the stored row has `valueBoolean == false` and polarity positive.
Check: the test passes, and fails with `valueBoolean: draft.valueBoolean` removed from `asCatalogObservation` (try it, then restore it).

**F-4 — The rest of the dead draft-row path goes** (from MNT-2, CON-7)
Files: `macos/App/Features/CitationComposer/CitationObservationRows.swift`; `macos/ProvenenciaTests/CitationComposerModelTests.swift`.
Do: delete `addDraft`, `focusedID`, `hasEditedOrSaving`, `unsavedObservationCount`, `isEmptyDraft`, the add branch of `performCommit`, the nil-baseline branch of `revert`, and the tests of those paths: T-6's `isEmptyDraft` line, T-8's draft-row case, and the tests at about `CitationComposerModelTests.swift:981` and `:1086`. `revert` keeps its edited-row branch (F-2 calls it). Do F-2 first.
Check: `rg -n "addDraft|focusedID|hasEditedOrSaving|unsavedObservationCount|isEmptyDraft" macos` finds nothing.

**F-5 — T-2 is a table with a real "unchanged"** (from CON-4)
Files: `core/database/observations/observations_test.go`.
Do: make `TestNegationFollowsTheValueType` table-driven; the refused update sends `false` over a stored `true`, and asserts the stored `value_boolean` is still 1 and the polarity still negative.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/` passes.

**F-6 — `ObservationValueDisplay` moves to Platform** (from MNT-5)
Files: `macos/App/Features/EvidenceGraph/ObservationValueDisplay.swift` → `macos/App/Platform/ObservationValueDisplay.swift`.
Do: `git mv`; the header says it words a saved Observation on every list surface. No `.pbxproj` entry if the groups are file-system synchronized; add one only if the build needs it.
Check: builds.

**F-7 — Docs and stale text** (from MNT-6, CON-3, CON-6, MNT-7)
Files: `docs/seeded-vocabulary.md`, `docs/conclusion-reconciliation.md`, `docs/deployment-plan/spike-14/completed.md`, `macos/App/Features/CitationComposer/CitationObservationValue.swift`.
Do: `seeded-vocabulary.md:351` drops "(from S14-37)" and adds that a boolean Observation is never negative, in the words of `docs/interpretation-layer-data-model.md:632`. `conclusion-reconciliation.md:144` adds that only a row edited by hand can be a negative boolean. `completed.md` quotes the removed copy as "No date set". The `CitationObservationValue` header drops "Display".
Check: `rg -n "S14-37" docs/*.md` finds nothing.

**F-8 — The plan's Landed note** (records F-1, F-2, F-4)
Files: `docs/deployment-plan/spike-14/prs/S14-37-boolean-value-type-ui.md`, and `completed.md` if a fix changed what shipped.
Do this last. Add to the `**Landed:**` line: segments take `.focusable()` and Space when given a handle (F-1); a saved line reads only `saved`, superseding D-14's clause (user, 2026-11-07); `markSaved` stores the written row and `discardEdit` reverts (F-2); the widened deletions (F-4); `ObservationValueDisplay` lives in Platform (F-6). Do not edit Decisions, Scope, or ACs.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-37-boolean-value-type-ui.md` reports no errors; `scripts/check-pr.sh --base s14-39-properties-single-only` PASS.
