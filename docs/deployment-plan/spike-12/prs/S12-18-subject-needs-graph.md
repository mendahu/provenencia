# S12-18 — A Subject needs its graph

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 3 — Graph id end to end
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: Go only; no view, RPC or copy changes)
**Board:** —
**Depends on:** S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528))
**Blocks:** S12-09
**Branch:** `s12-18-subject-needs-graph`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-08 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-queries.mdc`
**Context:** [data model](../source-structure.md#data-model), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) (`subjects.graph_id`), [`source-layer-data-model.md`](../../../source-layer-data-model.md) § 7.1, the S12-08 [Landed note](S12-08-graph-id-end-to-end.md#landed) and its D-3 / D-15; working rules: "One bridge, one function", "High tolerance for churn", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-10-24 against `main` @ `c93b3fb`
**Review:** [`S12-18-subject-needs-graph.review.md`](S12-18-subject-needs-graph.review.md) — Fits with changes / Ready with changes, 1 round; every finding answered

### Implementer — read this first

1. **This file is the whole task.** Read it top to bottom before you change code. Then read every file under **Current state** and load every skill under **Skills to load**.
2. **Check the base.** Compare `main` with the **Planned** commit. If a file under **Current state** has changed so a row is no longer true, stop and report which rows.
3. **Decisions are made.** Do not reopen a **Decision** or a **Scope** line, and do not swap in a design you like better. If one cannot work, stop: say which ID it is, what you found (path, line, output), and what you would need. Do not improvise around it.
4. **Do only In.** Out stays out, even when it is one line away. Do not refactor, rename, or tidy code the plan does not name.
5. **Work the Steps in order.** Run each step's check before you start the next. A failing check is fixed in that step.
6. **Hit a Stop condition → stop.** Report it. Do not work around it.
7. **Done is the Definition of done.** Report each box with its evidence: the command and its result, the test name, or the manual check you ran.
8. **Docs-only lines in the DoD are part of the PR.** Do not skip them.

---

## 1. Objective

Every new Subject names the Evidence graph it joins, and the engine reads the Subject's Source from that graph. `subjects.CreateInput` loses `SourceID`; `GraphID` is required. The last interim path that found "the Source's one graph" for a new Subject (`subjects.InsertTx`'s `SoleGraphTx` fallback) is deleted, so a second graph on a Source can never send a Subject to the wrong graph. The researcher sees no change: every product caller already passes a graph id (S12-08).

**Why now.** The working rule "One bridge, one function" retires the bridge in three PRs: S12-08 removed the handlers' use, this PR removes `subjects.InsertTx`'s, and S12-09 removes Promote's and deletes `SoleGraph` / `SoleGraphTx`. S12-09 depends on this PR so that, when it deletes `SoleGraphTx`, Promote is its only caller. S12-08 already moved the rule "a Subject's Source is its graph's Source" into `InsertTx` (S12-08 D-3), so this PR deletes the fallback and the field.

**What stays.** `subjects.Subject.SourceID` and the `subjects.source_id` column (working rule "`subjects.source_id` stays"). The composite FK `(graph_id, source_id)`. `evidencegraphs.SourceOf`, `SoleGraph` and `SoleGraphTx` (S12-09 deletes the last two). The test fixture `evidencegraphstest.SoleGraph` (S12-08 D-15). Every RPC, the wire and Swift. Promote's Go code.

## 2. User stories

**US-1.** As the engine, when a Subject is created, I take its graph from the caller and read its Source from that graph, so a Subject can never land in a graph its caller did not name.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source Deed with graphs Default and Jones; `subjects.Create` with `CreateInput{GraphID: jones.ID, SubjectTypeID: person, Label: "X"}` | The returned and stored Subject has `GraphID` = Jones and `SourceID` = Deed |
| AC-1.2 | `CreateInput{SubjectTypeID: person}` (no graph id) | `subjects.ErrInvalid` (`subjects.invalid`); no row is written. Before this PR an empty `GraphID` with a `SourceID` joined the Source's sole graph |
| AC-1.3 | `CreateInput{GraphID: <16 bytes no graph has>}` | `subjects.ErrInvalid` |
| AC-1.4 | `CreateInput{GraphID: []byte{1, 2, 3}}` | `subjects.ErrInvalid` |
| AC-1.5 | `connect.CreateCitedBridge` between two Subjects on Jones | The bridge Subject has `GraphID` = Jones and `SourceID` = Deed (from its endpoint's graph) |
| AC-1.6 | The repo after this PR; `rg -n 'SoleGraphTx' core api` | Matches only `core/database/evidencegraphs/evidencegraphs.go` (the definition and `SoleGraph`'s call), and `_test.go` files. `rg -n 'evidencegraphs\.SoleGraph\(' core api` matches only `core/database/promotealign/promotealign.go` and `_test.go` files |

## 3. Scope

| In | Covered by |
| --- | --- |
| `subjects.CreateInput.SourceID` deleted; `GraphID` required; `InsertTx` reads the Source with `evidencegraphs.SourceOf`; the `SoleGraphTx` fallback, `requireSource` and `sqlSourceExists` deleted (D-1, D-2) | STEP-1, AC-1.1 – AC-1.4 |
| `connect.CreateCitedBridge` passes only `GraphID` to `subjects.InsertTx` (D-3) | STEP-1, AC-1.5 |
| `evidencegraphs.SoleGraphTx` doc comment names only Promote (S12-09) (D-4) | STEP-1, AC-1.6 |
| Every test that builds a `subjects.CreateInput` passes a graph id (D-5 – D-9) | STEP-2, STEP-3, STEP-4 |
| `artifactstest.SubjectOn` drops its dead `sourceID` parameter (D-6) | STEP-2 |
| `interpretation-layer-data-model.md` and `source-layer-data-model.md` stop naming the fallback (D-10) | STEP-5 |

| Out | Home |
| --- | --- |
| Promote's `SoleGraph` use; deleting `SoleGraph`, `SoleGraphTx` and their tests in `evidencegraphs_test.go` and `core/database/audit/scopes_test.go` | S12-09 |
| Dropping the `subjects.source_id` column or `Subject.SourceID` | not planned (working rule "`subjects.source_id` stays") |
| Folding the ~25 test helpers that copy a Subject create into one shared fixture (`subjectstest`), moving `SubjectOn` out of `artifactstest`, one placement write per test Subject | [`ideas/test-subject-fixtures.md`](../../../ideas/test-subject-fixtures.md) (user, 2026-10-24) |
| Any RPC, proto, FakeStore or Swift change | not planned: `CreateSubjectRequest` already carries `graph_id` (S12-08) |
| `evidencegraphstest.SoleGraph` deletion | not planned (S12-08 D-15) |

## 4. Current state

Verified at `c93b3fb`.

| Area | Where | Today |
| --- | --- | --- |
| Create input | `core/database/subjects/subjects.go` `CreateInput` | Fields `SourceID`, `GraphID`, `SubjectTypeID`, `Label`, `Description`. The `GraphID` comment says "Empty means the Source's sole graph (SoleGraphTx)" |
| Insert rule | `core/database/subjects/subjects.go` `InsertTx` | Checks `SubjectTypeID` is 16 bytes; `SourceID` is empty or 16 bytes; a 16-byte `GraphID` with an empty `SourceID` reads the Source with `evidencegraphs.SourceOf` (`evidencegraphs.ErrInvalid` → `ErrInvalid`); an empty `GraphID` needs a 16-byte `SourceID`; then `requireSource(tx, in.SourceID)`; then, for an empty `GraphID`, `evidencegraphs.SoleGraphTx(tx, in.SourceID)`. A `SourceID` that is not the graph's Source is refused by the composite FK (`mapConstraint` → `ErrInvalid`) |
| Source check | `core/database/subjects/subjects.go` `requireSource`, `sqlSourceExists` | `SELECT 1 FROM sources WHERE id = ?`; only `InsertTx` calls it |
| Create | `core/database/subjects/subjects.go` `Create` | `RequireUserID`, then `InsertTx`, then the optional placement. No id checks of its own (S12-08 Landed note) |
| Graph → Source | `core/database/evidencegraphs/evidencegraphs.go` `SourceOf(q database.Querier, graphID []byte)` | `SELECT source_id FROM evidence_graphs WHERE id = ?`; `ErrInvalid` for a nil querier, a malformed id or no row |
| Bridge | `core/database/evidencegraphs/evidencegraphs.go` `SoleGraphTx`, `SoleGraph` | `SoleGraphTx` lists the Source's graphs and returns the one id. Its comment: "Promote's alignment (S12-09) and the empty-GraphID fallback of subjects.InsertTx (S12-18) still read it." Callers outside tests: `SoleGraph` and `subjects.InsertTx`. `SoleGraph` callers outside tests: `core/database/promotealign/promotealign.go` `loadInputs` |
| Product callers | `api/ffi/handlers/subjects.go` `CreateSubject` | Passes `GraphID` only. No change |
| | `core/database/connect/connect.go` `CreateCitedBridge` | `subjects.InsertTx(q, subjects.CreateInput{SourceID: from.SourceID, GraphID: from.GraphID, …})` |
| Test fixture | `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` `SoleGraph(t, c, sourceID) evidencegraphs.Graph` | Reads `evidencegraphs.List(c, sourceID)` and fails the test unless there is exactly one graph. Opens its own read on `c` |
| Test fixture | `core/database/artifacts/artifactstest/artifactstest.go` `SubjectOn(t, c, userID, sourceID, graphID, typeKey, label)` | Passes both ids to `subjects.Create`; 10 callers (`rg -n 'artifactstest\.SubjectOn\(' core api`) |
| One connection | `core/database/catalog.go` `db.SetMaxOpenConns(1)` | A catalog read (`evidencegraphstest.SoleGraph`, `evidencegraphs.List`) issued while a `writes.Run` / `writes.Call` closure or a raw `db.Begin()` transaction is open on the same catalog blocks forever |
| Raw-tx test | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve` | Opens `db.Begin()` after creating Source `base` and `baseArt`; at line 223 creates a Subject with `CreateInput{SourceID: base.ID, …}` inside that transaction |
| Handler test | `api/ffi/handlers/promote_test.go` (around line 330) | Reads `SELECT source_id, subject_type_id FROM subjects WHERE id = ?` and passes `SourceID: sourceID` |
| Internal-package test | `core/database/sources/sources_test.go` (package `sources`, line 749) | Cannot import `evidencegraphstest` (it imports `sources`: import cycle). `sources` already imports `evidencegraphs` |
| Tests of the rule | `core/database/subjects/subjects_test.go` `TestSubjects` cases "create joins the Source's graph", "graph id given", "graph of another source", "graph id wrong length"; `TestInsertReadsTheGraphsSource` map cases "another Source's id", "neither id", "unknown graph" | Cover the fallback, the `SourceID` / graph mismatch and the `SourceOf` path |
| Test literals | `rg -n 'CreateInput\{' --type go` filtered to `subjects.CreateInput` / package `subjects` | 64 `subjects.CreateInput{` literals in `_test.go` files across 40 files, 15 unqualified `CreateInput{` literals in `core/database/subjects/subjects_test.go`, and one in `artifactstest.SubjectOn`. Most set `SourceID` only; a few set both. Deleting the field makes the compiler list every literal that sets `SourceID` |
| Docs | `docs/interpretation-layer-data-model.md` line 587; `docs/source-layer-data-model.md` line 417 | Both say a new Subject takes `GraphID` when given, else `evidencegraphs.SoleGraphTx` |

**Precedent.** S12-08 ([#528](https://github.com/mendahu/provenencia/pull/528)) D-3 moved `CreateSubject` onto `CreateInput{GraphID: …}` with no `SourceID` and put the graph-to-Source read in `InsertTx`; copy its `SourceOf` error mapping unchanged. Its handler tests took graph ids through a fixture over the graph list (`soleGraphID` in `api/ffi/handlers/evidence_graphs_test.go`); this PR's Go tests use `evidencegraphstest.SoleGraph` the same way. Do not copy S12-08's handler-level helpers into `core/`.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `subjects.CreateInput` becomes `{GraphID, SubjectTypeID, Label, Description}`; the `GraphID` comment reads "GraphID is the Evidence graph the Subject joins. Required; the Subject's Source is the graph's Source." `InsertTx` order: trim; `len(SubjectTypeID) != 16` → `ErrInvalid`; `len(GraphID) != 16` → `ErrInvalid`; `sourceID, err := evidencegraphs.SourceOf(tx, in.GraphID)` (`evidencegraphs.ErrInvalid` → `ErrInvalid`, other errors returned); `requireTypePrefix`; insert with `sourceID` and `in.GraphID`. The returned `Subject.SourceID` and the change's `source_id` field come from `sourceID` | Working rule "High tolerance for churn": a field the engine derives is replaced, not kept beside. One rule, one place (S12-08 D-3) | Keeping `SourceID` as an optional cross-check — it can only disagree with the graph, and the FK already guards the column; making it a second required field — two ids for one fact |
| D-2 | Delete `requireSource` and `sqlSourceExists` | `SourceOf` reads the graph row; `evidence_graphs.source_id` is a NOT NULL FK, so a graph that exists proves its Source exists. Nothing else calls them | Keeping `requireSource(tx, sourceID)` after `SourceOf` — a read that cannot fail |
| D-3 | `connect.CreateCitedBridge` passes `subjects.CreateInput{GraphID: from.GraphID, SubjectTypeID: …, Description: …}` | The bridge's graph comes from its endpoints (S12-06); its Source follows from the graph | Reading `from.SourceID` to cross-check — `requireSameGraph` already pins both endpoints to one graph |
| D-4 | `SoleGraphTx` and `SoleGraph` stay exported and unchanged in code. `SoleGraphTx`'s whole doc comment becomes, verbatim: "// SoleGraphTx returns the id of the one graph of a Source.\n//\n// The one bridge left: Promote's alignment reads it through SoleGraph until\n// S12-09, which deletes both." `SoleGraph`'s comment stays. `TestSoleGraphTx` in `evidencegraphs_test.go` and `core/database/audit/scopes_test.go`'s use stay | S12-09 deletes them (map row S12-09); unexporting or folding them now is churn S12-09 undoes | Folding `SoleGraphTx` into `SoleGraph` — touches `scopes_test.go` and `TestSoleGraphTx`, which S12-09 deletes anyway |
| D-5 | A test that builds a Subject on a Source with one graph reads the graph with `evidencegraphstest.SoleGraph(t, c, sourceID).ID` and passes it as `GraphID`. The read runs **outside and before** the `writes.Run` / `writes.Call` / `db.Begin()` that creates the Subject, never inside its closure (one connection: §4 "One connection"). Three shapes: (a) **a fixture struct with one fixed Source** — `batchFixture` (`core/database/promote/batch_test.go`, built in `newBatchFixture`), `bridgeWorld` (`core/database/promote/bridges_test.go`, `newBridgeWorld`), `fixture` in `core/database/canonicalgraph/canonicalgraph_test.go` (`newFixture`), `fixture` in `core/database/conclusionheaders/conclusionheaders_test.go` (used by `walks_test.go` too), `handleFixture` (`core/search/handles_test.go`, `newHandleFixture`) — gains a field `graph []byte` *(new)*, set once where the fixture is built, right after its Source; its helpers (`bare`, `node`, `subject`) pass `f.graph` and read nothing. A conclusionheaders literal built from `art.SourceID` passes `f.graph` when `art` is on `f.source`, else reads `evidencegraphstest.SoleGraph(t, f.c, art.SourceID).ID` before its write. (b) **a helper or closure that builds a Subject from a Source id** it takes as a parameter (`(*fixture).bare(src, kind)` in promotealign, `(*fixture).subjectOn(src, kind)` in autoreconciler, `mustSubject` in identityclaims, the `mustPerson` closure in subjectpositions) or closes over (the `mk` / `mkSubject` / `newSubject` closures) keeps its signature and reads the graph on its first line, before its write. (c) **a test that already holds a graph** (from `artifactstest.Split`, `evidencegraphstest.Split` / `Create`, or a Subject's `GraphID`) passes it and adds no read. In every file, remove an import, helper or variable the change leaves unused, and nothing else | The map's scope entry names `evidencegraphstest.SoleGraph` and `artifactstest.Split`. A fixed-Source fixture stores its graph once, so S12-09, which makes Promote per graph, changes one field in each Promote fixture instead of a read hidden in each helper (review M-2) | Folding the ~25 copied create helpers into one shared fixture — user, 2026-10-24: keep this PR mechanical; logged in [`ideas/test-subject-fixtures.md`](../../../ideas/test-subject-fixtures.md) (review M-1, M-3, M-4); a new helper `artifactstest.SubjectOnSource` — a second fixture beside `SubjectOn` that hides the same `SoleGraph` read |
| D-6 | `artifactstest.SubjectOn` becomes `SubjectOn(t testing.TB, c *database.Catalog, userID, graphID []byte, typeKey, label string) subjects.Subject`; its 10 callers drop the Source argument | The `sourceID` parameter is dead once `CreateInput.SourceID` goes | Keeping an ignored parameter |
| D-7 | `core/database/sources/sources_test.go` (package `sources`, import cycle with `evidencegraphstest`) reads `graphs, err := evidencegraphs.List(c, uncited.ID)` between `runCreate` and the `writes.Call` that creates the Subject, calls `t.Fatalf("source %x has %d graphs, want 1", uncited.ID, len(graphs))` unless `err == nil && len(graphs) == 1`, and passes `graphs[0].ID` | `sources` already imports `evidencegraphs`; `List` is the one graph read (working rule "One graph list") | Moving the test to `sources_test` — a test move this PR does not need |
| D-8 | `TestEvidenceGraphsResolve` reads `baseGraph := evidencegraphstest.SoleGraph(t, c, base.ID).ID` right after `baseArt` is created, before `db.Begin()`, and the Subject at line 223 uses `GraphID: baseGraph` | A catalog read inside the open raw transaction blocks (one connection) | Taking the graph from `base`'s create changes — `writes.Run` returns no changes to the test |
| D-9 | `api/ffi/handlers/promote_test.go` reads `SELECT graph_id, subject_type_id FROM subjects WHERE id = ?` and passes `GraphID` | The row already holds the graph; no fixture read | `soleGraphID(t, dir, sourceID)` — it opens the project from `dir` while `withProjectCatalog` holds it |
| D-10 | `subjects_test.go`: delete case "create joins the Source's graph" (the behavior is gone; "graph id given" covers a created Subject's graph) and case "graph of another source" (a mismatch cannot be expressed). "graph id given" and "graph id wrong length" drop `SourceID`. In `TestInsertReadsTheGraphsSource` delete "another Source's id" and with it `other` (`src, other := newSource("Deed"), newSource("Other")` becomes `src := newSource("Deed")`; the `newSource` closure stays), rename "neither id" to "no graph id", keep "unknown graph", and add "graph id wrong length" (`GraphID: []byte{1, 2, 3}`). Every other `CreateInput{` literal in `subjects_test.go` and `delete_test.go` follows D-5. Remove any import, helper (`mustSource`, `mustSubjectTypes`) or variable the deletions leave unused, and nothing else | Tests follow the rule: each AC-1.x has a case; no test for a removed input | Keeping the mismatch tests through raw SQL — the composite FK is S12-03's schema, and after this PR no Go input can express a mismatch |
| D-11 | Docs: `interpretation-layer-data-model.md` line 587's last sentence becomes "`subjects.Create` takes the graph in `CreateInput.GraphID` (required) and reads the Subject's Source from it." `source-layer-data-model.md` line 417 becomes "A new Subject joins the graph in `subjects.CreateInput.GraphID`, which is required; its Source is the graph's Source." | Working rule "Each PR updates the docs and skills it makes false" | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| No production code — this PR adds no file, type, function, table, FFI method or helper outside tests. It deletes a field, a fallback and a check | — | `evidencegraphs.SourceOf` | Reused as it is |
| `graph []byte` field on five fixed-Source test fixtures *(new)* (D-5 a) | test fixture field | `evidencegraphstest.SoleGraph`, `artifactstest.SubjectOn` | Each fixture already stores its `source`; the graph sits beside it, read once with `SoleGraph` |

**Builds on.** S12-08 D-3 (`InsertTx` holds the rule; `SourceOf` is the one graph-to-Source read), S12-08 D-15 (`evidencegraphstest.SoleGraph` is the test fixture), working rule "One bridge, one function".
**Deletes.** `subjects.CreateInput.SourceID`; `InsertTx`'s empty-`GraphID` branch and its `SoleGraphTx` call; `subjects.requireSource`, `sqlSourceExists`; `artifactstest.SubjectOn`'s `sourceID` parameter; the `subjects_test.go` cases named in D-10; the fallback sentences in the two domain docs.
**Next adds.** S12-09 changes `promotealign.loadInputs` to a graph id and deletes `SoleGraph`, `SoleGraphTx`, `TestSoleGraphTx` and the `core/database/audit/scopes_test.go` use; after this PR those are the only callers left.

### Schema

None — no catalog change. The composite FK `(graph_id, source_id)` → `evidence_graphs (id, source_id)` stays.

### Go core

| Package | Change |
| --- | --- |
| `core/database/subjects` | `CreateInput` per D-1; `InsertTx` per D-1; `requireSource`, `sqlSourceExists` deleted (D-2). `Create` unchanged |
| `core/database/connect` | `CreateCitedBridge` per D-3 |
| `core/database/evidencegraphs` | `SoleGraphTx` doc comment per D-4; no code change |
| `core/database/artifacts/artifactstest` | `SubjectOn` signature per D-6 |

```go
// CreateInput is the mutable fields for a new Subject.
type CreateInput struct {
	// GraphID is the Evidence graph the Subject joins. Required; the
	// Subject's Source is the graph's Source.
	GraphID       []byte
	SubjectTypeID []byte
	Label         string
	Description   string
}
```

Errors: `subjects.invalid` (`ErrInvalid`, user kind) → a missing, malformed or unknown graph id, as today for a malformed or unknown one.

### FFI

None — `CreateSubjectRequest` already carries `graph_id` and the handler already sends `GraphID` only (S12-08).

### macOS

None — Go only.

### Strings

None — no copy changes; `subjects.invalid` already has its message.

## 7. Steps

**STEP-1 — A Subject needs its graph**
Files: `core/database/subjects/subjects.go`, `core/database/connect/connect.go`, `core/database/evidencegraphs/evidencegraphs.go`. Skill: `add-catalog-query`.
Do: D-1, D-2, D-3, D-4.
Check: `CGO_ENABLED=1 go build -tags fts5 ./core/... ./api/...` passes (non-test code). `rg -n 'SoleGraphTx' core api -g '!*_test.go'` lists only `core/database/evidencegraphs/evidencegraphs.go`.

**STEP-2 — Test fixtures and the subjects tests**
Files: `core/database/artifacts/artifactstest/artifactstest.go`, `core/database/subjects/subjects_test.go`, `core/database/subjects/delete_test.go`.
Do: D-6 (the function only; its 10 callers move in STEP-3); D-10; the other literals in those two subjects test files per D-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/... ./core/database/artifacts/...` passes.

**STEP-3 — Core test literals**
Files: every `_test.go` under `core/` that `CGO_ENABLED=1 go vet -tags fts5 ./core/...` reports for an unknown field `SourceID` in `subjects.CreateInput` or for too many arguments in a call to `artifactstest.SubjectOn` (`rg -n 'artifactstest\.SubjectOn\(' core api` lists the 10 calls), including `core/database/effects/evidence_graphs_test.go` (D-8) and `core/database/sources/sources_test.go` (D-7). The list at the planning base: `core/writes/run_test.go`, `core/search/handles_test.go`, and `core/database/audit/scopes_test.go`, `core/database/autoreconciler/autoreconciler_test.go`, `core/database/canonicalgraph/canonicalgraph_test.go`, `core/database/citations/citations_test.go`, `core/database/citations/delete_test.go`, `core/database/conclusiondetails/conclusiondetails_test.go`, `core/database/conclusionheaders/conclusionheaders_test.go`, `core/database/conclusionheaders/walks_test.go`, `core/database/connect/connect_test.go`, `core/database/deleteimpact/conclusion_release_test.go`, `core/database/deleteimpact/conclusion_test.go`, `core/database/deleteimpact/impact_test.go`, `core/database/deleteimpact/pins_e2e_test.go`, `core/database/effects/conclusion_handles_test.go`, `core/database/effects/evidence_graphs_test.go`, `core/database/effects/evidence_handles_test.go`, `core/database/effects/handle_paths_test.go`, `core/database/evidencegraphs/evidencegraphs_test.go`, `core/database/graphcache/evidence_graph_walk_test.go`, `core/database/grapheventtitles/grapheventtitles_test.go`, `core/database/graphfill/load_test.go`, `core/database/graphprogress/graphprogress_test.go`, `core/database/identityclaims/identityclaims_test.go`, `core/database/likenesses/likenesses_test.go`, `core/database/observations/observations_test.go`, `core/database/promote/batch_test.go`, `core/database/promote/bridges_test.go`, `core/database/promote/pins_test.go`, `core/database/promote/promote_test.go`, `core/database/promotealign/promotealign_test.go`, `core/database/properties/properties_delete_test.go`, `core/database/sourcecopy/sourcecopy_test.go`, `core/database/sources/sources_test.go`, `core/database/subjectpositions/subjectpositions_test.go`.
Do: per D-5, D-6 (callers), D-7, D-8. A literal that sets both `SourceID` and `GraphID` drops `SourceID`. A literal that sets neither id is not reported by `go vet`; it fails `go test` with `subjects.invalid` and is fixed the same way. Change only the literal, the graph read D-5 places, and what D-5 says to remove.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./core/...` reports nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes.

**STEP-4 — Handler test literals**
Files: `api/ffi/handlers/conclusion_headers_test.go`, `api/ffi/handlers/promote_align_test.go`, `api/ffi/handlers/promote_test.go`.
Do: add the `evidencegraphstest` import to `conclusion_headers_test.go` and `promote_align_test.go`. In `conclusion_headers_test.go` the creates at lines 234 and 342 run inside `withProjectCatalog(dir, func(c *database.Catalog) error {…})` (lines 228 and 336); read `graph := evidencegraphstest.SoleGraph(t, c, source[:]).ID` as the first line in that closure after the `subjecttypes.Lookup`, before `writes.Call`, and pass `GraphID: graph`. In `promote_align_test.go` the create at line 49 runs inside the `withProjectCatalog` closure that starts at line 33, after `src` is created; read `evidencegraphstest.SoleGraph(t, c, src.ID).ID` between the `sources.Create` write and the `subjects.Create` write. `promote_test.go` per D-9.
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/...` reports nothing; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes. `rg -n 'SoleGraphTx' core api` and `rg -n 'evidencegraphs\.SoleGraph\(' core api` match AC-1.6.

**STEP-5 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status, Landed note).
Do: D-11; map row status `Landed #N`; a `## S12-18 — A Subject needs its graph` paragraph at the top of `completed.md` (newest first, above S12-08), in S12-08's format (**Landed** date and PR link, what shipped, **Not this PR.**); per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-18-subject-needs-graph.md` reports no errors; `rg -n 'SoleGraphTx' docs -g '!docs/deployment-plan/**'` prints nothing.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/subjects/subjects_test.go` `TestInsertReadsTheGraphsSource` | `{GraphID: jones}` → Subject on Jones of Deed; "no graph id" `{}` → `ErrInvalid`; "unknown graph" → `ErrInvalid`; "graph id wrong length" *(new case)* `{GraphID: {1,2,3}}` → `ErrInvalid` | AC-1.1, AC-1.2, AC-1.3, AC-1.4, D-1, D-10 |
| T-2 | Go | `core/database/subjects/subjects_test.go` `TestSubjects` "graph id given", "graph id wrong length" | Jones graph id only → created and read back on Jones; `{1,2,3}` → `ErrInvalid` | AC-1.1, AC-1.4, D-10 |
| T-3 | Go | `core/database/connect/connect_test.go` `TestCreateCitedBridge` "bridge joins its endpoints' graph" | Bridge between Jones Subjects → `GraphID` Jones, `SourceID` the seed Source | AC-1.5, D-3 |
| T-4 | Go | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve` | Subject on `base`'s graph inside the raw tx → test completes (no hang) and "No other write flags the list" holds | D-8 |
| T-5 | Repo | `rg -n 'SoleGraphTx' core api`; `rg -n 'evidencegraphs\.SoleGraph\(' core api` | Output as AC-1.6 | AC-1.6, D-4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/subjects`, `core/database/effects`, `core/database/deleteimpact`, `core/database/promote`, and any test that hangs because a graph read landed inside a write closure (D-5).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy changes
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none: no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: interpretation-layer-data-model.md (subjects.graph_id), source-layer-data-model.md § 7.1 (D-11)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a Source that has an Artifact. Open its Evidence graph and add a Person from the canvas. The Person appears on the canvas, as on `main`.
2. Connect two Subjects on that graph with a relationship bridge. The bridge appears between them, as on `main`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A graph read placed inside a `writes.Run` / `writes.Call` closure or an open raw transaction hangs the test (one connection) | D-5, D-8: every read runs before the write. A test that hangs for 10 minutes is this; move the read out |
| A test Source has two graphs when `evidencegraphstest.SoleGraph` runs, so it fails the test | Use the graph the test already holds (D-5); the fixture failing names the Source |
| The compiler list misses a literal that sets neither id (`CreateInput{SubjectTypeID: …}`) | It now fails at run time with `subjects.invalid`; STEP-3 and STEP-4 run every package's tests |

**Stop and report if:**

- `subjects.InsertTx` or `CreateInput` no longer matches §4 (a field, branch or call was added after `c93b3fb`).
- A non-test caller of `subjects.Create` or `subjects.InsertTx` other than `api/ffi/handlers/subjects.go` and `core/database/connect/connect.go` exists at your base.
- A test can only get its Subject's graph by calling `evidencegraphs.SoleGraphTx` or `SoleGraph` from non-test code, or by adding a new helper (D-5 rejects both).
- `go vet` reports an import cycle for a test package other than `core/database/sources` (D-7 covers that one).
- Removing `requireSource` changes any test's expected error (D-2 says it cannot).

## 12. Open questions

None.

## 13. PR

**Title:** S12-18 — A Subject needs its graph
**Body:** objective in one sentence; what shipped against US-1; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-18-subject-needs-graph.md`.
