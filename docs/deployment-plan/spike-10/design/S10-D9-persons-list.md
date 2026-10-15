# S10-D9 — Persons list

**Kind:** Claude Design board (**enhancement** of the Persons list)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Persons list — the row
**Implements later as:** PR **S10-22** (the last-updated column). The same board draws the likeness thumb, which ships in **S10-35**, and the preferred portrait as that thumb, which ships in **S10-39**.
**Depends on:** —
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-22, S10-35, S10-39). Events is **S10-D10** and Places is **S10-D11**; both copy this board’s column. The Sources list is **S10-D8** and uses the same column in S10-41. The preferred-portrait control is drawn on the Person page board, **S10-D12**, not here.
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

Add one column to the Persons list row, and fill the thumb that is already there. Do not draw a second list. Do not throw away the section header, the count, the thumb, the title, the +N badge, the birth and death line, the ref, the chevron, the empty state, or the loading skeleton.

S10-22 builds the column. Draw the likeness thumb and the preferred portrait on this same list anyway. S10-35 and S10-39 fill those frames. They do not get another board.

---

## 1. Objective

The researcher can see, on each row, when they last worked on that person. The list also shows the person’s face once a likeness exists.

**What changes.** A last-updated column on each row, between the text and the ref. Later, the thumb shows a likeness crop instead of the person mark.

**What stays.** List order (by name, then ref), the title, the +N badge, the birth and death line with its +N places badge, the ref, the chevron, row height, hover, focus, keyboard movement, the count in the header, the empty state, and the loading skeleton. There is no sort control, no filter, and no search field on this page.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The row is the kit list row | Thumb (44pt), then the title over a secondary line, then a trailing mono ref and a chevron. The list is one card with hairlines between rows. |
| Last updated is the last work on this person | It moves when the researcher files, releases, or pins a Subject onto this person, or edits an Observation (its value, date, name, or note) on a Subject filed to this person. Editing a property, a term, a citation, or a credibility grade does not move it. It is the time of that change, not the time the person was created. |
| Recent reads as a day | A change today reads “Today, 3:15 PM”, with the time in the viewer’s locale. Yesterday reads “Yesterday”. Anything older reads as a short date, “12 Oct 2026”, in the viewer’s locale. There is no “3 minutes ago”. |
| The full time is the help | Hovering the cell shows the full stamp the note rows use: “14 Oct 2026 · 3:15 PM MDT”. |
| Unknown is an em dash | A person with no recorded change shows “—” in the column, muted. This is rare: history from before this spike that could not be mapped. It has no help. |
| The column is its own column | A fixed width, so the times line up down the list. It sits after the text and before the ref. It is not appended to the birth and death line, and it is not stacked under the ref. The text truncates first; the column and the ref keep their width. |
| The cell is quiet | Muted text, the caption size the secondary line uses. It is not a badge, not a chip, and not mono. One line. It is vertically centered on the row, like the ref. |
| Changes show up | After an edit elsewhere, the row’s time moves the next time the list is shown. The row does not reorder; the list stays sorted by name. |
| VoiceOver reads it | The row’s spoken label ends with “updated {the cell text}”, or “last updated unknown” for the em dash. |
| The thumb becomes a face (S10-35) | When any Subject filed to this person has a likeness (a cropped region of a photo), the thumb shows the earliest likeness crop, filled and cropped to the 44pt square with the kit’s thumb radius. A person with no likeness keeps the person mark. Events and Places keep their marks. |
| The preferred portrait leads (S10-39) | The researcher marks one likeness as preferred on the Person page. The list thumb then shows that crop instead of the earliest one. The list has no control for it and no badge saying it is preferred. Clearing the mark returns the earliest likeness. |
| A crop may still be loading | While the crop image loads, the thumb shows the kit’s loading state, then the image. A crop that cannot be read shows the person mark, not an error. |

### 2.1 What this board is not

- A sort by last updated, or any sort control. The list stays sorted by name.
- A filter, a status, or a search field. Those are the Sources list (S10-D8) only.
- A “created” time, or who made the change.
- A live clock that ticks “Today” into “Yesterday” while the list sits open.
- A control to pick the preferred portrait. That is the Person page (S10-D12).
- A gallery, or more than one face per row.
- A likeness on Events or Places.
- A new list component. This is the kit list row with one more column.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S10-22** | PSL-1…PSL-6. The last-updated column, its help, the em dash, the spoken label, and the skeleton. Events (S10-D10) and Places (S10-D11) ship the same column in the same PR. |
| Ships in **S10-35** | PSL-7. The earliest likeness as the thumb. |
| Ships in **S10-39** | PSL-8. The preferred portrait as the thumb. |
| Stays out | Sort, filter, search. A created time. A live clock. A preferred-portrait control on the list. A likeness on Events or Places. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PSL-1 | Each row has a last-updated column after the title and secondary line and before the ref. The column has one fixed width for every row, sized to fit “Yesterday”, “Today, 12:59 PM”, “12 Oct 2026”, and the US form “Sep 30, 2026” without truncating. Its text is leading-aligned in the column. |
| PSL-2 | The cell reads “Today, {time}” for a change on today’s date, “Yesterday” for yesterday’s date, and a short date (“12 Oct 2026”) for anything older. Dates and times are in the viewer’s locale and time zone. Caption size, muted, one line, not mono, not a badge. |
| PSL-3 | Hover help on the cell is the full stamp: “14 Oct 2026 · 3:15 PM MDT”. |
| PSL-4 | A person with no recorded change shows “—”, muted, in the same column, with no help. |
| PSL-5 | The title truncates before the column or the ref shrink. Row height, thumb, ref, chevron, hover, pressed, focus ring, and keyboard behavior are unchanged. The list order does not change when a time moves. |
| PSL-6 | The loading skeleton reserves the column with a short bar, so the first load does not shift the row when times arrive. The row’s spoken label ends with “updated {cell text}”, or “last updated unknown” for the em dash. |
| PSL-7 | (S10-35) A person with a likeness shows the earliest likeness crop in the 44pt thumb, filled to the square with the thumb radius. No likeness: the person mark. Loading: the kit thumb’s loading state. Unreadable: the person mark. |
| PSL-8 | (S10-39) A person with a preferred portrait shows that crop in the thumb. No preferred portrait: PSL-7. The list shows no mark or control for the preference. |

---

## 5. Suggested frames

One list. The later frames fill the thumb on that list. Do not draw a second page.

1. **Times.** Six people. Times: “Today, 3:15 PM”, “Today, 9:02 AM”, “Yesterday”, “12 Oct 2026”, “3 Mar 2025”, and “—”. The column lines up down the list. The list is still sorted by name, not by time.
2. **Help.** Hover on “Yesterday”: the help reads “13 Oct 2026 · 4:40 PM MDT”.
3. **Long name, narrow window.** A long name with +2 and a long birth place. The title and the birth line truncate. The column, the ref, and the chevron keep their width.
4. **No birth or death.** A person with only a name: one-line text, the column and ref vertically centered.
5. **Loading.** The skeleton with the column bar beside the ref bar.
6. **Empty.** No people: the empty state as today. No column.
7. **Likeness (S10-35).** Three people with face crops, one with the person mark. Crops fill the square.
8. **Preferred portrait (S10-39).** The same row as frame 7, now showing a different crop of the same person. Nothing on the row says it is preferred.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Persons list page | Snowflake | Ship | `Features/Conclusions/PersonsListView.swift` | Same page, same order. |
| Shared Conclusion list page | Snowflake | **Extend** | `Features/Conclusions/ConclusionListPage.swift` | Passes the column to the list for Persons, Events, and Places alike. |
| List row | Component | **Extend** | `DesignSystem/Components/List/PVList.swift` | PSL-1, PSL-5. One optional fixed-width text column before the ref, with optional help. Content-agnostic: the list does not know it is a time. Not a new list. |
| List skeleton | Component | **Extend** | `PVListSkeleton` in `DesignSystem/Components/List/PVList.swift` | PSL-6. A column bar when the list has the column. |
| Section header and count | Component | Ship | `PVSectionHeader` | Unchanged. |
| Title and +N | Snowflake | Ship | `Features/Conclusions/ConclusionListRow.swift` | Unchanged. The +N stays `PVBadge`. |
| Birth and death line | Snowflake | Ship | `PersonLifeLine` in `PersonsListView.swift` | Unchanged. The time is not added here. |
| Last-updated text | Snowflake | New | the time formatter beside the existing note stamp | PSL-2…PSL-4. Plain muted text in the list column. Not a Badge, not a Chip. Second call site: the Sources row in S10-41 (S10-D8) uses the same text and help. |
| Thumb | Component | Ship | `PVThumbnail` | PSL-7, PSL-8 use its image and loading states. The mark stays the fallback. |
| Empty state | Component | Ship | `PVEmptyState` | Unchanged. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A “last updated” kit component, badge, or chip | The column is plain text in the list row. |
| A second list component, or a table with columns and headers | The kit list gains one optional column. |
| Sort, filter, or search controls | Persons has none. Sources owns those (S10-D8). |
| A preferred badge or picker on the row | The Person page owns the choice (S10-D12). |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-22** against PSL-1…PSL-6 with S10-D10 and S10-D11. **S10-35** fills PSL-7 and **S10-39** fills PSL-8.
