# S13-D13 — The provenance sheet in the design system

**Kind:** Claude Design board (**enhancement** of the Provenencia design-system project: one new recipe specimen, the provenance sheet, moved from the Person page board into the kit)
**Spike:** Provenencia Spike 13 (Person detail page), Slice 5 — Gallery (preparatory)
**View:** The design system itself: a Provenance sheet specimen page in the kit, with its anatomy, slots, variants and states. No product screen changes.
**Implements later as:** PR **S13-36** (the provenance sheet joins the design system), before S13-14.
**Board:** not drawn yet. When agreed, record the Claude Design project, the specimen's file, turn and frames here, with every answer to § 7.
**Depends on:** the agreed S13-D2 board (the sheet as it ships; [archived brief](archive/S13-D2-provenance-sheet.md), built by S13-11 [#650](https://github.com/mendahu/provenencia/pull/650)); the agreed S13-D12 board (owner group lines; [archived brief](archive/S13-D12-life-rows-why.md)).
**Related:** [`../person-detail.md`](../person-detail.md) R-8 – R-15, R-13b, R-32; [`../deployment-plan.md`](../deployment-plan.md) working rules "The provenance sheet is a kit recipe", "The person card lives in the design system; display is decoupled from content"; brief [S13-D5](S13-D5-gallery-and-viewer.md) (the Gallery's one-record sheet and the viewer's record row, which instance this recipe); [`../../../ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Family trees and person pages" (a future caller).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** — (not reviewed yet)

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
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Work in the **Provenencia design-system project**, not on a product board. Add one specimen page to the kit: **Provenance sheet**, a recipe. Draw it from the agreed S13-D2 frames (the Person Detail board, turn 6): its look does not change. Do not redraw the Person page, and do not start a second sheet on the product boards. After this board, product boards (S13-D5 now, a family tree later) instance the kit's sheet instead of drawing their own.

---

## 1. Objective

The provenance sheet answers "where does this come from?" for one value: the records behind it, what the auto-reconciler did with each, and links to each record's Source, Citation and Observation. Today it is drawn and built for the Vitals rows of the Person page only. The Gallery needs it for one likeness (S13-D5, user 2026-11-03), and a family tree will need it for a relationship or a relative's fact. This board turns the sheet into a **kit recipe** with named slots, so every caller instances one sheet.

**What changes.** The sheet becomes a specimen in the kit: its anatomy and slot names (§ 4), its variants (many records, one record, owner groups, value groups), its states, and how a caller fills it. The one-record variant's header is decided here (PSK-4), so the Gallery board instances it.

**What stays.** Everything the sheet looks like and does today (S13-D2, S13-D12): the Panel, the header, the record table and its columns, the owner and value group lines, the three links per record, the reserved "Concluded by you" place, Done, the sizes at Regular and Wide, keyboard and VoiceOver. The Vitals provenance control (the info IconButton). The Event and Place pages' Why accordion, which keeps its own look (working rule "The provenance sheet replaces the Why on the Person page only").

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A value and its records | A value (a birth date, an occupation, a likeness) is concluded from **records**: Observations on Citations. Each record has what it was **read as** ("14 May 1817"), its **citation line** ("Parish register, Vol. 3, Baptism of Ada, p. 23"), its **outcome** (kept, folded, outvoted, weak, denied, against, no usable value) with a mark, its **Observation ref** and **Subject ref**, and three destinations: the Source, the Citation, the Observation. |
| Groups | Records sit under **owner** group lines when they come from several entities (a birth date read off two birth events, S13-D12), and under **value** group lines when they went into different values (a mixed row). Each group line is a heading. |
| The header | A title (the value's label: "Birth date", "Likeness from Hamburger Adressbuch 1861"), the shown value or values in their face (mono for a date, italic for a place) or the empty text in italic, the state Badge (merged, mixed, multiple) when there is one, and a summary line ("merged from 3 Sources · 4 records"). |
| The kit names no product types | A kit recipe takes **display-ready content**: strings, faces, Badges and link actions. The caller (a page) builds that content from its own data and decides what a link opens. The sheet never knows what a Person, a Source or a navigation place is. |
| Callers (known) | **Vitals rows** on the Person page (today); **a Gallery likeness** (S13-14: one record, no Badge); **the viewer** (S13-15: the record table under the image, without the Panel); **a family tree** (later: a relationship or a relative's fact). The Event and Place pages adopt it with their sections. |
| Opening and closing | A caller opens the sheet from its own control (the info IconButton on Vitals and Gallery). Escape, a click outside and Done close it; a link closes it, then the caller navigates. Focus returns to the control that opened it. |
| Sizes | One record to 30+ (records scroll inside the sheet past a cap; the title, header and Done stay). Citation lines are long and wrap; nothing is truncated. |

### 2.1 What this board is not

- A redesign of the sheet. Its look is S13-D2's.
- The Why accordion on the Event and Place pages.
- "Concluded by you" (Reconciliation Claims): its place stays reserved and empty.
- A family tree, or any product screen.
- New tokens, a new Panel, or a new table component.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-36** | PSK-1 … PSK-9 and PSK-W1: the sheet as a kit recipe and its record table as a kit part, filled by display-ready content; the Vitals rows moved onto it with no visible change. |
| Used by | S13-14 (the Gallery's one-record sheet, PSK-4), S13-15 (the record table under the viewer's image, PSK-6), later a family tree. |
| Stays out | Any visible change to Vitals; the Event and Place Why accordion. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PSK-1 | **Specimen page.** A kit page "Provenance sheet" (a recipe, beside the Person card) showing the sheet's anatomy with every slot named: **title**, **value** (one value, several, or the empty text), **state Badge** (optional), **summary** (optional), **reserved block** (the "Concluded by you" place, zero height), **record table** (PSK-5), **footer** (Done). Label each slot on the specimen. |
| PSK-2 | **Many records (the Vitals case).** The Birth date sheet from S13-D2: two owner groups (the shown owner first, marked "· shown"), five records, one outvoted. Unchanged from the S13-D2 frames. |
| PSK-3 | **Value groups.** A mixed Occupation sheet with two value groups and a group of records that went into no value. Unchanged from S13-D2 (PRV-6), headings only when a group has two or more buckets. |
| PSK-4 | **One record (the Gallery case; user, 2026-11-03).** The sheet for one record: the title names the thing ("Likeness from Hamburger Adressbuch 1861"); no value line, or the value in its face if the board prefers (say which); no state Badge; the summary "1 record"; the table with one record and no group lines. Draw it at Regular and Wide. This is the frame S13-D5's Gallery tile opens (its frame 2b instances it). |
| PSK-5 | **The record table.** Read as · Citation (the citation line, a link in ink that opens the Citation) · Outcome (mark and phrase) · Record (Observation and Subject refs, mono caption, muted) · the reserved 88pt column; the Source and Observation links under the citation line; primary ink for a counted record, secondary for the rest. As S13-D2 PRV-5, PRV-9, PRV-15 and the board's notes. Name it as the sheet's part: a caller can place it without the Panel (PSK-6). |
| PSK-6 | **The table alone (the viewer case).** The record table with one record, no column headings, placed under an image inside another Panel, at the viewer's width (S13-D5 LV-4). Draw it as a usage frame, not a new component. |
| PSK-7 | **Long.** 30 records: the table scrolls inside the sheet past its cap; title, header and Done stay. |
| PSK-8 | **Keyboard and VoiceOver.** As S13-D2 PRV-12, PRV-13: Tab through the citation line, Source and Observation links in reading order; the title and each group line are headings; each record reads as one element. A note on the specimen. |
| PSK-9 | **Dark mode.** PSK-2 and PSK-4 in dark. |
| PSK-W1 | **Wide.** The sheet's two widths (880pt Regular, 1280pt Wide, S13-D2), Read as capped at 220pt and Outcome at 280pt at Wide, the Citation column taking the rest. A resize across 1600 keeps the sheet, its scroll and focus (rule 5). |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Provenance sheet (any caller) | 880pt modal sheet | 1280pt modal sheet; the Citation column widens | Nothing moves; the sheet widens (PSK-W1) |
| Record table alone (viewer) | The host Panel's width | The host Panel's width | Nothing; its columns follow PSK-W1's caps |

---

## 5. Suggested frames

The sheet over a dimmed pane at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). Draw only the sheet and enough of the pane to show it is modal.

1. **Anatomy.** The PSK-2 sheet with every slot labelled (PSK-1).
2. **Many records.** Birth date, two owners (PSK-2), Regular and Wide (PSK-W1).
3. **Value groups.** Mixed Occupation (PSK-3).
4. **One record.** "Likeness from Hamburger Adressbuch 1861", one record: "Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23", kept, OBS-7Q2KD · PER-3HW8C (PSK-4), Regular and Wide.
5. **Table alone.** The one-record table under a 300 × 380-pixel likeness inside a viewer Panel (PSK-6).
6. **Long.** 30 records scrolled (PSK-7).
7. **Keyboard.** Focus on a record's Observation link, with the VoiceOver reading as a note (PSK-8).
8. **Dark.** Frames 2 and 4 in dark (PSK-9).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Provenance sheet | Recipe | **New** (promoted) | `DesignSystem/Recipes/ProvenanceSheet/` *(new, S13-36)*, from `Features/Conclusions/ProvenanceSheet.swift` `ProvenanceSheet` | PSK-1 – PSK-4, PSK-7. Second and third call sites are known: the Gallery (S13-14) and the viewer (S13-15); a family tree later. Takes display-ready content; names no `Features/` type. |
| Record table | Recipe (part of the sheet) | **New** (promoted) | the same folder, from `Features/Conclusions/ReconciliationRecordRows.swift` `ReconciliationRecordRows` (sheet style) | PSK-5, PSK-6. Placeable without the Panel. |
| Panel, footer, Done | Component | Ship | `PVPanel`, `PVButton` (primary) | As today. |
| Sheet presenter | Component | Ship | `pvSheet(item:)` | The caller presents; Escape and a click outside close. |
| State Badge | Component | Ship | `PVBadge` | As the Vitals row's Badge. |
| Group lines | Component | Ship | micro caps text, mono caption | As S13-D12. |
| Outcome mark | Recipe | Ship | the outcome marks the Why table uses today | Unchanged. |
| Provenance control | Component | Ship | `PVIconButton(.info)` | The caller's; not part of the recipe. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A visible change to the sheet | Its look is S13-D2's; this board moves it into the kit. |
| A new Panel, table or link component | `PVPanel` and the existing table cover it. |
| Product types in the kit (a Person, a Source, a navigation place) | Kit rule: a recipe takes display-ready content; the caller decides what a link opens. |
| The Event and Place Why accordion | Working rule "The provenance sheet replaces the Why on the Person page only". |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the kit project, the specimen's file, turn and frames; the slot names (PSK-1); the one-record header (PSK-4: value line or not); the table-alone frame (PSK-6); and every other change from this brief with the user's answer.
2. Write the S13-36 plan with [`plan-pr`](../../../../.agents/skills/plan-pr/SKILL.md) from this board; then refresh S13-D5 (frame 2b and the viewer's record row instance the kit sheet) and the S13-14 and S13-15 plans to use it.
3. S13-36 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
