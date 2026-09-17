# S5-D2 — Interpretation Sources list

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-08 (list + Source-page entry)
**Depends on:** S5-D1 (the sidebar destination that lands here), shipped Sources list (S2-04)
**Related briefs:** [`S5-D1`](S5-D1-interpretation-nav-entry.md), [`S5-D3`](S5-D3-source-page-entry.md), [`S5-D4`](S5-D4-source-nodes-destination.md) (where a row goes)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **the screen behind the Interpretation sidebar destination: a Source picker.**

Interpretation is Source-scoped, so before there is anything to look at, the researcher has to choose which document they are going to sit down and work through. That choice is this screen.

```text
[ S5-D1 ] ──▶ Interpretation Sources list ──▶ a Source's nodes [ S5-D4 ]
                     this board
```

**The hard problem on this board:** the rows are the *same Sources* as the shipped Sources destination, drawn from the same data, and they will look nearly identical. What differs is where a row goes and why the researcher is here. Making that legible without redesigning the row is the work.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Interpretation is Source-scoped | The section root is a **Source picker**, not a list of interpretation objects. There is nothing to show that is not scoped to a Source. |
| Same Sources, different intent | "Which document am I going to work through?" rather than "which document do I have?" The rows are identical; the intent is not. |
| Sources may have no Artifact | A Source with no attached file can never be cited — Citations require an Artifact. Not an error, but the researcher is about to sit down with something they cannot cite from. |
| This destination is read-only | Sources are created, renamed, retyped, and deleted in the Sources destination. None of that happens here. |

### 2.1 What this destination is not

- Not a list of Nodes, Citations, or Observations. Those are per-Source and live behind a row (S5-D4).
- Not a second Sources admin surface — no Add Source, no rename, no delete, no type editing.
- Not a dashboard. No charts, no progress rings, no "research completeness" scoring.

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| SL-1 | Reuse the shipped **`PVList`** evidence-list pattern from the Sources list (S2-04): leading thumbnail slot, primary title, secondary meta. Do **not** use `PVTable`, and do **not** design a new list component. |
| SL-2 | Row content: **title**, **`SRC-…`** in mono, and **source type name** — the same slots the Sources list uses. |
| SL-3 | Make the destination's *purpose* legible above the list, since the rows look identical to another destination's rows. A short header or description line explaining that picking a Source opens its interpretation is enough. Propose the copy. |
| SL-4 | Activating a row navigates to **that Source's nodes** (S5-D4), not to the Source page. |
| SL-5 | Optional, if it can be done without clutter: a quiet signal on rows for Sources with **no Artifact** (nothing to cite from yet). Not an error state, not a warning colour — an informational affordance at most. Propose or explicitly decline. |
| SL-6 | Empty state: no Sources in the project at all. The CTA should send the researcher to the Sources destination to add one — this list cannot create Sources. Reuse `PVEmptyState`. |
| SL-7 | Search/filter is **optional**. Include it only if it is a straight reuse of the Sources list's `PVInput` pattern; do not design a new filtering model. |

### 3.1 Explicitly deferred

| ID | Requirement |
| --- | --- |
| SL-8 | **No graph-specific columns on the rows** — no node counts, no "3 uncited," no last-worked timestamp. These are wanted eventually and are deliberately out of scope: they are derived counts, so placing a single Node would stale this list and force a cache invalidation between the node destination and its own landing page. Design the row without them and do not leave a visible gap where they would go. |

---

## 4. Screen / frame inventory (minimum)

1. Interpretation Sources list, populated — mixed types, a couple with thumbnails, at least one without.
2. Same list showing the no-Artifact affordance, if SL-5 is accepted.
3. Empty state.
4. Annotation of where a row goes, with S5-D4 as a stub frame.

---

## 5. Out of scope

- The sidebar destination that lands here — **S5-D1**.
- The Source-page entry control — **S5-D3**.
- The per-Source node destination behind a row — **S5-D4**.
- The canvas and anything spatial — Spike 6.
- Citations, Observations, properties, artifact viewing, transcription.
- Creating, editing, or deleting Sources from this destination.
- Node counts or any derived per-Source statistic (see SL-8).

---

## 6. Acceptance checklist

- [ ] The list reuses `PVList` and the Sources row anatomy — no new component, no `PVTable`.
- [ ] It is visually obvious *why* this destination exists and where a row goes, despite rows resembling the Sources list.
- [ ] Empty state points at the Sources destination rather than offering to create a Source here.
- [ ] No node counts, no progress indicators, no dashboard chrome.
- [ ] Nothing on any frame implies that Nodes are cited or that evidence has been recorded.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| List data | Reads the **existing** `.sourcesList(project:)` query key — sharing a key is encouraged, so SL-1 / SL-2 cost nothing at the data layer. SL-8 is what would break that. |
| Section root is load-bearing | This list is also the error-recovery destination: when a Source is deleted while its node destination is in history, navigation falls back here. It cannot be skipped. |
| No-Artifact signal (SL-5) | Requires knowing artifact presence per row. Confirm the existing list payload carries it before committing to the affordance. |
