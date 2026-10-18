# S10-65 — Render the likeness preview off the main thread

**Status:** Landed in [#472](https://github.com/mendahu/provenencia/pull/472) (2026-10-18)
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
**Review:** [`S10-65-depiction-preview-off-main.review.md`](S10-65-depiction-preview-off-main.review.md) — Fits / Ready with changes, 2 rounds

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

In the citation composer, the Likeness dialog's preview is cut from the open photo or PDF page inside a SwiftUI `body` pass, on the main thread. For a PDF it rasterizes the whole page at up to 4 px/pt to keep only the face box, so a scanned archive page decodes its page image on the main thread each time the region changes, the viewer finishes loading, or a saved likeness is opened. After this PR the cut runs once per region, off the main thread, and draws only the face box; `body` reads a stored result; and opening a saved likeness that is not redrawn shows its saved image instead of cutting again, which is also what its Save sends (no new image).

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
| AC-3.1 | Edit a saved likeness (image `objects/ab/cd.png` set, region unchanged) | `depictionPreviewState` is `.saved(relPath: "objects/ab/cd.png")`; `depictionPreviewKey` is nil, so no cut runs |
| AC-3.2 | Same edit; redraw a different region | `depictionPreviewState` becomes `.cutting`, then `.cut` |
| AC-3.3 | Edit a saved likeness whose image path is empty (Go hid an unreadable image), region unchanged | `depictionPreviewState` is `.none` and `depictionPreviewKey` is nil: no cut runs, matching `saveEdit`, which sends no image when the region is not redrawn |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ArtifactRegionCrop` splits into a main-actor snapshot and a nonisolated cut; the PDF cut draws only the box through one shared output context | STEP-1, AC-1.2, AC-1.3, AC-1.4 |
| The composer stores one cut per key, written only for the dialog's current key; `body` reads the stored result | STEP-1, AC-1.1, AC-1.5 |
| Save reuses the stored cut or runs the same cut | STEP-1, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| The value area starts the cut with `.task(id:)`, shows the loading state, and shows the saved image on an unchanged edit, labelled in every state | STEP-1, AC-1.1, AC-3.1, AC-3.2, AC-3.3 |
| One shared RGBA reader for test pixels | STEP-2 |

| Out | Home |
| --- | --- |
| Blocking Draw on files the viewer cannot cut, and a value-area message plus a disabled Save when a cut fails | audit action list, Low (UX-6) |
| Keeping a card's loaded likeness on the evidence graph card; decoding thumbnails at the drawn size | audit action list, Medium items 6 and 13 (PRF-2) |
| Decoding the crop before taking the catalog session in Go | audit action list, Low (PRF-3) |
| Storing likeness crops below 2,048 px | audit action list, Low (PRF-4) |
| Making `DepictionValueDraft.hasRegion` keep its validation result | audit SEC lens "Outside my lens"; not planned here |
| A Windows renderer | [`depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) ("Windows") |
| Moving the OCR crop in `CitationTranscriptionModel` off the main actor | not planned (it runs once per Auto Transcribe, not per `body` pass) |
| Clearing `depictionPreviewMemo` when the dialog closes or the Citation switches | not planned (the key changes with the artifact id, so a stale memo is never read; it holds one image) |
| Rendering a locked or encrypted PDF: `CGPDFDocument` returns no page until unlocked, so the cut is `.failed` where `PDFPage.thumbnail` drew something | not planned (the app has no unlock flow; ingest does not produce locked PDFs) |
| Disabling Save or showing a message when the cut failed | audit action list, Low (UX-6) |
| Moving the hand-added accessibility label into `CachedThumbnail` (the saved line and the `.saved` tile both add it at the call site) | not planned — move the label into `CachedThumbnail` when a third caller adds one |

## 4. Current state

Verified at `e97cf54`. The crop file's header comment still says "a PDFKit render of a PDF page".

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
| Viewer | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` | `@MainActor @Observable`. `load(_:)` resolves `url` with `ProjectFiles.objectURL`, builds `PDFDocument(url: url)` for a PDF (so `pdfDocument?.documentURL` is that URL), decodes images in `Task.detached`, keeps `content` (`ArtifactViewerContent.image(NSImage)` / `.pdf(PDFDocument)`), `displayImage`, `kind`, `isLoading`, `page`. `pdfDocument` derives from `content`, so `unload()`, `unloadPreservingGeneration()` and `installImageRasterForTesting(_:)` clear it. `ArtifactViewerContent.swift` already compares PDFs by `documentURL` |
| Saved image elsewhere | `macos/App/Features/CitationComposer/CitationComposerSubjectGroups.swift` | The saved line shows `CachedThumbnail(projectDir: model.session.projectKey.projectDir, relPath: relPath, mediaType: "image/png", size: 32)` |
| Saved image view | `macos/App/Platform/CachedThumbnail.swift` | Starts empty, sets `PVThumbnail.Content(loading: true)` (no label) in `.task`, then the image or `fallbackContent()` (the PNG file-type mark for a missing or undecodable file); `imageLabel` labels only the image state |
| Edit save | `macos/App/Features/CitationComposer/CitationComposerModel+Authoring.swift` `saveEdit` | Sends a new image only when `isDepictionRedrawn`; otherwise commits with no depiction input, so Go keeps the saved image |
| Thumbnail states | `macos/App/DesignSystem/Components/Thumbnail/PVThumbnail.swift` `PVThumbnail.Content` | `image`, `glyph`, `loading`, `label`; `loading: true` draws the kit's loading state |
| Tests | `macos/ProvenenciaTests/ArtifactRegionCropTests.swift` `imageCropIsTheBoundingBox`, `cropIsCappedAt2048`, `pdfRegionFollowsTheDisplayedPage`, private `pixels(_:)`; `macos/ProvenenciaTests/CitationComposerDepictionTests.swift` `drawingLeavesTheCitationRegion` (reads `depictionPreview` at lines 306 and 316), `savingSendsAPNGAndDeletesIt`, `editWithoutRedrawKeepsTheImage`, `redrawSendsANewImage`, `refusedSaveKeepsTheRegion`; fixtures `ArtifactFileFixtures.writePNG`, `writeMarkedPDF` (one 100×200 pt page, black 10×10 pt square at the unrotated top-left) | Cover the cut and the save synchronously |
| Concurrency settings | `macos/Provenencia.xcodeproj/project.pbxproj` | `SWIFT_STRICT_CONCURRENCY = complete`, `SWIFT_VERSION = 5.0` |

**Precedent.** `ThumbnailCache.image(projectDir:relPath:)` (`macos/App/Platform/ThumbnailCache.swift`) decodes in `Task.detached(priority: .userInitiated)` and hands a `CGImage`-backed result back to the main actor; `CachedThumbnail` starts its load with `.task(id:)`, so SwiftUI cancels a stale load when the key changes. Copy both: the detached cut and the `.task(id:)` trigger.

**Overlap with S10-63 (Planned, not landed).** S10-63 adds `useDepictionRegion(_:)` → `commitViewerRegion(_:)` and a button row in `CitationDepictionValueArea`. Its buttons need nothing from this PR: a region they set changes `depictionPreviewKey`, and the value area's `.task(id:)` cuts it. Its T-2 and T-9 save right after a button press with no cut run, which AC-2.3 covers. Both PRs edit `CitationComposerModel+Depiction.swift` and `CitationDepictionValueArea`, in different members: whichever lands second merges `main` and keeps both changes.

**Overlap with S10-62 and S10-64 (Planned, not landed).** S10-62 deletes the `fileExists` guard in `ArtifactViewerModel.load` and adds `ArtifactFileFixtures.linkToOutsidePNG`; S10-64 adds `ArtifactFileFixtures.writeImage`. This PR does not edit `ArtifactViewerModel.swift`, and adds `rgbaBytes` beside those fixtures: whichever lands second merges `main` and keeps every fixture. After S10-62, `cut` reopens the PDF URL that `objectURL` vetted at load; the gap between that check and the reopen is accepted (S10-62 D-4: the threat is a prepared folder, not a live race).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `ArtifactRegionCrop` gains `enum Snapshot: Sendable` *(new)* with `case image(CGImage)` and `case pdf(url: URL, pageIndex: Int)`; `@MainActor static func snapshot(_ viewer: ArtifactViewerModel, page: Int?) -> Snapshot?` *(new)* reads the viewer; `nonisolated static func cut(_ snapshot: Snapshot, region: ArtifactRegionDraft) -> CGImage?` *(new)* does the work and starts with `guard region.isValid else { return nil }` (the guard moves from `render`); `@MainActor static func render(_ viewer:page:region:) async -> CGImage?` keeps its name, becomes `async`, and is `snapshot` then `await Task.detached(priority: .userInitiated) { cut(snapshot, region: region) }.value`. | One cut path for preview, save, and tests (the S10-34 rule); only `Sendable` values cross the actor boundary; the type is named for the function that makes it | Passing `ArtifactViewerModel` or `PDFDocument` into the detached task — neither is `Sendable`, and PDFKit is not documented as thread-safe; naming the type `Source` — "Source" is the domain's central noun, and `ArtifactViewerSource` already means what the viewer loads (review M-5) |
| D-2 | The image snapshot is `cgImage(from: viewer.displayImage)` on the main actor; `boundingBox` and `scaledToFit` run in `cut`. | The viewer has already drawn the image, so its `CGImage` is decoded; `cropping(to:)` and the high-quality redraw are the work PRF-1 names. `CGImage` already crosses into a detached task in `withDepictionInput` | Re-decoding the file in the task with ImageIO — a second full decode, and ImageIO does not apply orientation the way `NSImage` does, so the cut could differ from what the researcher drew on |
| D-3 | The PDF snapshot is `.pdf(url: viewer.pdfDocument?.documentURL, pageIndex: page - 1)`, nil when the viewer kind is not `.pdf`, the page is nil or below 1, or `documentURL` is nil. `cut` opens its own `CGPDFDocument(url as CFURL)` and takes `page(at: pageIndex + 1)`. `ArtifactViewerModel.swift` is not edited. | The URL and the shown document come from one stored value (`content`), which every unload path already clears; the task owns every object it touches; `URL` is `Sendable` | A new `ArtifactViewerModel.fileURL` — a second copy of a fact `PDFDocument.documentURL` holds, with three reset sites to keep in step (review M-1); `PDFPage.pageRef` handed to the task — shared with PDFKit across threads; `PDFDocument.dataRepresentation()` — copies a whole archive PDF into memory per cut |
| D-4 | `private static func outputContext(width: Int, height: Int) -> CGContext?` *(new)* builds the one output bitmap: sRGB, 8 bits per component, `bytesPerRow: 0`, `premultipliedLast`, `interpolationQuality = .high`. `scaledToFit` draws into it, and so does the PDF cut. The PDF cut draws only the box, in this order (each call concatenates onto the context's transform, so a page point is mapped by the last call first):<br>1. `let media = page.getBoxRect(.mediaBox)`; `let rotated = abs(Int(page.rotationAngle)) % 180 == 90`; `let visual = rotated ? CGSize(width: media.height, height: media.width) : media.size`; return nil unless both sides are above 0.<br>2. `let n = ArtifactRegionGeometry.boundingRect(of: region.points)` (y-down, unit square); `let box = CGRect(x: n.minX * visual.width, y: n.minY * visual.height, width: n.width * visual.width, height: n.height * visual.height)` (y-down points).<br>3. `let scale = min(maxPageScale, maxLongEdge / max(box.width, box.height))`; `let width = max(1, Int((box.width * scale).rounded()))`, `let height = max(1, Int((box.height * scale).rounded()))` (the `scaledToFit` rule, so a tiny valid region yields at least 1×1).<br>4. `guard let ctx = outputContext(width:height:)`; fill `CGRect(x: 0, y: 0, width: width, height: height)` with white.<br>5. `ctx.scaleBy(x: scale, y: scale)`; `ctx.translateBy(x: -box.minX, y: -(visual.height - box.maxY))` (the box's y-up bottom-left goes to the origin); `ctx.concatenate(page.getDrawingTransform(.mediaBox, rect: CGRect(origin: .zero, size: visual), rotate: 0, preserveAspectRatio: true))`; `ctx.drawPDFPage(page)`.<br>6. `ctx.makeImage()`, then `scaledToFit(_:maxLongEdge:)` as for an image (a no-op when it fits).<br>No clip is set. `maxPageLongEdge` and `pageBitmap` are deleted. | The box is all that is kept; `getDrawingTransform` applies the page's own rotation the way `PDFPage.thumbnail` did, so AC-1.3 holds; `abs` catches −90; white fill matches the thumbnail's page background; one output context sets the saved PNG's pixel format for photos and PDFs, and the idea file's "Transparent polygon" changes one function (review M-2) | Keeping `page.thumbnail(of:for:)` and moving it off main — still renders the whole page; a second hand-built `CGContext` in the PDF draw — a copy of `scaledToFit`'s; returning nil for a sub-pixel box — `isValidPolygon` accepts it, so Save would refuse a region the researcher sees |
| D-5 | `DepictionPreviewMemo` holds `key: String` and `cut: (cgImage: CGImage, image: NSImage)?`; `cut == nil` means the cut failed for that key. `depictionPreviewMemo` loses `@ObservationIgnored`. | One optional admits no half-failed state (review M-6); the memo is now written from a task, so the view must observe it; a stored failure stops a retry loop for the same key (the key changes when the viewer finishes loading, so a cut tried while loading is tried again) | Two optionals, both nil for a failure — admits `cgImage` set and `image` nil; a separate `isCutting` flag — the state is derivable from "no memo for the current key" |
| D-6 | `private var unchangedDepictionImagePath: String?` *(new)* is the one "unchanged saved likeness" rule: it returns `currentDepictionDraft.imageRelPath` (which may be empty) when `dialogBase() == .edit`, `!isDepictionRedrawn`, and the draft `hasRegion`; nil otherwise. `private func depictionCutKey(for draft: DepictionValueDraft) -> String?` *(new)* returns nil unless `draft.hasRegion`, else the existing key string (artifact id, viewer kind, viewer loading, locator JSON). `var depictionPreviewKey: String?` *(new)* is nil when there is no depiction draft or `unchangedDepictionImagePath != nil`, else `depictionCutKey(for:)` of the current draft. | One predicate for the trigger, the read, and the state, so a cut never runs under a saved tile (review M-3); an unchanged edit — with or without a readable saved image — never cuts, which is what `saveEdit` sends (no image unless redrawn); two names for the two nil rules (review M-4) | A predicate that also requires a non-empty path — an unchanged edit with an empty path would then cut a preview that Save never sends; cutting on every edit open — the cost PRF-1 names |
| D-7 | `enum DepictionPreviewState: Equatable { case none, cutting, cut(NSImage), failed, saved(relPath: String) }` *(new)*, in `CitationComposerModel+Depiction.swift`; `var depictionPreviewState: DepictionPreviewState` *(new)* replaces `depictionPreview`: `unchangedDepictionImagePath` non-empty → `.saved(relPath:)`; empty → `.none`; else `depictionPreviewKey` nil → `.none` (no region, or a hidden region, as today's empty preview); memo key equal to `depictionPreviewKey` → `.cut(image)` or `.failed`; otherwise `.cutting`. | `body` reads stored state only; the cases are testable without a view; `.saved` and the key read the same predicate | Returning `PVThumbnail.Content` from the model — puts a kit type in the model |
| D-8 | `func depictionImage(for draft: DepictionValueDraft) async -> CGImage?` *(new)* is the one cut entry: nil when `depictionCutKey(for: draft)` is nil; the memo's `cut?.cgImage` when the memo's key equals that key (nil for a stored failure); otherwise it awaits `ArtifactRegionCrop.render`, and writes the memo (with `cut: nil` on a failed cut) only when `depictionCutKey(for: draft)` still equals `depictionPreviewKey` after the await, and returns the cut either way, so a save uses its own draft's pixels even when the preview key has moved on. `func renderDepictionPreview() async` *(new)* calls it for `currentDepictionDraft` when `depictionPreviewKey` is non-nil. `withDepictionInput` calls `depictionImage(for: values.depictionDraft)` and throws `DepictionRenderFailed` on nil. `renderedDepiction(for:)` and `depictionPreview` are deleted. | Preview and save share one function and one stored result, so the saved pixels are the previewed pixels; a stale cut never overwrites the current one (AC-1.5) | A stored `Task` handle the save awaits — cancellation bookkeeping for the few hundred milliseconds between commit and Save; a save pressed then runs the same deterministic cut (AC-2.3) |
| D-9 | `CitationDepictionValueArea` adds `.task(id: model.depictionPreviewKey) { await model.renderDepictionPreview() }` on its outer `VStack`, and maps `depictionPreviewState` in a private `@ViewBuilder var preview` *(new)*: `.none` and `.failed` → `PVThumbnail(PVThumbnail.Content(label: depictionPreviewLabel), size: 96)`; `.cutting` → the same with `loading: true`; `.cut(image)` → the same with `image: Image(nsImage: image)`; `.saved(relPath)` → `CachedThumbnail(projectDir: model.session.projectKey.projectDir, relPath: relPath, mediaType: "image/png", size: 96)` with `.accessibilityElement(children: .ignore)` and `.accessibilityLabel(Text(L10n.CitationComposer.depictionPreviewLabel))`, so VoiceOver reads the preview label while it loads and when the file is missing. A missing or undecodable saved image shows `CachedThumbnail`'s PNG file-type mark, as the saved line does. | SwiftUI cancels a stale task when the key changes and starts one for every trigger (region commit, Citation switch, viewer load, dialog open) with no model hook per trigger; both thumbnails are existing kit and platform views; the label holds in every state (review M-7 outside note) | A cut started from `commitViewerRegion` — misses the viewer-load and dialog-open triggers; `CachedThumbnail`'s `imageLabel` — labels only the image state |
| D-10 | The detached cut is not interrupted when its key changes; its result is dropped (D-8). | Core Graphics drawing cannot be cancelled mid-draw | `withTaskCancellationHandler` around the cut — nothing to cancel |
| D-11 | Tests read pixels through `static func rgbaBytes(_ image: CGImage) -> [UInt8]` *(new)* on `ArtifactFileFixtures`: the body of today's `ArtifactRegionCropTests.pixels(_:)` moved unchanged (DeviceRGB, 8 bits, `bytesPerRow: width * 4`, `premultipliedLast`, `[]` when the context fails), returning the flat RGBA bytes; `pixels(_:)` becomes a stride over it. | AC-2.1 and AC-2.2 compare pixels in a second suite; one RGBA reader instead of two | A second private reader in `CitationComposerDepictionTests` |
| D-12 | If the build warns that `CGImage` is not `Sendable` where it crosses the detached task (in `Snapshot.image` or the task's result), `ArtifactRegionCrop` gains `private struct SendableImage: @unchecked Sendable { let image: CGImage }` *(new)*, used at exactly those two crossings, with a comment that `CGImage` is immutable. Then `Snapshot` reads `case image(SendableImage)`, the detached task returns `SendableImage?`, and `render` unwraps it before returning `CGImage?`; `cut` still takes a `Snapshot`, and T-3 (which checks only the `.pdf` case) does not change. No other suppression is added. | `CGImage` is immutable, so the box is sound; the stop condition then fires only on a real isolation problem | Stopping on the warning — the same crossing already compiles in `withDepictionInput`; `@preconcurrency import CoreGraphics` — silences every CoreGraphics warning in the file |
| D-13 | `depictionPreviewKey` is computed twice per `body` pass (the state and `.task(id:)`); it is not cached. | It builds one short string from stored values; caching adds an invalidation rule | A stored key updated on every trigger — the bookkeeping D-9 avoids |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ArtifactRegionCrop.Snapshot` *(new)* | type | `ArtifactViewerContent` (`@unchecked Sendable`, holds `NSImage` / `PDFDocument`); `ArtifactViewerSource` (what the viewer loads) | Its cases are not safe off the main actor; `Snapshot` holds only `CGImage` and `URL` |
| `ArtifactRegionCrop.snapshot(_:page:)` *(new)* | func | `ArtifactRegionCrop.render` (the viewer half) | It is the viewer half of `render`, split out (D-1) |
| `ArtifactRegionCrop.cut(_:region:)` *(new)* | func | `ArtifactRegionCrop.render`, `pageBitmap` | It is the work half of `render`; `pageBitmap` is replaced by the box draw (D-4) |
| `ArtifactRegionCrop.outputContext(width:height:)` *(new)* | func | the `CGContext` inside `ArtifactRegionCrop.scaledToFit` | Extracted from it, so the PDF box draw and the scale share one pixel format (D-4) |
| `SendableImage` *(new, only if D-12 applies)* | type | `ArtifactViewerContent` (`@unchecked Sendable`) | Wraps one immutable `CGImage` for the two crossings (D-12) |
| `DepictionPreviewState` *(new)* | type | `PVThumbnail.Content` | A model type, not a kit type (D-7) |
| `depictionPreviewKey`, `depictionCutKey(for:)`, `unchangedDepictionImagePath` *(new)* | property, func, property | the key in `renderedDepiction(for:)`; `isDepictionRedrawn` | Same string, moved so the trigger, the read, and the save share it; the unchanged-edit rule is written once (D-6) |
| `depictionImage(for:)`, `renderDepictionPreview()` *(new)* | func | `renderedDepiction(for:)` | Async replacement (D-8) |
| `ArtifactFileFixtures.rgbaBytes(_:)` *(new)* | test helper | `ArtifactRegionCropTests.pixels(_:)` (private) | Moved to the shared fixture so two suites read pixels one way (D-11) |

**Builds on.** `ThumbnailCache`'s detached decode and `CachedThumbnail`'s `.task(id:)` load; `ArtifactFileFixtures` (add-swift-test: artifact files on disk come from it); `PVThumbnail.Content(loading:)`.
**Deletes.** `ArtifactRegionCrop.pageBitmap`, `ArtifactRegionCrop.maxPageLongEdge`, the hand-built context inside `scaledToFit`, `CitationComposerModel.depictionPreview`, `renderedDepiction(for:)`, `@ObservationIgnored` on `depictionPreviewMemo`.
**Next adds.** Audit UX-6 ("show a failed cut before Save") reads `depictionPreviewState == .failed` with no new state. The idea file's "Transparent polygon" crop changes only `ArtifactRegionCrop.cut` and `outputContext`. A Windows client keeps the same snapshot / cut split.

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
| Preview, saved | `CachedThumbnail` (`size: 96`) with `.accessibilityElement(children: .ignore)` and `.accessibilityLabel` | same | The saved line's view at the preview's size; the PNG file-type mark when the file is missing |
| Cut trigger | `.task(id: model.depictionPreviewKey)` | same | On the outer `VStack` |

States: no region → empty preview, empty caption; region just committed → loading preview; cut ready → the cut; cut failed → empty preview (the caption is unchanged; UX-6 is Out); unchanged edit → the saved image, loaded through `ThumbnailCache` (its first frame is the empty dash; a cache hit then shows the image at once, a miss shows the loading bar until the decode finishes), or the PNG file-type mark when that file is missing; unchanged edit with no saved image → empty preview. Accessibility: the preview keeps the label `depictionPreviewLabel` in every state. The empty caption (`depictionCaptionEmpty`) is unchanged.

### Strings

None — no new copy.

## 7. Steps

**STEP-1 — The app cuts off the main actor and reads a stored cut**
Files: `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift`, `macos/App/Features/CitationComposer/CitationComposerModel+Depiction.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/App/Features/CitationComposer/CitationComposerAuthoringBox.swift`. Skill: `add-ui-component`.
Do: in `ArtifactRegionCrop`, add `Snapshot`, `snapshot`, `cut`, `outputContext`, and the async `render` (D-1 to D-4); move `scaledToFit` onto `outputContext`; delete `pageBitmap` and `maxPageLongEdge`; rewrite the enum's header comment to "Cuts a region out of viewer content: a pixel crop of an image, or a Core Graphics draw of only the region's box on a PDF page, off the main actor." and the `render` / `cut` doc comments to match. In the model, change `DepictionPreviewMemo` (D-5) and drop `@ObservationIgnored`; add `unchangedDepictionImagePath`, `depictionCutKey(for:)`, `depictionPreviewKey` (D-6), `DepictionPreviewState` and `depictionPreviewState` (D-7), `depictionImage(for:)` and `renderDepictionPreview()` (D-8); rewrite `withDepictionInput` onto `depictionImage(for:)`; delete `depictionPreview` and `renderedDepiction(for:)`. In `CitationDepictionValueArea`, add the `.task(id:)` trigger and the `preview` view (D-9). Before editing, build `main` once and save the log (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO 2>&1 | tee /tmp/s10-65-base.log`). These edits land together: `render` turning async and `depictionPreview` being deleted break their callers until all four files are done, so this step has one check at its end.
Check: the same build command, teed to `/tmp/s10-65-head.log`, succeeds; `grep -E 'warning: .*(Sendable|actor-isolated|nonisolated)' /tmp/s10-65-head.log | grep -E 'ArtifactRegionCrop|CitationComposerModel|CitationComposerAuthoringBox' | wc -l` is not above the same count on `/tmp/s10-65-base.log` (apply D-12 first when the new warnings name `CGImage`).

**STEP-2 — Tests**
Files: `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ArtifactRegionCropTests.swift`, `macos/ProvenenciaTests/CitationComposerDepictionTests.swift`. Skill: `add-swift-test`.
Do: add `rgbaBytes` (D-11) and rewrite `pixels(_:)` on it. Make `imageCropIsTheBoundingBox`, `cropIsCappedAt2048`, and `pdfRegionFollowsTheDisplayedPage` `await` the render with their assertions unchanged; add T-2 and T-3. In `drawingLeavesTheCitationRegion`, replace `model.depictionPreview == nil` with `model.depictionPreviewState == .none` and the 4×4 size check with `await model.renderDepictionPreview()` followed by `guard case .cut(let image) = model.depictionPreviewState` and a 4×4 size check. Add `import ImageIO` to `CitationComposerDepictionTests.swift`. Add T-4 to T-10.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CitationComposerDepictionTests -only-testing:ProvenenciaTests/ArtifactRegionCropTests -only-testing:ProvenenciaTests/CitationAutoTranscribeTests -only-testing:ProvenenciaTests/ArtifactViewerModelTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-3 — Docs**
Files: `docs/ideas/depictions-and-likenesses.md`. The Spike 10 decision ("The app renders the region") still holds.
Do: in the "Windows" bullet, replace "Spike 10 draws it with PDFKit on the Mac." with "Spike 10 draws it with Core Graphics on the Mac." The plan Status and the S10-65 map row move in implement-pr step 8's commit, after the draft PR has a number.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-65-depiction-preview-off-main.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS` on a macOS host.

Commits, in order: STEP-1 and STEP-2 together, staged once STEP-2's check passes ("Cut the likeness preview once off the main thread, draw only the PDF box, and show a saved likeness without cutting it."). The test bundle reads the old API until STEP-2, so the two steps are one commit. Then the step-8 commit.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `ArtifactRegionCropTests` `imageCropIsTheBoundingBox`, `cropIsCappedAt2048`, `pdfRegionFollowsTheDisplayedPage` (now `await` the render) | Assertions unchanged: 4×4 and red top; 2048×512; rotated black square on the top-right box, 2:1 sides | AC-1.3, AC-1.4, D-1, D-2, D-4 |
| T-2 | Swift | `ArtifactRegionCropTests` `pdfCutDrawsOnlyTheBox` *(new)* | `writeMarkedPDF(rotation: 0)`, load, render page 1 region `rectangle(0, 0, 0.05, 0.1)` → 20×80; `rgbaBytes` first pixel black (R, G, B < 60), last pixel white (R, G, B > 200) | AC-1.2, D-4 |
| T-3 | Swift | `ArtifactRegionCropTests` `pdfSnapshotNeedsTheLoadedDocument` *(new)* | A fresh `ArtifactViewerModel` → `snapshot(viewer, page: 1)` is nil; after loading the marked PDF, `guard case .pdf(let url, let pageIndex) = snapshot(viewer, page: 1)` with `pageIndex == 0` and `url == viewer.pdfDocument?.documentURL`; `snapshot(viewer, page: 0)` is nil; after `unload()`, `snapshot(viewer, page: 1)` is nil | D-3 |
| T-4 | Swift | `CitationComposerDepictionTests` `previewIsCutAfterTheRegionCommits` *(new)* | `seeded()`, Likeness draft, `startDepictionDrawing()`, `commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))` → `depictionPreviewState == .cutting`; `await renderDepictionPreview()` → `.cut` with size 4×4 | AC-1.1, D-7, D-8 |
| T-5 | Swift | `CitationComposerDepictionTests` `savedPixelsAreThePreviewPixels` *(new)* | Photo: commit, `await renderDepictionPreview()`, then `try await model.withDepictionInput(for: draftValues(model)!)` with a closure that decodes `input.imagePath` through `CGImageSourceCreateWithURL` + `CGImageSourceCreateImageAtIndex` → width, height and `ArtifactFileFixtures.rgbaBytes` equal those of `depictionPreviewMemo?.cut?.cgImage`. PDF: `seeded(.pdf)`, overwrite `objects/news.pdf` with `writeMarkedPDF(rotation: 0)` before `prepare()`, commit `rect(0, 0, 0.5, 0.5)` on page 1, same comparison | AC-2.1, AC-2.2, D-8 |
| T-6 | Swift | `CitationComposerDepictionTests` `savingSendsAPNGAndDeletesIt` (unchanged) | Saves right after the commit with no cut run → ` depiction=1 depictionFile=present`, file removed | AC-2.3, D-8 |
| T-7 | Swift | `CitationComposerDepictionTests` `editWithoutRedrawShowsTheSavedImage` *(new)* | On the `editingSavedLikeness()` model: `depictionPreviewKey == nil`, `depictionPreviewState == .saved(relPath: savedRelPath)`; then `startDepictionDrawing()`, `commitViewerRegion(rect(0.5, 0.5, 0.9, 0.9))` on the same model → `.cutting`; `await renderDepictionPreview()` → `.cut`. A second model seeded as `hiddenRegionAsksForARedraw` seeds it (`likenessObservation(region: "")`, `observationID: "obs-1"`) → `.none` | AC-3.1, AC-3.2, D-6, D-7 |
| T-8 | Swift | `CitationComposerDepictionTests` `staleCutIsNotStored` *(new)* | Commit A = `rect(0.25, 0.25, 0.75, 0.75)`, keep `let a = draftValues(model)!.depictionDraft`; draw and commit B = `rect(0, 0, 1, 1)`; `_ = await model.depictionImage(for: a)` → `depictionPreviewState == .cutting` and `depictionPreviewMemo == nil`; `await renderDepictionPreview()` → `.cut` with size 8×8 | AC-1.5, D-8 |
| T-9 | Swift | `CitationComposerDepictionTests` `missingFileCutFails` *(new)* | `seeded()`, delete `objects/photo.png` under `projectDir` before `prepare()`; Likeness draft; `commitViewerRegion(rect(0.25, 0.25, 0.75, 0.75))`; `await renderDepictionPreview()` → `.failed`; `await saveAuthoring()` → the draft's error is `L10n.string(L10n.CitationComposer.depictionRenderFailed)` and its region is kept | AC-2.4, D-5 |
| T-10 | Swift | `CitationComposerDepictionTests` `editWithNoSavedImageDoesNotCut` *(new)* | `let (projectDir, store) = try seeded()`; `let region = regionJSON(rect(0.1, 0.1, 0.4, 0.4))`; `var observation = likenessObservation(region: region)`; `observation.depiction = CatalogDepictionValue(regionJSON: region, imageRelPath: "")` (as the suite does near line 257); `seedCitation(store, observations: [observation])`; `makeModel(store: store, projectDir: projectDir, citationID: "cit-1", observationID: "obs-1")`; `await model.prepare()` → `depictionPreviewKey == nil` and `depictionPreviewState == .none` | AC-3.3, D-6 |

Must stay green: everything `scripts/check-pr.sh` runs. The suites most likely to break: `CitationComposerDepictionTests` (`redrawSendsANewImage`, `refusedSaveKeepsTheRegion`, `keepEditingAfterACitationSwitchKeepsTheRegion`), `ArtifactRegionCropTests`, `CitationAutoTranscribeTests` (shares `ArtifactRegionCrop.cgImage(from:)` and `boundingBox`), `ArtifactViewerModelTests` (the viewer is read, not edited).

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
- [ ] Domain doc: none changes (STEP-3)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with a project that has a photo Source and a scanned multi-page PDF Source (a newspaper or register page scan of at least 2,000 px on its long edge).

1. Open the PDF Source's Citation on page 1 in the composer. Add a Likeness. Press **Draw on the file** and drag a box around one face. The dialog returns at once; the preview shows the loading state, then the face.
2. Draw again, a larger box. The preview shows the new face, never the old one.
3. Before pressing Save, take a screenshot of the preview. Save. Edit the saved likeness: its preview (the saved image) shows the same face and the same crop as the screenshot.
4. In that edit, change only the person. The preview shows the saved face from the thumbnail cache, with no cut: no cut loading bar appears, and the saved face shows as soon as the thumbnail cache returns it.
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
- The STEP-1 build reports a `Sendable` or actor-isolation error, or a new warning that D-12's `SendableImage` does not remove, for `ArtifactRegionCrop.Snapshot` or the detached cut.
- T-1's rotated-page assertions fail with the box draw of D-4.
- T-5's pixel comparison differs for either the photo or the PDF.
- `ArtifactViewerModel.load` no longer resolves its file through `ProjectFiles.objectURL` (S10-62 changes that function's checks, not its signature).
- `CitationComposerModel+Depiction.swift` already has `useDepictionRegion` (S10-63 landed first) and its call into `commitViewerRegion` changed shape: merge `main`, keep both, and report the merge in the Landed note.

## 12. Open questions

None.

## 13. PR

**Title:** S10-65 — Render the likeness preview off the main thread
**Body:** objective in one sentence; what shipped against US-1, US-2, US-3; AC evidence (test names); deviations from the plan; manual check results; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-65-depiction-preview-off-main.md`.

## Landed

Differences from the plan:

- The cut key is a typed `Equatable` value (`DepictionCutKey`: artifact id, viewer kind, loading flag, page and region draft), not `locator.encodeJSON()` as D-6 says: `encodeJSON()` has unstable key order (no `.sortedKeys`), so a string key never matched itself. Approved by the user against D-6's text.
- `main` had moved since the plan: S10-63 (`useDepictionRegion`; `commitViewerRegion` now returns `Bool`) and S10-62 (the `fileExists` guard in `ArtifactViewerModel.load` is gone) had landed. Both were kept and no Current state row was invalidated.
- The PDF cut draws through `PDFPage.draw(with:to:)` on a task-owned `PDFDocument`, not `CGContext.drawPDFPage`, so annotations the viewer shows are in the likeness (CON-1; user's call against D-3/D-4's text).
- D-12's `SendableImage` was not needed: the build has no new Sendable or isolation warnings.
- The manual check (section 10) is left for the user.
- STEP-3's idea-doc edit ("Core Graphics" in the Windows bullet) was undone by F-4, since the cut draws with PDFKit. The idea doc has no net change.
