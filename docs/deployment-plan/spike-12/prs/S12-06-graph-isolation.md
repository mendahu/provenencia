# S12-06 — Graph isolation

**Status:** Planned
**Spike / slice:** Spike 12 · Slice 2 — Graphs in the catalog
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: write-time checks in Go and one error string)
**Board:** —
**Depends on:** S12-03 (landed [#498](https://github.com/mendahu/provenencia/pull/498)); tests use S12-04's `evidencegraphs.Create` (landed [#505](https://github.com/mendahu/provenencia/pull/505))
**Blocks:** —
**Branch:** `s12-06-graph-isolation`, branched from the commit that adds this plan (so STEP-7 edits it in the same branch)
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-05 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [rules](../source-structure.md#rules) ("Graphs are fully isolated"), [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md) § 4 (subjects), § 5 (`mentions`) and § 8 (invariants), [`source-layer-data-model.md`](../../../source-layer-data-model.md) § 7.1; working rules: "Graphs are fully isolated", "One bridge, one function, deleted in S12-08", "`core/database/evidencegraphs` owns graphs and links from S12-03", "High tolerance for churn", "Each PR updates the docs and skills it makes false" (S12-06 line)
**Planned:** 2026-10-21 against `main` @ `10d3752` (refreshed after S11-19, [#513](https://github.com/mendahu/provenencia/pull/513))
**Review:** [`S12-06-graph-isolation.review.md`](S12-06-graph-isolation.review.md) — Fits / Ready with changes, 2 rounds; refreshed for S11-19

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

Each Evidence graph becomes a closed box at write time. An Observation is saved only when its Subject, the Subject it points at, and its Citation's Artifact all sit in, or are linked to, one graph. Anything else is refused with its own user-kind code, `observations.other_graph`. Today only the Subject's Source is checked, and the Subject an Observation points at is not checked at all (`observations.go`: "Value subjects are not checked"). A cited bridge joins its endpoints' graph instead of asking for "the Source's graph".

**Why now.** Slice 4 lets researchers make a second graph. Before that ships, Go must refuse a relationship that crosses graphs (dogfood bar 3). Every Source still has one graph, so nothing a researcher does today changes. The tests make second graphs with S12-04's `evidencegraphs.Create`.

**What stays.** The canvas, composer and Connect look and behave as before. The Evidence graph walk is untouched: S11-19 ([#513](https://github.com/mendahu/provenencia/pull/513)) moved it into memory (`graphcache.Graph.WalkSource`, which reaches only Subjects on the loaded `SourceGraph`), and S12-07 makes it per graph. A `mentions` Observation may still point at a `source` Subject of another Source (D-3). The `CreateCitedBridgeRequest` wire is unchanged.

## 2. User stories

**US-1.** As a researcher, when I record an Observation, the engine refuses it unless its Subject, the Subject it points at, and its Citation's Artifact are in, or linked to, one Evidence graph. Then a graph's evidence never leaks into another graph.

Fixture for every AC (built per T-1's recipe). Source **Census** with Artifacts **P1** and **P2**. Census's graph **Default** is linked to P1 only. Census's graph **Jones** is linked to P2 only. Census Subjects: **Ann** (person, Default), **Ben** (person, Jones), **Wed** (event, Default), **Part** (participation, Default), **Rec** (source, Default). Source **Book** has one graph, **Book Default**, with Artifact **B1** and Subjects **Cert** (source) and **Cal** (person).

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Citation on P1, Observation `name` on Ann | Saved |
| AC-1.2 | Citation on P1, Observation `name` on Ben | Refused `observations.other_graph`; no Citation or Observation row is left |
| AC-1.3 | Citation on P2, Observation `name` on Ann | Refused `observations.other_graph` |
| AC-1.4 | Citation on P1, Observation `name` on Cal (Book Default) | Refused `observations.other_graph` (today `observations.invalid`) |
| AC-1.5 | Citation on P1, edge Observation (`AllowEdgeRows`) on Part with `person` = Ben | Refused `observations.other_graph` |
| AC-1.6 | Citation on P1, edge Observation on Part with `person` = Ann | Saved |
| AC-1.7 | Citation on P1, Observation `mentions` on Rec with value Cert (Book Default) | Saved (D-3) |
| AC-1.8 | Citation on P1, Observation `mentions` on Rec with value Cal (a `person` in Book Default) | Refused `observations.other_graph` (only reification-typed values are exempt) |
| AC-1.9 | An existing `name` Observation on Ann (Citation on P1) is updated so its Subject is Ben | Refused `observations.other_graph`; the stored row is unchanged |
| AC-1.10 | The same Observation is updated with a new name, Subject still Ann | Saved |
| AC-1.11 | Citation on P1, edge Observation on Part with `person` = a `source`-typed Subject of another graph | Refused `observations.other_graph` (D-3: edge rows are never exempt) |

**US-2.** As a researcher, when I connect two Subjects, the bridge joins their graph, and a connection across graphs is refused.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Census has two graphs. Connect Ben (Jones) to a Jones event, new Citation on P2 | Saved. The bridge Subject's `GraphID` is Jones and its `SourceID` is Census. `SoleGraphTx` is not called (it would fail with two graphs) |
| AC-2.2 | Connect Ann (Default) to a Jones event, new Citation on P1 | Refused `observations.other_graph`. No bridge Subject, position, Citation or Observation row is left |
| AC-2.3 | Connect Ann to Wed, attaching an existing Citation on P2 | Refused `observations.other_graph` |
| AC-2.4 | Every existing `TestCreateCitedBridge` case on a one-graph Source | Passes unchanged, apart from removing the `SourceID:` input line |

**US-3.** As a researcher, an isolation refusal reads as its reason.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `L10n.Errors.message(code: "observations.other_graph")` | "That subject or artifact is in another evidence graph." |

## 3. Scope

| In | Covered by |
| --- | --- |
| `apperr.CodeObservationsOtherGraph` and `observations.ErrOtherGraph` (user kind) | STEP-1, AC-1.2 |
| `subjects.CreateInput.GraphID`: when set, the Subject is inserted into that graph; when empty, `SoleGraphTx` as today | STEP-2, AC-2.1 |
| `evidencegraphstest.Create` *(new)* test helper; `deleteimpact/impact_test.go`'s inline graph create moves onto it | STEP-2 |
| `requireSameGraph` in `observations` replaces `requireSubjectSource` in insert and update. It checks the Subject's graph, the Citation Artifact's link to it, and the Subject value's graph, with the reification exemption for non-edge rows | STEP-3, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11 |
| `sqlSubjectExists` in `resolveValue` is deleted; the graph read refuses a missing value Subject | STEP-3 |
| `connect.CreateCitedBridge` drops its Source and Citation-Source checks and `CreateInput.SourceID`. The bridge takes `SourceID` and `GraphID` from the from-Subject | STEP-4, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| The `create_cited_bridge` handler stops parsing `source_id` | STEP-4 |
| `citations.SourceIDTx` is deleted (connect was its only caller) | STEP-4 |
| `L10n.Errors.observationsOtherGraph` and its `Localizable.xcstrings` entry | STEP-5, AC-3.1 |
| `interpretation-layer-data-model.md` and `source-layer-data-model.md` § 7.1 isolation text; map scope for S12-08 and the working-rule doc line | STEP-7 |

| Out | Home |
| --- | --- |
| Removing `source_id` from `CreateCitedBridgeRequest` (the bridge's graph comes from its endpoints, so the RPC takes no graph id) | S12-08 |
| Making `subjects.CreateInput.GraphID` required and deleting its `SoleGraphTx` fallback | S12-08 |
| Any Evidence graph walk change. S11-19 ([#513](https://github.com/mendahu/provenencia/pull/513)) deleted the SQL `canonicalgraph.WalkSource`; the in-memory `graphcache.Graph.WalkSource` becomes per graph when `SourceGraph` is keyed by Evidence graph. `canonicalgraph.Walk` reads only canonical rows, which have no graph | S11-19, S12-07 |
| Filtering the composer's Artifacts or the canvas's Subjects by graph | S12-10, S12-08 |
| A `vocabkey` constant for the `source` type key and moving `sourcegraphprogress`'s literals onto it | not planned (D-3 reads the registry role, so this PR adds no literal) |
| Moving `depictionSaveErrorsHaveCopy` into the shared error-copy table | not planned (it checks more than copy) |
| Repairing stored rows that already cross Sources | not planned (every Source has one graph; R-2) |
| A preflight that lists what is in the other graph | not planned (one plain code, as for the link refusals) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Source check | `core/database/observations/observations.go` `requireSubjectSource` | Joins `subjects` → `citations` → `artifacts` on `source_id` and returns `ErrInvalid` (`observations.invalid`) on no row. It does not check the Subject value. Called only from `insertOne` and `updateOne` |
| Insert | `observations.go` `insertOne` | Calls `requireSubjectSource` right after the length checks, before `sqlSubjectType`, the binding check, `sqlSubjectTypeKey` + `connectrules.Edge` (edge lock) and `resolveValue` |
| Update | `core/database/observations/update.go` `updateOne` | Calls `requireSubjectSource(tx.Tx, …)` at the same point. `write.go` `Update` refuses a requested edge row before `updateOne` (`isEdgeTx`), so updates never write edge rows |
| Subject value | `observations.go` `resolveValue`, `case propertymeta.ValueTypeSubject` | Checks only that `in.ValueSubjectID` exists (`sqlSubjectExists`, its only user) and returns it as `resolvedValue.SubjectID` |
| Type key read | `observations.go` `sqlSubjectTypeKey`; users `insertOne` and `write.go` `isEdgeTx` | Selects `t.key` only |
| Codes | `core/apperr/apperr.go` `CodeObservationsInvalid`, `CodeObservationsEdgeLocked`, `CodeObservationsInUse` | No isolation code |
| Membership reads | `core/database/evidencegraphs/evidencegraphs.go` `SubjectGraphTx`, `IsLinkedTx` | Exist, with no callers outside the package's tests. `SubjectGraphTx` returns `evidencegraphs.ErrInvalid` for a missing Subject |
| Second graphs | `evidencegraphs.go` `Create`, `Unlink` | `Create` makes a graph linked to given Artifacts. `Unlink` refuses a graph's last link |
| Reification role | `core/database/subjectvocab/registry.go` `RoleReification`; `subjectvocab.go` `TypeByKey` | The `source` type has `Role: RoleReification`. `TypeByKey` reads the product registry. `subjectvocab` does not import `observations` |
| Subject insert | `core/database/subjects/subjects.go` `CreateInput`, `InsertTx` | `CreateInput` has no graph. `InsertTx` calls `evidencegraphs.SoleGraphTx(tx, in.SourceID)`, which fails when a Source has two graphs |
| Composite FK | `core/database/migrations/000055.sql` | `subjects (graph_id, source_id)` → `evidence_graphs (id, source_id)`; `foreign_keys` is on (`core/database/catalog.go`) |
| Connect | `core/database/connect/connect.go` `CreateCitedBridge` | Refuses unless `from.SourceID` and `to.SourceID` equal `in.SourceID`. For an attached Citation it also refuses unless `citations.SourceIDTx` equals `in.SourceID`. The bridge is inserted with `subjects.InsertTx(q, CreateInput{SourceID: in.SourceID, …})` |
| Connect rules | `core/connectrules/connectrules.go` `productBridges` | No product bridge has a `source` endpoint. Plugins can register bridges |
| Connect callers | `api/ffi/handlers/connect.go` `CreateCitedBridge`; test literals found by `rg -n 'connect.CreateInput\{'` and the unqualified `CreateInput{` literals in `connect_test.go` | The handler parses `req.GetSourceId()` and passes it as `SourceID` |
| Observation writers | `rg -l 'observations.InsertManyTx\|InsertWithObservationsTx\|observations.AddToCitation\|citations.CreateWithObservations\|observations.Update\(' core api --glob '!*_test.go'` | `citations`, `connect`, `api/ffi/handlers/citations.go`, `api/ffi/handlers/observations.go`, `observationstest`. Promote writes no Observations |
| Citation Source | `core/database/citations/citations.go` `SourceIDTx` | Only caller is `connect.go`. `SourceID` (catalog form, used by `api/ffi/handlers/citations.go`) and `sqlCitationSource` stay |
| Citation Artifact | `citations.go` `Update`, `CitationFieldsInput` | A Citation's `artifact_id` never changes after insert, so checking at Observation write is enough |
| `mentions` | `core/database/subjectvocab/registry.go` | `mentions` (value type subject, "Target hint: source") is bound to type `source`. `interpretation-layer-data-model.md` § 5 documents `BookSource -- mentions --> CertificateSource` across Sources |
| Docs | `docs/interpretation-layer-data-model.md` (subjects `source_id` paragraph, `graph_id` paragraph, invariants 3 and 16); `docs/source-layer-data-model.md` § 7.1 bullet "A new Subject joins its Source's graph through `evidencegraphs.SoleGraphTx`" | Say cross-Source Observations remain valid, and that a Subject's graph comes only from `SoleGraphTx` |
| Graph fixtures | `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go` `SeedRefless`, `SoleGraph`; `core/database/deleteimpact/impact_test.go` (graph `jones`, inline `writes.Call` → `evidencegraphs.Create`) | No shared helper creates a graph |
| Swift copy | `macos/App/Platform/L10n.swift` `L10n.Errors` (`observationsInvalid`, `message(code:)` switch); `macos/App/Resources/Localizable.xcstrings` `error.observations.invalid`; `macos/ProvenenciaTests/L10nResolutionTests.swift` `evidenceGraphErrorsHaveCopy` (a `[code: sentence]` table) | No `other_graph` copy |
| Tests | `core/database/observations/observations_test.go` `TestObservations`; `core/database/connect/connect_test.go` `TestCreateCitedBridge`; `core/database/subjects/subjects_test.go` `TestSubjects` | No test writes across Sources or graphs, and none expects `ErrInvalid` for a cross-Source write |

**Precedent.** S12-05 ([#511](https://github.com/mendahu/provenencia/pull/511)) added the five `evidencegraphs.*` codes end to end. Copy its shape for one code: the `apperr` constant, the `L10n.Errors` resource, the `message(code:)` case, the `xcstrings` entry, and one row in `L10nResolutionTests`. `evidencegraphs.Unlink` is the shape for a domain invariant that returns its own user-kind `apperr` error from a `…Tx` read.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Refusals return `observations.ErrOtherGraph` = `apperr.New(apperr.CodeObservationsOtherGraph, apperr.KindUser)`, code `observations.other_graph`. Copy: "That subject or artifact is in another evidence graph." | User, 2026-10-21. The reason is clear to the researcher and to tests | Reusing `observations.invalid` — user chose a separate code |
| D-2 | `requireSameGraph(tx *sql.Tx, citationID, subjectID, valueSubjectID []byte, edge bool) error` *(new)* in `observations.go` is the one isolation check. In order: (1) the Subject's graph by `evidencegraphs.SubjectGraphTx`; (2) the Citation's `artifact_id` by `sqlCitationArtifact`; (3) `evidencegraphs.IsLinkedTx(graph, artifact)`, false → `ErrOtherGraph`; (4) when `valueSubjectID` is non-empty, its graph by `SubjectGraphTx`, and when that differs from (1) → `ErrOtherGraph` unless D-3 applies. A missing Subject, Citation or value Subject returns `observations.ErrInvalid`: `errors.Is(err, evidencegraphs.ErrInvalid)` and `sql.ErrNoRows` both map to it | One check owns the rule (working rule). Graph membership is read only through `evidencegraphs`, so `observations` adds no second graph read and does not import `subjects` | A check per write path — two copies; a `graph_id` SQL in `observations` — a second membership read |
| D-3 | A value Subject is exempt from step (4) when `edge` is false and its type is a product reification: origin `vocabkey.OriginProvenencia` and `subjectvocab.TypeByKey(key)` returns `Role == subjectvocab.RoleReification`. `requireSameGraph` runs `sqlSubjectTypeKey` on `valueSubjectID` (a third caller); it now selects `t.key, t.origin`, and `insertOne` and `isEdgeTx` scan and discard `origin`. `sql.ErrNoRows` there maps to `ErrInvalid`. Edge rows (bridge endpoints) are never exempt | User, 2026-10-21: keep the documented `mentions` across Sources. The registry role names the concept, so a later reification type gets it by declaring its role. Edge rows are excluded so a plugin bridge with a `source` endpoint cannot join two graphs | `t.key = 'source'` in SQL — a third copy of the literal; no exemption — breaks `mentions` |
| D-4 | `insertOne` keeps the `connectrules.Edge` result as `edge`. `insertOne` and `updateOne` call `requireSameGraph(tx, citationID, in.SubjectID, resolved.SubjectID, edge)` right after `resolveValue` (`updateOne` passes `false`). Delete the `requireSubjectSource` call and function. Delete `sqlSubjectExists` and its read in `resolveValue`'s subject case | One call checks the Subject, the Artifact and the value together. `resolved.SubjectID` is set only for subject-valued Properties, so a stale `in.ValueSubjectID` on another type is never checked. A refusal after `resolveValue` wrote a date or name row is safe: `writes.Run` rolls back. Because the check now runs later, a write that is both cross-graph and invalid in another way returns the earlier error (`observations.invalid` for an unbound Property, `observations.edge_locked`, or a value error). No existing test changes its expected code | Checking first with `in.ValueSubjectID` — checks a field the Property may not use; two calls — the check splits |
| D-5 | `subjects.CreateInput` gains `GraphID []byte` *(new)*. When it has 16 bytes, `InsertTx` inserts into that graph and the composite FK refuses a graph of another Source (`mapConstraint` → `subjects.ErrInvalid`). When it is empty, `InsertTx` calls `SoleGraphTx` as today. Any other length is `ErrInvalid`, checked in `InsertTx` only (`Create` reaches it). The audit row and the returned `Subject` carry the graph used | The bridge needs its endpoints' graph on a two-graph Source, where `SoleGraphTx` fails. One field is the slot S12-08 makes required. The empty case is the sanctioned `SoleGraphTx` bridge, not a new path. `subjects.invalid` differs from `evidencegraphs.invalid` for an Artifact's wrong graph because each package reports its own input | A second insert function `InsertInGraphTx` — a near-copy of `InsertTx`; reading the graph's Source first — the FK already enforces it |
| D-6 | `CreateCitedBridge` deletes both `SourceID` checks and `CreateInput.SourceID`. The bridge is inserted with `SourceID: from.SourceID, GraphID: from.GraphID`. The to-endpoint and the Citation's Artifact are checked by `requireSameGraph` on the edge Observations. The handler stops parsing `req.GetSourceId()` and gets a one-line comment that `source_id` is unread until S12-08 removes it. Every `connect.CreateInput{` literal, and each `CreateInput{` literal in `connect_test.go`, loses its `SourceID:` line | Working rule: `connect` relies on `requireSameGraph`. A field nothing reads is dead | Keeping `SourceID` as an extra guard — a second scope check, which the rule removes; a proto comment — needs a regenerate for a field S12-08 deletes |
| D-7 | Delete `citations.SourceIDTx`. Keep `citations.SourceID` and `sqlCitationSource` | D-6 removes its only caller | Keeping it — dead code |
| D-8 | `evidencegraphstest.Create(t testing.TB, c *database.Catalog, userID, sourceID []byte, name string, artifactIDs ...[]byte) evidencegraphs.Graph` *(new)* wraps `writes.Call` → `evidencegraphs.Create` and fails the test on error. T-1 to T-3 use it, and `deleteimpact/impact_test.go`'s `jones` create moves onto it. Subjects go into a graph with `subjects.Create(…, CreateInput{GraphID: …})`. No fixture writes `graph_id` by direct SQL | Three packages in this PR create a second graph; one helper, beside `SoleGraph` | Inline creates in each test — five copies |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `apperr.CodeObservationsOtherGraph` *(new)*, `observations.ErrOtherGraph` *(new)* | code / var | `CodeObservationsInvalid`, `evidencegraphs.ErrArtifactInUse` | User asked for its own reason (D-1); same shape |
| `requireSameGraph` *(new)* | func | `requireSubjectSource` (deleted), `evidencegraphs.SubjectGraphTx`, `IsLinkedTx` | Replaces the old check and composes the existing membership reads |
| `sqlCitationArtifact` *(new)* const in `observations.go` | SQL | `sqlCitationExists`, `citations.GetTx` | `citations` imports `observations`, so `observations` cannot call `citations.GetTx` |
| `sqlSubjectTypeKey` gains `t.origin` | SQL (extended) | itself | Extended, not copied; `isEdgeTx` and `insertOne` scan the extra column |
| `subjects.CreateInput.GraphID` *(new)* | field | `artifacts.CreateInput.GraphIDs` | Same idea for one graph; the slot S12-08 makes required |
| `evidencegraphstest.Create` *(new)* | test helper | `evidencegraphstest.SoleGraph`; inline creates in `core/database/deleteimpact/impact_test.go` | Extends the package that holds graph fixtures |
| `L10n.Errors.observationsOtherGraph` *(new)*, `error.observations.other_graph` *(new)* | L10n key | `L10n.Errors.observationsInvalid` | New reason, same namespace |
| `TestObservationsStayInOneGraph` *(new)* | Go test | `TestObservations` table | Its own fixture (two graphs, two Sources); `TestObservations`'s `mustSeed` has one graph |
| `errorCodesHaveCopy` (renamed from `evidenceGraphErrorsHaveCopy`) | Swift test | itself | One row added; the next code is one row |

**Builds on.** `evidencegraphs` membership reads (S12-03), `evidencegraphs.Create` / `Unlink` (S12-04), `subjectvocab` roles, the `apperr` → `L10n.Errors.message(code:)` path (S12-05), `.cursor/rules/go-tests.mdc` tables.
**Deletes.** `observations.requireSubjectSource`; `observations.sqlSubjectExists`; the two `SourceID` checks in `connect.CreateCitedBridge`; `connect.CreateInput.SourceID` and its `SourceID:` line in every literal; the handler's `source_id` parse; `citations.SourceIDTx`; the comment "Value subjects are not checked"; the inline graph create in `core/database/deleteimpact/impact_test.go`.
**Next adds.** S12-08 makes `subjects.CreateInput.GraphID` required, drops `SoleGraphTx` from `InsertTx`, and removes `CreateCitedBridgeRequest.source_id` without adding a graph id. S12-10's composer filter and S12-13's link checkboxes rely on this check as their backstop. S12-10 to S12-13 tests reuse `evidencegraphstest.Create`.

### Schema

None — no catalog change. `subjects_graph_id_idx` and the composite FK exist (migration `000055.sql`).

### Go core

| Package | Change |
| --- | --- |
| `core/apperr` | `CodeObservationsOtherGraph = "observations.other_graph"` *(new)*, beside `CodeObservationsInUse` |
| `core/database/observations` | `ErrOtherGraph` *(new)*. `requireSameGraph` *(new)* per D-2 / D-3, called per D-4. `sqlSubjectTypeKey` selects `t.key, t.origin`. Delete `requireSubjectSource` and `sqlSubjectExists`. New imports: `evidencegraphs`, `subjectvocab`, `vocabkey` |
| `core/database/subjects` | `CreateInput.GraphID` *(new)* per D-5 |
| `core/database/evidencegraphs/evidencegraphstest` | `Create` *(new)* per D-8 |
| `core/database/connect` | Per D-6 |
| `core/database/citations` | Delete `SourceIDTx` (D-7) |
| `api/ffi/handlers` | `connect.go`: drop the `sourceID` parse and field (D-6) |

```go
// requireSameGraph refuses an Observation whose Subject, Subject value, or
// Citation Artifact is not in, or linked to, the Subject's Evidence graph.
// A product reification (a source Subject) may be the value of a non-edge
// row in any graph.
func requireSameGraph(tx *sql.Tx, citationID, subjectID, valueSubjectID []byte, edge bool) error

const sqlCitationArtifact = `SELECT artifact_id FROM citations WHERE id = ?`
```

Errors: `observations.other_graph` (user) → the Subject is not in a graph linked to the Citation's Artifact, or the Subject value is in another graph (D-3 aside). `observations.invalid` → the Subject, Citation or Subject value does not exist (as today).

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

**STEP-2 — Subject into a given graph; graph fixture**
Files: `core/database/subjects/subjects.go`, `subjects_test.go`, `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go`, `core/database/deleteimpact/impact_test.go`.
Do: `CreateInput.GraphID` per D-5; record `graph_id` from the graph used. Add `evidencegraphstest.Create` (D-8) and move the `jones` create in `impact_test.go` onto it. Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/subjects/... ./core/database/deleteimpact/...` passes.

**STEP-3 — `requireSameGraph`**
Files: `observations.go`, `update.go`, `write.go`, `observations_test.go`. Skill: `add-catalog-query`.
Do: add `requireSameGraph` and `sqlCitationArtifact` (D-2, D-3); extend `sqlSubjectTypeKey` and its two scans. In `insertOne` and `updateOne`, remove the `requireSubjectSource` call and call `requireSameGraph` right after `resolveValue` (D-4). Delete `requireSubjectSource` and `sqlSubjectExists`. Add T-1.
Check: `rg -n 'requireSubjectSource|sqlSubjectExists' core` prints nothing. `CGO_ENABLED=1 go test -tags fts5 ./core/database/observations/...` passes.

**STEP-4 — Connect relies on the check**
Files: `core/database/connect/connect.go`, `connect_test.go`, `api/ffi/handlers/connect.go`, `core/database/citations/citations.go`, every file `rg -l 'connect.CreateInput\{'` lists.
Do: D-6 and D-7. Add T-2 subtests.
Check: `rg -n 'SourceIDTx' core api` prints nothing. `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes.

**STEP-5 — Copy**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/L10nResolutionTests.swift`. Skills: `add-localized-string`, `add-swift-test`.
Do: per Strings. Put the resource, the `case` and the `xcstrings` entry right after their `observations.invalid` neighbours. Rename `evidenceGraphErrorsHaveCopy` to `errorCodesHaveCopy` and add the `observations.other_graph` row (T-4).
Check: `python3 scripts/check-localizable-xcstrings.py` passes; the Xcode test run in `scripts/check-pr.sh` passes `L10nResolutionTests`.

**STEP-6 — Full check**
Files: none new.
Do: run the repo check. On a host without Xcode, run `gofmt -l`, `go vet ./...`, `CGO_ENABLED=1 go test -tags fts5 ./...`, `python3 scripts/check-localizable-xcstrings.py` and `python3 scripts/check-plan-docs.py`, and report the Xcode boxes as not run, for the user to run.
Check: `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-7 — Docs**
Files: `docs/interpretation-layer-data-model.md`, `docs/source-layer-data-model.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row, S12-08 scope entry, working rule "Each PR updates the docs and skills it makes false"), this plan (status, Landed note).
Do:
- `interpretation-layer-data-model.md`: the paragraph that starts "`source_id` is the subject's home Source" and invariant 3 say an Observation's Subject, Subject value and Citation Artifact share one Evidence graph, except a reification Subject value on a non-edge row; name `observations.requireSameGraph`. The next paragraph (the one about Subjects of type `source`) stays. Append invariant 24 with the same rule (no renumbering). The `graph_id` paragraph says `subjects.Create` takes `CreateInput.GraphID` when given, else `SoleGraphTx`.
- `source-layer-data-model.md` § 7.1: rewrite the `SoleGraphTx` bullet the same way, and add one bullet for isolation pointing to `observations.requireSameGraph` and `observations.other_graph`.
- `deployment-plan.md`: in the S12-08 scope entry, "List, create-Subject and cited-bridge RPCs take a graph id" becomes "List and create-Subject RPCs take a graph id", and the entry adds "`CreateCitedBridgeRequest` drops `source_id` and takes no graph id (the bridge's graph comes from its endpoints, S12-06)", and it adds "makes `subjects.CreateInput.GraphID` (S12-06) required". The working rule's S12-06 line names both docs.
- Then the rest of the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-06-graph-isolation.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

**T-1 fixture recipe.** One function inside the test builds the §2 fixture, in this order: `database.Create(t.TempDir(), …)`, `users.Upsert`, `subjectvocab.Install`, a source type through `sourcetypes.Upsert` (as `canonicalgraph_test.go` `newFixture` does). Sources Census and Book through `sources.Create`. Artifacts P1, P2 (Census) and B1 (Book) through `artifacts.Create`, each with no `GraphIDs`, so both link to their Source's Default. `evidencegraphstest.Create(…, census.ID, "Jones", P2.ID)`, then `evidencegraphs.Unlink(Default, P2)` inside `writes.Call`. Subjects through `subjects.Create`: Ann, Wed, Part, Rec with Census's Default id as `GraphID`; Ben with Jones's id; Cert and Cal with no `GraphID` on Book. Type ids by `subjecttypes.Lookup`, Property ids by `properties.Lookup` (`name`, `person`, `mentions`), all with `vocabkey.OriginProvenencia`. Each case writes once: `writes.Call(… "create_citation_with_observations" …)` → `citations.InsertWithObservationsTx(tx.Tx, citations.CreateInput{ArtifactID: …, LocatorJSON: validLocator}, obs, observations.InsertOptions{AllowEdgeRows: edge})`, where `edge` is true only for AC-1.5, AC-1.6 and AC-1.11. "No row left" means `citations` and `observations` row counts are as before the call. AC-1.9 and AC-1.10 save AC-1.1's row first, then call `observationstest.UpdateObservation` and re-read the row.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/observations/observations_test.go` `TestObservationsStayInOneGraph` *(new)*, table-driven over the recipe above | Each AC-1.x row → saved, or `errors.Is(err, observations.ErrOtherGraph)` with no row left | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.9, AC-1.10, AC-1.11, D-2, D-3, D-4 |
| T-2 | Go | `core/database/connect/connect_test.go` `TestCreateCitedBridge`: subtests "bridge joins its endpoints' graph", "endpoints in two graphs refused", "attached citation on an unlinked artifact refused" *(new)*. Each starts from `mustSeed`, adds Artifact P2, `evidencegraphstest.Create(…, "Jones", P2)`, unlinks P2 from Default, and creates Jones Subjects with `GraphID` and placements | AC-2.1 → `Subject.GraphID` is Jones; AC-2.2 → `ErrOtherGraph` and `countBridges` unchanged; AC-2.3 → `ErrOtherGraph`. Existing subtests stay green | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-6 |
| T-3 | Go | `core/database/subjects/subjects_test.go` `TestSubjects`: cases "graph id given", "graph of another source", "graph id wrong length" *(new)* | A second graph of the Source → `Subject.GraphID` equals it, `Get` reads it back, and the create change's `Fields["graph_id"].New` is it (as the existing check after `Get` in `TestSubjects` does); another Source's graph → `subjects.ErrInvalid`; 3 bytes → `ErrInvalid` | D-5 |
| T-4 | Swift | `macos/ProvenenciaTests/L10nResolutionTests.swift` `errorCodesHaveCopy` (renamed), new row | `message(code: "observations.other_graph")` equals the D-1 sentence and is not the unknown copy | AC-3.1, D-1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: the tests that build cited bridges (`core/database/autoreconciler`, `core/database/effects`, `core/database/graphcache` (`source_walk_test.go`), `core/search`, `core/database/citations/delete_test.go`, `core/database/deleteimpact/impact_test.go`, `core/database/promotealign`, `core/database/sourceeventtitles`, `api/ffi/handlers/connect_test.go`), through the `SourceID` removal.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: interpretation-layer-data-model.md, source-layer-data-model.md § 7.1, and the S12-08 scope entry and working rule in deployment-plan.md (STEP-7)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open an existing project. On a Source's Evidence graph, add a Citation with a name Observation, and connect a person to an event. Both save as before.
2. Promote that Source and open a canonical person's page. The events and places shown are the same as before the PR.

Two-graph refusals cannot be reached from the app until slice 4. T-1 and T-2 cover them.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| R-1. A writer saves Observations whose Subject value is in another Source today, and it starts failing | Every Source has one graph, so this only refuses rows that already crossed Sources. The writers are the ones §4 "Observation writers" lists: the composer and observation handlers, Connect, and test helpers |
| R-2. A stored row already crosses Sources and can no longer be updated | Accepted. The researcher deletes and re-records it. Not planned |
| R-3. A test that builds a bridge across two Sources breaks | It now fails with `ErrOtherGraph`. Fix the fixture to use one Source; do not weaken the check |

**Stop and report if:**

- `requireSubjectSource` is called from anywhere other than `insertOne` and `updateOne`.
- `observations` importing `evidencegraphs` or `subjectvocab` creates an import cycle (`go build` reports `import cycle not allowed`).
- Some code other than `connect.go` calls `citations.SourceIDTx`.
- An existing test or caller writes a Subject value across Sources on purpose (a fixture comment or doc says so), other than a non-edge `source`-typed value.
- An existing test expects `observations.ErrInvalid` (or another code) for a write that now returns `ErrOtherGraph`.
- The composite FK does not refuse a `GraphID` of another Source (T-3's second case saves a row).

## 12. Open questions

None.

## 13. PR

**Title:** S12-06 — Graph isolation
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-06-graph-isolation.md`.
