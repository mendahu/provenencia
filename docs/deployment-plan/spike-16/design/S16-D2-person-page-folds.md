# S16-D2 — Person page folds: a page at a time, focus on what it reveals, and a folded record's highlight

**Kind:** Claude Design board (**enhancement** of the Spike 13 Person page boards, S13-D3 Family, S13-D4 Timeline, S13-D6 Sources and S13-D10 record links: states are added to the existing frames)
**Spike:** Provenencia Spike 16 (odds and ends), batch 1
**View:** the Person page, in its three sections that fold: Family (the relatives' tiles), Timeline (the event rows and each event's participant cards) and Sources (the Source rows and their record lines, beside the record links diagram)
**Implements later as:** PR **S16-07** (One paged fold that moves focus)
**Board:** agreed 2026-11-09 — [Claude Design — Person Detail](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ), turn "S16-D2 · Folds" on the pages *Person Detail - Family* (F5a, F5b), *Person Detail - Timeline* (F6a, F6b, F7a, F7b, F8, F9) and *Person Detail - Sources* (F1a, F1b, F1 narrow pane, F2a, F2b, F3a, F3b, F4). FH-1: the record line's highlight (`surfaceSelected`, `PVRadius.sm`) behind the whole Show more row. FH-4: a pointer click leaves focus where it is; Family and Sources scroll the revealed item just into view
**Depends on:** the agreed S13-D3, S13-D4, S13-D6 and S13-D10 boards (Family shipped in S13-12, [#676](https://github.com/mendahu/provenencia/pull/676), and Timeline in S13-13, [#691](https://github.com/mendahu/provenencia/pull/691); Sources and record links in S13-16, [#722](https://github.com/mendahu/provenencia/pull/722), and S13-28, [#744](https://github.com/mendahu/provenencia/pull/744))
**Related:** audit P-15 and P-14 ([`../person-page-audit.md`](../person-page-audit.md)); working rule "One paged fold that moves focus" in [`../deployment-plan.md`](../deployment-plan.md#working-rules); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S16-07-paged-fold-focus.md`](../prs/S16-07-paged-fold-focus.md)
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
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths, unless the section opts into Page columns' full-width mode (`primaryFill`: at Wide, with no secondary, the primary fills the pane between the gutters and the header keeps its band); its text still keeps its measure. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Draw on the Person page board as a new turn "S16-D2 · Folds", starting from the S13-D3 Family frames, the S13-D4 Timeline frames, the S13-D6 Sources frames and the S13-D10 frames with the record links diagram beside the rows (Regular and Wide). Add states and change one label. Do not redraw the page header, the section switcher, the Family tree's lines and captions, the Timeline's lifeline, rows and cards, the Source rows' columns, the record lines, the Unpromote buttons, the diagram or its column.

---

## 1. Objective

Three sections of the Person page fold their long lists behind a "Show N more" row: Family folds each list of relatives past 30 tiles; Timeline folds the dated and the undated events past 50 rows, and each event's other people past 5 cards at Regular and 12 at Wide; Sources folds as below. Today a Show more reveals the whole list at once and says how many are folded in all. The Timeline and Sources move the keyboard to the first item it revealed; Family drops it, so a keyboard user must Tab from the top of the page again (P-14).

The Sources section lists the Sources a person's records come from, one row each, with one line per record under it. Beside the rows, the record links diagram draws one node per Evidence graph. A node and the lines of its graph highlight each other: hovering or focusing a node gives that graph's lines the selected-surface background, and hovering a line lights its node.

A Source shows five record lines and folds the rest behind "Show N more" (six lines show whole). The section shows 50 Sources and folds the rest behind a "Show N more" row under the last one. When the hovered node's graph has lines only behind a fold, nothing in the list lights up today, so the researcher cannot tell which Source the node belongs to (P-15). A parish register filed one graph per entry reaches this with its sixth entry.

**What changes.**
- Every Show more on the Person page reveals at most 50 more (one page) and says how many the next click reveals: "Show 50 more" while more than 50 are folded, then "Show 15 more" for the last 15.
- After a Show more, the keyboard is on the first item it revealed, in Family as in Timeline and Sources.
- Crossing the Regular/Wide line keeps the keyboard; a Timeline card that folds away at Regular gives it to its event's title.
- When the highlighted graph has no line shown but has a line behind a Source's fold, that Source's Show more row takes the highlight. When the graph's Source is itself behind the section's Source fold, the section's Show more row takes it.

**What stays.** The highlight's look on a record line, the tiles, rows, cards and lines and their layout, the first fold limits (Family 30, Timeline 50 events and 5 / 12 cards, five lines per Source, 50 Sources), the "Show %lld more" wording, the lifeline (it still draws every event's point), the diagram, and the Wide arrangements S13-D3, S13-D4 and S13-D10 agreed.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A record line | One of the person's records on one Evidence graph of the Source: the graph's name as a link, the graph ref, the record's name and ref, Unpromote. |
| A graph has several lines | A person can have up to ten records on one graph, so a graph's lines can sit both before and behind a fold. A graph with any line shown highlights its shown lines only. |
| The highlight | One graph at a time: the one the pointer is on (a node or a line), else the one whose node or line has the keyboard. |
| A page | 50 more items per Show more click, in every section of the Person page. The fold's first limit is unchanged. |
| Family's lists | Parent sets, siblings (the person among them), partners, and each partner's children. Go returns at most 100 relatives per role, so a list has at most two pages past its first 30. The person's own tile is not a button and never takes the keyboard. |
| Timeline's folds | One fold for the dated events before the death marker, one for the undated group, one per event for its other people. A lifeline point on a folded row opens the dated fold through that row's page and scrolls to it. |
| Long lists | A Source can hold hundreds of a person's records (one graph per register entry); a person can have records on more than 50 Sources. |

### 2.1 What this board is not

- Hovering a Show more row does not highlight a node: the row stands for many graphs.
- The highlight is not spoken. A line's highlight is not spoken today either; VoiceOver reads the node and the line, not their link.
- The Gallery and the Vitals section: they do not fold. The Gallery keeps the keyboard across the Regular/Wide line as it does today (no frame).
- A focus ring of a new kind: the ring is each tile's, row title's, card's or link's own.
- A reveal-on-hover, an auto-unfold, or a count of the highlighted graph's folded lines.

---

## 3. Implementation gate

| S16-07 ships | Not S16-07 |
| --- | --- |
| FH-1 – FH-5, FH-W1, FH-W2 | Back restoring a section's folds (P-19, [idea](../../../ideas/person-page-followups.md)); the Sources rows' narrow arrangement ([wide layout idea](../../../ideas/wide-layout.md)) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| FH-1 | When the highlighted graph has no line shown in a Source row and at least one line behind that row's fold, the row's Show more takes the highlight. **The board decides** its look; the recommendation is the record line's own highlight (the selected-surface background with the line's small corner radius, behind the whole Show more row). The highlight goes when the node is no longer highlighted. |
| FH-2 | When the highlighted graph's Source is behind the section's Source fold, the section's Show more row (under the last shown Source) takes the same highlight as FH-1. |
| FH-3 | Every Show more says how many the next click reveals: the smaller of 50 and the number folded. A Source with 120 lines shows five and "Show 50 more"; one click shows 55 and "Show 50 more"; a second shows 105 and "Show 15 more"; a third shows all 120 and no Show more. A click never leaves a single line behind a Source's fold (six lines show whole today): a Source with 56 lines shows five and "Show 51 more". |
| FH-4 | After a Show more is pressed with the keyboard, the first item it revealed has the keyboard and its focus ring: in Family the first revealed relative's tile (never the person's own tile); in the Timeline the first revealed row's title, or an event's first revealed card; in Sources the first revealed record line's graph link, or the first revealed Source's first graph link. A Timeline row fold also scrolls the revealed row into view. |
| FH-5 | A click on a lifeline point whose row is folded opens the dated fold through that row's page only (the rows after that page stay folded behind "Show 50 more") and scrolls the row to the top. |
| FH-W1 | At Wide the diagram stays in the pinned column beside the rows, so the highlighted node and the highlighted Show more are on screen together. At Regular, when the pane is narrower than the rows' minimum, the diagram is under the rows, as today; the highlight works the same. Nothing moves between widths. |
| FH-W2 | The page is 50 at both widths. The Timeline shows 12 cards per event at Wide and 5 at Regular before its fold, as today; a page adds 50 to either. Crossing the line keeps the keyboard where it was; when the focused card is folded away at Regular (the 9th of 12), its event's title takes the keyboard. Family and Timeline keep the Wide arrangements S13-D3 and S13-D4 agreed. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Person page, Family section | The tree in its two columns, as S13-D3 agreed | As S13-D3 agreed | Nothing; this board adds the page count and the focus landing |
| Person page, Timeline section | The lifeline and the rows, 5 cards per event before the fold | Full width: the lifeline and the rows, 12 cards per event before the fold | The card fold's first limit only, as today (FH-W2) |
| Person page, Sources section | The rows, with the record links diagram in a pinned column beside them, or under them when the rows would be narrower than their minimum | Full width (Page columns' `primaryFill`): the rows fill the pane, the diagram in its pinned column beside them | Nothing; this board adds a state to the Show more rows |

---

## 5. Suggested frames

Frames with an a / b pair are drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440); the others as they say.

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | FH-1, FH-W1 | A person with nine records on "Hamburg parish register 1820–1840", one graph per entry, and records on three other Sources. The register's row shows five lines and "Show 4 more". The pointer is on the node of the seventh entry, "Entry 1834/112": the node and its arcs are highlighted, the register's Show more has the highlight, no line has it |
| 2a / 2b | FH-1 | Frame 1 with the pointer on the node of the third entry: its shown line has the highlight, the Show more does not |
| 3a / 3b | FH-2, FH-W1 | A person with records on 60 Sources: 50 rows and "Show 10 more". The pointer is on a node whose Source is the 55th: the section's Show more has the highlight |
| 4 | FH-3 | Regular only: a Source with 120 lines in three states side by side: folded ("Show 50 more" after five lines), after one click ("Show 50 more" after 55), after two ("Show 15 more" after 105) |
| 5a / 5b | FH-3, FH-4 | Family: a person with one partner and 100 children. Folded: 30 children and "Show 50 more". After a keyboard press: 80 children, the 31st child's tile with its focus ring, "Show 20 more" |
| 6a / 6b | FH-3, FH-4 | Timeline: a person with 2,000 dated events. Folded: 50 rows and "Show 50 more". After a keyboard press: the 51st row scrolled to the top with its title's focus ring, "Show 50 more" after the 100th row |
| 7a / 7b | FH-3, FH-4, FH-W2 | Timeline: one event with 500 other people. Folded: 5 cards (Regular) or 12 (Wide) and "Show 50 more". After a keyboard press: 55 or 62 cards, the 6th or 13th card with its focus ring |
| 8 | FH-W2 | Regular and Wide side by side: at Wide the 9th card of an event has the keyboard; at Regular that card is folded away and the event's title has the focus ring |
| 9 | FH-5 | Regular only: a lifeline click on the point of the 180th event: rows 1–200 shown, the 180th at the top, "Show 50 more" after the 200th |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Show more row | Snowflake | Extend | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` `ConclusionShowMoreRow` | Gains a highlighted state (FH-1, FH-2) and the page count (FH-3). Still a small ghost Button; not a link, not a Badge |
| Record line highlight | Snowflake | Ship | `macos/App/Features/Conclusions/PersonSourcesSection.swift` `PersonSourcesLineView` | `PVColor.surfaceSelected` behind the line, `PVRadius.sm` corners. FH-1's recommended look reuses these tokens; no new token |
| Source row | Snowflake | Ship | `PersonSourcesSection.swift` `PersonSourcesRowView` | Unchanged |
| Family tile | Recipe | Ship | `macos/App/DesignSystem/Recipes/PersonCard/PVPersonCard.swift` on `PVTileCard` | Its own focus ring (`PVFocusRingButtonStyle`) shows where FH-4 lands; no new ring |
| Timeline row title and participant card | Snowflake + recipe | Ship | `macos/App/Features/Conclusions/PersonTimelineSection.swift`; `PVPersonCard` `.compact` | Their own focus rings, as today |
| Lifeline | Snowflake | Ship | `PersonTimelineSection.swift` `PersonTimelineLine` | Unchanged; every point still drawn (FH-5 changes what a click opens, not the line) |
| Record links diagram | Snowflake | Ship | `macos/App/Features/Conclusions/PersonRecordLinksColumn.swift` | Unchanged; its node highlight drives FH-1 and FH-2 |
| Button | Component | Ship | `macos/App/DesignSystem/Components/Button/` | The Show more's Button, ghost, small, as today |

**Non-goals.** No new kit primitive, no new focus ring, no new color token, no new copy key: "Show %lld more" stays and only its number changes. No count of the highlighted graph's folded lines on the row.

---

## 7. Handoff

When the board is agreed: record the frames and the FH-1 look on the **Board** line here, in [`README.md`](README.md) and on the plan's **Board** line; the plan copies FH-1 – FH-5, FH-W1 and FH-W2 into its Scope and moves from `Blocked — on board S16-D2` to `Planned`. When S16-07 lands, the brief moves to the spike's design archive.
