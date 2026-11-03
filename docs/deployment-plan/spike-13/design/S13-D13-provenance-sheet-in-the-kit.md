# S13-D13 — The provenance sheet in the design system

**Kind:** Claude Design board (**enhancement** of the Provenencia design-system project: one new recipe specimen, the provenance sheet, moved from the Person page board into the kit)
**Spike:** Provenencia Spike 13 (Person detail page), Slice 5 — Gallery (preparatory)
**View:** The design system itself: a Provenance sheet specimen page in the kit, with its anatomy, slots, variants and states. No product screen changes.
**Implements later as:** PR **S13-36** (the provenance sheet joins the design system), before S13-14.
**Board:** not drawn yet. When agreed, record the Claude Design project, the specimen's file, turn and frames here, with every answer to § 7.
**Depends on:** the agreed S13-D2 board (the sheet as it ships; [archived brief](archive/S13-D2-provenance-sheet.md), built by S13-11 [#650](https://github.com/mendahu/provenencia/pull/650)); the agreed S13-D12 board (owner group lines; [archived brief](archive/S13-D12-life-rows-why.md)).
**Related:** [`../provenance-sheet.md`](../provenance-sheet.md) (where the boundary sits: the kit assembles; callers choose records, title and destinations); [`../person-detail.md`](../person-detail.md) R-8 – R-15, R-13b, R-32; [`../deployment-plan.md`](../deployment-plan.md) working rules "The provenance sheet is a kit recipe", "The person card lives in the design system; display is decoupled from content"; brief [S13-D5](S13-D5-gallery-and-viewer.md) (the Gallery's one-record sheet and the viewer's record row, which instance this recipe); [`../../../ideas/kinship-view.md`](../../../ideas/kinship-view.md) "Family trees and person pages" (a future caller).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** [`S13-D13-provenance-sheet-in-the-kit.review.md`](S13-D13-provenance-sheet-in-the-kit.review.md) — round 1 Fits with changes / Ready with changes, every finding Fixed or answered by the user

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

**Why a recipe.** Vitals, the Gallery and the viewer all live in the Person page's feature, so a shared feature view would serve them. A recipe is right because a family tree in another feature is a known caller and the user asked for the sheet in the kit (2026-11-03); a recipe takes display-ready content, so that caller fills it without reaching into the Person page's code.

**What stays.** Everything the sheet looks like and does today (S13-D2, S13-D12): the Panel, the header, the record table and its columns, the owner and value group lines, the three links per record, the reserved "Concluded by you" place, Done, the sizes at Regular and Wide, keyboard and VoiceOver. The Vitals provenance control (the info IconButton). The Event and Place pages' Why accordion keeps its own look **and its own copy of the record table** in the Person page's code until those pages are redesigned onto this sheet (user, 2026-11-03: they will be replaced soon); that copy is recorded debt (working rule "The provenance sheet is a kit recipe").

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A value and its records | A value (a birth date, an occupation, a likeness) is concluded from **records**: Observations on Citations. Each record has what it was **read as** ("14 May 1817"), its **citation line** ("Parish register, Vol. 3, Baptism of Ada, p. 23"), its **outcome** (kept, folded, outvoted, weak, denied, against, no usable value) with a mark, its **Observation ref** and **Subject ref**, and three destinations: the Source, the Citation, the Observation. |
| Groups | Records sit under **owner** group lines when they come from several entities (a birth date read off two birth events, S13-D12), and under **value** group lines when they went into different values (a mixed row). Each group line is a heading. |
| The header | A title (the value's label: "Birth date", "Likeness from Hamburger Adressbuch 1861"), the shown value or values in their face (mono for a date, italic for a place) or the empty text in italic, the state Badge (merged, mixed, multiple) when there is one, and a summary line ("merged from 3 Sources · 4 records"). |
| The kit names no product types | A kit recipe takes **display-ready content**: for the sheet, a title; zero or more values, each with its face (text, date in mono, place in italic) or an empty text; an optional state Badge (Merged, Mixed or Concluded, the three a Vitals row can show); an optional summary; and groups. A group has an optional owner line (title, ref, "shown"), value groups with an optional heading, and records. A record has its Read as (or a phrase such as "not …"), whether it counted, its citation line, its outcome (mark and phrase), its Observation and Subject refs, and a stable id. The recipe has three levels a caller mixes freely: a single **record row**, a **record table** of rows with group lines, and the **sheet** (PSK-6). Every level reports a link as "this record, and Source, Citation or Observation"; the caller maps that to a destination, closes the sheet, then navigates. The sheet never knows what a Person, a Source or a navigation place is. **The kit also assembles that content** from the reconciler's records every caller already holds (a field, a span of fields, a life row's owners, one record), so callers do not format records; a caller with records of another kind fills the content itself ([`../provenance-sheet.md`](../provenance-sheet.md)). On the board this is invisible: draw the content, not the assembly. |
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
| Ships in **S13-36** | The sheet as a kit recipe and its record table as a kit part, filled by display-ready content that the recipe's assembler builds from the reconciler's records ([`../provenance-sheet.md`](../provenance-sheet.md); the S13-36 planner may cut it in two), with the optional slots and columns of PSK-1, PSK-4 and PSK-5; the Vitals rows moved onto it with no visible change (PSK-2, PSK-3, PSK-5, PSK-7 – PSK-9, PSK-W1). PSK-4 and PSK-6 are previews only in S13-36: they are content shapes, not code paths, and gain their callers in S13-14 and S13-15. |
| Used by | S13-14 (the Gallery's one-record sheet, PSK-4), S13-15 (one record row under the viewer's image, PSK-6), later a family tree. |
| Stays out | Any visible change to Vitals. The Event and Place Why accordion and its own copy of the record table (`Features/Conclusions/ReconciliationRecordRows.swift` in its Why mode, `ReconciliationReasoningView`): it stays in the Person page's code, unchanged, until those pages move onto this sheet (user, 2026-11-03). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PSK-1 | **Specimen page.** A kit page "Provenance sheet" (a recipe; file *Provenance Sheet.dc.html* in the kit project, listed with the Person card under Recipes) showing the sheet's anatomy with every slot named: **title**, **value** (optional: one value, several, or the empty text), **state Badge** (optional), **summary** (optional), **reserved block** (the "Concluded by you" place, zero height), **record table** (PSK-5), **footer** (Done). Label each slot on the specimen. |
| PSK-2 | **Many records (the Vitals case).** The Birth date sheet from S13-D2: two owner groups (the shown owner first, marked "· shown"), five records, one outvoted. Unchanged from the S13-D2 frames. |
| PSK-3 | **Value groups.** A mixed Occupation sheet with two value groups and a group of records that went into no value. Unchanged from S13-D2 (PRV-6), headings only when a group has two or more buckets. |
| PSK-4 | **One record (the Gallery case; user, 2026-11-03).** The sheet for one record: the title names the thing ("Likeness from Hamburger Adressbuch 1861"); **no value line** (a likeness has no value to read out); no state Badge; the summary "1 record"; the table with one record, no group lines, and **no Read as column** (a likeness reads only as "Likeness", which the title already says). Draw it at Regular and Wide. This is the frame S13-D5's Gallery tile opens (its frame 2b instances it). |
| PSK-5 | **The record table.** Read as · Citation (the citation line, a link in ink that opens the Citation) · Outcome (mark and phrase) · Record (Observation and Subject refs, mono caption, muted) · the reserved 88pt column; the Source and Observation links under the citation line; primary ink for a counted record, secondary for the rest. As S13-D2 PRV-5, PRV-9, PRV-15 and the board's notes. The Read as column is optional (PSK-4). The citation line's copy button (the kit clipboard button) stays. Name it as the sheet's part: a caller can place it without the Panel (PSK-6). |
| PSK-6 | **Three levels (user, 2026-11-03).** Draw the recipe's three levels side by side on the specimen, each a named part a caller can place alone: **Record row** (one record's cells: Read as, the citation line with its copy button, outcome, refs, links; no Panel, no heading row, no group line), **Record table** (rows with group lines and an optional heading row), **Sheet** (a Panel, the header and a table). Then a usage frame: **one Record row on its own** under a likeness image inside the viewer's Panel, with no Read as column. The viewer's Panel takes the sheet's widths (880pt at Regular, 1280pt at Wide), so the row keeps its column caps (PSK-W1; S13-D5 LV-1). Show that a row and a table built with the same columns line up. |
| PSK-7 | **Long.** 30 records: the table scrolls inside the sheet past its cap; title, header and Done stay. |
| PSK-8 | **Keyboard and VoiceOver.** As S13-D2 PRV-12, PRV-13: Tab through the citation line, Source and Observation links in reading order; the title and each group line are headings; each record reads as one element. A note on the specimen. |
| PSK-9 | **Dark mode.** PSK-2 and PSK-4 in dark. |
| PSK-W1 | **Wide.** The sheet's two widths (880pt Regular, 1280pt Wide, S13-D2), Read as capped at 220pt and Outcome at 280pt at Wide, the Citation column taking the rest. A resize across 1600 keeps the sheet, its scroll and focus (rule 5). |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Provenance sheet (any caller) | 880pt modal sheet | 1280pt modal sheet; the Citation column widens | Nothing moves; the sheet widens (PSK-W1) |
| Record row alone (viewer) | The host Panel's width | The host Panel's width | Nothing; its columns follow PSK-W1's caps |

---

## 5. Suggested frames

The sheet over a dimmed pane at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). Draw only the sheet and enough of the pane to show it is modal.

1. **Anatomy.** The PSK-2 sheet with every slot labelled (PSK-1).
2. **Many records.** Birth date, two owners (PSK-2), Regular and Wide (PSK-W1).
3. **Value groups.** Mixed Occupation (PSK-3).
4. **One record.** "Likeness from Hamburger Adressbuch 1861", one record: "Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23", kept, OBS-7Q2KD · PER-3HW8C (PSK-4), Regular and Wide.
5. **Three levels.** The Record row, the Record table and the Sheet side by side, each labelled; then one Record row on its own under a 300 × 380-pixel likeness inside a viewer Panel (PSK-6).
6. **Long.** 30 records scrolled (PSK-7).
7. **Keyboard.** Focus on a record's Observation link, with the VoiceOver reading as a note (PSK-8).
8. **Dark.** Frames 2 and 4 in dark (PSK-9).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Provenance sheet | Recipe | **New** (promoted) | `DesignSystem/Recipes/ProvenanceSheet/` *(new, S13-36)* `PVProvenanceSheet`, from `Features/Conclusions/ProvenanceSheet.swift` `ProvenanceSheet` | PSK-1 – PSK-4, PSK-7. Later call sites are known: the Gallery (S13-14), the viewer (S13-15), a family tree. Takes `PVProvenanceSheetContent` (§ 2); names no `Features/` type; ids from a prefix the caller passes. |
| Assembler | Recipe (logic, no drawing) | **New** (moved) | the same folder: `PVProvenanceSheetContent.init(field:title:…)`, `.spanning`, `.derived`, `.narrowed(to:title:)`, from `ReconciledValueRowModel`'s provenance logic and `ReconciliationRecord.init` | Not drawn on the board. Builds the content every variant shows; [`../provenance-sheet.md`](../provenance-sheet.md) § 5. Record and state wording and the owner titles move to `Platform/` with it. |
| Record row | Recipe (part of the sheet) | **New** (promoted) | the same folder, `PVProvenanceRecordRow`, from one record of `ReconciliationRecordRows` (sheet mode) | PSK-5, PSK-6. Placeable alone, anywhere; the table is made of it. |
| Record table | Recipe (part of the sheet) | **New** (promoted) | the same folder, `PVProvenanceRecordTable`, from `ReconciliationRecordRows` (sheet mode only) | PSK-5, PSK-6. Rows plus group lines; placeable without the Panel. The Why mode stays in Features (§ 3). |
| Columns | Recipe (part of the sheet) | **New** | the same folder, `PVProvenanceColumns` | PSK-6. Which optional columns show and the width class; shared by row and table so they line up. |
| Group lines (owner, value heading) | Recipe (part of the table) | **New** (moved) | from `ReconciliationRecordRows`'s private `ownerLine`, `valueHeading` | PSK-2, PSK-3. |
| Citation link | Recipe (part of the table) | **New** (moved) | from the private `CitationLineButton`, `CitationLineButtonStyle` | PSK-5. A link in ink, as S13-D2's board note; it moves, it is not a new link component. |
| Outcome mark and phrase | Recipe (part of the table) | **New** (moved) | from `Features/Conclusions/ReconciliationOutcome.swift` `ReconciliationOutcome` | PSK-5. The Why accordion's copy keeps using the Features one until it is deleted. |
| Value faces | Recipe (part of the sheet) | **New** (moved) | from `ReconciledValueRowModel.ValueStyle` and `ReconciledValueRow.valueContent` | Text, date (mono), place (italic); the Vitals row keeps reading its own. |
| Copy the citation | Component | Ship | `PVClipboardButton` | PSK-5. As today. |
| Panel, footer, Done | Component | Ship | `PVPanel`, `PVButton` (primary) | As today. |
| Sheet presenter | Component | Ship | `pvSheet(item:)` | The caller presents; Escape and a click outside close. `ProvenanceSheetItem` and its `row(in:)` stay with the caller. |
| State Badge | Component | Ship | `PVBadge` | Merged, Mixed, Concluded, as the Vitals row's Badge. |
| Provenance control | Component | Ship | `PVIconButton(.info)` | The caller's; not part of the recipe. |
| Copy | — | Ship | an `L10n.ProvenanceSheet` namespace *(new, S13-36)* for the sheet's own words (Done, column headings, link labels); the caller's words (title, values, summary) arrive in the content | As S13-35 moved the card's copy to `L10n.PersonCard`. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A visible change to the sheet | Its look is S13-D2's; this board moves it into the kit. |
| A new Panel, table or link component | `PVPanel` and the moved table, group lines and citation link cover it. |
| Product types in the kit (a Person, a Source, a navigation place) | Kit rule: a recipe takes display-ready content; the caller decides what a link opens. |
| Moving the Event and Place Why accordion onto the kit table | User, 2026-11-03: those pages keep their own copy until they are redesigned onto this sheet. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the kit project, the specimen's file, turn and frames; the slot names (PSK-1); the one-record header (PSK-4: value line or not); the three levels and the lone-row frame (PSK-6); and every other change from this brief with the user's answer.
2. Write the S13-36 plan with [`plan-pr`](../../../../.agents/skills/plan-pr/SKILL.md) from this board; then refresh S13-D5 (frame 2b and the viewer's record row instance the kit sheet) and the S13-14 and S13-15 plans to use it.
3. S13-36 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
