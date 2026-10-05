# S9-D14 — Properties: cardinality

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Properties page (Configure → Properties)  
**Implements later as:** PR **S9-37**  
**Depends on:** S9-36 (cardinality in the engine)  
**Related:** S9-07b (the Properties page as shipped)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** to the shipped Properties page. Add to its frames; do not redesign the page.

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

Let a researcher say whether a Property holds **one value or several** on a Person, Event or Place. Today every Property is reconciled down to one value; a few hold several true values at once (a Place's concurrent names). Seeded Properties have a fixed answer; researcher-created ones choose.

```text
Properties › languages spoken            (inspector)
  Value type     Text
  Holds          ( ) One value   (•) Several values
                 Several: every distinct value is kept; spellings still merge.
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Two choices | **One value** (single; the default) or **several values** (multiple). |
| What it changes | Single: the engine narrows records down to one value, or shows *mixed* when it can't. Several: every distinct value is kept; duplicates, spellings and weak evidence still drop. |
| Seeded Properties are fixed | `toponym` holds several; every other seeded Property holds one. Show the value, read-only, like other seeded facts in the inspector. |
| Most repeating facts are events | Occupations, residences and religion over time belong in events, not a several-values Property. A one-line hint can say so. |
| Changing it recomputes | Every Person / Event / Place carrying the Property is reconciled again. Fast, but the page should not imply data is changed or lost. |
| Where it lives | The Property inspector, beside value type. Also in the create form. |

### 2.1 What this board is not

- Not a redesign of the Properties page or its bindings.
- Not term management.

---

## 3. Implementation gate (S9-37)

| Ships in **S9-37** | Does **not** ship there |
| --- | --- |
| Cardinality control in the inspector and the create form; read-only for seeded Properties | Term categories (S9-D15); any other Property setting |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PC-1 | The inspector shows **Holds: one value / several values** for every Property. |
| PC-2 | Researcher-created Properties can change it; seeded ones show it read-only. |
| PC-3 | A short explanation of each choice, and the hint that repeating facts over time are usually events. |
| PC-4 | The create form asks for it, defaulting to one value. |
| PC-5 | Changing it needs no confirm unless the board finds it confusing; nothing is deleted. |

---

## 5. Suggested frames

1. Inspector for a seeded single Property (`name`), read-only.
2. Inspector for `toponym`, read-only *several values*.
3. Inspector for a researcher Property, editable.
4. Create form with the choice.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Properties inspector | Snowflake | Extend | `Features/Properties/PropertiesView.swift` | Add the *Holds* row. |
| Create form | Snowflake | Extend | `Features/Properties/PropertiesView.swift` (`PropertiesCreateHost`) | |
| Choice | Component | Ship | Radio / Select from the kit | |
| Field / hint | Component | Ship | Field, caption text | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A new settings section or page | One row on the existing inspector. |

---

## 7. Out of scope

- Term categories (S9-D15)
- Reconciliation settings beyond cardinality

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-37** against the board and inventory (kit first).
