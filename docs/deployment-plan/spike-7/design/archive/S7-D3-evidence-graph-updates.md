# S7-D3 — Evidence graph updates (Add property, cited rows, connect handoff)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 7 (Citations / Observations / composer)  
**Implements later as:** PRs **S7-09** first (Add property + card growth), then **S7-10** (connect)  
**Depends on:** Spike 6 canvas + cards (S6-D1/D2). Prefer **S7-12** (Marks) and **S7-14** (`PVCallout` actions) already landed. Composer place chrome is **S7-D4** — this board only needs a handoff target (navigate away), not the finished composer.  
**Related briefs:** [`S7-D8`](S7-D8-pvcallout-actions.md) — Callout actions (gate CTA); [`S7-D4`](../S7-D4-citation-composer.md) — composer place (do not design it here)  
**Design system layers:** [`docs/design-system-layers.md`](../../../../design-system-layers.md) — inventory in §9  
**Dogfood order:** Design and ship **S7-09** before thick composer work so Add property is clickable early.

> **Shipped delta (S7-09):** Edit label/description reuses the create `PVFormDialog` (board copy said create-only).

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](../README.md) first. When implementing, treat **§9 UI building-block inventory** as binding: compose existing kit pieces; only invent what the table marks **New**.

---

## 1. Objective

Design **Evidence graph chrome updates** now that Citations / Observations exist:

- **Add property** affordance on primary (and bridge) cards
- **Cited-data rows** that make cards grow vertically
- Uncited → cited shell transition once Observations exist
- **Subject refs** visible on cards (designer judgment on density; slight widen OK)
- **Bridge edge summaries** that read as English phrases from cited Properties (not working `label` + type name)
- **No-Artifact** empty / disabled state for citing
- **Connect** disambiguation sheet, then **navigate away** to the citation composer place (Option B) — this board designs the graph-side handoff only

**Do not design the composer layout, viewer, or form.** That is **S7-D4**.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Graph is the only Subject surface | No parallel list of Observations. Rows live on the card. |
| Uncited = zero Observations about the subject | Working `label` is not an Observation. Keep Spike 6 uncited treatment until first Observation. |
| Cards grow | Cited property rows stack; width may widen **slightly** from today’s ~236 / ~188 if density needs it (designer judgment — do not jump to a second column); height grows **without** an inner scroll; edge attachment must still read correctly. |
| Subject refs on cards | Every subject (root and bridge) shows its catalog `ref` (mono / muted). Placement is a design finding (under type, trailing header, etc.). |
| Bridge body reads the edge | Durable bridges prefer an **edge summary phrase** from cited Properties (templates below), not `label` + type title (“Relationship”). Working `label` may stay off the bridge chrome when a summary exists. |
| Citing needs an Artifact | Prefer never opening the graph without Artifacts (Sources list already gates). If the graph is open with zero Artifacts anyway: **disable** cite controls (palette place tools, Connect, every Add property) and show **one** graph-wide callout — not per-button or per-card copy. Recovery → Source page add Artifact. |
| Composer is another place | Add property / connect completion **navigates** (leaves the canvas). Propose transition cue (brief), not a sheet over the graph. |
| Bridges require Citation at create | Connect never creates an uncited bridge. Disambiguation on-graph → composer with two edge Observations pre-scoped. Person→person is always `relationship` (shared events = Event + person→event lines). |
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
| **Subject ref** | Catalog `ref` on every card (primary and bridge). Quiet mono; must remain selectable / copyable in spirit of other ref chrome elsewhere. |
| **Growth / width** | Height **grows indefinitely** with cited rows (and description). **No** scroll-inside-card in Spike 7. Width: start from today’s ~236 (primary) / ~188 (bridge); designer may widen **slightly** if ref + rows crowd — call the chosen widths. Edge attach points follow real body height (EG-8). |
| **Uncited → cited** | After ≥1 Observation about the subject, drop uncited shell (Spike 6 contrast frame). |
| **Bridge cards** | Subordinate chrome remains. Body prefers **edge summary** (see §3.1) once durable; drop redundant working `label` + bare type title when the summary is available. Type mark / quiet type cue may remain for scanability. |
| **Palette / type chrome (implementation)** | Membership, copy, icons, card colors, and line/gradient tokens for **every** shown Subject type (roots and bridges) come from the Interpretation subject registry — design may keep today’s look; implementers must not keep hard-coded kind→style maps as the source of truth (see deployment plan registry section). |

### 3.1 Bridge edge summaries (authoritative)

Endpoint names live on the root cards. The bridge card carries the **relational phrase** between them (direction follows the locked edge Properties). Prefer short English templates seeded with the connect / presentation registry later; board proposes copy, not final L10n keys.

| Bridge | Inputs | Preferred summary (examples) | Notes |
| --- | --- | --- | --- |
| **relationship** | `relationship_type` term; edges `person` → `related_to` | **is the {type} of** — e.g. “is the father of”, “is the spouse of” | Matches model: *person is this type of related_to*. Omit working `label` when this phrase exists. Symmetric terms (spouse, sibling, cousin) still use the same template. |
| **location** | edges `event` → `place` | **took place in** | No disambiguation term; fixed phrase is enough. |
| **participation** | `role` term; edges `person` → `event` | Prefer **role-aware** copy — e.g. “participated as {role}”, “was the {role} at” — with a plain **participated in** fallback when role is missing | Bare “participated in” alone drops the role the disambiguation sheet just collected; still useful as fallback. |

Until durable Observations exist, keep today’s honesty body. After S7-10, replace honesty with the summary (and optional quiet type mark).

#### 3.1.1 Localization (product terms + phrase templates)

**Split (same as subject-type presentation):** the registry declares **which** `L10n` key a product term (or phrase template) uses; the macOS String Catalog / `L10n` holds the **translated strings** for `origin=provenencia`. Do **not** store locale strings in SQLite or in the Go registry. Catalog `property_terms.label` remains the English install seed / fallback (and the only display string for `origin=user`).

| Origin | Who declares the key | Who holds translations |
| --- | --- | --- |
| Product (`provenencia`) | `seedTerms` / presentation next to the registry (expose `L10nKey` over FFI, or a stable convention `propertyTerm.<propertyKey>.<termKey>`) | App String Catalog + typed `L10n` |
| Plugin (`plugin:<id>`) later | Same registry shape on the plugin module | Plugin-shipped locale table (or host merge) resolved by the same client helper — **not** a second hard-coded Swift map |
| User | — | Catalog `label` verbatim |

| Surface | Localization rule |
| --- | --- |
| Bridge phrase templates | Product UI chrome → typed `L10n` with `%@` for interpolated term display names. Plugin bridge types may declare their own template key via registry presentation later. |
| Product term display names | Resolve `L10nKey` → String Catalog; fall back to DB `label` only if missing. |
| User-minted terms (`origin=user`) | DB `label` only. |
| Disambiguation / composer pickers | Same resolver as bridge summaries; term lists use **ComboBox**, not chips. |

Ship product-term `L10n` members + catalog entries with **S7-10** (or earlier if composer pickers need them first). Client API shape: one `displayName(forTerm:)` (and phrase builders) that always goes key → catalog → label — so plugins do not force call-site churn.

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
2. Disambiguation sheet on graph (`role` / `relationship_type` via **ComboBox**; refuse unsupported pairs) — §3.2 matrix. Person→person is **always** a relationship (no shared-event choice).
3. Confirm → navigate to composer with bridge + two edge Observations pre-filled
4. Submit → Back to graph with bridge card + edges; cancel → nothing written
```

Disambiguation is **not** a history entry. Composer place is.

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| EG-1 | **Add property** on primary cards; propose bridge-card variant (bridges may add further properties later). |
| EG-2 | **Cited-data rows** with property + value summary; cards **grow indefinitely** (no inner scroll); slight width bump OK if crowded. |
| EG-3 | Uncited vs cited shell wired to Observations (not to working label alone). |
| EG-4 | No-Artifact gate: one graph-wide callout + disabled cite controls (not per-control messaging); clear recovery to add an Artifact. |
| EG-5 | Connect disambiguation: person→event (`role`), person→person (`relationship_type` only — always relationship), event→place (none), refusals. |
| EG-6 | Handoff to composer is navigational (leave canvas); propose any transient confirmation. |
| EG-7 | Graph a11y: new controls named; rotor/list still works; **do not** design composer a11y here. |
| EG-8 | Edge geometry still attaches sensibly to taller (and slightly wider) cards (note for implementers). |
| EG-9 | Catalog **ref** visible on primary and bridge cards. |
| EG-10 | Durable bridge body uses **edge summary** templates (§3.1); do not lead with working `label` + bare type title when a summary exists. |
| EG-11 | Product `property_terms` and bridge phrase templates resolve via **L10n** / String Catalog (§3.1.1); user terms stay catalog `label`. |

---

## 6. Screen / frame inventory (minimum)

### 6.1 Card chrome states (required board section)

Give **primary subject card chrome** its **own section** on the board (not only one-off frames mixed with connect). Show the same Person (or Event) card across progressive states so implementers can match density and shell transitions:

1. **Empty + uncited** — working label only; Spike 6 uncited shell; **ref** visible; Add property visible per EG-1.
2. **With a description** — uncited or cited shell as product rules dictate; description present, still zero (or no) cited property rows; ref visible.
3. **One cited property** — cited shell; single compact property row; ref visible.
4. **Several cited properties** — at least **five** cited rows so growth / stacking / edge attach read clearly; ref still legible.
5. Optional extras if useful: selected vs idle, Add property quiet vs prominent, a negative-polarity row, crowded header after slight widen.

Also show at least one **bridge card** after durable connect — edge summary per §3.1 (relationship / participation / location), with ref; may live in this section or adjacent. Include one frame of today’s provisional honesty body for contrast.

### 6.2 Graph flows (minimum)

1. No-Artifact gate — frame the **message center** (e.g. top-right) plus the same canvas with palette / Connect / Add property all disabled; no stacked per-card callouts.
2. Connect disambiguation (person→person → `relationship_type` ComboBox only).
3. Annotation of navigation away (graph → composer) without drawing the composer.

---

## 7. Out of scope

- Composer layout, PDF viewer, polygon tools, observation form — **S7-D4**.
- Conflicted / competing Observation badges (later honesty spike).
- Pinning a Citation while staying on the graph.
- **Scroll-inside-card** / max-height + clip for cited rows — grow only for now.

---

## 8. Deliverable

Claude Design board + short notes for S7-09 / S7-10. Archive this brief when done.

---

## 9. UI building-block inventory (binding for design + implement)

Provenencia UI is layered as **components / recipes / snowflakes** ([`docs/design-system-layers.md`](../../../design-system-layers.md)). This table is the repo source of truth for *what* this board may introduce. Paths are from `macos/App/` unless noted.

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
| Subject type marks | Recipe | **Ship** | `DesignSystem/Recipes/Marks/` (`PVMark` / `subject_*`; see colocated `MARKS.md`) | Card / palette icons. Pigment still follows registry / kind style — not Lucide. Landed in **S7-12**; this board does not redesign marks. |
| Create-subject form dialog | Component | Ship | `DesignSystem/Components/FormDialog/PVFormDialog.swift` | Already used for place/create. Disambiguation (§9.3) should reuse this chrome, not a hand-rolled sheet. |
| Create-subject form body | Snowflake | Ship | Private form inside `EvidenceGraphView` | Label field for new Person/Event/Place. Connect durable path may still land here after disambiguation, or skip straight to composer — board proposes. |
| Armed / connect hint banner | Snowflake | Ship | Private in `EvidenceGraphView` | Keep lightweight; not a new kit control. |
| Edge + rubber-band paint | Snowflake | Ship | `Features/EvidenceGraph/EvidenceGraphEdgeLayer.swift` (+ GraphCanvas edge geometry) | Recompute attach points when cards grow. |
| Feedback toast | Component | Ship | `DesignSystem/Components/Toast/` (+ vocabulary toast overlay) | Invalid connect already toasts; optional for gates if not a callout. |

### 9.2 Card chrome (S7-09 — Add property + cited rows)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| Primary subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | Baseline width ~236; **slight widen OK** (designer calls it). Height grows indefinitely with description + cited rows; show **ref**; keep uncited → cited shell. Private chrome today (`EvidenceSubjectCardChrome`). Board must include the **§6.1** state strip. |
| Bridge / relationship card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceBridgeCard.swift` | Subordinate width ~188; slight widen OK. Replace provisional honesty body with **§3.1 edge summary** when durable; show **ref**; de-emphasize or hide working `label` when summary exists. |
| Kind → ink / gradient map | Snowflake (debt) | **Extend** | `Features/EvidenceGraph/EvidenceSubjectKindStyle.swift` | Implementation must move the source of truth to the Interpretation subject registry (deployment plan). Design may keep today’s look. |
| **Add property** control | Component + Snowflake | **New** (control uses kit) | Affordance on card: prefer `DesignSystem/Components/Button/PVButton.swift` or `IconButton/PVIconButton.swift`; placement/wiring stays in `EvidenceSubjectCard` | Quiet on selected/activated card (board proposes always-visible vs activated-only). Opens navigation — not a sheet. |
| **Cited-data row** | Snowflake | **New** | Prefer `private` row view colocated with `EvidenceSubjectCard` (e.g. `EvidenceCitedPropertyRow`) | Property label + value summary + optional polarity. Canvas-zoom density is a design finding. **Do not** promote to a recipe until a second call site exists. |
| **Bridge edge summary** | Snowflake (+ registry later) | **New** | Phrase builder colocated with bridge card / graph model; templates should eventually live next to `seedConnect` / presentation in `subjectvocab` | Implement from §3.1; interpolate **localized** product term names (§3.1.1), not raw DB English. Do not hard-code endpoint names into the bridge body. |
| Card growth (no inner scroll) | Snowflake | **Extend** (behavior on card) | Same card files | Always grow with content. **Do not** implement scroll-inside-card. Update `edgeLayoutHeight` / attach math with real body height (and any new width). |

### 9.3 Gates, connect handoff (S7-09 gate + S7-10)

| Building block | Layer | Status | Home | Notes for Claude Design / implementers |
| --- | --- | --- | --- | --- |
| **No-Artifact** message center | Component + Snowflake | **New** (compose kit) | One `PVCallout` hosted in `EvidenceGraphView` — prefer a fixed corner (e.g. top-right), away from the floating palette | **Single** graph-wide explanation when Artifact count is zero + recovery action → Source page. Requires **S7-D8** / **S7-14** (`PVCallout` actions slot) — do not hand-roll a banner. **Disable** (do not hide) cite controls: Person/Event/Place tools, Connect, and every Add property. **Do not** attach callouts under each button or on each card. |
| **Connect disambiguation** sheet | Component + Snowflake | **New** (body) | Chrome: `PVFormDialog`. Body: private snowflake under `Features/EvidenceGraph/` | person→event → `role` ComboBox; person→person → `relationship_type` ComboBox only (**always** relationship — no shared-event fork); event→place → confirm / skip straight to composer; refusals toast. **Not** a history entry. Confirm → navigate to composer with pre-scoped bridge + edge Observations. |
| Term pickers (`role`, `relationship_type`) | Component | Ship | `DesignSystem/Components/ComboBox/PVComboBox.swift` | **Always** ComboBox — options are catalog `property_terms` (product + user) and can grow. Do **not** use `PVChip` for these lists. Load from store/FFI; display via §3.1.1 L10n. |
| Navigation handoff (graph → composer) | Platform (not DS) | **New** place later | `WorkspaceNavigation` + future composer `WorkspaceLocation` (S7-D4 / S7-08) | Board only annotates leave-canvas cue. Do not mock composer UI here. |

### 9.4 Explicit non-goals for this inventory

| Do not add | Why |
| --- | --- |
| New `DesignSystem/Components/*` for “graph card” or “cited row” | Cards and rows are Evidence-graph snowflakes until a second product surface needs them. |
| Composer viewer / observation form / term picker | **S7-D4**. |
| Subject fields admin chrome | **S7-D2** (shipped / archived). |
| Replacing subject marks with Lucide / SF Symbols | Curated pack stays at `Recipes/Marks/` (**S7-12** landed); this board does not redesign marks. |
| Scroll-inside-card / clipped max height | Spike 7 grows cards only; revisit later if real Sources force it. |

### 9.5 Suggested implement order (matches dogfood)

1. **Extend** `EvidenceSubjectCard` — Add property + cited-row snowflake + **ref** + **indefinite growth** (no inner scroll) + uncited shell; apply any agreed slight widen. Match **§6.1** states.  
2. **No-Artifact** message center (`PVCallout` once, **actions** from **S7-14**) + disable cite controls from one model flag.  
3. **Extend** `EvidenceBridgeCard` — **ref** + **§3.1 edge summary** once durable Observations exist (honesty until then); product-term + phrase **L10n** (§3.1.1).  
4. **Connect disambiguation** snowflake inside `PVFormDialog` → **PVComboBox** for `role` / `relationship_type` → navigate to composer place (same term L10n).
