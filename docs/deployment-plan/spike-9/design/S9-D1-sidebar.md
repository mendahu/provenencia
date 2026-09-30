# S9-D1 — Workspace sidebar: Conclusions group

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Workspace sidebar (`WorkspaceSidebar`)  
**Implements later as:** PR **S9-08** (later on this view: S9-23 / S9-26 turn Events and Places live)  
**Depends on:** S9-07 (Persons list read + counts); shipped sidebar ([`WorkspaceSidebar`](../../../../macos/App/Features/Workspace/WorkspaceSidebar.swift), [`PVSidebarNav`](../../../../macos/App/DesignSystem/Components/SidebarNav/PVSidebarNav.swift))  
**Related:** S9-D2…D4 (the lists these open); [`deployment-plan.md`](../deployment-plan.md) R5  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); [`add-workspace-place`](../../../../.cursor/skills/add-workspace-place/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of a shipped surface: extend its existing frames.

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

### Shared Spike 9 facts (all Conclusion boards)

- A canonical **Person / Event / Place** (`PER-…` / `EVT-…` / `PLC-…`) is a researcher's handle for one historical thing. In the UI it is a Person, never a "canonical entity." Interpretation Subjects on Evidence graphs keep candidate refs (`CPR-…`).
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Every Property is **multi-valued** across members.
- Each field shows **one resolved value** in a state: **single**, **merged** (names and dates auto-reconciled), **mixed** (members disagree; top-ranked value shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

Add a **Conclusions** group to the sidebar with three destinations — **Persons**, **Events**, **Places** — each with a count, alongside today's Sources and the configuration group (Source types, Source fields, Subject fields).

```text
  Sources                    128
  ─ Conclusions ─
  Persons                     42
  Events                      77
  Places                      19
  ─ Configure ─
  Source types · Source fields · Subject fields
```

Grouping label, order relative to Sources, and whether Conclusions is a disclosure group like the config children are board findings. Keep the shipped sidebar chrome (brand row, footer, collapse rail).

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Only three kinds get pages | Persons, Events, Places. Participations, Locations, Relationships are handles too but have **no** destination. |
| Counts are handles, not members | `42 Persons` = 42 handles, however many Subjects are promoted onto them. |
| Lists start empty | Until the researcher promotes something, counts are 0. |
| Sidebar collapses to a rail | Each new item needs an icon for the collapsed state. |
| Sections are history ids | New `WorkspaceSection` cases (`persons`, `events`, `places`) are persisted in navigation history. |

### 2.1 What this board is not

- Not the list pages themselves (S9-D2…D4).
- Not a Conclusions landing page or dashboard.
- Not reordering or restyling existing destinations.

---

## 3. Implementation gate (S9-08)

| Ships in **S9-08** | Does **not** ship there |
| --- | --- |
| Conclusions group + three destinations + rail icons; Persons live with its count | Events / Places pages (stubbed until S9-23 / S9-26); list content |
| L10n + VoiceOver labels | Association-kind destinations |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SB-1 | Sidebar shows Persons, Events, Places in a Conclusions group, each with a count (`PVSidebarNav` item count slot). |
| SB-2 | Order and grouping relative to Sources and the config group are decided on the board; Sources stays first. |
| SB-3 | Collapsed rail shows an icon per destination with the label as tooltip / accessibility label. |
| SB-4 | Zero counts read as zero, not hidden. |
| SB-5 | Selected state and keyboard focus follow shipped `PVSidebarNav` behavior. |
| SB-6 | No new kit primitive. |

---

## 5. Suggested frames

1. Expanded sidebar with the Conclusions group, typical counts.
2. Fresh project: all Conclusions counts 0.
3. Persons selected.
4. Collapsed rail with the three new icons.
5. Short window: sidebar scrolls, footer pinned (existing W-5b behavior).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Workspace sidebar | Snowflake | **Extend** | `Features/Workspace/WorkspaceSidebar.swift` | Add the group's `PVSidebarNavItem`s. |
| Sidebar nav | Component | Ship | `DesignSystem/Components/SidebarNav/PVSidebarNav.swift` | Items, children, count slot, rail. |
| Workspace section | Snowflake | **Extend** | `Features/Workspace/WorkspaceSection.swift` | `persons` / `events` / `places` + label + icon. |
| Icons | Token | Ship / **Add** | `PVSymbol` | Person, event, place symbols; add SF Symbol mappings if missing. |
| Catalog counts | Snowflake | **Extend** | `Features/Catalog/CatalogCounts.swift` | Three new counts. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A Conclusions header component | `PVSidebarNav` already groups. |
| Per-kind colored sidebar rows | Kind color belongs to graph cards, not navigation. |

---

## 7. Out of scope

- List content, empty states inside pages
- Omnibar
- Association-kind navigation

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-08** against the board and inventory (kit first).
