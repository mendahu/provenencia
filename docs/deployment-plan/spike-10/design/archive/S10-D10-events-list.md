# S10-D10 — Events list

**Kind:** Claude Design board (**enhancement** of the Events list)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Events list — the row
**Implements later as:** PR **S10-22**, together with [S10-D9](S10-D9-persons-list.md) and the other Conclusion list brief.
**Depends on:** [S10-D9](S10-D9-persons-list.md) (the column this board copies)
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-22). Persons is **S10-D9** and draws the column first. Places is **S10-D11**. The Sources list is **S10-D8** (S10-41).
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

Add one column to the Events list row. It is the same column the Persons list (S10-D9) draws: same width, same text, same help, same em dash. Copy it; do not restyle it. Do not draw a second list. Do not throw away the section header, the count, the event mark thumb, the title, the date and place line, the ref, the chevron, the empty state, or the loading skeleton.

---

## 1. Objective

The researcher can see, on each row, when they last worked on that event.

**What changes.** A last-updated column on each row, between the text and the ref, identical to the Persons list.

**What stays.** List order (dated events by the date window, then the rest by ref), the title, the date and place line, the ref, the chevron, the event mark thumb, row height, hover, focus, keyboard movement, the count in the header, the empty state, and the loading skeleton. There is no sort control, no filter, and no search field on this page. The thumb stays the event mark; events have no likeness.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The row is the kit list row | Thumb (44pt, the event mark), then the title over the date (mono) and place (italic) line, then a trailing mono ref and a chevron. |
| Last updated is the last work on this event | It moves when the researcher files, releases, or pins a Subject onto this event, or edits an Observation (its value, date, name, or note) on a Subject filed to it. Editing a property, a term, a citation, or a credibility grade does not move it. |
| One column for three lists | The column is the Persons list column (S10-D9): fixed width, after the text, before the ref. “Today, 3:15 PM”, “Yesterday”, then a short date (“12 Oct 2026”), in the viewer’s locale. Caption size, muted, not mono, not a badge. Help is the full stamp (“14 Oct 2026 · 3:15 PM MDT”). No recorded change: “—”, no help. |
| The date line keeps its width | The event’s own date stays mono and fixed on the secondary line. It is a different date from the column: the date the event happened, not the date the researcher worked on it. The column is not mono, so the two do not read as one value. | |
| VoiceOver reads it | The row’s spoken label ends with “updated {the cell text}”, or “last updated unknown” for the em dash. |

### 2.1 What this board is not

- A sort by last updated, or any sort control. The list stays in date order.
- A filter, a status, or a search field. Those are the Sources list (S10-D8) only.
- A different treatment of the column from the Persons list.
- A likeness or photo in the thumb.
- A new list component.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S10-22** | EVL-1…EVL-4, with S10-D9 and the other Conclusion list brief. |
| Stays out | Sort, filter, search. A likeness thumb. A created time. A live clock. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EVL-1 | Each row has the last-updated column from S10-D9 PSL-1…PSL-4: same fixed width, same position (after the text, before the ref), same text rules, same help, same em dash. |
| EVL-2 | The title and the date and place line truncate before the column or the ref shrink. Row height, thumb, ref, chevron, hover, pressed, focus ring, and keyboard behavior are unchanged. The list stays in date order, and does not change when a time moves. |
| EVL-3 | The loading skeleton reserves the column the way S10-D9 PSL-6 does. |
| EVL-4 | The row’s spoken label ends with “updated {cell text}”, or “last updated unknown” for the em dash. |

---

## 5. Suggested frames

1. **Times.** Six events: “Today, 3:15 PM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, “9 Jan 2024”, and “—”. The column lines up with the Persons list column.
2. **Narrow window.** A long title (“Marriage of James Robins and Eliza Hartley”), a long place, and +2 places. The title and the place truncate; the date keeps its width. The column, the ref, and the chevron keep their width.
3. **No secondary line.** An event with no date and no place: one-line title, the column and ref vertically centered.
4. **Loading.** The skeleton with the column bar beside the ref bar.
5. **Empty.** No events: the empty state as today. No column.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Events list page | Snowflake | Ship | `Features/Conclusions/EventsListView.swift` | Same page, same order. |
| Shared Conclusion list page | Snowflake | **Extend** | `Features/Conclusions/ConclusionListPage.swift` | The same column S10-D9 adds. Not a per-kind copy. |
| List row | Component | **Extend** | `DesignSystem/Components/List/PVList.swift` | The optional column S10-D9 adds. Not a second extension. |
| List skeleton | Component | **Extend** | `PVListSkeleton` in `DesignSystem/Components/List/PVList.swift` | The column bar S10-D9 adds. |
| Title | Snowflake | Ship | `Features/Conclusions/ConclusionListRow.swift` | Unchanged. |
| The date and place line | Snowflake | Ship | `Features/Conclusions/EventsListView.swift` | Unchanged. The time is not added here. |
| Last-updated text | Snowflake | Ship | from S10-D9 | Same text and help as Persons. |
| Thumb | Component | Ship | `PVThumbnail` | The event mark, unchanged. |
| Empty state | Component | Ship | `PVEmptyState` | Unchanged. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A event-specific last-updated treatment | One column for three lists (S10-D9). |
| Sort, filter, or search controls | Events has none. Sources owns those (S10-D8). |
| A likeness thumb | Likeness is a Person property. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-22** against EVL-1…EVL-4 with S10-D9 and the other Conclusion list brief.
