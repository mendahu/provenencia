# Ship log — S14-03 Rectangle-only depiction drawing

**Plan:** [`S14-03-rectangle-only-depiction-drawing.md`](S14-03-rectangle-only-depiction-drawing.md) · **Branch:** `s14-03-rectangle-only-depiction-drawing` · **PR:** #652
**State:** ready

## Round 0 — implement
Base `main` @ `baecdfb8` (plan planned against `f2dc2a95`; no Current-state row went stale). Head `7502c708`. Gate PASS on the orchestrator's third run: runs 1 and 2 FAILED `xcode-test` only in the known flake family in files this PR does not touch (run 1: `SourcePageModelTests.aGraphWriteMovesOnlyItsRow`, `SourcesModelTests.listReadsEachGraphsProgress`, `PVSegmentedControlTests.disabledItemDoesNotSelect`; run 2: `SourceGraphFormModelTests` edit cases). All were `waitUntil` timeouts. The implementer saw the same Sources pair once.
Report deviations: (1) AC-5.3 conflicts with D-2 as written: D-2 hides a tool armed before a dialog opens, so `calloutSavesTheCitationsPage` arms the tool with no dialog open, and only a tool armed on screen blocks the page save (recorded in the Landed note); (2) `drawingHoldsPasteTranscription` now loads a real text PDF through the new `ArtifactFileFixtures.writeTextPDF`; (3) one code commit instead of one per step; (4) manual checks 1 – 8 not run (no GUI session).

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Nothing skipped. UX was routed by `macos/App/**`, `L10n.swift` and `Localizable.xcstrings`; SEC+PRF by the `macos/App/Features/**` models and the cut path, which reads a Citation region from the catalog.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (+ PRF-1, SEC outside-lens note) | MNT | major | Retrying a failed cut depends on a Task spawned by the view's `useRegion` and on the dialog being presented again. A region button cuts twice, and a redraw onto an equal region never retries. | **To user** | Verified: `AuthoringBox.swift:908` `.task(id:)` plus `:948` `Task { renderDepictionPreview() }`. `depictionImage(for:)` keeps no record of a cut in flight. The fix replaces the mechanism D-9 prescribes. |
| UX-1 | UX | major | A failed "Save page N" shows its error behind the modal dialog, with no announcement. | **To user** (placement) / Fix now (announcement) | Verified: `savePage()` has no else branch, and `fields.error` is drawn only by the evidence column's `PVField`. D-11 put the error in `fields.error`, "shown where the composer shows it today". |
| UX-2 | UX | major | Focus is dropped when the Callout disappears after "Save page" succeeds. | Fix now → F-2 | Verified: `savePage()` makes no focus call. Draw is `.focused(focus, equals: .value)` at `:884`, and `FocusClaim.claim` is the existing pattern. |
| UX-3 | UX | minor | A failed cut appears with no announcement, and its caption ("Draw it again") loops when the file is missing. | Fix now (announce) / **To user** (copy, which the board fixed) | Verified: `useRegion` announces before the render. |
| MNT-2 | MNT | minor | "Cannot cut and not loading" is derived in two places (caption and preview). | Fix now → F-4 | Verified: `Depiction.swift:220` and `:283`. |
| MNT-3 | MNT | minor | Unused `wrongPageJSON` test fixture. | Fix now → F-5 | Verified: `rg` finds only the declaration at `:1045`. |

Nits (logged): `CON-1 · ship log untracked (deliberate; committed at finish)` · `MNT-4 · DepictionTests.swift:1333 · let (model, _) instead of _ = store` · `MNT-5 · ArtifactViewer.swift:10 · drop isSetPageEnabled default` · `MNT-6 · Depiction.swift:314 · comment why !isDrawingRegion is also checked` · `UX-4 · Depiction.swift:225 · .waiting caption falls through to .empty ("use a region button") while the buttons are disabled` · `UX-5 · L10n.swift:3347 · depictionDrawingHint "tools above" with one tool (kept by plan)`

**Paused for the user** on MNT-1, UX-1 placement and UX-3 copy.

**User, 2026-10-31: "Yes, all three"**, accepting the orchestrator's recommendations:
- **MNT-1** (with PRF-1): a pending-cut key drives the `.task(id:)`. This amends D-9's mechanism (F-1).
- **UX-1:** the failed page save shows in the needs-page Callout and is announced. This amends D-11's placement (F-3).
- **UX-3:** a new caption string for a failed cut that points at the file (F-6).

### Fix plan 1

**F-1 — One pending-cut value drives the preview cut** (from MNT-1, PRF-1, the SEC outside-lens note; user-approved amendment of D-9)
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerAuthoringBox.swift`, `CitationComposerDepictionTests.swift`.
Failing test first: `CitationComposerDepictionTests/failedCutRetriesOnAnEqualRegion`. After a cut fails (T-7's setup: PDF deleted after opening, then restored), commit the *same* region again. Assert that `depictionPendingCutKey` goes from nil (failed memo under the key) back to the key, and that `renderDepictionPreview()` then yields `.cut`. Also assert, by counting cuts (a test seam is fine only if one already exists; otherwise assert through the key), that a region-button commit leaves exactly one trigger: `useRegion` no longer calls `renderDepictionPreview`.
Do:
- Add `var depictionPendingCutKey: DepictionCutKey?` on the model. It is `depictionPreviewKey` when `depictionPreviewState == .cutting`, else nil. A failed memo under the key gives `.failed`, so the value is nil. When `commitViewerRegion` clears that memo, the value becomes the key again, so the `.task(id:)` restarts with no help from the view.
- In `CitationComposerAuthoringBox`, change `.task(id: model.depictionPreviewKey)` to `.task(id: model.depictionPendingCutKey)`, and delete the `Task { await model.renderDepictionPreview() }` in `useRegion`.
- Update the D-9 doc comments that mention the view's Task.
- Add one sentence to the plan's Landed note: "D-9 amended (user, 2026-10-31, ship log round 1): a pending-cut key drives the `.task(id:)`; `useRegion` no longer starts its own Task."
Check: the new test and T-7 pass; `rg -n "renderDepictionPreview" macos/App` prints only the model's definition and the one `.task(id:)` call; scripts/check-pr.sh PASS.

**F-2 — Focus lands on Draw after Save page succeeds** (from UX-2)
Files: `CitationComposerAuthoringBox.swift`.
Do: on success in `savePage()`, call `FocusClaim.claim(focus, .value)` with the caption announcement as its completion. Model it on `claimDrawingFocus` in `CitationComposerViewerPane.swift:71`.
Check: build; manual check 6 by keyboard (left for the user); scripts/check-pr.sh PASS.

**F-3 — A failed Save page shows in the Callout and is announced** (from UX-1; user-approved amendment of D-11)
Files: `CitationComposerAuthoringBox.swift`, `CitationComposerDepictionTests.swift` if a model-level assertion applies.
Do:
- In the needs-page `PVCallout` content, when `model.fields.error` is non-nil, show it as a line under the button: `Text(error)`, `PVColor.danger`, caption-size font, as `PVField`'s error line draws it. Compose it from existing views; add no new component. Check `PVCallout`'s content slot accepts it; if the slot takes only one view, wrap the button and the line in a `VStack`.
- In `savePage()`'s else branch, post `AccessibilityAnnouncement` with `model.fields.error` when it is set.
- Keep `fields.error` as the store (AC-5.2 unchanged).
- Add one sentence to the Landed note: "D-11 amended (user, 2026-10-31): a failed save shows in the needs-page Callout and is announced."
Check: T-9 still passes; scripts/check-pr.sh PASS.

**F-4 — One value for whether the viewer can cut** (from MNT-2)
Files: `CitationComposerModel+Depiction.swift`.
Failing test first: none (refactor); `captionFollowsTheDraft`, T-6 and T-7 must stay green.
Do:
- Add a derived `depictionCutAvailability` with cases `.ready`, `.loading` and `.unavailable`, built from `depictionCanCut` and `artifactViewer.isLoading`.
- The caption (`:220`) and `depictionPreviewState` (`:283`) both read it and no longer combine the two themselves.
- With it, fix UX-4: while `.loading`, the caption shows the same state the preview's `.waiting` implies. Do not show `.empty`, which tells the researcher to use buttons that are disabled. Reuse an existing caption case if one fits ("cannot cut" must not show while loading); if none fits and a new string would be needed, leave UX-4 alone and say so in the report.
Check: `rg -n "artifactViewer.isLoading" macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift` prints only the new derived value; scripts/check-pr.sh PASS.

**F-5 — Delete the unused fixture** (from MNT-3; the nits ride along)
Files: `CitationComposerDepictionTests.swift`, `ArtifactViewer.swift`, `CitationComposerModel+Depiction.swift`.
Do:
- Delete `wrongPageJSON` (`:1045`).
- MNT-4: `let (model, _) = …` instead of `_ = store` (`:1333`).
- MNT-5: drop the default on `isSetPageEnabled`.
- MNT-6: say in a doc comment why `canSaveDepictionCitationPage` also checks `!isDrawingRegion`: it covers a tool armed for the Citation.
Check: `rg -n wrongPageJSON macos` prints nothing; scripts/check-pr.sh PASS.

**F-6 — Announce a failed cut, and point the caption at the file** (from UX-3; the copy is user-approved)
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerAuthoringBox.swift`, `L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Do:
- Add `L10n.CitationComposer.depictionCaptionCutFailed`, key `citationComposer.depictionCaptionCutFailed`, with defaultValue "The region could not be cut from the file. Check that the file is still in the project, then draw it again." and comment "Caption under the depiction preview when the drawn region could not be cut from the open file".
- `DepictionCaptionState.cutFailed` reads the new string. `depictionRenderFailed` stays as the post-Save error wherever else it is used; if `.cutFailed` was its only reader, delete it and its catalog entry.
- Announce the caption when it becomes `.cutFailed`, with `.onChange(of: model.depictionCaptionState)` in the authoring box, posting only on the change to `.cutFailed`.
Check: `captionFollowsTheDraft` asserts the new text for `.cutFailed`; localizable check PASS; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `845774e0`. Gate PASS (orchestrator, first run). F-1 – F-6 done.
Report deviations:
1. One commit for the round.
2. F-1's failing test was red only because the new property did not exist; it never ran against the old behaviour.
3. `depictionRenderFailed` is kept, since `+Authoring.swift:446` and `:571` read it after Save.
4. UX-4 is left alone: no caption case fits a loading viewer without a new string.
5. The `completed.md` paragraph is updated.
6. No manual checks were run.

**PRF-1 (minor): Resolved.** The orchestrator read it: `rg renderDepictionPreview macos/App` prints only the definition and the one `.task(id: model.depictionPendingCutKey)`. `useRegion` no longer starts a Task.

## Round 2 — review
Reviewers: MNT (MNT-1 major), UX (UX-1 and UX-2 majors). Not sent: CON (clean) and SEC+PRF (minor only, resolved above).

Both came back clean: MNT-1, UX-1 and UX-2 are **Resolved**, and there are no new blockers.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-7 | MNT | minor (raised as a note) | F-3's Callout error line is one more hand-styled `Text` (`PVTypeScale.micro`, `PVColor.danger`), with no shared design-system piece behind it. | **To user** | This is a same-pass cleanup: F-3 added the copy. Others: `PVField.swift:66`, `DateValueEditorForm.swift:143,148`, `NameValueEditorForm.swift:80,228`, `SourcePageMetadataView.swift:204`, `SourcePageIdentityHeader.swift:232` and more. The rule makes it Fix now, but round 2 allows only blockers into Fix plan 2, so the two rules conflict. |
| MNT-7 → | | | **User, 2026-10-31: "Yes, extract it here".** | Fix now → F-7 (same-pass cleanup; no round 3: the orchestrator reads the commit and runs the gate) | |
| UX-6 | UX | minor (outside-lens note) | Load-failure paths also set `fields.error` (`CitationComposerModel.swift:618, 986, 1092`). One that fires while the needs-page Callout is up reads as a failed page save. | Follow-up | It needs a load error that coincides with the Callout. A separate save-error field would be a model change past what the round-2 cap allows. |

### Fix plan 2

**F-7 — One design-system error line, with every hand-styled copy moved onto it** (from MNT-7; same-pass cleanup, user-approved)
Files: new `macos/App/DesignSystem/Components/Field/PVErrorText.swift` (or the folder `add-ui-component` puts it in). Copies to move:
- `PVField.swift:65`
- `PVInlineEdit.swift:146`
- `CitationComposerAuthoringBox.swift:859`
- `DateValueEditorForm.swift:142, 147`
- `NameValueEditorForm.swift:78, 226`
- `SourcePageMetadataView.swift:202`
- `SourcePageIdentityHeader.swift:230`
- `EvidenceGraphView.swift:447`

Add the new file to the Xcode target. Skill: `add-ui-component` (design-system layer).
Failing test first: none. This is a visual refactor with no behaviour change, and the existing tests must stay green.
Do:
- Add `PVErrorText` to the design system. It takes the error text (a `String`, plus a verbatim init if the copies need one; follow how `Text(verbatim:)` sites pass it) and a size, `.micro` (default, the field-level line) or `.caption` (the form-level line).
- It draws `Text` with `PVFont.body(size:)` and `PVColor.danger`, exactly as the copies do, so nothing a researcher sees changes.
- Move every copy above onto it. Keep each call site's extra modifiers: `.padding`, `.fixedSize`, `.accessibilityIdentifier`.
- Do not touch icon-only danger uses: the `PVIcon` in `SourcesGraphCountLine` and `SourcesGraphButton`, and the required marker in `PVField.swift:53`.
- Add one line to the plan's Landed note: "MNT-7 (user, 2026-10-31): hand-styled error lines moved onto `PVErrorText`."
Check:
- `rg -n -U 'Text\((verbatim: )?[^)]*\)\s*\n\s*\.font\(PVFont\.body\(size: PVTypeScale\.(micro|caption)\)\)\s*\n\s*\.foregroundStyle\(PVColor\.danger\)' macos/App` prints only the line(s) inside `PVErrorText.swift`.
- `rg -n 'foregroundStyle\(PVColor\.danger\)' macos/App/Features` prints only `PVIcon` uses.
- scripts/check-pr.sh PASS.

## Round 2 — fix
Head `5c5b2331`. Gate PASS (orchestrator). F-7 done: there is a new `PVErrorText` (`.micro` / `.caption`) in `DesignSystem/Components/ErrorText`, and nine copies were moved onto it: `PVField`, `PVInlineEdit`, the authoring box, the date form (2), the name form (2), Source metadata, the Source identity header, and the evidence graph.
Report deviations:
1. The file was registered by hand in `project.pbxproj`.
2. Three `PVColor.danger` uses are left. They are not error lines: a canvas painter colour (`EvidenceCardView`), an icon (`SourcesGraphCountLine`), and a label that turns danger when invalid (`DateValueEditorForm:288`).

The orchestrator read the commit: the component matches the copies (`Text(verbatim:)` with a `String`, the same as `Text(String)`), and nothing a researcher sees changes.
**MNT-7: Resolved.**

## Result
Gate PASS (`5c5b2331`; `main` unmoved since `baecdfb8`, no merge) · CI pending · open: none · follow-ups: UX-6 (a load error in `fields.error` can show under the Save page button), UX-4 (the caption while the viewer loads needs a new string), UX-5 (`depictionDrawingHint` "tools above", for the board's next pass) · to user: answered (MNT-1, UX-1 placement, UX-3 copy, MNT-7) · manual checks 1 – 8 not run · CON miss: none · flakes seen outside this PR: `SourcePageModelTests.aGraphWriteMovesOnlyItsRow`, `SourcesModelTests.listReadsEachGraphsProgress`, `PVSegmentedControlTests.disabledItemDoesNotSelect`, `SourceGraphFormModelTests` edit cases
