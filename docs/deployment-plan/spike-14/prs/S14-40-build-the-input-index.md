# S14-40 — Build the input index

**Status:** Landed in #746 (2026-11-08)
**Landed:** No version bump. D-13 was amended by the user on 2026-11-08 to excuse persons reached only through a non-canonical birth or death event (9,987 on the large fixture). Differences from the plan: the shared test helpers live in `core/database/conclusionsummaries/conclusionsummariestest`; the test isolator reuses one graph for all rows; `InputScope` calls the exported `Graph.Node` and `Graph.Follow` so the recorder also hears `NoInputs()` reads; `writeInputs` uses `INSERT OR IGNORE` in 200-row statements; `lifeOwners` takes an `*InputScope`; `perf_test.go` uses `databasetest.Fastest/Since/WithinBudget` after the base moved. Manual check §10: none.
**Spike / slice:** Spike 14 · Slice 8 — Summary rows know their inputs
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: engine only; every row and document is composed as before)
**Board:** —
**Depends on:** —
**Blocks:** S14-42
**Branch:** `s14-40-build-the-input-index`
**Version bump:** none — a catalog format step and an engine speed-up nothing on screen shows; S14-42 carries the researcher-visible fix, as S14-10 did for its format step ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-catalog-model`](../../../../.agents/skills/add-catalog-model/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md) (its checklist gains a line)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/catalog-queries.mdc`, `.cursor/rules/versioning.mdc`
**Context:** [design](../summary-inputs.md); working rules "(Batch 8) Summary rows record their inputs; invalidation is a lookup", "(Batch 8) An input is a handle's values, one bridge's links at one end, or an association's values", "(Batch 8) Inputs are attributed per row inside the batch, and isolated composition is the oracle", "(Batch 8) One input index, owned by `searchindex`"; map review log § Batch 8 ([`../deployment-plan.review.md`](../deployment-plan.review.md))
**Planned:** 2026-11-07 against `main` @ `26ada55`
**Review:** [`S14-40-build-the-input-index.review.md`](S14-40-build-the-input-index.review.md) — Fits with changes — Ready with changes, 2 rounds; no open blockers; To user none

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

Every person, event and place summary row records which other handles, links and associations it was composed from, and the catalog stores that as an index beside the search documents. Nothing reads the index yet. S14-42 switches write invalidation to it and deletes the hand-written inverse (`conclusionsummaries.Dependents` and its neighbours). This PR proves the index is right. It equals what composing each row alone reads, it covers what `Dependents` finds today, and it equals a rebuild after every write the upkeep tests make. The PR also removes a quadratic cost from search reprojection (one graph per call instead of one per 500-id batch), and builds a person's place graph from its birth and death events only, as `PersonLifeOwners` already does.

**Why now.** First PR of Batch 8 (user, 2026-11-07); S14-38's stale documents are fixed on top of it in S14-42.

**What stays.** Every summary row, list and search document reads as before. Invalidation, `Dependents`, `AssociationEndpoints` and `subjects.Delete` are untouched.

## 2. User stories

**US-1.** As the engine, each summary row knows its inputs, so a later write can find it by lookup.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Any person, event or place in the scenario worlds (§8 T-2) or the `TestLargeSourceBudgets` fixture, composed in one batch through `conclusionsummaries.RowsForSearch` | `Summaries.Inputs(id)` equals the keys the per-graph test recorder reports when that row alone is composed, minus the row's own values |
| AC-1.2 | James (person) with a birth located at Yorkville, part_of Toronto, part_of Ontario | James's inputs hold `(birth, values)`, `(birth, location, event)`, the location's `(A, values)`, `(Yorkville, values)`, `(Toronto, values)` and `(Ontario, values)`, each place's `(place_relationship, from)` key, and each part_of association's values. They hold no key for the places of James's other (non-birth, non-death) events, and not `(James, values)` |
| AC-1.3 | Any handle X in those worlds | The rows whose inputs hold `(X, values)` or any link key of X include every id `conclusionsummaries.Dependents([X])` returns, except X's immediate part_of parents, its succession neighbours (detail-page hops), and persons reached only through a birth or death event with no accepted member (D-13). Extras are counted per kind and logged |

**US-2.** As the engine, the index is stored and kept with the search documents.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Any write in `TestRebuildEqualsUpkeep_Scenarios` or `_SeededSequences` | `catalog_search_inputs` equals the table `searchindex.RebuildAll` writes (both sorted) |
| AC-2.2 | A filed participation A between James and a birth | `catalog_search_inputs` holds `(A → James, participation, person)` and `(A → birth, participation, event)` |
| AC-2.3 | A handle's document is dropped (its last member unpromoted, or merged) | Its rows as a dependent are gone after that `ReprojectHandles` |
| AC-2.4 | A catalog whose `catalog_search_meta.projection_version` is 10 | After `searchindex.EnsureCatalog`, `catalog_search_inputs` holds rows and the version is 11 |

**US-3.** As the engine, search reprojection reads each handle's graph once per call.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ReprojectHandles` of 1,200 ids (3 batches) | One graph is built (`newGraph` called once); zero ids build none |
| AC-3.2 | The first part_of filed from the 6,000-event city (`TestLargeSourceBudgets` "second part_of from the city") | Faster than before this PR; planning measured 3.64 s → 1.63 s on one machine. The PR reports both numbers |
| AC-3.3 | Composing every person and every event row on the large fixture | Under `budgetComposeAll` (500 ms); planning measured 95 ms + 125 ms before this PR |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache.InputKey`, `InputScope`, `Graph.Inputs`, `Graph.NoInputs`, `Graph.AssociationInputs`, a per-graph test recorder | STEP-2, D-1, D-2, D-9, D-12 |
| Composers record per row through scopes; `placeGraph` records each place's reads and unions them per row; `Summaries.Inputs` | STEP-3, D-3 – D-6 |
| A person's place graph comes from its birth and death events only | STEP-3, D-5 |
| Migration `000062.sql`: table `catalog_search_inputs`, registered in `catalogmodel` | STEP-4, D-7 |
| `searchindex` writes, clears and rebuilds the table, association ends included; `ProjectionVersion` 11 | STEP-5, D-8, D-10 |
| `ReprojectHandles` composes on one graph per call | STEP-5, D-11 |
| Rebuild-equals-upkeep compares the table | STEP-6, AC-2.1 |
| Oracle, coverage and budget checks on the large fixture | STEP-1, STEP-7, AC-1.1, AC-1.3, AC-3.2, AC-3.3 |
| `catalog-queries.mdc` line amended; `add-searchable-kind` line; docs | STEP-8 |

| Out | Home |
| --- | --- |
| Invalidation from the index; deleting `Dependents`, `AssociationEndpoints`, `linkEnds`, `movedRoots`, `SetDisplayDeps`; `subjects.Delete` | S14-42 |
| The `core/database/graphcache/sql.go` copy of `database.SQLInPlaceholders` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (the S13-37 row): `core/database` imports `graphcache`, so the helper moves to a leaf package first (planning, 2026-11-07) |
| Recording inputs for the kinship memos (`canonlinks`) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the "Spike 14 Batch 8 design" row |
| The remaining cost of the first part_of filed from a city after the one-graph fix | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the "first `part_of` filed from a city" row, updated with the new figure in STEP-8 |

## 4. Current state

Verified at `26ada55`.

| Area | Where | Today |
| --- | --- | --- |
| Graph reads | `core/database/graphcache/graph.go` `Graph`, `Node`, `Nodes`, `Over`, `Node` struct (`SubjectTypeID`, `Merged`, `Values` by property id, `Links`), `(*Node).Kept`; `follow.go` `Follow`; `vocab.go` `Vocabulary`, `PropertiesByID`, `TypeKeysByID`, `PropertyID` | `Node` loads and returns a node; `Nodes` prefetches; `Follow(id, hop)` calls `g.Node(id)` and filters `n.Links` by the hop's bridge, end and term. No recording. `Graph` has no lock: callers are serialized by the catalog session |
| Test hook precedent | `graphcache` `(*Graph).SetQueryCounterForTest` (graph.go line 412), a per-graph func field | The shape of D-12 |
| Bridges | `core/connectrules/connectrules.go` `LookupBridge` (line 297), `Bridge.Endpoints`; `core/hops/hops.go` `Hop.Bridge`, `From`, `PlacesOfEvent`, `ParentsOfPlace`, `SuccessorsOfPlace`, `PredecessorsOfPlace` | |
| Person rows | `core/database/conclusionsummaries/compose.go` `composePersons` (line 547), `onePerson`, `lifePicks`, `leadLife`, `eventPlaceGraph` (line 787), `leadPictures`, `PersonLifeOwners` (line 751), `lifeOwners` | Batch: `follow(EventsOfSubject)` per person and `g.Nodes(events)`; then `eventPlaceGraph` over every subject event of every person (`Node(e)`, `follow(e, PlacesOfEvent)`, then `buildPlaceGraph`); `leadPictures` (raw SQL through `gallery.ForEntities`). Per row: `onePerson` → `lifePicks` → `leadLife` → `eventPlaces`. The row shows only the lead birth's and death's places. `PersonLifeOwners` builds its place graph from `lifePicks` births and deaths only |
| Event rows | `composeEvents` (line 865), `oneEvent`, `eventPlaces` | Batch: `follow(SubjectsOfEvent)`, `follow(PlacesOfEvent)`, `g.Nodes`, `buildPlaceGraph(places)`, `leadPictures`. Per row: own values, `eventPlaces` (`follow`, `Node(place)`, `pg.FoldLocations`), subjects (`follow`, `Node(person)`) |
| Place rows | `composePlaces` (line 1141), `placeShell`, `finishPlaces` | Shell from own values; `finishPlaces` builds one `placeGraph` from all rows and sets `Parents` per row through `pg.ParentChain` |
| Place chains | `compose.go` `buildPlaceGraph` (line 1200), `rememberPlaceLink` (line 1341), `placeStep`; `placechains.go` `placeGraph`, `linkHolds`, `placeNeighborhood` (line 295, a fourth caller of `buildPlaceGraph`) | BFS over `ParentsOfPlace` for `summaryChainDepth` (8) rounds from the seeds: per frontier place, `Node`, `placeShell`, `follow`, and `Node(assoc)` per link. `rememberPlaceLink` stores every link whatever its dates; `linkHolds` filters at query time. Places first reached in the last round are not read |
| Summaries | `compose.go` `Summaries`, `NewSummaries` (several callers, tests included), `RowsForSearch` (line 407), `persons`, `events`, `places`; `display_register.go` `checkDisplay` | `RowsForSearch` composes with `store` false; its only caller is `searchindex`. `persons`, `events`, `places` sort their rows (`sortPersons`, …), so a row's index is not its composition order. `checkDisplay` calls all three |
| Dependents | `core/database/conclusionsummaries/dependents.go` `Dependents`, `AssociationEndpoints` | Hand-written inverse; returns a place's parents and succession neighbours too; reaches a person from a place only through a birth or death event |
| Reprojection | `core/database/searchindex/handles.go` `ReprojectHandles` (line 42), `reprojectHandleBatch`, `loadHandleSummaries` | `sqlbatch.Each` in batches of 500; each batch calls `RowsForSearch(graphcache.Over(q), w)`, a fresh graph, so a place shared by many events is reloaded with all its links per batch. Associations have `kind == ""` and are skipped. A gone id drops its documents. No test calls `reprojectHandleBatch` or `loadHandleSummaries` directly |
| Rebuild | `searchindex.go` `ClearAll` (line 245), `RebuildAll`, `ProjectionVersion` (10), `NeedsRebuild`, `EnsureCatalog` (line 638) | Version mismatch rebuilds on open |
| Recompute | `core/database/autoreconciler/autoreconciler.go` `RecomputeReported` | Reprojects the roots and, when every root's rows moved, their dependents; otherwise only the dependents of moved roots |
| Catalog model | `core/database/catalogmodel/model.go` | `catalog_search_docs`, `catalog_search_meta`, `auto_reconciler_*` are `BucketSkip` |
| Migrations | `core/database/migrations/` | Last step `000061.sql`. New tables are `STRICT` (`evidence_graphs`) |
| Upkeep tests | `core/database/autoreconciler/autoreconciler_test.go` `(*fixture).assertUpkeepEqualsRebuild` (line 1103), `handleDocs(t, q autoreconciler.Querier)` | Compare value rows and person / event / place documents with a rebuild in a rolled-back transaction |
| Rebuild test | `core/search/handles_test.go` `TestRebuildIndexesHandles` (line 343) | Sets the projection version back and checks open rebuilds documents |
| Perf fixture | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `largeFixture` (embeds `*fixture`, whose catalog is `c`), `fastest`, `since`, `over`, subtest "second part_of from the city" (the first part_of is untimed, line 304) | 8,115 persons, 10,067 events, 51 places, 24,208 associations |
| Test worlds | `core/database/evidencelayer/evidencelayertest/evidencelayertest.go` `(*Records).Participation`, `Locate`, `PlaceRel` (each returns the bridge Subject); `conclusionsummaries` tests: `walks_test.go` `at`, `named`, `promoteSubject`; `placechains_test.go` `placeSubject`; `display_cache_test.go` `TestDisplayCacheFollowsAPlaceThreeStepsUp` (Country / Region / City / Town) | Reused by T-2, T-3, T-5 |
| Rule | `.cursor/rules/catalog-queries.mdc` line 16: "An open-catalog read of the canonical graph follows a `hops.Hop` with `graphcache.Graph.Follow`. Do not hand-write the association join in SQL …" | Amended in STEP-8 |

**Planning probes** (throwaway, reverted): recording `Node` / `Follow` while composing each row alone gave 96,676 entries, median 5 per row, identical to `Dependents` on 300 sampled handles at the coarser key. Composing per-row place graphs instead of the batch one cost 4× (persons 95 → 386 ms, events 125 → 533 ms). One graph per `ReprojectHandles` call took the first part_of from the city from 3.64 s to 1.63 s.

**Precedent.** `graphcache.RegisterNodeMemo` is the owner-agnostic `graphcache` API `canonlinks` uses. `searchindex` writing a derived table in the reprojection transaction is `catalog_search_docs` itself. `TestRebuildIndexesHandles` is the shape of T-6. `core/database/evidencegraphs/export_test.go` `SetMintRef` is the shape of D-11's seam.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `graphcache.InputKey` *(new)* `struct { ID []byte; Bridge, Endpoint string }`. `Bridge == ""` is the handle's values; otherwise `Bridge` is the bridge type key and `Endpoint` the endpoint Property key where `ID` sits (`hop.Bridge()`, `hop.From()`). An association's values are `{A, "", ""}` | The three kinds of the working rule in one comparable shape; product bridges only (`connectrules`), so a key carries no origin | A kind enum — the empty bridge already says it |
| D-2 | `core/database/graphcache/inputs.go` *(new)*. `func (g *Graph) Inputs(dependent []byte) *InputScope` records keys and leaves out `{dependent, "", ""}`. `g.Inputs(nil)` records every key (used per place, D-4). `func (g *Graph) NoInputs() *InputScope` records nothing. `InputScope` has `Node(id)`, `Nodes(ids)`, `Follow(id, hop)`, `Add(keys ...InputKey)` (applies the self rule) and `Keys() []InputKey` (deduplicated, sorted by ID, Bridge, Endpoint). `Node(id)` records `{id, "", ""}`, a missing node included. `Follow(id, hop)` records `{id, hop.Bridge(), hop.From()}` and `{l.Association, "", ""}` for each link it returns, and nothing for `id`'s values. `Nodes` records nothing. There is no `Graph()` accessor: a composer reads other handles only through the scope. `Graph.Node`, `Nodes` and `Follow` stay as they are. The key rules live once, in unexported `nodeKey(id) InputKey` and `followKeys(id, hop, links) []InputKey` in `inputs.go`; the scope adds the self rule, and the test recorder (D-12) calls the same two (review M-8) | An explicit value per row cannot be misattributed, and the kinship memos take the same API later | A "current dependent" field on `Graph` — shared mutable state |
| D-3 | Composers create one scope per composed row before any per-row read: `scopes := map[string]*graphcache.InputScope{}` filled with `g.Inputs(id)` for each row in `need` (persons, events), and per summary in `finishPlaces`. These functions take a `*graphcache.InputScope` in place of `*graphcache.Graph` for their reads, and every `g.Node` and `follow` inside them goes through the scope: `lifePicks`, `onePerson`, `leadLife`, `eventPlaces`, `oneEvent`. The package helper `follow` becomes `follow(in *graphcache.InputScope, id, hop)`, nil-safe as today. Batch-phase prefetches (`g.Nodes`) and the batch-phase follows that only collect ids for a prefetch call `follow(g.NoInputs(), …)`; each row repeats its own follows through its scope. `PersonLifeOwners`, `lifeOwners` and `placeStep` pass `g.NoInputs()` | One code path per row, and every key a row holds was recorded by a scope call it made | Recording in batch phases — cannot attribute |
| D-4 | `buildPlaceGraph(g, v, seeds)` creates one `g.Inputs(nil)` scope per place the first time it expands that place, keeps it in a local map for both of the round's loops, reads that place through it (`Node`, `follow`, and `rememberPlaceLink`'s `Node(assoc)`, which gains a scope parameter; `follow` already records each returned association, so `rememberPlaceLink`'s `linked` dedupe loses no key), and at the end stores each scope's `Keys()` in `placeGraph.inputs map[string][]graphcache.InputKey` *(new)*. `placeNeighborhood` and `placeStep` keep their calls; `buildPlaceGraph`'s signature does not change. `func (pg *placeGraph) closure(seeds [][]byte) []graphcache.InputKey` *(new)* is the union of `pg.inputs[p]` for every place a BFS from `seeds` over `pg.partFrom` expands in `summaryChainDepth` rounds (rounds 0 to 7, as the build). `closure` writes no key of its own. A row adds `closure` of its places with `in.Add`. `rememberPlaceLink` stores every link whatever its dates, so `partFrom` holds every link the BFS needs. A place within 7 rounds of a row's seed is within 7 rounds of the batch's nearest seed, so the batch build expanded it | Every key comes from a scope call, so a read added to the chain builder is recorded without a second edit (review M-1) | Literal keys in `closure` — a hand-written copy of the build's reads; per-row place graphs — 4× slower |
| D-5 | `composePersons`: the batch phase keeps its `follow(EventsOfSubject)` and `g.Nodes(events)` prefetch, then calls `lifePicks(scopes[id], v, n)` per person, collects the birth and death pick event ids of all persons, and calls `eventPlaceGraph(g, v, pickIDs)` with its signature unchanged, as `PersonLifeOwners` already calls it. `onePerson` takes the picks instead of calling `lifePicks` again. In the per-row phase each person records `in.Follow(pick.event.ID, hops.PlacesOfEvent)` for every birth and death pick (not only the lead) and adds `pg.closure` of those places. Rows read the same: the lead's places are still seeds, at the same depth | A person depends on its birth and death events' places, as `Dependents` says, not on every event's (review M-3); one attribution pattern, D-3's, and `eventPlaceGraph` keeps its second caller untouched (review M-10) | Every subject event — makes S14-42 reproject a person for a marriage or census place write; scopes passed into `eventPlaceGraph` — a second way to attribute |
| D-6 | `composePersons`, `composeEvents` and `finishPlaces` return `(rows, map[string][]graphcache.InputKey, error)`, keyed by raw id. `persons`, `events` and `places` return `(rows, map[string][]graphcache.InputKey, error)`. `ListPersons`, `ListEvents`, `ListPlaces`, `PersonsByIDs`, `EventsByIDs`, `PlacesByIDs`, `ByIDs` and `display_register.go` `checkDisplay` discard the map. `Summaries` gains `inputs map[string][]graphcache.InputKey` and `func (h Summaries) Inputs(id []byte) []graphcache.InputKey` *(new)*. `RowsForSearch` sets `s := NewSummaries(ps, es, pl)` then `s.inputs` from the three maps. `NewSummaries` keeps its signature | One composition path; the reprojector is the index's only writer | A new `NewSummaries` argument — touches its other callers |
| D-7 | Migration `core/database/migrations/000062.sql` *(new)*: `CREATE TABLE catalog_search_inputs (dependent_id BLOB NOT NULL, input_id BLOB NOT NULL, bridge TEXT NOT NULL DEFAULT '', endpoint TEXT NOT NULL DEFAULT '', PRIMARY KEY (dependent_id, input_id, bridge, endpoint)) STRICT, WITHOUT ROWID;` and `CREATE INDEX catalog_search_inputs_input_idx ON catalog_search_inputs (input_id, bridge, endpoint);`. No foreign key. `catalogmodel`: `{Name: "catalog_search_inputs", Bucket: BucketSkip}` | Working rule: no CASCADE, or S14-38 returns. Named with its owner's prefix (review M-7). The primary key serves the per-dependent delete; the index serves S14-42's lookup | A foreign key with CASCADE, as `auto_reconciler_values` has |
| D-8 | `core/database/searchindex/inputs.go` *(new)* `writeInputs(q Querier, g *graphcache.Graph, ids [][]byte, heads map[string]handleHead, s conclusionsummaries.Summaries) error`: deletes the rows of every id in the batch as a dependent, then inserts `s.Inputs(id)` for a person / event / place head that is not merged, and the `g.AssociationInputs` entry for a head with `kind == ""` (one call per batch, over those heads). A gone or merged id keeps no rows. `reprojectHandleBatch` calls it after the documents. `ClearAll` adds `DELETE FROM catalog_search_inputs` | One package writes, clears and rebuilds the table, in the documents' transaction | A second writer in `autoreconciler` |
| D-9 | `func (g *Graph) AssociationInputs(ids [][]byte) (map[string][]InputKey, error)` *(new)* in `inputs.go`, keyed by raw id: it reads `g.Vocabulary()` once per call. For each id: a missing or merged node, or a type that is not a `connectrules` bridge (`TypeKeysByID` then `LookupBridge`), gives no entry. Otherwise, for each `Bridge.Endpoints` entry, the endpoint Property id (`Vocabulary.PropertyID(key, bridge origin)`) and `graphcache.Rank1(n.Kept(propID))`: its `EntityID`, when set, gives `{EntityID, bridge key, endpoint key}`. Two endpoints on the same Property key give two keys (distinct `ID`) | The bridge shape stays in `connectrules` and `graphcache`, as `catalog-queries.mdc` line 16 says (review M-2); rank 1 kept is the end every reader uses (`linkLoadSQL`, `endpointIDs`, `AssociationEndpoints`, `canonlinks.EdgeOf`; review M-9); the vocabulary once per call, not per association | Raw SQL on `auto_reconciler_values`; every kept rank — ends no `Link` shows |
| D-10 | `ProjectionVersion` becomes 11, with the comment line `11: summary inputs beside the handle documents (S14-40).` | An existing catalog fills the table on open, through `NeedsRebuild` → `RebuildAll` | A backfill migration — the projector is Go |
| D-11 | `ReprojectHandles` returns early on zero ids, then builds one graph through `var newGraph = graphcache.Over` *(new, package var)* and passes it to every batch: `reprojectHandleBatch(q, g, ids)`, `loadHandleSummaries(g, heads)`. `core/database/searchindex/export_test.go` *(new, package `searchindex`)* has `SetNewGraphForTest(fn func(graphcache.Querier) *graphcache.Graph) (restore func())`. T-4 and T-5 are in `inputs_test.go` *(new)* as package `searchindex_test`, because they build catalogs through `promote`, which an in-package test cannot import without a cycle | One graph per call is safe: a reprojection writes documents, not the value rows the graph caches. Its memory at a full `RebuildAll` is every handle's node, what the open catalog's own graph holds after its warm; the graph is dropped when the call returns (review X-8) | A timing test alone — machine-dependent |
| D-12 | `func (g *Graph) SetInputRecorderForTest(fn func(InputKey))` *(new)*, a per-graph func field beside the query counter. When set, the exported `Graph.Node` and `Graph.Follow` report `nodeKey` and `followKeys` (D-2), with no self rule. `Follow` reaches its node through an unexported lookup, so it reports no node values | The oracle needs the reads of a row composed alone, independent of the scope code it checks; per graph, as `SetQueryCounterForTest` is (review M-5) | A package-level hook — global state every graph checks |
| D-13 | The coverage check (AC-1.3) is a superset check. For X, the index's dependents are every `dependent_id` whose inputs hold `(X, values)` or any `(X, bridge, endpoint)`. They must include `Dependents([X])` minus X's immediate part_of parents (`hops.ParentsOfPlace`, one hop) and X's succession neighbours (`SuccessorsOfPlace`, `PredecessorsOfPlace`, one hop). It also excuses a missing person d when every event e that is in `Dependents([X])`, is a subject event of d (`EventsOfSubject`, role subject) and has kept `event_type` birth or death, has no accepted member (`graphcache.Node.Canonical` false). `birthOrDeath` ignores canonicality; `lifePicks` reads only canonical events, so no person row reads such an event's places (user, 2026-11-08, after T-10 found 9,987 such misses on the large fixture, whose `bulkCanon` events have no accepted claims). A miss with at least one canonical birth or death event in `Dependents([X])` is still a miss. Excused persons are logged as a count, apart from the extras. Extras are logged as a count per kind | S14-42 must invalidate at least what today's walk does | Equality — fails on a correct over-read |
| D-14 | `budgetComposeAll` *(new)* = 500 ms in `perf_test.go`, for composing every person and every event row once, warm, fastest of 3 | 220 ms before recording on the planning machine; 500 ms allows recording plus CI and catches a per-row place graph | No budget — a 4× regression would pass |

**Rules this plan bends.** None. STEP-8 amends `catalog-queries.mdc` line 16 to name the scope for composers; it does not bend it.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphcache.InputKey`, `InputScope`, `Graph.Inputs`, `NoInputs`, `InputScope.Add`, `Keys` *(new)* | type / func | `graphcache.NodeMemo` / `MemoDeps` (a hand-written inverse per slot) | The memo API stores results, not inputs; the scope is what `NodeMemo` owners use later |
| `Graph.AssociationInputs` *(new)* | func | `conclusionsummaries.AssociationEndpoints` (raw SQL, deleted by S14-42), `graphcache` `(*Graph).endpointIDs` and `linkLoadSQL` (load.go) | Reads the loaded node through `connectrules`, so no SQL copy of the bridge shape survives S14-42 |
| `Graph.SetInputRecorderForTest` *(new)* | test hook | `SetQueryCounterForTest` | Same per-graph shape, reporting keys instead of counting queries |
| `placeGraph.inputs`, `placeGraph.closure` *(new)* | field / func | `buildPlaceGraph` | `closure` only unions what the build's scopes recorded |
| `Summaries.inputs`, `Summaries.Inputs` *(new)* | field / func | `Summaries.Person` / `Event` / `Place` | Same struct, one more lookup |
| `catalog_search_inputs` *(new)*, `000062.sql` *(new)* | table | `catalog_search_docs` | Many rows per dependent; same owner and bucket |
| `searchindex/inputs.go` `writeInputs` *(new)* | file / func | `searchindex.Upsert`, `Delete` | Per-dependent replace of many rows |
| `searchindex.newGraph`, `SetNewGraphForTest` *(new)* | package var / test seam | `graphcache.Over` inline; `evidencegraphs` `SetMintRef` | The seam T-4 counts |
| `budgetComposeAll` *(new)* | const | `budgetLocateAtCity` | One more row |

**Builds on.** `graphcache` as the shared read layer; `searchindex`'s ownership of derived search data and its version-driven rebuild; the rebuild-equals-upkeep guard; `PersonLifeOwners`'s place graph from picks.
**Deletes.** The per-batch `graphcache.Over(q)` in `loadHandleSummaries`. `onePerson`'s own `lifePicks` call (the picks come from the batch phase). Nothing else becomes dead; S14-42 deletes the old inverse.
**Next adds.** S14-42 looks dependents up through `catalog_search_inputs_input_idx`. The kinship memos record through `InputScope`. A new handle kind's composer records through its scope (the `add-searchable-kind` line).

### Reach

None — the index is written and not read; `RowsForSearch` and every list return the same rows.

### Scale

At the map's Scale (a project of 3,000 Sources, 37,000 Subjects and 92,000 Observations; 6,000 events on one place), measured on the `TestLargeSourceBudgets` fixture:
- **Composition** gains one small key set per row (median 5 keys) and one key set per place in each batch's place graph. A person's place graph shrinks to its life events. `budgetComposeAll` holds it (D-14).
- **Writes.** Each reprojected document replaces its dependent's few rows: one `DELETE … WHERE dependent_id IN (batch)` and the inserts. A rebuild writes about 97,000 rows on the fixture, a few hundred thousand at the map's Scale.
- **Reprojection** of N handles builds one graph, not N/500: a place's links load once.

### Edges

- **Assumptions.** Composition is deterministic for a fixed catalog, so isolated and batch reads agree (T-2, T-9). `buildPlaceGraph` and `closure` both run `summaryChainDepth` rounds. A cached node stays valid for a whole `ReprojectHandles` call, because a reprojection writes no value rows.
- **Neighbors.** A filtered hop (`EventsOfSubject`, role subject; `ParentsOfPlace`, part_of) records the bridge-end key, which covers the links the filter drops (working rule). A missing node read through `Node` is still recorded (T-1), so its later creation reaches the row. A part_of link whose dates exclude it at the row's date is still in `partFrom` and its association's values are recorded (T-2 world). Succession is not followed by row composers; D-13 excludes those hops.
- **Removed inputs.** A dropped dependent loses its rows (AC-2.3). A deleted input keeps its dependents' rows until they are reprojected; that is what S14-42 reads.
- **Not handle inputs.** Vocabulary labels (`loadVocab`), the name format profile (`nameformats.LoadFor`, sort only), `Graph.Kind` (list membership), the alias index, and the row's own raw reads (`gallery.ForEntities`; the document's `sqlHandleHeads`, `sqlHandleValues`) are not recorded. The working rule "A row is its own input" covers the self-reads in S14-42.
- **Upkeep reach.** `RecomputeReported` skips reprojecting a dependent when no root's rows moved. If some write changes a dependent's inputs without moving a root's rows, the upkeep table differs from a rebuild while the documents still match. That would be an invalidation miss today, and T-7 would show it (Stop condition; S14-42 owns invalidation).
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — the keys come from catalog rows the app wrote; per row they are bounded by its links and the chain depth (8).

### Schema

Migration `core/database/migrations/000062.sql` *(new)*, per D-7, with a leading comment: `-- S14-40: catalog_search_inputs records the inputs of each person, event and place summary (and each association's ends) so a write can find the rows it moved. Owned by searchindex; no foreign key, so a deleted input keeps its dependents' rows until they are reprojected.`

Registered in `catalogmodel` as `BucketSkip`, with no FK rows. Delete class: none (derived). Audit: none.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `inputs.go` *(new)*: `InputKey`, `InputScope`, `Inputs`, `NoInputs`, `Add`, `Keys`, `AssociationInputs`, `SetInputRecorderForTest`. `graph.go` `Node` and `follow.go` `Follow` report to the recorder when set; `Follow` uses an unexported node lookup |
| `core/database/conclusionsummaries` | `compose.go`: per-row scopes (D-3), person picks first (D-5), returned key maps (D-6), `Summaries.inputs` / `Inputs`. `placechains.go`: `placeGraph.inputs`, `closure`; `buildPlaceGraph` and `rememberPlaceLink` read through per-place scopes (D-4) |
| `core/database/searchindex` | `inputs.go` *(new)* `writeInputs`; `handles.go`: `newGraph`, early return, one graph per call, `writeInputs` call, header comment; `searchindex.go`: `ClearAll`, `ProjectionVersion` 11 |
| `core/database/catalogmodel` | `catalog_search_inputs` row |

Errors: none new.

### FFI

None — no wire change.

### macOS

None — no app change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Budgets first**
Files: `core/database/promotealign/perf_test.go`.
Do: on today's code, add `budgetComposeAll` (D-14) and the subtest "compose every summary" (T-8). In "second part_of from the city", wrap the untimed first `f.PlaceRel` (the one that mints the association) in `since(func() { … })` and `t.Logf` its duration (T-11); the budgeted second part_of is unchanged.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` passes; copy the two logged figures into the PR description as the before figures.

**STEP-2 — The input scope**
Files: `core/database/graphcache/inputs.go` *(new)*, `graph.go`, `follow.go`, `core/database/graphcache/inputs_test.go` *(new)*.
Do: D-1, D-2, D-9, D-12.
Check: T-1 passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes.

**STEP-3 — Composers record per row**
Files: `core/database/conclusionsummaries/compose.go`, `placechains.go`, `display_register.go`, `core/database/conclusionsummaries/inputs_test.go` *(new)*.
Do: D-3 – D-6. Write T-2 and T-3 first, with the worlds of §8.
Check: T-2 and T-3 pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/...` passes with no existing test changed.

**STEP-4 — The table**
Files: `core/database/migrations/000062.sql` *(new)*, `core/database/catalogmodel/model.go`. Skills: `add-catalog-migration`, `add-catalog-model`.
Do: D-7.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `TestPragmaHonesty`).

**STEP-5 — The reprojector writes the index**
Files: `core/database/searchindex/handles.go`, `inputs.go` *(new)*, `searchindex.go`, `core/database/searchindex/export_test.go` *(new)*, `core/database/searchindex/inputs_test.go` *(new)*, `core/search/handles_test.go`. Skill: `add-catalog-query`.
Do: D-8, D-10, D-11.
Check: T-4, T-5 and T-6 pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/... ./core/search/...` passes.

**STEP-6 — Upkeep equals rebuild, for the index too**
Files: `core/database/autoreconciler/autoreconciler_test.go`.
Do: `assertUpkeepEqualsRebuild` also compares `inputRows(t, db)` with `inputRows(t, tx)` after `searchindex.RebuildAll(tx)`. `inputRows(t testing.TB, q autoreconciler.Querier) []string` *(new test helper)* returns every row of `catalog_search_inputs` as `hex(dependent)|hex(input)|bridge|endpoint`, sorted.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/` passes (T-7).

**STEP-7 — At scale**
Files: `core/database/promotealign/perf_test.go`.
Do: T-9 and T-10 as subtests of `TestLargeSourceBudgets`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` passes; every budget is under its limit; the T-11 line is under 0.8 × its STEP-1 figure.

**STEP-8 — Rules, skill and docs**
Files: `.cursor/rules/catalog-queries.mdc`, `.agents/skills/add-searchable-kind/SKILL.md`, `docs/deployment-plan/spike-14/summary-inputs.md`, `docs/ideas/structural-debt.md`.
Do:
- `catalog-queries.mdc` line 16 becomes: "An open-catalog read of the canonical graph follows a `hops.Hop` with `graphcache.Graph.Follow`; a summary composer reads other handles through its `graphcache.InputScope` (`Node`, `Follow`) so its inputs are recorded, and a raw query in a composer reads no other handle. Do not hand-write the association join in SQL; the bridge shape lives in the connectrules registry."
- The `add-searchable-kind` checklist gains "A handle kind's composer records its inputs through `graphcache.InputScope`."
- In `summary-inputs.md`, § "Measured at planning (2026-11-07)" gains the final key counts from T-9.
- In `structural-debt.md`, the "first `part_of` filed from a city" row gets the new figure from T-11.
Check: `rg -n 'through `Graph\.(Node|Follow)' docs/deployment-plan/spike-14 .agents .cursor` finds nothing; `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-40-build-the-input-index.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Landing docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row), this plan (status and Landed note).
Do: per the Definition of done.
Check: links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/inputs_test.go` `TestInputScope` *(new)* | table on a `graphcachetest` world (a person, a birth, a participation with role subject, a witness participation): `Node(other)` → `{other}`; `Node(self)` → none; `Node(missing id)` → `{missing}`; `Follow(person, EventsOfSubject)` → `{person, participation, person}` plus each returned association, not `{person}`; `Nodes` → none; `NoInputs()` → none; `Inputs(nil)` keeps self keys; `AssociationInputs([participation])` → the two end keys; a rank-2 kept end gives no key; `AssociationInputs([person])` → no entry; with the recorder set, `Graph.Follow` reports no node values | D-1, D-2, D-9, D-12 |
| T-2 | Go | `core/database/conclusionsummaries/inputs_test.go` `TestInputsMatchIsolatedComposition` *(new)* | worlds built with `evidencelayertest.Records` (`Participation`, `Locate`, `PlaceRel`) and the package's `placeSubject`, `named`, `promoteSubject`: Country / Region / City / Town with a birth at Town; a part_of with link dates that exclude it at the event's date; York succeeded_by Toronto; a witness participation; an event at two places; a person with a birth and a marriage at different places → for every person, event and place, `RowsForSearch(graphcache.Over(db), all)` `Inputs(id)` equals the keys a fresh `graphcache.Over(db)` with `SetInputRecorderForTest` reports for `RowsForSearch(g, Want{id})`, minus `{id}`; James's inputs hold no key of the marriage's place | AC-1.1, AC-1.2, D-4, D-5 |
| T-3 | Go | same file `TestInputsCoverDependents` *(new)* | the same worlds; D-13 for every handle → no missing id; extras logged per kind | AC-1.3, D-13 |
| T-4 | Go | `core/database/searchindex/inputs_test.go` `TestReprojectHandlesBuildsOneGraph` *(new)* | 1,200 random 16-byte ids (none in the catalog, 3 batches) → `newGraph` called once; zero ids → not called | AC-3.1, D-11 |
| T-5 | Go | same file `TestReprojectWritesInputs` *(new)* | a filed participation A between James and a birth (`evidencelayertest.Records.Participation`; both promoted with `promotetest.Save`, which files A): `searchindex.ClearAll`, then `ReprojectHandles([birth, A])` → the birth's rows hold `{James}` (the upkeep that already ran cannot have written them); A's rows are `(A → James, participation, person)` and `(A → birth, participation, event)`. Unpromote the birth (`promote.Unpromote` through `writes.Call`) → no row with the birth as dependent. Set a handle's `merged_into_id` with SQL and `ReprojectHandles` it → no row with it as dependent. `ClearAll` → table empty | AC-2.2, AC-2.3, D-8, D-9 |
| T-6 | Go | `core/search/handles_test.go` `TestRebuildIndexesSummaryInputs` *(new)* | as `TestRebuildIndexesHandles`: set version 10, `EnsureCatalog` → table not empty (any person row records at least its `EventsOfSubject` key, links or not), version 11 | AC-2.4, D-10 |
| T-7 | Go | `core/database/autoreconciler/autoreconciler_test.go` every `TestRebuildEqualsUpkeep_*` and `assertUpkeepEqualsRebuild` caller | upkeep table equals rebuild table | AC-2.1 |
| T-8 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "compose every summary" *(new)* | every person and every event id (by T-9's query) through `RowsForSearch` on `lf.c.Graph()`, warm, fastest of 3 → under `budgetComposeAll` | AC-3.3, D-14 |
| T-9 | Go | same, "inputs match isolated composition" *(new)* | ids by kind from `SELECT e.id FROM canonical_entities e JOIN subject_types st ON st.id = e.subject_type_id WHERE st.key = ? AND e.merged_into_id IS NULL ORDER BY e.id`; every person, event and place, on one `graphcache.Over(db)` graph with the recorder collecting into a fresh set per row (the probe ran all rows this way in 2.3 s) → T-2's equality; logs key count, median and max per row | AC-1.1 |
| T-10 | Go | same, "inputs cover Dependents" *(new)* | the first 100 ids of each kind by the T-9 query → T-3's check; logs extras per kind and the D-13 excused count | AC-1.3 |
| T-11 | Go | same, "second part_of from the city" (the STEP-1 log line) | the first, untimed `f.PlaceRel` (it mints the association and moves rows), one sample, logged; the budgeted second part_of stays as it is → the STEP-7 figure is under 0.8 × the STEP-1 figure (planning: 3.64 → 1.63 s); both in the PR description | AC-3.2 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/conclusionsummaries` and `persondetail` (signatures), `core/search` (document tests around reprojection), the `TestLargeSourceBudgets` budgets, `catalogmodel` honesty.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated — none
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New — no UI
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated) — not gated
- [ ] Domain doc updated if a rule in docs/*.md changed: none; catalog-queries.mdc, add-searchable-kind, summary-inputs.md and structural-debt.md per STEP-8
- [ ] PR description: what shipped, AC evidence, before and after timings (STEP-1, STEP-7), deviations
```

## 10. Manual check

None — the index is not read by anything a researcher can see. The checks are T-1 – T-11 and the timings in the PR description.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A row records less or more than composing it alone reads | T-2 and T-9 fail on any difference |
| Recording slows the lists | T-8 budget |
| The index drifts from documents after some write | T-7 on every upkeep scenario and seed |

**Stop and report if:**

- T-2 or T-9 finds a difference that is not a bug in this PR's scopes: a composer reads another handle outside `Node` / `Follow` (a raw query, a memo, `Kind`). Name the read; it needs a decision.
- T-3 or T-10 finds an id `Dependents` returns that the index lacks, outside D-13's exclusions (including the non-canonical birth or death excuse).
- T-7 fails: the upkeep table differs from a rebuild after a write whose documents match. Name the write and the dependent. It is an invalidation miss for S14-42, not something to fix here by widening reprojection.
- `budgetComposeAll` fails, or any existing `TestLargeSourceBudgets` budget fails after STEP-3 or STEP-5.
- The T-11 figure is not under 0.8 × its STEP-1 figure.
- A test fails because `RowsForSearch`, a list or a detail returned different rows than before.
- `catalogmodel.TestPragmaHonesty` or the schema hash requires a foreign key or a different bucket for `catalog_search_inputs`.

## 12. Open questions

None.

## 13. PR

**Title:** S14-40 — Build the input index
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (T-1 – T-11 output, the key counts from T-9); before and after timings; deviations from the plan; `Plan: docs/deployment-plan/spike-14/prs/S14-40-build-the-input-index.md`.
