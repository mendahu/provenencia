# S13-D12 — Person page: the Why on the four life rows

**Kind:** Claude Design board (**enhancement** of the Person page board S13-D1: the four life rows gain today's Why disclosure)
**Spike:** Provenencia Spike 13 (Person page), slice 2 (provenance sheet)
**View:** Person page — Vitals, the birth date, birth place, death date and death place rows
**Implements later as:** PR **S13-06** (Derived rows carry their records: LW-1 – LW-7, LW-W1)
**Depends on:** S13-D1 (board agreed 2026-10-29; built in S13-10, #634)
**Related:** [`../prs/S13-06-derived-rows-carry-their-records.md`](../prs/S13-06-derived-rows-carry-their-records.md); [`../person-detail.md`](../person-detail.md) R-12, R-13a, R-14; working rule "A derived row carries its owning entity's records" in [`../deployment-plan.md`](../deployment-plan.md#engine); the provenance sheet that later replaces this Why on the Person page is S13-D2 (S13-11)
**Board:** — (not drawn yet). Draw on the Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), beside the S13-D1 frames
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
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Work on the S13-D1 Vitals frames (4a – 4y). Add the Why disclosure to the four life rows and draw it open. Do not redraw the header, the switcher, the other Vitals rows, the reserved provenance slot or the other sections.

---

## 1. Objective

A researcher reads where a birth date or a birth place comes from without leaving the Person page.

**What changes.** The four life rows (birth date, birth place, death date, death place) get the same **Why** disclosure every other Vitals row has. Opened, it lists the records behind the row's value: for a date, every record of that date on the person's birth (or death) events; for a place, every record naming the Place the event happened at.

**What stays.** The rows' label, value, Mixed badge and the open-arrow IconButton that goes to the Event or Place page; every other Vitals row and its Why; the reserved 28pt slot at each row's end for S13-D2's provenance control; the Event and Place pages.

This Why is interim: S13-D2 (S13-11) replaces the Why accordion on every Vitals row with the provenance sheet. This board draws only what S13-06 ships.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A life value belongs to another record | The birth date is a property of the birth event, and the birth place is the Place that event happened at (R-13a). The row reads them off those records; the arrow opens that Event or Place page. |
| Why today | A Vitals row's Why is a sunken Card titled "Why “14 May 1817”" (or "Why “Birth date”" when the row is empty) holding a table: **Read as** · **Source** · **Outcome**, plus an empty 88pt column reserved for a later action. Each record shows what the record said, the Source as a link that opens the Citation with its Observation in focus, and the outcome (Kept, Folded, Outvoted, Weak, Denied, …) with its mark. Records that went into the shown value are in primary ink; the rest secondary. |
| The date row's records | The records of the date the row reads (the event's date, else its start date) on **every** surviving birth (or death) event of the person, not only the one the row shows. |
| The place row's records | The toponym records (the names) of every Place the row's event happened at. A Place's parent chain (Yorkshire, England) and the record linking the event to the Place are not listed. |
| Two birth events | A person can have two birth events that do not merge (two records, two dates). The row shows the earlier one and the Mixed badge. The date row's Why lists the records of both events; each event's outcomes are relative to its own date, so a record can read Kept on each. |
| Two Places | An event can be at two Places that do not fold into one chain (Montréal and Paris). The place row shows the first and the Mixed badge; its Why lists the names of both. |
| Empty | A life row with no event or no Place has no records and no Why, as today. A row whose records all lost (every date Denied, say) is stated empty and still has a Why. |
| Sizes | Typically 1 – 4 records per life row; a much-cited birth can have 15. Two events or two Places is rare; three is the realistic most. |

### 2.1 What this board is not

- The provenance control or sheet (S13-D2).
- The Event and Place pages' rows (they keep their own Why on their date and name rows).
- The location link's evidence or the parent chain's names.
- Any change to the other Vitals rows.

---

## 3. Implementation gate

| Ships in **S13-06** | Stays out |
| --- | --- |
| LW-1 – LW-7, LW-W1: the Why disclosure on the four life rows, its title, the records of every owning event or Place, how records from several events or Places are told apart, VoiceOver. | The provenance sheet and control (S13-11). The Event page's place row (it keeps no Why). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| LW-1 | Each life row that has records shows the **Why** disclosure button where other rows show it, before the open-arrow IconButton. A row with no records shows no Why button and keeps its layout. |
| LW-2 | Opened, the Why is today's sunken Card and table under the row at the disclosure indent: Read as · Source · Outcome and the reserved column. The title is today's "Why “<value>”" ("Why “14 May 1817”", "Why “York, Yorkshire”"), or "Why “<label>”" ("Why “Birth date”") when the row is stated empty. |
| LW-3 | **One owner.** When the records come from one event (date row) or one Place (place row), the table lists them as any other row's Why does: by the value they went into, then the rest. Nothing names the event or Place. |
| LW-4 | **Several owners.** When the records come from two or more events or Places, the board shows which record belongs to which. **The board decides how.** Draw both and pick one: **(a)** one table, with a quiet group line before each owner's records naming it ("Birth · EVT-4KX2M", "Montréal · PLC-7QH3D"), the shown owner's group first; **(b)** one table with no group lines, each record's Outcome followed by the owner's name in secondary ink. Recommended: (a). |
| LW-5 | Records keep today's ink rule within each owner: a record that went into that owner's shown value is primary, the rest secondary. A record from an owner the row does not show (the second birth event) is secondary in its Read as column. |
| LW-6 | The Source link opens the record's Citation with its Observation in focus, as today. If LW-4 picks (a), the group line is text, not a link (the open arrow already goes to the shown owner). |
| LW-7 | VoiceOver: the Why button reads as it does on other rows. Inside, each record reads as today; under LW-4 (a) each group line is a heading read as "Birth, EVT-4KX2M". |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Person page, Vitals | Rows in the column; a Why opens under its row | Rows in the centered primary column (S13-D1, PPF-W2); a Why opens under its row | Nothing: the Why is part of its row and S13-D2 owns the secondary column (Wide rule 2) |

| ID | Requirement |
| --- | --- |
| LW-W1 | At Wide a life row's Why opens under its row inside the primary column at the Regular measure, as other rows' Why does. A resize across 1600 keeps an open Why open and the scroll position. |

---

## 5. Suggested frames

Frames 1 and 3 at **Regular** (pane 1102 × 900) and again at **Wide** (pane 2296 × 1440); the others at Regular. All on Vitals, from S13-D1's full person.

1. **Birth date Why open, one event.** Three records: two Kept that merge ("14 May 1817"), one Outvoted ("1816") (LW-1 – LW-3, LW-5).
2. **Birth place Why open, one Place.** Two records naming York and one naming "Eboracum" Folded (LW-3).
3. **Two birth events, both LW-4 options side by side.** The row shows "14 May 1817" Mixed; the Why lists the lead event's two records and the second event's one record ("2 Jun 1818"). The board keeps the option it picks and deletes the other (LW-4, LW-5, LW-7).
4. **Two Places.** Death place "Montréal" Mixed; the Why lists Montréal's names and Paris's name under the option LW-4 picks.
5. **Empty and all-lost.** Death date stated empty with no Why button; a birth date stated empty whose only record is Denied, with its Why open (LW-1, LW-2).

---

## 6. UI building-block inventory

This table is **binding**. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Vitals row | Snowflake | Ship | `Features/Conclusions/ReconciledValueRow.swift` `ReconciledValueRow` | Unchanged; it already shows the Why button when its model has records. |
| Life row model | Snowflake | **Extend** | `Features/Conclusions/ReconciledValueRowModel.swift` `ReconciledValueRowModel.derived` | Gains records and a Why title. |
| Why table | Snowflake | **Extend** (only if LW-4 picks (a)) | `Features/Conclusions/ReconciliationReasoningView.swift` | Under (a), a group line row before each owner's records, in today's heading style (`pvMicroCaps`, muted). Under (b), a secondary-ink suffix in the Outcome cell. No new table. |
| Why disclosure button | Component | Ship | `DesignSystem/Components/Disclosure/PVDisclosureButton.swift` | As on every row. |
| Open arrow | Component | Ship | `DesignSystem/Components/IconButton/PVIconButton.swift` | Unchanged, after the Why button. |
| Mixed badge | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` | Unchanged. |
| Sunken card | Component | Ship | `DesignSystem/Components/Card/PVCard.swift` (`tone: .sunken`) | As today. |

**Non-goals.** No new kit primitive. No second Why view for derived rows. No change to the reserved provenance slot.

---

## 7. Handoff

When the board is agreed: record the board link, the frames and the LW-4 option on this brief and in the S13-06 plan (**Board** line, Design → macOS, D-row that reads LW-4), move S13-06 from `Blocked — on board S13-D12` to `Planned` with `plan-pr`, and when S13-06 lands move this brief to the spike design archive folder and update [`README.md`](README.md).
