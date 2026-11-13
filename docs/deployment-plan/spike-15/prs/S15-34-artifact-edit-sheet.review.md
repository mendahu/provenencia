# Review — S15-34 Artifacts are edited in a sheet

**Verdict:** Fits with changes — Ready with changes, 2 rounds; every blocker Fixed; two user answers (Cancel, Save with unconfirmed text); M-8 Logged.
**Rounds:** 2. **Reviewers:** maintainability (fresh, read-only), executability (fresh, read-only)

## To user

- **X-3 · Cancel.** SA-8 says Cancel asks to discard; the kit's visible Cancel closes at once on every dialog. **Answer (user, 2026-11-13):** follow the kit; Escape and an outside click ask. AC-2.2 and D-8 say so; S15-D14's Board line records it.
- **X-11 · Save with unconfirmed text.** **Answer (user, 2026-11-13):** stop and say so — no write, a top Callout asks to press ✓ or ✕ (AC-2.6, D-6).

## Writer's own calls

- One inner edit at a time on a board: opening a row edit closes the add row (dropping its content, as ✕ would), and opening the add row ends a row edit (D-10).
- The add row closes on Escape through `.onExitCommand`, not a `.cancelAction` shortcut (D-10, X-13).
- The sheet's tests live in `SourcePageModelTests.swift` under their own MARK, reaching the model through `model.artifacts.openSheet`, so they reuse the file's private census fixtures (X-6).
- The field-name style in `ArtifactCopyDetailsView` stays its own caption style rather than widening the private `MetadataBoardLabel` (M-5's smaller point): the label is a board-row column with a fixed 132pt width and edit-alignment padding.

## Round 1

### M-1 · blocker · A second ScrollViewReader beside `pvScrollsRowIntoView`
**Answer:** Fixed — D-1 uses `.pvScrollsRowIntoView(scrollTarget, anchor: .top, animated: false, scrollsOnAppear: true)`.

### M-2 · blocker · `PVSheetKeys` beside `PVSheetDismiss`
**Answer:** Fixed — D-2 adds `innerEditOpen` to `PVSheetDismiss.action` (Escape and outside click) and `PVFormDialogControls.confirmShortcut(innerEditOpen:)`; T-1 extends the existing test structs.

### M-3 · blocker · `applySaved` beside `show`
**Answer:** Fixed — D-7 extends `show(_:metadata:effects:)` (replace or append, rows, one `setWorkspace`, `.wroteMetadataBoard`); `patchArtifacts` folds into it; the plan says which write the page reads and that the mutation is idempotent.

### M-4 · major · The components that own the edit should raise the signal
**Answer:** Fixed — `PVInlineEdit`, the add row and a suggestion row with typed text raise `pvSheetInnerEdit` (D-2); the sheet body reports nothing.

### M-5 · major · A named copy-details view
**Answer:** Fixed — `ArtifactCopyDetailsView` with `static shown(_:)` (D-16, T-8). The field-name style point: Declined, see Writer's own calls.

### M-6 · major · `add-workspace-place` line 173
**Answer:** Fixed — STEP-11 updates it, with an `rg` check.

### M-7 · major · A new PreferenceKey for the body height
**Answer:** Fixed — `onGeometryChange` into state (the `PVPageColumns` idiom); Codebase fit names the capped-ScrollView precedents and why a sheet body differs.

### M-8 · minor · Two unlink-lock loaders
**Answer:** Logged — a row in `docs/ideas/structural-debt.md` (STEP-11), to touch with S15-11.

### M-9 · minor · Dirty composed by hand
**Answer:** Fixed — `MetadataBoardHost.isChanged` and `MetadataBoardModel.hasUnsavedChanges` (D-5).

### M-10 · minor · Dialog leftovers
**Answer:** Fixed — `MetadataBoardAddResult` moves to `MetadataBoardModel.swift`; three comments rewritten (D-14).

### M-11 · nit · Duplicate title key
**Answer:** Fixed — the title reuses `editArtifactLabel(ref:)`.

### M-12 · nit · T-1's home
**Answer:** Fixed — existing `PVSheetDismissTests` and `PVFormDialogControlsTests`.

### Outside-lens notes
**Answer:** Fixed — the Scale sentence counts the `contains` comparisons; D-9 names the sheet board's `onClear`.

### X-1 · blocker · The draft board is not observable
**Answer:** Fixed — D-18 makes `MetadataDraftHost` `@Observable`; T-9.

### X-2 · blocker · Unconfirmed text is invisible to the model
**Answer:** Fixed — D-5: the local editors report when "differs from the starting text" flips (`setTyped`), never per keystroke; a suggestion row with typed text is an inner edit (D-2); Save stops on it (D-6, user).

### X-3 · blocker · Cancel versus SA-8
**Answer:** To user → answered (follow the kit).

### X-4 · major · AC-1.4 unreachable
**Answer:** Fixed — `canSave` leaves the label out; `save()` checks it (D-6).

### X-5 · major · Init and lock-read ownership
**Answer:** Fixed — D-4 names the init with `onSaved`, `linkLoadTask`; D-11 says `openSheet` starts it.

### X-6 · major · STEP-5 tests need a way in
**Answer:** Fixed — `openSheet` moves into STEP-5; the tests live in `SourcePageModelTests.swift` (Writer's own calls).

### X-7 · major · Graph list still loading
**Answer:** Fixed — `awaitingGraphs` and the loading branch (D-4, D-9); `graphIDs` set when the list first loads.

### X-8 · major · Tests with no fate
**Answer:** Fixed — § 8 lists every one as ported (with its new name), updated, kept or removed.

### X-9 · major · Two `rg` checks fail by construction
**Answer:** Fixed — STEP-7 is scoped to the Sources feature and qualifies `seedDrafts`; STEP-9 is scoped and names the expected history hits.

### X-10 · major · STEP-3 needs `SourcePageMetadataView`
**Answer:** Fixed — STEP-3 lists it; D-14 says which `acceptsArtifactPageDrop` line goes in which step.

### X-11 · minor · Save with open input
**Answer:** To user → answered (stop and say so).

### X-12 · minor · Skill line 173
**Answer:** Fixed (as M-6).

### X-13 · minor · `.cancelAction` in a feature view
**Answer:** Fixed — `.onExitCommand` on the add row; one inner edit at a time (D-10).

### X-14 · minor · Small choices
**Answer:** Fixed — (1) D-8's `verbatimTitle` init stores the title too; (2) a `ProgressView` in ✓'s place, `PVComboBox(size: .sm)`; (3) `MetadataValueText(chrome:)`; (4) `show` takes the rows; (5) `sheet != nil` or the id missing; (6) the Callout is not compact; (7) the dismiss rule covers an outside click; (8) `pickAddRowField` / `setAddRowValue` in the model.

### X-15 · nit · Lint warning
**Answer:** Logged — "Move to another Source…" is the menu's real label.

## Round 2

Fresh reviewers checked every round-1 answer against the code at `bc4dec3`; all hold. Five new blockers came from round-1 fix text; each is answered below. Round 2 is the cap: nothing is left open.

### M2-1 · blocker · Two typed-text signals (`typedFieldIDs` beside `hasTypedEdit`)
**Answer:** Fixed — D-5 deletes `hasTypedEdit`; `showRefusal` reads `typedFieldIDs`; ids are cleared where their buffer is (leaving edit, a fill, an Escape that clears a suggestion, `dismissSuggestion`, `clear`, `resetDrafts` for rows gone). T-6's typed-elsewhere case moves to T-2 as a direct `showRefusal` test, since D-6 stops Save first.

### M2-2 · blocker · The rows written into the cached key twice
**Answer:** Fixed — D-7: `setWorkspace` carries the Artifact list (no slot patches it); the rows go to the page's workspace only, as `SourcePageMetadataHost` does, and into the cache only through `.wroteMetadataBoard`'s slot. The `add-workspace-place` § 6 line S15-36 wrote stays true, so no rule changes and nothing goes To user.

### X2-1 · blocker · Escape still taken by the sheet's hidden button
**Answer:** Fixed — D-2: the hidden button's shortcut is `innerEditOpen ? nil : .cancelAction` (the `PVConfirm` precedent); `innerEditOpen` reaches `PVSheetCloseRule.outcome` for the outside click only; a content-initiated `pvSheetClose` follows the dirty rule as today. T-1 extends `sheetCloseOutcomes`.

### X2-2 · blocker · Two tests break STEP-7
**Answer:** Fixed — `restingCardMarksAnArtifactInNoGraph` keeps its section half and its card half is ported to `sheetMarksAnArtifactInNoGraph` (T-5); `pageDropStagesFileWithoutCreating` is updated for the sheet (T-7).

### X2-3 · blocker · A typed suggestion has no way out, and ✕ can lock Save
**Answer:** Fixed — `dismissSuggestion` and `clear` drop the id (D-5, T-2 `dismissingATypedSuggestionClearsItsFlag`); Escape in a suggestion with typed text clears it (D-2, AC-2.4); the Save-stop copy is reworded for every field kind, "Finish the open field first: save its value or clear it." (D-6, D-17, AC-2.6). Writer's call: the wording change keeps the user's decision (stop and say so) and fits a suggestion row's "Save" and ✕ controls.
