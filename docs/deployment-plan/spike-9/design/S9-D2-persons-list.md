# S9-D2 — Persons list

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Persons list page  
**Implements later as:** PR **S9-19**  
**Depends on:** S9-11 (list reads), S9-12 (formatters), S9-18 (destination)  
**Related:** S9-D3 / S9-D4 extend this row anatomy; S9-D5 (Person detail); precedent [`SourcesListView`](../../../../macos/App/Features/Sources/SourcesListView.swift)  
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
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Every Property is **multi-valued** across members.
- Each field shows **one resolved value** in a state: **single**, **merged** (names and dates auto-reconciled), **mixed** (members disagree; top-ranked value shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

A list of every Person in the project. One row per handle:

```text
[ ◯ ]  James Robins                                  PER-7KD45
       14 May 1817 – 2 Jan 1880   ·   York, Upper Canada → Toronto
[ ◯ ]  Mary Smith  ⚠ mixed                           PER-3QX91
       ABT 1790 –                  ·   Kingston  +2
[ ◯ ]  PER-9ZZ02                                     PER-9ZZ02
       —
```

Fields: **thumbnail slot · name · birth date – death date · birth place · death place · ref**. Clicking a row opens the Person detail (S9-D5). This board establishes the **row anatomy** the Events and Places lists reuse.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Every value is resolved from many members | Each cell shows one value. Design a small *mixed* marker for disagreement and a *+N* for extra distinct values (places). |
| Name falls back | Resolved name → working `label` → ref. A row with only a ref must still look intentional. |
| Life dates are derived | From birth / death Events linked to the Person. Missing ⇒ empty, not "Unknown". A lone birth reads `1817 –`. |
| Places are derived | Birth / death Event → Location → Place toponym. Either may be missing. |
| No likeness yet | Thumbnail is a per-kind placeholder (person glyph). Do not imply a photo exists. |
| Sort | Default by name (resolved). No sort controls this spike. |
| Dates are structures | Formatting (`ABT`, ranges, partial dates) is the app's existing date display. Show realistic variety. |

### 2.1 What this board is not

- Not the detail page (S9-D5).
- Not filters, sort controls, search-within-list (the omnibar covers search).
- Not member / Source counts per row (not in the row spec).
- Not bulk actions or row context menus beyond open.

---

## 3. Implementation gate (S9-19)

| Ships in **S9-19** | Does **not** ship there |
| --- | --- |
| Persons list: rows, markers, empty state, loading | Sort / filter controls |
| Row → Person detail navigation | Row actions (merge, delete, edit) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PL-1 | Row shows thumbnail slot, name, birth – death, birth place, death place, and ref. Ref in the shipped mono ref style. |
| PL-2 | *Mixed* marker on any cell whose value is mixed; compact (Badge / Chip / Mark), not a full breakdown. |
| PL-3 | *+N* on place cells when more than one distinct value exists. |
| PL-4 | Missing values are blank; a name-less Person shows its ref as title without looking broken. |
| PL-5 | Empty state (no Persons yet) explains Persons come from **Promote** on an Evidence graph card. |
| PL-6 | Row is one navigation target; VoiceOver reads the row as name, dates, places, ref. |
| PL-7 | Dense enough for hundreds of rows; long toponyms truncate gracefully. |
| PL-8 | Row anatomy is reusable by Events (S9-D3) and Places (S9-D4): define which slots are generic (thumbnail, title, secondary line, ref, markers). |

---

## 5. Suggested frames

1. Typical list: complete rows, partial rows (birth only), a ref-only row.
2. Mixed name and mixed birth date rows.
3. Place cell with *+2*.
4. Empty state.
5. Loading (first load) vs refreshing (stale content kept).
6. Long names / toponyms truncating.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Persons list | Snowflake | **New** | `Features/Conclusions/PersonsListView.swift` | Reads the list query key. |
| Conclusion list row | Snowflake | **New** | `Features/Conclusions/ConclusionListRow.swift` | Shared by D2–D4 (three call sites) — a feature snowflake, not kit, unless the board argues otherwise. |
| Thumbnail | Component | Ship | `DesignSystem/Components/Thumbnail/` | Placeholder per kind. |
| Mixed marker | Component | Ship | `Badge` / `Chip` / `Recipes/Marks` | Pick one on the board. |
| Empty state | Component | Ship | `DesignSystem/Components/EmptyState/` | Points at Promote. |
| Table / list | Component | Ship | `DesignSystem/Components/Table/` or the Sources list pattern | Board picks; match Sources list density. |
| Date / name display | Snowflake | Ship / Extend | `Features/Dates/DateValueDisplay.swift`, `Features/Names/NameValueDisplay.swift` | Formatting only. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit `PVEntityRow` | Three call sites in one feature; stay a feature snowflake. |
| Inline editing in rows | Pages are read-only this spike. |

---

## 7. Out of scope

- Detail page
- Promote
- Search

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-19** against the board and inventory (kit first).
