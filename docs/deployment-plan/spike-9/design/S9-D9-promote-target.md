# S9-D9 — Promote: shell + choose target

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Promote flow — choose target step (and the flow shell)  
**Implements later as:** PR **S9-26**  
**Depends on:** S9-15 (target suggestions), S9-25 (entry)  
**Related:** S9-D10…D12 extend this shell; conclusion model §5.3–5.4; [`deployment-plan.md`](../deployment-plan.md) R7  
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

Design the **Promote flow shell** and its first step. The researcher clicked Promote on a subject card and now chooses **where this subject goes**:

- **Mint a new** Person / Event / Place, or
- **Pick an existing** handle of the same kind.

```text
Promote  CPR-2AB91 · James Robins (1851 census, line 14)          [ Done ]
────────────────────────────────────────────────────────────────────────
Step 1 of 3 · Choose a Person

  ( ) New Person
  (•) Existing
      [ search Persons…                     ]
      Suggested
        ◯ James Robins     1817 – 1880 · York       PER-7KD45
        ◯ James Robbins    1819 –      · Kingston   PER-1QQ10
                                                        [ Next ]
```

**The shell decision is this board's:** a workspace **place** (like the citation composer, Option B — Back returns to the graph) or a **sheet** over the graph. The walk (S9-D12) can go on for many steps, which argues for a place. The shell must carry: which subject is being promoted, step progress, **Done** (off-ramp), and the leave guard for unsaved work.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Same kind only | A person subject can only join a Person. |
| Header rows | Existing handles show the same header as the list row (name, life dates, place, ref) so the researcher recognizes them. |
| Suggestions | During a walk, handles already related to the one just filed come first; then resemblance (name, date, toponym, event type). |
| Mint skips comparison | New handle → straight to claim fields (S9-D10). Existing handle with members → compare (S9-D11) first. |
| One-way | Promote creates claims only. No resume midstream; no editing past steps. |
| Steps save individually | Each subject's step saves on Next (S9-D10). Done ends the walk; saved steps stay. |

### 2.1 What this board is not

- Not the compare, claim-fields, or walk screens (later briefs extend this shell).
- Not merge or claim editing.

---

## 3. Implementation gate (S9-26)

| Ships in **S9-26** | Does **not** ship there |
| --- | --- |
| Promote shell (place or sheet) + choose-target step + leave guard | Claim fields save (S9-27), compare (S9-28), walk (S9-29) |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PT-1 | Shell shows the subject being promoted (ref, working label or name, its Source) throughout the flow. |
| PT-2 | Step indicator adapts: mint path has fewer steps than the join path. |
| PT-3 | Choose between **New {kind}** and **Existing**; existing uses a searchable pick of header rows. |
| PT-4 | Suggestions section with related-first ordering; empty suggestions state. |
| PT-5 | **Done** is always available; leaving with an unsaved step asks (Confirm). |
| PT-6 | Next is disabled until a target is chosen. |
| PT-7 | Shell decision (place vs sheet) documented on the board with the reason. |

---

## 5. Suggested frames

1. Shell + choose target, new selected.
2. Existing selected, suggestions list.
3. Existing with search typed, no match.
4. Leave-guard confirm.
5. During a walk: suggestions headed *Related to PER-7KD45*.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Promote shell | Snowflake | **New** | `Features/Promote/PromoteView.swift` | Place or sheet per board. |
| Promote model | Snowflake | **New** | `Features/Promote/PromoteModel.swift` | Step state, draft, leave guard. |
| Choose-target step | Snowflake | **New** | `Features/Promote/PromoteTargetStep.swift` | |
| ComboBox | Component | Ship | kit | Searchable existing-handle pick. |
| Header row | Snowflake | Ship (from D2) | `Features/Conclusions/ConclusionListRow.swift` | Reuse for candidates. |
| Button / Confirm | Component | Ship | kit | Next, Done, leave guard. |
| Section header | Component | Ship | kit | Suggested / Related. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A kit wizard / stepper | One flow; compose. |
| A second copy of the list row | Reuse D2. |

---

## 7. Out of scope

- Compare, claim fields, walk
- Provisional / rejected status

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-26** against the board and inventory (kit first).
