# S16-18 — Place entries

**Status:** Landed in [#817](https://github.com/mendahu/provenencia/pull/817) (2026-11-13)
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
**Review:** [`S16-18-place-entries.review.md`](S16-18-place-entries.review.md) — Fits with changes / Ready with changes, two rounds, no open blocker

## Landed

- `eventPlaceGraph` calls `g.Nodes(placeIDs)` before `buildPlaceGraph`: `eventPlaces` reads each place's node one at a time, and the chain build's old node load had batched it (`TestPersonDerivedRows`, `TestPersonEventsQueryCount`). S16-19 deletes both.
- `buildPlaceGraph` keeps its signature; its `rec` parameter is unused, since `closure` reads the entries' `inputs`.
- An entry's `partFrom`/`partTo` and `succFrom`/`succTo` are filled by orienting each followed link and filing it under the end that is the place, so a self-loop sits in both lists.
- An extra test hook, `FillPlaceEntriesForTest`, holds entries for T-8 and T-11.
- The plan's `\bbyID\b` sweep hits unrelated `byID` locals in `composePersons`.
- §11 was amended by the user (2026-11-13) after round 0: the hub-commit refill has a 1 s ceiling (measured 560 – 668 ms against 142 ms), the cold crafted-hub read a 3 s ceiling on any of three runs (measured 1706 – 2366 ms against 1328 – 1522 ms), and cold-list and after-location compare the median of three runs.
- The ship log's STEP-1 and STEP-9 figures are in the PR description.

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

Every place chain ("Toronto, York County, Upper Canada") is built today by `buildPlaceGraph`, which loads each place's whole canonical node, with every link it has, then re-reads the `part_of` links and their associations, on every Places list read, every place page and every summary batch. A city's node carries its 6,000 events' location links, so any chain that passes through the city loads them. This PR keeps one **place entry** per place in the graph cache's registry: its shell (names and period), its filed `part_of` and `succeeded_by` links in both directions with their decoded dates, and the input keys a summary row records for it. Entries fill a frontier at a time by a narrow load that reads only `place_relationship` links. The graph cache's registry learns to update keyed entries exactly by itself: an entry declares the `graphcache.InputKey`s its value was built from (`Inputs`, the same keys a summary row records), and a commit drops only the keys whose inputs it moved, so a place entry needs no Apply of its own and survives every write that does not touch its names or its place links. `buildPlaceGraph` becomes a walk over the entries, so the Places list, the place page's relationships and the event and person summary composers all read chains from them, and the node-reading helpers go.

**Why now.** Bar 9 and bar 13 ("Walks are lookups": a city's place chain reads without loading its 6,000 location links). S16-19 reads an event's own places from the entries, and S16-21 reads Promote's place links from them; both inherit the exact update. The user chose the registry-level fix over a drop on every place the commit names (2026-11-13: "not a band-aid").

**What stays.** Every chain, fold, place page relationship and summary row reads as today, in today's order; every summary row records the same input keys (S14-40's tests hold it). `summaryChainDepth` (8), the input cap (`inputKeyCap`), S16-01's `eventLocations` cap of 20 places per event and the composers' own reads of an event's places stay. The Places list still loads each listed place's node for its row. No SQL table, wire or Swift change.

## 2. User stories

**US-1.** As a researcher, when a page shows a chain that passes through a city with 6,000 events, the engine does not load those events' location links to draw it, so the page and every write at that city stay fast.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Town `part_of` City `part_of` Region, promoted; three events located at City; the graph is dropped. `ParentsAtDate(town, today)` | `["City", "Region"]`; afterwards `g.Cached(city)` and `g.Cached(town)` are false, and no event is cached |
| AC-1.2 | The same catalog; an event at the town, composed alone on a write's one-shot graph (`RowsForSearch(graphcache.Over(db), Want{Events: [event]})`) | The event row's places read "Town" with parents `["City", "Region"]`; `Cached(city)` is false on that graph |
| AC-1.3 | City `part_of` Region and an event located at City (`graphcachetest.Canon`). `NodesOnBridges([city], "place_relationship")` | One node: its values; exactly its `part_of` link to Region (no location link); nothing is cached on the graph |

**US-2.** As the engine, a place entry is filled once and dropped only by a commit that writes its place, one of its links, or a new `place_relationship` link at it, so repeat reads compose from memory.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | `ListPlaces` twice on the session graph (the second is warm), then a write that names a Person, then `ListPlaces` | `KeyCounts()["conclusionsummaries.places"]` is the same before and after the write, and the last `ListPlaces` issues no more queries than the warm one |
| AC-2.2 | Entries held for Town, City, Region; a toponym "Big City" is cited on City's Subject (a real write) | Right after the write City's entry is not held, and Town's and Region's are; the next `ParentsAtDate(town)` starts with `PlacesByIDs(f.C, [city])[0].Names[0]`, and that row's `Names` contain "Big City" (rank order decides which name shows) |
| AC-2.3 | Entries held for Town and City, neither node cached; the Town → City `place_relationship` Subject is deleted (a real write) | Town's and City's entries are not held (the association is in the write's `Handles` and both entries list its `NodeKey` in `Inputs`); Region's is; `ParentsAtDate(town)` is `[]` |
| AC-2.4 | A `graphcachetest.Canon` world: Town ⊂ City ⊂ Region, entries held. (a) `loc := Location(Event("birth"), city)`, `Commit(loc)`; (b) `link := PartOf(Entity("place"), city)`, `Commit(link)`; (c) `Merge(region, into)` for a new place `into`, `Commit(region, into)` | (a) every entry still held; (b) City's entry not held, Town's and Region's held; (c) City's entry not held (its `place_relationship` neighbour merged), Town's held |
| AC-2.5 | A vocabulary commit (`graphcachetest.NewCanon(t, f.C).CommitVocabulary()`) | Every place entry is dropped (`DerivedFrom` the vocabulary) |
| AC-2.6 | Verify on (`PROVENENCIA_GRAPH_VERIFY=1`) through the whole conclusionsummaries suite | Passes; a held entry corrupted by a test fails Verify with "conclusionsummaries: stored place entry" |
| AC-2.7 | The entry's Cap lowered to 2 (`SetPlaceEntryCapForTest`), a chain of five places read | The chain reads in full; at most 2 keys are held afterward |

**US-3.** As the engine, a summary row that reads a chain from the entries records the keys it recorded before, so S14-42 still finds every row a write moves.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `TestInputsMatchIsolatedComposition`, `TestInputsHoldEveryShownParent`, `TestInputsCapCountsNothing`, `TestListsRecordNoInputs` | Unchanged and passing |
| AC-3.2 | An unknown 16-byte id read as a seed (`ParentsAtDate(f.C, unknown, today)`, then `PlaceEntryForTest(g, unknown)`) | `ParentsAtDate` is `[]`; the entry is held with `found` false and inputs `[{unknown, "", ""}]`, as today's `in.Node` of a missing node records. (A link never names a missing place: `linkLoadSQL` joins the neighbour entity, so only an explicit seed reaches this case) |

**US-4.** As a cache owner, I declare the input keys my keyed values were built from and the registry drops exactly the keys a commit moves, so no owner writes a drop rule or guesses from roots.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | A test entry with `Inputs`: key A built from `{A}` and `{X}`; key B from `{B}` | A commit with `Handles` [X] drops A and keeps B; [Y] keeps both; [B] drops B |
| AC-4.2 | A test entry with `Inputs`: event E from `{E, "location", "event"}`, place P from `{P, "location", "place"}` | A committed new location between E and P drops both; a committed participation at E keeps both; merging a handle located at P drops P |
| AC-4.3 | Store, overwrite (new inputs), evict past the Cap, `Drop`, `Clear` on an `Inputs` entry | After each, the input index names only held keys (`InputIndexSizeForTest`) |
| AC-4.4 | `RegisterKeyed` with neither Apply nor Inputs | Panics with "needs an Apply or Inputs" |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache` exports `NodeKey` and `FollowKeys` (today `nodeKey`, `followKeys`) | STEP-2, D-9 |
| `graphcache.(*Node).Follow`, extracted from `Graph.follow` | STEP-2, D-9 |
| `graphcache.(*Graph).NodesOnBridges`: values and only the named bridges' links, uncached; one row-to-node constructor (`readNodes`) shared with `loadChunk` | STEP-2, D-2, AC-1.3 |
| `InputScope.Add` reports to the graph's test recorder | STEP-2, D-10, AC-3.1 |
| Keyed entries declare `Inputs` (`[]InputKey`); the registry keeps an input index and drops the keys whose inputs a commit moves; `OnCommit` hands the registry the moved inputs (written handles, committed links' ends by bridge, merged handles' links) | STEP-3, D-16, D-17, AC-4.1, AC-4.2, AC-4.3, AC-4.4 |
| Keyed entry `conclusionsummaries.places`: fill, `Inputs`, `DerivedFrom`, Check, Cap | STEP-4, D-1, D-3 – D-6, D-11, D-12, AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5, AC-2.6, AC-2.7 |
| `buildPlaceGraph` walks the entries; `placeGraph` reads entries; `placeStep`, `finishPlaces`, `AttachPlaceRelationships` read them; `rememberShell`, `rememberPlaceLink` go | STEP-5, STEP-6, D-7, D-8, AC-1.1, AC-1.2, AC-3.1, AC-3.2 |
| Tests for the entries, including a guard that their keys equal what a scope records | STEP-4, STEP-7, D-18 |
| `docs/catalog-caches.md` row and rule, `BenchmarkCacheSizes` fill, `cache_doc_test.go` import | STEP-8 |
| Measured before and after: `BenchmarkListPlaces` *(new)*, `BenchmarkPlaceMesh` *(new)*; `Canon.PartOf` *(new)* | STEP-1, STEP-9, D-14, D-15 |
| The map, design doc, `catalog-queries.mdc` and the maintainability lens say what landed | STEP-10 |

| Out | Home |
| --- | --- |
| The event and person composers reading an event's own places from the entries (`eventPlaces`' `in.Node`, `composeEvents`' `g.Nodes(placeIDs)`) | S16-19 |
| `PersonLifeOwners`, `lifeOwners`, `eventPlaceGraph` | S16-19 |
| Promote's connections reading place links from the entries (an exported reader) | S16-21 |
| The Places list's own row reads (`nodesOf` loads each listed place's node for its shell and lead picture) | not planned: a row reads its own node, as every summary row does |
| Storing a composed chain (`ParentChain` at a date) | not planned: rule "No new persisted projections, and no stored answer without a measurement" |
| Filling place entries at open | not planned: rule "Warm only what every proposal needs" |
| `ParentsAtDate`, `PartsAtDate`, `SuccessionNames` having only test callers | not planned: they are the tests' entry points to the chain rules |
| `canonlinks.kinship` declaring `Inputs` (its pivots' `FollowKeys`) instead of its cached-neighbour Apply (`deps`) | [structural debt idea](../../../ideas/structural-debt.md) (a row added in STEP-10) |
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
| Node load | `core/database/graphcache/load.go` `loadIDs`, `loadChunk`, `entityRows`, `StoredValues`, `loadMembers`, `loadOutcomes`, `loadLinks`, `linkRaws`, `linkLoadSQL`, `bridgePredicates`, `finishLinks` | `loadChunk` reads entity rows, values, members, outcomes and every bridge's links, builds each `*Node` in one literal (the only place a node is built from rows), and installs it; `linkLoadSQL` constrains to every registered bridge (`bridgePredicates`) |
| Commit | `core/database/graphcache/commit.go` `OnCommit`, `detachAssociation`, `mergedNeighbors` | Reads `entityRows` of every committed id, so it knows each committed association's bridge (`row.typeKey`); `detachAssociation` returns its ends (loaded holders plus `endpointIDs` from SQL); `mergedNeighbors` reads each merged handle's links (`linkRaws`, each with its `bridgeKey`). Roots: the commit's handles, association ends, merged neighbours; a removed node adds its links' neighbours only when it was cached. Then `applyEntries(roots, set)`. A removed association's ends are known only through loaded holders |
| Handles of a removed link | `core/database/effects/registry.go` `identity_claims` (`Handles: union(field("entity_id"), …)`), `observations` (`observationMembers`) | Deleting a bridge Subject deletes its claim and Observations, so the association id is in `set.Handles` |
| Registry | `core/database/graphcache/cache.go` `KeyedEntry`, `RegisterKeyed` (panics without `Apply`), `keyedValues`, `store`, `evict`, `Keyed.GetMany`, `Drop`, `Clear`, `SetCapForTest`, `applyEntries`, `DropNamedGraphs`; `cache_test.go` `TestEntryApply`, `TestKeyedEvictsLeastRecentlyUsed`, `TestRegisterRejects` | Keyed entries store on a one-shot graph (`Over`) and go with it. An entry's only commit hook is `Apply(g, k, roots, set)`; no entry can say what its values depend on, so `canonlinks.kinship` walks cached neighbours (`deps`) |
| Isolation oracle | `core/database/conclusionsummaries/conclusionsummariestest/inputs.go` `newIsolator`, `CheckMatchIsolated` | Composes each row alone on one `Over` graph and hears keys through `SetInputRecorderForTest`, which hears only `Graph.Node` and `Graph.Follow` |
| Display check | `core/database/conclusionsummaries/display_register.go` `checkDisplay` | Under Verify, composes fresh rows through `persons`, `events`, `places` |
| Cache doc | `docs/catalog-caches.md` (line 5: "The signal is a field of `effects.Set` that the entry's Apply reads, or a `DerivedFrom` declaration"); `core/catalogsession/cache_doc_test.go` (blank imports; `conclusionsummaries` is not imported); `core/database/promotealign/bench_test.go` `BenchmarkCacheSizes` | No place row |
| Test builders | `core/database/graphcache/graphcachetest/graphcachetest.go` `Canon` (`Entity`, `Link`, `Location`, `Join`, `Merge`, `Remove`, `Commit`, `CommitVocabulary`, `Bulk`) | No place-link builder |
| Tests | `placechains_test.go` (`TestPlaceParentsAtDateFromPeriods` … `TestUndatedPlacesAlwaysHold`, `TestAttachPlaceRelationships*`, `TestPersonLifeOwners`), `placechains_internal_test.go`, `inputs_test.go`, `display_cache_test.go`, `eventplaces_cap_test.go`, `walks_test.go`, `conclusionsummaries_test.go` `TestListPlacesQueryCountIsConstant`; `core/database/graphcache/inputs_test.go` `TestInputScope`; `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` | Must stay green |

**Precedent.** S16-10 ([#789](https://github.com/mendahu/provenencia/pull/789)) registered `canonlinks.kinship` (`core/database/canonlinks/reader.go` `pivots`): a keyed entry from `init()` with `DerivedFrom: Vocabulary{}`, an Apply that drops roots, a `Check` that re-derives and `reflect.DeepEqual`s, `Cap: maxKeys` (65,536). Copy that registration shape. S16-13 ([#809](https://github.com/mendahu/provenencia/pull/809), `core/database/graphfill/records.go` `RecordsOf`) fills a keyed entry through `GetMany` with a batched fill closure; copy its fill-closure shape. Do not copy `pivots`' Apply (`deps` walks cached neighbours, which misses a removed link whose ends are not loaded: the gap D-16 closes).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | A keyed registry entry `conclusionsummaries.places` *(new)*, handle `placeEntries graphcache.Keyed[placeEntry]` *(new)* in `core/database/conclusionsummaries/placeentries.go` *(new)*, keyed by place id. `placeEntry` *(new)*: `found bool`; `shell PlaceSummary` (from `placeShell`); `period valuecompare.Window`; `partFrom`, `partTo`, `succFrom`, `succTo []placeLink` (the place as part, as whole, as predecessor, as successor), each oriented as `rememberPlaceLink` does today and in the node's link order; `inputs []graphcache.InputKey` (D-4). Values are shared and read-only. Each list is deduped on from/to/association as `rememberPlaceLink` does today; a self-loop `part_of` sits in both `partFrom` and `partTo`, as today **Amended (user, 2026-11-13, round 1 SEC-1/PRF-1):** two keyed entries, not one. `conclusionsummaries.places` (`placeEntries`, `placeEntry`) keeps `found`, `shell`, `period`, `partFrom` and `inputs`: what a chain reads. `conclusionsummaries.placePages` *(new)* (`placePages graphcache.Keyed[placePage]`, `placePage{partTo, succFrom, succTo []placeLink; inputs []graphcache.InputKey}`) holds the lists only the place page reads, filled only by `placeStep` for its own id. A chain read never decodes a child or succession link's dates, so one bad date there fails only that place's page, as at base, and a write under an N-part whole reads no part link's association. A self-loop `part_of` sits in `partFrom` and in the page's `partTo`. | `conclusionsummaries` owns `placeGraph`, `placeShell` and the chain rules; per-key entries are the map's rule "results per key" | An entry in `graphcache` (it would import place semantics); a whole entry (a write's graph would build every place) |
| D-2 | Fill by a narrow load: `graphcache.(*Graph).NodesOnBridges(ids [][]byte, bridges ...string) (map[string]*Node, error)` *(new)* reads entity rows, values and only the named bridges' links (none when no bridge is named), in `sqlbatch.Each` chunks, and caches nothing: `Members` and `Outcomes` are empty. A missing id is absent; duplicate ids are read once (`uniqueIDs`); a bridge key that is not registered is an error, `"graphcache: bridge %q is not registered"`; with no bridge, no link or disambiguation query runs. Nodes are built from rows in one place: `loadChunk` splits into `readNodes(ids [][]byte, o readOpts) (map[string]*Node, error)` *(new)* (rows to nodes; `readOpts{bridges []string; full bool}`, `full` reads members and outcomes) and its install loop; `loadChunk` calls `readNodes(ids, readOpts{bridges: bridgeKeys(), full: true})`, `NodesOnBridges` calls `readNodes(ids, readOpts{bridges: bridges})`. `linkRaws`, `linkLoadSQL` and `bridgePredicates` take `bridges []string`; `bridgeKeys()` *(new)* lists every registered bridge key, which today's callers (`loadLinks`, `mergedNeighbors`, `LinkLoadSQL`) pass. The entry always fills through `NodesOnBridges`, even when the place's node is cached | One fill path; the narrow load is what skips location links on a write's graph and on a cold session graph; one node constructor | Reading the cached node when held (two fill paths); installing a partial node (other readers would see a node without its links); a second node literal beside `loadChunk`'s (review M-2: a new `Node` field would miss one path); a `""` sentinel for "no bridges" (review M-5: it would mean "every bridge" one call below) |
| D-3 | A link's dates come from `NodesOnBridges(assocIDs)` (values only) and `assocDatesOf`. One fill per frontier: places' narrow nodes, then all their links' associations in one call. An association absent from that result is an error, `"conclusionsummaries: association %x is missing"`, as `rememberPlaceLink` returns today **Amended (round 1):** each entry's fill reads only its own lists' associations: the chain fill `partFrom`'s, the page fill `partTo`'s and succession's. | Same loader; `assocDatesOf` and `keptDateAt` read a `*Node` today | `StoredValues` wrapped in a hand-made `Node` in `conclusionsummaries` |
| D-4 | `placeEntry.inputs` = `NodeKey(id)` plus `FollowKeys(id, hops.ParentsOfPlace, partOfLinks)`, where `partOfLinks` is `n.Follow(hops.ParentsOfPlace)`; sorted with `SortInputKeys`. A missing place's inputs are `[NodeKey(id)]`. No keys for `PartsOfPlace` or succession | These are exactly the keys today's per-place scope records (`in.Node(id)`, `follow(in, id, ParentsOfPlace)`, `in.Node(assoc)` = `NodeKey(assoc)`, which `FollowKeys` already holds), so every row's closure is unchanged (AC-3.1, AC-3.2) | Every hop's keys: a hub's 20,000 part links would enter every row whose chain passes it |
| D-5 | The place entry has no Apply. It declares `Inputs` = `placeInputs(id, e)` *(new)*: `e.inputs` (D-4) plus `FollowKeys` of `PartsOfPlace`, `SuccessorsOfPlace` and `PredecessorsOfPlace` over its other lists, deduped; a missing place: `[NodeKey(id)]` (D-16). So a commit drops a place's entry when it writes the place (names, period, merge or deletion: `{id}`), writes or removes one of its links (the association's `{assoc}` is in `Handles`), files a new `place_relationship` link at it (`{id, place_relationship}`), or merges a place it links to; a location, participation or any other write leaves it **Amended (round 1):** `placeEntries` declares `Inputs` = `e.inputs` (D-4; a missing place `[NodeKey(id)]`); `placePages` declares `Inputs` = its `inputs`, built in the page fill from the followed links: `NodeKey(id)` plus `FollowKeys` of `PartsOfPlace`, `SuccessorsOfPlace`, `PredecessorsOfPlace` over the real `Follow` results (no `Link` rebuilt from a `placeLink`). `placeInputs` goes. | User, 2026-11-13: the robust, exact update, not a drop on everything the commit names; in the registry, for every keyed entry, in the dependency type the summary rows and the input index already use (review M-10) | `Drop(roots ∪ SummaryDeps)` (renaming a 20,000-part hub would drop 20,000 unchanged entries, against the rule "updates exactly"; review M-1); id `Deps` plus `Bridges` names (a third way to say "built from" beside `InputKey`; review M-10) |
| D-6 | `DerivedFrom: []graphcache.Source{graphcache.Vocabulary{}}`; `Cap: maxPlaceEntries` *(new const, 65,536)*; `Check` per D-12; keyed entries are graph-lifetime, so `Drop` and a `Structure` commit empty it | The shell and link terms decode through vocabulary ids; the cap matches `canonlinks.maxKeys` and holds every place of a real catalog | No cap (a crafted catalog could grow it without bound) |
| D-7 | `buildPlaceGraph(g, v, seeds, rec)` keeps its signature and walks the entries: per round, `placeEntries.GetMany(g, frontier, fill)`; each visited place's entry goes into `pg.places`; the next frontier is each entry's `partFrom` wholes not yet seen; `summaryChainDepth` rounds, as today. `placeGraph` becomes `{places map[string]placeEntry}`; every method reads it (`byID[id]` → `places[id].shell`, `partFrom[id]` → `places[id].partFrom`, and so on; `periodOf` of an absent place stays `Always`). It records no scope: `closure` reads `places[id].inputs`. Where a method tested `byID`'s `ok` (`displayName`, `relationship`, `membershipDates`' callers), it tests that the place is in `places` and `found`: a `found: false` entry reads as absent (`relationship` returns `Entity{ID: id}`, `displayName` returns `""`). `rememberShell`, `rememberPlaceLink`, the `linked` maps and the per-place scopes go | User, 2026-11-13: one chain builder, so the Places list, the place page and the event and person composers all read entries in this PR and no node-walking builder survives beside it | A new walker for the Places list beside `buildPlaceGraph` until S16-19 (the map's first split: two builders between the PRs) |
| D-8 | `placeStep(c, id)` loses its hop list: `GetMany([id])`, then `GetMany` of every neighbour on the entry's four lists; `pg.places` holds the whole entries. `AttachPlaceRelationships`, `PartsAtDate` and `SuccessionNames` call `placeStep(c, id)`. `finishPlaces` and `AttachPlaceRelationships` stop overwriting `byID` and `period` with the row: the entry's shell is the same `placeShell` of the same values **Amended (round 1):** `placeStep(c, id)`: `placePages.GetMany([id])` and `placeEntries.GetMany` of `[id]` plus every neighbour on the chain entry's `partFrom` and the page's three lists; `placeGraph` gains `pages map[string]placePage`, which `ImmediateParts`, `succNeighbors` and `AttachPlaceRelationships` read. Neighbours hold chain entries only. | Every reader of a step graph reads lists keyed on the step's own id, and only shells and periods of neighbours (review M-4); the only production caller passes all four hops | Trimming the step's lists to the asked hops and clearing neighbours' lists (code no reader needs) |
| D-9 | `graphcache` renames `nodeKey` → `NodeKey` and `followKeys` → `FollowKeys` (exported, doc unchanged) and extracts `func (n *Node) Follow(hop hops.Hop) ([]Link, error)` *(new)* from `Graph.follow`, which now loads the node and calls it | The map's rule ("S16-18 exports `nodeKey` and `followKeys`"); the entry filters a narrow node's links with the same rule `Graph.Follow` uses | Re-writing the bridge, end and term filter in `conclusionsummaries` |
| D-10 | `InputScope.Add` hands each key to the graph's test recorder (when set) before its `off` and self checks; `InputScope.Node` and `Follow` record through a new unexported `add`, since `Graph.Node` and `Graph.Follow` already report. `SetInputRecorderForTest`'s doc: "hears every key `Node` and `Follow` read and every key handed to `InputScope.Add`" | The isolation oracle (`CheckMatchIsolated`) hears reads only through the recorder; a chain read from an entry reaches a row through `addClosure`'s `in.Add`, so the oracle hears the same keys the batch records (AC-3.1) | Leaving the recorder alone (the oracle would miss every chain key); a second `Add`-like method for entry keys |
| D-11 | A missing place (no entity row): the fill stores `placeEntry{found: false, inputs: [NodeKey(id)]}`, whose `Inputs` are `[{id}]`, so a commit that creates it drops the entry. A merged place: built from what the narrow load returns, as today's node read does | Today's `in.Node` records a missing node's key, and `buildPlaceGraph` shows a merged node's shell; one entry shape | Not storing missing places (every read would query again) |
| D-12 | Check: `v` from `vocabulary(g)`, then rebuild the one id with the entry's fill (`fillPlaceEntries(g, v, [][]byte{id})`) and `reflect.DeepEqual`; mismatch error `"conclusionsummaries: stored place entry %x does not match a fresh load"` **Amended (round 1):** each entry checks by rebuilding through its own fill; the page's mismatch error is `"conclusionsummaries: stored place page %x does not match a fresh load"`. | `sources.records` and `canonlinks.kinship` check by rebuild; the narrow load reads the committed rows | A Check that reads cached nodes (the entry is not built from them) |
| D-13 | `checkDisplay` composes through `places`, `events`, `persons`, which now `GetMany` place entries; under Verify that may fill or touch entries. Accepted | Verify runs only under `PROVENENCIA_GRAPH_VERIFY`; the entries it fills are exact | A side door that composes without the registry |
| D-14 | No time budget. STEP-1 and STEP-9 record `BenchmarkListPlaces` *(new)* at 4,000 and 16,000 persons (cold: `Drop`, then `ListPlaces`; warm: a second `ListPlaces`; write: `RowsForSearch` of every place on `graphcache.Over(db)`) and `BenchmarkPlaceMesh` *(new)* (D-15's crafted mesh) on the same machine | The PR claims a counter (AC-1.1); rule "Measure growth" applies to speed claims, and the figures still show the cost of the always-narrow fill (D-2) | A new constant in `perf_test.go` |
| D-15 | `BenchmarkPlaceMesh`: through `graphcachetest.Canon` in one `Bulk` (places `Canon.Entity("place")`, links `Canon.PartOf(part, whole)` *(new)*, locations `Canon.Location`), a hub place with 20,000 parts (bar 2's 20,000-child hub), the hub the lowest place of an eight-place `part_of` chain (the hub `part_of` W1, W1 `part_of` W2, and so on to W7: seven wholes above it), and one event located at the hub per 10 parts (2,000 location links). Sub-benchmarks: `ListPlaces` cold (`Drop` first) and warm; `AttachPlaceRelationships` of the hub's row, cold; `ListPlaces` after `Canon.Commit(hub)` (what a write of the hub's names commits); `ListPlaces` after a new location at the hub is committed | The map's Check (Apply's work stated for a crafted `part_of` mesh) and rule "Apply states its fan-out" | A crafted mesh in the large fixture (it has no `part_of` links) |
| D-16 | `KeyedEntry[T]` gains `Inputs func(id []byte, v T) []InputKey` *(new, optional)*: the keys value `v` of key `id` was built from, in `InputScope`'s sense. The registry keeps, per entry, each held key's inputs and an input index from `(ID, Bridge)` to keys, updated in `store` (an overwrite replaces the old inputs), `evict`, `Drop` and `Clear`. In `applyEntries`, before the entry's Apply, it drops every key with an input among the commit's moved inputs (D-17), matched on `(ID, Bridge)`. Wiring: `RegisterKeyed` builds, over the typed handle, an `entry.dropMoved func(g *Graph, moved []InputKey)` *(new)* closure (as it builds `info.apply`); `keyedValues` keeps `inputs` (key → its index keys) so `store`, `evict`, `Drop` and `Clear` unindex without calling `Inputs` again; `applyEntries` runs `dropMoved` when the entry is held, then Apply when it is set and the entry is still held. An entry that reads a link declares the link's association `{assoc}` (as `FollowKeys` does): that is all a removed link moves, because a removed association has no row to give a bridge | One dependency type across the persisted input index, the summary closure and the registry; `Handles` names what a write changed, and a removed link's association is in it (§4); the registry, not each owner, keeps the index in step with eviction. `Endpoint` is not matched: the commit has the bridge and ends without a new query but not each end's endpoint Property, and a bridge-level match drops nothing that a place entry did not read | Ids (`Deps`) and bridge names (`Bridges`) as two new concepts (review M-10, M-11); matching `Endpoint` too (a change to `endpointIDs`' query for no gain on this entry) |
| D-17 | `OnCommit` collects the commit's moved inputs `[]InputKey` with no new query: `{h}` for each id in `set.Handles`; `{end, bridge}` for each end in `detachAssociation`'s `touched` of each committed association that has a row (bridge `row.typeKey`); `{neighbor, bridge}` for each link of a merged handle (`mergedNeighbors` returns them from its `linkRaw`s' `bridgeKey` and `neighbor`). It passes them to `applyEntries(roots, set, moved)`. `RegisterKeyed` requires an Apply or Inputs; Apply runs only when set | A new link names no held key's `{assoc}`, so the link's ends at its bridge must move; the bridge and ends are already in hand in `OnCommit`. A removed link moves only `{assoc}` (D-16) | Passing the moves to every Apply (each owner would filter them again); a query per committed association |
| D-18 | A guard test pins D-4 and D-5 to real reads: for each place of the T-3 world, (a) the entry's `inputs` equal `in.Keys()` of a `g.Inputs(nil)` scope on a fresh `graphcache.Over(db)` after `in.Node(p)`, `links := in.Follow(p, hops.ParentsOfPlace)` and `in.Node(l.Association)` for each link; (b) `placeInputs` equals the same with all four place hops followed (T-8) **Amended (round 1):** (a) is the chain entry's `inputs`; (b) is the page entry's `inputs` against a scope that does `in.Node(p)` and follows the other three place hops, with `in.Node` of each link's association. | With D-10 the isolation oracle hears the entry's own keys, so it no longer checks them against real reads (review M-3) | Relying on `TestInputsMatchIsolatedComposition` alone |

**Rules this plan bends.**

- `.cursor/rules/catalog-queries.mdc`: "a summary composer reads other handles through its `graphcache.InputScope` (`Node`, `Follow`) so its inputs are recorded". The composers now read chains from a registry entry and record its keys through `InputScope.Add`. The map's working rule "A cache a composer reads hands its inputs to the row" (user, map review K-21, 2026-11-08) asked for exactly this. STEP-10 rewrites the line: "a summary composer reads other handles through its `graphcache.InputScope` (`Node`, `Follow`), or reads a registry entry that recorded its keys with `graphcache.NodeKey` and `FollowKeys` and adds them to the scope, so its inputs are recorded".
- The map's S16-18 / S16-19 split: changed by the user on 2026-11-13 (D-7). The map's rule "One cache registry, two ways to fill" and the design's registry section gain keyed `Inputs` (user, 2026-11-13, D-16, D-17). The map row, scope entries, rule and design doc are updated in the planning commit.
- `.cursor/rules/catalog-queries.mdc` and `docs/catalog-caches.md` say a keyed cache declares its Apply: STEP-8 and STEP-10 make them "its Apply or its `Inputs`" (same answer).

## 6. Design

### Codebase fit

Searched: `RegisterKeyed`, `GetMany(`, `nodeKey`, `followKeys`, `func (g \*Graph) follow`, `StoredValues(`, `linkLoadSQL`, `bridgePredicates`, `loadChunk`, `placeGraph`, `buildPlaceGraph`, `SetInputRecorderForTest`, `Canon) Link`, `Benchmark` in `core/database/conclusionsummaries` and `core/database/promotealign`.

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `core/database/conclusionsummaries/placeentries.go` *(new)* | file | `display_register.go` (the package's other registration) | That file registers the display check; the entry has its own type, fill, Apply and Check |
| `placeEntry` *(new)* | type | `placeGraph` (`placechains.go`) | `placeGraph` stays as the walk's view, now a map of entries (D-7) |
| `placeEntries` *(new)*, `maxPlaceEntries` *(new)* | keyed handle, const | `canonlinks` `pivots`, `maxKeys` | A different value type and owner; the registry is the shared part |
| `placeInputs(id []byte, e placeEntry) []graphcache.InputKey` *(new)* | func | `canonlinks` `deps` | That walks cached neighbours at commit time; this lists, with `FollowKeys`, what the value was built from |
| `fillPlaceEntries(g *graphcache.Graph, v vocab, ids [][]byte) (map[string]placeEntry, error)` *(new)* | func | `rememberShell` + `rememberPlaceLink` + `assocDatesOf` | It replaces the two `remember*` helpers (deleted); `assocDatesOf`, `placeShell`, `periodWindow` are reused |
| `placeEntriesOf(g *graphcache.Graph, v vocab, ids [][]byte) (map[string]placeEntry, error)` *(new)* | func | `graphfill.RecordsOf` | The one read: `placeEntries.GetMany` with `fillPlaceEntries`; `buildPlaceGraph`, `placeStep` and the Check call it |
| `graphcache.(*Graph).NodesOnBridges` *(new)*, `readNodes` *(new)*, `readOpts` *(new)*, `bridgeKeys` *(new)* | func / type | `loadChunk` (`load.go`) | `readNodes` is `loadChunk`'s row-to-node half, so one literal builds every node (D-2) |
| `linkRaws(ids, bridges []string)`, `linkLoadSQL(valueTable, entityTable string, ids [][]byte, bridges []string)`, `bridgePredicates(bridges []string)` (new parameter on each) | signature | themselves | An explicit bridge set; today's callers pass `bridgeKeys()` |
| `KeyedEntry.Inputs` *(new field)*, the input index in `keyedValues` *(new)*, `applyEntries`' `moved []InputKey` parameter | registry | `KeyedEntry.Apply`, `DropNamedGraphs`; `searchindex` `ReprojectChanged` / `dependentsOf` (the persisted input index) | Apply stays for entries whose update is not a drop (`sources.records` patches). The persisted index finds summary rows in SQL inside the write; this one finds held cache keys in memory after the commit, keyed by the same `InputKey` (D-16) |
| `func InputIndexSizeForTest[T any](g *Graph, k Keyed[T]) int` *(new, test)* in `core/database/graphcache/export_test.go`, counting (index key, held key) pairs | helper | `SetCapForTest` | Reports a keyed entry's index size so AC-4.3 can see eviction keep it in step |
| `TestKeyedInputs`, `TestKeyedInputLinks` *(new, test)* in `core/database/graphcache/cache_test.go` | test | `TestEntryApply` | — |
| `Canon.PartOf(part, whole []byte) []byte` *(new)* in `graphcachetest.go` | helper | `Canon.Location` | Review M-9: a place-link builder beside the location builder |
| `graphcache.(*Node).Follow` *(new)* | method | `Graph.follow` | Extracted; `Graph.follow` calls it |
| `graphcache.NodeKey`, `graphcache.FollowKeys` (renamed) | func | `nodeKey`, `followKeys` | Exported, per the map |
| `InputScope.add` *(new, unexported)* | method | `InputScope.Add` | `Add` now reports to the recorder (D-10); `Node` and `Follow` record without reporting twice |
| `BenchmarkListPlaces` *(new, test)* in `core/database/promotealign/bench_test.go` | benchmark | `BenchmarkCacheSizes` | The large fixture (`newLargeFixture`, `perf_test.go`) is in that package; `BenchmarkCacheSizes` measures heap, not read time |
| `BenchmarkPlaceMesh` *(new, test)* in `core/database/conclusionsummaries/placeentries_test.go` *(new)* | benchmark | `eventplaces_cap_test.go` `newPlacesWorld` (a crafted world through `graphcachetest.Canon`) | A different shape (a hub and a chain); the `Canon` calls are the reuse |
| `TestPlaceEntries`, `TestPlaceEntriesFollowLinkMoves`, `TestChainsSkipLocationLinks`, `TestPlaceEntryInputsMatchReads` *(new, test)* in `placeentries_test.go` | test | `TestForFollowsCommits` (`canonlinks`) | — |
| `HeldPlaceEntryForTest(g, id) bool`, `PlaceEntryForTest(g, id) (held, found bool, inputs, placeInputs []graphcache.InputKey)` (through `Peek`), `SetPlaceEntryCapForTest`, `CorruptPlaceEntryForTest` *(new, test)* in `core/database/conclusionsummaries/export_test.go` | helper | `SetInputKeyCapForTest` (same file) | The entry handle is unexported; these wrap `placeEntries.Peek`, `SetCapForTest` and `Set` (a copy of the held entry with `shell.Names` set to `["corrupt"]`) for the external test package |
| `placeRel` returns the bridge `subjects.Subject` (test helper) | helper | itself | Callers ignore the result; T-4 deletes the bridge |
| `TestNodesOnBridges` *(new, test)* in `core/database/graphcache/graph_test.go` | test | `TestLoadLinksUsesTheEdgeIndex` | That test checks the query plan, not what the load returns |

**Builds on.** The registry's keyed entries and `GetMany` batch fill (S16-10), which gain declared dependencies; `graphcache`'s link query and `connectrules` bridges (no hand-written association join); S14-40's `InputKey`, `inputBudget` and closure; `graphcachetest.Canon` for crafted catalogs.
**Deletes.** `rememberShell`; `rememberPlaceLink`; `placeGraph`'s `byID`, `partFrom`, `partTo`, `succFrom`, `succTo`, `period` and `inputs` maps (replaced by `places`); `buildPlaceGraph`'s `scopeOf` scopes, `linked` map, `g.Nodes(frontier)` and `g.Nodes(assoc)`; `placeStep`'s node, neighbour and association loads, its `pending` type and `linked` map; `finishPlaces`' `pg.byID[id] = summaries[i]` and `pg.period[id] = …`; `AttachPlaceRelationships`' `g.byID[...] = *h` and `g.period[...] = …`; the unexported names `nodeKey` and `followKeys`; `loadChunk`'s node literal (moved into `readNodes`); `placeStep`'s hop parameter. The place-chain node walk is replaced by the entries for every caller.
**Next adds.** S16-19 makes `eventPlaces` read an event's own places' shells from `placeEntriesOf` and drops `composeEvents`' `g.Nodes(placeIDs)`, then deletes `eventPlaceGraph` with `PersonLifeOwners`. S16-21 exports a reader of an entry's `part_of` and `succeeded_by` links for `graphalign.Connections`; that adds a production import of `conclusionsummaries` from `promotealign` (today only its tests import it), which S16-21 decides on (review M-8). `canonlinks.kinship` can declare its pivots' `FollowKeys` as `Inputs` and drop `deps` (Out). Any keyed entry declares what it read instead of writing an Apply, and a cache a composer reads hands the row the same keys it is dropped by. A later entry that a composer reads records its keys the same way (`NodeKey`, `FollowKeys`, `InputScope.Add`).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A chain reader sees an entry filled before an unrelated commit | `placeGraph` methods | Dropped by its `Inputs` (D-5); AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| An existing keyed entry with no `Inputs` | `applyEntries` | Unchanged: its Apply runs as today; the index is skipped for it |
| A place id with no entity row reaches the walk (a dangling link end) | `fillPlaceEntries` | Stored `found: false` with `[NodeKey(id)]` (D-11); AC-3.2; `displayName` is `""` as today |
| A neighbour on the place page whose entry the LRU evicted between `GetMany` calls | `placeStep` | Each `GetMany` returns every asked id even when a later store evicts it (`cache.go`); the step keeps its own map |
| A one-shot write graph reads chains | `buildPlaceGraph` on `graphcache.Over` | Keyed entries store on that graph and go with it (`Keyed.store` has no `alt` check); AC-1.2 |

### Scale

At the map's Scale (large fixture at 4,000 and 16,000 persons; one city with 6,000 events; 51 places and no `part_of` links on that fixture) and bar 2's crafted 20,000-child hub:

- **Fill.** One frontier round issues, per `sqlbatch` chunk, four queries for the places (`entityRows`, `StoredValues`, the place-relationship link query, `finishLinks`' disambiguation) and two for their links' associations (`entityRows`, `StoredValues`); at most `summaryChainDepth` (8) rounds per read. On the large fixture every chain is one round (no `part_of` links), so the first `ListPlaces` after open issues six more queries than today, once; STEP-9 records it. The city's fill reads its values and no location row.
- **Session graph after open.** `promotealign.Warm` has loaded every place node, so today's walk reads them from memory; the entries query once per place and then read from memory (D-2). `BenchmarkListPlaces` cold records the difference at both sizes (§11 stop).
- **Update per commit.** `OnCommit` builds the moved inputs from data it already reads: one key per written handle (a write names a handful; a 500-participant event, about 501), two per committed association, one per link of a merged handle. For each entry with `Inputs`, one index lookup per moved input. Filing a location at the 6,000-event city: the moved inputs are `{loc}`, `{event, location}`, `{city, location}`; no place entry holds a `location` input, so none drops (the city drops only if the write's `Handles` names it). Renaming the city or the crafted hub: one entry drops. Removing a `part_of` link: its two ends. Merging the crafted hub into another place: `mergedNeighbors` already reads its 20,000 part links today; their 20,000 part entries drop, because each now links to a merged place. `BenchmarkPlaceMesh` "after the hub commit" and "after a location at the hub" record the refill.
- **Hub entry.** The hub's entry holds its 20,000 `partTo` links with their decoded dates; its fill reads 20,000 association value sets. Measured in round 0: `ListPlaces` after the hub commit 142 ms before, 560–668 ms after; accepted under a 1 s ceiling (§11; user, 2026-11-13). Today a chain through the hub loads the hub's whole node (the same 20,000 links plus 2,000 location links). `BenchmarkPlaceMesh` records both.
- **Memory.** The input index holds each held entry's inputs: about one plus two per link per place (the hub: about 40,000). One entry per place reached: on the large fixture 51 keys; STEP-9 records `placeEntries-B` and the key count in `BenchmarkCacheSizes` and `docs/catalog-caches.md`. Past `maxPlaceEntries` the least recently used place leaves (AC-2.7).

### Edges

- **Assumptions.** `parentChain` and `isAncestor` assume `pg` holds shells only for places the walk expanded within `summaryChainDepth` ("a name the build did not read must not appear because another row of the batch happened to read it"): `pg.places` holds only the entries the walk visited, in the same rounds (`TestInputsHoldEveryShownParent`, `TestPlaceParentsAtDateFromPeriods`). `reach` and `closure` assume `partFrom` holds the part-end `part_of` links the build followed: the entry's `partFrom` is `n.Follow(hops.ParentsOfPlace)`, the same filter. `ImmediateParents` and the other lists sort by `displayName` with `sort.SliceStable`, so ties keep link order: entries keep the node's link order (`finishLinks` sorts by association ref then neighbour ref, the narrow load included). Cycles: the walk's `seen` map is unchanged (`TestUndatedPlacesAlwaysHold`, the walks tests).
- **Neighbors.** `partTo` used to hold only the batch's children of a visited whole; it now holds every part. Only `ImmediateParts` reads it, and only through `placeStep` (place page, `PartsAtDate`), where today's step already followed `PartsOfPlace` in full (`TestAttachPlaceRelationships`, `TestIrelandUKDatedMembershipDropsAfter1922`). Succession never builds a chain (`TestSuccessionNeverBuildsChain`). A place-page step reads neighbour shells without their links (D-8).
- **Removed inputs.** A removed `part_of` link: its association is in `Handles` (§4) and both ends list its `{assoc}` input (AC-2.3), whether or not their nodes are loaded. A deleted place: in `Handles`, so its own entry drops; its links' associations are removed by its writer, so the entries at their other ends drop. A merged place: in `Handles` (its own entry); its `place_relationship` neighbours drop through the merge's moved `{neighbor, place_relationship}` inputs (AC-2.4 c). A changed link date or term: the association is in `Handles`. A new link: through its ends' `{end, place_relationship}` inputs (AC-2.4 b). A vocabulary change: `DerivedFrom` drops every entry (AC-2.5). An evicted key leaves the input index with it (AC-4.3).
- **Real data.** None — no rule, seed or table changes.
- **Untrusted size.** A hub's parts (20,000 in bar 2): loaded in `sqlbatch` chunks, measured (D-15); never copied into a row's keys (D-4). A `part_of` mesh's frontier: bounded by `summaryChainDepth` rounds and the walk's `seen` map, as today; a row's recorded keys stay under `inputKeyCap`. Entries held: `maxPlaceEntries`. A crafted commit with many `Handles`: one index lookup each, bounded by the write.

### Schema

None — no catalog change.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` `inputs.go` | `NodeKey`, `FollowKeys` (D-9); `Add` reports to `g.inputRecorder`, then records; `add` *(new)* records only; `Node` and `Follow` call `add`; `SetInputRecorderForTest` doc per D-10 |
| `core/database/graphcache` `follow.go` | `func (n *Node) Follow(hop hops.Hop) ([]Link, error)` *(new)*: "the node's links that match hop, in the node's order"; `Graph.follow` loads the node and returns `n.Follow(hop)`; `Graph.Follow` reports `FollowKeys` |
| `core/database/graphcache` `graph.go` | `Node` reports `NodeKey(id)` |
| `core/database/graphcache` `load.go` | `readNodes`, `readOpts`, `bridgeKeys`, `NodesOnBridges` *(new)*, doc: "reads ids' entity rows and values, and their links on the named bridges only, without caching them: Members and Outcomes are empty. A missing id is absent." `loadChunk` = `readNodes` + install; `linkRaws`, `linkLoadSQL`, `bridgePredicates` take `bridges []string` (D-2) |
| `core/database/graphcache` `cache.go` | `KeyedEntry.Inputs`; `keyedValues` gains `inputs map[string][]string` (key → its `(ID, Bridge)` index keys) and `byInput map[string]map[string]bool` (`(ID, Bridge)` → keys); `store`, `evict`, `Drop`, `Clear` keep them in step; `applyEntries(roots, set, moved)` drops by moved inputs, then runs Apply when set; `RegisterKeyed` panics only with neither (D-16) |
| `core/database/graphcache` `commit.go` | `OnCommit` collects the moved inputs (D-17); `mergedNeighbors` returns its links' `{neighbor, bridge}`; the no-handle path calls `applyEntries(nil, set, nil)` |
| `core/database/conclusionsummaries` `placeentries.go` *(new)* | `placeEntry`, `placeEntries`, `maxPlaceEntries`, `init` registration (D-1, D-5, D-6, D-12), `placeInputs`, `placeEntriesOf`, `fillPlaceEntries`, `checkPlaceEntry` |
| `core/database/conclusionsummaries` `compose.go` | `buildPlaceGraph` walks entries (D-7); `placeStep(c, id)` per D-8; `finishPlaces` loses its overwrite; `rememberShell`, `rememberPlaceLink` deleted; `assocDatesOf`, `placeShell` unchanged |
| `core/database/conclusionsummaries` `placechains.go` | `placeGraph{places map[string]placeEntry}`; `newPlaceGraph`; every method reads `places` (D-7); `AttachPlaceRelationships` loses its overwrite (D-8) |

Sketch:

```go
// graphcache
type KeyedEntry[T any] struct {
	Name        string
	Apply       func(g *Graph, k Keyed[T], roots [][]byte, set effects.Set) error // optional with Inputs
	Inputs      func(id []byte, v T) []InputKey                                     // a commit moving one drops the key
	DerivedFrom []Source
	Check       func(g *Graph, id []byte, stored T) error
	Cap         int
}

// conclusionsummaries
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
	Inputs:      placeInputs,
	Check:       checkPlaceEntry,
	Cap:   maxPlaceEntries,
})
```

Errors: none new besides the Check message (D-12) and `RegisterKeyed`'s panic text (AC-4.4).

### FFI

None — the Place detail and list wire are unchanged.

### macOS

None — no Swift change.

### Strings

None — no user-facing copy.

## 7. Steps

**STEP-1 — Measure the base**
Files: `core/database/promotealign/bench_test.go`, `core/database/conclusionsummaries/placeentries_test.go` *(new)*, `core/database/graphcache/graphcachetest/graphcachetest.go`.
Do: add `Canon.PartOf` (a `Canon.Link("place_relationship", {"from": part, "to": whole}, "place_relationship_type", "part_of")`); add `BenchmarkListPlaces` (D-14) and `BenchmarkPlaceMesh` (D-15), the latter in `package conclusionsummaries_test` (an internal test importing `graphcachetest` would cycle); both call only exported functions that exist at the base. On the planning base (no other change) run both and `BenchmarkCacheSizes`; write the figures into the ship log with `hostname`.
Check: `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench 'BenchmarkListPlaces|BenchmarkCacheSizes' -benchtime 1x -timeout 60m ./core/database/promotealign` prints both sizes; `CGO_ENABLED=1 go test -tags fts5 -run '^$' -bench BenchmarkPlaceMesh -benchtime 1x ./core/database/conclusionsummaries` prints its five sub-benchmarks.

**STEP-2 — The graph cache exports its keys, a node's Follow and a narrow load**
Files: `core/database/graphcache/inputs.go`, `follow.go`, `graph.go`, `load.go`, `commit.go` (`mergedNeighbors`' `linkRaws` call), `graph_test.go`, `inputs_test.go`.
Do: D-9, D-10, D-2 per §6. Add `TestNodesOnBridges` (T-1) and the `TestInputScope` subtest "the recorder hears Add" (T-2).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes; `rg -n "\bnodeKey\b|\bfollowKeys\b" core` finds nothing; `rg -n "&Node\{" core/database/graphcache --glob '!*_test.go'` finds one literal, in `readNodes`.

**STEP-3 — Keyed entries declare what they depend on**
Files: `core/database/graphcache/cache.go`, `commit.go`, `export_test.go`, `cache_test.go`. Skill: `use-catalog-session`.
Do: D-16, D-17 per §6. Add `InputIndexSizeForTest`, `TestKeyedInputs` (T-9) and `TestKeyedInputLinks` (T-10); extend `TestRegisterRejects` (AC-4.4).
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (every existing entry still has an Apply and no Inputs, and behaves as before), and `PROVENENCIA_GRAPH_VERIFY=1 CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/... ./core/database/canonlinks/...` passes.

**STEP-4 — The place entry**
Files: `core/database/conclusionsummaries/placeentries.go` *(new)*, `export_test.go`, `placeentries_test.go` *(new)*, created in STEP-1.
Do: D-1, D-3 – D-6, D-11, D-12. `fillPlaceEntries`: `NodesOnBridges(ids, "place_relationship")`; for each present id, `placeShell`, `periodWindow`, the four lists from `n.Follow` of the four place hops, each link oriented and dated with `assocDatesOf` over one `NodesOnBridges(assocIDs)`; `inputs` per D-4; `placeInputs` per D-5; a missing id per D-11. Add the four `export_test.go` hooks (§6 Codebase fit) and `TestPlaceEntryInputsMatchReads` (T-8), so a wrong fill shows here. No chain reads the entry yet.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/...` passes, `TestPlaceEntryInputsMatchReads` included.

**STEP-5 — Chains and the Places list read the entries**
Files: `core/database/conclusionsummaries/compose.go` (`buildPlaceGraph`, `finishPlaces`, `placeStep`), `placechains.go` (`placeGraph`, `newPlaceGraph` and its methods).
Do: D-7: `placeGraph{places}`, its methods, `buildPlaceGraph` over the entries, `finishPlaces` without its overwrite; the per-place scopes and node loads in `buildPlaceGraph` go. `closure`, `closureSize`, `reach`, `addChains`, `inputBudget` keep their behavior. Until STEP-6, `placeStep` keeps its node reads and writes them into `pg.places` through `rememberShell` / `rememberPlaceLink`, adapted to the new struct.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/...` passes, `TestInputsMatchIsolatedComposition`, `TestInputsHoldEveryShownParent` and the walks tests included.

**STEP-6 — The place page reads the entries**
Files: `core/database/conclusionsummaries/compose.go` (`placeStep`, `rememberShell`, `rememberPlaceLink`), `placechains.go` (`AttachPlaceRelationships`, `PartsAtDate`, `SuccessionNames`).
Do: D-8; delete what **Deletes** lists that is still left.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/... ./core/database/persondetail/... ./api/...` passes, and with `PROVENENCIA_GRAPH_VERIFY=1` the same command passes; `rg -n "rememberShell|rememberPlaceLink|\bbyID\b|partFrom\[|partTo\[|succFrom\[|succTo\[" core/database/conclusionsummaries --glob '!*_test.go'` finds nothing.

**STEP-7 — Tests for the entries**
Files: `core/database/conclusionsummaries/placeentries_test.go` *(new)*, created in STEP-1, `placechains_test.go` (`placeRel` returns the bridge Subject).
Do: add `TestChainsSkipLocationLinks` (T-3), `TestPlaceEntries` (T-4) and `TestPlaceEntriesFollowLinkMoves` (T-11), in `package conclusionsummaries_test`.
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestChainsSkipLocationLinks|TestPlaceEntr' ./core/database/conclusionsummaries` passes, and again with `PROVENENCIA_GRAPH_VERIFY=1`.

**STEP-8 — The cache table and the sizes benchmark know the entry**
Files: `docs/catalog-caches.md`, `core/catalogsession/cache_doc_test.go`, `core/database/promotealign/bench_test.go`.
Do: blank-import `core/database/conclusionsummaries` in `cache_doc_test.go`; in `BenchmarkCacheSizes` add a `measure("placeEntries", fill)` whose fill runs `conclusionsummaries.ListPlaces(c)`, placed before the `"displays"` measure; add the `conclusionsummaries.places` row (Owner `core/database/conclusionsummaries/placeentries.go` *(new)* · `placeEntries`; Key place handle id; Fill `placeEntriesOf`: a narrow load of the places' values and `place_relationship` links and their associations' values, a frontier at a time, through `Keyed.GetMany`; Updated by `DerivedFrom` the vocabulary and `Inputs` (its values, its links' associations and its `place_relationship` ends: a write of the place or a link, a new link at it, or a merge of a neighbour); Lifetime graph; Cap 65,536; size from STEP-9). The doc's rule line (line 5) becomes: "The signal is a field of `effects.Set` that the entry's Apply reads, a moved key among its declared `Inputs` (the `InputKey`s summary rows use), or a `DerivedFrom` declaration".
Check: `CGO_ENABLED=1 go test -tags fts5 -run 'TestCacheDocListsEveryEntry|TestCacheDocSeesEveryOwner' ./core/catalogsession` passes.

**STEP-9 — Measure after**
Files: `docs/catalog-caches.md` (size cell).
Do: run the STEP-1 commands on the same machine; write the after figures beside the base in the ship log; fill the size cell with `placeEntries-B` and `conclusionsummaries.places-keys` at 4,000 / 16,000.
Check: §11's measured stop conditions hold; `CGO_ENABLED=1 go test -tags fts5 -run TestLargeSourceBudgets ./core/database/promotealign` passes.

**STEP-10 — Docs**
Files: `.cursor/rules/catalog-queries.mdc`, `.agents/skills/review-pr/lenses/maintainability.md`, `docs/ideas/structural-debt.md`, `docs/deployment-plan/spike-16/cache-architecture.md`, `docs/deployment-plan/spike-16/deployment-plan.md`, `docs/deployment-plan/spike-16/completed.md`, this plan.
Do: `catalog-queries.mdc` per **Rules this plan bends** (the composer line, and "declares its Apply" → "declares its Apply or its `Inputs`"); the maintainability lens' Known trap "a drop rule written in `OnCommit` instead of an entry's Apply and DerivedFrom" → "a drop rule written in `OnCommit` instead of an entry's `Inputs`, Apply and DerivedFrom"; `structural-debt.md`: a row "`canonlinks.kinship` declares its pivots' `FollowKeys` as `Inputs` instead of `deps`, which walks cached neighbours and misses a removed link whose ends are not loaded"; `cache-architecture.md` § A cache a composer reads hands its inputs to the row: "(exported by S16-18, today the unexported `nodeKey` and `followKeys`)" → "(`graphcache.NodeKey` and `FollowKeys`)"; the map's working rule "A cache a composer reads hands its inputs to the row": "(S16-18 exports `nodeKey` and `followKeys`)" → "(`graphcache.NodeKey` and `FollowKeys`, S16-18)"; map row status `Landed #N`; `completed.md` paragraph; this plan's status and Landed note.
Check: the DoD sweep finds nothing; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/graph_test.go` `TestNodesOnBridges` *(new)* | `Canon`: City `PartOf` Region, an event located at City. Table: `NodesOnBridges([city], "place_relationship")` → one node, one link (to Region, `TermKey` `part_of`); `NodesOnBridges([city])` → no links; `NodesOnBridges([city], "place_relationship", "location")` → both links; `NodesOnBridges([missing], "place_relationship")` → absent; after each, `g.Cached(city)` is false; with every bridge key the node equals `g.Node(city)` but for `Members` and `Outcomes` | AC-1.3, D-2 |
| T-2 | Go | `core/database/graphcache/inputs_test.go` `TestInputScope` subtest "the recorder hears Add" *(new)* | A recording scope and a `NoInputs` scope each `Add(k)` → the recorder hears `k` from both; `scope.Node(person)` → heard once | D-10 |
| T-3 | Go | `core/database/conclusionsummaries/placeentries_test.go` `TestChainsSkipLocationLinks` *(new)* | The package fixture (`newFixture`, `placeSubject`, `placeRel`, `promoteSubject`, `lifeEvent`): Town ⊂ City ⊂ Region, three births located at City, one at Town, all promoted. (a) `f.C.Graph().Drop()`, `ParentsAtDate(town, today)` → `["City", "Region"]`; `Cached` of Town, City and each event is false. (b) `RowsForSearch(graphcache.Over(db), Want{Events: [townBirth]})` → its place `Town` with parents `["City", "Region"]`; `Cached(city)` is false on that graph | AC-1.1, AC-1.2 |
| T-4 | Go | `placeentries_test.go` `TestPlaceEntries` *(new)*, table-driven: each case builds the T-3 world, holds entries for Town, City, Region (one `ParentsAtDate(town)` after `Drop`), runs its write, then asserts | "person write": `f.named(person, "Bo")` → the places key count is unchanged and `ListPlacesQueryCount` after equals the warm count before; "city rename": a toponym "Big City" cited on City → `HeldPlaceEntryForTest(g, city)` false, Town's and Region's true; the chain's first name is `PlacesByIDs(f.C, [city])[0].Names[0]` and those `Names` contain "Big City"; "remove link": `placeRel` returns the bridge Subject (`res.Subject`, whether or not dates were added), deleted by `writes.Call(f.C, writes.Op{Action: "delete_subject", UserID: userID}, func(tx *database.Tx) (struct{}, []rowchange.Change, error) { ch, err := subjects.Delete(tx, userID, bridge.ID); return struct{}{}, ch, err })` → Town and City not held, Region held, chain `[]`; "vocabulary": `graphcachetest.NewCanon(t, f.C).CommitVocabulary()` → key count 0; "cap": `SetPlaceEntryCapForTest(g, 2)`, a chain of five places → `ParentsAtDate` has four names, key count ≤ 2; "corrupt": `CorruptPlaceEntryForTest(g, city)` → `g.Verify()` error contains "stored place entry"; "missing place": `ParentsAtDate(f.C, unknown, today)` for a random 16-byte id → `[]`, and `PlaceEntryForTest(g, unknown)` → held, `found` false, inputs `[{unknown, "", ""}]` | AC-2.1, AC-2.2, AC-2.3, AC-2.5, AC-2.6, AC-2.7, AC-3.2, D-5, D-11, D-12 |
| T-5 | Go | `core/database/conclusionsummaries/inputs_test.go` (unchanged) `TestInputsMatchIsolatedComposition`, `TestInputsHoldEveryShownParent`, `TestInputsCapCountsNothing`, `TestListsRecordNoInputs` | As today | AC-3.1, D-4, D-10 |
| T-6 | Go | `placechains_test.go` (unchanged), `walks_test.go`, `display_cache_test.go`, `eventplaces_cap_test.go`, `conclusionsummaries_test.go` `TestListPlacesQueryCountIsConstant` | As today, and under `PROVENENCIA_GRAPH_VERIFY=1` | D-7, D-8, D-13 |
| T-7 | Go | `bench_test.go` `BenchmarkListPlaces` *(new)*; `placeentries_test.go` `BenchmarkPlaceMesh` *(new)* | Reported, not gated (§11 holds the stops) | D-14, D-15 |
| T-8 | Go | `placeentries_test.go` `TestPlaceEntryInputsMatchReads` *(new)* | Per D-18 (a) and (b), for Town, City, Region and an unknown id → each list equals the scope's keys | D-4, D-5, D-18 |
| T-9 | Go | `core/database/graphcache/cache_test.go` `TestKeyedInputs` *(new)*, table-driven | Under `graphcache.SnapshotRegistryForTest`, a test entry with `Inputs` and no Apply, key A from `{A}`, `{X}` and key B from `{B}`; commits are `g.OnCommit(effects.Commit{Effects: effects.Set{Handles: ids}})` with ids that have no rows: `Handles` [X] → A gone, B held; [Y] → both held; [B] → B gone; overwrite A with `{A}`, `{Z}`, then [X] → A held, [Z] → A gone; Cap 1, store A then B → `InputIndexSizeForTest` counts only B's; `Drop`, then `Clear` → 0. `TestRegisterRejects` gains a case: neither Apply nor Inputs panics with "needs an Apply or Inputs" | AC-4.1, AC-4.3, AC-4.4, D-16 |
| T-10 | Go | `cache_test.go` `TestKeyedInputLinks` *(new)* | Under `SnapshotRegistryForTest`, a test entry whose values for event E and place P (held with `Set`) declare `{E, "location", "event"}` and `{P, "location", "place"}`, and `{L}` once a location L joins them; through `Canon`: `L := Location(E, P)`, `Commit(L)` → E and P gone; set again with `{L}`, a `Join(person, E, "subject")` committed → both held; `Remove(L)`, `Commit(L)` → E and P gone (through `{L}`); set again, a handle Q located at P merged (`Merge(Q, other)`, `Commit(Q, other)`) → P gone | AC-4.2, D-16, D-17 |
| T-11 | Go | `placeentries_test.go` `TestPlaceEntriesFollowLinkMoves` *(new)* | A `Canon` world on `graphcachetest.OpenCatalog`: Town `PartOf` City `PartOf` Region, entries held by `ParentsAtDate(town)`. The three cases of AC-2.4, each from a fresh world | AC-2.4, D-5, D-17 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `TestInputsMatchIsolatedComposition` and `TestPlaceEntryInputsMatchReads` (if D-4 or D-10 drifts), `TestEntryApply` (the new `applyEntries`), `TestAttachPlaceRelationships*` (D-8), `TestListPlacesQueryCountIsConstant` (a fill per row instead of per round), `TestCacheDocListsEveryEntry`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] The conclusionsummaries, graphcache, persondetail and api suites pass with PROVENENCIA_GRAPH_VERIFY=1
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left — none in this PR
- [ ] Sweep: `rg -n "\bnodeKey\b|\bfollowKeys\b|rememberShell|rememberPlaceLink|NodesOnBridge\(" core .agents .cursor/rules docs/*.md` finds nothing; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated (catalog-queries.mdc, catalog-caches.md, the maintainability lens)
- [ ] UI matches the brief inventory — no UI in this PR
- [ ] Version bump: none
- [ ] Ship log: STEP-1 and STEP-9 figures with hostname, before and after
- [ ] docs/catalog-caches.md: the conclusionsummaries.places row with measured sizes, and the rule line naming Inputs
- [ ] docs/ideas/structural-debt.md: the canonlinks.kinship Inputs row
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
| An entry's keys drift from what a scope records (D-4) | T-8 compares them with real reads; T-5 holds batch against single-row composition |
| The input index falls out of step with eviction | T-9 counts it after every mutation; Verify's Check rebuilds held values |
| The always-narrow fill makes the first read after open slower | `BenchmarkListPlaces` cold at both sizes; stop below |
| A commit that removes a link reaches neither end through `roots` | The association is in `Handles` and both entries name it (D-5, D-16); T-4 "remove link" holds it |
| A merge leaves a neighbour's entry linked to a merged place | The merge's moved `{neighbor, place_relationship}` inputs drop it (D-17); T-11 (c) holds it |

**Stop and report if:**

- A Current state row is false at your base (for example `buildPlaceGraph` no longer walks nodes, or `OnCommit` no longer applies entries after its reload).
- `TestInputsMatchIsolatedComposition` fails after STEP-4 and the difference is a key D-4 does not list (report the key).
- In T-4 "remove link" the bridge Subject's association is not in the write's `Handles` (D-16's premise would be wrong).
- A Verify run reports a stale place entry after any write in the suites (a change D-5 does not cover: report the write).
- An existing entry's test changes result after STEP-3 (the registry change must not change an entry that declares no `Inputs`).
- `BenchmarkListPlaces` cold at 16,000 persons is more than 50 ms slower after than before, or warm is more than 10 % slower.
- `BenchmarkPlaceMesh` after a location at the hub is more than 1.25 × slower after than before, comparing the median of three `-benchtime 1x` runs on each side.
- `BenchmarkPlaceMesh` cold takes 3 s or more after (any of three `-benchtime 1x` runs). Same reason as the hub commit below: the crafted hub only has to stay bounded (user, 2026-11-13, after round 0 measured medians 1865 and 1974 ms against a 1362 ms base: the cold fill reads the hub's 20,000 association value sets through `NodesOnBridges`).
- `BenchmarkPlaceMesh` after the hub commit takes 1 s or more after. The crafted hub only has to stay bounded (the map's **Scale**), so this row has an absolute ceiling, not a ratio (user, 2026-11-13, after round 0 measured 560–668 ms against a 142 ms base: the refill reads the hub's 20,000 association value sets, see **Performance** "Hub entry").
- `BenchmarkCacheSizes` `placeEntries-B` exceeds 5 MB at 16,000 persons.
- `TestLargeSourceBudgets` fails a subtest it passed in STEP-1.

## 12. Open questions

None.

## 13. PR

**Title:** S16-18 — Place entries
**Body:** objective in one sentence; what shipped against US-1 – US-4; AC evidence; deviations from the plan; the STEP-1 and STEP-9 figures; manual check; `Plan: docs/deployment-plan/spike-16/prs/S16-18-place-entries.md`.
