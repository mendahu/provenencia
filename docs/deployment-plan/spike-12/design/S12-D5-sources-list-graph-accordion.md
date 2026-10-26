# S12-D5 — Sources list graph accordion

**Kind:** Claude Design board (**enhancement** of the Sources list and of the Source page's Evidence graphs rows)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Sources list — every row's Evidence graph zone becomes an expander over one sub-row per graph; the Status filter's last option is renamed. Source page — each Evidence graphs row gains the count line.
**Implements later as:** PR **S12-14**
**Depends on:** S12-08 (a graph is a place), S12-D2 (the Source page graph rows this board adds a line to)
**Related:** [`../source-structure.md`](../source-structure.md#where-graphs-show) (the design), [`../deployment-plan.md`](../deployment-plan.md) working rules "Sources list", "Source page", "One graph list", "Name and description are editable; graphs list alphabetically by name". S5-D2 drew today's split row; S12-D6 added its right-click menu; S12-D2, S12-D3 and S12-D4 drew the Source page graph row this board extends.
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

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### This brief is an enhancement

The Sources list keeps its header, toolbar (search, type filter, status filter, sort), caption band and split rows: the page zone on the left (thumbnail, title, type, last updated, ref, right-click Copy…) and the 248 pt Evidence graph zone on the right with its hairline. This board changes **what the graph zone holds** and adds **sub-rows under an expanded row**. On the Source page it adds **one line** to each Evidence graphs card. Do not redraw the page zone, the toolbar, or the Source page's other sections, and do not start a second list.

---

## 1. Objective

A census Source can now hold several Evidence graphs (Jones household, Smith household, …). The Sources list still has one **Open graph** zone per Source, which opens the alphabetically first graph and shows only its counts. The researcher cannot see a Source's other graphs, their progress, or open them from the list. The Source page lists the graphs but shows no progress.

**What changes.**

- **Sources list.** Every row's graph zone becomes an expander that reads the Source's graph count ("1 evidence graph", "3 evidence graphs"). Expanding it shows one sub-row per graph under the Source row, in name order. Each sub-row carries the graph's name and ref on the left and, in the graph zone column, today's **Open graph** zone for that graph: the label and its count line ("12 subjects · 40 observations", or "0 subjects · not started").
- **Status filter.** The statuses are now read per graph, with a new meaning for In progress (§2), and the last option reads **Promoted** instead of "Fully promoted".
- **Source page.** Each Evidence graphs card shows the same count line under its ref and description.

**What stays.** The page zone and its click (opens the Source page). The graph zone's width, hairline and hover. A Source with no Artifact keeps today's sunken "Needs an artifact" zone with no expander. Search matches Source title and ref only (graphs are found through the omnibar, S12-15).

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Every Source has at least one graph | Every row with an Artifact has an expander; a migrated project's rows read "1 evidence graph". |
| Graphs list alphabetically by name | Sub-rows are in name order, A to Z, not draggable. "Default" sorts with the other names. |
| A graph has a name, a ref and a description | Sub-row: the name (free text, shown as typed; "Default" is stored text, not a label) and the `GPH-` ref in mono. The description is not on the sub-row (it is on the Source page card). |
| Open needs a linked Artifact | A graph opens only when at least one Artifact is linked to it. A graph that links none has its Open zone disabled, with the tooltip and VoiceOver hint "Add an artifact to open this evidence graph" (today's Source page copy). |
| A Source with no Artifact | Its one Default graph can't open and has nothing to count. The row keeps today's sunken "Needs an artifact" zone with its tooltip; there is no expander. |
| Counts are per graph | A sub-row's count line counts that graph only. There is no Source-level total on the collapsed row. |
| The count line | Three states, today's rules: loading → a thin hairline in place of the text; 0 subjects → "0 subjects · not started" (italic "not started"); otherwise "N subjects · M observations" in mono micro. The list's sub-rows and the Source page cards draw it the same way. |
| Graph status | Each graph has one status from its counts. **Empty**: no subjects. **In progress**: at least one subject and none promoted yet. **Promoted**: at least one subject promoted (by any Promote run), even if others are not. |
| Status filter on a Source | A Source shows under a status when **at least one** of its graphs has that status, so a Source can appear under two statuses. Expanding a filtered row shows all of its graphs, not only the matching ones. |
| Expanded state | Remembered per Source until the app quits: leaving the list and coming back keeps the rows that were open. It is not part of Back / Forward history. A new filter or sort keeps it. |
| The graph list and counts load apart from the Sources | The Source rows can show before the graph list. Until the graph list loads, the graph zone shows a skeleton line in place of the expander label. If the graph list fails, the zone shows a danger icon whose tooltip is the app's standard message for the failure (the board writes "Couldn’t read the evidence graphs." as a stand-in). |
| Counts update live | Adding an Observation in one graph changes only that graph's count line, in the list and on the Source page, without a reload. |

### 2.1 What this board is not

- Searching graphs, or graph hits (S12-15, omnibar).
- Create, edit, delete or link graphs from the list; a Promote button on a sub-row; a status badge on a sub-row or row.
- Filed / primary counts ("3 of 5 filed") anywhere.
- A Source-level total of subjects or observations.
- Any change to the Source page card other than the count line.

---

## 3. Implementation gate (S12-14)

| Ships in **S12-14** | Stays out |
| --- | --- |
| The graph zone expander on every row with an Artifact; the graph sub-rows with per-graph Open and count line; the loading and failure states of the zone; the session-held expanded state; the Status filter's per-graph rule and the **Promoted** label; the count line on the Source page's Evidence graphs cards. | Graph search (S12-15). Graph create, edit, delete and links from the list. Promote from the list. A kit accordion or tree component. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GA-1 | On a row whose Source has an Artifact, the graph zone is one button, the expander: the network icon, the graph count in words ("1 evidence graph", "3 evidence graphs"), and a trailing chevron (down when collapsed, up when expanded). It fills the zone and keeps today's hover (muted text lifts to primary, hover fill). Clicking anywhere in the zone toggles it. |
| GA-2 | Expanded, the row is followed by one sub-row per graph, in name order, before the next Source row. A sub-row uses the same two columns and hairline as the row. Left column: the graph name (body small, medium, one line, truncated) and its ref (mono micro, muted) on one line, indented to align with the Source title (past the thumbnail). Right column: the graph's Open zone (GA-3). Sub-rows sit on the sunken surface and are separated by the same divider as rows; the last sub-row ends with the row divider. |
| GA-3 | A sub-row's graph zone is today's **Open graph** zone for that graph: network icon, "Open graph", and the graph's count line (§2). Clicking it opens the graph's canvas. The left column is not a button. |
| GA-4 | A graph that links no Artifact has its Open zone disabled with the tooltip "Add an artifact to open this evidence graph"; its count line still shows. |
| GA-5 | A Source with no Artifact keeps today's sunken "Needs an artifact" zone and tooltip; it has no expander and no sub-rows. |
| GA-6 | Loading: until the graph list has loaded, the expander zone shows the icon and a skeleton line in place of the label, and does not toggle. Each sub-row's count line shows the thin loading hairline until the counts arrive. Failure: the zone shows the icon and a danger mark with the standard error as tooltip, and does not toggle. |
| GA-7 | Expanded rows stay expanded when the researcher leaves the list and returns, and after a filter, sort or search change, until the app quits. Every row starts collapsed in a new launch. |
| GA-8 | The Status filter's options read **All statuses**, **Empty**, **In progress**, **Promoted**. A Source shows under a status when at least one of its graphs has it (§2). |
| GA-9 | VoiceOver: the expander reads "{N} evidence graphs" with the expanded / collapsed state; a sub-row reads "{graph name}, {ref}", then its Open zone reads "Open graph, {count line}" as the row's zone does today, with the blocked hint when disabled. Keyboard: the expander and each sub-row's Open zone are focusable buttons in reading order. |
| GA-10 | Many graphs (12) list in full under the row with no inner scroll; the list scrolls. A 60-character graph name truncates in the left column; the graph zone keeps its 248 pt width. |
| GA-11 | Source page: each Evidence graphs card shows the count line (§2) on its own line under the ref, above the description, in the same mono micro muted style as the list. Open, Edit and Delete do not move. |

---

## 5. Suggested frames

Draw the Sources list at its default window width; the Source page at its content width.

1. **Sources list, collapsed** — five rows: an obituary ("1 evidence graph"), a census ("3 evidence graphs"), a will ("1 evidence graph"), a photo with no Artifact (sunken "Needs an artifact"), a letter ("1 evidence graph"). Hover on the census zone (GA-1, GA-5).
2. **Census expanded** — the census row expanded under it: "Default" `GPH-4K2PQ` "Open graph / 0 subjects · not started"; "Jones household" `GPH-JONES` "Open graph / 12 subjects · 40 observations"; "Smith household" `GPH-SMITH` "Open graph / 7 subjects · 19 observations". The obituary row below collapsed (GA-2, GA-3).
3. **One-graph row expanded** — the obituary with its one "Default" sub-row (GA-2).
4. **Blocked graph** — a fourth census graph "Unlinked" with no Artifact, its Open zone disabled, tooltip shown (GA-4).
5. **Loading** — graph zones with skeleton labels; an expanded row's sub-row count lines as hairlines (GA-6).
6. **Graph list failed** — zones with the danger mark and tooltip (GA-6).
7. **Status menu open** — options All statuses, Empty, In progress, Promoted; then two frames of the list filtered, with Jones in progress and Smith promoted: under "In progress" the census shows, and under "Promoted" the census shows again (GA-8).
8. **Twelve graphs and a long name** — an expanded row with 12 sub-rows, one with a 60-character name, at the list's minimum width (GA-10).
9. **Source page** — the census's Evidence graphs section with three cards, each with its count line under the ref; one card with a description, one without; the "Default" card reading "0 subjects · not started" (GA-11).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Split row and columns | Snowflake | Extend | `Features/Sources/SourcesSplitRow.swift` `SourcesSplitRow`, `SourcesSplitLayout.columns` | The row's graph zone becomes the expander. Sub-rows reuse `SourcesSplitLayout.columns` so the hairline stays straight. Not a second row type with its own track. |
| Expander | Snowflake | **New** | `Features/Sources/SourcesSplitRow.swift` | A full-zone button with the existing `SourcesSplitZoneButtonStyle` hover, the network icon, the count, and the chevron from `PVExpandedState.chevron(isExpanded:)`; spoken state through `pvExpandedState`. Not `PVDisclosureButton`: that is a `PVButton` trigger, and this zone is a whole 248 pt hit area matching today's Open graph zone. |
| Graph sub-row | Snowflake | **New** | `Features/Sources/SourcesGraphSubRow.swift` (new) | Name, ref, and the Open zone of one graph. One call site (the list). Not a kit primitive, not the Source page card. |
| Open graph zone | Snowflake | Extend | `SourcesSplitRow.swift` `graphZone` | Today's zone, now drawn per graph on each sub-row instead of once per row. |
| Count line | Snowflake | **New** (lifted) | `Features/Sources/SourcesGraphCountLine.swift` (new) | Today's private `SourcesSplitRow.graphCountLine` lifted to one Sources view with its three states. Three call sites: the list sub-row, its VoiceOver text, the Source page card. Feature-shared, not kit. |
| Loading label | Component | Ship | `DesignSystem/Components/SkeletonBar/PVSkeletonBar.swift` | In place of the expander label. Not a spinner. |
| Failure mark | Component | Ship | `PVIcon` with the danger color | With the error tooltip. Not a Callout: the zone is 248 pt wide. |
| Blocked zone | Snowflake | Ship | `SourcesSplitRow.swift` (the `else` branch of `graphZone`) | Unchanged "Needs an artifact". |
| Status filter | Component | Ship | `PVSelect` in `Features/Sources/SourcesListView.swift` `statusMenu` | Only the last option's label changes. |
| Source page graph card | Snowflake | Extend | `Features/Sources/SourcePageGraphsView.swift` `SourcePageGraphRow` | Adds the count line. No other change. |

**Non-goals.** No kit accordion, tree or disclosure group: the expander is a Sources snowflake on today's zone style. No status badge or chip on a row or sub-row. No progress bar or meter. No local Button or Card.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](README.md), and move the S12-14 plan from `Blocked — on board S12-D5` to `Planned` with the frames' final copy in its Strings section.
