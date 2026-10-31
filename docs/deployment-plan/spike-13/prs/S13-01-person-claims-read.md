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
**Review:** [`S13-01-person-claims-read.review.md`](S13-01-person-claims-read.review.md) — round 1: Fits with changes / Not ready, all findings Fixed; round 2 running

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

**Why now.** S13-16 (Sources section) draws this read. S13-04 (family) and S13-05 (timeline) reuse the accepted-members helper and declare their keys the way this PR declares `.personClaims`, so it lands first.

**What stays.** `GetConclusionDetail` keeps its payload. No view changes. The Evidence graph's own reads (`ListSubjects`, `ListObservationsByEvidenceGraph`, `GetCitation`) return what they return today.

## 2. User stories

**US-1.** As the engine, when asked for a person's claims, I return their accepted claims grouped by Source so that the Sources section can draw one group per Source with Subject cards beneath it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source "Parish register" (one Source, two Evidence graphs "Burials" and "Baptisms"), each with a Subject accepted into person P | One group: Source id, ref, title "Parish register", source type id; two claims ordered "Baptisms", "Burials", each with claim id, graph id, ref and name, the Subject, and that Subject's Observations |
| AC-1.2 | P has accepted claims from Sources titled "Zion census", "Abbey roll", "Écluse ledger", and a second Source also titled "Abbey roll" | Groups ordered "Abbey roll", "Abbey roll", "Écluse ledger", "Zion census" (case and diacritics ignored); the two "Abbey roll" groups ordered by Source ref |
| AC-1.3 | P also has a provisional (not accepted) claim | It is not returned |
| AC-1.4 | Entity id unknown, or the entity was merged away | Error `conclusiondetails.not_found` |
| AC-1.5 | An unmerged handle with no accepted members | Empty group list, no error |
| AC-1.6 | A claimed Subject has an Observation whose value is another Subject on the graph | That Observation carries the other Subject's label, as `ListObservationsByEvidenceGraph` does |
| AC-1.7 | A claim's graph has no ref (`evidence_graphs.ref` NULL) | The claim reads `graph_ref` "" and sorts by graph name like any other |
| AC-1.8 | Two claims each with a depiction Observation; the RPC called twice | The second call issues at most one query (the depiction path lookup) |

**US-2.** As the app, I keep a person's claims in the session cache and refresh them when they go stale.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Warmed `.conclusionDetail` and `.personClaims` keys for e1 and e3; write effects with `touchedHandles` [e1, e2] | Both e1 keys are revalidated through their descriptor `refreshes`; the e3 keys are not; `invalidations(for:project:)` no longer names any per-handle key |
| AC-2.2 | A promote (`.promotedSubject` mutation) | `.allCached(.personClaims)` and `.allCached(.conclusionDetail)` |
| AC-2.3 | Warmed `.personClaims` for e1 and `.conclusionDetail` for e1; write effects with `evidenceGraphs` [g1] and no handles (a Source retitle, a Source type change, a graph rename) | The `.personClaims` key is revalidated; the `.conclusionDetail` key is not |
| AC-2.4 | `.personClaims` key loads through the store | `GoStore` calls `ListPersonClaims` and maps groups, claims, `CatalogSubject` and `CatalogObservation` with the existing mappers |
| AC-2.5 | A Source's type changes | The loaded Evidence graphs of that Source are dropped in Go, so the next read has the new type id |
| AC-2.6 | A key both invalidated and named by a refresh in one `apply` | It is revalidated once |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache.Node.AcceptedMembers` *(new)*; `conclusiondetails.acceptedMembers` deleted and its caller moved onto it | STEP-1, AC-1.3 |
| `graphcache.EvidenceGraph` carries `GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID`; the `"sources"` effect drops a Source's graphs on `source_type_id` as well as `title` | STEP-2, AC-1.7, AC-2.5 |
| `conclusionheaders.SortByTitle` exported (was `sortByTitle`) | STEP-3, AC-1.2 |
| `persondetail.Claims` *(new package)* | STEP-4, AC-1.1 – AC-1.5 |
| `listedProtosFromGraph` split into its steps so a response maps all its Observations in one pass | STEP-5, AC-1.8 |
| `ListPersonClaims` RPC, handler, dispatch, proto regenerated | STEP-6, AC-1.6, AC-1.8 |
| Swift: store method in `GenealogyStore`, `GoStore`, `FakeStore`; types `CatalogPersonClaimGroup`, `CatalogPersonClaim` | STEP-7, AC-2.4 |
| Swift: descriptor `refreshes` list; `applyHiddenPolicy` follows any slot and may match any id; `CatalogQueryFollows.touchedHandles` and `.evidenceGraphs`; `.conclusionDetail` moved onto it | STEP-8, AC-2.1, AC-2.6 |
| Swift: `CatalogQueryKey.personClaims` and its descriptor arm | STEP-9, AC-2.1 – AC-2.3 |
| Docs and skills: `add-workspace-place` (Refresh section), working rules, scope entry | STEP-10 |

| Out | Home |
| --- | --- |
| The Sources section UI, type label lookup, Subject card placement, Observation display order | S13-16 |
| The Source date on a group | Descoped 2026-10-31, [`docs/ideas/source-provenance-date.md`](../../../ideas/source-provenance-date.md) |
| Family, timeline and record-links keys declaring their refreshes | S13-04, S13-05, S13-27 |
| Moving the hand-written `touchedHandles` → persons, events and places list fan-out (`invalidations(for:project:)`) onto `refreshes` | log it — [review M-2](S13-01-person-claims-read.review.md) |
| Moving title order out of `conclusionheaders` to a neutral package | when a third non-header caller appears — [review M-8](S13-01-person-claims-read.review.md) |
| Remove a claim | S13-17 |
| Event titles for event Subjects on a person's claims (a person's claims are Person Subjects) | not planned |
| Warming person claims at project open | not planned |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Accepted members | `core/database/conclusiondetails/conclusiondetails.go` `acceptedMembers` (L258), called at L168 | Private count of `n.Members` with `Accepted`. The only copy since S13-24 deleted `promotealign.acceptedClaims` |
| Node members | `core/database/graphcache/graph.go` `Node.Members`, `Member` | `Member{SubjectID, SourceID, GraphID, Accepted}`, in claim id order. No claim id |
| Loaded Evidence graph | `core/database/graphcache/evidence_graph.go` `EvidenceGraph`, `Graph.EvidenceGraph` | `GraphID` (set by `Graph.EvidenceGraph` after the loader), `SourceID`, `SourceTitle`, `Subjects`, `Observations`, `Members` (`EvidenceMember` carries `ClaimID`, `EntityID`). LRU of `maxEvidenceGraphs` = 64 |
| Graph loader | `core/database/graphfill/load.go` `loadEvidenceGraph`, `loadGraphSource` (L65) | `loadGraphSource` returns `(sourceID, title, found, err)` from `SELECT g.source_id, s.title FROM evidence_graphs g JOIN sources s …`. `evidence_graphs.ref` is nullable (`evidencegraphstest.SeedRefless`); `sources.ref` is not. `checkEvidenceGraph` reloads and compares the whole struct |
| Graph drop on Source | `core/database/effects/registry.go` `"sources"` entry, `EvidenceGraph: onField("title", graphsOfSource(self()))`; `graphsOfSource` (L202); `union` (`core/database/effects/path.go` L283) | A Source type change does not drop the Source's graphs |
| Title order | `core/database/conclusionheaders/order.go` `sortByTitle`, `titleCollator` | Unexported generic sort: Unicode root collation, case and diacritics ignored, empty titles last, ties by `strings.ToLower(ref)`. Callers: `compose.go` (2), `export_test.go` (1). A Swift comment at `macos/App/Features/Conclusions/ConclusionListPage.swift` L162 names it |
| Observation mapping | `api/ffi/handlers/source_read.go` `listedProtosFromGraph` (L43); `listedFromSource`; `api/ffi/handlers/observations.go` `listedProtos` (L387), `depictionPaths` | Per call: rebuilds property info and terms from the vocabulary, builds a Subject label map, maps rows, then `listedProtos` (1:1, in order) with one `files.RelPaths` query when any row has a depiction. Callers: `citations.go` L131 (`GetCitation`), `observations.go` L32 |
| Handler precedents | `api/ffi/handlers/subjects.go` `ListSubjects` (L149), `subjectProto` (L268); `api/ffi/handlers/conclusion_detail.go` `GetConclusionDetail`; `api/ffi/dispatch.go` | `ListSubjects` converts `graphcache.EvidenceSubject` to `subjects.Subject` inline for `subjectProto` |
| Proto | `api/proto/engine.proto` `enum Method` (highest 111; 107, 108, 110 reserved), `Subject` (L1038), `Observation` (L1950) | No person claims messages. Generated: `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift` |
| Claims in tests | `core/database/identityclaims/identityclaims_test.go` `mustClaim` (L248) | Builds a provisional claim with `identityclaims.Create(… Status: identityclaims.StatusProvisional)` inside `writes.Call`; Promote never writes one |
| Fixture helpers | `subjectstest.On`, `evidencegraphstest.SoleGraph`, `evidencegraphstest.Create`, `evidencegraphstest.SeedRefless`, `promotetest.Save`, `observationstest.FileArtifact`, `observationstest.SaveCitation`, `sources.Update`; merge by `UPDATE canonical_entities SET merged_into_id` as in `conclusiondetails_test.go` L444 | |
| Swift registry | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for:project:)` (L109 – 131), `.conclusionDetail` arm (L571 – 586), `conclusionTriggers` (L88 – 90), `CatalogQueryDescriptor` (L27 – 43) with `refresh: CatalogQueryRefresh?`, `make` (L49 – 74); refresh users `.citationCounts` (L347), `.citationsByArtifact` (L451), `.evidenceGraphProgress` (L474) | `touchedHandles` → `.conclusionDetail` fan-out is hand-written at L127 – 129 |
| Swift refresh | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryFollows` (L25 – 52), `CatalogQueryRefresh`, `applyHiddenPolicy(follows:id:)` (L117 – 125); `WorkspaceSession.swift` `apply` (L187 – 203), `apply(invalidations:)`, `applyHiddenPolicy(to:)` (L240) | `applyHiddenPolicy` takes a `CatalogRead` only and needs an id per key. `apply` runs each warmed key's one `refresh` in a `Task` after `apply(invalidations:)`, with no shared dedupe. No follows case for `touchedHandles` (`CatalogMutation.swift:111`, `handles + headerHandles`) or `evidenceGraphs` |
| Swift key | `macos/App/Features/Workspace/Session/CatalogQueryKey.swift` `conclusionDetail(project:entityId:)` (L31), `Kind` (CaseIterable) | Only per-handle key |
| Swift store | `macos/App/Platform/GenealogyStore.swift` `getConclusionDetail` (L1332), `CatalogSubject` (L257), `CatalogObservation` (L707); `macos/App/Platform/GoStore.swift` `getConclusionDetail` (L1222), `mapSubject` (L2086, `private static`), `mapObservation` (L2202); `macos/App/Platform/FakeStore.swift` `getConclusionDetail` (L1959), `seededConclusionDetails` (L92); `macos/App/Platform/CoreInvoke.swift` `CoreMethod` (L125) | |
| Subject card inputs | `macos/App/Features/Workspace/Session/EvidenceGraphSnapshot.swift` `EvidenceGraphPlacedSubject` (L19) | `CatalogSubject`, type label and kind (joined in Swift), observations sorted by `observationDisplayOrder`, membership. S13-16 builds it from this read |
| Tests | `core/database/graphfill/load_test.go` `TestRetitleDropsEvidenceGraph` (L37); `api/ffi/handlers/conclusion_detail_test.go`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs` (L857), `handleScopesInvalidateNothing` (L1011), `everyKindHasOneDescriptor` (L449, asserts 18 kinds), `hiddenPolicies`; `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `eachReadFollowsOnlyWhatFeedsIt` (L911) | |

**Precedent.** `GetConclusionDetail` (handler, `GoStore`, `FakeStore`, `.conclusionDetail` arm) is the shape for the RPC and the key. `ListSubjects` / `ListObservationsByEvidenceGraph` are the shape for reading Subjects and Observations from a loaded graph. `.citationCounts`' `refresh: .applyHiddenPolicy(follows:id:)` is the shape for a per-id revalidation; `eachReadFollowsOnlyWhatFeedsIt` is the shape for its session test. `TestRetitleDropsEvidenceGraph` is the shape for the Source-type drop test.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `persondetail.Claims(c *database.Catalog, entityID []byte) ([]ClaimGroup, error)` in a new package `core/database/persondetail`, file `claims.go` | Working rule "Each section reads its own query" names the package | A method on `conclusiondetails` — the rule keeps `GetConclusionDetail` from growing |
| D-2 | An unexported `memberGraphs(c, n) ([]memberGraph, error)` returns each accepted member (`n.AcceptedMembers()`) with its loaded `graphcache.EvidenceGraph` (one `c.Graph().EvidenceGraph` per distinct `GraphID`), its `EvidenceMember` (matching `SubjectID` and `EntityID`) and its `EvidenceSubject`. `Claims` builds groups from it; the Subject's Observations are the `EvidenceObservation`s with that `SubjectID` | All of it is in `graphcache`; S13-27 reuses the walk ([review M-7](S13-01-person-claims-read.review.md)) | Adding `ClaimID` to `graphcache.Member` — the loaded graph carries it |
| D-3 | `graphcache.EvidenceGraph` gains `GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID`. `loadGraphSource` becomes `loadGraphHead(q, graphID) (graphHead, bool, error)` returning an unexported struct, from `SELECT COALESCE(g.ref, ''), g.name, g.source_id, s.ref, s.title, s.source_type_id FROM evidence_graphs g JOIN sources s ON s.id = g.source_id WHERE g.id = ?` | One join already reads `sources`; a graph ref may be NULL; eight positional returns avoided ([review M-6, X-3](S13-01-person-claims-read.review.md)) | A `sources.Get` per group — SQL on every view |
| D-4 | The Source type is returned as `source_type_id` only. S13-16 resolves the label from the cached source types list, as `SourcesModel` does (user, 2026-10-31) | Keeps the read SQL-free when warm; a type rename refreshes only the source types list | Go label via `sourcetypes.GetByID` — one query per type per view |
| D-5 | The `"sources"` `EvidenceGraph` path becomes `graphsOfSource(union(onField("title", self()), onField("source_type_id", self())))`. The `DocEvidenceGraph` search path stays on `title` | One walk; a stale type id would show the old type ([review M-4](S13-01-person-claims-read.review.md)) | Two `graphsOfSource` walks in a union; dropping on every Source write |
| D-6 | Groups: `conclusionheaders.SortByTitle(groups, Title, SourceRef)`. Claims in a group: first `sort.SliceStable` by Subject ref (`strings.ToLower`), then `SortByTitle(claims, GraphName, GraphRef)`; the stable sort keeps Subject ref order inside a graph-name-and-ref tie. A blank name sorts last, by ref; an empty ref sorts first among equal names | R-36; one owner of title order | A local collation |
| D-7 | Unknown or merged entity → `conclusiondetails.ErrNotFound`. No kind check | Same code and Swift handling as the detail page | A new error code for the same page |
| D-8 | A member whose graph is not `Found`, or whose Subject or `EvidenceMember` is missing from the loaded graph, is skipped, not an error | A mismatch means a racing drop; the next refresh is correct | Failing the whole read |
| D-9 | Proto: `METHOD_LIST_PERSON_CLAIMS = 112`; `ListPersonClaimsRequest{project_dir = 1, entity_id = 2}`; `ListPersonClaimsResponse{repeated PersonClaimGroup groups = 1}`; `PersonClaimGroup{source_id = 1, source_ref = 2, source_title = 3, source_type_id = 4, repeated PersonClaim claims = 5}`; `PersonClaim{claim_id = 1, graph_id = 2, graph_ref = 3, graph_name = 4, Subject subject = 5, repeated Observation observations = 6}`. One stem across layers: Go `ClaimGroup` / `Claim`, proto `PersonClaimGroup` / `PersonClaim`, Swift `CatalogPersonClaimGroup` / `CatalogPersonClaim` ([review M-5](S13-01-person-claims-read.review.md)) | Reuses `Subject` and `Observation` | A new card message |
| D-10 | `listedProtosFromGraph` is split in `source_read.go`: `newListedContext(c) (listedContext, error)` *(new; property info and terms, once)*; `(listedContext).rows(eg graphcache.EvidenceGraph, keep func(graphcache.EvidenceObservation) bool) ([]observations.Listed, error)` *(new; Subject labels from `eg.Subjects`, then `listedFromSource` per kept row)*; `listedProtosFromGraph` keeps its signature and calls the two then `listedProtos`. `ListPersonClaims` makes one context, collects every claim's rows into one slice (keep = Subject id match, graph = the claim's loaded graph), calls `listedProtos` once, and splits the result back by each claim's row count | One mapping path; one depiction query per response ([review M-3](S13-01-person-claims-read.review.md)) | One `listedProtosFromGraph` per claim — a query and two map rebuilds per claim |
| D-11 | `persondetail.Claim` carries `ClaimID`, `Graph graphcache.EvidenceGraph` (the loaded value, shared, read-only), `Subject graphcache.EvidenceSubject` and `Observations []graphcache.EvidenceObservation`. The handler converts `Subject` to `subjects.Subject` inline as `ListSubjects` does. A claim with no Observations sends an empty `observations` list | The handler needs the graph's Subjects for labels (AC-1.6) | A separate `GraphSubjects` field |
| D-12 | Swift: `CatalogQueryDescriptor.refresh` becomes `refreshes: [CatalogQueryRefresh]` (and the `make` parameter, default `[]`); the three existing users pass one-element lists. `CatalogQueryFollows` gains `.touchedHandles` (`effects.touchedHandles`) and `.evidenceGraphs` (`effects.evidenceGraphs`). `CatalogQueryRefresh.applyHiddenPolicy(follows: CatalogQueryFollows, id: (@Sendable (CatalogQueryKey) -> String?)? = nil)`: with `id`, revalidates a warmed key whose id is among the followed ids; without, revalidates any warmed key of the kind when the slot names any id. Existing call sites pass `.reads(.citations)`. `WorkspaceSession.apply` runs every refresh of every warmed key and skips a key `apply(invalidations:)` already revalidated in this call (it returns its `seen` set). The hand-written `touchedHandles` → `.conclusionDetail` loop goes | One mechanism for "reload this key when the write names it"; each next key is a list entry, no new field ([review M-1, M-2](S13-01-person-claims-read.review.md)); working rule "One invalidation rule for person-section keys" | A new `perHandle` struct beside `refresh` — a second mechanism; a `Bool` per trigger kind |
| D-13 | `.conclusionDetail` declares `refreshes: [.applyHiddenPolicy(follows: .touchedHandles, id: entityId)]`. `.personClaims` arm, declared right after `.conclusionDetail`: `stalePolicy: .sessionFresh`, `invalidateOn: conclusionTriggers`, `whenHidden: .evict`, `invalidation: .allCached(.personClaims)`, `refreshes: [.applyHiddenPolicy(follows: .touchedHandles, id: entityId), .applyHiddenPolicy(follows: .evidenceGraphs)]` | Same lifetime as the detail page; a Source retitle or graph rename names a graph, not a handle (AC-2.3) | Mapping a graph id to its persons in Swift — the app does not know it |
| D-14 | Swift types in `GenealogyStore.swift`: `CatalogPersonClaimGroup{sourceID, sourceRef, title, sourceTypeID, claims}` and `CatalogPersonClaim{claimID, graphID, graphRef, graphName, subject: CatalogSubject, observations: [CatalogObservation]}`, `Sendable, Equatable`. `GoStore.listPersonClaims` maps with `mapSubject` (made `static`, not `private`) and `mapObservation`; wire order kept | Reuse the mappers; S13-16 sorts for display | Sorting in the store |
| D-15 | `FakeStore.seededPersonClaims: [String: [CatalogPersonClaimGroup]]` *(new)*. `listPersonClaims` returns the seeded value for a seeded entity (an empty list included) and throws `CoreInvokeError.coded(… code: "conclusiondetails.not_found" …)` for any other | One definition of unknown | Falling back to `membershipBySubject` — FakeStore would mirror the engine |
| D-16 | Go tests build their own fixture in `persondetail_test` from the public helpers in §4 (`database.Create`, `users.Upsert`, `subjectvocab.Install`, `sourcetypes.Upsert`, `sources.Create` through `writes.Run`, `evidencegraphstest.SoleGraph` / `Create` / `SeedRefless`, `subjectstest.On(…, "person", …)`, `promotetest.Save`, `identityclaims.Create` with `StatusProvisional` through `writes.Call`, `observationstest.FileArtifact` / `SaveCitation`). No shared fixture is extracted | Each helper exists; `conclusiondetails_test`'s fixture is in a test package | Extracting `conclusiondetails_test`'s fixture — out of scope |

**Rules this plan bends.** Working rule "Claims are read per person from the graph cache, grouped by Source in Go" said S13-01 "adds Source type through `sourcetypes`" → D-4 returns the type id and leaves the label to Swift → user answered 2026-10-31: return the type id → STEP-10 rewrites the rule and the S13-01 scope entry.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphcache.(*Node).AcceptedMembers` *(new)* | func | `conclusiondetails.acceptedMembers` | Moved, not copied; the private one is deleted |
| `EvidenceGraph.GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID` *(new)*; `graphfill.loadGraphHead`, `graphHead` *(new, replace `loadGraphSource`)* | fields / func | `EvidenceGraph.SourceTitle`, `loadGraphSource` | Extends the same struct and query |
| `conclusionheaders.SortByTitle` | func (exported) | `sortByTitle` | Renamed for export |
| `core/database/persondetail` *(new)*, `Claims`, `ClaimGroup`, `Claim`, `memberGraphs` | package / func / types | `conclusiondetails.ForEntity` | The working rule puts the section reads here |
| `newListedContext`, `listedContext.rows` *(new)* | funcs | `listedProtosFromGraph` | Its steps, split; it calls them |
| `ListPersonClaims` *(new)* | FFI | `GetConclusionDetail`, `ListSubjects` | A new read |
| `PersonClaimGroup`, `PersonClaim` *(new)* | proto | `Subject`, `Observation` | Wraps them |
| `CatalogPersonClaimGroup`, `CatalogPersonClaim` *(new)* | Swift types | `CatalogSubject`, `CatalogObservation` | Wrap them |
| `listPersonClaims` *(new)*, `seededPersonClaims` *(new)* | store method / fake seed | `getConclusionDetail`, `seededConclusionDetails` | One per RPC |
| `CatalogQueryFollows.touchedHandles`, `.evidenceGraphs` *(new)* | enum cases | `.handle`, `.reads` | `.handle` follows `handleScopes`, a different set |
| `CatalogQueryDescriptor.refreshes` (was `refresh`) | field | `refresh` | Widened to a list, as `patches` is |
| `CatalogQueryKey.personClaims` *(new)* | key | `.conclusionDetail` | One key per section |

**Builds on.** `graphcache` loaded Evidence graphs and their LRU; `effects` registry paths; `CatalogQueryRefresh.applyHiddenPolicy` and `WorkspaceSession.applyHiddenPolicy(to:)` (`add-workspace-place` Refresh); `conclusionTriggers`.
**Deletes.** `conclusiondetails.acceptedMembers`; `graphfill.loadGraphSource` (replaced by `loadGraphHead`); the hand-written `touchedHandles` → `.conclusionDetail` loop in `invalidations(for:project:)`; `CatalogQueryDescriptor.refresh` (replaced by `refreshes`).
**Next adds.** S13-04, S13-05 and S13-27 each add `refreshes` entries (`.touchedHandles` keyed, plus an any-id entry their plan picks) and use `AcceptedMembers`; S13-27 uses `memberGraphs`. S13-16 reads `.personClaims`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source type change now drops loaded graphs | `graphcache.dropEvidenceGraphs` via `effects` | Promote's graph memos and the Evidence graph view reload that Source's graphs, as they do on a retitle (`TestRetitleDropsEvidenceGraph`) |
| `.conclusionDetail` revalidates through `refreshes` instead of `invalidations(for:)` | `WorkspaceSession.apply`, `applyHiddenPolicy(to:)` | Same policy path as before (`applyHiddenPolicy(to:)`); T-7 pins visible reload and hidden evict |
| Refresh skipped for a key already revalidated | `WorkspaceSession.apply` | Existing refreshes (citations, progress) run as before when their key was not invalidated (`eachReadFollowsOnlyWhatFeedsIt` stays green) |

### Scale

At the map's Scale (a person with 5 – 20 Evidence graphs typical, 50 the budget): one `Node` read and one `EvidenceGraph` read per distinct graph. Warm, the Go read issues no SQL (T-4); the handler issues at most one `files.RelPaths` query per response (T-6). Cold, each graph loads with its existing loader queries. 50 graphs fit the 64-graph LRU; reading them can evict graphs Promote had loaded, which reload on next use. Work per claim is linear in its graph's Observations. On the app side, each write runs the `.personClaims` refreshes once per warmed person key; a hidden key is evicted, only a visible one reloads.

### Edges

- **Assumptions.** A member's `GraphID` names a graph whose loaded `Members` and `Subjects` contain the member (D-8 skips otherwise; T-3). Groups key on `SourceID`, not title (two Sources with one title stay two groups; T-2). `listedProtos` returns one proto per row, in order — the split in D-10 relies on it (T-6).
- **Neighbors.** Provisional members are excluded (T-1). A Subject with zero Observations yields a claim with an empty list (T-1). A ref-less graph reads "" (T-5).
- **Removed inputs.** An unpromoted member disappears on the next read: unpromote touches the handle (AC-2.1). A deleted graph is not `Found` and is skipped (D-8).
- **Real data.** None — no table-driven rule.
- **Untrusted size.** None — the catalog is the app's own; the count is bounded by the person's members.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `func (n *Node) AcceptedMembers() []Member` *(new)*: accepted members in `Members` order; nil node → nil. `EvidenceGraph` gains `GraphRef`, `GraphName`, `SourceRef`, `SourceTypeID` |
| `core/database/graphfill` | `loadGraphHead` *(new)* replaces `loadGraphSource` (D-3); `loadEvidenceGraph` fills the new fields |
| `core/database/effects` | `"sources"` `EvidenceGraph` path per D-5 |
| `core/database/conclusionheaders` | `sortByTitle` → `SortByTitle`; its doc comment (now on `titleCollator`'s line) moves onto the function and says it orders any titled row |
| `core/database/conclusiondetails` | `MemberCount: len(n.AcceptedMembers())`; `acceptedMembers` deleted |
| `core/database/persondetail` *(new)* | `type ClaimGroup struct{ SourceID []byte; SourceRef, Title string; SourceTypeID []byte; Claims []Claim }`; `type Claim struct{ ClaimID []byte; Graph graphcache.EvidenceGraph; Subject graphcache.EvidenceSubject; Observations []graphcache.EvidenceObservation }`; `func Claims(…)`; `memberGraphs` |

Errors: `conclusiondetails.not_found` → unknown or merged entity.

### FFI

`api/proto/engine.proto`: D-9. Handler `ListPersonClaims` in `api/ffi/handlers/person_detail.go` *(new)*; dispatch entry in `api/ffi/dispatch.go`. Regenerate with `./scripts/generate-proto.sh`; commit `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Store method | `listPersonClaims(projectDir:entityID:) async throws -> [CatalogPersonClaimGroup]` | `GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift` | D-14, D-15 |
| Method id | `CoreMethod.listPersonClaims` | `CoreInvoke.swift` | |
| Refresh | `refreshes`, `applyHiddenPolicy(follows:id:)`, `.touchedHandles`, `.evidenceGraphs` | `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift` | D-12 |
| Key | `CatalogQueryKey.personClaims(project:entityId:)`, `Kind.personClaims` | `CatalogQueryKey.swift` | arms in `kind` and `project`; D-13 |

States: none drawn in this PR.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — One accepted-members helper**
Files: `core/database/graphcache/graph.go`, `core/database/graphcache/graph_test.go`, `core/database/conclusiondetails/conclusiondetails.go`.
Do: add `AcceptedMembers` with a table test; switch `ForEntity`; delete `acceptedMembers`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/conclusiondetails/...` passes; `rg 'func acceptedMembers' core` finds nothing.

**STEP-2 — The loaded graph knows its ref, name, Source ref and type**
Files: `core/database/graphcache/evidence_graph.go`, `core/database/graphfill/load.go`, `core/database/graphfill/load_test.go`, `core/database/effects/registry.go`. Skill: `add-catalog-query`.
Do: D-3, D-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphfill/... ./core/database/effects/... ./core/database/graphcache/...` passes, including T-5 and `effects`' registry completeness test.

**STEP-3 — Export the title order**
Files: `core/database/conclusionheaders/order.go`, `compose.go`, `export_test.go`, `macos/App/Features/Conclusions/ConclusionListPage.swift` (comment).
Do: rename to `SortByTitle`; move and widen its doc comment.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionheaders/...` passes; `rg 'sortByTitle' core macos` finds nothing.

**STEP-4 — `persondetail.Claims`**
Files: `core/database/persondetail/claims.go` *(new)*, `core/database/persondetail/claims_test.go` *(new)*.
Do: D-1, D-2, D-6, D-7, D-8, D-11; fixture per D-16.
Check: T-1 – T-4 pass; `go list -deps ./core/database/persondetail | rg 'provenencia/(core/database/(conclusionheaders|graphcache)|api)'` lists `conclusionheaders` and `graphcache` and no `api` package.

**STEP-5 — Split the Observation mapping**
Files: `api/ffi/handlers/source_read.go`.
Do: D-10's split; `listedProtosFromGraph` keeps its signature and output.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/handlers/...` passes unchanged.

**STEP-6 — `ListPersonClaims` RPC**
Files: `api/proto/engine.proto`, `api/proto/engine/engine.pb.go`, `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/person_detail.go` *(new)*, `api/ffi/handlers/person_detail_test.go` *(new)*, `api/ffi/dispatch.go`. Skills: `add-ffi-handler`, `use-catalog-session`.
Do: D-9, D-10, D-11; `withProjectCatalog`.
Check: `./scripts/generate-proto.sh` leaves no diff on a second run; T-6 passes.

**STEP-7 — Swift store**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `CoreInvoke.swift`.
Do: D-14, D-15.
Check: the app and test targets build.

**STEP-8 — Refreshes as a list, following any slot**
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: D-12; `.conclusionDetail` per D-13; delete the hand-written loop. In `invalidationsComeFromRegistrySpecs` (L904 – 921) and `handleScopesInvalidateNothing`, delete the `.conclusionDetail` per-handle entries from the expected lists (they now come from `refreshes`); no other entry or order changes.
Check: the existing citations and progress refresh tests and `eachReadFollowsOnlyWhatFeedsIt` pass; T-7, T-8 pass.

**STEP-9 — The `personClaims` key**
Files: `CatalogQueryKey.swift`, `CatalogQueryRegistry.swift`, `CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-13. `everyKindHasOneDescriptor`: 18 → 19. `invalidationsComeFromRegistrySpecs`: the `.deletedSource` and `.promotedSubject` expected lists gain `.allCached(.personClaims)` directly after `.allCached(.conclusionDetail)` (order follows `Kind` declaration order).
Check: `everyKindHasOneDescriptor`, `hiddenPolicies`, `invalidationsComeFromRegistrySpecs`, T-9 – T-11 pass.

**STEP-10 — Docs**
Files: `.agents/skills/add-workspace-place/SKILL.md` (Refresh section); `docs/deployment-plan/spike-13/deployment-plan.md` (working rules "One invalidation rule for person-section keys" and "Claims are read per person from the graph cache, grouped by Source in Go", S13-01 row and scope entry); `docs/deployment-plan/spike-13/completed.md`; this plan.
Do: the skill's Refresh row says `refreshes` is a list, `applyHiddenPolicy` follows any slot (with or without an id), and a per-handle key declares `.touchedHandles` there rather than in `invalidations(for:project:)`; then the DoD. (The two working rules were updated when this plan was written.)
Check: links resolve; `python3 scripts/check-guidance-names.py` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log`. No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/persondetail/claims_test.go` *(new)* `TestClaimsGroupsBySource` *(new)* | One Source; its sole graph renamed "Burials" and a second graph "Baptisms" (`evidencegraphstest.Create`); one Subject each promoted into P (the first with two Observations, the second with none); a third graph's Subject claimed provisional on P → one group, claims "Baptisms", "Burials" with fields per AC-1.1; the "Baptisms" claim has an empty Observation list; no provisional | AC-1.1, AC-1.3 |
| T-2 | Go | `claims_test.go` `TestClaimsOrderSources` *(new)* | Sources "Zion census", "Abbey roll", "Écluse ledger", second "Abbey roll" → order per AC-1.2 | AC-1.2 |
| T-3 | Go | `claims_test.go` `TestClaimsMissingAndEmpty` *(new, table)* | unknown id → `ErrNotFound`; merged handle (`merged_into_id` set) → `ErrNotFound`; handle with only a provisional member → empty | AC-1.4, AC-1.5, D-8 |
| T-4 | Go | `claims_test.go` `TestClaimsWarmReadIssuesNoQuery` *(new)* | T-1's fixture; second `Claims` call with `SetQueryCounterForTest` → 0 queries | D-2 |
| T-5 | Go | `core/database/graphfill/load_test.go` `TestSourceTypeChangeDropsEvidenceGraph` *(new)* | `sourcetypes.Upsert` a second type; `sources.Update` the Source's type through `writes.Run` → next `EvidenceGraph` read has the new `SourceTypeID`; `GraphName`, `SourceRef` read back; a `SeedRefless` graph reads `GraphRef` "" | AC-1.7, AC-2.5, D-3, D-5 |
| T-6 | Go | `api/ffi/handlers/person_detail_test.go` *(new)* `TestListPersonClaims` *(new, runRPC)* | T-1's shape plus an Observation whose value is another Subject, and a depiction Observation on each of two claims → wire groups and claims; the value Subject's label set; a second call counts at most 1 query; bad id → `conclusiondetails.not_found` | AC-1.1, AC-1.6, AC-1.8, D-10 |
| T-7 | Swift | `WorkspaceSessionTests.swift` `touchedHandlesRevalidateOnlyTheirDetailPages` *(new)* | warmed `.conclusionDetail` e1, e3; effects `touchedHandles` [e1, e2] → e1 revalidated (visible) or evicted (hidden), e3 untouched | AC-2.1, D-12 |
| T-8 | Swift | `WorkspaceSessionTests.swift` `aKeyInvalidatedAndRefreshedRevalidatesOnce` *(new)* | warmed visible `.conclusionDetail` e1; `apply` with `touchedHandles` [e1] and a `.promotedSubject` mutation → one reload | AC-2.6 |
| T-9 | Swift | `WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs` *(new)* | warmed `.personClaims` e1, `.conclusionDetail` e1; effects `evidenceGraphs` [g1] → `.personClaims` revalidated, `.conclusionDetail` not; effects `touchedHandles` [e1] → both | AC-2.1, AC-2.3, D-13 |
| T-10 | Swift | `CatalogQueryRegistryTests.swift` `personClaimsInvalidateOnPromote` *(new)* | `.promotedSubject` → includes `.allCached(.personClaims)` | AC-2.2 |
| T-11 | Swift | `CatalogQueryRegistryTests.swift` `personClaimsLoadThroughTheStore` *(new)* | `FakeStore` seeded with one group for e1 → key loads it; unseeded e9 → coded `conclusiondetails.not_found` | AC-2.4, D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `invalidationsComeFromRegistrySpecs`, `handleScopesInvalidateNothing`, the citations and progress refresh tests, `TestRetitleDropsEvidenceGraph`, the `graphcache` Verify check, `conclusionheaders` order tests.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none added
- [ ] Sweep: `sortByTitle`, `acceptedMembers`, `loadGraphSource` and the descriptor `refresh:` field `rg`'d across `.agents/`, `.cursor/rules/`, `docs/*.md`, `macos/App` comments and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
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

None in the app: no view reads the key until S13-16. The reviewer runs `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/handlers/ -run 'ListPersonClaims'` and the Swift tests named in §8.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Adding fields to `EvidenceGraph` breaks Verify's whole-struct comparison | `checkEvidenceGraph` reloads through the same loader, so both sides carry the fields |
| Reading 50 graphs evicts Promote's warm graphs | Within the 64 cap; they reload on use |
| Every graph write revalidates every warmed `.personClaims` key | Hidden keys are evicted (`whenHidden: .evict`); only the visible page reloads, from cache |
| Skipping a refresh for a key already revalidated changes an existing refresh's behavior | Only keys `apply(invalidations:)` revalidated in the same call are skipped; they reload anyway |

**Stop and report if:**

- `loadGraphSource` no longer reads the graph's Source in one query, or `EvidenceGraph` gained a graph name another way.
- `EvidenceMember` no longer carries `ClaimID`, or a member's Subject is not on its graph's loaded `Subjects` in T-1.
- `listedProtos` does not return one proto per row in order.
- `invalidationsComeFromRegistrySpecs` needs an edit other than the ones STEP-8 and STEP-9 name.
- `union` of two `onField` paths inside `graphsOfSource` fails `effects`' registry completeness test.
- T-4 shows any query on the second read, or T-6 more than one.

## 12. Open questions

None.

## 13. PR

**Title:** S13-01 — Person claims read
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-13/prs/S13-01-person-claims-read.md`.
