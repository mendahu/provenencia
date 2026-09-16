# S5-D1 — Interpretation section

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PRs S5-07 (sidebar + places) and S5-08 (list + Source-page entry)
**Depends on:** Shipped workspace chrome (S2-01), Sources list (S2-04), Source page (S2-23)
**Related brief:** [`S5-D2-source-nodes-destination.md`](S5-D2-source-nodes-destination.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **how a researcher gets into the Interpretation layer.** Interpretation is Source-scoped, so before there is anything to look at, the researcher must pick a Source. This board covers the three pieces of that:

```text
Sidebar → Interpretation ──▶ Interpretation Sources list ──▶ a Source's interpretation
                                                                    ▲
Sources → a Source page ──── "Open interpretation graph" ───────────┘
```

1. A new **Interpretation** sidebar destination (label + icon).
2. The **Interpretation Sources list** it lands on — a Source picker whose rows open a Source's interpretation rather than its Source page.
3. An **entry control on the Source page** that jumps to the same place for the Source already open.

The destination all three lead to is the subject of S5-D2. Annotate the jump; do not design that screen here.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Interpretation is Source-scoped | The section root is a **Source picker**, not a list of interpretation objects. There is nothing to show that is not scoped to a Source. |
| Same Sources, different intent | The rows are the same Sources as the Sources destination. The difference is **where they go** and **why you are here** — "which document am I going to work through?" rather than "which document do I have?" |
| Sources may have no Artifact | A Source with no attached file cannot ever be cited (Citations require an Artifact). That is not an error, but it is worth a signal — the researcher is about to sit down with something they cannot cite from. |
| Two sections, one Source | Clicking the Source-page control moves the researcher into a **different sidebar section** for the same Source. Back returns to the Source page. |

### 2.1 What this destination is not

- Not a list of Nodes, Citations, or Observations. Those are per-Source and live behind a row.
- Not a second Sources admin surface — no Add Source, no rename, no delete, no type editing. It is read-only.
- Not a dashboard. No charts, no progress rings, no "research completeness" scoring.

---

## 3. Requirements

### 3.1 Sidebar entry

| ID | Requirement |
| --- | --- |
| I-1 | Add **Interpretation** as a top-level sidebar destination, in the same nav treatment as Sources / Source types / Source fields. |
| I-2 | Propose the **label** and an **icon**. The label should read as the layer, not the view — the researcher is entering a layer of work. Note that "graph" and "map" were both considered and rejected for the section name; the *destination* behind a row is called the Interpretation graph. |
| I-3 | The destination shows **no count badge** in this spike. Sources shows one; leave Interpretation's slot empty rather than inventing a number. Show what an empty badge slot looks like beside a populated one so the sidebar does not read as broken. |
| I-4 | Ordering: Interpretation sits with the layer destinations, not among the vocabulary destinations (Source types / Source fields). Propose placement. |

### 3.2 Interpretation Sources list

| ID | Requirement |
| --- | --- |
| I-5 | Reuse the shipped **`PVList`** evidence-list pattern from the Sources list (S2-04): leading thumbnail slot, primary title, secondary meta. Do **not** use `PVTable`, and do **not** design a new list component. |
| I-6 | Row content: **title**, **`SRC-…`** in mono, and **source type name** — the same slots the Sources list uses. |
| I-7 | Make the destination's *purpose* legible above the list, since the rows look identical to another destination's rows. A short header or description line explaining that picking a Source opens its interpretation is enough. Propose the copy. |
| I-8 | Activating a row navigates to **that Source's interpretation** (S5-D2), not to the Source page. |
| I-9 | Optional, if it can be done without clutter: a quiet signal on rows for Sources with **no Artifact** (nothing to cite from yet). Not an error state, not a warning colour — an informational affordance at most. Propose or explicitly decline. |
| I-10 | Empty state: no Sources in the project at all. The CTA should send the researcher to the Sources destination to add one — this list cannot create Sources. Reuse `PVEmptyState`. |
| I-11 | Search/filter is **optional**. Include it only if it is a straight reuse of the Sources list's `PVInput` pattern; do not design a new filtering model. |

### 3.3 Source-page entry control

| ID | Requirement |
| --- | --- |
| I-12 | Add an **Open interpretation graph** control to the Source page. The identity header's meta row already carries the type pill and its edit button, which is the natural neighbourhood — but propose the placement you think is right and say why. |
| I-13 | This is a **navigation** action, not a mutation. It must not read like Save / Delete / Add. Weight it accordingly against the existing controls. |
| I-14 | Propose the label. It must survive Spike 6, when the destination becomes an actual spatial graph, and must not promise editing capability that does not exist yet. |
| I-15 | Show the control in context on the existing Source page board — do not redraw the page. |

### 3.4 Explicitly deferred

| ID | Requirement |
| --- | --- |
| I-16 | **No graph-specific columns on the rows** — no node counts, no "3 uncited," no last-worked timestamp. These are wanted eventually and are deliberately out of scope: they are derived counts, so placing a single Node would stale this list and force a cache invalidation between the canvas and its own landing page. Design the row without them and do not leave a visible gap where they would go. |

---

## 4. Screen / frame inventory (minimum)

1. Sidebar with **Interpretation** added, showing it selected, and showing the empty badge slot next to Sources' populated one.
2. Interpretation Sources list, populated — mixed types, a couple with thumbnails, at least one without.
3. Interpretation Sources list, empty state.
4. Source page (existing board) with the **Open interpretation graph** control in place.
5. Annotation of both navigation paths into the destination, with the destination itself as a stub frame.

---

## 5. Out of scope

- The interpretation destination behind a row — **S5-D2**.
- The canvas, bubbles, drawing, and anything spatial — Spike 6.
- Citations, Observations, properties, artifact viewing, transcription.
- Creating, editing, or deleting Sources from this destination.
- Node counts or any derived per-Source statistic (see I-16).
- A second app chrome or window shell.

---

## 6. Acceptance checklist

- [ ] Sidebar gains one destination with a proposed label and icon, and the empty badge slot reads as intentional.
- [ ] The list reuses `PVList` and the Sources row anatomy — no new component, no `PVTable`.
- [ ] It is visually obvious *why* this destination exists and where a row goes, despite rows resembling the Sources list.
- [ ] Empty state points at the Sources destination rather than offering to create a Source here.
- [ ] The Source page gains one navigation control that does not read as a mutation, shown in context rather than on a redrawn page.
- [ ] No node counts, no progress indicators, no dashboard chrome.
- [ ] Nothing on any frame implies that Nodes are cited or that evidence has been recorded.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| Sidebar case | `WorkspaceSection` is `CaseIterable` and the sidebar builds itself from `allCases`, so the label and icon from I-2 are the entire sidebar PR. Needs a `PVSymbol` case if no existing glyph fits. |
| List data | Reads the **existing** `.sourcesList(project:)` query key — sharing a key is encouraged, so I-5/I-6 cost nothing at the data layer. I-16 is what would break that. |
| Section root is load-bearing | This list is also the error-recovery destination: when a Source is deleted while its interpretation is in history, navigation falls back here. It cannot be skipped. |
| Source-page control | Nothing in the Source page feature currently holds `WorkspaceNavigation`; S5-08 injects it. |
| No-Artifact signal (I-9) | Requires knowing artifact presence per row. Confirm the existing list payload carries it before committing to the affordance. |
