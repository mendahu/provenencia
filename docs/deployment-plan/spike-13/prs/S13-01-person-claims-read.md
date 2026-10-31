# S13-01 — Person claims read

**Status:** Draft
**Spike / slice:** Spike 13 · Slice 3 — Family (the read serves slice 6's Sources section; its helper and invalidation spec are needed by S13-04)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: a read, a store method and a cache key; no view consumes them until S13-16)
**Board:** —
**Depends on:** —
**Blocks:** S13-04, S13-05, S13-16, S13-27
**Branch:** `s13-01-person-claims-read`
**Version bump:** none — Spike 11–14 PRs leave `VERSION` alone until a release is cut, as S13-06 and S13-10 did ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** [person-detail.md § 5. Sources](../person-detail.md#5-sources) (R-33 – R-37), working rules: "Each section reads its own query", "One invalidation rule for person-section keys", "Canonical only, filtered in Go", "Claims are read per person from the graph cache, grouped by Source in Go"
**Planned:** 2026-10-31 against `main` @ `474ca93`
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

The engine can answer "which records make up this person, and where do they come from": a canonical person's accepted identity claims, grouped by Source and ordered by Source title, each claim carrying the Subject card data the Evidence graph already renders. The app can fetch it under its own per-person cache key, which refreshes on its own when the person, their records, or the Source and Evidence graph around a record change.

**Why now.** S13-16 (Sources section) draws this read. S13-04 (family) and S13-05 (timeline) reuse the accepted-members helper and join the per-person invalidation rule this PR introduces, so it lands first.

**What stays.** `GetConclusionDetail` keeps its payload. No view changes. The Evidence graph's own reads (`ListSubjects`, `ListObservationsByEvidenceGraph`) are unchanged.

## 2. User stories

**US-1.** As the engine, when asked for a person's claims, I return their accepted claims grouped by Source so that the Sources section can draw one group per Source with Subject cards beneath it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source "Parish register" (one Source, two Evidence graphs "Baptisms" and "Burials"), each with a Subject accepted into person P | One group: Source id, ref, title "Parish register", source type id; two claims ordered "Baptisms", "Burials", each with claim id, graph id, ref and name, the Subject, and that Subject's Observations |
| AC-1.2 | P has accepted claims from Sources titled "Zion census", "Abbey roll", "Écluse ledger" | Groups ordered "Abbey roll", "Écluse ledger", "Zion census" (case and diacritics ignored, ties by ref) |
| AC-1.3 | P also has a provisional (not accepted) claim | It is not returned |
| AC-1.4 | Entity id unknown, or the entity was merged away | Error `conclusiondetails.not_found` |
| AC-1.5 | An unmerged handle with no accepted members | Empty group list, no error |
| AC-1.6 | A claimed Subject has an Observation whose value is another Subject on the graph | That Observation carries the other Subject's label, as `ListObservationsByEvidenceGraph` does |

**US-2.** As the app, I keep a person's claims in the session cache and refresh them when they go stale.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Write effects with `touchedHandles` [e1, e2] | Invalidations include `.conclusionDetail` and `.personClaims` for e1 and e2, both declared through the descriptor's per-handle entry, none hand-written in `invalidations(for:project:)` |
| AC-2.2 | A promote (`.promotedSubject` mutation) | `.allCached(.personClaims)` and `.allCached(.conclusionDetail)` |
| AC-2.3 | Write effects naming any Evidence graph (a Source retitle, a Source type change, a graph rename) and no handle | `.allCached(.personClaims)`; `.conclusionDetail` keys are not invalidated |
| AC-2.4 | `personClaims` key loads through `GoStore` | Calls `ListPersonClaims` and maps groups, claims, `CatalogSubject` and `CatalogObservation` with the existing mappers |
| AC-2.5 | Source type id changes on a Source | The loaded Evidence graphs of that Source are dropped in Go, so the next read has the new type id |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache.Node.AcceptedMembers` *(new)*; `conclusiondetails.acceptedMembers` deleted and its caller moved onto it | STEP-1, AC-1.3 |
| `graphcache.EvidenceGraph` carries `GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID`; `sources` effect drops a Source's graphs on `source_type_id` as well as `title` | STEP-2, AC-2.5 |
| `conclusionheaders.SortByTitle` exported (was `sortByTitle`) | STEP-3, AC-1.2 |
| `persondetail.Claims` *(new package)* | STEP-4, AC-1.1 – AC-1.5 |
| `ListPersonClaims` RPC, handler, dispatch, proto regenerated | STEP-5, AC-1.6 |
| Swift: store method in `GenealogyStore`, `GoStore`, `FakeStore`; types `CatalogPersonClaimSource`, `CatalogPersonClaim` | STEP-6, AC-2.4 |
| Swift: `CatalogQueryFollows.touchedHandles` *(new case)* and the descriptor's per-handle entry; `.conclusionDetail` moved onto it | STEP-7, AC-2.1 |
| Swift: `CatalogQueryKey.personClaims` and its descriptor arm | STEP-8, AC-2.1 – AC-2.3 |
| Docs and skills: `add-workspace-place` (per-handle invalidation), working rule wording | STEP-9 |

| Out | Home |
| --- | --- |
| The Sources section UI, type label lookup, Subject card placement | S13-16 |
| The Source date on a group | Descoped 2026-10-31, [`docs/ideas/source-provenance-date.md`](../../../ideas/source-provenance-date.md) |
| Family, timeline and record-links keys joining the per-handle entry | S13-04, S13-05, S13-27 |
| Remove a claim | S13-17 |
| Event titles for event Subjects on a person's claims (a person's claims are Person Subjects) | not planned |
| Warming person claims at project open | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Accepted members | `core/database/conclusiondetails/conclusiondetails.go` `acceptedMembers` (L258), called at L168 | Private count of `n.Members` with `Accepted`. The only copy since S13-24 deleted `promotealign.acceptedClaims` |
| Node members | `core/database/graphcache/graph.go` `Node.Members`, `Member` | `Member{SubjectID, SourceID, GraphID, Accepted}`, in claim id order. No claim id |
| Loaded Evidence graph | `core/database/graphcache/evidence_graph.go` `EvidenceGraph` | `GraphID` (set by `Graph.EvidenceGraph` after the loader), `SourceID`, `SourceTitle`, `Subjects`, `Observations`, `Members` (`EvidenceMember` carries `ClaimID`, `EntityID`). LRU of `maxEvidenceGraphs` = 64 |
| Graph loader | `core/database/graphfill/load.go` `loadEvidenceGraph`, `loadGraphSource` | `SELECT g.source_id, s.title FROM evidence_graphs g JOIN sources s …`. `checkEvidenceGraph` compares the whole struct |
| Graph drop on Source | `core/database/effects/registry.go` `"sources"` entry, `EvidenceGraph: onField("title", graphsOfSource(self()))` | A Source type change does not drop the Source's graphs. `union` exists in `core/database/effects/path.go` |
| Title order | `core/database/conclusionheaders/order.go` `sortByTitle`, `titleCollator` | Unexported generic sort: Unicode root collation, case and diacritics ignored, empty titles last, ties by ref. 3 callers in the package |
| Handler precedents | `api/ffi/handlers/subjects.go` `ListSubjects`, `subjectProto`; `api/ffi/handlers/observations.go` `ListObservationsByEvidenceGraph`; `api/ffi/handlers/source_read.go` `listedProtosFromGraph` | Read Subjects and Observations from `c.Graph().EvidenceGraph(graphID)`. `listedProtosFromGraph(c, eg, citationID)` labels value Subjects from `eg.Subjects`; `listedProtos` resolves depiction paths with one `files.RelPaths` query |
| Detail handler | `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`; `api/ffi/dispatch.go` | `withProjectCatalog` + `conclusiondetails.ForEntity`; dispatch by `MethodGetConclusionDetail` |
| Proto | `api/proto/engine.proto` `enum Method` (highest 111; 107 – 110 reserved), `Subject` (L1038), `Observation` (L1950) | No person claims messages |
| Swift registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for:project:)` (L109 – 131), `.conclusionDetail` arm (L571 – 586), `conclusionTriggers` (L88 – 90), `CatalogQueryDescriptor` (L27 – 43), `make` (L49 – 74) | `touchedHandles` → `.conclusionDetail` fan-out is hand-written at L127 – 129 |
| Swift follows | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryFollows` (L25 – 52) | Cases `.handle` (`handleScopes`), `.source`, `.reads`. None follows `touchedHandles` (`CatalogMutation.swift:111`, `handles + headerHandles`) |
| Swift key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `conclusionDetail(project:entityId:)` (L31), `Kind` (CaseIterable) | Only per-handle key |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `getConclusionDetail` (L1332), `CatalogSubject` (L257), `CatalogObservation` (L707); `macos/App/Platform/GoStore.swift` `getConclusionDetail` (L1222), `mapSubject` (L2086, private), `mapObservation` (L2202); `macos/App/Platform/FakeStore.swift` `getConclusionDetail` (L1959); `macos/App/Platform/CoreInvoke.swift` `CoreMethod` (L125) | |
| Subject card inputs | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphPlacedSubject` (L19) | `CatalogSubject`, type label and kind (joined in Swift from `propertiesWorkspace`), observations sorted by `observationDisplayOrder`, membership (claim id, entity). S13-16 builds it from this read |
| Tests | `core/database/graphfill/load_test.go` `TestRetitleDropsEvidenceGraph`; `api/ffi/handlers/conclusion_detail_test.go` `TestGetConclusionDetail`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs` (L857, asserts the `touchedHandles` order at L904 – 921), `everyKindHasOneDescriptor`, `hiddenPolicies`, `conclusionDetailLoadsAPersonsNameValuesAndOutcomes` | |

**Precedent.** `GetConclusionDetail` (handler, `GoStore`, `FakeStore`, `.conclusionDetail` arm) is the shape to copy for the RPC and the key. `ListSubjects` / `ListObservationsByEvidenceGraph` are the shape for reading Subjects and Observations from the loaded Evidence graph; copy their use of `subjectProto` and `listedProtosFromGraph`. `TestRetitleDropsEvidenceGraph` is the shape for the Source-type drop test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `persondetail.Claims(c *database.Catalog, entityID []byte) ([]SourceGroup, error)` in a new package `core/database/persondetail`, file `claims.go` | Working rule "Each section reads its own query" names the package | A method on `conclusiondetails` — the rule keeps `GetConclusionDetail` from growing |
| D-2 | Members come from `n.AcceptedMembers()`; each distinct `GraphID` is read once through `c.Graph().EvidenceGraph`; the claim id, Subject and Observations come from that loaded graph (`EvidenceMember` with matching `SubjectID` and `EntityID`; `EvidenceSubject` by id; `EvidenceObservation` with `SubjectID`) | All of it is in `graphcache` already; no new SQL on a warm read | Adding `ClaimID` to `graphcache.Member` — the loaded graph already carries it, and the Node loader would change for one reader |
| D-3 | `graphcache.EvidenceGraph` gains `GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID`, filled by `loadGraphSource` in the same query | The group and claim need them; one join already reads `sources` | A `sources.Get` per group — SQL on every view |
| D-4 | The Source type is returned as `source_type_id` only. S13-16 resolves the label from the cached source types list, as `SourcesModel` does (user, 2026-10-31) | Keeps the read SQL-free when warm; a type rename refreshes only the source types list | Go label via `sourcetypes.GetByID` — one query per type per view |
| D-5 | The `"sources"` effect drops a Source's graphs on `title` or `source_type_id`: `union(onField("title", …), onField("source_type_id", …))` | The loaded graph now holds the type id; a stale one would show the old type | Dropping on every Source write — drops graphs on notes and metadata edits |
| D-6 | Groups ordered with `conclusionheaders.SortByTitle` (exported from `sortByTitle`) on Source title, ties by Source ref. Claims in a group ordered by `SortByTitle` on graph name, ties by graph ref, then Subject ref | R-36; one owner of title order | A local `strings.ToLower` sort — a second collation |
| D-7 | Unknown or merged entity → `conclusiondetails.ErrNotFound`. No kind check: any unmerged handle's claims are returned | Same code and Swift handling as the detail page | A new `persondetail` error code — a second not-found for the same page |
| D-8 | A member whose graph is not `Found` or whose Subject is missing from the loaded graph is skipped, not an error | The two reads are of one committed state; a mismatch means a racing drop, and the next refresh is correct | Failing the whole read |
| D-9 | Proto: `METHOD_LIST_PERSON_CLAIMS = 112`; `ListPersonClaimsRequest{project_dir, entity_id}`; `ListPersonClaimsResponse{repeated PersonClaimSource sources}`; `PersonClaimSource{source_id, source_ref, source_title, source_type_id, repeated PersonClaim claims}`; `PersonClaim{claim_id, graph_id, graph_ref, graph_name, Subject subject, repeated Observation observations}` | Reuses the `Subject` and `Observation` messages the Subject card already maps | A new card message — a second mapping |
| D-10 | Observations are mapped by `listedProtosFromGraph(c, graphcache.EvidenceGraph{Subjects: eg.Subjects, Observations: <the Subject's>}, nil)` | Value Subjects get labels from the whole graph (AC-1.6) | Filtering inside `listedProtosFromGraph` — changes its signature for one caller |
| D-11 | Swift: `CatalogQueryFollows.touchedHandles` *(new)* returns `effects.touchedHandles`. `CatalogQueryDescriptor` gains `perHandle: CatalogQueryPerHandle?` *(new struct: `follows`, `key: @Sendable (ProjectKey, String) -> CatalogQueryKey`, `allOnEvidenceGraphs: Bool`)*. `invalidations(for:project:)` loops `CatalogQueryKey.Kind.allCases` in declaration order and, for each descriptor with `perHandle`, appends `.key(key(project, id))` for every followed id, then `.allCached(kind)` when `allOnEvidenceGraphs` and `effects.evidenceGraphs` is non-empty. The hand-written loop at L127 – 129 goes | Working rule "One invalidation rule for person-section keys"; S13-04, S13-05 and S13-27 each add one declaration | A second hand-written loop per key |
| D-12 | `.personClaims` arm: `stalePolicy: .sessionFresh`, `invalidateOn: conclusionTriggers`, `whenHidden: .evict`, `invalidation: .allCached(.personClaims)`, `perHandle` following `.touchedHandles` with `allOnEvidenceGraphs: true`. `.conclusionDetail` gets `perHandle` with `allOnEvidenceGraphs: false` | Same lifetime as the detail page; a Source retitle or graph rename touches no handle but changes a group or claim (AC-2.3) | Mapping a graph id to its persons in Swift — the app does not know it |
| D-13 | Swift types `CatalogPersonClaimSource{sourceID, sourceRef, title, sourceTypeID, claims}` and `CatalogPersonClaim{claimID, graphID, graphRef, graphName, subject: CatalogSubject, observations: [CatalogObservation]}` in `GenealogyStore.swift`; `GoStore` maps through `mapSubject` and `mapObservation`. Observations keep wire order; S13-16 sorts by `observationDisplayOrder` | Reuse the existing mappers | Sorting in the store — the snapshot owns display order |

**Rules this plan bends.** Working rule "Claims are read per person from the graph cache, grouped by Source in Go" said S13-01 "adds Source type through `sourcetypes`" → D-4 returns the type id and leaves the label to Swift → user answered 2026-10-31: return the type id → STEP-9 rewrites the rule and the S13-01 scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphcache.(*Node).AcceptedMembers` *(new)* | func | `conclusiondetails.acceptedMembers` | Moved, not copied; the private one is deleted |
| `EvidenceGraph.GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID` *(new)* | fields | `EvidenceGraph.SourceTitle` | Extends the same struct and query |
| `conclusionheaders.SortByTitle` | func (exported) | `sortByTitle` | Renamed for export; callers updated |
| `core/database/persondetail` *(new)*, `Claims`, `SourceGroup`, `Claim` | package / func / types | `conclusiondetails.ForEntity` | The working rule puts the section reads here |
| `ListPersonClaims` *(new)* | FFI | `GetConclusionDetail`, `ListSubjects` | A new read; reuses `subjectProto`, `listedProtosFromGraph` |
| `PersonClaimSource`, `PersonClaim` *(new)* | proto | `Subject`, `Observation` | Wraps them; no new card message |
| `CatalogPersonClaimSource`, `CatalogPersonClaim` *(new)* | Swift types | `CatalogSubject`, `CatalogObservation` | Wrap them |
| `listPersonClaims` *(new)* | store method | `getConclusionDetail` | One per RPC |
| `CatalogQueryFollows.touchedHandles` *(new)* | enum case | `.handle` | `.handle` follows `handleScopes`, a different set |
| `CatalogQueryPerHandle` *(new)*, `CatalogQueryDescriptor.perHandle` *(new)* | struct / field | `CatalogQueryRefresh` (`follows` + closure) | Refresh reloads part of a value; this invalidates whole keys |
| `CatalogQueryKey.personClaims` *(new)* | key | `.conclusionDetail` | One key per section |

**Builds on.** `graphcache` loaded Evidence graphs and their LRU; `effects` registry paths (`union`, `onField`); `CatalogQueryRegistry` descriptors (`add-workspace-place`); `conclusionTriggers`.
**Deletes.** `conclusiondetails.acceptedMembers`; the hand-written `touchedHandles` → `.conclusionDetail` loop in `invalidations(for:project:)`.
**Next adds.** S13-04, S13-05 and S13-27 each add one `perHandle` declaration (with `allOnEvidenceGraphs` as their plan decides) and use `AcceptedMembers`. S13-16 reads `.personClaims`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source type change now drops loaded graphs | `graphcache.dropEvidenceGraphs` via `effects` | Promote's graph memos and the Evidence graph view reload that Source's graphs on a type change; they did on a retitle already (`TestRetitleDropsEvidenceGraph`) |
| `perHandle` replaces the hand-written loop | `WorkspaceSession.apply` | `.conclusionDetail` invalidations are unchanged in set and order for the existing test (L904 – 921) |

### Scale

At the map's Scale (a person with 5 – 20 Evidence graphs typical, 50 the budget): one `Node` read and one `EvidenceGraph` read per distinct graph. Warm, no SQL except `listedProtos`' one batched `files.RelPaths` query when any Observation carries a depiction. Cold, each graph loads with its existing loader queries. 50 graphs fit the 64-graph LRU; reading them can evict graphs Promote had loaded, which reload on next use. Work per claim is linear in the graph's Observations (one filter pass per Subject). Invalidation adds one key per touched handle and one `.allCached` per write that names a graph.

### Edges

- **Assumptions.** A member's `GraphID` names a graph whose loaded `Members` contain the same Subject and entity (D-8 skips otherwise; T-3). Accepted members of one person can span many Sources; groups key on `SourceID`, not title (two Sources with the same title stay two groups; T-2).
- **Neighbors.** Provisional members are excluded (AC-1.3, T-1). A Subject with zero Observations still yields a claim with an empty list (T-1).
- **Removed inputs.** An unpromoted member disappears on the next read: unpromote touches the handle, so the key refreshes (AC-2.1). A deleted graph is not `Found` and is skipped (D-8).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — the catalog is the app's own; the count is bounded by the person's members.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `func (n *Node) AcceptedMembers() []Member` *(new)*: accepted members in `Members` order; nil node → nil. `EvidenceGraph` gains `GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID` |
| `core/database/graphfill` | `loadGraphSource` selects `g.ref, g.name, s.ref, s.source_type_id` too; `loadEvidenceGraph` sets the new fields |
| `core/database/effects` | `"sources"` `EvidenceGraph` path and the `DocEvidenceGraph` search path stay on `title`; only `EvidenceGraph` gains `source_type_id` via `union` |
| `core/database/conclusionheaders` | `sortByTitle` → `SortByTitle` |
| `core/database/conclusiondetails` | `MemberCount: len(n.AcceptedMembers())`; `acceptedMembers` deleted |
| `core/database/persondetail` *(new)* | `type SourceGroup struct{ SourceID []byte; SourceRef, Title string; SourceTypeID []byte; Claims []Claim }`; `type Claim struct{ ClaimID, GraphID []byte; GraphRef, GraphName string; Subject graphcache.EvidenceSubject; Observations []graphcache.EvidenceObservation; GraphSubjects []graphcache.EvidenceSubject }`; `func Claims(c *database.Catalog, entityID []byte) ([]SourceGroup, error)` |

Errors: `conclusiondetails.not_found` → unknown or merged entity.

### FFI

`api/proto/engine.proto`: `METHOD_LIST_PERSON_CLAIMS = 112` *(new)* and the messages in D-9. Handler `ListPersonClaims` in `api/ffi/handlers/person_detail.go` *(new)*; dispatch entry in `api/ffi/dispatch.go`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store method | `listPersonClaims(projectDir:entityID:) async throws -> [CatalogPersonClaimSource]` | `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift` | `FakeStore` returns `seededPersonClaims[entityID] ?? []`, throws `conclusiondetails.not_found` like `getConclusionDetail` for an unknown entity |
| Method id | `CoreMethod.listPersonClaims` | `CoreInvoke.swift` | |
| Key | `CatalogQueryKey.personClaims(project:entityId:)`, `Kind.personClaims` | `CatalogQueryKey.swift` | arms in `kind` and `project` |
| Descriptor | `.personClaims` arm, `perHandle` | `CatalogQueryRegistry.swift` | D-11, D-12 |

States: none drawn in this PR.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — One accepted-members helper**
Files: `core/database/graphcache/graph.go`, `core/database/conclusiondetails/conclusiondetails.go`, `core/database/graphcache/graph_test.go`.
Do: add `AcceptedMembers`; switch `ForEntity` to it; delete `acceptedMembers`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/conclusiondetails/...` passes; `rg 'func acceptedMembers' core` finds nothing.

**STEP-2 — The loaded graph knows its ref, name, Source ref and type**
Files: `core/database/graphcache/evidence_graph.go`, `core/database/graphfill/load.go`, `core/database/graphfill/load_test.go`, `core/database/effects/registry.go`. Skill: `add-catalog-query`.
Do: D-3, D-5.
Check: `go test -tags fts5 ./core/database/graphfill/... ./core/database/effects/... ./core/database/graphcache/...` passes, including T-5.

**STEP-3 — Export the title order**
Files: `core/database/conclusionheaders/order.go` and its callers.
Do: rename `sortByTitle` → `SortByTitle`; doc comment names it the owner of title order for lists and groups.
Check: `go test -tags fts5 ./core/database/conclusionheaders/...` passes.

**STEP-4 — `persondetail.Claims`**
Files: `core/database/persondetail/claims.go` *(new)*, `core/database/persondetail/claims_test.go` *(new)*.
Do: D-1, D-2, D-6, D-7, D-8.
Check: T-1 – T-4 pass.

**STEP-5 — `ListPersonClaims` RPC**
Files: `api/proto/engine.proto`, generated code, `api/ffi/handlers/person_detail.go` *(new)*, `api/ffi/handlers/person_detail_test.go` *(new)*, `api/ffi/dispatch.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-9, D-10; `withProjectCatalog`; map Subjects with `subjectProto`.
Check: `./scripts/generate-proto.sh` leaves no diff on rerun; T-6 passes.

**STEP-6 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`.
Do: D-13; the store method in all three.
Check: the app target builds.

**STEP-7 — Per-handle invalidation as a descriptor entry**
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`.
Do: D-11; declare `.conclusionDetail`'s `perHandle`; delete the hand-written loop.
Check: `invalidationsComeFromRegistrySpecs` passes unchanged; T-7 passes.

**STEP-8 — The `personClaims` key**
Files: `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `CatalogQueryRegistryTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-12.
Check: `everyKindHasOneDescriptor`, `hiddenPolicies`, T-8 – T-10 pass.

**STEP-9 — Docs**
Files: `.agents/skills/add-workspace-place/SKILL.md`; `docs/deployment-plan/spike-13/deployment-plan.md` (working rules "One invalidation rule for person-section keys" and "Claims are read per person from the graph cache, grouped by Source in Go", S13-01 row and scope entry); `completed.md`; this plan.
Do: the skill says a per-handle key declares `perHandle` (follows `.touchedHandles`) and never adds a loop to `invalidations(for:project:)`; the claims rule says the type is returned as an id and labelled in Swift; then the DoD.
Check: links resolve; `scripts/check-guidance-names.py` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/persondetail/claims_test.go` *(new)* `TestClaimsGroupsBySource` *(new)* | One Source, graphs "Burials" and "Baptisms", one accepted Subject each, one provisional in a third graph, one Subject with no Observations → one group, claims "Baptisms", "Burials", fields per AC-1.1; provisional absent | AC-1.1, AC-1.3 |
| T-2 | Go | `claims_test.go` `TestClaimsOrderSources` *(new)* | Titles "Zion census", "Abbey roll", "Écluse ledger", plus two Sources both titled "Abbey roll" → order by title then ref, two "Abbey roll" groups | AC-1.2 |
| T-3 | Go | `claims_test.go` `TestClaimsMissingAndEmpty` *(new, table)* | unknown id → `ErrNotFound`; merged handle → `ErrNotFound`; handle with no accepted members → empty | AC-1.4, AC-1.5, D-8 |
| T-4 | Go | `claims_test.go` `TestClaimsWarmReadIssuesNoQuery` *(new)* | Second call with `SetQueryCounterForTest` → 0 queries | D-2 |
| T-5 | Go | `core/database/graphfill/load_test.go` `TestSourceTypeChangeDropsEvidenceGraph` *(new)* | Change a Source's type → next `EvidenceGraph` read has the new `SourceTypeID`; graph name and ref read back | AC-2.5, D-3, D-5 |
| T-6 | Go | `api/ffi/handlers/person_detail_test.go` *(new)* `TestListPersonClaims` *(new, runRPC)* | Fixture of T-1 plus an Observation whose value is another Subject → wire groups and claims; that Observation's value Subject label set; bad id → `conclusiondetails.not_found` | AC-1.1, AC-1.6 |
| T-7 | Swift | `CatalogQueryRegistryTests.swift` `perHandleKeysFollowTouchedHandles` *(new)* | effects `touchedHandles` [e1, e2] → `.conclusionDetail` e1, e2 then `.personClaims` e1, e2; no hand-written path | AC-2.1, D-11 |
| T-8 | Swift | `CatalogQueryRegistryTests.swift` `personClaimsInvalidateOnPromote` *(new)* | `.promotedSubject` → `.allCached(.personClaims)` | AC-2.2 |
| T-9 | Swift | `CatalogQueryRegistryTests.swift` `evidenceGraphEffectsInvalidateAllPersonClaims` *(new)* | effects `evidenceGraphs` [g1], no handles → `.allCached(.personClaims)`, no `.conclusionDetail` key | AC-2.3, D-12 |
| T-10 | Swift | `CatalogQueryRegistryTests.swift` `personClaimsLoadThroughTheStore` *(new)* | `FakeStore` seeded with one group → key loads it; unknown entity → coded `conclusiondetails.not_found` | AC-2.4 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `invalidationsComeFromRegistrySpecs`, `TestRetitleDropsEvidenceGraph`, the `graphcache` Verify check (it compares the whole loaded graph), `conclusionheaders` order tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: `sortByTitle` and `acceptedMembers` `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none; add-workspace-place updated
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

None in the app: no view reads the key until S13-16. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/handlers/ -run ListPersonClaims` and the Swift registry tests named in §8.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Adding fields to `EvidenceGraph` breaks Verify's whole-struct comparison | `checkEvidenceGraph` reloads through the same loader, so both sides carry the fields; run `graphcache` tests |
| Reading 50 graphs evicts Promote's warm graphs | Within the 64 cap; they reload on use. Measured by T-4 for warmth only |
| `.allCached(.personClaims)` on every graph write reloads open person pages often | Hidden keys are evicted (`whenHidden: .evict`); only the visible page reloads, from cache |

**Stop and report if:**

- `loadGraphSource` no longer reads the graph's Source in one query, or `EvidenceGraph` gained a graph name another way.
- `EvidenceMember` no longer carries `ClaimID`, or a member's Subject is not on its graph's loaded `Subjects`.
- `invalidationsComeFromRegistrySpecs` cannot pass without changing its expected order.
- `effects` has no `union` path, or `union` of two `onField` paths fails `effects`' own checks.
- T-4 shows any query on the second read.

## 12. Open questions

None.

## 13. PR

**Title:** S13-01 — Person claims read
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-01-person-claims-read.md`.
