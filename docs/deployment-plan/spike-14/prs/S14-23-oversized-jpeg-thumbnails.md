# S14-23 — Thumbnails for oversized JPEGs

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 1 — Event and Place pictures
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a thumbnail that was missing now appears in the existing tile, header and cover slots; no control, copy or layout changes)
**Board:** —
**Depends on:** — (the map prefers S14-12 first; this plan goes first because S14-12 is not planned yet, and it leaves S14-12 one byte limit to read, `raster.MaxDecodeBytes`, recorded in S14-12's scope entry)
**Blocks:** —
**Branch:** `s14-23-oversized-jpeg-thumbnails`
**Version bump:** patch, 0.2.2 → 0.2.3 — a missing thumbnail is a bugfix with no new screen or RPC ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** — (no skill covers `core/raster`)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [design](../oversized-jpeg-thumbnails.md); working rule "(Batch 1) Oversized JPEGs get a thumbnail from a reduced decode of our own"; map scope entry [S14-23](../deployment-plan.md#s14-23--thumbnails-for-oversized-jpegs); likeness audit SEC-3 (a raster byte limit, [`docs/ideas/depictions-and-likenesses.md`](../../../ideas/depictions-and-likenesses.md)); bar outcome 12; ITU-T T.81 (Annex B syntax, Annex C and F.2.2 Huffman decoding, G.1.1.1.1 and G.1.2 progressive DC scans, A.2 block order, E.1.4 restart)
**Planned:** 2026-10-31 against `main` @ `b6a8f3e`
**Review:** [`S14-23-oversized-jpeg-thumbnails.review.md`](S14-23-oversized-jpeg-thumbnails.review.md) — round 1 answered, round 2 running

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

A JPEG scan over 64 megapixels (the dogfood file is 11,000 × 18,000, 39 MB) saves as an Artifact and then shows the missing-image glyph everywhere, with no error, and cannot be chosen as a Source cover. After this PR Go decodes such a JPEG at 1/8 scale from its DC coefficients only, then scales and encodes the usual 256 px thumbnail, so the tile, the Source cover and the omnibar show it. The decoder is pure Go, so a Windows build reuses it. `core/raster` also refuses any input over a byte limit of its own before it decodes, and has one rule for bytes it cannot decode, so an image Go cannot read shows the glyph instead of failing the Artifact response.

**Why now.** Batch 1's last engine item (bar outcome 12), moved from Spike 12 (S12-39). The design was agreed on 2026-10-27.

**What stays.** Every readable JPEG within 64 MP, and every readable PNG, GIF, BMP, TIFF and WebP, decodes exactly as today through the standard library. Other formats over the budget stay refused. Staged depiction images (`ingest.StageImage`, `EncodePNG`) never take the reduced path. The Mac viewer's own bounded decode (S14-13) is untouched.

## 2. User stories

**US-1.** As a researcher who attaches an oversized JPEG scan, I see its thumbnail and can use it as the Source cover, so that the scan is not a blank tile.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `raster.EncodeJPEG` on a baseline 4:2:0 JPEG of 8,008 × 8,000 (64.06 MP, over budget) with `Options{MaxEdge: 256}` | Returns a JPEG of 256 × 255 |
| AC-1.2 | `derivatives.EnsureThumbnail` on a `files` row (`image/jpeg`) whose object is the AC-1.1 JPEG | Returns a `file_derivatives` link; its derived File is `image/jpeg` and decodes to 256 × 255; a second call returns the same link id |
| AC-1.3 | `rastertest.JPEGHeader` declaring 60,000 × 60,000, one component (1/8 image 56.25 MP, within budget), with 16 bytes of scan data | `EncodeJPEG` returns an error that `errors.Is` `ErrUndecodable`; the call allocates less than 16 MiB in total |
| AC-1.4 | `rastertest.JPEGHeader` declaring 65,535 × 65,535 (1/8 image 8,192 × 8,192 = 67.1 MP, over budget) | `ErrTooLarge` before any coefficient buffer is allocated |
| AC-1.5 | `EncodePNG` (the staging path) on the AC-1.1 JPEG | `ErrTooLarge`, as today |
| AC-1.6 | The existing PNG and GIF header bombs in `TestEncodeJPEGRejects` | Still `ErrTooLarge` |
| AC-1.7 | `EncodeJPEG` with `MaxEdge: 2048` on the AC-1.1 JPEG | 1,001 × 1,000: on the reduced path the output is never larger than ⌈W/8⌉ × ⌈H/8⌉ |

**US-2.** As the engine, the reduced decoder gives the same picture as the standard library would at 1/8 scale, and refuses what it cannot read instead of guessing.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Each decodable fixture in D-18 through `decodeJPEGDC` | Bounds are (0, 0)–(19, 13) (⌈150/8⌉ × ⌈103/8⌉); over the 18 × 12 full 8 × 8 cells, the mean absolute difference per channel from the standard library's decode averaged over each cell is at most 3.0 |
| AC-2.2 | `video-001.cmyk.jpeg` (four components) | Error; no image |
| AC-2.3 | `video-001.progressive.truncated.jpeg`; `video-001.jpeg` cut to half its length; `video-001.jpeg` with its EOI removed | Error; no image |
| AC-2.4 | The fuzz seeds (every fixture, plus small JPEGs from `jpeg.Encode`) through `FuzzDecodeJPEGDC` | No panic; every success has bounds ⌈W/8⌉ × ⌈H/8⌉ and at most the budget the fuzz passes |

**US-3.** As the engine, every raster decode is bounded by bytes as well as pixels, and every decode failure is one refusal.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | With the byte limit set to 64 in the test, `EncodeJPEG` and `EncodePNG` on 65 bytes starting `FAKEIMG!` | `ErrTooLarge`; the fake format's config func is never called |
| AC-3.2 | With the byte limit set to 64, the same on exactly 64 bytes | `ErrUnsupportedFormat` (the fake format is not allowlisted); the config func was called once |
| AC-3.3 | `derivatives` refuses a `files` row declaring `raster.MaxDecodeBytes + 1` bytes | `ErrUnprocessable`, as today with its own constant |
| AC-3.4 | `EncodeJPEG` on `video-001.jpeg` cut to half its length (within budget, standard-library path) | `ErrUndecodable` |
| AC-3.5 | `EnsureThumbnail` on a `files` row (`image/jpeg`) whose object is `video-001.jpeg` cut to half | `ErrUnprocessable` (today: the raw `image/jpeg` error, which fails the Artifact response) |
| AC-3.6 | `EncodeJPEG` on four bytes `00 01 02 03` | `ErrUnsupportedFormat` (today: `image.ErrFormat` returned raw) |

## 3. Scope

| In | Covered by |
| --- | --- |
| The reduced decoder `decodeJPEGDC`: baseline and extended sequential (SOF0, SOF1) and progressive (SOF2) Huffman, 8-bit, one or three components, the sampling factors the standard library accepts, restart intervals, interleaved and single-component scans | STEP-4, STEP-5, STEP-6, AC-2.1 |
| Refusals: four components, and any marker the decoder does not read (arithmetic, lossless, hierarchical) return an error | STEP-4, AC-2.2 |
| Every count read from the file checked before use; the 1/8 image checked against `MaxDecodePixels`; scan data checked before the coefficient buffers are allocated | STEP-4, AC-1.3, AC-1.4 |
| `EncodeJPEG` takes the reduced path for a JPEG over `MaxDecodePixels` | STEP-7, AC-1.1, AC-1.5, AC-1.7 |
| `raster.MaxDecodeBytes` (128 MiB) checked first in every decode; `derivatives` uses it in place of `maxSourceBytes` | STEP-1, AC-3.1 – AC-3.3 |
| One refusal rule in `decode`: `image.ErrFormat` → `ErrUnsupportedFormat`; every other decode failure, on either path → `ErrUndecodable` *(new)*; `derivatives.generate` maps `ErrUndecodable` to `ErrUnprocessable` | STEP-2, AC-3.4 – AC-3.6 |
| `core/raster/rastertest` *(new)*: the shared over-budget fixtures, replacing `pngHeaderOnly` and `bombPNG` | STEP-3 |
| Fixtures from the Go distribution under `core/raster/testdata/third_party/go/` *(new)*, with Go's licence and a notice | STEP-3 |
| Go native fuzzing over the decoder, seeds run by `go test`; a "Fuzz targets" paragraph in `.cursor/rules/go-tests.mdc` | STEP-8, AC-2.4 |
| A derivatives test that stores the thumbnail row for an oversized JPEG | STEP-9, AC-1.2 |
| Docs: `completed.md`, the map row, S14-12's scope entry, the design note, `source-layer-data-model.md`, the security lens | STEP-10 |

| Out | Home |
| --- | --- |
| EXIF orientation | Not planned ([design § Not this design](../oversized-jpeg-thumbnails.md#not-this-design)) |
| A larger cover derivative (and its 1/8 cap on oversized JPEGs, AC-1.7) | [design § Not this design](../oversized-jpeg-thumbnails.md#not-this-design) |
| PNG, GIF, BMP, TIFF, WebP over the budget | Not planned (design) |
| The Mac viewer opening the original | Done in S14-13 |
| Persisting a refusal so `EnsureThumbnail` stops re-reading an unprocessable object on every Artifact response | Not planned; this PR removes the common case (oversized JPEGs) |
| Moving the decode outside `catalogsession.Do` | S14-12 |
| `raster_test.go` `encodeImage` (small full-buffer test images) | Not planned; it stays for the small images it builds |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Package doc | `core/raster/raster.go` lines 1–3 | "decodes only allowlisted formats within a pixel budget" |
| Pixel budget | `core/raster/raster.go` `MaxDecodePixels` (64,000,000), `ErrTooLarge` ("raster: image dimensions exceed decode budget") | The only bound. No byte limit in the package |
| Format refusal | `core/raster/raster.go` `ErrUnsupportedFormat` | Doc says "bytes are not one of the allowlisted raster formats, regardless of the recorded media type", but bytes no decoder recognises come back as raw `image.ErrFormat` |
| Decode | `core/raster/raster.go` `decode(src, formats)` | `image.DecodeConfig` (its error returned raw), format allowlist, then `width × height > MaxDecodePixels` → `ErrTooLarge`, then `image.Decode` (its error returned raw). Wrong for this PR: every JPEG over budget is refused, and a decode failure is not a refusal |
| Encode entry points | `core/raster/raster.go` `EncodeJPEG`, `EncodePNG`, `prepare(src, opts)` | Both call `prepare` → `decode` → `scaleToMaxEdge` → transforms. `EncodeJPEG` serves thumbnails; `EncodePNG` serves staged images |
| Scale | `core/raster/raster.go` `scaleToMaxEdge` | Nearest-neighbour; returns the source unchanged when its long edge ≤ `MaxEdge` |
| Slot | `core/raster/raster.go` `Slot`, `decodeSem` | Callers hold one slot across read, decode, encode. `raster` never acquires one itself |
| Thumbnail generation | `core/derivatives/ensure.go` `generate` | Holds `raster.Slot()`, reads the object, calls `raster.EncodeJPEG`; maps `raster.ErrTooLarge` and `raster.ErrUnsupportedFormat` to `ErrUnprocessable`; passes any other error up. It is the only code that maps raster errors |
| Derivatives byte cap | `core/derivatives/ensure.go` `maxSourceBytes` (128 MiB) | Checked on `files.byte_size` in `Ensure` and on the bytes read in `readSourceObject` |
| Staging | `core/ingest/stage.go` `StageImage` | Own 32 MiB cap (`imageMaxBytes`), holds a slot, calls `raster.EncodePNG` with `Formats: stageFormats`; maps any error to `ErrUnsupportedImage`. `ingest.MaxBytes` (512 MiB) is the import cap and never reaches `raster` |
| Handler | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath` | `derivatives.ErrUnprocessable` → empty path, `skipped`; any other error fails the response. So today an oversized JPEG shows the glyph and nothing is stored, and a truncated small JPEG fails the Artifact response |
| Cover | `core/database/sources/sources.go` `requireRasterCoverArtifact` | Calls `derivatives.EnsureThumbnail`; `ErrUnprocessable` → `ErrInvalid`. So "Use as thumbnail" fails for an oversized JPEG today |
| Standard library | `$(go env GOROOT)/src/image/jpeg/reader.go` `DecodeConfig`, `processSOF`, `isRGB`, `processApp0Marker`, `processApp14Marker` | Go 1.27. `DecodeConfig` reads markers up to SOF only; refuses precision ≠ 8, a component count other than 1, 3, 4, a sampling factor of 3, a 3-component frame where Hmax % Hc ≠ 0 or Vmax % Vc ≠ 0, and SOF markers other than C0–C2, **before** `decode` checks the budget. So arithmetic, lossless and 12-bit JPEGs never reach the reduced path from `EncodeJPEG` |
| Tests | `core/raster/raster_test.go` `TestEncodeJPEGRejects`, `TestEncodeJPEGAtBudgetBoundaries`, `TestDecode`, `TestEncodePNG`, `pngHeaderOnly`, `init` (registers format `"fake"`, magic `FAKEIMG!`); `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources`, `bombPNG` | `TestDecode` calls `decode(src, formats)` directly; `bombPNG` is a copy of `pngHeaderOnly`; the "declared ByteSize over source cap" case names `maxSourceBytes` |
| Shared test packages | `core/database/artifacts/artifactstest`, `core/database/databasetest` | The repo's pattern for helpers several packages' tests import |
| Vendored material | `core/database/namealiases/third_party/nicknames/LICENSE`, `NOTICE` | The repo's shape for upstream-licensed files |
| Swift mirror | `api/ffi/swift_vocabulary_test.go` row "viewer decode pixels" | Pins `ArtifactImageDecode.maxDecodePixels` to `raster.MaxDecodePixels`. Unchanged by this PR |
| Fuzzing | — | No `Fuzz…` target in the repo, and no guidance |

**Precedent.** `57123061` ("Share one bounded image decoder between thumbnails and staged images.") moved the codec to `core/raster` and added the bomb tests: copy its table shape for the new refusal rows. `core/ingest/file.go` `MaxBytes` / `var maxBytes = MaxBytes` (lowered in `file_test.go`) is the precedent for a byte limit a test lowers. `artifactstest` is the precedent for `rastertest`. There is no precedent for a hand-written codec or a fuzz target; this plan specifies both.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The decoder lives in `core/raster/jpegdc.go` *(new)* as the unexported `decodeJPEGDC(src []byte, maxPixels int64) (image.Image, error)`. Nothing outside `raster` calls it | Working rule names `core/raster`; the package is "the one bounded codec". Unexported keeps one entry (`EncodeJPEG`) and lets tests and the fuzz target pass a small budget | A new package — a second codec package for one caller; an exported func — a second public decode path that skips the allowlist |
| D-2 | `decode` takes a third parameter, `reduceJPEG bool`. When `true`, the format is `"jpeg"` and the declared size is over `MaxDecodePixels`, it returns `decodeJPEGDC(src, MaxDecodePixels)`. `prepare` takes the same flag; `EncodeJPEG` passes `true`, `EncodePNG` passes `false`. `EncodeJPEG`'s doc says that on the reduced path the output is at most ⌈W/8⌉ × ⌈H/8⌉ whatever `MaxEdge` asks | The design keeps staging off the reduced path. The flag sits where the over-budget branch already is. The doc sentence is how a later larger derivative learns its cap (AC-1.7) | A new `Options` field — callers outside `raster` would choose a decode strategy |
| D-3 | One refusal rule in `decode`. `ErrUndecodable` *(new, exported)* = "raster: image bytes cannot be decoded". A `DecodeConfig` error that is `image.ErrFormat` → `ErrUnsupportedFormat`; any other `DecodeConfig` error, any `image.Decode` error and any `decodeJPEGDC` error → `fmt.Errorf("%w: %w", ErrUndecodable, err)`. The budget checks keep `ErrTooLarge`. `derivatives.generate` adds `ErrUndecodable` to its `ErrUnprocessable` mapping | Every decode failure, on either path, refuses the same way: the glyph, never a failed Artifact response (AC-3.4 – AC-3.6). `generate` is the only code that maps raster errors, so this is one `errors.Is`. `ingest.StageImage` already maps every raster error | Wrapping reduced-path failures in `ErrTooLarge` only — two refusal rules in one function, and the standard-library path keeps failing the Artifact response; reusing `ErrTooLarge` for decode failures — the name would lie |
| D-4 | `raster.MaxDecodeBytes` *(new)* = 128 MiB, with `var maxDecodeBytes = MaxDecodeBytes` *(new)* that tests lower. `decode` refuses `len(src) > maxDecodeBytes` with `ErrTooLarge` as its first line. `ErrTooLarge`'s message becomes "raster: image exceeds decode budget" and its doc names bytes and pixels; the package doc names the byte limit, the pixel budget and the format allowlist | SEC-3: the codec bounds itself. 128 MiB is `derivatives`' cap today; `StageImage` keeps its stricter 32 MiB and its own user-facing error. The `var` follows `ingest.maxBytes` so a test does not allocate 128 MiB | A separate error for bytes — every caller already maps `ErrTooLarge`; an `Options.MaxBytes` — the point is that callers cannot widen it |
| D-5 | `derivatives` deletes `maxSourceBytes` and uses `raster.MaxDecodeBytes` in `Ensure` and `readSourceObject` | One number for "how many bytes a decode may see"; a copy can drift above the codec's | Keeping both constants equal by comment |
| D-6 | Markers read: SOI, SOF0, SOF1, SOF2, DHT, DQT, DRI, SOS, EOI; APP0 is JFIF when its length field is ≥ 7 and its payload starts `JFIF\x00`; APP14 is Adobe when its length field is ≥ 14 and its payload starts `Adobe`, and its transform is payload byte 11. Skipped by length: other APPn, COM. Any other marker, including RST outside a scan, DNL, DAC and SOF3/5–7/9–11/13–15, is an error. Every segment's length is ≥ 2 and fits in the remaining bytes, checked before reading it. Fill bytes (`0xFF` repeated) before a marker are allowed | T.81 Table B.1; APP0 and APP14 decide RGB vs YCbCr as the standard library's `processApp0Marker`, `processApp14Marker` and `isRGB` do | Skipping APP0/APP14 as the design note says — an Adobe RGB JPEG would come out with wrong colors |
| D-7 | SOF checks: P = 8; Nf ∈ {1, 3} (4 → error); every H and V ∈ {1, 2, 4}; for Nf = 3, Hmax % Hc = 0 and Vmax % Vc = 0 for every component; for Nf = 1 the component is treated as 1 × 1 (as the standard library does); Tq ≤ 3; component ids unique; X, Y ≥ 1; the segment length is 8 + 3·Nf. The reduced size is W8 = ⌈X/8⌉, H8 = ⌈Y/8⌉; `W8 × H8 > maxPixels` → `ErrTooLarge` (not wrapped), before any buffer is allocated | Bounds every count before use and accepts exactly the sampling the standard library accepts. 65,535² is the largest declarable frame, so W8 × H8 ≤ 67.1 M fits in `int64` | Accepting H or V = 3 — the standard library refuses it, so only crafted input would reach that code |
| D-8 | DHT: Tc ≤ 1, Th ≤ 3; the 16 code counts sum to at most 256 and the segment holds that many values; the canonical code is checked not to overflow at any length (Annex C); a DC table value over 11 is an error. A DHT replaces the table in its slot with a new, immutable table. Decoding uses a 9-bit lookahead table plus a bit-by-bit slow path, written for this package over a `[]byte` | Malformed tables are the classic decoder crash. The lookahead keeps a 39 MB baseline scan in the low seconds | Adapting the standard library's `huffman.go` — its reader is bound to the stdlib `decoder` struct, its `io.Reader` buffering and unread-byte rules, so the adaptation is a rewrite of the same size |
| D-9 | DQT: Pq 0 (8-bit) or 1 (16-bit) entries, Tq ≤ 3, segment length a whole number of tables; only entry 0 (the DC step) is kept. A component's step is latched at the first scan that includes it; a scan naming a missing quantisation table is an error | Only DC is used. Latching at the first scan is libjpeg's rule, so a later DQT cannot change already-decoded blocks | Reading the table at the end of the file |
| D-10 | SOS checks, in this order: segment length = 6 + 2·Ns; 1 ≤ Ns ≤ Nf; every Cs in the frame and not repeated; then by scan type. Sequential: Ss = 0, Se = 63, Ah = Al = 0, Td and Ta tables defined. Progressive DC first (Ss = 0, Se = 0, Ah = 0): Al ≤ 13, Td tables defined. Progressive DC refinement (Ss = 0, Se = 0, Ah > 0): Ah = Al + 1, and each component's last DC scan had Al = Ah; no table needed. Progressive AC (1 ≤ Ss ≤ Se ≤ 63): Ns = 1, Ah = 0 or Ah = Al + 1, Al ≤ 13; no table needed, the scan is skipped. Each scan takes its tables from the slots as they are at its SOS. Coverage: a component already covered by a sequential scan, or by a progressive DC first scan, in an earlier scan → error; DC refinement of a component with no DC first scan → error | T.81 B.2.3, G.1.1.1.1. Stating the AC-scan rules keeps a skipped scan from being an unchecked hole; the coverage rules keep a crafted file from decoding a block twice | — |
| D-11 | Block layout (T.81 A.2): Hmax, Vmax over the frame; MCUs across = ⌈X / (8·Hmax)⌉, down = ⌈Y / (8·Vmax)⌉. Each component stores one DC value per block in an `[]int16` of (MCUs across · Hc) × (MCUs down · Vc). An interleaved scan (Ns > 1) walks MCUs, each holding Vc rows × Hc columns of blocks per component in scan order. A single-component scan walks that component's ⌈⌈X·Hc/Hmax⌉/8⌉ × ⌈⌈Y·Vc/Vmax⌉/8⌉ blocks in raster order, one block per MCU. A one-component frame is always single-component, so a gray image tagged 2×2 has ⌈X/8⌉ × ⌈Y/8⌉ blocks | The shape every scan type and every fixture needs; the gray 2×2 and separate-DC fixtures hold it | Storing per scan — DC refinement updates blocks an earlier scan wrote |
| D-12 | At the first SOS, after every D-10 check passes and before the coefficient buffers are allocated, the decoder checks that (bytes after that SOS header) × 8 ≥ the number of blocks that scan covers; short → error | Each block of a first scan costs at least one bit, so a header declaring a 56 MP reduced image with no data is refused without a 113 MB allocation (AC-1.3). Running after D-10 means a header with valid tables reaches it (T-6). Later scans run on buffers already sized and fail on EOF | Allocating lazily per block row — complicates every scan walker for no extra safety |
| D-13 | Sequential scans decode every block: the DC difference (category ≤ 11) adds to the component's predictor; the AC symbols are decoded and discarded (ZRL = 16 zeros, EOB ends the block; position past 63 → error; AC category ≤ 10). Progressive DC first scan: DC = predictor sum `<< Al`. DC refinement: one bit per block, OR'd in at bit Al. AC scans: the decoder advances to the next marker that is neither a stuffed `0xFF00` nor RST0–7, without decoding | Design § What it reads. Skipping AC scans is what makes progressive cheap | Decoding progressive AC bands — no output uses them |
| D-14 | The predictor is an `int32`; the stored DC saturates to the `int16` range | A crafted run of differences cannot overflow; real 8-bit DC values stay within ±2,047 and are unaffected | Refusing out-of-range DC — rejects a file over a value no viewer would refuse |
| D-15 | Restart intervals: with DRI Ri > 0, after every Ri MCUs of a scan (except at its end) the next marker must be RST(n mod 8), counting from RST0 in each scan; the bit reader and predictors reset. A missing or out-of-order RST is an error. Skipped AC scans pass RST markers without checking them | T.81 E.1.4. The design says a malformed stream is an error, never a partial image | libjpeg-style resynchronisation — produces partial images |
| D-16 | The stream ends at EOI. EOF before EOI, a marker inside entropy data where no RST is due, or a component that no sequential or DC first scan covered, is an error. Bytes after EOI are ignored | Truncation never yields a partial picture (AC-2.3) | Accepting EOF as EOI — a half-written file would thumbnail as half grey |
| D-17 | Output: sample = clamp(round(DC × Q0 / 8) + 128, 0, 255) per block, round half away from zero. A pixel (x, y) of the W8 × H8 image takes, from each component, the block at (x·Hc/Hmax, y·Vc/Vmax) (integer division). Nf = 1 → `*image.Gray`. Nf = 3 → `*image.RGBA` (A = 255). The three samples are R, G, B exactly when the stdlib's `isRGB` holds: no JFIF APP0, and either an Adobe APP14 with transform 0, or component ids 'R', 'G', 'B' (whatever an Adobe transform says); otherwise `color.YCbCrToRGB` | The DC term over 8 is the block's mean before the level shift. Matching `isRGB` and `YCbCrToRGB` keeps colors as the full decode gives them | `*image.YCbCr` — needs one of the standard subsample ratios, which 2×2/1×2/1×2 does not fit |
| D-18 | Fixtures: copy into `core/raster/testdata/third_party/go/` *(new)* from `$(go env GOROOT)/src/image/testdata/`: `video-001.jpeg`, `video-001.progressive.jpeg`, `video-001.progressive.truncated.jpeg`, `video-001.q50.410.jpeg`, `video-001.q50.411.jpeg`, `video-001.q50.420.jpeg`, `video-001.q50.422.jpeg`, `video-001.q50.440.jpeg`, `video-001.q50.444.jpeg`, `video-001.q50.420.progressive.jpeg`, `video-001.q50.444.progressive.jpeg`, `video-001.221212.jpeg`, `video-001.restart2.jpeg`, `video-001.rgb.jpeg`, `video-001.cmyk.jpeg`, `video-001.separate.dc.progression.progressive.jpeg`, `video-005.gray.jpeg`, `video-005.gray.q50.2x2.jpeg`, `video-005.gray.q50.2x2.progressive.jpeg` (about 110 KB). Beside them: `LICENSE` (a copy of `$(go env GOROOT)/LICENSE`) and `NOTICE` (one paragraph: copied from the Go 1.27 distribution's `src/image/testdata/` for `core/raster` tests). The decodable set for T-1 is every file except `cmyk` and `truncated` | The standard library cannot encode progressive, RGB or odd-sampling JPEGs; these files cover each case the scope names (verified at planning: DC refinement, separate DC scans, DRI = 20, Adobe t=0 with ids R/G/B, gray tagged 2×2, Cb and Cr 1×2 under Y 2×2). `third_party/<upstream>/` with `LICENSE` and `NOTICE` follows `namealiases/third_party/nicknames` | Reading them from `GOROOT` at test time — CI images and future Go versions may move or drop them |
| D-19 | Fuzz target `FuzzDecodeJPEGDC` *(new)* in `core/raster/jpegdc_test.go` calls `decodeJPEGDC(data, 1<<20)`. Seeds: every D-18 file, plus `jpeg.Encode` output of a 37 × 21 RGBA gradient at quality 10 and 90 and a 16 × 16 `*image.Gray`. Invariants: no panic; on success the bounds are ⌈W/8⌉ × ⌈H/8⌉ for the SOF the input declares (the test reads W, H from the first SOF0–2 segment) and their product ≤ 1<<20 | `go test` runs the seeds in CI with no extra command; a 1 MP budget keeps each input fast | Fuzzing `EncodeJPEG` — only inputs declaring over 64 MP reach the reduced path |
| D-20 | `core/raster/rastertest` *(new package; imports only the standard library)* holds: `HeaderOnlyPNG(t, w, h uint32) []byte` (moved from `pngHeaderOnly`); `Gradient` *(new)*, an `image.Image` of a given size whose `At` computes a `color.YCbCr` (no pixel buffer); `OversizedJPEG(t) []byte`, the 8,008 × 8,000 `Gradient` encoded once per test binary (`sync.Once`) with `jpeg.Encode` at quality 50 (4:2:0, about 1 MB, about 2 s measured at planning); `JPEGSegment(marker byte, payload []byte) []byte` (`FF`, marker, big-endian length = len + 2, payload); `JPEGHeader(w, h uint16, comps int, scanData []byte) []byte`: SOI, DQT table 0 of 64 ones, DHT DC table 0 and AC table 0 each with one code of length 1 (value 0x00), SOF0 with `comps` components (ids 1, 2, 3 in order, H = V = 1, Tq = 0), SOS over all components (Td = Ta = 0, Ss 0, Se 63, Ah Al 0), then `scanData`, no EOI. `raster_test.go` and `thumbnail_test.go` use it; `pngHeaderOnly` and `bombPNG` are deleted | One home for "an image over budget", used by both packages and by S14-12's tests later; two existing copies become one. `JPEGHeader` has valid tables, so it reaches D-12 (T-6) | A `lazyImage` copied into each package — a third copy of a fixture beside the existing duplicate |

**Rules this plan bends.** None. STEP-8 adds a paragraph to `.cursor/rules/go-tests.mdc`; it contradicts nothing there.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/raster/jpegdc.go` `decodeJPEGDC` *(new)* | file, func | `image/jpeg` `Decode` (stdlib) | Its reader is unexported and always runs the full IDCT into a full-size buffer; `golang.org/x/image` has no reduced JPEG decoder |
| Huffman table, bit reader and `jpegErr(check string) error` in `jpegdc.go` *(new, unexported)* | type, func | `$(go env GOROOT)/src/image/jpeg/huffman.go` (stdlib, unexported) | Not importable; D-8 |
| `errJPEGDC` *(new, unexported)* | var | — | The one base error every `jpegErr` wraps, so a test can tell the reduced path from the stdlib path (T-4) |
| `raster.ErrUndecodable` *(new)* | var | `raster.ErrUnsupportedFormat`, `raster.ErrTooLarge` | Neither means "these bytes are a supported format but broken"; D-3 |
| `raster.MaxDecodeBytes`, `maxDecodeBytes` *(new)* | const, var | `derivatives.maxSourceBytes`, `ingest.imageMaxBytes` | Replaces `maxSourceBytes` (D-5); `StageImage`'s cap is stricter and user-facing, so it stays |
| `decode` / `prepare` parameter `reduceJPEG` *(new)* | param | `Options.Formats` | `Formats` limits what may decode; this chooses how an over-budget JPEG decodes, which only `raster` decides (D-2) |
| `core/raster/rastertest` *(new)* | package | `pngHeaderOnly` (`raster_test.go`), `bombPNG` (`thumbnail_test.go`), `artifactstest` (pattern) | Replaces the two copies (D-20) |
| `core/raster/testdata/third_party/go/` *(new)* | fixtures | `core/database/namealiases/third_party/nicknames/` (pattern), `core/derivatives/testdata/tiny.png` | The files are the stdlib's |
| `core/raster/jpegdc_test.go` *(new)*: `TestDecodeJPEGDCMatchesStdlib`, `TestDecodeJPEGDCRefuses`, `TestDecodeJPEGDCTruncated`, `TestJPEGDCHuffman`, `FuzzDecodeJPEGDC` | tests | `raster_test.go` tables | Same package, new file for the new codec |
| `fakeConfigCalls` *(new, in `raster_test.go`)* | test var | `init`'s `"fake"` format | Counts calls to the existing fake format's config func instead of registering a second format |

**Builds on.** `core/raster`'s single decode gate (`decode`), the `ErrTooLarge` and `ErrUnsupportedFormat` contracts, the caller-held `Slot`, the `ingest` test-lowered byte cap, the `<pkg>test` helper packages and the `third_party/<upstream>/` shape.
**Deletes.** `derivatives.maxSourceBytes` (D-5); `raster_test.go` `pngHeaderOnly` and `derivatives` `bombPNG` (D-20). No callee loses its last caller.
**Next adds.** S14-12's staged-object type calls `EncodeJPEG` under its own slot, gets the reduced path with no change, and reads `raster.MaxDecodeBytes` for its read cap (recorded in S14-12's scope entry, STEP-10). A larger cover derivative is a `derivatives.Spec` with a bigger `MaxEdge`; for an oversized JPEG it is capped at 1/8 (D-2, AC-1.7). A drawn-size decode is Swift's (`ThumbnailDecodePolicy.decode`'s `maxPixelSize`, S14-13's Next adds), not `raster`'s. A future Windows client calls the same Go. The next parser of untrusted bytes copies the fuzz pattern in `go-tests.mdc`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| An oversized JPEG now gets a `file_derivatives` row | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`, `existingThumbnailRelPath`; `core/database/sources/sources.go` `requireRasterCoverArtifact`; `derivatives.reprojectSourcesForNewThumbnail` | The same row shape as any JPEG thumbnail: 256 px JPEG, `TypeThumbnail`. Existing catalogs have no refusal stored, so the next Artifact response creates it; the Source cover and omnibar pick it up through the existing reprojection |
| A broken image now returns `ErrUnprocessable` instead of a raw error | `thumbnailRelPath` → empty path, `skipped`; `requireRasterCoverArtifact` → `ErrInvalid` | Both already handle `ErrUnprocessable`; the Artifact response now succeeds with no thumbnail, as for an oversized PNG |

### Scale

At the map's Scale (3,000 Sources) the cost is per oversized file, once: the row is stored and later reads are lookups. Per file: a 198 MP, 39 MB baseline scan has 3.1 M luma blocks; Huffman work is linear in the 39 MB of scan data. Memory peak at that size: `[]int16` DC buffers (3.1 M + 2 × 0.8 M blocks at 4:2:0 ≈ 9 MB) plus the 1,375 × 2,250 RGBA output (12 MB) plus the 39 MB input. Worst case the budget allows: a reduced image of 64 M pixels at 4:4:4 is 3 × 128 MB of DC buffers plus 256 MB RGBA ≈ 640 MB, below the standard library's own worst case at the same 64 MP budget for a progressive JPEG (64 `int32` coefficients per block). Concurrency stays bounded by the caller's `Slot` (at most 4). Time budget: in `TestEncodeJPEGOversized`, the `EncodeJPEG` call alone (not the fixture build) on the 64 MP fixture (1 M luma blocks) takes under 10 s on CI, timed in the test with `time.Since` and logged; Stop condition if not. A broken file that a broken-image refusal now covers is still re-read on every Artifact response, as an oversized PNG is today (Out).

### Edges

- **Assumptions.** `scaleToMaxEdge` uses `b.Min`, and the reduced image starts at (0, 0) (T-1 checks `Bounds().Min`). `decode` assumes `image.DecodeConfig` already accepted the SOF; the reduced decoder does not rely on it and re-checks every marker (D-6, D-7), because the fuzz target calls it directly (T-9). Callers assume `ErrTooLarge` and `ErrUnsupportedFormat` mean "refuse quietly"; D-3 adds `ErrUndecodable` with the same meaning, mapped in the one place that maps raster errors (`generate`), and `StageImage` maps every raster error already (T-12, T-13).
- **Neighbors.** 8,000 × 8,000 (exactly 64 MP) keeps the standard library (T-4's "at budget" row: the error is `ErrUndecodable` and not `errJPEGDC`). 8,008 × 8,000 takes the reduced path (T-4). PNG at 8,001 × 8,000 stays refused (existing). An oversized JPEG through `EncodePNG` stays refused (T-5). `MaxEdge` above ⌈W/8⌉ on the reduced path returns ⌈W/8⌉ × ⌈H/8⌉, not upscaled (AC-1.7). A small truncated JPEG, which failed the Artifact response, now refuses quietly (AC-3.4, AC-3.5).
- **Removed inputs.** No input is decided or deleted; a thumbnail row that a later delete removes is handled by the existing derivative release.
- **Real data.** No table or seed drives the rule. The planning marker dump over the 26 stdlib JPEG fixtures (2026-10-31) found every scan shape listed in D-18; none uses SOF1 or several sequential scans, so T-2 patches `video-001.jpeg`'s SOF0 to SOF1 (it must still decode).
- **Untrusted size.** Frame: W8 × H8 ≤ `maxPixels` before allocation (D-7). Coefficient buffers: at most 3 components, each ≤ (W8 + 4) × (H8 + 4) blocks because H, V ≤ 4. First scan data ≥ blocks / 8 bytes before allocation (D-12). Huffman counts ≤ 256 and segment-length checked (D-8). Scan count: not capped; every sequential or DC scan costs at least one bit per block, and AC scans are skipped linearly, so total work is linear in input bytes, which are ≤ `MaxDecodeBytes` (D-4). Marker segments: each length is checked against the remaining bytes before it is read (D-6).

### Schema

None — no catalog change. Existing catalogs need nothing: a refusal was never stored.

### Go core

| Package | Change |
| --- | --- |
| `core/raster` | `const MaxDecodeBytes = 128 << 20` *(new)*; `var maxDecodeBytes = MaxDecodeBytes` *(new)*; `ErrUndecodable` *(new)*; `ErrTooLarge` message and doc; package doc (D-4) |
| `core/raster` | `func decode(src []byte, formats []string, reduceJPEG bool) (image.Image, error)` — byte check, `DecodeConfig` with D-3's mapping, allowlist, budget, then the reduced or standard decode with D-3's mapping |
| `core/raster` | `func prepare(src []byte, opts Options, reduceJPEG bool)`; `EncodeJPEG` passes `true`, `EncodePNG` `false`. Doc comments say which path each takes and the 1/8 cap (D-2) |
| `core/raster/jpegdc.go` *(new)* | `decodeJPEGDC` and its unexported marker parser, Huffman table, bit reader, scan walkers (D-6 – D-17). Errors come from `jpegErr(check)`, which returns `fmt.Errorf("%w: %s", errJPEGDC, check)` |
| `core/raster/rastertest` *(new)* | D-20 |
| `core/derivatives` | `maxSourceBytes` → `raster.MaxDecodeBytes` (D-5); `generate` maps `raster.ErrUndecodable` to `ErrUnprocessable` (D-3) |

```go
// decodeJPEGDC decodes a JPEG at 1/8 scale from its DC coefficients only.
// It returns an image of ⌈W/8⌉ × ⌈H/8⌉ pixels with Min at (0, 0), or an
// error and no image. It allocates nothing sized by the frame until the
// frame is within maxPixels and the first scan has enough bytes (D-12).
func decodeJPEGDC(src []byte, maxPixels int64) (image.Image, error)
```

Errors: `raster.ErrTooLarge` (bytes or pixels over budget) and `raster.ErrUndecodable` (wrapping the decoder's own error) → `derivatives.ErrUnprocessable` → empty thumbnail path.

### FFI

None — no proto or handler change.

### macOS

None — no Swift change. The thumbnail cache decodes the 256 px derivative as for any JPEG.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — The byte limit.**
Files: `core/raster/raster.go`, `core/raster/raster_test.go`, `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: `MaxDecodeBytes`, `maxDecodeBytes`, the byte check at the top of `decode`, the new `ErrTooLarge` message and doc, the package doc (D-4). Replace `maxSourceBytes` with `raster.MaxDecodeBytes` (D-5). Add `fakeConfigCalls` to the `init` fake format. Add T-7 and make the T-10 edit.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster ./core/derivatives ./core/ingest` passes.

**STEP-2 — One refusal rule.**
Files: `core/raster/raster.go`, `core/raster/raster_test.go`, `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: `ErrUndecodable` and D-3's mapping in `decode` (both `DecodeConfig` and `image.Decode`); `generate` maps `ErrUndecodable`. Add T-12 and T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster ./core/derivatives ./core/ingest ./api/ffi/...` passes.

**STEP-3 — Shared fixtures.**
Files: `core/raster/rastertest/rastertest.go` *(new)*, `core/raster/testdata/third_party/go/` *(new: the D-18 files, `LICENSE`, `NOTICE`)*, `core/raster/raster_test.go`, `core/derivatives/thumbnail_test.go`.
Do: D-20's package; copy D-18's files with `cp "$(go env GOROOT)/src/image/testdata/<name>" core/raster/testdata/third_party/go/`. Move `pngHeaderOnly`'s callers to `rastertest.HeaderOnlyPNG` and `bombPNG`'s to the same, then delete both.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster/... ./core/derivatives` passes; `rg -n 'pngHeaderOnly|bombPNG' core` finds nothing.

**STEP-4 — Markers, frame, tables.**
Files: `core/raster/jpegdc.go` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: `jpegErr`, `errJPEGDC`; the marker loop, SOF, DHT, DQT, DRI, APP0, APP14, SOS header parsing and every check in D-6 – D-10; D-7's budget check; D-12's data check; buffer allocation per D-11. No entropy decoding yet: after allocation, return `jpegErr("scan decoding")`. Add T-2's header rows (listed in §8).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestDecodeJPEGDCRefuses/header' ./core/raster` passes.

**STEP-5 — Huffman and the bit reader.**
Files: `core/raster/jpegdc.go` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: canonical code build with the overflow check, 9-bit lookahead plus slow path, the byte-stuffing-aware bit reader that stops at a marker (D-8, D-16). Add T-11.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestJPEGDCHuffman' ./core/raster` passes.

**STEP-6 — Scans and output.**
Files: `core/raster/jpegdc.go` *(new)*, `core/raster/jpegdc_test.go` *(new)*.
Do: sequential scans and progressive DC first, refinement and AC skipping (D-13), restart handling (D-15), EOI and coverage checks (D-10, D-16), saturation (D-14), output per D-17. Add T-1, T-3 and T-2's scan rows.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestDecodeJPEGDC' ./core/raster` passes, every fixture within the AC-2.1 tolerance.

**STEP-7 — Wire into `EncodeJPEG`.**
Files: `core/raster/raster.go`, `core/raster/raster_test.go`.
Do: the `reduceJPEG` parameter on `decode` and `prepare` (D-2), the doc comments on `EncodeJPEG` and `EncodePNG`. Update `TestDecode` calls to pass `false` and add a `true` row. Add T-4, T-5, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 -v -run 'TestEncodeJPEGOversized' ./core/raster` passes and logs the `EncodeJPEG` time under 10 s; `CGO_ENABLED=1 go test -tags fts5 ./core/raster` passes.

**STEP-8 — Fuzzing.**
Files: `core/raster/jpegdc_test.go` *(new)*, `.cursor/rules/go-tests.mdc`.
Do: `FuzzDecodeJPEGDC` (D-19). Run the fuzzer for 5 minutes. When it fails, fix `jpegdc.go`, keep the crasher file it wrote under `core/raster/testdata/fuzz/FuzzDecodeJPEGDC/` *(new)* as a committed seed, and start a new 5-minute run; stop when one run ends with no failure. Add to `go-tests.mdc` a "Fuzz targets" paragraph: a package that parses bytes the app did not write gets a `Fuzz…` target over its parser; seeds come from `testdata/` and small encoder output, added with `f.Add`; every crasher is fixed and committed under `testdata/fuzz/<Target>/`, which `go test` then runs; the PR runs the fuzzer for 5 minutes locally (`go test -run '^$' -fuzz <Target> -fuzztime 5m`); `core/raster` `FuzzDecodeJPEGDC` is the example.
Check: the last `CGO_ENABLED=1 go test -tags fts5 -run '^$' -fuzz FuzzDecodeJPEGDC -fuzztime 5m ./core/raster` ends with no failure; `CGO_ENABLED=1 go test -tags fts5 ./core/raster` runs the seeds and passes; `python3 scripts/check-guidance-names.py` passes.

**STEP-9 — The derivative row.**
Files: `core/derivatives/thumbnail_test.go`.
Do: add T-8 with `rastertest.OversizedJPEG`. No production change in `derivatives` beyond STEP-1 and STEP-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/derivatives` passes.

**STEP-10 — Docs and version.**
Files: `docs/deployment-plan/spike-14/completed.md` (a new `## S14-23 — Oversized JPEG thumbnails` entry at the top, above S14-13, newest first); `docs/deployment-plan/spike-14/deployment-plan.md` (map row Status → `Landed #N`, as the other landed rows read); this plan (Status → `Landed in [#N](…) (date)` once the PR number exists, and a Landed note); `docs/deployment-plan/spike-14/oversized-jpeg-thumbnails.md` (§ What it reads: APP0 and APP14 are read for the color rule, D-6; § Bounds: decode failures are `ErrUndecodable`, D-3; a first line "Landed in #N"); `docs/source-layer-data-model.md` (line 313: "within the pixel budget" → "within the byte limit and pixel budget"; § 8 `file_derivatives`, one paragraph after the sentence beginning "Derivative Files use the same content-addressed": a thumbnail is decoded by `core/raster` within 128 MiB and 64 MP; a JPEG over 64 MP is decoded at 1/8 scale from its DC coefficients; other formats over the budget, and bytes that cannot be decoded, get no thumbnail); `.agents/skills/review-pr/lenses/security.md` (line 24, the `core/raster` row: "the one bounded codec (bytes, pixels, format allowlist), also used by staging, with a hand-written reduced JPEG decoder (fuzzed)"; line 134: "Go's `raster` has both" → "Go's `raster` has a byte, pixel and type bound"); `VERSION`, `core/version.go`, `MARKETING_VERSION` in `macos/Provenencia.xcodeproj/project.pbxproj` → 0.2.3.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-23-oversized-jpeg-thumbnails.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS` (its `xcode-test` needs macOS; on another OS report that check as not run).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show a thumbnail for a JPEG scan over 64 megapixels."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/raster/jpegdc_test.go` `TestDecodeJPEGDCMatchesStdlib` *(new)* | Each decodable D-18 fixture → bounds (0,0)–(19,13); over the 18 × 12 full cells, mean abs diff per channel ≤ 3.0 against `jpeg.Decode` averaged over each 8 × 8 cell; gray fixtures return `*image.Gray`, colour fixtures `*image.RGBA` | AC-2.1, D-11, D-17 |
| T-2 | Go | `core/raster/jpegdc_test.go` `TestDecodeJPEGDCRefuses` *(new)*, rows built with `rastertest.JPEGHeader` / `JPEGSegment` or by patching a fixture byte located by scanning for its marker | **header rows (STEP-4)**, each → error matching `errJPEGDC`, no image: `video-001.cmyk.jpeg`; `video-001.jpeg` with its SOF0 marker byte set to 0xC9; SOF P = 12; `JPEGHeader` with 2 components; component 0 sampling byte 0x51 (H = 5); 0x31 (H = 3); DHT whose 16 counts sum to 257; DHT DC table with value 12; DQT with Tq = 4; SOS naming component id 9; progressive AC SOS (SOF2, Ss 1, Se 63) with Ns = 2; SOS length one byte short. `JPEGHeader(65535, 65535, 1, nil)` → `ErrTooLarge`, not `errJPEGDC`. **scan rows (STEP-6)**: `video-001.restart2.jpeg` with the first `FF D0` after SOS changed to `FF D1` → error; `video-001.jpeg` SOF0 → SOF1 (0xC1) → decodes, same pixels as unpatched; a second sequential SOS over component 1 appended before EOI → error | AC-2.2, D-6 – D-10, D-15 |
| T-3 | Go | `core/raster/jpegdc_test.go` `TestDecodeJPEGDCTruncated` *(new)* | `video-001.progressive.truncated.jpeg`; `video-001.jpeg` cut at half; `video-001.jpeg` without its last two bytes (EOI) → error, no image | AC-2.3, D-16 |
| T-4 | Go | `core/raster/raster_test.go` `TestEncodeJPEGOversized` *(new)*; `TestEncodeJPEGAtBudgetBoundaries` *(extend)* | `rastertest.OversizedJPEG`, `MaxEdge: 256` → 256 × 255; `MaxEdge: 2048` → 1,001 × 1,000; the `EncodeJPEG` call's time logged. At-budget row: `JPEGHeader(8000, 8000, 1, nil)` through `decode(…, true)` → `ErrUndecodable` and not `errJPEGDC` (standard-library path) | AC-1.1, AC-1.7, D-2 |
| T-5 | Go | `core/raster/raster_test.go` `TestEncodePNG` *(new row "oversized jpeg never reduces")* | `rastertest.OversizedJPEG` through `EncodePNG(…, Options{MaxEdge: 2048})` → `ErrTooLarge` | AC-1.5, D-2 |
| T-6 | Go | `core/raster/raster_test.go` `TestEncodeJPEGRejects` *(new rows)* | `JPEGHeader(65535, 65535, 1, nil)` → `ErrTooLarge`; `JPEGHeader(60000, 60000, 1, 16 zero bytes)` → `ErrUndecodable` wrapping `errJPEGDC`, and `runtime.MemStats.TotalAlloc` grows < 16 MiB across the call (read with `runtime.ReadMemStats` before and after); `OversizedJPEG` cut at half → `ErrUndecodable`. Existing GIF and PNG bombs unchanged | AC-1.3, AC-1.4, AC-1.6, D-12 |
| T-7 | Go | `core/raster/raster_test.go` `TestDecodeByteLimit` *(new)* | `maxDecodeBytes = 64`, restored by `t.Cleanup`; `fakeConfigCalls` reset to 0 per row. 65 bytes (`FAKEIMG!` + 57 bytes) → `EncodeJPEG` and `EncodePNG` both `ErrTooLarge`, calls 0. 64 bytes → `ErrUnsupportedFormat`, calls 1 per call | AC-3.1, AC-3.2, D-4 |
| T-8 | Go | `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailOversizedJPEG` *(new)* | `installSourceFile(t, c, rastertest.OversizedJPEG(t), "image/jpeg", "big.jpg")` → `EnsureThumbnail` returns a link; derived File `image/jpeg`; its object decodes to 256 × 255; second call → same `Link.ID` | AC-1.2 |
| T-9 | Go | `core/raster/jpegdc_test.go` `FuzzDecodeJPEGDC` *(new)* | Seeds per D-19 → no panic; success implies bounds = ⌈W/8⌉ × ⌈H/8⌉ and ≤ 1<<20 pixels | AC-2.4, D-19 |
| T-10 | Go | `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources` *(edit)* | "declared ByteSize over source cap" uses `raster.MaxDecodeBytes + 1` → `ErrUnprocessable` | AC-3.3, D-5 |
| T-11 | Go | `core/raster/jpegdc_test.go` `TestJPEGDCHuffman` *(new)* | A DHT payload with counts {1, 2, 0…} (one 1-bit code, two 2-bit codes) and values {5, 6, 7} → codes `0` → 5, `10` → 6, `11` → 7 for the bit string `0 10 11`; counts {3, 0…} (three 1-bit codes) → overflow error; `video-001.jpeg`'s first DHT builds without error | D-8 |
| T-12 | Go | `core/raster/raster_test.go` `TestEncodeJPEGRejects` *(new rows)* | `video-001.jpeg` cut at half → `ErrUndecodable`; the existing "garbage bytes" row now asserts `ErrUnsupportedFormat` | AC-3.4, AC-3.6, D-3 |
| T-13 | Go | `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources` *(new row)* | `installSourceFile` with `video-001.jpeg` cut at half (read from `core/raster/testdata/third_party/go/` *(new)* through a relative path), `image/jpeg` → `ErrUnprocessable` | AC-3.5, D-3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/raster` `TestDecode` (signature change), `core/derivatives` `TestEnsureThumbnailRejectsUnsafeSources` (constant rename, helper move), `core/ingest` `TestStageImage…` (must be unaffected; it uses `EncodePNG`), `api/ffi/handlers` thumbnail tests (a decode error that used to surface is now skipped; if one asserts the old raw error, that is a Stop condition).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] The last 5-minute fuzz run in STEP-8 ended with no failure; every crasher found is a committed seed
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: `maxSourceBytes`, `pngHeaderOnly`, `bombPNG`, "image dimensions exceed decode budget" and "within a pixel budget" / "within the pixel budget" `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/` and `L10n.swift`, each hit either gone or naming the byte limit too; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump 0.2.2 → 0.2.3 (VERSION, core/version.go, MARKETING_VERSION together)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc and guidance updated: docs/source-layer-data-model.md (line 313, § 8); oversized-jpeg-thumbnails.md; security lens lines 24 and 134; go-tests.mdc "Fuzz targets"
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Build the app from the branch. In a test project, attach a JPEG over 64 MP to a Source (to make one on macOS: `sips -z 18000 11000 any.jpg --out big.jpg`, then `sips -g pixelWidth -g pixelHeight big.jpg` shows 11000 × 18000). Open the Source. The Artifact tile shows the scan's thumbnail, not the missing-image glyph.
2. Choose "Use as thumbnail" on that Artifact. The Source header and the Sources list row show it as the cover.
3. Type the Source's title in the omnibar. The hit shows the thumbnail.
4. Attach a JPEG under 64 MP and a PNG. Both thumbnail as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A hand-written decoder reads untrusted bytes | Every count checked before use (Edges), fuzzing (T-9), errors never yield partial images (D-16), the caller's slot bounds concurrency |
| Colour or rounding differs from the full decode | T-1 tolerance against the stdlib on every sampling and colour rule, over full cells |
| A real camera or scanner file uses a shape the fixtures miss (several sequential scans, SOF1) | D-10 and D-13 treat scans uniformly; T-2's SOF1 and second-scan rows; a real file that fails is refused quietly (D-3), never worse than today |
| The oversized tests slow CI | One 1 MB fixture per test binary (`sync.Once`); Stop condition at 10 s |

**Stop and report if:**

- `core/raster/raster.go` `decode` no longer has the signature `decode(src []byte, formats []string)` or no longer runs `image.DecodeConfig` before the budget check.
- A fixture in D-18 cannot reach a mean abs diff ≤ 3.0 per channel over full cells (T-1) with D-17's formula: report the fixture and the measured value; do not loosen the tolerance.
- The `EncodeJPEG` call in `TestEncodeJPEGOversized` takes over 10 s, or T-6's allocation check needs more than 16 MiB: report the measurement.
- A fuzz crasher's fix needs a rule not in D-6 – D-17.
- Any `ingest`, `derivatives` or `api/ffi` test needs a change beyond T-7, T-8, T-10, T-12, T-13 and the helper moves in STEP-3.
- A fixture file in D-18 is missing from the Go 1.27 distribution.

## 12. Open questions

None.

## 13. PR

**Title:** S14-23 — Thumbnails for oversized JPEGs
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (test names, the fuzz runs' count and the last one's result, `TestEncodeJPEGOversized` timing); deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-23-oversized-jpeg-thumbnails.md`.
