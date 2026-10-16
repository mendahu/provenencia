# Ship log — S10-55 File release from the declared model

**Plan:** [`S10-55-file-release.md`](S10-55-file-release.md) · **Branch:** `s10-55-file-release` · **PR:** #423
**State:** ready

## Round 0 — implement
Head `6d5a2589`. Gate PASS (orchestrator run). Report deviations:
- D-4's `releaseOwnedRow` split into `snapshotOwnedRow` + `releaseOwnedSnap` (caller deletes the parent between them); recorded in the plan's Landed note.
- `gofmt -w` on `artifacts.go`, `citations.go` (base drift under go 1.27), formatting only, `ee1ebd95`.
- `.agents/skills/add-catalog-migration/SKILL.md` reworded (STEP-7 grep required it).
- STEP-7 grep still hits "release entry" in S10-32's plan D-2; left for S10-32.
- Implementer's shell popped the user's `stash@{0}` by mistake; it restored the file and the stash is intact (orchestrator checked `git stash list`). No branch change to `Localizable.xcstrings`.
- Manual check not run (implementer); T-9 covers it.

## Round 1 — review
Lenses: all five. CON, MNT minors/nits; UX, SEC, PERF clean.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | minor | `purge_test.go` subtest "commit purges and rollback does not" tests rollback only | Follow-up | verified lines 109–124; commit path covered by `TestArtifactDeletePurgesDisk`; no fix plan touches the file |
| CON-2 | CON | minor | AC-1.3 case doesn't assert the `A → T` derivative link survives | Follow-up | verified `release_test.go:319-323`; current code keeps the link (early return); no fix plan |
| CON-3 | CON | nit | Landed note omits that `lookupParentColumn` was folded into `snapshotOwnedRow` | Logged | behavior matches D-4 |
| MNT-1 | MNT | minor | Release → `files.PurgeAfterCommit` pairing is hand-kept in six deletes, untested | Follow-up | D-1 rejected purge in `deleteimpact`; `Tx.Catalog()` reopens it; suggest folding into S10-53 before S10-32 adds a real File release |
| MNT-2 | MNT | minor | `artifacts.Delete`'s `c` parameter is now unused | Follow-up | D-8 keeps it in this PR |
| MNT-3 | MNT | nit | PK-missing check duplicated in `release.go` | Logged | — |
| MNT-4 | MNT | nit | `"files"` / `"source_file_id"` literals repeated | Logged | — |
| SEC-O1 | SEC (outside) | minor | `PurgeUnreferenced` follows a symlinked `objects/hh` dir | Follow-up | existed before this PR; object-store hardening |
| SEC-O2 | SEC (outside) | minor | Delete vs in-flight ingest of same bytes race | Follow-up | existed before this PR; S10-53 owns it |

No Fix-now findings → no Fix plan 1, no round 2.

## Round 1 — user decisions (2026-10-16)
After the ready report, the user re-decided the follow-ups:

| ID | New decision | Note |
| --- | --- | --- |
| MNT-1 | Requirement for S10-53, in its docs → F-5 | |
| MNT-2 | Fix now → F-1 | Overrides plan D-8 ("`artifacts.Delete` keeps its `c` parameter"), by the user |
| CON-1 | Fix now → F-2 | |
| CON-2 | Fix now → F-3 | |
| SEC-O1 | Fix now → F-4 | |
| SEC-O2 | Requirement for S10-53, in its docs → F-5 | |

### Fix plan 1

**F-1 — Drop the unused Catalog from `artifacts.Delete`** (from MNT-2; user overrides D-8)
Files: `core/database/artifacts/artifacts.go`, `api/ffi/handlers/artifacts.go`, any test caller (`grep -rn "artifacts.Delete(" core api`).
Do: remove the `c *database.Catalog` parameter; update every caller. In the S10-55 plan's Landed note (`docs/deployment-plan/spike-10/prs/S10-55-file-release.md`), add: "`artifacts.Delete` no longer takes the Catalog (user, 2026-10-16, overriding D-8): `Tx.Catalog()` serves the purge." Also add one clause there that `lookupParentColumn` was folded into `snapshotOwnedRow` (CON-3).
Check: `go build ./...`; scripts/check-pr.sh PASS.

**F-2 — Test both halves of `PurgeAfterCommit`** (from CON-1)
Files: `core/database/files/purge_test.go`.
Failing test first: none (coverage gap). In subtest "commit purges and rollback does not", after the rolled-back call keeps the object, add a second `writes.Call` whose function calls `files.PurgeAfterCommit(tx, []files.Object{obj(sumGone)})` and returns nil; assert the object no longer exists.
Check: `go test -tags fts5 ./core/database/files/` passes.

**F-3 — AC-1.3 keeps the derivative link** (from CON-2)
Files: `core/database/deleteimpact/release_test.go` (`TestReleaseReportsReleasedObjects`, `extra` branch).
Do: also assert `SELECT COUNT(*) FROM file_derivatives` == 1.
Check: the test passes.

**F-4 — Purge cannot leave the objects folder through a symlink** (from SEC-O1)
Files: `core/objectstore/objectstore.go`, `core/objectstore/objectstore_test.go` (new if absent), `core/database/files/purge.go`.
Failing test first: `TestRemoveObject` in `core/objectstore` (`package objectstore_test`), table-driven on `t.TempDir()` project folders:
  - regular object at `objects/ab/cd/<sum>.jpg` → removed, returns nil;
  - missing object → returns nil (no error);
  - `objects/ab` is a symlink to a directory outside the project that holds `cd/<sum>.jpg` → returns an error and the outside file still exists;
  - `objects` itself is a symlink to an outside directory → error, outside file still exists;
  - the object path is a symlink to an outside file → the outside file still exists (removing the link itself is fine);
  - a rel that does not start with `objects/` → error.
Do: add `objectstore.RemoveObject(projectDir, rel string) error`. It requires `rel` to start with `objects/`, opens `os.OpenRoot(filepath.Join(projectDir, "objects"))` (refuse if `objects` itself is a symlink: `os.Lstat` it first and return an error when `ModeSymlink` is set), and calls `root.Remove` on the rest of the path (slash-separated → `filepath.FromSlash`). A missing file (`errors.Is(err, fs.ErrNotExist)`) returns nil. Doc comment: "RemoveObject removes one object file under projectDir/objects without following a symlink out of that folder." Update the package doc to "Package objectstore verifies and removes content-addressed object files on disk." `files.PurgeUnreferenced` calls `objectstore.RemoveObject(c.Dir(), rel)` instead of `os.Remove(filepath.Join(...))` and ignores its error as today (best-effort). Do not touch `core/ingest`'s removes; S10-53 moves them onto the purge (F-5).
Check: `TestRemoveObject` red before, green after; `TestPurgeUnreferenced` and `TestArtifactDeletePurgesDisk` pass unedited; scripts/check-pr.sh PASS.

**F-5 — Carry two requirements into S10-53's docs** (from MNT-1, SEC-O2)
Files: `docs/deployment-plan/spike-10/prs/S10-53-stage-image.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (S10-53 scope entry).
Do: in the S10-53 plan, directly under the header block (before "Implementer — read this first"), add a section:

```markdown
### Requirements added after planning (user, 2026-10-16, from S10-55's review)

This plan must go back through `plan-pr` and `review-plan` to fold these in before it is Planned again.

1. **Releases schedule their own purge** (S10-55 MNT-1). Today each releasing delete (`artifacts`, `observations`, `subjects`, `sources`, `citations`, `metadatafields`) pairs `deleteimpact.ReleaseFacets` / `ReleaseSnapshot` with its own `files.PurgeAfterCommit` call, and nothing checks the pairing. `database.Tx` now carries its catalog (`Tx.Catalog()`), which removes the reason S10-55 D-1 kept the purge out of `deleteimpact`. S10-53 makes the release entry points take the `*database.Tx` and schedule the purge themselves, so callers no longer pass `Released.Objects`, and `ReleaseReplaced` (S10-32's update path) gets the same.
2. **A delete cannot race an import of the same bytes** (S10-55 SEC-O2). A delete's after-commit purge can remove an object that an `ingest.File` or `StageImage` running at the same time has just written or reused but not yet committed a row for. When ingest's rollback and `Discard` move onto `files.PurgeUnreferenced`, S10-53 closes this window (for example, by doing the object write and the purge's row check under one lock, or by re-verifying the object inside the write before it commits) and pins it with a test.
```

Set the plan's **Status** line to `Blocked — on S10-55; needs re-plan for the requirements below`. In `deployment-plan.md`, the S10-53 **Plan:** line gets the same status text, and its **In** cell ends with: " Also, from S10-55's review (user, 2026-10-16): release entry points schedule their own purge through `Tx.Catalog()`, and a delete cannot race an import of the same bytes."
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-53-stage-image.md docs/deployment-plan/spike-10/prs/S10-55-file-release.md` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `d5dd1f87`. Gate PASS (orchestrator run). F-1 to F-5 done. Deviation: F-2's committing write returns one `subject_position` create change (a table with no audit scope) because `writes.Run` rolls back an empty change list. Landed note and `completed.md` updated for F-4.

## Round 2 — review
Lenses: CON, MNT, SEC (the lenses with fixes). Scope `6d5a2589...d5dd1f87`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| — | CON | — | F-1 to F-4 and CON-3 resolved; F-5 done | — | tests run green at `d5dd1f87` |
| — | MNT | — | MNT-2 resolved; MNT-1 carried into S10-53 | — | — |
| MNT-R2-1 | MNT | minor (not raised) | `"objects"` folder name now spelled in four places | Logged | tidy when S10-53 adds the object write to `core/objectstore` |
| — | SEC | — | SEC-O1 resolved by `RemoveObject` | — | `TestRemoveObject` six cases |
| SEC-R2-1 | SEC | minor (not raised) | Gap between `Lstat(objects)` and `OpenRoot` | Follow-up | open root at the project folder, then `Root.OpenRoot("objects")` |

No new blockers → no Fix plan 2.

## Result
Gate PASS · CI see PR #423 · open: none · follow-ups: SEC-R2-1; MNT-1 and SEC-O2 are S10-53 requirements · logged: MNT-R2-1, MNT-3, MNT-4 · to user: — (user decisions recorded above)
