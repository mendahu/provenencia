# S14-13 — A crafted project can't crash the app

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a refused image uses the viewer's existing load-failed state; no control, copy or layout changes)
**Board:** —
**Depends on:** S12-23 (landed [#608](https://github.com/mendahu/provenencia/pull/608))
**Blocks:** S14-03
**Branch:** `s14-13-crafted-project-safe`
**Version bump:** patch — crash and memory fixes on crafted input, no feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/catalog-session.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/versioning.mdc`
**Context:** audit [A-26](../app-health-audit.md#a-26--crafted-graph-positions-crash-the-app), [A-27](../app-health-audit.md#a-27--the-artifact-viewer-decodes-originals-unbounded), [A-28](../app-health-audit.md#a-28--a-go-panic-kills-the-app); [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) (the PDF cut's file check); working rules "What the app hands to macOS to open is checked first, and Go decides the type", "Audit items keep their IDs"; map review K-8, K-25; bar outcome 15
**Planned:** 2026-10-29 against `main` @ `b8ca4cfc`
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

A project folder someone shares can hold a catalog row or a file the app did not write. Today three such inputs crash the app or exhaust its memory: a graph position near the edge of `Int64` traps Swift arithmetic (A-26), an image that declares a huge pixel size is decoded in full by the Artifact viewer (A-27), and a panic anywhere in the Go core kills the process, because the core is a c-shared library inside the app (A-28). After this PR each of those fails as an ordinary error, and the PDF cut opens its file through the same checked path as every other read.

**Why now.** Batch 5's first audit item. S14-03 gates Draw on the "can cut" predicate this PR changes in place (map review K-25).

**What stays.** Every position the app itself writes, every image up to 64 megapixels and every PDF look, zoom, crop and transcribe exactly as today. The thumbnail cache keeps its own PNG and JPEG list.

## 2. User stories

**US-1.** As a researcher who opens a shared project with crafted graph positions, the Evidence graph opens with the cards held to the edge of the grid instead of crashing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A `subject_positions` row holds `(9223372036854775807, -9223372036854775808)`; `ListSubjectPositions` runs | It returns `(1000000, -1000000)` |
| AC-1.2 | `SetSubjectPosition` with `grid_x = 1000001` | `subjectpositions.invalid`; the row is unchanged |
| AC-1.3 | `SetSubjectPosition` with `(1000000, -1000000)` | Saved and returned unchanged |
| AC-1.4 | The Swift placement and move code is fed `Int64.max` / `Int64.min` cells (a FakeStore or wire row the clamp did not see) | No trap; every cell it returns is within ±1,000,000 |
| AC-1.5 | A card at `gridX = 1000000` is moved right by the arrow key | The patched cell stays `1000000`; no store error is shown |

**US-2.** As a researcher who opens an image Artifact, the viewer decodes it within a bound, and an image that lies about its type or size shows the load-failed state instead of exhausting memory.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A PNG of 3000 × 2000 px at 300 dpi, `media_type` `image/png` | Shows as today: `displayImage.size` equals `NSImage(contentsOf:)`'s size (720 × 480 pt); the composer's detail reads "3000 × 2000"; a crop of the top half is 3000 × 1000 px before the 2048 cap |
| AC-2.2 | A PNG file stored with `media_type` `image/jpeg` | `emptyReason == .loadFailed`, `displayImage == nil`, no decode ran |
| AC-2.3 | A PNG header declaring 20,000 × 20,000 (400 MP) | `.loadFailed` before any pixel decode |
| AC-2.4 | A JPEG of 12,000 × 6,000 (72 MP, within the 256 MP budget) | Displays; the decoded raster has at most 64,000,000 pixels (about 11,313 × 5,656) and keeps the aspect; `displayImage.size` is the full image's point size; the detail reads "12000 × 6000" |
| AC-2.5 | A JPEG with EXIF orientation 6 | Displays with the same orientation and `size` as `NSImage(contentsOf:)` gives for that file |

**US-3.** As the engine, a panic in a handler returns `internal.unknown` and the project stays usable.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A function run through `ffi.Call`'s recovery panics | `Call` returns an error whose code is `internal.unknown`, kind internal; the process lives |
| AC-3.2 | A `catalogsession.Do` callback panics | The panic reaches the caller; the session's mutex is released; the next `Do` on the same project succeeds on a freshly opened catalog |

**US-4.** As the engine, the PDF cut opens its file through `ProjectFiles.objectURL`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A loaded PDF Artifact; `ArtifactRegionCrop.snapshot` | Carries the project folder, the object's path and the 0-based page, not a URL |
| AC-4.2 | The object file is replaced by a symlink after load; the region is cut | `cut` returns nil |
| AC-4.3 | The object is unchanged; a region is cut | The same box as today (`pdfCutDrawsOnlyTheBox` keeps passing) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-26: `subjectpositions` refuses outside ±`MaxCell` on write and clamps on read | STEP-1, AC-1.1 – 1.3 |
| A-26: Swift clamps positions at the wire and uses saturating cell arithmetic in placement and moves | STEP-3, AC-1.4, AC-1.5 |
| A-28: `ffi.Call` recovers a panic into `internal.unknown` | STEP-2, AC-3.1 |
| A-28: `catalogsession.Do` releases its lock with `defer` and drops a session whose callback panicked | STEP-2, AC-3.2 |
| A-27: one shared declared-size check in `ThumbnailDecodePolicy` | STEP-4 |
| A-27: the viewer checks sniffed type and the 256 MP budget before decoding, and decodes once at no more than 64 MP for display, crop and OCR | STEP-5, STEP-6, AC-2.1 – 2.5 |
| The PDF cut reopens its file through `ProjectFiles.objectURL` | STEP-7, AC-4.1 – 4.3 |

| Out | Home |
| --- | --- |
| Length caps on catalog text (audit L-10) | not planned (map scope entry) |
| The foreign-catalog open order (audit L-7) | not planned (map scope entry) |
| Draw blocked on a file the viewer cannot cut; the failed-cut state | S14-03 |
| Bounding PDFKit's own parse and render of a crafted PDF | not planned: the cut is already capped by `maxPageScale` and `maxLongEdge`, and PDFKit renders per page |
| A `recover` in `api/libprovenencia/main.go` around `EncodeError` | not planned: `ffi.Call` covers every handler; `EncodeError` marshals a fixed message |
| FakeStore's connect midpoint (`FakeStore.swift` `floorDiv(fromPos.gridX + toPos.gridX + 1, 2)`) | not planned: FakeStore positions are test-seeded |
| The JPEG reduced decoder and raster's byte limit | S14-23 |
| A new empty-state reason or copy for a refused image | not planned (D-9) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Position write | `core/database/subjectpositions/subjectpositions.go` `Set`, `SetTx` | Accepts any `int64`. Callers: `api/ffi/handlers/subjects.go` (`SetSubjectPosition`, `CreateSubject` placement through `core/database/subjects/subjects.go` `Create`), `core/database/connect/connect.go` (bridge midpoint) |
| Position read | same file `scanPosition` (used by `Get`, `GetTx`, `ListByEvidenceGraph`) | Returns stored values unchanged. `connect.go` adds two read values (`fromPos.GridX+toPos.GridX+1`), which wraps in Go on extreme rows |
| Position tests | `core/database/subjectpositions/subjectpositions_test.go` `TestSubjectPositions`; `api/ffi/handlers/subjects_test.go` `TestSubjectPositions` | No bound cases |
| FFI dispatch | `api/ffi/dispatch.go` `Call` | One `switch` over methods; no `recover`. `api/libprovenencia/main.go` `provenencia_call` calls it |
| Dispatch tests | `api/ffi/dispatch_test.go` `TestCallRouter` | Unknown-method cases only |
| Session | `core/catalogsession/session.go` `Do` | `s.mu.Lock()`, `fn`, `warmGraph`, then `s.mu.Unlock()` not deferred: a panic in `fn` leaves the session locked forever. `closeSession`, `getOrOpen`, `Close`; lock order is `reg.mu` → `s.mu` (`getOrOpen` closes other sessions while holding `reg.mu`) |
| Session tests | `core/catalogsession/session_test.go` `TestDo` | No panic case |
| Writes on unwind | `core/writes/run.go` `Run` | Already `defer c.EndWrite()` and `defer tx.Rollback()`: a panicking write rolls back. `AfterCommit` work (graph memo) can be half-applied |
| Goroutines | `api/`, `core/` (non-test) | None spawned, so a recover in `Call` covers every handler path |
| Wire mapping | `macos/App/Platform/GoStore.swift` `mapSubjectPosition` | Copies `gridX` / `gridY` unchanged |
| Grid mapping | `macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift` `gridCell(gridPoint:)` | `Int64(floor(x / spacing))` traps on a non-finite or out-of-range value |
| Placement | `macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift` `composerSlot` (`placed.map(\.gridX).max()! + horizontalStep`, `column + horizontalStep`), `rowInColumn` (`max()! + verticalStep`), `maxGridX` / `maxGridY` (`Int64(floor(raw + 1e-9))`) | Plain `+` and `Int64(_:)` trap on extreme cells |
| Moves | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `moveSubject` (`fromGridX + deltaX`), `moveSubjects` (`move.fromX + move.deltaX`), `snapAndPatch` (through `gridCell`) | Plain `+` |
| Placement tests | `macos/ProvenenciaTests/EvidenceGraphPlacementTests.swift`, `GraphCanvasGridMappingTests.swift` | No extreme cases |
| Mirrored limits | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` rows "region point cap", "thumbnail decode edge" | The precedent for a Swift constant held to Go's |
| Raster budget | `core/raster/raster.go` `maxDecodePixels` (unexported, 64,000,000) | Used by `decode` only |
| Thumbnail policy | `macos/App/Platform/ThumbnailCache.swift` `ThumbnailDecodePolicy.accepts` | PNG/JPEG list plus a per-edge check (`maxDeclaredEdge` 4096) inline |
| Thumbnail tests | `macos/ProvenenciaTests/ThumbnailCacheTests.swift` `decodePolicyCases`, `refusesAnOversizedDeclaredEdge` | — |
| Viewer load | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` `load` | `ArtifactViewerKind.classify(mediaType:)` from Go's type, then `NSImage(contentsOf:)`: full decode through every ImageIO decoder, no type or size check |
| Viewer consumers | `ArtifactViewer.swift` (viewport), `ArtifactRegionCrop.snapshot` (`viewer.displayImage` → `cgImage(from:)`), `CitationTranscriptionModel.imageRaster` / `runAutoTranscribe` (OCR crops `displayImage`), `needsWholePageWarning` (`OCRImage.isOversized(image.size)`), `CitationComposerEvidenceColumn.swift` `artifactDetail` + private `pixelSize(_:)` (reads `rep.pixelsWide`) and `artifactThumbnail` | All read `displayImage` as if it were the original |
| Viewport | `ArtifactMediaViewport.swift` `layoutDocumentCentered`, `ArtifactMediaDocumentView.draw` | The document is `image.size` in points; zoom 0.5…4 |
| PDF cut | `ArtifactRegionCrop.swift` `Snapshot.pdf(url:pageIndex:)`, `snapshot` (`viewer.pdfDocument?.documentURL`), `pageBox(url:…)` (`PDFDocument(url:)`) | Reopens the bare URL with no object check |
| Object check | `macos/App/Platform/ProjectFiles.swift` `objectURL(projectDir:relPath:)` | Not actor-isolated; refuses `..`, symlinks, aliases, non-regular files |
| Viewer / crop tests | `ArtifactViewerModelTests.swift` (`loadsPNGImageViaProjectFiles`, `symlinkedObjectSetsMissingFile`), `ArtifactRegionCropTests.swift` (`imageCropIsTheBoundingBox`, `pdfCutDrawsOnlyTheBox`, `pdfSnapshotNeedsTheLoadedDocument`) | — |

**Precedent.** S12-23 ([#608](https://github.com/mendahu/provenencia/pull/608)) did the same shape for opening files: Go decides the type, Swift checks only what it must, and one shared check replaced per-caller copies. Copy its test style in `ProjectFilesTests` / `ThumbnailCacheTests` (fixtures written to a temp project folder through `ImageIO`). The "thumbnail decode edge" row in `TestSwiftVocabularyMatchesGo` is the precedent for mirroring a Go limit in Swift; add rows the same way.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The grid range is ±1,000,000 cells on both axes: `subjectpositions.MaxCell = 1_000_000` *(new)*. Swift mirrors it as `GraphCanvasGridMapping.cellLimit` *(new)*, held equal by a new row in `TestSwiftVocabularyMatchesGo` | The audit's range. 40 M points of document is far past any board a researcher grows (800 pt per growth step), and the sum of two in-range cells cannot overflow | ±100,000 — no evidence it is needed, and it moves the refusal closer to real boards |
| D-2 | Go refuses a write outside the range with `subjectpositions.invalid` in `SetTx` (so `Set`, subject create with placement and the bridge midpoint all get it) and clamps every read in `scanPosition` | One check where the table is written, one where it is read; every reader goes through `scanPosition` | Clamping on write — hides a caller bug; a migration rewriting rows — the read clamp covers them without a format step |
| D-3 | Swift clamps at the wire in `GoStore.mapSubjectPosition`, and every cell computed in placement and moves goes through `GraphCanvasGridMapping.clampedCell(_:)` *(new)* and `offsetCell(_:by:)` *(new, saturating add then clamp)*. `gridCell(gridPoint:)` returns clamped cells; a NaN coordinate maps to cell 0 and ±infinity to ±`cellLimit` | Swift never sends a cell Go refuses (AC-1.5), and no `+` or `Int64(_:)` in these paths can trap whatever the snapshot holds | Swift trusting Go's clamp alone — FakeStore, previews and the composer's own arithmetic still reach the traps |
| D-4 | `ffi.Call` keeps its signature and recovers: the current `switch` moves to `route(method, in)` *(new, unexported)*; `Call` runs it through `recovered(func() ([]byte, error))` *(new, unexported)*, which turns a panic into `apperr.New(apperr.CodeInternalUnknown, apperr.KindInternal)` and logs `provenencia: panic in method <n>: <value>` with `log.Printf`, as `core/writes/run.go` logs | One recover covers every handler; `recovered` is testable without a panicking production handler; the log keeps the cause | A recover in each handler — 90+ copies; a recover only in `provenencia_call` — untestable from Go tests and misses direct `Call` users |
| D-5 | `catalogsession.Do` defers its unlock. When `fn` or `warmGraph` panics, the deferred function marks the session closed and closes its catalog while still holding `s.mu`, unlocks, removes the session from `reg.sessions` only if the entry is still this session (taking `reg.mu` after `s.mu` is released), and lets the panic continue | A panic inside `AfterCommit` can leave the graph memo half-updated; reopening rebuilds the warm caches from the catalog. Waiters already `continue` on `s.closed`, and releasing `s.mu` before taking `reg.mu` keeps the `reg.mu` → `s.mu` order | Defer the unlock only — the next call runs on caches that may disagree with the catalog; recovering inside `Do` — the handler would return success-shaped values |
| D-6 | `raster.maxDecodePixels` is exported as `raster.MaxDecodePixels` (same value). The viewer decodes at most that many pixels (`ArtifactImageDecode.maxDecodedPixels` *(new)*, mirrored by a `TestSwiftVocabularyMatchesGo` row) | One bound for "how much of an image the app holds decoded", Go and Swift | A separate Swift-only number — drifts from Go's budget |
| D-7 | The viewer refuses an image that declares more than 256,000,000 pixels (`ArtifactImageDecode.maxDeclaredPixels` *(new)*), or an edge outside 1 to 65,535. User, 2026-10-29: keep the 198 MP (11,000 × 18,000) scan viewable | A crafted small file can declare any size; 256 MP is the largest transient decode ImageIO may make before it scales down | 64 MP — the user's real scan would stop opening |
| D-8 | The viewer decodes **once**, at no more than `maxDecodedPixels`, and that one raster serves display, crop and OCR. User, 2026-10-29, replacing the map's "screen-sized decode" | A screen-sized display would blur zoomed reading and OCR on every scan over about 4096 px; images within 64 MP keep today's pixels exactly | Screen-sized display plus a second bounded decode for crop and OCR — softer zoom on ordinary census scans, two decode paths |
| D-9 | A refused image (type mismatch, over budget, unreadable header) sets the existing `ArtifactViewerEmptyReason.loadFailed`; no new reason or copy | The researcher cannot act on the distinction; the existing state already says the file could not be shown | A new reason and string — new copy with no new action |
| D-10 | The sniffed type check: `CGImageSourceGetType` must conform to `UTType(mimeType: source.mediaType)`; a media type with no `UTType` is refused. Lives in `ArtifactImageDecode`, not in `ThumbnailDecodePolicy` | Working rule "What the app hands to macOS to open is checked first": the viewer checks only type match and budget, with no list of its own | A Swift list of viewer types — the rule forbids a second allowlist |
| D-11 | `ThumbnailDecodePolicy.declaredPixelSize(_:maxEdge:maxPixels:)` *(new)* reads the first image's declared width and height (header only) and returns them when both are in `1...maxEdge` and `width * height <= maxPixels`, else nil. `accepts` keeps its PNG/JPEG list and calls it with `maxDeclaredEdge` and `maxDeclaredEdge * maxDeclaredEdge` | The working rule names this one shared function in `ThumbnailDecodePolicy` | A second header reader in the viewer |
| D-12 | The decode is `CGImageSourceCreateThumbnailAtIndex` with `kCGImageSourceCreateThumbnailFromImageAlways`, `kCGImageSourceShouldCacheImmediately`, and `kCGImageSourceThumbnailMaxPixelSize` = the declared long edge when `w × h ≤ maxDecodedPixels`, else `floor(longEdge × sqrt(maxDecodedPixels / (w × h)))`. `kCGImageSourceCreateThumbnailWithTransform` is set to whichever value makes the result match `NSImage(contentsOf:)`'s orientation on the EXIF-6 fixture (T-11) | Saved regions are normalized to the image as the viewer shows it today; changing orientation would move every saved region on a rotated JPEG | Applying EXIF always — moves existing regions if today's viewer does not |
| D-13 | `displayImage` is `NSImage(cgImage: raster, size: pointSize)`, where `pointSize` is the declared pixel size × 72 / DPI per axis (`kCGImagePropertyDPIWidth` / `DPIHeight`, 72 when absent), swapped for orientations 5–8 when D-12 applies the transform. The model also holds `displayRaster: CGImage?` *(new)* and `imagePixelSize: (width: Int, height: Int)?` *(new, the declared size)* and `source: ArtifactViewerSource?` *(new)* | Layout, zoom, the overlay and OCR's `isOversized` read `size` and stay as today (AC-2.1); crop and OCR read the decoded pixels without an `NSImage` conversion; the composer's detail shows the original's size (AC-2.4) | Letting `size` become the reduced pixel size — zoom 1 would show a smaller document than today |
| D-14 | `ArtifactRegionCrop.Snapshot.pdf` becomes `.pdf(projectDir: String, relPath: String, pageIndex: Int)`; `pageBox` resolves it with `ProjectFiles.objectURL` and returns nil when that refuses. `.image` takes `viewer.displayRaster` | The map's design; one predicate (`snapshot`) still says whether the viewer can cut, changed in place for S14-03 | Carrying a URL checked at snapshot time — the cut runs later off the main actor |
| D-15 | `CitationTranscriptionModel.runAutoTranscribe` reads `artifactViewer.displayRaster` instead of converting `displayImage`; `imageRaster` (an `NSImage`) stays for `canAutoTranscribe` and `needsWholePageWarning` | OCR crops the same pixels the depiction crop does | Leaving the `NSImage` → `CGImage` conversion — its pixel size on a `size`-scaled `NSImage` is not pinned by any test |

**Rules this plan bends.** None. D-8 replaces the map scope entry's "shows a screen-sized decode and crops from a bounded one" by the user's answer (2026-10-29); STEP-8 rewords the scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `subjectpositions.MaxCell` *(new)* | const | `locator.MaxRegionPoints` | Same shape of limit; it belongs with the table it bounds |
| `route`, `recovered` in `api/ffi/dispatch.go` *(new, unexported)* | func | `Call` | `Call` keeps its contract; the split makes the recover testable |
| `raster.MaxDecodePixels` | const (export) | `raster.maxDecodePixels` | Rename of the existing constant, no second value |
| `GraphCanvasGridMapping.cellLimit`, `clampedCell(_:)`, `offsetCell(_:by:)` *(new)* | const / func | `GraphCanvasGridMapping.gridCell(gridPoint:)` | Extends the one grid-math enum both placement and the model already call (searched `addingReportingOverflow`, `clamp` in `macos/App`: no grid helper exists) |
| `ThumbnailDecodePolicy.declaredPixelSize(_:maxEdge:maxPixels:)` *(new)* | func | `ThumbnailDecodePolicy.accepts` inline edge loop | Extracted from `accepts`, which then calls it |
| `ArtifactImageDecode` *(new enum, bottom of `ArtifactViewerModel.swift`)* with `maxDeclaredPixels`, `maxDecodedPixels`, `decode(url:mediaType:) -> Decoded?` and `struct Decoded { raster: CGImage; pointSize: CGSize; pixelSize: (width: Int, height: Int) }` | enum / func | `ThumbnailCache` detached decode block | The thumbnail cache decodes to 512 px with its own type list; the viewer needs full size up to the budget and Go's type. Kept in the existing file (like `ArtifactFindNote`) so no `project.pbxproj` entry is needed before S14-15 |
| `ArtifactViewerModel.displayRaster`, `imagePixelSize`, `source` *(new)* | property | `displayImage` | Consumers need the decoded pixels, the original size and the object path; `displayImage` stays the view's `NSImage` |
| Test fixture helpers in the new tests | helper | `ArtifactRegionCropTests` private `makeRGBImage`; `ThumbnailCacheTests` fixtures | Reuse those helpers in their files; a PNG header-only fixture (IHDR with a huge size) is written inline in `ArtifactViewerModelTests` |

**Builds on.** S12-23's single-type-authority rule; `ThumbnailDecodePolicy`; `ProjectFiles.objectURL`; `TestSwiftVocabularyMatchesGo` mirrored-limit rows; `apperr.CodeInternalUnknown`.
**Deletes.** The private `pixelSize(_:)` in `CitationComposerEvidenceColumn.swift` (its one caller reads `imagePixelSize`). The inline edge loop in `ThumbnailDecodePolicy.accepts`. `Snapshot.pdf(url:pageIndex:)` and the `documentURL` read in `snapshot`. `NSImage(contentsOf:)` in `ArtifactViewerModel.load`.
**Next adds.** S14-03 gates Draw on `ArtifactRegionCrop.snapshot` as left here. S14-23 reads `raster.MaxDecodePixels` by its new name.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An image refused before decode | `ArtifactViewer` empty state (`.loadFailed`), `ArtifactRegionCrop.snapshot` → nil, `CitationTranscriptionModel.canAutoTranscribe` → false | `load` already resets on each call (`unloadPreservingGeneration`); `source`, `displayRaster`, `imagePixelSize` reset there and in `unload` |
| A PDF object replaced after load | `pageBox` → nil → `render` → nil | The composer's existing nil-render path (`CitationComposerModel+Depiction.swift` near `ArtifactRegionCrop.render`) is unchanged; S14-03 adds its failed-cut state |
| A session dropped after a panic | `catalogsession.Do` reopen | `getOrOpen` opens a missing entry already (`Close then Do reopens` case) |
| A clamped position | Evidence graph canvas | Cards at ±1,000,000 draw off any board a researcher grows; the canvas math stays finite (T-6) |

### Scale

At the map's Scale (60 graphs, 37,000 Subjects): the read clamp is two comparisons per `subject_positions` row in `scanPosition`; no new query, key or invalidation. The viewer does one decode per Artifact load, as today, now bounded at 64 MP (≈ 256 MiB) resident and 256 MP transient. A session reopen after a panic costs one `openResearcher` (reconcile and warm) and happens only after a panic.

### Edges

- **Assumptions.** `connect.go` adds two read positions: safe because both are clamped (|sum| ≤ 2,000,001), T-1. `Set`'s "unchanged placement returns no changes" compares the clamped previous value: a stored out-of-range row that Swift writes back as the clamped value is a no-op and stays out of range in the file; reads keep clamping it (T-1 case). Waiters in `Do` loop on `s.closed` until the dropped entry leaves the registry: bounded by one `reg.mu` section (T-3).
- **Neighbors.** `±1000000` saves and reads unchanged; `±1000001` is refused on write and read back as `±1000000` (T-1). An image of exactly 64,000,000 px decodes at full size; one pixel more is reduced (T-9). 256,000,000 declared px opens; one more is refused (T-9).
- **Removed inputs.** A refused image contributes nothing to OCR or crop: `canAutoTranscribe` is false and `snapshot` nil (T-10).
- **Real data.** The 198 MP scan is under 256 MP and over 64 MP: it opens reduced to 64 MP (manual check 3).
- **Untrusted size.** Every dimension read from the file is checked before work: declared edges `1...65535` and product ≤ 256 MP (`declaredPixelSize`, before `CGImageSourceCreateThumbnailAtIndex`); positions on every read. A PDF's page count and box are bounded by the existing `maxPageScale` / `maxLongEdge` render.

### Schema

None — no catalog change; out-of-range rows are clamped on read, not rewritten.

### Go core

| Package | Change |
| --- | --- |
| `core/database/subjectpositions` | `const MaxCell int64 = 1_000_000` *(new)*. `SetTx` returns `ErrInvalid` when `gridX` or `gridY` is outside `[-MaxCell, MaxCell]`, before the subject lookup. `scanPosition` clamps both fields into the range. Package doc names the range |
| `core/catalogsession` | `Do`: `defer` the unlock; on panic, mark closed, close the catalog, unlock, drop the registry entry if it is still this session, re-panic (D-5) |
| `core/raster` | `maxDecodePixels` → `MaxDecodePixels` (all uses in the package) |

Errors: `subjectpositions.invalid` → a write outside the range. `internal.unknown` → a recovered panic.

### FFI

`api/ffi/dispatch.go`: `func Call(method int32, in []byte) ([]byte, error) { return recovered(method, func() ([]byte, error) { return route(method, in) }) }` — `recovered` takes the method for the log line. No proto change.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Grid limit | `GraphCanvasGridMapping` | `macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift` | `cellLimit`, `clampedCell`, `offsetCell`; `gridCell` clamps |
| Wire clamp | `GoStore.mapSubjectPosition` | `macos/App/Platform/GoStore.swift` | `clampedCell` on both fields |
| Placement | `EvidenceGraphPlacement` | `macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift` | `+ horizontalStep` / `+ verticalStep` → `offsetCell`; `Int64(floor(…))` / `Int64(ceil(…))` in `maxGridX`, `maxGridY` go through one private `cell(_ raw: CGFloat) -> Int64` that clamps finite values and maps NaN to 0 |
| Moves | `EvidenceGraphModel.moveSubject`, `moveSubjects` | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` | `offsetCell` |
| Header check | `ThumbnailDecodePolicy` | `macos/App/Platform/ThumbnailCache.swift` | `declaredPixelSize` |
| Viewer decode | `ArtifactImageDecode`, `ArtifactViewerModel.load` | `macos/App/Features/ArtifactViewer/ArtifactViewerModel.swift` | `.image` case calls `ArtifactImageDecode.decode`; sets `displayRaster`, `displayImage`, `imagePixelSize`, `source` |
| Cut | `ArtifactRegionCrop` | `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` | D-14 |
| Composer detail | `CitationComposerEvidenceColumn.artifactDetail` | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` | Reads `model.artifactViewer.imagePixelSize` |
| OCR | `CitationTranscriptionModel.runAutoTranscribe` | `macos/App/Features/CitationComposer/CitationTranscriptionModel.swift` | D-15 |

States: refused → `.loadFailed`. Accessibility: unchanged.

### Strings

None — no new or changed copy.

## 7. Steps

**STEP-1 — Bound positions in Go**
Files: `core/database/subjectpositions/subjectpositions.go`, `subjectpositions_test.go`, `api/ffi/handlers/subjects_test.go`. Skill: `add-catalog-query`.
Do: D-1, D-2. Add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectpositions/... ./core/database/connect/... ./api/ffi/handlers/...` passes.

**STEP-2 — Recover panics at the FFI boundary; release and drop the session**
Files: `api/ffi/dispatch.go`, `api/ffi/dispatch_test.go`, `core/catalogsession/session.go`, `core/catalogsession/session_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-4, D-5. Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ ./core/catalogsession/...` passes; `go test -race ./core/catalogsession/...` passes.

**STEP-3 — Saturating grid cells in Swift**
Files: `GraphCanvasGridMapping.swift`, `GoStore.swift`, `EvidenceGraphPlacement.swift`, `EvidenceGraphModel.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/GraphCanvasGridMappingTests.swift`, `EvidenceGraphPlacementTests.swift`, `EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Do: D-3. Add the "grid cell limit" row (`static let cellLimit: Int64 = (\d[\d_]*)` in `GraphCanvasGridMapping`, want `subjectpositions.MaxCell`; strip `_` before comparing). Add T-5, T-6, T-7.
Check: `go test ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `xcodebuild test` for the three test files passes.

**STEP-4 — One declared-size check**
Files: `core/raster/raster.go` (export), `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`.
Do: D-6 (export only), D-11. Add T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster/... ./core/derivatives/...` passes; `ThumbnailCacheTests` pass unchanged plus T-8.

**STEP-5 — Bounded viewer decode**
Files: `ArtifactViewerModel.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`.
Do: D-6 (Swift constant and the "viewer decoded pixels" row: `static let maxDecodedPixels = (\d[\d_]*)` in `ArtifactImageDecode`, want `raster.MaxDecodePixels`), D-7, D-9, D-10, D-12, D-13. Reset the new properties in `unload`, `unloadPreservingGeneration`, and set them in `installImageRasterForTesting` (raster via `ArtifactRegionCrop.cgImage(from:)`, pixel size from it). Add T-9, T-11.
Check: `ArtifactViewerModelTests` pass, including `loadsPNGImageViaProjectFiles` unchanged.

**STEP-6 — Consumers read the bounded raster**
Files: `CitationComposerEvidenceColumn.swift`, `CitationTranscriptionModel.swift`, `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift`.
Do: D-13 (detail), D-15. Delete the private `pixelSize(_:)`. Add T-10.
Check: `xcodebuild test` passes; `rg "pixelSize\(" macos/App/Features/CitationComposer` has no hit.

**STEP-7 — The cut reopens through ProjectFiles**
Files: `ArtifactRegionCrop.swift`, `macos/ProvenenciaTests/ArtifactRegionCropTests.swift`.
Do: D-14. Update `pdfSnapshotNeedsTheLoadedDocument` to the new case shape. Add T-12.
Check: `ArtifactRegionCropTests` pass; `rg "documentURL" macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` has no hit.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row status; S14-13 scope entry: replace "it shows a screen-sized decode and crops from a bounded one" with "it decodes once at no more than 64 MP for display, crop and OCR, and refuses an image declaring over 256 MP (user, 2026-10-29)"), this plan (status, Landed note), `.agents/skills/use-catalog-session/SKILL.md` (one line: a panicking `Do` callback drops the session; the next call reopens), `docs/deployment-plan/spike-14/app-health-audit.md` (A-26 – A-28 rows: fixed in S14-13).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-13-crafted-project-safe.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/subjectpositions/subjectpositions_test.go` `TestSubjectPositions` (new cases) | `Set(±MaxCell)` → saved; `Set(MaxCell+1, 0)` → `ErrInvalid`, row unchanged; raw `UPDATE` to `(math.MaxInt64, math.MinInt64)` → `Get`, `GetTx`, `ListByEvidenceGraph` return `(MaxCell, -MaxCell)`; then `Set(MaxCell, -MaxCell)` → no changes | AC-1.1, AC-1.2, AC-1.3, D-2 |
| T-2 | Go | `api/ffi/handlers/subjects_test.go` `TestSubjectPositions` (new cases) | RPC `SetSubjectPosition` with `grid_x = 1000001` → code `subjectpositions.invalid`; list after a raw extreme row → clamped | AC-1.1, AC-1.2 |
| T-3 | Go | `core/catalogsession/session_test.go` `TestDo` case "panic drops the session" *(new)* | `Do(dir, panics)` under `recover` in the test → panic observed; second `Do(dir, ok)` returns nil, within 5 s, on a `*database.Catalog` pointer different from the first | AC-3.2, D-5 |
| T-4 | Go | `api/ffi/dispatch_test.go` `TestRecovered` *(new)* | `recovered(1, func() { panic("x") })` → `apperr.From(err).Code() == "internal.unknown"`; a non-panicking func's result and error pass through unchanged | AC-3.1, D-4 |
| T-5 | Swift | `GraphCanvasGridMappingTests` `clampsExtremeCells` *(new)* | `offsetCell(Int64.max, by: 1)` → `cellLimit`; `offsetCell(Int64.min, by: -1)` → `-cellLimit`; `gridCell(gridPoint: (.infinity, .nan))` → `(cellLimit, 0)` | AC-1.4, D-3 |
| T-6 | Swift | `EvidenceGraphPlacementTests` `extremePositionsDoNotTrap` *(new)* | Snapshot with subjects at `(Int64.max, Int64.min)` and `(Int64.min, Int64.max)` → `composerSlot`, `canvas(fitting:)`, `maxCell(on:)` return; every cell within ±`cellLimit` | AC-1.4 |
| T-7 | Swift | `EvidenceGraphModelTests` `arrowAtTheGridEdgeStays` *(new)* | Card at `gridX = cellLimit`, arrow right → patched cell `cellLimit`; `toast` stays nil (no `positionPersistFailedTitle`) | AC-1.5 |
| T-8 | Swift | `ThumbnailCacheTests` `declaredPixelSizeBounds` *(new, parameterized)* | 10 × 10 PNG, `maxEdge 4096` → `(10, 10)`; 4097 × 1 → nil; 100 × 100 with `maxPixels 9999` → nil; with `maxPixels 10000` → `(100, 100)` | D-11 |
| T-9 | Swift | `ArtifactViewerModelTests` `refusesAMismatchedType`, `refusesAnOverBudgetDeclaration`, `reducesAnImageOverTheDecodeBudget`, `keepsAnImageWithinTheBudgetAtFullSize` *(new)* | AC-2.2 fixture → `.loadFailed`; PNG IHDR 20000 × 20000 → `.loadFailed`; a JPEG that declares 72 MP (a header-patched small JPEG, or a 12,000 × 6,000 solid JPEG written by ImageIO) → `displayRaster` ≤ 64,000,000 px, aspect within 1 px, `imagePixelSize == (12000, 6000)`; 3000 × 2000 @ 300 dpi PNG → `displayRaster` 3000 × 2000, `displayImage.size` == `NSImage(contentsOf:)!.size` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-7, D-8, D-13 |
| T-10 | Swift | `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift` `autoTranscribeUsesTheDecodedRaster` *(new)*; `refusedImageCannotTranscribe` *(new)* | Loaded 3000 × 2000 image, region top half → the fake OCR engine receives a 3000 × 1000 image; refused image → `canAutoTranscribe == false` | AC-2.1, D-15 |
| T-11 | Swift | `ArtifactViewerModelTests` `orientsAsNSImageDoes` *(new)* | 300 × 200 JPEG written with `kCGImagePropertyOrientation: 6` → `displayImage.size` and the raster's aspect equal `NSImage(contentsOf:)`'s | AC-2.5, D-12 |
| T-12 | Swift | `ArtifactRegionCropTests` `pdfSnapshotCarriesTheObjectPath`, `pdfCutRefusesASwappedObject` *(new)* | Loaded PDF → `.pdf(projectDir, relPath, 0)`; replace the object with a symlink to a copy → `cut` nil; unchanged object → `pdfCutDrawsOnlyTheBox` still passes | AC-4.1, AC-4.2, AC-4.3, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ArtifactRegionCropTests` (snapshot shape), `ArtifactViewerModelTests` (`installImageRasterForTesting` callers), `TestSwiftVocabularyMatchesGo`, `EvidenceGraphPlacementTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `maxDecodePixels`, `Snapshot.pdf(url:`, and `pixelSize(` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no stale hit; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive (no brief)
- [ ] Version bump: patch — VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status and the reworded scope entry (STEP-8); completed.md: one short paragraph of what shipped
- [ ] app-health-audit.md: A-26, A-27, A-28 marked fixed in S14-13
- [ ] use-catalog-session skill: the panic line (STEP-8)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. In a copy of a dogfood project, run `sqlite3 <project>/provenencia.sqlite "UPDATE subject_positions SET grid_x = 9223372036854775807, grid_y = -9223372036854775808 WHERE rowid = (SELECT min(rowid) FROM subject_positions);"`. Open that Source's Evidence graph: it opens; the card is far off the board; the other cards are where they were. Drag another card: it moves and saves.
2. Open a Citation composer on an image Artifact of an ordinary scan: it looks, zooms to 400% and auto-transcribes a region as before; the detail line shows the scan's pixel size.
3. Open the 11,000 × 18,000 scan: it shows; the detail reads "11000 × 18000"; a likeness crop saves.
4. In a project copy, overwrite one image object with a PNG whose IHDR declares 20000 × 20000 (T-9's fixture written to disk): the viewer shows the could-not-load state and the app stays responsive.
5. Open a PDF Citation and save a region: it saves as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `NSImage(cgImage:size:)` changes how the viewport draws a reduced raster | AC-2.1 / T-9 pin `size`; manual check 2 at 400% |
| Dropping a session on panic races a waiter | T-3 under `-race`; lock order `reg.mu` → `s.mu` kept |
| `UTType(mimeType:)` does not map a type Go stores (for example `image/tif`, `image/heif`) | Stop condition below |

**Stop and report if:**

- `NSImage(contentsOf:)` and `CGImageSourceCreateThumbnailAtIndex` with either `kCGImageSourceCreateThumbnailWithTransform` value disagree on the EXIF-6 fixture's orientation or size (D-12 cannot hold).
- A fixture of a type on Go's image allowlist (`core/ingest/mediatypes/mediatypes.go` `allowed`: jpeg, png, gif, webp, tiff, bmp, heic, heif) written by ImageIO fails D-10's conformance check against Go's media type string.
- `cgImage(forProposedRect:)` or the viewport needs the `NSImage`'s pixel size to equal its `size` anywhere outside §4's consumers (`rg displayImage macos/App` shows a reader not in §4).
- A position read path exists that does not go through `scanPosition` (`rg subject_positions core api` shows SQL outside `core/database/subjectpositions`, `core/database/effects/registry.go` and `catalogmodel`).
- `catalogsession.Do`'s drop needs `reg.mu` while `s.mu` is held.
- A test or existing behavior needs a new empty-state reason or string (D-9).

## 12. Open questions

None.

## 13. PR

**Title:** S14-13 — A crafted project can't crash the app
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-13-crafted-project-safe.md`.
