# S9-D1 — Workspace sidebar: Source / Conclude / Configure sections

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Workspace sidebar (`WorkspaceSidebar`)  
**Implements later as:** PR **S9-08** (later on this view: S9-23 / S9-26 turn Events and Places live). The configuration renames ship first, in **S9-07b**.  
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

Organize the sidebar into titled sections that follow the research workflow — **Source**, **Conclude**, **Narrate** — at the top, and **Configure** at the bottom:

| Section | Destinations | Alignment | Ships |
| --- | --- | --- | --- |
| **Source** | Sources (count) | Top | Shipped destination, now under a title |
| **Conclude** | **Persons**, **Events**, **Places** (counts) | Top, after Source | New in S9-08 |
| **Narrate** | — | Top, after Conclude | **Not implemented yet.** Not rendered in S9-08; one board frame shows where it will sit so it can be added later without a relayout. |
| **Configure** | Source types, Metadata, Properties | **Bottom**, directly above the session footer (username / account and the collapse toggle) | Shipped destinations, moved and renamed |

Empty space between the top sections and Configure is what separates research from configuration.
- **Renamed configuration destinations** (researcher's decision):

  | Today | New label | Notes |
  | --- | --- | --- |
  | Source types | **Source types** | Unchanged. Plain "Types" would be ambiguous once Subject types are configurable. |
  | Source fields | **Metadata** | The Source metadata fields. |
  | Subject fields | **Properties** | Properties and their bindings to Subject types. |

```text
  ┌──────────────────────────────┐
  │ Provenencia            brand │
  │ SOURCE                       │  ← top-aligned
  │ Sources                  128 │
  │ CONCLUDE                     │
  │ Persons                   42 │
  │ Events                    77 │
  │ Places                    19 │
  │                              │
  │            (space)           │
  │                              │
  │ CONFIGURE                    │  ← bottom-aligned
  │ Source types                 │
  │ Metadata                     │
  │ Properties                   │
  │ ──────────────────────────── │
  │ Researcher name        ⟨⟨    │  ← session footer (shipped)
  └──────────────────────────────┘
```

Section titles are fixed: **Source**, **Conclude**, **Narrate**, **Configure** (verbs for the workflow stage, not nouns for the content). **Sections do not collapse.** Every section is always open, and every destination is a top-level row in its section. Today the configuration destinations are nested as children under Sources; they move out to be Configure's own rows, and nothing in the sidebar uses disclosure. Board findings: the title treatment (it must survive the collapsed rail, where titles drop and only spacing or a rule separates sections), and how research and Configure read as separate beyond the space between them. Keep the shipped sidebar chrome (brand row, footer, collapse rail).

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Only three kinds get pages | Persons, Events, Places. Participations, Locations, Relationships are handles too but have **no** destination. |
| Counts are handles, not members | `42 Persons` = 42 handles, however many Subjects are promoted onto them. |
| Lists start empty | Until the researcher promotes something, counts are 0. |
| Sidebar collapses to a rail | Each new item needs an icon for the collapsed state. |
| Sections are history ids | New `WorkspaceSection` cases (`persons`, `events`, `places`) are persisted in navigation history. Renamed sections keep decoding their old ids (S9-07b). |
| Short windows scroll | The whole column scrolls when it can't fit (W-5b). The bottom alignment applies only when there is spare height; on a short window Configure follows the top sections directly, and nothing overlaps. |

### 2.1 What this board is not

- Not the list pages themselves (S9-D2…D4).
- Not a landing page or dashboard for any section.
- Not Narrate. Its title and destinations come with the Narrative layer; this board only reserves its position.
- Not restyling existing destination rows. Moving the configuration section and renaming its destinations is in scope; the rows themselves keep `PVSidebarNav` styling.
- Not the configuration pages' content. Their titles follow the new labels (S9-07b); nothing else on those pages changes.

---

## 3. Implementation gate (S9-08)

| Ships in **S9-08** | Does **not** ship there |
| --- | --- |
| Source / Conclude / Configure section titles; Conclude's three destinations + rail icons; Persons live with its count | Events / Places pages (stubbed until S9-23 / S9-26); list content |
| Configure bottom-aligned above the footer; Source and Conclude top-aligned | The renames themselves (**S9-07b**, which lands first); the Narrate section |
| L10n + VoiceOver labels | Association-kind destinations |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SB-0 | Sections are titled **Source**, **Conclude**, **Configure** (L10n). **Narrate** is reserved after Conclude but hidden until it has a destination. |
| SB-0b | Sections are not collapsible: no disclosure control on titles or rows. Configuration destinations are top-level rows under Configure, no longer children of Sources. |
| SB-1 | Sidebar shows Persons, Events, Places in the Conclude section, each with a count (`PVSidebarNav` item count slot). |
| SB-2 | Source then Conclude are top-aligned. Configure (Source types, Metadata, Properties, in that order) is bottom-aligned directly above the session footer. |
| SB-2b | When the window is too short for both sections plus spare space, the column scrolls as one (W-5b) with Configure after the top sections; nothing overlaps or clips. |
| SB-2c | Configuration labels read **Source types**, **Metadata**, **Properties** everywhere: sidebar, rail tooltips, VoiceOver, page titles. |
| SB-3 | Collapsed rail shows an icon per destination with the label as tooltip / accessibility label. |
| SB-4 | Zero counts read as zero, not hidden. |
| SB-5 | Selected state and keyboard focus follow shipped `PVSidebarNav` behavior. |
| SB-6 | No new kit primitive. |

---

## 5. Suggested frames

1. Expanded sidebar, tall window: research top-aligned, configuration bottom-aligned above the footer, typical counts.
2. Fresh project: all Conclude counts 0.
3. Persons selected; then Properties selected (a bottom-section selection).
4. Collapsed rail: the three new icons at the top, configuration icons at the bottom above the collapse toggle.
5. Short window: the column scrolls as one, Configure follows the top sections, footer behavior as shipped (W-5b).
6. Reserved: the same sidebar with a **Narrate** section (one placeholder destination) between Conclude and the space, to prove it fits without moving anything else. Not shipped.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Workspace sidebar | Snowflake | **Extend** | `Features/Workspace/WorkspaceSidebar.swift` | Add the group's `PVSidebarNavItem`s. |
| Sidebar nav | Component | Ship | `DesignSystem/Components/SidebarNav/PVSidebarNav.swift` | Items, children, count slot, rail. |
| Workspace section | Snowflake | **Extend** | `Features/Workspace/WorkspaceSection.swift` | `persons` / `events` / `places` + label + icon. Configuration cases arrive renamed from S9-07b. |
| Sidebar layout | Snowflake | **Extend** | `Features/Workspace/WorkspaceSidebar.swift` | Two `PVSidebarNav` sections with flexible space between them inside the existing `ScrollView`; no new component. |
| Icons | Token | Ship / **Add** | `PVSymbol` | Person, event, place symbols; add SF Symbol mappings if missing. |
| Catalog counts | Snowflake | **Extend** | `Features/Catalog/CatalogCounts.swift` | Three new counts. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A bespoke section-title component | `PVSidebarNav` already groups; titles use its group heading (or `SectionHeader` if the board shows the nav lacks one — flag it, don't draw a local copy). |
| Per-kind colored sidebar rows | Kind color belongs to graph cards, not navigation. |
| A second footer or a pinned overlay for configuration | It is ordinary nav content bottom-aligned by spacing, so short windows scroll it like everything else. |

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
