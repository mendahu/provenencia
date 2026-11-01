# S14-12 — File work before the catalog session

**Status:** Landed in [#664](https://github.com/mendahu/provenencia/pull/664) (2026-11-01)
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
**Landed:** [#664](https://github.com/mendahu/provenencia/pull/664), 2026-11-01. Differences from the plan: (1) D-12: `sources.CoverArtifactFile` keeps a query of its own (`sqlArtifactForCover` stays) because `core/database/artifacts/artifacts_test.go` is an internal test that imports `sources`, so `sources` importing `artifacts` is an import cycle in that test; (2) `objectstore.CheckObjects` (new) runs before `Staged` installs, so an `objects/` swapped for a symlink after the stage is still refused as `ingest.unsafe_objects`, which keeps `TestStagedInsertTxRefusesSymlinkedObjects` unchanged; (3) the ingest test helper `objectFiles` skips staged temps so `TestStageImageRefuses` keeps its "no object written" assertion (`stageTemps` counts temps); (4) the `DELETE FROM file_derivatives` block moved to `damageCatalog` in the same commit as the thumbnail pass, not its own; (5) T-10 is a row in `TestEnsureFileThumbnail`; (6) `Job` has no exported accessor, the pass keys jobs by the File id it already holds; (7) the version was 0.2.4 at branch time, so the bump is 0.2.5.
**Planned:** 2026-11-01 against `main` @ `a652313`
**Review:** [`S14-12-file-work-before-session.review.md`](S14-12-file-work-before-session.review.md) — Fits with changes / Ready with changes → Fixed, 2 rounds

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

Attaching a large scan, saving a depiction, setting a Source cover, or opening a Source page whose images have no thumbnail yet no longer holds the catalog session while Go copies, hashes, decodes and encodes. Each of those paths stages its bytes first, outside `catalogsession.Do`, and enters the session only to look up, insert and install. Other panes (Source page, graph loads, counts, search) run while the file work runs.

**Why now.** Audit A-24 and bar outcome 14. Since S14-23, a JPEG over 64 MP that fails late is decoded in full on every Artifact response (565 ms for a cut-short 11,000 × 18,000 scan, up to 3 s for a crafted one, S14-23 ship log PRF-1), and today all of it runs under the session lock.

**What stays.** Every response, error code, audit revision and `WriteEffects` value is as today, except the error precedence named in D-13. Objects are installed and purged only inside `catalogsession.Do`, so a purge and an import of the same bytes still never interleave ([`source-layer-data-model.md`](../../../source-layer-data-model.md) line 313). Thumbnails are still made lazily by the responses that make them today; nothing stores a refused decode.

## 2. User stories

**US-1.** As a researcher, when I attach a large scan to an Artifact, the other panes keep loading while it copies, so that the app does not freeze for the length of the copy.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `IngestArtifactFile` while another call holds the catalog session | The copy, hash and sniff finish while the session is held: the `.ingest-*` temp in the project's objects folder reaches the source file's full size before the session is released |
| AC-1.2 | The session is released | The ingest returns the same Artifact, File and `Reused` as an unblocked run |
| AC-1.3 | The same file attached twice, a file over `ingest.MaxBytes`, an executable, a symlink | Same results and error codes as today (`TestFile` and the artifact handler tests unchanged in their assertions) |
| AC-1.4 | `IngestArtifactFile` with a valid file on an Artifact that already has one | `artifacts.ErrFileAlreadyAttached`, as today; no temp and no new object left |

**US-2.** As a researcher, when I save a depiction, its image is decoded and re-encoded before the save enters the catalog session.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `CreateCitationWithObservations` with one depiction while another call holds the session | The staged PNG temp (`.stage-*` in the objects folder) exists before the session is free; the save completes after it with the same response as today |
| AC-2.2 | A depiction whose write fails (Observation refused) | No object is left for the staged PNG and no temp is left in the objects folder |
| AC-2.3 | `UpdateObservation`, `AddObservationsToCitation`, `CreateCitedBridge` with a depiction | Each stages before its session through `withStagedDepictions`; their handler tests pass unchanged in their assertions |

**US-3.** As a researcher, when a Source page lists images with no thumbnail yet, or I pick a cover, the decodes run outside the catalog session.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `derivatives.Plan` returns a `Job` inside a session; another call then holds the session | `Job.Stage` completes while the session is held |
| AC-3.2 | `GetSourceWorkspace` on a Source with two raster Artifacts and no derivatives | Both `thumbnail_rel_path` fields are filled, as today; two `file_derivatives` rows exist after the call; no audit revision is added |
| AC-3.3 | Two Artifacts with the same File on one Source page | One decode, one derivative row; both protos carry the same path |
| AC-3.4 | A JPEG that fails late (no EOI, over 64 MP) on a Source page | The page returns with an empty `thumbnail_rel_path` for it; no derivative row |
| AC-3.5 | A Source page with 17 raster Artifacts and no derivatives | All 17 paths filled and 17 link rows written (D-10 batches them 16 + 1) |
| AC-3.6 | `SetSourceCover` to a raster Artifact whose thumbnail does not exist yet | The thumbnail is made outside the session and the cover is set, as today; a cover on a non-raster Artifact, a late-failing JPEG, or an Artifact of another Source returns `sources.ErrInvalid`, as today |
| AC-3.7 | `EnsureFileThumbnail` for a raster File with no derivative, and for a PDF | Same `RelPath` / `Skipped` as today |

## 3. Scope

| In | Covered by |
| --- | --- |
| `depictionvalue.ImageMediaType` *(new)*, used by `ingest` and `ReadableImage` | STEP-1 |
| One staged-object type, `ingest.Staged` *(replaces `StagedImage`)*: stage outside the session, insert/install inside, purge what the write did not use | STEP-3, AC-2.2 |
| `withStaged` / `withStagedDepictions` *(new)*: the one stage-then-write entry point; the four depiction handlers use it | STEP-3, AC-2.1, AC-2.3 |
| `database.RequireProject`, `catalogsession.Key` (exported), `catalogsessiontest.Hold` *(new)* test helper | STEP-2 |
| `ingest.StageFile`, `Staged.CommitFile` *(new)*; `IngestArtifactFile` stages before its session | STEP-4, AC-1.1 – AC-1.4 |
| `derivatives.Plan`, `Job.Stage`, `derivatives.Commit`, `derivatives.Refused` *(new)*; `Ensure` becomes their composition | STEP-5, AC-3.1 |
| `thumbnailPass` *(new)* for `artifactProto` callers, `EnsureFileThumbnail` and `SetSourceCover`; `sources.CoverArtifactFile` *(new)*; `NormalizeCover` lookup-only; one link-to-path tail (`derivedRelPath`) | STEP-6, AC-3.2 – AC-3.7 |
| `thumbnails_test.go` uses the harness's `damageCatalog` in place of its raw `DELETE FROM file_derivatives` (Spike 12 ship-log leftover) | STEP-7 |
| `use-catalog-session`, `add-ffi-handler`, `add-depiction-value`, the maintainability and security lenses, and `source-layer-data-model.md` name the new shape | STEP-8 |

| Out | Home |
| --- | --- |
| Oversized-JPEG thumbnails | S14-23 (landed) |
| Storing crops below 2,048 px | [depictions and likenesses idea](../../../ideas/depictions-and-likenesses.md) |
| Remembering a refused decode so a late-failing JPEG is not decoded again on the next response (PRF-1's repeat cost) | not planned — this PR takes the decode off the session; reopen from the dogfood log if the Source page itself is slow |
| Decoding several thumbnails in parallel | not planned — `raster.Slot` bounds memory; one handler decodes in sequence as today |
| A Swift change, or an `EnsureFileThumbnail` call from the app | not planned — the app reads thumbnail paths from responses, as today |
| Moving `ingest.File` and `derivatives.Ensure` into test-helper packages | not planned — about 35 test call sites; D-5 and D-8 mark them test-and-fixture only and STEP-8 adds them to `use-catalog-session` "Do not" |
| PDF cover thumbnails | [PDF thumbnails idea](../../../ideas/artifact-pdf-thumbnails.md) |

## 4. Current state

Verified at `a652313`.

| Area | Where | Today |
| --- | --- | --- |
| Session lock | `core/catalogsession/session.go` `Do`, `runLocked`, `key` | One mutex per project; `fn` runs with it held. `key` trims and `filepath.Clean`s; empty → `database.ErrNotAProject`. `api/ffi/handlers/catalog.go` `withProjectCatalog` wraps `Do` |
| Import | `core/ingest/file.go` `File(c, absPath, userID)` | `database.RequireUserID` → `openChecked` → `streamSource` (stream up to 512 MiB to an `objectstore.Temp` in `objects/.ingest-*`, SHA-256, 512-byte sniff) → `mediatypes.Resolve` → `files.StorageRelPath` → `LookupByChecksum` (reuse: `tmp.Install`, then `reuseFile` Run) or `tmp.Install` → `writes.Run` `create_file` with `insertFile`; a lost checksum race falls back to reuse; a failed write purges what it wrote. All of it runs inside the caller's session |
| Import call site | `api/ffi/handlers/artifacts.go` `IngestArtifactFile` | Inside one `withProjectCatalog`: `artifacts.Get`, `ErrFileAlreadyAttached`, `ingest.File`, `update_artifact` Run, `artifactProto` |
| Depiction stage | `core/ingest/stage.go` `StageImage(c, absPath)`, `StagedImage` | Takes a catalog only for `c.DB()` and `c.Dir()`; reads ≤32 MiB under `raster.Slot`, `raster.EncodePNG` (PNG/JPEG, ≤2048 px) into memory; `InsertTx(tx *sql.Tx)` looks up by checksum in the tx, inserts with `insertFile`, then `objectstore.WriteIfAbsent`; `Discard()` purges through the catalog it kept. `"image/png"` literal at lines 76, 104, 128 |
| Depiction batch | `core/ingest/depictions.go` `StageDepictions(c, inputs) (discard func(), err)` | Stages each `observations.Input.Depiction` with no `Image`; discards all on error |
| Depiction call sites | `api/ffi/handlers/citations.go` `CreateCitationWithObservations` (line 36); `observations.go` `UpdateObservation` (64), `AddObservationsToCitation` (143); `connect.go` `CreateCitedBridge` (53) | Each calls `ingest.StageDepictions(c, …)` first inside `withProjectCatalog`, then `defer discard()` |
| Depiction interface | `core/database/observations/depiction.go` `StagedFile` (comments lines 16, 24); `observations.go` line 177 comment | `InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)`; comments name `*ingest.StagedImage` and `ingest.StageDepictions` |
| Test helper | `core/database/observations/observationstest/observationstest.go` `SaveCitation`, `UpdateObservation`, and `ingest.File` at line 64 | Stage inside the test's catalog |
| Stored image type | `core/depictionvalue/depictionvalue.go` `ReadableImage` (line 48; compare at 49), doc comment line 33 | Compares `mediaType != "image/png"`; comment names `ingest.StageImage` |
| Writes | `core/writes/run.go` `Run` | Records an audit revision unless every change is a none-effect change (`effects.AllNone`); a `file` create change is not none, a `file_derivative` change is |
| Thumbnail | `core/derivatives/ensure.go` `Ensure`, `EnsureThumbnail`, `generate`, `readSourceObject`, `insertLinkOnly`, `errDerivedFileExists`, `errDerivativeLinkExists` | `files.Lookup` → `raster.Supported` → existing link → `MaxDecodeBytes` → `ReadRelPath` → `generate` (Slot, bounded checksum-verified read, `raster.EncodeJPEG`) → `WriteIfAbsent` → `LookupByChecksum` → `writes.Run(writes.Op{})` inserting the `files` row with plain `files.Insert` (no change) and the link (`insertDerivativeLink`, search reprojection after commit). All inside the caller's session |
| Thumbnail call sites | `api/ffi/handlers/thumbnails.go` `thumbnailRelPath`; callers `artifactProto` (`artifacts.go`: `CreateArtifact`, `UpdateArtifact`, `IngestArtifactFile`, `listArtifactsProto` ← `sources.go` `GetSourceWorkspace`) and `EnsureFileThumbnail` | Each ensures inside `withProjectCatalog`; `ErrUnprocessable`, `ErrCorruptObject`, `ErrInvalid` → empty path, skipped; other errors fail the call. `EnsureFileThumbnailResponse` carries `RelPath` and `Skipped` |
| Cover | `core/database/sources/sources.go` `NormalizeCover` → `requireRasterCoverArtifact` (lines 471–500; reads the Artifact with its own `sqlArtifactForCover`, line 59, beside `core/database/artifacts/artifacts.go` `Get`, line 293) → `derivatives.EnsureThumbnail`, the same three errors → `ErrInvalid`; handler `api/ffi/handlers/sources.go` `SetSourceCover` | Ensures the cover's thumbnail inside the session before the `set_source_cover` Run |
| Lookup-only thumbnails | `thumbnails.go` `existingThumbnailRelPath`, `sourceCoverThumbnail(s)` | Never decode. `existingThumbnailRelPath` maps a link to a path with its own `files.Lookup` + `ReadRelPath` (a missing derived row → empty), a second copy of `thumbnailRelPath`'s tail (which returns that error) |
| Swift | `macos/App` | Never calls `EnsureFileThumbnail`; reads `thumbnailRelPath` from Artifact protos (`macos/App/Platform/GoStore.swift` line 2071) |
| Session tests | `core/catalogsession/session_test.go`, `api/ffi/handlers/catalog_session_test.go` `TestConcurrentCatalogRPCs` | Each hand-rolls a goroutine blocked in `Do` |
| Tests | `core/ingest/file_test.go` `TestFile`; `stage_test.go` `TestStageImageRefuses`, `TestStageImageReencodes`, `TestStagedImageInsertTx`, `TestStagedImageDiscard`, helper `stageInsert`; `unsafe_objects_test.go` `TestStagedImageInsertTxRefusesSymlinkedObjects`; `core/derivatives/thumbnail_test.go` `TestEnsureThumbnail*`; `core/database/sources/sources_test.go` `runSetCover` (line 54); `core/database/artifacts/artifacts_test.go` line 626; `core/database/sourcecopy/sourcecopy_test.go` line 174; `api/ffi/handlers/thumbnails_test.go` `TestEnsureFileThumbnail`, `TestListSourcesAndWorkspaceThumbnails` (raw `DELETE FROM file_derivatives` block, lines 205–225); harness `api/ffi/handlers/harness_test.go` `damageCatalog` (156), `sourceWithPNG` (130), `sourceFixture` (213) | Pass at the base |

**Precedent.** `3a24ed95` ("Stage an app-rendered image as a checked PNG that a write stores with its File row") split the depiction image into a stage before `writes.Run` and an `InsertTx` inside it. This PR moves that split's boundary from "before the write" to "before the session" and gives the import and the thumbnail the same type. Copy its `InsertTx` contract and its tests' shape; do not copy keeping the encoded bytes in memory (D-3).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | One type, `ingest.Staged`, replaces `StagedImage`. Fields: project folder, `*objectstore.Temp`, checksum, media type, byte size, storage path (`files.StorageRelPath`), sanitized filename (set by `StageFile` only), and the `wrote` / `file` / `haveFile` state `StagedImage` keeps. Constructors: `StageFile` (import), `StageImage` (depiction), `StageBytes` (bytes Go made: the thumbnail) | Working rule: one staged-object type in `ingest` | A type per path — the rule forbids it |
| D-2 | Every constructor takes the project folder (`projectDir string`), never a `*database.Catalog`, and first calls `database.RequireProject(dir)` *(new: returns exactly what `locateCatalog` returns, so the same error `database.Open` gives: the raw `os.Stat` error for a missing folder, `database.ErrNotAProject` for a non-folder or a folder with no catalog)*, so a stage never creates an objects folder in a folder that is not a project | A catalog exists only inside `Do`; `objectstore.CreateTemp` makes the objects folder when absent | Opening a session to read `c.Dir()` — it is the lock this PR avoids |
| D-3 | Every stage writes its bytes to an `objectstore.Temp` in the objects folder outside the session (`.ingest-` for imports, `.stage-` for images and thumbnails). Inside the session only `Temp.Install` (a rename) runs | Install is cheap and the same for all three; the rename never copies across volumes | PNG/JPEG bytes in memory with `WriteIfAbsent` inside the session — file writing inside the lock; a temp in `os.TempDir` — a cross-volume move copies inside the lock |
| D-4 | `Staged.InsertTx(tx *sql.Tx)` keeps `StagedImage.InsertTx`'s contract: look up by checksum in the tx, insert with `insertFile` (setting `original_filename` only when the stage has one), install the temp, record `wrote`; a second call returns the same File and no change. It satisfies `observations.StagedFile` | The depiction write and the thumbnail commit both run in a tx; the in-tx lookup removes the checksum race `derivatives` handles today | A new interface — `StagedFile` fits |
| D-5 | `Staged.CommitFile(c, userID) (Result, error)` is today's `File` body after `files.StorageRelPath`, reading checksum, media type, byte size, storage path and filename from the stage: `LookupByChecksum` (reuse: `Temp.Install` repairs a missing or mismatched object, else removes the temp; then the `reuse_file` Run), else install, `create_file` Run, lost race → reuse, failure → purge. A stage with no filename (from `StageImage` or `StageBytes`) returns `ErrInvalid`. `CommitFile` checks `database.RequireUserID` first. `ingest.File(c, absPath, userID)` keeps its signature and its refusal order: `RequireUserID`, then `StageFile(c.Dir(), …)`, `defer stages.Settle(c)`, `CommitFile`; its doc comment says it does file work inside the caller's session and is for tests and fixtures only | Keeps the import's audit actions and error paths; keeps about 35 test call sites | One `InsertTx` Run for imports — it would change a reuse's audit action |
| D-6 | Two cleanup methods. `Staged.Discard()` runs outside the session: it removes the temp (a no-op after `Install`) and never touches an installed object, as `Temp.Discard` does. `Staged.Settle(c *database.Catalog)` runs inside the session after the write: when this stage wrote its object and no `files` row names the checksum, it calls `files.PurgeUnreferenced`; then it clears `wrote` and calls `Discard`. Both are idempotent. The purge half is unexported (`purge`). `ingest.Stages` *(new, `[]*Staged`)* has `Discard()` and `Settle(c)`. `Settle` is the one in-session cleanup: `withStaged`, `ingest.File`, `derivatives.Ensure` and `observationstest` all `defer stages.Settle(c)` | Purge must stay inside `Do` (line 313); a stage that never reached the session must still lose its temp; one in-session cleanup, so the four compositions cannot drift | `Close` / `Discard(c)` — `Temp.Close` keeps the file; an exported `Purge` beside `Discard` — four callers would each write the order |
| D-7 | `catalogsession.key` is exported as `Key(projectDir string) (string, error)`. Handlers stage through one entry point in `api/ffi/handlers/catalog.go`: `withStaged(projectDir string, stage func(dir string) (ingest.Stages, error), fn func(*database.Catalog) error) error` *(new)*: it normalizes with `catalogsession.Key`, calls `stage(dir)`, defers `stages.Discard()`, runs `withProjectCatalog(dir, …)` and defers `stages.Settle(c)` first inside it. On a `stage` error it returns the error; the stage function itself discards what it staged before the error. `withStagedDepictions(projectDir, inputs, fn)` *(new)* is `withStaged` with `ingest.StageDepictions` as `stage`. The four depiction handlers, `IngestArtifactFile` and `thumbnailPass` use them. Outside `ingest`, only `withStaged` calls `Discard`, and only `withStaged`, `derivatives.Ensure` and `observationstest` call `Settle` | Map: "one stage-then-write entry point"; one folder normalization; the order of cleanup lives in `Settle` and `withStaged` | Per-handler defers — six copies of the ordering |
| D-8 | Thumbnails split in `derivatives`. `Plan(c, sourceFileID, spec) (Result, *Job, error)` *(new)* is `Ensure` up to `generate`: it returns a final `Result` and nil `Job`, or a `Job` (source File id, spec, source object path, source checksum). `(*Job).Stage(projectDir string) (*ingest.Staged, error)` *(new)* holds `raster.Slot`, reads with `readSourceObject`, `raster.EncodeJPEG`, then `ingest.StageBytes(projectDir, jpeg, "image/jpeg")`; it keeps the stage on the `Job` and returns it so the caller can add it to its `Stages`. `Commit(c, j) (Result, error)` *(new)* runs inside the session, outside any tx: `files.Lookup` the source File (gone, or its checksum differs from the Job's → `ErrInvalid`); an existing link → return it (the stage stays unused; the caller's `Settle` cleans it); else one `writes.Run(writes.Op{})` doing `j.staged.InsertTx(tx.Tx)`, **dropping the `file` change it returns**, then `insertDerivativeLink`; `errDerivativeLinkExists` → `filederivatives.Lookup`, as today; the after-commit search reprojection error, as today. `Commit` does not purge. `Ensure` = `Plan` → `Stage(c.Dir())` → `defer Settle(c)` → `Commit`, with its error contract unchanged; its doc comment says it decodes inside the caller's session and is for tests only. `errDerivedFileExists` and `insertLinkOnly` go | Map: "the thumbnail splits into read, decode outside, and write". Derivative Files stay unaudited (a `file` change would make `writes.Run` record a revision) | Moving `Plan` / `Commit` into handlers — the derivative rules belong to `derivatives`; a `Staged` "unaudited" flag — one call site needs it |
| D-9 | `derivatives.Refused(err error) bool` *(new)*: true for `ErrUnprocessable`, `ErrCorruptObject`, `ErrInvalid`. It is the one "show no thumbnail" predicate; the pass uses it, and `thumbnailRelPath`'s copy goes | One error mapping, not two | — |
| D-10 | Handlers make thumbnails through `thumbnailPass` *(new, in `api/ffi/handlers/thumbnail_pass.go` (new))*. `need(c, fileID []byte, fill func(relPath string, skipped bool)) error` runs inside a session: a File id that is not 16 bytes → `fill("", true)`; else `Plan`; a final link → `fill(derivedRelPath(c, link))` at once; `Skipped` or `Refused(err)` → `fill("", true)`; another error → returned; a `Job` → recorded once per File id, with every `fill` for that File appended. `derivedRelPath(c, link) (string, bool, error)` *(new, in `thumbnails.go`)* is the one link-to-path tail: `files.Lookup(link.DerivedFileID)` (`sql.ErrNoRows` → `""`, skipped, as `existingThumbnailRelPath` does today) + `files.ReadRelPath` (empty → skipped). `existingThumbnailRelPath` becomes `filederivatives.Lookup` + `derivedRelPath`. `run(projectDir string, then func(*database.Catalog) error) error` runs after the first session: jobs go in batches of 16 (`thumbnailBatch` const); each batch calls `withStaged` with a `stage` that calls `Job.Stage` per job (a `Refused` error → that job's fills get `("", true)` and the job drops; another error → the batch's stages are discarded and the error returned) and an `fn` that `Commit`s each staged job and fills (`Refused` → `("", true)`; another error → returned, failing the call). `then`, when not nil, runs inside the last batch's session, or in one `withProjectCatalog` when no batch ran. No job and nil `then` → no second session | Every production decode leaves the session; at most 16 temps (and 16 open handles on the objects folder) at once; the cover write shares the commit session | One session per job — N lock round trips; staging every job at once — unbounded open files on a large Source |
| D-11 | Callers of the pass. `artifactProto(c, a, pass)` and `listArtifactsProto(c, id, pass)` take it; `artifactProto`'s fill sets `ThumbnailRelPath`. `CreateArtifact`, `UpdateArtifact`, `IngestArtifactFile`, `GetSourceWorkspace` and `EnsureFileThumbnail` call `pass.run(dir, nil)` after their session and before `proto.Marshal`, and only when their session returned nil. `EnsureFileThumbnail`'s fill sets `RelPath` and `Skipped`. `thumbnailRelPath` goes | One mechanism for every response that carries a thumbnail | — |
| D-12 | The cover rule lives in `sources` once. `sources.CoverArtifactFile(c, sourceID, artifactID) ([]byte, error)` *(new)* reads the Artifact with `artifacts.Get` (`sql.ErrNoRows` → `ErrInvalid`) and checks it belongs to the Source and has a File (else `ErrInvalid`); `sqlArtifactForCover` goes (`sources` → `artifacts` adds no cycle: neither imports the other). `requireRasterCoverArtifact` becomes `CoverArtifactFile` + `filederivatives.Lookup(c, fileID, TypeThumbnail)`: no link → `ErrInvalid`; it never decodes, and `sources` stops importing `derivatives`. `SetSourceCover`: when `strings.TrimSpace` of the request's mode is `sources.CoverModeArtifact` and the primary id parsed to 16 bytes, a first session calls `CoverArtifactFile` (its error is returned, as today's `NormalizeCover` would) and `pass.need(c, fileID, nil)`; then `pass.run(dir, then)` where `then` is today's session body (`NormalizeCover`, the `set_source_cover` Run, `enrichSourceProto`). Any other mode, or artifact mode with an empty id, skips the first session and `then` runs in one session, where `NormalizeCover` decides (an empty id → `ErrInvalid`, as today). `need` accepts a nil `fill` | The cover is decided by `NormalizeCover` alone (no handler pre-check); no production path decodes inside `Do` | A handler that maps a skipped thumbnail to `ErrInvalid` — a third copy of the rule |
| D-13 | `IngestArtifactFile` stages before checking the Artifact. When a request has two faults, the file's error (missing, symlink, too large, refused type, not a project) now wins over `artifacts.Get`'s or `ErrFileAlreadyAttached`; an already-attached Artifact pays the copy and then discards it. Single-fault errors are unchanged | The app offers Attach only on an Artifact with no File; a pre-check session would add a lock round trip to every attach | A first session for `artifacts.Get` — the extra round trip buys nothing on the app's paths |
| D-14 | `depictionvalue.ImageMediaType = "image/png"` *(new const)*; `ingest.StageImage` and `ReadableImage` use it. Thumbnails keep their own `"image/jpeg"` | Map scope; one owner for the stored depiction type | A const in `ingest` — `depictionvalue` is the pure type both read |
| D-15 | Session-freedom tests use `catalogsessiontest.Hold(t, projectDir) (release func())` (new, in `core/catalogsession/catalogsessiontest/catalogsessiontest.go` (new)): it starts `Do` in a goroutine whose `fn` signals it is running and blocks on a channel, waits for the signal, and registers `release` with `t.Cleanup`. Tests that wait for a temp poll the objects folder every 10 ms with a 5 s deadline. A test that holds a project builds its fixtures and runs `Plan` and `Commit` inside `catalogsession.Do` (fixture helpers such as `installSourceFile` take the session's `c`), never on a catalog it opened with `database.Open`, which would hold the project's exclusive lock (`ErrAlreadyOpen`); it calls `Hold` only around the stage, and releases before the next `Do` | One helper for a racy handshake; `catalogsession` does not depend on `ingest`, `derivatives` or `sources` (`go list -deps ./core/catalogsession`), so no cycle | A test seam in production code |

**Rules this plan bends.** None. The `add-ffi-handler` sentence "stages it inside `withProjectCatalog`" changes because the working rule replaced it (STEP-8).

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `ingest.Staged` *(renamed from `StagedImage`, generalized)* | type | `core/ingest/stage.go` `StagedImage` | It is that type, extended; "Image" goes because it carries imports |
| `ingest.StageFile` | func | `ingest.File` (first half), `streamSource` | Reuses `openChecked`, `streamSource`, `mediatypes.Resolve` |
| `ingest.StageBytes` | func | `objectstore.WriteIfAbsent` | That writes in place; a stage defers install to the session |
| `Staged.CommitFile` | method | `ingest.File` (second half) | Moved, not rewritten |
| `Staged.Settle`, `Staged.Discard`, `ingest.Stages` | methods, type | `StagedImage.Discard`, `StageDepictions`'s `discardAll` closure | The closure becomes `Stages`; the split follows D-6 |
| `database.RequireProject` | func | `database.locateCatalog` (unexported) | Exports its check only |
| `catalogsession.Key` | func | `catalogsession.key` | Exported, not copied |
| `catalogsessiontest.Hold` | test helper | hand-rolled holds in `session_test.go`, `catalog_session_test.go` | Third and later copies; shaped like `observationstest` |
| `derivatives.Plan`, `Job`, `Job.Stage`, `derivatives.Commit` | funcs, type | `derivatives.Ensure`, `generate` | `Ensure`'s body cut at its two boundaries |
| `derivatives.Refused` | func | `thumbnailRelPath`'s and `requireRasterCoverArtifact`'s error checks | Replaces both copies |
| `sources.CoverArtifactFile` | func | `requireRasterCoverArtifact` (its checks), `artifacts.Get` | Reads through `artifacts.Get`; split out so the handler and `NormalizeCover` share it |
| `depictionvalue.ImageMediaType` | const | literals in `stage.go`, `depictionvalue.go` | Replaces them |
| `withStaged`, `withStagedDepictions` | helpers | `withProjectCatalog` | Wrap it; no second session doorway |
| `thumbnailPass`, `derivedRelPath`, `thumbnailBatch` | type, func, const | `thumbnailRelPath`, `existingThumbnailRelPath` | `thumbnailRelPath` splits into `need` and `derivedRelPath`; `existingThumbnailRelPath` keeps its lookup and uses `derivedRelPath` for its tail; no batch helper exists |

**Builds on.** `objectstore.Temp`; `files.LookupByChecksumTx`, `insertFile`, `files.PurgeUnreferenced`; `raster.Slot`; `withProjectCatalog`; `observations.StagedFile`; the `observationstest` helper-package shape.
**Deletes.** `StagedImage` (name), its in-memory `png` field and `WriteIfAbsent` call; `StageDepictions`'s returned closure; `derivatives.insertLinkOnly`, `errDerivedFileExists`, `Ensure`'s `WriteIfAbsent` and pre-Run `LookupByChecksum`; `thumbnailRelPath`; `requireRasterCoverArtifact`'s `EnsureThumbnail` call and error mapping, `sqlArtifactForCover`, and the `sources` → `derivatives` import; `existingThumbnailRelPath`'s own `files.Lookup` + `ReadRelPath` tail; the `"image/png"` literals in `stage.go` and `ReadableImage`; the raw `DELETE FROM file_derivatives` block in `thumbnails_test.go`.
**Next adds.** A later derivative (a medium preview) is a `Spec` through `Plan` / `Job.Stage` / `Commit` and the pass. A later client file (an attachment on a note) is a `stage` func for `withStaged`. A PDF cover ([idea](../../../ideas/artifact-pdf-thumbnails.md)) changes only `derivatives` (a raster for a PDF); `requireRasterCoverArtifact`'s link lookup accepts it unchanged.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A staged temp exists while another call holds the session | the objects folder (`objectstore.Temp`) | Purge removes only objects named by checksum, never `.ingest-*` / `.stage-*` temps (`files.PurgeUnreferenced` builds paths from checksums) — T-6 |
| A row or object changes between the pass's sessions | `derivatives.Commit` | Re-checks the source File and its checksum (gone → `ErrInvalid` → empty path) and the link (made meanwhile → returned, stage settled by `withStaged`) — T-5 |
| A stage runs against a folder that is not a project | `database.RequireProject` | The error `database.Open` returns today (`ErrNotAProject`, or the raw error for a missing folder); no objects folder made — T-1 |
| `NormalizeCover` on an Artifact with no thumbnail link | `requireRasterCoverArtifact` | `ErrInvalid`; reached only when the pass refused the File or a delete raced it — T-9, T-12 |

### Scale

At the map's Scale (a 60-graph census; scans up to `ingest.MaxBytes` 512 MiB; a Source page of up to a few hundred Artifacts): the import holds the session for one rename and two short Runs instead of a 512 MiB copy and hash. A Source page with N Artifacts lacking thumbnails holds the session for its read pass, then ⌈N/16⌉ times for 16 small Runs each, with the decodes (one `raster.Slot` at a time, each ≤ `raster.MaxDecodeBytes`) between them outside it. Open handles on the objects folder are at most 16 per call. Staged thumbnails are ≤256 px JPEGs on disk, so peak memory is one decode, as today. No new cache, no new key, no `WriteEffects` change, no audit revision: derivative rows stay unaudited (D-8).

### Edges

- **Assumptions.** `files.PurgeUnreferenced` assumes no import is mid-install: installs and purges stay inside `Do` (D-3, D-6), held by T-6. `insertDerivativeLink` assumes the derived File row exists in the same tx: `Commit` runs `InsertTx` first in that Run (T-5). `observations.StagedFile` callers assume a second `InsertTx` returns the same File (`TestStagedInsertTx`, renamed, unchanged assertions). `writes.Run` audits any non-none change: `Commit` drops `InsertTx`'s `file` change (T-5 asserts no revision).
- **Neighbors.** A File with an existing link: `Plan` returns it, no job (T-4). A non-raster File: `Skipped`, no job (existing `TestEnsureThumbnailSkipsNonImage`; T-10 PDF). A late-failing JPEG: `Stage` → `ErrUnprocessable` → empty path (T-7). Exactly 16 jobs: one batch; 17: two (T-7).
- **Removed inputs.** Artifact or File deleted between the pass's sessions: `Commit` → `ErrInvalid` → empty path; the stage is purged and discarded (T-5). A depiction write refused: `Settle(c)` removes the installed PNG (T-2, existing discard test renamed).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** Imports capped at `maxBytes` before and during the copy (`openChecked`, `LimitReader`); depiction images at `maxImageBytes`; thumbnail sources at `raster.MaxDecodeBytes`, verified against the stored checksum before decode. Jobs per response are bounded by Artifacts on the page, deduplicated by File id, and staged 16 at a time.
- **Imports.** `derivatives` → `ingest` brings `core/database/observations` (through `core/ingest/depictions.go`) under `derivatives`. `sources` no longer imports `derivatives` (D-12), so no `sources` → `observations` path appears through it. Stop if `go list -deps ./core/ingest` includes `derivatives` or `sources`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database` | `func RequireProject(dir string) error` *(new)* |
| `core/catalogsession` | `key` → exported `Key`; callers in `session.go` follow |
| `core/catalogsession/catalogsessiontest` *(new)* | `Hold(t testing.TB, projectDir string) (release func())` |
| `core/depictionvalue` | `const ImageMediaType = "image/png"` *(new)*; `ReadableImage` uses it |
| `core/ingest/stage.go` | `Staged` (D-1); `StageImage(projectDir, absPath)`; `StageBytes` *(new)*; `InsertTx`; `Settle(c)`; `Discard()`; `Stages` *(new)* |
| `core/ingest/file.go` | `StageFile` *(new)*; `CommitFile` *(new)*; `File` = composition (D-5) |
| `core/ingest/depictions.go` | `StageDepictions(projectDir string, inputs []observations.Input) (Stages, error)`; on error it discards every stage so far and returns nil |
| `core/database/observations` | Comments in `depiction.go` (16, 24) and `observations.go` (177) name `*ingest.Staged` and `withStagedDepictions` |
| `core/derivatives/ensure.go` | `Job`, `Plan`, `(*Job) Stage`, `Commit`, `Refused` *(new)* (D-8, D-9); `Ensure` composes them |
| `core/database/sources/sources.go` | `CoverArtifactFile` *(new)*; `requireRasterCoverArtifact` lookup-only; `sqlArtifactForCover` deleted (D-12) |

```go
// sketch — signatures only
func StageFile(projectDir, absPath string) (*Staged, error)
func StageImage(projectDir, absPath string) (*Staged, error)
func StageBytes(projectDir string, data []byte, mediaType string) (*Staged, error)
func (s *Staged) CommitFile(c *database.Catalog, userID []byte) (Result, error)
func (s *Staged) InsertTx(tx *sql.Tx) (files.File, []rowchange.Change, error)
func (s *Staged) Settle(c *database.Catalog)
func (s *Staged) Discard()

func Plan(c *database.Catalog, sourceFileID []byte, spec Spec) (Result, *Job, error)
func (j *Job) Stage(projectDir string) (*ingest.Staged, error)
func Commit(c *database.Catalog, j *Job) (Result, error)
func Refused(err error) bool

func CoverArtifactFile(c *database.Catalog, sourceID, artifactID []byte) ([]byte, error)
```

Errors: unchanged codes. `RequireProject` → `catalog.not_a_project`.

### FFI

None — no proto change. Handler files: `catalog.go` (`withStaged`, `withStagedDepictions`), `citations.go`, `observations.go`, `connect.go`, `artifacts.go`, `sources.go`, `thumbnails.go`, and the pass's file (D-10).

`IngestArtifactFile` shape:

```go
var staged *ingest.Staged
pass := &thumbnailPass{}
err = withStaged(req.GetProjectDir(), func(dir string) (ingest.Stages, error) {
	s, err := ingest.StageFile(dir, req.GetPath())
	if err != nil { return nil, err }
	staged = s
	return ingest.Stages{s}, nil
}, func(c *database.Catalog) error {
	// artifacts.Get, ErrFileAlreadyAttached, staged.CommitFile(c, userID),
	// update_artifact Run, artifactProto(c, got, pass)
})
if err == nil { err = pass.run(req.GetProjectDir(), nil) }
```

### macOS

None — Go only.

### Strings

None — no copy, and no action changes what it covers.

## 7. Steps

Every step leaves `go build ./...` and `go vet ./...` passing, so each is a commit.

**STEP-1 — One stored-image media type**
Files: `core/depictionvalue/depictionvalue.go`, `core/ingest/stage.go`. Skill: `add-depiction-value`.
Do: add `ImageMediaType`; use it in `ReadableImage` and the three `stage.go` literals.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/depictionvalue/... ./core/ingest/...` passes; `rg '"image/png"' core/ingest/stage.go core/depictionvalue/depictionvalue.go` has no hit.

**STEP-2 — Session helpers**
Files: `core/database/catalog.go`, `core/catalogsession/session.go`, `core/catalogsession/catalogsessiontest/catalogsessiontest.go` *(new)*, `core/database/catalog_test.go`. Skill: `use-catalog-session`.
Do: add `RequireProject` (D-2); export `Key` and update its callers in `session.go`; add `Hold` (D-15). Add T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/ ./core/catalogsession/...` passes; `go list -deps ./core/catalogsession | rg "core/ingest$|core/derivatives"` has no hit.

**STEP-3 — Depiction images stage before the session**
Files: `core/ingest/stage.go`, `core/ingest/depictions.go`, `core/database/observations/depiction.go`, `core/database/observations/observations.go`, `core/database/observations/observationstest/observationstest.go`, `core/depictionvalue/depictionvalue.go` (comment), `api/ffi/handlers/catalog.go`, `citations.go`, `observations.go`, `connect.go`, `core/ingest/*_test.go`, `api/ffi/handlers/citations_test.go`, every other `*_test.go` that calls `StageImage` or `StageDepictions`. Skills: `use-catalog-session`, `add-ffi-handler`.
Do: build `Staged` with `StageImage`, `StageBytes`, `InsertTx`, `Settle`, `Discard`, `Stages` (D-1 – D-4, D-6); `StageDepictions` takes the folder and returns `Stages`; add `withStaged` and `withStagedDepictions` (D-7) and move the four depiction handlers onto `withStagedDepictions`. `observationstest` stages with `c.Dir()` and defers `stages.Settle(c)`. Existing tests change only `StagedImage` → `Staged` (test names follow: `TestStagedInsertTx`, `TestStagedSettle`, `TestStagedInsertTxRefusesSymlinkedObjects`), `c` → `c.Dir()` in stage calls, and the cleanup calls. Add T-1 (image and bytes rows), T-2, T-3, T-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./...` passes; `rg StagedImage --type go` has no hit; `rg "StageDepictions" api/ffi/handlers` hits only `catalog.go`.

**STEP-4 — Import stages before the session**
Files: `core/ingest/file.go`, `core/ingest/stage_test.go`, `api/ffi/handlers/artifacts.go`, `api/ffi/handlers/artifacts_test.go`.
Do: add `StageFile` and `CommitFile`; rewrite `File` as the composition with its refusal order and test-only doc comment (D-5); `IngestArtifactFile` per the FFI sketch with today's `artifactProto(c, got)` (the pass comes in STEP-6); D-13's order. Add T-1's file rows, T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./...` passes; `rg "ingest\.File\(" api/` has no hit.

**STEP-5 — Split the thumbnail derivative**
Files: `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: add `Job`, `Plan`, `Job.Stage`, `Commit`, `Refused` (D-8, D-9); `Ensure` composes them, with its test-only doc comment; delete `insertLinkOnly` and `errDerivedFileExists`. Add T-4, T-5 (fixtures per D-15).
Check: `CGO_ENABLED=1 go test -tags fts5 ./...` passes with every existing `TestEnsureThumbnail*` assertion unchanged.

**STEP-6 — Thumbnails and covers decode between sessions**
Files: `api/ffi/handlers/thumbnail_pass.go` *(new)*, `artifacts.go`, `sources.go`, `thumbnails.go`, `core/database/sources/sources.go`, `core/database/sources/sources_test.go`, `core/database/artifacts/artifacts_test.go`, `core/database/sourcecopy/sourcecopy_test.go`, `api/ffi/handlers/thumbnails_test.go`, `api/ffi/handlers/cover_test.go`.
Do: add `thumbnailPass`, `thumbnailBatch` and `derivedRelPath`, and rebuild `existingThumbnailRelPath` on `derivedRelPath` (D-10); thread the pass through `artifactProto` and `listArtifactsProto` and call `run` in the five handlers (D-11); add `CoverArtifactFile` on `artifacts.Get`, delete `sqlArtifactForCover`, make `requireRasterCoverArtifact` lookup-only and split `SetSourceCover` (D-12); delete `thumbnailRelPath`. The three core tests that set an artifact cover (`sources_test.go` `runSetCover` for `CoverModeArtifact`, `artifacts_test.go` line 626, `sourcecopy_test.go` line 174) call `derivatives.EnsureThumbnail` on the cover's File first; their assertions do not change. Add T-7, T-9, T-10's row, T-12.
Check: `CGO_ENABLED=1 go test -tags fts5 ./...` passes; `rg "EnsureThumbnail\(|derivatives\.Ensure\(" --type go -g '!*_test.go'` hits only `core/derivatives/ensure.go`; `go list -deps ./core/database/sources | rg core/derivatives` has no hit; `rg sqlArtifactForCover core/` has no hit.

**STEP-7 — Thumbnail test uses the harness**
Files: `api/ffi/handlers/thumbnails_test.go`.
Do: replace the `Close` / `database.Open` / `DELETE FROM file_derivatives` block (lines 205–225 at the base) with `damageCatalog(t, f.dir, "DELETE FROM file_derivatives")`, as `sources_test.go` line 374 does.
Check: `rg "DELETE FROM file_derivatives" api/ffi/handlers/thumbnails_test.go` hits only the `damageCatalog` call; `go test` passes.

**STEP-8 — Guidance and domain doc**
Files: `.agents/skills/use-catalog-session/SKILL.md`, `.agents/skills/add-ffi-handler/SKILL.md` (§ Files from the client, line 67), `.agents/skills/add-depiction-value/SKILL.md` (lines 7, 33, 37), `.agents/skills/review-pr/lenses/maintainability.md` (line 313), `.agents/skills/review-pr/lenses/security.md` (line 22), `docs/source-layer-data-model.md` (lines 313 and 467).
Do: `use-catalog-session` gains "File work before the session": stage outside `Do` through `withStaged` / `withStagedDepictions`, thumbnails through `thumbnailPass`, install and purge only inside `Do` (`Settle`), `catalogsessiontest.Hold` for tests; its "Do not" adds calling `ingest.File` or `derivatives.Ensure` from a handler. `add-ffi-handler` says the handler stages before the session through `withStaged`. The lenses and `add-depiction-value` name `ingest.Staged`. The domain doc says file work and decoding run before the session, and install and purge inside it.
Check: `python3 scripts/check-guidance-names.py` passes; `rg "StagedImage|stages it inside" .agents docs/*.md` has no hit (plans under `docs/deployment-plan/` are history and keep their names).

**STEP-9 — Docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row), this plan (status, Landed note), `VERSION`, `core/version.go`, `macos/Provenencia.xcodeproj/project.pbxproj` (`MARKETING_VERSION`).
Do: per the Definition of done.
Check: links resolve; `scripts/check-pr.sh` prints `check-pr: PASS` (its Xcode step runs on macOS).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

The `catalogsessiontest.Hold` helper and the 10 ms / 5 s poll (D-15) serve T-1, T-3, T-4 and T-8.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/ingest/stage_test.go` `TestStageOutsideSession` *(new)* | Table: `StageFile`, `StageImage`, `StageBytes` on a project held by `Hold` → each returns before release; on a folder with no catalog → `database.ErrNotAProject` and no objects folder made; on a missing folder → the `os.Stat` error, as `database.Open` | AC-1.1, AC-2.1, D-2 |
| T-2 | Go | `core/ingest/stage_test.go` `TestStagedPurgeAndDiscard` *(new)* | `Discard` with no session → no temp left in the objects folder; `InsertTx` inside a `writes.Run` whose fn then returns an error, then `Settle(c)` → object gone, temp gone; a committed write then `Settle` and `Discard` → object kept; `CommitFile` on a `StageBytes` stage → `ErrInvalid` | AC-2.2, D-5, D-6 |
| T-3 | Go | `api/ffi/handlers/citations_test.go` `TestDepictionStagesBeforeSession` *(new)* | Project held; `CreateCitationWithObservations` with one depiction in a goroutine → a `.stage-*` temp appears before release; after release the response equals an unblocked run's (ids aside); a refused write leaves no PNG object and no temp | AC-2.1, AC-2.2 |
| T-4 | Go | `core/derivatives/thumbnail_test.go` `TestPlanStageCommit` *(new)* | Table: no link → `Plan` returns a `Job`; `Job.Stage` while `Hold` holds the session → completes; `Commit` → link and derived File; File with a link → `Plan` returns it and no `Job`; non-raster → `Skipped`, no `Job` | AC-3.1, D-8 |
| T-5 | Go | `core/derivatives/thumbnail_test.go` `TestCommitAfterChange` *(new)* | `Commit` adds no audit revision; between `Stage` and `Commit`: `EnsureThumbnail` made the link → `Commit` returns that link and, after `Settle`, no unreferenced object is left; the source File row deleted → `ErrInvalid`, no link row | D-8, D-10 |
| T-6 | Go | `core/ingest/stage_test.go` `TestPurgeLeavesTemps` *(new)* | A staged temp in the objects folder and `files.PurgeUnreferenced` for an unrelated unreferenced object → the temp still exists and installs afterwards | Reach row 1, D-3 |
| T-7 | Go | `api/ffi/handlers/thumbnails_test.go` `TestWorkspaceThumbnailPass` *(new)* | Fixtures built as `sourceWithPNG` builds its Source and Artifact. Table: two raster Artifacts, no derivatives → both paths, two link rows, audit revision count unchanged by the read; two Artifacts on one File → one link row, same path; a late-failing JPEG (`rastertest.OversizedJPEG` with its last two bytes, the EOI marker, cut off) → empty path, no link row; 17 raster Artifacts (small PNGs) → 17 paths and 17 link rows; all derivatives present → response unchanged, `file_derivatives` count and the objects folder's file count unchanged | AC-3.2, AC-3.3, AC-3.4, AC-3.5 |
| T-8 | Go | `api/ffi/handlers/artifacts_test.go` `TestIngestStagesBeforeSession` *(new)* | Project held; `IngestArtifactFile` in a goroutine → an `.ingest-*` temp of the source's full size appears before release; after release the Artifact, File and `Reused` equal an unblocked run's on a fresh fixture; Artifact already attached + valid file → `ErrFileAlreadyAttached`, no temp and no new object | AC-1.1, AC-1.2, AC-1.4, D-13 |
| T-9 | Go | `api/ffi/handlers/cover_test.go` `TestSetCoverMakesThumbnailFirst` *(new)* | Table: raster Artifact with no derivative → cover set, one link row; PDF Artifact → `sources.ErrInvalid` code; late-failing JPEG (as T-7) → `sources.ErrInvalid` code; Artifact of another Source → `sources.ErrInvalid` code; type-icon mode → set, no `.stage-*` temp created | AC-3.6, D-12 |
| T-10 | Go | `api/ffi/handlers/thumbnails_test.go` `TestEnsureFileThumbnail` *(existing, one row added)* | Existing rows unchanged; a PDF File on an Artifact → `Skipped` true, empty `RelPath` | AC-3.7 |
| T-11 | Go | `core/ingest/file_test.go` `TestFile`; `api/ffi/handlers/artifacts_test.go`, `citations_test.go`, `connect_test.go` *(existing)* | Unchanged assertions | AC-1.3, AC-2.3 |
| T-12 | Go | `core/database/sources/sources_test.go` `TestNormalizeCoverNeedsThumbnail` *(new)* | Raster Artifact with no thumbnail link → `ErrInvalid`; after `derivatives.EnsureThumbnail` → artifact mode accepted; an Artifact of another Source → `ErrInvalid` | D-12 |
| T-13 | Go | `core/catalogsession/session_test.go` `TestHold` *(new)*; `core/database/catalog_test.go` `TestRequireProject` *(new)* | `Hold` → a second `Do` on the project blocks until `release`; `RequireProject` on a project → nil, on a folder with no catalog → `ErrNotAProject`, on a missing folder → an `fs.ErrNotExist` error | D-2, D-15 |
| T-14 | Go | `api/ffi/handlers/thumbnails_test.go` `TestListSourcesAndWorkspaceThumbnails` *(existing)* | Unchanged assertions through the rebuilt `existingThumbnailRelPath` | D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/ingest`, `core/database/observations` (`observationstest`), `core/derivatives`, `core/database/sources`, `api/ffi/handlers/thumbnails_test.go` and `cover_test.go`, `TestConcurrentCatalogRPCs`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `StagedImage`, `thumbnailRelPath`, `insertLinkOnly`, `errDerivedFileExists` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory — not applicable, no UI change
- [ ] Version bump: patch — VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — not applicable, not gated
- [ ] Domain doc updated: source-layer-data-model.md lines 313 and 467 (file work and decoding before the session; install and purge inside)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source on a 60-graph census. Attach a scan of at least 300 MB to an Artifact. While it copies, switch to another Source page and open an Evidence graph: both load before the attach finishes.
2. On a copy of a project with at least 20 image Artifacts on one Source, with the app closed, run `DELETE FROM file_derivatives` on its catalog. Open that Source page and, while it loads, open another pane: the other pane is not held until the thumbnails finish; the thumbnails appear as before.
3. Pick an image Artifact as that Source's cover: it is set as before.
4. Draw a depiction on a Citation and save it: it saves and shows as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| An install and a purge of the same checksum interleave | Install and purge stay inside `Do` (D-3, D-6: `Settle`); T-6 |
| A staged temp leaks when the session never opens | `withStaged` defers `Discard` outside the session (D-7); T-2 |
| A response field changes because thumbnails fill after the first session | `pass.run` runs before `proto.Marshal`; T-7, T-10 and existing handler tests compare responses |
| A thumbnail starts writing audit revisions | `Commit` drops `InsertTx`'s `file` change (D-8); T-5, T-7 |
| A large Source opens hundreds of files at once | Batches of 16 (D-10); T-7's 17-Artifact row |

**Stop and report if:**

- `catalogsession.Do`, `ingest.File`, `StageImage`, `derivatives.Ensure`, `thumbnailRelPath` or `requireRasterCoverArtifact` no longer does what §4 says.
- An existing assertion in `TestFile`, `TestStageImage*`, `TestStagedImage*`, `TestEnsureThumbnail*`, `TestEnsureFileThumbnail` or `TestListSourcesAndWorkspaceThumbnails` must change for the tests to pass (only the names, stage arguments and cleanup calls of STEP-3 and the three setup lines of STEP-6 may change).
- `go list -deps ./core/ingest` includes `core/derivatives` or `core/database/sources`, or `go list -deps ./core/catalogsession` includes `core/ingest` or `core/derivatives`.
- A handler needs the catalog to decide what to stage before its session (D-2 assumes request fields are enough).
- A `WriteEffects` value or audit revision count differs from today in any handler test.

## 12. Open questions

None.

## 13. PR

**Title:** S14-12 — File work before the catalog session
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-12-file-work-before-session.md`.
