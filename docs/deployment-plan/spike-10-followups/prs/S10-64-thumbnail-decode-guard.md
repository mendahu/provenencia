# S10-64 — Check a thumbnail before decoding it

**Status:** Landed in [#469](https://github.com/mendahu/provenencia/pull/469) (2026-10-18)
**Spike / slice:** Spike 10 follow-ups · likeness hardening
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a refused image shows the missing state each reader already has)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s10-64-thumbnail-decode-guard`
**Version bump:** none — spike PRs leave `VERSION` at 0.2.0 and a release cuts the bump (`git log -- VERSION`: last moved in #402; S10-60 – S10-63 and Spike 11 do the same) ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [application stack, read path](../../../application-stack.md), [`core/raster`](../../../../core/raster/raster.go) (`MaxEdgeLimit`), audit 2026-10-18 finding SEC-2 second half (Medium item 5; likeness and depiction area audit, all five lenses)
**Planned:** 2026-10-18 against `main` @ `e97cf54`
**Review:** [`S10-64-thumbnail-decode-guard.review.md`](S10-64-thumbnail-decode-guard.review.md) — Fits with changes / Ready with changes, 1 round

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

Every thumbnail, likeness tile, portrait, and evidence-card face in the app decodes through `ThumbnailCache`. Today it hands whatever file sits at the path to ImageIO, which decodes any format it knows at whatever size the file declares, and it starts one decode per request with no limit. A project folder someone else made can put a TIFF, a HEIC, or a PNG that declares 20000×20000 pixels under `objects/`, and opening the Persons list or a Source graph decodes it on as many threads as there are faces. After this PR the cache decodes only a PNG or JPEG whose declared long edge is at most 4096 pixels — the largest edge Go's `raster` package ever writes — and runs at most four decodes at once. Anything else is stored as missing and shows the placeholder each reader already has.

**Why now.** The audit (SEC-2) found the Swift object read path had no format or size check while Go's own reader refuses the same inputs. S10-62 closes the path half (symlinks); this PR closes the decode half.

**What stays.** The 512-pixel thumbnail size, the 100-entry LRU, the in-flight coalescing, the `.missing` sentinel, every caller, and every placeholder. Go is not changed.

## 2. User stories

**US-1.** As a researcher who opens a project folder someone else prepared, when a stored object is not an image the app writes or declares a size no Go encoder produces, the app shows its usual placeholder instead of decoding it, so a crafted file cannot exhaust memory.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `objects/aa/bb/x.png` holds a PNG of 8×8 | `ThumbnailCache.image` returns an image (no regression) |
| AC-1.2 | `objects/aa/bb/x.jpg` holds a JPEG of 8×8 (the Go thumbnail derivative format) | returns an image (no regression) |
| AC-1.3 | `objects/aa/bb/x.png` holds TIFF bytes | returns nil and stores the key as missing |
| AC-1.4 | `objects/aa/bb/x.png` holds a PNG declaring 4097×8 | returns nil and stores the key as missing |
| AC-1.5 | a PNG declaring 4096×8 | returns an image (the limit is inclusive) |
| AC-1.6 | a file of random bytes | the policy refuses it (as today, the cache returns nil); covered at the policy layer by T-1 |

**US-2.** As a researcher opening a Source graph with many faces or scrolling a long Persons list, the app decodes at most four images at a time, so a burst of thumbnails never runs dozens of full decodes in parallel.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `maxConcurrentDecodes = 2`, eight distinct PNG keys requested at once | all eight return images; the recorded peak of running decodes is at least 1 and at most 2 |
| AC-2.2 | two requests for the same key at once | both get the image and the cache holds one entry (coalescing unchanged) |

**US-3.** As a contributor, when Go's largest encoder edge changes, a Go test fails until the Swift limit matches.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ThumbnailDecodePolicy.maxDeclaredEdge` and `raster.MaxEdgeLimit` differ | the "thumbnail decode edge" row of `TestSwiftVocabularyMatchesGo` fails |

## 3. Scope

| In | Covered by |
| --- | --- |
| A pure `ThumbnailDecodePolicy` that accepts a `CGImageSource` only when its type is PNG or JPEG and its declared pixel width and height are each at most `maxDeclaredEdge` (4096) | STEP-1, AC-1.1 – AC-1.6 |
| `ThumbnailCache.image` checks the policy before `CGImageSourceCreateThumbnailAtIndex`; a refusal returns nil and caches `.missing` | STEP-2, AC-1.3, AC-1.4 |
| A scoped decode slot on `ThumbnailCache` (`withDecodeSlot`, the Swift twin of Go's `raster.Slot`): at most `maxConcurrentDecodes` (4) decodes run; later ones wait in arrival order; the class doc comment names the policy and the cap | STEP-3, AC-2.1, AC-2.2 |
| `ArtifactFileFixtures.writeImage` as the one image writer (`writePNG` forwards to it) and `ArtifactFileFixtures.write` made internal for raw bytes, so tests write a JPEG, TIFF, PNG, or garbage bytes without a pasted literal | STEP-1, T-1 – T-5 |
| A row in `TestSwiftVocabularyMatchesGo` pinning `maxDeclaredEdge` to `raster.MaxEdgeLimit`; the mirrored-limits clause in `macos-client.mdc` names it | STEP-4, AC-3.1 |
| One sentence in `docs/application-stack.md` **Read path** | STEP-5 |

| Out | Home |
| --- | --- |
| Refusing a symlinked `objects/`, shard folder, or object | S10-62 |
| Cancelling a decode that is already running, decoding at the drawn size, bounding the cache by bytes, the evidence card keeping its own loaded image (audit PRF-2) | not planned (audit 2026-10-18 Medium items 6 and 13, not mapped) |
| A byte-based decode limit in Go `raster` (audit SEC-3) | not planned (audit 2026-10-18 Low item) |
| The Artifact viewer's full-resolution decode (`ArtifactViewerModel.load`) | not planned |
| New copy or a distinct "unreadable" glyph for a refused image | not planned (audit UX-11 covers the graph tile, not mapped) |
| Rewriting the existing `ThumbnailCacheTests` cases off their base64 literal | not planned |

## 4. Current state

Verified at `e97cf54`.

| Area | Where | Today |
| --- | --- | --- |
| Cache | `macos/App/Platform/ThumbnailCache.swift` `ThumbnailCache.image(projectDir:relPath:)` | `@MainActor` class. Coalesces in-flight keys. The unstructured `Task` resolves `ProjectFiles.objectURL`, then a `Task.detached(priority: .userInitiated)` calls `CGImageSourceCreateWithURL` and `CGImageSourceCreateThumbnailAtIndex` with `kCGImageSourceThumbnailMaxPixelSize: maxPixelSize` (512). No type or declared-size check; no limit on concurrent detached decodes. nil → `.missing` |
| Cache knobs | same file `maxPixelSize` (nonisolated static 512), `maxEntries` (var 100), `entryCount` (test helper) | The pattern this PR copies for `maxConcurrentDecodes` and `peakConcurrentDecodes` |
| Callers | `macos/App/Platform/CachedThumbnail.swift` `load()`; `macos/App/Features/EvidenceGraph/EvidenceCardView.swift` `loadMissingLikenesses()` and its `cachedImage` draw path | `CachedThumbnail` shows its fallback on nil; the card paints its placeholder when `cachedImage` is nil. Neither needs a change |
| What reaches it | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `sourceCoverThumbnail`; `core/derivatives/ensure.go` `ThumbnailSpec` (JPEG, `MaxEdge: 256`); `core/ingest/stage.go` `stageMaxEdge = 2048` via `raster.EncodePNG` | Every path the app passes is a Go-written JPEG thumbnail derivative or a Go-written likeness PNG |
| Go limit | `core/raster/raster.go` `MaxEdgeLimit = 4096`, `checkMaxEdge` | Both `EncodeJPEG` and `EncodePNG` refuse a `MaxEdge` over 4096, so no Go-written object has a longer edge |
| Go decode slot | `core/raster/raster.go` `decodeSem`, `decodeSlots()` (`min(GOMAXPROCS, 4)`), `Slot() (release func())` | Go's bound on concurrent decodes: a caller takes a slot and `defer release()` frees it. The Swift gate copies the name, the cap, and the scoped shape |
| Mirror test | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo`, `swiftEnumBody`, `matches`, path constants `swiftVocabulary` … | One table of `{name, path, enum, re, want}` rows; `swiftEnumBody` finds `enum <name>` and its closing `\n}\n` at column 0. The header comment above the path constants (lines 20–22) reads "The app reads Go's wire vocabulary through one Swift file. This test reads that file so a rename on either side fails here, not as a silently missing badge or row." |
| Mirror rule | `.cursor/rules/macos-client.mdc` line 20 ("Go wire vocabulary … `api/ffi/swift_vocabulary_test.go` fails if that file drifts from Go") | Names wire strings only. S10-61 D-9 and S10-62 D-10 (both Planned, not landed) add a "limits Swift mirrors" clause |
| Fixtures | `macos/ProvenenciaTests/ArtifactFileFixtures.swift` `uniqueProjectDir()`, `writePNG(projectDir:relPath:width:height:topHalfRed:)` (defaults 8×8), private `write(_:projectDir:relPath:)` | `writePNG` draws a `CGContext` bitmap and writes `NSBitmapImageRep(cgImage:).representation(using: .png …)`; `write` creates parent folders. The enum's doc comment says "Artifact files on disk for tests that load a viewer: a PNG, a PDF, and the catalog row that points at them." Raw `Data.write` plus `createDirectory` is hand-rolled again in `ThumbnailCacheTests` (four cases) |
| Tests | `macos/ProvenenciaTests/ThumbnailCacheTests.swift` `missingFileCachesSentinel`, `loadsAndHitsCache`, `decodesAtABoundedPixelSize`, `cachedImageIsAMissUntilLoaded`, `coalescesInflightLoads` | All decode Go-allowed PNGs (`loadsAndHitsCache` and `coalescesInflightLoads` write PNG bytes at a `.jpg` path; ImageIO reads content, not the extension, so they keep passing). `decodesAtABoundedPixelSize` writes 2000×1000, under 4096 |
| Precedent | `macos/App/Platform/IngestMediaPolicy.swift` `enum IngestMediaPolicy` | A pure Platform enum that holds a Swift-side file policy with a size constant (`maxBytes`). `ThumbnailDecodePolicy` takes the same shape |

No async semaphore, `AsyncSemaphore`, or `maxConcurrent` exists under `macos/App` (`rg -n "Semaphore|maxConcurrent" macos/App` finds none).

**Precedent.** `IngestMediaPolicy` for the policy enum; Go's `raster.Slot` for the scoped decode slot and its cap; S10-61 D-9 for the mirror row; `maxEntries` / `entryCount` for the cache knob and its test helper.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The check is a new file-level `enum ThumbnailDecodePolicy` at column 0 in `ThumbnailCache.swift`, after the class. It holds `static let maxDeclaredEdge = 4096`, `static let allowedTypes: Set<String> = [UTType.png.identifier, UTType.jpeg.identifier]`, and `static func accepts(_ source: CGImageSource) -> Bool`. All members are `nonisolated` by being on a plain enum (no actor). | A pure function is unit-testable without the cache; `swiftEnumBody` finds an `enum` but not a `class`, so the mirror row works unchanged; same shape as `IngestMediaPolicy` | Constants on `ThumbnailCache` (a class: `swiftEnumBody` cannot read it); a separate file (one small type used only here) |
| D-2 | `accepts` returns true only when `CGImageSourceGetType(source)` (as `String`) is in `allowedTypes`, and the properties of image index 0 only — `CGImageSourceCopyPropertiesAtIndex(source, 0, nil) as? [CFString: Any]` — hold `kCGImagePropertyPixelWidth` and `kCGImagePropertyPixelHeight`, each read as `(props[key] as? NSNumber)?.intValue`, each `>= 1` and `<= maxDeclaredEdge`. A nil type, nil properties, or a missing or non-numeric dimension refuses. | ImageIO reads the type and the header dimensions without decoding pixels; checking each edge against one number matches how Go bounds output (`MaxEdge` on the long edge) | A pixel-count limit mirroring `maxDecodePixels` (64 MP allows 8000×8000, twice anything Go writes); trusting the file extension (the existing tests write PNG bytes at `.jpg`) |
| D-3 | `maxDeclaredEdge` mirrors `raster.MaxEdgeLimit`. | It is the largest edge either Go encoder accepts, so every object the app is meant to show is within it, with headroom over the 2048 likeness and 256 thumbnail sizes | Mirroring `stageMaxEdge` (2048): a future larger derivative would silently show as missing |
| D-4 | `ThumbnailCache.image` creates the source in the detached task, calls `ThumbnailDecodePolicy.accepts(src)`, and returns nil when it refuses, before `CGImageSourceCreateThumbnailAtIndex`. A refusal caches `.missing` through the existing nil path. | One check on the one decode path; no new state or copy | A separate pre-flight read of the header bytes (a second file open) |
| D-5 | The decode gate lives on `ThumbnailCache` as main-actor state: `var maxConcurrentDecodes: Int = 4`, `private var runningDecodes = 0`, `private var decodeWaiters: [CheckedContinuation<Void, Never>] = []`, `private(set) var peakConcurrentDecodes = 0` (test helper, beside `entryCount`). One scoped entry point, `private func withDecodeSlot<T: Sendable>(_ body: () async -> T) async -> T`, mirrors Go's `raster.Slot`: it takes a slot when `runningDecodes < max(1, maxConcurrentDecodes)` (a cap of 0 or less counts as 1, so a test cannot deadlock the cache), else appends a continuation and suspends until a releasing caller hands it the slot; then it sets `peakConcurrentDecodes = max(peakConcurrentDecodes, runningDecodes)`, runs `body` with `defer` releasing the slot (resume the first waiter, which inherits the slot, or decrement `runningDecodes`), and returns `body`'s result. The unstructured task calls `await self.withDecodeSlot { await Task.detached { … }.value }` after `objectURL` returns non-nil. There is no separate acquire or release call. The class doc comment gains one sentence: "Decodes only what `ThumbnailDecodePolicy` accepts, at most `maxConcurrentDecodes` at a time." | The class is already `@MainActor`, so a counter and a FIFO need no lock or new type; a scoped call makes release structural, as Go's `Slot()` returning its releaser does, so no path (including PRF-2's later early exits) can leak a slot; `maxEntries` is the precedent for a var knob tests can lower | Separate `acquireDecodeSlot()` / `releaseDecodeSlot()` calls (a release on every path is the caller's job, and a missed one hangs every later waiter); a new `actor` semaphore type (no second user); `TaskGroup` (callers arrive independently, not as a batch); `OperationQueue` (mixes GCD into async code) |
| D-6 | The cap is 4. | Go's `raster.decodeSlots()` runs up to four decodes at once; four bound memory at about four 2048² decodes while keeping list scrolling responsive | `ProcessInfo.activeProcessorCount` (an 8- or 12-core Mac would allow 3× the memory) |
| D-7 | A waiting request is not cancelled when its caller goes away. | The inner task is unstructured today and stays so; cancelling waiters is PRF-2's job (Out) | Cancelling on `Task.isCancelled` (changes coalescing semantics; Out) |
| D-8 | `ArtifactFileFixtures.writeImage(projectDir: String, relPath: String, width: Int = 8, height: Int = 8, fileType: NSBitmapImageRep.FileType = .png, topHalfRed: Bool = false) throws` *(new)* is the one image writer: it takes today's `writePNG` body with `.png` replaced by `fileType`. `writePNG` keeps its signature and becomes a one-line forward to `writeImage(…, fileType: .png, topHalfRed: topHalfRed)`. The private `write(_:projectDir:relPath:)` drops `private` (same name and body) so a test writes raw bytes through it. The enum's doc comment becomes "Artifact files on disk for tests: an image (PNG, JPEG, TIFF), a PDF, raw bytes, and the catalog row that points at them." If S10-62 has landed, its `linkToOutsidePNG` and its tests stay; `writeImage` sits beside it. | `add-swift-test` says artifact files on disk come from `ArtifactFileFixtures` and not to write another copy; one drawing path and one parameter set, so the next fixture is one argument | A second public writer beside `writePNG` with a shared private `bitmap` helper (two entry points with different parameters); a suite-private writer or another base64 literal (three already exist in `ThumbnailCacheTests`) |
| D-9 | The mirror is a row in `TestSwiftVocabularyMatchesGo`: path constant `swiftThumbnailCache = "../../macos/App/Platform/ThumbnailCache.swift"`, `name: "thumbnail decode edge"`, `enum: "ThumbnailDecodePolicy"`, `re: regexp.MustCompile(`static let maxDeclaredEdge = (\d+)`)`, `want: []string{strconv.Itoa(raster.MaxEdgeLimit)}`, with imports `strconv` and `github.com/mendahu/provenencia/core/raster`. A regex with no match yields an empty list, which the row's existing comparison fails. | One table for every Swift-mirrors-Go fact (S10-61 D-9, S10-62 D-10; `go-tests.mdc` one test per family) | A separate test function (the shape both earlier reviews removed) |
| D-10 | After this PR, the `.cursor/rules/macos-client.mdc` Go-vocabulary bullet (line 20 at `e97cf54`) reads exactly: "- Go wire vocabulary (field states, reconciler reasons, seeded Property keys, seeded term keys, date qualifiers) is spelled once, in `Platform/ConclusionVocabulary.swift` (value types: `PropertyValueType`). Views and models use the typed cases. The limits Swift mirrors from Go (`ArtifactRegionGeometry.maxRegionPoints`, `ProjectFiles.objectsDir`, `ThumbnailDecodePolicy.maxDeclaredEdge`) are pinned the same way. `api/ffi/swift_vocabulary_test.go` fails if any of them drifts from Go." The parenthesized list names only the constants that exist on `main` after the merge, in that order: when S10-61 or S10-62 has not landed, its name is left out, and that PR adds it back in the same sentence when it lands. The `api/ffi/swift_vocabulary_test.go` header comment above the path constants is replaced, all three lines, with: "// The app reads Go's wire vocabulary, and mirrors some of Go's limits, through\n// Swift files this test reads. A rename or a changed limit on either side fails\n// here, not as a silently missing badge or row or a wrongly refused input." The new row sits last in the table, after any S10-61 or S10-62 row already there. Whichever of S10-61, S10-62, S10-64 lands later merges `main` and keeps every row, path constant, import, and name, and leaves this exact sentence and comment. | The three PRs edit the same table, header, and rule line with no declared order; one fixed text makes the merge mechanical. The wording carries S10-61 D-9's clause ("the limits Swift mirrors") and S10-62 D-10's name into one sentence | The three PRs edit the same table, header, and rule line with no declared order | Declaring a landing order (they are independent) |
| D-11 | No new copy, glyph, or state. | A refused object is indistinguishable to the researcher from a missing one; each reader already shows its placeholder | A distinct "unreadable image" glyph (Out, UX-11) |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ThumbnailDecodePolicy` *(new)* with `maxDeclaredEdge`, `allowedTypes`, `accepts(_:)` | type | `macos/App/Platform/IngestMediaPolicy.swift` `IngestMediaPolicy` | Ingest checks a file the researcher picks before Go stages it (size, extension, symlink); this checks a stored object before ImageIO decodes it. Different input, same shape. Searched `CGImageSourceGetType|kCGImagePropertyPixelWidth` in `macos/App`: no hits |
| `ThumbnailCache.maxConcurrentDecodes`, `peakConcurrentDecodes`, `withDecodeSlot(_:)` *(new)* | func / var | `core/raster/raster.go` `decodeSem`, `decodeSlots()`, `Slot()`; `ThumbnailCache.maxEntries`, `entryCount` | The Swift twin of Go's decode slot (same name, same cap of 4, same scoped release), kept on the class that owns the decode because ImageIO runs in-process in Swift. Searched `Semaphore|maxConcurrent` in `macos/App`: no hits |
| `ArtifactFileFixtures.writeImage(projectDir:relPath:width:height:fileType:topHalfRed:)` *(new)*; `ArtifactFileFixtures.write(_:projectDir:relPath:)` loses `private` | helper | `ArtifactFileFixtures.writePNG`, private `write`; the raw writers in `ThumbnailCacheTests` | `writeImage` takes `writePNG`'s body plus a file type and `writePNG` forwards to it, so there is one drawing path; the existing `write` covers raw bytes, so no fifth ad-hoc writer |
| Row "thumbnail decode edge" and `swiftThumbnailCache` *(new)* in `api/ffi/swift_vocabulary_test.go` | test row | `TestSwiftVocabularyMatchesGo` rows | A row, not a function |

**Builds on.** `IngestMediaPolicy`'s policy-enum shape, Go's `raster.Slot`, `ThumbnailCache`'s knob pattern, `TestSwiftVocabularyMatchesGo`, `ArtifactFileFixtures`.
**Deletes.** None — nothing becomes dead.
**Next adds.** PRF-2's cancellation and drawn-size keys change `ThumbnailCache.image` inside the same `withDecodeSlot` closure, where an early exit still releases its slot; decoding at the drawn size passes a different `kCGImageSourceThumbnailMaxPixelSize` after the same `accepts`. Another mirrored limit is one more table row.

### Schema

None — no catalog change.

### Go core

None — no Go behavior change. One test row in `api/ffi/swift_vocabulary_test.go` (D-9).

### FFI

None — no engine message changes.

### macOS

`ThumbnailCache.swift` sketch:

```swift
@MainActor
final class ThumbnailCache {
    // … existing members …
    /// Most decodes running at once; later requests wait in arrival order.
    var maxConcurrentDecodes: Int = 4
    /// Test helper: the most decodes that ran at once.
    private(set) var peakConcurrentDecodes = 0
    private var runningDecodes = 0
    private var decodeWaiters: [CheckedContinuation<Void, Never>] = []

    /// Runs `body` holding one decode slot, like Go's `raster.Slot`.
    private func withDecodeSlot<T: Sendable>(_ body: () async -> T) async -> T {
        if runningDecodes < max(1, maxConcurrentDecodes) {
            runningDecodes += 1
        } else {
            await withCheckedContinuation { decodeWaiters.append($0) } // slot handed over
        }
        peakConcurrentDecodes = max(peakConcurrentDecodes, runningDecodes)
        defer {
            if decodeWaiters.isEmpty { runningDecodes -= 1 }
            else { decodeWaiters.removeFirst().resume() }
        }
        return await body()
    }

    // in image(projectDir:relPath:), inside the unstructured Task:
    //   guard let url = ProjectFiles.objectURL(...) else { return nil }
    //   return await self.withDecodeSlot {
    //       await Task.detached(priority: .userInitiated) {
    //           guard let src = CGImageSourceCreateWithURL(url as CFURL, nil),
    //                 ThumbnailDecodePolicy.accepts(src),
    //                 let cg = CGImageSourceCreateThumbnailAtIndex(src, 0, [...] as CFDictionary)
    //           else { return nil }
    //           return NSImage(...)
    //       }.value
    //   }
}

/// What `ThumbnailCache` decodes: an image Go's `raster` package could have
/// written. Matches `raster.MaxEdgeLimit`.
enum ThumbnailDecodePolicy {
    static let maxDeclaredEdge = 4096
    static let allowedTypes: Set<String> = [UTType.png.identifier, UTType.jpeg.identifier]
    static func accepts(_ source: CGImageSource) -> Bool { … }
}
```

`ThumbnailCache.swift` adds `import UniformTypeIdentifiers`.

States: refused → nil → `.missing` → each caller's existing placeholder. Accessibility: unchanged.

### Strings

None — no copy (D-11).

## 7. Steps

**STEP-1 — Decode policy and image fixture**
Files: `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ArtifactFileFixtures.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`. Skill: `add-swift-test`.
Do: add `ThumbnailDecodePolicy` (D-1, D-2, D-3) and `import UniformTypeIdentifiers`. Add `writeImage`, make `writePNG` forward to it, drop `private` from `write`, and update the enum doc comment (D-8). Add T-1 (policy on a `CGImageSource` built from each fixture file with `CGImageSourceCreateWithURL`).
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ThumbnailCacheTests CODE_SIGNING_ALLOWED=NO` passes, including every existing case.

**STEP-2 — The cache checks the policy**
Files: `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`.
Do: call `ThumbnailDecodePolicy.accepts(src)` in the detached decode before `CGImageSourceCreateThumbnailAtIndex` (D-4). Add T-2 – T-4.
Check: the STEP-1 command passes. Removing the `accepts` call locally makes T-3 and T-4 fail (revert after).

**STEP-3 — Decode slot gate**
Files: `macos/App/Platform/ThumbnailCache.swift`, `macos/ProvenenciaTests/ThumbnailCacheTests.swift`.
Do: add `maxConcurrentDecodes`, `peakConcurrentDecodes`, `runningDecodes`, `decodeWaiters`, and `withDecodeSlot(_:)`, and wrap the detached decode in `withDecodeSlot` (D-5, D-6, D-7, the §6 sketch). Add the one-sentence class doc comment (D-5). Add T-5. T-6 is the existing `coalescesInflightLoads`, unchanged; nothing is added for it.
Check: the STEP-1 command passes; then `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/ThumbnailCacheTests -only-testing:ProvenenciaTests/PersonDetailTests -only-testing:ProvenenciaTests/PersonsListTests -only-testing:ProvenenciaTests/EvidenceCardLayoutTests -only-testing:ProvenenciaTests/CitationComposerDepictionTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-4 — Pin the edge to Go**
Files: `api/ffi/swift_vocabulary_test.go`, `.cursor/rules/macos-client.mdc`.
Do: add the T-7 row last in the table, the `swiftThumbnailCache` path constant, and the `strconv` and `core/raster` imports (D-9); replace the header comment and the rule bullet with the exact D-10 text.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; changing `maxDeclaredEdge` to 4095 locally makes the "thumbnail decode edge" row fail (revert after).

**STEP-5 — Docs**
Files: `docs/application-stack.md`.
Do: in the **Read path** paragraph, after "Swift already holds the folder it opened and reads the file with AppKit/SwiftUI.", add: "Thumbnails decode through `ThumbnailCache`, which decodes only a PNG or JPEG declaring at most `raster.MaxEdgeLimit` pixels on each edge, and caps concurrent decodes at `ThumbnailCache.maxConcurrentDecodes`." If S10-62 has landed and added its own sentence after that one, put this sentence after S10-62's.
Then the Definition of done lines for the map row and this plan, after the draft PR has a number (implement-pr step 8).
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10-followups/prs/S10-64-thumbnail-decode-guard.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS` on a macOS host.

Commits, one per behavior:
1. STEP-1 + STEP-2 + STEP-4: "Decode a project thumbnail only when it is a PNG or JPEG within Go's largest edge."
2. STEP-3: "Decode at most four project thumbnails at a time."
3. STEP-5: "Document what the app decodes from the object store."
4. After the PR opens: "Record S10-64 as landed in pull request N." (implement-pr step 8).

## 8. Tests

All Swift cases in `macos/ProvenenciaTests/ThumbnailCacheTests.swift`. Each builds a project with `ArtifactFileFixtures.uniqueProjectDir()`, writes files with `ArtifactFileFixtures.writePNG` or `writeImage`, uses a fresh `ThumbnailCache()`, and removes the project in a `defer`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `decodePolicyAcceptsOnlyGoImages` *(new)*, `@Test(arguments: decodePolicyCases)` over a file-private `struct DecodePolicyCase: Sendable { let label: String; let fileType: NSBitmapImageRep.FileType?; let width: Int; let height: Int; let expected: Bool }` | Each case writes `objects/aa/bb/x.png` in a fresh project — `writeImage(…, width:, height:, fileType:)` when `fileType` is set, else `ArtifactFileFixtures.write(Data((0..<64).map { UInt8(truncatingIfNeeded: $0 &* 37) }), …)` — then `let accepted = CGImageSourceCreateWithURL(url as CFURL, nil).map(ThumbnailDecodePolicy.accepts) ?? false` and `#expect(accepted == c.expected)`. Cases: 8×8 `.png` → true; 8×8 `.jpeg` → true; 8×8 `.tiff` → false; 4096×8 `.png` → true; 4097×8 `.png` → false; 8×4097 `.png` → false; random bytes (`fileType: nil`) → false | AC-1.1 – AC-1.6, D-2 |
| T-2 | Swift | `loadsAGoJPEG` *(new)* | 8×8 JPEG at `objects/aa/bb/x.jpg` → `image` non-nil | AC-1.2 |
| T-3 | Swift | `refusesAnotherFormat` *(new)* | 8×8 TIFF at `objects/aa/bb/x.png` → nil; then, before loading anything else, `entryCount == 1` and `cachedImage` for that key is nil; control: an 8×8 PNG at `objects/aa/bb/y.png` in the same project → non-nil | AC-1.3, D-4 |
| T-4 | Swift | `refusesAnOversizedDeclaredEdge` *(new)* | 4097×8 PNG at `objects/aa/bb/big.png` → nil; 4096×8 PNG at `objects/aa/bb/edge.png` → non-nil | AC-1.4, AC-1.5, D-3 |
| T-5 | Swift | `runsAtMostTheCappedDecodes` *(new)*, `.timeLimit(.minutes(1))` | `maxConcurrentDecodes = 2`; eight distinct 64×64 PNGs at `objects/aa/bb/k0.png` … `k7.png`, requested in a `withTaskGroup(of: Bool.self)` whose children return `await cache.image(…) != nil` → all eight true; `(1...2).contains(cache.peakConcurrentDecodes)`. The cap is the property under test; whether two decodes overlapped depends on main-actor scheduling, so the test does not require exactly 2 | AC-2.1, D-5 |
| T-6 | Swift | existing `coalescesInflightLoads` | unchanged; both results non-nil and `entryCount == 1` with the gate in place | AC-2.2 |
| T-7 | Go | `api/ffi/swift_vocabulary_test.go` `TestSwiftVocabularyMatchesGo` row "thumbnail decode edge" *(new)* | Swift literal equals `raster.MaxEdgeLimit` | AC-3.1, D-9 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `ThumbnailCacheTests` (existing cases write PNG bytes at `.jpg`; they pass because ImageIO reads content), `PersonDetailTests`, `PersonsListTests`, `EvidenceCardLayoutTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none added: D-11)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no brief: no UI change)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together): none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] docs/deployment-plan/spike-10-followups/deployment-plan.md: S10-64 row status → Landed #N (this folder has no completed.md)
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated): not gated
- [ ] Domain doc updated: docs/application-stack.md read path (STEP-5)
- [ ] Guidance updated: macos-client.mdc mirrored-limits clause and the swift_vocabulary_test.go header (STEP-4)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

On a Mac, with a copy of a dogfood project:

1. Open the project. The Persons list, a Person page gallery, and a Source page show their thumbnails and faces as before.
2. Quit the app. In the copy, pick one likeness object under `objects/` (the object lives at `objects/<first two hex>/<next two hex>/<checksum>.png` (`objectpath.Rel`); find a checksum with `sqlite3 <copy>/provenencia.sqlite "SELECT f.checksum_sha256 FROM depiction_values d JOIN files f ON f.id = d.image_file_id LIMIT 1"`) and overwrite it with a TIFF of the same face (`sips -s format tiff <png> --out <object path>`).
3. Relaunch and open that Person. The tile and, when it leads, the header and Persons list thumb show the person mark placeholder; nothing else changes.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Go derivative is written as something other than JPEG or PNG in a future spec | D-3's mirror pins only the edge; a new derivative format is a Decision change that adds the type to `allowedTypes` |
| ImageIO reports a JPEG with EXIF rotation as width/height before rotation | `accepts` checks both edges against one limit, so rotation cannot move a passing image over it |

**Stop and report if:**

- `ThumbnailCache.image` no longer has the shape §4 describes (an unstructured task around one detached decode).
- `CGImageSourceCopyPropertiesAtIndex` returns no pixel width or height for an 8×8 PNG or JPEG written by `writeImage`.
- Any existing `ThumbnailCacheTests` case fails after STEP-2.
- A caller other than `CachedThumbnail` and `EvidenceCardView` reaches `ThumbnailCache` (`rg -n "ThumbnailCache\.shared" macos/App` lists more).
- No macOS host with Xcode is available: every Swift check in this PR is `xcodebuild`, and `scripts/check-pr.sh` skips Xcode tests off macOS.

## 12. Open questions

None.

## 13. PR

**Title:** S10-64 — Check a thumbnail before decoding it
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10-followups/prs/S10-64-thumbnail-decode-guard.md`.

## Landed

Differences from the plan:

- STEP-1, STEP-2, STEP-3 and STEP-4 are one commit, not two, because the policy, the cache check and the decode slot all edit `ThumbnailCache.swift` and its tests together.
- `macos-client.mdc` and the `swift_vocabulary_test.go` header already carried S10-61 and S10-62 text on `main`; this PR only added `ThumbnailDecodePolicy.maxDeclaredEdge` to the rule's list and kept the existing header (a superset of the D-10 wording).
- The manual check (section 10) is left for the user.
- User-approved amendment to D-5 (2026-10-18): `ProjectFiles.objectURL` now runs inside the decode slot, right before the file opens, so the slot wait no longer sits between the object-path check and the open.
- User-approved amendment to D-7 (2026-10-18): a queued decode whose callers have all gone is dropped without decoding and is not stored as missing; a cancelled caller returns nil at once. This pulls part of PRF-2 forward. New test: `ThumbnailCacheTests/dropsAQueuedDecodeWhoseCallersHaveGone`.
- Fix round 2: `ThumbnailCache.clear()` had no caller and could split one load's state, so it was removed along with `clearGeneration`; `finish` now always clears `inflight[key]`.
- `CachedThumbnail.load()` ignores a cancelled load's result (user-requested, review follow-up).
