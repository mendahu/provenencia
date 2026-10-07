# S9-D15 — Custom term: category

> **Retired 2026-10-07.** Never built. Place relationships are locked `part_of` / `succeeded_by` (UI-driven); `place_nature` uses the ordinary custom-term dialog with no category. Kept for history only.

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Citation Composer — the custom term dialog  
**Implements later as:** ~~PR **S9-38b**~~ (retired)  
**Depends on:** S9-38 (term categories in the engine)  
**Related:** S9-D7 (Place detail shows parents grouped by the parent's nature)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** to the composer's existing custom term dialog. Add to it; do not redesign the composer.

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

When a researcher adds their own **place relationship type** (beyond *part of* and *succeeded by*), they say how it behaves: **hierarchical** (one place is part of another) or **temporal** (one place became another). Place **nature** (administrative / informal / ecclesiastical) lives on the Place itself and is not chosen here.

```text
New relationship type
  Label      [ Judicial district                ]
  Behaves as ( ) Part of — one place within another
             ( ) Succession — one place became another
                                   [Cancel] [Add type]
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Only some Properties use categories | Today only `place_relationship_type`. Terms for other Properties (roles, relationship types, event types, **place nature**) never show the choice. |
| Two categories | **Hierarchical** (*part of*): builds chains like *Toronto, Ontario, Canada*. **Temporal** (*succession*): links a lineage (*York → Toronto*), never a chain. |
| It is required | A place relationship type without a category can't be used; the dialog can't add one without a choice. |
| Seeded types already have one | *Part of* is hierarchical; *succeeded by* is temporal. |
| Where it's met | The researcher is citing a place relationship in the composer (drawing it between two place cards) and picks *Add type…* in its type picker. |

### 2.1 What this board is not

- Not a term manager.
- Not the composer's relationship row itself (unchanged).
- Not choosing a Place's nature (that is a Property on the Place).

---

## 3. Implementation gate (S9-38b)

| Ships in **S9-38b** | Does **not** ship there |
| --- | --- |
| The category choice in the custom term dialog, shown only for Properties whose terms carry a category | Editing a term's category after it exists |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| TC-1 | The dialog shows the category choice only when the Property's terms carry one. |
| TC-2 | Each choice has a plain-language label and an example. |
| TC-3 | Add is disabled until a category is chosen. |
| TC-4 | Errors (a duplicate label) show as today. |

---

## 5. Suggested frames

1. Custom term dialog for a role (unchanged, no category).
2. Custom term dialog for a place relationship type, nothing chosen.
3. Same, *Part of* chosen.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Custom term dialog | Snowflake | Extend | `Features/CitationComposer/CitationComposerView.swift` | Existing FormDialog. |
| Choice | Component | Ship | Radio from the kit | |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A category on every term | Only Properties that use one. |

---

## 7. Out of scope

- Term editing and deletion
- Cardinality (S9-D14)

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-38b** against the board and inventory (kit first).
