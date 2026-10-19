# Ship log — S12-03 Evidence graphs in the catalog

**Plan:** [`S12-03-evidence-graphs-catalog.md`](S12-03-evidence-graphs-catalog.md) · **Branch:** `s12-03-evidence-graphs-catalog` · **PR:** #498
**State:** ready

Worktree: `/Users/jakerobins/Repos/provenencia-s12-03`, branched from `origin/main` @ `84a9b318`.

## Round 0 — implement
Head `686c8e29`. Gate PASS (orchestrator run). Report: no deviations from Decisions; test details the plan left open (evidencegraphs tests strip the create-time links and link explicitly; T-15/T-16 use seeded `birth_certificate`; T-19 builds its second Source in the transaction). AC-1.1 migration backfill and §10 manual check not run by the implementer.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`core/apperr`, new error codes), SEC+PRF (security: `migrations/000055.sql`, `core/catalogsession`; performance: `core/database/**`). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 (+ SEC outside-lens note to CON) | PRF | major | `evidencegraphs.EnsureRefs` backfills one ref per Source with no collision retry and one autocommit per row; first open after upgrade can fail | To user → user 2026-10-19: override D-15, fix here → F-1 | verified: `evidencegraphs.go:361-` loop has no `IsUniqueConflict` branch and no transaction; 32^5 token space gives ~1.5% collision chance at 1,000 Sources, ~31% at 5,000. Contradicts D-15 "no retry (as `users`)" |
| MNT-3 | MNT | major | `add-catalog-ref` skill step 3 does not name `ReconcileOnOpen` as the home for open-time backfills | Fix now → F-2 | verified: `.agents/skills/add-catalog-ref/SKILL.md:74,88` name only `users.EnsureRefs`; `docs/catalog-refs.md:112` already says `ReconcileOnOpen` |
| MNT-1 | MNT | major | `EvidenceGraphs` is the third catalog-wide bool on `effects.Set`/`Effect` | Follow-up | plan Out line and working rule "The graph list has its own effect" chose the bool; already logged in `docs/ideas/shared-list-helpers.md#catalog-wide-list-flags`. Best landed before S12-05 puts it on the wire |
| MNT-2 | MNT | major | Seventh copy of the ref-mint-and-insert loop; more private `uuidJSON`/`nullStr` | Follow-up | D-8 and plan Out line chose the copy; already logged in `docs/ideas/shared-list-helpers.md#one-ref-mint-loop` |
| UX-1 | UX | minor | `subjects.graph_id` inbound edge has no `viaFormats` delete-confirm heading | Follow-up | §6 Strings: None; no macOS in this PR. Put the line in the S12-04/S12-05 plan Strings |
| UX-2 | UX | minor | Three new `evidencegraphs.*` codes have no `L10n.Errors` mapping | Follow-up | D-17 defers; unreachable at this head. S12-05 DoD should check each code |
| SEC-1 | SEC | minor | `PRAGMA foreign_key_check` result rows are discarded by `migrate` | Follow-up | mechanism is `migrate.go` (unchanged here); same gap in 000030/000052. `docs/ideas/` |
| MNT-4 | MNT | minor | `seedGraphRefless`/`graphRef` fixture copied into two test packages | Follow-up | T-15/T-16 are named plan tests; files not in the fix plan |
| MNT-5 | MNT | minor | `effects.Set` doc comment lists an incomplete set of catalog-wide flags | Follow-up | `registry.go:28`; file not in the fix plan (rides with MNT-1's follow-up) |

Nits (logged): `CON-1 · effects/registry_completeness_test.go:3-6 · regroup stdlib imports` · `MNT-6 · evidencegraphs.go:316-322 · LinkArtifact's 16-byte check repeats checkLink` · `MNT-7 · catalogsession/session.go:102-105 · comment says "Create"; name onboarding.createCatalog` · `MNT-8 · defaultGraphID copied into four test files; move onto S12-04's List` · `UX-3 · evidencegraphs.go:18 · DefaultName "Default" is English catalog data; note for S12-11/S12-14`

CON: clean. UX: clean (verdict). SEC+PRF, MNT: changes needed.

**Paused:** PRF-1 to user (contradicts D-15). User 2026-10-19: "Override D-15 for PRF-1: run the graph-ref backfill in one transaction with the maxRefRetries / IsUniqueConflict retry, plus a test that forces a collision."

### Fix plan 1

**F-1 — Backfill graph refs in one transaction, retrying a ref collision** (from PRF-1; user override of D-15's "no retry")
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`, plan D-15 Landed note / deviation line in `S12-03-evidence-graphs-catalog.md`.
Failing test first: `TestEnsureRefs/retries_a_ref_collision` — seed two Sources (graphs A and B); set B's `ref` NULL and keep A's; make the package's mint seam return A's existing ref on its first call and a fresh `ref.Mint` after; after `EnsureRefs`, B has a valid `GPH-` ref ≠ A's and A is unchanged. Red today: the UPDATE returns a UNIQUE error.
Do: add an unexported seam `var mintRef = func() (string, error) { return ref.Mint(ref.PrefixEvidenceGraph) }` and use it in both `insert` and `EnsureRefs` (tests restore it with `t.Cleanup`). Rewrite `EnsureRefs`: read the NULL-ref ids as today, then `db.Begin()`; for each id loop `attempt < maxRefRetries` — `mintRef()`, `tx.Exec(sqlSetRef, r, id)`, break on nil, continue on `database.IsUniqueConflict(err)`, roll back and return on any other error; after the loop, a still-failing row rolls back and returns `ErrInvalid` (as `insert` does); `Commit()` at the end. Same retry shape as `insert` (lines ~80–95); do not change `users.EnsureRefs`. Update D-15's doc line in the plan's Landed note as a deviation ("backfill runs in one transaction with the `insert` retry, per user override of D-15 in review round 1"), and any doc sentence this PR added that says the graph backfill has no retry (`grep -rn "no retry" docs .agents core/database/evidencegraphs`).
Check: `TestEnsureRefs` (all cases) passes; scripts/check-pr.sh PASS.

**F-2 — Name `ReconcileOnOpen` in the `add-catalog-ref` skill** (from MNT-3)
Files: `.agents/skills/add-catalog-ref/SKILL.md`.
Failing test first: none (docs).
Do: change checklist step 3 to: "**Backfill** — if opening old catalogs, `EnsureRefs(c)` in the table package (see `users.EnsureRefs`, `evidencegraphs.EnsureRefs`), mint in one transaction with the `maxRefRetries` / `IsUniqueConflict` retry when the table grows with the catalog, and call it once from `catalogsession.ReconcileOnOpen`; never from an opener directly." Add `evidencegraphs.EnsureRefs` and `catalogsession.ReconcileOnOpen` to the Reference implementation list beside `users.EnsureRefs`.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `e9cd7580`. Gate PASS (orchestrator run; `TestEnsureRefs` re-run uncached, 3/3 pass). F-1 done (`612cfd39`: `mintRef` seam, one-transaction backfill with retry; extra subtest `exhausted_retries_roll_back`; plan Landed note records the D-15 deviation). F-2 done (`e9cd7580`).

## Round 2 — review
Reviewers: SEC+PRF (PRF-1), MNT (MNT-3). Scope `686c8e29...e9cd7580`. Not sent: CON, UX (no Fix-now blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | major | (round 1) graph-ref backfill retry and one transaction | Resolved | `evidencegraphs.go:365-389`: one `db.Begin`, `maxRefRetries` loop on `IsUniqueConflict`, `ErrInvalid` + rollback on exhaustion; `TestEnsureRefs` 3/3 |
| MNT-3 | MNT | major | (round 1) skill names `ReconcileOnOpen` | Resolved | `add-catalog-ref/SKILL.md:74,88-89` match `session.go:49-55` and `docs/catalog-refs.md:112` |

New blockers: none. MNT note (not raised): the mint-and-retry loop now appears twice in `evidencegraphs.go` (`insert`, `EnsureRefs`); the "One ref mint loop" follow-up (MNT-2) should cover the backfill path too.

## Result
Gate PASS · CI pending · open: none · follow-ups: MNT-1, MNT-2 (incl. the backfill copy), MNT-4, MNT-5, UX-1, UX-2, SEC-1 · to user: PRF-1 (answered: override D-15, fixed) · CON miss: none
