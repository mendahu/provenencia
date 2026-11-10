# S16-D1 — Person page notices: a failed reload says so, and a partial diagram says so

**Kind:** Claude Design board (**enhancement** of the Spike 13 Person page boards, S13-D1 page frame and S13-D10 record links, and of the S16-D2 turn on the same board: states are added to the existing frames)
**Spike:** Provenencia Spike 16 (odds and ends), batch 1
**View:** the Person page: its header band (with the section switcher), a full-width section's band (Family, Timeline, Sources), and the Sources section's record links column. The Event and Place pages share the page frame and take the header-band notice the same way (one frame)
**Implements later as:** PR **S16-04** (The Person page says when a reload failed or the diagram is partial)
**Board:** — (not drawn yet). Draw on [Claude Design — Person Detail](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ) as a new turn "S16-D1 · Notices"
**Depends on:** the agreed S13-D1, S13-D10 and S16-D2 boards (the page frame shipped in S9-24 and S13-08; Sources and the record links in S13-16, [#722](https://github.com/mendahu/provenencia/pull/722), and S13-28, [#744](https://github.com/mendahu/provenencia/pull/744); the folds in S16-07, [#761](https://github.com/mendahu/provenencia/pull/761))
**Related:** audit P-5, P-10, P-17 ([`../person-page-audit.md`](../person-page-audit.md)); working rules "A refresh failure is read beside its value", "The record links read is bounded and says so", "A person's records are records" in [`../deployment-plan.md`](../deployment-plan.md#working-rules); the Promote page's and the Evidence graph's retry notices (S14-D7, S14-20) for the kit piece; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide)
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

Draw on the Person page board as a new turn "S16-D1 · Notices", starting from the S13-D1 page frames (header band and section switcher), the S16-D2 Family, Timeline and Sources frames, and the S13-D10 frames with the record links diagram beside the rows (Regular and Wide). Add notices and one note. Do not redraw the header, the switcher, the sections' rows, tiles or cards, the folds, the diagram's circle, its nodes or its title.

---

## 1. Objective

A Person page reads its data in parts: the page's own read (the header, Vitals and the Gallery), one read per section (Family, Timeline, Sources), and a second read beside Sources for the record links diagram. After a write, each part reloads behind what it already shows; the header's "Refreshing" spinner says so.

**What is wrong today.**
- When a reload fails, the old rows stay and nothing says so (P-5). After a Remove whose reload fails, the removed record line stays, and a second Remove does nothing.
- When a first load fails, the section shows a red Callout that says "Please try again" with nothing to press, and VoiceOver hears nothing.
- A person deleted or merged from another window leaves with "This record no longer exists", where the page calls a person's claims "records" (P-17).
- When the record links read was too large to compare, the diagram draws every link thin (as if the records barely agree) or no link at all (as if they stood apart). When more links or records were found than the diagram draws, the rest are dropped without a word (P-10).

**What changes.**
- A failed reload shows a danger notice with **Try again** above the content it leaves in place. A failed first load shows the same notice in place of the content. VoiceOver hears it when it appears and each time a Try again fails again.
- One notice per read, where that read draws: the page's read in the header band; a section's read at the top of its band; the record links read at the top of the diagram column, compact.
- A Person page whose person is gone says "This person no longer exists", in the Toast as it leaves and in any section notice that reports it. Event and Place pages keep today's line.
- The diagram column shows one note when its drawing is partial. A read that did not compare draws the records and no links.

**What stays.** The page frame, the header, the "Refreshing" spinner, the sections' rows and folds, the empty states, the diagram's circle, nodes, title and caption (when it counted), and the Wide arrangements S13-D1, S13-D10 and S16-D2 agreed.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The page's read | The header (name, life lines, ref, record count, portrait), Vitals and the Gallery. On Event and Place pages, the whole page. |
| A section's read | Family, Timeline and Sources each read on their own, only while open. Vitals and Gallery have no read of their own. |
| The record links read | The diagram beside the Sources rows has its own read. The rows and the diagram fail apart. |
| A reload | Runs behind the shown value after a write in this or another window. It can fail (a locked or moved project, a disk error); the next write or Try again reads again. |
| Gone | The person was deleted or merged into another person, here or in another window. Not a failure that passes: the page leaves for the Persons list with a Toast. A section can learn it first, while the page still shows. |
| A partial read | The record links read compares at most a fixed amount of work, and sends at most 75,000 links; the diagram draws at most 1,000 records. Past the first, nothing was compared; past the others, the drawing leaves some out. Real projects stay far under these; a crafted or badly broken project reaches them. |
| A link | One arc between two records' graphs per other person they share; thick when two or more facts agree, thin otherwise. |

### 2.1 What this board is not

- The Persons, Events and Places lists: their failed refresh gets its own brief later.
- A notice for the Source types the Sources rows label themselves with: a failed read of them leaves the type labels out, as today.
- A notice per line, or any data about a link (Spike 13: the diagram shows no data about a line).
- A dismiss button: a notice goes when its read succeeds.
- Moving the keyboard to a notice that appears on its own (a background reload). Only a Try again the researcher pressed keeps the keyboard on the notice.
- Back restoring a section's place (P-19).
- What Remove does on a record line under a stale notice: it stays as today (open question for the user, recorded in the plan's review log).

---

## 3. Implementation gate

| S16-04 ships | Not S16-04 |
| --- | --- |
| PN-1 – PN-10, PN-W1, PN-W2 | The lists' notices ([idea](../../../ideas/person-page-followups.md)); stale-while-refreshing of hidden pages (S16-09); the section-switch and Remove announcements (S16-06) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PN-1 | **A failed reload keeps the content and says so.** The notice sits above the content it covers, which stays as it was. Kit: Callout, danger tone, with one small secondary **Try again** Button under its message (the Promote page's and the Evidence graph's retry notice). Title: what could not be read (below). Message: the error's own text, then "What is shown is from the last time it was read." |
| PN-2 | **A failed first load offers Try again.** Same notice, in place of the content (no rows to keep). Message: the error's own text only. |
| PN-3 | **The page's notice** sits in the header band, under the section switcher, at the header's width, on every section (it covers the header, Vitals and the Gallery). On a first-load failure the page has no header: the notice alone at the top of the page column. Title: "This page could not be read". |
| PN-4 | **A section's notice** sits at the top of that section, at the page band's width (not stretched at Wide), above the rows, tiles or lifeline. Titles are today's: "Family could not be read", "Timeline could not be read", "Sources could not be read". |
| PN-5 | **The diagram's notice** sits at the top of the record links column, above the diagram's title, in the Callout's compact form. On a first-load failure it is all the column shows. Title: today's "The connections could not be read". The Sources rows and their notice are independent: either, both, or neither can show. |
| PN-6 | **Try again** reads again only what its notice reports. While it runs, the Button shows its loading state; over shown content the header's "Refreshing" spinner runs too, as for any reload. On success the notice goes. On a new failure the notice stays, the keyboard stays on Try again, and VoiceOver hears the notice again. |
| PN-7 | **VoiceOver hears each notice** when it appears and when its read fails again (danger priority). A notice that appears on its own does not take the keyboard. |
| PN-8 | **Gone.** When the person is gone, the page leaves for the Persons list with a Toast: "This person no longer exists. They may have been merged into another person." When a section's read learns it first, that section's notice says the same line under its title, with no Try again (it would fail the same way). An Event or Place page keeps "This record no longer exists. It may have been merged into another." |
| PN-9 | **A partial diagram says so, once.** One note in the record links column, under the title, in place of the caption ("An arc per shared person · thick when 2+ facts agree"). Kit: Callout, warning tone, compact, no title, no button; not announced (it is a state of the drawing, read in order before the diagram). Two messages: **did not compare** — "Too many records to compare. The links between them are not drawn." The diagram draws every record and no arc (the app is sent none), and each record is spoken as its ref, graph and Source only ("PER-7KD45, Entry 1834/112, Hamburg parish register"), never as sharing no one. **Left some out** (more links than are sent, or more than 1,000 records) — "Too many records to draw them all. Some records and links are left out." Did not compare wins when both are true. A full read shows no note. |
| PN-W1 | Wide changes no notice's place. The page's notice keeps the header band's width; a section's notice keeps the page band's width when the section fills the pane (Timeline, Sources); the diagram's notice and note keep the pinned column's width beside the rows. |
| PN-W2 | Crossing 1600 keeps every notice and the keyboard on its Try again. At Regular, when the pane is too narrow for the rows and the column side by side, the column (its notice and note with it) sits under the rows, as today. |
| PN-10 | **Combined states.** When the Sources rows' first load fails or the person is gone, the record links column is not drawn (the column belongs to the rows). A gone section shows only its notice, no rows. A reload that fails over an empty section puts the notice above the EmptyState. At Wide the top of the pinned column aligns with the top of the rows, not with the rows' notice. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Person page, header band | Header, switcher, then the page's notice at the header's width | The same, at the header's width (the band keeps its measure) | Nothing |
| Person page, Family section | The section's notice, then the tree, as S13-D3 agreed | As S13-D3 agreed, with the notice at the band's width | Nothing |
| Person page, Timeline section | The section's notice, then the lifeline and the rows | Full width (Page columns' `primaryFill`): the notice keeps the band's width, the lifeline and rows fill the pane | Nothing |
| Person page, Sources section | The rows' notice above the rows; the record links column (notice, title, note, diagram) beside them, or under them when the rows would be narrower than their minimum | Full width (`primaryFill`): the rows' notice keeps the band's width; the column pinned beside the rows | The column only, as today |
| Event and Place pages | Header, then the page's notice, then Details | Centred at the Regular measure, as today (not drawn: nothing moves) | Nothing |

---

## 5. Suggested frames

Frames with an a / b pair are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others as they say. Sample error text for the frames (the engine's own message is shown; this is not new copy): "Something went wrong. Please try again."

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | PN-1, PN-4, PN-W1 | Family open on a person with two parents, four siblings, a spouse and three children. The section's reload failed: the notice "Family could not be read" with Try again above the tree, which is unchanged |
| 2a / 2b | PN-1, PN-5, PN-10, PN-W1 | Sources open on a person with records on four Sources, the column beside the rows. The rows' reload failed and the record links' first load failed: the rows' notice above the rows, the compact notice alone in the column, its top aligned with the rows' top |
| 2c | PN-W2 | Regular, pane 900 pt wide: frame 2 with the column under the rows (notice first) |
| 3a / 3b | PN-3, PN-W1 | Timeline open. The page's reload failed: the notice under the switcher in the pinned band, the Timeline under it unchanged |
| 4 | PN-2, PN-3 | Regular only, two states side by side: the page's first load failed (the notice alone at the top of the page); Timeline's first load failed under a loaded header (the notice in place of the lifeline) |
| 5 | PN-6 | Regular only: frame 1 while Try again runs (the Button's loading state, "Refreshing" beside the switcher), then after a second failure (the notice again, Try again with its focus ring) |
| 6a / 6b | PN-9, PN-W1 | Sources with the record links read that did not compare, the column beside the rows: 40 records on the circle, no arcs, the warning note under the title in place of the caption |
| 7 | PN-9 | Regular only: the column with the "left some out" note: 1,000 records drawn with their arcs, the note under the title |
| 8 | PN-8, PN-10 | Regular only, two states: the Toast "This person no longer exists. They may have been merged into another person." alone at the pane's toast position over a blank pane (the Persons list is not on this board); and the Sources section's notice for a gone person (title "Sources could not be read", that line, no Try again, no rows, no column) |
| 9 | PN-3 | Regular only: an Event page whose reload failed: the notice under the header, above Details |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Retry notice | Component | Extend | `macos/App/DesignSystem/Components/Callout/PVRetryCallout.swift` `PVRetryCallout` | Every failed-read notice on this board (PN-1 – PN-6), full or compact, at its call site: no wrapper. It is the Callout with one small secondary Button and announces. It gains what PN-6 and PN-7 need for every caller: speak again on each new failure, and keep the keyboard on Try again after a pressed retry fails. The Button reads "Try again", as on Promote. Not a second retry control, not a Toast |
| Callout | Component | Ship | `macos/App/DesignSystem/Components/Callout/PVCallout.swift` `PVCallout` | The gone notice in a section (danger, no button, PN-8) and the diagram's partial note (warning, compact, no title, not announced, PN-9) |
| Button | Component | Ship | `macos/App/DesignSystem/Components/Button/` | Inside the retry notice only: secondary, small, its loading state while Try again runs |
| Toast | Component | Ship | `macos/App/DesignSystem/Components/Toast/`, shown by `macos/App/Features/Workspace/WorkspaceNavigation.swift` `leavesForSectionRoot` | The gone line as the page leaves (PN-8); only its words change |
| Page frame | Snowflake | Extend | `macos/App/Features/Conclusions/ConclusionDetailPage.swift` | Gains the page's notice under the switcher (PN-3), shared by Person, Event and Place |
| Section host | Snowflake | Extend | `macos/App/Features/Conclusions/ConclusionPageSectionHost.swift` `ConclusionPageSectionHost` | Draws each section's notice (band) and the diagram's (column); today's plain Callout becomes the retry notice |
| Record links column | Snowflake | Extend | `macos/App/Features/Conclusions/PersonRecordLinksColumn.swift` | Gains the note in place of the caption (PN-9); the circle, nodes and title are unchanged |
| Refreshing spinner | Snowflake | Ship | `ConclusionDetailPage.swift` `ConclusionDetailHeaderBand` | Runs while Try again reads, as for any reload |

**Non-goals.** No new kit primitive (the retry notice is extended, not forked), no feature wrapper around it, no new tone, no new color token, no dismiss control, no new stroke style on the diagram (a link that was not compared is not drawn at all), no per-section copy of the stale line.

---

## 7. Handoff

When the board is agreed: record the frames on the **Board** line here, in [`README.md`](README.md) and on the plan's **Board** line; the plan ([`../prs/S16-04-person-page-notices.md`](../prs/S16-04-person-page-notices.md)) copies PN-1 – PN-10, PN-W1 and PN-W2 into its Scope and moves from `Blocked — on board S16-D1` to `Planned`. When S16-04 lands, the brief moves to the spike's design archive.
