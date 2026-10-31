# Review — S14-23 Thumbnails for oversized JPEGs

**Verdict:** round 1: Fits with changes — Not ready. Round 2: pending.
**Rounds:** 1 (round 2 running). **Reviewers:** maintainability (Opus), executability (Sonnet)

## To user
- None.

## Writer's own calls
- **Broken images refuse quietly (D-3, M-1).** Every decode failure in `raster`, on the standard-library path too, becomes `ErrUndecodable` → `ErrUnprocessable` → no thumbnail. Today a truncated or arithmetic-coded small JPEG fails the whole Artifact response. This widens the PR by about six lines; it was an Out line in the first draft.
- **Byte limit 128 MiB (D-4).** It is `derivatives`' cap today; `StageImage` keeps 32 MiB.
- **Hand-written Huffman decoder (D-8).** The standard library's is not reusable as is (bound to its decoder struct and reader); adapting it would be a rewrite of the same size.
- **Vendored Go test fixtures (D-18)** under `core/raster/testdata/third_party/go/` with Go's BSD `LICENSE` and a `NOTICE`.
- **New guidance:** a "Fuzz targets" paragraph in `.cursor/rules/go-tests.mdc` (STEP-8).
- **Order before S14-12.** The map prefers S14-12 first; S14-12 is unplanned, and this PR leaves it one byte limit to read.

## Round 1

### M-1 · blocker · Two ways to handle "these bytes cannot be decoded" in one `decode`
D-3 wrapped reduced-path failures in `ErrTooLarge` while the stdlib path returned raw errors that fail the Artifact response; the rejection reason ("three new cases") was wrong — only `derivatives.generate` maps raster errors.
**Answer:** Fixed — D-3 is one rule: `image.ErrFormat` → `ErrUnsupportedFormat`, every other decode failure → `ErrUndecodable` *(new)*, mapped in `generate`. STEP-2, AC-3.4 – AC-3.6, T-12, T-13; the Out row is gone; T-4 tells the paths apart with `errJPEGDC`.

### M-2 · major · `lazyImage` copied into two packages beside an existing duplicate
**Answer:** Fixed — D-20 adds `core/raster/rastertest` (`HeaderOnlyPNG`, `Gradient`, `OversizedJPEG`, `JPEGSegment`, `JPEGHeader`); `pngHeaderOnly` and `bombPNG` are deleted (STEP-3). `encodeImage` stays (Out): it builds small full images for many existing rows.

### M-3 · minor · Docs the byte limit makes incomplete
**Answer:** Fixed — STEP-10 updates the package doc, security lens line 134 and `source-layer-data-model.md` line 313; the DoD sweep adds "pixel budget".

### M-4 · minor · Larger derivative silently capped at 1/8; wrong Next adds and Out pointer
**Answer:** Fixed — D-2 documents the cap in `EncodeJPEG`'s doc, AC-1.7 pins it, Next adds corrected (drawn-size decode is Swift's), the Out row points at the design note.

### M-5 · minor · Vendored licence does not follow `third_party/<upstream>/`
**Answer:** Fixed — D-18 uses `core/raster/testdata/third_party/go/` with `LICENSE` and `NOTICE`.

### M-6 · minor · D-8's licence reason does not hold
**Answer:** Fixed — the licence reason is dropped; D-8 rests on the stdlib reader being bound to its decoder struct and buffering. Recorded under Writer's own calls for veto.

### M-7 · minor · Test shape against `go-tests.mdc`; first fuzz target records no pattern
**Answer:** Fixed — T-5 is a row in `TestEncodePNG`, the at-budget case extends `TestEncodeJPEGAtBudgetBoundaries`, T-7 counts calls on the existing `"fake"` format, STEP-8 adds the "Fuzz targets" paragraph.

### M-8 · minor · S14-12 ordering and its byte cap unrecorded
**Answer:** Fixed — header **Depends on** explains the order; Next adds and STEP-10 add a line to S14-12's scope entry (done in the map with this plan).

### M-9 · nit · Per-check error vars nobody matches
**Answer:** Fixed — one `jpegErr(check)` wrapping `errJPEGDC`.

### X-1 · blocker · STEP-2 and STEP-3 use fixtures STEP-6 copies
**Answer:** Fixed — fixtures, `LICENSE`, `NOTICE` and `rastertest` are STEP-3, before any decoder step.

### X-2 · blocker · Header-only JPEG helper unspecified, so D-12 can go untested
**Answer:** Fixed — D-20 specifies `JPEGHeader` byte by byte with a DQT and both DHTs; D-12 runs after every D-10 check, so T-6's 60,000² row reaches it, and asserts `errJPEGDC`.

### X-3 · major · "grayscale" false for a lazy image
**Answer:** Fixed — AC-1.2 / T-8 use `OversizedJPEG` (YCbCr) and drop "grayscale"; one-component decoding is covered by T-1's gray fixtures.

### X-4 · major · T-2 split across steps with undefined crafting
**Answer:** Fixed — T-2 lists header rows (STEP-4) and scan rows (STEP-6), each with its construction; the RST row's patch is named.

### X-5 · major · Edges the walkers must guess
**Answer:** Fixed — D-7 matches the stdlib's sampling rule (H, V ∈ {1, 2, 4}, divisibility for 3 components); D-10 states order, length, coverage, refinement (Ah = Al + 1), AC-scan validation without tables, and tables taken at SOS; D-6 states the APP0/APP14 length and id checks; D-17 restates `isRGB` unambiguously.

### X-6 · major · Oversized fixture built three times
**Answer:** Fixed — `rastertest.OversizedJPEG` is built once per test binary; the 10 s budget times the `EncodeJPEG` call only.

### X-7 · minor · Test format underspecified
**Answer:** Fixed — T-7 uses the existing `FAKEIMG!` format; the 64-byte row expects `ErrUnsupportedFormat`.

### X-8 · minor · Checks not runnable here; fuzz reruns; edge-cell tolerance
**Answer:** Fixed — STEP-8 reruns until one full 5-minute run passes; T-1 compares full cells only; STEP-10 says how to report `xcode-test` off macOS.

### X-9 · minor · STEP-1 carries T-10 unnamed
**Answer:** Fixed.

### X-10 · minor · STEP-10 choices
**Answer:** Fixed — `completed.md` entry at the top; map status `Landed #N`; plan status filled once the PR number exists.

### X-11 · nit · Map does not link the plan
**Answer:** Fixed — map row links the plan.
