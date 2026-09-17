# S5-D2 — Sources list → Evidence graph

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-08
**Depends on:** S5-D3 (Sources-family nav), shipped Sources list (S2-04)
**Related brief:** [`S5-D3`](S5-D3-sources-section-nav.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **dual actions on the existing Sources list** so a researcher can open a Source's filing page *or* its **Evidence graph** — without a second Source picker or an Interpretation sidebar section.

```text
Sources list (shipped) ──▶ Source page        (filing — already works)
                       ──▶ Evidence graph     (stub now; canvas in Spike 6)
```

The list stays the Sources destination. The Evidence graph is a deep place under the same sidebar family. Early development: the graph destination may be a **"coming soon" stub**.

Also handle **Sources with no Artifact**: still listed; Evidence graph action disabled; shortcut to the Source page to add an Artifact.

**Naming:** use **Evidence graph** in chrome and actions (not "Interpretation graph"). The researcher is already Source-scoped from this list.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| One product layer | This is still the Sources list — not a parallel picker under another section. |
| Two jobs per Source | Filing (Source page) vs Evidence graph. Actions must be **named differently** — not twin "Edit" labels. |
| Graph needs an Artifact | Citations require an Artifact. No Artifact → cannot open the Evidence graph. |
| Graph is the only Node surface | Do not imply a table of nodes behind the graph action. |
| Stub is temporary | Evidence graph may be "coming soon"; keep it cheap. |

### 2.1 What this destination is not

- Not a second Sources list under another sidebar item.
- Not a node / citation / observation list.
- Not the Subject types / Subject fields screens — **S5-D3**.

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| SL-1 | Reuse the shipped **`PVList`** Sources list (S2-04). Do **not** invent a parallel list destination. |
| SL-2 | Keep existing row content: **title**, **`SRC-…`**, **source type**, thumbnail slot. |
| SL-3 | Each row offers **two clear actions**: open the **Source** (detail page) and open the **Evidence graph**. Propose labels, icons, and hierarchy (primary vs secondary). Avoid "Edit" for both. Prefer copy that says Evidence graph (or a short verb that clearly leads there). |
| SL-4 | Evidence graph on a Source that **has at least one Artifact** navigates to that Source's graph place. In this spike the destination may be a stub. |
| SL-5 | Sources with **no Artifact** stay listed. Evidence graph action is **disabled** (no click-through). |
| SL-5a | On a no-Artifact row, keep a clear path to the **Source page** so the researcher can add an Artifact. Propose how that reads with SL-3. |
| SL-5b | Make the blocked state legible (needs an Artifact first) without a hard error banner. |
| SL-6 | Empty state stays the shipped Sources empty state — unchanged. |
| SL-7 | Filter/sort may stay as shipped; do not redesign the toolbar unless dual actions force a small change. |
| SL-8 | **No graph-specific columns** — no node counts, "3 uncited," last-worked. |
| SL-9 | Evidence graph stub: one short "coming soon" line that uses the **Evidence graph** name. |

---

## 4. Screen / frame inventory (minimum)

1. Sources list with dual actions visible on openable rows.
2. Close-up of a no-Artifact row (Evidence graph disabled + path to Source page).
3. Evidence graph stub destination after the graph action.
4. Optional: comparison annotation so the list still feels like Sources, not a second product area.

---

## 5. Out of scope

- Nested sidebar config — **S5-D3**.
- Source-page Evidence graph button — Spike 6 is fine.
- The canvas itself.
- Subject vocabulary editors.

---

## 6. Acceptance checklist

- [ ] One Sources list — not a second picker.
- [ ] Source vs Evidence graph are visually and verbally distinct.
- [ ] Product copy uses **Evidence graph**, not Interpretation graph.
- [ ] No-Artifact rows cannot open the graph; Source page remains reachable.
- [ ] Stub is cheap.
- [ ] No node counts or dashboard chrome.

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| Navigation | `WorkspaceLocation` needs a **page vs graph** discriminator under `.sources` + `sourceId` (design note §11.1.2). |
| List data | Existing `.sourcesList(project:)`. Confirm artifact presence for SL-5. |
| Stub | Deep place + placeholder view titled Evidence graph; Spike 6 swaps the view. |
| L10n | New strings for the Evidence graph action and stub — use the skill. |
| Environment | Sources list must hold `@Environment(WorkspaceNavigation.self)` for the graph action. |
