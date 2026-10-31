# S14-03 — Rectangle-only depiction drawing

**Status:** Blocked — on board S14-D6
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S14-D6](../design/S14-D6-composer-drawing-feedback.md), board not drawn yet (user, 2026-10-31: the PR changes how the composer looks, so it waits on a brief)
**Board:** — (when agreed: the *Citation Composer* Claude Design project, turn "S14-D6 · Drawing feedback"; frames described in words in § 6 macOS until then)
**Depends on:** S14-13 (landed [#632](https://github.com/mendahu/provenencia/pull/632)), S14-D6
**Blocks:** —
**Branch:** `s14-03-rectangle-only-depiction-drawing`
**Version bump:** patch, 0.2.2 → 0.2.3 — fixes to existing drawing feedback and copy, no new feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md) (read only: what Go accepts for a depiction region)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/depiction-values.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rule "A depiction is drawn as a rectangle"; bar outcomes 4 and 9; the 2026-10-18 likeness audit items UX-5, UX-6, UX-7 (composer feedback), UX-12, UX-13 (copy) and the freehand cap feedback, whose text is in `git show f88bef38^:docs/ideas/depictions-and-likenesses.md` lines 19 – 21; S14-13's Landed note (the "can cut" predicate is `ArtifactRegionCrop.snapshot`)
**Planned:** 2026-10-31 against `main` @ `f2dc2a95`
**Review:** — (not reviewed yet)

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

A researcher drawing a likeness or a picture on a Citation's file is offered the rectangle tool alone, because the saved image is always the rectangle around the region; "Use the citation's region" on a circled entry sets the circle's bounding rectangle. While any region is drawn, the controls that would move the Citation's locator under the drawing are disabled. Draw is blocked on a file the viewer cannot cut, a failed cut shows at once with Save disabled, a PDF Citation saved without a page can save its page from the warning, the captions name the region buttons, and VoiceOver hears the freehand cap. The work comes from the user's rectangle-only call (2026-10-27) and the Spike 10 likeness audit leftovers moved here.

**Why now.** S14-02 shipped Event and Place pictures, so more depictions are drawn; S14-13 left the one "can cut" predicate this PR gates Draw on.

**What stays.** A Citation's own region keeps all seven tools. Go accepts any valid depiction region: a polygon depiction saved earlier opens, shows and saves as it is until it is drawn again. The composer's panes, the depiction row's layout, the drawing band, and the cut itself (`ArtifactRegionCrop.cut`) are unchanged.

## 2. User stories

**US-1.** As a researcher drawing a depiction, I am offered the rectangle tool alone, so that what I draw is what is saved.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Likeness dialog open on an image Citation, Draw pressed | `viewerRegionTools == [.rectangle]`; `armedRegionTool == .rectangle` |
| AC-1.2 | Freeform armed for the Citation region, then a likeness dialog opened and Draw pressed | `armedRegionTool == .rectangle` |
| AC-1.3 | Likeness dialog open, not drawing | `viewerRegionTools == [.rectangle]` |
| AC-1.4 | No depiction dialog open | `viewerRegionTools == ArtifactRegionTool.allCases` (seven) |
| AC-1.5 | Drawing a depiction, the selected rectangle radio clicked (`stripArmedRegionTool = nil`) | drawing ends: `presented == .authoring`, `armedRegionTool == nil` |
| AC-1.6 | Edit of a saved likeness whose region is a 6-point freeform polygon, not redrawn, saved | Save sends no new image; the region is unchanged |

**US-2.** As a researcher, "Use the citation's region" on a non-rectangle region sets the rectangle around it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Citation region a circle ring centered (0.5, 0.5), radius 0.2 in a square file | `depictionCitationRegion` is `.rectangle` with points (0.3, 0.3), (0.7, 0.3), (0.7, 0.7), (0.3, 0.7) within 1e-6 |
| AC-2.2 | Citation region `.lOpenTopRight` inside box (0.1, 0.2) – (0.6, 0.9) | `.rectangle` over (0.1, 0.2) – (0.6, 0.9) |
| AC-2.3 | Citation region `.rectangle` | returned unchanged (same points, same order) |

**US-3.** As a researcher drawing any region, the controls that would change the Citation's locator under me are disabled.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A Citation region tool armed (no dialog) | `isDrawingRegion`; `setPageFromViewer()` returns false; `removePage()` and `removeRegion()` leave `locator` unchanged |
| AC-3.2 | Drawing a depiction | `isDrawingRegion`, same as AC-3.1 |
| AC-3.3 | PDF Citation, Find open, a Citation region tool armed through `stripArmedRegionTool` | `artifactViewer.isFindPresented == false` |
| AC-3.4 | No tool armed, not drawing | `isDrawingRegion == false`; Set page works |

**US-4.** As a researcher, Draw is blocked on a file the viewer cannot cut, and a failed cut shows at once with Save disabled.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Image Artifact whose file is missing (viewer `emptyReason == .missingFile`), likeness dialog open | `depictionCanCut == false`; `canStartDepictionDrawing == false`; `useDepictionRegion(.whole)` returns false; `depictionCaption == depictionCaptionCannotCut` |
| AC-4.2 | Viewer loading (`isLoading`) | `canStartDepictionDrawing == false`; caption is the empty caption |
| AC-4.3 | PDF opened in the viewer, its object file then deleted, a region committed and the preview rendered | `depictionPreviewState == .failed`; `depictionCutFailed`; `authoringCanConfirm == false`; caption `depictionRenderFailed`; `canStartDepictionDrawing == true` |
| AC-4.4 | AC-4.3 on an edit with a redrawn region | `editCanConfirm == false` |

**US-5.** As a researcher on a PDF Citation saved without a page, I save the viewer's page to it from the warning and then draw.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | Existing PDF Citation with no page, viewer on page 2, likeness dialog open, `saveDepictionCitationPage()` | FakeStore's `citationsByID["cit-1"].locatorJSON` decodes to page 2; `depictionNeedsPage == false`; `canStartDepictionDrawing == true` |
| AC-5.2 | Same, with `cit-1` removed from FakeStore's `citationsByID` so `updateCitation` throws (its existing `StoreBoom.boom`; no new hook, working rule "FakeStore is a seeded fake") | `depictionNeedsPage == true`; `fields.error` is the store's message |
| AC-5.3 | Image Citation, or PDF Citation with a page | `saveDepictionCitationPage()` adds no `updateCitation` to FakeStore's `recordedCalls` |

**US-6.** As a researcher, the depiction copy names what I can do, and VoiceOver hears the freehand cap.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-6.1 | No region drawn | caption "Nothing drawn yet. Draw on the file, or use a region button below" |
| AC-6.2 | PDF Citation needs a page | caption "Draw and the region buttons wait until the citation has a page"; Callout "This citation has no page yet. Save the viewer's page to it to draw a region."; button "Save page 2 to this citation" |
| AC-6.3 | Saved region hidden | Callout "This region no longer matches the citation's page. Draw it again or use a region button below." |
| AC-6.4 | Edit with the region unchanged | `depictionPreviewLabel == depictionPreviewSavedLabel` ("Saved image"); a redrawn edit and a new draft read "Image to save" |
| AC-6.5 | Freeform tool with 256 points placed, one more click | overlay beeps and posts `L10n.ArtifactViewer.freeformCapReached` once |

## 3. Scope

| In | Covered by |
| --- | --- |
| A depiction dialog offers `.rectangle` alone; a Citation region all seven (DF-1, DF-2) | STEP-1, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5 |
| `ArtifactRegionDraft.boundingRectangle` for "Use the citation's region" | STEP-2, AC-2.1, AC-2.2, AC-2.3 |
| Drawing disables Set page, chip ×, Find; hides the locator hint (DF-3, UX-5) | STEP-3, AC-3.1, AC-3.2, AC-3.3, AC-3.4 |
| Draw and region buttons blocked on a file the viewer cannot cut (DF-4, UX-6) | STEP-4, AC-4.1, AC-4.2 |
| A failed cut shows with Save disabled (DF-5, UX-6) | STEP-5, AC-4.3, AC-4.4 |
| The needs-page Callout saves the viewer's page (DF-6, UX-7) | STEP-6, AC-5.1, AC-5.2, AC-5.3 |
| Captions and region-hidden Callout name the region buttons; "Saved image" label (DF-7 – DF-10, UX-12, UX-13) | STEP-7, AC-6.1, AC-6.2, AC-6.3, AC-6.4 |
| Freehand cap announced to VoiceOver (DF-11) | STEP-8, AC-6.5 |
| Wide: nothing moves (DF-W1) | Manual check 7 |

| Out | Home |
| --- | --- |
| Removing shapes from Citation regions | not planned (working rule "A depiction is drawn as a rectangle") |
| Go-side rectangle enforcement, rewriting saved polygon depictions | not planned (same rule: Go keeps accepting any valid region) |
| Reshaping a committed polygon depiction's handles into a rectangle | not planned: handle edits keep the saved kind (D-11) |
| The composer's 1400pt pane switch onto the width class | [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md), the composer's next layout brief |
| A refused image glyph (UX-11), thumbnail work (PRF-2), Go crop work (PRF-4), `encodeJSON` key order | [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |
| Showing every saved depiction on the photo | same idea file |

## 4. Current state

Verified at `f2dc2a95`.

| Area | Where | Today |
| --- | --- | --- |
| Tool enum | `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift` `ArtifactRegionTool`, `ArtifactRegionKind`, `ArtifactRegionDraft` (`isValid`, `static let whole`) | Seven tools; `ArtifactRegionDraft` has no bounding-rectangle conversion |
| Region geometry | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `boundingRect(of:)`, `rectanglePoints(from:to:)`, `maxRegionPoints` (256), `appendFreeformVertex` | The pieces `boundingRectangle` composes |
| Tool strip | `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` `ArtifactViewerToolChrome` `regionTools`, `regionRadio`, `setPageButton` | `ForEach(ArtifactRegionTool.allCases)`; a click on the selected radio sets the binding to nil; Set page has no disabled state. `ArtifactViewer` builds the same chrome when `showsToolStrip` (the composer passes false) |
| Composer strip | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `toolStrip`, `findToggleButton`, `locatorStrip`, `locatorChip` | Binds `$model.armedRegionTool`; Find toggle and ⌘F enabled while drawing; chip × enabled; `locatorDrawHint` always drawn; disables the whole strip only while transcribing |
| Drawing band | same file `drawingBand`, `claimDrawingFocus` | Hint or wrong-page warning, and Cancel. Unchanged |
| Composer tool state | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `armedRegionTool`, `setPageFromViewer`, `removePage`, `removeRegion`, `disarmRegionTool` | `setPageFromViewer`, `removePage`, `removeRegion` guard `isDrawingDepiction` only, not an armed Citation tool |
| Depiction drawing | `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift` `viewerArmedRegionTool`, `startDepictionDrawing`, `depictionCitationRegion`, `useDepictionRegion`, `depictionDrawBlocked`, `canStartDepictionDrawing`, `depictionCaption`, `depictionPreviewState`, `depictionNeedsPage` | `startDepictionDrawing` keeps an already armed tool (`if armedRegionTool == nil`); `depictionCitationRegion` returns the region as is; `depictionDrawBlocked` is `isTranscribing \|\| depictionNeedsPage` (no "can cut"); a failed cut only shows as an empty preview |
| Save gates | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `authoringCanConfirm`, `editCanConfirm`; catch of `DepictionRenderFailed` | Save enabled on a failed preview; the failure shows only after Save |
| Depiction row | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` (`preview`, `useRegion`) | Needs-page and region-hidden `PVCallout`s with no actions; `.failed` draws the same empty `PVThumbnail` as `.none`; `.saved` labelled `depictionPreviewLabel` ("Image to save") |
| Can-cut predicate | `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` `snapshot(_:page:)` | Nil when the viewer holds nothing `cut` can cut (no raster, no PDF document or source, no page). Read only by `render` |
| Freehand cap | `macos/App/Features/ArtifactViewer/ArtifactRegionOverlay.swift` `ArtifactRegionOverlayView.handleFreeformClick` (private) | `NSSound.beep()` only |
| Announcements | `macos/App/Platform/AccessibilityAnnouncement.swift` `post(_:priority:)` | The one VoiceOver announcement path |
| Citation save | `macos/App/Features/CitationComposer/CitationFieldsDraft.swift` `saveCitation()` | Saves the whole Citation draft, sets `baseline`, reports `error` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.depictionCaptionEmpty`, `depictionCaptionNeedsPage`, `depictionNeedsPage`, `depictionRegionHidden`, `depictionPreviewLabel`, `depictionRenderFailed`, `locatorDrawHint`; `macos/App/Resources/Localizable.xcstrings` | Copy names Draw only (UX-12) |
| Tests | `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` `existingPDFCitationKeepsItsPage`, `viewerToolIsUnarmedOnTheWrongPage`, `drawingDisablesCitationControlsAndLeaveReturns`, `missingFileCutFails`, `captionFollowsTheDraft`, `citationRegionIsCopied`, `regionButtonsFollowDraw`; `macos/ProvenenciaTests/ArtifactRegionGeometryTests.swift`; `macos/ProvenenciaTests/ArtifactRegionCropTests.swift` (PDF deleted after open → `cut` nil, lines 170 – 190) | `missingFileCutFails` draws on a missing file and expects a failed cut; after this PR Draw is blocked there (T-6 replaces it) |

**Precedent.** S10-63 (keyboard likeness region) added `useDepictionRegion` and the region buttons through `prepareDepictionRegion`, one rule for Draw and the buttons: extend that rule (`depictionDrawBlocked`) rather than adding a second gate. S14-13 ([#632](https://github.com/mendahu/provenencia/pull/632)) left `ArtifactRegionCrop.snapshot` as the one can-cut predicate: read it, do not re-derive from `emptyReason`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ArtifactViewerToolChrome` takes `regionTools: [ArtifactRegionTool]` (default `ArtifactRegionTool.allCases`) and draws only those. The composer passes `model.viewerRegionTools` *(new)*: `[.rectangle]` when `currentDepictionDraft != nil`, else `ArtifactRegionTool.allCases`. | The viewer stays unaware of depictions (map scope). Keyed on the open depiction draft, not on drawing mode, because `commitViewerRegion` sends any viewer commit to the open depiction draft: a freeform arm before Draw would otherwise save a polygon depiction. | Keying on `isDrawingDepiction` — leaks six tools into the depiction while its dialog is open. A depiction flag on the viewer — the viewer would learn the domain |
| D-2 | `viewerArmedRegionTool`'s getter also returns nil when `armedRegionTool` is not in `viewerRegionTools`. | A freeform tool armed for the Citation before the dialog opened must not commit into the depiction. | Clearing `armedRegionTool` when the dialog opens — a second place that writes tool state |
| D-3 | The strip binds `$model.stripArmedRegionTool` *(new)*: get `armedRegionTool`; set a tool → `artifactViewer.closeFind()` then `armedRegionTool = tool`; set nil → `disarmRegionTool()`. `startDepictionDrawing` sets `armedRegionTool = .rectangle` always. | One setter: clicking the selected rectangle in drawing mode leaves drawing, as Escape does (DF-1), and arming any tool closes Find (DF-3). | Blocking deselect in drawing mode — the radio would do nothing |
| D-4 | `ArtifactRegionDraft.boundingRectangle` *(new)*: `self` when `kind == .rectangle`; else `.rectangle` with `rectanglePoints` of `boundingRect(of: points)`'s corners. `depictionCitationRegion` returns `region.boundingRectangle` (still nil when the region is invalid). | One conversion beside `.whole`; a saved rectangle keeps its vertex order, so `citationRegionIsCopied` holds. Bounding rect area ≥ polygon area, so a valid region gives a valid rectangle. | Converting in `useDepictionRegion` — the button would show one region and set another |
| D-5 | `isDrawingRegion` *(new)* = `isDrawingDepiction \|\| armedRegionTool != nil`. `setPageFromViewer`, `removePage`, `removeRegion` guard it (replacing their `isDrawingDepiction` guard). The view disables Set page (`ArtifactViewerToolChrome.isSetPageEnabled` *(new)*, default true), the chip ×, the Find toggle and its ⌘F button on it, and drops `locatorDrawHint`. | The scope says "while a region or a depiction is drawn"; the model guards hold the rule, the view mirrors it. | Hiding the controls — the layout jumps (brief non-goal) |
| D-6 | `ArtifactRegionCrop.canCut(_:page:) -> Bool` *(new, @MainActor)* = `snapshot(viewer, page:) != nil`. `depictionCanCut` *(new)* = `canCut(artifactViewer, page: depictionPage ?? artifactViewer.page)`. `depictionDrawBlocked` = `isTranscribing \|\| depictionNeedsPage \|\| !depictionCanCut`. | One predicate, as S14-13 left it. Through `depictionDrawBlocked`, Draw, both region buttons, `prepareDepictionRegion` and the Tab order all gain it. | Reading `emptyReason` — a second definition of "can cut" |
| D-7 | `depictionCutFailed` *(new)* = `depictionPreviewState == .failed`. `authoringCanConfirm` and `editCanConfirm` return false for a property draft while it is true. The preview draws `PVThumbnail` with `glyph: .symbol(.warning)` for `.failed`; the caption is `depictionRenderFailed` in `PVColor.warningForeground`. | Shows the failure where the researcher looks, before Save (UX-6). Reuses the existing failure copy. | A Toast or Callout — brief non-goal |
| D-8 | Caption order in `depictionCaption`: needs page → cannot cut (`!artifactViewer.isLoading && !depictionCanCut`) → cut failed → region hidden → empty → redrawn → box. `depictionCaptionIsWarning` *(new)* is true for cut failed. | Each state has one caption; the earliest blocker wins. Loading keeps the empty caption (DF-4). | — |
| D-9 | `saveDepictionCitationPage() async` *(new)*: returns unless `depictionNeedsPage`, not transcribing and not `fields.isSaving`; `setPage(artifactViewer.page)`; `await fields.saveCitation()`; on success posts `depictionCaption` through `AccessibilityAnnouncement.post`. The Callout's Button reads `L10n.CitationComposer.depictionSavePage(artifactViewer.page)` with `loading: model.fields.isSaving`. It saves the whole Citation draft, as Save does. | User, 2026-10-31: set the page and save in one click. A failure stays in `fields.error`, shown where the composer shows it today. | Setting the page only — the user chose save |
| D-10 | The preview's label is `depictionPreviewLabel` *(new computed on the model)*: `L10n.CitationComposer.depictionPreviewSavedLabel` ("Saved image") when `depictionPreviewState` is `.saved`, else `depictionPreviewLabel` ("Image to save"). Used by every preview case. | UX-13: an unchanged edit saves no image. | — |
| D-11 | A committed polygon depiction keeps its handle edits as they are (the overlay edits the committed shape in its own kind); only a new draw is a rectangle. | Working rule: a polygon depiction "saves as it is until redrawn". | Converting on handle edit — rewrites a saved polygon without a redraw (Out) |
| D-12 | The freehand cap: `ArtifactRegionOverlayView` gains `var announce: @MainActor (String) -> Void = { AccessibilityAnnouncement.post($0) }` and `handleFreeformClick` becomes internal; at the cap it beeps and calls `announce(L10n.string(L10n.ArtifactViewer.freeformCapReached))`. | Keeps the beep; the seam makes the announcement testable without VoiceOver. | Posting from the composer — the cap lives in the overlay |
| D-13 | Tests live in `CitationComposerDepictionTests.swift` and `ArtifactRegionGeometryTests.swift`, plus `macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` *(new)* wired into the Xcode target per `add-swift-test`. | The overlay has no test file today. | — |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionDraft.boundingRectangle` *(new)* | func | `ArtifactRegionDraft.whole`, `ArtifactRegionCrop.boundingBox(of:region:)` | Extends the draft beside `.whole`; `boundingBox` crops pixels, not a region |
| `ArtifactRegionCrop.canCut(_:page:)` *(new)* | func | `ArtifactRegionCrop.snapshot` | A Bool over `snapshot`, so callers do not hold a `Snapshot` |
| `ArtifactViewerToolChrome.regionTools`, `.isSetPageEnabled` *(new)* | params | `armedRegionTool`, `onSetPage` | Extends the one strip |
| `viewerRegionTools`, `stripArmedRegionTool`, `isDrawingRegion`, `depictionCanCut`, `depictionCutFailed`, `depictionCaptionIsWarning`, `depictionPreviewLabel`, `saveDepictionCitationPage()` *(new)* | composer model members | `viewerArmedRegionTool`, `isDrawingDepiction`, `depictionDrawBlocked`, `depictionPreviewState`, `depictionCaption`, `setPageFromViewer` | Each derives from the existing state beside its closest member; none stores state |
| `ArtifactRegionOverlayView.announce` *(new)* | property | `onCommit`, `onDisarm` closures | Same closure-seam style |
| L10n keys *(new)*: `CitationComposer.depictionCaptionCannotCut`, `depictionSavePage(_:)`, `depictionPreviewSavedLabel`; `ArtifactViewer.freeformCapReached` | key | `L10n.CitationComposer.depiction*` | Same namespace |
| `ArtifactRegionOverlayTests.swift` *(new)* | test file | `ArtifactRegionGeometryTests.swift` | The overlay view is not geometry |

**Builds on.** `prepareDepictionRegion` / `depictionDrawBlocked` (one rule for Draw, buttons and Tab), `ArtifactRegionCrop.snapshot` (S14-13), `PVCallout`'s actions row, `PVThumbnail` symbol content, `AccessibilityAnnouncement`.
**Deletes.** The `if armedRegionTool == nil` in `startDepictionDrawing`; `missingFileCutFails` (replaced by T-6 and T-7); the `isDrawingDepiction` guards in `setPageFromViewer`, `removePage`, `removeRegion` (replaced by `isDrawingRegion`).
**Next adds.** A signature key (idea file) is one more depiction Property and gets the rectangle tool with no change. The composer's layout brief moves its pane switch without touching this.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Draw blocked while the viewer loads or shows an empty state | `CitationDepictionValueArea` buttons, Tab order (`claimTab`) | Tab skips a disabled Draw already (`tabFromPropertySkipsADisabledDraw`); the block lifts when the viewer finishes loading because `depictionCanCut` is derived |
| Citation saved from the observation dialog (D-9) | `CitationFieldsDraft.saveCitation` | Same call the evidence column's Save makes while a dialog is open; baseline moves, so `depictionNeedsPage` clears |
| Armed tool hidden by D-2 | overlay `armedTool` nil | The strip shows no selected radio; Draw arms `.rectangle` |

### Scale

None — no write, list or loop changes beyond one Citation save on a button press.

### Edges

- **Assumptions.** `commitViewerRegion` routes to the depiction draft whenever `currentDepictionDraft != nil`; D-1 and D-2 key on the same condition so they cannot disagree (T-1, T-2).
- **Neighbors.** A rectangle region → itself (T-4); circle, L and freeform → their box (T-4); a region touching the unit square's edge → `rectanglePoints` clamps (T-4 case at (0, 0) – (1, 1)).
- **Removed inputs.** A viewer that unloads (Artifact switch) makes `depictionCanCut` false until the new file loads (T-6).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — no list read from data the app did not write.

### Schema

None — no catalog change.

### Go core

None — Go keeps accepting any valid depiction region.

### FFI

None —.

### macOS

Until the board is agreed, each frame in words (brief § 5):

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Region radios | `PVIconButton` in `ArtifactViewerToolChrome` | `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` | Frame 1: rectangle alone, selected; frame 2: seven |
| Set page | `setPageButton` + `.disabled(!isSetPageEnabled)` | same | Frame 1 (PDF) |
| Chip × and Find | `PVIconButton` `.disabled(model.isDrawingRegion)` | `CitationComposerViewerPane.swift` | Frames 1, 2 |
| Locator hint | `Text(locatorDrawHint)` drawn only when `!model.isDrawingRegion` | same | Frames 1, 2: absent |
| Needs-page Callout | `PVCallout(tone: .warning, message:) { PVButton(..., variant: .secondary, size: .sm, loading:) }` | `CitationComposerAuthoringBox.swift` | Frames 7, 8 |
| Failed preview | `PVThumbnail(Content(glyph: .symbol(.warning), label:), size: 96)` | same | Frame 5 |
| Caption | existing `Text`, `foregroundStyle(model.depictionCaptionIsWarning ? PVColor.warningForeground : PVColor.textMuted)` | same | Frames 3 – 5, 7 |

States: loading (buttons disabled, empty caption), cannot cut, failed, needs page, hidden, empty, redrawn, box. Accessibility: the Save page button's label is its title; DF-10 label; DF-11 announcement; `useRegion` and D-9 announce the new caption.

### Strings

| Key | Value |
| --- | --- |
| `L10n.CitationComposer.depictionCaptionEmpty` (changed) | "Nothing drawn yet. Draw on the file, or use a region button below" |
| `L10n.CitationComposer.depictionCaptionNeedsPage` (changed) | "Draw and the region buttons wait until the citation has a page" |
| `L10n.CitationComposer.depictionNeedsPage` (changed) | "This citation has no page yet. Save the viewer's page to it to draw a region." |
| `L10n.CitationComposer.depictionRegionHidden` (changed) | "This region no longer matches the citation's page. Draw it again or use a region button below." |
| `L10n.CitationComposer.depictionCaptionCannotCut` *(new)* | "Draw is unavailable: the file is not showing in the viewer" |
| `L10n.CitationComposer.depictionSavePage(_:)` *(new)* | "Save page %lld to this citation" |
| `L10n.CitationComposer.depictionPreviewSavedLabel` *(new)* | "Saved image" |
| `L10n.ArtifactViewer.freeformCapReached` *(new)* | "This freehand region holds its most points, 256. Click the first point to close it." |

Old-scope copy: `depictionDrawingHint` ("Draw the region with the tools above. Escape goes back to the observation.") stays: one tool is still "the tools above" in the strip; the board may reword it (Stop condition).

## 7. Steps

**STEP-1 — One tool for a depiction**
Files: `ArtifactViewer.swift`, `CitationComposerViewerPane.swift`, `CitationComposerModel+Depiction.swift`. Skill: `add-ui-component`.
Do: D-1, D-2, D-3. The pane passes `regionTools: model.viewerRegionTools` and binds `$model.stripArmedRegionTool`.
Check: T-1, T-2, T-3 pass; `viewerToolIsUnarmedOnTheWrongPage` still passes.

**STEP-2 — Bounding rectangle**
Files: `ArtifactRegionTool.swift`, `CitationComposerModel+Depiction.swift`.
Do: D-4.
Check: T-4, T-5 pass; `citationRegionIsCopied`, `citationRegionOnItsPage` still pass.

**STEP-3 — Drawing locks the locator**
Files: `CitationComposerModel.swift`, `ArtifactViewer.swift`, `CitationComposerViewerPane.swift`.
Do: D-5.
Check: T-8 passes; `drawingDisablesCitationControlsAndLeaveReturns` still passes.

**STEP-4 — Draw needs a file the viewer can cut**
Files: `ArtifactRegionCrop.swift`, `CitationComposerModel+Depiction.swift`.
Do: D-6, the cannot-cut caption of D-8 (key added in STEP-7; add the key here with its final value).
Check: T-6 passes; `missingFileCutFails` removed; `tabFromPropertySkipsADisabledDraw`, `regionButtonsFollowDraw` still pass.

**STEP-5 — A failed cut shows with Save disabled**
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerModel+Authoring.swift`, `CitationComposerAuthoringBox.swift`.
Do: D-7, the failed caption of D-8.
Check: T-7 passes; `refusedSaveKeepsTheRegion`, `staleCutIsNotStored` still pass.

**STEP-6 — Save the page from the Callout**
Files: `CitationComposerModel+Depiction.swift`, `CitationComposerAuthoringBox.swift`, `Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`.
Do: D-9.
Check: T-9 passes; `existingPDFCitationKeepsItsPage` still passes.

**STEP-7 — Copy**
Files: `Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `CitationComposerModel+Depiction.swift`, `CitationComposerAuthoringBox.swift`. Skill: `add-localized-string`.
Do: § 6 Strings; D-10.
Check: T-10, T-11 pass; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-8 — Freehand cap announcement**
Files: `ArtifactRegionOverlay.swift`, `macos/ProvenenciaTests/ArtifactRegionOverlayTests.swift` *(new)*, `Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-12, D-13.
Check: T-12 passes.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row), this plan (status), `design/S14-D6-composer-drawing-feedback.md` → `design/archive/`, `design/README.md`, `docs/ideas/depictions-and-likenesses.md` (the audit paragraph names S14-03 as done).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-03-rectangle-only-depiction-drawing.md` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

All in `macos/ProvenenciaTests/`, Swift Testing.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CitationComposerDepictionTests` `depictionOffersTheRectangleAlone` *(new)* | no dialog → seven; likeness dialog → `[.rectangle]`; Draw → armed `.rectangle`; freeform armed first then Draw → `.rectangle` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1, D-3 |
| T-2 | Swift | same file `citationToolDoesNotReachTheDepiction` *(new)* | freeform armed, likeness dialog opened → `viewerArmedRegionTool == nil` | D-2 |
| T-3 | Swift | same file `deselectingTheRectangleEndsDrawing` *(new)* | drawing, `stripArmedRegionTool = nil` → `presented == .authoring`, armed nil | AC-1.5 |
| T-4 | Swift | `ArtifactRegionGeometryTests` `boundingRectangleOfEachKind` *(new)* | circle, L, freeform → box; rectangle → identical; full-square freeform → (0,0)–(1,1) | AC-2.1, AC-2.2, AC-2.3, D-4 |
| T-5 | Swift | `CitationComposerDepictionTests` `circledCitationRegionGivesItsBox` *(new)* | Citation circle region → `depictionCitationRegion` is its box; `useDepictionRegion` sets it | AC-2.1 |
| T-6 | Swift | same file `drawWaitsForACuttableFile` *(new; replaces `missingFileCutFails`)* | missing image file → cannot cut, caption cannot-cut, `useDepictionRegion(.whole)` false; loading → blocked, empty caption | AC-4.1, AC-4.2, D-6, D-8 |
| T-7 | Swift | same file `failedCutDisablesSave` *(new)* | PDF opened then its file deleted, region committed, preview rendered → `.failed`, `authoringCanConfirm == false`, caption render-failed and warning, Draw enabled; same on a redrawn edit → `editCanConfirm == false` | AC-4.3, AC-4.4, D-7 |
| T-8 | Swift | same file `drawingLocksTheLocator` *(new)* | Citation tool armed → `isDrawingRegion`, Set page / remove page / remove region refused; arming closes Find; nothing armed → Set page works | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-5 |
| T-9 | Swift | same file `calloutSavesTheCitationsPage` *(new)* | PDF Citation with no page, viewer page 2 → stored locator has page 2, needs-page clears, Draw enabled; citation removed from the fake → still needs page, `fields.error` set; image Citation → no `updateCitation` in `recordedCalls` | AC-5.1, AC-5.2, AC-5.3, D-9 |
| T-10 | Swift | same file `captionFollowsTheDraft` (extended) | empty, needs-page captions read the new values | AC-6.1, AC-6.2 |
| T-11 | Swift | same file `previewLabelFollowsTheSave` *(new)* | unchanged edit → "Saved image"; redrawn edit, new draft → "Image to save"; polygon edit not redrawn saves with no image | AC-6.4, AC-1.6, D-10, D-11 |
| T-12 | Swift | `ArtifactRegionOverlayTests` `freeformCapIsAnnounced` *(new)* | freeform armed, 256 clicks then one more through `handleFreeformClick` → `announce` called once with `freeformCapReached` | AC-6.5, D-12 |

AC-6.3 is covered by the manual check (Callout copy). Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerDepictionTests`, `CitationAutoTranscribeTests` (strip disabling while transcribing).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected; docs/ideas/depictions-and-likenesses.md audit paragraph updated
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. New project, an image Source with a family photo, a Citation with a circle region round a face. Add Likeness for a Person: the strip shows one radio. Draw: rectangle armed, Set page (none on an image), the region chip's × and Find (PDF only) disabled, no locator hint. Escape, then "Use the citation's region": the preview shows the circle's box. (Bar 4, bar 9.)
2. Close the dialog: seven radios. Arm freehand: chip × disabled, no hint. Place 256 points with VoiceOver on: the 257th click beeps and VoiceOver reads the cap. (Bar 9.)
3. Open a Citation whose image file was removed from the project's objects folder: the viewer's empty state, the likeness row's buttons disabled, the cannot-cut caption.
4. A PDF Citation: draw a likeness, delete the PDF from the project's objects folder before the preview finishes, then redraw: the warning preview, the warning caption, Save disabled.
5. A PDF Citation saved without a page, viewer on page 2: the Callout and "Save page 2 to this citation"; press it: the page chip appears, the Callout goes, Draw enabled. (Bar 9.)
6. Edit a saved likeness without redrawing: VoiceOver reads the preview as "Saved image". A saved polygon likeness (from an older build) opens and saves unchanged.
7. Widen the window past 1600pt of pane: frames 1 and 5 look the same, the canvas wider. Check light and dark.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Disabling chip × on an armed Citation tool surprises a researcher who arms a tool by accident | Clicking the selected radio, or Escape, disarms (D-3); manual check 2 |
| `depictionCanCut` false while the viewer loads hides Draw briefly | Empty caption stays (D-8); the block lifts on load |
| D-9 saves other unsaved Citation edits | User accepted, 2026-10-31 |

**Stop and report if:**

- The agreed S14-D6 board differs from § 6 or a DF requirement (copy, which controls disable, where the Save page button sits, the drawing hint's wording).
- `ArtifactRegionCrop.snapshot` is not nil for a viewer showing an empty state, or is nil for a loaded image or PDF page (D-6 would block or allow the wrong files).
- `commitViewerRegion` no longer routes to the depiction draft whenever `currentDepictionDraft != nil` (D-1, D-2 key on that).
- Disabling the chip × or Set page breaks a test in `CitationAutoTranscribeTests` that expects them enabled with a tool armed.
- `fields.saveCitation()` cannot run while the observation dialog is open (D-9).

## 12. Open questions

None. The plan is blocked on the board, not on a question.

## 13. PR

**Title:** S14-03 — Rectangle-only depiction drawing
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-03-rectangle-only-depiction-drawing.md`.
