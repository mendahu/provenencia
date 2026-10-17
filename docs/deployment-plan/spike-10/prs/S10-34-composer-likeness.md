# S10-34 — Composer likeness: draw a face, save it, see it

**Status:** Blocked — on board S10-D15
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-34--composer-depiction-row)
**Brief:** [S10-D15](../design/S10-D15-composer-depiction-row.md), board not agreed yet
**Board:** the user sends the Claude Design link when the board is agreed; it is recorded here, and before this plan is `Planned` §6 macOS describes each agreed frame in words
**Depends on:** S10-16 (landed [#384](https://github.com/mendahu/provenencia/pull/384)), S10-54 (landed [#433](https://github.com/mendahu/provenencia/pull/433)), S10-D15; S10-52 is in progress and this plan assumes it has landed — it changes no file this plan names (its plan §6 macOS: "None — no Swift change"), so its landing changes no row here
**Blocks:** —
**Branch:** `s10-34-composer-likeness`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [S10-D15](../design/S10-D15-composer-depiction-row.md) CDR-1 to CDR-13; [depictions and likenesses](../depictions-and-likenesses.md) ("The image"); [S10-54 plan](S10-54-depiction-ffi.md) (store types) and its Landed note; [S10-D3](../design/archive/S10-D3-citation-composer.md) CC-3, CC-4, CC-12; working rules "A depiction's value is its own row: region plus image", "The client renders the image; Go validates it and stores its own copy", "`likeness` is new catalogs only"
**Planned:** 2026-10-17 against `main` @ `ac05ae8` (S10-54 landed; S10-52 assumed landed)
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

The engine and the FFI can already store a likeness: a region on the Citation's page and an image Go re-encodes from the app's crop (S10-32, S10-54). Nothing in the app offers one. This PR lets the researcher choose **Likeness** on a person in the citation composer, draw the face on the photo or the PDF page with the viewer's existing region tools, see the face before saving, save it with an empty transcription, and see the saved face on its line. Editing can redraw the face, and Duplicate copies its region. The app renders the crop (a pixel crop of an image, a PDFKit render of a PDF page) to a temporary PNG, sends its path, and deletes the file after the call.

**Why now.** It is the last piece between the engine work and the gallery (S10-35): without it no likeness can be entered, so the Person page has nothing to show.

**What stays.** The composer layout from S10-16, the Citation chooser and its locator line, Auto Transcribe and Paste, bridges, every other property type, and the leave confirm. The Citation's own region stays the Citation's. The engine, the FFI, and the proto are unchanged.

## 2. User stories

**US-1.** As a researcher with a group photo, when I add an observation on a person, I can choose Likeness, draw that person's face on the photo, and save it, so that the face is evidence on that person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A catalog with `likeness` bound to person; an image Citation; the dialog on person Ada | The property options for Ada include Likeness (CDR-1) |
| AC-1.2 | The same, on an audio Artifact, or a catalog with no depiction Property | Likeness is not offered (CDR-1) |
| AC-1.3 | Likeness chosen, no region drawn | Save is disabled (CDR-4) |
| AC-1.4 | Draw a rectangle (0.25,0.25)–(0.75,0.75) on an 8×8 photo | The draft holds that region; the Citation's locator region is unchanged (CDR-3); the preview is a 4×4 image (CDR-4) |
| AC-1.5 | Save with an empty Citation transcription | One `addObservationsToCitation` call carries `depiction.regionJSON` with one region selector and no page selector, and `imagePath` names a PNG that exists during the call; after the call the file is gone (CDR-5) |
| AC-1.6 | After AC-1.5 succeeds | The draft is Likeness with an empty subject and no region; the next opening focus is the subject pick (CDR-6) |
| AC-1.7 | The saved observation is listed with `depiction.imageRelPath = "objects/ab/cd.png"` | Its line's `imageRelPath` is that path and its value text is empty (CDR-7) |
| AC-1.8 | The store throws `ingest.unsupported_image` | The dialog stays open with the region kept and the mapped error copy in its error; the temporary file is gone (CDR-11) |
| AC-1.9 | A face is drawn and not saved; navigation is requested | The leave guard holds (CDR-12) |

**US-2.** As a researcher, when a likeness is wrong, I can change its person or redraw it, or duplicate it for another face, without losing the original.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Edit a saved likeness, change only the subject, save | `updateObservation` gets `observation.valueDepictionID` equal to the saved id and `depictionInput == nil` (CDR-8) |
| AC-2.2 | Edit a saved likeness, redraw the region, save | `updateObservation` gets a `depictionInput` whose region is the new one and whose file exists during the call (CDR-8) |
| AC-2.3 | Duplicate a saved likeness | The draft has the same subject, property, polarity, and region, an empty `valueDepictionID`, and is not saved; the opening focus is the subject pick (CDR-9) |

**US-3.** As a researcher on a PDF, the face lands on the Citation's page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A new Citation on a 2-page PDF, viewer on page 2, face drawn | The Citation locator's page becomes 2; the region JSON's page selector is 2 (CDR-10) |
| AC-3.2 | An existing Citation saved with page 1, viewer on page 2, drawing starts | The viewer moves to page 1; a commit while the viewer is on page 2 is refused (CDR-10) |
| AC-3.3 | An existing Citation saved with no page, Likeness chosen | The draft reports the needs-page state; drawing cannot start (CDR-10) |
| AC-3.4 | A PDF page with `/Rotate 90`, region (0,0)–(0.5,0.5) | The render is the visual top-left quarter of the rotated page (D-5) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Likeness offered on a person on an image or PDF Artifact (CDR-1) | STEP-1, AC-1.1, AC-1.2 |
| Drawing the face with the existing region tools without changing the Citation's region (CDR-2, CDR-3) | STEP-4, STEP-8, AC-1.4 |
| The face preview and Save gating (CDR-4) | STEP-2, STEP-8, AC-1.3, AC-1.4 |
| Save with an empty transcription; temporary PNG written and deleted (CDR-5) | STEP-5, AC-1.5 |
| Next face one action away (CDR-6) | STEP-5, AC-1.6 |
| Saved line shows the face (CDR-7) | STEP-9, AC-1.7 |
| Edit with redraw (CDR-8) | STEP-6, AC-2.1, AC-2.2 |
| Duplicate copies the region (CDR-9) | STEP-7, AC-2.3 |
| PDF page rule (CDR-10) | STEP-4, AC-3.1–AC-3.3 |
| Refused save keeps the region (CDR-11) | STEP-5, AC-1.8 |
| Leave holds an unsaved face (CDR-12) | STEP-1, AC-1.9 |
| VoiceOver labels (CDR-13) | STEP-8, STEP-9, manual check |
| FakeStore stores and returns depictions | STEP-3 |

| Out | Home |
| --- | --- |
| Person gallery and Persons list thumb | S10-35 |
| Preferred portrait | S10-39 |
| Promote's side-by-side likenesses | S11-20 |
| Showing every saved face on the photo at once | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) (this PR adds the line) |
| A polygon-masked, transparent crop | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| Place, event, and signature keys | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| Guarding a Citation page change after likenesses are saved on it | dogfood log ([`docs/dogfood/ux.md`](../../../dogfood/ux.md)); Go already hides a region that no longer matches (`depictionvalues.Readable`) |
| Engine, FFI, or proto changes | not planned |
| Deleting the unused `CitationComposerModel.propertyOptions(for:)` | not planned in this PR |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Value types the composer offers | `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes`, `listedProperties(typeKey:properties:rules:)`, `propertyOptions(forSubjectID:)` | Text, integer, date, term, name. Depiction is filtered out, so Likeness never appears (seeded in `core/database/subjectvocab/registry.go`, key `likeness`, person, locked) |
| Draft values | `macos/App/Features/CitationComposer/CitationObservationRows.swift` `ObservationRowValues` | No depiction fields. `ObservationRow.loaded(from:)` drops `valueDepictionID` and `depiction`; `asCatalogObservation(property:)` builds a `CatalogObservation` without `valueDepictionID`, so a likeness update would fail Go's `resolve` (`core/database/observations/observations.go`) with `ErrInvalid` |
| Value logic | `macos/App/Features/CitationComposer/CitationObservationValue.swift` `Fields`, `summary`, `isDialogValid`, `apply` | `default` branches: summary `""`, dialog invalid, apply `.unsupported` |
| Subject change in the dialog | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `applyObservationSubject(_:field:values:)`, `clearObservationValue(_:)` | Clears the property when the subject's type changes, including from an empty subject — that would drop a kept Likeness (CDR-6) |
| Subject change on a row | `macos/App/Features/CitationComposer/CitationObservationRows.swift` `applySubject(rowID:subjectID:vocabulary:)` | Clears the property when it is not in `propertyOptions(forSubjectID:)` |
| New observation save | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `savePropertyDraft(_:)`, `ensuredCitationID()` | Builds a `CatalogObservationDraft` through `ObservationRow.asDraft`, creates the Citation first when new (`CitationComposerContext.ensureCitationID` → `createCitationWithObservations` with `fields.values.locator`), then `addObservationsToCitation`; then resets the draft to the same subject |
| Edit save | `CitationComposerModel+Authoring.swift` `saveEdit()`; `CitationObservationRows.swift` `commit(rowID:citation:)`, `performCommit(rowID:context:citation:)` | Calls the three-argument `updateObservation` overload; no `depictionInput` |
| Duplicate | `CitationComposerModel+Authoring.swift` `duplicate(line:)` | Copies `row.values` into a new `CitationPropertyDraft` with `isDuplicate: true` |
| Dialog presentation | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `ComposerPresented`, `show(_:)`, `showAuthoringDialog()`, `showEditDialog()`, `suspendedDialog()`, `restorePresented(_:)` | One presented dialog at a time; `returnTo:` cases suspend and restore the dialog under them |
| Opening focus | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationAuthoringNavigation.initialPropertyFocus(subjectID:)`; `CitationComposerAuthoringBox.swift` `propertyFields(_:)` `.onChange(of: draft.isSaving)` | Subject when empty, else property; after a save, focus goes to property |
| Value controls | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields` `valueControls`, `propertyOptions` | Switch on value type; `default` shows a disabled input. Shared with the edit sheet (`CitationComposerSubjectGroups.swift` `CitationComposerEditSheet.propertyForm`) |
| Viewer | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` | Passes `committedRegion: model.locator.region` and `onCommitRegion: { model.setRegion($0) }` to `ArtifactViewer`; one region at a time |
| Citation region | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `setRegion(_:)`, `armedRegionTool`, `disarmRegionTool()`, `setPage(_:)`, `locatorCapabilities` | Drawing writes the Citation's locator; `setRegion` stamps the viewer page on a PDF |
| Region JSON | `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` `encodeJSON()`, `decode(_:)` | Encodes artifact, page, and normalized region selectors; `core/locator/depiction.go` `ValidateDepictionRegion` accepts that shape (leading artifact selector allowed) |
| Region geometry | `macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift` `boundingRect(of:)`, `isValidPolygon(_:)` | Unit square, origin top-left, y down |
| Image crop | `macos/App/Features/OCR/OCRImage.swift` `cgImage(from:)`, `crop(_:normalizedRect:)` | Bounding-box crop used by Auto Transcribe (`CitationTranscriptionModel.runAutoTranscribe`) |
| PDF region render | — | None. `ArtifactPDFViewport.swift` `layoutPageOverlay()` normalizes against the media box as displayed (rotation applied) |
| Viewer content | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `kind`, `displayImage`, `pdfDocument`, `page` (1-based), `setPage(_:)` | The loaded file; no URL kept |
| PNG writer, temp file | — | None in the app |
| Saved lines | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationSavedLine`, `CitationComposerAuthoring.groups(observations:connections:vocabulary:nameOrders:)`; `CitationComposerSubjectGroups.swift` `lineRow(_:rule:)` | A likeness line today reads "Likeness —" with no image |
| Saved image display | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`, `macos/App/Platform/ThumbnailCache.swift` | Loads `objects/…` through `ProjectFiles.objectURL` |
| Store types | `macos/App/Platform/GenealogyStore.swift` `CatalogDepictionValueInput`, `CatalogDepictionValue`, `CatalogObservation.valueDepictionID` / `.depiction`, `CatalogObservationDraft.depiction`, `updateObservation(projectDir:userID:observation:depictionInput:)` | Landed in S10-54 |
| Fake store | `macos/App/Platform/FakeStore.swift` `appendFakeObservations`, `fakeObservation(citationID:id:ref:row:)`, `updateObservation(…depictionInput:)` | Ignores depiction input; returns no `depiction` |
| Error copy | `macos/App/Platform/L10n.swift` `L10n.Errors.ingestUnsupportedImage`, `depictionValuesInvalid` | Landed in S10-54 |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` (`firstRowSaveOnNewCreatesCitationAndExposesRef`, `saveKeepsTheSubject`, `duplicateDoesNotWriteAndConfirmsADirtyDraft`, `saveEditClosesTheDialog`); `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift` (`regionCropIsSmallerThanFullImage`, private `writePNG`, `writeBlankPDF`, `uniqueProjectDir`, `seedArtifact`, `waitUntil`); `macos/ProvenenciaTests/OCRImageTests.swift`; `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` | No likeness authoring test |

**Precedent.** The S10-54 plan and [#433](https://github.com/mendahu/provenencia/pull/433) for the store types this PR fills. Auto Transcribe's region crop (`CitationTranscriptionModel.runAutoTranscribe`: `OCRImage.cgImage(from:)` → `ArtifactRegionGeometry.boundingRect(of:)` → `OCRImage.crop(_:normalizedRect:)`) is the image half of D-5; copy it, do not copy its whole-page OCR warning. The custom-term and new-subject dialogs (`ComposerPresented.customTerm(_:over:)`, `.newSubject(_:over:)`) are the precedent for suspending the observation dialog and returning to it (D-7).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `supportedValueTypes` gains `depiction`. `listedProperties` and `propertyOptions(forSubjectID:)` take `allowsDepiction: Bool` (default `false`); when false, depiction Properties are dropped. `CitationPropertyFields.propertyOptions` passes `model.locatorCapabilities.supportsRegionLocator`. `CitationObservationRows.applySubject` passes `true` (a saved likeness row exists only on a region-capable Artifact). `make(...)`'s `propertyComboOptionsByTypeID` passes `false`. Any depiction Property bound to the subject's type is offered; only `likeness` exists. | The value type decides the control, the key decides where it shows (depictions doc). The filter by file kind sits beside the existing value-type filter, so there is one list. | A `likeness` key check — the next depiction key would need a second branch. A separate likeness button — a second authoring entry beside CC-3's one home |
| D-2 | `ObservationRowValues` gains `depictionRegion: ArtifactRegionDraft? = nil`, `depictionPage: Int? = nil`, `valueDepictionID: String = ""`, `depictionImageRelPath: String = ""`. `ObservationRow.loaded(from:)` fills them from `observation.valueDepictionID` and `observation.depiction` (region and page through `CitationLocatorDraft.decode(regionJSON)`). `asCatalogObservation` sets `valueDepictionID`. `clearObservationValue` and `CitationObservationRows.clearValue` clear all four. | The struct is `Equatable` and already drives dirty and leave (`CitationPropertyDraft.holdsLeave`), so a drawn face holds leave with no new flag (CDR-12). Keeping `valueDepictionID` fixes the update path Go would otherwise refuse | A separate likeness draft struct — a second draft type beside `CitationPropertyDraft` |
| D-3 | `CitationObservationValue.Fields` gains `depictionRegion`. `summary` returns `""` for depiction (the line shows the image). `isDialogValid` requires `depictionRegion?.isValid == true`. `apply` returns `nil` for a valid region and `.unsupported` otherwise, and does not set `draft.depiction`; the model attaches the input after rendering (D-6). | `apply` is pure and cannot render. One value enum keeps every type's rules together | Rendering inside `apply` — it would need the viewer content and file I/O |
| D-4 | The region JSON is `CitationLocatorDraft(page: values.depictionPage, region: region, extras: []).encodeJSON()`. Image: no page. PDF: the page from D-8. | Reuses the encoder the Citation uses; Go accepts the leading artifact selector | A second JSON encoder for regions |
| D-5 | `ArtifactRegionCrop` *(new)*, `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift`: `static func render(_ viewer: ArtifactViewerModel, page: Int?, region: ArtifactRegionDraft) -> CGImage?`. Image: `OCRImage.cgImage(from: viewer.displayImage)` → `OCRImage.crop(_:normalizedRect: boundingRect(of:))`. PDF: `viewer.pdfDocument?.page(at: page - 1)`; the visual page size is `bounds(for: .mediaBox).size`, width and height swapped when `rotation` is 90 or 270; scale `s = min(4, 2048 / (bbox long edge in points), 8192 / (visual page long edge))`; `page.thumbnail(of: visualSize × s, for: .mediaBox)` → `OCRImage.cgImage(from:)` → `OCRImage.crop(_:normalizedRect:)`. Both paths then scale down to at most `maxLongEdge = 2048` on the long edge. Returns nil on any failure. | The bbox crop is the saved image (working rule). `thumbnail(of:for:)` draws in the displayed orientation, the same space the overlay normalizes against. Go re-encodes to 2048 anyway; capping on the client keeps the file far under `ingest` `imageMaxBytes` (32 MiB) and `raster` decode budget | Extending `OCRImage` — it is the OCR feature's helper; this type reuses its crop and adds the PDF and size rules. Drawing the PDF with a hand-built CTM — rotation is where it goes wrong |
| D-6 | `ArtifactRegionCrop.writeTemporaryPNG(_:) throws -> URL` *(new)* encodes with `CGImageDestination` (`UTType.png`) to `FileManager.default.temporaryDirectory/provenencia-likeness-<UUID>.png`. The model renders and writes on the main actor immediately before the store call, and removes the file in a `defer` around the call, success or failure. A nil render or a write error sets the draft error to `L10n.CitationComposer.likenessRenderFailed` *(new)* and calls nothing. | Go never deletes the source (`ingest.StageImage`); the app owns its temp file. Rendering at save, from the region, makes the saved image deterministic | Rendering when the region is drawn and keeping the file until save — a discarded draft would leak files |
| D-7 | Drawing is a presented state: `ComposerPresented.drawingLikeness(returnTo: ComposerPresented)` *(new)*. `startLikenessDrawing()` *(new)* suspends the open observation or edit dialog (`show(.drawingLikeness(returnTo:))`), keeps `authoring` / `editSession`, and arms `armedRegionTool = .rectangle` when none is armed. While drawing, the viewer shows the draft's region (`viewerRegion` *(new)*) and commits go to `commitViewerRegion(_:)` *(new)*, which writes `depictionRegion` and `depictionPage` into the draft and restores the dialog. `cancelLikenessDrawing()` *(new)* (disarm or Escape) restores the dialog with the region unchanged. Outside drawing, `viewerRegion` is the Citation's region and `commitViewerRegion` is `setRegion`. The chrome that tells the researcher they are drawing comes from the agreed board (§6 macOS). | The dialog is a modal `FormDialog` over the viewer (S10-D15 domain facts), and the brief forbids a second viewer. Routing through the model keeps `ArtifactViewer`'s API unchanged and keeps the Citation's region untouched (CDR-3) | A viewer inside the dialog — the brief's non-goal. A second region layer on `ArtifactRegionOverlayView` — showing saved faces is Out |
| D-8 | PDF page rule (answered 2026-10-17): `likenessPage` *(new)* is `fields.baseline.locator.page` for an existing Citation, else `fields.locator.page`. Existing Citation with no saved page: `likenessNeedsPage` *(new)* is true, the draw control is disabled, and the dialog shows the Callout `L10n.CitationComposer.likenessNeedsPage` *(new)*. With a page: `startLikenessDrawing` calls `artifactViewer.setPage(page)`, and `commitViewerRegion` refuses (no write, tool stays armed) while `artifactViewer.page != page`. New Citation with no page: the commit calls `setPage(artifactViewer.page)` on the Citation locator, then uses that page. Images: no page, no rule. | `ValidateDepictionRegion` requires the region's page to equal the Citation's saved page. A new Citation is created with `fields.values.locator` by `ensuredCitationID`, so the stamped page is saved with it | Refusing on every Citation without a page — the user chose the stamp for a new Citation |
| D-9 | Edit: `CitationObservationRows.commit(rowID:citation:depictionInput:)` and `performCommit` gain `depictionInput: CatalogDepictionValueInput? = nil`, passed to `updateObservation(…depictionInput:)`. `saveEdit` renders and writes a new input only when `values.depictionRegion != propertyBaseline.depictionRegion`; otherwise it passes nil and the row keeps `valueDepictionID`. | Redraw is allowed (answered 2026-10-17). Go replaces the image only when an input is sent | Always re-rendering — every subject edit would store a new image |
| D-10 | Duplicate copies `row.values` and clears `valueDepictionID` and `depictionImageRelPath`; the region and page stay (answered 2026-10-17). Saving renders a new image (D-6). | CC-12 for every saved row; the copy owns its own image | Hiding Duplicate on likeness lines |
| D-11 | Next face (CDR-6): after a likeness saves, `savePropertyDraft` resets to `ObservationRowValues` with only `propertyID` kept. For every other value type it keeps the subject, as today. `applyObservationSubject` clears the property only when it is not in `propertyOptions(forSubjectID: raw, allowsDepiction: locatorCapabilities.supportsRegionLocator)`, instead of on any type change. Focus: `CitationAuthoringNavigation.initialPropertyFocus(subjectID:propertyID:)` returns `.subject` when the subject is empty, else `.property` when the property is empty, else `.subject` for a depiction property. The post-save focus in `propertyFields(_:)` uses the same function. | A group photo is a run of faces, one per person. The type-change rule would drop the kept Likeness on the first subject pick; the option-membership rule is the one `applySubject` already uses | Keeping the subject — every next face is a different person |
| D-12 | Saved line: `CitationSavedLine` gains `imageRelPath: String?` *(new)*, set in `groups(...)` for a depiction row with a non-empty `depictionImageRelPath`. `lineRow` draws `CachedThumbnail(projectDir:relPath:size:)` in place of the value text when `imageRelPath` is set. The thumbnail size, and the size and placement of the dialog preview, are the agreed board's. | `CachedThumbnail` is the loader for `objects/…` paths and owns loading and unreadable states (CDR-7) | `NSImage(contentsOf:)` on the render path — `CachedThumbnail`'s doc forbids it |
| D-13 | Dialog preview: `likenessPreview` *(new)* on the model, an `NSImage?` rendered with `ArtifactRegionCrop.render` whenever the draft's region changes (commit, duplicate, edit open), shown through `PVThumbnail` with content `.image`. | One renderer for preview and save (CDR-4) | Showing the saved file in edit — a redraw would then need a second path |
| D-14 | `FakeStore` stores depictions: a draft or update input with a depiction records `depiction=1` on the call, checks `FileManager.default.fileExists(atPath: input.imagePath)` and records `depictionFile=present` or `depictionFile=missing`, and returns `valueDepictionID = "dep-<n>"` and `depiction = CatalogDepictionValue(regionJSON: input.regionJSON, imageRelPath: "objects/fake/dep-<n>.png")`. An update with no input keeps the stored value. `FakeStore.depictionWriteError: Error?` *(new)* is thrown by every add or update that carries a depiction input while it is set. | Model tests assert the call order, the file during the call, and the error path without Go | A test-only store subclass — `FakeStore` is the one fake |
| D-15 | Tests: model tests go in `macos/ProvenenciaTests/CitationComposerLikenessTests.swift` *(new)*; render tests in `macos/ProvenenciaTests/ArtifactRegionCropTests.swift` *(new)*. `uniqueProjectDir`, `writePNG`, `writeBlankPDF`, `seedArtifact`, and `waitUntil` move from private members of `CitationAutoTranscribeTests` into `macos/ProvenenciaTests/ComposerArtifactFixtures.swift` *(new)* as `enum ComposerArtifactFixtures` static functions, and `CitationAutoTranscribeTests` calls them. | Both suites need an image or PDF Artifact on disk; one copy | Copying the helpers |
| D-16 | `xcodebuild`, the Swift tests, and the manual check run on a Mac. Without one, the implementer finishes the non-Xcode checks, stops before STEP-1's Swift check, and reports which checks did not run. | Precedent: S10-23 D-11 | Claiming a Swift check that was not run |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionCrop` *(new)* | type / file | `macos/App/Features/OCR/OCRImage.swift` `OCRImage` | Reuses `cgImage(from:)` and `crop(_:normalizedRect:)`. Adds the PDF render, the 2048 cap, and the PNG writer, which are not OCR concerns |
| `ArtifactRegionCrop.writeTemporaryPNG(_:)` *(new)* | func | `macos/App/Features/Workspace/NavigationHistoryStore.swift` (atomic writes) | No PNG encoder exists in the app (searched `CGImageDestination`, `representation(using:`) |
| `ComposerPresented.drawingLikeness(returnTo:)` *(new)* | enum case | `ComposerPresented.customTerm(_:over:)`, `.newSubject(_:over:)` | Extends the same presented-state registry |
| `startLikenessDrawing`, `cancelLikenessDrawing`, `commitViewerRegion`, `viewerRegion`, `likenessPage`, `likenessNeedsPage`, `likenessPreview` *(new)* | funcs / properties | `CitationComposerModel.setRegion(_:)`, `setPage(_:)` | Route the one viewer region between the Citation and the draft; `setRegion` stays the Citation's |
| `ObservationRowValues` depiction fields *(new)* | fields | `dateDraft`, `nameDraft` | Same struct, same pattern as the other structured values |
| `CitationSavedLine.imageRelPath` *(new)* | field | `CitationSavedLine.valueText` | An image is not text |
| `allowsDepiction:` parameter *(new)* | param | `CitationComposerVocabulary.supportedValueTypes` | The file kind is not in the vocabulary snapshot |
| `FakeStore.depictionWriteError` *(new)* | test hook | `FakeStore.updateSourceError`, `unpromoteSubjectError` | Same fake, same `<call>Error` hook shape |
| `ComposerArtifactFixtures` *(new)* | test helper | private helpers in `CitationAutoTranscribeTests` | Moved, not copied |
| `L10n.CitationComposer.likeness*` *(new)* | keys | `L10n.CitationComposer` | Same namespace |

**Builds on.** `ArtifactViewer` and its region tools unchanged; `OCRImage` crop; `CitationLocatorDraft` encoder; `ComposerPresented` `returnTo:` suspension; `CachedThumbnail`; the S10-54 store types; `CitationObservationValue` as the one value-rule switch.
**Deletes.** The `default` "unsupported" path stops being reachable for depiction. The private helpers in `CitationAutoTranscribeTests` move (D-15). No other code becomes dead.
**Next adds.** A place, event, or signature depiction key is one seeded Property: D-1 offers any depiction Property, and every path here keys on the value type. S10-35 reads the images this PR saves.

### Schema

None — no catalog change.

### Go core

None — no engine change.

### FFI

None — the S10-54 fields carry everything.

### macOS

Frame-dependent rows are filled from the agreed S10-D15 board when this plan leaves `Blocked`. The model behavior above does not depend on the frames.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Likeness value area | `CitationPropertyFields` `valueControls` depiction case | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` | Preview `PVThumbnail` (D-13), a `PVButton` `L10n.CitationComposer.likenessDraw` / `likenessRedraw` calling `startLikenessDrawing()`, disabled when `likenessNeedsPage`; `PVCallout` with `likenessNeedsPage` copy. Shared by the edit sheet |
| Viewer wiring | `ArtifactViewer` | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` | `committedRegion: model.viewerRegion`, `onCommitRegion: { model.commitViewerRegion($0) }`, `onDisarmRegionTool: { model.disarmRegionTool() }` where `disarmRegionTool` calls `cancelLikenessDrawing()` while drawing. The locator strip keeps reading `model.locator` |
| Drawing chrome | per the agreed board | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` | The hint text `L10n.CitationComposer.likenessDrawingHint` replaces `locatorDrawHint` while drawing; the rest is the board's |
| Saved likeness line | `CachedThumbnail` → `PVThumbnail` | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` `lineRow` | D-12. Accessibility label `likenessLineLabel(person)` on the thumbnail |
| Dialog presentation | `pvFormDialog` | `macos/App/Features/CitationComposer/CitationComposerView.swift` | `authoringSheetBinding` and `editSheetBinding` read false during `.drawingLikeness`; their setters must not discard while drawing |

States: no region → Save disabled, preview absent; render failed → Callout `likenessRenderFailed`, region kept; image loading or unreadable → kit thumbnail states; existing Citation with no page → Callout, draw disabled. Accessibility: CDR-13 labels; drawing stays pointer-only.

### Strings

L10n keys *(new)* in `L10n.CitationComposer` (copy is the brief's; the agreed board may change it before `Planned`):

- `likenessDraw` = "Draw on the file"
- `likenessRedraw` = "Draw again"
- `likenessDrawingHint` = "Draw the face. Press Escape to go back to the observation."
- `likenessPreviewLabel` = "Likeness to save"
- `likenessThumbLabel(_ person: String)` = "Likeness of %@"
- `likenessLineLabel(_ person: String)` = "Likeness, %@"
- `likenessNeedsPage` = "Set this citation's page to draw a likeness."
- `likenessRenderFailed` = "The likeness could not be cut from the file. Draw the region again."

Removed: none.

## 7. Steps

**STEP-1 — Values, rules, and the property list**
Files: `CitationObservationRows.swift`, `CitationObservationValue.swift`, `CitationComposerVocabulary.swift`, `CitationComposerModel+Authoring.swift` (`applyObservationSubject`, `clearObservationValue`), `CitationComposerAuthoring.swift` (`initialPropertyFocus`), `CitationComposerAuthoringBox.swift` (`propertyOptions` call only), `macos/ProvenenciaTests/ComposerArtifactFixtures.swift` *(new)*, `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift`, `macos/ProvenenciaTests/CitationComposerLikenessTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`. Skill: `add-swift-test`.
Do: D-1, D-2, D-3, D-11 (subject rule and focus function only), D-15.
Check: T-1, T-2, T-3, T-12 pass; `CitationAutoTranscribeTests` and `CitationComposerModelTests` pass unchanged.

**STEP-2 — Render and write the crop**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` *(new)*, `macos/ProvenenciaTests/ArtifactRegionCropTests.swift` *(new)*, `project.pbxproj`.
Do: D-5, D-6 (the writer).
Check: T-4, T-5, T-6, T-7 pass.

**STEP-3 — FakeStore depictions**
Files: `macos/App/Platform/FakeStore.swift`.
Do: D-14.
Check: T-8 passes; every existing suite that uses `FakeStore` passes.

**STEP-4 — Drawing state and the PDF page rule**
Files: `CitationComposerModel.swift`, `macos/App/Features/CitationComposer/CitationComposerModel+Likeness.swift` *(new)*, `CitationComposerView.swift` (bindings), `CitationComposerViewerPane.swift` (wiring), `project.pbxproj`.
Do: D-7, D-8, D-13. The new extension file holds every new likeness method and property.
Check: T-9, T-10, T-11 pass.

**STEP-5 — Save a new likeness**
Files: `CitationComposerModel+Authoring.swift` (`savePropertyDraft`), `CitationComposerModel+Likeness.swift`.
Do: D-6 (call site), D-11 (reset). Attach `CatalogDepictionValueInput(regionJSON:imagePath:)` to the draft after `asDraft`.
Check: T-13, T-14, T-15 pass.

**STEP-6 — Edit and redraw**
Files: `CitationObservationRows.swift` (`commit`, `performCommit`), `CitationComposerModel+Authoring.swift` (`saveEdit`).
Do: D-9.
Check: T-16, T-17 pass; `saveEditClosesTheDialog` passes.

**STEP-7 — Duplicate**
Files: `CitationComposerModel+Authoring.swift` (`duplicate(line:)`).
Do: D-10.
Check: T-18 passes; `duplicateDoesNotWriteAndConfirmsADirtyDraft` passes.

**STEP-8 — Dialog value area and drawing chrome**
Files: `CitationComposerAuthoringBox.swift` (`valueControls`, post-save focus), `CitationComposerViewerPane.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skills: `add-ui-component`, `add-localized-string`.
Do: §6 macOS rows "Likeness value area", "Drawing chrome", the strings in §6 Strings, per the agreed board.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; the app builds in `xcodebuild`; manual check steps 1–4.

**STEP-9 — Saved likeness line**
Files: `CitationComposerAuthoring.swift` (`CitationSavedLine`, `groups`), `CitationComposerSubjectGroups.swift` (`lineRow`).
Do: D-12.
Check: T-19 passes; manual check step 5.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (map row and Slice 8 line), this plan (status, Landed note), `docs/deployment-plan/spike-10/design/S10-D15-composer-depiction-row.md` → `design/archive/`, `docs/deployment-plan/spike-10/design/README.md`, `docs/ideas/depictions-and-likenesses.md` (add "Every saved face on the photo at once"; fix "as a JPEG" to "as a PNG").
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Draw a face on the photo and save it as a likeness.").

## 8. Tests

Swift Testing. Model tests use `FakeStore` with a `likeness` Property (`valueType: "depiction"`) bound to person, and an image or PDF Artifact from `ComposerArtifactFixtures`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CitationComposerLikenessTests.swift` `likenessIsOfferedOnImageAndPDFOnly` *(new)* | image and PDF Artifacts → options include Likeness; audio Artifact → absent; catalog with no depiction Property → absent | AC-1.1, AC-1.2, D-1 |
| T-2 | Swift | `CitationComposerLikenessTests.swift` `loadedLikenessKeepsItsValue` *(new)* | `CatalogObservation` with `valueDepictionID "dep-1"` and a page-2 region JSON → `ObservationRow.loaded` has region, page 2, `"dep-1"`, image path; `asCatalogObservation` keeps `"dep-1"` | D-2 |
| T-3 | Swift | `CitationComposerLikenessTests.swift` `likenessNeedsARegionToSave` *(new)* | Likeness, no region → `authoringCanConfirm == false`; valid region → true | AC-1.3, D-3 |
| T-4 | Swift | `ArtifactRegionCropTests.swift` `imageCropIsTheBoundingBox` *(new)* | 8×8 image, rectangle (0.25,0.25)–(0.75,0.75) → 4×4; top-half-red image, region in the top half → red pixel | AC-1.4, D-5 |
| T-5 | Swift | `ArtifactRegionCropTests.swift` `cropIsCappedAt2048` *(new)* | 4096×1024 image, full-width region → long edge 2048 | D-5 |
| T-6 | Swift | `ArtifactRegionCropTests.swift` `pdfRegionFollowsTheDisplayedPage` *(new)* | 2-page PDF, page 2 region renders non-nil; a page with `rotation = 90`, region (0,0)–(0.5,0.5) → aspect ratio of the rotated quarter, and a page-mark drawn at the visual top-left is inside it | AC-3.4, D-5 |
| T-7 | Swift | `ArtifactRegionCropTests.swift` `temporaryPNGIsAPNG` *(new)* | `writeTemporaryPNG` → file under the temporary directory whose first 8 bytes are the PNG signature | D-6 |
| T-8 | Swift | `CitationComposerLikenessTests.swift` `fakeStoreReturnsTheDepiction` *(new)* | add with a depiction input → returned `depiction.regionJSON` equals the input, `imageRelPath` starts with `objects/`; update with nil input keeps it | D-14 |
| T-9 | Swift | `CitationComposerLikenessTests.swift` `drawingALikenessLeavesTheCitationRegion` *(new)* | Citation with a region; start drawing, commit a region → draft holds it, `locator.region` unchanged, dialog restored; cancel → region unchanged | AC-1.4, D-7 |
| T-10 | Swift | `CitationComposerLikenessTests.swift` `newPDFCitationTakesTheDrawnPage` *(new)* | new Citation, viewer page 2, commit → `locator.page == 2`, draft page 2 | AC-3.1, D-8 |
| T-11 | Swift | `CitationComposerLikenessTests.swift` `existingPDFCitationKeepsItsPage` *(new)* | saved page 1, viewer page 2: start → viewer page 1; commit while page 2 → refused. Saved with no page → `likenessNeedsPage`, start does nothing | AC-3.2, AC-3.3, D-8 |
| T-12 | Swift | `CitationComposerLikenessTests.swift` `unsavedLikenessHoldsLeave` *(new)* | draft with a region → `shouldHoldNavigation` true | AC-1.9, D-2 |
| T-13 | Swift | `CitationComposerLikenessTests.swift` `savingALikenessSendsAPNGAndDeletesIt` *(new)* | image Citation, empty transcription, save → recorded `addObservationsToCitation … depiction=1 depictionFile=present`; region JSON has one region selector and no page; the temp file no longer exists | AC-1.5, D-4, D-6 |
| T-14 | Swift | `CitationComposerLikenessTests.swift` `nextFaceKeepsLikenessAndClearsThePerson` *(new)* | after T-13 → draft property Likeness, subject empty, no region; `initialPropertyFocus` → `.subject`; picking a person keeps Likeness | AC-1.6, D-11 |
| T-15 | Swift | `CitationComposerLikenessTests.swift` `refusedLikenessKeepsTheRegion` *(new)* | `depictionWriteError` set to the `ingest.unsupported_image` error → draft error is the mapped copy, region kept, temp file gone | AC-1.8, D-6 |
| T-16 | Swift | `CitationComposerLikenessTests.swift` `editWithoutRedrawKeepsTheImage` *(new)* | edit, change subject, save → `updateObservation` with `valueDepictionID` and no `depiction=1` | AC-2.1, D-9 |
| T-17 | Swift | `CitationComposerLikenessTests.swift` `redrawSendsANewImage` *(new)* | edit, redraw, save → `updateObservation … depiction=1 depictionFile=present` | AC-2.2, D-9 |
| T-18 | Swift | `CitationComposerLikenessTests.swift` `duplicateCopiesTheRegionNotTheImage` *(new)* | duplicate a saved likeness → same subject, property, polarity, region; empty `valueDepictionID`; no store write; focus `.subject` | AC-2.3, D-10 |
| T-19 | Swift | `CitationComposerLikenessTests.swift` `savedLikenessLineCarriesItsImage` *(new)* | listed likeness with `imageRelPath "objects/ab/cd.png"` → its `CitationSavedLine.imageRelPath` is that path, `valueText` empty | AC-1.7, D-12 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` (`saveKeepsTheSubject`, `incompatibleSubjectClearsPropertyAndDisablesSave`, `propertyDraftWalksSubjectThenPropertyThenValue`) and `CitationAutoTranscribeTests` after the fixture move.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); any check not run on this machine is reported as not run (D-16)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory and the agreed board; no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and Slice 8 line; completed.md: one short paragraph of what shipped
- [ ] S10-D15 moved to design/archive/ and design/README.md moves it to Done
- [ ] docs/ideas/depictions-and-likenesses.md: "every saved face on the photo" added; JPEG corrected to PNG
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, in a project created after S10-32 (so it has Likeness), with a Source that has a group photo Artifact (`wedding.jpg`) and a two-page PDF Artifact.

1. Open the composer on the photo, New citation, empty transcription. Add an observation on Ada: Likeness is in the property list. Draw a rectangle round her face: the dialog returns with her face preview. The Citation's locator line is unchanged.
2. Save: the dialog is ready for the next face with Likeness kept and the person pick focused. Add Ben the same way.
3. Switch the composer to the PDF with no Citation chosen, go to page 2, add Likeness on Clara, draw: the new Citation's locator shows page 2. On an existing Citation with page 1, start drawing from page 2: the viewer goes to page 1.
4. Leave with a drawn, unsaved face: the leave confirm asks.
5. Saved lines for Ada and Ben show their faces. VoiceOver reads "Likeness, Ada …" and "Likeness of Ada".
6. Edit Ben's likeness, redraw it, save: the line shows the new face. Duplicate Ada's: the dialog has her region; change the person to Clara and save: Ada's line is unchanged.
7. On an audio Artifact, Likeness is not offered.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `pvFormDialog` calls its binding setter with false when the composer hides it for drawing, which would discard the draft | §6 macOS: the setters ignore false during `.drawingLikeness`; T-9 covers the model, manual check step 1 the view |
| `PDFPage.thumbnail(of:for:)` does not draw the rotated page the way the overlay normalizes | T-6 compares against a mark at the visual top-left |
| A large photo's full `CGImage` costs memory on the main actor at save | The crop is computed from the already-loaded `displayImage`, as Auto Transcribe does |
| A temp file leaks on a crash between write and delete | Files are in the system temporary directory, named `provenencia-likeness-*`; accepted |

**Stop and report if:**

- The agreed board keeps the observation dialog on screen while the face is drawn, or draws the face inside the dialog: D-7 no longer holds.
- The agreed board shows saved faces on the photo, or offers a likeness outside the observation dialog: that is Out.
- `ArtifactViewer` cannot take a region from `viewerRegion` without a change to `ArtifactRegionOverlayView`.
- Go refuses a region JSON from `CitationLocatorDraft.encodeJSON()` (`depictionvalues.invalid` in T-13 against `GoStore`, or in the manual check).
- `ObservationRow.loaded(from:)` or `asCatalogObservation` no longer match §4 because S10-52 or another PR touched them.
- `saveKeepsTheSubject` or `incompatibleSubjectClearsPropertyAndDisablesSave` cannot pass with D-11's subject rule.

## 12. Open questions

- **Board not agreed.** S10-D15 is written; the user sends the Claude Design link when the board is agreed. Then §6 macOS gains a word description of each frame, the copy in §6 Strings is checked against it, and the plan goes through `review-plan`.

## 13. PR

**Title:** S10-34 — Draw a likeness in the citation composer
**Body:** objective in one sentence; what shipped against US-1 to US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-34-composer-likeness.md`.
