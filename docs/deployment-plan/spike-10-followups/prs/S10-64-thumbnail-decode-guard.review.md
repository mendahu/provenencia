# Review — S10-64 Check a thumbnail before decoding it

**Verdict:** Fits with changes — Ready with changes
**Rounds:** 1. **Reviewers:** maintainability (opus, pr-reviewer), executability (sonnet, Explore)

## To user
- None.

## Writer's own calls
- D-1: the check is a file-level `enum ThumbnailDecodePolicy` in `ThumbnailCache.swift` (`swiftEnumBody` reads an enum, not a class); same shape as `IngestMediaPolicy`.
- D-2: content type (not extension) must be PNG or JPEG, and each declared edge of image 0 must be 1 … 4096; anything missing refuses.
- D-3: the limit mirrors `raster.MaxEdgeLimit` (4096), not `stageMaxEdge` (2048) or `maxDecodePixels` (64 MP).
- D-4: a refusal returns nil before `CGImageSourceCreateThumbnailAtIndex` and caches `.missing` through the existing path.
- D-5: the gate is main-actor state on `ThumbnailCache` with one scoped `withDecodeSlot` (after round 1); no new semaphore type.
- D-6: the cap is a fixed 4, matching Go's `raster.decodeSlots()`.
- D-7: waiters are not cancelled (PRF-2, Out).
- D-8: one image writer, `ArtifactFileFixtures.writeImage`; `writePNG` forwards to it; `write` loses `private` for raw bytes (after round 1).
- D-9: the mirror is a "thumbnail decode edge" row in `TestSwiftVocabularyMatchesGo`.
- D-10: one fixed text for the `macos-client.mdc` bullet and the test-file header comment, naming every mirrored constant on `main` after the merge (after round 1).
- D-11: no new copy, glyph, or state.
- No version bump.

## Round 1

### M-1 · major · Make the decode slot one scoped call, modelled on Go's `raster.Slot`
**Plan:** D-5 added separate `acquireDecodeSlot()` / `releaseDecodeSlot()`; a leaked slot was guarded only by a T-5 time limit. **Codebase:** `core/raster/raster.go` `decodeSem`, `decodeSlots()`, `Slot() (release func())` — the same mechanism, uncited. **Better shape:** one `withDecodeSlot` with `defer` release; cite `raster.Slot`.
**Answer:** Fixed — D-5 now specifies `private func withDecodeSlot<T: Sendable>(_ body: () async -> T) async -> T` with a `defer` release that hands the slot to the first waiter; §6 sketch rewritten; Codebase fit row names `decodeSem` / `decodeSlots()` / `Slot()`; §4 gains a "Go decode slot" row; Precedent and Builds on cite it; D-6's reason cites `decodeSlots()`; the slot-leak risk row is removed; Next adds says PRF-2's early exits land inside the closure.

### M-2 · minor · One image writer in `ArtifactFileFixtures`, plus raw bytes for T-1
**Plan:** D-8 added `writeImage` beside `writePNG` with a shared private `bitmap`; T-1's random bytes had no writer. **Codebase:** `ArtifactFileFixtures.writePNG`, private `write`; raw writers hand-rolled in `ThumbnailCacheTests`; `add-swift-test` "do not write another copy".
**Answer:** Fixed — D-8: `writeImage(projectDir:relPath:width: = 8, height: = 8, fileType: = .png, topHalfRed: = false)` is the one writer, `writePNG` forwards, no private `bitmap`; `write(_:projectDir:relPath:)` drops `private` and T-1 writes its random bytes through it. Codebase fit and STEP-1 updated.

### M-3 · nit · The read-path sentence hard-codes "four"
**Answer:** Fixed — STEP-5's sentence names `ThumbnailCache.maxConcurrentDecodes`.

### X-1 · blocker (for this host) · No Swift check can run on Linux
**Answer:** Declined — a property of the dry-run host, not of the plan. §11 already stops when no macOS host with Xcode is available, worded as S10-62's stop condition; `ship-pr` runs on the user's Mac.

### X-2 · major · A 4096×1 image may not produce a 512-px thumbnail
**Answer:** Fixed — AC-1.4 / AC-1.5 and T-4 use 4097×8 and 4096×8; T-1 uses 4096×8, 4097×8, and 8×4097 for consistency.

### X-3 · major · The `.mdc` clause and header comment are not fully worded
**Plan:** D-10 described partial edits; three PRs (S10-61 D-9, S10-62 D-10, S10-64) edit the same line and comment. **Answer:** Fixed — D-10 gives the exact final bullet and the exact three-line header comment, naming `ArtifactRegionGeometry.maxRegionPoints`, `ProjectFiles.objectsDir`, `ThumbnailDecodePolicy.maxDeclaredEdge` (only those on `main` after the merge, in that order), and places the new row last. §4 Mirror test row quotes today's header comment.

### X-4 · minor · T-1 shape is undefined
**Answer:** Fixed — T-1 names `DecodePolicyCase` (label, `fileType?`, width, height, expected), `@Test(arguments: decodePolicyCases)`, the random-bytes recipe through `ArtifactFileFixtures.write`, and the single outcome `CGImageSourceCreateWithURL(…).map(accepts) ?? false`.

### X-5 · minor · `writeImage` signature has no defaults or argument order
**Answer:** Fixed — D-8 gives the full signature with defaults and the enum doc-comment replacement.

### X-6 · minor · T-5 can flake and its shape is underspecified
**Answer:** Fixed — T-5 carries `.timeLimit(.minutes(1))`, names keys `objects/aa/bb/k0.png` … `k7.png`, uses `withTaskGroup(of: Bool.self)`, and asserts `(1...2).contains(peakConcurrentDecodes)` with the reason; AC-2.1 matches.

### X-7 · minor · T-3 ordering of the `entryCount` assertion
**Answer:** Fixed — T-3 asserts `entryCount == 1` and `cachedImage == nil` before loading the control.

### X-8 · minor · Small step-level gaps
**Answer:** Fixed — STEP-3 says T-6 is unchanged and adds nothing; STEP-3's second check is a full command; STEP-4 names T-7 and puts the row last; D-2 reads index 0 only, as `[CFString: Any]`, dimensions via `NSNumber.intValue`, nil properties refuse; D-5 treats a cap of 0 or less as 1.

### X-9 · minor · Two ACs are covered less strongly than they read
**Answer:** Fixed — AC-1.6 says it is covered at the policy layer by T-1; AC-2.2 now reads "both get the image and the cache holds one entry", matching `coalescesInflightLoads`.

### X-10 · minor · Swift-file merge overlap with S10-62 not covered
**Answer:** Fixed — D-8: if S10-62 has landed, `linkToOutsidePNG` and its tests stay and `writeImage` sits beside it.

### X-11 · nit · Status and temptation
**Answer:** Status: Declined — Draft until review passes (`plan-pr` step 8); the plan is now Planned. Temptation: Fixed — the class doc comment update is in scope as one sentence (D-5, Scope, STEP-3) instead of an Out line, because the comment describes the decode the PR changes.
