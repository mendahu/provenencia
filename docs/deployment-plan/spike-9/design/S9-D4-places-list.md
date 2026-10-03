# S9-D4 — Places list

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Places list page  
**Implements later as:** PR **S9-26**  
**Depends on:** S9-D2 (row anatomy); S9-25 (place reads), S9-08 (destination)  
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
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Every Property is **multi-valued** across members.
- Each field shows **one resolved value** in a state: **single**, **merged** (names and dates auto-reconciled), **mixed** (members disagree; top-ranked value shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

A list of every Place. Row: **thumbnail slot · toponym · ref**. Default sort by toponym. The sparsest list; it should still read as the same family as Persons and Events.

```text
[ ⌖ ]  York                                           PLC-5YK01
[ ⌖ ]  Upper Canada  +2                               PLC-9UC11
[ ⌖ ]  PLC-3ZZ70                                      PLC-3ZZ70
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Toponyms differ by record | *Upper Canada*, *U.C.*, *Canada West* can all be one Place. Show the top-ranked one; *+N* for the rest. |
| Toponym is text, not reconciled | No merge; ranked distinct values. The top-ranked toponym shows with *+N*; no mixed marker in rows (S9-D2 decision). |
| Fallback | toponym → label → ref. |
| No geography | No map, coordinates, or containment this spike. |

### 2.1 What this board is not

- Not a map.
- Not containment (York ⊂ Upper Canada).

---

## 3. Implementation gate (S9-26)

| Ships in **S9-26** | Does **not** ship there |
| --- | --- |
| Places list per S9-D2 row anatomy | Map, coordinates, hierarchy |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PLL-1 | Reuse the S9-D2 row with title + ref only; decide whether the secondary line is empty or omitted. |
| PLL-2 | *+N* as in S9-D2. No *mixed* marker in rows (S9-D2 decision); disagreement shows on Place detail. |
| PLL-3 | Empty state points at Promote. |
| PLL-4 | Place placeholder thumbnail. |

---

## 5. Suggested frames

1. Typical list.
2. Place with several toponyms.
3. Ref-only row.
4. Empty state.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Places list | Snowflake | **New** | `Features/Conclusions/PlacesListView.swift` | |
| Conclusion list row | Snowflake | Ship (from D2) | `Features/Conclusions/ConclusionListRow.swift` | |
| Thumbnail / EmptyState / markers | Component | Ship | kit | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A tree / outline of places | No containment model yet. |

---

## 7. Out of scope

- Maps
- Gazetteer binding

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-26** against the board and inventory (kit first).
