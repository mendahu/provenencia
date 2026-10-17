# S10-34 — Composer depiction row: draw a face, save it, see it

**Status:** Landed in [#442](https://github.com/mendahu/provenencia/pull/442) (2026-10-17)
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-34--composer-depiction-row)
**Brief:** [S10-D15](../design/archive/S10-D15-composer-depiction-row.md), board agreed 2026-10-17
**Board:** Claude Design project *Citation Composer*, file *Citation Composer.dc.html*, section "Depiction row" (Enhancement · S10-D15 → PR S10-34), frames D1a–D10 and D5c: <https://claude.ai/design/p/8b120853-4ba4-4c0d-8492-75e76b9f9b7a?via=share> (the implementer opens it locally). §6 "Frames" describes each frame in words, and the plan was checked against an export of the board
**Depends on:** S10-16 (landed [#384](https://github.com/mendahu/provenencia/pull/384)), S10-54 (landed [#433](https://github.com/mendahu/provenencia/pull/433)), S10-D15; S10-52 (landed [#437](https://github.com/mendahu/provenencia/pull/437); it changed no file this plan names)
**Blocks:** —
**Branch:** `s10-34-composer-depiction-row`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** —
**Context:** [S10-D15](../design/archive/S10-D15-composer-depiction-row.md) CDR-1 to CDR-14; [depictions and likenesses](../depictions-and-likenesses.md) ("Shape": rendering hangs off the value type, the key decides where it shows; "The image"); [S10-54 plan](S10-54-depiction-ffi.md) (store types) and its Landed note; [S10-D3](../design/archive/S10-D3-citation-composer.md) CC-3, CC-4, CC-12; working rules "A depiction's value is its own row: region plus image", "The client renders the image; Go validates it and stores its own copy", "`likeness` is new catalogs only"
**Planned:** 2026-10-17 against `main` @ `cf1f051` (S10-52 and S10-54 landed); first drafted against `ac05ae8`
**Review:** [`S10-34-composer-depiction-row.review.md`](S10-34-composer-depiction-row.review.md) — round 1: Fits with changes / Not ready; round 2 (after the board): Fits with changes / Ready with changes → changes applied

## Landed

Differences from this plan: The evidence graph card draws a Likeness row's thumbnail (`EvidenceCardView`, `ThumbnailCache.cachedImage`) and a saved likeness line centres its label against the thumbnail, both added at the user's request after the PR went ready (outside §4/§6). Set page is disabled while drawing by a guard in `setPageFromViewer` instead of a greyed button, so `ArtifactViewer`'s API is unchanged. `depictionWrongPage` is computed (drawing, a PDF, a saved or unsaved Citation page, and a different viewer page) instead of a stored flag. On the wrong page (or a PDF Citation with no page) the model hands the viewer overlay no armed tool (`viewerArmedRegionTool`), so the overlay never commits there; the commit guard stays as a defence. `FakeStore.recordedDepictionPaths` was added so tests can assert the temporary file is gone. `ArtifactFileFixtures.writePNG` draws through a bitmap context, not `NSImage.lockFocus`, so a fixture's pixel size does not follow the display's backing scale. The 4096 page cap, `scaledToFit`, and `writeTemporaryPNG` are in `ArtifactRegionCrop` as D-5 and D-6 say; `ArtifactRegionCrop.render` is `@MainActor` because it reads the viewer model. The code and tests are one commit, since the Xcode project lists every new file at once. The manual check in section 10 was not run: it needs a person at the Mac. The Xcode tests pass in `scripts/check-pr.sh`.

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

The engine and the FFI can already store a depiction: a region on the Citation's page and an image Go re-encodes from the app's crop (S10-32, S10-54). Nothing in the app offers one. This PR lets the researcher choose **Likeness** on a person in the citation composer, draw the face on the photo or the PDF page with the viewer's existing region tools, see it before saving, save it with an empty transcription, and see the saved face on its line. Editing can redraw the face, and Duplicate copies its region. The app renders the crop (a pixel crop of an image, a PDFKit render of a PDF page) to a temporary PNG, sends its path, and deletes the file after the call. Every path keys on the `depiction` value type; Likeness is the one seeded depiction Property, so it is the one the researcher sees.

**Why now.** It is the last piece between the engine work and the gallery (S10-35): without it no likeness can be entered, so the Person page has nothing to show.

**What stays.** The composer layout from S10-16, the Citation chooser and its locator line, Auto Transcribe and Paste, bridges, every other value type, and the leave confirm. The Citation's own region stays the Citation's. The engine, the FFI, and the proto are unchanged.

## 2. User stories

**US-1.** As a researcher with a group photo, when I add an observation on a person, I can choose Likeness, draw that person's face on the photo, and save it, so that the face is evidence on that person.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A catalog with `likeness` bound to person; an image Artifact; the dialog on person Ada | The property options for Ada include Likeness (CDR-1) |
| AC-1.2 | The same on an audio Artifact, or a catalog with no depiction Property | Likeness is not offered (CDR-1) |
| AC-1.3 | Likeness chosen, no region drawn | Save is disabled (CDR-4) |
| AC-1.4 | Draw starts, a rectangle (0.25,0.25)–(0.75,0.75) is committed on an 8×8 photo | The dialog is restored; the draft holds that region; the Citation's locator region is unchanged (CDR-3); the preview is a 4×4 image (CDR-4) |
| AC-1.5 | Save with an empty Citation transcription | One `addObservationsToCitation` call carries `depiction.regionJSON` with one region selector and no page selector, and `imagePath` names a PNG that exists during the call; after the call the file is gone (CDR-5) |
| AC-1.6 | After AC-1.5 succeeds | The dialog stays open: property Likeness, subject empty, no region; the opening focus is the subject pick (CDR-6) |
| AC-1.7 | The saved observation is listed with `depiction.imageRelPath = "objects/ab/cd.png"` | Its line's `imageRelPath` is that path and its value text is empty (CDR-7) |
| AC-1.8 | The store throws `ingest.unsupported_image` | The dialog stays open with the region kept and the mapped error copy as its error; the temporary file is gone (CDR-11) |
| AC-1.9 | A region is drawn and not saved; navigation is requested | The leave guard holds; Keep editing returns to the dialog (CDR-12) |
| AC-1.10 | While drawing | Citation chooser, artifact switch, Auto Transcribe, Paste, Clear, and Set page are disabled (CDR-3) |

**US-2.** As a researcher, when a likeness is wrong, I can change its person or redraw it, or duplicate it for another face, without losing the original.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Edit a saved likeness, change only the subject, save | `updateObservation` gets `observation.valueDepictionID` equal to the saved id and no depiction input (CDR-8) |
| AC-2.2 | Edit a saved likeness, redraw the region, save | `updateObservation` gets a depiction input whose region is the new one and whose file exists during the call (CDR-8) |
| AC-2.3 | Duplicate a saved likeness | The draft has the same subject, property, polarity, and region, an empty `valueDepictionID`, and is not saved; the opening focus is the subject pick (CDR-9) |
| AC-2.4 | Edit a saved likeness whose region Go hid (`depiction.regionJSON` empty) | The dialog shows the redraw Callout and Save is disabled until a region is drawn (CDR-14) |

**US-3.** As a researcher on a PDF, the face lands on the Citation's page.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A new Citation on a 2-page PDF, viewer on page 2, region committed | The Citation locator's page becomes 2; the region JSON's page selector is 2 (CDR-10) |
| AC-3.2 | An existing Citation saved with page 1, viewer on page 2, Draw | The viewer moves to page 1; a commit while the viewer is on page 2 is refused and the wrong-page hint shows (CDR-10) |
| AC-3.3 | An existing Citation saved with no page, Likeness chosen | The needs-page Callout shows; Draw is disabled (CDR-10) |
| AC-3.4 | A 100×200 pt page with a black square at its unrotated top-left and `rotation = 90`; region (0.5,0)–(1,0.5) | The render contains black; region (0,0)–(0.5,0.5) does not; each render is twice as wide as tall (D-5) |

## 3. Scope

| In | Covered by |
| --- | --- |
| Depiction Properties offered on an image or PDF Artifact (CDR-1) | STEP-1, AC-1.1, AC-1.2 |
| Drawing with the existing region tools; the Citation's region and controls untouched (CDR-2, CDR-3) | STEP-4, STEP-8, AC-1.4, AC-1.10 |
| Preview and Save gating (CDR-4) | STEP-2, STEP-4, STEP-8, AC-1.3, AC-1.4 |
| Save with an empty transcription; temporary PNG written and deleted (CDR-5) | STEP-5, AC-1.5 |
| Next face (CDR-6) | STEP-5, AC-1.6 |
| Saved line shows the face (CDR-7) | STEP-9, AC-1.7 |
| Edit with redraw (CDR-8) | STEP-6, AC-2.1, AC-2.2 |
| Duplicate copies the region (CDR-9) | STEP-7, AC-2.3 |
| PDF page rule (CDR-10) | STEP-4, AC-3.1–AC-3.3 |
| Refused save keeps the region (CDR-11) | STEP-5, AC-1.8 |
| Leave holds an unsaved region (CDR-12) | STEP-1, STEP-4, AC-1.9 |
| VoiceOver labels (CDR-13) | STEP-8, STEP-9, manual check |
| A hidden region asks for a redraw (CDR-14) | STEP-1, STEP-8, AC-2.4 |
| One region-crop helper for Auto Transcribe and depictions | STEP-2 |
| FakeStore stores and returns depictions | STEP-3 |
| Delete the dead property-menu cache and `CitationObservationRows.applySubject` | STEP-1 |

| Out | Home |
| --- | --- |
| Person gallery and Persons list thumb | S10-35 |
| Preferred portrait | S10-39 |
| Promote's side-by-side likenesses | S11-20 |
| Showing every saved face on the photo at once | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) (this PR adds the line) |
| A polygon-masked, transparent crop | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| Place, event, and signature keys | [ideas: depictions and likenesses](../../../ideas/depictions-and-likenesses.md) |
| Guarding a Citation page change after depictions are saved on it | dogfood log ([`docs/dogfood/ux.md`](../../../dogfood/ux.md)); Go already hides a region that no longer matches, and CDR-14 asks for a redraw |
| Folding the other test copies of PNG / PDF writers and `waitUntil` into the shared fixtures | dogfood log ([`docs/dogfood/ux.md`](../../../dogfood/ux.md)) under tooling |
| S10-D12's stale "A depiction's value is its Citation" lines | S10-35 plan, which reads that brief |
| Engine, FFI, or proto changes | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Value types the composer offers | `macos/App/Features/CitationComposer/CitationComposerVocabulary.swift` `supportedValueTypes`, `listedProperties(typeKey:properties:rules:)`, `propertyOptions(forSubjectID:)` | Text, integer, date, term, name. Depiction is filtered out, so Likeness never appears (seeded in `core/database/subjectvocab/registry.go`, key `likeness`, person, locked) |
| Dead menu cache | `CitationComposerVocabulary.swift` `propertyComboOptionsByTypeID`, `propertyComboOptions(forSubjectID:)`; `macos/App/Features/CitationComposer/CitationComposerModel.swift` `propertyOptions(for:)` | `propertyOptions(for:)` has no caller; it is the only reader of the cache, which `make(...)` builds |
| Dead row helper | `macos/App/Features/CitationComposer/CitationObservationRows.swift` `applySubject(rowID:subjectID:vocabulary:)` | No caller |
| Draft values | `CitationObservationRows.swift` `ObservationRowValues`, `ObservationRow.loaded(from:)`, `valueFields()`, `asDraft(property:)`, `asCatalogObservation(property:)` | No depiction value. `loaded` drops `valueDepictionID` and `depiction`; `asCatalogObservation` omits `valueDepictionID`, so a likeness update would fail Go's `resolve` (`core/database/observations/observations.go`) with `ErrInvalid` |
| Clearing a value | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `clearObservationValue(_:)`; `CitationObservationRows.swift` `clearValue(_:)` | Two copies that already differ: only `clearValue` resets `valueSubjectID` |
| Value logic | `macos/App/Features/CitationComposer/CitationObservationValue.swift` `Fields`, `fields(from:)`, `summary`, `isDialogValid`, `apply` | `default` branches: summary `""`, dialog invalid, apply `.unsupported` |
| Save gating | `CitationComposerModel+Authoring.swift` `authoringCanConfirm`, `editCanConfirm`, `propertyValuesComplete(_:vocabulary:)` | Builds `CitationObservationValue.Fields` by hand |
| Subject change in the dialog | `CitationComposerModel+Authoring.swift` `applyObservationSubject(_:field:values:)` | Clears the property and value when the subject's type changes, including from an empty subject — that would drop a kept Likeness (CDR-6) |
| Citation change in the dialog | `CitationComposerModel.swift` `setAuthoringCitation(_:)`, `preservesAuthoringOnCitationChange` | The open draft survives a Citation change |
| New observation save | `CitationComposerModel+Authoring.swift` `savePropertyDraft(_:)`, `ensuredCitationID()` | Builds a `CatalogObservationDraft` through `ObservationRow.asDraft`, creates the Citation first when new (`CitationComposerContext.ensureCitationID` → `createCitationWithObservations` with `fields.values.locator`), then `addObservationsToCitation`; resets the draft to the same subject and keeps the dialog open |
| Edit save | `CitationComposerModel+Authoring.swift` `saveEdit()`; `CitationObservationRows.swift` `commit(rowID:citation:)`, `performCommit(rowID:context:citation:)` | Three-argument `updateObservation`; the add branch never attaches a depiction |
| Duplicate | `CitationComposerModel+Authoring.swift` `duplicate(line:)` | Copies `row.values` into a new `CitationPropertyDraft` with `isDuplicate: true` |
| Dialog presentation | `CitationComposerModel.swift` `ComposerPresented`, `ComposerBase`, `show(_:)`, `suspendedDialog()`, `restore(_:)`, `restorePresented(_:)` | `customTerm(_:over: ComposerBase)` and `newSubject(_:over:)` suspend the observation or edit dialog; `suspendedDialog()` returns them as they are, `restorePresented` shows them again |
| Opening focus | `macos/App/Features/CitationComposer/CitationComposerAuthoring.swift` `CitationAuthoringNavigation.initialPropertyFocus(subjectID:)`; call sites `CitationComposerAuthoringBox.swift` `liveOpeningFocus` and `CitationComposerSubjectGroups.swift` `CitationComposerEditSheet` opening focus, plus the test `macos/ProvenenciaTests/CitationComposerModelTests.swift` (the one-argument call); `CitationComposerAuthoringBox.swift` `propertyFields(_:)` `.onChange(of: draft.isSaving)` sets the literal `.property` | Subject when empty, else property; after a save, `.property` |
| Value controls | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationPropertyFields` `valueControls`, `propertyOptions` | Switch on value type; `default` shows a disabled input. Shared with `CitationComposerEditSheet.propertyForm` (`CitationComposerSubjectGroups.swift`) |
| Viewer | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` `body`, `clearMenuButton`, `locatorStrip`; `macos/App/Features/ArtifactViewer/ArtifactViewer.swift` `ArtifactViewer`, `ArtifactViewerToolChrome` | Passes `committedRegion: model.locator.region`, `onCommitRegion: { model.setRegion($0) }`; one region at a time; Clear, Set page, and the region radios sit in the strip |
| Citation region | `CitationComposerModel.swift` `setRegion(_:)`, `armedRegionTool`, `disarmRegionTool()`, `setPage(_:)`, `isTranscribing`, `locatorCapabilities` | Drawing writes the Citation's locator through `CitationLocatorDraft.setRegion(_:capabilities:viewerPage:)`, which stamps the viewer page on a PDF; `setRegion` disarms the tool |
| Region JSON | `macos/App/Features/CitationComposer/CitationLocatorDraft.swift` `CitationLocatorDraft` (`page`, `region`, `extras`), `setRegion`, `encodeJSON()`, `decode(_:)` | `core/locator/depiction.go` `ValidateDepictionRegion` accepts that shape with one region selector and a leading artifact selector |
| Region crop | `macos/App/Features/OCR/OCRImage.swift` `cgImage(from:)`, `crop(_:pixelRect:)`, `crop(_:normalizedRect:)`, `isOversized(_:)`; `macos/App/Features/CitationComposer/CitationTranscriptionModel.swift` `runAutoTranscribe()` | Auto Transcribe chains `cgImage(from:)` → `ArtifactRegionGeometry.boundingRect(of:)` → `crop(_:normalizedRect:)` |
| PDF region render | — | None. `macos/App/Features/ArtifactViewer/ArtifactPDFViewport.swift` `layoutPageOverlay()` normalizes against the media box as displayed (rotation applied) |
| Viewer content | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `kind`, `displayImage`, `pdfDocument`, `page` (1-based), `setPage(_:)`, `load(_:)` | The loaded file; no URL kept |
| PNG writer, temp file | — | None in the app (searched `CGImageDestination`, `representation(using:`) |
| Saved lines | `CitationComposerAuthoring.swift` `CitationSavedLine`, `CitationComposerAuthoring.groups(observations:connections:vocabulary:nameOrders:)`; `CitationComposerSubjectGroups.swift` `lineRow(_:rule:)` | A likeness line reads "Likeness —" with no image; the row is `.accessibilityElement(children: .combine)` |
| Saved image display | `macos/App/Platform/CachedThumbnail.swift` `CachedThumbnail`; `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.Content(image:glyph:loading:label:)` | Loads `objects/…` through `ThumbnailCache` |
| Store types | `macos/App/Platform/GenealogyStore.swift` `CatalogDepictionValueInput`, `CatalogDepictionValue`, `CatalogObservation.valueDepictionID` / `.depiction`, `CatalogObservationDraft.depiction`, `updateObservation(projectDir:userID:observation:depictionInput:)` | Landed in S10-54 |
| Fake store | `macos/App/Platform/FakeStore.swift` `addObservationsToCitation`, `appendFakeObservations`, `fakeObservation(citationID:id:ref:row:)`, `updateObservation(…depictionInput:)`, error hooks `updateSourceError`, `unpromoteSubjectError` | Ignores depictions; records `addObservationsToCitation citationID=… observations=N` and `updateObservation id=…` |
| Errors | `macos/App/Platform/CoreInvoke.swift` `CoreInvokeError.coded(status:code:kind:params:)`; `macos/App/Platform/L10n.swift` `L10n.Errors.ingestUnsupportedImage`, `depictionValuesInvalid` | Landed in S10-54 |
| Tests | `macos/ProvenenciaTests/CitationComposerModelTests.swift` (`makeStore`, `makeModel`, `saveKeepsTheSubject`, `incompatibleSubjectClearsPropertyAndDisablesSave`, `propertyDraftWalksSubjectThenPropertyThenValue`, `duplicateDoesNotWriteAndConfirmsADirtyDraft`, `saveEditClosesTheDialog`); `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift` (`regionCropIsSmallerThanFullImage`, private `uniqueProjectDir`, `writePNG`, `writeBlankPDF`, `seedArtifact`); `macos/ProvenenciaTests/OCRImageTests.swift`; `macos/ProvenenciaTests/ArtifactViewerModelTests.swift` (`loadsPDFAndPages`) | No depiction authoring test |

**Precedent.** The S10-54 plan and [#433](https://github.com/mendahu/provenencia/pull/433) for the store types this PR fills. Auto Transcribe's region crop is the image half of D-5; it moves into the shared helper instead of being copied. `ComposerPresented.customTerm(_:over:)` / `.newSubject(_:over:)` are the precedent for suspending the observation dialog and returning to it (D-7). `DateValueDraft` / `NameValueDraft` on `ObservationRowValues` are the precedent for one structured draft per value type (D-2).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `supportedValueTypes` gains `depiction`. `propertyOptions(forSubjectID:capabilities: ArtifactLocatorCapabilities)` replaces `propertyOptions(forSubjectID:)`; `listedProperties` drops a Property whose value type needs a capability the Artifact lacks, through one rule `static func isOffered(valueType: String, capabilities: ArtifactLocatorCapabilities) -> Bool` *(new)*: `depiction` needs `supportsRegionLocator`, every other supported type needs nothing. `CitationPropertyFields.propertyOptions` and `applyObservationSubject` pass `model.depictionCapabilities` *(new)*: `ArtifactViewerKind.classify(mediaType: selectedArtifact?.file?.mediaType ?? "").locatorCapabilities`, so the list does not change while the viewer is still loading. The dead chain (`CitationComposerModel.propertyOptions(for:)`, `propertyComboOptions(forSubjectID:)`, `propertyComboOptionsByTypeID` and its loop in `make(...)`) and the dead `CitationObservationRows.applySubject` are deleted. | The value type decides the control, the key decides where it shows (depictions doc). One rule table keeps the next gated type (a video frame) one entry. Deleting the cache leaves one property list | A `likeness` key check. A `Bool` flag with a hard-coded `true` caller. Keeping the cache, which would never list Likeness |
| D-2 | `DepictionValueDraft` *(new)*, `macos/App/Features/CitationComposer/DepictionValueDraft.swift`: `var locator: CitationLocatorDraft`, `var valueDepictionID: String`, `var imageRelPath: String`, `var regionHidden: Bool { !valueDepictionID.isEmpty && !hasRegion }` (computed: a saved value whose region Go hid, until a region is drawn), `static func empty()`, `init(from: CatalogObservation)` (locator through `CitationLocatorDraft.decode(depiction.regionJSON)`), `var hasRegion: Bool` (`locator.region?.isValid == true`). `ObservationRowValues` and `CitationObservationValue.Fields` gain `depictionDraft: DepictionValueDraft = .empty()`; `fields(from:)`, `valueFields()`, `propertyValuesComplete`, and `ObservationRow.loaded(from:)` carry it; `asCatalogObservation` sets `valueDepictionID`. `ObservationRowValues.clearValue()` *(new)* resets every value field including `valueSubjectID` and `depictionDraft`; `clearObservationValue` and `CitationObservationRows.clearValue` call it. | Same shape as `dateDraft` / `nameDraft`. `CitationLocatorDraft` already owns region validation, the PDF page stamp, the JSON round trip, and unknown extras. `ObservationRowValues` is `Equatable`, so a drawn region holds leave and makes the edit dirty with no new flag (CDR-12). One clearer ends the existing drift | Four loose fields and a second page-stamp rule. Two clearers |
| D-3 | `CitationObservationValue`: `summary` returns `""` for depiction. `isDialogValid` requires `depictionDraft.hasRegion`. `apply` returns `nil` when `hasRegion` and `.unsupported` otherwise, and does not set `draft.depiction`; the model attaches the input after rendering (D-6). | `apply` is pure and cannot render. One value switch keeps every type's rules together | Rendering inside `apply` |
| D-4 | The region JSON is `depictionDraft.locator.encodeJSON()`. Image: no page selector. PDF: the page from D-8. | Reuses the Citation's encoder; extras round-trip | A second encoder |
| D-5 | `ArtifactRegionCrop` *(new)*, `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift`, owns cutting a region from viewer content. `cgImage(from:)`, `crop(_:pixelRect:)`, and `crop(_:normalizedRect:)` move here from `OCRImage`; `OCRImage` keeps `isOversized` and its constants. `static func boundingBox(of image: CGImage, region: ArtifactRegionDraft) -> CGImage?` *(new)* is the uncapped crop; `runAutoTranscribe` calls it. `static func render(_ viewer: ArtifactViewerModel, page: Int?, region: ArtifactRegionDraft) -> CGImage?` *(new)*: image → `cgImage(from: viewer.displayImage)` → `boundingBox`; PDF → `viewer.pdfDocument?.page(at: page - 1)`, visual size = `bounds(for: .mediaBox).size` with width and height swapped when `rotation` is 90 or 270, bbox size in points = normalized bbox × visual size, `s = min(4, 2048 / bbox long edge, 4096 / visual long edge)`, `page.thumbnail(of: visual size × s, for: .mediaBox)` → `cgImage(from:)` → `boundingBox`; any other `kind` → nil. Both paths then call `scaledToFit(_:maxLongEdge:)` *(new)* (a `CGContext` in sRGB, `premultipliedLast`, `.high` interpolation) with `maxLongEdge = 2048`, whose doc comment cites `core/ingest/stage.go` `stageMaxEdge`. | One way to cut a region from an Artifact: the polygon mask (ideas) and a later PDF Auto Transcribe change one function. `thumbnail(of:for:)` draws in the displayed orientation, the space the overlay normalizes against. Cropping in `CGImage` pixels makes the result independent of backing scale. The 4096 page cap keeps the bitmap at or under 64 MB | A new type beside `OCRImage` that copies the crop chain. A hand-built CTM for rotation |
| D-6 | `ArtifactRegionCrop.writeTemporaryPNG(_:) throws -> URL` *(new)* encodes with `CGImageDestination` (`UTType.png`) to `FileManager.default.temporaryDirectory/provenencia-depiction-<UUID>.png` and throws `ArtifactRegionCropError.encodeFailed` *(new)* when the destination cannot be created or `CGImageDestinationFinalize` returns false. One model helper `withDepictionInput<T>(for values: ObservationRowValues, perform: (CatalogDepictionValueInput) async throws -> T) async throws -> T` *(new)* renders (D-5) with `artifactViewer`, writes, calls `perform`, and removes the file in a `defer`, success or failure. The render runs first, before `ensuredCitationID()`, so a failed render creates no Citation. A nil render or a write error throws `DepictionRenderFailed` *(new)*. `savePropertyDraft` catches it with its own `catch is DepictionRenderFailed` (before the generic catch), sets `draft.isSaving = false` and `draft.error = depictionRenderFailed`. `saveEdit` wraps `withDepictionInput` in a `do`/`catch`; on a throw it sets `session.isSaving = false`, `session.error` to `depictionRenderFailed` (or the mapped error), writes `editSession = session`, and returns without committing. | Go never deletes the source (`ingest.StageImage`); the app owns its temp file in one place | Rendering at draw time and holding the file until save — a discarded draft would leak files |
| D-7 | Drawing is a presented state: `ComposerPresented.drawingDepiction(over: ComposerBase)` *(new)*, `over` only `.authoring` or `.edit`. `suspendedDialog()` returns it as it is (the `customTerm`/`newSubject` arm); `restorePresented` shows it again. `startDepictionDrawing()` *(new)* does nothing while `isTranscribing` or `depictionNeedsPage`; otherwise it applies D-8's page, arms `armedRegionTool = .rectangle` when none is armed, and shows `.drawingDepiction(over:)`. `currentDepictionDraft` *(new)* is the `depictionDraft` of the open property draft (`authoring`) or edit session (`editSession`, `.property`) when its Property's value type is `depiction`, else nil. `viewerRegion` *(new)* is `currentDepictionDraft?.locator.region` when `currentDepictionDraft != nil`, else `locator.region`; `commitViewerRegion(_:)` *(new)* writes into `currentDepictionDraft` through `locator.setRegion(_:capabilities:viewerPage:)` (after D-8's guard), disarms the tool, and restores the dialog; with no depiction draft it is `setRegion`. `cancelDepictionDrawing()` *(new)* (Cancel, Escape, or `disarmRegionTool` while drawing) restores the dialog with the region unchanged. `isDrawingDepiction` *(new)* disables the Citation chooser, the artifact switch, Auto Transcribe, Paste, Clear, Set page, Delete citation, Save citation, Add observation, Done, each subject group's add button, and each line's Edit, Duplicate, and Delete (and the line tap). The model guards the same actions: `openAuthoring`, `beginEdit`, `askDeleteLine`, `duplicate(line:)`, and `selectCitation` return without effect while drawing. The label, transcription, uncertain, and note fields stay editable; page navigation, zoom, and the region radios stay enabled. `startDepictionDrawing` closes the find row (`artifactViewer.closeFind()`). `base(of:)` returns `over` for `.drawingDepiction`. Escape: while drawing, the Cancel button carries `.keyboardShortcut(.cancelAction)`. `authoringSheetBinding` and `editSheetBinding` read false and ignore a false write while drawing. | The dialog is a modal sheet over the viewer, and the brief forbids a second viewer. `over: ComposerBase` admits only the states drawing can suspend and reuses the restore arms. Showing the draft's region whenever a depiction draft is open makes edit show its face (CDR-8) | `returnTo: ComposerPresented` — admits drawing over a confirm. A region layer on `ArtifactRegionOverlayView` — saved faces are Out |
| D-8 | PDF page rule (answered 2026-10-17). An existing Citation is `context.citationID?.isEmpty == false`. `depictionPage` *(new)* is `fields.baseline.locator.page` for an existing Citation, else `fields.locator.page`. `depictionNeedsPage` *(new)* is true for an existing PDF Citation whose saved page is nil; the dialog shows the Callout `depictionNeedsPage` and Draw is disabled. Setting the page in the viewer changes only the unsaved Citation fields, so the Callout says to save it (§6 Strings; the board's shorter copy left that step out). With a page: `startDepictionDrawing` calls `artifactViewer.setPage(page)`, and `commitViewerRegion` refuses (no write, tool stays armed, `depictionWrongPage` *(new)* is true, the hint shows `depictionWrongPage(page)`) while `artifactViewer.page != page`. New Citation with no page: the commit calls `setPage(artifactViewer.page)` on the Citation locator first. A new Citation with no page also has no region, because `CitationLocatorDraft.setRegion` stamps a page on a PDF. Images: no page, no rule. | `ValidateDepictionRegion` requires the region's page to equal the Citation's saved page; `ensuredCitationID` creates a new Citation with `fields.values.locator`, so the stamped page is saved with it | Refusing on every Citation without a page |
| D-9 | Edit: `CitationObservationRows.commit(rowID:citation:depictionInput:)` and `performCommit` gain `depictionInput: CatalogDepictionValueInput? = nil`; the update branch passes it to `updateObservation(…depictionInput:)`, and the add branch sets `draft.depiction`. `isDepictionRedrawn` *(new)* on the model is true when the edit session's `.property` values have `depictionDraft.locator != propertyBaseline.depictionDraft.locator`. `saveEdit` wraps the commit in `withDepictionInput` only when `isDepictionRedrawn`; otherwise it commits with nil and the row keeps `valueDepictionID`. | Redraw is allowed (answered 2026-10-17). Go replaces the image only when an input is sent. The add branch stays correct if a later caller reaches it | Always re-rendering |
| D-10 | Duplicate copies `row.values` and sets `depictionDraft.valueDepictionID` and `imageRelPath` to `""`; the locator stays (answered 2026-10-17). | CC-12 for every saved row; the copy owns its own image | Hiding Duplicate |
| D-11 | Next face (CDR-6): after a depiction saves, `savePropertyDraft` resets to `ObservationRowValues` with only `propertyID` kept; every other value type keeps the subject, as today. `applyObservationSubject` clears the property and its value only when the property is not in `propertyOptions(forSubjectID: raw, capabilities: locatorCapabilities)`; a kept property keeps its value. Focus: `initialPropertyFocus(subjectID:propertyID:isDepiction:)` (the two new parameters default to `""` and `false`, so the existing test call compiles unchanged) returns `.subject` when the subject is empty, `.property` when the property is empty, `.subject` when `isDepiction`, else `.property`. `isDepiction` comes from `CitationComposerVocabulary.isDepiction(propertyID:)` *(new)* (`property(id:)?.valueType == PropertyValueType.depiction.rawValue`). The two call sites in §4 pass it; the post-save `.onChange` calls the same function instead of the literal `.property`. | A group photo is a run of faces, one per person. The membership rule is the one the property list shows | Keeping the subject — every next face is another person |
| D-12 | `setAuthoringCitation(_:)` resets `depictionDraft` to `.empty()` on the open property draft when the Citation changes. The artifact switch already resets `authoring` (`selectCitationAndLoad`). | A region belongs to its Citation's page; the render source always matches the viewer | Storing an Artifact id in the draft |
| D-13 | Saved line: `CitationSavedLine` gains `imageRelPath: String?` *(new)*, set in `groups(...)` when the row's `depictionDraft.imageRelPath` is non-empty. A depiction line uses `.padding(.vertical, PVSpacing.space1)` instead of `space3`, so the 32pt tile makes a line of about 36pt, as the board draws it. `lineRow` draws `CachedThumbnail(projectDir: model.session.projectKey.projectDir, relPath:, mediaType: "image/png", size: 32)` in place of the value text when it is set; an empty path keeps the "—". The italic "not" and the Negative badge stay. A depiction line sets `.accessibilityLabel(L10n.CitationComposer.depictionLineLabel(property, person))` on the row and `depictionThumbLabel(property, person)` on the thumbnail; `property` is `line.label`; `person` is `vocabulary.subjectsByID[line.subjectID]?.label`, and when it is nil the row and thumbnail labels are `line.label` alone. Tile size: S10-D15 CDR-7. | `CachedThumbnail` is the loader for `objects/…` and owns loading and placeholder states | `NSImage(contentsOf:)` on the render path |
| D-14 | Preview: `depictionPreview` *(new)* is a computed `NSImage?` over `currentDepictionDraft`: nil without a region, else `ArtifactRegionCrop.render` through a one-entry memo keyed by (`selectedArtifact?.id`, `locator.encodeJSON()`). The memo stores `NSImage(cgImage: rendered, size: .zero)`. Shown through `PVThumbnail(PVThumbnail.Content(image: Image(nsImage:), label: depictionPreviewLabel), size: 96)`. | Derived, so no mutation site keeps it in sync and it always matches what save renders | A stored property reset by hand at each site |
| D-15 | A depiction draft whose `regionHidden` is true shows the Callout `depictionRegionHidden` *(new)* in the value area; Save is disabled until a region is drawn (CDR-14). | Go hides a region that no longer matches the Citation's page (`depictionvalues.Readable`) | Saving without a region, which Go refuses |
| D-16 | `FakeStore` stores depictions. `appendFakeObservations` and `fakeObservation` map `valueDepictionID` and `depiction`. A draft or update input with a depiction appends ` depiction=1 depictionFile=present` or ` depictionFile=missing` (from `FileManager.default.fileExists(atPath:)`) to the recorded string, and returns `valueDepictionID = "dep-<n>"` from a private counter `depictionSequence` *(new)* and `depiction = CatalogDepictionValue(regionJSON: input.regionJSON, imageRelPath: "objects/fake/dep-<n>.png")`. An update records ` valueDepictionID=<sent id>` from the sent observation, and with no input copies the stored observation's `valueDepictionID` and `depiction`. Error hooks `addObservationsToCitationError` and `updateObservationError` *(new)* throw when set. | Model tests assert the call order, the file during the call, and the error path without Go; hooks are named after the call | A test-only store subclass |
| D-17 | Tests: `macos/ProvenenciaTests/ArtifactFileFixtures.swift` *(new)* `enum ArtifactFileFixtures` holds `uniqueProjectDir()`, `writePNG(projectDir:relPath:width:height:topHalfRed:)`, `writeBlankPDF(projectDir:relPath:pageCount:)`, `writeMarkedPDF(projectDir:relPath:rotation:)` (one 100×200 pt page, a black 10×10 pt square at the unrotated top-left), and `seedArtifact(_:mediaType:relPath:)`, moved from `CitationAutoTranscribeTests` (which calls them). Model tests go in `CitationComposerDepictionTests.swift` *(new)* with private `makeStore` / `makeModel` that follow `CitationComposerModelTests`' and add `prop-likeness` (key `likeness`, label `Likeness`, `valueType "depiction"`) bound to the person type. Render tests go in `ArtifactRegionCropTests.swift` *(new)*; the crop tests in `OCRImageTests` move there with the functions. `.agents/skills/add-swift-test/SKILL.md` gains one line: shared file fixtures live in `ArtifactFileFixtures`. | One home for Artifact files on disk | Copying the helpers |
| D-19 | `depictionCaption` *(new)* on the model picks the caption: needs page → `depictionCaptionNeedsPage`; `regionHidden` → `depictionCaptionHidden`; no region → `depictionCaptionEmpty`; `isDepictionRedrawn` → `depictionCaptionRedrawn`; otherwise `depictionCaptionBox`. | One derived value; the board draws five states | A caption stored on the draft |
| D-20 | The save-error `PVCallout` in `CitationPropertyFields` gains the title `rowStateError` ("Not saved") and moves from after Negative to directly under the value controls, for every value type. The two bridge error Callouts (`CitationComposerAuthoringBox.swift` `bridgeFields(_:)`, `CitationComposerSubjectGroups.swift` `bridgeForm(endA:endB:error:)`) gain the same title. `CitationObservationRows.performCommit`'s two fallback `markError` calls use `observationSaveFailed` *(new)* instead of `rowStateError`, so no Callout reads "Not saved" twice. The `rowStateError` comment in `L10n.swift` and `Localizable.xcstrings` becomes "Badge when a row-level Save failed; also the title of a save-error Callout". The message stays the mapped error: for `ingest.unsupported_image` it is the existing "The picture of that region couldn’t be read. Draw the region again, or try a smaller one.", longer than the board's frame. | The agreed board (D9a) places and titles it so; one error style for the composer's dialogs | A depiction-only error Callout. Shortening the existing error copy |
| D-21 | Depiction copy: the edit sheet subtitle is `editDepictionSubtitle(ref:)` for a depiction Property (every depiction edit, including a hidden region); the branch lives in `CitationComposerView.editSheetSubtitle`'s `.property` case. The duplicate note is `depictionDuplicatedFrom(person, property)`; the branch lives in `duplicateOrigin`, which takes the line (`duplicateOrigin(line:)`) instead of `subjectID` and `phrase`, so both `duplicate(line:)` call sites go through it. The leave message is `leaveDepictionDraft(property:person:ref:)`; the branch lives only in `unsavedSummary`'s `authoring.holdsLeave` arm, for a depiction draft with a region and a person, else the existing `leaveDraft`. With no person, every depiction copy falls back to the existing string. The property argument is the Property's label as stored, not lowercased, so "Likeness" keeps its capital in English; the board's lowercase is not reproduced. | The board's copy; lowercasing a label is wrong in languages that capitalize nouns | `localizedLowercase` on the label |
| D-18 | `xcodebuild`, the Swift tests, and the manual check run on a Mac. Without one, the implementer finishes the non-Xcode checks, stops before STEP-1's Swift check, and reports which checks did not run. | Precedent: S10-23 D-11 | Claiming a Swift check that was not run |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionCrop` *(new)* | type / file | `macos/App/Features/OCR/OCRImage.swift` `OCRImage` | Extends it by taking over its crop functions; OCR keeps its oversize rule |
| `ArtifactRegionCrop.writeTemporaryPNG(_:)`, `ArtifactRegionCropError` *(new)* | func / type | none (no PNG encoder in the app) | — |
| `DepictionValueDraft` *(new)* | type / file | `DateValueDraft`, `NameValueDraft` | Same pattern; wraps `CitationLocatorDraft` |
| `ObservationRowValues.clearValue()` *(new)* | func | `clearObservationValue`, `CitationObservationRows.clearValue` | Replaces their bodies |
| `CitationComposerVocabulary.isOffered(valueType:capabilities:)` *(new)* | func | `supportedValueTypes` | Adds the file-kind rule beside it |
| `ComposerPresented.drawingDepiction(over:)` *(new)* | enum case | `customTerm(_:over:)`, `newSubject(_:over:)` | Same registry and arms |
| `startDepictionDrawing`, `cancelDepictionDrawing`, `commitViewerRegion`, `viewerRegion`, `currentDepictionDraft`, `isDrawingDepiction`, `depictionPage`, `depictionNeedsPage`, `depictionWrongPage`, `depictionPreview`, `withDepictionInput` *(new)* in `CitationComposerModel+Depiction.swift` *(new)* | funcs / properties / file | `CitationComposerModel.setRegion(_:)`, `setPage(_:)` | Route the one viewer region between the Citation and the draft; `setRegion` stays the Citation's |
| `CitationSavedLine.imageRelPath` *(new)* | field | `CitationSavedLine.valueText` | An image is not text |
| `FakeStore.addObservationsToCitationError`, `updateObservationError`, `depictionSequence` *(new)* | test hooks | `FakeStore.updateSourceError`, `unpromoteSubjectError` | Same shape |
| `ArtifactFileFixtures` *(new)* | test helper | private helpers in `CitationAutoTranscribeTests` | Moved, not copied |
| `L10n.CitationComposer.depiction*` *(new)* | keys | `L10n.CitationComposer` | Same namespace |

**Builds on.** `ArtifactViewer` with its API unchanged; `CitationLocatorDraft`; `ComposerPresented` `over:` suspension; `CachedThumbnail`; the S10-54 store types; `CitationObservationValue` as the one value-rule switch.
**Deletes.** `CitationComposerModel.propertyOptions(for:)`, `CitationComposerVocabulary.propertyComboOptions(forSubjectID:)`, `propertyComboOptionsByTypeID` and its build loop; `CitationObservationRows.applySubject`; the crop functions in `OCRImage` (moved); the private file helpers in `CitationAutoTranscribeTests` (moved); the duplicate clearing bodies.
**Next adds.** A place, event, or signature depiction key is one seeded Property: every path keys on the value type. A video-frame depiction is one `isOffered` entry. S10-35 reads the images this PR saves.

### Schema

None — no catalog change.

### Go core

None — no engine change.

### FFI

None — the S10-54 fields carry everything.

### macOS

The rows follow the brief's inventory; "Frames" below describes the agreed board.

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Depiction value area | `CitationPropertyFields` `valueControls` `.depiction` case | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` | Under the existing "Value" label: an optional warning `PVCallout` (needs page, or hidden region) across the full width, then a row of a 96pt `PVThumbnail` preview (D-14; empty shows the kit's empty dash) and, to its right, a column of an italic caption (`depictionCaption` *(new)*, D-19) above a small secondary `PVButton` with `.crop` — `depictionDraw` or `depictionRedraw` → `startDepictionDrawing()`, disabled when `depictionNeedsPage`. Shared by the edit sheet |
| Save error | `PVCallout(tone: .danger, title: L10n.CitationComposer.rowStateError, message:)` | `CitationComposerAuthoringBox.swift` `CitationPropertyFields` | D-20: directly under the value controls and above Negative, for every value type |
| Viewer wiring | `ArtifactViewer` | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` | `committedRegion: model.viewerRegion`, `onCommitRegion: { model.commitViewerRegion($0) }`. The locator strip keeps reading `model.locator`. `clearMenuButton` and Set page disabled while `isDrawingDepiction` |
| Drawing hint | `PVCallout(tone: .info)` + ghost `PVButton` `depictionCancel` | `macos/App/Features/CitationComposer/CitationComposerViewerPane.swift` | A band between the locator strip and the canvas, at both widths: the Callout fills the band, Cancel sits to its right. `depictionDrawingHint`, or `depictionWrongPage(page)` with tone `.warning`. Escape calls `cancelDepictionDrawing()` |
| Saved depiction line | `CachedThumbnail` → `PVThumbnail` | `CitationComposerSubjectGroups.swift` `lineRow`; `CitationComposerAuthoring.swift` `CitationSavedLine` | D-13 |
| Dialog presentation | `pvFormDialog` | `macos/App/Features/CitationComposer/CitationComposerView.swift` | D-7 bindings; D-21 subtitle |

States: no region → Save disabled, empty preview; render failed → `depictionRenderFailed` in the save-error Callout, region kept; image loading → the kit thumbnail's skeleton; unreadable → the image placeholder glyph; existing PDF Citation with no page → warning Callout, Draw disabled; hidden region → warning Callout, Save disabled. Accessibility: D-13 labels; the preview reads `depictionPreviewLabel`; drawing stays pointer-only.

#### Frames

The agreed board (see **Board**) draws the composer as shipped by S10-16 with these changes. Sample data: a wedding photo of five, a two-page parish newsletter, Ada Lovelace, Ben Ford, Clara Ford.

- **D1a — Likeness in the pick.** The New observation dialog on Ada over the wedding photo; the property combo is open with Interests, Likeness, Name, Occupation. Value still reads "Choose a property first".
- **D1b — Likeness chosen, nothing drawn.** Subject Ada, Property Likeness. The Value field is the value area: an empty 96pt preview, the caption "Nothing drawn yet. Draw the region on the file", and **Draw on the file**. Save observation disabled. The Citation's transcription reads "No transcription yet".
- **D2a / D2b — Drawing mode, wide and narrow.** The dialog is gone. The rectangle radio is selected in the tool strip. The frame leaves Delete, Save citation, Add observation, and Done looking live; D-7 disables them. Between the locator strip and the canvas, the info Callout "Draw the region with the tools above. Escape goes back to the observation." with **Cancel** to its right. The Artifact and Citation selects, Auto transcribe, and Clear are disabled; the label and transcription fields stay as they are. The locator strip still shows the Citation's own region chip; the canvas shows only the rectangle being drawn. The Observations header reads "Unsaved: a draft observation" (existing).
- **D3 — Ready to save.** The dialog is back: Ada, Likeness, the preview of her face, the caption "The rectangle around the region, up to 2,048 px on its long edge", **Draw again**. Save observation enabled. The canvas behind shows her region.
- **D4 — Next face.** Ada's likeness line is under her with the existing Saved badge. The dialog stays open: Subject empty and focused ("Person, event or place"), Property **empty** in the frame, value area empty. See D-11: the plan keeps Likeness, as the brief and the board caption say; the frame's empty Property field is a drawing slip.
- **D5a / D5b — Saved lines, wide and narrow.** Ada: Likeness (32pt thumbnail), Occupation Bridesmaid. Ben: Likeness with the hover actions Edit, Duplicate, Delete. Clara: italic *not*, then the thumbnail, then the Negative badge. The line keeps its usual height.
- **D5c — Thumbnail states.** Ben's thumbnail loading: the kit skeleton. Clara's unreadable: the image placeholder glyph. VoiceOver: "Likeness, Ben Ford" on the line, "Likeness of Ben Ford" on the thumbnail.
- **D6a — Edit.** Edit observation, subtitle "OBS-L2BEN · saves on its own. A redraw replaces the image" (D-21). The kind chips are locked with the existing note "A saved row keeps its kind. Duplicate it to record the other". Ben's region shows on the canvas in place of the Citation's. Caption as D3; **Draw again**; Cancel and Save observation.
- **D6b — Edit, redraw.** Drawing mode over Ben's face, as D2a; the header reads "Unsaved: changes to an observation" (existing).
- **D6c — Edit, ready to save.** The new region on the canvas; the caption "Redrawn. Saving replaces the saved image".
- **D7 — Duplicate.** New observation with Ada, Likeness, her region, caption as D3, **Draw again**; the origin note "Duplicated from Ada Lovelace’s likeness. Saving cuts its own image; the original is unchanged" (D-21); the subject combo open, being changed to Clara.
- **D8a — PDF, new Citation.** Newsletter page 2, Citation "New". After the face is drawn the tool strip shows "Page 2 set" and the locator strip reads Entire document › Page 2 of 2 › the Citation's region chip. Dialog with Ben, Likeness, caption as D3.
- **D8b — PDF, wrong page.** Existing Citation on page 1, viewer on page 2 in drawing mode: the Callout is the warning "Draw on page 1, this citation's page." with Cancel.
- **D8c — PDF, no page.** Existing Citation with no page. In the value area, above the preview row, the warning Callout (copy per D-8; the frame's chip of a Citation region on a page-less PDF Citation cannot occur, so a test fixture has no region); caption "Draw is unavailable until the citation has a page"; **Draw on the file** disabled; Save disabled.
- **D9a — Save refused.** Under the value area and above Negative, the danger Callout titled "Not saved" with the mapped error (D-20: the full existing string, which ends "Draw the region again, or try a smaller one."); region and preview kept; Save enabled.
- **D9b — Hidden region.** Edit on Clara's negative likeness: the warning Callout "This region no longer matches the citation's page. Draw it again."; caption "Save waits for the region to be drawn again"; **Draw again**; Save disabled. Clara's line still shows her saved image. The board's subtitle here lacks the redraw sentence; D-21 shows it on every depiction edit.
- **D10 — Leave.** Drawing mode with a region drawn; the existing leave confirm "Leave with unsaved work?" with "The likeness of Clara Ford you drew is not saved. Everything already saved on CIT-0102 stays." (D-21), Keep editing and Discard and leave. Keep editing returns to drawing mode.

### Strings

L10n keys *(new)* in `L10n.CitationComposer`, copy from the agreed board:

- `depictionDraw` = "Draw on the file"
- `depictionRedraw` = "Draw again"
- `depictionDrawingHint` = "Draw the region with the tools above. Escape goes back to the observation."
- `depictionWrongPage(_ page: Int)` = "Draw on page %lld, this citation's page."
- `depictionCancel` = "Cancel"
- `depictionPreviewLabel` = "Image to save"
- `depictionThumbLabel(_ property: String, _ person: String)` = "%1$@ of %2$@"
- `depictionLineLabel(_ property: String, _ person: String)` = "%1$@, %2$@"
- `depictionNeedsPage` = "Set this citation's page and save the citation to draw a region." (the board reads "Set this citation's page to draw a region."; D-8)
- `observationSaveFailed` = "The observation could not be saved. Try again."
- `depictionRegionHidden` = "This region no longer matches the citation's page. Draw it again."
- `depictionRenderFailed` = "The region could not be cut from the file. Draw it again."
- `depictionCaptionEmpty` = "Nothing drawn yet. Draw the region on the file"
- `depictionCaptionBox` = "The rectangle around the region, up to 2,048 px on its long edge"
- `depictionCaptionRedrawn` = "Redrawn. Saving replaces the saved image"
- `depictionCaptionNeedsPage` = "Draw is unavailable until the citation has a page"
- `depictionCaptionHidden` = "Save waits for the region to be drawn again"
- `editDepictionSubtitle(ref:)` = "%1$@ · saves on its own. A redraw replaces the image"
- `depictionDuplicatedFrom(_ person: String, _ property: String)` = "Duplicated from %1$@’s %2$@. Saving cuts its own image; the original is unchanged"
- `leaveDepictionDraft(property:person:ref:)` = "The %1$@ of %2$@ you drew is not saved. Everything already saved on %3$@ stays."

Reused: `rowStateError` ("Not saved") as the save-error title (comment updated, D-20). Removed: none.

## 7. Steps

**STEP-1 — Values, rules, the property list, and the strings**
Files: `DepictionValueDraft.swift` *(new)*, `CitationObservationRows.swift`, `CitationObservationValue.swift`, `CitationComposerVocabulary.swift`, `CitationComposerModel.swift` (delete `propertyOptions(for:)`), `CitationComposerModel+Authoring.swift` (`applyObservationSubject`, `clearObservationValue`, `propertyValuesComplete`, `setAuthoringCitation` caller), `CitationComposerAuthoring.swift` (`initialPropertyFocus`), `macos/ProvenenciaTests/CitationComposerModelTests.swift` (only if the default parameters do not compile), `CitationComposerAuthoringBox.swift` (`propertyOptions`, `liveOpeningFocus`), `CitationComposerSubjectGroups.swift` (edit sheet opening focus), `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift` *(new)*, `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift`, `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj`, `.agents/skills/add-swift-test/SKILL.md`. Skills: `add-swift-test`, `add-localized-string`.
Do: D-1, D-2, D-3, D-11 (subject rule and focus), D-12, D-17, and every key in §6 Strings.
Check: T-1, T-2, T-3, T-12 pass; `CitationAutoTranscribeTests` and `CitationComposerModelTests` pass; `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-2 — One region crop**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` *(new)*, `macos/App/Features/OCR/OCRImage.swift`, `CitationTranscriptionModel.swift`, `macos/ProvenenciaTests/ArtifactRegionCropTests.swift` *(new)*, `macos/ProvenenciaTests/OCRImageTests.swift`, `project.pbxproj`.
Do: D-5, D-6 (the writer and error).
Check: T-4, T-5, T-6, T-7 pass; `regionCropIsSmallerThanFullImage` passes.

**STEP-3 — FakeStore depictions**
Files: `macos/App/Platform/FakeStore.swift`.
Do: D-16.
Check: T-8 passes; every suite that uses `FakeStore` passes.

**STEP-4 — Drawing state and the PDF page rule**
Files: `CitationComposerModel.swift` (`ComposerPresented`, `base(of:)`, `suspendedDialog`, `restorePresented`, `selectCitation`, `depictionCapabilities`), `CitationComposerModel+Authoring.swift` (guards in `openAuthoring`, `beginEdit`, `askDeleteLine`, `duplicate(line:)`), `CitationComposerSubjectGroups.swift` (line actions and group add disabled), `CitationComposerModel+Depiction.swift` *(new)*, `CitationComposerView.swift` (bindings), `CitationComposerViewerPane.swift` (wiring, disabled controls, hint), `CitationComposerEvidenceColumn.swift` (Auto Transcribe, Paste, chooser, Delete citation, Save citation disabled while drawing), `CitationComposerSubjectGroups.swift` `CitationComposerObservationsColumn` (Add observation, Done disabled), `project.pbxproj`.
Do: D-7, D-8, D-14.
Check: T-9, T-10, T-11, T-20 pass.

**STEP-5 — Save a new depiction**
Files: `CitationComposerModel+Authoring.swift` (`savePropertyDraft`), `CitationComposerModel+Depiction.swift` (`withDepictionInput`).
Do: D-6 (helper and call site), D-11 (reset). Attach `CatalogDepictionValueInput(regionJSON:imagePath:)` to the draft after `asDraft`.
Check: T-13, T-14, T-15 pass; `saveKeepsTheSubject` passes.

**STEP-6 — Edit and redraw**
Files: `CitationObservationRows.swift` (`commit`, `performCommit`), `CitationComposerModel+Authoring.swift` (`saveEdit`).
Do: D-9, D-15.
Check: T-16, T-17, T-21 pass; `saveEditClosesTheDialog` passes.

**STEP-7 — Duplicate**
Files: `CitationComposerModel+Authoring.swift` (`duplicate(line:)`).
Do: D-10.
Check: T-18 passes; `duplicateDoesNotWriteAndConfirmsADirtyDraft` passes.

**STEP-8 — Dialog value area and depiction copy**
Files: `CitationComposerAuthoringBox.swift` (`valueControls`, the save-error Callout, post-save focus), `CitationComposerModel+Depiction.swift` (`depictionCaption`), `CitationComposerModel.swift` (`leaveDraftDetail`, `unsavedSummary`), `CitationComposerView.swift` (`editSheetSubtitle`), `CitationComposerModel+Authoring.swift` (`duplicateOrigin(line:)`), `CitationObservationRows.swift` (`performCommit` fallback messages), `CitationComposerSubjectGroups.swift` (`bridgeForm` error title). Skill: `add-ui-component`.
Do: §6 macOS rows "Depiction value area" and "Save error", frames D1b, D3, D6a, D6c, D8c, D9a, D9b, D10; D-19, D-20, D-21.
Check: T-22, T-23 pass; the app builds in `xcodebuild`; manual check steps 1–4.

**STEP-9 — Saved depiction line**
Files: `CitationComposerAuthoring.swift` (`CitationSavedLine`, `groups`), `CitationComposerSubjectGroups.swift` (`lineRow`).
Do: D-13.
Check: T-19 passes; manual check step 5.

**STEP-10 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (map row, Slice 8 line), this plan (status, Landed note), `docs/deployment-plan/spike-10/design/archive/S10-D15-composer-depiction-row.md` (moved there from `design/`), `docs/deployment-plan/spike-10/design/README.md`, `docs/ideas/depictions-and-likenesses.md` (under "Still open", add the bullet "**Every saved face on the photo.** The composer's viewer shows one region at a time; showing every saved depiction on the photo needs a read-only region layer." and change "as a JPEG" to "as a PNG").
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Draw a face on the photo and save it as a likeness.").

## 8. Tests

Swift Testing. Model tests use `FakeStore` with `prop-likeness` bound to person (D-17) and an Artifact from `ArtifactFileFixtures`, loaded through `ArtifactViewerModel.load(_:)` the way `ArtifactViewerModelTests.loadsPDFAndPages` does. A refused save uses `CoreInvokeError.coded(status: 1, code: "ingest.unsupported_image", kind: .user, params: [])`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `CitationComposerDepictionTests.swift` `depictionIsOfferedOnImageAndPDFOnly` *(new)* | image and PDF → options include Likeness; audio → absent; catalog with no depiction Property → absent | AC-1.1, AC-1.2, D-1 |
| T-2 | Swift | `CitationComposerDepictionTests.swift` `loadedDepictionKeepsItsValue` *(new)* | `CatalogObservation` with `valueDepictionID "dep-1"` and a page-2 region JSON → `loaded` draft has region, page 2, `"dep-1"`, image path; `asCatalogObservation` keeps `"dep-1"`; empty `regionJSON` → `regionHidden` | D-2 |
| T-3 | Swift | `CitationComposerDepictionTests.swift` `depictionNeedsARegionToSave` *(new)* | Likeness, no region → `authoringCanConfirm == false`; valid region → true | AC-1.3, D-3 |
| T-4 | Swift | `ArtifactRegionCropTests.swift` `imageCropIsTheBoundingBox` *(new)* | 8×8, rectangle (0.25,0.25)–(0.75,0.75) → 4×4; top-half-red image, region in the top half → red pixel | AC-1.4, D-5 |
| T-5 | Swift | `ArtifactRegionCropTests.swift` `cropIsCappedAt2048` *(new)* | 4096×1024, full region → 2048×512 | D-5 |
| T-6 | Swift | `ArtifactRegionCropTests.swift` `pdfRegionFollowsTheDisplayedPage` *(new)* | `writeMarkedPDF(rotation: 90)`: region (0.5,0)–(1,0.5) contains black; (0,0)–(0.5,0.5) does not; width = 2 × height | AC-3.4, D-5 |
| T-7 | Swift | `ArtifactRegionCropTests.swift` `temporaryPNGIsAPNG` *(new)* | `writeTemporaryPNG` → a file in the temporary directory whose first 8 bytes are the PNG signature | D-6 |
| T-8 | Swift | `CitationComposerDepictionTests.swift` `fakeStoreReturnsTheDepiction` *(new)* | add with input → `depiction.regionJSON` equals input, `imageRelPath` starts with `objects/`; update with nil input keeps both | D-16 |
| T-9 | Swift | `CitationComposerDepictionTests.swift` `drawingLeavesTheCitationRegion` *(new)* | Citation with a region; start, commit → draft holds the new region, `locator.region` unchanged, `presented == .authoring`; start, cancel → unchanged | AC-1.4, D-7 |
| T-10 | Swift | `CitationComposerDepictionTests.swift` `newPDFCitationTakesTheDrawnPage` *(new)* | new Citation, viewer page 2, commit → `locator.page == 2`, draft page 2 | AC-3.1, D-8 |
| T-11 | Swift | `CitationComposerDepictionTests.swift` `existingPDFCitationKeepsItsPage` *(new)* | saved page 1, viewer page 2: start → viewer page 1; set viewer page 2, commit → refused, `depictionWrongPage`. Saved with no page → `depictionNeedsPage`, start does nothing | AC-3.2, AC-3.3, D-8 |
| T-12 | Swift | `CitationComposerDepictionTests.swift` `unsavedDepictionHoldsLeave` *(new)* | draft with a region → `shouldHoldNavigation` true | AC-1.9, D-2 |
| T-13 | Swift | `CitationComposerDepictionTests.swift` `savingSendsAPNGAndDeletesIt` *(new)* | image Citation, empty transcription, save → recorded call ends ` depiction=1 depictionFile=present`; region JSON has one region selector and no page; temp file gone | AC-1.5, D-4, D-6 |
| T-14 | Swift | `CitationComposerDepictionTests.swift` `nextFaceKeepsThePropertyAndClearsThePerson` *(new)* | after T-13 → property Likeness, subject empty, no region, `presented == .authoring`; opening focus `.subject`; picking a person keeps Likeness | AC-1.6, D-11 |
| T-15 | Swift | `CitationComposerDepictionTests.swift` `refusedSaveKeepsTheRegion` *(new)* | `addObservationsToCitationError` set → draft error is the mapped copy, region kept, temp file gone | AC-1.8, D-6 |
| T-16 | Swift | `CitationComposerDepictionTests.swift` `editWithoutRedrawKeepsTheImage` *(new)* | edit, change subject, save → recorded `updateObservation` contains ` valueDepictionID=<saved id>` and not ` depiction=1` | AC-2.1, D-9 |
| T-17 | Swift | `CitationComposerDepictionTests.swift` `redrawSendsANewImage` *(new)* | edit, redraw, save → `updateObservation … depiction=1 depictionFile=present` | AC-2.2, D-9 |
| T-18 | Swift | `CitationComposerDepictionTests.swift` `duplicateCopiesTheRegionNotTheImage` *(new)* | duplicate → same subject, property, polarity, region; empty `valueDepictionID` and `imageRelPath`; no write; focus `.subject` | AC-2.3, D-10 |
| T-19 | Swift | `CitationComposerDepictionTests.swift` `savedLineCarriesItsImage` *(new)* | listed depiction with `imageRelPath "objects/ab/cd.png"` → line `imageRelPath` that path, `valueText` empty | AC-1.7, D-13 |
| T-20 | Swift | `CitationComposerDepictionTests.swift` `drawingDisablesCitationControlsAndLeaveReturns` *(new)* | while drawing → `isDrawingDepiction`; navigation requested → leave confirm; Keep editing → `presented == .drawingDepiction(over: .authoring)`; while drawing, `beginEdit`, `duplicate(line:)`, `askDeleteLine`, and `beginAddObservation` change nothing | AC-1.9, AC-1.10, D-7 |
| T-22 | Swift | `CitationComposerDepictionTests.swift` `captionFollowsTheDraft` *(new)* | no region → `depictionCaptionEmpty`; region → `depictionCaptionBox`; edit after a redraw → `depictionCaptionRedrawn`; existing PDF Citation with no page → `depictionCaptionNeedsPage`; hidden region → `depictionCaptionHidden` | D-19 |
| T-23 | Swift | `CitationComposerDepictionTests.swift` `depictionCopyNamesThePropertyAndPerson` *(new)* | drawn draft on Clara → `unsavedSummary` is `leaveDepictionDraft("Likeness", "Clara Ford", ref)`; duplicate of Ada's line → note is `depictionDuplicatedFrom("Ada Lovelace", "Likeness")` | D-21 |
| T-21 | Swift | `CitationComposerDepictionTests.swift` `hiddenRegionAsksForARedraw` *(new)* | edit a depiction loaded with empty `regionJSON` → `regionHidden`, `editCanConfirm == false`; after a commit → `regionHidden == false`, `editCanConfirm == true` | AC-2.4, D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CitationComposerModelTests` (`saveKeepsTheSubject`, `incompatibleSubjectClearsPropertyAndDisablesSave`, `propertyDraftWalksSubjectThenPropertyThenValue`), `CitationAutoTranscribeTests`, and `OCRImageTests` after the moves.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); any check not run on this machine is reported as not run (D-18)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory and the agreed board; no new kit primitive
- [ ] Version bump: none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and Slice 8 line; completed.md: one short paragraph of what shipped
- [ ] S10-D15 moved to design/archive/ and design/README.md moves it to Done
- [ ] docs/ideas/depictions-and-likenesses.md: the saved-faces bullet added; JPEG corrected to PNG
- [ ] .agents/skills/add-swift-test/SKILL.md names ArtifactFileFixtures
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, in a project created after S10-32 (so it has Likeness), with a Source that has a group photo Artifact (`wedding.jpg`), a two-page PDF Artifact, and an audio Artifact.

1. Open the composer on the photo, New citation, empty transcription. Add an observation on Ada: Likeness is in the property list. Draw on the file: the dialog steps aside, the hint is above the canvas, and Clear, Set page, the Citation chooser, and Auto Transcribe are disabled. Draw a rectangle round her face: the dialog returns with her face. The Citation's locator line is unchanged.
2. Save: the dialog is ready for the next face with Likeness kept and the person pick focused. Add Ben the same way.
3. Switch to the PDF, New citation, go to page 2, add Likeness on Clara, draw: the locator shows page 2. On an existing Citation with page 1, Draw from page 2: the viewer goes to page 1. Page to 2 and draw: the hint says to draw on page 1.
4. Draw a face and click Back before saving: the leave confirm asks; Keep editing returns to drawing.
5. Saved lines for Ada and Ben show their faces. VoiceOver reads "Likeness, Ada" on the line and "Likeness of Ada" on the thumbnail.
6. Edit Ben's likeness, redraw it, save: the line shows the new face. Duplicate Ada's: the dialog has her region; change the person to Clara and save: Ada's line is unchanged.
7. On the audio Artifact, Likeness is not offered.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `pvFormDialog` writes its binding false when the composer hides it for drawing, which would discard the draft | D-7: the bindings ignore false while drawing; T-9 covers the model, manual check step 1 the view |
| `PDFPage.thumbnail(of:for:)` does not draw the rotated page the way the overlay normalizes | T-6 checks against a mark at a known corner |
| A large photo's `CGImage` on the main actor at save | The crop comes from the already-loaded `displayImage`, as Auto Transcribe's does; the PDF page bitmap is capped at 4096 |
| A temp file leaks on a crash between write and delete | System temporary directory, named `provenencia-depiction-*`; accepted |
| The subject rule (D-11) changes behavior for a non-depiction Property shared by two subject types | The property and value are kept only when the new subject's type also offers that Property, the same list the dialog shows |

**Stop and report if:**

- The agreed board keeps the observation dialog on screen while a region is drawn, draws inside the dialog, or starts drawing outside the dialog: D-7 no longer holds.
- The agreed board shows saved faces on the photo: that is Out.
- `ArtifactViewer` cannot take a region from `viewerRegion` without a change to `ArtifactRegionOverlayView`.
- Go refuses a region JSON from `CitationLocatorDraft.encodeJSON()` (`depictionvalues.invalid` in the manual check).
- `ObservationRow.loaded(from:)` or `asCatalogObservation` no longer match §4 because S10-52 or another PR touched them.
- `saveKeepsTheSubject` or `incompatibleSubjectClearsPropertyAndDisablesSave` cannot pass with D-11's subject rule.
- `applySubject` or `propertyOptions(for:)` has gained a caller since the planning base.

## 12. Open questions

None.

## 13. PR

**Title:** S10-34 — Draw a likeness in the citation composer
**Body:** objective in one sentence; what shipped against US-1 to US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-34-composer-depiction-row.md`.
