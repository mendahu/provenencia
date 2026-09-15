# Spike 3 — Completed steps

Finished Spike 3 work kept for history. Spike archived — optional later slices: [`deployment-plan.md`](deployment-plan.md) (S3-13+).

IDs stay stable (`S3-NN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| [S3-01](#s3-01--design-app-layout-toolbar) | Design | App Layout toolbar (Back/Forward, breadcrumbs, omnibar field) |
| [S3-02](#s3-02--design-omnibar-results) | Design | Omnibar results dropdown + hit row |
| [S3-03](#s3-03--pr-catalog-projectuuid) | PR | Catalog `project.uuid` (mint, heal, ProjectInfo) |
| [S3-04](#s3-04--pr-workspacelocation--navigation-history) | PR | WorkspaceLocation + persisted navigation history |
| [S3-05](#s3-05--pr-main-column-toolbar-chrome) | PR | Main-column toolbar (nav + crumbs + omnibar shell) |
| [S3-06](#s3-06--skipped--history-jump-menu-polish) | PR | Skipped — full jump menus shipped in S3-05 |
| [S3-07](#s3-07--pr-searchable-kind-registry--searchcatalog-rpc-shell) | PR | Searchable-kind registry + SearchCatalog RPC |
| [S3-08](#s3-08--pr-fts5-projection) | PR | FTS5 catalog search projection |
| [S3-09](#s3-09--pr-context-ranking--ref-fast-path) | PR | Context ranking + ref fast path |
| [S3-10](#s3-10--pr-omnibar-results-ui--wire-search--remove-list-search) | PR | Omnibar results + retire list search |
| [S3-11](#s3-11--pr-fuzzy--typo-matching) | PR | Trigram shortlist + Jaro–Winkler typo tolerance |
| [S3-12](#s3-12--pr-spike-3-dogfood-polish) | PR | Dogfood polish for history + omnibar |

---

## Steps

### S3-01 — Design: App Layout (toolbar)

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | — |
| **Deliverables** | Done. Board for the main-column toolbar: Back/Forward (long-press / secondary-click jump menus), flex-grow breadcrumbs, trailing omnibar field (`⌘K`). Layout summary: [`design/README.md`](design/README.md). Boards stay in Claude Design (not checked into git). |
| **Context** | [`navigation-history.md`](navigation-history.md); [`omnibar-search.md`](omnibar-search.md) field placement. |
| **Out** | Results dropdown (S3-02); engine search; history persistence. |
| **Feeds** | S3-05 (S3-06 skipped — full jump menus in S3-05) |

---

### S3-02 — Design: Omnibar results

| | |
| --- | --- |
| **Kind** | Design (Claude Design) |
| **Depends on** | S3-01 (omnibar field placement) |
| **Deliverables** | Done. Board for flat ranked rich hit rows (`PVOmnibarHitRow` slot contract), dropdown chrome, empty/loading/keyboard states. Brief archived: [`design/archive/omnibar-results-dropdown-brief.md`](design/archive/omnibar-results-dropdown-brief.md). Summary: [`design/README.md`](design/README.md). |
| **Context** | [`omnibar-search.md`](omnibar-search.md). |
| **Out** | Search engine / FTS; faceted query UI. |
| **Feeds** | S3-10 |

---

### S3-03 — PR: Catalog `project.uuid`

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — |
| **Deliverables** | Done. Migration `000017` adds `project.uuid` (`BLOB`, unique when set). `project.NewID` / Upsert mints UUIDv7 on Create; `project.EnsureUUID` heal-mints on Open / `createCatalog` / `catalogsession` (beside `users.EnsureRefs`). Existing uuid never rewritten on Upsert conflict. Exposed on protobuf / Swift `ProjectInfo.uuid` via Complete, Open, and GetProjectInfo. |
| **Context** | [`navigation-history.md`](navigation-history.md) § Project key; skill [`add-catalog-migration`](../../../../.cursor/skills/add-catalog-migration/SKILL.md). Active-project pointer stays path-based. |
| **Out** | History JSON; toolbar; short `PRJ-…` ref. |
| **Dogfood** | New + upgraded projects show a stable UUID on open; rename/move of the folder does not change it. |
| **Feeds** | S3-04 (history file key); later install-local chrome |

---

### S3-04 — PR: `WorkspaceLocation` + navigation history

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-03 (done) |
| **Deliverables** | Done. `WorkspaceLocation` (section + `sourceId` / `fieldId` / `typeId`, denormalized ref/title). Freestanding `WorkspaceNavigation` (environment-injected; not fused into sidebar chrome): owns `NavigationHistoryStore` with `go(to:)`, `goBack()`, `goForward()`, `go(toIndex:)`, coalesce, truncate forward, cap 100. Persist `Application Support/Provenencia/navigation/{uuid}.json` (v1). Sidebar, Source open/close/create, and vocabulary row selection all commit via `go(to:)`. Destinations apply `currentLocation` on appear/change; missing entities prune to section root. `⌘[` / `⌘]` via `NavigationCoordinator`. `WorkspaceModel` is sidebar collapse only. |
| **Context** | [`navigation-history.md`](navigation-history.md). Hand-rolled coordinator — not SwiftUI `NavigationStack` / `NavigationPath`. |
| **Out** | Final toolbar visuals / jump menus (S3-05); omnibar results; search RPC. |
| **Dogfood** | Keyboard Back/Forward restores deep locations across destinations; relaunch returns to the leave-off place for that `project.uuid`. |
| **Feeds** | S3-05 (toolbar chrome); S3-07+ Hit `location` shape |

---

### S3-05 — PR: Main-column toolbar chrome

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-04 (done); S3-01 board (done) |
| **Deliverables** | Done. App Layout main-column toolbar (`WorkspaceToolbar`): Back/Forward with chevron + mono shortcut hints, long-press (~400ms) / secondary-click jump menus (nearest 15, icon \| Destination › title \| mono ref), flex `PVBreadcrumbs` via `go(to:)` for ancestors, trailing omnibar **field shell** (`PVInput` + ⌘K suffix / `⌘K` focus via `OmnibarFocusCoordinator`). Header height stays **52** for traffic lights. Stripped Source-page identity-header breadcrumbs / `onBackToList` crumb wiring. L10n for omnibar + Focus Search. |
| **Context** | Claude Design App Layout board; [`design/README.md`](design/README.md); [`navigation-history.md`](navigation-history.md) § Toolbar. |
| **Out** | Omnibar results dropdown / `searchCatalog` wire / remove list search (S3-10). |
| **Dogfood** | Toolbar Back/Forward + jump menu + crumb click feel browser-like; ⌘K focuses the field; Source page has no local crumb trail. |
| **Feeds** | S3-10 (results + search wire) |

---

### S3-06 — Skipped: History jump-menu polish

| | |
| --- | --- |
| **Kind** | PR (skipped) |
| **Depends on** | S3-05 |
| **Deliverables** | Skipped. Full jump-menu anatomy (rich rows, long-press swallow, popover dismiss) shipped with **S3-05**; this optional split is unused. |
| **Context** | Was only needed if S3-05 shipped a minimal Back/Forward without menus. |
| **Out** | — |
| **Feeds** | — |

---

### S3-07 — PR: Searchable-kind registry + `SearchCatalog` RPC shell

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | — (parallel with S3-05 after S3-04 location shape) |
| **Deliverables** | Done. `core/search` kind registry (`source` / `source_type` / `source_field`) with field weights, context-section boosts, and `WorkspaceLocation` mappers. Protobuf `METHOD_SEARCH_CATALOG` + `SearchHit` / `WorkspaceLocation` messages; FFI handler via `withProjectCatalog` + `search.Engine`. `Searcher` interface with `NaiveScanner` bridge (domain `List` APIs — no handler SQL). Mac `GenealogyStore.searchCatalog` / GoStore / FakeStore + Swift tests. Skill [`add-searchable-kind`](../../../../.cursor/skills/add-searchable-kind/SKILL.md). |
| **Context** | [`omnibar-search.md`](omnibar-search.md) § Incremental delivery S1; [`add-ffi-handler`](../../../../.cursor/skills/add-ffi-handler/SKILL.md); [`use-catalog-session`](../../../../.cursor/skills/use-catalog-session/SKILL.md). |
| **Out** | FTS5 tables / write-path reproject (S3-08); full context/ref ranking polish (S3-09); omnibar UI / remove list search (S3-10). |
| **Dogfood** | RPC / FakeStore return ranked hits for known titles, refs, type/field labels with navigable locations. |
| **Feeds** | S3-08 (swap NaiveScanner for FTS behind same Hit/Query/FFI) |

---

### S3-08 — PR: FTS5 projection

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-07 |
| **Deliverables** | Done. Enabled `-tags fts5` on dylib build + CI/`go test`. Migration `000018`: `catalog_search_docs` + external-content `catalog_search_fts` + `catalog_search_meta`. `core/database/searchindex` projectors for Source (notes/metadata/artifact/filename rollup) and vocab roots; incremental reproject on domain writes; `searchindex.EnsureCatalog` rebuild/heal on Open/Create. `FTSSearcher` is `DefaultEngine` (NaiveScanner removed). Registry body weights for rolled Source text. Skill [`add-searchable-kind`](../../../../.cursor/skills/add-searchable-kind/SKILL.md) updated. |
| **Context** | [`omnibar-search.md`](omnibar-search.md) S2; [`application-stack.md`](../../../application-stack.md) §10. |
| **Out** | Fuzzy (S3-11); Files/Artifact as own hit kinds; omnibar UI (S3-10). |
| **Dogfood** | Edit a Source title/note → SearchCatalog updates; title beats body; Open heals a wiped index. |
| **Feeds** | S3-09 |

---

### S3-09 — PR: Context ranking + ref fast path

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-08 |
| **Deliverables** | Done. Registry `RefBoostWeights` + section `ContextBoost` (2.0). Exact/prefix ref fast path on `catalog_search_docs.ref` (skips FTS for ref-shaped queries). Controlled multi-token OR retrieve for partial term coverage. Tagged Source body rollup (`note:` / `metadata:` / `filename:`) with `ProjectionVersion` 2 + snippet `match_reason`. Hit cap remains 50. FakeStore parity for paste-ref. |
| **Context** | [`omnibar-search.md`](omnibar-search.md) S3. |
| **Out** | Fuzzy (S3-11). |
| **Dogfood** | Paste `SRC-…` → top hit with `match_reason` ref; Sources context floats Sources without hiding vocabulary. |
| **Feeds** | S3-10 |

---

### S3-10 — PR: Omnibar results UI + wire search + remove list search

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-05; S3-09; S3-02 board |
| **Deliverables** | Done. `PVOmnibarHitRow` + content-column results overlay (jump-menu family / `PVRadius.sm`, not NSPanel). `OmnibarResultsModel` debounces SearchCatalog (180ms, min 2 chars); keyboard ↑↓/Return/Esc + outside dismiss; select → `go(to:)` and clear query. Source leads from `listSources`/`CachedThumbnail`; type/field glyphs. Removed Sources + `VocabularyListPane` list-search chrome and query filters. L10n + `OmnibarResultsModelTests`. |
| **Context** | Claude Design Omnibar Results board; [`omnibar-search.md`](omnibar-search.md) S4. |
| **Out** | Files/Artifact first-class hit kinds. |
| **Dogfood** | ⌘K → find Source/type/field → navigate → Back; no list search fields. |
| **Feeds** | S3-11; S3-12 |

---

### S3-11 — PR: Fuzzy / typo matching

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-10 (engine could land after S3-09; omnibar already wired) |
| **Deliverables** | Done. Migration `000019`: external-content `catalog_search_fts_trigram` (`tokenize='trigram case_sensitive 0 remove_diacritics 1'`). `searchindex` keeps unicode61 + trigram indexes in sync; `ProjectionVersion` 3. `FTSSearcher` expands under-limit candidates via OR-of-character-trigrams shortlist (cap 150), Jaro–Winkler gate on title/label/ref/key (`FuzzyWeights`), fuzzy-only score scale so exact FTS still wins. Ref-shaped queries skip fuzzy. Latin-script dogfood focus (not CJK). Tests for typo recall, ranking, garbage, accent fold; FFI smoke. Skill [`add-searchable-kind`](../../../../.cursor/skills/add-searchable-kind/SKILL.md) updated. |
| **Context** | [`omnibar-search.md`](omnibar-search.md) S5. |
| **Out** | Cross-root association search; NL/AI; spellfix1 / stemming; CJK fuzzy models. |
| **Dogfood** | Common typos (`Ilminstr` → Ilminster) recover without flooding garbage. |
| **Feeds** | S3-12 |

---

### S3-12 — PR: Spike 3 dogfood polish

| | |
| --- | --- |
| **Kind** | PR |
| **Depends on** | S3-10; S3-11 preferred |
| **Deliverables** | Done. Gate `fallbackToSectionRoot` on `hasCompletedInitialLoad` (Sources + vocab) so relaunch deep places are not pruned before catalog load. Omnibar loading row + `searchError` callout (not blank panel / not “no matches”). A11y: Back/Forward identifiers, per-hit omnibar ids + labels, jump-menu accessibility label. Confirmed jump menu and omnibar already share `PVRadius.sm` + `PVElevation.overlay`. Regression tests for load gate, close/loading/empty/error/selection. Spike 3 dogfood closeout; open questions parked as deferred. |
| **Context** | [`deployment-plan.md`](deployment-plan.md) Definition of done; [`navigation-history.md`](navigation-history.md) restore; omnibar S4/S5. |
| **Out** | New catalog layers; project Files browser; empty-query recents; non-current stack prune on attach. |
| **Dogfood** | Relaunch on a Source page stays put; ⌘K search shows loading then hits or a clear error; identifiers support UI scripts. |
| **Feeds** | Optional S3-13+ |
