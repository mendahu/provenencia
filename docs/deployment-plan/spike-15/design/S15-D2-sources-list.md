# S15-D2 — Sources list: the Source's date on each row and as a sort

**Kind:** Claude Design board (**enhancement** of the Sources list as shipped: S2-04 / S12-D5 list, three columns Source · Type · Evidence graph, the S12-14 toolbar. Draw on the existing Sources list frames; add the date to each row and two options to the Sort menu, and give the list its first Wide frame. Drawn as option C: see **Board**)
**Spike:** Provenencia Spike 15, slice 1
**View:** the Sources list (the workspace's Sources destination: header, toolbar, caption band, rows)
**Implements later as:** PR **S15-04** (SL-1 – SL-7, SL-W1, SL-W2)
**Board:** agreed 2026-11-10 — [Claude Design — Sources List](https://claude.ai/artifact/JDm51QdorUvY79EFoci4cm), page *S2-04 Sources list*, turn "S15-D2 · Sources list": frames 1a/1b (date added), 2a/2b (Source date, oldest first), 3 (Sort menu), 4 (newest first), 5 (narrow, 860), 6 (type filter Census), and the panel "S15-D2 · Requirements as drawn"; the frame component is *Sources List Frame* (presets `s15added`, `s15asc`, `s15sortmenu`, `s15desc`, `s15narrow`, `s15census`). Board decisions (user, 2026-11-10), which replace the brief's recommendations where they differ: the date is named as the **published** date, never shown with an icon, and kept apart from the last update (option C): "Published 1 Jun 1880" under the type at Regular, its own **Published** column at Wide, a record line (type · Published date) above the activity line when narrow; undated reads "No published date" (or "Undated" in the column), faint; at Wide the whole list centers at 1600pt of content (Wide rule 3 rewritten to match). § 4 below is the requirements as drawn
**Depends on:** S15-02 (landed [#762](https://github.com/mendahu/provenencia/pull/762)): every Source carries an optional date and its place in date order; S15-03 (landed [#766](https://github.com/mendahu/provenencia/pull/766)): the Source page shows that date with a calendar icon
**Related:** [`../source-provenance-date.md`](../source-provenance-date.md); board [S15-D1](S15-D1-source-page.md) SD-1 (how the date reads on the Source page); [`../../../ideas/wide-layout.md`](../../../ideas/wide-layout.md) "Sources, Persons, Events and Places lists"; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); working rules in [`../deployment-plan.md`](../deployment-plan.md#working-rules); plan [`../prs/S15-04-sources-list-date.md`](../prs/S15-04-sources-list-date.md)
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board. This file is the whole prompt.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
- **Rethink** (this brief says replace): throw away the old frames. Do not keep a before/after to ship.
- **Enhancement**: add to the existing frames. Do not start a second composer / graph / page.

1. **Clear this board’s local design-system cache.** Claude Design keeps a stale pack; drawing against it invents local copies of kit controls.
2. **Delete this board’s reference** to the design-system bundle.
3. **Pull a fresh copy** of the Provenencia design system from the main project. Do not continue until the fetched kit lists current components. If the kit looks stale or empty, delete the cache and refetch. Do **not** draw a replacement kit locally.
4. **Compose from that kit.** Instance existing components. Reach for a **bespoke / local** control only when the use is truly this domain. One call site is not a new design-system primitive.

**Reach for (kit).** Instance these first. The **UI building-block inventory** later in this brief names the snowflakes and which kit piece each situation should use.

| Situation | Use |
| --- | --- |
| Labeled value, textarea, or trailing control | Field + TextArea / Input |
| Primary / secondary / ghost action | Button; icon-only → IconButton |
| Choose one from a short list | Select |
| Searchable pick | ComboBox |
| Warning, error, or inline hint | Callout |
| Page- or pane-level empty | EmptyState |
| Confirm replace or destroy | Confirm (`item:` snapshot, not a Bool) |
| Resource delete with inbound check | DeleteImpact recipe (S8-D9 / **S8-13**) — confirm if allowed, notice if blocked |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |
| Record or detail page at Regular and Wide | Page columns |
| Choose one of a few options inline | Segmented control (ChipGroup segmented) |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### Two widths — draw both

The window opens at 1366 × 900pt but is often stretched across a large monitor.
Every view has two width classes, measured on the view's own pane (the window
minus the sidebar), not the screen. A view with a side pane (an inspector, a detail
pane) plans that pane inside its Wide frame:

- **Regular** — pane below 1600pt. Draw the pane at 1102 × 900 (the default 1366pt window minus the open sidebar).
- **Wide** — pane 1600pt and up. Draw the pane at 2296 × 1440 (a 2560pt window minus the sidebar).

Draw only the view's own pane. Do not draw the app's sidebar, its navigation items or the window's title bar: they are not part of the board. A frame that needs the toolbar (breadcrumbs, Back and Forward) says so in the brief, and draws only the toolbar above the pane.

Rules for Wide:

1. **The extra space does a job.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. If the view truly has nothing to add, center its content. Content jammed to one side with empty space beside it is never the Wide design.
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band; with `fillSplitSideMax` the primary and the band are the main-and-side group the page lays out inside it, as the Source page does); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Fill the width only with something useful.** A list, table, grid or canvas takes the extra width when Wide has more to show in it: more columns, more cards, the detail beside the list. When it has nothing more to show, it keeps a maximum width (1600pt of content for a list) and is centered, with the page gutters outside it. Empty space stretched between a row's columns is never the Wide design. Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Draw on the Sources list frames as a new turn "S15-D2 · Sources list", starting from the list as it ships today (described in § 1). Keep the header (title, description, count line, Add Source), the toolbar (search, type filter, evidence status filter, Sort), the caption band and its three columns, each row's cover, title, ref, last update, type cell and Evidence graph buttons, the narrow arrangement below 880pt, the filter-empty state and the right-click Copy Source. Add the date to each row's second line and two date options to the Sort menu, and draw the list at Wide.

---

## 1. Objective

The Sources list is every Source in the project, one row each. Today, from the top: a header (the title "Sources", a one-line description, a mono count line "42 sources" and a primary "Add Source" button at the right); a toolbar on a card background (a search field "Search title or ref", a type filter Select "All types", an evidence status Select "All statuses", a Sort Select reading "Sorted by date added"); a caption band in micro caps (SOURCE · TYPE · EVIDENCE GRAPH) pinned at the top of the scroll; then the rows. A row has three cells: the **Source** cell (a 40pt cover thumbnail, the title in body-small medium, and under it a second line in muted text: the mono ref `SRC-4KD2P`, a dot, and the last update with a small clock icon "2 days ago"); the **Type** cell (the type name, caption, secondary, 170pt); and the **Evidence graph** cell (420pt, a hairline on its leading edge, one button per graph with its status badge). Below 880pt of list width the Type column goes, the type joins the start of the second line, and the graph column narrows to 380pt. The list has no Wide arrangement: at 2296pt the Source cell stretches and the rows stay one per line.

The Sort menu holds four options today: Date added (newest first, the default), Date updated, Alphabetical, Reverse alphabetical. The sort, the filters and the search are remembered per project.

**What changes** (as drawn, user 2026-11-10).
- **The Source's own date, named as published.** A Source has one optional date: when the record was made or first published. The list names it as the published date, apart from the last update: the Source date is a fact about the record, the last update is app activity, so they never share a line. "Published 1 Jun 1880" sits under the type at Regular; at Wide it is its own Published column; below 880pt a record line (type · Published date) sits above the activity line (ref · last update). An undated Source reads "No published date" (or "Undated" in the column), faint.
- **Two date sorts.** The Sort menu gains "Source date, oldest first" and "Source date, newest first". Undated Sources come last in both.
- **Wide.** The list gets its Wide frame: centered at 1600pt of content with a Published column (SL-W1).

**What stays.** The header, the toolbar's controls and order, the caption band and its columns, the Type and Evidence graph cells, the narrow arrangement, the filter-empty and catalog-empty states, Copy Source on right-click, and the four existing sorts with their labels.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The Source date | One optional structured date per Source: the original publication, or when the record was made. Set on the Source page only. It is not a column the researcher edits in the list. |
| How a date reads | As the app's lists and page summaries write dates, and as the Source page header writes it: "1880", "12 Mar 1880", "abt 1880", "bef 3 Mar 1830", "bet 1834 and 1840", or the phrase as the researcher typed it ("in the reign of Victoria"). |
| Undated | Most Sources start undated; there is no backfill. A project of 3,000 Sources may have 2,900 undated rows. Each reads "Undated", muted. |
| A phrase-only date | A date that is only words ("in the reign of Victoria") shows its words on the row but has no place in date order: it sorts with the undated Sources. |
| Date order | The engine gives each dated Source its place in date order, the same order as the Person page's Timeline: `bef. 1880` sorts before `1880`, `1880` before `aft. 1880`; `abt 1880` sorts with `1880`. The list follows that order and never orders dates itself. |
| Ties | Two Sources with the same place ("1880" and "1880") are ordered by title A to Z. Undated Sources, at the end, are ordered by title A to Z in both directions. |
| The sort is remembered | Per project and user, with the filters and search. Choosing a date sort and leaving the list keeps it. |
| A date set elsewhere | Setting or clearing a date on the Source page updates that row's date, and its place under a date sort, when the list is shown again. |

### 2.1 What this board is not

- Filtering by date, a date range filter, or a "Dated only" filter (not planned in this spike).
- A Private mark, glyph or filter: the Private switch does nothing in the list (working rule "Not public is a stored switch and nothing else").
- Editing the date from the list: the date is set on the Source page.
- Group headings by decade or century, or an "Undated" section heading: the undated rows simply come last.
- The Source page (S15-D1), the Person page's Sources section (S15-D3), the Copy Source form (S15-D13).
- List beside the selected detail at Wide (the lists' later redesign, [`wide-layout.md`](../../../ideas/wide-layout.md)).

---

## 3. Implementation gate

| PR | Ships | Not that PR |
| --- | --- | --- |
| **S15-04** | SL-1 – SL-7, SL-W1, SL-W2 | Date filters (not planned); the Person page's Sources section (S15-05) |

---

## 4. Requirements

As drawn on the board (user, 2026-11-10). The board's panel "S15-D2 · Requirements as drawn" is the long form.

| ID | Requirement |
| --- | --- |
| SL-1 | **The published date.** At Regular the Type cell reads the type, then under it "Published 1 Jun 1880": a muted "Published" label and the date (through the app's row date display, as the Source page header writes it) in secondary ink like the type, caption size, no icon. The last update stays on the Source cell's second line after the ref, behind its history icon, as shipped. The two dates differ in place, label and tone. Not a Chip or Badge. |
| SL-2 | **Undated.** "No published date" where the label and date stand (Regular, narrow), or "Undated" under the Published caption (Wide), in faint ink. A phrase-only date shows its phrase. |
| SL-3 | **A long date.** The date truncates with an ellipsis inside its column (Type 170 at Regular, Published 150 at Wide), with "Published {full date}" as help; the "Published" label never truncates. Narrow (below 880pt of list): the record line is type · Published date; the type gives way first, then the date; the activity line (ref · last update) stays below it and the ref never truncates. Neither line wraps; a narrow row gains one line. |
| SL-4 | **The Sort menu.** Two options after "Reverse alphabetical": "Source date, oldest first" and "Source date, newest first". The control reads "Sorted by source date, oldest first" when chosen. Same Select, same icon. |
| SL-5 | **The order.** The engine's place in date order, never parsed by the list: `bef. 1830` < `1830` < `abt 1834` = `bet 1834 and 1840` < `1880` < `aft. 1880`. Ties by title A to Z. Undated and phrase-only Sources last in both directions, by title A to Z. |
| SL-6 | **Empty and filtered.** A filter or search keeps the chosen date sort. The filter-empty and catalog-empty states are unchanged. |
| SL-7 | **VoiceOver.** Record facts, then app data: "1880 U.S. census, Ottawa County, Kansas, Census, published 1 Jun 1880, SRC-4KD2P, updated 2 days ago"; an undated Source reads "no published date" in the date's place. Sort options reachable by keyboard as today. |
| SL-W1 | **Wide: center, don't stretch.** From 1600pt of pane the whole list (header, toolbar, caption band, rows) sits in one centered column, at most 1600pt of content plus the page gutters, with hairlines on its sides; the pane takes what is left. The date moves out from under the type into its own Published column: Source fills (about 880pt), Type 150, Published 150, Evidence graph 420. Titles stay on one line. |
| SL-W2 | **Regular.** The three columns as today (Source fills, Type 170, Evidence graph 420) with the date under the type. Below 880pt of list width: record line, then activity line. Crossing 880pt or 1600pt keeps the scroll position and the chosen sort. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Sources list | Header, toolbar, caption band, rows across the pane; Source fills, Type 170 (type, then Published date under it), Evidence graph 420; below 880 the record and activity lines | One centered column of at most 1600pt of content plus gutters, hairlines at its sides; Source fills, Type 150, Published 150, Evidence graph 420 | The date leaves the Type cell for its own Published column; the list stops growing at 1600 and centers (rule 3) |

---

## 5. Suggested frames

Frames with an a / b pair are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others at Regular. Data: twelve Sources, invented, of types Census, Parish register, Photograph, Letter and Newspaper. Dated: "1880 U.S. census, Ottawa County, Kansas" (`1 Jun 1880`), "St Brigid parish baptisms" (`bet 1834 and 1840`), "Robins family portrait" (`abt 1895`), "Letter from Ada to her sister" (`bef 3 Mar 1830`), "Ottawa County Herald, 4 July issue" (`4 Jul 1881`), "Notes on the old mill" (phrase "in the reign of Victoria"); six undated. Invent every person and record.

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | SL-1, SL-2, SL-W1, SL-W2 | The list sorted by Date added: dated and undated rows mixed |
| 2a / 2b | SL-4, SL-5 | The list sorted "Source date, oldest first": the dated rows in order, the phrase-only and undated rows last; the Sort control reading "Sorted by source date, oldest first" |
| 3 | SL-4 | The Sort menu open, six options, "Source date, oldest first" checked |
| 4 | SL-5 | "Source date, newest first" |
| 5 | SL-3, SL-W2 | The list at 860pt (narrow): the record line (type · Published date) above the activity line (ref · last update); one long range date truncated with its tooltip shown (as drawn) |
| 6 | SL-6 | A type filter "Census" with the date sort kept |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sources list row | Snowflake | Extend | `macos/App/Features/Sources/SourcesListRow.swift` `SourcesListRow`, `SourcesLinkLabel` | The Type cell gains the published line (Regular); a Published cell (Wide); the narrow record line |
| Published date | Snowflake (inside the row) | Extend (the row) | `macos/App/Features/Sources/SourcesListRow.swift` | Muted "Published" label + the date in secondary ink through the row date display (`DateRowDisplay`); faint "No published date" / "Undated". No icon. Not a Chip, Badge or List detail cell (that cell is icon + muted text: the last update's look, which the board keeps apart) |
| Last update | Component | Ship | `macos/App/DesignSystem/Components/List/PVList.swift` `PVListDetailCell` via `PVLastUpdated.detail` | Unchanged, on the activity line |
| Sort control | Component | Ship | `macos/App/DesignSystem/Components/Select/` `PVSelect` | Two more options |
| Caption band | Snowflake | Extend | `SourcesListView` `captionBand` | Gains a "Published" caption at Wide |
| Columns | Snowflake | Extend | `macos/App/Features/Sources/SourcesListColumns.swift` `SourcesListColumns` | Learns the width class: Type 150 and Published 150 at Wide |
| Wide list column | Token | New | `macos/App/DesignSystem/Tokens/PVSpacing.swift` `widthListWideMax` | 1600; second call site: Promote's list (`PromoteLayout.listMaxWide`) |
| Side hairlines | Component | Ship | `macos/App/DesignSystem/Components/Divider/` `PVDivider` | Vertical, at the centered column's sides at Wide |
| Tooltip | Platform | Ship | `.help` | "Published {full date}" |

**Non-goals.** No new kit component: the published date is part of the row's snowflake; the only kit addition is the width token, which Promote already needs. No icon on the date, no Private glyph, no filter control, no group headings.

---

## 7. Handoff

When the board is agreed: record the link, the frames and the board's decisions on the **Board** line here, in [`README.md`](README.md) and on S15-04's plan **Board** line; the plan copies SL-1 – SL-7, SL-W1 and SL-W2 and moves from `Blocked — on board S15-D2` to `Planned`. When S15-04 lands, the brief moves to the spike's design archive.
