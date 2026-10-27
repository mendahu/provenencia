# S12-15 — Graphs in the omnibar

**Status:** Draft
**Spike / slice:** Spike 12 · Slice 4 — Graphs in the UI
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** — (no new chrome: a graph hit is one more kind in the existing `PVOmnibarHitRow`, with a kind chip label and the `gitBranch` glyph the Source page's Open button already uses)
**Board:** —
**Depends on:** S12-08 (landed [#528](https://github.com/mendahu/provenencia/pull/528))
**Blocks:** —
**Branch:** `s12-15-graph-search`
**Version bump:** none — Spike 11 and 12 PRs leave `VERSION` at 0.2.0 until a release is cut, as S12-01 to S12-14 did. The catalog format moves 55 → 56 (one additive migration, D-13); that is `user_version`, not the product version ([`docs/versioning.md`](../../../versioning.md#project-format-version))
**Skills to load:** [`add-searchable-kind`](../../../../.agents/skills/add-searchable-kind/SKILL.md), [`add-catalog-migration`](../../../../.agents/skills/add-catalog-migration/SKILL.md), [`add-workspace-location`](../../../../.agents/skills/add-workspace-location/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/searchable-kinds.mdc`, `.cursor/rules/catalog-migrations.mdc`, `.cursor/rules/go-tests.mdc`, `.cursor/rules/macos-client.mdc`, `.cursor/rules/macos-l10n.mdc`, `.cursor/rules/swift-tests.mdc`
**Context:** [omnibar contract](../../archive/spike-3/omnibar-search.md), [graph data model](../source-structure.md#data-model), the user's call (2026-10-27, review M-4) that a hit's parent record is projected onto its search document, not read at search time; working rules: "Omnibar", "A graph is a place", "One graph list", "A re-keyed symbol is renamed in the PR that re-keys it" (the name `EvidenceGraph`), "Each PR updates the docs and skills it makes false"
**Planned:** 2026-10-27 against `main` @ `b11d1cba`
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

A researcher types a graph's name, a word of its description, or its `GPH-` ref in the omnibar and gets an Evidence graph hit: the graph's name, the kind chip "Evidence graph", its Source's title and its ref. Return opens that graph's canvas, exactly as the graph's button in the Sources list does, so Back / Forward and restore return to it. This is bar item 7 ("The omnibar finds a graph by name, description or `GPH-` ref") and the last PR of slice 4.

**Why now.** S12-08 made a graph a place (`WorkspaceLocation.graphId`, `.evidenceGraph(…)`), and the Go search location already carries `GraphID` with nothing setting it (S12-08 STEP-6: "no search hit sets it until S12-15"). S12-11 to S12-14 let a Source hold several graphs, so the researcher now has graphs worth jumping to by name.

**What stays.** Source, type, field, Person, Event and Place hits score, rank and open as on `main`. A Source's document does not gain its graphs' text: a graph's name finds the graph, not its Source. The omnibar panel, `PVOmnibarHitRow`, `CatalogSearchModel` and Promote's find-another-record picker (kinds person / event / place) do not change. No FFI method, message or field is added. The one schema change is two display columns on the search index table, filled for graphs only.

## 2. User stories

**US-1.** As a researcher, when I type part of an Evidence graph's name, description or ref in the omnibar, I see that graph as a hit, so that I can reach one household of a census without opening its Source first.

Fixture for every AC: Source "1881 Census, Ward 3" (`SRC-W3CEN`) with Artifacts "Page 1" and "Page 2" and graphs "Default", "Jones household" (description "Thomas Jones and family, page 2", ref `GPH-J0NES`) and "Smith household".

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Type `jones hou` | A hit with title "Jones household", kind chip "Evidence graph", secondary line "1881 Census, Ward 3", ref `GPH-J0NES`, lead glyph `gitBranch`, and no match-context line |
| AC-1.2 | Type `thomas` (only in the description) | The Jones household hit, with the match-context line "Description" |
| AC-1.3 | Type `GPH-J0NES` (exact ref) | The Jones household hit is first, ref accented; `gph-j0` (prefix, any case) also returns it |
| AC-1.4 | Type `ward` (only in the Source's title) | The Source hit; no graph hit (graphs match their own name, description and ref only) |
| AC-1.5 | Rename "Jones household" to "Jones family", then type `jones fam` | The graph hit reads "Jones family"; `jones hou` returns no graph hit |
| AC-1.6 | Rename the Source to "1881 Census, Ward 3 (Toronto)", then type `jones` | The graph hit's secondary line reads "1881 Census, Ward 3 (Toronto)" |
| AC-1.7 | Delete "Smith household", then type `smith hou` | No graph hit |
| AC-1.8 | A Source "Empty deed" with no Artifact (one Default graph, ref `GPH-EMPTY`); delete that Source, then type `GPH-EMPTY` | No hit |
| AC-1.9 | Type `default` | One "Default" hit per Source, each with its Source's title on the secondary line |
| AC-1.10 | A project opened by a build before this PR (search index at projection version 8) | On open the index rebuilds once and `jones hou` returns the graph hit |

**US-2.** As a researcher, when I choose an Evidence graph hit, its canvas opens, so that the omnibar is one more way into the graph.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Return on the Jones household hit | The canvas opens on Jones household's graph id; the breadcrumbs read Sources › 1881 Census, Ward 3 › Jones household; the omnibar clears |
| AC-2.2 | AC-2.1, then Back, then Forward | Back returns to the place the search started from; Forward returns to the Jones household canvas |
| AC-2.3 | AC-2.1, quit and relaunch | The workspace restores to the Jones household canvas |
| AC-2.4 | A graph that links no Artifact (a Source's Default graph after its only Artifact was deleted); Return on its hit | Its canvas opens with the "no artifact" callout (`evidenceGraph.noArtifact.callout`) that S12-13 shows; nothing is refused |

## 3. Scope

| In | Covered by |
| --- | --- |
| Two parent display columns on `catalog_search_docs` (migration `000056`), written by `searchindex.Upsert` and read by every hit query | STEP-1, AC-2.1 |
| An `evidence_graph` search document per graph (name, ref, description; its Source's id, ref and title as display stubs), kept current by every graph write and every Source rename through `effects`, and healed by a projection version bump | STEP-2, AC-1.5, AC-1.6, AC-1.7, AC-1.8, AC-1.10 |
| The `evidence_graph` searchable kind: registry entry, field map, and a location built from the document's parent columns | STEP-3, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.9 |
| The proto comments that list the hit kinds, regenerated | STEP-4 |
| The kind chip, lead glyph and VoiceOver label of a graph hit | STEP-5, AC-1.1, AC-1.2 |
| FakeStore graph hits that follow the same rules, opening the location the Sources list's button opens | STEP-6, AC-1.1, AC-1.2, AC-1.3 |
| A graph hit opens the canvas through `navigation.go(to:)` | STEP-3, STEP-6, AC-2.1, AC-2.2, AC-2.3, AC-2.4 |
| Docs and skills the PR makes false: the omnibar contract's kind tables, `add-searchable-kind`'s location table and parent columns, the map's "One graph list" rule | STEP-7 |

| Out | Home |
| --- | --- |
| A Source's document rolling in its graphs' names, so a graph name finds the Source | not planned (the graph is its own hit) |
| Graph hits matching their Source's title | not planned (AC-1.4) |
| Hiding the Default graph of single-graph Sources from search | not planned (D-7) |
| A per-kind facet or filter in the omnibar | not planned ([omnibar contract](../../archive/spike-3/omnibar-search.md): the omnibar stays global) |
| Artifact, Citation or Observation hit kinds | not planned ([omnibar contract](../../archive/spike-3/omnibar-search.md) "What should be searchable") |
| One rule for `Title` / `Ref` on a graph-surface location across `deleteimpact`'s projectors (`projectEvidenceGraph` carries the graph name, `projectSubject` the Subject's) | suggested follow-up task (review M-1); the breadcrumb reads the session's Source title first, so neither shows today |
| Parent columns for kinds other than `evidence_graph` | not planned (no other kind opens inside a parent record) |
| A typed Swift `CatalogSearchKind` enum pinned to `search.Registry` | suggested follow-up task (review M-6); all seven kinds share the string literals |
| Map row status of S12-14 (merged in #576, map still says "In review") | S12-14's own DoD |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Kind ids | `core/search/search.go` consts `KindSource` … `KindPlace`; `SectionSources` … | No graph kind. `WorkspaceLocation.GraphID` and `SourceSurface` exist (S12-08) and no hit sets them. No `SourceSurface` constant in `core/search` |
| Registry | `core/search/registry.go` `Registry`, `init` (handle kinds) | Six kinds. Priorities: person 4, event 3, place 2, source 1, types and fields 0 |
| Field map | `core/search/fts.go` `fieldValuesForKind` | Cases for handles, `KindSource`, `KindSourceType` / `KindMetadataField`; `default` maps `title` / `ref` / `body`, so a graph's description (stored in `secondary`) would never score |
| Location | `core/search/fts.go` `locationFor(kind, id, refCol, title)`, called once in `Search` | One case per kind from the document row alone; a graph location needs its Source id and ref, which no column holds |
| Doc rows | `core/search/fts.go` `docRow`, the scans in `mergeFTS` and `mergeTrigramCandidates`; `core/search/refpath.go` `lookUpRefDocs` | Four `SELECT` texts (`mergeFTS`, `mergeTrigramCandidates`, and the exact and `LIKE` branches of `lookUpRefDocs`) each list the same eleven `catalog_search_docs` columns, and three `rows.Scan` calls read them positionally into `docRow` (`mergeFTS` adds `ftsRank`) |
| Search table | `core/database/migrations/000018.sql` `catalog_search_docs`, `000020.sql` (display stubs `display_icon_key`, `display_thumbnail_rel_path`); latest step `000055.sql`; `catalogmodel` `{Name: "catalog_search_docs", Bucket: BucketSkip}` | No column names a hit's parent record. `000020` is the precedent: display columns `TEXT NOT NULL DEFAULT ''` "for hit hydration without a second catalog round-trip" |
| Ref fast path | `core/search/refpath.go` `classifyRefQuery`, `lookUpRefDocs` | Prefix-agnostic (`[A-Z]{3}-…` and `ref.Valid`); a `GPH-` query already hits any document whose `ref` column holds it |
| Projector | `core/database/searchindex/searchindex.go` consts `KindSource` …, `ProjectionVersion = 8`, `Document`, `Upsert`, `Reproject` (switch on `effects.Doc*`), `ReprojectSourceType`, `ReprojectSourcesForType`, `RebuildAll` | No graph document; `Reproject` returns `unknown search document` for any other kind. `ReprojectSourcesForType` is the precedent for a parent label on a child document |
| Effects | `core/database/effects/registry.go` consts `DocSource`, `DocSourceType`, `DocMetadataField`; `registry["evidence_graphs"]`, `registry["sources"]`; `onField`, `graphsOfSource` | `evidence_graphs` sets `EvidenceGraph: self()`, `SourceScope`, `Lists` and no `Search`. `sources` already sets `EvidenceGraph: onField("title", graphsOfSource(self()))`. `registry["source_types"]` is the shape to copy: `{Kind: DocSourceType, Path: self()}` plus `{Kind: DocSource, Path: inbound("sources", "source_type_id")}` |
| Write path | `core/writes/run.go` `Run` | Calls `searchindex.Reproject(tx.Tx, set.Search)` in the write's transaction when `set.Search` is non-empty |
| Graph writes | `core/database/evidencegraphs/evidencegraphs.go` `insert` (Create, `CreateDefault` from `sources.Create`), `Update`, `Delete`; `core/database/sources/sources.go` `Update` (`sqlUpdate` sets type, title, description; never `ref`), `Delete` → `deleteimpact.ReleaseFacets` | Each graph write returns an `evidence_graph` row change (create, update of `name` / `description`, delete); a Source rename returns a `source` change with `title` in its fields |
| Kind owner | `core/database/catalogmodel/model.go` `KindEvidenceGraph Kind = "evidence_graph"`; `core/search/search.go` `KindPerson = vocabkey.Person` | Search kinds may derive from a declared owner; the handle kinds do |
| Refs at open | `core/catalogsession/session.go` `ReconcileOnOpen` | `evidencegraphs.EnsureRefs` runs before `searchindex.EnsureCatalog`, so a rebuild indexes minted refs |
| Graph location shape (app) | `macos/App/Features/Workspace/WorkspaceLocation.swift` `static func evidenceGraph(sourceId:graphId:subjectId:ref:title:sourceTitle:)`; `macos/App/Features/Sources/SourcesListNavigation.swift` `graphLocation(for:graph:)` | Gated on `graph.hasArtifact`. The Sources list's graph button opens `.evidenceGraph(sourceId:, graphId:, ref: source.ref, title: source.title)`: on the canvas `title` is the Source title (`WorkspaceToolbar.breadcrumbItems` reads it as the Source crumb's fallback) |
| Wire | `api/ffi/handlers/search_catalog.go` `locationToProto`, `hitToProto`; `macos/App/Platform/GoStore.swift` `mapSearchHit` → `mapWorkspaceLocationFromProto` | `graph_id` and `source_surface` already cross both ways; `SourceSurface(rawValue: "graph")` is `.graph`. Comments on `SearchHit.kind` and `SearchCatalogRequest.kinds` in `api/proto/engine.proto` list six kinds |
| Hit presentation | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` `kindLabel(for:)`, `leadThumbnail(for:)`, `showsMatchContext`, `accessibilityLabel` | Unknown kinds fall back to the Source chip label and the `search` glyph; `description` already shows the "Description" match context |
| Hit row | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` | `source` draws `CachedThumbnail`; every other kind `leadThumbnail(for:)`. No change needed |
| Activate | `macos/App/Features/Workspace/OmnibarResultsModel.swift` `activate(_:navigation:)` | `navigation.go(to: hit.location)` for every kind |
| Canvas, no Artifact | `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift` `canCite`; `EvidenceGraphView.swift` `noArtifactCallout` | A graph with no linked Artifact opens with a warning callout and a button to the Source page |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Workspace.omnibarKindPlace` …; `macos/App/Resources/Localizable.xcstrings` `workspace.omnibar.kind.*` | Six kind chip strings |
| FakeStore | `macos/App/Platform/FakeStore.swift` `searchCatalog`, `vocabularySearchHits`, `searchPriority`, `classifyRefQuery`, `scoreSearchFields`; `evidenceGraphsByProject` | Sources (with ref fast path and section boost), types, fields and handles; no graphs. The Source loop's ref fast path is inline |
| Go tests | `core/search/search_test.go` (`TestExactRefIsTopHit`, `TestEnsureCatalogRebuildsAfterWipe`, `seedUser`, `seedType`, `uuidString`); `core/database/searchindex/searchindex_test.go` `TestMigrationCreatesSearchTables`; `core/search/handles_test.go` `newHandleFixture`; `core/database/evidencegraphs/evidencegraphstest` `SoleGraph`, `Create`; `api/ffi/handlers/search_catalog_test.go` `TestSearchCatalog`; `api/ffi/handlers/evidence_graphs_test.go` `newGraphFixture` | No graph search coverage |
| Swift tests | `macos/ProvenenciaTests/CatalogSearchTests.swift` `matchContextLocalizesBodyFieldsAndHidesTitleRef`, `leadHit`; `SearchCatalogStoreTests.swift` `findsSourceByTitleAndMapsLocation`; `OmnibarResultsModelTests.swift` `activateFieldHitFromSearch` | No graph hit coverage |

**Precedent.** Two commits, one per half. Go: `d58855d3` ("Index Persons, Events and Places for search from the resolved-values cache", S9-34a) added document kinds to `searchindex` with a `ProjectionVersion` bump and `RebuildAll` entry; copy its projector + rebuild + version shape, not its handle-cache reads. The migration copies `000020.sql` (two display columns, `TEXT NOT NULL DEFAULT ''`, healed by the projection version). The effects half copies `registry["source_types"]`'s two `Search` docs: its own row, and the parent's change fanning out to its children. Mac: `189822ce` ([#321](https://github.com/mendahu/provenencia/pull/321), "Show Person, Event, and Place hits in the omnibar.") added kind chip strings, the presentation cases, the proto comment and regeneration, and the omnibar tests; copy those, not its header oneof (graphs need no header).

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | Kind id `evidence_graph`: `search.KindEvidenceGraph`, `searchindex.KindEvidenceGraph` and `effects.DocEvidenceGraph` *(new)*, each `= string(catalogmodel.KindEvidenceGraph)`, beside the existing constants | `catalogmodel` declares the kind; deriving from it, as `KindPerson = vocabkey.Person` derives from `vocabkey`, adds no hand-synced copy (review M-3). `catalogmodel` is a leaf the three packages can import | Three `"evidence_graph"` literals — three copies to keep in step; moving the six existing kinds onto owners — not this PR's kind |
| D-2 | The document: `Kind` `evidence_graph`, `EntityID` graph id, `Title` / `DisplayTitle` the name, `Ref` / `DisplayRef` the ref (empty when not yet minted), `Secondary` the description, `DisplaySubtitle` the Source title, `DisplayParentID` the Source id (`database.UUIDString`), `DisplayParentRef` the Source ref; `Body`, icon and thumbnail empty | Name, description and ref are the match text (working rule); the Source's title, id and ref are display stubs in non-FTS columns, so they never match (AC-1.4) and the hit and its location come from the document alone, as every other kind's do (user, 2026-10-27) | Reading the Source at search time with a post-cut join — one kind hydrating differently from the rest (review M-4) |
| D-3 | `effects.registry["evidence_graphs"]` gains `Search: []SearchDoc{{Kind: DocEvidenceGraph, Path: self()}}`; `registry["sources"].Search` gains `{Kind: DocEvidenceGraph, Path: onField("title", graphsOfSource(self()))}` after its `DocSource` doc. `searchindex.Reproject` gains `case effects.DocEvidenceGraph: err = ReprojectEvidenceGraph(q, id)` | Registry lines keep every graph write (create, Copy Source's and Add Source's Default, edit, delete, Source delete) and every Source rename current in the write's own transaction, the way `source_types` fans out to its Sources. A Source's ref never changes after insert (`sources.sqlUpdate`), so only `title` fans out. Link rows (`evidence_graph_artifacts`) change no document field, so they get no `Search` | A reproject call inside `evidencegraphs` or `sources` — a second mechanism beside the effects registry; fanning out on every Source write — a description or type edit would reproject graphs for nothing |
| D-4 | `searchindex.ReprojectEvidenceGraph(q Querier, graphID []byte) error` *(new)* reads `SELECT COALESCE(g.ref, ''), g.name, COALESCE(g.description, ''), s.id, s.ref, s.title FROM evidence_graphs g JOIN sources s ON s.id = g.source_id WHERE g.id = ?`; `sql.ErrNoRows` → `Delete(q, KindEvidenceGraph, …)`. `RebuildAll` reprojects every `SELECT id FROM evidence_graphs` after Sources and before handles. `ProjectionVersion` = 9 with the comment line `9: Evidence graphs, with their Source in the parent columns (S12-15).` | The `ReprojectSourceType` shape; the version bump makes `EnsureCatalog` fill the new kind and columns on open (AC-1.10) | Reading through `evidencegraphs.List` — it loads every graph's links; the projector reads its own row like `ReprojectSource` reads `sources` |
| D-5 | Registry entry: `Kind: KindEvidenceGraph`, `DefaultInEverything: true`, `Priority: 0`, `ContextSections: []string{SectionSources}`, `ContextBoost: 1.1`, `Fields`: `title` 10, `ref` 12, `description` 3. `fieldValuesForKind` gains `case KindEvidenceGraph: {"title": title, "ref": refCol, "description": secondary}` | Same weights as a Source's own fields. Priority 0 is the skill's "other 0", so a Source still edges out a graph of equal text. Graphs live under Sources, so the Sources section nudges them | Priority 1 like Sources — the skill fixes the shares; the `default` field map — it drops the description (§4) |
| D-6 | `locationFor` takes the row: `locationFor(d docRow) WorkspaceLocation` (its one caller passes `d`; the existing cases read `d.kind`, `d.entityID`, `d.displayRef`, `d.displayTitle` as today). `case KindEvidenceGraph: WorkspaceLocation{Section: SectionSources, SourceID: d.displayParentID, GraphID: d.entityID, SourceSurface: SurfaceGraph, Ref: d.displayParentRef, Title: d.displaySubtitle}` with `SurfaceGraph = "graph"` *(new)* in `search.go` (comment: matches macOS `SourceSurface` raw values). The hit's `Subtitle` is `d.displaySubtitle`, as for every kind | The location matches `SourcesListNavigation.graphLocation` (`ref` / `title` are the Source's), so Back / Forward, the history menu and the breadcrumbs behave as for the list's button. Passing the row keeps one argument list as columns grow | A five-argument `locationFor` — the next column grows every call site |
| D-7 | Every graph is indexed, the Default graph of a single-graph Source included (AC-1.9) | The bar says the omnibar finds "a graph"; a `GPH-` ref must find every graph, and a Default graph opens the same canvas the Source's only graph button opens. The hits sit below Sources (priority 0) | Skipping Default graphs of single-graph Sources — the index would depend on a sibling count, so creating a second graph would reproject the first |
| D-8 | A graph hit always opens the canvas, linked Artifacts or not (AC-2.4) | The canvas already handles a graph with no Artifact (`canCite`, `noArtifactCallout`), and that callout links to the Source page where the Artifact is added | Gating on `hasArtifact` as the list's button does — Go would read links on every search, and the hit would need a disabled state the row does not have |
| D-9 | Swift: `CatalogSearchHitPresentation.kindLabel(for:)` gains `case "evidence_graph": L10n.Workspace.omnibarKindEvidenceGraph`; `leadThumbnail(for:)` gains `case "evidence_graph": PVThumbnail.Content(glyph: .symbol(.gitBranch))`. Nothing else in the row changes; the secondary line is `hit.subtitle` (the Source title) | `gitBranch` is the Evidence graph glyph (`PVSymbol.gitBranch`, "Evidence graph jump"; the Source page's Open button). The row's existing slots carry every field | A new mark or kit primitive — that is new chrome and would need a brief |
| D-10 | Copy: `L10n.Workspace.omnibarKindEvidenceGraph` *(new)*, key `workspace.omnibar.kind.evidenceGraph`, value "Evidence graph", comment "Kind chip on an omnibar hit for an Evidence graph" | "Evidence graph" in copy (working rule); the key follows `workspace.omnibar.kind.metadataField` | "Graph" — bare "graph" is not the product word |
| D-11 | FakeStore: `vocabularySearchHits` scores each `evidenceGraphsByProject[projectDir]` graph whose Source is in `sourcesByProject[projectDir]` with fields `title` (name) 10, `ref` 12, `description` 3, the Sources section boost 1.1 and the same ref fast path as Sources; kind `evidence_graph`, subtitle the Source title, location `SourcesListNavigation.canvasLocation(for: source, graph: graph)`. `SourcesListNavigation` gains `static func canvasLocation(for source: CatalogSource, graph: CatalogEvidenceGraph) -> WorkspaceLocation` *(new)*, the ungated `.evidenceGraph(sourceId:graphId:ref:title:)` call that `graphLocation(for:graph:)` now returns after its `hasArtifact` guard (review M-7). The Source loop's inline ref fast path moves into `private static func refBoost(for ref: String, key: (key: String, exact: Bool)?) -> (boost: Double, matched: Bool)` *(new)* (8 exact, 3 prefix, 1 and `matched: false` otherwise), called by both loops. The callers keep their own rules: with a ref key, `matched` sets `score = max(score, 12)` and `reason = "ref"`, and an unmatched hit with `score <= 0` is skipped; with no ref key a hit needs `score > 0`. The graph loop sits right after the Source loop, outside the `if refKey == nil` block that holds the types and fields loops. The comment "Ref-shaped queries only promote Sources" becomes "Ref-shaped queries only promote Sources and Evidence graphs". `searchPriority` needs no entry (absent is 0) | The skill: FakeStore mirrors kinds, location rules and the ref boost. One helper instead of a second copy of the ref branch | A second inline copy of the ref branch — two copies to keep in step with Go |
| D-12 | Tests: Go table-driven in a new `core/search/graphs_test.go` *(new)*, one new case in `TestSearchCatalog`; Swift in the three existing suites (§8). No test of the migration itself | `.cursor/rules/go-tests.mdc`; the handle kinds got their own `handles_test.go`, graphs get theirs; `add-catalog-migration` step 4: test the behavior a migration enables | Growing `search_test.go` — it is Source / vocabulary coverage |
| D-13 | Migration `core/database/migrations/000056.sql` *(new)*: `ALTER TABLE catalog_search_docs ADD COLUMN display_parent_id TEXT NOT NULL DEFAULT '';` and `ADD COLUMN display_parent_ref TEXT NOT NULL DEFAULT '';`, with a header comment in `000020`'s voice: the navigable record a hit opens inside (an Evidence graph's Source), not FTS-indexed, empty for kinds that open on their own id, healed by the projection version. No FK, no index, no `catalogmodel` change (`catalog_search_docs` is `BucketSkip`; `entity_id` is a UUID string the same way). `searchindex.Document` gains `DisplayParentID`, `DisplayParentRef` *(new)*; `Upsert` writes them; `docRow` gains `displayParentID`, `displayParentRef` *(new)*, and the doc read gets one column list and one scan first (review M-8): `const docColumns` *(new)* in `fts.go` lists `d.kind, d.entity_id, d.display_ref, d.display_title, d.display_subtitle, d.display_icon_key, d.display_thumbnail_rel_path, d.display_parent_id, d.display_parent_ref, d.title, d.ref, d.secondary, d.body`, and `func scanDoc(rows *sql.Rows, d *docRow, extra ...any) error` *(new)* scans those fields in that order, then `extra`. All four `SELECT` texts use `docColumns` (both `lookUpRefDocs` branches alias `catalog_search_docs d`), and all three scans call `scanDoc` (`mergeFTS` passes `&d.ftsRank`) | The user's call (2026-10-27): a hit's parent is a projected display stub, the way `000020` added icons and thumbnails. Generic names, so a later kind that opens inside a parent fills the same columns. The Source title already rides in `display_subtitle`, so two columns suffice | One `display_source_id` column — names one parent kind; a third `display_parent_title` column — duplicates `display_subtitle`; a JSON location column — the location has a typed shape in Go already |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `search.KindEvidenceGraph`, `searchindex.KindEvidenceGraph`, `effects.DocEvidenceGraph` *(new)* | const | `catalogmodel.KindEvidenceGraph`; `KindPerson = vocabkey.Person` | Derived from the declared owner (D-1) |
| `display_parent_id`, `display_parent_ref` (`000056.sql`) *(new)* | column | `display_icon_key`, `display_thumbnail_rel_path` (`000020.sql`) | Same kind of display stub; no existing column names a parent record (D-13) |
| `Document.DisplayParentID`, `DisplayParentRef`; `docRow.displayParentID`, `displayParentRef` *(new)* | field | `DisplayThumbnailRelPath` / `displayThumbnailRelPath` | One field per new column |
| `search.docColumns`, `search.scanDoc` *(new)* | const / func | the four inline column lists and three `rows.Scan` calls they replace | One list and one scan, so a column added to `docRow` cannot miss a retrieval path (D-13) |
| `search.SurfaceGraph` *(new)* | const | `deleteimpact` `surfaceGraph` (unexported) | `core/search` cannot read an unexported constant of another package, and `core/search` must not import `deleteimpact` for one string; it sits beside the `Section*` constants that mirror macOS raw values the same way |
| `searchindex.ReprojectEvidenceGraph` *(new)* | func | `searchindex.ReprojectSourceType` | Same shape, one more document kind; there is no generic projector to extend |
| Registry entry `KindEvidenceGraph`; `fieldValuesForKind` and `locationFor` cases | registry entry / case | The `KindSourceType` entry and cases | Registry entries, not new mechanisms |
| `effects.registry["evidence_graphs"].Search`, one doc on `registry["sources"].Search` | registry field | `registry["source_types"].Search` | Lines on the existing entries (D-3) |
| `SourcesListNavigation.canvasLocation(for:graph:)` *(new)* | func | `SourcesListNavigation.graphLocation(for:graph:)` | Split out of it so the list and FakeStore build one location (D-11) |
| `L10n.Workspace.omnibarKindEvidenceGraph` *(new)* | key | `omnibarKindPlace` | One more kind chip |
| `FakeStore.refBoost(for:key:)` *(new)* | helper | the inline ref branch in `vocabularySearchHits`'s Source loop | Extracted from it so both loops call one copy (D-11) |
| `core/search/graphs_test.go` *(new)* | test file | `core/search/handles_test.go` | Graph search gets its own file, as handles did (D-12) |

**Builds on.** The `add-searchable-kind` registry and projector, `000020`'s display stubs, the `effects` registry's `Search` docs (with `source_types`' parent fan-out) and `writes.Run`'s same-transaction reproject, `ProjectionVersion` healing, `WorkspaceLocation.evidenceGraph(…)`'s shape from S12-08, and `CatalogSearchHitPresentation`'s per-kind cases.
**Deletes.** The four inline column lists and three positional scans in `fts.go` and `refpath.go` (replaced by `docColumns` / `scanDoc`). S12-08's note that no hit sets `GraphID` stops being true; that note lives only in the S12-08 plan, which is not edited.
**Next adds.** None named in the map. A kind that opens inside a parent record fills `display_parent_id` / `display_parent_ref` in its projector and reads them in its `locationFor` case. The S12-99 bar pass checks bar item 7 against this PR.

### Schema

Migration `core/database/migrations/000056.sql` *(new)*:

```sql
-- The navigable record a hit opens inside: an Evidence graph's Source (S12-15).
-- Display stubs like 000020's, not FTS-indexed; empty for kinds that open on
-- their own id. ProjectionVersion 9 fills them on open.
ALTER TABLE catalog_search_docs ADD COLUMN display_parent_id TEXT NOT NULL DEFAULT '';
ALTER TABLE catalog_search_docs ADD COLUMN display_parent_ref TEXT NOT NULL DEFAULT '';
```

Registered in `catalogmodel`: nothing new (`catalog_search_docs` stays `BucketSkip`, no FK). Delete class: none. Audit: none (the index is not audited). Catalog format 55 → 56.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `DocEvidenceGraph` *(new)* (D-1); the two `Search` docs (D-3) |
| `core/database/searchindex` | `Document` parent fields and `Upsert` (D-13); `KindEvidenceGraph` *(new)*; `ProjectionVersion = 9`; `Reproject` case; `ReprojectEvidenceGraph` *(new)*; `RebuildAll` loop over `evidence_graphs` (D-4). Package comment unchanged |
| `core/search` | `KindEvidenceGraph`, `SurfaceGraph` *(new)* in `search.go`; registry entry (D-5); `fieldValuesForKind` case; `docRow` fields and the three `SELECT`s (D-13); `locationFor(d docRow)` and its graph case (D-6) |

Errors: none new.

### FFI

None — no method, message or field. `api/proto/engine.proto` comment edits only: `SearchHit.kind` reads `// source | source_type | metadata_field | person | event | place | evidence_graph`; `SearchCatalogRequest.kinds` reads `// (sources, types, metadata fields, persons, events, places, and Evidence graphs).` Regenerate with `./scripts/generate-proto.sh` so `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift` carry the comment.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Graph hit row | `PVOmnibarHitRow` via `CatalogSearchHitRow` (unchanged) | `macos/App/Features/Catalog/Search/CatalogSearchHitRow.swift` | Title = name, kind chip "Evidence graph", secondary = Source title, ref = `GPH-…` |
| Kind chip and lead | `CatalogSearchHitPresentation.kindLabel(for:)`, `leadThumbnail(for:)` | `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift` | D-9 |
| Open | `OmnibarResultsModel.activate(_:navigation:)` (unchanged) | `macos/App/Features/Workspace/OmnibarResultsModel.swift` | The hit's location is `.evidenceGraph(…)`'s shape |
| Fake graph hits | `FakeStore.vocabularySearchHits`, `refBoost(for:key:)` *(new)* | `macos/App/Platform/FakeStore.swift` | D-11 |
| Canvas location | `SourcesListNavigation.canvasLocation(for:graph:)` *(new)*, `graphLocation(for:graph:)` | `macos/App/Features/Sources/SourcesListNavigation.swift` | D-11 |

States: no new state; loading, empty and error are the omnibar's. Accessibility: `accessibilityLabel(for:)` composes "Jones household, Evidence graph, 1881 Census, Ward 3, GPH-J0NES" from the existing slots.

### Strings

L10n keys *(new)*: `L10n.Workspace.omnibarKindEvidenceGraph` = "Evidence graph" (`workspace.omnibar.kind.evidenceGraph`), in `L10n.swift` after `omnibarKindPlace` and in `Localizable.xcstrings` (`extractionState` `manual`, `en` `translated`). Removed: none.

## 7. Steps

**STEP-1 — Parent columns on the search index**
Files: `core/database/migrations/000056.sql` *(new)*, `core/database/searchindex/searchindex.go`, `core/search/fts.go`, `core/search/refpath.go`. Skill: `add-catalog-migration`.
Do: D-13 in this order: first `docColumns` and `scanDoc` with today's eleven columns, all four `SELECT` texts and three scans moved onto them, and `go test ./core/search/...` green; then the migration, `Document` and `Upsert`, and the two new names in `docRow`, `docColumns` and `scanDoc`. No kind fills the columns yet. Commit the helper move on its own ("Read search documents through one column list and one scan.").
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/... ./core/search/...` passes (schema hash and `TestPragmaHonesty` accept the step; every existing search test stays green).

**STEP-2 — Project a document per graph**
Files: `core/database/effects/registry.go`, `core/database/searchindex/searchindex.go`. Skill: `add-searchable-kind` (§3 Projection).
Do: D-1's two constants in these packages, D-3, D-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/... ./core/database/searchindex/... ./core/writes/...` passes (the registry init check accepts both new `Search` paths).

**STEP-3 — Register the kind and its location**
Files: `core/search/search.go`, `core/search/registry.go`, `core/search/fts.go`, `core/search/graphs_test.go` *(new)*, `core/search/handles_test.go`. Skill: `add-searchable-kind` (§1, §2), `.cursor/rules/go-tests.mdc`.
Do: D-1's `core/search` constant, `SurfaceGraph`, D-5, D-6. Write T-1 to T-3 and extend T-4.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/search/...` passes, T-1 to T-4 included.

**STEP-4 — Handler test and proto comments**
Files: `api/proto/engine.proto`, generated `api/proto/engine/engine.pb.go` and `macos/App/Platform/Generated/engine.pb.swift`, `api/ffi/handlers/search_catalog_test.go`.
Do: §6 FFI comment edits; `./scripts/generate-proto.sh`; T-5.
Check: `git diff --stat` shows only comment lines in the generated files; `CGO_ENABLED=1 go test -tags fts5 ./api/ffi/...` passes.

**STEP-5 — Kind chip and lead glyph**
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Catalog/Search/CatalogSearchHitPresentation.swift`, `macos/ProvenenciaTests/CatalogSearchTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-10, D-9; T-6.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CatalogSearchTests` passes.

**STEP-6 — FakeStore graph hits and opening a hit**
Files: `macos/App/Platform/FakeStore.swift`, `macos/App/Features/Sources/SourcesListNavigation.swift`, `macos/ProvenenciaTests/SearchCatalogStoreTests.swift`, `macos/ProvenenciaTests/OmnibarResultsModelTests.swift`. Skill: `add-searchable-kind` (§4), `add-swift-test`.
Do: D-11; T-7, T-8.
Check: `xcodebuild test … -only-testing:ProvenenciaTests/SearchCatalogStoreTests -only-testing:ProvenenciaTests/OmnibarResultsModelTests` passes, and the existing `findsSourceByRef` and `prefixRefPromotesSource` stay green after the `refBoost` extraction, and every `SourcesList…` suite stays green after the `canvasLocation` split.

**STEP-7 — Docs**
Files: `docs/deployment-plan/archive/spike-3/omnibar-search.md`, `.agents/skills/add-searchable-kind/SKILL.md`, `docs/deployment-plan/spike-12/completed.md`, `docs/deployment-plan/spike-12/deployment-plan.md` (map row), this plan (status).
Do: `deployment-plan.md` Working rules › Client › "One graph list": "breadcrumbs and the omnibar location all read it" becomes "and breadcrumbs all read it; an omnibar graph hit carries its Source on its search document (S12-15)". `omnibar-search.md` — "What should be searchable" gains the row `| Evidence graphs | Source | name, description, ref | GPH-… |` after Sources; the Navigable root row lists "Evidence graph"; the "Spike 3 first kinds" paragraph gains one sentence: "Evidence graphs shipped in S12-15: one document per graph, and the hit opens the graph's canvas from its Source's id and ref, projected on the document's parent columns." `add-searchable-kind` — the §2 location table gains `| evidence_graph | sources | graph_id, with source_surface graph and source_id from display_parent_id |`, followed by one sentence: "A kind that opens inside a parent record projects the parent's id and ref into `display_parent_id` / `display_parent_ref` (and its title into `display_subtitle`); `locationFor` reads them.", and both priority lines (the Checklist line "Priority share of the kind tenth" and the §1 table's `Priority` row) end "other 0 (types, fields, Evidence graphs)". Then the Definition of done's doc lines: the map row's Status becomes `Landed #N` (the PR number once opened), its Plan link stays.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-12/prs/S12-15-graph-search.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Find Evidence graphs in the omnibar by name, description or ref."). No WIP commits.

## 8. Tests

Go fixture for T-1 to T-3 (`graphs_test.go`, a helper `newGraphSearchFixture(t)` *(new)* in that file): `newHandleFixture` (Source "Register", Artifact "Scan", one Default graph); a second Artifact "Page 2" from `artifactstest.Create(c, user, artifacts.CreateInput{SourceID: source, Label: "Page 2"})`; then "Jones household" with description "Thomas Jones and family, page 2" linked to "Page 2", created through `writes.Call(c, writes.Op{Action: "create_evidence_graph", UserID: user}, …)` around `evidencegraphs.Create(tx, user, evidencegraphs.CreateInput{SourceID: source, Name: "Jones household", Description: "Thomas Jones and family, page 2", ArtifactIDs: [][]byte{page2}})` (`evidencegraphstest.Create` takes no description); and "Smith household" from `evidencegraphstest.Create(t, c, user, source, "Smith household", scan)`. Refs are read from the returned `evidencegraphs.Graph`. Updates and deletes in T-2 go through `writes.Run` with actions `update_evidence_graph`, `delete_evidence_graph`, `update_source` and `delete_source`, as the FFI handlers name them. Querier calls in T-3 use `db, _ := c.DB()`.

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/search/graphs_test.go` `TestEvidenceGraphSearch` *(new)*, table-driven | `jones hou` → a hit with Kind `evidence_graph`, Title "Jones household", Subtitle "Register", Ref the graph ref, MatchReason `title`, Location `{Section: sources, SourceID: Register, GraphID: Jones, SourceSurface: graph, Ref: Register's ref, Title: "Register"}`; `thomas` → the same hit, MatchReason `description`; the exact ref → first hit, MatchReason `ref`; the ref's first six characters lower-cased → the hit, MatchReason `ref`, Location.SourceID Register (the `LIKE` branch of `lookUpRefDocs`); `register` → no `evidence_graph` hit; `default` → one `evidence_graph` hit titled "Default"; `jones` with `Kinds: []string{KindPerson}` → no `evidence_graph` hit | AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.9, AC-2.1, D-5, D-6, D-13 |
| T-2 | Go | `core/search/graphs_test.go` `TestEvidenceGraphSearchFollowsWrites` *(new)*, table-driven, each case a write then a search | Update Jones's name to "Jones family" → `jones fam` hits, `jones hou` has no graph hit; update its description to "Mary Jones" → `thomas` has no graph hit, `mary` hits; update the Source title to "Register (Toronto)" → the Jones hit's Subtitle and Location.Title read "Register (Toronto)"; delete Smith (`evidencegraphs.Delete` through `writes.Run`) → `smith hou` has no graph hit; a second Source with no Artifact, deleted with `sources.Delete` → its Default graph's ref returns no hit | AC-1.5, AC-1.6, AC-1.7, AC-1.8, D-3 |
| T-3 | Go | `core/search/graphs_test.go` `TestRebuildIndexesEvidenceGraphs` *(new)* | `searchindex.ClearAll` and `SetProjectionVersion(db, 8)` on the fixture, then `searchindex.EnsureCatalog` → `jones hou` hits with Location.SourceID Register and the stored version is 9 | AC-1.10, D-4 |
| T-4 | Go | `core/search/handles_test.go` `TestKindPriorityIsATenthOfTheScore` (extended) | The equal-score list gains `{Kind: KindEvidenceGraph, Title: "G", Score: 10}` → order person, event, place, source, evidence_graph, source_type (priority 0 ties break on kind name) | D-5 |
| T-5 | Go | `api/ffi/handlers/search_catalog_test.go` `TestSearchCatalog` case "finds an evidence graph by ref" *(new)* | `newGraphFixture`, `CreateEvidenceGraph` "Jones household" on `page2`, then `SearchCatalogRequest{Query: <its ref>}` → the first hit has kind `evidence_graph`, location `graph_id` the graph, `source_id` the fixture Source, `source_surface` `graph` | AC-1.3, D-6 |
| T-6 | Swift | `macos/ProvenenciaTests/CatalogSearchTests.swift` `evidenceGraphHitPresentation` *(new)* | `kindLabel(for: "evidence_graph")` == `omnibarKindEvidenceGraph`; `leadThumbnail(for: leadHit(kind: "evidence_graph")).glyph == .symbol(.gitBranch)`; `accessibilityLabel` of a hit (title "Jones household", subtitle "1881 Census, Ward 3", ref `GPH-J0NES`, reason `description`) equals `ListFormatter.localizedString(byJoining:)` of [title, "Evidence graph" (`L10n.string(omnibarKindEvidenceGraph)`), subtitle, ref, "Description"], so the test does not depend on the locale's separators | AC-1.1, AC-1.2, D-9, D-10 |
| T-7 | Swift | `macos/ProvenenciaTests/SearchCatalogStoreTests.swift` `findsEvidenceGraphByNameRefAndDescription` *(new)* | FakeStore with the §2 fixture, seeded by assigning `sourcesByProject[projectDir]` and `evidenceGraphsByProject[projectDir]` (`CatalogEvidenceGraph`) → `jones hou`: one `evidence_graph` hit, subtitle "1881 Census, Ward 3", location `SourcesListNavigation.canvasLocation(for:graph:)` of the fixture (Source `SRC-W3CEN`, title "1881 Census, Ward 3"); `thomas`: reason `description`; `GPH-J0NES`: first hit, reason `ref`; a graph whose Source is not in `sourcesByProject` → no hit; `heldCatalogProjectDir == projectDir` | AC-1.1, AC-1.2, AC-1.3, D-11 |
| T-8 | Swift | `macos/ProvenenciaTests/OmnibarResultsModelTests.swift` `activateEvidenceGraphHitOpensTheGraph` *(new)* | The T-7 store, query `jones`, activate the graph hit → `navigation.currentLocation.graphId` is Jones, `.sourceId` the Source, `.sourceSurface == .graph`, `selectedSection == .sources`, the query is empty | AC-2.1, D-8 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database` schema-hash and `catalogmodel.TestPragmaHonesty` (the migration), `core/search` ranking tests that count hits (`TestSharedTokenFindsSourceAndVocab`, `TestMultiTokenPartialCoverageStillRetrieves`: every new Source now has a Default graph document, which matches only `default`), `core/database/effects` `TestEvidenceGraphsResolve` and `core/writes` `run_test.go` (now non-empty `Set.Search` on graph writes and Source renames), and `SearchCatalogStoreTests` after the `refBoost` extraction.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches §6 macOS; no new kit primitive
- [ ] Version bump: none (VERSION stays 0.2.0)
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row Plan link and status; completed.md: one short paragraph of what shipped
- [ ] Brief: none (not gated)
- [ ] Domain doc updated: docs/deployment-plan/archive/spike-3/omnibar-search.md kind tables; .agents/skills/add-searchable-kind/SKILL.md location table, parent columns and priority lines; deployment-plan.md "One graph list" rule
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with a census Source, "1881 Census, Ward 3", with two page Artifacts. On its Source page add the graphs "Jones household" (description "Thomas Jones and family, page 2", Page 2 checked) and "Smith household".
2. (AC-1.1, AC-2.1) In the omnibar type `jones hou`. See "Jones household · Evidence graph · 1881 Census, Ward 3 · GPH-…" with the branch glyph. Press Return. The Jones household canvas opens; the breadcrumbs read Sources › 1881 Census, Ward 3 › Jones household.
3. (AC-2.2, AC-2.3) Press Back, then Forward: the canvas returns. Quit and relaunch: the canvas is restored.
4. (AC-1.2, AC-1.3) Copy the graph's `GPH-` ref from its Source page card, paste it in the omnibar: the graph is the first hit with the ref accented. Type `thomas`: the hit shows "Description".
5. (AC-1.5, AC-1.6, AC-1.7) Rename the graph to "Jones family" on the Source page; `jones fam` finds it. Rename the Source; the hit's second line follows. Delete "Smith household"; `smith hou` finds no graph.
6. (AC-2.4) On a Source whose only Artifact you delete, search its Default graph's ref and press Return: the canvas opens with the "no artifact" callout.
7. (AC-1.10) On a project last opened by a build before this PR, open it and type a graph's name: the hit is there.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A Default graph document per Source adds a hit to tests that count hits for words like `default` | T-1 pins the `default` case; §8 names the suites to rerun |
| A graph write path that returns no `evidence_graph` row change, or a Source rename that does not fan out, leaves a stale document | T-2 covers create, edit, delete, Source rename and Source delete |
| The migration widens a table every project carries | Two `TEXT NOT NULL DEFAULT ''` columns, the `000020` shape; the projection version fills them on first open |
| The rebuild on open is slower with one more document per Source | One row read per graph, inside the existing rebuild transaction; graphs are a few per Source |

**Stop and report if:**

- `./scripts/generate-proto.sh` cannot run (no `protoc` or plugins) or regenerates more than comment lines: do not hand-edit the generated files.
- No macOS toolchain is available for `xcodebuild test`: STEPs 4 and 5 and `check-pr.sh` need one ([`ship-pr`](../../../../.agents/skills/ship-pr/SKILL.md) runs on the Mac).

- `effects` package init panics on `self()` for `evidence_graphs`' `Search` path, or `Resolve` returns no `DocEvidenceGraph` id for a graph delete (T-2's delete case fails with the document still present).
- `sources.Delete` of a Source with no Artifact returns no `evidence_graph` delete change, so T-2's last case finds the Default graph's ref.
- `sources.Update` (or any Source write) can change `ref`, so the parent ref on graph documents could go stale with no `title` fan-out.
- The schema hash or `TestPragmaHonesty` refuses `000056.sql` as written.
- `catalogsession.ReconcileOnOpen` no longer runs `evidencegraphs.EnsureRefs` before `searchindex.EnsureCatalog`.
- `WorkspaceLocation.evidenceGraph(…)` or `SourcesListNavigation.graphLocation(for:graph:)` no longer puts the Source's ref and title on the location (D-6 and D-11 match them).
- `mapWorkspaceLocationFromProto` maps `source_surface` `graph` to anything but `.graph`, or drops `graph_id`.
- A test outside §8's list fails because of the new document and fixing it means changing the test's intent.

## 12. Open questions

None.

## 13. PR

**Title:** S12-15 — Graphs in the omnibar
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-12/prs/S12-15-graph-search.md`.
