# Review — S14-13 A crafted project can't crash the app

**Verdict:** Fits with changes — Ready with changes (round 2); every finding fixed, no blocker open.
**Rounds:** 2. **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- D-10 bends a working rule: the viewer checks against Go's media type, or, when macOS has no type for it, the extension Go assigned from the same table. **User, 2026-10-29: media type, else extension.** STEP-8 rewords the rule and the scope entry.
- Answered before planning (user, 2026-10-29): the viewer decodes once at no more than 64 MP for display, crop and OCR (D-8, replacing the map's "screen-sized decode"); the viewer refuses an image declaring over 256 MP, so the 198 MP scan stays viewable (D-7).

## Writer's own calls
- The grid range is ±1,000,000 cells (the audit's number) (D-1).
- A panic inside `catalogsession.Do` drops the session so the next call reopens with fresh caches, not only an unlock (D-5).
- The panic log carries the method number, the value and `debug.Stack()` (D-4).
- A refused image reuses `.loadFailed`, with no new copy (D-9).
- The sniffed type is checked against the object's extension, which Go assigns from its stored-type table, not against `UTType(mimeType:)` (D-10).
- The viewer's orientation matches `NSImage(contentsOf:)` today, pinned by a test, so saved regions do not move (D-12).
- The patch bump is 0.2.1 → 0.2.2.

## Round 1

### M-1 · blocker · The catalog's cell limit goes into GraphCanvas
The limit was on `GraphCanvasGridMapping`; `macos/App/Features/GraphCanvas/README.md` keeps catalog types and position tables out of `GraphCanvas/`.
**Answer:** Fixed — D-1 and D-3: `cellLimit` and `clampedCell` are on `EvidenceGraphPlacement`; `GraphCanvasGridMapping` keeps only generic `Int64` saturation (`offsetCell`, a saturating `gridCell`). The vocabulary row points at `EvidenceGraphPlacement`.

### M-2 · blocker · Two ImageIO bounded-decode paths and two header reads
**Answer:** Fixed — D-11: `ThumbnailDecodePolicy.declaredImage(_:maxEdge:maxPixels:)` (one properties read, with size, DPI and orientation) and `decode(_:maxPixelSize:applyTransform:)`, used by `ThumbnailCache` and `ArtifactImageDecode`. `ArtifactImageDecode` keeps only the viewer's policy.

### M-3 · blocker · A third close-a-session body
**Answer:** Fixed — D-5: `closeSession` splits into `closeLocked` and the locking wrapper; `forget(dir, s)` drops the registry entry only if it is still `s`.

### M-4 · major · The viewer would hold five parallel facts about one image
**Answer:** Fixed — D-13, D-14: `ArtifactViewerContent.image(ArtifactImageDecode.Decoded)`; `displayImage`, `displayRaster` and `imagePixelSize` are computed from `content`; `source` is stored once.

### M-5 · major · The ffi-handler skill and rule become false
**Answer:** Fixed — STEP-8 updates `.agents/skills/add-ffi-handler/SKILL.md` and `.cursor/rules/ffi-handlers.mdc`; the DoD lists it.

### M-6 · major · The sweep misses `maxDecodePixels` in spike docs
**Answer:** Fixed — STEP-8 renames the constant in `oversized-jpeg-thumbnails.md` and `deployment-plan.md`; the sweep is a recursive `rg` over `docs .agents .cursor core api macos/App`.

### M-7 · minor · OCR reads two forms of the same image
**Answer:** Declined in part. `imageRaster` stays: `CitationComposerModel.imageRaster` and `needsWholePageWarning` read it (D-16). The points-versus-pixels oversize warning is unchanged behavior, so it is an Out line pointing to the depictions idea file, not a question for this PR.

### M-8 · minor · The wire clamp in `GoStore` can never fire
**Answer:** Fixed — the `GoStore` clamp is dropped (D-3; Out line).

### M-9 · minor · `ArtifactImageDecode` tucked into the model file
**Answer:** Fixed — D-17: its own file, with a hand-written `project.pbxproj` entry while S14-15 has not landed.

### M-10 · minor · A new regex shape in the vocabulary test
**Answer:** Fixed — the Swift literals are written without `_` and the rows use `(\d+)` (D-1, D-6).

### M-11 · nit · The PDF snapshot could carry `ArtifactViewerSource`
**Answer:** Fixed — D-15: `.pdf(source:pageIndex:)`.

### X-1 · blocker · D-10's conformance check meets a media type with no UTType
**Answer:** Fixed — D-10 checks against `UTType(filenameExtension:)` of the object's extension. Go's table assigns `.tiff` to `image/tif` and `.heif` to `image/heif`, so every stored type maps. Only HEIC-sniffed-as-HEIF stays a Stop condition, with a fixture.

### X-2 · blocker · How viewer content and the detached load carry the decode
**Answer:** Fixed — D-14 states the content shape, the `Sendable` story, where `source` is set (after `objectURL`, every kind) and what `installImageRasterForTesting` does.

### X-3 · major · Orientation and size left to be found empirically
**Answer:** Fixed — D-12 and D-13: `appliesOrientation` is one constant pinned by T-11. `pixelSize` and `pointSize` swap for orientations 5–8 only when it is true. The budget and `maxPixelSize` are orientation-free. DPI and orientation come from `declaredImage`'s single properties read.

### X-4 · major · T-4 does not compile against `recovered`
**Answer:** Fixed — D-4 gives the full signature and the panic result `(nil, internal.unknown)`; T-4 uses `func() ([]byte, error)`; the log includes the stack.

### X-5 · major · STEP-4 omits `raster_test.go`
**Answer:** Fixed — §6 Go core and STEP-4 list `raster_test.go` and the doc comments.

### X-6 · major · D-5's panic path has unmade choices
**Answer:** Fixed — D-5: a `done` flag (no `recover`, so the stack is kept), `closeLocked` under the held lock, the close error dropped, and the defer covers `warmGraph`.

### X-7 · major · The "unchanged placement" Edges sentence is false
**Answer:** Fixed — §6 Edges: the upsert rewrites the row to the clamped value and returns no change; T-1 reads the raw column.

### X-8 · major · STEP-3 misses some arithmetic
**Answer:** Fixed — §6 macOS names every site, including `minCell.gridY + 2`, the hover and create paths, and the constants that stay. STEP-3 runs an `rg` and lists the kept hits in the PR description.

### X-9 · minor · Smaller gaps
**Answer:** Fixed — `installImageRasterForTesting` nil case (D-14); the fixtures are decided (a grayscale 12,000 × 6,000 JPEG; a shared `pngHeaderData(width:height:)` with CRC); an Out line for the 40 M-point board; version 0.2.2; the map link in STEP-8 (and made at planning time).

## Round 2

Both reviewers confirmed the round-1 answers against the code. New findings:

### M-12 · blocker · D-10 bends the working rule on the viewer's type check
**Answer:** To user → answered (2026-10-29): media type, else extension. D-10 rewritten; listed under **Rules this plan bends**; STEP-8 rewords the working rule and the S14-13 scope entry; DoD line.

### M-13 · blocker · New image fixtures sidestep `ArtifactFileFixtures`
**Answer:** Fixed — §6 Codebase fit: `ArtifactFileFixtures.writeImage` gains `dpi:`, `orientation:`, `grayscale:`; `writeDeclaredPNGHeader` / `declaredPNGHeaderData` live there; STEP-4 updates the `add-swift-test` fixture line.

### M-14 · blocker · A second NaN-and-saturate floor conversion
**Answer:** Fixed — D-3: `GraphCanvasGridMapping.saturatingFloor(_:)`, used by `gridCell` and by `maxGridX` / `maxGridY`; T-5 tests it.

### Below the bar (maintainability)
**Answer:** Fixed — `maxDecodedPixels` renamed `maxDecodePixels`; the viewer's edge constant is `maxEdge`; `applyTransform:` and `appliesOrientation` go if T-11 settles on `true`; the migrations folder is excluded from the stop condition; the sweeps exclude this plan, its log and the audits, and name the unrelated `GraphCanvasItemView` hit; `cgImage(from:)` keeps only its test-hook caller (noted under Deletes).

### X-10 · blocker · T-9 expects 7999 where D-12 gives 8000
**Answer:** Fixed — T-9 and Edges expect 8000; ImageIO scales the short edge.

### X-11 · major · The test hook's nil case breaks the whole-page OCR test
**Answer:** Fixed — D-14: `installImageRasterForTesting(_:raster:)`; the whole-page case passes a 1 × 1 raster; STEP-6 changes it.

### X-12 · major · DPI of 0 or absurd values
**Answer:** Fixed — D-11: DPI only when finite and in 1…100,000, else 72; orientation outside 1…8 → 1; T-8 case for `dpi: 0`.

### X-13 · major · Sweeps hit the plan itself
**Answer:** Fixed — STEP-8 and DoD sweeps exclude the plan, its log and the audits.

### X-14 · major · PDF snapshot non-nil while loading; rule contradiction
**Answer:** Fixed — D-15 requires `pdfDocument != nil` as today; the rule wording went to the user (M-12).

### X-15 · minor · Fixtures, tolerances, path constants
**Answer:** Fixed — fixtures through `ArtifactFileFixtures`; 0.01 tolerances for DPI and point size; T-10's fixture named; `swiftEvidencePlacement` and `swiftArtifactImageDecode` path constants named.

### X-16 · minor · STEP-3's `rg` misses sites
**Answer:** Fixed — STEP-3 runs a broader `rg` and lists kept hits.

### X-17 · minor · D-5 spin, defer placement, webp
**Answer:** Fixed — Edges describes the spin and where the defer is registered; the Stop condition notes `image/webp` is untested.
