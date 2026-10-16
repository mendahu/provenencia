# S10-53 — Stage a client image

**Status:** Blocked — on S10-51
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-53--stage-a-client-image)
**Brief:** — (no new chrome: an engine API with no caller until S10-32)
**Board:** —
**Depends on:** S10-51 (`files.PurgeUnreferenced`; not landed)
**Blocks:** S10-32
**Branch:** `s10-53-stage-image`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) "The image", [`docs/source-layer-data-model.md`](../../../source-layer-data-model.md) §6, [`.agents/skills/review-pr/lenses/security.md`](../../../../.agents/skills/review-pr/lenses/security.md) trust boundaries, working rule "The client renders the image; Go validates it and stores its own copy", map review [`../deployment-plan.review.md`](../deployment-plan.review.md) K-4, K-5
**Planned:** 2026-10-16 against `main` @ `98819f86`, with [S10-51](S10-51-file-release.md) as planned
**Review:** [`S10-53-stage-image.review.md`](S10-53-stage-image.review.md) — Fits, Ready with changes (applied); 2 rounds

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

Go can accept an image the Mac app rendered and store it safely as a File. The app writes a crop to a temporary file and hands Go its path, the way Artifact import already does. Go applies the checks every import gets, decodes the image with one bounded codec (PNG or JPEG only, within the pixel budget), re-encodes its own PNG at most 2048 pixels on the long edge, and stores that PNG, never the app's bytes. The work runs in two phases, so no decode happens while the catalog write is open, and a failed write leaves nothing on disk.

**Why now.** S10-32 stores each likeness image this way. The user asked that a client-sent file get every protection a file import gets (2026-10-16). Today the import checks live inline in `ingest.File`, which stores the client's bytes and never decodes; the bounded decode lives inline in `raster.EncodeJPEG` under `core/derivatives`; and writing bytes into the object store is hand-rolled in `derivatives`. S10-32 would otherwise copy all three.

**What stays.** `ingest.File` accepts and refuses exactly what it does today, with the same errors and audit. Thumbnails are generated exactly as today. No schema, FFI, or Swift change.

## 2. User stories

**US-1.** As a researcher, a picture the app sends to be stored is checked the way an imported file is, so a bad file cannot get in by that path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `StageImage` on a relative path, a missing path, a symlink, a directory, an empty file, and (with the test cap lowered to 1 KiB) a 2 KiB PNG | Refused with `ingest.ErrInvalid`, `ErrMissing`, `ErrSymlink`, `ErrNotAFile`, `ErrEmpty`, and `ErrTooLarge`, the same errors `ingest.File` returns (`errors.Is`). No object is written. |
| AC-1.2 | A real TIFF, GIF, and BMP encoding (from `golang.org/x/image/tiff`, `image/gif`, `golang.org/x/image/bmp`), an HTML file named `.png`, and random bytes | Refused with `ingest.ErrUnsupportedImage`; the TIFF, GIF, and BMP never reach a full decode. No object is written. |
| AC-1.3 | A PNG whose header declares 100000 × 100000 pixels | Refused with `ingest.ErrUnsupportedImage` before a full decode. |
| AC-1.4 | A valid 300 × 200 PNG followed by 1 KiB of appended ZIP bytes | Staged. The stored object is a PNG that decodes to 300 × 200, is not byte-equal to the input, and ends at its `IEND` chunk. |
| AC-1.5 | A valid 4000 × 1000 JPEG | Staged as a PNG of 2048 × 512. |

**US-2.** As a contributor, the staging API lets a write store the image with its row, and a failed write leaves nothing behind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Stage a valid PNG; inside `writes.Run`, `InsertTx`; the write commits | One `files` row with media type `image/png`, no original filename, and its object on disk. The change list holds one `file` create change. |
| AC-2.2 | Stage a valid PNG; the write function returns an error; the deferred `Discard` runs | No `files` row and no object. |
| AC-2.3 | Stage the same PNG twice; insert both in one write | One `files` row; the second `InsertTx` returns the same File and no change. |
| AC-2.4 | Stage an image whose re-encode equals the object of a `files` row that already exists; `Discard` | The existing object and row stay. |
| AC-2.5 | Stage, insert, commit; the deferred `Discard` runs | Nothing is removed. |

**US-3.** As a contributor, imports and thumbnails behave as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Every existing `ingest` test | Passes unedited. |
| AC-3.2 | Every existing `raster` test and thumbnail test | Passes with only the import path changed (STEP-2 moves the package). |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ingest.openChecked` *(new)* and `ingest.insertFile` *(new)*, shared by `ingest.File` and staging | STEP-1, AC-1.1, AC-2.1, AC-3.1 |
| `core/derivatives/raster` moves to `core/raster` and becomes the one bounded image codec: private `decode` and `prepare`, `Options.Formats`, `EncodePNG`, `Slot` | STEP-2, AC-1.2, AC-1.3, AC-1.5, AC-3.2 |
| `objectstore.WriteIfAbsent` and `objectstore.SyncDir` *(new)*; `derivatives.writeObjectIfAbsent` and both `syncDir` copies deleted | STEP-3, AC-2.1, AC-3.2 |
| `files.LookupByChecksumTx` *(new)* | STEP-4, AC-2.3 |
| `ingest.StageImage`, `StagedImage`, `InsertTx`, `Discard`, `ErrUnsupportedImage` | STEP-5, STEP-6, AC-1.1 to AC-2.5 |
| Docs, including the security lens trust table | STEP-7 |

| Out | Home |
| --- | --- |
| Calling `StageImage` from an Observation write | S10-32 |
| The FFI field that carries the crop path, and Swift copy for `ingest.unsupported_image` | S10-54 |
| Rendering the crop in the app | S10-34 |
| Moving `ingest.lookupFile`'s inline `SELECT … FROM files WHERE id = ?` into `core/database/files` | Logged in the review log; not planned |
| One helper for the three `io.LimitReader(…, cap+1)` reads | Logged in the review log; not planned |
| Moving `ingest.installObject` into `objectstore` | Not planned; only `ingest` installs a streamed temp file |
| A Go PDF renderer | Not planned ([depictions and likenesses](../depictions-and-likenesses.md)) |
| A better downscale filter than nearest-neighbour | Not planned; the app renders at most 2048 (S10-34), so Go's downscale is a guard |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Import checks | `core/ingest/file.go` `File` | Inline: trimmed absolute path (`ErrInvalid`), `os.Lstat` (`ErrMissing`, `mapOpenErr`), symlink (`ErrSymlink`), regular file (`ErrNotAFile`), size 0 (`ErrEmpty`), size over `maxBytes` (`ErrTooLarge.WithParams(formatByteSize(n))`); then `openSource` (`O_NOFOLLOW` on Unix, `open_unix.go`) and the same three checks on the opened fd. |
| Import stream | `core/ingest/file.go` `streamSource`, `installObject`; `core/ingest/sync_unix.go`, `sync_other.go` `syncDir` | `streamSource` copies to `objects/.ingest-*` through SHA-256 and a 512-byte sniff, capped by `maxBytes`. `installObject` renames into place unless `objectstore.MatchesChecksum` already matches, removes the temp file, and reports `wrote`. |
| Import write | `core/ingest/file.go` `File` (closure around lines 163–192), `lookupFile` | `files.LookupByChecksum` before `writes.Run` ("deadlocks while a Run transaction is open"). Inside `writes.Run(Op{Action: "create_file"})`: `files.Insert`, `files.IsUniqueConflict` → `errChecksumTaken`, and an inline `fields` map (`id`, `checksum_sha256`, `byte_size`, optional `original_filename`, optional `media_type`) in a `rowchange.Change{EntityType: "file", Action: ActionCreate}`. On a failed write it removes the object it wrote. |
| Media policy | `core/ingest/mediatypes/mediatypes.go` `Resolve`, `Allowed` | The import allowlist; `image/png` and `image/jpeg` are allowed. |
| Package doc | `core/ingest/file.go` lines 1–4 | Describes checksum reuse, `reuse_file`, `IngestedAs`, and ends "Ingest does not decode or execute file content." |
| Limits | `core/ingest/file.go` `MaxBytes` (512 MiB) and the test-override `var maxBytes`; `core/derivatives/ensure.go` `maxSourceBytes` (128 MiB) | — |
| Errors | `core/ingest/file.go` `Err*`; `core/apperr/apperr.go` `CodeIngest*` (`CodeIngestMissing` last) | No image-format error in `ingest`. No Go↔Swift error-code parity test exists; unknown codes fall to the default callout. |
| Codec | `core/derivatives/raster/raster.go` `EncodeJPEG`, `Options{MaxEdge, Quality, Transforms}`, `Supported`, `supportedMIME`, `supportedFormats`, `maxDecodePixels` (64 MP), `ErrTooLarge`, `ErrUnsupportedFormat`, `scaleToMaxEdge`, `MaxEdgeLimit` (4096), package doc "turns image bytes into derivative JPEGs" | Inline in `EncodeJPEG`: option checks, `image.DecodeConfig`, format in `supportedFormats`, declared pixels within budget, `image.Decode`, scale, transforms, `jpeg.Encode`. No PNG encoder. Importers: `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`. |
| Decode slot | `core/derivatives/ensure.go` `decodeSem`, `decodeSlots`, `generate` | `generate` holds one of `min(GOMAXPROCS, 4)` slots across read, decode, and encode. |
| Byte write | `core/derivatives/ensure.go` `writeObjectIfAbsent`, `sha256Hex`; `core/derivatives/sync_unix.go`, `sync_other.go` `syncDir` | Match by checksum, `os.CreateTemp(dir, ".deriv-*")`, write, sync, close, rename, re-match on a rename failure, `syncDir`. |
| Object store | `core/objectstore/objectstore.go` `MatchesChecksum` | Only verification lives here. |
| Files queries | `core/database/files/files.go` `sqlLookupChecksum`, `scanFile`, `isLowerHex`, `LookupByChecksum(c, …)`, `Insert(tx, f)`, `IsUniqueConflict` | No `*sql.Tx` checksum lookup. Convention elsewhere: `properties.LookupTx(tx *sql.Tx, …)`. |
| Purge | [S10-51](S10-51-file-release.md) D-1 `files.PurgeUnreferenced(c, objs []files.Object)` | Not landed. Removes an object whose checksum no `files` row names. |
| Security lens | `.agents/skills/review-pr/lenses/security.md` trust table | "Ingest: `core/ingest` …" (no decode); "Derivatives: `core/derivatives` (thumbnails, crops) — Image and PDF bytes decoded in-process". |
| Tests | `core/ingest/file_test.go` `TestFile` (helpers `mustUser`, `writeTemp`, `auditCount` are closures inside it), `TestMapOpenErr`; `core/ingest/file_unix_test.go`; `core/derivatives/raster/raster_test.go` `TestEncodeJPEGRejects`, `TestEncodeJPEGAtBudgetBoundaries`, `TestSupported`, `TestEncodeJPEGOptionsAndTransform`; `core/derivatives/thumbnail_test.go` | — |

**Precedent.** `ingest.File` is the "check, stage an object, look up by checksum before the write, insert inside the write, remove the object on failure" sequence; `StageImage` splits it at the caller's write. `90b4114f` ("share object helpers") moved object helpers into `core/objectstore`; STEP-3 does the same for `writeObjectIfAbsent` and `syncDir`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `StageImage` lives in `core/ingest`. The package doc keeps its first two sentences and its last becomes: "`File` does not decode or execute file content. `StageImage` decodes an image only through `raster.EncodePNG` and stores Go's own re-encode, never the caller's bytes." | A depiction image is a research File, not a disposable derivative. `ingest` already owns the path checks and `create_file`, so they stay private. | `core/derivatives` — its doc and invariant 18 say derivatives are disposable. |
| D-2 | `openChecked(absPath string, limit int64) (*os.File, error)` *(new, unexported)* in `file.go` holds every pre-read check with today's errors. `c.DB()`, the `strings.TrimSpace` of `absPath`, and `database.RequireUserID` stay in `File`, in today's order; `File` passes the trimmed path, and `openChecked` starts at the absolute-path test, so error precedence and `IngestedAs` are unchanged. `File` calls it with `maxBytes`. `File`'s failure cleanup (`if wrote { _ = os.Remove(objPath) }`) becomes `files.PurgeUnreferenced(c, []files.Object{{ChecksumSHA256: checksum, MediaType: mediaType}})`, so object bytes are removed only by the purge (review M-13). `insertFile(tx *sql.Tx, row files.File) (rowchange.Change, error)` *(new, unexported)* in `file.go` is `files.Insert` plus today's `fields` map and change; it returns `errChecksumTaken` on `files.IsUniqueConflict`. `File`'s closure calls it with `tx.Tx`. No behavior change to `File`. | One copy of each check and of the audited File-create shape (review M-3, K-5). | Copying either into `stage.go`. |
| D-3 | The staged-file cap: `const imageMaxBytes int64 = 32 << 20` and `var maxImageBytes = imageMaxBytes` (test override), both unexported. Over the cap is `ErrTooLarge` from `openChecked`, with the size in its params like an import. | No caller outside `ingest` needs the number. The override keeps tests from writing 32 MiB. A 2048 × 2048 RGBA PNG is about 17 MiB at worst. | An exported constant with no outside caller. |
| D-4 | `core/derivatives/raster` moves to `core/raster` (`git mv`, update two importers). Its package doc becomes: "Package raster is the one bounded image codec: it decodes only allowlisted formats within a pixel budget, and re-encodes. Thumbnails and staged images both use it." The existing "Security: golang.org/x/image decoders" paragraph stays. Inside it: private `decode(src []byte, formats []string) (image.Image, error)` (`DecodeConfig`; format in `formats`, or in `supportedFormats` when `formats` is nil, and always in `supportedFormats`, else `ErrUnsupportedFormat`; pixels within `maxDecodePixels`, else `ErrTooLarge`; then `image.Decode`) and private `prepare(src []byte, opts Options) (image.Image, error)` (the `MaxEdge` checks, `decode(src, opts.Formats)`, `scaleToMaxEdge`, transforms). `EncodeJPEG` = Quality check + `prepare` + `jpeg.Encode`. `EncodePNG(src []byte, opts Options) ([]byte, error)` *(new)* = `prepare` + `png.Encode`; it ignores `Quality`. `Options.Formats []string` *(new)*, honored by both. | The research layer must not depend on a subpackage of the disposable layer, and one codec means one decompression-bomb guard (review M-5, M-6). The format check runs before any full decode. | Exporting `decode` — no outside caller. A second guard in `ingest`. |
| D-5 | `raster.Slot() (release func())` *(new)* holds `decodeSem` and `decodeSlots`, moved from `derivatives`. `derivatives.generate` uses `defer raster.Slot()()`. `StageImage` takes `release := raster.Slot()` right after `openChecked` succeeds and releases it after `EncodePNG` returns, so the slot covers the read and the decode and encode, not the hash or the object write. | One bound on decode memory for both callers. | A second semaphore in `ingest`. |
| D-6 | `objectstore.WriteIfAbsent(objPath string, data []byte) (wrote bool, err error)` *(new)* is `derivatives.writeObjectIfAbsent` moved, returning whether this call wrote the object; its temp files are `.write-*`, it hashes with its own `sha256Hex` (the `derivatives` copy stays for `readSourceObject`), and it ignores the `SyncDir` error as `writeObjectIfAbsent` does today. The package doc becomes "Package objectstore verifies and writes content-addressed object files on disk." `objectstore.SyncDir(dir string) error` *(new)* holds the moved `syncDir` (with `sync_unix.go` / `sync_other.go` moved alongside into `core/objectstore`). `derivatives` and `ingest` call them; their `syncDir` copies and `writeObjectIfAbsent` are deleted. | One atomic, checksum-matched byte write and one directory sync (review M-4). | A third hand-rolled write in `ingest`. |
| D-7 | `files.LookupByChecksumTx(tx *sql.Tx, checksumHex string) (File, error)` *(new)* beside `LookupByChecksum`, reusing `sqlLookupChecksum`, `scanFile`, and the hex check. | Queries live in `core/database/<table>/` (`add-catalog-query`); the `Tx` variant follows `properties.LookupTx`. | Inline SQL in `ingest`. |
| D-8 | `stageFormats = []string{"png", "jpeg"}` *(new, unexported)* in `stage.go` is the staged decode allowlist and replaces the import sniff policy for staging. It must stay inside `mediatypes.Allowed` (T-4 checks `image/png` and `image/jpeg`). | The decode allowlist is stricter than the import policy today; the test keeps a future widening from admitting a type imports refuse (review M-7). | Also running `mediatypes.Resolve` — redundant once the bytes are decoded and re-encoded. |
| D-9 | `StageImage(c *database.Catalog, absPath string) (*StagedImage, error)`, never called inside `writes.Run`: `c.DB()` (as `File` does), trim the path, `openChecked(absPath, maxImageBytes)`; take the slot; `io.ReadAll(io.LimitReader(f, maxImageBytes+1))`, more than the cap → `ErrTooLarge`; `raster.EncodePNG(raw, raster.Options{MaxEdge: 2048, Formats: stageFormats})`; release the slot; `raster.ErrUnsupportedFormat`, `raster.ErrTooLarge`, or any decode error → `ErrUnsupportedImage`; SHA-256 of the PNG; `files.StorageRelPath(checksum, "image/png")`; `objectstore.WriteIfAbsent(filepath.Join(c.Dir(), filepath.FromSlash(rel)), png)` gives `wrote`; `files.LookupByChecksum(c, checksum)` (`sql.ErrNoRows` means none). Any error after `WriteIfAbsent` reports `wrote` true (including its own error) builds the `StagedImage`, calls its `Discard`, and returns `(nil, err)`, so a failed stage leaves no object. | Every decode and disk write happens before the caller's write. The client sends PNG (S10-34); JPEG is accepted so an image Artifact's pixels can be passed through. | Accepting every raster format — more decoders on an untrusted path for no caller. |
| D-10 | `type StagedImage struct { c *database.Catalog; checksum string; byteSize int64; wrote bool; file files.File; haveFile bool }`. `StageImage` sets `file`/`haveFile` when a row already exists. `(*StagedImage).InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)` (a `*sql.Tx`, because S10-32 calls it from `observations.resolveValue`, which holds one): if `haveFile`, return `file` and no change; else `files.LookupByChecksumTx(tx, checksum)` → found: cache it and return it, no change; else `id, err := files.NewID()` then `insertFile(tx, files.File{ID: id, ChecksumSHA256: checksum, MediaType: "image/png", ByteSize: byteSize})` → cache it (`haveFile = true`) and return it with its change. `errChecksumTaken` from `insertFile` is returned as is. | The caller's write carries the row and its audit; a second `InsertTx` on the same stage is free. | Inserting in its own `writes.Run` — a crash between writes would leave an unreferenced File. |
| D-11 | `(*StagedImage).Discard()`: if `wrote`, `files.PurgeUnreferenced(s.c, []files.Object{{ChecksumSHA256: s.checksum, MediaType: "image/png"}})`; otherwise nothing. Safe to call more than once and after a commit (a row then names the checksum, so the purge keeps it). The calling pattern is `s, err := ingest.StageImage(c, path); if err != nil {…}; defer s.Discard()`. | One rule for "never remove bytes a row still names": S10-51's purge (review M-1, M-9). | A second remove-if-unreferenced in `ingest`. |
| D-12 | `ErrUnsupportedImage = apperr.New(apperr.CodeIngestUnsupportedImage, apperr.KindUser)` *(new)*; `CodeIngestUnsupportedImage = "ingest.unsupported_image"` *(new)* after `CodeIngestMissing`. Tests compare errors with `errors.Is` (`ingest.ErrTooLarge` and `raster.ErrTooLarge` are different errors). | One user-facing reason for "not a PNG or JPEG we can safely decode". | Reusing `ErrUnsupportedType` — it means a disallowed import media type, with its own help text. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ingest.openChecked`, `ingest.insertFile` *(new)* | func | the inline checks and closure in `ingest.File` | Extracted, not copied; `File` calls both. |
| `ingest.StageImage`, `StagedImage`, `InsertTx`, `Discard`, `stageFormats`, `imageMaxBytes`, `maxImageBytes` in `core/ingest/stage.go` *(new)* | func / type / var | `ingest.File` | Splits the import sequence at the caller's write and stores Go's bytes. |
| `ingest.ErrUnsupportedImage`, `apperr.CodeIngestUnsupportedImage` *(new)* | error | `ErrUnsupportedType` | Different meaning (D-12). |
| `core/raster` (moved), `decode`, `prepare`, `EncodePNG`, `Options.Formats`, `Slot` *(new)* | package / func | `core/derivatives/raster` `EncodeJPEG`, `derivatives.decodeSem` | Moved and extracted; `EncodeJPEG` shares `prepare`. |
| `objectstore.WriteIfAbsent`, `objectstore.SyncDir` *(new)* | func | `derivatives.writeObjectIfAbsent`, two `syncDir` copies | Moved; the copies are deleted. |
| `files.LookupByChecksumTx` *(new)* | func | `files.LookupByChecksum`, `properties.LookupTx` | Same query on a transaction. |
| `core/ingest/stage_test.go` *(new)*, package-level `stageUser`, `stageWrite` helpers | test helpers | the closures inside `TestFile` | `file_test.go` stays unedited (AC-3.1). |

**Builds on.** `ingest.File`, `raster`'s allowlist and budget, `objectstore`, S10-51's `files.PurgeUnreferenced`, `writes.Run` change lists.
**Deletes.** The inline checks and change map in `ingest.File`, the inline decode in `EncodeJPEG`, `core/derivatives/raster` (moved), `derivatives.decodeSem` / `decodeSlots` / `writeObjectIfAbsent`, `core/ingest/sync_*.go`, `core/derivatives/sync_*.go`.
**Next adds.** S10-32 calls `StageImage` per depiction input before its write with `defer s.Discard()`, and `InsertTx` inside it. A later signature or place picture uses the same calls.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/ingest` | D-1, D-2, D-3, D-8 to D-12; `stage.go` *(new)*. |
| `core/raster` (moved from `core/derivatives/raster`) | D-4, D-5. |
| `core/derivatives` | Imports `core/raster`; `generate` uses `raster.Slot`; uses `objectstore.WriteIfAbsent` and `objectstore.SyncDir`. |
| `core/objectstore` | D-6; `sync_unix.go`, `sync_other.go` *(moved)*. |
| `core/database/files` | D-7. |
| `core/apperr` | D-12. |

```go
// core/ingest/stage.go (sketch)
type StagedImage struct {
	c        *database.Catalog
	checksum string
	byteSize int64
	wrote    bool
	file     files.File
	haveFile bool
}
func StageImage(c *database.Catalog, absPath string) (*StagedImage, error)
func (s *StagedImage) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)
func (s *StagedImage) Discard()
```

Errors: `ingest.unsupported_image` → not a PNG or JPEG within the pixel budget. Path and size errors are `ingest`'s existing ones.

### FFI

None — S10-54 adds the field that reaches `StageImage`.

### macOS

None — no Swift change.

### Strings

None — Swift copy for `ingest.unsupported_image` is S10-54.

## 7. Steps

**STEP-1 — Share the import checks and the File-create change**
Files: `core/ingest/file.go`.
Do: D-2. No behavior change.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/...` passes unedited.

**STEP-2 — One bounded codec in `core/raster`**
Files: `core/derivatives/raster/` moved to `core/raster/` with `git mv` (`core/raster/raster.go`, `core/raster/raster_test.go` (new) after the move), `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: D-4 and D-5. `core/raster/raster_test.go` moves unchanged (it is `package raster`), then gains T-1 and T-2; `core/derivatives/thumbnail_test.go` changes only its import path. (new) after the move
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster/... ./core/derivatives/...` passes.

**STEP-3 — One byte write in `objectstore`**
Files: `core/objectstore/objectstore.go`, `core/objectstore/sync_unix.go` and `core/objectstore/sync_other.go` (moved from `core/derivatives/`, *(new)* at this path), `core/objectstore/objectstore_test.go` *(new)*, `core/derivatives/ensure.go`, `core/ingest/file.go`, `core/ingest/sync_unix.go`, `core/ingest/sync_other.go` (deleted).
Do: D-6, including the `objectstore` package doc. Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/... ./core/ingest/... ./core/derivatives/...` passes.

**STEP-4 — Look up a File by checksum on a transaction**
Files: `core/database/files/files.go`, `core/database/files/files_test.go`. Skill: `add-catalog-query`.
Do: D-7. Add T-4a.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/files/...` passes.

**STEP-5 — Stage an image**
Files: `core/apperr/apperr.go`, `core/ingest/stage.go` *(new)*, `core/ingest/stage_test.go` *(new)*.
Do: D-1, D-3, D-5, D-8, D-9, D-12. In `stage_test.go` define package-level helpers `stageUser(t, c) []byte` (a user row, as `TestFile`'s `mustUser` does) and `stageWrite(t, dir, name string, data []byte) string` (writes a file and returns its absolute path); do not edit `file_test.go`. Build fixtures in the test with real encoders (`image/png`, `image/jpeg`, `image/gif`, `golang.org/x/image/tiff`, `golang.org/x/image/bmp`) and a hand-written PNG header for the 100000 × 100000 case, as `raster_test.go` does; no binary fixtures are checked in. Add T-4, T-5, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/...` passes.

**STEP-6 — Insert inside a write, or discard**
Files: `core/ingest/stage.go`, `core/ingest/stage_test.go` (both (new) in STEP-5).
Do: D-10 and D-11. Add T-7 and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/... ./core/database/files/...` passes.

**STEP-7 — Docs**
Files: `docs/source-layer-data-model.md`, `.agents/skills/review-pr/lenses/security.md`, `.agents/skills/ship-pr/SKILL.md`, `docs/ideas/large-jpeg-thumbnails.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `source-layer-data-model.md` §6: change "The stored object is the **original file bytes**" to "The stored object is the **original file bytes** of an import"; after that paragraph add "A File can also hold an image the app rendered (a depiction, S10-32). Go checks its path as for an import, decodes only PNG or JPEG within the pixel budget (`core/raster`), and stores its own PNG, never the app's bytes (`ingest.StageImage`)."
- `security.md` trust table: the Ingest row's Where gains "`ingest.StageImage` decodes and re-encodes app-rendered images"; the Derivatives row becomes "`core/derivatives` (thumbnails), `core/raster` (the one bounded codec, also used by staging)".
- `ship-pr` skill: the Security row of the routing table (around line 100) adds `core/raster`.
- `docs/ideas/large-jpeg-thumbnails.md`: repoint its link to `core/raster/raster.go`, and add "An ImageIO fallback goes in `EncodeJPEG`, not in the shared `decode`, so staged images never take it." (new) after the move
- `deployment-plan.md`: the S10-53 **Plan:** line's status becomes "— Landed #N", and " — done, #N" is appended to its SLICE 8 line (as `S10-50  … — done, #413`). `completed.md`: a `## S10-53 — Stage a client image` section above the current first section, in the shape of the `S10-50` section.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-53-stage-image.md` passes; `grep -rn "derivatives/raster" core .agents docs/ideas` returns nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/raster/raster_test.go` (new) after the move, `TestDecode` *(new, in-package)*, table-driven | PNG with formats `["png","jpeg"]` → ok; real TIFF with the same list → `ErrUnsupportedFormat`; real TIFF with `nil` → ok; PNG header declaring 100000 × 100000 → `ErrTooLarge` | AC-1.2, AC-1.3, D-4 |
| T-2 | Go | `raster_test.go` `TestEncodePNG` *(new)*, table-driven | 4000 × 1000 JPEG, `MaxEdge` 2048 → PNG 2048 × 512; 300 × 200 PNG → PNG 300 × 200; `MaxEdge` 0 → error; GIF with `Formats: ["png"]` → `ErrUnsupportedFormat`; `EncodeJPEG` with `Formats: ["png"]` on a JPEG → `ErrUnsupportedFormat` | AC-1.5, D-4 |
| T-3 | Go | `core/objectstore/objectstore_test.go` `TestWriteIfAbsent` *(new)* | absent → written, `wrote` true; same bytes again → `wrote` false; a different file at the path → replaced, `wrote` true | D-6 |
| T-4a | Go | `core/database/files/files_test.go` `TestLookupByChecksumTx` *(new)* | inserted row found inside the same tx; unknown checksum → `sql.ErrNoRows`; bad hex → `ErrInvalid` | D-7 |
| T-4 | Go | `core/ingest/stage_test.go` `TestStageImageRefuses` *(new)*, table-driven | every AC-1.1 input (`maxImageBytes` set to 1024 for the size row, restored after) and every AC-1.2 and AC-1.3 input → the named error by `errors.Is`; the objects directory has no new file. Plus: for each `f` in `stageFormats`, `mediatypes.Allowed("image/"+f)` is true | AC-1.1, AC-1.2, AC-1.3, D-3, D-8, D-12 |
| T-5 | Go | `stage_test.go` `TestStageImageReencodes` *(new)* | 300 × 200 PNG plus 1 KiB of ZIP bytes → staged object decodes to 300 × 200, differs from the input, ends at `IEND`; 4000 × 1000 JPEG → 2048 × 512 PNG | AC-1.4, AC-1.5, D-9 |
| T-6 | Go | existing `core/ingest/file_test.go`, `file_unix_test.go`, moved `core/raster/raster_test.go` (new) after the move, `core/derivatives/thumbnail_test.go` | pass with only import paths changed | AC-3.1, AC-3.2 |
| T-7 | Go | `stage_test.go` `TestStagedImageInsertTx` *(new)* | stage, `writes.Run(c, writes.Op{Action: "create_file", UserID: stageUser(t, c)}, …)` whose function returns `s.InsertTx(tx.Tx)`'s File and changes → one `files` row (`image/png`, empty `original_filename`), object present, one `file` create change; the same stage inserted twice in one write → one row, second call no change | AC-2.1, AC-2.3, D-10 |
| T-8 | Go | `stage_test.go` `TestStagedImageDiscard` *(new)*, table-driven | stage, write function returns an error, `Discard` → no row, no object; stage, `InsertTx`, commit, then stage the same input again (so a row exists), `Discard` → object stays; stage, insert, commit, `Discard` twice → object stays | AC-2.2, AC-2.4, AC-2.5, D-11 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestFile` (STEP-1), the moved raster tests and `TestEnsureThumbnailRejectsUnsafeSources` (STEP-2), thumbnail writes (STEP-3).

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
- [ ] Domain doc updated: source-layer-data-model.md §6, security lens trust table
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing calls `StageImage` until S10-32 and S10-54. Instead, run `CGO_ENABLED=1 go test -tags fts5 -run 'TestStageImage|TestStagedImage|TestDecode|TestEncodePNG|TestWriteIfAbsent|TestLookupByChecksumTx' -v ./core/ingest/... ./core/raster/... ./core/objectstore/... ./core/database/files/...` and paste the result in the PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| S10-51 lands with a different `files.PurgeUnreferenced` shape | This plan waits on S10-51; refresh D-11 against its Landed note. |
| Moving the decode slot changes thumbnail concurrency | `raster.Slot` keeps `decodeSlots`' size; T-6 keeps the thumbnail tests. |
| Two stages of the same pixels, one discarded before the other commits | `catalogsession.Do` runs a project's operations one at a time, so FFI callers cannot interleave; `ingest.File` has the same theoretical exposure today. Accepted. |
| `golang.org/x/image` decoders on an untrusted path | D-4 and D-8 refuse every format but PNG and JPEG before a full decode. |

**Stop and report if:**

- An existing test in `core/ingest` needs an edit to pass after STEP-1, or a raster or thumbnail test needs more than an import-path edit after STEP-2.
- `image.DecodeConfig` allocates the declared pixels for any format (check with the 100000 × 100000 header in T-1).
- `files.LookupByChecksum` blocks when `Discard` runs after a failed `writes.Run` returns.
- `files.PurgeUnreferenced` is not in `core/database/files` when you start (S10-51 has not landed).

## 12. Open questions

None.

## 13. PR

**Title:** S10-53 — Stage a client image
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-53-stage-image.md`.
