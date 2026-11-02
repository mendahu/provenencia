# S13-D4 — Person page: the Timeline section

**Kind:** Claude Design board (**enhancement** of the agreed S13-D1 Person page frame: it fills the Timeline section, which S13-D1 drew as an empty place, and places the S13-D3 person tile in a second view)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Timeline section
**Implements later as:** PR **S13-13** (the Timeline section).
**Board:** not drawn yet. When it is agreed, record here: the Claude Design turn and frames, and the answers the Handoff (§ 7) lists.
**Depends on:** the agreed S13-D1 board (the frame, the header, the switcher; [archived brief](archive/S13-D1-person-page-frame.md)); the agreed S13-D3 board (the person tile and the Family section's states; [archived brief](archive/S13-D3-family-section.md), built by S13-12 [#676](https://github.com/mendahu/provenencia/pull/676)); the agreed S14-D4 board (the Events list row's thumbnail; [archived brief](../../spike-14/design/archive/S14-D4-events-list-thumbnail.md)); S13-05's timeline read (landed [#680](https://github.com/mendahu/provenencia/pull/680), the data this board draws); S13-32 (the tile as a kit shell with person content at the call site; no visible change, so the board draws the tile as S13-D3 did).
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

A researcher opens a Person's Timeline section and reads the person's life from birth to death as a list of events: each event's picture, name, date and place, the person's role in it, and the other people who took part, each drawn as the same **person tile** as on Family. A click on an event opens its Event page; a click on a participant opens that person's Timeline, so the researcher can walk from one life to the next.

**What changes.** The Timeline section's empty place becomes the event list, with an Undated group at its end, a state for a person with no events, loading and error states of its own, and "Show N more" rows where a long list folds. The S13-D3 tile appears in a second view, with each participant's role.

**What stays.** The header, the switcher, Vitals, Family, Gallery and the Sources empty place (S13-D1, S13-D3). The person tile's size, content and colors (S13-D3 FS-1, FS-2). The gender colors (S10-D12). The section switcher still names the section; Timeline has no heading of its own, as Vitals and Family have none.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Direct participation, any role (R-23) | The list holds every event the person took part in directly, in any role: their own birth (Subject), a child's baptism where they stood as Father, a neighbour's marriage they witnessed, a death they reported as Informant. A family event they are not recorded at (a child's birth with no Father named) is not listed. |
| Canonical only (R-2) | Every event and every participant is promoted. Nothing unpromoted is drawn, and no placeholder stands in for an unknown person. An event with no other promoted participant shows the event alone. |
| The order comes from the engine | Events arrive earliest first, in the Events list's date order: by date, else by start date; ties by ref. The board does not reorder them. Events with neither a date nor a start date arrive last, flagged undated (R-24); they keep their order inside the Undated group. |
| An event row's facts (R-25, R-28) | The event's **title** as every Events row titles it ("Birth of Johann Weber", "Marriage of Johann Weber and Maria Schulz", "Census at Hamburg", a recorded name, or the ref in mono); its **date** as the Events list words it ("12 Mar 1817", "abt. 1846", "1844 – 1849", "bef. 1880"), in mono; its **place** (the kept place's name, italic, plus a "+N" Badge when more places are kept); the **person's role** ("Subject", "Father", "Witness"). A row with no date or no place leaves that part out. |
| The person can hold several roles in one event | Usually one. Rarely two (Witness and Informant on one death): both are shown, joined as a list ("Witness, Informant"). |
| The event's picture (R-26; user, 2026-11-02) | Events have pictures now (Spike 14, S14-02): an Event leads with its chosen or earliest picture, as the Events list row and the Event page header show it. The slot shows that picture; an event with no picture shows the **Event mark** on the sunken thumbnail, as the Events list does. No new placeholder artwork. |
| Participants (R-27) | The other promoted people on the event, each with their role terms ("Mother", "Witness", "Subject"). They arrive in the engine's order (by the participation's ref), which reads as arbitrary; the board may ask for another order (by role, by title): say which, and the plan adds it in Go. The person themself is never among them. |
| The person tile (S13-D3 FS-1, FS-2) | Each participant is the S13-D3 tile, unchanged: 92pt tall, a 70pt likeness with the gender tint and line, the name (two lines, then cut), the b. and d. lines with a missing half left out. The tile has a declared minimum width; a grid of tiles shares the row's width at that minimum or more. The tile does not draw a role: the role is spoken by VoiceOver only, and the board draws where a participant's role is **seen**, outside the tile (TL-6). |
| A tile opens the participant's Timeline (user, 2026-11-02) | A click (or Return / Space with focus) on a participant's tile opens their Person page on its **Timeline** section. One navigation step: Back returns here. As on Family, walking a section stays in that section; every other link to a Person still lands on Vitals. |
| An event opens its Event page (R-28) | A click on the event row's title (or the row, as the board decides) opens the Event page. The Event page has no sections; it opens as it does from the Events list. |
| Long lists fold (user, 2026-11-02) | An event's participants fold at **15** behind a "Show N more" row; the event list folds at **50** behind a "Show N more" row (the dated list and the Undated group each). Nothing is ever hidden for good: the row shows the rest, and the fold returns when the researcher opens another person. Same "Show N more" row as Family's. |
| Sizes | Typical: 8 – 25 events, 0 – 4 participants each. Large: 120 events over a long life, a census of 12 household members, a burial with 41 people named (an obituary). Titles and places can be long ("Marriage of Johanna Wilhelmina Catharina van der Berg-Schoonhoven and Johann Weber"; "St. Ansgarii, Bremen, Freie Hansestadt Bremen, Deutscher Bund"). |
| The read can be slow or fail | The section has its own read. First view: a short load. A failure is a danger Callout in the section, not over the page. Refreshing (after an event or a participant changes) keeps the list drawn; the refresh shows at the end of the switcher row, as on Vitals and Family. |
| No events | A person with no promoted events shows an EmptyState. |

### 2.1 What this board is not

- A chart or a scaled time axis: no proportional spacing by year, no zoom, no horizontal swimlane. The list is in date order; the board may add year or decade markers as text (TL-5).
- The person's age at each event, or the gap between events: not computed this spike.
- Editing: no add event, link, unlink, reorder or set-picture control, and no context menu on a row or tile.
- The provenance control on an event, a role or a tile (R-13b).
- Events the person is linked to only through a relative (a child's birth with no Father named) or through a place.
- A filter by role or event type, or a search box.
- A second person tile design.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-13** | TL-1 … TL-13 and TL-W1 … TL-W3: the Timeline section on the Person page, the event row with its picture slot, roles and participants, the Undated group, the folds, the empty, loading, error and refreshing states, keyboard and VoiceOver, light and dark, at Regular and Wide. The tile itself is unchanged: S13-32 first splits it into a content-agnostic kit shell and the person content Family and the Timeline build (user, 2026-11-02), with no visible change. |
| Later boards | None planned. Event and Place pages adopting sections (R-1) would reuse this row later. |
| Stays out | Any change to the header, switcher or other sections; the tile's size or content; the Event and Place pages. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| TL-1 | **The event row.** One row per event: the picture slot (TL-2), the title, the date and place line, the person's role (TL-4), and the participants (TL-6). Pick a row form from the kit: a Card per event, or rows divided by rules on one surface; say why. The title reads as a link to the Event page (R-28). |
| TL-2 | **The picture slot (R-26).** The event's lead picture on a Thumbnail, the size the Events list row uses or larger; the Event mark on the sunken tone when the event has no picture. Every row has the slot, so rows align. |
| TL-3 | **Order and the Undated group (R-23, R-24).** Events earliest first, as given. After the dated events, an "Undated" group: a caption, then its events in the same row form. With no undated events there is no group; with only undated events, the group is the whole list (decide whether its caption still shows; say why). |
| TL-4 | **The person's role (R-25).** The person's role on each row, plain to see without reading the participants: for example a Badge or a muted line ("as Witness"). Two roles read as a list. Decide how a "Subject" role reads on the person's own birth and death ("Subject" as worded, or a softer form); the role is always present (R-25). |
| TL-5 | **Reading the life.** Help the eye find a period: for example a year in the row's margin, a year or decade caption where the year changes, or a rail joining the rows. Text and kit lines only (§ 2.1). Recommend one, and say why. |
| TL-6 | **Participants (R-27).** The other promoted participants under the event, each as the S13-D3 tile (not redrawn), laid out in a wrapping grid that fills the row's width at the tile's Family size. Each participant's role is **seen** as a caption above the tile (the Family captions' micro caps); the tile itself does not change (its height, content and slots are Family's). Decide the caption's spacing, and the grid's columns at Regular and Wide. An event with no other participant shows no participant area. Participants keep the engine's order unless the board asks for another (§ 2). |
| TL-7 | **Folding (user, 2026-11-02).** An event with more than 15 participants shows the first 15 and a "Show N more" row; the event list shows the first 50 events and a "Show N more" row, and the Undated group the same. Use the Family section's "Show N more" row (a ghost Button) unchanged. |
| TL-8 | **Opening.** A click on an event's title opens its Event page. A click on a participant's tile opens that person's Person page on **Timeline**. Draw the toolbar for one frame: after clicking Anna Weber's tile on Johann's birth, the breadcrumbs read *Persons › Weber, Anna › Timeline* (a Person crumb uses the sorted title, as Family's does) and the Back menu lists *Johann Weber · Timeline*. |
| TL-9 | **Long and sparse rows.** A long title wraps to two lines, then truncates, with the full title in the tooltip and VoiceOver. A long place truncates at its tail. A row with no date, or no place, or neither, leaves the part out; a row whose title is a ref shows it in mono. |
| TL-10 | **No events.** A person with no promoted events shows the kit EmptyState with the Timeline glyph (calendar), a title and one line on where events come from. Recommended: "No events yet" and "Events appear here when events this person took part in are promoted from an Evidence graph." |
| TL-11 | **Loading, error, refreshing.** First load: a small spinner where the list goes (the header and switcher already show). Failure: a danger Callout in the section titled "Timeline could not be read", with the error message. Refreshing: the list stays drawn; "Refreshing" at the end of the switcher row (S13-D1 PPF-8), as on Family. |
| TL-12 | **Keyboard and VoiceOver.** Tab moves through each event's title, then its participants' tiles, then any "Show N more" row, in reading order; Return or Space opens the focused item. VoiceOver reads each event as a group named by its title and role, for example "Birth of Johann Weber, 12 Mar 1817, Bremen, as Subject", with its title as a link and its tiles as buttons ("Mother, Anna Weber, born 3 May 1790, Bremen, died 1851, button"). The Undated caption is read as a heading. Decoration (rails, year markers) is not read twice. |
| TL-13 | **Dark mode.** Frames 1, 5 and 9 in dark mode: rows, pictures and the Event mark, role marks, tiles and every gender tint. |
| TL-W1 | **The list at Wide: explore a secondary column (user, 2026-11-02).** Draw two Wide options and recommend one. (A) **Primary-only**, as Vitals and Family: the rows keep their Regular measure and the participant grids add tiles per row (rule 3). (B) **A secondary column** of supporting content beside the list in Page columns' secondary column: for example a jump list of decades (or years) with event counts that scrolls the list to that period, or a short summary of the life (first and last dated event, how many events, how many undated). Rule 4 binds B: whatever sits in the secondary column at Wide is stacked above or below the list, or behind a disclosure, at Regular; draw that Regular placement too. Say which option ages better as the list grows to 120 events. |
| TL-W2 | **Tiles keep their size.** At both widths the participant tile is the Family tile's size; Wide adds tiles per row, it does not stretch them. |
| TL-W3 | **Crossing 1600 keeps the place (rule 5).** Resizing across the line keeps the section, the scroll position, the focused item and any unfolded list. A note on the Wide frame. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Timeline section | The event list in the primary column; participant tiles wrap within each row | (A) the same list, more tiles per row, centered primary-only; or (B) the list in the primary column and supporting content in the secondary column (TL-W1) | (A) nothing; (B) the secondary column's content sits stacked or behind a disclosure at Regular |
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

1. **Full timeline.** Johann's Timeline: events 1 – 11, then the Undated group with 12 and 13, scrolled to the top (TL-1 – TL-6).
2. **Scrolled.** Frame 1 scrolled to the end: the Undated group (TL-3).
3. **Row sheet.** One row in each state, side by side: with a picture and with the Event mark; one role and two roles; no participants, one, and five; a two-line title; a long place cut at its tail; no date; no place; a ref title in mono; hover and keyboard focus on the title and on a tile (TL-1, TL-2, TL-4, TL-6, TL-9).
4. **Large event.** The Burial of an obituary's subject with 41 participants: 15 tiles and "Show 26 more"; beside it, the same unfolded (TL-7).
5. **Long life.** A person with 120 events: the first 50 and "Show 70 more" (TL-7), with TL-5's period markers.
6. **Only undated.** A person whose two events have no date (TL-3).
7. **Opening a participant.** Frame 1 after clicking Anna Weber's tile on event 1: her Timeline, with the toolbar above the pane showing the breadcrumbs and the Back menu (TL-8).
8. **No events.** The EmptyState (TL-10).
9. **Loading, error, refreshing.** The section's first load, its error Callout, and frame 1 refreshing (TL-11).
10. **Keyboard.** Frame 1 with focus on Anna Weber's tile under event 1; the VoiceOver reading of event 1 and of the tile as a note (TL-12).
11. **Dark mode.** Frames 1, 3 and 8 in dark (TL-13).
12. **Wide, option A.** Frame 1 at Wide, primary-only (TL-W1, TL-W2), with the rule-5 note (TL-W3).
13. **Wide, option B.** Frame 1 at Wide with the secondary column (TL-W1), and its Regular placement beside it.
14. **Wide large.** Frame 4 at Wide in the recommended option.
15. **Wide dark.** The recommended Wide frame in dark.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, header, switcher | Snowflake / component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `PVPageColumns`, `PVSegmentedControl` | As S13-D1. Timeline fills the primary column; the secondary column only if TL-W1 picks option B. |
| Timeline section | Snowflake | **New** | `Features/Conclusions/PersonTimelineSection.swift` *(new)* | The event list, the Undated group, the folds. Only the Person page uses it. Not a chart, not `GraphCanvas`. |
| Event row | Snowflake | **New** (inside the Timeline section) | `PersonTimelineSection.swift` | TL-1. A `PVCard` per event, or rows on one surface divided by the kit's rule; the board picks one. Not a new card style. |
| Event picture | Component | Ship | `CachedThumbnail` over `PVThumbnail`, fallback mark `subjectEvent` | TL-2. The Events list row's thumbnail and mark (S14-D4); no new placeholder artwork. |
| Event title | Snowflake | Ship | `ConclusionListRow.title(_:lineLimit: 2)` with `EventTitleDisplay`, as a link-styled button with the kit focus ring | TL-1, TL-9, TL-12. Name / italic label / mono ref, as every Events row. Return activates it through the kit's Return modifier (S13-32). |
| Date and place line | Snowflake | Ship | `EventSecondaryLine` and `EventSecondaryDisplay` (`Features/Conclusions/EventsListView.swift`) | TL-1, TL-9. The Events row's date (mono) · place (italic) · "+N" Badge, unchanged. |
| The person's role | Component | Ship | `PVBadge` (neutral, subtle) or muted caption text | TL-4. Worded from the role term the read gives, as the Evidence graph names roles. |
| Period markers | Snowflake | **New** (inside the Timeline section), if TL-5 uses them | `PersonTimelineSection.swift` | TL-5. Text in the kit's micro caps or caption style, and one kit line token if a rail is drawn. No new token. |
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
| A scaled time axis, a chart, zoom or a horizontal swimlane | § 2.1: a list in date order. |
| A second person tile, or a new tile size | Working rule "One person tile, built once"; TL-W2. |
| Placeholder artwork for an event with no picture | The Event mark is the fallback (S14-D4). |
| Age at the event or gaps between events | Not computed this spike (§ 2.1). |
| Add, link, unlink or reorder controls, or a context menu | This spike is view only. |
| The provenance control on a row, role or tile | R-13b. |
| A filter or search over the timeline | § 2.1. |
| A new Card style, color token or kit primitive | Card, Thumbnail, Badge and the tile cover it; one call site is not a primitive. |
| Placeholder tiles for unknown participants | R-2. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames; the row form (TL-1); the picture size (TL-2); the Undated caption with only undated events (TL-3); how the role reads, and how "Subject" reads (TL-4); the period markers (TL-5); where a participant's role is seen and the participant order (TL-6); the empty-state copy (TL-10); the Wide option and, for B, its Regular placement (TL-W1); and every other change from this brief with the user's answer.
2. Refresh the [S13-13 plan](../prs/S13-13-timeline-section.md): copy the answers into its Decisions and **Design → macOS**, describe every frame in words, close its Open questions, run `review-plan` again, and set it to `Planned`.
3. S13-13 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
