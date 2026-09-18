# S6-D1 — Evidence graph canvas + bubbles

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 6 (Evidence graph canvas)  
**Implements later as:** PRs S6-02, S6-03  
**Depends on:** Spike 5 Evidence graph place (stub exists)  
**Related brief:** [`S6-D2`](S6-D2-connect-edges.md) (edges — design after or in parallel, but implement after bubbles)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design the **visual system for the Evidence graph canvas** so engineering can replace the stub with a spatial surface that already looks like the product:

- empty / sparse graph
- tool palette to add **Person**, **Event**, **Place**
- root **bubbles** (default, selected, dragging)
- **tray** for unplaced subjects
- chrome that still reads as Sources-family work (Evidence graph title), not a new app mode

S6-01 may ship an empty scrollable grid before this board is done; **bubble and palette chrome wait on this brief.**

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| One Source per graph | Title/context is this Source; no multi-Source chrome. |
| Three placeable roots | Person / Event / Place are presentational variants of one bubble pattern — not three products. |
| Graph is the only Subject surface | Do not design a sibling subject list or view toggle. |
| Unplaced → tray | No auto-layout. Absence of a position = tray chip/row. |
| Layout is research arrangement | Bubbles sit on a calm grid; avoid dashboard chrome. |
| Accessibility is first-class | Selection and type must be obvious visually *and* nameable for VoiceOver (labels, roles) — engineering will map to `accessibilityRepresentation`. |

### 2.1 What this surface is not

- Not the Source filing page.
- Not connect / edge design — **S6-D2**.
- Not citation composer, property inspector, or artifact viewer.
- Not a family tree (no generational layout).

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| CB-1 | Full-bleed **canvas** inside the workspace content area; Evidence graph naming in chrome/breadcrumb as already established. |
| CB-2 | Subtle **grid** that supports snap placement without dominating. |
| CB-3 | **Palette** (or equivalent) for Add Person / Add Event / Add Place — clear, few controls, Mac-native density. |
| CB-4 | **Root bubble** anatomy: type cue (icon/color/shape — reuse Evidence icons where possible), working **label**, selected and dragging states. |
| CB-5 | Person / Event / Place must be distinguishable at a glance without reading fine print. |
| CB-6 | **Empty state:** short guidance to add a first subject (no “coming soon”). |
| CB-7 | **Tray** along an edge for unplaced subjects; drag-onto-grid affordance should be obvious. |
| CB-8 | Stay on existing tokens / type ramp; no purple-glow “node editor” cliché; no card farm. |
| CB-9 | Show at least one frame with **several** bubbles so density and overlap risk are visible. |

---

## 4. Screen / frame inventory (minimum)

1. Empty Evidence graph (palette visible).
2. Graph with a few Person / Event / Place bubbles (one selected).
3. Bubble close-ups: default / selected / dragging.
4. Tray with one or more unplaced subjects.
5. Optional: palette-only detail if the control is non-obvious.

---

## 5. Out of scope

- Relationship lines and connect cursor — **S6-D2**.
- Delete confirmation, conflict/uncited badges, filters.
- Source-page “Open Evidence graph” control.
- Subject types / Subject fields editors.

---

## 6. Acceptance checks

- [ ] A stranger can tell Person vs Event vs Place without a legend essay.
- [ ] Selected vs idle is obvious at thumbnail size.
- [ ] Tray is discoverable in the empty-adjacent and populated frames.
- [ ] Nothing in the frames implies a second tabular Subjects destination.
- [ ] Ready to hand to engineering for S6-02 / S6-03 without inventing chrome in code review.
