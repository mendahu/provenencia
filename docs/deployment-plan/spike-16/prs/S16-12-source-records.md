# S16-12 — Source records in the graph cache

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Sources section, the record links diagram and the provenance sheet draw the same words; only where they are read from moves)
**Board:** —
**Depends on:** S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789), the cache registry this entry joins), S16-01 (landed [#756](https://github.com/mendahu/provenencia/pull/756), same `claims.go` and `recordlinks.go`); S16-11 (landed [#793](https://github.com/mendahu/provenencia/pull/793)) left the outcome record shells as the one SQL read this PR narrows
**Blocks:** S16-13 (composes claims from the records tier and these records), S16-30 (a provenance record composes the Source record), S16-25 (replaces this PR's `.source` follows with `.followsInputs`), S15-05 (adds the Source date to the record), S15-24 (the move changes an Artifact's Source; this entry's Apply sees both)
**Branch:** `s16-12-source-records`
**Version bump:** none — no change a researcher can see; the wire change ships inside the one app binary ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (Refresh), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`docs/catalog-caches.md`](../../../catalog-caches.md); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); [`macos-client-patterns.md`](../../../macos-client-patterns.md) "Conclusion keys"; audit P-22 ([`person-page-audit.md`](../person-page-audit.md)); working rules: "(Batch 2) The point of this batch is simpler code", "(Batch 2) Cache entities and results per key; compose answers", "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) A write's one-shot graph fills only what it reads", "(Batch 2) A Source's words live once", "(Batch 2) Warm only what every proposal needs; no background re-warm", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth"; goal bar 11 (the rename clause)
**Planned:** 2026-11-12 against `main` @ `8d34930`
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

A Source's words (ref, title, type, credibility, its Artifacts' labels) are copied today into every loaded Evidence graph and joined again into the provenance sheet's record shells. So a Source rename evicts every loaded graph of that Source and drops their memos, and the next Sources section or Promote read reloads them whole. After this PR the graph cache holds one record per Source, built on first read and patched from `effects.Set.SourceScopes`; the Sources section, the record links diagram and the provenance sheet compose Source words from it; a rename reloads one Source record and no Evidence graph. The record links wire carries the Source's ref, so the app stops passing a ref map between the two Sources views and writes the "title, else ref" and "name, else ref" fallbacks once.

**Why now.** S16-13 (the records tier) and S16-30 (provenance composed) compose from Source records; S15-05 adds the Source date to the record and S15-24's move changes an Artifact's Source. Both spikes wait on the record existing. Audit P-22's wire half rides here because this PR changes the same message (user, 2026-11-08).

**What stays.** Every word, order and label the researcher sees; the claims wire (`PersonClaimGroup`); the Evidence-graph store's 64-graph LRU and its drop on `Set.EvidenceGraphs` for Subject, Observation, claim and Citation writes; the auto-reconciler's own candidate SQL (a write path); search documents (`DocEvidenceGraph` still reprojects on a title edit); `.evidenceGraphs` follows on the claims and record links keys.

## 2. User stories

**US-1.** As a researcher, when I rename a Source, the Person page's Sources section, its record links diagram and the provenance sheet show the new title, and the next Promote or Sources read does not reload that Source's Evidence graphs from disk.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Graphs G1 and G2 of Source "Register" are loaded, each with a memo stored (`SetEvidenceGraphMemo`); the Source is renamed "Parish register" through `writes.Run` | Both graphs are still loaded (`LoadedEvidenceGraphForTest`) and both memos still read back |
| AC-1.2 | A Person with a claim in G1; Source renamed as in AC-1.1 | `persondetail.Claims` returns the group with `Source.Title == "Parish register"`; `persondetail.RecordLinks` returns G1 with `Source.Title == "Parish register"` |
| AC-1.3 | A Person whose name Observation cites an Artifact "Page 1" of Source "Register" (credibility `low_trust`); then the Source is renamed, the Artifact relabelled "Folio 1" and the grade set to `high_trust` | `conclusiondetails.ForEntity` reads before: `SourceTitle == "Register"`, `ArtifactLabel == "Page 1"`, `CredibilityKey == "low_trust"`, `Provenance.Credibility < 0`; after: "Parish register", "Folio 1", "high_trust", `> 0` |
| AC-1.4 | The Person page's claims key and record links key are cached with Source "s1" in their values; a write reports `sourceScopes: ["s1"]` | Both keys revalidate; a write with `sourceScopes: ["s9"]` revalidates neither |
| AC-1.5 | A Source type change on a Source with a loaded graph | The graph stays loaded; the claims group reads the new type id |

**US-2.** As the engine, a Source's words live in one place and update exactly from the commit.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Source records held; a write creates an Artifact on Source A, deletes Source B, moves an Artifact from A to C (an `artifacts` update of `source_id`) | After each commit the held index equals a fresh build (`Check` passes); B is gone; the moved Artifact resolves to C |
| AC-2.2 | A write on a write's graph (`graphcache.Over(tx)`) calls `sources.RecordsOf` | It returns `sources.ErrRecordsOnWrite` and builds nothing |
| AC-2.3 | A Source title edit resolved by `effects.Resolve` | `Set.EvidenceGraphs` is empty; `Set.SourceScopes` holds the Source; the `DocEvidenceGraph` search docs still name every graph of the Source |
| AC-2.4 | Graph dropped, vocabulary loaded; `sources.RecordsOf` warms the records, then `conclusiondetails.ForEntity` reads a small and a big Person | Warming issues exactly two queries; the two reads issue the same count, and `sqlRecords` names no `sources`, `artifacts` or `source_credibility_` table (`conclusiondetails.SQLRecordsForTest` *(new)*, `export_test.go`) |

**US-3.** As a contributor, the record links wire says which Source each graph is on by ref, and the app writes each title fallback once.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `GetPersonRecordLinks` for a person in graphs on Sources "SRC-A" (titled "Census") and "SRC-B" (untitled) | Every `RecordLinksGraph` has `source_ref` = its Source's ref and `source_id` = its Source's id; `source_title` is "Census" and "" |
| AC-3.2 | `PersonRecordLinksContent` over a graph with `sourceTitle: ""`, `sourceRef: "SRC-1"` | Its node help names the Source "SRC-1" (as today, now without a ref map) |
| AC-3.3 | `PersonSourcesContent` over today's fixtures | Row titles, line titles and accessibility labels unchanged (existing tests green) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `sources.Records` whole registry entry: Build (two queries), Apply on `Set.SourceScopes`, Check, graph lifetime | STEP-2, AC-2.1, AC-2.2 |
| Credibility read through one shared SQL column (`autoreconciler.CredibilityColumn`) | STEP-1, STEP-2, AC-1.3 |
| Claims grouping and record links take Source words from the record | STEP-3, AC-1.2, AC-1.5 |
| The detail read's record shells drop the Source, Artifact and credibility joins | STEP-4, AC-1.3, AC-2.4 |
| `EvidenceGraph` stops copying Source words; `loadGraphHead`'s join goes | STEP-5, AC-1.1 |
| The `sources` effect names no graph for a title or type change | STEP-5, AC-2.3 |
| `RecordLinksGraph`: `source_ref` added, `source_type_id` removed | STEP-6, AC-3.1 |
| Swift: claims and record links keys follow `.source`; `graphSourceRefs` goes; fallbacks written once | STEP-7, AC-1.4, AC-3.2, AC-3.3 |
| `docs/catalog-caches.md` row, size figures, effects README, patterns doc | STEP-2, STEP-5, STEP-8 |

| Out | Home |
| --- | --- |
| The records tier and claims composed without loading whole graphs | S16-13 |
| The Evidence-graph catalog (graph shells, a Source's graphs) | S16-15 |
| Provenance records composed from cached records (the Citation and Subject halves of `sqlRecords`) | S16-30 |
| The claims read owning every graph label (the rest of P-22) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |
| The Source's cover and each Artifact's file on the record | the first PR that reads them from the record (D-3); not planned now |
| The Source date on the record | S15-05 |
| `.followsInputs` in place of the `.source` follows | S16-25 |
| The auto-reconciler reading credibility from the record (it runs on a write's graph) | not planned (D-4) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Graph copies Source words | `core/database/graphcache/evidence_graph.go` `EvidenceGraph` | Fields `SourceID`, `SourceRef`, `SourceTitle`, `SourceTypeID`. Only `SourceID` is needed by the graph itself (`loadTitleEdges`, claims grouping) |
| Graph head join | `core/database/graphfill/load.go` `loadGraphHead`, `graphHead` | `JOIN sources s` for `s.ref, s.title, s.source_type_id`; `loadEvidenceGraph` copies them onto the graph. `checkEvidenceGraph` compares whole graphs |
| Rename evicts graphs | `core/database/effects/registry.go` `"sources"` | `EvidenceGraph: graphsOfSource(union(onField("title", self()), onField("source_type_id", self())))`; `OnCommit` (`core/database/graphcache/commit.go`) calls `dropEvidenceGraphs(set.EvidenceGraphs)` |
| Effects doc | `core/database/effects/README.md` | "A Source row names every graph of the Source only on a write that changes `title` (`onField`), because a graph's search document and canvas hold the Source title" |
| Claims grouping | `core/database/persondetail/claims.go` `claims`, `ClaimGroup` | Group copies `m.graph.SourceRef`, `SourceTitle`, `SourceTypeID`; sorted by `Title` then `SourceRef` through `conclusionsummaries.SortByTitle` |
| Record links | `core/database/persondetail/recordlinks.go` `PersonRecordLinks`, `recordLinks` | `Graphs []graphcache.EvidenceGraph`, taken from the claims groups in their order |
| Record shells | `core/database/conclusiondetails/conclusiondetails.go` `sqlRecords` and its scan | Joins `artifacts a`, `sources src`, `autoreconciler.ProvenanceJoins`; scans `src.id`, `src.title`, `a.label`, `sg.key` and `ProvenanceColumns` (credibility relative to `standard`, `c.transcription_uncertain`) through `ProvenanceDest` |
| Credibility rule | `core/database/autoreconciler/autoreconciler.go` `ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` | Exported for `conclusiondetails`; also used by `sqlLoadCandidates` (the reconciler, on a write's transaction) |
| Grades | `core/database/effects/registry.go` `"source_credibility_grades": {None: true}`; `core/database/sourcecredibilitygrades/grades.go` `Install` | Seeded at create (`core/onboarding/ready.go`); no runtime write |
| Registry | `core/database/graphcache/cache.go` `RegisterWhole`, `WholeEntry`, `Whole.Get` | A `LifeGraph` whole entry is neither built nor stored on an `Over` graph (`Get` returns `ok == false`); Apply runs only when held |
| Precedent entry | `core/database/namealiases/namealiases.go` `Index` | A whole entry registered in the table's own package `init()`, with Build and a `reflect.DeepEqual` Check |
| Doc test | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry` | Blank-imports every owner package; holds `docs/catalog-caches.md` to `graphcache.EntryNames()` |
| Size benchmark | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` | One `measure` per cache on the large fixture |
| Handler | `api/ffi/handlers/person_detail.go` `ListPersonClaims`, `recordLinksProto` | Maps `ClaimGroup`'s three Source fields; maps `SourceId`, `SourceTitle`, `SourceTypeId` per record links graph |
| Wire | `api/proto/engine.proto` `RecordLinksGraph` | Fields 1–6: `graph_id`, `graph_ref`, `graph_name`, `source_id`, `source_title`, `source_type_id`. No `source_ref` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogRecordLinksGraph`; `macos/App/Platform/GoStore.swift` `mapPersonRecordLinks` | `sourceTypeID` unread anywhere in `macos/App/Features` |
| Ref map | `macos/App/Features/Conclusions/PersonSourcesContent.swift` `graphSourceRefs`; `PersonSourcesSection.swift` passes it; `PersonRecordLinksContent.swift` `init(links:graphLocations:graphSourceRefs:)` | Record links' Source label is `sourceTitle.trimmedNonEmpty ?? graphSourceRefs[graphID] ?? ""`; Sources' row title is `title.trimmedNonEmpty.map { .name($0) } ?? .ref(sourceRef)`; both write graph title as `name ?? ref ?? ""` |
| Key follows | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `personGraphsSection`, `.personClaims`, `.personRecordLinks` | Follow `.touchedHandles` (own id), `.evidenceGraphs` (graphs in value), property labels; no `.source`. `.conclusionDetail` follows `.source` (S16-09 precedent) |
| Tests | `core/database/graphfill/load_test.go` `TestRetitleDropsEvidenceGraph`, `TestOnlyATitleChangeDropsEvidenceGraph`, `TestSourceTypeChangeDropsEvidenceGraph` | Assert a title or type edit drops the graph |
| Tests | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve` | Asserts "title edit" names `def` and `jones` |
| Tests | `core/database/persondetail/claims_test.go`, `recordlinks_test.go` | Read `g.SourceRef`, `g.Title`, `g.SourceTypeID`, `rl.Graphs[i].SourceTitle` |
| Tests | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityQueryCountIsConstant`, `TestForEntityEvidence` | The count test drops the graph and counts the first and second read; the evidence test checks credibility keys |
| Tests | `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims`, `TestRecordLinksProto`, `TestGetPersonRecordLinks` | Read `GetSourceTypeId()` on record links graphs |
| Tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personRecordLinksRevalidate`; `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`; `macos/ProvenenciaTests/PersonSourcesContentTests.swift` | Claims key does not follow Sources; record links refreshes `== [.touchedHandles, .evidenceGraphs, .list(.propertyVocabulary)]`; record links content built with `graphSourceRefs` |

**Precedent.** [#789](https://github.com/mendahu/provenencia/pull/789) (S16-10) registered `namealiases.Index` as a whole entry in its own package: copy its `init()` shape, its `Check` by rebuild and `reflect.DeepEqual`, and its `docs/catalog-caches.md` row; do not copy its `LifeCatalog` lifetime (D-2). [#780](https://github.com/mendahu/provenencia/pull/780) (S16-09) added `.applyHiddenPolicy(followsValue: .source) { … }` to `.conclusionDetail`: copy it for the two keys, through `personGraphsSection` (D-10). [#793](https://github.com/mendahu/provenencia/pull/793) (S16-11) narrowed `sqlRecords` the same way: copy its shape of "the node supplies one half, SQL the rest".

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The entry lives in `core/database/sources` (file `records.go` *(new)*): `var Records graphcache.Whole[*RecordIndex]`, registered in `init()` as `"sources.records"`. `sources` already depends on `graphcache` (no cycle; `go list -deps` checked) | The table's package owns its cache, as `namealiases.Index` | A new `sourcerecords` package — a second home for one table |
| D-2 | Lifetime `graphcache.LifeGraph`. `RecordsOf(g)` *(new)* returns the held or built index; when `Whole.Get` returns `ok == false` (a write's graph) it returns `ErrRecordsOnWrite` *(new)*. No reader runs on a write's graph today (`persondetail` and `conclusiondetails` read `c.Graph()`) | Rule "A write's one-shot graph fills only what it reads": a whole Build of 3,000 Sources must not run per write; `LifeCatalog` would build once per write graph | A SQL fallback on `Over` — a second read path; `LifeCatalog` — builds per write |
| D-3 | `Record` holds `ID`, `Ref`, `Title`, `TypeID`, `CredibilityKey` ("" = no assessment), `Credibility` (int, relative to `standard`) and `ArtifactLabels map[string]string` (Artifact id → label, "" = none). No cover, no Artifact file, no date | Only fields a reader in this PR reads (rule "simpler code"); S15-05 adds the date, the first reader of cover or file adds those | The map's "cover" and "Artifacts with file" now — fields with no reader. The map's scope entry is updated (STEP-8) |
| D-4 | Credibility is read by SQL in Build and Apply with `autoreconciler.CredibilityColumn` *(new, exported)*: the relative sort-order expression now inside `ProvenanceColumns`. `ProvenanceColumns`, `ProvenanceJoins` and `ProvenanceDest` become unexported (`provenanceColumns`, `provenanceJoins`, `provenanceDest`), built from `CredibilityColumn`; their only caller left is `sqlLoadCandidates`. The vocabulary does not gain grades | One rule for "relative to standard", shared by the reconciler and the record; grades are written only by `Install` at create (`None: true` effect), so a value computed at Build cannot go stale in a session | Grades in `graphcache.Vocabulary` with a Go `relative` — a second copy of the reconciler's SQL rule. The map and design said "credibility grades join the vocabulary"; STEP-8 rewords both |
| D-5 | `RecordIndex` holds `bySource map[string]Record` and `sourceOfArtifact map[string]string` (Artifact id → Source id). Methods `Source(id []byte) (Record, bool)` and `Artifact(id []byte) (Record, string, bool)` (the Artifact's Source and label). No method iterates; readers only look up, so no order is defined or needed | The detail shells name a Citation's Artifact, not its Source | A slice per Source scanned per Artifact |
| D-6 | Build runs two queries: (1) `SELECT s.id, s.ref, COALESCE(s.title, ''), s.source_type_id, COALESCE(sg.key, ''), ` + `CredibilityColumn` + ` FROM sources s LEFT JOIN source_credibility_assessments sca ON sca.source_id = s.id LEFT JOIN source_credibility_grades sg ON sg.id = sca.credibility_grade_id`; (2) `SELECT id, source_id, COALESCE(label, '') FROM artifacts`. Apply runs the same two with `WHERE s.id IN (…)` / `WHERE source_id IN (…)` through `sqlbatch.Query` for the ids in `set.SourceScopes` | The design's two queries; `sqlbatch` bounds the IN list | One join — duplicates Source columns per Artifact |
| D-7 | Apply patches the held index in place: for each id in `set.SourceScopes`, delete the old record and its Artifacts from `sourceOfArtifact`, then insert what the two queries return (none for a deleted Source). It does not call `Whole.Set` (nothing derives from the entry). The index is shared and read-only between engine calls; `OnCommit` runs between calls, never during a reader | O(Sources named) per commit; copy-on-write would be O(all Sources) per commit | A full rebuild per commit; a copy per commit |
| D-8 | `EvidenceGraph` loses `SourceRef`, `SourceTitle`, `SourceTypeID`; keeps `SourceID`. `graphHead` keeps `ref`, `name`, `sourceID`; the head query reads `evidence_graphs` alone | Rule "A Source's words live once"; the compiler finds every reader | Keeping them unread — two copies again |
| D-9 | The `"sources"` effect drops its `EvidenceGraph:` field. `Search` keeps `{Kind: DocEvidenceGraph, Path: onField("title", graphsOfSource(self()))}` (a graph's search document still holds the Source title) | No reader takes Source words from a loaded graph after STEP-5; the Swift canvas (`.evidenceGraph` key) reads no Source words (its five reads return none; checked) | Keeping it for safety — the rename-evicts-graphs cost this PR removes |
| D-10 | `CatalogQueryDescriptor.personGraphsSection` gains a `sourceIDs: @escaping @Sendable (Value) -> Set<String>` parameter and adds `.applyHiddenPolicy(followsValue: .source, ids: sourceIDs)` after the `.evidenceGraphs` entry. Claims pass `Set(groups.map(\.sourceID))`; record links pass `Set(links.graphs.map(\.sourceID))` | One place for both keys; refreshes order: `[.touchedHandles, .evidenceGraphs, .source, .list(.propertyVocabulary)]` | A `.source` entry written in each descriptor |
| D-11 | `RecordLinksGraph` keeps `source_id` (now read by D-10's follow), adds `string source_ref = 7`, removes `source_type_id` with `reserved 6; reserved "source_type_id";`. Swift `CatalogRecordLinksGraph` gains `sourceRef`, loses `sourceTypeID` | The `.source` follow needs the Source id from the value, so `source_id` is no longer unread | Dropping `source_id` as the map said — the record links key could then not follow its Sources. The map row and scope entry are updated (STEP-8) |
| D-12 | `ClaimGroup` becomes `{Source sources.Record; Claims []Claim}`; `PersonRecordLinks.Graphs` becomes `[]SourcedGraph` *(new)* `{Graph graphcache.EvidenceGraph; Source sources.Record}`. Ordering is unchanged: groups by `Source.Title`, then `Source.Ref` | One Source value per answer instead of three copied fields | Keeping three fields filled from the record |
| D-13 | A loaded graph whose `SourceID` has no record, or a detail record whose Citation's Artifact has none, is an error (`fmt.Errorf("persondetail: graph %x names Source %x, which has no record", …)` and the same shape in `conclusiondetails`). `evidence_graphs.source_id` and `artifacts.source_id` cascade on delete, and Apply runs before the next read | A broken invariant, not a case to hide (precedent `personEventProto`) | Skipping the group — hides a cache bug |
| D-14 | Swift fallbacks: `PersonSourcesContent.graphTitle(name:ref:) -> String` *(new, static)* = `name.trimmedNonEmpty ?? ref.trimmedNonEmpty ?? ""` and `PersonSourcesContent.sourceTitle(title:ref:) -> PVTitleSource` *(new, static)* = `title.trimmedNonEmpty.map { .name($0) } ?? .ref(ref)`. `PersonRecordLinksContent` calls both (`sourceTitle(…).text`); its node `label` (`ref ?? name ?? ""`) is a different rule and stays | `PersonSourcesContent` is already "the one builder of a graph's location" that record links reads | A `PVTitleSource` extension in the design system — feature words, not a kit rule |
| D-15 | `TestForEntityQueryCountIsConstant` warms the records (`sources.RecordsOf(g)`) after `g.Vocabulary()` and before it starts counting | Build runs on first read; the test compares two warm reads | Counting the Build into the first read |

**Rules this plan bends.** None. D-3, D-4 and D-11 narrow the map's scope entry and the design's Entity records line; neither is a working rule. STEP-8 updates both, and the report to the user names the three.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/sources/records.go` *(new)* `Records`, `Record`, `RecordIndex`, `RecordsOf`, `ErrRecordsOnWrite` | whole registry entry, types, func | `core/database/namealiases/namealiases.go` `Index`; `sources.Source` | Searched `RegisterWhole`, `sources.Source`, `GetBySource`: `Source` is the edit row (description, cover mode, date, public) read one at a time; no cache holds Sources |
| `autoreconciler.CredibilityColumn` *(new)* | const | `autoreconciler.ProvenanceColumns` | Split out of it, not beside it: the one rule, two callers |
| `persondetail.SourcedGraph` *(new)* | type | `graphcache.EvidenceGraph` | The graph no longer carries Source words (D-8) |
| `RecordLinksGraph.source_ref` *(new)* | proto field | `PersonClaimGroup.source_ref` | Same name and meaning as the claims wire |
| `PersonSourcesContent.graphTitle(name:ref:)`, `sourceTitle(title:ref:)` *(new)* | static funcs | the inline expressions at `PersonSourcesContent.swift` and `PersonRecordLinksContent.swift` | Searched `trimmedNonEmpty.map { .name`, `titleOrRef`: no shared helper exists |
| `core/database/sources/records_test.go` *(new)*, `records_bench_test.go` *(new)* | tests | `core/database/namealiases` tests | New entry |

**Builds on.** The S16-10 registry (`RegisterWhole`, Apply on `effects.Set`, Check under `PROVENENCIA_GRAPH_VERIFY`); `effects.Set.SourceScopes` (16 entries already set it, including `artifacts` with old and new `source_id` on a move, `source_credibility_assessments`, and `sources` itself); `sqlbatch.Query`; S16-09's `.source` follow.

**Deletes.**
- `EvidenceGraph.SourceRef`, `SourceTitle`, `SourceTypeID` and their doc comment lines; `graphHead.sourceRef`, `sourceTitle`, `sourceTypeID`; the `JOIN sources s` in `loadGraphHead`.
- `ClaimGroup.SourceID`, `SourceRef`, `Title`, `SourceTypeID` (replaced by `Source`).
- In `sqlRecords`: `JOIN artifacts a`, `JOIN sources src`, `autoreconciler.ProvenanceJoins`, the columns `a.id, src.id, COALESCE(src.title, ''), COALESCE(a.label, ''), COALESCE(sg.key, '')` and `ProvenanceColumns`; `c.artifact_id` and `c.transcription_uncertain` take their place.
- The exported names `autoreconciler.ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` (unexported, D-4).
- The `"sources"` effect's `EvidenceGraph:` line.
- `RecordLinksGraph.source_type_id` (proto field 6, reserved); `CatalogRecordLinksGraph.sourceTypeID`; its line in `GoStore.mapPersonRecordLinks` and in fixtures.
- `PersonSourcesContent.graphSourceRefs` (property and builder), the `graphSourceRefs:` parameter of `PersonRecordLinksContent.init` and the argument in `PersonSourcesSection.swift`.
- The "title edit drops the graph" assertions in the three `graphfill` tests (inverted, T-5).

**Next adds.** S15-05 adds the date to `Record` and its Build/Apply query (one column). S16-13 composes claims from the records tier plus `RecordsOf`. S16-30 composes a provenance record's Source half from `RecordIndex.Artifact` and deletes what remains of `sqlRecords`. S15-24's move needs no change here: the Artifact's old and new Source are both in `SourceScopes`. S16-25 replaces D-10's `.source` entry with `.followsInputs`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Claims and record links keys revalidate on a write naming a cited Source | `CatalogQueryRegistry.swift` `personGraphsSection` | Same `applyHiddenPolicy(followsValue:)` path `.conclusionDetail` uses; a hidden key marks stale; before the value arrives `followsValue` matches any id (documented behavior) |
| A graph or Artifact with no record | `persondetail.claims`, `conclusiondetails` record scan | Error (D-13); the section shows S16-04's retry notice |
| `RecordsOf` on a write's graph | callers on `graphcache.Over` | None today; `ErrRecordsOnWrite` (AC-2.2) |

### Scale

At the map's Batch 2 Scale (3,000 Sources, 10,000 Evidence graphs, 50,000 Observations; large fixture at 4,000 and 16,000 persons):

- **Build** (first read of a session, or after `Drop`): two full scans, 3,000 Source rows and every Artifact row (est. 30,000). Est. 5–15 ms, 1.5–4 MB. Measured by `BenchmarkSourceRecords` *(new)* at 3,000 Sources × 10 Artifacts; stop if over 50 ms or 5 MB (Stop conditions).
- **Apply per commit:** runs only while held; two queries per batch of Sources named in `SourceScopes` (usually one Source). Work per Source is its Artifact count. Crafted: a Source with 10,000 Artifacts reloads 10,000 rows on every write inside it; measured by the same benchmark (`apply-crafted`), budget 20 ms. A write naming every Source (none exists today; a Source type delete is refused while in use) costs one Build.
- **Evidence graphs:** a rename now reloads zero graphs (today: every loaded graph of the Source, up to 64, each a full graph load on its next read).
- **Swift:** the claims and record links keys now also revalidate on writes in a cited Source that touch no claimed graph (another graph's Observation, an Artifact label, a credibility change, a note). While the Sources section is visible that is one `ListPersonClaims` and one `GetPersonRecordLinks` per such write; hidden keys mark stale. Both reads are within S13-27's 50-graph budget; S16-25 narrows the follow to the value's inputs.
- **Memory:** `docs/catalog-caches.md` gets the entry's size at 4,000 and 16,000 persons (`BenchmarkCacheSizes`) and at 3,000 Sources; the 250 MB ceiling holds (est. under 5 MB).

### Edges

- **Assumptions.** `SourceScopes` names every Source whose record a write changes: `sources` (self, create/update/delete), `artifacts` (field `source_id`, old and new on an update), `source_credibility_assessments` (field `source_id`), `files` (through `artifacts.file_id`, harmless here). T-2 holds create, delete, relabel, move and credibility. Grade rows change only at `Install` before any session read (D-4); if a runtime grade write appears, Stop.
- **Neighbors.** A description, cover, note, metadata or date edit reloads the record (same words), changes nothing visible, and evicts no graph (today a description edit already evicts none; a title or type edit did). A Subject or Observation write in a graph still drops that graph (unchanged, T-5's control case).
- **Removed inputs.** A deleted Source: Apply removes its record and its Artifacts; its graphs are deleted by cascade and name themselves in `Set.EvidenceGraphs`, so no loaded graph points at a missing record. A deleted Artifact: its Source is in `SourceScopes`; the reload omits it. T-2.
- **Real data.** Not table-driven by vocabulary; the credibility column runs over the shipped grades (`low_trust`, `standard`, `high_trust`, and the rest `Install` seeds) exactly as the reconciler does today.
- **Untrusted size.** Sources and Artifacts per catalog: Build is linear and bounded by the table sizes; Artifacts per Source multiplies Apply (crafted 10,000, budget above); `SourceScopes` length is bounded by `sqlbatch`'s chunking. No nested loop over Sources.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sources` | `records.go` *(new)*: `type Record struct { ID []byte; Ref, Title string; TypeID []byte; CredibilityKey string; Credibility int; ArtifactLabels map[string]string }`; `type RecordIndex struct { bySource map[string]Record; sourceOfArtifact map[string]string }` with `Source(id []byte) (Record, bool)` and `Artifact(id []byte) (Record, string, bool)`; `var Records graphcache.Whole[*RecordIndex]`; `init()` registers Name `"sources.records"`, `Lifetime: graphcache.LifeGraph`, Build = `loadRecords(graphcache.Querier(g), nil)`, Apply = D-7 over `set.SourceScopes` (returns early when empty), Check = rebuild and `reflect.DeepEqual`; `func RecordsOf(g *graphcache.Graph) (*RecordIndex, error)`; `var ErrRecordsOnWrite = errors.New("sources: records are not read on a write's graph")` |
| `core/database/autoreconciler` | `const CredibilityColumn` *(new)*; `provenanceColumns = CredibilityColumn + ",\n c.transcription_uncertain"`; `provenanceJoins`, `provenanceDest` unexported (D-4) |
| `core/database/persondetail` | `claims`: `recs, err := sources.RecordsOf(c.Graph())`; group by `m.graph.SourceID`, `Source: rec` (D-12, D-13); sort by `g.Source.Title`, `g.Source.Ref`. `recordLinks`: `Graphs []SourcedGraph` from the groups; `linksBetween` takes `[]graphcache.EvidenceGraph` mapped from them |
| `core/database/conclusiondetails` | `sqlRecords` selects `c.artifact_id, c.transcription_uncertain` in place of the dropped columns; the scan looks up `recs.Artifact(artifactID)` for `SourceID`, `SourceTitle`, `ArtifactLabel`, `CredibilityKey`, `Provenance.Credibility`; `Provenance.Uncertain` from the column; error per D-13 |
| `core/database/graphfill` | D-8 |
| `core/database/graphcache` | D-8 (`EvidenceGraph` fields and comment) |
| `core/database/effects` | D-9; README sentence rewritten (STEP-5) |

Errors: `sources.ErrRecordsOnWrite`; D-13's two invariant errors. No new `apperr` code.

### FFI

`api/proto/engine.proto` `RecordLinksGraph` per D-11; update its comment ("one Evidence graph of the person and its Source, by id and ref; `source_title` is empty for an untitled Source"). Handler `api/ffi/handlers/person_detail.go`: `ListPersonClaims` maps `g.Source.ID`, `.Ref`, `.Title`, `.TypeID`; `recordLinksProto` maps `GraphId`, `GraphRef`, `GraphName` from `g.Graph`, `SourceId`, `SourceRef`, `SourceTitle` from `g.Source`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Record links graph model | `CatalogRecordLinksGraph` | `macos/App/Platform/GenealogyStore.swift` | `sourceRef` added, `sourceTypeID` removed (D-11); `GoStore.mapPersonRecordLinks`, `PersonSourcesPreviewData.swift`, `FakeStore` seeds follow |
| Claims and record links follows | `CatalogQueryDescriptor.personGraphsSection` | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` | D-10; update the `.personRecordLinks` comment ("a Source write names its graph" → "a Source write names the Source") |
| Fallbacks | `PersonSourcesContent.graphTitle`, `sourceTitle` | `macos/App/Features/Conclusions/PersonSourcesContent.swift` | D-14 |
| Record links labels | `PersonRecordLinksContent` | `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` | `init(links:graphLocations:)`; Source label `PersonSourcesContent.sourceTitle(title: graph.sourceTitle, ref: graph.sourceRef).text` |

States: unchanged. Accessibility: unchanged (same strings, same inputs).

### Strings

None — no key added or removed; `recordLinksNode*` strings receive the same Source label as today.

## 7. Steps

**STEP-1 — One credibility column**
Files: `core/database/autoreconciler/autoreconciler.go`. Skill: `add-catalog-query`.
Do: Before any change, record `BenchmarkCacheSizes` on `main` in the ship log. Add `CredibilityColumn` and build `ProvenanceColumns` from it (still exported; STEP-4 unexports the trio once `conclusiondetails` stops using it).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/... ./core/database/conclusiondetails/...` passes unchanged.

**STEP-2 — The Source records entry**
Files: `core/database/sources/records.go` *(new)*, `core/database/sources/records_test.go` *(new)*, `core/database/sources/records_bench_test.go` *(new)*, `core/catalogsession/cache_doc_test.go` (blank import `core/database/sources`), `core/database/promotealign/bench_test.go` (`measure("sourceRecords", …)` after `"nodes"`), `docs/catalog-caches.md`. Skill: `add-catalog-query`.
Do: D-1 to D-7. Add the `sources.records` row to "Registry entries" (owner `core/database/sources/records.go` *(new)* · `Records`; key one value; Fill Build, two queries, on first read; Updated by Apply reloading the Sources in `effects.Set.SourceScopes`; Lifetime graph; Cap none; sizes from the benchmark).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sources/... ./core/catalogsession/...` passes (T-1, T-2, T-3, `TestCacheDocListsEveryEntry`); `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'SourceRecords' ./core/database/sources/` prints build ms, heap MB, apply ms within budget.

**STEP-3 — Claims and record links read the record**
Files: `core/database/persondetail/claims.go`, `recordlinks.go`, `claims_test.go`, `recordlinks_test.go`, `recordlinks_perf_test.go` (if it reads Source fields), `api/ffi/handlers/person_detail.go` (Go types only; the wire is STEP-7). Skill: `use-catalog-session`.
Do: D-12, D-13. The handler maps `g.Source.*` into today's proto fields (`SourceTypeId` still filled from `g.Source.TypeID` until STEP-6).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/handlers/...` passes (T-4).

**STEP-4 — The record shells take Source words from the record**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `conclusiondetails_test.go`.
Do: Narrow `sqlRecords` and its scan per Go core; D-13; D-15; add `SQLRecordsForTest` *(new)* to `core/database/conclusiondetails/export_test.go`. Then unexport `ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` in `core/database/autoreconciler/autoreconciler.go` (D-4).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/... ./core/database/autoreconciler/...` passes (T-6, T-7, `TestForEntityEvidence` unchanged); `rg -n "ProvenanceColumns|ProvenanceJoins|ProvenanceDest" --type go` has no hit.

**STEP-5 — Graphs stop copying Source words, and a rename names no graph**
Files: `core/database/graphcache/evidence_graph.go`, `core/database/graphfill/load.go`, `core/database/graphfill/load_test.go`, `core/database/effects/registry.go`, `core/database/effects/evidence_graphs_test.go`, `core/database/effects/README.md`.
Do: D-8 and D-9 together, one commit. Rename `TestRetitleDropsEvidenceGraph` → `TestRetitleKeepsEvidenceGraph`, `TestSourceTypeChangeDropsEvidenceGraph` → `TestSourceTypeChangeKeepsEvidenceGraph`, `TestOnlyATitleChangeDropsEvidenceGraph` → `TestSourceEditKeepsEvidenceGraph`, and rewrite them per T-5. In `TestEvidenceGraphsResolve` the "title edit" set names no graph and its `SourceScopes` holds the Source; update its doc comment. README: replace the sentence quoted in Current state with "A Source row names no graph: no loaded graph holds Source words, which live in the Source records entry (`sources.Records`) that `SourceScopes` updates; a title edit still reprojects the Source's graphs' search documents (`DocEvidenceGraph`)."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/database/graphfill/... ./core/database/graphcache/... ./core/database/persondetail/...` passes, the completeness test included (T-5, T-8); `rg -n "SourceTitle|SourceRef|SourceTypeID" core/database/graphcache core/database/graphfill` has no hit.

**STEP-6 — The record links wire carries the Source ref**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/person_detail.go`, `api/ffi/handlers/person_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-11; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes (T-9).

**STEP-7 — The app follows Sources and writes each fallback once**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift` (if it builds `CatalogRecordLinksGraph`), `macos/App/Features/Conclusions/PersonSourcesPreviewData.swift`, `PersonSourcesContent.swift`, `PersonSourcesSection.swift`, `PersonRecordLinksContent.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, and the four test files in Current state. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-10, D-11, D-14.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/PersonRecordLinksContentTests -only-testing:ProvenenciaTests/PersonSourcesContentTests` passes (T-10 – T-12), one run at a time.

**STEP-8 — Docs**
Files: `docs/catalog-caches.md` (sizes; the `evidenceGraphs` typed-field row's "Updated by" unchanged), `docs/macos-client-patterns.md` ("Conclusion keys": claims and record links follow the Sources in their value through `.source`; the long rule-2 line's "the Evidence graph, list, detail and person-claims keys follow `evidenceGraphs`" stays true), `docs/deployment-plan/spike-16/cache-architecture.md` (Entity records: Source records hold no cover or file yet, credibility is read by the reconciler's column, `RecordLinksGraph` keeps `source_id`), `docs/deployment-plan/spike-16/deployment-plan.md` (S16-12 map row and scope entry per D-3, D-4, D-11; status), `completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-12-source-records.md` passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sources/records_test.go` `TestRecordsBuild` *(new)* | Table: Source "Register" (low_trust, Artifacts "Page 1", unlabelled), untitled Source, Source with no assessment → `Source(id)` returns ref, title ("" for untitled), type id, key ("low_trust" / ""), `Credibility` (< 0 / 0); `Artifact(page1)` → Register, "Page 1"; unknown id → false | AC-1.3, D-3, D-4, D-5 |
| T-2 | Go | `records_test.go` `TestRecordsApply` *(new)* | Records held; table of writes through `writes.Run`: rename, type change, Artifact create, relabel, delete, move A→C, credibility set and cleared, Source delete → after each, `RecordsOf` equals a fresh build (`Check` returns nil) and the named fields match; a description edit leaves the record equal | AC-2.1, D-7, Edges |
| T-3 | Go | `records_test.go` `TestRecordsOnWrite` *(new)* | `RecordsOf(graphcache.Over(tx))` → `ErrRecordsOnWrite`; query counter shows no `sources` query | AC-2.2, D-2 |
| T-4 | Go | `core/database/persondetail/claims_test.go` (updated), `recordlinks_test.go` (updated) `TestClaimsAfterSourceEdit` *(new)* | Person with a claim in G1 of "Register"; rename to "Parish register" and change type → `Claims` group `Source.Title`/`TypeID` new; `RecordLinks` `Graphs[0].Source.Title` new; existing order tests unchanged with `Source.Title` | AC-1.2, AC-1.5, D-12 |
| T-5 | Go | `core/database/graphfill/load_test.go` `TestRetitleKeepsEvidenceGraph`, `TestSourceTypeChangeKeepsEvidenceGraph`, `TestSourceEditKeepsEvidenceGraph` (renamed) | Two loaded graphs of one Source, each with a memo; rename, type change, description edit → both stay loaded, memos read back; control: a Subject create in G1 drops G1 only | AC-1.1, D-8, D-9 |
| T-6 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityAfterSourceEdit` *(new)* | AC-1.3's before and after | AC-1.3 |
| T-7 | Go | `conclusiondetails_test.go` `TestForEntityQueryCountIsConstant` (updated, D-15) | After `g.Drop()` and `g.Vocabulary()`, counting: `RecordsOf` → 2 queries; small and big Person → equal counts; `SQLRecordsForTest()` contains none of `sources`, `artifacts`, `source_credibility_` | AC-2.4 |
| T-8 | Go | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve` (updated) | Title edit → `EvidenceGraphs` empty, `SourceScopes` = [src], `Search` still holds `DocEvidenceGraph` for def and jones | AC-2.3 |
| T-9 | Go | `api/ffi/handlers/person_detail_test.go` `TestGetPersonRecordLinks`, `TestRecordLinksProto` (updated) | Every graph: `GetSourceRef()` = the Source's ref, `GetSourceId()` = its id, `GetSourceTitle()` = title or ""; `TestListPersonClaims` unchanged | AC-3.1 |
| T-10 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs` (updated) | Claims value cites "s1": `revalidatedKeys(sourceScopes: ["s1"])` == {claimsKey, detailKey}; `["s9"]` → [] | AC-1.4, D-10 |
| T-11 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personRecordLinksRevalidate` (updated) | `sourceScopes: ["s1"]` → [key]; `["s9"]` → []; refreshes `== [.touchedHandles, .evidenceGraphs, .source, .list(.propertyVocabulary)]` | AC-1.4, D-10 |
| T-12 | Swift | `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift` (updated), `PersonSourcesContentTests.swift` `titleFallbacks` *(new)* | Graph `sourceTitle: ""`, `sourceRef: "SRC-1"` → help names "SRC-1"; `sourceTitle(title: " ", ref: "SRC-2")` → `.ref("SRC-2")`; `graphTitle(name: "", ref: "EG-1")` → "EG-1"; existing content tests unchanged | AC-3.2, AC-3.3, D-14 |
| T-13 | Go | `core/database/sources/records_bench_test.go` `BenchmarkSourceRecords` *(new)* | 3,000 Sources × 10 Artifacts: build ms and heap; Apply of one ordinary Source and of a 10,000-Artifact Source | D-6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/graphfill` (check compares whole graphs), `core/database/persondetail` order tests, `core/database/conclusiondetails` count test, `PROVENENCIA_GRAPH_VERIFY` runs (the new Check), `CatalogQueryRegistryTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-caches.md (row and sizes), docs/macos-client-patterns.md (Conclusion keys), core/database/effects/README.md (the Source row sentence), cache-architecture.md (Entity records)
- [ ] Ship log records BenchmarkCacheSizes and BenchmarkSourceRecords before and after, on the same machine
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person whose records sit in two Evidence graphs of Source "Register". Open the person on Sources: the row reads "Register"; hover a record links node: its help names "Register".
2. Open the person's Vitals and a field's provenance sheet: the record names "Register".
3. Open Source "Register" and rename it "Parish register". Go Back to the person: Sources row, record links help and provenance sheet read "Parish register".
4. Open one of the two graphs: the canvas shows as before (no Source word on it changes).
5. Rename an untitled Source's graph's Source to "" (clear the title): the Sources row and the record links help read its ref.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A reader of `EvidenceGraph` Source words outside the files found | Deleting the fields makes the compiler list them (STEP-5) |
| The claims and record links keys reload more often in Swift | Stated in Scale; hidden keys mark stale; S16-25 narrows |
| A grade's sort order changes at runtime, leaving `Credibility` stale | No such write exists (`None: true`); Stop condition below |
| In-place Apply surprises a reader holding the index | Readers hold it within one engine call; `OnCommit` runs between calls (D-7) |

**Stop and report if:**

- A row of §4 is no longer true at your base (notably `sqlRecords`, `loadGraphHead`, the `"sources"` effect).
- Any Go code outside `persondetail`, `graphfill`, `api/ffi/handlers/person_detail.go` fails to compile after STEP-5 because it reads `EvidenceGraph.SourceRef`, `SourceTitle` or `SourceTypeID` (a reader this plan missed).
- A Swift view outside the files in STEP-8 reads `CatalogRecordLinksGraph.sourceTypeID`.
- `rg -n "source_credibility_grades" --type go` shows a write other than `sourcecredibilitygrades.Upsert` and `Install`, or a caller of them outside create and onboarding (D-4's premise).
- A caller of `persondetail.Claims`, `persondetail.RecordLinks` or `conclusiondetails.ForEntity` runs on a `graphcache.Over` graph (D-2's premise).
- `BenchmarkSourceRecords` shows Build over 50 ms or 5 MB at 3,000 Sources × 10 Artifacts, or Apply over 20 ms for the 10,000-Artifact Source.
- The Swift Evidence graph canvas, Promote or the Citation composer shows a Source title read from the `.evidenceGraph` key's value (D-9's premise).

## 12. Open questions

None.

## 13. PR

**Title:** S16-12 — Source records in the graph cache
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-12-source-records.md`.
