# S13-D1 — Person page: header, section switching, Vitals

> **Update 2026-10-29: Wide layout.** Since the first version of this brief, the design system gained a Wide width class (a pane of 1600pt and up) and a **Page columns** component (Spike 14 brief S14-D1, archived by [#629](https://github.com/mendahu/provenencia/pull/629), board agreed 2026-10-29). If you already started this board, **extend your frames**; do not start a second page. For this revision:
> - **Pull the fresh design-system kit first.** It has Page columns, the `--width-wide-breakpoint` and `--width-secondary-max` tokens, and the Width classes README section.
> - The **"Two widths — draw both"** block is now in the Paste block, and the brief gains Page columns in the domain facts and inventory, requirements **PPF-W1 – PPF-W5**, a **Wide layout** table (§ 4.1), and Wide frames **14 – 19**.
> - PPF-3 now places the header and switcher in Page columns' header slot, pinned or scrolling; PPF-13 asks for Tab between switcher items, not arrow keys.

**Kind:** Claude Design board (**rethink** of the Person page layout below the header; the header itself is an **enhancement**)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail: the fixed header, the section switcher, and the five sections as places
**Implements later as:** PR **S13-10** (header, switcher, Vitals, the four other sections as places). Later boards fill the frame: S13-D2 (provenance sheet, S13-11), S13-D3 (Family, S13-12), S13-D4 (Timeline, S13-13), S13-D5 (Gallery, S13-14 and S13-15), S13-D6 (Sources, S13-16 and S13-17), S13-D10 (how the records connect, S13-28).
**Board:** not drawn yet. When agreed, record here: the Claude Design link, the frame numbers, and every change from this brief with the user's answer.
**Depends on:** S13-09 (landed [#624](https://github.com/mendahu/provenencia/pull/624)): the page already places the header, then a below-header slot. The agreed S10-D12 Person board (header thumb, gender Badge, likeness gallery, portrait). The agreed S14-D1 board (Page columns and the Wide width class, agreed 2026-10-29, [archived brief](../../spike-14/design/archive/S14-D1-wide-layout.md)): pull the fresh kit, which has Page columns.
**Related:** [`../person-detail.md`](../person-detail.md) R-1 – R-7, R-14, R-15, R-D1 – R-D3; [`../deployment-plan.md`](../deployment-plan.md) working rules "S13-D1 draws the whole frame", "Sections are places", "One sectioned Conclusion page, not a Person fork", "Redesigns adopt Page columns"; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide) (Regular and Wide); plan [`../prs/S13-10-header-and-sections.md`](../prs/S13-10-header-and-sections.md).
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

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### Two widths — draw both

The window opens at 1366 × 900pt but is often stretched across a large monitor.
Every view has two width classes, measured on the view's own pane (the window
minus the sidebar), not the screen. A view with a side pane (an inspector, a detail
pane) plans that pane inside its Wide frame:

- **Regular** — pane below 1600pt. Draw at the default window: 1366 × 900, sidebar open.
- **Wide** — pane 1600pt and up. Draw at 2560 × 1440, sidebar open.

Rules for Wide:

1. **The extra space does a job.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. If the view truly has nothing to add, center its content. Content jammed to one side with empty space beside it is never the Wide design.
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is a rethink below the header

Below the header, replace the old frames: today's single column (Details, then Likenesses) becomes five sections. Do not keep the old single column as a frame to ship. The header is an enhancement: keep its parts as the S10-D12 board drew them (thumb, gender Badge, name, b. and d. lines, ref and member count), and change only how it sits on the page.

This board draws the **whole frame**, at Regular and at Wide, on the kit's **Page columns** component: the header, the switcher, and all five sections as named places. Vitals is drawn in full. Family, Timeline, Gallery and Sources are drawn as they ship in S13-10 (see PPF-9 and PPF-10). Later boards fill those sections inside this frame; they do not move the header or the switcher.

---

## 1. Objective

A researcher opens a Person and sees, at a glance, who the person is (the header), and can move between five views of them (Vitals, Family, Timeline, Gallery, Sources) in one action from anywhere on the page. Back and Forward walk the sections the researcher visited.

**What changes.** The page is laid out on the kit's Page columns (a header slot, a primary column and an optional secondary column, from the S14-D1 board), so it has a Regular and a Wide arrangement. A section switcher sits under the header. The header stays the same whichever section shows. Today's Details rows become the Vitals section. Today's likeness grid moves, unchanged, into the Gallery section. Family, Timeline and Sources are places with an empty state until their PRs land.

**What stays.** Every part of the header; every Details row with its state Badge, Sources and Why disclosure (the provenance sheet replaces the Why in S13-D2, not here); the likeness tiles with Use as portrait, the Portrait Badge and open Citation; the header thumb's right-click menu "Use earliest likeness" (S13-D5 decides where it goes); the Event and Place pages, which keep today's layout (R-1).

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| One canonical person | Everything on the page is a promoted (canonical) person and what is concluded about them. Nothing unpromoted is drawn (R-2). |
| The header is display only (R-4) | Name, the b. and d. lines (date · place), the 120pt thumb (lead likeness or the person mark, with the gender tint from S10-D12), the gender Badge above the name, the ref and the member count at the right. No new buttons. The thumb's existing right-click "Use earliest likeness" stays until the Gallery board. |
| Five sections (R-5) | Vitals, Family, Timeline, Gallery, Sources, in that order. |
| A section is a place (R-6a) | Choosing a section is a navigation step. It goes into Back / Forward history, and quitting and reopening the app returns to it. A link to a Person from elsewhere (the Persons list, the omnibar, an Evidence graph card, later a family tile) opens on Vitals. |
| Vitals is today's rows (R-14, R-15) | Name and the four life rows (birth date, birth place, death date, death place) always show, stated empty when nothing is recorded. Every other Person property shows once a record speaks to it, in binding order. Each row: label, value, a state Badge (Mixed, Merged, …), its Sources, and the Why disclosure. A row with many values can be long; a person can have 20 rows. |
| The provenance control comes next (S13-D2) | Every Vitals value will get a control beside it that opens a sheet (R-8, R-13a). Leave room on the row's trailing edge; do not draw the sheet. |
| Gallery is today's grid until S13-D5 | 96pt crops in a wrapping grid, captioned with the Source title, earliest first; one carries the Portrait Badge, the others a ghost "Use as portrait" button; a click opens the Citation. A person with no likeness has an empty Gallery. |
| Family, Timeline, Sources are not built yet | Their content comes with S13-D3, S13-D4 and S13-D6. Until then each is a place that says the section is coming. The researcher can still choose it, and history still records it. |
| Some persons are sparse | A person may have only a ref (no name, no dates, no likeness, one member). The header then shows the ref as the title. |
| The page loads and can fail | Today: a small spinner while the first read loads; an error Callout if it fails; a spinner in the Details header while a stale page refreshes. |
| The window can be narrow | The content column can be as narrow as the workspace allows with the sidebar open. Five section names must still fit or degrade cleanly. |
| Two widths (§ "Two widths — draw both") | Regular is a pane below 1600pt, Wide 1600pt and up. The app measures the pane once; the page reads its class. Most researchers work at Regular (every MacBook full screen is Regular); a 1920pt-or-wider monitor is Wide. |
| Page columns is the page scaffold (S14-D1) | Three slots: an optional **header** that spans the column group and either scrolls with the page or is pinned above the scroll; a **primary** column (up to 1180pt); an optional **secondary** column (340 – 480pt). At Regular the secondary stacks below the primary at the same width; at Wide it sits on the trailing side, 32pt from the primary. With no secondary, the column is centered at both widths. Page columns owns the scroll. The Event and Place pages move onto it too, with no secondary and no other change. |
| A section fills the primary column | The section the researcher chose is the page's primary content. What, if anything, goes in the secondary column is this board's call, per section (PPF-W2 – PPF-W4), within rule 4: anything in it at Wide is stacked below the primary at Regular. |

### 2.1 What this board is not

- The provenance sheet or its control (S13-D2).
- Family, Timeline, Gallery or Sources content beyond PPF-9 and PPF-10 (S13-D3 – S13-D6).
- How the records connect (S13-D10; that board decides whether it sits inside Sources or beside it).
- Event or Place pages in sections (R-1).
- Editing anything from the header.
- A sixth section, or sections the researcher can reorder or hide.
- New Page columns behavior: the component is drawn on the S14-D1 board. This board places the Person page on it; it does not add a slot, a mode or a full-width variant. A section that would need one says so as a note for its own board.
- Wide frames for the Event and Place pages beyond "the same centered column as today" (their own redesigns).

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-10** | PPF-1 … PPF-14 and PPF-W1 … PPF-W5. The header placement, the switcher, section history, Vitals, Gallery with today's grid, the three empty sections, loading, error and refreshing states, at Regular and Wide. S13-10 also builds the Page columns component in the Swift kit from the S14-D1 board, and moves the Event and Place pages onto it. |
| Later boards | S13-D2 adds the provenance control to Vitals rows. S13-D3, S13-D4, S13-D5, S13-D6 replace PPF-9 and PPF-10 with each section's content, inside this frame. |
| Stays out | Event and Place page redesigns (they keep today's look on Page columns). Any engine read. The provenance sheet. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PPF-1 | **The header is the same in every section.** Same parts as today (thumb, Badge, name, b. and d. lines, ref, member count). Switching sections never moves, resizes or redraws it. |
| PPF-2 | **Explore the switching pattern (R-7).** Draw at least two patterns on Vitals with a long person (20 rows): (A) a switcher with one item per section that shows one section at a time, and (B) the five sections stacked on one scroll with jump shortcuts. Recommend one, and say why, against PPF-3, PPF-4 and PPF-12. The rest of the board draws the chosen pattern. |
| PPF-3 | **One action from anywhere (R-6).** From any scroll position in any section, the researcher reaches any other section in one click. Decide whether the header and switcher stay pinned while the section scrolls, or scroll with the page and the switcher stays reachable another way. Both go in Page columns' **header** slot, which is pinned or scrolling by the page's choice (S14-D1 WL-3, WL-8); do not build a second pinned bar. If pinned, decide whether the header compacts on scroll (for example the name and b./d. line only, thumb smaller) so a 120pt header does not take a third of a short window. Draw the scrolled state. |
| PPF-4 | **The current section is plain to see.** The switcher marks the current section without relying on colour alone, and VoiceOver reads it as selected ("Vitals, selected, 1 of 5"). |
| PPF-5 | **Sections are places.** Choosing a section adds one Back step. Back returns to the previous section on the same person, then to wherever the researcher came from. Choosing the section already showing adds nothing. Opening a Person from anywhere lands on Vitals. |
| PPF-6 | **History says which section.** The breadcrumbs and the Back / Forward jump menu tell two sections of one person apart. Recommended: the breadcrumbs read *Persons › Jane Doe › Family*, with *Jane Doe* going to Vitals, and Vitals shows no third crumb; the jump menu row reads *Jane Doe · Family*, and a Vitals entry reads *Jane Doe* alone, as today. Draw the toolbar for one frame. |
| PPF-7 | **Each section opens at its top.** Switching to a section shows it from its first line; the page does not keep one section's scroll for another. |
| PPF-8 | **Vitals is today's Details.** The rows, their order, Badges, Sources and Why disclosures are as today. Decide whether the section keeps a "Details" heading, takes "Vitals" as its heading, or has none because the switcher already names it. Keep the refreshing spinner somewhere visible while a stale page reloads. Leave room on each row's trailing edge for S13-D2's provenance control. |
| PPF-9 | **Gallery shows today's grid.** The Likenesses grid as S10-D12 drew it (count, 96pt crops, Source captions, Portrait Badge, Use as portrait, open Citation) moves into Gallery unchanged. A person with no likeness shows an EmptyState: "No likenesses yet" with one line on where they come from (a likeness is drawn on a Citation's image). |
| PPF-10 | **Family, Timeline and Sources say they are coming.** Each shows an EmptyState with the section's glyph (Family: branch, Timeline: calendar, Sources: library; existing kit symbols, no new mark), its name and one muted line that the section arrives in a later version. Same component and layout for the three; only the words differ. No fake content, no skeleton rows. |
| PPF-11 | **States.** First load: the header area and the switcher are not drawn until the person loads (today's single spinner), or draw their loading state from the kit. Failure: today's danger Callout in place of the page. Refreshing a stale page: the spinner from PPF-8. A section's own load and error states arrive with its later board. |
| PPF-12 | **Narrow window.** At the narrowest Regular content column, all five section names stay readable (no truncation to "Fam…") or the switcher degrades in a way the board shows. Draw it. |
| PPF-13 | **Keyboard.** Every section is reachable from the keyboard: each switcher item takes focus with Tab and shows the kit focus ring, and ⌘1 – ⌘5 choose Vitals … Sources while a Person page is showing (no other screen uses ⌘1 – ⌘5 today). Confirm or replace the shortcuts. |
| PPF-14 | **Dark mode.** The header, the switcher in its selected and unselected states, Vitals and one empty section in dark mode. |
| PPF-W1 | **The frame at Wide.** The page sits on Page columns at both widths. The header and the switcher go in its header slot and span the column group (WL-8), pinned or scrolling as PPF-3 decides, at Regular and Wide alike. The header keeps its Regular measure: its name and lines do not stretch (rule 2). |
| PPF-W2 | **Vitals at Wide.** Vitals' rows are the primary column at its Regular measure (rule 2). Decide what the secondary column holds on Vitals, if anything, from what the page already has: for example the likeness tiles or the member count and ref as supporting content. Rule 4 binds: whatever sits there at Wide is stacked below the rows at Regular, in the same order. If nothing earns the column, Vitals is primary-only and centered (WL-6); say so in one clause and leave the column for S13-D2's provenance sheet and S13-D6's Sources, which are the likely occupants. Recommended: primary-only for now. |
| PPF-W3 | **Gallery at Wide.** The likeness grid is a grid (rule 3): at Wide it adds columns at the same 96pt tile size. Decide whether it fills the primary column (up to 1180pt, primary-only) or the whole column group; the second needs a full-width Page columns mode that does not exist, so it is a note for S13-D5, not this board. Recommended: primary-only, centered. |
| PPF-W4 | **Empty sections at Wide.** Family, Timeline, Sources and the empty Gallery show their EmptyState centered in the primary column, primary-only (WL-6). |
| PPF-W5 | **Crossing 1600 keeps the place (rule 5).** Resizing across the line keeps the section, the scroll position within it and keyboard focus. Show it as a note on the Wide Vitals frame. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Person page frame | Page columns: header slot (header + switcher), one centered column up to 1180 | Same header slot spanning the column group; primary column up to 1180; secondary column per section | Nothing leaves the header slot; the secondary column is decided per section (PPF-W2 – PPF-W4) |
| Vitals | Rows in the column; anything the board puts in the secondary column stacked below the rows | Rows in the primary column; the secondary column if PPF-W2 fills it, otherwise centered primary-only | Per PPF-W2 |
| Gallery | Today's grid in the column | The grid in the primary column with more tiles per row | Nothing; more columns (rule 3) |
| Family, Timeline, Sources (empty) | EmptyState in the column | EmptyState centered, primary-only | Nothing (nothing to add yet; their boards draw their own Wide arrangement) |
| Event and Place pages | Today's page on Page columns, primary-only | The same centered column | Nothing (their own redesigns) |

---

## 5. Suggested frames

One Person page. Every later frame is this page with one thing changed. Draw each at **Regular** (1366 × 900, sidebar open) unless it says Wide; the Wide frames (14 – 18) sit beside their Regular frame.

1. **Pattern A on Vitals.** A full person (name, both life lines, gender Woman, three likenesses, 20 Vitals rows, one Mixed row) on Vitals, switcher pattern A (PPF-2).
2. **Pattern B on Vitals.** The same person, stacked sections with jump shortcuts (PPF-2). Then the recommendation.
3. **Scrolled.** Frame 1 (or the chosen pattern) scrolled to the last Vitals row: the header and the switcher as PPF-3 decides.
4. **Gallery.** Frame 1's person on Gallery: three tiles, the second the portrait (PPF-9).
5. **Gallery, none.** A person with no likeness on Gallery: the EmptyState (PPF-9).
6. **Family, Timeline, Sources.** Each section's empty state (PPF-10). One frame may show all three side by side.
7. **Sparse person.** Only a ref, no name, no dates, no likeness, on Vitals: the ref as title, the four life rows stated empty.
8. **History.** Frame 4 with the toolbar: breadcrumbs *Persons › Jane Doe › Gallery*, the Back jump menu open listing *Jane Doe* (the Vitals entry) and *Persons* (PPF-6).
9. **Loading and failure.** The first-load state and the error Callout (PPF-11).
10. **Refreshing.** Frame 1 while a stale page reloads (PPF-8).
11. **Narrow.** Frame 1 at the narrowest content column with the sidebar open (PPF-12).
12. **Focus.** The switcher with keyboard focus on Timeline (PPF-13).
13. **Dark mode.** Frames 1, 3 and 6 in dark mode (PPF-14).
14. **Wide Vitals.** Frame 1 (the chosen pattern) at 2560 × 1440: the header slot spanning the group, Vitals in the primary column, the secondary column as PPF-W2 decides; the rule-5 note (PPF-W1, PPF-W2, PPF-W5).
15. **Wide Vitals, growing.** Frame 14 at 1920 × 1080 (pane 1656), so the secondary column's 380pt width shows, if PPF-W2 uses one; otherwise the centered column (WL-5, WL-6).
16. **Wide, scrolled.** Frame 3 at 2560 × 1440: the header slot pinned or scrolled as PPF-3 decides (PPF-W1).
17. **Wide Gallery.** Frame 4 at 2560 × 1440: more tiles per row, same tile size (PPF-W3).
18. **Wide empty section.** Frame 6's Family at 2560 × 1440: the EmptyState centered (PPF-W4).
19. **Wide dark.** Frame 14 in dark mode.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Shared Conclusion page | Snowflake | **Extend** | `Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | The one page for Person, Event and Place. Person passes its own below-header fill through the `pageSections` slot (S13-09). Its `page` scaffold moves onto Page columns; Event and Place pass no secondary and look as today (R-1). Not a Person-only page. |
| Page columns | Component | **New** (agreed on S14-D1; built in S13-10) | `DesignSystem/Components/PageColumns/` *(new)* | PPF-W1 – PPF-W4. Header slot (pinned or scrolling), primary, optional secondary; owns the scroll. Draw from the fresh kit's component. Do not add slots, modes or a full-width variant on this board. Callers: the Person page, and the Event and Place pages through the same `ConclusionDetailPage.page` in S13-10; next, the Conclusion lists and the Source page in their own redesigns (S14-D1 inventory). |
| Width class | Component | Ship | `DesignSystem/Components/WidthClass/PVWidthClass.swift` | Not drawn. Regular or Wide, measured once by the workspace; Page columns reads it. |
| Header | Snowflake | Ship | same file `ConclusionDetailHeader` | PPF-1, PPF-W1. Unchanged parts, in Page columns' header slot with the switcher. A compact form (PPF-3) is a variant of this view, not a second header. |
| Gender Badge, thumb, b./d. lines | Component / snowflake | Ship | `PVBadge`, `CachedThumbnail` over `PVThumbnail`, `ConclusionDatePlaceLine` | PPF-1. As S10-D12. |
| Section switcher (pattern A) | Component | **Extend** | `PVChipGroup(style: .segmented)` with one `PVChip` per section (`DesignSystem/Components/Chip/PVChip.swift`) | PPF-2, PPF-4, PPF-13. The segmented track the date-kind picker and the Properties value-type picker use. Each item is a section name; the selected chip lifts. Extend: `PVChip` gains the selected accessibility trait and the kit focus ring, for every chip in the app. Not a new Tab component. If the board needs icons beside the names, use `PVChip`'s existing content; do not draw a tab bar. |
| Narrow-window fallback | Component | Ship | `PVSelect` (`DesignSystem/Components/Select/PVSelect.swift`), or the same chips with shorter labels | PPF-12. Only if five names cannot fit. One of these two; nothing else. |
| Section jump shortcuts (pattern B) | Component | Ship | `PVButton` ghost, small, in a row | PPF-2. Only if pattern B wins. |
| Section title inside a section | Component | Ship | `PVSectionHeader` | PPF-8, PPF-9. Title, count as meta, refreshing spinner as aside. |
| Vitals rows | Snowflake | Ship | `Features/Conclusions/ReconciledValueRow.swift` `ReconciledValueRow` | PPF-8. Unchanged; the Why disclosure stays until S13-D2. |
| Gallery grid and tiles | Snowflake | Ship | `ConclusionDetailPage.swift` `ConclusionDetailLikenessSection` | PPF-9. Moved, not redrawn. S13-D5 redesigns it. |
| Use as portrait, Portrait Badge, chosen toast, thumb menu | Component | Ship | `PVButton` ghost, `PVBadge` accent, `PVToast`, the header thumb's right-click menu | PPF-9. As S10-D12; the thumb menu stays on the header (§1). |
| Empty section, no likeness | Component | Ship | `PVEmptyState` (`DesignSystem/Components/EmptyState/`) | PPF-9, PPF-10. With a glyph, title and message. Same component for all four empties. Glyphs are existing `PVSymbol`s: Family `gitBranch`, Timeline `calendar`, Gallery `photo`, Sources `library`. |
| Loading, failure, refreshing | Component | Ship | `ProgressView` small, `PVCallout` danger | PPF-11, PPF-8. As today. |
| Breadcrumbs, jump menu | Component | Ship | `PVBreadcrumbs`, the toolbar jump menu (`Features/Workspace/WorkspaceToolbar.swift`) | PPF-6. Add a crumb and a row label; no new toolbar control. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A Tab or TabBar kit component, or a native `TabView` | The segmented `PVChipGroup` already switches between a few exclusive options. One page is not a second call site. |
| A Person-only page type beside `ConclusionDetailPage` | Working rule "One sectioned Conclusion page, not a Person fork". |
| Placeholder content (fake relatives, sample events, skeleton rows) in empty sections | They read as data. The empty state says the section is coming. |
| A sidebar or outline of sections beside the page | The switcher is the one way to switch; the workspace sidebar stays as it is. At Wide the secondary column is supporting content (rule 4, WL-9), not a section navigator. |
| A Wide-only element, or text and rows stretched to fill a Wide pane | Rules 2 and 4: Wide adds columns; it does not add content or widen the measure. |
| A Page columns full-width mode, a third column, or a second pinned bar | Page columns is agreed (S14-D1). A section that needs more is a note for its own board. |
| Section controls in the header | R-4: the header is display only. |
| An overflow or More menu, or a horizontally scrolling tab strip, for a narrow window | PPF-12 falls back to `PVSelect` or shorter labels; a second switching control for one case is not added. |
| A new mark or icon for an empty section | The four glyphs are existing kit symbols. |
| Remembering the last section per person | A link to a Person opens on Vitals (PPF-5); history is how the researcher returns. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design link, the chosen pattern (PPF-2), the pinned or scrolling answer (PPF-3), what Vitals' secondary column holds at Wide (PPF-W2) and the Gallery's Wide answer (PPF-W3), the section heading answer (PPF-8), the toolbar wording (PPF-6), the shortcuts (PPF-13), and every other change from this brief with the user's answer.
2. Refresh the [S13-10 plan](../prs/S13-10-header-and-sections.md): copy the answers into its Decisions and **Design → macOS**, close its Open questions, run `review-plan` again, and set it to `Planned`.
3. S13-10 archives this brief under `design/archive/` *(new)* and records it in [`../completed.md`](../completed.md). Later boards (S13-D2 – S13-D6, S13-D10) extend these frames.
