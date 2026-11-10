# S15-D2 — Sources list: the Source's date on each row and as a sort

**Kind:** Claude Design board (**enhancement** of the Sources list as shipped: S2-04 / S12-D5 list, three columns Source · Type · Evidence graph, the S12-14 toolbar. Draw on the existing Sources list frames; add the date to the row's second line and two options to the Sort menu, and give the list its first Wide frame)
**Spike:** Provenencia Spike 15, slice 1
**View:** the Sources list (the workspace's Sources destination: header, toolbar, caption band, rows)
**Implements later as:** PR **S15-04** (SL-1 – SL-7, SL-W1, SL-W2)
**Board:** — (not drawn yet). When agreed, a Claude Design artifact link, recorded here, in [`README.md`](README.md) and on S15-04's plan **Board** line (user, 2026-11-10)
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
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
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

**What changes.**
- **The Source's own date on each row.** A Source has one optional date: when the record was made or first published ("1880", "abt 12 Mar 1880", "bet 1834 and 1840", or a phrase "in the reign of Victoria"). The Source page shows it in its header with a 14pt calendar icon (board S15-D1, SD-1). The list row shows the same date, written the same way, on the Source cell's second line (user, 2026-11-10). An undated Source's row reads a muted "Undated" in that place (user, 2026-11-10).
- **Two date sorts.** The Sort menu gains "Source date, oldest first" and "Source date, newest first" (user, 2026-11-10). Undated Sources come last in both.
- **Wide.** The list gets its Wide frame (SL-W1).

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

| ID | Requirement |
| --- | --- |
| SL-1 | **The date on the second line.** A dated Source's second line reads: (type, narrow lists only) · ref · date · last update. The date is caption-size muted text after a 12pt calendar icon, matching the last update's icon size, so the line reads as one row of facts. It is text, not a Chip or Badge. **The board decides** the icon's tone and the exact spacing. |
| SL-2 | **Undated.** An undated Source shows the same calendar icon and the word "Undated" in the date's place, in the faintest text tone, so a dated row stands out from an undated one at a glance. |
| SL-3 | **A long date.** A long range or phrase ("bet 12 Mar 1834 and 3 Sep 1840", "in the reign of Victoria, before the fire") truncates with an ellipsis and shows the full date as a tooltip; the ref never truncates; the last update keeps its place. **The board decides** which item gives way first when the line is too narrow (recommended: the type (narrow lists), then the date, then the last update). The row never wraps its second line. |
| SL-4 | **The Sort menu.** Two options after "Reverse alphabetical": "Source date, oldest first" and "Source date, newest first". The control reads "Sorted by source date, oldest first" when chosen. No icon change. |
| SL-5 | **The order.** Under "oldest first": `bef. 1830`, `1830`, `abt 1834`, `bet 1834 and 1840`, `1880`, `aft. 1880`, then the undated rows by title. Under "newest first": the dated rows in reverse, then the same undated rows by title. A phrase-only date sits among the undated rows. |
| SL-6 | **Empty and filtered.** A filter or search that leaves rows keeps the chosen date sort. The filter-empty state and the catalog-empty state are unchanged. |
| SL-7 | **Keyboard and VoiceOver.** The row's link reads "1880 U.S. census, Ottawa County, Census, SRC-4KD2P, 1 Jun 1880, updated 2 days ago" (and "Undated" in the date's place for an undated Source). The Sort options are reachable by keyboard as today. |
| SL-W1 | **Wide: the list fills the width.** At Wide the toolbar, the caption band and the rows run the pane's full width between the page gutters, as lists do. The Source cell takes the extra width; the Type (170pt) and Evidence graph (420pt) cells keep their size; the second line, date included, never stretches its items apart. No second column: the list beside a selected Source is a later redesign. |
| SL-W2 | **Regular: as today plus the date.** At Regular the three columns as today with the date on the second line. Below 880pt the narrow arrangement, with the type first on the second line. Crossing 1600pt or 880pt keeps the scroll position and the chosen sort. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Sources list | Header, toolbar, caption band, rows; three columns (Source fills, Type 170, Evidence graph 420); below 880 the Type column joins the second line | The same, filling the pane between the gutters; the Source cell takes the extra width; Type and Evidence graph keep their widths | Nothing moves: a list fills the width (rule 3). The extra width lets long titles and long dates show in full |

---

## 5. Suggested frames

Frames with an a / b pair are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others at Regular. Data: twelve Sources, invented, of types Census, Parish register, Photograph, Letter and Newspaper. Dated: "1880 U.S. census, Ottawa County, Kansas" (`1 Jun 1880`), "St Brigid parish baptisms" (`bet 1834 and 1840`), "Robins family portrait" (`abt 1895`), "Letter from Ada to her sister" (`bef 3 Mar 1830`), "Ottawa County Herald, 4 July issue" (`4 Jul 1881`), "Notes on the old mill" (phrase "in the reign of Victoria"); six undated. Invent every person and record.

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | SL-1, SL-2, SL-W1, SL-W2 | The list sorted by Date added: dated and undated rows mixed |
| 2a / 2b | SL-4, SL-5 | The list sorted "Source date, oldest first": the dated rows in order, the phrase-only and undated rows last; the Sort control reading "Sorted by source date, oldest first" |
| 3 | SL-4 | The Sort menu open, six options, "Source date, oldest first" checked |
| 4 | SL-5 | "Source date, newest first" |
| 5 | SL-3, SL-W2 | The list at 860pt (narrow): type first on the second line; one long range date truncated with its tooltip shown |
| 6 | SL-6 | A type filter "Census" with the date sort kept |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sources list row | Snowflake | Extend | `macos/App/Features/Sources/SourcesListRow.swift` `SourcesListRow`, `SourcesLinkLabel` | The date joins `SourcesLinkLabel`'s second line between the ref and the last update. No new cell, no new column |
| Row date | Snowflake (inside the row) | Extend (the row) | `macos/App/Features/Sources/SourcesListRow.swift` | Icon + caption text through the app's row date display (`DateRowDisplay`), the same text the Source page header shows. Not a Chip, not a Badge, not `PVListDetailCell` (that one is the last update's relative time) |
| Calendar icon | Component | Ship | `macos/App/DesignSystem/Components/Icon/` `PVIcon(.calendar)` | 12pt, as the last update's icon |
| Separator dot | Snowflake | Ship | `SourcesLinkLabel.separator` | Between each item, as today |
| Sort control | Component | Ship | `macos/App/DesignSystem/Components/Select/` `PVSelect` | Two more options; no new control |
| Caption band | Snowflake | Ship | `SourcesListView` `captionBand` | Unchanged: no Date caption |
| Columns | Snowflake | Ship | `macos/App/Features/Sources/SourcesListColumns.swift` `SourcesListColumns` | Unchanged widths; Wide changes nothing in them |
| Tooltip | Platform | Ship | `.help` | The full date on a truncated date |

**Non-goals.** No new kit primitive: the date is part of the row's snowflake. No Date column, no Private glyph, no filter control, no group headings.

---

## 7. Handoff

When the board is agreed: record the link, the frames and the board's decisions (SL-1's tone and spacing, SL-3's give-way order) on the **Board** line here, in [`README.md`](README.md) and on S15-04's plan **Board** line; the plan copies SL-1 – SL-7, SL-W1 and SL-W2 and moves from `Blocked — on board S15-D2` to `Planned`. When S15-04 lands, the brief moves to the spike's design archive.
