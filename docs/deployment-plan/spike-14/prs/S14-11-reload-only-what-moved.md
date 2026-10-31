# S14-11 — Reload only what moved, only when it is on screen

**Status:** Draft
**Spike / slice:** Spike 14 · Slice 4 — Evidence graphs finished (batch 4; moved from Spike 12, where it was S12-26, on 2026-10-29; S12-25 landed in #621, S12-34's descriptor is on `main`)
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: every view draws the same rows; only when and how much the app reloads changes)
**Board:** —
**Depends on:** — (S12-25's keyed `reads` and S12-34's `CatalogQueryDescriptor` `patches` / `refresh` slots are both on `main` @ `85a5ce2`)
**Blocks:** —
**Branch:** `s14-11-reload-only-what-moved`
**Version bump:** patch — a performance fix and the stale-count fix, no feature ([`docs/versioning.md`](../../../versioning.md)); the next patch after `VERSION` at merge (0.2.3 → 0.2.4 at planning)
**Skills to load:** [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/swift-tests.mdc`, `.cursor/rules/macos-client.mdc`
**Context:** [`app-health-audit.md`](../app-health-audit.md) A-9, A-10, A-25; [`docs/ideas/archive/settled-list-count-publish.md`](../../../ideas/archive/settled-list-count-publish.md) (its leftover); working rules "A write names only the graphs whose loaded rows it moves, and the partial reads it moves by key", "A hidden graph or Source page goes stale; it does not reload", "FakeStore is a seeded fake, not a second engine"; [`macos-client-patterns.md`](../../../macos-client-patterns.md) cache contract rules 2–4; [`source-layer-data-model.md`](../../../source-layer-data-model.md) § 7.1
**Planned:** 2026-10-31 against `main` @ `85a5ce2`
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

On a census Source with 60 household graphs, graph work reloads only what it moved and only what is on screen. Linking an Artifact, editing or deleting a graph, adding an Artifact, or adding or copying a Source re-reads that one Source's graphs instead of the whole project's graph list, and Add Source and Copy Source put the new row into the Sources list instead of reloading it. A graph canvas or Source page the researcher visited earlier and left waits until it is shown again instead of reloading on every write. A list that comes back after going stale publishes its sidebar count only after its reload. Typing in Add Source or the list search no longer re-sorts every Source on each keystroke.

**Why now.** Bar outcome 14 ("Graph work stays fast on a split census") and the Spike 12 app-health audit: A-9 (every graph or link write reloads the project's graph list, 20.9 ms at 3,060 graphs, and Add Source reloads the Sources list, 14 ms at 3,061 Sources), A-10 (hidden graph and Source-page keys reload instead of going stale), A-25 (the Sources list filters and sorts every Source per body run), and the leftover in [`settled-list-count-publish.md`](../../../ideas/archive/settled-list-count-publish.md) (a stale hidden list publishes its cached count on appear).

**What stays.** What every view draws; the graph list's rows (`CatalogEvidenceGraph`) and the session's `EvidenceGraphIndex` reads; `evidencegraphs.List`; Source delete reporting no effects (S14-16 adds them); the Sources list's sort orders and filters; the sidebar recount after a Source create or delete (`navCountTriggers`); every other key's hidden policy.

## 2. User stories

**US-1.** As a researcher on a 60-graph census, when I link an Artifact to a graph, edit or delete a graph, or add an Artifact, the app re-reads that Source's graphs and no other Source's, so graph work stays fast.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Source S has graphs Default and Jones; a link row between Jones and page 2 is created | The write's effects carry `reads: [{read: "evidence_graphs", ids: [S]}]` and no `lists` entry `evidence_graphs` |
| AC-1.2 | The same write applied in the app with the graph list on screen and warmed for Sources S and T | The store is asked for S's graphs once (`listEvidenceGraphs(sourceID: S)`); T's rows are untouched; Jones's `artifactIDs` include page 2 after `settleWrites()` |
| AC-1.3 | `ListEvidenceGraphs` with `source_id` = S | Returns S's graphs only, by name; with `source_id` empty it returns every graph, as today; with a malformed id it fails with the shared invalid-id error |
| AC-1.4 | Source S is deleted (`.deletedSource(id: S)`, no effects) with the graph list warmed | S's rows leave the cached list through the same re-read (it returns none); the list is not reloaded whole |
| AC-1.5 | The graph list is warmed but not on screen (the researcher is on the Persons list) when a link write names S | Nothing is fetched; the list key goes stale and reloads whole on the next visit to a place that shows it |
| AC-1.6 | A partial re-read is running for S | `evidenceGraphsHandle.isFetching` is true, so `evidenceGraphPresence` reads `.unknown` and `artifactIDsInNoGraph` reads `nil` until it lands |

**US-2.** As a researcher, when I add or copy a Source, the new row appears in the Sources list at once without the list reloading.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | The Sources list is warmed with 2 Sources; Add Source creates "Deed" | The list holds 3 rows, the third being the row `createSource` returned (ref, title, type, `updated`); `listSources` was not called again |
| AC-2.2 | Copy Source creates "Census 1851, page 2" | The same: the returned row is inserted; no `listSources` call |
| AC-2.3 | The returned row's id is already in the cached list (a reload landed first) | The row is replaced, not duplicated |
| AC-2.4 | The Sources list was never warmed | Nothing happens to it; the next visit loads it |

**US-3.** As a researcher, when I leave a graph canvas or a Source page and keep working elsewhere, that page waits until I come back instead of reloading on each write.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Graph G's key is warmed, the researcher is on graph H's canvas, a write names G in `evidenceGraphs` | G's rows are not fetched; G's key is stale |
| AC-3.2 | The researcher goes back to G's canvas | G's rows load once; the canvas shows them |
| AC-3.3 | Source page P is warmed and hidden; `.mutatedSourceWorkspace(sourceId: P)` is applied | P is not fetched until the Source page or a canvas or composer of P is shown |

**US-4.** As a researcher, a list that comes back on screen after going stale shows its sidebar count only after it has reloaded.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | The Sources list key is marked stale while hidden | `handle.isStale` is true and `handle.isSettled` is false, though `status == .ready` and `isFetching` is false |
| AC-4.2 | The list is shown again and its reload finishes | `isStale` is false and `isSettled` is true; `onSettled` runs its publish once, after the reload |
| AC-4.3 | A stale key gets a value through `setQueryValue` | `isStale` is false |
| AC-4.4 | A stale key gets a partial patch through `patchCached` | `isStale` stays true |

**US-5.** As a researcher typing in Add Source or the list search, the list is not filtered and sorted again on each keystroke that does not change its inputs.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-5.1 | 3 Sources, sort A–Z; `visibleSources` read, then `draft.title` changed three times and `visibleSources` read after each | The rows are built once (`visibleSourcesBuildCount == 1`) |
| AC-5.2 | Then `settings.query` changes, or the Sources handle's revision moves, or (under a status filter) the graph list or progress revision moves | The rows are built again, once per change |

## 3. Scope

| In | Covered by |
| --- | --- |
| `effects.ReadEvidenceGraphs` *(new)* keyed by Sources, reported by graph and link rows; `effects.ListEvidenceGraphs` removed (A-9) | STEP-1, AC-1.1 |
| `ListEvidenceGraphsRequest.source_id` *(new)*; the handler reads one Source when it is set (A-9) | STEP-2, AC-1.3 |
| `GenealogyStore.listEvidenceGraphs(projectDir:sourceID:)` | STEP-3 |
| `QueryHandle.isStale` and `markStale()` *(new)*; `isSettled` false while stale; the session keeps it in step with its stale set | STEP-4, AC-4.1 – AC-4.4 |
| `WorkspaceSession.refreshCached` *(new)*: re-read part of a warmed value and merge it, with `isFetching` set while it runs | STEP-4, AC-1.6 |
| The graph list's `refresh` follows `.reads(.evidenceGraphs)` (and a deleted Source until S14-16) and merges per Source; `CatalogList.evidenceGraphs` and its invalidation case go (A-9) | STEP-5, AC-1.2, AC-1.4, AC-1.5 |
| `.createdSource(CatalogSource)` and the `insertedSourceRow` patch on the Sources list (A-9) | STEP-6, AC-2.1 – AC-2.4 |
| `.evidenceGraph` and `.sourceWorkspace` take `whenHidden: .markStale` (A-10) | STEP-7, AC-3.1 – AC-3.3 |
| `SourcesModel.visibleSources` memoized against the revisions and settings it reads (A-25) | STEP-8, AC-5.1, AC-5.2 |
| Docs: `macos-client-patterns.md` rules 2 and 4 and the `QueryHandle` row, `add-workspace-place` cache table, `source-layer-data-model.md` § 7.1 | STEP-9 |

| Out | Home |
| --- | --- |
| `DeleteSource` returning `WriteEffects` (and so the `.deletedSource` special case in `CatalogQueryFollows.ids`) | S14-16 |
| Moving the Add Source form into its own view (A-25's optional half) | not planned: the memo removes the cost |
| Narrowing `.deletedSource`'s `allCached(.evidenceGraph)` to the deleted Source's graphs | not planned: hidden graph keys now only go stale, so the cost is one flag per key |
| Hidden policy for `.evidenceGraphProgress`, `.sourceTypesList`, `.metadataFieldsList` | not planned: the audit names only A-10's two keys |
| The descriptor itself | done in S12-34; this PR adds one `refresh` and one `patches` entry |
| Removing `WorkspaceSession.invalidateAll(matching:)` (no app caller) | not planned in this PR: it predates it; [`structural-debt.md`](../../../ideas/structural-debt.md) if wanted |

## 4. Current state

Verified at `85a5ce2`.

| Area | Where | Today |
| --- | --- | --- |
| Read names | `core/database/effects/registry.go` `Read`, `ReadCitations`, `ReadEvidenceGraphProgress`, `Read.KeyTable`, `Reads` | Two keyed reads, both keyed by `evidence_graphs`. `KeyTable` returns `""` for anything else |
| Graph list flag | same file, `ListEvidenceGraphs`, `Lists`; registry entries `"evidence_graphs"` and `"evidence_graph_artifacts"` (`Lists: []List{ListEvidenceGraphs}`) | Every graph or link write flags the catalog-wide list. Their `SourceScope` paths are `field("source_id")` and `up("graph_id", "source_id")` |
| Effect resolution | `core/database/effects/resolve.go` `Resolve` | Walks each `Effect.Reads` path per change, including deleted rows through `ghostsFrom`; no change needed |
| Go tests | `core/database/effects/registry_completeness_test.go` `TestEvidenceGraphsList`; `core/database/effects/lists_test.go` (rows at lines 52–54); `core/database/effects/evidence_graphs_test.go` (`set.Has(effects.ListEvidenceGraphs)` at 141, 149, 184, 288) | Pin the list flag on graph and link tables |
| Handler | `api/ffi/handlers/evidence_graphs.go` `ListEvidenceGraphs` | Always `evidencegraphs.List(c, nil)`: the whole project |
| Query | `core/database/evidencegraphs/evidencegraphs.go` `List(c, sourceID)` | Already takes a Source (seeks `evidence_graphs_source_id_idx` through `sqlListBySource`); nil means all |
| Proto | `api/proto/engine.proto` `ListEvidenceGraphsRequest` (`project_dir = 1` only); `WriteEffects.lists` and `.reads` comments name `"evidence_graphs"` as a list | — |
| Handler tests | `api/ffi/handlers/evidence_graphs_test.go` `assertGraphEffects` (requires `lists` ∋ `evidence_graphs`), `TestListEvidenceGraphs`; `api/ffi/handlers/write_effects_test.go` (line 70 lists `ListEvidenceGraphs`) | — |
| Swift vocabulary test | `api/ffi/swift_vocabulary_test.go` rows "catalog-wide lists" and "catalog reads" | Driven by `effects.Lists` and `effects.Reads`: both rows stay and follow the Go change. The map's scope entry says the row goes; it does not (`CatalogList` keeps two cases) |
| Store | `macos/App/Platform/GenealogyStore.swift` `listEvidenceGraphs(projectDir:)`; `GoStore.swift`, `FakeStore.swift` (`evidenceGraphsByProject`, `listEvidenceGraphsError`), `macos/ProvenenciaTests/OnboardingModelTests.swift` `ThrowingStore` | No Source parameter. Callers: the registry loader and `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` (8 calls) |
| Wire enums | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogList.evidenceGraphs`, `CatalogRead`, `CatalogWriteEffectsRead` (doc: "Evidence graphs for both current reads"), `CatalogMutation.createdSource` (no payload), `CatalogMutation.evidenceGraphId` | — |
| Invalidation | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for list:project:)` case `.evidenceGraphs`; descriptor arm `.evidenceGraphsList` (`invalidateOn: [.deletedSource]`, `whenHidden: .markStale`, no `refresh`) | Every graph or link write invalidates the whole list; visible → reload of every graph |
| Sources list | same file, arm `.sourcesList` (`invalidateOn: [.createdSource, .deletedSource]`, patches `.lastUpdated`, `.updatedSourceRow()`) | Add and Copy Source reload the whole list |
| Hidden policy | same file, arms `.evidenceGraph` and `.sourceWorkspace` | No `whenHidden`, so `.revalidate`: a hidden visited page reloads on each write that names it |
| Follows | `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogQueryFollows.ids(in:mutation:)` | Special case: `.reads(.evidenceGraphProgress)` also follows `mutation?.evidenceGraphId` (Promote reports no effects) — the precedent for a deleted Source |
| Patch / refresh slots | same file, `CatalogQueryPatch.updatedSourceRow()`, `CatalogQueryRefresh`, `.applyHiddenPolicy(follows:id:)` | — |
| Session | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `invalidatedKeys`, `applyHiddenPolicy(to:)` (`.markStale` inserts into `invalidatedKeys` only), `replaceValue`, `patchCached`, `setQueryValue`, `ensureQuery`, `invalidate`, `evict`, `apply(_:_:)` (runs each warmed key's `refresh` through `trackWrite`) | Stale is the session's set only; the handle cannot tell |
| Handle | `macos/App/Features/Workspace/Session/QueryHandle.swift` `isSettled` (`status == .ready && !isFetching`), `beginLoading(stale:)` | A stale hidden key is settled, so `.onSettled` publishes its cached count on appear |
| Settled readers | `macos/App/Features/Workspace/QueryHandleSettled.swift` `onSettled`; `WorkspaceSession+EvidenceGraphs.swift` `evidenceGraphPresence`, `artifactIDsInNoGraph` | Read `isSettled` |
| Graph list index | `macos/App/Features/Workspace/Session/WorkspaceSession+EvidenceGraphs.swift` `EvidenceGraphIndex` | Groups by `sourceID` in list order; consumers read per Source or by id, never the flat order |
| Writes that create a Source | `macos/App/Features/Sources/SourcesModel.swift` `create()` and `macos/App/Features/Sources/SourceCopyModel.swift` (line 143): `session.apply(effects, .createdSource)` | — |
| Sources rows | `SourcesModel.visibleSources` (read in `SourcesListContent.body`, `macos/App/Features/Sources/SourcesListView.swift` line 63) | Filters and sorts every Source on every body run; `searchKeyCache` memoizes only the search keys |
| Swift tests | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (lines 644, 865–874, 1169, 1249, 1252–1274), `SourcesModelTests.swift` (396, the hidden-list test), `WorkspaceSessionTests.swift` (1271, 1284), `EvidenceGraphModelTests.swift` `graphPresenceFollowsTheList` | Use `.createdSource` without a payload and expect `.evidenceGraphsList` invalidations |

**Precedent.** Graph progress is the shape to copy: `CatalogQueryRefresh(follows: .reads(.evidenceGraphProgress))` in `CatalogQueryRegistry.swift` re-reads one graph per id and merges it into the warmed map, and `CatalogQueryFollows.ids` adds the mutation's graph while Promote reports no effects (S12-25, #621). The Go side copies `ReadEvidenceGraphProgress`: one `Read` constant, its `KeyTable`, an entry in `Reads`, and `ReadPath`s on the tables that move it. `CitationComposerContext.refreshMemo` (`MemoToken` of handle revisions) is the precedent for the `visibleSources` memo. Do not copy the progress refresh's silent catch: a failed partial re-read here reloads the list (D-7).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | The read is `ReadEvidenceGraphs Read = "evidence_graphs"`, `KeyTable() == "sources"`, appended to `Reads` after `ReadEvidenceGraphProgress`. Swift: `CatalogRead.evidenceGraphs = "evidence_graphs"` | A read is named for what it reads, keyed by its key table, as `citations` and `evidence_graph_progress` are; the list key of the same name goes in this PR, so the wire string has one meaning | `"source_graphs"` — a second name for the one graph list read |
| D-2 | `"evidence_graphs"` gets `Reads: []ReadPath{{ReadEvidenceGraphs, field("source_id")}}`; `"evidence_graph_artifacts"` gets `Reads: []ReadPath{{ReadEvidenceGraphs, up("graph_id", "source_id")}}`. Both drop `Lists`. `ListEvidenceGraphs` and its `Lists` entry are deleted. No entry on `"sources"` | Graph and link rows are what the list holds. A Source create or Copy writes a Default graph row, so they report the read through `evidence_graphs`; an Artifact create or delete writes or releases link rows. `CatalogEvidenceGraph` carries no Source field, so a Source edit moves nothing in the list. The paths are the entries' own `SourceScope` paths, which already resolve deleted rows through ghosts | A path on `"sources"` — reports the read on every title edit for nothing |
| D-3 | `ListEvidenceGraphsRequest` gains `string source_id = 2;`. Empty means every graph (today's behavior); otherwise `parseID` and `evidencegraphs.List(c, id)`. A Source with no graphs (deleted) returns an empty list, not an error | One Source per call matches `List`'s signature and the progress refresh's one-call-per-id loop; a write names one Source in practice | `repeated string source_ids` — no write names several Sources, and `List` takes one |
| D-4 | The store method is one requirement, `listEvidenceGraphs(projectDir: String, sourceID: String?)`; the loader passes `nil`. No convenience overload. `FakeStore` filters its seeded `evidenceGraphsByProject` rows by `sourceID` (a key filter, not an engine rule) and records each call's `sourceID` in `listEvidenceGraphsSourceIDs: [String?]` *(new)* | One entry point; the recorded ids are how tests prove only S was asked, as `citationCountsGraphIDs` does (S14-10) | A second method `listEvidenceGraphs(projectDir:sourceID:)` beside the old one — two paths to one read |
| D-5 | `QueryHandle` gains `private(set) var isStale = false`, `func markStale()` and `func clearStale()`. `beginLoading` and `resetForInvalidation` clear it. `isSettled` is `status == .ready && !isFetching && !isStale`. In `WorkspaceSession`, every insert into `invalidatedKeys` goes through a private `markStale(_ key:)` *(new)* that also calls the handle's `markStale()`, and every removal through a private `clearStale(_ key:)` *(new)* that also calls `clearStale()` (`applyHiddenPolicy` `.markStale`, `invalidate`, `replaceValue`, `ensureQuery`, `evict`) | The working rule names `QueryHandle.isStale`; one helper pair keeps the handle flag and the session's set from disagreeing. `setQueryValue` makes a key current, so it clears both; `patchCached` does not (rule 4) | Deriving `isStale` from the session (`session.isStale(key)`) — `onSettled` takes only the handle; dropping `invalidatedKeys` — `readyValue` and `ensureQuery` read it for keys whose handle is absent |
| D-6 | `WorkspaceSession.refreshCached<Value, Part>(_ key:, as:, read:, merge:) async` *(new)*: returns at once if the key has no value; if the key is not in `visibleKeys` it calls `applyHiddenPolicy(to:)` and returns; otherwise it calls the handle's `beginLoading(stale: true)`, awaits `read()`, and on success applies `merge` through `patchCached` (whose `replaceValue` clears `isFetching`) | A partial re-read must look like a refetch to `evidenceGraphPresence` and `artifactIDsInNoGraph`, or a just-created graph reads `.gone` and a just-linked Artifact reads "in no graph" for a frame. A hidden list gains nothing from a merge: it goes stale and reloads once when shown (the working rule's "only when it is on screen") | Merging into a hidden list — work for a screen nobody sees; leaving `isFetching` false — breaks `graphPresenceFollowsTheList` (e) for partial reads |
| D-7 | On a failed `read()`, `refreshCached` calls `invalidate(key)` and starts the descriptor's load (the visible-key `.revalidate` path of `applyHiddenPolicy`) | A failed partial read leaves rows the write moved; a full reload reports its own failure through the handle | Keeping the old rows silently (the progress refresh's catch) — the list would show a link the catalog no longer has |
| D-8 | The `.evidenceGraphsList` arm gets `refresh: CatalogQueryRefresh(follows: .reads(.evidenceGraphs))`, which for each id runs `refreshCached` with `read` = `store.listEvidenceGraphs(projectDir:sourceID: id)` and `merge` = drop every cached row whose `sourceID == id`, then append the re-read rows in Go's order. `invalidateOn` becomes `[]`. `CatalogQueryFollows.ids` adds `mutation?.deletedSourceId` *(new accessor on `CatalogMutation`)* to `.reads(.evidenceGraphs)` when the effects did not name it, commented "S14-16 removes this when DeleteSource reports effects" | A graph's `source_id` never changes, so the rows of one Source are exactly what one re-read replaces. Appending is safe because every reader goes through `EvidenceGraphIndex`, which keeps each Source's rows in their order. The deleted Source follows the Promote precedent, so a delete uses the same refresh | Sorting the merged list by name — no reader needs flat order; keeping `.deletedSource` in `invalidateOn` — a whole reload, which A-9 removes |
| D-9 | `CatalogList.evidenceGraphs` and the `.evidenceGraphs` case of `CatalogQueryRegistry.invalidations(for list:project:)` are deleted | The read replaces the list; a key only one mechanism moves | Keeping the list beside the read — two reloads per write |
| D-10 | `CatalogMutation.createdSource` becomes `.createdSource(CatalogSource)`. `CatalogQueryPatch.insertedSourceRow()` *(new)*, `follows: nil`, on the `.sourcesList` arm after `.updatedSourceRow()`: for `.createdSource(let source)` it replaces the row with that id if present, else appends it, through `patchCached`. `.sourcesList` `invalidateOn` becomes `[.deletedSource]`. `invalidationKind` and `navCountTriggers` keep `.createdSource` (the type list's `usedBy` and the sidebar recount still need it) | `createSource` and `copySource` return the enriched row with its activity (`s.Updated = res.Activity`), so the write reports everything the patch needs (rule 4). Appending is enough: `visibleSources` sorts. This is the cache layer's patch slot, not a feature write-through (S12-16) | A feature calling `patchCached` itself — S12-16 removed that hand-written insert; a separate `.copiedSource` mutation — Copy and Add move the list the same way |
| D-11 | `.evidenceGraph` and `.sourceWorkspace` arms take `whenHidden: .markStale` | Working rule "A hidden graph or Source page goes stale; it does not reload". Every place that reads them lists them in `PlaceRegistry.queryKeys` (canvas, composer, Promote, Source page), so `apply(location:)` reloads a stale one on the visit; Promote's and the composer's `readyValue` run after that | `.evict` — drops the camera-adjacent cached frame Back shows; the rule says stale |
| D-12 | `SourcesModel` memoizes `visibleSources` with a private `VisibleRowsToken` *(new)*: the Sources handle revision, the types handle revision, `settings`, the query's locale identifier, and — only when `settings.status != nil` — the graph list and progress handle revisions (a missing handle reads `-1`). The rows and token live in `@ObservationIgnored` properties; reading the revisions keeps the view observing them. `@ObservationIgnored private(set) var visibleSourcesBuildCount = 0` *(new)* counts builds, as a test hook | `CitationComposerContext.refreshMemo` precedent; the token is every input the build reads, so a hit cannot be stale | Moving the Add form into its own view only — the search field and any other body run still re-sort; memoizing in the view with `@State` — the model owns the derived list |

**Rules this plan bends.** None. D-10 is the patch that `macos-client-patterns.md` rule 4 allows ("unless the write reported everything the patch needs"); STEP-9 rewords rule 2's "keeps `.createdSource` (a new row the effects cannot describe)" to say the row comes from the response.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `effects.ReadEvidenceGraphs` *(new)* | const | `effects.ReadEvidenceGraphProgress` | Extends the same registry: one constant, one `KeyTable` arm, one `Reads` entry. Searched `Read =` in `core/database/effects` |
| `ListEvidenceGraphsRequest.source_id` *(new)* | proto field | `CreateArtifactRequest.source_id`, `evidencegraphs.List(c, sourceID)` | Extends the existing request and query; no new RPC |
| `FakeStore.listEvidenceGraphsSourceIDs` *(new)* | test helper | `FakeStore.citationCountsGraphIDs` | Same recording shape |
| `QueryHandle.isStale`, `markStale()`, `clearStale()` *(new)* | property / func | `WorkspaceSession.invalidatedKeys` | The handle is what `onSettled` sees; the session keeps its set for handle-less reads (D-5) |
| `WorkspaceSession.markStale(_:)`, `clearStale(_:)` *(new, private)* | func | the inline `invalidatedKeys.insert/remove` calls | Replace those calls so both flags move together |
| `WorkspaceSession.refreshCached` *(new)* | func | `patchCached`, the inline `evidenceGraphProgress` refresh | `patchCached` has no read or fetching state; the progress refresh is inline and keeps rows on failure. Searched `refresh`, `merge`, `patchCached` in `Session/` |
| `CatalogRead.evidenceGraphs` *(new)* | enum case | `CatalogRead.evidenceGraphProgress` | Same enum, wire-checked by `TestSwiftVocabularyMatchesGo` |
| `CatalogMutation.deletedSourceId` *(new)* | accessor | `CatalogMutation.evidenceGraphId` | Same pattern for a mutation-named id |
| `CatalogQueryPatch.insertedSourceRow()` *(new)* | static func | `CatalogQueryPatch.updatedSourceRow()` | Same slot and shape; `updatedSourceRow` only replaces and reads `patchedSource`, which a create must not set (it would also patch the page shell) |
| `SourcesModel.VisibleRowsToken`, `visibleSourcesBuildCount` *(new)* | struct / test hook | `CitationComposerContext.MemoToken`, `SourcesModel.searchKeyCache` | Same memo-by-revision shape; the search-key cache stays as the inner memo |

**Builds on.** S12-25's keyed reads (`effects.Reads`, `WriteEffects.reads`, `CatalogQueryFollows.reads`); S12-34's descriptor `patches` / `refresh` slots; the progress refresh; `add-workspace-place` "Refresh" and "Mark stale" rows.
**Deletes.** `effects.ListEvidenceGraphs` and its `Lists` entry; the `Lists` fields on the `evidence_graphs` and `evidence_graph_artifacts` registry entries; `TestEvidenceGraphsList` (its job moves into `TestReadsDeclare`, T-2); the `ListEvidenceGraphs` rows in `lists_test.go` and `write_effects_test.go`; `CatalogList.evidenceGraphs`; the `.evidenceGraphs` case in `CatalogQueryRegistry.invalidations(for list:project:)`; `.deletedSource` from the `.evidenceGraphsList` arm's `invalidateOn` and `.createdSource` from the `.sourcesList` arm's; the `.evidenceGraphsList` arm comment that begins "Graph and link writes report the `evidence_graphs` list"; the `"evidence_graphs"` mention in the `WriteEffects.lists` proto comment; `source-layer-data-model.md`'s "flag the graph list (`effects.ListEvidenceGraphs` in `effects.Set.Lists`)".
**Next adds.** S14-16 gives `DeleteSource` effects: the graph read then rides `WriteEffects.reads` and the `deletedSourceId` line in `CatalogQueryFollows.ids` is deleted. Another per-key list read (a Source's Artifacts) is one `Read` constant, its paths, one `CatalogRead` case and one `refreshCached` refresh.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A hidden graph key that is stale when its canvas, composer or Promote opens | `PlaceRegistry` places `.sourceGraph`, `.sourceCitationComposer`, `.sourcePromote`; `EvidenceGraphModel.prepare`, `CitationComposerContext.waitForGraph`, `PromoteModel.graphSnapshot` | `apply(location:)` warms every listed key, and `ensureQuery` reloads a stale one; `readyValue` waits on that load. `viewIdentity` is the graph id, so the view resets on a new graph as today. Shows the cached rows with `isRereadingGraph` while reloading (stale-while-revalidate) |
| A hidden Source page key that is stale on revisit | `.sourceDetail` place; `SourcePageModel` (`reread`) | Same: listed in `queryKeys`; the page already rereads on appear |
| The graph list handle `isFetching` during a partial re-read | `evidenceGraphPresence`, `artifactIDsInNoGraph`, `evidenceGraphCountsLoading` (reads `read.isLoading`, which a value keeps false) | `.unknown` / `nil` while it runs, as during a full refetch today; counts keep showing |
| A list handle stale while `.ready` | `onSettled`, `WorkspaceDestinationHostTests` (`isSettled`) | `isSettled` is false until the reload; `onSettled`'s `revision` change after the reload publishes |
| A re-read for a Source the list does not hold (new Source) | `.evidenceGraphsList` merge | Drop finds nothing, append adds its Default graph |
| A re-read for a deleted Source | same | Returns none; its rows are dropped |

### Scale

At the map's Scale (a census Source with 60 household graphs and 300 Citations on one page; 3,000 Sources):

- **Link, graph and Artifact writes** report one Source in `reads.evidence_graphs`. Visible list: one `ListEvidenceGraphs(source_id)` (an index seek returning ≤ 60 graphs and their links) and a merge over the cached list (an O(rows) filter over ~3,060 rows on the main actor, then the index rebuild on the next read, which today's full reload also pays). Hidden list: one flag. Today: a project-wide read (20.9 ms at 3,060 graphs, A-9) on every such write.
- **Add / Copy Source**: one row appended to the cached Sources list, no `listSources` (14 ms at 3,061 Sources). The Default graph's row arrives through the graph read for the new Source only.
- **Writes that name graphs** (`evidenceGraphs`): each hidden visited graph key gets a flag instead of a five-call reload (`listSubjects`, positions, observations, memberships, event titles). A session that visited all 60 graphs reloads only the one on screen.
- **Typing in Add Source or the search field**: a memo hit is a token compare; a build stays O(n log n) over 3,000 Sources and runs once per input change.
- No budget test: AC-1.2 and AC-5.1 assert the count of reads and builds, which is the property; the audit's timings are the reference.

### Edges

- **Assumptions.** (1) A graph's `source_id` never changes (`evidencegraphs.Update` writes name, description and links only; "No moving a Subject and no splitting a graph"), so dropping one Source's rows and appending its re-read is exact — T-6. (2) Readers do not depend on the flat list order: `EvidenceGraphIndex` groups by Source preserving order; `evidenceGraphsBySource`, `evidenceGraphs(sourceID:)`, `evidenceGraph(id:)` and the composer's `readyValue` read through it or only check non-nil — T-6 asserts per-Source order after a merge. (3) `Resolve` walks a deleted link's `up("graph_id", "source_id")` through ghosts when its graph is deleted in the same batch, as `SourceScope` already does — T-1 "graph delete". (4) `createSource` and `copySource` return the row `listSources` would (both go through `enrichSourceProto`, with `Updated` set) — T-8 compares fields.
- **Neighbors.** A Source title edit: no graph read (D-2), still patches the row through `.updatedSourceRow` — T-1 "title edit". An Artifact label edit: no link row, no read — T-1. An Artifact create with links: one read for its Source — T-1. A Source type change (`.changedSourceType`): the Sources list still patches through `patchedSource`; `insertedSourceRow` ignores it — T-8. The Sources list on `.deletedSource`: still invalidates (row removal is not patched here) — T-8.
- **Removed inputs.** A deleted Source's graphs leave the list through the refresh (AC-1.4). A deleted graph leaves through its Source's re-read (T-6). An Artifact deleted with its links: the link releases report the read for its Source (T-1).
- **Concurrency.** Two partial re-reads of the same key overlap: the first to land clears `isFetching` while the second is out; the second merges its Source when it lands. Each merge reads the current value after its `await`, so neither loses the other's Source. A full load in flight when a merge lands is cancelled and restarted by `replaceValue`, so an older full answer cannot overwrite the merge — T-7.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — `source_id` is parsed as a UUID before the read; the rows are the catalog's own, bounded by the same query as today.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `ReadEvidenceGraphs Read = "evidence_graphs"` *(new)*, doc "the graphs of a Source, with their linked Artifacts (the graph list), keyed by Source"; `KeyTable`: `case ReadEvidenceGraphs: return "sources"`; `Reads = []Read{ReadCitations, ReadEvidenceGraphProgress, ReadEvidenceGraphs}`. Registry per D-2. `ListEvidenceGraphs` deleted; `Lists = []List{ListVocabulary, ListPropertyUsage}`; the `List` doc's example list follows |
| `api/ffi/handlers` | `ListEvidenceGraphs`: `var sourceID []byte; if s := strings.TrimSpace(req.GetSourceId()); s != "" { sourceID, err = parseID(s) … }` before `withProjectCatalog`; `evidencegraphs.List(c, sourceID)`. Doc: "returns every graph of the project, or of one Source, by name" |

Errors: a malformed `source_id` → the shared invalid-id error `parseID` returns (`errInvalidID`).

### FFI

`api/proto/engine.proto`: `ListEvidenceGraphsRequest` gains `string source_id = 2; // empty: every graph of the project`. The `WriteEffects.lists` comment drops `"evidence_graphs"`; the `reads` comment reads "each with the ids of the read's key table (Evidence graphs for "citations" and "evidence_graph_progress", Sources for "evidence_graphs")". Regenerate with `./scripts/generate-proto.sh`. No new method.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Stale flag | `QueryHandle.isStale` | `macos/App/Features/Workspace/Session/QueryHandle.swift` | D-5 |
| Partial refresh primitive | `WorkspaceSession.refreshCached` | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` | D-6, D-7 |
| Graph list refresh | `.evidenceGraphsList` arm `refresh` | `CatalogQueryRegistry.swift` | D-8 |
| Sources insert | `CatalogQueryPatch.insertedSourceRow()` | `CatalogLastUpdated.swift` | D-10 |
| Hidden policy | `.evidenceGraph`, `.sourceWorkspace` arms | `CatalogQueryRegistry.swift` | D-11 |
| Rows memo | `SourcesModel.visibleSources` | `macos/App/Features/Sources/SourcesModel.swift` | D-12 |

Sketch (signatures only):

```swift
// WorkspaceSession
func refreshCached<Value, Part>(
    _ key: CatalogQueryKey,
    as _: Value.Type,
    read: () async throws -> Part,
    merge: @escaping (inout Value, Part) -> Void
) async

// CatalogMutation
var deletedSourceId: String? // .deletedSource(let id) → id

// CatalogQueryPatch
static func insertedSourceRow() -> Self
```

States: no visible change. Loading: a partial re-read shows the cached rows. Error: a failed partial re-read becomes a full reload whose failure shows where it does today (`SourceGraphsSection`'s error, `evidenceGraphsRead` `.failed`). Accessibility: unchanged.

### Strings

None — no copy changes; no action changes what it covers.

## 7. Steps

**STEP-1 — The graph list is a keyed read in Go.**
Files: `core/database/effects/registry.go`, `core/database/effects/registry_completeness_test.go`, `core/database/effects/lists_test.go`, `core/database/effects/evidence_graphs_test.go`, `core/database/effects/reads_test.go`, `core/database/effects/reads_resolve_test.go`.
Do: D-1, D-2. Delete `TestEvidenceGraphsList`; extend `TestReadsDeclare` (T-2, T-3); drop the `ListEvidenceGraphs` rows from `lists_test.go`; turn each `set.Has(effects.ListEvidenceGraphs)` assertion in `evidence_graphs_test.go` into a `reads` assertion for the Source (T-1); add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/...` passes; `rg -n "ListEvidenceGraphs" core` prints nothing.

**STEP-2 — `ListEvidenceGraphs` reads one Source.**
Files: `api/proto/engine.proto`, generated `api/proto/engine/*.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` (by the script), `api/ffi/handlers/evidence_graphs.go`, `api/ffi/handlers/evidence_graphs_test.go`, `api/ffi/handlers/write_effects_test.go`. Skill: `add-ffi-handler`, `use-catalog-session`.
Do: D-3 and the proto comments in §6 FFI; run `./scripts/generate-proto.sh`. `assertGraphEffects` checks `reads` holds `evidence_graphs` with the graph's Source instead of `lists`; T-4 and T-5; drop `ListEvidenceGraphs` from the `write_effects_test.go` row.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes, including `TestSwiftVocabularyMatchesGo` failing only on the Swift enums STEP-5 changes (expected until then — note the failure text in the step report).

**STEP-3 — The store asks for one Source.**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/ProvenenciaTests/OnboardingModelTests.swift` (`ThrowingStore`), `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift`, `CatalogQueryRegistry.swift` (loader passes `sourceID: nil`). Skill: `add-swift-test`.
Do: D-4. Every existing call passes `sourceID: nil`. Add T-9.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/EvidenceGraphStoreTests CODE_SIGNING_ALLOWED=NO` passes. Later steps run this same command with other `-only-testing:` targets.

**STEP-4 — A stale key is not settled, and a partial re-read looks like a refetch.**
Files: `QueryHandle.swift`, `WorkspaceSession.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-workspace-place`.
Do: D-5 (route every `invalidatedKeys` insert and remove through `markStale(_:)` / `clearStale(_:)`), D-6, D-7. Add T-10, T-11, T-12.
Check: `xcodebuild test` (STEP-3's command) with `-only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/SourcesModelTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests` passes.

**STEP-5 — The graph list follows the keyed read.**
Files: `CatalogMutation.swift`, `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`, `SourcesModelTests.swift`, `EvidenceGraphModelTests.swift`. Skill: `add-workspace-place`.
Do: D-1 (Swift case and doc comments on `CatalogRead` and `CatalogWriteEffectsRead`), D-8, D-9. Rewrite the `CatalogQueryRegistryTests` that expect `.evidenceGraphsList` invalidations from a list or from `.deletedSource` (lines 865–874, 1249, 1268–1269) to expect none; add T-6, T-7, T-13.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/ -run TestSwiftVocabularyMatchesGo` passes; `xcodebuild test` (STEP-3's command) with `-only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/WorkspaceSessionTests` passes; `rg -n "CatalogList.evidenceGraphs|case evidenceGraphs = \"evidence_graphs\"" macos/App/Features/Workspace/Session/CatalogMutation.swift` shows only the `CatalogRead` case.

**STEP-6 — Add and Copy Source insert their row.**
Files: `CatalogMutation.swift`, `CatalogLastUpdated.swift`, `CatalogQueryRegistry.swift`, `SourcesModel.swift`, `SourceCopyModel.swift`, tests that apply `.createdSource` (`WorkspaceSessionTests.swift` 1271 and 1284, `SourcesModelTests.swift` 396, `CatalogQueryRegistryTests.swift` 644, 865, 968, 1169).
Do: D-10. `SourcesModel.create()` applies `.createdSource(created)`; `SourceCopyModel` applies `.createdSource` with the returned Source. The hidden-list test at `SourcesModelTests.swift` 396 switches its trigger to `.deletedSource(id:)` on a Source it then removes from the store, keeping what it proves. Add T-8.
Check: `xcodebuild test` (STEP-3's command) with `-only-testing:ProvenenciaTests/SourcesModelTests -only-testing:ProvenenciaTests/SourceCopyModelTests -only-testing:ProvenenciaTests/CatalogQueryRegistryTests` passes.

**STEP-7 — Hidden graph and Source page keys go stale.**
Files: `CatalogQueryRegistry.swift`, `CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`.
Do: D-11. Add T-14 and T-15.
Check: `xcodebuild test` (STEP-3's command) with `-only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/EvidenceGraphModelTests -only-testing:ProvenenciaTests/CitationComposerModelTests -only-testing:ProvenenciaTests/SourcePageModelTests` passes.

**STEP-8 — The Sources list builds its rows once per input change.**
Files: `SourcesModel.swift`, `SourcesModelTests.swift`.
Do: D-12. Add T-16.
Check: `xcodebuild test` (STEP-3's command) with `-only-testing:ProvenenciaTests/SourcesModelTests` passes.

**STEP-9 — Docs.**
Files: `docs/macos-client-patterns.md` (cache contract rule 2: `reads` names `evidence_graphs` by Source; Add and Copy Source insert their returned row through `.createdSource(source)`; `.evidenceGraph` and `.sourceWorkspace` join the `.markStale` keys. Rule 4: a create that returns its row patches list membership through a descriptor patch (`insertedSourceRow`). The `QueryHandle` table row: `isStale`, and `isSettled` = ready, not refetching and not stale), `.agents/skills/add-workspace-place/SKILL.md` (cache table: the Refresh row names the graph list's per-Source merge through `refreshCached`; the Patch row names `insertedSourceRow`), `docs/source-layer-data-model.md` § 7.1 (last bullet: graph and link rows report the `evidence_graphs` read for their Source, `effects.ReadEvidenceGraphs`), `completed.md`, `deployment-plan.md` (map row, and the S14-11 scope entry's `swift_vocabulary_test.go` clause corrected to "follows"), this plan (status, Landed note).
Do: per the Definition of done.
Check: links resolve; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-11-reload-only-what-moved.md` reports no errors.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/evidence_graphs_test.go` (the existing test whose assertions sit at lines 141–288) | Source create (Default graph + link) → `Reads` = `[{evidence_graphs, [src]}]`, no `Lists`; link row alone → same read; graph create, rename, delete (with released links) → read for the Source; Source title edit → no `evidence_graphs` read; Artifact label edit → none; Artifact create with links → read for its Source | AC-1.1, D-2 |
| T-2 | Go | `core/database/effects/reads_test.go` `TestReadsDeclare` (its `want` map gains a row) | `ReadEvidenceGraphs: {"evidence_graph_artifacts", "evidence_graphs"}`; the test's own loop then holds both paths to `KeyTable() == "sources"`. `TestEvidenceGraphsList` in `registry_completeness_test.go` is deleted | D-1, D-2 |
| T-3 | Go | `core/database/effects/reads_test.go` `TestReadsDeclare` (one more bad-entry check) | `Effect{Reads: []ReadPath{{ReadEvidenceGraphs, self()}}}.validate("evidence_graphs")` → error (the path ends at `evidence_graphs`, not `sources`) | D-1 |
| T-4 | Go | `api/ffi/handlers/evidence_graphs_test.go` `TestListEvidenceGraphs` subtests "one Source" and "bad source id" *(new)* | `source_id` = srcB → one graph with `pageB`; empty → every graph as today; `"x"` → error | AC-1.3 |
| T-5 | Go | `api/ffi/handlers/evidence_graphs_test.go` `assertGraphEffects` | Every graph and link handler test asserts `reads` ∋ `{evidence_graphs, [source]}` and no `lists` ∋ `evidence_graphs` | AC-1.1 |
| T-6 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `graphListMergesOneSourcesReRead` *(new)* | Cached list S: [Default, Jones], T: [Default]; graph list key visible; FakeStore S now [Default, Jones(+page2), Smith]; apply effects `reads: [evidence_graphs: [S]]`; `settleWrites()` → `listEvidenceGraphsSourceIDs == ["S"]`; S's rows are [Default, Jones, Smith] in that order with page2 on Jones; T's rows unchanged | AC-1.2, D-8 |
| T-7 | Swift | `WorkspaceSessionTests.swift` `graphListPartialReReadSetsFetchingAndFallsBackOnFailure` *(new)* | A gated FakeStore read: while it waits, `evidenceGraphsHandle.isFetching` and `evidenceGraphPresence` is `.unknown`; then `listEvidenceGraphsError` set → after `settleWrites()` a full `listEvidenceGraphs(sourceID: nil)` call was made | AC-1.6, D-6, D-7 |
| T-8 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `addSourceInsertsTheReturnedRow` *(new)* and `SourceCopyModelTests.swift` `copyInsertsTheReturnedRow` *(new)* | 2 cached Sources; create "Deed" → 3 rows, the new one equal to the returned row; `listSources` call count unchanged; a second apply of the same `.createdSource` keeps 3 rows; with the list never warmed → no handle created; `.changedSourceType` does not insert | AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| T-9 | Swift | `macos/ProvenenciaTests/EvidenceGraphStoreTests.swift` `listsOneSourcesGraphs` *(new)* | Real store: two Sources → `listEvidenceGraphs(sourceID: a)` returns only a's graphs | AC-1.3, D-4 |
| T-10 | Swift | `WorkspaceSessionTests.swift` `staleKeyIsNotSettled` *(new)* | Warm the Sources list, hide it, apply `.deletedSource(id:)` → `isStale`, `!isSettled`, `status == .ready`, `!isFetching`; `query` it → after `waitForQuery`, `!isStale`, `isSettled` | AC-4.1, AC-4.2 |
| T-11 | Swift | `WorkspaceSessionTests.swift` `setQueryValueClearsStaleAndPatchKeepsIt` *(new)* | Stale key + `patchCached` → still `isStale`; + `setQueryValue` → `!isStale` | AC-4.3, AC-4.4, D-5 |
| T-12 | Swift | `WorkspaceSessionTests.swift` `hiddenGraphListGoesStaleOnRead` *(new)* | Graph list warmed, `visibleKeys` without it; apply `reads: [evidence_graphs: [S]]` → no store call, key stale; `query` → one full load | AC-1.5 |
| T-13 | Swift | `WorkspaceSessionTests.swift` `deletedSourceDropsItsGraphs` *(new)* | Visible list with S and T; FakeStore drops S; apply `.deletedSource(id: "S")` → `listEvidenceGraphsSourceIDs == ["S"]`, S's rows gone, T's kept, no `nil` call | AC-1.4, D-8 |
| T-14 | Swift | `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` `graphAndSourcePageKeysMarkStaleWhenHidden` *(new)* | `hiddenPolicy(.evidenceGraph(…)) == .markStale`; `hiddenPolicy(.sourceWorkspace(…)) == .markStale` | D-11 |
| T-15 | Swift | `WorkspaceSessionTests.swift` `hiddenGraphReloadsOnlyWhenRevisited` *(new)* | Graphs G and H warmed; on H's place apply `evidenceGraphs: [G, H]` → H reloads, G does not (FakeStore `listSubjects` graph ids); `apply(location:)` G's canvas → G loads once | AC-3.1, AC-3.2 |
| T-16 | Swift | `SourcesModelTests.swift` `visibleSourcesBuildsOncePerInputChange` *(new)* | 3 Sources, A–Z: read `visibleSources`, change `draft.title` 3×, reading after each → `visibleSourcesBuildCount == 1`; change `settings.query` → 2; a Sources revision bump → 3; with a status filter, a progress revision bump → +1; without one → no change | AC-5.1, AC-5.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CatalogQueryRegistryTests` (list invalidations, `.createdSource`), `SourcesModelTests` (hidden-list count test), `EvidenceGraphModelTests.graphPresenceFollowsTheList`, `WorkspaceDestinationHostTests` (`isSettled`), `SourcePageModelTests` (graph list waits), `TestSwiftVocabularyMatchesGo`.

AC-3.3 is covered by T-14 plus the manual check (step 3): the Source page key follows the same `applyHiddenPolicy` path T-15 exercises.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left (none expected)
- [ ] Sweep: `rg -nP "ListEvidenceGraphs\b|effects\.ListEvidenceGraphs|CatalogList\.evidenceGraphs|\.createdSource\b(?!\()"` across `.agents/`, `.cursor/rules/`, `docs/*.md`, `core/`, `api/` and `macos/` leaves only the RPC name `ListEvidenceGraphs` (method, request, response, handler) and `.createdSource(` / `CatalogMutationKind.createdSource`; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (none)
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New (no UI change)
- [ ] Version bump: the next patch in VERSION, core/version.go, MARKETING_VERSION together
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (not gated: none)
- [ ] Domain doc updated: source-layer-data-model.md § 7.1; macos-client-patterns.md rules 2 and 4 and the QueryHandle row; add-workspace-place cache table
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with a Source of at least three graphs (or the 60-graph census when available) and a second Source. Run with `PROVENENCIA_DEBUG_SESSION_APPLY=1` to see place applies.

1. Open the Source page of the census. Open a graph's Edit form and link a second Artifact; save. The Source page's graph row and the Artifact card update; no other Source's row flickers. (A breakpoint or log in `GoStore.listEvidenceGraphs` shows one call with the census's id.)
2. Add a Source from the Sources list. The new row appears at once in the right sort position; the list does not blank or show a progress state. Copy a Source: the same.
3. Open graph A's canvas, then graph B's canvas, and add a Subject on B. Go to Persons, then Back to B, then to A: A loads when shown (its canvas shows its cached rows, then current ones) and not before.
4. With the Sources list showing a count of N in the sidebar, go to Persons, delete a Source from its page, and come back to Sources: the sidebar count changes to N−1 once, after the list reloads.
5. Open Add Source on a project with many Sources and type a title quickly: no lag; the list behind the sheet does not reorder.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A reader depends on the flat graph list order | §6 Edges (2); T-6 asserts per-Source order; stop condition below |
| `isSettled` now false while stale breaks a test that warmed a key, hid it and expected it settled | STEP-4 runs the three suites that read `isSettled`; fix the test only where it asserted the old stale-but-settled state |
| A partial re-read overlapping a full load leaves `isFetching` stuck | `replaceValue` clears it on every value write; T-7 covers the failure path |
| Removing `.createdSource` from the Sources list's `invalidateOn` leaves a Source write unpatched | Only Add and Copy apply `.createdSource`; both now carry the row (STEP-6 sweep) |

**Stop and report if:**

- `evidencegraphs.Update` or any write can change a graph's `source_id` (D-8's merge would be wrong).
- A reader of `evidenceGraphsHandle.value` outside `EvidenceGraphIndex` uses its flat order (`rg -n "evidenceGraphsHandle\?\.value|evidenceGraphsKey\) as \[CatalogEvidenceGraph\]"` finds one that iterates it).
- `CreateSourceResponse.source` or `CopySourceResponse.source` lacks a field `listSources` rows carry (T-8 comparison fails on a field the handler does not set).
- `Resolve` cannot resolve `up("graph_id", "source_id")` for a link released in the same batch as its graph (T-1 graph delete fails).
- A place reads `.evidenceGraph` or `.sourceWorkspace` without listing it in `PlaceRegistry.queryKeys` (D-11's reload-on-visit would not happen).
- A test outside §8's list needs its assertions changed for reasons other than `.createdSource`'s payload, the removed list, or `isSettled`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-11 — Reload only what moved, only when it is on screen
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-14/prs/S14-11-reload-only-what-moved.md`.
