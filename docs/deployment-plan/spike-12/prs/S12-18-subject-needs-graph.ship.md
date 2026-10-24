# Ship log — S12-18 A Subject needs its graph

**Plan:** [`S12-18-subject-needs-graph.md`](S12-18-subject-needs-graph.md) · **Branch:** `s12-18-subject-needs-graph` · **PR:** #532
**State:** ready

## Round 0 — implement
Head `94d1ee2f`. Gate PASS (orchestrator run). PR #532 (draft). Report deviations (all recorded in the plan's Landed note):
- `personBeside` in `api/ffi/handlers/promote_test.go` takes its Artifact's Source from the created Subject (D-9's select returns `graph_id, subject_type_id` only).
- `newSubject` closures in `graphfill/load_test.go` and `observations/observations_test.go` fall back to `evidencegraphstest.SoleGraph` on a nil graph id instead of changing callers.
- Deleting the D-10 "create joins the Source's graph" case drops its create/delete-change `graph_id` check; "graph id given" still checks the create change.
- Helpers outside the D-5 (a) list read `SoleGraph` on their first line, per D-5 (b).
- AC-1.6 grep run over non-test files only.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or new error codes). SEC+PRF routed by `core/database/**` and the Insert path that loads a stored graph id.

CON: clean (no findings). SEC+PRF: clean (no findings). MNT: 1 major, 1 minor, 1 nit, 1 outside-lens note.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | ~25 copied test Subject-create helpers each grew a `SoleGraph` read instead of one shared fixture | To user → Fix now → F-1 | Verified: 60 added `evidencegraphstest.SoleGraph(` lines; `docs/ideas/test-subject-fixtures.md` names the same copies. Same-pass cleanup shape (this PR touches the copies) → Fix now by rule, but plan D-5 / Out records an explicit user ruling (2026-10-24) to keep this PR mechanical. Rule vs specific ruling → user. **User (2026-10-23): fold now.** Supersedes D-5's deferral and the Out line "Folding the ~25 test helpers…" |
| MNT-2 | MNT | minor | Five unrelated `SoleGraph(t, …)` calls rewritten to `SoleGraph(f.t, …)` | Fix now → F-1 | verified in diff: `conclusiondetails_test.go:380`, `evidence_graph_walk_test.go:219,266,314`, `grapheventtitles_test.go:302`; none builds a `CreateInput` |
| MNT-4 (CON, routed from MNT outside-lens) | CON | minor | Deleting D-10's case removed the only check that a Subject's delete change carries `graph_id` | Fix now → F-2 | verified: `rg '"graph_id"\]' core` matches only the create check at `subjects_test.go:136`; Landed note acknowledges the gap |

Nits (logged): `MNT-3 · core/database/subjects/subjects.go:99 · InsertTx doc could say the graph is required and the Source comes from it`

### Fix plan 1
**F-1 — One shared test fixture for creating a Subject** (from MNT-1; user ruling 2026-10-23 supersedes plan D-5's deferral and the Scope Out line "Folding the ~25 test helpers…")
Files: new `core/database/subjects/subjectstest/subjectstest.go`; `core/database/artifacts/artifactstest/artifactstest.go`; every `_test.go` in `core` and `api` that calls `subjects.Create(` (list: `rg -l 'subjects\.Create\(' core api -g '*_test.go' -g '!core/database/subjects/**'`); `docs/ideas/test-subject-fixtures.md`; the plan's Landed note. Skill: `add-catalog-query` (test helper package layout).
Failing test first: none (test-only refactor); the existing suites are the proof.
Do:
1. Move `artifactstest.SubjectOn` to `subjectstest.On(t, c, userID, graphID, typeKey, label) subjects.Subject` (same body: type lookup by key, `writes.Call("create_subject")`, `subjects.Create`). Add a sibling `subjectstest.OfType(t, c, userID, graphID, subjectTypeID, label) subjects.Subject` for callers that already hold a type id. Delete `artifactstest.SubjectOn`; move its callers to `subjectstest.On`.
2. Point every test helper or inline block whose whole job is "create one Subject in its own write" at `On` / `OfType`: the copies MNT-1 names (`(*fixture).bare` in promotealign and conclusionheaders/walks, `(*batchFixture).bare`, `(*bridgeWorld).node`, `(*walkFixture).subject`, `mustSubject` in identityclaims, `subjectOn` in autoreconciler, the `mk`/`mkSubject`/`newSubject` closures in effects, graphfill, observations, `mustPerson` in subjectpositions, and the helpers in likenesses, conclusiondetails, grapheventtitles, promote/pins, promote/promote, graphprogress, deleteimpact, citations, sources, sourcecopy, properties, connect, search/handles, and the two `api/ffi/handlers` tests). A helper that becomes a one-line wrapper is deleted and its callers call `subjectstest` directly. Keep D-5's placement rule: graph reads stay outside any write; fixed-Source fixtures keep their `graph` field.
3. Leave `subjects.Create` direct where the test is about Create itself or creates the Subject inside a larger transaction with other ops (for example `core/writes/run_test.go`, `core/database/audit/scopes_test.go`, `core/database/evidencegraphs/evidencegraphs_test.go`, and the `core/database/subjects/**` tests). Name each one you leave in your report with a one-line reason.
4. Helpers that follow `subjects.Create` with a separate `subjectpositions.Set` write keep doing so (the idea file's "One placement write" section stays an idea).
5. Docs: in `docs/ideas/test-subject-fixtures.md`, remove the "One way to make a test Subject" section (shipped in S12-18) and keep "One placement write"; update its Status line. Add to the plan's Landed note: "MNT-1 (ship review): user ruled 2026-10-23 to fold the copied create helpers into `subjectstest` on this PR, superseding D-5's deferral and the Out line."
Check: `rg -n 'subjects\.Create\(' core api -g '*_test.go' -g '!core/database/subjects/**'` prints only the files you list under step 3; `rg -n 'SubjectOn\b' core api` prints nothing; `rg -c 'evidencegraphstest\.SoleGraph\(' core api -g '*_test.go'` total is lower than at head `94d1ee2f`; scripts/check-pr.sh PASS.

**F-2 — Restore `SoleGraph(t, …)` where the round only renamed the handle** (from MNT-2)
Files: `core/database/conclusiondetails/conclusiondetails_test.go:380`, `core/database/graphcache/evidence_graph_walk_test.go:219,266,314`, `core/database/grapheventtitles/grapheventtitles_test.go:302`.
Do: after F-1, any of these five lines that still exist and still read `SoleGraph(f.t, …)` in a function that has its own `t` go back to `t`. Lines F-1 removed need nothing.
Check: `git diff 1843cd4f -- <those files> | grep -E '^[-+].*SoleGraph\((f\.)?t,'` shows no `t`→`f.t` rewrite on a line that builds no Subject; scripts/check-pr.sh PASS.

**F-3 — Check that a Subject's delete change carries `graph_id`** (from MNT outside-lens note, CON)
Files: `core/database/subjects/subjects_test.go` (case "graph id given").
Failing test first: none needed — it is a restored assertion; write it, run it, and confirm it passes against current code (if it fails, report it: that is a real defect).
Do: in "graph id given", after the create check, delete the Subject with `subjects.Delete` through the same write path the deleted D-10 case used, and assert the delete change's `graph_id` equals the graph id given.
Check: `rg -n '"graph_id"\]' core/database/subjects` matches the create and the delete assertion; scripts/check-pr.sh PASS.

## Round 1 — fix
Head `9006b72b` (commits `691b4a6b` F-3, `9006b72b` F-1+F-2). Gate PASS (orchestrator run). F-1 done: `subjectstest.On` / `OfType`, `artifactstest.SubjectOn` deleted, ~40 test files moved. Left on `subjects.Create`: `core/writes/run_test.go`, `audit/scopes_test.go`, `effects/*` (Placement or effects txn), `connect_test.go`, `citations/delete_test.go`, `observations_test.go`, `deleteimpact/impact_test.go`, `promote/bridges_test.go` `node` — every one passes a `Placement` or runs inside a larger transaction; `subjects/**` tests. Implementer deviation: the fixture takes no Placement, so Placement creates stay direct (idea file's "One placement write" section stays). `subjectstest` fails with `t.Fatal`, including inside `withProjectCatalog` callbacks in `api/ffi/handlers` tests.
- MNT-2 → **Resolved** (orchestrator read): the three remaining `SoleGraph(f.t, …)` lines in those files are new reads inside Subject helpers; the five handle-only rewrites are gone.
- CON (F-3) → **Resolved** (orchestrator read): `subjects_test.go:148` asserts the delete change's `graph_id` Old.

## Round 2 — review
Reviewers: MNT (had the Fix-now major MNT-1). Not sent: CON, SEC+PRF (no Fix-now blocker or major).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 (r2) | MNT | major | Fold only partly done: 14 single-write creates that pass a `Placement` (8 files: connect, effects ×3, promote/bridges `node`, citations/delete, deleteimpact/impact, observations) still copy the fixture body | To user → Fix now → F-4 | **User (2026-10-23): finish the fold on this PR, no third review round; orchestrator checks the diff.** Not resolved. Verified: `rg -n 'subjects\.Create\(' core api -g '*_test.go' -g '!core/database/subjects/**'` prints 19 lines; F-1 step 3 allowed no Placement exception. Open major after round 2, and the cap allows Fix plan 2 for blockers only → user decides between a placement-aware fixture on this PR or deferring to the idea file |

New round-2 findings: none.

### Fix plan 2
**F-4 — Placement-aware fixture; move the 14 Placement creates onto it** (from MNT-1 r2; user ruling 2026-10-23)
Files: `core/database/subjects/subjectstest/subjectstest.go`; `core/database/connect/connect_test.go`; `core/database/effects/handle_paths_test.go`, `evidence_handles_test.go`, `conclusion_handles_test.go`; `core/database/promote/bridges_test.go`; `core/database/citations/delete_test.go`; `core/database/deleteimpact/impact_test.go`; `core/database/observations/observations_test.go`; `docs/ideas/test-subject-fixtures.md`; the plan's Landed note.
Failing test first: none (test-only refactor).
Do:
1. In `subjectstest`, add `At(t testing.TB, c *database.Catalog, userID, graphID, subjectTypeID []byte, label string, at subjects.Placement) subjects.Subject`. It has the same body as `OfType` but passes `&at`. Make `OfType` and `At` share one unexported body, so the create is written once; `On` keeps resolving its type key and calls that body.
2. Move every single-write create that passes a `Placement` onto `At`: the two `create` closures in `connect_test.go` (≈:110, :429), the `mk` closures in the three effects files, `(*bridgeWorld).node` in `promote/bridges_test.go`, and the inline creates in `citations/delete_test.go` (≈:78, :86), `deleteimpact/impact_test.go` (≈:232, :240, :516) and `observations/observations_test.go` (≈:114, :122, :130). Delete any closure that becomes a one-line wrapper and call `subjectstest.At` directly. Keep the graph reads outside any write (D-5). A site that uses the create's changes (not `_`) and needs them stays direct: name it in your report with the line.
3. Leave direct: `core/writes/run_test.go`, `core/database/audit/scopes_test.go`, `core/database/effects/evidence_graphs_test.go`, and the `core/database/subjects/**` tests.
4. `docs/ideas/test-subject-fixtures.md`: fix the dangling "Several of those helpers" so the remaining "One placement write" section stands on its own (name the helpers that run a second `subjectpositions.Set` write: promotealign, canonicalgraph, graphcache walk, promote batch, autoreconciler, grapheventtitles, conclusionheaders walks, search handles, and say `subjectstest.At` is the target). Add one line to the plan's Landed note: the fixture gained `At` for creates with a Placement.
Check: `rg -n 'subjects\.Create\(' core api -g '*_test.go' -g '!core/database/subjects/**'` prints only lines in `core/writes/run_test.go`, `core/database/audit/scopes_test.go` and `core/database/effects/evidence_graphs_test.go` (plus any site you name under step 2); `rg -n 'subjects\.Create\(' core/database/subjects/subjectstest` prints one line; scripts/check-pr.sh PASS.

## Round 2 — fix
Head `b6574336`. F-4 done: `subjectstest.At` added; `On` → `OfType` → shared `create`; the 14 Placement creates moved. Orchestrator read the diff (11 files, +35 −124): `rg 'subjects\.Create\(' core api -g '*_test.go' -g '!core/database/subjects/**'` prints only `core/writes/run_test.go:157`, `audit/scopes_test.go:101,389`, `effects/evidence_graphs_test.go:192,225` (Create's own pipeline, larger transactions, or the changes are used). MNT-1 → **Resolved**.
Gate: first run FAIL in `xcode-test` on `GraphCanvasEdgeGeometryTests.bottomApproachEndsUnderTheBridgeFill()`. The diff has no Swift (only `.go` and `.md` files), and the same test passed on the three earlier gates. Rerun PASS. Logged as flaky, not a gate failure of this PR.
`origin/main` has not moved since the branch was cut (`1843cd4f`); no merge needed.

## Result
Gate PASS · CI <pending> · open: none · follow-ups: flaky `GraphCanvasEdgeGeometryTests.bottomApproachEndsUnderTheBridgeFill()`; `docs/ideas/test-subject-fixtures.md` "One placement write" · to user: MNT-1 (fold now, 2026-10-23), MNT-1 r2 (finish fold, Fix plan 2 for a major, 2026-10-23) · CON miss: none
