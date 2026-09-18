# Deployment Plan — Spike 6

Evidence graph **canvas prototype**: pan/zoom, place and drag bubbles, draw relationship lines, and do it accessibly. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §7 / §11.3 / §13. Foundation: [Spike 5 archive](../archive/spike-5/).

## Status

**Open.** Landings go in [`completed.md`](completed.md).

> **Goal of this spike:** retire UI risk. If the canvas feels wrong or a11y is intractable, stop here and rethink — do not pile Citations on a failed surface.

## Goal (dogfood bar)

All of the following must be true in the app (VoiceOver / keyboard where noted):

1. Open **Evidence graph** on a Source → real canvas (stub gone).
2. **Zoom in** and **zoom out** (trackpad pinch and/or menu/keyboard); pan with scroll.
3. **Arm a palette tool** (Add Person / Event / Place), **click the grid** → create modal → bubble at that point; subject + position persist across relaunch.
4. **Drag** a bubble; it snaps to the grid; position persists (flush on gesture end, not per frame).
5. **Connect** two bubbles → relationship **line** with a mid-bubble; gesture works under magnification.
6. **Accessibility:** VoiceOver can list subjects (and links), move focus, and perform add / move / connect without relying on spatial pointer alone. Keyboard parity for the same actions.
7. Geometry lives in a **neutral module** (not baked into Interpretation-only types) — design note §13.

**Explicitly not required to pass the bar:** durable Observation/Citation subgraphs, disambiguation sheet, property inspector, unplaced-subjects tray, Source-page entry button.

## Design track (no code PRs)

**All canvas chrome is designed in Claude Design before the matching UI PR.** Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S6-D1** | Canvas + primary cards | Grid, toggle tools, click-to-place + create modal, **Person/Event/Place floating cards** (header chrome; tray/connect/cited-rows out) | S6-02, S6-03 |
| **S6-D2** | Connect + bridge cards | Connect tool, lines, **subordinate bridge cards**, card growth / cited-row language (wiring later) | S6-04 |

Run **S6-D1 before any bubble chrome.** **S6-01** (scroll shell) may start in parallel with S6-D1 — it has no product chrome. **S6-D2 before S6-04.**

## PR sequence

```text
   design                         build
─────────────               ──────────────────────────────────────────

S6-D1  Primary cards          S6-01  NSScrollView shell + coords     ◀── may start now
  │                           │    (replace stub; pan/zoom only)
  │                           ▼
  └────── gates ──────────▶ S6-02  Primary cards + a11y
                              │    (icon/color/label; empty body OK)
                              ▼
                            S6-03  Click-to-place + create modal + drag
                              │
S6-D2  Connect + bridges      │
  │                           ▼
  └────── gates ──────────▶ S6-04  Connect + bridge cards (UI prototype)
                              │    (cited-row wiring later)
                              ▼
                            S6-05  Dogfood close / go-nogo
```

---

## Checklist

- [ ] S6-D1 — Design: canvas + primary cards → [`design/`](design/)
- [ ] S6-D2 — Design: connect + bridge cards → [`design/`](design/)
- [x] S6-01 — `NSScrollView` shell + coordinate conversion → [`completed.md`](completed.md)
- [ ] S6-02 — Primary cards on the canvas + accessibility representation
- [ ] S6-03 — Click-to-place, create modal, drag, snap, persist
- [ ] S6-04 — Connect + bridge cards (accessible)
- [ ] S6-05 — Docs, dogfood, go-nogo

---

## S6-01 — Pan/zoom shell

Replace the Evidence graph stub with an `NSScrollView` bridge (`NSViewRepresentable`), empty grid content, and **one** unit-tested coordinate-conversion seam (design note §7.2).

| | |
| --- | --- |
| **In** | Magnification min/max; pinch zoom; scroll pan; content-space point conversion; camera in app-local state only (not catalog). |
| **Out** | Bubbles, palette, edges, subject RPCs. |
| **Testable** | Zoom in/out by hand; Swift tests for conversion under magnification. |
| **Depends on** | Spike 5 graph place. **Not** gated on S6-D1. |

---

## S6-02 — Primary cards + accessibility representation

Render Person / Event / Place as **floating subject cards** (real SwiftUI views in a `ZStack` over a `Canvas` grid — design note §7.1). Load subjects + positions via the existing graph query key. Ship `accessibilityRepresentation` / children / rotor skeleton **in this PR** (§7.4).

| | |
| --- | --- |
| **In** | Neutral geometry module; primary cards per S6-D1 (icon, color, label, selection); VoiceOver subject list. Empty / header-only body OK. |
| **Out** | Create/drag/connect; bridge cards; cited-data rows; add-property. |
| **Testable** | Open a Source that already has subjects (or seed via FakeStore); VoiceOver sees them. |
| **Depends on** | S6-01, **S6-D1**. |

---

## S6-03 — Click-to-place + drag + persist

Palette tools toggle Add Person / Event / Place. With a tool armed, click the grid → create modal (type fixed from the tool; label, description, other create-time fields) → subject + position at the click. Drag snaps to grid; patch position on gesture end (design note §5.3).

**Tray descoped** — unplaced subjects (no position row) are out of this spike’s UI; revisit when imports / off-canvas creates exist.

| | |
| --- | --- |
| **In** | Toggle tools; place-click; create modal; drag; debounce flush; keyboard add/move/select parity for what the pointer can do. |
| **Out** | Unplaced tray; connect; bridge cards; cited-data rows; delete confirmation UX; label editing polish beyond create. |
| **Testable** | Arm Add Person, click, label, confirm — three bubbles; drag them; relaunch — positions stick; keyboard/VoiceOver can add and move. |
| **Depends on** | S6-02, S6-D1. |

---

## S6-04 — Connect + bridge cards (UI prototype)

Connect tool: click A then B → line through a **bridge card**. Gesture and hit-testing must work under zoom (same coordinate seam). Bridge chrome per S6-D2 (subordinate to primaries). Extend a11y so links are traversable and connect is keyboard-capable.

**Persistence scope (deliberate):** retires *drawing and gesture* risk, not Observation macros. Prefer creating a bridge `subjects` row + position. Endpoint association may be **provisional** until Citations + Observations exist — **no new migration**. Do not pretend provisional links are research data.

**Cited-data rows / add-property:** designed in S6-D2; **not required** to wire in this PR (empty bridge body OK).

| | |
| --- | --- |
| **In** | Connect gesture; edge rendering; bridge card chrome per S6-D2; a11y for links. |
| **Out** | Disambiguation form, Observations, Citations, pinned composer, person→person macro matrix (§3.2), live cited rows. |
| **Testable** | Draw links under zoom; VoiceOver announces the link; relaunch keeps bridge subjects if created (links may be provisional — honesty in S6-05). |
| **Depends on** | S6-03, **S6-D2**. |

---

## S6-05 — Dogfood close / go-nogo

Honesty pass against the [goal bar](#goal-dogfood-bar). Record the verdict in [`completed.md`](completed.md) and update [`README.md`](README.md).

| Verdict | Meaning |
| --- | --- |
| **Go** | Canvas stays the only Subject surface; next spike can take Citations / Observations / real connect macros. |
| **No-go** | Stop. Do not route around with a subject list (§1.3) — fix or replace the canvas approach first. |

No SemVer bump for docs-only close. Archive this folder when the spike is accepted complete.

---

## Scope boundary

| In Spike 6 | Out |
| --- | --- |
| `NSScrollView` + coords | Deployment-target bump |
| Bubbles P/E/P | `source` subjects on canvas (§4.6) |
| Drag / snap (no tray) | Auto-layout; unplaced tray |
| Connect **UI** + lines | Observation/Citation macros |
| Canvas a11y representation | Alternate list destination |
| Neutral `GraphCanvas` (name flexible) | Polymorphic position table (§13.3) |

---

## Gotchas (read before coding)

1. **Magnified hit-testing is haunted** without the shared conversion seam (§7.2). Unit-test it in S6-01; reuse everywhere.
2. **A11y is not polish** (§7.4). If S6-02 ships bubbles without a representation, stop and fix before S6-03.
3. **Positions travel; camera does not** (§5.2). Do not write pan/zoom into the catalog.
4. **Drag must not write per frame** (§5.3). Patch on gesture end.
5. **Geometry stays neutral** (§13). Parameterize over id / type / label / grid cell — do not couple to `CatalogSubject` inside the scroll bridge.
6. **Connect without Observations is provisional.** Label it in UI/docs so dogfood does not treat lines as cited evidence.
7. **Scope creep default:** no minimap, multi-select, alignment guides, or fancy edge routers in this spike (§9).

---

## Definition of done

- [ ] Checklist above complete
- [ ] Dogfood bar items 1–7 met (or explicit no-go recorded)
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
- [ ] Design note status line points at this spike as scheduled (slice 2)

## What the next spike inherits

On **go:** a real Evidence graph place with pan/zoom, placeable/draggable bubbles, accessible structure, and a connect gesture whose durable write path is still Observations — ready for Citations + macros.

On **no-go:** a written verdict and no further Interpretation UI until the canvas question is reopened.
