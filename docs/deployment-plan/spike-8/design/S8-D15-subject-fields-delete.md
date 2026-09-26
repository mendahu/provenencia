# S8-D15 — Subject Fields delete

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-17** (cut over `properties.Delete` + DeleteImpact on this page)  
**Depends on:** Frozen policy + Impact report in [`../deployment-plan.md`](../deployment-plan.md) §S8-09.1–§S8-09.4; **S8-12** `GetDeleteImpact`; **S8-D9** / **S8-13** DeleteImpact recipe  
**Related:** Shared confirm / notice is [`S8-D9-impact.md`](S8-D9-impact.md) — **instance it, do not restyle.** Source Types is [`S8-D13-source-types-delete.md`](S8-D13-source-types-delete.md). Source Fields is [`S8-D14-source-fields-delete.md`](S8-D14-source-fields-delete.md). Do not draw those pages.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped Subject Fields page. Do not start a second properties app.

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

## 1. Objective

Subject Fields already deletes unused properties (`.pvConfirm` yes/no). Replace that with DeleteImpact so an in-use property **names the Observations that use it**.

**This board designs** trash + DeleteImpact on **this page only**. Unused **user** property with **no terms** → confirm. In-use → notice lists `OBS-…` (composer locations). Terms remaining → notice lists those terms (no Terms page this spike). Seeded / plugin → `origin_locked`. Trash stays offered when inbound is non-empty. Type-bindings are not blockers (`usedBy` becomes Observation count).

`subject_type_fields` CASCADE with the property — a type-binding is **not** a blocker. Do not invent a Property Terms page. Do not add Subject type delete (sidebar stub).

---

## 2. Domain facts (this board)

| Fact | UI implication |
| --- | --- |
| Blocked by `observations.property_id` | Notice lists those Observations. |
| Type bindings CASCADE | A property used only as a suggested field on a type can still go. |
| Property terms | No Terms page. Do not draw term-row trash unless it already exists on this page. |
| After erase | Clear selection as today. |

### 2.1 What this board is not

- Not Source Types (**S8-D13**) or Source Fields (**S8-D14**).
- Not restyling DeleteImpact (**S8-D9**).
- Not composer Observation-row delete (**S8-D11**).

---

## 3. Implementation gate (S8-17)

| Ships in **S8-17** | Does **not** ship there |
| --- | --- |
| `properties.Delete` through `deleteimpact`; drop `sqlInUse` | Any other page |
| DeleteImpact on Subject Fields | The recipe itself (**S8-13**) |
| L10n + VoiceOver | A Terms page; Subject type delete |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PF-1 | Trash always offered. DeleteImpact confirm when allowed; notice lists `OBS-…` when not. |
| PF-2 | Type-bindings are not blockers and are not listed. |
| PF-3 | After erase, selection falls back as today. |
| PF-4 | Instance DeleteImpact. No page-local refuse. |
| PF-5 | VoiceOver names the property and erase vs blocked. |

---

## 5. Suggested frames

1. Unused user property, no terms — confirm.
2. Property used on two Observations — notice names `OBS-…`.
3. Property bound to a type but unused on Observations — confirm (bindings CASCADE).
4. User property with unused terms — notice names the terms.
5. Seeded property — `origin_locked`.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Subject Fields trash | Snowflake | **Extend** | `SubjectFieldsView` | Replace `.pvConfirm`. |
| DeleteImpact | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/` | **S8-13**. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Disable-trash-when-used as the only UX | No names. |
| Cascade “delete property and Observations” | Policy is refuse. |
| Drawing Source Types or Source Fields | Other briefs. |

---

## 7. Out of scope

- Other vocabulary pages
- Redesigning the properties list
- ⌘Z

---

## 8. Handoff

1. Archive this brief under [`archive/`](archive/) and write [`../completed.md`](../completed.md).
2. Implement **S8-17** against these frames. Writer follows [`add-catalog-delete`](../../../../.cursor/skills/add-catalog-delete/SKILL.md).
