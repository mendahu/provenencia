# S14-23 — Thumbnails for oversized JPEGs

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a thumbnail that was missing now appears in the existing tile, header and cover slots; no control, copy or layout changes)
**Board:** —
**Depends on:** —
**Blocks:** —
**Branch:** `s14-23-oversized-jpeg-thumbnails`
**Version bump:** patch, 0.2.2 → 0.2.3 — a missing thumbnail is a bugfix with no new screen or RPC ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** — (no skill covers `core/raster`; `.cursor/rules/go-tests.mdc` is the only binding rule)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [design](../oversized-jpeg-thumbnails.md); working rule "(Batch 1) Oversized JPEGs get a thumbnail from a reduced decode of our own"; map scope entry [S14-23](../deployment-plan.md#s14-23--thumbnails-for-oversized-jpegs); likeness audit SEC-3 (a raster byte limit, [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md)); bar outcome 12; ITU-T T.81 (Annex B syntax, Annex F.2.2 Huffman decoding, Annex G.1.2 progressive DC scans, A.2 block order)
**Planned:** 2026-10-31 against `main` @ `b6a8f3e`
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

A JPEG scan over 64 megapixels (the dogfood file is 11,000 × 18,000, 39 MB) saves as an Artifact and then shows the missing-image glyph everywhere, with no error, and cannot be chosen as a Source cover. After this PR Go decodes such a JPEG at 1/8 scale from its DC coefficients only, then scales and encodes the usual 256 px thumbnail, so the tile, the Source cover and the omnibar show it. The decoder is pure Go, so a Windows build reuses it. `core/raster` also refuses any input over a byte limit of its own before it decodes, instead of trusting each caller's cap.

**Why now.** Batch 1's last engine item (bar outcome 12), moved from Spike 12 (S12-39). The design was agreed on 2026-10-27.

**What stays.** Every JPEG within 64 MP, and every PNG, GIF, BMP, TIFF and WebP, decodes exactly as today through the standard library. Other formats over the budget stay refused. Staged depiction images (`ingest.StageImage`, `EncodePNG`) never take the reduced path. The Mac viewer's own bounded decode (S14-13) is untouched.

## 2. User stories

**US-1.** As a researcher who attaches an oversized JPEG scan, I see its thumbnail and can use it as the Source cover, so that the scan is not a blank tile.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `raster.EncodeJPEG` on a baseline 4:2:0 JPEG of 8,008 × 8,000 (64.06 MP, over budget) with `Options{MaxEdge: 256}` | Returns a JPEG of 256 × 255 |
| AC-1.2 | `derivatives.EnsureThumbnail` on a `files` row whose object is a grayscale JPEG of 8,008 × 8,000 | Returns a `file_derivatives` link; its derived File is `image/jpeg` and decodes to 256 × 255; a second call returns the same link id |
| AC-1.3 | A JPEG whose SOF declares 60,000 × 60,000 (1/8 image 56.3 MP, within budget) followed by an SOS header and 16 bytes of scan data | `EncodeJPEG` returns an error that `errors.Is` `ErrTooLarge`; the call allocates less than 16 MiB in total |
| AC-1.4 | A JPEG whose SOF declares 65,535 × 65,535 (1/8 image 8,192 × 8,192 = 67.1 MP, over budget) | `ErrTooLarge` before any coefficient buffer is allocated |
| AC-1.5 | `EncodePNG` (the staging path) on the 8,008 × 8,000 JPEG of AC-1.1 | `ErrTooLarge`, as today |
| AC-1.6 | The existing PNG and GIF header bombs in `TestEncodeJPEGRejects` | Still `ErrTooLarge` |

**US-2.** As the engine, the reduced decoder gives the same picture as the standard library would at 1/8 scale, and refuses what it cannot read instead of guessing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Each decodable fixture in `core/raster/testdata/` *(new)* (baseline and progressive; 4:4:4, 4:2:2, 4:2:0, 4:4:0, 4:1:1, 4:1:0, 2×2/1×2/1×2; RGB; gray; gray tagged 2×2; restart interval 20; separate DC scans) through `decodeJPEGDC` | Bounds are 19 × 13 (⌈150/8⌉ × ⌈103/8⌉); per channel, the mean absolute difference from the standard library's decode averaged over each 8 × 8 cell is at most 3.0 |
| AC-2.2 | `video-001.cmyk.jpeg` (four components) | Error; no image |
| AC-2.3 | `video-001.progressive.truncated.jpeg`; the AC-1.1 JPEG cut to half its length; the same JPEG with its EOI removed | Error; no image |
| AC-2.4 | The fuzz seeds (every fixture, plus small JPEGs from `jpeg.Encode`) through `FuzzDecodeJPEGDC` | No panic; every success has bounds ⌈W/8⌉ × ⌈H/8⌉ and at most the budget the fuzz passes |

**US-3.** As the engine, every raster decode is bounded by bytes as well as pixels.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | With the byte limit set to 64 in the test, `EncodeJPEG` and `EncodePNG` on 65 bytes of a registered test format | `ErrTooLarge`; the format's `DecodeConfig` is never called |
| AC-3.2 | With the byte limit set to 64, the same on exactly 64 bytes | The call reaches `DecodeConfig` (the test format's counter is 1) |
| AC-3.3 | `derivatives` refuses a `files` row declaring `raster.MaxDecodeBytes + 1` bytes | `ErrUnprocessable`, as today with its own constant |

## 3. Scope

| In | Covered by |
| --- | --- |
| The reduced decoder `decodeJPEGDC`: baseline and extended sequential (SOF0, SOF1) and progressive (SOF2) Huffman, 8-bit, one or three components, any sampling factors the standard library accepts, restart intervals, interleaved and single-component scans | STEP-2, STEP-3, STEP-4, AC-2.1 |
| Refusals: four components, and any marker the decoder does not read (arithmetic, lossless, hierarchical) return an error | STEP-2, AC-2.2 |
| Every count read from the file checked before use; the 1/8 image checked against `MaxDecodePixels`; scan data checked before the coefficient buffers are allocated | STEP-2, STEP-3, AC-1.3, AC-1.4 |
| `EncodeJPEG` takes the reduced path for a JPEG over `MaxDecodePixels`; any reduced-path error wraps `ErrTooLarge` | STEP-5, AC-1.1, AC-1.5 |
| `raster.MaxDecodeBytes` (128 MiB) checked first in every decode; `derivatives` uses it in place of `maxSourceBytes` | STEP-1, AC-3.1 – AC-3.3 |
| Fixtures from the Go distribution in `core/raster/testdata/` *(new)*, with Go's licence | STEP-6 |
| Go native fuzzing over the decoder, seeds run by `go test` | STEP-6, AC-2.4 |
| A derivatives test that stores the thumbnail row for an oversized JPEG | STEP-7, AC-1.2 |
| Docs: `completed.md`, the map row, the design note, `source-layer-data-model.md`, the security lens row | STEP-8 |

| Out | Home |
| --- | --- |
| EXIF orientation | Not planned ([design § Not this design](../oversized-jpeg-thumbnails.md#not-this-design)) |
| A larger cover derivative | [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) thumbnail work |
| PNG, GIF, BMP, TIFF, WebP over the budget | Not planned (design) |
| The Mac viewer opening the original | Done in S14-13 |
| Persisting a refusal so `EnsureThumbnail` stops re-reading an unprocessable object on every Artifact response | Not planned; this PR removes the common case (oversized JPEGs) |
| A small JPEG the standard library cannot read (arithmetic coding, 12-bit, a truncated file) makes `generate` return the raw decode error, which `thumbnailRelPath` passes up, so the Artifact response fails | Dogfood log; found while planning (2026-10-31), not this PR's behavior to change |
| Moving the decode outside `catalogsession.Do` | S14-12 |
| Drawn-size decode, cache bound by bytes | [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Pixel budget | `core/raster/raster.go` `MaxDecodePixels` (64,000,000), `ErrTooLarge` ("raster: image dimensions exceed decode budget") | The only bound. No byte limit in the package |
| Decode | `core/raster/raster.go` `decode(src, formats)` | `image.DecodeConfig`, format allowlist, then `width × height > MaxDecodePixels` → `ErrTooLarge`, then `image.Decode`. Wrong for this PR: every JPEG over budget is refused |
| Encode entry points | `core/raster/raster.go` `EncodeJPEG`, `EncodePNG`, `prepare(src, opts)` | Both call `prepare` → `decode` → `scaleToMaxEdge` → transforms. `EncodeJPEG` serves thumbnails; `EncodePNG` serves staged images |
| Scale | `core/raster/raster.go` `scaleToMaxEdge` | Nearest-neighbour; returns the source unchanged when its long edge ≤ `MaxEdge` |
| Slot | `core/raster/raster.go` `Slot`, `decodeSem` | Callers hold one slot across read, decode, encode. `raster` never acquires one itself |
| Thumbnail generation | `core/derivatives/ensure.go` `generate` | Holds `raster.Slot()`, reads the object, calls `raster.EncodeJPEG`; maps `raster.ErrTooLarge` and `raster.ErrUnsupportedFormat` to `ErrUnprocessable`; passes any other error up |
| Derivatives byte cap | `core/derivatives/ensure.go` `maxSourceBytes` (128 MiB) | Checked on `files.byte_size` in `Ensure` and on the bytes read in `readSourceObject` |
| Staging | `core/ingest/stage.go` `StageImage` | Own 32 MiB cap (`imageMaxBytes`), holds a slot, calls `raster.EncodePNG` with `Formats: stageFormats`; maps any error to `ErrUnsupportedImage` |
| Handler | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath` | `derivatives.ErrUnprocessable` → empty path, `skipped`. So today an oversized JPEG shows the glyph and nothing is stored |
| Cover | `core/database/sources/sources.go` (cover set, ~line 487) | Calls `derivatives.EnsureThumbnail`; `ErrUnprocessable` → `ErrInvalid`. So "Use as thumbnail" fails for an oversized JPEG today |
| Standard library | `$(go env GOROOT)/src/image/jpeg/reader.go` `DecodeConfig`, `processSOF`, `isRGB` | Go 1.27. `DecodeConfig` reads markers up to SOF only; refuses precision ≠ 8, a component count other than 1, 3, 4, a sampling factor of 3, and SOF markers other than C0–C2 (as "unknown marker") **before** `decode` checks the budget. So arithmetic, lossless and 12-bit JPEGs never reach the reduced path |
| Tests | `core/raster/raster_test.go` `TestEncodeJPEGRejects`, `TestDecode`, `TestEncodePNG`; `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources` | Bombs, allowlist, sizes. `TestDecode` calls `decode(src, formats)` directly; the "declared ByteSize over source cap" case names `maxSourceBytes` |
| Swift mirror | `api/ffi/swift_vocabulary_test.go` row "viewer decode pixels" | Pins `ArtifactImageDecode.maxDecodePixels` to `raster.MaxDecodePixels`. Unchanged by this PR |
| Fuzzing | — | No `Fuzz…` target in the repo yet |

**Precedent.** `57123061` ("Share one bounded image decoder between thumbnails and staged images.") moved the codec to `core/raster` and added the bomb tests: copy its table shape for the new refusal cases and its header-only fixture style (`pngHeaderOnly`) for the crafted SOF. `core/ingest/file.go` `MaxBytes` / `var maxBytes = MaxBytes` (overridden in `file_test.go`) is the precedent for a byte limit a test lowers. There is no precedent for a hand-written codec or a fuzz target; this plan specifies both.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The decoder lives in `core/raster/jpegdc.go` *(new)* as the unexported `decodeJPEGDC(src []byte, maxPixels int64) (image.Image, error)`. Nothing outside `raster` calls it | Working rule names `core/raster`; the package is "the one bounded codec". Unexported keeps one entry (`EncodeJPEG`) and lets tests and the fuzz target pass a small budget | A new package `core/raster/jpegdc` — a second codec package for one caller; an exported func — a second public decode path that skips the allowlist |
| D-2 | `decode` takes a third parameter, `reduceJPEG bool`. When `true`, the format is `"jpeg"` and the declared size is over `MaxDecodePixels`, it returns `decodeJPEGDC(src, MaxDecodePixels)`. `prepare` takes the same flag; `EncodeJPEG` passes `true`, `EncodePNG` passes `false` | The design keeps staging off the reduced path. The flag sits where the over-budget branch already is | A new `Options` field — callers outside `raster` would choose a decode strategy; trying the reduced path after `ErrTooLarge` in `EncodeJPEG` — needs the format a second time |
| D-3 | Any error from `decodeJPEGDC` reaches the caller as `fmt.Errorf("%w: %w", ErrTooLarge, err)` | An oversized JPEG the decoder cannot read stays exactly as today (refused as too large → `ErrUnprocessable` → the glyph), so no Artifact response starts failing. `derivatives` needs no change to its error mapping | A new `raster` error — `derivatives`, `sources` and the handler would each need a new case; a plain error — `thumbnailRelPath` passes it up and the Artifact response fails (the Out bug, made worse) |
| D-4 | `raster.MaxDecodeBytes` *(new)* = 128 MiB, with `var maxDecodeBytes = MaxDecodeBytes` *(new)* that tests lower. `decode` refuses `len(src) > maxDecodeBytes` with `ErrTooLarge` as its first line. `ErrTooLarge`'s message becomes "raster: image exceeds decode budget" and its doc names bytes and pixels | SEC-3: the codec bounds itself. 128 MiB is the largest caller cap today (`derivatives`); `ingest` keeps its stricter 32 MiB and its own user-facing error. The `var` follows `ingest.maxBytes` so a test does not allocate 128 MiB | A separate `ErrTooManyBytes` — every caller already maps `ErrTooLarge`; an `Options.MaxBytes` — the point is that callers cannot widen it |
| D-5 | `derivatives` deletes `maxSourceBytes` and uses `raster.MaxDecodeBytes` in `Ensure` and `readSourceObject` | One number for "how many bytes a decode may see"; a copy can drift above the codec's | Keeping both constants equal by comment |
| D-6 | Markers read: SOI, SOF0, SOF1, SOF2, DHT, DQT, DRI, SOS, EOI, APP0 (JFIF id only), APP14 (Adobe transform only). Skipped by length: APP1–APP13, APP15, COM. Any other marker, including RST outside a scan, DNL, DAC and SOF3/5–7/9–11/13–15, is an error. Fill bytes (`0xFF` repeated) before a marker are allowed | T.81 Table B.1; APP0 and APP14 decide RGB vs YCbCr exactly as the standard library's `isRGB` does, so colors match the full decode | Skipping APP0/APP14 as the design note says — an Adobe RGB JPEG would come out with wrong colors |
| D-7 | SOF checks: P = 8; Nf ∈ {1, 3} (4 → error); each H, V in 1–4; Tq ≤ 3; component ids unique; X, Y ≥ 1 (already true: `decode` refused 0). The reduced size is W8 = ⌈X/8⌉, H8 = ⌈Y/8⌉; `W8 × H8 > maxPixels` → error, before any buffer is allocated | Bounds every count before use. 65,535² is the largest declarable frame, so W8 × H8 ≤ 67.1 M fits in `int64` arithmetic | — |
| D-8 | DHT: Tc ≤ 1, Th ≤ 3; the 16 code counts sum to at most 256 and the segment holds that many values; the canonical code is checked not to overflow at any length (Annex C); a DC table value over 11 is an error. Decoding uses a 9-bit lookahead table plus a bit-by-bit slow path, written for this package, not copied from the standard library | Malformed tables are the classic decoder crash. The lookahead keeps a 39 MB baseline scan in the low seconds | Copying the stdlib's `$(go env GOROOT)/src/image/jpeg/huffman.go` — Go's BSD licence would then cover code in this repo, and its decoder is shaped for full IDCT |
| D-9 | DQT: Pq 0 (8-bit) or 1 (16-bit) entries, Tq ≤ 3; only entry 0 (the DC step) is kept. A component's step is latched at the first scan that includes it; a scan naming a missing table is an error | Only DC is used. Latching at the first scan is libjpeg's rule, so a later DQT cannot change already-decoded blocks | Reading the table at the end of the file |
| D-10 | SOS checks: 1 ≤ Ns ≤ Nf, every Cs in the frame and not repeated, Td and Ta ≤ 3 and the table defined when the scan needs it. Sequential: Ss = 0, Se = 63, Ah = Al = 0. Progressive DC scan: Ss = Se = 0, Al ≤ 13. Progressive AC scan: 1 ≤ Ss ≤ Se ≤ 63, Ns = 1. Anything else is an error | T.81 B.2.3 and G.1.1.1.1 | — |
| D-11 | Block layout (T.81 A.2): Hmax, Vmax over the frame; MCUs across = ⌈X / (8·Hmax)⌉, down = ⌈Y / (8·Vmax)⌉. Each component stores one DC value per block in an `[]int16` of (MCUs across · Hc) × (MCUs down · Vc). An interleaved scan (Ns > 1) walks MCUs, each holding Vc rows × Hc columns of blocks per component in scan order. A single-component scan walks that component's ⌈⌈X·Hc/Hmax⌉/8⌉ × ⌈⌈Y·Vc/Vmax⌉/8⌉ blocks in raster order, one block per MCU. A one-component frame is always single-component (so a gray image tagged 2×2 has ⌈X/8⌉ × ⌈Y/8⌉ blocks) | This is the shape every scan type and every fixture needs; the gray 2×2 and separate-DC fixtures exist to hold it | Storing per scan — progressive DC refinement updates blocks a previous scan wrote |
| D-12 | Before allocating the coefficient buffers (at the first SOS), the decoder checks that the bytes after that SOS header, times 8, are at least the number of blocks that scan covers; short → error | Each block of a first DC scan costs at least one bit, so a header declaring a 56 MP reduced image with no data is refused without a 113 MB allocation (AC-1.3). Later scans run on buffers already sized and fail on EOF | Allocating lazily per block row — complicates every scan walker for no extra safety |
| D-13 | Sequential scans decode every block: the DC difference (category ≤ 11) adds to the component's predictor; the AC symbols are decoded and discarded (ZRL = 16 zeros, EOB ends the block; position past 63 → error; AC category ≤ 10). Progressive first DC scan (Ah = 0): DC = predictor sum `<< Al`. DC refinement (Ah > 0): one bit per block, OR'd in at bit Al. AC scans are skipped: the decoder advances to the next marker that is neither a stuffed `0xFF00` nor RST0–7, without decoding | Design § What it reads. Skipping AC scans is what makes progressive cheap | Decoding progressive AC bands — no output uses them |
| D-14 | The predictor is an `int32`; the stored DC saturates to the `int16` range | A crafted run of differences cannot overflow; real 8-bit DC values stay within ±2,047 and are unaffected | Refusing out-of-range DC — rejects a file over a value no viewer would refuse |
| D-15 | Restart intervals: with DRI Ri > 0, after every Ri MCUs of a scan (except at its end) the next marker must be RST(n mod 8), counting from RST0 in each scan; the bit reader and predictors reset. A missing or out-of-order RST is an error | T.81 E.1.4. The design says a malformed stream is an error, never a partial image | libjpeg-style resynchronisation — produces partial images |
| D-16 | The stream ends at EOI. EOF before EOI, a marker inside entropy data other than RSTn where one is due, or a component that no first DC scan covered, is an error. Bytes after EOI are ignored | Truncation must never yield a partial picture (AC-2.3) | Accepting EOF as EOI — a half-written file would thumbnail as half grey |
| D-17 | Output: sample = clamp(round(DC × Q0 / 8) + 128, 0, 255) per block, with round half away from zero. A pixel (x, y) of the W8 × H8 image takes, from each component, the block at (x·Hc/Hmax, y·Vc/Vmax) (integer division). Nf = 1 → `*image.Gray`. Nf = 3 → `*image.RGBA` (A = 255): when the stdlib's `isRGB` rule holds (no JFIF APP0, and either Adobe transform 0 or component ids 'R','G','B') the three samples are R, G, B; otherwise `color.YCbCrToRGB` | The DC term over 8 is the block's mean before the level shift. Matching `isRGB` and `YCbCrToRGB` keeps colors as the full decode gives them | `*image.YCbCr` — needs one of the standard subsample ratios, which "any sampling factors" (2×2/1×2/1×2) does not fit |
| D-18 | Fixtures: copy these files from `$(go env GOROOT)/src/image/testdata/` into `core/raster/testdata/` *(new)*: `video-001.jpeg`, `video-001.progressive.jpeg`, `video-001.progressive.truncated.jpeg`, `video-001.q50.{410,411,420,422,440,444}.jpeg`, `video-001.q50.{420,444}.progressive.jpeg`, `video-001.221212.jpeg`, `video-001.restart2.jpeg`, `video-001.rgb.jpeg`, `video-001.cmyk.jpeg`, `video-001.separate.dc.progression.progressive.jpeg`, `video-005.gray.jpeg`, `video-005.gray.q50.2x2.jpeg`, `video-005.gray.q50.2x2.progressive.jpeg` (about 110 KB). Add `core/raster/testdata/LICENSE-go` *(new)* (a copy of `$(go env GOROOT)/LICENSE`) and `core/raster/testdata/README.md` *(new)* (one paragraph: where they came from, which Go version, which licence) | The standard library cannot encode progressive, RGB or odd-sampling JPEGs; these files already exercise each case the scope names (verified at planning: DC refinement, separate DC scans, DRI = 20, Adobe t=0 with ids R/G/B, gray tagged 2×2, Cb/Cr 1×2 under Y 2×2) | Reading them from `GOROOT` at test time — CI images and future Go versions may move or drop them |
| D-19 | Fuzz target `FuzzDecodeJPEGDC` *(new)* in `core/raster/jpegdc_test.go` calls `decodeJPEGDC(data, 1<<20)`. Seeds: every file in D-18, plus `jpeg.Encode` output of a 37 × 21 RGBA gradient at quality 10 and 90 and a 16 × 16 gray image. Invariants: no panic; on success the bounds are ⌈W/8⌉ × ⌈H/8⌉ for the SOF the input declares and their product ≤ 1<<20 | `go test` runs the seeds in CI with no extra command; a 1 MP budget keeps each fuzz input fast | Fuzzing `EncodeJPEG` — only inputs declaring over 64 MP reach the reduced path |
| D-20 | The oversized fixtures are built in each test from a lazy `image.Image` whose `At` computes the pixel (no full buffer), encoded with `jpeg.Encode` at quality 50: 8,008 × 8,000 `color.YCbCr` in `raster` (about 2 s, 1 MB), 8,008 × 8,000 `color.Gray` in `derivatives` (about 1 s). Measured at planning on Go 1.27 | The scope asks for a just-over-64 MP JPEG with no full buffer; 8,008 is a multiple of 8, so the expected sizes are exact | A committed 1 MB fixture — binary churn for a file the test can make |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/raster/jpegdc.go` `decodeJPEGDC` *(new)* | file, func | `image/jpeg` `Decode` (stdlib) | Its reader is unexported and always runs the full IDCT into a full-size buffer; searched for "DC only", "reduced", "scale" decoders in `golang.org/x/image` (none for JPEG) |
| Huffman table and bit reader types in `jpegdc.go` *(new, unexported)* | type | `$(go env GOROOT)/src/image/jpeg/huffman.go` (stdlib, unexported) | Not importable; D-8 |
| `raster.MaxDecodeBytes`, `maxDecodeBytes` *(new)* | const, var | `derivatives.maxSourceBytes`, `ingest.MaxBytes` / `maxBytes` | Replaces `maxSourceBytes` (D-5); `ingest`'s caps are stricter and user-facing, so they stay |
| `decode` / `prepare` parameter `reduceJPEG` *(new)* | param | `Options.Formats` | `Formats` limits what may decode; this chooses how an over-budget JPEG decodes, which only `raster` decides (D-2) |
| `core/raster/testdata/` *(new)* | fixtures | `core/derivatives/testdata/tiny.png` | Different package; the files are the stdlib's |
| `core/raster/jpegdc_test.go` `TestDecodeJPEGDC…`, `FuzzDecodeJPEGDC` *(new)* | tests | `raster_test.go` tables | Same package, new file for the new codec |
| `lazyImage` test type *(new, in `raster_test.go` and `thumbnail_test.go`)* | test helper | `encodeImage` in `raster_test.go` (allocates a full RGBA) | A full buffer at 64 MP is 256 MB; the two packages cannot share a test helper |

**Builds on.** `core/raster`'s single decode gate (`decode`), its `ErrTooLarge` contract, which every caller already maps, and the caller-held `Slot`. The `ingest` test-lowered byte cap.
**Deletes.** `derivatives.maxSourceBytes` (D-5). The sentence "every JPEG over the budget is refused" is not written anywhere else; `docs/source-layer-data-model.md` § thumbnails gains the reduced path (STEP-8). No callee loses its last caller.
**Next adds.** S14-12's staged-object type calls `EncodeJPEG` under its own slot and gets the reduced path with no change. A drawn-size decode is a smaller `MaxEdge`, unchanged here. A future Windows client calls the same Go.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An oversized JPEG now gets a `file_derivatives` row | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `existingThumbnailRelPath`; `core/database/sources/sources.go` cover set; `derivatives.reprojectSourcesForNewThumbnail` | The same row shape as any JPEG thumbnail: 256 px JPEG, `TypeThumbnail`. Existing catalogs have no refusal stored, so the next Artifact response creates it; the Source cover and omnibar pick it up through the existing reprojection |

### Scale

At the map's Scale (3,000 Sources) the cost is per oversized file, once: the row is stored and later reads are lookups. Per file: a 198 MP, 39 MB baseline scan has 3.1 M luma blocks; Huffman work is linear in the 39 MB of scan data. Memory peak at that size: `[]int16` DC buffers (3.1 M + 2 × 0.8 M blocks at 4:2:0 ≈ 9 MB) plus the 1,375 × 2,250 RGBA output (12 MB) plus the 39 MB input. Worst case the budget allows: a reduced image of 64 M pixels at 4:4:4 is 3 × 128 MB of DC buffers plus 256 MB RGBA ≈ 640 MB, below the standard library's own worst case at the same 64 MP budget for a progressive JPEG (64 `int32` coefficients per block). Concurrency stays bounded by the caller's `Slot` (at most 4). Time budget: `TestEncodeJPEGOversized` (64 MP, 1 M blocks) stays under 10 s on CI including the fixture encode; Stop condition if not.

### Edges

- **Assumptions.** `scaleToMaxEdge` assumes bounds starting at (0, 0) or uses `b.Min` — the reduced image starts at (0, 0) (T-1 checks `Bounds().Min`). `decode` assumes `image.DecodeConfig` already proved the bytes are a JPEG with an SOF the stdlib accepts; the reduced decoder does not rely on it and re-parses every marker (fuzz T-9 calls it directly). Callers assume `ErrTooLarge` means "refuse quietly" — D-3 keeps that for every reduced-path failure (T-6).
- **Neighbors.** 8,000 × 8,000 (exactly 64 MP) keeps the standard library (existing `TestEncodeJPEGAtBudgetBoundaries` plus T-4 case "at budget"); 8,008 × 8,000 takes the reduced path (T-4). PNG at 8,001 × 8,000 stays refused (existing). An oversized JPEG through `EncodePNG` stays refused (T-5).
- **Removed inputs.** No input is decided or deleted; a thumbnail row that a later delete removes is handled by the existing derivative release.
- **Real data.** No table or seed drives the rule. The planning run over the 26 stdlib JPEG fixtures (marker dump, 2026-10-31) found every scan shape listed in D-18; none uses SOF1 or multiple sequential scans, so T-2 adds a hand-built SOF1 case by patching `video-001.jpeg`'s SOF0 byte to 0xC1.
- **Untrusted size.** Frame: W8 × H8 ≤ `maxPixels` before allocation (D-7). Coefficient buffers: at most 3 components, each ≤ (W8 + 4) × (H8 + 4) blocks because H, V ≤ 4. First scan data ≥ blocks / 8 bytes before allocation (D-12). Huffman counts ≤ 256 and segment-length checked (D-8). Scan count: not capped; every DC scan costs at least one bit per block and AC scans are skipped linearly, so total work is linear in input bytes, which are ≤ `MaxDecodeBytes` (D-4). Marker segments: each length is checked against the remaining bytes before it is read.

### Schema

None — no catalog change. Existing catalogs need nothing: a refusal was never stored.

### Go core

| Package | Change |
| --- | --- |
| `core/raster` | `const MaxDecodeBytes = 128 << 20` *(new)*; `var maxDecodeBytes = MaxDecodeBytes` *(new)*; `ErrTooLarge` message and doc (D-4) |
| `core/raster` | `func decode(src []byte, formats []string, reduceJPEG bool) (image.Image, error)` — byte check first; over-budget JPEG with `reduceJPEG` → `decodeJPEGDC` wrapped per D-3 |
| `core/raster` | `func prepare(src []byte, opts Options, reduceJPEG bool)`; `EncodeJPEG` passes `true`, `EncodePNG` `false`. Doc comments say which path each takes |
| `core/raster/jpegdc.go` *(new)* | `func decodeJPEGDC(src []byte, maxPixels int64) (image.Image, error)` and its unexported marker parser, Huffman table, bit reader, scan walkers (D-6 – D-17). Errors are unexported `errors.New("raster: jpeg …")` values naming the failed check |
| `core/derivatives` | `maxSourceBytes` → `raster.MaxDecodeBytes` (D-5). No change to error mapping |

```go
// decodeJPEGDC decodes a JPEG at 1/8 scale from its DC coefficients only.
// It returns an image of ⌈W/8⌉ × ⌈H/8⌉ pixels with Min at (0, 0), or an
// error and no image. It allocates nothing sized by the frame until the
// frame is within maxPixels and the first scan has enough bytes (D-12).
func decodeJPEGDC(src []byte, maxPixels int64) (image.Image, error)
```

Errors: `raster.ErrTooLarge` (wrapping the decoder's own error) → `derivatives.ErrUnprocessable` → empty thumbnail path, as today.

### FFI

None — no proto or handler change.

### macOS

None — no Swift change. The thumbnail cache decodes the 256 px derivative as for any JPEG.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The byte limit.**
Files: `core/raster/raster.go`, `core/raster/raster_test.go`, `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: add `MaxDecodeBytes`, `maxDecodeBytes`, the byte check at the top of `decode`, the new `ErrTooLarge` message and doc (D-4). Replace `maxSourceBytes` with `raster.MaxDecodeBytes` in `ensure.go` and in the "declared ByteSize over source cap" case (D-5). Add T-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster ./core/derivatives ./core/ingest` passes.

**STEP-2 — Markers, frame, tables.**
Files: `core/raster/jpegdc.go` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: the marker loop, SOF, DHT, DQT, DRI, APP0, APP14, SOS header parsing and every check in D-6 – D-10, D-7's budget check, D-12's data check, buffer allocation per D-11. No entropy decoding yet: return an error "not implemented" after the first SOS. Add T-2's refusal cases that need no scan decoding (CMYK, SOF1 patched to 0xC9, 65,535 × 65,535, AC-1.3's header, a DHT with counts summing to 257, a DQT with Tq = 4).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestDecodeJPEGDC' ./core/raster` passes for those cases.

**STEP-3 — Huffman and the bit reader.**
Files: `core/raster/jpegdc.go` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: canonical code build with the overflow check, 9-bit lookahead plus slow path, the byte-stuffing-aware bit reader that stops at a marker (D-8, D-16). Unit-test decoding a table from `video-001.jpeg`'s first DHT against a hand-written bit string.
Check: `go test -tags fts5 -run 'TestJPEGDCHuffman' ./core/raster` passes.

**STEP-4 — Scans and output.**
Files: `core/raster/jpegdc.go` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: sequential scans (D-13), progressive DC first and refinement scans and AC skipping (D-13), restart handling (D-15), EOI and coverage checks (D-16), saturation (D-14), output per D-17. Add T-1, T-3 and the rest of T-2.
Check: `go test -tags fts5 -run 'TestDecodeJPEGDC' ./core/raster` passes, every fixture within the AC-2.1 tolerance.

**STEP-5 — Wire into `EncodeJPEG`.**
Files: `core/raster/raster.go`, `core/raster/raster_test.go`.
Do: the `reduceJPEG` parameter on `decode` and `prepare` (D-2), the wrap (D-3), doc comments on `EncodeJPEG` and `EncodePNG`. Update `TestDecode` calls to pass `false` and add a `true` row. Add T-4, T-5, T-6.
Check: `go test -tags fts5 ./core/raster` passes; `TestEncodeJPEGOversized` runs under 10 s (`go test -v -run TestEncodeJPEGOversized ./core/raster` prints its time).

**STEP-6 — Fixtures and fuzzing.**
Files: `core/raster/testdata/` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: copy the D-18 files, `LICENSE-go` and the README (T-1 reads them from `core/raster/testdata/` *(new)*, so STEP-4 copies the fixtures it needs first and this step adds the rest). Add `FuzzDecodeJPEGDC` (D-19). Then run the fuzzer for 5 minutes; every crasher it writes under `core/raster/testdata/fuzz/FuzzDecodeJPEGDC/` *(new)* is fixed in `jpegdc.go` and its file is committed as a regression seed.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -fuzz FuzzDecodeJPEGDC -fuzztime 5m ./core/raster` ends with no failure; `go test -tags fts5 ./core/raster` runs the seeds and passes.

**STEP-7 — The derivative row.**
Files: `core/derivatives/thumbnail_test.go`.
Do: add T-8 with its own lazy gray image (D-20). No production change in `derivatives` beyond STEP-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/derivatives` passes.

**STEP-8 — Docs and version.**
Files: `docs/deployment-plan/spike-14/completed.md`; `docs/deployment-plan/spike-14/deployment-plan.md` (map row Status → Landed); this plan (Status, Landed note); `docs/deployment-plan/spike-14/oversized-jpeg-thumbnails.md` (§ What it reads: APP0 and APP14 are read for the color rule, D-6; add "Landed in #N"); `docs/source-layer-data-model.md` (§ 8 `file_derivatives`, one paragraph after the sentence beginning "Derivative Files use the same content-addressed": a thumbnail is decoded by `core/raster` within 128 MiB and 64 MP; a JPEG over 64 MP is decoded at 1/8 scale from its DC coefficients; other formats over the budget get no thumbnail); `.agents/skills/review-pr/lenses/security.md` (the `core/raster` row: "the one bounded codec, also used by staging, including a hand-written reduced JPEG decoder (fuzzed)"); `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj` → 0.2.3.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-23-oversized-jpeg-thumbnails.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a thumbnail for a JPEG scan over 64 megapixels."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/raster/jpegdc_test.go` `TestDecodeJPEGDCMatchesStdlib` *(new)* | Each decodable D-18 fixture → bounds (0,0)–(19,13); mean abs diff per channel ≤ 3.0 against `jpeg.Decode` averaged over each 8 × 8 cell (in-bounds pixels only); Gray fixtures return `*image.Gray`, colour fixtures `*image.RGBA` | AC-2.1, D-11, D-17 |
| T-2 | Go | `core/raster/jpegdc_test.go` `TestDecodeJPEGDCRefuses` *(new)* | `video-001.cmyk.jpeg`; SOF byte patched to 0xC9 (arithmetic); SOF1 patch of `video-001.jpeg` decodes (not refused); P = 12; Nf = 2; H = 5; DHT counts sum 257; DHT DC value 12; DQT Tq = 4; SOS naming a missing component; progressive AC scan with Ns = 2; restart marker RST1 where RST0 is due → error, no image (SOF1 case → success) | AC-2.2, D-6 – D-10, D-15 |
| T-3 | Go | `core/raster/jpegdc_test.go` `TestDecodeJPEGDCTruncated` *(new)* | `video-001.progressive.truncated.jpeg`; `video-001.jpeg` cut at half; `video-001.jpeg` with its last two bytes (EOI) removed → error, no image | AC-2.3, D-16 |
| T-4 | Go | `core/raster/raster_test.go` `TestEncodeJPEGOversized` *(new)* | Lazy 8,008 × 8,000 YCbCr JPEG, `MaxEdge: 256` → JPEG decoding to 256 × 255; "at budget": 8,000 × 8,000 header only (no scan data) through `decode(…, true)` takes the stdlib path (error is not wrapped from `decodeJPEGDC`: `errors.Is(err, ErrTooLarge)` is false) | AC-1.1, D-2 |
| T-5 | Go | `core/raster/raster_test.go` `TestEncodePNGNeverReduces` *(new)* | The T-4 JPEG through `EncodePNG(…, Options{MaxEdge: 2048})` → `ErrTooLarge` | AC-1.5, D-2 |
| T-6 | Go | `core/raster/raster_test.go` `TestEncodeJPEGRejects` *(extend)* | New rows: SOF 65,535 × 65,535 → `ErrTooLarge`; SOF 60,000 × 60,000 + SOS + 16 bytes → `ErrTooLarge` and `runtime.MemStats.TotalAlloc` grows < 16 MiB across the call; T-4's JPEG cut at half → `ErrTooLarge`. Existing GIF and PNG bombs unchanged | AC-1.3, AC-1.4, AC-1.6, D-3, D-12 |
| T-7 | Go | `core/raster/raster_test.go` `TestDecodeByteLimit` *(new)* | `maxDecodeBytes = 64` (restored by `t.Cleanup`); 65 bytes of a registered format `"count"` whose config func increments a counter → `ErrTooLarge` from `EncodeJPEG` and `EncodePNG`, counter 0; 64 bytes → counter 1 | AC-3.1, AC-3.2, D-4 |
| T-8 | Go | `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailOversizedJPEG` *(new)* | `installSourceFile` with a lazy 8,008 × 8,000 gray JPEG (`image/jpeg`) → `EnsureThumbnail` returns a link; derived File `image/jpeg`; its object decodes to 256 × 255; second call → same `Link.ID` | AC-1.2 |
| T-9 | Go | `core/raster/jpegdc_test.go` `FuzzDecodeJPEGDC` *(new)* | Seeds per D-19 → no panic; success implies bounds = ⌈W/8⌉ × ⌈H/8⌉ (W, H read from the input's SOF by the test) and ≤ 1<<20 pixels | AC-2.4, D-19 |
| T-10 | Go | `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources` *(edit)* | "declared ByteSize over source cap" uses `raster.MaxDecodeBytes + 1` → `ErrUnprocessable` | AC-3.3, D-5 |
| T-11 | Go | `core/raster/jpegdc_test.go` `TestJPEGDCHuffman` *(new)* | `video-001.jpeg`'s first DHT built; a hand-written bit string → the expected symbols; a table whose counts overflow the code space → error | D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/raster` `TestDecode` (signature change), `core/derivatives` `TestEnsureThumbnailRejectsUnsafeSources` (constant rename), `core/ingest` `TestStageImage…` (must be unaffected; it uses `EncodePNG`).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] The 5-minute fuzz run in STEP-6 ended with no failure; any crasher it found is a committed seed
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: `maxSourceBytes` and "image dimensions exceed decode budget" `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump 0.2.2 → 0.2.3 (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: docs/source-layer-data-model.md § 8 (reduced path, byte limit); oversized-jpeg-thumbnails.md (APP0/APP14, landed); security lens row
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Build the app from the branch. In a test project, attach a JPEG over 64 MP to a Source (to make one: `sips -z 18000 11000 any.jpg --out big.jpg` on macOS, then check `sips -g pixelWidth -g pixelHeight big.jpg`). Open the Source. The Artifact tile shows the scan's thumbnail, not the missing-image glyph.
2. Choose "Use as thumbnail" on that Artifact. The Source header and the Sources list row show it as the cover.
3. Type the Source's title in the omnibar. The hit shows the thumbnail.
4. Attach a JPEG under 64 MP and a PNG. Both thumbnail as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A hand-written decoder reads untrusted bytes | Every count checked before use (Edges), fuzzing (T-9), errors never yield partial images (D-16), the caller's slot bounds concurrency |
| Colour or rounding differs from the full decode | T-1 tolerance against the stdlib on every sampling and colour rule |
| A real camera or scanner file uses a shape the fixtures miss (multiple sequential scans, SOF1) | D-11 and D-13 treat scans uniformly; T-2's SOF1 case; a real file that fails stays refused as today (D-3), never worse |
| The oversized tests slow CI | 1 MB fixtures built in about 2 s; Stop condition at 10 s |

**Stop and report if:**

- `core/raster/raster.go` `decode` no longer has the signature `decode(src []byte, formats []string)` or no longer runs `image.DecodeConfig` before the budget check.
- A fixture in D-18 cannot reach a mean abs diff ≤ 3.0 per channel (T-1) with D-17's formula: report the fixture and the measured value; do not loosen the tolerance.
- `TestEncodeJPEGOversized` takes over 10 s, or T-6's allocation check needs more than 16 MiB: report the measurement.
- The 5-minute fuzz run finds a crasher whose fix needs a rule not in D-6 – D-17.
- Any `ingest` or `derivatives` test needs a change beyond T-8 and T-10.
- A fixture file in D-18 is missing from the Go 1.27 distribution.

## 12. Open questions

None.

## 13. PR

**Title:** S14-23 — Thumbnails for oversized JPEGs
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names, the fuzz run's duration and result, `TestEncodeJPEGOversized` timing); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-23-oversized-jpeg-thumbnails.md`.
