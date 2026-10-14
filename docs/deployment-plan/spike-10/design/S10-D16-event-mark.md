# S10-D16 — Event mark

**Kind:** Claude Design board (**rethink** of one kit mark)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Design system — the `subject_event` mark
**Implements later as:** PR **S10-19**, together with [S10-D7](S10-D7-sidebar-marks.md)
**Depends on:** —
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-19). The sidebar, toolbar, and list empty states that use this mark are **S10-D7**. Cards, omnibar, Promote, and Properties already call `PVMark` and pick up the new artwork with no board of their own.
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

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

### This brief is the design-system board

Paste this file into the Provenencia design-system project. You are already on the kit. Do not fetch a second copy of it, and do not draw this mark on an application screen.

### This brief replaces one mark

Throw away the hourglass drawn for `subject_event`. Draw the new Event mark in that same slot. Leave every other mark on the sheet.

---

## 1. Objective

The Event mark is an hourglass, so an event reads as a timer. This board replaces that artwork with a calendar or datebook, in the same line as the Person and Place marks.

**What changes.** The `subject_event` glyph.

**What stays.** The key `subject_event`. Person, Place, and the bridge and source marks. Marks stay monochrome templates: the caller supplies the ink. No new mark component.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| One key, one asset | `subject_event` is the Event mark. The hourglass in that slot is what goes. The key name stays, so every screen that already asks for `subject_event` shows the new drawing. |
| Match the family | `subject_person` and `subject_place` are the neighbors. The Event mark uses their stroke and optical weight. It is not an SF Symbol, and it is not a filled pictogram sitting next to two line marks. |
| Template, not a colored icon | Artwork is one color, rendered as a template. Event copper, and every other ink, is applied by the screen. Do not bake a color into the glyph. |
| It has to read small | Sidebar and toolbar use it around 15–16pt. List rows use 24pt. A card uses about 15pt. A tile uses 40pt. Draw it so the calendar or datebook is still an event at 15pt. |
| No lettering | Subject marks do not carry text. A datebook page stays a shape. |

### 2.1 What this board is not

- A redraw of Person, Place, relationship, participation, location, or source marks.
- A new mark family, or a second Event key.
- An SF Symbol named calendar.
- The sidebar, the toolbar, or a list. Those are S10-D7, and they instance this glyph.
- A colored Event icon. Ink stays at the call site.

---

## 3. Implementation gate (S10-19)

| Ships in **S10-19** | Stays out |
| --- | --- |
| A new `subject_event` asset, calendar or datebook, template-rendered, at the sizes in §2. The key is unchanged. Person and Place assets stay. | A new `PVMark` type. Recoloring the glyph. Sidebar and toolbar placement (S10-D7). |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EM-1 | Replace the `subject_event` hourglass with one calendar or datebook glyph. The key stays `subject_event`. |
| EM-2 | The glyph matches the stroke and optical weight of `subject_person` and `subject_place`. It is a template: one color, no baked ink, no text. |
| EM-3 | It remains legible at 15pt, 16pt, 24pt, and 40pt. |
| EM-4 | No other mark on the sheet changes. |

---

## 5. Suggested frames

Draw on the marks sheet. Do not draw a sidebar or a card layout.

1. **The family.** Person, the new Event mark, and Place, same size, same stroke. The hourglass is gone.
2. **Small.** The Event mark at 15pt and at 16pt, still readable as a calendar or datebook.
3. **Row and tile.** The same mark at 24pt and at 40pt.
4. **Ink is not in the asset.** The mark drawn once in Event copper and once in a neutral ink, to show the color comes from outside the glyph.

---

## 6. UI building-block inventory

This table is **binding**. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Event mark | Recipe | **Rethink** | `DesignSystem/Recipes/Marks/` asset `subject_event` | EM-1…EM-3. The artwork only. The `PVMark` type stays. |
| Person mark | Recipe | Ship | `subject_person` | Unchanged. The weight reference. |
| Place mark | Recipe | Ship | `subject_place` | Unchanged. The weight reference. |
| Other subject marks | Recipe | Ship | `subject_relationship`, `subject_participation`, `subject_location`, `subject_source` | Unchanged. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A new mark component, or a second Event key | One slot already exists. Replace its artwork. |
| Color, text, or an SF Symbol | EM-2. Call sites tint a template. |
| A sidebar or a card | S10-D7 places the mark. Existing `PVMark` screens pick up the asset. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-19** against this board and [S10-D7](S10-D7-sidebar-marks.md). The shell instances this glyph. This board does not draw the shell.
