# S5-D1 — Interpretation nav entry

> **Superseded.** Do not implement. Product IA moved to a unified Sources family with nested config and an **Evidence graph** deep place — see [`../S5-D3-sources-section-nav.md`](../S5-D3-sources-section-nav.md). Kept for history of the Claude Design board only.

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** ~~PR S5-07~~ — cancelled
**Depends on:** Shipped workspace chrome (S2-01)
**Related brief:** [`S5-D3`](../S5-D3-sources-section-nav.md) (replacement)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **one new top-level sidebar destination: Interpretation.**

This board is deliberately narrow — it is the nav chrome only. The screen behind the destination is S5-D2, a separate board.

```text
Sidebar → Interpretation ──▶ [ S5-D2 ]
          ▲
          this board
```

The work is a label, an icon, a position in the list, and how the destination behaves in the states the sidebar already supports. That is a small board, and it should stay small.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Interpretation is a **layer**, not a view | The label should name the layer the researcher is entering, the way Sources does. The *destination* behind a row is called the Interpretation graph — the section is not. |
| The layer is Source-scoped | The destination lands on a Source picker, not on a list of interpretation objects. Nothing in the nav should promise a cross-Source view. |
| The layer has no data yet in this spike | No count is available, and one should not be invented. |
| Sources / Source types / Source fields are already there | This is an addition to a shipped, working sidebar — match it rather than restyling it. |

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| NAV-1 | Add **Interpretation** as a top-level sidebar destination in the same nav treatment as Sources / Source types / Source fields. Do not introduce grouping headers, disclosure triangles, or nesting that the sidebar does not already have. |
| NAV-2 | Propose the **label**. It should read as the layer being entered, not the view being opened. Note that "graph" and "map" were both considered and rejected for the section name. |
| NAV-3 | Propose an **icon**, drawn from the existing symbol vocabulary where one fits. If nothing fits, say so explicitly — that is a useful finding, not a failure. |
| NAV-4 | Ordering and placement: Interpretation belongs with the layer destinations, not among the vocabulary destinations (Source types / Source fields). Propose where it sits and why. |
| NAV-5 | **No count badge** in this spike. Sources shows one, so leave Interpretation's slot empty rather than inventing a number — and show the empty slot beside a populated one so the sidebar does not read as broken or half-loaded. |
| NAV-6 | Show the destination **selected** and **unselected**, using the shipped selection treatment. |

### 3.1 Explicitly deferred

| ID | Requirement |
| --- | --- |
| NAV-7 | No badge, count, or activity indicator, now or designed-for-later. These would be derived counts, and placing a single Node would stale them. Do not leave a visible gap where one would go. |
| NAV-8 | No second-level nav under Interpretation. Sources are reached through the destination's own screen (S5-D2), not through the sidebar. |
| NAV-9 | Nothing in the sidebar should signal that the layer is incomplete. The destination is permanent; any "graph not built yet" caveat belongs on the stub behind a Source (S5-D2), not in the nav. |

---

## 4. Screen / frame inventory (minimum)

1. Full sidebar with **Interpretation** added, unselected.
2. Same sidebar with Interpretation selected.
3. Detail crop showing Interpretation's empty badge slot beside Sources' populated one.
4. Icon candidates, if more than one is worth considering.

The destination's content may be a stub frame. Do not design it here.

---

## 5. Out of scope

- The Interpretation Sources list this destination lands on — **S5-D2**.
- The interpretation graph, the canvas, and anything spatial — Spike 6.
- A Source-page control that jumps into the layer — deferred to Spike 6 with the graph.
- Any restyling of the existing sidebar, chrome, or window shell.

---

## 6. Acceptance checklist

- [ ] Exactly one destination is added, in the shipped nav treatment, with a proposed label and icon.
- [ ] The empty badge slot reads as intentional rather than as a loading or error state.
- [ ] Placement among the existing destinations is proposed with a reason.
- [ ] Selected and unselected states are both shown.
- [ ] Nothing implies a cross-Source interpretation view, a count, or nested navigation.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| Sidebar case | `WorkspaceSection` is `CaseIterable` and the sidebar builds itself from `allCases`, so the label and icon from NAV-2 / NAV-3 are essentially the entire sidebar PR. |
| Symbol | Needs a `PVSymbol` case if no existing glyph fits — hence NAV-3 asking for an explicit answer either way. |
| Gating | This brief gates S5-07, because that PR commits the label and symbol in code. Stub destination views in S5-07 are exempt: they are compile scaffolding, replaced before anything ships. |
