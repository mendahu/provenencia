# S5-D2 — Interpretation Sources list

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-08
**Depends on:** S5-D1 (done — the sidebar destination that lands here), shipped Sources list (S2-04)
**Related brief:** [`S5-D1`](archive/S5-D1-interpretation-nav-entry.md) (done)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **the screen behind the Interpretation sidebar destination: a Source picker.**

Interpretation is Source-scoped, so before there is anything to look at, the researcher has to choose which document they are going to sit down and work through. That choice is this screen.

```text
[ S5-D1 ] ──▶ Interpretation Sources list ──▶ that Source's graph (Spike 6)
                     this board                 stub / placeholder for now
```

The list itself is permanent — it is the real Source picker for the layer. In this spike the destination behind a row is still the canvas, which is not built yet, so a **placeholder / "coming soon" screen** is fine. The app is early development; do not over-design the temporary state. Design the permanent picker, and treat the stub as scaffolding Spike 6 replaces.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Interpretation is Source-scoped | The section root is a **Source picker**, not a list of interpretation objects. There is nothing to show that is not scoped to a Source. |
| Same Sources, different intent | "Which document am I going to work through?" rather than "which document do I have?" The rows are identical; the intent is not. |
| Sources may have no Artifact | Citations require an Artifact, so a Source with none cannot be interpreted yet. **Still show it in the list**, but do **not** open the graph. Offer a shortcut to that Source's detail page so the researcher can add an Artifact. |
| This destination is read-only | Sources are created, renamed, retyped, and deleted in the Sources destination. None of that happens here. |
| **The graph is the only Node surface** | There is no list or table view of Nodes anywhere in the product, now or later. Do not design toward one, and do not imply a row leads to a table of things. |
| **The graph does not exist yet** | Activating a row can land on a stub. Temporary; Spike 6 replaces it with the canvas. |

### 2.1 What this destination is not

- Not a list of Nodes, Citations, or Observations. Those live in the graph, which is the only place they ever live.
- Not a second Sources admin surface — no Add Source, no rename, no delete, no type editing.
- Not a dashboard. No charts, no progress rings, no "research completeness" scoring.

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| SL-1 | Reuse the shipped **`PVList`** evidence-list pattern from the Sources list (S2-04): leading thumbnail slot, primary title, secondary meta. Do **not** use `PVTable`, and do **not** design a new list component. |
| SL-2 | Row content: **title**, **`SRC-…`** in mono, and **source type name** — the same slots the Sources list uses. |
| SL-3 | Make the destination's *purpose* legible above the list, since the rows look identical to another destination's rows. A short header or description line explaining that picking a Source opens its interpretation graph is enough. Propose the copy. |
| SL-4 | Activating a row that **has at least one Artifact** navigates toward **that Source's interpretation graph** — never the Source page. In this spike that destination may be a placeholder; design the permanent behaviour and a light stub for now. |
| SL-5 | Sources with **no Artifact** stay in the list (do not filter them out). The row itself must **not** open the graph / stub — no click-through on the row body. |
| SL-5a | On a no-Artifact row, provide a clear shortcut that navigates to **that Source's detail page** (Sources section) so the researcher can add an Artifact. Propose the control (link, button, inline action) and copy. This is the one allowed jump *out* of Interpretation from this list. |
| SL-5b | Make the blocked state legible: it should be obvious *why* the row does not open the graph (needs an Artifact first), without reading as a hard error or warning banner. |
| SL-6 | Empty state: no Sources in the project at all. The CTA should send the researcher to the Sources destination to add one — this list cannot create Sources. Reuse `PVEmptyState`. |
| SL-7 | Search/filter is **optional**. Include it only if it is a straight reuse of the Sources list's `PVInput` pattern; do not design a new filtering model. |
| SL-8 | **No graph-specific columns on the rows** — no node counts, no "3 uncited," no last-worked timestamp. These are wanted eventually and are deliberately out of scope: they are derived counts, so placing a single Node would stale this list and force a cache invalidation between the graph and its own landing page. Design the row without them and do not leave a visible gap where they would go. |
| SL-9 | The stub behind a row can be a plain "coming soon" / not-yet-available screen. Keep it cheap — one short line of copy is enough. Do not invent an illustrated empty state or marketing panel for it. |

---

## 4. Screen / frame inventory (minimum)

1. Interpretation Sources list, populated — mixed types, a couple with thumbnails, at least one without; include **at least one no-Artifact row** alongside openable rows.
2. Close-up / annotation of the no-Artifact row: blocked click-through + shortcut to the Source page (SL-5…SL-5b).
3. Empty state (no Sources in the project at all).
4. The placeholder destination behind an openable row (SL-9) — one frame is enough.

---

## 5. Out of scope

- The sidebar destination that lands here — **S5-D1**.
- The interpretation graph itself, and anything spatial — Spike 6.
- A Source-page control that jumps into the layer — deferred to Spike 6 with the graph.
- Any list or table view of Nodes — the graph is the only Node surface, permanently.
- Citations, Observations, properties, artifact viewing, transcription.
- Creating, editing, or deleting Sources from this destination.
- Node counts or any derived per-Source statistic (see SL-8).

---

## 6. Acceptance checklist

- [ ] The list reuses `PVList` and the Sources row anatomy — no new component, no `PVTable`.
- [ ] It is visually obvious *why* this destination exists, despite rows resembling the Sources list.
- [ ] Activating a row with an Artifact is clearly aimed at that Source's graph, even if the stub is temporary.
- [ ] A no-Artifact Source is still listed, cannot open the graph, and has a shortcut to its Source detail page.
- [ ] Empty state points at the Sources destination rather than offering to create a Source here.
- [ ] No node counts, no progress indicators, no dashboard chrome.
- [ ] Nothing on any frame implies that Nodes exist, are cited, or can be created yet.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| List data | Reads the **existing** `.sourcesList(project:)` query key — sharing a key is encouraged, so SL-1 / SL-2 cost nothing at the data layer. SL-8 is what would break that. |
| Section root is load-bearing | This is why the list ships before the graph exists. `WorkspaceSidebar` builds items from `WorkspaceSection.allCases`, so adding the section automatically adds a sidebar row pointing at `.sectionRoot(section)` — something has to render there. It is also the error-recovery destination once deep places exist. |
| Stub destination | A deep place that lands on a placeholder view is fine for this spike. Spike 6 replaces the view; keep the place identity stable if convenient. |
| No-Artifact gate (SL-5…) | Requires knowing artifact presence per row. Confirm the existing list payload carries it (or extend the query) before implementing. The shortcut is cross-section navigation: `WorkspaceLocation(section: .sources, sourceId: X)`. |
