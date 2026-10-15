# S10-22 — Last updated on People, Events, and Places

**Status:** Planned
**Spike / slice:** Spike 10 · Slice 6 — Lists and the Person page
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#s10-22--last-updated-on-people-events-and-places)
**Brief:** [S10-D9](../design/S10-D9-persons-list.md) (PSL-1…PSL-6), [S10-D10](../design/S10-D10-events-list.md) (EVL-1…EVL-4), [S10-D11](../design/S10-D11-places-list.md) (PLL-1…PLL-4); boards agreed 2026-10-15
**Board:** Claude Design projects *Persons list* (id 5c00cb55-8261-4579-b0e6-bc571ab23f78, file *Persons List.dc.html*, frames 01–07; 08–09 are S10-35 and S10-39), *Events list* (id 2799d7a0-387e-4570-b138-fc387562ac91, file *Events List.dc.html*, the frames marked S10-D10), and *Places list* (id e8910bb1-f564-4d7a-b829-69de0e60af5a, file *Places List.dc.html*, frames S10-D11 A–D). §6 macOS describes every frame this PR ships in words. The implementer does not need the boards.
**Depends on:** S10-21 (landed [#398](https://github.com/mendahu/provenencia/pull/398)), S10-D9, S10-D10, S10-D11
**Blocks:** —
**Branch:** `s10-22-list-last-updated`
**Version bump:** none — Spike 10 PRs leave `VERSION` at 0.2.0, and the product bump is cut when a release ships ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ffi-handler`](../../../../.agents/skills/add-ffi-handler/SKILL.md), [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md), [`add-workspace-place`](../../../../.agents/skills/add-workspace-place/SKILL.md)
**Rules that bind:** `.cursor/rules/go-tests.mdc`, `.cursor/rules/design-system-layers.mdc`
**Context:** [`core/database/effects/README.md`](../../../../core/database/effects/README.md), [S10-21 plan](S10-21-handle-last-activity.md), working rules "**Last updated** is an audit scope, not a timestamp column" and "**The Sources list is one page.**" in [`../deployment-plan.md`](../deployment-plan.md)
**Planned:** 2026-10-15 against `main` @ `360b405f` (boards folded in 2026-10-15)
**Review:** [`S10-22-list-last-updated.review.md`](S10-22-list-last-updated.review.md) — Fits with changes, Ready with changes (applied); 2 rounds

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

Each row on the People, Events, and Places lists shows when the researcher last worked on that handle, in its own column before the ref: “Today, 3:15 PM”, “Yesterday”, or a short date, with the full stamp as help and “—” when nothing is recorded. The value is the `canonical_entity` revision time S10-21 already returns. A note-only Observation save now refreshes the lists, so the cell moves after every edit that moves the revision.

**Why now.** S10-21 landed the engine fields (`updated_at` on `PersonHeader`, `EventHeader`, `PlaceHeader`). The rows do not show them, and the app drops them in `GoStore`.

**What stays.** List order (Go sorts by title or date), the title, the secondary lines, the thumb, the ref, the chevron, the empty state, and the count. No sort, filter, or search on these lists. The Sources list is untouched (S10-41). The Person, Event, and Place detail pages do not show the time.

## 2. User stories

**US-1.** As a researcher, on People, Events, and Places, I see when I last worked on each row, so I can find what I was doing.

Times below use `now` = 2026-10-14 18:00 in `America/Edmonton`, locale `en_US`.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Row `updatedAt` = `2026-10-14T21:15:00Z` (15:15 local) | Cell reads `Today, 3:15 PM`. Help is `L10n.LastUpdated.help(stamp: TimestampFormat.boardStamp(updatedAt))` (on a machine in Edmonton: `Last updated 14 Oct 2026 · 3:15 PM MDT`). |
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
| AC-2.2 | `CatalogWriteEffects(handleScopes: ["e1"])` | Invalidations are the three list keys and nothing else (no `conclusionDetail`). |
| AC-2.3 | `CatalogWriteEffects(handles: ["e1"], handleScopes: ["e1", "e2"])` | The three list keys once each, and `conclusionDetail` for `e1` only. |
| AC-2.4 | In the app: edit a note on Bob's Observation, then open People | Bob's cell reads `Today, {now}`. |

## 3. Scope

| In | Covered by |
| --- | --- |
| `effects.Set.HandleScopes` filled by `effects.Resolve` | STEP-1, AC-2.1 |
| `WriteEffects.handle_scopes` on the wire, mapped in Swift, and invalidating the three lists | STEP-2, STEP-3, AC-2.1, AC-2.2, AC-2.3 |
| `updatedAt` on `CatalogPersonHeader`, `CatalogEventHeader`, `CatalogPlaceHeader`, mapped from the list responses | STEP-4, AC-1.1 |
| Last-updated text, help, and em dash (PSL-2, PSL-3, PSL-4) | STEP-5, AC-1.1, AC-1.2, AC-1.3, AC-1.4, AC-1.5 |
| `PVList` optional detail column and skeleton bar (PSL-1, PSL-5, PSL-6) | STEP-6, AC-1.8 |
| The column on all three lists and the spoken suffix (PSL-6; EVL-1 to EVL-4; PLL-1 to PLL-4) | STEP-7, AC-1.6, AC-1.7 |

| Out | Home |
| --- | --- |
| The Sources list cell, filters, search, persisted settings | S10-41 (S10-D8) |
| Likeness thumb on Persons (PSL-7) | S10-35 |
| Preferred portrait as the thumb (PSL-8) | S10-39 |
| Sort by last updated on People, Events, or Places | not planned (working rule "The Sources list is one page") |
| Last updated on the Person, Event, or Place detail page | not planned |
| A timer that re-renders “Today” into “Yesterday” while a list sits open | not planned (brief §2.1) |
| `audit.Record` reusing `Set.HandleScopes` instead of resolving again | not planned: two reads of one path inside one write; no measured cost |
| Fixture times in `FakeStore` or `PreviewFixtures` headers | not planned: previews and UI tests show the em dash |

## 4. Current state

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
| Tests | `core/database/effects/handle_paths_test.go` `TestHandlePaths` "observation note reaches member handle"; `macos/ProvenenciaTests/CatalogQueryRegistryTests.swift` (the `invalidations(for: CatalogWriteEffects(…))` expectation); `macos/ProvenenciaTests/TimestampFormatTests.swift`; `macos/ProvenenciaTests/PersonsListTests.swift`; `macos/ProvenenciaTests/PVListTests.swift` | Cover the paths, invalidation, stamp, labels, and keyboard as today. |

**Precedent.** `header_handles` (field 6) was added to `WriteEffects` the same way: a field on `effects.Set`, one line in `writeEffectsProto`, one in `mapWriteEffects`, and a branch in `invalidations(for:)`. `header_handles` is filled from `Set.HeaderDeps`, which `writes.Run` computes after `Resolve`; `HandleScopes` instead comes out of `Resolve` (D-1). Copy the proto, mapper, and invalidation parts. The S10-21 plan added the header fields; this PR only reads them.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `effects.Set` gets `HandleScopes [][]byte` *(new)*, filled in `Resolve` with `appendPath(w, table, eff.HandleScope, set.HandleScopes)` beside `Handles`, with a field comment matching `Effect.HandleScope`'s. | The registry already owns the path; `Resolve` is "the whole effect of a batch". | Widening `observation_notes` `Handles` — that recomputes the reconciler for a note, which changes no value. Reading audit scopes back after commit — a second source of truth. |
| D-2 | `WriteEffects.handle_scopes = 7` *(new)*; Swift `CatalogWriteEffects.handleScopes` *(new)*. One name at every layer, Go's `HandleScopes`. | Separate from `handles`: a scope move needs a list refetch, not a detail refetch. | Folding into `handles` — would refetch every open detail page on each note save. |
| D-3 | `invalidations(for:)` adds the three list keys when `touchedHandles` or `handleScopes` is non-empty, once each, in the existing order. `conclusionDetail` stays per `touchedHandles` only. Shape: `if !effects.touchedHandles.isEmpty || !effects.handleScopes.isEmpty { append the three lists }`, then `for entityID in effects.touchedHandles { append conclusionDetail }`. `touchedHandles` does not change, so `WorkspaceSession.apply` and its other readers ignore scope handles. | The detail page does not show the time (§3 Out). | A separate list branch emitting duplicates. |
| D-4 | Header models get `var updatedAt: String = ""` *(new)* only, as the last stored property. Not `updatedRevision`. | The row needs the time; empty means unknown. Nothing sorts by revision here. | Mapping both — a field with no reader. |
| D-5 | `LastUpdatedDisplay` *(new)*, a `struct` in `macos/App/Platform/TimestampFormat.swift` with `let text: String`, `let help: String?`, `let spoken: String`, `let isKnown: Bool` (D-6), `static let columnWidth: CGFloat` (D-11), and `init(_ rfc3339: String, now: Date, calendar: Calendar = .current, locale: Locale = .current)`. It parses once with `TimestampFormat.parse`. Formats use `Date.FormatStyle(locale: locale, calendar: calendar, timeZone: calendar.timeZone)`. `calendar.isDate(date, inSameDayAs: now)` → text `L10n.LastUpdated.today(time:)` with that style's `.hour().minute()`; `calendar.isDate(date, inSameDayAs: calendar.date(byAdding: .day, value: -1, to: now)!)` → `L10n.LastUpdated.yesterday`; otherwise that style's `.day().month(.abbreviated).year()`. A future date on another day uses the date form. The brief's “12 Oct 2026” is the `en_GB` form; the viewer's locale decides the order (`Oct 12, 2026` in `en_US`). | Brief PSL-2, and the brief's inventory homes the text beside the note stamp. One value carries text, help, and spoken form, so S10-41's `SourcesSplitRow` (not a `PVList` row) calls it instead of copying rules out of `Features/Conclusions`. `now`, calendar, and locale are parameters so tests are deterministic. | Text in `TimestampFormat` with the dash, help, and suffix in `ConclusionListRow` — S10-41 would reach into another feature. `RelativeDateTimeFormatter` (“2 days ago”) — the user chose day words then a date. `DateFormatter.doesRelativeDateFormatting` — no time on today, wording varies by OS. |
| D-6 | Empty or unparseable `rfc3339`: `text` is `"—"` (a glyph, not translated), `help` is `nil`, `spoken` is `L10n.LastUpdated.a11yUnknown`. Otherwise `help` is `L10n.LastUpdated.help(stamp: TimestampFormat.boardStamp(rfc3339))` (“Last updated 14 Oct 2026 · 3:15 PM MDT”) and `spoken` is `L10n.LastUpdated.a11yUpdated(text)`. `let isKnown: Bool` *(new)* is `false` for the dash. | PSL-3, PSL-4, PSL-6. The boards name the column in the help, since the column has no header. The note-row stamp already exists. | A new stamp format; an L10n key for a dash. |
| D-7 | `PVListDetailColumn<Item>` *(new)* `struct { var width: CGFloat; var content: (Item) -> PVListDetail? }` and `PVListDetail` *(new)* `struct { var text: String; var help: String?; var icon: PVSymbol?; var faint: Bool = false }`, both in `PVList.swift`. Both `PVList` inits (the primary one and the `Secondary == EmptyView` one, which forwards) get `detail: PVListDetailColumn<Item>? = nil`, placed right after `meta`. When set, every row draws, between the text stack and the trailing ref, a column of `detail.width` holding an `HStack(spacing: 6)` of `PVIcon(icon, size: 14)` (when `icon` is set) and a one-line, caption, non-mono text with `.fixedSize()` (it never truncates; the column frame is fixed and keeps `layoutPriority(1)` like the trailing group, so the title stack gives way), leading-aligned and vertically centered; icon and text are `PVColor.textMuted`, or `PVColor.textFaint` when `faint`; `content` returning `nil` draws an empty column of that width. Help covers the icon and the text and is attached only when non-nil (`if let help { cell.help(help) } else { cell }` in a `@ViewBuilder`, never `.help(help ?? "")`). No `.accessibilityHidden`: the row is already `.accessibilityElement(children: .ignore)`. `PVListSkeleton` gets `detailWidth: CGFloat? = nil` *(new)* and, when set, draws a 64 × 10 bar, leading-aligned, in that column before the existing 72pt ref bar, which stays. | The brief marks the list row Extend. One parameter carries closure and width, so a width with no column cannot be written, and `Components/` holds no board value or time type. | Separate `detail` and `detailWidth` parameters with a default width. A view laid out beside `meta` from `Features/` — the row layout is private to `PVList`. Putting the time in `meta` — the user chose its own column. |
| D-8 | `ConclusionListKind` gets a required `static func updatedAt(_ row: Row) -> String` *(new)*; each kind returns `header.updatedAt`. `ConclusionListRow.updatedColumn<Kind: ConclusionListKind>(_: Kind.Type) -> PVListDetailColumn<Kind.Row>` *(new)* returns width `LastUpdatedDisplay.columnWidth` and content `{ let d = LastUpdatedDisplay(Kind.updatedAt($0), now: Date()); return PVListDetail(text: d.text, help: d.help, icon: .history, faint: !d.isKnown) }`. `PVSymbol.history` *(new)* is `"clock.arrow.circlepath"`, doc comment “Lucide `history` — last updated on a list row”. `ConclusionListRow.labelWithUpdated(_ base: String, _ updated: LastUpdatedDisplay) -> String` *(new)* returns `L10n.Conclusions.a11yList(base, rest: updated.spoken)`. `ConclusionListBody` passes `detail: ConclusionListRow.updatedColumn(Kind.self)` and `itemAccessibilityLabel: { ConclusionListRow.labelWithUpdated(Kind.accessibilityLabel($0, orders: session.nameFormat), LastUpdatedDisplay(Kind.updatedAt($0), now: Date())) }`. `ConclusionListPage` gets `static func skeleton() -> PVListSkeleton` *(new)* returning `PVListSkeleton(detailWidth: LastUpdatedDisplay.columnWidth)`, used at both skeleton sites (`ConclusionListPage.body` fallback and the `.loading` case in `ConclusionListBody`). The Kinds' own `accessibilityLabel` stays without the suffix. | One place for three lists (Events and Places copy Persons); one skeleton builder so the two sites cannot drift. | Per-kind copies; an `accessibilityLabel(_ base: String, …)` overload beside the existing `accessibilityLabel(_:ref:)`, which means title and ref. |
| D-9 | `now` is `Date()` read when the row builds. No timer. | Brief §2.1. The list rebuilds on refetch and navigation. | `TimelineView` — an open list re-rendering every minute for a day boundary. |
| D-10 | The spoken suffix is `LastUpdatedDisplay.spoken`, joined by the existing `L10n.Conclusions.a11yList` in `labelWithUpdated`. | PSL-6 wording, reused by S10-41. | A new joiner. |
| D-11 | `LastUpdatedDisplay.columnWidth` *(new)* `= 136`, the agreed boards' width, icon included. It fits “Today, 12:59 PM”, “Yesterday”, “12 Oct 2026”, and “Sep 30, 2026”. | One width for three lists and the Sources row (PSL-1, EVL-1, PLL-1). | A width per kind; a constant in `Features/Conclusions` that S10-41 cannot reach. |

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `effects.Set.HandleScopes` *(new)* | field | `effects.HandleScopes` func | Extends `Resolve` with the same path the func reads; the func stays for `audit.Record`. |
| `WriteEffects.handle_scopes` *(new)* | proto field | `header_handles` | Different invalidation (D-2). |
| `CatalogWriteEffects.handleScopes` *(new)* | field | `headerHandles` | Same as above. |
| `updatedAt` on three header models *(new)* | field | `CatalogSource.updatedRevision` | Header models had no time; Sources stores a revision for sort. |
| `LastUpdatedDisplay` *(new)* | struct | `TimestampFormat.boardStamp` | The text, help, spoken form, and width of one cell, shared by the three lists and S10-41's `SourcesSplitRow`. Lives beside the stamp it reuses. |
| `L10n.LastUpdated` *(new)* | L10n namespace | `L10n.Conclusions` | Shared with the Sources row in S10-41, so not under Conclusions. |
| `PVListDetail`, `PVListDetailColumn`, `PVList.detail`, `PVListSkeleton.detailWidth` *(new)* | component slot | `PVList.meta` | `meta` is the mono ref; the brief puts the time in its own column. |
| `PVSymbol.history` *(new)* | icon case | `PVSymbol.refresh` (`arrow.clockwise`) | `refresh` means a moved Promote row; the boards draw Lucide `history`. |
| `ConclusionListKind.updatedAt`, `ConclusionListRow.updatedColumn`, `ConclusionListRow.labelWithUpdated`, `ConclusionListPage.skeleton` *(new)* | protocol req / funcs | `ConclusionListRow.title(_:extraCount:)` | Same helper enum; maps the shared value to the kit slot. |
| `api/ffi/handlers/write_effects_test.go` *(new)* | test file | `observations_test.go` | `writeEffectsProto` had no test. |

**Builds on.** The effects registry (`registry.go`), the `WriteEffects` → `CatalogQueryRegistry` invalidation path, `TimestampFormat`, and `PVList`.
**Deletes.** None — nothing becomes dead.
**Next adds.** S10-41 builds `LastUpdatedDisplay` in `SourcesSplitRow` for the text, help, spoken form, and width. S10-35 fills the `PVList` thumbnail closure in `ConclusionListBody` with the likeness.

### Schema

None — no catalog change. S10-21 stored the scopes.

### Go core

| Package | Change |
| --- | --- |
| `core/database/effects` | `Set.HandleScopes [][]byte` *(new)*; `Resolve` appends `eff.HandleScope`. README: one sentence that `Resolve` also returns `HandleScopes`. |

### FFI

`api/proto/engine.proto` `WriteEffects`: `repeated string handle_scopes = 7;` *(new)* with a comment: handles whose last-updated revision this write moved. Handler `api/ffi/handlers/write_effects.go`: `HandleScopes: idStrings(set.HandleScopes)`. Regenerate with `./scripts/generate-proto.sh`.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| List row column | `PVList` (Extend) | `macos/App/DesignSystem/Components/List/PVList.swift` | D-7. Column between the text `VStack` and the trailing `HStack`, `.frame(width: detail.width, alignment: .leading)`. Text truncates first: the column and trailing stack keep `layoutPriority(1)`. |
| Skeleton bar | `PVListSkeleton` (Extend) | same file | D-7. |
| Conclusion page | snowflake | `macos/App/Features/Conclusions/ConclusionListPage.swift` | D-8: `detail:`, the label wrap, and `skeleton()` at both skeleton sites. |
| Cell value | `LastUpdatedDisplay` | `macos/App/Platform/TimestampFormat.swift` | D-5, D-6, D-11. |

Agreed frames (today in the frames is 14 Oct 2026, Mountain Time): **Times** — six rows per list, still in list order (Persons by name, Events by date, Places by name); each column cell is the history icon then “Today, 3:15 PM”, “Today, 9:02 AM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, or a faint “—”; the 136pt column lines up down the list between the text and the mono ref. **Help** — hovering “Yesterday” shows “Last updated 13 Oct 2026 · 4:40 PM MDT”; the dash shows nothing. **Narrow window** — a long title with +2 and a long secondary line truncate; dates, +N, column, ref, and chevron keep their width. **No secondary line** — one-line title, column and ref centered; a second row shows the `en_US` form “Sep 30, 2026”. **First load** — the skeleton has a 64pt bar in the column before the 72pt ref bar. **Empty** — the unchanged empty state; no rows, no column. Persons frames 08–09 (likeness thumb) are S10-35 and S10-39, not this PR. States: empty list → `PVEmptyState`, no column; failed → `PVCallout`; 1000 rows → two `LastUpdatedDisplay` builds per visible row (cell and spoken label, lazy); across midnight the two can differ for one render, which D-9 accepts. Accessibility: D-10.

### Strings

L10n keys *(new)*, namespace `enum LastUpdated` in `macos/App/Platform/L10n.swift` directly after `enum Conclusions`, each with a `comment`:
- `lastUpdated.today` = "Today, %@" — time of a change made today.
- `lastUpdated.yesterday` = "Yesterday".
- `lastUpdated.a11y.updated` = "updated %@".
- `lastUpdated.a11y.unknown` = "last updated unknown".
- `lastUpdated.help` = "Last updated %@" — hover help; the argument is the full stamp.

Accessors, in the style of `L10n.Conclusions.a11yList`: `static func today(time: String) -> String`, `static var yesterday: String`, `static func a11yUpdated(_ text: String) -> String`, `static var a11yUnknown: String`, `static func help(stamp: String) -> String`; each returns `String` through `L10n.string` / `L10n.format`.

Removed: none.

## 7. Steps

**STEP-1 — Resolve returns handle scopes**
Files: `core/database/effects/registry.go`, `core/database/effects/resolve.go`, `core/database/effects/README.md`, `core/database/effects/handle_paths_test.go`.
Do: D-1. Add T-1.
Check: `CGO_ENABLED=1 go test -tags fts5 ./core/database/effects/ ./core/writes/` passes.

**STEP-2 — Scope handles on the wire**
Files: `api/proto/engine.proto`, generated files, `api/ffi/handlers/write_effects.go`, `api/ffi/handlers/write_effects_test.go` *(new)*. Skill: `add-ffi-handler`.
Do: D-2 Go side. Run `./scripts/generate-proto.sh`. Add T-2.
Check: `CGO_ENABLED=1 go test -tags fts5 ./api/...` passes; `git diff --stat` shows the regenerated Go and Swift protobuf files.

**STEP-3 — Lists reload on scope handles**
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

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period, like the rest of `git log` ("Restore list scroll after the sort is applied."). No WIP commits.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Go | `core/database/effects/handle_paths_test.go` `TestHandlePaths` subtest "resolve reports note handle scopes" *(new)* | The `observation_note` create change from "observation note reaches member handle" → call `effects.Resolve(tx, []rowchange.Change{ch})` inline (no new helper); `.HandleScopes` is `{bostonRes.Entity.ID}`, `.Handles` is empty | AC-2.1, D-1 |
| T-2 | Go | `api/ffi/handlers/write_effects_test.go` `TestWriteEffectsProto` *(new)*, table-driven | `Set{}` → all empty; `Set{HandleScopes: [id]}` → `HandleScopes == [uuid string]`, `Handles` empty; `Set{Handles: [a], HeaderDeps: [b]}` → `Handles [a]`, `HeaderHandles [b]`, `HandleScopes` empty | AC-2.1, D-2 |
| T-3 | Swift | `CatalogQueryRegistryTests` `handleScopesReloadListsNotDetails` *(new)* | `handleScopes: ["e1"]` → `[personsList, eventsList, placesList]`; `handles: ["e1"], handleScopes: ["e1","e2"]` → the three lists then `conclusionDetail(e1)` | AC-2.2, AC-2.3, D-3 |
| T-4 | Swift | `TimestampFormatTests` `lastUpdatedNamesTodayYesterdayThenDate` *(new)* | `LastUpdatedDisplay` with the AC-1.1, AC-1.2, AC-1.3 inputs, fixed `now`, a `Calendar(identifier: .gregorian)` whose `timeZone` is `America/Edmonton`, and locales `en_US` and `en_GB` → the `text` values in §2 | AC-1.1, AC-1.2, AC-1.3, D-5 |
| T-5 | Swift | `TimestampFormatTests` `lastUpdatedUnknownIsDashWithoutHelp` *(new)* | `""` and `"garbage"` → text `—`, help `nil`, spoken `last updated unknown`; yesterday's input (Edmonton calendar) → `isKnown`, help starts `Last updated `, spoken `updated Yesterday`; `""` and `"garbage"` → `isKnown == false` | AC-1.4, AC-1.5, D-6 |
| T-6 | Swift | `PersonsListTests` `rowLabelEndsWithUpdated` *(new)* | `ConclusionListRow.labelWithUpdated("Robins, James, PER-7KD45, b. 1817", LastUpdatedDisplay(yesterday, now:, calendar:, locale: en_US))` → ends `, updated Yesterday`; with `""` → ends `, last updated unknown` | AC-1.6, D-8, D-10 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `CatalogQueryRegistryTests` (invalidation order), `TestHandlePaths`, and `PersonsListTests` (row labels).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: core/database/effects/README.md (Resolve returns HandleScopes); macos/App/DesignSystem/README.md (List row detail column)
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with promoted people, events, and places. Open People: each row shows a time in its own column before the ref; the column lines up; the list order is unchanged. Hover a cell: the full stamp appears.
2. Open Events and Places: the same column, text, help, and em dash as People (AC-1.7).
3. Narrow the window: titles truncate; the column and ref keep their width.
4. Quit and reopen the project, open People: the loading skeleton shows a column bar, and rows do not shift when the times arrive (AC-1.8).
5. Open a Source, edit only the note on an Observation of a Subject filed to a Person, save. Open People: that row reads `Today, {now}` (AC-2.4).
6. Edit an Observation's value on an Event's member Subject. Open Events: that row moved.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| Lists refetch on every composer note save | One read per list, only when a scope handle is reported; lists not on screen refetch on next show per their spec. |
| Locale strings differ from the examples | Tests pin locale and time zone (T-4). |
| `PVList` change alters lists that do not pass `detail` | The column draws only when `detail` is set (D-7); the existing `#Preview` stays unchanged. |

**Stop and report if:**

- `effects.Resolve` or `writeEffectsProto` no longer match §4.
- The agreed board places the time somewhere other than its own column before the ref, or changes PSL-2's wording.
- `ListPersonHeaders`, `ListEventHeaders`, or `ListPlaceHeaders` return empty `updated_at` for a promoted handle in the manual check (S10-21 regression).
- A note-only save in the app still leaves the row's time unchanged after step 5 of the manual check.
- Field number 7 on `WriteEffects` is taken at your base.

## 12. Open questions

None.

## 13. PR

**Title:** S10-22 — Last updated on People, Events, and Places
**Body:** objective in one sentence; what shipped against each US; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-10/prs/S10-22-list-last-updated.md`.
