# S13-D3 — Person page: the Family section and the person tile

**Kind:** Claude Design board (**enhancement** of the agreed S13-D1 Person page frame: it fills the Family section, which S13-D1 drew as an empty place)
**Spike:** Provenencia Spike 13 (Person detail page)
**View:** Person detail, Family section; the shared person tile
**Implements later as:** PR **S13-12** (the Family section and the tile). S13-D4 (Timeline, S13-13) reuses this board's tile for event participants.
**Board:** — (not drawn yet). Draw it in the Claude Design project [*Person Detail*](https://claude.ai/design/p/c8660cdb-2c61-4c30-bc39-b2411036e6c0?via=share), file *Person Detail.dc.html*, as a new turn after S13-D2's.
**Depends on:** the agreed S13-D1 board (the frame, the header, the switcher; agreed 2026-10-29, [archived brief](archive/S13-D1-person-page-frame.md)); the agreed S10-D12 Person board (the gender treatment on the thumb); S13-04's family read (branch `s13-04-person-family-read`, the data this board draws).
**Related:** [`../person-detail.md`](../person-detail.md) R-2, R-3, R-16 – R-22; [`../deployment-plan.md`](../deployment-plan.md) working rules "Family groups in Go", "One person tile, built once", "Colors reuse `PersonGenderTreatment`", "A family tile opens the relative on Family"; [`macos-client-patterns.md` § 9](../../../macos-client-patterns.md#9-window-widths-regular-and-wide); plan [`../prs/S13-12-family-section.md`](../prs/S13-12-family-section.md).
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
2. **Readable things keep their measure.** Prose, forms and the primary column keep their Regular maximum widths. Wide adds columns; it does not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Crossing the line keeps state.** Selection, scroll position, focus and unsaved edits survive a resize across 1600.
6. **Past every column's maximum, center the group.** Page gutters are never narrower than at Regular; they take what is left.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.

### This brief is an enhancement

Extend the S13-D1 frames. The header, the switcher and the other four sections stay exactly as agreed; only the Family section's content changes, from its empty place (S13-D1 PPF-10) to the family diagram. Draw on the S13-D1 Person page frame, with the switcher on **Family**. Do not start a second Person page.

---

## 1. Objective

A researcher opens a Person's Family section and sees, in one plain diagram, the people the person grew up with and the people who grew up with them: their parents, their siblings, each spouse or partner, and each family's children. Each relative is a **person tile** (likeness, name, birth and death) in their gender color, and a click on a tile opens that relative's Person page on its own Family section, so the researcher can walk the family from person to person.

**What changes.** The Family section's empty place becomes the family diagram, with lines between the tiles, a state for a person with no known family, and loading and error states of its own. The **person tile** is drawn here for the first time; the Timeline board (S13-D4) reuses it unchanged.

**What stays.** The header, the switcher, Vitals, Gallery, and the Timeline and Sources empty places (S13-D1). The gender colors (S10-D12). The section switcher still names the section; Family has no heading of its own, as Vitals has none.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Canonical people only (R-2) | Every tile is a promoted (canonical) person. Nothing unpromoted is drawn, and no placeholder stands in for an unknown relative: where nothing is promoted, nothing is drawn. |
| The data comes grouped (working rule "Family groups in Go") | The engine returns the person's family as **units**, in this order: (1) each **parent set** with the siblings who share it; (2) each **partner** with the children the two share; (3) the person's **children with no other known parent**; (4) **siblings whose parents are not known** (a recorded sibling link with no parent behind it). The board lays out these units; it does not regroup people. |
| A parent set is one or two parents | Two parents are a set when the evidence pairs them (a spouse link, or both named as parents in one birth). A parent with no pair is a set of one. A person can have **two or more parent sets** that conflict (for example a birth record naming one father and a baptism naming another): each set is its own unit (R-19). |
| Siblings are recorded or implied | A sibling is recorded (a sibling link) or implied by a shared parent. The diagram draws both the same way (user, 2026-11-01): it does not mark which is which. Siblings through a parent set sit under that set; a sibling who shares two of the person's parent sets sits once, under the set of exactly the parents they share. |
| A partner can be a co-parent | A partner unit is a spouse (recorded, or implied by a marriage) or a **co-parent**: the other parent of one of the person's children who is not a spouse. Both look the same on the tile; the line between the person and a co-parent is the same partner line (user, 2026-11-01: recorded versus implied is not drawn). |
| Children are in birth order | Within each unit, the children (siblings, or the person's children) come earliest born first; children with no birth date come after the dated ones (user, 2026-11-01). Units keep the engine's order. |
| The person is in the diagram | The person whose page this is sits in the diagram where they belong: among the siblings under each parent set, and beside each partner. Their tile is the **focus**: it looks set apart from the relatives' tiles and does not open anything (the page is already theirs). |
| A tile's content (R-21) | A likeness (the relative's lead picture, else the person mark), the name (the reconciled name, else the working label in italic, else the ref in mono, as every Person row does), and the b. and d. lines (date · place). A missing half is left out, as in the page header: a tile with no dates and no places shows the name alone. Colored per the gender treatment (R-3): the thumb's tint and line, as the page header's thumb has them. |
| Gender colors (R-3, S10-D12) | Man: accent tint. Woman: plum tint. Non-binary: info tint. Other or unknown: sunken tint, inked line. Gender wins; sex at birth stands in when no gender is recorded. A person with two disagreeing values gets no tint (the unknown treatment). The colors live on the thumb only, as on the page header; do not tint the whole tile. |
| A tile opens the relative's Family (R-20; user, 2026-11-01) | A click (or Return / Space with focus) on a relative's tile opens their Person page on its **Family** section. It is one navigation step: Back returns to this Person's Family. This replaces S13-D1 PPF-5's "opens on Vitals" for family tiles only; every other link to a Person still lands on Vitals. |
| Sizes | Typical: two parents, three to six siblings, one or two partners, two to eight children: 10 – 20 tiles. Large: two parent sets, ten siblings, three partners, twelve children: about 30 tiles. A malformed import can name 50 relatives; the diagram must still lay out, wrap and scroll, not overlap. |
| Names can be long | "Johanna Wilhelmina Catharina van der Berg-Schoonhoven". Tiles have a fixed width; names wrap to two lines, then truncate, with the full name in the tooltip and VoiceOver. |
| Each relative once | A relative appears once in the diagram: a sibling under the one parent set they belong to, a partner in their own unit with the children the two share. Only the focus tile can appear in more than one unit (FS-3). |
| The read can be slow or fail | The section has its own read. First view: a short load. A failure is a danger Callout in the section, not over the page. Refreshing (after a relative changes) keeps the diagram drawn. |
| No family | A person with no promoted parents, siblings, partners or children shows an EmptyState. |

### 2.1 What this board is not

- A family tree: grandparents, aunts and uncles, cousins, more than one generation up or down (R-17; [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md)).
- A canvas: no pan, zoom, drag or free placement (R-16). The diagram is laid out by the app from the units.
- Editing a family: no add parent, add child, link, unlink or reorder.
- The provenance sheet on a relative or a line (R-13b), or any mark of which records a relationship rests on.
- A distinction between recorded and implied relationships (user, 2026-11-01).
- The Timeline participants (S13-D4), which reuse this board's tile.
- Step-, half- and adoptive relationships as named kinds.

---

## 3. Implementation gate

| | |
| --- | --- |
| Ships in **S13-12** | FS-1 … FS-14 and FS-W1 … FS-W3: the Family section on the Person page, the person tile, the lines, birth order of children, the empty, loading, error and refreshing states, keyboard and VoiceOver, light and dark, at Regular and Wide. |
| Later boards | S13-D4 places the same tile in the Timeline's participant lists; the tile is not redrawn there. |
| Stays out | Any change to the header, switcher or other sections; the Event and Place pages; a tree beyond the nuclear family. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| FS-1 | **The person tile.** Draw one tile, used for every person on the diagram: a likeness thumb with the gender tint and line, the name, the b. line and the d. line (date · place), from the kit's Card and Thumbnail. Pick its width and the thumb size so a row of four or more tiles fits the Regular primary column. The same tile, unchanged, serves the Timeline later (working rule "One person tile, built once"). |
| FS-2 | **Tile states.** A tile with a likeness; with the person mark (no likeness); with no dates or places (name only); with only a ref (no name); with a long name (two lines, then truncated). Gender: man, woman, non-binary, unknown. Hover and keyboard focus (kit focus ring). |
| FS-3 | **The focus tile.** The person whose page this is appears in the diagram as a set-apart tile (for example an inked border or a raised Card) that is not a link: no hover, no pointer, no focus stop. Draw it where it sits in each unit: among the siblings, beside each partner. Decide whether it repeats once per unit or is drawn once, with lines from it to each unit; say why. |
| FS-4 | **A parent set.** One or two parent tiles side by side, joined by a partner line, with a line down to the row of their children: the siblings and the focus tile, children in birth order (§ 2). A set of one parent has the same line down. |
| FS-5 | **Two conflicting parent sets (R-19).** Two parent sets, each with its own row of siblings. The researcher can tell at a glance that these are two different accounts of who the parents are, not four parents. Decide how they sit (side by side, or one above the other) and whether each set has a short caption ("Parents", "Other parents") or nothing; say why. |
| FS-6 | **Partners and their children (R-19).** The focus tile beside each partner, joined by a partner line, with a line down to the children the two share, in birth order. A person with two marriages shows two partner units. A co-parent (not a spouse) draws exactly like a spouse. |
| FS-7 | **Children with no other known parent.** The person's children who have no known other parent: a line from the focus tile down to them, with no partner tile and no empty slot for one. |
| FS-8 | **Siblings whose parents are not known.** Recorded siblings with no shown parent sit with the focus tile in a row joined by a sibling line, with no parent above them. |
| FS-9 | **Where each unit sits.** Lay out the four kinds of unit in one plain reading order: parents and siblings above or first, the person's own partners and children below or after (R-22 "parent to child, spouse to spouse, siblings through their parents"). Lines are thin, in one kit line color in light and dark; they never cross a tile or a name. A long row of tiles wraps; the lines still read as one family. Recommended: units stack top to bottom in the engine's order, each unit a short block of parents (or the focus and partner) over its children. |
| FS-10 | **Opening a relative (R-20).** A click on any relative's tile opens that relative's Person page on Family. Draw the toolbar for one frame: after the click, the breadcrumbs read *Persons › Johann Weber › Family* and the Back menu lists *Anna Weber · Family*. |
| FS-11 | **No family.** A person with no promoted relatives shows the kit EmptyState, with the Family glyph, a title and one line on where relatives come from. Recommended: "No family yet" and "Relatives appear here when people linked to this person in an Evidence graph are promoted." |
| FS-12 | **Loading, error, refreshing.** First load of the section: a small spinner where the diagram goes (the header and switcher already show). Failure: a danger Callout in the section with the error message. Refreshing after a relative changed: the diagram stays drawn; show the refresh where Vitals shows it, at the end of the switcher row (S13-D1 PPF-8), not a second spinner. |
| FS-13 | **Keyboard and VoiceOver.** Tab moves through the relatives' tiles in reading order (FS-9); Return or Space opens the focused tile. VoiceOver reads each tile as one button: its role, name and life, for example "Parent, Anna Weber, born 3 May 1790, Bremen, died 1851, button". Roles: Parent, Sibling, Partner, Child. A unit is a group VoiceOver names ("Parents", "Partner Maria Schulz and children", "Children", "Siblings"). Lines are not read. |
| FS-14 | **Dark mode.** Frames 1, 3 and 9 in dark mode: tiles, the focus tile, every gender tint and the lines. |
| FS-W1 | **The diagram at Wide.** Family is primary-only, as Vitals is (S13-D1 PPF-W2), and the diagram is a canvas-like layout, so rule 3 applies: rows of tiles fill the primary column at the same tile size, and at Wide two units may sit side by side where they fit (for example two partner units, or two parent sets). Tiles keep their Regular size. Decide what sits side by side and what stacks; say why in one clause. |
| FS-W2 | **No secondary column.** The section adds nothing to Page columns' secondary column; the diagram is centered primary-only at both widths (WL-6). |
| FS-W3 | **Crossing 1600 keeps the place (rule 5).** Resizing across the line keeps the section, the scroll position and the focused tile. A note on the Wide frame. |

### 4.1 Wide layout

| View | Regular (pane below 1600) | Wide (pane 1600 and up) | What moves where |
| --- | --- | --- | --- |
| Family section | Units stacked in the primary column, rows of tiles wrapping | The same units in the primary column, more tiles per row, two units side by side where FS-W1 says | Nothing leaves the primary column; no secondary column (FS-W2) |
| Person tile | Fixed size | Same size | Nothing (rule 3: cards keep their size) |
| No family, loading, error | EmptyState, spinner or Callout in the primary column | Centered primary-only | Nothing |

---

## 5. Suggested frames

Each frame is the S13-D1 Person page pane only (no app sidebar), with the switcher on **Family**, at **Regular** (pane 1102 × 900) unless it says Wide (pane 2296 × 1440). The Wide frames sit beside their Regular frame.

Sample family for frames 1 – 3: **Johann Weber** (the person; man), b. 12 Mar 1817 · Bremen, d. 1880 · Hamburg. Parents Heinrich Weber (man, 1788 – 1849) and Anna Weber née Kohl (woman, 1790 – 1851). Siblings Margarethe (woman, 1815), Friedrich (man, 1820) and Wilhelm (no gender recorded, no dates). Wife Maria Schulz (woman, 1822 – 1859) with children Karl (1845), Luise (1848) and Emma (1851); second wife Sophie Brandt (woman, 1830 –) with child Otto (1862); a son Paul (1870) with no other known parent. Half of the tiles have a likeness.

1. **Full family.** Johann's Family: one parent set with three siblings and the focus tile, two partner units, one child with no other parent (FS-1, FS-3, FS-4, FS-6, FS-7, FS-9).
2. **Tile sheet.** The tile states of FS-2, side by side, including the focus tile and the hover and focus states.
3. **Conflicting parents.** Johann with a second parent set (Heinrich Weber and Katharina Weber, from a baptism) and one sibling under it (FS-5).
4. **Recorded siblings only.** A person whose only family is two recorded siblings, no parents (FS-8).
5. **One parent.** A person with a mother only, two siblings under her (FS-4, set of one).
6. **Large family.** Two parent sets, ten siblings, three partners, twelve children: wrapping rows, the page scrolled halfway (FS-9; § 2 Sizes).
7. **Long names and sparse tiles.** Frame 1 with the long name from § 2 on one sibling and a ref-only child.
8. **Opening a relative.** Frame 1 after clicking Anna Weber: her Family, with the toolbar above the pane showing the breadcrumbs and the Back menu (FS-10).
9. **No family.** The EmptyState (FS-11).
10. **Loading and error.** The section's first load and its error Callout (FS-12).
11. **Keyboard.** Frame 1 with focus on Friedrich's tile; the VoiceOver reading of it as a note (FS-13).
12. **Dark mode.** Frames 1, 3 and 9 in dark (FS-14).
13. **Wide full family.** Frame 1 at Wide (FS-W1, FS-W2), with the rule-5 note (FS-W3).
14. **Wide large family.** Frame 6 at Wide (FS-W1).
15. **Wide dark.** Frame 13 in dark.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person page frame, header, switcher | Snowflake / component | Ship | `Features/Conclusions/ConclusionDetailPage.swift`, `PVPageColumns`, `PVSegmentedControl` | As S13-D1. Family fills the primary column; nothing in the secondary (FS-W2). |
| Family section | Snowflake | **New** | `Features/Conclusions/PersonFamilySection.swift` *(new)* | The diagram: units laid out per FS-9, tiles, lines. Only the Person page uses it. Not a canvas, not `GraphCanvas`. |
| Person tile | Snowflake (recipe when the Timeline adopts it) | **New** | `Features/Conclusions/PersonTile.swift` *(new)* | FS-1, FS-2. One tile for Family now and Timeline next (working rule "One person tile, built once"). Composed from Card and Thumbnail; no new kit primitive. |
| Tile frame | Component | Ship | `PVCard` (`DesignSystem/Components/Card/PVCard.swift`) | FS-1, FS-3. The relative tile is a card; the focus tile is the same card set apart (an inked border or the raised surface; not a new card style). |
| Likeness with gender tint | Component | Ship | `CachedThumbnail` over `PVThumbnail` with `PVThumbnail.Colors` from `PersonGenderTreatment` | FS-1, FS-2. The person mark when there is no likeness, as in the page header and the Persons list. |
| Name | Snowflake | Ship | `ConclusionListRow.title` with `PersonTitleDisplay` | Name plain, label italic, ref mono, as every Person row. |
| b. and d. lines | Snowflake | Ship | `ConclusionDatePlaceLine` | The same line the page header draws, at the tile's size. |
| Lines between tiles | Snowflake | **New** (inside the Family section) | `PersonFamilySection.swift` | FS-4 – FS-9. Thin lines in one existing kit color token (border or the lineage-line tokens); not a new token, and not the inferred line (no recorded/implied difference). |
| Unit caption, if FS-5 uses one | Component | Ship | small muted text, or `PVSectionHeader` | FS-5. Only if the board gives parent sets a caption. |
| No family | Component | Ship | `PVEmptyState` | FS-11. The Family glyph `gitBranch`, as S13-D1's empty place. |
| Loading, error, refreshing | Component | Ship | `ProgressView` small, `PVCallout` danger; the switcher row's refresh indicator | FS-12. As S13-D1. |
| Breadcrumbs, Back menu | Component | Ship | `PVBreadcrumbs`, the toolbar jump menu | FS-10. Unchanged; they already name the section. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A canvas, pan, zoom or free placement | R-16: a plain laid-out diagram. |
| A tree beyond the nuclear family, or an "expand" to grandparents | R-17; [`docs/ideas/kinship-view.md`](../../../ideas/kinship-view.md). |
| A recorded / implied mark: dashed lines, the inferred line color, a Badge | User, 2026-11-01: the diagram draws both the same. |
| Gendered role words on tiles (Mother, Father, Son) | The roles are Parent, Sibling, Partner, Child; the gender is in the colors. |
| Placeholder tiles for unknown relatives ("Unknown father") | R-2: where nothing is promoted, nothing is drawn. |
| Add, link, unlink or reorder controls, or a context menu on a tile | This spike is view only. |
| The provenance control on a tile or a line | R-13b. |
| A second tile design for the Timeline | Working rule "One person tile, built once". |
| A new Card style, color token or kit primitive for the tile | Card and Thumbnail cover it; one call site is not a primitive. |
| Content in the secondary column at Wide | FS-W2. |

---

## 7. Handoff

1. When the board is agreed, record on the **Board** line: the Claude Design turn and frames, the tile's size and thumb size (FS-1), how the focus tile is set apart and whether it repeats per unit (FS-3), how two parent sets sit and their caption (FS-5), the unit order and the line color token (FS-9), the empty-state copy (FS-11), what sits side by side at Wide (FS-W1), and every other change from this brief with the user's answer.
2. Refresh the [S13-12 plan](../prs/S13-12-family-section.md): copy the answers into its Decisions and **Design → macOS**, close its Open questions, run `review-plan` again, and set it to `Planned` once S13-04 has landed.
3. S13-12 archives this brief under `design/archive/` and records it in [`../completed.md`](../completed.md). S13-D4 reuses the tile from this board.
