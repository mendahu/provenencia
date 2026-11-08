# Cache architecture: entities and indexes in memory, answers composed

Design for Spike 16 Batch 2 (user, 2026-11-08). The PRs are in the [map](deployment-plan.md#batch-2--cache-architecture); the decisions are its Batch 2 working rules. It comes from a review of the catalog's caches on 2026-11-08: five reviewers (persisted projections, the Go graph cache and its walkers, Promote, the Conclusion page reads, the Swift session) and a survey of which catalog entities no cache holds, read against `main` @ 78f2681 and the S14-40 and S14-42 branches. Nothing was run; timings are from the [second app-health audit](../spike-14/app-health-audit-2.md) (2026-10-28), [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md), [`spike-14/summary-inputs.md`](../spike-14/summary-inputs.md) and the Spike 13 ship logs, and every "est." is an estimate to be measured by the PR that relies on it.

## Problem

The catalog keeps derived data in three layers: SQLite (the auto-reconciler's values and outcomes, search documents and, from S14-40, the input index), the Go catalog graph (`core/database/graphcache`) and the Swift session (`WorkspaceSession`). Most of the conclusion layer's complexity is code that rebuilds what a cache could hold, for four reasons.

1. **Caches that empty on any write.** Any commit that names a handle drops Promote's match index and its handle profiles (`graphcache/commit.go:41`) and the `canonlinks` `for` and `siblings` memos (`DropOnHandleCommit`, `canonlinks/reader.go:65-76`). The next proposal and the next Family read rebuild them from scratch: at 16,000 persons a proposal after a handle commit took 4.05–4.19 s against 661–779 ms warm (audit 2, B-8).
2. **Walks repeated per request.** Each Promote proposal walks the whole connected canon from its seeds (`promotealign/canon.go:115-176`: about 18,000 handles, 24,000 links and 6,000 implied links, 122 MB allocated, on the large fixture). Each summary batch rebuilds place chains (`buildPlaceGraph`); the Places list rebuilds them on every read (`finishPlaces`); the Person detail derives the life rows a second time (`PersonLifeOwners`, `compose.go:751-863`); the Timeline walks each event's participants.
3. **No cache between "one entity" and "one whole graph".** The graph cache holds whole canonical entities (`Node`) and whole Evidence graphs (a 64-graph LRU). Nothing holds a Subject, a Source or a graph's shell, so readers join SQL or load whole graphs: the provenance sheet's records are an 8-table join on every page open ("The Why stays a SQL read each time the page opens", `conclusiondetails.go:8`); the Sources section loads every member graph whole and rescans it per record (audit P-4); graph progress runs full-table `GROUP BY`s (`graphprogress.go:20-41`); a Source rename evicts every loaded graph of the Source because each graph copies the Source's words (`effects/registry.go:309`).
4. **Swift invalidation written per key, and pages evicted.** Family and Timeline reload on any handle write and every hidden one is evicted (`CatalogQueryRegistry.swift` `personSection`); per-entity pages are `.evict`, so Back shows a spinner; each key kind hand-writes which ids it follows.

## Shape

**Cache entities and indexes; compose answers.** Small records keyed by id, and a few indexes for the walks that repeat, held in Go memory, each updated exactly from the commit's effects. Readers compose their answer from them at read time, which costs microseconds; nothing composed is stored except the summary rows search needs, which S14-42 keeps exact.

| Layer | Holds | Updated by |
| --- | --- | --- |
| SQLite | Truth; auto-reconciler values and outcomes; search documents; the input index (S14-40) | Written in the commit |
| Go entity records | Canonical nodes with their outcome records (S16-11); Source records (S16-12); Subject records (S16-13); the Evidence-graph catalog (S16-15); vocabulary | The commit signal for that entity (`effects.Set`: `Handles`, `SourceScopes`, `EvidenceGraphs`, `Reads`) |
| Go indexes | Kinship (S16-16); place hierarchy (S16-18); Promote's profiles and name index (S16-20) | Apply on commit from declared dependencies |
| Go stored answers | Summary rows (displays) only | `Set.SummaryDeps` from the input index (S14-42) |
| Swift | One value per view key | The inputs the value reports (S16-24, S16-25); stale while refreshing (S16-23) |

### One home for a catalog-wide cache (S16-10)

`graphcache` gains a registry beside `RegisterNodeMemo`: a **graph index** names a catalog-wide cache with

- **Build**, run lazily on the first read on any graph (the session graph, or a write's one-shot graph);
- **Apply**, run by `OnCommit` with the commit's `effects.Set` on the session graph, to update the entries it names, instead of dropping the whole;
- **Inputs**, the `graphcache.InputKey`s an entry was built from, for indexes a summary composer reads (below);
- **Check**, which Verify runs (`PROVENENCIA_GRAPH_VERIFY`): a fresh Build compared with the kept index;
- **DropOnVocabulary**, and a memory **Cap**.

The match index, alias index and name format, today three hand-written fields with their own setters and drop rules (`graph.go`), move onto it (audit 2 B-8: "one keyed slot"). Node memos stay for results per node. Every later cache in this batch registers here, so there is one way to add a cache and one place Verify looks.

The registry comes with **one table of every cache** — Go and Swift, its key, what builds it and the signal that updates it — in a new domain doc, `docs/catalog-caches.md` (S16-10), which each later PR updates.

### The write path composes on a one-shot graph

A write composes summary rows on a throwaway graph over its own transaction (`searchindex` `newGraph = graphcache.Over`), because the session graph is still on the previous commit and the catalog has one connection. So an index a composer reads cannot only be built at open: it builds lazily on whichever graph asks and is kept, and updated by Apply, only on the session graph. The place index (S16-18) is the one that matters; places are few, and the per-write build is to be measured (est. milliseconds). The kinship index and Promote's caches are read only on the session graph.

### Indexes hand their inputs to the row

S14-40 records each summary row's inputs through `graphcache.InputScope` (`Node`, `Follow`), and S14-42 finds a write's dependent rows in that record. A composer that reads an index instead of `Node` and `Follow` must still record what it read, or S14-42's lookup misses the row and it goes stale. So an index entry keeps the `InputKey`s it was built from and the composer adds them to its scope. `buildPlaceGraph` already does this per place (`pg.inputs`, `inputBudget.addClosure`), with a cap past which a chain stops at the depth its inputs cover; the place index keeps both.

### Entity records

- **Outcome records on the node (S16-11).** `Node` gains each Property's outcome rows (Observation, Subject, reason, value rank, denied by, vote) and the chosen likeness, loaded in one more batched query beside `Values`. `Set.Handles` reloads the node, so they are exact. `sqlPropertyIDs`, the reconciler half of `sqlOutcomes` and `gallery.ForEntities` go; the gallery and the summary's lead picture compose from the node. Est. 4–6 MB at 16,000 persons.
- **Source records (S16-12).** Per Source: ref, title, type, cover, credibility grade and its Artifacts (label, file); a Source's graphs are the Evidence-graph catalog's (S16-15). Two queries at open (est. 5–15 ms, 1.5 MB); `Set.SourceScopes` reloads the named Sources. A Source's words live only here: `EvidenceGraph` stops copying them, `loadGraphHead`'s join and the outcome SQL's Source, Artifact and credibility joins go, and the `sources` effect stops naming a Source's graphs for a title or type change once no reader takes Source words from a graph (Swift's claims and record-links keys follow `.source`). Credibility grades join the vocabulary.
- **Subject records (S16-13).** Per Subject: ref, label, type, graph, Source, accepted claim and its Observations with each Citation's label, page and Artifact and the uncertainty flag. Built lazily, sliced from a loaded graph or one batched query (`graphfill`'s columns); a graph in `Set.EvidenceGraphs` drops its Subjects' records. Records hold ids, not copied words: a provenance record is the node's outcome (S16-11) + the Subject record + the Source record. `sqlOutcomes`, `toponymsOf` and the date and name lookups of the detail read go; est. a few MB in a session, 25–35 MB if every Subject is read, under its own cap. The Sources section composes from them (S16-14): `memberGraphs` loads no whole graph, so P-4's rescans and the past-64-graphs thrash leave that section.
- **The Evidence-graph catalog (S16-15).** Per graph: Source, ref, name and its subject, filed and Observation counts; per Source, its graphs in list order. Built at open or on first read (est. 30–80 ms, 2 MB); updated from `Set.Reads` (`evidence_graph_progress`, `evidence_graphs`). `graphprogress`'s full-table scans become map reads, and claims and record links group and order graphs without loading them.

### Indexes for the walks

- **Kinship (S16-16, S16-17).** The `canonlinks` per-pivot results (which already have exact one-hop `Deps`) plus a reverse map from each person to the implied links naming them and their recorded relationship links. `For(id)` is a lookup; siblings are read through the parents' children. Apply recomputes the pivots of the commit's roots and their neighbours. Built at open from the nodes `promotealign.Warm` already loads (est. 0.1–0.3 s at 16,000 persons, 10–20 MB). The `for` and `siblings` slots and their `DropOnHandleCommit` go, with about 300 of `reader.go`'s 668 lines; the Family read becomes one pass. Its link order is `For`'s today, so Promote's tie-breaks do not move.
- **Place hierarchy (S16-18).** Every place's shell and its `part_of` and `succeeded_by` links with decoded dates, both directions, and each place's input keys. Chains that depend on a date stay functions over it. Built lazily (one batched association load); Apply rebuilds a place only when its own shell or place-relationship links move, so filing a location at a city does not churn it, and reading a city's chain no longer loads its 6,000 location links. `buildPlaceGraph`, `placeStep`, `rememberShell`, `rememberPlaceLink`, `eventPlaceGraph` and `placeNeighborhood` (about 210 lines) collapse to one accessor.
- **Summary rows carry what readers re-derive (S16-19).** A person's row keeps its birth and death owner events and their event–place links; an event's row keeps the Property its date came from and its participants by role. `PersonLifeOwners` and `lifeOwners` go, the Timeline reads participants from rows it already loads, and the two event–person shapes (structural debt A-33) become one.

### Promote

- **Profiles and the name index survive writes (S16-20).** A handle's profile is a node memo whose `Deps` are the handle and its events (audit 2 B-8's fix). The shortlist index moves onto the graph-index registry and Apply patches it: remove the dirty handles' old keys, add their new ones, keep the generation for the rest; the candidates memo re-ranks only Subjects whose lookup keys meet a dirty handle. `indexGen`, the generation checks and the profiler's shared map go. `Warm` fills full profiles in the same pass. Verify compares the patched index with a fresh build. This is [`promote-ranking.md`](../../ideas/promote-ranking.md)'s "Name index on commit".
- **Neighbourhoods instead of the whole canon (S16-21, S16-22).** `Align` reads a `CanonView` — one handle's profile and links at a time, from the kinship index and S16-20's profiles — instead of `loadCanon`'s walk of the whole component. `Connections`' loop check, which needs every filed `part_of` and `succeeded_by` link, reads the place index (S16-21). Cost then grows with the Source, not the catalog (est. warm proposal 661–779 ms → about 150 ms at 16,000 persons). An equivalence test against today's walk on the large fixture holds the result; this is "Narrower canon walk".

### Swift

- **Stale while refreshing (S16-23).** The per-entity Conclusion keys go from `.evict` to `.markStale`; `apply` evicts only the ids in `effects.deleted`; a not-found reload over a stale value reads as gone (S16-04's reading). Back to a page shows its last value at once. A cap on hidden per-entity keys bounds memory.
- **Values report their inputs (S16-24, S16-25).** Each per-entity read returns the handles, graphs and Sources it was built from (recorded through `InputScope`, the kinship index's entries, and explicit graph and Source ids); Swift has one `.followsInputs` refresh in place of `personSection`, `personGraphsSection`, the per-kind id closures and S16-09's Source follow. Cache keys become plain values first (S16-28), which removes the key enum's switches and `guard case` blocks.
- **Lists merge the touched rows (S16-26)** with the existing `.merged` read, by an id filter and a sort key from Go; exact once S14-42 makes `summary_handles` exact.
- **One primitive for derived values (S16-27):** `session.derived(...)`, recomputed only when an input's revision changes, replacing four hand memos and building a section's content once per value.
- **Prefetch last (S16-29):** a person's other sections and a hovered person, once S16-23 and S16-25 stop the next write throwing them away.

## Warming

- At open, after `promotealign.Warm` (which already loads every person, event and place node): build the kinship index and fill full profiles in the same pass; the place index builds on first read. CPU only, behind the open spinner (est. 0.1–0.3 s at 16,000 persons; open is 4.16 s there today).
- After a write, nothing re-warms in the background: every engine call waits in `catalogsession`'s one queue, so a background build would delay the user's next action. Apply keeps each cache current inside `OnCommit` instead.
- In Swift, stale-while-refreshing (S16-23) is the largest perceived-speed win; prefetch (S16-29) comes after it.

## What goes

`canonicalgraph` (its last production caller, `Dependents`, leaves with S14-42); `DropOnHandleCommit` and `dropHandleMemos`; dropping the match index on every handle commit; the three warm fields on `Graph`; `sqlOutcomes`, `sqlPropertyIDs`, `gallery.ForEntities`; the Source fields on `EvidenceGraph`; `graphprogress`'s scans; most of `canonlinks`' loading; the place-chain builders; `PersonLifeOwners`; `loadCanon`'s walk; `personSection` and `personGraphsSection` in Swift. Rough total from the review: about 1,500 lines out and a few hundred in (est.).

## Not this batch

- Persisted projections (stored summary rows, a canonical links table, stored shortlist keys): refused for now (rule "No new persisted projections").
- The record links pair memo (P-9), participants sent once (P-8), lazy Timeline cards and siblings sent once (P-23): [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md).
- A batch Evidence-graph read for a person past 64 graphs: [`ideas/structural-debt.md`](../../ideas/structural-debt.md) (S13-16 PRF-2); record links still load whole graphs.
- Promote's per-graph compare and alike-pair memos (audit 2 B-11), coverage-aware similarity: [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md).
- A file-path memo, a last-activity table, stored graph progress counters: low value.
