# S9-D5 — Person detail

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Person detail page  
**Implements later as:** PR **S9-16** (later on this view: S9-32 fills life dates and places)  
**Depends on:** S9-15 (detail read + value-state formatting, reasoning), S9-13 / S9-13b (reconciler, name module), S9-14 (evidence + reasoning in the cache)  
**Revision:** 2026-10-05, for the reconciliation design: values explain themselves.  
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
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Each Property collects values from every member, and the engine **reconciles** them ([`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md)).
- Each field shows its reconciled value in a state: **single**, **merged** (several records agree once reconciled), **mixed** (records disagree and the evidence couldn't narrow them; every surviving value is shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- **Every value can explain itself.** The engine returns every record it considered with an outcome: kept, folded into a fuller value (*J.* into *James*), outvoted by a majority of Sources, dropped as weak evidence (low-trust Source, uncertain transcription, low-confidence claim), denied by a stronger negative record, or no usable value. Support counts **Sources**, not records.
- A few Properties hold **several true values** (a Place's concurrent names, *Montréal* and *Montreal*). Those show every value; most fields show one.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

The page for one Person. Read-only this spike. Header fields: **thumbnail slot, name, birth date, death date, birth place, death place, ref** — a superset of the list row.

Every field shows its **state**, its value(s), and **why**: the records the engine considered and what happened to each.

```text
[ ◯ ]  James Robins                          PER-7KD45
       merged · 3 Sources                    ▸ Why
       Born   14 May 1817 · York, Upper Canada  merged · 2 Sources
       Died   ABT 1880    · Toronto, Ontario    mixed  ▸ 2 values · Why

  Why "James Robins"
    James Robins     1850 census          kept
    J. Robins        Baptism register     folded into James Robins
    James Robbins    Family Bible         outvoted (2 of 3 Sources)
    James Robins     Newspaper notice     weak · low-trust Source
```

This board defines the **value-state and reasoning vocabulary** that Event and Place detail reuse.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Four states now, five later | single, merged, mixed, empty; **concluded** (a researcher's Reconciliation Claim) arrives later and must fit without a relayout. |
| Mixed shows every survivor | When the evidence can't narrow a value, every surviving value is shown, not a "winner". The first one leads; the rest disclose. |
| Values explain themselves | Each value has a **Why**: every record considered, its Source, the value it read, and its outcome. Outcomes: *kept*, *folded into …*, *outvoted (n of m Sources)*, *weak* (low-trust Source / uncertain transcription / low-confidence claim), *denied by …* (a stronger record saying "not this"), *no usable value*. |
| Support counts Sources | "2 Sources", not "2 records": two Observations from one Source are one vote. |
| Against | A negative record ("not James Robins") that didn't eliminate anything still counts against a value: show *1 record disagrees*. |
| Derived fields | Birth / death date and place come from linked Events; places read with their parent chain at the event's date (*York, Upper Canada*). If the birth Event isn't promoted, the field is empty. |
| Read-only | No edit, merge, delete, or reconcile controls this spike. The *Why* is where a future "conclude this value" action will sit. |
| Members exist | A member list (Subjects + Sources) is a **stretch** section; design a slot for it if cheap. |

### 2.1 What this board is not

- Not reconciliation (choosing a value) — later; leave room for it in *Why*.
- Not editing identity claims (Spike 10).
- Not a family tree or timeline.
- Not Observation editing: *Why* rows may link to the Citation, nothing more.

---

## 3. Implementation gate (S9-16)

| Ships in **S9-16** | Does **not** ship there |
| --- | --- |
| Header, value states, values, *Why* disclosure, empty states; name populated, life-date and place rows empty | Filling life dates and places (S9-32 — design them fully here); edit / reconcile / merge controls |
| Optional member list if the board keeps it | Tree, timeline, map |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PD-1 | Header shows thumbnail slot, reconciled name (fallback label → ref), and the ref. |
| PD-2 | Birth and death rows each show date and place with their state; places read with their chain at the event's date, and a date that straddles a change shows every candidate (*Upper Canada or Province of Canada*). |
| PD-3 | State vocabulary is visual and textual (VoiceOver reads *mixed*, *merged from 3 Sources*). |
| PD-4 | Mixed fields show every surviving value; the first leads, the rest disclose. |
| PD-5 | Every field has a **Why** disclosure listing each record considered: Source, value as read, outcome. Outcomes have one short phrase each (§2) and are distinguishable without colour. |
| PD-6 | Support reads in Sources; a value with negatives against it says so. |
| PD-7 | Empty fields are honest (*No birth recorded*), not hidden, so the page shape is stable. |
| PD-8 | Layout leaves room for a *concluded* state and a future "conclude this value" action in *Why* (Reconciliation Claims) without redesign. |
| PD-9 | Back returns to wherever the researcher came from (list, graph, omnibar). |

---

## 5. Suggested frames

1. Complete Person, all single values.
2. Merged name: *J. Robins* folded into *James Robins*, *Why* open.
3. Mixed death date: two surviving values, *Why* open.
4. Weak and denied: a low-trust spelling dropped; a stronger "not James Robins" record denying one value; *1 record disagrees* on another.
5. Birth place with an undecided chain (*Toronto, Upper Canada or Province of Canada*).
6. Sparse Person: name only, no events.
7. Ref-only Person (no name, no label).
8. Optional: member list section.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Person detail | Snowflake | **New** | `Features/Conclusions/PersonDetailView.swift` | Reads the detail key. |
| Value-state field row | Snowflake | **New** | `Features/Conclusions/ResolvedValueRow.swift` | Shared by D5–D7. |
| Reasoning list (*Why*) | Snowflake | **New** | `Features/Conclusions/ReconciliationReasoningView.swift` | One row per record: Source, value, outcome. Shared by D5–D7. |
| Outcome phrase | Snowflake | **New** | `Features/Conclusions/ReconciliationOutcome.swift` | Maps the engine's reason keys to L10n phrases and marks. |
| Field | Component | Ship | `DesignSystem/Components/Field/` | Label + value. |
| State / outcome mark | Component | Ship | `Badge` / `Chip` / `Recipes/Marks` | Mixed / merged; outcome marks. |
| Disclosure | Component | Ship | kit | *Why*, other values. |
| Section header | Component | Ship | `DesignSystem/Components/SectionHeader/` | Life events, members. |
| Thumbnail | Component | Ship | kit | Placeholder. |
| Date / name display | Snowflake | Ship / Extend | `Features/Dates/`, `Features/Names/` | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit resolved-value or reasoning component | Three call sites in one feature. |
| Numeric confidence scores | Reconciliation is elimination, not a score; show outcomes and Source counts. |

---

## 7. Out of scope

- Reconciliation UI (concluding a value)
- Claim management
- Relationships section (parents, spouses) — later

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-16** against the board and inventory (kit first).
