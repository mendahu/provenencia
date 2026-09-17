# S5-D2 — Interpretation Sources list

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-08
**Depends on:** S5-D1 (the sidebar destination that lands here), shipped Sources list (S2-04)
**Related brief:** [`S5-D1`](S5-D1-interpretation-nav-entry.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **the screen behind the Interpretation sidebar destination: a Source picker.**

Interpretation is Source-scoped, so before there is anything to look at, the researcher has to choose which document they are going to sit down and work through. That choice is this screen.

```text
[ S5-D1 ] ──▶ Interpretation Sources list ──▶ (nothing yet — the graph is Spike 6)
                     this board
```

**Read this before anything else, because it is unusual.** In this spike the rows **go nowhere.** The only way to work with Nodes is the spatial graph, and the graph is not built yet, so there is no destination behind a Source. The list ships now because the sidebar destination needs a root to land on, and because the list itself is permanent — it is the real, final Source picker for the layer. Only its rows are temporarily inert.

So this board has two jobs, and the second one is the interesting one:

1. Design the Source picker as it will permanently be.
2. Decide how a row reads **when it cannot be opened**, without designing something that has to be undone once it can.

**The trap to avoid:** a row that looks tappable and silently does nothing. To the only user of this build, that is indistinguishable from a bug, and "it's a stub" is not visible in the UI.

**The second trap:** over-designing the unavailable state. Whatever you draw here gets deleted the moment the graph lands, so it should be the smallest honest thing — not an illustrated empty state, not a "coming soon" marketing panel.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| Interpretation is Source-scoped | The section root is a **Source picker**, not a list of interpretation objects. There is nothing to show that is not scoped to a Source. |
| Same Sources, different intent | "Which document am I going to work through?" rather than "which document do I have?" The rows are identical; the intent is not. |
| Sources may have no Artifact | A Source with no attached file can never be cited — Citations require an Artifact. Not an error, but the researcher is about to sit down with something they cannot cite from. |
| This destination is read-only | Sources are created, renamed, retyped, and deleted in the Sources destination. None of that happens here. |
| **The graph is the only Node surface** | There is no list or table view of Nodes anywhere in the product, now or later. Do not design toward one, and do not imply a row leads to a table of things. |
| **The graph does not exist yet** | Rows cannot be opened in this spike. This is the one temporary thing on the board. |

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
| SL-4 | Design the row's **eventual** behaviour as opening that Source's interpretation graph — never the Source page. This is the permanent target and the copy in SL-3 should describe it. |
| SL-5 | Optional, if it can be done without clutter: a quiet signal on rows for Sources with **no Artifact** (nothing to cite from yet). Not an error state, not a warning colour — an informational affordance at most. Propose or explicitly decline. |
| SL-6 | Empty state: no Sources in the project at all. The CTA should send the researcher to the Sources destination to add one — this list cannot create Sources. Reuse `PVEmptyState`. |
| SL-7 | Search/filter is **optional**. Include it only if it is a straight reuse of the Sources list's `PVInput` pattern; do not design a new filtering model. |

### 3.1 The inert state (this spike only)

| ID | Requirement |
| --- | --- |
| SL-9 | Rows cannot be opened in this spike. **Decide and justify one approach**: (a) rows are plainly non-interactive — no hover, no pressed state, no chevron, so nothing invites a click; or (b) rows keep their normal appearance but the destination is explicitly marked unavailable. Do not propose (c) a normal-looking row with a silent no-op. |
| SL-10 | Whatever you choose must be **cheap to remove.** It is deleted the moment the graph lands. A styling variant is fine; a separate row component, a new empty-state illustration, or bespoke copy per row is not. |
| SL-11 | If the purpose copy from SL-3 needs a temporary sentence acknowledging the graph is not built yet, propose it — and propose the permanent version alongside, so the diff that removes it is obvious. |
| SL-12 | Do **not** design a placeholder destination. There is no screen behind a row in this spike; the row simply does not activate. |

### 3.2 Explicitly deferred

| ID | Requirement |
| --- | --- |
| SL-8 | **No graph-specific columns on the rows** — no node counts, no "3 uncited," no last-worked timestamp. These are wanted eventually and are deliberately out of scope: they are derived counts, so placing a single Node would stale this list and force a cache invalidation between the graph and its own landing page. Design the row without them and do not leave a visible gap where they would go. |

---

## 4. Screen / frame inventory (minimum)

1. Interpretation Sources list, populated, in its **permanent** form — mixed types, a couple with thumbnails, at least one without.
2. The same list in its **inert** form per SL-9, so the difference between the two is a visible, reviewable diff.
3. Same list showing the no-Artifact affordance, if SL-5 is accepted.
4. Empty state (no Sources in the project at all).

No frame for the destination behind a row — there isn't one.

---

## 5. Out of scope

- The sidebar destination that lands here — **S5-D1**.
- The interpretation graph itself, and anything spatial — Spike 6.
- A placeholder or "coming soon" destination behind a row — explicitly not wanted (SL-12).
- A Source-page control that jumps into the layer — deferred to Spike 6 with the graph.
- Any list or table view of Nodes — the graph is the only Node surface, permanently.
- Citations, Observations, properties, artifact viewing, transcription.
- Creating, editing, or deleting Sources from this destination.
- Node counts or any derived per-Source statistic (see SL-8).

---

## 6. Acceptance checklist

- [ ] The list reuses `PVList` and the Sources row anatomy — no new component, no `PVTable`.
- [ ] It is visually obvious *why* this destination exists, despite rows resembling the Sources list.
- [ ] The permanent form and the inert form are both shown, and the difference between them is small and obviously removable.
- [ ] No row looks actionable while doing nothing.
- [ ] No placeholder destination was designed.
- [ ] Empty state points at the Sources destination rather than offering to create a Source here.
- [ ] No node counts, no progress indicators, no dashboard chrome.
- [ ] Nothing on any frame implies that Nodes exist, are cited, or can be created yet.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| List data | Reads the **existing** `.sourcesList(project:)` query key — sharing a key is encouraged, so SL-1 / SL-2 cost nothing at the data layer. SL-8 is what would break that. |
| Section root is load-bearing | This is why the list ships before it can do anything. `WorkspaceSidebar` builds items from `WorkspaceSection.allCases`, so adding the section automatically adds a sidebar row pointing at `.sectionRoot(section)` — something has to render there. It is also the error-recovery destination once deep places exist. |
| Inert rows are not a no-op handler | Whatever SL-9 chooses has to be implemented as that, not as a tap target wired to nothing. See the deployment plan's gotcha list. |
| No-Artifact signal (SL-5) | Requires knowing artifact presence per row. Confirm the existing list payload carries it before committing to the affordance. |
