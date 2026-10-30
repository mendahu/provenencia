# S14-D2 — Event page: Pictures and the thumbnail

**Kind:** Claude Design board (**enhancement** of the Event page)
**Spike:** Provenencia Spike 14 (odds and ends), batch 1 — Event and Place pictures
**View:** the Event page (`EventDetailView`, on the shared Conclusion page)
**Implements later as:** PR **S14-02** (Pictures in the app), with [S14-D3](S14-D3-place-page-pictures.md) (Place page), [S14-D4](S14-D4-events-list-thumbnail.md) (Events list), [S14-D5](S14-D5-places-list-thumbnail.md) (Places list)
**Board:** — (not yet drawn; the user records the link and answers here, see § 7)
**Depends on:** S13-10 (Spike 13, Planned): builds Page columns in the app and moves the Event page onto it, primary-only. S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615)): an Event's gallery and header lead come from the engine. The agreed S13-D1 board (*Person Detail* project): the Person page frame and its Gallery frames 4i and 4j, whose tiles, captions, Badge and button this board reuses. The agreed S14-D1 board (Page columns, the Wide width class).
**Related:** [`../event-place-pictures.md`](../event-place-pictures.md) § Thumbnail parity, § What changes; [`../deployment-plan.md`](../deployment-plan.md) working rules "Event and Place each get their own picture Property, labelled Picture", "An Event or Place chooses its thumbnail as a Person chooses its portrait", "The gallery is one slot on the detail body", "A depiction is worded as a picture outside the Person's own copy"; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S14-02-pictures-in-the-app.md`](../prs/S14-02-pictures-in-the-app.md).
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

Draw on the *Person Detail* project, beside the S13-D1 frames, as a new turn "S14-D2 · Event page pictures". Add Event page frames only. Do not redraw the Person page, and do not start a second Conclusion page: the Event page is the same page as the Person page with its own header and rows.

---

## 1. Objective

A researcher who draws a picture on a Citation and saves it to an Event (Property "Picture", key `event_picture`) sees it on the Event page: in a **Pictures** section and in the header thumb. With two or more pictures, they choose which one leads, as they choose a Person's portrait.

**What changes.** The Event page gains a Pictures section below its rows, with the Person page's tiles. The header thumb shows the lead picture instead of the Event mark. The thumbnail can be chosen from a tile and set back to the earliest from the header thumb.

**What stays.** Every other part of the Event page: the header's title, date and place lines, ref and member count; the Date and Place rows with their Why. The page layout (Page columns, one centered primary column, no section switcher). The Person page and its wording.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A picture is a crop | A rectangle drawn on a Citation's image and saved to the Property **Picture** (`event_picture`). It is the same kind of value as a Person's likeness. |
| Events have 0 – many pictures | Most have none. A wedding may have one; a battle or a parade may have several. The gallery lists them earliest first. |
| The lead picture | The earliest picture leads unless the researcher chose another. The lead is the header thumb and the Events list's row thumb (S14-D4). A choice is a display preference, not evidence. |
| One mechanism with the Person | The Person page already does this with likenesses: 96pt tiles captioned with the Source title, a Badge on the lead tile, a ghost button on the others, a click opens the Citation, a right-click on the header thumb offers going back to the earliest, a success Toast after each choice. The Event page reuses it with its own words. |
| Words | A Person's crop is a **likeness** and its lead the **portrait**. An Event's crop is a **picture** and its lead the **thumbnail**. The section is **Pictures**. |
| The Event mark | With no picture, the header thumb shows the Event mark, as today. A tile whose image cannot load shows the same mark. |
| The page is short | A header (title, date and place lines, ref, member count), then two rows: Date and Place. The Pictures section is often the largest thing on the page. |
| Two widths | Regular is a pane below 1600pt, Wide 1600pt and up. The Event page sits on Page columns, primary-only and centered (S13-D1 § 4.1). |

### 2.1 What this board is not

- The Person page or its Gallery section (S13-D1, S13-D5).
- The Place page (S14-D3) or the lists (S14-D4, S14-D5).
- Sections or a switcher on the Event page.
- A picture viewer or a larger preview (S13-D5 decides that for the Person first).
- Drawing a picture (the composer; S14-03).
- An empty Pictures state: an Event with no picture shows no Pictures section.
- A Wide second column.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S14-02** | EVP-1 … EVP-8 and EVP-W1: the Pictures section, the header thumb, choosing a thumbnail, the earliest menu, the Toast, at Regular and Wide, light and dark. |
| Later boards | S13-D5 redraws the Person Gallery; the Event page adopts it in its own redesign. |
| Stays out | Any engine change. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EVP-1 | **Pictures section.** Below the rows (where the Person page draws its Likenesses today), a SectionHeader "Pictures" with the count as meta, then the Person page's tile grid: 96pt crops, adaptive columns, earliest first, each captioned with its Source title in the Person tile's caption style. Same spacing as the Person page's Likenesses grid. |
| EVP-2 | **No pictures, no section.** An Event with no picture shows no Pictures section and no empty state. Draw it to show the page is as today. |
| EVP-3 | **Open a picture.** A click on a tile opens its Citation, as on the Person page. Keep the Person tile's hover and the tooltip "Open citation". |
| EVP-4 | **Header thumb.** The 120pt header thumb shows the lead picture, cropped as the Person header crops a portrait. With no picture, the Event mark as today. |
| EVP-5 | **Choose the thumbnail.** With two or more pictures, the lead tile wears a Badge "Thumbnail" (accent, photo glyph, as the Person's "Portrait"), and every other tile a ghost small Button "Use as thumbnail". With one picture, that tile wears the Badge and there is no button. |
| EVP-6 | **Back to the earliest.** When a chosen picture leads (not the earliest), a right-click on the header thumb opens the native menu with "Use earliest picture". The menu's title reads "Event thumbnail", and its hint "Right-click for thumbnail options". When the earliest leads, there is no menu, as on the Person. |
| EVP-7 | **Toast.** After a choice, the success Toast "Thumbnail updated" with "Using the picture from <Source title> as the thumbnail." or "Using the earliest picture as the thumbnail.". A failed save shows the danger Toast with the error, as on the Person. |
| EVP-8 | **Accessibility and dark mode.** Each tile reads "Picture from <Source title>. Open citation"; the Badge and button read as on the Person, worded "thumbnail". The header thumb reads "Picture from <Source title>" when a picture leads. Draw frame 1 in dark mode. |
| EVP-W1 | **Wide.** The page stays primary-only and centered (WL-6), as S13-D1 decided for the Person Gallery (PPF-W3): the rows keep their Regular measure (rule 2), and the Pictures grid fills the primary column at the same 96pt tiles. The primary column is 1038pt wide in the Regular pane and at most 1180pt at Wide, so Wide gains about one tile per row; it is otherwise the Regular page, centered. No secondary column: the page has nothing supporting to put in it, and a full-width grid needs a Page columns mode that does not exist (a note for S13-D5, not this board). |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Event page | Page columns, primary-only: header, rows, Pictures grid | The same centered primary column, up to 1180pt; the grid fills it (about one more tile per row) | Nothing |

---

## 5. Suggested frames

Draw each as the pane only, with no app sidebar, at Regular (1102 × 900) unless it says Wide (2296 × 1440).

1. **Event with pictures.** An Event "Wedding of John Smith and Mary Brown, 1902" (header date and place lines; Date and Place rows) with three pictures (Sources "Wedding photograph, 1902", "Family album, 1920", "Newspaper clipping, 1902"); the earliest leads: header thumb is its crop; the Pictures section under the rows; the lead tile wears the Thumbnail Badge, the two others "Use as thumbnail". Tiles, captions, Badge and button are the Person page's, as the S13-D1 board draws them in its Gallery frames 4i and 4j (EVP-1, EVP-4, EVP-5).
2. **Chosen thumbnail.** Frame 1 after "Use as thumbnail" on another tile: the header shows it, the Badge moved, the Toast showing (EVP-5, EVP-7).
3. **Earliest menu.** Frame 2 with the header thumb's right-click menu open: "Use earliest picture" (EVP-6).
4. **One picture.** The same Event with one picture: Badge, no button (EVP-5).
5. **No picture.** The same Event with none: the mark in the header, no Pictures section (EVP-2, EVP-4).
6. **Dark mode.** Frame 1 in dark mode (EVP-8).
7. **Wide.** Frame 1 at Wide, with eight pictures: the centered primary column, the grid filling it (EVP-W1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Shared Conclusion page | Snowflake | Ship | `Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | The one page for Person, Event and Place. The Event page draws the gallery through the same slot the Person uses. Not a new page. |
| Gallery section and tiles | Snowflake | **Extend** | same file, today `ConclusionDetailLikenessSection` and `LikenessTileView` (renamed gallery in S14-02) | EVP-1, EVP-3, EVP-5. The Person's grid, worded by the page's kind: heading, Badge and button words, fallback mark. Not a second grid. |
| Section title | Component | Ship | `PVSectionHeader` | EVP-1. "Pictures" with the count as meta. |
| Tile crop and header thumb | Component | Ship | `CachedThumbnail` over `PVThumbnail` | EVP-1, EVP-4. 96pt tiles, 120pt header; the Event mark as fallback. |
| Thumbnail Badge | Component | Ship | `PVBadge` accent with the photo glyph | EVP-5. As the Person's Portrait Badge. |
| Use as thumbnail | Component | Ship | `PVButton` ghost, small | EVP-5. As "Use as portrait". |
| Earliest menu | Component | Ship | the header thumb's right-click menu (`pvRightClickMenu`, `PVActionMenuItem`) | EVP-6. As the Person's. |
| Toast | Component | Ship | `PVToast`, drawn by the page's `vocabularyToastOverlay` | EVP-7. Success and danger tones. The kit's Toast. |
| Page columns | Component | Ship (built in S13-10, a dependency of this board) | `DesignSystem/Components/PageColumns/` *(new)* | EVP-W1. Primary-only. Not drawn anew. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A second gallery grid or tile for the Event page | One gallery slot serves all three kinds (working rule "The gallery is one slot on the detail body"). |
| A Pictures empty state | EVP-2: no pictures, no section. |
| New buttons in the header | The header stays display only; choosing happens on the tiles and the thumb's right-click menu, as on the Person. |
| A picture viewer or lightbox | Out (S13-D5 decides for the Person first). |
| A secondary column, or a stretched grid tile, at Wide | EVP-W1. |
| A new mark for a picture | The Event mark already exists. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design link, the turn and frame numbers, and every change from this brief with the user's answer.
2. Refresh the [S14-02 plan](../prs/S14-02-pictures-in-the-app.md) once all four S14-02 boards are agreed ([S14-D3](S14-D3-place-page-pictures.md) (Place page), [S14-D4](S14-D4-events-list-thumbnail.md) (Events list), [S14-D5](S14-D5-places-list-thumbnail.md) (Places list) and this one): copy the answers into its Decisions and **Design → macOS**, describe each frame in words, run `review-plan` again, and set it to `Planned`.
3. S14-02 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
