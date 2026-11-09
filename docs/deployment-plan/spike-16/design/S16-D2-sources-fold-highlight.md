# S16-D2 — Person page Sources: a folded record's highlight and paged Show more

**Kind:** Claude Design board (**enhancement** of the Spike 13 Person page boards, S13-D6 Sources section and S13-D10 record links: one state is added to the existing Sources frames)
**Spike:** Provenencia Spike 16 (odds and ends), batch 1
**View:** the Person page's Sources section: the Source rows' Show more rows, beside the record links diagram
**Implements later as:** PR **S16-07** (One paged fold that moves focus)
**Board:** — (not agreed yet)
**Depends on:** the agreed S13-D6 and S13-D10 boards (shipped in S13-16, [#722](https://github.com/mendahu/provenencia/pull/722), and S13-28, [#744](https://github.com/mendahu/provenencia/pull/744))
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

Draw on the Person page board as a new turn "S16-D2 · Folded highlight", starting from the S13-D6 Sources frames and the S13-D10 frames with the record links diagram beside the rows (Regular and Wide). Add one state and change one label. Do not redraw the header line, the Source rows' columns, the record lines, the Unpromote buttons, the diagram or its column.

---

## 1. Objective

The Sources section lists the Sources a person's records come from, one row each, with one line per record under it. Beside the rows, the record links diagram draws one node per Evidence graph. A node and the lines of its graph highlight each other: hovering or focusing a node gives that graph's lines the selected-surface background, and hovering a line lights its node.

A Source shows five record lines and folds the rest behind "Show N more" (six lines show whole). The section shows 50 Sources and folds the rest behind a "Show N more" row under the last one. When the hovered node's graph has lines only behind a fold, nothing in the list lights up today, so the researcher cannot tell which Source the node belongs to (P-15). A parish register filed one graph per entry reaches this with its sixth entry.

**What changes.**
- When the highlighted graph has no line shown but has a line behind a Source's fold, that Source's Show more row takes the highlight. When the graph's Source is itself behind the section's Source fold, the section's Show more row takes it.
- Every Show more on the Person page reveals at most 50 more (one page) and says how many the next click reveals: "Show 50 more" while more than 50 are folded, then "Show 15 more" for the last 15. Today it says how many are folded in all and reveals them all at once.

**What stays.** The highlight's look on a record line, the line and row layout, the first fold limits (five lines per Source, 50 Sources), the "Show %lld more" wording, the diagram, and the Wide arrangement S13-D10 agreed.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A record line | One of the person's records on one Evidence graph of the Source: the graph's name as a link, the graph ref, the record's name and ref, Unpromote. |
| A graph has several lines | A person can have up to ten records on one graph, so a graph's lines can sit both before and behind a fold. A graph with any line shown highlights its shown lines only. |
| The highlight | One graph at a time: the one the pointer is on (a node or a line), else the one whose node or line has the keyboard. |
| A page | 50 more items per Show more click, in every section of the Person page. The fold's first limit is unchanged. |
| Long lists | A Source can hold hundreds of a person's records (one graph per register entry); a person can have records on more than 50 Sources. |

### 2.1 What this board is not

- Hovering a Show more row does not highlight a node: the row stands for many graphs.
- The highlight is not spoken. A line's highlight is not spoken today either; VoiceOver reads the node and the line, not their link.
- Family, Timeline and Gallery. Their Show more rows change only their number (FH-3), with the same control and wording, so they are not drawn here.
- A reveal-on-hover, an auto-unfold, or a count of the highlighted graph's folded lines.

---

## 3. Implementation gate

| S16-07 ships | Not S16-07 |
| --- | --- |
| FH-1 – FH-3, FH-W1 | Back restoring a section's folds (P-19, [idea](../../../ideas/person-page-followups.md)); the Sources rows' narrow arrangement ([wide layout idea](../../../ideas/wide-layout.md)) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| FH-1 | When the highlighted graph has no line shown in a Source row and at least one line behind that row's fold, the row's Show more takes the highlight. **The board decides** its look; the recommendation is the record line's own highlight (the selected-surface background with the line's small corner radius, behind the whole Show more row). The highlight goes when the node is no longer highlighted. |
| FH-2 | When the highlighted graph's Source is behind the section's Source fold, the section's Show more row (under the last shown Source) takes the same highlight as FH-1. |
| FH-3 | Every Show more says how many the next click reveals: the smaller of 50 and the number folded. A Source with 120 lines shows five and "Show 50 more"; one click shows 55 and "Show 50 more"; a second shows 105 and "Show 15 more"; a third shows all 120 and no Show more. A click never leaves a single line behind a Source's fold (six lines show whole today): a Source with 56 lines shows five and "Show 51 more". |
| FH-W1 | At Wide the diagram stays in the pinned column beside the rows, so the highlighted node and the highlighted Show more are on screen together. At Regular, when the pane is narrower than the rows' minimum, the diagram is under the rows, as today; the highlight works the same. Nothing moves between widths. |

### 4.1 Wide layout

| View | Regular (pane below 1600 pt) | Wide (pane 1600 pt and up) | What moves where |
| --- | --- | --- | --- |
| Person page, Sources section | The rows, with the record links diagram in a pinned column beside them, or under them when the rows would be narrower than their minimum | Full width (Page columns' `primaryFill`): the rows fill the pane, the diagram in its pinned column beside them | Nothing; this board adds a state to the Show more rows only |

---

## 5. Suggested frames

Each frame is drawn at **Regular** (pane 1102 × 900) and **Wide** (pane 2296 × 1440).

| Frame | Requirement | Draw |
| --- | --- | --- |
| 1a / 1b | FH-1, FH-W1 | A person with nine records on "Hamburg parish register 1820–1840", one graph per entry, and records on three other Sources. The register's row shows five lines and "Show 4 more". The pointer is on the node of the seventh entry, "Entry 1834/112": the node and its arcs are highlighted, the register's Show more has the highlight, no line has it |
| 2a / 2b | FH-1 | Frame 1 with the pointer on the node of the third entry: its shown line has the highlight, the Show more does not |
| 3a / 3b | FH-2, FH-W1 | A person with records on 60 Sources: 50 rows and "Show 10 more". The pointer is on a node whose Source is the 55th: the section's Show more has the highlight |
| 4 | FH-3 | Regular only: a Source with 120 lines in three states side by side: folded ("Show 50 more" after five lines), after one click ("Show 50 more" after 55), after two ("Show 15 more" after 105) |

---

## 6. UI building-block inventory

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Show more row | Snowflake | Extend | `macos/App/Features/Conclusions/ConclusionShowMoreRow.swift` `ConclusionShowMoreRow` | Gains a highlighted state (FH-1, FH-2) and the page count (FH-3). Still a small ghost Button; not a link, not a Badge |
| Record line highlight | Snowflake | Ship | `macos/App/Features/Conclusions/PersonSourcesSection.swift` `PersonSourcesLineView` | `PVColor.surfaceSelected` behind the line, `PVRadius.sm` corners. FH-1's recommended look reuses these tokens; no new token |
| Source row | Snowflake | Ship | `PersonSourcesSection.swift` `PersonSourcesRowView` | Unchanged |
| Record links diagram | Snowflake | Ship | `macos/App/Features/Conclusions/PersonRecordLinksColumn.swift` | Unchanged; its node highlight drives FH-1 and FH-2 |
| Button | Component | Ship | `macos/App/DesignSystem/Components/Button/` | The Show more's Button, ghost, small, as today |

**Non-goals.** No new kit primitive, no new color token, no new copy key: "Show %lld more" stays and only its number changes. No count of the highlighted graph's folded lines on the row.

---

## 7. Handoff

When the board is agreed: record the frames and the FH-1 look on the **Board** line here, in [`README.md`](README.md) and on the plan's **Board** line; the plan copies FH-1 – FH-3 and FH-W1 into its Scope and moves from `Blocked — on board S16-D2` to `Planned`. When S16-07 lands, the brief moves to the spike's design archive.
