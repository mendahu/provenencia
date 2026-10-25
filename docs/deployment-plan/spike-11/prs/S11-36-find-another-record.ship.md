# Ship log — S11-36 Find another record in the target menu

**Plan:** [`S11-36-find-another-record.md`](S11-36-find-another-record.md) · **Branch:** `s11-36-find-another-record` · **PR:** #558
**State:** ready

## Round 0 — implement
Head `a7d3f84f`. Gate PASS (orchestrator re-run against `181dd413`). PR #558 (draft). Report deviations: four commits, not one per step; T-7 uses a Weak proposal row with a None alternative (a None-band row starts unset); the general `PVComboBox` init takes `focus` before `onCommit`, matching the Text init. Not run: the plan's §10 manual checks 1–7, VoiceOver, and the runtime-only Stop conditions (sentinel shown after close, Escape reaching `.onExitCommand`, `WorkspaceSession` reaching popup rows through D-17, other `PVComboBox` callers, the omnibar's spoken hit).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none. UX: `macos/App/**` and `Localizable.xcstrings`. SEC+PRF: `macos/App/Features/**` models (`PromoteModel`, `CatalogSearchModel`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Every way out of the record search drops keyboard focus | Fix now → F-2 | verified: `PromoteRecordSearch` holds the only `@FocusState` in the row and unmounts on pick, Cancel and Escape; `PVSelect` (`PVSelect.swift:71`) has only a private `isFocused` and no focus parameter, so nothing hands focus back to the remounted `targetSelect`. D-8 covers focus on open only. Code this PR adds |
| UX-2 | UX | major | Stepping a closed target menu (↑ on an unset row, ↓ from Skip, End, type-select, VoiceOver adjust) commits the sentinel and opens the search | To user → Fix now → F-1 | verified: `PVSelectSession.handleClosedKey` → `moveCommitted` / `jumpCommitted` → `commit(index:close:false)`; `moveIndex(from: -1, delta: -1)` returns `count - 1`, the sentinel. Any fix that makes the sentinel unsteppable touches the working rule "The menu entry is a sentinel option" and the Out line "An action item … in `PVSelect`" |
| CON-1 | CON | minor | §10 manual check not run; view-only ACs have no automated proof | Rejected | §10 is the user's manual check at hand-off (ship-pr Finish); listed in the PR body and the report. The reviewer traced the Escape, click-away and pick paths and found them right |
| MNT-1 | MNT | minor | Found-record fixture written twice in `PromoteModelTests.swift` | Fix now → F-3 | verified: `rg -n '"e-far", ref: "PER-FAR"' macos/ProvenenciaTests` gives lines 421 and 448. Same-pass cleanup: this PR adds the second copy |
| MNT-2 | MNT | minor | Picker and omnibar each order the search list's states | Follow-up | `docs/ideas/shared-list-helpers.md` "One popup list under the combo box and the omnibar" already names it; the two lines differ (two-letter state only in the picker, loading only in the omnibar), so this is not a copy to fold now |
| MNT-3 (+UX out-of-lens) | MNT | minor | The `Empty == Text` init does not call the D-15 general init, so the parameter list is kept in three places | Fix now → F-4 | verified: `PVComboBox.swift:798-830` forwards to the private memberwise init; D-15 says it calls the general init. Minor in a file this PR changes |

Nits (logged): `MNT-4 · macos/App/Features/Promote/PromoteModel.swift:183 · move RecordSearch beside the other private(set) state` · `MNT-5 · macos/App/Features/Promote/PromoteView.swift:565 · pass the CatalogSearchModel in instead of re-unwrapping` · `MNT-6 · macos/App/Features/Promote/PromoteView.swift:534 · one label view for the New/Skip and Find option rows` · `UX-3 · macos/App/Features/Promote/PromoteView.swift:572 · emptyLine could take the query the box passes`

SEC+PRF: clean, no findings.

**UX-2, user (2026-10-25):** "Unsteppable option". `PVSelectOption` gets a flag for an option that can be chosen only from the open menu; stepping a closed select skips it. The sentinel stays an ordinary option (the working rule holds, its wording gains "chosen only from the open menu"). Decision UX-2 → Fix now → F-1.

### Fix plan 1

**F-1 — Stepping a closed select never commits a menu-only option** (from UX-2, user decision)
Files: `macos/App/DesignSystem/Components/Select/PVSelect.swift` (`PVSelectOption`), `macos/App/DesignSystem/Components/Select/PVSelectSession.swift`, `macos/App/Features/Promote/PromoteView.swift` (`targetOptions`), `macos/ProvenenciaTests/PVSelectSessionTests.swift`, `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`, `add-swift-test`.
Failing test first: `PVSelectSessionTests` — a session over `[a, b, find(menuOnly)]`: closed `.up` from no selection commits `b`; closed `.down` from `b` stays on `b`; closed `.end` commits `b`; `decrement()` / `increment()` behave the same; closed type-select of "f" commits nothing; with the menu open, `.down` to the menu-only row and `.return` commits it.
Do: add `var isMenuOnly: Bool = false` to `PVSelectOption` (documented: chosen only from the open menu; closed-key stepping, Home/End, closed type-select and VoiceOver adjust skip it). In `PVSelectSession`, closed `moveCommitted`, `jumpCommitted` and `applyTypeSelect(commit: true)` pick among options that are not menu-only (skip past them in the step direction; stay put when none is left). Open-menu highlight and commit are unchanged. Set `isMenuOnly: true` on the `findToken` option in `PromoteView.targetOptions`. One line in the DesignSystem README's select section. Update the working rule's wording in the plan's Landed note ("chosen only from the open menu").
Check: the new test passes; existing `PVSelectSessionTests` pass unchanged; scripts/check-pr.sh PASS.

**F-2 — Focus returns to the row's target menu when the record search closes** (from UX-1)
Files: `macos/App/DesignSystem/Components/Select/PVSelect.swift`, `macos/App/Features/Promote/PromoteView.swift` (`PromoteAlignmentRow`, `PromoteRecordSearch`), `macos/App/DesignSystem/README.md`. Skill: `add-ui-component`.
Failing test first: none (focus is view state; manual check 3 from the keyboard covers it). Say so in the report.
Do: give `PVSelect` an optional `focus:` parameter in the shape of `PVComboFocus` (`PVComboBox.swift:272`): a small `PVSelectFocus` built from `FocusState<Value?>.Binding` + `equals:` that the select applies with `.focused(binding, equals:)` instead of its private `isFocused` when given. Reuse `PVComboFocus` itself if its `install` shape can be generalized without changing combo callers; otherwise a sibling type beside it, documented as the same contract. Existing callers pass nothing and keep the private `@FocusState`. In `PromoteAlignmentRow`, keep `@FocusState private var targetFocus: Bool?`-style state (or an enum), pass it to `targetSelect`, and when `model.recordSearch?.subjectID` changes from this row's id to anything else, set it on the next runloop (`DispatchQueue.main.async`, the same idiom as D-8's focus-on-open). This covers pick, Cancel and Escape alike.
Check: scripts/check-pr.sh PASS; other `PVSelect` callers compile unchanged.

**F-3 — Point the older found-record test at `storeWithFoundRecords`** (from MNT-1)
Files: `macos/ProvenenciaTests/PromoteModelTests.swift`.
Failing test first: none (test refactor).
Do: `aFoundRecordIsProposedHeldAndFiled` starts from `storeWithFoundRecords(...)` and drops its own `e-far` membership block.
Check: `rg -n '"e-far", ref: "PER-FAR"' macos/ProvenenciaTests` prints exactly one line (inside `storeWithFoundRecords`); the test passes.

**F-4 — The `Empty == Text` combo-box init calls the D-15 general init** (from MNT-3, UX out-of-lens)
Files: `macos/App/DesignSystem/Components/ComboBox/PVComboBox.swift`.
Failing test first: none (no behavior change; existing `PVComboBoxTests` pin it).
Do: make the `Empty == Text` init (line ~800) call the general init from the `extension PVComboBox` (line ~763), passing `empty: { _ in Text(emptyLabel) }` (whatever the current empty text is), so the parameter list lives in the general init and the private memberwise init only. Keep the plain-row init delegating as it does.
Check: `PVComboBoxTests` pass; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `a5c7cb5c`. Gate PASS (orchestrator re-run against `181dd413`). F-1 and F-2 in `a5c7cb5c` (shared files), F-3 `04b1ed29`, F-4 `4595979a`. Report deviations: F-1's tests could not run red first (they need `isMenuOnly` to compile); F-2 adds `PVSelectFocus` as a sibling of `PVComboFocus` because the combo's `install` builds a TextField; F-3's grep still prints the `model.hold(… "e-far", ref: "PER-FAR" …)` call at line 424, which is not a fixture copy. The keyboard focus return was not run by hand.
- MNT-1 (F-3): **Resolved** — read the diff; `aFoundRecordIsProposedHeldAndFiled` starts from `storeWithFoundRecords` and the membership block is gone.
- MNT-3 (F-4): **Resolved** — read the diff; the `Empty == Text` init passes `focus:` before `onCommit:`, so it resolves to the general init rather than the memberwise one.

## Round 2 — review
Reviewers: UX (had UX-1 and UX-2 major). Scope `a7d3f84f...a5c7cb5c`. Not re-sent: CON, SEC+PRF (no findings to fix); MNT (minors only, read above).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1 | UX | major | Focus dropped on closing the search | Resolved | `PVSelectFocus`; `PromoteAlignmentRow.targetFocus` set when `recordSearch?.subjectID` leaves the row |
| UX-2 | UX | major | Closed stepping commits the sentinel | Resolved | `PVSelectOption.isMenuOnly`; `PVSelectSessionTests` cover ↑/↓/Home/End/adjust/type-select and the open menu |

No new findings. Watch item (reviewer, low confidence, not a finding): handing the search from Ada's row to another row (AC-3.2) queues two focus writes, Ada's `targetFocus` and the new row's `focus = .query`; the new row's should win. Manual check 3 from the keyboard confirms; narrowing the `onChange` to `new == nil` removes the race if not.

## Result
Gate PASS · CI green · open: none · follow-ups: MNT-2 (`docs/ideas/shared-list-helpers.md`, already named there); the Source page type editor has the same focus-on-close gap as UX-1 (can adopt `PVSelectFocus`/`PVComboFocus` restore) · to user: UX-2 (answered: unsteppable option) · CON miss: none
