# S13-D10 — Person page: how the records connect

**Kind:** Claude Design board (**enhancement** of the agreed S13-D6 Sources section: it fills the place S13-D6 reserved for How the records connect, and moves that place beside the Source rows at both widths, user 2026-11-07)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Sources section: the record links diagram and its hover link to the Source rows
**Implements later as:** PR **S13-28** (How the records connect)
**Board:** agreed 2026-11-07, turn 15 "S13-D10 · How the records connect" on the [Person Detail canvas](https://claude.ai/artifact/N9V52ewQjxeNvh3Az3YkXQ) (page "Person Detail - Sources"), migrated from the [Claude Design project](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share). Placement answered before the board (user, 2026-11-07): beside the Source rows at both widths, pinned while the rows scroll; this replaces S13-D6's band above the rows at Regular and secondary column at Wide. The board stacks the diagram below a 1078pt pane, not 900 (user accepted); the Wide header keeps its centred band rather than the board's left alignment (user, 2026-11-07).
**Depends on:** the agreed S13-D6 board (the Sources section, its Source rows and graph lines, and the place it reserved for this view; [archived brief](archive/S13-D6-sources-section.md), built by S13-16 [#722](https://github.com/mendahu/provenencia/pull/722) and S13-17 [#729](https://github.com/mendahu/provenencia/pull/729)); the agreed S13-D1 frame (header, switcher, Page columns; [archived brief](archive/S13-D1-person-page-frame.md)); the read S13-27 ([`../prs/S13-27-person-record-links-read.md`](../prs/S13-27-person-record-links-read.md)).
**Related:** [`../evidence-graph-comparison.md`](../evidence-graph-comparison.md) Requirement 2 and § First version; [`../deployment-plan.md`](../deployment-plan.md) S13-28 scope entry, working rules "The view is a circular link diagram", "Slice 7 starts with a simple count", "Scoring and the view are two concerns", "Sections are places", "A full-width section keeps its states in the band"; dogfood bar 10; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** [`S13-D10-how-the-records-connect.review.md`](S13-D10-how-the-records-connect.review.md)

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

Extend the agreed S13-D6 Sources frames. The header, the switcher, the Source rows' content, their graph lines, Remove and the confirm stay exactly as agreed. **One change to S13-D6's layout** (user, 2026-11-07): the diagram sits **beside** the Source rows at **both** widths, not in a band above them at Regular, and it **stays pinned in view while the Source rows scroll**, so the researcher browses the list and the diagram side by side. At Wide the section uses the whole pane (Page columns' full-width mode) and the extra space goes to the diagram. Do not start a second Sources section.

---

## 1. Objective

A researcher on a Person's Sources section sees, beside the list of Sources and pinned in view while the list scrolls, **how the person's records connect**: each Evidence graph that holds a record of this person sits on a circle, labelled with the graph ref the Source rows already show, and between two graphs there is one arc for each **other person** both records name. An arc is thin when the two records share at most one fact about that person, thick when they share two or more. A record that shares no one with any other stands apart: it sits on the circle with no arc. Hovering a graph on the circle highlights its line in the Source rows, and hovering a graph line in the rows highlights the graph on the circle, so the researcher knows which record to go and look at. Clicking a graph on the circle opens that Evidence graph, as its line in the rows does.

This is a first version, built to be looked at on many real people before anything richer is designed (working rule "Slice 7 starts with a simple count"). **The arcs and their weights are the only data.** No label, tooltip or detail on an arc, on hover or on click.

**What changes.** The section becomes two columns at both widths: the Source rows, and beside them the diagram, which stays in view while the rows scroll. The Source rows' graph lines gain a highlight state, driven from the circle.

**What stays.** What S13-D6 agreed for the rows: their columns, graph lines, folds, Remove, the confirm, the empty, loading and error states of the section. Only the rows' column width changes (RL-W2). The header, the switcher, the other sections.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A node is an Evidence graph | One record's reading of one group of people (a census household page, an obituary, a register page). The circle holds every Evidence graph that holds an accepted record of this person: the same graphs the Source rows' graph lines name. A graph holding two records of the person (a duplicate) is one node. |
| The label | The graph's ref ("EVG-2K7PQ") in mono, the same ref the graph line shows. No numbers or letters are added. A graph with no ref (rare, an old project) is labelled with its name, cut to the width of a ref (RL-1 decides how labels fit). |
| The order | The read gives the graphs in the Source rows' order (Source title, then graph name). They go around the circle in that order, clockwise from the top, so graphs of one Source sit side by side. |
| An arc | One other person both graphs name, filed on one canonical person (the page's own person never draws an arc: every graph holds them). Two graphs that share three people get **three arcs**: the number of arcs is the data (user, 2026-11-07). |
| Thin or thick | Thin (weak): the two records agree on at most one fact about that person (a name, a birth, a relationship). Thick (strong): two or more. A shared person with nothing in common still draws a thin arc: sharing the person is itself the link. Conflicts are ignored in this version. |
| Stands apart | A graph with no arc to any other. It stays on the circle, unconnected. Bar 10: "A record that shares no one with another stands apart." |
| Groups | Two clusters of graphs with no arc between them are visible as two clusters; nothing names or boxes them. |
| No line data | Hovering or clicking an arc does nothing. The view never shows who the shared person is, how many facts, or why. |
| Hover both ways | Hovering a node highlights the node, its arcs and the graph's line in the Source rows (every line naming that graph, if there are two records). Hovering a graph line in the rows highlights its node and arcs. The other nodes and arcs recede. |
| Open | Clicking a node opens its Evidence graph with this person's Subject selected, exactly as the graph line's link does (a graph with two records opens on the first, in the rows' order). |
| The read | The diagram reads its own query (S13-27), separate from the Source rows' query: it loads and fails on its own. While the rows show, the diagram can still be loading. Its reload shows in the switcher row's "Refreshing" like any section read. |
| Sizes | Typical: 5 – 20 graphs. Budget: 50 graphs, 1,225 possible pairs; a household graph shares 2 – 4 people with another, an obituary a dozen. A graph that shares people with everyone can have 50+ arcs. Small: 1 graph (nothing to connect), 2 graphs. |

### 2.1 What this board is not

- What a line counted, its facts, its conflicts, or the shared person's name (§ Open of the requirements; a later version).
- A score beyond thin and thick, or a colour per strength beyond the weight.
- Shared events and places as arcs (persons only in this version).
- Pan, zoom, drag, or a layout other than the circle.
- Editing a claim or a link from the diagram; Promote.
- A legend that explains thin and thick in prose beyond one caption line (RL-9 decides whether there is one).
- Event and Place pages.
- Any change to the Source rows beyond their highlight state.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-28** | RL-1 … RL-14 and RL-W1 … RL-W5: the diagram beside the Source rows at Regular and Wide, pinned while they scroll, its states, hover both ways, open on click, keyboard and VoiceOver, light and dark. |
| Stays out | Everything in § 2.1. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| RL-1 | **The circle.** The person's graphs evenly spaced on a circle, in the read's order, clockwise from the top. Each node is a small dot drawn by the diagram (decide its diameter; recommended 8pt) with its graph ref beside it, outside the circle, in mono caption. The circle fills the diagram column less the room its labels need (decide that room; recommended: the width of one ref plus 8pt on each side). Decide how labels avoid each other at 50 graphs (recommended: labels follow the circle's tangent outward, and past 30 nodes the ref shows on hover and focus only, the dot staying). |
| RL-2 | **The arcs.** Each shared person is a curve between its two nodes inside the circle (decide: a chord bowed toward the centre, as a chord diagram draws it, recommended; or a straight line). Several arcs between the same two nodes fan out side by side so each can be counted. Thin and thick are two stroke weights of one colour; draw them so the difference reads at a glance at 5 and at 50 graphs. Decide the colour (recommended: the kit's muted line colour at rest). |
| RL-3 | **Stands apart.** A node with no arc is drawn the same as any node; nothing else marks it (recommended: the missing arcs are the mark). Say if the board wants a muted tone, and why. |
| RL-4 | **Hover a node.** The node, its label and its arcs take the highlight (recommended: the accent line colour); the arcs' other ends' nodes take a second, lighter highlight; every other node and arc recedes. At the same time the graph's line in the Source rows takes a highlight state (recommended: the kit's selected-row background, `surfaceSelected`, on the whole line). Pointer leaves: everything returns. |
| RL-5 | **Hover a graph line.** Hovering a graph line in the Source rows (anywhere on it, its Unpromote button included) does the same to its node as RL-4. A graph whose line is folded away ("Show N more") still highlights on the circle when hovered there; the rows do not unfold. |
| RL-6 | **Click a node.** Opens the Evidence graph with this person's Subject selected (as the graph line's link). The node takes the kit's hover and pressed states and focus ring, as the graph line's link does; no cursor change (no kit link changes the cursor). An arc is not clickable. |
| RL-7 | **No line data.** No tooltip, label or popover on an arc or on the circle's centre. A node's tooltip is its graph name and Source title ("Household 214 — Hamburg census 1867"): the circle shows only refs, so it names the Source the line's row already shows. |
| RL-8 | **Small.** One graph: decide what the region shows (recommended: the one node alone, centred, and one muted caption line "One record — nothing to connect yet"). Two graphs with no arc: two nodes, no arc. |
| RL-9 | **Caption.** Decide whether the region carries a one-line caption under a section-style title (recommended: the micro-caps title "How the records connect" and, muted, "An arc is a person both records name; thick when they agree on two or more facts."). Keep it to one line; no legend box. |
| RL-10 | **Loading, error, refreshing.** The diagram's first load: a small spinner centred in the region, the Source rows drawn beside it. A failed read: a compact danger Callout in the diagram column ("The connections could not be read") with the error, the tone every failed section read uses, the Source rows unaffected. A reload keeps the diagram drawn ("Refreshing" in the switcher row, as the rows). |
| RL-11 | **Keyboard.** Decide where the diagram sits in the Tab order (recommended: after the Source rows, as it is drawn beside them at both widths) and that Tab reaches each node in circle order, with the kit focus ring; Return opens the graph. Focus on a node highlights it as hover does (RL-4). |
| RL-12 | **VoiceOver, dark mode.** VoiceOver reads the region as a group "How the records connect, 7 records", then each node as a button "EVG-2K7PQ, Household 214, Hamburg census 1867, connected to 3 records" (the count of other graphs it has an arc to, the only thing a listener cannot see; no arc counts or weights) or "…, shares no one". Decide the wording. Frames 1, 1b, 3 and 4 also in dark. |
| RL-14 | **No diagram column while the section itself shows a state.** While the Source rows load, fail, or have no records, the section shows S13-D6's spinner, Callout or EmptyState full width, as today, and there is no diagram column. The diagram column appears with the rows. |
| RL-13 | **Pinned.** The diagram stays in view while the page scrolls the Source rows: it scrolls with the page until its top is 24pt (the page's vertical padding) below the top of the visible area under the pinned header and switcher, then holds there; at the end of the section it scrolls away with the last rows. It never grows taller than the visible area; decide how it fits when the window is short (recommended: the circle shrinks to fit, labels keep their size, down to a 240pt circle; below that the title and caption stay and the column scrolls with the page). |
| RL-W1 | **Wide.** The section uses the whole pane between the gutters (full-width mode). The Source rows keep their Regular measure (the citation column at most its S13-D6 width); the diagram takes the rest, up to a maximum the board decides, and the two columns are centred as a group past that. The header keeps its band. |
| RL-W2 | **Regular.** The rows and the diagram share the primary column side by side (pane 1102: about 1038pt between the gutters). Decide the diagram column's width (recommended: 340pt, the kit inspector width; at most 364pt, so the citation column keeps at least 200pt) and what the rows give up: the citation column narrows first; the Type, Date and Ref columns keep their S13-D6 widths unless the board shows a reason. Say how the long title (two lines, cut) reads at that width, and draw the **graph lines** at that width: a line holds the graph link, the graph ref, the Subject's name and ref, and Unpromote, about 250pt of fixed parts in a column of about 180 – 200pt after the indent (recommended: the refs and the Subject's name drop to a second caption line under the graph link, Unpromote staying on the first). |
| RL-W5 | **Narrow panes.** The window can be as narrow as 700pt (a pane of about 436pt). Decide the pane width below which the side-by-side arrangement gives way, and what it becomes (recommended: below a 900pt pane, the diagram stacks under the Source rows at the full column width and is not pinned; the user asked for the same arrangement at every width, so say why if the board stacks it, or show it beside at 436pt). |
| RL-W4 | **States at Wide.** The diagram's spinner, Callout and one-record caption sit centred in the wider diagram column (RL-10, RL-8). |
| RL-W3 | **Crossing 1600 keeps state.** The hovered or focused node, the pin and the scroll position survive a resize across 1600 (rule 5). A note on the Wide frame. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Sources section with the diagram | Source rows and the diagram side by side in the primary column; the diagram pinned while the rows scroll | Full-width mode: the rows at their Regular measure, the diagram taking the rest up to its maximum, pinned; the group centred past that | Only the diagram's width; the arrangement is the same |
| Diagram states | Spinner, Callout or one-node caption in the diagram column | The same, centred in the wider diagram column | Nothing |

---

## 5. Suggested frames

Each frame is the S13-D1 Person page pane only (no app sidebar), switcher on **Sources**, at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440); the Wide frame sits beside its Regular frame.

Sample person: **Johann Weber** (man, PER-3HK9W), with S13-D6's seven records on seven graphs, in the read's order around the circle:

| # | Graph (ref) | Source |
| --- | --- | --- |
| 1 | Household 214 (EVG-2K7PQ) | Hamburg census 1867 |
| 2 | Household 388 (EVG-5MD8R) | Hamburg census 1867 |
| 3 | Volume 3 (EVG-8TZ4N) | Hamburger Adressbuch 1861 |
| 4 | Issue 221 (EVG-1QR9D) | Hamburger Nachrichten, obituary 1880 |
| 5 | Register 2 (EVG-6WB3M) | Konfirmandenbuch St. Ansgarii 1831 |
| 6 | Register 3 (EVG-4PL7T) | Konfirmandenbuch St. Ansgarii 1831 |
| 7 | Page 7 (EVG-9CN2H) | Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte… |

Arcs (one per shared person):

| Graphs | Shared person | Weight |
| --- | --- | --- |
| 1 – 2 | Anna Weber (wife) | thick |
| 1 – 2 | Karl Weber (son) | thin |
| 1 – 4 | Anna Weber | thick |
| 1 – 4 | Karl Weber | thick |
| 2 – 4 | Anna Weber | thin |
| 1 – 5 | Heinrich Weber (father) | thick |
| 5 – 6 | Heinrich Weber | thin |
| 5 – 6 | Maria Weber (mother) | thin |

Graphs 3 (Volume 3) and 7 (Page 7) stand apart. Remember: the shared persons' names are for the board author only; the view never shows them.

1. **Sources with the diagram.** Johann's section at Regular: the Source rows and, beside them, the diagram (RL-1, RL-2, RL-3, RL-9, RL-W2).
1b. **Pinned.** Frame 3's large person scrolled half way down the rows: the diagram held under the switcher; beside it, the end of the section, where the diagram scrolls away with the last rows (RL-13).
2. **Hover.** Frame 1 with the pointer on node 1 (EVG-2K7PQ): its arcs and node highlighted, nodes 2, 4 and 5 lightly highlighted, the rest receded; the Household 214 line in the rows highlighted. Beside it, the pointer on the Register 3 line in the rows, node 6 highlighted (RL-4, RL-5).
3. **Large.** A person in 50 graphs: 14 Sources, the Adressbuch as one Source of 9 volumes side by side on the circle, about 120 arcs, three graphs standing apart; labels legible (RL-1, RL-2).
4. **Small and states.** One graph (RL-8); two graphs with no arc; the diagram spinner beside drawn rows; the diagram's Callout beside drawn rows (RL-10).
5. **Keyboard.** The Tab order over frame 1 with a node focused, and the VoiceOver readings as notes (RL-11, RL-12).
6. **Dark mode.** Frames 1, 1b, 3 and 4 in dark (RL-12).
7. **Wide.** Frame 1 at Wide in full-width mode: the rows at their measure, the diagram larger beside them; frame 2's hover at Wide; frame 3 (50 graphs) and frame 1b (pinned) at Wide; frame 4's states at Wide; the rule-5 note (RL-W1, RL-W3, RL-W4).
9. **Narrow.** Frame 1 at a 436pt pane (the 700pt window) and at the RL-W5 breakpoint (RL-W5).
8. **Section states.** The section's own spinner, Callout and EmptyState, full width with no diagram column (RL-14).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, Page columns | Component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `DesignSystem/Components/PageColumns/PVPageColumns.swift` | RL-W1, RL-W2. Sources opts into full-width mode at Wide, as Timeline and Gallery do; the section lays out its two columns itself. No page secondary column. |
| Sources section, Source rows, graph lines | Snowflake | Extend | `Features/Conclusions/PersonSourcesSection.swift` | RL-4, RL-5, RL-W2. The section holds the rows and the diagram side by side; the graph line gains a highlight state; the rows' column widths are unchanged but the citation's. |
| List with a pinned side column | Snowflake | **New** | `Features/Conclusions/` *(new, named in the S13-28 plan)*, widths from `PVPageColumnsMetrics` | RL-13, RL-W1, RL-W2. One arrangement for a list with a side column beside it at both widths, pinned; starts from § 9's "List with a side pane". Second call site: the Event and Place pages with this view. |
| Visible-area reading | Snowflake | Extend | the Timeline's lifeline reading where the viewport sits over its section (`PersonTimelineTethers`), extracted to one page reader | RL-13. |
| Record links diagram | Snowflake | **New** | `Features/Conclusions/` *(new file, named in the S13-28 plan)* | RL-1 – RL-12. Drawn by the view itself on a SwiftUI canvas: a fixed circle with no pan, zoom or drag, so not the Evidence graph's `GraphCanvas` host. Second call site: the Event and Place pages, when they adopt sections. |
| Node | Snowflake | New | inside the diagram | RL-1, RL-6. A dot and the mono ref; hover and pressed from `PVHoverEffect`, the focus ring from `pvFocusRing` (the `PVInkLink` link style's parts), Return from `pvActivatesOnReturn`. Part of the diagram; second call site with it: the Event and Place pages. |
| Arc | Snowflake | New | inside the diagram | RL-2. Two stroke weights of one line colour token. Part of the diagram; second call site with it: the Event and Place pages. |
| Title and caption | Component | Ship | micro-caps and caption text styles | RL-9. |
| Graph line highlight | Component | Ship | the kit's selected-row background (`PVColor.surfaceSelected`) | RL-4, RL-5. |
| Loading | Component | Ship | the kit spinner (`ProgressView`) | RL-10. |
| Error | Component | Ship | `PVCallout` (danger, compact) | RL-10, as the section host's failed read. |
| One-record caption | Component | Ship | caption text style, muted | RL-8. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A tooltip, label or popover on an arc | Working rule "The view is a circular link diagram": no line data. |
| A legend box, a score, a colour per strength | Two weights are the data (§ 2.1). |
| Pan, zoom, drag, or the `GraphCanvas` host | A fixed circle of at most about 50 nodes. |
| A new kit primitive (a chord diagram component) | One call site until Event and Place adopt sections. |
| A change to the Source rows beyond the highlight | S13-D6 is agreed. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames; the node mark and the label placement at 50 graphs (RL-1); the arc shape, the two weights and the colour (RL-2); the stands-apart answer (RL-3); the highlight treatment on the circle and on the graph line (RL-4); the one-graph answer and copy (RL-8); the caption answer and copy (RL-9); the error copy (RL-10); the Tab order (RL-11); the VoiceOver wording (RL-12); the diagram's width at Regular and its maximum at Wide, and the citation column's width at Regular (RL-W1, RL-W2); how it fits a short window (RL-13); the graph lines at Regular (RL-W2); the narrow-pane breakpoint and arrangement (RL-W5); and every other change from this brief with the user's answer.
2. Refresh the [S13-28 plan](../prs/S13-28-how-the-records-connect.md): copy the answers into its Decisions and **Design → macOS**, describe every frame in words, close its Open questions, run `review-plan` again, and set it to `Planned`.
3. S13-28 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
