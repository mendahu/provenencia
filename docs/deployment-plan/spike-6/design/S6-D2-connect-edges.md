# S6-D2 — Connect + bridge cards (+ card growth language)

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 6 (Evidence graph canvas)  
**Implements later as:** PR **S6-04** (connect + bridge card chrome). Cited-data **wiring** stays later — this board still designs how cards grow so we don’t paint ourselves into a corner.  
**Depends on:** S6-D1 (primary card language done)  
**Related brief:** [`S6-D1`](S6-D1-canvas-bubbles.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **connections** and **bridge / relationship cards**, and the **card-body growth** model that both primaries and bridges will use when cited data exists:

- **Connect tool** (distinct from Add Person/Event/Place)
- **Line + mid-card + line** so a relationship is a real subject, not a stroke alone
- **Bridge cards** — same family as primaries, but **visually subordinate** (smaller / quieter)
- **Cited data rows** + **Add cited data** affordance on cards (design the target; S6-04 does not have to wire Observations)

Reuse S6-D1 primary chrome; do not redesign Person/Event/Place from scratch.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| A drawn line is a macro | Mid-node is a **bridge card** (relationship / participation / location family). |
| Bridges hold data | Role and metadata live on the mid-card — that is why it is a node. Typically fewer rows than a primary, **not as a hard rule**. |
| Primaries vs connectors | Bridge cards must never read as a fourth peer Person/Event/Place — smaller scale, quieter color, tighter type OK (propose). |
| Cards grow | Floating cards **expand vertically** as cited rows accumulate; keep readable at normal zoom. |
| Zoom exists | Lines and mid-cards stay legible slightly zoomed out. |
| A11y | Connect steps and mid-cards need clear names (“Relationship between A and B”). |

### 2.1 What this is not

- Not the disambiguation sheet (person→person → relationship vs shared event) — later.
- Not a redesign of the create-primary flow — **S6-D1**.
- Not the full citation composer / artifact viewer / locators — later (show row + “view citation” affordance only).
- Not arrowheads that imply concluded genealogy direction unless a light cue is clearly helpful.

### 2.2 Implementation gate (do not grow S6-04 past connect risk)

| Ships in S6-04 | Design here but **not** required to implement in S6-04 |
| --- | --- |
| Connect gesture; line through bridge card; bridge chrome vs primary | Live Observation/Citation writes |
| Bridge card selected/dragging; a11y for links | Full add-property modal + durable cited rows |
| Provisional link persistence (per deployment plan) | Honesty states (negated/conflicted) |

---

## 3. Bridge card model (authoritative for D2)

Same floating rounded-card family as S6-D1 primaries, but **subordinate**:

| Element | Intent |
| --- | --- |
| **Scale** | Smaller overall and/or tighter padding; type one step down is on the table — propose what keeps hierarchy obvious. |
| **Chrome** | Muted wash / thinner border / quieter icon so primaries stay the protagonists. |
| **Header** | Bridge kind cue + working label (e.g. relationship label / role summary — propose). |
| **Body** | Can hold cited rows / role metadata (see §4). Empty bridge after connect is OK for the prototype. |
| **States** | Default, selected, dragging. |

Composition: **primary — line — bridge card — line — primary**.

---

## 4. Cited data on cards (design target; later wiring)

Applies to **primaries and bridges** once Observations exist. Design it here so connect/hierarchy choices account for growth.

| Element | Intent |
| --- | --- |
| **Add cited data** | Control on the card (propose copy) → will open a modal later. |
| **Data row** | Compact property/value sub-element on the card. |
| **Citation affordance** | Way to view supporting citation/detail from the row (propose; not full composer). |
| **Growth** | Frames for 0 / 1 / several rows (card taller). |

S6-04 may ship bridge cards with empty bodies; the board must still show the grown states so we accept the visual system.

---

## 5. Connect flow (authoritative)

```text
1. Arm Connect (distinct from Add Person/Event/Place)
2. Click primary A → waiting for B (propose feedback)
3. Click primary B → bridge card created between them; lines A—bridge—B
4. Escape / cancel clears in-progress connect
```

Optional later: create modal for bridge label/role — only if needed for the prototype; otherwise a default label is fine for S6-04.

---

## 6. Requirements

| ID | Requirement |
| --- | --- |
| CE-1 | **Connect tool** distinct from Add Person/Event/Place. |
| CE-2 | **In-progress connect** feedback after choosing A. |
| CE-3 | **Completed edge:** line through **bridge card**; bridge styling per §3. |
| CE-4 | Selected bridge card / edge states. |
| CE-5 | Cancel / Escape for in-progress connect. |
| CE-6 | Frame with **two** relationships (crossing/clutter risk). |
| CE-7 | Match S6-D1 tokens; same canvas, not a pasted diagramming app. |
| CE-8 | **Cited-data growth** language per §4 (0 / 1 / many rows; add control; citation affordance) on primary and/or bridge — design target for later wiring. |
| CE-9 | Hierarchy proof: primaries remain dominant when a bridge sits between them, even if the bridge has a data row. |

---

## 7. Screen / frame inventory (minimum)

1. Connect mode idle.
2. In-progress: A selected, waiting for B.
3. Completed: A — bridge card — B (empty bridge body OK).
4. Bridge card selected.
5. Two relationships sharing a subject.
6. Primary and/or bridge with **several cited data rows** (grown card) + add-property control.
7. Optional: one data row showing “view citation” affordance.

---

## 8. Out of scope

- Disambiguation form.
- Full citation composer, artifact viewer, locator capture.
- Negated / conflicted honesty styles.
- Auto edge-routing / orthogonal algorithms — straight or simple curves fine.
- Unplaced tray; Subject vocabulary admin.

---

## 9. Acceptance checks

- [ ] Bridge mid-card is obviously “the relationship,” not a fourth primary.
- [ ] Connect in-progress / complete is understandable without a tutorial.
- [ ] Grown cited-data cards are designed, but it is clear S6-04 need not wire Observations.
- [ ] Frames are ready for **S6-04** connect risk without pulling Observation scope into that PR.
