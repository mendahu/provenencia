# S9-D4 — Places list

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Places list page  
**Implements later as:** PR **S9-26** (later on this view: S9-40 fills the chain cell)  
**Depends on:** S9-D2 (row anatomy); S9-25 (place reads), S9-39 (chains), S9-08 (destination)  
**Revision:** 2026-10-05: several names per Place; place hierarchy.  
**Related:** S9-D7 (Place detail)  
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
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Each Property collects values from every member, and the engine **reconciles** them ([`conclusion-reconciliation.md`](../../../conclusion-reconciliation.md)).
- Each field shows its reconciled value in a state: **single**, **merged** (several records agree once reconciled), **mixed** (records disagree and the evidence couldn't narrow them; every surviving value is shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- **Every value can explain itself.** The engine returns every record it considered with an outcome: kept, folded into a fuller value (*J.* into *James*), outvoted by a majority of Sources, dropped as weak evidence (low-trust Source, uncertain transcription, low-confidence claim), denied by a stronger negative record, or no usable value. Support counts **Sources**, not records.
- A few Properties hold **several true values** (a Place's concurrent names, *Montréal* and *Montreal*). Those show every value; most fields show one.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

A list of every Place. Row: **thumbnail slot · name (+N) · parent chain · ref**. Default sort by name. It should read as the same family as Persons and Events.

```text
[ ⌖ ]  Toronto                 Ontario, Canada                 PLC-5YK01
[ ⌖ ]  Montréal  +1            Québec, Canada                  PLC-2MT40
[ ⌖ ]  York                    Upper Canada                    PLC-7YK18
[ ⌖ ]  Robins family farm      Scarborough Township            PLC-9RF02
[ ⌖ ]  PLC-3ZZ70                                               PLC-3ZZ70
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| A Place can have several names at once | *Montréal* and *Montreal* are both true names of one Place. The row shows the first; *+N* for the rest. Spellings and case already merged. |
| A rename is a different Place | *York* and *Toronto* are two rows; one succeeded the other (shown on detail, not here). |
| Places are any grain | Township, county, province, country, a region, a family farm. No fixed levels. |
| Parent chain depends on the date | The row shows **today's** chain (or the latest one, for a place that no longer exists: *York → Upper Canada*). A place with no parents shows no chain. |
| Several parents | A place can be part of several places (administrative, geographic, ecclesiastical). The row follows the administrative chain. |
| Fallback | name → label → ref. |
| No geography | No map or coordinates. |

### 2.1 What this board is not

- Not a map.
- Not a tree or outline view of the hierarchy (a later view; the list stays flat).

---

## 3. Implementation gate

| Ships in **S9-26** | Ships in **S9-40** | Does **not** ship |
| --- | --- | --- |
| Places list per S9-D2 row anatomy: name, *+N*, ref; chain cell present but empty | The parent chain cell | Map, coordinates, tree view |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PLL-1 | Reuse the S9-D2 row: name as title, chain in the secondary slot, ref trailing. |
| PLL-2 | *+N* for extra names, as in S9-D2. No *mixed* marker in rows (S9-D2 decision); disagreement shows on Place detail. |
| PLL-3 | The chain reads as one line of names (*Ontario, Canada*), truncating from the top of the hierarchy. |
| PLL-4 | Empty state points at Promote. |
| PLL-5 | Place placeholder thumbnail. |

---

## 5. Suggested frames

1. Typical list, with chains.
2. A Place with two concurrent names (*+1*).
3. A Place that no longer exists, showing its last chain.
4. A top-level Place (no chain) and a ref-only row.
5. Empty state.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Places list | Snowflake | **New** | `Features/Conclusions/PlacesListView.swift` | |
| Conclusion list row | Snowflake | Ship (from D2) | `Features/Conclusions/ConclusionListRow.swift` | Secondary slot holds the chain. |
| Place chain text | Snowflake | **New** | `Features/Conclusions/PlaceChainDisplay.swift` | Formats chain parts; shared with D5 / D7 and Person rows. |
| Thumbnail / EmptyState / markers | Component | Ship | kit | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A tree / outline of places | Later view; the list stays flat. |
| Place kind badges | Not in this spike. |

---

## 7. Out of scope

- Maps
- Gazetteer lookup ([`ideas/place-gazetteer-service.md`](../../../ideas/place-gazetteer-service.md))

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-26** (and the chain cell in **S9-40**) against the board and inventory (kit first).
