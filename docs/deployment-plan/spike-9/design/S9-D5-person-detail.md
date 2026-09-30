# S9-D5 — Person detail

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Person detail page  
**Implements later as:** PR **S9-22**  
**Depends on:** S9-11 (detail read), S9-12 (formatters)  
**Related:** S9-D6 / S9-D7 extend this page; S9-D2 (the row it expands)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); [`add-workspace-place`](../../../../.cursor/skills/add-workspace-place/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief designs a **new** surface. Establish its frames here; later Spike 9 briefs extend them.

### Claude Design — do this first (in order)

Work **in place** on this board. Do not fork a parallel copy of the surface.
- **Rethink** (this brief says replace): throw away the old frames. Do not keep a before/after to ship.
- **Enhancement**: add to the existing frames. Do not start a second composer / graph / page.

1. **Clear this board’s local design-system cache.** Claude Design keeps a stale pack; drawing against it invents local copies of kit controls.
2. **Delete this board’s reference** to the design-system bundle.
3. **Pull a fresh copy** of the Provenencia design system from the main project. Do not continue until the fetched kit lists current components. If the kit looks stale or empty, delete the cache and refetch. Do **not** draw a replacement kit locally.
4. **Compose from that kit.** Instance existing components. Reach for a **bespoke / local** control only when the use is truly this domain. One call site is not a new design-system primitive.

**Reach for (kit).** Instance these first. The **UI building-block inventory** later in this brief names the snowflakes and which kit piece each situation should use.

| Situation | Use |
| --- | --- |
| Labeled value, textarea, or trailing control | Field + TextArea / Input |
| Primary / secondary / ghost action | Button; icon-only → IconButton |
| Choose one from a short list | Select |
| Searchable pick | ComboBox |
| Warning, error, or inline hint | Callout |
| Page- or pane-level empty | EmptyState |
| Confirm replace or destroy | Confirm (`item:` snapshot, not a Bool) |
| Resource delete with inbound check | DeleteImpact recipe (S8-D9 / **S8-13**) — confirm if allowed, notice if blocked |
| Short create / edit form | FormDialog |
| Status / count / polarity mark | Badge; compact token → Chip |
| Cover or file thumb | Thumbnail |
| Grouping / raised or sunken row | Card |
| Section title | SectionHeader |
| Transient after-save notice | Toast |
| Native menu of actions | ContextMenu |

Do **not** invent a local Field, Button, Card, Select, Callout, or Confirm.

---

### Shared Spike 9 facts (all Conclusion boards)

- A canonical **Person / Event / Place** (`PER-…` / `EVT-…` / `PLC-…`) is a researcher's handle for one historical thing. In the UI it is a Person, never a "canonical entity." Interpretation Subjects on Evidence graphs keep candidate refs (`CPR-…`).
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Every Property is **multi-valued** across members.
- Each field shows **one resolved value** in a state: **single**, **merged** (names and dates auto-reconciled), **mixed** (members disagree; top-ranked value shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

The page for one Person. Read-only this spike. Header fields: **thumbnail slot, name, birth date, death date, birth place, death place, ref** — a superset of the list row.

Every field shows its **resolution state** and can reveal the **ranked alternatives** behind it:

```text
[ ◯ ]  James Robins                          PER-7KD45
       Born   14 May 1817 · York                merged (3 records)
       Died   ABT 1880    · Toronto   ⚠ mixed   ▸ 2 other values
```

This board defines the **value-state vocabulary** (single / merged / mixed / empty, and space for a future *concluded*) that Event and Place detail reuse.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Four states now, five later | single, merged, mixed, empty; **concluded** (researcher's Reconciliation Claim) arrives later and must fit without a relayout. |
| Alternatives exist | Each field has ranked clusters (distinct values). Mixed fields especially need a way to see them. |
| Ranking reflects provenance | Source credibility, transcription certainty, claim confidence, agreement. Show *support* (how many records back a value) — not a numeric score. |
| Derived fields | Birth / death date and place come from linked Events; if the birth Event is not promoted, the field is empty. |
| Read-only | No edit, merge, delete, or reconcile controls this spike. |
| Members exist | A member list (Subjects + Sources) is a **stretch** section; design a slot for it if cheap. |

### 2.1 What this board is not

- Not reconciliation (choosing a value) — later.
- Not editing identity claims (Spike 10).
- Not a family tree or timeline.
- Not Observation-level evidence drill-down beyond the cluster list.

---

## 3. Implementation gate (S9-22)

| Ships in **S9-22** | Does **not** ship there |
| --- | --- |
| Header, value states, cluster disclosure, empty states | Edit / reconcile / merge controls |
| Optional member list if the board keeps it | Tree, timeline, map |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PD-1 | Header shows thumbnail slot, resolved name (fallback label → ref), and the ref. |
| PD-2 | Birth and death rows each show date and place with their state. |
| PD-3 | State vocabulary is visual and textual (VoiceOver reads *mixed*, *merged from 3 records*). |
| PD-4 | Mixed and multi-cluster fields disclose ranked alternatives with their support count. |
| PD-5 | Empty fields are honest (*No birth recorded*), not hidden, so the page shape is stable. |
| PD-6 | Layout leaves room for a *concluded* state and future actions (Spike 10) without redesign. |
| PD-7 | Back returns to wherever the researcher came from (list, graph, omnibar). |

---

## 5. Suggested frames

1. Complete Person, all single values.
2. Merged name and date (compatible records).
3. Mixed death date with alternatives disclosed.
4. Sparse Person: name only, no events.
5. Ref-only Person (no name, no label).
6. Optional: member list section.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person detail | Snowflake | **New** | `Features/Conclusions/PersonDetailView.swift` | Reads the detail key. |
| Value-state field row | Snowflake | **New** | `Features/Conclusions/ResolvedValueRow.swift` | Shared by D5–D7. |
| Field | Component | Ship | `DesignSystem/Components/Field/` | Label + value. |
| State mark | Component | Ship | `Badge` / `Chip` / `Recipes/Marks` | Mixed / merged. |
| Section header | Component | Ship | `DesignSystem/Components/SectionHeader/` | Life events, members. |
| Thumbnail | Component | Ship | kit | Placeholder. |
| Date / name display | Snowflake | Ship / Extend | `Features/Dates/`, `Features/Names/` | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit resolved-value component | Three call sites in one feature. |
| Numeric confidence scores | Ranking is display policy; show support, not scores. |

---

## 7. Out of scope

- Reconciliation UI
- Claim management
- Relationships section (parents, spouses) — later

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-22** against the board and inventory (kit first).
