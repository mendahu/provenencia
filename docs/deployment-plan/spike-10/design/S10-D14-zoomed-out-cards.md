# S10-D14 — Evidence graph: zoomed-out card

**Kind:** Claude Design board  
**Status:** Open — draw this board just before **S10-46**  
**Spike:** Provenencia Spike 10 (dogfood UX)  
**View:** Evidence graph, zoomed out (`EvidenceGraphView`)  
**Implements later as:** PR **S10-46**  
**Depends on:** Shipped subject and bridge cards ([`EvidenceSubjectCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift), [`EvidenceBridgeCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceBridgeCard.swift)); kind color ([`EvidenceSubjectKindStyle`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectKindStyle.swift))  
**Related:** S10-02 (tiled grid and edges — not this board); S10-D6 (open file, box-select, Unpromote — a different board on this same view)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board. This board is the Evidence graph only.

This brief is an **enhancement** of a shipped surface: add the zoomed-out face to the existing graph. Do not redraw the card you see at reading zoom.

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

---

## 1. Objective

A real Source can put about a hundred subject cards on one graph. Up close, with a handful of cards on screen, the full card is fine. Zoomed out, with every card in the pane, that same full card (property rows, icons, footer) is what makes panning and dragging sluggish.

Draw the face the card wears **below about 0.65×**. The camera already stops at 0.25× and 1.25×; this board does not change that range. The full card comes back only above about 0.75×, so the two faces do not flicker on the boundary.

The zoomed-out face contains **the kind icon and the label, and nothing else.** No type line, no ref, no property rows, no edit or delete, no Promote, no Add property. The icon is the mark the full card already uses for that kind (person, event, place, or the bridge's own mark). The label is the one string the full card already shows as its name, or the bridge's one-line summary. Person, event, place, and bridge stay distinguishable by that icon and by the kind color they already use. The selected card still reads as selected.

The icon and the label are **larger than on the full card.** The full card's name is 16pt in document space, which is about 10pt on screen at the cutoff and about 4pt at 0.25×. The simplified face spends the room the rows used to take on a bigger icon and a bigger label, still laid out inside the card, so just under 0.65× the name is a comfortable reading size again. That type is in document points: zoom still scales it. At 0.25× the card is about a quarter of its 1× size (a person card is about 66pt wide on screen), so the name cannot stay at that comfortable size. Draw both moments. Do not specify a font that stays a fixed size on screen at every zoom; it will not fit the card at the floor.

The face occupies the **same rectangle** the full card already occupies on the document. Cards do not move, shrink, or reflow when the researcher zooms. They look small because the graph is zoomed out. Pinch back in and the full card returns in that same place.

```text
reading zoom, above ~0.75× (unchanged)     zoomed out, below ~0.65× (this board)
┌ person · CPR-2AB91 ── ✎ 🗑 ┐              ┌────────────────────┐
│ James Robins               │              │  [person]          │
│ age        34              │              │  James Robins      │
│ + Add property             │              └────────────────────┘
│ Promote                    │              icon + label only, larger type
└────────────────────────────┘              same rectangle, kind color kept
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Camera range stays | Magnification is `0.25`…`1.25`. The simplified face is shown below about `0.65`. The full card returns above about `0.75`. |
| Contents are icon + label | The mark the card already uses, and one line of the label. Nothing else is drawn inside the face. |
| Type grows, then zoom scales it | Larger than the full card's 16pt name, so the name is readable again just under `0.65×`. Document points, not a fixed size on screen. At `0.25×` the card is too small for that reading size to hold. |
| Three primary kinds | Person, event, and place already have a mark, tint, line, and ink (`EvidenceSubjectKindStyle`). The zoomed-out face uses those. |
| Bridges are not primaries | A bridge (participation, relationship, location, place relationship) uses its own mark and its one-line summary. It must not wear a person, event, or place mark. |
| The label is one string | A primary's line is the working label the full card already shows (`James Robins`). A bridge's line is the one-line summary the full bridge already shows. Truncate the tail. No second line. |
| Selection already has a treatment | The full card uses an accent ring. The zoomed-out face needs a selected state that survives at that size. |
| Edges attach to the card rectangle | The rectangle does not change, so connectors do not jump when the face swaps. |
| The grid is not this board | S10-02 draws the background as a tile and keeps the connectors. Leave the grid and the lines as they are. |

### 2.1 What this board is not

- Not a redesign of the reading-zoom card.
- Not a new zoom range, a minimap, or a control that hides part of the Source.
- Not the tiled grid, the edge stroke, canvas growth, or the camera (S10-02, S10-03, S10-04).
- Not open file, box-select, or Unpromote (S10-D6).
- Not a smaller layout. The document positions stay.

---

## 3. Implementation gate (S10-46)

| Ships in **S10-46** | Does **not** ship there |
| --- | --- |
| The zoomed-out face below about `0.65`, the full card back above about `0.75` | A researcher-set cutoff. A new camera min or max. |
| Kind icon and one label line, larger than on the full card; truncation; selection | Property rows, type line, ref, edit, delete, Promote, Add property |
| Same document rectangle as the full card | A label that stays a fixed size on screen. Moving cards to pack the overview. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| ZC-1 | Below about 0.65×, a person, event, or place card shows its kind mark and one truncated line of its label, and nothing else. The mark and the label are larger than on the full card. |
| ZC-2 | A bridge shows its own mark and one truncated line of its summary. It does not use a person, event, or place mark. |
| ZC-3 | Just under the cutoff, the label is a comfortable reading size. At 0.25×, draw the same face scaled down with the card; the name may no longer be comfortable, and the mark and kind color still identify the card. |
| ZC-4 | The selected card is identifiable without reading the label. |
| ZC-5 | The face fills the card's existing rectangle. Above about 0.75× the full card returns in that same rectangle, edges unmoved. |
| ZC-6 | A dragged card keeps the zoomed-out face while it moves. It does not pop into the full card for the duration of the drag. |
| ZC-7 | A graph of about a hundred cards, all in view, still reads as a map of kinds and labels, not as a field of identical chips. |

---

## 5. Suggested frames

1. Reading zoom, above 0.75×: one full person card, unchanged, so the board has the face this does not replace.
2. Just under 0.65×: the same card as icon + enlarged label only, inside the same rectangle. The name is a comfortable reading size.
3. At 0.25×: that same face scaled with the card. The name is no longer comfortable; the mark and kind color still read.
4. A zoomed-out graph of many person, event, and place cards.
5. The same graph with one card selected.
6. A bridge between two people, its own mark, not a person mark.
7. A long label, truncated, next to a short one.
8. A card mid-drag below 0.65×, still the simple face.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Reading-zoom card stays. The zoomed-out face is the existing kind mark plus the label, larger, not a second component. |
| Bridge card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | The bridge's existing mark plus its summary, larger. |
| Kind style | Snowflake | Ship | `Features/EvidenceGraph/EvidenceSubjectKindStyle.swift` | Tint, line, and ink for the three primaries. Do not invent a new palette. |
| Evidence graph | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceGraphView.swift` | Shows the face when the card is too small to read. |
| Card (kit) | Component | — | kit | Do not instance a kit Card for this face. One call site, and the graph card is already the snowflake. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit "mini card" or zoom control | One call site. The cutoff is not a setting. |
| Anything but the icon and the label | Type line, ref, rows, and actions stay on the full card. |
| A font that ignores zoom | Type is larger in the card, then scales with magnification. |
| A new kind mark or kind color | Person, event, place, and bridge already have a mark. |

---

## 7. Out of scope

- Grid tile and connector lines (S10-02)
- Canvas growth (S10-03)
- Camera restore (S10-04)
- Hit-target geometry (S10-01)
- Layers or filters that hide cards

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record the agreement in [`../completed.md`](../completed.md).
3. Implement **S10-46** against the board and inventory (kit first). The reading-zoom card is unchanged except that it yields to this face while the rows are too small to read.
