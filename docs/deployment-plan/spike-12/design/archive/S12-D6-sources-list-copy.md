# S12-D6 — Copy Source from the Sources list

**Kind:** Claude Design board (**enhancement** of the Sources list row)
**Spike:** Provenencia Spike 12 (Source structure)
**View:** Sources list — a row's right-click menu with Copy…
**Implements later as:** PR **S12-17**
**Depends on:** S12-D1 (the Copy Source form this menu opens), S12-02
**Related:** [`../source-structure.md`](../../source-structure.md#copy-source) (the design), [`../deployment-plan.md`](../../deployment-plan.md) working rule "Entry points". S12-D5 (Sources list graph accordion, S12-14) later puts graph sub-rows under each row.
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

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

### This brief is an enhancement

Add a right-click menu to the Sources list row that exists. Do not redraw the row, the list, its filters or its header. Do not redraw the Copy Source form: it is S12-D1's form, opened from here. Frames that show the form reuse S12-D1's frames.

---

## 1. Objective

Copying a Source should not require opening its page first. From the Sources list, a researcher right-clicks the obituary they just finished and starts the next one.

**What changes.** A Sources list row gets a right-click menu with one item, **Copy…**, which opens the S12-D1 Copy Source form for that row's Source.

**What stays.** The row's two click zones (the page zone opens the Source page; the graph zone opens its Evidence graph), its thumbnail, title, type, progress counts and Updated column. The list header, filters and Add Source.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| A row has two zones | Left: thumbnail, title, type, which opens the Source page. Right: the Evidence graph progress, which opens the graph. Left-click behavior does not change. |
| The menu is on the page zone | Right-click on the left zone opens the menu. Right-click on the graph zone does nothing new. S12-14 later adds graph sub-rows under each row (S12-D5); they do not get this menu. |
| Kit right-click menus exist | The person page portrait has one, with a VoiceOver action for its item: a small panel at the pointer, a muted caps title row optional, keyboard-highlightable items. Use the same ContextMenu. |
| One item | **Copy…**. No Delete, Open graph or other items on this board. |
| The form loads from here | A list row carries only the Source's title, type and cover. The form (S12-D1) opens with title and type at once, and shows its loading state (S12-D1 CS-8) while the description, metadata, credibility and Artifacts load. |
| Saving opens the copy | As from the Source page. Returning to the list shows the copy as a new row. |
| VoiceOver | A right-click is not reachable with VoiceOver. The row offers Copy… in its actions rotor. |

### 2.1 What this board is not

- The Copy Source form (S12-D1).
- More menu items, a row hover button, a row overflow (⋯) button or multi-select.
- The graph accordion (S12-D5).

---

## 3. Implementation gate (S12-17)

| Ships in **S12-17** | Stays out |
| --- | --- |
| A right-click menu with Copy… on a Sources list row's page zone, the row's accessibility action, opening the S12-D1 form from the list. | Any other menu item. A hover or overflow control. A shared kit menu modifier (logged in `docs/ideas/shared-list-helpers.md`). Any change to the Copy form. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CL-1 | Right-click (or Ctrl-click) on a row's page zone opens a kit context menu at the pointer with one item, **Copy…**. |
| CL-2 | Choosing Copy… opens the S12-D1 Copy Source form for that row's Source, prefilled as on the Source page. |
| CL-3 | Left-click on the page zone still opens the Source page; left-click on the graph zone still opens the graph. Right-click on the graph zone opens nothing. |
| CL-4 | The row's VoiceOver actions include Copy…, doing the same as CL-2. |
| CL-5 | The menu is keyboard-highlightable like the cover menu (arrow, Return, Escape). |

---

## 5. Suggested frames

1. **A Sources list row, right-clicked** on the page zone, the one-item menu open at the pointer (CL-1).
2. **The S12-D1 form opened from that row, loading** (S12-D1 CS-8), over the list (CL-2).
3. **The list after the copy**, with the copy as a new row (returning from the copy's page).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sources list row | Snowflake | **Extend** | `Features/Sources/SourcesSplitRow.swift` | The page zone gains the menu trigger. Nothing else changes. |
| Row menu | Component | Ship | `DesignSystem/Components/ContextMenu/PVContextMenu.swift` (`pvContextMenuTrigger`, `PVContextMenuPanel`, `PVContextMenuItem`) | One item. Same pattern as the person page portrait menu (`Features/Conclusions/ConclusionDetailPage.swift`), including its VoiceOver action. |
| Copy Source form | Snowflake | Ship | S12-D1, `Features/Sources/` (from S12-02) | Opened as is. Not redrawn. |

**Non-goals.** No new kit primitive. No local menu. No hover or overflow button.

---

## 7. Handoff

When the board is agreed: move this brief to the spike's `design/archive` folder, update [`README.md`](../README.md), and plan S12-17 with `plan-pr`.
