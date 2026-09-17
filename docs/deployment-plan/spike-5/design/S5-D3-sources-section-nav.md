# S5-D3 — Sources section nav (nested config)

**Kind:** Claude Design board
**Spike:** Provenencia Spike 5 (Interpretation foundation)
**Implements later as:** PR S5-07 (sidebar + stub destinations)
**Depends on:** Shipped workspace chrome (S2-01), Source types / Source fields nav (S2-01)
**Related briefs:** [`S5-D2`](S5-D2-sources-list-graph-entry.md) (Sources list → Evidence graph); archived [`S5-D1`](archive/S5-D1-interpretation-nav-entry.md) (**superseded** — do not implement)

Paste this entire document into Claude Design as the requirements for one board/flow. Read the shared product facts in [`README.md`](README.md) first.

---

## 1. Objective

Design **a nested Sources-family sidebar**: one primary work destination, four de-emphasized config children.

```text
Sources                         ← primary (same visual weight as today)
  ├─ Source types               ← config (smaller / indented / nested)
  ├─ Source fields
  ├─ Subject types              ← new config
  └─ Subject fields             ← new config
```

**Sources** is where research work starts (list → Source page or Evidence graph). The four children are vocabulary / admin for that work — not peer places. Nesting (or equivalent de-emphasis) should make that hierarchy obvious.

This board is **nav chrome only**: hierarchy, labels, icons, selected/idle, badges, stub destinations. It does **not** design Subject types / Subject fields editors.

**Supersedes S5-D1.** That board designed a flat top-level Interpretation item. Do not draw it.

---

## 2. Domain facts the UI must reflect

| Fact | UI implication |
| --- | --- |
| One product layer for Source + Evidence graph | No Interpretation / Graph top-level item. |
| Hierarchy | Sources = work; four others = config under Sources. |
| Subject ≠ Source | Subject types / fields = what a document can *talk about*; Source types / fields = how it is *filed*. |
| Avoid "claim" | Conclusion owns that word. Prefer **Subject**. |
| Subject editors not built yet | Stub / "coming soon" destinations are fine. |
| Engine names stay internal | No "node," `subject_types`, or "observation" in the rail. |

### 2.1 What this board is not

- Not the Sources list dual-action design — **S5-D2**.
- Not the Subject types / Subject fields admin UI.
- Not the Evidence graph / canvas.

---

## 3. Requirements

| ID | Requirement |
| --- | --- |
| NAV-1 | **Sources** remains a primary sidebar destination at the same prominence as today (peer to whatever other primary sections exist). |
| NAV-2 | **Source types**, **Source fields**, **Subject types**, and **Subject fields** appear as **nested / de-emphasized** children of Sources — not five equal-weight flat rows. Propose the pattern (indent, smaller type, disclosure group, section label, etc.). The shipped rail is currently flat; inventing hierarchy here is intentional. |
| NAV-3 | Propose **final labels** (working defaults: Subject types / Subject fields). Mirror the Source types / Source fields grammar. |
| NAV-4 | Propose **icons** from `PVSymbol` where possible. Config children may share a quieter treatment than Sources. |
| NAV-5 | Propose **order** among the four children (Source types → Source fields → Subject types → Subject fields is the default — justify if different). |
| NAV-6 | **No** Interpretation / Graph / Map / Claims top-level destination. |
| NAV-7 | **No count badges** on the four config children in this spike (or empty badge slots that do not look broken). Sources may keep its existing count. |
| NAV-8 | Selected / unselected: selecting a nested item must be obvious; selecting Sources vs a child must not look identical. |
| NAV-9 | Stub content for Subject types and Subject fields: one short "coming soon" line. Source types / Source fields keep their shipped screens. |
| NAV-10 | Collapsed vs expanded parent (if using disclosure): propose default (expanded is fine for a solo dogfood). Do not hide config behind a menu that is hard to discover. |

---

## 4. Screen / frame inventory (minimum)

1. Sidebar: Sources selected, nested config visible (expanded).
2. Same: Subject types selected (stub content).
3. Same: Subject fields selected (stub content).
4. If disclosure exists: collapsed parent state.
5. Annotation of hierarchy / size / weight vs a flat five-item alternative (why nested wins).

---

## 5. Out of scope

- Sources list actions and Evidence graph stub — **S5-D2**.
- Real Subject types / Subject fields editors.
- The canvas.
- Other future layer groups (Conclusion, Narration) — show Sources family only, or light context if needed.

---

## 6. Acceptance checklist

- [ ] Sources reads as primary work; the four config items read as subordinate.
- [ ] Subject types / Subject fields parallel Source types / Source fields.
- [ ] No Interpretation top-level item.
- [ ] Labels avoid "claim," "node," and engine jargon.
- [ ] Stub destinations are cheap.
- [ ] Nesting pattern is specified enough to implement (not "make it nested somehow").

---

## 7. Implementation notes (for PRs — not Design homework)

| Item | Status |
| --- | --- |
| Today | `WorkspaceSection` is flat `CaseIterable` → one sidebar row each. Nesting needs a sidebar presentation model (group + children), not only new cases. |
| Sections | Still add `subjectTypes` / `subjectFields`; parent grouping is view-layer. |
| Place registry | Stub presentations for the two new sections; exhaustive tests update. |
| S5-D1 | Superseded; do not implement. |
