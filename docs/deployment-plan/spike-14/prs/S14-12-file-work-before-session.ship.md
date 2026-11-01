# Ship log — S14-12 File work before the catalog session

**Plan:** [`S14-12-file-work-before-session.md`](S14-12-file-work-before-session.md) · **Branch:** `s14-12-file-work-before-session` · **PR:** #664
**State:** ready

## Round 0 — implement
Base `0c166624`. Head `8712d463`. Gate PASS (orchestrator run; implementer saw one `PVSegmentedControlTests.disabledItemDoesNotSelect` failure that passed on rerun, untouched code).
Report deviations:
1. **D-12 not as written (To user).** `sources.CoverArtifactFile` keeps `sqlArtifactForCover` instead of `artifacts.Get`: `core/database/artifacts/artifacts_test.go` is in-package and imports `sources` (25 uses), so `sources → artifacts` is a test import cycle. Verified. Cover rule is still in one place; `sources` no longer imports `derivatives`.
2. `objectstore.CheckObjects` added; `Staged` calls it before `Install` so a symlink-swapped objects folder still returns `ErrUnsafeObjects`.
3. `objectFiles` test helper skips `.ingest-`/`.stage-` temps; new `stageTemps` helper.
4. STEP-7 shares STEP-6's commit.
5. T-10 is a row in `TestEnsureFileThumbnail`.
6. `CommitFile` / `InsertTx` do not purge; cleanup is `Settle`/`Discard` only (D-6).
7. Manual check §10 not run.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, or error codes; `project.pbxproj` is the version bump only).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | `sources` keeps its own Artifact query, now exported as `CoverArtifactFile` | Fix now → F-2 | verified: `artifacts_test.go` uses one unexported name (`getByRef`, line 233); the repo's `export_test.go` pattern (`evidencegraphs`, `identityclaims`) breaks the test cycle. Restores D-12 as written; resolves Round 0 deviation 1, so it no longer goes to the user |
| SEC-x (CON) | SEC→CON | major | `Job.Stage` reads the source object after `Plan`'s session; an Artifact delete in between purges it, `readSourceObject` returns raw `fs.ErrNotExist`, `Refused` misses it, and the whole `GetSourceWorkspace` / `SetSourceCover` fails | Fix now → F-1 | verified: `core/derivatives/ensure.go:251-258` returns the `OpenObject` error unmapped except `ErrNotRegular`/`ErrUnsafeObjects`; `Refused` (203) covers only the three sentinels. A new race this PR opens (the read used to be under the session). Re-rated from an outside-lens note to major: one deleted Artifact fails the whole Source page |
| MNT-2 | MNT | major | `SetSourceCover` re-derives the cover-mode rule (`TrimSpace(mode) == CoverModeArtifact && len(id) == 16`) | To user → Fix now → F-7 (user 2026-11-01: "Predicate in sources") | verified at `api/ffi/handlers/sources.go:329`; the same parse is in `NormalizeCover` (154-160) and `SetCover`. D-12 spells out this handler check, so the fix contradicts a Decision |
| SEC-1 | SEC | minor | `CheckObjects` checks the current `objects/` path, not the folder the temp's root holds; a rename-and-recreate passes | Fix now → F-3 | verified: `stage.go:170-185` → `Temp.Install` uses `t.root` from `CreateTemp`; nothing compares it with a fresh open. Code this PR adds |
| MNT-3 | MNT | minor | The swap check sits in `Staged.install`, not `Temp.Install` | Fix now → F-3 (with SEC-1) | same place as SEC-1; one fix: `Temp` keeps `projectDir`, `Install` re-opens and compares with `os.SameFile` |
| PRF (MNT outside lens) | PRF | minor | `thumbnailPass.run` opens a session for a batch whose jobs were all refused at stage | Fix now → F-4 | verified at `thumbnail_pass.go:107-139`; code this PR adds |
| MNT-4 | MNT | minor | Project-folder-for-session-tests helper written four times | Fix now → F-5 | same-pass cleanup: this PR adds two copies |
| MNT-5 | MNT | minor | `EnsureThumbnail` became test-only but its doc and the skill's "Do not" do not say so | Fix now → F-6 | `rg 'EnsureThumbnail\(' -g '!*_test.go'` hits only `ensure.go` |

Nits (logged): `CON-1 · core/ingest/file.go:1-8 · rewrap package doc` · `CON-2 · core/database/sources/sources.go:56 · sqlArtifactForCover moved in its const block (goes with F-2)` · `CON-3 · sources.go CoverArtifactFile comment rewrap (goes with F-2)` · `MNT-6 · core/ingest/file.go:1-6 · rewrap` · `MNT-7 · core/database/observations/depiction.go:24 · rewrap` · `MNT-8 · sources.go:474-475 · rewrap (goes with F-2)` · `MNT-9 · thumbnail_pass.go:79 · fillLink repeats derivedRelPath's 16-byte check` · `MNT-10 · thumbnails.go:50 · "single-link" tail`

User on MNT-2 (2026-11-01): "Predicate in sources". This supersedes D-12's handler-side mode check.

### Fix plan 1

**F-1 — A source object purged between Plan and Stage leaves only that thumbnail empty** (from SEC outside-lens note)
Files: `core/derivatives/ensure.go`, `core/derivatives/thumbnail_test.go`.
Failing test first: a new row in `TestCommitAfterChange` (or a sibling `TestStageAfterPurge`): `Plan` returns a Job inside a session, the source object is removed from `objects/` (as `files.PurgeAfterCommit` would after an Artifact delete), then `Job.Stage` → today returns a raw `fs.ErrNotExist` that `derivatives.Refused` is false for. Add a handler-level case too if cheap: `thumbnailPass` with a job whose object is gone fills `("", true)` and the call succeeds.
Do: in `readSourceObject`, map `errors.Is(err, fs.ErrNotExist)` from `objectstore.OpenObject` to `ErrInvalid` (the File is gone, as `Commit` reports a gone row), so `Refused` covers it. No new predicate.
Check: the test passes; `Refused(err)` is true for it; scripts/check-pr.sh PASS.

**F-2 — `CoverArtifactFile` reads the Artifact through `artifacts.Get` (D-12 as written)** (from MNT-1; resolves Round 0 deviation 1; folds in nits CON-2, CON-3, MNT-8)
Files: `core/database/artifacts/artifacts_test.go`, `core/database/artifacts/export_test.go` (new), `core/database/sources/sources.go`, the plan's Landed note.
Do: make `artifacts_test.go` `package artifacts_test` and qualify its identifiers. Add `export_test.go` (`package artifacts`) with `var GetByRef = getByRef`, following `core/database/evidencegraphs/export_test.go`. `CoverArtifactFile` calls `artifacts.Get` (`sql.ErrNoRows` → `ErrInvalid`). Delete `sqlArtifactForCover` and the import-cycle sentence. Remove deviation (1) from the plan's Landed note.
Check: `rg sqlArtifactForCover core/` → no hit; `go list -deps ./core/database/artifacts | grep database/sources` → no hit (production); tests pass; scripts/check-pr.sh PASS.

**F-7 — The cover-mode rule lives once in `sources`** (from MNT-2; user decision; supersedes D-12's handler check)
Files: `core/database/sources/sources.go`, `core/database/sources/sources_test.go`, `api/ffi/handlers/sources.go`, the plan's Landed note.
Failing test first: `TestCoverArtifactID` (table): `"artifact"` + 16-byte id → id, true; `" artifact "` + 16-byte id → id, true; `"artifact"` + empty id → false; `"type_icon"` → false; unknown mode → false.
Do: add pure `sources.CoverArtifactID(mode string, primaryArtifactID []byte) ([]byte, bool)`, true only when the trimmed mode is `CoverModeArtifact` and the id is 16 bytes. `NormalizeCover` and `SetCover` share one unexported `normalizeCoverMode(mode) string` (the `strings.TrimSpace`) with it. The handler replaces `strings.TrimSpace(req.GetCoverMode()) == sources.CoverModeArtifact && len(primaryID) == 16` with `if fileArtifactID, ok := sources.CoverArtifactID(req.GetCoverMode(), primaryID); ok`. No mode string compare is left in `api/`. Note the supersession in the Landed note.
Check: `rg 'CoverModeArtifact' api/` → only the proto/request pass-through if any, no comparison; `TestSetCoverMakesThumbnailFirst` unchanged and passing; scripts/check-pr.sh PASS.

**F-3 — `Temp.Install` refuses when `objects/` is not the folder the temp was made in** (from SEC-1, MNT-3)
Files: `core/objectstore/objectstore.go`, `core/ingest/stage.go`, `core/ingest/unsafe_objects_test.go`.
Failing test first: a case in `unsafe_objects_test.go`: stage, rename `objects/` aside and create a new real `objects/` folder, then `InsertTx` / `CommitFile` → want `objectstore.ErrUnsafeObjects`; today it installs into the moved folder.
Do: `CreateTemp` keeps `projectDir` on the `Temp`. `Temp.Install` first opens a fresh root with `openObjectsRoot(projectDir, false)` (that refuses a symlink, as `CheckObjects` does), `Stat(".")` on both roots, and returns `ErrUnsafeObjects` unless `os.SameFile`. Delete exported `CheckObjects` and its call in `Staged.install`.
Check: the new case and `TestStagedInsertTxRefusesSymlinkedObjects` pass; `rg CheckObjects core api` → no hit; scripts/check-pr.sh PASS.

**F-4 — No session for a batch with nothing staged** (from MNT outside-lens PRF note)
Files: `api/ffi/handlers/thumbnail_pass.go`, `api/ffi/handlers/thumbnails_test.go`.
Do: in `thumbnailPass.run`, skip `withStaged` for a batch whose stage left no staged job, unless it is the last batch and `then` is not nil (then must still run in a session; when no batch commits, it runs in its own `withProjectCatalog`, as today). Keep the "then runs exactly once" contract.
Check: a refused-only page (late-failing JPEG case in `TestWorkspaceThumbnailPass`) still returns empty paths; `TestSetCoverMakesThumbnailFirst` passes; scripts/check-pr.sh PASS.

**F-5 — One `catalogsessiontest.Project(t)` helper** (from MNT-4; same-pass cleanup)
Files: `core/catalogsession/catalogsessiontest/catalogsessiontest.go`, `core/derivatives/thumbnail_test.go` (`sessionProject`), `core/ingest/stage_test.go` (`stageProjectDir`), `core/catalogsession/session_test.go` (`TestHold`), `core/catalogsession/warm_test.go` (`newProject`).
Do: add `Project(t) string` beside `Hold`: create a catalog in `t.TempDir()`, close it, register `catalogsession.CloseAll` cleanup the way `newProject` does, return the folder. Move all four onto it and delete the copies. If `warm_test.go`'s `newProject` does something extra (e.g. onboarding), keep that as a wrapper over `Project`.
Check: `grep -rnE 'database\.Create\(t\.TempDir\(\)' core/derivatives core/ingest core/catalogsession` prints only `catalogsessiontest.go`; tests pass; scripts/check-pr.sh PASS.

**F-6 — `EnsureThumbnail` says it is for tests** (from MNT-5)
Files: `core/derivatives/ensure.go`, `.agents/skills/use-catalog-session/SKILL.md`.
Do: give `EnsureThumbnail`'s doc the same test-only sentence as `Ensure`; add `derivatives.EnsureThumbnail` to the skill's "Do not" line beside `derivatives.Ensure`.
Check: `scripts/check-guidance-names.py` OK; scripts/check-pr.sh PASS.

Order: F-1, F-2, F-7, F-3, F-4, F-5, F-6. Nits stay logged; the implementer need not touch them beyond what F-2 removes.

## Round 1 — fix
Head `a463d7d4`. Gate PASS (orchestrator run). F-1, F-2, F-7, F-3, F-4, F-5, F-6 done; every failing-test-first ran red, then green.
Implementer deviations: F-5 keeps `warm_test.go` `newProject` (it runs `onboarding.Complete`; the warm tests need it), so `Project` replaces the other three. F-2 also maps `artifacts.ErrInvalid` (non-16-byte id) to `sources.ErrInvalid`. F-7 also moved the two cover-mode compares in `api/ffi/handlers/thumbnails.go` onto `CoverArtifactID`. F-1 has no handler-level case ("if cheap"). Plan Landed note rewritten (items 1–9).

Minor fixes read by the orchestrator (no round-2 reviewer for SEC+PRF):
- SEC-1 / MNT-3 (F-3): **Resolved.** `Temp` keeps `projectDir`; `Install` → `requireSameObjects` (fresh `openObjectsRoot`, `os.SameFile` on both roots), discards the temp on refusal; `CheckObjects` gone.
- PRF note (F-4): **Resolved.** `run` stages each batch first and skips `withStaged` when nothing staged, unless last batch with `then`; normalizes with `catalogsession.Key` once.
- MNT-4 (F-5): **Resolved** (with the `newProject` exception above).
- MNT-5 (F-6): **Resolved.**

## Round 2 — review
Reviewers: MNT (MNT-1, MNT-2), CON (the routed stage-after-purge major). Skipped: SEC+PRF (minors only, read above), UX (not routed in round 1).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) | Resolved | `CoverArtifactFile` → `artifacts.Get`; `sqlArtifactForCover` gone; no production cycle |
| MNT-2 | MNT | major | (round 1) | Resolved | `sources.CoverArtifactID`; `rg CoverModeArtifact api/` → none |
| SEC-x (CON) | CON | major | (round 1) | Resolved | `readSourceObject` maps `fs.ErrNotExist` → `ErrInvalid`; `TestStageAfterPurge` |

No new blockers from either reviewer.

Nits (logged): `CON-4 · core/derivatives/thumbnail_test.go:560 · orphaned sessionProject doc comment above tinyPNG`

## Result
Gate PASS · CI pending · open: none · follow-ups: none · to user: MNT-2 (answered: "Predicate in sources") · CON miss: none
