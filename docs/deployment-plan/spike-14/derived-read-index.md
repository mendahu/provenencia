# Derived rows know what they read

Design for Spike 14 Batch 8 (user, 2026-11-07). The PRs are in the [map](deployment-plan.md#slice-8--derived-rows-know-what-they-read). It came out of planning S14-38, where fixing one stale search document took three new pieces of invalidation code and still left bridge delete half-fixed.

## Problem

The catalog keeps three layers of derived data:

| Layer | Built from | How a write finds what to rebuild today |
| --- | --- | --- |
| 1. Handle values (`auto_reconciler_values`) | Members' Observations | The effects registry (`core/database/effects`): declarative paths from the changed row forward, using a deleted row's old fields. This layer works and does not change |
| 2. Summary rows (graph cache displays) and search documents (`catalog_search_docs`) | Layer 1, read through `graphcache.Graph` `Node`, `Nodes` and `Follow` in `core/database/conclusionsummaries/compose.go`; a search document adds its handle's own values (`searchindex/handles.go`) | A hand-written inverse of the summary reads, run after the write against the data as it now is |
| 3. App pages and lists | Layers 1 and 2 | `WriteEffects.handles` and `summary_handles` |

Layer 2's inverse is spread over:
- `conclusionsummaries.Dependents` with `walkDepth` and `birthOrDeath`: the reverse of each read, written by hand.
- `conclusionsummaries.AssociationEndpoints`, called before and after the recompute in `autoreconciler.RecomputeReported`, plus `linkEnds`: which ends of a link to walk from.
- The `movedRoots` narrowing in `RecomputeReported`.
- `subjects.Delete`'s own snapshot (`linkedHandles`, `Dependents`) and its own `RecomputeTx`, outside `writes.Run`.
- `graphcache.SetDisplayDeps` and `detachAssociation`.

Every new read in a composer needs a matching edit to the inverse, and nothing checks that the two agree except the rebuild-equals-upkeep tests. Deletes are where it fails structurally. When a link's handle is deleted, its value rows go by CASCADE inside the write, so nothing is left to walk from. That is S14-38: 234 stale documents over 451 unpromotes in the S14-33 probe.

## Shape

**Record what each derived row read, and look the readers up.** When a summary row or a search document is composed, every graph read it makes is recorded against that row. The inverse is then a stored index rather than code. A write invalidates the rows that read what it changed, deleted or linked.

- **A delete works by construction.** "The birth's document read the participation link" stays in the index until the birth is rebuilt, so deleting that link finds the birth with no snapshot.
- **An end that nothing reads through needs no rule.** The city's row never follows its `location` links, so filing or removing a location at the city reaches the event, not the city's 6,000 other events. `linkEnds` and the `Unread` mark S14-38 proposed are not needed.
- **The app is told the same set.** The rebuilt readers and the ends of each link the write changed go to the app in `summary_handles`. That includes the place after a bridge delete, which S14-38 could not reach.

### What a read is

A read key is one of:

| Key | Recorded by | Changes when |
| --- | --- | --- |
| A handle's values `(H, values)` | `Graph.Node(H)` | H's stored value rows move (the existing `SameDisplayRows` check in `recomputeBatch`), H is merged or deleted |
| A handle's links of one bridge at one end `(H, bridge, end)` | `Graph.Follow(H, hop)` | A link of that bridge with H at that end is created, deleted, or merged away |
| An association's values `(A, values)` | `Follow` (for each link returned) and `Node(A)` | A's rows move (role, type) or A is deleted |

`Nodes(ids)` is a prefetch and records nothing. Vocabulary labels (the one raw query in `compose.go`, `loadVocab`) are not handle reads. They follow the existing vocabulary drop.

Keying links by bridge and end is what keeps the city cheap. Events at the city read `(city, values)` for its name and `(city, place_relationship, from)` for its parent chain, never `(city, location, place)`.

### An association records its ends

When layer 1 writes an association's values, the index also stores the association as a reader of its ends' link keys. A deleted association's ends are then a forward lookup in the index before its entries are cleared. This is what names the place to the app after a bridge delete or an Unpromote.

### Attribution inside a batch

Composers work in batches. `composePersons` prefetches every row's events, builds one `placeGraph` for the batch, then composes each row from it (`onePerson`). A recorder on `Node` and `Follow` alone cannot tell which row a read belongs to, and `placeGraph` lookups never reach the graph.

So reads are attributed per row explicitly:
- Each row is composed under a read scope.
- `placeGraph` records the chain members a row consumes.
- Shared prefetches record nothing.

The proof is an oracle test. Composing each row alone with a recorder on the graph reads gives that row's true read set. On fixtures, the batch-attributed sets must equal it.

### Storage

Search documents persist across sessions, so the index is a catalog table, written in the same transaction as the documents it describes, and rebuilt by `searchindex.RebuildAll`. Each row is (reader, read handle, key aspect), with an index on the read side. The graph cache's display rows are composed by the same composers and use the same index, so there is one index, not one per cache. The table and column names are fixed in S14-40's plan.

### Invalidation

After layer 1's recompute, `writes.Run`:
1. Collects the changed keys: moved handle values, deleted and merged handles, and the link keys of each created, deleted or moved association. A deleted association's link keys come from the index.
2. Looks up their readers.
3. Reprojects those search documents and records their new reads.
4. Drops their displays.
5. Reports the readers, plus the changed links' ends, to the app as `summary_handles`.

This replaces `Dependents`, `AssociationEndpoints`, `linkEnds` and `movedRoots`. `subjects.Delete` stops snapshotting and recomputing on its own and runs on `writes.Run` like every other write.

## Measured at planning (2026-11-07)

Probe on the `TestLargeSourceBudgets` fixture: 8,115 persons, 10,067 events, 51 places and 24,208 associations, with one city holding 6,000 events. It recorded `Node`, `Nodes` and `Follow` reads while composing each person, event and place row on its own (throwaway code, reverted).

| Measure | Result |
| --- | --- |
| Index size | 96,676 reader→read entries over 38,347 read ids (about 18,000 rows) |
| Reads per row | median 5, p99 16, max 31 |
| Readers per read id | median 2, p99 4, max 10,001 (the city: every event there shows its name, as today) |
| Agreement with `Dependents` | identical on 300 sampled handles (100 each kind): no id only one side had |
| Composing every row in one batch | 267 ms warm |
| Composing every row alone with recording | 2.27 s, so production attributes inside the batch, and isolated composition is only the test oracle |

At the map's Scale (37,000 Subjects) the index is a few hundred thousand rows. The probe did not key links by bridge and end; S14-40 does, and repeats the agreement check with the finer keys.

## What stays

- The effects registry for layer 1 (`Set.Handles`), the audit scopes and `Set.Search` for Source and Evidence graph documents.
- The rebuild-equals-upkeep tests, which now also compare the index with a rebuild.
- The graph cache's memo slots for kinship (`canonlinks`, `DropOnHandleCommit`) stay as they are. Moving them onto the index is a later step after S14-42 ([`docs/ideas/structural-debt.md`](../../ideas/structural-debt.md)).
- Detail pages are not readers in the index. They reload by id from `touchedHandles`, and the index tells the app the ends of changed links.
