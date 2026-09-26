# S8-D14 — Source Fields delete

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 8 (pause and refine / data entry)  
**Implements later as:** PR **S8-16** (cut over `sourcefields.Delete` + DeleteImpact on this page)  
**Depends on:** Frozen policy + Impact report in [`../deployment-plan.md`](../deployment-plan.md) §S8-09.1–§S8-09.4; **S8-12** `GetDeleteImpact`; **S8-D9** / **S8-13** DeleteImpact recipe  
**Related:** Shared confirm / notice is [`S8-D9-impact.md`](S8-D9-impact.md) — **instance it, do not restyle.** Source Types is [`S8-D13-source-types-delete.md`](S8-D13-source-types-delete.md). Subject Fields is [`S8-D15-subject-fields-delete.md`](S8-D15-subject-fields-delete.md). Do not draw those pages.  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of the shipped Source Fields page. Do not start a second fields app.

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

Source Fields already deletes unused fields (`used_by == 0` disables trash; `.pvConfirm` yes/no). Replace that with DeleteImpact so an in-use field **names the Sources that store a value for it**.

**This board designs** trash + DeleteImpact on **this page only**. Unused field → confirm. In-use → notice lists those Sources (`SRC-…`). Trash stays offered when `used_by > 0`. Seeded fields may still be deleted when unused.

Type-suggestion joins and layout rows CASCADE — do not list them as blockers. `used_by` may stay as a column; it must match the Impact totals that block. **Origin:** plugin never; seeded and user may erase when unused. Clearing a metadata *value* on the Source page is a facet write (already shipped); that is not this trash.

---

## 2. Domain facts (this board)

| Fact | UI implication |
| --- | --- |
| Blocked by `source_metadata.field_id` | Notice lists Sources that have a saved value. |
| `used_by` on the list | May stay as a column; not the only gate. |
| Suggestion / layout CASCADE | Unused field can go even if types still suggest it. |
| After erase | `fallbackToSectionRoot` as today. |

### 2.1 What this board is not

- Not Source Types (**S8-D13**) or Subject Fields (**S8-D15**).
- Not restyling DeleteImpact (**S8-D9**).
- Not Source-page metadata-value delete (facet Confirm).

---

## 3. Implementation gate (S8-16)

| Ships in **S8-16** | Does **not** ship there |
| --- | --- |
| `sourcefields.Delete` through `deleteimpact`; drop `sqlInUse` | Any other page |
| DeleteImpact on Source Fields | The recipe itself (**S8-13**) |
| L10n + VoiceOver | Redesigning the fields list |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| SF-1 | Trash always offered. DeleteImpact confirm when allowed; notice lists `SRC-…` when not. |
| SF-2 | Do not require `used_by == 0` to enable trash. |
| SF-3 | After erase, selection falls back as today. |
| SF-4 | Instance DeleteImpact. No page-local refuse. |
| SF-5 | VoiceOver names the field and erase vs blocked. |

---

## 5. Suggested frames

1. Unused user field — confirm.
2. Field with saved values on two Sources — notice names `SRC-…`.
3. Seeded unused field — still deletable.
4. Plugin field — `origin_locked`.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Source Fields trash | Snowflake | **Extend** | `SourceFieldsView` | Replace `.pvConfirm` + disable-on-count. |
| DeleteImpact | Recipe | Ship | `DesignSystem/Recipes/DeleteImpact/` | **S8-13**. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Disable-trash-when-used as the only UX | No names. |
| Cascade “delete field and values” | Policy is refuse. |
| Drawing Source Types or Subject Fields | Other briefs. |

---

## 7. Out of scope

- Other vocabulary pages
- Redesigning the fields list
- ⌘Z

---

## 8. Handoff

1. Archive this brief under [`archive/`](archive/) and write [`../completed.md`](../completed.md).
2. Implement **S8-16** against these frames. Writer follows [`add-catalog-delete`](../../../../.cursor/skills/add-catalog-delete/SKILL.md).
