# S14-13 — A crafted project can't crash the app

**Status:** Landed in #632 (2026-10-29)
**Landed:** Differences from the plan: re-saving a clamped position rolls back (`writes.Run` rolls back a write with no changes), so the raw column keeps the file's value and every read clamps it, rather than being rewritten (T-1 asserts the raw value unchanged). `declaredPNGHeaderData` carries filler IDAT data (one byte per 3,000 declared pixels) because ImageIO reads no size from an IHDR-then-IEND header. The `Do` close-and-forget sits in a `runLocked` helper. `installImageRasterForTesting` (now `installImageForTesting`) installs nothing when it has no raster. Manual check 1 to 5 not run in the app by the implementer. Round 1 changes, each User, 2026-10-29, ship-log round 1: D-10 is superseded (a `.heif` Artifact sniffs as `public.heic`): Go's stored-type table now gives each image type its accepted type identifiers, `SourceFileRef.type_identifiers` carries them, and the viewer checks the sniffed type conforms to one of them, with no Swift alias map and no media-type or extension fallback. D-9 is superseded for the size case: an image refused for its declared size shows `ArtifactViewerEmptyReason.imageTooLarge` ("Image too large to show"); every other refusal keeps the load-failed copy. D-11's home moves: the shared header read and bounded decode sit in `ImageHeaderDecode` (`Platform/`), not `ThumbnailDecodePolicy`. D-6 and D-7 are refined: both constants stay as 8-bit pixel counts and are divided by the bytes per color component, so a 16-bit image is refused above 128 MP and decoded at no more than 32 MP.
**Spike / slice:** Spike 14 · Slice 5 — Audit fixes
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a refused image uses the viewer's existing load-failed state; no control, copy or layout changes)
**Board:** —
**Depends on:** S12-23 (landed [#608](https://github.com/mendahu/provenencia/pull/608))
**Blocks:** S14-03
**Branch:** `s14-13-crafted-project-safe`
**Version bump:** patch, 0.2.1 → 0.2.2 — crash and memory fixes on crafted input, no feature ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/catalog-session.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/versioning.mdc`; `macos/App/Features/GraphCanvas/README.md` (no catalog types under `macos/App/Features/GraphCanvas/`)
**Context:** audit [A-26](../app-health-audit.md#a-26--crafted-graph-positions-crash-the-app), [A-27](../app-health-audit.md#a-27--the-artifact-viewer-decodes-originals-unbounded), [A-28](../app-health-audit.md#a-28--a-go-panic-kills-the-app); [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) (the PDF cut's file check); working rules "What the app hands to macOS to open is checked first, and Go decides the type", "Audit items keep their IDs"; map review K-8, K-25; bar outcome 15
**Planned:** 2026-10-29 against `main` @ `b8ca4cfc`
**Review:** [`S14-13-crafted-project-safe.review.md`](S14-13-crafted-project-safe.review.md) — Fits with changes — Ready with changes, 2 rounds; all findings fixed

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
| AC-1.4 | The Swift placement and move code is fed `Int64.max` / `Int64.min` cells (a FakeStore row the Go clamp did not see) | No trap; every cell it returns is within ±1,000,000 |
| AC-1.5 | A card at `gridX = 1000000` is moved right by the arrow key | The patched cell stays `1000000`; no store error is shown |

**US-2.** As a researcher who opens an image Artifact, the viewer decodes it within a bound, and an image that lies about its type or size shows the load-failed state instead of exhausting memory.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A PNG of 3000 × 2000 px at 300 dpi, object `….png`, `media_type` `image/png` | Shows as today: `displayImage.size` equals `NSImage(contentsOf:)`'s size (720 × 480 pt); the composer's detail reads "3000 × 2000"; OCR of the top-half region receives a 3000 × 1000 image |
| AC-2.2 | PNG bytes stored as an object named `….jpg` | `emptyReason == .loadFailed`, `displayImage == nil`, no decode ran |
| AC-2.3 | A PNG header declaring 20,000 × 20,000 (400 MP) | `.loadFailed` before any pixel decode |
| AC-2.4 | A grayscale JPEG of 12,000 × 6,000 (72 MP, within the 256 MP budget) | Displays; `displayRaster` has at most 64,000,000 pixels (about 11,313 × 5,656) and keeps the aspect within 1 px; `displayImage.size` is the full image's point size; the detail reads "12000 × 6000" |
| AC-2.5 | A JPEG with EXIF orientation 6 | Displays with the same orientation and `size` as `NSImage(contentsOf:)` gives for that file |

**US-3.** As the engine, a panic in a handler returns `internal.unknown` and the project stays usable.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | A function run through `ffi.Call`'s recovery panics | `Call` returns `nil` and an error whose code is `internal.unknown`, kind internal; the process lives |
| AC-3.2 | A `catalogsession.Do` callback panics | The panic reaches the caller; the session's mutex is released; the next `Do` on the same project succeeds on a freshly opened catalog |

**US-4.** As the engine, the PDF cut opens its file through `ProjectFiles.objectURL`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A loaded PDF Artifact; `ArtifactRegionCrop.snapshot` | Carries the viewer's `ArtifactViewerSource` and the 0-based page, not a URL |
| AC-4.2 | The object file is replaced by a symlink after load; the region is cut | `cut` returns nil |
| AC-4.3 | The object is unchanged; a region is cut | The same box as today (`pdfCutDrawsOnlyTheBox` keeps passing) |

## 3. Scope

| In | Covered by |
| --- | --- |
| A-26: `subjectpositions` refuses outside ±`MaxCell` on write and clamps on read | STEP-1, AC-1.1, AC-1.2, AC-1.3 |
| A-26: Swift uses saturating cell arithmetic in placement and moves and clamps to the Evidence graph's limit | STEP-3, AC-1.4, AC-1.5 |
| A-28: `ffi.Call` recovers a panic into `internal.unknown` | STEP-2, AC-3.1 |
| A-28: `catalogsession.Do` releases its lock with `defer` and drops a session whose callback panicked | STEP-2, AC-3.2 |
| A-27: one shared header read and one shared bounded decode in `ThumbnailDecodePolicy`, used by the thumbnail cache and the viewer | STEP-4, STEP-5 |
| A-27: the viewer checks sniffed type and the 256 MP budget before decoding, and decodes once at no more than 64 MP for display, crop and OCR | STEP-5, STEP-6, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| The PDF cut reopens its file through `ProjectFiles.objectURL` | STEP-7, AC-4.1, AC-4.2, AC-4.3 |

| Out | Home |
| --- | --- |
| Length caps on catalog text (audit L-10) | not planned (map scope entry) |
| The foreign-catalog open order (audit L-7) | not planned (map scope entry) |
| Draw blocked on a file the viewer cannot cut; the failed-cut state | S14-03 |
| Bounding PDFKit's own parse and render of a crafted PDF | not planned: the cut is already capped by `maxPageScale` and `maxLongEdge`, and PDFKit renders per page |
| A `recover` in `api/libprovenencia/main.go` around `EncodeError` | not planned: `ffi.Call` covers every handler; `EncodeError` marshals a fixed message |
| Capping the graph document size or the scroll view for a card at ±1,000,000 cells (a 40 M-point board) | not planned: only a crafted row reaches it, and the canvas math stays finite (T-6) |
| FakeStore's connect midpoint (`FakeStore.swift` `floorDiv(fromPos.gridX + toPos.gridX + 1, 2)`) | not planned: FakeStore positions are test-seeded |
| OCR's oversize warning reading points (`needsWholePageWarning` → `OCRImage.isOversized(image.size)`) rather than pixels | not planned: unchanged behavior; [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) if dogfood finds it |
| A clamp in `GoStore.mapSubjectPosition` | not planned (D-3): Go clamps every read |
| The JPEG reduced decoder and raster's byte limit | S14-23 |
| A new empty-state reason or copy for a refused image | not planned (D-9) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Position write | `core/database/subjectpositions/subjectpositions.go` `Set`, `SetTx` | Accepts any `int64`. `Set` always runs `SetTx`'s upsert, then compares the previous row to decide whether to return a change. Callers: `api/ffi/handlers/subjects.go` (`SetSubjectPosition`, `CreateSubject` placement through `core/database/subjects/subjects.go` `Create`), `core/database/connect/connect.go` (bridge midpoint) |
| Position read | same file `scanPosition` (used by `Get`, `GetTx`, `ListByEvidenceGraph`) | Returns stored values unchanged. `connect.go` adds two read values (`fromPos.GridX+toPos.GridX+1`), which wraps in Go on extreme rows |
| Position tests | `core/database/subjectpositions/subjectpositions_test.go` `TestSubjectPositions`; `api/ffi/handlers/subjects_test.go` `TestSubjectPositions` | No bound cases |
| FFI dispatch | `api/ffi/dispatch.go` `Call` | One `switch` over methods; no `recover`. `api/libprovenencia/main.go` `provenencia_call` calls it |
| Dispatch tests | `api/ffi/dispatch_test.go` `TestCallRouter` | Unknown-method cases only |
| FFI guidance | `.agents/skills/add-ffi-handler/SKILL.md` ("`dispatch.go` is **only** a method switch", "one `case`" in `Call`), `.cursor/rules/ffi-handlers.mdc` ("a thin Call switch", "dispatch.go only routes") | Become false when the switch moves to `route` |
| Session | `core/catalogsession/session.go` `Do` | `s.mu.Lock()`, `fn`, `warmGraph`, then `s.mu.Unlock()` not deferred: a panic in `fn` leaves the session locked forever. `closeSession` takes `s.mu`, sets `closed`, closes and nils `cat`; `Close`, `getOrOpen` and `CloseAll` delete the entry under `reg.mu` and call `closeSession`. Lock order is `reg.mu` → `s.mu` |
| Session tests | `core/catalogsession/session_test.go` `TestDo` | No panic case |
| Writes on unwind | `core/writes/run.go` `Run` | Already `defer c.EndWrite()` and `defer tx.Rollback()`: a panicking write rolls back. `AfterCommit` work (graph memo) can be half-applied. Logs with `log.Printf("provenencia: …")` |
| Goroutines | `api/`, `core/` (non-test) | None spawned, so a recover in `Call` covers every handler path |
| Grid mapping | `macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift` `gridCell(gridPoint:)` | `Int64(floor(x / spacing))` traps on a non-finite or out-of-range value. The module README keeps catalog types out of `macos/App/Features/GraphCanvas/` |
| Placement | `macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift` `composerSlot` (`placed.map(\.gridX).max()! + horizontalStep`, `column + horizontalStep`, `minCell.gridY + 2`), `rowInColumn` (`max()! + verticalStep`), `maxGridX` / `maxGridY` (`Int64(floor(raw + 1e-9))`), `minGridX` / `minGridY` (`Int64(ceil(…))` of constants) | Plain `+` and `Int64(_:)` trap on extreme cells. Owns `minCell` / `maxCell` |
| Moves | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `moveSubject` (`fromGridX + deltaX`), `moveSubjects` (`move.fromX + move.deltaX`), `snapAndPatch`, `updateHover`, `beginCreate` (through `gridCell`) | Plain `+` |
| Placement tests | `macos/ProvenenciaTests/EvidenceGraphPlacementTests.swift`, `GraphCanvasGridMappingTests.swift`, `EvidenceGraphModelTests.swift` | No extreme cases |
| Mirrored limits | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` rows "region point cap", "thumbnail decode edge" (`(\d+)` regexes) | The precedent for a Swift constant held to Go's |
| Raster budget | `core/raster/raster.go` `maxDecodePixels` (unexported, 64,000,000; doc comments at lines 41, 46, 102, 155); `core/raster/raster_test.go` (lines 139–140) | — |
| Old name in docs | `docs/deployment-plan/spike-14/oversized-jpeg-thumbnails.md`, `docs/deployment-plan/spike-14/deployment-plan.md` (working rule "Oversized JPEGs…", S14-23 scope entry) | Name `maxDecodePixels` |
| Thumbnail policy | `macos/App/Platform/ThumbnailCache.swift` `ThumbnailDecodePolicy.accepts`; the detached block in `ThumbnailCache.image(projectDir:relPath:)` | `accepts`: PNG/JPEG list plus an inline per-edge check (`maxDeclaredEdge` 4096). The block opens the source, calls `accepts`, then `CGImageSourceCreateThumbnailAtIndex` (`FromImageAlways`, `WithTransform`, `ShouldCacheImmediately`, `MaxPixelSize: 512`) |
| Thumbnail tests | `macos/ProvenenciaTests/ThumbnailCacheTests.swift` `decodePolicyCases`, `loadsAGoJPEG`, `refusesAnOversizedDeclaredEdge` | — |
| Viewer content | `macos/App/Features/ArtifactViewer/ArtifactViewerContent.swift` `ArtifactViewerContent.image(NSImage)` (`@unchecked Sendable`, `==` compares `tiffRepresentation`) | — |
| Viewer load | `ArtifactViewerModel.swift` `load` | `ArtifactViewerKind.classify(mediaType:)` from Go's type, then `NSImage(contentsOf:)` in a `Task.detached` returning `Result<ArtifactViewerContent, Error>`: full decode through every ImageIO decoder, no type or size check. Stores the image twice: `content = .image(image)` and `displayImage = image`, reset in `unload`, `unloadPreservingGeneration` and set in `installImageRasterForTesting` |
| Viewer consumers | `ArtifactViewer.swift` (viewport, `displayImage?.size`), `ArtifactRegionCrop.snapshot` (`viewer.displayImage` → `cgImage(from:)`), `CitationTranscriptionModel.imageRaster` / `runAutoTranscribe` (OCR crops `displayImage`), `needsWholePageWarning`, `CitationComposerModel.imageRaster`, `CitationComposerEvidenceColumn.swift` `artifactDetail` + private `pixelSize(_:)` (reads `rep.pixelsWide`) and `artifactThumbnail` | All read `displayImage` as if it were the original |
| Viewport | `ArtifactMediaViewport.swift` `layoutDocumentCentered`, `ArtifactMediaDocumentView.draw` | The document is `image.size` in points; zoom 0.5…4 |
| PDF cut | `ArtifactRegionCrop.swift` `Snapshot.pdf(url:pageIndex:)`, `snapshot` (`viewer.pdfDocument?.documentURL`), `pageBox(url:…)` (`PDFDocument(url:)`) | Reopens the bare URL with no object check |
| Object check | `macos/App/Platform/ProjectFiles.swift` `objectURL(projectDir:relPath:)` | Not actor-isolated; refuses `..`, symlinks, aliases, non-regular files |
| Viewer source | `ArtifactViewerKind.swift` `ArtifactViewerSource` (`projectDir`, `relPath`, `mediaType`, `Sendable`) | Passed to `load`, not kept |
| Viewer / crop / OCR tests | `ArtifactViewerModelTests.swift` (`loadsPNGImageViaProjectFiles`, `symlinkedObjectSetsMissingFile`, private `minimalPNGData()`), `ArtifactRegionCropTests.swift` (`imageCropIsTheBoundingBox`, `pdfCutDrawsOnlyTheBox`, `pdfSnapshotNeedsTheLoadedDocument`), `CitationAutoTranscribeTests.swift` (private `FakeOCREngine` with `lastSize`), `CitationComposerDepictionTests.swift` | `installImageRasterForTesting` is called from `CitationComposerDepictionTests` and `CitationAutoTranscribeTests` |
| Xcode project | `macos/Provenencia.xcodeproj/project.pbxproj` | Hand-listed files (S14-15 not landed): `ArtifactViewerKind.swift` has its four entries |

**Precedent.** S12-23 ([#608](https://github.com/mendahu/provenencia/pull/608)) did the same shape for opening files: Go decides the type, Swift checks only what it must, and one shared check replaced per-caller copies. Copy its test style in `ProjectFilesTests` / `ThumbnailCacheTests` (fixtures written to a temp project folder through `ImageIO`). The "thumbnail decode edge" row in `TestSwiftVocabularyMatchesGo` is the precedent for mirroring a Go limit in Swift; add rows the same way.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The grid range is ±1,000,000 cells on both axes: `subjectpositions.MaxCell = 1_000_000` *(new)*. Swift mirrors it as `EvidenceGraphPlacement.cellLimit: Int64 = 1000000` *(new, literal without `_`)*, held equal by a new row in `TestSwiftVocabularyMatchesGo` (`static let cellLimit: Int64 = (\d+)` in enum `EvidenceGraphPlacement`) | The audit's range. 40 M points of document is far past any board a researcher grows (800 pt per growth step), and the sum of two in-range cells cannot overflow. The limit belongs to the Evidence graph's table, so it lives with `minCell` / `maxCell`, not in `GraphCanvas/` | ±100,000 — no evidence it is needed; the limit on `GraphCanvasGridMapping` — breaks the module's no-catalog-types contract |
| D-2 | Go refuses a write outside the range with `subjectpositions.invalid` in `SetTx`, before the subject lookup (so `Set`, subject create with placement and the bridge midpoint all get it), and clamps every read in `scanPosition` | One check where the table is written, one where it is read; every reader goes through `scanPosition` | Clamping on write — hides a caller bug; a migration rewriting rows — the read clamp covers them without a format step |
| D-3 | `GraphCanvasGridMapping` gains generic saturating helpers with no catalog limit: `offsetCell(_:by:) -> Int64` *(new)* adds with `addingReportingOverflow` and saturates at `Int64.min` / `Int64.max`; `saturatingFloor(_ value: CGFloat) -> Int64` *(new)* floors a value, maps NaN to 0 and a floor outside `Int64` (±infinity included) to `Int64.min` / `Int64.max`; `gridCell(gridPoint:)` calls it for x and y. `EvidenceGraphPlacement.clampedCell(_:) -> Int64` *(new)* clamps to ±`cellLimit`. Every cell that placement and `EvidenceGraphModel` compute is `clampedCell(offsetCell(…))` or `clampedCell(gridCell(…))`. `GoStore.mapSubjectPosition` is unchanged | Swift never sends a cell Go refuses (AC-1.5); no `+` or `Int64(_:)` in these paths traps whatever the snapshot holds; `GraphCanvas/` stays free of the Evidence graph bound | A clamp in `GoStore` — never fires, since Go clamps every read; trusting Go alone — FakeStore and the composer's own arithmetic still reach the traps |
| D-4 | `ffi.Call` keeps its signature. The current `switch` moves verbatim to `route(method int32, in []byte) ([]byte, error)` *(new, unexported)*. `Call` returns `recovered(method, func() ([]byte, error) { return route(method, in) })`. `recovered(method int32, fn func() ([]byte, error)) (out []byte, err error)` *(new, unexported)* defers a `recover()`; on a panic it sets `out = nil`, `err = apperr.New(apperr.CodeInternalUnknown, apperr.KindInternal)` and logs `log.Printf("provenencia: panic in method %d: %v\n%s", method, r, debug.Stack())`. New methods are added as `case`s in `route` | One recover covers every handler; `recovered` is testable without a panicking production handler; the log keeps the cause and the stack, as `core/writes/run.go` logs | A recover in each handler — 90+ copies; a recover only in `provenencia_call` — untestable from Go tests and misses direct `Call` users |
| D-5 | `catalogsession` gets one definition of closing: `closeSession` splits into `closeLocked(s *session) error` *(new; caller holds `s.mu`; sets `closed`, closes and nils `cat`)* and `closeSession` (locks, calls `closeLocked`). `forget(dir string, s *session)` *(new)* takes `reg.mu` and deletes `reg.sessions[dir]` only when it is `s`. `Do` sets `done := false` before `fn`, sets `done = true` after `fn` and `warmGraph` return, and defers: if `!done`, `_ = closeLocked(s)` (its error is dropped: the panic is the error the caller sees); then `s.mu.Unlock()`; then, if `!done`, `forget(dir, s)`. No `recover` in `Do`: the panic continues to `ffi.Call` with its stack | A panic inside `AfterCommit` can leave the graph memo half-updated; reopening rebuilds the warm caches. Waiters already `continue` on `s.closed`; `forget` runs after `s.mu` is released, so the `reg.mu` → `s.mu` order holds. `Close`, `getOrOpen` and `CloseAll` keep calling `closeSession` | Defer the unlock only — the next call runs on caches that may disagree with the catalog; a third inline close body — a later `session` field is missed on the panic path; recovering inside `Do` — the handler would return success-shaped values |
| D-6 | `raster.maxDecodePixels` is exported as `raster.MaxDecodePixels` (same value). The viewer decodes at most that many pixels (`ArtifactImageDecode.maxDecodePixels = 64000000` *(new)*, mirrored by a `TestSwiftVocabularyMatchesGo` row, `(\d+)`) | One bound for "how much of an image the app holds decoded", Go and Swift | A separate Swift-only number — drifts from Go's budget |
| D-7 | The viewer refuses an image that declares more than 256,000,000 pixels (`ArtifactImageDecode.maxDeclaredPixels = 256000000` *(new)*), or an edge outside 1 to 65,535 (`ArtifactImageDecode.maxEdge = 65535` *(new)*). User, 2026-10-29: keep the 198 MP (11,000 × 18,000) scan viewable | A crafted small file can declare any size; 256 MP is the largest transient decode ImageIO may make before it scales down | 64 MP — the user's real scan would stop opening |
| D-8 | The viewer decodes **once**, at no more than `maxDecodePixels`, and that one raster serves display, crop and OCR. User, 2026-10-29, replacing the map's "screen-sized decode" | A screen-sized display would blur zoomed reading and OCR on every scan over about 4096 px; images within 64 MP keep today's pixels exactly | Screen-sized display plus a second bounded decode for crop and OCR — softer zoom on ordinary census scans, two decode paths |
| D-9 | A refused image (type mismatch, over budget, unreadable header, failed decode) sets the existing `ArtifactViewerEmptyReason.loadFailed`; no new reason or copy | The researcher cannot act on the distinction; the existing state already says the file could not be shown | A new reason and string — new copy with no new action |
| D-10 | The sniffed type check: the expected type is `UTType(mimeType: source.mediaType)`, or, when macOS has none for that string, `UTType(filenameExtension:)` of the object path's extension (`(relPath as NSString).pathExtension`). `CGImageSourceGetType` must conform to it; neither available → refused. `source.mediaType` still picks the viewer kind (`ArtifactViewerKind.classify`). User, 2026-10-29 | Go assigns every object's extension from the same stored-type table as its media type (`core/ingest/mediatypes/mediatypes.go` `allowed`, `StoredExtension`), so the fallback is still Go's type: `image/tif` → `.tiff`. The viewer keeps no list of its own | Media type only — every `image/tif` Artifact would be refused; extension only — departs from the media type the rule names; a Swift alias map — a second table of Go's types |
| D-11 | `ThumbnailDecodePolicy` owns the two ImageIO steps both callers use. `struct DeclaredImage { width: Int; height: Int; dpiWidth: CGFloat; dpiHeight: CGFloat; orientation: Int }` *(new)*. `declaredImage(_ source: CGImageSource, maxEdge: Int, maxPixels: Int) -> DeclaredImage?` *(new)* makes one `CGImageSourceCopyPropertiesAtIndex(source, 0, nil)` call and returns nil unless both edges are in `1...maxEdge` and `width * height <= maxPixels`; DPI is taken only when finite and in `1...100000`, else 72 (absent, 0, negative or absurd values from the file); orientation defaults to 1 when absent or outside `1...8`. `decode(_ source: CGImageSource, maxPixelSize: Int, applyTransform: Bool) -> CGImage?` *(new)* calls `CGImageSourceCreateThumbnailAtIndex` with `FromImageAlways`, `ShouldCacheImmediately`, `MaxPixelSize` and `WithTransform: applyTransform`. `accepts` keeps its PNG/JPEG list and calls `declaredImage(source, maxEdge: maxDeclaredEdge, maxPixels: maxDeclaredEdge * maxDeclaredEdge)`. `ThumbnailCache`'s block calls `decode(src, maxPixelSize: maxPixelSize, applyTransform: true)` | The working rule names one shared function in `ThumbnailDecodePolicy`; one header read and one decode mean a later byte limit or drawn-size option reaches both callers ([idea](../../../ideas/depictions-and-likenesses.md) "decode at the drawn size") | A second header read and decode in the viewer — the two drift on options |
| D-12 | The viewer's `maxPixelSize` is the declared long edge when `w × h ≤ maxDecodePixels`, else `floor(longEdge × sqrt(maxDecodePixels / (w × h)))` (orientation-free: the long edge and the product do not change under rotation). `applyTransform` is the constant `ArtifactImageDecode.appliesOrientation` *(new)*, set to the value that makes T-11 pass: the result matches `NSImage(contentsOf:)`'s orientation and size on the EXIF-6 fixture. If that value is `true`, delete the constant and the `applyTransform:` parameter instead (`decode` always applies the transform, as the thumbnail cache does), so no parameter has a single value | Saved regions are normalized to the image as the viewer shows it today; changing orientation would move every saved region on a rotated JPEG | Applying EXIF always — moves existing regions if today's viewer does not |
| D-13 | `ArtifactImageDecode.Decoded` *(new, `@unchecked Sendable` like `ArtifactViewerContent`)* holds `raster: CGImage`, `image: NSImage` and `pixelSize: (width: Int, height: Int)`. `image` is `NSImage(cgImage: raster, size: pointSize)` with `pointSize = (width × 72 / dpiWidth, height × 72 / dpiHeight)`. `pixelSize` is the declared `(width, height)`. When `appliesOrientation` is true and the orientation is 5–8, both `pointSize` and `pixelSize` swap width and height, so they describe the image as shown | Layout, zoom, the overlay and OCR's `isOversized` read `size` and stay as today (AC-2.1); crop and OCR read the decoded pixels; the composer's detail shows the full size the way it is displayed (AC-2.4) | Letting `size` become the reduced pixel size — zoom 1 would show a smaller document than today |
| D-14 | `ArtifactViewerContent.image` becomes `.image(ArtifactImageDecode.Decoded)`; its `==` compares `raster` by identity (`===`). `ArtifactViewerModel` stores no image separately: `displayImage` (`NSImage?`), `displayRaster` (`CGImage?`, *new*) and `imagePixelSize` (*new*) are computed from `content`. `source: ArtifactViewerSource?` *(new, `private(set)`)* is set in `load` right after `ProjectFiles.objectURL` succeeds, for every renderable kind, and cleared wherever `content` is cleared. The detached task returns `Result<ArtifactViewerContent, Error>` as today. `installImageRasterForTesting(_ image: NSImage, raster: CGImage? = nil)` builds a `Decoded` with `image` as given and `raster` as given or else `ArtifactRegionCrop.cgImage(from: image)` (pixel size from the raster); when neither gives a raster, `content` stays nil. `CitationAutoTranscribeTests`' whole-page case installs `NSImage(size: 6200 × 8400)` with a 1 × 1 `raster:`, so `needsWholePageWarning` still reads the 6200 × 8400 size | One write sets all of it and one `content = nil` clears it; no window where the raster and the `NSImage` disagree | Stored `displayRaster` / `imagePixelSize` beside `displayImage` — five parallel facts, three reset sites each |
| D-15 | `ArtifactRegionCrop.Snapshot.pdf` becomes `.pdf(source: ArtifactViewerSource, pageIndex: Int)`; `snapshot` returns `.pdf` only when `viewer.pdfDocument != nil` and `viewer.source != nil` (nil while the PDF is loading, as today); `pageBox` resolves `ProjectFiles.objectURL(projectDir: source.projectDir, relPath: source.relPath)` and returns nil when that refuses. `.image` takes `viewer.displayRaster` | The map's design; one predicate (`snapshot`) still says whether the viewer can cut, changed in place for S14-03 | Carrying a URL checked at snapshot time — the cut runs later off the main actor |
| D-16 | `CitationTranscriptionModel.runAutoTranscribe` reads `artifactViewer.displayRaster` instead of converting `displayImage`; `imageRaster` (an `NSImage`) stays, because `CitationComposerModel.imageRaster` and `needsWholePageWarning` read it | OCR crops the same pixels the depiction crop does | Removing `imageRaster` — two callers outside this PR's change |
| D-17 | `ArtifactImageDecode` *(new)* is its own file, `macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift`. If `project.pbxproj` at the base still lists files by hand (no `PBXFileSystemSynchronizedRootGroup`), add its PBXFileReference, PBXBuildFile, ArtifactViewer group child and app Sources entry the way `ArtifactViewerKind.swift` is listed | One-type files are the folder's pattern (`ArtifactViewerKind.swift`, `ArtifactViewerContent.swift`) | Tucking it into `ArtifactViewerModel.swift` to avoid the project edit |

**Rules this plan bends.** D-10 bends working rule "What the app hands to macOS to open is checked first, and Go decides the type" ("the sniffed type matches the media type Go returned"): when macOS has no type for Go's media type string, the check uses the extension Go assigned from the same table. User, 2026-10-29 (option: media type, else extension). STEP-8 rewords the rule and the S14-13 scope entry. D-8 replaces the map scope entry's "shows a screen-sized decode and crops from a bounded one" by the user's answer (2026-10-29); STEP-8 rewords the scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `subjectpositions.MaxCell` *(new)* | const | `locator.MaxRegionPoints` | Same shape of limit; it belongs with the table it bounds |
| `route`, `recovered` in `api/ffi/dispatch.go` *(new, unexported)* | func | `Call` | `Call` keeps its contract; the split makes the recover testable |
| `closeLocked`, `forget` in `core/catalogsession/session.go` *(new)* | func | `closeSession`, `Close` | Split out of them, so the panic path reuses the one close body |
| `raster.MaxDecodePixels` | const (export) | `raster.maxDecodePixels` | Rename of the existing constant, no second value |
| `GraphCanvasGridMapping.offsetCell(_:by:)`, `saturatingFloor(_:)` *(new)* | func | `GraphCanvasGridMapping.gridCell(gridPoint:)` | Generic grid math beside it; no catalog limit (README contract). Searched `addingReportingOverflow`, `clamp` in `macos/App`: no grid helper exists |
| `EvidenceGraphPlacement.cellLimit`, `clampedCell(_:)` *(new)* | const / func | `EvidenceGraphPlacement.minCell`, `maxCell` | The Evidence graph's other bounds live here |
| `ThumbnailDecodePolicy.DeclaredImage`, `declaredImage(_:maxEdge:maxPixels:)`, `decode(_:maxPixelSize:applyTransform:)` *(new)* | type / func | `ThumbnailDecodePolicy.accepts` inline edge loop; `ThumbnailCache` detached decode | Extracted from those two; both callers use them |
| `ArtifactImageDecode` *(new file)*: `maxEdge`, `maxDeclaredPixels`, `maxDecodePixels`, `appliesOrientation`, `decode(url: URL, relPath: String) -> Decoded?`, `struct Decoded` | enum / func / type | `ThumbnailDecodePolicy` | Holds only the viewer's policy (type match, budgets, point size) on top of the shared steps |
| `ArtifactViewerModel.displayRaster`, `imagePixelSize` (computed), `source` *(new)* | property | `displayImage` | `displayImage` becomes computed from the same `content` |
| Test fixtures (`macos/ProvenenciaTests/ArtifactFileFixtures.swift` `ArtifactFileFixtures`) | helper | `ArtifactFileFixtures.writeImage(projectDir:relPath:width:height:fileType:topHalfRed:)`, `write(_:projectDir:relPath:mode:)` | Extended, not copied (`add-swift-test`: "do not write another copy"): `writeImage` gains `dpi: CGFloat? = nil`, `orientation: Int? = nil` and `grayscale: Bool = false`, writing through `CGImageDestination` with properties (`NSBitmapImageRep.representation` cannot set EXIF orientation); `writeDeclaredPNGHeader(projectDir:relPath:width:height:)` *(new)* writes the 8-byte signature, an IHDR with the given size (big-endian, depth 8, color type 2) and its CRC-32 over type and data (a private `crc32(_:)` beside it), then IEND, so ImageIO reads the declared size with no pixels; `declaredPNGHeaderData(width:height:)` *(new)* returns the same bytes for T-8's in-memory source. `FakeOCREngine.lastSize` is reused as is |

**Builds on.** S12-23's single-type-authority rule; `ThumbnailDecodePolicy`; `ProjectFiles.objectURL`; `ArtifactViewerSource`; `TestSwiftVocabularyMatchesGo` mirrored-limit rows; `apperr.CodeInternalUnknown`.
**Deletes.** The private `pixelSize(_:)` in `CitationComposerEvidenceColumn.swift` (its one caller reads `imagePixelSize`). The inline edge loop in `ThumbnailDecodePolicy.accepts` and the inline `CGImageSourceCreateThumbnailAtIndex` options in `ThumbnailCache`. The stored `displayImage` property (now computed). `Snapshot.pdf(url:pageIndex:)` and the `documentURL` read in `snapshot`. `NSImage(contentsOf:)` in `ArtifactViewerModel.load`. The close body inside `closeSession` (moves to `closeLocked`). `ArtifactRegionCrop.cgImage(from:)` loses its two production callers and stays for `installImageRasterForTesting` and its own tests.
**Next adds.** S14-03 gates Draw on `ArtifactRegionCrop.snapshot` as left here. S14-23 reads `raster.MaxDecodePixels` by its new name. A drawn-size thumbnail decode is a `maxPixelSize` argument to `ThumbnailDecodePolicy.decode`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An image refused before decode | `ArtifactViewer` empty state (`.loadFailed`), `ArtifactRegionCrop.snapshot` → nil, `CitationTranscriptionModel.canAutoTranscribe` → false | `load` resets `content` and `source` on each call (`unloadPreservingGeneration`); the computed properties follow |
| A PDF object replaced after load | `pageBox` → nil → `render` → nil | The composer's existing nil-render path (`CitationComposerModel+Depiction.swift` at the `ArtifactRegionCrop.render` call) is unchanged; S14-03 adds its failed-cut state |
| A session dropped after a panic | `catalogsession.Do` reopen | `getOrOpen` opens a missing entry already (`Close then Do reopens` case) |
| A clamped position | Evidence graph canvas | Cards at ±1,000,000 draw far off any board a researcher grows; the canvas math stays finite (T-6) |

### Scale

At the map's Scale (60 graphs, 37,000 Subjects): the read clamp is two comparisons per `subject_positions` row in `scanPosition`; no new query, key or invalidation. The viewer does one decode per Artifact load, as today, now bounded at 64 MP (≈ 256 MiB) resident and 256 MP transient. The thumbnail cache's decode is unchanged in cost. A session reopen after a panic costs one `openResearcher` (reconcile and warm) and happens only after a panic.

### Edges

- **Assumptions.** `connect.go` adds two read positions: safe because both are clamped (|sum| ≤ 2,000,001), T-1. `Set` always runs `SetTx`'s upsert: a stored out-of-range row that Swift writes back as the clamped value is rewritten in the file to the clamped value and returns no change (the table is unaudited, `effects` registry `None`), T-1 checks the raw column afterwards. Waiters in `Do` spin: between the unlock and `forget`, `getOrOpen` still returns the closed session, they lock it, see `closed` and `continue`, until `forget`'s one `reg.mu` section drops the entry (T-3, `-race`). The defer is registered after the existing `s.closed || s.cat == nil` check and its inline unlock.
- **Neighbors.** `±1000000` saves and reads unchanged; `±1000001` is refused on write and read back as `±1000000` (T-1). An image of exactly 64,000,000 px decodes at full size; 8000 × 8001 (64,008,000 px) gets `maxPixelSize` 8000 and decodes to about 8000 × 7999 (T-9, through the pure helper `ArtifactImageDecode.maxPixelSize(width:height:)` *(new)*). A header declaring 16,000 × 16,000 (256,000,000 px) passes `declaredImage(maxPixels: maxDeclaredPixels)`; 16,000 × 16,001 does not (T-8, header-only fixtures).
- **Removed inputs.** A refused image contributes nothing to OCR or crop: `canAutoTranscribe` is false and `snapshot` nil (T-10).
- **Real data.** The 198 MP scan is under 256 MP and over 64 MP: it opens reduced to 64 MP (manual check 3).
- **Untrusted size.** Every dimension read from the file is checked before work: declared edges `1...65535` and product ≤ 256 MP (`declaredImage`, before `decode`); positions on every read. A PDF's page count and box are bounded by the existing `maxPageScale` / `maxLongEdge` render.

### Schema

None — no catalog change; out-of-range rows are clamped on read, and rewritten only when the app saves that card again.

### Go core

| Package | Change |
| --- | --- |
| `core/database/subjectpositions` | `const MaxCell int64 = 1_000_000` *(new)*. `SetTx` returns `ErrInvalid` when `gridX` or `gridY` is outside `[-MaxCell, MaxCell]`, before the subject lookup. `scanPosition` clamps both fields into the range. Package doc names the range |
| `core/catalogsession` | D-5 |
| `core/raster` | `maxDecodePixels` → `MaxDecodePixels` in `raster.go` (code and the doc comments at lines 41, 46, 102, 155) and `raster_test.go` (lines 139–140) |

Errors: `subjectpositions.invalid` → a write outside the range. `internal.unknown` → a recovered panic.

### FFI

`api/ffi/dispatch.go`: D-4. No proto change. Imports `log` and `runtime/debug`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Saturating cells | `GraphCanvasGridMapping` | `macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift` | `offsetCell`; `gridCell` saturates (D-3) |
| Evidence limit | `EvidenceGraphPlacement` | `macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift` | `cellLimit`, `clampedCell`. In `composerSlot`: `max()! + horizontalStep`, `column + horizontalStep`, `minCell.gridY + 2` → `clampedCell(offsetCell(…))`; in `rowInColumn`: `max()! + verticalStep` the same. `maxGridX` / `maxGridY` become `clampedCell(GraphCanvasGridMapping.saturatingFloor(raw + 1e-9))`. `minGridX` / `minGridY` and `horizontalStep` / `verticalStep` / `originReserveCells` stay: they are constants of fixed card sizes |
| Moves | `EvidenceGraphModel` | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` | `moveSubject`, `moveSubjects`: `clampedCell(offsetCell(…))`. `snapAndPatch`, `updateHover`, `beginCreate`: `clampedCell` of each `gridCell` result |
| Shared ImageIO steps | `ThumbnailDecodePolicy` | `macos/App/Platform/ThumbnailCache.swift` | D-11 |
| Viewer decode | `ArtifactImageDecode` *(new file)* | `macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift` *(new)* | `decode(url:relPath:)`: `CGImageSourceCreateWithURL` → type check (D-10) → `declaredImage(maxEdge: maxEdge, maxPixels: maxDeclaredPixels)` → `ThumbnailDecodePolicy.decode(maxPixelSize: maxPixelSize(width:height:), applyTransform: appliesOrientation)` → `Decoded` (D-13); nil at any refusal |
| Viewer model | `ArtifactViewerModel`, `ArtifactViewerContent` | `macos/App/Features/ArtifactViewer/` | D-14; the `.image` case of the detached task returns `.success(.image(decoded))` or `.failure(.decodeFailed)` |
| Cut | `ArtifactRegionCrop` | `macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` | D-15 |
| Composer detail | `CitationComposerEvidenceColumn.artifactDetail` | `macos/App/Features/CitationComposer/CitationComposerEvidenceColumn.swift` | Reads `model.artifactViewer.imagePixelSize` |
| OCR | `CitationTranscriptionModel.runAutoTranscribe` | `macos/App/Features/CitationComposer/CitationTranscriptionModel.swift` | D-16 |

States: refused → `.loadFailed`. Accessibility: unchanged.

### Strings

None — no new or changed copy.

## 7. Steps

**STEP-1 — Bound positions in Go**
Files: `core/database/subjectpositions/subjectpositions.go`, `subjectpositions_test.go`, `api/ffi/handlers/subjects_test.go`. Skill: `add-catalog-query`.
Do: D-1 (Go constant), D-2. Add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjectpositions/... ./core/database/connect/... ./api/ffi/handlers/...` passes.

**STEP-2 — Recover panics at the FFI boundary; release and drop the session**
Files: `api/ffi/dispatch.go`, `api/ffi/dispatch_test.go`, `core/catalogsession/session.go`, `core/catalogsession/session_test.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-4, D-5. Add T-3, T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ ./core/catalogsession/...` passes; `CGO_ENABLED=1 go test -race -tags fts5 ./core/catalogsession/...` passes.

**STEP-3 — Saturating grid cells in Swift**
Files: `GraphCanvasGridMapping.swift`, `EvidenceGraphPlacement.swift`, `EvidenceGraphModel.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/GraphCanvasGridMappingTests.swift`, `EvidenceGraphPlacementTests.swift`, `EvidenceGraphModelTests.swift`. Skill: `add-swift-test`.
Do: D-1 (Swift constant and vocabulary row "evidence grid cell limit": path constant `swiftEvidencePlacement = "../../macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift"` *(new)*, want `strconv.FormatInt(subjectpositions.MaxCell, 10)`), D-3. Then run `rg -n ' [+-] |Int64\(' macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift macos/App/Features/GraphCanvas/GraphCanvasGridMapping.swift` and `rg -n 'grid[XY]|from[XY]|Grid[XY]' macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift | rg ' [+-] '`: every hit that adds to or converts a cell is converted per §6 macOS; every other hit (point arithmetic, the constants §6 says stay) is listed as kept in the PR description. Add T-5, T-6, T-7.
Check: `go test ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `xcodebuild test` for the three test files passes.

**STEP-4 — Shared header read and decode**
Files: `core/raster/raster.go`, `core/raster/raster_test.go`, `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`, `.agents/skills/add-swift-test/SKILL.md`.
Do: D-6 (Go export only), D-11. Extend `ArtifactFileFixtures` per §6 Codebase fit and name the new options on the skill's `ArtifactFileFixtures` line. Add T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster/... ./core/derivatives/... ./api/ffi/...` passes; `ThumbnailCacheTests` pass (existing cases unchanged) plus T-8.

**STEP-5 — Bounded viewer decode**
Files: `macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift` *(new)*, `macos/Provenencia.xcodeproj/project.pbxproj` (D-17), `ArtifactViewerContent.swift`, `ArtifactViewerModel.swift`, `api/ffi/swift_vocabulary_test.go`, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`.
Do: D-6 (Swift constant and vocabulary row "viewer decode pixels": path constant `swiftArtifactImageDecode = "../../macos/App/Features/ArtifactViewer/ArtifactImageDecode.swift"` *(new)*, want `strconv.Itoa(raster.MaxDecodePixels)`), D-7, D-9, D-10, D-12, D-13, D-14, D-17. Add T-9, T-11.
Check: `ArtifactViewerModelTests` pass, including `loadsPNGImageViaProjectFiles` unchanged; `go test ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes.

**STEP-6 — Consumers read the bounded raster**
Files: `CitationComposerEvidenceColumn.swift`, `CitationTranscriptionModel.swift`, `macos/ProvenenciaTests/CitationAutoTranscribeTests.swift`.
Do: D-13 (detail line), D-16. Delete the private `pixelSize(_:)`. Change the whole-page case's `installImageRasterForTesting` call per D-14. Add T-10.
Check: `xcodebuild test` passes (`CitationAutoTranscribeTests`, `CitationComposerDepictionTests` included); `rg "pixelSize\(" macos/App/Features/CitationComposer` has no hit.

**STEP-7 — The cut reopens through ProjectFiles**
Files: `ArtifactRegionCrop.swift`, `macos/ProvenenciaTests/ArtifactRegionCropTests.swift`.
Do: D-15. Update `pdfSnapshotNeedsTheLoadedDocument` to the new case shape. Add T-12.
Check: `ArtifactRegionCropTests` pass; `rg "documentURL" macos/App/Features/ArtifactViewer/ArtifactRegionCrop.swift` has no hit.

**STEP-8 — Docs and guidance**
Files: `docs/deployment-plan/spike-14/completed.md`; `docs/deployment-plan/spike-14/deployment-plan.md` (map row status; S14-13 scope entry: replace "it shows a screen-sized decode and crops from a bounded one" with "it decodes once at no more than 64 MP for display, crop and OCR, and refuses an image declaring over 256 MP (user, 2026-10-29)"; `maxDecodePixels` → `raster.MaxDecodePixels` in the working rule "Oversized JPEGs get a thumbnail from a reduced decode of our own" and the S14-23 scope entry; in the working rule "What the app hands to macOS to open is checked first, and Go decides the type" and the S14-13 scope entry, "the sniffed type matches the media type Go returned" → "the sniffed type matches the media type Go returned, or, when macOS has no type for it, the extension Go assigned from the same table (user, 2026-10-29)"); `docs/deployment-plan/spike-14/oversized-jpeg-thumbnails.md` (`maxDecodePixels` → `MaxDecodePixels`); this plan (status, Landed note); `.agents/skills/use-catalog-session/SKILL.md` (one line: a panicking `Do` callback drops the session and the next call reopens); `.agents/skills/add-ffi-handler/SKILL.md` and `.cursor/rules/ffi-handlers.mdc` (dispatch is the `route` switch; add the `case` in `route`; `Call` recovers panics into `internal.unknown`, so handlers never `recover`); `docs/deployment-plan/spike-14/app-health-audit.md` (A-26 – A-28 rows: fixed in S14-13); `VERSION`, `core/version.go`, `MARKETING_VERSION` in `project.pbxproj` → 0.2.2.
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-13-crafted-project-safe.md` reports no errors; `rg -n 'maxDecodePixels' docs .agents .cursor core api -g '!docs/deployment-plan/spike-14/prs/S14-13-*' -g '!docs/deployment-plan/spike-14/app-health-audit*.md'` has no hit.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/subjectpositions/subjectpositions_test.go` `TestSubjectPositions` (new cases) | `Set(±MaxCell)` → saved; `Set(MaxCell+1, 0)` → `ErrInvalid`, row unchanged; raw `UPDATE` to `(math.MaxInt64, math.MinInt64)` → `Get`, `GetTx`, `ListByEvidenceGraph` return `(MaxCell, -MaxCell)`; then `Set(MaxCell, -MaxCell)` → no changes, and a raw `SELECT grid_x, grid_y` reads `(MaxCell, -MaxCell)` | AC-1.1, AC-1.2, AC-1.3, D-2 |
| T-2 | Go | `api/ffi/handlers/subjects_test.go` `TestSubjectPositions` (new cases) | RPC `SetSubjectPosition` with `grid_x = 1000001` → code `subjectpositions.invalid`; list after a raw extreme row → clamped | AC-1.1, AC-1.2 |
| T-3 | Go | `core/catalogsession/session_test.go` `TestDo` case "panic drops the session" *(new)* | `Do(dir, func(*database.Catalog) error { panic("x") })` inside a test `defer recover()` → panic observed; second `Do(dir, ok)` returns nil within 5 s on a `*database.Catalog` pointer different from the first; run under `-race` | AC-3.2, D-5 |
| T-4 | Go | `api/ffi/dispatch_test.go` `TestRecovered` *(new, table-driven)* | `recovered(1, func() ([]byte, error) { panic("x") })` → `nil`, `apperr.From(err).Code() == "internal.unknown"`, kind internal; `recovered(1, func() ([]byte, error) { return []byte{1}, nil })` → `[]byte{1}`, nil; an `errors.New("e")` result passes through unchanged | AC-3.1, D-4 |
| T-5 | Swift | `GraphCanvasGridMappingTests` `saturatesExtremeCells` *(new)* | `offsetCell(Int64.max, by: 1)` → `Int64.max`; `offsetCell(Int64.min, by: -1)` → `Int64.min`; `gridCell(gridPoint: (.infinity, .nan))` → `(Int64.max, 0)`; `gridCell(gridPoint: (-1e30, 40))` → `(Int64.min, 1)`; `saturatingFloor(.nan)` → 0, `saturatingFloor(2.5)` → 2, `saturatingFloor(-2.5)` → -3 | AC-1.4, D-3 |
| T-6 | Swift | `EvidenceGraphPlacementTests` `extremePositionsDoNotTrap` *(new)* | Snapshot with subjects at `(Int64.max, Int64.min)` and `(Int64.min, Int64.max)` → `composerSlot`, `canvas(fitting:)`, `maxCell(on:)` return; every cell within ±`cellLimit`; `clampedCell(Int64.max)` → `cellLimit` | AC-1.4 |
| T-7 | Swift | `EvidenceGraphModelTests` `arrowAtTheGridEdgeStays` *(new)* | Card at `gridX = cellLimit`, arrow right → patched cell `cellLimit`; `toast` stays nil (no `positionPersistFailedTitle`) | AC-1.5 |
| T-8 | Swift | `ThumbnailCacheTests` `declaredImageBounds` *(new, parameterized)* | 10 × 10 PNG, `maxEdge 4096, maxPixels 4096*4096` → width 10, height 10, dpi 72, orientation 1; `declaredPNGHeaderData(4097, 1)` → nil; 100 × 100 with `maxPixels 9999` → nil, with `10000` → non-nil; 300-dpi PNG → `abs(dpiWidth - 300) < 0.01`; a PNG written with `dpi: 0` → `dpiWidth == 72`; `declaredPNGHeaderData(16000, 16000)` with `maxEdge 65535, maxPixels 256000000` → non-nil, `(16000, 16001)` → nil | D-7, D-11 |
| T-9 | Swift | `ArtifactViewerModelTests` `refusesAMismatchedType`, `refusesAnOverBudgetDeclaration`, `reducesAnImageOverTheDecodeBudget`, `keepsAnImageWithinTheBudgetAtFullSize`, `maxPixelSizeAtTheBudget` *(new)* | `writeImage(fileType: .png)` at `objects/ab/cd/x.jpg` with `mediaType` `image/jpeg` → `.loadFailed`, `displayImage == nil`; `writeDeclaredPNGHeader(width: 20000, height: 20000)` → `.loadFailed`; `writeImage(width: 12000, height: 6000, fileType: .jpeg, grayscale: true)` → `displayRaster` ≤ 64,000,000 px, aspect within 1 px, `imagePixelSize == (12000, 6000)`; 3000 × 2000 PNG written with `dpi: 300` → `displayRaster` 3000 × 2000, `displayImage.size` within 0.01 pt of `NSImage(contentsOf:)!.size`; `maxPixelSize(width: 8000, height: 8000)` → 8000, `(8000, 8001)` → 8000 (`floor(8001 × sqrt(64e6 / 64.008e6))`; ImageIO scales the short edge, so the result is about 8000 × 7999) | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-7, D-8, D-10, D-13 |
| T-10 | Swift | `CitationAutoTranscribeTests` `autoTranscribeUsesTheDecodedRaster`, `refusedImageCannotTranscribe` *(new)* | `writeImage(width: 3000, height: 2000, fileType: .png)` loaded, region top half → `FakeOCREngine.lastSize == 3000 × 1000`; refused image → `canAutoTranscribe == false` | AC-2.1, D-16 |
| T-11 | Swift | `ArtifactViewerModelTests` `orientsAsNSImageDoes` *(new)* | `writeImage(width: 300, height: 200, fileType: .jpeg, orientation: 6)` → `displayImage.size` and the raster's aspect equal `NSImage(contentsOf:)`'s | AC-2.5, D-12 |
| T-12 | Swift | `ArtifactRegionCropTests` `pdfSnapshotCarriesTheSource`, `pdfCutRefusesASwappedObject` *(new)* | Loaded PDF → `.pdf(source: <the loaded source>, pageIndex: 0)`; replace the object with a symlink to a copy → `cut` nil; unchanged object → `pdfCutDrawsOnlyTheBox` still passes | AC-4.1, AC-4.2, AC-4.3, D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ArtifactRegionCropTests` (snapshot shape), `ArtifactViewerModelTests` and `CitationComposerDepictionTests` (`installImageRasterForTesting`), `ThumbnailCacheTests`, `TestSwiftVocabularyMatchesGo`, `EvidenceGraphPlacementTests`, `core/raster` tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg -n 'maxDecodePixels|Snapshot.pdf\(url|pixelSize\(' docs .agents .cursor core api macos/App -g '!docs/deployment-plan/spike-14/prs/S14-13-*' -g '!docs/deployment-plan/spike-14/app-health-audit*.md'` has no hit except `GraphCanvasItemView.swift` `pixelSize(scale:)` (unrelated); `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (D-10: the working rule and the S14-13 scope entry reworded in STEP-8)
- [ ] UI matches the brief inventory; no new kit primitive (no brief)
- [ ] Version bump: 0.2.2 in VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status, the reworded scope entry and the renamed constant (STEP-8); completed.md: one short paragraph of what shipped
- [ ] app-health-audit.md: A-26, A-27, A-28 marked fixed in S14-13
- [ ] use-catalog-session skill, add-ffi-handler skill and ffi-handlers rule updated (STEP-8)
- [ ] PR description: what shipped, AC evidence, the STEP-3 kept-hit list, deviations, manual check
```

## 10. Manual check

1. In a copy of a dogfood project, run `sqlite3 <project>/provenencia.sqlite "UPDATE subject_positions SET grid_x = 9223372036854775807, grid_y = -9223372036854775808 WHERE rowid = (SELECT min(rowid) FROM subject_positions);"`. Open that Source's Evidence graph: it opens; the card is far off the board; the other cards are where they were. Drag another card: it moves and saves.
2. Open a Citation composer on an image Artifact of an ordinary scan: it looks, zooms to 400% and auto-transcribes a region as before; the detail line shows the scan's pixel size.
3. Open the 11,000 × 18,000 scan: it shows; the detail reads "11000 × 18000"; a likeness crop saves.
4. In a project copy, overwrite one image object with a PNG whose IHDR declares 20000 × 20000 (the bytes of `ArtifactFileFixtures.declaredPNGHeaderData(width: 20000, height: 20000)` written over the object): the viewer shows the could-not-load state and the app stays responsive.
5. Open a PDF Citation and save a region: it saves as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| `NSImage(cgImage:size:)` changes how the viewport draws a reduced raster | AC-2.1 / T-9 pin `size`; manual check 2 at 400% |
| Dropping a session on panic races a waiter | T-3 under `-race`; lock order `reg.mu` → `s.mu` kept |
| HEIC sniffed against a `.heif` object | Stop condition below |

**Stop and report if:**

- Neither value of `appliesOrientation` makes T-11 pass (D-12 cannot hold).
- A fixture of a type on Go's image allowlist (`core/ingest/mediatypes/mediatypes.go` `allowed`) that ImageIO can write (jpeg, png, gif, tiff, bmp, heic) fails D-10's check (media type, else the extension Go's table gives it). `image/webp` is not writable by ImageIO and stays untested here.
- `rg -n 'displayImage' macos/App` shows a reader not listed in §4 "Viewer consumers".
- A position read path exists that does not go through `scanPosition` (`rg -n subject_positions core api` shows SQL outside `core/database/subjectpositions`, `core/database/effects/registry.go`, `core/database/catalogmodel` and `core/database/migrations`).
- `forget` would need `reg.mu` while `s.mu` is held.
- A test or existing behavior needs a new empty-state reason or string (D-9).

## 12. Open questions

None.

## 13. PR

**Title:** S14-13 — A crafted project can't crash the app
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-13-crafted-project-safe.md`.
