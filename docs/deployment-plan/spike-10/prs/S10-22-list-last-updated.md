# S10-22 — Last updated on People, Events, and Places

**Status:** In progress
**Spike / slice:** Spike 10 · Slice 6 — Lists and the Person page
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-22--last-updated-on-people-events-and-places)
**Brief:** [S10-D9](../design/archive/S10-D9-persons-list.md) (PSL-1…PSL-6), [S10-D10](../design/archive/S10-D10-events-list.md) (EVL-1…EVL-4), [S10-D11](../design/archive/S10-D11-places-list.md) (PLL-1…PLL-4); boards agreed 2026-10-15
**Board:** Claude Design projects *Persons list* (id 5c00cb55-8261-4579-b0e6-bc571ab23f78, file *Persons List.dc.html*, frames 01–07; 08–09 are S10-35 and S10-39), *Events list* (id 2799d7a0-387e-4570-b138-fc387562ac91, file *Events List.dc.html*, the frames marked S10-D10), and *Places list* (id e8910bb1-f564-4d7a-b829-69de0e60af5a, file *Places List.dc.html*, frames S10-D11 A–D). §6 macOS describes every frame this PR ships in words. The implementer does not need the boards.
**Depends on:** S10-21 (landed [#398](https://github.com/mendahu/provenencia/pull/398)), S10-D9, S10-D10, S10-D11
**Blocks:** —
**Branch:** `s10-22-list-last-updated`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-audit-scope`](../../../../.agents/skills/add-audit-scope/SKILL.md), [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [`core/database/effects/README.md`](../../../../core/database/effects/README.md), [S10-21 plan](S10-21-handle-last-activity.md), working rules "**Last updated** is an audit scope, not a timestamp column" and "**The Sources list is one page.**" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-15 against `main` @ `360b405f` (boards folded in 2026-10-15; rework amended 2026-10-15 against branch head `97a208b1`)
**Review:** [`S10-22-list-last-updated.review.md`](S10-22-list-last-updated.review.md) — initial plan 2 rounds; boards refresh; rework 2 rounds, no open blocker

## Rework (2026-10-15)

The user reviewed round 0 and asked for two refactors on this PR. They supersede the parts of the plan named here; the rest of the plan stands and is already built (STEP-1…STEP-8).

1. **One timestamp formatter.** `LastUpdatedDisplay` was a second display path for timestamps beside `TimestampFormat`, shaped for one screen. Every catalog timestamp is now formatted by one function that takes the timestamp, a style, and a context (locale, calendar, now) and returns one line of text. It carries no styling. The list cell is composed by the row from that function. D-5, D-6, D-8, D-10, D-11 are rewritten; D-12 is new.
2. **Surgical cache updates.** A write carries its activity (revision and time). The app patches the cached rows the write moved, in People, Events, Places, and Sources, through one declared patch in the session cache; the engine's display cache patches its own last-activity map on commit, so no list load scans the audit history. Lists not on screen are marked stale instead of reloaded. The user chose (2026-10-15) to move the Sources list onto the same patch and engine cache in this PR. D-3 and D-4 are rewritten; D-13…D-21 are new. This also retires PRF-2 (scopes resolved twice per write).

3. **Source-page and graph writes report their effects** (user, 2026-10-15, after the rework review). The rework review's performance lens found that a note, metadata, artifact, credibility, or Subject write on a Source still made People, Events, Places, and the Sources list reload in full on the next visit, because those writes report no effects and ride blanket mutation triggers. They now return `WriteEffects` like citation writes, and the blanket triggers go. D-22…D-24 are new; D-5's stamp, D-14, and D-15 are adjusted by the rework review (ship log, Fix plan R1).

4. **Activity is not evidence** (rework review round 2, 2026-10-15). `effects.sources` meant both "this Source's last activity moved" and "this Source's evidence graph changed", so once Source-page writes reported effects (item 3), a note save reloaded the Source graph, citation reads, and the Properties workspace, in the engine and the app. D-25 splits it the way D-13 already split handles; D-26 keeps one `apply` body. Both are built from the ship log's Fix plan R2.

New work is STEP-9…STEP-15, then STEP-16…STEP-18. The **Landed** note is written when the PR merges.

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

Each row on the People, Events, and Places lists shows when the researcher last worked on that handle, in its own column before the ref: “Today, 3:15 PM”, “Yesterday”, or a short date, with the full stamp as help and “—” when nothing is recorded. The value is the `canonical_entity` revision time S10-21 already returns. A note or Observation save moves the cell by patching the cached rows in place; it does not reload the lists. Every catalog timestamp in the app is formatted by one function with named styles.

**Why now.** S10-21 landed the engine fields (`updated_at` on `PersonHeader`, `EventHeader`, `PlaceHeader`). The rows do not show them, and the app drops them in `GoStore`.

**What stays.** List order (Go sorts by title or date), the title, the secondary lines, the thumb, the ref, the chevron, the empty state, and the count. No sort, filter, or search on these lists. The Sources list's chrome is untouched (S10-41); its cache and engine read change (D-20). The Person, Event, and Place detail pages do not show the time.

## 2. User stories

**US-1.** As a researcher, on People, Events, and Places, I see when I last worked on each row, so I can find what I was doing.

Times below use `now` = 2026-10-14 18:00 in `America/Edmonton`, locale `en_US`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Row `updatedAt` = `2026-10-14T21:15:00Z` (15:15 local) | Cell reads `Today, 3:15 PM`. Help is `L10n.LastUpdated.help(stamp:)` of `TimestampFormat.string(updatedAt, .stamp, in:)`: `Last updated Oct 14, 2026 · 3:15 PM MDT` (`en_US`; the date part follows the locale, `14 Oct 2026` in `en_GB`). |
| AC-1.2 | `updatedAt` = `2026-10-13T22:40:00Z` (16:40 local, yesterday) | Cell reads `Yesterday`. |
| AC-1.3 | `updatedAt` = `2026-10-12T15:00:00Z` | Cell reads `Oct 12, 2026` (`en_US`); with locale `en_GB` it reads `12 Oct 2026`. |
| AC-1.4 | `updatedAt` = `""` (revision 0) | Cell reads `—` in the faint color, with the history icon, no help. |
| AC-1.5 | `updatedAt` = `"garbage"` | Cell reads `—` in the faint color, with the history icon, no help (unparseable is unknown). |
| AC-1.6 | Persons row *Robins, James* with life line `b. 1817`, `updatedAt` yesterday | Spoken label is `Robins, James, PER-7KD45, b. 1817, updated Yesterday`. With `updatedAt` empty it ends `, last updated unknown`. |
| AC-1.7 | Events and Places rows | The same column, text, help, and em dash as Persons (EVL-1, PLL-1). |
| AC-1.8 | First load | The skeleton shows a column bar; the row does not shift when times arrive (PSL-6). |

**US-2.** As a researcher, when I edit evidence on a filed Subject, the list row's time moves without reopening the project.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Save only a note on an Observation of a Subject filed to Person *Bob* | The write reports Bob's handle in `WriteEffects.handle_scopes`. `handles` and `header_handles` are empty. |
| AC-2.2 | `CatalogWriteEffects(handleScopes: ["e1"])` | `invalidations(for:)` is empty. |
| AC-2.3 | `CatalogWriteEffects(handles: ["e1"], handleScopes: ["e1", "e2"])` | The three list keys once each, and `conclusionDetail` for `e1` only (from `handles`, as before S10-22). |
| AC-2.4 | In the app: edit a note on Bob's Observation, then open People | Bob's cell reads `Today, {now}`. |
| AC-2.5 | The People list is cached with *e1* (`updated` revision 10) and *e2* (revision 9). Apply `CatalogWriteEffects(handleScopes: ["e1"], activity: CatalogActivity(revision: 12, at: "2026-10-15T16:00:00Z"))` | *e1* is `{12, 2026-10-15T16:00:00Z}`, *e2* is unchanged, no list load starts (`isFetching == false`). Applying revision 11 changes nothing (older). Revision 0 changes nothing. |
| AC-2.5b | The Sources list is cached with *s1* (revision 5). Apply `CatalogWriteEffects(sources: ["s1"], activity: {12, T})` | *s1* is `{12, T}` and moves to the top of the Updated sort; no Sources load starts; the Sources list is not marked stale. |
| AC-2.5c | A list load is in flight when the patch applies | The in-flight load is cancelled and restarted, so an older response cannot replace the patched value (`setQueryValue`'s rule). |
| AC-2.6 | The three lists are cached and none is on screen. Apply `CatalogWriteEffects(handles: ["e1"])` | No list load starts (`isFetching == false`, old value). The next `query` of a list reloads it. |
| AC-2.7 | The People list is on screen. Apply `CatalogWriteEffects(handles: ["e1"])` | People reloads now; Events and Places (hidden) are marked stale. |
| AC-2.8 | A write commits at revision 42, time `T`, moving handle *H* | `writes.Result` has `Revision 42`, `At T`, equal to that `audit_transactions` row. `WriteEffects.revision` is 42 and `revision_at` is `T`. |
| AC-2.9 | The engine has loaded last activity once (`ListPersons`, or `sources.List` for Sources). A note save on *H* commits | The next `ListPersons` returns *H*'s `Updated` equal to that commit's revision and time, and runs no `audit_transaction_scopes` query; the same holds for a Source and `sources.List` / `AttachLatestRevision`. A commit older than the cached revision does not move it. `Drop` clears the map; the next read loads it with one query. |
| AC-2.10 | A write | `audit_transaction_scopes` rows equal `effects.Resolve`'s `Sources` and `HandleScopes` for that revision, and `audit` calls no resolver: `rg "effects\.(Resolve|Handles)" core/database/audit --glob '!*_test.go'` finds nothing, so each scope path is walked once per write, in `writes.Run`. |
| AC-2.11 | Save a note on Source *S* (no filed Subject involved) with People, Events, Places, and Sources cached and hidden | The response carries `WriteEffects` with `sources == [S]` and a revision. Sources row *S* is patched in place; the Sources list is not stale. People, Events, and Places are neither reloaded nor stale. |
| AC-2.12 | Change *S*'s credibility grade while *S* has a Subject filed to Person *P* | The response's `handles` include *P*; People, Events, and Places hidden are marked stale (D-19), and *P*'s detail page is invalidated. |
| AC-2.13 | Delete a Subject filed to Person *P* from the graph | The response's effects include *P* in `handles` and *S* in `sources`; lists behave as AC-2.12; Sources row *S* is patched. |
| AC-2.14 | The registry after this PR | `conclusionTriggers` contains neither `.mutatedSourceWorkspace` nor `.deletedSubject`; `sourcesList` has no `markStaleOn`; `CatalogMutation.savedCitation` does not exist. |

**US-3.** As a contributor, every catalog timestamp in the app is formatted by one function with named styles, so a new screen asks for a style instead of adding a formatter.

Context: `en_US` unless noted, a Gregorian calendar in `America/Edmonton`, `now` = 2026-10-14 18:00 local. `T` = `2026-10-12T15:00:00Z`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-3.1 | `TimestampFormat.string(T, .shortDate, in:)` | `Oct 12, 2026`; `en_GB`: `12 Oct 2026`. |
| AC-3.2 | `.longDate` | `October 12, 2026`. |
| AC-3.3 | `.dateTime` | `Oct 12, 2026 at 9:00 AM` (the locale's own joiner). |
| AC-3.4 | `.stamp` | `Oct 12, 2026 · 9:00 AM MDT` (`en_US`; `en_GB` `12 Oct 2026 · 09:00 GMT-6`); `·` comes from L10n. A single-digit day is not padded: `2026-10-04T15:00:00Z` → `Oct 4, 2026 · 9:00 AM MDT`, matching the cell and the board's `3 Mar 2025 · 8:47 PM MST`. |
| AC-3.5 | `.relative` for today 15:15, yesterday 16:40, and `T` | `Today, 3:15 PM`; `Yesterday`; `Oct 12, 2026`. |
| AC-3.6 | `.spoken` for the same three | `Today at 3:15 PM`; `Yesterday`; `October 12, 2026`. |
| AC-3.7 | Any style with `""` or `"garbage"` | `nil`. |
| AC-3.8 | The repo after this PR | `boardStamp`, `abbreviatedDateTime`, and `LastUpdatedDisplay` do not exist; no `ISO8601DateFormatter`, `DateFormatter`, or `Date.FormatStyle` formats a catalog timestamp outside `TimestampFormat.swift`. Not timestamps, and left alone: `PVDeleteImpact.swift` (list joining), `ReconciledValueDisplay.swift` (numbers), `Features/Dates/DateValueDisplay.swift` (record dates), `FakeStore` (fixtures). |

## 3. Scope

| In | Covered by |
| --- | --- |
| `effects.Set.HandleScopes` filled by `effects.Resolve` | STEP-1, AC-2.1 |
| `WriteEffects.handle_scopes` on the wire and mapped in Swift | STEP-2, STEP-3, AC-2.1 |
| `updatedAt` on `CatalogPersonHeader`, `CatalogEventHeader`, `CatalogPlaceHeader`, mapped from the list responses | STEP-4, AC-1.1 |
| Last-updated text, help, and em dash (PSL-2, PSL-3, PSL-4) | STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5 |
| `PVList` optional detail column and skeleton bar (PSL-1, PSL-5, PSL-6) | STEP-6, AC-1.8 |
| The column on all three lists and the spoken suffix (PSL-6; EVL-1 to EVL-4; PLL-1 to PLL-4) | STEP-7, STEP-13, AC-1.6, AC-1.7 |
| Rework: scope list declared once in `effects`; scopes resolved once per write | STEP-9, AC-2.10 |
| Rework: commit carries its time to listeners and the write result | STEP-9, AC-2.8 |
| Rework: engine last-activity cache patched on commit | STEP-10, AC-2.9 |
| Rework: `WriteEffects.revision` / `revision_at` | STEP-11, AC-2.8 |
| Rework: one timestamp formatter; callers moved; old formatters deleted | STEP-12, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.7, AC-3.8 |
| Rework: the row composes the cell from the formatter | STEP-13, AC-1.1, AC-1.4, AC-1.6 |
| Rework: one patch primitive; declared last-updated patches on People, Events, Places, and Sources; hidden lists marked stale | STEP-14, AC-2.2, AC-2.3, AC-2.5, AC-2.5b, AC-2.5c, AC-2.6, AC-2.7 |
| Rework: Sources list and `AttachLatestRevision` read the engine cache; `updated_at` on `Source` | STEP-10, STEP-11, AC-2.9 |
| Rework: Source-page and graph writes return `WriteEffects`; one apply for effects plus mutation; blanket triggers removed | STEP-16, STEP-17, AC-2.11, AC-2.12, AC-2.13, AC-2.14 |

| Out | Home |
| --- | --- |
| The Sources list cell, filters, search, persisted settings | S10-41 (S10-D8) |
| Likeness thumb on Persons (PSL-7) | S10-35 |
| Preferred portrait as the thumb (PSL-8) | S10-39 |
| Sort by last updated on People, Events, or Places | not planned (working rule "The Sources list is one page") |
| Last updated on the Person, Event, or Place detail page | not planned |
| A timer that re-renders “Today” into “Yesterday” while a list sits open | not planned (brief §2.1) |
| Fetching only the changed rows when a recompute changes a list row's title or secondary line | Follow-up idea: a visible list still reloads in full when a handle's values change (pre-existing since S9) |
| Writes that still report no `WriteEffects`: `CreateSource`, `UpdateSource`, `SetSourceCover`, `DeleteSource`, promote, Property writes | Follow-up idea: they keep their mutation kinds (`createdSource`, `patchedSource`, `deletedSource`, `promotedSubject`, `updatedProperty`); `SourcesModel.createSource` reloading the visible list before inserting the row it has is logged with them |
| Other hidden keys revalidated on every Observation save (`propertiesWorkspace`, `citationsByArtifact`) | Follow-up idea: they cannot simply go stale because `readyValue` returns `nil` for a stale key and `CitationComposerModel` relies on it |
| The Sources row's last-updated cell (S10-D8) | S10-41; it moves the cell composition (`updatedDetail`, `labelWithUpdated`, `updatedColumnWidth`) to one feature-neutral home both rows call |
| Fixture times in `FakeStore` or `PreviewFixtures` headers | not planned: previews and UI tests show the em dash |

## 4. Current state

The rows down to **Tests** describe the planning base `360b405f`; STEP-1…STEP-8 changed them as §5 and §6 say, and the branch head has that code. The rework rows (Audit scopes, Commit listeners, Display cache, List reads, Sources reads, Timestamp callers, App cache) describe the branch head.

| Area | Where | Today |
| --- | --- | --- |
| Effect set | `core/database/effects/registry.go` `Set` | Has `Handles`, `Sources`, `CacheSources`, `Search`, `Vocabulary`, `Structure`, `HeaderDeps`. No handle scopes. |
| Resolve | `core/database/effects/resolve.go` `Resolve` | Appends `eff.Source`, `eff.CacheSource`, `eff.Handles` per change via `appendPath`. Does not read `eff.HandleScope`. `HandleScopes` (same file) resolves that path on its own, for `audit.Record`. |
| Note path | `core/database/effects/registry.go` `"observation_notes"` | `Source` and `HandleScope` only, no `Handles`. So a note save reports no handles and the lists are not reloaded. |
| Write | `core/writes/run.go` `Run` | Records audit, then `effects.Resolve`, recomputes `set.Handles`, returns `Result{Effects: set}`. |
| Wire | `api/proto/engine.proto` `WriteEffects` (fields 1–6) | `handles`, `sources`, `cache_sources`, `vocabulary`, `structure`, `header_handles`. |
| Handler | `api/ffi/handlers/write_effects.go` `writeEffectsProto` | Maps `effects.Set` to `engine.WriteEffects`; `idStrings` turns ids into UUID strings. No test file. |
| Header fields | `api/proto/engine.proto` `PersonHeader` 6–7, `EventHeader` 15–16, `PlaceHeader` 12–13 | `updated_revision`, `updated_at` (RFC3339 UTC, empty when 0), filled on `List*Headers` only (S10-21). |
| Swift effects | `macos/App/Features/Workspace/Session/CatalogMutation.swift` `CatalogWriteEffects` | `handles`, `sources`, `cacheSources`, `vocabulary`, `structure`, `headerHandles`; `touchedHandles` = handles + headerHandles. |
| Swift mapping | `macos/App/Platform/GoStore.swift` `mapWriteEffects`, `mapPersonHeader`, `mapEventHeader`, `mapPlaceHeader` | Effects map fields 1–6. Header mappers drop `updatedAt`. |
| Invalidation | `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift` `invalidations(for:project:)` | Non-empty `touchedHandles` → `personsList`, `eventsList`, `placesList`, then `conclusionDetail` per handle. |
| Header models | `macos/App/Platform/GenealogyStore.swift` `CatalogPersonHeader`, `CatalogEventHeader`, `CatalogPlaceHeader` | No time field. |
| Timestamps | `macos/App/Platform/TimestampFormat.swift` `TimestampFormat` | `parse`, `boardStamp` (`04 Mar 2026 · 7:22 PM MST`), `abbreviatedDateTime`. No relative form. |
| List kit | `macos/App/DesignSystem/Components/List/PVList.swift` `PVList`, `PVListSkeleton` | Row: thumb, text stack, trailing `HStack` of mono `meta` and chevron. Skeleton draws a 72pt trailing bar. No extra column. |
| List page | `macos/App/Features/Conclusions/ConclusionListPage.swift` `ConclusionListKind`, `ConclusionListBody` | Builds `PVList` with `meta: Kind.ref`, `itemAccessibilityLabel: Kind.accessibilityLabel(_:orders:)`. |
| Kinds | `macos/App/Features/Conclusions/PersonsListView.swift` `PersonsList`, `EventsListView.swift` `EventsList`, `PlacesListView.swift` `PlacesList` | Conform to `ConclusionListKind`. |
| Row helpers | `macos/App/Features/Conclusions/ConclusionListRow.swift` `ConclusionListRow` | `thumbnail(mark:)`, `title(_:extraCount:)`, `accessibilityLabel(_:ref:)`. |
| Audit scopes | `core/database/audit/scopes.go` `ScopeSource`, `ScopeCanonicalEntity`, `scopeResolvers`, `Latest`, `LatestByScope`; `core/database/audit/audit.go` `Record` | `Record` walks `effects.Sources` and `effects.HandleScopes` itself, so `writes.Run` walks each scope path twice (here and in `Resolve`). `LatestByScope` groups every scope row of a type, joined to `audit_transactions`; its only non-test caller is `conclusionheaders.latestHandles`. `audit`'s internal test imports `core/database`, so a package `core/database` imports cannot import `audit`. |
| Commit listeners | `core/database/listen.go` `CommitListener`, `Notify`; `core/writes/run.go` `Result`, `Run` | `OnCommit(rev int64, set effects.Set)`; the commit time (`project.NowUTC()` passed to `Record`) is not kept. `Result{Revision, Effects}`. |
| Display cache | `core/database/graphcache/commit.go` `OnCommit`; `graph.go` `clear`, `Drop` | Patches handles, sources, labels on commit. Holds no last activity. |
| List reads | `core/database/conclusionheaders/compose.go` `latestHandles`, `ListPersons`, `ListEvents`, `ListPlaces` | Each list read runs `audit.LatestByScope` over the whole history, after `persons()` / `events()` / `places()` compose. Those three are shared with `PersonsFromQuerier` (search reprojection inside a write, over `graphcache.Over(q)`). `PersonsByIDs` and its siblings do not fill `Updated`. |
| Sources reads | `core/database/sources/sources.go` `sqlList`, `List`, `AttachLatestRevision`, `sqlLatestRevision`; `api/ffi/handlers/thumbnails.go` `enrichSourceProto`; `api/proto/engine.proto` `Source` (fields 1–12) | `sqlList` runs a correlated `MAX(t.revision)` over `audit_transaction_scopes` per Source; `AttachLatestRevision` runs one per Source read. `Source` carries `updated_revision = 12` only. Swift `CatalogSource.updatedRevision` (`GenealogyStore.swift`) feeds `SourcesModel`'s Updated sort; `WorkspaceSession.apply(_ effects:)` marks `sourcesList` stale on any touched Source, and `sourcesList`'s spec uses `markStaleOn`. |
| Timestamp callers | `macos/App/Features/Sources/SourcePageNotesView.swift` (`boardStamp`), `macos/App/Features/Onboarding/OnboardingProjectMetaLines.swift` `formatDate` (`abbreviatedDateTime`), `ConclusionListRow.updatedColumn` / `ConclusionListPage` (`LastUpdatedDisplay`) | Three formatting entry points with different inputs; only `LastUpdatedDisplay` takes locale and calendar. |
| App cache | `macos/App/Features/Workspace/Session/WorkspaceSession.swift` `handles` (`[CatalogQueryKey: AnyObject]`), `apply(_ effects:)`, `apply(invalidations:stale:)`, `setQueryValue`, `patchUpdatedSource`, `visibleKeys`, `warmQuery`; `CatalogQueryRegistry.swift` `Spec.evictWhenHidden`, `evictsWhenHidden(_:)` | Invalidated keys that are cached and not evicted are revalidated at once, on or off screen. Two patch paths with different in-flight rules: `setQueryValue` cancels and restarts an in-flight load; `patchUpdatedSource` calls `applySuccess` directly. Three per-kind switches already stay in step (`CatalogQueryRegistry.load`, `warmQuery`, `Kind.invalidation`). |
| Tests | `core/database/effects/handle_paths_test.go` `TestHandlePaths` "observation note reaches member handle"; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (the `invalidations(for: CatalogWriteEffects(…))` expectation); `macos/ProvenenciaTests/TimestampFormatTests.swift`; `macos/ProvenenciaTests/PersonsListTests.swift`; `macos/ProvenenciaTests/PVListTests.swift` | Cover the paths, invalidation, stamp, labels, and keyboard as today. |

**Precedent.** `header_handles` (field 6) was added to `WriteEffects` the same way: a field on `effects.Set`, one line in `writeEffectsProto`, one in `mapWriteEffects`, and a branch in `invalidations(for:)`. `header_handles` is filled from `Set.HeaderDeps`, which `writes.Run` computes after `Resolve`; `HandleScopes` instead comes out of `Resolve` (D-1). Copy the proto, mapper, and invalidation parts. The S10-21 plan added the header fields; this PR only reads them.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `effects.Set` gets `HandleScopes [][]byte` *(new)*, filled in `Resolve` with `appendPath(w, table, eff.HandleScope, set.HandleScopes)` beside `Handles`, with a field comment matching `Effect.HandleScope`'s. | The registry already owns the path; `Resolve` is "the whole effect of a batch". | Widening `observation_notes` `Handles` — that recomputes the reconciler for a note, which changes no value. Reading audit scopes back after commit — a second source of truth. |
| D-2 | `WriteEffects.handle_scopes = 7` *(new)*; Swift `CatalogWriteEffects.handleScopes` *(new)*. One name at every layer, Go's `HandleScopes`. | Separate from `handles`: a scope move needs a list refetch, not a detail refetch. | Folding into `handles` — would refetch every open detail page on each note save. |
| D-3 | `invalidations(for:)` reads `touchedHandles` only, as before S10-22: the three list keys and `conclusionDetail` per touched handle. `handleScopes` produce no invalidation; they move the time through D-17. `apply(_ effects:)` drops its hard-coded `stale: [.sourcesList]`; touched Sources move through D-17 too. `touchedHandles` does not change. | A scope move changes one field the write already reports (D-16). | Invalidating the lists on scope handles (round 0) — a full rebuild of three lists per note save, growing with the catalog. |
| D-4 | One activity value in Swift, mirroring the engine's `graphcache.Activity`: `struct CatalogActivity: Sendable, Equatable { var revision: Int64 = 0; var at: String = "" }` *(new, in `CatalogLastUpdated.swift`)*; revision 0 is "none". The three header models replace `updatedAt` with `var updated = CatalogActivity()` (last stored property), mapped from `updated_revision` and `updated_at`. `CatalogSource.updatedRevision` becomes `var updated = CatalogActivity()`, mapped from `updated_revision` and the new `updated_at`; `SourcesModel`'s Updated sort reads `updated.revision`. `ConclusionListKind.updatedAt(_:)` returns `row.updated.at`. | One fact, one shape, ordered by revision in both layers; S10-41 reads the same value for its cell and sort. | A time string alone (round 0) — the wall clock is not monotonic and Sources sort by revision. |
| D-5 | One formatter. In `macos/App/Platform/TimestampFormat.swift`, `enum TimestampFormat` keeps `parse` and gains `enum Style { case shortDate, longDate, dateTime, stamp, relative, spoken }` *(new)*, `struct Context { var locale: Locale; var calendar: Calendar; var now: Date; static var current: Context { Context(locale: .current, calendar: .current, now: Date()) } }` *(new; `calendar.timeZone` is the zone)*, `static func string(_ rfc3339: String, _ style: Style, in context: Context = .current) -> String?` *(new; `nil` when empty or unparseable)* and `static func string(_ date: Date, _ style: Style, in context: Context = .current) -> String` *(new)*. Every style formats through one `Date.FormatStyle(locale: context.locale, calendar: context.calendar, timeZone: context.calendar.timeZone)`, and each case formats only what it returns. `shortDate` = `.day().month(.abbreviated).year()`; `longDate` = `.day().month(.wide).year()`; `dateTime` = `Date.FormatStyle(date: .abbreviated, time: .shortened, locale:, calendar:, timeZone:)`; `stamp` = `L10n.Timestamp.stamp(date:time:)` of the `shortDate` date part (unpadded day, as the board draws `3 Mar 2025`) and `.hour().minute().timeZone(.specificName(.short))` (the locale orders the date: `Oct 12, 2026` in `en_US`); `relative` = `L10n.Timestamp.today(time:)` (`.hour().minute()`) when `context.calendar.isDate(date, inSameDayAs: now)`, `L10n.Timestamp.yesterday` when same day as `calendar.date(byAdding: .day, value: -1, to: now)`, else `shortDate`; `spoken` = the same tests with `L10n.Timestamp.todayAt(time:)`, `yesterday`, else `longDate`. Text only: no font, color, or view. `shortDate` and `longDate` stay public styles though no screen calls them yet: the user named them as part of the API. | The user's call: one function taking the timestamp, the locale, and the display option, returning one line of text. Every style is testable with a fixed context. | `LastUpdatedDisplay` (round 0) — a second, screen-shaped display path beside the formatter. One function per style — the API grows with every style. A `View` — styling is the caller's. |
| D-6 | The row composes the cell. `ConclusionListRow.updatedDetail(_ updatedAt: String, in context: TimestampFormat.Context) -> PVListDetail` *(new)*: `text` = `string(updatedAt, .relative, in:) ?? "—"`; `help` = `string(updatedAt, .stamp, in:).map(L10n.LastUpdated.help(stamp:))`; `icon` = `.history`; `faint` = the relative string is `nil`. `ConclusionListRow.labelWithUpdated(_ base: String, updatedAt: String, in context:) -> String` *(new signature)* = `L10n.Conclusions.a11yList(base, rest: string(updatedAt, .spoken, in:).map(L10n.LastUpdated.a11yUpdated) ?? L10n.LastUpdated.a11yUnknown)`. | The cell is this row's composition of shared pieces: formatter, L10n, kit slot. | Composition in `Platform/` — a feature's cell in the platform layer. |
| D-7 | `PVListDetailColumn<Item>` *(new)* `struct { var width: CGFloat; var content: (Item) -> PVListDetail? }` and `PVListDetail` *(new)* `struct { var text: String; var help: String?; var icon: PVSymbol?; var faint: Bool = false }`, both in `PVList.swift`. Both `PVList` inits (the primary one and the `Secondary == EmptyView` one, which forwards) get `detail: PVListDetailColumn<Item>? = nil`, placed right after `meta`. When set, every row draws, between the text stack and the trailing ref, a column of `detail.width` holding an `HStack(spacing: 6)` of `PVIcon(icon, size: 14)` (when `icon` is set) and a one-line, caption, non-mono text with `.fixedSize()` (it never truncates; the column frame is fixed and keeps `layoutPriority(1)` like the trailing group, so the title stack gives way), leading-aligned and vertically centered; icon and text are `PVColor.textMuted`, or `PVColor.textFaint` when `faint`; `content` returning `nil` draws an empty column of that width. Help covers the icon and the text and is attached only when non-nil (`if let help { cell.help(help) } else { cell }` in a `@ViewBuilder`, never `.help(help ?? "")`). No `.accessibilityHidden`: the row is already `.accessibilityElement(children: .ignore)`. `PVListSkeleton` gets `detailWidth: CGFloat? = nil` *(new)* and, when set, draws a 64 × 10 bar, leading-aligned, in that column before the existing 72pt ref bar, which stays. | The brief marks the list row Extend. One parameter carries closure and width, so a width with no column cannot be written, and `Components/` holds no board value or time type. | Separate `detail` and `detailWidth` parameters with a default width. A view laid out beside `meta` from `Features/` — the row layout is private to `PVList`. Putting the time in `meta` — the user chose its own column. |
| D-8 | `ConclusionListKind.updatedAt(_ row: Row) -> String` returns `row.updated.at`. `ConclusionListRow.updatedColumn<Kind: ConclusionListKind>(_: Kind.Type) -> PVListDetailColumn<Kind.Row>` (signature kept) returns width `ConclusionListRow.updatedColumnWidth` and content `{ updatedDetail(Kind.updatedAt($0), in: .current) }`. `ConclusionListBody` passes `detail: ConclusionListRow.updatedColumn(Kind.self)` and `itemAccessibilityLabel: { ConclusionListRow.labelWithUpdated(Kind.accessibilityLabel($0, orders: session.nameFormat), updatedAt: Kind.updatedAt($0), in: .current) }`. `ConclusionListPage.skeleton()` returns `PVListSkeleton(detailWidth: ConclusionListRow.updatedColumnWidth)`, at both skeleton sites. `updatedColumn`, `labelWithUpdated`, and `skeleton` exist on the branch; only their bodies or signatures change. The Kinds' own `accessibilityLabel` stays without the suffix. | One place for three lists. | Per-kind copies; an `accessibilityLabel(_ base: String, …)` overload beside `accessibilityLabel(_:ref:)`. |
| D-9 | `now` is `Date()` read when the row builds. No timer. | Brief §2.1. The list rebuilds on refetch and navigation. | `TimelineView` — an open list re-rendering every minute for a day boundary. |
| D-10 | The spoken suffix uses the `.spoken` style: `updated Yesterday`, `updated Today at 3:15 PM`, `updated October 12, 2026`; `last updated unknown` when there is no time. | VoiceOver reads a full month, not an abbreviation. | Reusing the visible `.relative` text. |
| D-11 | `ConclusionListRow.updatedColumnWidth: CGFloat = 136` *(new)*, the agreed boards' width, icon included. | One width for three lists. The width is layout, not formatting, so it is not in `TimestampFormat`. | A width per kind. A constant in `Platform/`. S10-41 moves it beside the cell it shares when the Sources row adopts the cell. |
| D-12 | Callers move to `TimestampFormat.string`: `SourcePageNotesView` uses `string(note.createdAt, .stamp) ?? note.createdAt`; `OnboardingProjectMetaLines.formatDate` uses `string(rfc3339, .dateTime) ?? rfc3339` (both keep today's raw fallback). `boardStamp`, `abbreviatedDateTime`, and `LastUpdatedDisplay` are deleted with their tests. New L10n namespace `L10n.Timestamp` *(new)*: `today(time:)` (key `timestamp.today`, "Today, %@"), `todayAt(time:)` (`timestamp.todayAt`, "Today at %@"), `yesterday` (`timestamp.yesterday`, "Yesterday"), `stamp(date:time:)` (`timestamp.stamp`, "%1$@ · %2$@"). `L10n.LastUpdated` keeps only `help(stamp:)`, `a11yUpdated(_:)`, `a11yUnknown`; keys `lastUpdated.today` and `lastUpdated.yesterday` are removed from `Localizable.xcstrings`. | Day words and the stamp joiner are date formatting, not last-updated copy. | Keeping `boardStamp` as a wrapper — two entry points. |
| D-13 | The scope list is declared once in `core/database/effects/scopes.go` *(new)*: `const ScopeSource = "source"`, `const ScopeCanonicalEntity = "canonical_entity"`, `type Scope struct { Type string; IDs func(Set) [][]byte }`, and `var Scopes = []Scope{{ScopeSource, func(s Set) [][]byte { return s.Sources }}, {ScopeCanonicalEntity, func(s Set) [][]byte { return s.HandleScopes }}}`. `writes.Run` calls `effects.Resolve` before `audit.Record`; `audit.Record(tx *sql.Tx, rev Revision, set effects.Set) (int64, error)` stores one scope row per id from each `effects.Scopes` entry's `IDs(set)` and no longer resolves; an empty `Set` stores no scope rows (its doc comment says so, replacing the "scopeResolvers entry resolves" sentence). Tests that call `Record` (`audit_test.go` ~18 calls, `scopes_internal_test.go`, `promote/bridges_test.go`) go through one helper per test package, `recordResolved(t, tx, rev)` *(new, test-only)*, which runs `effects.Resolve(tx, rev.Changes)` then `Record`. `audit.ScopeSource`, `audit.ScopeCanonicalEntity`, `scopeResolvers`, and `export_test.go`'s `ScopeResolvers` are deleted; callers use the `effects` constants. The standalone resolvers `effects.Sources` and `effects.HandleScopes` are deleted; `handle_paths_test.go` `mustHandleScopes` reads `effects.Resolve(tx, changes)` `.HandleScopes`. `effects.Handles` stays (`subjects.go` calls it). | One declared place maps a scope type to its ids; `audit` and the display cache both read it, and `effects` is the one package both may import. Each scope path is walked once per write (retires PRF-2). | Constants in `audit` — the display cache cannot import `audit` (`audit`'s internal test imports `core/database`, which imports `graphcache`). Keeping `Record` resolving — two walks per write. |
| D-14 | A commit carries its time. `core/database/effects/commit.go` *(new)*: `type Activity struct { Revision int64; At string }` (zero = no revision) and `type Commit struct { Activity; Effects Set }` (embedded, so `c.Revision` and `c.At` read as before). It lives in `effects` because `graphcache` must accept it in `OnCommit` and `core/database` imports `graphcache`, so a `database.Commit` would cycle. `database.CommitListener.OnCommit(c effects.Commit) error` and `Notify(commit effects.Commit)` (parameter named `commit`; the receiver is `c`) replace the `(rev, set)` forms; `graphcache.Graph.OnCommit`, its four calls in `graph_test.go`, and `writes/run_test.go` `recListener` follow. `core/writes/run.go`: `type Result = effects.Commit` (alias; `Revision` and `Effects` keep their names); `Run` keeps the `project.NowUTC()` it passes to `Record` as `At`. | The engine cache and the app need the time of the change; the write already has it. | `At` on `effects.Set` — the set says what changed, not when. `database.Commit` — import cycle. A second struct beside `Result` — one concept. |
| D-15 | The display cache holds last activity. `core/database/graphcache/activity.go` *(new)*: `func (g *Graph) Activity(scopeType string, id []byte) (effects.Activity, error)` returns zero and `nil` on a nil `*Graph`, an error (`graphcache: unknown scope type %q`) for a type not in `effects.Scopes` (including `""`), and otherwise loads that type's whole map once with the grouped SQL moved from `audit.LatestByScope` into `graphcache/sql.go`, then reads memory. `OnCommit` first, before the `Structure` early return, runs for each `effects.Scopes` entry whose map is loaded: for each id in `IDs(c.Effects)`, `if c.Revision > cur.Revision { set c.Activity }`. `clear()` empties the maps (so `Drop` and a structure change reload with one query). Readers: `conclusionheaders.ListPersons`, `ListEvents`, `ListPlaces` fill `Updated` from `c.Graph().Activity(effects.ScopeCanonicalEntity, id)` after composing, exactly where `latestHandles` is used today — never inside `persons()` / `events()` / `places()`, which search runs over a throwaway `graphcache.Over(q)` during writes. `latestHandles`, `audit.Latest`, `audit.LatestByScope` are deleted; the three `Updated` fields become `effects.Activity` (row types do not name the cache's package); `api/ffi/handlers/conclusion_headers.go` keeps reading `.Revision` and `.At`. In `audit/scopes_test.go`, `scopeSeed.handleRevision` reads with its own one-line SQL (`SELECT COALESCE(MAX(t.revision),0) …` for one `scope_id`). `conclusionheaders_test.go` `TestListHeadersCarryUpdated` keeps its assertions with the new type. Update the `compose.go` comment that said the display cache never holds it. Reads are serialized by `catalogsession.Do`, so the map needs no lock. | A list load costs the rows it returns, not the audit history. The commit knows which ids moved and when. | Grouped query per list load (round 0) — grows with history. A new listener package — `core/database` would import it, the same cycle. |
| D-16 | The write carries its activity to the app. `WriteEffects`: `int64 revision = 8; string revision_at = 9;` *(new)*. `writeEffectsProto(res writes.Result)` *(new signature)* sets them from `res.Revision` and `res.At`; its seven call sites (`citations.go` ×3, `observations.go` ×3, `connect.go`) pass `wrote` instead of `wrote.Effects`. Swift `CatalogWriteEffects` gains `var activity = CatalogActivity()`, mapped in `mapWriteEffects`. The `engine.proto` comment on `handle_scopes` and the Swift `handleScopes` comment say the app patches those rows; they no longer say it reloads lists. | The app patches rows with the write's own activity; no read. | A follow-up read of the moved rows — an FFI call per save. |
| D-17 | Declared last-updated patches. In `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` *(new)*: `protocol CatalogLastUpdatedRow { var id: String { get }; var updated: CatalogActivity { get set } }`, conformed by `CatalogPersonHeader`, `CatalogEventHeader`, `CatalogPlaceHeader`, `CatalogSource`; `enum CatalogLastUpdatedScope { case handle, source }`; `struct CatalogLastUpdatedPatch: Sendable { let scope: CatalogLastUpdatedScope; let apply: @MainActor @Sendable (WorkspaceSession, CatalogQueryKey, Set<String>, CatalogActivity) -> Void; static func rows<Row: CatalogLastUpdatedRow>(_: Row.Type, scope: CatalogLastUpdatedScope) -> Self }`. `rows` builds `apply` as `session.patchCached(key, as: [Row].self) { rows in` move each row whose `id` is in the set and whose `updated.revision < activity.revision` to `activity`; return whether any moved `}`. `CatalogQueryRegistry.Spec` gains `var lastUpdated: CatalogLastUpdatedPatch? = nil` and `func lastUpdatedPatch(_ key: CatalogQueryKey) -> CatalogLastUpdatedPatch?` *(new; takes a key like `hiddenPolicy(_:)`)*. Specs: `personsList`, `eventsList`, `placesList` → `.rows(CatalogPersonHeader.self / CatalogEventHeader.self / CatalogPlaceHeader.self, scope: .handle)`; `sourcesList` → `.rows(CatalogSource.self, scope: .source)`. `WorkspaceSession.apply(_ effects:)` first calls `patchLastUpdated(effects)` *(new, private)*: if `effects.activity.revision > 0`, for each cached key of this project whose spec declares a patch, it calls `patch.apply(self, key, ids, effects.activity)` with `ids` = `effects.handleScopes` for `.handle` and `effects.sources` for `.source`. No `switch key`. No loader runs. `sourcesList` keeps `markStaleOn` for mutation-path writes, which report no activity (Out). | Cache infrastructure: each list declares its patch in its registry row, one line per list. | A `switch key` mapping kinds to row types — a fourth per-kind switch to keep in step. Invalidate and refetch (round 0). Fetching the moved rows by id — an FFI call, and Go owns list order. |
| D-18 | One patch primitive. A private `replaceValue<Value>(_ key: CatalogQueryKey, _ value: Value, clearStale: Bool)` *(new)* holds today's `setQueryValue` body (bump the generation, cancel an in-flight load, `applySuccess`, restart the load if one was in flight); `setQueryValue` calls it with `clearStale: true` and keeps its behavior. `WorkspaceSession.patchCached<Value>(_ key: CatalogQueryKey, as: Value.Type, _ change: (inout Value) -> Bool)` *(new, internal so `CatalogLastUpdatedPatch` can call it)*: when the key has a value and `change` returns `true`, it calls `replaceValue(key, value, clearStale: false)`, so an older in-flight response cannot overwrite the patch and stale flags stay. `patchUpdatedSource` moves both its patches onto `patchCached`, and keeps `invalidatedKeys.remove(workspaceKey)` as one line after the workspace patch (that page is now current). | One answer to "what if a load is in flight". | A third patch path with its own rules. |
| D-19 | One hidden-key policy. `CatalogQueryRegistry.Spec.evictWhenHidden: Bool` is replaced by `var whenHidden: CatalogHiddenPolicy = .revalidate` with `enum CatalogHiddenPolicy { case revalidate, markStale, evict }` *(new)*; `evictsWhenHidden(_:)` becomes `hiddenPolicy(_:)`. `apply(invalidations:stale:)` switches on it once for an invalidated, cached key not in `visibleKeys`: `.evict` evicts (today's `conclusionDetail`), `.markStale` adds it to `invalidatedKeys` so the next `query` reloads it while the old value shows, `.revalidate` reloads now. `conclusionDetail` → `.evict`; the three conclusion lists and `sourcesList` → `.markStale`. It covers the effects and mutation paths alike. `CatalogQueryRegistryTests.onlyConclusionDetailKeysEvictWhenHidden` is renamed `hiddenPolicies` and expects `.evict` for `conclusionDetail`, `.markStale` for `personsList` and `sourcesList`, `.revalidate` for `sourceWorkspace`. Two existing tests change because People is hidden in both: rename `applyPromoteRevalidatesWarmedPersonsList` to `applyPromoteMarksHiddenPersonsListStale`; in `applyTriggerEvictsHiddenDetailPagesAndRevalidatesTheVisibleOne`, replace the list's `waitForFetchComplete` and new-name check and its doc comment's "revalidates list keys as usual". Both assert stale the observable way: right after `apply`, `list.isFetching == false` and the old value; then `let _: QueryHandle<[CatalogPersonHeader]> = session.query(listKey)`, `await waitForFetchComplete(list)`, and the new value. | One concept with three cases; setting two flags is impossible. A save made away from a list reloads no list until the researcher looks at one. | A second Boolean beside `evictWhenHidden`. `.evict` for lists — Back would show a skeleton and lose the S10-40 anchor's rows until reload. |
| D-20 | The Sources list reads the engine cache. `api/proto/engine.proto` `Source` gains `string updated_at = 13;` *(new)*. In `core/database/sources/sources.go`, `sqlList` drops the correlated `MAX(t.revision)` subquery; `Source.UpdatedRevision int64` becomes `Updated effects.Activity`, filled by `List` from `c.Graph().Activity(effects.ScopeSource, s.ID)` per row after the scan, and by `AttachActivity` (renamed from `AttachLatestRevision`; one non-test caller) the same way (`sqlLatestRevision` is deleted). `sourceProto` maps `Updated.Revision` and `Updated.At`. Swift: D-4. | The same cache and the same patch for every last-updated fact; no Sources read scans history. | Leaving Sources to S10-41 — the user chose this PR (2026-10-15). |
| D-21 | The cell composition stays in `ConclusionListRow` for this PR. S10-41 moves `updatedDetail`, `labelWithUpdated`, and `updatedColumnWidth` to one feature-neutral home that both the conclusion rows and `SourcesSplitRow` call; it does not copy them. The S10-41 map row says so. | One cell composition. Only one caller exists today. | Moving it now with no second caller (Q6). |
| D-22 | Source-page and graph writes return their effects. Each of these responses gains `WriteEffects effects` at its next free field number: `AddSourceNoteResponse`, `UpdateSourceNoteResponse`, `DeleteSourceNoteResponse`, `SetSourceMetadataResponse`, `ClearSourceMetadataResponse`, `DismissSourceMetadataSuggestionResponse`, `ReorderSourceMetadataResponse`, `CreateArtifactResponse`, `UpdateArtifactResponse`, `DeleteArtifactResponse`, `IngestArtifactFileResponse`, `UpsertSourceCredibilityAssessmentResponse`, `CreateSubjectResponse`, `UpdateSubjectResponse`, `DeleteSubjectResponse`. Each handler (`api/ffi/handlers/sources.go`, `artifacts.go`, `credibility.go`, `subjects.go`) keeps the `writes.Result` it now discards and sets `Effects: writeEffectsProto(wrote)`. `SetSubjectPosition` is not in the list: it writes a layout row with no audit and no effects. The matching `GenealogyStore` methods return `CatalogWriteEffects` with their current value (a tuple when there is one, as the composer methods already do); `GoStore` maps with `mapWriteEffects`; `FakeStore` returns `CatalogWriteEffects(sources: [sourceID], activity: CatalogActivity(revision: nextAuditRevision, at: now))` where it already bumps a Source's revision and `.empty` elsewhere. | The write already knows exactly what it touched and when; the app should be told instead of guessing with a blanket trigger. | A preparatory PR — the user chose this PR (2026-10-15). |
| D-23 | One apply for a write that has effects and a mutation. `WorkspaceSession.apply(_ effects: CatalogWriteEffects, _ mutation: CatalogMutation)` *(new)*: runs `patchLastUpdated(effects)`, then `apply(invalidations:stale:)` once with the union, without duplicates, of `registry.invalidations(for: effects)` and `registry.invalidations(by: mutation)`, and `registry.staleMarks(by: mutation)`; then the mutation's `patchedSource` and both `refreshSourceGraphProgress` calls as the two single applies do today. The single-argument `apply(_ effects:)` and `apply(_ mutation:)` stay for writes that have only one. `SourcePageContext.notifyWorkspaceMutated()` and `notifyMetadataMutated()` become `applyWrite(_ effects:)` and `applyMetadataWrite(_ effects:)`, calling it with `.mutatedSourceWorkspace(sourceId:)` / `.mutatedSourceMetadata(sourceId:)`. The graph and composer callers of `.mutatedSourceGraph` and `.deletedSubject` (`EvidenceGraphModel`, `CitationComposerModel`) pass the write's effects the same way. Every place that applies `.mutatedSourceWorkspace`, `.mutatedSourceMetadata`, `.mutatedSourceGraph`, or `.deletedSubject` passes the write's effects. | Keys the effects cannot name (the Source page's own `sourceWorkspace`, metadata lists) still follow the mutation; lists follow the effects. One key is never invalidated twice. | Two separate applies — the same key could be cancelled and restarted twice. |
| D-24 | The blanket triggers go. `CatalogQueryRegistry.conclusionTriggers` drops `.mutatedSourceWorkspace` and `.deletedSubject` (lists and detail pages follow `touchedHandles`; a credibility change or a Subject delete reports its handles). `sourcesList` drops `markStaleOn` entirely (its rows are patched from `effects.sources`). `CatalogMutation.savedCitation(sourceId:)`, `CatalogMutationKind.savedCitation`, and every registry entry and `switch` arm naming them are deleted: nothing constructs it (the composer applies effects). The registry comments that called this trigger an over-bust are rewritten. | A Source note no longer reloads three conclusion lists and the Sources list. | Keeping the kinds in the triggers "to be safe" — that is the over-bust. |
| D-25 | Split Source activity from Source evidence, mirroring `Handles` / `HandleScope`. `effects.Effect` gains `SourceScope Path` *(new)*: the Source whose last activity (audit `source` scope) a change moves. Every table that has a `Source` path today gets that same path as `SourceScope`, so stored audit scopes do not change (T-8 and `TestSourceScopes` hold them). `Source` keeps one meaning: "this Source's evidence changed" — the reads the Source graph (`graphcache/source.go`), graph progress, citation counts, and citations-by-artifact load. A table keeps `Source` only if one of those reads loads it; the implementer checks each read's SQL and removes `Source` from the rest (expected: `source_notes`, `source_metadata`, `source_metadata_layout`, `source_credibility_assessments`, and `sources` itself unless the graph reads a column of it). `CacheSource` has the same meaning as the new `Source`, so it is merged into `Source` on its two tables (positions and `identity_claim_evidence`) and `Effect.CacheSource`, `Set.CacheSources`, and `WriteEffects.cache_sources` (field 3, now `reserved 3;`) are deleted. `Set.SourceScopes` *(new)*; `effects.Scopes` maps `ScopeSource` to `Set.SourceScopes`, so `audit.Record` and `graphcache.applyActivity` read it; `graphcache.OnCommit` drops graphs for `Set.Sources` only. Wire: `WriteEffects.source_scopes = 10` *(new)*; Swift `CatalogWriteEffects.sourceScopes` *(new)*, `cacheSources` deleted, `touchedSources` deleted (callers read `sources`); `CatalogLastUpdatedScope.source.ids(in:)` returns `sourceScopes`. `invalidations(for:)` keys `sourceGraph`, `citationCounts`, `citationsByArtifact`, and `propertiesWorkspace` on `effects.sources` (evidence) only, and graph progress refreshes for `sources` only. `propertiesWorkspace` and `citationsByArtifact` take `whenHidden: .markStale`. | A note save moves the Sources row and reloads only the Source page, in the engine and the app. One field, one meaning. | An app-side filter by write kind — the engine already knows which tables feed which reads. Keeping `CacheSource` beside the new `Source` — two names for one fact. |
| D-26 | One `apply` body. `WorkspaceSession.apply(_ effects: CatalogWriteEffects = .empty, _ mutation: CatalogMutation? = nil)` holds the only body: `patchLastUpdated`, then the mutation's `patchedSource` (before invalidation, as `apply(_ mutation:)` does today), then one `apply(invalidations:)` of the deduped union, then graph progress for `effects.sources` and the mutation's Source. `apply(_ effects:)` and `apply(_ mutation:)` forward to it; `refreshSourceGraphProgress(for:)` is deleted; the stacked doc comment is split. After an Artifact delete, `SourcePageModel` calls `context.refreshCoverFromStore()` as create and attach do, so the Sources row gets `hasArtifact`, cover, and activity from the engine. | One place to add a step; the copies already drifted. Artifact delete changes row fields the activity patch cannot. | Three bodies. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `effects.Set.HandleScopes` *(new)* | field | `effects.HandleScopes` func | Extends `Resolve` with the same path; the standalone func is deleted (D-13). |
| `WriteEffects.handle_scopes` *(new)* | proto field | `header_handles` | Different invalidation (D-2). |
| `CatalogWriteEffects.handleScopes` *(new)* | field | `headerHandles` | Same as above. |
| `updated: CatalogActivity` on three header models and `CatalogSource` *(new)* | field | `CatalogSource.updatedRevision` | One value for revision and time (D-4); replaces `updatedRevision`. |
| `TimestampFormat.Style`, `TimestampFormat.Context`, `TimestampFormat.string` *(new)* | enum / struct / func | `TimestampFormat.boardStamp`, `abbreviatedDateTime` | Replaces both; one entry point for every catalog timestamp. |
| `L10n.Timestamp` *(new)* | L10n namespace | `L10n.LastUpdated` | Day words and the stamp joiner belong to date formatting, used by every style. |
| `effects.Scopes`, `effects.Scope`, `effects.ScopeSource`, `effects.ScopeCanonicalEntity`, `effects/scopes.go` *(new)* | registry | `audit.scopeResolvers`, `audit.Scope*` | Moved, not copied: `audit` and `graphcache` both read it; `audit`'s copies are deleted. |
| `effects.Commit`, `effects.Activity`, `effects/commit.go` *(new)* | type / file | `writes.Result`; `graphcache.Activity` (round 0 of the rework) | `Result` becomes an alias of `Commit`; one `Activity` type for the commit, the cache, and the row types. |
| `Graph.Activity`, `graphcache/activity.go` *(new)* | func / file | `audit.LatestByScope` | Moved into the cache that patches it; the audit copy is deleted. |
| `WriteEffects.revision`, `revision_at` *(new)* | proto fields | — | The write's own time; nothing carried it. |
| `CatalogActivity`, `CatalogLastUpdatedRow`, `CatalogLastUpdatedScope`, `CatalogLastUpdatedPatch`, `CatalogLastUpdated.swift`, `Spec.lastUpdated`, `lastUpdatedPatch(_:)`, `patchLastUpdated` *(new)* | value / protocol / enum / struct / funcs | `graphcache.Activity`; `patchUpdatedSource` | `CatalogActivity` mirrors the engine value. The patch is declared per spec, so a list adds one line, not a switch arm. |
| `WorkspaceSession.patchCached` *(new)* | func | `setQueryValue`, `patchUpdatedSource` | The one patch primitive; `patchUpdatedSource` moves onto it. |
| `CatalogHiddenPolicy`, `Spec.whenHidden`, `hiddenPolicy(_:)` *(new)* | enum / field / func | `Spec.evictWhenHidden`, `evictsWhenHidden(_:)` | Replaces them; one policy with three cases. |
| `Source.updated_at` *(new)* | proto field | `Source.updated_revision` | The Sources row needs the time beside the revision, as the headers have. |
| `WriteEffects effects` on fifteen Source-page and graph responses *(new)* | proto fields | the citation and observation responses' `effects` | Same field, same mapping. |
| `WorkspaceSession.apply(_:_:)` *(new)* | func | `apply(_ effects:)`, `apply(_ mutation:)` | One application of both, without double invalidation. |
| `SourcePageContext.applyWrite`, `applyMetadataWrite` *(new; replace `notifyWorkspaceMutated`, `notifyMetadataMutated`)* | funcs | `CitationComposerContext.applyWrite` | Same shape on the Source page. |
| `L10n.LastUpdated` *(new)* | L10n namespace | `L10n.Conclusions` | Shared with the Sources row in S10-41, so not under Conclusions. |
| `PVListDetail`, `PVListDetailColumn`, `PVList.detail`, `PVListSkeleton.detailWidth` *(new)* | component slot | `PVList.meta` | `meta` is the mono ref; the brief puts the time in its own column. |
| `PVSymbol.history` *(new)* | icon case | `PVSymbol.refresh` (`arrow.clockwise`) | `refresh` means a moved Promote row; the boards draw Lucide `history`. |
| `ConclusionListKind.updatedAt`, `ConclusionListRow.updatedColumn`, `updatedDetail`, `labelWithUpdated`, `updatedColumnWidth`, `ConclusionListPage.skeleton` *(new)* | protocol req / funcs | `ConclusionListRow.title(_:extraCount:)` | Same helper enum; composes the formatter, L10n, and the kit slot. |
| `api/ffi/handlers/write_effects_test.go` *(new)* | test file | `observations_test.go` | `writeEffectsProto` had no test. |

**Builds on.** The effects registry (`registry.go`), the `WriteEffects` → `CatalogQueryRegistry` invalidation path, `TimestampFormat`, and `PVList`.
**Deletes.** `CatalogMutation.savedCitation` and its kind; `.mutatedSourceWorkspace` and `.deletedSubject` in `conclusionTriggers`; `sourcesList.markStaleOn`; `SourcePageContext.notifyWorkspaceMutated` / `notifyMetadataMutated`; `sources.sqlLatestRevision` and `sqlList`'s scope subquery; `Spec.evictWhenHidden` and `evictsWhenHidden(_:)`; the hard-coded `stale: [.sourcesList]` in `apply(_ effects:)`; `LastUpdatedDisplay`, `TimestampFormat.boardStamp`, `TimestampFormat.abbreviatedDateTime` and their tests; `L10n.LastUpdated.today`, `.yesterday` and their catalog keys; `audit.scopeResolvers`, `audit.ScopeSource`, `audit.ScopeCanonicalEntity`, `audit.Latest`, `audit.LatestByScope` (tests move to `graphcache`); `effects.Sources`, `effects.HandleScopes` (standalone resolvers); `conclusionheaders.latestHandles`; the `handleScopes` branch in `invalidations(for:)`.
**Next adds.** S10-41 moves the cell composition to a shared home (D-21) and draws it on `SourcesSplitRow`, reading `CatalogSource.updated`; the patch, the engine cache, and the sort already work. S10-35 fills the `PVList` thumbnail closure in `ConclusionListBody` with the likeness.

### Schema

None — no catalog change. S10-21 stored the scopes.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `Set.HandleScopes [][]byte` *(new)*; `Resolve` appends `eff.HandleScope`. Rework: `scopes.go`, `commit.go` (D-13, D-14). |
| `core/database/audit`, `core/writes`, `core/database` (`listen.go`) | Rework: D-13, D-14. |
| `core/database/graphcache`, `core/database/conclusionheaders`, `core/database/sources` | Rework: D-15, D-20. |

### FFI

`api/proto/engine.proto` `WriteEffects`: `repeated string handle_scopes = 7;` *(new)* with a comment: handles whose last-updated revision this write moved. Handler `api/ffi/handlers/write_effects.go`: `HandleScopes: idStrings(set.HandleScopes)`. Rework: `WriteEffects.revision = 8`, `revision_at = 9` (D-16); `Source.updated_at = 13` (D-20). Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| List row column | `PVList` (Extend) | `macos/App/DesignSystem/Components/List/PVList.swift` | D-7. Column between the text `VStack` and the trailing `HStack`, `.frame(width: detail.width, alignment: .leading)`. Text truncates first: the column and trailing stack keep `layoutPriority(1)`. |
| Skeleton bar | `PVListSkeleton` (Extend) | same file | D-7. |
| Conclusion page | snowflake | `macos/App/Features/Conclusions/ConclusionListPage.swift` | D-8: `detail:`, the label wrap, and `skeleton()` at both skeleton sites. |
| Cell text | `TimestampFormat.string` | `macos/App/Platform/TimestampFormat.swift` | D-5 (rework). The row composes it (D-6). |
| Cache patch | `CatalogLastUpdatedPatch`, `patchCached` | `macos/App/Features/Workspace/Session/` | D-17, D-18, D-19. |

Agreed frames (today in the frames is 14 Oct 2026, Mountain Time): **Times** — six rows per list, still in list order (Persons by name, Events by date, Places by name); each column cell is the history icon then “Today, 3:15 PM”, “Today, 9:02 AM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, or a faint “—”; the 136pt column lines up down the list between the text and the mono ref. **Help** — hovering “Yesterday” shows “Last updated 13 Oct 2026 · 4:40 PM MDT”; the dash shows nothing. **Narrow window** — a long title with +2 and a long secondary line truncate; dates, +N, column, ref, and chevron keep their width. **No secondary line** — one-line title, column and ref centered; a second row shows the `en_US` form “Sep 30, 2026”. **First load** — the skeleton has a 64pt bar in the column before the 72pt ref bar. **Empty** — the unchanged empty state; no rows, no column. Persons frames 08–09 (likeness thumb) are S10-35 and S10-39, not this PR. States: empty list → `PVEmptyState`, no column; failed → `PVCallout`; 1000 rows → two `TimestampFormat.string` calls per visible row for the cell and help and one for the spoken label (lazy); a save patches O(rows) in memory per cached list with no FFI call; across midnight the two can differ for one render, which D-9 accepts. Accessibility: D-10.

### Strings

Round 0 (superseded in part by D-12): L10n keys *(new)*, namespace `enum LastUpdated` in `macos/App/Platform/L10n.swift` directly after `enum Conclusions`, each with a `comment`:
- `lastUpdated.today` = "Today, %@" — time of a change made today.
- `lastUpdated.yesterday` = "Yesterday".
- `lastUpdated.a11y.updated` = "updated %@".
- `lastUpdated.a11y.unknown` = "last updated unknown".
- `lastUpdated.help` = "Last updated %@" — hover help; the argument is the full stamp.

Accessors, in the style of `L10n.Conclusions.a11yList`: `static func today(time: String) -> String`, `static var yesterday: String`, `static func a11yUpdated(_ text: String) -> String`, `static var a11yUnknown: String`, `static func help(stamp: String) -> String`; each returns `String` through `L10n.string` / `L10n.format`.

Rework (D-12): `L10n.Timestamp` *(new)* with `today(time:)`, `todayAt(time:)`, `yesterday`, `stamp(date:time:)`. Removed: `L10n.LastUpdated.today`, `.yesterday` and keys `lastUpdated.today`, `lastUpdated.yesterday`.

## 7. Steps

**STEP-1 — Resolve returns handle scopes**
Files: `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/database/effects/README.md`, `core/database/effects/handle_paths_test.go`.
Do: D-1. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/writes/` passes.

**STEP-2 — Scope handles on the wire**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/write_effects.go`, `api/ffi/handlers/write_effects_test.go` *(new)*. Skill: `add-ffi-handler`.
Do: D-2 Go side. Run `./scripts/generate-proto.sh`. Add T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `git diff --stat` shows the regenerated Go and Swift protobuf files.

**STEP-3 — Scope handles reach the app** (round 0 also invalidated the lists; D-3 removes that)
Files: `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Platform/GoStore.swift` `mapWriteEffects`, `macos/App/Features/Workspace/Session/CatalogQueryRegistry.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: D-2 Swift side, D-3. Add T-3. The existing `invalidations(for:)` expectation stays unchanged.
Check: `xcodebuild test -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' -only-testing:ProvenenciaTests/CatalogQueryRegistryTests CODE_SIGNING_ALLOWED=NO` passes.

**STEP-4 — Header models carry the time**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift` (`mapPersonHeader`, `mapEventHeader`, `mapPlaceHeader`).
Do: D-4. Map `h.updatedAt`.
Check: the app target builds (`xcodebuild build -project macos/Provenencia.xcodeproj -scheme Provenencia -destination 'platform=macOS' CODE_SIGNING_ALLOWED=NO`).

**STEP-5 — Last-updated value**
Files: `macos/App/Platform/TimestampFormat.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/TimestampFormatTests.swift`. Skill: `add-localized-string`.
Do: D-5, D-6, D-11 (`columnWidth = 136`), Strings. Add T-4, T-5.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `-only-testing:ProvenenciaTests/TimestampFormatTests` passes.

**STEP-6 — List detail column**
Files: `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/DesignSystem/Components/Icon/PVIcon.swift` (`PVSymbol.history`), `macos/App/DesignSystem/README.md` (List row). Skill: `add-ui-component`.
Do: D-7. Change the `PVList` header comment from “no columns” to: one optional fixed-width text column before the ref, content-agnostic; still no selection and no sort. Add the detail column to the List row in `macos/App/DesignSystem/README.md`. Add a second `#Preview` list with `detail:`; the existing preview is unchanged. `ConclusionListPage.swift` is the only `PVList` caller.
Check: app builds; `-only-testing:ProvenenciaTests/PVListTests` passes.

**STEP-7 — The column on three lists**
Files: `macos/App/Features/Conclusions/ConclusionListPage.swift`, `ConclusionListRow.swift`, `PersonsListView.swift`, `EventsListView.swift`, `PlacesListView.swift`, `macos/ProvenenciaTests/PersonsListTests.swift`.
Do: D-8, D-9, D-10. Add T-6.
Check: `-only-testing:ProvenenciaTests/PersonsListTests` passes; manual check steps 1–4.

**STEP-8 — Docs**
Files: `docs/deployment-plan/spike-10/completed.md`, `docs/deployment-plan/spike-10/deployment-plan.md` (map row), this plan (status, Landed note), `docs/deployment-plan/spike-10/design/README.md` and the three briefs → `design/archive/`.
Do: per the Definition of done.
Check: `python3 scripts/check-plan-docs.py` reports no errors; links resolve.

**STEP-9 — Declare scopes once; resolve once per write; commit carries its time**
Files: `core/database/effects/scopes.go` *(new)*, `core/database/effects/resolve.go`, `core/database/effects/README.md`, `core/database/effects/handle_paths_test.go`, `core/database/audit/scopes.go`, `core/database/audit/audit.go`, `core/database/audit/*_test.go`, `core/database/listen.go`, `core/writes/run.go`, `core/writes/run_test.go`, `core/database/graphcache/commit.go` (signature only), `core/database/effects/commit.go` *(new)*, `core/database/graphcache/graph_test.go`, `core/database/audit/export_test.go`, `core/database/audit/scopes_effects_test.go`, `core/database/promote/bridges_test.go`, and every other `OnCommit(` / `Notify(` / `audit.Record(` / `audit.Scope` caller (`rg` them). Skill: `add-audit-scope`.
Do: D-13, D-14. Add T-7; rewrite T-8. Update `add-audit-scope` if it names `scopeResolvers` or `audit.Scope*`.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg "scopeResolvers|audit\.Scope(Source|CanonicalEntity)|effects\.(Sources|HandleScopes)\(" core api` finds nothing.

**STEP-10 — Engine last-activity cache**
Files: `core/database/graphcache/activity.go` *(new)*, `graphcache/sql.go`, `graphcache/commit.go`, `graphcache/graph.go`, `graphcache/graph_test.go`, `core/database/conclusionheaders/compose.go`, `conclusionheaders.go`, `events.go`, `places.go`, `conclusionheaders_test.go`, `core/database/sources/sources.go` and its tests, `api/ffi/handlers/sources.go`, `api/ffi/handlers/thumbnails.go`, `core/database/audit/scopes.go`, `audit/scopes_test.go`.
Do: D-15 and the Go side of D-20, except `updated_at`: `sourceProto` maps `Updated.Revision` only (STEP-11 adds `UpdatedAt` with the field); `api/ffi/handlers/sources.go` `s.UpdatedRevision = res.Revision` becomes `s.Updated = graphcache.Activity{Revision: res.Revision, At: res.At}`; `AttachLatestRevision`'s early return checks `s.Updated.Revision > 0`. Move `TestLatestByScope`'s cases into T-9; add T-14.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/... ./api/...` passes; `rg "LatestByScope|audit\.Latest" core api` finds nothing.

**STEP-11 — Revision time on the wire**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/write_effects.go`, its seven call sites, `write_effects_test.go`, the `Source` proto mapping in `api/ffi/handlers`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `macos/App/Platform/GoStore.swift` (`mapWriteEffects`, the three header mappers, the `CatalogSource` mapping), `macos/App/Platform/GenealogyStore.swift`, `macos/App/Features/Conclusions/PersonsListView.swift`, `EventsListView.swift`, `PlacesListView.swift` (`updatedAt(_:)` returns `header.updated.at`), `macos/App/Features/Sources/SourcesModel.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (its `updatedRevision` lines), and `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` *(new; only `CatalogActivity` in this step; in `project.pbxproj` add a `PBXFileReference`, a `PBXBuildFile`, the `Session` group child, and the app Sources phase entry, copying `CatalogMutation.swift`'s four entries with the next unused `AA0000…` ids)*. Skill: `add-ffi-handler`. (new)
Do: D-4, D-16, D-20 (wire field `updated_at`, `sourceProto` maps `Updated.At`, Swift model). Extend T-2. Run `./scripts/generate-proto.sh`.
Check: Go tests pass; the app target builds.

**STEP-12 — One timestamp formatter**
Files: `macos/App/Platform/TimestampFormat.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/App/Features/Sources/SourcePageNotesView.swift`, `macos/App/Features/Onboarding/OnboardingProjectMetaLines.swift`, `macos/ProvenenciaTests/TimestampFormatTests.swift`. Skill: `add-localized-string`, `add-swift-test`.
Do: D-5 and the additive half of D-12: add `L10n.Timestamp`, `TimestampFormat.Style`, `Context`, `string`; move `SourcePageNotesView` and `OnboardingProjectMetaLines` onto it; delete `abbreviatedDateTime` and its test. Leave `boardStamp`, `L10n.LastUpdated.today` / `.yesterday`, and `LastUpdatedDisplay` in place (STEP-13 deletes them). Add T-10.
Check: `python3 scripts/check-localizable-xcstrings.py` passes; `-only-testing:ProvenenciaTests/TimestampFormatTests` passes.

**STEP-13 — The row composes the cell**
Files: `macos/App/Features/Conclusions/ConclusionListRow.swift`, `ConclusionListPage.swift`, `macos/App/Platform/TimestampFormat.swift`, `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`, `macos/ProvenenciaTests/PersonsListTests.swift`, `macos/ProvenenciaTests/TimestampFormatTests.swift`.
Do: D-6, D-8, D-10, D-11, and the deleting half of D-12: delete `LastUpdatedDisplay`, `boardStamp`, `L10n.LastUpdated.today` / `.yesterday` and their catalog keys, and the T-4, T-5, and `boardStamp` tests. Rewrite T-6; add T-11.
Check: `-only-testing:ProvenenciaTests/PersonsListTests -only-testing:ProvenenciaTests/TimestampFormatTests` passes; `python3 scripts/check-localizable-xcstrings.py` passes; AC-3.8 holds.

**STEP-14 — Patch in place; stale when hidden**
Files: `macos/App/Features/Workspace/Session/CatalogLastUpdated.swift` *(created in STEP-11)*, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `macos/ProvenenciaTests/WorkspaceSessionTests.swift`. Skill: `add-workspace-place`. (new)
Do: D-3, D-17, D-18, D-19. Rewrite T-3; add T-12, T-13; change the three tests D-19 names.
Check: `-only-testing:ProvenenciaTests/CatalogQueryRegistryTests -only-testing:ProvenenciaTests/WorkspaceSessionTests` passes.

**STEP-15 — Docs for the rework**
Files and the rule each rewrites:
- `docs/macos-client-patterns.md` rule 4 and the sentence that a patch is "never a substitute" for invalidation: a write that reports its activity is patched through a declared `CatalogLastUpdatedPatch`; invalidation is for changes the client cannot know; `whenHidden` policies.
- `.agents/skills/add-workspace-place/SKILL.md` (the "always pair with `apply` so the registry still refetches" line): same.
- `core/database/effects/README.md`: `Scopes`, `Commit`, `Resolve` once per write.
- `core/writes/README.md`, `core/database/catalogmodel/README.md`, `core/database/rowchange/README.md`: where they say `audit.Record` resolves scopes from `effects.Sources`.
- `.cursor/rules/catalog-audit.mdc` (reading with `audit.LatestByScope`): read last activity from `Graph.Activity`.
- `.agents/skills/add-audit-scope/SKILL.md`, `.agents/skills/add-catalog-query/SKILL.md`, `.agents/skills/add-catalog-delete/SKILL.md`: moved symbols.
- `docs/audit-revision-history.md` lines 330–337 (`scopeResolvers`, `effects.HandleScopes`, `LatestByScope`).
- `docs/deployment-plan/spike-10/deployment-plan.md`: the S10-22 In line (patched, not invalidated; Sources included); the S10-41 In line (`CatalogSource.updated` from the engine cache, not `audit.LatestByScope(db, audit.ScopeSource)`; it moves the cell composition per D-21).
- This plan's Landed note at merge.
Do: then sweep with `rg "scopeResolvers|LatestByScope|audit\.Latest|audit\.Scope(Source|CanonicalEntity)|effects\.(Sources|HandleScopes)\(|boardStamp|abbreviatedDateTime|LastUpdatedDisplay|evictWhenHidden" .agents .cursor docs core api macos` and fix every hit outside `docs/deployment-plan/archive/` and this PR's plan, review, and ship logs.
Check: that sweep is clean; `python3 scripts/check-plan-docs.py` reports no errors; `scripts/check-pr.sh` PASS.

**STEP-16 — Source-page and graph writes return effects (engine)**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/sources.go`, `artifacts.go`, `credibility.go`, `subjects.go`, and their `_test.go` files. Skill: `add-ffi-handler`.
Do: D-22 (Go side). Add T-15.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/... ./core/...` passes.

**STEP-17 — The app applies those effects; blanket triggers go**
Files: `macos/App/Platform/GenealogyStore.swift`, `GoStore.swift`, `FakeStore.swift`, `macos/App/Features/Sources/SourcePageContext.swift`, `SourceNotesSection.swift`, `SourceMetadataSection.swift`, `SourceArtifactsSection.swift`, `SourceCredibilitySection.swift`, `SourcePageModel.swift`, `macos/App/Features/EvidenceGraph/EvidenceGraphModel.swift`, `macos/App/Features/CitationComposer/CitationComposerModel.swift`, `macos/App/Features/Workspace/Session/CatalogMutation.swift`, `CatalogQueryRegistry.swift`, `WorkspaceSession.swift`, every other caller `rg` finds for the changed store methods, `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift`, `WorkspaceSessionTests.swift`. Skill: `add-workspace-place`, `add-swift-test`.
Do: D-22 (Swift side), D-23, D-24. Add T-16, T-17; fix tests that named `savedCitation`.
Check: the app target builds; `-only-testing:ProvenenciaTests` passes.

**STEP-18 — Docs for the effects path**
Files: `docs/macos-client-patterns.md` and `.agents/skills/add-workspace-place/SKILL.md` (a write returns `WriteEffects`; `apply(effects, mutation)` when a mutation still names keys the effects cannot), `.agents/skills/add-ffi-handler/SKILL.md` if it describes write responses, `docs/deployment-plan/spike-10/deployment-plan.md` (S10-22 In line).
Check: `scripts/check-pr.sh` PASS.

**STEP-19 — Activity is not evidence; one apply body** — built from the ship log's Fix plan R2 (D-25, D-26).
Files: as Fix plan R2 lists. Do: D-25, D-26; T-18, T-19, T-20.
Check: `scripts/check-pr.sh` PASS.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/handle_paths_test.go` `TestHandlePaths` subtest "resolve reports note handle scopes" *(new)* | The `observation_note` create change from "observation note reaches member handle" → call `effects.Resolve(tx, []rowchange.Change{ch})` inline (no new helper); `.HandleScopes` is `{bostonRes.Entity.ID}`, `.Handles` is empty | AC-2.1, D-1 |
| T-2 | Go | `api/ffi/handlers/write_effects_test.go` `TestWriteEffectsProto`, table-driven | `Result{}` → all empty, `Revision 0`, `RevisionAt ""`; `Result{Revision: 42, At: T, Effects: Set{HandleScopes: [id]}}` → `HandleScopes == [uuid string]`, `Revision 42`, `RevisionAt T`; `Set{Handles: [a], HeaderDeps: [b]}` → `Handles [a]`, `HeaderHandles [b]` | AC-2.1, AC-2.8, D-2, D-16 |
| T-3 | Swift | `CatalogQueryRegistryTests` `handleScopesInvalidateNothing` *(rename of `handleScopesReloadListsNotDetails`)* | `handleScopes: ["e1"]` → `[]`; `handles: ["e1"], handleScopes: ["e1","e2"]` → the three lists then `conclusionDetail(e1)` | AC-2.2, AC-2.3, D-3 |
| T-4 | — | (replaced by T-10) | — | D-5 |
| T-5 | — | (replaced by T-10, T-11) | — | D-6 |
| T-6 | Swift | `PersonsListTests` `rowLabelEndsWithUpdated` | `ConclusionListRow.labelWithUpdated("Robins, James, PER-7KD45, b. 1817", updatedAt: yesterday, in: edmonton en_US)` → ends `, updated Yesterday`; `T` → ends `, updated October 12, 2026`; `""` → ends `, last updated unknown` | AC-1.6, D-6, D-10 |
| T-7 | Go | `core/writes/run_test.go` `TestRunReportsCommitTime` *(new)* | A write → `Result.At` equals the `audit_transactions.created_at` of `Result.Revision`; a listener registered with `Listen` receives the same `Commit` | AC-2.8, D-14 |
| T-8 | Go | `core/database/audit/scopes_effects_test.go` `checkNewEffects` (helper), driven by its existing callers in `scopes_test.go` (`TestSourceScopes`, `TestHandleScopes`: note save, member Observation delete, promote) | Rewritten over `effects.Scopes` and `effects.Resolve`: for each scope, stored rows for the revision equal `Scope.IDs(Resolve(changes))` | AC-2.10, D-13 |
| T-9 | Go | `core/database/graphcache/graph_test.go` `TestActivity` *(new)*, table-driven | First `Activity` read loads once (cases moved from `TestLatestByScope`: newest revision with its time, zero for no rows, `""` and an unknown type return an error, a nil `*Graph` returns zero; queries counted with `SetQueryCounterForTest`); `OnCommit(Commit{Revision: 7, At: T, Effects: {HandleScopes: [h]}})` → `Activity(canonical_entity, h) == {7, T}` with no new query; an older commit leaves it; `Drop` then read → reloaded with one query | AC-2.9, D-15 |
| T-10 | Swift | `TimestampFormatTests` `stylesFormatOneLine` *(new)*, parameterized | The AC-1.2 and AC-1.3 inputs through `.relative` (`Yesterday`; `Oct 12, 2026`, `en_GB` `12 Oct 2026`), and each AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6 input → its string (U+202F normalized to a space before comparing, as round 0 does); AC-3.7 inputs → `nil` | AC-1.2, AC-1.3, AC-3.1, AC-3.2, AC-3.3, AC-3.4, AC-3.5, AC-3.6, AC-3.7, D-5 |
| T-11 | Swift | `PersonsListTests` `updatedDetailComposesTheCell` *(new)* | AC-1.1 input → text `Today, 3:15 PM`, help starts `Last updated `, not faint; `""` and `"garbage"` → text `—`, help `nil`, faint | AC-1.1, AC-1.4, AC-1.5, D-6 |
| T-12 | Swift | `WorkspaceSessionTests` `applyEffectsPatchesLastUpdatedInPlace` *(new)* | AC-2.5, AC-2.5b, AC-2.5c: patched values, `isFetching == false` after the apply (AC-2.5c: start a load, apply, the load restarts and its old response does not replace the patch) | AC-2.5, AC-2.5b, AC-2.5c, D-17, D-18 |
| T-13 | Swift | `WorkspaceSessionTests` `hiddenListsGoStaleVisibleListReloads` *(new)* | AC-2.6 and AC-2.7, asserted as D-19 describes | AC-2.6, AC-2.7, D-19 |
| T-14 | Go | `core/database/sources` list test (the existing one that checks `UpdatedRevision`) | After a write under a Source, `List` and `AttachLatestRevision` return `Updated` equal to that commit's `{Revision, At}`; a second `List` runs no scope query (`SetQueryCounterForTest` on the graph) | AC-2.9, D-20 |
| T-15 | Go | `api/ffi/handlers` tests for the four handler files *(new cases)*, table-driven | `AddSourceNote` → `effects.sources == [S]`, `revision > 0`, `handles` empty; `UpsertSourceCredibilityAssessment` on a Source with a filed Subject → `handles` include the handle; `DeleteSubject` of a filed Subject → `handles` and `sources` set | AC-2.11, AC-2.12, AC-2.13, D-22 |
| T-16 | Swift | `WorkspaceSessionTests` `sourceWriteEffectsPatchSourcesAndLeaveListsAlone` *(new)* | AC-2.11 with `apply(effects, .mutatedSourceWorkspace)`; AC-2.12 with `handles` → lists stale (next `query` reloads); the Source page's `sourceWorkspace` key is invalidated once | AC-2.11, AC-2.12, D-23 |
| T-17 | Swift | `CatalogQueryRegistryTests` `blanketTriggersAreGone` *(new)* | AC-2.14 | AC-2.14, D-24 |
| T-18 | Go | `core/database/effects` table test *(new)* and `TestSourceWritesReturnEffects` | A `source_note` change → `SourceScopes == [S]`, `Sources` empty; a `subjects` change → both `[S]`; a position change → `Sources == [S]`, `SourceScopes` empty; audit scope rows unchanged (T-8) | D-25 |
| T-19 | Swift | `WorkspaceSessionTests` `sourceNoteLeavesEvidenceReadsAlone` *(new)* | With `propertiesWorkspace`, `sourceGraph(s1)`, and a `citationsByArtifact` key warmed, apply a note's effects (`sourceScopes: ["s1"]`) with `.mutatedSourceWorkspace` → none fetching or stale; the Sources row patched | D-25, D-26 |
| T-20 | Swift | `SourcePageModel` or `WorkspaceSessionTests` *(new)* | Cached Sources list; delete the only Artifact → the row's `hasArtifact == false` | D-26 |

Must stay green: everything `scripts/check-pr.sh` runs. The rework is most likely to break `audit` scope tests, `graphcache` tests, `conclusionheaders` tests, and `WorkspaceSessionTests` (promote revalidation). Most likely to break: `CatalogQueryRegistryTests` (invalidation order), `TestHandlePaths`, and `PersonsListTests` (row labels).

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests, catalog and plan-doc checks)
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: core/database/effects/README.md (Resolve returns HandleScopes; Scopes; one walk per write); macos/App/DesignSystem/README.md (List row detail column); add-workspace-place and docs/macos-client-patterns.md (patch in place, `whenHidden` policies); add-audit-scope and docs/audit-revision-history.md where they name moved symbols
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with promoted people, events, and places. Open People: each row shows a time in its own column before the ref; the column lines up; the list order is unchanged. Hover a cell: the full stamp appears.
2. Open Events and Places: the same column, text, help, and em dash as People (AC-1.7).
3. Narrow the window: titles truncate; the column and ref keep their width.
4. Quit and reopen the project, open People: the loading skeleton shows a column bar, and rows do not shift when the times arrive (AC-1.8).
5. Open a Source, edit only the note on an Observation of a Subject filed to a Person, save. Open People: that row reads `Today, {now}` (AC-2.4).
6. Edit an Observation's value on an Event's member Subject. Open Events: that row moved.
7. With People open in one window state, edit a note from a Source page and come back with Back: the row's time moved, and the list did not flash a skeleton (patched, not reloaded).
8. A Source page note still shows its full stamp; the onboarding project card still shows its created and updated times.
9. AC-3.8: `rg "LastUpdatedDisplay|boardStamp|abbreviatedDateTime" macos` finds nothing, and no catalog timestamp is formatted outside `TimestampFormat.swift` (the exceptions AC-3.8 lists are not timestamps).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| A patched row and the engine disagree | Both take the same commit's revision and time; the app only moves a row forward (D-17); a stale-marked list reloads from the engine's cache on show (D-18). |
| The scope move breaks audit history | T-8 checks stored scope rows against `Resolve` for each write shape. No migration: stored rows are unchanged. |
| Locale strings differ from the examples | Tests pin locale and time zone (T-10). |
| `PVList` change alters lists that do not pass `detail` | The column draws only when `detail` is set (D-7); the existing `#Preview` stays unchanged. |

**Stop and report if:**

- `effects.Resolve` or `writeEffectsProto` no longer match §4.
- The agreed board places the time somewhere other than its own column before the ref, or changes PSL-2's wording.
- `ListPersonHeaders`, `ListEventHeaders`, or `ListPlaceHeaders` return empty `updated_at` for a promoted handle in the manual check (S10-21 regression).
- A note-only save in the app still leaves the row's time unchanged after step 5 of the manual check.
- Field number 7, 8, or 9 on `WriteEffects` is taken at your base.
- Moving the scope list or `Commit` into `effects`, or the activity map into `graphcache`, still produces an import cycle.
- `sources.List` cannot reach `c.Graph()` without an import cycle.
- `effects.Resolve` run before `audit.Record` resolves a different id set than it did after (T-8 fails on a delete).
- A Subject delete or a credibility change on a Source with a filed Subject reports no handles (T-15), so removing the triggers would leave lists wrong.
- A place applies `.mutatedSourceWorkspace`, `.mutatedSourceMetadata`, `.mutatedSourceGraph`, or `.deletedSubject` after a write whose store method cannot return effects.
- A `graphcache` read path is reachable without the catalog session's serialization, so the activity map would need a lock the rest of `Graph` does not have.
- `CatalogLastUpdatedPatch.apply` cannot be stored in the `Sendable` `Spec` with a `@MainActor` closure; report the compiler error.

## 12. Open questions

None.

## 13. PR

**Title:** S10-22 — Last updated on People, Events, and Places
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-22-list-last-updated.md`.
