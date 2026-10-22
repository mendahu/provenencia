# Ship log — S12-07 Graph-keyed reads

**Plan:** [`S12-07-graph-keyed-reads.md`](S12-07-graph-keyed-reads.md) · **Branch:** `s12-07-graph-keyed-reads` · **PR:** #520
**State:** ready

## Round 0 — implement
Base `376eb9ad`. Head `cb010094`. Gate PASS (orchestrator run). Report deviations:
- STEP-2 and STEP-3 landed as one commit (STEP-3 cannot build part-way).
- T-5 is a subtest of `TestRetitleDropsEvidenceGraph`.
- AC-1.1's Bea is a place with a `toponym` text Observation (text is invalid on a person).
- Split tests add a second Artifact first (a graph cannot lose its last link).
- STEP-8's `rg` also prints `api/ffi/handlers/artifacts.go:215` (`artifacts.ListBySource`), unrelated; left alone.
- Handler file name `sourcegraphprogress.go` kept.
- Under `PROVENENCIA_GRAPH_VERIFY=1`, three tests fail (`TestMemoFollowsCommit/a_handle_merged`, `TestCommitOfAnAssociationLoadsEndpoints`, `TestCommitReloadsASharedEndpointOnce`); they fail the same way on `main` @ `376eb9ad`.
- Manual check (§10) not run in the app.

## Round 1 — review
Reviewers: CON (Sonnet), MNT, SEC+PRF. Skipped: UX (no `macos/App/**`, strings, `core/apperr`, or error codes in the diff).

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | Two-graph test setup (new Artifact, `Split`, Subject on graph) copied six more times | Fix now → F-1 | verified: `git grep evidencegraphstest.Split` lists 8 callers, 6 added here, each preceded by an Artifact create. Same-pass cleanup: supersedes plan §8 "no new helper". The PR also re-copies the `runArtifactCreate` body inline (`graphfill/load_test.go` `newArtifact`, `graphcache/evidence_graph_walk_test.go`), and that helper has 20 copies, so it moves onto the same fixture package |
| MNT-2 | MNT | minor | `graphprogress.Progress.SourceID` set by `List`, empty from `Get`; handler patches it | Fix now → F-2 | verified: `graphprogress.go:55` doc, handler `sourcegraphprogress.go` `row.SourceID = sourceID`. Minor in a file this PR changes |
| MNT-3 | MNT | minor | `docs/promote-graph-alignment.md:165` still says graphcache keeps "Source graphs" | Fix now → F-3 | verified: line 165 |
| SEC+PRF out-of-lens (→ CON) | SEC | — | Source with two graphs: Source-keyed handlers fail with `CodeInternalUnknown`; `ListSourceGraphProgress` returns two rows per Source | Rejected | accepted by D-7, D-8, D-9 until S12-08 / S12-14; no app path creates a second graph yet |

CON: clean, no findings. SEC+PRF: clean, no findings.

Nits (logged): `MNT-4 · core/database/promotealign/promotealign.go:95 · pass db directly instead of var q database.Querier` · `MNT-5 · core/database/promotealign/evidence.go:690 · titles' add(source, subject) passes the Source id memberOf already holds`

### Fix plan 1

**F-1 — One fixture package for "a second Artifact, a split graph, a Subject on a graph"** (from MNT-1)
Files: new `core/database/artifacts/artifactstest/artifactstest.go`; every test file listed below.
Failing test first: none (test-only refactor); the moved tests are the check.
Do:
- Add package `artifactstest` (it may import `artifacts`, `evidencegraphs`, `evidencegraphstest`, `subjects`, `subjecttypes`, `writes`; `evidencegraphstest` cannot import `artifacts` because `artifacts`' own tests import it). Three functions:
  - `Create(c *database.Catalog, userID []byte, in artifacts.CreateInput) (artifacts.Artifact, error)`: the `runArtifactCreate` body (`writes.Run` with `Action: "create_artifact"`), same signature so it drops in.
  - `Split(t testing.TB, c *database.Catalog, userID, sourceID []byte, name string) (def, graph evidencegraphs.Graph, page artifacts.Artifact)`: creates an Artifact labelled `name` on the Source, then calls `evidencegraphstest.Split(t, c, userID, sourceID, name, page.ID)`. Doc comment states why the Artifact is needed (a graph cannot lose its last link).
  - `SubjectOn(t testing.TB, c *database.Catalog, userID, sourceID, graphID []byte, typeKey, label string) subjects.Subject`: looks up the provenencia subject type by key and runs `subjects.Create` with `CreateInput{SourceID, GraphID, SubjectTypeID, Label}` through `writes.Call` (`Action: "create_subject"`), fatal on error.
- Move all 8 `evidencegraphstest.Split` callers onto `artifactstest.Split` (and `SubjectOn` where they create a Subject on a graph): `connect/connect_test.go`, `graphcache/evidence_graph_walk_test.go`, `grapheventtitles/grapheventtitles_test.go`, `graphfill/load_test.go` (fixture and the retitle subtest), `graphprogress/graphprogress_test.go`, `observations/observations_test.go`, `promotealign/evidence_test.go`, `subjectpositions/subjectpositions_test.go`. Where a test needs the extra Artifact under a specific label or reuses an existing one, `evidencegraphstest.Split` stays the call; say which in the report.
- Delete every `runArtifactCreate` copy (20 files: `grep -rln "func runArtifactCreate" core api`) and the inline `newArtifact` closure in `graphfill/load_test.go`; callers use `artifactstest.Create`. Keep a copy only where importing `artifactstest` makes an import cycle (an in-package test of a package `artifacts` imports, directly or not); name each kept copy and the cycle in the report.
- This supersedes plan §8 "Reused as the fixture for every new case; no new helper" — note it in the plan's Landed note.
Check: `grep -rnE "func runArtifactCreate|newArtifact :?= func" core api` prints only the cycle-kept copies the report names; `git grep -n "evidencegraphstest.Split" -- '*.go'` prints only `artifactstest` and the callers the report names; `scripts/check-pr.sh` PASS.

**F-2 — `graphprogress.Get` fills `SourceID`** (from MNT-2)
Files: `core/database/graphprogress/graphprogress.go`, `graphprogress_test.go`, `api/ffi/handlers/sourcegraphprogress.go`.
Failing test first: a `graphprogress` test case asserting `Get(c, graphID).SourceID` equals the graph's Source id (red now: it is empty).
Do: `Get` reads the graph's `source_id` from `evidence_graphs` by primary key in the same read (an unknown graph keeps zeros and an empty `SourceID`). Rewrite the `Progress` doc so `SourceID` means "the graph's Source; empty when the graph does not exist". Remove `row.SourceID = sourceID` after `Get` in the handler (keep the zero row it builds when `sourceGraphID` is not ok).
Check: the new case passes; `TestGetSourceGraphProgress` passes unchanged; `scripts/check-pr.sh` PASS.

**F-3 — Doc line** (from MNT-3)
Files: `docs/promote-graph-alignment.md:165`.
Do: "Source graphs" → "Evidence graphs (keyed by graph id)".
Check: `rg -n -i 'source graphs' docs/promote-graph-alignment.md` prints nothing.

## Round 1 — fix
Head `61a75310`. Gate PASS (orchestrator run). F-1, F-2, F-3 done.
- F-1: new `core/database/artifacts/artifactstest` (`Create`, `Split`, `SubjectOn`); all 8 `Split` callers moved; `runArtifactCreate` kept only in `sources/sources_test.go` (cycle: `artifactstest` → `evidencegraphstest` → `sources`). `evidencegraphstest.Split` and `SoleGraph` now take `testing.TB`. `observations` and `connect` keep their own Subject creators (they pass a `Placement` or a nil graph).
- MNT-2: **Resolved** — `Get` reads `source_id` from `evidence_graphs`; handler patch removed; field doc rewritten.
- MNT-3: **Resolved** — doc line reads "Evidence graphs (keyed by graph id)".

## Round 2 — review
Reviewers: MNT (only lens with a Fix-now major). Scope `cb010094...61a75310`.

| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| MNT-1 | MNT | major | (round 1) two-graph setup copies | Resolved | `git grep evidencegraphstest.Split` prints only `artifactstest.go:39`; one cycle-kept `runArtifactCreate` |

No new findings. Verdict: Clean. No Fix plan 2.

## Result
Gate PASS (`61a75310`, `main` unmoved at `376eb9ad`) · CI pending · open: none · follow-ups: — · to user: — · nits logged: MNT-4, MNT-5 · CON miss: none
