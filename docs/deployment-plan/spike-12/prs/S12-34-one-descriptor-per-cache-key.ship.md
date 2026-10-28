# Ship log — S12-34 One descriptor per cache key

**Plan:** [`S12-34-one-descriptor-per-cache-key.md`](S12-34-one-descriptor-per-cache-key.md) · **Branch:** `s12-34-one-descriptor-per-cache-key` · **PR:** #614
**State:** ready

## Round 0 — implement
Head `766c5fac`. Gate PASS (orchestrator, on the pushed head). Report: none (historical mentions of deleted names left in landed plans/audit as current-state records; STEP-6 paths clean).

## Round 1 — review
Reviewers: CON (Composer), MNT (Composer), UX (Composer), SEC+PRF (Composer). UX for `macos/App/**`; SEC+PRF for `macos/App/Features/**` session/models (performance). Skipped: none. (MNT/UX/SEC+PRF re-dispatched after Other Models usage abort; CON interrupted onto Composer.)

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 / SEC-1 | CON+SEC | blocker | `updatedSourceShell` writes patched Source onto every warmed workspace | Fix now → F-1 | verified: `CatalogLastUpdated.swift:95-101` had no `sourceId` match; old helper keyed one `sourceWorkspace`; test warmed only `s1` |
| MNT-1 | MNT | major | Skill still teaches “registry loader” in two entry points | Fix now → F-2 | verified: `add-workspace-place` lines 27 and 97 still said registry loader; body already said descriptor |
| MNT-2 | MNT | minor | T-2 does not pin `sourcesList` patch composition order | Fix now → F-3 | verified: test only checked first `follows`; D-5 requires `[.lastUpdated, .updatedSourceRow]` |
| UX | UX | — | no findings | — | Clean |

Nits (logged): `MNT-3 · CatalogQueryRegistry.swift:485-489 · hoist queryHandle guard before graphId loop` · `MNT-4 · CatalogLastUpdated.swift · file name still LastUpdated (S12-31)`

### Fix plan 1

**F-1 — Scope `updatedSourceShell` to the matching Source id** (from CON-1, SEC-1)
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-swift-test`.
Failing test first: `applyUpdatedSourceLeavesOtherCachedWorkspacesUntouched` — warm `s1` and `s2` workspaces; `apply(.updatedSource(…s1…))`; expect `s2` shell still `s2`.
Do: In `updatedSourceShell`, `guard case .sourceWorkspace(_, let sourceId) = key, sourceId == source.id else { return }` before mutating. Keep `.updatedSourceRow`’s row-id check as is.
Check: the new test passes; `applyUpdatedSourcePatchesListAndWorkspace` still passes; scripts/check-pr.sh PASS.

**F-2 — Finish `add-workspace-place` skill sweep** (from MNT-1)
Files: `.agents/skills/add-workspace-place/SKILL.md` (`.claude/skills` is the symlink).
Failing test first: none (docs).
Do: Retitle step 3 to `CatalogQueryKey` + descriptor arm. Change the “When this applies” Also-use cell from “registry loader in `CatalogQueryRegistry`” to the descriptor arm (including `patches` / `refresh` when needed), matching the checklist.
Check: `rg -n "registry loader" .agents/skills/add-workspace-place` prints nothing; scripts/check-pr.sh PASS.

**F-3 — Pin `sourcesList` patch order in T-2** (from MNT-2)
Files: `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`.
Failing test first: extend `lastUpdatedPatchesAreDeclaredOnTheFourLists` (or a sibling assertion) so `sourcesList` has exactly two patches, first `follows == .source`, second `follows == nil`.
Do: Assert count and order against D-5; do not change production patch order.
Check: the assertion fails if order is swapped; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `e8c2cbda`. Gate PASS (orchestrator). F-1, F-2, F-3 done.

## Round 2 — review
Reviewers: CON, MNT, SEC+PRF (had Fix-now blocker/major). Skipped: UX (Clean in round 1). Scope: `766c5fac...e8c2cbda`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | blocker | updatedSourceShell scope | Resolved | `sourceId` guard + `applyUpdatedSourceLeavesOtherCachedWorkspacesUntouched` |
| SEC-1 | SEC | blocker | same | Resolved | same evidence |
| MNT-1 | MNT | major | skill registry loader | Resolved | `rg "registry loader"` empty; step 3 retitled |
| MNT-2 | MNT | minor | sourcesList patch order | Resolved | T-2 asserts count 2 and follows order |

No new blockers.

## Result
Gate PASS · CI green · open: none · follow-ups: MNT-3 (hoist refresh queryHandle), MNT-4 (rename CatalogLastUpdated.swift → S12-31) · to user: — · CON miss: none
