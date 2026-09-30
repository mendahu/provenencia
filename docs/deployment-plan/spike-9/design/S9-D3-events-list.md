# S9-D3 — Events list

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Events list page  
**Implements later as:** PR **S9-20**  
**Depends on:** S9-D2 (row anatomy); S9-11, S9-12, S9-18  
**Related:** S9-D6 (Event detail); naming matrix in [`deployment-plan.md`](../deployment-plan.md) R4  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief **extends S9-D2**'s frames. Reuse its layout; change only what this kind needs.

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

A list of every Event. Row: **thumbnail slot · event title · event date · event place · ref**. Default sort by date.

```text
[ ◇ ]  Birth of James Robins                          EVT-4MA10
       14 May 1817   ·   York
[ ◇ ]  Marriage of James Robins and Mary Smith        EVT-8PL22
       ABT 1810      ·   Kingston  +1
[ ◇ ]  Census of James Robins et al.                  EVT-2CC05
       1851          ·   Toronto
[ ◇ ]  Fire at York                                   EVT-7QQ31
       1849
[ ◇ ]  Birth of unnamed person                        EVT-1ZX88
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Events are rarely named | Titles are **composed**: recorded `event_name` → *{Type} of {subject}* → working label → *{Type} at {place}* / *Unspecified {type}* → ref. Show every case. |
| Several subjects | *Marriage of A and B*; otherwise *{Type} of {first} et al.* |
| Unnamed subject | *Birth of unnamed person*. |
| Date | Resolved `date`, else start–end span. Mixed ⇒ marker. |
| Place | From the Event's Locations; several ⇒ first + *+N*. |
| Sort | By date (undated last). |

### 2.1 What this board is not

- Not a timeline.
- Not the detail page (S9-D6).
- Not editing titles.

---

## 3. Implementation gate (S9-20)

| Ships in **S9-20** | Does **not** ship there |
| --- | --- |
| Events list per the S9-D2 row anatomy | Timeline / grouping by year |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| EL-1 | Reuse the S9-D2 row; title slot holds the composed title. |
| EL-2 | Show each title case from the naming precedence so long titles and *et al.* are designed, not discovered. |
| EL-3 | Date cell supports points, ranges, qualifiers, and spans. |
| EL-4 | *Mixed* and *+N* markers behave as in S9-D2. |
| EL-5 | Empty state points at Promote. |
| EL-6 | Event placeholder thumbnail distinct from Person / Place. |

---

## 5. Suggested frames

1. Typical list across title cases.
2. Mixed date; multi-place row.
3. Undated events sorting last.
4. Empty state.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Events list | Snowflake | **New** | `Features/Conclusions/EventsListView.swift` | |
| Conclusion list row | Snowflake | Ship (from D2) | `Features/Conclusions/ConclusionListRow.swift` | Same row. |
| Event title formatter | Snowflake | Ship (S9-12) | `Features/Conclusions/` | Title parts → text via L10n. |
| Thumbnail / EmptyState / markers | Component | Ship | kit | As D2. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A second row design | Reuse D2. |

---

## 7. Out of scope

- Timeline / calendar views
- Event type filters

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-20** against the board and inventory (kit first).
