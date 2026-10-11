# Catalog graph in memory (archived)

**Done.** Catalog writes go through one orchestrator, and the open catalog holds the canonical graph in memory (`core/database/graphcache`). Promote, the People / Events / Places lists, place chains, conclusion detail, and a Source's Evidence graph read that graph. Live notes: [`writes`](../../../core/writes/README.md), [`effects`](../../../core/database/effects/README.md), [`catalogmodel`](../../../core/database/catalogmodel/README.md). How Promote loads it: [`promote-graph-alignment.md`](../../promote-graph-alignment.md) §7–8. Matching reads candidates from the same graph: [`matching.md`](../../matching.md).

## Decisions

- `writes.Run` owns the transaction. A write function writes its rows and returns `[]rowchange.Change`. Run records audit, resolves effects, recomputes reconciled values, and reprojects search before commit, then tells the memory tier. A rollback never updates the graph.
- Effects are declared on `catalogmodel`, the same table model `deleteimpact` reads. A path names a declared foreign key or fails at startup. Every non-skip table has an entry, and a cascade into a table with effects is audited.
- Two stores sit on the catalog session: one node per handle, and one graph per Source. An identity claim is the pointer from a Subject to its handle. Nodes load on first read, in batches, and a reload replaces the node. The session already serializes catalog work, so the store has no lock of its own.
- A commit drops the handles it rewrote, their header dependents, the Source graphs it named, and vocabulary when labels or structure changed. A read does not consult the audit log.
- Product keys stay in registries (`connectrules`, `subjectvocab`, `match`, [`core/hops`](../../../core/hops)). The graph carries ids and opaque signatures.
- List headers, place chains, and conclusion values are memos built from the node. Term labels, Source titles, and "today" are applied when the row is shown.
- The Mac session cache is a separate tier ([`page-navigation-performance.md`](page-navigation-performance.md)). Citation, observation, and cited-bridge responses return `WriteEffects`, and the composer drops those cached reads.

Refused: a marks table, updating memory from inside the uncommitted transaction, a persisted change feed, hop caps on the Promote walk, and dropping the whole graph on every revision.

Full-text search, nav counts, the citation row and its notes, Evidence-graph positions, and the conclusion Why (outcomes and Source titles) stay SQL reads.

## Leftovers

| Theme | Home |
| --- | --- |
| Seed `Upsert` (source types, metadata fields, properties, grades, subject-type bindings) still reprojects by hand, off `Run` | [`writes`](../../../core/writes/README.md) |
| Subject delete still calls `RecomputeTx` once after the rows are gone. Header dependents stay outside the effects registry | [`writes`](../../../core/writes/README.md) |
| Writes other than citations, observations, and cited bridges still invalidate the Mac cache through `CatalogMutation` | [`CatalogMutation.swift`](../../../macos/App/Features/Workspace/Session/CatalogMutation.swift) |
