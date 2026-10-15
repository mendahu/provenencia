# S10-41 — Sources list: last updated, status filter, search, saved settings

**Status:** Landed in [#405](https://github.com/mendahu/provenencia/pull/405) (2026-10-15)
**Spike / slice:** Spike 10 · Slice 6 — Lists and the Person page
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-41--sources-list)
**Brief:** [S10-D8](../design/archive/S10-D8-sources-list.md) (SRL-1…SRL-10), board agreed 2026-10-15
**Board:** Claude Design project *Sources list* (id 118d2eb7-3b46-4912-a688-09ec7b0070fa, file *S2-04 Sources list.dc.html*, section “S10-D8 enhancement” and frames S10-D8 1–8 with 7b). §6 macOS describes every frame this PR ships in words. The implementer does not need the board.
**Depends on:** S10-22 (landed [#402](https://github.com/mendahu/provenencia/pull/402)), S10-D8 (board agreed 2026-10-15)
**Blocks:** —
**Branch:** `s10-41-sources-list`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-catalog-query`](../../../../.agents/skills/add-catalog-query/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`
**Context:** working rules "The Sources list is one page", "Last updated is an audit scope", "Sources evidence filters count primaries in one read", "List scroll is a field on the history entry"; [`navigation-history.md`](../../archive/spike-3/navigation-history.md); [`conclusion-layer-data-model.md`](../../../conclusion-layer-data-model.md) (Identity Claims, bridges); [S10-22 plan](S10-22-list-last-updated.md) Landed note, D-20, D-21
**Planned:** 2026-10-15 against `main` @ `832ccc2`; refreshed 2026-10-15 against `main` @ `6f70bf4` after S10-22 landed; D-13 revised 2026-10-15 at `daf4e086` after the round-0 stop (project attached in `WorkspaceView.init`)
**Review:** [`S10-41-sources-list.review.md`](S10-41-sources-list.review.md) — two rounds; Fits / Ready; one To user item

## Landed

Differences from this plan:
- None in behavior. The STEP-6 `rg -n "updatedDetail|labelWithUpdated|updatedColumnWidth" macos/App` check still prints `L10n.Promote.updatedDetail` and `PromoteView`: an unrelated Promote string that shares the word. No Conclusion or Sources symbol remains.
- The plan's `.onChange(of: model.settings)` replaces the sort and type handlers in STEP-3 rather than STEP-5, because the model change removed `model.sort` and `model.typeFilterID` from the view in the same commit.
- `lastUpdated.a11y.join` was added with `PVLastUpdated` (STEP-4); the three status labels with the model (STEP-3); the other `sources.list` keys with the toolbar (STEP-5).

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

**Why now.** Slice 6 is the lists slice. S10-09 (persist sort and type filter) was folded into this PR so the Sources toolbar is designed and persisted once (map, 2026-10-12). S10-22 already reads each Source's last activity from the engine's display cache, puts it on `CatalogSource.updated`, and patches it after every write; it left the Sources row's cell, and moving the cell composition to a shared home, to this PR (S10-22 D-21).

**What stays.** The split row (filing zone | Evidence graph zone), its counts, the caption band, the type filter and sort options, row hover and keyboard, the Add Source dialog, the empty catalog state, and the load error. Scroll restore is S10-40 and is unchanged. People, Events, and Places get none of these controls.

## 2. User stories

**US-1.** As a researcher, when I scan the Sources list, I see when each Source last changed, so I can pick up recent work.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source whose newest scoped revision was today at 15:15 local | The row's column reads “Today, 3:15 PM” (`en_US`) behind the history icon; help reads “Last updated 14 Oct 2026 · 3:15 PM MDT”. |
| AC-1.2 | A Source with no recorded activity (`CatalogSource.updated.at` empty) | The column reads a faint “—” with no help; the filing zone's spoken value ends “last updated unknown”. |
| AC-1.3 | Edit a Source's note, go to another place, come back | That row's time moved; the list order follows the sort, not the time (unless the sort is Date updated). |
| AC-1.4 | The Conclusion lists after the cell composition moves (D-10) | Persons, Events, and Places rows read and speak exactly as before. |

**US-2.** As a researcher, I filter the list by evidence status so I can find graphs that still need promoting.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | A Source with no person, event, or place subjects (a participation-only graph counts too) | Matches **Empty**, not In progress or Fully promoted. |
| AC-2.2 | A Source with one person subject and no accepted claim on it | Matches **In progress**. |
| AC-2.3 | Promote that person (accepted Identity Claim) on the Promote page, return to Sources | The Source matches **Fully promoted**, without relaunch. |
| AC-2.4 | A person with only a `provisional` claim | Counts as not filed: **In progress**. |
| AC-2.5 | 200 Sources | `ListSourceGraphProgress` issues the same fixed number of queries as for 2 Sources (no per-row query). |
| AC-2.6 | Status **Empty** set and the progress map not loaded yet | Below the toolbar, the first-load `ProgressView`, not the filter-empty state. If the progress load fails, a danger callout instead. |

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
| AC-4.1 | Set sort A–Z, type Census, status Fully promoted, search `smith`; open a Source; press Back | All four are still set. |
| AC-4.2 | Same settings; quit and relaunch | All four are still set. |
| AC-4.3 | Saved type id was deleted since | Type reads All types; other settings restore. |
| AC-4.4 | Two install users on one project, or one user on two projects | Each pair has its own settings. |
| AC-4.5 | Change status or search on this visit | The list scrolls to the top (scroll anchor cleared), as sort and type do today. |

## 3. Scope

| In | Covered by |
| --- | --- |
| Last-updated column on the Sources row (SRL-1, SRL-2), drawn from `CatalogSource.updated` | STEP-4; AC-1.1, AC-1.2, AC-1.3 |
| The cell composition moves from `ConclusionListRow` to one shared home; `PVList`'s cell becomes a view both rows draw (S10-22 D-21) | STEP-4; AC-1.4 |
| Primary and filed counts in one grouped read beside `sourcegraphprogress` | STEP-1; AC-2.1, AC-2.2, AC-2.3, AC-2.4, AC-2.5 |
| Progress refresh after Promote | STEP-2; AC-2.3 |
| Status select (SRL-4), search field (SRL-3, SRL-5), combined filters and count line (SRL-6) | STEP-3, STEP-5; AC-2.6, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5 |
| Filter-empty state with Clear filters (SRL-7) and counting state (SRL-8) | STEP-3, STEP-5; AC-2.6, AC-3.6 |
| Sort, type, status, query persisted per project UUID and user (SRL-9, folded S10-09) | STEP-3; AC-4.1, AC-4.2, AC-4.3, AC-4.4 |
| Status and search clear the scroll anchor (SRL-10) | STEP-5; AC-4.5 |
| Docs: completed.md, map, brief archive | STEP-6 |

| Out | Home |
| --- | --- |
| Filters, search, or last updated on People, Events, Places | not planned (working rule); last updated there is S10-22 |
| Omnibar or FTS search of Sources | not planned (working rule) |
| A per-row count query | not planned (working rule) |
| Status badge on the row; bridges as a status axis | not planned (brief §2.1, working rule) |
| A new sort (by status, by time) beyond the existing four | not planned (brief §2.1) |
| Scroll restore | S10-40, landed [#392](https://github.com/mendahu/provenencia/pull/392) |
| Live “Today” → “Yesterday” clock | not planned (S10-22) |
| Reading or patching the Sources time (engine cache, `WriteEffects.sourceScopes`, the session patch) | landed in S10-22 |
| Promote returning `WriteEffects` | S10-22 follow-up (its Landed note); this PR adds one mutation case instead (D-8) |
| Reconciling settings across installs, or syncing them into the catalog | not planned |

## 4. Current state

Verified at `6f70bf4` (S10-22 merged).

| Area | Where | Today |
| --- | --- | --- |
| Source time | `core/database/sources/sources.go` `List`, `AttachActivity`, `Source.Updated effects.Activity`; `core/database/graphcache/activity.go` `Graph.Activity`; `api/ffi/handlers/sources.go` `sourceProto`; `engine.proto` `Source.updated_revision = 12`, `updated_at = 13` | Done by S10-22: every Source read carries its last activity from the display cache. Nothing to change. |
| Swift Source time | `macos/App/Platform/GenealogyStore.swift` `CatalogSource.updated: CatalogActivity`; `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` `CatalogActivity`, `CatalogLastUpdatedPatch`; `CatalogQueryRegistry` `.sourcesList` spec (`lastUpdated: .rows(CatalogSource.self, scope: .source)`) | Done by S10-22: mapped in `GoStore` and patched in place after writes. `SourcesModel.visibleSources` `.updated` sorts by `updated.revision`. |
| Cell composition | `macos/App/Features/Conclusions/ConclusionListRow.swift` `updatedColumnWidth` (136), `updatedColumn(_:)`, `updatedDetail(_:in:)`, `labelWithUpdated(_:updatedAt:in:)` | Builds the cell from `TimestampFormat.string(_, .relative / .stamp / .spoken, in:)`; the label joins with `L10n.Conclusions.a11yList`. Callers: `ConclusionListPage.swift` (`PVListSkeleton(detailWidth:)` at line 107, `updatedColumn` and `labelWithUpdated` in `ConclusionListBody`), `macos/ProvenenciaTests/PersonsListTests.swift` `rowLabelEndsWithUpdated`, `updatedDetailComposesTheCell`. |
| Cell view | `macos/App/DesignSystem/Components/List/PVList.swift` `private func detailCell(_:)`; `PVListDetail`, `PVListDetailColumn` | The cell (icon 14, caption text, muted or faint, help when non-nil) is private to `PVList`, which frames it to `detail.width` with `layoutPriority(1)`. |
| Strings | `macos/App/Platform/L10n.swift` `L10n.LastUpdated` (`help`, `a11yUpdated`, `a11yUnknown`), `L10n.Timestamp`; `PVSymbol.history` | Landed by S10-22. |
| Graph progress | `core/database/sourcegraphprogress/sourcegraphprogress.go` `List`, `Get`, `Progress`, `scanGrouped`, `canvasTypeSQL` | Unchanged by S10-22. Two grouped counts (non-`source` subjects; observations). No primary or claim counts. |
| Graph progress FFI | `api/ffi/handlers/sourcegraphprogress.go` `protoProgress`; `engine.proto` `SourceGraphProgress` (fields 2, 3; 4 reserved) | Maps two counts. |
| Primary kinds | `core/database/promote/promote.go` `PrimaryKind(key, origin)` | `person`, `event`, `place` with origin `provenencia`. |
| Claim status | `core/database/migrations/000032.sql` `identity_claims` | `status IN ('provisional','accepted','rejected')`; unique accepted claim per subject. |
| Swift progress | `macos/App/Platform/GenealogyStore.swift` `SourceGraphProgress` (`subjectCount`, `observationCount`, `isZero`); `GoStore.mapSourceGraphProgress` | No primary or filed counts. |
| Fakes | `macos/App/Platform/FakeStore.swift` `graphProgress(for:projectDir:)`, `graphProgressBySource` | Progress computed from fake subjects, or the override map. |
| Progress refresh | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `apply(_:_:)`, `graphProgressSource(of:)`, `refreshSourceGraphProgress(sourceId:)` | Patches one Source for each id in `effects.sources` and for `.mutatedSourceGraph` / `.deletedSubject`. **Not** `.promotedSubject`, which `PromoteModel.file()` dispatches with no effects. Claim writes that return effects already patch (`identity_claims` keeps its `Source` path in `core/database/effects/registry.go`). |
| List model | `macos/App/Features/Sources/SourcesModel.swift` `SourcesModel` (`Sort`, `typeFilterID`, `sort`, `visibleSources`, `countLine`) | Sort and type are in-memory; the model is `@State` in `SourcesListView`, so both reset on each visit. |
| List view | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent` (`toolbar`, `filterMenu`, `sortMenu`, `listBody`, `.onChange` for sort and type clearing the anchor) | Toolbar: type `PVSelect`, sort `PVSelect`, both with icons (chips). Filter-empty: `PVEmptyState(icon: .filter, title: noFilterMatchesTitle, …, compact: true)`. |
| Row | `macos/App/Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow.pageZone` | Thumb, title/type stack, mono ref. `accessibilityValue` = `"\(title), \(ref)"`. |
| Host | `macos/App/Features/Workspace/WorkspaceDestinationHost.swift` `body` `.sourcesList` | Builds `SourcesListView(session:userID:store:catalogCounts:)`. |
| Project id | `macos/App/Features/Workspace/WorkspaceNavigation.swift` `projectUuid` | `private var`; set by `attachProject(uuid:)`; keys the history file. |
| Project attach | `macos/App/Features/Workspace/WorkspaceView.swift` `init`, `.onAppear`, `.onChange(of: model.project?.uuid)` | `init` builds `WorkspaceNavigation()` with `projectUuid == ""` and wires `onLocationCommit`; `attachProject(uuid:)` runs only in `.onAppear`, after the first body evaluation, so `WorkspaceDestinationHost` (always mounted by `WorkspaceContent`) builds the seed `.sourcesList` view first. `OnboardingModel` sets `project` before `enterHome` on both paths (`loadProjectInfo` is awaited; `applyOpened` sets it first), so `model.project?.uuid` is readable in `init`. `attachProject` is synchronous and idempotent for the same uuid once history loads. When `projectInfo` fails, `loadProjectInfo` sets a fallback `ProjectInfo` whose `uuid` is `""`: history is not attached (as today), `SourcesListSettingsStore` is nil, and the list keeps settings in memory only. That path is by design, not a stop. |
| Persistence home | `macos/App/Platform/InstallPaths.swift` `navigationDirectory`, `navigationFileName(projectUuid:)`, `navigationFile`; `macos/App/Features/Workspace/NavigationHistoryStore.swift` `persist`; [`navigation-history.md`](../../archive/spike-3/navigation-history.md) “Persistence format” | Per-project state is JSON under Application Support keyed by the project UUID hex, written atomically. Not `UserDefaults` (also `add-workspace-location` Do-not list). |
| Catalog connection | `core/database/catalog.go` `db.SetMaxOpenConns(1)` | One connection: a query issued while another's rows are open blocks. |
| Search precedent | `macos/App/Features/Properties/PropertiesView.swift` `listToolbar`; `PropertiesModel.searchQuery` | `PVInput(text:size:prompt:icon: .search)` filtering a loaded list. |
| Select chip | `macos/App/DesignSystem/Components/Select/PVSelect.swift` `body` | `icon != nil` draws `chipTrigger`; no icon draws the field trigger. |
| Copy | `macos/App/Platform/L10n.swift` `L10n.Sources` (`noFilterMatchesTitle`, `noFilterMatchesMessage`, `countLine`, `countLineFiltered`, `sortUpdated` comment “stand-in until audit timestamps surface”) | |
| Tests | `core/database/sourcegraphprogress/sourcegraphprogress_test.go` `TestSourceGraphProgress`; `api/ffi/handlers/sourcegraphprogress_test.go` `TestListSourceGraphProgress`; `macos/ProvenenciaTests/SourcesModelTests.swift`; `WorkspaceSessionTests.swift` `applyMutatedSourceGraphPatchesProgressOnly`; `PersonsListTests.swift`; `PVListTests.swift` | Cover today's behavior. |

**Precedent.** [#402](https://github.com/mendahu/provenencia/pull/402) (S10-22) built the cell on `PVList` rows and `ConclusionListRow`: move it, do not copy it. `PropertiesView.listToolbar` is the search field. `NavigationHistoryStore` and `InstallPaths.navigationFile` are the per-project JSON file under Application Support; `SourcesListSettingsStore` is a smaller copy of that shape.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | **Done in S10-22 — no work.** The Sources time is read from the display cache (`Graph.Activity`) and carried on `Source.Updated`, `updated_at`, and `CatalogSource.updated`. The earlier D-1 to D-4 (an `audit.LatestOf` read, `List` ordering, `CreateSource`, the proto field) are superseded. | S10-22 D-20 and its Landed note. | Re-adding a scope read in `sources` — S10-22 removed the per-list history scan on purpose. |
| D-5 | `sourcegraphprogress` replaces `sqlCountSubjectsBySource` and `sqlCountSubjectsOne` with `sqlCountSubjectsByType` *(new)*: `SELECT s.source_id, t.key, t.origin, COUNT(*), COUNT(ic.subject_id) FROM subjects s JOIN subject_types t ON t.id = s.subject_type_id LEFT JOIN identity_claims ic ON ic.subject_id = s.id AND ic.status = 'accepted' GROUP BY s.source_id, t.key, t.origin`, and `sqlCountSubjectsByTypeOne` *(new)*, the same with `WHERE s.source_id = ?` and without `s.source_id` in the select and group. One scan loop per row: `SubjectCount += n` when `key != "source"`; `PrimaryCount += n` and `FiledCount += filed` when `promote.PrimaryKind(key, origin)`. `Progress` gets `PrimaryCount int32` and `FiledCount int32` *(new)*. `canvasTypeSQL` is deleted. The loop creates a Source's map entry only for a row with `key != "source"`, so a Source whose only subject is the `source` reification still has no subject entry. The observation queries and `scanGrouped` for them stay; each query's rows are closed before the next runs. `List` still omits Sources with zero subjects and zero observations (a primary is a subject, so no new rule). | One read of `subjects` per call, two fixed queries for the list (working rule; AC-2.5). `promote.PrimaryKind` owns which kinds are primary, so the list and Promote cannot disagree. The unique accepted-claim index makes `COUNT(ic.subject_id)` the filed count. | A separate primaries query beside the subject count — two reads of the same join. A `t.key IN ('person','event','place')` literal — a second list of primaries. |
| D-6 | `engine.proto` `SourceGraphProgress` gets `int32 primary_count = 5;` and `int32 filed_count = 6;` *(new)*; field 4 stays reserved. `protoProgress` maps both. | Additive; same message the list already loads and patches. | A new FFI method — a second load and cache key. |
| D-7 | Swift `SourceGraphProgress` gets `var primaryCount: Int = 0` and `var filedCount: Int = 0` *(new)*; `isZero` is unchanged. `GoStore.mapSourceGraphProgress` sets `primaryCount: Int(row.primaryCount)` and `filedCount: Int(row.filedCount)`. `SourcesModel.EvidenceStatus` *(new)* is nested beside `Sort`: `enum EvidenceStatus: String, CaseIterable, Identifiable, Codable { case empty, inProgress = "in_progress", filed }`, with `var label: LocalizedStringResource` and `init(_ progress: SourceGraphProgress?)`: nil or `primaryCount == 0` → `.empty`; `filedCount < primaryCount` → `.inProgress`; else `.filed`. | The rule lives once, beside the sibling list option and its copy. `GenealogyStore.swift` keeps only the counts. A Source missing from the map has zero counts (Go omits it). | A status string from Go — counts are the data, the label is UI. The enum in `GenealogyStore.swift` with its label elsewhere. |
| D-8 | `WorkspaceSession.graphProgressSource(of:)` also returns the id for `.promotedSubject(let id)`. | AC-2.3: `PromoteModel.file()` dispatches `.promotedSubject` with no effects, so nothing refreshes that Source's counts today. | Invalidating the whole progress map — the registry's `sourceGraphProgress` spec patches one Source. Making Promote return effects — an S10-22 follow-up, wider than this PR. |
| D-9 | `SourcesSplitRow` reads `source.updated.at`. It gets `private var updatedAt: String { source.updated.at }` *(new)*, read by `pageZone`. Its init does not change, so its two call sites in `SourcesListView.swift` do not change. | `CatalogSource.updated` exists (S10-22). | A new field on the row's init. |
| D-10 | `enum PVLastUpdated` *(new)*, a recipe in `macos/App/DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift` *(new)* with a `#Preview` of a known and an unknown cell, holds the composition S10-22 left in `ConclusionListRow`, moved, not copied: `static let columnWidth: CGFloat = 136` (from `updatedColumnWidth`), `static func detail(_ updatedAt: String, in context: TimestampFormat.Context) -> PVListDetail` (body of `updatedDetail`), and `static func label(_ base: String, updatedAt: String, in context: TimestampFormat.Context) -> String` (body of `labelWithUpdated`, joining with `L10n.LastUpdated.a11yJoin` *(new)* instead of `L10n.Conclusions.a11yList`; same "%1$@, %2$@" output). `ConclusionListRow` deletes `updatedColumnWidth`, `updatedDetail`, `labelWithUpdated`; `updatedColumn` stays and returns `PVListDetailColumn(width: PVLastUpdated.columnWidth) { PVLastUpdated.detail(Kind.updatedAt($0), in: .current) }`. `ConclusionListPage` calls `PVLastUpdated.columnWidth` (skeleton) and `PVLastUpdated.label` (row label). `PersonsListTests` `rowLabelEndsWithUpdated` and `updatedDetailComposesTheCell` move to `macos/ProvenenciaTests/PVLastUpdatedTests.swift` *(new)*, calling `PVLastUpdated` with the same expectations. `PVList`'s private `detailCell(_:)` becomes `struct PVListDetailCell: View` *(new)* in `PVList.swift` with `let detail: PVListDetail?` and `let width: CGFloat`; its body is the old cell plus the frame `PVList.row(_:index:)` applies today (`.frame(width: width, alignment: .leading).layoutPriority(1)`), so `PVList` draws `PVListDetailCell(detail: detail.content(item), width: detail.width)`. `SourcesSplitRow.pageZone` places `PVListDetailCell(detail: PVLastUpdated.detail(updatedAt, in: .current), width: PVLastUpdated.columnWidth)` between the title `VStack` and the ref `Text`, and the ref `Text` gets `.fixedSize(horizontal: true, vertical: false)` so only the title stack truncates. | S10-22 D-21 and the map's S10-41 In line: one feature-neutral home both rows call. It maps a product meaning (last activity) to `PVListDetail` props and now has a second call site, which is a recipe (`docs/design-system-layers.md` §2–3; `add-ui-component` step 3), beside `Recipes/Marks/PVMark.swift`. The cell view owns its column frame, so the width-and-priority rule lives once. | A home in `Platform/` — not a screen layer (`docs/macos-client-patterns.md`). The row re-spelling `PVList`'s frame. | Leaving it in `ConclusionListRow` — Sources would reach into `Features/Conclusions`. Copying it into `SourcesSplitRow`. Keeping `L10n.Conclusions.a11yList` — a Conclusions key in a shared home. Rebuilding the Sources row on `PVList` — the split zones are not a `PVList` row. |
| D-11 | `SourcesSplitRow.pageZone` `accessibilityValue` becomes `Text(verbatim: PVLastUpdated.label("\(source.title), \(source.ref)", updatedAt: updatedAt, in: .current))`. | SRL-2. | — |
| D-12 | `SourcesListSettings` *(new)*: `struct SourcesListSettings: Codable, Equatable { var sort: SourcesModel.Sort = .added; var typeFilterID = ""; var status: SourcesModel.EvidenceStatus? = nil; var query = "" }` with a custom `init(from:)` that decodes each field with `try?` and keeps its default on failure, in `macos/App/Features/Sources/SourcesListSettings.swift` *(new)*. `SourcesModel.Sort` gains `Codable`. `SourcesListSettingsStore` *(new)*, same file: `init?(projectUuid: String, userID: String, fileURL: URL? = nil, fileManager: FileManager = .default)` returns nil when `InstallPaths.projectKey(projectUuid:)` trimmed of `.whitespacesAndNewlines` is empty, or the trimmed `userID` is empty, or when the default URL cannot be built; `func load() -> SourcesListSettings`; `func save(_ settings: SourcesListSettings)`. The file is a JSON `[String: SourcesListSettings]` keyed by user id, at `InstallPaths.sourcesListSettingsFile(projectUuid:)` *(new)* = `{Application Support}/Provenencia/sources-list/{projectKey}.json`. `save` reads the file, replaces this user's entry, and writes with `Data.write(to:options: .atomic)` after creating the directory. Missing or undecodable file → defaults; a write error is ignored (the list keeps working in memory). `InstallPaths.projectKey(projectUuid:)` *(new)* is the hex step `navigationFileName` does today, extracted unchanged (trim `{}`, drop `-`, lowercase; it does not trim whitespace, so `navigationFileName` output does not change), and `navigationFileName` returns `"\(projectKey(projectUuid:)).json"`. The `InstallPaths.identityDirectory` doc comment adds `sources-list/` to the folders it lists. | Per-project state lives in JSON under Application Support keyed by project UUID, not `UserDefaults` (`navigation-history.md` decision; `add-workspace-location` Do-not list). Typed fields: one representation, field-by-field fallback (AC-4.3). | `UserDefaults` — the documented rule says no, and keys for deleted projects pile up. A field on `NavigationHistoryDocument` — settings are not history and would reset with it. A catalog table — not research data. A format version field — undecodable data already falls back. |
| D-13 | `WorkspaceNavigation.projectUuid` becomes `private(set) var`. `WorkspaceView.init` calls `navigation.attachProject(uuid:)` with `model.project?.uuid` when it is non-empty, **before** `navigation.onLocationCommit` is set and before `_navigation = State(initialValue:)`, so the history is loaded and `projectUuid` is set before the first body evaluation mounts `WorkspaceDestinationHost`, and the init attach starts no session loads. In `.onAppear`, the `attachProject` call is replaced by `session.apply(location: navigation.currentLocation)` as its first line, which warms the restored location at the same moment it is warmed today. The rest of `.onAppear` stays: `presentHistoryIssueIfNeeded()` surfaces an issue the init attach recorded. `.onChange(of: model.project?.uuid)` stays as is; the hook is wired by then, so a project switch still commits through it. `WorkspaceDestinationHost` passes `projectUuid: navigation.projectUuid` to `SourcesListView`, which builds `SourcesListSettingsStore(projectUuid:userID:)` and passes it to `SourcesModel.init(…, settingsStore: SourcesListSettingsStore? = nil)`. The model holds it as `private let settingsStore` and sets `settings = settingsStore?.load() ?? SourcesListSettings()` in `init`, before the first render. | Restored values are the initial values, so the anchor-clearing `.onChange` does not fire on restore (S10-40 anchor survives). Attaching in `init` also restores the history location before the first frame instead of one frame later. A repeated `WorkspaceView.init` (parent re-render: `model.error`, reduce motion, `model.project`) builds a throwaway navigation that reads the history file once and makes no store call, because its commit hook is not wired during the attach. It writes only when the file is missing or unusable, which the live attach has already fixed, and everything runs on the main actor, so it cannot race the live store. Wiring the hook first would send discarded catalog loads through the one held connection on every re-init. | Loading in `.task` — the first frame shows defaults and the change clears the restored anchor. Resolving the store lazily in `SourcesModel` once `projectUuid` is set — the settings arrive after the first render, the same anchor problem. Passing the uuid from `OnboardingView` into a new `WorkspaceNavigation` init parameter — a second attach path beside `attachProject`. |
| D-14 | `SourcesModel` replaces `var typeFilterID` and `var sort` with `var settings: SourcesListSettings { didSet { if settings != oldValue { settingsStore?.save(settings) } } }` *(new)*. Callers (`SourcesListView`, `SourcesModelTests`) read and bind `model.settings.sort`, `.typeFilterID`, `.status`, `.query`. `sortControlLabel` and `filterLabel` read `settings`. | One representation, one save, one change to observe. | Four stored properties with four `didSet`s and string mapping. |
| D-15 | `visibleSources` filters in order type → status → search, then sorts as today. Type uses `effectiveTypeFilterID` (D-17). Status: `EvidenceStatus(graphProgress(for: source.id)) == settings.status`. Search: `let q = settings.query.trimmingCharacters(in: .whitespacesAndNewlines)`; when non-empty, keep rows where `title.localizedStandardContains(q) \|\| ref.localizedStandardContains(q)`; the query is one phrase, not split into words. | AC-3.1 to AC-3.4: case- and accent-insensitive substring. | `lowercased().contains` (Properties) — misses `jose` → `José`. Word splitting — the brief says one phrase. |
| D-16 | `var isFiltering: Bool` *(new)*: `effectiveTypeFilterID` non-empty, or `settings.status` non-nil, or trimmed query non-empty. `countLine` uses the filtered form when `isFiltering` and `visible != total`; while `isAwaitingFilterData` or `statusCountsError != nil` it is the unfiltered `countLine(total:)`. `func clearFilters()` *(new)* sets `typeFilterID = ""`, `status = nil`, `query = ""` in one assignment to `settings`; sort unchanged. | SRL-6, SRL-7. | — |
| D-17 | `var effectiveTypeFilterID: String` *(new)*: `settings.typeFilterID` when it is in `types`, else `""`. `var isAwaitingFilterData: Bool` *(new)*: (`settings.status != nil` and `graphCountsLoading`) or (`settings.typeFilterID` non-empty and the types handle has no value and is not `.error`). `var statusCountsError: Error?` *(new)*: the progress handle's `error` when `settings.status != nil` and the handle status is `.error`, else nil. `filterMenu`'s selection reads `effectiveTypeFilterID` and writes `settings.typeFilterID`. The stored id is overwritten only when the researcher changes the type or presses Clear filters. | AC-4.3 and SRL-8 without a flash of the empty state and without wiping a saved type while types load. | Ignoring the type until types load — the full list flashes, then narrows. Clearing an unknown id on load — a write the researcher did not make. |
| D-18 | `SourcesListContent.listBody`, in order: `model.statusCountsError` → `PVCallout(tone: .danger, message: L10n.Errors.message(for: error))`, same padding and identifier pattern as `sources.loadError`, id `sources.statusError`; `model.isAwaitingFilterData` → the first-load `ProgressView`, no caption band; `visibleSources.isEmpty` → `PVEmptyState(icon: .filter, title: L10n.Sources.noFilterMatchesTitle, message: L10n.string(L10n.Sources.noFilterMatchesMessage), compact: true) { PVButton(L10n.Sources.clearFilters, variant: .secondary) { model.clearFilters() }.accessibilityIdentifier("sources.clearFilters") }`; else the list as today. The two `.onChange` handlers for `model.sort` and `model.typeFilterID` are replaced by one `.onChange(of: model.settings) { navigation.go(to: navigation.currentLocation.withListScrollAnchor(nil)) }`, which covers sort, type, status, search, and Clear filters. | SRL-7, SRL-8, SRL-10. | A separate “no search results” state — one state for every filter (brief). |
| D-19 | Toolbar order: `searchField`, `filterMenu`, `statusMenu`, `sortMenu`, `Spacer(minLength: 0)`. `searchField` *(new)* is `PVInput(text: $model.settings.query, prompt: L10n.Sources.searchPrompt, icon: .search)` at the default `.md` size, `.frame(minWidth: 160, maxWidth: 280)`, identifier `sources.search`. `statusMenu` *(new)* is a `PVSelect` built like `sortMenu`: `selection: Binding(get: { model.settings.status?.rawValue ?? "" }, set: { model.settings.status = SourcesModel.EvidenceStatus(rawValue: $0) })`, options `""` → `L10n.string(L10n.Sources.statusAll)` then each case → `L10n.string($0.label)`, `icon: .network` (the Evidence graph glyph; a `PVSelect` without an icon draws the field trigger, not the toolbar chip, `PVSelect.swift` `body`), `menuWidth: 220`, `fillsWidth: false`, `accessibilityLabel: L10n.Sources.statusMenu`, identifiers `sources.status`, `sources.status.all`, `sources.status.<rawValue>`. | Brief SRL-3, SRL-4; the `sortMenu` binding shape. | A segmented control or chips (brief non-goal). No icon — draws the field trigger beside two chips (board kit note). An iconless chip variant — a kit change for one call site. |
| D-20 | `FakeStore` gets `private static let primaryTypeKeys: Set<String> = ["person", "event", "place"]` *(new)*. `graphProgress(for:projectDir:)` sets `primaryCount` to the subjects whose type key is in it (key only; the fake has no origin) and `filedCount` to 0, unless `graphProgressBySource` overrides. The four older literal sites in `FakeStore` are not changed. | One copy at the new site. | Editing the four older sites — outside this PR. |
## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `Progress.PrimaryCount`, `FiledCount`, `sqlCountSubjectsByType`, `sqlCountSubjectsByTypeOne` *(new)* | fields / SQL | `sqlCountSubjectsBySource`, `sqlCountSubjectsOne` | Replace them: one read yields subject, primary, and filed counts. |
| `SourceGraphProgress.primary_count`, `filed_count` *(new)* | proto fields | `subject_count` | Same message. |
| `SourcesModel.EvidenceStatus` *(new)* | enum | `SourcesModel.Sort` | Sibling list option, same shape. |
| `PVListDetailCell` *(new)* | component view | `PVList.detailCell(_:)` (private) | Extracted so the Sources row draws the same cell. |
| `PVLastUpdated` *(new)*, `L10n.LastUpdated.a11yJoin` *(new)* | recipe / key | `ConclusionListRow.updatedColumnWidth`, `updatedDetail`, `labelWithUpdated`, `L10n.Conclusions.a11yList` | Moved from `ConclusionListRow` to a shared home (S10-22 D-21); the join gets a neutral key. |
| `SourcesListSettings`, `SourcesListSettingsStore`, `InstallPaths.sourcesListSettingsFile`, `InstallPaths.projectKey` *(new)* | struct / file / funcs | `NavigationHistoryStore`, `InstallPaths.navigationFile`, `navigationFileName` | Per-project JSON under Application Support, the documented home; `projectKey` is extracted from `navigationFileName`. |
| `SourcesModel.settings`, `isFiltering`, `clearFilters`, `effectiveTypeFilterID`, `isAwaitingFilterData`, `statusCountsError` *(new)* | properties / funcs | `typeFilterID`, `sort`, `graphCountsLoading` | Same model; `settings` replaces `typeFilterID` and `sort`. |
| `FakeStore.primaryTypeKeys` *(new)* | constant | four literal sites in `FakeStore` | One copy at the new site. |
| `L10n.Sources.searchPrompt`, `statusAll`, `statusEmpty`, `statusInProgress`, `statusFullyPromoted`, `statusMenu`, `clearFilters` *(new)* | keys | `filterAllTypes`, `filterMenu` | Same namespace. |

**Builds on.** S10-22's `CatalogSource.updated` and its patch, the `sourcegraphprogress` counts and their session cache key and patch path, `promote.PrimaryKind`, S10-22's cell composition and `PVList` detail cell, the history file's Application Support home and project key, `PVInput` / `PVSelect` / `PVEmptyState`, the S10-40 anchor clearing.
**Deletes.** `sourcegraphprogress.sqlCountSubjectsBySource`, `sqlCountSubjectsOne`, `canvasTypeSQL`; `SourcesModel.typeFilterID` and `sort` as stored properties; the two `.onChange` handlers for sort and type; `ConclusionListRow.updatedColumnWidth`, `updatedDetail`, `labelWithUpdated` (moved to `PVLastUpdated`); `PVList.detailCell(_:)` (now `PVListDetailCell`); the stale `sortUpdated` comment (“stand-in until audit timestamps surface”).
**Next adds.** Another list that shows a last-updated time calls `PVLastUpdated`. A status shown on the row reads `SourcesModel.EvidenceStatus`. A fifth Sources setting is one field on `SourcesListSettings`.

### Schema

None — no catalog change. Scopes are already stored.

### Go core

| Package | Change |
| --- | --- |
| `core/database/sourcegraphprogress` | D-5. Package comment: “aggregates Evidence-graph subject, observation, primary, and filed counts per Source for the Sources list”. |

### FFI

`api/proto/engine.proto`: `SourceGraphProgress.primary_count = 5`, `filed_count = 6` (D-6). Handler: `sourcegraphprogress.go` `protoProgress`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Last-updated cell | `PVListDetailCell` (extracted) | `macos/App/DesignSystem/Components/List/PVList.swift` | D-10. |
| Cell composition | `PVLastUpdated` *(new)* recipe | `macos/App/DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift`; callers `ConclusionListRow.swift`, `ConclusionListPage.swift`, `SourcesSplitRow.swift` | D-10. Register it in the App target and `PVLastUpdatedTests.swift` in the test target in `project.pbxproj`. |
| Split row | snowflake | `macos/App/Features/Sources/SourcesSplitRow.swift` | D-9, D-10, D-11. |
| Toolbar, list states | snowflake | `macos/App/Features/Sources/SourcesListView.swift` | D-18, D-19. |
| Model | — | `macos/App/Features/Sources/SourcesModel.swift` | D-13, D-14, D-15, D-16, D-17. |
| Settings | — | `macos/App/Features/Sources/SourcesListSettings.swift` *(new)*; `macos/App/Platform/InstallPaths.swift` | D-12. Add the new file to the App target and `macos/ProvenenciaTests/SourcesListSettingsTests.swift` to the test target in `project.pbxproj`, the four objects per file add-swift-test describes. |
| Host | — | `WorkspaceView.swift`, `WorkspaceDestinationHost.swift`, `WorkspaceNavigation.swift` | D-13. |

Frames, in words, from the agreed board (today 14 Oct 2026, Mountain Time; content host 1136px). **1 Times** — no filters, eight rows; in each filing zone the grid is thumb 44, title stack (flexible), the 136pt last-updated column (history icon 13pt, caption, muted), then the mono ref; “Today, 3:15 PM”, “Today, 9:02 AM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, and a faint “—” on the 1891 census, which has no revision; the 248pt Evidence graph zone with its count line is unchanged; count line “8 sources”. **2 Help** — hovering “Yesterday” shows “Last updated 13 Oct 2026 · 4:40 PM MDT” (native help on macOS). **3 Status menu** — the status chip (network icon) open over the list: All statuses (checked), Empty, In progress, Fully promoted. **4 Filtered** — In progress + Census schedule: three rows, graph zones show counts, “3 of 8 sources”. **5 Search** — `smith` matches two titles, “2 of 8 sources”. **6 Nothing matches** — Fully promoted + `zzz`: toolbar stays; compact `PVEmptyState` “No matching sources” / “Try another filter, or clear them to see every source.” with a secondary Clear filters button; count line “0 of 8 sources”. **7 Counting** — Empty set, counts not in: toolbar, then only the first-load spinner; no caption band, no empty state; count line “8 sources”. **7b Counts failed** — In progress set, counts unreadable: the page's load-error `PVCallout`. **8 Narrow** — 650px host: long title and type truncate; the 136pt column, the ref, and the 248pt graph zone hold; search is at its 160pt minimum; the three chips keep their size. Toolbar in every frame: search (280pt, shrinks to 160), type chip, status chip, sort chip, leading-aligned.

States: empty catalog → unchanged `sources.empty`; load error → unchanged; counts error with a status set → `sources.statusError` callout (D-18); 1000 Sources → filtering and sorting are in-memory on the loaded array, one cell build per visible row in a `LazyVStack`. Accessibility: D-11; the search field's prompt is its label; the status select's `accessibilityLabel` is `L10n.Sources.statusMenu`.

### Strings

L10n key *(new)* in `L10n.LastUpdated` (S10-22's namespace): `lastUpdated.a11y.join` = "%1$@, %2$@", comment “Joins a list row's spoken label and its last-updated phrase”; accessor `static func a11yJoin(_ base: String, _ updated: String) -> String`.

L10n keys *(new)* in `L10n.Sources`, each with a `comment`, added to `Localizable.xcstrings`:
- `sources.list.searchPrompt` = "Search title or ref"
- `sources.list.statusMenu` = "Evidence status" — accessibility label of the status select
- `sources.list.status.all` = "All statuses"
- `sources.list.status.empty` = "Empty"
- `sources.list.status.inProgress` = "In progress"
- `sources.list.status.fullyPromoted` = "Fully promoted" — every person, event, and place on the graph is promoted
- `sources.list.clearFilters` = "Clear filters"

`SourcesModel.EvidenceStatus.label` maps to the three status keys. The case stays `filed` (raw value `"filed"`, stored in settings); only its label reads “Fully promoted”.

Changed: `sources.list.noFilterMatchesTitle` → "No matching sources"; `sources.list.noFilterMatchesMessage` → "Try another filter, or clear them to see every source."; both comments say they cover type, status, and search. `sources.list.sort.updated` comment → “Sources sort option: most recent work first (audit revision)”.

## 7. Steps

**STEP-1 — Primary and filed counts**
Files: `core/database/sourcegraphprogress/sourcegraphprogress.go`, `core/database/sourcegraphprogress/sourcegraphprogress_test.go`, `api/proto/engine.proto`, generated files, `api/ffi/handlers/sourcegraphprogress.go`, `api/ffi/handlers/sourcegraphprogress_test.go`. Skill: `add-catalog-query`, `add-ffi-handler`.
Do: D-5, D-6. Run `./scripts/generate-proto.sh`. Add T-4, T-5.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/sourcegraphprogress/ ./api/...` passes; `git diff --stat` lists the regenerated Go and Swift protobuf files.

**STEP-2 — Swift counts, fakes, promote refresh**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift` (`mapSourceGraphProgress`), `FakeStore.swift`, `macos/App/Features/Workspace/Session/WorkspaceSession.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-swift-test`.
Do: D-7 counts on `SourceGraphProgress` and the mapper, D-8, D-20. Add T-7.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/WorkspaceSessionTests -only-testing:ProvenenciaTests/SourcesModelTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-3 — Model: status, search, saved settings**
Files: `macos/App/Features/Sources/SourcesListSettings.swift` *(new)*, `macos/App/Platform/InstallPaths.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `SourcesModel.swift`, `SourcesListView.swift` (init, and `model.sort` / `model.typeFilterID` reads moved to `model.settings`), `macos/App/Features/Workspace/WorkspaceView.swift`, `macos/App/Features/Workspace/WorkspaceDestinationHost.swift`, `WorkspaceNavigation.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/ProvenenciaTests/SourcesListSettingsTests.swift` *(new)*, `SourcesModelTests.swift`, `macos/ProvenenciaTests/WorkspaceNavigationTests.swift`. Skill: `add-workspace-place`, `add-swift-test`, `add-localized-string`.
Do: D-7 `EvidenceStatus` and its three label keys (`statusEmpty`, `statusInProgress`, `statusFullyPromoted`, §6 Strings), D-12, D-13, D-14, D-15, D-16, D-17. Update `SourcesModelTests` for `settings`. Add T-6, T-8, T-9, T-10, T-11, T-12, T-14.
Check: the STEP-2 `xcodebuild` command plus `-only-testing:ProvenenciaTests/SourcesListSettingsTests -only-testing:ProvenenciaTests/WorkspaceNavigationTests` passes.

**STEP-4 — Shared cell and the Sources column**
Files: `macos/App/DesignSystem/Recipes/LastUpdated/PVLastUpdated.swift` *(new)*, `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/Platform/L10n.swift` (`a11yJoin`), `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Conclusions/ConclusionListRow.swift`, `ConclusionListPage.swift`, `macos/App/Features/Sources/SourcesSplitRow.swift`, `macos/Provenencia.xcodeproj/project.pbxproj`, `macos/ProvenenciaTests/PersonsListTests.swift`, `macos/ProvenenciaTests/PVLastUpdatedTests.swift` *(new)*. Skill: `add-ui-component`, `add-localized-string`.
Do: D-9, D-10, D-11. `PVList` output and the Conclusion lists' spoken labels do not change. Add T-13. Narrow the window until the title truncates: the column, the ref, and the 248pt graph zone keep their width.
Check: the STEP-2 `xcodebuild` command with `-only-testing:ProvenenciaTests/PVListTests -only-testing:ProvenenciaTests/PersonsListTests -only-testing:ProvenenciaTests/PVLastUpdatedTests` passes (`PVListTests` unchanged, guarding the extraction); `rg -n "updatedDetail|labelWithUpdated|updatedColumnWidth" macos/App` prints nothing; Times frame visible in the app.

**STEP-5 — Toolbar, empty and counting states, copy**
Files: `macos/App/Features/Sources/SourcesListView.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`. Skill: `add-localized-string`, `add-ui-component`.
Do: D-18, D-19, and the rest of §6 Strings for `L10n.Sources` (`searchPrompt`, `statusMenu`, `statusAll`, `clearFilters`, and the changed `noFilterMatches` copy and `sortUpdated` comment).
Check: `python3 scripts/check-localizable-xcstrings.py` passes; manual check steps 2–6.

**STEP-6 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `deployment-plan.md` (S10-41 row and slice 6 line), this plan (status, Landed note), the S10-D8 brief → `design/archive/`, `design/README.md`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py docs/deployment-plan/spike-10/prs/S10-41-sources-list.md` reports no errors; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Show when each Source last changed on the Sources list."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-4 | Go | `core/database/sourcegraphprogress/sourcegraphprogress_test.go` `TestSourceGraphProgress` (extend, table) | no subjects → absent / zeros; one person, no claim → primary 1 filed 0; same person with accepted claim → 1/1; provisional claim → 1/0; participation bridge alone → primary 0; person + place, one accepted → 2/1; only a `source`-type subject → absent from `List`; `Get` equals the `List` row | AC-2.1, AC-2.2, AC-2.3, AC-2.4, D-5 |
| T-5 | Go | `api/ffi/handlers/sourcegraphprogress_test.go` `TestListSourceGraphProgress` (extend) | `primary_count` and `filed_count` on the wire | D-6 |
| T-6 | Swift | `macos/ProvenenciaTests/SourcesModelTests.swift` `evidenceStatusFromCounts` *(new)* | `EvidenceStatus(nil)` → empty; counts (0,0) → empty; (2,1) → inProgress; (2,2) → filed | D-7 |
| T-7 | Swift | `macos/ProvenenciaTests/WorkspaceSessionTests.swift` `applyPromotedSubjectPatchesProgress` *(new)* | warm progress; set fake override for s1 to filed; `apply(.promotedSubject(sourceId: "s1"))` → `getSourceGraphProgressCalls == 1` and the map's s1 row is filed | AC-2.3, D-8 |
| T-8 | Swift | `macos/ProvenenciaTests/SourcesListSettingsTests.swift` *(new)* `roundTrip`, `perProjectAndUser`, `badDataReadsDefaults`, `unknownFieldValuesFallBack`, `emptyKeysMakeNoStore`, `projectKeyMatchesNavigationFileName` | each test uses a temp directory `fileURL`; save then load equal; (p1,u1) vs (p1,u2) vs (p2,u1) independent; garbage file → defaults; `{"sort":"bogus","status":"x","query":"smith"}` → sort `.added`, status nil, query `smith`; `""`, `"  "`, `"{}"` uuid or empty user → `init?` returns nil; `navigationFileName(projectUuid:) == projectKey(projectUuid:) + ".json"` | AC-4.4, D-12 |
| T-9 | Swift | `SourcesModelTests.swift` `settingsRestoreOnNewModel` *(new)* | model A sets sort `.az`, type `t1`, status `.filed`, query `smith`; model B with a store on the same temp file, uuid, user → `settings` equal | AC-4.1, AC-4.2, D-13, D-14 |
| T-10 | Swift | `SourcesModelTests.swift` `searchMatchesTitleAndRefIgnoringCaseAndAccents` *(new)* | the AC-3.1, AC-3.2, AC-3.3, AC-3.4 table | AC-3.1, AC-3.2, AC-3.3, AC-3.4, D-15 |
| T-11 | Swift | `SourcesModelTests.swift` `filtersCombineAndClear` *(new)* | type + status + query → expected rows and “3 of 8 sources”; `clearFilters()` → all rows, sort unchanged; restored type id not in loaded types → all rows, `isFiltering == false`, `settings.typeFilterID` unchanged | AC-3.5, AC-3.6, AC-4.3, D-16, D-17 |
| T-12 | Swift | `SourcesModelTests.swift` `statusFilterWaitsForCounts` *(new)* | status `.empty` before progress loads → `isAwaitingFilterData == true`; after load → false and rows filtered; restored type before types load → true; while awaiting, `countLine` is the unfiltered total; `listSourceGraphProgressError` set and status `.filed` → `statusCountsError != nil` and `countLine` is the unfiltered total | AC-2.6, D-17 |
| T-13 | Swift | `macos/ProvenenciaTests/PVLastUpdatedTests.swift` *(new)*: `rowLabelEndsWithUpdated` and `updatedDetailComposesTheCell` moved from `PersonsListTests` (same expectations, calling `PVLastUpdated`), plus `sourceRowLabelEndsWithUpdated` *(new)* | `PVLastUpdated.label("1851 census, SRC-7KQ2D", updatedAt: yesterday, in: edmonton)` → `"1851 census, SRC-7KQ2D, updated Yesterday"`; empty → ends `", last updated unknown"` | AC-1.1, AC-1.2, AC-1.4, D-10, D-11 |
| T-14 | Swift | `macos/ProvenenciaTests/WorkspaceNavigationTests.swift` `attachProjectSetsUuidAndCommitsSynchronously` *(new)* | `let (seed, url) = try attachedNavigation()`; `seed.go(to: .sectionRoot(.sources).withListScrollAnchor("s3"))`; a new `WorkspaceNavigation` with `onLocationCommit` appending to `recorded`; `attachProject(uuid: projectUuid, fileURL: url)` → on return `projectUuid` equals it, `currentLocation == .sectionRoot(.sources)`, `currentLocation.listScrollAnchor == "s3"`, `recorded.count == 1`, `recorded[0].listScrollAnchor == "s3"`; a second `attachProject` with the same uuid → `recorded.count == 1`. This pins the navigation contract D-13 relies on; the `WorkspaceView.init` wiring is checked by manual check step 7 (relaunch) | AC-4.2, D-13 |

Must stay green: everything `scripts/check-pr.sh` runs. T-1, T-2, T-3 were dropped at the 2026-10-15 refresh: S10-22 landed and tested the Source time. Most likely to break: `SourcesModelTests.sortOrdersByAddedUpdatedAndTitle`, `PersonsListTests`, `PVListTests`, `WorkspaceSessionTests`.

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
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

Use a project with at least eight Sources of two types: one with no subjects, one with an unpromoted person, one whose person is promoted, one with a note edited today.

1. Open Sources. Each row shows the history icon and a time; the edited Source reads “Today,” and its time; hover shows “Last updated” and the full stamp (AC-1.1). A Source imported before scopes existed, if any, shows a faint “—”; VoiceOver on its filing zone ends “last updated unknown” (AC-1.2).
2. Status → Empty: only the subject-less Source (AC-2.1). In progress: the unpromoted one (AC-2.2). Fully promoted: the promoted one.
3. Open the In progress Source's Promote page, file the person, go back to Sources (status still In progress): the Source is gone from the list; switch to Fully promoted: it is there.
4. Type part of a title, then part of a ref (AC-3.3), then a title without its accent (AC-3.2): rows narrow as you type; the count line reads “N of M sources”.
5. Type `zzz`: the empty state shows Clear filters; press it: every row is back, the sort is unchanged.
6. Scroll down, change status, then scroll down and type in search: each time the list jumps to the top (AC-4.5).
7. Set sort A–Z, a type, Fully promoted, and a query. Open a Source, press Back: all four are set. Quit and relaunch: all four are set.
8. Edit a Source's note, open People, return to Sources: that row's time moved (AC-1.3). People shows no search, status, or type controls.
9. In the diff, `sourcegraphprogress.List` runs a fixed set of grouped queries with no loop over Sources (AC-2.5).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Proto regeneration, Xcode tests, the app build, and the manual check need macOS with `protoc`, `protoc-gen-go`, `protoc-gen-swift`, and Xcode | Run the PR on macOS. Never hand-edit `engine.pb.go` or `engine.pb.swift`. |
| A query while another's rows are open deadlocks on the one catalog connection | D-5 closes each grouped query's rows before the next; T-4 hangs if it does not. |
| A persisted query hides rows after relaunch and looks like data loss | Count line “N of M sources” and the Clear filters empty state (SRL-6, SRL-7). |
| Promote via a path that neither dispatches `.promotedSubject` nor returns `CatalogWriteEffects` | Manual check step 3; STEP-2 greps `promotedSubject` dispatch sites. |

**Stop and report if:**

- A §4 row is no longer true at your base (S10-22 code moved again), or `PVList.detailCell` cannot become `PVListDetailCell` without changing `PVList` output.
- `./scripts/generate-proto.sh` or `xcodebuild` cannot run on this machine. Stop after the `.proto` edit and report; do not hand-write generated code or claim a check you did not run.
- The agreed S10-D8 board differs from SRL-1 to SRL-10 or the copy in §6 Strings (the plan is refreshed first).
- `promote.PrimaryKind` is no longer exported, or importing `promote` from `sourcegraphprogress` creates an import cycle.
- A promote path in `macos/App` dispatches neither `.promotedSubject` nor `apply(_ effects:)` with the Source.
- `model.project?.uuid` is empty in `WorkspaceView.init` on the relaunch or open path even though `projectInfo` succeeded (D-13 relies on `OnboardingModel` setting `project` before `enterHome`).
- `navigation.projectUuid` read in `WorkspaceDestinationHost` is empty on the first `.sourcesList` build after the D-13 `init` attach. Check with a temporary `print` in `WorkspaceDestinationHost` on both the relaunch and the open path; do not commit it.

## 12. Open questions

None.

## 13. PR

**Title:** S10-41 — Sources list: last updated, status filter, search, saved settings
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-41-sources-list.md`.
