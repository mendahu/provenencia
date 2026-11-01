# S14-12 — File work before the catalog session

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 4 — Evidence graphs finished (batch 4; moved from Spike 12, where it was S12-27, on 2026-10-29)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: every view draws the same rows and thumbnails; only where Go copies, hashes and decodes moves)
**Board:** —
**Depends on:** — (S14-23's `raster.MaxDecodeBytes` and DC decode are on `main` @ `a652313`)
**Blocks:** —
**Branch:** `s14-12-file-work-before-session`
**Version bump:** patch — a performance fix with no feature and no wire change ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge (0.2.4 → 0.2.5 at planning)
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-depiction-value`](../../../../.agents/skills/add-depiction-value/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`app-health-audit.md`](../app-health-audit.md) A-24; working rule "File work and decoding happen before the catalog session" (map review K-4); bar outcome 14 ("attaching a large scan does not stall the other panes"); [`source-layer-data-model.md`](../../../source-layer-data-model.md) (line 313: a purge and an import never interleave; line 467: thumbnails); [`oversized-jpeg-thumbnails.md`](../oversized-jpeg-thumbnails.md); S14-23 ship log PRF-1 ([`S14-23-oversized-jpeg-thumbnails.ship.md`](S14-23-oversized-jpeg-thumbnails.ship.md))
**Planned:** 2026-11-01 against `main` @ `a652313`
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

Attaching a large scan, saving a depiction, or opening a Source page whose images have no thumbnail yet no longer holds the catalog session while Go copies, hashes, decodes and encodes. Each of those paths stages its bytes first, outside `catalogsession.Do`, and enters the session only to look up, insert and install. Other panes (Source page, graph loads, counts, search) run while the file work runs.

**Why now.** Audit A-24 and bar outcome 14. Since S14-23, a JPEG over 64 MP that fails late is decoded in full on every Artifact response (565 ms for a cut-short 11,000 × 18,000 scan, up to 3 s for a crafted one, S14-23 ship log PRF-1), and today all of it runs under the session lock.

**What stays.** Every response, error code, audit revision and `WriteEffects` value is as today. Objects are installed and purged only inside `catalogsession.Do`, so a purge and an import of the same bytes still never interleave ([`source-layer-data-model.md`](../../../source-layer-data-model.md) line 313). Thumbnails are still made lazily by the responses that make them today; nothing stores a refused decode.

## 2. User stories

**US-1.** As a researcher, when I attach a large scan to an Artifact, the other panes keep loading while it copies, so that the app does not freeze for the length of the copy.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `IngestArtifactFile` while another call holds the catalog session | The copy, hash and sniff finish while the session is held: the `.ingest-*` temp in the project's objects folder reaches the source file's full size before the session is released |
| AC-1.2 | The session is released | The ingest completes with the same response as an unblocked run, and a `GetSourceWorkspace` queued behind the held call is not held for the copy |
| AC-1.3 | The same file attached twice, a file over `ingest.MaxBytes`, an executable, a symlink | Same results and error codes as today (`TestFile` and the artifact handler tests unchanged in their assertions) |

**US-2.** As a researcher, when I save a depiction, its image is decoded and re-encoded before the save enters the catalog session.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CreateCitationWithObservations` with one depiction while another call holds the session | The staged PNG temp (`.stage-*` in the objects folder) exists before the session is free; the save completes after it with the same response as today |
| AC-2.2 | A depiction whose write fails (Observation refused) | No object is left for the staged PNG and no temp is left in the objects folder |
| AC-2.3 | `UpdateObservation`, `AddObservationsToCitation`, `CreateCitedBridge` with a depiction | Each stages before its session through the same entry point (`withStagedDepictions`); their handler tests pass unchanged in their assertions |

**US-3.** As a researcher, when a Source page lists images with no thumbnail yet, the decodes run outside the catalog session.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `derivatives.Plan` returns a `Job` inside a session; another call then holds the session | `Job.Stage` completes while the session is held |
| AC-3.2 | `GetSourceWorkspace` on a Source with two raster Artifacts and no derivatives | Both `thumbnail_rel_path` fields are filled, as today; two `file_derivatives` rows exist after the call |
| AC-3.3 | Two Artifacts with the same File on one Source page | One decode, one derivative row; both protos carry the same path |
| AC-3.4 | A JPEG that fails late (no EOI, over 64 MP) on a Source page | The page returns with an empty `thumbnail_rel_path` for it; the decode ran outside the session; no derivative row |
| AC-3.5 | `SetSourceCover` to an Artifact whose thumbnail does not exist yet | The thumbnail is made outside the session and the cover is set, as today; a cover on a non-raster or unprocessable Artifact returns `sources.ErrInvalid`, as today |
| AC-3.6 | `EnsureFileThumbnail` for a raster File with no derivative | Same response as today; the decode ran outside the session |

## 3. Scope

| In | Covered by |
| --- | --- |
| `depictionvalue.ImageMediaType` *(new)*, used by `ingest` and `ReadableImage` | STEP-1 |
| One staged-object type, `ingest.Staged` *(replaces `StagedImage`)*: stage outside the session, insert/install inside, discard what the write did not use | STEP-2, AC-2.2 |
| `ingest.StageFile` *(new)* and `Staged.CommitFile` *(new)*; `ingest.File` becomes their composition for callers already in a session | STEP-2, AC-1.3 |
| `StageImage` and `StageDepictions` take a project folder, not a catalog | STEP-2 |
| `derivatives.Plan`, `Job.Stage`, `derivatives.Commit` *(new)*; `Ensure` becomes their composition | STEP-3, AC-3.1 |
| `withStaged` / `withStagedDepictions` *(new)*: the one stage-then-write entry point for the citation, observation (×2) and connect handlers | STEP-4, AC-2.1, AC-2.3 |
| `IngestArtifactFile` stages before `withProjectCatalog` | STEP-5, AC-1.1, AC-1.2 |
| `thumbnailPass` *(new)*: `artifactProto` callers, `EnsureFileThumbnail` and `SetSourceCover` plan in the session, decode outside it, commit in a second session | STEP-6, AC-3.2 – AC-3.6 |
| `thumbnails_test.go` uses the harness's `damageCatalog` in place of its raw `DELETE FROM file_derivatives` (Spike 12 ship-log leftover) | STEP-7 |
| `use-catalog-session`, `add-ffi-handler`, `add-depiction-value`, the maintainability and security lenses, and `source-layer-data-model.md` name the new shape | STEP-8 |

| Out | Home |
| --- | --- |
| Oversized-JPEG thumbnails | S14-23 (landed) |
| Storing crops below 2,048 px | [depictions and likenesses idea](../../../ideas/depictions-and-likenesses.md) |
| Remembering a refused decode so a late-failing JPEG is not decoded again on the next response (PRF-1's repeat cost) | not planned — this PR takes the decode off the session; reopen from the dogfood log if the Source page itself is slow |
| Decoding several thumbnails in parallel | not planned — `raster.Slot` bounds memory; one handler decodes in sequence as today |
| A Swift change, or an `EnsureFileThumbnail` call from the app | not planned — the app reads thumbnail paths from responses, as today |
| Making `sources.NormalizeCover` lookup-only | not planned — the handler's pass makes the thumbnail first (D-9), and core tests rely on `NormalizeCover` generating |

## 4. Current state

Verified at `a652313`.

| Area | Where | Today |
| --- | --- | --- |
| Session lock | `core/catalogsession/session.go` `Do`, `runLocked` | One mutex per project; `fn` runs with it held. `api/ffi/handlers/catalog.go` `withProjectCatalog` wraps `Do` |
| Import | `core/ingest/file.go` `File(c, absPath, userID)` | `openChecked` → `streamSource` (stream up to 512 MiB to an `objectstore.Temp` in `objects/.ingest-*`, SHA-256, 512-byte sniff) → `mediatypes.Resolve` → `files.StorageRelPath` → `LookupByChecksum` (reuse: `tmp.Install`, then `reuseFile` Run) or `tmp.Install` → `writes.Run` `create_file` with `insertFile`; a lost checksum race falls back to reuse; a failed write purges what it wrote. All of it runs inside the caller's session |
| Import call site | `api/ffi/handlers/artifacts.go` `IngestArtifactFile` | Inside one `withProjectCatalog`: `artifacts.Get`, `ingest.File`, `update_artifact` Run, `artifactProto` (which ensures a thumbnail, below) |
| Depiction stage | `core/ingest/stage.go` `StageImage(c, absPath)`, `StagedImage` | Takes a catalog only for `c.DB()` and `c.Dir()`; reads ≤32 MiB under `raster.Slot`, `raster.EncodePNG` (PNG/JPEG, ≤2048 px) into memory; `InsertTx(tx *sql.Tx)` looks up by checksum in the tx, inserts with `insertFile`, then `objectstore.WriteIfAbsent`; `Discard()` purges through the catalog it kept. `"image/png"` literal three times (lines 76, 104, 128) |
| Depiction batch | `core/ingest/depictions.go` `StageDepictions(c, inputs) (discard func(), err)` | Stages each `observations.Input.Depiction` with no `Image`; discards all on error |
| Depiction call sites | `api/ffi/handlers/citations.go` `CreateCitationWithObservations`; `observations.go` `UpdateObservation`, `AddObservationsToCitation`; `connect.go` `CreateCitedBridge` | Each calls `ingest.StageDepictions(c, …)` as the first line inside `withProjectCatalog`, then `defer discard()` |
| Depiction interface | `core/database/observations/depiction.go` `StagedFile` | `InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)`; comment names `*ingest.StagedImage` |
| Test helper | `core/database/observations/observationstest/observationstest.go` `SaveCitation`, `UpdateObservation` | Call `ingest.StageDepictions(c, …)` inside the test's catalog |
| Stored image type | `core/depictionvalue/depictionvalue.go` `ReadableImage` | Compares `mediaType != "image/png"` (line 49); doc comment names `ingest.StageImage` |
| Thumbnail | `core/derivatives/ensure.go` `Ensure`, `EnsureThumbnail`, `generate`, `readSourceObject`, `insertLinkOnly`, `errDerivedFileExists` | `files.Lookup` → `raster.Supported` → existing link → `MaxDecodeBytes` → `ReadRelPath` → `generate` (Slot, bounded checksum-verified read, `raster.EncodeJPEG`) → `WriteIfAbsent` → `LookupByChecksum` → unaudited `writes.Run(writes.Op{})` inserting the `files` row and the link (`insertDerivativeLink`, search reprojection after commit). All inside the caller's session |
| Thumbnail call sites | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`; callers `artifactProto` (`artifacts.go`: `CreateArtifact`, `UpdateArtifact`, `IngestArtifactFile`, `listArtifactsProto` ← `sources.go` `GetSourceWorkspace`) and `EnsureFileThumbnail` | Each ensures inside `withProjectCatalog`; `ErrUnprocessable`, `ErrCorruptObject`, `ErrInvalid` → empty path (skipped); other errors fail the call |
| Cover | `core/database/sources/sources.go` `NormalizeCover` → `requireRasterCoverArtifact` → `derivatives.EnsureThumbnail`; handler `api/ffi/handlers/sources.go` `SetSourceCover` | Ensures the cover's thumbnail inside the session before the `set_source_cover` Run |
| Lookup-only thumbnails | `thumbnails.go` `existingThumbnailRelPath`, `sourceCoverThumbnail(s)` | Never decode; unchanged by this PR |
| Swift | `macos/App` | Never calls `EnsureFileThumbnail`; reads `thumbnailRelPath` from Artifact protos (`GoStore.swift` line 2071) |
| Tests | `core/ingest/file_test.go` `TestFile`; `stage_test.go` `TestStageImageRefuses`, `TestStageImageReencodes`, `TestStagedImageInsertTx`, `TestStagedImageDiscard`, helper `stageInsert`; `unsafe_objects_test.go` `TestStagedImageInsertTxRefusesSymlinkedObjects`; `core/derivatives/thumbnail_test.go` (`TestEnsureThumbnail*`); `api/ffi/handlers/thumbnails_test.go` `TestEnsureFileThumbnail`, `TestListSourcesAndWorkspaceThumbnails` (raw `DELETE FROM file_derivatives` at line 217); `api/ffi/handlers/catalog_session_test.go` `TestConcurrentCatalogRPCs`; harness `api/ffi/handlers/harness_test.go` `damageCatalog` (line 156) | Pass at the base |

**Precedent.** `3a24ed95` ("Stage an app-rendered image as a checked PNG that a write stores with its File row") split the depiction image into a stage before `writes.Run` and an `InsertTx` inside it. This PR moves that split's boundary from "before the write" to "before the session" and gives the import and the thumbnail the same type. Copy its `InsertTx` / `Discard` contract and its tests' shape; do not copy keeping the encoded bytes in memory (D-3).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One type, `ingest.Staged`, replaces `StagedImage`. Fields: project folder, an `*objectstore.Temp`, checksum, media type, byte size, storage path, sanitized filename (imports only), and the `wrote` / `file` / `haveFile` state `StagedImage` keeps today. Constructors: `StageFile` (import), `StageImage` (depiction), `StageBytes` (bytes Go made: the thumbnail) | Working rule: one staged-object type in `ingest`, used by all three | A second type per path — the rule forbids it; an interface with three implementations — three paths share every field |
| D-2 | Staging takes the project folder (`projectDir string`), never a `*database.Catalog`, and first calls `database.RequireProject(dir)` *(new: exports the `locateCatalog` check, returns `database.ErrNotAProject`)* so a stage never creates an objects folder in a folder that is not a project. The folder is `filepath.Clean`ed, the same key `catalogsession.Do` uses | A catalog exists only inside `Do`; `objectstore.CreateTemp` makes the objects folder when absent | Opening a session just to read `c.Dir()` — it is the lock this PR avoids |
| D-3 | Every stage writes its bytes to an `objectstore.Temp` in the project's objects folder outside the session (`.ingest-` prefix for imports, `.stage-` for images and thumbnails). Inside the session only `Temp.Install` (a rename) runs | Install is then cheap and the same for all three; the temp is on the objects volume, so the rename never copies | Keeping PNG/JPEG bytes in memory and `WriteIfAbsent` inside the session — a 2048 px PNG write is file work inside the lock; a temp in `os.TempDir` — a cross-volume move copies inside the lock |
| D-4 | `Staged.InsertTx(tx *sql.Tx)` keeps `StagedImage.InsertTx`'s contract: look up by checksum in the tx, insert with `insertFile` (setting `original_filename` only when the stage has one), install the temp, record `wrote`; a second call returns the same File and no change. It still satisfies `observations.StagedFile` | The depiction write and the thumbnail commit both run inside a tx; the in-tx lookup removes the checksum race the derivative code handles today | A new interface — `StagedFile` already fits |
| D-5 | `Staged.CommitFile(c, userID) (Result, error)` is today's `File` body from `files.StorageRelPath` on: `LookupByChecksum` (reuse: install to repair, `reuse_file` Run), else install, `create_file` Run, lost race → reuse, failure → purge. It runs inside the session. `ingest.File(c, absPath, userID)` keeps its signature as `StageFile(c.Dir(), …)` + `CommitFile` + discard, for callers already in a session (fixtures and tests); no handler calls it | Keeps the import's audit actions (`create_file`, `reuse_file`) and error paths byte for byte; keeps 30+ test call sites unchanged | Moving the import into one `InsertTx` Run — it would change the audit action of a reuse |
| D-6 | Two discard methods. `Staged.Discard(c *database.Catalog)` runs inside the session: it purges the object when this stage wrote it and no row names the checksum, then removes the temp. `Staged.Close()` runs outside: it removes the temp only, never touches an installed object. Both are idempotent. `ingest.Stages` *(new, `[]*Staged`)* has the same two methods | Purge must stay inside `Do` (line 313); a stage that never reached the session (open failed) must still lose its temp | One `Discard()` called from both places — it would purge outside the lock if the inner call were forgotten |
| D-7 | Handlers stage through one entry point in `api/ffi/handlers/catalog.go`: `withStaged(projectDir string, stages ingest.Stages, fn func(*database.Catalog) error) error` *(new)* defers `stages.Close()`, runs `withProjectCatalog`, and defers `stages.Discard(c)` first inside it. `withStagedDepictions(projectDir, inputs, fn)` *(new)* is `ingest.StageDepictions` + `withStaged`. The four depiction handlers and `IngestArtifactFile` use them; no handler calls `Discard` or `Close` itself | Map: "one stage-then-write entry point"; the two-defer order cannot be got wrong per handler | A per-handler pair of defers — five copies of the same ordering |
| D-8 | Thumbnails split in `derivatives`: `Plan(c, sourceFileID, spec) (Result, *Job, error)` *(new)* is `Ensure` up to `generate` (returns a final `Result` and a nil `Job`, or a `Job`); `(*Job).Stage(projectDir) (*ingest.Staged, error)` *(new)* holds `raster.Slot`, reads with `readSourceObject` (capped at `raster.MaxDecodeBytes`), `raster.EncodeJPEG`, then `ingest.StageBytes(projectDir, jpeg, "image/jpeg")`; `Commit(c, job, staged) (Result, error)` *(new)* re-checks the source File and an existing link, then one unaudited `writes.Run(writes.Op{})` doing `staged.InsertTx` and `insertDerivativeLink`. `Ensure` becomes `Plan` → `Stage(c.Dir())` → `Commit` with discard, so its error contract and tests are unchanged. `errDerivedFileExists` and `insertLinkOnly` go (D-4's in-tx lookup removes their race) | Map: "the thumbnail splits into read, decode outside, and write"; keeps `Ensure` for callers inside a session | Moving `Plan` / `Commit` into the handlers — the derivative rules belong to `derivatives` |
| D-9 | Handlers make thumbnails through `thumbnailPass` (new, in `api/ffi/handlers/thumbnail_pass.go` (new)): inside the first session `need(c, fileID, dst *string)` calls `Plan`, fills `dst` from an existing link at once, or records the `Job` (one per File id, any number of `dst`); after it, `run(projectDir)` stages every job outside the session, then, only when a job staged, enters one second session to `Commit` each, fills each `dst`, and discards. `artifactProto(c, a, pass)` takes the pass; `CreateArtifact`, `UpdateArtifact`, `IngestArtifactFile`, `GetSourceWorkspace` and `EnsureFileThumbnail` call `pass.run` after their session and before `proto.Marshal`. `SetSourceCover` in artifact mode runs a first session that plans the cover Artifact's File, runs the pass, and then runs today's session; when the pass skipped that File it returns `sources.ErrInvalid` without calling `NormalizeCover` | Every production decode leaves the session; `NormalizeCover` then finds the link and does not decode | Making the Source page lookup-only and the app ask for thumbnails — a Swift change (Out); one session per job — N extra lock round trips |
| D-10 | The pass maps errors as `thumbnailRelPath` does today: `ErrUnprocessable`, `ErrCorruptObject`, `ErrInvalid` from `Plan`, `Stage` or `Commit` → empty path; any other error fails the call. A source object missing at `Stage` fails the call, as a missing object does today; a source File gone at `Commit` is `ErrInvalid` → empty path | No response changes; the Source page after the delete re-plans | Treating a missing object as skipped — a behavior change this PR does not need |
| D-11 | `depictionvalue.ImageMediaType = "image/png"` *(new const)*; `ingest.StageImage` and `ReadableImage` use it. Thumbnails keep their own `"image/jpeg"` | Map scope; one owner for the stored depiction type | A const in `ingest` — `depictionvalue` is the pure type both read |
| D-12 | Session-freedom tests hold the session with `catalogsession.Do` blocking on a channel and show the stage finishes (or its temp is complete) before release; they need no hook in production code | The map's "slow reader in the stage step" is shown without a test seam | A package-var seam around `ingest.StageFile` in handlers — production indirection for a test |

**Rules this plan bends.** None. The `add-ffi-handler` skill's "stages it inside `withProjectCatalog`" sentence changes because the working rule replaced it (STEP-8).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ingest.Staged` *(renamed from `StagedImage`, generalized)* | type | `core/ingest/stage.go` `StagedImage` | It is that type, extended; the name drops "Image" because it now carries imports |
| `ingest.StageFile` | func | `ingest.File` (its first half), `streamSource` | Reuses `openChecked`, `streamSource`, `mediatypes.Resolve` as they are |
| `ingest.StageBytes` | func | `objectstore.WriteIfAbsent` | That writes in place; a stage must defer install to the session |
| `Staged.CommitFile` | method | `ingest.File` (its second half) | Moved, not rewritten |
| `Staged.Close`, `ingest.Stages` | method, type | `StageDepictions`'s `discardAll` closure | The closure becomes `Stages.Discard`; `Close` is the outside half D-6 needs |
| `database.RequireProject` | func | `database.locateCatalog` (unexported) | Exports its check only |
| `derivatives.Plan`, `Job`, `Job.Stage`, `derivatives.Commit` | funcs, type | `derivatives.Ensure`, `generate` | `Ensure`'s body cut at its two boundaries |
| `depictionvalue.ImageMediaType` | const | literals in `stage.go`, `depictionvalue.go` | Replaces them |
| `withStaged`, `withStagedDepictions` | helpers | `withProjectCatalog` | Wraps it; no second session doorway |
| `thumbnailPass` | type | `thumbnailRelPath` | `thumbnailRelPath` becomes `pass.need`'s mapping; searched `rg "type .*Pass|thumbnail" api/ffi/handlers` — no batch thumbnail helper exists |

**Builds on.** `objectstore.Temp` (`CreateTemp`, `Install`, `Discard`); `files.LookupByChecksumTx`, `insertFile`, `files.PurgeUnreferenced`; `raster.Slot`; `withProjectCatalog`; the `observations.StagedFile` interface.
**Deletes.** `StagedImage` (name), its in-memory `png` field and `WriteIfAbsent` call; `StageDepictions`'s returned closure; `derivatives.insertLinkOnly`, `errDerivedFileExists`, `Ensure`'s `WriteIfAbsent` and pre-Run `LookupByChecksum`; `thumbnailRelPath` (folded into `thumbnailPass.need`); the three `"image/png"` literals in `stage.go` and the one in `ReadableImage`; the raw `DELETE FROM file_derivatives` block in `thumbnails_test.go`.
**Next adds.** A later derivative (a medium preview) is a `Spec` through the same `Plan` / `Job.Stage` / `Commit`. A later client file (an attachment on a note) stages with `StageFile` and writes through `withStaged`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A staged temp exists while the session is held by another call | the objects folder (`objectstore.Temp`) | Purge only removes named objects, never `.ingest-*` / `.stage-*` temps (`files.PurgeUnreferenced` builds paths from checksums) — T-6 |
| A row or object changes between the pass's two sessions | `derivatives.Commit` | Re-checks the source File (gone → `ErrInvalid` → empty path) and the link (made by another call → returned, staged temp discarded) — T-5 |
| A stage runs against a folder that is not a project | `database.RequireProject` | `ErrNotAProject`, the error `Do` returns today; no objects folder created — T-1 |

### Scale

At the map's Scale (a 60-graph census; scans up to `ingest.MaxBytes` 512 MiB; a Source page of up to a few hundred Artifacts): the import holds the session for one rename and two short Runs instead of a 512 MiB copy and hash. A Source page with N Artifacts lacking thumbnails holds the session for one read pass, then N decodes outside it (one `raster.Slot` at a time, each ≤ `raster.MaxDecodeBytes`), then one session with N small Runs. Staged thumbnails are ≤256 px JPEGs on disk, so peak memory is one decode, as today. No new cache, no new key, no `WriteEffects` change: derivative rows stay unaudited.

### Edges

- **Assumptions.** `files.PurgeUnreferenced` assumes no import is mid-install: installs and purges both stay inside `Do` (D-3, D-6), held by T-6. `insertDerivativeLink` assumes the derived File row exists in the same tx: `Commit` runs `InsertTx` first in that Run (T-5). `observations.StagedFile` callers assume a second `InsertTx` returns the same File (`TestStagedImageInsertTx` → renamed, unchanged assertions).
- **Neighbors.** A File with an existing link: `Plan` returns it, no job, no second session (T-4). A non-raster File: `Skipped`, no job (existing `TestEnsureThumbnailSkipsNonImage`). A late-failing JPEG: `Stage` returns `ErrUnprocessable`, no second session, empty path (T-7).
- **Removed inputs.** Artifact deleted between the pass's sessions: `Commit` finds no source File → empty path; its staged temp is discarded (T-5). A depiction write refused: `Discard(c)` purges the installed PNG (existing `TestStagedImageDiscard`).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** Bytes read from the project folder: imports capped at `maxBytes` before and during the copy (`openChecked`, `LimitReader`); depiction images at `maxImageBytes`; thumbnail sources at `raster.MaxDecodeBytes`, verified against the stored checksum before decode. Jobs per response are bounded by Artifacts on the page, deduplicated by File id.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database` | `func RequireProject(dir string) error` *(new)* — `locateCatalog`'s check without opening |
| `core/depictionvalue` | `const ImageMediaType = "image/png"` *(new)*; `ReadableImage` uses it; its doc comment names `ingest.StageImage` still |
| `core/ingest/stage.go` | `type Staged struct` (D-1); `StageImage(projectDir, absPath string) (*Staged, error)`; `StageBytes(projectDir string, data []byte, mediaType string) (*Staged, error)` *(new)*; `(*Staged) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)`; `(*Staged) Discard(c *database.Catalog)`; `(*Staged) Close()`; `type Stages []*Staged` with `Discard(c)` and `Close()` *(new)* |
| `core/ingest/file.go` | `StageFile(projectDir, absPath string) (*Staged, error)` *(new)*; `(*Staged) CommitFile(c *database.Catalog, userID []byte) (Result, error)` *(new)*; `File(c, absPath, userID)` = stage + commit + discard (D-5) |
| `core/ingest/depictions.go` | `StageDepictions(projectDir string, inputs []observations.Input) (Stages, error)`; on error it closes every stage so far and returns nil |
| `core/database/observations/depiction.go` | Comment: `*ingest.Staged` satisfies `StagedFile` |
| `core/derivatives/ensure.go` | `type Job struct` *(new: source File id, spec, object path, checksum)*; `Plan`, `(*Job) Stage`, `Commit` *(new)* (D-8); `Ensure` composes them |

```go
// sketch — signatures only
func StageFile(projectDir, absPath string) (*Staged, error)
func StageImage(projectDir, absPath string) (*Staged, error)
func StageBytes(projectDir string, data []byte, mediaType string) (*Staged, error)
func (s *Staged) CommitFile(c *database.Catalog, userID []byte) (Result, error)
func (s *Staged) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)
func (s *Staged) Discard(c *database.Catalog)
func (s *Staged) Close()

func Plan(c *database.Catalog, sourceFileID []byte, spec Spec) (Result, *Job, error)
func (j *Job) Stage(projectDir string) (*ingest.Staged, error)
func Commit(c *database.Catalog, j *Job, staged *ingest.Staged) (Result, error)
```

Errors: unchanged codes. `RequireProject` → `catalog.not_a_project`.

### FFI

None — no proto change. Handlers change inside `api/ffi/handlers/`: `catalog.go` (`withStaged`, `withStagedDepictions`), `citations.go`, `observations.go`, `connect.go`, `artifacts.go` (`IngestArtifactFile`, `artifactProto(c, a, pass)`, `listArtifactsProto(c, id, pass)`), `sources.go` (`GetSourceWorkspace`, `SetSourceCover`), `thumbnails.go` (`EnsureFileThumbnail`; `thumbnailRelPath` removed), and the pass's own file (new, below).

`IngestArtifactFile` shape:

```go
staged, err := ingest.StageFile(req.GetProjectDir(), req.GetPath())
if err != nil { return nil, err }
pass := &thumbnailPass{}
err = withStaged(req.GetProjectDir(), ingest.Stages{staged}, func(c *database.Catalog) error {
	// artifacts.Get, ErrFileAlreadyAttached, staged.CommitFile(c, userID), update_artifact Run, artifactProto(c, got, pass)
})
if err == nil { err = pass.run(req.GetProjectDir()) }
```

The `ErrFileAlreadyAttached` check stays inside the session; a file staged for an Artifact that already has one is closed unused.

### macOS

None — Go only.

### Strings

None — no copy, and no action changes what it covers.

## 7. Steps

**STEP-1 — One stored-image media type**
Files: `core/depictionvalue/depictionvalue.go`, `core/ingest/stage.go`. Skill: `add-depiction-value`.
Do: add `ImageMediaType`; use it in `ReadableImage` and the three `stage.go` literals.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/depictionvalue/... ./core/ingest/...` passes; `rg '"image/png"' core/ingest/stage.go core/depictionvalue` has no hit.

**STEP-2 — One staged-object type in `ingest`**
Files: `core/database/catalog.go`, `core/ingest/stage.go`, `core/ingest/file.go`, `core/ingest/depictions.go`, `core/database/observations/depiction.go`, `core/database/observations/observationstest/observationstest.go`, `core/ingest/*_test.go`, every `*_test.go` that calls `StageImage` or `StageDepictions`. Skill: `use-catalog-session`.
Do: add `RequireProject`; build `Staged`, `StageFile`, `StageImage`, `StageBytes`, `InsertTx`, `CommitFile`, `Discard`, `Close`, `Stages` (D-1 – D-6); rewrite `File` as the composition; `StageDepictions` takes the folder and returns `Stages`. `observationstest` stages with `c.Dir()` and defers `Discard(c)` then `Close()`. Tests change only `StagedImage` → `Staged`, `c` → `c.Dir()` in stage calls, and the discard call shape; no assertion changes. Add T-1, T-2, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes; `rg StagedImage --type go` has no hit.

**STEP-3 — Split the thumbnail derivative**
Files: `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: add `Job`, `Plan`, `Job.Stage`, `Commit` (D-8); `Ensure` composes them; delete `insertLinkOnly` and `errDerivedFileExists`. Add T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/derivatives/... ./core/database/sources/... ./core/database/artifacts/...` passes with every existing `TestEnsureThumbnail*` assertion unchanged.

**STEP-4 — Depiction handlers stage before the session**
Files: `api/ffi/handlers/catalog.go`, `citations.go`, `observations.go`, `connect.go`, `api/ffi/handlers/citations_test.go`. Skill: `add-ffi-handler`.
Do: add `withStaged`, `withStagedDepictions` (D-7); the four handlers call `withStagedDepictions` in place of `withProjectCatalog` + `StageDepictions`. Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `rg "StageDepictions" api/ffi/handlers` hits only `catalog.go`.

**STEP-5 — Import stages before the session**
Files: `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/artifacts_test.go`.
Do: `IngestArtifactFile` per the FFI sketch (thumbnail through the pass comes in STEP-6; in this step `artifactProto` keeps its signature). Add T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes; `rg "ingest\.File\(" api/` has no hit.

**STEP-6 — Thumbnails decode between two sessions**
Files: `api/ffi/handlers/thumbnail_pass.go` *(new)*, `artifacts.go`, `sources.go`, `thumbnails.go`, `api/ffi/handlers/thumbnails_test.go`, `api/ffi/handlers/cover_test.go`.
Do: add `thumbnailPass` (D-9, D-10); thread it through `artifactProto` and `listArtifactsProto`; call `pass.run` in `CreateArtifact`, `UpdateArtifact`, `IngestArtifactFile`, `GetSourceWorkspace`, `EnsureFileThumbnail`; split `SetSourceCover`; delete `thumbnailRelPath`. Add T-7, T-9, T-10.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/... ./core/...` passes; `rg "EnsureThumbnail\(|derivatives\.Ensure\(" api/` has no hit.

**STEP-7 — Thumbnail test uses the harness**
Files: `api/ffi/handlers/thumbnails_test.go`.
Do: replace the `Close` / `database.Open` / `DELETE FROM file_derivatives` block (lines 205–225 at the base) with `damageCatalog(t, f.dir, "DELETE FROM file_derivatives")`, as `sources_test.go` line 374 does.
Check: `rg "DELETE FROM file_derivatives" api/ffi/handlers/thumbnails_test.go` hits only the `damageCatalog` call; `go test` passes.

**STEP-8 — Guidance and domain doc**
Files: `.agents/skills/use-catalog-session/SKILL.md`, `.agents/skills/add-ffi-handler/SKILL.md` (§ Files from the client), `.agents/skills/add-depiction-value/SKILL.md`, `.agents/skills/review-pr/lenses/maintainability.md` (line 313), `.agents/skills/review-pr/lenses/security.md` (line 22), `docs/source-layer-data-model.md` (lines 313 and 467).
Do: `use-catalog-session` gains a "File work before the session" section: stage outside `Do`, `withStaged` / `withStagedDepictions` for client files, `thumbnailPass` for thumbnails, install and purge only inside `Do`. `add-ffi-handler` says the handler stages before `withProjectCatalog` through `withStaged`. The lenses name `ingest.Staged`. The domain doc says file work and decoding run before the session and the object install and purge inside it.
Check: `python3 scripts/check-guidance-names.py` passes; `rg "StagedImage|stages it inside" .agents docs` has no hit.

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `VERSION`, `core/version.go`, `MARKETING_VERSION`.
Do: per the Definition of done.
Check: links resolve; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/ingest/stage_test.go` `TestStageOutsideSession` *(new)* | Table: `StageFile`, `StageImage`, `StageBytes` on a project while a goroutine holds `catalogsession.Do` blocked on a channel → each returns before release; on a folder with no catalog → `database.ErrNotAProject` and no `objects/` folder made | AC-1.1, AC-2.1, D-2, D-12 |
| T-2 | Go | `core/ingest/stage_test.go` `TestStagedCloseAndDiscard` *(new)* | Close before any session → no temp left in the objects folder; `InsertTx` in a rolled-back write then `Discard(c)` → object purged, temp gone; `Discard` then `Close` → no error, committed object kept | AC-2.2, D-6 |
| T-3 | Go | `api/ffi/handlers/citations_test.go` `TestDepictionStagesBeforeSession` *(new)* | Session held by a blocked `catalogsession.Do`; `CreateCitationWithObservations` with one depiction in a goroutine → a `.stage-*` temp appears in the objects folder before release; after release the response matches the unblocked run; a refused write leaves no PNG object and no temp | AC-2.1, AC-2.2 |
| T-4 | Go | `core/derivatives/thumbnail_test.go` `TestPlanStageCommit` *(new)* | Table: no link → `Plan` returns a `Job`; `Job.Stage` while another `Do` holds the session → completes; `Commit` → link and derived File; File with a link → `Plan` returns it and no `Job`; non-raster → `Skipped`, no `Job` | AC-3.1, D-8 |
| T-5 | Go | `core/derivatives/thumbnail_test.go` `TestCommitAfterChange` *(new)* | Between `Stage` and `Commit`: another `Ensure` made the link → `Commit` returns that link, staged object not left unreferenced; the source File deleted → `ErrInvalid`, no link row, temp gone | D-8, D-10 |
| T-6 | Go | `core/ingest/stage_test.go` `TestPurgeLeavesTemps` *(new)* | A staged temp in `objects/` and `files.PurgeUnreferenced` for an unrelated unreferenced object → the temp still exists and installs afterwards | Reach row 1, D-3 |
| T-7 | Go | `api/ffi/handlers/thumbnails_test.go` `TestWorkspaceThumbnailPass` *(new)* | Table: Source with two raster Artifacts, no derivatives → both paths, two link rows; two Artifacts on one File → one link row, same path; a late-failing JPEG (`rastertest.OversizedJPEG` with its last two bytes, the EOI marker, cut off) → empty path, no link row; all derivatives present → no second session (assert with a `file_derivatives` count unchanged and no `.stage-*` temp created) | AC-3.2, AC-3.3, AC-3.4 |
| T-8 | Go | `api/ffi/handlers/artifacts_test.go` `TestIngestStagesBeforeSession` *(new)* | Session held by a blocked `catalogsession.Do`; `IngestArtifactFile` in a goroutine → an `.ingest-*` temp of the source's full size appears in the objects folder before release (poll with a 5 s deadline); after release the ingest returns the same Artifact, File and `Reused` as an unblocked run on a fresh project | AC-1.1, AC-1.2 |
| T-9 | Go | `api/ffi/handlers/cover_test.go` `TestSetCoverMakesThumbnailFirst` *(new)* | Cover to a raster Artifact with no derivative → cover set, one link row; to a PDF Artifact → `sources.ErrInvalid` code; to a late-failing JPEG (as in T-7) → `sources.ErrInvalid` code | AC-3.5 |
| T-10 | Go | `api/ffi/handlers/thumbnails_test.go` `TestEnsureFileThumbnail` *(existing)* | Unchanged assertions pass through the pass | AC-3.6 |
| T-11 | Go | `core/ingest/file_test.go` `TestFile`; `api/ffi/handlers/artifacts_test.go`, `citations_test.go`, `connect_test.go` *(existing)* | Unchanged assertions | AC-1.3, AC-2.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/ingest` (signature changes), `core/database/observations` (`observationstest`), `core/derivatives`, `api/ffi/handlers/thumbnails_test.go` and `cover_test.go`, `TestConcurrentCatalogRPCs`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `StagedImage`, `thumbnailRelPath`, `insertLinkOnly`, `errDerivedFileExists` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory (no UI change)
- [ ] Version bump: patch — VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ (not gated)
- [ ] Domain doc updated: source-layer-data-model.md lines 313 and 467 (file work and decoding before the session; install and purge inside)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source on a 60-graph census. Attach a scan of at least 300 MB to an Artifact. While it copies, switch to another Source page and open an Evidence graph: both load before the attach finishes.
2. On a Source whose Artifacts have no thumbnails yet (run `DELETE FROM file_derivatives` on a copy of the project with the app closed), open the Source page while another pane loads: the other pane is not held until the thumbnails finish; the thumbnails appear as before.
3. Draw a depiction on a Citation and save it: it saves and shows as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| An install and a purge of the same checksum interleave | Install and purge stay inside `Do` (D-3, D-6); T-6 |
| A staged temp leaks when the session never opens | `withStaged` defers `Close` outside the session (D-7); T-2 |
| A response field changes because thumbnails fill after the first session | `pass.run` runs before `proto.Marshal`; T-7 and existing handler tests compare responses |
| The second session adds latency to every Source page | No job → no second session (T-7's last row) |

**Stop and report if:**

- `catalogsession.Do`, `ingest.File`, `StageImage`, `derivatives.Ensure` or `thumbnailRelPath` no longer does what §4 says.
- An existing assertion in `TestFile`, `TestStageImage*`, `TestStagedImage*`, `TestEnsureThumbnail*`, `TestEnsureFileThumbnail` or `TestListSourcesAndWorkspaceThumbnails` must change for the tests to pass (only names, stage arguments and discard calls may change, STEP-2).
- `go list -deps ./core/ingest` comes to include `core/derivatives` or `core/database/sources` (an import cycle with D-8).
- A handler needs the catalog to decide what to stage before its session (D-2 assumes request fields are enough).
- A `WriteEffects` or audit revision count differs from today in any handler test.

## 12. Open questions

None.

## 13. PR

**Title:** S14-12 — File work before the catalog session
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-12-file-work-before-session.md`.
