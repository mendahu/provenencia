# S7-D3 — Evidence graph updates (Add property, cited rows, connect handoff)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PRs **S7-09** first (Add property + card growth), then **S7-10** (connect)  
**Depends on:** Spike 6 canvas + cards (S6-D1/D2). Composer place chrome is **S7-D4** — this board only needs a handoff target (navigate away), not the finished composer.  
**Related briefs:** [`S7-D4`](S7-D4-citation-composer.md) — composer place (do not design it here)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md) — inventory in §9  
**Dogfood order:** Design and ship **S7-09** before thick composer work so Add property is clickable early.

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first. When implementing, treat **§9 UI building-block inventory** as binding: compose existing kit pieces; only invent what the table marks **New**.

---

## 1. Objective

Design **Evidence graph chrome updates** now that Citations / Observations exist:

- **Add property** affordance on primary (and bridge) cards
- **Cited-data rows** that make cards grow vertically
- Uncited → cited shell transition once Observations exist
- **No-Artifact** empty / disabled state for citing
- **Connect** disambiguation sheet, then **navigate away** to the citation composer place (Option B) — this board designs the graph-side handoff only

**Do not design the composer layout, viewer, or form.** That is **S7-D4**.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Graph is the only Subject surface | No parallel list of Observations. Rows live on the card. |
| Uncited = zero Observations about the subject | Working `label` is not an Observation. Keep Spike 6 uncited treatment until first Observation. |
| Cards grow | Cited property rows stack; width stays fixed (~236pt family); height grows; edge attachment must still read correctly. |
| Add property needs an Artifact | Source with no Artifact: disable or explain; route toward adding a file on the Source page. |
| Composer is another place | Add property / connect completion **navigates** (leaves the canvas). Propose transition cue (brief), not a sheet over the graph. |
| Bridges require Citation at create | Connect never creates an uncited bridge. Disambiguation on-graph → composer with two edge Observations pre-scoped. |
| Provisional Spike 6 links | Replace honesty labels with real cited edges after S7-10; propose interim vs final bridge body. |

### 2.1 What this board is not

- Not the citation composer place (viewer\|form, breadcrumbs) — **S7-D4**.
- Not Subject types / fields admin — fields are **S7-D2**; types editor is **descoped**.
- Not NameValue editor chrome — **S7-D5**.
- Not conflicted / negated Observation visual language (thin OK; full polish later).
- Not unplaced-subjects tray / minimap.

### 2.2 Implementation gate

| Ships in S7-09 / S7-10 | Does **not** ship there |
| --- | --- |
| Add property control → navigate to composer | Composer layout (S7-08 / D4) |
| Cited-data rows + card growth | Locator tools |
| Connect disambiguation sheet + handoff | Pinning Citations across edits |
| Bridge card body once edges are real | Companion window / modal composer |

---

## 3. Card model updates (authoritative for D3)

Build on Spike 6 primary / bridge cards.

| Element | Intent |
| --- | --- |
| **Add property** | Clear control on activated/selected card (and/or always-visible quiet control). Opens navigation to composer for that subject. |
| **Cited rows** | Each Observation (or Property summary) as a compact row: property label + value summary + polarity hint if negative. Propose density at canvas zoom. |
| **Growth** | Card height grows with rows; propose max before scroll-inside-card vs always grow (prefer grow for tens of rows on a Source). |
| **Uncited → cited** | After ≥1 Observation about the subject, drop uncited shell (Spike 6 contrast frame). |
| **Bridge cards** | Subordinate chrome remains; body shows relationship/role summary from Observations once durable. |
| **Palette / type chrome (implementation)** | Membership, copy, icons, card colors, and line/gradient tokens for **every** shown Subject type (roots and bridges) come from the Interpretation subject registry — design may keep today’s look; implementers must not keep hard-coded kind→style maps as SoT (see deployment plan registry section). |

---

## 4. Flows (authoritative)

### 4.1 Add property (primary)

```text
1. Select / activate card
2. Add property
3. If Source has no Artifact → gate (do not enter composer empty)
4. If multiple Artifacts → may pick on graph or defer pick to composer (propose; prefer composer-owned pick if simpler)
5. Navigate to citation composer place for this subject
6. On return (submit or Back): graph shows updated rows / selection restored if possible
```

### 4.2 Connect (durable)

```text
1. Connect tool: click A, then B
2. Disambiguation sheet on graph (role / relationship_type / refuse / shared-event choice) — §3.2 matrix
3. Confirm → navigate to composer with bridge + two edge Observations pre-filled
4. Submit → Back to graph with bridge card + edges; cancel → nothing written
```

Disambiguation is **not** a history entry. Composer place is.

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| EG-1 | **Add property** on primary cards; propose bridge-card variant (bridges may add further properties later). |
| EG-2 | **Cited-data rows** with property + value summary; cards grow. |
| EG-3 | Uncited vs cited shell wired to Observations (not to working label alone). |
| EG-4 | No-Artifact gate with clear recovery path. |
| EG-5 | Connect disambiguation sheet covering person→event, person→person, event→place, refusals. |
| EG-6 | Handoff to composer is navigational (leave canvas); propose any transient confirmation. |
| EG-7 | Graph a11y: new controls named; rotor/list still works; **do not** design composer a11y here. |
| EG-8 | Edge geometry still attaches sensibly to taller cards (note for implementers). |

---

## 6. Screen / frame inventory (minimum)

1. Uncited Person card with Add property.
2. Same card after two cited properties (grown body).
3. No-Artifact gate.
4. Connect disambiguation (person→person choice).
5. Bridge card after durable connect (cited body).
6. Annotation of navigation away (graph → composer) without drawing the composer.

---

## 7. Out of scope

- Composer layout, PDF viewer, polygon tools, observation form — **S7-D4**.
- Conflicted / competing Observation badges (later honesty spike).
- Pinning a Citation while staying on the graph.

---

## 8. Deliverable

Claude Design board + short notes for S7-09 / S7-10. Archive this brief when done.

---

## 9. UI building-block inventory (binding for design + implement)

Provenencia UI is layered as **components / recipes / snowflakes** ([`docs/design-system-layers.md`](../../../design-system-layers.md)). This table is the repo SoT for *what* this board may introduce. Paths are from `macos/App/` unless noted.

**How to read**

| Column | Meaning |
| --- | --- |
| **Layer** | Frost layer: Component (`DesignSystem/Components/<Name>/`), Recipe (`DesignSystem/Recipes/<Name>/`), or Snowflake (`Features/…` or rare `DesignSystem/Snowflakes/`). |
| **Status** | **Ship** = already correct; compose as-is. **Extend** = exists; this brief changes it. **New** = create (prefer snowflake under `Features/EvidenceGraph/` unless a second call site is already known). |
| **Home** | Intended file / folder after S7-09 / S7-10. |

Do **not** invent new design-system **components** on this board unless a row says so and a finding is raised. Prefer slots on existing kit chrome (`PVFormDialog`, `PVButton`, `PVCallout`, …).

### 9.1 Destination shell (already shipping)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| Evidence graph destination | Snowflake | Ship | `Features/EvidenceGraph/EvidenceGraphView.swift` | Hosts canvas, palette, create sheet, banners. Extend wiring only (handoff, gates). |
| Graph canvas shell | Snowflake | Ship | `Features/GraphCanvas/` (`GraphCanvasScrollView`, pointer, edges) | Pan/zoom/hit targets. Not redesigned here; taller cards affect edge attachment (EG-8). |
| Floating place / connect palette | Snowflake | Ship | `Features/EvidenceGraph/EvidenceGraphPalette.swift` | Keep Connect tool; do not redesign palette IA. |
| Subject type marks | Recipe | Ship → **S7-D6 / S7-12** | Today: `DesignSystem/Recipes/SubjectIcon/`. After: `DesignSystem/Recipes/Marks/` (`subject_*` beside `file_*` / `type_*`). | Card / palette icons. Pigment still follows registry / kind style — not Lucide. **S7-D6** owns the Marks merge; this board does not redesign marks. |
| Create-subject form dialog | Component | Ship | `DesignSystem/Components/FormDialog/PVFormDialog.swift` | Already used for place/create. Disambiguation (§9.3) should reuse this chrome, not a hand-rolled sheet. |
| Create-subject form body | Snowflake | Ship | Private form inside `EvidenceGraphView` | Label field for new Person/Event/Place. Connect durable path may still land here after disambiguation, or skip straight to composer — board proposes. |
| Armed / connect hint banner | Snowflake | Ship | Private in `EvidenceGraphView` | Keep lightweight; not a new kit control. |
| Edge + rubber-band paint | Snowflake | Ship | `Features/EvidenceGraph/EvidenceGraphEdgeLayer.swift` (+ GraphCanvas edge geometry) | Recompute attach points when cards grow. |
| Feedback toast | Component | Ship | `DesignSystem/Components/Toast/` (+ vocabulary toast overlay) | Invalid connect already toasts; optional for gates if not a callout. |

### 9.2 Card chrome (S7-09 — Add property + cited rows)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| Primary subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Fixed width (~236). Grow height with cited rows; keep uncited → cited shell transition. Private chrome today (`EvidenceSubjectCardChrome`). |
| Bridge / relationship card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Subordinate width (~188). Replace provisional honesty body with cited relationship/role summary when durable. |
| Kind → ink / gradient map | Snowflake (debt) | **Extend** | `Features/EvidenceGraph/EvidenceSubjectKindStyle.swift` | Implementation must move SoT to Interpretation subject registry (deployment plan). Design may keep today’s look. |
| **Add property** control | Component + Snowflake | **New** (control uses kit) | Affordance on card: prefer `DesignSystem/Components/Button/PVButton.swift` or `IconButton/PVIconButton.swift`; placement/wiring stays in `EvidenceSubjectCard` | Quiet on selected/activated card (board proposes always-visible vs activated-only). Opens navigation — not a sheet. |
| **Cited-data row** | Snowflake | **New** | Prefer `private` row view colocated with `EvidenceSubjectCard` (e.g. `EvidenceCitedPropertyRow`) | Property label + value summary + optional polarity. Canvas-zoom density is a design finding. **Do not** promote to a recipe until a second call site exists. |
| Card growth / scroll rule | Snowflake | **New** (behavior on card) | Same card files | Prefer grow for tens of rows; propose max-before-inner-scroll if needed. Update `edgeLayoutHeight` / attach math with real body height. |

### 9.3 Gates, connect handoff (S7-09 gate + S7-10)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| **No-Artifact** gate | Component + Snowflake | **New** (compose kit) | Prefer `DesignSystem/Components/Callout/PVCallout.swift` or EmptyState; copy + recovery action in `EvidenceGraphView` / model | Disable Add property or intercept click; recovery → Source page add Artifact. Do not invent a new dialog primitive. |
| **Connect disambiguation** sheet | Component + Snowflake | **New** (body) | Chrome: `PVFormDialog` (or Confirm if the choice set stays tiny). Body: private snowflake under `Features/EvidenceGraph/` | Covers person→event, person→person, event→place, refusals. **Not** a history entry. Confirm → navigate to composer with pre-scoped bridge + edge Observations. |
| Choice chips / segmented options (if needed in disambiguation) | Component | Ship | `DesignSystem/Components/Chip/PVChip.swift` (+ `PVChipGroup`) | Use for short closed lists (role / relationship_type). Longer vocabularies → ComboBox (composer / later). |
| Navigation handoff (graph → composer) | Platform (not DS) | **New** place later | `WorkspaceNavigation` + future composer `WorkspaceLocation` (S7-D4 / S7-08) | Board only annotates leave-canvas cue. Do not mock composer UI here. |

### 9.4 Explicit non-goals for this inventory

| Do not add | Why |
| --- | --- |
| New `DesignSystem/Components/*` for “graph card” or “cited row” | Cards and rows are Evidence-graph snowflakes until a second product surface needs them. |
| Composer viewer / observation form / term picker | **S7-D4**. |
| Subject fields admin chrome | **S7-D2** (shipped / archived). |
| Replacing subject marks with Lucide / SF Symbols | Curated pack stays; consolidation is **S7-D6** / **S7-12**, not this board. |

### 9.5 Suggested implement order (matches dogfood)

1. **Extend** `EvidenceSubjectCard` — Add property (`PVButton` / `PVIconButton`) + cited-row snowflake + growth + uncited shell.  
2. **No-Artifact** gate via `PVCallout` (or toast) + model check.  
3. **Extend** `EvidenceBridgeCard` body once durable Observations exist.  
4. **Connect disambiguation** snowflake inside `PVFormDialog` → navigate to composer place.
