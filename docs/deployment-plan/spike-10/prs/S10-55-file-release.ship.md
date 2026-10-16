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

## Result
Gate PASS · CI see PR #423 · open: none · follow-ups: CON-1, CON-2, MNT-1, MNT-2, SEC-O1, SEC-O2 · to user: —
