# S6-D2 — Connect + bridge cards

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 6 (Evidence graph canvas)  
**Implements later as:** PR **S6-04** (connect gesture, lines, bridge card chrome)  
**Depends on:** S6-D1 (primary card language done)  
**Related brief:** [`S6-D1`](archive/S6-D1-canvas-bubbles.md) (completed)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **connections** and **bridge / relationship cards** so S6-04 can retire drawing and gesture risk **without** building Citations, Observations, or a Source viewer:

- **Connect tool** (distinct from Add Person/Event/Place)
- **Line + mid-card + line** so a relationship is a real subject, not a stroke alone
- **Bridge cards** — same family as primaries, but **visually subordinate** (smaller / quieter)
- **One create step** — label/description sheet only (same pattern as S6-D1 primary create)
- **Honest prototype links** — endpoint association may be session/app-only until a later spike wires Citations; label that so dogfood does not treat lines as cited evidence

Reuse S6-D1 primary chrome; do not redesign Person/Event/Place from scratch.

**Do not design or frame a citation / Source-viewer modal on this board.** That surface is a later spike. This board stops at connect + bridge chrome + label create.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| A drawn line is a macro | Mid-node is a **bridge card** (`relationship` / `participation` / `location`, …), not a bare stroke. |
| Primaries may be uncited | Person/Event/Place may land with label only (**S6-D1** uncited shell). |
| Bridges are not a fourth primary | Bridge cards must never read as peer Person/Event/Place — smaller / quieter (propose). |
| Prototype links are temporary | S6-04 may keep A↔bridge↔B without Citations. Show **honesty copy** on the bridge (e.g. link prototype / not cited yet) — not the D1 dashed uncited primary shell. |
| Zoom / a11y | Lines and mid-cards stay legible slightly zoomed out; connect + mid-card need clear VoiceOver names. |
| Lines follow cards | Moving A, B, or the bridge updates both segments live. |

### 2.1 What this is not

- Not a citation composer, Source viewer, Observation list, or locator UI — **out of this board and out of S6-04**.
- Not a catalog “provisional edge” table / migration — endpoint ids may live in session or FakeStore only.
- Not the D1 **uncited** dashed shell on bridges (that treatment is **primaries only**).
- Not the full disambiguation sheet (person→person → relationship vs shared event) — later / light kind pick OK.
- Not a redesign of the create-primary **label** modal — **S6-D1**; bridges reuse that pattern.
- Not arrowheads that imply concluded genealogy direction unless a light cue is clearly helpful.
- Not cited-data rows / “Add cited data” growth chrome (later, when Observations exist).

### 2.2 Implementation gate (S6-04)

| Ships in S6-04 | Explicitly out |
| --- | --- |
| Connect gesture; line through bridge card; bridge chrome vs primary | Citation / Source-viewer modal |
| Bridge card selected / dragging; a11y for links | Observation / Citation writes |
| Label + description create for the bridge (D1 pattern) | Locator / artifact capture |
| Honesty labeling for prototype links | Live cited-row growth on cards |

---

## 3. Bridge card model (authoritative for D2)

Same floating rounded-card family as S6-D1 primaries, but **subordinate**:

| Element | Intent |
| --- | --- |
| **Scale** | Smaller overall and/or tighter padding; type one step down is on the table — propose what keeps hierarchy obvious. |
| **Chrome** | Muted wash / thinner border / quieter icon so primaries stay the protagonists. Never the D1 **uncited** shell. |
| **Header** | Bridge kind cue + short summary (e.g. “Location”, “Participation”, relationship label — propose). |
| **Body (prototype)** | Quiet honesty that this link is canvas prototype / not yet cited evidence — complete as a *link*, not an empty “add properties” state. |
| **States** | Default, selected, dragging. |

Composition: **primary — line — bridge card — line — primary**.

---

## 4. Connect flow (authoritative for this board / S6-04)

```text
1. Arm Connect (distinct from Add Person/Event/Place)
2. Select primary A, then primary B (click or drag-line — propose; under zoom)
3. Bridge kind if needed (location vs relationship vs … — light; full disambiguation later)
4. Modal — same pattern as S6-D1 primary create: working label + description only
5. Save → create bridge subject + position; record endpoint association (session/app OK);
   draw A — bridge — B; mid-card shows honesty labeling
6. Escape / cancel at any step clears in-progress connect; cancel on label sheet → no bridge
```

**Primary vs bridge:** Primaries may stay uncited (**S6-D1**). Bridges from Connect get a real subject + lines for gesture dogfood; they are **not** presented as cited research until a later spike.

---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| CE-1 | **Connect tool** distinct from Add Person/Event/Place. |
| CE-2 | **In-progress connect** feedback after choosing A (and during drag if used). |
| CE-3 | **Completed edge:** line through **bridge card**; bridge styling per §3. |
| CE-4 | Selected bridge card / edge states. |
| CE-5 | Cancel / Escape; no stranded mid-card. |
| CE-6 | Frame with **two** relationships (crossing/clutter risk). |
| CE-7 | Match S6-D1 tokens; same canvas, not a pasted diagramming app. |
| CE-8 | Bridge body uses **honesty** for prototype links; no D1 uncited shell; no citation modal. |
| CE-9 | **One** create modal: label + description only (D1 pattern). |
| CE-10 | Hierarchy: primaries stay dominant next to quieter bridge cards. |
| CE-11 | Lines stay attached when A, B, or the bridge moves. |

---

## 6. Screen / frame inventory (minimum)

1. Connect mode idle.
2. In-progress: A selected, waiting for B (or drag in progress).
3. Create modal: label + description (bridge; same pattern as primary create).
4. Completed bridge: A — bridge card — B (honesty body OK).
5. Bridge card selected / dragging.
6. Two relationships sharing a subject.
7. Optional: birth Event — location — Place as the canonical example.
8. Optional: zoomed-out canvas with lines still legible.

**Do not include:** citation / Source-viewer frames, locked Observation lists, “Add cited data” from a primary, or populated cited-row growth.

---

## 7. Out of scope

- Citation composer / Source viewer / Observation forms (any modal beyond label+description).
- Catalog provisional-edge migrations.
- Full disambiguation form (unless a minimal kind picker is needed for the board).
- Negated / conflicted honesty styles.
- Auto edge-routing / orthogonal algorithms — straight or simple curves fine.
- Unplaced tray; Subject vocabulary admin.
- Cited-data rows and add-property growth (later spike).

---

## 8. Acceptance checks

- [ ] Connect → label modal → bridge + lines is the whole create path; **no second modal**.
- [ ] Bridge mid-card is obviously “the relationship,” not a fourth primary.
- [ ] Prototype honesty is readable; lines are not presented as cited evidence.
- [ ] Cancel never leaves orphan mid-cards.
- [ ] Frames are ready for **S6-04** gesture / chrome dogfood only.
- [ ] No Source-viewer or Observation-list chrome appears on the board.

---

## 9. Later product (not this board)

When Citations / Observations exist, Connect will chain into a **shared** citation composer (locked endpoint Observations, same surface as primary “Add cited data”). That handoff is **out of S6-D2 and S6-04** — do not invent bridge-only cite UI here, and do not frame it on this board.
