# S9-D6 — Event detail

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Event detail page  
**Implements later as:** PR **S9-24** (later on this view: S9-32 fills subject titles and places)  
**Depends on:** S9-D5 (page + value states); S9-22 (event reads)  
**Related:** S9-D3 (row it expands)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief **extends S9-D5**'s frames. Reuse its layout; change only what this kind needs.

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

The page for one Event: **thumbnail slot, title, date, place(s), ref**, using S9-D5's value-state rows.

```text
[ ◇ ]  Marriage of James Robins and Mary Smith     EVT-8PL22
       Date    ABT 1810           ⚠ mixed   ▸ 1 other value
       Place   Kingston · Frontenac
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Title is composed | Same precedence as the list (S9-D3). A recorded `event_name` wins. |
| Several places | An Event may have several Locations (York *and* Upper Canada). Show all, not *+N*. |
| Date may be a span | `start_date`–`end_date` when there is no single date. |
| Participants | Subject Persons are what the title names; linking them to their Person pages is welcome if cheap. |

### 2.1 What this board is not

- Not a participants / roles editor.
- Not a map.

---

## 3. Implementation gate (S9-24)

| Ships in **S9-24** | Does **not** ship there |
| --- | --- |
| Event header + date using S9-D5 rows; place rows empty | Subject titles and places (S9-32 — design them here); participant management |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| ED-1 | Title as composed; the recorded-name case looks no different from composed titles except its source. |
| ED-2 | Date row with states; span form designed. |
| ED-3 | Places listed in full. |
| ED-4 | If subject Persons are shown, each links to its Person page. |
| ED-5 | Reuse S9-D5 rows and states; no new row style. |

---

## 5. Suggested frames

1. Birth with one subject.
2. Marriage with mixed date.
3. Census *et al.* with several places.
4. No-subject event (*Fire at York*).
5. Sparse event (ref only).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Event detail | Snowflake | **New** | `Features/Conclusions/EventDetailView.swift` | |
| Value-state field row | Snowflake | Ship (from D5) | `Features/Conclusions/ReconciledValueRow.swift` | |
| Event title formatter | Snowflake | Ship (S9-22) | `Features/Conclusions/` | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A second page layout | Reuse D5. |

---

## 7. Out of scope

- Editing
- Timelines

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-24** against the board and inventory (kit first).
