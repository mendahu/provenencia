# Ship log — S10-56 Object store stays inside objects/; deletes prune empty folders

**Plan:** [`S10-56-object-store-root.md`](S10-56-object-store-root.md) · **Branch:** `s10-56-object-store-root` · **PR:** #425
**State:** fixing 1

## Round 0 — implement
Base: branch `s10-56-object-store-root` cut from `s10-53-stage-image` @ `39ee4860` (#424, open); plan committed on it. PR targets `s10-53-stage-image` until #424 merges.
Head `6a7ab5db`. Gate PASS (orchestrator run). Report deviations:
- The prune cases are a separate `TestRemoveObjectPrunesEmptyShards` (the existing `TestRemoveObject` table's shape cannot assert folder removal); same cases as T-1.
- Plan Status written as "Landed in [#425](…) (2026-10-17)", the form `check-plan-docs.py` accepts.
- In-app manual check not run.

## Round 1 — review
Lenses: all five. PERF clean; CON, SEC one major each; MNT, UX minors.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| SEC-1 | SEC | major | A source under a symlinked shard (`objects/ab` → elsewhere) makes `root.Lstat` fail with a path-escape error that is not `ErrNotRegular`, so `readSourceObject` returns it raw and a Source's whole Artifact list fails to load | Fix now → F-1 | trace: Go `os` root returns `errPathEscapes`; `api/ffi/handlers/thumbnails.go` skips only `ErrCorruptObject` etc.; `artifacts.go` list loop aborts. F-1's failing test verifies |
| CON-1 | CON | major | No Landed note; PR body says "Deviations: None" | Fix now → F-5 | verified: plan has no `## Landed`; ship log lists the separate prune test and the skipped manual check |
| CON-O1 (from CON outside) | UX | major | Finder's `.DS_Store` keeps a browsed shard folder, so the cleanup misses exactly the folders the user looks at | To user → **Fix now → F-2** (user, 2026-10-17: remove a lone `.DS_Store`) | extends D-7; decided by the user |
| UX-1 | UX | minor | A symlinked objects/ now fails an import with "Something went wrong. Please try again." | To user → **Fix now → F-3** (user, 2026-10-17: keep the refusal, add a clear message in this PR; overrides the plan's "no Swift, no strings") | `ErrUnsafeObjects` is a plain error → `internal.unknown` |
| MNT-1 | MNT | minor | `WriteIfAbsent` removes its temp in three places | Fix now → F-4 (rides along) | same file as F-1 |
| MNT-2 | MNT | minor | Package doc does not mention reads, temps, or the objects/ guarantee | Fix now → F-4 (rides along) | same file |
| MNT-3 | MNT | minor | Test setup copied (prune test, `TestOpenObject`, artifacts subtest) | Follow-up | test-only; more than a few lines |
| CON-2 | CON | minor | `TestTemp`, `TestWriteIfAbsent` not table-driven | Follow-up | coverage complete; style rule only |

### Fix plan 1

**F-1 — A symlinked shard folder is refused like a symlinked object** (from SEC-1)
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_test.go`, `core/derivatives/thumbnail_test.go`.
Failing test first: `TestOpenObject` new row "objects/ab symlinked outside" → want `ErrNotRegular` (today: a path-escape error); `TestEnsureThumbnailRejectsUnsafeSources` new row "source's shard folder is a symlink" (move `objects/<hh>` of an installed source to a folder elsewhere in the temp dir and symlink it back) → want `ErrCorruptObject`.
Do: in `openRegular`, before the object's own `root.Lstat`, walk `dir := filepath.Dir(inner)` up to `"."` and `root.Lstat` each; missing → `fs.ErrNotExist`; anything that is not a real directory (a symlink included) → `ErrNotRegular`. `matches` and `readSourceObject` need no change (they already map `ErrNotRegular`).
Check: both rows red before, green after; `go test -tags fts5 ./core/objectstore/... ./core/derivatives/...` passes.

**F-2 — The cleanup treats a folder holding only `.DS_Store` as empty** (user decision)
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_test.go`.
Failing test first: `TestRemoveObjectPrunesEmptyShards` new row: `objects/ab/cd` holds the object and a `.DS_Store`, `objects/ab` also holds a `.DS_Store` → after `RemoveObject`, `objects/ab/cd` and `objects/ab` are gone, objects/ stays. And a row where `objects/ab/cd` holds the object, a `.DS_Store`, and another file → `cd` stays with both files.
Do: in `RemoveObject`'s prune walk, when `root.Remove(dir)` fails, read the folder's entries through the root (`root.Open(dir)` + `ReadDir(2)`); if the only entry is a regular file named `.DS_Store`, `root.Remove(filepath.Join(dir, ".DS_Store"))` and retry `root.Remove(dir)` once; otherwise stop, as today. Never touch objects/ itself. Update `RemoveObject`'s doc: "A folder whose only entry is a Finder `.DS_Store` counts as empty."
Check: both rows red/green as stated; `go test -tags fts5 ./core/objectstore/... ./core/database/...` passes.

**F-3 — A symlinked objects/ folder gets its own message on import** (from UX-1, user decision)
Files: `core/apperr/apperr.go`, `core/ingest/file.go`, `core/ingest/stage.go`, `core/ingest/stage_test.go` or a new `core/ingest/unsafe_objects_test.go`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, the Swift test that covers `L10n.Errors.ingestCallout(code:)` if one exists (else add one under `macos/ProvenenciaTests`). Skills: `add-localized-string`, `add-swift-test`.
Failing test first: Go — `ingest.File` into a project whose objects/ is a symlink to an outside folder → `errors.Is(err, ingest.ErrUnsafeObjects)` and `apperr.From(err).Code == "ingest.unsafe_objects"`; the same for `StagedImage.InsertTx`. Swift — `L10n.Errors.ingestCallout(code: "ingest.unsafe_objects")` is non-nil with the new title.
Do: `apperr.CodeIngestUnsafeObjects = "ingest.unsafe_objects"` after `CodeIngestUnsupportedImage`; `ingest.ErrUnsafeObjects = apperr.New(apperr.CodeIngestUnsafeObjects, apperr.KindUser)`. In `ingest.File` (the `CreateTemp`/`streamSource` error and the `Install` error) and in `InsertTx` (the `WriteIfAbsent` error): `errors.Is(err, objectstore.ErrUnsafeObjects)` → return `ErrUnsafeObjects`. `objectstore` stays free of `apperr`. Swift: `L10n.Errors.ingestUnsafeObjectsTitle` = "This project’s objects folder can’t be used", `ingestUnsafeObjectsHelp` = "Its objects folder is a link to another location. Provenencia stores files only inside the project folder. Replace the link with a real folder, then try again."; keys `error.ingest.unsafe_objects.title` / `.help` in `Localizable.xcstrings` (en, translated, manual); a `case "ingest.unsafe_objects":` in `ingestCallout(code:)` returning an `IngestCallout` with them, shaped like the `ingest.invalid` case.
Check: Go and Swift tests red/green; scripts/check-pr.sh PASS (includes localizable and xcode-test).

**F-4 — One deferred temp removal; a package doc that says what the package does** (from MNT-1, MNT-2)
Files: `core/objectstore/objectstore.go`.
Do: in `WriteIfAbsent`, right after `createTempIn` succeeds, `defer func() { _ = root.Remove(name) }()`; delete the three inline `root.Remove(name)` calls (keep the `tmp.Close()` calls). Package doc: "Package objectstore reads, writes, streams, and removes content-addressed object files, always through an os.Root on the project's objects/ folder."
Check: `go test -tags fts5 ./core/objectstore/...` passes unedited.

**F-5 — Landed note** (from CON-1)
Files: `docs/deployment-plan/spike-10/prs/S10-56-object-store-root.md`, `docs/deployment-plan/spike-10/completed.md`.
Do: add `## Landed` under the header (shape of S10-53's) listing: the prune cases are a separate `TestRemoveObjectPrunesEmptyShards` (the `TestRemoveObject` table cannot assert folder removal); a lone `.DS_Store` counts as empty (user, 2026-10-17, F-2); `ingest.unsafe_objects` with Swift copy (user, 2026-10-17, F-3, overriding "no Swift, no strings"); a symlinked shard folder is `ErrNotRegular` (F-1); the in-app manual check is left to the user. Add one sentence about the new message and `.DS_Store` to the S10-56 `completed.md` section.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-56-object-store-root.md` passes.
