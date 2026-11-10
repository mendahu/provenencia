# S15-04 — The Sources list shows and sorts by the date

**Status:** Landed in #N (2026-11-10)
**Spike / slice:** Spike 15 · Slice 1 — Source date, privacy, credibility
**Map row:** [`../deployment-plan.md`](../deployment-plan.md#pr-map)
**Brief:** [S15-D2](../design/archive/S15-D2-sources-list.md), board agreed 2026-11-10
**Board:** [Claude Design — Sources List](https://claude.ai/artifact/JDm51QdorUvY79EFoci4cm), page *S2-04 Sources list*, turn "S15-D2 · Sources list": frames 1a/1b, 2a/2b, 3, 4, 5, 6 and the panel "S15-D2 · Requirements as drawn"; the row and list code is the component *Sources List Frame* (presets `s15added`, `s15asc`, `s15sortmenu`, `s15desc`, `s15narrow`, `s15census`). Read it with the Artifact tool's `read` on that link, `path` "project/S2-04 Sources list.dc.html" (frames and requirements) and "project/Sources List Frame.dc.html" (the row). §6 macOS describes each frame in words
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)), S15-03 (landed [#766](https://github.com/mendahu/provenencia/pull/766)), S15-D2
**Blocks:** —
**Branch:** `s15-04-sources-list-date`
**Version bump:** minor, the next minor at merge (0.6.0 → 0.7.0 at the planning base) — the date on list rows, two new sorts and a Wide arrangement ([`docs/versioning.md`](../../../versioning.md))
**Skills to load:** [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md), [`add-localized-string`](../../../../.agents/skills/add-localized-string/SKILL.md), [`add-swift-test`](../../../../.agents/skills/add-swift-test/SKILL.md)
**Rules that bind:** `.cursor/rules/macos-client.mdc`, `.cursor/rules/versioning.mdc`
**Context:** working rules "The Source date is one optional DateValue on `sources`", "One date order", "Not public is a stored switch and nothing else" ([`../deployment-plan.md`](../deployment-plan.md#working-rules)); [`../source-provenance-date.md`](../source-provenance-date.md); [`../../../source-layer-data-model.md`](../../../source-layer-data-model.md) § 4 (`date_sort_key`); [`../../../structured-date-model.md`](../../../structured-date-model.md) (the one date order); S15-02's [plan](S15-02-source-date-and-public-flag.md) D-5, D-13; S15-03's [plan](S15-03-source-page-date-and-private-switch.md) D-13 and **Next adds**; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide) (rule 3, rewritten 2026-11-10)
**Planned:** 2026-11-10 against `main` @ `16efd08`
**Review:** [`S15-04-sources-list-date.review.md`](S15-04-sources-list-date.review.md) — Fits with changes / Ready with changes, 2 rounds; every blocker Fixed or answered by the user (2026-11-10); nothing open

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

## Landed

- `SourcesListContent` keeps `listWidth` in `@State` and reads `columns` through a computed property built from it and `\.pvWidthClass`, instead of a `let columns` in `body`; every reader still gets the one value.
- `pvListBand` is documented in the design system README's Page columns row (the kit's inventory of its modifiers).
- The comments on the D-14 strings in `L10n.swift` and `Localizable.xcstrings` now say "published date dialog" along with the values.
- Not run: the manual check (§ 10) and the board comparison of frames 1a, 1b, 2b and 5 (this build was verified by the Xcode tests only).

## 1. Objective

A researcher sees each Source's own date on its Sources list row, named as its published date and kept apart from the app's last update, and sorts the list by it, oldest or newest first, with undated Sources last. At Wide the list stops stretching: it centers at 1600pt of content and gives the date its own column. This is the list half of dogfood bar 1 ("The Sources list shows it on each row and sorts by it, undated last"). The date and its place in date order already reach Swift's wire model from S15-02; this PR maps the order key, draws the date, adds two sorts and the Wide arrangement. Swift never orders dates itself (working rule "One date order").

**Why now.** S15-02 and S15-03 landed: Sources carry a date and researchers are setting it on the Source page, but the list neither shows nor uses it.

**What stays.** The four existing sorts and their order, the Source page's layout (only its date copy changes, D-14), the filters, search, the Source link and its activity line (ref · last update), the graph cells, Copy Source on right-click, and the settings file's shape.

## 2. User stories

**US-1.** As a researcher, when I scan the Sources list, I see each Source's published date on its row so that I can tell my records apart by when they were made.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-1.1 | Regular (list 1000pt), a Source dated `abt 12 Mar 1880` | The Type cell reads the type, then under it a muted "Published" and `abt 12 Mar 1880` in secondary ink; help "Published abt 12 Mar 1880"; the Source cell's second line is ref · last update as today |
| AC-1.2 | Regular, an undated Source, or one whose date gives an empty line | Under the type, a faint "No published date" |
| AC-1.3 | A Source whose date is only the phrase "in the reign of Victoria" | Its date reads "in the reign of Victoria" |
| AC-1.4 | Wide (pane 2296pt) | The list is one centered column at most 1600pt of content plus gutters, hairlines at its sides; columns Source · Type (150) · Published (150) · Evidence graph (420); the caption band reads SOURCE · TYPE · PUBLISHED · EVIDENCE GRAPH; a dated row's Published cell shows the date alone (secondary), an undated one "Undated" (faint) |
| AC-1.5 | Narrow (list 860pt) | Under the title: the record line `type · Published <date>` (or `type · No published date`), then the activity line ref · last update; the type truncates first, then the date; the ref never truncates |
| AC-1.6 | A Source dated on its page while the list is cached | Back on the list, its row shows the new date (the `.updatedSource` patch, no reload) |
| AC-1.8 | The Source page header of a dated Source; of an undated Source; its date dialog | VoiceOver reads "Published, 1 Jun 1880"; the button reads "Add published date"; the pencil's label is "Edit published date"; the dialog is titled "Published date" |
| AC-1.7 | VoiceOver on a dated row's link | The value reads "<title>, <type>, published 1 Jun 1880, <ref>, updated …"; undated reads "…, <type>, no published date, <ref>, updated …" |

**US-2.** As a researcher, I sort the Sources list by the published date, oldest or newest first, so that I can work through my records in time order; undated Sources wait at the end.

| AC | Given / when | Then |
| --- | --- | --- |
| AC-2.1 | Sources with keys from `bef. 1830` (k1) < `1830` (k2) < `1880` (k3) < `aft. 1880` (k4), and two undated "Bravo", "Alpha"; sort "Published date, oldest first" | Order: k1, k2, k3, k4, Alpha, Bravo |
| AC-2.2 | The same; "Published date, newest first" | Order: k4, k3, k2, k1, Alpha, Bravo |
| AC-2.3 | Two Sources with equal keys, "Zeta" and "Eta"; either date sort | Eta before Zeta |
| AC-2.4 | A phrase-only Source (date set, key absent) titled "Mill notes", and an undated "Abbey roll" | Both sit in the undated tail: Abbey roll, Mill notes |
| AC-2.5 | Choose "Published date, newest first", leave the list, reopen the project | The sort is still "Published date, newest first"; the control reads "Sorted by published date, newest first" |
| AC-2.6 | A settings file written by this build with `"sort":"dateNewest"`, read by a build without the case | Falls back to Date added (existing per-field fallback; no code change, stated for the reviewer) |
| AC-2.7 | Date sort chosen and a type filter applied | The filtered rows keep the date order |

## 3. Scope

| In | Covered by |
| --- | --- |
| `CatalogSource.dateSortKey` mapped from `Source.date_sort_key` | STEP-1, AC-2.1 |
| Two sorts, labels, order and tie rule (SL-4, SL-5, SL-6) | STEP-2, STEP-3, AC-2.1 – AC-2.7 |
| The published date under the type, in the Published column and on the narrow record line; undated; truncation and help (SL-1, SL-2, SL-3) | STEP-4, STEP-5, AC-1.1 – AC-1.6 |
| Row accessibility value with the date (SL-7) | STEP-5, AC-1.7 |
| Wide: centered at `widthListWideMax`, side hairlines, Published caption (SL-W1); Regular and narrow (SL-W2) | STEP-4, STEP-6, AC-1.4, AC-1.5 |
| `PVSpacing.widthListWideMax` and the kit modifier `pvListBand`; Promote's list adopts the token (1600 of content at Wide) | STEP-4, STEP-6 |
| "Published" as the date's one name: the sorts, the Source page's date copy (D-14) | STEP-2, AC-1.8 |
| Bug (user, 2026-11-10): the vertical hairline between the Source/Type cells and the Evidence graph column stops at the bottom of the graph stack when a wrapped title makes the row taller; it spans the full row | STEP-5a |

| Out | Home |
| --- | --- |
| Filtering by date or by Private | not planned (map scope entry "Not this PR") |
| The list beside a selected Source at Wide | [`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md) |
| Centering the other lists (Persons, Events, Places) at `widthListWideMax` | their redesigns ([`docs/ideas/wide-layout.md`](../../../ideas/wide-layout.md)) |
| The Person page's Sources section date | S15-05 |
| Any Go change: the key is already on the wire | not needed (S15-02 D-5) |
| `.az` and `.za` reading `titleOrder` | not planned (D-3) |
| The date in the list's search keys (`SourcesModel.searchKeys`) | not planned |
| Truncating titles to one line at Wide | not planned: the board's "titles stay on one line" is the effect of the width, not a line limit |
| Rewriting Wide rule 3 in `docs/macos-client-patterns.md` § 9 and the `add-design-brief` Paste block | done in the planning PR [#770](https://github.com/mendahu/provenencia/pull/770) |

## 4. Current state

| Area | Where | Today |
| --- | --- | --- |
| Wire | `api/proto/engine.proto` `Source.date_sort_key` (optional int64, 15); `api/ffi/handlers/sources.go` `sourceProto` (line ~719) fills it from `sources.DateSortKey` | Absent when undated or phrase-only. Generated Swift: `macos/App/Platform/Generated/engine.pb.swift` `dateSortKey`, `hasDateSortKey` |
| Swift model | `macos/App/Platform/GenealogyStore.swift` `CatalogSource` | Has `date: CatalogDateValueInput?`, `isPublic`; no sort key (S15-03 D-13 left it to this PR) |
| Mapping | `macos/App/Platform/GoStore.swift` `mapSource` (`private static`) | Maps `date` when `s.hasDate`; ignores `dateSortKey` |
| Fake | `macos/App/Platform/FakeStore.swift` `setSourceDate` → `mutateSource` | Sets `date`; knows no key |
| Sorts | `macos/App/Features/Sources/SourcesModel.swift` `Sort` (`added`, `updated`, `az`, `za`), `buildVisibleSources` switch | No date case. `visibleSources` memo token includes the sources handle revision, so a patched row re-sorts |
| Settings | `macos/App/Features/Sources/SourcesListSettings.swift` `init(from:)` | Unknown `sort` raw value falls back to `.added` |
| Columns | `macos/App/Features/Sources/SourcesListColumns.swift` `SourcesListColumns(listWidth:)`, `sourcesSourceCell`, `sourcesTypeCell`, `sourcesGraphCell` | Compact below 880; Type 170 (nil compact); graphs 420 / 380; no width class |
| Row | `macos/App/Features/Sources/SourcesListRow.swift` `SourcesListRow.sourceLink`, `private struct SourcesLinkLabel` `secondLine` | Type cell: one `Text` framed `minHeight: firstSlot`. Second line: (type, compact) · ref · `PVListDetailCell` last update. Accessibility value: `PVLastUpdated.label("title, type, ref", …)` |
| List | `macos/App/Features/Sources/SourcesListView.swift` `SourcesListContent` `body`, `captionBand`; `@State columns = SourcesListColumns(listWidth: .infinity)` (line ~58) set in `.onGeometryChange` (line ~310) | The content `VStack` (`.frame(maxWidth: .infinity, maxHeight: .infinity)`, line ~82) fills the pane at every width; no `pvWidthClass` read; captions Source · Type · Evidence graph |
| Width class | `macos/App/DesignSystem/Components/WidthClass/PVWidthClass.swift` `\.pvWidthClass`; `macos/App/DesignSystem/Tokens/PVSpacing.swift` | `widthWideBreakpoint` 1600; no list maximum token |
| Promote | `macos/App/Features/Promote/PromoteView.swift` `private enum PromoteLayout` `listMaxRegular` (1200), `listMaxWide` (1600), `listMax(_:)`; the list body `.padding(.horizontal, PVSpacing.space10)` then `.frame(maxWidth: PromoteLayout.listMax(widthClass))` | The max is the outer frame with 40pt padding inside: 1120 / 1520 of content |
| Kit band | `macos/App/DesignSystem/Components/PageColumns/PVPageColumns.swift` `pvKeepsPageBand()` | Frames at `widthContentMax` and centers; the nearest kit modifier |
| Source page date copy | `L10n.swift` `L10n.Sources.sourceDateAdd` ("Add publication date"), `sourceDateEdit` ("Edit date"), `sourceDateAccessibility(date:)` ("Date, %@"), `sourceDateDialogTitle` ("Source date"), `sourceDateClearsCaption` ("Saving removes this Source's date.") | Used by `SourcePageIdentityHeader.sourceDate` and `SourceDateDialog` (S15-03) |
| Mapping tests | `macos/ProvenenciaTests/ObservationDepictionMappingTests.swift` | Calls internal `GoStore.mapObservation` on a built proto: the allowed exception to `swift-tests.mdc`'s no-`engine.pb.swift` rule |
| Date text | `macos/App/Platform/DateRowDisplay.swift` `DateRowDisplay.line(date:locale:)`, private `rowDate` | A range with a side that has no year returns `""` (the `guard let start, end` in `rowDate`); the header (`SourcePageIdentityHeader.sourceDate`) shows the same text. Tests pass `Locale(identifier: "en_US")` (`DateRowDisplayTests`) |
| L10n | `macos/App/Platform/L10n.swift` `L10n.Sources.sortAdded` … `sortZA`, `sortedBy`, `columnSource`, `columnType`, `columnEvidenceGraph`; `L10n.Dates`; `L10n.DesignSystem.spokenJoin(_ parts: [String])`; `macos/App/Resources/Localizable.xcstrings` | |
| Tests | `macos/ProvenenciaTests/SourcesModelTests.swift` `sortOrdersByAddedUpdatedAndTitle`, `updatedSourcePatchesListRow`, `listColumnsMatchBoard`; `macos/ProvenenciaTests/SourcesListSettingsTests.swift` `unknownFieldValuesFallBack` | Cover the four sorts, the row patch, the columns and settings fallback |

**Precedent.** The `az` / `za` pair in `SourcesModel.Sort` and `sortOrdersByAddedUpdatedAndTitle`: one case per direction, one label each, one test asserting the order. `PromoteLayout.listMax(_:)` (`PromoteView.swift`): a list that centers at 1600 at Wide, read off `pvWidthClass`. `SourcePageOverviewLayout(widthClass:)` (S15-03, #766): a feature layout that takes the width class.

## 5. Decisions

| ID | Decision | Why | Rejected |
| --- | --- | --- | --- |
| D-1 | `CatalogSource` gains `var dateSortKey: Int64? = nil`, declared after `date` (so `mapSource`'s labeled call passes `dateSortKey:` between `date:` and `isPublic:`); `GoStore.mapSource` sets it to `s.dateSortKey` only when `s.hasDateSortKey`, and drops `private` (`static func mapSource`, as `mapObservation` is) so T-7 can call it | The wire key is the only order Swift uses (rule "One date order") | Deriving a key from `date` in Swift — a second date order |
| D-2 | `Sort` gains `case dateOldest`, `case dateNewest` (raw values `"dateOldest"`, `"dateNewest"`), after `za` in `allCases`, labeled "Published date, oldest first" and "Published date, newest first" | User, 2026-11-10: two options, like A–Z / Z–A; "Published" everywhere (D-14), which replaces the board's "Source date, oldest first" and "Source date, newest first" labels | One option with a direction toggle — no such control |
| D-3 | Order: rows with a key first, by key (ascending for `dateOldest`, descending for `dateNewest`); equal keys, and the rows without a key after them, by one private `static func titleOrder(_ a: CatalogSource, _ b: CatalogSource) -> Bool` *(new)* in `SourcesModel` (`title.localizedCaseInsensitiveCompare` ascending, then `id` descending), the same in both directions. `.az` and `.za` are not changed | "Undated last" from the working rule; SL-5 | Ties by added order — not visible to the researcher |
| D-4 | A phrase-only date (date set, key absent) sorts in the undated tail and shows its phrase | The engine gives it no place (structured-date-model); the row shows what was typed | Hiding the phrase |
| D-5 | One text seam, internal `enum SourcesRowDate` *(new)* in `SourcesListRow.swift`, each function taking `locale: Locale = .autoupdatingCurrent`: `static func line(_ date: CatalogDateValueInput?, locale:) -> String?` = `DateRowDisplay.line(date:locale:)` when `date` is set and that line is not empty, else nil; `help(_:) -> String?` = `L10n.Sources.publishedHelp(date:)` of the line; `spoken(_:) -> String` = `L10n.Sources.publishedSpoken(date:)` of the line, or `L10n.string(L10n.Sources.noPublishedDateSpoken)` | One place decides dated vs undated (an empty line counts as undated); testable; S15-05 can reuse it | A helper on the `private` `SourcesLinkLabel`, which tests cannot reach |
| D-6 | The date is drawn by one private view `SourcesPublishedDate` *(new)* in `SourcesListRow.swift`, `init(date:style:)`, `style` `.labeled` or `.bare`, caption size (`PVFont.body(size: PVTypeScale.caption)`), no icon. `.labeled` dated: `HStack(spacing: PVSpacing.space2)` of `L10n.Sources.publishedLabel` in `PVColor.textMuted` with `.fixedSize()` and the line in `PVColor.textSecondary` with `.lineLimit(1).truncationMode(.tail)`; undated: `L10n.Sources.noPublishedDate` in `PVColor.textFaint`, `.lineLimit(1)`. `.bare` dated: the line alone in `textSecondary`, truncating; undated: `L10n.Dates.rowUndated` in `textFaint`. `.help(help)` on the view when dated | Board SL-1 – SL-3 (option C): a labeled value in the type's ink, apart from the last update's icon + muted look | `PVListDetailCell` (round-1 M-1, reversed with the user, 2026-11-10): it would need a leading label, a secondary tone and truncation; `PVDatePlaceLine` would need a plain (not abbreviation) leading label and the "No published date" wording. One caller until S15-05, which promotes the view to the kit |
| D-7 | `SourcesListColumns` *(extended)* takes `init(listWidth: CGFloat, widthClass: PVWidthClass)` (no default; STEP-4 updates both call sites) and gains `enum DatePlacement { case underType, column, recordLine }` *(new)* and `var datePlacement`: `isCompact` → `.recordLine`; else `.wide` → `.column`; else `.underType`. `type`: nil compact, 150 at `.column`, else 170. `published: CGFloat?` *(new)*: 150 at `.column`, else nil. `graphs` unchanged. A cell helper `sourcesPublishedCell(_:)` *(new)* beside `sourcesTypeCell`, same padding | One value the caption band and every row read, as today; board column widths. `SourcesListContent` measures only the width: `@State private var listWidth: CGFloat = .infinity`, set in `.onGeometryChange(for: CGFloat.self) { $0.size.width }`, and `body` builds `let columns = SourcesListColumns(listWidth: listWidth, widthClass: widthClass)` from it and `@Environment(\.pvWidthClass)` | Reading the width class inside each row — the band and rows could drift; caching columns in `@State` with the width class captured in the geometry closure — stale at the 1600 crossing |
| D-8 | Row placements. `.underType`: the Type cell becomes a `VStack(alignment: .leading, spacing: PVSpacing.space1)` of the type `Text` (unchanged) and `SourcesPublishedDate(.labeled)`, framed `minHeight: firstSlot` as today. `.column`: the Type cell is the type alone; a Published cell after it holds `SourcesPublishedDate(.bare)` with `.sourcesPublishedCell(columns)` and `minHeight: firstSlot`. `.recordLine`: `SourcesLinkLabel` *(extended)*: `typeLabel: String` is passed only at `.recordLine` (empty otherwise, as today when not compact), and it gains `date: CatalogDateValueInput?` and `showsRecordLine: Bool` (`columns.datePlacement == .recordLine`). With `showsRecordLine` it shows, between the title and the activity line (`VStack` spacing `PVSpacing.space1`, as today), a record line `HStack(alignment: .firstTextBaseline, spacing: PVSpacing.space3)` of the type (`.layoutPriority(0)`, truncating, as today), the `separator`, and `SourcesPublishedDate(.labeled)` (`.layoutPriority(1)`); an empty type drops the type and its separator. The activity line becomes ref · last update only (the type leaves it). The Published cell's text sits in a `firstSlot`-high frame, centered, like the type. Titles keep today's wrapping at every width | Board frames 1a, 1b, 5 | Keeping the type on the activity line when narrow — the board splits record facts from app activity |
| D-9 | Wide list. `PVSpacing.widthListWideMax: CGFloat = 1600` *(new)*: a list's content width at Wide, page gutters outside it (doc comment says so). Kit modifier `func pvListBand(_ widthClass: PVWidthClass) -> some View` *(new)* beside `pvKeepsPageBand()` in `PVPageColumns.swift`, one structure at both widths so view identity and scroll survive the crossing: `frame(maxWidth: widthClass == .wide ? PVSpacing.widthListWideMax + 2 * PVSpacing.gutterPage : .infinity, maxHeight: .infinity)`, then `.overlay` of an `HStack(spacing: 0)` holding, only at `.wide`, a full-height `PVDivider(axis: .vertical)`, a `Spacer()`, and another, then `.frame(maxWidth: .infinity, maxHeight: .infinity)` (centered). At `.regular` it changes nothing visible. `SourcesListContent` applies `.pvListBand(widthClass)` right after the content `VStack`'s existing `.frame(maxWidth: .infinity, maxHeight: .infinity)`; `.background(PVColor.surfacePage)`, the toast, the dialogs and the `.onChange` modifiers follow it, so they stay full-width. Promote adopts the token (user, 2026-11-10): `PromoteLayout.listMax(_:)` returns content widths, `.regular` → `listMaxRegular` = 1120 (unchanged content), `.wide` → `PVSpacing.widthListWideMax`; `listMaxWide` is deleted; the list body applies `.frame(maxWidth: PromoteLayout.listMax(widthClass), alignment: .leading)` before `.padding(.horizontal, PVSpacing.space10)`, then centers as today. Promote's list is 80pt wider at Wide, unchanged at Regular; it does not use `pvListBand` (no board draws hairlines or gutters there) | Board SL-W1; rule 3 as rewritten (user, 2026-11-10); one token with one meaning, one band modifier for the lists that follow (Next adds) | A hand-rolled frame in the Sources list — a fourth copy of the idiom beside `pvKeepsPageBand`, `ConclusionListPage.page`, Promote; the token meaning the outer frame — Promote's and Sources' 1600 would differ |
| D-10 | The caption band adds `L10n.Sources.columnPublished` with `.sourcesPublishedCell(columns)` after Type when `columns.published != nil` | Board frame 1b | — |
| D-11 | Accessibility: the base passed to `PVLastUpdated.label` in `sourceLink` becomes `L10n.DesignSystem.spokenJoin([title, typeLabel, SourcesRowDate.spoken(source.date), ref]) ?? title`, built by internal `static func spokenBase(title:type:ref:date:) -> String` *(new)* on `SourcesListRow` | Board SL-7 order: record facts, then app data | Interpolation — the catalog owns the join |
| D-12 | `FakeStore` gains `var nextSourceDateSortKey: Int64?` *(new)*, default nil; `setSourceDate` sets the row's `dateSortKey` to it. Seeded rows set keys through the test helper `source(...)`, which STEP-3 extends with `date:` and `dateSortKey:` | `.cursor/rules/swift-tests.mdc`: the fake scripts what Go composes | A cleared-means-nil rule or a key formula in the fake |
| D-13 | No change to `SourcesListSettings` decoding | The per-field fallback covers an unknown raw value (AC-2.6) | A settings migration |
| D-14 | "Published" is the date's one name (user, 2026-11-10). On the Source page: `sourceDateAdd` → "Add published date", `sourceDateEdit` → "Edit published date", `sourceDateAccessibility(date:)` → "Published, %@", `sourceDateDialogTitle` → "Published date", `sourceDateClearsCaption` → "Saving removes this Source's published date." Keys unchanged; values and comments change in `L10n.swift` and `Localizable.xcstrings`. The dialog subtitle "When the record was made or first published." stays | One word for one concept in copy, VoiceOver and tests; S15-05 uses it too | Leaving "Source date" / "publication date" on the page beside "Published" on the list |

**Rules this plan bends.** `docs/macos-client-patterns.md` § 9 Wide rule 3 ("Lists, tables, grids and canvases fill the width", also in the `add-design-brief` Paste block) → the board centers the list at 1600 instead → user, 2026-11-10: "fill the width if you have something useful to fill it with, but otherwise constrain it" → rule 3 rewritten in § 9, the Paste block, `docs/ideas/wide-layout.md` and this brief in the planning PR [#770](https://github.com/mendahu/provenencia/pull/770); STEP-8 checks the token sentence names the landed token. Also: the S15-D1 board shows "Add publication date" and the S15-D2 board "Source date, oldest first" and "newest first" sorts; D-14 changes that copy without a new board (user, 2026-11-10); STEP-8 records it on both briefs' **Board** lines.

## 6. Design

### Codebase fit

| New *(this PR adds)* | Kind | Closest existing | Why not reuse or extend it |
| --- | --- | --- | --- |
| `CatalogSource.dateSortKey` *(new)* | field | `CatalogSource.date`, `updated.revision` | Extends the model; the key is not derivable in Swift |
| `SourcesModel.Sort.dateOldest`, `.dateNewest` *(new)* | enum cases | `.az`, `.za` | Extends the enum |
| `SourcesModel.titleOrder` *(new)* | private func | `buildVisibleSources` `.az` inline compare | One tie rule for both date sorts |
| `SourcesRowDate` *(new)* | enum | `PVLastUpdated.detail`, `.label` | Same shape for the published date; stays in the feature until S15-05 is a second caller |
| `SourcesPublishedDate` *(new)* | private view | `PVListDetailCell`; `PVDatePlaceLine`; `SourcePageIdentityHeader.sourceDate` | D-6: the cell lacks a label, the secondary tone and truncation; `PVDatePlaceLine`'s leading is an abbreviation and its missing wording is "unknown"; the header has a pencil, focus and a 14pt icon. Promoted at S15-05 |
| `SourcesListColumns.DatePlacement`, `published`, `sourcesPublishedCell` *(new)* | enum, field, helper | `SourcesListColumns.type`, `sourcesTypeCell` | Extends the one column value |
| `SourcesListRow.spokenBase` *(new)* | static func | `L10n.DesignSystem.spokenJoin`, `PVLastUpdated.label` | Uses both |
| `PVSpacing.widthListWideMax` *(new)* | token | `PromoteLayout.listMaxWide`, `PVSpacing.widthContentMax` | Promote's literal is deleted and Promote reads the token (D-9); `widthContentMax` (1180) is the record-page column |
| `pvListBand(_:)` *(new)* | kit modifier | `pvKeepsPageBand()`, `ConclusionListPage.page(meta:body:)`, Promote's frame | The same idiom at a list's width with side hairlines; the next list redesigns call it |
| `FakeStore.nextSourceDateSortKey` *(new)* | test hook | `FakeStore.nextWriteEffects` | The fake's scripting shape |
| L10n keys (§ Strings) *(new)* | keys | `L10n.Sources.sortAZ`, `columnType`; `L10n.Dates`; `L10n.Conclusions.timelineUndated` | Sorts and captions sit beside theirs; "Undated" is a row date value beside `DateRowDisplay`'s strings; the Timeline's heading may change on its own |

**Builds on.** `DateRowDisplay`, `L10n.DesignSystem.spokenJoin`, `PVLastUpdated`, `pvWidthClass`, `PromoteLayout.listMax`'s Wide pattern, the wire key from `sources.DateSortKey`, the `Sort` enum and settings store, the `.updatedSource` row patch.
**Deletes.** `PromoteLayout.listMaxWide`. The "Source date" and "publication date" values on the Source page (D-14). The type on the narrow activity line moves to the record line. Nothing else becomes dead.
**Next adds.** S15-05 shows the published date in the Person page's Sources section through `SourcesRowDate` (promote it to a recipe then). A later list redesign calls `pvListBand` until its detail sits beside it.

### Reach

| New path or state | Destination | Checked: resets on the new id · handles a missing referent · shows the state |
| --- | --- | --- |
| A Source with a date reaches the list row | `SourcesPublishedDate` | Dated, phrase-only, empty-line and undated each render (AC-1.1 – AC-1.3, T-6) |
| The list at Wide | `SourcesListContent` | Centered column; empty, loading, error and filter-empty states sit inside it (they are in the same `VStack`) |
| A stored settings `sort` this build adds | `SourcesListSettings.init(from:)` | Older builds fall back to `.added` (AC-2.6) |

### Scale

At the map's Scale (3,000 Sources): the sort adds one `Int64?` comparison per pair inside the existing memoized `buildVisibleSources`. Each drawn row calls `DateRowDisplay.line` once inside a `LazyVStack`. T-5 runs the date sort over 3,000 rows and asserts the order and one build; no timing assertion.

### Edges

- **Assumptions.** The key is a total order and equal keys mean equal places (S15-02 T-10). Ties fall to title, then id, so the sort is deterministic (T-2).
- **Neighbors.** `abt 1880` and `1880` share a key and tie to title; `bef. 1880` before `1880` before `aft. 1880` (T-2). A list exactly 880pt wide is not compact (T-8).
- **Removed inputs.** A cleared date drops the key and the row moves to the tail (T-3).
- **Real data.** None — no table or seed drives the rule.
- **Untrusted size.** None — rows come from the app's own catalog read.

### Schema

None — no catalog change.

### Go core

None — `sources.DateSortKey` and the wire field exist (S15-02).

### FFI

None — `Source.date_sort_key` is on the wire.

### macOS

The board in words (implementer may also read it, see **Board**):

- **Frame 1a, Regular 1102 × 900, sorted by Date added.** Rows as today. The Type cell (170) reads "Census" and under it "Published 1 Jun 1880" ("Published" muted, the date secondary); undated rows read "No published date" faint. The Source cell's second line is unchanged: `SRC-4KD2P` · history icon "2 days ago".
- **Frame 1b, Wide 2296 × 1440.** The whole list (title, description, count, Add Source, toolbar, caption band, rows) in one column centered in the pane, 1600 of content plus 32pt gutters, a hairline down each side, page background beyond. Captions SOURCE · TYPE · PUBLISHED · EVIDENCE GRAPH. Type 150 holds the type alone; Published 150 holds "1 Jun 1880" (secondary) or "Undated" (faint). Titles stay on one line.
- **Frames 2a / 2b, sorted "Published date, oldest first"** (the board reads "Source date, oldest first"; D-14). `bef 3 Mar 1830`, `bet 1834 and 1840`, `1 Jun 1880`, `4 Jul 1881`, `abt 1895`, then the phrase-only and undated rows by title.
- **Frame 3.** The Sort menu open with six options, "Published date, oldest first" checked.
- **Frame 4.** "Published date, newest first": dated rows reversed, undated tail by title.
- **Frame 5, narrow 860.** Under the title: "Census · Published bet 12 Mar 1834 and 3 Sep 1840" truncated with an ellipsis and its help shown; under it `SRC-…` · last update.
- **Frame 6.** Type filter Census with the date sort kept, count "2 of 12".

| Building block | Component | Home | Notes |
| --- | --- | --- | --- |
| Published date | `SourcesPublishedDate` *(new, private)* | `macos/App/Features/Sources/SourcesListRow.swift` | D-6, D-8 |
| Columns | `SourcesListColumns` *(extended)* | `macos/App/Features/Sources/SourcesListColumns.swift` | D-7 |
| Wide column | `pvListBand` *(new)*, `PVSpacing.widthListWideMax` *(new)* | `PVPageColumns.swift`; `PVSpacing.swift`; applied in `SourcesListView.swift` | D-9 |
| Caption band | `captionBand` | `SourcesListView.swift` | D-10 |
| Sort options | `PVSelect` | `SourcesListView.sortMenu` | Options come from `Sort.allCases`; no view change beyond the new cases |
| Last update | `PVListDetailCell` | unchanged | |

States: undated, or a set date with an empty line → "No published date" / "Undated" (faint); phrase-only → phrase; long → truncated with help; empty, loading, error and filter-empty unchanged, inside the centered column at Wide. Accessibility: D-11; the row link merges its children, so the date has no identifier of its own; sort options keep `sources.sort.<raw>` identifiers.

### Strings

L10n *(new)* in `L10n.swift` and `Localizable.xcstrings` (key → value; comment):
- `L10n.Sources.sortDateOldest`: `sources.list.sort.dateOldest` = "Published date, oldest first"; "Sources sort option: published date, oldest first, undated last"
- `L10n.Sources.sortDateNewest`: `sources.list.sort.dateNewest` = "Published date, newest first"; "Sources sort option: published date, newest first, undated last"
- `L10n.Sources.publishedLabel`: `sources.list.publishedLabel` = "Published"; "Label before a Source's date on a Sources list row"
- `L10n.Sources.columnPublished`: `sources.list.columnPublished` = "Published"; "Sources list column caption: the Source's published date (Wide)"
- `L10n.Sources.noPublishedDate`: `sources.list.noPublishedDate` = "No published date"; "Sources list row: the Source has no date"
- `L10n.Sources.publishedHelp(date:)`: `sources.list.publishedHelp` = "Published %@"; "Hover text on a Sources list row's date; argument is the date"
- `L10n.Sources.publishedSpoken(date:)`: `sources.list.publishedSpoken` = "published %@"; "VoiceOver part of a Sources list row; argument is the date"
- `L10n.Sources.noPublishedDateSpoken`: `sources.list.noPublishedDateSpoken` = "no published date"; "VoiceOver part of a Sources list row with no date"
- `L10n.Dates.rowUndated`: `dates.row.undated` = "Undated"; "A list cell's date when there is none"

Changed values (keys kept), D-14: `sources.page.sourceDate.add` "Add published date"; `.edit` "Edit published date"; `.accessibility` "Published, %@"; `.dialogTitle` "Published date"; `.clearsCaption` "Saving removes this Source's published date."

`sortedBy` lowercases the label: "Sorted by published date, oldest first". Removed: none.

## 7. Steps

**STEP-1 — Map the key.**
Files: `macos/App/Platform/GenealogyStore.swift`, `macos/App/Platform/GoStore.swift`, `macos/App/Platform/FakeStore.swift`, `macos/ProvenenciaTests/SourceMappingTests.swift` *(new)*.
Do: D-1, D-12. Add T-7.
Check: T-7 passes; existing `SourcesModelTests` pass.

**STEP-2 — Strings.** Skill: `add-localized-string`.
Files: `macos/App/Platform/L10n.swift`, `macos/App/Resources/Localizable.xcstrings`.
Do: every key and changed value in §6 Strings (D-14 included).
Check: `python3 scripts/check-localizable-xcstrings.py` passes.

**STEP-3 — The two sorts.** Skill: `add-swift-test`.
Files: `macos/App/Features/Sources/SourcesModel.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`, `macos/ProvenenciaTests/SourcesListSettingsTests.swift`.
Do: D-2, D-3, D-4; labels in `Sort.label`. Extend the test helper `source(id:title:typeID:ref:updatedRevision:)` with `date: CatalogDateValueInput? = nil, dateSortKey: Int64? = nil`. Add T-1 – T-5.
Check: T-1 – T-5 pass.

**STEP-4 — Columns and the width token.** Skill: `add-ui-component`.
Files: `macos/App/Features/Sources/SourcesListColumns.swift`, `macos/App/Features/Sources/SourcesListView.swift`, `macos/App/DesignSystem/Tokens/PVSpacing.swift`, `macos/App/Features/Promote/PromoteView.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`.
Do: D-7, including the `listWidth` state and the `columns` built in `body`; the token and Promote's adoption (D-9). Update `listColumnsMatchBoard` to pass `widthClass: .regular`. Add T-8.
Check: the app builds; T-8 and `listColumnsMatchBoard` pass; Promote at Regular looks as before.

**STEP-5 — The row.** Skill: `add-ui-component`.
Files: `macos/App/Features/Sources/SourcesListRow.swift`, `macos/ProvenenciaTests/SourcesModelTests.swift`.
Do: D-5, D-6, D-8, D-11. Add T-6.
Check: T-6 passes; frames 1a and 5 match in the app.

**STEP-5a — The graph column's hairline spans the row.** Skill: `add-ui-component`. Added at the user's request, 2026-11-10.
Files: `macos/App/Features/Sources/SourcesListRow.swift`.
Do: today `graphStack` draws `PVDivider(axis: .vertical)` as an `.overlay(alignment: .leading)` on the stack itself, so the line is only as tall as the graph buttons; a title that wraps makes the Source cell, and so the row, taller, and the line stops short. Tie the line to the row's height, not the stack's: give `graphStack` `.frame(maxHeight: .infinity, alignment: .top)` after its `.padding(.vertical, …)` and before the `.overlay` (the row already has `.fixedSize(horizontal: false, vertical: true)`, and `sourceLink` uses the same `maxHeight: .infinity` idiom), so the stack fills the row and the hairline runs top to bottom. The buttons stay top-aligned where they are. One commit: "Run the Sources list's graph column hairline down the whole row."
Check: the app builds; with a title long enough to wrap to three lines at Regular, the hairline meets the row's bottom divider; a row with two graphs (taller stack than the Source cell) looks as before.

**STEP-6 — The list at Wide.** Skill: `add-ui-component`.
Files: `macos/App/Features/Sources/SourcesListView.swift`.
Do: `pvListBand` and its use (D-9), D-10.
Check: frames 1b and 2b match in the app; crossing 1600 keeps scroll and sort; Promote at Wide is 1600 of content, centered.

**STEP-7 — Version.**
Files: `VERSION`, `core/version.go`, every `MARKETING_VERSION` line in `macos/Provenencia.xcodeproj/project.pbxproj` (four at the planning base).
Do: the minor above `VERSION` on `main` when you branch; if `main` bumps again before merge, rebase and take the next one.
Check: `VERSION`, `core.Version` and every `MARKETING_VERSION` line agree.

**STEP-8 — Docs.**
Files: `../completed.md`, `../deployment-plan.md` (map row), this plan (status), `../design/S15-D2-sources-list.md`, `../design/S15-D1-source-page.md` (**Board** line: D-14's copy), `../design/README.md`, `docs/macos-client-patterns.md` § 9.
Do: per the Definition of done. Create the folder `archive` *(new)* under `../design/` and `git mv` the brief into it; in `../design/README.md` move the S15-D2 row from Scheduled to Archived, its link pointing into that folder, its Board cell the board link, its Built by cell "S15-04 (#N)". `completed.md`: one paragraph at the top in the shape of the S15-03 entry. § 9: the sentence on `widthListWideMax` matches the landed token, `pvListBand` and Promote.
Check: `python3 scripts/check-plan-docs.py` on this plan passes; links resolve.

Commits: one per behavior. The subject is one full imperative sentence that says what changed for the researcher, ending in a period.

## 8. Tests

| ID | Layer | File · test | Case: input → expected | Covers |
| --- | --- | --- | --- | --- |
| T-1 | Swift | `SourcesModelTests.swift` `sortByDateOldestPutsUndatedLast` *(new)* | `s1` "Delta" key 100, `s2` "Echo" key 200, `s3` "Foxtrot" key 300, `s4` "Golf" key 400, `s5` "Bravo" no date, `s6` "Alpha" no date; `.dateOldest` → s1, s2, s3, s4, s6, s5 | AC-2.1, D-3 |
| T-2 | Swift | `SourcesModelTests.swift` `sortByDateNewestKeepsUndatedTailAndTies` *(new)* | T-1's six plus `s7` "Zeta" key 500, `s8` "Eta" key 500, `s9` "Mill notes" with `date: CatalogDateValueInput(kind: "point", phrase: "in the reign of Victoria")` and no key, `s10` "Abbey roll" no date; `.dateNewest` → s8, s7, s4, s3, s2, s1, s10, s6, s5, s9; `.dateOldest` → s1, s2, s3, s4, s8, s7, s10, s6, s5, s9 | AC-2.2, AC-2.3, AC-2.4, D-3, D-4 |
| T-3 | Swift | `SourcesModelTests.swift` `clearedDateMovesRowToUndatedTail` *(new)* | T-1's six, `.dateOldest`; `nextSourceDateSortKey` nil, `store.setSourceDate(…, sourceID: "s1", date: nil)`, apply with `.updatedSource` → s2, s3, s4, s6, s5, s1. Then `nextSourceDateSortKey = 450`, set s5's date to `CatalogDateValueInput(kind: "point", startYear: 1880, startMonth: 6, startDay: 1)`, apply → s2, s3, s4, s5, s6, s1 | AC-1.6, D-12 |
| T-4 | Swift | `SourcesListSettingsTests.swift` `dateSortRoundTrips` *(new)*; `SourcesModelTests.swift` `sortControlLabelForDateSort` *(new)* | Save `.dateNewest`, load → `.dateNewest`; `sortControlLabel` → "Sorted by published date, newest first". AC-2.6 stays covered by `unknownFieldValuesFallBack` | AC-2.5 |
| T-5 | Swift | `SourcesModelTests.swift` `dateSortAtScaleWithFilter` *(new)* | 3,000 Sources `s0000` … `s2999`, title "Source NNNN"; type `t1` for even N, `t2` for odd; each N divisible by 30 dated with key N (100 dated, all even since 30 is even; `s0000` has key 0, a present key); type filter `t1`, `.dateOldest` → 1,500 rows: the 100 dated rows by key, then the 1,400 undated even rows by title; `visibleSourcesBuildCount` rises by 1 | AC-2.7, §6 Scale |
| T-6 | Swift | `SourcesModelTests.swift` `publishedDateTextAndSpokenBase` *(new)* | With `locale: Locale(identifier: "en_US")`: `SourcesRowDate.line` of `CatalogDateValueInput(kind: "point", startYear: 1880, startMonth: 6, startDay: 1)` → "1 Jun 1880"; phrase-only → "in the reign of Victoria"; nil → nil; `CatalogDateValueInput(kind: "range", startYear: 1834)` (no end) → nil. `help` of the point → "Published 1 Jun 1880", of nil → nil. `SourcesListRow.spokenBase(title: "T", type: "Census", ref: "SRC-1", date: nil)` → "T, Census, no published date, SRC-1" | AC-1.1, AC-1.2, AC-1.3, AC-1.7, D-5, D-11 |
| T-7 | Swift | `SourceMappingTests.swift` `mapsDateSortKey` *(new)*, shaped as `ObservationDepictionMappingTests` (the allowed mapping exception to `swift-tests.mdc`) | `Provenencia_Engine_V1_Source` with `dateSortKey = 42` → `GoStore.mapSource(_:).dateSortKey == 42`; without it → nil | D-1 |
| T-8 | Swift | `SourcesModelTests.swift` `columnsPlaceTheDateByWidth` *(new)* | `(1000, .regular)` → `.underType`, type 170, published nil, graphs 420; `(880, .regular)` → `.underType`; `(860, .regular)` → `.recordLine`, type nil, graphs 380; `(1664, .wide)` → `.column`, type 150, published 150, graphs 420; `(860, .wide)` → `.recordLine`. `PVSpacing.widthListWideMax == 1600` | AC-1.4, AC-1.5, D-7, D-9 |
| T-9 | Swift | `SourcePageModelTests.swift` `sourceDateCopySaysPublished` *(new)* | `L10n.Sources.sourceDateAccessibility(date: "1 Jun 1880")` → "Published, 1 Jun 1880"; `L10n.string(L10n.Sources.sourceDateAdd)` → "Add published date"; `sourceDateDialogTitle` → "Published date" | AC-1.8, D-14 |

Must stay green: everything `scripts/check-pr.sh` runs. Most likely to break: `listColumnsMatchBoard` (new init); any S15-03 test asserting the old Source page date copy (`SourcePageModelTests`, around the "Source date, Private" mark).

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
- [ ] Domain doc updated if a rule in docs/*.md changed: § 9's `widthListWideMax` sentence matches the code
- [ ] PR description: what shipped, AC evidence, deviations, manual check
```

## 10. Manual check

1. Open a project with at least six Sources. On three Source pages set dates `1 Jun 1880`, `bet 1834 and 1840`, `bef 3 Mar 1830`; on one set the phrase "in the reign of Victoria"; leave the rest undated. Go to Sources at the default window: each Type cell shows "Published …" or "No published date" under the type (AC-1.1 – AC-1.3, AC-1.6). Compare with frame 1a.
2. Sort "Published date, oldest first": `bef 3 Mar 1830`, `bet 1834 and 1840`, `1 Jun 1880`, then the phrase-only and undated rows by title. Sort "newest first": the dated three reversed, the same tail (AC-2.1 – AC-2.4). Frames 2a, 4.
3. Quit and reopen the project: the sort holds (AC-2.5). Apply a type filter: the date order holds (AC-2.7). Frame 6.
4. Narrow the window below 880pt of list: the record line and activity line; a long range truncates and shows its help (AC-1.5). Frame 5.
5. Widen past 1600pt of pane: the list centers with side hairlines and a Published column; scroll and sort are kept (AC-1.4). Frame 1b.
6. VoiceOver on a dated and an undated row reads AC-1.7. On a Source page, the header and dialog read AC-1.8.
7. Open Promote at Wide: its list is 1600 of content, centered; at Regular it looks as before.
8. Give a Source a title long enough to wrap to two or three lines: the hairline left of its Evidence graph column runs down to the row's bottom divider (STEP-5a).

## 11. Risks and stop conditions

| Risk | Mitigation |
| --- | --- |
| The narrow row's extra line breaks the graph slot alignment (`firstSlot`) | The Source cell is top-aligned and grows; the graph stack keeps its own padding; frame 5 is the check |
| A test helper constructs `CatalogSource` positionally | The new field has a default (D-1) |
| Centering at Wide pushes the toast or dialogs inside the column | D-9 keeps those modifiers on the outer full-width view |

**Stop and report if:**

- `Provenencia_Engine_V1_Source.hasDateSortKey` does not exist in `engine.pb.swift` at your base.
- The board's frames and §6 macOS's words disagree (other than D-14's copy). If you cannot read the board, §6 is the authority.
- `pvWidthClass` is not set in the Sources list's environment (it reads `.regular` at a 2296pt pane).

## 12. Open questions

None.

## 13. PR

**Title:** S15-04 — The Sources list shows and sorts by the date
**Body:** objective in one sentence; what shipped against US-1 and US-2; AC evidence; deviations from the plan; manual check; `Plan: docs/deployment-plan/spike-15/prs/S15-04-sources-list-date.md`.
