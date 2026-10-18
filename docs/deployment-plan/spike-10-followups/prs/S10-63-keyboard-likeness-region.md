# S10-63 — Keyboard region for a likeness

**Status:** Draft
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (waived by the user 2026-10-18: two PVButtons in the existing Likeness value area, described in words in §6)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-63-keyboard-likeness-region`
**Version bump:** none — Spike 10 and Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [Spike 10 decisions](../../archive/spike-10/README.md#decisions) ("A depiction is an owned value row"), working rules: "CDR-13 is amended (user, 2026-10-18)"; the S10-D15 brief CDR-2, CDR-3, CDR-4, CDR-8, CDR-10, CDR-13 (`git show d1c00b5^2:docs/deployment-plan/spike-10/design/archive/S10-D15-composer-depiction-row.md`); the 2026-10-18 likeness area audit, finding UX-1
**Planned:** 2026-10-18 against `main` @ `aa60b63`
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

A researcher who works from the keyboard or with VoiceOver can save a likeness. Today a likeness's value is its region, the region can only be drawn with the pointer, and Save stays disabled until a region exists, so no keyboard or VoiceOver user can finish the Likeness observation (audit UX-1). This PR adds two keyboard-reachable buttons to the Likeness value area: one sets the region to the whole photo or the whole PDF page, and one copies the Citation's own region. Pressing **Draw on the file** also moves keyboard focus to the drawing band's **Cancel** and announces drawing mode, so a keyboard user who starts drawing by mistake learns how to go back.

**Why now.** The user decided on 2026-10-18 to amend CDR-13 ("Drawing stays a pointer action") rather than keep it: drawing stays pointer-only, and the two buttons are the keyboard path. A single portrait photo, or a Citation already drawn around one face, is the common case and needs no drawing at all.

**What stays.** Drawing on the canvas, its tools, the PDF page rule (CDR-10), the preview and caption (CDR-4), the save path (`withDepictionInput`), and Go's region validation stay as they are. A region set by a button is an ordinary region: the preview cuts it, Save sends it, and an edit treats it as a redraw when it differs from the saved one (CDR-8).

## 2. User stories

**US-1.** As a researcher citing a single portrait photo, when I open a Likeness observation, I press **Use the whole photo** so that the whole photo is the region and I can save without drawing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | An image Citation, a new Likeness draft for Ada; press **Use the whole photo** | The draft's region is the rectangle (0,0)–(1,1) with kind `rectangle`; `hasRegion` is true; the dialog stays on `.authoring`; the preview shows the whole photo; Save is enabled |
| AC-1.2 | Same draft, region set by AC-1.1; Save | `addObservationsToCitation` is called with `depiction=1 depictionFile=present`; the saved region JSON has one `region` selector and no `page` selector; the temporary PNG is deleted |
| AC-1.3 | A PDF Citation saved on page 2, viewer on page 1; new Likeness draft; the button reads **Use the whole page**; press it | The viewer moves to page 2; the draft's region is (0,0)–(1,1) on page 2 |
| AC-1.4 | A new, unsaved PDF Citation with no page, viewer on page 2; press **Use the whole page** | The Citation's page becomes 2 and the draft's region is the whole of page 2 (the same page stamp a drawn region gets) |
| AC-1.5 | An existing PDF Citation saved with no page (`depictionNeedsPage`) | Both buttons are disabled, as Draw is |
| AC-1.6 | A transcription is running | Both buttons are disabled, as Draw is |
| AC-1.7 | The Go region check receives a whole-file rectangle `[(0,0),(1,0),(1,1),(0,1)]` with the Citation's page | `locator.ValidateDepictionRegion` returns nil |

**US-2.** As a researcher whose Citation is already drawn around one face, I press **Use the citation's region** so that the likeness takes that region without drawing it again.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | An existing image Citation saved with the region (0.2,0.2)–(0.6,0.7); new Likeness draft; press **Use the citation's region** | The draft's region has the same kind and points as the Citation's saved region; Save is enabled |
| AC-2.2 | An existing PDF Citation saved on page 1 with a region; viewer on page 2; press the button | The viewer moves to page 1; the draft's region is the Citation's region on page 1 |
| AC-2.3 | A Citation with no region (artifact only, or page only) | **Use the citation's region** is not shown; **Use the whole photo** / **Use the whole page** is |
| AC-2.4 | A new, unsaved Citation where the researcher drew a region before opening the Likeness draft | The button uses that unsaved region |
| AC-2.5 | Editing a saved likeness with region (0.1,0.1)–(0.4,0.4); press **Use the citation's region** (a different region) | `isDepictionRedrawn` is true; the caption reads the Redrawn caption; Save sends a new image (`depiction=1 depictionFile=present`) |
| AC-2.6 | Editing a saved likeness whose region already equals the whole photo; press **Use the whole photo** | `isDepictionRedrawn` stays false; Save keeps the saved image |

**US-3.** As a keyboard or VoiceOver user who presses **Draw on the file**, I land on **Cancel** and hear that drawing has started, so that I can go back with Space or Escape.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Press **Draw on the file** from the keyboard | The drawing band appears and keyboard focus is on its **Cancel** button (full keyboard access on) |
| AC-3.2 | Same, with VoiceOver on | VoiceOver announces "Draw the region with the tools above. Escape goes back to the observation." |
| AC-3.3 | Focus on **Cancel**; press Space, or press Escape | Drawing ends and the dialog returns with the region unchanged (existing behavior) |

## 3. Scope

| In | Covered by |
| --- | --- |
| **Use the whole photo** / **Use the whole page** button in the Likeness value area (one label per viewer kind) | STEP-2, STEP-3, AC-1.1–AC-1.6 |
| **Use the citation's region** button, shown only when the Citation has a valid region | STEP-2, STEP-3, AC-2.1–AC-2.6 |
| Both buttons go through `commitViewerRegion`, so the PDF page rule and the page stamp apply as for a drawn region | STEP-2, AC-1.3, AC-1.4, AC-2.2 |
| Both buttons follow Draw's enable rule (`canStartDepictionDrawing`) | STEP-2, AC-1.5, AC-1.6 |
| Draw moves focus to the band's Cancel and posts a VoiceOver announcement | STEP-4, AC-3.1–AC-3.3 |
| A Go test row proving a whole-file rectangle passes the depiction region check | STEP-1, AC-1.7 |
| Three L10n keys and their String Catalog entries | STEP-3 |
| Spike 10 decisions note records the CDR-13 amendment | STEP-6 |

| Out | Home |
| --- | --- |
| Keyboard drawing on the canvas (arrow keys move or resize a rectangle) | not planned (user chose the region buttons, 2026-10-18) |
| Disabling Set page, the locator chip × buttons, and Find while drawing; hiding the locator draw hint (audit UX-5) | audit action list, Low |
| Blocking Draw and the two buttons on a file the viewer cannot cut, and showing a failed cut before Save (audit UX-6) | audit action list, Low |
| Setting a PDF Citation's page from the needs-page Callout (audit UX-7) | audit action list, Low |
| Where focus lands when the dialog returns from drawing | dogfood log ([`docs/dogfood/ux.md`](../../../dogfood/ux.md)) |
| Rendering the preview off the main actor (audit PRF-1) | audit action list, Medium |
| Any change to Go validation or the save RPCs | not planned (AC-1.7 shows the region already passes) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Value area | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | A `VStack`: an optional `PVCallout` (needs page / region hidden), then an `HStack` of a 96pt `PVThumbnail` preview and a `VStack` with the caption `Text` and one `PVButton` (**Draw on the file** / **Draw again**, `.secondary`, `.sm`, icon `.crop`), `.disabled(!model.canStartDepictionDrawing)`, `.focused(focus, equals: .value)`, identifier `citationComposer.depiction.draw`. It is used by `CitationPropertyFields` (same file, `case PropertyValueType.depiction.rawValue`) for both the add dialog and `CitationComposerEditSheet` |
| Wrapping row | `CitationComposerAuthoringBox.swift` `FlowLayout` (private, same file) | Lays out children left to right and wraps when the row runs out of width |
| Enable rule | `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift` `canStartDepictionDrawing`, `depictionDrawBlocked` | Draw needs a depiction draft and no transcription in flight and no `depictionNeedsPage` |
| Region commit | same file `commitViewerRegion(_:)` | With a depiction dialog open: refuses while transcribing; runs `depictionPageCheck` (`.ok` needs `artifactViewer.page == depictionPage`; `.needsPage` / `.wrongPage` refuse; `.stampViewerPage` sets the Citation's page from the viewer); writes the region to the draft through `CitationLocatorDraft.setRegion(_:capabilities:viewerPage:)`; clears `armedRegionTool`; restores the dialog when drawing |
| Page | same file `depictionPage`, `isExistingCitation` | Existing Citation: `fields.baseline.locator.page`; new Citation: `fields.values.locator.page` |
| Draw start | same file `startDepictionDrawing()` | Moves the viewer to `depictionPage` on a PDF, closes Find, arms `.rectangle`, then `show(.drawingDepiction(over:))`. No focus change, no announcement |
| Redraw flag | same file `isDepictionRedrawn` | True when the edit's `depictionDraft.locator` differs from `session.propertyBaseline.depictionDraft.locator`; it is derived, so any region write counts |
| Viewer page | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `setPage(_:)` | Synchronous; clamps to the page count |
| Region type | `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift` `ArtifactRegionDraft` (`kind`, `points`, `isValid`) | `isValid` is `ArtifactRegionGeometry.isValidPolygon`, which accepts points on 0 and 1 |
| Rectangle helper | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `rectanglePoints(from:to:)` | Returns the four clamped corners, top-left first, clockwise |
| Citation region | `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` `CitationLocatorDraft.region` | The Citation's region; `fields.baseline.locator` is the saved one, `fields.values.locator` the unsaved one |
| Drawing band | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `drawingBand` | An info or wrong-page `PVCallout` and a ghost **Cancel** `PVButton` with `.keyboardShortcut(.cancelAction)`, identifier `citationComposer.depiction.cancel`. The pane already holds a `@FocusState` (`findFieldFocused`) |
| Announcement precedent | `macos/App/Features/Names/NameValueEditorForm.swift` `moveAndAnnounce(at:by:)` | `NSAccessibility.post(element: NSApp as Any, notification: .announcementRequested, userInfo: [.announcement: message, .priority: NSAccessibilityPriorityLevel.medium])` |
| Copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.depictionDraw`, `depictionRedraw`, `depictionDrawingHint`, `depictionCancel` | "Draw on the file", "Draw again", "Draw the region with the tools above. Escape goes back to the observation.", "Cancel" |
| Go check | `core/locator/depiction.go` `ValidateDepictionRegion`; `core/locator/validate.go` `validateRegion` | Points must lie in [0,1] inclusive, the first must not repeat at the end, area ≥ 1e-12, no self-crossing; the pages must agree. The region need not lie inside the Citation's region |
| Tests | `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` `savingSendsAPNGAndDeletesIt`, `newPDFCitationTakesTheDrawnPage`, `existingPDFCitationKeepsItsPage`, `redrawSendsANewImage`, `editWithoutRedrawKeepsTheImage`, `drawCannotStartWhileTranscribing`; helpers `seeded(_:)`, `makeModel`, `seedCitation`, `openLikenessDraft`, `editingSavedLikeness`, `rect`, `regionJSON`, `draftValues`, `editValues`, `isMissing` | Cover drawing and save; nothing sets a region without the canvas |
| Go tests | `core/locator/depiction_test.go` `TestValidateDepictionRegion` | Table-driven; uses a triangle `tri`; no whole-file rectangle row |

**Precedent.** S10-34 ([#442](https://github.com/mendahu/provenencia/pull/442), `d711307`) added the value area, `commitViewerRegion`, and the depiction tests. Copy its model-extension style (one computed rule read by the view and the action) and its test shape (`seeded`, `openLikenessDraft`, then `#expect` on `draftValues`). The announcement copies `NameValueEditorForm.moveAndAnnounce`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Two buttons: **Use the whole photo** / **Use the whole page**, and **Use the citation's region**. Drawing stays pointer-only. | User decision 2026-10-18 (map working rule) | Keyboard drawing on the canvas — the user chose the buttons |
| D-2 | No Claude Design brief; §6 describes the buttons in words. | User decision 2026-10-18 | A brief and board — waived |
| D-3 | The whole-file label is two L10n keys chosen by `depictionCapabilities.supportsPageLocator`: page on a PDF, photo otherwise. | A PDF has pages and a photo does not; one key with a placeholder word does not translate | One key "Use the whole file" — "file" is vaguer than CDR-1's "photo" and the page rule makes the PDF case a page |
| D-4 | Both buttons call one model entry, `useDepictionRegion(_:)` *(new)*, which checks `canStartDepictionDrawing`, moves the viewer to `depictionPage` on a PDF when there is one (as `startDepictionDrawing` does), and calls `commitViewerRegion(_:)`. | One path for every region a depiction gets, so the page rule, the page stamp, and the transcription guard stay in one place | Writing `draft.locator` directly from the button — bypasses `depictionPageCheck` and the page stamp |
| D-5 | The whole-file region is `ArtifactRegionDraft(kind: .rectangle, points: ArtifactRegionGeometry.rectanglePoints(from: .zero, to: CGPoint(x: 1, y: 1)))`, held as `depictionWholeFileRegion` *(new, static let)* in `CitationComposerModel+Depiction.swift`. | It is a valid polygon in Swift and in Go (AC-1.7); `rectanglePoints` is the existing builder | A new helper on `ArtifactRegionGeometry` — one caller, and the geometry type does not know about depictions |
| D-6 | The Citation's region is `depictionCitationRegion` *(new)*: `fields.baseline.locator.region` on an existing Citation, `fields.values.locator.region` on a new one, and nil unless `isValid`. | Mirrors `depictionPage`, so the region and the page come from the same locator and pass `depictionPageCheck` | Always `fields.values` — on an existing Citation the save checks against the saved locator, and the page rule already reads the baseline |
| D-7 | **Use the citation's region** is absent when `depictionCitationRegion` is nil; both buttons are `.disabled(!model.canStartDepictionDrawing)` otherwise. | A button that can never work is noise; disabled-with-Draw keeps one rule (AC-1.5, AC-1.6) | Showing it disabled with no region — a dead control with no way to enable it from the dialog |
| D-8 | A button replaces any region the draft has. On an edit, the derived `isDepictionRedrawn` decides whether Save sends a new image; no new flag. | `isDepictionRedrawn` already compares locators (AC-2.5, AC-2.6) | A confirm before replacing — the old region is one Cancel (Discard) away, as with Draw again |
| D-9 | Layout: the caption `Text` stays; under it the three buttons sit in the file's `FlowLayout(spacing: PVSpacing.space4, lineSpacing: PVSpacing.space4)` in this order: Draw (unchanged), whole file, citation's region. The two new buttons are `PVButton`, `.secondary`, `.sm`, no icon. | `FlowLayout` already wraps chip rows in this dialog; same variant and size as Draw keeps one visual weight; the dialog is narrow enough that three buttons wrap | A new kit component — two buttons need none; icons — no kit symbol says "whole page" or "citation region", and a wrong glyph misleads |
| D-10 | Draw keeps `.focused(focus, equals: .value)`; the two new buttons take no `CitationAuthoringFocus` case and are reached with Tab after Draw. `propertyValueTakesFocus` is unchanged. | Draw is the value area's first control and the Tab claim already lands there; the buttons are disabled exactly when Draw is, so the skip rule stays right | New focus cases — `CitationAuthoringNavigation` would need new destinations with no new behavior |
| D-11 | When the drawing band appears, the viewer pane sets a new `@FocusState private var drawingCancelFocused: Bool` *(new)* to true on the band's **Cancel** and posts `.announcementRequested` with `L10n.string(L10n.CitationComposer.depictionDrawingHint)` at `.medium` priority, from `.onAppear` of `drawingBand`. | The band is the visible element of drawing mode; the existing hint already says what to do and how to leave | A new announcement string — the hint already exists and says Escape; posting from the model — the model does not import AppKit accessibility for any other copy |
| D-12 | Copy uses the straight apostrophe, as every other `citationComposer.depiction*` string does. | Consistency inside the namespace | The curly apostrophe — mixing inside one dialog |
| D-13 | Accessibility labels are the button titles; identifiers `citationComposer.depiction.useWhole` *(new)* and `citationComposer.depiction.useCitationRegion` *(new)*. | `PVButton` reads its title; the titles say the action | Extra hints — the titles are full sentences of what happens |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CitationComposerModel.useDepictionRegion(_:)` *(new)* | func | `commitViewerRegion(_:)`, `startDepictionDrawing()` in `CitationComposerModel+Depiction.swift` | It calls `commitViewerRegion`; it only adds the page move `startDepictionDrawing` does and the enable guard |
| `CitationComposerModel.depictionWholeFileRegion` *(new)* | static let | `ArtifactRegionGeometry.rectanglePoints(from:to:)`, `ArtifactRegionGeometry.unitSquare` | Built from `rectanglePoints`; searched `unitSquare`, `wholePage`, `x: 1, y: 1` in `Features/ArtifactViewer` and `Features/CitationComposer` and found no region constant |
| `CitationComposerModel.depictionCitationRegion` *(new)* | var | `depictionPage` (same file), `viewerRegion` | `viewerRegion` returns the depiction draft's region while the dialog is open; this reads the Citation's region the way `depictionPage` reads its page |
| `CitationComposerViewerPane.drawingCancelFocused` *(new)* | `@FocusState` | `findFieldFocused` in the same view | A second field of the same kind for a different control |
| `L10n.CitationComposer.depictionUseWholePhoto`, `depictionUseWholePage`, `depictionUseCitationRegion` *(new)* | key | `depictionDraw`, `depictionRedraw` | New actions need new words |
| Tests in `CitationComposerDepictionTests` *(new cases)* | test | the file's helpers | Same suite and helpers |

**Builds on.** `commitViewerRegion` and `depictionPageCheck` (the one region rule), `FlowLayout`, `PVButton`, the `NameValueEditorForm` announcement call, the `add-localized-string` key pattern.
**Deletes.** None — nothing becomes dead.
**Next adds.** The audit's UX-6 (block Draw on files that cannot be cut) changes `canStartDepictionDrawing` once and the two buttons follow it. A second depiction key (signature, a place photo) uses the same value area and gets both buttons for free.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/locator` | No code change. `TestValidateDepictionRegion` gains a row: a whole-file rectangle with the Citation's page passes. |

### FFI

None — the region goes over the existing observation save calls as region JSON.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Whole-file button | `PVButton` (`.secondary`, `.sm`) | `CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | Title `depictionUseWholePage` when `model.depictionCapabilities.supportsPageLocator`, else `depictionUseWholePhoto`. Action `model.useDepictionRegion(CitationComposerModel.depictionWholeFileRegion)`. `.disabled(!model.canStartDepictionDrawing)`. Identifier `citationComposer.depiction.useWhole` |
| Citation-region button | `PVButton` (`.secondary`, `.sm`) | same | Shown only when `model.depictionCitationRegion != nil`. Action `model.useDepictionRegion(region)`. Same disabled rule. Identifier `citationComposer.depiction.useCitationRegion` |
| Button row | `FlowLayout` | same file | Replaces the single Draw button in the caption `VStack`: Draw, whole file, citation's region, in that order |
| Drawing band Cancel | `PVButton` (existing) | `CitationComposerViewerPane.swift` `drawingBand` | `.focused($drawingCancelFocused)`; `.onAppear` of the band sets `drawingCancelFocused = true` on the next main-queue turn and posts the announcement |

Frame in words (no board): the Likeness value area keeps its 96pt preview on the left. On the right, the italic caption, then a row of three small secondary buttons: **Draw on the file** (crop icon), **Use the whole photo** (or **Use the whole page** on a PDF), **Use the citation's region** (only when the Citation has a region). When the row is too narrow the third, then the second, button wraps to a new line under Draw, with `space4` between lines.

States: no region → the empty caption (`depictionCaptionEmpty`), all buttons enabled; region set → the preview shows the cut, the caption reads the box caption, Draw reads **Draw again**; needs page → the warning Callout shows and all three buttons are disabled; transcribing → all three disabled. Large file: the preview render is unchanged (audit PRF-1 is Out). Accessibility: each button reads its title; Tab moves Draw → whole file → citation's region; drawing mode moves focus to Cancel and is announced.

### Strings

L10n keys *(new)*, in `L10n.CitationComposer` beside `depictionRedraw`, each with a String Catalog entry in `macos/App/Resources/Localizable.xcstrings`:

- `L10n.CitationComposer.depictionUseWholePhoto` = "Use the whole photo" — comment: "Button that sets a depiction's region to the whole photo"
- `L10n.CitationComposer.depictionUseWholePage` = "Use the whole page" — comment: "Button that sets a depiction's region to the whole PDF page"
- `L10n.CitationComposer.depictionUseCitationRegion` = "Use the citation's region" — comment: "Button that sets a depiction's region to the citation's own region"

Removed: none.

## 7. Steps

**STEP-1 — Prove the whole-file rectangle passes Go**
Files: `core/locator/depiction_test.go`. Rule: `.cursor/rules/go-tests.mdc`.
Do: add a `whole` constant beside `tri`: `{"type":"region","unit":"normalized","points":[{"x":0,"y":0},{"x":1,"y":0},{"x":1,"y":1},{"x":0,"y":1}]}`. Add two rows to `TestValidateDepictionRegion`: `"whole file, page 1 both"` → `doc(page("1"), whole)` with `doc(page("1"))`, no error; `"whole file, no page"` → `doc(whole)` with `doc(art)`, no error.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/locator/` passes.

**STEP-2 — Model: one entry for a region the dialog sets**
Files: `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`.
Do: add `static let depictionWholeFileRegion` (D-5), `var depictionCitationRegion: ArtifactRegionDraft?` (D-6), and `@discardableResult func useDepictionRegion(_ region: ArtifactRegionDraft) -> Bool` (D-4): return false unless `canStartDepictionDrawing`; when `depictionCapabilities.supportsPageLocator` and `depictionPage` is set, call `artifactViewer.setPage(page)`; return `commitViewerRegion(region)`. Doc comments in the file's style.
Check: the app target builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS'`).

**STEP-3 — View: the two buttons and their copy**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-localized-string`, `add-ui-component`.
Do: add the three keys (§6 Strings). In `CitationDepictionValueArea`, replace the lone Draw button with the `FlowLayout` row of §6 macOS (D-9, D-7, D-13). Draw keeps its modifiers.
Check: `python3 scripts/check-localizable-xcstrings.py` prints OK; the app target builds.

**STEP-4 — Drawing band: focus Cancel and announce**
Files: `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`.
Do: per D-11, add `@FocusState private var drawingCancelFocused: Bool`, put `.focused($drawingCancelFocused)` on the band's Cancel, and on the band's `.onAppear` run `DispatchQueue.main.async { drawingCancelFocused = true }` and the `NSAccessibility.post` call copied from `NameValueEditorForm.moveAndAnnounce` with `L10n.string(L10n.CitationComposer.depictionDrawingHint)`. Add `import AppKit` only if the file does not compile without it.
Check: the app target builds; manual check step 6.

**STEP-5 — Swift tests**
Files: `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`. Skill: `add-swift-test`.
Do: add T-2 to T-9 (§8) with the file's existing helpers. Give `editingSavedLikeness` a `region: String` parameter whose default is today's `regionJSON(rect(0.1, 0.1, 0.4, 0.4))`, so existing callers are unchanged. Compare copy through `L10n.string(…)`, not raw English.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CitationComposerDepictionTests` passes.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-10-followups/deployment-plan.md` (map row), this plan (status and Landed note), `docs/deployment-plan/archive/spike-10/README.md`.
Do: per the Definition of done. In the Spike 10 decisions note, add one bullet after the depiction bullet: "**A likeness takes its region from the keyboard too.** Drawing stays a pointer action; **Use the whole photo** (or page) and **Use the citation's region** set the region without drawing, and Draw moves focus to Cancel and announces drawing mode."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-63-keyboard-likeness-region.md` reports no errors; the links in the edited files resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/locator/depiction_test.go` `TestValidateDepictionRegion` (two new rows) | whole rectangle on page 1 with Citation page 1 → nil; whole rectangle, no page, artifact-only Citation → nil | AC-1.7, D-5 |
| T-2 | Swift | `CitationComposerDepictionTests` `wholePhotoSetsTheRegionAndSaves` *(new)* | `seeded()`, new Likeness draft, `useDepictionRegion(CitationComposerModel.depictionWholeFileRegion)` → true; `draftValues` region points equal `rect(0, 0, 1, 1).points`, kind `.rectangle`, `presented == .authoring`; `saveAuthoring()` → last `addObservationsToCitation` call ends with ` depiction=1 depictionFile=present`, saved region JSON has no `page` selector, `isMissing(store.recordedDepictionPaths)` | AC-1.1, AC-1.2 |
| T-3 | Swift | `CitationComposerDepictionTests` `wholePageMovesToTheCitationsPage` *(new)* | `seeded(.pdf)`, `seedCitation` with page 2, edit entry, viewer on page 1, new Likeness draft, whole region → viewer page 2, draft page 2, region whole | AC-1.3, D-4 |
| T-4 | Swift | `CitationComposerDepictionTests` `wholePageStampsANewCitationsPage` *(new)* | `seeded(.pdf)`, add entry, `artifactViewer.setPage(2)`, whole region → `model.locator.page == 2`, draft page 2 | AC-1.4 |
| T-5 | Swift | `CitationComposerDepictionTests` `regionButtonsFollowDraw` *(new)* | existing PDF Citation with no page → `useDepictionRegion` returns false and the draft has no region; image Citation with `model.transcriptions.isTranscribing = true` (as `drawCannotStartWhileTranscribing` sets it) → returns false | AC-1.5, AC-1.6, D-7 |
| T-6 | Swift | `CitationComposerDepictionTests` `citationRegionIsCopied` *(new)* | `seedCitation` with `regionJSON(rect(0.2, 0.2, 0.6, 0.7))`, edit entry, new Likeness draft → `depictionCitationRegion` equals that region; `useDepictionRegion` with it → draft region equals it | AC-2.1, D-6 |
| T-7 | Swift | `CitationComposerDepictionTests` `citationRegionOnItsPage` *(new)* | `seeded(.pdf)`, Citation page 1 with a region, viewer page 2 → after the call, viewer page 1 and draft region on page 1 | AC-2.2 |
| T-8 | Swift | `CitationComposerDepictionTests` `noCitationRegionMeansNone` *(new)* | artifact-only Citation → `depictionCitationRegion == nil`; new Citation after `setRegion(rect(0.1, 0.1, 0.5, 0.5))` then a Likeness draft → equals that region | AC-2.3, AC-2.4 |
| T-9 | Swift | `CitationComposerDepictionTests` `regionButtonOnAnEditIsARedraw` *(new)* | `editingSavedLikeness()`, `useDepictionRegion(rect(0.5, 0.5, 0.9, 0.9))` → `isDepictionRedrawn`, the file's `caption(model)` helper returns `L10n.string(L10n.CitationComposer.depictionCaptionRedrawn)`, `saveEdit()` sends an `updateObservation` call ending ` depiction=1 depictionFile=present`; `editingSavedLikeness(region: regionJSON(rect(0, 0, 1, 1)))` then the whole region again → `isDepictionRedrawn == false`. STEP-5 adds the `region:` parameter to `editingSavedLikeness`, defaulting to today's `regionJSON(rect(0.1, 0.1, 0.4, 0.4))` | AC-2.5, AC-2.6, D-8 |

AC-3.1, AC-3.2 and AC-3.3 are view-level focus and VoiceOver behavior: manual check steps 6 and 7.

Must stay green: everything `scripts/check-pr.sh` runs. The suite most likely to break: `CitationComposerDepictionTests` (`tabFromPropertySkipsADisabledDraw`, `drawingDisablesCitationControlsAndLeaveReturns`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches §6 macOS (brief waived); no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] spike-10-followups/deployment-plan.md: S10-63 row status → Landed #N
- [ ] Brief: none (waived)
- [ ] Domain doc: archive/spike-10/README.md gains the CDR-13 amendment bullet (STEP-6)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with System Settings → Keyboard → Keyboard navigation on.

1. Open a new project. Add a Source with one photo (a JPEG or PNG portrait) and one 2-page PDF. Add a Subject Ada (person) on it.
2. Open the composer on the photo. Add a Likeness observation for Ada without touching the mouse: Tab to **Use the whole photo**, press Space. See the preview fill with the whole photo and Save enable. Save. The saved line shows a 32pt thumbnail.
3. Draw a Citation region around one face on the photo, save the Citation. Add another Likeness for Ada; **Use the citation's region** appears; press it. The preview shows that face. Save.
4. Edit the second likeness; press **Use the whole photo**. The caption reads "Redrawn. Saving replaces the saved image". Save; the thumbnail shows the whole photo.
5. Open the PDF on page 2, save the Citation with page 2, move the viewer to page 1. Add a Likeness: the button reads **Use the whole page**; press it. The viewer moves to page 2 and the preview shows page 2.
6. In a Likeness dialog press **Draw on the file** with Space. The drawing band appears and the focus ring is on **Cancel**. Press Space: drawing ends and the dialog returns. Press Draw again, then Escape: the same.
7. Turn on VoiceOver. Press **Draw on the file**. VoiceOver says "Draw the region with the tools above. Escape goes back to the observation." Press Escape.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Focus does not land on Cancel because the sheet is still dismissing when the band appears | D-11 sets focus on the next main-queue turn, as `placeOpeningFocusIfNeeded` does |
| Three buttons overflow the dialog width | `FlowLayout` wraps them (D-9) |
| The whole-page preview of a large PDF page renders slowly on the main thread | Same cost as drawing a page-sized box today; audit PRF-1 is Out |

**Stop and report if:**

- `commitViewerRegion(_:)` or `depictionPageCheck` no longer behaves as §4 says.
- A whole-file rectangle fails `ArtifactRegionGeometry.isValidPolygon` or `locator.ValidateDepictionRegion` (T-1 fails without a code change in `core/locator`).
- `CitationDepictionValueArea` is no longer the value area for both the add dialog and `CitationComposerEditSheet`.
- Focus on Cancel needs a change to `PVButton` or another kit component.
- The save of a button-set region on an existing Citation is refused by Go with `locator.invalid` (the saved Citation locator and `depictionCitationRegion` disagree).

## 12. Open questions

None.

## 13. PR

**Title:** S10-63 — Keyboard region for a likeness
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-63-keyboard-likeness-region.md`.
