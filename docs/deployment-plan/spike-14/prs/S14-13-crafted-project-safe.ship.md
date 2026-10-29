# Ship log — S14-13 A crafted project can't crash the app

**Plan:** [`S14-13-crafted-project-safe.md`](S14-13-crafted-project-safe.md) · **Branch:** `s14-13-crafted-project-safe` · **PR:** #632
**State:** ready

## Round 0 — implement
Head `ecf72552`. Gate PASS (orchestrator run, base `51cf4aef`). PR #632 (draft).
Report deviations: (1) T-1 raw column keeps the file's extreme value because `writes.Run` rolls back a no-change write; reads clamp. (2) T-8 PNG fixture needs IDAT filler (~85 KB at 16000²) before ImageIO reports a size. (3) D-10: `UTType(mimeType: "image/tif")` returns a dynamic type, treated as unknown so Go's extension fallback runs; HEIC not fixtured. (4) D-5 close/unlock/forget in a `runLocked` helper. (5) `installImageRasterForTesting` no-ops without a raster. (6) STEP-3 kept hits listed in the PR body. (7) `maxDecodePixels` sweep hits are the D-6 constant, its doc comment, the vocabulary regex, and `pixelSize(scale:)`. (8) Manual check not run by hand.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none (diff touches `macos/App/**`, `core/raster`, `core/catalogsession`, `api/ffi/**`, `ThumbnailCache.swift`, Features models).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1, CON-1 | UX, CON | blocker | A `.heif` Artifact (HEVC, sniffed `public.heic`) is refused and shows load-failed | To user | Plan §11 stop condition 2 fired: `public.heic` does not conform to `public.heif` (both reviewers' probes); Go maps `.heif` → `image/heif` (`mediatypes.go:124,166`). Every fix contradicts D-10 ("no Swift alias map") |
| SEC-1 | SEC, PRF | major | Pixel budgets ignore bytes per pixel; 16-bit/float images decode at 2–4× the planned memory | To user | Probe: 16-bit PNG/TIFF decode at 64 bpp, float TIFF at 128 bpp. Byte-correct budget would refuse a 16-bit scan above 128 MP, against D-7's intent |
| UX-2 | UX | major | An image refused only for size gets "couldn't be read" copy | To user | Contradicts D-9 (no new reason or copy); §11 stop condition 6 |
| MNT-1, CON-2 | MNT, CON | major | `appliesOrientation` and `applyTransform:` kept at the single value `true` | Fix now → F-1 | verified: D-12 says delete both when `true`; `ThumbnailCache.swift:82` and the viewer both pass `true` |
| MNT-2 | MNT | major | The clamp-cell composition is written out at 7 sites plus a private `movedCell` | Fix now → F-2 | Same-pass cleanup; verified `rg 'clampedCell\|offsetCell\|movedCell\|gridCell\(' macos/App`. Meets D-3's invariant through one helper |
| MNT-3 | MNT | minor | `ThumbnailDecodePolicy` now holds the viewer's header read and decode under a thumbnail name | To user | D-11 chose the name and home |
| MNT-4 | MNT | minor | `canAutoTranscribe` gates on `imageRaster` (NSImage) but runs on `displayRaster`; test hook name says raster | Fix now → F-4 | Minor in files this PR changes; gating on `displayRaster` keeps D-16 (`imageRaster` stays) |
| SEC-2 | SEC | minor | A panic in `openResearcher` after `database.Open` leaks the catalog on every retry | Fix now → F-3 | verified: `session.go` closes `c` only on error returns; the new recover makes this reachable |
| CON-3 | CON | minor | Landed note omits deviations | Fix now → F-5 | DoD |

Nits (logged): `MNT-5 · .cursor/rules/macos-client.mdc:20 · add cellLimit and maxDecodePixels to the mirrored-limits list` · `MNT-6 · ArtifactRegionCrop.swift:22 · cgImage(from:) only feeds the test hook` · `MNT-7 · subjectpositions_test.go:426 · max shadows builtin` · `UX-3 · L10n.swift:3817 · load-failed copy now covers a type mismatch` · `CON-4 · no change`

**User answers (2026-10-29):** UX-1 → Go supplies the accepted type identifiers (F-6, supersedes D-10 and its rule bend). SEC-1 → scale both budgets by bit depth (F-7, refines D-6/D-7). UX-2 → add a "too large" empty state (F-8, supersedes D-9). MNT-3 → move the shared ImageIO steps to their own Platform file (F-9, supersedes D-11's home). Rows above now read Fix now.

### Fix plan 1

Order: F-6 (blocker), F-7, F-8, F-1, F-2 (majors), then F-9, F-3, F-4, F-5. Each lands as its own commit. Superseded Decisions are named in each item; record each in the plan's Landed note (F-5) as "User, 2026-10-29, ship-log round 1".

**F-6 — Go supplies the type identifiers an image object must sniff as** (from UX-1, CON-1; supersedes D-10 and the "media type, else extension" bend)
Files: `core/ingest/mediatypes/mediatypes.go` (+ test), `api/proto/engine.proto` (+ regenerated Go and `macos/App/Platform/Generated/engine.pb.swift`), `api/ffi/handlers/sources.go:650`, `macos/App/Platform/GenealogyStore.swift` (`CatalogFileRef`), `GoStore` mapping, `FakeStore`, `macos/App/Features/ArtifactViewer/ArtifactViewerKind.swift` (`ArtifactViewerSource`), `CitationComposerModel.swift:1208`, `ArtifactImageDecode.swift`, `macos/ProvenenciaTests/ArtifactViewerModelTests.swift`, `ArtifactFileFixtures.swift`. Skill: `add-ffi-handler` (proto regen steps).
Failing test first: `opensEveryAllowedImageType` gains rows `("image/heif", ".heif")` and `("image/heic", ".heic")` whose fixture is written through `CGImageDestination` as `public.heic` (extend `ArtifactFileFixtures.writeImage`'s file-type switch; do not add a second writer). Run red on the current head: the `.heif` row fails.
Do:
1. In `mediatypes.go`, turn the `allowed` table's value into a small struct `{ext string; typeIdentifiers []string}` so the one table of stored types also says which Uniform Type Identifiers a stored file of that type may sniff as. Fill identifiers for the image rows only: jpeg `public.jpeg`; png `public.png`; gif `com.compuserve.gif`; webp `org.webmproject.webp`; tiff and tif `public.tiff`; bmp `com.microsoft.bmp`; heic `public.heic`, `public.heif`; heif `public.heif`, `public.heic`. Other rows keep none. `Allowed` and `StoredExtension` read the struct unchanged in behavior. Add exported `TypeIdentifiers(mediaType string) []string` (copy of the slice; nil when unknown or none). Go test: every `image/*` row has at least one identifier; `TypeIdentifiers("image/heif")` contains `public.heic`; unknown → nil.
2. `SourceFileRef` gains `repeated string type_identifiers = 6;` with a one-line comment ("UTIs a stored image may sniff as; Go's stored-type table. Empty for non-images."). `sources.go` fills it from `mediatypes.TypeIdentifiers(f.MediaType)`. Handler test asserts a png file ref carries `public.png`.
3. Swift: `CatalogFileRef` gains `typeIdentifiers: [String]`, mapped in `GoStore`; `FakeStore` fills it the same way for the types its fixtures use (a small private helper mapping media type → identifiers is test-double code, allowed). `ArtifactViewerSource` gains `typeIdentifiers: [String]`; `CitationComposerModel` passes `artifact.file?.typeIdentifiers ?? []`.
4. `ArtifactImageDecode.decode(url:mediaType:relPath:)` becomes `decode(url:typeIdentifiers:)`: refuse unless `UTType(sniffed)` conforms to at least one `UTType(identifier)` in the list (empty list → refused). Delete `expectedType` and its `isDynamic` / extension fallback.
5. Docs: reword the working rule "What the app hands to macOS to open is checked first, and Go decides the type" and the S14-13 scope entry back to "the sniffed type conforms to one of the type identifiers Go's stored-type table gives the media type"; drop the D-10 bend paragraph's effect in the Landed note.
Check: `opensEveryAllowedImageType` passes for jpeg, png, gif, tiff, tif, bmp, heic, heif; `refusesAMismatchedType` still passes; `rg -n 'expectedType|isDynamic' macos/App` prints nothing; `go test ./core/ingest/mediatypes/... ./api/ffi/...`; scripts/check-pr.sh PASS.

**F-7 — Both viewer budgets count bytes, scaled by bit depth** (from SEC-1; refines D-6/D-7: the constants stay, as 8-bit pixel counts)
Files: `macos/App/Platform/<F-9's file>` (`DeclaredImage`, `declaredImage`), `ArtifactImageDecode.swift`, `ArtifactFileFixtures.swift`, `ArtifactViewerModelTests.swift` / T-8 and T-9 test files.
Failing test first: `writeDeclaredPNGHeader` / `declaredPNGHeaderData` gain `bitDepth: Int = 8` (IHDR depth byte). New cases: (a) `declaredImageBounds` reads depth 16 from a 16-bit header; (b) the viewer refuses a 16-bit header declaring 12000 × 12000 (144 MP > 256 MP / 2); (c) `maxPixelSize(width:height:depth:)` for a 16-bit 10000 × 10000 is `floor(10000 × sqrt(32,000,000 / 100,000,000))` = 5656.
Do: `DeclaredImage` gains `bitsPerComponent: Int` from `kCGImagePropertyDepth` (default 8 when absent or outside `1...64`). In `ArtifactImageDecode`, a private `depthFactor = max(1, bitsPerComponent / 8)`; the declared gate uses `maxDeclaredPixels / depthFactor` (checked after the header read, same refusal), and `maxPixelSize` uses `maxDecodePixels / depthFactor`. The thumbnail cache's `accepts` is unchanged. Doc comments on the two constants say "8-bit pixels; divided by bytes per component".
Check: new cases pass; AC-2.4 and every existing T-9 case unchanged; scripts/check-pr.sh PASS.

**F-8 — A size refusal gets its own empty state** (from UX-2; supersedes D-9 for the size case only)
Files: `ArtifactImageDecode.swift`, `ArtifactViewerModel.swift`, `ArtifactViewerContent.swift` (`ArtifactViewerEmptyReason`), `ArtifactViewer.swift` (empty-state mapping ~289-305), `L10n.swift`, `Localizable.xcstrings`, viewer tests. Skill: `add-localized-string`.
Failing test first: `refusesAnOverBudgetDeclaration` asserts `emptyReason == .imageTooLarge` (was `.loadFailed`); `refusesAMismatchedType` keeps `.loadFailed`.
Do: `decode` returns `Result<Decoded, ArtifactImageDecode.Refusal>` with `enum Refusal: Error { case tooLarge, unreadable }`. `tooLarge` = `declaredImage` read the header but the size is out of bounds (edge outside `1...maxEdge`, or over the depth-scaled pixel gate) — so `declaredImage` must let the caller tell "no header" from "over budget": return the parsed `DeclaredImage` and let each caller apply its own bounds through a `fits(maxEdge:maxPixels:)` method, with `accepts` calling it. Every other failure is `unreadable`. The model maps `.tooLarge` to new `ArtifactViewerEmptyReason.imageTooLarge`, `.unreadable` to `.loadFailed`. Strings: `artifactViewer.imageTooLargeTitle` "Image too large to show"; `artifactViewer.imageTooLargeMessage` "This image is larger than Provenencia can display. A smaller copy of the file will open." Comments describe when each shows. The composer's auto-transcribe hint for this case stays the existing no-image hint.
Check: tests pass; `localizable` check PASS; scripts/check-pr.sh PASS.

**F-1 — Delete the single-valued orientation flag** (from MNT-1, CON-2; does what D-12 says)
Files: `ArtifactImageDecode.swift`, `ThumbnailCache.swift` (or F-9's file).
Do: delete `appliesOrientation`; drop `applyTransform:` from `decode` and pass `kCGImageSourceCreateThumbnailWithTransform: true`; `turned = declared.orientation >= 5`; update the `ThumbnailCache` call site and the doc comments (`Decoded.pixelSize`).
Check: `rg -n 'applyTransform|appliesOrientation' macos` prints nothing; T-11 `orientsAsNSImageDoes` passes.

**F-2 — One helper per clamped-cell composition** (from MNT-2; same-pass cleanup, meets D-3 through helpers)
Files: `macos/App/Features/Catalog/Connect/EvidenceGraphPlacement.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`.
Do: beside `clampedCell` add `static func offsetCell(_ cell: Int64, by delta: Int64) -> Int64` (= `clampedCell(GraphCanvasGridMapping.offsetCell(cell, by: delta))`) and `static func cell(at gridPoint: CGPoint) -> (gridX: Int64, gridY: Int64)` (calls `GraphCanvasGridMapping.gridCell(gridPoint:)`, clamps both axes; match `gridCell`'s return shape). Move the four placement sites (181, 187, 193, 213) and the model's `moveSubject`/`moveSubjects`, `updateHover`, `beginCreate`, `snapAndPatch` onto them; delete private `movedCell`. `GraphCanvas/` stays free of `cellLimit`.
Check: `rg -n 'clampedCell\(' macos/App` prints only the definition and the two new helpers' bodies; `rg -n 'movedCell' macos/App` prints nothing; `rg -n 'GraphCanvasGridMapping\.(offsetCell|gridCell)' macos/App/Features/EvidenceGraph macos/App/Features/Catalog` prints only the two helpers' bodies; T-5–T-7 pass.

**F-9 — The shared ImageIO steps get their own Platform file** (from MNT-3; supersedes D-11's home, not its content)
Files: new `macos/App/Platform/ImageHeaderDecode.swift` (enum `ImageHeaderDecode`), `ThumbnailCache.swift`, `ArtifactImageDecode.swift`, `project.pbxproj`, `.cursor/rules/macos-client.mdc`, any skill or doc naming `ThumbnailDecodePolicy.declaredImage`/`.decode`.
Do: move `DeclaredImage`, `declaredImage` and `decode` into `ImageHeaderDecode`; `ThumbnailDecodePolicy` keeps `allowedTypes`, `maxDeclaredEdge` and `accepts`, and drops the "Also owns…" comment. Do this commit before F-7, F-8 and F-1 so they edit the new file. Fold in nit MNT-5 (add `EvidenceGraphPlacement.cellLimit` and `ArtifactImageDecode.maxDecodePixels` to the mirrored-limits list in `macos-client.mdc`).
Check: `rg -n 'ThumbnailDecodePolicy\.(declaredImage|decode)|DeclaredImage' macos/App` shows only `ImageHeaderDecode` users; `TestSwiftVocabularyMatchesGo` passes.

**F-3 — A panic while opening a researcher catalog closes it** (from SEC-2)
Files: `core/catalogsession/session.go` (`openResearcher`), `session_test.go`.
Failing test first: if `openResearcher` has a seam for reconcile/warm, a test that panics there and asserts the catalog was closed; if there is no seam, skip the test and say so in the report (do not add a seam only for this).
Do: after `database.Open` succeeds, `ok := false; defer func() { if !ok { _ = c.Close() } }()`; set `ok = true` just before `return c, nil`; drop the now-redundant explicit `c.Close()` calls on the error returns. No `recover`.
Check: `go test -race ./core/catalogsession/...`.

**F-4 — Auto-transcribe gates on the raster it runs on** (from MNT-4; keeps D-16)
Files: `CitationTranscriptionModel.swift`, `ArtifactViewerModel.swift`, tests calling the hook.
Do: `canAutoTranscribe` requires `artifactViewer.displayRaster != nil` (not `imageRaster`); rename `installImageRasterForTesting(_:raster:)` to `installImageForTesting(_:raster:)` at every caller. Fold in nit MNT-7 (`max` → `edge` in `subjectpositions_test.go:426`).
Check: `refusedImageCannotTranscribe` and `autoTranscribeUsesTheDecodedRaster` pass.

**F-5 — Landed note** (from CON-3)
Files: `docs/deployment-plan/spike-14/prs/S14-13-crafted-project-safe.md`.
Do: add to the Landed note every round-0 deviation (`installImageRasterForTesting` no-op, T-1 raw column, T-8 filler, D-5 helper) and each round-1 supersession above (D-9, D-10, D-11 home, D-6/D-7 depth scaling), each as "User, 2026-10-29".
Check: scripts/check-pr.sh PASS (plan-docs).

## Round 1 — fix
Head `31bc43f5`. Gate: first run FAIL in `xcode-test` (23 issues, all `SourceGraphFormModelTests` edit cases and `SourcesModelTests.listReadsEachGraphsProgress`, in suites and code this PR does not touch); immediate rerun on the same head PASS. Logged as a flaky-test follow-up.
F-1, F-2, F-3, F-4, F-5, F-7, F-8, F-9 done. F-6: the failing `.heif` row was **not reproduced**: ImageIO sniffs a `.heif`-named file as `public.heif` (`CGImageSourceCreateWithURL` uses the extension), so the old check accepted it; the reviewers' `public.heic` sniff came from the same bytes without a `.heif` name. UX-1/CON-1 → Rejected as a defect (evidence: implementer probe). F-6 kept as the user's design answer (Go supplies the type identifiers; replaces D-10's extension fallback and its `isDynamic` special case).
Deviations: `writeImage` takes an optional `contentType: UTType?` for HEIC; identifiers default to `[]` (an image with none is refused); the 16-bit header fixture's filler scales with depth; F-2 added a third private helper `flooredCell(_:)` for `maxGridX`/`maxGridY`; F-3 has no test (no seam in `openResearcher`).
Nits MNT-5 and MNT-7 folded in.

## Round 2 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF — each had a round-1 blocker or major. Scope `ecf72552...31bc43f5`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| UX-1, CON-1 | UX, CON | — | `.heif` refused | Resolved | F-6; `opensEveryAllowedImageType` heic and heif rows |
| UX-2 | UX | — | Size refusal copy | Resolved | F-8; `.imageTooLarge` with its own strings |
| SEC-1 | SEC, PRF | — | Bytes per pixel | Resolved | F-7; reviewer probe: 16-bit, float and 10-bit HEIC all held at 4 bytes per budgeted pixel |
| MNT-1, CON-2 | MNT, CON | — | Single-valued orientation flag | Resolved | F-1 |
| MNT-2 | MNT | — | Clamped-cell copies | Resolved | F-2; `flooredCell` third helper accepted |
| SEC-2, MNT-3, MNT-4, CON-3 | — | — | Minors | Resolved (orchestrator check) | F-3 `ok`/deferred Close in `openResearcher`; F-9 `ImageHeaderDecode.swift`; F-4 `canAutoTranscribe` reads `displayRaster`; F-5 Landed note |
| MNT-8 | MNT | blocker | `FakeStore.typeIdentifiers(forMediaType:)` copies Go's identifier table with no parity test; the Swift test that proves every type opens reads the copy | Fix now → F-10 | verified `FakeStore.swift:875-888`; CON noted the same copy as a minor. Kept at blocker: Fix plan 1 created the copy, and the user's F-6 answer (Go owns the identifiers) is unproven in Swift without the pin; a Go typo would ship refused images with every test green |

Nits (logged): `UX-4 · L10n.swift:3829 · "A smaller copy of the file will open." reads as if the app opens one; suggest "Attach a smaller copy of the file to view it here."`

### Fix plan 2

**F-10 — Pin FakeStore's type identifiers to Go's table** (from MNT-8)
Files: `macos/App/Platform/FakeStore.swift`, `api/ffi/swift_vocabulary_test.go`, `.cursor/rules/macos-client.mdc:22` (add the new fixture to the mirror list).
Failing test first: the new Go test below, with one Swift row deliberately wrong locally, fails; then fix the row. (Do not commit the wrong row.)
Do: replace the `switch` in `FakeStore.typeIdentifiers(forMediaType:)` with `static let imageTypeIdentifierFixtures: [(String, String)]`, one row per image media type in Go's table, the second element the identifiers joined by a single space in Go's order (`("image/heic", "public.heic public.heif")`). `typeIdentifiers(forMediaType:)` reads that table and splits on spaces (unknown → `[]`). In `swift_vocabulary_test.go` add `TestFakeStoreTypeIdentifiersMatchGo`: read the pairs with `swiftFixturePairs(t, <FakeStore path>, "FakeStore", "imageTypeIdentifierFixtures")`; assert each row's second element equals `strings.Join(mediatypes.TypeIdentifiers(row[0]), " ")`, and that the rows' media types are exactly the set of image media types in Go's table that have identifiers. If `mediatypes` has no way to list them, add exported `ImageMediaTypes() []string` (sorted) beside `TypeIdentifiers` and use it. If `swiftFixturePairs` cannot read a `static let` on a `final class` (it was written for enums), extend its matcher minimally and add a case to `TestSwiftFixturePairsDecodesEscapes` or its sibling.
Check: `go test -tags fts5 ./api/ffi/...` passes; Swift tests that call `typeIdentifiers(forMediaType:)` pass unchanged; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `86416e20`. Gate PASS (orchestrator run). F-10 done: `TestFakeStoreTypeIdentifiersMatchGo` red with one row wrong, green after; `mediatypes.ImageMediaTypes()` added; `swiftFixturePairs` also reads a `final class` (covered by `TestSwiftFixturePairsReadsFinalClass`). MNT-8 Resolved (orchestrator check; round 2 has no re-review).

## Finish
Merged `origin/main` (`cb828f82`, plan docs only) → `b431fee0`. Gate PASS against `cb828f82`.

## Result
Gate PASS · CI green (Go tests, macOS tests) · open: none · follow-ups: flaky `SourceGraphFormModelTests` edit cases and `SourcesModelTests.listReadsEachGraphsProgress` (one failing run on the round-1 head, green on rerun); UX-4 copy nit · to user: UX-1, SEC-1, UX-2, MNT-3 (answered 2026-10-29) · CON miss: none
