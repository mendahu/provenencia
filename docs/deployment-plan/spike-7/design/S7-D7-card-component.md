# S7-D7 — Card component (design-system reference)

**Kind:** Claude Design board / design-system component  
**Spike:** Provenencia Spike 7  
**Implements later as:** PR **S7-13**  
**Depends on:** Shipped macOS [`PVCard`](../../../../macos/App/DesignSystem/Components/Card/PVCard.swift) (already in the app kit — this brief **documents and references** it in Claude Design, not a greenfield invent)  
**Related:** Evidence graph subject/bridge cards stay **snowflakes** (**S7-D3** / **S7-09**) — out of scope here  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)

Paste this document into Claude Design as the requirements for a **Card** kit board (or component page). Read shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Publish a **Card** component in the Provenencia Claude Design system whose visual contract matches what macOS already ships as `PVCard`:

- Surface tones: card / raised / sunken (`PVColor.surfaceCard` / `surfaceRaised` / `surfaceSunken`)
- Border: solid hairline (`borderSubtle`) or dashed (`borderDefault`, provisional rows)
- Corner radius default `md` (overrideable `sm`)
- Optional shallow elevation (`PVElevation.sm`)
- Optional content padding
- Content-agnostic container — **no** header/footer slots required in this pass (call sites compose titles outside, same as today’s Swift)

Use **implemented** Artifacts / Subject fields / metadata call sites as visual reference — not a speculative redesign.

---

## 2. Implementation gate (S7-13)

| Ships in S7-13 | Does **not** ship there |
| --- | --- |
| Migrate the four manual “good candidates” to `PVCard` (see deployment plan) | Evidence graph subject/bridge/palette chrome (**S7-D3** snowflakes) |
| Align Swift `PVCard` with any board-clarified props if needed | Selectable strip tiles / file-choice / icon-picker cells |
| DesignSystem README already lists Card — keep in sync | Full-bleed pane fills (sidebar, list bands) |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| CD-1 | Board shows Card at default, elevated, dashed, sunken, and `sm` radius. |
| CD-2 | Document that Evidence graph cards are **not** this component (kind tint / selection halo). |
| CD-3 | No new kit primitives beyond Card; tokens already exist. |
| CD-4 | UI inventory: Card = Component **Ship** / light **Extend** if board adds documented slots; S7-13 call sites = Snowflake **Extend**. |

---

## 4. Out of scope

- Graph card growth / cited rows (**S7-D3**)
- Selectable “chip card” API for type strips
- Omnibar overlay elevation (`PVElevation.overlay`) unless the board explicitly unifies it

---

## 5. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S7-13** against the board + existing `PVCard` API.
