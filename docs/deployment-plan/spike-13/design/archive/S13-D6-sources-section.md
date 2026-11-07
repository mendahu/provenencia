# S13-D6 — Person page: the Sources section, Remove and the last-claim warning

**Kind:** Claude Design board (**enhancement** of the agreed S13-D1 Person page frame: it fills the Sources section, which S13-D1 drew as a placeholder place (PPF-10), and adds the Remove confirm over the page; a **rethink** of this board's first Sources draft, user 2026-11-05)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Sources section; the Remove confirm
**Implements later as:** PR **S13-16** (the Sources section) and PR **S13-17** (Remove, the last-claim warning, landing on the Persons list).
**Board:** agreed 2026-11-05 — Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail - Sources.dc.html*, turn "S13-D6 rev 2 · Sources as Source rows, Remove and the last-record warning", frames 14a – 14y (the first, Subject-card draft was rejected the same day). Answers: a micro-caps column header line SOURCE · TYPE · DATE · REF, provenance rhythm (8pt row padding, a subtle hairline between Sources, a strong rule under the header) (SS-1); the title is the provenance citation link, two lines then cut, no copy button (SS-2); Citation flexible · Type 176pt · Date 120pt · Ref 104pt, 14pt gaps (SS-3); **no record count (user, 2026-11-05; departs from SS-4)**; a caption graph line indented 20pt, no glyph, the Subject name cut first (SS-5); **Remove is a ghost Button with the Unpromote glyph and the word, 6pt after the line's last ref (user, 2026-11-05; departs from RC-1's IconButton at the far right)**; the empty copy as written (SS-7); a Source with more than six lines shows five and "Show N more" (SS-9); How the records connect: the Wide secondary column and a band above the rows at Regular, drawn on the frames only; Wide today primary-only centered at 1180pt (SS-11, SS-W1); no name beside the ref (RC-3); no Toast (RC-6). The web kit's stand-ins (the type pill composed from Mark; arrow-down-to-line for the Unpromote glyph) are `CatalogSourceTypePill` and `PVSymbol.unpromote` in the app.
**Depends on:** the agreed S13-D2 and S13-D13 boards (the provenance sheet's row and table, which this section's rows borrow; [S13-D2](archive/S13-D2-provenance-sheet.md), [S13-D13](archive/S13-D13-provenance-sheet-in-the-kit.md)); the agreed S13-D1 board (the frame, the header, the switcher; [archived brief](archive/S13-D1-person-page-frame.md), built by S13-10 [#634](https://github.com/mendahu/provenencia/pull/634)); the Unpromote confirm and its last-member wording (S13-08, [#707](https://github.com/mendahu/provenencia/pull/707)); the agreed S13-D4 and S13-D5 boards (Page columns' full-width mode; [S13-D4](archive/S13-D4-timeline-section.md), [S13-D5](archive/S13-D5-gallery-and-viewer.md)).
**Related:** [`../person-detail.md`](../../person-detail.md) R-33 – R-39 (and R-2, R-13b); [`../deployment-plan.md`](../../deployment-plan.md) S13-16 and S13-17 scope entries, working rules "Sections are places", "One Unpromote, deletes through the registry", "A full-width section keeps its states in the band", "Claims are read per person from the graph cache, grouped by Source in Go"; [`macos-client-patterns.md` § 9](../../../../macos-client-patterns.md#9-window-widths-regular-and-wide); slice 7's later board S13-D10 ([`../evidence-graph-comparison.md`](../../evidence-graph-comparison.md)), which sits in or beside this section.
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** [`S13-D6-sources-section.review.md`](S13-D6-sources-section.review.md) — round 1 on the card version, every finding Fixed; revision 2 (Source rows, 2026-11-05) not reviewed yet

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

### This brief is an enhancement of the page, and a rethink of this section's first draft

Extend the S13-D1 frames. The header, the switcher and the other four sections stay exactly as agreed (S13-D1, S13-D3, S13-D4, S13-D5); only the Sources section's content changes, and the Remove confirm is added over the page. Draw on the S13-D1 Person page frame, with the switcher on **Sources**. Do not start a second Person page.

**Rethink:** the board's first Sources draft drew each claim as the Evidence graph's Subject card. Throw those frames away (user, 2026-11-05: too much information, the wrong emphasis). Do not keep a card variant. The Evidence graph canvas is not on this board; its Unpromote confirm keeps its look (this board's confirm is the same confirm, RC-2).

---

## 1. Objective

A researcher opens a Person's Sources section and sees **the Sources behind the person**: one row per Source, laid out like a row of the provenance sheet: its citation (the Source part only), a link to the Source page, and its key data (type, publication date) in their own columns. Under each Source, small and secondary, one link per Evidence graph of that Source that holds a record of this person, with the graph's and the record's refs. A graph link opens that Evidence graph with the person's Subject selected and in view. On each record the researcher can **Remove** it: the Subject is unpromoted from this person, exactly as Unpromote on the Evidence graph does. Removing the last record deletes the person; the confirm says so first, and on proceeding the app lands on the Persons list.

The section also **keeps a place for "How the records connect"** (slice 7, board S13-D10), which will show how the person's Evidence graphs link to each other. This board draws where it goes, empty; S13-D10 designs it.

**What changes.** The Sources place, today an EmptyState ("Sources arrives in a later version of Provenencia."), shows the Source rows and their graph links. The Remove confirm opens over the page.

**What stays.** The header, the switcher, Vitals, Family, Timeline and Gallery. The provenance sheet's look (S13-D2, S13-D13), which this section borrows from. The Evidence graph's own Unpromote confirm (title, copy, the "Also unpromoted" list, the danger tone when a handle is deleted): the Person page's confirm is the same one (user, 2026-11-05). One wording fix rides along: when only the subject's own handle is deleted, the graph's confirm today shows the generic DeleteImpact line; it takes the per-kind sentence approved for S13-08 on 2026-11-04 ("PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph."), on both pages.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| An identity claim (R-33) | One accepted promotion of a Subject (a person as one Source records them) into this canonical person. On this board a claim is a **record**: the Subject on one Evidence graph. Nothing unpromoted is drawn (R-2). |
| Source rows (R-34, R-36) | A Source can hold several Evidence graphs (a census has a graph per household page), so one Source can hold several of the person's records. One row per Source; its records sit under it. Rows are in Source title order; within a Source, records are in Evidence graph name order. The engine orders them; the board does not reorder. |
| The Source's citation | The Source part of the provenance sheet's citation line: the Source's **title**, else its ref ("SRC-4KQ7M") in mono. Not the Artifact, Citation or page: those belong to one Observation, and this row stands for the whole Source. When Spike 15 adds citation fields (author, publisher, place), the Source part of a formatted citation lands here. |
| The Source's key data (R-35) | **Type** ("Census", "Church register", "Newspaper"), shown with the type's icon as the Source page header's type pill shows it; **publication date**, which Sources do not have yet: Spike 15 adds it (user, 2026-11-05: reserve the column). Until then the date column shows "—". The Source's **ref** ("SRC-9TQ2M") in mono. |
| Open the Source (R-37) | The Source page. |
| An Evidence graph link | Secondary. One line per record: the graph's name ("Household 214") and ref ("EVG-2K7PQ"), the record's Subject label as the Source wrote it ("J. Weber") and its ref ("CPR-2M8XF"), in mono for the refs. A Source with one graph still shows its one link. Two records of the person on one graph (a duplicate) are two lines naming the same graph. |
| Open the graph (user, 2026-11-05) | The Evidence graph, with this record's Subject selected and scrolled into view. Back returns to the Sources section. |
| Remove (R-38) | Unpromote this record's Subject from the person. It takes back the Subject's claim and the claims of its relationships, participations and locations that have this Subject as an end; the cards stay on the Evidence graph. The confirm is the Evidence graph's Unpromote confirm (user, 2026-11-05): title "Unpromote J. Weber?", the message, the "Also unpromoted" list of those connections with "The other end of each stays promoted.", "Unpromote" and "Keep promotion". |
| The last record (R-39) | When the record is the person's last, Remove deletes the person. The engine says so before the write, and the confirm turns to the danger tone with "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." When the check cannot run, the confirm says the person may be deleted. On proceeding, the app lands on the **Persons list**, and Back does not return to the deleted page. |
| Another handle can go too | Unpromoting a Subject can empty a relationship or participation handle whose last member was one of its connections; the confirm's message names them all with the kit DeleteImpact line. |
| How the records connect (slice 7, S13-D10) | Later, the page will show the person's Evidence graphs as connected records: which graphs agree, which conflict, which stand apart. It is about the same graphs this section links. Its form and its size are S13-D10's; this board only reserves its place (SS-11). |
| The read and refresh | The section reads its own query (the person's records by Source), with its own spinner on the first load and "Refreshing" at the end of the switcher row on a reload (S13-D1 PPF-8). A Remove that keeps the person refreshes the section: the removed line leaves, an emptied Source row leaves. |
| Sizes | Typical: 3 – 20 records from 2 – 12 Sources. Large: 50 records from 14 Sources, up to 9 under one Source. Source titles can be long ("Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten"). |

### 2.1 What this board is not

- Subject cards, property rows, Observations or values of any record (user, 2026-11-05).
- The citation of one Observation (Artifact, Citation, page): this section cites Sources, not records.
- Setting the Source date, or sorting by it (Spike 15). The column is reserved and shows "—".
- How the records connect itself (S13-D10): only its reserved place.
- Editing a Source, a Subject or a claim from this page; Promote.
- A provenance control on a row (R-13b).
- A claim review, pins, a confidence or an argument (descoped 2026-10-29); rejecting a claim or marking it provisional.
- A bulk Remove, a multi-select, a filter or a search.
- A change to the Evidence graph canvas; its confirm changes only by the last-member sentence (§ 1).

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-16** | SS-1 … SS-10 and SS-W1 … SS-W3: the Sources section, its Source rows (citation, Source link, type, the reserved date column, ref), the graph links, the empty, loading, error and refreshing states, keyboard and VoiceOver, light and dark, at Regular and Wide. The reserved place for How the records connect (SS-11) is a layout rule only: S13-16 draws nothing in it. Each graph link's Remove slot stays empty until S13-17. |
| Ships in **S13-17** | RC-1 … RC-7: Remove on each graph link, the confirm (the Evidence graph's confirm, shared), the last-record warning, the pending and failure states, landing on the Persons list, and the section after a Remove that keeps the person. |
| Later | Spike 15 fills the date column. S13-D10 / S13-28 fill the reserved place. |
| Stays out | Any change to the header, the switcher, the other sections, the Evidence graph canvas, or its confirm beyond the shared last-member sentence (§ 1). |

---

## 4. Requirements

### 4.1 Sources section

| ID | Requirement |
| --- | --- |
| SS-1 | **The Source rows.** One row per Source, in the engine's order, drawn like a row of the provenance sheet's record table (S13-D2, S13-D13): the same row height, padding, hairline separators and column rhythm, a table the eye can scan down. Under each Source row, its graph links (SS-5). Decide whether the rows carry a column header line (as the provenance sheet's table does) or not; say why. |
| SS-2 | **The citation column.** The widest column, first: the Source's title in the provenance sheet's citation type (else its ref in mono), up to two lines then cut at its tail, the full title in the tooltip. The title is the link to the Source page (R-37): the same link treatment the provenance sheet uses for its citation line, with the kit focus ring and Return. Decide whether a copy button sits beside it, as on the provenance sheet; recommended: no (there is no citation to copy until Spike 15 formats one). |
| SS-3 | **The key-data columns.** Fixed-width columns after the citation: **Type**, the Source type pill (icon and label, as the Source page header shows it); **Date**, the publication date, which shows "—" in muted text until Spike 15; **Ref**, the Source's ref in mono. Keep a column of the same width for the date, so Spike 15 fills it with no relayout. Decide the column order and widths at Regular; say why. |
| SS-4 | **Record count.** Answered on the board (user, 2026-11-05): **no count**. The graph lines under each Source are easy to count, and they are Evidence graphs, not a number to read. |
| SS-5 | **The graph links (secondary).** Under each Source row, indented to the citation column, one compact line per record: a small link with the Evidence graph's name, then in mono caption its graph ref, then the record's Subject label and its ref ("Household 214 · EVG-2K7PQ · J. Weber · CPR-2M8XF"). The line is visibly secondary to the Source row: smaller type, muted, tight spacing; keep it to one line, cutting the Subject label first. The graph name is the link (opens the graph, § 2). At the line's end, a reserved slot for Remove (RC-1). Decide whether the lines carry a leading graph glyph; say why. |
| SS-6 | **Hover and focus.** The Source title link and each graph link take hover and the kit focus ring; a row does not lift as a whole. Draw both states. |
| SS-7 | **No records.** A person with no accepted record is deleted by the engine (R-39), so this state shows only when the read races a delete. The kit EmptyState with the Sources glyph (library): "No records" · "This person has no promoted records." Decide the copy; keep it short. |
| SS-8 | **Loading, error, refreshing.** First load: the section's spinner, as Family and Timeline. A failed read: the danger Callout "Sources could not be read" with the error (the Family and Timeline wording). A reload keeps the rows drawn and shows "Refreshing" at the end of the switcher row (PPF-8). |
| SS-9 | **Large.** 50 records from 14 Sources, one with 9: every Source row and every graph link is shown. Decide whether a Source's graph links fold behind "Show N more" (the Family and Timeline row) and at what count; recommended: fold past 5 links under one Source. |
| SS-10 | **Keyboard, VoiceOver, dark mode.** Tab goes through each Source's title link, then its graph links (then each Remove, S13-17), in reading order. VoiceOver reads a Source row as one element: "Hamburg census 1867, Census, date not recorded, SRC-9TQ2M, 2 records", then its title as a link "Open Hamburg census 1867"; a graph link as "Open Household 214, record J. Weber, CPR-2M8XF". Frames 1, 3 and 5 also in dark. |
| SS-11 | **Room for How the records connect (S13-D10).** Mark where that view will sit, as an empty dashed region labelled "How the records connect — S13-D10" on the frames only (it ships as nothing). Recommended: at Wide, the secondary column beside the Source rows (it is supporting content about the same graphs); at Regular, a band above the Source rows, full width of the primary column. The Source rows' layout must not move when it is added: at Wide they keep their place in the primary column; at Regular they move down by the band's height and nothing else. Say why if the board places it elsewhere. |
| SS-W1 | **Sources at Wide.** Primary column at the Regular measure holding the Source rows, the key-data columns keeping their widths and the citation column its measure (rule 2); the secondary column is kept for How the records connect (SS-11). Until S13-28 ships it, the group of columns is centered as primary-only (rule 1: the section has nothing else to put there now). Draw both: today's centered frame, and the frame with the reserved secondary region. |
| SS-W2 | **Crossing 1600 keeps the place (rule 5).** Resizing keeps the section, the scroll position and the focused link. A note on the Wide frame. |
| SS-W3 | **States at Wide.** The EmptyState, the spinner and the Callout stay centered at the Regular primary's width. |

### 4.2 Remove and the last-record warning

| ID | Requirement |
| --- | --- |
| RC-1 | **Remove.** A small ghost Button with the Unpromote glyph (the one the Evidence graph card uses) and the word "Unpromote", 6pt after each graph link's last ref (SS-5; agreed 2026-11-05: an icon alone at the far right read as detached). It removes that one record. It is on every link, the last one included. While a Remove opens or writes, every Remove in the section is disabled. |
| RC-2 | **The confirm.** The kit Confirm, the same confirm the Evidence graph shows for Unpromote (one shared confirm): title "Unpromote J. Weber?"; the message "This promotion is removed. The cards stay on the graph, and the handle stays."; under it the micro-caps header "Also unpromoted", the list of connection sentences ("J. Weber — daughter of — Heinrich Weber", "J. Weber — subject of — Baptism 1841") and "The other end of each stays promoted."; buttons "Unpromote" (irreversible tone) and "Keep promotion". With no connections, no list. Draw it over the Person page, Sources section. |
| RC-3 | **The last record (R-39).** When the preflight says the person is deleted, the confirm takes the danger tone and the message reads "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." (Maria Weber, frame 6). When it names more handles than the person, the message is the kit DeleteImpact line that names them all (today's Evidence graph wording for that case). The list of connections stays. Decide whether the Person page shows the person's name beside its ref in this message (the Evidence graph does not); say why. |
| RC-4 | **The check could not run.** The danger tone with "The check for whether this person is deleted could not run. Confirming may still delete this person. The cards stay on the graph." Unpromote stays enabled. |
| RC-5 | **Writing and failure.** While writing, the confirm's buttons show the kit's running state. A failed write keeps the confirm open with the error under the list; "Unpromote" stays available to retry. |
| RC-6 | **After a Remove that keeps the person.** The confirm closes; the graph link leaves, an emptied Source row leaves, and the section keeps its scroll position; focus moves to the next link's Remove, else the previous one's, else the first Source title. No Toast: the line leaving is the confirmation. Say if the board wants one. |
| RC-7 | **After the last record.** The confirm closes and the app lands on the Persons list; the deleted person is gone from the list and the sidebar count. Back does not return to the deleted person's page. Draw the Persons list frame with a note. |

### 4.3 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Sources section | Source rows in the primary column, graph links under each; the How the records connect band (empty, reserved) above them | Source rows in the primary column at the Regular measure; the secondary column reserved for How the records connect; centered primary-only until S13-28 (SS-W1) | Later, How the records connect moves from the band above the rows (Regular) to the secondary column (Wide) |
| Remove confirm | The kit Confirm over the page | The same Confirm, same size | Nothing |
| No records, loading, error | EmptyState, spinner, Callout in the primary column | Centered at the Regular primary's width (SS-W3) | Nothing |

---

## 5. Suggested frames

Each frame is the S13-D1 Person page pane only (no app sidebar), with the switcher on **Sources**, at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). The Wide frames sit beside their Regular frame.

Sample person: **Johann Weber** (man, PER-3HK9W), as on the Family, Timeline and Gallery boards. Seven records from five Sources, in the engine's order:

| Source (type, ref) | Evidence graph (ref) | Record (Subject label, ref) |
| --- | --- | --- |
| Hamburg census 1867 (Census, SRC-9TQ2M) | Household 214 (EVG-2K7PQ) | Johann Weber (CPR-7Q2KD) |
| | Household 388 (EVG-5MD8R) | J. Weber (CPR-2M8XF) |
| Hamburger Adressbuch 1861 (Directory, SRC-3HW6B) | Volume 3 (EVG-8TZ4N) | Weber, Joh., Tischler (CPR-9PA4D) |
| Hamburger Nachrichten, obituary 1880 (Newspaper, SRC-7KP2Q) | Issue 221 (EVG-1QR9D) | Johann Heinrich Weber (CPR-4TZ6K) |
| Konfirmandenbuch St. Ansgarii 1831 (Church register, SRC-2DN5X) | Register 2 (EVG-6WB3M) | Johann Weber (CPR-6HQ3R) |
| | Register 3 (EVG-4PL7T) | Joh. Weber (CPR-1XG8K) |
| Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten (Membership list, SRC-5RJ4W) | Page 7 (EVG-9CN2H) | Weber, J. H. (CPR-8WK2N) |

Every date cell shows "—" (Spike 15).

1. **Sources.** Johann's section: five Source rows (the long title cut at two lines), their seven graph links, the reserved How the records connect band above (dashed, labelled) (SS-1 – SS-5, SS-11).
2. **Row states.** One Source row and its links: resting; hover and focus on the title link; hover and focus on a graph link; a link whose Subject label is cut to fit (SS-5, SS-6).
3. **Large.** 50 records from 14 Sources: Hamburger Adressbuch 1855 – 1880 as one Source with 9 graphs ("Volume 1855" … "Volume 1879"), the census, the Konfirmandenbuch, the obituary, and ten more one- to six-record Sources (yearbooks, a guild register, a passenger list), with the SS-9 fold if the board folds (SS-9).
4. **States.** No records (SS-7), the first-load spinner, the error Callout, and "Refreshing" in the switcher row with the rows drawn (SS-8).
5. **Remove confirm.** Over frame 1: Remove on the Household 388 link, the confirm with two connection sentences under "Also unpromoted" (RC-1, RC-2).
6. **The last record.** A person with one record (Maria Weber, PER-8JD2L, one record from the obituary): the confirm in the danger tone with the deletes message; beside it the variant that deletes two other handles, and the could-not-check variant (RC-3, RC-4).
7. **Writing and failure.** The confirm running, then with a write error under the list (RC-5).
8. **After Remove.** Frame 1 with the Household 388 link gone and focus on the next Remove (RC-6); the Persons list after Maria's last record: Anna Weber, Heinrich Weber, Johann Weber, Karl Weber (Maria gone), the sidebar's Persons count 4 (was 5), with a note that Back does not return (RC-7).
9. **Keyboard.** The Tab order over frame 1, and the VoiceOver readings as notes (SS-10).
10. **Dark mode.** Frames 1, 3 and 5 in dark (SS-10).
11. **Wide.** Frame 1 at Wide centered primary-only (today), and frame 1 at Wide with the reserved secondary region for How the records connect, with the rule-5 note (SS-W1, SS-W2, SS-11); frame 4's states at Wide (SS-W3).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, header, switcher | Snowflake / component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `PVPageColumns`, `PVSegmentedControl` | As S13-D1. Sources fills the primary column; the secondary column is reserved for S13-D10 (SS-W1, SS-11). |
| Section host | Snowflake | Ship | `Features/Conclusions/ConclusionPageSectionHost.swift` | SS-7, SS-8, SS-W3. Spinner, error Callout and EmptyState, as Family and Timeline. |
| Sources section | Snowflake | **New** | `Features/Conclusions/` *(new file, named in the S13-16 plan)* | SS-1 – SS-11. Replaces the `.sources` EmptyState in `PersonDetailView.swift`. Its second call site: the Event and Place pages when they get sections (R-1). |
| Row columns and rhythm | Recipe | Ship | `DesignSystem/Recipes/ProvenanceSheet/PVProvenanceColumns.swift`, `PVProvenanceRecordTable.swift` | SS-1, SS-3. Borrow the provenance table's column spacing, row padding, hairlines and header line; the Sources row has its own columns (Citation · Type · Date · Ref), so it is not a provenance record row. |
| Citation (Source part) | Snowflake | Ship | `Platform/CitationLineDisplay.swift` (`Parts.title`, `titleIsRef`) | SS-2. The Source part of the citation line, not its detail. |
| Source title link | Recipe | Ship | the provenance record row's citation link (`PVProvenanceRecordRow.swift`) | SS-2. The same link treatment, focus ring and Return. |
| Source type | Snowflake | Ship | `Features/Catalog/SourceTypeMarkers.swift` `CatalogSourceTypePill` | SS-3. The Source page header's type pill. |
| Date, ref, count, graph ref, Subject ref | Component | Ship | caption and mono caption text styles | SS-3 – SS-5. "—" in muted text for the reserved date. |
| Graph link | Component | Ship | `PVButton` (ghost, `sm`) or the kit's small link style, as the record row's Source and Observation links | SS-5. Secondary: small, muted. |
| Remove | Component | Ship | `PVButton` (ghost, `sm`, icon `PVSymbol.unpromote`, label "Unpromote") | RC-1 as agreed (user, 2026-11-05): the glyph and the word, 6pt after the line's last ref. S13-17. |
| Fold | Snowflake | Ship | `Features/Conclusions/ConclusionShowMoreRow.swift` | SS-9, if the board folds. |
| How the records connect place | Snowflake | Ship | — (a dashed region on the frames only; nothing ships) | SS-11. S13-D10 designs it. |
| No records | Component | Ship | `PVEmptyState`, glyph `library` | SS-7. |
| Error | Component | Ship | `PVCallout` (danger) | SS-8. |
| Remove confirm | Component | Ship | `PVConfirm` (`item:` snapshot), its detail slot | RC-2 – RC-5. The Evidence graph's Unpromote confirm, lifted so both pages show one; no second confirm copy. |
| Connection sentences | Snowflake | Ship | `Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` | RC-2. The graph confirm's "Also unpromoted" lines. |
| Deletes-the-person line | Recipe | Ship | `PVDeleteImpactCopy` (`DesignSystem/Recipes/DeleteImpact/`) | RC-3, as the graph confirm. |
| Persons list | Snowflake | Ship | `Features/Conclusions/PersonsListView.swift` | RC-7. Unchanged; the landing place. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Subject cards or any record's property rows | User, 2026-11-05: too much information here. |
| An Observation's citation (Artifact, Citation, page) | This section cites Sources (§ 2). |
| A date editor or a placeholder date | The column is reserved; Spike 15 owns the date. |
| A drawn How the records connect | S13-D10; this board only reserves its place. |
| A second Unpromote confirm, or new confirm copy | User, 2026-11-05: the same confirm as the graph. |
| A Toast after Remove | RC-6, unless the board argues for one. |
| A filter, a search or a sort control | § 2.1. |
| A new kit primitive (a sources table) | One call site; borrow the provenance table's rhythm. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames; the column header answer (SS-1); the copy button answer (SS-2); the column order and widths (SS-3); where the count sits (SS-4); the graph link line and its glyph (SS-5); the empty copy (SS-7); the fold answer (SS-9); where How the records connect sits (SS-11); the last-record message's name (RC-3); the Toast answer (RC-6); and every other change from this brief with the user's answer.
2. Refresh the [S13-16 plan](../../prs/S13-16-sources-section.md) and the [S13-17 plan](../../prs/S13-17-remove-a-claim.md): copy the answers into their Decisions and **Design → macOS**, describe every frame in words, close their Open questions, run `review-plan` again, and set them to `Planned`.
3. S13-17 archives this brief under `design/archive/` and records it in [`../completed.md`](../../completed.md) (S13-16 lands first and leaves the brief open for S13-17).
