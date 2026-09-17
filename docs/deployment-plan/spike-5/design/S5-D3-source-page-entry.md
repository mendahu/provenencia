# S5-D3 — Source page entry control

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-08 (list + Source-page entry)
**Depends on:** Shipped Source page (S2-23, exported board at `archive/spike-2/design/boards/source-page.dc.html`)
**Related briefs:** [`S5-D1`](S5-D1-interpretation-nav-entry.md), [`S5-D2`](S5-D2-interpretation-sources-list.md), [`S5-D4`](S5-D4-source-nodes-destination.md) (where this control goes)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Add **one control to the existing Source page** that takes the researcher from the Source they are looking at straight into interpreting it.

This is the second of the two ways into the layer. The first is the sidebar destination (S5-D1) and its Source picker (S5-D2); this one skips the picker for the Source already open.

```text
Sources → a Source page ──── "Open interpretation graph" ───▶ [ S5-D4 ]
                                     this board
```

**Scope discipline matters more than usual here.** The Source page is a shipped, dense, already-designed screen. This board adds one control to it. Do not redraw the page, restructure its header, or revisit decisions made in S2-23.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| This is **navigation**, not mutation | It changes nothing about the Source. It must not sit in the same visual class as Save, Delete, or Add. |
| It crosses sidebar sections | Activating it moves the researcher into a **different** sidebar section for the same Source. Back returns to the Source page, and the sidebar selection changes under them — worth annotating even though the nav behaviour is not yours to design. |
| The destination is a plain list today, a canvas in Spike 6 | The label must stay true after Spike 6 turns that destination into a spatial graph, and must not promise editing capability that does not exist yet. |
| A Source with no Artifact can still be interpreted | Nodes do not require an Artifact; Citations will. Do not disable or hide the control for Sources with no file attached. |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| SP-1 | Add an **Open interpretation graph** control to the Source page. The identity header's meta row already carries the type pill and its edit button, which is the natural neighbourhood — but propose the placement you think is right and say why. |
| SP-2 | Weight it as a **navigation** affordance against the page's existing controls. It must not read like Save / Delete / Add. |
| SP-3 | Propose the **label**. It must survive Spike 6, when the destination becomes an actual spatial graph, and must not imply capability that does not ship in this spike. |
| SP-4 | Show the control **in context on the existing Source page board** — do not redraw the page. A crop of the relevant region is enough. |
| SP-5 | Show the control's interactive states (rest, hover, pressed, focus) only insofar as they differ from the shipped component you are reusing. If it is a straight `PVButton` / `PVIconButton` reuse, say so and skip the state matrix. |
| SP-6 | Annotate the navigation: which sidebar section the researcher lands in, and what Back does. Behaviour note only — the destination itself is S5-D4. |

### 3.1 Explicitly deferred

| ID | Requirement |
| --- | --- |
| SP-7 | No node count, no "already started" indicator, no badge on the control. Derived counts are out of scope for the whole spike, and this control must look identical for a Source with no Nodes and a Source with fifty. |
| SP-8 | No secondary actions, menu, or split button. One control, one destination. |

---

## 4. Screen / frame inventory (minimum)

1. Source page (existing board) with the control in place, shown in context.
2. Detail crop of the control and its immediate neighbours.
3. Annotation of the navigation path and Back behaviour, with S5-D4 as a stub frame.

---

## 5. Out of scope

- The sidebar destination and its label — **S5-D1**.
- The Interpretation Sources list — **S5-D2**.
- The destination this control opens — **S5-D4**.
- Any other change to the Source page: layout, metadata, type pill, notes, artifacts.
- The canvas and anything spatial — Spike 6.

---

## 6. Acceptance checklist

- [ ] Exactly one control is added, with a proposed label and placement backed by a reason.
- [ ] It does not read as a mutation against the page's existing controls.
- [ ] It is shown in context on the existing board rather than on a redrawn page.
- [ ] The label still makes sense once the destination becomes a canvas.
- [ ] No count, badge, or "started" state on the control.
- [ ] Nothing else on the Source page changed.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| Navigation injection | Nothing in the Source page feature currently holds `WorkspaceNavigation`; S5-08 injects it. |
| History | The jump is a first-class navigation, so it must go through `WorkspaceNavigation.go(to:)` rather than setting a selection directly — otherwise Back breaks. See the `add-workspace-location` skill. |
| Shared PR | S5-08 implements this brief and S5-D2 together; both must be reviewable before it opens. |
