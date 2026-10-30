# S14-D2 — Event and Place pages: Pictures and the thumbnail

**Kind:** Claude Design board (**enhancement** of the Event and Place pages and of the Events and Places list rows)
**Spike:** Provenencia Spike 14 (odds and ends), batch 1 — Event and Place pictures
**View:** the Event and Place pages (one shared Conclusion page) and their list rows
**Implements later as:** PR **S14-02** (Pictures in the app)
**Board:** — (not yet drawn; the user records the link and answers here, see § 7)
**Depends on:** S13-10 (Spike 13, Planned): builds Page columns in the app and moves the Event and Place pages onto it, primary-only. S14-01 (landed [#615](https://github.com/mendahu/provenencia/pull/615)): an Event's and a Place's gallery and header lead come from the engine. The agreed S13-D1 board (*Person Detail* project): the Person page frame, its Gallery section and the Page columns arrangement the Event and Place pages also sit on. The agreed S14-D1 board (Page columns, the Wide width class).
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

Draw on the *Person Detail* project, beside the S13-D1 frames, as a new turn "S14-D2 · Event and Place pictures". Add Event and Place page frames and list-row frames. Do not redraw the Person page, and do not start a second Conclusion page: the Event and Place pages are the same page as the Person page with a different header and rows.

---

## 1. Objective

A researcher who draws a picture of a church on a Citation and saves it to a Place (Property "Picture") sees that picture on the Place: in a **Pictures** section on the page, in the header thumb, and on the Place's row in the Places list. With two or more pictures, they choose which one leads, as they choose a Person's portrait. The same holds for an Event.

**What changes.** The Event and Place pages gain a Pictures section below their rows, with the Person page's tiles. The header thumb shows the lead picture instead of the kind's mark. The thumbnail can be chosen from a tile and set back to the earliest from the header thumb. The Events and Places list rows show the lead picture in the row thumb, where they show the kind's mark today.

**What stays.** Every other part of the Event and Place pages: the header's title, date and place lines, ref and member count; the Date, Place, Names and Period rows with their Why; a Place's relationship sections. The page layout (Page columns, one centered primary column, no section switcher: Event and Place are not sectioned). The Person page, its "Likenesses" heading, "Portrait" Badge and "Use as portrait" wording. Every list's columns, sort and row height.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A picture is a crop | A rectangle drawn on a Citation's image and saved to the Property **Picture** (`event_picture` on an Event, `place_picture` on a Place). It is the same kind of value as a Person's likeness. |
| An Event or Place has 0 – many pictures | Most have none. A church may have five photos over a century; a wedding may have one. The gallery lists them earliest first. |
| The lead picture | The earliest picture leads unless the researcher chose another. The lead is the header thumb and the list-row thumb. A choice is a display preference, not evidence. |
| One mechanism with the Person | The Person page already does all of this with likenesses: 96pt tiles captioned with the Source title, a Badge on the lead tile, a ghost button on the others, a click opens the Citation, a right-click on the header thumb offers going back to the earliest, a success Toast after each choice. Event and Place reuse it with their own words. |
| Words | A Person's crop is a **likeness** and its lead the **portrait**. An Event's or Place's crop is a **picture** and its lead the **thumbnail**. The section is **Pictures**. |
| The kind's mark | With no picture, the header thumb and the list-row thumb show the Event or Place mark, as today. A tile whose image cannot load shows the same mark. |
| Event and Place pages are short | A header, then two rows (Event: Date and Place; Place: Names and Period), then a Place's relationship sections (the places it sits in or holds). The Pictures section is often the largest thing on the page. |
| Two widths | Regular is a pane below 1600pt, Wide 1600pt and up. The Event and Place pages sit on Page columns, primary-only and centered (S13-D1 § 4.1). The lists are full-width tables at both widths. |

### 2.1 What this board is not

- The Person page or its Gallery section (S13-D1, S13-D5).
- Sections or a switcher on the Event or Place page.
- A picture viewer or a larger preview (S13-D5 decides that for the Person; Event and Place follow later).
- Drawing a picture (the composer; S14-03).
- An empty Pictures state with a message: an Event or Place with no picture shows no Pictures section, as the Person page did before S13-D1.
- A Wide second column on the Event or Place page.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S14-02** | EPP-1 … EPP-9 and EPP-W1 … EPP-W2: the Pictures section, the header thumb, choosing a thumbnail, the earliest menu, the Toast, the list-row thumb, at Regular and Wide, light and dark. |
| Later boards | S13-D5 redraws the Person Gallery; Event and Place adopt it in their own redesign. |
| Stays out | Any engine change; Promote (its evidence sheet was removed by Spike 13's S13-23). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EPP-1 | **Pictures section.** Below the rows and above any relationship section (where the Person page draws its Likenesses today), a SectionHeader "Pictures" with the count as meta, then the Person page's tile grid: 96pt crops, adaptive columns, earliest first, each captioned with its Source title in the Person tile's caption style. Same spacing as the Person page's Likenesses grid. |
| EPP-2 | **No pictures, no section.** An Event or Place with no picture shows no Pictures section and no empty state. Draw it to show the page is as today. |
| EPP-3 | **Open a picture.** A click on a tile opens its Citation, as on the Person page. Keep the Person tile's hover and the tooltip "Open citation". |
| EPP-4 | **Header thumb.** The 120pt header thumb shows the lead picture, cropped as the Person header crops a portrait. With no picture, the kind's mark as today. |
| EPP-5 | **Choose the thumbnail.** With two or more pictures, the lead tile wears a Badge "Thumbnail" (accent, photo glyph, as the Person's "Portrait"), and every other tile a ghost small Button "Use as thumbnail". With one picture, that tile wears the Badge and there is no button. |
| EPP-6 | **Back to the earliest.** When a chosen picture leads (not the earliest), a right-click on the header thumb opens the native menu with "Use earliest picture". The menu's title reads "Event thumbnail" or "Place thumbnail", and its hint "Right-click for thumbnail options". When the earliest leads, there is no menu, as on the Person. |
| EPP-7 | **Toast.** After a choice, the success Toast "Thumbnail updated" with "Using the picture from <Source title> as the thumbnail." or "Using the earliest picture as the thumbnail.". A failed save shows the danger Toast with the error, as on the Person. |
| EPP-8 | **List rows.** In the Events and Places lists, the row thumb shows the lead picture, at the size and corner the Persons list uses for a portrait (the List's 44pt thumbnail slot). With no picture, the kind's mark as today. Nothing else in the row changes. |
| EPP-9 | **Accessibility and dark mode.** Each tile reads "Picture from <Source title>. Open citation"; the Badge and button read as on the Person, worded "thumbnail". The header thumb reads "Picture from <Source title>" when a picture leads. Draw the Place page frames and one list in dark mode. |
| EPP-W1 | **Pages at Wide.** The Event and Place pages stay primary-only and centered (WL-6), as S13-D1 decided for the Person Gallery (PPF-W3): the rows keep their Regular measure (rule 2), and the Pictures grid fills the primary column at the same 96pt tiles. The primary column is 1038pt wide in the Regular pane and at most 1180pt at Wide, so Wide gains about one tile per row; it is otherwise the Regular page, centered. No secondary column: the page has nothing supporting to put in it, and a full-width grid needs a Page columns mode that does not exist (a note for S13-D5, not this board). |
| EPP-W2 | **Lists at Wide.** The lists fill the width as today; the row thumb keeps its size. Show it on one Wide list frame. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Place page (and Event page, the same page) | Page columns, primary-only: header, rows, Pictures grid, relationship sections | The same centered primary column, up to 1180pt; the grid fills it (about one more tile per row) | Nothing |
| Places list (and Events list) | Full-width list, row thumb at the left | The same list across the width | Nothing (the list fills, rule 3) |

---

## 5. Suggested frames

Draw each as the pane only, with no app sidebar, at Regular (1102 × 900) unless it says Wide (2296 × 1440).

1. **Place with pictures.** A Place "St. Mary's Church, Halifax" (Names and Period rows, and a "Part of: Halifax" relationship section) with three pictures (Sources "Church photo, 1890", "Parish history, 1923", "Postcard, 1951"); the earliest leads: header thumb is the 1890 crop; the Pictures section sits between the rows and the relationship section; the 1890 tile wears the Thumbnail Badge, the two others "Use as thumbnail". Tiles, captions, Badge and button are the Person page's, as the S13-D1 board draws them in its Gallery frames 4i and 4j (EPP-1, EPP-4, EPP-5).
2. **Chosen thumbnail.** Frame 1 after "Use as thumbnail" on the 1951 tile: the header shows it, the Badge moved, the Toast showing (EPP-5, EPP-7).
3. **Earliest menu.** Frame 2 with the header thumb's right-click menu open: "Use earliest picture" (EPP-6).
4. **One picture.** An Event "Wedding of John Smith and Mary Brown, 1902" (header date and place lines as today; Date and Place rows) with one picture: Badge, no button (EPP-5).
5. **No picture.** The same Event with none: the mark in the header, no Pictures section (EPP-2, EPP-4).
6. **Places list.** Five rows: two with pictures, three with the mark (EPP-8).
7. **Events list.** The same, for Events (EPP-8).
8. **Dark mode.** Frames 1 and 6 in dark mode (EPP-9).
9. **Wide Place.** Frame 1 at Wide, with eight pictures: the centered primary column, the grid filling it (EPP-W1).
10. **Wide list.** Frame 6 at Wide (EPP-W2).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Shared Conclusion page | Snowflake | Ship | `Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailPage` | The one page for Person, Event and Place. Event and Place draw the gallery through the same slot the Person uses. Not a new page. |
| Gallery section and tiles | Snowflake | **Extend** | same file, today `ConclusionDetailLikenessSection` and `LikenessTileView` (renamed gallery in S14-02) | EPP-1, EPP-3, EPP-5. The Person's grid, worded by the page's kind: heading, Badge and button words, fallback mark. Not a second grid. |
| Section title | Component | Ship | `PVSectionHeader` | EPP-1. "Pictures" with the count as meta. |
| Tile crop and header thumb | Component | Ship | `CachedThumbnail` over `PVThumbnail` | EPP-1, EPP-4, EPP-8. 96pt tiles, 120pt header, the list's row size; the kind's mark as fallback. |
| Thumbnail Badge | Component | Ship | `PVBadge` accent with the photo glyph | EPP-5. As the Person's Portrait Badge. |
| Use as thumbnail | Component | Ship | `PVButton` ghost, small | EPP-5. As "Use as portrait". |
| Earliest menu | Component | Ship | the header thumb's right-click menu (`pvRightClickMenu`, `PVActionMenuItem`) | EPP-6. As the Person's. |
| Toast | Component | Ship | `PVToast`, drawn by the page's `vocabularyToastOverlay` | EPP-7. Success and danger tones. The kit's Toast. |
| List row thumb | Component | Ship | `PVList` thumbnail slot in `Features/Conclusions/ConclusionListPage.swift` | EPP-8. The slot every Conclusion list already has; only the image changes. |
| Page columns | Component | Ship (built in S13-10, a dependency of this board) | `DesignSystem/Components/PageColumns/` *(new)* | EPP-W1. Primary-only. Not drawn anew. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A second gallery grid or tile for Event and Place | One gallery slot serves all three kinds (working rule "The gallery is one slot on the detail body"). |
| A Pictures empty state | EPP-2: no pictures, no section. |
| New buttons in the header | The header stays display only; choosing happens on the tiles and the thumb's right-click menu, as on the Person. |
| A picture viewer or lightbox | Out (S13-D5 decides for the Person first). |
| A secondary column, or a stretched grid tile, at Wide | EPP-W1. |
| A new mark for a picture | The Event and Place marks already exist. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design link, the turn and frame numbers, and every change from this brief with the user's answer.
2. Refresh the [S14-02 plan](../prs/S14-02-pictures-in-the-app.md): copy the answers into its Decisions and **Design → macOS**, describe each frame in words, run `review-plan` again, and set it to `Planned`.
3. S14-02 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md).
