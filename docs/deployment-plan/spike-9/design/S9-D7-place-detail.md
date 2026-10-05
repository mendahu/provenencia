# S9-D7 — Place detail

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Place detail page  
**Implements later as:** PR **S9-27** (later on this view: S9-40 fills period, hierarchy and succession)  
**Depends on:** S9-D5 (page, value states, *Why*); S9-25 (place reads), S9-38 / S9-39 (place model, chains)  
**Revision:** 2026-10-05: several names per Place; period; hierarchy; succession.  
**Related:** S9-D4 (row it expands)  
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

The page for one Place: **thumbnail slot, names, ref, period, the places it is part of, the places part of it, and what it succeeded or was succeeded by.**

```text
[ ⌖ ]  Toronto                                   PLC-5YK01
       Ontario, Canada                            (today)
       Names    Toronto · Tkaronto               ▸ Why
       Period   1834 –

  Part of
    Administrative   Upper Canada           1834 – 1841
                     Province of Canada     1841 – 1867
                     Ontario                1867 –
    Geographic       Golden Horseshoe

  Contains
    Scarborough Township · Etobicoke · …

  Succession
    Succeeded   York  (1793 – 1834)
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Several names at once | All of a Place's reconciled names show (*Toronto*, *Tkaronto*); each has the D5 *Why*. Spellings and case already merged; a low-trust spelling may be dropped as *weak*. |
| Period | When the Place existed or mattered (*1834 –*, *until 1867*, unknown). Either end may be missing. |
| "Part of" has a type | Administrative, geographic, ecclesiastical, plus researcher-added types. Group parents by type. |
| Parents change with time | A link holds while both places' periods overlap, so a list of administrative parents reads as a timeline. The header shows **today's** chain (or the last one for a place that ended). |
| Contains | Places that are part of this one (direct children), grouped like parents. Can be long. |
| Succession | *Succeeded* / *succeeded by* links to the place before or after a rename or merger (York → Toronto). Not part of the hierarchy. |
| Everything is evidence | Every name, period and relationship came from a cited record (the hard way). Relationship rows can show their *Why* like any value. |
| Read-only | Relationships are drawn on the Evidence graph, not here. |
| No geography | No map or coordinates. |

### 2.1 What this board is not

- Not a map or gazetteer view.
- Not an editor for the hierarchy.

---

## 3. Implementation gate

| Ships in **S9-27** | Ships in **S9-40** | Does **not** ship |
| --- | --- | --- |
| Header, names with values, states and *Why* (D5 rows); period, hierarchy and succession sections present but empty | Period, Part of, Contains, Succession, the header chain | Maps, events-at-this-place, editing |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PLD-1 | Title is the first name (fallback label → ref); the other names show beneath with D5 value rows and *Why*. |
| PLD-2 | The header shows today's chain, or the last one for a place whose period has ended. |
| PLD-3 | **Period** row: start and end, either may be unknown. |
| PLD-4 | **Part of** lists parents grouped by relationship type, each with the span the link holds; an undated link reads as always. |
| PLD-5 | **Contains** lists direct children; long lists truncate with a count. |
| PLD-6 | **Succession** shows *Succeeded* and *Succeeded by* links with the other place's period. |
| PLD-7 | Every related place links to its own page. |
| PLD-8 | Reuse S9-D5 rows, states and *Why*. |

---

## 5. Suggested frames

1. A city with two names, a period, three administrative parents over time, a geographic parent, children, and a predecessor (Toronto).
2. A place that ended (Upper Canada: period closed, succeeded by Province of Canada, many children).
3. A researcher's place (a family farm: one name, one geographic parent, no period).
4. A top-level place with no parents (Canada).
5. Ref-only Place.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Place detail | Snowflake | **New** | `Features/Conclusions/PlaceDetailView.swift` | |
| Value-state field row | Snowflake | Ship (from D5) | `Features/Conclusions/ReconciledValueRow.swift` | |
| Reasoning list (*Why*) | Snowflake | Ship (from D5) | `Features/Conclusions/ReconciliationReasoningView.swift` | |
| Place chain text | Snowflake | Ship (from D4) | `Features/Conclusions/PlaceChainDisplay.swift` | |
| Place relationship row | Snowflake | **New** | `Features/Conclusions/PlaceRelationshipRow.swift` | Related place, type, span. |
| Section header | Component | Ship | kit | Part of, Contains, Succession. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Map placeholder | Not in scope; would imply geography exists. |
| Place kind badges | Not in this spike. |
| Inline relationship editing | Relationships are evidence, drawn on the Evidence graph. |

---

## 7. Out of scope

- Events at this place (later)
- Maps, coordinates, boundaries
- Gazetteer lookup ([`ideas/place-gazetteer-service.md`](../../../ideas/place-gazetteer-service.md))

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-27** (and the hierarchy sections in **S9-40**) against the board and inventory (kit first).
