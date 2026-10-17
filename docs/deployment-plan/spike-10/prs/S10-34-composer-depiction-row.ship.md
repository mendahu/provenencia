# Ship log — S10-34 Composer depiction row: draw a face, save it, see it

**Plan:** [`S10-34-composer-depiction-row.md`](S10-34-composer-depiction-row.md) · **Branch:** `s10-34-composer-depiction-row` · **PR:** #442
**State:** ready

## Round 0 — implement
Base `500396f`. Head `4c664a6`. Gate PASS (orchestrator run). Report deviations:
- Set page disabled while drawing by a model guard in `setPageFromViewer`, not a greyed button; Clear menu disabled in the view.
- `depictionWrongPage` computed, not stored.
- `ignoreNextDisarm` added so a refused wrong-page commit does not end drawing.
- `FakeStore.recordedDepictionPaths` test hook added.
- `ArtifactFileFixtures.writePNG` draws through a bitmap context (Retina-safe).
- `ArtifactRegionCrop.render` is `@MainActor`.
- STEP-10 brief path in the plan updated to the archive path.
- Code and tests in one commit (Xcode project file lists every new file).
- `setAuthoringCitation` resets `depictionDraft` only when the picked Citation differs (D-12 gives no guard).

## Round 1 — review
Security: clean. Performance: 1 major. Maintainability: 1 blocker, 1 major, 5 minor, 1 nit. UX: 3 major, 3 minor. Conformance: 1 major, 3 minor.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (+UX-4, CON-2) | MNT | blocker | Wrong-page refusal uses one-shot `ignoreNextDisarm`; overlay shows the refused rect disarmed; flag goes stale after a handle edit | Fix now → F-1 | verified: `commitAndDisarm` sets `armedTool = nil`, then `onCommit`, then `onDisarm`; handle edit at `mouseUp` calls `onCommit` alone, leaving the flag set. Fix keeps `ArtifactViewer`'s API unchanged (Builds on) instead of the reviewer's Bool-returning callback |
| CON-1 | CON | major | Picking another Citation then Keep editing erases the drawn region | Fix now → F-2 | carries a failing test; matches the implementer's D-12 deviation note |
| UX-1 | UX | major | Tab after picking Likeness lands on `.value`, which nothing binds | Fix now → F-3 | verified: `.depiction` arm renders `CitationDepictionValueArea` with no `.focused(_, equals: .value)` |
| UX-2 | UX | major | After next face, Property combo reads empty while the Likeness area shows | Fix now → F-4 | verified: `propertyOptions(forSubjectID:)` returns `[]` with no subject; `selectedLabel` falls back to "" |
| UX-3 | UX | major | VoiceOver reads a negative likeness (and Saved) the same as a positive one | Fix now → F-5 | verified: `depictionLineLabel` overrides the combined label with no polarity input |
| PRF-1 | PRF | major | Full PDF page rasterized on the main actor at preview (from `body`) and again at Save | Fix now (memo reuse at Save, PNG encode off main) → F-6; box-only render To user → **Follow-up** (user, 2026-10-17: measure on a 300 dpi scan first) | Box-only render contradicts D-5 (`thumbnail(of:)` over the visual size; hand-built CTM rejected). Reuse and off-main encode fit D-5/D-6/D-14 |
| MNT-2 | MNT | major | Two Out lines homed in `docs/dogfood/ux.md` never written | Fix now → F-7 | verified: `git diff --stat … -- docs/dogfood` empty |
| MNT-3 | MNT | minor | `commitViewerRegion` restates the page rule | Fix now (rides F-1) | same function |
| MNT-4 | MNT | minor | Two inline `valueType == .depiction` checks beside `isDepiction(propertyID:)` | Fix now (rides F-4) | few lines, files touched |
| MNT-5 | MNT | minor | `initialPropertyFocus` call built in two views | Fix now (rides F-3) | few lines, files touched |
| MNT-6 | MNT | minor | Three things named `depictionLineLabel`; branching modifier | Fix now (rides F-5) | same code |
| MNT-7 | MNT | minor | `FakeStore.depictionRecord` formats and records | Follow-up | file not otherwise touched |
| MNT-8 | MNT | nit | `depictionCapabilities` name | Logged | |
| UX-5 | UX | minor | Hidden-region edit says "Draw on the file", board D9b says "Draw again" | Fix now (rides F-3) | one line, file touched |
| UX-6 | UX | minor | `propertyRequiredAfterSubject` left with no caller | Fix now (rides F-5) | DoD "no stale strings" |
| CON-3 | CON | minor | Archived S10-D15 brief links to `archive/archive/…` | Fix now (rides F-7) | verified path does not exist |
| CON-4 | CON | minor | `depictionPreview` and drawing-time guards untested | Fix now (rides F-6, F-1) | test file touched |
| CON outside-lens | UX | nit | S10-D15 review log left in `design/` after brief moved | Fix now (rides F-7) | one `git mv` |

### Fix plan 1

**F-1 — Never let the overlay commit on the wrong page; delete `ignoreNextDisarm`** (from MNT-1, UX-4, CON-2, MNT-3, CON-4)
Files: `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`, `CitationComposerModel+Depiction.swift`, `CitationComposerModel.swift`, `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`. `ArtifactViewer`'s API stays unchanged (plan "Builds on").
Failing test first: `CitationComposerDepictionTests/viewerToolIsUnarmedOnTheWrongPage`: an existing PDF Citation saved on page 1, Likeness draft, `startDepictionDrawing()`, `artifactViewer.setPage(2)` → the model's viewer-facing armed tool (new `viewerArmedRegionTool`, below) is nil, `armedRegionTool` is still `.rectangle`, `depictionWrongPage` is true; back to page 1 → `viewerArmedRegionTool == .rectangle`. Then `disarmRegionTool()` on page 1 while drawing cancels drawing (no swallowed disarm).
Do:
- Add `viewerArmedRegionTool: ArtifactRegionTool?` (get: nil while `depictionWrongPage` or `depictionNeedsPage` is true while drawing, else `armedRegionTool`; set: writes `armedRegionTool`). In `CitationComposerViewerPane`, pass a `Binding` over it to `armedRegionTool:` in place of `$model.armedRegionTool`. The overlay is then never armed on the wrong page, so no refused commit and no trailing disarm reach the model from the UI.
- Delete `ignoreNextDisarm` and its branch in `disarmRegionTool`, and its resets in `startDepictionDrawing` / `cancelDepictionDrawing`.
- MNT-3: one private predicate (e.g. `enum DepictionPageCheck { case ok, needsPage, wrongPage(Int), stampViewerPage }`) computed once; `depictionNeedsPage`, `depictionWrongPage`, and the guard in `commitViewerRegion` all read it. `commitViewerRegion` still refuses on `.needsPage` / `.wrongPage` (defence; returns false, no write).
- CON-4 (guards): in `drawingDisablesCitationControlsAndLeaveReturns`, also call `clearRegion()`, `requestSelectArtifact(_:)` with another Artifact, `requestAutoTranscribe()`, and `requestPasteTranscription()` while drawing, and assert `locator`, `selectedArtifactID`, `pendingTranscriptionConfirm` (and the paste state) are unchanged.
Check: the new and extended tests pass; `rg ignoreNextDisarm macos` finds nothing; scripts/check-pr.sh PASS.

**F-2 — Reset the depiction region only after the open Citation actually changes** (from CON-1)
Files: `CitationComposerModel+Authoring.swift`, `CitationComposerModel.swift`, `CitationComposerDepictionTests.swift`.
Failing test first: `CitationComposerDepictionTests/keepEditingAfterACitationSwitchKeepsTheRegion`, the code in CON-1 (unsaved transcription, `setAuthoringCitation("cit-1")`, `pendingLeave?.keepDraft == true`, `keepEditingAfterLeave()`, region still the drawn face).
Do: remove the early `mutateProperty { $0.values.depictionDraft = .empty() }` from `setAuthoringCitation`. Reset `depictionDraft` through `mutateProperty` in `loadCitationPreservingAuthoring(_:)` after the load succeeds and the open Citation id differs from before, before `presentAfterIdentityChange`. Also add a case that a confirmed switch (Discard) does clear the region (D-12).
Check: both cases pass; scripts/check-pr.sh PASS.

**F-3 — Tab from Property lands on the Draw button** (from UX-1, MNT-5, UX-5)
Files: `CitationComposerAuthoringBox.swift`, `CitationComposerSubjectGroups.swift`, `CitationComposerAuthoring.swift` (`CitationAuthoringNavigation`).
Failing test first: none at unit level for focus; if `CitationAuthoringNavigation` decides the Tab destination, add `CitationComposerDepictionTests/tabFromPropertySkipsADisabledDraw` (needs-page depiction → destination after `.property` is nil/footer, not `.value`).
Do:
- Pass `focus` into `CitationDepictionValueArea` and bind its Draw / Draw again `PVButton` with `.focused(focus, equals: .value)`, the same as the other value controls. When Draw is disabled (`depictionNeedsPage` or transcribing), the Tab destination after `.property` is not `.value`.
- MNT-5: add `CitationAuthoringNavigation.initialPropertyFocus(values:vocabulary:)` and call it from `CitationComposerAuthoringBox.propertyFocus(for:)` and `CitationComposerSubjectGroups` (~line 351).
- UX-5: the button reads `depictionRedraw` when `draft.hasRegion || !draft.valueDepictionID.isEmpty` (board D9b).
Check: build; existing tests green; scripts/check-pr.sh PASS.

**F-4 — The Property field shows the kept Likeness after "next face"** (from UX-2, MNT-4)
Files: `CitationComposerAuthoringBox.swift` (`CitationPropertyFields.propertyOptions`), `CitationComposerModel+Authoring.swift`, `CitationComposerModel.swift`, `CitationComposerDepictionTests.swift`.
Failing test first: extend `nextFaceKeepsThePropertyAndClearsThePerson`, or add `keptPropertyIsAnOptionWithNoSubject`: after a save and reset, the options the property combo is given (factor `CitationPropertyFields.propertyOptions` into a model/vocabulary function the test can call if it is view-private) contain the kept Likeness Property, so its label is "Likeness".
Do: when `values.propertyID` is set and not in the subject-filtered list, append `vocabulary.property(id: values.propertyID)`. `applyObservationSubject` keeps dropping it for a subject type that does not offer it. MNT-4: replace the inline `property.valueType == PropertyValueType.depiction.rawValue` checks (`CitationComposerModel+Authoring.swift:541`, `CitationComposerModel.swift:340`) with `vocabulary.isDepiction(propertyID:)`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-5 — VoiceOver hears Negative and Saved on a likeness line** (from UX-3, MNT-6, UX-6)
Files: `CitationComposerSubjectGroups.swift`, `macos/App/Platform/L10n.swift`, `Localizable.xcstrings`. Skill: `add-localized-string`.
Failing test first: if the label is computed by a non-view function, add `CitationComposerDepictionTests/negativeLikenessLabelSaysNegative`; else none (manual check line below).
Do:
- Compute one `String?` row label per line (depiction label or nil) and apply `.accessibilityLabel` only when set, with no branching `@ViewBuilder` modifier (MNT-6); rename so the L10n key, the helper, and the modifier no longer share `depictionLineLabel`.
- Add `.accessibilityValue` with the existing `negativeLabel` when `line.negative`, and the existing Saved badge text when the line is fresh (reuse the strings the badges already use; no new copy).
- UX-6: delete `propertyRequiredAfterSubject` from `L10n.swift` and `Localizable.xcstrings`.
- Add a manual-check line to the PR body: "VoiceOver on Clara's negative likeness reads Negative."
Check: `python3 scripts/check-localizable-xcstrings.py` passes; scripts/check-pr.sh PASS.

**F-6 — Save reuses the preview render; PNG encode leaves the main actor** (from PRF-1 partial, CON-4)
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerModel.swift` (`depictionPreviewMemo`), `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift`, `CitationComposerDepictionTests.swift`.
Failing test first: in `drawingLeavesTheCitationRegion`, after the commit, `model.depictionPreview?.size` is 4×4 on the 8×8 fixture, and nil with no region (CON-4). For reuse, add a render counter seam only if one is cheap; otherwise assert behaviour (Save still sends a PNG; `savingSendsAPNGAndDeletesIt` stays green).
Do: keep the rendered `CGImage` in `depictionPreviewMemo` alongside the `NSImage` (same key: Artifact id + `locator.encodeJSON()`). `withDepictionInput` takes the memo's `CGImage` when the key matches the draft being saved, else renders (D-5). Encode the PNG (`writeTemporaryPNG`) in a detached task and await it; the `defer` removal stays. Do not change `pageBitmap` (D-5 stands; box-only render is a follow-up).
Check: tests pass; scripts/check-pr.sh PASS.

**F-7 — Docs: dogfood entries, broken link, review log beside its brief** (from MNT-2, CON-3, CON outside-lens)
Files: `docs/dogfood/ux.md`, `docs/deployment-plan/spike-10/design/archive/S10-D15-composer-depiction-row.md`, `docs/deployment-plan/spike-10/design/S10-D15-composer-depiction-row.review.md` (if it exists).
Do: add the plan's two Out entries to `docs/dogfood/ux.md` ("Guarding a Citation page change after depictions are saved on it"; under tooling, "Folding the other test copies of PNG / PDF writers and `waitUntil` into `ArtifactFileFixtures`", naming `ArtifactViewerModelTests.swift`, `PromoteModelTests.swift`, `CitationAutoTranscribeTests.swift`, `EvidenceGraphModelTests.swift`, `CitationComposerModelTests.swift` `seedArtifact`). Fix the archived brief's link to `S10-D3-citation-composer.md`. `git mv` the S10-D15 review log into `design/archive/` and fix any links to it.
Check: links resolve (`ls` each target); `python3 scripts/check-plan-docs.py` adds no new error; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0891a1a` (commits `fd7ea3a`, `2c50839`, `0891a1a`). Gate PASS (orchestrator run; 1128 Swift tests). F-1 to F-7 done. Deviations: F-3/F-4/F-6 tests written with the fix (new API needed to compile); F-1–F-4 and F-6 share one commit; F-5 keeps an `if let` modifier (renamed `lineReading(_:)`), L10n key `depictionLineLabel` unchanged; F-1 binding goes to `ArtifactViewer` only, the tool-strip radios still bind `armedRegionTool`; dogfood tooling entry sits under Open with Where: Tooling; PR body line for VoiceOver left to the orchestrator.

## Round 2 — review
Lenses: CON, MNT, UX, PRF. Scope `4c664a61...0891a1af`.
PRF: clean (PRF-1 Fix-now part resolved). UX: clean (UX-1–6 resolved). CON: clean (CON-1–3 resolved; CON-4 guards half-done). MNT: MNT-1–6 resolved; 1 new blocker, 1 minor.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-9 (+UX, CON outside-lens notes) | MNT | blocker | "Draw can start" written three times (model guard, button, Tab rule) and already drifted: Draw looks enabled while transcribing | Fix now → F-8 | verified: button `.disabled(model.depictionNeedsPage)` only; guard and Tab rule add `isTranscribing`. Three lenses saw it |
| CON-4 (residual) | CON | minor | Drawing-guard asserts pass with the guards deleted | Fix now (rides F-8) | test file touched by F-8 |
| MNT-10 | MNT | minor | `propertyOptions(keeping:)` comment narrower than the code | Fix now (rides F-8) | comment only |

### Fix plan 2

**F-8 — One `canStartDepictionDrawing` for the guard, the button, and Tab** (from MNT-9, CON-4, MNT-10)
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerAuthoringBox.swift`, `CitationComposerVocabulary.swift`, `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`.
Failing test first: `CitationComposerDepictionTests/drawCannotStartWhileTranscribing` — Likeness draft with a region-capable Artifact, put the model in `isTranscribing` (the way existing Auto Transcribe tests do) → `canStartDepictionDrawing` is false and `startDepictionDrawing()` leaves `isDrawingDepiction` false; not transcribing → true.
Do:
- Add `var canStartDepictionDrawing: Bool` (`currentDepictionDraft != nil && !isTranscribing && !depictionNeedsPage`). `startDepictionDrawing` guards on it; the Draw button uses `.disabled(!model.canStartDepictionDrawing)`; `propertyValueTakesFocus` uses `!(isDepiction(committed) && !model.canStartDepictionDrawing)` (or the equivalent that does not require a current draft at Tab time — keep today's Tab behaviour when no draft yet). No other copy of the rule remains.
- CON-4: make `drawingDisablesCitationControlsAndLeaveReturns` meaningful: seed the Citation with a region, seed a second Artifact and await the switch attempt, and use a fixture where `canAutoTranscribe` / `canPasteTranscription` hold (assert them true first), so each assert fails if its guard is removed. Check by deleting one guard locally and seeing the test fail, then restore.
- MNT-10: reword the `propertyOptions(forSubjectID:keeping:capabilities:)` doc comment to say the held Property is added whenever it is missing from the subject's list, and that `applyObservationSubject` drops an incompatible Property first.
Check: tests pass; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `e749bb7`. Gate PASS (orchestrator run; 1130 Swift tests). F-8 done. Deviation: `depictionDrawBlocked` (`isTranscribing || depictionNeedsPage`) is the one rule; `canStartDepictionDrawing` adds "a draft is open" for the guard and button, and the Tab rule reads `depictionDrawBlocked` (no draft may exist at Tab time). Each of the four drawing guards was deleted in turn and the suite failed each time.

## Result
Gate PASS · CI see PR · open: none · follow-ups: PRF-1 (box-only PDF render, user 2026-10-17), MNT-7 · to user: PRF-1 box-only render → Follow-up (answered)
