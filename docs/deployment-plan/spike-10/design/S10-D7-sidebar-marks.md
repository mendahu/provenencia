# S10-D7 — Sidebar marks

**Kind:** Claude Design board (**enhancement** of the workspace shell)
**Spike:** Provenencia Spike 10 (dogfood UX)
**View:** Workspace shell — sidebar and the toolbar section icon
**Implements later as:** PR **S10-19**, together with [S10-D16](S10-D16-event-mark.md)
**Depends on:** S10-D16 (the Event glyph this board instances)
**Related:** [`../deployment-plan.md`](../deployment-plan.md) (S10-19). The Event artwork is drawn on the design-system board, **S10-D16**. List empty states for People, Events, and Places use the same marks in this PR. Later list briefs (S10-D9…D11) do not redraw that icon.
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

Instance the subject marks on the workspace shell that exists. Do not draw a second sidebar. Do not draw a new Event glyph on this board. The glyph comes from the kit (S10-D16). Sources, Source types, Metadata, and Properties keep their symbols.

---

## 1. Objective

People, Events, and Places in the sidebar use ordinary symbols, while the same three kinds on a card use the subject marks. This board puts the marks in the shell so the sidebar, the toolbar, and the card agree.

**What changes.** The People, Events, and Places sidebar rows, the toolbar icon when one of those sections is showing, and the empty-state icon on those three lists.

**What stays.** The other sidebar rows. The row labels, order, and selection. The toolbar’s other controls. Person and Place artwork. The Event artwork is the kit’s, not a drawing made here.

---

## 2. Domain facts

| Fact | What it means on this board |
| --- | --- |
| Three rows change | People, Events, and Places. Today they are the symbols person, calendar, and map pin. They become `subject_person`, `subject_event`, and `subject_place`. |
| The rest of the sidebar stays | Sources, Source types, Metadata, and Properties keep the symbols they have (library, tag, list, list-tree). |
| The toolbar shares the section icon | The toolbar section control reads the same icon as the sidebar row for the place you are in. When that place is People, Events, or Places, the toolbar shows the subject mark. Other sections keep the symbol. |
| Size | An expanded sidebar row draws the icon at about 15pt. Collapsed, about 16pt. The toolbar draws it at 15pt. Use the kit mark at that size. Do not invent a sidebar-only drawing. |
| Empty lists | People, Events, and Places lists use an empty state whose icon is still the old symbol. That icon becomes the same subject mark. The empty state component stays. Title and message stay. |
| Everywhere else already uses the mark | An Event card, an omnibar Event hit, Promote, and Properties already call `PVMark`. They are not redrawn here. After S10-D16 they show the new Event glyph on their own. |

### 2.1 What this board is not

- A place to design the Event glyph. That is S10-D16. If the fetched kit still shows an hourglass, stop and use the kit after that board. Do not draw a local calendar.
- A redraw of Person or Place marks.
- New sidebar rows, or a new order.
- A new toolbar, or a new empty-state component.
- Last updated, filters, or a likeness thumb on the lists. Those are later briefs.

---

## 3. Implementation gate (S10-19)

| Ships in **S10-19** | Stays out |
| --- | --- |
| Sidebar People, Events, and Places use the subject marks. The toolbar section icon uses that same mark for those three sections. The three list empty states use the mark in the existing empty-state icon slot. The Event asset is the one S10-D16 drew. | A local Event glyph. New sidebar rows. Redrawing Person or Place. A new empty state. Changing Sources, Source types, Metadata, or Properties icons. |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| WS-1 | The People, Events, and Places sidebar rows show `subject_person`, `subject_event`, and `subject_place`. The other sidebar rows keep their symbols. Labels, order, and selection stay. |
| WS-2 | The mark is the kit mark, at the row’s current icon size, expanded and collapsed. This board does not draw a new glyph. |
| WS-3 | The toolbar section icon shows that same mark when the section is People, Events, or Places, and the existing symbol for every other section. |
| WS-4 | The People, Events, and Places empty states keep `PVEmptyState`. The icon slot shows the subject mark. Title and message stay. |
| WS-5 | Event cards, omnibar hits, Promote, and Properties are not on this board. They already use the mark. |

---

## 5. Suggested frames

Draw the workspace shell. One extra frame is only the empty-state icon. Do not redraw a list row or a card.

1. **Sidebar, expanded.** People, Events, and Places use the subject marks. Sources still uses its library symbol. Events is selected.
2. **Sidebar, collapsed.** The three marks are still readable. The other symbols are unchanged.
3. **Toolbar on Events.** The section icon is the Event mark, the same drawing as the sidebar row.
4. **Toolbar on Sources.** The section icon is still the library symbol.
5. **Empty Events.** The existing empty state, with the Event mark in the icon slot. People and Places are the same treatment. The title and message are the ones the list already has.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Sidebar | Component | **Extend** | `DesignSystem/Components/SidebarNav/PVSidebarNav.swift` | WS-1, WS-2. People, Events, and Places host `PVMark`. Other rows stay `PVIcon`. Not a new sidebar. |
| Section icon source | Snowflake | **Extend** | `Features/Workspace/WorkspaceSection.swift` | The three kinds gain a subject-mark key. The toolbar reads this slot (`WorkspaceToolbar`). |
| Toolbar section icon | Snowflake | **Extend** | `Features/Workspace/WorkspaceToolbar.swift` | WS-3. Same mark as the sidebar row. |
| Subject marks | Recipe | Ship | `PVMark` | Instance `subject_person`, `subject_event`, `subject_place`. The Event artwork is S10-D16. |
| Empty state | Component | **Extend** | `PVEmptyState` | WS-4. Keep the component. The icon slot for these three lists is the mark. |
| List empty call sites | Snowflake | **Extend** | `PersonsList`, `EventsList`, `PlacesList` | Swap the symbol for the mark. Do not redraw the list. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A local Event glyph, or a new mark component | WS-2. The kit already has the key. S10-D16 draws it. |
| A second sidebar or a second empty state | Compose the shell and `PVEmptyState` that exist. |
| New rows, or new icons for the other sections | WS-1. Four sections stay on symbols. |

---

## 7. Handoff

1. Archive this brief under `design/archive/` when the board is agreed.
2. Record it in [`../completed.md`](../completed.md).
3. Implement **S10-19** against this board and [S10-D16](S10-D16-event-mark.md). Design the mark first. This board instances it.
