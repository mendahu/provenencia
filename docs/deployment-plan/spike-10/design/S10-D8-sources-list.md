# S10-D8 — Sources list

**Kind:** Claude Design board (**enhancement** of the Sources list)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Sources list — the toolbar, the row, and the empty states
**Implements later as:** PR **S10-41**
**Depends on:** the agreed S10-D9 Persons board (the last-updated column this board copies). This brief restates that cell in full; where the S10-D9 file and this brief differ, this brief, the working rule “Last updated”, and the [S10-22 plan](../prs/S10-22-list-last-updated.md) D-5, D-6, D-11 describe the agreed board.
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-41; working rules "The Sources list is one page", "Last updated", "Sources evidence filters count primaries in one read", "List scroll is a field on the history entry"). Persons is **S10-D9**, Events is **S10-D10**, and Places is **S10-D11**; none of them get the toolbar controls drawn here.
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

Add to the Sources list that exists. Do not draw a second Sources page, and do not turn it into a table with column headers. Keep the page header (title, description, count line, Add Source), the toolbar band with the type filter and the sort, the caption band (“Source” | “Evidence graph”), the split row (filing zone | Evidence graph zone), the empty catalog state, and the load error.

Copy the last-updated column from the agreed Persons board (S10-D9) as §2 restates it. Do not redesign it.

---

## 1. Objective

The researcher can find a Source quickly and see where each one stands. They can narrow the list to Sources whose graph is empty, still in progress, or filed. They can type part of a title or ref to find one. Each row shows when they last worked on that Source. The list comes back the way they left it: the same sort, type, status, and search, including after relaunch.

**What changes.** A last-updated column on each row. An evidence-status select and a search field in the toolbar. One empty state for “filters match nothing”, with a Clear filters button.

**What stays.** The row’s thumb, title, type line, ref, and Evidence graph zone with its counts. The type filter and the sort, with their options. Row height, hover, focus, and keyboard. The page header and the count line. The empty catalog state. No sort, filter, or search appears on Persons, Events, or Places.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The row is a split row | The filing zone (thumb 44pt, title over the type line, then the mono ref) opens the Source page. The Evidence graph zone (248pt) opens the graph and shows “N subjects · M observations”, or “Needs an artifact” when the Source has no artifact. A hairline divides the zones down the whole list. |
| Last updated is the last work under this Source | It moves when the researcher edits the Source, its notes, metadata, credibility, artifacts, files, or anything on its Evidence graph (subjects, citations, observations). It is the time of that change. |
| The cell is the Persons cell | A history icon, then “Today, 3:15 PM”, “Yesterday”, or a short date in the viewer’s locale (the frames use day month year, “12 Oct 2026”; `en_US` reads “Oct 12, 2026”). Caption size, muted, one line, not mono, not a badge. A fixed 136pt column, leading-aligned, vertically centered. Help is “Last updated 14 Oct 2026 · 3:15 PM MDT”. No recorded change (no audit revision under the Source): a faint “—” with no help. The cell is not its own click target; a click opens the Source page like the rest of the filing zone. |
| The column is in the filing zone | Between the title stack and the ref. The title and the type line truncate first; the column and the ref keep their width. The Evidence graph zone keeps its 248pt. The caption band has no label for the column; the help names it. |
| Evidence status has three values | A primary is a person, event, or place on the Source’s graph. **Empty**: no primaries. **In progress**: at least one primary is not yet promoted. **Filed**: at least one primary, and every primary is promoted. Promoted means the subject has an accepted Identity Claim; a provisional claim counts as not promoted. Bridges (participation, relationship, location) are not counted. |
| The status is a filter, not a mark | The row does not show a status badge. The Evidence graph zone already shows the counts. |
| Search filters as you type | The query is trimmed; a query of only spaces is no search. The whole trimmed query must appear somewhere in the title or in the ref (“src-4f” finds SRC-4F2NP), ignoring case and accents. Several words are one phrase, not separate terms. “SRC4F” without the hyphen does not match. It is not the omnibar and does not search inside the Source. |
| Filters combine | Type, status, and search all apply at once. The count line reads “12 of 40 sources” whenever any of them narrows the list. |
| Settings come back | Sort, type, status, and search are remembered for this project and this researcher. Leaving and coming back, or relaunching, shows the same four. A remembered value that no longer applies (a deleted type) reads as its default, silently. |
| Filtering returns to the top | Changing the type, the status, the search, or the sort, or pressing Clear filters, scrolls the list to the top. Coming back with Back, Forward, or after relaunch returns to the row the researcher left; that restore already exists (S10-40) and is not redrawn. |
| Nothing matches | When the filters hide every row, the list area shows one compact empty state with a Clear filters button. Clear filters resets type, status, and search. It does not change the sort. |
| Counts can load after the rows | The rows and their times arrive together. The status comes from the Evidence graph counts the graph zone already shows, so it arrives when those counts do, a moment after the rows (today the graph zone shows a loading hairline then). While they load with a status other than All set, the list area below the toolbar shows the same spinner as the first load, with no caption band and no empty state. If the counts fail to load with a status set, the list area shows the error callout the page uses for a failed load. With status All, rows show as usual and the graph zone shows its loading line. |

### 2.1 What this board is not

- A status badge or chip on the row.
- A sort by status, or new sort options. Sort keeps Date added, Date updated, A–Z, Z–A.
- A search inside Source content, metadata, notes, or transcriptions. That is the omnibar.
- Saved views, several filter sets, or a filter sidebar.
- A table with column headers, or a column chooser.
- These controls on Persons, Events, or Places.
- A “created” time, or who made the change.
- A live clock that ticks “Today” into “Yesterday” while the list sits open.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S10-41** | SRL-1…SRL-9, and SRL-10’s scroll to the top for status, search, and Clear filters. |
| Already shipped | SRL-10’s scroll to the top for sort and type, and the restore on Back, Forward, a jump, and relaunch (S10-40). |
| Stays out | Status on the row. Omnibar search. Controls on the Conclusion lists. A new sort. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SRL-1 | Each row has the S10-D9 last-updated column in the filing zone, after the title stack and before the ref: history icon, then the text, 136pt wide including the icon, caption, muted, one line, leading-aligned. The title and type line truncate before the column or the ref shrink. The cell is not a separate click target. The Evidence graph zone is unchanged. |
| SRL-2 | The cell reads “Today, {time}”, “Yesterday”, or a short date, in the viewer’s locale and time zone. Hover help is “Last updated {full stamp}”. No recorded change: a faint “—”, no help. The filing zone’s spoken value is “{title}, {ref}, updated {cell text}” (“1851 census, SRC-4F2NP, updated Yesterday”), or ends “last updated unknown”. |
| SRL-3 | The toolbar holds, left to right: the search field, the type filter, the evidence-status select, then the sort. All four are the regular control size the type filter and sort use today. The search field has a search icon and the prompt “Search title or ref”. It is the only control that changes width: it grows to 280pt and shrinks to 160pt. The selects keep their size and never wrap. The controls sit at the leading edge; free space is after the sort. |
| SRL-4 | The status select has four options: “All statuses”, “Empty”, “In progress”, “Filed”. Its closed label is the option name. It is the same select as the type filter, without an icon. |
| SRL-5 | Typing in search filters the rows on each keystroke: the trimmed query as one phrase, anywhere in the title or the ref, ignoring case and accents. Spaces alone are no search. Clearing the field shows the rows again. |
| SRL-6 | Type, status, and search combine. The count line reads “{visible} of {total} sources” when any of them hides a row, and “{total} sources” otherwise. |
| SRL-7 | When the filters hide every row, the list area shows a compact empty state: title “No matching sources”, message “Try another filter, or clear them to see every source.”, and a secondary Clear filters button. The button resets type to All types, status to All statuses, and search to empty. Sort is unchanged. The toolbar stays visible above it. |
| SRL-8 | While the Evidence graph counts load and a status other than All is set, the list area below the toolbar shows the first-load spinner, without the caption band. It never shows the SRL-7 empty state for a list that has not been counted yet. If the counts fail with a status set, the list area shows the page’s load-error callout. |
| SRL-9 | Sort, type, status, and search persist per project and researcher, and restore when the Sources list is shown again, including after relaunch. A restored value that no longer applies reads as its default (a deleted type reads as All types), with no message. |
| SRL-10 | Changing the type, status, search, or sort, or pressing Clear filters, on this visit scrolls the list to the top. Coming back to the list with Back, Forward, a history jump, or after relaunch returns to the row it was left on (S10-40, already shipped). |

---

## 5. Suggested frames

One Sources list. Today in the frames is 14 Oct 2026, Mountain Time. Draw it with eight Sources of mixed types and covers.

1. **Times.** No filters. Rows show “Today, 3:15 PM”, “Today, 9:02 AM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, and one faint “—”. The column lines up down the list. The count line reads “8 sources”.
2. **Help.** Hover on “Yesterday”: the help reads “Last updated 13 Oct 2026 · 4:40 PM MDT”.
3. **Status menu open.** The status select is open over the list with All statuses, Empty, In progress, Filed. All statuses is checked.
4. **Filtered.** Status In progress and type Census. Three rows. The count line reads “3 of 8 sources”. Each row’s Evidence graph zone shows its counts as today.
5. **Search.** “smith” in the search field. Two rows whose titles contain Smith. The count line reads “2 of 8 sources”.
6. **Nothing matches.** Status Filed and “zzz” in search. The toolbar stays. The list area shows the compact empty state with the Clear filters button.
7. **Counting.** Status Empty is set and the counts have not arrived: the toolbar stays; below it, no caption band, only the first-load spinner.
8. **Narrow window.** A long title and a long type label. The title and type line truncate. The column, the ref, and the 248pt Evidence graph zone keep their width. The search field has shrunk to 160pt; the three selects keep their size.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sources list page | Snowflake | **Extend** | `Features/Sources/SourcesListView.swift` | Same page. The toolbar gains search and status; the filter-empty state changes (SRL-7); the spinner also covers counting (SRL-8). |
| Split row | Snowflake | **Extend** | `Features/Sources/SourcesSplitRow.swift` | SRL-1, SRL-2. The column goes in the filing zone before the ref. The graph zone is unchanged. |
| Last-updated cell | Component | **Extend** | `PVListDetailCell` *(extracted in S10-41)* from the detail cell S10-22 draws inside `PVList`, in `DesignSystem/Components/List/PVList.swift` | SRL-1, SRL-2. S10-22 draws the cell inline in the `PVList` row. S10-41 lifts it into one small view that `PVList` and `SourcesSplitRow` both instance. Not a Badge, not a Chip, not a second copy in the Sources feature. |
| Last-updated text | — (Platform value) | Ship | `LastUpdatedDisplay` in `Platform/TimestampFormat.swift` (S10-22) | SRL-2. Text, help, spoken form, and the 136pt width. |
| Search field | Component | Ship | `PVInput` with the search icon | SRL-3, SRL-5. The same composition as the Properties search field (`PVInput`, search icon), at the regular size of the neighboring selects, not the Properties page’s small size. Not a new search component. |
| Type filter | Component | Ship | `PVSelect` | Unchanged options. |
| Status select | Component | Ship | `PVSelect` | SRL-4. A second `PVSelect` beside the type filter. Not a segmented control, not chips. |
| Sort | Component | Ship | `PVSelect` | Unchanged options. |
| Count line | Snowflake | Ship | `SourcesListView` header | SRL-6. Existing “{visible} of {total} sources” string. |
| Nothing matches | Component | **Extend** (copy and action) | `PVEmptyState`, compact, in `SourcesListView.listBody` | SRL-7. The state exists today for the type filter. It gets new copy (replacing “No sources for this type” / “Try another type, or choose All types.”) and a Clear filters `PVButton`, secondary, in its existing action slot. |
| Counting | Snowflake | Ship | the first-load `ProgressView` in `SourcesListView` | SRL-8. Same spinner, below the toolbar, while the Evidence graph counts load with a status set. Not a skeleton: do not draw a split-row skeleton. |
| Evidence graph zone | Snowflake | Ship | `SourcesSplitRow` | Unchanged. No status badge here. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A status Badge or Chip on the row | Status is a filter. The graph zone already shows counts. |
| A segmented control or chip row for status | Four options in a toolbar: the kit Select already does this, like the type filter. |
| A search component | `PVInput` with the search icon is the kit search field. |
| A second last-updated cell drawn for Sources | The cell is the S10-D9 cell. |
| A filter bar or sidebar | The toolbar band already holds the filters. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-41** against SRL-1…SRL-10 (SRL-10 in part, §3), after S10-22 lands.
