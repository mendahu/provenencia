# Ship log — S14-23 Thumbnails for oversized JPEGs

**Plan:** [`S14-23-oversized-jpeg-thumbnails.md`](S14-23-oversized-jpeg-thumbnails.md) · **Branch:** `s14-23-oversized-jpeg-thumbnails` · **PR:** #654
**State:** ready

## Round 0 — implement
Base `85a5ce2e`.
Head `0d5781d1` (local, not pushed). Gate not run.
Commits: STEP-1 `8e5a98d6`, STEP-2 `76c3a9b6`, STEP-3 `0d5781d1`. Uncommitted: `core/raster/jpegdc.go`, `core/raster/jpegdc_test.go` (STEP-4 – STEP-6).

**Stopped on:** T-1 / AC-2.1 tolerance (Stop condition §11 line 2). Every chroma-subsampled fixture misses mean abs diff ≤ 3.0 per channel against `jpeg.Decode` averaged over each 8 × 8 cell:

| Fixture | R / G / B |
| --- | --- |
| `q50.410` | 11.34 / 3.63 / 11.55 |
| `q50.411` | 8.76 / <3 / 9.29 |
| `q50.420` | 7.19 / <3 / 7.32 |
| `q50.420.progressive` | 7.19 / <3 / 7.32 |
| `q50.422` | 4.00 / <3 / 4.36 |
| `q50.440` | 6.09 / <3 / 6.28 |
| `221212` | 4.08 / <3 / 4.44 |
| `restart2` | 7.21 / <3 / 7.50 |
| `separate.dc.progression.progressive` | 7.20 / <3 / 7.48 |
| `rgb` (R 2×2, G B 1×1) | <3 / 21.15 / 20.96 |

Passing: `video-001.jpeg`, both 4:4:4, `progressive`, `progressive.truncated`, all three gray.

Implementer's diagnosis: the formula, not the decoder. D-17 gives every 8 × 8 luma cell under one chroma block that block's DC, while the full decode's chroma varies inside the 16 × 16 (4:2:0) or 32 × 16 (4:1:0) area. Against a reference built from the stdlib's own `image.YCbCr` planes (Y averaged per cell, Cb/Cr averaged over the covering chroma block, then `YCbCrToRGB`) the decoder measured 0.15 – 0.21 on 4:2:0, 4:1:0, 4:2:2, 4:4:0. Throwaway test, deleted.

Taken to the user: AC-2.1 / T-1 needs a plan change through `plan-pr`.

**User (2026-10-31):** "Block-level reference". Plan amended in `bd9a9a4e`: AC-2.1, T-1 and the Stop condition cite new D-22 (mean of the stdlib's own component over each block's covering rectangle, clipped; 3.0 unchanged). One fresh reviewer measured D-22 with the round-0 decoder on every D-18 fixture: 0.00 – 0.84 worst channel; three wording minors fixed, one declined (review log, "Amendment 2026-10-31"). Round 0 resumes at STEP-6 with the same implementer.

Head `1c96ecb6`. Gate PASS (orchestrator run). Report deviations: version already 0.2.3 from S14-03 (no bump; plan line stale); T-6 allocation row is its own test `TestEncodeJPEGOversizedHeaderAllocation`; two extra assertions (`TestDecode` "oversized jpeg reduces only when asked", 8000 × 8000 header stays on stdlib path in `TestEncodeJPEGAtBudgetBoundaries`). Fuzz: one 5-minute run, 12.8M execs, no failure. `TestEncodeJPEGOversized` ~35–40 ms per call. Worst T-1 channel 0.84.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF (`core/raster`, `core/derivatives`, hand-written decoder over untrusted bytes). Skipped: UX (no `macos/App/**`, strings, or new `apperr` code; `ErrUndecodable` maps to the existing `ErrUnprocessable`).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Three readers of the vendored Go fixture folder, one a cross-package `../raster/testdata` path | Fix now → F-1 | verified: `rg -n third_party core --type go` → `raster_test.go` `goFixture`, inline `os.ReadFile` in `FuzzDecodeJPEGDC`, `thumbnail_test.go:364`; all added by this PR |
| CON-2 | CON (from MNT outside-lens) | minor | "oversized jpeg reduces only when asked" never shows the flag switches paths | Fix now → F-2 | verified: `raster_test.go:317-325` feeds a 65,535² header that both paths refuse with `ErrTooLarge` |
| MNT-2 | MNT | minor | DHT count/length rule in both `parseDHT` and `newHuffTable` | Fix now → F-3 | verified: `jpegdc.go:59-68`, `:485-494`; code this PR adds |
| MNT-3 | MNT | minor | Unreachable `dc category over 11` in `block` | Fix now → F-4 | verified: `jpegdc.go:748`; `sc.dcTab` only from `d.huff[0]`, built with `dc=true`, which refuses values over 11 |
| MNT-4 | MNT | minor | Test header builder always writes SOF0; one row byte-searches for `FF C0` | Fix now → F-5 | verified: `jpegdc_test.go:286`; D-21 rules out byte search |
| CON-1, MNT-9 | CON, MNT | minor | Stale "pixel budget" docs: `StageImage` (`stage.go:46`), `EncodePNG` (`raster.go:146`) omit the byte limit | Fix now → F-6 | DoD sweep line; `rg 'within the pixel budget' core` |
| MNT-7, MNT-8, MNT-10, MNT-11 | MNT | nit | Wrong or misleading comments/docs in files this PR adds or changes | Fix now → F-6 | stale comment in a file this PR adds or changes |
| PRF-1 | PRF | minor | A broken or crafted oversized JPEG is fully re-decoded on every Artifact response, inside the session | Later PR (S14-12) | measured by reviewer: 565 ms realistic truncated scan, 3.04 s / 592 MB crafted 63,000² 3-comp, linear, within the plan's ~640 MB worst case; plan Out lines put persisting a refusal at "not planned" and decode-outside-session at S14-12. Written into S14-12's map scope entry (**Ships**, after the `MaxDecodeBytes` sentence) |

Nits (logged): `MNT-5 · core/raster/jpegdc.go:275,647,828 · dcDecoder.scans redundant with covered check` · `MNT-6 · core/raster/jpegdc_test.go:183 · sofPayload duplicates find`

SEC: no findings. Reviewer's two 120 s fuzz runs (1.46 M, 2.66 M execs) passed.

### Fix plan 1

**F-1 — One reader for the vendored Go fixtures** (from MNT-1)
Files: `core/raster/rastertest/rastertest.go`, `core/raster/raster_test.go`, `core/raster/jpegdc_test.go`, `core/derivatives/thumbnail_test.go`.
Failing test first: none (refactor; the existing tests that read fixtures are the check).
Do: add `rastertest.GoFixture(tb testing.TB, name string) []byte` *(new)*, which finds `core/raster/testdata/third_party/go/` from its own source file through `runtime.Caller(0)` (the files stay where D-18 put them) and fails `tb` on a read error. Delete `goFixture` from `raster_test.go` and move its callers to `rastertest.GoFixture`; replace the inline `os.ReadFile` in `FuzzDecodeJPEGDC` and the `../raster/testdata/...` read in `thumbnail_test.go` with it.
Check: `rg -n 'third_party' core --type go` prints only the line in `rastertest.go`; `go test -tags fts5 ./core/raster/... ./core/derivatives ./core/ingest` passes; scripts/check-pr.sh PASS.

**F-2 — Prove the reduce flag switches paths** (from CON-2)
Files: `core/raster/raster_test.go`.
Failing test first: the subtest rewritten as below must fail if `decode` ignores the flag (check by temporarily passing `false` through; do not commit that).
Do: in "oversized jpeg reduces only when asked", use `rastertest.OversizedJPEG(t)`: `decode(src, both, false)` → `ErrTooLarge`; `decode(src, both, true)` → no error and bounds 1,001 × 1,000. Keep the 65,535² header as a third case asserting `ErrTooLarge` with `true`.
Check: the test passes; scripts/check-pr.sh PASS.

**F-3 — One home for the DHT count rule** (from MNT-2)
Files: `core/raster/jpegdc.go`, `core/raster/jpegdc_test.go`.
Failing test first: none (refactor; `TestJPEGDCHuffman` and `TestDecodeJPEGDCRefuses` `header/` DHT rows are the check).
Do: `newHuffTable` takes the raw table payload after Tc/Th (16 counts then values) and returns the table and the bytes it consumed; it alone sums the counts, refuses a total over 256 and a payload shorter than the total. `parseDHT` only reads Tc/Th, calls it, and advances. `TestJPEGDCHuffman` calls the same entry point and drops its own count sum.
Check: `rg -n 'more than 256 codes' core/raster/jpegdc.go` prints one line; tests pass; scripts/check-pr.sh PASS.

**F-4 — One DC category guard** (from MNT-3)
Files: `core/raster/jpegdc.go`.
Do: delete the `s > 11` check in `block`; on the table-time check in `newHuffTable` add one comment that it is the guard `block` relies on.
Check: tests pass (including `FuzzDecodeJPEGDC` seeds); scripts/check-pr.sh PASS.

**F-5 — Header builder names its SOF marker** (from MNT-4)
Files: `core/raster/jpegdc_test.go`.
Do: add `sofMarker byte` to `headerParts` (default `0xC0` in `defaultParts`); `build` writes it; the progressive-AC row sets `0xC2` and drops the `bytes.Index` patch.
Check: `rg -n 'bytes.Index' core/raster/jpegdc_test.go` prints nothing; tests pass.

**F-6 — Docs and comments** (from CON-1, MNT-9, MNT-7, MNT-8, MNT-10, MNT-11)
Files: `core/ingest/stage.go:46`, `core/raster/raster.go:146`, `core/raster/jpegdc_test.go:198` and `:266`, `.agents/skills/review-pr/lenses/security.md:24`, `docs/ideas/depictions-and-likenesses.md:15`.
Do: `StageImage` doc: "within the byte limit and pixel budget". `EncodePNG` doc: "the same byte limit, format allowlist and pixel budget as EncodeJPEG". `reducedPix` comment: "reducedPix returns the pixels of src decoded by the reduced decoder". Line 266 comment: "Cb 4×4 over Y 2×2: Hmax % Hc ≠ 0". Security lens line 24: "the one bounded codec: bytes, pixels, format allowlist; also used by staging; holds a hand-written, fuzzed reduced JPEG decoder". Ideas line 15: move the S14-23 byte-limit item out of the "Spike 12 takes" list into a sentence of its own saying S14-23 landed it.
Check: `rg -n 'within the pixel budget' core` prints nothing; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `0ebd6f3f`. Gate PASS (orchestrator run). F-1 – F-6 done (`469c8ce9`, `164f267d`, `b45e688c`, `0ebd6f3f`). Deviations: F-2's temporary flag-ignore check not run (the `true` call on `OversizedJPEG` cannot pass if the flag is ignored); `164f267d` alone does not compile its tests, head passes.
CON-2: **Resolved** — read `raster_test.go:304-320`: `OversizedJPEG` → `ErrTooLarge` with `false`, 1,001 × 1,000 with `true`; 65,535² still `ErrTooLarge`. CON-1 / F-6 doc lines: **Resolved** (`rg 'within the pixel budget' core` empty).

## Round 2 — review
Reviewers: MNT (had Fix-now major MNT-1). Scope `1c96ecb6...0ebd6f3f` (`dd701c5d` is the orchestrator's S14-12 scope note). CON, SEC+PRF not re-sent: no Fix-now blocker or major.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) Three fixture readers | Resolved | `rg -n third_party core --type go` → only `rastertest.go:136,144` in the raster tree |

No new findings. Nits (logged): `MNT · core/raster/jpegdc_test.go:293 · "scan/wrong restart marker" still byte-searches for RST0 after SOS` — accepted: an RST marker lives in entropy-coded data `jpegSegments` does not walk, and T-2 specifies "the first FF D0 after SOS". (The round-1 fix report's "`bytes.Index` prints nothing" was wrong for this line.)

## Finish
Merged `origin/main` (`14c226a3`, S14-11 plan docs only). Gate PASS after the merge.

## Result
Gate PASS · CI pending · open: none · follow-ups: PRF-1 → S14-12 scope entry · to user: — (AC-2.1 amendment answered: "Block-level reference") · CON miss: none
