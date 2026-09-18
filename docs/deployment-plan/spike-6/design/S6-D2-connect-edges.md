# S6-D2 — Connect + relationship edges

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 6 (Evidence graph canvas)  
**Implements later as:** PR S6-04  
**Depends on:** S6-D1 (bubble language must exist first)  
**Related brief:** [`S6-D1`](S6-D1-canvas-bubbles.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **how two root bubbles become a relationship** on the canvas: the connect affordance, the **line**, and the **mid-bubble** (bridge subject) so the association reads as a thing — not a mysterious extra node.

This board is for Spike 6’s **UI prototype**. Engineering may not yet write Citations/Observations; the visuals should still match the intended product so we can judge whether connect feels right.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| A drawn line is a macro | Mid-bubble is the bridge (relationship / participation / location family) — visually distinct from Person/Event/Place roots. |
| Prototype honesty | Do not design citation badges or “cited” checkmarks yet; keep the edge calm. |
| Zoom exists | Lines and mid-bubbles must stay legible when slightly zoomed out. |
| A11y | Selected edge / mid-bubble need clear names (“Relationship between A and B”) for VoiceOver mapping. |

### 2.1 What this is not

- Not the disambiguation sheet (person→person → relationship vs shared event) — later.
- Not property chips on bubbles.
- Not arrowheads that imply directed genealogy conclusions the model has not asserted yet — prefer neutral connectors unless a light direction cue is clearly helpful.

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| CE-1 | **Connect tool** (or mode) distinct from Add Person/Event/Place. |
| CE-2 | **In-progress connect:** after choosing A, show that B is the next click (cursor, highlight, or rubber-band). |
| CE-3 | **Completed edge:** line from A through **mid-bubble** to B; mid-bubble uses bridge styling (not a fourth root type lookalike). |
| CE-4 | Selected mid-bubble / edge state. |
| CE-5 | Cancel / escape path for an in-progress connect (visual only is enough). |
| CE-6 | One frame with **two** relationships so crossing/clutter risk is visible. |
| CE-7 | Match S6-D1 tokens; edges should feel like the same canvas, not a diagramming app pasted on. |

---

## 4. Screen / frame inventory (minimum)

1. Connect mode idle (tool selected).
2. In-progress: A selected, waiting for B.
3. Completed: A — mid-bubble — B.
4. Mid-bubble selected.
5. Optional: two links sharing a subject.

---

## 5. Out of scope

- Disambiguation form copy and layout.
- Observation/Citation chrome.
- Negated / conflicted edge styles (honesty slice later).
- Automatic orthogonal routing / avoidance algorithms — straight or simple curves are fine.

---

## 6. Acceptance checks

- [ ] Mid-bubble is obviously “the relationship,” not a third person.
- [ ] In-progress connect is understandable without a tutorial paragraph.
- [ ] Frames are implementation-ready for S6-04.
- [ ] No design implies durable cited evidence beyond what Spike 6 will ship.
