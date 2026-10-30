# S14-D4 — Events list: the lead picture in the row thumb

**Kind:** Claude Design board (**enhancement** of the Events list)
**Spike:** Provenencia Spike 14 (odds and ends), batch 1 — Event and Place pictures
**View:** the Events list (`EventsListView`, on the shared Conclusion list page)
**Implements later as:** PR **S14-02** (Pictures in the app), with [S14-D2](S14-D2-event-page-pictures.md) (Event page), [S14-D3](S14-D3-place-page-pictures.md) (Place page), [S14-D5](S14-D5-places-list-thumbnail.md) (Places list)
**Board:** agreed 2026-10-30 — [*Events List*](https://claude.ai/design/p/2799d7a0-387e-4570-b138-fc387562ac91?via=share), file *Events List.dc.html*, "S14-D4", frames 1 – 2. Answers: dark mode is the board's theme toggle; at Wide the list keeps today's centered content column (EVL-W1 corrected).
**Depends on:** S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615)): each Event header carries the path of the picture it leads with. The Persons list as it ships today, whose row thumb already shows a Person's portrait. The agreed S14-D1 board (the Wide width class).
**Related:** [`../event-place-pictures.md`](../../event-place-pictures.md) § Thumbnail parity; [`../deployment-plan.md`](../../deployment-plan.md) working rules "Event and Place each get their own picture Property, labelled Picture", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "The gallery is one slot on the detail body", "A depiction is worded as a picture outside the Person's own copy"; [`macos-client-patterns.md` § 9](../../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S14-02-pictures-in-the-app.md`](../../prs/S14-02-pictures-in-the-app.md).
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../../.agents/skills/add-ui-component/SKILL.md)

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

Draw on the *Person Detail* project as a new turn "S14-D4 · Events list thumbnail". Add Events list frames only. Do not redraw the list's columns, sort or toolbar: only the row thumb changes.

---

## 1. Objective

A researcher scanning the Events list sees, on each Event that has a picture, that picture in the row thumb, the way the Persons list shows a Person's portrait.

**What changes.** The row thumb shows the Event's lead picture where it shows the Event mark today.

**What stays.** Everything else in the list: its columns (title with its subjects, the place line under it, ref, updated), sort, row height, selection, empty and loading states.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| The lead picture | The earliest picture saved to the Event's Property **Picture** (`event_picture`), unless the researcher chose another on the Event page (S14-D2). The row shows that lead. |
| Most events have none | Rows with a picture sit among rows with the mark. |
| The row thumb slot exists | Every Conclusion list draws a 44pt thumbnail at the row's leading edge through the List component; the Persons list fills it with the portrait, the Events list with the Event mark. |
| Two widths | Regular is a pane below 1600pt, Wide 1600pt and up. The list sits in the content column (up to 1180pt), centered, at both. |

### 2.1 What this board is not

- The Event page (S14-D2) or the Places list (S14-D5).
- A larger thumb, a hover preview, or a picture column.
- Choosing the thumbnail from the list.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S14-02** | EVL-1 … EVL-3 and EVL-W1: the row thumb at Regular and Wide, light and dark. |
| Stays out | Any other list change; the list's own redesign. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EVL-1 | **Row thumb.** An Event with a picture shows its lead picture in the row thumb, at the size and corner the Persons list uses for a portrait (the List's 44pt thumbnail slot). |
| EVL-2 | **No picture.** An Event with none shows the Event mark, as today. An image that cannot load shows the mark too. |
| EVL-3 | **Accessibility and dark mode.** The row's spoken label is unchanged, as on the Persons list when a portrait shows. Draw frame 1 in dark mode. |
| EVL-W1 | **Wide.** The list stays as today: its content column up to 1180pt, centered in the pane (`ConclusionListPage.page`); the row thumb keeps its size. (Corrected 2026-10-30: this line said the list fills the width, which is not today's layout.) |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Events list | The list in the content column, row thumb at the leading edge | The same column, centered | Nothing (today's layout) |

---

## 5. Suggested frames

Draw each as the pane only, with no app sidebar, at Regular (1102 × 900) unless it says Wide (2296 × 1440).

1. **Events list.** Six rows: two with pictures, four with the mark (EVL-1, EVL-2).
2. **Dark mode.** Frame 1 in dark mode (EVL-3).
3. **Wide.** Frame 1 at Wide (EVL-W1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Conclusion list page | Snowflake | Ship | `Features/Conclusions/ConclusionListPage.swift` | The one list page for Persons, Events and Places. Unchanged. |
| Row thumb | Component | Ship | `PVList` thumbnail slot, `CachedThumbnail` over `PVThumbnail` | EVL-1, EVL-2. The slot every Conclusion list already has; only the image changes. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A bigger thumb or a picture column | EVL-1: the Persons list's size. |
| A thumbnail menu on the row | Choosing happens on the Event page. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design link, the turn and frame numbers, and every change from this brief with the user's answer.
2. Refresh the [S14-02 plan](../../prs/S14-02-pictures-in-the-app.md) once all four S14-02 boards are agreed ([S14-D2](S14-D2-event-page-pictures.md) (Event page), [S14-D3](S14-D3-place-page-pictures.md) (Place page), [S14-D5](S14-D5-places-list-thumbnail.md) (Places list) and this one): copy the answers into its Decisions and **Design → macOS**, describe each frame in words, run `review-plan` again, and set it to `Planned`.
3. S14-02 archives this brief under `design/archive/` and records it in [`../completed.md`](../../completed.md).
