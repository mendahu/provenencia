# S12-06 — Graph isolation

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 2 — Graphs in the catalog
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: write-time checks in Go, one walk join, and one error string)
**Board:** —
**Depends on:** S12-03 (landed [#498](https://github.com/mendahu/provenencia/pull/498)); tests use S12-04's `evidencegraphs.Create` (landed [#505](https://github.com/mendahu/provenencia/pull/505))
**Blocks:** —
**Branch:** `s12-06-graph-isolation`, branched from the commit that adds this plan (so STEP-8 edits it in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-05 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [rules](../source-structure.md#rules) ("Graphs are fully isolated"), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) § 4 (subjects) and § 5 (`mentions`); working rules: "Graphs are fully isolated", "One bridge, one function, deleted in S12-08", "`core/database/evidencegraphs` owns graphs and links from S12-03", "High tolerance for churn", "Each PR updates the docs and skills it makes false" (S12-06 line)
**Planned:** 2026-10-21 against `main` @ `9ca7472`
**Review:** — (not reviewed yet)

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

Each Evidence graph becomes a closed box at write time. An Observation is saved only when its Subject, the Subject it points at, and its Citation's Artifact all sit in, or are linked to, one graph. Anything else is refused with its own user-kind code, `observations.other_graph`. Today only the Subject's Source is checked, and the Subject an Observation points at is not checked at all (`observations.go`: "Value subjects are not checked"). A cited bridge joins its endpoints' graph instead of asking for "the Source's graph", and the Evidence graph walk stops at the graph's edge.

**Why now.** Slice 4 lets researchers make a second graph. Before that ships, Go must refuse a relationship that crosses graphs (dogfood bar 3). Every Source still has one graph, so nothing a researcher does today changes. The tests make second graphs with S12-04's `evidencegraphs.Create`.

**What stays.** The canvas, composer and Connect look and behave as before. `canonicalgraph.WalkSource` keeps its signature and its callers (S12-07 and S12-09 re-key them). A `mentions` Observation may still point at a `source` Subject of another Source (D-3). The `CreateCitedBridgeRequest` wire is unchanged.

## 2. User stories

**US-1.** As a researcher, when I record an Observation, the engine refuses it unless its Subject, the Subject it points at, and its Citation's Artifact are in, or linked to, one Evidence graph. Then a graph's evidence never leaks into another graph.

Fixture for every AC: Source **Census** with Artifacts **P1** and **P2**. Graph **Default** is linked to P1 only. Graph **Jones** is linked to P2 only. Subjects **Ann** (person, Default), **Ben** (person, Jones), **Wed** (event, Default). Source **Book** has graph Default, Artifact **B1**, Subject **Cert** (type `source`) and Subject **Cal** (person).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Citation on P1, Observation `name` on Ann | Saved |
| AC-1.2 | Citation on P1, Observation `name` on Ben | Refused `observations.other_graph`; no Citation or Observation row is left |
| AC-1.3 | Citation on P2, Observation `name` on Ann | Refused `observations.other_graph` |
| AC-1.4 | Citation on P1, Observation `name` on Cal (another Source) | Refused `observations.other_graph` (today `observations.invalid`) |
| AC-1.5 | Citation on P1, an edge Observation (`AllowEdgeRows`) on a Default `participation` bridge with `person` = Ben | Refused `observations.other_graph` |
| AC-1.6 | Citation on P1, the same edge Observation with `person` = Ann | Saved |
| AC-1.7 | Citation on Census's P1, a Census `source` Subject, Observation `mentions` = Cert (Book's graph) | Saved (D-3) |
| AC-1.8 | Citation on P1, Observation `mentions` = a `person` Subject of another graph | Refused `observations.other_graph` (only `source`-typed values are exempt) |
| AC-1.9 | An existing `name` Observation on Ann (Citation on P1) is updated so its Subject is Ben | Refused `observations.other_graph`; the stored row is unchanged |
| AC-1.10 | The same Observation is updated with a new name, Subject still Ann | Saved |

**US-2.** As a researcher, when I connect two Subjects, the bridge joins their graph, and a connection across graphs is refused.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Census has two graphs. Connect Ben (Jones) to a Jones event, new Citation on P2 | Saved. The bridge Subject's `GraphID` is Jones and its `SourceID` is Census. `SoleGraphTx` is not called (it would fail with two graphs) |
| AC-2.2 | Connect Ann (Default) to a Jones event, new Citation on P1 | Refused `observations.other_graph`. No bridge Subject, position, Citation or Observation row is left |
| AC-2.3 | Connect Ann to Wed, attaching an existing Citation on P2 | Refused `observations.other_graph` |
| AC-2.4 | Every existing `TestCreateCitedBridge` case on a one-graph Source | Passes unchanged, apart from removing the `SourceID:` input line |

**US-3.** As the engine, the Evidence graph walk only follows edges inside one graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `WalkSource` from Birth over `PlacesOfEvent`, where the location bridge's place York was moved to a second graph of the same Source by direct SQL (a stored row from before this PR) | No edge |
| AC-3.2 | The walk's `EXPLAIN QUERY PLAN` | No `SCAN` (the existing index test stays green) |

**US-4.** As a researcher, an isolation refusal reads as its reason.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | `L10n.Errors.message(code: "observations.other_graph")` | "That subject or artifact is in another evidence graph." |

## 3. Scope

| In | Covered by |
| --- | --- |
| `apperr.CodeObservationsOtherGraph` and `observations.ErrOtherGraph` (user kind) | STEP-1, AC-1.2 |
| `requireSameGraph` in `observations` replaces `requireSubjectSource` in insert and update. It checks the Subject's graph, the Citation Artifact's link to it, and the Subject value's graph, with the `source`-type exemption | STEP-2, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10 |
| `subjects.CreateInput.GraphID`: when set, the Subject is inserted into that graph; when empty, `SoleGraphTx` as today | STEP-3, AC-2.1 |
| `connect.CreateCitedBridge` drops its Source and Citation-Source checks and `CreateInput.SourceID`. The bridge takes `SourceID` and `GraphID` from the from-Subject | STEP-4, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| The `create_cited_bridge` handler stops parsing `source_id` | STEP-4 |
| `citations.SourceIDTx` is deleted (connect was its only caller) | STEP-4 |
| `canonicalgraph` `sqlSourceSelect` joins `dst` on `graph_id` | STEP-5, AC-3.1, AC-3.2 |
| `L10n.Errors.observationsOtherGraph` and its `Localizable.xcstrings` entry | STEP-6, AC-4.1 |
| `interpretation-layer-data-model.md` isolation text and invariants; map scope for S12-08 | STEP-8 |

| Out | Home |
| --- | --- |
| Removing `source_id` from `CreateCitedBridgeRequest` and sending a graph id from Swift | S12-08 ("cited-bridge RPCs take a graph id") |
| Making `subjects.CreateInput.GraphID` required and deleting its `SoleGraphTx` fallback | S12-08 |
| Re-keying `WalkSource`'s callers (`sourcefill`, `sourceeventtitles`, `promotealign`) and its `sourceID` parameter | S12-07, S12-09 |
| Filtering the composer's Artifacts or the canvas's Subjects by graph | S12-10, S12-08 |
| Repairing stored rows that already cross Sources | not planned (every Source has one graph; R-2) |
| A preflight that lists what is in the other graph | not planned (one plain code, as for the link refusals) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Source check | `core/database/observations/observations.go` `requireSubjectSource` | Joins `subjects` → `citations` → `artifacts` on `source_id` and returns `ErrInvalid` (`observations.invalid`) on no row. It does not check the Subject value |
| Insert | `observations.go` `insertOne` | Calls `requireSubjectSource` right after the length checks, before the binding, edge-lock and `resolveValue` steps |
| Update | `core/database/observations/update.go` `updateOne` | Calls `requireSubjectSource(tx.Tx, …)` at the same point |
| Subject value | `observations.go` `resolveValue`, `case propertymeta.ValueTypeSubject` | Checks only that `in.ValueSubjectID` exists (`sqlSubjectExists`) and returns it as `resolvedValue.SubjectID` |
| Codes | `core/apperr/apperr.go` `CodeObservationsInvalid`, `CodeObservationsEdgeLocked`, `CodeObservationsInUse` | No isolation code |
| Membership reads | `core/database/evidencegraphs/evidencegraphs.go` `SubjectGraphTx`, `IsLinkedTx` | Exist, with no callers outside the package's tests. S12-03 added them for this PR |
| Second graphs | `evidencegraphs.go` `Create` | Creates a graph linked to given Artifacts. `Unlink` refuses a graph's last link |
| Subject insert | `core/database/subjects/subjects.go` `CreateInput`, `InsertTx` | `CreateInput` has no graph. `InsertTx` calls `evidencegraphs.SoleGraphTx(tx, in.SourceID)`, which fails when a Source has two graphs |
| Composite FK | `core/database/migrations/000055.sql` | `subjects (graph_id, source_id)` → `evidence_graphs (id, source_id)`; `foreign_keys` is on (`core/database/catalog.go`) |
| Connect | `core/database/connect/connect.go` `CreateCitedBridge` | Refuses unless `from.SourceID` and `to.SourceID` equal `in.SourceID`. For an attached Citation it also refuses unless `citations.SourceIDTx` equals `in.SourceID`. The bridge is inserted with `subjects.InsertTx(q, CreateInput{SourceID: in.SourceID, …})` |
| Connect input | `connect.go` `CreateInput.SourceID` | Read only by the two checks above and the bridge insert |
| Connect callers | `api/ffi/handlers/connect.go` `CreateCitedBridge`; 30 `CreateInput{` literals: 21 `connect.CreateInput{` across 16 files (`rg -l 'connect.CreateInput\{'`) and 9 in `connect_test.go` | The handler parses `req.GetSourceId()` and passes it as `SourceID` |
| Citation Source | `core/database/citations/citations.go` `SourceIDTx` | Only caller is `connect.go`. `SourceID` (catalog form, used by `api/ffi/handlers/citations.go`) and `sqlCitationSource` stay |
| Citation Artifact | `citations.go` `Update`, `CitationFieldsInput` | A Citation's `artifact_id` never changes after insert, so checking at Observation write is enough |
| Walk | `core/database/canonicalgraph/canonicalgraph.go` `sqlSourceSelect` | `JOIN subjects dst ON dst.id = b.value_subject_id AND dst.source_id = bs.source_id` |
| `mentions` | `core/database/subjectvocab/registry.go` | `mentions` (value type subject, "Target hint: source") is bound to type `source`. `interpretation-layer-data-model.md` § 5 documents `BookSource -- mentions --> CertificateSource` across Sources |
| Doc | `docs/interpretation-layer-data-model.md` (subjects `source_id` paragraph, `graph_id` paragraph, invariant 3) | Says cross-Source Observations remain valid, and that `subjects.Create` takes the graph only from `SoleGraphTx` |
| Swift copy | `macos/App/Platform/L10n.swift` `L10n.Errors` (`observationsInvalid`, `message(code:)` switch); `macos/App/Resources/Localizable.xcstrings` `error.observations.invalid` | No `other_graph` copy |
| Tests | `core/database/observations/observations_test.go` `TestObservations`; `core/database/connect/connect_test.go` `TestCreateCitedBridge`; `core/database/canonicalgraph/canonicalgraph_test.go` `TestWalkSource`; `core/database/subjects/subjects_test.go` `TestSubjects`; `macos/ProvenenciaTests/L10nResolutionTests.swift` `evidenceGraphErrorsHaveCopy` | No test writes across Sources or graphs |

**Precedent.** S12-05 ([#511](https://github.com/mendahu/provenencia/pull/511)) added the five `evidencegraphs.*` codes end to end. Copy its shape for one code: the `apperr` constant, the `L10n.Errors` resource, the `message(code:)` case, the `xcstrings` entry, and one `L10nResolutionTests` test. `evidencegraphs.Unlink` is the shape for a domain invariant that returns its own user-kind `apperr` error from a `…Tx` read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Refusals return `observations.ErrOtherGraph` = `apperr.New(apperr.CodeObservationsOtherGraph, apperr.KindUser)`, code `observations.other_graph`. Copy: "That subject or artifact is in another evidence graph." | User, 2026-10-21. The reason is clear to the researcher and to tests | Reusing `observations.invalid` — user chose a separate code |
| D-2 | `requireSameGraph(tx *sql.Tx, citationID, subjectID, valueSubjectID []byte) error` *(new)* in `observations.go` is the one isolation check. It reads the Subject's graph (`evidencegraphs.SubjectGraphTx`; a missing Subject is `ErrInvalid`), the Citation's `artifact_id` (no row is `ErrInvalid`), and `evidencegraphs.IsLinkedTx(graph, artifact)` (false is `ErrOtherGraph`). When `valueSubjectID` is non-empty, it reads that Subject's graph and type, and returns `ErrOtherGraph` when the graph differs, unless D-3 applies | One check owns the rule (working rule). The membership reads live in `evidencegraphs` so `observations` does not import `subjects` | A check per write path — two copies; a check in `connect` — the working rule moves it out |
| D-3 | A Subject value whose type is `(key = 'source', origin = 'provenencia')` is exempt from the graph check | User, 2026-10-21. A `source` Subject reifies a Source; `mentions` across Sources is the documented use (§ 5) | No exemption — breaks the documented `mentions` edge |
| D-4 | `insertOne` and `updateOne` call `requireSameGraph(tx, citationID, in.SubjectID, resolved.SubjectID)` right after `resolveValue`, and delete the `requireSubjectSource` call and function. `resolved.SubjectID` is set only for subject-valued Properties, so a stale `in.ValueSubjectID` on another type is never checked | One call checks the Subject, the Artifact and the value together. A refusal after `resolveValue` wrote a date or name row is safe: `writes.Run` rolls back the transaction | Checking before `resolveValue` with `in.ValueSubjectID` — checks a field the Property may not use; two calls — the check splits |
| D-5 | `subjects.CreateInput` gains `GraphID []byte` *(new)*. When it has 16 bytes, `InsertTx` inserts into that graph and the composite FK refuses a graph of another Source (`mapConstraint` → `subjects.ErrInvalid`). When it is empty, `InsertTx` calls `SoleGraphTx` as today. Any other length is `ErrInvalid`. The audit row and the returned `Subject` carry the graph used | The bridge needs its endpoints' graph on a two-graph Source, where `SoleGraphTx` fails. One field is the slot S12-08 makes required. The empty case is the sanctioned `SoleGraphTx` bridge, not a new path | A second insert function `InsertInGraphTx` — a near-copy of `InsertTx`; reading the graph's Source first — the FK already enforces it |
| D-6 | `CreateCitedBridge` deletes both `SourceID` checks and `CreateInput.SourceID`. The bridge is inserted with `SourceID: from.SourceID, GraphID: from.GraphID`. The to-endpoint and the Citation's Artifact are checked by `requireSameGraph` on the edge Observations. The handler stops parsing `req.GetSourceId()`. Every `connect.CreateInput{` literal loses its `SourceID:` line | Working rule: `connect` relies on `requireSameGraph`. A field nothing reads is dead | Keeping `SourceID` as an extra guard — a second scope check, which the rule removes |
| D-7 | Delete `citations.SourceIDTx`. Keep `citations.SourceID` and `sqlCitationSource` | D-6 removes its only caller | Keeping it — dead code |
| D-8 | In `sqlSourceSelect`, `dst.source_id = bs.source_id` becomes `dst.graph_id = bs.graph_id`. `bs.source_id = ?` and the `WalkSource` signature stay | Scope entry: "joins on `graph_id` (no signature change)". `subjects_graph_id_idx` exists; `dst` is read by primary key | Re-keying `WalkSource` to a graph id — S12-07 / S12-09 |
| D-9 | Test fixtures make a second graph with `evidencegraphs.Create` and move a link with `evidencegraphs.Unlink`. They put Subjects into it with `subjects.Create(…, CreateInput{GraphID: …})`. Only AC-3.1 uses direct SQL (`UPDATE subjects SET graph_id = ?`), because the write path now refuses that row | Use the real write paths; direct SQL only for a row the engine can no longer write | A shared fixture helper in `evidencegraphstest` — each test needs a different shape; add one when a third package needs it |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `apperr.CodeObservationsOtherGraph` *(new)*, `observations.ErrOtherGraph` *(new)* | code / var | `CodeObservationsInvalid`, `evidencegraphs.ErrArtifactInUse` | User asked for its own reason (D-1); same shape |
| `requireSameGraph` *(new)* | func | `requireSubjectSource` (deleted), `evidencegraphs.SubjectGraphTx`, `IsLinkedTx` | Replaces the old check and composes the existing membership reads |
| `sqlValueSubjectGraph` *(new)* const in `observations.go` | SQL | `sqlSubjectTypeKey` (key only), `evidencegraphs.sqlSubjectGraph` (graph only) | One row needs the graph, key and origin together; searched `rg "t.key" core/database/observations` |
| `sqlCitationArtifact` *(new)* const in `observations.go` | SQL | `sqlCitationExists`, `citations.GetTx` | `citations` imports `observations`, so `observations` cannot call `citations.GetTx` |
| `subjects.CreateInput.GraphID` *(new)* | field | `artifacts.CreateInput.GraphIDs` | Same idea for one graph; the slot S12-08 makes required |
| `L10n.Errors.observationsOtherGraph` *(new)*, `error.observations.other_graph` *(new)* | L10n key | `L10n.Errors.observationsInvalid` | New reason, same namespace |
| `TestRequireSameGraph` *(new)* | Go test | `TestObservations` table | Its own fixture (two graphs, two Sources); the table's `mustSeed` has one graph |
| `observationsOtherGraphHasCopy` *(new)* | Swift test | `evidenceGraphErrorsHaveCopy` | One code, same shape |

**Builds on.** `evidencegraphs` membership reads (S12-03), `evidencegraphs.Create` / `Unlink` (S12-04), the `apperr` → `L10n.Errors.message(code:)` path (S12-05), `.cursor/rules/go-tests.mdc` tables.
**Deletes.** `observations.requireSubjectSource`; the two `SourceID` checks in `connect.CreateCitedBridge`; `connect.CreateInput.SourceID` and its `SourceID:` line in each of the 30 literals; the handler's `source_id` parse; `citations.SourceIDTx`; the comment "Value subjects are not checked".
**Next adds.** S12-08 makes `subjects.CreateInput.GraphID` required, drops `SoleGraphTx` from `InsertTx`, and removes `CreateCitedBridgeRequest.source_id`. S12-10's composer filter and S12-13's link checkboxes rely on this check as their backstop.

### Schema

None — no catalog change. `subjects_graph_id_idx` and the composite FK exist (migration `000055.sql`).

### Go core

| Package | Change |
| --- | --- |
| `core/apperr` | `CodeObservationsOtherGraph = "observations.other_graph"` *(new)*, beside `CodeObservationsInUse` |
| `core/database/observations` | `ErrOtherGraph` *(new)*. `requireSameGraph` *(new)* per D-2 / D-3, called per D-4. Delete `requireSubjectSource`. Imports `evidencegraphs` and `vocabkey` |
| `core/database/subjects` | `CreateInput.GraphID` *(new)* per D-5, in `InsertTx` and `Create`'s length checks |
| `core/database/connect` | Per D-6 |
| `core/database/citations` | Delete `SourceIDTx` (D-7) |
| `core/database/canonicalgraph` | `sqlSourceSelect` join per D-8; update the comment above it to say "one graph" |
| `api/ffi/handlers` | `connect.go`: drop the `sourceID` parse and field (D-6) |

```go
// requireSameGraph refuses an Observation whose Subject, Subject value, or
// Citation Artifact is not in, or linked to, the Subject's Evidence graph.
// A source-typed Subject value may be in any graph.
func requireSameGraph(tx *sql.Tx, citationID, subjectID, valueSubjectID []byte) error

const sqlCitationArtifact = `SELECT artifact_id FROM citations WHERE id = ?`
const sqlValueSubjectGraph = `SELECT s.graph_id, t.key = 'source' AND t.origin = ?
	FROM subjects s JOIN subject_types t ON t.id = s.subject_type_id
	WHERE s.id = ?` // arg: vocabkey.OriginProvenencia
```

Errors: `observations.other_graph` (user) → Subject not in the Citation Artifact's graphs, or Subject value in another graph (D-3 aside). `observations.invalid` → the Subject, Citation or Subject value does not exist (as today).

### FFI

None — no proto change. `CreateCitedBridgeRequest.source_id` stays on the wire, unread, until S12-08.

### macOS

None — no view change. Only copy (Strings).

### Strings

L10n keys *(new)*: `L10n.Errors.observationsOtherGraph` = "That subject or artifact is in another evidence graph." (`error.observations.other_graph`, comment `FFI error observations.other_graph`), with a `case "observations.other_graph":` in `L10n.Errors.message(code:)`. Removed: none.

## 7. Steps

**STEP-1 — Error code**
Files: `core/apperr/apperr.go`, `core/database/observations/observations.go`.
Do: add `CodeObservationsOtherGraph` and `ErrOtherGraph` (D-1).
Check: `CGO_ENABLED=1 go build ./...` passes.

**STEP-2 — `requireSameGraph`**
Files: `observations.go`, `update.go`, `observations_test.go`. Skill: `add-catalog-query`.
Do: add `requireSameGraph`, `sqlCitationArtifact` and `sqlValueSubjectGraph` (D-2, D-3). In `insertOne` and `updateOne`, remove the `requireSubjectSource` call and call `requireSameGraph` right after `resolveValue` (D-4). Delete `requireSubjectSource`. Add `TestRequireSameGraph` (T-1). Its fixture uses STEP-3's `GraphID`, so write the test file first and land it with STEP-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/...` passes after STEP-3.

**STEP-3 — Subject into a given graph**
Files: `core/database/subjects/subjects.go`, `subjects_test.go`.
Do: `CreateInput.GraphID` per D-5; record `graph_id` from the graph used. Add T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/... ./core/database/observations/...` passes.

**STEP-4 — Connect relies on the check**
Files: `core/database/connect/connect.go`, `connect_test.go`, `api/ffi/handlers/connect.go`, `core/database/citations/citations.go`, every file `rg -l 'connect.CreateInput\{'` lists, and `connect_test.go`.
Do: D-6 and D-7. Add T-2 subtests.
Check: `rg -n 'SourceIDTx|requireSubjectSource' core api` prints nothing. `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-5 — Walk joins on graph**
Files: `core/database/canonicalgraph/canonicalgraph.go`, `canonicalgraph_test.go`.
Do: D-8. Add the T-3 subtest to `TestWalkSource`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/canonicalgraph/... ./core/database/sourcefill/... ./core/database/sourceeventtitles/... ./core/database/promotealign/...` passes.

**STEP-6 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/L10nResolutionTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: per Strings. Add T-5.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; the Xcode test run in `scripts/check-pr.sh` passes `L10nResolutionTests`.

**STEP-7 — Full check**
Files: none new.
Do: run the repo check.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-8 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row, S12-08 scope entry), this plan (status, Landed note).
Do: in `interpretation-layer-data-model.md`, change the `source_id` paragraph and invariant 3 to say an Observation's Subject, Subject value and Citation Artifact share one Evidence graph, except a `source`-typed Subject value. Name `observations.requireSameGraph`. Add an invariant for it after invariant 16. Change the `graph_id` paragraph: `subjects.Create` takes `CreateInput.GraphID` when given, else `SoleGraphTx`. In the S12-08 scope entry, add "makes `subjects.CreateInput.GraphID` (S12-06) required and removes `CreateCitedBridgeRequest.source_id`". Then do the rest of the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-06-graph-isolation.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/observations/observations_test.go` `TestRequireSameGraph` *(new)*, table-driven over the §2 fixture | Each AC-1.x row → saved, or `errors.Is(err, observations.ErrOtherGraph)` with no Citation or Observation row left. AC-1.5 – AC-1.8 insert through `observations.InsertManyTx` with `AllowEdgeRows: true` inside `writes.Call`. AC-1.9 / AC-1.10 use `observationstest.UpdateObservation` and re-read the row | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, D-2, D-3, D-4 |
| T-2 | Go | `core/database/connect/connect_test.go` `TestCreateCitedBridge`: subtests "bridge joins its endpoints' graph", "endpoints in two graphs refused", "attached citation on an unlinked artifact refused" *(new)* | AC-2.1 → `Subject.GraphID` is Jones; AC-2.2 → `ErrOtherGraph` and `countBridges` unchanged; AC-2.3 → `ErrOtherGraph`. Existing subtests stay green | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-6 |
| T-3 | Go | `core/database/canonicalgraph/canonicalgraph_test.go` `TestWalkSource`: subtest "a Subject in another graph is not reached" *(new)* | Second graph made with `evidencegraphs.Create`, York moved into it by `UPDATE subjects SET graph_id`, `PlacesOfEvent` from Birth → 0 edges. Existing index subtest green | AC-3.1, AC-3.2, D-8 |
| T-4 | Go | `core/database/subjects/subjects_test.go` `TestSubjects`: cases "graph id given", "graph of another source", "graph id wrong length" *(new)* | A second graph of the Source → `Subject.GraphID` equals it; another Source's graph → `subjects.ErrInvalid`; 3 bytes → `ErrInvalid` | D-5 |
| T-5 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` `observationsOtherGraphHasCopy` *(new)* | `message(code: "observations.other_graph")` equals the D-1 sentence and is not the unknown copy | AC-4.1, D-1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the tests that build cited bridges (`core/database/autoreconciler`, `core/database/effects`, `core/database/citations/delete_test.go`, `core/database/deleteimpact/impact_test.go`, `core/database/promotealign`, `core/database/sourceeventtitles`, `api/ffi/handlers/connect_test.go`), through the `SourceID` removal.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — n/a, no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — n/a, no brief
- [ ] Domain doc updated if a rule in docs/*.md changed: interpretation-layer-data-model.md (STEP-8); S12-08 scope entry in deployment-plan.md
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open an existing project. On a Source's Evidence graph, add a Citation with a name Observation, and connect a person to an event. Both save as before.
2. Promote that Source and open a canonical person's page. The events and places shown are the same as before the PR.

Two-graph refusals cannot be reached from the app until slice 4. T-1 to T-3 cover them.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| R-1. A caller saves Observations whose Subject value is in another Source today, and it starts failing | Every Source has one graph, so this only refuses rows that already crossed Sources. `rg 'ValueSubjectID' core api --glob '!*_test.go'` callers are the composer, Connect and Promote. Promote writes no Observations |
| R-2. A stored row already crosses Sources and can no longer be updated | Accepted. The researcher deletes and re-records it. Not planned |
| R-3. A test that builds a bridge across two Sources breaks | It now fails with `ErrOtherGraph`. Fix the fixture to use one Source; do not weaken the check |

**Stop and report if:**

- `requireSubjectSource` is called from anywhere other than `insertOne` and `updateOne`.
- `observations` importing `evidencegraphs` creates an import cycle (`go build` reports `import cycle not allowed`).
- Some code other than `connect.go` calls `citations.SourceIDTx`.
- An existing test or caller writes a `mentions` value of a non-`source` type, or any other Subject value, across Sources on purpose (a fixture comment or doc says so).
- The composite FK does not refuse a `GraphID` of another Source (T-4's second case saves a row).
- `sqlSourceSelect` with the `graph_id` join shows a `SCAN` in T-3's index subtest.

## 12. Open questions

None.

## 13. PR

**Title:** S12-06 — Graph isolation
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-06-graph-isolation.md`.
