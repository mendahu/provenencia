# S9-D13 — Omnibar: Person / Event / Place hits

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Omnibar results dropdown  
**Implements later as:** PR **S9-35**  
**Depends on:** S9-34 (search kinds + structured headers); shipped [`PVOmnibarHitRow`](../../../../macos/App/DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift)  
**Related:** Omnibar contract [`omnibar-search.md`](../../../deployment-plan/archive/spike-3/omnibar-search.md); S9-D2 (row these should echo)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); [`add-searchable-kind`](../../../../.cursor/skills/add-searchable-kind/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief is an **enhancement** of a shipped surface: extend its existing frames.

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

Add Person, Event, and Place hits to the omnibar results. They use the shipped `PVOmnibarHitRow` slots (lead, title, kind, secondary, ref, match context) filled from the same header as the list row, so a hit reads like its list row.

```text
[ ◯ ]  James Robins                    Person    PER-7KD45
       1817 – 1880 · York              matched: "Jim Robins"
[ ◇ ]  Birth of James Robins           Event     EVT-4MA10
       14 May 1817 · York
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Alternates match | *Jim Robins* can find a Person reconciled as *James Robins*; show why (match context). |
| Structured header | The app formats title and secondary from structures (unlike today's kinds). |
| Ref match | Exact ref match uses the accent ref style. |
| Mixed flat ranking | Conclusion hits interleave with Sources etc. by score; no per-kind sections. |

### 2.1 What this board is not

- Not omnibar chrome, ranking, or facets.

---

## 3. Implementation gate (S9-35)

| Ships in **S9-35** | Does **not** ship there |
| --- | --- |
| Three hit kinds in the shipped row | New row layout, facets |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| OH-1 | Lead slot: per-kind placeholder thumbnail consistent with lists. |
| OH-2 | Title / secondary match the list row's title and secondary line. |
| OH-3 | Match context shows an alternate name / toponym when that is what matched. |
| OH-4 | Kind label: Person / Event / Place. |
| OH-5 | No new row variant. |

---

## 5. Suggested frames

1. Mixed results for *Robins* (Sources + Persons + Events).
2. Exact ref hit.
3. Alternate-name match.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Omnibar hit row | Recipe | Ship | `DesignSystem/Recipes/OmnibarHitRow/PVOmnibarHitRow.swift` | Fill slots. |
| Hit presentation | Snowflake | **Extend** | `Features/Workspace/OmnibarHitPresentation.swift` | Map the three kinds. |
| Thumbnail | Component | Ship | kit | Lead. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A Conclusion-specific hit row | Row is kind-agnostic by design. |

---

## 7. Out of scope

- Search ranking
- Subjects (`CPR-…`) as hits

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-35** against the board and inventory (kit first).
