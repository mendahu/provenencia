# Ship log — S12-04 Graph lifecycle

**Plan:** [`S12-04-graph-lifecycle.md`](S12-04-graph-lifecycle.md) · **Branch:** `s12-04-graph-lifecycle` · **PR:** #505
**State:** review 2

Worktree: `/Users/jakerobins/Repos/provenencia-s12-04`, branched from `origin/main` @ `9cf8e33c`.

## Round 0 — implement
Head `beb8be43`. Gate PASS (orchestrator run). Report: no deviations from the plan; T-6 and T-7 needed no production change; subtest names partly the implementer's. The implementer committed this ship log stub in `ffae89aa` (reported as uncommitted); harmless.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, UX (`core/apperr`, new error codes `evidencegraphs.in_use`, `evidencegraphs.last_graph`), SEC+PRF (performance: `core/database/**`; security: `core/catalogsession` test, stored lists loaded by `List`). Skipped: none.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| PRF-1 | PRF | blocker | `? IS NULL OR source_id = ?` defeats `evidence_graphs_source_id_idx`; every `SoleGraphTx` / `LinkArtifact` / `Update` scans all graph links | To user → Fix now → F-1 | verified: `EXPLAIN QUERY PLAN` on the migration-55 shape gives `SCAN l` / `SCAN evidence_graphs`. Contradicts D-2's SQL text; user chose fix on this PR (2026-10-20) |
| MNT-1 | MNT | major | `Update` adds a third private `containsID` copy | Fix now → F-2 (same-pass cleanup) | verified: copies at `identityclaims/bridgefile.go:449` `containsID`, `graphcache/verify.go:266` `hasID`, tests `promote/batch_test.go:213` `containsID`, `conclusionheaders/walks_test.go:389` `hasID` |
| MNT-2 | MNT | major | `add-catalog-query` teaches `Querier` reads only; the multi-query read-transaction shape `List` uses is undocumented | Fix now → F-3 | verified: skill § Reads inside a write has no `db.Begin` read; this PR adds the pattern's second instance |
| MNT-3 | MNT | minor | `Graph.ArtifactIDs` comment says "by Artifact ref"; `Create` returns input order | Fix now → F-4 | verified `evidencegraphs.go:74`; D-1's comment wording, a stale-comment fix not a behavior change |
| MNT-4 | MNT | minor | `TestImpactEvidenceGraph` inlines `evidencegraphstest.SoleGraph` | Fix now → F-4 | verified `impact_test.go:818-822` |
| UX-1 | UX | minor | A refused graph edit cannot say which removed Artifact is in use | Later PR | written into the S12-05 map scope entry (`deployment-plan.md` § S12-05 Ships): preflight takes the removed-Artifact set |

Nits (logged): `CON-1 · ship log State/PR stale · updated by orchestrator` · `CON-2 · TestUpdate AC-3.6 fixture not read by reviewer` · `UX-2 · same-named graphs share a delete-confirm title; S12-11/12 show the ref` · `MNT-5 · evidencegraphs.go:29 · ErrLastGraph doc comment style` · `MNT-6 · evidencegraphs.go:334 · requireDistinct could use database.UniqueBlobIDs`
Security: no findings. CON: clean.

### Fix plan 1

**F-1 — The per-Source graph read seeks the `source_id` index** (from PRF-1)
Files: `core/database/evidencegraphs/evidencegraphs.go`, `core/database/evidencegraphs/evidencegraphs_test.go`.
Failing test first: `TestListQueryPlan` — runs `EXPLAIN QUERY PLAN` on the per-Source graph query and link query (through the package's own SQL consts, in-package test or an exported-for-test hook, whichever the package's tests already use) against a created catalog, and fails on any `SCAN evidence_graphs` or `SCAN l`/`SCAN evidence_graph_artifacts` row; it must be red on the current SQL.
Do: keep one `list(tx, sourceID)`. Replace the `? IS NULL OR …` SQL with two forms chosen by `len(sourceID) == 0`: per Source, `sqlListBySource = … FROM evidence_graphs WHERE source_id = ? ORDER BY name COLLATE NOCASE, ref, id` and the link query driven from `evidence_graphs g` with `g.source_id = ?`, joined to `evidence_graph_artifacts l` on `l.graph_id = g.id` and to `artifacts a`, same ORDER BY; project-wide, the same SELECTs with no WHERE. Bind `sourceID` directly (drop the twice-bound `NullBlob`). Share the column list and ORDER BY between the forms (a const prefix/suffix) so the two forms cannot drift. Supersedes D-2's SQL text (user, 2026-10-20); write that in the plan's Landed note as a review deviation.
Check: `TestListQueryPlan` passes; `TestList` and every existing test pass; scripts/check-pr.sh PASS.

**F-2 — One shared byte-id membership helper** (from MNT-1; same-pass cleanup)
Files: `core/database/sqlin.go` (beside `UniqueBlobIDs`), `core/database/evidencegraphs/evidencegraphs.go`, `core/database/identityclaims/bridgefile.go`, `core/database/graphcache/verify.go`, `core/database/promote/batch_test.go`, `core/database/conclusionheaders/walks_test.go`.
Do: add `func ContainsBlobID(ids [][]byte, id []byte) bool` to package `database` (one-line doc comment). Delete `containsID` in `evidencegraphs`, `identityclaims`, `promote` tests and `hasID` in `graphcache`, `conclusionheaders` tests; call `database.ContainsBlobID`. Leave `core/subjectprofile` and `core/graphalign` (pure layers that do not import `database`; their inline `slices.ContainsFunc` stays).
Check: `grep -rnE "^func \w+\(\w+ \[\]\[\]byte, \w+ \[\]byte\) bool" core api` prints only `core/database/sqlin.go:<n>:func ContainsBlobID(ids [][]byte, id []byte) bool {`; scripts/check-pr.sh PASS.

**F-3 — `add-catalog-query` documents the read-transaction read** (from MNT-2)
Files: `.agents/skills/add-catalog-query/SKILL.md`.
Do: in § Reads inside a write (after the `Querier` paragraph), add a short paragraph: a read made of more than one query opens one read transaction (`db.Begin`, deferred `Rollback`), as `deleteimpact.Snapshot` and `evidencegraphs.List` do, so its queries see one state; its unexported body takes `*sql.Tx` and the package's writes call it with theirs; an exported `…Tx` twin is added only with its first caller outside the package.
Check: scripts/check-pr.sh PASS (plan-docs).

**F-4 — Comment and fixture minors** (from MNT-3, MNT-4)
Files: `core/database/evidencegraphs/evidencegraphs.go:74`, `core/database/deleteimpact/impact_test.go:818-822`.
Do: `ArtifactIDs` comment → "linked Artifacts: by Artifact ref from List, in input order from Create". In `TestImpactEvidenceGraph`, replace the inline `List` + length check with `def := evidencegraphstest.SoleGraph(t, c, src.ID)`. Update D-1's comment quote in the plan's Landed note as a deviation alongside F-1's.
Check: `go test -tags fts5 ./core/database/deleteimpact/` passes; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `5d7a4bdb`. Gate PASS (orchestrator run; the implementer saw one flaky `GraphCanvasEdgeGeometryTests.bottomApproachEndsUnderTheBridgeFill` failure with no Swift change, green on rerun). F-1 done (`TestListQueryPlan` red first on `SCAN evidence_graphs` / `SCAN l`). F-2 done; deviation: `graphcache` cannot import `database` (`database/catalog.go` imports `graphcache`), so its `hasID` became an inline `slices.ContainsFunc` over its own `sameID`. F-3, F-4 done. Landed note records the D-2 SQL and D-1 comment changes.
MNT-3, MNT-4 (minors): Resolved — read in `5d7a4bdb` (comment at `evidencegraphs.go`, `SoleGraph` in `impact_test.go`).

## Round 2 — review
Reviewers: MNT (MNT-1, MNT-2), SEC+PRF (PRF-1). Scope `beb8be43...5d7a4bdb`.
