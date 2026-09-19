# S6-D1 — Evidence graph canvas + primary subject cards

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 6 (Evidence graph canvas)  
**Implements later as:** PRs **S6-02**, **S6-03** only  
**Depends on:** Spike 5 Evidence graph place (S6-01 shell may already exist)  
**Related brief:** [`S6-D2`](S6-D2-connect-edges.md) — connect, bridge/relationship cards, and cited-data-on-card growth

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the **primary** Evidence graph surface for Spike 6’s first UI PRs:

- empty / sparse graph
- **tool palette** — Add Person / Add Event / Add Place as **mutually exclusive tools** (toggle one on)
- **click-to-place** → **create modal** (type from tool; label, description, …)
- **primary subject cards** — Person / Event / Place as floating rounded cards (not round graph-viz bubbles)
- chrome that still reads as Sources-family work (Evidence graph title)

**Keep this board small.** Connect, relationship mid-nodes, and cited-data rows on cards are **S6-D2** (and later Observation work) — do not expand D1 into that.

S6-01 may ship an empty scrollable grid before this board is done; **primary card + palette chrome wait on this brief.**

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| One Source per graph | Title/context is this Source; no multi-Source chrome. |
| Three primary roots | Person / Event / Place share one **primary card** pattern; differentiate with **icon + color tinge**, not three unrelated layouts. |
| Graph is the only Subject surface | Do not design a sibling subject list or view toggle. |
| Place then fill | Position from the canvas click; type from the armed tool. Modal is for create-time fields (label, description, …). Citation is **not** required to create a primary. |
| Uncited primaries are common | A Person/Event/Place with only a working label (no Observations) is valid but thin. The card needs a clear **uncited** visual — ghost / shell / badge — so the board does not look “done.” |
| Cards will grow later | Shape should *allow* vertical growth (rounded floating card), but **this board does not design populated cited-data rows** — that is S6-D2 / later. Once a primary has any Observation, it leaves the uncited treatment (exact threshold: Observations about this subject). |
| Layout is research arrangement | Cards on a calm grid; avoid dashboard chrome. |
| Accessibility is first-class | Type, selection, tool mode, and **uncited vs cited** must be nameable for VoiceOver. |

### 2.1 What this surface is not

- Not round node-editor “bubbles.”
- Not **connect**, edge lines, or **bridge / relationship cards** — **S6-D2**.
- Not cited-data rows, “Add cited data,” or citation composer — **S6-D2** (visual) / later slices (wiring).
- Not an unplaced-subjects tray.
- Not a family tree.

### 2.2 Implementation gate (do not grow S6-02 / S6-03)

| Ships in S6-02 / S6-03 | Does **not** ship there |
| --- | --- |
| Primary card: icon, color tinge, label, selected / dragging | Bridge cards, connect tool |
| **Uncited** chrome on primaries (all new cards start this way until Observations exist) | True “became cited” transition wiring (needs Observations later) |
| Palette tools + create modal + place + drag | Cited-data rows, add-property control, full citation affordances |
| Empty / header-only card body is fine | Growing multi-row dossiers |

---

## 3. Primary card model (authoritative for D1)

Prefer a **floating card** with modest rounded corners over a circle or pill. Reserve the idea that the body can grow later; for D1, a **header-forward** card (icon + label, quiet empty body) is enough.

| Element | Intent |
| --- | --- |
| **Icon** | Clear Person / Event / Place marks. Existing Evidence icons are file/source-type — **likely three new subject icons** (reusable later for Conclusion canonical entities). |
| **Color** | Light tinge per kind (wash / border / accent). On (or extending) the design-system ramp — no neon / purple-glow. |
| **Header** | Icon + working **label** (+ optional quiet type name). |
| **Uncited state** | Default for a primary with **no Observations**. Propose: ghosted / desaturated shell, dashed border, or a small “Uncited” badge / mark — readable at canvas zoom, not a shouty error. Must stay distinct from selected/dragging. VoiceOver: include uncited. |
| **Cited shell** | After at least one Observation about the subject, drop the uncited treatment (body may still be sparse until rows exist — D2). Contrast frame: same card family, solid/confident chrome. |
| **States** | Default (uncited or cited), selected, dragging (optional hover). |

**Bridges never use this treatment** — Connect mandates a Citation and writes endpoint Observations at create time (**S6-D2**). Do not invent an uncited bridge card.

---

## 4. Create flow (authoritative)

```text
1. Toggle palette tool: Add Person | Add Event | Add Place
2. Canvas shows tool armed (propose how)
3. Click empty grid → that point is the position
4. Modal: type already set from the tool (show it). Collect label, description, …
5. Confirm → **uncited** primary card at the click; cancel → no subject
6. Toggle off / Escape / switch tool clears armed mode
```

Citation is not part of this modal. Uncited is expected and honest until Add cited data / later Observation work.
---

## 5. Requirements

| ID | Requirement |
| --- | --- |
| CB-1 | Full-bleed **canvas**; Evidence graph naming as established. |
| CB-2 | Subtle **grid** for snap placement. |
| CB-3 | **Palette** of three toggle tools; clear armed vs idle. |
| CB-4 | **Primary card** per §3 — floating rounded card; icon + color + label; default / selected / dragging. |
| CB-4a | **Uncited** visual for primaries with no Observations (ghost / shell / badge — propose). A11y name includes uncited. |
| CB-4b | At least one frame of the **cited shell** contrast (same card, no uncited treatment) even if D1 never wires the transition. |
| CB-5 | Person / Event / Place obvious via **icon + color** without a legend. Note if new icons are required. |
| CB-6 | **Empty graph** guidance: pick a tool, click the grid. |
| CB-7 | **Create modal** after place-click: type fixed/shown; label, description, other create-time fields. Prefer `PVDialog` / existing forms. |
| CB-8 | Tokens: shipped system; no purple-glow node-editor cliché. |
| CB-9 | Frames: idle palette; tool armed; modal; sparse graph of primaries; card state close-ups. |

---

## 6. Screen / frame inventory (minimum)

1. Empty Evidence graph (palette visible; no tool armed).
2. Tool armed.
3. Create modal (type shown; label, description; confirm/cancel).
4. Graph with Person / Event / Place primary cards (differentiation) — **uncited** chrome.
5. Card states: default (uncited) / selected / dragging.
6. Contrast: one **cited-shell** primary (no uncited treatment) next to an uncited peer.
7. Optional: proposed Person / Event / Place icon marks.

---

## 7. Out of scope (→ S6-D2 or later)

- Connect tool, lines, bridge/relationship cards (bridges are never uncited — D2).
- Cited-data rows, add-property control, citation peek / composer.
- Unplaced tray; delete confirm; honesty badges (negated/conflicted); filters.
- Source-page graph entry; Subject types/fields admin.

---

## 8. Acceptance checks

- [ ] Cards read as floating rounded dossiers, not round graph-viz nodes.
- [ ] Uncited primaries are obviously “shell / not yet evidenced,” not broken and not finished research.
- [ ] Person vs Event vs Place is obvious via icon + color (including when uncited).
- [ ] Armed tool → click → modal → uncited card is clear without a tutorial.
- [ ] No connect / bridge / cited-row chrome sneaks into this board.
- [ ] Ready for **S6-02 / S6-03** without inventing extra product surface area.
