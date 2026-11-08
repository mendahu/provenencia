# Summary rows know their inputs

Design for Spike 14 Batch 8 (user, 2026-11-07). The PRs are in the [map](deployment-plan.md#slice-8--summary-rows-know-their-inputs): S14-40 builds the input index, and S14-42 switches invalidation to it. It came out of planning S14-38, where fixing one stale search document took three new pieces of invalidation code and still left bridge delete half-fixed.

The noun is **input**, not "read": `core/database/effects` already uses `Read`, `ReadPath` and `Set.Reads` for keyed partial reads, and the wire has `WriteEffects.reads`.

## Problem

The catalog keeps three layers of derived data:

| Layer | Built from | How a write found what to rebuild before S14-42 |
| --- | --- | --- |
| 1. Handle values (`auto_reconciler_values`, `auto_reconciler_outcomes`) | Members' Observations | The effects registry (`core/database/effects`): declarative paths from the changed row forward, using a deleted row's old fields. This layer works and does not change |
| 2. Summary rows (graph cache displays) and handle search documents (`catalog_search_docs`) | Layer 1, through `graphcache.Graph` `Node`, `Nodes` and `Follow` in `core/database/conclusionsummaries/compose.go`; plus raw reads of the row's own handle (`gallery.ForEntities` for the lead picture; `searchindex/handles.go` `sqlHandleHeads`, `sqlHandleValues`) | A hand-written inverse of the summary reads, run after the write against the data as it now is |
| 3. App pages and lists | Layers 1 and 2 | `WriteEffects.handles` and `summary_handles` |

Layer 2's inverse was spread over:
- `conclusionsummaries.Dependents` with `walkDepth` and `birthOrDeath`: the reverse of each read, written by hand.
- `conclusionsummaries.AssociationEndpoints`, called before and after the recompute in `autoreconciler.RecomputeReported`, plus `linkEnds`: which ends of a link to walk from.
- The `movedRoots` narrowing in `RecomputeReported`.
- `subjects.Delete`'s own snapshot (`linkedHandles`, `HandlesObservingSubject`, `Dependents`) and its own `RecomputeTx`.
- `graphcache.SetDisplayDeps` / `forgetDisplays`, a fallback that only the test kit still reaches (`writes.Run` always sets `SummaryDepsWalked`).

Every new read in a composer needs a matching edit to the inverse, and nothing checks that the two agree except the rebuild-equals-upkeep tests. Deletes are where it fails structurally. When a link's handle is deleted, its value rows go by CASCADE inside the write, so nothing is left to walk from. That is S14-38: 234 stale documents over 451 unpromotes in the S14-33 probe.

## Shape

**Record each summary row's inputs, and look the dependents up.** When a summary row (and with it the handle's search document) is composed, every graph input it uses is recorded against that row. The inverse is then a stored index rather than code. A write invalidates the rows whose inputs it changed, deleted or linked.

- **A delete works by construction.** "The birth's row used the participation link" stays in the index until the birth is rebuilt, so deleting that link finds the birth with no snapshot.
- **An end that nothing reads through needs no rule.** The city's row never follows its `location` links, so filing or removing a location at the city reaches the event, not the city's 6,000 other events. `linkEnds` and the `Unread` mark S14-38 proposed are not needed.
- **The app is told the same set.** The rebuilt rows and the ends of each link the write changed go to the app in `summary_handles`. That includes the place after a bridge delete, which S14-38 could not reach.

### What an input is

An input key is one of:

| Key | Recorded by | Changes when |
| --- | --- | --- |
| A handle's values `(H, values)` | `InputScope.Node(H)`: a composer reads through its row's `graphcache.InputScope` | H's stored value rows move (the existing `SameDisplayRows` check in `recomputeBatch`), H is merged or deleted |
| A handle's links of one bridge at one end `(H, bridge, end)` | `InputScope.Follow(H, hop)`. `Follow`'s own internal `Node(H)` records nothing more | A link of that bridge with H at that end is created, deleted or merged away, or its values move: a filtered hop (`EventsOfSubject` keeps role subject, `ParentsOfPlace` keeps part_of) can start or stop returning a link whose role or type changed, so the key covers every link of the bridge at that end, not only those the filter returned |
| An association's values `(A, values)` | `InputScope.Follow` for each link it returns, and `InputScope.Node(A)` | A's rows move (role, type, link dates) or A is deleted |

`InputScope.Nodes(ids)` is a prefetch and records nothing. Vocabulary labels (the one raw query in `compose.go`, `loadVocab`) are not handle inputs. They follow the existing vocabulary drop, which does not rewrite stored search documents (logged in [`docs/ideas/structural-debt.md`](../../ideas/structural-debt.md)).

Keying links by bridge and end is what keeps the city cheap. Events at the city use `(city, values)` for its name and `(city, place_relationship, from)` for its parent chain, never `(city, location, place)`.

### A row is its own input

Some inputs a row takes from its own handle do not go through `Node`: the lead picture (`gallery.ForEntities` reads `auto_reconciler_outcomes`, `canonical_entities.preferred_likeness_observation_id` and `files`), and the search document's own head and value queries. A preferred-picture write moves none of the value rows that `SameDisplayRows` compares. So **every handle in `Set.Handles` is rebuilt, its document rewritten and its display dropped, whether or not its values moved**, as `RecomputeReported` does with its roots today. The index records only inputs from other handles and links. The isolated-composition oracle (below) cannot see these raw self-reads; this rule covers them.

### An association records its ends

The search reprojector also stores each association as a dependent of its ends' link keys. Association ids already reach `ReprojectHandles` (`RecomputeReported` passes them as roots), where today they are skipped for having no document kind. A deleted association's ends are then a forward lookup in the index before its entries are cleared. This is what names the place to the app after a bridge delete or an Unpromote.

### Attribution inside a batch

Composers work in batches. `composePersons` prefetches every row's events, builds one `placeGraph` for the batch, then composes each row from it (`onePerson`). A recorder on `Node` and `Follow` alone cannot tell which row an input belongs to, and `placeGraph` lookups never reach the graph.

So inputs are attributed per row explicitly:
- The scope is a `graphcache.Graph` API keyed by a dependent id, not a private type in `conclusionsummaries`. The kinship memo slots (`canonlinks`) can become its next users instead of building a second recorder.
- Each row is composed under its scope.
- For place chains, the batch's `placeGraph` reads each place through its own scope and keeps those keys per place; a row takes the union over the places within the chain depth of its own places, so no key is written by hand. That is the same set composing the row alone reads, so the oracle holds. A row shows ancestors only as far as the build expanded places, so a batch never shows a parent its inputs lack.
- One `RowsForSearch` call counts the place-chain keys it records (`maxInputKeysPerCall`, 250,000). Over the cap the call records no inputs for any row; only a crafted catalog, such as a dense part_of mesh, reaches it.
- A person's place graph is built from its birth and death events only, as `PersonLifeOwners` does, so a marriage's place is not a person's input.
- Shared prefetches record nothing.

The proof is an oracle test. Composing each row alone with a recorder on the graph gives that row's true graph inputs. On fixtures, the batch-attributed sets must equal it.

### Storage

Search documents persist across sessions, so the index is a catalog table, owned by `core/database/searchindex` alone:
- It is written in the same transaction as each handle document (`ReprojectHandles`), cleared by `ClearAll` and rebuilt by `RebuildAll`.
- Each row is (dependent, input handle, key aspect), with an index on the input side.
- **It has no foreign key to `canonical_entities`.** A CASCADE would delete a deleted association's rows inside the write, which is the S14-38 failure again. Rows go only when the reprojector rewrites or drops their dependent.
- `searchindex.ProjectionVersion` is bumped, so an existing catalog fills the index on open. Otherwise invalidation would find nothing on old catalogs, with no error.

The graph cache's display rows are composed by the same composers and use the same index, so there is one index, not one per cache. The table is `catalog_search_inputs` (dependent, input, bridge, endpoint), with the `catalog_search_` prefix of its owner's other tables (S14-40's plan, D-7).

### Invalidation

After layer 1's recompute, `searchindex.ReprojectChanged` (called by `autoreconciler.RecomputeReported`, which `writes.Run` and the `graphcachetest` kit call) does the following for a write:
1. Collects the changed keys: the values of every handle in `Set.Handles` that moved, deleted and merged handles, and the link keys of each created, deleted or moved association. A deleted association's link keys come from the index.
2. Looks up their dependents.
3. Reprojects those search documents, plus every handle in `Set.Handles` (a row is its own input), and records their new inputs.
4. Drops their displays.
5. Reports the dependents, plus the changed links' ends, to the app as `summary_handles`.

This replaces `Dependents`, `AssociationEndpoints`, `linkEnds`, `movedRoots` and `SetDisplayDeps`. `subjects.Delete` stops snapshotting and recomputing on its own; the write around it finds what it moved.

`autoreconciler.Rebuild` recomputes and calls `ReprojectHandles` for every handle, and `ReprojectChanged` reprojects through `reprojectOn` on its own graph, with the `dependent_is_association` flag telling an association's end rows apart.

## Measured at planning (2026-11-07)

Probe on the `TestLargeSourceBudgets` fixture: 8,115 persons, 10,067 events, 51 places and 24,208 associations, with one city holding 6,000 events. It recorded `Node`, `Nodes` and `Follow` reads while composing each person, event and place row on its own (throwaway code, reverted).

| Measure | Result |
| --- | --- |
| Index size | 96,676 dependent→input entries over 38,347 input ids (about 18,000 rows) |
| Inputs per row | median 5, p99 16, max 31 |
| Dependents per input id | median 2, p99 4, max 10,001 (the city: every event there shows its name, as today) |
| Against `Dependents` | identical on 300 sampled handles (100 each kind). With finer keys, the index's inverse is expected to be a subset: `Dependents` also returns a place's parents and succession neighbors "so a later place detail is covered", which are detail pages, not rows. Those are the ends of the changed link, which step 5 reports |
| Composing every row in one batch | 267 ms warm |
| Composing every row alone with recording | 2.27 s, so production attributes inside the batch, and isolated composition is only the test oracle |

Final counts from S14-40 (T-9, the recorder-checked index on the same fixture, every row):

| Kind | Rows | Keys | Median | Max |
| --- | --- | --- | --- | --- |
| Person | 8,117 | 32,365 | 3 | 17 |
| Event | 10,070 | 86,386 | 10 | 10 |
| Place | 52 | 55 | 1 | 4 |

No person row on this fixture holds a place key: the fixture's `bulkCanon` events have no accepted member, so no person has a life pick and its place graph is empty. A person's place inputs are proven by T-2 and T-3 only, on the scenario worlds. Against `Dependents`, T-10 found 9,987 persons excused as reached only through a birth or death event with no accepted member (D-13); the fixture's `bulkCanon` events have no accepted claims.

At the map's Scale (37,000 Subjects) the index is a few hundred thousand rows.

## What stays

- The effects registry for layer 1 (`Set.Handles`), the audit scopes and `Set.Search` for Source and Evidence graph documents.
- The rebuild-equals-upkeep tests, which now also compare the index with a rebuild.
- `graphcache` `detachAssociation` and `mergedNeighbors`, for node reloads and memo roots only; display rows drop from `SummaryDeps` alone.
- The kinship memo slots (`canonlinks`, `DropOnHandleCommit`) stay as they are. Moving them onto the scope API is a later step after S14-42 ([`docs/ideas/structural-debt.md`](../../ideas/structural-debt.md)).
- Detail pages are not dependents in the index. They reload by id from `touchedHandles`, and step 5 tells the app the ends of changed links.
- No production code merges handles today (`merged_into_id` is set only by `graphcachetest`). The merge rows above are held by a `graphcachetest` merge test in S14-42.
