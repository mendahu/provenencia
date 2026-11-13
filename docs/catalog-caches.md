# Catalog caches

Every cache the app keeps over the catalog, with what fills it, what updates it and what it costs. The Go caches are one registry (`core/database/graphcache/cache.go`): an owner registers a typed handle in `init()` and reads and writes only through it. `TestCacheDocListsEveryEntry` (`core/catalogsession/cache_doc_test.go`) holds the first table to the registry, so a PR that registers a cache adds its row here in the same change. `BenchmarkCacheSizes` (`core/database/promotealign/bench_test.go`) measures the sizes; a PR that adds an entry adds its fill to the benchmark.

Rule: every cache names the signal that updates it, and updates exactly. The signal is a field of `effects.Set` that the entry's Apply reads, or a `DerivedFrom` declaration (the vocabulary, or an earlier whole entry). A cache that empties itself on any write is a known exception, listed below.

## Registry entries

Sizes are `runtime.MemStats.HeapAlloc` after a GC, before and after the fill, on the large fixture (`BenchmarkCacheSizes`, Apple M1, 2026-11-11); Swift sizes are not measured. Keys are the keys the entry holds after the fill.

| Cache | Owner (path · symbol) | Key | Fill | Updated by | Lifetime | Cap | Size at 4,000 / 16,000 persons |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `canonlinks.kinship` | `core/database/canonlinks/reader.go` · `pivots` | event or person handle id | `Reader.pivot`: the caller's derive closure, through `Keyed.Get` | `DerivedFrom` the vocabulary; Apply drops the commit's roots and their cached neighbours | graph | 65,536 | 14,162 / 14,162 keys; 1.4 / 4.4 MB with `canonlinks.for` and `canonlinks.siblings` |
| `canonlinks.for` | `core/database/canonlinks/reader.go` · `forAnswers` | person handle id | `Reader.For`, through `Keyed.Get` | `DerivedFrom` the vocabulary; Apply empties it on any commit that names a handle (known exception) | graph | 65,536 | 8,115 / 20,115 keys |
| `canonlinks.siblings` | `core/database/canonlinks/reader.go` · `siblingAnswers` | person handle id | `Reader.LoadSiblings` and `Reader.Relationships`, through `Keyed.GetMany` | `DerivedFrom` the vocabulary; Apply empties it on any commit that names a handle (known exception) | graph | 65,536 | 8,115 / 20,115 keys |
| `evidencegraphs.index` | `core/database/evidencegraphs/index.go` · `Index` | one value | Build: four queries (every graph row, every link, the Subject and Observation counts), on first read | Apply re-lists the Sources in `effects.Set.Reads` `evidence_graphs` and recounts the graphs in `evidence_graph_progress`, then `Set` | graph | none | 12.4 KB / 12.4 KB on the large fixture (`BenchmarkCacheSizes`); 5.2 MB and 97 ms to build at 3,000 Sources, 10,000 graphs, 50,000 Observations (`BenchmarkGraphIndex`) |
| `graphfill.graphs` | `core/database/graphfill/load.go` · `Graphs` | Evidence graph id | `graphfill.EvidenceGraph`: `LoadEvidenceGraph` (head, Subjects, Observations with their Citation's label, page, Artifact and uncertainty, members, title edges) | Apply drops the graphs in `effects.Set.EvidenceGraphs`; the least recently used leaves at the Cap | graph | 64 | see `evidencelayer.layers` |
| `graphfill.records` | `core/database/graphfill/records.go` · `Records` | Evidence graph id | `graphfill.RecordsOf`: from the loaded graph when held, else one Subject, Observation and member query set; each Observation carries its Citation's label, page, Artifact and uncertainty, and the records index Observations by id | Apply drops the graphs in `effects.Set.EvidenceGraphs`; the least recently used leaves at the Cap | graph | 4,096 | 0.70 MB / 0.70 MB for 23 graphs; 22.4 MB and 0.74 s for 4,096 held of 10,000 graphs (`BenchmarkGraphRecords`) |
| `evidencelayer.layers` | `core/database/evidencelayer/evidencelayer.go` · `Layers` | Evidence graph id | `evidencelayer.Load`: the loaded graph, then `Build` | `DerivedFrom` the vocabulary and `nameformats.default`; Apply drops the graphs in `effects.Set.EvidenceGraphs` | graph | 64 | with `graphfill.graphs` and `promotealign.candidates`: 34.5 / 45.0 MB (one proposal; 34.6 / 44.6 MB before, as the store with its memos) |
| `namealiases.index` | `core/database/namealiases/namealiases.go` · `Index` | one value | Build: every alias table, at open (`namealiases.Warm`) or on first read | none; `Invalidate` drops it | catalog | none | 0.77 / 0.77 MB |
| `nameformats.default` | `core/database/nameformats/nameformats.go` · `Default` | one value | Build: the project's default name format, at open (`nameformats.Warm`) or on first read | none; `Invalidate` drops it | catalog | none | 1.5 KB / 1.5 KB |
| `promotealign.candidates` | `core/database/promotealign/canon.go` · `candidates` | Evidence graph id | `loadCandidatesMemo`: `rankCandidates` for the Subjects not yet ranked | `DerivedFrom` the vocabulary and `nameformats.default`; Apply drops the graphs in `effects.Set.EvidenceGraphs`; a memo ranked against an older match index generation is not used | graph | 64 | see `evidencelayer.layers` |
| `promotealign.matchindex` | `core/database/promotealign/index.go` · `MatchIndex` | one value | The owner stores it (`loadShortlist`, the profiler's `keep`); no Build until S16-20 | `DerivedFrom` the vocabulary, `namealiases.index` and `nameformats.default`; Apply drops it on any commit that names a handle (known exception) | graph | none | 6.2 / 9.8 MB |
| `sources.records` | `core/database/sources/records.go` · `Records` | one value | Build: two queries (every Source with its grade, every Artifact), on first read | Apply reloads the Sources in `effects.Set.SourceScopes`, then `Set` | graph | none | 16 KB / 16 KB on the large fixture (one Source); 5.7 MB at 3,000 Sources of 10 Artifacts (`BenchmarkSourceRecords`) |

Past its Cap a keyed entry drops its least recently used key. A whole entry has no cap: it is bounded by what it is built from. On a write's graph (`graphcache.Over`) a catalog-lifetime entry builds once through the write's transaction and lives with that graph; a graph-lifetime entry is neither built nor stored there, and is read through `database.ReadWhole`, which returns `database.ErrWholeOnWrite` for a write's graph and `database.ErrClosed` for a nil one.

## Typed `Graph` fields

These stay fields of `graphcache.Graph`, not registry entries, because each is either the graph itself or has a shape the registry's two kinds do not fit.

| Field | Holds | Why typed | Updated by | Size at 4,000 / 16,000 persons |
| --- | --- | --- | --- | --- |
| `nodes`, `holders`, `kinds`, `kindOn`, `kindSet` | the canonical nodes (values, members, links, outcome rows and the chosen likeness) and the per-type id lists | They are the graph the entries derive from; OnCommit reloads them before any Apply runs | `effects.Set.Handles` (reload), `Structure` (Drop) | 57.0 / 76.9 MB (Person, Event and Place kinds loaded; 45.3 / 58.2 MB before S16-11 added outcome rows) |
| `vocab` | properties, terms, subject types | Read by nearly every feature; the vocabulary is the source other entries declare in `DerivedFrom` | `ListPropertyVocabulary` | part of `nodes` |
| `labels` | the vocabulary label map rows resolve at read time | Stored by `conclusionsummaries`; dropped with the vocabulary | `ListPropertyVocabulary` | not measured |
| `displays` | composed, not finished, summary rows by handle | Verified by `conclusionsummaries`'s display check; dropped by roots and `SummaryDeps` | `Handles`, `SummaryDeps` | 11 KB / 11 KB (`RowsForSearch` keeps none) |
| `activity` | last activity per scope type and scope id | A patched index, not a derived value: OnCommit moves it forward | every commit's `Activity` | not measured |

## Held sessions

| Cache | Owner | Key | Lifetime |
| --- | --- | --- | --- |
| The held catalog (connection, graph and every entry above) | `core/catalogsession/session.go` · `reg` | project directory | until `Close`, `CloseAll`, opening another project, or a panic in `Do` |

## Swift caches

Sizes are not measured. S16-24 – S16-29 reshape these.

| Cache | Owner (path · symbol) | Key | Fill | Updated by | Lifetime | Cap |
| --- | --- | --- | --- | --- | --- | --- |
| Per-view query results | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` · `CatalogQueryKey.Kind.descriptor` (the source of truth for each kind's key, fill and invalidation) | `CatalogQueryKey` | the descriptor's loader | the descriptor's `invalidateOn` mutation kinds | the session | 50 hidden entities (`WorkspaceSession.maxHiddenEntities`) |
| Evidence-graph index (Swift) | `WorkspaceSession.evidenceGraphIndex` | catalog revision | built on first read, patched per Source | `WriteEffects`' revision | the session | none |
| Name format | `WorkspaceSession.nameFormat` | one value | loaded at open | a name format write | the session | none |
| Sidebar counts | `CatalogCounts` | section | `refreshAll` | `WriteEffects.lists` with `counts` | the workspace | none |
| Thumbnails | `ThumbnailCache.shared` | artifact and size | decoded on first draw | none (immutable files) | the process | 100 entries, LRU |
| Visible Sources | `SourcesModel.visibleRowsMemo` | filter and sort token | the model's filter | the token | the model | one entry |
| Citation composer snapshot | `CitationComposerContext.memo` | memo token | the composer's load | the token | the composer | one entry |
| Depiction preview | `CitationComposerModel.depictionPreviewMemo` | cut key | the preview render | the key | the composer | one entry |
| Source search keys | `SourcesModel.searchKeyCache` | row id, with its revision | built per row | the row's revision | the model | one per row |
| Evidence-graph snapshot | `EvidenceGraphModel.snapshotMemo` | rows, types and rules | the snapshot builder | the key | the model | one entry |

## Known exceptions

The drop-on-any-handle Applies of `canonlinks.for` and `canonlinks.siblings` (S16-16 deletes both entries) and of `promotealign.matchindex` (S16-20 patches it instead) are the registry's known exceptions to updating exactly.
