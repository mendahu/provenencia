# S9-D12 — Promote: walk + off-ramp

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Promote flow — walk step (connected subjects)  
**Implements later as:** PR **S9-29**  
**Depends on:** S9-D9 (shell), S9-D10 (save); S9-16 (neighborhood)  
**Related:** conclusion model §5.4 (neighborhood walk)  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief **extends S9-D9**'s frames. Reuse its layout; change only what this kind needs.

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

After a subject saves, show the subjects **connected to the one just filed** and offer each for promotion. Picking one runs the steps again for it. **Done** ends the walk; everything saved stays.

```text
Saved: CPR-2AB91 → PER-7KD45 James Robins

Connected to this record
  ◇ Birth (CEV-…)                        [ Promote ]
  ⌖ York (CPL-…)                          [ Promote ]
  ⋯ Participation: subject of Birth       waiting for Birth
  ⋯ Location: Birth at York               waiting for Birth, York
                                                        [ Done ]
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Queue grows | Each promoted subject adds its own unpromoted neighbors (not already queued or handled). Skipping does not reach through. |
| Bridges wait | Participation / Location / Relationship subjects become ready when both ends have handles, then are a **short confirm**: file onto the association the ends already share, or mint one. |
| Queue is not saved | Leaving drops unvisited subjects; they stay unpromoted on the graph. |
| Already-promoted neighbors | Shown as done (with their handle) or omitted — board decides. |

### 2.1 What this board is not

- Not a graph view of the neighborhood.
- Not bulk promote.

---

## 3. Implementation gate (S9-29)

| Ships in **S9-29** | Does **not** ship there |
| --- | --- |
| Walk list, bridge confirm, progress, Done | Queue persistence, bulk actions |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| WK-1 | After each save, a success note and the connected-subjects list. |
| WK-2 | Each primary subject row: kind, working label / value, ref, Promote. |
| WK-3 | Bridge rows show *waiting for …* until ready, then a one-click confirm (share existing / mint). |
| WK-4 | Progress: how many handled this pass. |
| WK-5 | Done always visible; ends the flow and returns to the graph. |
| WK-6 | Empty neighborhood: *Nothing else connected* + Done. |

---

## 5. Suggested frames

1. Birth-record walk after the person saves.
2. Bridge becomes ready; confirm.
3. Mid-walk after several saves.
4. Nothing connected.
5. Done.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Walk step | Snowflake | **New** | `Features/Promote/PromoteWalkStep.swift` | |
| Rows | Component | Ship | Card / list rows | |
| Button | Component | Ship | kit | Promote, confirm, Done. |
| Toast / Callout | Component | Ship | kit | Saved note. |
| Kind style | Snowflake | Ship | `EvidenceSubjectKindStyle` | Kind glyphs match the graph. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A mini graph | The list is the MVP. |

---

## 7. Out of scope

- Persisted queues
- Claim editing

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-29** against the board and inventory (kit first).
