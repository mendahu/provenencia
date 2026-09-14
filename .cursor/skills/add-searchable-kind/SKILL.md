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

Authoritative behavior: [`docs/deployment-plan/spike-3/omnibar-search.md`](../../../docs/deployment-plan/spike-3/omnibar-search.md).

## Checklist

```
- [ ] Kind id string stable (`source`, `source_type`, `source_field`, …)
- [ ] Registry entry: fields + weights (title/ref/label ≫ description ≫ body)
- [ ] Location mapper → WorkspaceLocation (section + deep id)
- [ ] DefaultInEverything / ContextSections for ranking boosts
- [ ] FTS projector in core/database/searchindex (Upsert/Delete/Rebuild)
- [ ] Write-path reproject on domain mutators (same tx when practical)
- [ ] FakeStore mirrors kinds + location rules; markCatalogSessionHeld
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
| `DefaultInEverything` | Include when no facet filter (Spike 3: true for Sources/types/fields) |

Kind constants live in `core/search/search.go`. Section strings must match
macOS `WorkspaceSection` raw values.

FTS documents map fields into columns `title` / `ref` / `secondary` / `body`
(`core/database/searchindex`). Keep weights in this registry — do not fork
weight tables in SQL migrations.

### 2. Location mapping

Every hit must carry a navigable `WorkspaceLocation` so S3-10 can
`go(to: hit.location)`:

| Kind | Section | Deep id |
| --- | --- | --- |
| `source` | `sources` | `source_id` |
| `source_type` | `source-types` | `type_id` |
| `source_field` | `source-fields` | `field_id` |

Child text (notes, metadata values, artifact filenames) should **roll into** a
navigable root hit — do not invent note/metadata hit kinds without a
destination UI.

### 3. Projection + retrieval

1. Extend `core/database/searchindex` to build/upsert the kind’s document
   (and delete on remove). Bump `ProjectionVersion` when the document shape
   changes so `searchindex.EnsureCatalog` rebuilds on Open.
2. Call reproject from domain mutators in the same transaction when practical.
3. `FTSSearcher` (`DefaultEngine`) reads `catalog_search_fts` + docs; empty /
   whitespace query → **no hits**; cap at `DefaultHitLimit` (50).

### 4. FFI + Mac store

- Handler already maps core hits ↔ proto (`SearchCatalog`). New kinds need no
  new Method if they fit `SearchHit`.
- Swift: `CatalogSearchHit` + `GenealogyStore.searchCatalog`; FakeStore scores
  in memory with the same kind/location rules; call `markCatalogSessionHeld`.

Catalog RPCs use `withProjectCatalog` — see
[`use-catalog-session`](../use-catalog-session/SKILL.md). New Method →
[`add-ffi-handler`](../add-ffi-handler/SKILL.md).

### 5. Tests

| Layer | Cover |
| --- | --- |
| `core/search` | Ranking (title > description), note rollup, context boost, location, empty query, EnsureCatalog heal |
| `core/database/searchindex` | Migration / FTS smoke |
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
- Omnibar UI: Spike 3 S3-10 (not this skill)
