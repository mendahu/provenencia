# S13-D6 — Person page: the Sources section, Remove and the last-claim warning

**Kind:** Claude Design board (**enhancement** of the agreed S13-D1 Person page frame: it fills the Sources section, which S13-D1 drew as a placeholder place (PPF-10), and adds the Remove confirm over the page)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Sources section; the Remove confirm
**Implements later as:** PR **S13-16** (the Sources section) and PR **S13-17** (Remove, the last-claim warning, landing on the Persons list).
**Board:** not drawn yet.
**Depends on:** the agreed S13-D1 board (the frame, the header, the switcher; [archived brief](archive/S13-D1-person-page-frame.md), built by S13-10 [#634](https://github.com/mendahu/provenencia/pull/634)); the Evidence graph's Subject card (S9-D8, the card the canvas draws today); the Unpromote confirm and its last-member wording (S13-08, [#707](https://github.com/mendahu/provenencia/pull/707)); the agreed S13-D4 and S13-D5 boards (Page columns' full-width mode; [S13-D4](archive/S13-D4-timeline-section.md), [S13-D5](archive/S13-D5-gallery-and-viewer.md)).
**Related:** [`../person-detail.md`](../person-detail.md) R-33 – R-39 (and R-2, R-13b); [`../deployment-plan.md`](../deployment-plan.md) S13-16 and S13-17 scope entries, working rules "Sections are places", "One Unpromote, deletes through the registry", "A full-width section keeps its states in the band", "Claims are read per person from the graph cache, grouped by Source in Go"; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); slice 7's later board S13-D10 ([`../evidence-graph-comparison.md`](../evidence-graph-comparison.md)), which sits in or beside this section.
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** [`S13-D6-sources-section.review.md`](S13-D6-sources-section.review.md) — round 1 Fits with changes / Ready with changes, every finding Fixed

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

Extend the S13-D1 frames. The header, the switcher and the other four sections stay exactly as agreed (S13-D1, S13-D3, S13-D4, S13-D5); only the Sources section's content changes, and the Remove confirm is added over the page. Draw on the S13-D1 Person page frame, with the switcher on **Sources**. Do not start a second Person page. The Evidence graph canvas is not on this board: its Subject card keeps its look there, and its Unpromote confirm keeps its look (this board's confirm is the same confirm, RC-3).

---

## 1. Objective

A researcher opens a Person's Sources section and sees every record that makes up this person: one entry per **identity claim**, the claimed Subject drawn as the card the Evidence graph draws, grouped under the Source it was recorded from, the groups in Source title order. A Source opens its Source page; a claim opens its Evidence graph with that Subject selected and scrolled into view. On any claim the researcher can **Remove** it: the Subject is unpromoted from this person, exactly as Unpromote on the Evidence graph does. Removing the last claim deletes the person; the confirm says so first, and on proceeding the app lands on the Persons list.

**What changes.** The Sources place, today an EmptyState ("Sources arrives in a later version of Provenencia."), shows the Source groups. Each claim card gains the Person page's two actions: open its Evidence graph, and Remove. The Remove confirm opens over the Person page.

**What stays.** The header, the switcher, Vitals, Family, Timeline and Gallery. The Subject card's paint on the Evidence graph. The Evidence graph's own Unpromote confirm (title, copy, the "Also unpromoted" list, the danger tone when a handle is deleted): the Person page's confirm is the same one (user, 2026-11-05). One wording fix rides along: when only the subject's own handle is deleted, the graph's confirm today shows the generic DeleteImpact line; it takes the per-kind sentence approved for S13-08 on 2026-11-04 ("PER-3HK9W has no other records, so this person is deleted. The cards stay on the graph."), on both pages. The engine: the claims read (S13-01) and Unpromote with its preflight (S13-08) already exist; this board adds no data.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| An identity claim (R-33) | One accepted promotion of a Subject (a person as one Source records them) into this canonical person. The section lists every accepted claim; nothing unpromoted is drawn (R-2). |
| Source groups (R-34, R-36) | A Source can hold several Evidence graphs (a census has a graph per household page), so one Source can contribute several claims. Claims from one Source sit under one group. Groups are in Source title order; within a group, claims are in Evidence graph name order. The order comes from the engine; the board does not reorder. |
| What a group knows (R-35) | The Source's title (else its ref, "SRC-4KQ7M", in mono), its type ("Census", "Church register", "Obituary"; with the type's icon, as the Sources list shows it), its ref, and its claims. No date: the Source date was descoped (2026-10-31, Spike 15). |
| What a claim knows | The claimed Subject: its type ("Person"), working label, ref ("CPR-7Q2KD"), description, and its Observations (the property rows the Evidence graph card lists: name, sex, birth date, likeness, occupation and so on, each with its value). The Evidence graph it is on: name ("Household 214") and ref. No pins, no confidence, no argument: an identity claim is membership only (slice 0). |
| The Subject card (R-35) | The Evidence graph draws each Subject as a card 264pt wide with medium corner radius and a 1pt stroke in the kind's line color (a person's is the person tint; an uncited Subject's stroke is dashed). Top to bottom: a header (a 28pt rounded chip in the kind tint holding the person mark, beside a mono 10pt line "Person · CPR-7Q2KD" over the Subject's name in 15pt semibold; on the graph a pen and a trash icon at the header's end); the description in italic 12.5pt, up to three lines; a stack of property rows separated by 1pt hairlines in the kind line color, each row a small-caps property label over its value (a likeness row shows its thumbnail; competing values of one property carry a warning-colored bracket at the leading edge); an Add property row; and a 36pt footer band in the kind tint holding the handle's ref (mono) and name, with a 1pt hairline and a 36pt square holding the Unpromote icon at its end. Draw the card from this description; the Evidence graph board (S9-D8) is the reference if the project has it. On the Person page the card is **read-only**: header and property rows as the graph paints them, no pen, trash or Add property, and its footer is this page's (SS-4). |
| Open the Source (R-37) | The Source page. |
| Open the claim's graph (user, 2026-11-05) | The Evidence graph the claim is on, with its Subject's card selected and scrolled into view. Back returns to the Sources section. |
| Remove (R-38) | Unpromote this Subject from the person. It takes back the Subject's claim and the claims of its relationships, participations and locations that have this Subject as an end; the cards stay on the Evidence graph. The confirm is the Evidence graph's Unpromote confirm (user, 2026-11-05): title "Unpromote Anna Weber?", the message, the "Also unpromoted" list of those connections with "The other end of each stays promoted.", "Unpromote" and "Keep promotion". |
| The last claim (R-39) | When the claim is the person's last, Remove deletes the person. The engine says so before the write (its preflight), and the confirm turns to the danger tone with "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." When the check cannot run, the confirm says the person may be deleted. On proceeding, the app lands on the **Persons list**, and Back does not return to the deleted page. |
| Another handle can go too | Unpromoting a Subject can empty a relationship or participation handle whose last member was one of its connections; the preflight names those too, and the confirm's message names them all with the kit DeleteImpact line. |
| The read and refresh | The section reads its own query (the person's claims), with its own spinner on the first load and "Refreshing" at the end of the switcher row on a reload (S13-D1 PPF-8). A Remove that keeps the person refreshes the section: the removed card leaves, an emptied group leaves. |
| Sizes | Typical: 3 – 20 claims from 2 – 12 Sources. Large: 50 claims (a person through decades of yearbooks and censuses), up to 9 from one Source. A card has 0 – 30 property rows (typical 4 – 8). Source titles can be long ("Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten"). |

### 2.1 What this board is not

- The Source date, or sorting groups by date (descoped 2026-10-31, Spike 15).
- How the person's records connect (slice 7, board S13-D10): no links between cards, no comparison, no grouping by connection. S13-D10 decides later whether it sits inside this section or beside it; this board leaves room only by keeping the section's own layout simple.
- Editing a Subject, its Observations or its claim from this page: no edit, delete, Add property, or Promote. Those stay on the Evidence graph.
- A provenance control on a card (R-13b).
- A claim review, pins, a confidence or an argument (descoped 2026-10-29).
- Rejecting a claim, or marking it provisional: Remove only unpromotes (R-38).
- A bulk Remove, a multi-select, a filter or a search.
- A change to the Evidence graph canvas or its card; its confirm changes only by the last-member sentence (§ 1).

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-16** | SS-1 … SS-10 and SS-W1 … SS-W3: the Sources section, its groups and their headings, the read-only claim card, Open Source, Open graph, the empty, loading, error and refreshing states, keyboard and VoiceOver, light and dark, at Regular and Wide. The footer's end square stays empty until S13-17 builds Remove into it. |
| Ships in **S13-17** | RC-1 … RC-7: Remove on each claim card, the confirm (the Evidence graph's confirm, shared), the last-claim warning, the write's pending and failure states, landing on the Persons list, and the section after a Remove that keeps the person. |
| Later boards | S13-D10 (how the records connect) places its view in or beside this section. |
| Stays out | Any change to the header, the switcher, the other sections, the Evidence graph canvas, or its confirm beyond the shared last-member sentence (§ 1). |

---

## 4. Requirements

### 4.1 Sources section

| ID | Requirement |
| --- | --- |
| SS-1 | **The groups.** One group per Source, in the engine's order (Source title). Each group is its heading line (SS-2) over its claim cards (SS-3). Decide how groups are set apart (spacing only, a hairline, or each group on a sunken Card); say why. |
| SS-2 | **The group heading.** The Source title (else its ref in mono), its type with the type's icon, and the number of claims ("3 records"). The title opens the Source page (R-37) and is the same title link the Timeline section uses for an event's title: the heading-style title, two lines then cut at its tail, the kit focus ring, Return opens it, the full title in the tooltip. The type is the **Source type pill** the Source page's header shows (icon and label). The count is a muted caption after the pill. |
| SS-3 | **The claim cards.** Each claim as the Evidence graph's Subject card, at its graph width (264pt), the same header and property rows the graph paints, read-only (§ 2). Cards in a group wrap in rows, left to right, in the engine's order. Decide the gap; cards in one row align at their top. |
| SS-4 | **The card's footer.** The graph's footer geometry: where the graph shows the handle row, the Evidence graph the claim is on ("Household 214", its ref in mono), opening that graph with the Subject selected and in view; after the 1pt hairline, in the 36pt square where the graph puts Unpromote, the Remove control (an IconButton, the same glyph, tooltip "Unpromote"; S13-17). Decide whether the graph name is the whole handle part's click or a link in it; say why. When a Source has one Evidence graph, the footer still names it. |
| SS-5 | **Hover and focus.** A card does not lift as a whole (it is not one button): its two controls (open graph, Remove) take hover and the kit focus ring. Draw both states. |
| SS-6 | **No claims.** A person with no accepted claim is deleted by the engine (R-39), so this state shows only when the read races a delete. The kit EmptyState with the Sources glyph (library): "No records" · "This person has no promoted records." Decide the copy; keep it short. |
| SS-7 | **Loading, error, refreshing.** First load: the section's spinner, as Family and Timeline. A failed read: the danger Callout "Sources could not be read" with the error (the Family and Timeline wording). A reload keeps the cards drawn and shows "Refreshing" at the end of the switcher row (PPF-8). |
| SS-8 | **Large.** 50 claims from 14 Sources, one with 9: every card is shown, no fold. Decide whether a long group folds behind "Show N more" (the Family and Timeline row) and at what count; recommended: no fold (one card is one record the researcher may want to remove). |
| SS-9 | **Keyboard and VoiceOver.** Tab goes through each group's title link, then each card's open-graph control and Remove, in reading order. VoiceOver reads a group heading as a heading: "Hamburg census 1867, Census, 2 records"; a card as one element with the graph card's own label plus its graph: "Person, CPR-7Q2KD, Anna Weber, cited, on Household 214" (the graph card does not read its rows one by one, and neither does this page); the controls as buttons: "Open Household 214", "Unpromote Anna Weber". |
| SS-10 | **Dark mode.** Frames 1, 4 and 6 in dark: headings, Chips, cards with their kind tint, the footer. |
| SS-W1 | **Sources at Wide.** Decide between two and say why, from rule 1 and rule 3 (the user leaves it to the board, 2026-11-05): (A) full-width mode, as Timeline and Gallery: each group's cards wrap across the pane between the gutters and add columns; headings keep their measure; or (B) primary-only and centered at the Regular measure, leaving the secondary column for S13-D10. Draw the chosen one, and the other as an alternative frame marked as such. A secondary column of other content is not an option on this board. |
| SS-W2 | **Crossing 1600 keeps the place (rule 5).** Resizing keeps the section, the scroll position and the focused control. A note on the Wide frame. |
| SS-W3 | **States at Wide.** The EmptyState, the spinner and the Callout stay centered at the Regular primary's width (working rule "A full-width section keeps its states in the band"). |

### 4.2 Remove and the last-claim warning

| ID | Requirement |
| --- | --- |
| RC-1 | **Remove.** The Remove IconButton at the end of a card's footer (SS-4) opens the confirm for that claim's Subject. It is on every card, the last one included. While a Remove is writing, every Remove in the section is disabled. |
| RC-2 | **The confirm.** The kit Confirm, the same confirm the Evidence graph shows for Unpromote (one shared confirm): title "Unpromote Anna Weber?"; the message "This promotion is removed. The cards stay on the graph, and the handle stays."; under it the micro-caps header "Also unpromoted", the list of connection sentences ("Anna Weber — daughter of — Johann Weber", "Anna Weber — subject of — Baptism 1841") and "The other end of each stays promoted."; buttons "Unpromote" (irreversible tone) and "Keep promotion". With no connections, no list. Draw it over the Person page, Sources section. |
| RC-3 | **The last claim (R-39).** When the preflight says the person is deleted, the confirm takes the danger tone and the message reads "PER-8JD2L has no other records, so this person is deleted. The cards stay on the graph." (Maria Weber, frame 6) When it names more handles than the person, the message is the kit DeleteImpact line that names them all (today's Evidence graph wording for that case). The list of connections stays. Decide whether the Person page shows the person's name beside its ref in this message (the Evidence graph does not); say why. |
| RC-4 | **The check could not run.** The danger tone with "The check for whether this person is deleted could not run. Confirming may still delete this person. The cards stay on the graph." Unpromote stays enabled. |
| RC-5 | **Writing and failure.** While writing, the confirm's buttons show the kit's running state. A failed write keeps the confirm open with the error under the list (the kit Confirm's error line); "Unpromote" stays available to retry. |
| RC-6 | **After a Remove that keeps the person.** The confirm closes; the card leaves its group, an emptied group leaves, and the section stays on the same scroll position; focus moves to the next card's Remove, else the previous one's, else the section. No Toast: the card leaving is the confirmation. Say if the board wants one. |
| RC-7 | **After the last claim.** The confirm closes and the app lands on the Persons list, with the list's own selection and scroll; the deleted person is gone from the list and the sidebar count. Back does not return to the deleted person's page. Draw the Persons list frame with a note. |

### 4.3 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Sources section | Groups stacked in the primary column; each group's 264pt cards wrap, two or three a row | The board's choice from SS-W1 (A or B) | (A) the primary widens and cards add columns; (B) nothing moves |
| Remove confirm | The kit Confirm over the page | The same Confirm, same size | Nothing |
| No claims, loading, error | EmptyState, spinner, Callout in the primary column | Centered at the Regular primary's width (SS-W3) | Nothing |

---

## 5. Suggested frames

Each frame is the S13-D1 Person page pane only (no app sidebar), with the switcher on **Sources**, at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). The Wide frames sit beside their Regular frame.

Sample person: **Johann Weber** (man, PER-3HK9W), as on the Family, Timeline and Gallery boards. Seven claims from five Sources, in the engine's order:

| Source (type) | Evidence graph | Subject (ref) | Property rows |
| --- | --- | --- | --- |
| Hamburg census 1867 (Census) | Household 214 | Johann Weber (CPR-7Q2KD) | name, sex, age 32, occupation "Tischler", birthplace "Hamburg" |
| Hamburg census 1867 (Census) | Household 388 | J. Weber (CPR-2M8XF) | name, age 33, occupation "Tischlermeister" |
| Hamburger Adressbuch 1861 (Directory) | Volume 3 | Weber, Joh., Tischler (CPR-9PA4D) | name, occupation, residence "Steinstraße 12", likeness |
| Hamburger Nachrichten, obituary 1880 (Newspaper) | Issue 221 | Johann Heinrich Weber (CPR-4TZ6K) | name, death date 3 Mar 1880, age 45, residence |
| Konfirmandenbuch St. Ansgarii 1831 (Church register) | Register 2 | Johann Weber (CPR-6HQ3R) | name, birth date 12 Mar 1817, father "Heinrich Weber" |
| Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten (Membership list) | Page 7 | Weber, J. H. (CPR-8WK2N) | name, occupation |
| SRC-4KQ7M (a Source with no title; Letter) | Graph 1 | Johann (CPR-3NB5V) | name |

1. **Sources.** Johann's Sources section with the seven claims in five groups (the long title cut at two lines, the untitled Source's ref in mono), the census group's two cards side by side, each footer naming its graph with Remove at its end (SS-1 – SS-4).
2. **Card states.** One card: resting; hover and focus on the open-graph control; hover and focus on Remove; a card with 12 property rows (Johann Weber, Hamburg census 1867 full entry: name, sex, age 32, birth date 1835, birthplace Hamburg, occupation Tischler, religion lutherisch, marital status verheiratet, residence Steinstraße 12, household position Haushaltsvorstand, nationality Hamburg, literacy ja) beside one with 1 (Johann, the letter), aligned at the top (SS-3 – SS-5).
3. **Large.** 50 claims from 14 Sources, one with 9 (SS-8): Hamburger Adressbuch 1855 – 1880 as one Source with 9 volumes (graphs "Volume 1855" … "Volume 1879"), the two censuses, the Konfirmandenbuch, the obituary, and nine more one- to six-record Sources (yearbooks, a guild register, a passenger list); names and rows as in the sample table.
4. **States.** No claims (SS-6), the first-load spinner, the error Callout, and "Refreshing" in the switcher row with the cards drawn (SS-7).
5. **Remove confirm.** Over frame 1: Remove on the Household 388 card, the confirm with two connection sentences in "Also unpromoted" (RC-1, RC-2).
6. **The last claim.** A person with one claim (Maria Weber, PER-8JD2L, one claim from the obituary): the confirm in the danger tone with the deletes message; beside it the variant that deletes two other handles, and the could-not-check variant (RC-3, RC-4).
7. **Writing and failure.** The confirm running, then with a write error under the list (RC-5).
8. **After Remove.** Frame 1 with Household 388's card gone and focus on the next Remove (RC-6); and the Persons list after Maria's last claim: Anna Weber, Heinrich Weber, Johann Weber, Karl Weber (Maria gone), the sidebar's Persons count 4 (was 5), with a note that Back does not return (RC-7).
9. **Keyboard.** The Tab order over frame 1, and the VoiceOver readings as notes (SS-9).
10. **Dark mode.** Frames 1, 4 and 6 in dark (SS-10).
11. **Wide.** Frame 1 and frame 3 at Wide in the SS-W1 choice, with the rule-5 note (SS-W1, SS-W2); the alternative SS-W1 frame marked as such; frame 4's states at Wide (SS-W3).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, header, switcher | Snowflake / component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `PVPageColumns`, `PVSegmentedControl` | As S13-D1. Sources fills the primary column; at Wide as SS-W1 decides (full-width mode is `ConclusionPageSection.fillsWidth`, as Timeline and Gallery). |
| Section host | Snowflake | Ship | `Features/Conclusions/ConclusionPageSectionHost.swift` | SS-6, SS-7, SS-W3. The spinner, the error Callout and the EmptyState of a section that reads its own query, as Family and Timeline. |
| Sources section | Snowflake | **New** | `Features/Conclusions/` *(new file, named in the S13-16 plan)* | SS-1 – SS-9. Replaces the `.sources` EmptyState in `PersonDetailView.swift`. Its second call site: the Event and Place pages when they get sections (R-1). |
| Group heading | Snowflake | Ship | the Timeline's title link: `PVTitleText` (`.heading`, two lines) in a Button with `PVFocusRingButtonStyle` and `pvActivatesOnReturn` (`Features/Conclusions/PersonTimelineSection.swift`) | SS-2. The heading row is that link, the type pill and the count; it carries the heading trait. |
| Source type | Snowflake | Ship | `Features/Catalog/SourceTypeMarkers.swift` `CatalogSourceTypePill` | SS-2. The Source page header's type pill (icon and label); one look for a Source's type. Not a Chip (a Chip is a button). |
| Subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceCardView.swift`, `EvidenceSubjectCard.swift` | SS-3. The Evidence graph's own paint, read-only: header and property rows, no edit or delete icon, no Add property, this page's footer. One card, two hosts; not a second card drawn after it. |
| Card footer: graph link | Component | Ship | `PVButton` (ghost, `sm`) over the footer's handle part (`EvidenceSubjectCard.promotedFooterParts`) | SS-4. The Evidence graph's name, its ref in mono. |
| Card footer: Remove | Component | Ship | `PVIconButton` with `PVSymbol.unpromote` in the footer's 36pt slot (`promotedFooterParts`) | SS-4, RC-1. Tooltip and label "Unpromote". S13-17. |
| Card wrap | Component | Ship | `PVFlowLayout` | SS-3. Cards wrap at their width; Wide adds columns (SS-W1 A). |
| No claims | Component | Ship | `PVEmptyState`, glyph `library` | SS-6. |
| Error | Component | Ship | `PVCallout` (danger) | SS-7. |
| Remove confirm | Component | Ship | `PVConfirm` (`item:` snapshot), its detail slot | RC-2 – RC-5. The Evidence graph's Unpromote confirm, lifted so both pages show one confirm; no second confirm copy. |
| Connection sentences | Snowflake | Ship | `Features/Catalog/Connect/EvidenceBridgeEdgeSummary.swift` | RC-2. The graph confirm's "Also unpromoted" lines. |
| Deletes-the-person line | Recipe | Ship | `PVDeleteImpactCopy` (`DesignSystem/Recipes/DeleteImpact/`) | RC-3. The preflight's wording, as the graph confirm. |
| Persons list | Snowflake | Ship | `Features/Conclusions/PersonsListView.swift` | RC-7. Unchanged; the landing place. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A second Subject card (a SwiftUI look-alike, a compact summary card) | R-35: the Evidence graph's Subject card. One paint, two hosts. |
| Edit, delete, Add property or Promote on the page's card | Those stay on the Evidence graph (§ 2.1). |
| A second Unpromote confirm, or new confirm copy for the Person page | User, 2026-11-05: the same confirm as the graph. |
| A Toast after Remove | RC-6: the card leaving confirms it, unless the board argues for one. |
| A fold, a filter, a search or a sort control | § 2.1, SS-8. |
| Links or comparison between cards | Slice 7, S13-D10. |
| A new kit primitive (a group list, a claim card) | One call site. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames; how groups are set apart (SS-1); the heading's title control and type Chip (SS-2); the card gap (SS-3); the footer's graph control and Remove (SS-4); the empty copy (SS-6); the fold answer (SS-8); the Wide choice (SS-W1); the last-claim message's name (RC-3); the Toast answer (RC-6); and every other change from this brief with the user's answer.
2. Refresh the [S13-16 plan](../prs/S13-16-sources-section.md) and the [S13-17 plan](../prs/S13-17-remove-a-claim.md): copy the answers into their Decisions and **Design → macOS**, describe every frame in words, close their Open questions, run `review-plan` again, and set them to `Planned`.
3. S13-17 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md) (S13-16 lands first and leaves the brief open for S13-17).
