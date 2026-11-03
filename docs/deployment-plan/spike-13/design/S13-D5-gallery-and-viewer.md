# S13-D5 — Person page: the Gallery section and the likeness viewer

**Kind:** Claude Design board (**enhancement** of the agreed S13-D1 Person page frame: it fills the Gallery section, which S13-D1 drew as today's likeness grid moved in unchanged (PPF-9), and adds the likeness viewer, a new surface over the page)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Gallery section; the likeness viewer
**Implements later as:** PR **S13-14** (the Gallery section) and PR **S13-15** (the likeness viewer).
**Board:** not drawn yet. When agreed, record the Claude Design project, file, turn and frames here, with every answer to § 7.
**Depends on:** the agreed S13-D1 board (the frame, the header, the switcher; [archived brief](archive/S13-D1-person-page-frame.md), built by S13-10 [#634](https://github.com/mendahu/provenencia/pull/634)); S14-02's gallery slot and lead model (landed [#641](https://github.com/mendahu/provenencia/pull/641); the S14-D2 Pictures section, [archived brief](../../spike-14/design/archive/S14-D2-event-page-pictures.md)); the S13-D13 kit board (the provenance sheet as a kit recipe: the one-record sheet GA-4 opens and the record table LV-4 places; [brief](S13-D13-provenance-sheet-in-the-kit.md)); the agreed S13-D4 board (Page columns' full-width mode, [archived brief](archive/S13-D4-timeline-section.md), built by S13-13 [#691](https://github.com/mendahu/provenencia/pull/691)).
**Related:** [`../person-detail.md`](../person-detail.md) R-29 – R-32 (and R-1, R-2, R-13b); [`../deployment-plan.md`](../deployment-plan.md) S13-14 and S13-15 scope entries, working rule "Sections are places"; Spike 14 working rules "The gallery layer is named gallery, not likeness", "An Event or Place chooses its thumbnail as a Person chooses its portrait" ([`../../spike-14/deployment-plan.md`](../../spike-14/deployment-plan.md#working-rules)); [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide).
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.agents/skills/add-ui-component/SKILL.md)
**Review:** [`S13-D5-gallery-and-viewer.review.md`](S13-D5-gallery-and-viewer.review.md) — round 1 Fits with changes / Ready with changes, every finding Fixed

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

Extend the S13-D1 frames. The header, the switcher and the other four sections stay exactly as agreed (S13-D1, S13-D3, S13-D4); only the Gallery section's content changes, and the viewer is added over the page. Draw on the S13-D1 Person page frame, with the switcher on **Gallery**. Do not start a second Person page. The Event and Place pages' Pictures section (S14-D2, S14-D3) is not on this board and keeps its look.

---

## 1. Objective

A researcher opens a Person's Gallery section and sees every likeness of the person at a size where faces can be told apart, each with where it came from. A click on a likeness opens it in a **viewer** over the page, at the size the window allows, and the researcher steps through the gallery likeness by likeness. From the gallery and from the viewer a likeness can be made the portrait, and its provenance read, with links on to its Source, Citation and Observation.

**What changes.** The Gallery section's grid is redrawn: tile size, caption, where Set as portrait, the Portrait mark and the provenance control sit, and a visible way back to the earliest likeness. A click on a tile now opens the viewer (R-30). Each tile gets the **provenance control** the Vitals rows have, opening the same provenance sheet filled with that one likeness's record; the sheet's links reach the Source, the Citation and the Observation (R-32, R-13b; user, 2026-11-03). The tile has no direct Open citation. The viewer is new: the large image, the likeness's provenance record right under it, previous and next, a position counter, and Set as portrait.

**What stays.** The header, including the thumb's right-click "Use earliest likeness" menu, which the Event and Place headers share (user, 2026-11-03). The switcher, Vitals, Family, Timeline and the Sources empty place. The gallery's order (earliest recorded first, from the engine). The gender colors on the header thumb (S10-D12). The empty Gallery's EmptyState, unless the board rewords it (GA-8). The Event and Place pages and their Pictures sections (user, 2026-11-03: Person only).

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A likeness (R-29) | A rectangle a researcher drew on one page of a Citation's image and recorded as a **likeness** Observation of a Subject. The app stores the cut-out as its own PNG, at most 2048 pixels on its long edge; many are small (a face from a group photograph, 120 – 400 pixels). The Gallery lists every likeness of every accepted member of the person. |
| Canonical only (R-2) | Every likeness belongs to an accepted member. Nothing unpromoted is drawn. |
| The order comes from the engine | Earliest recorded first. The board does not reorder or group the tiles. |
| What a likeness knows | Its image; its **Source** title (else the Observation's ref, in mono); the **citation line** the provenance sheet already shows ("Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23": Source, Artifact, Citation, page, each left out when empty); the Observation's ref ("OBS-7Q2KD"). No date of its own: the Source date was descoped (2026-10-31). |
| The portrait (R-31) | Exactly one likeness leads: the one the researcher chose as the portrait, else the earliest. The header thumb shows it. The leading tile carries the Portrait mark; every other tile offers **Set as portrait**. Choosing writes at once (no confirm) and shows a Toast ("Portrait updated" · "The likeness from Hamburger Adressbuch 1861 is now the portrait."). |
| Back to the earliest | While a chosen likeness leads and it is not the earliest, the researcher can go back to the earliest. Today that is only the header thumb's right-click menu. This board adds a **visible** control for it in the Gallery (user, 2026-11-03); the header menu stays. With no choice, or when the chosen one is the earliest, the control is absent. |
| Provenance of a likeness (R-13b, R-32; user, 2026-11-03) | The likeness is one record of the Person's Likeness field, the field the Vitals section lists as "Likeness" with its own provenance sheet (S13-D2). From a tile, that sheet opens with **only this likeness's record**: its citation line, what the auto-reconciler did with it (kept, weak), its Observation and Subject refs, and the sheet's links to the Source page, the Citation's composer at the likeness, and the Observation. A link closes the sheet first; Back returns to the Gallery. There is no other Open citation control on the tile, in its menu or in the viewer. |
| The viewer (R-30; user, 2026-11-03) | Click (or Return / Space on a focused tile) opens the viewer on that likeness. It steps through the gallery in its order with previous and next and the ← → keys, and shows the position ("3 of 7"). **No autoplay**: no timer, no play button. It stops at the ends (no wrap). Escape or a click outside closes it and puts focus back on the tile of the likeness last shown. It is not a place in the navigation history: Back does not reopen it, and a restored window opens on the Gallery with the viewer closed. |
| Image size in the viewer | The image is shown whole (fit, never cropped) at the largest size the viewer allows, but never above its own pixel size at 2× (a 300-pixel face is drawn at most 150pt wide, not blown up to fill a 1440pt pane). No zoom or pan. The board draws a small likeness and a large one. |
| The read and refresh | The gallery arrives with the Person page's detail read; the section has no read of its own. Choosing a portrait refreshes the page; the gallery stays drawn and "Refreshing" shows at the end of the switcher row (S13-D1 PPF-8). |
| A likeness can go while the viewer is open | Another window can delete a Citation's likeness. When the viewer's likeness leaves the gallery, the viewer shows the likeness that took its place (the next, else the previous); when the gallery empties, the viewer closes. |
| An image can fail to load | A missing or unreadable file shows the Person mark on the sunken Thumbnail, in the gender colors, in the tile and in the viewer; captions and actions stay. |
| Sizes | Typical: 1 – 6 likenesses. Large: 40 (a person photographed in a long run of yearbooks). Source titles can be long ("Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten"). |

### 2.1 What this board is not

- Autoplay, a slideshow timer or transitions (user, 2026-11-03).
- Zoom, pan, rotate, crop, or redrawing the region (S14-03 owns drawing).
- Comparing two likenesses side by side.
- Editing: no delete, reorder, caption editing or tagging, and no context menu actions beyond Set as portrait and the provenance sheet.
- A direct Open citation control (user, 2026-11-03: the provenance sheet's Citation link replaces it).
- A redesign of the provenance sheet: it is S13-D2's sheet, with one record.
- Grouping by Source, a filter, or a search.
- The Event and Place Pictures sections, or a viewer for them (user, 2026-11-03: Person only; they adopt this gallery when those pages get sections).
- A change to the header, its thumb or its right-click menu.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-14** | GA-1 … GA-10 and GA-W1 … GA-W3: the Gallery section on the Person page, the tile, its caption, the Portrait mark, Set as portrait, the provenance control and its sheet, the visible way back to the earliest, the empty and refreshing states, keyboard and VoiceOver, light and dark, at Regular and Wide. Until S13-15 a click on a tile's image opens its provenance sheet, as the provenance control does. |
| Ships in **S13-15** | LV-1 … LV-9 and LV-W1: the viewer, opened by a click on a tile, with stepping, the counter, the likeness's provenance record under the image, Set as portrait, closing, keyboard and VoiceOver, light and dark, at Regular and Wide. The tile's click moves from Open Citation to the viewer. |
| Later boards | The Event and Place pages adopt this gallery and the viewer when they get sections (R-1). |
| Stays out | Any change to the header, switcher or other sections; the Event and Place pages. |

---

## 4. Requirements

### 4.1 Gallery section

| ID | Requirement |
| --- | --- |
| GA-1 | **The tile.** One tile per likeness: the image on a Thumbnail, cover-cropped into a square, with its caption (GA-2) and its actions (GA-3, GA-4). Today's tile is a 96pt crop in a 120pt column. Pick the tile's size so a face in a 300-pixel crop is recognisable (recommended: 160pt); say why. Every tile has the same size, so the grid aligns. |
| GA-2 | **The caption.** The Source title, up to two lines then cut at its tail, with the full citation line (§ 2) in the tooltip and VoiceOver. A title that is a ref shows in mono. Decide whether the tile shows more of the citation line (the page, "p. 23") on a second muted line; say why. |
| GA-3 | **The Portrait mark and Set as portrait (R-31).** The leading tile carries the Portrait Badge (accent, photo icon, as today); every other tile offers Set as portrait, a ghost Button. Decide where on the tile they sit (under the caption, as today; or over the image's corner); a control that only appears on hover is not enough on its own (keyboard and VoiceOver reach it). While a choice saves, every Set as portrait is disabled. |
| GA-4 | **The provenance control (R-13b, R-32; user, 2026-11-03).** Each tile has the Vitals rows' provenance control: the small info IconButton with the tooltip "Provenance of Likeness from Hamburger Adressbuch 1861". It opens the S13-D2 provenance sheet with only this likeness's record (§ 2). Decide where it sits on the tile beside Set as portrait. The tile's context menu offers Set as portrait and the provenance sheet. No Open citation control. Draw the sheet open from tile 2 (frame 2b) by instancing the kit's **Provenance sheet** recipe in its one-record variant (S13-D13 PSK-4); do not draw a sheet of your own. |
| GA-5 | **Back to the earliest.** While a chosen likeness leads and it is not the earliest, the Gallery shows a visible control to go back to the earliest (recommended: a ghost Button "Use earliest likeness" at the end of the section's header line); otherwise none. It writes at once with the Toast "Portrait updated" · "The earliest likeness is the portrait again." The header thumb's menu stays (§ 1). |
| GA-6 | **The section's heading.** Today the section repeats a "Likenesses" heading and a count under the switcher. Vitals, Family and Timeline have no heading of their own. Decide whether Gallery keeps a heading line (it carries the count and GA-5's control) or drops it and moves the control elsewhere; say why. |
| GA-7 | **The grid.** Tiles wrap in an adaptive grid at the tile's size; the grid starts at the section's leading edge. Every likeness is shown: no fold at the typical and large sizes (§ 2). |
| GA-8 | **No likenesses.** The kit EmptyState with the Gallery glyph (photo): "No likenesses yet" and "A likeness is drawn on a Citation’s image — draw one there and it appears here." (today's copy), or the board's rewording. |
| GA-9 | **Keyboard and VoiceOver.** Tab moves through each tile's image, then its actions, in reading order, then the next tile; Return or Space on the image opens the viewer (S13-15; until then the provenance sheet). VoiceOver reads each tile as a group: "Likeness from Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23, portrait" for the leading tile; the actions are buttons with the Source in their label ("Set the likeness from Hamburger Adressbuch 1861 as portrait"). The count is read with the heading, if GA-6 keeps one. |
| GA-10 | **Dark mode.** Frames 1, 3 and 5 in dark: tiles, the sunken fallback with the Person mark in its gender colors, the Badge, the buttons. |
| GA-W1 | **The grid at Wide.** Rule 3: the grid fills the pane between the page gutters (Page columns' full-width mode, as the Timeline section; the header keeps its band) and adds tiles per row; tiles keep their size. No secondary column: the section has nothing supporting to put in it. Captions keep their width. |
| GA-W2 | **Crossing 1600 keeps the place (rule 5).** Resizing keeps the section, the scroll position and the focused tile. A note on the Wide frame. |
| GA-W3 | **No likenesses at Wide.** The EmptyState stays centered in the primary column; the section does not stretch it. |

### 4.2 Likeness viewer

| ID | Requirement |
| --- | --- |
| LV-1 | **Opening.** A click on a tile's image, or Return / Space with the image focused, opens the viewer on that likeness. Pick its form and say why: (A) a large kit sheet over the page (the sheet presenter every app sheet uses: a click outside and Escape close it); (B) an overlay that dims and covers the pane, with the image in its middle; (C) the tile expanding in place inside the grid. **Recommended: (A)**, the kit sheet (Panel inside the sheet presenter), which already owns Escape, a click outside and focus. (B) would need a new page-level presenter the kit does not have; draw it only as an alternative frame, marked as that cost. (C) moves every tile below it. |
| LV-2 | **The image.** The whole likeness, fit inside the viewer's image area, centered, on the sunken surface; never above its pixel size at 2× (§ 2). While the full image loads the tile's image shows, scaled; a missing image shows the Person mark (§ 2). |
| LV-3 | **Stepping.** Previous and next IconButtons beside the image or under it, and the ← → keys; previous is disabled on the first likeness, next on the last (no wrap). A position counter ("3 of 7") in mono. A gallery of one shows neither button nor counter. |
| LV-4 | **The record under the image (user, 2026-11-03).** Right under the image, the likeness's provenance record, drawn by instancing the kit Provenance sheet's **Record row** on its own (S13-D13 PSK-6): the citation line, the outcome, the Observation and Subject refs, and its links to the Source, the Citation and the Observation. A link closes the viewer first. This replaces a separate caption and any Open citation button: the viewer shows the citation itself. Long lines wrap; nothing is cut. |
| LV-5 | **Actions.** Set as portrait (a Button) on a likeness that does not lead; the Portrait Badge on the one that does; a choice made here updates the header thumb and the grid behind the viewer, and the viewer stays open on the same likeness. Decide where it sits (the footer, recommended). The record's links (LV-4) are the only way on to the Citation. No "Use earliest" in the viewer: GA-5 owns it. |
| LV-6 | **Closing.** Escape, a click outside (A, B) and a visible close IconButton. Focus returns to the tile of the likeness last shown, scrolled into view. |
| LV-7 | **A likeness leaves while open.** As § 2: the viewer shows the likeness that took its place, else closes when none is left. Draw it as a note on frame 7. |
| LV-8 | **Keyboard and VoiceOver.** Focus starts on the image area, which reads "Likeness 3 of 7, from Hamburger Adressbuch 1861, Volume 3, Entry 214, p. 23"; ← → step and VoiceOver reads the new likeness; Tab moves through previous, next, the record's links, Set as portrait and close; the record row reads as on the provenance sheet. |
| LV-9 | **Dark mode.** Frames 6 and 8 in dark. |
| LV-W1 | **The viewer at Wide.** The image area grows with the pane (it is a canvas, rule 3), still capped by the image's own size (LV-2); the caption keeps its measure. Draw the Wide frame with a 2048-pixel likeness and a 300-pixel one. |

### 4.3 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Gallery section | Tiles wrap in the primary column | The grid fills the pane between the gutters (full-width mode); more tiles per row; tiles keep their size (GA-W1) | Nothing moves; the primary widens |
| Likeness viewer | The viewer at the size LV-1's form gives in a 1102 × 900 pane | A larger image area, capped by the image's size; the caption keeps its measure (LV-W1) | Nothing moves |
| No likenesses | EmptyState in the primary column | Centered primary-only (GA-W3) | Nothing |

---

## 5. Suggested frames

Each frame is the S13-D1 Person page pane only (no app sidebar), with the switcher on **Gallery**, at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). The Wide frames sit beside their Regular frame.

Sample person: **Johann Weber** (man), as on the Family and Timeline boards. Seven likenesses, earliest recorded first; number 4 is the chosen portrait:

| # | Source title | Citation line's other parts | Pixel size |
| --- | --- | --- | --- |
| 1 | Konfirmandenbuch St. Ansgarii 1831 | Register 2, Entry 18, p. 41 | 240 × 300 |
| 2 | Hamburger Adressbuch 1861 | Volume 3, Entry 214, p. 23 | 300 × 380 |
| 3 | Photograph: the Weber family, 1862 | Print, Johann (back row) | 180 × 220 |
| 4 | Carte de visite, Atelier Strumper, Hamburg | Front | 1400 × 2048 (the portrait) |
| 5 | Verzeichniß der Mitglieder des Vereins für Hamburgische Geschichte, nebst einem Anhang über die Statuten | Page 7, image 12 | 320 × 400 |
| 6 | OBS-7Q2KD (a Source with no title) | — | 260 × 260 |
| 7 | Hamburger Nachrichten, obituary 1880 | Issue 221, p. 4 | missing image |

1. **Gallery.** Johann's Gallery with the seven tiles: the Portrait Badge on 4, Set as portrait on the rest, GA-5's control, the long title on 5 cut at two lines, the ref title on 6 in mono, the fallback on 7 (GA-1 – GA-7).
2b. **A tile's provenance.** The provenance sheet open from tile 2, with its one record (GA-4).
2. **Tile sheet.** One tile in each state: the portrait; another with Set as portrait; hover and keyboard focus on the image and on each action; saving (buttons disabled); the context menu open (GA-1 – GA-4, GA-9).
3. **Earliest leads.** The gallery with no chosen portrait: the Badge on 1, no GA-5 control (GA-5).
4. **Large gallery.** 40 likenesses (GA-7).
5. **No likenesses.** The EmptyState (GA-8).
6. **Viewer.** The viewer open on 2 (a small likeness: drawn at its 2× size, not stretched): counter "2 of 7", previous and next, the provenance record under the image, Set as portrait, close (LV-1 – LV-6). Beside it the viewer on 4 (a large likeness filling the image area, the Portrait Badge in place of Set as portrait).
7. **Viewer ends and edges.** On 1 (previous disabled), on 7 (next disabled, missing image), a gallery of one (no stepping), and a note for a likeness that leaves while shown (LV-3, LV-7).
8. **Viewer, keyboard.** Focus on next after a → press; the VoiceOver reading of the image area as a note (LV-8).
9. **Dark mode.** Frames 1, 3, 5, 6 and 8 in dark (GA-10, LV-9).
10. **Wide gallery.** Frame 1 at Wide, full width, with the rule-5 note (GA-W1, GA-W2).
10b. **Wide large and empty.** Frame 4 (40 likenesses) at Wide, and frame 5 (no likenesses) at Wide (GA-W1, GA-W3).
11. **Wide viewer.** The viewer at Wide on 4 and on 2 (LV-W1).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, header, switcher | Snowflake / component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `PVPageColumns`, `PVSegmentedControl` | As S13-D1. Gallery fills the primary column; full-width mode at Wide (GA-W1), as Timeline. The header's thumb menu is unchanged. |
| Gallery section | Snowflake | **Rethink** | `Features/Conclusions/ConclusionDetailPage.swift` `ConclusionDetailGallerySection` today | GA-1 – GA-9. Today's section stays for the Event and Place Pictures (they keep their look); the Person's Gallery is the redrawn one. It reads the kind-blind gallery slot (`ConclusionDetailGallery`), so Event and Place can adopt it later. |
| Likeness tile | Snowflake | **Rethink** | `GalleryTileView` in `ConclusionDetailPage.swift` | GA-1, GA-2. The image on `CachedThumbnail` over `PVThumbnail`, the caption in caption type. |
| Likeness image | Component | Ship | `CachedThumbnail` / `PVThumbnail`, fallback mark `subjectPerson` in the gender colors | GA-1, LV-2. No new placeholder artwork. |
| Portrait mark | Component | Ship | `PVBadge` (accent, `photo` icon) | GA-3, LV-5. Today's Badge. |
| Set as portrait, Use earliest likeness | Component | Ship | `PVButton` (ghost, `sm`) | GA-3, GA-5, LV-5. Not a new action control. |
| Provenance control | Component | Ship | `PVIconButton(.info)`, as `ReconciledValueRow`'s provenance slot | GA-4. The Vitals rows' control, same glyph and tooltip wording. |
| Provenance sheet | Recipe | Ship | the kit's Provenance sheet (S13-D13; `DesignSystem/Recipes/ProvenanceSheet/` *(new)* after S13-36) | GA-4. Its one-record variant (PSK-4). |
| Provenance record row | Recipe | Ship | the same recipe's Record row (S13-D13 PSK-6) | LV-4. One row, placed under the viewer's image, without the table or the sheet. |
| Tile context menu | Component | Ship | `PVContextMenu` / `pvRightClickMenu` with `PVActionMenuItem` | GA-4. Set as portrait and the provenance sheet only. |
| Tile image button | Component | Ship | `PVFocusRingButtonStyle`; `pvActivatesOnReturn(when:)` (`DesignSystem/Components/ReturnActivation/`) | GA-9, LV-1. The kit focus ring, and Return opens it (a plain button does not on macOS), as the Timeline title does. |
| Grid | Component | Ship | an adaptive `LazyVGrid` at the tile's size, as today | GA-7, GA-W1. |
| Section heading line | Component | Ship | `PVSectionHeader` (title, meta count, trailing control) | GA-6. If the board keeps a heading. |
| No likenesses | Component | Ship | `PVEmptyState` | GA-8. The Gallery glyph `photo`. |
| Portrait Toast | Component | Ship | `PVToast` through the page's lead model | GA-3, GA-5, LV-5. Today's Toast and copy. |
| Viewer panel | Component | Ship | `PVPanel` (title, subtitle, title accessory, footer) inside `pvSheet` | LV-1, LV-4 – LV-6. The provenance sheet's body: the title row holds the close IconButton, the footer the actions. |
| Viewer image surface | Component | Ship | the sunken surface token (`surfaceSunken`) | LV-2. No new token. |
| Likeness viewer | Snowflake | **New** | `Features/Conclusions/` *(new file, named in the S13-15 plan)* | LV-1 – LV-8. Presented through the kit sheet presenter (`pvSheet`) for form (A); (B) would be a new page-level presenter. Not the Artifact viewer (no zoom, no region tools). A feature snowflake, not a kit primitive; its known second call site is the Event and Place pages when they adopt sections (R-1). |
| Previous, next, close | Component | Ship | `PVIconButton` (`chevronBack`, `chevronForward`, `dismiss`) | LV-3, LV-6. |
| Position counter, Observation ref | Component | Ship | mono caption text | LV-3, LV-4. |
| Citation line | Snowflake | Ship | `CitationLineDisplay` (the provenance sheet's citation line) | GA-2's tooltip and VoiceOver. One wording of a citation for the page. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Autoplay, a timer, transitions | User, 2026-11-03 (§ 2.1). |
| Zoom, pan, rotate, crop or redraw | § 2.1; the Artifact viewer and S14-03 own those. |
| A new kit primitive (a lightbox, a carousel, a gallery grid) | One call site. The viewer is a feature snowflake on the kit sheet or an overlay; a second caller (Event, Place) promotes it later. |
| A viewer or the new grid on the Event and Place pages | User, 2026-11-03: Person only (R-1). |
| A change to the header thumb or its right-click menu | User, 2026-11-03: the menu stays; Event and Place share it. |
| Delete, reorder, caption editing | This spike is view only. |
| A direct Open citation control, or a second sheet over the viewer | User, 2026-11-03: the provenance sheet and the viewer's record row link to the Citation. |
| A new color token or card style | Thumbnail, Badge and Button cover it. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames; the tile size and caption (GA-1, GA-2); where the Portrait mark, Set as portrait and the provenance control sit, and the one-record sheet's title and header (GA-3, GA-4); GA-5's control and its place; whether the heading stays (GA-6); any empty-state rewording (GA-8); the viewer's form (LV-1); its image area's size at Regular and Wide (LV-2, LV-W1); how the record row sits under the image and where Set as portrait sits (LV-4, LV-5); and every other change from this brief with the user's answer.
2. Refresh the [S13-14 plan](../prs/S13-14-gallery-section.md) and the [S13-15 plan](../prs/S13-15-likeness-viewer.md): copy the answers into their Decisions and **Design → macOS**, describe every frame in words, close their Open questions, run `review-plan` again, and set them to `Planned`.
3. S13-15 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md) (S13-14 lands first and leaves the brief open for S13-15).
