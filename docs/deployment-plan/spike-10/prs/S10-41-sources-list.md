# S10-41 — Sources list: last updated, status filter, search, saved settings

**Status:** Blocked — on S10-22 and board S10-D8
**Spike / slice:** Spike 10 · Slice 6 — Lists and the Person page
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-41--sources-list)
**Brief:** [S10-D8](../design/S10-D8-sources-list.md) (SRL-1…SRL-10), board not agreed yet
**Board:** Claude Design project *Sources list* (project id and frame numbers recorded here when the board is agreed). §6 macOS describes every frame this PR ships in words. The implementer does not need the board.
**Depends on:** S10-22 (in progress; adds `LastUpdatedDisplay`, `L10n.LastUpdated`, `PVSymbol.history`, and the `PVList` detail column), S10-D8
**Blocks:** —
**Branch:** `s10-41-sources-list`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rules "The Sources list is one page", "Last updated is an audit scope", "Sources evidence filters count primaries in one read", "List scroll is a field on the history entry"; [`navigation-history.md`](../../archive/spike-3/navigation-history.md); [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) (Identity Claims, bridges); [S10-22 plan](S10-22-list-last-updated.md) D-5…D-11
**Planned:** 2026-10-15 against `main` @ `832ccc2`
**Review:** [`S10-41-sources-list.review.md`](S10-41-sources-list.review.md)

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

The researcher can find a Source and see where it stands. Each Sources row shows when they last worked on it. They can narrow the list to Sources whose graph is empty, in progress, or filed, and type part of a title or ref to find one. The list comes back with the same sort, type, status, and search when they return to it, including after relaunch.

**Why now.** Slice 6 is the lists slice. S10-09 (persist sort and type filter) was folded into this PR so the Sources toolbar is designed and persisted once (map, 2026-10-12). S10-21 left the Sources revision read for this PR to move onto `audit.LatestByScope`. S10-22 builds the last-updated cell this row reuses.

**What stays.** The split row (filing zone | Evidence graph zone), its counts, the caption band, the type filter and sort options, row hover and keyboard, the Add Source dialog, the empty catalog state, and the load error. Scroll restore is S10-40 and is unchanged. People, Events, and Places get none of these controls.

## 2. User stories

**US-1.** As a researcher, when I scan the Sources list, I see when each Source last changed, so I can pick up recent work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source whose newest scoped revision was today at 15:15 local | The row's column reads “Today, 3:15 PM” (`en_US`) behind the history icon; help reads “Last updated 14 Oct 2026 · 3:15 PM MDT”. |
| AC-1.2 | A Source with no scoped revision (`audit.Latest` zero) | The column reads a faint “—” with no help; the filing zone's spoken value ends “last updated unknown”. |
| AC-1.3 | Edit a Source's note, go to another place, come back | That row's time moved; the list order follows the sort, not the time (unless the sort is Date updated). |
| AC-1.4 | `ListSources` on a catalog of 3 Sources | Each `Source.updated_at` is RFC3339 and `updated_revision` > 0; one `audit_transaction_scopes` read serves the whole list. |

**US-2.** As a researcher, I filter the list by evidence status so I can find graphs that still need promoting.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source with no person, event, or place subjects (a participation-only graph counts too) | Matches **Empty**, not In progress or Filed. |
| AC-2.2 | A Source with one person subject and no accepted claim on it | Matches **In progress**. |
| AC-2.3 | Promote that person (accepted Identity Claim) on the Promote page, return to Sources | The Source matches **Filed**, without relaunch. |
| AC-2.4 | A person with only a `provisional` claim | Counts as not filed: **In progress**. |
| AC-2.5 | 200 Sources | `ListSourceGraphProgress` issues the same fixed number of queries as for 2 Sources (no per-row query). |
| AC-2.6 | Status **Empty** set and the progress map not loaded yet | The list area shows the loading `ProgressView`, not the filter-empty state. |

**US-3.** As a researcher, I type part of a title or ref to find a Source.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | Titles “Smith family Bible”, “1851 census”, “José Ruiz letter”; type `smith` | Only “Smith family Bible” is shown. |
| AC-3.2 | Type `jose` | “José Ruiz letter” is shown (accent- and case-insensitive). |
| AC-3.3 | Type part of a ref, `src-4f` for `SRC-4F2NP` | That Source is shown. |
| AC-3.4 | Type `  ` (spaces only) | No filtering. |
| AC-3.5 | Type, status, and search all set, 3 of 8 rows match | Count line reads “3 of 8 sources”. |
| AC-3.6 | Filters match nothing; press Clear filters | Type is All types, status All statuses, search empty; sort unchanged; all rows shown. |

**US-4.** As a researcher, the list keeps my sort, type, status, and search.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-4.1 | Set sort A–Z, type Census, status Filed, search `smith`; open a Source; press Back | All four are still set. |
| AC-4.2 | Same settings; quit and relaunch | All four are still set. |
| AC-4.3 | Saved type id was deleted since | Type reads All types; other settings restore. |
| AC-4.4 | Two install users on one project, or one user on two projects | Each pair has its own settings. |
| AC-4.5 | Change status or search on this visit | The list scrolls to the top (scroll anchor cleared), as sort and type do today. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Last-updated column on the Sources row (SRL-1, SRL-2) | STEP-1, STEP-2, STEP-4, STEP-6; AC-1.1, AC-1.2, AC-1.3, AC-1.4 |
| `sources` reads move to `audit.LatestByScope` / `audit.LatestOf`; `sqlLatestRevision` and the `sqlList` subquery are removed | STEP-1; AC-1.4 |
| Primary and filed counts in one grouped read beside `sourcegraphprogress` | STEP-3; AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| Progress refresh after Promote | STEP-4; AC-2.3 |
| Status select (SRL-4), search field (SRL-3, SRL-5), combined filters and count line (SRL-6) | STEP-5, STEP-7; AC-2.x, AC-3.x |
| Filter-empty state with Clear filters (SRL-7) and counting state (SRL-8) | STEP-5, STEP-7; AC-2.6, AC-3.6 |
| Sort, type, status, query persisted per project UUID and user (SRL-9, folded S10-09) | STEP-5; AC-4.1, AC-4.2, AC-4.3, AC-4.4 |
| Status and search clear the scroll anchor (SRL-10) | STEP-7; AC-4.5 |
| Docs: audit skill, completed.md, map, brief archive | STEP-8 |

| Out | Home |
| --- | --- |
| Filters, search, or last updated on People, Events, Places | not planned (working rule); last updated there is S10-22 |
| Omnibar or FTS search of Sources | not planned (working rule) |
| A per-row count query | not planned (working rule) |
| Status badge on the row; bridges as a status axis | not planned (brief §2.1, working rule) |
| A new sort (by status, by time) beyond the existing four | not planned (brief §2.1) |
| Scroll restore | S10-40, landed [#392](https://github.com/mendahu/provenencia/pull/392) |
| Live “Today” → “Yesterday” clock | not planned (S10-22 D-9) |
| Reconciling settings across installs, or syncing them into the catalog | not planned |

## 4. Current state

Verified at `832ccc2`. Rows marked **(S10-22)** describe what S10-22's plan adds; recheck them when S10-22 lands (Implementer rule 2).

| Area | Where | Today |
| --- | --- | --- |
| Source list read | `core/database/sources/sources.go` `sqlList`, `List`, `scanSourceList` | One query with a correlated `MAX(t.revision)` subquery over `audit_transaction_scopes` per row, `scope_type = 'source'`. Fills `Source.UpdatedRevision int64`. No time. |
| One-Source revision | `core/database/sources/sources.go` `sqlLatestRevision`, `AttachLatestRevision` | Private copy of the scope read; skips the read when `UpdatedRevision > 0`. Called from `api/ffi/handlers/thumbnails.go` `enrichSourceProto` (Create, Update, SetCover, Get workspace). |
| Scope reads | `core/database/audit/scopes.go` `sqlLatestByScope`, `LatestByScope`, `Latest`, `Querier` | Map of every scope id to `{Revision, At}` for one scope type. No single-id read. Used by `conclusionheaders.latestHandles`. |
| Source FFI | `api/ffi/handlers/sources.go` `ListSources`, `CreateSource` (sets `s.UpdatedRevision = res.Revision`), `sourceProto` | Maps `UpdatedRevision` only. |
| Source proto | `api/proto/engine.proto` `message Source` | Fields 1–12; `updated_revision = 12`. No time. |
| Graph progress | `core/database/sourcegraphprogress/sourcegraphprogress.go` `List`, `Get`, `Progress`, `scanGrouped`, `canvasTypeSQL` | Two grouped counts (non-`source` subjects; observations). No primary or claim counts. |
| Graph progress FFI | `api/ffi/handlers/sourcegraphprogress.go` `protoProgress`; `engine.proto` `SourceGraphProgress` (fields 2, 3; 4 reserved) | Maps two counts. |
| Primary kinds | `core/database/promote/promote.go` `PrimaryKind(key, origin)` | `person`, `event`, `place` with origin `provenencia`. |
| Claim status | `core/database/migrations/000032.sql` `identity_claims` | `status IN ('provisional','accepted','rejected')`; unique accepted claim per subject. |
| Swift models | `macos/App/Platform/GenealogyStore.swift` `CatalogSource` (`updatedRevision`), `SourceGraphProgress` (`subjectCount`, `observationCount`, `isZero`) | No time; no primary or filed counts. |
| Swift mappers | `macos/App/Platform/GoStore.swift` `mapSource`, `mapSourceGraphProgress` | Map the fields above. |
| Fakes | `macos/App/Platform/FakeStore.swift` `listSources`, `graphProgress(for:projectDir:)`, `graphProgressBySource` | Progress computed from fake subjects, or the override map. |
| Progress refresh | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `refreshSourceGraphProgress(for:)` | Patches one Source on `.mutatedSourceGraph`, `.savedCitation`, `.deletedSubject`. **Not** on `.promotedSubject`, which `PromoteModel.file()` and the S9-04 card dispatch. Claim writes that return `CatalogWriteEffects` already patch via `apply(_ effects:)` (`identity_claims` has a `CacheSource` path in `core/database/effects/registry.go`). |
| List model | `macos/App/Features/Sources/SourcesModel.swift` `SourcesModel` (`Sort`, `typeFilterID`, `sort`, `visibleSources`, `countLine`) | Sort and type are in-memory; the model is `@State` in `SourcesListView`, so both reset on each visit. |
| List view | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent` (`toolbar`, `filterMenu`, `sortMenu`, `listBody`, `.onChange` for sort and type clearing the anchor) | Toolbar: type `PVSelect`, sort `PVSelect`. Filter-empty: `PVEmptyState(icon: .filter, title: noFilterMatchesTitle, …, compact: true)`. |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow.pageZone` | Thumb, title/type stack, mono ref. `accessibilityValue` = `"\(title), \(ref)"`. |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `body` `.sourcesList` | Builds `SourcesListView(session:userID:store:catalogCounts:)`. |
| Project id | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `projectUuid` | `private var`; set by `attachProject(uuid:)`; keys the history file. |
| Persistence precedent | `macos/App/Features/Workspace/WorkspaceModel.swift` `sidebarCollapsedDefaultsKey`; `macos/ProvenenciaTests/WorkspaceModelTests.swift` `makeDefaults()` | `UserDefaults` with an injectable suite in tests. |
| Search precedent | `macos/App/Features/Properties/PropertiesView.swift` `listToolbar`; `PropertiesModel.searchQuery` | `PVInput(text:size:prompt:icon: .search)` filtering a loaded list. |
| Last-updated cell **(S10-22)** | `macos/App/Platform/TimestampFormat.swift` `LastUpdatedDisplay` (`text`, `help`, `spoken`, `isKnown`, `columnWidth = 136`); `PVList.swift` `PVListDetail`, `PVListDetailColumn`, the row's detail cell (inline in `PVList`); `PVSymbol.history`; `L10n.LastUpdated` | Built for `PVList` rows. Sources rows are not `PVList`. |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources` (`noFilterMatchesTitle`, `noFilterMatchesMessage`, `countLine`, `countLineFiltered`, `sortUpdated` comment “stand-in until audit timestamps surface”) | |
| Docs | `.agents/skills/add-audit-scope/SKILL.md` “Read last activity under X” | Points at `sqlList` / `sqlLatestRevision`. |
| Tests | `core/database/sources/sources_test.go` `TestSources` (revision ordering); `core/database/audit/scopes_test.go` `scopeSeed.sourceRevisions`, `TestLatestByScope`; `core/database/sourcegraphprogress/sourcegraphprogress_test.go` `TestSourceGraphProgress`; `api/ffi/handlers/sources_test.go` `TestListSources`; `api/ffi/handlers/sourcegraphprogress_test.go`; `macos/ProvenenciaTests/SourcesModelTests.swift`; `WorkspaceSessionTests.swift` `applyMutatedSourceGraphPatchesProgressOnly` | Cover today's behavior. |

**Precedent.** [#398](https://github.com/mendahu/provenencia/pull/398) (S10-21) moved the Conclusion lists onto `audit.LatestByScope` and put `Updated audit.Latest` on each header: copy `conclusionheaders.latestHandles` and the `Updated` field. S10-22 maps `updated_at` from proto to `updatedAt` and builds `LastUpdatedDisplay` per row: copy its mapper and row build. `PropertiesView.listToolbar` is the search field; `WorkspaceModel` is the `UserDefaults` store with an injected suite.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `audit.LatestOf(q Querier, scopeType string, scopeID []byte) (Latest, error)` *(new)* in `scopes.go`, beside `LatestByScope`. SQL `sqlLatestOf` *(new)*: `SELECT COALESCE(MAX(t.revision), 0), COALESCE(t.created_at, '') FROM audit_transaction_scopes sc JOIN audit_transactions t ON t.id = sc.audit_transaction_id WHERE sc.scope_type = ? AND sc.scope_id = ?`, read with `q.Query` and one `rows.Next()`. Empty `scopeType` or `len(scopeID) != 16` → `ErrInvalid`. | One-Source responses (Create, Update, SetCover, workspace) patch the list row (`WorkspaceSession.patchUpdatedSource`), so they need the time too. Scope SQL lives in `audit`, not in each table package (add-audit-scope). | Keeping `sqlLatestRevision` in `sources` — a second copy of the scope read. Calling `LatestByScope` and picking one id — reads every Source for one row. |
| D-2 | `sources.Source.UpdatedRevision` is replaced by `Updated audit.Latest`. `List` drops the subquery from `sqlList` and fills `Updated` from `audit.LatestByScope(db, audit.ScopeSource)`. `AttachLatestRevision` is renamed `AttachLatest` *(new name)* and always reads `audit.LatestOf(db, audit.ScopeSource, s.ID)` (no `> 0` shortcut). `sqlLatestRevision` is deleted. | Same shape as `conclusionheaders` (`Updated audit.Latest`). The shortcut existed only so `CreateSource` could pass `res.Revision`; the create's own revision is committed before `enrichSourceProto` reads it. | Adding `UpdatedAt` beside `UpdatedRevision` — two fields for one value. |
| D-3 | `CreateSource` drops `s.UpdatedRevision = res.Revision`. | D-2. | — |
| D-4 | `engine.proto` `Source` gets `string updated_at = 13;` *(new)*, comment “RFC3339 UTC time of updated_revision; empty when 0”. `sourceProto` sets `UpdatedRevision: s.Updated.Revision, UpdatedAt: s.Updated.At`. The `updated_revision` comment changes “Sources list "Updated" sort” to “Sources list Date updated sort and last-updated column”. | S10-22 D-4 form. The revision stays: `SourcesModel.Sort.updated` sorts by it. | Sorting by the time string — revisions are the order. |
| D-5 | `sourcegraphprogress.Progress` gets `PrimaryCount int32` and `FiledCount int32` *(new)*. One grouped query `sqlCountPrimariesBySource` *(new)*: `SELECT s.source_id, t.key, t.origin, COUNT(*), COUNT(ic.subject_id) FROM subjects s JOIN subject_types t ON t.id = s.subject_type_id LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted' GROUP BY s.source_id, t.key, t.origin`. Go keeps rows where `promote.PrimaryKind(key, origin)` and adds them per Source. `Get` uses the same SQL with `WHERE s.source_id = ?` (`sqlCountPrimariesOne` *(new)*). `List` still omits Sources whose four counts are all zero. | One read for the list (working rule). `promote.PrimaryKind` already owns which kinds are primary, so the list and Promote cannot disagree. The unique accepted-claim index makes `COUNT(ic.subject_id)` the filed count. | A `t.key IN ('person','event','place')` literal — a second list of primaries. A per-Source query. Exporting a new key list from `promote`. |
| D-6 | `engine.proto` `SourceGraphProgress` gets `int32 primary_count = 5;` and `int32 filed_count = 6;` *(new)*; field 4 stays reserved. `protoProgress` maps both. | Additive; same message the Sources list already loads and patches. | A new FFI method — a second load and cache key for the same rows. |
| D-7 | Swift `SourceGraphProgress` gets `var primaryCount: Int = 0`, `var filedCount: Int = 0` *(new)*, and `var evidenceStatus: SourceEvidenceStatus` *(new)*: `primaryCount == 0` → `.empty`; `filedCount < primaryCount` → `.inProgress`; else `.filed`. `isZero` also requires both new counts to be 0. `SourceEvidenceStatus` *(new)* is `enum SourceEvidenceStatus: String, CaseIterable, Sendable { case empty, inProgress = "in_progress", filed }` beside it in `GenealogyStore.swift`. A Source missing from the progress map is `.empty`. | The rule is three lines; Go returns counts, which the graph zone may show later. Missing from the map means all zeros (Go omits them). | A status string from Go — the counts are the data; the label is UI. |
| D-8 | `WorkspaceSession.refreshSourceGraphProgress(for:)` also patches on `.promotedSubject(let id)`. | AC-2.3: `PromoteModel.file()` dispatches `.promotedSubject` and discards the effects. | Invalidating the whole progress map on promote — the registry comment forbids it. |
| D-9 | `CatalogSource` gets `var updatedAt: String = ""` *(new)*, mapped in `GoStore.mapSource`. `SourcesSplitRow` builds `LastUpdatedDisplay(source.updatedAt, now: Date())` once in `body`. | S10-22 D-4, D-9. | Formatting in the model. |
| D-10 | The detail cell S10-22 draws inline in `PVList` is extracted into `PVListDetailCell` *(new)*, a `View` in `PVList.swift` with `let detail: PVListDetail?` and `let width: CGFloat`, drawing exactly what S10-22 D-7 specifies (icon + text, muted or faint, help only when non-nil, `layoutPriority(1)`, fixed width, `nil` → empty column). `PVList` uses it; `SourcesSplitRow.pageZone` places `PVListDetailCell(detail: PVListDetail(text: d.text, help: d.help, icon: .history, faint: !d.isKnown), width: LastUpdatedDisplay.columnWidth)` between the title `VStack` and the ref `Text`. | Brief inventory: one cell, two hosts. The Sources row is a split row, not `PVList`. | Copying the cell into `SourcesSplitRow` — two copies that drift. Rebuilding the Sources row on `PVList` — the split zones are not a `PVList` row. |
| D-11 | `SourcesSplitRow.pageZone` `accessibilityValue` becomes `L10n.Conclusions.a11yList("\(source.title), \(source.ref)", rest: d.spoken)` (S10-22 D-10 joiner). | SRL-2. | A new joiner key. |
| D-12 | `SourcesListSettings` *(new)*: `struct SourcesListSettings: Codable, Equatable { var v = 1; var sort: String = "added"; var typeFilterID = ""; var status = ""; var query = "" }`, in `macos/App/Features/Sources/SourcesListSettings.swift` *(new)*, with `SourcesListSettingsStore` *(new)*: `init(defaults: UserDefaults, projectUuid: String, userID: String)`, `func load() -> SourcesListSettings`, `func save(_:)`. Key: `"pv.sourcesList.\(InstallPaths.navigationFileName(projectUuid:) without ".json").\(userID)"`. JSON-encoded `Data`. Missing, undecodable, or `v != 1` → defaults. Empty `projectUuid` or `userID` → `load` returns defaults and `save` does nothing. | Per project and user (working rule), keyed like the history file so a moved project folder keeps its settings. `UserDefaults` is the app's per-install preference store (`WorkspaceModel`). | A field on `NavigationHistoryDocument` — settings are not history and would reset with it. A catalog table — not research data. A file under Application Support — more code for four strings. |
| D-13 | `WorkspaceNavigation.projectUuid` becomes `private(set) var`. `WorkspaceDestinationHost` passes `projectUuid: navigation.projectUuid` to `SourcesListView`, which passes it and `defaults: .standard` to `SourcesModel.init`. `SourcesModel.init` gains `projectUuid: String = ""` and `defaults: UserDefaults = .standard` and loads settings there, before the first render. | Loading in `init` means the restored values are the initial values, so the `.onChange` handlers that clear the scroll anchor do not fire on restore (S10-40 anchor survives). | Loading in `.task` — the first frame shows defaults and the change clears the restored anchor. |
| D-14 | `SourcesModel` gains `var statusFilter: SourceEvidenceStatus?` (nil = all) and `var searchQuery = ""` *(new)*. `sort`, `typeFilterID`, `statusFilter`, `searchQuery` each save the whole settings value in `didSet`. On load, an unknown `sort` raw value → `.added`; unknown `status` → nil. | One write per change; four small strings. | Debouncing the query write — a keystroke-rate `UserDefaults` write is cheap. |
| D-15 | `visibleSources` filters in order type → status → search, then sorts. Status: `evidenceStatus(for:) == statusFilter`, where `evidenceStatus(for:)` *(new)* reads the progress map (missing → `.empty`). Search: `let q = searchQuery.trimmingCharacters(in: .whitespacesAndNewlines)`; when non-empty, keep rows where `title.localizedStandardContains(q) || ref.localizedStandardContains(q)`. | AC-3.1 to AC-3.4: case- and accent-insensitive, locale-aware. | `lowercased().contains` (Properties) — misses `jose` → `José`. |
| D-16 | `var isFiltering: Bool` *(new)*: type non-empty, or status non-nil, or trimmed query non-empty. `countLine` uses it: filtered form when `isFiltering` and `visible != total`. `func clearFilters()` *(new)* sets type `""`, status nil, query `""`; sort unchanged. `var isAwaitingStatusCounts: Bool` *(new)*: `statusFilter != nil && graphCountsLoading`. | SRL-6, SRL-7, SRL-8. `graphCountsLoading` already exists. | — |
| D-17 | A restored `typeFilterID` that is not in `types` once types load reads as All types: `visibleSources` ignores a type id not in `types` when `types` is non-empty, and `filterMenu` shows All types. The stored value is overwritten only when the researcher changes the type. | AC-4.3 without a write on load. | Clearing it on load — types load after the list. |
| D-18 | `SourcesListContent.listBody`: if `model.isAwaitingStatusCounts` → the existing loading `ProgressView`; else if `visibleSources.isEmpty` → `PVEmptyState(icon: .filter, title: L10n.Sources.noFilterMatchesTitle, message: …, compact: true) { PVButton(L10n.Sources.clearFilters, variant: .secondary) { model.clearFilters() } }` with id `sources.clearFilters` on the button. `.onChange(of: model.statusFilter)` and `.onChange(of: model.searchQuery)` clear the anchor the same way the existing sort and type handlers do. | SRL-7, SRL-8, SRL-10. | A separate “no search results” state — one state for every filter (brief). |
| D-19 | Toolbar order: `searchField`, `filterMenu`, `statusMenu`, `sortMenu`, `Spacer`. `searchField` *(new)* is `PVInput(text: $model.searchQuery, prompt: L10n.Sources.searchPrompt, icon: .search)` with `.frame(minWidth: 160, maxWidth: 280)` and identifier `sources.search`. `statusMenu` *(new)* is `PVSelect` built like `filterMenu`, options `""` → All statuses then `SourceEvidenceStatus.allCases`, no `icon`, `menuWidth: 220`, `fillsWidth: false`, identifiers `sources.status`, `sources.status.all`, `sources.status.<rawValue>`. | Brief SRL-3, SRL-4; Properties search precedent; same Select as type. | A segmented control or chips (brief non-goal). |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `audit.LatestOf`, `sqlLatestOf` *(new)* | func | `audit.LatestByScope`; `sources.sqlLatestRevision` | Single-id sibling of `LatestByScope`; replaces the private copy in `sources`. |
| `sources.AttachLatest` *(new name)* | func | `sources.AttachLatestRevision` | Same job, now fills `Updated`. |
| `Source.updated_at` *(new)* | proto field | `PersonHeader.updated_at` | Same field on a different message. |
| `Progress.PrimaryCount`, `FiledCount`, `sqlCountPrimariesBySource`, `sqlCountPrimariesOne` *(new)* | fields / SQL | `sqlCountSubjectsBySource` | Needs claim join and the type key for `promote.PrimaryKind`. |
| `SourceGraphProgress.primary_count`, `filed_count` *(new)* | proto fields | `subject_count` | Same message. |
| `SourceEvidenceStatus`, `SourceGraphProgress.evidenceStatus` *(new)* | enum / computed | `SourceGraphProgress.isZero` | Derived beside the counts it reads. |
| `CatalogSource.updatedAt` *(new)* | field | `CatalogPersonHeader.updatedAt` (S10-22) | Same shape. |
| `PVListDetailCell` *(new)* | component view | S10-22's inline cell in `PVList` | Extracted so the Sources row uses the same cell. |
| `SourcesListSettings`, `SourcesListSettingsStore` *(new)* | struct / file | `WorkspaceModel` defaults key; `NavigationHistoryStore` | Four values for one place; a defaults key keyed by project and user. |
| `SourcesModel.statusFilter`, `searchQuery`, `isFiltering`, `clearFilters`, `isAwaitingStatusCounts`, `evidenceStatus(for:)` *(new)* | properties / funcs | `typeFilterID`, `graphCountsLoading` | Same model. |
| `L10n.Sources.searchPrompt`, `statusAll`, `statusEmpty`, `statusInProgress`, `statusFiled`, `statusMenu`, `clearFilters` *(new)* | keys | `filterAllTypes`, `filterMenu` | Same namespace. |

**Builds on.** `audit` scope reads (add-audit-scope), the `sourcegraphprogress` grouped counts and their session cache key and patch path, `promote.PrimaryKind`, S10-22's `LastUpdatedDisplay` and `PVList` detail cell, `PVInput` / `PVSelect` / `PVEmptyState`, the S10-40 anchor clearing.
**Deletes.** `sources.sqlLatestRevision`; the correlated subquery in `sqlList`; `Source.UpdatedRevision`; `AttachLatestRevision`'s `> 0` shortcut; `CreateSource`'s `s.UpdatedRevision = res.Revision`; the `add-audit-scope` Read section's pointer to `sqlList` / `sqlLatestRevision` (replaced with `audit.LatestOf` / `LatestByScope`); the stale `sortUpdated` comment (“stand-in until audit timestamps surface”).
**Next adds.** A later per-place settings need (a Persons sort) gets its own settings value with the same key scheme. A status shown on the row reads `evidenceStatus`. The Conclusion detail pages could show a time with `audit.LatestOf(db, ScopeCanonicalEntity, id)`.

### Schema

None — no catalog change. Scopes are already stored.

### Go core

| Package | Change |
| --- | --- |
| `core/database/audit` | `LatestOf` *(new)* (D-1). |
| `core/database/sources` | `Source.Updated audit.Latest` (D-2); `sqlList` without the subquery; `List` fills from `LatestByScope`; `AttachLatest` (D-2); delete `sqlLatestRevision`. |
| `core/database/sourcegraphprogress` | D-5. Package comment: “aggregates Evidence-graph subject, observation, primary, and filed counts per Source for the Sources list”. |

```go
// LatestOf returns one scope id's newest revision and its time. The zero
// value means no revision.
func LatestOf(q Querier, scopeType string, scopeID []byte) (Latest, error)
```

### FFI

`api/proto/engine.proto`: `Source.updated_at = 13` (D-4); `SourceGraphProgress.primary_count = 5`, `filed_count = 6` (D-6). Handlers: `sources.go` `sourceProto`, `CreateSource` (D-3); `thumbnails.go` `enrichSourceProto` calls `sources.AttachLatest`; `sourcegraphprogress.go` `protoProgress`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Last-updated cell | `PVListDetailCell` (extracted) | `macos/App/DesignSystem/Components/List/PVList.swift` | D-10. |
| Split row | snowflake | `macos/App/Features/Sources/SourcesSplitRow.swift` | D-9, D-10, D-11. |
| Toolbar, list states | snowflake | `macos/App/Features/Sources/SourcesListView.swift` | D-18, D-19. |
| Model | — | `macos/App/Features/Sources/SourcesModel.swift` | D-13, D-14, D-15, D-16, D-17. |
| Settings | — | `macos/App/Features/Sources/SourcesListSettings.swift` *(new)* | D-12. Add to the App target in `project.pbxproj` (add-swift-test explains target membership). |
| Host | — | `WorkspaceDestinationHost.swift`, `WorkspaceNavigation.swift` | D-13. |

Frames, in words (today 14 Oct 2026, Mountain Time; the board may adjust wording — see Stop conditions). **Times** — eight rows; in each filing zone, after the title over the type line, a 136pt column with the history icon and “Today, 3:15 PM”, “Today, 9:02 AM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, or a faint “—”, then the mono ref; the 248pt Evidence graph zone unchanged; count line “8 sources”. **Help** — “Last updated 13 Oct 2026 · 4:40 PM MDT” on hover. **Status menu** — All statuses, Empty, In progress, Filed; All checked. **Filtered** — In progress + Census → three rows, “3 of 8 sources”. **Search** — `smith` → two rows, “2 of 8 sources”. **Nothing matches** — toolbar stays; compact empty state “No matching sources” / “Try another filter, or clear them to see every source.” with a secondary Clear filters button. **Counting** — status set, counts not loaded → loading spinner. **Narrow** — title and type truncate; column, ref, graph zone, and the three selects keep width; search shrinks to its 160pt minimum.

States: empty catalog → unchanged `sources.empty`; load error → unchanged; 1000 Sources → filtering and sorting are in-memory on the loaded array, one `LastUpdatedDisplay` per visible row in a `LazyVStack`. Accessibility: D-11; the search field's prompt is its label; the status select's `accessibilityLabel` is `L10n.Sources.statusMenu`.

### Strings

L10n keys *(new)* in `L10n.Sources`, each with a `comment`, added to `Localizable.xcstrings`:
- `sources.list.searchPrompt` = "Search title or ref"
- `sources.list.statusMenu` = "Evidence status" — accessibility label of the status select
- `sources.list.status.all` = "All statuses"
- `sources.list.status.empty` = "Empty"
- `sources.list.status.inProgress` = "In progress"
- `sources.list.status.filed` = "Filed"
- `sources.list.clearFilters` = "Clear filters"

`SourceEvidenceStatus` gets `var label: LocalizedStringResource` mapping to the three status keys (in `SourcesModel.swift` as an extension, so the store file stays UI-free).

Changed: `sources.list.noFilterMatchesTitle` → "No matching sources"; `sources.list.noFilterMatchesMessage` → "Try another filter, or clear them to see every source."; both comments say they cover type, status, and search. `sources.list.sort.updated` comment → “Sources sort option: most recent work first (audit revision)”.

## 7. Steps

**STEP-1 — Scope reads live in audit**
Files: `core/database/audit/scopes.go`, `core/database/audit/scopes_test.go`, `core/database/sources/sources.go`, `core/database/sources/sources_test.go`, `api/ffi/handlers/thumbnails.go`, `api/ffi/handlers/sources.go`. Skill: `add-audit-scope`, `add-catalog-query`.
Do: D-1, D-2, D-3. Update `scopeSeed.sourceRevisions` and `TestSources` to read `Updated.Revision`. Add T-1, T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/audit/ ./core/database/sources/ ./api/...` passes; `rg -n "sqlLatestRevision|UpdatedRevision" core/database/sources` prints nothing.

**STEP-2 — Source time on the wire**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/sources.go`, `api/ffi/handlers/sources_test.go`. Skill: `add-ffi-handler`.
Do: D-4. Run `./scripts/generate-proto.sh`. Add T-3.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `git diff --stat` lists the regenerated Go and Swift protobuf files.

**STEP-3 — Primary and filed counts**
Files: `core/database/sourcegraphprogress/sourcegraphprogress.go`, `core/database/sourcegraphprogress/sourcegraphprogress_test.go`, `api/proto/engine.proto`, generated files, `api/ffi/handlers/sourcegraphprogress.go`, `api/ffi/handlers/sourcegraphprogress_test.go`. Skill: `add-catalog-query`, `add-ffi-handler`.
Do: D-5, D-6. Run `./scripts/generate-proto.sh`. Add T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcegraphprogress/ ./api/...` passes.

**STEP-4 — Swift models, mappers, fakes, promote refresh**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`. Skill: `add-swift-test`.
Do: D-7, D-8, D-9 model field. `FakeStore.graphProgress(for:projectDir:)` sets `primaryCount` to the subjects whose type key is `person`, `event`, or `place`, and `filedCount` 0, unless `graphProgressBySource` overrides. Add T-6, T-7.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/SourcesModelTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-5 — Model: status, search, saved settings**
Files: `macos/App/Features/Sources/SourcesListSettings.swift` *(new)*, `SourcesModel.swift`, `SourcesListView.swift` (init only), `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `WorkspaceNavigation.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/ProvenenciaTests/SourcesListSettingsTests.swift` *(new)*, `SourcesModelTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: D-12, D-13, D-14, D-15, D-16, D-17. Add T-8, T-9, T-10, T-11, T-12.
Check: the STEP-4 `xcodebuild` command plus `-only-testing:ProvenenciaTests/SourcesListSettingsTests` passes.

**STEP-6 — Last-updated column on the row**
Files: `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/Features/Sources/SourcesSplitRow.swift`, `macos/ProvenenciaTests/PVListTests.swift`. Skill: `add-ui-component`.
Do: D-10, D-11. `PVList` output is unchanged by the extraction.
Check: the STEP-4 `xcodebuild` command with `-only-testing:ProvenenciaTests/PVListTests` passes; app builds; Times frame visible in the app.

**STEP-7 — Toolbar, empty and counting states, copy**
Files: `macos/App/Features/Sources/SourcesListView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`, `add-ui-component`.
Do: D-18, D-19, §6 Strings.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; manual check steps 2–6.

**STEP-8 — Docs**
Files: `.agents/skills/add-audit-scope/SKILL.md` (Read section names `audit.LatestOf` for one id and `audit.LatestByScope` for a list), `docs/deployment-plan/spike-10/completed.md`, `deployment-plan.md` (S10-41 row and slice 6 line), this plan (status, Landed note), `design/S10-D8-sources-list.md` → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-41-sources-list.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show when each Source last changed on the Sources list."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/audit/scopes_test.go` `TestLatestOf` *(new)*, table | source with two writes → revision of the second and its `created_at`; unknown id → zero `Latest`; empty type → `ErrInvalid`; 15-byte id → `ErrInvalid` | D-1 |
| T-2 | Go | `core/database/sources/sources_test.go` `TestSources` (extend) | `List`: each `Updated.Revision` > 0 and `Updated.At` parses as RFC3339; updated Source leads on revision; `AttachLatest` on a fresh `Get` equals the `List` value | AC-1.4, D-2 |
| T-3 | Go | `api/ffi/handlers/sources_test.go` `TestListSources`, `TestCreateSourceAndWorkspace` (extend) | list rows and the created Source carry non-empty `updated_at` and `updated_revision` > 0 | AC-1.4, D-3, D-4 |
| T-4 | Go | `core/database/sourcegraphprogress/sourcegraphprogress_test.go` `TestSourceGraphProgress` (extend, table) | no subjects → absent / zeros; one person, no claim → primary 1 filed 0; same person with accepted claim → 1/1; provisional claim → 1/0; participation bridge alone → primary 0; person + place, one accepted → 2/1; `Get` equals the `List` row | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5 |
| T-5 | Go | `api/ffi/handlers/sourcegraphprogress_test.go` `TestListSourceGraphProgress` (extend) | `primary_count` and `filed_count` on the wire | D-6 |
| T-6 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `evidenceStatusFromCounts` *(new)* | (0,0) → empty; (2,1) → inProgress; (2,2) → filed; missing from map → empty | D-7 |
| T-7 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `applyPromotedSubjectPatchesProgress` *(new)* | warm progress; set fake override for s1 to filed; `apply(.promotedSubject(sourceId: "s1"))` → `getSourceGraphProgressCalls == 1` and the map's s1 row is filed | AC-2.3, D-8 |
| T-8 | Swift | `macos/ProvenenciaTests/SourcesListSettingsTests.swift` *(new)* `roundTrip`, `perProjectAndUser`, `badDataReadsDefaults`, `emptyKeysDoNotSave` | save then load equal; (p1,u1) vs (p1,u2) vs (p2,u1) independent; garbage `Data` and `v = 2` → defaults; empty uuid → defaults, nothing written | AC-4.4, D-12 |
| T-9 | Swift | `SourcesModelTests.swift` `settingsRestoreOnNewModel` *(new)* | model A sets sort `.az`, type `t1`, status `.filed`, query `smith`; model B on the same defaults, uuid, user → all four equal | AC-4.1, AC-4.2, D-13, D-14 |
| T-10 | Swift | `SourcesModelTests.swift` `searchMatchesTitleAndRefIgnoringCaseAndAccents` *(new)* | the AC-3.1, AC-3.2, AC-3.3, AC-3.4 table | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-15 |
| T-11 | Swift | `SourcesModelTests.swift` `filtersCombineAndClear` *(new)* | type + status + query → expected rows and “3 of 8 sources”; `clearFilters()` → all rows, sort unchanged; unknown restored type id with types loaded → all rows | AC-3.5, AC-3.6, AC-4.3, D-16, D-17 |
| T-12 | Swift | `SourcesModelTests.swift` `statusFilterWaitsForCounts` *(new)* | status `.empty` before progress loads → `isAwaitingStatusCounts == true`; after load → false and rows filtered | AC-2.6, D-16 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `core/database/audit` `TestSourceScopes` (reads `sources.List`), `SourcesModelTests.sortOrdersByAddedUpdatedAndTitle`, `PVListTests`, `CatalogQueryRegistryTests`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says — none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: S10-41 status and slice 6 line; completed.md: one short paragraph of what shipped
- [ ] Brief S10-D8 moved to design/archive/ and design/README.md updated
- [ ] add-audit-scope skill: Read section names audit.LatestOf / audit.LatestByScope
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with at least eight Sources of two types: one with no subjects, one with an unpromoted person, one whose person is promoted, one with a note edited today.

1. Open Sources. Each row shows the history icon and a time; the edited Source reads “Today,” and its time; hover shows “Last updated” and the full stamp (AC-1.1). A Source imported before scopes existed, if any, shows a faint “—”; VoiceOver on its filing zone ends “last updated unknown” (AC-1.2).
2. Status → Empty: only the subject-less Source (AC-2.1). In progress: the unpromoted one (AC-2.2). Filed: the promoted one.
3. Open the In progress Source's Promote page, file the person, go back to Sources (status still In progress): the Source is gone from the list; switch to Filed: it is there.
4. Type part of a title, then part of a ref (AC-3.3), then a title without its accent (AC-3.2): rows narrow as you type; the count line reads “N of M sources”.
5. Type `zzz`: the empty state shows Clear filters; press it: every row is back, the sort is unchanged.
6. Scroll down, change status, then scroll down and type in search: each time the list jumps to the top (AC-4.5).
7. Set sort A–Z, a type, Filed, and a query. Open a Source, press Back: all four are set. Quit and relaunch: all four are set.
8. Edit a Source's note, open People, return to Sources: that row's time moved (AC-1.3). People shows no search, status, or type controls.
9. In the diff, `sourcegraphprogress.List` runs a fixed set of grouped queries with no loop over Sources (AC-2.5).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A persisted query hides rows after relaunch and looks like data loss | Count line “N of M sources” and the Clear filters empty state (SRL-6, SRL-7). |
| `MAX()` with a bare `created_at` in SQLite returns the row of the max | Same idiom `sqlLatestByScope` already relies on; T-1 checks the time belongs to the newest revision. |
| Promote via a path that neither dispatches `.promotedSubject` nor returns `CatalogWriteEffects` | Manual check step 3; STEP-4 greps `promotedSubject` dispatch sites. |

**Stop and report if:**

- S10-22 landed without `LastUpdatedDisplay`, `PVListDetail`, `PVSymbol.history`, or `L10n.LastUpdated` as its D-5, D-6, D-7, D-8 describe, or its detail cell cannot be extracted without changing `PVList` output.
- The agreed S10-D8 board differs from SRL-1 to SRL-10 or the copy in §6 Strings (the plan is refreshed first).
- `sources.List` or `AttachLatestRevision` has a caller not listed in §4.
- `promote.PrimaryKind` is no longer exported, or importing `promote` from `sourcegraphprogress` creates an import cycle.
- A promote path in `macos/App` dispatches neither `.promotedSubject` nor `apply(_ effects:)` with the Source.
- `WorkspaceNavigation.projectUuid` is empty when the Sources list first mounts after relaunch.

## 12. Open questions

None. The plan is Blocked until S10-22 lands and the S10-D8 board is agreed; both trigger a refresh (plan-pr “Refresh an existing plan”).

## 13. PR

**Title:** S10-41 — Sources list: last updated, status filter, search, saved settings
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-41-sources-list.md`.
