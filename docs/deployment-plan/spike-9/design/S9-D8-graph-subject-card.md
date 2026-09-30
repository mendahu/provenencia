# S9-D8 — Evidence graph subject card: Promote + membership

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Evidence graph subject card (`EvidenceSubjectCard`)  
**Implements later as:** PR **S9-25**  
**Depends on:** S9-13 (graph read carries membership); shipped card ([`EvidenceSubjectCard`](../../../../macos/App/Features/EvidenceGraph/EvidenceSubjectCard.swift)) and Spike 8 graph chrome  
**Related:** S9-D9 (Promote opens there); conclusion model §5.3–5.4  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

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
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Every Property is **multi-valued** across members.
- Each field shows **one resolved value** in a state: **single**, **merged** (names and dates auto-reconciled), **mixed** (members disagree; top-ranked value shown), or **empty**. A future **concluded** state (a researcher's Reconciliation Claim) needs room but does not ship in Spike 9.
- The engine returns **structures** (dates, names, title parts); the app formats text. Do not design copy that assumes a pre-built sentence from the engine.
- **No likeness/photo value exists yet.** Thumbnails are a slot with a per-kind placeholder.
- Promote only **creates** claims, one subject per saved step, with a **Done** off-ramp. Editing or removing claims is a later workflow (Spike 10).

---

## 1. Objective

Two additions to the **primary-kind** subject cards (person, event, place):

1. **Promote** — a control at the **bottom** of an unpromoted card that starts the Promote flow (S9-D9).
2. **Membership** — a promoted card shows the handle it belongs to (`PER-7KD45` + resolved name) and links to its page. Promote is not offered again.

```text
┌ person · CPR-2AB91 ──────────── ✎ 🗑 ┐      ┌ person · CPR-2AB91 ──────── ✎ 🗑 ┐
│ name       James Robins              │      │ name       James Robins          │
│ age        34                        │      │ age        34                    │
│ + Add property                       │      │ + Add property                   │
│ [ Promote ]                          │      │ ● PER-7KD45 · James Robins  ›    │
└──────────────────────────────────────┘      └──────────────────────────────────┘
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Only primary kinds | Bridge cards (participation, location, relationship) get **no** Promote; they are filed during the walk (S9-D12). They may show membership once filed. |
| One accepted handle per subject | A promoted card never shows Promote. |
| Cards are paint-only | Actions are AppKit hit targets (`GraphCanvasActionTarget`). Promote and the membership link each need a target id and a fixed frame. |
| Card height is computed | `contentHeight(for:)` drives edges and hits; the new row must have a known height. |
| Membership is a Conclusion fact | Style must not read as an Observation row (it is not cited evidence). |

### 2.1 What this board is not

- Not the Promote flow (S9-D9…D12).
- Not redesigning cards, edges, or the canvas.
- Not a graph-wide *promoted* filter.

---

## 3. Implementation gate (S9-25)

| Ships in **S9-25** | Does **not** ship there |
| --- | --- |
| Promote control + membership row on primary cards; bridge membership if cheap | Promote flow screens |
| Action target ids + accessibility actions | Card redesign |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| GC-1 | Unpromoted person / event / place cards show **Promote** at the bottom, below Add property. |
| GC-2 | Promoted cards replace Promote with a membership row: handle ref + resolved name (formatted in-app) + disclosure to open the page. |
| GC-3 | Membership row is visually a Conclusion link, distinct from Observation rows and kind color. |
| GC-4 | Both are canvas action targets with stable ids and VoiceOver actions (*Promote*, *Open James Robins*). |
| GC-5 | Card height grows by a fixed row; edges still attach correctly. |
| GC-6 | Bridge cards: no Promote; board decides whether filed bridges show a small membership mark. |

---

## 5. Suggested frames

1. Unpromoted person card with Promote.
2. Promoted person card with membership row.
3. Event and place variants.
4. Bridge card unchanged (or with a small filed mark).
5. Hover / pressed states on the new controls.
6. Dense graph: several promoted cards.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Subject card | Snowflake | **Extend** | `Features/EvidenceGraph/EvidenceSubjectCard.swift` | New row + action targets. |
| Canvas pointer controller | Snowflake | **Extend** | `Features/GraphCanvas/` | New action ids. |
| Button | Component | Ship | kit | Promote (ghost / secondary). |
| Chip / Badge | Component | Ship | kit | Membership ref mark. |
| Kind style | Snowflake | Ship | `Features/EvidenceGraph/EvidenceSubjectKindStyle.swift` | Do not tint membership with kind color. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit membership component | One call site. |
| Promote in the header icon row | The ask is a bottom control; header stays edit / delete. |

---

## 7. Out of scope

- Promote screens
- Bulk promote

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-25** against the board and inventory (kit first).
