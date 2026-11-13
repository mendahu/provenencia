# Ship log — S15-34 Artifacts are edited in a sheet

**Plan:** [`S15-34-artifact-edit-sheet.md`](S15-34-artifact-edit-sheet.md) · **Branch:** `s15-34-artifact-edit-sheet` · **PR:** #819
**State:** ready

## Round 0 — implement
Base `main` @ `745575882` (plan written against `b9a881b`). Head `548ddd7f6`. Gate PASS (orchestrator rerun). Report deviations (also in the plan's Landed note): `ArtifactSheetModel.onSaved` takes the reply rows as optional, nil for a text-only Save (D-4 had it non-optional; matches D-7); the sheet's `init` sets graph choices and starts `linkLoadTask` when the graph list is loaded (D-11 had `openSheet` do it); `MetadataBoardModel` clears a field's typed flag in `applySet` and `cancelEdit`; STEP-3+4 one commit, STEPs 5–9 one commit; comment rewording commit `a8d53800f`. §10 manual check not run (no GUI): Escape/Return with a row in edit, the add row's `.onExitCommand` inside the sheet, and the scrolling body's height (D-2, D-10, D-1 stop-condition risks) are unexercised. One chained docs-commit+gate+push command was denied by the permission classifier; the implementer re-ran the pieces separately.

## Round 1 — review
Reviewers: CON, MNT, UX, SEC+PRF. UX routed by `macos/App/**` and `Localizable.xcstrings`; SEC+PRF by `macos/App/Features/**` models (`ArtifactSheetModel`, `MetadataBoardModel`, `SourcePageModel`). Skipped: none.

SEC+PRF: no findings. Outside-lens note (scroll target fires while the body is 0pt tall) deduped into CON-1 → F-3.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1, MNT-1 | UX, MNT | major | Add row opens unfocused; its Escape is focus-bound `.onExitCommand` while it stands the sheet's Escape/Return/outside click down; ✓/✕ hand-copy `PVInlineEditActions(.iconRow)` | Fix now → F-1 | verified: `MetadataBoardView.swift` `MetadataBoardAddRowView` has no focus and `.onExitCommand`; `PVInlineEdit.swift:310-330` `iconRowBody` is the same shape with a `.cancelAction` ✕. Contradicted D-10's rejected alternative → **user, 2026-11-13: reuse the kit ✓/✕, drop `.onExitCommand`, focus the picker on open**; suggestion row: **raise the signal only while focused with text** (supersedes D-2's `hasTypedText`) |
| UX-2 | UX | major | Save errors appear at the top of a scrolled body | Fix now → F-2 | verified: `ArtifactSheet.swift:52` Callout is the body's first child; `formBody` scrolls only on appear. Naming the field in the copy: Rejected — D-6's wording is the user's call; scrolling to the Callout fixes visibility |
| CON-1 | CON | major | §10 manual check not run; it is the only proof for AC-1.1, 2.4, 2.5, 4.2, 5.1, 5.6 | To user (manual check before merge) + F-3 | the scroll-on-appear risk (`bodyHeight` starts 0, `.onChange(of: row, initial: true)`) gets a code fix in F-3; the rest is the user's §10 check at finish |
| MNT-2 | MNT | major | `ArtifactSheetModel.visibleGraphs` is a third copy of the page's graph read | Fix now → F-4 (same-pass) | verified: `SourceArtifactsSection.swift:60`, `SourceGraphsSection.swift:25`, `ArtifactSheetModel.swift:100` |
| MNT-3 | MNT | major | Unlink-lock read copied into the sheet model beside `SourceGraphFormModel`'s | Fix now → F-5 (same-pass; supersedes the structural-debt row this PR added deferring it to S15-11) | verified: `ArtifactSheetModel.swift:193`, `SourceGraphFormModel.swift:122`, same call and state shape |
| MNT-4 | MNT | major | Kit README and `add-ui-component` don't teach `pvSheetInnerEdit` | Fix now → F-6 | verified: only README component rows mention it |
| UX-3 | UX | minor | "1 more copy details" has no plural | Fix now → F-7 | xcstrings entry has no `variations.plural` |
| UX-4 | UX | minor | Truncated card values have no tooltip | Fix now → F-7 | `MetadataValueText` plain branch has no `.help` |
| MNT-5 | MNT | minor | Sheet inlines the "no graph" line | Fix now → F-7 | same font/colour as `noGraphLine(identifier:)` |
| MNT-6 | MNT, CON outside | minor | Bare `do { }` in the card meta row | Fix now → F-7 | `SourcePageArtifactsView.swift:232` |
| CON-2 | CON | minor | `sheetSaveIsOneWriteWithTheDraft` skips AC-1.2's suggestion fill and add row | Fix now → F-7 | test body only saves provider and reorders |
| CON-3 | CON | minor | `metadataDialogs` comment names the removed dialog | Fix now → F-7 | `SourcePageModel.swift:18` |

Nits (logged): `UX-5 · L10n.swift saveArtifact/artifactLabelRequired comments` · `UX-6 · ArtifactCopyDetailsView.swift:41 hard-coded ", " in VO label` · `UX-7 · metadata intro reachable only by hover/VO hint (board decision)` · `MNT-7 · MetadataDraftHost @ObservationIgnored on computed confirmsClear` · `MNT-8 · confirm id from local prefix` (moot after F-1) · `MNT-9 · MetadataBoardDialogs name` · `MNT-10 · artifactCardSaveError* key names` · `CON-4 · SourcePageModelTests.swift:892 comment` · `CON-5 · PlaceRegistry.swift:107 comment` · `CON-6 · Landed note: record sheetSavePatchesWithoutReload rename`

### Fix plan 1

**F-1 — The add row and a typed suggestion take Escape like a row in edit** (from UX-1, MNT-1; user 2026-11-13)
Files: `macos/App/Features/Sources/MetadataBoardView.swift`, tests reading `.addRow.confirm`/`.addRow.cancel` ids. Skill: `add-ui-component`.
Do: in `MetadataBoardAddRowView` replace `confirmControl` and the ✕ `PVIconButton` with `PVInlineEditActions(isSaving: board.isSavingAddRow, saveLabel: L10n.Sources.metadataAddRowConfirmLabel, cancelLabel: L10n.Sources.metadataAddRowCancelLabel, saveDisabled: !board.canConfirmAddRow, style: .iconRow, accessibilityIdentifierPrefix: prefix, onSave: { Task { await board.confirmAddRow() } }, onCancel: { board.closeAddRow() })` (match its real init); drop the row's `.onExitCommand`; keep `.pvSheetInnerEdit(true)`. Give the `PVComboBox` a `PVFieldFocus` from a `@FocusState` and claim it on appear, as `PromoteView.swift:530-536` does. In `MetadataBoardSuggestionRow`, add a `@FocusState` on its `PVInput` and raise `.pvSheetInnerEdit(isFocused && !draft.isEmpty)`; keep its `.onExitCommand { draft = "" }`. Ids become `.addRow.save` / `.addRow.cancel`; update any reader. In the plan's Landed note: D-10's rejected ".cancelAction ✕" and D-2's suggestion-row signal superseded by the user's 2026-11-13 call in review.
Check: `rg -n "onExitCommand" macos/App/Features/Sources/MetadataBoardView.swift` prints only the suggestion row's line; `rg -n "addRow.confirm" macos` prints nothing; scripts/check-pr.sh PASS.

**F-2 — A Save failure scrolls the sheet's body to its message** (from UX-2)
Files: `macos/App/Features/Sources/ArtifactSheet.swift`, `ArtifactSheetModel.swift`, and `PVFormDialog.swift` only if the scroll target must be passed as a request. 
Failing test first: `sheetSaveErrorAsksToScrollToTop` — after `save()` stops on unconfirmed input (and on empty label), the model's scroll request token has changed and its target is the top section.
Do: on setting `saveError` or `labelError` in `save()`, bump a request token and set the scroll target to the top section; drive the existing `pvScrollsRowIntoView(_, request:)` path (no new `ScrollViewReader`, D-1). Keep D-6's copy unchanged.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — The scrolling body applies its scroll target after its first measured height** (from CON-1, SEC+PRF outside note)
Files: `macos/App/DesignSystem/Components/FormDialog/PVFormDialog.swift`.
Do: in `formBody`, run the on-open scroll once `bodyHeight` first becomes > 0 (for example a `request` token derived from `bodyHeight > 0`, through `pvScrollsRowIntoView`'s existing request input) instead of relying on appear alone. Do not change behaviour for non-scrolling dialogs.
Check: scripts/check-pr.sh PASS; the `#Preview` with a scroll target still opens at its target (manual).

**F-4 — One read of the Source page's graphs** (from MNT-2, same-pass)
Files: `SourcePageContext.swift`, `ArtifactSheetModel.swift`, `SourceArtifactsSection.swift`, `SourceGraphsSection.swift`.
Do: add `evidenceGraphs: [CatalogEvidenceGraph]?` (and `evidenceGraphsLoading: Bool`) on `SourcePageContext` built from `session.evidenceGraphsRead(sourceID:)`; make the sheet model's `visibleGraphs`/`awaitingGraphs`, `SourceArtifactsSection.visibleGraphs` and `SourceGraphsSection.rows` read it; delete the copies.
Check: `rg -n "evidenceGraphsRead" macos/App/Features/Sources` prints only `SourcePageContext.swift`, `SourcesListRow.swift:161`, `SourcesModel.swift:258` and `SourceGraphsSection.swift`'s `.failed` read (if kept); scripts/check-pr.sh PASS.

**F-5 — One unlink-lock read for the sheet and the graph form** (from MNT-3, same-pass; supersedes the `structural-debt.md` row deferring it to S15-11)
Files: `ArtifactSheetModel.swift`, `SourceGraphFormModel.swift`, new small type beside `SourceArtifactGraphLinks.swift`, `docs/ideas/structural-debt.md`.
Do: extract one `@Observable` value (e.g. `UnlinkLocks`) holding `blocks`, `state: SourceFormLoadState`, `load(store:pairs:onError:)` and `locked(by:)` over `lockHints(by:)`. Each caller keeps its current error mapping by passing it in (sheet: `context.fail`; graph form: `L10n.Errors.message(for:)`) — no visible behaviour change. Both models hold it; delete the copies and the structural-debt row; note in the plan's Landed note.
Check: `rg -n "getEvidenceGraphUnlinkImpact" macos/App/Features/Sources` prints only the new type; the existing sheet and graph-form lock tests pass unchanged; scripts/check-pr.sh PASS.

**F-6 — The kit teaches the inner-edit signal** (from MNT-4)
Files: `macos/App/DesignSystem/README.md` (the "Every sheet presents through `.pvSheet`" paragraph), `.agents/skills/add-ui-component/SKILL.md` (sheet paragraph).
Do: give `PVSheetDismiss.action`'s new signature and the inner-edit branch (an open inner edit stands down the sheet's Escape, Return and outside click); one line in the skill: "A control that owns an open edit inside a sheet raises `.pvSheetInnerEdit(isOpen)` and takes Escape with a `.cancelAction` Cancel (`PVInlineEditActions` does both); `PVInlineEdit` already does."
Check: `scripts/check-guidance-names.py` and check-pr PASS.

**F-7 — Minors riding along** (from UX-3, UX-4, MNT-5, MNT-6, CON-2, CON-3)
- UX-3: plural variations for `sources.page.artifactCopyDetailsMoreLabel` (one/other) per `add-localized-string`.
- UX-4: in `MetadataValueText`, when `lineLimit != nil`, `.help(entry.valueText)` on the plain branch, and the address in the URL branch's help.
- MNT-5: make `noGraphLine` a file-level view in `SourceArtifactGraphLinks.swift`; use it from the Add dialog and `ArtifactSheet`.
- MNT-6: drop the bare `do { }` in `SourcePageArtifactsView.swift` `metaRow`.
- CON-2: extend `sheetSaveIsOneWriteWithTheDraft` per AC-1.2: a suggested `image` filled through `board.save`, `film` added through `openAddRow`/`pickAddRowField`/`setAddRowValue`/`confirmAddRow`, URL dragged above Provider; expect one write with field ids `["url", "provider", "accessed", "image", "film"]` (or the census fixture's actual saved set, in that order).
- CON-3: `SourcePageModel.swift:18` → "The page's one clear confirm, for every board on it."
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `3dd3fdf1c`. Gate PASS (orchestrator rerun). F-1 – F-7 done. Deviations: F-2 – F-5 and the MNT-5/MNT-6/CON-2 parts of F-7 share one commit (`286f50de8`; shared files would not compile split); F-3 drops `scrollsOnAppear` for a `BodyScrollRequest(measured: bodyHeight > 0, caller:)` token; CON-2's expected order is `["url","provider","image","film"]` (the census fixture has no `accessed`). F-1's user call recorded in the plan's Landed note. `#Preview` scroll check not run.

Minors read by the orchestrator: UX-3 **Resolved** (`%#@details@` with one/other variations in xcstrings and L10n default) · UX-4 **Resolved** (`.help` on both `MetadataValueText` branches when `lineLimit != nil`) · MNT-5 **Resolved** (`SourceArtifactNoGraphLine`, used by the sheet and the Add dialog) · MNT-6 **Resolved** (no bare `do {` in `SourcePageArtifactsView.swift`) · CON-3 **Resolved** (`SourcePageModel.swift:18`) · CON-2 → the CON round-2 reviewer reads it with its scope.

## Round 2 — review
Reviewers: CON (CON-1 → F-3), MNT (MNT-1 – MNT-4), UX (UX-1, UX-2). Scope `548ddd7f6...3dd3fdf1c`. Not sent: SEC+PRF (no findings in round 1).

Round-1 items: UX-1/MNT-1 **Resolved** · UX-2 **Resolved** · MNT-2 **Resolved** · MNT-3 **Resolved** · CON-1 **Resolved in code** (F-3; §10 manual check stays with the user) · CON-2 **Resolved** (CON reviewer) · MNT-4 **Not resolved** → CON-7.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-7, MNT-4 | CON, MNT | blocker | New skill and README lines say `PVInlineEditActions` raises `.pvSheetInnerEdit`; it does not | Fix now → F-8 | verified: `rg pvSheetInnerEdit macos/App/DesignSystem/Components/InlineEdit/` → only `PVInlineEdit.swift:82`; the wrong wording came from the orchestrator's F-6 text |

Nits (logged): `CON-8 · MetadataBoardView.swift:388 empty .help when lineLimit == nil`
UX: clean. Outside-lens note (MNT): first scroll waits for `bodyHeight > 0` then one main-actor turn — covered by the §10 manual check.

### Fix plan 2

**F-8 — The kit guidance says who raises the inner-edit signal** (from CON-7, MNT-4)
Files: `.agents/skills/add-ui-component/SKILL.md:92`, `macos/App/DesignSystem/README.md:247-249`.
Do: reword both: a control that owns an open edit inside a sheet raises `.pvSheetInnerEdit(isOpen)` and takes Escape with a `.cancelAction` Cancel. `PVInlineEdit` does both. `PVInlineEditActions` gives only the Cancel, so a control built on it raises the signal itself (as `MetadataBoardAddRowView` does). An input whose only inner edit is typed text raises the signal while focused with text and clears on `.onExitCommand` (the suggestion row). No code change.
Check: `rg -n "PVInlineEditActions\` do both|PVInlineEditActions do both|and \`PVInlineEditActions\` do both" .agents macos/App/DesignSystem/README.md` prints nothing; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `ba4595a7d`. F-8 done (docs only); orchestrator read both lines: CON-7/MNT-4 **Resolved** (`PVInlineEdit` does both; `PVInlineEditActions` gives only the Cancel; the typed-text case named).

## Finish
Merged `origin/main` @ `b95dbb1a7` (docs only: S16-19 plan). Gate PASS (orchestrator, against `b95dbb1a7`).

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: §10 manual check (CON-1; the only evidence for AC-1.1, AC-2.4, AC-2.5, AC-4.2, AC-5.1, AC-5.6 and the §11 runtime stop conditions)

## User change (2026-11-13, after ready)
The user asked that the Artifact sheet's "Copy details" heading read "Artifact metadata", to match the Source's Metadata section. Sentence case, as the sheet's other labels ("Evidence graphs").

**F-9 — The sheet's copy-details heading reads "Artifact metadata"** (user, 2026-11-13)
Files: `macos/App/Platform/L10n.swift` (`artifactCopyDetailsLabel`), `macos/App/Resources/Localizable.xcstrings`, the plan's Landed note. Skill: `add-localized-string`.
Do: change the default value of `sources.page.artifactCopyDetailsLabel` to "Artifact metadata" in L10n and the xcstrings en value; comment "Label of the Artifact metadata section of the Artifact sheet". Keep the key, the section id `.copyDetails`, the accessibility ids, and the card's "+n more" / "more copy details of" copy unchanged. Add one Landed-note bullet: the sheet heading reads "Artifact metadata" (user, 2026-11-13), superseding AC-1.1's "Copy details" label.
Check: `rg -n '"Copy details"' macos/App` prints nothing; scripts/check-pr.sh PASS.
