# Ship log — S12-09 Promote per graph

**Plan:** [`S12-09-promote-per-graph.md`](S12-09-promote-per-graph.md) · **Branch:** `s12-09-promote-per-graph` · **PR:** #535
**State:** ready

## Round 0 — implement
Head `6b053767`. Gate PASS (orchestrator re-run against `f855aaa0`). Report: deviations none; §10 manual check not run in the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX, SEC+PRF. Skipped: none — the diff touches `macos/App/**` (UX), `api/ffi/**` (SEC), and `core/database/**` plus `macos/App/Features/Promote/PromoteModel.swift` (PRF).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | `bridgeHeads` starts from `st.key IN (…)` + `subjects_subject_type_id_idx`, so Unpromote (scope lost with `s.source_id`), every `FileBridgesTx` and every Done read every bridge Subject in the catalog | Fix now → F-1 | verified: `bridgefile.go` query joins `subject_types` first with the type IN list and no index-steering; the catalog never runs `ANALYZE`; reviewer's `EXPLAIN QUERY PLAN` on a migrated scratch DB shows the type index for all three conds (36.6 / 94.7 / 21.7 ms at 66k bridges) vs `subjects_source_id_idx` at base |
| MNT-1 | MNT | major | New `promoteOne` is another copy of the promote-in-a-write block; ~30 copies across ~15 test packages | Fix now → F-2 (same-pass cleanup) | verified: `rg -n 'promote\.Save\(tx' core api --glob '*_test.go'`; this PR adds `promoteOne` and edits files holding copies (`audit/scopes_test.go`, `deleteimpact/*`, `effects/conclusion_handles_test.go`, `promotealign/*`, `search/handles_test.go`) |
| CON-1 | CON | minor (re-rated from major: a missing doc paragraph, no code effect) | Plan has no Landed note; DoD requires one | Fix now → F-4 | verified: `grep -n Landed` on the plan finds only the Status line; DoD line 277 asks for it |
| PRF-2 | PRF | minor | `UnfileBridges` runs one `AcceptedEntityForSubjectTx` per bridge head | Fix now → F-3 | minor in a file this PR changes; base joined `identity_claims` in one query |

Nits (logged): `CON-N1 · api/proto/engine.proto:1151 · re-wrap UnpromoteSubjectRequest comment` · `CON-N2 · core/database/promote/batch.go:38 · re-wrap Batch comment` · `CON-N3 / MNT-3 · core/database/promote/unpromote.go:18 · reflow short comment line` · `MNT-2 · core/database/identityclaims/bridgefile.go:3-8 · move rowchange import to the module group`

UX: clean, no findings.

### Fix plan 1

**F-1 — Bridge-head queries start from their selective condition** (from PRF-1)
Files: `core/database/identityclaims/bridgefile.go`, `core/database/identityclaims/*_test.go`.
Failing test first: `TestBridgeHeadQueryPlans` in `identityclaims` — for each of the three conditions (FileBridgesTx, FileEvidenceGraphBridgesTx, UnfileBridges), run `EXPLAIN QUERY PLAN` on the exact SQL `bridgeHeads` builds (expose the builder to the test via `export_test.go` or build the string in a small unexported func the test calls) and fail when any row names `subjects_subject_type_id_idx` or starts with `SCAN s` / `SCAN subjects`; require the subjects step to use the primary key (`sqlite_autoindex_subjects_1` / `INTEGER PRIMARY KEY` / `id=?`) for the two subject conds and `subjects_graph_id_idx` for the graph cond. Follow the shape of `evidencegraphs_test.go:1205-1240`. Red on head `6b053767`.
Do: in `bridgeHeads` write `FROM subjects s CROSS JOIN subject_types st ON st.id = s.subject_type_id AND st.origin = 'provenencia'` (as `graphcache/load.go:242-253` does, with a one-line comment why). Make `touchesEnd` `s.id IN (SELECT o.subject_id FROM observations o WHERE o.value_subject_id = ?)`. Make `FileBridgesTx`'s cond `s.id IN (SELECT ? UNION ALL SELECT o.subject_id FROM observations o WHERE o.value_subject_id = ?)` (same two args). Keep `ORDER BY s.id` and results unchanged.
Check: the new test passes; existing `identityclaims`, `promote`, `promotealign` tests pass unchanged; scripts/check-pr.sh PASS.

**F-2 — One promote test helper, every copy moved onto it** (from MNT-1; same-pass cleanup)
Files: new `core/database/promote/promotetest/promotetest.go`; every `*_test.go` under `core` and `api` that the grep below lists.
Failing test first: none (refactor; the moved tests are the check).
Do: add package `promotetest` (doc comment in the `evidencegraphstest` style) with `func Save(t testing.TB, c *database.Catalog, userID []byte, in promote.Input) promote.Result` that runs `writes.Call(c, writes.Op{Action: "promote_subject", UserID: userID}, …promote.Save(tx, userID, in)…)`, calls `t.Helper()`, and `t.Fatal`s on error. Replace every call site whose write closure does nothing but `return promote.Save(tx, u, in)` and whose error is fatal (`t.Fatal`, `must`, `f.t.Fatal`) with `promotetest.Save(...)`. Delete the now-thin wrappers or make them one-line calls to it: `(*batchFixture).promoteOne` (delete; its callers and the four inline blocks in `batch_test.go` use `promotetest.Save`), `(*bridgeWorld).promote` / `join`, `(*fixture).promote` in `promotealign_test.go`, `(*handleFixture).promote` in `core/search/handles_test.go`, and the inline blocks in `effects`, `deleteimpact`, `audit`, `conclusionheaders`, `conclusiondetails`, `likenesses`, `canonicalgraph`, `autoreconciler`, `graphcache`, and `api/ffi/handlers/conclusion_headers_test.go`. Leave a site alone when it asserts on Save's error, runs other writes in the same closure, or is in an in-package (non-`_test` package) test that would create an import cycle; list each such site in your report.
Check: `rg -n 'promote\.Save\(tx' core api --glob '*_test.go'` prints only `promotetest/promotetest.go` (not a test file — it will not print) plus `core/database/promote/promote_test.go` error-path cases and the sites your report lists with their reason; `rg -n 'func \(\w+ \*\w+\) promote\w*\(' core api --glob '*_test.go'` prints only one-line delegates or nothing; scripts/check-pr.sh PASS.

**F-3 — Unfile reads claims with the heads** (from PRF-2)
Files: `core/database/identityclaims/bridgefile.go`.
Failing test first: none (behavior unchanged; `TestUnpromote` covers it).
Do: give `bridgeHead` a nullable `claimID []byte` filled by `LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted'` in `bridgeHeads` (the `identity_claims_one_accepted_per_subject` partial unique index keeps it one row per head). `UnfileBridges` skips heads with nil `claimID` and passes it to `deleteimpact.ReleaseClaim`; drop the per-head `AcceptedEntityForSubjectTx`. `fileHeads` ignores the field. The F-1 plan test must still pass with the join added.
Check: `TestUnpromote`, F-1's plan test, `identityclaims` tests pass; scripts/check-pr.sh PASS.

**F-4 — Landed note and nits** (from CON-1; nits CON-N1..N3, MNT-2, MNT-3)
Files: `docs/deployment-plan/spike-12/prs/S12-09-promote-per-graph.md`, `api/proto/engine.proto` (then regenerate), `core/database/promote/batch.go`, `core/database/promote/unpromote.go`, `core/database/identityclaims/bridgefile.go`.
Do: add a `## Landed` section at the end of the plan listing every difference from the plan, including F-1 (CROSS JOIN, `IN` subquery conds — supersede D-1/D-2's query shape only), F-2 (new `promotetest` package; supersedes the Codebase-fit list not naming a shared helper), and F-3. Re-wrap the three comments; move the `rowchange` import into the module group.
Check: scripts/check-pr.sh PASS.

## Round 1 — fix
Head `2bd19fd4`. Gate PASS (orchestrator re-run, test cache cleared). F-1 `966e8695`, F-2 `45e3b9b1`, F-3 `4829760a`, F-4 `2bd19fd4` — all done. F-1 red shown by reverting the CROSS JOIN only (`file`, `unfile` subtests fail); the `graph` subtest was already green. F-2 leaves the `promote_test.go` error-path cases and the `autoreconciler_test.go:1509` random walk (tolerates errors). CON-1: **Resolved** (orchestrator read the plan's `## Landed`, lines 320–327).

## Round 2 — review
Reviewers: MNT (MNT-1 major), SEC+PRF (PRF-1 blocker). Not sent: CON (only a minor, resolved above), UX (no findings). Scope `6b053767...2bd19fd4`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | (round 1) bridge-head queries | Resolved | SEC+PRF round 2: all three plans seek `sqlite_autoindex_subjects_1` / `subjects_graph_id_idx`; `TestBridgeHeadQueryPlans` passes |
| PRF-2 | PRF | minor | (round 1) per-head claim read | Resolved | orchestrator read F-3: `claimID` from `LEFT JOIN identity_claims … 'accepted'`, one row per head by `identity_claims_one_accepted_per_subject`; SEC+PRF concurs |
| MNT-1 | MNT | major | (round 1) promote test helper | Resolved | MNT round 2: only error-path `promote_test.go` sites and the `autoreconciler_test.go:1508` walk remain; wrappers are thin delegates |
| MNT-7 | MNT | major (reviewer: below the round-2 bar) | `TestBridgeHeadQueryPlans` is a 4th copy of the `EXPLAIN QUERY PLAN` scan loop (`canonicalgraph_test.go:212`, `evidencegraphs_test.go:1221`, `graphcache/graph_test.go:230`, `identityclaims_test.go:417`) | To user → Fix now → F-5 | user 2026-10-24: "Yes — run Fix plan 2 (F-5 query-plan helper and F-6 nits) on this PR, gate it, then finish." |

Nits (logged): `MNT-4 · core/database/promote/promote_test.go:150-264 · p6a6…p6a15 leftover names` · `MNT-5 · handles_test.go:105, bridges_test.go:187, promotealign_test.go:252 · return promotetest.Save(...) directly` · `MNT-6 · core/database/promote/unpromote.go:20 · short comment line`

### Fix plan 2 (approved by the user 2026-10-24)

**F-5 — One query-plan test helper** (from MNT-7; same-pass cleanup)
Files: a shared test-support home for `core/database` (reuse an existing `*test` package if one fits; otherwise a new `core/database/dbtest` package with a `planTest`-style doc comment), and the four test files above.
Failing test first: none (refactor; the four plan tests are the check).
Do: add `QueryPlan(t testing.TB, db *sql.DB, query string, args ...any) []string` that runs `EXPLAIN QUERY PLAN`, scans `(id, parent, notused, detail)`, `t.Fatal`s on error, closes rows, and returns the detail lines. Move all four loops onto it, keeping each test's own assertions (SCAN prefix, CONSTANT exception, index-name checks) unchanged.
Check: `grep -rn 'EXPLAIN QUERY PLAN' core api --include='*.go'` prints only the helper; the four tests pass; scripts/check-pr.sh PASS.

**F-6 — Nits** (MNT-4, MNT-5, MNT-6) as logged above.

## Round 2 — fix
Head `309e1e05`. Gate PASS (orchestrator re-run, test cache cleared). F-5 `b41d562b`: new `core/database/dbtest.QueryPlan`; `grep -rn 'EXPLAIN QUERY PLAN' core api --include='*.go'` prints only the helper; orchestrator read the diff, and each of the four tests keeps its own assertions unchanged. F-6 `309e1e05`: MNT-4, MNT-5, MNT-6 done. MNT-7: **Resolved** (orchestrator read; no round 3, by the user's go-ahead for a non-blocker Fix plan 2). `main` has not moved since `f855aaa0`.

## Result
Gate PASS · CI green (test, localizable) · open: none · follow-ups: none · to user: MNT-7 (answered: fold in) · CON miss: none
