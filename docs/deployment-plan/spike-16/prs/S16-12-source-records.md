# S16-12 — Source records in the graph cache

**Status:** Planned
**Spike / slice:** Spike 16 · Slice 2b — Entity records
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: the Sources section, the record links diagram and the provenance sheet draw the same words; only where they are read from moves)
**Board:** —
**Depends on:** S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789), the cache registry this entry joins), S16-01 (landed [#756](https://github.com/mendahu/provenencia/pull/756), same `claims.go` and `recordlinks.go`); S16-11 (landed [#793](https://github.com/mendahu/provenencia/pull/793)) left the outcome record shells as the one SQL read this PR narrows
**Blocks:** S16-13 (composes claims from the records tier and these records), S16-30 (a provenance record composes the Source record), S16-25 (replaces this PR's `.source` follows with `.followsInputs`), S15-05 (adds the Source date to the record), S15-24 (the move changes an Artifact's Source; this entry's Apply sees both)
**Branch:** `s16-12-source-records`
**Version bump:** none — no change a researcher can see; the wire change ships inside the one app binary ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md) (Refresh), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/ffi-handlers.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [design § Entity records](../cache-architecture.md#entity-records); [`docs/catalog-caches.md`](../../../catalog-caches.md); [`core/database/effects/README.md`](../../../../core/database/effects/README.md); [`macos-client-patterns.md`](../../../macos-client-patterns.md) "Conclusion keys"; [`source-layer-data-model.md`](../../../source-layer-data-model.md) (a Source's title is required); audit P-22 ([`person-page-audit.md`](../person-page-audit.md)); working rules: "(Batch 2) The point of this batch is simpler code", "(Batch 2) Cache entities and results per key; compose answers", "(Batch 2) One cache registry, two ways to fill", "(Batch 2) Every cache names the signal that updates it, and updates exactly", "(Batch 2) Apply states its fan-out", "(Batch 2) A write's one-shot graph fills only what it reads", "(Batch 2) A Source's words live once", "(Batch 2) Warm only what every proposal needs; no background re-warm", "(Batch 2) Memory is budgeted", "(Batch 2) Measure growth"; goal bar 11 (the rename clause)
**Planned:** 2026-11-12 against `main` @ `8d34930`
**Review:** [`S16-12-source-records.review.md`](S16-12-source-records.review.md) — Fits — Ready with changes, 2 rounds; every finding answered; To user answered 2026-11-12

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

A Source's words (ref, title, type, credibility, its Artifacts' labels) are copied today into every loaded Evidence graph and joined again into the provenance sheet's record shells. So a Source rename evicts every loaded graph of that Source and drops their memos, and the next Sources section or Promote read reloads them whole. After this PR the graph cache holds one record per Source, built on first read and patched from `effects.Set.SourceScopes`; the Sources section, the record links diagram and the provenance sheet compose Source words from it; a rename reloads one Source record and no Evidence graph. The record links wire drops its one unread field, and the app drops the title fallbacks that guard states the catalog refuses (an empty Source title, an empty graph name).

**Why now.** S16-13 (the records tier) and S16-30 (provenance composed) compose from Source records; S15-05 adds the Source date to the record and S15-24's move changes an Artifact's Source. Both spikes wait on the record existing. Audit P-22's wire half rides here because this PR changes the same message (user, 2026-11-08); planning found its "missing `source_ref`" unneeded, since a Source always has a title (user, 2026-11-12).

**What stays.** Every word, order and label the researcher sees; the claims wire (`PersonClaimGroup`); the Evidence-graph store's 64-graph LRU and its drop on `Set.EvidenceGraphs` for Subject, Observation, claim and Citation writes; the auto-reconciler's own candidate SQL (a write path); search documents (`DocEvidenceGraph` still reprojects on a title edit); `.evidenceGraphs` follows on the claims and record links keys; the record links node label (graph ref, else name; a graph's ref may be empty).

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
| AC-2.1 | Source records held; each write of T-2's table commits (rename, type change, Artifact create, relabel, Artifact delete, credibility set, an Observation edit in one of the Source's graphs, then a delete of a second Source with no Artifact or Subject) | After each commit `g.Verify()` passes (the held index equals a fresh build); the deleted Artifact and the deleted second Source resolve to nothing; an entry registered `DerivedFrom: Records` is emptied by each commit that names a Source |
| AC-2.2 | `sources.RecordsOf` on a write's graph (`graphcache.Over(tx)`) | It returns `sources.ErrRecordsOnWrite` and builds nothing; on a nil graph it returns `database.ErrClosed` |
| AC-2.3 | A Source title edit resolved by `effects.Resolve` | `Set.EvidenceGraphs` is empty; `Set.SourceScopes` holds the Source; the `DocEvidenceGraph` search docs still name every graph of the Source |
| AC-2.4 | Records warmed (`sources.RecordsOf`), then `conclusiondetails.ForEntity` reads a small and a big Person | The graph query counts are equal (as today), and `sqlRecords` names no `sources`, `artifacts` or `source_credibility_` table (`SQLRecordsForTest` *(new)*) |

**US-3.** As a contributor, the record links wire carries no field nobody reads, and the Person page's Sources section and record links diagram show a Source's title and a graph's name as stored, with no fallback for a state the catalog refuses.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `GetPersonRecordLinks` for a person in graphs on two Sources | Every `RecordLinksGraph` has `source_id` and `source_title` of its Source; the message has no `source_type_id` (field 6 reserved) |
| AC-3.2 | `PersonRecordLinksContent` over a graph named "Page 4" with no ref, on Source "Census" | Node label "Page 4", help "Page 4 — Census" |
| AC-3.3 | `PersonSourcesContent` over the test fixtures, every Source titled and every graph named | Row titles `.name(title)`, line titles the graph name, accessibility labels as today for titled Sources |

## 3. Scope

| In | Covered by |
| --- | --- |
| `sources.Records` whole registry entry: Build (two queries), Apply on `Set.SourceScopes` through `Records.Set`, Check, graph lifetime | STEP-2, AC-2.1, AC-2.2 |
| Credibility read through one shared SQL column owned by the grades (`sourcecredibilitygrades.RelativeOrderColumn`) | STEP-1, STEP-2, AC-1.3 |
| Claims grouping and record links take Source words from the record | STEP-3, AC-1.2, AC-1.5 |
| The detail read's record shells drop the Source, Artifact and credibility joins | STEP-4, AC-1.3, AC-2.4 |
| `EvidenceGraph` stops copying Source words; `loadGraphHead`'s join goes | STEP-5, AC-1.1 |
| The `sources` effect names no graph for a title or type change | STEP-5, AC-2.3 |
| `RecordLinksGraph.source_type_id` removed | STEP-6, AC-3.1 |
| Swift: claims and record links keys follow `.source`; `graphSourceRefs` and the Source-title and graph-name fallbacks go | STEP-7, AC-1.4, AC-3.2, AC-3.3 |
| Docs: `docs/catalog-caches.md` row and sizes, effects README, `add-audit-scope`, patterns doc, map and design | STEP-2, STEP-5, STEP-8 |

| Out | Home |
| --- | --- |
| The records tier and claims composed without loading whole graphs | S16-13 |
| The Evidence-graph catalog (graph shells, a Source's graphs) | S16-15 |
| Provenance records composed from cached records (the Citation and Subject halves of `sqlRecords`) | S16-30 |
| The claims read owning every graph label (the rest of P-22) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) |
| The Source's cover and each Artifact's file on the record | the first PR that reads them from the record (D-3; user, 2026-11-12) |
| The Source date on the record | S15-05 |
| An Artifact moved between Sources, tested against this entry | S15-24 (no write moves one before it) |
| A narrower update signal than `SourceScopes` (a keyed read fired only by Source, Artifact and credibility rows) | not planned (user, 2026-11-12) |
| The provenance sheet's empty-Source-title fallback (`CitationLineDisplay.sourceTitle(for:)`, `Parts.titleIsRef`, read by `ConclusionGallerySection.swift` and `PVProvenanceRecordTable.swift`) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (STEP-8 adds the line; a separate view and component) |
| `.followsInputs` in place of the `.source` follows | S16-25 |
| The auto-reconciler reading credibility from the record (it runs on a write's graph) | not planned (D-4) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Graph copies Source words | `core/database/graphcache/evidence_graph.go` `EvidenceGraph` | Fields `SourceID`, `SourceRef`, `SourceTitle`, `SourceTypeID`. Only `SourceID` is needed by the graph itself (`loadTitleEdges`, claims grouping) |
| Graph head join | `core/database/graphfill/load.go` `loadGraphHead`, `graphHead` | `JOIN sources s` for `s.ref, s.title, s.source_type_id`; `loadEvidenceGraph` copies them onto the graph. `checkEvidenceGraph` compares whole graphs |
| Rename evicts graphs | `core/database/effects/registry.go` `"sources"` | `EvidenceGraph: graphsOfSource(union(onField("title", self()), onField("source_type_id", self())))`; `OnCommit` (`core/database/graphcache/commit.go`) calls `dropEvidenceGraphs(set.EvidenceGraphs)` |
| `SourceScopes` | `core/database/effects/registry.go` | 19 entries set `SourceScope`: `sources` (self), `artifacts` (`source_id`), `source_credibility_assessments` (`source_id`), and every evidence row (`observations`, `citations`, `subjects`, `evidence_graphs`, the notes, the owned values, `files`). It is the audit "last activity" scope ([`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md)) |
| Effects doc and tests | `core/database/effects/README.md`; `core/database/effects/registry_completeness_test.go` `TestSourceEvidenceIsNotActivity` (`{"sources", true, true}`), `TestEvidenceGraphPathsEndAtGraphs` (`want` lists `"sources"`) | README: "A Source row names every graph of the Source only on a write that changes `title` (`onField`), because a graph's search document and canvas hold the Source title …" |
| Claims grouping | `core/database/persondetail/claims.go` `claims`, `ClaimGroup` | Group copies `m.graph.SourceRef`, `SourceTitle`, `SourceTypeID`; sorted by `Title` then `SourceRef` through `conclusionsummaries.SortByTitle` |
| Record links | `core/database/persondetail/recordlinks.go` `PersonRecordLinks`, `recordLinks` | `Graphs []graphcache.EvidenceGraph`, taken from the claims groups in their order |
| Record shells | `core/database/conclusiondetails/conclusiondetails.go` `sqlRecords`, `loadRecords`, `fieldsOf` | `sqlRecords` joins `artifacts a`, `sources src`, `autoreconciler.ProvenanceJoins`; the scan in `loadRecords(q, toponym, nodes, byProp, leadRows)` reads `src.id`, `src.title`, `a.label`, `sg.key` and `ProvenanceColumns` through `ProvenanceDest`. `loadRecords` has no graph; `fieldsOf(g, q, …)` does. `sqlRecords` runs on the database handle, not through the graph's query counter |
| Credibility rule | `core/database/autoreconciler/autoreconciler.go` `ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` | Exported for `conclusiondetails`; also used by `sqlLoadCandidates`. The column repeats the literals `'standard'` and `'provenencia'` |
| Grades | `core/database/sourcecredibilitygrades/registry.go` `seedGrades`, `Install`; `core/database/effects/registry.go` `"source_credibility_grades": {None: true}` | Seeded at create (`core/onboarding/ready.go`); no runtime write |
| Titles are required | `core/database/sources/sources.go` `Create`, `Update` (refuse an empty trimmed title); migration `core/database/migrations/000011.sql` (`title TEXT NOT NULL`); `core/database/evidencegraphs/evidencegraphs.go` (refuses an empty graph name); `core/database/migrations/000055.sql` (`name TEXT NOT NULL`) | A Source title and a graph name are never empty; a graph's `ref` may be NULL |
| Registry | `core/database/graphcache/cache.go` `RegisterWhole`, `WholeEntry`, `Whole.Get`, `Whole.Set` | A `LifeGraph` whole entry is neither built nor stored on an `Over` graph or a nil graph (`Get` returns `ok == false`); Apply runs only when held; `Set` stores and runs `dropReaders` |
| Precedent entry | `core/database/namealiases/namealiases.go` `Index` | A whole entry registered in the table's own package `init()`, with Build and a `reflect.DeepEqual` Check |
| Doc test | `core/catalogsession/cache_doc_test.go` `TestCacheDocListsEveryEntry` | Blank-imports every owner package; holds `docs/catalog-caches.md` to `graphcache.EntryNames()` |
| Size benchmark | `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` | One `measure` per cache on the large fixture |
| Handler | `api/ffi/handlers/person_detail.go` `ListPersonClaims`, `recordLinksProto` | Maps `ClaimGroup`'s three Source fields; maps `SourceId`, `SourceTitle`, `SourceTypeId` per record links graph |
| Wire | `api/proto/engine.proto` `RecordLinksGraph` | Fields 1–6: `graph_id`, `graph_ref`, `graph_name`, `source_id`, `source_title`, `source_type_id` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogRecordLinksGraph`; `macos/App/Platform/GoStore.swift` `mapPersonRecordLinks` | `sourceTypeID` unread anywhere in `macos/App/Features` |
| Fallbacks | `macos/App/Features/Conclusions/PersonSourcesContent.swift` `graphSourceRefs`, row `title`, line `graphTitle`/`graphRef`; `PersonSourcesSection.swift` passes `graphSourceRefs`; `PersonRecordLinksContent.swift` `init(links:graphLocations:graphSourceRefs:)` | Record links' Source label is `sourceTitle.trimmedNonEmpty ?? graphSourceRefs[graphID] ?? ""`; Sources' row title is `title.trimmedNonEmpty.map { .name($0) } ?? .ref(sourceRef)`; both write a graph's title as `name ?? ref ?? ""` |
| Key follows | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `personGraphsSection`, `.personClaims`, `.personRecordLinks` | Follow `.touchedHandles` (own id), `.evidenceGraphs` (graphs in value), property labels; no `.source`. `.conclusionDetail` follows `.source` (S16-09 precedent) |
| Go tests | `core/database/graphfill/load_test.go` `TestRetitleDropsEvidenceGraph`, `TestOnlyATitleChangeDropsEvidenceGraph`, `TestSourceTypeChangeDropsEvidenceGraph`, `TestTwoEvidenceGraphsLoadApart` | The first three assert a title or type edit drops the graph (and also: unknown id not found; graph name and ref read back; no ref reads ""); the fourth reads `eg.SourceTitle` |
| Go tests | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve` | "title edit" names `def` and `jones` |
| Go tests | `core/database/persondetail/claims_test.go`, `recordlinks_test.go` | Read `g.SourceRef`, `g.Title`, `g.SourceTypeID`, `rl.Graphs[i].SourceTitle` |
| Go tests | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityQueryCountIsConstant`, `TestForEntityEvidence`; `export_test.go` | The count test compares a small and a big Person after `g.Drop()` and `g.Vocabulary()` |
| Go tests | `api/ffi/handlers/person_detail_test.go` `TestListPersonClaims`, `TestRecordLinksProto` (builds `PersonRecordLinks{Graphs: []graphcache.EvidenceGraph{…SourceTypeID…}}`), `TestGetPersonRecordLinks` | Read `GetSourceTypeId()` |
| Swift tests | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs`, `deletedEntitiesLoseTheirHiddenKeys`, `personRecordLinksFollowEveryGraphUntilLoaded`; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `invalidationsComeFromRegistrySpecs`, `personRecordLinksRevalidate`, `personRecordLinksLoadThroughTheStore`; `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift` `aGraphWithNoRefIsLabelledWithItsName` and the fixture helper; `macos/ProvenenciaTests/PersonSourcesContentTests.swift` `rowsKeepTheReadsOrder`, `cellsAndFallbacks`, `graphLocationsClaimGraphsAndSourceRefsCoverFoldedLines` | Pin today's refreshes, build `CatalogRecordLinksGraph(…, sourceTypeID:)`, and use fixtures with an empty Source title or graph name |

**Precedent.** [#789](https://github.com/mendahu/provenencia/pull/789) (S16-10) registered `namealiases.Index` as a whole entry in its own package: copy its `init()` shape, its `Check` by rebuild and `reflect.DeepEqual`, and its `docs/catalog-caches.md` row; do not copy its `LifeCatalog` lifetime (D-2). [#780](https://github.com/mendahu/provenencia/pull/780) (S16-09) added `.applyHiddenPolicy(followsValue: .source) { … }` to `.conclusionDetail`: copy it for the two keys, through `personGraphsSection` (D-10). [#793](https://github.com/mendahu/provenencia/pull/793) (S16-11) narrowed `sqlRecords` the same way: the cache supplies one half, SQL the rest.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The entry lives in `core/database/sources`, file `records.go` *(new)*: `var Records graphcache.Whole[*RecordIndex]`, registered in `init()` as `"sources.records"`. `sources` already depends on `graphcache` (no cycle). Its tests are in package `sources_test` (`records_test.go`, `records_bench_test.go`), which may import `evidencelayertest` (that fixture imports `sources`, so an internal test cannot), and each calls `sourcecredibilitygrades.Install(c)` before its first `RecordsOf` (`database.Create` seeds no grades) | The table's package owns its cache, as `namealiases.Index` | A new `sourcerecords` package — a second home for one table |
| D-2 | Lifetime `graphcache.LifeGraph`. `RecordsOf(g *graphcache.Graph) (*RecordIndex, error)` *(new)*: nil graph → `database.ErrClosed`; `Whole.Get` `ok == false` on a non-nil graph (a write's graph) → `ErrRecordsOnWrite` *(new)*. No reader runs on a write's graph (`persondetail` and `conclusiondetails` read `c.Graph()`) | Rule "A write's one-shot graph fills only what it reads": a whole Build of every Source must not run per write | A SQL fallback on `Over` — a second read path; `LifeCatalog` — builds once per write graph |
| D-3 | `Record` holds `ID []byte`, `Ref string`, `Title string`, `SourceTypeID []byte` (the name `sources.Source` uses), `CredibilityKey string` ("" = no assessment), `Credibility int` (relative to the default grade) and `ArtifactLabels map[string]string` (Artifact id → label as stored), always non-nil. No cover, no Artifact file, no date | Only fields a reader in this PR reads (user, 2026-11-12: cover is added with its first reader) | Cover and file now — fields with no reader |
| D-4 | `sourcecredibilitygrades.DefaultKey = "standard"` *(new)*, used by `seedGrades`, and `sourcecredibilitygrades.RelativeOrderColumn` *(new)*: `COALESCE(sg.sort_order - (SELECT sort_order FROM source_credibility_grades WHERE key = '<DefaultKey>' AND origin = '<vocabkey.OriginProvenencia>'), 0)`, built from the two constants, binding alias `sg`. `autoreconciler`'s column uses it; `ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest` become unexported (`provenanceColumns`, `provenanceJoins`, `provenanceDest`) once their last outside caller goes (STEP-4). The vocabulary does not gain grades | The rule "grade relative to the default" lives with the grades, one copy of the seed key; `sources` imports the light grades package, not the reconciler; grades change only at `Install` (`None: true`), so a value computed at Build cannot go stale | The column in `autoreconciler` — `sources` would import the reconciler and box out a later reconciler read of Source records (cycle); grades in `graphcache.Vocabulary` with a Go copy of the rule |
| D-5 | `RecordIndex` holds `bySource map[string]Record` and `sourceOfArtifact map[string]string` (Artifact id → Source id). Methods `Source(id []byte) (Record, bool)` and `Artifact(id []byte) (Record, string, bool)` (the Artifact's Source and label). No method iterates, so no order is defined | The detail shells name a Citation's Artifact, not its Source | A slice per Source scanned per Artifact |
| D-6 | Loader `loadRecords(q graphcache.Querier, ids [][]byte) (map[string]Record, error)` *(new, unexported)*, two queries: (1) `SELECT s.id, s.ref, s.title, s.source_type_id, COALESCE(sg.key, ''), ` + `RelativeOrderColumn` + ` FROM sources s LEFT JOIN source_credibility_assessments sca ON sca.source_id = s.id LEFT JOIN source_credibility_grades sg ON sg.id = sca.credibility_grade_id`; (2) `SELECT id, source_id, label FROM artifacts`. With `ids` nil they read every row (Build); with ids they add `WHERE s.id IN (` / `WHERE source_id IN (` through `sqlbatch.Query` (Apply). `s.title` and `label` need no COALESCE (both NOT NULL) | The design's two queries; `sqlbatch` bounds the IN list | One join — repeats Source columns per Artifact |
| D-7 | Apply returns at once when `set.SourceScopes` is empty. Otherwise it patches the held index in place: for each named id, delete the old record and its Artifacts from `sourceOfArtifact`, insert what `loadRecords` returns (none for a deleted Source), then `Records.Set(g, idx)` with the same pointer, so `dropReaders` empties any entry derived from it | O(Sources named) per commit; the registry's one write path keeps `DerivedFrom` correct for S16-13, S16-15 and S16-30 | Skipping `Set` — bypasses `catalog-queries.mdc`'s "reads and writes only through its handle"; a copy per commit — O(all Sources) |
| D-8 | `EvidenceGraph` loses `SourceRef`, `SourceTitle`, `SourceTypeID`; keeps `SourceID`. `graphHead` keeps `ref`, `name`, `sourceID`; the head query reads `evidence_graphs` alone | Rule "A Source's words live once"; the compiler finds every reader | Keeping them unread — two copies again |
| D-9 | The `"sources"` effect drops its `EvidenceGraph:` field. `Search` keeps `{Kind: DocEvidenceGraph, Path: onField("title", graphsOfSource(self()))}`. `TestSourceEvidenceIsNotActivity`'s row becomes `{"sources", false, true}` and `TestEvidenceGraphPathsEndAtGraphs` drops `"sources"` from `want`; both comments say a Source row names no graph | No reader takes Source words from a loaded graph after STEP-5; the Swift canvas (`.evidenceGraph` key) reads no Source words (its five reads return none) | Keeping it — the rename-evicts-graphs cost this PR removes |
| D-10 | `CatalogQueryDescriptor.personGraphsSection` gains `sourceIDs: @escaping @Sendable (Value) -> Set<String>` and adds `.applyHiddenPolicy(followsValue: .source, ids: sourceIDs)` after the `.evidenceGraphs` entry. Claims pass `Set(groups.map(\.sourceID))`; record links pass `Set(links.graphs.map(\.sourceID))`. Refreshes become `[.touchedHandles, .evidenceGraphs, .source, .list(.propertyVocabulary)]` for both | One place for both keys | A `.source` entry written in each descriptor |
| D-11 | `RecordLinksGraph` removes `source_type_id` with `reserved 6; reserved "source_type_id";` and keeps `source_id` (read by D-10's follow) and `source_title`. No `source_ref`. Swift `CatalogRecordLinksGraph` loses `sourceTypeID` | The `.source` follow needs the Source id; the title is never empty, so no ref is needed (user, 2026-11-12) | Adding `source_ref` and dropping `source_id`, as the map said |
| D-12 | `ClaimGroup` becomes `{Source sources.Record; Claims []Claim}`; `PersonRecordLinks.Graphs` becomes `[]SourcedGraph` *(new)* `{Graph graphcache.EvidenceGraph; Source sources.Record}`; `linksBetween` keeps `[]graphcache.EvidenceGraph`, mapped from them. Groups sort by `Source.Title`, then `Source.Ref` | One Source value per answer instead of three copied fields | Three fields filled from the record |
| D-13 | A loaded graph whose `SourceID` has no record is an error: `fmt.Errorf("persondetail: graph %x names Source %x, which has no record", …)`; a detail record whose Citation's Artifact has none: `fmt.Errorf("conclusiondetails: Artifact %x has no Source record", …)` | `sources.Delete` refuses a Source with Artifacts or Subjects (`deleteimpact`), its graphs cascade (`evidence_graphs.source_id`), an Artifact delete names its Source, and Apply runs before the next read: a miss is a cache bug (precedent `personEventProto`) | Skipping the group — hides a cache bug |
| D-14 | Swift shows a Source's title and a graph's name as stored: Sources row `title: .name(group.title)`; its spoken label always includes the ref; a line's `graphTitle = claim.graphName`, `graphRef = claim.graphRef.trimmedNonEmpty`; record links help uses `graph.graphName` and `graph.sourceTitle`; the node label stays `graph.graphRef.trimmedNonEmpty ?? graph.graphName` | The catalog refuses an empty title or name (Current state "Titles are required"); fallbacks for impossible states go (user, 2026-11-12) | Writing the fallbacks once in a shared helper |
| D-15 | `fieldsOf` calls `sources.RecordsOf(g)` once and passes the `*sources.RecordIndex` to `loadRecords`, which gains a `recs` parameter. `sqlRecords` selects `c.artifact_id` and `c.transcription_uncertain` in place of the dropped columns | One lookup per read | Calling `RecordsOf` per row |
| D-16 | `TestForEntityQueryCountIsConstant` warms the records (`sources.RecordsOf(g)`) after `g.Vocabulary()`, before counting | Build runs on first read; the test compares two warm reads | Counting Build into the first read |

**Rules this plan bends.**

| Rule | Decision | Why | User's answer | Step |
| --- | --- | --- | --- | --- |
| Working rule "(Batch 2) A Source's words live once" names cover among the record's fields | D-3 | No reader in this PR reads cover | 2026-11-12: leave cover out; the rule says cover is added with its first reader | STEP-8 |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/sources/records.go` *(new)* `Records`, `Record`, `RecordIndex`, `RecordsOf`, `ErrRecordsOnWrite`, `loadRecords` | whole registry entry, types, funcs | `core/database/namealiases/namealiases.go` `Index`; `sources.Source` | Searched `RegisterWhole`, `sources.Source`, `GetBySource`: `Source` is the edit row read one at a time; no cache holds Sources |
| `sourcecredibilitygrades.DefaultKey`, `RelativeOrderColumn` *(new)* | consts | `autoreconciler.ProvenanceColumns` (holds the rule today), `seedGrades` (holds the key) | Moved, not added beside: the reconciler's column is built from it |
| `persondetail.SourcedGraph` *(new)* | type | `graphcache.EvidenceGraph` | The graph no longer carries Source words (D-8) |
| `conclusiondetails.SQLRecordsForTest` *(new)* | test export | `core/database/conclusiondetails/export_test.go` | Existing file, one more line |
| `core/database/sources/records_test.go` *(new)*, `records_bench_test.go` *(new)* | tests | `core/database/namealiases` tests | New entry |

**Builds on.** The S16-10 registry (`RegisterWhole`, Apply on `effects.Set`, `Whole.Set`, Check under `PROVENENCIA_GRAPH_VERIFY`); `effects.Set.SourceScopes`; `sqlbatch.Query`; S16-09's `.source` follow.

**Deletes.**
- `EvidenceGraph.SourceRef`, `SourceTitle`, `SourceTypeID` and their doc comment; `graphHead.sourceRef`, `sourceTitle`, `sourceTypeID`; the `JOIN sources s` in `loadGraphHead`.
- `ClaimGroup.SourceID`, `SourceRef`, `Title`, `SourceTypeID` (replaced by `Source`).
- In `sqlRecords`: `JOIN artifacts a`, `JOIN sources src`, `autoreconciler.ProvenanceJoins`, and the columns `a.id, src.id, COALESCE(src.title, ''), COALESCE(a.label, ''), COALESCE(sg.key, '')` and `ProvenanceColumns`.
- The exported names `autoreconciler.ProvenanceColumns`, `ProvenanceJoins`, `ProvenanceDest`; the `'standard'` and `'provenencia'` literals in the reconciler's column.
- The `"sources"` effect's `EvidenceGraph:` line.
- `RecordLinksGraph.source_type_id` (field 6, reserved); `CatalogRecordLinksGraph.sourceTypeID` and its lines in `GoStore.mapPersonRecordLinks`, `PersonSourcesPreviewData.swift` and test fixtures.
- `PersonSourcesContent.graphSourceRefs` (property and builder), the `graphSourceRefs:` parameter of `PersonRecordLinksContent.init` and its argument in `PersonSourcesSection.swift`; the `trimmedNonEmpty` fallbacks for a Source title and a graph name in both content builders.
- `PersonSourcesPreviewData.untitled` and `#Preview("Untitled Source")`.
- The "edit drops the graph" assertions in the three `graphfill` tests (rewritten, T-5).

**Next adds.** S15-05 adds the date to `Record` and to `loadRecords`' first query. S16-13 composes claims from the records tier plus `RecordsOf`, and may register an entry `DerivedFrom: sources.Records` (D-7 empties it). S16-30 composes a provenance record's Source half from `RecordIndex.Artifact` and deletes the rest of `sqlRecords`. S15-24's move needs no change here: an `artifacts` update of `source_id` names old and new Source (`readChangeColumn`). S16-25 replaces D-10's `.source` entry with `.followsInputs`.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| Claims and record links keys revalidate on a write naming a cited Source | `CatalogQueryRegistry.swift` `personGraphsSection` | Same `applyHiddenPolicy(followsValue:)` path `.conclusionDetail` uses; a hidden key marks stale; before the value arrives `followsValue` matches any id |
| A graph or Artifact with no record | `persondetail.claims`, `conclusiondetails.loadRecords` | Error (D-13); the section shows S16-04's retry notice |
| `RecordsOf` on a write's graph or a nil graph | callers on `graphcache.Over` | None today; `ErrRecordsOnWrite` / `database.ErrClosed` (AC-2.2) |

### Scale

At the map's Batch 2 Scale (3,000 Sources, 10,000 Evidence graphs, 50,000 Observations; large fixture at 4,000 and 16,000 persons):

- **Build** (first read of a session, or after `Drop`): two full scans, 3,000 Source rows and every Artifact row (est. 30,000). Est. 5–15 ms, about 6 MB (measured 5.76 MB in round 0: two string-keyed maps of 30,000 Artifacts). Measured by `BenchmarkSourceRecords` (T-13); Stop over 50 ms or 8 MB (user, 2026-11-12, raised from 5 MB).
- **Apply per commit.** `SourceScopes` is the audit activity scope, so nearly every write inside a Source names it: Source, Artifact and credibility edits, and also every Observation, Subject, Citation, note and owned-value edit in the Source's graphs (user, 2026-11-12: keep this signal). While the records are held, each such commit runs two queries for its Source (usually one) and reads that Source's whole Artifact list. Real: a few hundred Artifacts at most. Crafted: a Source with 10,000 Artifacts reloads 10,000 rows on every Observation edit inside it; T-13 measures that Apply, budget 20 ms. A write naming every Source does not exist (a Source type delete is refused while in use).
- **Evidence graphs:** a rename or type change reloads zero graphs (today: every loaded graph of the Source, up to 64, each a full load on its next read).
- **Swift:** the claims and record links keys now also revalidate on writes in a cited Source that touch no claimed graph (another graph's Observation, an Artifact label, a credibility change, a note). While the Sources section is visible that is one `ListPersonClaims` and one `GetPersonRecordLinks` per such write; hidden keys mark stale. Both reads are within S13-27's 50-graph budget; S16-25 narrows the follow to the value's inputs.
- **Memory:** `docs/catalog-caches.md` gets the entry's size at 4,000 and 16,000 persons (`BenchmarkCacheSizes`) and at 3,000 Sources (T-13); est. under 8 MB against the 250 MB ceiling.

### Edges

- **Assumptions.** `SourceScopes` names every Source whose record a write changes: `sources` (self, create/update/delete), `artifacts` (`source_id`, read from the old row on delete), `source_credibility_assessments` (`source_id`). An audit-scope edit that dropped one of these would leave the cache stale; STEP-8 writes that dependency into `add-audit-scope` and the effects README, and T-2 holds each. Grade rows change only at `Install` before any session read (D-4); a runtime grade write is a Stop.
- **Neighbors.** A description, cover, note, metadata or date edit reloads the record with the same words and evicts no graph. An Observation edit in a graph of the Source reloads the record (unchanged) and still drops that one graph (T-5's control). A type change keeps graphs loaded and changes the claims group's type (AC-1.5).
- **Removed inputs.** A deleted Source (only one with no Artifact and no Subject can be deleted; `sources.Delete` refuses the rest): Apply removes its record; its graphs are deleted by cascade and name themselves in `Set.EvidenceGraphs` (`deleteimpact` releases them), so no loaded graph points at a missing record. A deleted Artifact: its Source is in `SourceScopes`; the reload omits it. T-2.
- **Real data.** The credibility column runs over the shipped grades (`low_trust` 1, `standard` 2, `high_trust` 3): relative −1, 0, +1, and 0 with no assessment, as the reconciler weighs them today (T-1).
- **Untrusted size.** Build is linear in Sources and Artifacts; Artifacts per Source multiplies Apply (crafted 10,000, budget above); `SourceScopes` length is chunked by `sqlbatch`. No loop over Sources inside a loop over Artifacts.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcecredibilitygrades` | `const DefaultKey = "standard"` *(new)* used by `seedGrades`; `const RelativeOrderColumn` *(new)* (D-4) |
| `core/database/sources` | `records.go` *(new)*: `Record`, `RecordIndex` with `Source` and `Artifact`, `Records` registered in `init()` (Name `"sources.records"`, `Lifetime: graphcache.LifeGraph`, Build = `loadRecords(graphcache.Querier(g), nil)`, Apply per D-7, Check = `loadRecords(…, nil)` and `reflect.DeepEqual` with the held `bySource` and `sourceOfArtifact`), `RecordsOf`, `ErrRecordsOnWrite = errors.New("sources: records are not read on a write's graph")` |
| `core/database/autoreconciler` | `provenanceColumns` built from `sourcecredibilitygrades.RelativeOrderColumn` + `c.transcription_uncertain`; the trio unexported in STEP-4 |
| `core/database/persondetail` | `claims`: `recs, err := sources.RecordsOf(c.Graph())`; group by `m.graph.SourceID`, `Source: rec` (D-12, D-13). `recordLinks`: `Graphs []SourcedGraph` from the groups |
| `core/database/conclusiondetails` | D-15; the scan sets `SourceID`, `SourceTitle`, `ArtifactLabel`, `CredibilityKey`, `Provenance.Credibility` from `recs.Artifact(artifactID)` and `Provenance.Uncertain` from the column; D-13 |
| `core/database/graphfill`, `core/database/graphcache` | D-8 |
| `core/database/effects` | D-9; README (STEP-5) |

Errors: `sources.ErrRecordsOnWrite`; D-13's two invariant errors. No new `apperr` code.

### FFI

`api/proto/engine.proto` `RecordLinksGraph` per D-11; its comment says "one Evidence graph of the person and its Source". Handler `api/ffi/handlers/person_detail.go`: `ListPersonClaims` maps `g.Source.ID`, `.Ref`, `.Title`, `.SourceTypeID` into today's `PersonClaimGroup` fields; `recordLinksProto` maps `GraphId`, `GraphRef`, `GraphName` from `g.Graph` and `SourceId`, `SourceTitle` from `g.Source`. Regenerate with `./scripts/generate-proto.sh` (needs `protoc`, `protoc-gen-go`, `protoc-gen-swift`).

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Record links graph model | `CatalogRecordLinksGraph` | `macos/App/Platform/GenealogyStore.swift` | `sourceTypeID` removed; `GoStore.mapPersonRecordLinks`, `PersonSourcesPreviewData.swift` and every test constructor drop `sourceTypeID:` |
| Claims and record links follows | `CatalogQueryDescriptor.personGraphsSection` | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` | D-10; the `.personRecordLinks` comment's "a Source write names its graph" becomes "a Source write names the Source" |
| Sources rows and lines | `PersonSourcesContent` | `macos/App/Features/Conclusions/PersonSourcesContent.swift` | D-14; `graphSourceRefs` deleted |
| Record links labels | `PersonRecordLinksContent` | `macos/App/Features/Conclusions/PersonRecordLinksContent.swift` | `init(links:graphLocations:)`; D-14 |

States: unchanged. Accessibility: the same strings; a Source row's spoken label always ends with its ref (it already did for every titled Source).

### Strings

None — no key added or removed; `recordLinksNode*` and `sourcesOpenGraph` receive the same values as today for titled Sources and named graphs.

## 7. Steps

**STEP-1 — The credibility rule moves to the grades**
Files: `core/database/sourcecredibilitygrades/registry.go`, `core/database/autoreconciler/autoreconciler.go`. Skill: `add-catalog-query`.
Do: Before any change, run `BenchmarkCacheSizes` on `main` and put the figures in your report to the orchestrator. Add `DefaultKey` and `RelativeOrderColumn` (D-4); `seedGrades` uses `DefaultKey`; the reconciler's column is built from `RelativeOrderColumn` (names still exported).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcecredibilitygrades/... ./core/database/autoreconciler/... ./core/database/conclusiondetails/...` passes unchanged.

**STEP-2 — The Source records entry**
Files: `core/database/sources/records.go` *(new)*, `core/database/sources/records_test.go` *(new)*, `core/database/sources/records_bench_test.go` *(new)*, `core/catalogsession/cache_doc_test.go` (blank import `core/database/sources`), `core/database/promotealign/bench_test.go` (`measure("sourceRecords", func() { _, err := sources.RecordsOf(g); must(b, err) })` after `"nodes"`), `docs/catalog-caches.md`. Skill: `add-catalog-query`.
Do: D-1 to D-7. Add the `sources.records` row to "Registry entries": owner `core/database/sources/records.go` *(new)* · `Records`; key one value; Fill Build, two queries, on first read; Updated by Apply reloading the Sources in `effects.Set.SourceScopes`, then `Set`; Lifetime graph; Cap none; sizes from the benchmarks.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sources/... ./core/catalogsession/...` passes (T-1, T-2, T-3, `TestCacheDocListsEveryEntry`); `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'SourceRecords' ./core/database/sources/` reports within budget.

**STEP-3 — Claims and record links read the record**
Files: `core/database/persondetail/claims.go`, `recordlinks.go`, `claims_test.go`, `recordlinks_test.go`, `api/ffi/handlers/person_detail.go`, `api/ffi/handlers/person_detail_test.go` (Go types only; the wire is STEP-6). Skill: `use-catalog-session`.
Do: D-12, D-13. The handler fills today's proto fields from `g.Source` (`SourceTypeId` from `g.Source.SourceTypeID` until STEP-6). `TestRecordLinksProto` builds `[]persondetail.SourcedGraph`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/persondetail/... ./api/ffi/handlers/...` passes (T-4).

**STEP-4 — The record shells take Source words from the record**
Files: `core/database/conclusiondetails/conclusiondetails.go`, `conclusiondetails_test.go`, `export_test.go`, `core/database/autoreconciler/autoreconciler.go`.
Do: D-15, D-13, D-16; add `SQLRecordsForTest` to `export_test.go`. Then unexport the trio in `autoreconciler` (D-4).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusiondetails/... ./core/database/autoreconciler/...` passes (T-6, T-7, `TestForEntityEvidence` unchanged); `rg -n "ProvenanceColumns|ProvenanceJoins|ProvenanceDest" --type go` has no hit.

**STEP-5 — Graphs stop copying Source words, and a rename names no graph**
Files: `core/database/graphcache/evidence_graph.go`, `core/database/graphfill/load.go`, `core/database/graphfill/load_test.go`, `core/database/effects/registry.go`, `core/database/effects/evidence_graphs_test.go`, `core/database/effects/registry_completeness_test.go`, `core/database/effects/README.md`.
Do: D-8 and D-9 in one commit. Rewrite the `graphfill` tests per T-5. In `TestEvidenceGraphsResolve` the "title edit" set names no graph and its `SourceScopes` holds the Source; reword its doc comment. README: replace the whole sentence from "A Source row names every graph of the Source" through the end of that sentence with "A Source row names no graph: no loaded graph holds Source words, which live in the Source records entry (`sources.Records`), updated from `SourceScopes`; a title edit still reprojects the Source's graphs' search documents (`DocEvidenceGraph`)."
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/database/graphfill/... ./core/database/graphcache/... ./core/database/persondetail/... ./api/ffi/handlers/...` passes (T-5, T-8); `rg -n "SourceRef|SourceTitle|SourceTypeID" core/database/graphcache/evidence_graph.go core/database/graphfill/load.go` has no hit, and `rg -n "eg\.Source(Title|Ref|TypeID)" core/database/graphfill` has no hit.

**STEP-6 — The record links wire drops its unread field**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/person_detail.go`, `api/ffi/handlers/person_detail_test.go`. Skill: `add-ffi-handler`.
Do: D-11; `./scripts/generate-proto.sh`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes (T-9).

**STEP-7 — The app follows Sources and shows titles as stored**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Features/Conclusions/PersonSourcesPreviewData.swift`, `PersonSourcesContent.swift`, `PersonSourcesSection.swift`, `PersonRecordLinksContent.swift`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, and the four Swift test files in Current state. Skills: `add-workspace-place`, `add-swift-test`.
Do: D-10, D-11, D-14, and T-10 – T-12. Delete `PersonSourcesPreviewData.untitled` and `#Preview("Untitled Source")` in `PersonSourcesSection.swift` (a state the catalog refuses). Reword the doc comment of `personGraphsSection` ("its own handle, the graphs its value holds, or the Sources they are on") and the `.personClaims` case comment the same way.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/PersonRecordLinksContentTests -only-testing:ProvenenciaTests/PersonSourcesContentTests` passes, one run at a time.

**STEP-8 — Docs**
Files: `docs/catalog-caches.md` (sizes), `core/database/effects/README.md` (one sentence by `SourceScopes`: the Source records entry reads it, so `sources`, `artifacts` and `source_credibility_assessments` keep their `SourceScope` paths), `.agents/skills/add-audit-scope/SKILL.md` (the same line where it decides `SourceScope` paths), `docs/macos-client-patterns.md` ("Conclusion keys": claims and record links follow the Sources in their value through `.source`), `docs/deployment-plan/spike-16/cache-architecture.md` (Entity records: no cover or file yet, credibility from the grades' column, the record links wire keeps `source_id` and gains no ref), `docs/deployment-plan/spike-16/deployment-plan.md` (S16-12 row and scope entry per D-3, D-4, D-11; the working rule "A Source's words live once": cover is added with its first reader; status), `docs/deployment-plan/spike-16/person-page-audit.md` (P-22 part: the ref was found unneeded), `docs/ideas/structural-debt.md` (one line: the provenance sheet's empty-Source-title fallback guards a state the catalog refuses), `completed.md`, this plan.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-16/prs/S16-12-source-records.md` passes; `python3 scripts/check-guidance-names.py` passes.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/sources/records_test.go` `TestRecordsBuild` *(new)* | Table: Source "Register" (low_trust; Artifacts "Page 1" and "Page 2"), Source "Census" (no assessment, no Artifact), Source "Gazette" (high_trust) → `Source(id)` gives ref, title, type id, key ("low_trust" / "" / "high_trust"), `Credibility` (−1 / 0 / +1), `ArtifactLabels` non-nil (empty for Census); `Artifact(page1)` → Register, "Page 1"; unknown id → false | AC-1.3, D-3, D-4, D-5, D-6 |
| T-2 | Go | `records_test.go` `TestRecordsApply` *(new)* | Records held, plus a test-only entry registered as a package-level `var` in `sources_test` with `graphcache.RegisterWhole`, `DerivedFrom: []graphcache.Source{sources.Records}`, default (`LifeGraph`) lifetime, holding a value. Table of writes through `writes.Run` on Source "Register": rename, type change, Artifact create, Artifact relabel, Artifact delete, credibility set (`sourcecredibility.Upsert`), an Observation edit in its graph; then delete a second Source that has only its default graph → after each, `g.Verify()` returns nil, the named fields match, a deleted Source or Artifact resolves to false, and the derived entry is empty | AC-2.1, D-7, Edges |
| T-3 | Go | `records_test.go` `TestRecordsOf` *(new)* | `RecordsOf(graphcache.Over(tx))` → `ErrRecordsOnWrite`, and `Over`'s query counter shows no query; `RecordsOf(nil)` → `database.ErrClosed` | AC-2.2, D-2 |
| T-4 | Go | `core/database/persondetail/claims_test.go` `TestClaimsAfterSourceEdit` *(new)*; `recordlinks_test.go` (updated) | Person with a claim in G1 of "Register"; rename to "Parish register", then change type → `Claims` group `Source.Title` and `Source.SourceTypeID` new; `RecordLinks` `Graphs[0].Source.Title` new. Existing order tests read `Source.Title` | AC-1.2, AC-1.5, D-12 |
| T-5 | Go | `core/database/graphfill/load_test.go` | `TestRetitleKeepsEvidenceGraph` (renamed): two loaded graphs of one Source, each with a memo; rename → both stay loaded, memos read back, `SourceID` unchanged; unknown id still not found. `TestSourceTypeChangeKeepsEvidenceGraph` (renamed): type change → graph stays loaded; graph name and ref read back, no ref reads "". `TestSourceEditKeepsEvidenceGraph` (renamed from `TestOnlyATitleChangeDropsEvidenceGraph`): description edit keeps it; control: a Subject create in G1 drops G1 only. `TestTwoEvidenceGraphsLoadApart`: checks `SourceID` in place of the title | AC-1.1, AC-1.5, D-8, D-9 |
| T-6 | Go | `core/database/conclusiondetails/conclusiondetails_test.go` `TestForEntityAfterSourceEdit` *(new)* | AC-1.3's before and after | AC-1.3 |
| T-7 | Go | `conclusiondetails_test.go` `TestForEntityQueryCountIsConstant` (updated, D-16) | After `g.Drop()`, `g.Vocabulary()` and `sources.RecordsOf(g)`: small and big Person → equal graph query counts; `SQLRecordsForTest()` contains none of `sources`, `artifacts`, `source_credibility_` | AC-2.4 |
| T-8 | Go | `core/database/effects/evidence_graphs_test.go` `TestEvidenceGraphsResolve`; `registry_completeness_test.go` `TestSourceEvidenceIsNotActivity`, `TestEvidenceGraphPathsEndAtGraphs` (updated) | Title edit → `EvidenceGraphs` empty, `SourceScopes` = [src], `Search` still holds `DocEvidenceGraph` for def and jones; the two completeness rows per D-9 | AC-2.3, D-9 |
| T-9 | Go | `api/ffi/handlers/person_detail_test.go` `TestGetPersonRecordLinks`, `TestRecordLinksProto` (updated) | Every graph: `GetSourceId()` = its Source's id, `GetSourceTitle()` = its title; `TestListPersonClaims` unchanged | AC-3.1 |
| T-10 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `personClaimsFollowHandlesAndGraphs` (updated) | Claims value cites "s1": `revalidatedKeys(sourceScopes: ["s1"])` == {claimsKey, detailKey}; `["s9"]` → []. Other `CatalogRecordLinksGraph(…)` constructors in the file drop `sourceTypeID:` | AC-1.4, D-10 |
| T-11 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `personRecordLinksRevalidate`, `invalidationsComeFromRegistrySpecs` (updated) | Record links: `sourceScopes: ["s1"]` → [key], `["s9"]` → []; both keys' refreshes `== [.touchedHandles, .evidenceGraphs, .source, .list(.propertyVocabulary)]`; `personRecordLinksLoadThroughTheStore` drops `sourceTypeID:` | AC-1.4, D-10 |
| T-12 | Swift | `macos/ProvenenciaTests/PersonRecordLinksContentTests.swift`, `PersonSourcesContentTests.swift` (updated) | `aGraphWithNoRefIsLabelledWithItsName`: Source title "Census", no ref map → help "Page 4 — Census"; the fixture helper gives each graph a name and Source title. Sources fixtures give "src-3" the title "Letter" and every claim a graph name; `rowsKeepTheReadsOrder` expects `rows[2].title == .name("Letter")`; `cellsAndFallbacks` expects `letter.accessibilityLabel == "Letter, date not recorded, SRC-4KQ7M"` and drops the no-name graph case; `graphLocationsClaimGraphsAndSourceRefsCoverFoldedLines` is renamed `graphLocationsAndClaimGraphsCoverFoldedLines` and loses its `graphSourceRefs` assertion | AC-3.2, AC-3.3, D-14 |
| T-13 | Go | `core/database/sources/records_bench_test.go` `BenchmarkSourceRecords` *(new)* | Seeded by raw `INSERT` inside one transaction (as `core/database/graphcache/graphcachetest/graphcachetest.go` seeds): 3,000 Sources × 10 Artifacts, one assessment per tenth Source, plus one Source with 10,000 Artifacts. Reports Build ms and heap MB; Apply ms through `g.OnCommit(effects.Commit{Effects: effects.Set{SourceScopes: [][]byte{id}}})` for an ordinary Source and for the 10,000-Artifact one | D-6, D-7 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/graphfill` (its check compares whole graphs), `core/database/persondetail` order tests, `core/database/conclusiondetails` count test, `PROVENENCIA_GRAPH_VERIFY` runs (the new Check), `CatalogQueryRegistryTests`.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: docs/catalog-caches.md (row and sizes), docs/macos-client-patterns.md (Conclusion keys), core/database/effects/README.md, add-audit-scope, cache-architecture.md (Entity records)
- [ ] Report: BenchmarkCacheSizes before (STEP-1) and after, BenchmarkSourceRecords after, on the same machine
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a person whose records sit in two Evidence graphs of Source "Register". Open the person on Sources: the row reads "Register"; hover a record links node: its help names "Register".
2. Open the person's Vitals and a field's provenance sheet: the record names "Register".
3. Open Source "Register" and rename it "Parish register". Go Back to the person: the Sources row, the record links help and the provenance sheet read "Parish register".
4. Change the Source's type: the Sources row shows the new type icon and label.
5. Open one of the two graphs: the canvas shows as before.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A reader of `EvidenceGraph` Source words outside the files found | Deleting the fields makes the compiler list them (STEP-5) |
| The claims and record links keys reload more often in Swift | Stated in Scale; hidden keys mark stale; S16-25 narrows |
| An audit-scope change drops a `SourceScope` path the cache needs | STEP-8 documents the dependency; T-2 fails if it happens |
| In-place Apply surprises a reader holding the index | Readers hold it within one engine call; `OnCommit` runs between calls (D-7) |

**Stop and report if:**

- A row of §4 is no longer true at your base (notably `sqlRecords`, `loadGraphHead`, the `"sources"` effect).
- Any Go code outside `persondetail`, `graphfill`, `api/ffi/handlers/person_detail.go` and their tests fails to compile after STEP-5 because it reads `EvidenceGraph.SourceRef`, `SourceTitle` or `SourceTypeID`.
- A Swift view outside the files in STEP-7 reads `CatalogRecordLinksGraph.sourceTypeID`, or a Source title or graph name can be stored empty (a write path not in "Titles are required").
- `rg -n "source_credibility_grades" --type go` shows a write other than `sourcecredibilitygrades.Upsert` from `Install` (D-4's premise).
- A caller of `persondetail.Claims`, `persondetail.RecordLinks` or `conclusiondetails.ForEntity` runs on a `graphcache.Over` graph (D-2's premise).
- `BenchmarkSourceRecords` shows Build over 50 ms or 8 MB at 3,000 Sources × 10 Artifacts, or Apply over 20 ms for the 10,000-Artifact Source.
- The Swift Evidence graph canvas, Promote or the Citation composer shows a Source title read from the `.evidenceGraph` key's value (D-9's premise).

## 12. Open questions

None.

## 13. PR

**Title:** S16-12 — Source records in the graph cache
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-12-source-records.md`.
