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
**Review:** [`S10-63-keyboard-likeness-region.review.md`](S10-63-keyboard-likeness-region.review.md) — Fits with changes / Ready with changes, 1 round

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

A researcher who works from the keyboard or with VoiceOver can save a likeness. Today a likeness's value is its region, the region can only be drawn with the pointer, and Save stays disabled until a region exists, so no keyboard or VoiceOver user can finish the Likeness observation (audit UX-1). This PR adds two keyboard-reachable buttons to the Likeness value area: one sets the region to the whole photo or the whole PDF page, and one copies the Citation's own region. Pressing **Draw on the file** also moves keyboard focus to the drawing band's **Cancel** and announces drawing mode, so a keyboard user who starts drawing by mistake learns how to go back. A region set by a button is announced too. The VoiceOver announcement call moves into one small Platform helper, which the name editor's part move already needs.

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
| AC-1.8 | A PDF Citation; then an image Citation | `depictionWholeFileTitle` is `depictionUseWholePage` on the PDF and `depictionUseWholePhoto` on the image |
| AC-1.9 | VoiceOver on; press **Use the whole photo** on an image Citation | VoiceOver announces the new caption, "The rectangle around the region, up to 2,048 px on its long edge" |
| AC-1.7 | The Go region check receives a whole-file rectangle `[(0,0),(1,0),(1,1),(0,1)]` with the Citation's page | `locator.ValidateDepictionRegion` returns nil |

**US-2.** As a researcher whose Citation is already drawn around one face, I press **Use the citation's region** so that the likeness takes that region without drawing it again.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | An existing image Citation saved with the region (0.2,0.2)–(0.6,0.7); new Likeness draft; press **Use the citation's region** | The draft's region has the same kind and points as the Citation's saved region; Save is enabled |
| AC-2.2 | An existing PDF Citation saved on page 1 with a region; viewer on page 2; press the button | The viewer moves to page 1; the draft's region is the Citation's region on page 1 |
| AC-2.3 | A Citation with no region (artifact only, or page only) | **Use the citation's region** is not shown; **Use the whole photo** / **Use the whole page** is |
| AC-2.4 | A new, unsaved Citation where the researcher drew a region before opening the Likeness draft | The button uses that unsaved region |
| AC-2.7 | An existing Citation saved with region R1; the researcher redraws the Citation region to R2 without saving the Citation; then opens a Likeness draft | `depictionCitationRegion` is R1, the saved region, because the save checks the depiction against the saved locator (D-6) |
| AC-2.5 | Editing a saved likeness with region (0.1,0.1)–(0.4,0.4); press **Use the citation's region** (a different region) | `isDepictionRedrawn` is true; the caption reads the Redrawn caption; Save sends a new image (`depiction=1 depictionFile=present`) |
| AC-2.6 | Editing a saved likeness whose region already equals the whole photo; press **Use the whole photo** | `isDepictionRedrawn` stays false; Save keeps the saved image |

**US-3.** As a keyboard or VoiceOver user who presses **Draw on the file**, I land on **Cancel** and hear that drawing has started, so that I can go back with Space or Escape.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Press **Draw on the file** from the keyboard | The drawing band appears and keyboard focus is on its **Cancel** button (full keyboard access on) |
| AC-3.2 | Same, with VoiceOver on | VoiceOver announces "Draw the region with the tools above. Escape goes back to the observation." |
| AC-3.3 | Focus on **Cancel**; press Space, or press Escape | Drawing ends and the dialog returns with the region unchanged (existing behavior) |
| AC-3.4 | `NameValueEditorForm` moves a name part with its move buttons (VoiceOver on) | The part-moved announcement still plays, now through `AccessibilityAnnouncement.post` |

## 3. Scope

| In | Covered by |
| --- | --- |
| **Use the whole photo** / **Use the whole page** button in the Likeness value area (one label per viewer kind) | STEP-2, STEP-3, AC-1.1–AC-1.6 |
| **Use the citation's region** button, shown only when the Citation has a valid region | STEP-2, STEP-3, AC-2.1–AC-2.6 |
| Both buttons go through `commitViewerRegion`, so the PDF page rule and the page stamp apply as for a drawn region | STEP-2, AC-1.3, AC-1.4, AC-2.2 |
| Both buttons follow Draw's enable rule (`canStartDepictionDrawing`) | STEP-2, AC-1.5, AC-1.6 |
| Draw moves focus to the band's Cancel and posts a VoiceOver announcement | STEP-4, AC-3.1–AC-3.3 |
| A button press that sets a region announces the new caption | STEP-3, AC-1.9 |
| One announcement helper, `AccessibilityAnnouncement.post` in `macos/App/Platform/AccessibilityAnnouncement.swift` *(new)*; `NameValueEditorForm.moveAndAnnounce` moves onto it | STEP-2, AC-3.4 |
| Draw and the buttons share one preamble (`prepareDepictionRegion`) and one Citation-locator choice (`depictionCitationLocator`) | STEP-2, AC-1.3, AC-2.2, AC-2.7 |
| `ArtifactRegionDraft.whole` *(new)*, the unit-square rectangle on the region type | STEP-2, AC-1.1 |
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
| Feedback when a button press is refused (viewer still loading, wrong page) beyond the existing caption and Callout | audit action list, Low (UX-6) |
| Rewording the empty caption "Nothing drawn yet. Draw the region on the file" now that two buttons sit beside it | audit action list, Low (copy and label nits, UX-12/UX-13) |
| Renaming `canStartDepictionDrawing` / `depictionDrawBlocked` (D-14 rewrites the doc comment only) | not planned |
| Moving the test files' hand-written rectangle builders (`CitationComposerDepictionTests.rect`, `ArtifactRegionCropTests.rectangle`, `CitationAutoTranscribeTests`) onto a shared builder | not planned (logged in the review log, M-5) |
| `CitationComposerModel.identityAnnouncement`, read as an accessibility hint rather than posted | not planned (logged in the review log, M-3) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Value area | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | A `VStack`: an optional `PVCallout` (needs page / region hidden), then an `HStack` of a 96pt `PVThumbnail` preview and a `VStack` with the caption `Text` and one `PVButton` (**Draw on the file** / **Draw again**, `.secondary`, `.sm`, icon `.crop`), `.disabled(!model.canStartDepictionDrawing)`, `.focused(focus, equals: .value)`, identifier `citationComposer.depiction.draw`. It is used by `CitationPropertyFields` (same file, `case PropertyValueType.depiction.rawValue`) for both the add dialog and `CitationComposerEditSheet` |
| Wrapping row | `CitationComposerAuthoringBox.swift` `FlowLayout` (private, same file) | Lays out children left to right and wraps when the row runs out of width |
| Enable rule | `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift` `canStartDepictionDrawing`, `depictionDrawBlocked` | Draw needs a depiction draft and no transcription in flight and no `depictionNeedsPage` |
| Region commit | same file `commitViewerRegion(_:)` | With a depiction dialog open: refuses while transcribing; runs `depictionPageCheck` (`.ok` needs `artifactViewer.page == depictionPage`; `.needsPage` / `.wrongPage` refuse; `.stampViewerPage` sets the Citation's page from the viewer); writes the region to the draft through `CitationLocatorDraft.setRegion(_:capabilities:viewerPage:)`; clears `armedRegionTool`; restores the dialog when drawing |
| Page | same file `depictionPage`, `isExistingCitation` | Existing Citation: `fields.baseline.locator.page`; new Citation: `fields.values.locator.page` |
| Draw start | same file `startDepictionDrawing()` | Guards `canStartDepictionDrawing`, takes `dialogBase()` and returns unless it is `.authoring` or `.edit`, moves the viewer to `depictionPage` on a PDF, closes Find, arms `.rectangle`, then `show(.drawingDepiction(over: base))`. No focus change, no announcement |
| Draw rule comment | same file `canStartDepictionDrawing` doc comment | "The one rule for starting to draw: the Draw button and ``startDepictionDrawing()`` read it, and the Tab order reads ``depictionDrawBlocked``." Readers: `CitationComposerAuthoringBox.swift` (Draw's `.disabled`, the Tab skip on `depictionDrawBlocked`), `startDepictionDrawing`, `CitationComposerDepictionTests` |
| Redraw flag | same file `isDepictionRedrawn` | True when the edit's `depictionDraft.locator` differs from `session.propertyBaseline.depictionDraft.locator`; it is derived, so any region write counts |
| Viewer page | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `setPage(_:)` | Synchronous; clamps to the page count |
| Region type | `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift` `ArtifactRegionDraft` (`kind`, `points`, `isValid`) | `isValid` is `ArtifactRegionGeometry.isValidPolygon`, which accepts points on 0 and 1 |
| Rectangle helper | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `rectanglePoints(from:to:)` | Returns the four clamped corners, top-left first, clockwise |
| Unit square | same file `extension CGRect` `static let unitSquare` | `CGRect(x: 0, y: 0, width: 1, height: 1)`, the default `imageRect` of `handlePoints`; a rectangle, not a region draft |
| Citation region | `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` `CitationLocatorDraft.region` | The Citation's region; `fields.baseline.locator` is the saved one, `fields.values.locator` the unsaved one |
| Drawing band | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `drawingBand` | An info or wrong-page `PVCallout` and a ghost **Cancel** `PVButton` with `.keyboardShortcut(.cancelAction)`, identifier `citationComposer.depiction.cancel`. The pane already holds a `@FocusState` (`findFieldFocused`) |
| Announcement | `macos/App/Features/Names/NameValueEditorForm.swift` `moveAndAnnounce(at:by:)` | The only `.announcementRequested` post in the app: `NSAccessibility.post(element: NSApp as Any, notification: .announcementRequested, userInfo: [.announcement: message, .priority: NSAccessibilityPriorityLevel.medium])` |
| Focus double hop | `CitationComposerAuthoringBox.swift` `claimTab(from:committed:via:bridgeHasTerm:)` | Writes the focus binding on the next main-queue turn, then on the turn after that writes it again when it did not land ("a deep form can miss the first write") |
| Platform files | `macos/App/Platform/ProjectFiles.swift` and its four `project.pbxproj` objects (PBXBuildFile, PBXFileReference, the Platform group child, the app Sources phase entry) | The shape a new Platform file copies; the project uses no synchronized folders |
| Wrapping row | `CitationComposerAuthoringBox.swift` `FlowLayout(spacing: PVSpacing.space4, lineSpacing: PVSpacing.space3)` | The file's one other `FlowLayout` call (the composed name pieces) |
| Copy | `macos/App/Platform/L10n.swift` `L10n.CitationComposer.depictionDraw`, `depictionRedraw`, `depictionDrawingHint`, `depictionCancel` | "Draw on the file", "Draw again", "Draw the region with the tools above. Escape goes back to the observation.", "Cancel" |
| Go check | `core/locator/depiction.go` `ValidateDepictionRegion`; `core/locator/validate.go` `validateRegion` | Points must lie in [0,1] inclusive, the first must not repeat at the end, area ≥ 1e-12, no self-crossing; the pages must agree. The region need not lie inside the Citation's region |
| Tests | `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` `savingSendsAPNGAndDeletesIt`, `newPDFCitationTakesTheDrawnPage`, `existingPDFCitationKeepsItsPage`, `redrawSendsANewImage`, `editWithoutRedrawKeepsTheImage`, `drawCannotStartWhileTranscribing`; helpers `seeded(_:)`, `makeModel`, `seedCitation(locatorJSON:)`, `openLikenessDraft`, `editingSavedLikeness`, `rect`, `regionJSON`, `draftValues`, `editValues`, `isMissing`, `caption(_:)` | Cover drawing and save; nothing sets a region without the canvas |
| Go tests | `core/locator/depiction_test.go` `TestValidateDepictionRegion` | Table-driven; uses a triangle `tri`; no whole-file rectangle row |

**Precedent.** S10-34 ([#442](https://github.com/mendahu/provenencia/pull/442), `d711307`) added the value area, `commitViewerRegion`, and the depiction tests. Copy its model-extension style (one computed rule read by the view and the action) and its test shape (`seeded`, `openLikenessDraft`, then `#expect` on `draftValues`). The announcement call moves out of `NameValueEditorForm.moveAndAnnounce` into the Platform helper; the focus double hop copies `claimTab`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Two buttons: **Use the whole photo** / **Use the whole page**, and **Use the citation's region**. Drawing stays pointer-only. | User decision 2026-10-18 (map working rule) | Keyboard drawing on the canvas — the user chose the buttons |
| D-2 | No Claude Design brief; §6 describes the buttons in words. | User decision 2026-10-18 | A brief and board — waived |
| D-3 | The whole-file label is two L10n keys chosen by `depictionCapabilities.supportsPageLocator`: page on a PDF, photo otherwise. The model exposes the choice as `var depictionWholeFileTitle: LocalizedStringResource` *(new)* so a test reads it (AC-1.8). | A PDF has pages and a photo does not; one key with a placeholder word does not translate | One key "Use the whole file" — "file" is vaguer than CDR-1's "photo" and the page rule makes the PDF case a page |
| D-4 | Draw and both buttons share one private preamble, `prepareDepictionRegion() -> ComposerBase?` *(new)*: it returns nil unless `canStartDepictionDrawing` and `dialogBase()` is `.authoring` or `.edit`, moves the viewer to `depictionPage` on a PDF when there is one, and returns the base. `startDepictionDrawing()` becomes `guard let base = prepareDepictionRegion() else { return }` followed by its existing Find, tool, and `show` lines. Both buttons call `useDepictionRegion(_:)` *(new)*: `prepareDepictionRegion() != nil && commitViewerRegion(region)`. | One rule for "a depiction region is set on the Citation's page, and only when Draw could start"; audit UX-6 then changes one function and the buttons cannot miss it | Writing `draft.locator` directly from the button — bypasses `depictionPageCheck` and the page stamp; copying Draw's guard and page move into `useDepictionRegion` — two copies in one file |
| D-5 | The whole-file region is `ArtifactRegionDraft.whole` *(new, static let)* in `ArtifactRegionTool.swift`, beside the type: `ArtifactRegionDraft(kind: .rectangle, points: ArtifactRegionGeometry.rectanglePoints(from: CGPoint(x: CGRect.unitSquare.minX, y: CGRect.unitSquare.minY), to: CGPoint(x: CGRect.unitSquare.maxX, y: CGRect.unitSquare.maxY)))`. | A unit-square rectangle says nothing about depictions; region values live on the region type, built from the square the coordinates are defined on. It is valid in Swift and in Go (AC-1.7) | A static on the feature model — a region constant on a 1,000-line feature file; a literal (0,0)–(1,1) — restates `CGRect.unitSquare` |
| D-6 | One private `depictionCitationLocator: CitationLocatorDraft` *(new)* is `isExistingCitation ? fields.baseline.locator : fields.values.locator`. `depictionPage` becomes `depictionCitationLocator.page`, and `depictionCitationRegion` *(new)* is `depictionCitationLocator.region` when `isValid`, else nil. | The page and the region come from one locator by structure, so they cannot split and draw the Go refusal stop condition 5 names; an existing Citation's unsaved region is ignored because the save checks against the saved locator (AC-2.7) | Two ternaries that must stay in step; always `fields.values` — on an existing Citation the save checks the saved locator |
| D-7 | **Use the citation's region** is absent when `depictionCitationRegion` is nil; both buttons are `.disabled(!model.canStartDepictionDrawing)` otherwise. | A button that can never work is noise; disabled-with-Draw keeps one rule (AC-1.5, AC-1.6) | Showing it disabled with no region — a dead control with no way to enable it from the dialog |
| D-8 | A button replaces any region the draft has. On an edit, the derived `isDepictionRedrawn` decides whether Save sends a new image; no new flag. | `isDepictionRedrawn` already compares locators (AC-2.5, AC-2.6) | A confirm before replacing — the old region is one Cancel (Discard) away, as with Draw again |
| D-9 | Layout: the caption `Text` stays; under it the three buttons sit in the file's `FlowLayout(spacing: PVSpacing.space4, lineSpacing: PVSpacing.space3)` (the spacing of the file's other `FlowLayout` call) in this order: Draw (unchanged), whole file, citation's region. The two new buttons are `PVButton`, `.secondary`, `.sm`, no icon. | `FlowLayout` already wraps chip rows in this dialog; same variant and size as Draw keeps one visual weight; the dialog is narrow enough that three buttons wrap | A new kit component — two buttons need none; icons — no kit symbol says "whole page" or "citation region", and a wrong glyph misleads |
| D-10 | Draw keeps `.focused(focus, equals: .value)`; the two new buttons take no `CitationAuthoringFocus` case and are reached with Tab after Draw. `propertyValueTakesFocus` is unchanged. | Draw is the value area's first control and the Tab claim already lands there; the buttons are disabled exactly when Draw is, so the skip rule stays right | New focus cases — `CitationAuthoringNavigation` would need new destinations with no new behavior |
| D-11 | When the drawing band appears, the viewer pane focuses its **Cancel** through a new `@FocusState private var drawingCancelFocused: Bool` *(new)* with `claimTab`'s double hop: on the next main-queue turn it sets `drawingCancelFocused = true` and posts `AccessibilityAnnouncement.post(L10n.string(L10n.CitationComposer.depictionDrawingHint))`; on the turn after, it sets the focus again when `drawingCancelFocused` is still false. `.onDisappear` of the band sets `drawingCancelFocused = false`, so a second Draw starts clean. | The band is the visible element of drawing mode and the existing hint says what to do and how to leave; the sheet is still dismissing when the band appears, which is the case `claimTab`'s second write exists for; the announcement follows the first focus write so the focus move does not cut it off | A single `DispatchQueue.main.async` — the sheet dismissal can swallow it; a new announcement string — the hint already says Escape |
| D-14 | `canStartDepictionDrawing` keeps its name; its doc comment becomes: "The one rule for setting a depiction region: Draw, the two region buttons, and ``prepareDepictionRegion()`` read it, and the Tab order reads ``depictionDrawBlocked``." | Every way to set a region reads one rule, and the comment says so for the next reader (UX-6) | A rename to `canSetDepictionRegion` — about eight call sites and test lines for no behavior; the comment carries the meaning |
| D-15 | When `useDepictionRegion` returns true, the view announces the caption the press produced, `L10n.string(model.depictionCaption)`, through `AccessibilityAnnouncement.post`. When it returns false, nothing more happens: the disabled state, the Callout, and the caption already show why (audit UX-6 is Out). | A VoiceOver user otherwise hears nothing when the region is set; the caption already says what was set and whether Save replaces the image, so no new string | A new "Region set" string — duplicates the caption; failure copy — UX-6 |
| D-16 | One Platform helper, `enum AccessibilityAnnouncement` *(new)* in `macos/App/Platform/AccessibilityAnnouncement.swift`, with `@MainActor static func post(_ message: String, priority: NSAccessibilityPriorityLevel = .medium)` wrapping the `NSAccessibility.post(element: NSApp as Any, notification: .announcementRequested, …)` call. `NameValueEditorForm.moveAndAnnounce` calls it instead of its own copy. The file registers in `project.pbxproj` as `ProjectFiles.swift` does (four objects). | This PR adds the second and third callers; one place to fix the call (priority, a delay) | A third hand copy of the `userInfo` dictionary; posting from the model — announcements are view events here, and the helper keeps the model free of the AppKit call |
| D-12 | Copy uses the straight apostrophe, as every other `citationComposer.depiction*` string does. | Consistency inside the namespace | The curly apostrophe — mixing inside one dialog |
| D-13 | Accessibility labels are the button titles; identifiers `citationComposer.depiction.useWhole` *(new)* and `citationComposer.depiction.useCitationRegion` *(new)*. | `PVButton` reads its title; the titles say the action | Extra hints — the titles are full sentences of what happens |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CitationComposerModel.prepareDepictionRegion()` *(new, private)* | func | `startDepictionDrawing()` (same file) | It is `startDepictionDrawing`'s existing guard, base check, and page move, pulled out so the buttons share it (D-4) |
| `CitationComposerModel.useDepictionRegion(_:)` *(new)* | func | `commitViewerRegion(_:)` | `prepareDepictionRegion()` then `commitViewerRegion`; no rule of its own |
| `ArtifactRegionDraft.whole` *(new)* | static let | `CGRect.unitSquare`, `ArtifactRegionGeometry.rectanglePoints(from:to:)` | `CGRect.unitSquare` is a rectangle, not a region draft; searched `whole`, `wholePage`, `x: 1, y: 1` in `Features/ArtifactViewer` and `Features/CitationComposer` and found no region constant (only test builders) |
| `CitationComposerModel.depictionCitationLocator` *(new, private)* | var | `depictionPage` (same file) | `depictionPage` is rewritten onto it (D-6) |
| `CitationComposerModel.depictionCitationRegion` *(new)* | var | `viewerRegion` | `viewerRegion` returns the depiction draft's region while the dialog is open; this reads the Citation's |
| `CitationComposerModel.depictionWholeFileTitle` *(new)* | var | `depictionCaption` (same file) | Same shape: a resource the view reads and a test checks (D-3) |
| `AccessibilityAnnouncement.post(_:priority:)` *(new)* in `macos/App/Platform/AccessibilityAnnouncement.swift` *(new)* | func / file | `NameValueEditorForm.moveAndAnnounce` (the one `.announcementRequested` post); `CitationComposerModel.identityAnnouncement` (a hint, not a post) | `moveAndAnnounce` moves onto it; no kit or Platform helper posts announcements today (`rg announcementRequested macos/App`) |
| `CitationComposerViewerPane.drawingCancelFocused` *(new)* | `@FocusState` | `findFieldFocused` in the same view | A second field of the same kind for a different control |
| `L10n.CitationComposer.depictionUseWholePhoto`, `depictionUseWholePage`, `depictionUseCitationRegion` *(new)* | key | `depictionDraw`, `depictionRedraw` | New actions need new words |
| Tests in `CitationComposerDepictionTests` *(new cases)* | test | the file's helpers | Same suite and helpers |

**Builds on.** `commitViewerRegion` and `depictionPageCheck` (the one region rule), `FlowLayout`, `PVButton`, `claimTab`'s focus double hop, the `add-localized-string` key pattern.
**Deletes.** The inline `NSAccessibility.post` call in `NameValueEditorForm.moveAndAnnounce` (replaced by the helper) and the ternary in `depictionPage` (now read from `depictionCitationLocator`).
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
| Whole-file button | `PVButton` (`.secondary`, `.sm`) | `CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | Title `model.depictionWholeFileTitle`. Action: `if model.useDepictionRegion(.whole) { AccessibilityAnnouncement.post(L10n.string(model.depictionCaption)) }` (D-15). `.disabled(!model.canStartDepictionDrawing)`. Identifier `citationComposer.depiction.useWhole` |
| Citation-region button | `PVButton` (`.secondary`, `.sm`) | same | Shown only when `model.depictionCitationRegion != nil`. Action: the same announce-on-true call with `model.useDepictionRegion(region)`. Same disabled rule. Identifier `citationComposer.depiction.useCitationRegion` |
| Button row | `FlowLayout` | same file | Replaces the single Draw button in the caption `VStack`: Draw, whole file, citation's region, in that order |
| Drawing band Cancel | `PVButton` (existing) | `CitationComposerViewerPane.swift` `drawingBand` | `.focused($drawingCancelFocused)`; `.onAppear` of the band runs D-11's double hop and posts the announcement; `.onDisappear` clears the focus flag |

Frame in words (no board): the Likeness value area keeps its 96pt preview on the left. On the right, the italic caption, then a row of three small secondary buttons: **Draw on the file** (crop icon), **Use the whole photo** (or **Use the whole page** on a PDF), **Use the citation's region** (only when the Citation has a region). When the row is too narrow the third, then the second, button wraps to a new line under Draw, with `space3` between lines.

States: no region → the empty caption (`depictionCaptionEmpty`), all buttons enabled; region set → the preview shows the cut, the caption reads the box caption, Draw reads **Draw again**; needs page → the warning Callout shows and all three buttons are disabled; transcribing → all three disabled. Large file: the preview render is unchanged (audit PRF-1 is Out). Accessibility: each button reads its title; Tab moves Draw → whole file → citation's region; a press that sets the region announces the new caption; drawing mode moves focus to Cancel and is announced.

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

**STEP-2 — Model and Platform: one preamble, one locator, one announcement helper**
Files: `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`, `macos/App/Features/ArtifactViewer/ArtifactRegionTool.swift`, `macos/App/Platform/AccessibilityAnnouncement.swift` *(new)*, `macos/App/Features/Names/NameValueEditorForm.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-ui-component` (pbxproj line of its checklist).
Do:
- `ArtifactRegionTool.swift`: add `static let whole` on `ArtifactRegionDraft` (D-5).
- Model: add `private var depictionCitationLocator` and rewrite `depictionPage` onto it; add `var depictionCitationRegion: ArtifactRegionDraft?` (D-6); add `var depictionWholeFileTitle: LocalizedStringResource` (D-3); add `private func prepareDepictionRegion() -> ComposerBase?` and rewrite `startDepictionDrawing()` onto it; add `@discardableResult func useDepictionRegion(_ region: ArtifactRegionDraft) -> Bool` (D-4); rewrite the `canStartDepictionDrawing` doc comment (D-14). Doc comments in the file's style.
- Platform: add `AccessibilityAnnouncement.swift` (`import AppKit`) with D-16's `post`; register it in `project.pbxproj` with four new objects copied from `ProjectFiles.swift`'s (PBXBuildFile, PBXFileReference, Platform group child, app Sources phase), using unused `AA00000000000000000000xx` IDs.
- `NameValueEditorForm.moveAndAnnounce`: replace its `NSAccessibility.post` call with `AccessibilityAnnouncement.post(message)`.
Check: the app target builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`); the existing `CitationComposerDepictionTests` pass (`xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CitationComposerDepictionTests CODE_SIGNING_ALLOWED=NO`), proving Draw is unchanged.

**STEP-3 — View: the two buttons and their copy**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-localized-string`, `add-ui-component`.
Do: add the three keys (§6 Strings) and return them from `depictionWholeFileTitle`. In `CitationDepictionValueArea`, replace the lone Draw button with the `FlowLayout` row of §6 macOS (D-9, D-7, D-13, D-15). Draw keeps its modifiers.
Check: `python3 scripts/check-localizable-xcstrings.py` prints OK; the app target builds.

**STEP-4 — Drawing band: focus Cancel and announce**
Files: `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift`.
Do: per D-11, add `@FocusState private var drawingCancelFocused: Bool`, put `.focused($drawingCancelFocused)` on the band's Cancel, run the double hop and `AccessibilityAnnouncement.post(L10n.string(L10n.CitationComposer.depictionDrawingHint))` from the band's `.onAppear`, and clear the flag in `.onDisappear`. The file needs no new import: the AppKit call lives in the helper.
Check: the app target builds; manual check step 6.

**STEP-5 — Swift tests**
Files: `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`. Skill: `add-swift-test`.
Do: add T-2 to T-10 (§8) with the file's existing helpers. Give `editingSavedLikeness` a `region: String` parameter whose default is today's `regionJSON(rect(0.1, 0.1, 0.4, 0.4))`, so existing callers are unchanged. Build a Citation locator with `regionJSON(_:page:)` (nil rect for page only) and pass it to `seedCitation(locatorJSON:)`; an "edit entry" is `makeModel(citationID: "cit-1")` with no `observationID`, as `existingPDFCitationKeepsItsPage` does; move the viewer with `artifactViewer.setPage(_:)` after `prepare()`. Compare copy through `L10n.string(…)`, not raw English.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CitationComposerDepictionTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-6 — Docs**
Files: `docs/deployment-plan/archive/spike-10/README.md`.
Do: in the Spike 10 decisions note, add one bullet after the depiction bullet: "**A likeness takes its region from the keyboard too.** Drawing stays a pointer action; **Use the whole photo** (or page) and **Use the citation's region** set the region without drawing, and Draw moves focus to Cancel and announces drawing mode."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-63-keyboard-likeness-region.md` reports no errors; the links in the edited files resolve.

The map row (`docs/deployment-plan/spike-10-followups/deployment-plan.md`, S10-63 → `Landed #N`), this plan's Status (`Landed in #N (date)`), and its Landed note go in `implement-pr` step 8's commit "Record S10-63 as landed in pull request N." after the draft PR has a number; this folder has no `completed.md`.

Commits, in order: STEP-1 ("Prove a whole-photo rectangle passes the likeness region check."); STEP-2 ("Share one announcement helper and one depiction region preamble."); STEP-4, its one file staged alone ("Move focus to Cancel and announce when likeness drawing starts."); STEP-3 and STEP-5 together, staged once T-2 to T-10 pass ("Let a keyboard user give a likeness the whole photo or the citation's region."); STEP-6 ("Record that a likeness takes its region from the keyboard too."). Then the step-8 commit.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/locator/depiction_test.go` `TestValidateDepictionRegion` (two new rows) | whole rectangle on page 1 with Citation page 1 → nil; whole rectangle, no page, artifact-only Citation → nil | AC-1.7, D-5 |
| T-2 | Swift | `CitationComposerDepictionTests` `wholePhotoSetsTheRegionAndSaves` *(new)* | `seeded()`, new Likeness draft, `useDepictionRegion(.whole)` → true; `draftValues` region points equal `rect(0, 0, 1, 1).points`, kind `.rectangle`, `presented == .authoring`; `saveAuthoring()` → last `addObservationsToCitation` call ends with ` depiction=1 depictionFile=present`, saved region JSON has no `page` selector, `isMissing(store.recordedDepictionPaths)` | AC-1.1, AC-1.2, D-5 |
| T-3 | Swift | `CitationComposerDepictionTests` `wholePageMovesToTheCitationsPage` *(new)* | `seeded(.pdf)`, `seedCitation` with page 2, edit entry, viewer on page 1, new Likeness draft, whole region → viewer page 2, draft page 2, region whole | AC-1.3, D-4 |
| T-4 | Swift | `CitationComposerDepictionTests` `wholePageStampsANewCitationsPage` *(new)* | `seeded(.pdf)`, add entry, `artifactViewer.setPage(2)`, whole region → `model.locator.page == 2`, draft page 2 | AC-1.4 |
| T-5 | Swift | `CitationComposerDepictionTests` `regionButtonsFollowDraw` *(new)* | existing PDF Citation with no page → `useDepictionRegion` returns false and the draft has no region; image Citation with `model.transcriptions.isTranscribing = true` (as `drawCannotStartWhileTranscribing` sets it) → returns false | AC-1.5, AC-1.6, D-7 |
| T-6 | Swift | `CitationComposerDepictionTests` `citationRegionIsCopied` *(new)* | `seedCitation(locatorJSON: regionJSON(rect(0.2, 0.2, 0.6, 0.7)))`, edit entry, new Likeness draft → `depictionCitationRegion` equals that region; `useDepictionRegion` with it → draft region equals it. Then, on the same model before the draft: `setRegion(rect(0.5, 0.5, 0.9, 0.9))` on the Citation without saving, open a Likeness draft → `depictionCitationRegion` is still `rect(0.2, 0.2, 0.6, 0.7)` | AC-2.1, AC-2.7, D-6 |
| T-7 | Swift | `CitationComposerDepictionTests` `citationRegionOnItsPage` *(new)* | `seeded(.pdf)`, Citation page 1 with a region, viewer page 2 → after the call, viewer page 1 and draft region on page 1 | AC-2.2 |
| T-8 | Swift | `CitationComposerDepictionTests` `noCitationRegionMeansNone` *(new)* | artifact-only Citation → `depictionCitationRegion == nil`; new Citation after `setRegion(rect(0.1, 0.1, 0.5, 0.5))` then a Likeness draft → equals that region | AC-2.3, AC-2.4 |
| T-9 | Swift | `CitationComposerDepictionTests` `regionButtonOnAnEditIsARedraw` *(new)* | `editingSavedLikeness()`, `useDepictionRegion(rect(0.5, 0.5, 0.9, 0.9))` → `isDepictionRedrawn`, the file's `caption(model)` helper returns `L10n.string(L10n.CitationComposer.depictionCaptionRedrawn)`, `saveEdit()` sends an `updateObservation` call ending ` depiction=1 depictionFile=present`; `editingSavedLikeness(region: regionJSON(rect(0, 0, 1, 1)))` then the whole region again → `isDepictionRedrawn == false`. STEP-5 adds the `region:` parameter to `editingSavedLikeness`, defaulting to today's `regionJSON(rect(0.1, 0.1, 0.4, 0.4))` | AC-2.5, AC-2.6, D-8 |
| T-10 | Swift | `CitationComposerDepictionTests` `wholeFileTitleFollowsTheFile` *(new)* | `seeded(.pdf)` → `L10n.string(model.depictionWholeFileTitle) == L10n.string(L10n.CitationComposer.depictionUseWholePage)`; `seeded()` (image) → `depictionUseWholePhoto` | AC-1.8, D-3 |

AC-1.9, AC-3.1, AC-3.2, AC-3.3 and AC-3.4 are view-level focus and VoiceOver behavior: manual check steps 6, 7 and 8. T-3, T-4 and T-7 also prove the shared preamble (`prepareDepictionRegion`) moves the page for the buttons; the existing Draw tests prove it still does for Draw.

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break: `CitationComposerDepictionTests` (`tabFromPropertySkipsADisabledDraw`, `drawingDisablesCitationControlsAndLeaveReturns`, the PDF page tests that now run through `prepareDepictionRegion`), and `NameValueDraftTests` (part moves; the draft logic is untouched, only the announcement call moves).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches §6 macOS (brief waived); no new kit primitive
- [ ] Version bump: none
- [ ] After the draft PR has a number, implement-pr step 8's commit "Record S10-63 as landed in pull request N.": this plan's Status → Landed in #N (date) with a Landed note listing every difference from the plan, and the S10-63 row in spike-10-followups/deployment-plan.md → Landed #N (this folder has no completed.md)
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
7. Turn on VoiceOver. Press **Draw on the file**. VoiceOver says "Draw the region with the tools above. Escape goes back to the observation." Press Escape. Then press **Use the whole photo**: VoiceOver reads the new caption.
8. With VoiceOver on, open a person's name editor and move a name part with its move button: the part-moved announcement still plays.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Focus does not land on Cancel because the sheet is still dismissing when the band appears | D-11 writes the focus twice, a turn apart, as `claimTab` does |
| Three buttons overflow the dialog width | `FlowLayout` wraps them (D-9) |
| The whole-page preview of a large PDF page renders slowly on the main thread | Same cost as drawing a page-sized box today; audit PRF-1 is Out |

**Stop and report if:**

- `commitViewerRegion(_:)` or `depictionPageCheck` no longer behaves as §4 says.
- A whole-file rectangle fails `ArtifactRegionGeometry.isValidPolygon` or `locator.ValidateDepictionRegion` (T-1 fails without a code change in `core/locator`).
- `CitationDepictionValueArea` is no longer the value area for both the add dialog and `CitationComposerEditSheet`.
- Focus on Cancel needs a change to `PVButton` or another kit component.
- Manual check step 6 shows focus is not on Cancel after D-11's double hop.
- Moving `NameValueEditorForm.moveAndAnnounce` onto the helper changes when or what it announces (manual check step 8).
- The save of a button-set region on an existing Citation is refused by Go with `locator.invalid` (the saved Citation locator and `depictionCitationRegion` disagree).

## 12. Open questions

None.

## 13. PR

**Title:** S10-63 — Keyboard region for a likeness
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-63-keyboard-likeness-region.md`.
