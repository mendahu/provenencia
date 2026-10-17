# S10-53 — Stage a client image

**Status:** Landed in [#424](https://github.com/mendahu/provenencia/pull/424) (2026-10-16)
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-53--stage-a-client-image)
**Brief:** — (no new chrome: an engine API with no caller until S10-32)
**Board:** —
**Depends on:** S10-55 (landed in #423: `files.PurgeUnreferenced`, `files.PurgeAfterCommit`, `objectstore.RemoveObject`, `Tx.Catalog()`)
**Blocks:** S10-32
**Branch:** `s10-53-stage-image`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-catalog-delete`](../../../../.agents/skills/add-catalog-delete/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/catalog-deletes.mdc`, `.cursor/rules/go-tests.mdc`
**Context:** [depictions and likenesses](../depictions-and-likenesses.md) "The image", [`docs/source-layer-data-model.md`](../../../source-layer-data-model.md) §6, [`docs/catalog-deletes.md`](../../../catalog-deletes.md), [`.agents/skills/review-pr/lenses/security.md`](../../../../.agents/skills/review-pr/lenses/security.md) trust boundaries, working rules "The client renders the image; Go validates it and stores its own copy" and "A File is released only when nothing points at it, and one purge removes its object", map review [`../deployment-plan.review.md`](../deployment-plan.review.md) K-4, K-5, S10-55 ship log [`S10-55-file-release.ship.md`](S10-55-file-release.ship.md) MNT-1, SEC-O2
**Planned:** 2026-10-16 against `main` @ `1c99ef4e` (S10-55 landed); first planned against `98819f86`
**Review:** [`S10-53-stage-image.review.md`](S10-53-stage-image.review.md) — refresh: Fits, Ready with changes (applied); 2 rounds; M-1 answered by the user

## Landed

Differences from this plan: `objectstore.WriteIfAbsent` hashes `data` and returns an error when it does not match `checksum` (T-3 requires it, D-6 said it would not hash); the 2048 long edge is a named constant, `stageMaxEdge`, beside `stageFormats`; T-9 reuses the package's `testCatalog` and `runArtifactCreate` helpers instead of building its own fixture. The **Current state** rows that name `core/ingest/sync_*.go`, `core/derivatives/sync_*.go`, and `core/derivatives/raster` describe the code before this PR moved them, so their paths are no longer in code spans (the plan-docs check requires existing paths).

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

Go can accept an image the Mac app rendered and store it safely as a File. The app writes a crop to a temporary file and hands Go its path, the way Artifact import already does. Go applies the checks every import gets, decodes the image with one bounded codec (PNG or JPEG only, within the pixel budget), re-encodes its own PNG at most 2048 pixels on the long edge, and stores that PNG, never the app's bytes. The decode runs before the catalog write; the staged object is written inside the write that commits its `files` row, so a stage that is never inserted, or a failed write, leaves nothing on disk.

Two rules from S10-55's review land with it (user, 2026-10-16). **Releases schedule their own purge:** every exported release in `deleteimpact` takes the write's `*database.Tx` and schedules the after-commit purge itself, so no caller can drop the objects a release freed. **A delete cannot race an import of the same bytes:** every production path that writes or purges an object (imports, thumbnails, deletes, and S10-32's staging) runs inside `catalogsession.Do`, which runs a project's operations one at a time, so no purge can fall between another operation's object write and its commit. This PR names that guarantee where the purge and the data model describe object removal, and adds no second serializer (user, 2026-10-16, narrowing S10-55 SEC-O2 after review M-1).

**Why now.** S10-32 stores each likeness image this way and is the first update path that releases a File (`ReleaseReplaced`). The user asked that a client-sent file get every protection a file import gets (2026-10-16). Today the import checks live inline in `ingest.File`, which stores the client's bytes and never decodes; the bounded decode lives inline in `raster.EncodeJPEG` under `core/derivatives`; writing bytes into the object store is hand-rolled in `derivatives`; and six deletes pair a release with a hand-written `files.PurgeAfterCommit`.

**What stays.** `ingest.File` accepts and refuses exactly what it does today, with the same errors and audit. Thumbnails are generated exactly as today. Every delete removes and audits exactly what it does today. No schema, FFI, or Swift change.

## 2. User stories

**US-1.** As a researcher, a picture the app sends to be stored is checked the way an imported file is, so a bad file cannot get in by that path.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `StageImage` on a relative path, a missing path, a symlink, a directory, an empty file, and (with the test cap lowered to 1 KiB) a 2 KiB PNG | Refused with `ingest.ErrInvalid`, `ErrMissing`, `ErrSymlink`, `ErrNotAFile`, `ErrEmpty`, and `ErrTooLarge`, the same errors `ingest.File` returns (`errors.Is`). No object is written. |
| AC-1.2 | A real TIFF, GIF, and BMP encoding (from `golang.org/x/image/tiff`, `image/gif`, `golang.org/x/image/bmp`), an HTML file named `.png`, and random bytes | Refused with `ingest.ErrUnsupportedImage`. No object is written. (The format check before a full decode is T-1's.) |
| AC-1.3 | A PNG whose header declares 100000 × 100000 pixels | Refused with `ingest.ErrUnsupportedImage` before a full decode. |
| AC-1.4 | A valid 300 × 200 PNG followed by 1 KiB of appended ZIP bytes; staged, inserted, committed | The stored object is a PNG that decodes to 300 × 200, is not byte-equal to the input, and ends at its `IEND` chunk. |
| AC-1.5 | A valid 4000 × 1000 JPEG; staged, inserted, committed | Stored as a PNG of 2048 × 512. |
| AC-1.6 | A valid PNG; `StageImage` returns | No new file under the project's objects folder (the object is written inside the write, AC-2.1). |

**US-2.** As a contributor, the staging API lets a write store the image with its row, and a failed write leaves nothing behind.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Stage a valid PNG; inside `writes.Run`, `InsertTx`; the write commits | One `files` row with media type `image/png`, no original filename, and its object on disk. The change list holds one `file` create change. |
| AC-2.2 | Stage a valid PNG; `InsertTx`; the write function then returns an error; the deferred `Discard` runs | No `files` row and no object. |
| AC-2.3 | Stage the same PNG twice; insert both in one write | One `files` row; the second `InsertTx` returns the same File and no change. |
| AC-2.4 | A `files` row and object for a PNG already exist; stage the same PNG; insert it in a write that fails; `Discard` | The existing object and row stay. |
| AC-2.5 | Stage, insert, commit; the deferred `Discard` runs twice | Nothing is removed. |

**US-3.** As a contributor, imports, thumbnails, and deletes behave as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Every existing test in `core/ingest` | Passes unedited. |
| AC-3.2 | Every existing `raster` test and thumbnail test | Passes with only the import path changed (STEP-2 moves the package). |
| AC-3.3 | Every existing delete test in `core/database/artifacts`, `observations`, `subjects`, `sources`, `citations`, `metadatafields`, `promote`, `identityclaims` | Passes with only the release call signatures changed (STEP-7). |

**US-4.** As a contributor, a release cannot free a File without removing its bytes (S10-55 MNT-1).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | An Artifact with a File and its object; inside `writes.Call`, `UPDATE artifacts SET file_id = NULL`, then `deleteimpact.ReleaseReplaced(tx, "artifacts.file_id", fileID)`; the write commits | The `files` row and the object are gone. The caller passed no object list. |
| AC-4.2 | The same, but the write function returns an error | The `files` row and the object stay. |
| AC-4.3 | The same Artifact; inside `writes.Call`, `SnapshotOwned`, `DELETE FROM artifacts`, `ReleaseSnapshot(tx, snap)`; commit | The object is gone. |
| AC-4.4 | Any package outside `core/database/deleteimpact` | Has no way to read the objects a release freed (`Released` is unexported; the exported releases return no `files.Object`). |

**US-5.** As a contributor, the rule that keeps a purge from racing an import is written where the purge lives (S10-55 SEC-O2).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | `core/catalogsession/session_test.go` `TestDo` case "serializes ops on one session" | Passes unedited; it is the test that pins the guarantee. |
| AC-5.2 | Read `files.PurgeUnreferenced`'s doc comment and data model §6 | Both name `catalogsession.Do` as what keeps a purge from falling between an object write and its row's commit. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `ingest.openChecked` *(new)* and `ingest.insertFile` *(new)*, shared by `ingest.File` and staging | STEP-1, AC-1.1, AC-2.1, AC-3.1 |
| `core/derivatives/raster` moves to `core/raster` and becomes the one bounded image codec: private `decode` and `prepare`, `Options.Formats`, `EncodePNG`, `Slot` | STEP-2, AC-1.2, AC-1.3, AC-1.5, AC-3.2 |
| `objectstore.WriteIfAbsent` and `objectstore.SyncDir` *(new)*; `derivatives.writeObjectIfAbsent` and both `syncDir` copies deleted | STEP-3, AC-2.1, AC-3.2 |
| `files.LookupByChecksumTx` *(new)* | STEP-4, AC-2.3 |
| `ingest.StageImage`, `StagedImage`, `InsertTx`, `Discard`, `ErrUnsupportedImage`; `ingest.File`'s failure cleanup moves onto `files.PurgeUnreferenced` | STEP-1, STEP-5, STEP-6, AC-1.1 to AC-2.5 |
| Every exported release in `deleteimpact` takes `*database.Tx` and schedules its own purge; `Released` unexported; six deletes and `Unpromote` updated; `identityclaims.UnfileBridgesTx` becomes `UnfileBridges(tx *database.Tx, …)` | STEP-7, AC-4.1 to AC-4.4, AC-3.3 |
| `files.PurgeUnreferenced`'s doc comment names `catalogsession.Do` as the guarantee. Docs: data model §6, catalog deletes, `deleteimpact` README, `add-catalog-delete`, the catalog-deletes rule, the security lens trust table, the `ship-pr` routing table, `docs/ideas/large-jpeg-thumbnails.md`, the map's S10-53 status and SLICE 8 lines, `completed.md` | STEP-8, AC-5.2 |

| Out | Home |
| --- | --- |
| Calling `StageImage` from an Observation write | S10-32 |
| S10-32's plan text that still names `ReleaseReplaced`'s returned objects and `files.PurgeAfterCommit` (its D-9 `releaseReplacedValue`). After D-13, `observations.updateOne` and S10-32's `releaseReplacedValue` take `*database.Tx` (`Update` passes `tx`), because `ReleaseReplaced` needs it | S10-32 refreshes its plan against this PR's Landed note |
| The FFI field that carries the crop path, and Swift copy for `ingest.unsupported_image` | S10-54 |
| Rendering the crop in the app | S10-34 |
| `derivatives.Ensure` removing its own object when its write fails | Not planned; a thumbnail is disposable and orphan bytes are harmless |
| A core-level race guard independent of `catalogsession` (the purge holding the one connection; `ingest.File` and `derivatives.Ensure` writing objects inside the write) | Not planned (user, 2026-10-16): `catalogsession.Do` is the guarantee. Revisit if any object write or purge runs outside it |
| `derivatives.Ensure` looking up the derived File inside its write and dropping the `errDerivedFileExists` retry (review M-5) | Not planned; logged in the review log |
| Opening the `objects` root through the project root to close the `Lstat` / `OpenRoot` gap in `RemoveObject` (S10-55 SEC-R2-1) | Not planned here; S10-55 ship log follow-up |
| One constant for the `"objects"` folder name (S10-55 MNT-R2-1) | Not planned; logged in S10-55's ship log |
| Moving `ingest.lookupFile`'s inline `SELECT … FROM files WHERE id = ?` into `core/database/files` | Logged in the review log; not planned |
| One helper for the three `io.LimitReader(…, cap+1)` reads | Logged in the review log; not planned |
| Moving `ingest.installObject` into `objectstore` | Not planned; only `ingest` installs a streamed temp file |
| A Go PDF renderer | Not planned ([depictions and likenesses](../depictions-and-likenesses.md)) |
| A better downscale filter than nearest-neighbour | Not planned; the app renders at most 2048 (S10-34), so Go's downscale is a guard |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Import checks | `core/ingest/file.go` `File` | Inline: trimmed absolute path (`ErrInvalid`), `os.Lstat` (`ErrMissing`, `mapOpenErr`), symlink (`ErrSymlink`), regular file (`ErrNotAFile`), size 0 (`ErrEmpty`), size over `maxBytes` (`ErrTooLarge.WithParams(formatByteSize(n))`); then `openSource` (`O_NOFOLLOW` on Unix, `open_unix.go`) and the same three checks on the opened fd. |
| Import stream | `core/ingest/file.go` `streamSource`, `installObject`; core/ingest/sync_unix.go, `sync_other.go` `syncDir` | `streamSource` copies to `objects/.ingest-*` through SHA-256 and a 512-byte sniff, capped by `maxBytes`. `installObject` renames into place unless `objectstore.MatchesChecksum` already matches, always removes the temp file, and reports `wrote`. |
| Import write | `core/ingest/file.go` `File`, `reuseFile`, `lookupFile` | `files.LookupByChecksum` before `writes.Run` ("deadlocks while a Run transaction is open"). Found → `installObject`, then `reuseFile`. Not found → **`installObject` before the write**, then `writes.Run(Op{Action: "create_file"})`: `files.Insert`, `files.IsUniqueConflict` → `errChecksumTaken`, and an inline `fields` map (`id`, `checksum_sha256`, `byte_size`, optional `original_filename`, optional `media_type`) in a `rowchange.Change{EntityType: "file", Action: ActionCreate}`. `errChecksumTaken` → `LookupByChecksum` + `reuseFile`. Any other failure → `os.Remove(objPath)` when it wrote. |
| Media policy | `core/ingest/mediatypes/mediatypes.go` `Resolve`, `Allowed` | The import allowlist; `image/png` and `image/jpeg` are allowed. |
| Package doc | `core/ingest/file.go` lines 1–4 | Describes checksum reuse, `reuse_file`, `IngestedAs`, and ends "Ingest does not decode or execute file content." |
| Limits | `core/ingest/file.go` `MaxBytes` (512 MiB) and the test-override `var maxBytes`; `core/derivatives/ensure.go` `maxSourceBytes` (128 MiB) | — |
| Errors | `core/ingest/file.go` `Err*`; `core/apperr/apperr.go` `CodeIngest*` (`CodeIngestMissing` last) | No image-format error in `ingest`. No Go↔Swift error-code parity test exists; unknown codes fall to the default callout. |
| Codec | core/derivatives/raster/raster.go `EncodeJPEG`, `Options{MaxEdge, Quality, Transforms}`, `Supported`, `supportedMIME`, `supportedFormats`, `maxDecodePixels` (64 MP), `ErrTooLarge`, `ErrUnsupportedFormat`, `scaleToMaxEdge`, `MaxEdgeLimit` (4096), package doc "turns image bytes into derivative JPEGs" | Inline in `EncodeJPEG`: option checks, `image.DecodeConfig`, format in `supportedFormats`, declared pixels within budget, `image.Decode`, scale, transforms, `jpeg.Encode`. No PNG encoder. Importers: `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`. |
| Decode slot | `core/derivatives/ensure.go` `decodeSem`, `decodeSlots`, `generate` | `generate` holds one of `min(GOMAXPROCS, 4)` slots across read, decode, and encode. |
| Thumbnail write | `core/derivatives/ensure.go` `Ensure` (around lines 140–200), `writeObjectIfAbsent`, `sha256Hex`; core/derivatives/sync_unix.go, `sync_other.go` `syncDir` | `Ensure` calls **`writeObjectIfAbsent(derivedPath, derivedJPEG)` before `writes.Run`**, then looks up the derived File by checksum, then inside the write inserts the derived `files` row (if new) and the `file_derivatives` link. `writeObjectIfAbsent`: match by checksum, `os.CreateTemp(dir, ".deriv-*")`, write, sync, close, rename, re-match on a rename failure, `syncDir`. A failed write leaves the object. |
| Object store | `core/objectstore/objectstore.go` `MatchesChecksum`, `RemoveObject`; `core/objectstore/objectstore_test.go` `TestRemoveObject` (`package objectstore_test`) | Package doc "verifies and removes content-addressed object files on disk." No write. |
| Files queries | `core/database/files/files.go` `sqlLookupChecksum`, `scanFile`, `isLowerHex`, `LookupByChecksum(c, …)` (uses `c.DB()`), `Insert(tx, f)`, `IsUniqueConflict` | No `*sql.Tx` checksum lookup. Convention elsewhere: `properties.LookupTx(tx *sql.Tx, …)`. |
| Purge | `core/database/files/purge.go` `Object`, `PurgeUnreferenced(c, objs)`, `PurgeAfterCommit(tx *database.Tx, objs)`; `purge_test.go` `TestPurgeUnreferenced` | `PurgeUnreferenced` calls `LookupByChecksum(c, …)` per object, then `objectstore.RemoveObject(c.Dir(), rel)` when no row names the checksum. The lookup and the removal are separate pool checkouts, so a write can begin and commit between them. Its doc says "S10-53 moves ingest.File's rollback and StagedImage.Discard onto this function." `PurgeAfterCommit` schedules it with `tx.AfterCommit`; a non-empty list on a Tx with no catalog panics. |
| Serialization | `core/catalogsession/session.go` `Do` (around line 120); `api/ffi/handlers/catalog.go` `withProjectCatalog`; `api/ffi/handlers/artifacts.go` (around line 159, `ingest.File`), `api/ffi/handlers/thumbnails.go` (`derivatives.EnsureThumbnail`); `docs/application-stack.md` (around line 310); `core/catalogsession/session_test.go` `TestDo` | `Do` holds the session lock across its function, so a project's operations run one at a time. Every production object write and purge runs inside it; nothing in `core/` or `api/` starts a goroutine. `writes.Run` does not take the session lock. `PurgeUnreferenced`'s doc does not mention it. |
| Release API | `core/database/deleteimpact/facets.go` `Released{Changes, Objects}`, `ReleaseFacets(tx *sql.Tx, kind, id) (Released, error)`, `ReleaseClaim(tx *sql.Tx, claimID) (Released, error)`, `releaseSubjectClaims`, `eraseObservation` (calls `ReleaseFacets`, `SnapshotOwned`, `ReleaseSnapshot`); the `facetRelease.Release` field type, `pinRelease`, `releaseConnectionFacets`, and an inline release closure (around line 426), all in `facets.go`, return `Released`; `core/database/deleteimpact/release.go` `SnapshotOwned(tx *sql.Tx, …)`, `ReleaseSnapshot(tx *sql.Tx, snap) ([]files.Object, error)`, `ReleaseOwned(tx *sql.Tx, kind, id) ([]files.Object, error)`, `ReleaseReplaced(tx *sql.Tx, column, oldID) ([]files.Object, error)`, private `releaseReplaced`, `releaseOwnedSnap` | Every release returns the objects it freed; the caller must pass them on. |
| Release callers | `core/database/artifacts/artifacts.go` `Delete` (`ReleaseSnapshot` then `files.PurgeAfterCommit(tx, objects)`); `core/database/observations/write.go` `Delete` (`ReleaseFacets`, `ReleaseSnapshot`, `PurgeAfterCommit(tx, append(released.Objects, owned...))`); `core/database/subjects/subjects.go`, `core/database/sources/sources.go`, `core/database/citations/citations.go`, `core/database/metadatafields/metadatafields.go` `Delete` (`ReleaseFacets(tx.Tx, …)` then `PurgeAfterCommit(tx, released.Objects)`); `core/database/promote/unpromote.go` `Unpromote` and `core/database/identityclaims/bridgefile.go` `UnfileBridgesTx(tx *sql.Tx, sourceID, subjectID)` (`ReleaseClaim`, objects dropped) | Six hand-written pairings and two callers that drop the objects (claims own no File today). |
| Release tests | `core/database/deleteimpact/release_test.go` (`package deleteimpact`, calls `ReleaseSnapshot` near line 201, `ReleaseOwned` near line 348); `impact_test.go` (`package deleteimpact_test`, `ReleaseFacets` near line 399 and `ReleaseSnapshot` near lines 629 and 654 on a `db.Begin()` tx); `facets_test.go` `TestReleaseFacetsReportsHandles` (near line 170); `core/database/promote/unpromote_test.go` (`UnfileBridgesTx(tx.Tx, …)` near line 150) | — |
| Delete docs | `docs/catalog-deletes.md` items 1, 2, the "One call per delete" line, and recipe step 6; `core/database/deleteimpact/README.md`; `.agents/skills/add-catalog-delete/SKILL.md` (checklist line about `files.PurgeAfterCommit`, the sketch near line 75–79, the owned-FK paragraph near line 96); `.cursor/rules/catalog-deletes.mdc` | All say each releasing delete calls `files.PurgeAfterCommit` itself. |
| Security lens | `.agents/skills/review-pr/lenses/security.md` trust table | "Ingest: `core/ingest` …" (no decode); "Derivatives: `core/derivatives` (thumbnails, crops) — Image and PDF bytes decoded in-process". |
| Map | `docs/deployment-plan/spike-10/deployment-plan.md` working rule "The client renders the image" (around line 138), the S10-53 entry (around line 652), SLICE 8 line (around line 263) | The working rule and the entry's In and Testable cells were refreshed with this plan and need no edit. The entry's **Plan:** status line and the SLICE 8 line change at landing (STEP-8). |
| Tests | `core/ingest/file_test.go` `TestFile` (helpers `mustUser`, `writeTemp`, `auditCount` are closures inside it), `TestMapOpenErr`; `core/ingest/file_unix_test.go`; core/derivatives/raster/raster_test.go `TestEncodeJPEGRejects`, `TestEncodeJPEGAtBudgetBoundaries`, `TestSupported`, `TestEncodeJPEGOptionsAndTransform`; `core/derivatives/thumbnail_test.go` | — |

**Precedent.** `ingest.File` is the "check, stage, look up by checksum, insert inside the write, clean up on failure" sequence; `StageImage` splits it at the caller's write. `90b4114f` ("share object helpers") moved object helpers into `core/objectstore`; STEP-3 does the same for `writeObjectIfAbsent` and `syncDir`. S10-55 F-2's committing subtest in `core/database/files/purge_test.go` ("commit purges and rollback does not") is the shape for a test write that must commit with no audited change: it returns one `subject_position` create change. S10-55's `artifacts.Delete` dropping its `Catalog` parameter for `Tx.Catalog()` is the precedent for D-13.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `StageImage` lives in `core/ingest`. The package doc keeps its first two sentences and its last becomes: "`File` does not decode or execute file content. `StageImage` decodes an image only through `raster.EncodePNG` and stores Go's own re-encode, never the caller's bytes." | A depiction image is a research File, not a disposable derivative. `ingest` already owns the path checks and `create_file`, so they stay private. | `core/derivatives` — its doc and invariant 18 say derivatives are disposable. |
| D-2 | `openChecked(absPath string, limit int64) (*os.File, error)` *(new, unexported)* in `file.go` holds every pre-read check with today's errors. `openChecked` starts at the absolute-path test (`ErrInvalid`) and runs through the checks on the opened fd. `File` keeps `c.DB()`, the `strings.TrimSpace` of `absPath` (it needs the trimmed path for `IngestedAs`), and `database.RequireUserID`, now called before `openChecked`; both a bad path and a bad user are `ErrInvalid`, so no caller can see the swap. `File` passes the trimmed path. `File` calls it with `maxBytes`. `insertFile(tx *sql.Tx, row files.File) (rowchange.Change, error)` *(new, unexported)* in `file.go` is `files.Insert` plus today's `fields` map and change; it returns `errChecksumTaken` on `files.IsUniqueConflict`. `File`'s closure calls it with `tx.Tx`. `File`'s failure cleanup (`if wrote { _ = os.Remove(objPath) }`, both places) becomes `files.PurgeUnreferenced(c, []files.Object{{ChecksumSHA256: checksum, MediaType: mediaType}})` under the same `if wrote`, so object bytes are removed only by the purge (review M-13). `File` keeps today's order: the object is installed before `create_file`'s write. No behavior change to `File`. | One copy of each check and of the audited File-create shape (review M-3, K-5). | Copying either into `stage.go`. |
| D-3 | The staged-file cap: `const imageMaxBytes int64 = 32 << 20` and `var maxImageBytes = imageMaxBytes` (test override), both unexported. Over the cap is `ErrTooLarge` from `openChecked`, with the size in its params like an import. | No caller outside `ingest` needs the number. The override keeps tests from writing 32 MiB. A 2048 × 2048 RGBA PNG is about 17 MiB at worst. | An exported constant with no outside caller. |
| D-4 | `core/derivatives/raster` moves to `core/raster` (`git mv`, update two importers). Its package doc becomes: "Package raster is the one bounded image codec: it decodes only allowlisted formats within a pixel budget, and re-encodes. Thumbnails and staged images both use it." The existing "Security: golang.org/x/image decoders" paragraph stays. Inside it: private `checkMaxEdge(opts Options) error` (today's two `MaxEdge` checks and messages); private `decode(src []byte, formats []string) (image.Image, error)` (`DecodeConfig`; format in `formats`, or in `supportedFormats` when `formats` is nil, and always in `supportedFormats`, else `ErrUnsupportedFormat`; pixels within `maxDecodePixels`, else `ErrTooLarge`; then `image.Decode`) and private `prepare(src []byte, opts Options) (image.Image, error)` (`decode(src, opts.Formats)`, `scaleToMaxEdge`, transforms; no option checks). `EncodeJPEG` = `checkMaxEdge`, then today's Quality check, then `prepare`, then `jpeg.Encode`, so today's error order is kept. `EncodePNG(src []byte, opts Options) ([]byte, error)` *(new)* = `checkMaxEdge`, `prepare`, `png.Encode`; it ignores `Quality`. `Options.Formats []string` *(new)*, honored by both. | The research layer must not depend on a subpackage of the disposable layer, and one codec means one decompression-bomb guard (review M-5, M-6). The format check runs before any full decode. | Exporting `decode` — no outside caller. A second guard in `ingest`. |
| D-5 | `raster.Slot() (release func())` *(new)* holds `decodeSem` and `decodeSlots`, moved from `derivatives`. `derivatives.generate` uses `defer raster.Slot()()`. `StageImage` takes the slot right after `openChecked` succeeds and releases it with `defer` (D-9). | One bound on decode memory for both callers. The slot is a caller-held pairing, not taken inside `prepare`, because the raw read (up to 128 MiB for a thumbnail source, 32 MiB for a stage) is part of the memory it bounds. | A second semaphore in `ingest`. Taking the slot inside `prepare` (review M-8) — the read would run unbounded. |
| D-6 | `objectstore.WriteIfAbsent(objPath, checksum string, data []byte) (wrote bool, err error)` *(new)* is `derivatives.writeObjectIfAbsent` moved, taking the checksum its caller already holds (as `MatchesChecksum` does) and returning whether this call wrote the object; its temp files are `.write-*`, it does not hash `data` itself (the `derivatives` `sha256Hex` stays for `readSourceObject`), and it ignores the `SyncDir` error as `writeObjectIfAbsent` does today. The package doc becomes "Package objectstore verifies, writes, and removes content-addressed object files on disk." `objectstore.SyncDir(dir string) error` *(new)* holds the moved `syncDir` (with `sync_unix.go` / `sync_other.go` moved alongside into `core/objectstore`). `derivatives` and `ingest` call them; their `syncDir` copies and `writeObjectIfAbsent` are deleted. | One atomic, checksum-matched byte write and one directory sync (review M-4). | A third hand-rolled write in `ingest`. |
| D-7 | `files.LookupByChecksumTx(tx *sql.Tx, checksumHex string) (File, error)` *(new)* beside `LookupByChecksum`, reusing `sqlLookupChecksum`, `scanFile`, and the trim and hex check (`ErrInvalid`). | Queries live in `core/database/<table>/` (`add-catalog-query`); the `Tx` variant follows `properties.LookupTx`. | Inline SQL in `ingest`. |
| D-8 | `stageFormats = []string{"png", "jpeg"}` *(new, unexported)* in `stage.go` is the staged decode allowlist and replaces the import sniff policy for staging. It must stay inside `mediatypes.Allowed` (T-4 checks `image/png` and `image/jpeg`). | The decode allowlist is stricter than the import policy today; the test keeps a future widening from admitting a type imports refuse (review M-7). | Also running `mediatypes.Resolve` — redundant once the bytes are decoded and re-encoded. |
| D-9 | `StageImage(c *database.Catalog, absPath string) (*StagedImage, error)`, never called inside `writes.Run`: `c.DB()` (as `File` does), trim the path, `openChecked(absPath, maxImageBytes)`; `release := raster.Slot()` then `defer release()` (the slot covers the read, decode, and encode; the hash after it is cheap); `io.ReadAll(io.LimitReader(f, maxImageBytes+1))`, a read error returned as is, more than the cap → `ErrTooLarge.WithParams(formatByteSize(n))`; `raster.EncodePNG(raw, raster.Options{MaxEdge: 2048, Formats: stageFormats})`; `raster.ErrUnsupportedFormat`, `raster.ErrTooLarge`, or any decode error → `ErrUnsupportedImage`; SHA-256 of the PNG; `files.StorageRelPath(checksum, "image/png")`. It returns a `StagedImage` holding the PNG bytes. It writes nothing to disk and reads no catalog row. | Every decode happens before the caller's write; the disk write moves into the write (D-15). The client sends PNG (S10-34); JPEG is accepted so an image Artifact's pixels can be passed through. | Writing the object in `StageImage` — see D-15. Accepting every raster format — more decoders on an untrusted path for no caller. |
| D-10 | `type StagedImage struct { c *database.Catalog; png []byte; checksum, rel string; wrote bool; file files.File; haveFile bool }`. `(*StagedImage).InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)` (a `*sql.Tx`, because S10-32 calls it from `observations.resolveValue`, which holds one). If `haveFile`, return `file` and no change. Else: `files.LookupByChecksumTx(tx, checksum)`; found → keep it, no change; `sql.ErrNoRows` → `id, err := files.NewID()` then `ch, err := insertFile(tx, files.File{ID: id, ChecksumSHA256: checksum, MediaType: "image/png", ByteSize: int64(len(png))})`. Then, in both branches, `w, err := objectstore.WriteIfAbsent(filepath.Join(s.c.Dir(), filepath.FromSlash(s.rel)), s.checksum, s.png)`; `s.wrote = s.wrote \|\| w`; an error is returned as is (the caller's write rolls back). Then cache (`file`, `haveFile = true`) and return the File with its change (nil when found). `errChecksumTaken` from `insertFile` is returned as is. A stage serves one write: after that write fails, the caller calls `Discard` and stages again. | The caller's write carries the row, its audit, and its bytes; a second `InsertTx` on the same stage is free. Writing in both branches also repairs a missing object under an existing row, as `ingest.File`'s reuse path does. | Inserting in its own `writes.Run` — a crash between writes would leave an unreferenced File. |
| D-11 | `(*StagedImage).Discard()`: if `wrote`, `files.PurgeUnreferenced(s.c, []files.Object{{ChecksumSHA256: s.checksum, MediaType: "image/png"}})`; then `s.png = nil`. Safe to call more than once and after a commit (a row then names the checksum, so the purge keeps it). It is called after the write returns, never inside the write function (D-14), and inside `catalogsession.Do`, as every caller of a write is. The calling pattern is `s, err := ingest.StageImage(c, path); if err != nil {…}; defer s.Discard()` placed before the `writes.Run`. | One rule for "never remove bytes a row still names": S10-55's purge (review M-1, M-9). | A second remove-if-unreferenced in `ingest`. |
| D-12 | `ErrUnsupportedImage = apperr.New(apperr.CodeIngestUnsupportedImage, apperr.KindUser)` *(new)*; `CodeIngestUnsupportedImage = "ingest.unsupported_image"` *(new)* after `CodeIngestMissing`. Tests compare errors with `errors.Is` (`ingest.ErrTooLarge` and `raster.ErrTooLarge` are different errors). | One user-facing reason for "not a PNG or JPEG we can safely decode". | Reusing `ErrUnsupportedType` — it means a disallowed import media type, with its own help text. |
| D-13 | Every exported release in `deleteimpact` takes the write's `*database.Tx` and schedules the purge itself with `files.PurgeAfterCommit(tx, objs)`: `ReleaseFacets(tx *database.Tx, kind catalogmodel.Kind, id []byte) ([]rowchange.Change, error)`, `ReleaseClaim(tx *database.Tx, claimID []byte) ([]rowchange.Change, error)`, `ReleaseSnapshot(tx *database.Tx, snap OwnedSnapshot) error`, `ReleaseOwned(tx *database.Tx, kind catalogmodel.Kind, id []byte) error`, `ReleaseReplaced(tx *database.Tx, column string, oldID []byte) error`. A nil `tx` is `ErrInvalid`. Each wraps an unexported body on `tx.Tx` that keeps today's code and returns the objects: `releaseFacets`, `releaseClaim`, `releaseSnapshot` (today's exported `ReleaseSnapshot` body), `releaseOwned` (`SnapshotOwned` then `releaseSnapshot`, both on `*sql.Tx`), `releaseReplaced` (exists). Inside the package, `releaseSubjectClaims`, `eraseObservation`, and the facet registry call only the unexported bodies, so a nested release is purged once, by the outermost exported call. `Released` is renamed `released` (fields unchanged), including the `facetRelease.Release` field type. `SnapshotOwned` keeps `*sql.Tx` (it releases nothing). `identityclaims.UnfileBridgesTx` is renamed `identityclaims.UnfileBridges` and takes `*database.Tx` (a `Tx` suffix means `*sql.Tx`, as its sibling `FileBridgesTx`; its two callers hold a `*database.Tx`). The six deletes drop their `files.PurgeAfterCommit` lines and pass `tx` instead of `tx.Tx` to the release. `Unpromote` passes `tx` to `ReleaseClaim` and `UnfileBridges` and keeps `q := tx.Tx` for its other calls. | A caller cannot hold the objects, so it cannot drop them (S10-55 MNT-1); `Tx.Catalog()` removed the reason S10-55 D-1 kept the purge out of `deleteimpact`. `ReleaseReplaced` gets it before S10-32 is its first caller. | A test that greps for the pairing — it checks the habit, not the API. Keeping `*sql.Tx` and returning an error when a release frees a File — a check at run time for what the type can rule out. |
| D-14 | `files.PurgeUnreferenced`'s doc comment: "Object bytes are removed only here (S10-53 moves ingest.File's rollback and StagedImage.Discard onto this function)." becomes "Object bytes are removed only here. Callers run inside catalogsession.Do (an AfterCommit callback runs inside the write that scheduled it), which runs a project's operations one at a time; that is what keeps a purge from falling between another operation's object write and its row's commit. Never call it inside a write function: its lookup waits for the connection the write holds. Inside a write, use PurgeAfterCommit." No code change to the purge. | The user chose the existing serializer as the guarantee (2026-10-16, review M-1); writing it at the purge tells the next caller the rule. | Holding the catalog's one connection across the purge's check and removal, and moving `ingest.File`'s and `derivatives.Ensure`'s object writes into their writes — a second serializer beside `catalogsession`, and a test hook in production code. |
| D-15 | `StageImage` writes nothing to disk; `InsertTx` writes the staged object inside the caller's write (D-10). `ingest.File` and `derivatives.Ensure` keep today's order (object first, then the write). | A stage the caller never inserts leaves nothing behind without relying on `Discard`, and the object and its row land in one write. | Writing the object in `StageImage` (the first plan) — a stage dropped without `Discard` leaves orphan bytes. |

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
| `deleteimpact.releaseFacets`, `releaseClaim`, `releaseSnapshot` *(new, unexported)* | func | today's exported bodies | Renamed bodies; the exported names become one-line wrappers that purge. |
| `core/ingest/stage_test.go` *(new)*, package-level `stageUser`, `stageWrite` helpers | test | the closures inside `TestFile` | `file_test.go` stays unedited (AC-3.1). |

**Builds on.** `ingest.File`, `raster`'s allowlist and budget, `objectstore`, S10-55's `files.PurgeUnreferenced` / `PurgeAfterCommit` / `Tx.Catalog()`, `catalogsession.Do` serialization, `writes.Run` change lists.
**Deletes.** The inline checks and change map in `ingest.File`, the inline decode in `EncodeJPEG`, `core/derivatives/raster` (moved), `derivatives.decodeSem` / `decodeSlots` / `writeObjectIfAbsent`, `core/ingest/sync_*.go`, `core/derivatives/sync_*.go`, the six hand-written `files.PurgeAfterCommit` calls in domain deletes, the exported `Released` type.
**Next adds.** S10-32 calls `StageImage` per depiction input before its write with `defer s.Discard()`, `InsertTx` inside it, and `deleteimpact.ReleaseReplaced(tx, "depiction_values.image_file_id", old)` with no object handling. A later signature or place picture uses the same calls. A new releasing delete calls the release and nothing else.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/ingest` | D-1, D-2, D-3, D-8 to D-12, D-15; `stage.go` *(new)*. |
| `core/raster` (moved from `core/derivatives/raster`) | D-4, D-5. |
| `core/derivatives` | Imports `core/raster`; `generate` uses `raster.Slot`; uses `objectstore.WriteIfAbsent` (same place as today) and `objectstore.SyncDir`. |
| `core/objectstore` | D-6; `sync_unix.go`, `sync_other.go` *(moved)*. |
| `core/database/files` | D-7, D-14. |
| `core/database/deleteimpact` | D-13. |
| `core/database/artifacts`, `observations`, `subjects`, `sources`, `citations`, `metadatafields`, `promote`, `identityclaims` | D-13 call sites. |
| `core/apperr` | D-12. |

```go
// core/ingest/stage.go (sketch)
type StagedImage struct {
	c        *database.Catalog
	png      []byte
	checksum string
	rel      string
	wrote    bool
	file     files.File
	haveFile bool
}
func StageImage(c *database.Catalog, absPath string) (*StagedImage, error)
func (s *StagedImage) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)
func (s *StagedImage) Discard()

// core/database/deleteimpact (sketch)
func ReleaseFacets(tx *database.Tx, kind catalogmodel.Kind, id []byte) ([]rowchange.Change, error)
func ReleaseClaim(tx *database.Tx, claimID []byte) ([]rowchange.Change, error)
func ReleaseSnapshot(tx *database.Tx, snap OwnedSnapshot) error
func ReleaseOwned(tx *database.Tx, kind catalogmodel.Kind, id []byte) error
func ReleaseReplaced(tx *database.Tx, column string, oldID []byte) error
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
Files: `core/derivatives/raster/` moved to `core/raster/` with `git mv` (`core/raster/raster.go` and `core/raster/raster_test.go` are (new) only as paths; their content moves), `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: D-4 and D-5. `core/raster/raster_test.go` (new) by the move arrives unchanged (it is `package raster`), then gains T-1 and T-2; `core/derivatives/thumbnail_test.go` changes only its import path.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/raster/... ./core/derivatives/...` passes.

**STEP-3 — One byte write in `objectstore`**
Files: `core/objectstore/objectstore.go`, `core/objectstore/sync_unix.go` and `core/objectstore/sync_other.go` (moved from `core/derivatives/`, *(new)* at this path), `core/objectstore/objectstore_test.go`, `core/derivatives/ensure.go`, `core/ingest/file.go`, `core/ingest/sync_unix.go`, `core/ingest/sync_other.go` (deleted).
Do: D-6, including the `objectstore` package doc. Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/... ./core/ingest/... ./core/derivatives/...` passes.

**STEP-4 — Look up a File by checksum on a transaction**
Files: `core/database/files/files.go`, `core/database/files/files_test.go`. Skill: `add-catalog-query`.
Do: D-7. Add T-4a.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/files/...` passes.

**STEP-5 — Stage an image**
Files: `core/apperr/apperr.go`, `core/ingest/stage.go` *(new)*, `core/ingest/stage_test.go` *(new)*.
Do: D-1, D-3, D-5, D-8, D-9, D-12. In `stage_test.go` define package-level helpers `stageUser(t, c) []byte` (a user row, as `TestFile`'s `mustUser` does) and `stageWrite(t, dir, name string, data []byte) string` (writes a file and returns its absolute path); do not edit `file_test.go`. Build fixtures in the test with real encoders (`image/png`, `image/jpeg`, `image/gif`, `golang.org/x/image/tiff`, `golang.org/x/image/bmp`) and a hand-written PNG header for the 100000 × 100000 case, as `raster_test.go` does; no binary fixtures are checked in. Add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/...` passes.

**STEP-6 — Insert inside a write, or discard**
Files: `core/ingest/stage.go`, `core/ingest/stage_test.go` (both *(new)* in STEP-5).
Do: D-10 and D-11. Add T-5, T-7, and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/ingest/... ./core/database/files/...` passes.

**STEP-7 — Releases schedule their own purge**
Files: `core/database/deleteimpact/facets.go`, `release.go`, `release_test.go`, `impact_test.go`, `facets_test.go`, `core/database/deleteimpact/purge_test.go` *(new)*; `core/database/artifacts/artifacts.go`, `core/database/observations/write.go`, `core/database/subjects/subjects.go`, `core/database/sources/sources.go`, `core/database/citations/citations.go`, `core/database/metadatafields/metadatafields.go`, `core/database/promote/unpromote.go`, `core/database/promote/unpromote_test.go`, `core/database/identityclaims/bridgefile.go`. Skill: `add-catalog-delete`.
Do: D-13. In-package tests (`release_test.go`) call the unexported bodies and keep their object assertions. `release_test.go` near line 348 calls `releaseOwned`. In external tests exactly these transactions change from `db.Begin()` to the fixture catalog's `BeginWrite()`, with `c.EndWrite()` called after that transaction's commit or rollback: `impact_test.go` the subtest transaction near line 391 (`ReleaseFacets`), `TestSnapshotOwnedThenRelease`'s transactions near lines 618 and 643 (`ReleaseSnapshot`); `facets_test.go` `TestReleaseFacetsReportsHandles` near line 164. Every other `db.Begin()` stays. Inside a changed transaction, every non-release call gets `tx.Tx` and every query goes through `tx`, never `db` (the one connection is held). Their raw `tx.Commit()` does not run the scheduled purge; the tests assert rows, as today. `TestReleaseFacetsReportsHandles` reads `got` as the change list. `unpromote.go` and `unpromote_test.go` call `identityclaims.UnfileBridges(tx, …)` with `tx` instead of `tx.Tx`. Add T-9. Then `grep -rn "files.PurgeAfterCommit(" core api --include='*.go' | grep -v _test.go` lists only `core/database/deleteimpact` and `core/database/files/purge.go`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/... ./api/...` passes; the grep holds.

**STEP-8 — Docs**
Files: `core/database/files/purge.go`, `docs/source-layer-data-model.md`, `docs/catalog-deletes.md`, `core/database/deleteimpact/README.md`, `.agents/skills/add-catalog-delete/SKILL.md`, `.cursor/rules/catalog-deletes.mdc`, `.agents/skills/review-pr/lenses/security.md`, `.agents/skills/ship-pr/SKILL.md`, `docs/ideas/large-jpeg-thumbnails.md`, `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, this plan.
Do:
- `purge.go`: D-14.
- `source-layer-data-model.md` §6: change "The stored object is the **original file bytes**" to "The stored object is the **original file bytes** of an import"; after that paragraph add "A File can also hold an image the app rendered (a depiction, S10-32). Go checks its path as for an import, decodes only PNG or JPEG within the pixel budget (`core/raster`), and stores its own PNG, never the app's bytes (`ingest.StageImage`). A staged image's object is written inside the write that commits its `files` row. An object is removed only by `files.PurgeUnreferenced`, when no row names its checksum; every object write and purge runs inside `catalogsession.Do`, which runs a project's operations one at a time, so a delete and an import of the same bytes cannot interleave."
- `catalog-deletes.md`: item 1's "Every releasing `Delete` then calls `files.PurgeAfterCommit`, which unlinks" becomes "Every exported release takes the write's `*database.Tx` and schedules `files.PurgeAfterCommit` itself, which unlinks"; item 2's "then `ReleaseSnapshot`, which returns the released objects for `files.PurgeAfterCommit`" becomes "then `ReleaseSnapshot`, which schedules the purge of what it released"; the "One call per delete" line's "prepends `Released.Changes`" becomes "prepends the changes it returns"; recipe step 6 drops "→ `files.PurgeAfterCommit` (everything released, facets included)".
- `core/database/deleteimpact/README.md`: "`ReleaseSnapshot` and `ReleaseReplaced` return the released objects, and every releasing delete passes them to `files.PurgeAfterCommit`" becomes "Every exported release takes the write's `*database.Tx` and schedules `files.PurgeAfterCommit` for the Files it released; callers never see the objects."
- `.agents/skills/add-catalog-delete/SKILL.md`: the checklist line drops "→ `files.PurgeAfterCommit`"; in the sketch (lines 70–82), `released, err := deleteimpact.ReleaseFacets(` becomes `changes, err := deleteimpact.ReleaseFacets(`, the `objs, err := deleteimpact.ReleaseSnapshot(…); files.PurgeAfterCommit(…)` comment becomes `// err = deleteimpact.ReleaseSnapshot(tx, snap) — it schedules the purge`, and both `released.Changes` become `changes`; the owned-FK paragraph's "its `Delete` calls `files.PurgeAfterCommit` with what it releases" becomes "its `Delete` calls the release, which schedules the purge".
- `.cursor/rules/catalog-deletes.mdc`: "`ReleaseSnapshot`, then `files.PurgeAfterCommit` with what it released" becomes "`ReleaseSnapshot` (it schedules the purge)".
- `security.md` trust table: the Ingest row's Where gains "`ingest.StageImage` decodes and re-encodes app-rendered images"; the Derivatives row becomes "`core/derivatives` (thumbnails), `core/raster` (the one bounded codec, also used by staging)".
- `ship-pr` skill: the Security and Performance rows of the routing table add `core/raster`.
- `security.md`: "as `raster.Thumbnail` does" (around line 90) becomes "as `raster.EncodeJPEG` and `raster.EncodePNG` do".
- `docs/ideas/large-jpeg-thumbnails.md`: repoint its link to `core/raster/raster.go` (new) by the STEP-2 move, and at the end of the "Suggested fix" paragraph that starts "Keep the 64-megapixel full decode" add "An ImageIO fallback goes in `EncodeJPEG`, not in the shared `decode`, so staged images never take it."
- `deployment-plan.md`: the S10-53 **Plan:** line's status becomes "— Landed #N", and " — done, #N" is appended to its SLICE 8 line (as `S10-50  … — done, #413`). `completed.md`: a `## S10-53 — Stage a client image` section above the current first section, in the shape of the `S10-55` section.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-53-stage-image.md` passes; `grep -rn "derivatives/raster" core .agents docs/ideas` returns nothing; `grep -rn "Released.Objects\|released.Objects\|released.Changes\|Released.Changes" docs .agents .cursor core/database/deleteimpact/README.md | grep -v docs/deployment-plan/` returns nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/raster/raster_test.go` (new) by the STEP-2 move `TestDecode` *(new, in-package)*, table-driven | PNG with formats `["png","jpeg"]` → ok; real TIFF with the same list → `ErrUnsupportedFormat`; the test file's registered `fake` format with the same list → `ErrUnsupportedFormat`, not the fake decoder's "fake: no decode" error (so no full decode ran); real TIFF with `nil` → ok; PNG header declaring 100000 × 100000 → `ErrTooLarge` | AC-1.2, AC-1.3, D-4 |
| T-2 | Go | `core/raster/raster_test.go` (new) by the STEP-2 move `TestEncodePNG` *(new)*, table-driven | 4000 × 1000 JPEG, `MaxEdge` 2048 → PNG 2048 × 512; 300 × 200 PNG → PNG 300 × 200; `MaxEdge` 0 → error; GIF with `Formats: ["png"]` → `ErrUnsupportedFormat`; `EncodeJPEG` with `Formats: ["png"]` on a JPEG → `ErrUnsupportedFormat` | AC-1.5, D-4 |
| T-3 | Go | `core/objectstore/objectstore_test.go` `TestWriteIfAbsent` *(new)* | absent → written, `wrote` true; a checksum that does not match `data` → error, nothing written; same bytes again → `wrote` false; a different file at the path → replaced, `wrote` true | D-6 |
| T-4a | Go | `core/database/files/files_test.go` `TestLookupByChecksumTx` *(new)* | inserted row found inside the same tx; unknown checksum → `sql.ErrNoRows`; bad hex → `ErrInvalid` | D-7 |
| T-4 | Go | `core/ingest/stage_test.go` `TestStageImageRefuses` *(new)*, table-driven | every AC-1.1 input (`maxImageBytes` set to 1024 for the size row, restored after) and every AC-1.2 and AC-1.3 input → the named error by `errors.Is`; the objects directory has no new file. A valid PNG → staged, and the objects directory still has no new file (AC-1.6). Plus: for each `f` in `stageFormats`, `mediatypes.Allowed("image/"+f)` is true | AC-1.1, AC-1.2, AC-1.3, AC-1.6, D-3, D-8, D-9, D-12 |
| T-5 | Go | `stage_test.go` `TestStageImageReencodes` *(new)* | stage, insert in a committing write (T-7's shape), read the object: 300 × 200 PNG plus 1 KiB of ZIP bytes → object decodes to 300 × 200, differs from the input, ends at `IEND`; 4000 × 1000 JPEG → 2048 × 512 PNG | AC-1.4, AC-1.5, D-9 |
| T-6 | Go | existing `core/ingest/file_test.go`, `file_unix_test.go`, moved `core/raster/raster_test.go` (new) by the STEP-2 move, `core/derivatives/thumbnail_test.go`, every delete test named in AC-3.3 | pass with only import paths (raster) or release signatures (STEP-7) changed | AC-3.1, AC-3.2, AC-3.3 |
| T-7 | Go | `stage_test.go` `TestStagedImageInsertTx` *(new)* | stage, `writes.Run(c, writes.Op{Action: "create_file", UserID: stageUser(t, c)}, …)` whose function returns `s.InsertTx(tx.Tx)`'s File and changes → one `files` row (`image/png`, empty `original_filename`), object present, one `file` create change; the same stage inserted twice in one write → one row, second call no change | AC-2.1, AC-2.3, D-10 |
| T-8 | Go | `stage_test.go` `TestStagedImageDiscard` *(new)*, table-driven | stage, `InsertTx`, write function returns an error, `Discard` → no row, no object; a row and object for the PNG exist, stage the same PNG, `InsertTx` in a failing write, `Discard` → row and object stay; stage, insert, commit, `Discard` twice → object stays | AC-2.2, AC-2.4, AC-2.5, D-11 |
| T-9 | Go | `core/database/deleteimpact/purge_test.go` (new) `TestReleaseSchedulesPurge` *(new, `package deleteimpact_test`)*, table-driven | fixture: a catalog from `database.Create(t.TempDir(), "t.provenencia")` (as `core/database/files/purge_test.go` does), a `files` row inserted with `files.Insert`, its object written with `objectstore.WriteIfAbsent` at `files.StorageRelPath`, and an Artifact pointing at it created with `artifacts.Create` inside `writes.Call` (a Source first, as `core/database/artifacts/artifacts_test.go` does). `ReleaseReplaced` after `UPDATE artifacts SET file_id = NULL` in a committing `writes.Call` (returning one change as S10-55's committing subtest in `core/database/files/purge_test.go` does) → `files` row and object gone; the same in a write that returns an error → both stay; `SnapshotOwned`, `DELETE FROM artifacts`, `ReleaseSnapshot` in a committing write → object gone | AC-4.1, AC-4.2, AC-4.3, D-13 |
| T-10 | Go | existing `core/catalogsession/session_test.go` `TestDo` case "serializes ops on one session" | passes unedited | AC-5.1, D-14 |

AC-4.4 is checked by the compiler: STEP-7's `go build` fails while any caller reads `Released` or a returned object list.

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestFile` (STEP-1), the moved raster tests and `TestEnsureThumbnailRejectsUnsafeSources` (STEP-2), thumbnail writes (STEP-3), the delete and Unpromote tests (STEP-7).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] The STEP-7 grep lists files.PurgeAfterCommit only in core/database/deleteimpact and core/database/files/purge.go
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none in this PR)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI in this PR)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: source-layer-data-model.md §6, catalog-deletes.md, deleteimpact README, add-catalog-delete, catalog-deletes rule, security lens trust table, ship-pr routing table, large-jpeg-thumbnails idea
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app — nothing calls `StageImage` until S10-32 and S10-54, and deletes look the same to the researcher. Instead, run `CGO_ENABLED=1 go test -tags fts5 -race -run 'TestDo|TestStageImage|TestStagedImage|TestDecode|TestEncodePNG|TestWriteIfAbsent|TestLookupByChecksumTx|TestReleaseSchedulesPurge' -v ./core/ingest/... ./core/raster/... ./core/objectstore/... ./core/database/files/... ./core/database/deleteimpact/... ./core/catalogsession/...` and paste the result in the PR. Read `files.PurgeUnreferenced`'s doc comment and data model §6 and confirm both name `catalogsession.Do` (AC-5.2). Then, in the app, delete an Artifact with an attached image and confirm its thumbnail disappears from the project's objects folder (the existing S10-55 behavior, now through the release).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Moving the decode slot changes thumbnail concurrency | `raster.Slot` keeps `decodeSlots`' size; T-6 keeps the thumbnail tests. |
| A delete and an import of the same bytes rest on `catalogsession.Do` | Every production path runs inside it today (§4 Serialization); D-14's doc and data model §6 name the rule; `TestDo` pins the serialization. |
| A purge called inside a write function waits forever for the connection that write holds | D-11 and D-14 docs; `Discard` is deferred before the write, so it runs after `Run` returns; `PurgeAfterCommit` runs after commit. |
| The staged object write inside the transaction lengthens the write | One ≤ 17 MiB write; no decode moves in. |
| `golang.org/x/image` decoders on an untrusted path | D-4 and D-8 refuse every format but PNG and JPEG before a full decode. |

**Stop and report if:**

- An existing test in `core/ingest` needs an edit to pass after STEP-1, or a raster or thumbnail test needs more than an import-path edit after STEP-2.
- `image.DecodeConfig` allocates the declared pixels for any format (check with the 100000 × 100000 header in T-1).
- `PurgeUnreferenced` or `Discard` hangs in any test (something calls it inside a write).
- A production caller of `ingest.File`, `derivatives.Ensure`, or a releasing delete runs outside `catalogsession.Do` (the guarantee D-14 documents would not hold).
- `deleteimpact` importing `core/database` makes an import cycle.
- A delete test in AC-3.3 needs more than the release signature change to pass.

## 12. Open questions

None.

## 13. PR

**Title:** S10-53 — Stage a client image
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-53-stage-image.md`.
