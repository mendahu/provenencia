# S13-D4 — Person page: the Timeline section

> **Update 2026-11-02: the lifeline.** If you already started this board from the first version of this brief, **extend your frames**; do not start a second Timeline. For this revision:
> - **Add a lifeline (TL-5, replaces "period markers").** A line drawn **to scale** over the life, beside the event list: equal stretches of line are equal stretches of time, so silences in the record are long empty stretches. Gaps are drawn, never labelled. Each dated event points to its place: a point for a day, a short bar for a range, month or year, an open end for "bef." or "aft.". Rows stay a readable list (their heights vary), so draw how a row points to its place, and pick the line's orientation and how its scale fits a 90-year and a 3-year life. Year ticks or decade labels. No zoom, pan or swimlane.
> - **Birth and death always have a marker (new TL-5a).** A recorded birth or death sits at its date at the line's ends. When unrecorded, a placeholder with a question mark, "Birth unknown" or "Death unknown", with **no estimated year**: the birth placeholder just before the earliest dated event; the death placeholder just after the latest event that cannot come after death (burial and funeral can), so a recorded funeral with no death puts the placeholder **before** the funeral. A placeholder sits a short fixed distance from its neighbour, joined by a dashed stretch that is not to scale.
> - **Every event sits at its own date**, a burial after death included: the line runs on to it. No warning marks.
> - **Undated events move (TL-3).** They sit together under "Undated" **between the birth and death markers**, not at the end of the list. Decide where between them, and how the reader sees they have no place on the line.
> - **Folds:** the line still shows a point for every dated event, folded or not.
> - **The person tile is unchanged and the role sits only in a caption above it (TL-6)**; drop any "role on the tile" option. Participants lay out in an adaptive grid at the tile's minimum width (Wide adds columns).
> - **The event title** is a link-styled button with the kit focus ring.
> - **Crumbs (TL-8)** read *Persons › Weber, Anna › Timeline* (sorted title).
> - **New frames:** 6b *Unknown ends* (Johann without his birth and death: "Birth unknown" before the baptism, "Death unknown" after the 1861 marriage and before the burial); 6c *Short life* (a child born 1845, died 1848, three events). Frame 1 now shows the line from 1817 to 1880 with the 1861 – 1880 stretch empty and the burial past death; frame 6 shows both placeholders around two undated events; dark mode adds frame 6b.

**Kind:** Claude Design board (**enhancement** of the agreed S13-D1 Person page frame: it fills the Timeline section, which S13-D1 drew as an empty place, and places the S13-D3 person tile in a second view)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Timeline section
**Implements later as:** PR **S13-13** (the Timeline section).
**Board:** drawn 2026-11-02 — Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Timeline.dc.html*, turn 12 "S13-D4 · Timeline — the event list and the lifeline (update 2026-11-02)", frames 12a – 12y. Answers: rows divided by the kit rule, a 72pt picture, the title in the display face at heading size as a link button, the date in mono at body size (TL-1, TL-2); role Badges, Subject in the outline Badge (TL-4); a vertical lifeline in an 80pt column, stretched to the list's height, ticks every 1 – 50 years at least 34pt apart with quarter ticks when a year is tall, dots for days, bars for spans, open ends for bef./aft., curves from the rows in view, clickable points (TL-5); end caps for birth and death, a dashed "?" and a muted row 40pt from the neighbour for a placeholder (TL-5a); undated just before the death marker with the hint "No date, so no place on the line" (TL-3); participants as **compact 52pt cards** (40pt likeness, one-line name, life years; full lines in the tooltip and VoiceOver) at a 232pt grid minimum, the user's call, so the tile gains a compact size (TL-6; user, 2026-11-02); participants ordered by role then name, in Go (S13-33; user, 2026-11-02); at Wide the section fills the pane and the line keeps its column, no secondary column (TL-W1; user, 2026-11-02: full width). The S13-13 plan describes every frame in words.
**Depends on:** S13-33 (each event's place on the line and the birth and death markers; the data TL-5 and TL-5a draw); the agreed S13-D1 board (the frame, the header, the switcher; [archived brief](archive/S13-D1-person-page-frame.md)); the agreed S13-D3 board (the person tile and the Family section's states; [archived brief](archive/S13-D3-family-section.md), built by S13-12 [#676](https://github.com/mendahu/provenencia/pull/676)); the agreed S14-D4 board (the Events list row's thumbnail; [archived brief](../../spike-14/design/archive/S14-D4-events-list-thumbnail.md)); S13-05's timeline read (landed [#680](https://github.com/mendahu/provenencia/pull/680), the data this board draws); S13-32 (the tile as a kit shell with person content at the call site; no visible change, so the board draws the tile as S13-D3 did).
**Related:** [`../person-detail.md`](../person-detail.md) R-2, R-3, R-23 – R-28; [`../deployment-plan.md`](../deployment-plan.md) working rules "Timeline is direct participation, any role", "One person tile, built once", "Colors reuse `PersonGenderTreatment`", "A timeline tile opens the participant on Timeline", "Timeline lists fold"; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S13-13-timeline-section.md`](../prs/S13-13-timeline-section.md).
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

Extend the S13-D1 frames, with the S13-D3 person tile as drawn. The header, the switcher and the other four sections stay exactly as agreed; only the Timeline section's content changes, from its empty place (S13-D1 PPF-10) to the list of events. Draw on the S13-D1 Person page frame, with the switcher on **Timeline**. Do not start a second Person page, and do not redraw the person tile.

---

## 1. Objective

A researcher opens a Person's Timeline section and reads the person's life from birth to death on a **lifeline**, a line drawn to scale from birth to death, where the stretches with no records show as empty line, beside a list of events that point to their places on it: each event's picture, name, date and place, the person's role in it, and the other people who took part, each drawn as the same **person tile** as on Family. A click on an event opens its Event page; a click on a participant opens that person's Timeline, so the researcher can walk from one life to the next.

**What changes.** The Timeline section's empty place becomes the lifeline and the event list, with birth and death markers (placeholders when unrecorded) and the undated events between them, a state for a person with no events, loading and error states of its own, and "Show N more" rows where a long list folds. The S13-D3 tile appears in a second view, with each participant's role.

**What stays.** The header, the switcher, Vitals, Family, Gallery and the Sources empty place (S13-D1, S13-D3). The person tile's size, content and colors (S13-D3 FS-1, FS-2). The gender colors (S10-D12). The section switcher still names the section; Timeline has no heading of its own, as Vitals and Family have none.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Direct participation, any role (R-23) | The list holds every event the person took part in directly, in any role: their own birth (Subject), a child's baptism where they stood as Father, a neighbour's marriage they witnessed, a death they reported as Informant. A family event they are not recorded at (a child's birth with no Father named) is not listed. |
| Canonical only (R-2) | Every event and every participant is promoted. Nothing unpromoted is drawn, and no placeholder stands in for an unknown person. An event with no other promoted participant shows the event alone. |
| The order comes from the engine | Events arrive earliest first, in the Events list's date order: by date, else by start date; ties by ref. The board does not reorder them. Events with neither a date nor a start date arrive flagged undated; they keep their order among themselves. |
| The lifeline (R-24a; user, 2026-11-02) | A line drawn **to scale** over the life: equal stretches of line are equal stretches of time, so a 17-year silence in the records is a long empty stretch, and two events a month apart sit close. Each dated event has its place on it: a day ("12 Mar 1817"), or a span for a range, a month or a year ("1844 – 1849", "1835"); "abt. 1846" is a point; "bef. 1880" has an open start. Every event sits at its own date, even after death (a burial) or past where the board expected. Gaps are only drawn, never labelled. |
| Birth and death markers (user, 2026-11-02) | The line always starts at a birth marker and ends at a death marker. A recorded birth or death event is its marker, at its date. When there is none, a **placeholder** marks it ("Birth unknown", "Death unknown", with a question mark) and invents no date: the birth placeholder sits just before the earliest dated event, the death placeholder just after the latest dated event that cannot come after death. Burial and funeral may come after death, so a person with a recorded funeral and no death gets the death placeholder **before** the funeral. Placeholders sit a short fixed distance from their neighbour, joined by a dashed stretch that is not to scale. |
| Undated events (R-24; user, 2026-11-02) | They have no place on the line. They sit together **between the birth and death markers**, not at the end of the list, so they read as part of the life. |
| An event row's facts (R-25, R-28) | The event's **title** as every Events row titles it ("Birth of Johann Weber", "Marriage of Johann Weber and Maria Schulz", "Census at Hamburg", a recorded name, or the ref in mono); its **date** as the Events list words it ("12 Mar 1817", "abt. 1846", "1844 – 1849", "bef. 1880"), in mono; its **place** (the kept place's name, italic, plus a "+N" Badge when more places are kept); the **person's role** ("Subject", "Father", "Witness"). A row with no date or no place leaves that part out. |
| The person can hold several roles in one event | Usually one. Rarely two (Witness and Informant on one death): both are shown, joined as a list ("Witness, Informant"). |
| The event's picture (R-26; user, 2026-11-02) | Events have pictures now (Spike 14, S14-02): an Event leads with its chosen or earliest picture, as the Events list row and the Event page header show it. The slot shows that picture; an event with no picture shows the **Event mark** on the sunken thumbnail, as the Events list does. No new placeholder artwork. |
| Participants (R-27) | The other promoted people on the event, each with their role terms ("Mother", "Witness", "Subject"). They arrive in the engine's order (by the participation's ref), which reads as arbitrary; the board may ask for another order (by role, by title): say which, and the plan adds it in Go. The person themself is never among them. |
| The person tile (S13-D3 FS-1, FS-2) | Each participant is the S13-D3 tile, unchanged: 92pt tall, a 70pt likeness with the gender tint and line, the name (two lines, then cut), the b. and d. lines with a missing half left out. The tile has a declared minimum width; a grid of tiles shares the row's width at that minimum or more. The tile does not draw a role: the role is spoken by VoiceOver only, and the board draws where a participant's role is **seen**, outside the tile (TL-6). |
| A tile opens the participant's Timeline (user, 2026-11-02) | A click (or Return / Space with focus) on a participant's tile opens their Person page on its **Timeline** section. One navigation step: Back returns here. As on Family, walking a section stays in that section; every other link to a Person still lands on Vitals. |
| An event opens its Event page (R-28) | A click on the event row's title (or the row, as the board decides) opens the Event page. The Event page has no sections; it opens as it does from the Events list. |
| Long lists fold (user, 2026-11-02) | An event's participants fold at **15** behind a "Show N more" row; the event list folds at **50** behind a "Show N more" row (the dated events and the undated ones each). The lifeline still shows a point for every dated event, folded or not; the board draws how a folded event's point looks and what a click on it does. Nothing is ever hidden for good: the row shows the rest, and the fold returns when the researcher opens another person. Same "Show N more" row as Family's. |
| Sizes | Typical: 8 – 25 events, 0 – 4 participants each. Large: 120 events over a long life, a census of 12 household members, a burial with 41 people named (an obituary). Titles and places can be long ("Marriage of Johanna Wilhelmina Catharina van der Berg-Schoonhoven and Johann Weber"; "St. Ansgarii, Bremen, Freie Hansestadt Bremen, Deutscher Bund"). |
| The read can be slow or fail | The section has its own read. First view: a short load. A failure is a danger Callout in the section, not over the page. Refreshing (after an event or a participant changes) keeps the list drawn; the refresh shows at the end of the switcher row, as on Vitals and Family. |
| No events | A person with no promoted events shows an EmptyState. |

### 2.1 What this board is not

- Zoom, pan, a horizontal swimlane, or a canvas the researcher drags. The lifeline is drawn by the app at a scale it picks for the space it has.
- Labels on gaps ("no records 1862 – 1879"): the empty stretch says it (user, 2026-11-02).
- An estimated year for an unknown birth or death: the placeholder has none.
- A mark on an event that falls outside the life: it sits at its date like any other.
- The person's age at each event: not computed this spike.
- Editing: no add event, link, unlink, reorder or set-picture control, and no context menu on a row or tile.
- The provenance control on an event, a role or a tile (R-13b).
- Events the person is linked to only through a relative (a child's birth with no Father named) or through a place.
- A filter by role or event type, or a search box.
- A second person tile design.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-13** | TL-1 … TL-13 and TL-W1 … TL-W3: the Timeline section on the Person page, the lifeline with its birth and death markers, the event row with its picture slot, roles and participants, the undated events between birth and death, the folds, the empty, loading, error and refreshing states, keyboard and VoiceOver, light and dark, at Regular and Wide. The tile itself is unchanged: S13-32 first splits it into a content-agnostic kit shell and the person content Family and the Timeline build (user, 2026-11-02), with no visible change. |
| Later boards | None planned. Event and Place pages adopting sections (R-1) would reuse this row later. |
| Stays out | Any change to the header, switcher or other sections; the tile's size or content; the Event and Place pages. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| TL-1 | **The event row.** One row per event: the picture slot (TL-2), the title, the date and place line, the person's role (TL-4), and the participants (TL-6). Pick a row form from the kit: a Card per event, or rows divided by rules on one surface; say why. The title reads as a link to the Event page (R-28). |
| TL-2 | **The picture slot (R-26).** The event's lead picture on a Thumbnail, the size the Events list row uses or larger; the Event mark on the sunken tone when the event has no picture. Every row has the slot, so rows align. |
| TL-3 | **Order and the undated events (R-23, R-24).** Events earliest first, as given. The undated events sit together between the birth and death markers (user, 2026-11-02), under an "Undated" caption, in the same row form, with no place on the line. Decide where between the markers they sit (for example after the last dated event before death, or in a band of their own beside the line) and how the reader sees that they are not placed; say why. With no undated events there is no caption. |
| TL-4 | **The person's role (R-25).** The person's role on each row, plain to see without reading the participants: for example a Badge or a muted line ("as Witness"). Two roles read as a list. Decide how a "Subject" role reads on the person's own birth and death ("Subject" as worded, or a softer form); the role is always present (R-25). |
| TL-5 | **The lifeline (R-24a).** Draw the line to scale beside the list, with year ticks or decade labels so a place on it reads as a year. Each dated event points to its place: a point for a day, a short bar for a span, an open end for BEF or AFT. The rows stay a readable list (their heights vary with their participants), so a row cannot sit exactly at its date: draw how a row points to its place (a connector from the row to the line, a dot on the line that highlights its row on hover or focus, or another way) and say why. Several events close together on the line still each point to their own place. Pick the line's orientation (a vertical rail beside the list is the natural fit for a scrolling page) and how its scale fits the space (the whole life in the section's height, or a fixed number of points per year that scrolls with the list); say why, with a 90-year life and a 30-year one. Kit lines and tokens only. |
| TL-5a | **Birth and death markers.** The recorded birth and death as markers at their dates, ends of the line. A placeholder: a question mark and "Birth unknown" or "Death unknown", a short fixed distance from its neighbouring event, joined by a dashed stretch that is not to scale. Draw: both recorded; birth unknown (first record an 1835 census); death unknown with a recorded funeral (the placeholder before the funeral); both unknown; a person with no dated events (both placeholders and only undated events between them). |
| TL-6 | **Participants (R-27).** The other promoted participants under the event, each as the S13-D3 tile (not redrawn), laid out in a wrapping grid that fills the row's width at the tile's Family size. Each participant's role is **seen** as a caption above the tile (the Family captions' micro caps); the tile itself does not change (its height, content and slots are Family's). Decide the caption's spacing, and the grid's columns at Regular and Wide. An event with no other participant shows no participant area. Participants keep the engine's order unless the board asks for another (§ 2). |
| TL-7 | **Folding (user, 2026-11-02).** An event with more than 15 participants shows the first 15 and a "Show N more" row; the event list shows the first 50 events and a "Show N more" row, and the Undated group the same. Use the Family section's "Show N more" row (a ghost Button) unchanged. |
| TL-8 | **Opening.** A click on an event's title opens its Event page. A click on a participant's tile opens that person's Person page on **Timeline**. Draw the toolbar for one frame: after clicking Anna Weber's tile on Johann's birth, the breadcrumbs read *Persons › Weber, Anna › Timeline* (a Person crumb uses the sorted title, as Family's does) and the Back menu lists *Johann Weber · Timeline*. |
| TL-9 | **Long and sparse rows.** A long title wraps to two lines, then truncates, with the full title in the tooltip and VoiceOver. A long place truncates at its tail. A row with no date, or no place, or neither, leaves the part out; a row whose title is a ref shows it in mono. |
| TL-10 | **No events.** A person with no promoted events shows the kit EmptyState with the Timeline glyph (calendar), a title and one line on where events come from. Recommended: "No events yet" and "Events appear here when events this person took part in are promoted from an Evidence graph." |
| TL-11 | **Loading, error, refreshing.** First load: a small spinner where the list goes (the header and switcher already show). Failure: a danger Callout in the section titled "Timeline could not be read", with the error message. Refreshing: the list stays drawn; "Refreshing" at the end of the switcher row (S13-D1 PPF-8), as on Family. |
| TL-12 | **Keyboard and VoiceOver.** Tab moves through each event's title, then its participants' tiles, then any "Show N more" row, in reading order; Return or Space opens the focused item. VoiceOver reads each event as a group named by its title and role, for example "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject", with its title as a link and its tiles as buttons ("Mother, Anna Weber, born 3 May 1790, Bremen, died 1851, button"). The Undated caption is read as a heading. The birth and death markers are read in their place in the list ("Born 12 Mar 1817, Bremen", "Death unknown"); the line, ticks and connectors are decoration and are not read. |
| TL-13 | **Dark mode.** Frames 1, 5, 6b and 9 in dark mode: the lifeline, markers and placeholders, rows, pictures and the Event mark, role marks, tiles and every gender tint. |
| TL-W1 | **The list at Wide: explore a secondary column (user, 2026-11-02).** The lifeline stays beside the list at both widths, at its Regular measure. Draw two Wide options and recommend one. (A) **Primary-only**, as Vitals and Family: the rows keep their Regular measure and the participant grids add tiles per row (rule 3). (B) **A secondary column** of supporting content beside the list in Page columns' secondary column: for example a jump list of decades (or years) with event counts that scrolls the list to that period, or a short summary of the life (first and last dated event, how many events, how many undated). Rule 4 binds B: whatever sits in the secondary column at Wide is stacked above or below the list, or behind a disclosure, at Regular; draw that Regular placement too. Say which option ages better as the list grows to 120 events. |
| TL-W2 | **Tiles keep their size.** At both widths the participant tile is the Family tile's size; Wide adds tiles per row, it does not stretch them. |
| TL-W3 | **Crossing 1600 keeps the place (rule 5).** Resizing across the line keeps the section, the scroll position, the focused item and any unfolded list. A note on the Wide frame. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Timeline section | The lifeline and the event list in the primary column; participant tiles wrap within each row | (A) the same list, more tiles per row, centered primary-only; or (B) the list in the primary column and supporting content in the secondary column (TL-W1) | (A) nothing; (B) the secondary column's content sits stacked or behind a disclosure at Regular |
| Person tile | The Family tile's size | Same size | Nothing (rule 3: cards keep their size) |
| No events, loading, error | EmptyState, spinner or Callout in the primary column | Centered primary-only | Nothing |

---

## 5. Suggested frames

Each frame is the S13-D1 Person page pane only (no app sidebar), with the switcher on **Timeline**, at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). The Wide frames sit beside their Regular frame.

Sample life for frames 1 – 4: **Johann Weber** (the person, as on the Family board; man), b. 12 Mar 1817 · Bremen, d. 1880 · Hamburg. Half of the events have a picture; every participant is a promoted person from the Family board, with their gender colors and likenesses as drawn there.

| # | Event (title as the Events list gives it) | Date · place | Johann's role | Other participants (role) |
| --- | --- | --- | --- | --- |
| 1 | Birth of Johann Weber | 12 Mar 1817 · Bremen | Subject | Heinrich Weber (Father), Anna Weber (Mother) |
| 2 | Baptism of Johann Weber | 16 Mar 1817 · St. Ansgarii, Bremen | Subject | Heinrich Weber (Father), Anna Weber (Mother), Friedrich Kohl (Witness) |
| 3 | Census at Bremen | 1835 · Bremen | Subject | Heinrich, Anna, Margarethe, Friedrich, Wilhelm Weber (Subject each) |
| 4 | Marriage of Johann Weber and Maria Schulz | 4 Jun 1844 · Bremen | Subject | Maria Schulz (Subject), Friedrich Weber (Witness) |
| 5 | Baptism of Karl Weber | 2 Feb 1845 · Hamburg | Father | Karl Weber (Subject), Maria Schulz (Mother) |
| 6 | Census at Hamburg | abt. 1846 · Hamburg | Subject | Maria Schulz, Karl Weber (Subject each) |
| 7 | Marriage of Friedrich Weber and Clara Lenz | 1850 · Bremen | Witness | Friedrich Weber (Subject), Clara Lenz (Subject) |
| 8 | Death of Maria Schulz | 1859 · Hamburg | Witness, Informant | Maria Schulz (Subject) |
| 9 | Marriage of Johann Weber and Sophie Brandt | 1861 · Hamburg | Subject | Sophie Brandt (Subject) |
| 10 | Death of Johann Weber | 1880 · Hamburg | Subject | Otto Weber (Informant) |
| 11 | Burial of Johann Weber | bef. 1880 · Ohlsdorf, Hamburg | Subject | — |
| 12 | Residence at Altona (undated) | — · Altona | Subject | — |
| 13 | Migration (undated, no place) | — | Subject | Sophie Brandt (Subject) |

Not on the list: Luise's birth (1848), whose record names no father (R-23).

1. **Full timeline.** Johann's Timeline: the lifeline from his recorded birth (1817) to his recorded death (1880), events 1 – 11 pointing to their places, the 1861 – 1880 stretch empty, the burial just past death, and the undated 12 and 13 between birth and death (TL-1 – TL-6, TL-5a).
2. **Scrolled.** Frame 1 scrolled to the end: the death marker, the burial past it, and the undated events (TL-3, TL-5a).
3. **Row sheet.** One row in each state, side by side: with a picture and with the Event mark; one role and two roles; no participants, one, and five; a two-line title; a long place cut at its tail; no date; no place; a ref title in mono; hover and keyboard focus on the title and on a tile (TL-1, TL-2, TL-4, TL-6, TL-9).
4. **Large event.** The Burial of an obituary's subject with 41 participants: 15 tiles and "Show 26 more"; beside it, the same unfolded (TL-7).
5. **Long life.** A person with 120 events over 85 years: the first 50 and "Show 70 more", the line still showing all 120 points (TL-5, TL-7).
6. **Only undated.** A person whose two events have no date: both placeholders, the two events between them (TL-3, TL-5a).
6b. **Unknown ends.** Johann without event 1 or 10: "Birth unknown" just before the baptism, "Death unknown" just after the 1861 marriage and before the burial (TL-5a).
6c. **Short life.** A child born 1845 and died 1848 with three events: the scale at a 3-year life (TL-5).
7. **Opening a participant.** Frame 1 after clicking Anna Weber's tile on event 1: her Timeline, with the toolbar above the pane showing the breadcrumbs and the Back menu (TL-8).
8. **No events.** The EmptyState (TL-10).
9. **Loading, error, refreshing.** The section's first load, its error Callout, and frame 1 refreshing (TL-11).
10. **Keyboard.** Frame 1 with focus on Anna Weber's tile under event 1; the VoiceOver reading of event 1 and of the tile as a note (TL-12).
11. **Dark mode.** Frames 1, 3 and 8 in dark (TL-13).
12. **Wide, option A.** Frame 1 at Wide, the lifeline at its Regular measure, primary-only (TL-W1, TL-W2), with the rule-5 note (TL-W3).
13. **Wide, option B.** Frame 1 at Wide with the secondary column (TL-W1), and its Regular placement beside it.
14. **Wide large.** Frame 4 at Wide in the recommended option.
15. **Wide dark.** The recommended Wide frame in dark.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, header, switcher | Snowflake / component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `PVPageColumns`, `PVSegmentedControl` | As S13-D1. Timeline fills the primary column; the secondary column only if TL-W1 picks option B. |
| Timeline section | Snowflake | **New** | `Features/Conclusions/PersonTimelineSection.swift` *(new)* | The lifeline beside the event list, the markers, the undated events, the folds. Only the Person page uses it. Not a chart, not `GraphCanvas`. |
| Event row | Snowflake | **New** (inside the Timeline section) | `PersonTimelineSection.swift` | TL-1. A `PVCard` per event, or rows on one surface divided by the kit's rule; the board picks one. Not a new card style. |
| Event picture | Component | Ship | `CachedThumbnail` over `PVThumbnail`, fallback mark `subjectEvent` | TL-2. The Events list row's thumbnail and mark (S14-D4); no new placeholder artwork. |
| Event title | Snowflake | Ship | `ConclusionListRow.title(_:lineLimit: 2)` with `EventTitleDisplay`, as a link-styled button with the kit focus ring | TL-1, TL-9, TL-12. Name / italic label / mono ref, as every Events row. Return activates it through the kit's Return modifier (S13-32). |
| Date and place line | Snowflake | Ship | `EventSecondaryLine` and `EventSecondaryDisplay` (`Features/Conclusions/EventsListView.swift`) | TL-1, TL-9. The Events row's date (mono) · place (italic) · "+N" Badge, unchanged. |
| The person's role | Component | Ship | `PVBadge` (neutral, subtle) or muted caption text | TL-4. Worded from the role term the read gives, as the Evidence graph names roles. |
| Lifeline | Snowflake | **New** (inside the Timeline section) | `PersonTimelineLifeline.swift` *(new)* | TL-5, TL-5a. The line, ticks, event points, connectors and markers, drawn with `Path` in kit line and text tokens (`--border-strong` as the Family lines, micro caps for years). Not `GraphCanvas`, no new token. |
| Birth / death placeholder | Component | Ship | `PVBadge` (neutral) or muted text with a `questionmark` glyph | TL-5a. Not new artwork. |
| Person tile | Component (shell) + feature content | Ship | the tile shell S13-32 adds to the design system; the person content builder in `Features/Conclusions` | TL-6. The S13-D3 tile, unchanged, as Family draws it (S13-32 splits it, no visible change). Its role is spoken; a visible role is the caption above it. |
| Participant role caption | Component | Ship | micro caps text (`pvMicroCaps`), as the Family captions | TL-6. |
| Participant grid | Component | Ship | an adaptive `LazyVGrid` at the tile's minimum width, as the Gallery's likeness grid | TL-6, TL-W2. Tiles keep their minimum and wrap; Wide adds columns. |
| "Show N more" row | Snowflake | Ship | Family's "Show N more" row (`PersonFamilySection.swift`, a ghost `PVButton`) | TL-7. The same row and wording; the plan shares it rather than copying it. |
| Undated caption | Component | Ship | micro caps text, or `PVSectionHeader` | TL-3. Read as a heading. |
| No events | Component | Ship | `PVEmptyState` | TL-10. The Timeline glyph `calendar`, as S13-D1's empty place. |
| Loading, error, refreshing | Component | Ship | `ProgressView` small, `PVCallout` danger; the switcher row's refresh indicator | TL-11. As Family. |
| Secondary column content | Snowflake | **New**, only if TL-W1 picks B | `PersonTimelineSection.swift` | TL-W1. Composed from kit text, `PVButton` (ghost) or `PVChip` for jump targets; not a new navigation component. |
| Breadcrumbs, Back menu | Component | Ship | `PVBreadcrumbs`, the toolbar jump menu | TL-8. Unchanged; they already name the section. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Zoom, pan, a draggable canvas or a horizontal swimlane | § 2.1: the app draws the line at its own scale. |
| Gap labels, estimated years for unknown ends, or warning marks on events outside the life | User, 2026-11-02 (§ 2.1). |
| A second person tile, or a new tile size | Working rule "One person tile, built once"; TL-W2. |
| Placeholder artwork for an event with no picture | The Event mark is the fallback (S14-D4). |
| Age at each event | Not computed this spike (§ 2.1). |
| Add, link, unlink or reorder controls, or a context menu | This spike is view only. |
| The provenance control on a row, role or tile | R-13b. |
| A filter or search over the timeline | § 2.1. |
| A new Card style, color token or kit primitive | Card, Thumbnail, Badge and the tile cover it; one call site is not a primitive. |
| Placeholder tiles for unknown participants | R-2. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames; the row form (TL-1); the picture size (TL-2); the Undated caption with only undated events (TL-3); how the role reads, and how "Subject" reads (TL-4); the lifeline's orientation, scale and how a row points to its place (TL-5); the markers and placeholders (TL-5a); where a participant's role is seen and the participant order (TL-6); the empty-state copy (TL-10); the Wide option and, for B, its Regular placement (TL-W1); and every other change from this brief with the user's answer.
2. Refresh the [S13-13 plan](../prs/S13-13-timeline-section.md): copy the answers into its Decisions and **Design → macOS**, describe every frame in words, close its Open questions, run `review-plan` again, and set it to `Planned`.
3. S13-13 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
