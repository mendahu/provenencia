# Cache architecture: entities and indexes in memory, answers composed

Design for Spike 16 Batch 2 (user, 2026-11-08). The PRs are in the [map](deployment-plan.md#batch-2--cache-architecture); the decisions are its Batch 2 working rules, reshaped by the map review's round 2 ([review](deployment-plan.review.md), K-17 – K-28). It comes from a review of the catalog's caches on 2026-11-08: five reviewers (persisted projections, the Go graph cache and its walkers, Promote, the Conclusion page reads, the Swift session) and a survey of which catalog entities no cache holds, read against `main` @ 78f2681 and the S14-40 and S14-42 branches, and re-checked on `main` @ 2ac041a after both landed (#746, #748). Nothing was run; timings are from the [second app-health audit](../spike-14/app-health-audit-2.md) (2026-10-28), [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md), [`spike-14/summary-inputs.md`](../spike-14/summary-inputs.md) and the Spike 13 ship logs, and every "est." is an estimate the PR that relies on it measures first.

## Problem

The catalog keeps derived data in three layers: SQLite (the auto-reconciler's values and outcomes, search documents and, from S14-40, the input index), the Go catalog graph (`core/database/graphcache`) and the Swift session (`WorkspaceSession`). Most of the conclusion layer's complexity is code that rebuilds what a cache could hold, for four reasons.

1. **Caches that empty on any write.** Any commit that names a handle drops Promote's match index and its handle profiles (`graphcache/commit.go:41`) and the `canonlinks` `for` and `siblings` memos (`DropOnHandleCommit`, `canonlinks/reader.go:65-76`). Setting the alias index or the name format drops the match index and every Evidence-graph memo as a side effect (`dropDerived`, `graph.go:231`). The next proposal and the next Family read rebuild them from scratch: at 16,000 persons a proposal after a handle commit took 4.05–4.19 s against 661–779 ms warm (audit 2, B-8).
2. **Walks repeated per request.** Each Promote proposal walks the whole connected canon from its seeds (`promotealign/canon.go:115-176`: about 18,000 handles, 24,000 links and 6,000 implied links, 122 MB allocated, on the large fixture). Each summary batch rebuilds place chains (`buildPlaceGraph`), the Places list rebuilds them on every read (`finishPlaces`), the Person detail derives the life rows a second time (`PersonLifeOwners` in `compose.go`), and the Timeline walks each event's participants.
3. **No cache between "one entity" and "one whole graph".** The graph cache holds whole canonical entities (`Node`) and whole Evidence graphs (a 64-graph LRU). Nothing holds a Subject's record, a Source or a graph's shell, so readers join SQL or load whole graphs: the provenance sheet's records are an 8-table join on every page open ("The Why stays a SQL read each time the page opens", `conclusiondetails.go:8`); the Sources section loads every member graph whole and rescans it per record (audit P-4); graph progress runs full-table `GROUP BY`s (`graphprogress.go:20-41`); a Source rename evicts every loaded graph of the Source because each graph copies the Source's words (`effects/registry.go:310`).
4. **Swift invalidation written per key, and pages evicted.** Family and Timeline reload on any handle write and every hidden one is evicted (`CatalogQueryRegistry.swift` `personSection`); per-entity pages are `.evict`, so Back shows a spinner; each key kind hand-writes which ids it follows.

## Shape

**Cache entities and results per key; compose answers.** Records keyed by id, and results per key for the walks that repeat, held in Go memory and each updated exactly from the commit. Readers compose their answer at read time, which costs microseconds; nothing composed is stored except the summary rows search needs, which S14-42 keeps exact.

| Layer | Holds | Updated by |
| --- | --- | --- |
| SQLite | Truth; auto-reconciler values and outcomes; search documents; the input index (S14-40) | Written in the commit |
| Go entity records | Canonical nodes with their outcome records (S16-11); Source records (S16-12); the Evidence-graph store with its records tier of Subject records (S16-13); the Evidence-graph catalog (S16-15); vocabulary | The commit signal for that entity (`effects.Set`: `Handles`, `SourceScopes`, `EvidenceGraphs`, `Reads`) |
| Go results per key | Kinship pivots (today's `canonlinks.kinship` slot, S16-16); place entries (S16-18); the match index of profiles, keys and postings (S16-20); Promote's candidates (S16-32) | Apply on commit, from the roots the commit reloaded |
| Go stored answers | Summary rows (displays) only | `Set.SummaryDeps` from the input index (S14-42) |
| Swift | One value per view key | The inputs the value reports (S16-24, S16-25); stale while refreshing (S16-23) |

### One cache registry, two ways to fill (S16-10)

`graphcache` registers every cache that is not a typed field of `Graph`. An entry declares:

- **Fill:** per key (today's `NodeMemo`, a Fill plus `Deps`, which folds into the registry) or whole (a lazy Build).
- **Apply(g, roots, set):** run by `OnCommit` after its reload, with the same roots `dropMemos` gets today — association ids become their ends (`detachAssociation`), merges add their neighbours (`mergedNeighbors`), removed nodes add their links' neighbours — and the commit's `effects.Set`, so no entry works out what a commit touched again.
- **Reads:** what else the entry is derived from (the vocabulary, the alias index, the name format, another entry). This replaces `DropOnVocabulary` and `dropDerived`: a name-format change drops exactly what was keyed on it, such as profiles.
- **Lifetime:** the graph, or the catalog connection for an entry that survives `Drop` (the alias index and the name format do today).
- **Check,** which Verify (`PROVENENCIA_GRAPH_VERIFY`) runs, and a memory **Cap**.

The match index, alias index, name format and the `canonlinks` slots move onto it. `Activity`, the summary displays and the labels stay typed fields of `Graph` (each is one map patched by `OnCommit` with its own rule); the cache table says why. The registry comes with **one table of every cache** — Go and Swift, its key, its fill, the signal that updates it, its cap and its size at Scale — in a new domain doc, `docs/catalog-caches.md` (S16-10), which each later PR updates.

### The write path fills only what it reads

A write composes summary rows on a throwaway graph over its own transaction (`searchindex` `newGraph = graphcache.Over`), because the session graph is still on the previous commit and the catalog has one connection. A per-key entry there fills only the keys the write's composers read and is discarded with the graph; no whole Build runs on a write path, so a write's cost follows the write, not the catalog. The session graph keeps entries and drops or updates them per key.

### A cache a composer reads hands its inputs to the row

S14-40 records each summary row's inputs through `graphcache.InputScope` (`Node`, `Follow`), and S14-42 finds a write's dependent rows in that record. A composer that reads a cache entry instead of `Node` and `Follow` must still record what the entry was built from, or S14-42's lookup misses the row and it goes stale. So an entry keeps the `InputKey`s it was built from, built with `graphcache`'s key constructors (exported by S16-18, today the unexported `nodeKey` and `followKeys`), and the composer adds them to its scope. `buildPlaceGraph` already does this per place (`pg.inputs`, `inputBudget.addClosure`), with a cap past which a chain stops at the depth its inputs cover; the place entries keep both.

### Entity records

- **Outcome records on the node (S16-11).** `Node` gains each Property's outcome rows (Observation, Subject, reason, value rank, denied by, vote) and the chosen likeness, loaded in one more batched query beside `Values`; `Set.Handles` reloads the node, so they are exact. `sqlPropertyIDs`, the reconciler half of `sqlOutcomes` and `gallery.ForEntities` go: the detail page, the summaries' lead picture and `SetPreferredLikeness`'s validation (`canonicalentities.go:275`, through the node on `graphcache.Over(tx)`) compose the gallery from the node. Every node load carries the outcomes, at open and on each write's one-shot graph, so open time and a write at the 6,000-event city are measured with the memory (est. 4–6 MB at 16,000 persons).
- **Source records (S16-12).** Per Source: ref, title, type, cover, credibility grade and its Artifacts (label, file); a Source's graphs are the catalog's (S16-15). A whole entry built on first read (two queries, est. 5–15 ms, 1.5 MB); Apply reloads the Sources in `Set.SourceScopes`, usually one per write (two small queries). A Source's words live only here: `EvidenceGraph` stops copying them, `loadGraphHead`'s join and the outcome SQL's Source, Artifact and credibility joins go, and the `sources` effect stops naming a Source's graphs for a title or type change once no reader takes Source words from a graph (the app's claims and record links keys follow `.source`). Credibility grades join the vocabulary.
- **The records tier of the Evidence-graph store (S16-13).** The 64-graph store gains a second tier, per graph: its Subject records (ref, label, type, Source id, accepted claim, and the Observations with each Citation's label, page and Artifact id and the uncertainty flag) and the memos that must outlive the loaded graph (Promote's candidates, S16-32), under a cap of its own. `dropEvidenceGraphs` drops both tiers, so there is one per-graph drop path. The tier fills from a loaded graph or from a step split out of `graphfill/load.go`, so there is one loader of Subject and Observation rows, and graphfill's registered check covers it. Records hold ids, not copied words. Est. a few MB in a session, 25–35 MB if every graph's records are read.
- **Provenance composed (S16-30).** A provenance record is the node's outcome row (S16-11), the Subject record and the Source record (S16-12): `sqlOutcomes`, `toponymsOf`, the detail read's date and name lookups and the second pass of `PersonDerivedRows` go.
- **The Sources section (S16-14).** The claims read composes from the records tier and Source records, and orders graphs from the catalog: `memberGraphs` loads no whole graph and rescans none, so P-4's rescans and the past-64-graphs thrash leave that section. Record links still load whole graphs to compare them.
- **The Evidence-graph catalog (S16-15).** Per graph: Source, ref, name and its subject, filed and Observation counts; per Source, its graphs in list order. A whole entry built on first read (est. 30–80 ms, 2 MB); Apply recounts the graphs and re-lists the Sources named in `Set.Reads` (`evidence_graph_progress`, `evidence_graphs`). `graphprogress`'s full-table scans become map reads.

### Results per key for the walks

- **Kinship answers composed from the pivots (S16-16, S16-17).** The `canonlinks.kinship` slot already holds each pivot's implied links with exact one-hop `Deps`. A person's implied links (`For`) and siblings are composed at read time from the pivots around them, parents bounded by `MaxPivotEnd`, instead of stored in the `for` and `siblings` slots that drop on every handle commit. Those two slots, `DropOnHandleCommit` and `dropHandleMemos` go; there is no catalog-wide kinship index, warm or reverse map (map review K-20), and a term edit drops only the pivots. The Family read becomes one pass (S16-17). Composition order is `For`'s today, so Promote's tie-breaks do not move.
- **Place entries (S16-18, S16-31).** Per place: its shell and its `part_of` and `succeeded_by` links with decoded dates, both directions, and the input keys it was built from. Filled a frontier at a time by a narrow load that skips location links, so reading a city's chain no longer loads its 6,000 events' links; on the session graph Apply refreshes a place only when its shell or place-relationship links move. Date-dependent chains stay functions over the entries. The Places list and the place steps read them first (S16-18), then the event and life composers (S16-31), with S14-40's input cap. The composers keep S16-08's helper for an event's places.
- **Summary rows carry what readers re-derive (S16-19).** A person's row keeps its birth and death owner events and their event–place links; an event's row keeps the Property its date came from and its participants by role. `PersonLifeOwners` and `lifeOwners` go, the Timeline reads participants from rows it already loads, and the two event–person shapes (structural debt A-33) become one.

### Promote

- **One match index entry (S16-20).** Profiles by handle, lookup keys by handle and the shortlist postings, as one whole entry filled by `promotealign.Warm` at open and declaring Reads on the alias index and the name format. Apply re-profiles the commit's roots and the people at the events they touched, once, diffs each changed handle's old and new keys, patches the postings, and hands the changed keys to the candidates. `indexGen`, the generation checks and the profiler's shared map go; Verify compares the patched entry with a fresh build. This is [`promote-ranking.md`](../../ideas/promote-ranking.md)'s "Name index on commit".
- **Candidates re-rank only what moved (S16-32).** Each Subject's ranked candidates live in the records tier and are re-ranked only when a changed key meets the Subject's lookup keys or a changed handle is in its list; they survive the loaded graph's eviction.
- **Neighbourhoods instead of the whole canon (S16-33, S16-21, S16-22).** First `Align` reads a `CanonView` interface served by today's `Canon`, with goldens unchanged (S16-33). `Connections`' loop check takes the filed `part_of` and `succeeded_by` links from the place entries instead of every canon edge (S16-21). Then the session-graph view serves one handle's profile and links at a time from the match index and the kinship pivots, and `loadCanon`'s walk of the whole component goes (S16-22). Cost then grows with the Source, not the catalog (est. warm proposal 661–779 ms → about 150 ms at 16,000 persons); an equivalence test on the large fixture holds the result. This is "Narrower canon walk".

### Swift

- **Stale while refreshing (S16-23).** The per-entity Conclusion keys go from `.evict` to `.markStale`; `apply` evicts only the ids in `effects.deleted`; a not-found reload over a stale value reads as gone (S16-04's reading). Back to a page shows its last value at once. A cap of 50 hidden entities bounds memory.
- **Values report their inputs (S16-24, S16-25).** Each per-entity read returns the handles (from `InputScope` and the kinship pivots), graphs and Sources it was built from, as three wire lists; `graphcache.InputKey`, which the persisted input index stores, keeps its one kind. Swift has one `.followsInputs` refresh in place of `personSection`, `personGraphsSection`, the per-kind id closures, S16-09's Source follow and S16-12's. Cache keys become plain values first (S16-28).
- **Lists merge the touched rows (S16-26)** through the existing `.merged` read, by an id filter and a sort key from Go.
- **One primitive for derived values (S16-27):** `session.derived(...)`, rebuilt only when an input's revision changes, replacing four hand memos and building a section's content once per value.
- **Prefetch last (S16-29):** a `prefetch` tag in S16-03's list of section reads, for Family and Timeline, and a hovered person's page, once S16-23 and S16-25 stop the next write throwing them away.

## Warming

- At open, `promotealign.Warm` already loads every person, event and place node; S16-20 fills the match index's full profiles in the same pass. Every other cache fills on first read. CPU only, behind the open spinner; open is 4.16 s at 16,000 persons today and is measured after S16-11 and S16-20.
- After a write, nothing re-warms in the background: every engine call waits in `catalogsession`'s one queue, so a background build would delay the user's next action. Apply keeps each cache current inside `OnCommit`, and each PR states that Apply's work per commit (rule "Apply states its fan-out").
- In Swift, stale-while-refreshing (S16-23) is the largest perceived-speed win; prefetch (S16-29) comes after it.

## What goes

`canonicalgraph` (its last production caller, `Dependents`, leaves with S14-42); `RegisterNodeMemo` as a separate mechanism, `DropOnHandleCommit`, `dropHandleMemos` and `dropDerived`; the `for` and `siblings` slots; dropping the match index on every handle commit; the three warm fields on `Graph`; `sqlOutcomes`, `sqlPropertyIDs`, `gallery.ForEntities`; the Source fields on `EvidenceGraph`; `graphprogress`'s scans; most of `canonlinks`' loading; the place-chain builders; `PersonLifeOwners`; `loadCanon`'s walk; `personSection` and `personGraphsSection` in Swift. Rough total from the review: about 1,500 lines out and a few hundred in (est.).

## Not this batch

- Persisted projections (stored summary rows, a canonical links table, stored shortlist keys): refused for now (rule "No new persisted projections").
- A catalog-wide kinship index with a reverse map: replaced by composing from the pivots (K-20); revisit only if composing measures slow.
- The record links pair memo (P-9), participants sent once (P-8), lazy Timeline cards and siblings sent once (P-23): [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md).
- A batch Evidence-graph read for a person past 64 graphs: [`ideas/structural-debt.md`](../../ideas/structural-debt.md) (S13-16 PRF-2); record links still load whole graphs.
- Promote's per-graph compare and alike-pair memos (audit 2 B-11), coverage-aware similarity: [`ideas/promote-ranking.md`](../../ideas/promote-ranking.md).
- A file-path memo, a last-activity table, stored graph progress counters: low value.
