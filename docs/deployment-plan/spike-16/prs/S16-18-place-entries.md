# S16-18 — Place entries

**Status:** Draft
**Spike / slice:** Spike 16 · Slice 2c — Walks
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no chrome change: an engine refactor; every place chain, place page and summary row reads the same as today)
**Board:** —
**Depends on:** S16-10 (landed [#789](https://github.com/mendahu/provenencia/pull/789)), S14-42 (Spike 14, landed [#748](https://github.com/mendahu/provenencia/pull/748))
**Blocks:** S16-19 (reads an event's own places from the entries), S16-21 (reads the filed `part_of` and `succeeded_by` links from the entries)
**Branch:** `s16-18-place-entries`
**Version bump:** none — no change a researcher can see, no wire and no catalog change ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`use-catalog-session`](../../../../.agents/skills/use-catalog-session/SKILL.md) (the graph cache, its registry and Verify)
**Rules that bind:** `.cursor/rules/go-tests.mdc` (table-driven tests), `.cursor/rules/catalog-queries.mdc` (a derived cache registers a typed handle from `init()`, reads only through `Get` / `GetMany` / `Peek`, adds its row to `docs/catalog-caches.md`, its fill to `BenchmarkCacheSizes`, and its package to `core/catalogsession/cache_doc_test.go`; this plan updates its composer line, see **Rules this plan bends**)
**Context:** [design § Results per key](../cache-architecture.md#results-per-key-for-the-walks), [design § A cache a composer reads hands its inputs to the row](../cache-architecture.md#a-cache-a-composer-reads-hands-its-inputs-to-the-row), [design § The write path fills only what it reads](../cache-architecture.md#the-write-path-fills-only-what-it-reads), [`docs/catalog-caches.md`](../../../catalog-caches.md), [Spike 14 summary inputs](../../archive/spike-14/summary-inputs.md), working rules: "The point of this batch is simpler code", "Cache entities and results per key; compose answers", "One cache registry, two ways to fill", "Every cache names the signal that updates it, and updates exactly", "Apply states its fan-out", "A write's one-shot graph fills only what it reads", "A cache a composer reads hands its inputs to the row", "Orders are defined and results do not move", "Memory is budgeted", "Batch 1's caps stay where S16-01 puts them", "No new persisted projections, and no stored answer without a measurement", "Measure growth"
**Planned:** 2026-11-13 against `main` @ `aa150de`
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

Every place chain ("Toronto, York County, Upper Canada") is built today by `buildPlaceGraph`, which loads each place's whole canonical node, with every link it has, then re-reads the `part_of` links and their associations, on every Places list read, every place page and every summary batch. A city's node carries its 6,000 events' location links, so any chain that passes through the city loads them. This PR keeps one **place entry** per place in the graph cache's registry: its shell (names and period), its filed `part_of` and `succeeded_by` links in both directions with their decoded dates, and the input keys a summary row records for it. Entries fill a frontier at a time by a narrow load that reads only `place_relationship` links. `buildPlaceGraph` becomes a walk over the entries, so the Places list, the place page's relationships and the event and person summary composers all read chains from them, and the node-reading helpers go.

**Why now.** Bar 13 ("Walks are lookups": a city's place chain reads without loading its 6,000 location links). S16-19 reads an event's own places from the entries, and S16-21 reads Promote's place links from them.

**What stays.** Every chain, fold, place page relationship and summary row reads as today, in today's order; every summary row records the same input keys (S14-40's tests hold it). `summaryChainDepth` (8), the input cap (`inputKeyCap`), S16-01's `eventLocations` cap of 20 places per event and the composers' own reads of an event's places stay. The Places list still loads each listed place's node for its row. No SQL table, wire or Swift change.

## 2. User stories

**US-1.** As a researcher, when a page shows a chain that passes through a city with 6,000 events, the engine does not load those events' location links to draw it, so the page and every write at that city stay fast.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Town `part_of` City `part_of` Region, promoted; three events located at City; the graph is dropped. `ParentsAtDate(town, today)` | `["City", "Region"]`; afterwards `g.Cached(city)` and `g.Cached(town)` are false, and no event is cached |
| AC-1.2 | The same catalog; an event at the town, composed alone on a write's one-shot graph (`RowsForSearch(graphcache.Over(db), Want{Events: [event]})`) | The event row's places read "Town" with parents `["City", "Region"]`; `Cached(city)` is false on that graph |
| AC-1.3 | City `part_of` Region and an event located at City (`graphcachetest.Canon`). `NodesOnBridge([city], "place_relationship")` | One node: its values; exactly its `part_of` link to Region (no location link); nothing is cached on the graph |

**US-2.** As the engine, a place entry is filled once and kept until a commit names that place, so repeat reads compose from memory and a commit elsewhere keeps it.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `ListPlaces` twice on the session graph (the second is warm), then a write that names a Person, then `ListPlaces` | `KeyCounts()["conclusionsummaries.places"]` is the same before and after the write, and the last `ListPlaces` issues no more queries than the warm one |
| AC-2.2 | Entries held for Town, City, Region; a toponym "Big City" is cited on City's Subject (a real write) | Right after the write City's entry is not held and Region's is; the next `ParentsAtDate(town)` starts with `PlacesByIDs(f.C, [city])[0].Names[0]`, and that row's `Names` contain "Big City" (rank order decides which name shows) |
| AC-2.3 | Entries held for Town and City, neither node cached; the Town → City `place_relationship` Subject is deleted (`subjects.Delete` through `writes.Call`) | Both entries are dropped (their ends come through `SummaryDeps`); `ParentsAtDate(town)` is `[]` |
| AC-2.4 | A new event is located at City (`lifeEvent` then `promoteSubject`) | City's entry is dropped; its refill reads no location link (AC-1.3's loader) |
| AC-2.5 | A vocabulary commit (`graphcachetest.NewCanon(t, f.C).CommitVocabulary()`) | Every place entry is dropped (`DerivedFrom` the vocabulary) |
| AC-2.6 | Verify on (`PROVENENCIA_GRAPH_VERIFY=1`) through the whole conclusionsummaries suite | Passes; a held entry corrupted by a test fails Verify with "conclusionsummaries: stored place entry" |
| AC-2.7 | The entry's Cap lowered to 2 (`SetPlaceEntryCapForTest`), a chain of five places read | The chain reads in full; at most 2 keys are held afterward |

**US-3.** As the engine, a summary row that reads a chain from the entries records the keys it recorded before, so S14-42 still finds every row a write moves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `TestInputsMatchIsolatedComposition`, `TestInputsHoldEveryShownParent`, `TestInputsCapCountsNothing`, `TestListsRecordNoInputs` | Unchanged and passing |
| AC-3.2 | An unknown 16-byte id read as a seed (`ParentsAtDate(f.C, unknown, today)`, then `PlaceEntryForTest(g, unknown)`) | `ParentsAtDate` is `[]`; the entry is held with `found` false and inputs `[{unknown, "", ""}]`, as today's `in.Node` of a missing node records. (A link never names a missing place: `linkLoadSQL` joins the neighbour entity, so only an explicit seed reaches this case) |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` exports `NodeKey` and `FollowKeys` (today `nodeKey`, `followKeys`) | STEP-2, D-9 |
| `graphcache.(*Node).Follow`, extracted from `Graph.follow` | STEP-2, D-9 |
| `graphcache.(*Graph).NodesOnBridge`: values and one bridge's links, uncached | STEP-2, D-2, AC-1.3 |
| `InputScope.Add` reports to the graph's test recorder | STEP-2, D-10, AC-3.1 |
| Keyed entry `conclusionsummaries.places`: fill, Apply, `DerivedFrom`, Check, Cap | STEP-3, D-1, D-3 – D-6, D-11, D-12, AC-2.1 – AC-2.7 |
| `buildPlaceGraph` walks the entries; `placeGraph` reads entries; `placeStep`, `finishPlaces`, `AttachPlaceRelationships` read them; `rememberShell`, `rememberPlaceLink` go | STEP-4, D-7, D-8, AC-1.1, AC-1.2, AC-3.1, AC-3.2 |
| Tests for the entries and the narrow load | STEP-5 |
| `docs/catalog-caches.md` row, `BenchmarkCacheSizes` fill, `cache_doc_test.go` import | STEP-6 |
| Measured before and after: `BenchmarkListPlaces` *(new)*, `BenchmarkPlaceMesh` *(new)* | STEP-1, STEP-7, D-14 |
| The map, design doc and `catalog-queries.mdc` say what landed | STEP-8 |

| Out | Home |
| --- | --- |
| The event and person composers reading an event's own places from the entries (`eventPlaces`' `in.Node`, `composeEvents`' `g.Nodes(placeIDs)`) | S16-19 |
| `PersonLifeOwners`, `lifeOwners`, `eventPlaceGraph` | S16-19 |
| Promote's connections reading place links from the entries (an exported reader) | S16-21 |
| The Places list's own row reads (`nodesOf` loads each listed place's node for its shell and lead picture) | not planned: a row reads its own node, as every summary row does |
| Storing a composed chain (`ParentChain` at a date) | not planned: rule "No new persisted projections, and no stored answer without a measurement" |
| Filling place entries at open | not planned: rule "Warm only what every proposal needs" |
| `ParentsAtDate`, `PartsAtDate`, `SuccessionNames` having only test callers | not planned: they are the tests' entry points to the chain rules |
| A time budget for chain reads | not planned: this PR's claim is a counter (AC-1.1), not a time (D-14) |

## 4. Current state

Verified at `aa150de`.

| Area | Where | Today |
| --- | --- | --- |
| Chain build | `core/database/conclusionsummaries/compose.go` `buildPlaceGraph` (line 1292) | Walks up `part_of` from seeds for `summaryChainDepth` rounds: `g.Nodes(frontier)` loads each place's whole node (every link, outcomes, members); per place a scope `rec.scope(g, nil)` records `in.Node(id)` and `follow(in, id, hops.ParentsOfPlace)`; `g.Nodes(assoc)` loads each link's association; `rememberPlaceLink` decodes its dates; `pg.inputs[id] = in.Keys()` |
| Place step | `compose.go` `placeStep` (line 1379) | Loads the place's node, follows the asked hops once, loads the neighbours' and associations' whole nodes, `rememberShell` each neighbour, `rememberPlaceLink` each link |
| Shell and link helpers | `compose.go` `rememberShell`, `rememberPlaceLink`, `assocDatesOf`, `placeShell` | `rememberShell` reads a node into `pg.byID` / `pg.period`; `rememberPlaceLink` orients a link (`FromEnd`), dedupes on from/to/association, reads its association's dates, appends to `partFrom`/`partTo` or `succFrom`/`succTo` |
| Places list | `compose.go` `finishPlaces` (line 1265) | On every read: `buildPlaceGraph` over every listed place, overwrites `pg.byID` / `pg.period` with each row, `ParentChain` at `chainQueryDate`, `rec.addClosure` per row |
| Composers | `compose.go` `composeEvents` (`buildPlaceGraph(g, v, placeIDs, rec)`, line 995), `composePersons` via `eventPlaceGraph` (line 607), `PersonLifeOwners` via `eventPlaceGraph` (line 820) | Read chains through `buildPlaceGraph`; `rec.addClosure` / `pg.addChains` hand each row the closure of its places |
| Chain view | `core/database/conclusionsummaries/placechains.go` `placeGraph`, `newPlaceGraph`, `reach`, `closure`, `closureSize`, `addChains`, `ImmediateParents`, `ImmediateParts`, `Successors`, `Predecessors`, `displayName`, `ParentChain`, `parentChain`, `isAncestor`, `FoldLocations`, `periodOf`, `linkHolds` | Read the maps `byID`, `partFrom`, `partTo`, `succFrom`, `succTo`, `period`, `inputs`; `partTo` is read only by `ImmediateParts` |
| Place page | `placechains.go` `AttachPlaceRelationships`; caller `api/ffi/handlers/conclusion_detail.go:235` | `placeStep` with all four place hops, then overwrites the page's own shell and period; reads `partFrom`, `ImmediateParts`, `Predecessors`, `Successors` |
| Test-only readers | `placechains.go` `placeNeighborhood`, `ParentsAtDate`, `PartsAtDate`, `SuccessionNames` | Called only from `placechains_test.go` |
| Input budget | `placechains.go` `inputBudget`, `scope`, `keep`, `addClosure`, `inputKeyCap` | Unchanged by this PR |
| Key constructors | `core/database/graphcache/inputs.go` `nodeKey`, `followKeys`; used in `inputs.go`, `graph.go:270`, `follow.go:15` | Unexported |
| Scope | `inputs.go` `InputScope.Node`, `Follow`, `Add` | `Node` and `Follow` record through `Add`; `Add` does not call the test recorder (`Graph.Node` and `Graph.Follow` do) |
| Follow | `core/database/graphcache/follow.go` `Graph.Follow`, `Graph.follow` | `follow` loads the node and filters `n.Links` by bridge, end and term |
| Node load | `core/database/graphcache/load.go` `loadIDs`, `loadChunk`, `entityRows`, `StoredValues`, `linkRaws`, `linkLoadSQL`, `bridgePredicates`, `finishLinks` | `loadChunk` reads entity rows, values, members, outcomes and every bridge's links, then installs the node; `linkLoadSQL` constrains to every registered bridge (`bridgePredicates`) |
| Commit | `core/database/graphcache/commit.go` `OnCommit` | Roots: the commit's handles, association ends (`detachAssociation`: loaded holders plus `endpointIDs` from SQL), merged neighbours; a removed node adds its links' neighbours only when it was cached. Then `applyEntries(roots, set)` |
| `SummaryDeps` | `core/database/searchindex/inputs.go` `ReprojectChanged`; `core/writes/run.go:88` | The rows the input index found plus the old and new ends of every association that moved or went |
| Registry | `core/database/graphcache/cache.go` `RegisterKeyed`, `Keyed.GetMany`, `Drop`, `SetCapForTest`; precedent `core/database/canonlinks/reader.go` `pivots` | Keyed entries store on a one-shot graph (`Over`) and go with it |
| Isolation oracle | `core/database/conclusionsummaries/conclusionsummariestest/inputs.go` `newIsolator`, `CheckMatchIsolated` | Composes each row alone on one `Over` graph and hears keys through `SetInputRecorderForTest`, which hears only `Graph.Node` and `Graph.Follow` |
| Display check | `core/database/conclusionsummaries/display_register.go` `checkDisplay` | Under Verify, composes fresh rows through `persons`, `events`, `places` |
| Cache doc | `docs/catalog-caches.md`; `core/catalogsession/cache_doc_test.go` (blank imports; `conclusionsummaries` is not imported); `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` | No place row |
| Tests | `placechains_test.go` (`TestPlaceParentsAtDateFromPeriods` … `TestUndatedPlacesAlwaysHold`, `TestAttachPlaceRelationships*`, `TestPersonLifeOwners`), `placechains_internal_test.go`, `inputs_test.go`, `display_cache_test.go`, `eventplaces_cap_test.go`, `walks_test.go`, `conclusionsummaries_test.go` `TestListPlacesQueryCountIsConstant`; `core/database/graphcache/inputs_test.go` `TestInputScope`; `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | Must stay green |

**Precedent.** S16-10 ([#789](https://github.com/mendahu/provenencia/pull/789)) registered `canonlinks.kinship` (`core/database/canonlinks/reader.go` `pivots`): a keyed entry from `init()` with `DerivedFrom: Vocabulary{}`, an Apply that drops roots, a `Check` that re-derives and `reflect.DeepEqual`s, `Cap: maxKeys` (65,536). Copy that registration shape. S16-13 ([#809](https://github.com/mendahu/provenencia/pull/809), `core/database/graphfill/records.go` `RecordsOf`) fills a keyed entry through `GetMany` with a batched fill closure; copy its fill-closure shape. Do not copy `pivots`' Apply (`deps` walks cached neighbours, which the place entry does not need: D-5).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A keyed registry entry `conclusionsummaries.places` *(new)*, handle `placeEntries graphcache.Keyed[placeEntry]` *(new)* in `core/database/conclusionsummaries/placeentries.go` *(new)*, keyed by place id. `placeEntry` *(new)*: `found bool`; `shell PlaceSummary` (from `placeShell`); `period valuecompare.Window`; `partFrom`, `partTo`, `succFrom`, `succTo []placeLink` (the place as part, as whole, as predecessor, as successor), each oriented as `rememberPlaceLink` does today and in the node's link order; `inputs []graphcache.InputKey` (D-4). Values are shared and read-only. Each list is deduped on from/to/association as `rememberPlaceLink` does today; a self-loop `part_of` sits in both `partFrom` and `partTo`, as today | `conclusionsummaries` owns `placeGraph`, `placeShell` and the chain rules; per-key entries are the map's rule "results per key" | An entry in `graphcache` (it would import place semantics); a whole entry (a write's graph would build every place) |
| D-2 | Fill by a narrow load: `graphcache.(*Graph).NodesOnBridge(ids [][]byte, bridge string) (map[string]*Node, error)` *(new)* reads `entityRows`, `StoredValues` and, when `bridge != ""`, that bridge's links (`linkRaws` through `linkLoadSQL` restricted to the bridge, then `finishLinks`), in `sqlbatch.Each` chunks, and caches nothing: the returned nodes have empty `Members` and `Outcomes`, and `Links` holds only that bridge's links. A missing id is absent. The entry always fills through it, even when the place's node is cached | One fill path; the narrow load is what skips location links on a write's graph and on a cold session graph | Reading the cached node when held (two fill paths, and the result would depend on what else loaded); installing a partial node (other readers would see a node without its links) |
| D-3 | A link's dates come from `NodesOnBridge(assocIDs, "")` (values only) and `assocDatesOf`. One fill per frontier: places' narrow nodes, then all their links' associations in one call. An association absent from that result is an error, `"conclusionsummaries: association %x is missing"`, as `rememberPlaceLink` returns today | Same loader; `assocDatesOf` and `keptDateAt` read a `*Node` today | `StoredValues` wrapped in a hand-made `Node` in `conclusionsummaries` |
| D-4 | `placeEntry.inputs` = `NodeKey(id)` plus `FollowKeys(id, hops.ParentsOfPlace, partOfLinks)`, where `partOfLinks` is `n.Follow(hops.ParentsOfPlace)`; sorted with `SortInputKeys`. A missing place's inputs are `[NodeKey(id)]`. No keys for `PartsOfPlace` or succession | These are exactly the keys today's per-place scope records (`in.Node(id)`, `follow(in, id, ParentsOfPlace)`, `in.Node(assoc)` = `NodeKey(assoc)`, which `FollowKeys` already holds), so every row's closure is unchanged (AC-3.1, AC-3.2) | Every hop's keys: a hub's 20,000 part links would enter every row whose chain passes it |
| D-5 | Apply: `k.Drop(g, roots)` then `k.Drop(g, set.SummaryDeps)`. The next read refills a dropped place by the narrow load | User, 2026-11-13: telling a place-relationship commit from a location commit costs a query per root, the same as a refill. `SummaryDeps` carries the old ends of a removed link, which `roots` misses when neither end's node is cached (`OnCommit` adds a removed node's neighbours only when it was cached) | Exact drop on shell or place-link moves only (a bridge check per root, more code, same cost); `roots` alone (misses AC-2.3) |
| D-6 | `DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}}`; `Cap: maxPlaceEntries` *(new const, 65,536)*; keyed entries are graph-lifetime, so `Drop` and a `Structure` commit empty it | The shell and link terms decode through vocabulary ids; the cap matches `canonlinks.maxKeys` and holds every place of a real catalog | No cap (a crafted catalog could grow it without bound) |
| D-7 | `buildPlaceGraph(g, v, seeds, rec)` keeps its signature and walks the entries: per round, `placeEntries.GetMany(g, frontier, fill)`; each visited place's entry goes into `pg.places`; the next frontier is each entry's `partFrom` wholes not yet seen; `summaryChainDepth` rounds, as today. `placeGraph` becomes `{places map[string]placeEntry}`; every method reads it (`byID[id]` → `places[id].shell`, `partFrom[id]` → `places[id].partFrom`, and so on; `periodOf` of an absent place stays `Always`). It records no scope: `closure` reads `places[id].inputs`. Where a method tested `byID`'s `ok` (`displayName`, `relationship`, `membershipDates`' callers), it tests that the place is in `places` and `found`: a `found: false` entry reads as absent (`relationship` returns `Entity{ID: id}`, `displayName` returns `""`). `rememberShell`, `rememberPlaceLink`, the `linked` maps and the per-place scopes go | User, 2026-11-13: one chain builder, so the Places list, the place page and the event and person composers all read entries in this PR and no node-walking builder survives beside it | A new walker for the Places list beside `buildPlaceGraph` until S16-19 (the map's first split: two builders between the PRs) |
| D-8 | `placeStep(c, id, hopList...)`: `GetMany([id])`, then `GetMany` of every neighbour on the asked hops (from the entry's four lists, by hop: `ParentsOfPlace` → `partFrom` wholes, `PartsOfPlace` → `partTo` parts, `SuccessorsOfPlace` → `succFrom` successors, `PredecessorsOfPlace` → `succTo` predecessors); `pg.places` holds the place's entry with only the asked hops' lists kept, and the neighbours' entries with their lists cleared, so a reader sees what `placeStep` reads today. `finishPlaces` and `AttachPlaceRelationships` stop overwriting `byID` and `period` with the row: the entry's shell is the same `placeShell` of the same values | Same results as today's step (a parts question does not climb) | Handing readers whole neighbour entries (a neighbour's own links would join the step) |
| D-9 | `graphcache` renames `nodeKey` → `NodeKey` and `followKeys` → `FollowKeys` (exported, doc unchanged) and extracts `func (n *Node) Follow(hop hops.Hop) ([]Link, error)` *(new)* from `Graph.follow`, which now loads the node and calls it | The map's rule ("S16-18 exports `nodeKey` and `followKeys`"); the entry filters a narrow node's links with the same rule `Graph.Follow` uses | Re-writing the bridge, end and term filter in `conclusionsummaries` |
| D-10 | `InputScope.Add` hands each key to the graph's test recorder (when set) before its `off` and self checks; `InputScope.Node` and `Follow` record through a new unexported `add`, since `Graph.Node` and `Graph.Follow` already report. `SetInputRecorderForTest`'s doc: "hears every key `Node` and `Follow` read and every key handed to `InputScope.Add`" | The isolation oracle (`CheckMatchIsolated`) hears reads only through the recorder; a chain read from an entry reaches a row through `addClosure`'s `in.Add`, so the oracle hears the same keys the batch records (AC-3.1) | Leaving the recorder alone (the oracle would miss every chain key); a second `Add`-like method for entry keys |
| D-11 | A missing place (no entity row): the fill stores `placeEntry{found: false, inputs: [NodeKey(id)]}`. A merged place: built from what the narrow load returns, as today's node read does | Today's `in.Node` records a missing node's key, and `buildPlaceGraph` shows a merged node's shell; one entry shape | Not storing missing places (every read would query again) |
| D-12 | Check: `v` from `vocabulary(g)`, then rebuild the one id with the entry's fill (`fillPlaceEntries(g, v, [][]byte{id})`) and `reflect.DeepEqual`; mismatch error `"conclusionsummaries: stored place entry %x does not match a fresh load"` | `sources.records` and `canonlinks.kinship` check by rebuild; the narrow load reads the committed rows | A Check that reads cached nodes (the entry is not built from them) |
| D-13 | `checkDisplay` composes through `places`, `events`, `persons`, which now `GetMany` place entries; under Verify that may fill or touch entries. Accepted | Verify runs only under `PROVENENCIA_GRAPH_VERIFY`; the entries it fills are exact | A side door that composes without the registry |
| D-14 | No time budget. STEP-1 and STEP-7 record `BenchmarkListPlaces` *(new)* at 4,000 and 16,000 persons (cold: `Drop`, then `ListPlaces`; warm: a second `ListPlaces`; write: `RowsForSearch` of every place on `graphcache.Over(db)`) and `BenchmarkPlaceMesh` *(new)* (D-15's crafted mesh) on the same machine | The PR claims a counter (AC-1.1); rule "Measure growth" applies to speed claims, and the figures still show the cost of the always-narrow fill (D-2) | A new constant in `perf_test.go` |
| D-15 | `BenchmarkPlaceMesh`: through `graphcachetest.Canon` in one `Bulk` (places are `Canon.Entity("place")`, links `Canon.Link("place_relationship", {"from": part, "to": whole}, "place_relationship_type", "part_of")`, locations `Canon.Location`), a hub place with 20,000 parts (bar 2's 20,000-child hub), the hub the lowest place of an eight-place `part_of` chain (the hub `part_of` W1, W1 `part_of` W2, and so on to W7: seven wholes above it), and one event located at the hub per 10 parts (2,000 location links). Sub-benchmarks: `ListPlaces` cold (`Drop` first) and warm; `AttachPlaceRelationships` of the hub's row, cold; and `ListPlaces` after `g.OnCommit` of a commit whose `Handles` is the hub and whose `SummaryDeps` is the hub and its 20,000 parts (what citing a new toponym on the hub reports, since each part's row reads the hub in its chain) | The map's Check (Apply's work stated for a crafted `part_of` mesh) and rule "Apply states its fan-out"; `Canon` has no toponym writer, so the rename's effect is committed directly | A crafted mesh in the large fixture (it has no `part_of` links) |

**Rules this plan bends.**

- `.cursor/rules/catalog-queries.mdc`: "a summary composer reads other handles through its `graphcache.InputScope` (`Node`, `Follow`) so its inputs are recorded". The composers now read chains from a registry entry and record its keys through `InputScope.Add`. The map's working rule "A cache a composer reads hands its inputs to the row" (user, map review K-21, 2026-11-08) asked for exactly this. STEP-8 rewrites the line: "a summary composer reads other handles through its `graphcache.InputScope` (`Node`, `Follow`), or reads a registry entry that recorded its keys with `graphcache.NodeKey` and `FollowKeys` and adds them to the scope, so its inputs are recorded".
- The map's S16-18 / S16-19 split and S16-18's Apply wording ("Apply refreshes a place only when its shell or place-relationship links move"): changed by the user on 2026-11-13 (D-5, D-7). The map row, scope entries and design doc are updated in the planning commit.

## 6. Design

### Codebase fit

Searched: `RegisterKeyed`, `GetMany(`, `nodeKey`, `followKeys`, `func (g \*Graph) follow`, `StoredValues(`, `linkLoadSQL`, `bridgePredicates`, `loadChunk`, `placeGraph`, `buildPlaceGraph`, `SetInputRecorderForTest`, `Canon) Link`, `Benchmark` in `core/database/conclusionsummaries` and `core/database/promotealign`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/conclusionsummaries/placeentries.go` *(new)* | file | `display_register.go` (the package's other registration) | That file registers the display check; the entry has its own type, fill, Apply and Check |
| `placeEntry` *(new)* | type | `placeGraph` (`placechains.go`) | `placeGraph` stays as the walk's view, now a map of entries (D-7) |
| `placeEntries` *(new)*, `maxPlaceEntries` *(new)* | keyed handle, const | `canonlinks` `pivots`, `maxKeys` | A different value type and owner; the registry is the shared part |
| `fillPlaceEntries(g *graphcache.Graph, v vocab, ids [][]byte) (map[string]placeEntry, error)` *(new)* | func | `rememberShell` + `rememberPlaceLink` + `assocDatesOf` | It replaces the two `remember*` helpers (deleted); `assocDatesOf`, `placeShell`, `periodWindow` are reused |
| `placeEntriesOf(g *graphcache.Graph, v vocab, ids [][]byte) (map[string]placeEntry, error)` *(new)* | func | `graphfill.RecordsOf` | The one read: `placeEntries.GetMany` with `fillPlaceEntries`; `buildPlaceGraph`, `placeStep` and the Check call it |
| `graphcache.(*Graph).NodesOnBridge` *(new)* | func | `loadIDs` / `loadChunk` (`load.go`) | Those install full nodes with every bridge's links, members and outcomes; this returns uncached nodes with one bridge's links. It reuses `entityRows`, `StoredValues`, `linkRaws`, `finishLinks` |
| `linkRaws(ids [][]byte, bridge string)`, `linkLoadSQL(valueTable, entityTable string, ids [][]byte, bridge string)`, `bridgePredicates(bridge string)` (new parameter on each) | signature | themselves | `""` keeps today's every-bridge query; `loadLinks`, `mergedNeighbors` and `LinkLoadSQL(ids)` pass `""` |
| `graphcache.(*Node).Follow` *(new)* | method | `Graph.follow` | Extracted; `Graph.follow` calls it |
| `graphcache.NodeKey`, `graphcache.FollowKeys` (renamed) | func | `nodeKey`, `followKeys` | Exported, per the map |
| `InputScope.add` *(new, unexported)* | method | `InputScope.Add` | `Add` now reports to the recorder (D-10); `Node` and `Follow` record without reporting twice |
| `BenchmarkListPlaces` *(new, test)* in `core/database/promotealign/bench_test.go` | benchmark | `BenchmarkCacheSizes` | The large fixture (`newLargeFixture`) lives there; `BenchmarkCacheSizes` measures heap, not read time |
| `BenchmarkPlaceMesh` *(new, test)* in `core/database/conclusionsummaries/placeentries_test.go` *(new)* | benchmark | `eventplaces_cap_test.go` `newPlacesWorld` (a crafted world through `graphcachetest.Canon`) | A different shape (a hub and a chain); the `Canon` calls are the reuse |
| `TestPlaceEntries` *(new, test)*, `TestChainsSkipLocationLinks` *(new, test)* in `placeentries_test.go` | test | `TestForFollowsCommits` (`canonlinks`) | — |
| `HeldPlaceEntryForTest`, `PlaceEntryForTest` (returns `found` and `inputs` through `Peek`), `SetPlaceEntryCapForTest`, `CorruptPlaceEntryForTest` *(new, test)* in `core/database/conclusionsummaries/export_test.go` | helper | `SetInputKeyCapForTest` (same file) | The entry handle is unexported; these wrap `placeEntries.Peek`, `SetCapForTest` and `Set` (a copy of the held entry with `shell.Names` set to `["corrupt"]`) for the external test package |
| `placeRel` returns the bridge `subjects.Subject` (test helper) | helper | itself | Callers ignore the result; T-4 deletes the bridge |
| `TestNodesOnBridge` *(new, test)* in `core/database/graphcache/graph_test.go` | test | `TestLoadLinksUsesTheEdgeIndex` | That test checks the query plan, not what the load returns |

**Builds on.** The registry's keyed entries and `GetMany` batch fill (S16-10); `graphcache`'s link query and `connectrules` bridges (no hand-written association join); S14-40's `InputKey`, `inputBudget` and closure; `graphcachetest.Canon` for crafted catalogs.
**Deletes.** `rememberShell`; `rememberPlaceLink`; `placeGraph`'s `byID`, `partFrom`, `partTo`, `succFrom`, `succTo`, `period` and `inputs` maps (replaced by `places`); `buildPlaceGraph`'s `scopeOf` scopes, `linked` map, `g.Nodes(frontier)` and `g.Nodes(assoc)`; `placeStep`'s node, neighbour and association loads, its `pending` type and `linked` map; `finishPlaces`' `pg.byID[id] = summaries[i]` and `pg.period[id] = …`; `AttachPlaceRelationships`' `g.byID[...] = *h` and `g.period[...] = …`; the unexported names `nodeKey` and `followKeys`. The place-chain node walk is replaced by the entries for every caller.
**Next adds.** S16-19 makes `eventPlaces` read an event's own places' shells from `placeEntriesOf` and drops `composeEvents`' `g.Nodes(placeIDs)`, then deletes `eventPlaceGraph` with `PersonLifeOwners`. S16-21 exports a reader of an entry's `part_of` and `succeeded_by` links for `graphalign.Connections`. A later entry that a composer reads records its keys the same way (`NodeKey`, `FollowKeys`, `InputScope.Add`).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A chain reader sees an entry filled before an unrelated commit | `placeGraph` methods | Apply drops every place a commit names (D-5); AC-2.1 – AC-2.5 |
| A place id with no entity row reaches the walk (a dangling link end) | `fillPlaceEntries` | Stored `found: false` with `[NodeKey(id)]` (D-11); AC-3.2; `displayName` is `""` as today |
| A neighbour on the place page whose entry the LRU evicted between `GetMany` calls | `placeStep` | Each `GetMany` returns every asked id even when a later store evicts it (`cache.go`); the step keeps its own map |
| A one-shot write graph reads chains | `buildPlaceGraph` on `graphcache.Over` | Keyed entries store on that graph and go with it (`Keyed.store` has no `alt` check); AC-1.2 |

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; one city with 6,000 events; 51 places and no `part_of` links on that fixture) and bar 2's crafted 20,000-child hub:

- **Fill.** One frontier round issues, per `sqlbatch` chunk, four queries for the places (`entityRows`, `StoredValues`, the place-relationship link query, `finishLinks`' disambiguation) and two for their links' associations (`entityRows`, `StoredValues`); at most `summaryChainDepth` (8) rounds per read. On the large fixture every chain is one round (no `part_of` links), so the first `ListPlaces` after open issues six more queries than today, once; STEP-7 records it. The city's fill reads its values and no location row.
- **Session graph after open.** `promotealign.Warm` has loaded every place node, so today's walk reads them from memory; the entries query once per place and then read from memory (D-2). `BenchmarkListPlaces` cold records the difference at both sizes (§11 stop).
- **Apply per commit.** Two `Keyed.Drop` calls: one map delete attempt per root and per `SummaryDeps` id. Filing a location at the 6,000-event city: roots are the location, the event and the city; `SummaryDeps` holds the event and the city (`budgetLocateAtCity`'s case); the city's entry drops, refilled by one narrow round. Renaming the city: `SummaryDeps` holds every row that reads the city's name (its 6,000 events and their persons' life rows, about 12,000 ids): about 12,000 map lookups (well under a millisecond); the only place key among them is the city's. Renaming the crafted hub: `SummaryDeps` holds its 20,000 parts' rows (each part's row reads the hub in its chain), so all 20,000 part entries drop and the next `ListPlaces` refills them in `ceil(20,000 / chunk)` rounds of six queries; `BenchmarkPlaceMesh` "after the hub commit" records it.
- **Hub entry.** The hub's entry holds its 20,000 `partTo` links with their decoded dates; its fill reads 20,000 association value sets. Today a chain through the hub loads the hub's whole node (the same 20,000 links plus 2,000 location links). `BenchmarkPlaceMesh` records both.
- **Memory.** One entry per place reached: on the large fixture 51 keys; STEP-6 records `placeEntries-B` and the key count in `BenchmarkCacheSizes` and `docs/catalog-caches.md`. Past `maxPlaceEntries` the least recently used place leaves (AC-2.7).

### Edges

- **Assumptions.** `parentChain` and `isAncestor` assume `pg` holds shells only for places the walk expanded within `summaryChainDepth` ("a name the build did not read must not appear because another row of the batch happened to read it"): `pg.places` holds only the entries the walk visited, in the same rounds (`TestInputsHoldEveryShownParent`, `TestPlaceParentsAtDateFromPeriods`). `reach` and `closure` assume `partFrom` holds the part-end `part_of` links the build followed: the entry's `partFrom` is `n.Follow(hops.ParentsOfPlace)`, the same filter. `ImmediateParents` and the other lists sort by `displayName` with `sort.SliceStable`, so ties keep link order: entries keep the node's link order (`finishLinks` sorts by association ref then neighbour ref, the narrow load included). Cycles: the walk's `seen` map is unchanged (`TestUndatedPlacesAlwaysHold`, the walks tests).
- **Neighbors.** `partTo` used to hold only the batch's children of a visited whole; it now holds every part. Only `ImmediateParts` reads it, and only through `placeStep` (place page, `PartsAtDate`), where today's step already followed `PartsOfPlace` in full (`TestAttachPlaceRelationships`, `TestIrelandUKDatedMembershipDropsAfter1922`). Succession never builds a chain (`TestSuccessionNeverBuildsChain`). A place-page step reads neighbour shells without their links (D-8).
- **Removed inputs.** A removed `part_of` link: its association's old ends come through `SummaryDeps`, both entries drop (AC-2.3). A deleted place: in `roots`, its entry drops; its links' associations are committed by its writer, so their ends drop too. A merged place: `mergedNeighbors` puts its neighbours in `roots`. A changed link date or term: the association is in `Handles`, its ends are roots. A vocabulary change: `DerivedFrom` drops every entry (AC-2.5).
- **Real data.** None — no rule, seed or table changes.
- **Untrusted size.** A hub's parts (20,000 in bar 2): loaded in `sqlbatch` chunks, measured (D-15); never copied into a row's keys (D-4). A `part_of` mesh's frontier: bounded by `summaryChainDepth` rounds and the walk's `seen` map, as today; a row's recorded keys stay under `inputKeyCap`. Entries held: `maxPlaceEntries`.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` `inputs.go` | `NodeKey`, `FollowKeys` (D-9); `Add` reports to `g.inputRecorder`, then records; `add` *(new)* records only; `Node` and `Follow` call `add`; `SetInputRecorderForTest` doc per D-10 |
| `core/database/graphcache` `follow.go` | `func (n *Node) Follow(hop hops.Hop) ([]Link, error)` *(new)*: "the node's links that match hop, in the node's order"; `Graph.follow` loads the node and returns `n.Follow(hop)`; `Graph.Follow` reports `FollowKeys` |
| `core/database/graphcache` `graph.go` | `Node` reports `NodeKey(id)` |
| `core/database/graphcache` `load.go` | `func (g *Graph) NodesOnBridge(ids [][]byte, bridge string) (map[string]*Node, error)` *(new)*, doc: "reads ids' entity rows and values, and their links on bridge only (none when bridge is ""), without caching them: Members and Outcomes are empty. A missing id is absent." `linkRaws`, `linkLoadSQL` and `bridgePredicates` take `bridge`, and their existing callers pass `""` (D-2) |
| `core/database/conclusionsummaries` `placeentries.go` *(new)* | `placeEntry`, `placeEntries`, `maxPlaceEntries`, `init` registration (D-1, D-5, D-6, D-12), `placeEntriesOf`, `fillPlaceEntries`, `checkPlaceEntry` |
| `core/database/conclusionsummaries` `compose.go` | `buildPlaceGraph` walks entries (D-7); `placeStep` per D-8; `finishPlaces` loses its overwrite; `rememberShell`, `rememberPlaceLink` deleted; `assocDatesOf`, `placeShell` unchanged |
| `core/database/conclusionsummaries` `placechains.go` | `placeGraph{places map[string]placeEntry}`; `newPlaceGraph`; every method reads `places` (D-7); `AttachPlaceRelationships` loses its overwrite (D-8) |

Sketch:

```go
type placeEntry struct {
	found    bool
	shell    PlaceSummary
	period   valuecompare.Window
	partFrom []placeLink // this place is the part
	partTo   []placeLink // this place is the whole
	succFrom []placeLink // this place is the predecessor
	succTo   []placeLink // this place is the successor
	inputs   []graphcache.InputKey
}

placeEntries = graphcache.RegisterKeyed(graphcache.KeyedEntry[placeEntry]{
	Name:        "conclusionsummaries.places",
	DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}},
	Apply: func(g *graphcache.Graph, k graphcache.Keyed[placeEntry], roots [][]byte, set effects.Set) error {
		k.Drop(g, roots)
		k.Drop(g, set.SummaryDeps)
		return nil
	},
	Check: checkPlaceEntry,
	Cap:   maxPlaceEntries,
})
```

Errors: none new besides the Check message (D-12).

### FFI

None — the Place detail and list wire are unchanged.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base**
Files: `core/database/promotealign/bench_test.go`, `core/database/conclusionsummaries/placeentries_test.go` *(new)*.
Do: add `BenchmarkListPlaces` (D-14) and `BenchmarkPlaceMesh` (D-15), the latter in `package conclusionsummaries_test`; both call only exported functions that exist at the base (`ListPlaces`, `AttachPlaceRelationships`, `RowsForSearch`, `graphcache.Over`). On the planning base (no other change) run both and `BenchmarkCacheSizes`; write the figures into the ship log with `hostname`.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkListPlaces|BenchmarkCacheSizes' -benchtime 1x -timeout 60m ./core/database/promotealign` prints both sizes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkPlaceMesh -benchtime 1x ./core/database/conclusionsummaries` prints its four sub-benchmarks.

**STEP-2 — The graph cache exports its keys, a node's Follow and a narrow load**
Files: `core/database/graphcache/inputs.go`, `follow.go`, `graph.go`, `load.go`, `graph_test.go`, `inputs_test.go`.
Do: D-9, D-10, D-2 per §6. Add `TestNodesOnBridge` (T-1) and the `TestInputScope` subtest "the recorder hears Add" (T-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes; `rg -n "\bnodeKey\b|\bfollowKeys\b" core` finds nothing.

**STEP-3 — The place entry**
Files: `core/database/conclusionsummaries/placeentries.go` *(new)*. Skill: `use-catalog-session`.
Do: D-1, D-3 – D-6, D-11, D-12. `fillPlaceEntries`: `NodesOnBridge(ids, "place_relationship")`; for each present id, `placeShell`, `periodWindow`, the four lists from `n.Follow` of the four place hops, each link oriented and dated with `assocDatesOf` over one `NodesOnBridge(assocIDs, "")`; `inputs` per D-4; a missing id per D-11. Nothing reads the entry yet.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/...` passes.

**STEP-4 — Every chain reads the entries**
Files: `core/database/conclusionsummaries/compose.go`, `placechains.go`, `placechains_internal_test.go` (only if a test builds `placeGraph` maps directly).
Do: D-7, D-8 per §6; delete what **Deletes** lists in these files. `closure`, `closureSize`, `reach`, `addChains`, `inputBudget` keep their behavior.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/... ./api/...` passes, and `PROVENENCIA_GRAPH_VERIFY=1` the same command passes; `rg -n "rememberShell|rememberPlaceLink|pg\.byID|\.partFrom\[|\.succTo\[" core/database/conclusionsummaries` finds only `places[...]` reads.

**STEP-5 — Tests for the entries**
Files: `core/database/conclusionsummaries/placeentries_test.go` *(new)*, created in STEP-1, `export_test.go`, `placechains_test.go` (`placeRel` returns the bridge Subject).
Do: add the four `export_test.go` hooks (§6 Codebase fit), `TestChainsSkipLocationLinks` (T-3) and `TestPlaceEntries` (T-4). `placeentries_test.go` is the external package `conclusionsummaries_test` (as `eventplaces_cap_test.go`): `graphcachetest` imports `conclusionsummaries`, so an internal test importing it would cycle.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestChainsSkipLocationLinks|TestPlaceEntries' ./core/database/conclusionsummaries` passes, and again with `PROVENENCIA_GRAPH_VERIFY=1`.

**STEP-6 — The cache table and the sizes benchmark know the entry**
Files: `docs/catalog-caches.md`, `core/catalogsession/cache_doc_test.go`, `core/database/promotealign/bench_test.go`.
Do: blank-import `core/database/conclusionsummaries` in `cache_doc_test.go`; in `BenchmarkCacheSizes` add a `measure("placeEntries", fill)` whose fill runs `conclusionsummaries.ListPlaces(c)`, placed before `measure("displays", …)`; add the `conclusionsummaries.places` row (Owner `core/database/conclusionsummaries/placeentries.go` *(new)* · `placeEntries`; Key place handle id; Fill `placeEntriesOf`: a narrow load of the places' values and `place_relationship` links and their associations' values, a frontier at a time, through `Keyed.GetMany`; Updated by `DerivedFrom` the vocabulary, Apply drops the commit's roots and `effects.Set.SummaryDeps`; Lifetime graph; Cap 65,536; size from STEP-7).
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestCacheDocListsEveryEntry|TestCacheDocSeesEveryOwner' ./core/catalogsession` passes.

**STEP-7 — Measure after**
Files: `docs/catalog-caches.md` (size cell).
Do: run the STEP-1 commands on the same machine; write the after figures beside the base in the ship log; fill the size cell with `placeEntries-B` and `conclusionsummaries.places-keys` at 4,000 / 16,000.
Check: §11's measured stop conditions hold; `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets ./core/database/promotealign` passes.

**STEP-8 — Docs**
Files: `.cursor/rules/catalog-queries.mdc`, `docs/deployment-plan/spike-16/cache-architecture.md`, `docs/deployment-plan/spike-16/deployment-plan.md`, `docs/deployment-plan/spike-16/completed.md`, this plan.
Do: the `catalog-queries.mdc` line per **Rules this plan bends**; `cache-architecture.md` § A cache a composer reads hands its inputs to the row: "(exported by S16-18, today the unexported `nodeKey` and `followKeys`)" → "(`graphcache.NodeKey` and `FollowKeys`)"; the map's working rule "A cache a composer reads hands its inputs to the row": "(S16-18 exports `nodeKey` and `followKeys`)" → "(`graphcache.NodeKey` and `FollowKeys`, S16-18)"; map row status `Landed #N`; `completed.md` paragraph; this plan's status and Landed note.
Check: the DoD sweep finds nothing; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/graph_test.go` `TestNodesOnBridge` *(new)* | `Canon`: city `part_of` region, an event located at the city. Table: `NodesOnBridge([city], "place_relationship")` → one node, toponym values, one link (to region, `TermKey` `part_of`); `NodesOnBridge([city], "")` → values, no links; `NodesOnBridge([missing], …)` → absent; after each, `g.Cached(city)` is false | AC-1.3, D-2 |
| T-2 | Go | `core/database/graphcache/inputs_test.go` `TestInputScope` subtest "the recorder hears Add" *(new)* | A recording scope and a `NoInputs` scope each `Add(k)` → the recorder hears `k` from both; `scope.Node(person)` → heard once | D-10 |
| T-3 | Go | `core/database/conclusionsummaries/placeentries_test.go` `TestChainsSkipLocationLinks` *(new)* | The package fixture (`newFixture`, `placeSubject`, `placeRel`, `promoteSubject`, `lifeEvent`): Town ⊂ City ⊂ Region, three births located at City, one at Town, all promoted. (a) `f.C.Graph().Drop()`, `ParentsAtDate(town, today)` → `["City", "Region"]`; `Cached` of Town, City and each event is false. (b) `RowsForSearch(graphcache.Over(db), Want{Events: [townBirth]})` → its place `Town` with parents `["City", "Region"]`; `Cached(city)` is false on that graph | AC-1.1, AC-1.2 |
| T-4 | Go | `placeentries_test.go` `TestPlaceEntries` *(new)*, table-driven: each case builds the T-3 world, holds entries for Town, City, Region (one `ParentsAtDate(town)` after `Drop`), runs its write, then asserts | "person write": `f.named(person, "Bo")` → the places key count is unchanged and `ListPlacesQueryCount` after equals the warm count before; "city rename": a toponym "Big City" cited on City → `HeldPlaceEntryForTest(g, city)` false, Region's true; the chain's first name is `PlacesByIDs(f.C, [city])[0].Names[0]` and those `Names` contain "Big City"; "remove link": `placeRel` returns the bridge Subject (`res.Subject`, whether or not dates were added), deleted by `writes.Call(f.C, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) { ch, err := subjects.Delete(tx, userID, bridge.ID); return struct{}{}, ch, err })` → Town and City not held, chain `[]`; "location at city": `f.lifeEvent(f.lifePerson(), "birth", "date", 1900, 0, citySubject)`, promoted → City not held; "vocabulary": `graphcachetest.NewCanon(t, f.C).CommitVocabulary()` → key count 0; "cap": `SetPlaceEntryCapForTest(g, 2)`, a chain of five places → `ParentsAtDate` has four names, key count ≤ 2; "corrupt": `CorruptPlaceEntryForTest(g, city)` → `g.Verify()` error contains "stored place entry"; "missing place": `ParentsAtDate(f.C, unknown, today)` for a random 16-byte id → `[]`, and `PlaceEntryForTest(g, unknown)` → held, `found` false, inputs `[{unknown, "", ""}]` | AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7, AC-3.2, D-5, D-11, D-12 |
| T-5 | Go | `core/database/conclusionsummaries/inputs_test.go` (unchanged) `TestInputsMatchIsolatedComposition`, `TestInputsHoldEveryShownParent`, `TestInputsCapCountsNothing`, `TestListsRecordNoInputs` | As today | AC-3.1, D-4, D-10 |
| T-6 | Go | `placechains_test.go` (unchanged), `walks_test.go`, `display_cache_test.go`, `eventplaces_cap_test.go`, `conclusionsummaries_test.go` `TestListPlacesQueryCountIsConstant` | As today, and under `PROVENENCIA_GRAPH_VERIFY=1` | D-7, D-8, D-13 |
| T-7 | Go | `bench_test.go` `BenchmarkListPlaces` *(new)*; `placeentries_test.go` `BenchmarkPlaceMesh` *(new)* | Reported, not gated (§11 holds the stops) | D-14, D-15 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestInputsMatchIsolatedComposition` (if D-4 or D-10 drifts), `TestAttachPlaceRelationships*` (D-8), `TestListPlacesQueryCountIsConstant` (a fill per row instead of per round), `TestCacheDocListsEveryEntry`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] The conclusionsummaries, graphcache, persondetail and api suites pass with PROVENENCIA_GRAPH_VERIFY=1
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: `rg -n "\bnodeKey\b|\bfollowKeys\b|rememberShell|rememberPlaceLink" core .agents .cursor/rules docs/*.md` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (catalog-queries.mdc)
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump: none
- [ ] Ship log: STEP-1 and STEP-7 figures with hostname, before and after
- [ ] docs/catalog-caches.md: the conclusionsummaries.places row with measured sizes
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ — no brief
- [ ] Domain doc: cache-architecture.md § A cache a composer reads hands its inputs to the row names NodeKey and FollowKeys
- [ ] PR description: what shipped, AC evidence, deviations, the measured figures
```

## 10. Manual check

1. Open a project with a place chain (Toronto `part_of` York County `part_of` Upper Canada) and events at Toronto. Open Places: each row's parents read as before. Open Toronto's page: Part of, Contains, Predecessors and Successors read as before.
2. Rename York County in its Source; return to Places and to an event at Toronto: the chain shows the new name.
3. Remove the Toronto → York County `part_of` record; Toronto's row and page show no parent.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The isolation oracle hears a different key set (D-4, D-10 drift) | T-5 fails first; D-4 lists the exact keys |
| The always-narrow fill makes the first read after open slower | `BenchmarkListPlaces` cold at both sizes; stop below |
| A hub rename drops 20,000 part entries | Stated in §6 Scale; `BenchmarkPlaceMesh` measures the refill |
| A commit that removes a link reaches neither end through `roots` | `SummaryDeps` carries the old ends (D-5); T-4 "remove link" holds it |

**Stop and report if:**

- A Current state row is false at your base (for example `buildPlaceGraph` no longer walks nodes, or `OnCommit` no longer applies entries after its reload).
- `TestInputsMatchIsolatedComposition` fails after STEP-4 and the difference is a key D-4 does not list (report the key).
- `SummaryDeps` does not hold both old ends of a removed `place_relationship` link in T-4 (D-5 would be wrong).
- `BenchmarkListPlaces` cold at 16,000 persons is more than 50 ms slower after than before, or warm is more than 10 % slower.
- `BenchmarkPlaceMesh` cold or after-rename `ListPlaces` is more than 1.25 × slower after than before.
- `BenchmarkCacheSizes` `placeEntries-B` exceeds 5 MB at 16,000 persons.
- `TestLargeSourceBudgets` fails a subtest it passed in STEP-1.

## 12. Open questions

None.

## 13. PR

**Title:** S16-18 — Place entries
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence; deviations from the plan; the STEP-1 and STEP-7 figures; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-18-place-entries.md`.
