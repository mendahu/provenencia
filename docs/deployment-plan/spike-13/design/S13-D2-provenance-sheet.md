# S13-D2 — Provenance sheet

**Kind:** Claude Design board (**enhancement** of the Person page board S13-D1 and S13-D12: the reserved slot on each Vitals row gains the provenance control, and the Why accordion on the Person page becomes a sheet)
**Spike:** Provenencia Spike 13 (Person page), slice 2 (provenance sheet)
**View:** Person page — Vitals: the provenance control on every row, and the provenance sheet it opens
**Implements later as:** PR **S13-11** (Provenance sheet: PRV-1 – PRV-14, PRV-W1 – PRV-W2)
**Depends on:** S13-D1 (board agreed 2026-10-29; built in S13-10, #634), S13-D12 (board agreed 2026-10-30; built in S13-06, #640)
**Related:** [`../prs/S13-11-provenance-sheet.md`](../prs/S13-11-provenance-sheet.md); [`../person-detail.md`](../person-detail.md) R-8 – R-13b, R-14, R-15; working rules "The provenance sheet replaces the Why on the Person page only" and "A derived row carries its owning entity's records" in [`../deployment-plan.md`](../deployment-plan.md#working-rules); the S13-D12 brief [`archive/S13-D12-life-rows-why.md`](archive/S13-D12-life-rows-why.md) (its group line per owner)
**Board:** — (not drawn yet). Draw it in the Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, as a new turn "S13-D2 · Provenance sheet"
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

Work on the S13-D1 Vitals frames (4a – 4y) and the S13-D12 frames (5a – 5h). Put the provenance control in the 28pt slot S13-D1 reserved at the end of every Vitals row, take the Why disclosure off the Person page's rows, and draw the sheet the control opens. Do not redraw the header, the switcher, the other sections, or the Event and Place pages (they keep the Why accordion).

---

## 1. Objective

A researcher clicks one control beside any value in Vitals and sees, in one place, everything that value rests on: every record that speaks to it, what the auto-reconciler did with each and why, and links that open the record's Source, its Citation, and the exact Observation.

**What changes.** On the Person page only:

- Every Vitals row that has records shows a **provenance control** in its reserved trailing slot.
- The row's **Why** disclosure goes away; what it showed moves into the **provenance sheet**, with more detail.
- The sheet keeps S13-D12's **group line per owner** for the life rows whose records come from several birth (or death) events or several Places.

**What stays.** Each row's label, value, state Badge (Mixed, Merged, Concluded), Source count, "1 disagrees" note, the other-values disclosure ("2 other values") and the open-arrow IconButton on the life rows. The header, the switcher and the other sections. The Event and Place pages keep today's Why accordion on their rows (working rule "The provenance sheet replaces the Why on the Person page only").

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A Vitals row | Label · value · state Badge · Source count · disagree note · disclosures (other values; Why today) · open arrow (life rows only) · the reserved 28pt slot at the end (S13-D1, PPF-8). A person has 5 – 20 rows. |
| A record | One Observation: a Subject on a Citation of a Source said this about the person. It has a **Read as** value (what it said, "not …" for a negative record, "—" when it said nothing usable), a **Source** (its title; a Source with no title shows the Observation ref), a **Citation** (no title on the wire today: the Citation is reached through its Source), an **Observation ref** (`OBS-7QH3D`), and the **Subject ref** it sits on (`PER-4KX2M` for a Person Subject). |
| An outcome | What the auto-reconciler did with the record, as a mark and a phrase: **Kept**; **Folded into "James Robins"**; **Outvoted (2 of 3 Sources)**; **Weak · low-trust Source, uncertain transcription**; **Denied by Court deposition, 1862**; **Provisional**; **No evidence**; a negative record reads **Kept** when it eliminated a value, else **Disagrees · did not eliminate**. Each outcome has its own mark so it reads without colour. These phrases are what the engine gives; the sheet words what it is given and invents no new cause. |
| Ink | A record that went into the shown value is in primary ink; the rest secondary. On a life row, a record of an owner the row does not show (the second birth event) is secondary. |
| Why today | A sunken Card under the row: title "Why “14 May 1817”", a table Read as · Source · Outcome, and an empty 88pt column reserved for a later "conclude this value" action. The Source is a link that opens the Citation in the composer with the Observation in focus. S13-D12 added, on the life rows, a quiet group line before each owner's records ("Birth of Ada Lovelace · EVT-4KX2M"), the shown owner first. |
| Three links per record (R-11) | **Source** opens the Source's page. **Citation** opens the Citation in the citation composer. **Observation** opens the composer with that Observation in focus (what today's Source link does). Each link leaves the Person page; Back returns to it. |
| Reconciliation Claims come later (R-10) | A researcher will later be able to conclude a value by hand (a Reconciliation Claim). It is not built. The sheet leaves a place for it so it slots in without a redesign: today's reserved action column, and a place above the records for "concluded by the researcher". Draw no fake content for it. |
| A row with no records | The name row of a person whose name nothing records, a life row with no event or no Place: no records, so no control. The slot stays empty so values still line up. |
| A row whose records all lost | A birth date whose only record is Denied: the row is stated empty and still has records, so it has the control (S13-D12 frame 5h). |
| A multi-valued field | Some Properties keep several values (occupations). The row lists each value; today's Why is titled "Why these 3 Occupations". |
| A span row | Not on the Person page today. |
| Sizes | Typically 1 – 4 records per row; a much-cited birth can have 15; a heavily researched name 30. Two owners on a life row is rare; three is the realistic most. Source titles can be long ("Register of Baptisms, St Michael le Belfrey, York, 1813 – 1837"). |
| The page refreshes | The page can reload while the sheet is open (another window promotes a record). |
| Reuse (R-13) | The sheet is built as a shared control, not Person-specific: the Event and Place pages adopt it in their own redesign. Draw nothing Person-only into the sheet. |
| Not elsewhere (R-13b) | Family, Timeline, Gallery and Sources get no control this spike. |

### 2.1 What this board is not

- The provenance control in Family, Timeline, Gallery or Sources (R-13b).
- Reconciliation Claims, a "conclude this value" action, or any editing in the sheet (R-10).
- The Event and Place pages (they keep the Why accordion).
- New data. The sheet shows what the Person page already reads. If the board needs something more (a Citation's own title, the Source's type or date, an image crop), name it in a board note; it becomes a follow-up engine PR (working rule "Engine PRs run ahead of their slice"), not part of this board.
- A change to the row's value, Badge, Source count or the other-values disclosure.

---

## 3. Implementation gate

| Ships in **S13-11** | Stays out |
| --- | --- |
| PRV-1 – PRV-14, PRV-W1 – PRV-W2: the control on every Vitals row with records, the sheet with its summary, records, group lines, three links per record and the room for Reconciliation Claims; keyboard, VoiceOver, dark mode, Regular and Wide. The Why disclosure leaves the Person page's rows. | Reconciliation Claims (R-10); the control outside Vitals (R-13b); the Event and Place pages; any new field on the wire. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PRV-1 | **The control.** Every Vitals row that has records shows the provenance control in its reserved 28pt trailing slot: an IconButton, small, ghost, with an existing kit glyph. **The board picks the glyph** from the kit (recommended: `info`; alternatives `listTree`, `history`). Its tooltip and spoken label name the row: "Provenance of Birth date". A row with no records leaves the slot empty. |
| PRV-2 | **The Why leaves the row.** On the Person page, no row shows the Why disclosure button or the Why Card. The other-values disclosure and the open arrow stay where S13-D1 and S13-D12 drew them. |
| PRV-3 | **The sheet opens.** Clicking the control, or pressing Space or Return on it, opens the provenance sheet for that row. **The board decides the form**, and draws both before picking: **(A)** a modal sheet over the page at both widths; **(B)** a modal sheet at Regular, and at Wide a panel docked in Page columns' secondary column beside Vitals, showing the provenance of the row last opened (the row marked as selected). Recommended: (A): one presenter, and the Event and Place pages, which have no secondary column, reuse it as is. |
| PRV-4 | **Sheet header.** The sheet's title is the row's label ("Birth date"). Under it: the shown value in the row's face (mono for a date, italic for a place), the row's state Badge, and one summary line in the reconciler's words ("Merged from 3 Sources · 1 record disagrees"; "Mixed · 2 values"; "Stated empty · 1 record denied"). A multi-valued row lists its values. A close control (Button "Done", or the kit's close IconButton; the board picks one) and Escape and a click outside close it. |
| PRV-5 | **Records.** The sheet lists every record behind the value, as today's Why orders them (by the value they went into, then the rest). Each record shows: Read as; its Source title; its outcome (mark + phrase); its Observation ref and Subject ref in mono caption, muted. **The board decides the layout**: (a) today's table widened with a fourth column for the refs, or (b) one block per record (Read as and outcome on the first line, Source and refs on the second). Recommended: (a), because it reads like the Why the researcher knows. A long Source title wraps; it is never cut. |
| PRV-6 | **Values as groups.** A row whose records went into more than one value (a Mixed row) shows which records stand behind which value. The board decides how: a value line before each value's records ("14 May 1817 · 2 Sources"), the shown value first, records that went into no value last under "Not counted"; or today's flat order with the value in the outcome phrase. Recommended: the value lines. |
| PRV-7 | **Owners as groups (S13-D12).** On a life row whose records come from several events or Places, each owner's records sit under S13-D12's group line (title in micro-caps, muted, then "· ref" in mono caption, muted; a subtle top rule; wraps; text, not a link), the shown owner first. With PRV-6, the owner line is the outer group and the value lines sit inside it. A life row with one owner shows no group line. |
| PRV-8 | **Which owner is shown.** The shown owner's group line says so, visually and to VoiceOver: recommended a muted "· shown" after the ref, spoken "Birth of Ada Lovelace, EVT-4KX2M, shown value". The other groups say nothing extra. |
| PRV-9 | **Three links per record (R-11).** Each record offers **Source** (opens the Source page), **Citation** (opens the Citation in the composer) and **Observation** (opens the composer with the Observation in focus). The Source link is the record's Source title, as today. The board decides how Citation and Observation show: recommended, the Observation ref as a link and a small "Citation" link beside it; or one ContextMenu / IconButton per record with the three. Every link closes the sheet, then opens its place; Back returns to the Person page on Vitals with the sheet closed. |
| PRV-10 | **Room for Reconciliation Claims (R-10).** Keep today's 88pt reserved action column at the end of each record. Leave a place between the header and the records where a "Concluded by you" block will later sit; draw it empty (no placeholder text, no skeleton), and add a board note marking where it goes. |
| PRV-11 | **Long and refreshing.** A row with 30 records: the sheet scrolls inside itself; the header stays. If the page reloads while the sheet is open, the sheet shows the row's new records; if the row is gone (its last record removed), the sheet closes. |
| PRV-12 | **Keyboard.** The control takes focus with Tab and shows the kit focus ring. In the sheet, Tab moves through the links in reading order; Escape closes and returns focus to the control that opened it. |
| PRV-13 | **VoiceOver.** The control reads "Provenance of Birth date, button". The sheet's title is a heading. Each group line (owner or value) is a heading. Each record reads as one element: "14 May 1817, Parish register, kept", followed by its links as separate controls. |
| PRV-14 | **Dark mode.** The row with the control, and the sheet with two owner groups, in dark mode. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Person page, Vitals | Rows in the centered column; the control at each row's end | Rows in the centered primary column (PPF-W2); the control at each row's end | Nothing in the rows |
| Provenance sheet | A modal sheet over the page | (A) the same sheet, wider so its table has room; or (B) docked in the secondary column beside Vitals | Per PRV-3 |

| ID | Requirement |
| --- | --- |
| PRV-W1 | At Wide, Vitals' rows keep the primary column's Regular measure (rule 2). Under PRV-3 (A) the sheet is wider at Wide than at Regular (the board names both widths) and its table columns take the extra width; the text measure of Read as and the outcome phrase does not stretch past what reads well. Under (B) the docked panel fills the secondary column (340 – 480pt) and the Regular sheet's content stacks in it. |
| PRV-W2 | A resize across 1600 keeps the sheet open on the same row, its scroll position and focus (rule 5). Under (B), crossing back to Regular turns the docked panel into the sheet without losing the row. |

---

## 5. Suggested frames

All on Vitals, from S13-D1's full person (20 rows, one Mixed row) and S13-D12's two-birth-events person. **Regular** pane 1102 × 900 unless the frame says Wide (pane 2296 × 1440).

1. **Rows with the control.** Vitals with the control on every row that has records and an empty slot on a row that has none; no Why buttons (PRV-1, PRV-2).
2. **Form options side by side.** The sheet as (A) and as (B) at Wide, for the birth date. The board keeps the option it picks and deletes the other (PRV-3).
3. **Sheet, merged name.** The name row's sheet: three records, two Kept, one Folded; the summary line; the refs; the links; the reserved column (PRV-4, PRV-5, PRV-9, PRV-10).
4. **Sheet, Mixed date.** A Mixed birth date with two values: the value lines (or the board's PRV-6 choice), one Outvoted record, one Weak record with its causes (PRV-6).
5. **Sheet, two birth events.** S13-D12 frame 5d's person: two owner groups, the shown one first and marked, value lines inside (PRV-7, PRV-8).
6. **Sheet, two Places.** Death place Montréal and Paris, two owner groups (PRV-7, PRV-8).
7. **Sheet, all lost.** A birth date stated empty whose only record is Denied by a negative record: the summary says so (PRV-4).
8. **Sheet, long.** 30 records on a name, scrolled to the middle, the header still showing (PRV-11).
9. **Focus.** The control focused with the ring; the sheet with focus on a record's Observation link (PRV-12).
10. **Wide.** Frame 1 and frame 5 at Wide, in the form PRV-3 picks (PRV-W1, PRV-W2).
11. **Dark.** Frames 1 and 5 in dark mode (PRV-14).

---

## 6. UI building-block inventory

This table is **binding**. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Vitals row | Snowflake | **Extend** | `Features/Conclusions/ReconciledValueRow.swift` `ReconciledValueRow` (`Provenance.sheet`) | Its reserved slot gains the control; in `.sheet` mode it shows no Why. Event and Place pass `.why` and look as today. |
| Provenance control | Component | Ship | `DesignSystem/Components/IconButton/PVIconButton.swift` (`size: .sm`) with an existing `PVSymbol` | PRV-1. One IconButton, not a new control. |
| Sheet presenter | Component | Ship | `DesignSystem/Components/Panel/PVSheet.swift` `.pvSheet` | PRV-3 (A), PRV-11, PRV-12. Escape and outside click follow its one rule; the sheet is never dirty. |
| Sheet chrome | Component | Ship | `DesignSystem/Components/Panel/PVPanel.swift` `PVPanel` | PRV-4. Title, subtitle, body; footer only if the board's close is a "Done" Button. |
| Docked panel (only if PRV-3 picks B) | Component | Ship | `DesignSystem/Components/PageColumns/PVPageColumns.swift` secondary column | PRV-3 (B), PRV-W1. The secondary slot Page columns already has; no new mode. |
| Provenance sheet body | Snowflake | **New** | `Features/Conclusions/ProvenanceSheet.swift` *(new)* | PRV-4 – PRV-11. Takes a row model, nothing Person-only (R-13). Second caller known: the Event and Place pages' redesign. |
| Record table | Snowflake | **Rethink** (on the Person page) | `Features/Conclusions/ReconciliationReasoningView.swift` | Its group line per owner moves into the sheet body; the accordion keeps today's flat table for Event and Place. Outcome cells keep `ReconciliationOutcome` (`Features/Conclusions/ReconciliationOutcome.swift`). |
| State Badge | Component | Ship | `DesignSystem/Components/Badge/PVBadge.swift` | PRV-4. The row's own Badge, repeated in the sheet header. |
| Links | Component | Ship | `DesignSystem/Components/Button/PVButton.swift` (`variant: .link`, `size: .sm`) | PRV-9. Source title, Observation ref, "Citation". If the board picks one menu per record, `PVIconButton` `.ellipsis` with `ContextMenu`. |
| Group lines (owner, value) | Snowflake | Ship (moved) | today's owner line in `ReconciliationReasoningView` | PRV-6 – PRV-8. Same micro-caps heading style for owner and value lines. No new kit piece. |
| Disclosure (other values) | Component | Ship | `DesignSystem/Components/Disclosure/PVDisclosureButton.swift` | PRV-2. Stays on the row. |
| Divider rules | Component | Ship | `DesignSystem/Components/Divider/PVDivider.swift` | Between records, as today. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A new Inspector or Drawer kit component | Page columns' secondary column and `.pvSheet` already cover both forms. |
| A Person-only sheet | R-13: the Event and Place pages adopt the same sheet. |
| Placeholder content for Reconciliation Claims | It reads as data; leave the place empty with a board note (PRV-10). |
| New outcome words or causes | The sheet words what the engine gives (cross-spike: S11-28, S11-44 and S11-45 add causes). |
| The control in Family, Timeline, Gallery or Sources | R-13b. |
| A Why disclosure beside the control on the Person page | PRV-2: one way to see provenance. |

---

## 7. Handoff

When the board is agreed: record the board link, the frames and the answers to PRV-1 (glyph), PRV-3 (form), PRV-4 (close), PRV-5 (layout), PRV-6 (value grouping) and PRV-9 (links) on the **Board** line of this brief and in the [S13-11 plan](../prs/S13-11-provenance-sheet.md) (**Board** line, Design → macOS, its Decisions), close the plan's Open questions, run `review-plan` on the plan, and move it from `Blocked — on board S13-D2` to `Planned` with `plan-pr`. When S13-11 lands, move this brief to `design/archive/` and update [`README.md`](README.md).
