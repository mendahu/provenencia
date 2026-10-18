# S10-65 — Render the likeness preview off the main thread

**Status:** Draft
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the preview keeps its 96pt `PVThumbnail`, which already has a loading state, and an unchanged edit shows the saved image through the existing `CachedThumbnail`)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-65-depiction-preview-off-main`
**Version bump:** none — Spike 10 and Spike 11 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Rules that bind:** `.cursor/rules/swift-tests.mdc`, `.cursor/rules/design-system-layers.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [Spike 10 decisions](../../archive/spike-10/README.md#decisions) ("A depiction is an owned value row … The app renders the region"); the S10-34 rule "render a face once for preview and save" (commit `0891a1a`); the 2026-10-18 likeness area audit, finding PRF-1 (action list Medium item 7)
**Planned:** 2026-10-18 against `main` @ `e97cf54`
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

In the citation composer, the Likeness dialog's preview is cut from the open photo or PDF page inside a SwiftUI `body` pass, on the main thread. For a PDF it rasterizes the whole page at up to 4 px/pt to keep only the face box, so a scanned archive page decodes its page image on the main thread each time the region changes, the viewer finishes loading, or a saved likeness is opened. After this PR the cut runs once per region, off the main thread, and draws only the face box; `body` reads a stored result; and opening a saved likeness that is not redrawn shows its saved image instead of cutting again.

**Why now.** The 2026-10-18 likeness audit rated this Medium (PRF-1). S10-63 adds two more ways to set a region, each of which would pay the same main-thread cut.

**What stays.** The image a save sends is the same pixels as the preview (the S10-34 rule): one cut function serves both. The crop geometry (y-down normalized box, rotation as displayed, the 2,048 px long-edge cap), the caption, the PDF page rule, the temporary PNG, and every Go path stay as they are.

## 2. User stories

**US-1.** As a researcher drawing a face on a large scanned PDF page, when I commit the region, the dialog stays responsive and the preview appears when the cut is ready.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Likeness draft on the 8×8 test photo; `commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))` | Right after the commit `depictionPreviewState` is `.cutting`; after `await renderDepictionPreview()` it is `.cut` with a 4×4 image |
| AC-1.2 | The marked 100×200 pt test PDF, unrotated; cut region (0, 0)–(0.05, 0.1) on page 1 | The cut is 20×80 px (5×20 pt at 4 px/pt); its first pixel is black and its last pixel is white — only the box was drawn |
| AC-1.3 | The marked PDF rotated 90°; regions (0.5, 0)–(1, 0.5) and (0, 0)–(0.5, 0.5) | As today: the first contains black, the second does not, and each is twice as wide as it is tall |
| AC-1.4 | The 4096×1024 test PNG; region (0, 0)–(1, 1) | As today: 2048×512 |
| AC-1.5 | Region A committed, then region B committed before A's cut finishes | The preview shows B's cut; A's result is never stored |

**US-2.** As a researcher, what I see in the preview is exactly what Save sends.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Photo Citation; region committed; preview cut | The PNG `withDepictionInput` writes decodes to the same width, height and RGBA bytes as the preview's image |
| AC-2.2 | Marked PDF Citation, page 1; region committed; preview cut | Same as AC-2.1 |
| AC-2.3 | Region committed and Save pressed before any preview cut ran | Save cuts with the same function; the call ends ` depiction=1 depictionFile=present` and the temporary file is deleted |
| AC-2.4 | A region on an Artifact whose file is missing (the viewer shows `.missingFile`) | `depictionPreviewState` is `.failed`; Save shows `depictionRenderFailed` and keeps the draft, as today |

**US-3.** As a researcher opening a saved likeness to change its person, I see the saved face at once, with no cut.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Edit a saved likeness (image `objects/…` path set, region unchanged) | `depictionPreviewState` is `.saved(relPath:)` with that path; `depictionPreviewKey` is nil, so no cut runs |
| AC-3.2 | Same edit; redraw a different region | `depictionPreviewState` becomes `.cutting`, then `.cut` |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ArtifactRegionCrop` splits into a main-actor snapshot and a nonisolated cut; the PDF cut draws only the box | STEP-1, AC-1.2, AC-1.3, AC-1.4 |
| `ArtifactViewerModel.fileURL` for the PDF cut | STEP-1, AC-1.2 |
| The composer stores one cut per key, written only for the dialog's current key; `body` reads the stored result | STEP-2, AC-1.1, AC-1.5 |
| Save reuses the stored cut or runs the same cut | STEP-2, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| The value area starts the cut with `.task(id:)`, shows the loading state, and shows the saved image on an unchanged edit | STEP-3, AC-1.1, AC-3.1, AC-3.2 |
| One shared RGBA reader for test pixels | STEP-1, STEP-4 |

| Out | Home |
| --- | --- |
| Blocking Draw on files the viewer cannot cut, and a value-area message plus a disabled Save when a cut fails | audit action list, Low (UX-6) |
| Keeping a card's loaded likeness on the evidence graph card; decoding thumbnails at the drawn size | audit action list, Medium items 6 and 13 (PRF-2) |
| Decoding the crop before taking the catalog session in Go | audit action list, Low (PRF-3) |
| Storing likeness crops below 2,048 px | audit action list, Low (PRF-4) |
| Making `DepictionValueDraft.hasRegion` keep its validation result | audit SEC lens "Outside my lens"; not planned here |
| A Windows renderer | [`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) ("Windows") |

## 4. Current state

Verified at `e97cf54`.

| Area | Where | Today |
| --- | --- | --- |
| Preview read | `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | `PVThumbnail(PVThumbnail.Content(image: model.depictionPreview.map { Image(nsImage: $0) }, label: L10n.CitationComposer.depictionPreviewLabel), size: 96)` — `body` reads `depictionPreview`, which cuts on a memo miss |
| Preview cut | `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift` `depictionPreview`, private `renderedDepiction(for:)` | Builds a key from `selectedArtifact?.id`, `artifactViewer.kind.rawValue`, `String(artifactViewer.isLoading)`, `draft.locator.encodeJSON()`; on a miss calls `ArtifactRegionCrop.render(artifactViewer, page:, region:)` synchronously and writes `depictionPreviewMemo` |
| Memo | `macos/App/Features/CitationComposer/CitationComposerModel.swift` `depictionPreviewMemo` | `@ObservationIgnored var depictionPreviewMemo: DepictionPreviewMemo?` — ignored because `body` writes it; `DepictionPreviewMemo` (in `+Depiction.swift`) holds `key`, `cgImage: CGImage`, `image: NSImage` |
| Save | `CitationComposerModel+Depiction.swift` `withDepictionInput(for:perform:)` | `renderedDepiction(for:)?.cgImage` or throws `DepictionRenderFailed`; writes the PNG in `Task.detached { try ArtifactRegionCrop.writeTemporaryPNG(rendered) }` (a `CGImage` already crosses into a detached task here); removes the file in `defer`. Called from `saveAuthoring` and `saveEdit` in `CitationComposerModel+Authoring.swift`, which map `DepictionRenderFailed` to `L10n.CitationComposer.depictionRenderFailed` |
| Redraw flag | `CitationComposerModel+Depiction.swift` `isDepictionRedrawn` | True when the edit's region differs from `session.propertyBaseline` |
| Region commit | `CitationComposerModel+Depiction.swift` `commitViewerRegion(_:)` | Page rule, `draft.locator.setRegion`, `writeDepictionDraft`, disarm, restore the dialog. No cut |
| Cut | `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` `render(_:page:region:)` | `@MainActor`. Image: `cgImage(from: viewer.displayImage)`, `boundingBox`, `scaledToFit(maxLongEdge: 2048)`. PDF: `pageBitmap` renders the whole page with `page.thumbnail(of:for: .mediaBox)` at `min(maxPageScale 4, maxLongEdge / boxLong, maxPageLongEdge 4096 / visualLong)`, then crops the box |
| Other crop users | `macos/App/Features/CitationComposer/CitationTranscriptionModel.swift` | Uses `ArtifactRegionCrop.cgImage(from:)` and `boundingBox(of:region:)` on the main actor for OCR; untouched |
| Viewer | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` | `@MainActor @Observable`. `load(_:)` resolves `url` with `ProjectFiles.objectURL`, decodes in `Task.detached`, keeps `content` (`ArtifactViewerContent.image(NSImage)` / `.pdf(PDFDocument)`), `displayImage`, `kind`, `isLoading`, `page`. The URL is not kept. `unload()` and `unloadPreservingGeneration()` reset state |
| Saved image elsewhere | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` | The saved line shows `CachedThumbnail(projectDir: model.session.projectKey.projectDir, relPath: relPath, mediaType: "image/png", size: 32)` |
| Thumbnail states | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.Content` | `image`, `glyph`, `loading`, `label`; `loading: true` draws the kit's loading state |
| Tests | `macos/ProvenenciaTests/ArtifactRegionCropTests.swift` `imageCropIsTheBoundingBox`, `cropIsCappedAt2048`, `pdfRegionFollowsTheDisplayedPage`, private `pixels(_:)`; `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` `drawingLeavesTheCitationRegion` (reads `depictionPreview` at lines 306 and 316), `savingSendsAPNGAndDeletesIt`, `editWithoutRedrawKeepsTheImage`, `redrawSendsANewImage`, `refusedSaveKeepsTheRegion`; fixtures `ArtifactFileFixtures.writePNG`, `writeMarkedPDF` (one 100×200 pt page, black 10×10 pt square at the unrotated top-left) | Cover the cut and the save synchronously |
| Concurrency settings | `macos/Provenencia.xcodeproj/project.pbxproj` | `SWIFT_STRICT_CONCURRENCY = complete`, `SWIFT_VERSION = 5.0` |

**Precedent.** `ThumbnailCache.image(projectDir:relPath:)` (`macos/App/Platform/ThumbnailCache.swift`) decodes in `Task.detached(priority: .userInitiated)` and hands a `CGImage`-backed result back to the main actor; `CachedThumbnail` starts its load with `.task(id:)`, so SwiftUI cancels a stale load when the key changes. Copy both: the detached cut and the `.task(id:)` trigger.

**Overlap with S10-63 (Planned, not landed).** S10-63 adds `useDepictionRegion(_:)` → `commitViewerRegion(_:)` and a button row in `CitationDepictionValueArea`. Its buttons need nothing from this PR: a region they set changes `depictionPreviewKey`, and the value area's `.task(id:)` cuts it. Its T-2 and T-9 save right after a button press with no cut run, which AC-2.3 covers. Both PRs edit `CitationComposerModel+Depiction.swift` and `CitationDepictionValueArea`, in different members: whichever lands second merges `main` and keeps both changes.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ArtifactRegionCrop` gains `enum Source: Sendable` *(new)* with `case image(CGImage)` and `case pdf(url: URL, pageIndex: Int)`; `@MainActor static func snapshot(_ viewer: ArtifactViewerModel, page: Int?) -> Source?` *(new)* reads the viewer; `nonisolated static func cut(_ source: Source, region: ArtifactRegionDraft) -> CGImage?` *(new)* does the work; `@MainActor static func render(_ viewer:page:region:) async -> CGImage?` keeps its name, becomes `async`, and is `snapshot` then `await Task.detached(priority: .userInitiated) { cut(source, region: region) }.value`. | One cut path for preview, save, and tests (the S10-34 rule); only `Sendable` values cross the actor boundary | Passing `ArtifactViewerModel` or `PDFDocument` into the detached task — neither is `Sendable`, and PDFKit is not documented as thread-safe |
| D-2 | The image snapshot is `cgImage(from: viewer.displayImage)` on the main actor; cropping and `scaledToFit` run in `cut`. | The viewer has already drawn the image, so its `CGImage` is decoded; `cropping(to:)` and the high-quality redraw are the work PRF-1 names. `CGImage` already crosses into a detached task in `withDepictionInput` | Re-decoding the file in the task with ImageIO — a second full decode, and ImageIO does not apply orientation the way `NSImage` does, so the cut could differ from what the researcher drew on |
| D-3 | The PDF snapshot is the viewer's file URL and the 0-based page index. `ArtifactViewerModel` gains `private(set) var fileURL: URL?` *(new)*, set from the resolved `url` when `load` succeeds and set to nil in `unload()`, `unloadPreservingGeneration()`, and the failure branch. `cut` opens its own `CGPDFDocument(url as CFURL)` and takes `page(at: pageIndex + 1)`. | The task owns every object it touches; `URL` is `Sendable` | `PDFPage.pageRef` handed to the task — shared with PDFKit across threads; `PDFDocument.dataRepresentation()` — copies a whole archive PDF into memory per cut |
| D-4 | The PDF cut draws only the box. Visual size = media box size with width and height swapped when `rotationAngle % 180 == 90`. Box in visual points = the region's normalized bounding rect × visual size. Scale = `min(maxPageScale, maxLongEdge / max(box width, box height))`. The context is `round(box width × scale) × round(box height × scale)`, sRGB, 8-bit, `premultipliedLast` (as `scaledToFit`), filled white, then scaled, translated so the box's top-left sits at the origin (y-up: `-(visualHeight − box.maxY)`), concatenated with `page.getDrawingTransform(.mediaBox, rect: CGRect(origin: .zero, size: visual), rotate: 0, preserveAspectRatio: true)`, and `drawPDFPage`. `maxPageLongEdge` and `pageBitmap` are deleted. | The box is all that is kept; `getDrawingTransform` applies the page's own rotation the way `PDFPage.thumbnail` did, so AC-1.3 holds; white fill matches the thumbnail's page background | Keeping `page.thumbnail(of:for:)` and moving it off main — still renders the whole page; keeping the `maxPageLongEdge` cap — it only bounded the whole-page bitmap |
| D-5 | `DepictionPreviewMemo` keeps `key` and holds `cgImage: CGImage?` and `image: NSImage?`; both nil means the cut failed for that key. `depictionPreviewMemo` loses `@ObservationIgnored`. | The memo is now written from a task, so the view must observe it; a stored failure stops a retry loop for the same key (the key changes when the viewer finishes loading, so a cut tried while loading is tried again) | A separate `isCutting` flag — the state is derivable from "no memo for the current key" |
| D-6 | `var depictionPreviewKey: String?` *(new)* is nil when there is no depiction draft, the draft has no valid region, or the dialog is an edit that is not redrawn, has a valid region, and has a saved image path; otherwise the existing key string (artifact id, viewer kind, viewer loading, locator JSON) built by `private func depictionPreviewKey(for draft: DepictionValueDraft) -> String?` *(new)*. | One key rule for the trigger, the read, and the save; an unchanged saved likeness needs no cut (US-3) | Cutting on every edit open — the cost PRF-1 names |
| D-7 | `enum DepictionPreviewState: Equatable { case none, cutting, cut(NSImage), failed, saved(relPath: String) }` *(new)*; `var depictionPreviewState: DepictionPreviewState` *(new)* replaces `depictionPreview`. `.saved` when the edit is not redrawn, the draft `hasRegion`, and `imageRelPath` is set (a hidden region keeps today's empty preview); `.none` when `depictionPreviewKey` is nil otherwise; `.cut` / `.failed` when the memo's key equals `depictionPreviewKey`; `.cutting` otherwise. | `body` reads stored state only; the cases are testable without a view | Returning `PVThumbnail.Content` from the model — puts a kit type in the model |
| D-8 | `func depictionImage(for draft: DepictionValueDraft) async -> CGImage?` *(new)* is the one cut entry: it returns the memo's image when the memo's key equals the draft's key (nil for a stored failure); otherwise it awaits `ArtifactRegionCrop.render`, and writes the memo only when the draft's key still equals `depictionPreviewKey` after the await. `func renderDepictionPreview() async` *(new)* calls it for `currentDepictionDraft` when `depictionPreviewKey` is non-nil. `withDepictionInput` calls `depictionImage(for: values.depictionDraft)` and throws `DepictionRenderFailed` on nil. `renderedDepiction(for:)` and `depictionPreview` are deleted. | Preview and save share one function and one stored result, so the saved pixels are the previewed pixels; a stale cut never overwrites the current one (AC-1.5) | A stored `Task` handle the save awaits — cancellation bookkeeping for the few hundred milliseconds between commit and Save; a save pressed then runs the same deterministic cut (AC-2.3) |
| D-9 | `CitationDepictionValueArea` adds `.task(id: model.depictionPreviewKey) { await model.renderDepictionPreview() }` on its outer `VStack`. The preview maps `depictionPreviewState`: `.none` and `.failed` → `PVThumbnail(PVThumbnail.Content(label: depictionPreviewLabel), size: 96)`; `.cutting` → the same with `loading: true`; `.cut(image)` → the same with `image: Image(nsImage: image)`; `.saved(relPath)` → `CachedThumbnail(projectDir: model.session.projectKey.projectDir, relPath: relPath, mediaType: "image/png", size: 96, imageLabel: L10n.string(L10n.CitationComposer.depictionPreviewLabel))`. | SwiftUI cancels a stale task when the key changes and starts one for every trigger (region commit, Citation switch, viewer load, dialog open) with no model hook per trigger; both thumbnails are existing kit and platform views | A cut started from `commitViewerRegion` — misses the viewer-load and dialog-open triggers |
| D-10 | The detached cut is not interrupted when its key changes; its result is dropped (D-8). | Core Graphics drawing cannot be cancelled mid-draw | `withTaskCancellationHandler` around the cut — nothing to cancel |
| D-11 | Tests read pixels through `static func rgbaBytes(_ image: CGImage) -> [UInt8]` *(new)* on `ArtifactFileFixtures`; `ArtifactRegionCropTests.pixels(_:)` is rewritten on top of it. | AC-2.1 and AC-2.2 compare pixels in a second suite; one RGBA reader instead of two | A second private reader in `CitationComposerDepictionTests` |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionCrop.Source` *(new)* | type | `ArtifactViewerContent` (`@unchecked Sendable`, holds `NSImage` / `PDFDocument`) | Its cases are not safe off the main actor; `Source` holds only `CGImage` and `URL` |
| `ArtifactRegionCrop.snapshot(_:page:)` *(new)* | func | `ArtifactRegionCrop.render` (the viewer half) | It is the viewer half of `render`, split out (D-1) |
| `ArtifactRegionCrop.cut(_:region:)` *(new)* | func | `ArtifactRegionCrop.render`, `pageBitmap` | It is the work half of `render`; `pageBitmap` is replaced by the box draw (D-4) |
| `ArtifactViewerModel.fileURL` *(new)* | property | the local `url` in `ArtifactViewerModel.load` | Kept instead of discarded (D-3) |
| `DepictionPreviewState` *(new)* | type | `PVThumbnail.Content` | A model type, not a kit type (D-7) |
| `depictionPreviewKey`, `depictionPreviewKey(for:)` *(new)* | property, func | the key in `renderedDepiction(for:)` | Same string, moved so the trigger, the read, and the save share it (D-6) |
| `depictionImage(for:)`, `renderDepictionPreview()` *(new)* | func | `renderedDepiction(for:)` | Async replacement (D-8) |
| `ArtifactFileFixtures.rgbaBytes(_:)` *(new)* | test helper | `ArtifactRegionCropTests.pixels(_:)` (private) | Moved to the shared fixture so two suites read pixels one way (D-11) |

**Builds on.** `ThumbnailCache`'s detached decode and `CachedThumbnail`'s `.task(id:)` load; `ArtifactFileFixtures` (add-swift-test: artifact files on disk come from it); `PVThumbnail.Content(loading:)`.
**Deletes.** `ArtifactRegionCrop.pageBitmap`, `ArtifactRegionCrop.maxPageLongEdge`, `CitationComposerModel.depictionPreview`, `renderedDepiction(for:)`, `@ObservationIgnored` on `depictionPreviewMemo`.
**Next adds.** Audit UX-6 ("show a failed cut before Save") reads `depictionPreviewState == .failed` with no new state. The idea file's "Transparent polygon" crop changes only `ArtifactRegionCrop.cut`. A Windows client keeps the same snapshot / cut split.

### Schema

None — no catalog change.

### Go core

None — Go receives the same temporary PNG.

### FFI

None — no message changes.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Preview, cutting | `PVThumbnail` (`Content(loading: true, label:)`) | `CitationComposerAuthoringBox.swift` `CitationDepictionValueArea` | 96pt, the existing loading state |
| Preview, cut | `PVThumbnail` (`Content(image:label:)`) | same | As today |
| Preview, none or failed | `PVThumbnail` (`Content(label:)`) | same | As today's empty preview |
| Preview, saved | `CachedThumbnail` (`size: 96`, `imageLabel:`) | same | The saved line's view at the preview's size |
| Cut trigger | `.task(id: model.depictionPreviewKey)` | same | On the outer `VStack` |

States: no region → empty preview, empty caption; region just committed → loading preview; cut ready → the cut; cut failed → empty preview (the caption is unchanged; UX-6 is Out); unchanged edit → the saved image, loaded through `ThumbnailCache`. Accessibility: the preview keeps the label `depictionPreviewLabel` in every state.

### Strings

None — no new copy.

## 7. Steps

**STEP-1 — Cut off the main actor, and only the PDF box**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift`, `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift`, `macos/ProvenenciaTests/ArtifactRegionCropTests.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`. Skill: `add-swift-test`.
Do: add `Source`, `snapshot`, `cut`, and the async `render` (D-1, D-2, D-4); delete `pageBitmap` and `maxPageLongEdge`; add `fileURL` (D-3). Add `rgbaBytes` to `ArtifactFileFixtures` and rewrite `pixels(_:)` on it (D-11). Make `imageCropIsTheBoundingBox`, `cropIsCappedAt2048`, and `pdfRegionFollowsTheDisplayedPage` `await` the render with their assertions unchanged; add T-2 and T-3. The app's only caller of `render`, `renderedDepiction(for:)`, is synchronous and is replaced in STEP-2, so the app target builds again only after STEP-2: work STEP-1 and STEP-2 before the first build.
Check: after STEP-2, `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ArtifactRegionCropTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-2 — One stored cut for preview and save**
Files: `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift`.
Do: change `DepictionPreviewMemo` (D-5) and drop `@ObservationIgnored`; add `depictionPreviewKey`, `depictionPreviewKey(for:)` (D-6), `DepictionPreviewState` and `depictionPreviewState` (D-7), `depictionImage(for:)` and `renderDepictionPreview()` (D-8); rewrite `withDepictionInput` onto `depictionImage(for:)`; delete `depictionPreview` and `renderedDepiction(for:)`. Doc comments in the file's style.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO` succeeds with no new concurrency warning in the two changed files; the STEP-1 check passes.

**STEP-3 — The value area reads the stored state**
Files: `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`. Skill: `add-ui-component`.
Do: in `CitationDepictionValueArea`, add the `.task(id:)` trigger and map `depictionPreviewState` to the preview as §6 macOS lists (D-9).
Check: the STEP-2 build succeeds.

**STEP-4 — Model tests**
Files: `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`. Skill: `add-swift-test`.
Do: in `drawingLeavesTheCitationRegion`, replace `model.depictionPreview == nil` with `model.depictionPreviewState == .none` and the 4×4 size check with `await model.renderDepictionPreview()` followed by a `.cut` image whose size is 4×4. Add T-4 to T-9.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CitationComposerDepictionTests -only-testing:ProvenenciaTests/ArtifactRegionCropTests -only-testing:ProvenenciaTests/CitationAutoTranscribeTests -only-testing:ProvenenciaTests/ArtifactViewerModelTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-5 — Docs**
Files: none in this step. No domain doc states how the preview is cut; the Spike 10 decision ("The app renders the region") still holds.
Do: nothing in content. The plan Status and the S10-65 map row move in implement-pr step 8's commit, after the draft PR has a number.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-65-depiction-preview-off-main.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS` on a macOS host.

Commits, in order: STEP-1 to STEP-4 together, staged once STEP-4's check passes ("Cut the likeness preview once off the main thread, draw only the PDF box, and show a saved likeness without cutting it."). The cut's call site moves with its caller, so no earlier commit builds. Then the step-8 commit.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `ArtifactRegionCropTests` `imageCropIsTheBoundingBox`, `cropIsCappedAt2048`, `pdfRegionFollowsTheDisplayedPage` (now `await` the render) | Assertions unchanged: 4×4 and red top; 2048×512; rotated black square on the top-right box, 2:1 sides | AC-1.3, AC-1.4, D-1, D-2, D-4 |
| T-2 | Swift | `ArtifactRegionCropTests` `pdfCutDrawsOnlyTheBox` *(new)* | `writeMarkedPDF(rotation: 0)`, load, render page 1 region `rectangle(0, 0, 0.05, 0.1)` → 20×80; `rgbaBytes` first pixel black (R, G, B < 60), last pixel white (R, G, B > 200) | AC-1.2, D-4 |
| T-3 | Swift | `ArtifactRegionCropTests` `pdfCutNeedsTheLoadedFile` *(new)* | A fresh `ArtifactViewerModel` with nothing loaded → `snapshot(viewer, page: 1)` is nil and `fileURL` is nil; after loading the marked PDF, `fileURL` is non-nil and `snapshot` is `.pdf` with `pageIndex == 0`; after `unload()`, `fileURL` is nil | D-3 |
| T-4 | Swift | `CitationComposerDepictionTests` `previewIsCutAfterTheRegionCommits` *(new)* | `seeded()`, Likeness draft, `startDepictionDrawing()`, `commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))` → `depictionPreviewState == .cutting`; `await renderDepictionPreview()` → `.cut` with size 4×4 | AC-1.1, D-7, D-8 |
| T-5 | Swift | `CitationComposerDepictionTests` `savedPixelsAreThePreviewPixels` *(new)* | Photo: commit, `await renderDepictionPreview()`, then `try await model.withDepictionInput(for: draftValues(model)!)` with a closure that decodes `input.imagePath` through `CGImageSourceCreateWithURL` + `CGImageSourceCreateImageAtIndex` → width, height and `ArtifactFileFixtures.rgbaBytes` equal those of the `.cut` image's `CGImage` (`depictionPreviewMemo?.cgImage`). PDF: `seeded(.pdf)`, overwrite `objects/news.pdf` with `writeMarkedPDF(rotation: 0)` before `prepare()`, commit `rect(0, 0, 0.5, 0.5)` on page 1, same comparison | AC-2.1, AC-2.2, D-8 |
| T-6 | Swift | `CitationComposerDepictionTests` `savingSendsAPNGAndDeletesIt` (unchanged) | Saves right after the commit with no cut run → ` depiction=1 depictionFile=present`, file removed | AC-2.3, D-8 |
| T-7 | Swift | `CitationComposerDepictionTests` `editWithoutRedrawShowsTheSavedImage` *(new)* | `editingSavedLikeness()` → `depictionPreviewKey == nil`, `depictionPreviewState == .saved(relPath: savedRelPath)` (the suite's saved path constant); a second model seeded as `hiddenRegionAsksForARedraw` seeds it (`likenessObservation(region: "")`, `observationID: "obs-1"`) → `.none`; `startDepictionDrawing()`, `commitViewerRegion(rect(0.5, 0.5, 0.9, 0.9))` → `.cutting`; `await renderDepictionPreview()` → `.cut` | AC-3.1, AC-3.2, D-6 |
| T-8 | Swift | `CitationComposerDepictionTests` `staleCutIsNotStored` *(new)* | Commit A = `rect(0.25, 0.25, 0.75, 0.75)`, keep `let a = draftValues(model)!.depictionDraft`; draw and commit B = `rect(0, 0, 1, 1)`; `_ = await model.depictionImage(for: a)` → `depictionPreviewState == .cutting` and `depictionPreviewMemo == nil`; `await renderDepictionPreview()` → `.cut` with size 8×8 | AC-1.5, D-8 |
| T-9 | Swift | `CitationComposerDepictionTests` `missingFileCutFails` *(new)* | `seeded()`, delete `objects/photo.png` under `projectDir` before `prepare()`; Likeness draft; `commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))`; `await renderDepictionPreview()` → `.failed`; `await saveAuthoring()` → the draft's error is `L10n.string(L10n.CitationComposer.depictionRenderFailed)` and its region is kept | AC-2.4, D-5 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break: `CitationComposerDepictionTests` (`redrawSendsANewImage`, `refusedSaveKeepsTheRegion`, `keepEditingAfterACitationSwitchKeepsTheRegion`), `ArtifactRegionCropTests`, `CitationAutoTranscribeTests` (shares `ArtifactRegionCrop.cgImage(from:)` and `boundingBox`), `ArtifactViewerModelTests` (`fileURL` in `load`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none added)
- [ ] UI matches §6 macOS; no new kit primitive
- [ ] Version bump: none
- [ ] After the draft PR has a number, implement-pr step 8's commit "Record S10-65 as landed in pull request N.": this plan's Status → Landed in #N (date) with a Landed note listing every difference from the plan, and the S10-65 row in spike-10-followups/deployment-plan.md → Landed #N (this folder has no completed.md)
- [ ] Brief: none (no new chrome)
- [ ] Domain doc: none changes (STEP-5)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with a project that has a photo Source and a scanned multi-page PDF Source (a newspaper or register page scan of at least 2,000 px on its long edge).

1. Open the PDF Source's Citation on page 1 in the composer. Add a Likeness. Press **Draw on the file** and drag a box around one face. The dialog returns at once; the preview shows the loading state, then the face.
2. Draw again, a larger box. The preview shows the new face, never the old one.
3. Save. Open the saved line's thumbnail and the preview from step 2 side by side: the same face, the same crop.
4. Edit the saved likeness and change only the person. The preview shows the saved face immediately, with no loading state.
5. In that edit, press **Draw again** and draw a new box. The preview shows the loading state, then the new face; the caption reads the redrawn caption. Save.
6. Repeat steps 1 to 3 on the photo Source.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The box draw places the face differently from the old whole-page crop on a rotated page | T-1 keeps the rotated-page assertions; T-2 checks the box contents |
| Opening a second `CGPDFDocument` per cut is slow on a very large PDF | It runs off the main actor once per key; manual check 1 uses a scanned page |
| `cgImage(from:)` on the main actor decodes a photo the viewer has not drawn yet | The viewer draws the image before a region can be drawn on it; a cut started during load has a different key and is tried again when loading ends |

**Stop and report if:**

- No macOS host with Xcode is available: every Swift check in this PR is `xcodebuild`, and `scripts/check-pr.sh` skips Xcode tests off macOS.
- The STEP-2 build reports a `Sendable` or actor-isolation warning or error for `ArtifactRegionCrop.Source` or the detached cut.
- T-1's rotated-page assertions fail with the box draw of D-4.
- T-5's pixel comparison differs for either the photo or the PDF.
- `ArtifactViewerModel.load` no longer resolves its file through `ProjectFiles.objectURL` (S10-62 changes that function's checks, not its signature).
- `CitationComposerModel+Depiction.swift` already has `useDepictionRegion` (S10-63 landed first) and its call into `commitViewerRegion` changed shape: merge `main`, keep both, and report the merge in the Landed note.

## 12. Open questions

None.

## 13. PR

**Title:** S10-65 — Render the likeness preview off the main thread
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3; AC evidence (test names); deviations from the plan; manual check results; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-65-depiction-preview-off-main.md`.
