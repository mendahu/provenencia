# S10-53 — Stage a client image

**Status:** Draft
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-53--stage-a-client-image)
**Brief:** — (no new chrome: an engine API with no caller until S10-32)
**Board:** —
**Depends on:** —
**Blocks:** S10-32
**Branch:** `s10-53-stage-image`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) "The image", [`docs/source-layer-data-model.md`](../../../source-layer-data-model.md) §6, working rule "The client renders the image; Go validates it and stores its own copy", map review [`../deployment-plan.review.md`](../deployment-plan.review.md) K-4, K-5
**Planned:** 2026-10-16 against `main` @ `98819f86`
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

Go can accept an image the Mac app rendered and store it safely as a File. The app writes a crop to a temporary file and hands Go its path, the way Artifact import already does. Go applies the checks every import gets, decodes the image inside the decoder's format allowlist and pixel budget, re-encodes its own PNG at most 2048 pixels on the long edge, and stores that PNG, never the app's bytes. The work runs in two phases, so no decode happens while the catalog write is open, and a failed write leaves nothing on disk.

**Why now.** S10-32 stores each likeness image this way. The user asked that a client-sent file get every protection a file import gets (2026-10-16). Today the import checks live inline in `ingest.File`, which stores the client's bytes and never decodes, and the bounded decode lives inline in `raster.EncodeJPEG`. S10-32 would otherwise copy both.

**What stays.** `ingest.File` accepts and refuses exactly what it does today, with the same errors and audit. Thumbnails are generated exactly as today. No schema, FFI, or Swift change.

## 2. User stories

**US-1.** As a researcher, a picture the app sends to be stored is checked the way an imported file is, so a bad file cannot get in by that path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `StageImage` on a relative path, a missing path, a symlink, a directory, an empty file, and a file one byte over `MaxImageBytes` | Refused with `ingest.ErrInvalid`, `ErrMissing`, `ErrSymlink`, `ErrNotAFile`, `ErrEmpty`, and `ErrTooLarge`, the same errors `ingest.File` returns. No object is written. |
| AC-1.2 | `StageImage` on a TIFF, a GIF, a WebP, a BMP, an HTML file renamed `.png`, and bytes that are not an image | Refused with `ingest.ErrUnsupportedImage` before a full decode. No object is written. |
| AC-1.3 | A PNG whose header declares 100000 × 100000 pixels | Refused with `ingest.ErrUnsupportedImage` before a full decode. |
| AC-1.4 | A valid 300 × 200 PNG followed by 1 KiB of appended ZIP bytes | Staged. The stored object is a PNG that decodes to 300 × 200 and is not byte-equal to the input, and has no trailing bytes. |
| AC-1.5 | A valid 4000 × 1000 JPEG | Staged as a PNG of 2048 × 512. |

**US-2.** As a contributor, the staging API lets a write store the image with its row, and a failed write leaves nothing behind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Stage a valid PNG; inside `writes.Run`, `InsertTx`; the write commits | One `files` row with media type `image/png`, no original filename, and its object on disk. The change list holds one `file` create change. |
| AC-2.2 | Stage a valid PNG; the write function returns an error; then `Discard` | No `files` row and no object. |
| AC-2.3 | Stage the same PNG twice; insert both in one write | One `files` row; the second `InsertTx` returns the same File and no change. |
| AC-2.4 | Stage an image whose re-encode equals an object a `files` row already holds; `Discard` without inserting | The existing object stays. |

**US-3.** As a contributor, imports and thumbnails behave as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Every existing `ingest` test and `raster` test | Passes unedited. |
| AC-3.2 | Every existing thumbnail test in `core/derivatives` | Passes unedited. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ingest.openChecked` *(new)*: the open-and-check step, shared by `ingest.File` and `StageImage` | STEP-1, AC-1.1, AC-3.1 |
| `raster.Decode` *(new)* bounded decode with a format list; `EncodeJPEG` uses it; `raster.EncodePNG` *(new)* | STEP-2, AC-1.2, AC-1.3, AC-3.1 |
| `raster.Slot` *(new)*: the decode semaphore moves from `derivatives` into `raster` | STEP-2, AC-3.2 |
| `ingest.StageImage`, `StagedImage.InsertTx`, `StagedImage.Discard`, `ingest.ErrUnsupportedImage`, `ingest.MaxImageBytes` *(new)* | STEP-3, STEP-4, AC-1.1 to AC-2.4 |
| Docs | STEP-5 |

| Out | Home |
| --- | --- |
| Calling `StageImage` from an Observation write | S10-32 |
| The FFI field that carries the crop path, and the Swift error copy for `ingest.unsupported_image` | S10-54 |
| Rendering the crop in the app | S10-34 |
| A Go PDF renderer | Not planned ([depictions and likenesses](../depictions-and-likenesses.md)) |
| A better downscale filter than nearest-neighbour | Not planned; the app renders at most 2048 (S10-34), so Go's downscale is a guard |
| Auditing thumbnails | Not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Import checks | `core/ingest/file.go` `File` | Inline: trimmed absolute path (`ErrInvalid`), `os.Lstat` (`ErrMissing`, `mapOpenErr`), symlink (`ErrSymlink`), regular file (`ErrNotAFile`), size 0 (`ErrEmpty`), size over `maxBytes` (`ErrTooLarge.WithParams(formatByteSize(n))`); then `openSource` (`O_NOFOLLOW` on Unix, `open_unix.go`) and the same three checks on the opened fd. |
| Import stream | `core/ingest/file.go` `streamSource`, `installObject`, `syncDir` | `streamSource` copies to `objects/.ingest-*` through SHA-256 and a 512-byte sniff, capped by the package `maxBytes`. `installObject(tmpPath, objPath, checksum)` renames into place unless `objectstore.MatchesChecksum` already matches, always removes the temp file, and reports `wrote`. |
| Import write | `core/ingest/file.go` `File` | `files.LookupByChecksum` before `writes.Run` ("deadlocks while a Run transaction is open"); a hit reuses the File with a `reuse_file` revision; otherwise `writes.Run(Op{Action: "create_file"})` inserts the row with a `file` create change. On a failed write it removes the object it wrote. |
| Package doc | `core/ingest/file.go` | "Ingest does not decode or execute file content." |
| Limits | `core/ingest/file.go` `MaxBytes` (512 MiB), `maxBytes` (test override); `core/derivatives/ensure.go` `maxSourceBytes` (128 MiB) | — |
| Errors | `core/ingest/file.go` `Err*`; `core/apperr/apperr.go` `CodeIngest*` | No image-format error in `ingest`. |
| Bounded decode | `core/derivatives/raster/raster.go` `EncodeJPEG`, `supportedFormats`, `maxDecodePixels` (64 MP), `ErrTooLarge`, `ErrUnsupportedFormat`, `scaleToMaxEdge`, `MaxEdgeLimit` (4096) | Inline in `EncodeJPEG`: `image.DecodeConfig`, format in `supportedFormats` (jpeg, png, gif, bmp, tiff, webp), declared pixels within budget, then `image.Decode`, scale, transforms, `jpeg.Encode`. No PNG encoder. |
| Decode semaphore | `core/derivatives/ensure.go` `decodeSem`, `decodeSlots`, `generate` | `generate` holds one of `min(GOMAXPROCS, 4)` slots across read, decode, and encode. |
| Effects | `core/database/effects/registry.go` `"files"` entry | `create_file` changes resolve; `ingest.File` already records them. |
| Tests | `core/ingest/file_test.go` `TestFile`, `TestMapOpenErr`; `core/ingest/file_unix_test.go`; `core/derivatives/raster/raster_test.go` `TestEncodeJPEGRejects`, `TestEncodeJPEGAtBudgetBoundaries`; `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources` | Cover the import checks, the decode budget, and unsafe thumbnail sources. |

**Precedent.** `ingest.File` is the pattern for "check, stream to a temp object, look up by checksum before the write, insert inside the write, remove the object on failure". `StageImage` splits that sequence at the write so a caller's own `writes.Run` can carry the insert. `90b4114f` ("share object helpers") is the precedent for moving a shared helper rather than copying it.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `StageImage` lives in `core/ingest`, not `core/derivatives`. The package doc becomes: "Ingest copies files into the object store. `File` does not decode content; `StageImage` decodes an image only through `raster.Decode` and stores Go's own re-encode." | A depiction image is a research File, not a disposable derivative (working rule, 2026-10-16). `ingest` already owns the path checks, the object install, and `create_file`, so the checks stay private and are not copied. | `core/derivatives` — its doc and invariant 18 say derivatives are reproducible and disposable. A new package — it would need the checks exported. |
| D-2 | `openChecked(absPath string, limit int64) (*os.File, error)` *(new, unexported)* in `core/ingest/file.go` holds every pre-read check (absolute path, Lstat, symlink, regular, empty, size over `limit`, `openSource`, and the fd re-checks) with today's errors. `File` calls it with `maxBytes`, and `streamSource` keeps its own copy cap at `maxBytes`. | One copy of each check (map review K-5). | Exporting the checks — no caller outside `ingest`. |
| D-3 | `ingest.MaxImageBytes int64 = 32 << 20` *(new)*. `StageImage` reads at most `MaxImageBytes + 1` bytes and refuses more with `ErrTooLarge`. | A 2048 × 2048 RGBA PNG is about 17 MiB at worst; 32 MiB leaves room and keeps a hostile file small. | 512 MiB (import) or 128 MiB (thumbnails) — far above any crop. |
| D-4 | `raster.Decode(src []byte, formats []string) (image.Image, error)` *(new)*: `image.DecodeConfig`; the format must be in `formats` (nil means `supportedFormats`) and in `supportedFormats`, else `ErrUnsupportedFormat`; declared pixels within `maxDecodePixels`, else `ErrTooLarge`; then `image.Decode`. `EncodeJPEG` calls `Decode(src, nil)` and keeps its option checks, scale, transforms, and encode. `raster.EncodePNG(src []byte, opts Options) ([]byte, error)` *(new)*: the same option checks for `MaxEdge`, then `Decode(src, opts.Formats)`, `scaleToMaxEdge`, transforms, `png.Encode`. `Options` gains `Formats []string` *(new)*; `Quality` is ignored by `EncodePNG`. | One bounded decode for thumbnails and staging. The format check runs before any full decode, so the TIFF, WebP, and BMP decoders never see a staged file. | A second decode in `ingest` — a second copy of the decompression-bomb guard. |
| D-5 | `raster.Slot() (release func())` *(new)* moves `decodeSem` and `decodeSlots` from `core/derivatives` into `raster`. `derivatives.generate` calls `defer raster.Slot()()` in place of its channel send and receive. `StageImage` holds a slot across its read, decode, and encode. | One bound on concurrent decode memory for both callers. | A second semaphore in `ingest` — two pools could together exceed the bound. |
| D-6 | `StageImage(c *database.Catalog, absPath string) (*StagedImage, error)`: `openChecked(absPath, MaxImageBytes)`, read, `raster.EncodePNG(raw, raster.Options{MaxEdge: 2048, Formats: []string{"png", "jpeg"}})`, SHA-256 of the PNG, `files.StorageRelPath(checksum, "image/png")`, then write the PNG to a temp file in the objects directory and `installObject`. It records whether this call wrote the object, and looks up an existing `files` row by checksum with `files.LookupByChecksum` (outside any write). It must not be called inside `writes.Run`. A `raster.ErrUnsupportedFormat`, `raster.ErrTooLarge`, or decode error maps to `ErrUnsupportedImage`. | The client sends PNG (S10-34); JPEG is accepted so an image Artifact's pixels can be passed through. Every decode happens before the write. | Accepting every raster format — more decoders on an untrusted path for no caller. |
| D-7 | `ErrUnsupportedImage = apperr.New(apperr.CodeIngestUnsupportedImage, apperr.KindUser)` *(new)*, code `"ingest.unsupported_image"` *(new)* in `core/apperr/apperr.go` beside `CodeIngestMissing`. | One user-facing reason for "not a PNG or JPEG we can safely decode". Swift copy for it is S10-54, when it can reach the app. | Reusing `ErrUnsupportedType` — that one means a disallowed import media type, with its own help text. |
| D-8 | `(*StagedImage).InsertTx(tx *database.Tx) (files.File, []rowchange.Change, error)`: if the stage found an existing row, or a `SELECT … FROM files WHERE checksum_sha256 = ?` on `tx` finds one, return it and no change. Otherwise `files.Insert(tx.Tx, files.File{ID: new, ChecksumSHA256, MediaType: "image/png", ByteSize})` and return a `file` create change with the same fields `ingest.File` records (no `original_filename`). A second call on the same stage returns the same File and no change. | The caller's write carries the row and its audit. Lookup through `tx` does not deadlock (only the pool does). | Inserting in a separate `writes.Run` — a crash between the two writes would leave an unreferenced File. |
| D-9 | `(*StagedImage).Discard()`: if this stage wrote the object and `files.LookupByChecksum(c, checksum)` finds no row, remove the object. Call it after the caller's `writes.Run` returns an error. It is a no-op after a committed insert. | Same rule as the purge: never remove bytes a row still names. | Removing unconditionally — a concurrent write could have stored the same pixels. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ingest.openChecked` *(new)* | func | the inline checks in `ingest.File` | Extracted, not copied; `File` calls it. |
| `ingest.StageImage`, `StagedImage`, `InsertTx`, `Discard` *(new)* | func / type | `ingest.File` | `File` runs its own write and stores the client's bytes; staging splits at the caller's write and stores Go's bytes. Both share `openChecked` and `installObject`. |
| `ingest.MaxImageBytes`, `ingest.ErrUnsupportedImage`, `apperr.CodeIngestUnsupportedImage` *(new)* | const / error | `MaxBytes`, `ErrUnsupportedType` | Different limit and meaning (D-3, D-7). |
| `raster.Decode` *(new)* | func | the inline decode in `raster.EncodeJPEG` | Extracted; `EncodeJPEG` calls it. |
| `raster.EncodePNG`, `Options.Formats` *(new)* | func / field | `raster.EncodeJPEG`, `Options` | Same shape, PNG out. |
| `raster.Slot` *(new)* | func | `derivatives.decodeSem` | Moved; the `derivatives` copy is deleted. |
| `core/ingest/stage.go`, `core/ingest/stage_test.go` *(new)* | files | `core/ingest/file.go`, `file_test.go` | Keeps `file.go` about imports. |

**Builds on.** `ingest.File`'s check-stream-install-insert sequence; `raster`'s allowlist and budget; `writes.Run` change lists.
**Deletes.** The inline checks in `ingest.File`, the inline decode in `raster.EncodeJPEG`, and `derivatives.decodeSem` / `decodeSlots`.
**Next adds.** S10-32 calls `StageImage` for each depiction input before its write, `InsertTx` inside it, and `Discard` on failure. A later signature or place picture uses the same three calls.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/ingest` | `openChecked` (D-2) in `file.go`; `StageImage`, `StagedImage`, `InsertTx`, `Discard`, `MaxImageBytes`, `ErrUnsupportedImage` (D-3, D-6 to D-9) in `stage.go` *(new)*; package doc (D-1). |
| `core/derivatives/raster` | `Decode`, `EncodePNG`, `Options.Formats`, `Slot` (D-4, D-5). |
| `core/derivatives` | `generate` uses `raster.Slot`; `decodeSem` and `decodeSlots` deleted. |
| `core/apperr` | `CodeIngestUnsupportedImage` (D-7). |

```go
// core/ingest/stage.go (sketch)
const MaxImageBytes int64 = 32 << 20
var ErrUnsupportedImage = apperr.New(apperr.CodeIngestUnsupportedImage, apperr.KindUser)

type StagedImage struct { /* c, checksum, relPath, byteSize, wrote, existing */ }
func StageImage(c *database.Catalog, absPath string) (*StagedImage, error)
func (s *StagedImage) InsertTx(tx *database.Tx) (files.File, []rowchange.Change, error)
func (s *StagedImage) Discard()
```

Errors: `ingest.unsupported_image` → the staged file is not a PNG or JPEG within the pixel budget. Path and size errors are `ingest`'s existing ones.

### FFI

None — no RPC; S10-54 adds the field that reaches `StageImage`.

### macOS

None — no Swift change.

### Strings

None — Swift copy for `ingest.unsupported_image` is S10-54.

## 7. Steps

**STEP-1 — Share the import checks**
Files: `core/ingest/file.go`.
Do: D-2. Move the checks into `openChecked` and call it from `File`. No behavior change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/...` passes unedited.

**STEP-2 — One bounded decode and one decode slot**
Files: `core/derivatives/raster/raster.go`, `core/derivatives/raster/raster_test.go`, `core/derivatives/ensure.go`.
Do: D-4 and D-5. Add T-1 and T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/derivatives/...` passes, with the existing tests unedited.

**STEP-3 — Stage an image**
Files: `core/apperr/apperr.go`, `core/ingest/stage.go` *(new)*, `core/ingest/stage_test.go` *(new)*.
Do: D-3, D-6, D-7. Add T-3, T-4, T-5. Build fixtures in the test with `image/png`, `image/jpeg`, and `golang.org/x/image/tiff` encoders and by hand-writing a PNG header (as `raster_test.go` does for its budget cases); no binary test files are checked in.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/...` passes.

**STEP-4 — Insert inside a write, or discard**
Files: `core/ingest/stage.go`, `core/ingest/stage_test.go` (both created in STEP-3, *(new)*).
Do: D-8 and D-9. Add T-6, T-7, T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/... ./core/derivatives/...` passes.

**STEP-5 — Docs**
Files: `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `source-layer-data-model.md` §6 (`files`): after the paragraph on stored objects, add "A File can also hold an image the app rendered and Go re-encoded (`ingest.StageImage`): Go checks the path as for an import, decodes only PNG or JPEG within the pixel budget, and stores its own PNG, never the app's bytes."
- `deployment-plan.md`: the S10-53 **Plan:** line and SLICE 8 line read "— done, #N"; `completed.md`: a `## S10-53 — Stage a client image` section at the top in the S10-50 shape.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-53-stage-image.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/derivatives/raster/raster_test.go` `TestDecode` *(new)*, table-driven | PNG with `formats` `["png","jpeg"]` → ok; TIFF with the same list → `ErrUnsupportedFormat`; TIFF with `nil` → ok; a PNG header declaring 100000 × 100000 → `ErrTooLarge` | AC-1.2, AC-1.3, D-4 |
| T-2 | Go | `raster_test.go` `TestEncodePNG` *(new)*, table-driven | 4000 × 1000 JPEG, `MaxEdge` 2048 → PNG 2048 × 512; 300 × 200 PNG → PNG 300 × 200; `MaxEdge` 0 → error | AC-1.5, D-4 |
| T-3 | Go | `core/ingest/stage_test.go` `TestStageImageRefusesPaths` *(new)*, table-driven | each AC-1.1 input → the named error; the `objects/` directory has no new file | AC-1.1, D-2, D-3 |
| T-4 | Go | `stage_test.go` `TestStageImageRefusesContent` *(new)*, table-driven | TIFF, GIF, WebP header bytes, BMP, HTML named `.png`, random bytes, the 100000 × 100000 PNG header → `ErrUnsupportedImage`; no new object | AC-1.2, AC-1.3, D-6, D-7 |
| T-5 | Go | `stage_test.go` `TestStageImageReencodes` *(new)* | 300 × 200 PNG plus 1 KiB of ZIP bytes → staged object decodes to 300 × 200, differs from the input bytes, and ends at the PNG IEND chunk; 4000 × 1000 JPEG → 2048 × 512 PNG | AC-1.4, AC-1.5, D-6 |
| T-6 | Go | `stage_test.go` `TestStagedImageInsertTx` *(new)* | stage, `writes.Run` with `InsertTx` → one `files` row (`image/png`, empty `original_filename`), object present, one `file` create change; same stage inserted twice in one write → one row, second call no change | AC-2.1, AC-2.3, D-8 |
| T-7 | Go | `stage_test.go` `TestStagedImageDiscard` *(new)* | stage, write function returns an error, `Discard` → no row, no object; stage pixels already stored by a committed File, `Discard` → object stays | AC-2.2, AC-2.4, D-9 |
| T-8 | Go | existing `core/ingest/file_test.go`, `core/derivatives/raster/raster_test.go`, `core/derivatives/thumbnail_test.go` | unedited, still pass | AC-3.1, AC-3.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestFile` (STEP-1), `TestEncodeJPEGRejects` and `TestEnsureThumbnailRejectsUnsafeSources` (STEP-2).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none in this PR)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI in this PR)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated: source-layer-data-model.md §6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing calls `StageImage` until S10-32 and S10-54. Instead, run `CGO_ENABLED=1 go test -tags fts5 -run 'TestStageImage|TestStagedImage|TestDecode|TestEncodePNG' -v ./core/ingest/... ./core/derivatives/raster/...` and paste the result in the PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving the decode semaphore changes thumbnail concurrency | `raster.Slot` keeps `decodeSlots`' size; T-8 keeps the thumbnail tests unedited. |
| `golang.org/x/image` decoders on an untrusted path | D-4 and D-6 refuse every format but PNG and JPEG before full decode. |
| Nearest-neighbour downscale looks rough | The app renders at most 2048 (S10-34), so Go scales only a misbehaving client's image. |

**Stop and report if:**

- An existing test in `core/ingest` or `core/derivatives` needs an edit to pass after STEP-1 or STEP-2.
- `image.DecodeConfig` fully decodes a format (so the allowlist would run after a decoder already allocated): check with the 100000 × 100000 header case in T-1.
- `files.LookupByChecksum` blocks when `Discard` is called after a failed `writes.Run` returns.
- The effects registry refuses a `file` create change inside a write whose `Op.Action` is not `create_file`; report the Run's error.

## 12. Open questions

None.

## 13. PR

**Title:** S10-53 — Stage a client image
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-53-stage-image.md`.
