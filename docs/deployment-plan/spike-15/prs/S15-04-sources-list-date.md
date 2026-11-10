# S15-04 — The Sources list shows and sorts by the date

**Status:** Blocked — on board S15-D2
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D2](../design/S15-D2-sources-list.md), board not drawn yet
**Board:** — (not drawn yet). When agreed: a Claude Design artifact link, read with the Artifact tool's `read` (user, 2026-11-10); the board's SL-3 give-way order is copied into D-6
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)), S15-03 (landed [#766](https://github.com/mendahu/provenencia/pull/766)), S15-D2
**Blocks:** —
**Branch:** `s15-04-sources-list-date`
**Version bump:** minor, the next minor at merge (0.6.0 → 0.7.0 at the planning base) — the date on list rows and two new sorts ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "The Source date is one optional DateValue on `sources`", "One date order", "Not public is a stored switch and nothing else" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`../source-provenance-date.md`](../source-provenance-date.md); [`../../../source-layer-data-model.md`](../../../source-layer-data-model.md) § 4 (`date_sort_key`); [`../../../structured-date-model.md`](../../../structured-date-model.md) (the one date order); S15-02's [plan](S15-02-source-date-and-public-flag.md) D-5, D-13; S15-03's [plan](S15-03-source-page-date-and-private-switch.md) D-13 and **Next adds**
**Planned:** 2026-11-10 against `main` @ `16efd08`
**Review:** [`S15-04-sources-list-date.review.md`](S15-04-sources-list-date.review.md) — round 1: Fits with changes / Ready with changes; every finding Fixed; no open blocker

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

A researcher sees each Source's own date on its Sources list row and sorts the list by it, oldest or newest first, with undated Sources last. This is the list half of dogfood bar 1 ("The Sources list shows it on each row and sorts by it, undated last"). The date and its place in date order already reach Swift's wire model from S15-02; this PR maps the order key, draws the date on the row and adds two sorts. Swift never orders dates itself (working rule "One date order").

**Why now.** S15-02 and S15-03 landed: Sources carry a date and researchers are setting it on the Source page, but the list neither shows nor uses it.

**What stays.** The four existing sorts and their order, the filters, search, the caption band and columns, the graph cells, Copy Source on right-click, and the settings file's shape.

## 2. User stories

**US-1.** As a researcher, when I scan the Sources list, I see each Source's date on its row so that I can tell my records apart by when they were made.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | A Source dated `abt 12 Mar 1880` | Its row's second line reads `SRC-…` · calendar icon `abt 12 Mar 1880` · last update; the text is `DateRowDisplay.line(date:)` of the Source's date |
| AC-1.2 | An undated Source, or one whose date gives an empty line | Its second line reads `SRC-…` · calendar icon `Undated` (faint tone) · last update |
| AC-1.3 | A Source whose date is only the phrase "in the reign of Victoria" | The row reads "in the reign of Victoria" |
| AC-1.4 | A Source dated on its page while the list is cached | Back on the list, its row shows the new date (the `.updatedSource` patch, no reload) |
| AC-1.5 | VoiceOver on a dated row's link | The value reads "<title>, <type>, <ref>, 1 Jun 1880, updated …"; undated reads "…, <ref>, Undated, updated …" |

**US-2.** As a researcher, I sort the Sources list by the Source date, oldest or newest first, so that I can work through my records in time order; undated Sources wait at the end.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Sources with keys from `bef. 1830` (k1) < `1830` (k2) < `1880` (k3) < `aft. 1880` (k4), and two undated "Bravo", "Alpha"; sort "Source date, oldest first" | Order: k1, k2, k3, k4, Alpha, Bravo |
| AC-2.2 | The same; "Source date, newest first" | Order: k4, k3, k2, k1, Alpha, Bravo |
| AC-2.3 | Two Sources with equal keys, "Zeta" and "Eta"; either date sort | Eta before Zeta |
| AC-2.4 | A phrase-only Source (date set, key absent) titled "Mill notes", and an undated "Abbey roll" | Both sit in the undated tail: Abbey roll, Mill notes |
| AC-2.5 | Choose "Source date, newest first", leave the list, reopen the project | The sort is still "Source date, newest first"; the control reads "Sorted by source date, newest first" |
| AC-2.6 | A settings file written by this build with `"sort":"dateNewest"`, read by a build without the case | Falls back to Date added (existing per-field fallback; no code change, stated for the reviewer) |
| AC-2.7 | Date sort chosen and a type filter applied | The filtered rows keep the date order |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CatalogSource.dateSortKey` mapped from `Source.date_sort_key` | STEP-1, AC-2.1 |
| Two sorts, labels, order and tie rule (SL-4, SL-5, SL-6) | STEP-2, STEP-3, AC-2.1 – AC-2.7 |
| The date and "Undated" on the row's second line, truncation and tooltip (SL-1, SL-2, SL-3) | STEP-4, AC-1.1 – AC-1.4 |
| Row accessibility value with the date (SL-7) | STEP-4, AC-1.5 |
| Regular, narrow and Wide arrangement (SL-W1, SL-W2) | STEP-4, manual check 4 |

| Out | Home |
| --- | --- |
| Filtering by date or by Private | not planned (map scope entry "Not this PR") |
| A Date column in the caption band | not planned (user, 2026-11-10: second line) |
| The list beside a selected Source at Wide | [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md) |
| The Person page's Sources section date | S15-05 |
| Any Go change: the key is already on the wire | not needed (S15-02 D-5) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Wire | `api/proto/engine.proto` `Source.date_sort_key` (optional int64, 15); `api/ffi/handlers/sources.go` `sourceProto` (line ~719) fills it from `sources.DateSortKey` | Absent when undated or phrase-only. Generated Swift: `macos/App/Platform/Generated/engine.pb.swift` `dateSortKey`, `hasDateSortKey` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogSource` | Has `date: CatalogDateValueInput?`, `isPublic`; no sort key (S15-03 D-13 left it to this PR) |
| Mapping | `macos/App/Platform/GoStore.swift` `mapSource` | Maps `date` when `s.hasDate`; ignores `dateSortKey` |
| Fake | `macos/App/Platform/FakeStore.swift` `setSourceDate` → `mutateSource` | Sets `date`; knows no key |
| Sorts | `macos/App/Features/Sources/SourcesModel.swift` `Sort` (`added`, `updated`, `az`, `za`), `buildVisibleSources` switch | No date case. `visibleSources` memo token includes the sources handle revision, so a patched row re-sorts |
| Settings | `macos/App/Features/Sources/SourcesListSettings.swift` `init(from:)` | Unknown `sort` raw value falls back to `.added` |
| Row | `macos/App/Features/Sources/SourcesListRow.swift` `SourcesListRow.sourceLink`, `SourcesLinkLabel.secondLine` | Second line: (type, compact) · ref · `PVListDetailCell` last update; `lineLimit(1)`. Accessibility value: `PVLastUpdated.label("title, type, ref", …)` |
| Date text | `macos/App/Platform/DateRowDisplay.swift` `DateRowDisplay.line(date:)`; `macos/App/Features/Sources/SourcePageIdentityHeader.swift` `sourceDate` | The header shows the same text with `PVIcon(.calendar, size: 14)`, `.lineLimit(1)`, `.help(line)` |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Sources.sortAdded` … `sortZA`, `sortedBy`; `macos/App/Resources/Localizable.xcstrings` | Keys `sources.list.sort.*` |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` `sortOrdersByAddedUpdatedAndTitle`, `updatedSourcePatchesListRow`, `settingsRestoreOnNewModel`; `macos/ProvenenciaTests/SourcesListSettingsTests.swift` `unknownFieldValuesFallBack` | Cover the four sorts, the row patch and settings fallback |

**Precedent.** The `az` / `za` pair in `SourcesModel.Sort` and its test `sortOrdersByAddedUpdatedAndTitle`: copy the shape (one case per direction, one label each, one test asserting the order). The header's `sourceDate` in `SourcePageIdentityHeader.swift` (S15-03, #766): copy the icon + `DateRowDisplay.line` + `.help` pattern at the row's caption size.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `CatalogSource` gains `var dateSortKey: Int64? = nil`, declared after `date`; `GoStore.mapSource` sets it to `s.dateSortKey` only when `s.hasDateSortKey`, and drops `private` (`static func mapSource`, as `mapObservation` is) so T-7 can call it | The wire key is the only order Swift uses (rule "One date order") | Deriving a key from `date` in Swift — a second date order |
| D-2 | `Sort` gains `case dateOldest`, `case dateNewest` (raw values `"dateOldest"`, `"dateNewest"`), after `za` in `allCases` | User, 2026-11-10: two options, like A–Z / Z–A; appended so menu order matches SL-4 | One option with a direction toggle — no such control in the toolbar |
| D-3 | Order: rows with a key first, by key (ascending for `dateOldest`, descending for `dateNewest`); equal keys, and the rows without a key after them, by one private `static func titleOrder(_ a: CatalogSource, _ b: CatalogSource) -> Bool` *(new)* in `SourcesModel` (`title.localizedCaseInsensitiveCompare` ascending, then `id` descending), the same in both directions. `.az` and `.za` are not changed | "Undated last" from the working rule; a stable, readable tie rule; the undated tail does not flip with direction (SL-5) | Ties by added order — not visible to the researcher |
| D-4 | A phrase-only date (date set, key absent) sorts in the undated tail and shows its phrase on the row | The engine gives it no place (structured-date-model); the row still shows what was typed | Hiding the phrase |
| D-5 | The row date is a `PVListDetailCell`, as the last update beside it is. `PVListDetail` *(extended)* gains `var truncates: Bool = false`; when true, `PVListDetailCell` gives its `Text` `.lineLimit(1).truncationMode(.tail)` instead of `.fixedSize()`; every existing caller keeps the default. `SourcesListRow` gains internal `static func dateDetail(_ date: CatalogDateValueInput?) -> PVListDetail` *(new)*: with `line = date.map { DateRowDisplay.line(date: $0) } ?? ""`, a non-empty line gives `PVListDetail(text: line, help: line, icon: .calendar, faint: false, truncates: true)`; an empty line (undated, or a set date whose line is empty, such as a range missing a side) gives `PVListDetail(text: L10n.string(L10n.Dates.rowUndated), help: nil, icon: .calendar, faint: true, truncates: true)`. `SourcesLinkLabel` takes `date: CatalogDateValueInput?` and draws `PVListDetailCell(detail: SourcesListRow.dateDetail(date), width: nil, iconSize: 12)` | Brief SL-1 – SL-3; one cell draws every fact on the line, so spacing, icon size, tones and tooltip cannot drift; S15-05 reuses the builder | A hand-built icon + `Text` (a second copy of the cell); a helper on the `private` `SourcesLinkLabel`, which tests cannot reach |
| D-6 | Second-line order: type (compact) · ref · date · last update, each separated by the existing `separator`. The date cell gets `.layoutPriority(0.5)` at its call site (outside the cell's own 1); the last update keeps the cell's 1; the type keeps 0; the ref keeps `fixedSize`. So the type truncates first, then the date (brief SL-3 recommendation; the agreed board may reorder these and nothing else) | Brief SL-3 | Wrapping the line — the row height is fixed to the graph slot |
| D-7 | Accessibility: `SourcesListRow` gains internal `static func spokenBase(title: String, type: String, ref: String, date: CatalogDateValueInput?) -> String` *(new)* = `L10n.DesignSystem.spokenJoin([title, type, ref, dateDetail(date).text]) ?? title`; `sourceLink` passes it to `PVLastUpdated.label` in place of today's interpolated string | SL-7; one spoken join | A separate accessibility element for the date |
| D-8 | `FakeStore` gains `var nextSourceDateSortKey: Int64?` *(new)*, default nil; `setSourceDate` sets the row's `dateSortKey` to it. Seeded rows set keys through the test helper `source(...)`, which STEP-3 extends with `date:` and `dateSortKey:` | `.cursor/rules/swift-tests.mdc`: the fake scripts what Go composes and does not re-implement it; a set date never keeps an old key | A cleared-means-nil rule or a key formula in the fake |
| D-9 | No change to `SourcesListSettings` decoding | The existing per-field fallback covers an unknown raw value (AC-2.6) | A migration of the settings file |

**Rules this plan bends.** None.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogSource.dateSortKey` *(new)* | field | `CatalogSource.date`, `updated.revision` | Extends the model; the key is not derivable in Swift |
| `SourcesModel.Sort.dateOldest`, `.dateNewest` *(new)* | enum cases | `.az`, `.za` | Extends the enum |
| `L10n.Sources.sortDateOldest`, `sortDateNewest` *(new)*; `L10n.Dates.rowUndated` *(new)* | L10n keys | `L10n.Sources.sortAZ`; `L10n.Dates` row strings; `L10n.Conclusions.timelineUndated` | Sorts sit beside the other sorts; "Undated" is a row date value and sits beside `DateRowDisplay`'s strings for S15-05; `timelineUndated` is the Timeline's heading and its copy may change on its own |
| `PVListDetail.truncates` *(new)* | kit field | `PVListDetailCell` `.fixedSize()` | Extends the kit cell additively; the only gap was truncation (SL-3) |
| `SourcesListRow.dateDetail`, `spokenBase` *(new)* | static funcs | `PVLastUpdated.detail(_:in:)`, `PVLastUpdated.label`; `SourcePageIdentityHeader.sourceDate` | Same shape as `PVLastUpdated.detail`, in the feature until S15-05 is a second caller; the header keeps its own size, pencil and focus |
| `FakeStore.nextSourceDateSortKey` *(new)* | test hook | `FakeStore.nextWriteEffects` | The fake's scripting shape |
| `SourcesModel.titleOrder` *(new)* | private func | `buildVisibleSources` `.az` inline compare | One tie rule for both date sorts |

**Builds on.** `PVListDetailCell` (the line's one cell), `DateRowDisplay` (one date text), `L10n.DesignSystem.spokenJoin`, the wire key from `sources.DateSortKey`, the `Sort` enum and settings store, the `.updatedSource` row patch.
**Deletes.** None — nothing becomes dead.
**Next adds.** S15-05 shows the date on the Person page's Sources section with the same `dateDetail` (promote it to a recipe then).

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source with a date reaches the list row | `SourcesLinkLabel` | Dated, phrase-only and undated each render (AC-1.1 – AC-1.3) |
| A stored settings `sort` this build adds | `SourcesListSettings.init(from:)` | Older builds fall back to `.added` (AC-2.6) |

### Scale

At the map's Scale (3,000 Sources): the sort adds one `Int64?` comparison per pair, inside the existing memoized `buildVisibleSources`, which already sorts by title for A–Z. The row computes `DateRowDisplay.line` once per drawn row inside a `LazyVStack`. T-5 runs the date sort over 3,000 rows and asserts the order and one build; no timing assertion (the cost is one optional-integer comparison per pair beside the title compare A–Z already does).

### Edges

- **Assumptions.** The key is a total order over dated Sources and equal keys mean equal places (S15-02 T-10). Ties fall to title, then id, so the sort is deterministic (T-2).
- **Neighbors.** `abt 1880` and `1880` share a key and tie to title; `bef. 1880` before `1880` before `aft. 1880` (T-2).
- **Removed inputs.** A cleared date drops the key and the row moves to the tail (T-3).
- **Real data.** None — no table or seed drives the rule.
- **Untrusted size.** None — rows come from the app's own catalog read, already bounded by the list.

### Schema

None — no catalog change.

### Go core

None — `sources.DateSortKey` and the wire field exist (S15-02).

### FFI

None — `Source.date_sort_key` is on the wire.

### macOS

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Row date | `PVListDetailCell` (Extend: `truncates`) | `macos/App/DesignSystem/Components/List/PVList.swift`; builder in `macos/App/Features/Sources/SourcesListRow.swift` | D-5, D-6; brief SL-1 – SL-3 |
| Sort options | `PVSelect` | `SourcesListView.sortMenu` | Options come from `Sort.allCases`; no view change beyond the new cases |
| Columns | `SourcesListColumns` | unchanged | Wide: no code change. The list already fills the pane (the Source cell is `maxWidth: .infinity`); manual check 4 confirms it |

States: undated, or a set date with an empty line → "Undated" (faint); phrase-only → phrase; long → truncated with tooltip; empty and filter-empty unchanged. Accessibility: D-7 (the row link merges its children, so the cell has no identifier of its own); sort options keep `sources.sort.<raw>` identifiers.

### Strings

L10n keys *(new)*, in `L10n.swift` and `Localizable.xcstrings`:
- `L10n.Sources.sortDateOldest`: `sources.list.sort.dateOldest` = "Source date, oldest first"
- `L10n.Sources.sortDateNewest`: `sources.list.sort.dateNewest` = "Source date, newest first"
- `L10n.Dates.rowUndated`: `dates.row.undated` = "Undated"

`sortedBy` lowercases the label: "Sorted by source date, oldest first". Removed: none.

## 7. Steps

**STEP-1 — Map the key.**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourceMappingTests.swift` *(new)*.
Do: D-1, D-8. Add T-7.
Check: T-7 passes; existing `SourcesModelTests` pass.

**STEP-2 — Strings.** Skill: `add-localized-string`.
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: the three keys in §6 Strings.
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The two sorts.** Skill: `add-swift-test`.
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`, `macos/ProvenenciaTests/SourcesListSettingsTests.swift`.
Do: D-2, D-3, D-4; labels in `Sort.label`. Extend the test helper `source(id:title:typeID:ref:updatedRevision:)` with `date: CatalogDateValueInput? = nil, dateSortKey: Int64? = nil`. Add T-1 – T-5.
Check: T-1 – T-5 pass.

**STEP-4 — The row.** Skill: `add-ui-component`.
Files: `macos/App/DesignSystem/Components/List/PVList.swift`, `macos/App/Features/Sources/SourcesListRow.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`.
Do: D-5, D-6, D-7, with the board's SL-3 order. Add T-6. No Wide code (§6 macOS Columns).
Check: T-6 passes; the app shows manual checks 1 – 4.

**STEP-5 — Version.**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` (four at the planning base).
Do: the minor above `VERSION` on `main` when you branch; if `main` bumps again before merge, rebase and take the next one.
Check: `VERSION`, `core.Version` and every `MARKETING_VERSION` line agree.

**STEP-6 — Docs.**
Files: `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S15-D2-sources-list.md` → `../design/archive/` *(new)*, `../design/README.md`.
Do: per the Definition of done. Create the folder `archive` *(new)* under `../design/` and `git mv` the brief into it; in `../design/README.md` move the S15-D2 row from Scheduled to Archived, its link pointing into that folder, its Board cell the board link, its Built by cell "S15-04 (#N)". `completed.md`: one paragraph at the top in the shape of the S15-03 entry (bold title, PR link, what a researcher can now do, product version).
Check: `python3 scripts/check-plan-docs.py` on this plan passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests.swift` `sortByDateOldestPutsUndatedLast` *(new)* | `s1` "Delta" key 100, `s2` "Echo" key 200, `s3` "Foxtrot" key 300, `s4` "Golf" key 400, `s5` "Bravo" no date, `s6` "Alpha" no date; `.dateOldest` → s1, s2, s3, s4, s6, s5 | AC-2.1, D-3 |
| T-2 | Swift | `SourcesModelTests.swift` `sortByDateNewestKeepsUndatedTailAndTies` *(new)* | T-1's six plus `s7` "Zeta" key 500, `s8` "Eta" key 500, `s9` "Mill notes" with `date: CatalogDateValueInput(kind: "point", phrase: "in the reign of Victoria")` and no key, `s10` "Abbey roll" no date; `.dateNewest` → s8, s7, s4, s3, s2, s1, s10, s6, s5, s9; `.dateOldest` → s1, s2, s3, s4, s8, s7, s10, s6, s5, s9 | AC-2.2, AC-2.3, AC-2.4, D-3, D-4 |
| T-3 | Swift | `SourcesModelTests.swift` `clearedDateMovesRowToUndatedTail` *(new)* | T-1's six, `.dateOldest`; `nextSourceDateSortKey` nil, `store.setSourceDate(…, sourceID: "s1", date: nil)`, apply its result with `.updatedSource` → order s2, s3, s4, s6, s5, s1. Then `nextSourceDateSortKey = 450`, set s5's date to `1 Jun 1880` and apply → s2, s3, s4, s5, s6, s1 | AC-1.4, D-8 |
| T-4 | Swift | `SourcesListSettingsTests.swift` `dateSortRoundTrips` *(new)*; `SourcesModelTests.swift` `sortControlLabelForDateSort` *(new)* | Save `.dateNewest`, load → `.dateNewest`; `sortControlLabel` → "Sorted by source date, newest first". AC-2.6 stays covered by the existing `unknownFieldValuesFallBack` | AC-2.5 |
| T-5 | Swift | `SourcesModelTests.swift` `dateSortAtScaleWithFilter` *(new)* | 3,000 Sources `s0000` … `s2999`, title "Source NNNN"; type `t1` for even N, `t2` for odd; each N divisible by 30 dated with key N (100 dated, 50 of them even); type filter `t1`, `.dateOldest` → 1,500 rows: the 50 dated even rows by key, then the 1,450 undated even rows by title; `visibleSourcesBuildCount` rises by 1 | AC-2.7, §6 Scale |
| T-6 | Swift | `SourcesModelTests.swift` `rowDateDetailAndSpokenBase` *(new)* | `SourcesListRow.dateDetail`: `1 Jun 1880` point → text "1 Jun 1880", help "1 Jun 1880", not faint, truncates; phrase-only → text "in the reign of Victoria"; nil → "Undated", faint, no help; `CatalogDateValueInput(kind: "range", startYear: 1834)` with no end, if its line is empty → "Undated". `spokenBase(title: "T", type: "Census", ref: "SRC-1", date: nil)` → "T, Census, SRC-1, Undated" | AC-1.1, AC-1.2, AC-1.3, AC-1.5, D-5, D-7 |
| T-7 | Swift | `SourceMappingTests.swift` `mapsDateSortKey` *(new)* | `Provenencia_Engine_V1_Source` with `dateSortKey = 42` → `GoStore.mapSource(_:).dateSortKey == 42`; without it → nil | D-1 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `sortOrdersByAddedUpdatedAndTitle` (if `allCases` order is asserted elsewhere), `listColumnsMatchBoard`.

## 9. Definition of done

```
- [ ] Every AC in §2 is true: test name or manual check line for each
- [ ] Every T-# in §8 exists and passes
- [ ] scripts/check-pr.sh prints check-pr: PASS (gofmt, go vet, go test, Xcode tests with no Swift warning, catalog and plan-doc checks). The Xcode tests and the manual check need macOS: without it, report those boxes "not run (no macOS)"; never mark them done
- [ ] No Out line was implemented; no code outside §4 / §6 was refactored
- [ ] New user-facing copy is in L10n + Localizable.xcstrings; no stale strings left
- [ ] Sweep: each renamed or deleted name `rg`'d by its old, qualified name across `.agents/`, `.cursor/rules/`, `docs/*.md` and `L10n.swift` with no hit left; `scripts/check-guidance-names.py` passes (check-pr runs it)
- [ ] Every row of **Rules this plan bends** has the user's answer, and its rule is updated
- [ ] UI matches the brief inventory; no new kit primitive unless the brief marks it New
- [ ] Version bump as the header says (VERSION, core/version.go, MARKETING_VERSION together) — or none
- [ ] This plan: Status → Landed in #N (date); a Landed note lists every difference from the plan
- [ ] deployment-plan.md: map row status; completed.md: one short paragraph of what shipped
- [ ] Brief moved to design/archive/ and design/README.md updated (if gated)
- [ ] Domain doc updated if a rule in docs/*.md changed: none expected
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with at least six Sources. On three Source pages set dates `1 Jun 1880`, `bet 1834 and 1840`, `bef 3 Mar 1830`; on one set the phrase "in the reign of Victoria"; leave the rest undated. Go to Sources: each row's second line shows its date or "Undated" (AC-1.1 – AC-1.4).
2. Sort "Source date, oldest first": `bef 3 Mar 1830`, `bet 1834 and 1840`, `1 Jun 1880`, then the phrase-only and undated rows by title. Sort "newest first": the dated three reversed, the same tail (AC-2.1 – AC-2.4).
3. Quit and reopen the project: the sort holds (AC-2.5). Apply a type filter: the date order holds (AC-2.7).
4. Narrow the window below 880pt list width: the type joins the second line; a long range truncates and shows its tooltip. Widen past 1600pt pane: the list fills the width; scroll and sort are kept (SL-W1, SL-W2). Compare each with the board's frames.
5. VoiceOver on a dated and an undated row reads AC-1.5.

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The second line overflows with a long title type and date in a compact row | D-6 priorities; manual check 4 |
| A Swift test helper constructs `CatalogSource` positionally and breaks on the new field | The field has a default (D-1) and sits after `date` |

**Stop and report if:**

- `Provenencia_Engine_V1_Source.hasDateSortKey` does not exist in `engine.pb.swift` at your base.
- The agreed board puts the date anywhere other than the second line, or adds a Date column (the user chose the second line on 2026-11-10).
- The board changes the sort labels or the undated copy from §6 Strings.
- At Wide (pane ≥ 1600pt) the list does not fill the pane between the gutters: Wide needs code this plan does not name.
- Adding `truncates` changes how any existing `PVListDetailCell` caller draws (a Conclusion list snapshot or test moves).

## 12. Open questions

- **The board.** S15-D2 is not drawn. Recommended: paste the brief into Claude Design, agree the frames, record the link; then copy the board's SL-3 order into D-6 and move this plan to Planned.

## 13. PR

**Title:** S15-04 — The Sources list shows and sorts by the date
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-04-sources-list-date.md`.
