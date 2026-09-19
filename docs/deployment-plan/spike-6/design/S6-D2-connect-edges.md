# S6-D2 — Connect + bridge cards (+ card growth language)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 6 (Evidence graph canvas)  
**Implements later as:** PR **S6-04** (connect + bridge card chrome). Citation-composer wiring may land later — this board still designs the **chained create → cite** flow so we don’t invent a second evidence UI.  
**Depends on:** S6-D1 (primary card language done)  
**Related brief:** [`S6-D1`](S6-D1-canvas-bubbles.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **connections** and **bridge / relationship cards**, and how they reuse the **same citation composer** primaries will use:

- **Connect tool** (distinct from Add Person/Event/Place)
- **Line + mid-card + line** so a relationship is a real subject, not a stroke alone
- **Bridge cards** — same family as primaries, but **visually subordinate** (smaller / quieter)
- **Two-step create, one evidence surface:** label/description sheet → immediately the citation / Source-viewer modal (shared with primaries)
- **Prefills:** endpoint Observations locked; optional extra Observations on the same Citation
- **No provisional edge store** — no draft lines without Citations/Observations

Reuse S6-D1 primary chrome; do not redesign Person/Event/Place from scratch.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| A drawn line is a macro | Mid-node is a **bridge card** (`relationship` / `participation` / `location`, …). Edges off it are Observations; every Observation needs a Citation. |
| Primaries may be uncited | Person/Event/Place may land with label only (**S6-D1** uncited shell). They open the citation composer later via Add cited data. |
| Bridges must finish cite in-flow | Connect does **not** leave an uncited bridge with pretend lines. After label/description save, the **same** citation modal opens immediately. |
| One Citation, N Observations | Endpoint edges are separate Observations under one Citation. Prefill + lock those rows; researcher may add more on the same Citation. |
| Often the citation *is* the link | Birth Event — location — Place may need only the two locked endpoint Observations. Extra properties optional. |
| Same modal, two entry points | Bridge connect and primary “Add cited data” share one citation / Source-viewer composer — do not invent a bridge-only cite UI. |
| Primaries vs connectors | Bridge cards must never read as a fourth peer Person/Event/Place — smaller / quieter (propose). |
| Zoom / a11y | Lines and mid-cards stay legible slightly zoomed out; connect + mid-card need clear VoiceOver names. |

### 2.1 What this is not

- Not provisional / draft edge persistence (no second table for “intended” links).
- Not an uncited bridge shell (that treatment is **primaries only** — S6-D1).
- Not the full disambiguation sheet (person→person → relationship vs shared event) — later / light kind pick OK.
- Not a redesign of the create-primary **label** modal — **S6-D1**; bridges reuse that pattern for step 1.
- Not building the full citation composer in S6-04 — **design the handoff and prefilled Observation list** so later wiring drops in.
- Not arrowheads that imply concluded genealogy direction unless a light cue is clearly helpful.

### 2.2 Implementation gate (do not grow S6-04 past connect risk)

| Ships in S6-04 | Design here but **not** required to implement in S6-04 |
| --- | --- |
| Connect gesture; line through bridge card; bridge chrome vs primary | Live Citation / Observation writes |
| Bridge card selected/dragging; a11y for links | Full Source-viewer / locator capture |
| Stub or chained “citation step” chrome if useful for dogfood honesty | Extra freeform Observation rows beyond the locked endpoints |

---

## 3. Bridge card model (authoritative for D2)

Same floating rounded-card family as S6-D1 primaries, but **subordinate**:

| Element | Intent |
| --- | --- |
| **Scale** | Smaller overall and/or tighter padding; type one step down is on the table — propose what keeps hierarchy obvious. |
| **Chrome** | Muted wash / thinner border / quieter icon so primaries stay the protagonists. Never the D1 **uncited** shell. |
| **Header** | Bridge kind cue + short summary (e.g. “Location”, “Participation”, relationship label — propose). |
| **Default body (link-only)** | Cited association after composer completes — citation affordance; not an empty “Add cited data” empty state. |
| **Optional body (extra rows)** | Additional Observations beyond the locked endpoints (role, type, …). |
| **States** | Default, selected, dragging. |

Composition: **primary — line — bridge card — line — primary**.

---

## 4. Shared citation composer (authoritative product shape)

One modal (Source viewer / citation composer) serves:

| Entry | Prefill | Locked |
| --- | --- | --- |
| **Connect → after bridge label save** | Endpoint Observations for the bridge (e.g. `event` + `place`, or `participant` ×2) | Those endpoint Property rows (subject values fixed to A and B) |
| **Primary → Add cited data** (later) | Empty or researcher-chosen Property | Nothing locked by default |

On the bridge path, researcher may **add more** Observation rows on the same Citation (optional). They do not re-pick the two ends.

```text
Cancel rules (design clearly):
- Cancel on label/description sheet → no bridge, no citation.
- Cancel on citation composer → abort whole connect (no orphan bridge / no pretend lines).
  Prefer: don’t persist the bridge until citation saves, or delete it on cite-cancel.
```

No provisional edge catalog — lines appear when endpoint Observations exist (or S6-04 stubs with honesty labeling only).

---

## 5. Connect flow (authoritative)

```text
1. Arm Connect (distinct from Add Person/Event/Place)
2. Select primary A, then primary B (click or drag-line — propose; under zoom)
3. Bridge kind if needed (location vs relationship vs … — light; full disambiguation later)
4. Modal 1 — same pattern as S6-D1 primary create: working label + description only
5. Save on Modal 1 → immediately open Modal 2 (shared citation / Source-viewer composer)
6. Modal 2: Observation list prefilled with locked endpoint subject-valued Properties;
   locator / transcription / other citation fields as the composer defines;
   optional “add another Observation” for extras on this Citation
7. Save Modal 2 → commit Citation + Observations (+ bridge subject); mid-card is a cited link
8. Escape / cancel at any step clears in-progress connect; no stranded bridge
```

**Primary vs bridge:** Primaries may stay on the canvas uncited (**S6-D1**). Bridges from Connect always complete through Modal 2 (or nothing is kept).

S6-04 may stub Modal 2; the **designed** flow still shows the handoff and prefilled locked rows.

---

## 6. Requirements

| ID | Requirement |
| --- | --- |
| CE-1 | **Connect tool** distinct from Add Person/Event/Place. |
| CE-2 | **In-progress connect** feedback after choosing A (and during drag if used). |
| CE-3 | **Completed edge:** line through **bridge card**; bridge styling per §3. |
| CE-4 | Selected bridge card / edge states. |
| CE-5 | Cancel / Escape; no stranded mid-card / no provisional edge store. |
| CE-6 | Frame with **two** relationships (crossing/clutter risk). |
| CE-7 | Match S6-D1 tokens; same canvas, not a pasted diagramming app. |
| CE-8 | **Link-only bridge** after cite — citation affordance; no D1 uncited shell. |
| CE-9 | **Two modals in sequence:** (1) label + description; (2) **shared** citation composer with **prefilled locked** endpoint Observations; optional extra rows. |
| CE-10 | Composer is recognizably the **same** surface primaries will use for Add cited data. |
| CE-11 | **Optional growth** on cards after extras exist; frames for link-only vs “has extra rows.” |
| CE-12 | Hierarchy: primaries stay dominant even with citation chrome or a data row on the bridge. |

---

## 7. Screen / frame inventory (minimum)

1. Connect mode idle.
2. In-progress: A selected, waiting for B (or drag in progress).
3. Modal 1: label + description (bridge; same pattern as primary create).
4. Modal 2: citation / Source-viewer with **two locked endpoint Observations** prefilled; room to add another.
5. Completed **link-only** bridge: A — bridge card — B.
6. Bridge card selected (citation / view-evidence affordance).
7. Two relationships sharing a subject.
8. Bridge **with** extra cited rows beyond endpoints.
9. Optional: same Modal 2 entered from a **primary** (Add cited data) — empty/unlocked — to prove shared chrome.
10. Optional: birth Event — location — Place as the canonical link-only example.

---

## 8. Out of scope

- Full disambiguation form (unless a minimal kind picker is needed for the board).
- Full locator / polygon / artifact tooling fidelity (show the composer shell + Observation list).
- Provisional edge tables or session-only pretend links as a product path.
- Negated / conflicted honesty styles.
- Auto edge-routing / orthogonal algorithms — straight or simple curves fine.
- Unplaced tray; Subject vocabulary admin.

---

## 9. Acceptance checks

- [ ] Connect → label modal → citation modal is one continuous create; no “uncited bridge” resting state.
- [ ] Endpoint Observations are prefilled and locked; extras are optional on the same Citation.
- [ ] Citation UI is clearly the same composer primaries will use later.
- [ ] Link-only bridge looks **complete**, not “missing properties.”
- [ ] Bridge mid-card is obviously “the relationship,” not a fourth primary.
- [ ] Cancel never leaves orphan bridges or pretend lines.
- [ ] Frames are ready for **S6-04** without requiring full Observation wiring in that PR.
