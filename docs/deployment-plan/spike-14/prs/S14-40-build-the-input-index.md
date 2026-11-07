# S14-40 — Build the input index

**Status:** Draft
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

Every person, event and place summary row records which other handles, links and associations it was composed from, and the catalog stores that as an index beside the search documents. Nothing reads the index yet. S14-42 switches write invalidation to it and deletes the hand-written inverse (`conclusionsummaries.Dependents` and its neighbours). This PR proves the index is right. It equals what composing each row alone reads, it covers everything `Dependents` finds today, and it equals a rebuild after every write the upkeep tests make. The PR also removes a quadratic cost from search reprojection: one graph per call instead of one per 500-id batch.

**Why now.** First PR of Batch 8 (user, 2026-11-07); S14-38's stale documents are fixed on top of it in S14-42.

**What stays.** Every summary row, list and search document reads as before. Invalidation, `Dependents`, `AssociationEndpoints` and `subjects.Delete` are untouched.

## 2. User stories

**US-1.** As the engine, each summary row knows its inputs, so a later write can find it by lookup.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Any person, event or place in the scenario fixtures (§8 T-2) or the `TestLargeSourceBudgets` fixture, composed in one batch through `conclusionsummaries.RowsForSearch` | `Summaries.Inputs(id)` equals the input keys recorded when that row alone is composed with the test recorder on the graph, minus the row's own values |
| AC-1.2 | James (person) with a birth event located at Yorkville, part_of Toronto, part_of Ontario | James's inputs hold `(birth, values)`, `(birth, location, event)`, the location's `(A, values)`, and `(Yorkville, values)`, `(Toronto, values)`, `(Ontario, values)` with their `place_relationship` / `from` link keys and each part_of association's values. They do not hold `(James, values)` |
| AC-1.3 | Any handle X in those fixtures | The rows whose inputs hold `(X, values)` or a link key of X include every id `conclusionsummaries.Dependents([X])` returns, except X's immediate part_of parents and its succession neighbours (detail-page hops). The extras are counted and logged |

**US-2.** As the engine, the index is stored and kept with the search documents.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Any write in `TestRebuildEqualsUpkeep_Scenarios` or `_SeededSequences` | `summary_inputs` equals the table `searchindex.RebuildAll` writes (both sorted) |
| AC-2.2 | A filed participation A between James and a birth | `summary_inputs` holds `(A → James, participation, person)` and `(A → birth, participation, event)` |
| AC-2.3 | A handle's document is dropped (deleted, merged) | Its rows as a dependent are gone after that `ReprojectHandles` |
| AC-2.4 | A catalog whose `catalog_search_meta.projection_version` is 10 | After `searchindex.EnsureCatalog`, `summary_inputs` holds rows and the version is 11 |

**US-3.** As the engine, search reprojection reads each handle's graph once per call.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `ReprojectHandles` of 1,200 ids (3 batches) | One graph is built (`newGraph` called once) |
| AC-3.2 | The first part_of filed from the 6,000-event city (`TestLargeSourceBudgets` "second part_of from the city") | Faster than before; planning measured 3.64 s → 1.63 s on one machine. The PR reports both numbers |
| AC-3.3 | Composing every person and every event row on the large fixture | Under `budgetComposeAll` (500 ms); planning measured 95 ms + 125 ms before recording |

## 3. Scope

| In | Covered by |
| --- | --- |
| `graphcache.InputKey`, `InputScope`, `Graph.Inputs` and a test recorder hook | STEP-2, D-1, D-2, D-10 |
| Composers attribute inputs per row; `placeGraph` replays its chain closure per row; `Summaries.Inputs` | STEP-3, D-3 – D-5 |
| Migration `000062.sql`: table `summary_inputs`, registered in `catalogmodel` | STEP-4, D-6 |
| `searchindex` writes, clears and rebuilds the table; association ends; `ProjectionVersion` 11 | STEP-5, D-7, D-8 |
| `ReprojectHandles` composes on one graph per call | STEP-5, D-9 |
| Rebuild-equals-upkeep compares the table | STEP-6, AC-2.1 |
| Oracle, coverage and budget checks on the large fixture | STEP-7, AC-1.1, AC-1.3, AC-3.2, AC-3.3 |
| Rule and skill lines; design and debt updates | STEP-8 |

| Out | Home |
| --- | --- |
| Invalidation from the index; deleting `Dependents`, `AssociationEndpoints`, `linkEnds`, `movedRoots`, `SetDisplayDeps`; `subjects.Delete` | S14-42 |
| The `core/database/graphcache/sql.go` copy of `database.SQLInPlaceholders` | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md) (the S13-37 row): `core/database` imports `graphcache`, so `graphcache` cannot import `database`; the helper has to move to a leaf package first. the row and the map say so (planning, 2026-11-07) |
| Recording inputs for the kinship memos (`canonlinks`) | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the "Spike 14 Batch 8 design" row |
| Narrowing a person's place graph to its birth and death events (it reads the places of all its events today) | not planned: an over-read costs a rebuild, never a miss; the coverage check logs it |
| The remaining cost of the first part_of filed from a city after the one-graph fix | [`docs/ideas/structural-debt.md`](../../../ideas/structural-debt.md), the "first `part_of` filed from a city" row, updated with the new number in STEP-8 |

## 4. Current state

Verified at `26ada55`.

| Area | Where | Today |
| --- | --- | --- |
| Graph reads | `core/database/graphcache/graph.go` `Graph`, `Node`, `Nodes`, `Over`; `follow.go` `Follow` | `Node` loads and returns a node; `Nodes` prefetches; `Follow(id, hop)` calls `g.Node(id)` and filters `n.Links` by the hop's bridge, end and term. No recording. `Graph` has no lock: callers are serialized by the catalog session |
| Test hooks | `graphcache.SetQueryCounterForTest` (graph.go), `identityclaims.SetReadCounterForTest` | Precedents for a package-level test hook |
| Person rows | `core/database/conclusionsummaries/compose.go` `composePersons` (line 547), `onePerson`, `lifePicks`, `leadLife`, `eventPlaceGraph`, `leadPictures` | Batch phase: `follow(EventsOfSubject)` per person, `g.Nodes(events)`, `eventPlaceGraph` (`Node(e)` and `follow(e, PlacesOfEvent)` for every subject event, then `buildPlaceGraph`), `leadPictures` (raw SQL through `gallery.ForEntities`). Per row: `onePerson` → `lifePicks` (`follow`, `Node` of each event) → `leadLife` → `eventPlaces` |
| Event rows | `composeEvents` (line 865), `oneEvent`, `eventPlaces` | Batch: `follow(SubjectsOfEvent)`, `follow(PlacesOfEvent)`, `g.Nodes`, `buildPlaceGraph(places)`, `leadPictures`. Per row: own values, `eventPlaces` (`follow`, `Node(place)`, `pg.FoldLocations`), subjects (`follow`, `Node(person)`) |
| Place rows | `composePlaces` (line 1141), `placeShell`, `finishPlaces` | Shell from own values; `finishPlaces` builds one `placeGraph` from all rows and sets `Parents` per row through `pg.ParentChain` |
| Place chains | `buildPlaceGraph` (line 1200), `rememberPlaceLink`; `placechains.go` `placeGraph` (`byID`, `partFrom`, `partTo`, `succFrom`, `succTo`, `period`), `linkHolds` | BFS over `ParentsOfPlace` for `summaryChainDepth` (8) rounds from the seeds: per frontier place, `Node`, `placeShell`, `follow`, and `Node(assoc)` per link (its dates). Places first reached in the last round are not read |
| Summaries | `compose.go` `Summaries`, `NewSummaries`, `RowsForSearch` (line 407), `persons`, `events`, `places` | `RowsForSearch` composes with `store` false; its only caller is `searchindex` |
| Dependents | `core/database/conclusionsummaries/dependents.go` `Dependents` | Hand-written inverse; returns a place's parents and succession neighbours too |
| Reprojection | `core/database/searchindex/handles.go` `ReprojectHandles` (line 42), `reprojectHandleBatch`, `loadHandleSummaries` | `sqlbatch.Each` in batches of 500 (`sqlbatch.Size`); each batch calls `RowsForSearch(graphcache.Over(q), w)`, a fresh graph, so a place shared by many events is reloaded with all its links per batch. Associations have `kind == ""` and are skipped. A gone id drops its documents |
| Rebuild | `searchindex.go` `ClearAll` (line 245), `RebuildAll`, `ProjectionVersion` (10), `NeedsRebuild`, `EnsureCatalog` (line 643) | Version mismatch rebuilds on open |
| Catalog model | `core/database/catalogmodel/model.go` | `catalog_search_docs`, `catalog_search_meta`, `auto_reconciler_*` are `BucketSkip`; `auto_reconciler_values.entity_id` is `ON DELETE CASCADE` (`000034.sql:13`) |
| Migrations | `core/database/migrations/` | Last step `000061.sql`. New tables are `STRICT` (`evidence_graphs`) |
| Upkeep tests | `core/database/autoreconciler/autoreconciler_test.go` `(*fixture).assertUpkeepEqualsRebuild` (line 1103), `handleDocs` | Compare value rows and person / event / place documents with a rebuild in a rolled-back transaction |
| Rebuild test | `core/search/handles_test.go` `TestRebuildIndexesHandles` (line 343) | Sets the projection version back and checks open rebuilds documents |
| Perf fixture | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets`, `newLargeFixture`, `fastest`, `over`, subtest "second part_of from the city" (the first part_of is untimed, line 304) | 8,115 persons, 10,067 events, 51 places, 24,208 associations |
| Composer fixtures | `core/database/conclusionsummaries/walks_test.go` `participation`, `location`, `named`, `at`, `promoteSubject`; `placechains_test.go` `placeSubject`, `placeRel`; `display_cache_test.go` `TestDisplayCacheFollowsAPlaceThreeStepsUp` (the Country / Region / City / Town world) | Reused by T-2, T-3 |

**Planning probes** (throwaway, reverted): recording `Node` / `Follow` while composing each row alone gave 96,676 entries, median 5 per row, identical to `Dependents` on 300 sampled handles at the coarser key. Composing per-row place graphs instead of the batch one cost 4× (persons 95 → 386 ms, events 125 → 533 ms). One graph per `ReprojectHandles` call took the first part_of from the city from 3.64 s to 1.63 s.

**Precedent.** `graphcache.RegisterNodeMemo` (memo.go) is the shape of an owner-agnostic `graphcache` API used by `canonlinks`. `searchindex` writing a derived table in the reprojection transaction is `catalog_search_docs` itself. `TestRebuildIndexesHandles` is the shape of AC-2.4.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `graphcache.InputKey` *(new)* `struct { ID []byte; Bridge, Endpoint string }`. `Bridge == ""` is the handle's values; otherwise `Bridge` is the bridge type key and `Endpoint` the endpoint Property key where `ID` sits (`hop.Bridge()`, `hop.From()`). An association's values are `{A, "", ""}` | The three kinds of the working rule in one comparable shape; product bridges only today (`connectrules`), so a key carries no origin | A kind enum — the empty bridge already says it |
| D-2 | `func (g *Graph) Inputs(dependent []byte) *InputScope` *(new)*. `InputScope` has `Node(id)`, `Nodes(ids)`, `Follow(id, hop)`, `Graph() *Graph` and `Keys() []InputKey` (deduplicated, sorted by ID, Bridge, Endpoint). `Node(id)` records `{id, "", ""}` unless `id` equals the dependent, including for a missing node. `Follow(id, hop)` records `{id, hop.Bridge(), hop.From()}` and `{l.Association, "", ""}` for each link it returns, and nothing for `id`'s values. `Nodes` records nothing. A scope from `g.Inputs(nil)` records nothing. `Graph.Node`, `Nodes` and `Follow` are unchanged; recording lives in the scope value, not in graph state | An explicit value per row cannot be misattributed by interleaved calls, and the next users (kinship memos) take the same API |  A "current dependent" field on `Graph` — shared mutable state on a graph many callers hold |
| D-3 | The per-row compose functions take `*graphcache.InputScope` where they read through `*graphcache.Graph`: `onePerson`, `lifePicks`, `leadLife`, `eventPlaces`, `oneEvent`. `PersonLifeOwners` and `lifeOwners`, which share `eventPlaces`, pass `g.Inputs(nil)`. Batch phases keep `*graphcache.Graph` | One code path per row; only the batch prefetch stays unscoped | Recording in batch phases — cannot attribute |
| D-4 | `placeGraph` gains `closure(id []byte) []InputKey` *(new)*, memoized per place in the graph: a BFS from `id` over `partFrom` for `summaryChainDepth` rounds, matching `buildPlaceGraph`'s rounds. Each place expanded in rounds 0 to 7 gives `{p, "", ""}`, `{p, "place_relationship", "from"}`, and `{assoc, "", ""}` for each part_of link out of it. Places first reached in the last round give nothing. A scope records a closure through `InputScope.Add(keys ...InputKey)` *(new)*, which applies the self rule. Replay per row: a person records `Node(e)` and `Follow(e, hops.PlacesOfEvent)` for each of its subject events, then the closure of each place those links reach. An event records the closure of each of its places. A place records its own closure in `finishPlaces` | Matches what composing the row alone reads (the oracle). A place within 7 rounds of a row's seed is within 7 rounds of the batch's nearest seed, so the batch `placeGraph` holds every link the closure walks | Per-row `placeGraph` — 4× slower (§4 probes) |
| D-5 | `Summaries` gains `inputs map[string][]graphcache.InputKey` and `func (h Summaries) Inputs(id []byte) []graphcache.InputKey` *(new)*. `persons`, `events` and `places` return each composed row's keys beside the rows. `RowsForSearch` fills `Summaries.inputs`; the list and `ByIDs` paths drop them | The search reprojector is the index's only writer (working rule) and `RowsForSearch`'s only caller | Recording only when asked — two code paths through the composers |
| D-6 | Migration `core/database/migrations/000062.sql` *(new)*: `CREATE TABLE summary_inputs (dependent_id BLOB NOT NULL, input_id BLOB NOT NULL, bridge TEXT NOT NULL DEFAULT '', endpoint TEXT NOT NULL DEFAULT '', PRIMARY KEY (dependent_id, input_id, bridge, endpoint)) STRICT, WITHOUT ROWID;` and `CREATE INDEX summary_inputs_input_idx ON summary_inputs (input_id, bridge, endpoint);`. No foreign key. `catalogmodel`: `{Name: "summary_inputs", Bucket: BucketSkip}` | Working rule: no CASCADE, or S14-38 returns. The primary key serves the per-dependent delete, the index serves S14-42's lookup | A foreign key with CASCADE, as `auto_reconciler_values` has — deletes the rows S14-42 needs |
| D-7 | `core/database/searchindex/inputs.go` *(new)* owns the table: `writeInputs(q Querier, ids [][]byte, heads map[string]handleHead, s conclusionsummaries.Summaries) error` deletes the rows of every id in the batch as a dependent, then inserts, for a person / event / place handle that is not merged, `s.Inputs(id)`, and, for an association handle (its type is a `connectrules` bridge, `connectrules.LookupBridge(typeKey)`; `handleHead` gains a `typeKey` field, set where `reprojectHandleBatch` scans the heads), one row per kept entity end: `(A, end, bridge key, endpoint Property key)`, read from `auto_reconciler_values` (`rank = 1`, `reason = 'kept'`, `value_entity_id` not null) joined to `properties` for the key. `reprojectHandleBatch` calls it after the documents. `ClearAll` adds `DELETE FROM summary_inputs` | One package writes, clears and rebuilds the table, in the documents' transaction | Writing association ends from `autoreconciler` — a second writer outside `ClearAll` |
| D-8 | `ProjectionVersion` becomes 11, with the comment line `11: summary inputs beside the handle documents (S14-40).` | An existing catalog fills the table on open, through the existing `NeedsRebuild` → `RebuildAll` path | A backfill migration — the projector is Go |
| D-9 | `ReprojectHandles` builds one graph per call through `var newGraph = graphcache.Over` *(new, package var)* and passes it to every batch: `reprojectHandleBatch(q, g, ids)`, `loadHandleSummaries(g, heads)`. `core/database/searchindex/export_test.go` *(new, package `searchindex`)* exposes `SetNewGraphForTest(func(graphcache.Querier) *graphcache.Graph) (restore func())` so T-4 counts calls; T-4 and T-5 live in `inputs_test.go` *(new)* as package `searchindex_test`, because they build catalogs through `promote` and `connect`, which an in-package test cannot import without a cycle | One graph per call is safe: a reprojection writes documents, not the value rows the graph caches. The package var is the smallest seam for a deterministic test | A timing test alone — machine-dependent |
| D-10 | `graphcache.SetInputRecorderForTest(fn func(InputKey))` *(new)*. When set, the exported `Graph.Node` and `Graph.Follow` report keys by D-2's rules, with no self rule and no scope. `Follow` reaches its node through an unexported lookup, so it does not also report the node's values | The oracle needs the reads of a row composed alone, independent of the scope code it checks | Running the oracle through scopes — it would check the code with itself |
| D-11 | The coverage check (AC-1.3) is a superset check. For X, the index's dependents of X are every `dependent_id` whose inputs hold `(X, values)` or any `(X, bridge, endpoint)`. They must include `Dependents([X])` minus X's immediate part_of parents (`hops.ParentsOfPlace`, one hop) and X's succession neighbours (`SuccessorsOfPlace`, `PredecessorsOfPlace`, one hop). Extras are logged as a count per kind, not failed | S14-42 must invalidate at least what today's walk does. A person reads the places of all its events, so the index holds more than `Dependents` (Out row 4) | An equality check — fails on correct over-reads |
| D-12 | `budgetComposeAll` *(new)* = 500 ms in `perf_test.go`, for composing every person and every event row once, warm, fastest of 3 | The planning measure was 220 ms before recording. 500 ms allows recording plus CI and catches a per-row place graph | No budget — a 4× regression would pass |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `graphcache.InputKey`, `InputScope`, `Graph.Inputs`, `InputScope.Add` *(new)* | type / func | `graphcache.NodeMemo` / `MemoDeps` (a hand-written inverse per slot) | The memo API stores results, not inputs; the scope is what `NodeMemo` owners use later (structural-debt row) |
| `graphcache.SetInputRecorderForTest` *(new)* | test hook | `SetQueryCounterForTest` | That one is per graph and counts queries; the oracle needs keys from the graph `RowsForSearch` builds inside `searchindex` |
| `placeGraph.closure` *(new)* | func | `buildPlaceGraph` (the same BFS, forward) | It reads `pg`'s maps, not the graph; it replays per row what the batch build read once |
| `Summaries.Inputs` *(new)* | func | `Summaries.Person` / `Event` / `Place` | Same struct, one more lookup |
| `summary_inputs` *(new)*, `000062.sql` *(new)* | table | `catalog_search_docs` | A different shape (many rows per dependent); same owner and bucket |
| `searchindex/inputs.go` `writeInputs` *(new)* | file / func | `searchindex.Upsert`, `Delete` | Per-dependent replace of many rows; documents stay as they are |
| `searchindex.newGraph` *(new)* | package var | `graphcache.Over` (called inline today) | The seam T-4 counts |
| `budgetComposeAll` *(new)* | const | `budgetLocateAtCity` | One more row |

**Builds on.** `graphcache` as the shared read layer; `searchindex`'s ownership of derived search data and its version-driven rebuild; the rebuild-equals-upkeep guard.
**Deletes.** The per-batch `graphcache.Over(q)` in `loadHandleSummaries`. Nothing else becomes dead in this PR; S14-42 deletes the old inverse.
**Next adds.** S14-42 looks dependents up through `summary_inputs_input_idx`. The kinship memos record through `InputScope`. A new handle kind's composer records through its scope (the `add-searchable-kind` line).

### Reach

None — no new path into existing code reaches a researcher: the index is written and not read, and `RowsForSearch` returns the same rows.

### Scale

At the map's Scale (a project of 3,000 Sources, 37,000 Subjects and 92,000 Observations; 6,000 events on one place), measured on the `TestLargeSourceBudgets` fixture:
- **Composition** gains one small key set per row (median 5 keys) and one memoized closure per distinct place (51 places on the fixture). `budgetComposeAll` holds it (D-12).
- **Writes.** Each reprojected document replaces its dependent's few rows: one `DELETE … WHERE dependent_id IN (batch)` and the inserts. A rebuild writes about 97,000 rows on the fixture, a few hundred thousand at the map's Scale.
- **Reprojection** of N handles builds one graph, not N/500: a place's links load once.

### Edges

- **Assumptions.** Composition is deterministic for a fixed catalog, so isolated and batch reads agree; T-2 and T-7 hold it. `buildPlaceGraph`'s round count is `summaryChainDepth`; `closure` reads the same constant. A cached node is valid for the whole `ReprojectHandles` call, because a reprojection writes no value rows.
- **Neighbors.** A filtered hop (`EventsOfSubject`, role subject; `ParentsOfPlace`, part_of) records the bridge-end key, which covers the links the filter drops (working rule). A missing node read through `Node` is still recorded (T-1), so its later creation reaches the row. A succeeded_by link is not followed by any row composer, but its bridge-end key equals part_of's; the coverage check excludes succession hops, as D-11 says.
- **Removed inputs.** A deleted or merged dependent loses its rows when its document is dropped (AC-2.3). A deleted input keeps its dependents' rows until they are reprojected; that is what S14-42 reads.
- **Not handle inputs.** Vocabulary labels (`loadVocab`), the name format profile (`nameformats.LoadFor`, sort only), the alias index, and the row's own raw reads (`gallery.ForEntities`; the document's `sqlHandleHeads`, `sqlHandleValues`) are not recorded. The working rule "A row is its own input" covers the self-reads in S14-42.
- **Real data.** None — no table- or seed-driven rule.
- **Untrusted size.** None — the keys come from catalog rows the app wrote; per row they are bounded by its links and the chain depth (8).

### Schema

Migration `core/database/migrations/000062.sql` *(new)*, per D-6, with a leading comment: `-- S14-40: summary_inputs records the inputs of each person, event and place summary (and each association's ends) so a write can find the rows it moved. Owned by searchindex; no foreign key, so a deleted input keeps its dependents' rows until they are reprojected.`

Registered in `catalogmodel` as `BucketSkip`, with no FK rows. Delete class: none (derived). Audit: none.

### Go core

| Package | Change |
| --- | --- |
| `core/database/graphcache` | `inputs.go` *(new)*: `InputKey`, `InputScope`, `Graph.Inputs`, `Add`, `Keys`, `SetInputRecorderForTest`. `graph.go` `Node` and `follow.go` `Follow` call the recorder when set; `Follow` uses an unexported node lookup |
| `core/database/conclusionsummaries` | `compose.go`: per-row functions on `*graphcache.InputScope` (D-3); replay per D-4; `Summaries.inputs`, `Inputs` (D-5). `placechains.go`: `placeGraph.closure` with its memo |
| `core/database/searchindex` | `inputs.go` *(new)* `writeInputs`; `handles.go`: `newGraph`, one graph per call, `writeInputs` call; `searchindex.go`: `ClearAll`, `ProjectionVersion` 11 |
| `core/database/catalogmodel` | `summary_inputs` row |

Errors: none new.

### FFI

None — no wire change.

### macOS

None — no app change.

### Strings

None — no copy.

## 7. Steps

**STEP-1 — Budget first**
Files: `core/database/promotealign/perf_test.go`.
Do: add `budgetComposeAll` (D-12) and the subtest "compose every summary" (T-8) on today's code, and time and log the first part_of in "second part_of from the city" (`t.Logf`). Log the timings.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` passes; record the logged times in the PR description as the before figure.

**STEP-2 — The input scope**
Files: `core/database/graphcache/inputs.go` *(new)*, `graph.go`, `follow.go`, `core/database/graphcache/inputs_test.go` *(new)*.
Do: D-1, D-2, D-10.
Check: T-1 passes; `CGO_ENABLED=1 go test -tags fts5 ./core/database/graphcache/...` passes.

**STEP-3 — Composers attribute inputs per row**
Files: `core/database/conclusionsummaries/compose.go`, `placechains.go`, `core/database/conclusionsummaries/inputs_test.go` *(new)*.
Do: D-3, D-4, D-5. Write T-2 and T-3 first, with the scenario worlds of §8.
Check: T-2 and T-3 pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/conclusionsummaries/...` passes, with no existing test changed.

**STEP-4 — The table**
Files: `core/database/migrations/000062.sql` *(new)*, `core/database/catalogmodel/model.go`. Skills: `add-catalog-migration`, `add-catalog-model`.
Do: D-6.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/...` passes (schema hash, `TestPragmaHonesty`).

**STEP-5 — The reprojector writes the index**
Files: `core/database/searchindex/handles.go`, `inputs.go` *(new)*, `searchindex.go`, `core/database/searchindex/export_test.go` *(new)*, `core/database/searchindex/inputs_test.go` *(new)*, `core/search/handles_test.go`. Skill: `add-catalog-query`.
Do: D-7, D-8, D-9.
Check: T-4, T-5 and T-6 pass; `CGO_ENABLED=1 go test -tags fts5 ./core/database/searchindex/... ./core/search/...` passes.

**STEP-6 — Upkeep equals rebuild, for the index too**
Files: `core/database/autoreconciler/autoreconciler_test.go`.
Do: `assertUpkeepEqualsRebuild` also compares `inputRows(db)` with the rows after `searchindex.RebuildAll` in its rolled-back transaction. `inputRows` *(new test helper)* is every row of `summary_inputs` as `hex(dependent)|hex(input)|bridge|endpoint`, sorted.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/autoreconciler/` passes (T-7).

**STEP-7 — At scale**
Files: `core/database/promotealign/perf_test.go`.
Do: T-9 and T-10 as subtests of `TestLargeSourceBudgets`. The first part_of log from STEP-1 runs again.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/promotealign/ -run TestLargeSourceBudgets -v` passes. Every budget is under its limit, and the logged first part_of is below the STEP-1 figure.

**STEP-8 — Rules, skill and docs**
Files: `.cursor/rules/catalog-queries.mdc`, `.agents/skills/add-searchable-kind/SKILL.md`, `core/database/searchindex/handles.go` (header comment), `docs/deployment-plan/spike-14/summary-inputs.md`, `docs/ideas/structural-debt.md`.
Do:
- `catalog-queries.mdc` gains the rule "A summary composer reads other handles only through a `graphcache.InputScope` (`Node`, `Follow`) so its inputs are recorded; a raw query in a composer reads only the row's own handle."
- The `add-searchable-kind` checklist gains "A handle kind's composer records its inputs through `graphcache.InputScope`."
- The `handles.go` header names `summary_inputs`.
- The design doc's "Measured" section gains the final key counts from T-9.
- In `structural-debt.md`, the "first `part_of` filed from a city" row gets the new figure.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-14/prs/S14-40-build-the-input-index.md` reports no errors; `scripts/check-pr.sh` prints `check-pr: PASS`.

**STEP-9 — Landing docs**
Files: `docs/deployment-plan/spike-14/completed.md`, `deployment-plan.md` (map row), this plan (status and Landed note).
Do: per the Definition of done.
Check: links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/graphcache/inputs_test.go` `TestInputScope` *(new)* | table on a `graphcachetest` world (a person, a birth, a participation with role subject, a witness participation): `Node(other)` → `{other}`; `Node(self)` → none; `Node(missing id)` → `{missing}`; `Follow(person, EventsOfSubject)` → `{person, participation, person}` plus each returned association, and not `{person}`; `Nodes` → none; `g.Inputs(nil)` → none; with the recorder set, `Graph.Follow` reports no node values | D-1, D-2, D-10 |
| T-2 | Go | `core/database/conclusionsummaries/inputs_test.go` `TestInputsMatchIsolatedComposition` *(new)* | worlds: Country / Region / City / Town with a birth at Town (as `TestDisplayCacheFollowsAPlaceThreeStepsUp`); a part_of with link dates that exclude it at the event's date; York succeeded_by Toronto; a witness participation; an event at two places → for every person, event and place, `RowsForSearch` (batch over all) `Inputs(id)` equals the recorder's keys from `RowsForSearch(g, Want{one id})` minus `{id}` | AC-1.1, AC-1.2, D-4 |
| T-3 | Go | same file `TestInputsCoverDependents` *(new)* | the same worlds; per D-11 for every handle → no missing id; extras logged | AC-1.3, D-11 |
| T-4 | Go | `core/database/searchindex/inputs_test.go` `TestReprojectHandlesBuildsOneGraph` *(new)* | 1,200 ids (3 batches) → `newGraph` called once | AC-3.1, D-9 |
| T-5 | Go | same file `TestReprojectWritesInputs` *(new)* | a filed participation A between James and a birth: after `ReprojectHandles` → the birth's rows hold `{James}`; A's rows are `(A → James, participation, person)`, `(A → birth, participation, event)`; after the birth's handle is deleted and reprojected → no row with dependent birth; after `ClearAll` → table empty | AC-2.2, AC-2.3, D-7 |
| T-6 | Go | `core/search/handles_test.go` `TestRebuildIndexesSummaryInputs` *(new)* | as `TestRebuildIndexesHandles`: set version 10, `EnsureCatalog` → table not empty, version 11 | AC-2.4, D-8 |
| T-7 | Go | `core/database/autoreconciler/autoreconciler_test.go` every `TestRebuildEqualsUpkeep_*` and `assertUpkeepEqualsRebuild` caller | upkeep table equals rebuild table | AC-2.1 |
| T-8 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "compose every summary" *(new)* | every person and every event row composed, fastest of 3 → under `budgetComposeAll` | AC-3.3, D-12 |
| T-9 | Go | same, "inputs match isolated composition" *(new)* | every person, event and place on the large fixture → T-2's equality; logs key count, median and max per row | AC-1.1 |
| T-11 | Go | `core/database/promotealign/perf_test.go` `TestLargeSourceBudgets` "second part_of from the city" (the STEP-1 log line) | the first part_of from the city, timed → logged; the STEP-7 figure is below the STEP-1 figure (Stop condition), both in the PR description | AC-3.2 |
| T-10 | Go | same, "inputs cover Dependents" *(new)* | 300 handles, the first 100 of each kind by id → T-3's check; logs extras per kind | AC-1.3 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/conclusionsummaries` (signatures), `core/search` (document tests around reprojection), the `TestLargeSourceBudgets` budgets, `catalogmodel` honesty.

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
- [ ] Domain doc updated if a rule in docs/*.md changed: none; summary-inputs.md, catalog-queries.mdc, add-searchable-kind and structural-debt.md per STEP-8
- [ ] PR description: what shipped, AC evidence, before and after timings (STEP-1, STEP-7), deviations
```

## 10. Manual check

None — the index is not read by anything a researcher can see. The checks are T-1 – T-10 and the timings in the PR description.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The replay records less or more than the isolated row reads | T-2 and T-9 fail on any difference |
| Recording slows the lists | T-8 budget |
| The index drifts from documents after some write | T-7 on every upkeep scenario and seed |

**Stop and report if:**

- T-2 or T-9 finds a difference that is not a replay bug: a composer reads another handle outside `Node` / `Follow` (a raw query, a memo, `Kind`). Name the read; it needs a decision.
- T-3 or T-10 finds an id `Dependents` returns that the index lacks, outside D-11's exclusions.
- `budgetComposeAll` fails, or any existing `TestLargeSourceBudgets` budget fails after STEP-3 or STEP-5.
- The first part_of in STEP-7 is not below the STEP-1 figure.
- `catalogmodel.TestPragmaHonesty` or the schema hash requires a foreign key or a different bucket for `summary_inputs`.
- A test fails because `RowsForSearch` returned different rows than before.

## 12. Open questions

None.

## 13. PR

**Title:** S14-40 — Build the input index
**Body:** objective in one sentence; what shipped against US-1 – US-3; AC evidence (T-1 – T-10 output, the key counts from T-9); before and after timings; deviations from the plan; `Plan: docs/deployment-plan/spike-14/prs/S14-40-build-the-input-index.md`.
