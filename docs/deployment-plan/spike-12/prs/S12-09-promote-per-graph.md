# S12-09 — Promote per graph

**Status:** Landed in [#535](https://github.com/mendahu/provenencia/pull/535) (2026-10-24)
**Spike / slice:** Spike 12 · Slice 3 — Graph id end to end
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: the Promote page looks and behaves as today; only the id it sends changes)
**Board:** —
**Depends on:** S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528)), S12-18 (landed [#532](https://github.com/mendahu/provenencia/pull/532)), S11-17 (landed #504), S11-25 (landed #517), S11-26 (landed #527), S11-27 (landed #530)
**Blocks:** —
**Branch:** `s12-09-promote-per-graph`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-18 did; no schema change, so the catalog format stays 55 ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [data model](../source-structure.md#data-model), [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) § 5.1 and § 7, [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) (Unpromote, bridge filing), the S12-08 [Landed note](S12-08-graph-id-end-to-end.md#landed) and its D-1, the S12-18 [Landed note](S12-18-subject-needs-graph.md#landed); working rules: "Promote is per graph", "Agreement counts per Source", "One bridge, one function", "A re-keyed symbol is renamed in the PR that re-keys it", "High tolerance for churn", "Each PR updates the docs and skills it makes false"
**Planned:** 2026-10-24 against `main` @ `eb5ec06`
**Review:** [`S12-09-promote-per-graph.review.md`](S12-09-promote-per-graph.review.md) — Fits with changes / Ready with changes, 2 rounds; every finding answered

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

Promote works on one Evidence graph. The proposal, the claim evidence sheet, Done (claims, pins and the bridges it files) and Unpromote's bridge unfiling each read or write the Subjects of one graph, named by its id. On a Source with two graphs, Promote all on the Jones graph proposes and files only the Jones Subjects, and a waiting bridge on the other graph stays unfiled until that graph's own Done. The last interim path that found "the Source's one graph" (`promotealign.loadInputs` through `evidencegraphs.SoleGraph`) goes, and `SoleGraph` and `SoleGraphTx` are deleted.

**Why now.** The working rule "One bridge, one function" retires the bridge in three PRs: S12-08 removed the handlers' use, S12-18 removed `subjects.InsertTx`'s, and this PR removes Promote's and deletes the two functions. S12-08 already put the graph on `PromoteEntry`, its `WorkspaceLocation` and its query key, so the Swift page already holds the graph id it must send. Spike 11 has finished rewriting Promote's matching and loaders (S11-17, S11-25, S11-26, S11-27 landed), so this PR re-keys settled code.

**What stays.** Every matching rule, score and band (Spike 11). The auto-reconciler and its count of distinct Sources (working rule "Agreement counts per Source"). `promote.Save` and `identityclaims.FileBridgesTx` (they work from one Subject). `UnpromoteSubjectRequest` (it names a Subject, whose graph the engine reads). The Promote page's chrome, copy and navigation. `PromoteEntry.sourceID` (the page still reads its Source's title and builds the canvas location from it). The test fixture `evidencegraphstest.SoleGraph` (S12-08 D-15).

## 2. User stories

**US-1.** As a researcher with a census split into a Smith graph and a Jones graph, when I open Promote on the Jones graph, I see only the Jones Subjects, and Done files only them and the bridges between them, so promoting one household never touches the other.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source Census with graphs Default (person Smith) and Jones (person Jones); `promotealign.Propose(c, jones.ID, nil)` | Rows name Jones only. `Propose(c, def.ID, nil)` rows name Smith only |
| AC-1.2 | Same Source; `promotealign.Evidence` with `GraphID: jones.ID` and `SubjectID: smith.ID` | `promote.ErrInvalid` (Smith is not a primary Subject of the Jones layer) |
| AC-1.3 | Same Source; on Default, a person and an event are each promoted with `promote.Save`, then a participation bridge joins them (it waits, unfiled); `promote.SaveBatch` with `Batch{GraphID: jones.ID, Rows: [{Jones, New}]}` | Jones is filed; the Default participation bridge has no accepted claim. A second `SaveBatch` with `Batch{GraphID: def.ID}` and no rows files it |
| AC-1.4 | `SaveBatch` with `Batch{GraphID: jones.ID}` and a row for Smith (a Default Subject) | `promote.ErrInvalid`; nothing is written |
| AC-1.5 | `Propose`, `Evidence` or `SaveBatch` with a graph id of 16 bytes that names no graph, or of 3 bytes; `ResolvePins` with such a graph id and a handle row that asks for default pins | `promote.ErrInvalid` |
| AC-1.6 | James, John and Mary promoted on Default, with filed bridges son-of (James–John), participation (James–event) and spouse (John–Mary); `promote.Unpromote(tx, user, james.ID)` | James's claim and the son-of and participation claims go; the spouse claim stays. `identityclaims.UnfileBridges` takes only the Subject: a bridge that ends on it is on its graph (working rule "Graphs are fully isolated") |

**US-2.** As the app, every Promote RPC names the graph the page is on, so the engine never guesses a Source's graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The Promote page opened on graph `g-2` (`PromoteEntry.graphID`) | `proposePromoteGraphAlignment`, `getPromoteClaimEvidence` and `applyPromoteGraphAlignment` each send `graphID: "g-2"` |
| AC-2.2 | `ProposePromoteGraphAlignmentRequest`, `GetPromoteClaimEvidenceRequest`, `ApplyPromoteGraphAlignmentRequest` with `graph_id: "nope"` | Each handler returns an error (`runRPC` `wantErr`) |
| AC-2.3 | The repo after this PR; `rg -n 'SoleGraph' core api -g '!*_test.go'` | Matches only `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go`. `rg -n 'evidencegraphs\.SoleGraph' core api` prints nothing |

## 3. Scope

| In | Covered by |
| --- | --- |
| `identityclaims.FileSourceBridgesTx` renamed `FileEvidenceGraphBridgesTx` and keyed by graph; the two head queries share one private `bridgeHeads`; `UnfileBridges` takes only the Subject (D-1, D-2) | STEP-1, AC-1.3, AC-1.6 |
| `promote.Batch.SourceID` replaced by `GraphID`; `SaveBatch` and `applyRow` check the graph; `Unpromote` unfiles on the Subject's graph (D-3, D-4) | STEP-1, AC-1.3, AC-1.4, AC-1.6 |
| `promotealign.Propose`, `EvidenceRequest`, `ResolvePins` and `loadInputs` take a graph id; `SoleGraph` use deleted (D-5) | STEP-1, AC-1.1, AC-1.2, AC-1.5 |
| `evidencegraphs.SoleGraph` and `SoleGraphTx` deleted (D-6) | STEP-1, AC-2.3 |
| Proto: the three Promote requests carry `graph_id` in place of `source_id`; handlers read it (D-7, D-8) | STEP-1, AC-2.2 |
| Go tests move onto graph ids; `TestSoleGraphTx` and the `core/database/audit/scopes_test.go` use go; two-graph tests added (D-9 – D-12) | STEP-2, STEP-3, STEP-4, STEP-5 |
| Swift store protocol, `GoStore`, `FakeStore`, the `OnboardingModelTests` stub and `PromoteModel` send `graphID` (D-13, D-14) | STEP-6, STEP-7, AC-2.1 |
| `promote-graph-alignment.md` and `conclusion-layer-data-model.md` say a layer, a Done's filing and Unpromote's unfiling are per graph (D-15) | STEP-8 |

| Out | Home |
| --- | --- |
| Any matching, scoring or band change | not planned (Spike 11 owns matching) |
| Layer memoization for large Sources | S11-21 (the layer memo is already keyed per graph since S12-07; S11-21 keys anything it adds the same way) |
| Renaming `ProposePromoteGraphAlignment` / `ApplyPromoteGraphAlignment` / `CatalogPromoteGraphAlignment*` | not planned: "graph alignment" names `graphalign`, not a Source |
| A Promote entry point per graph on the Source page or Sources list | S12-11, S12-14 |
| Graph names in Promote breadcrumbs | S12-11 |
| A per-graph credibility or reconciler change | not planned (working rules "Credibility stays per Source", "Agreement counts per Source") |
| Deleting `evidencegraphstest.SoleGraph` | not planned (S12-08 D-15) |
| Changing `UnpromoteSubjectRequest` or `promote.Save` | not planned (D-4: both start from one Subject) |

## 4. Current state

Verified at `eb5ec06`.

| Area | Where | Today |
| --- | --- | --- |
| Layer inputs | `core/database/promotealign/promotealign.go` `loadInputs(c, sourceID)` | Checks a 16-byte Source id, reads `evidencegraphs.SoleGraph(c, sourceID)` (`evidencegraphs.ErrInvalid` → `promote.ErrInvalid`), then `c.Graph().EvidenceGraph(graphID)` (`!eg.Found` → `promote.ErrInvalid`), the rules, terms and `loadLayerMemo(…, graphID, …)`. The only `SoleGraph` caller outside tests |
| Proposal | `core/database/promotealign/promotealign.go` `Propose(c, sourceID, fixed)`, `propose` | Check `len(sourceID) != 16`, read the audit revision, `loadInputs(c, sourceID)`, align. Doc comments say "for sourceID" / "the Source has no Subject" |
| Claim evidence | `core/database/promotealign/evidence.go` `EvidenceRequest{SourceID, SubjectID, HandleID []byte; Neighbors}`, `Evidence` | `loadInputs(c, req.SourceID)`; doc comments name "an unknown Source", "more neighbors than the Source has primary Subjects"; `evidence`'s comment says "every Subject the Source already filed" |
| Pins | `core/database/promotealign/pins.go` `ResolvePins(c, sourceID, rows)` | Lazily `loadInputs(c, sourceID)`; builds `EvidenceRequest{SourceID: sourceID, …}`; comment "A row already filed on the Source gets none" |
| Batch | `core/database/promote/batch.go` `Batch{SourceID, SeenRevision, Rows}`, `SaveBatch`, `applyRow(tx, sourceID, row)` | `SaveBatch` checks `len(in.SourceID) != 16`, `SELECT 1 FROM sources WHERE id = ?`, the revision, then each row; `applyRow` refuses a Subject whose `SourceID` differs; then `identityclaims.FileSourceBridgesTx(q, in.SourceID)`. Comment "Batch is one Done for a Source" |
| Unpromote | `core/database/promote/unpromote.go` `Unpromote` | `identityclaims.UnfileBridges(tx, subject.SourceID, subject.ID)`; comment "the bridges on its Source" |
| Bridge filing | `core/database/identityclaims/bridgefile.go` `FileSourceBridgesTx(tx, sourceID)`, `UnfileBridges(tx, sourceID, subjectID)` | Both select bridge Subjects `WHERE s.source_id = ?`. `FileSourceBridgesTx` and `bridgesTouching` (behind `FileBridgesTx`) each build the same key list, join, `SELECT s.id, st.key, s.subject_type_id` and `bridgeHead` scan loop; only the WHERE differs. `FileSourceBridgesTx` callers: `core/database/promote/batch.go`, `core/database/effects/conclusion_handles_test.go`. `UnfileBridges` callers: `core/database/promote/unpromote.go`, `core/database/promote/unpromote_test.go` |
| Bridge | `core/database/evidencegraphs/evidencegraphs.go` `SoleGraphTx`, `SoleGraph` | `SoleGraphTx` lists the Source's graphs and returns the one id; `SoleGraph` wraps it in a read transaction. Comment: "Promote's alignment reads it through SoleGraph until S12-09, which deletes both." `SourceOf(q, graphID)` beside them reads a graph's Source (`ErrInvalid` for a nil querier, a malformed id or no row) |
| Bridge tests | `core/database/evidencegraphs/evidencegraphs_test.go` `TestSoleGraphTx` (with its subtest "two graphs is not ErrInvalid"); `core/database/audit/scopes_test.go` case "graph link and unlink" `linkWrite` | `linkWrite` calls `evidencegraphs.SoleGraphTx(tx.Tx, s.a.ID)` inside its `writes.Call` closure |
| Test hooks | `core/database/promotealign/export_test.go` `LayerBridges(c, sourceID)`, `ProfiledRefsForTest(c, sourceID)` | `LayerBridges` reads `evidencegraphs.SoleGraph`; `ProfiledRefsForTest` calls `propose(c, sourceID, nil)`. One caller each in `promotealign_test.go` |
| promotealign tests | `core/database/promotealign/promotealign_test.go` `(*fixture).propose(sourceID, fixed)`; `pins_test.go` `(*fixture).resolve(src, rows...)`; `memo_test.go` `(*warmFixture).request()`; `evidence_test.go` (≈23 `EvidenceRequest{SourceID: …}` literals); `bench_test.go` (`promotealign.Propose(f.c, f.source.ID, nil)`) | Pass Source ids. `(*fixture).bare(src, kind)` reads `evidencegraphstest.SoleGraph` on its first line (S12-18), so it fails on a two-graph Source |
| Two-graph precedent | `core/database/promotealign/evidence_test.go` `TestEvidenceReadsMemberGraph` | Splits Census with `artifactstest.Split` and creates a Subject on the new graph with `subjectstest.On(t, f.c, userID, jones.ID, "person", "")` |
| promote tests | `core/database/promote/batch_test.go` `batchFixture` (field `graph`), `bridges_test.go` `bridgeWorld` (field `graph`), `pins_test.go`, `unpromote_test.go` subtest "a bridge on another Source stays filed" | Build `promote.Batch{SourceID: …}`; the fixtures already hold their Source's graph (S12-18 D-5 a) |
| Other Batch literals | `core/search/handles_test.go` (`handleFixture.graph`), `core/database/deleteimpact/pins_e2e_test.go`, `core/database/deleteimpact/conclusion_release_test.go` | `promote.Batch{SourceID: …}`; each holds a `graph` local or a fixture field from S12-18 |
| Proto | `api/proto/engine.proto` `ProposePromoteGraphAlignmentRequest` (`source_id = 2`), `GetPromoteClaimEvidenceRequest` (`source_id = 2`), `ApplyPromoteGraphAlignmentRequest` (`source_id = 3`, `reserved 6; reserved "skip_bridge_ids";`), `UnpromoteSubjectRequest` comment ("bridges on its Source") | Source-keyed |
| Handlers | `api/ffi/handlers/promote_align.go` `ProposePromoteGraphAlignment`, `promote_evidence.go` `GetPromoteClaimEvidence`, `promote_batch.go` `ApplyPromoteGraphAlignment` | `parseID(req.GetSourceId())`, passed to `Propose`, `EvidenceRequest.SourceID`, `ResolvePins` and `promote.Batch.SourceID` |
| Handler tests | `api/ffi/handlers/promote_align_test.go`, `promote_batch_test.go`, `promote_evidence_test.go` (`evidenceFixture.sourceID`, `request()`) | Requests set `SourceId`; cases "bad source id" send `SourceId: "nope"`. `soleGraphID(t, dir, sourceID)` in `api/ffi/handlers/evidence_graphs_test.go` reads a Source's one graph over the list RPC |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `proposePromoteGraphAlignment(projectDir:sourceID:fixed:)`, `getPromoteClaimEvidence(projectDir:sourceID:subjectID:handleID:neighbors:)`, `applyPromoteGraphAlignment(projectDir:userID:sourceID:seenRevision:rows:)`; `GoStore.swift` the same three; `FakeStore.swift` the same three and `buildPromoteProposal(projectDir:sourceID:fixed:)`; `macos/ProvenenciaTests/OnboardingModelTests.swift` a throwing stub of the three | Take `sourceID`. `FakeStore` reads `subjectsOfSource(sourceID)` for the proposal and Done, and records `"proposePromoteGraphAlignment"`, `"getPromoteClaimEvidence subject=… handle=… neighbors=…"`, `"applyPromoteGraphAlignment source=… rows=…"`. `FakeStore.subjectsByGraph` holds each graph's Subjects |
| Swift page | `macos/App/Features/Promote/PromoteModel.swift` `PromoteModel.load`, `loadEvidence`, `repropose`, `file` | Pass `sourceID: entry.sourceID` to the three calls; `PromoteEntry` already has `graphID` (S12-08) and `file` already applies `.promotedSubject(graphId: entry.graphID)` |
| Swift tests | `macos/ProvenenciaTests/PromoteModelTests.swift` `openingASheetLoadsItsEvidence` (asserts `"getPromoteClaimEvidence subject=a handle=e1 neighbors=0"`), `anOlderProposalAnsweringLastIsDropped` (filters `$0 == "proposePromoteGraphAlignment"`), helpers `store(subjects:)`, `model(store:subjectID:)` | One graph `g-1` on `src-1` |
| Docs | `docs/promote-graph-alignment.md` § 5.1 ("A layer is **one Source's Subjects**"), § 7 item 1 ("the Source's one Evidence graph"); `docs/conclusion-layer-data-model.md` Unpromote paragraph ("the bridges on that Source"), bridge filing paragraph ("That happens on Done") | Source-keyed wording |

**Precedent.** S12-08 ([#528](https://github.com/mendahu/provenencia/pull/528)) D-1 replaced `source_id` with `graph_id` on eight requests, reserving the old number and name and taking a new number (`ListSubjectMembershipsRequest`: `reserved 2; reserved "source_id"; string graph_id = 3;`); copy that shape. Its handler tests take graph ids through `soleGraphID`; copy that. S12-18 ([#532](https://github.com/mendahu/provenencia/pull/532)) gave each fixed-Source Promote fixture a `graph` field so this PR changes one field per fixture; use those fields. `TestEvidenceReadsMemberGraph` is the two-graph test shape to copy.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `identityclaims.FileSourceBridgesTx(tx, sourceID)` becomes `FileEvidenceGraphBridgesTx(tx, graphID)`, with early return for `len(graphID) != 16`. Doc comment: "FileEvidenceGraphBridgesTx files every unfiled bridge on the Evidence graph whose ends are both handles, including a bridge whose ends were claimed before this call. Does not commit." Its query and `bridgesTouching`'s share one private helper *(new)*: `func bridgeHeads(tx *sql.Tx, cond string, condArgs ...any) ([]bridgeHead, error)`, which owns the `connectrules.BridgeTypeKeys()` list (no keys → `nil, nil`), `SELECT s.id, st.key, s.subject_type_id FROM subjects s JOIN subject_types st ON st.id = s.subject_type_id AND st.origin = 'provenencia' WHERE st.key IN (…) AND (` + cond + `)`, the key arguments followed by `condArgs`, the copy-on-scan loop into `bridgeHead`, and `ORDER BY s.id`. `FileEvidenceGraphBridgesTx` calls `bridgeHeads(tx, "s.graph_id = ?", graphID)` then `fileHeads`; `bridgesTouching(tx, subjectID)` becomes `bridgeHeads(tx, "s.id = ? OR "+touchesEnd, subjectID, subjectID)` (delete `bridgesTouching`; `FileBridgesTx` calls `bridgeHeads` directly) | Working rules "Promote is per graph" and "A re-keyed symbol is renamed in the PR that re-keys it"; bare "graph" is taken (`graphcache.Graph`). One query owns the bridge-type and origin filter for filing and unfiling (D-2), so the next re-key or origin change edits one place (review M-1, M-5) | Keeping the name with a graph parameter — the name would say Source; a second function beside it — the churn rule; editing two copies of the same query in place — the copies drift (review M-1) |
| D-2 | `identityclaims.UnfileBridges(tx, sourceID, subjectID)` becomes `UnfileBridges(tx *database.Tx, subjectID []byte) ([]rowchange.Change, error)`: early return for `len(subjectID) != 16`; `heads, err := bridgeHeads(tx.Tx, touchesEnd, subjectID)`; for each head, `claim, err := AcceptedEntityForSubjectTx(tx.Tx, head.id)` (`sql.ErrNoRows` → skip the head; other errors returned), then `deleteimpact.ReleaseClaim(tx, claim.ID)` as today. Its own query, key list and scan loop are deleted. `bridgeHeads` orders by `s.id` (D-1), so claims release in bridge Subject id order as today. Doc comment: "UnfileBridges is the inverse of FileBridgesTx: it deletes, audited, the accepted claim of every bridge that has subjectID as an end, pins first, in bridge Subject id order. A bridge is on its ends' Evidence graph, so this is per graph. The bridges' handles (their associations) are left alone. Does not commit." | It mirrors `FileBridgesTx(tx, subjectID)` and selects bridges through the same `bridgeHeads`, so "which Subjects are bridges" has one owner (review M-5; `docs/seeded-vocabulary.md` "Connect policy" names plugin bridge types, which will change the origin filter). Isolation (S12-06, `observations.requireSameGraph`) keeps every bridge on its ends' graph, so unfiling is per graph by construction, and no caller can pass a graph that disagrees with the Subject (D-4's reason, one layer down; review M-2) | Re-keying the scope parameter to a graph id — two ids for one fact; keeping `UnfileBridges`' own query with the scope clause dropped — a third copy of the bridge-type filter (review M-5) |
| D-3 | `promote.Batch` becomes `{GraphID []byte; SeenRevision int64; Rows []BatchRow}`, comment "Batch is one Done for an Evidence graph." `SaveBatch`: after the user check, the `len(in.SourceID)` check and the `sources` existence query are replaced by `if _, err := evidencegraphs.SourceOf(q, in.GraphID); err != nil` (`evidencegraphs.ErrInvalid` → `ErrInvalid`, other errors returned; `SourceOf` owns the length check); then the revision check as today. `applyRow(tx, graphID, row)` refuses `!bytes.Equal(subject.GraphID, graphID)` with `ErrInvalid`. The filing call is `identityclaims.FileEvidenceGraphBridgesTx(q, in.GraphID)` | `SourceOf` is the one graph-row read and id check (S12-08 D-3, S12-18 D-1; review M-4); `promote` already reaches `evidencegraphs` through `subjects`, so no import cycle | A raw `SELECT 1 FROM evidence_graphs` in `promote` — a second graph read beside `SourceOf`; keeping `SourceID` and deriving the graph — the churn rule |
| D-4 | `promote.Unpromote` calls `identityclaims.UnfileBridges(tx, subject.ID)`; its comment's "the bridges on its Source" becomes "the bridges that have it as an end". `UnpromoteSubjectRequest` keeps its fields; its comment's "bridges on its Source" becomes "bridges on its Evidence graph". `promote.Save` does not change | The request names a Subject, and a Subject has one graph forever (working rule "Graphs are fully isolated"); a graph id on the wire could only disagree with it | Adding `graph_id` to `UnpromoteSubjectRequest` — a second id for one fact |
| D-5 | `promotealign`: `Propose(c, graphID, fixed)`, `propose(c, graphID, fixed)`, `loadInputs(c, graphID)`, `EvidenceRequest{GraphID, SubjectID, HandleID []byte; Neighbors}`, `ResolvePins(c, graphID, rows)`. `loadInputs` keeps its nil-catalog and `len != 16` checks, deletes the `SoleGraph` read and its error mapping, and keeps `c.Graph().EvidenceGraph(graphID)` with `!eg.Found` → `promote.ErrInvalid`. `inputs.graphID` stays and is set from the argument. Doc comments say "graph" / "Evidence graph" where they say "Source" today: `Propose` ("loads the Evidence layer of graphID"; "An unknown or empty graphID is promote.ErrInvalid"), `propose` ("when the graph has no Subject to place"), `inputs` ("the Evidence graph's layer"), `loadInputs` ("checks the catalog and graph id" and "loads the graph's layer"), `checkFixed` ("a primary Subject of the graph"), `EvidenceRequest` / `Evidence` ("An unknown graph, a row Subject that is not a primary Subject on it" and "more neighbors than the graph has primary Subjects", the rest of each sentence unchanged), `evidence` ("every Subject the graph already filed"), `ResolvePins` ("A row already filed on the graph gets none") | Working rule "Promote is per graph"; the layer, its memo and its members were already per graph (S12-07), so only the entry argument changes | Keeping a Source id and choosing "the first graph" — a second bridge |
| D-6 | Delete `evidencegraphs.SoleGraphTx` and `SoleGraph` and their doc comments. Remove any import (`fmt`) the deletion leaves unused in `evidencegraphs.go`, and nothing else | Working rule "One bridge, one function": this PR deletes both | — |
| D-7 | Proto, one edit per message, as S12-08 D-1: `ProposePromoteGraphAlignmentRequest`: `reserved 2; reserved "source_id"; string graph_id = 4;` and its comment "loads one Evidence graph's layer". `GetPromoteClaimEvidenceRequest`: `reserved 2; reserved "source_id"; string graph_id = 6;`. `ApplyPromoteGraphAlignmentRequest`: its existing `reserved 6; reserved "skip_bridge_ids";` become `reserved 3, 6; reserved "source_id", "skip_bridge_ids";`, plus `string graph_id = 7;`. `UnpromoteSubjectRequest` comment per D-4. Regenerate with `./scripts/generate-proto.sh`; never hand-edit `api/proto/engine/engine.pb.go` or `macos/App/Platform/Generated/engine.pb.swift` | A reused number would decode an old Source id as a graph id (S12-08 D-1) | Renaming in place (`source_id = 2` → `graph_id = 2`) — the wire would silently accept a Source id |
| D-8 | Handlers read `graphID, err := parseID(req.GetGraphId())` where they read `sourceID` today, and pass it to `Propose`, `EvidenceRequest.GraphID`, `ResolvePins` and `promote.Batch.GraphID`. `ProposePromoteGraphAlignment`'s comment: "loads one Evidence graph's layer"; `ApplyPromoteGraphAlignment`'s: "files one Done for an Evidence graph". No other handler change | `add-ffi-handler`: parse ids at the edge, pass typed values | — |
| D-9 | promotealign tests: add `func (f *fixture) graph(src sources.Source) []byte` *(new)* to `promotealign_test.go` beside `propose`, returning `evidencegraphstest.SoleGraph(f.t, f.c, src.ID).ID`. `(*fixture).propose(graphID, fixed)` and `(*fixture).resolve(graphID, rows...)` take a graph id; every call passes `f.graph(<src>)` where it passes `<src>.ID` or `<src>` today. Every `EvidenceRequest{SourceID: <src>.ID, …}` becomes `EvidenceRequest{GraphID: f.graph(<src>), …}` (in `memo_test.go` `(*warmFixture).request()` uses `w.f.graph(w.srcB)`; in `evidence_test.go` `likenessFixture.evidence()` uses `l.f.graph(l.srcB)`). `bench_test.go` passes `f.graph(f.source)`. `export_test.go`: `LayerBridges(c, graphID)` (delete its `SoleGraph` read and the `evidencegraphs` import) and `ProfiledRefsForTest(c, graphID)`; their callers pass `f.graph(f.source)` and `f.graph(srcB)`. `evidence_test.go` cases: "no such Source" becomes "no such graph" (`GraphID: missing`); "a Subject of another Source" becomes "a Subject of another graph" (same Subject, `GraphID: f.graph(srcB)`). A test that already holds a graph (`artifactstest.Split`, `evidencegraphstest.Split`) passes it. The existing sole-graph reads in the package move onto the helper: `(*fixture).bare` reads `f.graph(src)`, `memo_test.go`'s two `evidencegraphstest.SoleGraph` calls become `w.f.graph(w.srcB)` and `w.f.graph(w.f.source)`, and `source_store_test.go`'s three become `f.graph(f.source)`; drop `evidencegraphstest` imports left unused. The call sites include `memo_test.go`'s direct `promotealign.Propose(w.f.c, w.srcB.ID, nil)` calls, `pins_test.go`'s `EvidenceRequest` literals and its direct `promotealign.ResolvePins(f.c, srcB.ID, …)` calls. `f.graph` is called only outside a `writes.Run` / `writes.Call` closure | ≈60 call sites pass a single-graph Source; one helper keeps each site a one-token change and makes the graph explicit at the call, so a two-graph test passes its own graph. S12-18 D-5 (b) is the same read-before-write shape. One sole-graph read per fixture (review M-3) | A `propose` that takes a Source and reads its sole graph inside — hides the graph from the call and cannot express two graphs; spelling `evidencegraphstest.SoleGraph(t, f.c, src.ID).ID` at ≈60 sites |
| D-10 | promote, identityclaims and other core tests: `batchFixture`, `bridgeWorld`, `handleFixture` pass their `graph` field in `promote.Batch{GraphID: …}`; `core/database/promote/pins_test.go` passes `evidencegraphstest.SoleGraph(t, f.c, f.artifact.SourceID).ID` read before its `writes.Call`; `core/database/deleteimpact/pins_e2e_test.go` and `conclusion_release_test.go` pass the Subject's `GraphID` they already hold; `core/database/effects/conclusion_handles_test.go` calls `identityclaims.FileEvidenceGraphBridgesTx(tx.Tx, bob.GraphID)`. `unpromote_test.go`'s subtest "a bridge on another Source stays filed" is deleted: no call can name a scope now (D-2), and the first subtest already asserts the spouse bridge, which does not end on James, stays filed | Each test already holds the graph (S12-18 D-5); a Subject's `GraphID` is its graph | — |
| D-11 | `TestSoleGraphTx` is deleted whole. In `core/database/audit/scopes_test.go` case "graph link and unlink", `linkWrite` reads `graphID := evidencegraphstest.SoleGraph(t, s.c, s.a.ID).ID` before its `writes.Call` and passes it to `fn`; the closure no longer returns an error from a graph read | One connection: a catalog read inside a write closure blocks (S12-18 §4 "One connection"). `scopes_test.go` already imports `evidencegraphstest` | — |
| D-12 | Two-graph tests *(new)*. Each splits one Source with `artifactstest.Split(t, f.c, userID, src.ID, "Jones")` (it returns `def, jones, page2`; the Source's first Artifact stays on Default) and makes its two-graph Subjects with `subjectstest.On(t, f.c, userID, <graph>.ID, "person", "")` — never `f.bare` / `f.person` on a split Source, which read a sole graph — with no canvas position (as `TestEvidenceReadsMemberGraph`). In `promotealign`, a Subject that needs a name is cited with the fixture's `f.cite(<artifact>, s, observations.Input{PropertyID: f.prop("name").ID, Name: namevaluetest.Western(<name>)})`: the Source's first Artifact for Default, `page2` for Jones. `batch_test.go` cites nothing: the Jones Subject is not cited (a New row needs no Observation), and the Default person and event are `f.bare("person")` / `f.bare("event")` made **before** the split (`f.bare` uses the fixed `f.graph`, which is Default). T-4's error subtests call `writes.Call` around `promote.SaveBatch` directly, as `TestSaveBatchStaleWritesNothing` does, because `f.save` fails the test on any error. Tests: `promotealign_test.go` `TestProposeReadsOneGraph` (T-1), `evidence_test.go` `TestEvidenceRefusesASubjectOfAnotherGraph` (T-2), `batch_test.go` `TestSaveBatchFilesOnlyItsGraph` (T-4) | The map's Check is "On a two-graph Source, Promote all on one graph proposes and files only its Subjects"; Go tests are the only place a second graph can be made before S12-12 | A Swift two-graph Done test only — `FakeStore` does not file bridges |
| D-13 | Swift: the three `GenealogyStore` methods take `graphID: String` in the position of `sourceID:` (`proposePromoteGraphAlignment(projectDir:graphID:fixed:)`, `getPromoteClaimEvidence(projectDir:graphID:subjectID:handleID:neighbors:)`, `applyPromoteGraphAlignment(projectDir:userID:graphID:seenRevision:rows:)`); doc comments say "one Evidence graph". `GoStore` sets `req.graphID`. The `OnboardingModelTests` stub renames its ignored label. `PromoteModel` passes `graphID: entry.graphID` at its four call sites. `PromoteEntry.sourceID` stays | S12-08 put the graph on `PromoteEntry`; the churn rule replaces the label | Keeping `sourceID:` as the label for a graph id — the name would lie |
| D-14 | `FakeStore`: `proposePromoteGraphAlignment` records `"proposePromoteGraphAlignment graph=\(graphID)"`; `buildPromoteProposal(projectDir:graphID:fixed:)` reads `subjectsByGraph[graphID] ?? []`; `getPromoteClaimEvidence` records `"getPromoteClaimEvidence graph=\(graphID) subject=\(subjectID) handle=\(handleID) neighbors=\(neighbors.count)"`; `applyPromoteGraphAlignment` records `"applyPromoteGraphAlignment graph=\(graphID) rows=\(rows.count)"` and finds each row's Subject in `subjectsByGraph[graphID] ?? []` (else `promote.invalid`, as today). `PromoteModelTests`: `openingASheetLoadsItsEvidence` expects `"getPromoteClaimEvidence graph=g-1 subject=a handle=e1 neighbors=0"`; `anOlderProposalAnsweringLastIsDropped` filters `$0 == "proposePromoteGraphAlignment graph=g-1"`; `model(store:subjectID:)` gains `graphID: String? = nil` (`graphID ?? self.graphID` in the entry); new test `promoteSendsItsGraph` (T-11) | The fake mirrors Go's per-graph reads (`add-swift-test`), and the recorded calls let a test see which graph went out | A `lastPromoteGraphID` property — a second record beside `recordedCalls` |
| D-15 | Docs. `promote-graph-alignment.md` § 5.1 (keep the file's curly apostrophe): "A layer is **one Evidence graph’s Subjects**, not one connected component. *Map the rest of this graph* lists every primary Subject on that graph, including stragglers." § 7 item 1: its opening "**The layer:** the Source's one Evidence graph's Subjects" becomes "**The layer:** one Evidence graph's Subjects" (rest unchanged). `conclusion-layer-data-model.md` Unpromote paragraph: "the bridges on that Source" becomes "the bridges on that Evidence graph". Bridge filing paragraph: "That happens on Done for the bridge's Evidence graph, or later, when a claim makes its second end a member." | Working rule "Each PR updates the docs and skills it makes false" (map: S12-09 → these two docs) | — |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `identityclaims.FileEvidenceGraphBridgesTx` | func (rename of `FileSourceBridgesTx`) | `FileSourceBridgesTx`, `FileBridgesTx` | It is the same function re-keyed and renamed (D-1); `FileBridgesTx` files from one Subject and cannot catch a bridge whose ends were promoted earlier |
| `identityclaims.bridgeHeads` (private) | func | `bridgesTouching`, and the inline queries of `FileSourceBridgesTx` and `UnfileBridges` | It replaces all three copies (D-1, D-2); `bridgesTouching` is deleted |
| `graph_id` on three Promote requests | proto field | `source_id` on the same messages | Replaced, old number reserved (D-7) |
| `(*fixture).graph(src)` in `promotealign_test.go` | test helper | `evidencegraphstest.SoleGraph` | A one-line wrapper that binds the fixture's `t` and catalog for ≈60 call sites; the six existing spelled-out reads in `promotealign_test.go` `bare`, `memo_test.go` and `source_store_test.go` move onto it (D-9) |
| `TestProposeReadsOneGraph`, `TestEvidenceRefusesASubjectOfAnotherGraph`, `TestSaveBatchFilesOnlyItsGraph`, `promoteSendsItsGraph` | tests | `TestEvidenceReadsMemberGraph`, `doneSendsDefaultOrChosenPins` | New behavior: a second graph on one Source |

No production file, exported type, table, FFI method, component or L10n key is added.

**Builds on.** S12-07 (layer, memo and members keyed per graph), S12-08 D-1 (the reserved-field proto shape, `PromoteEntry.graphID`, `.promotedSubject(graphId:)`), S12-08 D-3 and S12-18 D-1 (`evidencegraphs.SourceOf` as the one graph-row read), S12-18 D-5 (fixtures hold their graph), working rule "One bridge, one function".
**Deletes.** `evidencegraphs.SoleGraphTx`, `SoleGraph` and `TestSoleGraphTx`; `promote.Batch.SourceID`; `SaveBatch`'s `sources` query; `FileSourceBridgesTx` (renamed); `bridgesTouching` and the second copy of the bridge-head query (D-1); `UnfileBridges`' scope parameter, its own query and scan loop, and the `unpromote_test.go` subtest that exercised the parameter (D-2, D-10); the `SoleGraph` read in `promotealign.loadInputs` and in `export_test.go` `LayerBridges`; `source_id` on three requests (reserved); the `sourceID:` label on three Swift store methods; `core/database/audit/scopes_test.go`'s in-closure `SoleGraphTx` call; the Source wording in two domain docs.
**Next adds.** S12-11 and S12-14 add per-graph Open buttons that build a Promote location with any graph id; nothing in Promote needs to change for them. S11-21, if it memoizes more, keys by the graph id `loadInputs` now receives.

### Schema

None — no catalog change. `subjects.graph_id` (S12-03) is what the filing queries filter on.

### Go core

| Package | Change |
| --- | --- |
| `core/database/identityclaims` | `FileEvidenceGraphBridgesTx(tx *sql.Tx, graphID []byte) (assocIDs [][]byte, changes []rowchange.Change, err error)` (D-1); `bridgeHeads` (private, D-1); `UnfileBridges(tx *database.Tx, subjectID []byte) ([]rowchange.Change, error)` (D-2) |
| `core/database/promote` | `Batch` per D-3; `SaveBatch`, `applyRow` per D-3; `Unpromote` per D-4 |
| `core/database/promotealign` | `Propose`, `propose`, `loadInputs`, `EvidenceRequest`, `Evidence`, `ResolvePins` per D-5 |
| `core/database/evidencegraphs` | `SoleGraphTx`, `SoleGraph` deleted (D-6) |

```go
// Batch is one Done for an Evidence graph. SeenRevision is the audit revision
// the proposal was read at.
type Batch struct {
	GraphID      []byte
	SeenRevision int64
	Rows         []BatchRow
}

type EvidenceRequest struct {
	GraphID, SubjectID, HandleID []byte
	Neighbors                    []graphalign.Fixed
}
```

Errors: `promote.invalid` (`promote.ErrInvalid`, user kind) → a missing, malformed or unknown graph id, or a row Subject on another graph, as today for a Source id or a Subject of another Source.

### FFI

`api/proto/engine.proto` per D-7; handlers `api/ffi/handlers/promote_align.go`, `promote_evidence.go`, `promote_batch.go` per D-8. Regenerate with `./scripts/generate-proto.sh` (writes `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`).

```proto
message ProposePromoteGraphAlignmentRequest {
  string project_dir = 1;
  reserved 2;
  reserved "source_id";
  repeated PromoteGraphAlignmentFixed fixed = 3;
  string graph_id = 4;
}
```

### macOS

None — no chrome change. Store and model per D-13 and D-14; files `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`, `macos/ProvenenciaTests/PromoteModelTests.swift`.

### Strings

None — no copy change; `promote.invalid` already has its message.

## 7. Steps

**STEP-1 — The engine works on one graph**
Files: `core/database/identityclaims/bridgefile.go`, `core/database/promote/batch.go`, `core/database/promote/unpromote.go`, `core/database/promotealign/promotealign.go`, `core/database/promotealign/evidence.go`, `core/database/promotealign/pins.go`, `core/database/evidencegraphs/evidencegraphs.go`, `api/proto/engine.proto`, `api/ffi/handlers/promote_align.go`, `api/ffi/handlers/promote_evidence.go`, `api/ffi/handlers/promote_batch.go`, the two generated files. Skills: `add-ffi-handler`, `add-catalog-query`.
Do: D-1 – D-8.
Check: `./scripts/generate-proto.sh` succeeds; `CGO_ENABLED=1 go build -tags fts5 ./core/... ./api/...` passes (non-test code); `rg -n 'SoleGraph' core api -g '!*_test.go'` lists only `core/database/evidencegraphs/evidencegraphstest/evidencegraphstest.go`.

**STEP-2 — identityclaims, promote and their callers' tests**
Files: `core/database/promote/batch_test.go`, `bridges_test.go`, `pins_test.go`, `unpromote_test.go`, `core/database/effects/conclusion_handles_test.go`, `core/database/deleteimpact/pins_e2e_test.go`, `core/database/deleteimpact/conclusion_release_test.go`, `core/search/handles_test.go`.
Do: D-10; `TestSaveBatchFilesOnlyItsGraph` per D-12 and T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promote/... ./core/database/identityclaims/... ./core/database/effects/... ./core/database/deleteimpact/... ./core/search/...` passes.

**STEP-3 — promotealign tests**
Files: `core/database/promotealign/promotealign_test.go`, `pins_test.go`, `memo_test.go`, `evidence_test.go`, `bench_test.go`, `export_test.go`, `source_store_test.go`.
Do: D-9; `TestProposeReadsOneGraph` and `TestEvidenceRefusesASubjectOfAnotherGraph` per D-12.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/...` passes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench . -benchtime 1x ./core/database/promotealign/` passes.

**STEP-4 — The bridge's tests go**
Files: `core/database/evidencegraphs/evidencegraphs_test.go`, `core/database/audit/scopes_test.go`.
Do: D-11. Remove any import the deletions leave unused, and nothing else (`uuid` stays in `evidencegraphs_test.go`: other tests use it).
Check: `CGO_ENABLED=1 go vet -tags fts5 ./core/...` reports nothing; `CGO_ENABLED=1 go test -tags fts5 ./core/...` passes; `rg -n 'evidencegraphs\.SoleGraph' core api` prints nothing.

**STEP-5 — Handler tests**
Files: `api/ffi/handlers/promote_align_test.go`, `promote_batch_test.go`, `promote_evidence_test.go`. Skill: `add-ffi-handler`.
Do: every Promote request sets `GraphId: soleGraphID(t, dir, sourceID)` where it sets `SourceId: sourceID` today (`evidenceFixture` gains `graphID string`, set with `soleGraphID` in `newEvidenceFixture` right after `f.sourceID` is set, and `request()` sends it; `sourceID` stays on the fixture only if another test reads it). Never call `soleGraphID` inside a `withProjectCatalog` closure. Cases "bad source id" become "bad graph id" with `GraphId: "nope"`. Add to `TestProposePromoteGraphAlignment` a case "unknown graph" sending a fresh `uuid.NewString()` as `GraphId`, `wantErr: true`. "empty source returns no rows" becomes "empty graph returns no rows". In `promote_evidence_test.go` `TestClaimEvidenceNeighborReading`'s `reserved(...)` checks, add `reserved("GetPromoteClaimEvidenceRequest", "source_id")`; add no other reservation assertion. A case that builds its request right after `personOnNewSource` reads `soleGraphID(t, dir, sourceID)` on the line before the request. In `promote_evidence_test.go` "a request that names nothing valid is promote.invalid": read `otherGraph := soleGraphID(t, f.dir, otherSource)` right after `otherSource` is made; the mutator "an unknown Source" becomes "an unknown graph" (`r.GraphId = uuid.Must(uuid.NewV7()).String()`), and "a Subject not on the Source" becomes "a Subject not on the graph" (`r.GraphId = otherGraph`).
Check: `CGO_ENABLED=1 go vet -tags fts5 ./api/...` reports nothing; `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes.

**STEP-6 — The app sends the graph**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/App/Features/Promote/PromoteModel.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift`.
Do: D-13; the `FakeStore` half of D-14.
Check: `xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO` succeeds; `rg -n 'sourceID: entry\.sourceID' macos/App/Features/Promote` prints nothing.

**STEP-7 — Swift tests**
Files: `macos/ProvenenciaTests/PromoteModelTests.swift`. Skill: `add-swift-test`.
Do: the `PromoteModelTests` half of D-14; T-11.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-8 — Docs**
Files: `docs/promote-graph-alignment.md`, `docs/conclusion-layer-data-model.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status, Landed note).
Do: D-15; map row status `Landed #N`; a `## S12-09 — Promote per graph` paragraph at the top of `completed.md` (newest first, above S12-18), in S12-18's format (**Landed** date and PR link, what shipped, **Not this PR.**); per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-09-promote-per-graph.md` reports no errors; `rg -n 'SoleGraph' docs -g '!docs/deployment-plan/**'` prints nothing; `scripts/check-pr.sh` prints `check-pr: PASS`.

Commits: three, because STEP-1 leaves the tests red until STEP-5: STEP-1 – STEP-5 in one commit, STEP-6 – STEP-7 in one, STEP-8 last. Each is one behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/promotealign/promotealign_test.go` `TestProposeReadsOneGraph` *(new)* | `census, censusArt := f.newSource("Census")` split into `def` (Smith, cited on `censusArt` with the name John Smith) and `jones` (Jones, cited on `page2` with the name Mary Jones); `Propose(f.c, jones.ID, nil)` → one row, Jones; `Propose(f.c, def.ID, nil)` → one row, Smith; `Propose` with 16 zero bytes and with `[]byte{1,2,3}` → `promote.ErrInvalid` | AC-1.1, AC-1.5, D-5 |
| T-2 | Go | `core/database/promotealign/evidence_test.go` `TestEvidenceRefusesASubjectOfAnotherGraph` *(new)* | A handle `h` from `f.promote(f.person(f.source, f.artifact, "John Smith"))` on the fixture's own Source; `census, censusArt := f.newSource("Census")`, split into `def` and `jones`; an unpromoted Smith on `def` cited on `censusArt` with the name John Smith; `Evidence(f.c, {GraphID: jones.ID, SubjectID: smith.ID, HandleID: h})` → `promote.ErrInvalid`; `{GraphID: def.ID, SubjectID: smith.ID, HandleID: h}` → no error | AC-1.2, D-5 |
| T-3 | Go | `core/database/promotealign/evidence_test.go` (the invalid-request table) "no such graph", "a Subject of another graph" | `GraphID: missing` → `ErrInvalid`; `aPerson` with `GraphID: f.graph(srcB)` → `ErrInvalid` | AC-1.5, D-9 |
| T-4 | Go | `core/database/promote/batch_test.go` `TestSaveBatchFilesOnlyItsGraph` *(new)* | `person := f.bare("person")`, `event := f.bare("event")` (Default), then `f.source` split into Default (`f.graph`) and Jones. Person and event each `promote.Save`d, then `f.participation(person, event)` → `f.bridgeClaims("participation") == 0`. Jones person `j` via `subjectstest.On`. `SaveBatch{GraphID: jones.ID, SeenRevision: f.revision(), Rows: [{j, New}]}` → `f.hasClaim(j.ID)`, `f.bridgeClaims("participation") == 0`. `SaveBatch{GraphID: f.graph, SeenRevision: f.revision()}` → `== 1`. Subtests: a row for the Default person with `GraphID: jones.ID` → `ErrInvalid` and the revision count is unchanged; `GraphID` 16 zero bytes → `ErrInvalid` | AC-1.3, AC-1.4, AC-1.5, D-1, D-3 |
| T-5 | Go | `core/database/promote/unpromote_test.go` `TestUnpromote` "takes the subject's claim and the bridges it ends off, and nothing else" (unchanged; the subtest "a bridge on another Source stays filed" is deleted) | Unpromote James → his claim, son-of and participation unfiled; spouse (John–Mary) stays filed; no handle deleted | AC-1.6, D-2, D-4 |
| T-6 | Go | `core/database/promotealign/pins_test.go` `TestResolvePins` (existing, now through `f.resolve(f.graph(srcB), …)`), plus subtest "an unknown graph is invalid" *(new)* | As today; `ResolvePins(f.c, make([]byte, 16), []ApplyRow{handleRow(incoming, handle, true)})` → `promote.ErrInvalid` | AC-1.5, D-5, D-9 |
| T-7 | Go | `core/database/effects/conclusion_handles_test.go` (existing case calling `FileEvidenceGraphBridgesTx(tx.Tx, bob.GraphID)`) | Both ends file, as today | D-1, D-10 |
| T-8 | Go | `api/ffi/handlers/promote_align_test.go` `TestProposePromoteGraphAlignment` "bad graph id", "unknown graph" *(new case)*, "empty graph returns no rows"; `promote_batch_test.go` `TestApplyPromoteGraphAlignment` "bad graph id", "mints a person in one revision"; `promote_evidence_test.go` "bad graph id" | `"nope"` → error; unknown uuid → error; a Source with no Subjects → empty response; Apply with the graph → one claim | AC-2.2, D-7, D-8 |
| T-9 | Go | `core/database/audit/scopes_test.go` `TestSourceScopes` "graph link and unlink" | Bumps Source A's scope, as today, with the graph read before the write | D-11 |
| T-10 | Repo | `rg -n 'SoleGraph' core api -g '!*_test.go'`; `rg -n 'evidencegraphs\.SoleGraph' core api` | Output as AC-2.3 | AC-2.3, D-6 |
| T-11 | Swift | `macos/ProvenenciaTests/PromoteModelTests.swift` `promoteSendsItsGraph` *(new)* | `store(subjects: [Ada "a"])` (graph `g-1`), plus graph `g-2` on `src-1` appended to `store.evidenceGraphsByProject[projectDir]` with `store.subjectsByGraph["g-2"] = [Zed "z"]` and a position for `z`; `store.promoteProposals = [CatalogPromoteGraphAlignmentProposal(revision: 1, rows: [row("z", handleID: "e1")])]`; `store.promoteEvidence["z"] = evidence([("l:1", true)])`; model on `subjectID: "z", graphID: "g-2"`; `load()`, `openSheet("z")`, wait for the board, `closeSheet()`, `done()`, wait for `lastPromoteBatch` → `recordedCalls` contains `"proposePromoteGraphAlignment graph=g-2"`, `"getPromoteClaimEvidence graph=g-2 subject=z handle=e1 neighbors=0"` and `"applyPromoteGraphAlignment graph=g-2 rows=1"`, and `lastPromoteBatch` holds only `z` | AC-2.1, D-13, D-14 |
| T-12 | Swift | `PromoteModelTests` `openingASheetLoadsItsEvidence`, `anOlderProposalAnsweringLastIsDropped` | Expected strings carry `graph=g-1`; otherwise as today | D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/promotealign` (≈60 call sites), `core/database/promote`, `api/ffi/handlers` Promote tests, and any test that hangs because a graph read landed inside a write closure (D-9, D-11).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none: no copy changes
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — none: no UI change
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: promote-graph-alignment.md § 5.1 and § 7, conclusion-layer-data-model.md (D-15)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

The app cannot make a second graph until S12-12, so the two-graph behavior is T-1, T-2 and T-4. In the app:

1. Open a project with a Source that has two unpromoted people on its Evidence graph. Click Promote on one, then *Map the rest of this graph*. Both rows show, as on `main`.
2. Open a row's evidence sheet. The evidence loads, as on `main`.
3. Click Done. Both people are filed, the toast reads as on `main`, and the canvas shows them promoted.
4. On the canvas, Unpromote one of them. Its claim and its bridges' claims go, as on `main`.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A graph read placed inside a `writes.Run` / `writes.Call` closure or a `withProjectCatalog` closure hangs a test (one connection) | D-9, D-11, STEP-5: every read runs before the write. A test that hangs for 10 minutes is this; move the read out |
| `f.bare` or another sole-graph helper called on a split Source fails the test | D-12: two-graph tests create Subjects with `subjectstest.On` on an explicit graph |
| Per-graph filing leaves a waiting bridge on another graph unfiled after a Done | That is the working rule "Promote is per graph" (T-4 asserts it); the bridge files on its own graph's Done or when a claim completes it (`FileBridgesTx`) |
| A stale Swift caller still passes a Source id | `rg -n 'sourceID: entry\.sourceID' macos/App/Features/Promote` in STEP-6; the Swift compiler rejects the old label |

**Stop and report if:**

- `promotealign.loadInputs`, `promote.SaveBatch`, `identityclaims.FileSourceBridgesTx` or `UnfileBridges` no longer matches §4 (a parameter, query or caller was added after `eb5ec06`).
- A non-test caller of `evidencegraphs.SoleGraph` or `SoleGraphTx` other than `promotealign.loadInputs` exists at your base.
- A non-test caller of `FileSourceBridgesTx`, `UnfileBridges`, `promotealign.Propose`, `Evidence`, `ResolvePins` or `promote.SaveBatch` exists outside the files in §4.
- `./scripts/generate-proto.sh` cannot run (no `protoc`, `protoc-gen-go` or `protoc-gen-swift`), or `xcodebuild` is not available: stop before STEP-1 commits and report; the generated files must be regenerated, not hand-edited.
- `promote` cannot import `evidencegraphs` (an import cycle `go build` reports).
- A test outside §4 / STEP-2 – STEP-5 / STEP-7 fails after the change for a reason other than a Source id passed where a graph id is now required.

## 12. Open questions

None.

## 13. PR

**Title:** S12-09 — Promote per graph
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-09-promote-per-graph.md`.

## Landed

Round 1 review changed these against the plan; nothing else differs.

- **Bridge-head query shape (F-1; supersedes D-1 and D-2's query shape only).** `bridgeHeads` joins `subject_types` with `CROSS JOIN` so SQLite starts at `subjects` through the condition's seek, not at the type index (which read every bridge Subject in the catalog). The touch condition is `s.id IN (SELECT o.subject_id FROM observations o WHERE o.value_subject_id = ?)`; `FileBridgesTx`'s is `s.id IN (SELECT ? UNION ALL SELECT o.subject_id ...)`. `TestBridgeHeadQueryPlans` pins the plans.
- **Shared promote test helper (F-2; supersedes the Codebase-fit list, which named no shared helper).** New package `core/database/promote/promotetest` with `Save`; the copies of the promote-in-a-write block in tests across `core` and `api` now call it, and `(*batchFixture).promoteOne` is deleted. Sites that assert on `Save`'s error stay as they were: the error-path cases in `promote_test.go`, and the random-operation walk in `autoreconciler_test.go`, which tolerates errors.
- **Unfile reads claims with the heads (F-3).** `bridgeHead` carries the accepted claim id from a `LEFT JOIN identity_claims`; `UnfileBridges` no longer runs one `AcceptedEntityForSubjectTx` per bridge. Behavior is unchanged.
- **Comment and import tidy (F-4).** Three comments re-wrapped (`engine.proto`, `batch.go`, `unpromote.go`) and the `rowchange` import moved into the module group of `bridgefile.go`.
- **Shared query-plan test helper (round 2 F-5).** New package `core/database/dbtest` with `QueryPlan`, which runs `EXPLAIN QUERY PLAN` and returns the detail lines; the plan tests in `canonicalgraph`, `evidencegraphs`, `graphcache` and `identityclaims` (`TestBridgeHeadQueryPlans`) now call it and keep their own assertions.
