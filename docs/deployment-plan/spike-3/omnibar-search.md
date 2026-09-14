# Omnibar search

**Status:** pulled into [Spike 3](README.md) — requirements. PR sequence: [`deployment-plan.md`](deployment-plan.md).

Visual chrome for the **field** (placement + access): Claude Design **App Layout** board. Results dropdown: Claude Design **Omnibar Results** board (brief archived under [`design/archive/`](design/archive/); summary [`design/README.md`](design/README.md)). The [Display ideas](#display-problem) section below is historical; implement against the board.

Engine / ranking architecture is **in scope** and first-class (not a client-side `LIKE` forever). Ship it in **successive PRs**; the end state below is the product bar.

## Problem

As the workspace grows (Sources, vocabulary, Files, later Interpretation and Conclusion), sidebar destinations become a slow way to *find* something you already know exists. Researchers often remember a title fragment, a filename, a person name, or a short ref (`SRC-…`, `PER-C-…`) and want to jump there without hunting through menus and nested lists.

Today each destination carries its **own** list search (Sources list filter; Source fields / types `VocabularyListPane` search on label/key/description). That fragments findability, duplicates chrome, and still fails when the researcher is unsure *which* destination owns the thing, or when the thing lives several clicks deep (Artifact under a Source, Observation on a Node).

## Goal

A single **omnibar** in the workspace toolbar: one search field that queries across catalog entities and offers **navigable hits**. Choosing a hit jumps via workspace `go(to:)` (see [`navigation-history.md`](navigation-history.md)) into the right destination and deep location.

Think command-palette / Spotlight energy, scoped to the **open project** — not web search, not Settings.

Search quality bar:

- **Weighted** field matching (title/ref beat notes)
- **Context-aware** ranking (boost kinds that match the current `WorkspaceLocation`)
- **Token / prefix** matching via SQLite **FTS5**
- **Typo / fuzzy** tolerance on a shortlist (trigram and/or Go fuzzy) — in scope, not a forever follow-up
- **Exact / prefix ref** as a first-class fast path (`SRC-F4N2P` → obvious top hit)

### Remove per-destination search

**Destroy** the existing in-destination search UI and the model/query plumbing that exists only to drive it:

- Sources list search field / clear-search / “no matches for query” empty state tied to that query
- Source fields and Source types search chrome in `VocabularyListPane` (and matching `query` filter behavior used only for that bar)

Findability moves entirely to the omnibar. Do not leave a second, quieter filter bar “just in case.” Type filters, sort controls, and other non-search list chrome can remain where they still earn their keep. Timing: remove old search when the omnibar can actually find those entities (same PR train, not months later).

## Implementation posture (first-class search)

Early product: **prefer correct Go/SQLite search infrastructure over a Swift filter that we know we will throw away.** Broad wiring and several PRs are fine; a permanent brute-force `LIKE` across live tables is not the destination.

- **Engine owns search** — One catalog-open RPC; ranking and column policy live in Go. Swift renders hits and calls `go(to:)`.
- **Registry, not ad hoc SQL per screen** — New searchable kinds register fields, weights, and location mapping.
- **Index in the project DB** — FTS projection travels with the `.provenencia` package; incremental upkeep on writes; rebuild/heal path for migrations and corruption.
- **No second find surface** — Omnibar replaces per-destination search.
- **Evaluate as we go** — Incremental PRs (below) so dogfood can reject ranking/UX before the next slice lands.

Aligned with [`application-stack.md`](../../application-stack.md) (FTS5 build tag; search index rebuild as a long-running task class).

## Chrome (from App Layout board)

Toolbar sits in the **main column** (shared with Back/Forward and breadcrumbs — see navigation history). Omnibar is the **trailing** control:

```text
[ ← ] [ → ]   Breadcrumbs …                    [ 🔍  Search …          ⌘K ]
```

| Control | Requirement |
| --- | --- |
| **Placement** | Always-visible search field, top-**right** of the main-column toolbar (~420px wide / max ~60% in the board). Search icon leading; placeholder copy per board (e.g. people, sources, places, files). |
| **⌘K** | Keyboard focuses / activates the omnibar (suffix hint on the field in the board). Same field — not a separate hidden-only palette. |
| **Results dropdown** | Opens from the field while searching per Omnibar Results board (rich `PVOmnibarHitRow`, flat rank, ~640px widen-left). Behavioral: navigable hits; selecting one calls `go(to:)` and dismisses. Debounce typing before RPC; empty/short query does not open a panel. |

## What should be searchable (horizon)

Grouped by layer; not all exist in the shipped catalog yet:

| Kind | Layer | Typical match fields | Ref? |
| --- | --- | --- | --- |
| Source types | Source (vocab) | `label`, `key`, `description` | no |
| Source fields | Source (vocab) | `label`, `key`, `description` | no |
| Sources | Source | `title`, `ref`, type label, `description`, notes, metadata text | `SRC-…` |
| Artifacts | Source | `ref`, `description`, primary File filename | `ART-…` |
| Files | Source | `original_filename`, `media_type` | no |
| Citations | Interpretation | `ref`, transcription / quote / page label, description | `CIT-…` |
| Observations | Interpretation | `ref`, property label + value text / name form / date | `OBS-…` |
| Nodes (candidates) | Interpretation | `ref`, `label`, type, description | `{P}-C-…` |
| Canonical entities | Conclusion | `ref`, working `label`, projected names | `{P}-…` |
| Relationships | Conclusion | mostly claims/views over entities — weak standalone identity today | usually none |

Contributors (`USR-…` / display name) are optional later if “who added this” becomes a jump target.

**Spike 3 first kinds:** Sources, source types, source fields (UI exists). Files / Artifacts as soon as project-wide listing or projection allows. Later layers register when their destinations ship — same pipeline.

## Search architecture (end state)

```text
Swift omnibar (query + current WorkspaceLocation)
  → SearchCatalog RPC  (one catalog open)
       → ref fast path (exact / prefix on ref column)
       → FTS5 over projected search documents
       → optional fuzzy pass on shortlist (trigram / Go)
       → rank: retrieval score × field weights × context boosts × ref boosts
       → Hit DTO[] (kind, id, ref?, title, subtitle?, match_reason?, location)
  → go(to: location)
```

### Searchable-kind registry (Go)

Declarative registry (one entry per kind), not scattered `LIKE` in handlers:

| Registry field | Role |
| --- | --- |
| `kind` | Stable id (`source`, `source_field`, …) — the **hit** kind, almost always a **navigable root** |
| **Projected fields** | Own columns **plus rolled-up text from child tables** that should not be separate hits |
| **Weights** | Relative importance (e.g. title/ref ≫ description ≫ notes/metadata body ≫ filename) |
| **Default in “everything”** | Whether the kind participates without a facet |
| **Location mapper** | Build `WorkspaceLocation` (`section` + per-kind deep id) for `go(to:)` |
| **Display stubs** | Enough to fill Hit DTO title/subtitle/kind label without a second fetch when possible |

Adding a kind later = register + projector + FTS upkeep on that table’s writes — not a new Mac search implementation.

### Navigable roots vs contributory text

FTS5 does **not** invent entities. It indexes **documents you project**. Provenencia should project documents around **what the researcher can open**, not every child row.

| Role | Examples | In search index? | Omnibar hit? |
| --- | --- | --- | --- |
| **Navigable root** | Source, Source field, Source type, (later) Node, canonical Person | Yes — one document per root | **Yes** — `kind` + `location` open that root |
| **Contributory / associated** | `source_notes`, `source_metadata` values, Artifact label/filename, File `original_filename` under a Source | Text **rolled into the parent Source document** (weighted lower than title/ref) | **No** separate hit for “the note” or “the metadata cell” — matching that text returns the **Source** |
| **Optional first-class hit** | Artifact (when we want jump-to-artifact-on-Source-page) | Either roll into Source **or** own document whose `location` is still Source-scoped | Only if UX wants Artifact as its own row; still not a free-floating note entity |

So a query that matches a word only in a note or metadata value still yields a **Source** hit (title/`SRC-…` in the dropdown), ideally with `match_reason` like “note: …” or “metadata author: …”. The researcher never lands on an orphan note entity — those tables are not destinations.

**Registry implication:** the `source` entry lists:

- Own fields: `title`, `ref`, `description`, type label, …
- **Associated projectors:** concatenate or column-feed notes bodies, metadata display text, artifact descriptions, primary filenames into the Source document (with lower weights)

**Write-path implication:** changing a note or metadata value **reprojects that Source’s** search document (same for artifact/file attach). Rebuilding walks roots and joins children.

Vocabulary rows (`source_fields`, `source_types`) are their **own** roots — they are sidebar destinations, not children of a Source.

Same pattern later: Observation text may roll into a **Node** or **Citation** root depending on what `go(to:)` can open; do not index every join table as a hit kind by default.

### FTS5 projection (SQLite)

- Maintain a **search document** projection inside `provenencia.sqlite` (FTS5 virtual table + side table for `kind` / `entity_id` / display fields as needed).
- Enable **`fts5`** on the embedded amalgamation ([`application-stack.md`](../../application-stack.md) §10).
- **Incremental updates** in the same transaction as domain writes when practical; **rebuild/heal** on Open/migrate when the projection lags or schema changes (long-running rebuild may be async later; sync rebuild OK while catalogs are small).
- Do **not** rely on caching whole tables in Swift as the source of truth. Optional debounce / in-flight request cancellation only.

Brute-force scan of live tables is **not** the end state. A temporary stub behind `SearchCatalog` during an early PR is acceptable only if the Hit DTO and RPC stay stable and the next PR swaps in FTS without rewriting the omnibar.

### Ranking and context

1. **Retrieve** candidates (ref path + FTS; fuzzy expands the set carefully).
2. **Score** in Go (or SQL helpers), conceptually:

   `score = retrieval_rank × field_weight × context_boost × ref_boost × term_coverage`

3. **Context:** pass current `WorkspaceLocation` (at least `section`). Boost matching kinds — e.g. on Sources / Source page, prioritize Source hits; on Source fields, prioritize fields; still include other kinds at lower weight so the omnibar remains global.
4. **Ref boost:** exact ref ≫ prefix ref ≫ FTS-only match.
5. Return a bounded list (e.g. top 20–50) for the dropdown; UI may show fewer.

### Multi-word queries

FTS5 tokenizes both the query and each document. `"John Smith's birth certificate"` becomes tokens roughly like `john`, `smith`, `birth`, `certificate` (apostrophe/punctuation handling depends on tokenizer).

**Within one document (the common case):**

| Candidate | Why it ranks |
| --- | --- |
| Source titled “John Smith's birth certificate” | Many/all tokens in a **high-weight** field (title) → strong `retrieval_rank` + field weight |
| Source type “Birth certificate” | Only `birth` + `certificate` → lower **term coverage**; still a valid hit if we don’t require every token |
| Person Node “John Smith” (later) | Only `john` + `smith` → partial coverage; competing with Sources |

Default FTS `MATCH` is typically **AND** (all tokens must appear in that document). Strict AND alone would **drop** the type and the person for this query, leaving mainly the full Source title — often good, sometimes too harsh.

**Product rule for Provenencia:**

1. Prefer documents that match **more tokens** and match them in **higher-weight fields** (title/ref ≫ body).
2. Allow **partial term coverage** (or a controlled OR / “best effort” rewrite) so useful narrower kinds still appear, but **rank them below** fuller matches.
3. Apply **context boosts** on top: if the researcher is on Sources / a Source page, multiply Source hits; if on Source types, multiply type hits; later on People, multiply person hits. Context changes **order**, not “only this kind” (omnibar stays global unless the user facets).
4. Phrase / nearness (e.g. `"birth certificate"` as a phrase, or tokens close together) can further boost the Source title over scattered token hits — tune in dogfood.

Example ordering for `John Smith's birth certificate` while browsing **Sources**:

1. Source “John Smith's birth certificate” (full coverage + title + context)  
2. Other Sources that mention both name and “birth certificate” in notes/metadata (rolled into Source doc, lower field weight)  
3. Source type “Birth certificate” (partial coverage; lower unless context is types)  
4. Person “John Smith” when that kind exists (partial coverage; higher if context is People)

### Associated entities across the graph (later)

“A person named John Smith **who has** a birth certificate associated” is **not** solved by FTS alone on isolated documents. That needs either:

- **Denormalized related labels** on the root document (e.g. Source doc also contains linked person names once Interpretation links exist), and/or  
- A **second-stage** association query after FTS shortlist  

Roll child **Source-layer** text (notes, metadata, files) into the Source document now. **Cross-root** association (Person ↔ Source) is a later slice when those links exist — don’t pretend multi-word FTS invents joins between separate hit kinds.

### Fuzzy / typo matching (in scope)

Ship real typo tolerance, phased after solid FTS + weights:

| Capability | Approach |
| --- | --- |
| Token / prefix (`par` → parish) | FTS5 (unicode61 / prefix queries) |
| Exact / prefix **ref** | Dedicated lookup on `ref` (not FTS alone) |
| Typos (`Ilminstr` → Ilminster) | FTS5 **trigram** and/or Go fuzzy (e.g. Jaro–Winkler / Levenshtein) on a **shortlist** of FTS or trigram candidates — never fuzzy-scan the entire catalog row-by-row |

Tune thresholds in dogfood so typos help without turning every query into noise.

### Hit DTO (FFI)

Normalized hit for all kinds (stable across PR slices):

- `kind`, `id` (UUID), `ref?`, `title`, `subtitle?`, `match_reason?` (optional “why”), `location` (payload aligned with navigation `WorkspaceLocation` / per-kind deep fields)

Swift does not re-implement ranking.

### Artifacts and weak identities

**Default:** Artifact/File text contributes to the owning **Source** document (see Navigable roots). Optionally later, register Artifact as its own hit kind whose location opens the Source page focused on that Artifact — still not a separate app destination for notes/metadata.

Relationships without refs: poor primary hits — prefer related Person/Event entities until the model gives clearer identity.

## Display problem

Hits are **heterogeneous**. A Source wants title + type + thumbnail + `SRC-…`. A File wants filename + media type. A Person Node wants name / label + `PER-C-…`. A vocabulary field wants label + key + data type.

**Results UI is locked** by the Claude Design Omnibar Results board: shared rich row skeleton + flat ranked list. The ideas below remain useful rationale; implement against the board / `PVOmnibarHitRow`.

### Display ideas (board-aligned)

**1. Shared skeleton, kind-specific slots** — shipped direction on the board.

```text
[icon / thumb]  primary title                 kind · secondary
                tertiary / match context        REF-…
```

Same spacing and typography; only the slots change. Avoid per-kind card layouts inside the results list.

**2. Grouped results** — optional exploration only; board default is flat.

**3. Faceted omnibar** — later; not Spike 3 chrome.

**4. Rank, then unify** — flat list ordered by engine score; kind chip disambiguates (**board default**).

**5. Two densities** — not required by the board; denser rich rows are the v1.

## Incremental delivery

Break into successive PRs so each slice is dogfoodable. Sequenced as **S3-07…S3-11** in [`deployment-plan.md`](deployment-plan.md) (S3-07 / S1 and S3-08 / S2 done — [`completed.md`](completed.md)); intended order:

| Slice | Delivers | Evaluate |
| --- | --- | --- |
| **S1 — Registry + RPC shell** | **Done (S3-07).** Searchable-kind registry; `SearchCatalog` protobuf/FFI; Hit DTO; FakeStore; context location on the request. Naïve scanner **behind the same API** as bridge. | RPC shape; location mapping; tests without UI. |
| **S2 — FTS5 projection** | **Done (S3-08).** Migration + FTS documents for Sources / types / fields; incremental upkeep on writes; rebuild/heal; FTS retrieval + field weights; Source child text rolled into Source docs. | Latency; relevance on real dogfood catalogs; index correctness after edits. |
| **S3 — Context ranking + ref fast path** | Section/kind boosts; exact/prefix ref promotion; match_reason where cheap. | “I’m on Sources → Sources float” feels right; paste-ref UX. |
| **S4 — Omnibar chrome + remove list search** | Toolbar field + ⌘K (S3-05 shell); results UI per board (S3-10); wire hits to `go(to:)`; **delete** per-destination search. | End-to-end find; no dual search chrome. |
| **S5 — Fuzzy / typo** | Trigram and/or Go fuzzy shortlist pass; tune thresholds. | Typos recover without garbage. |
| **S6+ — More kinds / depth** | Files, Artifacts projection; deeper note/metadata/transcription weight tuning; Interpretation when UI exists. | Noise vs recall; registry extensibility. |

S4 can overlap S2/S3 if chrome is blocked on Design for the dropdown — field chrome can ship with a simple list against the Hit DTO before visual polish.

## Why it fits Provenencia

Local-first catalogs invite keyboard navigation. Short refs were designed to be citable and memorable — an omnibar is the natural consumer. A single project-scoped search also reinforces that Evidence / Interpretation / Conclusion are one workspace, not separate products. FTS in the project file keeps search portable with the catalog across machines.

## Resolved decisions

| Question | Decision |
| --- | --- |
| Where does the field live? | **Main-column toolbar, trailing** (App Layout board). |
| Always-visible vs ⌘K-only? | **Both:** always-visible field; **⌘K** focuses/activates it. |
| Per-destination list search? | **Remove** when omnibar covers those kinds. |
| Results dropdown visuals? | **Board locked** — Claude Design Omnibar Results; shared rich row + flat ranked list. |
| Hit navigation? | **`go(to:)`** — same session history as sidebar / breadcrumbs. |
| Engine? | **Go `SearchCatalog` + kind registry + FTS5 projection** in the catalog DB — not Swift-as-search-engine. |
| Brute force forever? | **No.** Naïve scan was only a short bridge (S3-07); FTS5 projection is the engine (S3-08). |
| Context-aware ranking? | **Yes** — boost by current workspace section / location. |
| Multi-word queries? | **Yes** — tokenize; prefer higher **term coverage** + high-weight fields; allow partial matches at lower rank; context reorders kinds. |
| Cross-root “associated with”? | **Later** — denormalize related labels and/or second-stage joins; not implied by multi-word FTS alone. |
| Field weights? | **Yes** — declared per kind in the registry. |
| Child tables (notes, metadata, files)? | **Roll into navigable root** (e.g. Source) — match returns the Source, not a note/metadata entity. |
| Fuzzy / typo matching? | **Yes, in scope** (phased after FTS + weights + ref path). |
| Delivery? | **Successive PRs** per incremental table above. |

## Open questions

- Should vocabulary (types/fields) appear in the default “everything” query at full weight, or slightly demoted unless context is vocabulary?
- Thumbnail cost: show Source/File thumbs in the palette, or icons only until selected?
- Empty query: recent destinations / recent entities (from navigation history?), or blank until type?
- How aggressively to weight note bodies / metadata / (later) transcriptions by default?
- Candidate `PER-C-…` vs canonical `PER-…` when Interpretation/Conclusion search lands — same list with a layer badge, or separate sections?
- Sync vs async FTS rebuild when catalogs get large?

## Explicitly out of scope

- Searching audit history, settings, or help docs
- Natural-language / AI “ask the catalog” queries
- Cross-project search
- Replacing SQLite with an external search server

## Related docs

- [`deployment-plan.md`](deployment-plan.md) — sequenced PRs
- [`design/README.md`](design/README.md) — App Layout + Omnibar Results boards
- [`navigation-history.md`](navigation-history.md) (Back/Forward after omnibar jumps; shared toolbar; `WorkspaceLocation`)
- [`application-stack.md`](../../application-stack.md) (FTS5; index rebuild task class)
- [`catalog-refs.md`](../../catalog-refs.md)
- [`source-layer-data-model.md`](../../source-layer-data-model.md)
- [`artifact-file-storage.md`](../../artifact-file-storage.md)
- [`interpretation-layer-data-model.md`](../../interpretation-layer-data-model.md)
- [`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md)
- [`macos-client-patterns.md`](../../macos-client-patterns.md)
- Historical briefs that assumed local list search (superseded for find chrome): [`S2-04-sources-list.md`](../archive/spike-2/design/archive/S2-04-sources-list.md); [`S2-20-files-list.md`](../archive/spike-2/design/archive/S2-20-files-list.md)
