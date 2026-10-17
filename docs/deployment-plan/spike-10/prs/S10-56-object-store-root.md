# S10-56 — Object store stays inside objects/; deletes prune empty folders

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 8 — Likeness
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-56--object-store-stays-inside-objects-deletes-prune-empty-folders)
**Brief:** — (no new chrome: an engine change to how object files are read, written, and removed)
**Board:** —
**Depends on:** S10-53 (#424, open: `objectstore.WriteIfAbsent`). This PR's branch is cut from #424's branch and its PR targets `s10-53-stage-image` until #424 merges.
**Blocks:** —
**Branch:** `s10-56-object-store-root`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** —
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [`.agents/skills/review-pr/lenses/security.md`](../../../../.agents/skills/review-pr/lenses/security.md) trust boundaries "Object store" and "Project open"; [`docs/source-layer-data-model.md`](../../../source-layer-data-model.md) §6; S10-53 ship log [`S10-53-stage-image.ship.md`](S10-53-stage-image.ship.md) SEC-1; S10-55 ship log [`S10-55-file-release.ship.md`](S10-55-file-release.ship.md) SEC-R2-1; the user's report (2026-10-17) that deletes leave empty folders under objects/
**Planned:** 2026-10-17 against `s10-53-stage-image` @ `39ee4860` (#424's head; `main` @ `1c99ef4e` plus S10-53)
**Review:** [`S10-56-object-store-root.review.md`](S10-56-object-store-root.review.md) — Fits, Ready with changes (applied); 2 rounds

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

Every object file Go reads, writes, installs, or removes stays inside the project's objects/ folder. A project folder from someone else whose objects/ or `objects/ab` is a symlink to a folder outside it cannot make an import, a staged image, or a thumbnail write outside the project, and a FIFO at an object path cannot block a read. A delete that removes the last object in a shard folder also removes the `objects/ab/cd` and `objects/ab` folders it leaves empty.

**Why now.** S10-55 made removal safe (`objectstore.RemoveObject` opens an `os.Root`), but writes still use plain `os` calls on joined paths: S10-53 moved the thumbnail write into `objectstore.WriteIfAbsent` and pointed staging at it (S10-53 review SEC-1). `RemoveObject` has a gap between its `Lstat` of `objects` and `os.OpenRoot` (S10-55 review SEC-R2-1). The user found many empty shard folders after testing deletes (2026-10-17): `RemoveObject` removes only the file.

**What stays.** Imports, staging, thumbnails, and deletes accept, refuse, store, and remove exactly what they do today in a normal project. No schema, FFI, or Swift change.

## 2. User stories

**US-1.** As a researcher who opens a project folder someone sent me, the app cannot be made to write or read files outside that folder through its objects/ folder.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | `objects/ab` is a symlink to a folder outside the project; `WriteIfAbsent` of bytes whose checksum starts `ab` | Error; the outside folder holds no new file. |
| AC-1.2 | `objects` itself is a symlink to an outside folder; `WriteIfAbsent`, `CreateTemp`, `OpenObject`, `RemoveObject` | `WriteIfAbsent`, `CreateTemp`, and `OpenObject` return `ErrUnsafeObjects`; `RemoveObject` returns an error; nothing outside is written, read, or removed. |
| AC-1.3 | The object path is a symlink to an outside file holding the right bytes; `WriteIfAbsent` | `wrote` is true, the object path is now a regular file with the bytes, and the outside file is unchanged. |
| AC-1.4 | The object path is a FIFO; `OpenObject` | `ErrNotRegular`, without blocking. |
| AC-1.5 | A thumbnail source whose object path is a symlink to an outside file holding the right bytes; `derivatives.EnsureThumbnail` | `ErrCorruptObject`, no `file_derivatives` row. |
| AC-1.6 | A temp from `CreateTemp` installed at a `rel` whose `objects/ab` is a symlink outside | `Install` returns an error, the temp is gone, nothing is written outside. |

**US-2.** As a researcher, deleting things does not leave empty folders behind in my project.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The only object in `objects/ab/cd`; `RemoveObject` | The object, `objects/ab/cd`, and `objects/ab` are gone; objects/ stays. |
| AC-2.2 | Two objects in `objects/ab/cd`; `RemoveObject` of one | The other object and both folders stay. |
| AC-2.3 | One object in `objects/ab/cd` and one in `objects/ab/ef`; `RemoveObject` of the first | `objects/ab/cd` is gone; `objects/ab` and the other object stay. |
| AC-2.4 | An Artifact whose File is the only object in its shard (and its thumbnail the only one in its shard); `artifacts.Delete` commits | Both objects and their shard folders are gone. |

**US-3.** As a contributor, imports, staging, thumbnails, and deletes behave as before.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Every existing test in `core/ingest`, `core/derivatives`, `core/database/files`, `core/database/artifacts`, `core/database/deleteimpact`, `core/objectpath` | Passes. Allowed edits: the `WriteIfAbsent` call in `core/database/deleteimpact/purge_test.go`, the `TestWriteIfAbsent` signature change, and new rows or tests this plan names. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `objectpath.Dir` *(new)*; private `innerPath`, `openObjectsRoot`, `openRegular` in `core/objectstore`; `RemoveObject` opens through the project root (SEC-R2-1) | STEP-1, AC-1.2 |
| `OpenObject`, `ErrNotRegular`, `ErrUnsafeObjects` *(new)* | STEP-1, AC-1.2, AC-1.4 |
| `WriteIfAbsent(projectDir, rel, checksum, data)` through the root on one private install sequence; callers in `core/ingest/stage.go` and `core/derivatives/ensure.go` | STEP-2, AC-1.1 to AC-1.3 |
| `CreateTemp` and `*objectstore.Temp` *(new)*; `ingest.streamSource` and `ingest.File` use them; `ingest.installObject`, the exported `MatchesChecksum`, and `SyncDir` deleted | STEP-3, AC-1.6, AC-3.1 |
| `derivatives.readSourceObject` reads through `OpenObject` | STEP-4, AC-1.5 |
| `RemoveObject` removes the shard folders it empties | STEP-5, AC-2.1 to AC-2.4 |
| Docs: security lens trust table, data model §6, map row, `completed.md` | STEP-6 |

| Out | Home |
| --- | --- |
| Removing empty shard folders already left in existing projects | Not planned; they are harmless, and the next removal in each shard clears its own. The user can delete them by hand |
| Swift reading object files by relative path (thumbnails, Quick Look) | Not planned; Swift opens paths Go returns, and Go now never writes outside objects/ |
| `core/database/catalog.go`'s private `objectsDir` moving onto `objectpath.Dir` | Not planned (S10-55 MNT-R2-1); it creates the folder once at catalog create |
| `objectstore.Checksum` helper for the inline SHA-256s (S10-53 MNT-3) | Not planned here; S10-53 follow-up |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Removal | `core/objectstore/objectstore.go` `RemoveObject(projectDir, rel string) error` | Requires `rel` to start objects/; `os.Lstat(projectDir/objects)` (missing → nil, symlink → error); `os.OpenRoot(projectDir/objects)`; `root.Remove(inner)`; a missing file is nil. Removes only the file. The `Lstat` and `OpenRoot` are separate path lookups (SEC-R2-1). Its one caller is `core/database/files/purge.go` `PurgeUnreferenced`. |
| Check | `core/objectstore/objectstore.go` `MatchesChecksum(objPath, wantChecksum string) (bool, error)` | `os.Open` on an absolute path (follows symlinks; a FIFO blocks the open), then `f.Stat()` regular check, then SHA-256. Missing → `false, nil`. Callers: `WriteIfAbsent`, `core/ingest/file.go` `installObject`. |
| Write | `core/objectstore/objectstore.go` `WriteIfAbsent(objPath, checksum string, data []byte) (bool, error)` | Verifies `data` against `checksum`, `MatchesChecksum`, `os.MkdirAll`, `os.CreateTemp(dir, ".write-*")`, write, sync, close, `os.Rename`, re-match on a rename failure, `SyncDir`. Follows a symlinked objects/ or shard (SEC-1). Callers: `core/ingest/stage.go` `InsertTx`, `core/derivatives/ensure.go` `Ensure`, `core/database/deleteimpact/purge_test.go`. |
| Dir sync | `core/objectstore/sync_unix.go`, `sync_other.go` `SyncDir(dir string) error` | Unix: `os.Open(dir)` + `Sync`. Other: no-op. Callers: `WriteIfAbsent`, `core/ingest/file.go` `installObject`. |
| Package doc | `core/objectstore/objectstore.go` line 1 | "Package objectstore verifies, writes, and removes content-addressed object files on disk." |
| Import | `core/ingest/file.go` `File`, `streamSource(src, objectsDir)`, `installObject(tmpPath, objPath, checksum)` | `File` passes `filepath.Join(c.Dir(), "objects")`; `streamSource` does `os.MkdirAll` and `os.CreateTemp(objectsDir, ".ingest-*")` and returns the absolute temp path; `File` removes it with `os.Remove(tmpPath)` on early failures; `installObject` matches, `os.MkdirAll`, `os.Remove(objPath)`, `os.Rename`, re-matches on a failure, `SyncDir`, always removes the temp. |
| Staging | `core/ingest/stage.go` `InsertTx` | `objectstore.WriteIfAbsent(filepath.Join(s.c.Dir(), filepath.FromSlash(s.rel)), s.checksum, s.png)`. |
| Thumbnails | `core/derivatives/ensure.go` `Ensure`, `generate(objPath, wantChecksum, opts)`, `readSourceObject(objPath, wantChecksum)` | `Ensure` joins `c.Dir()` and `rel`; `readSourceObject` does `os.Open` (follows symlinks; a FIFO blocks), reads at most `maxSourceBytes+1`, and returns `ErrCorruptObject` on a checksum mismatch. The derived JPEG is written with `WriteIfAbsent(derivedPath, checksum, derivedJPEG)`. |
| Layout | `core/objectpath/objectpath.go` `Rel` (line 87 builds `"objects/" + …`); `core/database/catalog.go` private `objectsDir = "objects"` (creates the folder at line 108) | `objects/<hh>/<hh>/<sha256><ext>`. `objectpath` imports only `errors` and `strings`. |
| Tests | `core/objectstore/objectstore_test.go` (`package objectstore_test`, no build tag) `TestRemoveObject` (six table rows, three of them using `os.Symlink`), `TestWriteIfAbsent` (a joined absolute path; its mismatch case asserts no folder is created); `core/database/artifacts/artifacts_test.go` `TestArtifactDeletePurgesDisk`, `runDelete`, `objectPath`; `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources`; `core/database/observations/observations_test.go` (a sentinel file at `objects/ab/cd/sentinel`, near line 593) | — |
| Go | `go.mod` `go 1.27` | `os.Root` has `OpenRoot`, `Lstat`, `OpenFile`, `Mkdir`, `MkdirAll`, `Remove`, `Rename`. |

**Precedent.** S10-55 F-4 (`48fd37bc`, "Keep a purge from removing files outside the project's objects folder through a symlink") added `RemoveObject` with `os.OpenRoot` and its table test; every function here follows that shape and that test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `objectpath.Dir = "objects"` *(new, exported const)* in `core/objectpath/objectpath.go`; `Rel` builds its prefix from it. Every exported function in `core/objectstore` that names an object takes `projectDir` and a slash-separated `rel` that starts with `objectpath.Dir + "/"` (the form `files.StorageRelPath` returns), as `RemoveObject` does. Private `innerPath(rel string) (string, error)` *(new)* holds `RemoveObject`'s prefix cut (empty remainder or no prefix → error) and returns `filepath.FromSlash(inner)`. | The package that owns the layout owns the folder name; one way to name an object; a caller cannot hand in a path outside the store (review M-5). | A private constant in `objectstore` — a third spelling. Absolute paths checked by string prefix — does not stop a symlink. |
| D-2 | Private `openObjectsRoot(projectDir string, create bool) (*os.Root, error)` *(new)*: `os.OpenRoot(projectDir)`; with `create`, `Mkdir(objectpath.Dir, 0o755)` on that root, ignoring `fs.ErrExist`; `Lstat(objectpath.Dir)` on that root: missing → `fs.ErrNotExist`; a symlink or not a directory → `ErrUnsafeObjects`; then `OpenRoot(objectpath.Dir)` on that root; close the project root; return the objects root. `ErrUnsafeObjects = errors.New("objectstore: objects folder is a symlink or not a folder")` *(new, exported)*. | The `Lstat` and the open happen inside one opened project root, so a swap of `objects` between them cannot escape (SEC-R2-1); `Root.OpenRoot` refuses a symlink that leaves the project root. | `os.Lstat` on the joined path, then `os.OpenRoot` — today's gap. |
| D-3 | Private `openRegular(root *os.Root, inner string) (*os.File, error)` *(new)*: `root.Lstat(inner)` (missing → `fs.ErrNotExist`); not a regular file (symlink, FIFO, directory, device) → `ErrNotRegular`; `root.OpenFile(inner, os.O_RDONLY, 0)`; `f.Stat()` not regular → close, `ErrNotRegular`. `ErrNotRegular = errors.New("objectstore: object is not a regular file")` *(new, exported)*. `OpenObject(projectDir, rel string) (*os.File, error)` *(new)* = `openObjectsRoot(projectDir, false)` (its `fs.ErrNotExist` and `ErrUnsafeObjects` returned as is) then `openRegular`. Private `matches(root *os.Root, inner, checksum string) (bool, error)` *(new)* replaces the exported `MatchesChecksum`: `openRegular`; `fs.ErrNotExist` or `ErrNotRegular` → `false, nil`; otherwise hash. The exported `MatchesChecksum` is deleted in STEP-3, when its last outside caller (`ingest.installObject`) goes. | One opener, so a check and a read refuse the same things; `Lstat` refuses a FIFO before any open can block. No exported function takes an unchecked path (review M-2). | `O_NONBLOCK` — not portable, and still opens a device. |
| D-4 | Private `createTempIn(root *os.Root, prefix string) (*os.File, string, error)` *(new)*: name `prefix` + 16 hex chars from `crypto/rand`, at the top of the objects root; `root.OpenFile(name, os.O_RDWR\|os.O_CREATE\|os.O_EXCL, 0o600)`; up to 10 tries on `fs.ErrExist`; returns the root-relative name, never `File.Name()`. Private `installTemp(root *os.Root, tmp, inner, checksum string) (wrote bool, err error)` *(new)* is the one install sequence: it always removes `tmp` through `root` on return; `matches(root, inner, checksum)` true → `false, nil`; `root.MkdirAll(filepath.Dir(inner), 0o755)`; `root.Remove(inner)` ignoring `fs.ErrNotExist` (a symlink at the path is replaced, never followed); `root.Rename(tmp, inner)`; on a rename failure, a match → `false, nil`, else the error; then `syncDirIn(root, filepath.Dir(inner))`, whose error is returned with `wrote` true. | One copy of the install steps for both writers (review M-1). Temps at the top of objects/ never sit in a shard folder, so the prune (D-7) and writes never meet. | Two near-copies in `WriteIfAbsent` and `Install` — today's state across two packages. `os.CreateTemp` — it has no `os.Root` form. |
| D-5 | `WriteIfAbsent(projectDir, rel, checksum string, data []byte) (wrote bool, err error)`: verify `data` against `checksum` (as today, before any folder is touched); `openObjectsRoot(projectDir, true)`; `matches` true → `false, nil`; `createTempIn(root, ".write-")`; write, sync, close (any failure removes the temp through the root); `installTemp`; a sync error with `wrote` true is ignored, as today. | Every write step resolves inside objects/ (SEC-1). | — |
| D-6 | `CreateTemp(projectDir, prefix string) (*Temp, error)` *(new)*: `openObjectsRoot(projectDir, true)`, `createTempIn(root, prefix)`. `type Temp struct { root *os.Root; f *os.File; name string; closed, done bool }` *(new)* with (`closed` tracks the file; `Close` on a closed file returns nil, and `Install` and `Discard` skip the file close when it is closed): `Write(p []byte) (int, error)` (the file's); `Close() error` (sync then close the file; the root stays open, and its doc says `Install` or `Discard` is what releases the handle); `Install(rel, checksum string) (wrote bool, err error)` (`innerPath(rel)`, close the file if still open, `installTemp(root, name, inner, checksum)`, close the root, `done = true`); `Discard()` (no-op when `done`; else close the file, remove `name` through the root, close the root, `done = true`). | A streamed temp is its own concept with its own handle: callers never hold a temp path, and `RemoveObject` stays about objects (review M-3). `Install` on a value does not collide with the package-level `Install(c)` vocabulary seeders. | Package-level `Install(projectDir, tmpRel, …)` — reuses a word that means "seed vocabulary" across `core/database`, and routes temps through `RemoveObject`. |
| D-7 | `RemoveObject(projectDir, rel string) error` keeps its signature. It uses `innerPath` and `openObjectsRoot(projectDir, false)` (`fs.ErrNotExist` → `nil`; `ErrUnsafeObjects` returned). After `root.Remove(inner)` succeeds or reports `fs.ErrNotExist`, it walks `dir := filepath.Dir(inner)` while `dir != "."`, calling `root.Remove(dir)`, then `dir = filepath.Dir(dir)`; it stops at the first error (a folder that is not empty, or gone) and returns `nil` for it. It never removes objects/ itself. Its doc gains: "It then removes the shard folders the removal left empty, up to but not including objects/." | Deletes stop leaving empty folders (the user's report). Every object write and purge runs inside `catalogsession.Do` (S10-53), so a prune cannot race a write's `MkdirAll`; `root.Remove` refuses a non-empty folder. | A sweep of every empty folder on each purge — reads the whole store. `os.RemoveAll` — could remove files. |
| D-8 | The directory sync becomes private `syncDirIn(root *os.Root, dir string) error` in `sync_unix.go` (`root.Open(dir)` + `Sync`) and `sync_other.go` (no-op). The exported `SyncDir` is deleted in STEP-3 with its last outside caller. | No exported helper takes an unchecked path. | — |
| D-9 | `core/ingest/file.go`: `streamSource(src *os.File, projectDir string) (t *objectstore.Temp, checksum string, byteSize int64, mediaType string, err error)` uses `objectstore.CreateTemp(projectDir, ".ingest-")` and an `io.MultiWriter(t, h, sniff)`; it calls `t.Close()` before returning and `t.Discard()` on any failure. `File` passes `c.Dir()` and adds `defer t.Discard()` right after `streamSource` succeeds; its early `os.Remove(tmpPath)` calls are deleted (the deferred `Discard` covers them). Both `installObject` calls become `t.Install(relPath, checksum)`. `installObject`, `objPath`, and `objectsDir` are deleted. `core/ingest/stage.go` `InsertTx`: `objectstore.WriteIfAbsent(s.c.Dir(), s.rel, s.checksum, s.png)`. | Callers name objects by `rel` and temps by handle. | — |
| D-10 | `core/derivatives/ensure.go`: `generate(projectDir, rel, wantChecksum string, opts raster.Options)` and `readSourceObject(projectDir, rel, wantChecksum string)`; `readSourceObject` opens with `objectstore.OpenObject` and maps `objectstore.ErrNotRegular` and `objectstore.ErrUnsafeObjects` to `ErrCorruptObject`, so the thumbnail handler skips them; other errors (including `fs.ErrNotExist`) pass through as today. `Ensure` calls `objectstore.WriteIfAbsent(c.Dir(), derivedRel, checksum, derivedJPEG)`; `objPath` and `derivedPath` are deleted. | The thumbnail handler (`api/ffi/handlers/thumbnails.go`) already skips `ErrCorruptObject`, so an unsafe source gets no thumbnail, as a tampered one does. | A new derivatives error code — the handler treats both the same. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `objectpath.Dir` *(new)* | const | `"objects/"` literal in `objectpath.Rel`, `catalog.go` `objectsDir` | The layout owner exports it; `Rel` and `objectstore` use it. |
| `objectstore.innerPath`, `openObjectsRoot`, `openRegular`, `matches`, `createTempIn`, `installTemp`, `syncDirIn` *(new, private)* | func | `RemoveObject`'s inline `Lstat` + `OpenRoot` + prefix cut; `WriteIfAbsent`; `ingest.installObject`; `SyncDir` | Extracted and merged: one root, one opener, one install sequence. |
| `objectstore.OpenObject`, `ErrNotRegular`, `ErrUnsafeObjects` *(new)* | func / error | `derivatives.readSourceObject`'s `os.Open` | The guarded open lives with the store; `readSourceObject` keeps its limit and checksum check. |
| `objectstore.CreateTemp`, `objectstore.Temp` *(new)* | func / type | `ingest.streamSource`'s `os.CreateTemp`, `ingest.installObject` | Moved, so the root guards them; a handle instead of a temp path. |
| `core/objectstore/objectstore_unix_test.go` *(new)* | test | `objectstore_test.go` | Only the FIFO case needs `//go:build unix` (`syscall.Mkfifo`); symlink cases stay in the portable file, as `TestRemoveObject`'s do. |

**Builds on.** S10-55's `RemoveObject` and its table test; `files.StorageRelPath`; the `catalogsession.Do` guarantee S10-53 documented.
**Deletes.** `objectstore.MatchesChecksum` and `SyncDir` (exported), `ingest.installObject`, the absolute-path form of `WriteIfAbsent`, the `objPath` / `derivedPath` / `objectsDir` joins in `ingest` and `derivatives`, `File`'s explicit temp removals.
**Next adds.** Any new object writer (S10-32's staged depiction already goes through `InsertTx`) calls `WriteIfAbsent(projectDir, rel, …)`; a streamed writer uses `CreateTemp`; a reader calls `OpenObject`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/objectpath` | D-1. |
| `core/objectstore` | D-1 to D-8. |
| `core/ingest` | D-9. |
| `core/derivatives` | D-10. |

```go
// core/objectstore (sketch)
var ErrNotRegular = errors.New("objectstore: object is not a regular file")
var ErrUnsafeObjects = errors.New("objectstore: objects folder is a symlink or not a folder")
func OpenObject(projectDir, rel string) (*os.File, error)
func WriteIfAbsent(projectDir, rel, checksum string, data []byte) (wrote bool, err error)
func CreateTemp(projectDir, prefix string) (*Temp, error)
func (t *Temp) Write(p []byte) (int, error)
func (t *Temp) Close() error
func (t *Temp) Install(rel, checksum string) (wrote bool, err error)
func (t *Temp) Discard()
func RemoveObject(projectDir, rel string) error
```

### FFI

None.

### macOS

None.

### Strings

None.

## 7. Steps

**STEP-1 — One guarded root and one guarded open**
Files: `core/objectpath/objectpath.go`, `core/objectstore/objectstore.go`, `core/objectstore/objectstore_test.go`, `core/objectstore/objectstore_unix_test.go` *(new)*.
Do: D-1, D-2, D-3 except deleting `MatchesChecksum` (it keeps its absolute-path form until STEP-3). `RemoveObject` moves onto `innerPath` and `openObjectsRoot` (no prune yet). Add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectpath/... ./core/objectstore/... ./core/ingest/...` passes; `TestRemoveObject` passes unedited.

**STEP-2 — Writes through the root**
Files: `core/objectstore/objectstore.go`, `core/objectstore/sync_unix.go`, `core/objectstore/sync_other.go`, `core/objectstore/objectstore_test.go`, `core/ingest/stage.go`, `core/derivatives/ensure.go`, `core/database/deleteimpact/purge_test.go`.
Do: D-4, D-5, D-8 (add `syncDirIn`; keep `SyncDir` for `installObject` until STEP-3), D-9's `InsertTx` line, D-10's `WriteIfAbsent` line. Update `TestWriteIfAbsent` to the new signature: one project folder from `t.TempDir()` and `rel` objects/ab/cd/obj.png; its mismatch case asserts no objects/ folder was created. Add T-2's rows to it.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/... ./core/ingest/... ./core/derivatives/... ./core/database/deleteimpact/...` passes.

**STEP-3 — Imports stream and install through the root**
Files: `core/objectstore/objectstore.go`, `core/objectstore/sync_unix.go`, `core/objectstore/sync_other.go`, `core/objectstore/objectstore_test.go`, `core/ingest/file.go`.
Do: D-6, D-9; delete `MatchesChecksum` (D-3) and `SyncDir` (D-8). Add T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/... ./core/ingest/...` passes with `file_test.go`, `file_unix_test.go`, and `stage_test.go` unedited; `grep -rn "SyncDir\|installObject\|MatchesChecksum" core api --include='*.go'` returns nothing.

**STEP-4 — Thumbnails read through the root**
Files: `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Do: D-10. Add T-7 as a row of `TestEnsureThumbnailRejectsUnsafeSources`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/derivatives/...` passes; the existing rows of `TestEnsureThumbnailRejectsUnsafeSources` are unedited.

**STEP-5 — Deletes prune empty shard folders**
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_test.go`, `core/database/artifacts/artifacts_test.go`.
Do: D-7. Add T-1's rows and T-8.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/objectstore/... ./core/database/...` passes.

**STEP-6 — Docs**
Files: `.agents/skills/review-pr/lenses/security.md`, `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-10/deployment-plan.md`, `docs/deployment-plan/spike-10/completed.md`, this plan.
Do:
- `security.md` trust table, Object store row: the Where cell becomes "`core/objectpath`, `core/objectstore` (every object read, write, install, and removal goes through an `os.Root` on objects/)".
- `source-layer-data-model.md` §6: at the end of the paragraph that names `files.PurgeUnreferenced` (around line 290), add "Object files are read, written, and removed only through `core/objectstore`, inside the project's objects/ folder; a symlinked objects/ or a symlink out of it is refused, and a removal also removes the shard folders it leaves empty."
- `deployment-plan.md`: the S10-56 **Plan:** line's status becomes "— Landed #N", and " — done, #N" is appended to its SLICE 8 line. `completed.md`: a `## S10-56 — Object store stays inside objects/` section above the current first section, in the shape of the `S10-53` section.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-56-object-store-root.md` passes; `grep -n "os.CreateTemp\|os.Rename\|os.MkdirAll" core/objectstore/objectstore.go core/objectstore/sync_unix.go core/objectstore/sync_other.go core/ingest/file.go core/derivatives/ensure.go` returns nothing (test files keep their fixtures).

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/objectstore/objectstore_test.go` `TestRemoveObject` (new rows) | only object in `objects/ab/cd` → object, `objects/ab/cd`, `objects/ab` gone, objects/ stays; two objects in `objects/ab/cd`, remove one → other and folders stay; objects in `objects/ab/cd` and `objects/ab/ef`, remove the first → `cd` gone, `ab` and the other stay | AC-2.1, AC-2.2, AC-2.3, D-7 |
| T-2 | Go | `objectstore_test.go` `TestWriteIfAbsent` (new rows) | `objects/ab` symlinked to an outside `t.TempDir()` → error, outside folder empty; `objects` symlinked outside → `ErrUnsafeObjects`, outside empty; object path a symlink to an outside file with the right bytes → `wrote` true, path is a regular file with the bytes, outside file unchanged; no objects/ yet → created, `wrote` true; no `.write-` file left at the top of objects/ after any row | AC-1.1, AC-1.2, AC-1.3, D-2, D-4, D-5 |
| T-4 | Go | `core/objectstore/objectstore_test.go` `TestOpenObject` *(new)*, table-driven, plus `core/objectstore/objectstore_unix_test.go` *(new, `//go:build unix`, `package objectstore_test`)* `TestOpenObjectFIFO` *(new)* | portable table: regular → opens, reads the bytes; symlink to an outside file → `ErrNotRegular`; missing → `fs.ErrNotExist`; a directory → `ErrNotRegular`; `objects` symlinked outside → `ErrUnsafeObjects`. Unix file: FIFO (`syscall.Mkfifo`) → `ErrNotRegular`, returns without blocking | AC-1.2, AC-1.4, D-3 |
| T-5 | Go | `objectstore_test.go` `TestTemp` *(new)*, table-driven | `CreateTemp(dir, ".ingest-")`, write, `Close`, `Install` at a fresh `rel` → `wrote` true, object holds the bytes, no `.ingest-` file left; same bytes into a second temp, `Install` → `wrote` false, temp gone; `Discard` without `Install` → temp gone, no object; `Discard` after `Install` → no-op; `objects/ab` symlinked outside, `Install` → error, temp gone, outside empty; `objects` symlinked outside → `CreateTemp` returns `ErrUnsafeObjects` | AC-1.2, AC-1.6, D-4, D-6 |
| T-7 | Go | `core/derivatives/thumbnail_test.go` `TestEnsureThumbnailRejectsUnsafeSources` (new row) | a source File from `installSourceFile`; its object replaced by a symlink to an outside file with the same bytes → `ErrCorruptObject`, no link rows (the table's existing assertion) | AC-1.5, D-10 |
| T-8 | Go | `core/database/artifacts/artifacts_test.go` `TestArtifactDeletePurgesDisk`, new `t.Run` "prunes empty shard folders" | inside the existing test, reusing its `objectPath` closure and the setup of its image-plus-thumbnail subtest (an ingested image File and its thumbnail from `derivatives.EnsureThumbnail`); assert each object is the only entry in its shard folder; delete with `runDelete` → both objects and their `objects/<hh>/<hh>` and `objects/<hh>` folders gone | AC-2.4, D-7 |
| T-9 | Go | existing tests named in AC-3.1 | pass with only the allowed edits | AC-3.1 |

There is no T-3 or T-6 (merged into T-4 and T-5 after review M-2 and M-3).

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestFile` and `TestStagedImage*` (STEP-2, STEP-3), thumbnail tests (STEP-2, STEP-4), `TestPurgeUnreferenced` and the observations sentinel test (STEP-5).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks); off macOS it is reported as failing on xcode-test only, with every other check passing
- [ ] STEP-3 and STEP-6 greps return nothing
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none in this PR)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI in this PR)
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain docs updated: security lens trust table, source-layer-data-model.md §6
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

In the app, on a test project: delete an Artifact whose image (and thumbnail) are the only objects in their shard folders, then confirm in Finder that neither the files nor their `objects/<hh>/<hh>` and `objects/<hh>` folders remain. Import a new image and confirm it opens and gets a thumbnail. Then run `CGO_ENABLED=1 go test -tags fts5 -run 'TestRemoveObject|TestWriteIfAbsent|TestOpenObject|TestTemp|TestEnsureThumbnailRejectsUnsafeSources|TestArtifactDeletePurgesDisk' -v ./core/objectstore/... ./core/derivatives/... ./core/database/artifacts/...` and paste the result in the PR.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A prune races a write recreating the same shard folder | Every object write and purge runs inside `catalogsession.Do` (S10-53); `root.Remove` refuses a non-empty folder; temps never sit in a shard folder (D-4). |
| `os.Root` behaves differently on macOS for `Rename` or `MkdirAll` | T-2 and T-5 run on the gate's macOS host. |
| #424 changes `core/objectstore` before it merges | The plan's base is #424's head; see Stop conditions. |

**Stop and report if:**

- `s10-53-stage-image` has moved past `39ee4860` with a change to `core/objectstore`, `core/ingest/file.go`, `core/ingest/stage.go`, or `core/derivatives/ensure.go`.
- `os.Root` in Go 1.27 lacks a method D-2, D-3, D-4, or D-7 names.
- An existing test in AC-3.1 needs an edit the AC does not allow.
- `TestEnsureThumbnailRejectsUnsafeSources`'s tampered-bytes row stops returning `ErrCorruptObject`.
- The observations sentinel test (a file at `objects/ab/cd/sentinel`) fails after STEP-5.

## 12. Open questions

None.

## 13. PR

**Title:** S10-56 — Object store stays inside objects/; deletes prune empty folders
**Base:** `s10-53-stage-image` until #424 merges, then `main`.
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-56-object-store-root.md`.
