---
name: add-searchable-kind
description: >-
  Registers a Provenencia catalog entity as an omnibar SearchCatalog hit kind
  (core/search registry, location mapper, field weights, FTS projector hooks,
  FakeStore, tests). Use when adding or changing searchable kinds,
  SearchHit.kind strings, WorkspaceLocation mapping for hits, Searcher
  implementations, FTS document projection for a kind, or omnibar-search.md
  registry work — not for omnibar UI chrome alone.
---

# Add a searchable kind

Catalog search is **Go-owned**: declarative kind registry + FTS5 projection +
`Searcher` behind stable `SearchCatalog` FFI. Do **not** put ad hoc `LIKE` SQL
in FFI handlers or build a permanent Swift search engine.

Authoritative behavior: [`docs/deployment-plan/archive/spike-3/omnibar-search.md`](../../../docs/deployment-plan/archive/spike-3/omnibar-search.md).

## Checklist

```
- [ ] Kind id string stable (`source`, `source_type`, `metadata_field`, …)
- [ ] Registry entry: fields + weights (title/ref/label ≫ description ≫ body)
- [ ] Location mapper → WorkspaceLocation (section + deep id)
- [ ] DefaultInEverything / ContextSections for ranking boosts
- [ ] Priority share of the kind tenth (persons 4, events 3, places 2, sources 1, other 0: types, fields, Evidence graphs)
- [ ] FTS projector in core/database/searchindex (Upsert/Delete/Rebuild; unicode61 **and** trigram indexes)
- [ ] A `fieldValuesForKind` case (`core/search/fts.go`) mapping the document's columns onto every registry field the kind declares.
- [ ] Write-path reproject on domain mutators (same tx when practical)
- [ ] Tagged Source body rollup when contributing child text (`note:` / `metadata:` / `filename:`)
- [ ] Swift `CatalogSearchKind` case (`TestSwiftVocabularyMatchesGo` fails without it); the compiler then asks for its `kindLabel` and `leadThumbnail`, and for FakeStore's `searchPriority`
- [ ] FakeStore mirrors kinds + location rules + ref exact/prefix boost; markCatalogSessionHeld; fake hits for the kind (the compiler does not ask for these)
- [ ] Package + FFI + Swift FakeStore tests
```

## Steps

### 1. Registry (`core/search/registry.go`)

Add a `KindSpec`:

| Field | Purpose |
| --- | --- |
| `Kind` | Stable wire id on `SearchHit.kind` |
| `Fields` | Match targets + weights |
| `ContextSections` | Sections that boost this kind (`sources`, `source-types`, …) |
| `ContextBoost` | Multiplier when request location section matches |
| `Priority` | Share of the kind tenth of the score (`ScoreMix`). Persons 4, events 3, places 2, sources 1, other 0 (types, fields, Evidence graphs). The searcher normalizes by the maximum and does not read the name. |
| `Handle` | True for Conclusion handle kinds (person, event, place). Set only by the `handle` helper in `init`; drives member counts and the handle field map. `TestHandleKindsArePrimaryKinds` ties it to `vocabkey.PrimaryKinds()`. |
| `DefaultInEverything` | Include when no facet filter (Spike 3: true for Sources/types/fields) |

Kind constants live in `core/search/search.go`. Section strings must match
macOS `WorkspaceSection` raw values.

FTS documents map fields into columns `title` / `ref` / `secondary` / `body`
(`core/database/searchindex`). Keep weights in this registry — do not fork
weight tables in SQL migrations.

Global knobs in the same file: `FTSBM25Weights` (SQL bm25 columns),
`RefBoostWeights` (exact ≫ prefix ≫ none), and `FuzzyWeights` (trigram
shortlist cap + Jaro–Winkler gate / score scale). Ref-shaped queries use the
docs `ref` column fast path in `FTSSearcher` (hyphens break FTS MATCH).

### 2. Location mapping

Every hit must carry a navigable `WorkspaceLocation` so S3-10 can
`go(to: hit.location)`:

| Kind | Section | Deep id |
| --- | --- | --- |
| `source` | `sources` | `source_id` |
| `source_type` | `source-types` | `type_id` |
| `metadata_field` | `metadata` | `field_id` |
| `evidence_graph` | `sources` | `graph_id`, with `source_surface` `graph` and `source_id` from `display_parent_id` |

A kind that opens inside a parent record projects the parent's id and ref into `display_parent_id` / `display_parent_ref` (and its title into `display_subtitle`); `locationFor` reads them. `locationFor` returns a `core/workspacelocation.Location`, the one Go location type search and delete impact share; a graph hit is built by `workspacelocation.Graph`.

Child text (notes, metadata values, artifact filenames) should **roll into** a
navigable root hit — do not invent note/metadata hit kinds without a
destination UI.

### 3. Projection + retrieval

1. Extend `core/database/searchindex` to build/upsert the kind’s document
   (and delete on remove). Keep **both** `catalog_search_fts` (unicode61) and
   `catalog_search_fts_trigram` in sync on Upsert/Delete/ClearAll. Bump
   `ProjectionVersion` when the document shape or index set changes so
   `searchindex.EnsureCatalog` rebuilds on Open.
2. Add a `Search` doc on the table's `effects` registry entry (with the parent
   fan-out, as `source_types` and `sources` do), a case in
   `searchindex.Reproject`, and a loop in `RebuildAll`. A new display column
   goes through `searchindex.Document`/`Upsert` and `core/search`
   `docColumns`/`scanDoc`. A kind whose document fills `secondary` or `body`
   with a declared field also needs a case in `core/search/fts.go`
   `fieldValuesForKind`; without one the kind falls back to `title`, `ref` and
   `body`, and a declared field such as `description` scores zero with no error.
3. `FTSSearcher` (`DefaultEngine`) reads unicode61 FTS + docs; when the
   candidate map is still under the hit limit, expands via trigram OR shortlist
   + Jaro–Winkler gate (never full-catalog fuzzy). Empty / whitespace query →
   **no hits**; cap at `DefaultHitLimit` (50).

### 4. FFI + Mac store

- Handler already maps core hits ↔ proto (`SearchCatalog`). New kinds need no
  new Method if they fit `SearchHit`.
- Swift: `CatalogSearchHit` + `GenealogyStore.searchCatalog`; FakeStore scores
  in memory with the same kind/location rules; call `markCatalogSessionHeld`.
- Swift spells kinds once, in `macos/App/Platform/CatalogSearchKind.swift`
  (`hit.searchKind`; nil for a kind this build doesn't know). Switches on it
  are exhaustive, so a new case names each view spot. Never compare
  `hit.kind` to a string literal.
- Swift callers search through `CatalogSearchModel` (`macos/App/Features/Catalog/Search/`): kinds at init, an optional location per search (nil: no section nudge). The omnibar and Promote's find-another-record picker both use it; draw hits with `CatalogSearchHitRow`, which carries its own VoiceOver label (the session's name order). A picker is `PVComboBox(onQueryChange:isLoading:empty:)` over `CatalogSearchModel(kinds:)`, as `PromoteRecordSearch` does.

Catalog RPCs use `withProjectCatalog` — see
[`use-catalog-session`](../use-catalog-session/SKILL.md). New Method →
[`add-ffi-handler`](../add-ffi-handler/SKILL.md).

### 5. Tests

| Layer | Cover |
| --- | --- |
| `core/search` | Ranking (title > description), note rollup, context boost, location, empty query, EnsureCatalog heal, typo shortlist |
| `core/database/searchindex` | Migration / FTS + trigram smoke |
| `api/ffi/handlers` | `runRPC` hit on new kind; location populated |
| `ProvenenciaTests` | FakeStore title/ref hit + location ([`add-swift-test`](../add-swift-test/SKILL.md)) |

```sh
CGO_ENABLED=1 go test -tags fts5 ./core/search/... ./core/database/searchindex/... ./api/ffi/...
```

## Do not

- Add searchable kinds only in Swift / FakeStore without Go registry
- Return hits without a `WorkspaceLocation` the Mac can `go(to:)`
- Open the catalog outside `withProjectCatalog` / `catalogsession`
- Ship Files / Artifact / Interpretation kinds before destinations exist
- Reintroduce a live-table naïve scan as the product Searcher

## Related

- Hit navigation: [`add-workspace-location`](../add-workspace-location/SKILL.md)
- Migrations: [`add-catalog-migration`](../add-catalog-migration/SKILL.md)
- Omnibar UI: Spike 3 S3-10; the shared search model and hit row: `macos/App/Features/Catalog/Search/` (S11-37)
